package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
	"github.com/tacit7/eye-in-the-sky/internal/ui/services"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
	"github.com/tacit7/eye-in-the-sky/internal/ui/viewmodel"
)

// ViewType represents the current view state
type ViewType int

const (
	ViewList ViewType = iota
	ViewDetail
)

// TaskStateType represents the loading state of tasks
type TaskStateType int

const (
	TaskLoading TaskStateType = iota
	TaskLoaded
	TaskError
)

// TasksErrorMsg is sent when tasks fail to load
type TasksErrorMsg struct {
	Error error
}

// Bounds represents screen coordinates and dimensions
type Bounds struct {
	X, Y, W, H int
}

// ListLayout tracks the layout of the list view for click hit testing
type ListLayout struct {
	HeaderH int    // lines used by renderHeader
	TabsH   int    // lines used by listTabs box
	FooterH int    // lines used by renderFooter
	Content Bounds // bordereded content box
	RowH    int    // height per row; with your current render this is 1
}

// ViewRenderer is a function that renders a view
type ViewRenderer func(m *Model) string

// Model represents the application state
type Model struct {
	// Data access layer
	data *DataClient

	// Configuration
	config      Config
	keys        KeyBindings
	theme       Theme
	styles      Styles

	// View renderers map
	renderers map[ViewType]ViewRenderer

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
	currentView ViewType
	showAll     bool // Show all agents or only active ones

	// Agent list state
	agents        []domain.Agent
	selectedIndex int
	listOffset    int

	// Agent detail state
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
	allSessionMetrics  []domain.SessionMetric // All metrics from all agents
	monthlyCosts       []*domain.SessionMetric // Monthly costs with timestamps

	// Per-view scroll state
	tasksIndex    int
	tasksOffset   int
	taskIndex     int  // Current selected task in detail view
	commitsIndex  int
	commitsOffset int
	notesIndex    int
	notesOffset   int
	logsIndex     int
	logsOffset    int
	actionsIndex  int
	actionsOffset int
	projectTicketsIndex int
	projectTicketsOffset int

	// Right pane scroll state
	rightPaneOffset int

	// Task loading state
	taskState TaskStateType
	taskError error

	// UI dimensions
	width  int
	height int

	// Layout management
	layoutManager *LayoutManager

	// Polling
	lastRefresh time.Time
	lastUpdate  time.Time
	isLoading   bool
	err         error

	// Status message
	statusMsg  string
	statusTime time.Time

	// CCUsage database connection
	ccusageDB *db.CCUsageDB

	// Cached ccusage data
	ccusageDaily     []api.DailyReport
	ccusageSessions  []api.SessionReport
	ccusageMonthly   []api.MonthlyReport
	ccusageBlock     *api.ActiveBlockReport
	ccusageCosts     *api.CostSummary

	// CCUsage sync state
	ccusageSyncing    bool
	ccusageSyncStatus string
	ccusageEntryCount int
	lastCCUsageSync   time.Time

	// Layout tracking for click hit testing
	listLayout    ListLayout
	lastClickAt   time.Time
	lastClickRow  int

	// Project information
	projectInfo *ProjectInfo

	// Project view data
	projectTasks      []ProjectTask
	projectMDFiles    []ProjectFile
	claudeMDContent   string
	projectTasksIndex int
	projectMDFilesIndex int
	projectSelectedSection int    // 0: tasks, 1: CLAUDE.md, 2: .md files
	projectSection        string  // Current section in project tab
	projectTasksOffset    int
	projectMDFilesOffset  int

	// Usage view viewport
	usageViewport viewport.Model

	// Usage service and caching
	usageSvc           *services.UsageService
	cachedUsageVM      *viewmodel.UsageViewModel
	cachedUsageRender  string
	usageDirty         bool
	widthUnchanged     bool
	lastRenderWidth    int

	// Claude tab state
	claudeFiles         []ClaudeFile
	claudeSelectedIndex int
	claudeContent       string
	claudeViewport      viewport.Model
	claudeValidStatus   string
	claudeShowingContent bool
}

// Type aliases for backward compatibility during migration
type Agent = domain.Agent
type Action = domain.Action
type Commit = domain.Commit
type Note = domain.Note
type Task = domain.Task
type SessionMetric = domain.SessionMetric
type Log = domain.Log

