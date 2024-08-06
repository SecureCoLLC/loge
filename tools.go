package loge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

const dateTimeStringLength = 27

const maxFileSize = 10 * 1024 * 1024

func getLogName(path string) string {
	t := time.Now()
	ret := fmt.Sprintf("%d%02d%02d_", t.Year(), t.Month(), t.Day())
	fileList, _ := os.ReadDir(path)
	sort.Slice(fileList,
		func (x int, y int) bool {
			return fileList[x].Name() < fileList[y].Name()
		})

	fileNum := 0

	if len(fileList) == 0 {
		return filepath.Join(path, ret + fmt.Sprintf("%04d.log", fileNum))
	}
	
	lastFileName := fileList[len(fileList) - 1].Name()
	lastNum, _ := strconv.Atoi(lastFileName[9:13])
	fileNum = lastNum

	pathToFile := filepath.Join(path, ret + fmt.Sprintf("%04d.log", fileNum))
	fi, err := os.Stat(pathToFile)
	if err == nil {
		if fi.Size() > maxFileSize {
			fileNum += 1
		}
	}

	if fileNum > 9999 {
		//idk man
	}
	ret += fmt.Sprintf("%04d.log", fileNum)
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
