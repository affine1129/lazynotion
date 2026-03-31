package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var (
	selectedIndex int  = 0
	inEditor      bool = false
	treeNodes     []TreeNode
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

	for i, node := range treeNodes {
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
		// Move cursor to selectedIndex
		if i == selectedIndex {
			v.SetCursor(0, i)
		}
	}

	// Right pane: preview
	p, err := g.SetView("preview", maxX/3+1, 0, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	p.Title = "Preview"
	node := treeNodes[selectedIndex]
	if inEditor {
		p.Editable = true
		p.Editor = gocui.DefaultEditor
		g.SetCurrentView("preview")
	} else {
		p.Clear()
		p.Editable = false
		page := d[node.DBIdx].Pages[node.PageIdx]
		if node.IsDB {
			p.Write([]byte("<Database>: select a page and press Enter"))
		} else {
			p.Write([]byte(page.Content))
		}
		g.SetCurrentView("tree")
	}

	if !inEditor {
		g.SetCurrentView("tree")
	}

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
	if selectedIndex >= len(treeNodes) {
		selectedIndex = len(treeNodes) - 1
	}
}

// savePreview saves edited content back to the in-memory database state.
func savePreview(g *gocui.Gui, v *gocui.View) error {
	n := treeNodes[selectedIndex]
	databases[n.DBIdx].Pages[n.PageIdx].Content = v.Buffer()
	inEditor = false
	g.SetCurrentView("tree")
	return nil
}
