# madesu

A small terminal app for editing Markdown files. A file explorer sits on the left, a text editor on the right, and you can flip the right pane into a rendered Markdown preview.

Markdown simple utility.

## Features

- File explorer with directory navigation (folders first, dotfiles hidden)
- Plain text editor for `.md` and other text files
- Rendered Markdown preview, scrollable with keyboard and mouse
- Resizable panes: drag the divider with the mouse, or use `ctrl+←` / `ctrl+→`
- Mouse wheel / trackpad scrolling in both editor and preview
- Single binary, no runtime dependencies

## Requirements

- Go 1.26 or newer (see `go.mod`)
- A terminal with 256-color support and xterm mouse reporting. Terminal.app and iTerm2 on macOS are the setups I have used; other modern terminals should work too.

## Install

From source:

```sh
git clone git@github.com:dhinugraha/madesu.git
cd madesu
go build -o madesu .
```

Put the binary somewhere on your `PATH`, for example:

```sh
mv madesu ~/.local/bin/
```

Or install straight into `$(go env GOPATH)/bin`:

```sh
go install .
```

## Usage

```sh
madesu            # open the current directory
madesu ~/Notes    # open a specific directory
```

The first argument is always treated as a directory path. There are no other flags yet.

## Keyboard

### Explorer (left pane)

| Key | Action |
| --- | --- |
| `↑` / `k`, `↓` / `j` | Move selection |
| `enter` | Open a file, or enter a directory (`..` goes up) |
| `tab` | Switch to the right pane |
| `ctrl+c` | Quit |

### Editor (right pane)

| Key | Action |
| --- | --- |
| `ctrl+s` | Save |
| `ctrl+p` | Switch to the preview |
| `tab` | Switch to the explorer |
| `pgup` / `pgdown` | Move the cursor one screen up or down |
| `home` / `end` | Start / end of the current line |
| `ctrl+home` / `ctrl+end` | Start / end of the document |

Editing follows the standard `bubbles/textarea` keys, such as `ctrl+a` / `ctrl+e` for line start and end, and `ctrl+w` to delete a word.

### Preview (right pane)

| Key | Action |
| --- | --- |
| `↑` / `k`, `↓` / `j` | Scroll one line |
| `pgup` / `pgdown` | Scroll one screen |
| `home` / `end` | Top / bottom of the document |
| `ctrl+p` | Return to the editor |
| `tab` | Switch to the explorer |

### Opening a file

Opening a file puts the cursor on its first line, and the editor gets focus.

## Resizing panes

| Key | Action |
| --- | --- |
| `ctrl+←` / `ctrl+→` | Move the divider 2 columns left or right (works from either pane) |

The mouse can also drag the divider; see below.

## Mouse

- **Drag the divider** between the explorer and the right pane to resize them. Press on the border or the gap (±1 column), drag, and release. The explorer keeps at least 16 columns and the right pane at least 30.
- **Wheel / trackpad** scrolls the right pane by 3 lines per notch. In the editor, the cursor moves with the view so the text under the pointer keeps scrolling.

Because the app captures mouse events, your terminal's native text selection is disabled while it runs. To select text anyway, hold the terminal's bypass modifier while dragging: `option` in iTerm2, and usually `fn` in Terminal.app.

## Markdown support

The preview renders CommonMark plus these GitHub-flavored extensions:

- Headings (levels 1–6, with the `#` markers kept visible)
- Emphasis, bold, bold-italic, strikethrough, inline code
- Links (text followed by the URL) and autolinks; images show as `[img: alt] url`
- Block quotes, including nested quotes
- Bullet and ordered lists, nested up to any depth, tight and loose
- Task lists (`☑` / `☐`)
- Fenced and indented code blocks, with the language label
- Tables with left, center, and right alignment; long cells wrap inside their column
- Footnotes (`[^1]`), rendered as `[1]` with the notes listed at the end
- Definition lists
- Backslash escapes and HTML entities
- Horizontal rules

Not rendered yet:

- **Math** (`$...$`, `$$...$$`) is shown as plain text.
- **Raw HTML** is shown as plain text, since a terminal cannot render it.
- **Syntax highlighting** in code blocks is not implemented. Code blocks only get a background color.
- **Mermaid** and other diagram languages are shown as plain code blocks.
- **Images** are not displayed; the alt text and URL are shown.

## Configuration

| Variable | Values | Default | Purpose |
| --- | --- | --- | --- |
| `MADESU_THEME` | `dark`, `light` | `dark` | Color palette for the UI and the Markdown preview. Use `light` for a light terminal background. |

Terminal background detection is not automatic. The default is dark, so set `MADESU_THEME=light` if your background is light.

## Development

```sh
go vet ./...
go test -race -count=1 ./...
```

The code is a single Go package:

| File | Contents |
| --- | --- |
| `main.go` | Bubble Tea model, key and mouse handling, layout, theme |
| `explorer.go` | Directory listing (folders first, dotfiles hidden) |
| `mdrender.go` | Markdown to ANSI renderer on top of goldmark |
| `*_test.go` | Tests for layout geometry, keys, mouse drag and wheel, and the renderer |
| `testdata/` | Sample Markdown used by the layout tests |

Dependencies: [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [goldmark](https://github.com/yuin/goldmark).

## Known limitations

- The explorer has no mouse wheel support yet.
- The editor's wheel scrolling moves the cursor along with the view, because `bubbles/textarea` does not expose its scroll offset.
- The whole file is loaded into memory. Very large files will be slow to render in the preview (roughly 100 ms per 600 KB).
- No undo beyond what the editor component provides.
- There is no prompt for unsaved changes when opening another file.

## License

No license has been chosen yet.
