package tabs

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// commitItem wraps a domain.Commit to implement list.Item
type commitItem struct {
	commit domain.Commit
}

func (i commitItem) FilterValue() string {
	return string(i.commit.Hash) + " " + i.commit.Message
}

func (i commitItem) Title() string {
	hash := string(i.commit.Hash)
	if len(hash) > 7 {
		hash = hash[:7]
	}
	return hash
}

func (i commitItem) Description() string {
	// Get first line of commit message
	lines := strings.Split(i.commit.Message, "\n")
	if len(lines) > 0 {
		message := strings.TrimSpace(lines[0])
		maxLen := 60
		if len(message) > maxLen {
			message = message[:maxLen-3] + "..."
		}
		return message
	}
	return ""
}

// commitDelegate is a custom list delegate for commits
type commitDelegate struct{}

func (d commitDelegate) Height() int { return 1 }
func (d commitDelegate) Spacing() int { return 0 }
func (d commitDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d commitDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(commitItem)
	if !ok {
		return
	}

	var str string
	hash := i.Title()
	desc := i.Description()

	// Use list model's selected index
	if index == m.Index() {
		str = theme.ListItemSelected.Render(fmt.Sprintf("> %-7s %s", hash, desc))
	} else {
		str = theme.ListItem.Render(fmt.Sprintf("  %-7s %s", hash, desc))
	}

	fmt.Fprint(w, str)
}

// RenderCommits renders the commits tab with split-pane view
func RenderCommits(ctx *DataContext, overviewStyles OverviewStyles) string {
	return RenderCommitsSplitPane(ctx, overviewStyles, ctx.CommitsIndex, ctx.Width)
}

// RenderCommitsSplitPane renders commits in split-pane layout using bubbles/list and viewport
func RenderCommitsSplitPane(ctx *DataContext, overviewStyles OverviewStyles, selectedIndex int, width int) string {
	if len(ctx.Commits) == 0 {
		return theme.TextMuted.Render("\n  No commits found for this agent\n")
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1
	viewportHeight := 30

	// Convert commits to list items
	items := make([]list.Item, len(ctx.Commits))
	for i, commit := range ctx.Commits {
		items[i] = commitItem{commit: commit}
	}

	// Create list model
	delegate := commitDelegate{}
	commitList := list.New(items, delegate, leftWidth, viewportHeight)
	commitList.SetShowStatusBar(false)
	commitList.SetShowTitle(false)
	commitList.SetFilteringEnabled(false)
	commitList.SetShowHelp(false)

	// Set selected index
	if selectedIndex >= 0 && selectedIndex < len(items) {
		commitList.Select(selectedIndex)
	}

	// Render left pane (commit list)
	leftPane := commitList.View()

	// Render right pane (selected commit details with viewport)
	rightPane := ""
	if selectedIndex >= 0 && selectedIndex < len(ctx.Commits) {
		// Get git worktree path from agent
		gitPath := ""
		if ctx.Agent != nil {
			gitPath = ctx.Agent.GitWorktreePath
		}

		// Get commit details content
		content := renderCommitDetailsContent(ctx.Commits[selectedIndex], rightWidth, overviewStyles, gitPath)

		// Update viewport size and content
		ctx.CommitDetailViewport.Width = rightWidth
		ctx.CommitDetailViewport.Height = viewportHeight
		ctx.CommitDetailViewport.SetContent(content)

		// Render viewport
		rightPane = ctx.CommitDetailViewport.View()
	} else {
		rightPane = theme.TextMuted.Copy().
			Width(rightWidth).
			Render("Select a commit to view details")
	}

	// Join panes horizontally
	result := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		theme.PanelSidebar.Copy().Height(viewportHeight).Render(""),
		rightPane,
	)

	return result
}

// renderCommitDetailsContent renders commit details content for viewport
func renderCommitDetailsContent(commit domain.Commit, width int, styles OverviewStyles, gitPath string) string {
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
			sb.WriteString(colorizeGitDiff(diff))
		}
	}

	return sb.String()
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

// colorizeGitDiff adds color to git diff output (viewport handles wrapping)
func colorizeGitDiff(diff string) string {
	var sb strings.Builder
	lines := strings.Split(diff, "\n")

	for _, line := range lines {
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
