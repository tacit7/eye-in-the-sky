package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/overview"
)

// renderOverviewView renders the new modular overview view
func (m *Model) renderOverviewView() string {
	return m.overviewView.View()
}

// handleOverviewKeys handles key presses for the overview view
func (m *Model) handleOverviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleOverviewUpdate(msg)
}

// handleOverviewUpdate routes messages to the overview view
func (m *Model) handleOverviewUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle messages that should be processed by overview
	var cmd tea.Cmd
	m.overviewView, cmd = m.overviewView.Update(msg)

	// Check for messages from overview that root should handle
	switch msg := msg.(type) {
	case overview.SelectAgentMsg:
		// User selected an agent - switch to detail view
		m.selectedAgent = msg.Agent
		m.currentView = ViewDetail
		m.detailOffset = 0
		m.tabs.Set(1) // Set to Overview tab (index 1, since 0 is back arrow)
		return m, loadAgentDetailsCmd(m.data, m.selectedAgent.ID)

	case overview.NewSessionRequestMsg:
		// User wants to create note - open note modal
		selectedAgent := m.overviewView.(overview.Model).SelectedAgent()
		if selectedAgent != nil {
			// Set context for note modal
			sessionID := ""
			if selectedAgent.SessionID != nil {
				sessionID = *selectedAgent.SessionID
			}
			projectName := ""
			if selectedAgent.ProjectName != nil {
				projectName = *selectedAgent.ProjectName
			}

			// Get project ID from project name
			projectID := ""
			if projectName != "" {
				if proj, err := m.data.DB.GetProjectByName(projectName); err == nil && proj != nil {
					projectID = proj.ID
				}
			}

			m.noteModal.SetContext(selectedAgent.ID, sessionID, projectID)
			m.noteModal.Show()
		} else {
			m.statusMsg = "No agent selected"
		}
		return m, nil

	case overview.ViewportUpdateMsg:
		// Forward to appropriate viewport
		var viewportCmd tea.Cmd
		switch msg.Target {
		case "usage":
			m.usageViewport, viewportCmd = m.usageViewport.Update(msg.Msg)
		case "claude":
			m.claudeViewport, viewportCmd = m.claudeViewport.Update(msg.Msg)
		case "keybindings":
			m.keybindingsViewport, viewportCmd = m.keybindingsViewport.Update(msg.Msg)
		case "project":
			m.projectViewport, viewportCmd = m.projectViewport.Update(msg.Msg)
		}
		return m, viewportCmd
	}

	return m, cmd
}
