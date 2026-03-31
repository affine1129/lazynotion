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

	// treeWidth and previewWidth are the number of usable terminal columns
	// inside each pane's frame (gocui Size() = x1 - x0 - 1).
	treeWidth := treeRight - 1
	previewWidth := maxX - treeRight - 3

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
			fmt.Fprintf(v, "%s%s\n", prefix, truncateName(db.Name, nameWidth))
		} else {
			fmt.Fprintf(v, "  %s\n", truncateName(d[node.DBIdx].Pages[node.PageIdx].Name, nameWidth))
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
			p.Write([]byte(wrapText(page.Content, previewWidth)))
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

// wrapText splits text into lines that fit within width terminal columns,
// accounting for double-width Unicode characters (e.g. Japanese/Chinese).
// Existing newlines are preserved; each logical line is wrapped independently.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, wrapLine(line, width)...)
	}
	return strings.Join(out, "\n")
}

// wrapLine splits a single line into segments each no wider than width columns.
func wrapLine(line string, width int) []string {
	if runewidth.StringWidth(line) <= width {
		return []string{line}
	}
	var result []string
	colWidth := 0
	lineStart := 0
	for i, r := range line {
		w := runewidth.RuneWidth(r)
		// Only wrap if the current line is non-empty; this avoids producing an
		// empty leading segment when the very first character exceeds width.
		if colWidth > 0 && colWidth+w > width {
			result = append(result, line[lineStart:i])
			lineStart = i
			colWidth = w
		} else {
			colWidth += w
		}
	}
	result = append(result, line[lineStart:])
	return result
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
