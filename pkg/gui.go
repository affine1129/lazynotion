package main

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/mattn/go-runewidth"
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

	// treeRight is the x-coordinate of the tree pane's right border, which
	// is also the divider between the two panes.
	treeRight := maxX / 3

	// treeWidth is the number of usable terminal columns inside the tree
	// pane's frame (gocui Size() = x1 - x0 - 1).
	treeWidth := treeRight - 1

	// Left pane: tree
	v, err := g.SetView("tree", 0, 0, treeRight, maxY-1)
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

		p, err := g.SetView("preview", treeRight+1, 0, maxX-1, maxY-1)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
		p.Clear()
		p.Title = "Preview"
		p.Editable = false
		p.Write([]byte("No pages to preview."))
		g.SetCurrentView("tree")
		return nil
	}

	// nameWidth is the max visual width for a DB/page name: prefix is always
	// 2 ASCII chars ("+ ", "- ", "  "), leaving the rest for the name.
	nameWidth := treeWidth - 2
	if nameWidth < 1 {
		nameWidth = 1
	}
	for _, node := range treeNodes {
		if node.IsDB {
			db := d[node.DBIdx]
			prefix := "+ "
			if !db.Collapsed {
				prefix = "- "
			}
			fmt.Fprintf(v, "%s%s\n", prefix, padWideRunes(truncateName(db.Name, nameWidth)))
		} else {
			fmt.Fprintf(v, "  %s\n", padWideRunes(truncateName(d[node.DBIdx].Pages[node.PageIdx].Name, nameWidth)))
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
	p, err := g.SetView("preview", treeRight+1, 0, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
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
			p.Write([]byte(padWideRunes(page.Content)))
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

// truncateName shortens name so its visual width does not exceed maxWidth
// terminal columns, accounting for double-width Unicode characters.
func truncateName(name string, maxWidth int) string {
	if runewidth.StringWidth(name) <= maxWidth {
		return name
	}
	width := 0
	for i, r := range name {
		w := runewidth.RuneWidth(r)
		if width+w > maxWidth {
			return name[:i]
		}
		width += w
	}
	return name
}

// padWideRunes inserts a space after each double-width rune (CJK characters)
// so that gocui's single-cell-per-rune renderer allocates two cells per wide
// character, matching the two terminal columns the character actually occupies.
// The inserted spaces are consumed as "continuation cells" by termbox's flush
// loop and are never visible on screen.
func padWideRunes(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(r)
		if runewidth.RuneWidth(r) == 2 {
			b.WriteByte(' ')
		}
	}
	return b.String()
}
