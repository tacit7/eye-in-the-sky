package app

import (
	"fmt"
	"strings"
)

// renderNotesView renders the notes view with split panes
func (m *Model) renderNotesView() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Build item list for left pane
	items := make([]string, len(m.notes))
	for i, note := range m.notes {
		date := note.CreatedAt.Format("2006-01-02 15:04")
		firstLine := getFirstLine(note.Content)
		firstLine = truncate(firstLine, 50)
		items[i] = fmt.Sprintf("%s  %s", date, firstLine)
	}

	// Build detail content for right pane
	var detailContent string
	if m.notesIndex >= 0 && m.notesIndex < len(m.notes) {
		detailContent = m.renderNoteDetails(m.notes[m.notesIndex])
	}

	// Footer
	footer := m.renderFooterWithKeys("[j/k] Move  [r] Refresh  [q/esc] Back")

	return m.renderSplitPaneView(
		"Notes",
		items,
		m.notesIndex,
		len(m.notes),
		detailContent,
		footer,
	)
}

// renderNoteDetails renders the full note content with markdown syntax highlighting
func (m *Model) renderNoteDetails(note Note) string {
	var b strings.Builder

	// Timestamp
	b.WriteString(m.styles.Primary.Render("Date: "))
	b.WriteString(m.styles.Text.Render(note.CreatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Full content with markdown rendering
	b.WriteString(m.styles.Title.Render("Content"))
	b.WriteString("\n")

	// Try to render as markdown, fall back to plain text if it fails
	rendered, err := m.renderMarkdown(note.Content)
	if err != nil {
		b.WriteString(m.styles.Text.Render(note.Content))
	} else {
		b.WriteString(rendered)
	}
	b.WriteString("\n")

	return b.String()
}
