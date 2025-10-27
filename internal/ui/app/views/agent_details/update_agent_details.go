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
		case "p":
			m.tabs.Set(tabProjects)
			return m, nil

		// Back to overview
		case "esc", "left":
			return m, func() tea.Msg { return BackToOverviewMsg{} }
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
