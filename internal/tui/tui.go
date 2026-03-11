package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
	"tuidit/internal/config"
	"tuidit/internal/editor"
	"tuidit/internal/explorer"
	gitutil "tuidit/internal/git"
	"tuidit/internal/model"
	"tuidit/internal/utils"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A5568")).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")).
				Padding(0, 1)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#7C3AED")).
			Bold(true)

	modifiedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0AEC0")).
			Background(lipgloss.Color("#1E1E2E"))

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2).
			Background(lipgloss.Color("#1E1E2E"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7C3AED")).
			Foreground(lipgloss.Color("#FFFFFF"))

	previewItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	previewDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA"))

	previewSelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#7C3AED"))

	gutterAddedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3FB950"))

	gutterModifiedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E3B341"))

	gutterDeletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F85149"))

	gitStatusModifiedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E3B341"))

	gitStatusAddedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3FB950"))

	gitStatusUntrackedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F0883E"))

	gitStatusDeletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F85149"))

	gitStatusConflictStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F472B6"))

	gitStatusRenamedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#39C5CF"))

	gitIgnoredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A5568"))
)

// TUI represents the main TUI application
type TUI struct {
	State       *model.AppState
	Config      *config.Config
	FileTree    *explorer.FileTree
	Editor      *editor.Editor
	FileOps     *utils.FileOperations
	
	// View state
	selectedIndex int
	visibleNodes  []*model.TreeNode
	
	// Dialog state
	dialogInput         string
	dialogMessage       string
	dialogError         string
	dialogPreviewScroll int

	// Clipboard for cut/copy/paste (explorer)
	clipboardPath string
	clipboardCut  bool

	// Mouse state for double-click detection
	lastClickTime  int64 // Unix nanoseconds
	lastClickX     int
	lastClickY     int
	lastClickPanel int // 0=none, 1=explorer, 2=editor

	// Git status cache for explorer
	gitStatus gitutil.RepoFileStatus
}

// NewTUI creates a new TUI application
func NewTUI() *TUI {
	return &TUI{
		State:     model.NewAppState(),
		Config:    config.DefaultConfig(),
		FileTree:  explorer.NewFileTree(),
		Editor:    editor.NewEditor(),
		FileOps:   utils.NewFileOperations(),
	}
}

func (t *TUI) refreshGitStatus() {
	if t.FileTree.RootPath != "" {
		t.gitStatus = gitutil.GetRepoStatus(t.FileTree.RootPath)
	} else {
		t.gitStatus = nil
	}
}

// getSizeCmd returns a command that reads terminal size and sends WindowSizeMsg (so height/width update without calling GetSize in View).
func getSizeCmd() tea.Cmd {
	return func() tea.Msg {
		w, h, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil || w <= 0 || h <= 0 {
			return nil
		}
		return tea.WindowSizeMsg{Width: w, Height: h}
	}
}

// sizePollInterval is how often we read terminal size so resize is reflected (e.g. when terminal doesn't send WindowSizeMsg).
const sizePollInterval = 100 * time.Millisecond

// Init initializes the TUI
func (t *TUI) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnterAltScreen, getSizeCmd()}

	// Windows CMD/PowerShell often don't send SIGWINCH, so poll terminal size there.
	// Linux/macOS terminals handle resize natively via SIGWINCH — polling there causes flicker.
	if runtime.GOOS == "windows" {
		cmds = append(cmds, tea.Tick(sizePollInterval, func(t time.Time) tea.Msg { return t }))
	}

	if t.FileTree.RootPath != "" {
		_ = t.FileTree.StartWatch(t.FileTree.RootPath)
		t.refreshGitStatus()
		cmds = append(cmds, t.FileTree.WatchCmd())
	}
	return tea.Sequence(cmds...)
}

