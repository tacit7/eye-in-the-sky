package overview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/messages"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/shared"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
)

// Model is the OverviewView Bubble Tea model
// Owns agent list state and manages top-level tabs
type Model struct {
	// Data access (injected by root)
	dataClient DataClient

	// Agent list state (owned by this view)
	agents        []domain.Agent
	selectedIndex int
	listOffset    int
	showAll       bool // Show all agents or only active ones

	// Tab navigation
	tabs components.TabsModel

	// UI dimensions
	width  int
	height int

	// Status message
	statusMsg string

	// Styles (injected from parent) - actual type is *app.Styles
	styles shared.Styles

	// Viewport provider for tab content (injected from root)
	viewports shared.ViewportProvider
}

// DataClient interface for data access
// Root will inject an implementation that wraps the database layer
type DataClient interface {
	LoadAgents() ([]domain.Agent, error)
}

// ErrMsg is sent when an error occurs
type ErrMsg struct {
	Error error
}

// New creates a new OverviewView model
func New(dataClient DataClient, styles shared.Styles, viewports shared.ViewportProvider) Model {
	// Create tabs for overview view
	// Tab indices: 0=Agents, 1=Project, 2=Claude, 3=Usage, 4=Config
	tabs := components.NewTabsModel(
		[]string{"[O]verview", "[P]roject", "[C]laude", "[T]oken Usage", "[K]eybindings"},
		"#00ADD8", // Active color (placeholder, should come from styles)
		"#FFFFFF", // Text color (placeholder, should come from styles)
	)

	return Model{
		dataClient:    dataClient,
		styles:        styles,
		viewports:     viewports,
		tabs:          tabs,
		selectedIndex: 0,
		listOffset:    0,
		showAll:       false,
		agents:        []domain.Agent{},
	}
}

// Init initializes the overview model
// Follows Bubble Tea convention: Init() returns Cmd that loads data
func (m Model) Init() tea.Cmd {
	return m.loadAgentsCmd()
}

// SetSize updates the view dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SelectedAgent returns the currently selected agent
// Root calls this when transitioning to AgentDetailsView
func (m *Model) SelectedAgent() *domain.Agent {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.agents) {
		return &m.agents[m.selectedIndex]
	}
	return nil
}

// GetVisibleAgents returns the agents to display based on showAll filter
func (m *Model) GetVisibleAgents() []domain.Agent {
	if m.showAll {
		return m.agents
	}

	// Filter for active, working, and idle agents
	var visible []domain.Agent
	for _, agent := range m.agents {
		if agent.Status == "active" || agent.Status == "working" || agent.Status == "idle" {
			visible = append(visible, agent)
		}
	}
	return visible
}

// adjustListScroll adjusts the list offset to keep selected item visible
func (m *Model) adjustListScroll() {
	// Calculate available height for content
	// Header (3 lines) + Tabs (2 lines) + Footer (1 line) = 6 lines overhead
	visibleHeight := m.height - 6
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Keep selected index visible
	if m.selectedIndex < m.listOffset {
		m.listOffset = m.selectedIndex
	}
	if m.selectedIndex >= m.listOffset+visibleHeight {
		m.listOffset = m.selectedIndex - visibleHeight + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

// loadAgentsCmd loads agents from the data client
func (m *Model) loadAgentsCmd() tea.Cmd {
	return func() tea.Msg {
		agents, err := m.dataClient.LoadAgents()
		if err != nil {
			return ErrMsg{Error: err}
		}
		return messages.AgentsLoadedMsg{Agents: agents}
	}
}
