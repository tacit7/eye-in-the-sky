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
	// Tab navigation (direct keys)
	switch msg.String() {
	case "tab", "right":
		m.tabs.Next()
		return m, nil
	case "shift+tab", "left":
		m.tabs.Prev()
		return m, nil
	}

	// Route to active tab handler
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
	switch msg.String() {
	case "j", "down":
		// Move selection down
		if m.selectedIndex < len(m.agents)-1 {
			m.selectedIndex++
			m.adjustListScroll()
		}
		return m, nil

	case "k", "up":
		// Move selection up
		if m.selectedIndex > 0 {
			m.selectedIndex--
			m.adjustListScroll()
		}
		return m, nil

	case "g":
		// Go to top
		m.selectedIndex = 0
		m.listOffset = 0
		return m, nil

	case "G":
		// Go to bottom
		if len(m.agents) > 0 {
			m.selectedIndex = len(m.agents) - 1
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

	case "enter", " ":
		// Select agent - emit message for root to switch to detail view
		if m.SelectedAgent() != nil {
			return m, func() tea.Msg {
				return SelectAgentMsg{Agent: m.SelectedAgent()}
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
	// Forward navigation keys to project viewport
	switch msg.String() {
	case "j", "k", "down", "up", "pgdown", "pgup", "home", "end":
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
	// Forward navigation keys to usage viewport
	switch msg.String() {
	case "j", "k", "down", "up", "pgdown", "pgup", "home", "end":
		return m, func() tea.Msg {
			return ViewportUpdateMsg{Target: "usage", Msg: msg}
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

// NewSessionRequestMsg is sent when user wants to create a new session
// Root will handle this by opening a modal
type NewSessionRequestMsg struct{}

// SelectAgentMsg is sent when user selects an agent
// Root will handle this by switching to AgentDetailsView
type SelectAgentMsg struct {
	Agent *domain.Agent
}
