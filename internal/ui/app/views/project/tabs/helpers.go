package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// DataContext wraps all data needed by tabs (avoids circular import)
type DataContext struct {
	Project            *database.Project
	Agents             []domain.Agent    // Active agents only
	Notes              []domain.Note
	Tasks              []domain.Task
	TaskNotes          []domain.TaskNote // Annotations for selected task
	SelectedAgentIndex int               // Index of selected agent in agents list
	NotesIndex         int               // Index of selected note in notes list
	TasksIndex         int               // Index of selected task in tasks list
	RightPaneOffset    int               // Scroll offset for right pane
	Width              int               // Terminal width for responsive layout
	MarkdownRenderer   MarkdownRenderer  // Markdown renderer for formatting
}

// MarkdownRenderer is an interface for rendering markdown
type MarkdownRenderer interface {
	Render(in string) (string, error)
}

// OverviewStyles contains styles for project detail tabs
type OverviewStyles struct {
	SectionTitle lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Subtle       lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Primary      lipgloss.Style
	Secondary    lipgloss.Style
	Git          lipgloss.Style
	Error        lipgloss.Style
}

// filterNonEmpty removes empty strings from a slice
func filterNonEmpty(sections []string) []string {
	var result []string
	for _, s := range sections {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

// renderLabelValue renders a label-value pair with consistent formatting
func renderLabelValue(label, value string, valueStyle lipgloss.Style, labelStyle lipgloss.Style) string {
	return fmt.Sprintf("  %s %s\n",
		labelStyle.Render(label),
		valueStyle.Render(value))
}

// formatTimestamp formats a time with standard format
func formatTimestamp(t time.Time) string {
	return t.Format("Jan 2, 2006 15:04:05 MST")
}

// formatTime formats a time with short format (HH:MM:SS)
func formatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
