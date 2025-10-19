package app

import (
	"database/sql"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// ViewMode represents the current view state
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewLogs
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
	ID               string
	Status           string
	Source           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	GitWorktreePath  string
	FeatureDesc      string
	CurrentTask      string
	LastActivityAt   time.Time
	WindowID         string
	AgentDescription string
	ProjectName      string
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
		// Don't fail startup, just log warning
		claudePath = ""
	}

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
		       last_activity_at, window_id, description, project_name
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
		var gitPath, featureDesc, currentTask, windowID, desc, projectName sql.NullString
		var lastActivity sql.NullTime

		err := rows.Scan(
			&a.ID, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt,
			&gitPath, &featureDesc, &currentTask, &lastActivity,
			&windowID, &desc, &projectName,
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
		if desc.Valid {
			a.AgentDescription = desc.String
		}
		if projectName.Valid {
			a.ProjectName = projectName.String
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

	return commitRows.Err()
}

// SelectedAgent returns the currently selected agent
func (m *Model) SelectedAgent() *Agent {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.agents) {
		return nil
	}
	return &m.agents[m.selectedIndex]
}
