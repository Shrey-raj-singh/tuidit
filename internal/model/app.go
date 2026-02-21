package model

import (
	"github.com/charmbracelet/bubbletea"
)

// AppState represents the main application state
type AppState struct {
	// Mode and Focus
	Mode       FileMode
	FocusPanel FocusPanel
	
	// File Explorer
	RootPath     string
	FileTree     *TreeNode
	SelectedNode *TreeNode
	TreeScrollY  int
	
	// Editor
	Buffers      []*EditorBuffer
	ActiveBuffer *EditorBuffer
	
	// Dialog
	Dialog       Dialog
	
	// Dimensions
	Width        int
	Height       int
	ExplorerWidth int
	
	// Status
	StatusMessage string
	ShowHelp      bool
	
	// Startup
	StartupMode   StartupMode
	ShowStartup   bool
}

// StartupMode represents how the editor was started
type StartupMode int

const (
	StartupSelect StartupMode = iota
	StartupDirect
)

// Init initializes the app state
func (a *AppState) Init() tea.Cmd {
	return nil
}

// NewAppState creates a new application state
func NewAppState() *AppState {
	return &AppState{
		Mode:          ModeNormal,
		FocusPanel:    PanelExplorer,
		Buffers:       make([]*EditorBuffer, 0),
		ExplorerWidth: 30,
		ShowStartup:   true,
		Dialog: Dialog{
			Type: DialogNone,
		},
	}
}

// GetVisibleLines returns all visible tree nodes in order
func (a *AppState) GetVisibleLines() []*TreeNode {
	if a.FileTree == nil {
		return nil
	}
	
	var result []*TreeNode
	var traverse func(node *TreeNode, depth int)
	
	traverse = func(node *TreeNode, depth int) {
		result = append(result, node)
		if node.Expanded && len(node.Children) > 0 {
			for _, child := range node.Children {
				traverse(child, depth+1)
			}
		}
	}
	
	for _, child := range a.FileTree.Children {
		traverse(child, 0)
	}
	
	return result
}

// FindNodeByPath finds a node by its path
func (a *AppState) FindNodeByPath(path string) *TreeNode {
	if a.FileTree == nil {
		return nil
	}
	
	var search func(node *TreeNode) *TreeNode
	search = func(node *TreeNode) *TreeNode {
		if node.Path == path {
			return node
		}
		for _, child := range node.Children {
			if found := search(child); found != nil {
				return found
			}
		}
		return nil
	}
	
	return search(a.FileTree)
}
