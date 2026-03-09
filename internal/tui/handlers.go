package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"tuidit/internal/config"
	"tuidit/internal/editor"
	"tuidit/internal/model"
	"tuidit/internal/utils"
)

// handleKeyPress handles keyboard input
func (t *TUI) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle dialog input first
	if t.State.Dialog.Type != model.DialogNone {
		return t.handleDialogInput(msg)
	}

	// Handle input based on current mode and focus
	switch t.State.FocusPanel {
	case model.PanelExplorer:
		return t.handleExplorerInput(msg)
	case model.PanelEditor:
		return t.handleEditorInput(msg)
	}
	
	return t, nil
}

// handleDialogInput handles dialog input
func (t *TUI) handleDialogInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay: any key closes it
	if t.State.Dialog.Type == model.DialogHelp {
		t.State.Dialog.Type = model.DialogNone
		return t, nil
	}
	switch msg.String() {
	case "esc":
		t.State.Dialog.Type = model.DialogNone
		t.dialogInput = ""
		t.dialogError = ""
		t.State.Dialog.Preview = nil
		t.State.Dialog.PreviewIdx = 0
		t.dialogPreviewScroll = 0
		
	case "enter":
		// Enter: Select the current directory/file for editing (confirm the input field)
		return t.confirmDialog()
		
	case "right", "l", "ctrl+enter":
		// Right/L/Ctrl+Enter: Navigate into directory (if selected item is a directory)
		if len(t.State.Dialog.Preview) > 0 && t.State.Dialog.PreviewIdx >= 0 {
			selected := t.State.Dialog.Preview[t.State.Dialog.PreviewIdx]
			
			// If it's a directory, navigate into it
			if t.FileOps.IsDirectory(selected) {
				t.dialogInput = selected + string(filepath.Separator)
				t.State.Dialog.PreviewIdx = 0
				t.dialogPreviewScroll = 0
				t.updateDialogPreview()
				return t, nil
			}
		}
		// If not a directory or no selection, confirm dialog
		return t.confirmDialog()
		
	case "tab":
		// Tab completion
		t.tabComplete()
		
	case "up", "ctrl+p":
		// Navigate preview up
		if len(t.State.Dialog.Preview) > 0 {
			if t.State.Dialog.PreviewIdx > 0 {
				t.State.Dialog.PreviewIdx--
			} else {
				t.State.Dialog.PreviewIdx = len(t.State.Dialog.Preview) - 1
			}
			t.updatePreviewScroll()
		}
		
	case "down", "ctrl+n":
		// Navigate preview down
		if len(t.State.Dialog.Preview) > 0 {
			if t.State.Dialog.PreviewIdx < len(t.State.Dialog.Preview)-1 {
				t.State.Dialog.PreviewIdx++
			} else {
				t.State.Dialog.PreviewIdx = 0
			}
			t.updatePreviewScroll()
		}
		
	case "backspace":
		if len(t.dialogInput) > 0 {
			t.dialogInput = t.dialogInput[:len(t.dialogInput)-1]
			t.State.Dialog.PreviewIdx = 0
			t.dialogPreviewScroll = 0
			t.updateDialogPreview()
		}
		t.dialogError = ""
		
	case "y", "Y":
		if t.State.Dialog.Type == model.DialogDelete ||
			t.State.Dialog.Type == model.DialogSave ||
			t.State.Dialog.Type == model.DialogConfirmSwitch ||
			t.State.Dialog.Type == model.DialogQuit {
			return t.confirmDialogYes()
		} else {
			t.dialogInput += "y"
			t.State.Dialog.PreviewIdx = 0
			t.dialogPreviewScroll = 0
			t.updateDialogPreview()
		}
		
	case "n", "N":
		if t.State.Dialog.Type == model.DialogDelete ||
			t.State.Dialog.Type == model.DialogSave ||
			t.State.Dialog.Type == model.DialogConfirmSwitch ||
			t.State.Dialog.Type == model.DialogQuit {
			return t.confirmDialogNo()
		} else {
			t.dialogInput += "n"
			t.State.Dialog.PreviewIdx = 0
			t.dialogPreviewScroll = 0
			t.updateDialogPreview()
		}
		
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			t.dialogInput += msg.String()
			t.State.Dialog.PreviewIdx = 0
			t.dialogPreviewScroll = 0
			t.updateDialogPreview()
			t.dialogError = ""
		}
	}
	
	return t, nil
}

