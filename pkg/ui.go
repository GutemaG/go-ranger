package pkg

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/rivo/tview"
)

// set up ui
func (a *App) setupUI() {
	// create left pane
	a.leftPane = tview.NewList().ShowSecondaryText(false)
	a.leftPane.SetBorder(true).SetTitle("Parent")

	// create middle pane
	a.middlePane = tview.NewList().ShowSecondaryText(false)
	a.middlePane.SetBorder(true).SetTitle("Current")

	// create right pane
	a.rightPane = tview.NewTextView()
	a.rightPane.SetBorder(true).SetTitle("Preview")

	// create flex from left,middle and right pane
	mainFlex := tview.NewFlex().
		AddItem(a.leftPane, 0, 1, false).
		AddItem(a.middlePane, 0, 3, false).
		AddItem(a.rightPane, 0, 3, false)

	// Create Grid
	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		AddItem(mainFlex, 1, 0, 1, 1, 0, 0, false)
	// create bottom bar
	a.bottomBarLeft = tview.NewTextView()
	a.bottomBarRight = tview.NewTextView()
	a.bottomBar = tview.NewFlex().
		AddItem(a.bottomBarLeft, 0, 2, false).
		AddItem(a.bottomBarRight, 0, 2, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, false).
		AddItem(a.bottomBar, 1, 0, false)
	a.pages.AddPage("main", flex, true, true)
	a.pages.SetBorder(true)
	a.tviewApp.SetFocus(a.middlePane)
}

func (a *App) updateLeftPane() {
	a.leftPane.Clear()
	parentDir := filepath.Dir(a.currentDir)

	if parentDir != a.currentDir { // go upup one dir
		a.leftPane.AddItem("..", "", 'h', a.goUpDirectory)
	} else {
		a.leftPane.AddItem(".", "", 0, nil)
	}

	dirs, _, err := GetEntries(parentDir)
	if err != nil {
		return
	}
	for _, entry := range dirs {

		if entry.Name() == filepath.Base(a.currentDir) {
			a.leftPane.AddItem(fmt.Sprintf("[yellow]%s/[white]", entry.Name()), "", 0, nil)
		} else {
			a.leftPane.AddItem(fmt.Sprintf("[darkcyan]%s/[white]", entry.Name()), "", 0, nil)
		}
	}

}

func (a *App) updateMiddlePane() {
	a.middlePane.Clear()
	directories, files, err := GetEntries(a.currentDir)

	if err != nil {
		log.Printf("Error reading directory %s: %v", a.currentDir, err)
		return
	}
	// a.middlePane.SetCurrentItem(a.selectedPaths[a.currentDir])

	if len(directories)+len(files) == 0 {
		a.middlePane.SetTitle("[red]Current (Empty)[/]")
		return
	}

	for _, entry := range directories {
		title := fmt.Sprintf("[darkcyan]%s/[white]", entry.Name())
		// title := fmt.Sprintf("[darkcyan]%s/[white] %v %v", entry.Name(), entry.Count, convertFileSize(entry.Size))
		a.middlePane.AddItem(title, "", 0, nil)
	}

	for _, entry := range files {
		a.middlePane.AddItem(entry.Name(), "", 0, nil)
	}

	a.middlePane.SetTitle(fmt.Sprintf("Current (%d items)", len(directories)+len(files)))
	a.pages.SetTitle(a.currentDir)

	if a.middlePane.GetItemCount() > 0 {
		a.middlePane.SetCurrentItem(0)
	}
}

func (a *App) updatePanes() {
	a.rightPane.SetText("")

	// update left panes
	a.updateLeftPane()
	// update middle pane
	a.updateMiddlePane()

	a.updateBottomBarInfo()

	a.middlePane.SetCurrentItem(a.selectedPaths[a.currentDir])

}

func (a *App) updateBottomBarInfo() {
	if a.middlePane.GetItemCount() == 0 {
		a.bottomBarLeft.SetText("")
		a.bottomBarRight.SetText("")
		return
	}
	item, _ := a.middlePane.GetItemText(a.middlePane.GetCurrentItem())
	cleanedItem := GetCleanedItemName(item)
	newPath := filepath.Join(a.currentDir, cleanedItem)
	fileInfo, err := os.Stat(newPath)
	if err != nil {
		return
	}

	stat, _ := fileInfo.Sys().(*syscall.Stat_t)
	ownerUser, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
	fileOwner := ""
	if err == nil {
		fileOwner = ownerUser.Username
	}

	file_mode := fileInfo.Mode().String()
	modified_time := fileInfo.ModTime().Format("2006-01-02 15:04:05")
	a.bottomBarLeft.SetText(file_mode + " " + fileOwner + " " + modified_time + "(" + fileInfo.Name() + ")")

	diskInfo, _ := GetDiskInfo("/")
	total_items := strconv.Itoa(a.middlePane.GetItemCount())
	currentIndex := strconv.Itoa(a.middlePane.GetCurrentItem() + 1)
	a.bottomBarRight.SetText(convertFileSize(fileInfo.Size()) + ", " + diskInfo["free"] + " free, (" + currentIndex + "/" + total_items + ")")
}
