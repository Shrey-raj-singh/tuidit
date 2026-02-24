package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileOperations provides file system operations
type FileOperations struct{}

// NewFileOperations creates a new FileOperations instance
func NewFileOperations() *FileOperations {
	return &FileOperations{}
}

// CreateFile creates a new file
func (f *FileOperations) CreateFile(path string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	file.Close()

	return nil
}

// CreateDirectory creates a new directory
func (f *FileOperations) CreateDirectory(path string) error {
	// Check if directory already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("directory already exists: %s", path)
	}

	// Create the directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// DeleteFile deletes a file
func (f *FileOperations) DeleteFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteDirectory deletes a directory and all its contents
func (f *FileOperations) DeleteDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// RenameFile renames a file or directory
func (f *FileOperations) RenameFile(oldPath, newPath string) error {
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("source does not exist: %s", oldPath)
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("destination already exists: %s", newPath)
	}

	// Create parent directories for destination if needed
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

// MoveFile moves a file to a new location
func (f *FileOperations) MoveFile(src, dst string) error {
	return f.RenameFile(src, dst)
}

// CopyFile copies a file to a new location
func (f *FileOperations) CopyFile(src, dst string) error {
	// Read source file
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create parent directories for destination if needed
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(dst, content, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// CopyDirectory recursively copies a directory to a new location
func (f *FileOperations) CopyDirectory(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := f.CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := f.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// IsDirectory checks if a path is a directory
func (f *FileOperations) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists checks if a file exists
func (f *FileOperations) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileInfo returns information about a file
func (f *FileOperations) GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}

// ExpandPath expands ~ to home directory
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(GetHomeDir(), path[2:])
	}
	return path
}

// GetAbsolutePath returns the absolute path
func GetAbsolutePath(path string) (string, error) {
	expanded := ExpandPath(path)
	return filepath.Abs(expanded)
}

// ListDirectory lists the contents of a directory
func (f *FileOperations) ListDirectory(path string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	return entries, nil
}

// GetFileExtension returns the file extension
func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// GetFileName returns the file name without extension
func GetFileName(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}

// IsValidFileName checks if a file name is valid
func IsValidFileName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	
	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return false
		}
	}
	
	return true
}
