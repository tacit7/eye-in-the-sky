package tabs

import (
	"fmt"
	"os/exec"
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
		// Get git worktree path from agent
		gitPath := ""
		if ctx.Agent != nil {
			gitPath = ctx.Agent.GitWorktreePath
		}
		rightPane = renderCommitDetails(ctx.Commits[selectedIndex], rightWidth, overviewStyles, gitPath)
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

		// Hash (first 6 chars)
		hash := string(commit.Hash)
		if len(hash) > 6 {
			hash = hash[:6]
		}

		// Message preview (first line only)
		message := commit.Message
		lines := strings.Split(message, "\n")
		if len(lines) > 0 {
			message = strings.TrimSpace(lines[0])
		}

		// Calculate max message length
		maxLen := width - len(hash) - 5 // 5 for padding and spacing
		if len(message) > maxLen {
			message = message[:maxLen-3] + "..."
		}

		// Build line: hash message
		line := fmt.Sprintf("  %-6s %s", hash, message)

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
func renderCommitDetails(commit domain.Commit, width int, styles OverviewStyles, gitPath string) string {
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

	// Diff section
	if gitPath != "" {
		diff := getCommitDiff(gitPath, string(commit.Hash))
		if diff != "" {
			sb.WriteString("\n")
			sb.WriteString(styles.SectionTitle.Render("Changes"))
			sb.WriteString("\n")
			sb.WriteString(theme.TextMuted.Render(strings.Repeat("─", width-4)))
			sb.WriteString("\n\n")
			sb.WriteString(colorizeGitDiff(diff, width-4))
		}
	}

	return theme.PanelNoBorder.Copy().Width(width).Render(sb.String())
}

// getCommitDiff fetches the git diff for a specific commit
func getCommitDiff(gitPath string, commitHash string) string {
	if gitPath == "" || commitHash == "" {
		return ""
	}

	cmd := exec.Command("git", "-C", gitPath, "show", "--no-color", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(output)
}

// colorizeGitDiff adds color to git diff output
func colorizeGitDiff(diff string, maxWidth int) string {
	var sb strings.Builder
	lines := strings.Split(diff, "\n")

	for _, line := range lines {
		// Truncate long lines
		if len(line) > maxWidth {
			line = line[:maxWidth-3] + "..."
		}

		// Color based on diff markers
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			sb.WriteString(theme.TextSubtitle.Render(line))
		case strings.HasPrefix(line, "+"):
			sb.WriteString(theme.TextSuccess.Render(line))
		case strings.HasPrefix(line, "-"):
			sb.WriteString(theme.TextError.Render(line))
		case strings.HasPrefix(line, "@@"):
			sb.WriteString(theme.TextHighlight.Render(line))
		case strings.HasPrefix(line, "diff --git"):
			sb.WriteString(theme.TextTitle.Render(line))
		default:
			sb.WriteString(theme.TextMuted.Render(line))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
