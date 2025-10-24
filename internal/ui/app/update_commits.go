package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleCommitsKeys handles keys in commits view
func (m *Model) handleCommitsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// In commits view, q/esc go back to detail instead of quitting
	if keyStr == "q" || keyStr == "esc" {
		// Go back to detail view
		m.currentView = ViewDetail
		m.commitsIndex = 0
		m.commitsOffset = 0
		return m, nil
	}

	if keyStr == "j" || keyStr == "down" {
		// Move down in list
		if m.commitsIndex < len(m.commits)-1 {
			m.commitsIndex++
			m.adjustCommitsScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if keyStr == "k" || keyStr == "up" {
		// Move up in list
		if m.commitsIndex > 0 {
			m.commitsIndex--
			m.adjustCommitsScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if keyStr == "l" {
		// Scroll detail pane down
		m.rightPaneOffset++
		return m, nil
	}

	if keyStr == "h" {
		// Scroll detail pane up
		if m.rightPaneOffset > 0 {
			m.rightPaneOffset--
		}
		return m, nil
	}

	if keyStr == "r" || keyStr == "R" {
		// Reload commits using command
		if m.selectedAgent != nil {
			m.statusMsg = "Refreshing commits..."
			return m, loadAgentDetailsCmd(m.data, m.selectedAgent.ID)
		}
		return m, nil
	}

	if keyStr == "pgdown" {
		// Page down
		m.commitsIndex += 10
		if m.commitsIndex >= len(m.commits) {
			m.commitsIndex = len(m.commits) - 1
		}
		if m.commitsIndex < 0 {
			m.commitsIndex = 0
		}
		m.adjustCommitsScroll()
		return m, nil
	}

	if keyStr == "pgup" {
		// Page up
		m.commitsIndex -= 10
		if m.commitsIndex < 0 {
			m.commitsIndex = 0
		}
		m.adjustCommitsScroll()
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.commitsIndex = 0
		m.commitsOffset = 0
	case "G":
		// Go to bottom
		if len(m.commits) > 0 {
			m.commitsIndex = len(m.commits) - 1
			m.adjustCommitsScroll()
		}
	}

	return m, nil
}
