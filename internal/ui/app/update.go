package app

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.MouseMsg:
		// Handle mouse clicks on tabs
		if msg.Type == tea.MouseLeft {
			if m.currentView == ViewList {
				// List view tab clicks
				m.listTabs.Update(msg)
			} else if m.currentView == ViewDetail {
				// Detail view tab clicks
				m.tabs.Update(msg)
				// If user clicked on back arrow (index 0), go back to list
				if m.tabs.ActiveIndex == 0 {
					m.currentView = ViewList
					m.detailOffset = 0
					return m, nil
				}
				// Load data for the new tab
				if err := m.loadTabData(); err != nil {
					m.err = err
				}
			}
		}
		return m, nil

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
		// Refresh CCUsage data if needed (every 30 seconds)
		if m.ccusageDB != nil && time.Since(m.lastCCUsageSync) > 30*time.Second {
			if err := m.loadCCUsageData(); err != nil {
				// Log but don't fail
				log.Printf("Warning: Failed to reload ccusage data: %v\n", err)
			}
		}
		// Schedule next tick
		return m, m.tickCmd()

	case initCCUsageMsg:
		// Sync complete, data reloaded
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


	// List-specific keys (only if we're in list view)
	if m.currentView == ViewList {
		return m.handleListKeys(msg)
	}

	return m, nil
}

// handleListKeys handles keys in list view
func (m *Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		// Always set a status message so we know the key was detected
		if m.claudePath == "" {
			m.statusMsg = "ERROR: Claude binary not found in PATH"
			m.err = nil // Clear any previous errors
			return m, nil
		}
		m.statusMsg = "Creating new session..."
		m.err = nil // Clear any previous errors
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

// handleDetailKeys handles keys in detail view
func (m *Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In detail view, q/esc go back to list instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to list view
		m.currentView = ViewList
		m.detailOffset = 0
		return m, nil
	}

	// Tab navigation (left/right arrow keys or tab/shift-tab)
	switch msg.String() {
	case "tab", "right":
		m.tabs.Next()
		// If user navigated to index 0 (back arrow), go back to list
		if m.tabs.ActiveIndex == 0 {
			m.currentView = ViewList
			m.detailOffset = 0
			return m, nil
		}
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "shift+tab", "left":
		m.tabs.Prev()
		// If user navigated to index 0 (back arrow), go back to list
		if m.tabs.ActiveIndex == 0 {
			m.currentView = ViewList
			m.detailOffset = 0
			return m, nil
		}
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "o":
		m.tabs.Set(1) // Overview tab (index 1 now, 0 is back arrow)
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "c":
		m.tabs.Set(2) // Commits tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "l":
		m.tabs.Set(3) // Logs tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "n":
		m.tabs.Set(4) // Notes tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "a":
		m.tabs.Set(5) // Actions tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	}

	// View navigation keys (only from Detail view)
	switch msg.String() {
	}

	// j/k navigation for split-pane tabs
	switch msg.String() {
	case "j", "down":
		// Navigate down in the current tab's list
		switch m.tabs.ActiveIndex {
		case 2: // Commits
			if m.commitsIndex < len(m.commits)-1 {
				m.commitsIndex++
				m.rightPaneOffset = 0 // Reset right pane scroll
			}
		case 3: // Logs
			if m.logsIndex < len(m.logs)-1 {
				m.logsIndex++
				m.rightPaneOffset = 0
			}
		case 4: // Notes
			if m.notesIndex < len(m.notes)-1 {
				m.notesIndex++
				m.rightPaneOffset = 0
			}
		case 5: // Actions
			if m.actionsIndex < len(m.actions)-1 {
				m.actionsIndex++
				m.rightPaneOffset = 0
			}
		default:
			// Overview tab - use default scrolling
			m.detailOffset++
		}
		return m, nil

	case "k", "up":
		// Navigate up in the current tab's list
		switch m.tabs.ActiveIndex {
		case 2: // Commits
			if m.commitsIndex > 0 {
				m.commitsIndex--
				m.rightPaneOffset = 0
			}
		case 3: // Logs
			if m.logsIndex > 0 {
				m.logsIndex--
				m.rightPaneOffset = 0
			}
		case 4: // Notes
			if m.notesIndex > 0 {
				m.notesIndex--
				m.rightPaneOffset = 0
			}
		case 5: // Actions
			if m.actionsIndex > 0 {
				m.actionsIndex--
				m.rightPaneOffset = 0
			}
		default:
			// Overview tab - use default scrolling
			if m.detailOffset > 0 {
				m.detailOffset--
			}
		}
		return m, nil

	case "h":
		// Scroll right pane up (only for split-pane tabs)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 5 {
			if m.rightPaneOffset > 0 {
				m.rightPaneOffset--
			}
			return m, nil
		}

	case "l":
		// Scroll right pane down (only for split-pane tabs)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 5 {
			m.rightPaneOffset++
			return m, nil
		}

	case "ctrl+u":
		// Page up in right pane (or overview)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 5 {
			// Split-pane tabs: page up right pane
			m.rightPaneOffset -= 10
			if m.rightPaneOffset < 0 {
				m.rightPaneOffset = 0
			}
		} else {
			// Overview tab: page up content
			m.detailOffset -= 10
			if m.detailOffset < 0 {
				m.detailOffset = 0
			}
		}
		return m, nil

	case "ctrl+d":
		// Page down in right pane (or overview)
		if m.tabs.ActiveIndex >= 2 && m.tabs.ActiveIndex <= 5 {
			// Split-pane tabs: page down right pane
			m.rightPaneOffset += 10
		} else {
			// Overview tab: page down content
			m.detailOffset += 10
		}
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
			message: fmt.Sprintf("Archived agent %s", truncateID(agentID, 8)),
		}
	}
}