// Update handles updates
func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return t.handleKeyPress(msg)
	case tea.MouseMsg:
		return t.handleMouse(msg)
	case tea.WindowSizeMsg:
		if msg.Width == t.State.Width && msg.Height == t.State.Height {
			return t, nil
		}
		t.State.Width = msg.Width
		t.State.Height = msg.Height
		// Keep explorer selection in view when height shrinks
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		if len(t.visibleNodes) > 0 {
			visibleHeight := t.State.Height - 4
			if visibleHeight < 1 {
				visibleHeight = 1
			}
			if t.selectedIndex >= len(t.visibleNodes) {
				t.selectedIndex = len(t.visibleNodes) - 1
			}
			if t.selectedIndex < 0 {
				t.selectedIndex = 0
			}
			// Clamp TreeScrollY so selected item stays in view
			if t.selectedIndex < t.State.TreeScrollY {
				t.State.TreeScrollY = t.selectedIndex
			}
			if t.selectedIndex >= t.State.TreeScrollY+visibleHeight {
				t.State.TreeScrollY = t.selectedIndex - visibleHeight + 1
			}
			if t.State.TreeScrollY < 0 {
				t.State.TreeScrollY = 0
			}
		}
		return t, nil
	case model.DirChangedMsg:
		_ = t.FileTree.Refresh()
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		if t.selectedIndex >= len(t.visibleNodes) && len(t.visibleNodes) > 0 {
			t.selectedIndex = len(t.visibleNodes) - 1
		}
		if t.selectedIndex < 0 {
			t.selectedIndex = 0
		}
		t.refreshGitStatus()
		return t, t.FileTree.WatchCmd()
	// Size poll tick: read terminal size so resize is reflected (terminals that don't send WindowSizeMsg, e.g. Windows CMD)
	case time.Time:
		return t, tea.Batch(getSizeCmd(), tea.Tick(sizePollInterval, func(t time.Time) tea.Msg { return t }))
	}
	
	return t, nil
}

// handleMouse handles mouse events
func (t *TUI) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Only handle left clicks
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return t, nil
	}
	
	// Calculate panel boundaries
	explorerWidth := t.State.ExplorerWidth
	editorStartX := explorerWidth + 2 // Account for borders
	
	// Double-click detection (typically 300-500ms)
	const doubleClickThreshold int64 = 500_000_000 // 500ms in nanoseconds
	currentTime := time.Now().UnixNano()
	
	// Check if click is in explorer panel
	if msg.X < editorStartX {
		if t.State.Dialog.Type == model.DialogNone {
			t.State.FocusPanel = model.PanelExplorer
			
			// Calculate which item was clicked
			// Layout: border(1) + title(1) + separator(1) = 3 lines before content
			clickY := msg.Y - 3
			if clickY >= 0 && clickY < len(t.visibleNodes) {
				clickedIndex := t.State.TreeScrollY + clickY
				if clickedIndex >= len(t.visibleNodes) {
					clickedIndex = len(t.visibleNodes) - 1
				}
				
				// Check for double-click
				isDoubleClick := t.lastClickPanel == 1 &&
					currentTime-t.lastClickTime < doubleClickThreshold &&
					t.lastClickX == msg.X && t.lastClickY == msg.Y &&
					t.selectedIndex == clickedIndex
				
				// Update click state
				t.lastClickTime = currentTime
				t.lastClickX = msg.X
				t.lastClickY = msg.Y
				t.lastClickPanel = 1
				
				if isDoubleClick {
					// Double-click: open file or toggle directory
					if clickedIndex >= 0 && clickedIndex < len(t.visibleNodes) {
						node := t.visibleNodes[clickedIndex]
						if node.Type == model.FileTypeDirectory {
							// Toggle directory expansion
							t.FileTree.ToggleNode(node)
							t.visibleNodes = t.FileTree.GetVisibleNodes()
						} else {
							// Open file in editor
							if t.Editor.IsModified() {
								t.State.Dialog.Type = model.DialogConfirmSwitch
								t.State.Dialog.TargetPath = node.Path
								t.dialogMessage = "Current file has unsaved changes. Save before switching?"
								return t, nil
							}
							t.Editor.OpenFile(node.Path)
							t.State.FocusPanel = model.PanelEditor
							t.State.Mode = model.ModeNormal
						}
					}
				} else {
					// Single click: just select
					t.selectedIndex = clickedIndex
				}
			}
		}
	} else {
		// Click is in editor panel
		if t.State.Dialog.Type == model.DialogNone {
			t.State.FocusPanel = model.PanelEditor
			
			// Handle click in editor
			if t.Editor.Buffer != nil {
				// Calculate line number from click
				// Layout: border(1) + title(1) + separator(1) = 3 lines before content
				clickY := msg.Y - 3
				lineNum := t.Editor.Buffer.ScrollY + clickY
				
				if lineNum >= 0 && lineNum < len(t.Editor.Buffer.Lines) {
					// Calculate column from click (account for line numbers)
					lineNumWidth := 4
					col := msg.X - editorStartX - lineNumWidth - 2
					if col < 0 {
						col = 0
					}
					
					line := t.Editor.Buffer.Lines[lineNum]
					if col > len(line) {
						col = len(line)
					}
					
					// Check for double-click
					isDoubleClick := t.lastClickPanel == 2 &&
						currentTime-t.lastClickTime < doubleClickThreshold &&
						t.lastClickX == msg.X && t.lastClickY == msg.Y &&
						t.Editor.Buffer.Cursor.Line == lineNum &&
						t.Editor.Buffer.Cursor.Column == col
					
					// Update click state
					t.lastClickTime = currentTime
					t.lastClickX = msg.X
					t.lastClickY = msg.Y
					t.lastClickPanel = 2
					
					// Move cursor to clicked position
					t.Editor.Buffer.Cursor.Line = lineNum
					t.Editor.Buffer.Cursor.Column = col
					
					if isDoubleClick {
						// Double-click: enter insert mode
						t.State.Mode = model.ModeInsert
					}
				}
			}
		}
	}
	
	return t, nil
}

