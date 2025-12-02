package app

import (
	"github.com/charmbracelet/lipgloss"
)

// GetStatusStyle returns the appropriate style for a given status
func GetStatusStyle(status string, styles Styles) lipgloss.Style {
	switch status {
	case "active":
		return styles.Success
	case "working":
		return styles.Warning
	case "idle":
		return styles.Subtle
	case "failed":
		return styles.Error
	case "completed":
		return styles.Primary
	default:
		return styles.Subtle
	}
}
