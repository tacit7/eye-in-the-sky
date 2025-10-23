package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// scanClaudeDir scans ~/.claude directory for config files
func scanClaudeDir() ([]ClaudeFile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")

	// Check if directory exists
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return []ClaudeFile{}, nil
	}

	var files []ClaudeFile

	// Scan top-level files and immediate subdirectories
	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read .claude directory: %w", err)
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
				Name:    entry.Name(),
				Path:    fullPath,
				IsDir:   entry.IsDir(),
				ModTime: info.ModTime(),
			})
		}
	}

	// Sort: directories first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
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