// View renders the TUI
func (t *TUI) View() string {
	if t.State.Dialog.Type != model.DialogNone {
		return t.renderMainWithDialog()
	}
	
	return t.renderMain()
}

// centerText centers text within width
func (t *TUI) centerText(text string, width int) string {
	padding := (width - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	return strings.Repeat(" ", padding) + text
}

// minTermWidth/Height are the smallest dimensions we use (before first resize or tiny terminals).
// We use actual terminal size when larger so the UI is not cropped on small terminals.
const minTermWidth = 40
const minTermHeight = 6

// renderMain renders the main editor view. Size comes from State (updated by WindowSizeMsg or our size-poll tick).
func (t *TUI) renderMain() string {
	width := t.State.Width
	height := t.State.Height
	if width < minTermWidth {
		width = minTermWidth
	}
	if height < minTermHeight {
		height = minTermHeight
	}

	// Calculate panel widths
	explorerWidth := t.State.ExplorerWidth
	editorWidth := width - explorerWidth - 3
	if editorWidth < 10 {
		editorWidth = 10
	}
	
	// Render panels
	explorerPanel := t.renderExplorer(explorerWidth, height-4)
	editorPanel := t.renderEditor(editorWidth, height-4)
	statusBar := t.renderStatusBar(width)
	helpBar := t.renderHelpBar(width)
	
	// Layout
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		explorerPanel,
		editorPanel,
	)
	
	out := lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		statusBar,
		helpBar,
	)
	// Cap to terminal height so the top is never cropped when terminal is small
	lines := strings.Split(out, "\n")
	if len(lines) > height {
		lines = lines[:height]
		out = strings.Join(lines, "\n")
	}
	return out
}

// renderMainWithDialog renders main view with dialog overlay
func (t *TUI) renderMainWithDialog() string {
	mainView := t.renderMain()

	dialog := t.renderDialog()

	lines := strings.Split(mainView, "\n")
	dialogLines := strings.Split(dialog, "\n")

	startRow := (len(lines) - len(dialogLines)) / 2
	if startRow < 0 {
		startRow = 0
	}

	for i, dl := range dialogLines {
		row := startRow + i
		if row >= len(lines) {
			break
		}
		visualW := lipgloss.Width(dl)
		padLeft := (t.State.Width - visualW) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		lines[row] = strings.Repeat(" ", padLeft) + dl
	}

	return strings.Join(lines, "\n")
}

