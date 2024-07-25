package loge

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
)

const storageThreshold = 0.9

type fileOutputTransport struct {
	buffer          TransactionList
	currentFilename string
	file            *os.File
	writer          *bufio.Writer
	done            chan struct{}
	wg              sync.WaitGroup

	path     string
	filename string
	rotation bool
	json     bool

	terminated bool

	signal chan struct{}

	trans       []uint64
	transLocker sync.Mutex
}

func newFileTransport(buffer TransactionList, path string, filename string, rotation bool, json bool) *fileOutputTransport {
	ft := &fileOutputTransport{
		buffer:   buffer,
		done:     make(chan struct{}),
		signal:   make(chan struct{}, 1),
		trans:    make([]uint64, 0),
		path:     path,
		filename: filename,
		rotation: rotation,
		json:     json,
	}

	go ft.loop()
	return ft
}

func (ft *fileOutputTransport) loop() {
	ft.wg.Add(1)
	defer ft.wg.Done()

	for {
		select {
		case <-ft.done:
			ft.flushAll()
			return
		case <-ft.signal:
			ft.flushAll()
		}
	}
}

func (ft *fileOutputTransport) NewTransaction(id uint64) {
	ft.transLocker.Lock()
	ft.trans = append(ft.trans, id)
	ft.transLocker.Unlock()

	select {
	case ft.signal <- struct{}{}:
	default:
	}
}

func (ft *fileOutputTransport) Stop() {
	close(ft.done)
	ft.wg.Wait()
}

func (ft *fileOutputTransport) flushAll() {
	if ft.terminated {
		return
	}

	fileList, _ := os.ReadDir(ft.path)
	sort.Slice(fileList,
		func (x int, y int) bool {
			return fileList[x].Name() > fileList[y].Name()
		})
	for getStoragePercent(ft.path) > storageThreshold && len(fileList) >= 1 {
		os.Remove(filepath.Join(ft.path, fileList[0].Name()))
		fileList = fileList[1:]
	}
	
	if ft.file != nil {
		if ft.rotation {
			if ft.currentFilename != getLogName(ft.path) {
				ft.file.Close()
				ft.file = nil
				ft.writer = nil
			}
		}
	}

	if ft.file == nil {
		ft.createFile()
		if ft.file == nil {
			ft.terminated = true
			os.Stderr.Write([]byte("Unable to create the output file.  Log file output is disabled.\n"))
			return
		}
	}

	ft.transLocker.Lock()
	if len(ft.trans) == 0 {
		ft.transLocker.Unlock()
		return
	}

	ids := ft.trans
	ft.trans = make([]uint64, 0)
	ft.transLocker.Unlock()

	for _, id := range ids {
		tr, ok := ft.buffer.Get(id, true)
		if ok {
			for _, be := range tr.Items {
				if ft.json {
					json, err := be.Marshal()
					if err == nil {
						ft.writer.Write(json)
						ft.writer.Write([]byte("\n"))
					}
				} else {
					ft.writer.Write(be.Timestring[:])
					ft.writer.Write([]byte(be.Message))
					ft.writer.Write([]byte("\n"))
				}
			}
		}
	}

	ft.writer.Flush()
}

func (ft *fileOutputTransport) createFile() {
	if ft.rotation {
		ft.currentFilename = getLogName(ft.path)
	} else {
		ft.currentFilename = filepath.Join(ft.path, ft.filename)
	}

	var err error
	ft.file, err = os.OpenFile(ft.currentFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		ft.file = nil
		return
	}

	ft.writer = bufio.NewWriter(ft.file)
}

func getStoragePercent(logDir string) float64 {
	// TODO: Windows code
	// import "golang.org/x/sys/windows"

	// var freeBytesAvailable uint64
	// var totalNumberOfBytes uint64
	// var totalNumberOfFreeBytes uint64

	// err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr("C:"),
	//     &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)

	// Returns a float64 between 0 and 1 representing the percent of disk space taken up
	var fileSystemStats syscall.Statfs_t
	if err := syscall.Statfs(logDir, &fileSystemStats); err != nil {
		return -1
	}

	totalBytes := float64(fileSystemStats.Blocks * uint64(fileSystemStats.Bsize))
	usedBytes := float64((fileSystemStats.Blocks - fileSystemStats.Bavail) * uint64(fileSystemStats.Bsize))

	percentUsage := (usedBytes / totalBytes)
	return percentUsage
}
