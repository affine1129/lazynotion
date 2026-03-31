package main

import (
	"log"

	"github.com/jomei/notionapi"
	"github.com/jroimartin/gocui"
)

func Run() {
	if err := loadEnv(); err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}

	client := GetClient()
	databases, err := GetDatabases(client)
	if err != nil {
		log.Fatalf("failed to fetch databases: %v", err)
	}

	SetDatabase(databases)

	// Initialize GUI
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	// Set layout manager
	g.SetManagerFunc(func(g *gocui.Gui) error {
		return Layout(g)
	})
	g.Cursor = true

	// Set key bindings
	SetKeyBindings(g)

	startBackgroundPrefetch(g, client)

	// Start main loop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func startBackgroundPrefetch(g *gocui.Gui, client *notionapi.Client) {
	if client == nil {
		return
	}

	type dbTarget struct {
		index int
		id    notionapi.DatabaseID
	}

	d := GetDatabase()
	targets := make([]dbTarget, 0, len(d))
	for index, db := range d {
		targets = append(targets, dbTarget{index: index, id: db.ID})
	}

	go func() {
		for _, target := range targets {
			pages, err := FetchPages(client, target.id)
			if err != nil {
				log.Printf("background prefetch pages failed for %s: %v", target.id, err)
				continue
			}

			g.Update(func(*gocui.Gui) error {
				d := GetDatabase()
				if target.index >= len(d) {
					return nil
				}
				db := &d[target.index]
				if !db.PagesLoaded {
					db.Pages = pages
					db.PagesLoaded = true
				}
				return nil
			})

			for pageIndex, page := range pages {
				blocks, content, err := FetchPageContent(client, page.ID)
				if err != nil {
					log.Printf("background prefetch page failed for %s: %v", page.ID, err)
					continue
				}

				pageIndex := pageIndex
				g.Update(func(*gocui.Gui) error {
					d := GetDatabase()
					if target.index >= len(d) {
						return nil
					}
					db := &d[target.index]
					if pageIndex >= len(db.Pages) {
						return nil
					}
					prefetchedPage := &db.Pages[pageIndex]
					if prefetchedPage.ContentLoaded {
						return nil
					}
					prefetchedPage.Blocks = blocks
					prefetchedPage.Content = content
					prefetchedPage.ContentLoaded = true
					return nil
				})
			}
		}
	}()
}
