package tabs

import (
	"fmt"
	"strings"
)

// RenderNotes renders the notes tab content with markdown support
func RenderNotes(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.Notes) == 0 {
		return overviewStyles.Subtle.Render("\n  No notes available for this agent\n")
	}

	var sb strings.Builder

	// Header with columns: Time | Scope | Content Preview
	sb.WriteString(overviewStyles.Primary.Render("Time        Scope        Content"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// Render notes with scope indication
	for _, note := range ctx.Notes {
		// Timestamp (HH:MM:SS format)
		timestamp := note.CreatedAt.Format("15:04:05")

		// Scope indicator (from parent_type)
		scope := getScopeLabel(note.ParentType)

		// Content preview (first line, max 45 chars)
		preview := strings.Split(note.Body, "\n")[0]
		if len(preview) > 45 {
			preview = preview[:42] + "..."
		}

		line := fmt.Sprintf("%-11s %-12s %s\n", timestamp, scope, preview)
		sb.WriteString(line)
	}

	return sb.String()
}

// getScopeLabel returns a human-readable label for the note scope
func getScopeLabel(parentType string) string {
	switch parentType {
	case "global":
		return "🌐 Global"
	case "projects":
		return "📁 Project"
	case "agents":
		return "🤖 Agent"
	case "sessions":
		return "💼 Session"
	default:
		return "❓ Unknown"
	}
}

// TODO: Split-pane view with note details requires state migration
/*
// renderNoteDetails renders details for a single note
func renderNoteDetails(note domain.Note, styles OverviewStyles) string {
	var b strings.Builder

	// Note header
	b.WriteString(styles.Primary.Render("Created: "))
	b.WriteString(note.CreatedAt.Format("2006-01-02 15:04:05"))
	b.WriteString("\n\n")

	// Content
	b.WriteString(note.Content)

	return b.String()
}
*/
