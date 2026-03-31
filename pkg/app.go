package main

import (
	"log"

	"github.com/jroimartin/gocui"
)

func Run() {
	// Initialize GUI
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	client := GetClient()
	databases, err := GetDatabases(client)
	if err != nil {
		log.Fatalf("failed to fetch databases: %v", err)
	}

	SetDatabase(databases)

	// Set layout manager
	g.SetManagerFunc(func(g *gocui.Gui) error {
		return Layout(g)
	})
	g.Cursor = true

	// Set key bindings
	SetKeyBindings(g)

	// Start main loop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
