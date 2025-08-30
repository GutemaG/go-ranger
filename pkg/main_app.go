package pkg

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

type App struct {
	tviewApp       *tview.Application
	pages          *tview.Pages
	leftPane       *tview.List
	middlePane     *tview.List
	rightPane      *tview.TextView
	bottomBar      *tview.Flex
	bottomBarLeft  *tview.TextView
	bottomBarRight *tview.TextView
	currentDir     string
}

// creating new application

func NewApp() *App {
	a := &App{
		tviewApp: tview.NewApplication(),
		pages:    tview.NewPages(),
	}
	var err error
	a.currentDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// ui setup
	a.setupUI()
	// set event handlers
	a.setEventHandlers()

	// update pens
	a.updatePanes()

	return a
}

func (a *App) Run() error {
	a.tviewApp.SetRoot(a.pages, true)
	return a.tviewApp.Run()
}

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

func (a *App) goUpDirectory() {
	a.currentDir = filepath.Clean(filepath.Join(a.currentDir, ".."))
	a.updatePanes()
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
	// update left panes
	a.updateLeftPane()
	// update middle pane
	a.updateMiddlePane()

	a.updateBottomBarInfo()
	a.rightPane.SetText("")
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
	a.bottomBarLeft.SetText(file_mode + " " + fileOwner + " " + modified_time)

	diskInfo, _ := GetDiskInfo("/")
	total_items := strconv.Itoa(a.middlePane.GetItemCount())
	currentIndex := strconv.Itoa(a.middlePane.GetCurrentItem() + 1)
	a.bottomBarRight.SetText(convertFileSize(fileInfo.Size()) + ", " + diskInfo["free"] + " free, (" + currentIndex + "/" + total_items + ")")
}

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

		switch event.Rune() {
		case 'j':
			// tview handles this, but we could override
			a.middlePane.SetCurrentItem(a.middlePane.GetCurrentItem() + 1)
			return event
		case 'k':
			a.middlePane.SetCurrentItem(a.middlePane.GetCurrentItem() - 1)
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
		return // Can't stat, do nothing
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

// Function to get the item name without TUI color tags or trailing slashes
func getCleanedItemName(item string) string {
	cleanedItem := strings.TrimSuffix(item, "/")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[darkcyan]", "")
	cleanedItem = strings.ReplaceAll(cleanedItem, "[white]", "")
	return strings.TrimSpace(cleanedItem)
}

// navigateOrShowFile navigates into a directory or shows file content.
func (a *App) navigateOrShowFile(item string) {
	cleanedItem := getCleanedItemName(item)
	newPath := filepath.Join(a.currentDir, cleanedItem)
	fileInfo, err := os.Stat(newPath)

	if err != nil {
		log.Printf("Error stating file: %v", err)
		return
	}

	if fileInfo.IsDir() {
		a.currentDir = newPath
		a.updatePanes()
	} else {
		content, err := os.ReadFile(newPath)
		if err != nil {
			modal := tview.NewModal().
				SetText("Error reading file: " + err.Error()).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.SwitchToPage("main")
					a.tviewApp.SetFocus(a.middlePane)
				})
			a.pages.AddPage("fileViewer", modal, true, true)
			return
		}

		modal := tview.NewModal().
			SetText(string(content)).
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.pages.SwitchToPage("main")
				a.tviewApp.SetFocus(a.middlePane)
			})

		modal.SetTitle("Viewing: " + cleanedItem)
		a.pages.AddPage("fileViewer", modal, true, true)
		a.tviewApp.SetFocus(a.pages)
	}
}

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
