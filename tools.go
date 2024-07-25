package loge

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const dateTimeStringLength = 27

const maxFileSize = 1000

func getLogName(path string) string {
	t := time.Now()
	ret := fmt.Sprintf("%d%02d%02d_", t.Year(), t.Month(), t.Day())
	fileNum := 0
	for fileNum <= 9999 {
		tempPath := filepath.Join(path, ret + fmt.Sprintf("%04d.log", fileNum))
		_, err := os.Stat(tempPath)
		if err != nil {
			break
		}
		fileNum += 1
	}
	prevFileNum := fileNum - 1
	if prevFileNum >= 0 {
		tempPath := filepath.Join(path, ret + fmt.Sprintf("%04d.log", prevFileNum))
		fi, _ := os.Stat(tempPath)
		if fi.Size() < maxFileSize {
			fileNum = prevFileNum
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
