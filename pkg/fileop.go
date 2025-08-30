package pkg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
