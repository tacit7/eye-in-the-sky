package app

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// ViewMode represents the current view state
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewLogs
	ViewTasks
	ViewCommits
	ViewNotes
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

	// Tabs for detail view
	tabs components.TabsModel

	// View state
	currentView ViewMode
	showAll     bool // Show all agents or only active ones

	// Agent list state
	agents        []Agent
	selectedIndex int
	listOffset    int

	// Agent detail state
	selectedAgent *Agent
	detailOffset  int
	actions       []Action
	commits       []Commit
	notes         []Note
	tasks         []Task
	logs          []Log

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
func NewModel(db *sql.DB) (*Model, error) {
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

	// Create tabs for detail view
	tabs := components.NewTabsModel(
		[]string{"[O]verview", "[C]ommits", "[L]ogs", "[N]otes", "[A]ctions"},
		theme.Colors.Active,
		theme.Colors.Text,
	)

	m := &Model{
		db:            db,
		config:        config,
		keys:          keys,
		theme:         theme,
		styles:        styles,
		help:          helpModel,
		showHelp:      false,
		windowFocuser: windowFocuser,
		claudePath:    claudePath,
		tabs:          tabs,
		currentView:   ViewList,
		showAll:       config.ShowAllAgents,
		agents:        []Agent{},
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

// loadAgents loads agents from database
func (m *Model) loadAgents() error {
	query := `
		SELECT id, status, source, created_at, updated_at,
		       git_worktree_path, feature_description, current_task,
		       last_activity_at, window_id, terminal_application, description, project_name, current_session_id
		FROM agents
	`

	if !m.showAll {
		query += ` WHERE status IN ('active', 'working', 'idle', 'stale', 'unknown')`
	}

	query += ` ORDER BY last_activity_at DESC`

	rows, err := m.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	agents := []Agent{}
	for rows.Next() {
		var a Agent
		var gitPath, featureDesc, currentTask, windowID, terminalApp, desc, projectName, sessionID sql.NullString
		var lastActivity sql.NullTime

		err := rows.Scan(
			&a.ID, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt,
			&gitPath, &featureDesc, &currentTask, &lastActivity,
			&windowID, &terminalApp, &desc, &projectName, &sessionID,
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

		agents = append(agents, a)
	}

	m.agents = agents
	m.lastRefresh = time.Now()

	// Adjust selected index if needed
	if m.selectedIndex >= len(m.agents) && len(m.agents) > 0 {
		m.selectedIndex = len(m.agents) - 1
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