// ProjectTask represents a TaskWarrior task
type ProjectTask struct {
	UUID        string
	Description string
	Status      string
	Priority    string
	DueDate     time.Time
}

// ProjectFile represents a markdown file in the project
type ProjectFile struct {
	Name    string // Just the filename
	Path    string // Full path relative to project root
	Content string // File contents
}

// ClaudeFile represents a config file in ~/.claude
type ClaudeFile struct {
	Name    string // Filename (e.g., "settings.json")
	Path    string // Full path
	IsDir   bool   // True if it's a directory
	ModTime time.Time
}

// TaskAnnotation is kept local as it's referenced by domain.Task
type TaskAnnotation = domain.TaskAnnotation

// Styles holds all lipgloss styles
type Styles struct {
	Active       lipgloss.Style
	Working      lipgloss.Style
	Idle         lipgloss.Style
	Stale        lipgloss.Style
	Unknown      lipgloss.Style
	Completed    lipgloss.Style
	Failed       lipgloss.Style
	Primary      lipgloss.Style
	Secondary    lipgloss.Style
	Border       lipgloss.Style
	Title        lipgloss.Style
	Text         lipgloss.Style
	Subtle       lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Error        lipgloss.Style
	ContentBox   lipgloss.Style
	InfoBox      lipgloss.Style
	ErrorBox     lipgloss.Style
	Header       lipgloss.Style
	Footer       lipgloss.Style
	StatusBar    lipgloss.Style
	KeyHelp      lipgloss.Style
	Selected     lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	SectionTitle lipgloss.Style
	Git          lipgloss.Style
	Code         lipgloss.Style
	Highlight    lipgloss.Style
	Bold         lipgloss.Style
}

// NewModel creates a new application model
func NewModel(db *sql.DB, ccusageDB *db.CCUsageDB) (*Model, error) {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	keys, err := LoadKeyBindings()
	if err != nil {
		return nil, err
	}

	theme, err := LoadTheme(config.Theme)
	if err != nil {
		return nil, err
	}

	// Create styles from theme
	styles := createStyles(theme)

	// Create help model
	helpModel := NewHelp()

	// Create window focuser
	windowFocuser := util.NewWindowFocuser()

	// Resolve claude path
	claudePath, err := ResolveClaudePath(config.ClaudePath)
	if err != nil {
		// Don't fail startup, just log warning and leave empty
		log.Printf("Warning: %v", err)
		claudePath = ""
	}

	// Create tabs for agent detail view (← is back arrow, unicode 8678)
	tabs := components.NewTabsModel(
		[]string{"← Back", "[A]gent View", "[C]ommits", "[L]ogs", "[N]otes", "[A]ctions", "[T]asks", "[P]rojects"},
		theme.Colors.Active,
		theme.Colors.Text,
	)

	// Create tabs for overview (agent list)
	listTabs := components.NewTabsModel(
		[]string{"[O]verview", "[P]roject", "[C]laude", "[T]oken Usage"},
		theme.Colors.Active,
		theme.Colors.Text,
	)

	// Create markdown renderer (initialize with default width, will be updated on resize)
	mdRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		// Don't fail startup, just log warning
		log.Printf("Warning: Failed to create markdown renderer: %v", err)
		mdRenderer = nil
	}

	// Create viewport for usage tab (will be resized on window size updates)
	usageViewport := viewport.New(80, 20)

	// Create usage service with system clock
	usageSvc := services.NewUsageService(services.SystemClock{})

	m := &Model{
		data:          NewDataClient(db),
		ccusageDB:     ccusageDB,
		config:        config,
		keys:          keys,
		theme:         theme,
		styles:        styles,
		help:          helpModel,
		showHelp:      false,
		windowFocuser: windowFocuser,
		claudePath:    claudePath,
		mdRenderer:    mdRenderer,
		tabs:          tabs,
		listTabs:      listTabs,
		usageViewport: usageViewport,
		currentView:   ViewList,
		showAll:       config.ShowAllAgents,
		agents:        []Agent{},
		ccusageSyncing: false,
		lastClickRow:  -1, // Initialize to -1 so first click doesn't trigger double-click
		usageSvc:      usageSvc,
		usageDirty:    true, // Start dirty to force initial build
	}

	// Initialize view renderers map
	m.renderers = map[ViewType]ViewRenderer{
		ViewList:   (*Model).renderListView,
		ViewDetail: (*Model).renderDetail,
	}

	// Detect project information at startup
	m.projectInfo = DetectProject()

	// Load project data if detected
	if m.projectInfo != nil {
		// Load markdown files and CLAUDE.md
		m.projectMDFiles = LoadProjectMarkdownFiles(m.projectInfo.GitRootPath)
		m.claudeMDContent = LoadClaudeMDFile(m.projectInfo.GitRootPath)
	}

	// Initial load will happen in Init()
	return m, nil
}

