package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/config"
	"github.com/tacit7/eye-in-the-sky/internal/utils"
)

// HeaderState contains header focus and sorting state
type HeaderState struct {
	HeaderFocused  bool
	SelectedColumn int
	SortField      string
	SortAscending  bool
}

// Styles interface defines the styling methods needed by AgentLineRenderer
type Styles interface {
	// Style accessors
	GetSuccess() lipgloss.Style
	GetWarning() lipgloss.Style
	GetSubtle() lipgloss.Style
	GetError() lipgloss.Style
	GetPrimary() lipgloss.Style
	GetSelected() lipgloss.Style
}

// AgentTableColumn defines a column in the agent table with dynamic width support
type AgentTableColumn struct {
	Title  string
	Hidden *bool
	Width  *int
	Grow   *bool
}

// ColumnConfig controls which columns are visible in the agent table
type ColumnConfig struct {
	Icon         bool
	Status       bool
	Task         bool
	Source       bool
	Session      bool
	LastActivity bool
	ProjectName  bool
	LastLog      bool
}

// DefaultColumnConfig returns the default column configuration
func DefaultColumnConfig() ColumnConfig {
	return ColumnConfig{
		Icon:    true,
		Status:  true,
		Task:    true,
		Source:  true,
		Session: true,
	}
}

// AgentLineRenderer renders agent data into table rows
type AgentLineRenderer struct {
	Columns     ColumnConfig
	StatusStyle map[string]lipgloss.Style
	Styles      Styles
}

