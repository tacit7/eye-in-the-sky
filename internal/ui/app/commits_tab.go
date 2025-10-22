package app

import (
	"fmt"
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderCommitsTabView renders the commits tab with split-pane view
func (m *Model) renderCommitsTabView() string {
	if len(m.commits) == 0 {
		return m.styles.Subtle.Render("No commits found")
	}

	// Build item list for left pane
	items := make([]string, len(m.commits))
	for i, commit := range m.commits {
		date := commit.Timestamp.Format("2006-01-02")
		hash := truncateCommitHash(string(commit.Hash), 8)
		subject := truncate(commit.Message, 40)
		items[i] = fmt.Sprintf("%s  %s  %s", date, hash, subject)
	}

	// Build detail content for right pane
	var detailContent string
	if m.commitsIndex >= 0 && m.commitsIndex < len(m.commits) {
		detailContent = renderCommitDetails(m.commits[m.commitsIndex], m.styles)
	}

	return SplitPane{
		LeftItems:     items,
		SelectedIndex: m.commitsIndex,
		RightContent:  detailContent,
		Styles:        m.styles,
		Width:         m.width,
	}.View()
}

// renderCommitDetails renders details for a single commit
func renderCommitDetails(commit domain.Commit, styles Styles) string {
	var b strings.Builder

	// Commit hash
	b.WriteString(styles.Git.Render("commit "))
	b.WriteString(styles.Value.Render(string(commit.Hash)))
	b.WriteString("\n")

	// Author info (if available)
	if commit.Author != "" {
		b.WriteString(styles.Label.Render("Author: "))
		b.WriteString(styles.Value.Render(commit.Author))
		b.WriteString("\n")
	}

	// Date
	b.WriteString(styles.Label.Render("Date:   "))
	b.WriteString(styles.Value.Render(commit.Timestamp.Format("Mon Jan 2 15:04:05 2006 MST")))
	b.WriteString("\n\n")

	// Message
	b.WriteString(styles.Text.Render(commit.Message))
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

	// Diff preview (if available)
	if commit.Diff != "" {
		b.WriteString("\n")
		b.WriteString(styles.SectionTitle.Render("Diff Preview"))
		b.WriteString("\n\n")
		// Show first 20 lines of diff
		lines := strings.Split(commit.Diff, "\n")
		maxLines := 20
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		for i := 0; i < maxLines; i++ {
			line := lines[i]
			if strings.HasPrefix(line, "+") {
				b.WriteString(styles.Success.Render(line))
			} else if strings.HasPrefix(line, "-") {
				b.WriteString(styles.Error.Render(line))
			} else {
				b.WriteString(styles.Subtle.Render(line))
			}
			b.WriteString("\n")
		}
		if len(lines) > maxLines {
			b.WriteString(styles.Subtle.Render(fmt.Sprintf("... +%d more lines", len(lines)-maxLines)))
		}
	}

	return b.String()
}
