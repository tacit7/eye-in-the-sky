package app

import (
	"strings"
)

// renderUsageEmptyState renders the empty state when no usage data exists
func (m *Model) renderUsageEmptyState() string {
	return m.styles.Subtle.Render("  No usage data available")
}

// renderClaudeCodeEmptyState renders empty state for Claude Code database
func (m *Model) renderClaudeCodeEmptyState() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Claude Code Usage"))
	b.WriteString("\n\n")

	if m.ccusageSyncing {
		b.WriteString(m.styles.Working.Render("  \uf021 " + m.ccusageSyncStatus))
	} else {
		b.WriteString(m.styles.Subtle.Render("  No Claude Code usage data found"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.Primary.Render("  Press 'I' to initialize database"))
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("  This will discover and parse JSONL files from:"))
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("  ~/.config/claude/projects/"))
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("  ~/.claude/projects/"))
	}

	b.WriteString("\n\n")
	b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
	b.WriteString("\n\n")

	return b.String()
}

// renderClaudeCodeSyncPrompt renders prompt to sync Claude Code data
func (m *Model) renderClaudeCodeSyncPrompt() string {
	var b strings.Builder

	b.WriteString(m.styles.Subtle.Render("No data available\n\n"))
	b.WriteString(m.styles.Text.Render("Press 'I' to sync and load Claude Code usage data from your local projects.\n"))
	b.WriteString(m.styles.Subtle.Render("This will scan ~/.config/claude/projects/ and ~/.claude/projects/ for usage logs.\n"))

	return b.String()
}