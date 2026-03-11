# Tuidit v0.1.0 — Feature Reference

A comprehensive list of every feature in the current version of tuidit.

---

## Table of Contents

- [CLI & Startup](#cli--startup)
- [File Explorer](#file-explorer)
- [Editor — Normal Mode](#editor--normal-mode)
- [Editor — Insert Mode](#editor--insert-mode)
- [Mouse Support](#mouse-support)
- [Dialog System](#dialog-system)
- [Git Integration](#git-integration)
- [File Watcher & Auto-Refresh](#file-watcher--auto-refresh)
- [Rendering & UI](#rendering--ui)
- [Editor Core](#editor-core)
- [Configuration](#configuration)
- [File Operations](#file-operations)

---

## CLI & Startup

| Feature | Description |
|---------|-------------|
| **Version flag** | `tuidit -v` or `tuidit --version` prints the current version |
| **Open directory** | `tuidit /path/to/dir` opens the directory in the explorer panel |
| **Open file** | `tuidit /path/to/file` opens the file in the editor and its parent directory in the explorer |
| **Last workspace restore** | Running `tuidit` with no arguments reopens the last used directory (persisted in `~/.tuidit/last_workspace`) |
| **Path expansion** | Supports `~/` tilde expansion for home directory paths |
| **Alt screen** | Uses the terminal's alternate screen buffer for a clean exit |
| **Mouse support** | Mouse cell motion events are enabled at startup |

---

## File Explorer

### Navigation

| Keybinding | Action |
|------------|--------|
| `Up` / `k` | Move selection up |
| `Down` / `j` | Move selection down |
| `Left` / `h` / `Backspace` | Collapse directory, or jump to parent |
| `Right` / `l` | Expand directory |
| `Enter` | Open file in editor / toggle directory expand-collapse |
| `Tab` | Switch focus to editor panel |

### File & Folder Management

| Keybinding | Action |
|------------|--------|
| `n` | Create new file (dialog pre-filled with selected directory) |
| `N` (Shift+N) | Create new folder |
| `F2` | Rename file or folder |
| `Delete` / `d` | Delete file or folder (with confirmation prompt) |
| `x` / `Ctrl+X` | Cut file or folder |
| `y` / `Ctrl+C` | Copy file or folder |
| `p` / `Ctrl+V` | Paste (move if cut, copy if copied; supports recursive directory copy) |
| `r` | Refresh file tree manually |
| `Ctrl+O` | Open directory dialog |

### Panel Controls

| Keybinding | Action |
|------------|--------|
| `Ctrl+Right` | Widen explorer panel (+3 columns, max 60) |
| `Ctrl+Left` | Narrow explorer panel (-3 columns, min 15) |
| `Ctrl+H` | Open help guide |
| `Ctrl+Q` / `Esc` | Quit application |

### Behaviors

- **Unsaved changes guard** — prompts to save before switching files if the editor has modifications
- **Auto-scroll** — viewport follows the selection as you navigate
- **Directories-first sorting** — directories are always listed before files, alphabetically
- **Expansion state preserved** — refreshing the tree preserves which directories were expanded
- **Lazy loading** — directory children are only loaded when first expanded

---

## Editor — Normal Mode

### Mode Switching

| Keybinding | Action |
|------------|--------|
| `i` | Enter insert mode at cursor |
| `a` | Enter insert mode after cursor |
| `o` | Open new line below, enter insert mode |
| `O` (Shift+O) | Open new line above, enter insert mode |

### Editing

| Keybinding | Action |
|------------|--------|
| `d` | Delete entire current line |
| `x` | Delete character under cursor |
| `X` (Shift+X) | Delete character before cursor |

### Cursor Movement

| Keybinding | Action |
|------------|--------|
| `Up` / `k` | Move cursor up |
| `Down` / `j` | Move cursor down |
| `Left` / `h` | Move cursor left (wraps to previous line) |
| `Right` / `l` | Move cursor right (wraps to next line) |
| `0` / `^` | Jump to beginning of line |
| `$` | Jump to end of line |
| `w` | Move to next word boundary |
| `b` | Move to previous word boundary |
| `g` | Jump to first line |
| `G` (Shift+G) | Jump to last line |

### File & Panel Operations

| Keybinding | Action |
|------------|--------|
| `Tab` | Switch focus to explorer |
| `Ctrl+S` | Save file |
| `Ctrl+O` | Open file dialog |
| `Ctrl+N` | New file dialog |
| `Ctrl+H` | Open help guide |
| `Ctrl+Q` | Quit (prompts to save if modified) |
| `Ctrl+Right` / `Ctrl+Left` | Resize explorer panel |

---

## Editor — Insert Mode

| Keybinding | Action |
|------------|--------|
| `Esc` | Return to normal mode |
| `Enter` | Insert newline (splits line at cursor) |
| `Backspace` | Delete character before cursor (merges lines at column 0) |
| `Delete` | Delete character at cursor (merges with next line at end) |
| `Tab` | Insert 4 spaces |
| `Up` / `Down` / `Left` / `Right` | Cursor movement |
| `Home` / `End` | Jump to line start / end |
| `Ctrl+S` | Save file |
| `Ctrl+H` | Open help guide |
| `Ctrl+Q` | Quit (prompts to save if modified) |
| Any printable character | Inserted at cursor position |

---

## Mouse Support

| Action | Behavior |
|--------|----------|
| **Single click on explorer** | Selects the clicked file or folder |
| **Double click on explorer** | Opens file in editor / toggles directory expand-collapse |
| **Single click on editor** | Places cursor at clicked position (accounts for line numbers and scroll offset) |
| **Double click on editor** | Places cursor and enters insert mode |
| **Click on panel** | Focuses that panel |

---

## Dialog System

### Dialog Types

| Dialog | Trigger | Purpose |
|--------|---------|---------|
| **Open File** | `Ctrl+O` (editor) | Browse and open a file |
| **Open Directory** | `Ctrl+O` (explorer) | Browse and open a directory |
| **New File** | `n` (explorer) / `Ctrl+N` (editor) | Create a new file |
| **New Folder** | `N` (explorer) | Create a new folder |
| **Rename** | `F2` (explorer) | Rename a file or folder |
| **Delete** | `Delete` / `d` (explorer) | Confirm deletion |
| **Save** | Automatic | Prompt to save before closing |
| **Confirm Switch** | Automatic | Save before switching files |
| **Quit** | `Ctrl+Q` with unsaved changes | Save before quitting |
| **Help** | `Ctrl+H` | Context-aware shortcut guide |

### Dialog Features

- **Live file preview** — real-time filtered directory listing as you type
- **Preview navigation** — scroll through suggestions with `Up`/`Down` (wrapping)
- **Tab completion** — auto-completes the selected preview item; appends path separator for directories
- **Navigate into directory** — `Right`/`l` navigates into a highlighted directory
- **Directories-first sorting** — directories shown before files in preview
- **Scroll indicators** — shows "more above" / "more below" when preview list overflows (max 10 visible)
- **Error display** — input errors shown in red below the input field
- **Max width capped** — dialog width is capped at 70 columns for readability on wide terminals

---

## Git Integration

### Editor Gutter Indicators

A 1-character gutter column appears between line numbers and content when a file has git changes:

| Indicator | Color | Meaning |
|-----------|-------|---------|
| `│` | Green (`#3FB950`) | Line added (not in HEAD) |
| `│` | Yellow (`#E3B341`) | Line modified from HEAD |
| `─` | Red (`#F85149`) | Line(s) deleted at this position |

- Computed on file open and after every save
- Compares working copy against `HEAD` using `git diff HEAD --unified=0`
- Untracked files show all lines as added
- No gutter column is shown if the file is not in a git repository

### Explorer Git Status

Each file and folder in the explorer shows a colored status label:

| Label | Color | Meaning |
|-------|-------|---------|
| `M` | Yellow (`#E3B341`) | Modified |
| `A` | Green (`#3FB950`) | Added / staged |
| `U` | Orange (`#F0883E`) | Untracked |
| `D` | Red (`#F85149`) | Deleted |
| `R` | Cyan (`#39C5CF`) | Renamed |
| `C` | Pink (`#F472B6`) | Merge conflict |

Additional behaviors:

- **File names are colored** to match their status (modified files appear yellow, untracked appear orange, etc.)
- **Directory status propagation** — a folder inherits the highest-priority status from its children
- **Gitignored files** — files and folders listed in `.gitignore` are rendered in dim gray (`#4A5568`)
- **Status refreshes** on: app init, directory open, filesystem change, and file save
- Uses `git status --porcelain -u` for reliable parsing
- Uses `git ls-files --others --ignored --exclude-standard --directory` for gitignored detection

---

## File Watcher & Auto-Refresh

| Feature | Description |
|---------|-------------|
| **fsnotify watcher** | Watches the workspace root and all subdirectories for filesystem events |
| **Debounced events** | Changes are debounced at 150ms to avoid excessive refreshes |
| **Polling fallback** | A 2-second polling interval catches changes missed by fsnotify (e.g., WSL `/mnt/c` mounts) |
| **Background subdirectory watching** | Subdirectories are added to the watcher in a background goroutine so the UI stays responsive |
| **Inotify limit handling** | On Linux, if `max_user_watches` is hit, falls back to watching only the root directory |
| **Smart directory skipping** | Skips heavy directories during watching and polling: `.git`, `node_modules`, `vendor`, `__pycache__`, `dist`, `build`, `target`, `out`, `bin`, `obj` |
| **Tree signature comparison** | Polling uses a fast checksum of file counts + modification times to detect changes |
| **Expansion state preserved** | Tree refreshes preserve which directories were expanded |

---

## Rendering & UI

### Layout

- **Two-panel layout** — side-by-side explorer + editor with horizontal join
- **Active panel highlighting** — focused panel has a purple border (`#7C3AED`), unfocused panel has gray
- **Rounded borders** — all panels use rounded corner borders
- **Dynamic resizing** — panels resize when the terminal is resized
- **Windows size polling** — on Windows, terminal size is polled every 100ms (CMD/PowerShell don't send SIGWINCH)
- **Minimum terminal size** — enforces 40x6 minimum dimensions
- **Height capping** — output is capped to terminal height to prevent top-cropping on small terminals

### Editor Area

- **Line numbers** — 4-digit right-aligned, styled in gray
- **Current line highlighting** — dark blue-gray background in insert mode, darker blue in normal mode
- **Block cursor** — cursor character highlighted with purple background and white foreground
- **Horizontal scroll** — lines longer than the viewport scroll horizontally
- **Vertical scroll** — viewport follows the cursor automatically

### Status & Help

- **Status bar** — shows mode indicator (`[NORMAL]`/`[INSERT]`/`[COMMAND]`/`[EXPLORER]`), filename, modified badge, and cursor position (`Ln X/Y, Col Z`)
- **Context-aware help bar** — bottom line showing relevant shortcuts for the current panel and mode
- **Context-aware help guide** — full command reference popup (Ctrl+H) that reorders sections based on current context, marking the active section with a "current" badge

### Explorer Area

- **Tree icons** — `▼` expanded, `▶` collapsed, space for files
- **Indentation** — 2 spaces per depth level
- **Long name truncation** — names are truncated with `...` to fit panel width
- **Modified indicator** — `[Modified]` badge in editor title when buffer has unsaved changes

### Theme

Dark theme with purple accent (`#7C3AED`), inspired by Catppuccin:

| Element | Color |
|---------|-------|
| Accent / cursor / borders | `#7C3AED` |
| Files | `#E2E8F0` |
| Directories | `#60A5FA` |
| Line numbers | `#6B7280` |
| Help text | `#9CA3AF` |
| Errors | `#EF4444` |
| Success | `#10B981` |
| Modified | `#F59E0B` |
| Status bar | `#A0AEC0` on `#1E1E2E` |
| Dialog background | `#1E1E2E` |

---

## Editor Core

| Feature | Description |
|---------|-------------|
| **Open file** | Reads file line-by-line; handles non-existent files as new empty buffers |
| **Save file** | Writes buffer to disk with newline terminators; creates parent directories if needed |
| **Save As** | Save buffer to a new path |
| **Insert character** | Single character insertion at cursor position |
| **Tab to spaces** | Tab key inserts 4 spaces |
| **Insert newline** | Splits line at cursor, creates new line below |
| **Backspace** | Deletes character before cursor; merges lines at column 0 |
| **Delete forward** | Deletes character at cursor; merges with next line at end of line |
| **Cursor movement** | Full directional movement with line wrapping |
| **Column clamping** | Cursor column is always clamped to line length when moving between lines |
| **Auto-scroll to cursor** | Viewport follows the cursor on every movement |
| **Go to line** | Programmatic jump to a specific line number |
| **Search** | String search with wrap-around from cursor position |
| **Find next** | Finds and moves cursor to next occurrence |
| **Modified tracking** | Tracks unsaved changes since last save |

---

## Configuration

| Feature | Description |
|---------|-------------|
| **Config path** | `~/.tuidit/config.json` (directory auto-created) |
| **Last workspace** | Persisted in `~/.tuidit/last_workspace`; validates path on load |
| **Default settings** | TabSize=4, ShowLineNumbers=true, WordWrap=false, AutoSave=false, AutoIndent=true, ExplorerWidth=30 |
| **Keybinding fields** | Config struct supports custom keybindings (KeySave, KeyQuit, KeyOpen, etc.) |

---

## File Operations

All file operations are available through the explorer panel and dialog system:

| Operation | Description |
|-----------|-------------|
| **Create file** | Creates file with parent directory creation; errors if already exists |
| **Create directory** | Creates directory recursively; errors if already exists |
| **Delete file** | Removes a single file |
| **Delete directory** | Recursively removes directory and all contents |
| **Rename** | Renames file or directory; creates parent directories for destination |
| **Move** | Moves file or directory (via cut + paste) |
| **Copy file** | Copies file content to a new location |
| **Copy directory** | Deep recursive copy preserving directory structure and permissions |
| **Path utilities** | Home directory expansion, absolute path resolution, file extension extraction, valid filename checking |
