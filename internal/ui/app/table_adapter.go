package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/view/usage"
)

// UsageTableBuilder adapts app.TableBuilder to usageview.TableBuilder interface
type UsageTableBuilder struct {
	*TableBuilder
}

// SetBorderStyle adapts the method to return the interface type
func (t *UsageTableBuilder) SetBorderStyle(style int) usageview.TableBuilder {
	t.TableBuilder.SetBorderStyle(BorderStyle(style))
	return t
}

// SetHeaderStyle adapts the method to return the interface type
func (t *UsageTableBuilder) SetHeaderStyle(style lipgloss.Style) usageview.TableBuilder {
	t.TableBuilder.SetHeaderStyle(style)
	return t
}

// SetAlternatingRowStyle adapts the method to return the interface type
func (t *UsageTableBuilder) SetAlternatingRowStyle(even, odd lipgloss.Color) usageview.TableBuilder {
	t.TableBuilder.SetAlternatingRowStyle(even, odd)
	return t
}

// AddColumn adapts the method to return the interface type
func (t *UsageTableBuilder) AddColumn(name string, width int, align int, truncate bool) usageview.TableBuilder {
	t.TableBuilder.AddColumn(name, width, Alignment(align), truncate)
	return t
}

// NewUsageTableBuilder creates a table builder for usage views
func NewUsageTableBuilder() usageview.TableBuilder {
	return &UsageTableBuilder{TableBuilder: NewTableBuilder()}
}

// UsageStyles adapts app.Styles to usageview.Styles interface
type UsageStyles struct {
	styles *Styles
}

// NewUsageStyles creates a usage styles adapter
func NewUsageStyles(s *Styles) *UsageStyles {
	return &UsageStyles{styles: s}
}

// Primary returns the primary style
func (s *UsageStyles) Primary() lipgloss.Style {
	return s.styles.Primary
}

// Text returns the text style
func (s *UsageStyles) Text() lipgloss.Style {
	return s.styles.Text
}

// Subtle returns the subtle style
func (s *UsageStyles) Subtle() lipgloss.Style {
	return s.styles.Subtle
}

// Title returns the title style
func (s *UsageStyles) Title() lipgloss.Style {
	return s.styles.Title
}

// Border returns the border style
func (s *UsageStyles) Border() lipgloss.Style {
	return s.styles.Border
}

// Color methods for theme configuration
func (s *UsageStyles) HeaderFg() lipgloss.Color {
	return lipgloss.Color("15") // White
}

func (s *UsageStyles) HeaderBg() lipgloss.Color {
	return lipgloss.Color("33") // Blue
}

func (s *UsageStyles) AlternatingRowDark() lipgloss.Color {
	return lipgloss.Color("233") // Dark gray
}

func (s *UsageStyles) AlternatingRowLight() lipgloss.Color {
	return lipgloss.Color("") // Default/transparent
}

func (s *UsageStyles) SectionBorderColor() lipgloss.Color {
	return lipgloss.Color("240") // Gray
}
