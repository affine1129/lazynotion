package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/jomei/notionapi"
	"github.com/jroimartin/gocui"
)

type Block struct {
	ID        string
	Content   string
	AddFlg    bool
	UpdateFlg bool
	DeleteFlg bool
}

type Page struct {
	Name  string
	Block string
}

type Database struct {
	Name      string
	Pages     []Page
	Collapsed bool
}

var (
	mockPages = []Page{
		{Name: "Introduction", Content: "Welcome to the mock page."},
		{Name: "Details", Content: "Here are some detailed contents."},
		{Name: "Conclusion", Content: "Summary and closing remarks."},
	}

	mockDBs = []Database{
		{
			Name:      "Sample DB One",
			Pages:     []Page{mockPages[0], mockPages[1]},
			Collapsed: false,
		},
		{
			Name:      "Sample DB Two",
			Pages:     []Page{mockPages[1], mockPages[2]},
			Collapsed: true,
		},
		{
			Name:      "Sample DB Three",
			Pages:     []Page{{Name: "Only Page", Content: "Single page content."}},
			Collapsed: true,
		},
	}
	selectedIndex int  = 0
	inEditor      bool = false
	treeNodes     []TreeNode
)

type TreeNode struct {
	DBIdx   int  // index in mockDBs
	IsDB    bool // true if this line is a DB, false if a page
	PageIdx int  // index in Pages slice (if IsDB is false)
}

func main() {
	// Initialize GUI
	// g, err := gocui.NewGui(gocui.OutputNormal)
	// if err != nil {
	//   log.Panicln(err)
	// }
	// defer g.Close()

	// // Set layout manager
	// g.SetManagerFunc(layout)

	// g.Cursor = true
	// rebuildTreeNodes()

	// // Keybinding: Ctrl+C to quit
	// if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
	//   log.Panicln(err)
	// }
	// if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
	//   log.Panicln(err)
	// }
	// if err := g.SetKeybinding("tree", 'j', gocui.ModNone, cursorDown); err != nil {
	//   log.Panicln(err)
	// }
	// if err := g.SetKeybinding("tree", 'k', gocui.ModNone, cursorUp); err != nil {
	//   log.Panicln(err)
	// }
	// if err := g.SetKeybinding("tree", gocui.KeyEnter, gocui.ModNone, toggleDB); err != nil {
	//   log.Panicln(err)
	// }
	// // edit page in preview
	// if err := g.SetKeybinding("tree", 'e', gocui.ModNone, toggleEdit); err != nil {
	//   log.Panicln(err)
	// }
	// // save in preview
	// if err := g.SetKeybinding("preview", gocui.KeyCtrlS, gocui.ModNone, savePreview); err != nil {
	//   log.Panicln(err)
	// }

	// Start main loop
	// if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
	//   log.Panicln(err)
	// }

	// Notion API

	// 1) トークンを環境変数から取得
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		log.Fatalln("ERROR: NOTION_TOKEN is not set")
	}

	// 2) クライアント初期化
	client := notionapi.NewClient(notionapi.Token(token))

	// 3) Search API で object=database を指定して検索
	resp, err := client.Search.Do(context.Background(), &notionapi.SearchRequest{
		// 空の Query でフィルタのみ適用
		Filter: notionapi.SearchFilter{
			Property: "object",   // 検索対象はオブジェクトの種類
			Value:    "database", // database のみを取得
		},
		PageSize: 100, // 一度に取る件数
	})
	if err != nil {
		log.Fatalf("failed to search databases: %v", err)
	}

	// 4) 結果を走査し、*notionapi.Database 型だけを取り出す
	fmt.Println("Found databases:")
	for _, r := range resp.Results {
		if db, ok := r.(*notionapi.Database); ok {
			// ID とタイトル（最初のリッチテキスト）を出力
			title := ""
			if len(db.Title) > 0 {
				title = db.Title[0].PlainText
			}
			fmt.Printf(" • %s — %s\n", db.ID, title)
		}
	}

}

// layout draws two panes with mock and static content
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Tree view
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
			db := mockDBs[node.DBIdx]
			prefix := "+ "
			if !db.Collapsed {
				prefix = "- "
			}
			fmt.Fprintf(v, "%s%s\n", prefix, db.Name)
		} else {
			fmt.Fprintf(v, "  %s\n", mockDBs[node.DBIdx].Pages[node.PageIdx].Name)
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
		page := mockDBs[node.DBIdx].Pages[node.PageIdx]
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

func rebuildTreeNodes() {
	treeNodes = treeNodes[:0]
	for dbIdx, db := range mockDBs {
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

// toggleEdit opens the selected page in Neovim for editing
func toggleEdit(g *gocui.Gui, v *gocui.View) error {
	n := treeNodes[selectedIndex]
	if n.IsDB {
		return nil
	}
	// get pointer to the page
	pg := &mockDBs[n.DBIdx].Pages[n.PageIdx]
	// write content to temporary file
	tmpfile, err := os.CreateTemp("", "page-*.md")
	if err != nil {
		return err
	}
	tmpname := tmpfile.Name()
	tmpfile.WriteString(pg.Content)
	tmpfile.Close()
	defer os.Remove(tmpname)

	// close GUI to return to terminal
	g.Close()

	// launch Neovim
	cmd := exec.Command("nvim", tmpname)
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

	g2.SetManagerFunc(layout)
	// read updated content
	data, err := os.ReadFile(tmpname)
	if err != nil {
		return err
	}
	pg.Content = string(data)

	g.SetCurrentView("tree")
	return nil
}

// savePreview saves edited content back to mock data
func savePreview(g *gocui.Gui, v *gocui.View) error {
	n := treeNodes[selectedIndex]
	mockDBs[n.DBIdx].Pages[n.PageIdx].Content = v.Buffer()
	inEditor = false
	g.SetCurrentView("tree")
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
		rebuildTreeNodes()
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

// quit handler
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
