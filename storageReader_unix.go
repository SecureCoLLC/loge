package loge

import (
	"syscall"
)

func getStoragePercent(logDir string) float64 {
	var fileSystemStats syscall.Statfs_t
	if err := syscall.Statfs(logDir, &fileSystemStats); err != nil {
		return -1
	}

	totalBytes := float64(fileSystemStats.Blocks * uint64(fileSystemStats.Bsize))
	usedBytes := float64((fileSystemStats.Blocks - fileSystemStats.Bavail) * uint64(fileSystemStats.Bsize))

	percentUsage := (usedBytes / totalBytes)
	return percentUsage
}