// NewAgentLineRenderer creates a new agent line renderer
func NewAgentLineRenderer(styles Styles) *AgentLineRenderer {
	return &AgentLineRenderer{
		Columns: DefaultColumnConfig(),
		StatusStyle: map[string]lipgloss.Style{
			"active":    styles.GetSuccess(),
			"working":   styles.GetWarning(),
			"idle":      styles.GetSubtle(),
			"failed":    styles.GetError(),
			"completed": styles.GetPrimary(),
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
		headers = append(headers, " Status  ")
	}
	if r.Columns.LastLog {
		headers = append(headers, "Last Activity ")
	}
	if r.Columns.Session {
		headers = append(headers, "Session  ")
	}
	if r.Columns.Task {
		headers = append(headers, "Description  ")
	}
	if r.Columns.ProjectName {
		headers = append(headers, "Project ")
	}
	if r.Columns.Source {
		headers = append(headers, "Source ")
	}

	return headers
}

// RenderAgent renders an agent as a table row
func (r *AgentLineRenderer) RenderAgent(agent domain.Agent, selected bool) []string {
	var cells []string

	isChild := agent.ParentAgentID != ""

	// Icon column
	if r.Columns.Icon {
		icon := r.getAgentIcon(agent)
		cells = append(cells, icon)
	}

	// Status column with padding (first column gets leading space)
	if r.Columns.Status {
		status := r.formatStatus(agent.Status, isChild)
		cells = append(cells, " "+status+"  ")
	}

	// LastLog column (timestamp) - right after Status
	if r.Columns.LastLog {
		lastLog := agent.LastLog
		if lastLog == "" {
			lastLog = "-"
		}
		cells = append(cells, lastLog+"  ")
	}

	// Session column with padding
	if r.Columns.Session {
		session := utils.TruncateID(agent.SessionID, 8)
		cells = append(cells, session+"  ")
	}

	// Task column with padding
	if r.Columns.Task {
		task := r.formatTask(agent)
		cells = append(cells, task+"  ")
	}

	// ProjectName column (replaces Source)
	if r.Columns.ProjectName {
		projectName := agent.ProjectName
		if projectName == "" {
			projectName = "-"
		}
		cells = append(cells, projectName+"  ")
	}

	// Source column
	if r.Columns.Source {
		source := string(agent.Source)
		cells = append(cells, source+" ")
	}

	// LastActivity column
	if r.Columns.LastActivity {
		lastActivity := r.formatLastActivity(agent.LastActivityAt)
		cells = append(cells, lastActivity)
	}

	return cells
}

// getAgentIcon returns the appropriate icon for an agent
func (r *AgentLineRenderer) getAgentIcon(agent domain.Agent) string {
	if agent.Bookmarked {
		// Bookmarked agents show a checkmark
		icon := getIcon("bookmark")
		return r.Styles.GetSuccess().Render(icon)
	}
	if agent.ParentAgentID != "" {
		// This is a subagent - use vertical line
		icon := getIcon("pipe")
		return r.Styles.GetSuccess().Render(icon)
	}
	return " "
}

// formatStatus formats and styles the status with icon
func (r *AgentLineRenderer) formatStatus(status string, isChild bool) string {
	var icon string
	if isChild {
		// Child agents show cyan vertical line
		pipeIcon := getIcon("pipe")
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Render(pipeIcon)
	} else {
		icon = r.getStatusIcon(status)
	}
	style, exists := r.StatusStyle[status]
	if !exists {
		style = r.Styles.GetSubtle()
	}
	return icon + " " + style.Render(status)
}

// getStatusIcon returns the icon for a given status with orange color
func (r *AgentLineRenderer) getStatusIcon(status string) string {
	icon := getIcon(status)
	// Render icon in orange
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00"))
	return orangeStyle.Render(icon)
}

// formatTask formats the task/feature description with truncation
func (r *AgentLineRenderer) formatTask(agent domain.Agent) string {
	// Prioritize feature description (session description)
	desc := agent.FeatureDesc
	if desc == "" {
		desc = agent.CurrentTask
	}
	if desc == "" {
		return r.Styles.GetSubtle().Render("(no description)")
	}

	// Truncate description to max 50 characters
	maxWidth := 50
	if len(desc) > maxWidth {
		return desc[:maxWidth-3] + "..."
	}
	return desc
}

// RenderAgentRow renders a complete styled row for an agent
func (r *AgentLineRenderer) RenderAgentRow(agent domain.Agent, selected bool) string {
	cells := r.RenderAgent(agent, selected)
	row := r.formatRow(cells)

	if selected {
		return r.Styles.GetSelected().Render(row)
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
	if r.Columns.LastLog {
		widths = append(widths, 15)
	}
	if r.Columns.Session {
		widths = append(widths, 10)
	}
	if r.Columns.Task {
		widths = append(widths, 40)
	}
	if r.Columns.ProjectName {
		widths = append(widths, 20)
	}
	if r.Columns.Source {
		widths = append(widths, 30)
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
func (r *AgentLineRenderer) CreateTableBuilder(styles Styles) *TableBuilder {
	tb := NewTableBuilder()
	tb.SetHeaderStyle(styles.GetPrimary())
	tb.SetBorderStyle(BorderSimple) // Simple border with separator line only

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

// GetTableColumns returns column definitions with dynamic width support
func (r *AgentLineRenderer) GetTableColumns() []AgentTableColumn {
	hiddenFalse := false

	columns := []AgentTableColumn{}

	if r.Columns.Status {
		columns = append(columns, AgentTableColumn{
			Title:  " Status  ",
			Width:  intPtr(15),
			Hidden: &hiddenFalse,
		})
	}
	if r.Columns.LastLog {
		columns = append(columns, AgentTableColumn{
			Title:  "Last Activity ",
			Width:  intPtr(15), // Fixed width for timestamp
			Hidden: &hiddenFalse,
		})
	}
	if r.Columns.Session {
		columns = append(columns, AgentTableColumn{
			Title:  "Session  ",
			Width:  intPtr(12),
			Hidden: &hiddenFalse,
		})
	}
	if r.Columns.Task {
		columns = append(columns, AgentTableColumn{
			Title:  "Description  ",
			Grow:   boolPtr(true), // This column grows to fill space
			Hidden: &hiddenFalse,
		})
	}
	if r.Columns.ProjectName {
		columns = append(columns, AgentTableColumn{
			Title:  "Project ",
			Width:  intPtr(20),
			Hidden: &hiddenFalse,
		})
	}
	if r.Columns.Source {
		columns = append(columns, AgentTableColumn{
			Title:  "Source ",
			Width:  intPtr(12),
			Hidden: &hiddenFalse,
		})
	}

	return columns
}

// RenderAgentTable renders a complete table of agents using dynamic column widths
func (r *AgentLineRenderer) RenderAgentTable(agents []domain.Agent, selectedIndex int, width int, headerState HeaderState) string {
	columns := r.GetTableColumns()

	// Render header
	header := r.renderHeaderRow(columns, width, headerState)

	// Render rows
	var rowStrings []string
	for i, agent := range agents {
		isSelected := i == selectedIndex
		rowStr := r.renderAgentRow(agent, columns, width, isSelected)
		rowStrings = append(rowStrings, rowStr)
	}

	// Join all rows
	body := lipgloss.JoinVertical(lipgloss.Left, rowStrings...)

	// Add border around entire table
	border := lipgloss.RoundedBorder()
	if !config.UseNerdFonts {
		border = lipgloss.ASCIIBorder()
	}

	tableStyle := lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color("cyan")).
		BorderTop(false).
		Width(width)

	content := header + "\n" + body
	return tableStyle.Render(content)
}

// renderHeaderRow renders the table header with dynamic column widths
func (r *AgentLineRenderer) renderHeaderRow(columns []AgentTableColumn, tableWidth int, headerState HeaderState) string {
	renderedColumns := r.computeColumnWidths(columns, tableWidth)

	// Map columns to field names for sort indicator
	columnFields := []string{"status", "lastlog", "session", "task", "project"}

	var headerCells []string
	for i, col := range columns {
		if col.Hidden != nil && *col.Hidden {
			continue
		}
		width := renderedColumns[i]

		// Add sort indicator if this column is sorted
		title := col.Title
		if i < len(columnFields) && columnFields[i] == headerState.SortField {
			if headerState.SortAscending {
				title = title + " ↑"
			} else {
				title = title + " ↓"
			}
		}

		// Highlight selected column if header is focused
		style := r.Styles.GetPrimary().Bold(true)
		if headerState.HeaderFocused && i == headerState.SelectedColumn {
			style = r.Styles.GetSelected().Bold(true)
		}

		cell := style.
			Width(width).
			MaxWidth(width).
			Render(title)
		headerCells = append(headerCells, cell)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)
}

// renderAgentRow renders a single agent row with dynamic column widths
func (r *AgentLineRenderer) renderAgentRow(agent domain.Agent, columns []AgentTableColumn, tableWidth int, isSelected bool) string {
	renderedColumns := r.computeColumnWidths(columns, tableWidth)
	cellData := r.RenderAgent(agent, isSelected)

	var cells []string
	for i, data := range cellData {
		if i >= len(renderedColumns) {
			break
		}
		width := renderedColumns[i]

		style := lipgloss.NewStyle()
		if isSelected {
			style = r.Styles.GetSelected()
		}

		cell := style.
			Width(width).
			MaxWidth(width).
			Render(data)
		cells = append(cells, cell)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// computeColumnWidths calculates actual widths for each column including growing columns
func (r *AgentLineRenderer) computeColumnWidths(columns []AgentTableColumn, tableWidth int) []int {
	widths := make([]int, len(columns))
	takenWidth := 0
	numGrowingColumns := 0

	// First pass: calculate fixed widths
	for i, col := range columns {
		if col.Hidden != nil && *col.Hidden {
			continue
		}

		if col.Grow != nil && *col.Grow {
			numGrowingColumns++
			continue
		}

		if col.Width != nil {
			widths[i] = *col.Width
			takenWidth += *col.Width
			continue
		}

		// Use title width as default
		widths[i] = lipgloss.Width(col.Title)
		takenWidth += widths[i]
	}

	// Second pass: distribute remaining width to growing columns
	if numGrowingColumns > 0 {
		leftoverWidth := tableWidth - takenWidth - 4 // Reserve 4 for border padding
		if leftoverWidth < 0 {
			leftoverWidth = 20 // Minimum width
		}
		growWidth := leftoverWidth / numGrowingColumns

		for i, col := range columns {
			if col.Grow != nil && *col.Grow {
				widths[i] = growWidth
			}
		}
	}

	return widths
}

// Helper functions
func intPtr(i int) *int          { return &i }
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

// ConfigureColumns sets which columns are visible
func (r *AgentLineRenderer) ConfigureColumns(config ColumnConfig) {
	r.Columns = config
}

// getIcon returns the appropriate icon based on font mode
func getIcon(iconType string) string {
	var icons map[string]string
	if config.UseNerdFonts {
		icons = nerdIcons
	} else {
		icons = plainIcons
	}

	if icon, exists := icons[iconType]; exists {
		return icon
	}
	return icons["default"]
}

// nerdIcons contains Nerd Font glyphs
var nerdIcons = map[string]string{
	"active":    "\uf069", // Nerd Font asterisk (nf-fa-asterisk)
	"working":   "\uf069", // Nerd Font asterisk (nf-fa-asterisk)
	"idle":      "\uf069", // Nerd Font asterisk (nf-fa-asterisk)
	"completed": "✓",      // Checkmark for completed
	"failed":    "✗",      // X for failed
	"bookmark":  "★",      // Star for bookmarks
	"pipe":      "│",      // Vertical line for hierarchy
	"default":   "·",      // Middle dot default
}

// plainIcons contains ASCII fallback characters
var plainIcons = map[string]string{
	"active":    "*",
	"working":   "@",
	"idle":      "o",
	"completed": "+",
	"failed":    "x",
	"bookmark":  "*",
	"pipe":      "|",
	"default":   ".",
}

// StatusIcon returns an icon for the given status
func StatusIcon(status string) string {
	return getIcon(status)
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
// formatLastActivity formats the last activity timestamp
func (r *AgentLineRenderer) formatLastActivity(lastActivity time.Time) string {
	if lastActivity.IsZero() {
		return r.Styles.GetSubtle().Render("-")
	}
	elapsed := time.Since(lastActivity)
	if elapsed < time.Minute {
		return "just now"
	} else if elapsed < time.Hour {
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	} else if elapsed < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(elapsed.Hours()/24))
	}
}