// updatePreviewScroll updates the scroll position to keep selected item visible
func (t *TUI) updatePreviewScroll() {
	maxVisible := 10 // Maximum items visible in preview
	
	if t.State.Dialog.PreviewIdx < t.dialogPreviewScroll {
		t.dialogPreviewScroll = t.State.Dialog.PreviewIdx
	} else if t.State.Dialog.PreviewIdx >= t.dialogPreviewScroll+maxVisible {
		t.dialogPreviewScroll = t.State.Dialog.PreviewIdx - maxVisible + 1
	}
	
	// Ensure scroll doesn't go negative
	if t.dialogPreviewScroll < 0 {
		t.dialogPreviewScroll = 0
	}
}

// updateDialogPreview updates the preview list based on current input
func (t *TUI) updateDialogPreview() {
	if t.State.Dialog.Type != model.DialogOpenFile &&
		t.State.Dialog.Type != model.DialogOpenDir &&
		t.State.Dialog.Type != model.DialogNewFile &&
		t.State.Dialog.Type != model.DialogNewFolder {
		return
	}
	
	expanded := utils.ExpandPath(t.dialogInput)
	
	// Determine the directory to list and the filter prefix
	var dirPath, filterPrefix string
	
	// Check if the path ends with a separator
	if strings.HasSuffix(t.dialogInput, string(filepath.Separator)) {
		dirPath = expanded
		filterPrefix = ""
	} else {
		// Check if the path exists and is a directory
		info, err := os.Stat(expanded)
		if err == nil && info.IsDir() {
			dirPath = expanded
			filterPrefix = ""
		} else {
			// Treat the last component as a filter prefix
			dirPath = filepath.Dir(expanded)
			filterPrefix = filepath.Base(expanded)
		}
	}
	
	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.State.Dialog.Preview = nil
		return
	}
	
	// Filter and sort entries
	var results []string
	for _, entry := range entries {
		name := entry.Name()
		
		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}
		
		// Apply filter
		if filterPrefix != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(filterPrefix)) {
			continue
		}
		
		fullPath := filepath.Join(dirPath, name)
		
		// For open directory dialog, only show directories
		if t.State.Dialog.Type == model.DialogOpenDir && !entry.IsDir() {
			continue
		}
		
		results = append(results, fullPath)
	}
	
	// Sort: directories first, then files, alphabetically
	sort.Slice(results, func(i, j int) bool {
		infoI, errI := os.Stat(results[i])
		infoJ, errJ := os.Stat(results[j])
		if errI != nil || errJ != nil {
			return results[i] < results[j]
		}
		if infoI.IsDir() != infoJ.IsDir() {
			return infoI.IsDir()
		}
		return strings.ToLower(results[i]) < strings.ToLower(results[j])
	})
	
	t.State.Dialog.Preview = results
	t.State.Dialog.PreviewIdx = 0
}

// tabComplete performs tab completion on the current input
func (t *TUI) tabComplete() {
	if len(t.State.Dialog.Preview) == 0 {
		return
	}
	
	// Use the selected preview item
	if t.State.Dialog.PreviewIdx >= 0 && t.State.Dialog.PreviewIdx < len(t.State.Dialog.Preview) {
		selected := t.State.Dialog.Preview[t.State.Dialog.PreviewIdx]
		t.dialogInput = selected
		
		// If it's a directory, add separator and update preview
		if t.FileOps.IsDirectory(selected) {
			t.dialogInput += string(filepath.Separator)
		}
		
		t.updateDialogPreview()
	}
}

// confirmDialog confirms the current dialog
func (t *TUI) confirmDialog() (tea.Model, tea.Cmd) {
	switch t.State.Dialog.Type {
	case model.DialogOpenDir:
		return t.openDirectory(t.dialogInput)
		
	case model.DialogOpenFile:
		return t.openFile(t.dialogInput)
		
	case model.DialogNewFile:
		return t.createNewFile(t.dialogInput)
		
	case model.DialogNewFolder:
		return t.createNewFolder(t.dialogInput)
		
	case model.DialogRename:
		return t.renameItem(t.dialogInput)

	case model.DialogDelete, model.DialogSave, model.DialogConfirmSwitch, model.DialogQuit:
		return t.confirmDialogYes()
	}
	
	t.State.Dialog.Type = model.DialogNone
	return t, nil
}

