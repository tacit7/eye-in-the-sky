package usageview

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StylesProvider provides access to styling
type StylesProvider interface {
	Title() lipgloss.Style
	Border() lipgloss.Style
	Subtle() lipgloss.Style
	// Color method for theme configuration
	SectionBorderColor() lipgloss.Color
}

// RenderUsageSections composes multiple usage sections with proper spacing and borders
func RenderUsageSections(styles StylesProvider, sections ...string) string {
	// Create rounded border style for sections
	sectionBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SectionBorderColor()).
		Padding(1, 2).
		MarginTop(1)

	var boxedSections []string

	for _, section := range sections {
		trimmed := strings.TrimSpace(section)
		if trimmed != "" {
			// Wrap each section in a rounded border box
			boxedSections = append(boxedSections, sectionBox.Render(trimmed))
		}
	}

	if len(boxedSections) == 0 {
		return ""
	}

	// Join boxed sections with newlines
	return strings.Join(boxedSections, "\n")
}

// RenderSectionWithTitle renders a section with a styled title
func RenderSectionWithTitle(title string, content string, styles StylesProvider) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(styles.Title().Render(title))
	b.WriteString("\n\n")
	b.WriteString(content)

	return b.String()
}

// RenderUsageFooter renders the footer help hints and status for the usage tab
func RenderUsageFooter(styles StylesProvider, syncStatus string, width int) string {
	hints := "↑/↓ Scroll  |  R Refresh  |  Q Quit"

	// Use adaptive width for separator
	if width == 0 {
		width = 70
	}
	separator := styles.Border().Render(strings.Repeat("─", width))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n")

	// Show hints on left, status on right if present
	if syncStatus != "" {
		statusBar := lipgloss.JoinHorizontal(lipgloss.Top,
			styles.Subtle().Render(hints),
			styles.Subtle().Render("  |  "),
			styles.Subtle().Render(syncStatus),
		)
		b.WriteString(statusBar)
	} else {
		b.WriteString(styles.Subtle().Render(hints))
	}

	return b.String()
}
