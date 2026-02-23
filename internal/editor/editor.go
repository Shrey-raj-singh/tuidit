package editor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
	
	"tuidit/internal/model"
)

// Editor represents a text editor
type Editor struct {
	Buffer *model.EditorBuffer
}

// NewEditor creates a new editor
func NewEditor() *Editor {
	return &Editor{
		Buffer: nil,
	}
}

// OpenFile opens a file in the editor
func (e *Editor) OpenFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new empty buffer
			e.Buffer = &model.EditorBuffer{
				File: &model.FileNode{
					Path:     absPath,
					Name:     filepath.Base(absPath),
					Content:  []string{},
					Modified: false,
				},
				Lines:    []string{""},
				Cursor:   model.Cursor{Line: 0, Column: 0},
				FilePath: absPath,
			}
			return nil
		}
		return err
	}

	// Don't open directories
	if info.IsDir() {
		return fmt.Errorf("cannot open directory as file")
	}

	// Read file content
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(lines) == 0 {
		lines = []string{""}
	}

	e.Buffer = &model.EditorBuffer{
		File: &model.FileNode{
			Path:     absPath,
			Name:     filepath.Base(absPath),
			Content:  lines,
			Modified: false,
		},
		Lines:    lines,
		Cursor:   model.Cursor{Line: 0, Column: 0},
		FilePath: absPath,
	}

	return nil
}