// confirmDialogYes confirms dialog with Yes
func (t *TUI) confirmDialogYes() (tea.Model, tea.Cmd) {
	switch t.State.Dialog.Type {
	case model.DialogDelete:
		return t.deleteItem()
		
	case model.DialogSave:
		if err := t.Editor.SaveFile(); err != nil {
			t.dialogError = err.Error()
			return t, nil
		}
		t.State.Dialog.Type = model.DialogNone
		// If there's a pending file to open, open it
		if t.State.Dialog.TargetPath != "" {
			return t.openFile(t.State.Dialog.TargetPath)
		}
		
	case model.DialogConfirmSwitch:
		// Save current file and open the pending file
		if err := t.Editor.SaveFile(); err != nil {
			t.dialogError = err.Error()
			return t, nil
		}
		t.State.Dialog.Type = model.DialogNone
		if t.State.Dialog.TargetPath != "" {
			return t.openFile(t.State.Dialog.TargetPath)
		}
		
	case model.DialogQuit:
		if t.Editor.IsModified() {
			if err := t.Editor.SaveFile(); err != nil {
				t.dialogError = err.Error()
				return t, nil
			}
		}
		return t, tea.Quit
	}
	
	t.State.Dialog.Type = model.DialogNone
	return t, nil
}

// confirmDialogNo confirms dialog with No
func (t *TUI) confirmDialogNo() (tea.Model, tea.Cmd) {
	switch t.State.Dialog.Type {
	case model.DialogSave:
		t.State.Dialog.Type = model.DialogNone
		// If there's a pending file to open, open it without saving
		if t.State.Dialog.TargetPath != "" {
			return t.openFile(t.State.Dialog.TargetPath)
		}
		
	case model.DialogConfirmSwitch:
		// Don't save, just open the pending file
		t.State.Dialog.Type = model.DialogNone
		if t.State.Dialog.TargetPath != "" {
			return t.openFile(t.State.Dialog.TargetPath)
		}
		
	case model.DialogQuit:
		return t, tea.Quit
	}
	
	t.State.Dialog.Type = model.DialogNone
	return t, nil
}

// handleExplorerInput handles file explorer input
func (t *TUI) handleExplorerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if t.FileTree.Root != nil {
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		if t.selectedIndex >= len(t.visibleNodes) && len(t.visibleNodes) > 0 {
			t.selectedIndex = len(t.visibleNodes) - 1
		}
	}
	switch msg.String() {
	case "up", "k":
		t.moveSelectionUp()
		
	case "down", "j":
		t.moveSelectionDown()
		
	case "left", "h":
		t.collapseSelected()
		
	case "right", "l":
		t.expandSelected()
		
	case "enter":
		return t.openSelected()
		
	case "backspace":
		t.collapseSelected()
		
	case "tab":
		t.State.FocusPanel = model.PanelEditor
		t.State.Mode = model.ModeNormal
		
	case "n":
		// New file
		t.State.Dialog.Type = model.DialogNewFile
		t.dialogInput = ""
		if t.FileTree.Root != nil && t.selectedIndex < len(t.visibleNodes) {
			node := t.visibleNodes[t.selectedIndex]
			if node.Type == model.FileTypeDirectory {
				t.dialogInput = node.Path + string(filepath.Separator)
			} else if node.Parent != nil {
				t.dialogInput = node.Parent.Path + string(filepath.Separator)
			}
		}
		t.updateDialogPreview()
		
	case "N":
		// New folder
		t.State.Dialog.Type = model.DialogNewFolder
		t.dialogInput = ""
		if t.FileTree.Root != nil && t.selectedIndex < len(t.visibleNodes) {
			node := t.visibleNodes[t.selectedIndex]
			if node.Type == model.FileTypeDirectory {
				t.dialogInput = node.Path + string(filepath.Separator)
			} else if node.Parent != nil {
				t.dialogInput = node.Parent.Path + string(filepath.Separator)
			}
		}
		t.updateDialogPreview()
		
	case "f2":
		// Rename
		if t.FileTree.Root != nil && t.selectedIndex < len(t.visibleNodes) {
			node := t.visibleNodes[t.selectedIndex]
			if node != t.FileTree.Root {
				t.State.Dialog.Type = model.DialogRename
				t.dialogInput = node.Name
				t.State.Dialog.TargetPath = node.Path
			}
		}
		
	case "delete", "d":
		// Delete
		if t.FileTree.Root != nil && t.selectedIndex < len(t.visibleNodes) {
			node := t.visibleNodes[t.selectedIndex]
			if node != t.FileTree.Root {
				t.State.Dialog.Type = model.DialogDelete
				t.dialogMessage = node.Path
				t.State.Dialog.TargetPath = node.Path
			}
		}
		
	case "ctrl+o":
		// Open file/directory dialog
		t.State.Dialog.Type = model.DialogOpenDir
		t.dialogInput = "/"
		t.updateDialogPreview()

	case "ctrl+right":
		// Widen explorer
		t.resizeExplorerWidth(3)
	case "ctrl+left":
		// Narrow explorer
		t.resizeExplorerWidth(-3)

	case "ctrl+x", "x":
		if t.FileTree.Root != nil && t.selectedIndex < len(t.visibleNodes) {
			node := t.visibleNodes[t.selectedIndex]
			if node != t.FileTree.Root {
				t.clipboardPath = node.Path
				t.clipboardCut = true
				t.State.StatusMessage = "Cut: " + node.Name
			}
		}

	case "ctrl+c", "y":
		if t.FileTree.Root != nil && t.selectedIndex < len(t.visibleNodes) {
			node := t.visibleNodes[t.selectedIndex]
			if node != t.FileTree.Root {
				t.clipboardPath = node.Path
				t.clipboardCut = false
				t.State.StatusMessage = "Copied: " + node.Name
			}
		}

	case "ctrl+v", "p":
		if t.clipboardPath != "" {
			return t.pasteFromClipboard()
		}

	case "ctrl+h":
		// Show full command guide (popup)
		t.State.Dialog.Type = model.DialogHelp

	case "ctrl+q", "ctrl+shift+q", "esc":
		return t, tea.Quit

	case "r":
		// Refresh
		if t.FileTree.Root != nil {
			t.FileTree.Refresh()
			t.visibleNodes = t.FileTree.GetVisibleNodes()
		}
	}
	
	return t, nil
}

