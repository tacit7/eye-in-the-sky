package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleLogsKeys handles keys in logs view
func (m *Model) handleLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In logs view, q/esc go back to detail instead of quitting
	if msg.String() == "q" || msg.String() == "esc" {
		// Go back to detail view
		m.currentView = ViewDetail
		return m, nil
	}

	// Use resolver for navigation actions
	if m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		if found {
			switch action {
			case "down":
				// Move down in list
				if m.logsIndex < len(m.logs)-1 {
					m.logsIndex++
					m.rightPaneOffset = 0
				}
				return m, nil

			case "up":
				// Move up in list
				if m.logsIndex > 0 {
					m.logsIndex--
					m.rightPaneOffset = 0
				}
				return m, nil

			case "page_down":
				// Page down
				m.rightPaneOffset += 10
				return m, nil

			case "page_up":
				// Page up
				m.rightPaneOffset -= 10
				if m.rightPaneOffset < 0 {
					m.rightPaneOffset = 0
				}
				return m, nil
			}
		}
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.logsIndex = 0
		m.rightPaneOffset = 0
	case "G":
		// Go to bottom
		if len(m.logs) > 0 {
			m.logsIndex = len(m.logs) - 1
			m.rightPaneOffset = 0
		}
	}

	return m, nil
}
