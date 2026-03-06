package explorer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"tuidit/internal/model"
)

// FileTree manages the file tree structure
type FileTree struct {
	Root    *model.TreeNode
	RootPath string

	// watcher for filesystem changes (nil if not watching)
	watcher *fsnotify.Watcher
	watchCh chan struct{}
}

// NewFileTree creates a new file tree
func NewFileTree() *FileTree {
	return &FileTree{}
}

// LoadDirectory loads a directory into the file tree
func (ft *FileTree) LoadDirectory(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	
	ft.RootPath = absPath
	
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	
	ft.Root = &model.TreeNode{
		Name:     filepath.Base(absPath),
		Path:     absPath,
		Type:     model.FileTypeDirectory,
		Expanded: true,
		IsLoaded: false,
	}
	
	if info.IsDir() {
		ft.loadChildren(ft.Root)
	} else {
		ft.Root.Type = model.FileTypeFile
	}
	
	return nil
}

// loadChildren loads the children of a directory node
func (ft *FileTree) loadChildren(node *model.TreeNode) error {
	if node.Type != model.FileTypeDirectory {
		return nil
	}
	
	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return err
	}
	
	node.Children = make([]*model.TreeNode, 0)
	
	// Sort entries: directories first, then files, alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})
	
	for _, entry := range entries {
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		
		childPath := filepath.Join(node.Path, entry.Name())
		childType := model.FileTypeFile
		if entry.IsDir() {
			childType = model.FileTypeDirectory
		}
		
		// Check for symlink
		if entry.Type()&os.ModeSymlink != 0 {
			childType = model.FileTypeSymlink
		}
		
		child := &model.TreeNode{
			Name:     entry.Name(),
			Path:     childPath,
			Type:     childType,
			Expanded: false,
			Parent:   node,
			IsLoaded: false,
		}
		
		node.Children = append(node.Children, child)
	}
	
	node.IsLoaded = true
	return nil
}

// ToggleNode expands or collapses a directory node
func (ft *FileTree) ToggleNode(node *model.TreeNode) error {
	if node.Type != model.FileTypeDirectory {
		return nil
	}
	
	if !node.IsLoaded {
		if err := ft.loadChildren(node); err != nil {
			return err
		}
	}
	
	node.Expanded = !node.Expanded
	return nil
}

// ExpandNode expands a directory node
func (ft *FileTree) ExpandNode(node *model.TreeNode) error {
	if node.Type != model.FileTypeDirectory {
		return nil
	}
	
	if !node.IsLoaded {
		if err := ft.loadChildren(node); err != nil {
			return err
		}
	}
	
	node.Expanded = true
	return nil
}

// CollapseNode collapses a directory node
func (ft *FileTree) CollapseNode(node *model.TreeNode) {
	if node.Type == model.FileTypeDirectory {
		node.Expanded = false
	}
}

// RefreshNode refreshes the children of a node while preserving expansion state
func (ft *FileTree) RefreshNode(node *model.TreeNode) error {
	if node.Type != model.FileTypeDirectory {
		return nil
	}
	
	// Save expansion state of existing children
	expandedStates := make(map[string]bool)
	for _, child := range node.Children {
		expandedStates[child.Path] = child.Expanded
	}
	
	node.IsLoaded = false
	node.Children = nil
	if err := ft.loadChildren(node); err != nil {
		return err
	}
	
	// Restore expansion state for new children
	for _, child := range node.Children {
		if expanded, ok := expandedStates[child.Path]; ok {
			child.Expanded = expanded
			// If child was expanded and is a directory, refresh its children too
			if child.Expanded && child.Type == model.FileTypeDirectory {
				ft.RefreshNode(child)
			}
		}
	}
	
	return nil
}

// Refresh refreshes the entire tree
func (ft *FileTree) Refresh() error {
	if ft.Root == nil {
		return nil
	}
	return ft.RefreshNode(ft.Root)
}

// FindNode finds a node by path
func (ft *FileTree) FindNode(path string) *model.TreeNode {
	if ft.Root == nil {
		return nil
	}
	
	var search func(node *model.TreeNode) *model.TreeNode
	search = func(node *model.TreeNode) *model.TreeNode {
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
	
	return search(ft.Root)
}

// GetVisibleNodes returns all visible nodes in the tree
func (ft *FileTree) GetVisibleNodes() []*model.TreeNode {
	if ft.Root == nil {
		return nil
	}
	
	var result []*model.TreeNode
	var traverse func(node *model.TreeNode)
	
	traverse = func(node *model.TreeNode) {
		result = append(result, node)
		if node.Expanded && len(node.Children) > 0 {
			for _, child := range node.Children {
				traverse(child)
			}
		}
	}
	
	traverse(ft.Root)
	return result
}

// GetNodeDepth returns the depth of a node in the tree
func GetNodeDepth(node *model.TreeNode) int {
	depth := 0
	for node.Parent != nil {
		depth++
		node = node.Parent
	}
	return depth
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// StartWatch starts watching the tree root (and subdirs) for filesystem changes.
// Sends on watchCh when a change is detected (debounced). Call StopWatch before starting a new watch.
func (ft *FileTree) StartWatch(rootPath string) error {
	if rootPath == "" {
		return nil
	}
	ft.StopWatch()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	ft.watcher = watcher
	ft.watchCh = make(chan struct{}, 1)

	// Add root and all subdirs
	_ = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			_ = watcher.Add(path)
		}
		return nil
	})

	go func() {
		defer close(ft.watchCh)
		var debounce *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				_ = event
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(150*time.Millisecond, func() {
					select {
					case ft.watchCh <- struct{}{}:
					default:
					}
				})
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return nil
}

// StopWatch stops the filesystem watcher.
func (ft *FileTree) StopWatch() {
	if ft.watcher == nil {
		return
	}
	_ = ft.watcher.Close()
	ft.watcher = nil
	ft.watchCh = nil
}

// WatchCmd returns a Bubble Tea command that completes when a filesystem change is detected.
// Call this from Init/Update and re-return it after handling DirChangedMsg to keep watching.
func (ft *FileTree) WatchCmd() tea.Cmd {
	if ft.watchCh == nil {
		return nil
	}
	ch := ft.watchCh
	return func() tea.Msg {
		_, ok := <-ch
		if !ok {
			return nil
		}
		return model.DirChangedMsg{}
	}
}
