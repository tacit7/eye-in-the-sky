package app

import (
	"strings"
)

// renderUsageSections composes multiple usage sections with proper spacing and borders
func (m *Model) renderUsageSections(sections ...string) string {
	var validSections []string

	for _, section := range sections {
		trimmed := strings.TrimSpace(section)
		if trimmed != "" {
			validSections = append(validSections, trimmed)
		}
	}

	if len(validSections) == 0 {
		return ""
	}

	// Join sections with double newlines for spacing
	result := strings.Join(validSections, "\n\n")

	// Add final border at the end
	if len(validSections) > 0 {
		result += "\n\n" + m.styles.Border.Render(strings.Repeat("─", 70))
	}

	return result
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