package loge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const dateTimeStringLength = 27
const maxFileSize = 500

func getLogName(path string) string {
	fileList, _ := os.ReadDir(path)
	filesExist := len(fileList) > 0
	if filesExist {
		sort.Slice(fileList,
			func(x int, y int) bool {
				return fileList[x].Name() < fileList[y].Name()
			})
	}
	t := time.Now()
	if filesExist {
		dateStr := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())
		if fileList[0].Name()[:dateTimeStringLength] != dateStr {
			return filepath.Join(path, dateStr+"_1.log")
		}
	}
	fileNum := int(fileList[0].Name()[dateTimeStringLength+2])
	ret := fmt.Sprintf("%d%02d%02d_%d.log", t.Year(), t.Month(), t.Day(), fileNum)

	if getFileSize(filepath.Join(path, ret)) > maxFileSize {
		ret = fmt.Sprintf("%d%02d%02d_%d.log", t.Year(), t.Month(), t.Day(), fileNum+1)
	}

	return filepath.Join(path, ret)
}

func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func dumpTimeToBuffer(buf *[]byte, t time.Time) {
	*buf = (*buf)[:0]
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '/')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '/')
	itoa(buf, day, 2)
	*buf = append(*buf, ' ')

	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	*buf = append(*buf, '.')
	itoa(buf, t.Nanosecond()/1e3, 6)
	*buf = append(*buf, ' ')
}

func getFileSize(file string) int64 {
	var size int64
	size = -1
	f, err := os.Open(file)
	if err != nil {
		return size
	}
	defer f.Close()

	fInfo, _ := f.Stat()
	size = fInfo.Size()
	return size
}
