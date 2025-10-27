package tabs

import (
	"fmt"
	"strings"
)

// RenderCommits renders the commits tab content
// TODO: State migration in progress - needs commitsIndex and split-pane state
func RenderCommits(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.Commits) == 0 {
		return overviewStyles.Subtle.Render("\n  No commits found for this agent\n")
	}

	var sb strings.Builder

	// Simple header
	sb.WriteString(overviewStyles.Primary.Render("Date        Hash     Message"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// Render commits in simple list format
	for _, commit := range ctx.Commits {
		date := commit.Timestamp.Format("2006-01-02")
		hash := string(commit.Hash)
		if len(hash) > 8 {
			hash = hash[:8]
		}

		message := commit.Message
		if len(message) > 45 {
			message = message[:42] + "..."
		}

		line := fmt.Sprintf("%-11s %-8s %s\n", date, hash, message)
		sb.WriteString(line)
	}

	return sb.String()
}

// TODO: Split-pane view with commit details requires state migration
/*
// renderCommitDetails renders details for a single commit
func renderCommitDetails(commit domain.Commit, styles OverviewStyles) string {
	var b strings.Builder

	// Commit hash
	b.WriteString(styles.Git.Render("commit "))
	b.WriteString(string(commit.Hash))
	b.WriteString("\n")

	// Author info (if available)
	if commit.Author != "" {
		b.WriteString(styles.Label.Render("Author: "))
		b.WriteString(commit.Author)
		b.WriteString("\n")
	}

	// Date
	b.WriteString(styles.Label.Render("Date:   "))
	b.WriteString(commit.Timestamp.Format("Mon Jan 2 15:04:05 2006 MST"))
	b.WriteString("\n\n")

	// Message
	b.WriteString(commit.Message)
	b.WriteString("\n")

	// Files changed (if available)
	if len(commit.FilesChanged) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.SectionTitle.Render("Files Changed"))
		b.WriteString("\n\n")
		for _, file := range commit.FilesChanged {
			b.WriteString(fmt.Sprintf("  %s\n", file))
		}
	}

	return b.String()
}
*/