// renderExplorer renders the file explorer panel
func (t *TUI) renderExplorer(width, height int) string {
	if t.FileTree.Root == nil {
		content := "No directory open\n\nPress Ctrl+O to open"
		style := panelStyle
		if t.State.FocusPanel == model.PanelExplorer {
			style = activePanelStyle
		}
		return style.Width(width).Height(height).Render(content)
	}
	
	// Get visible nodes
	t.visibleNodes = t.FileTree.GetVisibleNodes()
	
	var lines []string
	
	// Title
	title := titleStyle.Render(" Explorer ")
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-2))
	
	// Content area height (title + separator take 2 lines; borders take space)
	contentHeight := height - 4
	if contentHeight < 0 {
		contentHeight = 0
	}
	startIdx := t.State.TreeScrollY
	endIdx := startIdx + contentHeight
	if endIdx > len(t.visibleNodes) {
		endIdx = len(t.visibleNodes)
	}
	if startIdx < 0 {
		startIdx = 0
	}
	
	// Render nodes
	for i := startIdx; i < endIdx; i++ {
		node := t.visibleNodes[i]
		line := t.renderTreeNode(node, i == t.selectedIndex, width-4)
		lines = append(lines, line)
	}
	
	// Fill remaining space
	minLines := height - 2
	if minLines < 2 {
		minLines = 2
	}
	for len(lines) < minLines {
		lines = append(lines, "")
	}
	
	content := strings.Join(lines, "\n")
	
	style := panelStyle
	if t.State.FocusPanel == model.PanelExplorer {
		style = activePanelStyle
	}
	
	return style.Width(width).Height(height).Render(content)
}

// renderTreeNode renders a single tree node
func (t *TUI) renderTreeNode(node *model.TreeNode, selected bool, width int) string {
	depth := explorer.GetNodeDepth(node)
	indent := strings.Repeat("  ", depth)

	var icon, name string

	switch node.Type {
	case model.FileTypeDirectory:
		if node.Expanded {
			icon = "▼"
		} else {
			icon = "▶"
		}
		name = node.Name + "/"
	default:
		icon = " "
		name = node.Name
	}

	gs := gitutil.StatusForPath(t.gitStatus, t.FileTree.RootPath, node.Path)
	label := gitutil.StatusLabel(gs)

	// Reserve 2 chars for " M" suffix when git status is present
	maxNameWidth := width
	if label != "" {
		maxNameWidth = width - 2
	}

	text := indent + icon + " " + name
	if len(text) > maxNameWidth && maxNameWidth > 3 {
		text = text[:maxNameWidth-3] + "..."
	}

	if gs == gitutil.FileIgnored {
		if selected {
			return selectedStyle.Render(text)
		}
		return gitIgnoredStyle.Render(text)
	}

	if selected {
		rendered := selectedStyle.Render(text)
		if label != "" {
			rendered += " " + t.styledGitLabel(gs, label)
		}
		return rendered
	}

	if node.Type == model.FileTypeDirectory {
		if gs != gitutil.FileClean {
			return dirStyle.Render(text) + " " + t.styledGitLabel(gs, label)
		}
		return dirStyle.Render(text)
	}

	if gs != gitutil.FileClean {
		styledName := t.gitFileStyle(gs).Render(text)
		return styledName + " " + t.styledGitLabel(gs, label)
	}

	return fileStyle.Render(text)
}

func (t *TUI) styledGitLabel(gs gitutil.FileGitStatus, label string) string {
	switch gs {
	case gitutil.FileModified:
		return gitStatusModifiedStyle.Render(label)
	case gitutil.FileAdded:
		return gitStatusAddedStyle.Render(label)
	case gitutil.FileUntracked:
		return gitStatusUntrackedStyle.Render(label)
	case gitutil.FileDeleted:
		return gitStatusDeletedStyle.Render(label)
	case gitutil.FileConflicted:
		return gitStatusConflictStyle.Render(label)
	case gitutil.FileRenamed:
		return gitStatusRenamedStyle.Render(label)
	}
	return label
}

func (t *TUI) gitFileStyle(gs gitutil.FileGitStatus) lipgloss.Style {
	switch gs {
	case gitutil.FileModified:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#E3B341"))
	case gitutil.FileAdded:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3FB950"))
	case gitutil.FileUntracked:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F0883E"))
	case gitutil.FileDeleted:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149"))
	case gitutil.FileConflicted:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F472B6"))
	case gitutil.FileRenamed:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#39C5CF"))
	}
	return fileStyle
}

