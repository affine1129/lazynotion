package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var (
	selectedIndex     int  = 0
	inEditor          bool = false
	treeNodes         []TreeNode
	prevSelectedIndex int = -1 // tracks last selected index to reset preview scroll
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

	_, viewHeight := v.Size()
	origin := 0
	if selectedIndex >= viewHeight {
		origin = selectedIndex - viewHeight + 1
	}
	v.SetOrigin(0, origin)

	for i, node := range treeNodes {
		if node.IsDB {
			db := d[node.DBIdx]
			prefix := "+ "
			if !db.Collapsed {
				prefix = "- "
			}
			fmt.Fprintf(v, "%s%s (%d)\n", prefix, db.Name, len(db.Pages))
		} else {
			fmt.Fprintf(v, "  %s\n", d[node.DBIdx].Pages[node.PageIdx].Name)
		}
		// Move cursor to selectedIndex
		if i == selectedIndex {
			v.SetCursor(0, i-origin)
		}
	}

	// Right pane: preview
	p, err := g.SetView("preview", maxX/3+1, 0, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		p.Wrap = true // wrap long lines so content is not cut off
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
		// Reset scroll origin only when the selection changes
		if prevSelectedIndex != selectedIndex {
			p.SetOrigin(0, 0)
			prevSelectedIndex = selectedIndex
		}
		if node.IsDB {
			p.Write([]byte("<Database>: select a page and press Enter"))
		} else {
			page := d[node.DBIdx].Pages[node.PageIdx]
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

// savePreview saves edited content back to mock data
func savePreview(g *gocui.Gui, v *gocui.View) error {
	n := treeNodes[selectedIndex]
	mockDBs[n.DBIdx].Pages[n.PageIdx].Content = v.Buffer()
	inEditor = false
	g.SetCurrentView("tree")
	return nil
}
