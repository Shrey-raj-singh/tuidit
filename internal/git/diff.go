package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type LineStatus int

const (
	StatusNone     LineStatus = iota
	StatusAdded
	StatusModified
	StatusDeleted
)

type FileGitStatus int

const (
	FileClean       FileGitStatus = iota
	FileModified                  // tracked, modified in worktree or index
	FileAdded                     // new file staged
	FileUntracked                 // not tracked by git
	FileDeleted                   // deleted in worktree or index
	FileRenamed                   // renamed
	FileConflicted                // merge conflict
	FileIgnored                   // listed in .gitignore
)

func gitAvailable() bool {
	p, err := exec.LookPath("git")
	return err == nil && p != ""
}

func repoRoot(dir string) (string, bool) {
	if !gitAvailable() {
		return "", false
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// GetFileGutter returns a per-line gutter status for the given file by
// comparing the working copy against HEAD. Returns nil on any error.
func GetFileGutter(filePath string, totalLines int) []LineStatus {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil
	}
	dir := filepath.Dir(absPath)

	if _, ok := repoRoot(dir); !ok {
		return nil
	}

	gitPath := filepath.ToSlash(absPath)

	lsCmd := exec.Command("git", "ls-files", "--error-unmatch", gitPath)
	lsCmd.Dir = dir
	if err := lsCmd.Run(); err != nil {
		// Untracked file — mark all lines as added
		status := make([]LineStatus, totalLines)
		for i := range status {
			status[i] = StatusAdded
		}
		return status
	}

	cmd := exec.Command("git", "diff", "HEAD", "--unified=0", "--no-color", "--", gitPath)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) == 0 {
			return make([]LineStatus, totalLines)
		}
		return nil
	}

	diffOutput := string(out)
	if strings.TrimSpace(diffOutput) == "" {
		return make([]LineStatus, totalLines)
	}

	return parseDiff(diffOutput, totalLines)
}

func parseDiff(diff string, totalLines int) []LineStatus {
	status := make([]LineStatus, totalLines)
	lines := strings.Split(diff, "\n")

	for _, line := range lines {
		if !strings.HasPrefix(line, "@@") {
			continue
		}
		newStart, newCount, oldCount := parseHunkHeader(line)
		if newStart < 1 {
			continue
		}

		if oldCount == 0 && newCount > 0 {
			for j := 0; j < newCount; j++ {
				idx := newStart - 1 + j
				if idx >= 0 && idx < totalLines {
					status[idx] = StatusAdded
				}
			}
		} else if newCount == 0 && oldCount > 0 {
			idx := newStart - 1
			if idx >= 0 && idx < totalLines {
				status[idx] = StatusDeleted
			}
		} else {
			for j := 0; j < newCount; j++ {
				idx := newStart - 1 + j
				if idx >= 0 && idx < totalLines {
					status[idx] = StatusModified
				}
			}
		}
	}

	return status
}

func parseHunkHeader(line string) (int, int, int) {
	idx := strings.Index(line, "@@")
	if idx < 0 {
		return 0, 0, 0
	}
	rest := line[idx+2:]
	idx2 := strings.Index(rest, "@@")
	if idx2 < 0 {
		return 0, 0, 0
	}
	header := strings.TrimSpace(rest[:idx2])

	parts := strings.Fields(header)
	if len(parts) < 2 {
		return 0, 0, 0
	}

	_, oldCount := parseRange(parts[0])
	newStart, newCount := parseRange(parts[1])

	return newStart, newCount, oldCount
}

func parseRange(s string) (int, int) {
	if len(s) < 2 {
		return 0, 0
	}
	s = s[1:]
	if idx := strings.Index(s, ","); idx >= 0 {
		start, _ := strconv.Atoi(s[:idx])
		count, _ := strconv.Atoi(s[idx+1:])
		return start, count
	}
	start, _ := strconv.Atoi(s)
	return start, 1
}