// createStyles creates lipgloss styles from theme
func createStyles(theme Theme) Styles {
	return Styles{
		Active:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Active)),
		Working:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Working)),
		Idle:      lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Idle)),
		Stale:     lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Stale)),
		Unknown:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Unknown)),
		Completed: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Completed)),
		Failed:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Failed)),
		Primary:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Primary)),
		Secondary: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Secondary)),
		Border:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Border)),
		Title:     lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Title)).Bold(true),
		Text:      lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Text)),
		Subtle:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Subtle)),

		// Additional styles for new views
		Success:      lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Active)),
		Warning:      lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Working)),
		Error:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Failed)),
		ContentBox:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(theme.Colors.Border)),
		InfoBox:      lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(theme.Colors.Primary)),
		ErrorBox:     lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color(theme.Colors.Failed)),
		Header:       lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color(theme.Colors.Title)),
		Footer:       lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color(theme.Colors.Text)),
		StatusBar:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Primary)),
		KeyHelp:      lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Subtle)),
		Selected:     lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color(theme.Colors.Active)),
		Label:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Secondary)),
		Value:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Text)),
		SectionTitle: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Primary)).Bold(true),
		Git:          lipgloss.NewStyle().Foreground(lipgloss.Color("#f97316")),
		Code:         lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color(theme.Colors.Text)),
		Highlight:    lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color(theme.Colors.Active)).Bold(true),
		Bold:         lipgloss.NewStyle().Bold(true),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		loadAgentsCmd(m.data.Agents), // Load initial agents
		m.tickCmd(),                   // Start ticker
	)
}

// tickCmd returns a command that triggers a refresh
func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(m.config.RefreshDuration(), func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// tickMsg is sent on each refresh interval
type tickMsg time.Time

// initCCUsageMsg triggers CCUsage database initialization
type initCCUsageMsg struct{}

// RefreshUsageMsg triggers a usage tab content refresh
type RefreshUsageMsg struct{}

// loadAgents loads agents from database
func (m *Model) loadAgents() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the data store to load agents
	agents, err := m.data.Agents.LoadAgents(ctx)
	if err != nil {
		return err
	}

	m.agents = agents
	m.lastRefresh = time.Now()

	// Load task counts for each agent from TaskWarrior
	m.loadAgentTaskCounts()

	// Adjust selected index if needed
	if m.selectedIndex >= len(m.agents) && len(m.agents) > 0 {
		m.selectedIndex = len(m.agents) - 1
	}

	// Also load all session metrics for the Usage tab
	if err := m.loadAllSessionMetrics(); err != nil {
		// Don't fail if metrics fail to load
		return nil
	}

	// Load monthly costs for the Usage tab
	if err := m.loadMonthlyCosts(); err != nil {
		// Don't fail if monthly costs fail to load
		return nil
	}

	m.usageDirty = true // Mark cache dirty after agent load
	return nil
}

// loadAgentDetails loads details for the currently selected agent
func (m *Model) loadAgentDetails() error {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.agents) {
		return nil
	}

	agent := m.agents[m.selectedIndex]
	m.selectedAgent = &agent

	// Load actions
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	actions, err := m.data.Actions.LoadByAgent(ctx, domain.AgentID(agent.ID), 50)
	cancel()
	if err != nil {
		return err
	}
	m.actions = actions

	// Load commits - includes commits from child agents (hierarchical visibility)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	commits, err := m.data.Commits.LoadByAgentHierarchy(ctx2, domain.AgentID(agent.ID), 20)
	cancel2()
	if err != nil {
		return err
	}
	m.commits = commits

	// Load notes for current agent
	ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	notes, err := m.data.Notes.LoadByAgent(ctx3, domain.AgentID(agent.ID))
	cancel3()
	if err != nil {
		// Don't fail if notes loading fails
		m.err = err
		notes = []Note{}
	}
	m.notes = notes

	// Load session metrics
	if err := m.loadSessionMetrics(); err != nil {
		// Don't fail if metrics loading fails, just log it
		m.err = err
	}

	return nil
}

