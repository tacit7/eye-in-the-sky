package tabs

import (
	"github.com/charmbracelet/bubbles/table"
)

// RenderAgentsTab renders the agents list tab using Bubbles table.Model
func RenderAgentsTab(agentsTable table.Model) string {
	// Simply render the table - all styling and data is already configured
	return agentsTable.View()
}
