// Package convert provides utilities for converting Notion API block structures
// into LazyNotion-flavored Markdown.
//
// Supported block types:
//   - paragraph
//   - heading_1, heading_2, heading_3
//   - bulleted_list_item
//   - numbered_list_item
//   - to_do
//   - code
//   - quote
//   - divider
//   - image
//
// Rich-text decorations (bold, italic, code, strikethrough, links) are mapped
// to their standard Markdown equivalents.  Nested child blocks are rendered
// with a two-space indent per depth level.
package convert

import (
	"strings"

	"github.com/jomei/notionapi"
)

// Convert converts a top-level slice of Notion blocks to a Markdown string.
func Convert(blocks notionapi.Blocks) string {
	return convertBlocks(blocks, 0)
}

// convertBlocks renders a slice of blocks at the given nesting depth.
func convertBlocks(blocks notionapi.Blocks, depth int) string {
	var sb strings.Builder
	for _, b := range blocks {
		sb.WriteString(ConvertBlock(b, depth))
	}
	return sb.String()
}

// ConvertBlock converts a single Notion block to its Markdown representation.
// depth controls the indentation level for nested (child) blocks.
func ConvertBlock(block notionapi.Block, depth int) string {
	indent := strings.Repeat("  ", depth)

	switch b := block.(type) {
	case *notionapi.ParagraphBlock:
		text := ConvertRichText(b.Paragraph.RichText)
		result := indent + text + "\n"
		if len(b.Paragraph.Children) > 0 {
			result += convertBlocks(b.Paragraph.Children, depth+1)
		}
		return result

	case *notionapi.Heading1Block:
		text := ConvertRichText(b.Heading1.RichText)
		result := indent + "# " + text + "\n"
		if len(b.Heading1.Children) > 0 {
			result += convertBlocks(b.Heading1.Children, depth+1)
		}
		return result

	case *notionapi.Heading2Block:
		text := ConvertRichText(b.Heading2.RichText)
		result := indent + "## " + text + "\n"
		if len(b.Heading2.Children) > 0 {
			result += convertBlocks(b.Heading2.Children, depth+1)
		}
		return result

	case *notionapi.Heading3Block:
		text := ConvertRichText(b.Heading3.RichText)
		result := indent + "### " + text + "\n"
		if len(b.Heading3.Children) > 0 {
			result += convertBlocks(b.Heading3.Children, depth+1)
		}
		return result

	case *notionapi.BulletedListItemBlock:
		text := ConvertRichText(b.BulletedListItem.RichText)
		result := indent + "- " + text + "\n"
		if len(b.BulletedListItem.Children) > 0 {
			result += convertBlocks(b.BulletedListItem.Children, depth+1)
		}
		return result

	case *notionapi.NumberedListItemBlock:
		text := ConvertRichText(b.NumberedListItem.RichText)
		result := indent + "1. " + text + "\n"
		if len(b.NumberedListItem.Children) > 0 {
			result += convertBlocks(b.NumberedListItem.Children, depth+1)
		}
		return result

	case *notionapi.ToDoBlock:
		text := ConvertRichText(b.ToDo.RichText)
		check := "[ ]"
		if b.ToDo.Checked {
			check = "[x]"
		}
		result := indent + "- " + check + " " + text + "\n"
		if len(b.ToDo.Children) > 0 {
			result += convertBlocks(b.ToDo.Children, depth+1)
		}
		return result

	case *notionapi.CodeBlock:
		lang := b.Code.Language
		code := richTextToPlain(b.Code.RichText)
		return indent + "```" + lang + "\n" + code + "\n" + indent + "```\n"

	case *notionapi.QuoteBlock:
		text := ConvertRichText(b.Quote.RichText)
		result := indent + "> " + text + "\n"
		if len(b.Quote.Children) > 0 {
			result += convertBlocks(b.Quote.Children, depth+1)
		}
		return result

	case *notionapi.DividerBlock:
		return indent + "---\n"

	case *notionapi.ImageBlock:
		url := b.Image.GetURL()
		caption := richTextToPlain(b.Image.Caption)
		return indent + "![" + caption + "](" + url + ")\n"

	default:
		return ""
	}
}

// ConvertRichText converts a slice of Notion RichText objects to a Markdown
// string, applying inline decorations (bold, italic, code, strikethrough,
// links).
func ConvertRichText(richTexts []notionapi.RichText) string {
	var sb strings.Builder
	for _, rt := range richTexts {
		content := rt.PlainText
		if content == "" {
			continue
		}

		// Resolve the display URL for linked text.
		href := rt.Href
		if rt.Text != nil && rt.Text.Link != nil && rt.Text.Link.Url != "" {
			href = rt.Text.Link.Url
		}

		// Apply inline annotations.
		if rt.Annotations != nil {
			if rt.Annotations.Code {
				content = "`" + content + "`"
			} else {
				if rt.Annotations.Bold && rt.Annotations.Italic {
					content = "**_" + content + "_**"
				} else if rt.Annotations.Bold {
					content = "**" + content + "**"
				} else if rt.Annotations.Italic {
					content = "_" + content + "_"
				}
				if rt.Annotations.Strikethrough {
					content = "~~" + content + "~~"
				}
			}
		}

		// Wrap in a Markdown link if there is a URL.
		if href != "" {
			content = "[" + content + "](" + href + ")"
		}

		sb.WriteString(content)
	}
	return sb.String()
}

// richTextToPlain concatenates the PlainText of all RichText segments without
// applying any Markdown decoration.  This is used where raw text is needed
// (e.g. code block content, image captions).
func richTextToPlain(richTexts []notionapi.RichText) string {
	var sb strings.Builder
	for _, rt := range richTexts {
		sb.WriteString(rt.PlainText)
	}
	return sb.String()
}