// loadLogs loads logs for the current agent session
func (m *Model) loadLogs() error {
	if m.selectedAgent == nil || m.selectedAgent.SessionID == "" {
		m.logs = []Log{}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logs, err := m.data.Logs.LoadBySession(ctx, m.selectedAgent.SessionID, 100)
	if err != nil {
		return err
	}

	m.logs = logs

	// Reset scroll position if needed
	if len(m.logs) > 0 && m.logsIndex >= len(m.logs) {
		m.logsIndex = len(m.logs) - 1
	}

	return nil
}

// loadMonthlyCosts loads all session metrics from the current month with timestamps
func (m *Model) loadMonthlyCosts() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	metrics, err := m.data.Metrics.LoadMonthly(ctx, now.Year(), now.Month())
	if err != nil {
		return err
	}

	// Convert to pointer slice for compatibility
	ptrMetrics := make([]*SessionMetric, len(metrics))
	for i := range metrics {
		ptrMetrics[i] = &metrics[i]
	}

	m.monthlyCosts = ptrMetrics
	return nil
}

// loadAllSessionMetrics loads the latest session metrics for all agents
func (m *Model) loadAllSessionMetrics() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metrics, err := m.data.Metrics.LoadAll(ctx)
	if err != nil {
		return err
	}

	m.allSessionMetrics = metrics
	return nil
}

// loadSessionMetrics loads session metrics for the current agent
func (m *Model) loadSessionMetrics() error {
	if m.selectedAgent == nil {
		m.sessionMetrics = []SessionMetric{}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metrics, err := m.data.Metrics.LoadByAgent(ctx, domain.AgentID(m.selectedAgent.ID), 10)
	if err != nil {
		return err
	}

	m.sessionMetrics = metrics
	return nil
}

// syncCCUsageData performs on-demand sync of ccusage data
func (m *Model) syncCCUsageData() error {
	if m.ccusageDB == nil {
		return nil
	}

	log.Println("[MODEL] Starting on-demand CCUsage sync...")
	syncMgr := parser.NewSyncManager(m.ccusageDB)
	if err := syncMgr.Sync(); err != nil {
		log.Printf("[MODEL] Warning: CCUsage sync failed: %v", err)
		return fmt.Errorf("sync failed: %w", err)
	}
	log.Println("[MODEL] CCUsage sync completed")
	return nil
}

// loadCCUsageData loads Claude Code usage data from ccusage database
func (m *Model) loadCCUsageData() error {
	if m.ccusageDB == nil {
		return nil
	}

	// Get entry count
	count, err := m.ccusageDB.GetEntryCount()
	if err != nil {
		m.ccusageEntryCount = 0
	} else {
		m.ccusageEntryCount = count
	}

	// If no entries, don't bother querying
	if m.ccusageEntryCount == 0 {
		m.ccusageDaily = []api.DailyReport{}
		m.ccusageSessions = []api.SessionReport{}
		m.ccusageMonthly = []api.MonthlyReport{}
		m.ccusageBlock = nil
		m.ccusageCosts = nil
		return nil
	}

	// Load daily usage (all days from start)
	daily, err := api.GetDailyUsageReport(m.ccusageDB, 0)
	if err != nil {
		return fmt.Errorf("failed to load daily usage: %w", err)
	}
	m.ccusageDaily = daily

	// Load sessions
	sessions, err := api.GetSessionUsageReport(m.ccusageDB, 10)
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}
	m.ccusageSessions = sessions

	// Load monthly summaries for all months with data
	monthly, err := api.GetAllMonthlyReports(m.ccusageDB)
	if err != nil {
		return fmt.Errorf("failed to load monthly reports: %w", err)
	}
	m.ccusageMonthly = monthly

	// Load active block
	block, err := api.GetActiveBlockReport(m.ccusageDB)
	if err != nil {
		return fmt.Errorf("failed to load active block: %w", err)
	}
	m.ccusageBlock = block

	// Load cost summary
	costs, err := api.GetTotalCostSummary(m.ccusageDB, 30)
	if err != nil {
		return fmt.Errorf("failed to load cost summary: %w", err)
	}
	m.ccusageCosts = costs

	m.lastCCUsageSync = time.Now()
	m.usageDirty = true // Mark cache dirty after data load
	return nil
}

