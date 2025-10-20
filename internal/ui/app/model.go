package app

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// ViewMode represents the current view state
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
)

// Model represents the application state
type Model struct {
	// Database connection
	db *sql.DB

	// Configuration
	config      Config
	keys        KeyBindings
	theme       Theme
	styles      Styles

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
	currentView ViewMode
	showAll     bool // Show all agents or only active ones

	// Agent list state
	agents        []Agent
	selectedIndex int
	listOffset    int

	// Agent detail state
	selectedAgent  *Agent
	detailOffset   int
	actions        []Action
	commits        []Commit
	notes          []Note
	tasks          []Task
	logs           []Log
	sessionMetrics []SessionMetric

	// Overview state
	allSessionMetrics  []SessionMetric // All metrics from all agents
	monthlyCosts       []*SessionMetric // Monthly costs with timestamps

	// Per-view scroll state
	tasksIndex    int
	tasksOffset   int
	commitsIndex  int
	commitsOffset int
	notesIndex    int
	notesOffset   int
	logsIndex     int
	logsOffset    int
	actionsIndex  int
	actionsOffset int

	// Right pane scroll state
	rightPaneOffset int

	// UI dimensions
	width  int
	height int

	// Polling
	lastRefresh time.Time
	err         error

	// Status message
	statusMsg string

	// CCUsage database connection
	ccusageDB *db.CCUsageDB

	// Cached ccusage data
	ccusageDaily    []api.DailyReport
	ccusageSessions []api.SessionReport
	ccusageMonthly  *api.MonthlyReport
	ccusageBlock    *api.ActiveBlockReport
	ccusageCosts    *api.CostSummary

	// CCUsage sync state
	ccusageSyncing    bool
	ccusageSyncStatus string
	ccusageEntryCount int
	lastCCUsageSync   time.Time
}

// Agent represents an agent from the database
type Agent struct {
	ID                  string
	Status              string
	Source              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	GitWorktreePath     string
	FeatureDesc         string
	CurrentTask         string
	LastActivityAt      time.Time
	WindowID            string
	TerminalApplication string
	AgentDescription    string
	ProjectName         string
	CurrentSessionID    string
	ParentAgentID       string
}

// Action represents an agent action from the database
type Action struct {
	ID          int
	AgentID     string
	ActionType  string
	Description string
	Details     string
	Timestamp   time.Time
}

// Commit represents a git commit from the database
type Commit struct {
	ID            int
	AgentID       string
	CommitHash    string
	CommitMessage string
	Timestamp     time.Time
}

// Note represents a session note from the database
type Note struct {
	ID        int
	SessionID string
	Content   string
	Timestamp time.Time
}

// Task represents a Taskwarrior task
type Task struct {
	UUID        string
	Description string
	Status      string
	Priority    string
	Project     string
	Tags        []string
	Due         time.Time
	Entry       time.Time
	Annotations []TaskAnnotation
}

// TaskAnnotation represents a task annotation with timestamp
type TaskAnnotation struct {
	Entry       time.Time
	Description string
}

// Log represents a session log entry
type Log struct {
	ID        int
	SessionID string
	Type      string
	Message   string
	Timestamp time.Time
}

// SessionMetric represents token usage and cost tracking
type SessionMetric struct {
	ID               int
	AgentID          string
	SessionID        string
	TokensUsed       int
	TokensBudget     int
	TokensRemaining  int
	InputTokens      int
	OutputTokens     int
	EstimatedCostUSD float64
	ModelName        string
	Timestamp        time.Time
	CreatedAt        time.Time
	Notes            string
}