// renderEditor renders the editor panel
func (t *TUI) renderEditor(width, height int) string {
	style := panelStyle
	if t.State.FocusPanel == model.PanelEditor {
		style = activePanelStyle
	}
	
	if t.Editor.Buffer == nil {
		content := "\n\n  No file open\n\n  Select a file from the explorer\n  or press Ctrl+N to create a new file"
		return style.Width(width).Height(height).Render(content)
	}
	
	var lines []string
	
	// Title with file name
	title := t.Editor.GetFileName()
	if t.Editor.IsModified() {
		title = title + " [Modified]"
	}
	lines = append(lines, titleStyle.Render(" "+title+" "))
	lines = append(lines, strings.Repeat("─", width-2))
	
	editorContentHeight := height - 4
	if editorContentHeight < 1 {
		editorContentHeight = 1
	}
	t.Editor.ScrollToCursor(editorContentHeight)
	visibleLines := t.Editor.GetVisibleLines(editorContentHeight)
	
	lineNumWidth := 4
	hasGutter := len(t.Editor.Buffer.GutterStatus) > 0
	gutterWidth := 0
	if hasGutter {
		gutterWidth = 1
	}

	for i, line := range visibleLines {
		lineNum := t.Editor.Buffer.ScrollY + i + 1
		lineNumStr := fmt.Sprintf("%*d ", lineNumWidth-1, lineNum)
		absLine := t.Editor.Buffer.ScrollY + i

		cursorLine := t.Editor.Buffer.Cursor.Line
		isCurrentLine := absLine == cursorLine

		renderedLine := lineNumStyle.Render(lineNumStr)

		gutterChar := ""
		if hasGutter {
			gutterChar = " "
			if absLine < len(t.Editor.Buffer.GutterStatus) {
				switch gitutil.LineStatus(t.Editor.Buffer.GutterStatus[absLine]) {
				case gitutil.StatusAdded:
					gutterChar = gutterAddedStyle.Render("│")
				case gitutil.StatusModified:
					gutterChar = gutterModifiedStyle.Render("│")
				case gitutil.StatusDeleted:
					gutterChar = gutterDeletedStyle.Render("─")
				}
			}
		}

		showCursor := isCurrentLine
		cursorPos := t.Editor.Buffer.Cursor.Column
		content := t.renderLineContent(line, width-lineNumWidth-gutterWidth-2, showCursor, cursorPos)

		if isCurrentLine && t.State.FocusPanel == model.PanelEditor {
			if t.State.Mode == model.ModeInsert {
				content = lipgloss.NewStyle().Background(lipgloss.Color("#2D3748")).Render(content)
			} else {
				content = lipgloss.NewStyle().Background(lipgloss.Color("#1E3A5F")).Render(content)
			}
		}

		lines = append(lines, renderedLine+gutterChar+content)
	}
	
	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}
	
	content := strings.Join(lines, "\n")
	return style.Width(width).Height(height).Render(content)
}

// renderLineContent renders a line of content with cursor
func (t *TUI) renderLineContent(line string, width int, showCursor bool, cursorPos int) string {
	// Truncate if too long
	if len(line) > width {
		// Handle horizontal scroll
		scrollX := 0
		if t.Editor.Buffer != nil {
			scrollX = t.Editor.Buffer.ScrollX
		}
		if scrollX > len(line) {
			scrollX = 0
		}
		endX := scrollX + width
		if endX > len(line) {
			endX = len(line)
		}
		line = line[scrollX:endX]
		cursorPos = cursorPos - scrollX
	}
	
	// Pad to width
	if len(line) < width {
		line = line + strings.Repeat(" ", width-len(line))
	}
	
	// Show cursor if this is the current line and editor is focused
	if showCursor && t.State.FocusPanel == model.PanelEditor && t.State.Mode == model.ModeInsert {
		if cursorPos >= 0 && cursorPos < len(line) {
			// Highlight the character at cursor position
			before := line[:cursorPos]
			atCursor := string(line[cursorPos])
			after := line[cursorPos+1:]
			return before + cursorStyle.Render(atCursor) + after
		}
	}
	
	return line
}