// handleEditorInput handles editor input
func (t *TUI) handleEditorInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In insert mode, handle text input
	if t.State.Mode == model.ModeInsert {
		switch msg.String() {
		case "esc":
			t.State.Mode = model.ModeNormal
			
		case "ctrl+s":
			if t.Editor.Buffer != nil {
				if err := t.Editor.SaveFile(); err != nil {
					t.State.StatusMessage = "Error saving: " + err.Error()
				} else {
					t.State.StatusMessage = "File saved"
				}
			}
			
		case "ctrl+q":
			if t.Editor.IsModified() {
				t.State.Dialog.Type = model.DialogQuit
			} else {
				return t, tea.Quit
			}
			
		case "enter":
			t.Editor.InsertNewline()
			
		case "backspace":
			t.Editor.DeleteChar()
			
		case "delete":
			t.Editor.DeleteCharForward()
			
		case "up":
			t.Editor.MoveCursor("up")
			
		case "down":
			t.Editor.MoveCursor("down")
			
		case "left":
			t.Editor.MoveCursor("left")
			
		case "right":
			t.Editor.MoveCursor("right")
			
		case "home":
			t.Editor.MoveCursor("home")
			
		case "end":
			t.Editor.MoveCursor("end")
			
		case "tab":
			t.Editor.InsertChar('\t')

		case "ctrl+h":
			t.State.Dialog.Type = model.DialogHelp
			return t, nil
			
		default:
			// Insert character
			if len(msg.String()) == 1 {
				t.Editor.InsertChar([]rune(msg.String())[0])
			}
		}
		return t, nil
	}
	
	// Normal mode
	switch msg.String() {
	case "i":
		t.State.Mode = model.ModeInsert
		
	case "a":
		// Append after cursor
		t.State.Mode = model.ModeInsert
		t.Editor.MoveCursor("right")
		
	case "o":
		// Open new line below
		t.State.Mode = model.ModeInsert
		t.Editor.InsertNewline()
		t.Editor.MoveCursor("left")
		
	case "O":
		// Open new line above
		t.State.Mode = model.ModeInsert
		t.Editor.MoveCursor("home")
		t.Editor.InsertNewline()
		t.Editor.MoveCursor("up")
		
	case "d":
		// Delete line
		if t.Editor.Buffer != nil {
			line := t.Editor.Buffer.Cursor.Line
			if line < len(t.Editor.Buffer.Lines) {
				t.Editor.Buffer.Lines = append(
					t.Editor.Buffer.Lines[:line],
					t.Editor.Buffer.Lines[line+1:]...,
				)
				if t.Editor.Buffer.Cursor.Line >= len(t.Editor.Buffer.Lines) {
					t.Editor.Buffer.Cursor.Line = len(t.Editor.Buffer.Lines) - 1
				}
				if t.Editor.Buffer.Cursor.Line < 0 {
					t.Editor.Buffer.Cursor.Line = 0
					t.Editor.Buffer.Lines = []string{""}
				}
				t.Editor.Buffer.Modified = true
			}
		}
		
	case "dd":
		// Delete line (vim style)
		// Already handled by single 'd'
		
	case "x":
		// Delete character under cursor
		t.Editor.DeleteCharForward()
		
	case "X":
		// Delete character before cursor
		t.Editor.DeleteChar()
		
	case "0", "^":
		t.Editor.MoveCursor("home")
		
	case "$":
		t.Editor.MoveCursor("end")
		
	case "w":
		// Move to next word
		t.moveWordForward()
		
	case "b":
		// Move to previous word
		t.moveWordBackward()
		
	case "G":
		// Go to last line
		if t.Editor.Buffer != nil {
			t.Editor.Buffer.Cursor.Line = len(t.Editor.Buffer.Lines) - 1
			t.Editor.ClampColumn()
		}
		
	case "g":
		// Go to first line
		t.Editor.Buffer.Cursor.Line = 0
		t.Editor.Buffer.Cursor.Column = 0
		
	case "up", "k":
		t.Editor.MoveCursor("up")
		
	case "down", "j":
		t.Editor.MoveCursor("down")
		
	case "left", "h":
		t.Editor.MoveCursor("left")
		
	case "right", "l":
		t.Editor.MoveCursor("right")
		
	case "tab":
		t.State.FocusPanel = model.PanelExplorer

	case "ctrl+right":
		t.resizeExplorerWidth(3)
	case "ctrl+left":
		t.resizeExplorerWidth(-3)

	case "ctrl+h":
		// Show full command guide (popup)
		t.State.Dialog.Type = model.DialogHelp
		
	case "ctrl+s":
		if t.Editor.Buffer != nil {
			if err := t.Editor.SaveFile(); err != nil {
				t.State.StatusMessage = "Error saving: " + err.Error()
			} else {
				t.State.StatusMessage = "File saved"
			}
		}
		
	case "ctrl+o":
		t.State.Dialog.Type = model.DialogOpenFile
		t.dialogInput = "/"
		t.updateDialogPreview()
		
	case "ctrl+n":
		t.State.Dialog.Type = model.DialogNewFile
		t.dialogInput = ""
		t.updateDialogPreview()
		
	case "ctrl+q":
		if t.Editor.IsModified() {
			t.State.Dialog.Type = model.DialogQuit
		} else {
			return t, tea.Quit
		}
		
	case "ctrl+c":
		return t, tea.Quit
	}
	
	return t, nil
}

