package config

import (
	"os"
	"path/filepath"
	"strings"
)

const lastWorkspaceFileName = "last_workspace"

// lastWorkspacePath returns the path to the last workspace file (~/.tuidit/last_workspace).
func lastWorkspacePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tuidit", lastWorkspaceFileName), nil
}

// GetLastWorkspace returns the last used workspace path, or empty string if none or invalid.
func GetLastWorkspace() string {
	p, err := lastWorkspacePath()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	path := strings.TrimSpace(string(b))
	if path == "" {
		return ""
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return ""
	}
	return path
}

// SaveLastWorkspace saves a workspace path (directory). Pass the directory to restore next time.
// If path is a file, its parent directory is saved.
func SaveLastWorkspace(path string) error {
	if path == "" {
		return nil
	}
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		path = filepath.Dir(path)
	}
	p, err := lastWorkspacePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(path), 0644)
}
