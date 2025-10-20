package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.MouseMsg:
		// Handle mouse clicks in detail view for tab navigation
		if m.currentView == ViewDetail && msg.Type == tea.MouseLeft {
			m.tabs.Update(msg)
			// Load data for the new tab
			if err := m.loadTabData(); err != nil {
				m.err = err
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
	case ViewTasks:
		return m.handleTasksKeys(msg)
	case ViewCommits:
		return m.handleCommitsKeys(msg)
	case ViewNotes:
		return m.handleNotesKeys(msg)
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
			m.statusMsg = "ERROR: Claude binary not found in PATH"
			return m, nil
		}
		m.statusMsg = "Creating new session..."
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
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "shift+tab", "left":
		m.tabs.Prev()
		// Load data for the new tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "c":
		m.tabs.Set(0) // Commits tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "l":
		m.tabs.Set(1) // Logs tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "n":
		m.tabs.Set(2) // Notes tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	case "a":
		m.tabs.Set(3) // Actions tab
		if err := m.loadTabData(); err != nil {
			m.err = err
		}
		return m, nil
	}

	// View navigation keys (only from Detail view)
	switch msg.String() {
	case "t":
		// Switch to Tasks view
		m.currentView = ViewTasks
		if err := m.loadTasks(); err != nil {
			m.err = err
			m.statusMsg = fmt.Sprintf("Failed to load tasks: %v", err)
		}
		return m, nil

	case "C":
		// Switch to Commits view
		m.currentView = ViewCommits
		m.commitsIndex = 0
		return m, nil

	case "N":
		// Switch to Notes view
		m.currentView = ViewNotes
		m.notesIndex = 0
		return m, nil

	case "L":
		// Switch to Logs view
		m.currentView = ViewLogs
		if err := m.loadLogs(); err != nil {
			m.err = err
			m.statusMsg = fmt.Sprintf("Failed to load logs: %v", err)
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
