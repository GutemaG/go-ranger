package main

import (
	"log"
	"os"

	"github.com/rivo/tview"
)

type BottomBarInfo struct {
	LeftInfo  string
	RightInfo string
}

// The current working directory.
var currentDir string

// UI components
var app *tview.Application
var pages *tview.Pages
var leftPane *tview.List
var middlePane *tview.List
var rightPane *tview.TextView
var mainFlex *tview.Flex
var bottomBar *tview.Flex
var bottomBarLeft *tview.TextView
var bottomBarRight *tview.TextView

// Main function to initialize and run the application.
func main() {
	// Initialize UI components and their event handlers.
	setupUI()

	var err error
	currentDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Update the UI with the initial directory contents.
	updatePanes()

	// Set the root component and run the application.
	app.SetRoot(pages, true).SetFocus(middlePane)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