// renderStatusBar renders the status bar
func (t *TUI) renderStatusBar(width int) string {
	left := ""
	right := ""
	
	// Left side: mode and file
	if t.State.FocusPanel == model.PanelEditor {
		modeStr := "NORMAL"
		if t.State.Mode == model.ModeInsert {
			modeStr = "INSERT"
		} else if t.State.Mode == model.ModeCommand {
			modeStr = "COMMAND"
		}
		left = fmt.Sprintf(" [%s] %s", modeStr, t.Editor.GetFileName())
		
		// Right side: cursor position
		line, col := t.Editor.GetCursorPosition()
		totalLines := t.Editor.GetLineCount()
		right = fmt.Sprintf("Ln %d/%d, Col %d ", line, totalLines, col)
	} else if t.State.FocusPanel == model.PanelExplorer {
		left = " [EXPLORER] "
		if t.FileTree.Root != nil {
			left += t.FileTree.RootPath
		}
	}
	
	// Pad between left and right
	middle := width - len(left) - len(right)
	if middle < 0 {
		middle = 0
	}
	
	return statusStyle.Width(width).Render(left + strings.Repeat(" ", middle) + right)
}

// renderHelpBar renders the help bar
func (t *TUI) renderHelpBar(width int) string {
	help := ""
	
	if t.State.FocusPanel == model.PanelExplorer {
		help = "Enter: Open/Expand | n/N: New | F2: Rename | Del | x/y/p: Cut/Copy/Paste | Ctrl+←→: Resize | Ctrl+H: Guide | Tab: Editor | Esc/Ctrl+Q: Quit"
	} else if t.State.FocusPanel == model.PanelEditor {
		if t.State.Mode == model.ModeInsert {
			help = "Esc: Normal | Ctrl+S: Save | Ctrl+H: Guide | Ctrl+Q: Quit"
		} else {
			help = "i a o: Insert/Append/Line | d: Delete Line | Ctrl+S: Save | Ctrl+←→: Resize | Ctrl+H: Guide | Tab: Explorer | Ctrl+Q: Quit"
		}
	}
	
	const minHelpWidth = 40
	if width < minHelpWidth {
		width = minHelpWidth
	}
	if len(help) > width {
		help = help[:width-3] + "..."
	}
	return helpStyle.Width(width).Render(help)
}

// renderHelpGuide returns the command guide for the Ctrl+H popup (dynamic by focus and mode, like bottom bar)
func (t *TUI) renderHelpGuide() string {
	sectionTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	currentTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).Background(lipgloss.Color("#2D3748"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))

	explorerLines := []string{
		"  " + keyStyle.Render("Enter") + "  Open file / Expand folder",
		"  " + keyStyle.Render("↑↓ j k") + "  Move selection",
		"  " + keyStyle.Render("← h  → l") + "  Collapse / Expand",
		"  " + keyStyle.Render("n  N") + "  New file / folder",
		"  " + keyStyle.Render("F2") + "  Rename  " + keyStyle.Render("Del d") + "  Delete",
		"  " + keyStyle.Render("x") + " / " + keyStyle.Render("y") + " / " + keyStyle.Render("p") + "  Cut / Copy / Paste",
		"  " + keyStyle.Render("Ctrl+O") + "  Open  " + keyStyle.Render("Tab") + "  Focus editor  " + keyStyle.Render("r") + "  Refresh",
		"  " + keyStyle.Render("Ctrl+←") + " / " + keyStyle.Render("Ctrl+→") + "  Resize explorer panel",
		"  " + keyStyle.Render("Esc") + "  or  " + keyStyle.Render("Ctrl+Q") + "  Quit",
	}
	editorNormalLines := []string{
		"  " + keyStyle.Render("i a o O") + "  Insert / Append / New line",
		"  " + keyStyle.Render("d") + "  Delete line  " + keyStyle.Render("x X") + "  Delete char",
		"  " + keyStyle.Render("↑↓ j k  ← →") + "  Move cursor",
		"  " + keyStyle.Render("0 $  w b  g G") + "  Line / word / first–last line",
		"  " + keyStyle.Render("Ctrl+S") + "  Save  " + keyStyle.Render("Ctrl+O N") + "  Open / New file",
		"  " + keyStyle.Render("Ctrl+← →") + "  Resize explorer  " + keyStyle.Render("Tab") + "  Focus explorer",
	}
	editorInsertLines := []string{
		"  " + keyStyle.Render("Esc") + "  Normal mode",
		"  " + keyStyle.Render("Ctrl+S") + "  Save  " + keyStyle.Render("Ctrl+Q") + "  Quit",
	}
	generalLines := []string{
		"  " + keyStyle.Render("Ctrl+H") + "  This guide  " + keyStyle.Render("Esc") + " / " + keyStyle.Render("Ctrl+Q") + "  Quit",
	}

	var out []string
	add := func(title string, isCurrent bool, section []string) {
		if isCurrent {
			out = append(out, currentTitle.Render(" ▶ "+title+" (current) "))
		} else {
			out = append(out, sectionTitle.Render(title))
		}
		out = append(out, section...)
		out = append(out, "")
	}

	// Order: current context first (like bottom suggestions), then others
	switch t.State.FocusPanel {
	case model.PanelExplorer:
		add("Explorer", true, explorerLines)
		add("Editor (normal)", false, editorNormalLines)
		add("Editor (insert)", false, editorInsertLines)
	case model.PanelEditor:
		if t.State.Mode == model.ModeInsert {
			add("Editor (insert)", true, editorInsertLines)
			add("Explorer", false, explorerLines)
			add("Editor (normal)", false, editorNormalLines)
		} else {
			add("Editor (normal)", true, editorNormalLines)
			add("Explorer", false, explorerLines)
			add("Editor (insert)", false, editorInsertLines)
		}
	default:
		add("Explorer", false, explorerLines)
		add("Editor (normal)", false, editorNormalLines)
		add("Editor (insert)", false, editorInsertLines)
	}
	out = append(out, sectionTitle.Render("General"))
	out = append(out, generalLines...)
	return strings.Join(out, "\n")
}

