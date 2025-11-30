package agent_details

import tea "github.com/charmbracelet/bubbletea"

// BackToOverviewMsg is sent when user wants to go back to overview
type BackToOverviewMsg struct{}

// Update handles messages for the agent details view
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Tab navigation - single-letter jumps
		case "o":
			m.tabs.Set(tabOverview)
			return m, nil
		case "t":
			m.tabs.Set(tabTasks)
			return m, nil
		case "a":
			m.tabs.Set(tabActions)
			return m, nil
		case "l":
			m.tabs.Set(tabLogs)
			return m, nil
		case "c":
			m.tabs.Set(tabCommits)
			return m, nil
		case "n":
			m.tabs.Set(tabNotes)
			return m, nil
		case "s":
			m.tabs.Set(tabSessionContext)
			return m, nil

		// Back to overview
		case "esc", "left":
			return m, func() tea.Msg { return BackToOverviewMsg{} }

		// Task navigation (j/k when on tasks tab)
		case "j", "down":
			if m.tabs.ActiveIndex == tabTasks && m.ctx != nil {
				if m.ctx.SelectedTaskIndex < len(m.ctx.Tasks)-1 {
					m.ctx.SelectedTaskIndex++
					// TODO: Load task notes for selected task
				}
			}
			// Notes navigation (j/k when on notes tab)
			if m.tabs.ActiveIndex == tabNotes && m.ctx != nil {
				if m.ctx.NotesIndex < len(m.ctx.Notes)-1 {
					m.ctx.NotesIndex++
				}
			}
			// Commits navigation (j when on commits tab)
			if m.tabs.ActiveIndex == tabCommits && m.ctx != nil {
				if m.ctx.CommitsIndex < len(m.ctx.Commits)-1 {
					m.ctx.CommitsIndex++
				}
			}
			return m, nil
		case "k", "up":
			if m.tabs.ActiveIndex == tabTasks && m.ctx != nil {
				if m.ctx.SelectedTaskIndex > 0 {
					m.ctx.SelectedTaskIndex--
					// TODO: Load task notes for selected task
				}
			}
			// Notes navigation (k when on notes tab)
			if m.tabs.ActiveIndex == tabNotes && m.ctx != nil {
				if m.ctx.NotesIndex > 0 {
					m.ctx.NotesIndex--
				}
			}
			// Commits navigation (k when on commits tab)
			if m.tabs.ActiveIndex == tabCommits && m.ctx != nil {
				if m.ctx.CommitsIndex > 0 {
					m.ctx.CommitsIndex--
				}
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
