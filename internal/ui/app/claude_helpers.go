package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// scanClaudeDir scans a directory for config files
// pathSuffix is the subdirectory within ~/.claude (e.g., "", "hooks", "agents")
func scanClaudeDir(pathSuffix string) ([]ClaudeFile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude", pathSuffix)

	// Check if directory exists
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return []ClaudeFile{}, nil
	}

	var files []ClaudeFile

	// Add parent directory (..) if not at root
	if pathSuffix != "" {
		parentPath := filepath.Join(homeDir, ".claude")
		parentSuffix := filepath.Dir(pathSuffix)
		if parentSuffix == "." {
			parentSuffix = ""
		}
		files = append(files, ClaudeFile{
			Name:     "..",
			Path:     parentPath,
			IsDir:    true,
			IsParent: true,
			ModTime:  time.Now(),
		})
	}

	// Scan top-level files and immediate subdirectories
	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(claudeDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Include .json, .md files and directories like hooks/, agents/
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".json") || strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, ClaudeFile{
				Name:     entry.Name(),
				Path:     fullPath,
				IsDir:    entry.IsDir(),
				IsParent: false,
				ModTime:  info.ModTime(),
			})
		}
	}

	// Sort: parent at top, then files first, then directories
	sort.Slice(files, func(i, j int) bool {
		// Keep parent directory at top
		if files[i].IsParent {
			return true
		}
		if files[j].IsParent {
			return false
		}

		// Files come before directories
		if files[i].IsDir != files[j].IsDir {
			return !files[i].IsDir // files first (IsDir=false before IsDir=true)
		}

		// Within files/dirs, special handling for CLAUDE file
		if !files[i].IsDir && !files[j].IsDir {
			if files[i].Name == "CLAUDE.md" {
				return true
			}
			if files[j].Name == "CLAUDE.md" {
				return false
			}
		}

		// Alphabetical sort
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// readClaudeFile reads the content of a file
func readClaudeFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

// prettifyJSON formats JSON with indentation
func prettifyJSON(content string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// Not valid JSON, return as-is
		return content, err
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return content, err
	}

	return buf.String(), nil
}

// validateJSON checks if content is valid JSON
func validateJSON(content string) (bool, string) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return false, fmt.Sprintf("Invalid JSON: %v", err)
	}
	return true, "✓ Valid JSON"
}

// colorizeJSON applies syntax highlighting to JSON content
func colorizeJSON(content string) (string, error) {
	// Get JSON lexer
	lexer := lexers.Get("json")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use terminal-friendly style
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Use terminal formatter with 256-color support
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize and format
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content, err
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return content, err
	}

	return buf.String(), nil
}
