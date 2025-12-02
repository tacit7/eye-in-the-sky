package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
	"github.com/tacit7/eye-in-the-sky/internal/ui/keybindings"
	"github.com/tacit7/eye-in-the-sky/internal/ui/modal"
	"github.com/tacit7/eye-in-the-sky/internal/ui/services"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
	"github.com/tacit7/eye-in-the-sky/internal/ui/viewmodel"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/overview"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/shared"
)

// ViewType represents the current view state
type ViewType int

const (
	ViewList ViewType = iota
	ViewDetail
	ViewProjectDetail
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
	config Config
	theme  Theme
	styles Styles

	// View models (new modular architecture)
	overviewView        tea.Model // overview.Model
	agentDetailsView    tea.Model // agent_details.Model (stores as interface to avoid import issues)
	projectDetailsView  tea.Model // project.Model (stores as interface to avoid import issues)

	// View renderers map
	renderers map[ViewType]ViewRenderer

	// Help system (old help overlay deprecated - using new modal-based help)
	// help field removed
	showHelp bool

	// Window focuser
	windowFocuser util.WindowFocuser

	// Claude binary path
	claudePath string

	// Markdown renderer
	mdRenderer *glamour.TermRenderer

	// Tabs for detail view
	tabs components.NavBar

	// Tabs for list view
	listTabs components.NavBar

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
	actions         []domain.Action
	commits         []domain.Commit
	notes           []domain.Note
	tasks           []domain.Task
	logs            []Log // Keep Log as local type for now
	sessionMetrics  []domain.SessionMetric
	sessionContexts []domain.SessionContext
	projectTickets  []domain.Task // Tickets for current agent's project

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

	// Logs tab state for auto-refresh
	lastFetchedAt time.Time // Last timestamp for incremental log fetching

	// Task loading state
	taskState TaskStateType
	taskError error

	// UI dimensions
	width  int
	height int

	// Project detail view state
	projectTabsIndex int // Active tab index in project detail view

	// Project detail view table models (stored here to preserve state)
	projectAgentsTable table.Model
	projectNotesTable  table.Model
	projectTasksTable  table.Model
	projectAgentsIndex int // Selected agent index in project agents table
	projectNotesIndex  int // Selected note index in project notes table
	projectTasksIndex  int // Selected task index in project tasks table
	projectAgentsForceRefresh bool // Force refresh of agents table on next render

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
	ccusageLoading    bool   // Loading data from database
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
	projectMDFiles      []ProjectFile
	claudeMDContent     string
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
	claudeCurrentPath    string         // Current directory path in ~/.claude
	claudeFiles          []ClaudeFile
	claudeSelectedIndex  int
	claudeFilesViewport  viewport.Model // Scrollable file list viewport
	claudeContent        string
	claudeViewport       viewport.Model
	claudeValidStatus    string
	claudeShowingContent bool

	// Modal system
	modalManager         *modal.Modal
	noteModal            components.NoteModal
	taskAnnotationModal  components.TaskAnnotationModal
	keybindResolver      *keybindings.Resolver

	// Pending note context (for EDITOR workflow)
	pendingNoteAgentID    string
	pendingNoteSessionID  string
	pendingNoteProjectID  string

	// Config tab state
	keybindingsYAML     string // Current YAML content
	keybindingsEditing  bool   // Is user editing?
	keybindingsModified bool   // Has content changed?
	keybindingsEditBuf  string // Edit buffer for changes
	keybindingsViewport viewport.Model
	keybindingsError    string // Validation error message (empty if valid)

	// Project tab state and viewport
	projectViewport viewport.Model

	// Overview rendering cache (for pure, fast View() calls)
	overviewCache  string         // Cached rendered overview tab content
	overviewDirty  bool           // True if cache needs refresh
	overviewStyles OverviewStyles // Styles specific to overview tab
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
	Name     string    // Filename (e.g., "settings.json")
	Path     string    // Full path
	IsDir    bool      // True if it's a directory
	IsParent bool      // True if this is a ".." parent directory entry
	ModTime  time.Time
}

