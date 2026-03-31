package convert_test

import (
	"testing"

	"github.com/affine1129/lazynotion/pkg/convert"
	"github.com/jomei/notionapi"
)

// richText is a helper that builds a plain RichText segment.
func richText(content string) notionapi.RichText {
	return notionapi.RichText{
		PlainText: content,
	}
}

// richTextBold builds a bold RichText segment.
func richTextBold(content string) notionapi.RichText {
	return notionapi.RichText{
		PlainText:   content,
		Annotations: &notionapi.Annotations{Bold: true},
	}
}

// richTextItalic builds an italic RichText segment.
func richTextItalic(content string) notionapi.RichText {
	return notionapi.RichText{
		PlainText:   content,
		Annotations: &notionapi.Annotations{Italic: true},
	}
}

// richTextCode builds an inline-code RichText segment.
func richTextCode(content string) notionapi.RichText {
	return notionapi.RichText{
		PlainText:   content,
		Annotations: &notionapi.Annotations{Code: true},
	}
}

// richTextStrikethrough builds a strikethrough RichText segment.
func richTextStrikethrough(content string) notionapi.RichText {
	return notionapi.RichText{
		PlainText:   content,
		Annotations: &notionapi.Annotations{Strikethrough: true},
	}
}

// richTextLink builds a linked RichText segment.
func richTextLink(content, href string) notionapi.RichText {
	return notionapi.RichText{
		PlainText: content,
		Href:      href,
	}
}

// ─── golden tests ─────────────────────────────────────────────────────────────

func TestParagraph(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richText("Hello, world!")},
			},
		},
	}
	want := "Hello, world!\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestHeadings(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.Heading1Block{
			Heading1: notionapi.Heading{
				RichText: []notionapi.RichText{richText("Title")},
			},
		},
		&notionapi.Heading2Block{
			Heading2: notionapi.Heading{
				RichText: []notionapi.RichText{richText("Section")},
			},
		},
		&notionapi.Heading3Block{
			Heading3: notionapi.Heading{
				RichText: []notionapi.RichText{richText("Subsection")},
			},
		},
	}
	want := "# Title\n\n## Section\n\n### Subsection\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestBulletedList(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.BulletedListItemBlock{
			BulletedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("Item A")},
			},
		},
		&notionapi.BulletedListItemBlock{
			BulletedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("Item B")},
			},
		},
	}
	want := "- Item A\n- Item B\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestNumberedList(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("First")},
			},
		},
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("Second")},
			},
		},
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("Third")},
			},
		},
	}
	want := "1. First\n2. Second\n3. Third\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestNumberedListCounterReset(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("First")},
			},
		},
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richText("Break")},
			},
		},
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("Restart")},
			},
		},
	}
	want := "1. First\nBreak\n\n1. Restart\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestToDo(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ToDoBlock{
			ToDo: notionapi.ToDo{
				RichText: []notionapi.RichText{richText("Unchecked task")},
				Checked:  false,
			},
		},
		&notionapi.ToDoBlock{
			ToDo: notionapi.ToDo{
				RichText: []notionapi.RichText{richText("Checked task")},
				Checked:  true,
			},
		},
	}
	want := "- [ ] Unchecked task\n- [x] Checked task\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestCode(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.CodeBlock{
			Code: notionapi.Code{
				Language: "go",
				RichText: []notionapi.RichText{richText("fmt.Println(\"hello\")")},
			},
		},
	}
	want := "```go\nfmt.Println(\"hello\")\n```\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestQuote(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.QuoteBlock{
			Quote: notionapi.Quote{
				RichText: []notionapi.RichText{richText("To be or not to be")},
			},
		},
	}
	want := "> To be or not to be\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestDivider(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.DividerBlock{},
	}
	want := "---\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestImageExternal(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ImageBlock{
			Image: notionapi.Image{
				Type: notionapi.FileTypeExternal,
				External: &notionapi.FileObject{
					URL: "https://example.com/image.png",
				},
				Caption: []notionapi.RichText{richText("A caption")},
			},
		},
	}
	want := "![A caption](https://example.com/image.png)\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestImageNoCaption(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ImageBlock{
			Image: notionapi.Image{
				Type: notionapi.FileTypeExternal,
				External: &notionapi.FileObject{
					URL: "https://example.com/photo.jpg",
				},
			},
		},
	}
	want := "![](https://example.com/photo.jpg)\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

// ─── rich text decoration tests ───────────────────────────────────────────────

func TestRichTextBold(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richTextBold("important")},
			},
		},
	}
	want := "**important**\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRichTextItalic(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richTextItalic("emphasis")},
			},
		},
	}
	want := "*emphasis*\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRichTextInlineCode(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richTextCode("os.Exit(1)")},
			},
		},
	}
	want := "`os.Exit(1)`\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRichTextStrikethrough(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richTextStrikethrough("deleted")},
			},
		},
	}
	want := "~~deleted~~\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRichTextLink(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					richTextLink("Notion", "https://notion.so"),
				},
			},
		},
	}
	want := "[Notion](https://notion.so)\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRichTextMixed(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					richText("Visit "),
					richTextLink("our site", "https://example.com"),
					richText(" for "),
					richTextBold("details"),
					richText("."),
				},
			},
		},
	}
	want := "Visit [our site](https://example.com) for **details**.\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

// ─── nesting tests ────────────────────────────────────────────────────────────

func TestNestedBulletedList(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.BulletedListItemBlock{
			BasicBlock: notionapi.BasicBlock{HasChildren: true},
			BulletedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("Parent")},
				Children: notionapi.Blocks{
					&notionapi.BulletedListItemBlock{
						BulletedListItem: notionapi.ListItem{
							RichText: []notionapi.RichText{richText("Child")},
						},
					},
				},
			},
		},
	}
	want := "- Parent\n  - Child\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestNestedParagraph(t *testing.T) {
	blocks := []notionapi.Block{
		&notionapi.ParagraphBlock{
			BasicBlock: notionapi.BasicBlock{HasChildren: true},
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richText("Outer")},
				Children: notionapi.Blocks{
					&notionapi.ParagraphBlock{
						Paragraph: notionapi.Paragraph{
							RichText: []notionapi.RichText{richText("Inner")},
						},
					},
				},
			},
		},
	}
	want := "Outer\n\n  Inner\n\n"
	got := convert.BlocksToMarkdown(blocks)
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

// ─── empty input ──────────────────────────────────────────────────────────────

func TestEmptyBlocks(t *testing.T) {
	got := convert.BlocksToMarkdown(nil)
	if got != "" {
		t.Errorf("got:\n%q\nwant empty string", got)
	}
}