// Styles holds all lipgloss styles
type Styles struct {
	Active    lipgloss.Style
	Working   lipgloss.Style
	Idle      lipgloss.Style
	Stale     lipgloss.Style
	Unknown   lipgloss.Style
	Completed lipgloss.Style
	Failed    lipgloss.Style
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Border    lipgloss.Style
	Title     lipgloss.Style
	Text      lipgloss.Style
	Subtle    lipgloss.Style
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
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		claudePath = ""
	}

	// Create tabs for agent detail view (← is back arrow, unicode 8678)
	tabs := components.NewTabsModel(
		[]string{"← Back", "[A]gent View", "[C]ommits", "[L]ogs", "[N]otes", "[A]ctions"},
		theme.Colors.Active,
		theme.Colors.Text,
	)

	// Create tabs for overview (agent list)
	listTabs := components.NewTabsModel(
		[]string{"[O]verview", "[P]roject", "[C]laude", "[U]sage"},
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
		fmt.Fprintf(os.Stderr, "Warning: Failed to create markdown renderer: %v\n", err)
		mdRenderer = nil
	}

	m := &Model{
		db:            db,
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
		currentView:   ViewList,
		showAll:       config.ShowAllAgents,
		agents:        []Agent{},
		ccusageSyncing: false,
	}

	// Load initial agent list
	if err := m.loadAgents(); err != nil {
		m.err = err
	}

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
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
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

// loadAgents loads agents from database
func (m *Model) loadAgents() error {
	query := `
		SELECT id, status, source, created_at, updated_at,
		       git_worktree_path, feature_description, current_task,
		       last_activity_at, window_id, terminal_application, description, project_name, current_session_id, parent_agent_id
		FROM agents
	`

	if !m.showAll {
		query += ` WHERE status IN ('active', 'working', 'idle', 'stale', 'unknown')`
	}

	query += ` ORDER BY
		CASE WHEN parent_agent_id IS NULL THEN id ELSE parent_agent_id END,
		CASE WHEN parent_agent_id IS NULL THEN 0 ELSE 1 END,
		last_activity_at DESC`

	rows, err := m.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	agents := []Agent{}
	for rows.Next() {
		var a Agent
		var gitPath, featureDesc, currentTask, windowID, terminalApp, desc, projectName, sessionID, parentAgentID sql.NullString
		var lastActivity sql.NullTime

		err := rows.Scan(
			&a.ID, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt,
			&gitPath, &featureDesc, &currentTask, &lastActivity,
			&windowID, &terminalApp, &desc, &projectName, &sessionID, &parentAgentID,
		)
		if err != nil {
			return err
		}

		if gitPath.Valid {
			a.GitWorktreePath = gitPath.String
		}
		if featureDesc.Valid {
			a.FeatureDesc = featureDesc.String
		}
		if currentTask.Valid {
			a.CurrentTask = currentTask.String
		}
		if lastActivity.Valid {
			a.LastActivityAt = lastActivity.Time
		}
		if windowID.Valid {
			a.WindowID = windowID.String
		}
		if terminalApp.Valid {
			a.TerminalApplication = terminalApp.String
		}
		if desc.Valid {
			a.AgentDescription = desc.String
		}
		if projectName.Valid {
			a.ProjectName = projectName.String
		}
		if sessionID.Valid {
			a.CurrentSessionID = sessionID.String
		}
		if parentAgentID.Valid {
			a.ParentAgentID = parentAgentID.String
		}

		agents = append(agents, a)
	}

	m.agents = agents
	m.lastRefresh = time.Now()

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

	return rows.Err()
}

// loadAgentDetails loads details for the currently selected agent
func (m *Model) loadAgentDetails() error {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.agents) {
		return nil
	}

	agent := m.agents[m.selectedIndex]
	m.selectedAgent = &agent

	// Load actions
	actionsQuery := `
		SELECT id, agent_id, action_type, description, details, timestamp
		FROM actions
		WHERE agent_id = ?
		ORDER BY timestamp DESC
		LIMIT 50
	`

	rows, err := m.db.Query(actionsQuery, agent.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	actions := []Action{}
	for rows.Next() {
		var a Action
		var details sql.NullString

		err := rows.Scan(&a.ID, &a.AgentID, &a.ActionType, &a.Description, &details, &a.Timestamp)
		if err != nil {
			return err
		}

		if details.Valid {
			a.Details = details.String
		}

		actions = append(actions, a)
	}
	m.actions = actions

	// Load commits
	commitsQuery := `
		SELECT id, agent_id, commit_hash, commit_message, timestamp
		FROM commits
		WHERE agent_id = ?
		ORDER BY timestamp DESC
		LIMIT 20
	`

	commitRows, err := m.db.Query(commitsQuery, agent.ID)
	if err != nil {
		return err
	}
	defer commitRows.Close()

	commits := []Commit{}
	for commitRows.Next() {
		var c Commit
		var message sql.NullString

		err := commitRows.Scan(&c.ID, &c.AgentID, &c.CommitHash, &message, &c.Timestamp)
		if err != nil {
			return err
		}

		if message.Valid {
			c.CommitMessage = message.String
		}

		commits = append(commits, c)
	}
	m.commits = commits

	// Load notes for current session
	notes := []Note{}
	if agent.CurrentSessionID != "" {
		notesQuery := `
			SELECT id, session_id, content, created_at
			FROM notes
			WHERE session_id = ?
			ORDER BY created_at DESC
			LIMIT 20
		`

		noteRows, err := m.db.Query(notesQuery, agent.CurrentSessionID)
		if err != nil {
			return err
		}
		defer noteRows.Close()

		for noteRows.Next() {
			var n Note

			err := noteRows.Scan(&n.ID, &n.SessionID, &n.Content, &n.Timestamp)
			if err != nil {
				return err
			}

			notes = append(notes, n)
		}

		if err := noteRows.Err(); err != nil {
			return err
		}
	}
	m.notes = notes

	// Load session metrics
	if err := m.loadSessionMetrics(); err != nil {
		// Don't fail if metrics loading fails, just log it
		m.err = err
	}

	return commitRows.Err()
}

