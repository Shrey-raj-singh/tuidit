package config

import (
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	// Editor settings
	TabSize      int
	ShowLineNum  bool
	WordWrap     bool
	AutoSave     bool
	AutoIndent   bool
	
	// UI settings
	ExplorerWidth int
	ShowStatusBar bool
	Theme        string
	
	// Keybindings
	KeySave      string
	KeyQuit      string
	KeyOpen      string
	KeyNewFile   string
	KeyNewFolder string
	KeyDelete    string
	KeyRename    string
	KeyHelp      string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		TabSize:       4,
		ShowLineNum:   true,
		WordWrap:      false,
		AutoSave:      false,
		AutoIndent:    true,
		ExplorerWidth: 30,
		ShowStatusBar: true,
		Theme:         "default",
		KeySave:       "ctrl+s",
		KeyQuit:       "ctrl+q",
		KeyOpen:       "ctrl+o",
		KeyNewFile:    "ctrl+n",
		KeyNewFolder:  "ctrl+shift+n",
		KeyDelete:     "delete",
		KeyRename:     "f2",
		KeyHelp:       "f1",
	}
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()
	
	// Check if config file exists
	configPath := path
	if configPath == "" {
		// Use default config location
		home, err := os.UserHomeDir()
		if err != nil {
			return config, nil
		}
		configPath = filepath.Join(home, ".tuidit", "config.json")
	}
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}
	
	// TODO: Parse config file
	
	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, path string) error {
	configPath := path
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(home, ".tuidit", "config.json")
	}
	
	// Create directory if needed
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// TODO: Write config file
	
	return nil
}