// renderDialog renders a dialog box
func (t *TUI) renderDialog() string {
	var content string
	
	switch t.State.Dialog.Type {
	case model.DialogOpenFile, model.DialogOpenDir:
		title := "Open File"
		if t.State.Dialog.Type == model.DialogOpenDir {
			title = "Open Directory"
		}
		content = fmt.Sprintf("%s\n\nPath: %s_", title, t.dialogInput)
		
		// Add preview list with scroll offset
		if len(t.State.Dialog.Preview) > 0 {
			content += "\n\nMatching items:"
			maxVisible := 10
			
			// Calculate visible range with scroll
			startIdx := t.dialogPreviewScroll
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx := startIdx + maxVisible
			if endIdx > len(t.State.Dialog.Preview) {
				endIdx = len(t.State.Dialog.Preview)
			}
			
			// Show scroll indicator if needed
			if t.dialogPreviewScroll > 0 {
				content += fmt.Sprintf(" (↑ %d more above)", t.dialogPreviewScroll)
			}
			
			for i := startIdx; i < endIdx; i++ {
				path := t.State.Dialog.Preview[i]
				name := path
				isDir := t.FileOps.IsDirectory(path)
				
				prefix := "  "
				if i == t.State.Dialog.PreviewIdx {
					prefix = "> "
				}
				
				if isDir && !strings.HasSuffix(name, "/") {
					name += "/"
				}
				
				if i == t.State.Dialog.PreviewIdx {
					content += "\n" + prefix + previewSelectedStyle.Render(name)
				} else if isDir {
					content += "\n" + prefix + previewDirStyle.Render(name)
				} else {
					content += "\n" + prefix + previewItemStyle.Render(name)
				}
			}
			
			if endIdx < len(t.State.Dialog.Preview) {
				content += fmt.Sprintf("\n  ... and %d more below", len(t.State.Dialog.Preview)-endIdx)
			}
		} else {
			content += "\n\nNo matching items"
		}
		
		if t.dialogError != "" {
			content += fmt.Sprintf("\n\n%s", errorStyle.Render(t.dialogError))
		}
		
	case model.DialogNewFile:
		content = fmt.Sprintf("Create New File\n\nPath: %s_", t.dialogInput)
		
		// Add preview list for directory selection
		if len(t.State.Dialog.Preview) > 0 {
			content += "\n\nDirectories:"
			maxItems := 5
			if len(t.State.Dialog.Preview) < maxItems {
				maxItems = len(t.State.Dialog.Preview)
			}
			for i := 0; i < maxItems; i++ {
				path := t.State.Dialog.Preview[i]
				name := filepath.Base(path) + "/"
				
				prefix := "  "
				if i == t.State.Dialog.PreviewIdx {
					prefix = "> "
				}
				
				if i == t.State.Dialog.PreviewIdx {
					content += "\n" + prefix + previewSelectedStyle.Render(name)
				} else {
					content += "\n" + prefix + previewDirStyle.Render(name)
				}
			}
		}
		
		if t.dialogError != "" {
			content += fmt.Sprintf("\n\n%s", errorStyle.Render(t.dialogError))
		}
		
	case model.DialogNewFolder:
		content = fmt.Sprintf("Create New Folder\n\nPath: %s_", t.dialogInput)
		
		// Add preview list for parent directory selection
		if len(t.State.Dialog.Preview) > 0 {
			content += "\n\nParent directories:"
			maxItems := 5
			if len(t.State.Dialog.Preview) < maxItems {
				maxItems = len(t.State.Dialog.Preview)
			}
			for i := 0; i < maxItems; i++ {
				path := t.State.Dialog.Preview[i]
				name := filepath.Base(path) + "/"
				
				prefix := "  "
				if i == t.State.Dialog.PreviewIdx {
					prefix = "> "
				}
				
				if i == t.State.Dialog.PreviewIdx {
					content += "\n" + prefix + previewSelectedStyle.Render(name)
				} else {
					content += "\n" + prefix + previewDirStyle.Render(name)
				}
			}
		}
		
		if t.dialogError != "" {
			content += fmt.Sprintf("\n\n%s", errorStyle.Render(t.dialogError))
		}
		
	case model.DialogRename:
		content = fmt.Sprintf("Rename\n\nNew name: %s_", t.dialogInput)
		if t.dialogError != "" {
			content += fmt.Sprintf("\n\n%s", errorStyle.Render(t.dialogError))
		}
		
	case model.DialogDelete:
		content = fmt.Sprintf("Delete\n\nAre you sure you want to delete?\n%s\n\nPress Y to confirm, N to cancel", t.dialogMessage)
		
	case model.DialogSave:
		content = fmt.Sprintf("Save Changes\n\nFile has unsaved changes. Save before closing?\n\nPress Y: Save | N: Don't Save | Esc: Cancel")
		
	case model.DialogConfirmSwitch:
		content = fmt.Sprintf("Save Changes\n\n%s\n\nPress Y: Save | N: Don't Save | Esc: Cancel", t.dialogMessage)
		
	case model.DialogQuit:
		if t.Editor.IsModified() {
			content = "Save Changes\n\nFile has unsaved changes. Save before quitting?\n\nPress Y: Save | N: Don't Save | Esc: Cancel"
		} else {
			content = "Quit\n\nPress Y to confirm, N to cancel"
		}

	case model.DialogHelp:
		content = t.renderHelpGuide()

	default:
		content = "Dialog"
	}
	
	// Update help text based on dialog type
	if t.State.Dialog.Type == model.DialogHelp {
		content += "\n\n" + helpStyle.Render("Press any key or Esc to close")
	} else if t.State.Dialog.Type == model.DialogOpenDir {
		content += "\n\n[Enter: Select Current Path] [Right/L: Navigate Into] [Tab: Complete] [↑↓: Navigate] [Esc: Cancel]"
	} else if t.State.Dialog.Type == model.DialogOpenFile {
		content += "\n\n[Enter: Open/Select] [Right/L: Navigate Into Dir] [Tab: Complete] [↑↓: Navigate] [Esc: Cancel]"
	} else {
		content += "\n\n[Enter: Confirm] [Tab: Complete] [↑↓: Navigate] [Esc: Cancel]"
	}
	
	maxW := t.State.Width - 4
	if maxW > 70 {
		maxW = 70
	}
	if maxW < 40 {
		maxW = 40
	}
	return dialogBoxStyle.MaxWidth(maxW).Render(content)
}
