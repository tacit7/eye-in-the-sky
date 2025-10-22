package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleNotesKeys handles keys in notes view
func (m *Model) handleNotesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In notes view, q/esc go back to detail instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to detail view
		m.currentView = ViewDetail
		m.notesIndex = 0
		m.notesOffset = 0
		return m, nil
	}

	if Matches(msg, m.keys.NavigateDown) {
		// Move down in list
		if m.notesIndex < len(m.notes)-1 {
			m.notesIndex++
			m.adjustNotesScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.NavigateUp) {
		// Move up in list
		if m.notesIndex > 0 {
			m.notesIndex--
			m.adjustNotesScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.ScrollDown) {
		// Scroll detail pane down
		m.rightPaneOffset++
		return m, nil
	}

	if Matches(msg, m.keys.ScrollUp) {
		// Scroll detail pane up
		if m.rightPaneOffset > 0 {
			m.rightPaneOffset--
		}
		return m, nil
	}

	if Matches(msg, m.keys.Refresh) {
		// Reload notes using command
		if m.selectedAgent != nil {
			m.statusMsg = "Refreshing notes..."
			return m, loadAgentDetailsCmd(m.data, m.selectedAgent.ID)
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
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
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.notesIndex -= 10
		if m.notesIndex < 0 {
			m.notesIndex = 0
		}
		m.adjustNotesScroll()
		return m, nil
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
