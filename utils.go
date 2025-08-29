package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func convertFileSize(size int64) string {
	const (
		byteUnit = 1
		kbUnit   = 1024
		mbUnit   = 1024 * 1024
		gbUnit   = 1024 * 1024 * 1024
		tbUnit   = 1024 * 1024 * 1024 * 1024
	)

	switch {
	case size == 0:
		return "0 bytes"
	case size < kbUnit:
		return fmt.Sprintf("%d bytes", size)
	case size < mbUnit:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(kbUnit))
	case size < gbUnit:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(mbUnit))
	case size < tbUnit:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(gbUnit))
	default:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(tbUnit))
	}
}

func GetDiskInfo(path string) (map[string]string, error) {
	var stat unix.Statfs_t

	err := unix.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	// Calculate raw sizes in bytes
	blockSize := uint64(stat.Bsize)
	freeBytes := stat.Bfree * blockSize
	totalBytes := stat.Blocks * blockSize
	usedBytes := totalBytes - freeBytes

	// Convert to human-readable format
	return map[string]string{
		"free":        formatBytes(freeBytes),
		"total":       formatBytes(totalBytes),
		"used":        formatBytes(usedBytes),
		"free_bytes":  fmt.Sprintf("%d", freeBytes),
		"total_bytes": fmt.Sprintf("%d", totalBytes),
		"used_bytes":  fmt.Sprintf("%d", usedBytes),
		"usage_percentage": fmt.Sprintf("%.1f%%",
			float64(usedBytes)/float64(totalBytes)*100),
	}, nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
