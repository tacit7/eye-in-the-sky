package app

import (
	"fmt"
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderLogsTabView renders the logs tab with split-pane view
func (m *Model) renderLogsTabView() string {
	if len(m.logs) == 0 {
		return m.styles.Subtle.Render("No logs found")
	}

	// Build item list for left pane
	items := make([]string, len(m.logs))
	for i, log := range m.logs {
		timestamp := log.Timestamp.Format("15:04:05")
		logType := truncate(log.Type, 10)
		items[i] = fmt.Sprintf("%s  %-10s", timestamp, logType)
	}

	// Build detail content for right pane
	var detailContent string
	if m.logsIndex >= 0 && m.logsIndex < len(m.logs) {
		detailContent = renderLogDetails(m.logs[m.logsIndex], m.styles)
	}

	return SplitPane{
		LeftItems:     items,
		SelectedIndex: m.logsIndex,
		RightContent:  detailContent,
		Styles:        m.styles,
		Width:         m.width,
	}.View()
}

// renderLogDetails renders details for a single log entry
func renderLogDetails(log domain.Log, styles Styles) string {
	var b strings.Builder

	// Log header
	b.WriteString(styles.Primary.Render("Type: "))
	b.WriteString(styles.Value.Render(log.Type))
	b.WriteString("\n")

	b.WriteString(styles.Primary.Render("Time: "))
	b.WriteString(styles.Value.Render(log.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Content (with markdown rendering if needed)
	if strings.Contains(log.Message, "```") || strings.Contains(log.Message, "#") {
		// Try to render as markdown
		if rendered, err := RenderMarkdown(log.Message, styles); err == nil {
			b.WriteString(rendered)
		} else {
			b.WriteString(styles.Text.Render(log.Message))
		}
	} else {
		b.WriteString(styles.Text.Render(log.Message))
	}

	return b.String()
}
