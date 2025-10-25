package app

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// handleTasksMessages handles messages specific to task loading
func (m *Model) handleTasksMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TasksLoadedMsg:
		// Tasks tab
		m.tasks = msg.Tasks
		m.taskState = TaskLoaded
		m.tasksIndex = 0
		m.tasksOffset = 0
		m.taskError = nil
		m.statusMsg = fmt.Sprintf("Loaded %d task(s)", len(msg.Tasks))
		log.Printf("[TASKS] Successfully loaded %d tasks", len(msg.Tasks))
		// Invalidate overview cache when tasks change
		m.overviewDirty = true
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
	if msg.String() == "q" || msg.String() == "esc" {
		// Go back to detail view
		m.currentView = ViewDetail
		m.tasksIndex = 0
		m.tasksOffset = 0
		return m, nil
	}

	// Use resolver for navigation actions
	if m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		if found {
			switch action {
			case "down":
				// Move down in list
				if m.tasksIndex < len(m.tasks)-1 {
					m.tasksIndex++
					m.adjustTasksScroll()
					m.rightPaneOffset = 0 // Reset detail scroll
				}
				return m, nil

			case "up":
				// Move up in list
				if m.tasksIndex > 0 {
					m.tasksIndex--
					m.adjustTasksScroll()
					m.rightPaneOffset = 0 // Reset detail scroll
				}
				return m, nil

			case "refresh":
				// Reload tasks asynchronously
				m.taskState = TaskLoading
				m.statusMsg = "Loading tasks..."
				return m, m.loadTasksCmd()

			case "page_down":
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

			case "page_up":
				// Page up
				m.tasksIndex -= 10
				if m.tasksIndex < 0 {
					m.tasksIndex = 0
				}
				m.adjustTasksScroll()
				return m, nil
			}
		}
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
		// Mark task done via the data store
		if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
			task := m.tasks[m.tasksIndex]
			err := m.data.Tasks.MarkDone(context.Background(), task.ID)
			if err != nil {
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
