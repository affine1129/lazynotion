package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var (
	selectedIndex  int = 0
	treeNodes      []TreeNode
	previewOriginY int = 0
)

type TreeNode struct {
	IsDB    bool // true if this line is a DB, false if a page
	DBIdx   int  // index in mockDBs
	PageIdx int  // index in Pages slice (if IsDB is false)
}

// layout draws two panes with mock and static content
func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	d := GetDatabase()

	RebuildTreeNodes()

	// Left pane: tree
	v, err := g.SetView("tree", 0, 0, maxX/3, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true              // enable highlight
		v.SelFgColor = gocui.ColorGreen // cursor color
	}
	v.Clear()
	v.Title = "Databases"

	if len(treeNodes) == 0 {
		fmt.Fprintln(v, "No databases found.")
		fmt.Fprintln(v, "Share at least one database with your Notion integration.")

		p, err := g.SetView("preview", maxX/3+1, 0, maxX-1, maxY-1)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			p.Wrap = true
		}
		p.Clear()
		p.Title = "Preview"
		p.Editable = false
		p.Write([]byte("No pages to preview."))
		g.SetCurrentView("tree")
		return nil
	}

	for _, node := range treeNodes {
		if node.IsDB {
			db := d[node.DBIdx]
			prefix := "+ "
			if !db.Collapsed {
				prefix = "- "
			}
			fmt.Fprintf(v, "%s%s\n", prefix, db.Name)
		} else {
			fmt.Fprintf(v, "  %s\n", d[node.DBIdx].Pages[node.PageIdx].Name)
		}
	}

	// Scroll the view so that selectedIndex is always visible, then place
	// the cursor at the correct position within the visible area.
	_, viewHeight := v.Size()
	_, oy := v.Origin()
	if selectedIndex < oy {
		oy = selectedIndex
	} else if selectedIndex >= oy+viewHeight {
		oy = selectedIndex - viewHeight + 1
	}
	v.SetOrigin(0, oy)
	v.SetCursor(0, selectedIndex-oy)

	// Right pane: preview
	p, err := g.SetView("preview", maxX/3+1, 0, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		p.Wrap = true
	}
	p.Title = "Preview"
	node := treeNodes[selectedIndex]
	p.Clear()
	p.Editable = false
	if node.IsDB {
		p.Write([]byte("<Database>: select a page and press Enter"))
	} else {
		page := d[node.DBIdx].Pages[node.PageIdx]
		if page.ContentLoaded {
			p.Write([]byte(page.Content))
		} else {
			p.Write([]byte("Loading page content..."))
		}
	}
	if previewOriginY > previewMaxOriginY(p) {
		previewOriginY = previewMaxOriginY(p)
	}
	p.SetOrigin(0, previewOriginY)
	g.SetCurrentView("tree")

	return nil
}

func RebuildTreeNodes() {
	d := GetDatabase()
	treeNodes = treeNodes[:0]
	for dbIdx, db := range d {
		treeNodes = append(treeNodes, TreeNode{DBIdx: dbIdx, IsDB: true})
		if !db.Collapsed {
			for pageIdx := range db.Pages {
				treeNodes = append(treeNodes, TreeNode{DBIdx: dbIdx, IsDB: false, PageIdx: pageIdx})
			}
		}
	}
	if len(treeNodes) == 0 {
		selectedIndex = 0
		return
	}
	if selectedIndex < 0 {
		selectedIndex = 0
	}
	if selectedIndex >= len(treeNodes) {
		selectedIndex = len(treeNodes) - 1
	}
}
