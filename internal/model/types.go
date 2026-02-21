package model

import (
	"time"
)

// FocusPanel represents which panel is currently focused
type FocusPanel int

const (
	PanelExplorer FocusPanel = iota
	PanelEditor
	PanelDialog
)

// FileMode represents the mode the editor is in
type FileMode int

const (
	ModeNormal FileMode = iota
	ModeInsert
	ModeCommand
	ModeVisual
)

// FileType represents the type of file
type FileType int

const (
	FileTypeFile FileType = iota
	FileTypeDirectory
	FileTypeSymlink
)

// TreeNode represents a node in the file tree
type TreeNode struct {
	Name      string
	Path      string
	Type      FileType
	Expanded  bool
	Children  []*TreeNode
	Parent    *TreeNode
	IsLoaded  bool
}

// FileNode represents a single file with its content
type FileNode struct {
	Path     string
	Name     string
	Content  []string
	Modified bool
	LastRead time.Time
}

// Cursor represents cursor position
type Cursor struct {
	Line   int
	Column int
}

// Selection represents text selection
type Selection struct {
	Start Cursor
	End   Cursor
	Active bool
}

// EditorBuffer represents an open file buffer
type EditorBuffer struct {
	File       *FileNode
	Lines      []string
	Cursor     Cursor
	Selection  Selection
	ScrollY    int
	ScrollX    int
	Modified   bool
	FilePath   string
}

// DialogType represents the type of dialog shown
type DialogType int

const (
	DialogNone DialogType = iota
	DialogOpenFile
	DialogOpenDir
	DialogNewFile
	DialogNewFolder
	DialogRename
	DialogDelete
	DialogSave
	DialogConfirmSwitch
	DialogQuit
	DialogHelp
)

// Dialog represents a dialog state
type Dialog struct {
	Type        DialogType
	Input       string
	Message     string
	TargetPath  string
	Preview     []string
	PreviewIdx  int
}

// OperationType represents file operation types
type OperationType int

const (
	OpNone OperationType = iota
	OpCreateFile
	OpCreateFolder
	OpRename
	OpDelete
	OpSave
)

// PendingOperation represents a pending file operation
type PendingOperation struct {
	Type      OperationType
	Source    string
	Target    string
	Confirmed bool
}
