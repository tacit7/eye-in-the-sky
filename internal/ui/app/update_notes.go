package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleNotesKeys handles keys in notes view
func (m *Model) handleNotesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In notes view, q/esc go back to detail instead of quitting
	if msg.String() == "q" || msg.String() == "esc" {
		// Go back to detail view
		m.currentView = ViewDetail
		m.notesIndex = 0
		m.notesOffset = 0
		return m, nil
	}

	// Use resolver for navigation actions
	if m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		if found {
			switch action {
			case "down":
				// Move down in list
				if m.notesIndex < len(m.notes)-1 {
					m.notesIndex++
					m.adjustNotesScroll()
					m.rightPaneOffset = 0 // Reset detail scroll
				}
				return m, nil

			case "up":
				// Move up in list
				if m.notesIndex > 0 {
					m.notesIndex--
					m.adjustNotesScroll()
					m.rightPaneOffset = 0 // Reset detail scroll
				}
				return m, nil

			case "refresh":
				// Reload notes using command
				if m.selectedAgent != nil {
					m.statusMsg = "Refreshing notes..."
					return m, loadAgentDetailsCmd(m.data, m.selectedAgent.ID)
				}
				return m, nil

			case "page_down":
				// Page down
				m.notesIndex += 10
				if m.notesIndex >= len(m.notes) {
					m.notesIndex = len(m.notes) - 1
				}
				if m.notesIndex < 0 {
					m.notesIndex = 0
				}
				m.adjustNotesScroll()
				return m, nil

			case "page_up":
				// Page up
				m.notesIndex -= 10
				if m.notesIndex < 0 {
					m.notesIndex = 0
				}
				m.adjustNotesScroll()
				return m, nil
			}
		}
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.notesIndex = 0
		m.notesOffset = 0
	case "G":
		// Go to bottom
		if len(m.notes) > 0 {
			m.notesIndex = len(m.notes) - 1
			m.adjustNotesScroll()
		}
	}

	return m, nil
}
