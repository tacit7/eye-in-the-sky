package project

import tea "github.com/charmbracelet/bubbletea"

// BackToOverviewMsg is sent when user wants to go back to overview
type BackToOverviewMsg struct{}

// Update handles messages for the project details view
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Tab navigation - single-letter jumps
		case "o":
			m.tabs.Set(tabOverview)
			return m, nil
		case "a":
			m.tabs.Set(tabAgents)
			return m, nil
		case "n":
			m.tabs.Set(tabNotes)
			return m, nil
		case "t":
			m.tabs.Set(tabTasks)
			return m, nil
		case "f":
			m.tabs.Set(tabFiles)
			return m, nil

		// Back to overview
		case "esc", "left":
			return m, func() tea.Msg { return BackToOverviewMsg{} }

		// Navigation within tabs
		// For table-based tabs, delegate to table.Update()
		// For other tabs, handle manually
		case "j", "down", "k", "up":
			if m.ctx == nil {
				return m, nil
			}

			// Delegate to table for table-based tabs
			switch m.tabs.ActiveIndex {
			case tabAgents:
				var cmd tea.Cmd
				m.agentsTable, cmd = m.agentsTable.Update(msg)
				// Sync selection index with table cursor
				m.ctx.SelectedAgentIndex = m.agentsTable.Cursor()
				m.ctx.RightPaneOffset = 0
				return m, cmd
			case tabNotes:
				var cmd tea.Cmd
				m.notesTable, cmd = m.notesTable.Update(msg)
				m.ctx.NotesIndex = m.notesTable.Cursor()
				m.ctx.RightPaneOffset = 0
				return m, cmd
			case tabTasks:
				var cmd tea.Cmd
				m.tasksTable, cmd = m.tasksTable.Update(msg)
				m.ctx.TasksIndex = m.tasksTable.Cursor()
				m.ctx.RightPaneOffset = 0
				return m, cmd
			}
			return m, nil

		// Scrolling in right pane (h/l)
		case "h":
			if m.ctx != nil && m.ctx.RightPaneOffset > 0 {
				m.ctx.RightPaneOffset -= 5
				if m.ctx.RightPaneOffset < 0 {
					m.ctx.RightPaneOffset = 0
				}
			}
			return m, nil
		case "l":
			if m.ctx != nil {
				m.ctx.RightPaneOffset += 5
			}
			return m, nil
		}

		// Handle enter on back arrow tab
		if m.tabs.ActiveIndex == tabBack && msg.String() == "enter" {
			return m, func() tea.Msg { return BackToOverviewMsg{} }
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	}

	return m, nil
}
