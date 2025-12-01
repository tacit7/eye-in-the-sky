package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// ThemeMode represents light or dark theme preference
type ThemeMode int

const (
	ThemeModeAuto ThemeMode = iota // Use terminal's preference
	ThemeModeLight
	ThemeModeDark
)

// Controller manages theme switching and provides dynamic colors
type Controller struct {
	mode ThemeMode
}

// NewController creates a new theme controller
func NewController() *Controller {
	return &Controller{
		mode: ThemeModeAuto, // Default to auto-detect from terminal
	}
}

// SetMode changes the theme mode
func (c *Controller) SetMode(mode ThemeMode) {
	c.mode = mode
}

// GetMode returns the current theme mode
func (c *Controller) GetMode() ThemeMode {
	return c.mode
}

// ToggleMode cycles through Auto → Light → Dark → Auto
func (c *Controller) ToggleMode() ThemeMode {
	switch c.mode {
	case ThemeModeAuto:
		c.mode = ThemeModeLight
	case ThemeModeLight:
		c.mode = ThemeModeDark
	case ThemeModeDark:
		c.mode = ThemeModeAuto
	}
	return c.mode
}

// GetModeString returns a human-readable string for the current mode
func (c *Controller) GetModeString() string {
	switch c.mode {
	case ThemeModeAuto:
		return "Auto"
	case ThemeModeLight:
		return "Light"
	case ThemeModeDark:
		return "Dark"
	default:
		return "Auto"
	}
}

// GetAdaptiveColor returns a lipgloss AdaptiveColor based on theme mode
// For Auto mode, lipgloss will detect terminal's background automatically
func (c *Controller) GetAdaptiveColor(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{
		Light: light,
		Dark:  dark,
	}
}

// GetPrimaryColor returns the primary color based on current theme
func (c *Controller) GetPrimaryColor() lipgloss.Color {
	switch c.mode {
	case ThemeModeLight:
		return lipgloss.Color("#00ADD8") // Cyan for light mode
	case ThemeModeDark:
		return lipgloss.Color("#5DADE2") // Lighter cyan for dark mode
	default:
		// Auto mode - use adaptive color, return dark variant as default
		return lipgloss.Color("#5DADE2")
	}
}

// GetBackgroundColor returns the background color based on current theme
func (c *Controller) GetBackgroundColor() lipgloss.Color {
	switch c.mode {
	case ThemeModeLight:
		return lipgloss.Color("#FFFFFF") // White for light mode
	case ThemeModeDark:
		return lipgloss.Color("#1E1E1E") // Dark gray for dark mode
	default:
		return lipgloss.Color("#1E1E1E") // Default to dark
	}
}

// GetTextColor returns the text color based on current theme
func (c *Controller) GetTextColor() lipgloss.Color {
	switch c.mode {
	case ThemeModeLight:
		return lipgloss.Color("#333333") // Dark gray for light mode
	case ThemeModeDark:
		return lipgloss.Color("#FFFFFF") // White for dark mode
	default:
		return lipgloss.Color("#FFFFFF") // Default to light text
	}
}

// GetBorderColor returns the border color based on current theme
func (c *Controller) GetBorderColor() lipgloss.Color {
	switch c.mode {
	case ThemeModeLight:
		return lipgloss.Color("#CCCCCC") // Light gray for light mode
	case ThemeModeDark:
		return lipgloss.Color("#444444") // Dark gray for dark mode
	default:
		return lipgloss.Color("#444444")
	}
}

// GetMutedColor returns the muted/dimmed text color based on current theme
func (c *Controller) GetMutedColor() lipgloss.Color {
	switch c.mode {
	case ThemeModeLight:
		return lipgloss.Color("#666666") // Medium gray for light mode
	case ThemeModeDark:
		return lipgloss.Color("#888888") // Light gray for dark mode
	default:
		return lipgloss.Color("#888888")
	}
}
