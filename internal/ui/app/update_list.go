package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// handleListKeys handles keys in list view
func (m *Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log all key presses for debugging
	debugf("Key pressed: %s", msg.String())

	// Tab navigation
	switch msg.String() {
	case "tab", "right":
		m.listTabs.Next()
		return m, nil
	case "shift+tab", "left":
		m.listTabs.Prev()
		return m, nil
	case "o":
		m.listTabs.Set(0) // Overview
		return m, nil
	case "p":
		m.listTabs.Set(1) // Project
		return m, nil
	case "c":
		m.listTabs.Set(2) // Claude
		return m, nil
	case "t", "T":
		m.listTabs.Set(3) // Usage
		m.statusMsg = "Switched to Usage tab"
		return m, nil
	case "u", "U":
		m.listTabs.Set(3) // Usage (support both u and U for backward compatibility)
		m.statusMsg = "Switched to Usage tab"
		return m, nil
	case "i", "I":
		// Initialize CCUsage database (only in Usage tab when empty)
		if m.listTabs.ActiveIndex == 3 && m.ccusageDB != nil && m.ccusageEntryCount == 0 && !m.ccusageSyncing {
			m.ccusageSyncing = true
			m.ccusageSyncStatus = "Initializing database..."
			return m, tea.Batch(
				m.initCCUsageCmd(),
				m.tickCmd(),
			)
		}
		return m, nil
	case "n":
		// 'n' is reserved for NewSession - don't handle it here, let it fall through
	}

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
		// Set active tab to Overview (index 1, since 0 is back arrow)
		m.tabs.Set(1)
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

	if Matches(msg, m.keys.NewSession) {
		debugf("NewSession key detected: %v, claudePath: %s", msg.String(), m.claudePath)
		// Always set a status message so we know the key was detected
		if m.claudePath == "" {
			m.statusMsg = "ERROR: Claude binary not found in PATH"
			m.err = nil // Clear any previous errors
			debugf("Claude path is empty")
			return m, nil
		}
		m.statusMsg = "Creating new session..."
		m.err = nil // Clear any previous errors
		debugf("Launching new session...")
		return m, NewSession(m.claudePath, m.config.DefaultTerminal)
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
		return m, util.FocusWindowCmd(m.windowFocuser, agent.WindowID, agent.TerminalApplication)
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
