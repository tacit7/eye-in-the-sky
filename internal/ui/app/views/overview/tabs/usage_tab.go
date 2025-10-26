package tabs

import (
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/shared"
)

// RenderUsageTab renders the token usage metrics tab using the viewport from root Model
func RenderUsageTab(viewport shared.ViewportAccess, styles shared.Styles) string {
	if viewport == nil {
		return styles.RenderSubtle("  No usage data available")
	}

	// Get viewport view
	view := viewport.View()

	// Add scroll indicators
	var b strings.Builder

	// Top indicator
	if viewport.AtTop() {
		b.WriteString("\n")
	} else {
		b.WriteString(styles.RenderSubtle("▲") + "\n")
	}

	b.WriteString(view)

	// Bottom indicator
	if !viewport.AtBottom() {
		b.WriteString("\n" + styles.RenderSubtle("▼"))
	}

	return b.String()
}
