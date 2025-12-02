package usageview

import "github.com/charmbracelet/lipgloss"

// AppStyles wraps the app.Styles type to provide the interface methods
type AppStyles struct {
	PrimaryStyle lipgloss.Style
	TextStyle    lipgloss.Style
	SubtleStyle  lipgloss.Style
	TitleStyle   lipgloss.Style
	BorderStyle  lipgloss.Style
}

// Primary returns the primary style
func (s *AppStyles) Primary() lipgloss.Style {
	return s.PrimaryStyle
}

// Text returns the text style
func (s *AppStyles) Text() lipgloss.Style {
	return s.TextStyle
}

// Subtle returns the subtle style
func (s *AppStyles) Subtle() lipgloss.Style {
	return s.SubtleStyle
}

// Title returns the title style
func (s *AppStyles) Title() lipgloss.Style {
	return s.TitleStyle
}

// Border returns the border style
func (s *AppStyles) Border() lipgloss.Style {
	return s.BorderStyle
}
