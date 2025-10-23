package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/keybindings"
)

// Keybindings tab messages

// keybindingsSavedMsg is sent when keybindings are saved
type keybindingsSavedMsg struct {
	err error
}

// keybindingsReloadedMsg is sent when keybindings are reloaded
type keybindingsReloadedMsg struct {
	content string
	err     error
}

// Keybindings tab commands

// saveKeybindingsCmd returns a command to save keybindings
func (m *Model) saveKeybindingsCmd(content string) tea.Cmd {
	return func() tea.Msg {
		err := keybindings.SaveKeybindingsYAML(content)
		if err != nil {
			return keybindingsSavedMsg{err: err}
		}
		// Reload the keybindings in the resolver after saving
		if resolver, err := keybindings.LoadKeybindings(); err == nil {
			m.keybindResolver = resolver
		}
		return keybindingsSavedMsg{err: nil}
	}
}

// reloadKeybindingsCmd returns a command to reload keybindings from file
func (m *Model) reloadKeybindingsCmd() tea.Cmd {
	return func() tea.Msg {
		content, err := keybindings.LoadKeybindingsYAML()
		if err != nil {
			return keybindingsReloadedMsg{
				content: "",
				err:     err,
			}
		}

		// Also reload the resolver
		if resolver, err := keybindings.LoadKeybindings(); err == nil {
			m.keybindResolver = resolver
		}

		return keybindingsReloadedMsg{
			content: content,
			err:     nil,
		}
	}
}
