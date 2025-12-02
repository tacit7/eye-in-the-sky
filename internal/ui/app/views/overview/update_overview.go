package overview

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/messages"
)

// Update handles messages for the OverviewView
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		log.Printf("[OVERVIEW] Window size: %dx%d", msg.Width, msg.Height)
		return m, nil

	case messages.AgentsLoadedMsg:
		// Agents loaded successfully
		m.agents = msg.Agents
		m.statusMsg = "Agents loaded"
		log.Printf("[OVERVIEW] Agents loaded: %d agents", len(msg.Agents))
		for i, a := range msg.Agents {
			log.Printf("[OVERVIEW]   Agent %d: %s (status=%s, desc=%s)", i, a.ID, a.Status, a.FeatureDesc)
		}
		// Refresh the agents table with new data
		m.refreshAgentsTable()
		return m, nil

	case ErrMsg:
		// Error occurred
		m.statusMsg = "Error: " + msg.Error.Error()
		log.Printf("[OVERVIEW] Error: %v", msg.Error)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Direct tab navigation shortcuts (global, checked before per-tab handlers)
	// Note: Some keys like 'c', 'k', 'n' are intentionally omitted to avoid conflicts with agents tab actions
	switch msg.String() {
	case "o", "O":
		// Jump to Overview tab (index 0)
		oldTab := m.tabs.ActiveIndex
		m.tabs.Set(0)
		if oldTab != 0 {
			return m, nil
		}
		return m, nil
	case "p":
		// Jump to Project tab (index 1)
		oldTab := m.tabs.ActiveIndex
		m.tabs.Set(1)
		if oldTab != 1 {
			return m, nil
		}
		return m, nil
	case "P":
		// Shift+P: Navigate directly to Project Detail View
		return m, func() tea.Msg {
			return SelectProjectMsg{}
		}
	case "t", "T":
		// Jump to Token Usage tab (index 3)
		oldTab := m.tabs.ActiveIndex
		m.tabs.Set(3)
		// Emit tab changed message when switching to token usage
		if oldTab != 3 {
			return m, func() tea.Msg {
				return TabChangedMsg{NewTabIndex: 3}
			}
		}
		return m, nil
	}

	// Tab navigation (sequential)
	switch msg.String() {
	case "tab":
		oldTab := m.tabs.ActiveIndex
		m.tabs.Next()
		newTab := m.tabs.ActiveIndex
		// Emit tab changed message if switched to token usage tab (index 3)
		if oldTab != newTab && newTab == 3 {
			return m, func() tea.Msg {
				return TabChangedMsg{NewTabIndex: newTab}
			}
		}
		return m, nil
	case "shift+tab", "left":
		oldTab := m.tabs.ActiveIndex
		m.tabs.Prev()
		newTab := m.tabs.ActiveIndex
		// Emit tab changed message if switched to token usage tab (index 3)
		if oldTab != newTab && newTab == 3 {
			return m, func() tea.Msg {
				return TabChangedMsg{NewTabIndex: newTab}
			}
		}
		return m, nil
	}

	// Route to active tab handler
	log.Printf("[OVERVIEW] Routing key '%s' to tab handler for activeTab=%d", msg.String(), m.tabs.ActiveIndex)
	switch m.tabs.ActiveIndex {
	case 0:
		return m.handleAgentsTabKeys(msg)
	case 1:
		return m.handleProjectTabKeys(msg)
	case 2:
		return m.handleClaudeTabKeys(msg)
	case 3:
		return m.handleUsageTabKeys(msg)
	case 4:
		return m.handleConfigTabKeys(msg)
	}

	return m, nil
}

// handleAgentsTabKeys handles keys for the agents list tab
func (m Model) handleAgentsTabKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If header is focused, handle header navigation
	if m.headerFocused {
		return m.handleHeaderKeys(msg)
	}

	switch msg.String() {
	case "j", "down", "k", "up":
		// Delegate navigation to table.Model
		var cmd tea.Cmd
		m.agentsTable, cmd = m.agentsTable.Update(msg)
		// Sync selectedIndex with table cursor
		m.selectedIndex = m.agentsTable.Cursor()
		m.adjustListScroll()

		// If at first row and pressing up, move to header
		if msg.String() == "k" || msg.String() == "up" {
			if m.selectedIndex == 0 && m.agentsTable.Cursor() == 0 {
				m.headerFocused = true
				return m, nil
			}
		}
		return m, cmd

	case "g":
		// Go to top
		m.selectedIndex = 0
		m.listOffset = 0
		m.agentsTable.SetCursor(0)
		return m, nil

	case "G":
		// Go to bottom
		visibleAgents := m.GetVisibleAgents()
		if len(visibleAgents) > 0 {
			m.selectedIndex = len(visibleAgents) - 1
			m.agentsTable.SetCursor(m.selectedIndex)
			m.adjustListScroll()
		}
		return m, nil

	case "a", "A":
		// Toggle show all agents
		m.showAll = !m.showAll
		return m, m.loadAgentsCmd()

	case "n", "N":
		// New session - emit message for root to handle
		// (Root will open modal)
		return m, func() tea.Msg {
			return NewSessionRequestMsg{}
		}

	case "enter", " ", "right":
		// Select agent - emit message for root to switch to detail view
		if m.SelectedAgent() != nil {
			return m, func() tea.Msg {
				return SelectAgentMsg{Agent: m.SelectedAgent()}
			}
		}
		return m, nil

	case "c", "C", "d", "D":
		// Mark current session as completed
		if m.SelectedAgent() != nil {
			return m, func() tea.Msg {
				return MarkSessionCompletedMsg{Agent: m.SelectedAgent()}
			}
		}
		return m, nil

	case "r", "R":
		// Refresh agents
		return m, m.loadAgentsCmd()
	}

	return m, nil
}

