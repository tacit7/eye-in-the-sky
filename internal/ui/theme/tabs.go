package theme

import "github.com/charmbracelet/lipgloss"

// Tab Styles - Consistent tab appearance for navigation

var (
	// TabActive for the currently selected tab
	// Usage: theme.TabActive.Render("Overview")
	TabActive = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Background(lipgloss.Color("#282A36")).
			Padding(0, 2).
			Bold(true).
			Border(lipgloss.Border{
			Top:         "─",
			Bottom:      " ",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┘",
			BottomRight: "└",
		}).
		BorderForeground(PrimaryColor)

	// TabInactive for unselected tabs
	// Usage: theme.TabInactive.Render("Logs")
	TabInactive = lipgloss.NewStyle().
			Foreground(MutedColor).
			Background(lipgloss.Color("#1E1E1E")).
			Padding(0, 2).
			Border(lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "╰",
			BottomRight: "╯",
		}).
		BorderForeground(BorderColor)

	// TabFocused for keyboard-focused tab (not necessarily active)
	// Usage: theme.TabFocused.Render("Tasks")
	TabFocused = TabInactive.Copy().
			BorderForeground(PrimaryColor).
			Foreground(SecondaryColor)

	// TabHeader for the tab bar container
	// Usage: lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
	TabHeader = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1E1E")).
			Padding(0, 1)
)
