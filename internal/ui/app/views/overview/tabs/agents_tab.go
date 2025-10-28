package tabs

import (
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
)

// LayoutInfo contains layout dimensions for rendering
type LayoutInfo struct {
	Width    int
	Height   int
	ContentH int
}

// RenderAgentsTab renders the agents list tab using lipgloss table with borders
func RenderAgentsTab(agents []domain.Agent, selectedIndex int, listOffset int, styles components.Styles, layout LayoutInfo) string {
	// Create agent renderer with proper styles
	renderer := components.NewAgentLineRenderer(styles)

	// Configure columns to show
	renderer.ConfigureColumns(components.ColumnConfig{
		Icon:         false,
		Status:       true,
		ID:           true,
		Task:         true,
		Source:       true,
		Session:      true,
		LastActivity: false,
		ProjectName:  false,
	})

	// Calculate viewport
	start := listOffset
	end := listOffset + layout.ContentH
	if end > len(agents) {
		end = len(agents)
	}
	if start > len(agents) {
		start = 0
	}

	// Get visible agents slice
	visibleAgents := agents[start:end]

	// Adjust selected index to be relative to visible slice
	adjustedSelectedIndex := selectedIndex - listOffset

	// Render the table using lipgloss table with borders
	return renderer.RenderAgentTable(visibleAgents, adjustedSelectedIndex)
}
