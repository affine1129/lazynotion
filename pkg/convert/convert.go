// Package convert provides functions to convert Notion API block objects into
// LazyNotion's Markdown representation.
//
// Supported block types:
//   - paragraph
//   - heading_1 / heading_2 / heading_3
//   - bulleted_list_item
//   - numbered_list_item
//   - to_do
//   - code
//   - quote
//   - divider
//   - image
//
// Rich-text decorations (bold, italic, strikethrough, inline-code, hyperlinks)
// are applied to every block that contains rich text.
//
// Nested child blocks are rendered with two-space indentation per depth level.
package convert

import (
	"fmt"
	"strings"

	"github.com/jomei/notionapi"
)

// BlocksToMarkdown converts a slice of Notion API Block objects into a single
// Markdown string suitable for use within LazyNotion.
//
// Consecutive numbered-list items are automatically numbered starting from 1;
// the counter resets whenever a non-numbered-list block appears.
func BlocksToMarkdown(blocks []notionapi.Block) string {
	return blocksToMarkdown(blocks, 0)
}

func blocksToMarkdown(blocks []notionapi.Block, depth int) string {
	var sb strings.Builder
	numberedListCounter := 0

	for _, block := range blocks {
		if _, ok := block.(*notionapi.NumberedListItemBlock); !ok {
			numberedListCounter = 0
		}

		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			sb.WriteString(indent(depth))
			sb.WriteString(richTextsToMarkdown(b.Paragraph.RichText))
			sb.WriteString("\n\n")
			if b.HasChildren {
				sb.WriteString(blocksToMarkdown(b.Paragraph.Children, depth+1))
			}

		case *notionapi.Heading1Block:
			sb.WriteString(indent(depth))
			sb.WriteString("# ")
			sb.WriteString(richTextsToMarkdown(b.Heading1.RichText))
			sb.WriteString("\n\n")

		case *notionapi.Heading2Block:
			sb.WriteString(indent(depth))
			sb.WriteString("## ")
			sb.WriteString(richTextsToMarkdown(b.Heading2.RichText))
			sb.WriteString("\n\n")

		case *notionapi.Heading3Block:
			sb.WriteString(indent(depth))
			sb.WriteString("### ")
			sb.WriteString(richTextsToMarkdown(b.Heading3.RichText))
			sb.WriteString("\n\n")

		case *notionapi.BulletedListItemBlock:
			sb.WriteString(indent(depth))
			sb.WriteString("- ")
			sb.WriteString(richTextsToMarkdown(b.BulletedListItem.RichText))
			sb.WriteString("\n")
			if b.HasChildren {
				sb.WriteString(blocksToMarkdown(b.BulletedListItem.Children, depth+1))
			}

		case *notionapi.NumberedListItemBlock:
			numberedListCounter++
			sb.WriteString(indent(depth))
			sb.WriteString(fmt.Sprintf("%d. ", numberedListCounter))
			sb.WriteString(richTextsToMarkdown(b.NumberedListItem.RichText))
			sb.WriteString("\n")
			if b.HasChildren {
				sb.WriteString(blocksToMarkdown(b.NumberedListItem.Children, depth+1))
			}

		case *notionapi.ToDoBlock:
			sb.WriteString(indent(depth))
			if b.ToDo.Checked {
				sb.WriteString("- [x] ")
			} else {
				sb.WriteString("- [ ] ")
			}
			sb.WriteString(richTextsToMarkdown(b.ToDo.RichText))
			sb.WriteString("\n")
			if b.HasChildren {
				sb.WriteString(blocksToMarkdown(b.ToDo.Children, depth+1))
			}

		case *notionapi.CodeBlock:
			sb.WriteString(indent(depth))
			sb.WriteString("```")
			sb.WriteString(b.Code.Language)
			sb.WriteString("\n")
			sb.WriteString(richTextsToMarkdown(b.Code.RichText))
			sb.WriteString("\n")
			sb.WriteString(indent(depth))
			sb.WriteString("```\n\n")

		case *notionapi.QuoteBlock:
			text := strings.TrimRight(richTextsToMarkdown(b.Quote.RichText), "\n")
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				sb.WriteString(indent(depth))
				sb.WriteString("> ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
			if b.HasChildren {
				sb.WriteString(blocksToMarkdown(b.Quote.Children, depth+1))
			}

		case *notionapi.DividerBlock:
			sb.WriteString(indent(depth))
			sb.WriteString("---\n\n")

		case *notionapi.ImageBlock:
			url := b.Image.GetURL()
			caption := richTextsToMarkdown(b.Image.Caption)
			sb.WriteString(indent(depth))
			sb.WriteString(fmt.Sprintf("![%s](%s)\n\n", caption, url))
		}
	}

	return sb.String()
}

// richTextsToMarkdown concatenates the Markdown representation of each rich
// text segment in the slice.
func richTextsToMarkdown(rts []notionapi.RichText) string {
	var sb strings.Builder
	for _, rt := range rts {
		sb.WriteString(richTextToMarkdown(rt))
	}
	return sb.String()
}

// richTextToMarkdown converts a single RichText segment into Markdown.
//
// Decoration order applied (inner → outer):
//  1. inline code  (`text`)
//  2. bold         (**text**)
//  3. italic       (*text*)
//  4. strikethrough (~~text~~)
//  5. hyperlink    ([text](url))
func richTextToMarkdown(rt notionapi.RichText) string {
	text := rt.PlainText

	if rt.Annotations != nil {
		if rt.Annotations.Code {
			text = "`" + text + "`"
		}
		if rt.Annotations.Bold {
			text = "**" + text + "**"
		}
		if rt.Annotations.Italic {
			text = "*" + text + "*"
		}
		if rt.Annotations.Strikethrough {
			text = "~~" + text + "~~"
		}
	}

	if rt.Href != "" {
		text = fmt.Sprintf("[%s](%s)", text, rt.Href)
	}

	return text
}

// indent returns a string of two spaces repeated depth times.
func indent(depth int) string {
	return strings.Repeat("  ", depth)
}
