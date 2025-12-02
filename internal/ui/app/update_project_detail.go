package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"log"
)

// handleProjectDetailKeys handles keys in project detail view
func (m *Model) handleProjectDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Printf("[PROJECT-DETAIL] handleProjectDetailKeys: received key='%s'", msg.String())

	// In project detail view, q/esc go back to list instead of quitting
	if msg.String() == "q" || msg.String() == "esc" {
		// Go back to list view
		m.currentView = ViewList
		return m, nil
	}

	// Tab navigation shortcuts
	switch msg.String() {
	case "tab", "right":
		// Navigate forward, skipping tab 0 (back arrow)
		// Project tabs: 0=Back, 1=Overview, 2=Agents, 3=Notes, 4=Tasks, 5=Files
		if m.projectTabsIndex == 5 {
			m.projectTabsIndex = 1 // Wrap to Overview
		} else {
			m.projectTabsIndex++
			if m.projectTabsIndex == 0 {
				m.projectTabsIndex = 1 // Skip back button
			}
		}
		return m, nil
	case "shift+tab", "left":
		// Navigate backward, skipping tab 0 (back arrow)
		if m.projectTabsIndex == 1 {
			m.projectTabsIndex = 5 // Wrap to Files
		} else {
			m.projectTabsIndex--
			if m.projectTabsIndex == 0 {
				m.projectTabsIndex = 5 // Skip back button
			}
		}
		return m, nil
	case "o", "O":
		log.Printf("[PROJECT-DETAIL] KEY O PRESSED: Switching to Overview tab (1)")
		m.projectTabsIndex = 1
		return m, nil
	case "a", "A":
		log.Printf("[PROJECT-DETAIL] KEY A PRESSED: Switching to Agents tab (2)")
		m.projectTabsIndex = 2
		return m, nil
	case "n", "N":
		log.Printf("[PROJECT-DETAIL] KEY N PRESSED: Switching to Notes tab (3)")
		m.projectTabsIndex = 3
		return m, nil
	case "t", "T":
		log.Printf("[PROJECT-DETAIL] KEY T PRESSED: Switching to Tasks tab (4)")
		m.projectTabsIndex = 4
		return m, nil
	case "f", "F":
		log.Printf("[PROJECT-DETAIL] KEY F PRESSED: Switching to Files tab (5)")
		m.projectTabsIndex = 5
		return m, nil
	case "r", "R":
		// Refresh/rerender the agents table
		if m.projectTabsIndex == 2 { // Agents tab
			log.Printf("[PROJECT-DETAIL] KEY R PRESSED: Forcing table refresh")
			// Set flag to force recreation on next render
			m.projectAgentsForceRefresh = true
			m.statusMsg = "Refreshing agents table..."
			m.statusTime = time.Now()
			return m, nil
		}
		return m, nil
	case "j", "down", "k", "up":
		// Handle navigation in the agents table
		if m.projectTabsIndex == 2 { // Agents tab
			var cmd tea.Cmd
			m.projectAgentsTable, cmd = m.projectAgentsTable.Update(msg)
			m.projectAgentsIndex = m.projectAgentsTable.Cursor()
			log.Printf("[PROJECT-DETAIL] Agents table navigation: cursor now at %d", m.projectAgentsIndex)
			return m, cmd
		}
		// TODO: Handle notes and tasks tabs navigation
		return m, nil
	case "h", "l":
		// Handle horizontal scrolling in right pane (if needed)
		return m, nil
	}

	return m, nil
}
