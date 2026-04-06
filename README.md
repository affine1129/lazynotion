# LazyNotion

A terminal UI for browsing and editing Notion databases — inspired by [lazygit](https://github.com/jesseduffield/lazygit).

## Features

- Browse Notion databases and pages in a two-pane terminal UI
- Load database pages and page content on demand to keep startup fast
- Navigate with Vim-style keys (`j` / `k`)
- Collapse and expand databases with `Enter`
- Open pages for editing in Neovim/Vim/vi with `e`
- Convert Notion API block objects to Markdown (`pkg/convert`)

## Requirements

- Go 1.24 or later
- A Notion integration token (for real API access)

## Getting Started

```sh
# Clone the repository
git clone https://github.com/affine1129/lazynotion
cd lazynotion

# Configure your Notion integration token (optional — mock data is used when absent)
cp .env.example .env
# Edit .env and set NOTION_TOKEN=secret_...

# Run
./exec.sh
# or
go run ./pkg/
```

## Install via Homebrew

```sh
# One-time: add tap
brew tap affine1129/lazynotion

# Install
brew install lazynotion
```

To publish Homebrew updates automatically on release tags:

1. Create a tap repository: `affine1129/homebrew-lazynotion`
2. Add a fine-grained GitHub token as repository secret:
	`HOMEBREW_TAP_GITHUB_TOKEN`
3. Push a version tag (e.g. `v0.1.0`) to trigger release workflow

The workflow uses `.goreleaser.yaml` and updates the formula in your tap repo.

`.env` is loaded automatically on startup, so values such as `NOTION_TOKEN`
are available without running `export` manually.

## Key Bindings

| Key        | Action                              |
|------------|-------------------------------------|
| `j`        | Move cursor down                    |
| `k`        | Move cursor up                      |
| `Enter`    | Toggle database collapsed/expanded  |
| `e`        | Open page in Neovim/Vim/vi for editing |
| `q`        | Quit                                |
| `Ctrl+C`   | Quit                                |

Edits are saved by your external editor. Use `:w` to write changes and `:wq`
to write and return to LazyNotion.

## Package: `pkg/convert`

`pkg/convert` converts Notion API block objects (`notionapi.Block`) into
LazyNotion's Markdown representation.

### Entry Point

```go
import "github.com/affine1129/lazynotion/pkg/convert"

md := convert.BlocksToMarkdown(blocks) // blocks is []notionapi.Block
```

### Supported Block Types

| Notion block type      | Markdown output                          |
|------------------------|------------------------------------------|
| `paragraph`            | Plain text followed by a blank line      |
| `heading_1`            | `# text`                                 |
| `heading_2`            | `## text`                                |
| `heading_3`            | `### text`                               |
| `bulleted_list_item`   | `- text`                                 |
| `numbered_list_item`   | `1. text`, `2. text`, … (auto-numbered)  |
| `to_do`                | `- [ ] text` / `- [x] text`             |
| `code`                 | ` ```language … ``` `                    |
| `quote`                | `> text`                                 |
| `divider`              | `---`                                    |
| `image`                | `![caption](url)`                        |

### Rich-Text Decorations

All block types that carry rich text support the following inline decorations:

| Notion annotation | Markdown syntax   |
|-------------------|-------------------|
| Bold              | `**text**`        |
| Italic            | `*text*`          |
| Inline code       | `` `text` ``      |
| Strikethrough     | `~~text~~`        |
| Hyperlink         | `[text](url)`     |

Decorations are applied in the order: inline code → bold → italic →
strikethrough → hyperlink.

### Nested Blocks

Child blocks are rendered with **two-space indentation** per depth level.
Nesting is supported for `paragraph`, `bulleted_list_item`,
`numbered_list_item`, `to_do`, and `quote`.

### Numbered List Counter

The counter for `numbered_list_item` resets to `1` whenever a block of a
different type appears in the sequence, matching standard Markdown behaviour.

## Development

```sh
# Run all tests
go test ./...

# Format code
go fmt ./...

# Vet
go vet ./...
```

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| [github.com/jomei/notionapi](https://github.com/jomei/notionapi) | v1.13.3 | Notion API client |
| [github.com/jroimartin/gocui](https://github.com/jroimartin/gocui) | v0.5.0 | Terminal UI |

## License

MIT
