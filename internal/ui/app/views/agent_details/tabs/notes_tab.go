package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// RenderNotes renders the notes tab with split-pane view
func RenderNotes(ctx *DataContext, overviewStyles OverviewStyles) string {
	return RenderNotesSplitPane(ctx, overviewStyles, ctx.NotesIndex, ctx.Width)
}

// RenderNotesSplitPane renders notes in split-pane layout
func RenderNotesSplitPane(ctx *DataContext, overviewStyles OverviewStyles, selectedIndex int, width int) string {
	if len(ctx.Notes) == 0 {
		return theme.TextMuted.Render("\n  No notes available for this session\n")
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	// Render left pane (note list)
	leftPane := renderNoteList(ctx.Notes, selectedIndex, leftWidth)

	// Render right pane (selected note details)
	rightPane := ""
	if selectedIndex >= 0 && selectedIndex < len(ctx.Notes) {
		rightPane = renderNoteDetails(ctx.Notes[selectedIndex], rightWidth, ctx.MarkdownRenderer)
	} else {
		rightPane = theme.TextMuted.Copy().
			Width(rightWidth).
			Render("Select a note to view details")
	}

	// Join panes horizontally
	result := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		theme.PanelSidebar.Copy().Height(30).Render(""),
		rightPane,
	)

	return result
}

// renderNoteList renders the left pane note list
func renderNoteList(notes []domain.Note, selectedIndex int, width int) string {
	var sb strings.Builder

	// Note items
	for i, note := range notes {
		isSelected := i == selectedIndex

		// Cursor indicator
		cursor := "  "
		if isSelected {
			cursor = "> "
		}

		// Timestamp: "Mon, Jan 2, 3:04 PM"
		timestamp := note.CreatedAt.Format("Mon, Jan 2, 3:04 PM")

		// First line preview
		lines := strings.Split(note.Body, "\n")
		preview := ""
		if len(lines) > 0 {
			preview = strings.TrimSpace(lines[0])
			// Remove markdown headers
			preview = strings.TrimPrefix(preview, "# ")
			preview = strings.TrimPrefix(preview, "## ")
			preview = strings.TrimPrefix(preview, "### ")

			maxLen := width - len(cursor) - len(timestamp) - 5
			if len(preview) > maxLen {
				preview = preview[:maxLen-3] + "..."
			}
		}

		// Build line: cursor timestamp preview
		line := fmt.Sprintf("%s%-23s %s", cursor, timestamp, preview)

		// Apply style
		if isSelected {
			line = theme.ListItemSelected.Copy().Width(width - 2).Render(line)
		} else {
			line = theme.ListItem.Copy().Width(width - 2).Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return theme.List.Copy().Width(width).Render(sb.String())
}

// renderNoteDetails renders the right pane with full note content
func renderNoteDetails(note domain.Note, width int, mdRenderer MarkdownRenderer) string {
	var sb strings.Builder

	// Note header (timestamp)
	timestamp := note.CreatedAt.Format("Monday, January 2, 2006 at 3:04 PM")
	sb.WriteString(theme.TextSubtitle.Render(timestamp))
	sb.WriteString("\n\n")

	// Render markdown if renderer available, otherwise plain text
	renderedBody := note.Body
	if mdRenderer != nil {
		rendered, err := mdRenderer.Render(note.Body)
		if err == nil {
			renderedBody = rendered
		}
		// On error, fall back to plain text
	}

	sb.WriteString(renderedBody)

	return theme.PanelNoBorder.Copy().Width(width).Render(sb.String())
}
