package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// setupUI initializes and configures all TUI components.
func setupUI() {
	app = tview.NewApplication()

	leftPane = tview.NewList()
	leftPane.ShowSecondaryText(false)
	leftPane.SetBorder(true)
	leftPane.SetTitle("Parent")

	middlePane = tview.NewList()
	middlePane.ShowSecondaryText(false)
	middlePane.SetBorder(true)
	middlePane.SetTitle("Current")

	rightPane = tview.NewTextView()
	rightPane.SetBorder(true)
	rightPane.SetTitle("Preview")
	rightPane.SetDynamicColors(true)
	rightPane.SetText("Select a file to preview.")

	mainFlex = tview.NewFlex().
		AddItem(leftPane, 0, 1, false).
		AddItem(middlePane, 0, 3, true).
		AddItem(rightPane, 0, 3, false)

	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		AddItem(mainFlex, 1, 0, 1, 1, 0, 0, false)

	bottomBarLeft = tview.NewTextView()
	bottomBarRight = tview.NewTextView()
	bottomBar = tview.NewFlex().
		AddItem(bottomBarLeft, 0, 2, false).
		AddItem(bottomBarRight, 0, 2, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, false).
		AddItem(bottomBar, 1, 0, false)

	pages = tview.NewPages().AddPage("main", flex, true, true)
	pages.SetBorder(true)

	setEventHandlers()
}

func updateBottomBarInfo() {
	item, _ := middlePane.GetItemText(middlePane.GetCurrentItem())
	fileInfo, err := getFileInfo(item)
	if err != nil {
		return
	}
	stat, _ := fileInfo.Sys().(*syscall.Stat_t)
	uid := stat.Uid
	var fileOwner string
	ownerUser, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		fileOwner = ""
	} else {
		fileOwner = ownerUser.Username
	}
	file_mode := fileInfo.Mode().String()
	modified_time := fileInfo.ModTime().Format("2006-01-02 15:04:05")
	bottomBarLeft.SetText(file_mode + " " + fileOwner + " " + modified_time)

	diskInfo, _ := GetDiskInfo("/")
	total_items := strconv.Itoa(middlePane.GetItemCount())
	currentIndex := strconv.Itoa(middlePane.GetCurrentItem())

	bottomBarRight.SetText(convertFileSize(fileInfo.Size()) + "," + diskInfo["free"] + " free" + ", (" + currentIndex + "/" + total_items + ")")

}

// setEventHandlers sets up all the key bindings and event functions.
func setEventHandlers() {
	leftPane.AddItem("..", "", 'h', goUpDirectory)

	middlePane.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if middlePane.GetItemCount() > 0 {
			cleanedItem := getCleanedItemName(mainText)
			newPath := filepath.Join(currentDir, cleanedItem)
			fileInfo, err := os.Stat(newPath)
			updateBottomBarInfo()
			if err == nil {
				if !fileInfo.IsDir() {
					previewFile(newPath, cleanedItem)
				} else {
					entries, dirErr := os.ReadDir(newPath)
					if dirErr != nil {
						rightPane.SetText("Error reading directory: " + dirErr.Error())
						rightPane.SetTitle("Preview Error")
						return
					}

					if len(entries) == 0 {
						rightPane.SetText("[red]Empty Directory[/]")
					} else {
						var sb strings.Builder
						for _, entry := range entries {
							if entry.IsDir() {
								sb.WriteString(fmt.Sprintf("[darkcyan]%s/[white]\n", entry.Name()))
							} else {
								sb.WriteString(fmt.Sprintf("%s\n", entry.Name()))
							}
						}
						rightPane.SetText(sb.String())
					}
					rightPane.SetTitle("Contents: " + cleanedItem)
				}
			}
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
			}
		case tcell.KeyCtrlC:
			app.Stop()
		}
		return event
	})

	middlePane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if middlePane.GetItemCount() == 0 {

			if event.Key() == tcell.KeyLeft || (event.Key() == tcell.KeyRune && event.Rune() == 'h') {
				goUpDirectory()
				return nil
			} else if event.Key() == tcell.KeyRune && event.Rune() == 'a' {
				addNewFile()
				return nil
			}
			return event
		}

		switch event.Key() {
		case tcell.KeyEnter:
			item, _ := middlePane.GetItemText(middlePane.GetCurrentItem())
			navigateOrShowFile(item)
			return nil
		case tcell.KeyLeft:
			goUpDirectory()
			return nil

		case tcell.KeyRight:
			item, _ := middlePane.GetItemText(middlePane.GetCurrentItem())
			newPath := filepath.Join(currentDir, getCleanedItemName(item))
			fileInfo, err := os.Stat(newPath)
			if err != nil {
				log.Printf("Error stating file: %v", err)
				return event
			}

			if fileInfo.IsDir() {
				currentDir = newPath
				updatePanes()
				app.SetFocus(middlePane)
				return nil
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				middlePane.SetCurrentItem(middlePane.GetCurrentItem() + 1)
				return nil
			case 'k':
				middlePane.SetCurrentItem(middlePane.GetCurrentItem() - 1)
				return nil
			case 'h':
				goUpDirectory()
				return nil
			case 'l':
				item, _ := middlePane.GetItemText(middlePane.GetCurrentItem())
				navigateOrShowFile(item)
				return nil
			case 'a':
				addNewFile()
				return nil
			case 'd':
				item, _ := middlePane.GetItemText(middlePane.GetCurrentItem())
				deleteFile(item)
				return nil

			}
		}
		return event
	})
}
