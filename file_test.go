package loge

import (
	"fmt"
	"os"
	"sort"
	"testing"
)

func TestStoragePercent(t *testing.T) {
	fmt.Println("Testing getStoragePercent")
	fmt.Println(getStoragePercent("."))
}

func TestFlushAll(t *testing.T) {
	fmt.Println("Testing flushAll")

	ft := newFileTransport(nil, "./logs", "", true, false)
	ft.flushAll()

	storageThreshold := 0.0
	ft.path = "./logs"

	fileList, _ := os.ReadDir(ft.path)
	sort.Slice(fileList,
		func(x int, y int) bool {
			return fileList[x].Name() > fileList[y].Name()
		})
	fmt.Println(fileList)
	fmt.Println(storageThreshold)
	fmt.Println(len(fileList))
	for getStoragePercent(ft.path) > storageThreshold && len(fileList) >= 1 {
		fmt.Println(fileList[0].Name())
		os.Remove(ft.path + "/" + fileList[0].Name())
		fileList = fileList[1:]
	}
}
