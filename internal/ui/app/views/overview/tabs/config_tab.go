package tabs

import (
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/shared"
)

// RenderConfigTab renders the keybindings configuration tab using the viewport from root Model
func RenderConfigTab(viewport shared.ViewportAccess, styles shared.Styles) string {
	if viewport == nil {
		return styles.RenderSubtle("  No keybindings data available")
	}

	var b strings.Builder

	// Title
	title := "⚙️  Keybindings Configuration"
	b.WriteString(styles.RenderPrimary(title))
	b.WriteString("\n")

	// Instructions
	b.WriteString(styles.RenderSubtle("Press 'e' to edit, 'r' to reload, j/k or ↑↓ to scroll.\n"))

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
