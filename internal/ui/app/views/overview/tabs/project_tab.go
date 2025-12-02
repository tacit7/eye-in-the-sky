package tabs

import (
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/shared"
)

// RenderProjectTab renders the project tasks tab using the viewport from root Model
func RenderProjectTab(viewport shared.ViewportAccess, styles shared.Styles) string {
	if viewport == nil {
		return styles.RenderSubtle("  No project data available")
	}

	var b strings.Builder

	// Title
	title := "📋 Project Tasks"
	b.WriteString(styles.RenderPrimary(title))
	b.WriteString("\n")

	// Instructions
	b.WriteString(styles.RenderSubtle("Press j/k or ↑↓ to scroll, PageUp/PageDown for faster scrolling.\n"))

	// Top scroll indicator
	if viewport.AtTop() {
		b.WriteString("\n")
	} else {
		b.WriteString(styles.RenderSubtle("▲") + "\n")
	}

	// Viewport content
	b.WriteString(viewport.View())

	// Bottom scroll indicator
	if !viewport.AtBottom() {
		b.WriteString("\n" + styles.RenderSubtle("▼"))
	}

	return b.String()
}
