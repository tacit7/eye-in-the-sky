package overview

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
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
	agents         []domain.Agent
	selectedIndex  int
	listOffset     int
	showAll        bool // Show all agents or only active ones

	// Bubble Tea table component for agents list
	agentsTable table.Model

	// Header sorting state
	headerFocused  bool   // True when header is focused for sorting
	selectedColumn int    // Which column is selected in header (0-4: Status, Session, ID, Task, Source)
	sortField      string // Current sort field: "status", "session", "id", "task", "source"
	sortAscending  bool   // Sort order

	// Tab navigation
	tabs components.NavBar

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
	LoadAgents(showAll bool) ([]domain.Agent, error)
}

// ErrMsg is sent when an error occurs
type ErrMsg struct {
	Error error
}

// New creates a new OverviewView model
func New(dataClient DataClient, styles shared.Styles, viewports shared.ViewportProvider) Model {
	// Create tabs for overview view
	// Tab indices: 0=Agents, 1=Project, 2=Claude, 3=Usage, 4=Config
	tabs := components.NewNavBar(
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
	// Refresh table with new height
	if len(m.agents) > 0 {
		m.refreshAgentsTable()
	}
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
	// Always show all agents except deleted ones
	// The showAll flag is kept for potential future use
	var visible []domain.Agent
	for _, agent := range m.agents {
		if agent.Status != "deleted" {
			visible = append(visible, agent)
		}
	}
	return visible
}

// GetActiveTabIndex returns the currently active tab index
// Used by root to determine which viewport to update
func (m *Model) GetActiveTabIndex() int {
	return m.tabs.ActiveIndex
}

// sortAgents sorts the agents list by the current sort field and order
func (m *Model) sortAgents() {
	if m.sortField == "" {
		return
	}

	agents := m.agents
	field := m.sortField
	ascending := m.sortAscending

	// Custom sort based on field
	for i := 0; i < len(agents); i++ {
		for j := i + 1; j < len(agents); j++ {
			var swap bool
			switch field {
			case "status":
				if ascending {
					swap = agents[i].Status > agents[j].Status
				} else {
					swap = agents[i].Status < agents[j].Status
				}
			case "session":
				if ascending {
					swap = agents[i].SessionID > agents[j].SessionID
				} else {
					swap = agents[i].SessionID < agents[j].SessionID
				}
			case "id":
				if ascending {
					swap = string(agents[i].ID) > string(agents[j].ID)
				} else {
					swap = string(agents[i].ID) < string(agents[j].ID)
				}
			case "task":
				task1 := agents[i].FeatureDesc
				if task1 == "" {
					task1 = agents[i].CurrentTask
				}
				task2 := agents[j].FeatureDesc
				if task2 == "" {
					task2 = agents[j].CurrentTask
				}
				if ascending {
					swap = task1 > task2
				} else {
					swap = task1 < task2
				}
			case "source":
				if ascending {
					swap = string(agents[i].Source) > string(agents[j].Source)
				} else {
					swap = string(agents[i].Source) < string(agents[j].Source)
				}
			case "project":
				if ascending {
					swap = agents[i].ProjectName > agents[j].ProjectName
				} else {
					swap = agents[i].ProjectName < agents[j].ProjectName
				}
			case "lastlog":
				if ascending {
					swap = agents[i].LastLog > agents[j].LastLog
				} else {
					swap = agents[i].LastLog < agents[j].LastLog
				}
			}

			if swap {
				agents[i], agents[j] = agents[j], agents[i]
			}
		}
	}
	m.agents = agents
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
		agents, err := m.dataClient.LoadAgents(m.showAll)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return messages.AgentsLoadedMsg{Agents: agents}
	}
}

// createAgentsTable builds a table.Model from the current agents list
func (m *Model) createAgentsTable() table.Model {
	columns := []table.Column{
		{Title: "Status", Width: 10},
		{Title: "Last Log", Width: 20},
		{Title: "Session", Width: 12},
		{Title: "Task", Width: 40},
		{Title: "Project", Width: 20},
	}

	rows := []table.Row{}
	visibleAgents := m.GetVisibleAgents()

	for _, agent := range visibleAgents {
		// Get task description (prefer FeatureDesc, fallback to CurrentTask)
		task := agent.FeatureDesc
		if task == "" {
			task = agent.CurrentTask
		}
		if len(task) > 40 {
			task = task[:37] + "..."
		}

		// Format last log preview
		lastLog := agent.LastLog
		if len(lastLog) > 20 {
			lastLog = lastLog[:17] + "..."
		}

		rows = append(rows, table.Row{
			agent.Status,
			lastLog,
			agent.SessionID,
			task,
			agent.ProjectName,
		})
	}

	// Calculate table height (leave room for header, tabs, footer)
	tableHeight := m.height - 8
	if tableHeight < 5 {
		tableHeight = 5
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)

	// Use primary color from styles
	primaryStyle := m.styles.GetPrimary()
	primaryColor := primaryStyle.GetForeground()
	// Use primary color if available, otherwise default to cyan
	if primaryColor == nil {
		primaryColor = lipgloss.Color("#00ADD8")
	}
	s.Selected = s.Selected.Bold(true).Foreground(primaryColor)
	t.SetStyles(s)

	return t
}

// refreshAgentsTable rebuilds the agents table with current data
func (m *Model) refreshAgentsTable() {
	m.agentsTable = m.createAgentsTable()
	// Sync table cursor with selectedIndex
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.GetVisibleAgents()) {
		m.agentsTable.SetCursor(m.selectedIndex)
	}
}
