package tabs

import (
	"fmt"
	"strings"
)

// RenderNotes renders the notes tab content with timestamp, title, and content preview
func RenderNotes(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.Notes) == 0 {
		return overviewStyles.Subtle.Render("\n  No notes available for this agent\n")
	}

	var sb strings.Builder

	// Header with three columns: Timestamp | Title | Content
	sb.WriteString(overviewStyles.Primary.Render("Time        Title                 Content"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// Render notes in three-column format
	for _, note := range ctx.Notes {
		// Timestamp (HH:MM:SS format)
		timestamp := note.CreatedAt.Format("15:04:05")

		// Title (max 20 chars)
		title := note.Title
		if len(title) > 20 {
			title = title[:17] + "..."
		}

		// Content preview (max 35 chars)
		preview := note.Content
		if len(preview) > 35 {
			preview = preview[:32] + "..."
		}

		line := fmt.Sprintf("%-11s %-21s %s\n", timestamp, title, preview)
		sb.WriteString(line)
	}

	return sb.String()
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
