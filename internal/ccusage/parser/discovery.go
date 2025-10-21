package parser

import (
	"fmt"
	"log"
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
	log.Printf("[DISCOVER] Searching in %d paths", len(paths))

	for _, path := range paths {
		log.Printf("[DISCOVER] Searching path: %s", path)
		discovered, err := walkClaudeDirectory(path)
		if err != nil {
			log.Printf("[DISCOVER] Error in %s: %v", path, err)
			// Log but continue if one directory fails
			continue
		}
		log.Printf("[DISCOVER] Found %d files in %s", len(discovered), path)
		files = append(files, discovered...)
	}

	log.Printf("[DISCOVER] Total files discovered: %d", len(files))
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
		log.Printf("[DISCOVER] Using CLAUDE_CONFIG_DIR: %v", paths)
		return paths
	}

	// Default paths
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("[DISCOVER] Error getting home directory: %v", err)
		return []string{}
	}

	paths := []string{
		filepath.Join(home, ".config", "claude", "projects"),
		filepath.Join(home, ".claude", "projects"),
	}
	log.Printf("[DISCOVER] Using default paths: %v", paths)
	return paths
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