// TaskNote is kept local as it's referenced by domain.Task
type TaskNote = domain.TaskNote

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

// Interface methods for overview.Styles compatibility
func (s *Styles) RenderSuccess(text string) string {
	return s.Success.Render(text)
}

func (s *Styles) RenderWarning(text string) string {
	return s.Warning.Render(text)
}

func (s *Styles) RenderSubtle(text string) string {
	return s.Subtle.Render(text)
}

func (s *Styles) RenderError(text string) string {
	return s.Error.Render(text)
}

func (s *Styles) RenderPrimary(text string) string {
	return s.Primary.Render(text)
}

func (s *Styles) RenderSelected(text string) string {
	return s.Selected.Render(text)
}

// GetSelf returns the concrete Styles pointer for use with TableBuilder
// This allows views to use Styles through an interface but still access the concrete type
func (s *Styles) GetSelf() *Styles {
	return s
}

// Get* methods for components.Styles interface
func (s *Styles) GetSuccess() lipgloss.Style { return s.Success }
func (s *Styles) GetWarning() lipgloss.Style { return s.Warning }
func (s *Styles) GetSubtle() lipgloss.Style  { return s.Subtle }
func (s *Styles) GetError() lipgloss.Style   { return s.Error }
func (s *Styles) GetPrimary() lipgloss.Style { return s.Primary }
func (s *Styles) GetSelected() lipgloss.Style { return s.Selected }

