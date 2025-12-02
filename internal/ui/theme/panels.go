package theme

import "github.com/charmbracelet/lipgloss"

// Panel Styles - Containers and box layouts

var (
	// PanelNormal for standard container panels
	// Usage: theme.PanelNormal.Render(content)
	PanelNormal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	// PanelHighlight for emphasized/important panels
	// Usage: theme.PanelHighlight.Render(content)
	PanelHighlight = PanelNormal.Copy().
			BorderForeground(PrimaryColor)

	// PanelError for error display panels
	// Usage: theme.PanelError.Render(errorMessage)
	PanelError = PanelNormal.Copy().
			BorderForeground(ErrorColor).
			Background(lipgloss.Color("#3A2222"))

	// PanelSuccess for success display panels
	// Usage: theme.PanelSuccess.Render(successMessage)
	PanelSuccess = PanelNormal.Copy().
			BorderForeground(SuccessColor).
			Background(lipgloss.Color("#1E3A1E"))

	// PanelNoBorder for borderless containers
	// Usage: theme.PanelNoBorder.Render(content)
	PanelNoBorder = lipgloss.NewStyle().
			Padding(1, 2)

	// PanelModal for modal dialog boxes
	// Usage: theme.PanelModal.Render(modalContent)
	PanelModal = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(PrimaryColor).
			Background(BackgroundDark).
			Padding(1, 2).
			Width(60)

	// PanelSidebar for sidebar containers
	// Usage: theme.PanelSidebar.Render(sidebarContent)
	PanelSidebar = lipgloss.NewStyle().
			Border(lipgloss.Border{Right: "│"}).
			BorderForeground(BorderColor).
			Padding(1, 1)
)
