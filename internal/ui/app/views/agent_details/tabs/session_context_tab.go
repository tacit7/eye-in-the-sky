package tabs

import (
	"fmt"
	"strings"
)

// RenderSessionContext renders the session context tab content
// After migration 002, context is stored as markdown instead of structured fields
func RenderSessionContext(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.SessionContexts) == 0 {
		return overviewStyles.Subtle.Render("\n  No session context saved for this agent\n")
	}

	var sb strings.Builder

	// Display most recent context
	context := ctx.SessionContexts[0]

	// Header
	sb.WriteString(overviewStyles.Primary.Render("Session Context"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", ctx.Width-4)))
	sb.WriteString("\n\n")

	// Render markdown context
	if context.Context != "" {
		// Use markdown renderer if available, otherwise fallback to plain text with indentation
		renderedContext := context.Context
		if ctx.MarkdownRenderer != nil {
			rendered, err := ctx.MarkdownRenderer.Render(context.Context)
			if err == nil {
				renderedContext = rendered
			}
		} else {
			// Fallback: add indentation to each line for better formatting
			lines := strings.Split(context.Context, "\n")
			renderedContext = ""
			for _, line := range lines {
				renderedContext += "  " + line + "\n"
			}
		}
		sb.WriteString(renderedContext)
		sb.WriteString("\n")
	} else {
		sb.WriteString(overviewStyles.Subtle.Render("  (No context available)"))
		sb.WriteString("\n\n")
	}

	// Metadata
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", ctx.Width-4)))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(fmt.Sprintf("Session: %s", context.SessionID)))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(fmt.Sprintf("Last updated: %s", context.UpdatedAt.Format("2006-01-02 15:04:05"))))
	sb.WriteString("\n")

	return sb.String()
}
