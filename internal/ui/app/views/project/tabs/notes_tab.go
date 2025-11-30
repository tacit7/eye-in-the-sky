package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// RenderNotes renders the notes tab with split-pane: table (left) + details (right)
func RenderNotes(ctx *DataContext, styles OverviewStyles, notesTable table.Model, width, height int) string {
	if ctx == nil || len(ctx.Notes) == 0 {
		return theme.TextMuted.Render("\n  No notes for this project\n")
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	// Left pane: table.Model view
	leftPane := renderNotesTable(notesTable, leftWidth, height)

	// Right pane: selected note details
	rightPane := renderNoteDetails(ctx, rightWidth, height, styles)

	// Join panes horizontally with separator
	result := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		theme.PanelSidebar.Copy().Height(height).Render(""),
		rightPane,
	)

	return result
}

// renderNotesTable renders the left pane table
func renderNotesTable(t table.Model, width, height int) string {
	return theme.PanelNormal.Copy().
		Width(width).
		Height(height).
		Render(t.View())
}

// renderNoteDetails renders the right pane with selected note info
func renderNoteDetails(ctx *DataContext, width, height int, styles OverviewStyles) string {
	if ctx.NotesIndex < 0 || ctx.NotesIndex >= len(ctx.Notes) {
		return theme.PanelNoBorder.Copy().
			Width(width).
			Height(height).
			Render(theme.TextMuted.Render("Select a note to view details"))
	}

	note := ctx.Notes[ctx.NotesIndex]

	var details strings.Builder

	// Header
	details.WriteString(styles.SectionTitle.Render("Note Details") + "\n\n")

	// Note info
	details.WriteString(renderNoteField("ID", string(note.ID), styles))
	details.WriteString(renderNoteField("Parent Type", note.ParentType, styles))
	details.WriteString(renderNoteField("Parent ID", note.ParentID, styles))

	if !note.CreatedAt.IsZero() {
		details.WriteString(renderNoteField("Created", note.CreatedAt.Format("2006-01-02 15:04:05"), styles))
	}

	// Body section with markdown rendering
	details.WriteString("\n" + styles.Label.Render("Body:") + "\n")
	if note.Body != "" {
		// Render markdown if available
		if ctx.MarkdownRenderer != nil {
			rendered, err := ctx.MarkdownRenderer.Render(note.Body)
			if err == nil {
				details.WriteString(rendered)
			} else {
				details.WriteString(styles.Value.Render(note.Body))
			}
		} else {
			details.WriteString(styles.Value.Render(note.Body))
		}
	} else {
		details.WriteString(theme.TextMuted.Render("(empty)"))
	}

	return theme.PanelNoBorder.Copy().
		Width(width).
		Height(height).
		Render(details.String())
}

// renderNoteField renders a label-value pair for notes
func renderNoteField(label, value string, styles OverviewStyles) string {
	if value == "" {
		value = theme.TextMuted.Render("(none)")
	}
	return fmt.Sprintf("%s %s\n", styles.Label.Render(label+":"), styles.Value.Render(value))
}
