package agent_details

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/agent_details/tabs"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
)

// Tab index constants - keep explicit and frozen
const (
	tabBack           = 0 // iota // 0 ← back
	tabOverview       = 1 // 1
	tabTasks          = 2 // 2
	tabActions        = 3 // 3 (removed from UI but kept for compatibility)
	tabLogs           = 3 // 3 (was 4)
	tabCommits        = 4 // 4 (was 5)
	tabNotes          = 5 // 5 (was 6)
	tabSessionContext = 6 // 6 (was 7)
)

type Styles struct {
	OverviewStyles tabs.OverviewStyles
}

// Model is the agent details view model
type Model struct {
	// Data context (pointer to avoid accidental copies)
	ctx *DataContext

	// View cache for the overview sub-tab
	overviewCache string
	overviewDirty bool

	// Tabs
	tabs components.NavBar

	// Dimensions
	width  int
	height int

	// Styles
	styles Styles
}

// New creates a new agent details model
func New(styles Styles, tabs components.NavBar) Model {
	return Model{
		styles:        styles,
		tabs:          tabs, // Pass prebuilt tab bar to keep consistency with overview
		overviewDirty: true,
	}
}

// Init initializes the agent details model
func (m Model) Init() tea.Cmd {
	return nil
}

// SetContext updates the data context and marks overview as dirty
func (m *Model) SetContext(ctx *DataContext) {
	m.ctx = ctx
	m.overviewDirty = true
}

// SetSize updates the view dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Update context width for responsive rendering
	if m.ctx != nil {
		m.ctx.Width = width
	}
}
