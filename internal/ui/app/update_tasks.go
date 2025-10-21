package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// handleTasksKeys handles keys in tasks view
func (m *Model) handleTasksKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In tasks view, q/esc go back to detail instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to detail view
		m.currentView = ViewDetail
		m.tasksIndex = 0
		m.tasksOffset = 0
		return m, nil
	}

	if Matches(msg, m.keys.NavigateDown) {
		// Move down in list
		if m.tasksIndex < len(m.tasks)-1 {
			m.tasksIndex++
			m.adjustTasksScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.NavigateUp) {
		// Move up in list
		if m.tasksIndex > 0 {
			m.tasksIndex--
			m.adjustTasksScroll()
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
		// Reload tasks
		if err := m.loadTasks(); err != nil {
			m.err = err
			m.statusMsg = fmt.Sprintf("Failed to reload tasks: %v", err)
		} else {
			m.statusMsg = "Tasks refreshed"
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
		// Page down
		m.tasksIndex += 10
		if m.tasksIndex >= len(m.tasks) {
			m.tasksIndex = len(m.tasks) - 1
		}
		if m.tasksIndex < 0 {
			m.tasksIndex = 0
		}
		m.adjustTasksScroll()
		return m, nil
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.tasksIndex -= 10
		if m.tasksIndex < 0 {
			m.tasksIndex = 0
		}
		m.adjustTasksScroll()
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.tasksIndex = 0
		m.tasksOffset = 0
	case "G":
		// Go to bottom
		if len(m.tasks) > 0 {
			m.tasksIndex = len(m.tasks) - 1
			m.adjustTasksScroll()
		}
	case "d":
		// Mark task done
		if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
			if err := m.markTaskDone(); err != nil {
				m.err = err
				m.statusMsg = fmt.Sprintf("Failed to mark task done: %v", err)
			} else {
				m.statusMsg = "Task marked done"
			}
		}
	}

	return m, nil
}
