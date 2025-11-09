package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// RenderCommits renders the commits tab with split-pane view
func RenderCommits(ctx *DataContext, overviewStyles OverviewStyles) string {
	return RenderCommitsSplitPane(ctx, overviewStyles, ctx.CommitsIndex, ctx.Width)
}

// RenderCommitsSplitPane renders commits in split-pane layout
func RenderCommitsSplitPane(ctx *DataContext, overviewStyles OverviewStyles, selectedIndex int, width int) string {
	if len(ctx.Commits) == 0 {
		return theme.TextMuted.Render("\n  No commits found for this agent\n")
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	// Render left pane (commit list)
	leftPane := renderCommitList(ctx.Commits, selectedIndex, leftWidth)

	// Render right pane (selected commit details)
	rightPane := ""
	if selectedIndex >= 0 && selectedIndex < len(ctx.Commits) {
		rightPane = renderCommitDetails(ctx.Commits[selectedIndex], rightWidth, overviewStyles)
	} else {
		rightPane = theme.TextMuted.Copy().
			Width(rightWidth).
			Render("Select a commit to view details")
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

// renderCommitList renders the left pane commit list
func renderCommitList(commits []domain.Commit, selectedIndex int, width int) string {
	var sb strings.Builder

	// Commit items
	for i, commit := range commits {
		isSelected := i == selectedIndex

		// Date and hash
		date := commit.Timestamp.Format("2006-01-02")
		hash := string(commit.Hash)
		if len(hash) > 8 {
			hash = hash[:8]
		}

		// Message preview
		message := commit.Message
		// Get first line only
		lines := strings.Split(message, "\n")
		if len(lines) > 0 {
			message = strings.TrimSpace(lines[0])
		}

		maxLen := width - len(date) - len(hash) - 7
		if len(message) > maxLen {
			message = message[:maxLen-3] + "..."
		}

		// Build line: date hash message
		line := fmt.Sprintf("  %-11s %-8s %s", date, hash, message)

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

// renderCommitDetails renders the right pane with full commit details
func renderCommitDetails(commit domain.Commit, width int, styles OverviewStyles) string {
	var sb strings.Builder

	// Commit hash
	sb.WriteString(styles.Git.Render("commit "))
	sb.WriteString(styles.Value.Render(string(commit.Hash)))
	sb.WriteString("\n")

	// Author info (if available)
	if commit.Author != "" {
		sb.WriteString(styles.Label.Render("Author: "))
		sb.WriteString(styles.Value.Render(commit.Author))
		sb.WriteString("\n")
	}

	// Date
	sb.WriteString(styles.Label.Render("Date:   "))
	sb.WriteString(styles.Value.Render(commit.Timestamp.Format("Mon Jan 2 15:04:05 2006 MST")))
	sb.WriteString("\n\n")

	// Message
	sb.WriteString(styles.Primary.Render(commit.Message))
	sb.WriteString("\n")

	// Files changed (if available)
	if len(commit.FilesChanged) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styles.SectionTitle.Render("Files Changed"))
		sb.WriteString("\n\n")
		for _, file := range commit.FilesChanged {
			sb.WriteString(fmt.Sprintf("  %s\n", file))
		}
	}

	return theme.PanelNoBorder.Copy().Width(width).Render(sb.String())
}
