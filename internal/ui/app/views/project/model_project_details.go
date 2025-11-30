package project

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/project/tabs"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
)

// Tab index constants - keep explicit and frozen
const (
	tabBack     = 0 // ← back
	tabOverview = 1 // Overview
	tabAgents   = 2 // Agents (active only)
	tabNotes    = 3 // Notes
	tabTasks    = 4 // Tasks (Eye in the Sky i-todo)
	tabFiles    = 5 // Files (ranger/lf browser)
)

type Styles struct {
	OverviewStyles tabs.OverviewStyles
}

// Model is the project details view model
type Model struct {
	// Data context (pointer to avoid accidental copies)
	ctx *DataContext

	// View cache for the overview sub-tab
	overviewCache string
	overviewDirty bool

	// Tabs
	tabs components.NavBar

	// Table models for split-pane tabs
	agentsTable table.Model
	notesTable  table.Model
	tasksTable  table.Model

	// Dimensions
	width  int
	height int

	// Styles
	styles Styles
}

// New creates a new project details model
func New(styles Styles, tabs components.NavBar) Model {
	return Model{
		styles:        styles,
		tabs:          tabs, // Pass prebuilt tab bar to keep consistency with overview
		overviewDirty: true,
	}
}

// Init initializes the project details model
func (m *Model) Init() tea.Cmd {
	return nil
}

// SetContext updates the data context and marks overview as dirty
func (m *Model) SetContext(ctx *DataContext) {
	m.ctx = ctx
	m.overviewDirty = true

	// Initialize tables when context is set
	m.initializeTables()
}

// SetTabs updates the tabs navbar
func (m *Model) SetTabs(tabs components.NavBar) {
	m.tabs = tabs
}

// SetSize updates the view dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Update context width for responsive rendering
	if m.ctx != nil {
		m.ctx.Width = width
	}

	// Reinitialize tables with new width
	m.initializeTables()
}

// initializeTables creates table models from context data
func (m *Model) initializeTables() {
	if m.ctx == nil {
		return
	}

	// Initialize agents table
	m.agentsTable = m.createAgentsTable()

	// Initialize notes table (will implement later)
	m.notesTable = m.createNotesTable()

	// Initialize tasks table (will implement later)
	m.tasksTable = m.createTasksTable()
}

// createAgentsTable builds a table.Model for the agents list
func (m *Model) createAgentsTable() table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Status", Width: 10},
		{Title: "Description", Width: 30},
		{Title: "Session", Width: 12},
	}

	rows := []table.Row{}
	for _, agent := range m.ctx.Agents {
		rows = append(rows, table.Row{
			string(agent.ID),
			agent.Status,
			agent.AgentDescription,
			agent.SessionID,
		})
	}

	// Calculate table width (left pane is half the width)
	tableWidth := m.width / 2
	if tableWidth < 60 {
		tableWidth = 60
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height - 5), // Leave room for header/footer
	)

	// Set table styles using our theme
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	s.Selected = s.Selected.Bold(true).Foreground(m.styles.OverviewStyles.Primary.GetForeground())
	t.SetStyles(s)

	return t
}

// createNotesTable builds a table.Model for the notes list
func (m *Model) createNotesTable() table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Type", Width: 12},
		{Title: "Preview", Width: 40},
		{Title: "Created", Width: 12},
	}

	rows := []table.Row{}
	for _, note := range m.ctx.Notes {
		// Create preview from body (first 40 chars)
		preview := note.Body
		if len(preview) > 40 {
			preview = preview[:37] + "..."
		}

		// Format timestamp
		created := ""
		if !note.CreatedAt.IsZero() {
			created = note.CreatedAt.Format("2006-01-02")
		}

		rows = append(rows, table.Row{
			string(note.ID),
			note.ParentType,
			preview,
			created,
		})
	}

	// Calculate table width (left pane is half the width)
	tableWidth := m.width / 2
	if tableWidth < 60 {
		tableWidth = 60
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height - 5), // Leave room for header/footer
	)

	// Set table styles using our theme
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	s.Selected = s.Selected.Bold(true).Foreground(m.styles.OverviewStyles.Primary.GetForeground())
	t.SetStyles(s)

	return t
}

// createTasksTable builds a table.Model for the tasks list
func (m *Model) createTasksTable() table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Title", Width: 30},
		{Title: "State", Width: 12},
		{Title: "Priority", Width: 8},
	}

	rows := []table.Row{}
	for _, task := range m.ctx.Tasks {
		// Format state
		stateStr := formatTaskStateShort(task.StateID)

		// Format priority
		priorityStr := formatPriorityShort(task.Priority)

		rows = append(rows, table.Row{
			string(task.ID),
			task.Title,
			stateStr,
			priorityStr,
		})
	}

	// Calculate table width (left pane is half the width)
	tableWidth := m.width / 2
	if tableWidth < 60 {
		tableWidth = 60
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height - 5), // Leave room for header/footer
	)

	// Set table styles using our theme
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	s.Selected = s.Selected.Bold(true).Foreground(m.styles.OverviewStyles.Primary.GetForeground())
	t.SetStyles(s)

	return t
}

// formatTaskStateShort converts state ID to short string
func formatTaskStateShort(stateID int) string {
	switch stateID {
	case 1:
		return "Todo"
	case 2:
		return "In Progress"
	case 3:
		return "Done"
	default:
		return "Unknown"
	}
}

// formatPriorityShort converts priority to short string
func formatPriorityShort(priority int) string {
	switch {
	case priority >= 4:
		return "High"
	case priority >= 2:
		return "Medium"
	default:
		return "Low"
	}
}
