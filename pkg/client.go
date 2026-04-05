package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jomei/notionapi"
)

const (
	notionAPIURL     = "https://api.notion.com/v1"
	notionAPIVersion = "2022-06-28"
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
		{Name: "Introduction", Blocks: []Block{mockBlocks[0], mockBlocks[1]}, Content: "This is the introduction page.", ContentLoaded: true},
		{Name: "Details", Blocks: []Block{mockBlocks[2], mockBlocks[3]}, ContentLoaded: true},
		{Name: "Conclusion", Blocks: []Block{mockBlocks[4]}, ContentLoaded: true},
	}

	mockDBs = []Database{
		{
			Name:        "Sample DB One",
			Pages:       []Page{mockPages[0], mockPages[1]},
			Collapsed:   false,
			PagesLoaded: true,
		},
		{
			Name:        "Sample DB Two",
			Pages:       []Page{mockPages[1], mockPages[2]},
			Collapsed:   true,
			PagesLoaded: true,
		},
		{
			Name:        "Sample DB Three",
			Pages:       []Page{mockPages[0], mockPages[2]},
			Collapsed:   true,
			PagesLoaded: true,
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

// GetDatabases fetches the database list visible to the integration. Page and
// block content are loaded lazily when the user expands a database or selects a
// page.
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
		dbs = append(dbs, Database{
			ID:          notionapi.DatabaseID(db.ID),
			Name:        title,
			Collapsed:   true,
			PagesLoaded: false,
		})
	}

	return dbs, nil
}

func FetchPages(client *notionapi.Client, dbID notionapi.DatabaseID) ([]Page, error) {
	if client == nil {
		return nil, nil
	}

	resp, err := client.Database.Query(context.Background(), dbID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query database %s: %w", dbID, err)
	}

	var pages []Page
	for _, r := range resp.Results {
		title := getPageTitle(r)
		pages = append(pages, Page{Name: title, ID: notionapi.PageID(r.ID)})
	}

	return pages, nil
}

func LoadPages(client *notionapi.Client, db *Database) error {
	if db == nil || db.PagesLoaded || client == nil {
		return nil
	}

	pages, err := FetchPages(client, db.ID)
	if err != nil {
		return err
	}
	db.Pages = pages
	db.PagesLoaded = true
	return nil
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

func LoadPageContent(client *notionapi.Client, page *Page) error {
	if page == nil || page.ContentLoaded || client == nil {
		return nil
	}

	blocks, content, err := FetchPageContent(client, page.ID)
	if err != nil {
		return err
	}

	page.Blocks = blocks
	page.Content = content
	page.ContentLoaded = true
	return nil
}

func FetchPageContent(client *notionapi.Client, pageID notionapi.PageID) ([]Block, string, error) {
	if client == nil {
		return nil, "", nil
	}

	blocks, err := getBlocks(client, pageID)
	if err != nil {
		return nil, "", err
	}

	content := ""
	for _, b := range blocks {
		content += b.Content + "\n"
	}

	return blocks, content, nil
}

func getBlocks(client *notionapi.Client, pageID notionapi.PageID) ([]Block, error) {
	md, err := fetchPageMarkdown(client.Token.String(), pageID)
	if err != nil {
		return nil, err
	}
	if md == "" {
		return nil, nil
	}
	return []Block{{Content: md}}, nil
}

// fetchPageMarkdown retrieves the content of a Notion page as Markdown using
// the GET /v1/pages/{page_id}/markdown endpoint.
func fetchPageMarkdown(token string, pageID notionapi.PageID) (string, error) {
	url := fmt.Sprintf("%s/pages/%s/markdown", notionAPIURL, pageID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", notionAPIVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve page markdown for %s: %w", pageID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("retrieve page markdown returned status %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Markdown string `json:"markdown"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode page markdown response: %w", err)
	}
	return result.Markdown, nil
}

// UpdatePageMarkdown replaces the content of a Notion page with the provided
// markdown string using the PATCH /v1/pages/{page_id}/markdown endpoint with
// the replace_content action.
func UpdatePageMarkdown(client *notionapi.Client, pageID notionapi.PageID, markdown string) error {
	if client == nil {
		return nil
	}

	type replaceContent struct {
		NewStr string `json:"new_str"`
	}
	type requestBody struct {
		Type           string         `json:"type"`
		ReplaceContent replaceContent `json:"replace_content"`
	}

	body, err := json.Marshal(requestBody{
		Type:           "replace_content",
		ReplaceContent: replaceContent{NewStr: markdown},
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/pages/%s/markdown", notionAPIURL, pageID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+client.Token.String())
	req.Header.Set("Notion-Version", notionAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update page markdown for %s: %w", pageID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update page markdown returned status %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