// resizeExplorerWidth changes the explorer panel width by delta (positive = wider, negative = narrower).
func (t *TUI) resizeExplorerWidth(delta int) {
	const minExplorerWidth = 15
	maxExplorerWidth := 60
	if t.State.Width > 0 {
		if m := t.State.Width - 3 - 20; m < maxExplorerWidth {
			maxExplorerWidth = m
		}
	}
	if maxExplorerWidth < minExplorerWidth {
		maxExplorerWidth = minExplorerWidth
	}
	t.State.ExplorerWidth += delta
	if t.State.ExplorerWidth < minExplorerWidth {
		t.State.ExplorerWidth = minExplorerWidth
	}
	if t.State.ExplorerWidth > maxExplorerWidth {
		t.State.ExplorerWidth = maxExplorerWidth
	}
}

// Navigation helpers

func (t *TUI) moveSelectionUp() {
	if t.selectedIndex > 0 {
		t.selectedIndex--
		
		// Scroll if needed
		if t.selectedIndex < t.State.TreeScrollY {
			t.State.TreeScrollY = t.selectedIndex
		}
	}
}

func (t *TUI) moveSelectionDown() {
	if t.selectedIndex < len(t.visibleNodes)-1 {
		t.selectedIndex++
		
		// Scroll if needed
		height := t.State.Height - 6
		if t.selectedIndex >= t.State.TreeScrollY+height {
			t.State.TreeScrollY = t.selectedIndex - height + 1
		}
	}
}

func (t *TUI) expandSelected() {
	if t.FileTree.Root == nil || t.selectedIndex >= len(t.visibleNodes) {
		return
	}
	
	node := t.visibleNodes[t.selectedIndex]
	if node.Type == model.FileTypeDirectory {
		t.FileTree.ExpandNode(node)
		t.visibleNodes = t.FileTree.GetVisibleNodes()
	}
}

func (t *TUI) collapseSelected() {
	if t.FileTree.Root == nil || t.selectedIndex >= len(t.visibleNodes) {
		return
	}
	
	node := t.visibleNodes[t.selectedIndex]
	if node.Type == model.FileTypeDirectory && node.Expanded {
		t.FileTree.CollapseNode(node)
		t.visibleNodes = t.FileTree.GetVisibleNodes()
	} else if node.Parent != nil {
		// Move to parent
		for i, n := range t.visibleNodes {
			if n == node.Parent {
				t.selectedIndex = i
				break
			}
		}
	}
}

