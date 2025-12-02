package app

import (
	"strings"
)

// RenderMarkdown renders markdown content with styling
func RenderMarkdown(content string, styles Styles) (string, error) {
	// Simple markdown rendering
	lines := strings.Split(content, "\n")
	var rendered strings.Builder

	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			rendered.WriteString(styles.Code.Render(line))
		} else if strings.HasPrefix(line, "# ") {
			rendered.WriteString(styles.Title.Render(strings.TrimPrefix(line, "# ")))
		} else if strings.HasPrefix(line, "## ") {
			rendered.WriteString(styles.SectionTitle.Render(strings.TrimPrefix(line, "## ")))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			rendered.WriteString("  • " + strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		} else {
			rendered.WriteString(line)
		}
		rendered.WriteString("\n")
	}

	return rendered.String(), nil
}
