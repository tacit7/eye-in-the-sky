package app

import (
	"strings"
	"github.com/charmbracelet/lipgloss"
)

// renderUsageSections composes multiple usage sections with proper spacing and borders
func (m *Model) renderUsageSections(sections ...string) string {
	// Create rounded border style for sections
	sectionBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
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

// renderSectionWithTitle renders a section with a styled title
func (m *Model) renderSectionWithTitle(title string, content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.styles.Title.Render(title))
	b.WriteString("\n\n")
	b.WriteString(content)

	return b.String()
}

// renderUsageFooter renders the footer help hints for the usage tab
func (m *Model) renderUsageFooter() string {
	hints := "↑/↓ Scroll  |  R Refresh  |  Q Quit"
	separator := m.styles.Border.Render(strings.Repeat("─", 70))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render(hints))

	return b.String()
}