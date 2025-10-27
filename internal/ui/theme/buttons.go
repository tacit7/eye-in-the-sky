package theme

import "github.com/charmbracelet/lipgloss"

// Button Styles - Consistent button appearance across the application

var (
	// BtnPrimary is the default button style for primary actions (submit, confirm)
	// Usage: theme.BtnPrimary.Render("Submit")
	BtnPrimary = lipgloss.NewStyle().
			Background(PrimaryColor).
			Foreground(SecondaryColor).
			Padding(0, 2).
			Bold(true)

	// BtnSecondary for secondary/alternative actions
	// Usage: theme.BtnSecondary.Render("Cancel")
	BtnSecondary = lipgloss.NewStyle().
			Background(lipgloss.Color("#6272A4")).
			Foreground(SecondaryColor).
			Padding(0, 2)

	// BtnDisabled for inactive/disabled buttons
	// Usage: theme.BtnDisabled.Render("Submit")
	BtnDisabled = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")).
			Foreground(MutedColor).
			Padding(0, 2)

	// BtnActive for currently pressed/active button state
	// Usage: theme.BtnActive.Render("Confirm")
	BtnActive = BtnPrimary.Copy().
			Background(lipgloss.Color("#50FA7B")).
			Bold(true)

	// BtnFocused for focused button (keyboard navigation)
	// Usage: theme.BtnFocused.Render("Submit")
	BtnFocused = BtnPrimary.Copy().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor)

	// BtnDanger for destructive actions (delete, remove)
	// Usage: theme.BtnDanger.Render("Delete")
	BtnDanger = lipgloss.NewStyle().
			Background(ErrorColor).
			Foreground(SecondaryColor).
			Padding(0, 2).
			Bold(true)
)