// RepoFileStatus maps relative file paths to their git status.
type RepoFileStatus map[string]FileGitStatus

// GetRepoStatus runs `git status --porcelain -u` in the given directory
// and returns a map from repo-relative path to status.
// Also propagates status to parent directories so folders show as modified.
func GetRepoStatus(rootDir string) RepoFileStatus {
	if !gitAvailable() {
		return nil
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil
	}

	root, ok := repoRoot(absRoot)
	if !ok {
		return nil
	}

	cmd := exec.Command("git", "status", "--porcelain", "-u", "--no-renames")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	result := make(RepoFileStatus)
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")

	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		xy := line[:2]
		relPath := strings.TrimSpace(line[3:])
		if relPath == "" {
			continue
		}
		relPath = filepath.FromSlash(relPath)

		absPath := filepath.Join(root, relPath)

		rel, err := filepath.Rel(absRoot, absPath)
		if err != nil {
			continue
		}
		// Skip files outside rootDir
		if strings.HasPrefix(rel, "..") {
			continue
		}

		st := classifyStatus(xy)
		result[rel] = st

		propagateToParents(result, rel, st)
	}

	addIgnoredPaths(result, root, absRoot)

	return result
}

// addIgnoredPaths runs `git ls-files --others --ignored --exclude-standard --directory`
// to find gitignored files/directories and marks them in the result map.
func addIgnoredPaths(result RepoFileStatus, repoRoot, absRoot string) {
	cmd := exec.Command("git", "ls-files", "--others", "--ignored", "--exclude-standard", "--directory")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		relPath := filepath.FromSlash(strings.TrimSuffix(line, "/"))
		absPath := filepath.Join(repoRoot, relPath)

		rel, err := filepath.Rel(absRoot, absPath)
		if err != nil {
			continue
		}
		if strings.HasPrefix(rel, "..") {
			continue
		}

		if _, exists := result[rel]; !exists {
			result[rel] = FileIgnored
		}
	}
}

// StatusForPath looks up the status for an absolute path given the root directory.
func StatusForPath(statuses RepoFileStatus, rootDir, absPath string) FileGitStatus {
	if statuses == nil {
		return FileClean
	}
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		return FileClean
	}
	if s, ok := statuses[rel]; ok {
		return s
	}
	return FileClean
}

func classifyStatus(xy string) FileGitStatus {
	x := xy[0]
	y := xy[1]

	if x == '?' && y == '?' {
		return FileUntracked
	}
	if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
		return FileConflicted
	}
	if x == 'A' || y == 'A' {
		return FileAdded
	}
	if x == 'D' || y == 'D' {
		return FileDeleted
	}
	if x == 'R' || y == 'R' {
		return FileRenamed
	}
	if x == 'M' || y == 'M' {
		return FileModified
	}
	return FileClean
}

func propagateToParents(result RepoFileStatus, relPath string, st FileGitStatus) {
	dir := filepath.Dir(relPath)
	for dir != "." && dir != "" {
		if existing, ok := result[dir]; ok {
			if statusPriority(st) <= statusPriority(existing) {
				break
			}
		}
		result[dir] = st
		dir = filepath.Dir(dir)
	}
}

func statusPriority(s FileGitStatus) int {
	switch s {
	case FileConflicted:
		return 5
	case FileDeleted:
		return 4
	case FileUntracked:
		return 3
	case FileModified:
		return 2
	case FileAdded:
		return 1
	default:
		return 0
	}
}

// StatusLabel returns a short colored label for the status.
func StatusLabel(s FileGitStatus) string {
	switch s {
	case FileModified:
		return "M"
	case FileAdded:
		return "A"
	case FileUntracked:
		return "U"
	case FileDeleted:
		return "D"
	case FileRenamed:
		return "R"
	case FileConflicted:
		return "C"
	default:
		return ""
	}
}

// IsInGitRepo checks if the given directory is inside a git repo.
func IsInGitRepo(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	_, ok := repoRoot(dir)
	return ok
}