// loadLogs loads logs for the current agent session
func (m *Model) loadLogs() error {
	if m.selectedAgent == nil || m.selectedAgent.CurrentSessionID == "" {
		m.logs = []Log{}
		return nil
	}

	logsQuery := `
		SELECT id, session_id, type, message, created_at
		FROM logs
		WHERE session_id = ?
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := m.db.Query(logsQuery, m.selectedAgent.CurrentSessionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	logs := []Log{}
	for rows.Next() {
		var log Log
		err := rows.Scan(&log.ID, &log.SessionID, &log.Type, &log.Message, &log.Timestamp)
		if err != nil {
			return err
		}
		logs = append(logs, log)
	}

	m.logs = logs

	// Reset scroll position if needed
	if len(m.logs) > 0 && m.logsIndex >= len(m.logs) {
		m.logsIndex = len(m.logs) - 1
	}

	return rows.Err()
}

// loadMonthlyCosts loads all session metrics from the current month with timestamps
func (m *Model) loadMonthlyCosts() error {
	query := `
		SELECT id, agent_id, session_id, tokens_used, tokens_budget, tokens_remaining,
		       input_tokens, output_tokens, estimated_cost_usd, model_name, timestamp, created_at, notes
		FROM session_metrics
		WHERE strftime('%Y-%m', timestamp) = strftime('%Y-%m', 'now')
		ORDER BY timestamp DESC
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	metrics := []*SessionMetric{}
	for rows.Next() {
		var metric SessionMetric
		var sessionID, modelName, notes sql.NullString
		var inputTokens, outputTokens sql.NullInt64
		var estimatedCost sql.NullFloat64

		err := rows.Scan(
			&metric.ID,
			&metric.AgentID,
			&sessionID,
			&metric.TokensUsed,
			&metric.TokensBudget,
			&metric.TokensRemaining,
			&inputTokens,
			&outputTokens,
			&estimatedCost,
			&modelName,
			&metric.Timestamp,
			&metric.CreatedAt,
			&notes,
		)
		if err != nil {
			return err
		}

		if sessionID.Valid {
			metric.SessionID = sessionID.String
		}
		if inputTokens.Valid {
			metric.InputTokens = int(inputTokens.Int64)
		}
		if outputTokens.Valid {
			metric.OutputTokens = int(outputTokens.Int64)
		}
		if estimatedCost.Valid {
			metric.EstimatedCostUSD = estimatedCost.Float64
		}
		if modelName.Valid {
			metric.ModelName = modelName.String
		}
		if notes.Valid {
			metric.Notes = notes.String
		}

		metrics = append(metrics, &metric)
	}

	m.monthlyCosts = metrics
	return rows.Err()
}

