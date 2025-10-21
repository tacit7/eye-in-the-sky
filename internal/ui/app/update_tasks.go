package app

import (
	"fmt"
	"log"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// handleTasksMessages handles messages specific to task loading
func (m *Model) handleTasksMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TasksLoadedMsg:
		m.tasks = msg.Tasks
		m.taskState = TaskLoaded
		m.tasksIndex = 0
		m.tasksOffset = 0
		m.taskError = nil
		m.statusMsg = fmt.Sprintf("Loaded %d task(s)", len(msg.Tasks))
		log.Printf("[TASKS] Successfully loaded %d tasks", len(msg.Tasks))
		return m, nil

	case TasksErrorMsg:
		m.taskState = TaskError
		m.taskError = msg.Error
		m.statusMsg = fmt.Sprintf("Error loading tasks: %v", msg.Error)
		log.Printf("[TASKS] Error loading tasks: %v", msg.Error)
		return m, nil
	}
	return m, nil
}

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
		// Reload tasks asynchronously
		m.taskState = TaskLoading
		m.statusMsg = "Loading tasks..."
		return m, m.loadTasksCmd()
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
		// Mark task done and reload asynchronously
		if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
			task := m.tasks[m.tasksIndex]
			cmd := exec.Command("task", task.UUID, "done")
			if err := cmd.Run(); err != nil {
				m.err = err
				m.statusMsg = fmt.Sprintf("Failed to mark task done: %v", err)
			} else {
				m.statusMsg = "Task marked done, reloading..."
				// Trigger async reload
				m.taskState = TaskLoading
				return m, m.loadTasksCmd()
			}
		}
	}

	return m, nil
}