// loadTasks loads tasks for the current agent
func (m *Model) loadTasks() error {
	if m.selectedAgent == nil {
		m.tasks = []Task{}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := m.data.Tasks.LoadByAgent(ctx, domain.AgentID(m.selectedAgent.ID), 100, 0)
	if err != nil {
		return err
	}

	m.tasks = tasks
	return nil
}

// loadProjectTickets loads project tickets for the current agent
func (m *Model) loadProjectTickets() error {
	if m.selectedAgent == nil || m.selectedAgent.ProjectName == "" {
		m.projectTickets = []Task{}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Load tickets by project tag
	tasks, err := m.data.Tasks.LoadByProject(ctx, m.selectedAgent.ProjectName, 100)
	if err != nil {
		return err
	}

	m.projectTickets = tasks
	return nil
}

// loadTasksCmd returns a command to load tasks
func (m *Model) loadTasksCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.loadTasks(); err != nil {
			return TasksErrorMsg{Error: err}
		}
		return nil
	}
}

// loadProjectTicketsCmd returns a command to load project tickets
func (m *Model) loadProjectTicketsCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.loadProjectTickets(); err != nil {
			return TasksErrorMsg{Error: err}
		}
		return nil
	}
}

// loadTabData loads data for the currently active tab
func (m *Model) loadTabData() error {
	if m.selectedAgent == nil {
		return nil
	}

	switch m.tabs.ActiveIndex {
	case 3: // Logs tab
		return m.loadLogs()
	case 6: // Tasks tab
		return m.loadTasks()
	case 7: // Projects tab
		return m.loadProjectTickets()
	default:
		// Other tabs (Agent View, Commits, Notes, Actions) already loaded by loadAgentDetails
		return nil
	}
}

// SelectedAgent returns the currently selected agent
func (m *Model) SelectedAgent() *Agent {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.agents) {
		return nil
	}
	return &m.agents[m.selectedIndex]
}

// loadAgentTaskCounts queries TaskWarrior for task counts for each agent
func (m *Model) loadAgentTaskCounts() {
	// Check if taskwarrior is installed
	if _, err := exec.LookPath("task"); err != nil {
		// TaskWarrior not installed, skip
		debugf("TaskWarrior not found in PATH")
		return
	}

	// Get all tasks in one export
	cmd := exec.Command("task", "export")
	output, err := cmd.Output()
	if err != nil {
		// Failed to get tasks, skip
		debugf("Failed to export tasks: %v", err)
		return
	}

	// Parse JSON output
	var tasks []map[string]interface{}
	if err := json.Unmarshal(output, &tasks); err != nil {
		debugf("Failed to parse task JSON: %v", err)
		return
	}

	debugf("Found %d total tasks from TaskWarrior", len(tasks))

	// Count tasks for each agent
	for i := range m.agents {
		count := 0
		var searchTag string

		// Determine the tag to search for
		if m.agents[i].ParentAgentID != "" {
			// Subagent: search for +subagent_<agent_id>
			searchTag = fmt.Sprintf("+subagent_%s", strings.ReplaceAll(string(m.agents[i].ID), "-", "_"))
			debugf("Agent %s (subagent): searching for tag %s", m.agents[i].ID, searchTag)
		} else if m.agents[i].SessionID != "" {
			// Parent agent: search for +session_<session_id>
			searchTag = fmt.Sprintf("+session_%s", strings.ReplaceAll(m.agents[i].SessionID, "-", "_"))
			debugf("Agent %s (parent): searching for tag %s", m.agents[i].ID, searchTag)
		} else {
			// No session or parent, skip
			debugf("Agent %s: no session or parent, skipping", m.agents[i].ID)
			continue
		}

		// Count tasks with matching tag
		for _, task := range tasks {
			// Check if task is pending first
			if status, ok := task["status"].(string); ok && status == "pending" {
				// Check tags array for matching tag
				if tags, ok := task["tags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagStr, ok := tag.(string); ok {
							// Remove the + prefix from searchTag for comparison
							cleanSearchTag := strings.TrimPrefix(searchTag, "+")
							if tagStr == cleanSearchTag {
								count++
								break // Found matching tag, count this task
							}
						}
					}
				}
			}
		}

		m.agents[i].TaskCount = count
		if count > 0 {
			debugf("Agent %s: found %d tasks", m.agents[i].ID, count)
		}
	}
}
