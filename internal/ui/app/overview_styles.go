package app

import (
	"github.com/charmbracelet/lipgloss"
)

// OverviewStyles contains all styles specific to the overview tab
// Centralizes style definitions to reduce inline style calls
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
	Success2     lipgloss.Style
	Danger       lipgloss.Style
	Info         lipgloss.Style
	Error        lipgloss.Style
}

// NewOverviewStyles creates and returns an OverviewStyles with theme-aware styling
// This allows centralized theming without changing render functions
func NewOverviewStyles(styles Styles) OverviewStyles {
	return OverviewStyles{
		SectionTitle: styles.SectionTitle,
		Label:        styles.Label,
		Value:        styles.Value,
		Subtle:       styles.Subtle,
		Success:      styles.Success,
		Warning:      styles.Warning,
		Primary:      styles.Primary,
		Secondary:    styles.Secondary,
		Git:          styles.Git,
		Success2:     styles.Success,
		Danger:       styles.Error,
		Info:         styles.Secondary,
		Error:        styles.Error,
	}
}
