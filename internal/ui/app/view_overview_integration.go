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
		overviewModel := m.overviewView.(overview.Model)
		selectedAgent := overviewModel.SelectedAgent()
		if selectedAgent != nil {
			// Get project ID from project name
			projectID := ""
			if selectedAgent.ProjectName != "" {
				if proj, err := m.data.DB.GetProjectByName(selectedAgent.ProjectName); err == nil && proj != nil {
					projectID = proj.ID
				}
			}

			m.noteModal.SetContext(string(selectedAgent.ID), selectedAgent.SessionID, projectID)
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

	case overview.TabChangedMsg:
		// Tab changed in overview view - trigger refresh if it's the usage tab
		if msg.NewTabIndex == 3 {
			// Token Usage tab - load data if not loaded yet, then refresh viewport
			// If data hasn't been synced yet, load it now (async)
			if m.ccusageDB != nil && m.lastCCUsageSync.IsZero() {
				return m, m.loadCCUsageDataCmd()
			}

			// Data already loaded, just refresh
			return m, func() tea.Msg {
				return RefreshUsageMsg{}
			}
		}
		return m, nil
	}

	return m, cmd
}
