package theme

import "github.com/charmbracelet/lipgloss"

// Text Texture Styles - Typography and text formatting

var (
	// TextTitle for main titles and headings
	// Usage: theme.TextTitle.Render("Eye in the Sky")
	TextTitle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Underline(true)

	// TextSubtitle for section subtitles
	// Usage: theme.TextSubtitle.Render("Active Agents")
	TextSubtitle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// TextLabel for form labels and field names
	// Usage: theme.TextLabel.Render("Agent ID:")
	TextLabel = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(15).
			Align(lipgloss.Right)

	// TextValue for data values paired with labels
	// Usage: theme.TextValue.Render("a1b2c3d4")
	TextValue = lipgloss.NewStyle().
			Foreground(SecondaryColor)

	// TextHighlight for emphasized inline text
	// Usage: theme.TextHighlight.Render("important")
	TextHighlight = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// TextMuted for de-emphasized text
	// Usage: theme.TextMuted.Render("optional info")
	TextMuted = lipgloss.NewStyle().
			Foreground(MutedColor)

	// TextLink for clickable/selectable items
	// Usage: theme.TextLink.Render("View details")
	TextLink = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Underline(true)

	// TextCode for inline code snippets
	// Usage: theme.TextCode.Render("git commit -m")
	TextCode = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Background(lipgloss.Color("#282A36")).
			Padding(0, 1)

	// TextCodeBlock for multi-line code blocks
	// Usage: theme.TextCodeBlock.Render(codeSnippet)
	TextCodeBlock = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#282A36")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6272A4"))

	// TextTimestamp for timestamps and dates
	// Usage: theme.TextTimestamp.Render("2025-10-26 15:30:00")
	TextTimestamp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Italic(true)

	// TextStatus for status indicators
	// Usage: theme.TextStatus.Render("● active")
	TextStatus = lipgloss.NewStyle().
			Bold(true)

	// TextWarning for warning messages
	// Usage: theme.TextWarning.Render("⚠ Warning: ")
	TextWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	// TextPlaceholder for placeholder text in inputs
	// Usage: theme.TextPlaceholder.Render("Enter note...")
	TextPlaceholder = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)
)

// Status-specific text styles with color coding

var (
	// StatusActive for "active" status
	StatusActive = TextStatus.Copy().Foreground(SuccessColor)

	// StatusIdle for "idle" status
	StatusIdle = TextStatus.Copy().Foreground(lipgloss.Color("#F1FA8C"))

	// StatusWorking for "working" status
	StatusWorking = TextStatus.Copy().Foreground(PrimaryColor)

	// StatusCompleted for "completed" status
	StatusCompleted = TextStatus.Copy().Foreground(lipgloss.Color("#8BE9FD"))

	// StatusFailed for "failed" status
	StatusFailed = TextStatus.Copy().Foreground(ErrorColor)

	// StatusArchived for "archived" status
	StatusArchived = TextStatus.Copy().Foreground(MutedColor)
)
