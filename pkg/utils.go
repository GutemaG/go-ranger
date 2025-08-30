package pkg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

// Function to get the item name without TUI color tags or trailing slashes
func getCleanedItemName(item string) string {
	cleanedItem := strings.TrimSuffix(item, "/")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[darkcyan]", "")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[white]", "")
	return strings.TrimSpace(cleanedItem)
}

// FileEntry is exported for use in the tui package.
type FileEntry struct {
	os.DirEntry
	Info   os.FileInfo
	Size   int64
	Count  int
	IsFile bool
}

// getDirSizeAndCount recursively calculates the total size and element count of a directory.
func getDirSizeAndCount(path string) (int64, int, error) {
	var totalSize int64
	var count int
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == path {
			return nil // Skip the directory itself
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		totalSize += info.Size()
		count++
		return nil
	})
	return totalSize, count, err
}

// GetEntries reads a directory and returns sorted lists of directories and files.
func GetEntries(path string) ([]FileEntry, []FileEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading directory %s: %w", path, err)
	}

	var directories []FileEntry
	var files []FileEntry

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't get info for
		}
		fileEntry := FileEntry{DirEntry: entry, Info: info}
		if entry.IsDir() {
			directories = append(directories, fileEntry)
		} else {
			files = append(files, fileEntry)
		}
	}

	// Sort directories and files alphabetically
	sort.Slice(directories, func(i, j int) bool { return directories[i].Name() < directories[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	return directories, files, nil
}

// CreateEntry creates a new file or directory.
func CreateEntry(path string, isDir bool) error {
	if isDir {
		return os.MkdirAll(path, 0755)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// GetCleanedItemName removes TUI formatting from a string.
func GetCleanedItemName(item string) string {
	cleanedItem := strings.TrimSuffix(item, "/")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[darkcyan]", "")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[white]", "")
	return strings.TrimSpace(cleanedItem)
}
