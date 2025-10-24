package app

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Claude tab messages

// claudeFilesLoadedMsg is sent when Claude config files are loaded
type claudeFilesLoadedMsg struct {
	files []ClaudeFile
	err   error
}

// claudeFileContentLoadedMsg is sent when a file's content is loaded
type claudeFileContentLoadedMsg struct {
	content string
	path    string
	err     error
}

// claudeEditorClosedMsg is sent when external editor closes
type claudeEditorClosedMsg struct {
	err error
}

// claudeValidationResultMsg is sent when JSON validation completes
type claudeValidationResultMsg struct {
	valid  bool
	status string
}

// Claude tab commands

// loadClaudeFilesCmd returns a command to load Claude config files
func (m *Model) loadClaudeFilesCmd() tea.Cmd {
	return func() tea.Msg {
		files, err := scanClaudeDir(m.claudeCurrentPath)
		return claudeFilesLoadedMsg{
			files: files,
			err:   err,
		}
	}
}

// loadClaudeFileContentCmd returns a command to load a file's content
func loadClaudeFileContentCmd(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := readClaudeFile(path)
		if err != nil {
			return claudeFileContentLoadedMsg{
				path: path,
				err:  err,
			}
		}

		// Try to prettify if JSON
		prettified, prettifyErr := prettifyJSON(content)
		if prettifyErr == nil {
			content = prettified
		}

		return claudeFileContentLoadedMsg{
			content: content,
			path:    path,
			err:     nil,
		}
	}
}

// validateClaudeFileCmd returns a command to validate JSON content
func validateClaudeFileCmd(content string) tea.Cmd {
	return func() tea.Msg {
		valid, status := validateJSON(content)
		return claudeValidationResultMsg{
			valid:  valid,
			status: status,
		}
	}
}

// loadClaudeDirectoryContentCmd returns a command to load directory contents
func (m *Model) loadClaudeDirectoryContentCmd(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := os.ReadDir(path)
		if err != nil {
			return claudeFileContentLoadedMsg{
				path: path,
				err:  err,
			}
		}

		var lines []string
		for _, file := range files {
			icon := "📄"
			if file.IsDir() {
				icon = "📁"
			}
			lines = append(lines, icon+" "+file.Name())
		}

		content := "📁 Directory contents:\n\n" + strings.Join(lines, "\n")
		return claudeFileContentLoadedMsg{
			content: content,
			path:    path,
			err:     nil,
		}
	}
}

// openClaudeFileInEditor opens a file in the user's $EDITOR
func openClaudeFileInEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim" // Fallback to nvim
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return claudeEditorClosedMsg{err: err}
	})
}
