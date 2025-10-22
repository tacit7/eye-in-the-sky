package app

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// ViewMode represents the current view state
type ViewModeRefactored int

const (
	ViewListRefactored ViewModeRefactored = iota
	ViewDetailRefactored
)

// BoundsRefactored represents screen coordinates and dimensions
type BoundsRefactored struct {
	X, Y, W, H int
}

// ListLayoutRefactored tracks the layout of the list view for click hit testing
type ListLayoutRefactored struct {
	HeaderH int              // lines used by renderHeader
	TabsH   int              // lines used by listTabs box
	FooterH int              // lines used by renderFooter
	Content BoundsRefactored // bordered content box
	RowH    int              // height per row
}

// ModelRefactored represents the application state with clean architecture
type ModelRefactored struct {
	// Data access layer
	data *DataClient

	// Configuration
	config Config
	keys   KeyBindings
	theme  Theme
	styles Styles

	// Help system
	help     help.Model
	showHelp bool

	// Window focuser
	windowFocuser util.WindowFocuser

	// Claude binary path
	claudePath string

	// Markdown renderer
	mdRenderer *glamour.TermRenderer

	// Tabs for detail view
	tabs components.TabsModel

	// Tabs for list view
	listTabs components.TabsModel

	// View state
	currentView ViewModeRefactored
	showAll     bool // Show all agents or only active ones

	// Agent list state (using domain types)
	agents        []domain.Agent
	selectedIndex int
	listOffset    int

	// Agent detail state (using domain types)
	selectedAgent  *domain.Agent
	detailOffset   int
	actions        []domain.Action
	commits        []domain.Commit
	notes          []domain.Note
	tasks          []domain.Task
	logs           []Log // Keep Log as local type for now
	sessionMetrics []domain.SessionMetric
	projectTickets []domain.Task // Tickets for current agent's project

	// Overview state
	allSessionMetrics []domain.SessionMetric  // All metrics from all agents
	monthlyCosts      []*domain.SessionMetric // Monthly costs with timestamps

	// Per-view scroll state
	tasksIndex           int
	tasksOffset          int
	commitsIndex         int
	commitsOffset        int
	notesIndex           int
	notesOffset          int
	logsIndex            int
	logsOffset           int
	actionsIndex         int
	actionsOffset        int
	projectTicketsIndex  int
	projectTicketsOffset int

	// Right pane scroll state
	rightPaneOffset int

	// Task loading state
	taskState TaskStateType
	taskError error

	// UI dimensions
	width  int
	height int

	// Polling
	lastRefresh time.Time
	err         error

	// Status message
	statusMsg string

	// CCUsage database connection (keep for now)
	ccusageDB *db.CCUsageDB

	// Cached ccusage data
	ccusageDaily    []api.DailyReport
	ccusageSessions []api.SessionReport
	ccusageMonthly  []api.MonthlyReport
	ccusageBlock    *api.ActiveBlockReport
	ccusageCosts    *api.CostSummary

	// CCUsage sync state
	ccusageSyncing    bool
	ccusageSyncStatus string
	ccusageEntryCount int
	lastCCUsageSync   time.Time

	// Layout tracking for click hit testing
	listLayout   ListLayoutRefactored
	lastClickAt  time.Time
	lastClickRow int

	// Project information
	projectInfo *ProjectInfo

	// Project view data
	projectTasks           []ProjectTask
	projectMDFiles         []ProjectFile
	claudeMDContent        string
	projectTasksIndex      int
	projectMDFilesIndex    int
	projectSelectedSection int // 0: tasks, 1: CLAUDE.md, 2: .md files
	projectTasksOffset     int
	projectMDFilesOffset   int
}

// NewModelRefactored creates a new model with clean architecture
func NewModelRefactored(db *sql.DB, config Config) *ModelRefactored {
	// Create data client with all stores
	dataClient := NewDataClient(db)

	// Initialize the model
	model := &ModelRefactored{
		data:        dataClient,
		config:      config,
		currentView: ViewListRefactored,
		showAll:     false,
		agents:      []domain.Agent{},
		tasks:       []domain.Task{},
		commits:     []domain.Commit{},
		notes:       []domain.Note{},
		actions:     []domain.Action{},
		width:       80,
		height:      24,
	}

	// Initialize themes and styles
	model.theme = GetTheme(config.Theme)
	model.styles = GetStyles(model.theme)

	// Initialize key bindings
	model.keys = GetKeyBindings()

	// Initialize help
	model.help = help.New()
	model.help.ShowAll = false

	// Initialize tabs
	model.tabs = components.NewTabsModel([]string{
		"[O]verview",
		"[C]ommits",
		"[L]ogs",
		"[N]otes",
		"[A]ctions",
		"[T]asks",
		"[P] Tickets",
		"[U]sage",
		"[Y] Project",
		"[S]ession",
	}, model.styles.Primary, model.styles.Subtle)

	model.listTabs = components.NewTabsModel([]string{
		"[A]gents",
		"[M]etrics",
		"[P]roject",
	}, model.styles.Primary, model.styles.Subtle)

	// Initialize markdown renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(model.width-4),
	)
	if err == nil {
		model.mdRenderer = renderer
	}

	// Initialize window focuser
	model.windowFocuser = util.NewWindowFocuser()

	// Find claude binary
	model.claudePath = findClaudeBinary()

	return model
}

