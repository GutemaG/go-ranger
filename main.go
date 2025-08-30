package main

import (
	"log"

	app "github.com/GutemaG/go-ranger/pkg"
)

func main() {
	app := app.NewApp()
	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
