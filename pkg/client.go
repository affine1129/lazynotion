package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jomei/notionapi"
)

var (
	mockBlocks = []Block{
		{ID: "1", Content: "This is the first block."},
		{ID: "2", Content: "This is the second block."},
		{ID: "3", Content: "This is the third block."},
		{ID: "4", Content: "This is the fourth block."},
		{ID: "5", Content: "This is the fifth block."},
	}

	mockPages = []Page{
		{Name: "Introduction", Blocks: []Block{mockBlocks[0], mockBlocks[1]}, Content: "This is the introduction page."},
		{Name: "Details", Blocks: []Block{mockBlocks[2], mockBlocks[3]}},
		{Name: "Conclusion", Blocks: []Block{mockBlocks[4]}},
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
			Pages:     []Page{mockPages[0], mockPages[2]},
			Collapsed: true,
		},
	}
)

func GetClient() *notionapi.Client {
	// トークンを環境変数から取得
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		log.Fatalln("ERROR: NOTION_TOKEN is not set")
	}

	// クライアント初期化
	client := notionapi.NewClient(notionapi.Token(token))

	return client
}

func GetDatabases(client *notionapi.Client) ([]Database, error) {
	return mockDBs, nil
	// Search API で object=database を指定して検索
	resp, err := client.Search.Do(context.Background(), &notionapi.SearchRequest{
		Filter: notionapi.SearchFilter{
			Property: "object",   // 検索対象はオブジェクトの種類
			Value:    "database", // database のみを取得
		},
		PageSize: 100, // 一度に取る件数
	})
	if err != nil {
		return nil, err
	}

	// 4) 結果を走査し、*notionapi.Database 型だけを取り出す
	fmt.Println("Found databases:")
	var dbs []Database
	for _, r := range resp.Results {
		if db, ok := r.(*notionapi.Database); ok {
			// ID とタイトル（最初のリッチテキスト）を出力
			title := ""
			if len(db.Title) > 0 {
				title = db.Title[0].PlainText
			}
			pages, err := getPages(client, db.Parent.DatabaseID)
			if err != nil {
				log.Fatalf("failed to search pages: %v", err)
			}

			dbs = append(dbs, Database{Name: title, Pages: pages})
			fmt.Printf(" • %s — %s\n", db.ID, title)
		}
	}

	return dbs, nil
}

func getPages(client *notionapi.Client, dbID notionapi.DatabaseID) ([]Page, error) {
	// 6) データベースのページを取得
	resp, err := client.Database.Query(context.Background(), dbID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query database %s: %w", dbID, err)
	}

	var pages []Page
	for _, r := range resp.Results {
		if title, ok := r.Properties["Name"].(*notionapi.TitleProperty); ok {
			pages = append(pages, Page{ID: r.ID.String(), Name: title.Title[0].PlainText})
			// ここではページのブロックを取得する例
			blocks, err := getBlocks(client, notionapi.PageID(r.ID))
			if err != nil {
				return nil, fmt.Errorf("failed to get blocks for page %s: %w", r.ID, err)
			}
			pages[len(pages)-1].Blocks = blocks
			// ページのコンテンツを取得（必要に応じて）
			content := ""
			for _, block := range blocks {
				content += block.Content + "\n"
			}
			pages[len(pages)-1].Content = content
		}
	}
	return pages, nil
}

func getBlocks(client *notionapi.Client, pageID notionapi.PageID) ([]Block, error) {
	return nil, nil
	// 7) ページのブロックを取得
	// resp, err := client.Block.GetChildren(context.Background(), notionapi.BlockID(pageID), nil)
	// if err != nil {
	//   return nil, fmt.Errorf("failed to get blocks for page %s: %w", pageID, err)
	// }
	// var blocks []Block
	// for _, r := range resp.Results {
	//   if block, ok := r.(*notionapi.ParagraphBlock); ok {
	//     // ここでは ParagraphBlock のみを扱う例
	//     content := ""
	//     for _, text := range block.RichText {
	//       content += text.PlainText
	//     }
	//     blocks = append(blocks, Block{ID: block.ID, Content: content})
	//   }
	// }
	// return blocks, nil
}