func (t *TUI) openSelected() (tea.Model, tea.Cmd) {
	if t.FileTree.Root == nil || t.selectedIndex >= len(t.visibleNodes) {
		return t, nil
	}
	
	node := t.visibleNodes[t.selectedIndex]
	
	if node.Type == model.FileTypeDirectory {
		// Toggle directory
		t.FileTree.ToggleNode(node)
		t.visibleNodes = t.FileTree.GetVisibleNodes()
	} else {
		// Check for unsaved changes before switching files
		if t.Editor.IsModified() {
			t.State.Dialog.Type = model.DialogConfirmSwitch
			t.State.Dialog.TargetPath = node.Path
			t.dialogMessage = "Current file has unsaved changes. Save before switching?"
			return t, nil
		}
		
		// Open file
		t.Editor.OpenFile(node.Path)
		t.State.FocusPanel = model.PanelEditor
		t.State.Mode = model.ModeNormal
	}
	
	return t, nil
}

// getPasteDestDir returns the directory where paste should put the item (selected folder, or parent of selected file, or tree root).
func (t *TUI) getPasteDestDir() string {
	if t.FileTree.Root == nil {
		return ""
	}
	if t.selectedIndex < 0 || t.selectedIndex >= len(t.visibleNodes) {
		return t.FileTree.RootPath
	}
	node := t.visibleNodes[t.selectedIndex]
	if node.Type == model.FileTypeDirectory {
		return node.Path
	}
	if node.Parent != nil {
		return node.Parent.Path
	}
	return t.FileTree.RootPath
}

func (t *TUI) pasteFromClipboard() (tea.Model, tea.Cmd) {
	destDir := t.getPasteDestDir()
	if destDir == "" {
		return t, nil
	}
	name := filepath.Base(t.clipboardPath)
	destPath := filepath.Join(destDir, name)
	if destPath == t.clipboardPath {
		t.State.StatusMessage = "Cannot paste into same location"
		return t, nil
	}
	// Prevent copying a directory into itself or its subtree
	if !t.clipboardCut {
		clipAbs, _ := filepath.Abs(t.clipboardPath)
		destAbs, _ := filepath.Abs(destPath)
		if strings.HasPrefix(destAbs+string(filepath.Separator), clipAbs+string(filepath.Separator)) {
			t.State.StatusMessage = "Cannot copy folder into itself"
			return t, nil
		}
	}
	if _, err := os.Stat(destPath); err == nil {
		t.State.StatusMessage = "Destination already exists: " + name
		return t, nil
	}
	srcInfo, err := os.Stat(t.clipboardPath)
	if err != nil {
		t.State.StatusMessage = "Source no longer exists"
		t.clipboardPath = ""
		t.clipboardCut = false
		return t, nil
	}
	if t.clipboardCut {
		// Move
		if err := t.FileOps.RenameFile(t.clipboardPath, destPath); err != nil {
			t.State.StatusMessage = "Move failed: " + err.Error()
			return t, nil
		}
		if t.Editor.GetFilePath() == t.clipboardPath {
			t.Editor.Buffer.FilePath = destPath
			if t.Editor.Buffer.File != nil {
				t.Editor.Buffer.File.Path = destPath
				t.Editor.Buffer.File.Name = filepath.Base(destPath)
			}
		}
		t.clipboardPath = ""
		t.clipboardCut = false
		t.State.StatusMessage = "Moved: " + name
	} else {
		// Copy
		if srcInfo.IsDir() {
			if err := t.FileOps.CopyDirectory(t.clipboardPath, destPath); err != nil {
				t.State.StatusMessage = "Copy failed: " + err.Error()
				return t, nil
			}
		} else {
			if err := t.FileOps.CopyFile(t.clipboardPath, destPath); err != nil {
				t.State.StatusMessage = "Copy failed: " + err.Error()
				return t, nil
			}
		}
		t.State.StatusMessage = "Copied: " + name
	}
	if t.FileTree.Root != nil {
		t.FileTree.Refresh()
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		for i, node := range t.visibleNodes {
			if node.Path == destPath {
				t.selectedIndex = i
				break
			}
		}
	}
	return t, nil
}

// File operations

