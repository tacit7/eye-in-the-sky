package tabs

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// noteItem wraps a domain.Note to implement list.Item
type noteItem struct {
	note domain.Note
}

func (i noteItem) FilterValue() string {
	return i.note.Body
}

func (i noteItem) Title() string {
	return i.note.CreatedAt.Format("Mon, Jan 2, 3:04 PM")
}

func (i noteItem) Description() string {
	lines := strings.Split(i.note.Body, "\n")
	if len(lines) > 0 {
		preview := strings.TrimSpace(lines[0])
		// Remove markdown headers
		preview = strings.TrimPrefix(preview, "# ")
		preview = strings.TrimPrefix(preview, "## ")
		preview = strings.TrimPrefix(preview, "### ")

		maxLen := 60
		if len(preview) > maxLen {
			preview = preview[:maxLen-3] + "..."
		}
		return preview
	}
	return ""
}

// noteDelegate is a custom list delegate for notes
type noteDelegate struct{}

func (d noteDelegate) Height() int { return 2 }
func (d noteDelegate) Spacing() int { return 1 }
func (d noteDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d noteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(noteItem)
	if !ok {
		return
	}

	var str string

	// Use list model's selected index
	if index == m.Index() {
		str = theme.ListItemSelected.Render(fmt.Sprintf("> %s\n  %s", i.Title(), i.Description()))
	} else {
		str = theme.ListItem.Render(fmt.Sprintf("  %s\n  %s", i.Title(), i.Description()))
	}

	fmt.Fprint(w, str)
}

// RenderNotes renders the notes tab with split-pane view
func RenderNotes(ctx *DataContext, overviewStyles OverviewStyles) string {
	return RenderNotesSplitPane(ctx, overviewStyles, ctx.NotesIndex, ctx.Width)
}

// RenderNotesSplitPane renders notes in split-pane layout using bubbles/list
func RenderNotesSplitPane(ctx *DataContext, overviewStyles OverviewStyles, selectedIndex int, width int) string {
	if len(ctx.Notes) == 0 {
		return theme.TextMuted.Render("\n  No notes available for this session\n")
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	// Convert notes to list items
	items := make([]list.Item, len(ctx.Notes))
	for i, note := range ctx.Notes {
		items[i] = noteItem{note: note}
	}

	// Create list model
	delegate := noteDelegate{}
	noteList := list.New(items, delegate, leftWidth, 30)
	noteList.SetShowStatusBar(false)
	noteList.SetShowTitle(false)
	noteList.SetFilteringEnabled(false)
	noteList.SetShowHelp(false)

	// Set selected index
	if selectedIndex >= 0 && selectedIndex < len(items) {
		noteList.Select(selectedIndex)
	}

	// Render left pane (note list)
	leftPane := noteList.View()

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
