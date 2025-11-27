package app

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// noteEditorClosedMsg is sent when the note editor closes
type noteEditorClosedMsg struct {
	tempPath string
	err      error
}

// openNoteEditorCmd opens the user's EDITOR to create a note
func openNoteEditorCmd() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Fallback to vim
	}

	// Create temporary file for note
	tmpFile, err := os.CreateTemp("", "eits-note-*.md")
	if err != nil {
		return func() tea.Msg {
			return noteEditorClosedMsg{err: err}
		}
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(execErr error) tea.Msg {
		return noteEditorClosedMsg{
			tempPath: tmpPath,
			err:      execErr,
		}
	})
}
