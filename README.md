# tuidit - Terminal Code Editor

A terminal UI based CLI code editor written in Go with modular code architecture.

## Features

- **Two Startup Modes**:
  - Interactive mode: Open editor without selecting any directory/file, then use TUI to open directories or files
  - Direct mode: Specify a directory or file path via command line argument

- **File Explorer Panel**:
  - Collapsible folder tree on the left side
  - Navigate through directories with keyboard
  - Create, rename, and delete files and folders

- **Editor Panel**:
  - Full-featured text editing with syntax highlighting support
  - Vim-like modal editing (Normal/Insert modes)
  - Line numbers display
  - Save files with Ctrl+S

## Installation

### Prerequisites
- Go 1.21 or higher

### Build from Source

```bash
go build -o tuidit ./cmd/editor/
```

## Usage

### Interactive Mode
Run without arguments to see the startup menu:

```bash
./tuidit
```

Options:
1. Open Directory - Browse and select a directory
2. Open File - Open a specific file for editing
3. New File - Start with an empty editor

### Direct Mode
Open a specific directory or file:

```bash
# Open a directory
./tuidit /path/to/directory

# Open a file
./tuidit /path/to/file.go
```

## Keyboard Shortcuts

### Startup Screen
| Key | Action |
|-----|--------|
| `1` | Open Directory |
| `2` | Open File |
| `3` | New File (Empty Editor) |
| `Q` | Quit |

### File Explorer (Normal Mode)
| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `←/h` | Collapse directory |
| `→/l` | Expand directory |
| `Enter` | Open file/Expand directory |
| `n` | Create new file |
| `N` | Create new folder |
| `F2` | Rename file/folder |
| `Del/d` | Delete file/folder |
| `r` | Refresh directory |
| `Tab` | Focus editor panel |
| `Ctrl+O` | Open file/directory dialog |
| `Ctrl+Q` | Quit |

### Editor (Normal Mode)
| Key | Action |
|-----|--------|
| `i` | Enter insert mode |
| `a` | Append after cursor |
| `o` | Open new line below |
| `O` | Open new line above |
| `d` | Delete current line |
| `x` | Delete character under cursor |
| `X` | Delete character before cursor |
| `0/^` | Go to line start |
| `$` | Go to line end |
| `w` | Move to next word |
| `b` | Move to previous word |
| `g` | Go to first line |
| `G` | Go to last line |
| `↑/k` | Move up |
| `↓/j` | Move down |
| `←/h` | Move left |
| `→/l` | Move right |
| `Tab` | Focus explorer panel |
| `Ctrl+S` | Save file |
| `Ctrl+O` | Open file dialog |
| `Ctrl+N` | New file dialog |
| `Ctrl+Q` | Quit |

### Editor (Insert Mode)
| Key | Action |
|-----|--------|
| `Esc` | Return to normal mode |
| `Enter` | Insert newline |
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character under cursor |
| `Tab` | Insert tab (4 spaces) |
| `Ctrl+S` | Save file |
| `Ctrl+Q` | Quit |

## Project Structure

```
tuidit/
├── cmd/
│   └── editor/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── editor/
│   │   └── editor.go        # Text editor component
│   ├── explorer/
│   │   └── filetree.go      # File tree/explorer component
│   ├── model/
│   │   ├── app.go           # Application state
│   │   └── types.go         # Type definitions
│   ├── tui/
│   │   ├── handlers.go      # Keyboard input handlers
│   │   └── tui.go           # TUI rendering
│   └── utils/
│       └── fileops.go       # File operations utilities
├── go.mod
├── go.sum
└── README.md
```

## Architecture

The editor is built with a modular architecture:

- **Model**: Defines the core data structures and application state
- **Explorer**: Manages the file tree navigation and display
- **Editor**: Handles text editing operations
- **TUI**: Renders the terminal UI and handles user input
- **Utils**: Provides file system operations
- **Config**: Manages application configuration

## License

MIT License
