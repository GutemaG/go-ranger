package pkg

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
)

func (a *App) addNewFile() {
	var fileName string
	form := tview.NewForm().
		AddInputField("Name (add '/' for directory)", "", 20, nil, func(text string) {
			fileName = text
		}).
		AddButton("Create", func() {
			if fileName != "" {
				filePath := filepath.Join(a.currentDir, fileName)
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
					a.updatePanes()
				}
			}
			a.pages.SwitchToPage("main")
			a.tviewApp.SetFocus(a.middlePane)
		}).
		AddButton("Cancel", func() {
			a.pages.SwitchToPage("main")
			a.tviewApp.SetFocus(a.middlePane)
		})

	form.SetBorder(true).SetTitle("Add New File/Directory").SetTitleAlign(tview.AlignCenter)

	a.pages.AddPage("addFile", form, true, true)
	a.tviewApp.SetFocus(form)
}

func (a *App) deleteFile(item string) {
	cleanedItem := getCleanedItemName(item)
	path := filepath.Join(a.currentDir, cleanedItem)
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
					a.updatePanes()
				}
			}
			a.pages.SwitchToPage("main")
			a.tviewApp.SetFocus(a.middlePane)
		})

	a.pages.AddPage("deleteConfirm", modal, true, true)
	a.tviewApp.SetFocus(modal)
}