// UpdatePageMarkdown replaces the content of a Notion page with the provided markdown string.
// It deletes all existing blocks on the page and appends new blocks parsed from the markdown.
// Returns an error if any API call fails.
func UpdatePageMarkdown(client *notionapi.Client, pageID string, markdown string) error {
	ctx := context.Background()
	bid := notionapi.BlockID(pageID)

	// Step 1: 既存ブロックをすべて取得して削除
	resp, err := client.Block.GetChildren(ctx, bid, nil)
	if err != nil {
		return fmt.Errorf("failed to get blocks for page %s: %w", pageID, err)
	}
	for _, block := range resp.Results {
		if _, err := client.Block.Delete(ctx, block.GetID()); err != nil {
			log.Printf("failed to delete block %s: %v", block.GetID(), err)
		}
	}

	// Step 2: Markdown をパースして Notion ブロックに変換
	blocks := markdownToBlocks(markdown)
	if len(blocks) == 0 {
		return nil
	}

	// Step 3: 新しいブロックをページに追加
	_, err = client.Block.AppendChildren(ctx, bid, &notionapi.AppendBlockChildrenRequest{
		Children: blocks,
	})
	if err != nil {
		return fmt.Errorf("failed to append blocks to page %s: %w", pageID, err)
	}

	return nil
}

// richText は与えられたテキストから RichText スライスを生成するヘルパー関数
func richText(content string) []notionapi.RichText {
	return []notionapi.RichText{
		{
			Type: notionapi.ObjectTypeText,
			Text: &notionapi.Text{Content: content},
		},
	}
}

// markdownToBlocks converts a markdown string to a slice of Notion Block objects.
// Supported syntax: headings (#, ##, ###), bullet lists (- or *), numbered lists (1.), and paragraphs.
func markdownToBlocks(markdown string) []notionapi.Block {
	var blocks []notionapi.Block
	lines := strings.Split(markdown, "\n")

	for _, line := range lines {
		// 空行はスキップ
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "### "):
			text := strings.TrimPrefix(trimmed, "### ")
			blocks = append(blocks, &notionapi.Heading3Block{
				BasicBlock: notionapi.BasicBlock{Type: notionapi.BlockTypeHeading3},
				Heading3:   notionapi.Heading{RichText: richText(text)},
			})
		case strings.HasPrefix(trimmed, "## "):
			text := strings.TrimPrefix(trimmed, "## ")
			blocks = append(blocks, &notionapi.Heading2Block{
				BasicBlock: notionapi.BasicBlock{Type: notionapi.BlockTypeHeading2},
				Heading2:   notionapi.Heading{RichText: richText(text)},
			})
		case strings.HasPrefix(trimmed, "# "):
			text := strings.TrimPrefix(trimmed, "# ")
			blocks = append(blocks, &notionapi.Heading1Block{
				BasicBlock: notionapi.BasicBlock{Type: notionapi.BlockTypeHeading1},
				Heading1:   notionapi.Heading{RichText: richText(text)},
			})
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			text := trimmed[2:]
			blocks = append(blocks, &notionapi.BulletedListItemBlock{
				BasicBlock:       notionapi.BasicBlock{Type: notionapi.BlockTypeBulletedListItem},
				BulletedListItem: notionapi.ListItem{RichText: richText(text)},
			})
		default:
			// 番号付きリスト: "数字. テキスト" の形式を検出
			if text, ok := parseNumberedListItem(trimmed); ok {
				blocks = append(blocks, &notionapi.NumberedListItemBlock{
					BasicBlock:       notionapi.BasicBlock{Type: notionapi.BlockTypeNumberedListItem},
					NumberedListItem: notionapi.ListItem{RichText: richText(text)},
				})
			} else {
				blocks = append(blocks, &notionapi.ParagraphBlock{
					BasicBlock: notionapi.BasicBlock{Type: notionapi.BlockTypeParagraph},
					Paragraph:  notionapi.Paragraph{RichText: richText(trimmed)},
				})
			}
		}
	}

	return blocks
}

// parseNumberedListItem detects lines of the form "N. text" (N = one or more digits)
// and returns the text portion and true. Otherwise returns "", false.
func parseNumberedListItem(s string) (string, bool) {
	dot := strings.IndexByte(s, '.')
	if dot <= 0 {
		return "", false
	}
	prefix := s[:dot]
	for _, r := range prefix {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	if len(s) <= dot+1 || s[dot+1] != ' ' {
		return "", false
	}
	return strings.TrimSpace(s[dot+2:]), true
}
