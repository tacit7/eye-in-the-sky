package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderLogs renders the logs tab with scrollable viewport
// Follows LOG_VIEWER_ACTION_PLAN.md for tail -f style behavior
func RenderLogs(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.Logs) == 0 {
		return overviewStyles.Subtle.Render("\n  No logs available for this session\n")
	}

	var sb strings.Builder

	// Background style for table rows
	rowBg := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF0000")).
		Padding(0, 2)

	// Header with column labels (with padding)
	headerText := fmt.Sprintf("  %-12s  %-12s  %s", "Time", "Type", "Message")
	sb.WriteString(overviewStyles.Primary.Render(headerText))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", ctx.Width-4)))
	sb.WriteString("\n\n")

	// Render all logs (viewport will handle scrolling)
	for _, log := range ctx.Logs {
		timestamp := log.Timestamp.Format("15:04:05")
		logType := truncateString(log.Type, 10)
		message := log.Message

		// Color code by log type
		var typeStyled string
		switch strings.ToLower(log.Type) {
		case "error":
			typeStyled = overviewStyles.Error.Render(padRight(logType, 12))
		case "warning", "warn":
			typeStyled = overviewStyles.Warning.Render(padRight(logType, 12))
		case "info":
			typeStyled = overviewStyles.Primary.Render(padRight(logType, 12))
		case "commit":
			typeStyled = overviewStyles.Git.Render(padRight(logType, 12))
		default:
			typeStyled = padRight(logType, 12)
		}

		// Format line with increased padding
		line := fmt.Sprintf("%-12s  %s  %s",
			overviewStyles.Subtle.Render(padRight(timestamp, 12)),
			typeStyled,
			message,
		)

		// Apply background and render
		sb.WriteString(rowBg.Render(line))
		sb.WriteString("\n")
	}

	return sb.String()
}

// truncateString truncates string to max length with ellipsis
func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// padRight pads string to specified length with spaces
func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
