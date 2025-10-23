package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderClaudeTab renders the Claude config browser tab
func (m *Model) renderClaudeTab() string {
	if len(m.claudeFiles) == 0 {
		return m.renderClaudeEmptyState()
	}

	// Calculate dimensions for two-pane layout
	// Use simple division, no boxes (parent handles that)
	leftPaneWidth := (m.width - 6) / 3
	if leftPaneWidth < 25 {
		leftPaneWidth = 25
	}
	rightPaneWidth := m.width - leftPaneWidth - 10

	// Render left pane (file list)
	leftPane := m.renderClaudeFileList()

	// Render right pane (content viewer or instructions)
	rightPane := m.renderClaudeContentViewer()

	// Style the panes without borders
	leftStyle := lipgloss.NewStyle().
		Width(leftPaneWidth).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Width(rightPaneWidth).
		Padding(0, 1)

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Colors.Border)).
		Render(" │ ")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(leftPane),
		separator,
		rightStyle.Render(rightPane),
	)
}

// renderClaudeFileList renders the file list on the left pane
func (m *Model) renderClaudeFileList() string {
	var lines []string

	// Title
	title := m.styles.SectionTitle.Render("📁 ~/.claude")
	lines = append(lines, title)
	lines = append(lines, "")

	// Render files
	for i, file := range m.claudeFiles {
		var line string

		// Selection indicator
		if i == m.claudeSelectedIndex {
			// Selected file
			icon := "📄"
			if file.IsDir {
				icon = "📁"
			}
			line = m.styles.Selected.Render(fmt.Sprintf(" %s %s", icon, file.Name))
		} else {
			// Unselected file
			icon := "  "
			if file.IsDir {
				icon = "📁"
			} else {
				icon = "📄"
			}
			line = m.styles.Text.Render(fmt.Sprintf(" %s %s", icon, file.Name))
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderClaudeContentViewer renders the content viewer on the right pane
func (m *Model) renderClaudeContentViewer() string {
	if !m.claudeShowingContent || m.claudeContent == "" {
		return m.renderClaudeInstructions()
	}

	// Show validation status if available
	var header string
	if m.claudeValidStatus != "" {
		statusStyle := m.styles.Success
		if strings.Contains(m.claudeValidStatus, "Invalid") {
			statusStyle = m.styles.Error
		}
		header = statusStyle.Render(m.claudeValidStatus) + "\n\n"
	}

	// Return viewport content with header
	return header + m.claudeViewport.View()
}

// renderClaudeInstructions renders help text when no file is selected
func (m *Model) renderClaudeInstructions() string {
	instructions := []string{
		m.styles.Title.Render("Claude Config Browser"),
		"",
		m.styles.Text.Render("Navigate and view Claude Code configuration files."),
		"",
		m.styles.SectionTitle.Render("Keybindings:"),
		"",
		m.styles.Label.Render("  j/k or ↓/↑  ") + m.styles.Subtle.Render("Navigate file list"),
		m.styles.Label.Render("  Enter       ") + m.styles.Subtle.Render("Load and view file"),
		m.styles.Label.Render("  v           ") + m.styles.Subtle.Render("Validate JSON"),
		m.styles.Label.Render("  e           ") + m.styles.Subtle.Render("Edit in $EDITOR"),
		m.styles.Label.Render("  r           ") + m.styles.Subtle.Render("Refresh file list"),
		m.styles.Label.Render("  Esc/q       ") + m.styles.Subtle.Render("Back to file list"),
		"",
		m.styles.Subtle.Render("Select a file to get started."),
	}

	return strings.Join(instructions, "\n")
}

// renderClaudeEmptyState renders the empty state when no files found
func (m *Model) renderClaudeEmptyState() string {
	emptyMsg := []string{
		"",
		"",
		m.styles.Subtle.Render("  No files found in ~/.claude"),
		"",
		m.styles.Subtle.Render("  The directory may not exist or is empty."),
		"",
		m.styles.Subtle.Render("  Press 'r' to refresh."),
	}

	return strings.Join(emptyMsg, "\n")
}