// handleTasksKeys handles keys in tasks view
func (m *Model) handleTasksKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In tasks view, q/esc go back to detail instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to detail view
		m.currentView = ViewDetail
		m.tasksIndex = 0
		m.tasksOffset = 0
		return m, nil
	}

	if Matches(msg, m.keys.NavigateDown) {
		// Move down in list
		if m.tasksIndex < len(m.tasks)-1 {
			m.tasksIndex++
			m.adjustTasksScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.NavigateUp) {
		// Move up in list
		if m.tasksIndex > 0 {
			m.tasksIndex--
			m.adjustTasksScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.ScrollDown) {
		// Scroll detail pane down
		m.rightPaneOffset++
		return m, nil
	}

	if Matches(msg, m.keys.ScrollUp) {
		// Scroll detail pane up
		if m.rightPaneOffset > 0 {
			m.rightPaneOffset--
		}
		return m, nil
	}

	if Matches(msg, m.keys.Refresh) {
		// Reload tasks
		if err := m.loadTasks(); err != nil {
			m.err = err
			m.statusMsg = fmt.Sprintf("Failed to reload tasks: %v", err)
		} else {
			m.statusMsg = "Tasks refreshed"
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
		// Page down
		m.tasksIndex += 10
		if m.tasksIndex >= len(m.tasks) {
			m.tasksIndex = len(m.tasks) - 1
		}
		if m.tasksIndex < 0 {
			m.tasksIndex = 0
		}
		m.adjustTasksScroll()
		return m, nil
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.tasksIndex -= 10
		if m.tasksIndex < 0 {
			m.tasksIndex = 0
		}
		m.adjustTasksScroll()
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.tasksIndex = 0
		m.tasksOffset = 0
	case "G":
		// Go to bottom
		if len(m.tasks) > 0 {
			m.tasksIndex = len(m.tasks) - 1
			m.adjustTasksScroll()
		}
	case "d":
		// Mark task done
		if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
			if err := m.markTaskDone(); err != nil {
				m.err = err
				m.statusMsg = fmt.Sprintf("Failed to mark task done: %v", err)
			} else {
				m.statusMsg = "Task marked done"
			}
		}
	}

	return m, nil
}

// handleCommitsKeys handles keys in commits view
func (m *Model) handleCommitsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In commits view, q/esc go back to detail instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to detail view
		m.currentView = ViewDetail
		m.commitsIndex = 0
		m.commitsOffset = 0
		return m, nil
	}

	if Matches(msg, m.keys.NavigateDown) {
		// Move down in list
		if m.commitsIndex < len(m.commits)-1 {
			m.commitsIndex++
			m.adjustCommitsScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.NavigateUp) {
		// Move up in list
		if m.commitsIndex > 0 {
			m.commitsIndex--
			m.adjustCommitsScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.ScrollDown) {
		// Scroll detail pane down
		m.rightPaneOffset++
		return m, nil
	}

	if Matches(msg, m.keys.ScrollUp) {
		// Scroll detail pane up
		if m.rightPaneOffset > 0 {
			m.rightPaneOffset--
		}
		return m, nil
	}

	if Matches(msg, m.keys.Refresh) {
		// Reload commits
		if err := m.loadAgentDetails(); err != nil {
			m.err = err
			m.statusMsg = fmt.Sprintf("Failed to reload commits: %v", err)
		} else {
			m.statusMsg = "Commits refreshed"
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
		// Page down
		m.commitsIndex += 10
		if m.commitsIndex >= len(m.commits) {
			m.commitsIndex = len(m.commits) - 1
		}
		if m.commitsIndex < 0 {
			m.commitsIndex = 0
		}
		m.adjustCommitsScroll()
		return m, nil
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.commitsIndex -= 10
		if m.commitsIndex < 0 {
			m.commitsIndex = 0
		}
		m.adjustCommitsScroll()
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.commitsIndex = 0
		m.commitsOffset = 0
	case "G":
		// Go to bottom
		if len(m.commits) > 0 {
			m.commitsIndex = len(m.commits) - 1
			m.adjustCommitsScroll()
		}
	}

	return m, nil
}

// handleNotesKeys handles keys in notes view
func (m *Model) handleNotesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In notes view, q/esc go back to detail instead of quitting
	if Matches(msg, m.keys.Back) || Matches(msg, m.keys.Quit) {
		// Go back to detail view
		m.currentView = ViewDetail
		m.notesIndex = 0
		m.notesOffset = 0
		return m, nil
	}

	if Matches(msg, m.keys.NavigateDown) {
		// Move down in list
		if m.notesIndex < len(m.notes)-1 {
			m.notesIndex++
			m.adjustNotesScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.NavigateUp) {
		// Move up in list
		if m.notesIndex > 0 {
			m.notesIndex--
			m.adjustNotesScroll()
			m.rightPaneOffset = 0 // Reset detail scroll
		}
		return m, nil
	}

	if Matches(msg, m.keys.ScrollDown) {
		// Scroll detail pane down
		m.rightPaneOffset++
		return m, nil
	}

	if Matches(msg, m.keys.ScrollUp) {
		// Scroll detail pane up
		if m.rightPaneOffset > 0 {
			m.rightPaneOffset--
		}
		return m, nil
	}

	if Matches(msg, m.keys.Refresh) {
		// Reload notes
		if err := m.loadAgentDetails(); err != nil {
			m.err = err
			m.statusMsg = fmt.Sprintf("Failed to reload notes: %v", err)
		} else {
			m.statusMsg = "Notes refreshed"
		}
		return m, nil
	}

	if Matches(msg, m.keys.PageDown) {
		// Page down
		m.notesIndex += 10
		if m.notesIndex >= len(m.notes) {
			m.notesIndex = len(m.notes) - 1
		}
		if m.notesIndex < 0 {
			m.notesIndex = 0
		}
		m.adjustNotesScroll()
		return m, nil
	}

	if Matches(msg, m.keys.PageUp) {
		// Page up
		m.notesIndex -= 10
		if m.notesIndex < 0 {
			m.notesIndex = 0
		}
		m.adjustNotesScroll()
		return m, nil
	}

	// Fallback keys (temporary)
	switch msg.String() {
	case "g":
		// Go to top
		m.notesIndex = 0
		m.notesOffset = 0
	case "G":
		// Go to bottom
		if len(m.notes) > 0 {
			m.notesIndex = len(m.notes) - 1
			m.adjustNotesScroll()
		}
	}

	return m, nil
}

// adjustTasksScroll adjusts tasks scroll offset to keep selected item visible
func (m *Model) adjustTasksScroll() {
	visibleHeight := m.height - 8
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	if m.tasksIndex < m.tasksOffset {
		m.tasksOffset = m.tasksIndex
	}

	if m.tasksIndex >= m.tasksOffset+visibleHeight {
		m.tasksOffset = m.tasksIndex - visibleHeight + 1
	}
}

// adjustCommitsScroll adjusts commits scroll offset to keep selected item visible
func (m *Model) adjustCommitsScroll() {
	visibleHeight := m.height - 8
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	if m.commitsIndex < m.commitsOffset {
		m.commitsOffset = m.commitsIndex
	}

	if m.commitsIndex >= m.commitsOffset+visibleHeight {
		m.commitsOffset = m.commitsIndex - visibleHeight + 1
	}
}

// adjustNotesScroll adjusts notes scroll offset to keep selected item visible
func (m *Model) adjustNotesScroll() {
	visibleHeight := m.height - 8
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	if m.notesIndex < m.notesOffset {
		m.notesOffset = m.notesIndex
	}

	if m.notesIndex >= m.notesOffset+visibleHeight {
		m.notesOffset = m.notesIndex - visibleHeight + 1
	}
}

// initCCUsageCmd runs database initialization in background
func (m *Model) initCCUsageCmd() tea.Cmd {
	return func() tea.Msg {
		if m.ccusageDB == nil {
			return initCCUsageMsg{}
		}

		syncMgr := parser.NewSyncManager(m.ccusageDB)
		if err := syncMgr.Sync(); err != nil {
			m.ccusageSyncStatus = fmt.Sprintf("Error: %v", err)
			m.ccusageSyncing = false
			return initCCUsageMsg{}
		}

		// Reload data
		if err := m.loadCCUsageData(); err != nil {
			m.ccusageSyncStatus = fmt.Sprintf("Error loading data: %v", err)
		} else {
			m.ccusageSyncStatus = fmt.Sprintf("Initialized! Found %d entries", m.ccusageEntryCount)
		}

		m.ccusageSyncing = false
		return initCCUsageMsg{}
	}
}