// NewModel creates a new application model
func NewModel(db *sql.DB, ccusageDB *db.CCUsageDB) (*Model, error) {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	theme, err := LoadTheme(config.Theme)
	if err != nil {
		return nil, err
	}

	// Create styles from theme
	styles := createStyles(theme)

	// Note: Old help model removed - using new modal-based help system
	// The help modal is now opened via ? or Shift+H keys

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
	tabs := components.NewNavBar(
		[]string{"← Back", "[O]verview", "[T]asks", "[L]ogs", "[C]ommits", "[N]otes", "[S]ession Context"},
		theme.Colors.Active,
		theme.Colors.Text,
	)

	// Create tabs for overview (agent list)
	listTabs := components.NewNavBar(
		[]string{"[O]verview", "[P]roject", "[C]laude", "[T]oken Usage", "[K]eybindings"},
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

	// Create viewport for Claude tab (will be resized on window size updates)
	claudeViewport := viewport.New(80, 20)
	claudeFilesViewport := viewport.New(40, 20)

	// Create viewport for Config tab (will be resized on window size updates)
	keybindingsViewport := viewport.New(80, 20)

	// Create viewport for Project tab (will be resized on window size updates)
	projectViewport := viewport.New(80, 20)

	// Create usage service with system clock
	usageSvc := services.NewUsageService(services.SystemClock{})

	// Load keybindings - use defaults from code, skip YAML file for now
	var keybindResolver *keybindings.Resolver
	var keybindingsError string
	resolver, err := keybindings.LoadKeybindingsDefaults()
	if err != nil {
		log.Printf("Warning: Failed to load default keybindings: %v\n", err)
		resolver = &keybindings.Resolver{}
		keybindingsError = err.Error()
	} else {
		keybindingsError = "" // No error if defaults loaded successfully
	}
	keybindResolver = resolver

	// Initialize modal manager
	modalManager := modal.New()

	// Create overview view (new modular architecture)
	// Will be initialized properly after m is created
	var overviewView tea.Model

	m := &Model{
		data:         NewDataClient(db),
		overviewView: overviewView,
		ccusageDB: ccusageDB,
		config:    config,
		theme:     theme,
		styles:    styles,
		// help field removed - using new modal-based help system
		showHelp:      false,
		windowFocuser: windowFocuser,
		claudePath:    claudePath,
		mdRenderer:    mdRenderer,
		tabs:           tabs,
		listTabs:       listTabs,
		usageViewport:  usageViewport,
		claudeViewport: claudeViewport,
		claudeFilesViewport: claudeFilesViewport,
		keybindingsViewport: keybindingsViewport,
		projectViewport: projectViewport,
		currentView:    ViewList,
		showAll:        config.ShowAllAgents,
		agents:         []Agent{},
		ccusageSyncing: false,
		lastClickRow:   -1, // Initialize to -1 so first click doesn't trigger double-click
		usageSvc:       usageSvc,
		usageDirty:     true, // Start dirty to force initial build
		claudeFiles:    []ClaudeFile{},
		claudeShowingContent: false,
		modalManager:         modalManager,
		noteModal:            components.NewNoteModal(80, 24), // Initialize with default size
		taskAnnotationModal:  components.NewTaskAnnotationModal(80, 24),
		keybindResolver:      keybindResolver,
		keybindingsError:     keybindingsError,
		overviewStyles:       NewOverviewStyles(styles),
		overviewDirty:        true, // Start dirty to force initial render
	}

	// Initialize view renderers map
	m.renderers = map[ViewType]ViewRenderer{
		ViewList:          (*Model).renderOverviewView, // New modular overview view
		ViewDetail:        (*Model).renderDetail,
		ViewProjectDetail: (*Model).renderProjectDetail,
	}

	// Detect project information at startup
	m.projectInfo = DetectProject()

	// Load project data if detected
	if m.projectInfo != nil {
		// Load markdown files and CLAUDE.md
		m.projectMDFiles = LoadProjectMarkdownFiles(m.projectInfo.GitRootPath)
		m.claudeMDContent = LoadClaudeMDFile(m.projectInfo.GitRootPath)
	}

	// Load keybindings YAML content
	if keybindingsContent, err := keybindings.LoadKeybindingsYAML(); err == nil {
		m.keybindingsYAML = keybindingsContent
		m.keybindingsViewport.SetContent(keybindingsContent)
	}

	// Initialize overview view (new modular architecture)
	m.overviewView = m.createOverviewView()

	// Initial load will happen in Init()
	return m, nil
}

// createOverviewView creates the overview view with injected dependencies
func (m *Model) createOverviewView() tea.Model {
	// Create data client adapter
	dataClient := overview.NewDataClientAdapter(m.data.Agents)

	// Create overview view with viewport provider
	return overview.New(dataClient, &m.styles, m)
}

// GetUsageViewport implements shared.ViewportProvider
func (m *Model) GetUsageViewport() shared.ViewportAccess {
	return &m.usageViewport
}

// GetClaudeViewport implements shared.ViewportProvider
func (m *Model) GetClaudeViewport() shared.ViewportAccess {
	return &m.claudeViewport
}

// GetKeybindingsViewport implements shared.ViewportProvider
func (m *Model) GetKeybindingsViewport() shared.ViewportAccess {
	return &m.keybindingsViewport
}

// GetProjectViewport implements shared.ViewportProvider
func (m *Model) GetProjectViewport() shared.ViewportAccess {
	return &m.projectViewport
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
	// Initialize overview view (will load its own agents)
	overviewCmd := m.overviewView.Init()

	return tea.Batch(
		overviewCmd,                   // Initialize overview view
		loadAgentsCmd(m.data.Agents, m.showAll),  // Load initial agents (for old code, can be removed later)
		m.loadClaudeFilesCmd(),        // Load Claude config files
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

// ccusageDataLoadedMsg is sent when CCUsage data has finished loading
type ccusageDataLoadedMsg struct {
	err error
}

// RefreshUsageMsg triggers a usage tab content refresh
type RefreshUsageMsg struct{}

// loadAgents loads agents from database
func (m *Model) loadAgents() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the data store to load agents
	agents, err := m.data.Agents.LoadAgents(ctx, m.showAll)
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

	// Load notes for current agent's session
	ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	notes, err := m.data.Notes.LoadByAgent(ctx3, domain.AgentID(agent.ID), agent.SessionID)
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

	// Set initial lastFetchedAt to most recent log timestamp
	if len(m.logs) > 0 {
		m.lastFetchedAt = m.logs[len(m.logs)-1].Timestamp
	}

	// Reset scroll position if needed
	if len(m.logs) > 0 && m.logsIndex >= len(m.logs) {
		m.logsIndex = len(m.logs) - 1
	}

	return nil
}

// loadLogsIncremental appends new logs since last fetch (for auto-refresh)
func (m *Model) loadLogsIncremental() error {
	if m.selectedAgent == nil || m.selectedAgent.SessionID == "" {
		return nil
	}

	// Use the last fetched timestamp for incremental pull
	newLogs, err := m.data.DB.GetLogsAfter(m.selectedAgent.SessionID, m.lastFetchedAt)
	if err != nil {
		return err
	}

	if len(newLogs) == 0 {
		return nil
	}

	// Append new logs
	for _, dbLog := range newLogs {
		m.logs = append(m.logs, Log{
			ID:        domain.LogID(dbLog.ID),
			SessionID: dbLog.SessionID,
			Type:      dbLog.Type,
			Message:   dbLog.Message,
			Timestamp: dbLog.Timestamp,
		})
	}

	// Update lastFetchedAt to most recent log
	m.lastFetchedAt = newLogs[len(newLogs)-1].Timestamp

	return nil
}

// loadSessionContexts loads session contexts for the selected agent
func (m *Model) loadSessionContexts() error {
	if m.selectedAgent == nil {
		m.sessionContexts = []domain.SessionContext{}
		return nil
	}

	// Load session contexts from database
	dbContexts, err := m.data.DB.GetSessionContextsForAgent(string(m.selectedAgent.ID))
	if err != nil {
		return err
	}

	// Convert database models to domain models
	m.sessionContexts = make([]domain.SessionContext, len(dbContexts))
	for i, dbCtx := range dbContexts {
		m.sessionContexts[i] = domain.SessionContext{
			ID:        dbCtx.ID,
			AgentID:   domain.AgentID(dbCtx.AgentID),
			SessionID: dbCtx.SessionID,
			Context:   dbCtx.Context,
			CreatedAt: dbCtx.CreatedAt,
			UpdatedAt: dbCtx.UpdatedAt,
		}
	}

	return nil
}

// Helper functions for pointer conversions
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func float32PtrToFloat32(f *float32) float32 {
	if f == nil {
		return 0.0
	}
	return *f
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

	start := time.Now()
	startTime := start.Format("15:04:05.000")
	log.Printf("[PERF][%s] loadTasks() sync starting for agent %s", startTime, m.selectedAgent.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := m.data.Tasks.LoadByAgent(ctx, domain.AgentID(m.selectedAgent.ID), 100, 0)
	if err != nil {
		return err
	}

	// Sort tasks by priority, status, and creation date (same logic as renderTaskList)
	sort.Slice(tasks, func(i, j int) bool {
		// Archived tasks to bottom
		if tasks[i].Archived != tasks[j].Archived {
			return !tasks[i].Archived
		}
		// Higher priority first
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		// Earlier state first (todo < in_progress < done)
		if tasks[i].StateID != tasks[j].StateID {
			return tasks[i].StateID < tasks[j].StateID
		}
		// Newer first
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	elapsed := time.Since(start)
	endTime := time.Now().Format("15:04:05.000")
	log.Printf("[PERF][%s] loadTasks() sync complete in %v", endTime, elapsed)
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
		return TasksLoadedMsg{Tasks: m.tasks}
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
	case 2: // Tasks tab
		return m.loadTasks()
	case 3: // Logs tab
		return m.loadLogs()
	case 6: // Session Context tab
		return m.loadSessionContexts()
	default:
		// Other tabs (Overview, Commits, Notes) already loaded by loadAgentDetails
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
