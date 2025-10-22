package app

import (
	"fmt"
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderNotesTabView renders the notes tab with split-pane view
func (m *Model) renderNotesTabView() string {
	if len(m.notes) == 0 {
		return m.styles.Subtle.Render("No notes available")
	}

	// Build item list for left pane
	items := make([]string, len(m.notes))
	for i, note := range m.notes {
		timestamp := note.CreatedAt.Format("15:04:05")
		preview := truncate(note.Content, 30)
		items[i] = fmt.Sprintf("%s  %s", timestamp, preview)
	}

	// Build detail content for right pane
	var detailContent string
	if m.notesIndex >= 0 && m.notesIndex < len(m.notes) {
		detailContent = renderNoteDetails(m.notes[m.notesIndex], m.styles)
	}

	return SplitPane{
		LeftItems:     items,
		SelectedIndex: m.notesIndex,
		RightContent:  detailContent,
		Styles:        m.styles,
		Width:         m.width,
	}.View()
}

// renderNoteDetails renders details for a single note
func renderNoteDetails(note domain.Note, styles Styles) string {
	var b strings.Builder

	// Note header
	b.WriteString(styles.Primary.Render("Created: "))
	b.WriteString(styles.Value.Render(note.CreatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Content
	b.WriteString(styles.Text.Render(note.Content))

	return b.String()
}