func (t *TUI) openDirectory(path string) (tea.Model, tea.Cmd) {
	expanded := utils.ExpandPath(path)
	
	if err := t.FileTree.LoadDirectory(expanded); err != nil {
		t.dialogError = "Error opening directory: " + err.Error()
		return t, nil
	}
	
	t.State.RootPath = expanded
	t.visibleNodes = t.FileTree.GetVisibleNodes()
	t.selectedIndex = 0
	t.State.Dialog.Type = model.DialogNone
	t.State.FocusPanel = model.PanelExplorer
	_ = config.SaveLastWorkspace(expanded)
	_ = t.FileTree.StartWatch(expanded)

	return t, t.FileTree.WatchCmd()
}

func (t *TUI) openFile(path string) (tea.Model, tea.Cmd) {
	expanded := utils.ExpandPath(path)
	
	// Check if it's a directory
	if t.FileOps.IsDirectory(expanded) {
		return t.openDirectory(expanded)
	}
	
	if err := t.Editor.OpenFile(expanded); err != nil {
		t.dialogError = "Error opening file: " + err.Error()
		return t, nil
	}
	
	// Check if the file is already in the current tree
	found := false
	for i, node := range t.visibleNodes {
		if node.Path == expanded {
			t.selectedIndex = i
			found = true
			break
		}
	}
	
	// If not found, check if the file's parent directory is within the current root
	if !found && t.FileTree.Root != nil {
		fileDir := filepath.Dir(expanded)
		
		// Check if the file's directory is within or is the current root
		if strings.HasPrefix(fileDir, t.FileTree.RootPath) || fileDir == t.FileTree.RootPath {
			// Try to find and expand the parent directory in the tree
			parentNode := t.FileTree.FindNode(fileDir)
			if parentNode != nil {
				// Expand the parent directory to show the file
				t.FileTree.ExpandNode(parentNode)
				t.visibleNodes = t.FileTree.GetVisibleNodes()
				
				// Find and select the file
				for i, node := range t.visibleNodes {
					if node.Path == expanded {
						t.selectedIndex = i
						found = true
						break
					}
				}
			}
		}
		
		// If still not found, refresh the tree while preserving expansion state
		if !found {
			t.FileTree.Refresh()
			t.visibleNodes = t.FileTree.GetVisibleNodes()
			
			// Try to find the file again
			for i, node := range t.visibleNodes {
				if node.Path == expanded {
					t.selectedIndex = i
					found = true
					break
				}
			}
		}
	}
	
	t.State.Dialog.Type = model.DialogNone
	t.State.FocusPanel = model.PanelEditor
	t.State.Mode = model.ModeNormal
	_ = config.SaveLastWorkspace(expanded)

	return t, nil
}

func (t *TUI) createNewFile(path string) (tea.Model, tea.Cmd) {
	expanded := utils.ExpandPath(path)
	
	if !utils.IsValidFileName(filepath.Base(expanded)) {
		t.dialogError = "Invalid file name"
		return t, nil
	}
	
	if err := t.FileOps.CreateFile(expanded); err != nil {
		t.dialogError = "Error creating file: " + err.Error()
		return t, nil
	}
	
	// Refresh explorer while preserving expansion state
	if t.FileTree.Root != nil {
		t.FileTree.Refresh()
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		
		// Select the newly created file
		for i, node := range t.visibleNodes {
			if node.Path == expanded {
				t.selectedIndex = i
				break
			}
		}
	}
	
	// Open the new file
	t.Editor.OpenFile(expanded)
	t.State.Dialog.Type = model.DialogNone
	t.State.FocusPanel = model.PanelEditor
	t.State.Mode = model.ModeInsert
	
	return t, nil
}

func (t *TUI) createNewFolder(path string) (tea.Model, tea.Cmd) {
	expanded := utils.ExpandPath(path)
	
	if !utils.IsValidFileName(filepath.Base(expanded)) {
		t.dialogError = "Invalid folder name"
		return t, nil
	}
	
	if err := t.FileOps.CreateDirectory(expanded); err != nil {
		t.dialogError = "Error creating folder: " + err.Error()
		return t, nil
	}
	
	// Refresh explorer while preserving expansion state
	if t.FileTree.Root != nil {
		t.FileTree.Refresh()
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		
		// Select the newly created folder
		for i, node := range t.visibleNodes {
			if node.Path == expanded {
				t.selectedIndex = i
				break
			}
		}
	}
	
	t.State.Dialog.Type = model.DialogNone
	
	return t, nil
}

