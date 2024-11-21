package loge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateTimeStringLength = 27

const maxFileSize = 10 * 1024 * 1024

func getLogName(path string) (string, error) {
	t := time.Now()
	ret := fmt.Sprintf("%d%02d%02d_", t.Year(), t.Month(), t.Day())
	fileDirList, _ := os.ReadDir(path)

	fileList := make([]string, 0)
	for _, file := range fileDirList {
		if len(file.Name()) == 17 && file.Name()[len(file.Name())-4:] == ".log" {
			fileList = append(fileList, filepath.Join(path, file.Name()))
		}
	}
	if len(fileList) > 1 {
		sort.Slice(fileList,
			func(x int, y int) bool {
				f1, err1 := os.Stat(fileList[x])
				f2, err2 := os.Stat(fileList[y])
				if err1 != nil || err2 != nil {
					return true
				}
				return f1.ModTime().Before(f2.ModTime())
			})
	}

	for len(fileList) >= 1 && !strings.Contains(fileList[0], ret) {
		fileList = fileList[1:]
	}

	if len(fileList) == 0 {
		ret += fmt.Sprintf("%04d.log", 0)
		return filepath.Join(path, ret), nil
	}

	fileNum := 0
	recentFile := fileList[len(fileList)-1]
	newNum, err := strconv.Atoi(recentFile[len(recentFile)-8 : len(recentFile)-4])
	if err == nil {
		fileNum = newNum
	}

	pathToFile := filepath.Join(path, ret+fmt.Sprintf("%04d.log", fileNum))
	fi, err := os.Stat(pathToFile)
	if err == nil {
		if fi.Size() > maxFileSize {
			fileNum += 1
		}
	} else {
		return "", err
	}

	delNum := fileNum + 1
	if fileNum > 9999 {
		delNum = 0
		fileNum = 0
	}
	delFile := ret + fmt.Sprintf("%04d.log", delNum)
	delPath := filepath.Join(path, delFile)
	_, delErr := os.Stat(delPath)
	if delErr != nil {
		os.Remove(delPath)
	}
	ret += fmt.Sprintf("%04d.log", fileNum)
	return filepath.Join(path, ret), nil
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
