package main

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
	Name   string
	Blocks []Block
	// for preview
	Content string
}

func NewPage(name string, blocks []Block) Page {
	return Page{
		Name:   name,
		Blocks: blocks,
	}
}

type Database struct {
	Name      string
	Pages     []Page
	Collapsed bool
}

func NewDatabase(name string, pages []Page, collapsed bool) Database {
	return Database{
		Name:      name,
		Pages:     pages,
		Collapsed: collapsed,
	}
}

func SetDatabase(db []Database) {
	databases = db
}

func GetDatabase() []Database {
	return databases
}
