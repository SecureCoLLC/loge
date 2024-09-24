package loge

import (
	"golang.org/x/sys/windows"
)

func getStoragePercent(logDir string) float64 {
	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr("."),
		&freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return -1
	}

	// Returns a float64 between 0 and 1 representing the percent of disk space taken up
	var percentUsage float64 = 1 - (float64(freeBytesAvailable) / float64(totalNumberOfBytes))
	return percentUsage
}
