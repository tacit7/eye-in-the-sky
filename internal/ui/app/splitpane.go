package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SplitPane represents a split-pane layout with a list on the left and details on the right
type SplitPane struct {
	LeftItems      []string
	SelectedIndex  int
	RightContent   string
	Styles         Styles
	Width          int
}

// View renders the split-pane content
func (p SplitPane) View() string {
	// Calculate dimensions
	totalWidth := p.Width - 6
	leftPaneWidth := totalWidth / 3
	// rightPaneWidth := totalWidth - leftPaneWidth - 1 // -1 for separator (unused)

	// Build left pane (list)
	var leftPane strings.Builder
	for i, item := range p.LeftItems {
		if i == p.SelectedIndex {
			leftPane.WriteString(p.Styles.Selected.Render(truncate(item, leftPaneWidth-2)))
		} else {
			leftPane.WriteString(truncate(item, leftPaneWidth-2))
		}
		if i < len(p.LeftItems)-1 {
			leftPane.WriteString("\n")
		}
	}

	// Build right pane (details)
	rightPane := p.RightContent
	if rightPane == "" {
		rightPane = p.Styles.Subtle.Render("Select an item to view details")
	}

	// Combine panes side by side
	leftLines := strings.Split(leftPane.String(), "\n")
	rightLines := strings.Split(rightPane, "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var combined strings.Builder
	for i := 0; i < maxLines; i++ {
		// Left pane line
		if i < len(leftLines) {
			line := leftLines[i]
			combined.WriteString(line)
			// Pad to fill width
			padding := leftPaneWidth - lipgloss.Width(line)
			if padding > 0 {
				combined.WriteString(strings.Repeat(" ", padding))
			}
		} else {
			combined.WriteString(strings.Repeat(" ", leftPaneWidth))
		}

		// Separator
		combined.WriteString(p.Styles.Border.Render("│"))

		// Right pane line
		if i < len(rightLines) {
			combined.WriteString(" " + rightLines[i])
		}

		if i < maxLines-1 {
			combined.WriteString("\n")
		}
	}

	return combined.String()
}
