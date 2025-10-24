package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleDetailKeys handles keys in detail view
func (m *Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Use keybindings resolver for current tab context
	if m.keybindResolver != nil {
		currentTab := getCurrentDetailTabName(m.tabs.ActiveIndex)
		m.keybindResolver.SetContext("detail", currentTab)
	}

	// In detail view, q/esc go back to list instead of quitting
	if msg.String() == "q" || msg.String() == "esc" {
		// Go back to list view, preserving selection
		m.currentView = ViewList
		m.detailOffset = 0
		return m, nil
	}

	// Tab navigation (left/right arrow keys or tab/shift-tab)
	switch msg.String() {
	case "tab", "right":
		m.tabs.Next()
		// If user navigated to index 0 (back arrow), go back to list
		if m.tabs.ActiveIndex == 0 {
			m.currentView = ViewList
			m.detailOffset = 0
			return m, nil
		}
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "shift+tab", "left":
		m.tabs.Prev()
		// If user navigated to index 0 (back arrow), go back to list
		if m.tabs.ActiveIndex == 0 {
			m.currentView = ViewList
			m.detailOffset = 0
			return m, nil
		}
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "A":
		m.tabs.Set(1) // Agent View tab (index 1 now, 0 is back arrow)
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "C":
		m.tabs.Set(2) // Commits tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "L":
		m.tabs.Set(3) // Logs tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "N":
		m.tabs.Set(4) // Notes tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "T":
		m.tabs.Set(6) // Tasks tab
		// Trigger async task loading
		m.taskState = TaskLoading
		m.statusMsg = "Loading tasks..."
		return m, m.loadTasksCmd()
	case "P":
		m.tabs.Set(7) // Projects tab
		// Trigger async project tickets loading
		m.statusMsg = "Loading project tickets..."
		return m, m.loadProjectTicketsCmd()
	}

	// View navigation keys (only from Detail view)
	switch msg.String() {
	}

	// j/k navigation for split-pane tabs
	switch msg.String() {
	case "j", "down":
		// Navigate down in the current tab's list
		switch m.tabs.ActiveIndex {
		case 2: // Commits
			if m.commitsIndex < len(m.commits)-1 {
				m.commitsIndex++
				m.rightPaneOffset = 0 // Reset right pane scroll
			}
		case 3: // Logs
			if m.logsIndex < len(m.logs)-1 {
				m.logsIndex++
				m.rightPaneOffset = 0
			}
		case 4: // Notes
			if m.notesIndex < len(m.notes)-1 {
				m.notesIndex++
				m.rightPaneOffset = 0
			}
		case 5: // Actions
			if m.actionsIndex < len(m.actions)-1 {
				m.actionsIndex++
				m.rightPaneOffset = 0
			}
		case 6: // Tasks
			if m.tasksIndex < len(m.tasks)-1 {
				m.tasksIndex++
				m.rightPaneOffset = 0
			}
		case 7: // Projects
			if m.projectTicketsIndex < len(m.projectTickets)-1 {
				m.projectTicketsIndex++
				m.rightPaneOffset = 0
			}
		default:
			// Overview tab - use default scrolling
			m.detailOffset++
		}
		return m, nil

	case "k", "up":
		// Navigate up in the current tab's list
		switch m.tabs.ActiveIndex {
		case 2: // Commits
			if m.commitsIndex > 0 {
				m.commitsIndex--
				m.rightPaneOffset = 0
			}
		case 3: // Logs
			if m.logsIndex > 0 {
				m.logsIndex--
				m.rightPaneOffset = 0
			}
		case 4: // Notes
			if m.notesIndex > 0 {
				m.notesIndex--
				m.rightPaneOffset = 0
			}
		case 5: // Actions
			if m.actionsIndex > 0 {
				m.actionsIndex--
				m.rightPaneOffset = 0
			}
		case 6: // Tasks
			if m.tasksIndex > 0 {
				m.tasksIndex--
				m.rightPaneOffset = 0
			}
		case 7: // Projects
			if m.projectTicketsIndex > 0 {
				m.projectTicketsIndex--
				m.rightPaneOffset = 0
			}
		default:
			// Overview tab - use default scrolling
			if m.detailOffset > 0 {
				m.detailOffset--
			}
		}
		return m, nil

	case "h":
		// Scroll right pane up (only for split-pane tabs)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 7 {
			if m.rightPaneOffset > 0 {
				m.rightPaneOffset--
			}
			return m, nil
		}

	case "l":
		// Scroll right pane down (only for split-pane tabs)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 7 {
			m.rightPaneOffset++
			return m, nil
		}

	case "ctrl+u":
		// Page up in right pane (or overview)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 7 {
			// Split-pane tabs: page up right pane
			m.rightPaneOffset -= 10
			if m.rightPaneOffset < 0 {
				m.rightPaneOffset = 0
			}
		} else {
			// Overview tab: page up content
			m.detailOffset -= 10
			if m.detailOffset < 0 {
				m.detailOffset = 0
			}
		}
		return m, nil

	case "ctrl+d":
		// Page down in right pane (or overview)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 7 {
			// Split-pane tabs: page down right pane
			m.rightPaneOffset += 10
		} else {
			// Overview tab: page down content
			m.detailOffset += 10
		}
		return m, nil
	}

	// Note: ScrollDown/ScrollUp are already handled above via j/k switch cases

	if msg.String() == "pgdown" {
		// Page down
		m.detailOffset += 10
		return m, nil
	}

	if msg.String() == "pgup" {
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

// getCurrentDetailTabName returns the keybindings scope name for the current detail tab
func getCurrentDetailTabName(tabIndex int) string {
	switch tabIndex {
	case 0:
		return "back"
	case 1:
		return "overview"
	case 2:
		return "commits"
	case 3:
		return "logs"
	case 4:
		return "notes"
	case 5:
		return "actions"
	case 6:
		return "tasks"
	case 7:
		return "projects"
	default:
		return "overview"
	}
}