// loadAllSessionMetrics loads the latest session metrics for all agents
func (m *Model) loadAllSessionMetrics() error {
	query := `
		SELECT DISTINCT ON (agent_id) id, agent_id, session_id, tokens_used, tokens_budget, tokens_remaining,
		       input_tokens, output_tokens, estimated_cost_usd, model_name, timestamp, created_at, notes
		FROM session_metrics
		ORDER BY agent_id, timestamp DESC
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	metrics := []SessionMetric{}
	for rows.Next() {
		var metric SessionMetric
		var sessionID, modelName, notes sql.NullString
		var inputTokens, outputTokens sql.NullInt64
		var estimatedCost sql.NullFloat64

		err := rows.Scan(
			&metric.ID,
			&metric.AgentID,
			&sessionID,
			&metric.TokensUsed,
			&metric.TokensBudget,
			&metric.TokensRemaining,
			&inputTokens,
			&outputTokens,
			&estimatedCost,
			&modelName,
			&metric.Timestamp,
			&metric.CreatedAt,
			&notes,
		)
		if err != nil {
			return err
		}

		if sessionID.Valid {
			metric.SessionID = sessionID.String
		}
		if inputTokens.Valid {
			metric.InputTokens = int(inputTokens.Int64)
		}
		if outputTokens.Valid {
			metric.OutputTokens = int(outputTokens.Int64)
		}
		if estimatedCost.Valid {
			metric.EstimatedCostUSD = estimatedCost.Float64
		}
		if modelName.Valid {
			metric.ModelName = modelName.String
		}
		if notes.Valid {
			metric.Notes = notes.String
		}

		metrics = append(metrics, metric)
	}

	m.allSessionMetrics = metrics
	return rows.Err()
}

// loadSessionMetrics loads session metrics for the current agent
func (m *Model) loadSessionMetrics() error {
	if m.selectedAgent == nil {
		m.sessionMetrics = []SessionMetric{}
		return nil
	}

	metricsQuery := `
		SELECT id, agent_id, session_id, tokens_used, tokens_budget, tokens_remaining,
		       input_tokens, output_tokens, estimated_cost_usd, model_name, timestamp, created_at, notes
		FROM session_metrics
		WHERE agent_id = ?
		ORDER BY timestamp DESC
		LIMIT 10
	`

	rows, err := m.db.Query(metricsQuery, m.selectedAgent.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	metrics := []SessionMetric{}
	for rows.Next() {
		var metric SessionMetric
		var sessionID, modelName, notes sql.NullString
		var inputTokens, outputTokens sql.NullInt64
		var estimatedCost sql.NullFloat64

		err := rows.Scan(
			&metric.ID,
			&metric.AgentID,
			&sessionID,
			&metric.TokensUsed,
			&metric.TokensBudget,
			&metric.TokensRemaining,
			&inputTokens,
			&outputTokens,
			&estimatedCost,
			&modelName,
			&metric.Timestamp,
			&metric.CreatedAt,
			&notes,
		)
		if err != nil {
			return err
		}

		if sessionID.Valid {
			metric.SessionID = sessionID.String
		}
		if inputTokens.Valid {
			metric.InputTokens = int(inputTokens.Int64)
		}
		if outputTokens.Valid {
			metric.OutputTokens = int(outputTokens.Int64)
		}
		if estimatedCost.Valid {
			metric.EstimatedCostUSD = estimatedCost.Float64
		}
		if modelName.Valid {
			metric.ModelName = modelName.String
		}
		if notes.Valid {
			metric.Notes = notes.String
		}

		metrics = append(metrics, metric)
	}

	m.sessionMetrics = metrics

	return rows.Err()
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
		m.ccusageMonthly = nil
		m.ccusageBlock = nil
		m.ccusageCosts = nil
		return nil
	}

	// Load daily usage (last 7 days)
	daily, err := api.GetDailyUsageReport(m.ccusageDB, 7)
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

	// Load monthly summary
	now := time.Now()
	monthly, err := api.GetMonthlyUsageReport(m.ccusageDB, now.Year(), int(now.Month()))
	if err != nil {
		return fmt.Errorf("failed to load monthly: %w", err)
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
	return nil
}

// loadTabData loads data for the currently active tab
func (m *Model) loadTabData() error {
	if m.selectedAgent == nil {
		return nil
	}

	switch m.tabs.ActiveIndex {
	case 2: // Logs tab
		return m.loadLogs()
	default:
		// Other tabs (Overview, Commits, Notes, Actions) already loaded by loadAgentDetails
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
