# Tuidit – Interview Q&A

---

## Q1. What is Tuidit and what problem does it solve?

**Answer:**

**Tuidit** is a **terminal UI (TUI) based CLI code editor** written in Go. It runs entirely inside the terminal and provides a lightweight, keyboard-driven way to browse and edit code without leaving the command line.

**Problems it solves:**

1. **Quick edits without leaving the terminal** — When you’re already in a terminal (SSH, WSL, or local), you can open a file or folder with `tuidit` and edit in place instead of switching to a GUI editor or a separate window.

2. **Lightweight, no GUI required** — It doesn’t depend on a desktop environment or a heavy IDE. It’s useful on servers, minimal setups, or when you want a fast, low-resource editor.

3. **Familiar workflow** — It remembers the last workspace (like VS Code/Cursor) so running `tuidit` with no arguments reopens your last directory. You can also open any path: `tuidit /path/to/project` or `tuidit file.go`.

4. **Integrated file explorer and editor** — It combines a collapsible file tree (create/rename/delete, cut/copy/paste) with a Vim-like editor (Normal/Insert modes, line numbers, Ctrl+S to save), so you can navigate and edit in one tool.

5. **Portable and easy to run** — A single Go binary can be added to PATH and run from any directory, with support for Linux, macOS, and Windows.

In short, Tuidit solves the need for a **fast, terminal-native code editor** that fits into a CLI workflow and works in environments where a full IDE or GUI isn’t available or desired.

---

## Q2. Why did you build a terminal-based code editor instead of a GUI editor?

**Answer:**

I built a TUI-based code editor because it is **lightweight** and can be used in **small or embedded systems**, unlike GUI-based code editors, which are **system-heavy**.

- **Lightweight** — A terminal UI runs inside the existing terminal, with no windowing system or graphical toolkit. The binary is small, startup is fast, and it uses minimal memory and CPU compared to a full IDE or GUI editor.

- **Usable in constrained environments** — It can run on embedded Linux, headless servers, SSH sessions, minimal installs, or any environment that has a terminal but not a full desktop. GUI editors need a display server, more RAM, and more resources, so they aren’t practical there.

- **GUI editors are system-heavy** — They depend on display servers, graphical libraries, and heavier runtimes, which makes them unsuitable for resource-limited or embedded setups. A TUI editor avoids that overhead while still giving a structured editing experience.

So the choice was intentional: a terminal-based editor for **low resource usage** and **use in small or embedded systems**, where a GUI editor would be too heavy.

---

## Q3. What are the main features of Tuidit compared to traditional editors like Vim or Nano?

**Answer:**

Compared to traditional terminal editors like Vim or Nano, Tuidit adds a **project-centric UI** and **discoverability** while staying in the terminal:

| Area | Tuidit | Vim | Nano |
|------|--------|-----|------|
| **File explorer** | Built-in collapsible tree, create/rename/delete, cut/copy/paste | Netrw or plugins | None |
| **Last workspace** | Reopens last directory by default | Needs session/plugins | No |
| **Shortcut help** | Ctrl+H context-aware guide | No built-in cheat sheet | Bottom bar hints only |
| **Editing** | Vim-like (Normal/Insert) + Ctrl+S / Ctrl+O | Full Vim, steeper learning curve | Simple, no modes |
| **Layout** | Resizable explorer + editor panels | Single buffer / splits | Single buffer |

**1. Integrated file explorer (vs Vim/Nano)**  
Vim has netrw or relies on plugins for a file tree; Nano has no built-in tree. Tuidit has a **collapsible file-explorer panel** side by side with the editor: you navigate the folder tree with the keyboard, create/rename/delete files and folders, and use cut/copy/paste (e.g. Ctrl+X/C/V) without leaving the app. So you get an IDE-like project view without plugins.

**2. Last-workspace restoration (vs Vim/Nano)**  
Running `tuidit` with no arguments **reopens the last used directory**, similar to VS Code or Cursor. Vim and Nano don’t do this by default—you typically pass a path or rely on session plugins. Tuidit makes “continue where you left off” built-in.

**3. Context-aware shortcut guide (vs Vim/Nano)**  
**Ctrl+H** opens a **shortcut guide** that shows the main keybindings for the current context (explorer vs editor, normal vs insert). Nano shows a few hints at the bottom; Vim has no built-in cheat sheet. Tuidit makes the keybindings discoverable without leaving the editor.

**4. Familiar editing with less to learn (vs Vim)**  
The editor is **Vim-like** (Normal/Insert modes, motions like j/k/h/l, d, x, etc.) so it feels familiar to Vim users, but the feature set is focused. **Ctrl+S** to save and **Ctrl+O** to open are consistent across modes, which is easier for people used to GUI editors than pure Vim.

**5. Resizable panels and open-from-anywhere**  
You can **resize the explorer** (e.g. Ctrl+Left/Right). You can also add `tuidit` to PATH and run it from any directory with `tuidit` or `tuidit /path/to/project`, giving a single-tool workflow for opening and editing projects.

