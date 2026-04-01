package main

import (
	"fmt"
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
	// scroll preview pane
	if err := g.SetKeybinding("tree", 'J', gocui.ModNone, scrollPreviewDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("tree", 'K', gocui.ModNone, scrollPreviewUp); err != nil {
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
		d := GetDatabase()
		db := &d[node.DBIdx]
		if db.Collapsed && !db.PagesLoaded {
			if err := LoadPages(GetClient(), db); err != nil {
				return err
			}
		}
		db.Collapsed = !db.Collapsed
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if selectedIndex < len(treeNodes)-1 {
		selectedIndex++
		previewOriginY = 0
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if selectedIndex > 0 {
		selectedIndex--
		previewOriginY = 0
	}
	return nil
}

// quit handler
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// previewMaxOriginY returns the maximum scroll offset for the preview pane.
func previewMaxOriginY(p *gocui.View) int {
	_, pvHeight := p.Size()
	totalLines := len(p.BufferLines())
	max := totalLines - pvHeight
	if max < 0 {
		return 0
	}
	return max
}

func scrollPreviewDown(g *gocui.Gui, v *gocui.View) error {
	p, err := g.View("preview")
	if err != nil {
		return nil
	}
	if previewOriginY < previewMaxOriginY(p) {
		previewOriginY++
		p.SetOrigin(0, previewOriginY)
	}
	return nil
}

func scrollPreviewUp(g *gocui.Gui, v *gocui.View) error {
	if previewOriginY > 0 {
		previewOriginY--
		p, err := g.View("preview")
		if err != nil {
			return nil
		}
		p.SetOrigin(0, previewOriginY)
	}
	return nil
}

func pickEditor() (string, error) {
	if _, err := exec.LookPath("nvim"); err == nil {
		return "nvim", nil
	}
	if _, err := exec.LookPath("vim"); err == nil {
		return "vim", nil
	}
	return "", fmt.Errorf("no supported editor found: install nvim or vim")
}

// toggleEdit opens the selected page in the configured editor.
func toggleEdit(g *gocui.Gui, v *gocui.View) error {
	if len(treeNodes) == 0 || selectedIndex < 0 || selectedIndex >= len(treeNodes) {
		return nil
	}
	n := treeNodes[selectedIndex]
	d := GetDatabase()

	if n.IsDB {
		return nil
	}
	if err := LoadPageContent(GetClient(), &d[n.DBIdx].Pages[n.PageIdx]); err != nil {
		return err
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

	editor, err := pickEditor()
	if err != nil {
		return err
	}

	// launch editor
	cmd := exec.Command(editor, tmpName)
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
