package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tickMsg:
		// Refresh data
		if err := m.loadAgents(); err != nil {
			m.err = err
		}
		// Schedule next tick
		return m, m.tickCmd()

	case cmdResult:
		// Handle command execution results
		if msg.success {
			m.statusMsg = msg.message
		} else {
			m.err = msg.err
			m.statusMsg = msg.message
		}
		// Refresh agent list after command
		if err := m.loadAgents(); err != nil {
			m.err = err
		}
		return m, nil

	case util.FocusResult:
		// Handle window focus results
		if msg.Success {
			m.statusMsg = msg.Message
		} else {
			m.err = msg.Err
			m.statusMsg = msg.Message
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// errMsg is sent when an error occurs
type errMsg struct {
	err error
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle view-specific keys first (before global keys)
	switch m.currentView {
	case ViewDetail:
		return m.handleDetailKeys(msg)
	case ViewLogs:
		return m.handleLogsKeys(msg)
	case ViewList:
		// List view falls through to global keys, then list-specific
	}

	// Global keys (only checked if not handled by view-specific)
	if Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	if Matches(msg, m.keys.Refresh) {
		if err := m.loadAgents(); err != nil {
			m.err = err
		}
		m.statusMsg = "Refreshed"
		return m, nil
	}

	if Matches(msg, m.keys.HelpToggle) {
		m.showHelp = !m.showHelp
		return m, nil
	}

	// Fallback logging (temporary)
	if msg.String() == "j" {
		fmt.Fprintf(os.Stderr, "fallback key j used\n")
	}
	if msg.String() == "k" {
		fmt.Fprintf(os.Stderr, "fallback key k used\n")
	}

	// List-specific keys (only if we're in list view)
	if m.currentView == ViewList {
		return m.handleListKeys(msg)
	}

	return m, nil
}

// handleListKeys handles keys in list view
func (m *Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if Matches(msg, m.keys.NavigateDown) {
		if m.selectedIndex < len(m.agents)-1 {
			m.selectedIndex++
			m.adjustListScroll()
		}
		return m, nil
	}

	if Matches(msg, m.keys.NavigateUp) {
		if m.selectedIndex > 0 {
			m.selectedIndex--
			m.adjustListScroll()
		}
		return m, nil
	}

	if Matches(msg, m.keys.Select) {
		// Switch to detail view
		if err := m.loadAgentDetails(); err != nil {
			m.err = err
			return m, nil
		}
		m.currentView = ViewDetail
		m.detailOffset = 0
		return m, nil
	}

	if Matches(msg, m.keys.ToggleFilter) {
		// Toggle show all agents
		m.showAll = !m.showAll
		if err := m.loadAgents(); err != nil {
			m.err = err
		}
		return m, nil
	}

	if Matches(msg, m.keys.ShowLogs) {
		m.currentView = ViewLogs
		return m, nil
	}

	if Matches(msg, m.keys.NewSession) {
		if m.claudePath == "" {
			m.statusMsg = "Claude binary not found"
			return m, nil
		}
		return m, NewSession(m.claudePath)
	}

	if Matches(msg, m.keys.ContinueSession) {
		agent := m.SelectedAgent()
		if agent == nil {
			m.statusMsg = "No agent selected"
			return m, nil
		}
		if m.claudePath == "" {
			m.statusMsg = "Claude binary not found"
			return m, nil
		}
		// Use agent ID as session ID for now
		return m, ResumeSession(m.claudePath, agent.ID)
	}

	if Matches(msg, m.keys.StartSession) {
		agent := m.SelectedAgent()
		if agent == nil {
			m.statusMsg = "No agent selected"
			return m, nil
		}
		if m.claudePath == "" {
			m.statusMsg = "Claude binary not found"
			return m, nil
		}
		return m, StartSession(m.claudePath, agent.ID)
	}

	if Matches(msg, m.keys.GoToWindow) {
		agent := m.SelectedAgent()
		if agent == nil {
			m.statusMsg = "No agent selected"
			return m, nil
		}
		if agent.WindowID == "" {
			m.statusMsg = "No window ID for agent"
			return m, nil
		}
		return m, util.FocusWindowCmd(m.windowFocuser, agent.WindowID)
	}

	if Matches(msg, m.keys.Archive) {
		agent := m.SelectedAgent()
		if agent == nil {
			m.statusMsg = "No agent selected"
			return m, nil
		}
		return m, m.archiveAgentCmd(agent.ID)
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.selectedIndex = 0
		m.listOffset = 0
	case "G":
		// Go to bottom
		if len(m.agents) > 0 {
			m.selectedIndex = len(m.agents) - 1
			m.adjustListScroll()
		}
	}

	return m, nil
}

// handleDetailKeys handles keys in detail view
func (m *Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In detail view, q/esc go back to list instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to list view
		m.currentView = ViewList
		m.detailOffset = 0
		return m, nil
	}

	if Matches(msg, m.keys.ScrollDown) {
		// Scroll down
		m.detailOffset++
		return m, nil
	}

	if Matches(msg, m.keys.ScrollUp) {
		// Scroll up
		if m.detailOffset > 0 {
			m.detailOffset--
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
		// Page down
		m.detailOffset += 10
		return m, nil
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.detailOffset -= 10
		if m.detailOffset < 0 {
			m.detailOffset = 0
		}
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.detailOffset = 0
	case "G":
		// Go to bottom
		m.detailOffset = 9999
	}

	return m, nil
}

// handleLogsKeys handles keys in logs view
func (m *Model) handleLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In logs view, q/esc go back to list instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to list view
		m.currentView = ViewList
		return m, nil
	}

	if Matches(msg, m.keys.ScrollDown) {
		// Scroll down
		m.detailOffset++
		return m, nil
	}

	if Matches(msg, m.keys.ScrollUp) {
		// Scroll up
		if m.detailOffset > 0 {
			m.detailOffset--
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
		// Page down
		m.detailOffset += 10
		return m, nil
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.detailOffset -= 10
		if m.detailOffset < 0 {
			m.detailOffset = 0
		}
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.detailOffset = 0
	case "G":
		// Go to bottom
		m.detailOffset = 9999
	}

	return m, nil
}

// adjustListScroll adjusts list scroll offset to keep selected item visible
func (m *Model) adjustListScroll() {
	// Reserve space for header (3 lines) and footer (2 lines)
	visibleHeight := m.height - 5
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	// If selected item is above visible area, scroll up
	if m.selectedIndex < m.listOffset {
		m.listOffset = m.selectedIndex
	}

	// If selected item is below visible area, scroll down
	if m.selectedIndex >= m.listOffset+visibleHeight {
		m.listOffset = m.selectedIndex - visibleHeight + 1
	}
}

// archiveAgentCmd creates a command to archive an agent
func (m *Model) archiveAgentCmd(agentID string) tea.Cmd {
	return func() tea.Msg {
		// Update agent status to archived in database
		query := `UPDATE agents SET status = 'archived', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		if _, err := m.db.Exec(query, agentID); err != nil {
			return cmdResult{
				success: false,
				message: "Failed to archive agent",
				err:     err,
			}
		}

		return cmdResult{
			success: true,
			message: fmt.Sprintf("Archived agent %s", agentID[:8]),
		}
	}
}