// Init initializes the model
func (m *ModelRefactored) Init() tea.Cmd {
	// Load initial agents
	return tea.Batch(
		loadAgentsCmd(m.data.Agents),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return RefreshMsg{}
		}),
	)
}

// Update handles messages
func (m *ModelRefactored) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case AgentsLoadedMsg:
		m.agents = msg.Agents
		// After agents loaded, load task counts
		return m, loadTaskCountsCmd(m.data.Tasks, m.agents)

	case TaskCountsLoadedMsg:
		// Update agent task counts
		for i := range m.agents {
			for _, count := range msg.Counts {
				if m.agents[i].ID == count.AgentID {
					// This needs a new field in domain.Agent or separate tracking
					// For now, we'll need to track this separately
					break
				}
			}
		}
		return m, nil

	case TasksLoadedMsg:
		m.tasks = msg.Tasks
		m.taskState = TaskLoaded
		return m, nil

	case CommitsLoadedMsg:
		m.commits = msg.Commits
		return m, nil

	case NotesLoadedMsg:
		m.notes = msg.Notes
		return m, nil

	case ActionsLoadedMsg:
		m.actions = msg.Actions
		return m, nil

	case MetricsLoadedMsg:
		m.sessionMetrics = msg.Metrics
		return m, nil

	case ErrMsg:
		m.err = msg.Error
		m.statusMsg = msg.Error.Error()
		return m, nil

	case RefreshMsg:
		// Refresh agents periodically
		return m, tea.Batch(
			loadAgentsCmd(m.data.Agents),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
				return RefreshMsg{}
			}),
		)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			// Manual refresh
			return m, loadAgentsCmd(m.data.Agents)

		case "enter":
			// Select agent for detail view
			if m.currentView == ViewListRefactored && m.selectedIndex < len(m.agents) {
				m.selectedAgent = &m.agents[m.selectedIndex]
				m.currentView = ViewDetailRefactored
				// Load agent details
				return m, tea.Batch(
					loadTasksCmd(m.data.Tasks, m.selectedAgent.ID, 0, 0),
					loadCommitsCmd(m.data.Commits, m.selectedAgent.ID),
					loadNotesCmd(m.data.Notes, m.selectedAgent.ID),
					loadActionsCmd(m.data.Actions, m.selectedAgent.ID),
					loadMetricsByAgentCmd(m.data.Metrics, m.selectedAgent.ID, 10),
				)
			}
			return m, nil

		case "esc":
			// Go back to list view
			if m.currentView == ViewDetailRefactored {
				m.currentView = ViewListRefactored
				m.selectedAgent = nil
				return m, loadAgentsCmd(m.data.Agents)
			}
			return m, nil

		case "j", "down":
			// Move selection down
			if m.currentView == ViewListRefactored {
				if m.selectedIndex < len(m.agents)-1 {
					m.selectedIndex++
				}
			}
			return m, nil

		case "k", "up":
			// Move selection up
			if m.currentView == ViewListRefactored {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the UI
func (m *ModelRefactored) View() string {
	if m.currentView == ViewDetailRefactored {
		return m.renderDetailView()
	}
	return m.renderListView()
}

// renderListView renders the agent list
func (m *ModelRefactored) renderListView() string {
	var lines []string

	// Header
	lines = append(lines, m.styles.Primary.Render("Eye in the Sky - Agents"))
	lines = append(lines, "")

	// Agent list
	for i, agent := range m.agents {
		prefix := "  "
		if i == m.selectedIndex {
			prefix = "> "
		}

		status := m.getStatusStyle(agent.Status).Render(agent.Status)
		line := fmt.Sprintf("%s%-8s %s", prefix, agent.ID, status)

		if agent.CurrentTask != "" {
			line += " - " + agent.CurrentTask
		}

		lines = append(lines, line)
	}

	// Help
	lines = append(lines, "")
	lines = append(lines, m.styles.Subtle.Render("[j/k] Navigate  [enter] Details  [r] Refresh  [q] Quit"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderDetailView renders the agent detail view
func (m *ModelRefactored) renderDetailView() string {
	if m.selectedAgent == nil {
		return "No agent selected"
	}

	var lines []string

	// Header
	lines = append(lines, m.styles.Primary.Render(fmt.Sprintf("Agent: %s", m.selectedAgent.ID)))
	lines = append(lines, "")

	// Agent info
	lines = append(lines, fmt.Sprintf("Status: %s", m.getStatusStyle(m.selectedAgent.Status).Render(m.selectedAgent.Status)))
	lines = append(lines, fmt.Sprintf("Current Task: %s", m.selectedAgent.CurrentTask))
	lines = append(lines, "")

	// Tasks
	lines = append(lines, m.styles.Secondary.Render("Tasks:"))
	for _, task := range m.tasks {
		lines = append(lines, fmt.Sprintf("  - [%s] %s", task.Priority, task.Description))
	}

	// Help
	lines = append(lines, "")
	lines = append(lines, m.styles.Subtle.Render("[esc] Back  [q] Quit"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// getStatusStyle returns the style for a status
func (m *ModelRefactored) getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "active":
		return m.styles.Active
	case "working":
		return m.styles.Working
	case "idle":
		return m.styles.Idle
	case "completed":
		return m.styles.Completed
	case "failed":
		return m.styles.Failed
	default:
		return m.styles.Subtle
	}
}

// RefreshMsg is sent periodically to refresh data
type RefreshMsg struct{}