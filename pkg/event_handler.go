package pkg

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

func (a *App) setEventHandlers() {

	a.middlePane.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		a.preview()
	})

	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// exit by Ctr-c or q
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			a.tviewApp.Stop()
		}
		return event
	})

	a.middlePane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle navigation and actions for empty directory
		if a.middlePane.GetItemCount() == 0 {
			if event.Key() == tcell.KeyLeft || event.Rune() == 'h' {
				a.goUpDirectory()
				return nil
			}
			// We can add other actions like 'a' for add here
			return event
		}
		a.selectedPaths[a.currentDir] = a.middlePane.GetCurrentItem()

		// Handle navigation for populated directory
		switch event.Key() {
		case tcell.KeyEnter:
			item, _ := a.middlePane.GetItemText(a.middlePane.GetCurrentItem())
			a.navigateOrShowFile(item)
			return nil
		case tcell.KeyLeft:
			a.goUpDirectory()
			return nil
		case tcell.KeyRight:
			item, _ := a.middlePane.GetItemText(a.middlePane.GetCurrentItem())
			newPath := filepath.Join(a.currentDir, getCleanedItemName(item))
			fileInfo, err := os.Stat(newPath)
			if err != nil {
				log.Printf("Error stating file: %v", err)
				return event
			}

			if fileInfo.IsDir() {
				a.currentDir = newPath
				a.updatePanes()
				a.tviewApp.SetFocus(a.middlePane)
				return nil
			}
		}

		index := a.middlePane.GetCurrentItem()
		switch event.Rune() {
		case 'j':
			a.middlePane.SetCurrentItem(index + 1)
			a.selectedPaths[a.currentDir] = index + 1
			a.preview()
			return event
		case 'k':
			a.middlePane.SetCurrentItem(index - 1)
			a.selectedPaths[a.currentDir] = index - 1
			a.preview()
			return event
		case 'h':
			a.goUpDirectory()
			return nil
		case 'l':
			item, _ := a.middlePane.GetItemText(a.middlePane.GetCurrentItem())
			a.navigateOrShowFile(item)
			return nil
		case 'a':
			a.addNewFile()
			return nil
		case 'd':
			item, _ := a.middlePane.GetItemText(a.middlePane.GetCurrentItem())
			a.deleteFile(item)
			return nil
		}

		return event
	})

}