**Summary:** Compared to Vim, Tuidit offers a **built-in file tree, last-workspace memory, and an in-app shortcut guide** without needing plugins. Compared to Nano, it offers **modal editing, a full file explorer, workspace memory, and a structured two-panel TUI**, while remaining a lightweight terminal editor.

---

## Q4. What is the architecture of Tuidit? How are its modules structured?

**Answer:**

Tuidit uses a **modular architecture** with a clear separation of concerns. The app is structured as a thin entry point that creates a single TUI “model,” which composes several internal packages. The UI and event loop are driven by **Bubble Tea** (Charm Bracelet).

**Directory layout**

| Layer | Path | Role |
|-------|------|------|
| **Entry point** | `cmd/editor/main.go` | Parses CLI args, resolves path, loads last workspace via config, creates the TUI, runs `tea.NewProgram(app)`. |
| **Model** | `internal/model/` | Core data structures (`types.go`: `TreeNode`, `FileNode`, `Cursor`, `EditorBuffer`, `Dialog`, etc.) and **application state** (`app.go`: `AppState` — focus, mode, root path, buffers, dimensions, dialog). |
| **TUI** | `internal/tui/` | Main Bubble Tea **Model**: composes state, explorer, editor, config, and file ops; implements `Update()` (key/mouse/window size) and `View()` (rendering); holds styles (Lipgloss). Input handling lives in `handlers.go`. |
| **Explorer** | `internal/explorer/` | **File tree**: loads directory tree, expand/collapse, selection, fs watcher; exposes `FileTree` and visible nodes. |
| **Editor** | `internal/editor/` | **Text editing**: open/save file, buffer, cursor, scroll, Vim-like Normal/Insert actions (insert, delete, move, etc.). |
| **Config** | `internal/config/` | **Configuration** and **last-workspace** persistence (e.g. where to reopen on `tuidit` with no args). |
| **Utils** | `internal/utils/` | **File operations** used by the app (e.g. path expansion, create/rename/delete for explorer and dialogs). |

**How the modules fit together**

- **`main.go`** does not contain UI logic. It: reads args → gets/validates path (and optional last workspace from **config**) → creates **TUI** → runs the Bubble Tea program.
- The **TUI** struct holds: `State` (model), `Config`, `FileTree` (explorer), `Editor`, and `FileOps` (utils). So the TUI is the single place that ties **state**, **explorer**, **editor**, and **file ops** together.
- **Update flow:** Key/mouse/window messages go to TUI’s `Update()`. The TUI (and `handlers.go`) decide what to do and then call into **FileTree** (e.g. open folder, expand node) or **Editor** (e.g. open file, type, save). **Model** types are used everywhere; **AppState** is updated (focus, mode, dimensions, dialog).
- **View flow:** TUI’s `View()` asks **FileTree** for the visible tree and **Editor** for the current buffer/lines and cursor, then renders explorer panel + editor panel + status/dialogs using **Lipgloss** styles.
- **Config** and **utils** are used at startup and when opening/saving files or performing file operations from the explorer/dialogs.

So the architecture is **modular**: each package has a single responsibility (model = state/types, explorer = file tree, editor = text editing, tui = orchestration + rendering, config = settings, utils = file I/O helpers), and the **TUI layer** is the only one that composes them and talks to Bubble Tea.

---

## Q5. What is Bubble Tea? (with sample code)

**Answer:**

**Bubble Tea** is a Go TUI (terminal user interface) framework from **Charm Bracelet**. It lets you build interactive terminal apps (keyboard/mouse, redraws, multiple “screens”) instead of simple line-based CLIs. It’s inspired by **The Elm Architecture**: your app is a **model** (state), and you implement **Init**, **Update**, and **View** so the framework can run the event loop and render the UI.

**Core ideas**

- **Model** — A value (usually a struct) that holds all app state. It must implement the Bubble Tea `Model` interface.
- **Init()** — Called once at start; can return a `tea.Cmd` to run (e.g. subscribe to events). Return `nil` for no initial command.
- **Update(msg)** — Called for every event (key press, mouse, window resize, custom messages). Returns the updated model and an optional `tea.Cmd` (e.g. for async work).
- **View()** — Returns a string that is drawn in the terminal. You build this string from the current model (often using Lipgloss for styling).

You run the app with `tea.NewProgram(model).Run()`; Bubble Tea handles the event loop and calls `Update` / `View` as needed.

**Minimal sample: counter**

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	count int
}

func (m model) Init() tea.Cmd {
	return nil // no initial command
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.count++
		case "down", "j":
			m.count--
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("\n Count: %d\n\n (↑/k up, ↓/j down, q quit)\n", m.count)
}