// SaveFile saves the current buffer to disk
func (e *Editor) SaveFile() error {
	if e.Buffer == nil {
		return fmt.Errorf("no file open")
	}

	// Create parent directories if needed
	dir := filepath.Dir(e.Buffer.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(e.Buffer.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range e.Buffer.Lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	e.Buffer.Modified = false
	e.Buffer.File.Modified = false
	return nil
}

// SaveFileAs saves the current buffer to a new path
func (e *Editor) SaveFileAs(path string) error {
	if e.Buffer == nil {
		return fmt.Errorf("no file open")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	e.Buffer.FilePath = absPath
	e.Buffer.File.Path = absPath
	e.Buffer.File.Name = filepath.Base(absPath)

	return e.SaveFile()
}

// InsertChar inserts a character at the current cursor position
func (e *Editor) InsertChar(ch rune) {
	if e.Buffer == nil {
		return
	}

	line := e.Buffer.Cursor.Line
	col := e.Buffer.Cursor.Column

	if line < 0 || line >= len(e.Buffer.Lines) {
		return
	}

	currentLine := e.Buffer.Lines[line]
	
	// Handle tab
	if ch == '\t' {
		ch = ' '
		for i := 0; i < 4; i++ {
			currentLine = currentLine[:col] + string(ch) + currentLine[col:]
			col++
		}
		e.Buffer.Lines[line] = currentLine
		e.Buffer.Cursor.Column = col
		e.Buffer.Modified = true
		return
	}

	// Insert character
	if col > len(currentLine) {
		col = len(currentLine)
	}
	
	newLine := currentLine[:col] + string(ch) + currentLine[col:]
	e.Buffer.Lines[line] = newLine
	e.Buffer.Cursor.Column++
	e.Buffer.Modified = true
}

// InsertNewline inserts a newline at the current cursor position
func (e *Editor) InsertNewline() {
	if e.Buffer == nil {
		return
	}

	line := e.Buffer.Cursor.Line
	col := e.Buffer.Cursor.Column

	if line < 0 || line >= len(e.Buffer.Lines) {
		return
	}

	currentLine := e.Buffer.Lines[line]
	
	// Split line at cursor
	before := currentLine[:col]
	after := currentLine[col:]
	
	// Insert new line
	newLines := make([]string, 0)
	newLines = append(newLines, e.Buffer.Lines[:line]...)
	newLines = append(newLines, before)
	newLines = append(newLines, after)
	newLines = append(newLines, e.Buffer.Lines[line+1:]...)
	
	e.Buffer.Lines = newLines
	e.Buffer.Cursor.Line++
	e.Buffer.Cursor.Column = 0
	e.Buffer.Modified = true
}

// DeleteChar deletes the character before the cursor (backspace)
func (e *Editor) DeleteChar() {
	if e.Buffer == nil {
		return
	}

	line := e.Buffer.Cursor.Line
	col := e.Buffer.Cursor.Column

	if line < 0 || line >= len(e.Buffer.Lines) {
		return
	}

	// If at beginning of line and not first line, merge with previous line
	if col == 0 && line > 0 {
		prevLine := e.Buffer.Lines[line-1]
		currentLine := e.Buffer.Lines[line]
		
		newLines := make([]string, 0)
		newLines = append(newLines, e.Buffer.Lines[:line-1]...)
		newLines = append(newLines, prevLine+currentLine)
		newLines = append(newLines, e.Buffer.Lines[line+1:]...)
		
		e.Buffer.Lines = newLines
		e.Buffer.Cursor.Line--
		e.Buffer.Cursor.Column = utf8.RuneCountInString(prevLine)
		e.Buffer.Modified = true
		return
	}

	// Delete character before cursor
	if col > 0 {
		currentLine := e.Buffer.Lines[line]
		newLine := currentLine[:col-1] + currentLine[col:]
		e.Buffer.Lines[line] = newLine
		e.Buffer.Cursor.Column--
		e.Buffer.Modified = true
	}
}

// DeleteCharForward deletes the character at the cursor (delete key)
func (e *Editor) DeleteCharForward() {
	if e.Buffer == nil {
		return
	}

	line := e.Buffer.Cursor.Line
	col := e.Buffer.Cursor.Column

	if line < 0 || line >= len(e.Buffer.Lines) {
		return
	}

	currentLine := e.Buffer.Lines[line]
	
	// If at end of line and not last line, merge with next line
	if col >= utf8.RuneCountInString(currentLine) && line < len(e.Buffer.Lines)-1 {
		nextLine := e.Buffer.Lines[line+1]
		
		newLines := make([]string, 0)
		newLines = append(newLines, e.Buffer.Lines[:line]...)
		newLines = append(newLines, currentLine+nextLine)
		newLines = append(newLines, e.Buffer.Lines[line+2:]...)
		
		e.Buffer.Lines = newLines
		e.Buffer.Modified = true
		return
	}

	// Delete character at cursor
	if col < utf8.RuneCountInString(currentLine) {
		newLine := currentLine[:col] + currentLine[col+1:]
		e.Buffer.Lines[line] = newLine
		e.Buffer.Modified = true
	}
}

// MoveCursor moves the cursor in the specified direction
func (e *Editor) MoveCursor(direction string) {
	if e.Buffer == nil {
		return
	}

	switch direction {
	case "up":
		if e.Buffer.Cursor.Line > 0 {
			e.Buffer.Cursor.Line--
			e.clampColumn()
		}
	case "down":
		if e.Buffer.Cursor.Line < len(e.Buffer.Lines)-1 {
			e.Buffer.Cursor.Line++
			e.clampColumn()
		}
	case "left":
		if e.Buffer.Cursor.Column > 0 {
			e.Buffer.Cursor.Column--
		} else if e.Buffer.Cursor.Line > 0 {
			e.Buffer.Cursor.Line--
			e.Buffer.Cursor.Column = utf8.RuneCountInString(e.Buffer.Lines[e.Buffer.Cursor.Line])
		}
	case "right":
		lineLen := utf8.RuneCountInString(e.Buffer.Lines[e.Buffer.Cursor.Line])
		if e.Buffer.Cursor.Column < lineLen {
			e.Buffer.Cursor.Column++
		} else if e.Buffer.Cursor.Line < len(e.Buffer.Lines)-1 {
			e.Buffer.Cursor.Line++
			e.Buffer.Cursor.Column = 0
		}
	case "home":
		e.Buffer.Cursor.Column = 0
	case "end":
		e.Buffer.Cursor.Column = utf8.RuneCountInString(e.Buffer.Lines[e.Buffer.Cursor.Line])
	}
}

// clampColumn ensures cursor column is within line bounds
func (e *Editor) ClampColumn() {
	if e.Buffer == nil {
		return
	}
	
	lineLen := utf8.RuneCountInString(e.Buffer.Lines[e.Buffer.Cursor.Line])
	if e.Buffer.Cursor.Column > lineLen {
		e.Buffer.Cursor.Column = lineLen
	}
}

// clampColumn ensures cursor column is within line bounds (internal)
func (e *Editor) clampColumn() {
	e.ClampColumn()
}

// ScrollToCursor ensures cursor is visible
func (e *Editor) ScrollToCursor(height int) {
	if e.Buffer == nil {
		return
	}

	cursor := e.Buffer.Cursor.Line
	
	// Scroll up if cursor is above visible area
	if cursor < e.Buffer.ScrollY {
		e.Buffer.ScrollY = cursor
	}
	
	// Scroll down if cursor is below visible area
	if cursor >= e.Buffer.ScrollY+height {
		e.Buffer.ScrollY = cursor - height + 1
	}
}

// GetVisibleLines returns the lines visible in the viewport
func (e *Editor) GetVisibleLines(height int) []string {
	if e.Buffer == nil {
		return []string{}
	}

	start := e.Buffer.ScrollY
	end := start + height
	
	if end > len(e.Buffer.Lines) {
		end = len(e.Buffer.Lines)
	}
	
	if start > len(e.Buffer.Lines) {
		start = len(e.Buffer.Lines)
	}
	
	return e.Buffer.Lines[start:end]
}

// GetCurrentLine returns the current line
func (e *Editor) GetCurrentLine() string {
	if e.Buffer == nil || e.Buffer.Cursor.Line >= len(e.Buffer.Lines) {
		return ""
	}
	return e.Buffer.Lines[e.Buffer.Cursor.Line]
}

// SetModified sets the modified flag
func (e *Editor) SetModified(modified bool) {
	if e.Buffer != nil {
		e.Buffer.Modified = modified
		e.Buffer.File.Modified = modified
	}
}

// IsModified returns whether the buffer is modified
func (e *Editor) IsModified() bool {
	return e.Buffer != nil && e.Buffer.Modified
}

// GetFileName returns the current file name
func (e *Editor) GetFileName() string {
	if e.Buffer == nil || e.Buffer.File == nil {
		return "[No File]"
	}
	return e.Buffer.File.Name
}

// GetFilePath returns the current file path
func (e *Editor) GetFilePath() string {
	if e.Buffer == nil {
		return ""
	}
	return e.Buffer.FilePath
}

// GetLineCount returns the number of lines
func (e *Editor) GetLineCount() int {
	if e.Buffer == nil {
		return 0
	}
	return len(e.Buffer.Lines)
}

// GetCursorPosition returns the cursor position (1-indexed for display)
func (e *Editor) GetCursorPosition() (line, col int) {
	if e.Buffer == nil {
		return 1, 1
	}
	return e.Buffer.Cursor.Line + 1, e.Buffer.Cursor.Column + 1
}

// GoToLine moves cursor to specified line (1-indexed)
func (e *Editor) GoToLine(lineNum int) {
	if e.Buffer == nil {
		return
	}
	
	if lineNum < 1 {
		lineNum = 1
	}
	if lineNum > len(e.Buffer.Lines) {
		lineNum = len(e.Buffer.Lines)
	}
	
	e.Buffer.Cursor.Line = lineNum - 1
	e.Buffer.Cursor.Column = 0
	e.clampColumn()
}

// Search searches for a string in the buffer
func (e *Editor) Search(query string) (found bool, line, col int) {
	if e.Buffer == nil || query == "" {
		return false, 0, 0
	}
	
	startLine := e.Buffer.Cursor.Line
	startCol := e.Buffer.Cursor.Column
	
	for i := startLine; i < len(e.Buffer.Lines); i++ {
		lineText := e.Buffer.Lines[i]
		searchStart := 0
		if i == startLine {
			searchStart = startCol
		}
		
		if idx := strings.Index(lineText[searchStart:], query); idx != -1 {
			return true, i, searchStart + idx
		}
	}
	
	// Wrap around
	for i := 0; i <= startLine; i++ {
		lineText := e.Buffer.Lines[i]
		endCol := len(lineText)
		if i == startLine {
			endCol = startCol
		}
		
		if idx := strings.Index(lineText[:endCol], query); idx != -1 {
			return true, i, idx
		}
	}
	
	return false, 0, 0
}

// FindNext finds the next occurrence of the query
func (e *Editor) FindNext(query string) bool {
	found, line, col := e.Search(query)
	if found {
		e.Buffer.Cursor.Line = line
		e.Buffer.Cursor.Column = col
		return true
	}
	return false
}
