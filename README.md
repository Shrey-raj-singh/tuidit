# tuidit - Terminal Code Editor

A terminal UI based CLI code editor written in Go with modular code architecture.

## Features

- **Open from anywhere**: Add `tuidit` to your PATH (see [Use from anywhere](#use-from-anywhere)) and run `tuidit` from any directory.
- **Last workspace**: Automatically reopens the last used directory when you run `tuidit` with no arguments (like VS Code/Cursor).
- **File Explorer Panel**: Collapsible folder tree, navigate with keyboard, create/rename/delete files and folders, cut/copy/paste, resize panel (Ctrl+Left/Right).
- **Editor Panel**: Vim-like modal editing (Normal/Insert), line numbers, save with Ctrl+S.
- **Command guide**: Press **Ctrl+H** anytime to open a context-aware shortcut guide.

## Installation

### Pre-built binaries (recommended)

- **GitHub Releases** — Download the binary for your platform from the [Releases](https://github.com/Shrey-raj-singh/tuidit/releases) page (Linux amd64/arm64, Windows amd64, macOS amd64/arm64). Extract and add the directory to your PATH.

- **Homebrew** (macOS / Linux):
  ```bash
  brew tap Shrey-raj-singh/tuidit https://github.com/Shrey-raj-singh/tuidit
  brew install tuidit
  ```
  Or install the formula from this repo:
  ```bash
  brew install Shrey-raj-singh/tuidit/tuidit
  ```

- **Winget** (Windows) — *Not yet available.* The package must be submitted to the [winget-pkgs](https://github.com/microsoft/winget-pkgs) repo first. Until then, use the **GitHub Releases** download or **build from source** (see below).

- **APT** (Debian/Ubuntu) — Add the repo and install (after a release and with [GitHub Pages enabled](https://docs.github.com/en/pages)):
  ```bash
  sudo mkdir -p /etc/apt/keyrings
  echo "deb [trusted=yes] https://shrey-raj-singh.github.io/tuidit/ stable main" | sudo tee /etc/apt/sources.list.d/tuidit.list
  sudo apt update
  sudo apt install tuidit
  ```

- **DNF** (Fedora / RHEL / CentOS / Rocky) — Add the repo and install (after a release and with [GitHub Pages enabled](https://docs.github.com/en/pages)):
  ```bash
  echo '[tuidit]
  name=tuidit
  baseurl=https://shrey-raj-singh.github.io/tuidit/rpm/$basearch
  enabled=1
  gpgcheck=0' | sudo tee /etc/yum.repos.d/tuidit.repo
  sudo dnf install tuidit
  ```

### Build from source

**Prerequisites:** Go 1.21 or higher.

```bash
go build -o tuidit ./cmd/editor/
```

On Windows (executable name can include `.exe`):

```powershell
go build -o tuidit.exe ./cmd/editor/
```

### Use from anywhere

Add the directory that contains the `tuidit` (or `tuidit.exe`) binary to your **PATH** so you can run it from any folder.

**Linux / macOS (bash or zsh)**

1. Choose a directory for binaries, e.g. `$HOME/bin` or `/usr/local/bin`.
2. Build and copy the binary there:
   ```bash
   go build -o tuidit ./cmd/editor/
   mkdir -p ~/bin && mv tuidit ~/bin/
   ```
3. Add to PATH in `~/.bashrc` or `~/.zshrc`:
   ```bash
   export PATH="$HOME/bin:$PATH"
   ```
4. Reload the shell:
   ```bash
   source ~/.bashrc   # or source ~/.zshrc
   ```
5. Run from anywhere:
   ```bash
   tuidit
   tuidit /path/to/project
   ```

**Windows (PowerShell)**

1. Build the binary (e.g. in your project folder):
   ```powershell
   go build -o tuidit.exe ./cmd/editor/
   ```
2. Pick a folder that will be on PATH (e.g. `C:\Tools` or a folder you created). Copy `tuidit.exe` there.
3. Add that folder to your user PATH:
   - Open **Settings** → **System** → **About** → **Advanced system settings** → **Environment Variables**.
   - Under **User variables**, select **Path** → **Edit** → **New**, and add the folder (e.g. `C:\Tools`).
   - Confirm with **OK**.
4. Restart the terminal (or open a new one), then run:
   ```powershell
   tuidit
   tuidit C:\path\to\project
   ```

**Windows (Command Prompt)**

- Same as PowerShell: add the folder containing `tuidit.exe` to the **Path** user variable via **Environment Variables**, then use a new `cmd` window and run `tuidit` from any directory.

### Linux / WSL: file watcher (inotify and polling)

The file explorer auto-refreshes when files change on disk. On Linux it uses inotify; if your project has many directories, you may hit the system limit (`fs.inotify.max_user_watches`, often 8192). Tuidit then watches only the root directory so refresh still works for top-level changes. For full tree watching on large projects, increase the limit (e.g. `echo 524288 | sudo tee /proc/sys/fs/inotify/max_user_watches` or set `fs.inotify.max_user_watches=524288` in `/etc/sysctl.conf` and run `sudo sysctl -p`).

**WSL:** On Windows mounts (e.g. `/mnt/c/...`), inotify does not work. Tuidit falls back to polling the tree every 2 seconds, so the explorer still auto-refreshes when you add, remove, or change files from Windows or another terminal.

## Usage

### With no arguments
Opens the editor and restores the **last used workspace** (directory) if there was one; otherwise shows an empty explorer. Use **Ctrl+O** to open a folder or file.

```bash
tuidit
```

### With a path
Open a specific directory or file:

```bash
# Open a directory
tuidit /path/to/directory

# Open a file
tuidit /path/to/file.go
```

On Windows:

```powershell
tuidit C:\Users\Me\project
tuidit .\src\main.go
```

## Keyboard Shortcuts

Press **Ctrl+H** anytime to open the full, context-aware shortcut guide in a popup.

### File Explorer
| Key | Action |
|-----|--------|
| `↑/k` `↓/j` | Move selection |
| `←/h` `→/l` | Collapse / Expand directory |
| `Enter` | Open file / Expand directory |
| `n` `N` | New file / New folder |
| `F2` | Rename |
| `Del` `d` | Delete |
| `Ctrl+X` `Ctrl+C` `Ctrl+V` | Cut / Copy / Paste |
| `Ctrl+O` | Open file or directory |
| `Ctrl+Left` `Ctrl+Right` | Resize explorer panel |
| `Tab` | Focus editor |
| `r` | Refresh |
| `Esc` `Ctrl+Q` | Quit |
| `Ctrl+H` | Open shortcut guide |

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
| `Ctrl+Left` `Ctrl+Right` | Resize explorer panel |
| `Ctrl+H` | Open shortcut guide |
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
| `Ctrl+H` | Open shortcut guide |
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