// handleProjectTabKeys handles keys for the project tab
func (m Model) handleProjectTabKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ", "right":
		// Navigate to project detail view - emit message for root to handle
		return m, func() tea.Msg {
			return SelectProjectMsg{}
		}
	case "j", "k", "down", "up", "pgdown", "pgup", "home", "end":
		// Forward navigation keys to project viewport
		return m, func() tea.Msg {
			return ViewportUpdateMsg{Target: "project", Msg: msg}
		}
	}
	return m, nil
}

// handleClaudeTabKeys handles keys for the Claude config tab
func (m Model) handleClaudeTabKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Forward navigation keys to Claude viewport
	switch msg.String() {
	case "j", "k", "down", "up", "pgdown", "pgup", "home", "end":
		return m, func() tea.Msg {
			return ViewportUpdateMsg{Target: "claude", Msg: msg}
		}
	}
	return m, nil
}

// handleUsageTabKeys handles keys for the token usage tab
func (m Model) handleUsageTabKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Printf("[OVERVIEW] handleUsageTabKeys received key: '%s'", msg.String())
	switch msg.String() {
	case "j", "k", "down", "up", "pgdown", "pgup", "home", "end":
		// Forward navigation keys to usage viewport
		return m, func() tea.Msg {
			return ViewportUpdateMsg{Target: "usage", Msg: msg}
		}
	case "i", "I":
		// Request CCUsage sync - emit message for root to handle
		log.Printf("[OVERVIEW] Emitting SyncCCUsageRequestMsg from handleUsageTabKeys")
		return m, func() tea.Msg {
			return SyncCCUsageRequestMsg{}
		}
	}
	return m, nil
}

// handleConfigTabKeys handles keys for the keybindings config tab
func (m Model) handleConfigTabKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Forward navigation keys to keybindings viewport
	switch msg.String() {
	case "j", "k", "down", "up", "pgdown", "pgup", "home", "end":
		return m, func() tea.Msg {
			return ViewportUpdateMsg{Target: "keybindings", Msg: msg}
		}
	}
	return m, nil
}

// handleHeaderKeys handles keyboard input when header is focused
func (m Model) handleHeaderKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left":
		// Move to previous column
		if m.selectedColumn > 0 {
			m.selectedColumn--
		}
		return m, nil

	case "l", "right":
		// Move to next column (5 columns: Status, LastActivity, Session, Task, Project)
		if m.selectedColumn < 4 {
			m.selectedColumn++
		}
		return m, nil

	case "j", "down":
		// Move back to first row
		m.headerFocused = false
		m.selectedIndex = 0
		return m, nil

	case "enter", " ":
		// Sort by selected column
		return m.sortByColumn()

	case "esc":
		// Exit header focus
		m.headerFocused = false
		return m, nil
	}

	return m, nil
}

// sortByColumn sorts agents by the selected column
func (m Model) sortByColumn() (tea.Model, tea.Cmd) {
	// Map column index to field name (matches order in renderHeaderRow)
	columnFields := []string{"status", "lastlog", "session", "task", "project"}
	newField := columnFields[m.selectedColumn]

	// Toggle sort order if same field, otherwise default to ascending
	if m.sortField == newField {
		m.sortAscending = !m.sortAscending
	} else {
		m.sortField = newField
		m.sortAscending = true
	}

	// Sort agents based on field and order
	m.sortAgents()

	// Refresh the table with sorted data
	m.refreshAgentsTable()

	return m, nil
}

// NewSessionRequestMsg is sent when user wants to create a new session
// Root will handle this by opening a modal
type NewSessionRequestMsg struct{}

// SelectAgentMsg is sent when user selects an agent
// Root will handle this by switching to AgentDetailsView
type SelectAgentMsg struct {
	Agent *domain.Agent
}

// SelectProjectMsg is sent when user selects project (presses Enter on Project tab)
// Root will handle this by switching to ProjectDetailsView
type SelectProjectMsg struct{}

// MarkSessionCompletedMsg is sent when user wants to mark a session as completed
// Root will handle this by updating the agent status
type MarkSessionCompletedMsg struct {
	Agent *domain.Agent
}
