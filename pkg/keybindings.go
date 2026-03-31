package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/jroimartin/gocui"
)

func SetKeyBindings(g *gocui.Gui) error {
	// Keybinding: Ctrl+C to quit
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("tree", 'j', gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("tree", 'k', gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("tree", gocui.KeyEnter, gocui.ModNone, toggleDB); err != nil {
		log.Panicln(err)
	}
	// edit page in preview
	if err := g.SetKeybinding("tree", 'e', gocui.ModNone, toggleEdit); err != nil {
		log.Panicln(err)
	}
	// save in preview
	if err := g.SetKeybinding("preview", gocui.KeyCtrlS, gocui.ModNone, savePreview); err != nil {
		log.Panicln(err)
	}
	// scroll preview from tree view using arrow keys
	if err := g.SetKeybinding("tree", gocui.KeyArrowDown, gocui.ModNone, scrollPreviewDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("tree", gocui.KeyArrowUp, gocui.ModNone, scrollPreviewUp); err != nil {
		log.Panicln(err)
	}
	return nil
}

// toggleDB toggles the collapsed state of a Database node
func toggleDB(g *gocui.Gui, v *gocui.View) error {
	if selectedIndex < 0 || selectedIndex >= len(treeNodes) {
		return nil
	}
	node := treeNodes[selectedIndex]
	if node.IsDB {
		mockDBs[node.DBIdx].Collapsed = !mockDBs[node.DBIdx].Collapsed
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if selectedIndex < len(treeNodes)-1 {
		selectedIndex++
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if selectedIndex > 0 {
		selectedIndex--
	}
	return nil
}

// scrollPreviewDown scrolls the preview pane down by one line.
func scrollPreviewDown(g *gocui.Gui, v *gocui.View) error {
	p, err := g.View("preview")
	if err != nil {
		return err
	}
	_, oy := p.Origin()
	_, ph := p.Size()
	if oy+ph < len(p.BufferLines()) {
		p.SetOrigin(0, oy+1)
	}
	return nil
}

// scrollPreviewUp scrolls the preview pane up by one line.
func scrollPreviewUp(g *gocui.Gui, v *gocui.View) error {
	p, err := g.View("preview")
	if err != nil {
		return err
	}
	_, oy := p.Origin()
	if oy > 0 {
		p.SetOrigin(0, oy-1)
	}
	return nil
}

// quit handler
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// toggleEdit opens the selected page in Neovim for editing
func toggleEdit(g *gocui.Gui, v *gocui.View) error {
	n := treeNodes[selectedIndex]
	d := GetDatabase()

	if n.IsDB {
		return nil
	}
	// get pointer to the page
	pg := &d[n.DBIdx].Pages[n.PageIdx]
	// write content to temporary file
	tmpfile, err := os.CreateTemp("", "page-*.md")
	if err != nil {
		return err
	}
	tmpName := tmpfile.Name()
	tmpfile.WriteString(pg.Content)
	tmpfile.Close()
	defer os.Remove(tmpName)

	// close GUI to return to terminal
	g.Close()

	// launch Neovim
	cmd := exec.Command("nvim", tmpName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// after edit, reopen GUI
	g2, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	g2.SetManagerFunc(Layout)
	// read updated content
	data, err := os.ReadFile(tmpName)
	if err != nil {
		return err
	}
	pg.Content = string(data)

	g.SetCurrentView("tree")
	return nil
}
