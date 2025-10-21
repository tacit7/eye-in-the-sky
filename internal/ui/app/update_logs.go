package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleLogsKeys handles keys in logs view
func (m *Model) handleLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In logs view, q/esc go back to list instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to list view
		m.currentView = ViewList
		return m, nil
	}

	if Matches(msg, m.keys.ScrollDown) {
		// Scroll down
		m.detailOffset++
		return m, nil
	}

	if Matches(msg, m.keys.ScrollUp) {
		// Scroll up
		if m.detailOffset > 0 {
			m.detailOffset--
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
		// Page down
		m.detailOffset += 10
		return m, nil
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.detailOffset -= 10
		if m.detailOffset < 0 {
			m.detailOffset = 0
		}
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.detailOffset = 0
	case "G":
		// Go to bottom
		m.detailOffset = 9999
	}

	return m, nil
}
