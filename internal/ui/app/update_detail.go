package app

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// handleDetailKeys handles keys in detail view
func (m *Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Printf("[DETAIL] handleDetailKeys: received key='%s'", msg.String())

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
	log.Printf("[DETAIL] TAB Navigation Message: key='%s'", msg.String())
	switch msg.String() {
	case "tab", "right":
		// Navigate forward, skipping tab 0 (back arrow)
		// If on tab 6, go to tab 1; otherwise increment
		if m.tabs.ActiveIndex == 6 {
			m.tabs.Set(1)
		} else {
			m.tabs.Next()
			// Skip tab 0 if we landed on it
			if m.tabs.ActiveIndex == 0 {
				m.tabs.Set(1)
			}
		}
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "shift+tab", "left":
		// Navigate backward, skipping tab 0 (back arrow)
		// If on tab 1, go to tab 6; otherwise decrement
		if m.tabs.ActiveIndex == 1 {
			m.tabs.Set(6)
		} else {
			m.tabs.Prev()
			// Skip tab 0 if we landed on it
			if m.tabs.ActiveIndex == 0 {
				m.tabs.Set(6)
			}
		}
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "O":
		log.Printf("[DETAIL] KEY O PRESSED: Switching to Overview tab (1)")
		m.tabs.Set(1) // Overview tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "T":
		log.Printf("[DETAIL] KEY T PRESSED: Switching to Tasks tab (2)")
		m.tabs.Set(2) // Tasks tab
		// Trigger async task loading
		m.taskState = TaskLoading
		m.statusMsg = "Loading tasks..."
		return m, m.loadTasksCmd()
	case "L":
		log.Printf("[DETAIL] KEY L PRESSED: Switching to Logs tab (3)")
		m.tabs.Set(3) // Logs tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "C":
		log.Printf("[DETAIL] KEY C PRESSED: Switching to Commits tab (4)")
		m.tabs.Set(4) // Commits tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "N":
		log.Printf("[DETAIL] KEY N PRESSED: Switching to Notes tab (5)")
		m.tabs.Set(5) // Notes tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "S":
		log.Printf("[DETAIL] KEY S PRESSED: Switching to Session Context tab (6)")
		m.tabs.Set(6) // Session Context tab
		if err := m.loadTabData(); err != nil {
			m.err = err
			log.Printf("[DETAIL] ERROR loading session context: %v", err)
		}
		log.Printf("[DETAIL] Tab set to: %d", m.tabs.ActiveIndex)
		return m, nil
	case "n":
		// Create new note for current agent
		if m.selectedAgent != nil {
			// Get project ID from project name
			projectID := ""
			if m.selectedAgent.ProjectName != "" {
				if proj, err := m.data.DB.GetProjectByName(m.selectedAgent.ProjectName); err == nil && proj != nil {
					projectID = proj.ID
				}
			}

			m.noteModal.SetContext(string(m.selectedAgent.ID), m.selectedAgent.SessionID, projectID)
			m.noteModal.Show()
		}
		return m, nil
	}

	// Navigation using resolver
	if m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		if found {
			switch action {
			case "down":
				// Navigate down in the current tab's list
				// m.tabs.ActiveIndex includes the Back button at index 0
				switch m.tabs.ActiveIndex {
				case 1: // Overview
					m.detailOffset++
				case 2: // Tasks
					if m.tasksIndex < len(m.tasks)-1 {
						m.tasksIndex++
						m.rightPaneOffset = 0
					}
				case 3: // Logs
					if m.logsIndex < len(m.logs)-1 {
						m.logsIndex++
						m.rightPaneOffset = 0
					}
				case 4: // Commits
					if m.commitsIndex < len(m.commits)-1 {
						m.commitsIndex++
						m.rightPaneOffset = 0
					}
				case 5: // Notes
					if m.notesIndex < len(m.notes)-1 {
						m.notesIndex++
						m.rightPaneOffset = 0
					}
				}
				return m, nil

			case "up":
				// Navigate up in the current tab's list
				// m.tabs.ActiveIndex includes the Back button at index 0
				switch m.tabs.ActiveIndex {
				case 1: // Overview
					if m.detailOffset > 0 {
						m.detailOffset--
					}
				case 2: // Tasks
					if m.tasksIndex > 0 {
						m.tasksIndex--
						m.rightPaneOffset = 0
					}
				case 3: // Logs
					if m.logsIndex > 0 {
						m.logsIndex--
						m.rightPaneOffset = 0
					}
				case 4: // Commits
					if m.commitsIndex > 0 {
						m.commitsIndex--
						m.rightPaneOffset = 0
					}
				case 5: // Notes
					if m.notesIndex > 0 {
						m.notesIndex--
						m.rightPaneOffset = 0
					}
				}
				return m, nil

			case "page_down":
				// Page down in right pane (or overview)
				if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 5 {
					m.rightPaneOffset += 10
				} else {
					m.detailOffset += 10
				}
				return m, nil

			case "page_up":
				// Page up in right pane (or overview)
				if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 5 {
					m.rightPaneOffset -= 10
					if m.rightPaneOffset < 0 {
						m.rightPaneOffset = 0
					}
				} else {
					m.detailOffset -= 10
					if m.detailOffset < 0 {
						m.detailOffset = 0
					}
				}
				return m, nil
			}
		}
	}

	// Task-specific commands (only when on Tasks tab)
	if m.tabs.ActiveIndex == 2 { // Tasks tab
		switch msg.String() {
		case "a":
			// Annotate task
			if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
				task := m.tasks[m.tasksIndex]
				m.taskAnnotationModal.SetTask(task.ID, task.Title)
				m.taskAnnotationModal.Show()
				return m, nil
			}
		case "d":
			// Mark task done
			if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
				task := m.tasks[m.tasksIndex]
				ctx := context.Background()
				err := m.data.Tasks.MarkDone(ctx, task.ID)
				if err != nil {
					m.statusMsg = fmt.Sprintf("Failed to mark done: %v", err)
				} else {
					m.statusMsg = "Task marked done, reloading..."
					m.taskState = TaskLoading
					return m, m.loadTasksCmd()
				}
			}
		case "t":
			// Mark task as todo
			if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
				task := m.tasks[m.tasksIndex]
				ctx := context.Background()
				err := m.data.Tasks.MarkTodo(ctx, task.ID)
				if err != nil {
					m.statusMsg = fmt.Sprintf("Failed to mark todo: %v", err)
				} else {
					m.statusMsg = "Task marked todo, reloading..."
					m.taskState = TaskLoading
					return m, m.loadTasksCmd()
				}
			}
		}
	}

	// Fallback keys for agent list navigation (not in resolver yet)
	switch msg.String() {
	case "g":
		// Go to top
		m.detailOffset = 0
		return m, nil
	case "G":
		// Go to bottom
		m.detailOffset = 9999
		return m, nil
	}

	return m, nil
}

// getCurrentDetailTabName returns the keybindings scope name for the current detail tab
func getCurrentDetailTabName(tabIndex int) string {
	switch tabIndex {
	case 0: // tabBack
		return "back"
	case 1: // tabOverview
		return "overview"
	case 2: // tabTasks
		return "tasks"
	case 3: // tabLogs
		return "logs"
	case 4: // tabCommits
		return "commits"
	case 5: // tabNotes
		return "notes"
	case 6: // tabSessionContext
		return "session_context"
	default:
		return "overview"
	}
}
