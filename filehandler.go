package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rivo/tview"
)

// FileEntry represents a file or directory with its associated info.
type FileEntry struct {
	os.DirEntry
	Info os.FileInfo
}

func updateLeftPane() {
	leftPane.Clear()
	var parentDirectories []FileEntry
	parentDir := filepath.Dir(currentDir)

	if parentDir != currentDir {
		leftPane.AddItem(filepath.Base(parentDir), "", 'h', nil)
	}

	parentEntries, err := os.ReadDir(parentDir)

	if err != nil {
		log.Printf("Error reading directory %s: %v", parentDir, err)
		return
	}
	for _, entry := range parentEntries {
		info, err := entry.Info()
		if err != nil {
			log.Printf("Error getting file info for %s: %v", entry.Name(), err)
			continue
		}
		fileEntry := FileEntry{DirEntry: entry, Info: info}
		if entry.IsDir() {
			parentDirectories = append(parentDirectories, fileEntry)
		}
	}
	for _, entry := range parentDirectories {
		leftPane.AddItem(fmt.Sprintf("[darkcyan]%s/[white]", entry.Name()), "", 0, nil)
	}
}

func updateMiddlePane() {
	middlePane.Clear()

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", currentDir, err)
		return
	}

	if len(entries) == 0 {
		middlePane.SetTitle("[red]Current (Empty Directory)[/]")
		return
	}

	var directories []FileEntry
	var files []FileEntry

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Printf("Error getting file info for %s: %v", entry.Name(), err)
			continue
		}
		fileEntry := FileEntry{DirEntry: entry, Info: info}
		if entry.IsDir() {
			directories = append(directories, fileEntry)
		} else {
			files = append(files, fileEntry)
		}
	}

	// Sort directories and files alphabetically by name
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Name() < directories[j].Name()
	})
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Populate the middle pane with sorted directories first
	for _, entry := range directories {
		middlePane.AddItem(fmt.Sprintf("[darkcyan]%s/[white]", entry.Name()), "", 0, nil)
	}

	// Then populate with sorted files
	for _, entry := range files {
		middlePane.AddItem(entry.Name(), "", 0, nil)
	}

	middlePane.SetTitle(fmt.Sprintf("Current (%d items)", len(entries)))
	pages.SetTitle(currentDir) // Update the main page title with the current directory

	// Manually set the current item to 0 to trigger the preview for the first item
	if middlePane.GetItemCount() > 0 {
		middlePane.SetCurrentItem(0)
	}
}

// updatePanes updates the contents of the file lists.
func updatePanes() {
	updateLeftPane()
	updateMiddlePane()
	rightPane.SetText("")

}

// Function to get the item name without TUI color tags or trailing slashes
func getCleanedItemName(item string) string {
	cleanedItem := strings.TrimSuffix(item, "/")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[darkcyan]", "")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[white]", "")
	return strings.TrimSpace(cleanedItem)
}

// previewFile previews the content of a file.
func previewFile(newPath string, cleanedItem string) {
	content, err := os.ReadFile(newPath)
	if err != nil {
		rightPane.SetText("Error reading file: " + err.Error())
		rightPane.SetTitle("Preview Error")
		return
	}
	rightPane.SetText(string(content))
	rightPane.SetTitle("Preview: " + cleanedItem)
}

func getFileInfo(item string) (os.FileInfo, error) {
	cleanedItem := getCleanedItemName(item)
	newPath := filepath.Join(currentDir, cleanedItem)
	fileInfo, err := os.Stat(newPath)
	return fileInfo, err
}

// navigateOrShowFile navigates into a directory or shows file content.
func navigateOrShowFile(item string) {
	cleanedItem := getCleanedItemName(item)
	newPath := filepath.Join(currentDir, cleanedItem)
	fileInfo, err := os.Stat(newPath)

	if err != nil {
		log.Printf("Error stating file: %v", err)
		return
	}

	if fileInfo.IsDir() {
		currentDir = newPath
		updatePanes()
	} else {
		content, err := os.ReadFile(newPath)
		if err != nil {
			modal := tview.NewModal().
				SetText("Error reading file: " + err.Error()).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					pages.SwitchToPage("main")
					app.SetFocus(middlePane)
				})
			pages.AddPage("fileViewer", modal, true, true)
			return
		}

		modal := tview.NewModal().
			SetText(string(content)).
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				pages.SwitchToPage("main")
				app.SetFocus(middlePane)
			})

		modal.SetTitle("Viewing: " + cleanedItem)
		pages.AddPage("fileViewer", modal, true, true)
		app.SetFocus(pages)
	}
}

// addNewFile creates a new file or directory in the current directory.
func addNewFile() {
	var fileName string
	form := tview.NewForm().
		AddInputField("Name (add '/' for directory)", "", 20, nil, func(text string) {
			fileName = text
		}).
		AddButton("Create", func() {
			if fileName != "" {
				filePath := filepath.Join(currentDir, fileName)
				var err error
				if strings.HasSuffix(fileName, "/") {
					err = os.MkdirAll(filePath, 0755)
				} else {
					var file *os.File
					file, err = os.Create(filePath)
					if err == nil {
						file.Close()
					}
				}
				if err != nil {
					log.Printf("Error creating entry %s: %v", filePath, err)
				} else {
					updatePanes()
				}
			}
			pages.SwitchToPage("main")
			app.SetFocus(middlePane)
		}).
		AddButton("Cancel", func() {
			pages.SwitchToPage("main")
			app.SetFocus(middlePane)
		})

	form.SetBorder(true).SetTitle("Add New File/Directory").SetTitleAlign(tview.AlignCenter)

	pages.AddPage("addFile", form, true, true)
	app.SetFocus(form)
}

// deleteFile confirms and deletes the selected file or directory.
func deleteFile(item string) {
	cleanedItem := getCleanedItemName(item)
	path := filepath.Join(currentDir, cleanedItem)
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("Error stating file: %v", err)
		return
	}

	modalText := fmt.Sprintf("Are you sure you want to delete '%s'?", cleanedItem)
	if fileInfo.IsDir() {
		modalText = fmt.Sprintf("Are you sure you want to delete the directory '%s' and its contents?", cleanedItem)
	}

	modal := tview.NewModal().
		SetText(modalText).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				err := os.RemoveAll(path)
				if err != nil {
					log.Printf("Error deleting file: %v", err)
				} else {
					updatePanes()
				}
			}
			pages.SwitchToPage("main")
			app.SetFocus(middlePane)
		})

	pages.AddPage("deleteConfirm", modal, true, true)
	app.SetFocus(modal)
}

// goUpDirectory navigates up a directory.
func goUpDirectory() {
	currentDir = filepath.Clean(filepath.Join(currentDir, ".."))
	updatePanes()
}
