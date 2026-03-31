package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/affine1129/lazynotion/pkg/convert"
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

// GetClient returns a Notion API client using the NOTION_TOKEN environment
// variable. If the token is not set it returns nil; callers treat nil as
// "use mock data".
func GetClient() *notionapi.Client {
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		return nil
	}
	return notionapi.NewClient(notionapi.Token(token))
}

// GetDatabases fetches all databases visible to the integration. When client
// is nil (no token configured) it falls back to mock data so the app remains
// usable during development.
func GetDatabases(client *notionapi.Client) ([]Database, error) {
	if client == nil {
		log.Println("NOTION_TOKEN not set – using mock data")
		return mockDBs, nil
	}

	resp, err := client.Search.Do(context.Background(), &notionapi.SearchRequest{
		Filter: notionapi.SearchFilter{
			Property: "object",
			Value:    "database",
		},
		PageSize: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var dbs []Database
	for _, r := range resp.Results {
		db, ok := r.(*notionapi.Database)
		if !ok {
			continue
		}
		title := ""
		if len(db.Title) > 0 {
			title = db.Title[0].PlainText
		}
		pages, err := getPages(client, notionapi.DatabaseID(db.ID))
		if err != nil {
			return nil, fmt.Errorf("failed to get pages for database %q: %w", title, err)
		}
		dbs = append(dbs, Database{Name: title, Pages: pages})
	}

	return dbs, nil
}

func getPages(client *notionapi.Client, dbID notionapi.DatabaseID) ([]Page, error) {
	resp, err := client.Database.Query(context.Background(), dbID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query database %s: %w", dbID, err)
	}

	var pages []Page
	for _, r := range resp.Results {
		title := getPageTitle(r)
		blocks, err := getBlocks(client, notionapi.PageID(r.ID))
		if err != nil {
			return nil, fmt.Errorf("failed to get blocks for page %s: %w", r.ID, err)
		}
		content := ""
		for _, b := range blocks {
			content += b.Content + "\n"
		}
		pages = append(pages, Page{Name: title, Blocks: blocks, Content: content})
	}
	return pages, nil
}

// getPageTitle finds the title-type property of a Notion page regardless of
// its property name (the default is "Name" but workspaces can rename it).
func getPageTitle(p notionapi.Page) string {
	for _, prop := range p.Properties {
		if titleProp, ok := prop.(*notionapi.TitleProperty); ok {
			if len(titleProp.Title) > 0 {
				return titleProp.Title[0].PlainText
			}
		}
	}
	return "Untitled"
}

func getBlocks(client *notionapi.Client, pageID notionapi.PageID) ([]Block, error) {
	resp, err := client.Block.GetChildren(context.Background(), notionapi.BlockID(pageID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks for page %s: %w", pageID, err)
	}

	md := convert.BlocksToMarkdown(resp.Results)
	if md == "" {
		return nil, nil
	}
	return []Block{{Content: md}}, nil
}
