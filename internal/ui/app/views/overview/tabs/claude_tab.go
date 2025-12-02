package tabs

import (
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/shared"
)

// RenderClaudeTab renders the Claude configuration files tab using the viewport from root Model
func RenderClaudeTab(viewport shared.ViewportAccess, styles shared.Styles) string {
	if viewport == nil {
		return styles.RenderSubtle("  No Claude config data available")
	}

	// Render Claude config using viewport
	// The viewport already has the full two-pane layout from view_claude.go
	return viewport.View()
}
