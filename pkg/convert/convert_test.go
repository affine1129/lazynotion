package convert_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/affine1129/lazynotion/pkg/convert"
	"github.com/jomei/notionapi"
)

// update is a flag that regenerates the golden files when set.
// Run: go test ./pkg/convert/... -update
var update = flag.Bool("update", false, "update golden files")

// goldenTest is a helper that compares got against a golden file.
// When -update is passed the golden file is overwritten with got.
func goldenTest(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("failed to update golden file %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	if got != string(want) {
		t.Errorf("output mismatch for %s\ngot:\n%s\nwant:\n%s", name, got, string(want))
	}
}

// richText is a small helper that builds a plain notionapi.RichText.
func richText(content string) notionapi.RichText {
	return notionapi.RichText{
		PlainText: content,
		Text:      &notionapi.Text{Content: content},
	}
}

// annotated builds a notionapi.RichText with the given annotations applied.
func annotated(content string, a notionapi.Annotations) notionapi.RichText {
	return notionapi.RichText{
		PlainText:   content,
		Text:        &notionapi.Text{Content: content},
		Annotations: &a,
	}
}

// linked builds a notionapi.RichText that is wrapped in a hyperlink.
func linked(content, url string) notionapi.RichText {
	return notionapi.RichText{
		PlainText: content,
		Text:      &notionapi.Text{Content: content, Link: &notionapi.Link{Url: url}},
		Href:      url,
	}
}

// ── paragraph ────────────────────────────────────────────────────────────────

func TestConvert_Paragraph(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{richText("Hello World")},
			},
		},
	}
	goldenTest(t, "paragraph", convert.Convert(blocks))
}

// ── headings ─────────────────────────────────────────────────────────────────

func TestConvert_Headings(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.Heading1Block{
			Heading1: notionapi.Heading{RichText: []notionapi.RichText{richText("Heading One")}},
		},
		&notionapi.Heading2Block{
			Heading2: notionapi.Heading{RichText: []notionapi.RichText{richText("Heading Two")}},
		},
		&notionapi.Heading3Block{
			Heading3: notionapi.Heading{RichText: []notionapi.RichText{richText("Heading Three")}},
		},
	}
	goldenTest(t, "headings", convert.Convert(blocks))
}

// ── bulleted list ─────────────────────────────────────────────────────────────

func TestConvert_BulletedList(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.BulletedListItemBlock{
			BulletedListItem: notionapi.ListItem{RichText: []notionapi.RichText{richText("Item A")}},
		},
		&notionapi.BulletedListItemBlock{
			BulletedListItem: notionapi.ListItem{RichText: []notionapi.RichText{richText("Item B")}},
		},
		&notionapi.BulletedListItemBlock{
			BulletedListItem: notionapi.ListItem{RichText: []notionapi.RichText{richText("Item C")}},
		},
	}
	goldenTest(t, "bulleted_list", convert.Convert(blocks))
}

// ── numbered list ─────────────────────────────────────────────────────────────

func TestConvert_NumberedList(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{RichText: []notionapi.RichText{richText("First")}},
		},
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{RichText: []notionapi.RichText{richText("Second")}},
		},
		&notionapi.NumberedListItemBlock{
			NumberedListItem: notionapi.ListItem{RichText: []notionapi.RichText{richText("Third")}},
		},
	}
	goldenTest(t, "numbered_list", convert.Convert(blocks))
}

// ── to_do ─────────────────────────────────────────────────────────────────────

func TestConvert_ToDo(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.ToDoBlock{
			ToDo: notionapi.ToDo{
				RichText: []notionapi.RichText{richText("Buy milk")},
				Checked:  false,
			},
		},
		&notionapi.ToDoBlock{
			ToDo: notionapi.ToDo{
				RichText: []notionapi.RichText{richText("Write tests")},
				Checked:  true,
			},
		},
	}
	goldenTest(t, "todo", convert.Convert(blocks))
}

// ── code ──────────────────────────────────────────────────────────────────────

func TestConvert_Code(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.CodeBlock{
			Code: notionapi.Code{
				Language: "go",
				RichText: []notionapi.RichText{richText(`fmt.Println("hello")`)},
			},
		},
	}
	goldenTest(t, "code", convert.Convert(blocks))
}

// ── quote ─────────────────────────────────────────────────────────────────────

func TestConvert_Quote(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.QuoteBlock{
			Quote: notionapi.Quote{
				RichText: []notionapi.RichText{richText("A wise quote")},
			},
		},
	}
	goldenTest(t, "quote", convert.Convert(blocks))
}

// ── divider ───────────────────────────────────────────────────────────────────

func TestConvert_Divider(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.DividerBlock{},
	}
	goldenTest(t, "divider", convert.Convert(blocks))
}

// ── image ─────────────────────────────────────────────────────────────────────

func TestConvert_Image(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.ImageBlock{
			Image: notionapi.Image{
				Caption:  []notionapi.RichText{richText("a cat")},
				Type:     "external",
				External: &notionapi.FileObject{URL: "https://example.com/cat.png"},
			},
		},
	}
	goldenTest(t, "image", convert.Convert(blocks))
}

// ── rich text decorations ─────────────────────────────────────────────────────

func TestConvert_RichText(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.ParagraphBlock{
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					richText("plain "),
					annotated("bold", notionapi.Annotations{Bold: true}),
					richText(" "),
					annotated("italic", notionapi.Annotations{Italic: true}),
					richText(" "),
					annotated("bold-italic", notionapi.Annotations{Bold: true, Italic: true}),
					richText(" "),
					annotated("code", notionapi.Annotations{Code: true}),
					richText(" "),
					annotated("strike", notionapi.Annotations{Strikethrough: true}),
					richText(" "),
					linked("link", "https://example.com"),
				},
			},
		},
	}
	goldenTest(t, "rich_text", convert.Convert(blocks))
}

// ── nested blocks ─────────────────────────────────────────────────────────────

func TestConvert_Nested(t *testing.T) {
	blocks := notionapi.Blocks{
		&notionapi.BulletedListItemBlock{
			BulletedListItem: notionapi.ListItem{
				RichText: []notionapi.RichText{richText("Parent item")},
				Children: notionapi.Blocks{
					&notionapi.BulletedListItemBlock{
						BulletedListItem: notionapi.ListItem{
							RichText: []notionapi.RichText{richText("Child item")},
						},
					},
				},
			},
		},
	}
	goldenTest(t, "nested", convert.Convert(blocks))
}