func main() {
	p := tea.NewProgram(model{count: 0})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

**How Tuidit uses Bubble Tea**

Tuidit’s main “app” is a Bubble Tea model (`TUI` in `internal/tui/tui.go`). It implements `Init`, `Update`, and `View` and is run from `cmd/editor/main.go` like this:

```go
app := tui.NewTUI()
// ... set app.State, load directory/file ...

p := tea.NewProgram(
	app,
	tea.WithAltScreen(),       // full-screen TUI
	tea.WithMouseCellMotion(), // mouse support
)
if _, err := p.Run(); err != nil {
	// handle error
}
```

- **Init()** returns `tea.EnterAltScreen` (and optionally a file-watcher command).
- **Update()** switches on `tea.KeyMsg`, `tea.MouseMsg`, `tea.WindowSizeMsg`, and custom messages (e.g. `DirChangedMsg`), then updates state and calls into the explorer/editor.
- **View()** builds the full UI string (file tree + editor panel + status bar + dialogs) from `State`, `FileTree`, and `Editor`, using Lipgloss for borders and colors.

So Bubble Tea is the **runtime** for Tuidit: it owns the event loop and redraws; Tuidit just implements the model and the three methods.

---

## Q6. How does the file management system work in your editor?

**Answer:**

File management in Tuidit is split into **three parts**: an in-memory **file tree** (explorer), a **file operations** layer (utils), and **TUI logic** that wires keys/dialogs to those layers and keeps the tree in sync.

**1. File tree (explorer)**

- **`internal/explorer/filetree.go`** defines a `FileTree` with a `Root` (`*model.TreeNode`) and `RootPath`.
- **Loading a directory:** `LoadDirectory(path)` resolves to an absolute path, creates a root node, and loads its children with `loadChildren()`: `os.ReadDir`, then sort (directories first, then alphabetically), skip hidden (`.`), and build `TreeNode` entries (Name, Path, Type, Parent, Expanded, IsLoaded). So the tree is **lazy**: children are loaded when you first expand a folder.
- **Expand/collapse:** `ToggleNode` / `ExpandNode` load children if not yet loaded, then set `Expanded`. `CollapseNode` just sets `Expanded = false`.
- **Refresh:** `Refresh()` / `RefreshNode()` re-read the directory and rebuild children while **preserving expansion state** for existing paths, so the UI doesn’t collapse when you add/remove files.
- **Visibility:** `GetVisibleNodes()` does a depth-first walk and returns only nodes that are visible (expanded parents), which the TUI uses to render the explorer list.
- **Filesystem watcher:** `StartWatch(rootPath)` uses **fsnotify** to watch the root and all subdirs. On events (debounced ~150 ms) it sends on a channel; `WatchCmd()` turns that into a Bubble Tea command that produces `model.DirChangedMsg`. When the TUI receives `DirChangedMsg`, it calls `FileTree.Refresh()` and restarts the watcher so the tree stays in sync with external changes (e.g. git checkout, another editor).

**2. File operations (utils)**

- **`internal/utils/fileops.go`** defines `FileOperations` with methods that perform actual I/O:
  - **CreateFile(path)** — `MkdirAll` for parent, then `os.Create`.
  - **CreateDirectory(path)** — `os.MkdirAll`.
  - **DeleteFile** / **DeleteDirectory** — `os.Remove` / `os.RemoveAll`.
  - **RenameFile(old, new)** — ensure destination doesn’t exist, `MkdirAll` for new’s parent, then `os.Rename` (works for move within same filesystem).
  - **CopyFile(src, dst)** — read all, `MkdirAll` for dst parent, write.
  - **CopyDirectory(src, dst)** — recursive: create dst dir, then for each entry either `CopyDirectory` or `CopyFile`.
- Helpers: **ExpandPath** (e.g. `~` → home), **IsValidFileName**, **ListDirectory**, etc. The editor and TUI use these when resolving paths and validating names in dialogs.

**3. TUI: dialogs and clipboard**

- **New file/folder:** User presses `n` / `N` in the explorer; TUI opens a dialog (e.g. `DialogNewFile` / `DialogNewFolder`), user types a name and confirms. Handler resolves the path (current node or parent), then calls `FileOps.CreateFile` or `FileOps.CreateDirectory`, then `FileTree.Refresh()` and updates `visibleNodes` and selection.
- **Rename:** F2 → `DialogRename` with current name; on confirm, `FileOps.RenameFile(oldPath, newPath)`. If the editor has that file open, the buffer’s path is updated. Then refresh the tree.
- **Delete:** Delete key → `DialogDelete` (confirm). On confirm, `FileOps.DeleteFile` or `FileOps.DeleteDirectory`. If the deleted path was open in the editor, that’s handled (e.g. clear or switch buffer). Then refresh the tree.
- **Cut / Copy / Paste:** TUI keeps `clipboardPath` and `clipboardCut`. **Cut** (Ctrl+X) and **Copy** (Ctrl+C) store the selected node’s path and a flag. **Paste** (Ctrl+V) uses `getPasteDestDir()` (selected folder, or parent of selected file, or tree root), then:
  - If **cut:** `FileOps.RenameFile(clipboardPath, destPath)` (move); if the moved file was open, update the editor buffer’s path; clear clipboard.
  - If **copy:** `FileOps.CopyFile` or `FileOps.CopyDirectory`; no change to editor; clear clipboard.
  After paste, `FileTree.Refresh()` and visible nodes/selection are updated. Guards: no paste into same path, no copy of a directory into itself.

**End-to-end flow**

1. User opens a directory (e.g. `tuidit /path`) or uses “Open folder” → `LoadDirectory` and optionally `StartWatch` + `WatchCmd()`.
2. User selects a node; keys trigger dialogs or clipboard. On confirmation, **only** `FileOperations` does disk I/O.
3. After any create/rename/delete/cut/copy/paste, the TUI calls **`FileTree.Refresh()`** (and restarts the watcher if needed), then recomputes `visibleNodes` so the explorer panel matches the filesystem.

So the **file management system** is: **explorer** = in-memory tree + lazy load + refresh + fsnotify; **utils** = all create/delete/rename/copy/move; **TUI** = dialogs, clipboard, and calling those two and refreshing the tree after every change.

---

## Q7. How do you initiate (initialize) Go and Rust projects?

**Answer:**

**Go**

1. Create a directory and go into it:
   ```bash
   mkdir myapp && cd myapp
   ```
2. Initialize the Go module (this creates `go.mod`). The module path is usually your repo path or a simple name:
   ```bash
   go mod init myapp
   # or: go mod init github.com/username/myapp
   ```
3. Add a main package, e.g. `main.go`:
   ```go
   package main
   func main() { println("hello") }
   ```
4. Run or build:
   ```bash
   go run .
   go build -o myapp .
   ```
5. Add dependencies by importing in code and running:
   ```bash
   go mod tidy
   ```
   Or add a specific package:
   ```bash
   go get github.com/some/package
   ```

**Rust**

1. Create a new project (also creates a git repo by default):
   ```bash
   cargo new myapp        # binary (has main)
   cargo new mylib --lib  # library (has lib.rs)
   ```
   To create inside the current directory without a new folder:
   ```bash
   cargo init        # binary in current dir
   cargo init --lib   # library in current dir
   ```
2. This creates:
   - `Cargo.toml` — project name, version, dependencies
   - `src/main.rs` (binary) or `src/lib.rs` (library)
3. Run or build:
   ```bash
   cargo run
   cargo build          # debug build
   cargo build --release # release build
   ```
4. Add dependencies: edit `Cargo.toml` under `[dependencies]`, e.g.:
   ```toml
   [dependencies]
   serde = "1.0"
   ```
   Then `cargo build` will fetch and compile them.

**Summary**

| Step        | Go                    | Rust                          |
|------------|------------------------|-------------------------------|
| Init       | `go mod init <module>` | `cargo new <name>` / `cargo init` |
| Config file| `go.mod` (+ `go.sum`)  | `Cargo.toml`                  |
| Entry      | `main.go` (package main) | `src/main.rs` or `src/lib.rs` |
| Run        | `go run .`             | `cargo run`                   |
| Build      | `go build -o out .`    | `cargo build` / `cargo build --release` |
| Deps       | `go get pkg` / `go mod tidy` | Add in `Cargo.toml`, then `cargo build` |

---

## Q8. Why did you choose Go over Rust (for Tuidit)?

**Answer:**

I chose Go over Rust for this project for a few practical reasons:

**1. TUI ecosystem**  
I wanted a terminal UI with panels, keyboard/mouse, and a clear model–update–view style. **Bubble Tea** (Charm Bracelet) in Go is a mature, well-documented framework with that exact design. Rust has solid options (e.g. ratatui, crossterm), but for this kind of app the Go TUI stack (Bubble Tea + Lipgloss) was a strong fit and had more examples and tutorials for building editor-like UIs quickly.

**2. Faster iteration**  
Go’s simpler memory model and lack of a borrow checker made it easier to prototype and change the UI and file logic without fighting the type system. For a terminal editor where correctness matters but isn’t at the level of a kernel or safety-critical system, Go’s trade-off favored development speed.

**3. Concurrency that matches the design**  
Bubble Tea is message-driven (events → `Update` → new model). We have a filesystem watcher that sends events into the same loop. Go’s goroutines and channels map naturally to that: the watcher runs in a goroutine and sends on a channel; the TUI turns that into a `tea.Cmd`. Rust can do the same, but the Go version was straightforward to wire up without async/await or more complex concurrency.

**4. Single binary and cross-compilation**  
Go’s `GOOS`/`GOARCH` cross-compilation and static binaries by default made it easy to target Linux, macOS, and Windows from one codebase and ship a single executable. Rust can do this too, but Go’s tooling and defaults are very simple for this use case.

**5. Simpler tooling and onboarding**  
One standard formatter (`gofmt`), one main build/test flow (`go build`, `go test`), and a small, readable standard library made it easy to keep the project minimal and understandable. That helped for a side project where the goal was to ship a working editor, not to maximize runtime performance or formal safety.

I’d consider Rust for a project that needed maximum performance, no GC, or stronger safety guarantees; for a TUI code editor, Go’s ecosystem and productivity were the right fit for me.

---

## Q9. What is the distribution size of Tuidit?

**Answer:**

Tuidit is intentionally **small and self-contained**:

- The **statically-linked binary** for each platform is typically around **4–6 MB** (Linux, macOS, Windows, amd64/arm64).
- When compressed in a `.tar.gz` or `.zip`, the download size is usually around **2–3 MB**.
- There are **no external runtime dependencies** (no Node, no Python, no JVM). You just download the binary, make it executable, and add it to your PATH.

From a distribution perspective this is closer to shipping a **single small CLI tool** than deploying a full IDE. It’s fast to download even on slow networks and cheap to distribute via GitHub Releases, package managers (Homebrew, APT, DNF), or copying the binary over SSH.

---

## Q10. How is Tuidit better or different than Vim, Nano, and other existing CLI editors?

**Answer:**

“Better” depends on the use case, but Tuidit is designed to be **more approachable and project-centric** than traditional editors while staying terminal-native:

1. **Built-in project view, no plugins required**  
   Editors like Vim rely on plugins (e.g. netrw, NERDTree) for a file tree; Nano has no project view at all. Tuidit ships with a **first-class file explorer panel**: you can see the whole project, expand folders, and do file operations with obvious keys (n/N, F2, Del, Ctrl+X/C/V).

2. **Discoverable keybindings**  
   Vim is extremely powerful but has a steep learning curve and expects you to memorize commands or use external cheat sheets. Nano shows some shortcuts but is limited. In Tuidit, **Ctrl+H** always brings up a **context-aware shortcut guide**, which makes it easier for new users to become productive quickly.

3. **Opinionated, simpler Vim-like editing**  
   Tuidit borrows the parts of Vim that help most (Normal/Insert modes, hjkl navigation, word/line motions) and combines them with more familiar shortcuts like **Ctrl+S** to save and **Ctrl+O** to open. This gives a **gentler on-ramp** for people coming from GUI editors who still want modal editing.

4. **IDE-style workflow in the terminal**  
   By combining a project tree, editor panel, status bar, and shortcut guide, Tuidit feels closer to a **minimal terminal IDE** than a single-buffer editor. For quick edits, Vim or Nano are great; for **browsing and working in an entire project** from the terminal, Tuidit provides more structure out of the box.

5. **Last workspace and auto-refresh**  
   Tuidit remembers your **last workspace** and automatically reopens it, and it can **auto-refresh the file tree** when files change on disk (fsnotify + polling for WSL). This makes it behave more like modern editors (VS Code, Cursor) while still running purely in the terminal.

In short, Tuidit is not trying to replace Vim for power users, but to offer a **small, opinionated, project-focused editor** that feels familiar to GUI-editor users and productive for everyday terminal work.

---

## Q11. What is Lipgloss and how does Tuidit use it?

**Answer:**

**Lipgloss** is a Go library (also from Charmbracelet) for **styling and laying out text in the terminal**. It’s like a CSS box model for terminal UIs:

- You define reusable **styles** (`lipgloss.Style`) with colors, borders, padding, and alignment.
- You call `.Render(text)` on a style to get a **string with ANSI escape codes**.
- You arrange those strings with helpers like `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical` to build panels and full-screen layouts.

Lipgloss is **purely presentational**: it doesn’t know about keys, state, or files. Bubble Tea handles the event loop and messages; Lipgloss turns the current model into a styled string that you see in the terminal.

**Core Lipgloss ideas:**

- **Styles:** foreground/background colors, bold/italic/underline, borders (rounded, thick, etc.), padding, margins.
- **Sizing:** `Width(n)` / `Height(n)` to pad or truncate content to a given size.
- **Layout:** `JoinHorizontal` and `JoinVertical` to stitch multiple blocks together (e.g. explorer + editor, then status bar + help bar).

In Tuidit (`internal/tui/tui.go`), Lipgloss is used to define the **visual identity and layout**:

1. **Global theme styles**  
   Tuidit centralizes styles like:
   - `titleStyle` — purple, bold text with padding (used for `" Explorer "` and file titles).
   - `panelStyle` — rounded gray border with inner padding (used for both panels).
   - `activePanelStyle` — same border but with a purple border color to highlight the focused panel.
   - `fileStyle` / `dirStyle` / `selectedStyle` — different colors and boldness for files, directories, and the selected row in the explorer.
   - `statusStyle` / `helpStyle` — colors for the bottom status line and help bar.
   - `dialogBoxStyle` — bordered, padded box for dialogs (e.g. open file, rename, delete, Ctrl+H guide).

2. **Panels and layout**  
   In `renderMain()`, Tuidit:
   - Computes the **explorer width**, **editor width**, and **height** from the current terminal size.
   - Renders explorer and editor content as plain lines, then wraps them with `panelStyle` / `activePanelStyle` using `Width(height)` + `Height(height)`.
   - Joins them side-by-side with `lipgloss.JoinHorizontal(...)` and then stacks `statusBar` and `helpBar` underneath with `lipgloss.JoinVertical(...)`.

   This is how the UI gets a consistent two-panel layout with borders that adapt to terminal resizes.

3. **Highlighting and readability**  
   When rendering explorer nodes, Tuidit chooses a style based on type and selection:
   - Selected row → `selectedStyle.Render(text)` (purple background, bold).
   - Directory → `dirStyle.Render(text)` (blue, bold).
   - File → `fileStyle.Render(text)`.

   Similar style decisions are made for editor titles (e.g. `[Modified]`), status messages, and dialogs.

4. **Dialogs and overlays**  
   For dialogs, Tuidit:
   - Renders the base screen with `renderMain()`.
   - Renders the dialog content with `dialogBoxStyle.Render(...)`.
   - Overlays dialog lines onto the main view by computing a center position and splicing the dialog’s strings into the main string.

So, **Bubble Tea** gives Tuidit the **model–update–view loop**, and **Lipgloss** gives it the **look and layout**: borders, colors, padding, and how panels and bars fit together for a modern, polished TUI inside the terminal.

---

## Q12. Why is the Go distribution (binary) of Tuidit larger than a Rust build of the same editor?

**Answer:**

The short version is that **Go ships a bigger runtime and tends to statically link more code by default**, while **Rust has a much thinner runtime and relies on LLVM to strip away more unused code**, often with dynamic linking to system libraries. Both produce “single binaries”, but the contents are very different.

Let’s break it down:

**1. Go embeds a full runtime (GC + scheduler + reflection)**  
When you build Tuidit in Go, the binary includes:

- A **garbage collector** (heap scanner, write barriers, safepoints).
- A **goroutine scheduler** (M:N scheduling, timers, stack management).
- **Runtime type information** and reflection support for interfaces, maps, slices, errors, etc.
- Panic/recover, stack traces, and other runtime helpers.

Even if Tuidit doesn’t use every feature explicitly, this machinery is always there for non-trivial Go programs. That baseline alone is 1–2 MB+ before you add your own code.

Rust, by design, has **no GC** and a very small runtime: ownership and lifetimes are enforced at compile time, so there’s no need to ship a collector or scheduler in the binary. If you don’t use threads, async runtimes, or unwinding, a lot of runtime code simply isn’t linked at all.

**2. Static vs dynamic linking defaults**  

- Go typically produces **statically linked binaries** for pure-Go code on Linux/macOS. That means:
  - Your binary carries its own implementations of low-level routines instead of depending on the system’s libc and friends.
  - Distribution is trivial (copy one file, run it), but the file is larger because nothing is shared with the OS.

- Rust binaries often rely on **dynamic linking** to system libraries by default (e.g. glibc). That means:
  - Some code lives in shared libraries that are already present on the system.
  - The Rust executable itself can be smaller, because it doesn’t embed everything it needs.

If you build Rust statically (e.g. against musl), the size gap between Rust and Go shrinks, but Rust still usually wins because of the smaller runtime and more aggressive dead-code elimination.

**3. Different optimization pipelines (gc vs LLVM)**  

- Go’s `gc` compiler is optimized for **fast builds and simple tooling**, not maximum code shrink. It does SSA, inlining, and dead-code elimination, but it doesn’t run the deep set of whole-program optimizations that LLVM does.
- Rust uses **LLVM**, which has very strong optimizations, especially with `-C opt-level=3` and **LTO/ThinLTO**:
  - Unused functions and types across crates can be stripped.
  - Cross-crate inlining, constant folding, and devirtualization help remove entire branches of code that are never reached.

For a TUI app like Tuidit, that means the Rust compiler can more aggressively throw away unused pieces of libraries, while Go tends to keep a bit more “just in case”.

**4. Library ecosystem and what Tuidit pulls in**  
Tuidit in Go uses **Bubble Tea**, **Lipgloss**, **fsnotify**, and standard library packages for terminal control, styling, Unicode, filesystem watching, and more. All of that code (plus their dependencies) is compiled into one static binary.

An equivalent Rust TUI app might use `ratatui` + `crossterm` or similar. Those crates are usually designed to be **feature-flagged and minimal**, and LLVM can aggressively prune code paths you don’t touch. Combined with the thinner runtime, the resulting Rust binary is often **1–3 MB where Go might be 4–6 MB** for a similar app.

**5. Why this trade-off is acceptable for Tuidit**  

In practice:

- A **4–6 MB** Go binary is still very small by modern standards.
- Static linking and the larger runtime make distribution **dead simple**: one file, no “missing DLL/so” issues across Linux/macOS/Windows/WSL.
- The Go ecosystem (Bubble Tea + Lipgloss) let Tuidit ship faster, and the cost is just a couple of extra megabytes on disk.

So, the Go build of Tuidit is larger mostly because it **includes more runtime machinery and is more static by default**. A Rust port could produce a smaller binary, but at the cost of a different ecosystem and more complex distribution story, while the current size is already perfectly acceptable for a terminal editor.

---

## Q13. How do you handle opening, editing, and saving files internally?

**Answer:**

Internally, Tuidit separates **filesystem I/O** from **in‑memory editing**, so the editor always works on a safe buffer and only touches disk when you explicitly save.

**1. Opening files**

There are two main flows:

- **From the CLI:**  
  - If you run `tuidit path/to/file.go`, `cmd/editor/main.go` detects that the path is a file.  
  - It calls `Editor.OpenFile(absPath)` and also loads the parent directory into the explorer so the project tree is available.

- **From the explorer panel:**  
  - You move the selection to a file and press `Enter`.  
  - The TUI handler (`internal/tui/handlers.go`) gets the selected `TreeNode` and calls `Editor.OpenFile(node.Path)`.

Inside `Editor.OpenFile` (simplified):

1. Read the file via the `FileOperations` helper (backed by `os.ReadFile`).
2. Split the contents into a `[]string` of lines.
3. Create an `EditorBuffer` with:
   - `Lines []string`
   - `Cursor { Line, Col }`
   - `ScrollY` (top visible line)
   - `Path` and `Modified` flag
4. Set this buffer as the active buffer in the editor, and track it in the list of open buffers.

At this stage, **all editing happens in memory**; the on‑disk file is untouched.

**2. Editing in memory**

Tuidit’s editor is **Vim‑like** (Normal/Insert modes), but under the hood it’s just manipulating the `EditorBuffer`:

- **Modes** live in `AppState.Mode` (Normal vs Insert).  
  Key handlers decide whether to interpret keys as commands or as literal text.

- **Cursor and scrolling:**  
  Movements (`h/j/k/l`, arrows, `w/b/0/$/g/G`) update `Buffer.Cursor`.  
  `Editor.ScrollToCursor(viewportHeight)` adjusts `ScrollY` so the cursor stays in the visible window.

- **Text operations:**  
  All edits are transformations on `Lines`:
  - Inserting a character: modify `Lines[cursor.Line]` at `cursor.Col`, then advance the column.
  - Inserting a newline: split the current line into two entries in `Lines` and move the cursor to the new line.
  - Deleting a line (`d`): remove one element from `Lines` and clamp cursor/scroll.
  - Deleting a character (`x`/`X`): remove from the string and adjust indexes.

Every change:

- Marks the buffer as **`Modified = true`**.
- Leaves the filesystem alone until you explicitly save.
- Is reflected in the TUI through `Editor.GetVisibleLines(height)` + `Cursor` and `ScrollY` when `renderEditor` runs.

**3. Saving files**

Save is always an explicit action (e.g. **Ctrl+S**):

1. The TUI sees the `Ctrl+S` key in the editor context.
2. It calls `Editor.SaveActiveBuffer()` (or equivalent), which:
   - Joins `Lines` with `\n` into a single byte slice.
   - Uses `FileOperations` to write to disk (backed by `os.WriteFile(path, data, mode)`).
   - Clears the buffer’s `Modified` flag.
3. The status bar is updated via `AppState.StatusMessage` to show feedback like “Saved”.

If the file was **renamed or moved** from the explorer:

- The explorer uses `FileOperations.RenameFile` or move helpers.
- The TUI updates any open `EditorBuffer.Path` that pointed to the old location.
- Subsequent saves go to the new path.

**4. Keeping explorer and editor in sync**

Tuidit keeps the file tree and buffers consistent:

- **Open file reuse:**  
  If you open a file that already has a buffer, the editor reuses that buffer instead of re-reading from disk, preserving unsaved changes.

- **File operations from the explorer:**  
  The explorer uses `FileOperations` to create/rename/delete/copy/move paths.  
  After each operation, `FileTree.Refresh()` is called to rebuild the tree while preserving expansion state.  
  If an edited file is deleted or moved, the TUI updates or closes the corresponding buffer.

- **External changes:**  
  A filesystem watcher (fsnotify on native Linux/macOS, plus polling fallback for WSL) sends `DirChangedMsg` into the Bubble Tea loop.  
  On that message, Tuidit calls `FileTree.Refresh()` and restarts the watcher, so the explorer reflects external edits, git checkouts, or newly created files. The editor buffers remain as-is until you explicitly reopen or reload, to avoid losing unsaved edits.

In summary, Tuidit’s flow is:

- **Open** → read once into an in‑memory `EditorBuffer` (lines + cursor + scroll).
- **Edit** → apply all changes to the buffer only; keep track of `Modified`.
- **Save** → on demand, write the buffer to disk via `FileOperations` and update UI status.

---

## Q14. How do you manage cursor movement and text editing operations?

**Answer:**

Cursor movement and text editing are handled entirely in an **in‑memory buffer model**, and all key events are interpreted through Bubble Tea’s `Update` loop.

**1. Buffer and cursor model**

Each open file has an `EditorBuffer` (see `internal/model` / `internal/editor`), which contains:

- `Lines []string` — the file contents, one string per visual line.
- `Cursor` — a struct with `Line` and `Col` (0‑based indices).
- `ScrollY` — the index of the first visible line in the viewport.
- Metadata: `Path`, `Modified`, etc.

All movement and editing operations modify **only** these fields; rendering (`renderEditor`) reads them and draws the UI.

**2. Modes and key handling**

The editor is **modal**:

- `ModeNormal` — keys are interpreted as **commands** (motions, deletes, mode switches).
- `ModeInsert` — most keys are interpreted as **literal text input**.

In `internal/tui/handlers.go`, Tuidit:

- Receives `tea.KeyMsg` from Bubble Tea.
- Switches on `AppState.Mode` and current focus (explorer vs editor).
- For editor focus:
  - In Normal mode: maps keys like `h/j/k/l`, `w/b/0/$/g/G`, `d`, `x`, `i`, `a`, `o/O` to **cursor/motion/edit commands**.
  - In Insert mode: interprets printable keys as text insertion; Enter/Backspace/Delete, etc., have special handlers.

These handlers call into the editor package to mutate the buffer.

**3. Cursor movement**

Movement commands adjust `Cursor.Line` and `Cursor.Col` with clamping:

- **Horizontal**:
  - `h` / Left arrow → `Col--` (min 0).
  - `l` / Right arrow → `Col++` (max = `len(Lines[Line])`).
  - `0` / `^` → move to start of line.
  - `$` → move to end of line.

- **Vertical**:
  - `j` / Down arrow → `Line++` (max = `len(Lines)-1`).
  - `k` / Up arrow → `Line--` (min 0).
  - `g` / `G` → first/last line.

- **Word motions**:
  - `w` / `b` walk forwards/backwards over non‑space characters to find the next/previous word boundary within or across lines.

After any movement, the editor:

- Clamps the cursor to a valid column on the new line (no off‑by‑one at line ends).
- Calls `ScrollToCursor(viewportHeight)` so `ScrollY` is adjusted if the cursor goes off‑screen (top or bottom).

**4. Text editing operations**

Text edits are just **transformations on `Lines`** at the cursor:

- **Insert mode**:
  - Printable key: insert rune into `Lines[cursor.Line]` at `cursor.Col` and increment `Col`.
  - Enter: split the current line at `Col` into two lines (above/below), adjust `Line`/`Col`.
  - Backspace/Delete: remove characters before/under `Col`, merge lines when deleting at boundaries.

- **Normal mode commands**:
  - `d` (delete line): remove `Lines[cursor.Line]` from the slice and clamp `Line`/`ScrollY`.
  - `x` / `X`: delete the character under/before the cursor in the current line.
  - `o` / `O`: insert a new empty line below/above the current line, move cursor into Insert mode on that line.

Every edit:

- Updates `Lines` and `Cursor` (and sometimes `ScrollY`).
- Marks the buffer as **`Modified = true`**.
- Does **not** touch the on‑disk file until a save is triggered.

**5. Rendering and status**

When `renderEditor` is called:

- It asks the editor for visible lines via `GetVisibleLines(height)`, which slices `Lines[ScrollY : ScrollY+height]`.
- It uses the cursor position (`Cursor.Line`, `Cursor.Col`) to:
  - Highlight the current line.
  - Draw a visual cursor (via Lipgloss) at the right column.
- It also feeds cursor/position info into the status bar (`Ln`, `Col`) so the user sees where they are.

In summary, cursor movement and editing are implemented as **pure state updates** on an `EditorBuffer` (lines + cursor + scroll), driven by key events from Bubble Tea. The TUI is just a view over that state, and save operations are the only time changes are flushed back to disk.

---

## Q15. How does your editor handle large files efficiently?

**Answer:**

Today Tuidit is optimized for **interactive responsiveness** on typical source files (hundreds to low tens of thousands of lines), not for huge logs. It keeps a **line-based buffer in memory** and only **renders what’s visible**, which keeps the UI smooth for normal code editing.

**1. Line-based in-memory buffer**

- When you open a file, Tuidit reads it once from disk and splits it into `[]string` lines in an `EditorBuffer`.
- All navigation and editing operations work on this line slice:
  - Cursor movement just changes `Cursor.Line` / `Cursor.Col`.
  - Deletes/inserts adjust the relevant line(s) and occasionally the slice length.
- That means operations like moving the cursor, scrolling, or editing a single line are effectively **O(1)** relative to file size (except when you insert/delete whole lines, which may shift part of the slice).

For extremely large files (hundreds of MBs or millions of lines), this “load everything into memory” model will eventually hit memory and GC pressure, but for typical source code projects it’s a good trade‑off: simpler logic, predictable behavior, and low per‑operation cost.

**2. Viewport-based rendering**

The editor never tries to draw the entire file on screen:

- The buffer tracks a `ScrollY` (top visible line) and viewport height.
- `GetVisibleLines(height)` returns only `Lines[ScrollY : ScrollY+height]` (clamped).
- `renderEditor` uses just that slice to build the output string for the editor panel.

So, even if the file has 50,000 lines, each frame only processes **the subset that is visible** (plus some small overhead for borders and status/help bars). This keeps redraws fast because rendering work scales with the window height, not with file length.

**3. Input and scrolling**

- Keyboard events (like `j/k`, PageUp/PageDown) adjust `Cursor` and `ScrollY` and then trigger a re‑render.
- Since only the visible window is rendered and Bubble Tea handles diffing efficiently, scrolling through a large file feels smooth as long as the machine has enough RAM for the buffer itself.

**4. When this model is enough (and when it isn’t)**

- **Enough for:**  
  - Source trees (Go/Rust/TS/etc.).  
  - Config files, small/medium logs, documentation.  
  - Most day‑to‑day editing scenarios.

- **Limitations:**  
  - Very large logs or generated files (hundreds of MBs, millions of lines) will:
    - Take noticeable time to load initially (one big read + split).
    - Use a lot of RAM for the `[]string` slice.

In those extreme cases, tools specialized for log viewing (e.g. `less`, `tail`, `sed/awk`, or dedicated log viewers) are still a better fit. For Tuidit’s main target — **terminal‑native code editing with good UX** — the current approach (in‑memory lines + viewport rendering) gives a responsive experience without the complexity of a paged or chunked file representation.
