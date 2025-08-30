package pkg

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
	selectedPaths  map[string]int
}

// creating new application

func (a *App) updateSelectedPaths() {

	split_paths := strings.Split(a.currentDir, "/")
	for i, path := range split_paths {
		current_path := strings.Join(split_paths[:i], "/")
		if i == 0 {
			current_path = "/"
			path = "home"
		}

		dirs, _, _ := GetEntries(current_path)

		var info []string
		for _, dir := range dirs {
			info = append(info, dir.Info.Name())
		}

		index := slices.Index(info, path)

		if index == -1 {
			continue
		}

		if path == "" {
			index = slices.Index(info, "home")
		}

		a.selectedPaths[current_path] = index
	}

}

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
	a.selectedPaths = make(map[string]int)
	a.updateSelectedPaths()

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
	a.tviewApp.EnableMouse(true)

	return a.tviewApp.Run()
}

func (a *App) goUpDirectory() {
	a.currentDir = filepath.Clean(filepath.Join(a.currentDir, ".."))
	a.updatePanes()
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
