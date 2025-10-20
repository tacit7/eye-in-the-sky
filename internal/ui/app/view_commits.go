package app

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// renderCommitsView renders the commits view with split panes
func (m *Model) renderCommitsView() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Build item list for left pane
	items := make([]string, len(m.commits))
	for i, commit := range m.commits {
		date := commit.Timestamp.Format("2006-01-02")
		hash := truncateCommitHash(commit.CommitHash, 8)
		subject := truncate(commit.CommitMessage, 40)
		items[i] = fmt.Sprintf("%s  %s  %s", date, hash, subject)
	}

	// Build detail content for right pane
	var detailContent string
	if m.commitsIndex >= 0 && m.commitsIndex < len(m.commits) {
		detailContent = m.renderCommitDetails(m.commits[m.commitsIndex])
	} else if len(m.commits) == 0 && m.selectedAgent.GitWorktreePath == "" {
		detailContent = m.styles.Subtle.Render("No git repository configured for this agent")
	}

	// Footer
	footer := m.renderFooterWithKeys("[j/k] Move  [r] Refresh  [q/esc] Back")

	return m.renderSplitPaneView(
		"Commits",
		items,
		m.commitsIndex,
		len(m.commits),
		detailContent,
		footer,
	)
}

// renderCommitDetails renders the full commit message and diff
func (m *Model) renderCommitDetails(commit Commit) string {
	var b strings.Builder

	// Commit info
	b.WriteString(m.styles.Primary.Render("Commit: "))
	b.WriteString(m.styles.Text.Render(commit.CommitHash))
	b.WriteString("\n")

	b.WriteString(m.styles.Primary.Render("Date: "))
	b.WriteString(m.styles.Text.Render(commit.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Commit message
	b.WriteString(m.styles.Title.Render("Message"))
	b.WriteString("\n")
	b.WriteString(m.styles.Text.Render(commit.CommitMessage))
	b.WriteString("\n\n")

	// Get diff from git
	if m.selectedAgent != nil && m.selectedAgent.GitWorktreePath != "" {
		diff, err := getCommitDiff(commit.CommitHash, m.selectedAgent.GitWorktreePath)
		if err == nil && diff != "" {
			b.WriteString(m.styles.Title.Render("Diff"))
			b.WriteString("\n")
			// Apply syntax highlighting to the diff
			highlightedDiff := highlightDiff(diff)
			b.WriteString(highlightedDiff)
		}
	}

	return b.String()
}

// getCommitDiff retrieves the diff for a commit
func getCommitDiff(hash string, worktreePath string) (string, error) {
	cmd := exec.Command("git", "show", "--color=never", hash)
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	diff := string(output)

	// Truncate if too large (>5000 chars)
	if len(diff) > 5000 {
		diff = diff[:5000] + "\n... (diff truncated)"
	}

	return diff, nil
}

// highlightDiff applies syntax highlighting to git diff output
func highlightDiff(diff string) string {
	// Get the diff lexer
	lexer := lexers.Get("diff")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use a terminal-friendly style
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Create a terminal256 formatter with ANSI colors
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize the diff
	iterator, err := lexer.Tokenise(nil, diff)
	if err != nil {
		return diff // Return unhighlighted on error
	}

	// Format with colors
	var builder strings.Builder
	err = formatter.Format(&builder, style, iterator)
	if err != nil {
		return diff // Return unhighlighted on error
	}

	return builder.String()
}