func (t *TUI) renameItem(newName string) (tea.Model, tea.Cmd) {
	if !utils.IsValidFileName(newName) {
		t.dialogError = "Invalid name"
		return t, nil
	}
	
	oldPath := t.State.Dialog.TargetPath
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)
	
	if err := t.FileOps.RenameFile(oldPath, newPath); err != nil {
		t.dialogError = "Error renaming: " + err.Error()
		return t, nil
	}
	
	// Refresh explorer while preserving expansion state
	if t.FileTree.Root != nil {
		t.FileTree.Refresh()
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		
		// Select the renamed item
		for i, node := range t.visibleNodes {
			if node.Path == newPath {
				t.selectedIndex = i
				break
			}
		}
	}
	
	// Update editor if the renamed file is open
	if t.Editor.GetFilePath() == oldPath {
		t.Editor.Buffer.FilePath = newPath
		t.Editor.Buffer.File.Path = newPath
		t.Editor.Buffer.File.Name = newName
	}
	
	t.State.Dialog.Type = model.DialogNone
	
	return t, nil
}

func (t *TUI) deleteItem() (tea.Model, tea.Cmd) {
	path := t.State.Dialog.TargetPath
	
	var err error
	if t.FileOps.IsDirectory(path) {
		err = t.FileOps.DeleteDirectory(path)
	} else {
		err = t.FileOps.DeleteFile(path)
	}
	
	if err != nil {
		t.dialogError = "Error deleting: " + err.Error()
		return t, nil
	}
	
	// Close editor if deleted file was open
	if t.Editor.GetFilePath() == path {
		t.Editor = newEditorInstance()
	}
	
	// Refresh explorer while preserving expansion state
	if t.FileTree.Root != nil {
		// Try to find the next sibling or parent to select after deletion
		var nextPath string
		for i, node := range t.visibleNodes {
			if node.Path == path {
				// Try to select the next item
				if i+1 < len(t.visibleNodes) {
					nextPath = t.visibleNodes[i+1].Path
				} else if i > 0 {
					nextPath = t.visibleNodes[i-1].Path
				}
				break
			}
		}
		
		t.FileTree.Refresh()
		t.visibleNodes = t.FileTree.GetVisibleNodes()
		
		// Select the next appropriate item
		if nextPath != "" {
			for i, node := range t.visibleNodes {
				if node.Path == nextPath {
					t.selectedIndex = i
					break
				}
			}
		}
		
		// Clamp selection to valid range
		if t.selectedIndex >= len(t.visibleNodes) {
			t.selectedIndex = len(t.visibleNodes) - 1
		}
		if t.selectedIndex < 0 {
			t.selectedIndex = 0
		}
	}
	
	t.State.Dialog.Type = model.DialogNone
	
	return t, nil
}

// Word movement helpers

func (t *TUI) moveWordForward() {
	if t.Editor.Buffer == nil {
		return
	}
	
	line := t.Editor.Buffer.Cursor.Line
	col := t.Editor.Buffer.Cursor.Column
	
	if line >= len(t.Editor.Buffer.Lines) {
		return
	}
	
	text := t.Editor.Buffer.Lines[line]
	
	// Skip current word
	for col < len(text) && !isWordChar(text[col]) {
		col++
	}
	
	// Skip to next word
	for col < len(text) && isWordChar(text[col]) {
		col++
	}
	
	if col >= len(text) && line < len(t.Editor.Buffer.Lines)-1 {
		// Move to next line
		t.Editor.Buffer.Cursor.Line++
		t.Editor.Buffer.Cursor.Column = 0
	} else {
		t.Editor.Buffer.Cursor.Column = col
	}
}

func (t *TUI) moveWordBackward() {
	if t.Editor.Buffer == nil {
		return
	}
	
	line := t.Editor.Buffer.Cursor.Line
	col := t.Editor.Buffer.Cursor.Column
	
	if line == 0 && col == 0 {
		return
	}
	
	if col == 0 && line > 0 {
		// Move to previous line
		t.Editor.Buffer.Cursor.Line--
		t.Editor.Buffer.Cursor.Column = len(t.Editor.Buffer.Lines[t.Editor.Buffer.Cursor.Line])
		return
	}
	
	text := t.Editor.Buffer.Lines[line]
	
	// Skip to previous word
	for col > 0 && !isWordChar(text[col-1]) {
		col--
	}
	
	for col > 0 && isWordChar(text[col-1]) {
		col--
	}
	
	t.Editor.Buffer.Cursor.Column = col
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// Helper to create new editor
func newEditorInstance() *editor.Editor {
	return editor.NewEditor()
}

// Helper for strings
func init() {
	_ = strings.Builder{}
}
