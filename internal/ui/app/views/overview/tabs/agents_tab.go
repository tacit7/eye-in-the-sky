package tabs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
)

// LayoutInfo contains layout dimensions for rendering
type LayoutInfo struct {
	Width    int
	Height   int
	ContentH int
}

// RenderAgentsTab renders the agents list tab using the proper AgentLineRenderer + TableBuilder system
// This ensures consistent lipgloss styling with the old code
func RenderAgentsTab(agents []domain.Agent, selectedIndex int, listOffset int, styles components.Styles, layout LayoutInfo) string {
	// Create agent renderer with proper styles
	renderer := components.NewAgentLineRenderer(styles)

	// Configure columns to show
	renderer.ConfigureColumns(components.ColumnConfig{
		Icon:    true,
		Status:  true,
		ID:      true,
		Task:    true,
		Source:  true,
		Session: true,
	})

	// Create table builder
	tb := renderer.CreateTableBuilder(styles)

	// Calculate viewport
	start := listOffset
	end := listOffset + layout.ContentH
	if end > len(agents) {
		end = len(agents)
	}
	if start > len(agents) {
		start = 0
	}

	// Set row styling for selection and alternating rows
	tb.SetRowStyleFunc(func(rowIndex int) lipgloss.Style {
		actualIndex := listOffset + rowIndex
		if actualIndex == selectedIndex {
			return styles.GetSelected()
		}
		// Alternating row colors for better readability
		if rowIndex%2 == 1 {
			return lipgloss.NewStyle().Background(lipgloss.Color("233"))
		}
		return lipgloss.NewStyle()
	})

	// Convert agents to table rows using renderer
	var rows [][]string
	for i := start; i < end && i < len(agents); i++ {
		row := renderer.RenderAgent(agents[i], false)
		rows = append(rows, row)
	}

	// Render the table
	result, _ := tb.RenderTable(rows)
	return result
}
