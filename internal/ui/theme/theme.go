package theme

import "github.com/charmbracelet/lipgloss"

// Base Colors - Core color palette for the application
var (
	PrimaryColor   = lipgloss.Color("#00ADD8") // Cyan - primary actions and highlights
	SecondaryColor = lipgloss.Color("#FFFFFF") // White - text and backgrounds
	BorderColor    = lipgloss.Color("#444444") // Dark gray - borders and separators
	ErrorColor     = lipgloss.Color("#FF5555") // Red - errors and warnings
	SuccessColor   = lipgloss.Color("#50FA7B") // Green - success messages
	MutedColor     = lipgloss.Color("#888888") // Light gray - dimmed/inactive text
	BackgroundDark = lipgloss.Color("#1E1E1E") // Very dark gray - dark backgrounds
)

// Adaptive Colors - Colors that change based on terminal theme
var (
	AdaptivePrimary = lipgloss.AdaptiveColor{
		Light: "#00ADD8",
		Dark:  "#5DADE2",
	}
	AdaptiveText = lipgloss.AdaptiveColor{
		Light: "#333333",
		Dark:  "#FFFFFF",
	}
)

// Base Styles - Foundational styles used across the application
var (
	// Header style for section headers and titles
	Header = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	// Panel style for containers and boxes
	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(1, 2)

	// PanelFocused style for focused/active panels
	PanelFocused = Panel.Copy().
		BorderForeground(PrimaryColor).
		BorderStyle(lipgloss.ThickBorder())

	// List style for list containers
	List = lipgloss.NewStyle().
		Padding(0, 1)

	// ListItem style for individual list items
	ListItem = lipgloss.NewStyle().
		Padding(0, 2)

	// ListItemSelected style for selected list items
	ListItemSelected = ListItem.Copy().
		Foreground(PrimaryColor).
		Bold(true)
)

// Text Styles - Common text formatting
var (
	// TextNormal is the default text style
	TextNormal = lipgloss.NewStyle()

	// TextBold for emphasized text
	TextBold = lipgloss.NewStyle().Bold(true)

	// TextDimmed for de-emphasized/secondary text
	TextDimmed = lipgloss.NewStyle().
		Foreground(MutedColor)

	// TextError for error messages
	TextError = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	// TextSuccess for success messages
	TextSuccess = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	// CodeBlock for inline code or code blocks
	CodeBlock = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Background(lipgloss.Color("#282A36")).
		Padding(0, 1)
)

// Status Bar Styles
var (
	// StatusBar is the base status bar style
	StatusBar = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Background(BackgroundDark).
		Padding(0, 1)

	// StatusBarKey for keyboard shortcuts in status bar
	StatusBarKey = StatusBar.Copy().
		Foreground(PrimaryColor).
		Bold(true)
)

// OverviewStyles provides backward compatibility with existing tab code
// This struct matches the style interface used in agent_details tabs
type OverviewStyles struct {
	SectionTitle lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Subtle       lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Primary      lipgloss.Style
	Secondary    lipgloss.Style
	Git          lipgloss.Style
	Error        lipgloss.Style
}

// GetOverviewStyles returns styles compatible with the OverviewStyles struct
// This bridges the gap during migration from old style system to centralized theme
func GetOverviewStyles() OverviewStyles {
	return OverviewStyles{
		SectionTitle: Header,
		Label:        TextLabel,
		Value:        TextValue,
		Subtle:       TextDimmed,
		Success:      TextSuccess,
		Warning:      TextWarning, // From textures.go
		Primary:      lipgloss.NewStyle().Foreground(PrimaryColor),
		Secondary:    lipgloss.NewStyle().Foreground(SecondaryColor),
		Git:          lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")), // Purple for git
		Error:        TextError,
	}
}

// Helper Functions

// MaxWidth creates a style with maximum width constraint
func MaxWidth(width int) lipgloss.Style {
	return lipgloss.NewStyle().MaxWidth(width)
}

// Padding creates a style with specified padding
func Padding(vertical, horizontal int) lipgloss.Style {
	return lipgloss.NewStyle().Padding(vertical, horizontal)
}

// Margin creates a style with specified margin
func Margin(vertical, horizontal int) lipgloss.Style {
	return lipgloss.NewStyle().Margin(vertical, horizontal)
}
