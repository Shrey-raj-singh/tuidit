package main

import (
	"fmt"
	"os"
	"path/filepath"
	
	tea "github.com/charmbracelet/bubbletea"
	"tuidit/internal/model"
	"tuidit/internal/tui"
	"tuidit/internal/utils"
)

func main() {
	// Parse command line arguments
	var startPath string
	
	if len(os.Args) > 1 {
		startPath = os.Args[1]
	}
	
	// Create TUI application
	app := tui.NewTUI()
	
	// If a path is provided, try to open it
	if startPath != "" {
		expanded := utils.ExpandPath(startPath)
		absPath, err := filepath.Abs(expanded)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}
		
		// Check if path exists
		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		// Set up the app based on path type
		app.State.ShowStartup = false
		app.State.RootPath = absPath
		
		if info.IsDir() {
			// Open directory
			if err := app.FileTree.LoadDirectory(absPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading directory: %v\n", err)
				os.Exit(1)
			}
			app.State.FocusPanel = model.PanelExplorer
		} else {
			// Open file
			if err := app.Editor.OpenFile(absPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
				os.Exit(1)
			}
			// Also load parent directory
			dir := filepath.Dir(absPath)
			app.FileTree.LoadDirectory(dir)
			app.State.FocusPanel = model.PanelEditor
			app.State.Mode = model.ModeNormal
		}
		
		app.State.StartupMode = model.StartupDirect
	}
	
	// Run the application
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

