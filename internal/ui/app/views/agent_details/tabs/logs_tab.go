package tabs

import (
	"fmt"
	"strings"
)

// RenderLogs renders the logs tab content
func RenderLogs(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.Logs) == 0 {
		return overviewStyles.Subtle.Render("\n  No logs available for this session\n")
	}

	var sb strings.Builder

	// Header
	sb.WriteString(overviewStyles.Primary.Render("Time        Type        Message"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// Render logs
	for _, log := range ctx.Logs {
		timestamp := log.Timestamp.Format("15:04:05")
		logType := log.Type
		if len(logType) > 12 {
			logType = logType[:9] + "..."
		}
		message := log.Message
		if len(message) > 50 {
			message = message[:47] + "..."
		}

		line := fmt.Sprintf("%-11s %-11s %s\n", timestamp, logType, message)
		sb.WriteString(line)
	}

	return sb.String()
}

// TODO: Split-pane view with log details requires state migration
/*
// renderLogDetails renders details for a single log entry
func renderLogDetails(log domain.Log, styles OverviewStyles) string {
	var b strings.Builder

	// Log header
	b.WriteString(styles.Primary.Render("Type: "))
	b.WriteString(log.Type)
	b.WriteString("\n")

	b.WriteString(styles.Primary.Render("Time: "))
	b.WriteString(log.Timestamp.Format("2006-01-02 15:04:05"))
	b.WriteString("\n\n")

	// Content
	b.WriteString(log.Message)

	return b.String()
}
*/
