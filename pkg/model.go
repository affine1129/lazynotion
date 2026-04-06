package main

import "github.com/jomei/notionapi"

var databases []Database

type Block struct {
	ID        string
	Type      string
	Content   string
	Children  []Block // for nested blocks
	AddFlg    bool
	UpdateFlg bool
	DeleteFlg bool
}

func NewBlock(id, blockType, content string) Block {
	return Block{
		ID:      id,
		Type:    blockType,
		Content: content,
	}
}

type Page struct {
	ID     notionapi.PageID
	Name   string
	Blocks []Block
	// for preview
	Content       string
	ContentLoaded bool
}

func NewPage(name string, blocks []Block) Page {
	return Page{
		Name:          name,
		Blocks:        blocks,
		ContentLoaded: true,
	}
}

type Database struct {
	ID          notionapi.DatabaseID
	Name        string
	Pages       []Page
	Collapsed   bool
	PagesLoaded bool
	Loading     bool // true while pages are being fetched for the first time
}

func NewDatabase(name string, pages []Page, collapsed bool) Database {
	return Database{
		Name:        name,
		Pages:       pages,
		Collapsed:   collapsed,
		PagesLoaded: true,
	}
}

func SetDatabase(db []Database) {
	databases = db
}

func GetDatabase() []Database {
	return databases
}
