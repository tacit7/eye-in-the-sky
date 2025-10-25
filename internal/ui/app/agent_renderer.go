package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// ColumnConfig controls which columns are visible in the agent table
type ColumnConfig struct {
	Icon    bool
	Status  bool
	ID      bool
	Task    bool
	Source  bool
	Session bool
}

// DefaultColumnConfig returns the default column configuration
func DefaultColumnConfig() ColumnConfig {
	return ColumnConfig{
		Icon:    true,
		Status:  true,
		ID:      true,
		Task:    true,
		Source:  true,
		Session: true,
	}
}

// AgentLineRenderer renders agent data into table rows
type AgentLineRenderer struct {
	Columns     ColumnConfig
	StatusStyle map[string]lipgloss.Style
	Styles      *Styles
}

// NewAgentLineRenderer creates a new agent line renderer
func NewAgentLineRenderer(styles *Styles) *AgentLineRenderer {
	return &AgentLineRenderer{
		Columns: DefaultColumnConfig(),
		StatusStyle: map[string]lipgloss.Style{
			"active":    styles.Success,
			"working":   styles.Warning,
			"idle":      styles.Subtle,
			"failed":    styles.Error,
			"completed": styles.Primary,
		},
		Styles: styles,
	}
}

// RenderHeaders returns the column headers for the table
func (r *AgentLineRenderer) RenderHeaders() []string {
	var headers []string

	if r.Columns.Icon {
		headers = append(headers, "")
	}
	if r.Columns.Status {
		headers = append(headers, "Status")
	}
	if r.Columns.ID {
		headers = append(headers, "Agent ID")
	}
	if r.Columns.Task {
		headers = append(headers, "Task")
	}
	if r.Columns.Source {
		headers = append(headers, "Source")
	}
	if r.Columns.Session {
		headers = append(headers, "Session")
	}

	return headers
}

// RenderAgent renders an agent as a table row
func (r *AgentLineRenderer) RenderAgent(agent domain.Agent, selected bool) []string {
	var cells []string

	// Icon column
	if r.Columns.Icon {
		icon := r.getAgentIcon(agent)
		cells = append(cells, icon)
	}

	// Status column
	if r.Columns.Status {
		status := r.formatStatus(agent.Status)
		cells = append(cells, status)
	}

	// ID column
	if r.Columns.ID {
		id := truncateID(string(agent.ID), 8)
		cells = append(cells, id)
	}

	// Task column
	if r.Columns.Task {
		task := r.formatTask(agent)
		cells = append(cells, task)
	}

	// Source column
	if r.Columns.Source {
		source := string(agent.Source)
		cells = append(cells, source)
	}

	// Session column
	if r.Columns.Session {
		session := truncateID(agent.SessionID, 8)
		cells = append(cells, session)
	}

	return cells
}

// getAgentIcon returns the appropriate icon for an agent
func (r *AgentLineRenderer) getAgentIcon(agent domain.Agent) string {
	if agent.ParentSessionID != "" {
		// This is a subagent - use green pipe
		return r.Styles.Success.Render("│")
	}
	return " "
}

// formatStatus formats and styles the status
func (r *AgentLineRenderer) formatStatus(status string) string {
	style, exists := r.StatusStyle[status]
	if !exists {
		style = r.Styles.Subtle
	}
	return style.Render(status)
}

// formatTask formats the task/feature description
func (r *AgentLineRenderer) formatTask(agent domain.Agent) string {
	task := agent.CurrentTask
	if task == "" {
		task = agent.FeatureDesc
	}
	if task == "" {
		return r.Styles.Subtle.Render("(no task)")
	}
	return task
}

// RenderAgentRow renders a complete styled row for an agent
func (r *AgentLineRenderer) RenderAgentRow(agent domain.Agent, selected bool) string {
	cells := r.RenderAgent(agent, selected)
	row := r.formatRow(cells)

	if selected {
		return r.Styles.Selected.Render(row)
	}
	return row
}

// formatRow formats cells into a properly spaced row
func (r *AgentLineRenderer) formatRow(cells []string) string {
	// Define column widths
	widths := r.getColumnWidths()

	var formatted []string
	for i, cell := range cells {
		if i < len(widths) {
			formatted = append(formatted, truncateField(cell, widths[i]))
		}
	}

	return strings.Join(formatted, "  ")
}

// getColumnWidths returns the configured column widths
func (r *AgentLineRenderer) getColumnWidths() []int {
	var widths []int

	if r.Columns.Icon {
		widths = append(widths, 2)
	}
	if r.Columns.Status {
		widths = append(widths, 12)
	}
	if r.Columns.ID {
		widths = append(widths, 10)
	}
	if r.Columns.Task {
		widths = append(widths, 40)
	}
	if r.Columns.Source {
		widths = append(widths, 30)
	}
	if r.Columns.Session {
		widths = append(widths, 10)
	}

	return widths
}

// truncateField truncates a field to the specified width
func truncateField(value string, maxWidth int) string {
	if lipgloss.Width(value) <= maxWidth {
		return padRight(value, maxWidth)
	}
	if maxWidth > 3 {
		return value[:maxWidth-3] + "..."
	}
	return "..."
}

// CreateTableBuilder creates a TableBuilder configured for agent display
func (r *AgentLineRenderer) CreateTableBuilder(styles *Styles) *TableBuilder {
	tb := NewTableBuilder()
	tb.SetHeaderStyle(styles.Primary)
	tb.SetBorderStyle(BorderSimple)

	headers := r.RenderHeaders()
	widths := r.getColumnWidths()

	for i, header := range headers {
		width := 20 // default
		if i < len(widths) {
			width = widths[i]
		}
		tb.AddColumn(header, width, AlignLeft, true)
	}

	return tb
}

// RenderAgentTable renders a complete table of agents
func (r *AgentLineRenderer) RenderAgentTable(agents []domain.Agent, selectedIndex int) string {
	tb := r.CreateTableBuilder(r.Styles)

	// Set row styling function for selection
	tb.SetRowStyleFunc(func(rowIndex int) lipgloss.Style {
		if rowIndex == selectedIndex {
			return r.Styles.Selected
		}
		return lipgloss.NewStyle()
	})

	// Convert agents to rows
	var rows [][]string
	for _, agent := range agents {
		row := r.RenderAgent(agent, false) // selection handled by style func
		rows = append(rows, row)
	}

	result, _ := tb.RenderTable(rows)
	return result
}

// ConfigureColumns sets which columns are visible
func (r *AgentLineRenderer) ConfigureColumns(config ColumnConfig) {
	r.Columns = config
}

// StatusIcon returns an icon for the given status
func StatusIcon(status string) string {
	icons := map[string]string{
		"active":    "●",
		"working":   "◉",
		"idle":      "○",
		"completed": "✓",
		"failed":    "✗",
	}
	if icon, exists := icons[status]; exists {
		return icon
	}
	return "·"
}

// StatusColor returns a color name for the given status
func StatusColor(status string) string {
	colors := map[string]string{
		"active":    "green",
		"working":   "yellow",
		"idle":      "gray",
		"completed": "blue",
		"failed":    "red",
	}
	if color, exists := colors[status]; exists {
		return color
	}
	return "white"
}