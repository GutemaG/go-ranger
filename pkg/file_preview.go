package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// preview shows a file's content or a directory's listing.
func (a *App) preview() {
	if a.middlePane.GetItemCount() == 0 {
		return
	}
	item, _ := a.middlePane.GetItemText(a.middlePane.GetCurrentItem())
	a.updateBottomBarInfo()

	cleanedItem := GetCleanedItemName(item)
	newPath := filepath.Join(a.currentDir, cleanedItem)
	fileInfo, err := os.Stat(newPath)
	if err != nil {
		return
	}

	if fileInfo.IsDir() {
		dirs, files, dirErr := GetEntries(newPath)
		if dirErr != nil {
			a.rightPane.SetText("Error reading directory: " + dirErr.Error())
			return
		}
		if len(dirs)+len(files) == 0 {
			a.rightPane.SetText("[red]Empty Directory[/]")
		} else {
			var sb strings.Builder
			for _, entry := range dirs {
				sb.WriteString(fmt.Sprintf("[darkcyan]%s/[white]\n", entry.Name()))
			}
			for _, entry := range files {
				sb.WriteString(fmt.Sprintf("%s\n", entry.Name()))
			}
			a.rightPane.SetText(sb.String())
		}
		a.rightPane.SetTitle("Contents: " + cleanedItem)
	} else {
		content, err := os.ReadFile(newPath)
		if err != nil {
			a.rightPane.SetText("Error reading file: " + err.Error())
			return
		}
		a.rightPane.SetText(string(content))
		a.rightPane.SetTitle("Preview: " + cleanedItem)
	}
}
