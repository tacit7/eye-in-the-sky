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
func RenderAgentsTab(agents []domain.Agent, selectedIndex int, listOffset int, styles components.Styles, layout LayoutInfo, headerState components.HeaderState) string {
	// Create agent renderer with proper styles
	renderer := components.NewAgentLineRenderer(styles)

	// Configure columns to show
	renderer.ConfigureColumns(components.ColumnConfig{
		Icon:         false,
		Status:       true,
		Task:         true,
		Source:       false,
		Session:      true,
		LastActivity: false,
		ProjectName:  true,
		LastLog:      true,
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

	// Render the table using dynamic column system
	return renderer.RenderAgentTable(visibleAgents, adjustedSelectedIndex, layout.Width, headerState)
}
