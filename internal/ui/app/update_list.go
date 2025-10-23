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
	case "p", "P":
		m.listTabs.Set(1) // Project
		return m, nil
	case "c":
		m.listTabs.Set(2) // Claude
		return m, nil
	case "t", "T":
		m.listTabs.Set(3) // Usage
		m.statusMsg = "Switched to Usage tab"
		// Trigger usage content refresh
		return m, func() tea.Msg { return RefreshUsageMsg{} }
	case "u", "U":
		m.listTabs.Set(3) // Usage (support both u and U for backward compatibility)
		m.statusMsg = "Switched to Usage tab"
		// Trigger usage content refresh
		return m, func() tea.Msg { return RefreshUsageMsg{} }
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
		// Switch to detail view - load details with command
		if m.selectedIndex >= 0 && m.selectedIndex < len(m.agents) {
			m.selectedAgent = &m.agents[m.selectedIndex]
			m.currentView = ViewDetail
			m.detailOffset = 0
			// Set active tab to Overview (index 1, since 0 is back arrow)
			m.tabs.Set(1)
			return m, loadAgentDetailsCmd(m.data, m.selectedAgent.ID)
		}
		return m, nil
	}

	// Usage tab scrolling (only when Usage tab is active)
	if m.listTabs.ActiveIndex == 3 {
		switch msg.String() {
		case "up", "k":
			m.usageViewport.LineUp(1)
			return m, nil
		case "down", "j":
			m.usageViewport.LineDown(1)
			return m, nil
		case "pgup":
			m.usageViewport.HalfViewUp()
			return m, nil
		case "pgdown":
			m.usageViewport.HalfViewDown()
			return m, nil
		}
	}

	// Project tab navigation and scrolling (only when Project tab is active)
	if m.listTabs.ActiveIndex == 1 {
		switch msg.String() {
		case "1":
			m.projectSelectedSection = 0 // Tasks
			m.projectTasksIndex = 0
			return m, nil
		case "2":
			m.projectSelectedSection = 1 // CLAUDE.md
			return m, nil
		case "3":
			m.projectSelectedSection = 2 // Markdown files
			m.projectMDFilesIndex = 0
			return m, nil
		case "j", "down":
			// Navigate within section
			switch m.projectSelectedSection {
			case 0: // Tasks
				if m.projectTasksIndex < len(m.projectTasks)-1 {
					m.projectTasksIndex++
				}
			case 2: // Markdown files
				if m.projectMDFilesIndex < len(m.projectMDFiles)-1 {
					m.projectMDFilesIndex++
				}
			}
			return m, nil
		case "k", "up":
			// Navigate within section
			switch m.projectSelectedSection {
			case 0: // Tasks
				if m.projectTasksIndex > 0 {
					m.projectTasksIndex--
				}
			case 2: // Markdown files
				if m.projectMDFilesIndex > 0 {
					m.projectMDFilesIndex--
				}
			}
			return m, nil
		}
	}

	// Claude tab navigation and operations (only when Claude tab is active)
	if m.listTabs.ActiveIndex == 2 {
		switch msg.String() {
		case "j", "down":
			// Navigate file list down
			if !m.claudeShowingContent && m.claudeSelectedIndex < len(m.claudeFiles)-1 {
				m.claudeSelectedIndex++
			} else if m.claudeShowingContent {
				// Scroll content viewport down
				m.claudeViewport.LineDown(1)
			}
			return m, nil
		case "k", "up":
			// Navigate file list up
			if !m.claudeShowingContent && m.claudeSelectedIndex > 0 {
				m.claudeSelectedIndex--
			} else if m.claudeShowingContent {
				// Scroll content viewport up
				m.claudeViewport.LineUp(1)
			}
			return m, nil
		case "enter":
			// Load selected file
			if m.claudeSelectedIndex >= 0 && m.claudeSelectedIndex < len(m.claudeFiles) {
				selectedFile := m.claudeFiles[m.claudeSelectedIndex]
				if !selectedFile.IsDir {
					m.claudeShowingContent = true
					return m, loadClaudeFileContentCmd(selectedFile.Path)
				}
			}
			return m, nil
		case "v":
			// Validate JSON
			if m.claudeShowingContent && m.claudeContent != "" {
				return m, validateClaudeFileCmd(m.claudeContent)
			}
			return m, nil
		case "e":
			// Open in editor (only for files, not directories)
			if m.claudeSelectedIndex >= 0 && m.claudeSelectedIndex < len(m.claudeFiles) {
				selectedFile := m.claudeFiles[m.claudeSelectedIndex]
				if !selectedFile.IsDir {
					return m, openClaudeFileInEditor(selectedFile.Path)
				}
			}
			return m, nil
		case "r":
			// Refresh file list
			m.statusMsg = "Refreshing Claude config files..."
			return m, m.loadClaudeFilesCmd()
		case "esc", "q":
			// Back to file list from content view
			if m.claudeShowingContent {
				m.claudeShowingContent = false
				m.claudeContent = ""
				m.claudeValidStatus = ""
				return m, nil
			}
			return m, nil
		case "pgup":
			if m.claudeShowingContent {
				m.claudeViewport.HalfViewUp()
			}
			return m, nil
		case "pgdown":
			if m.claudeShowingContent {
				m.claudeViewport.HalfViewDown()
			}
			return m, nil
		}
	}

	if Matches(msg, m.keys.ToggleFilter) {
		// Toggle show all agents
		m.showAll = !m.showAll
		return m, loadAgentsCmd(m.data.Agents)
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
		return m, ResumeSession(m.claudePath, string(agent.ID))
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
		return m, StartSession(m.claudePath, string(agent.ID))
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
		return m, m.archiveAgentCmd(string(agent.ID))
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
