package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileInfo contains information about a discovered JSONL file
type FileInfo struct {
	Path    string // Absolute path to the file
	Project string // Project name extracted from path
	MTime   int64  // File modification time
}

// DiscoverFiles discovers all JSONL files from Claude data directories
func DiscoverFiles() ([]FileInfo, error) {
	var files []FileInfo
	paths := getClaudePaths()

	for _, path := range paths {
		discovered, err := walkClaudeDirectory(path)
		if err != nil {
			// Log but continue if one directory fails
			continue
		}
		files = append(files, discovered...)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no JSONL files found in Claude data directories")
	}

	return files, nil
}

// getClaudePaths returns the Claude data directory paths to search
func getClaudePaths() []string {
	// Check CLAUDE_CONFIG_DIR environment variable
	if env := os.Getenv("CLAUDE_CONFIG_DIR"); env != "" {
		// Support comma-separated paths
		var paths []string
		for _, p := range strings.Split(env, ",") {
			if p = strings.TrimSpace(p); p != "" {
				paths = append(paths, p)
			}
		}
		return paths
	}

	// Default paths
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	return []string{
		filepath.Join(home, ".config", "claude", "projects"),
		filepath.Join(home, ".claude", "projects"),
	}
}

// walkClaudeDirectory walks a Claude data directory and discovers JSONL files
func walkClaudeDirectory(rootPath string) ([]FileInfo, error) {
	var files []FileInfo

	// Check if directory exists
	info, err := os.Stat(rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Directory doesn't exist, skip
		}
		return nil, fmt.Errorf("failed to stat directory %s: %w", rootPath, err)
	}

	if !info.IsDir() {
		return nil, nil
	}

	// Walk the directory structure: {rootPath}/{project}/{sessionId}.jsonl
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", rootPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectName := entry.Name()
		projectPath := filepath.Join(rootPath, projectName)

		// Look for JSONL files in the project directory
		projectEntries, err := os.ReadDir(projectPath)
		if err != nil {
			continue // Skip on error
		}

		for _, projectEntry := range projectEntries {
			if projectEntry.IsDir() {
				continue
			}

			if !strings.HasSuffix(projectEntry.Name(), ".jsonl") {
				continue
			}

			filePath := filepath.Join(projectPath, projectEntry.Name())
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				continue
			}

			files = append(files, FileInfo{
				Path:    filePath,
				Project: projectName,
				MTime:   fileInfo.ModTime().Unix(),
			})
		}
	}

	return files, nil
}
