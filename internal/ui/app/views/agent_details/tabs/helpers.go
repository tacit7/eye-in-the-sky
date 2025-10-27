package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/utils"
)

// filterNonEmpty removes empty strings from a slice
func filterNonEmpty(sections []string) []string {
	var result []string
	for _, s := range sections {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

// renderLabelValue renders a label-value pair with consistent formatting
func renderLabelValue(label, value string, valueStyle lipgloss.Style, labelStyle lipgloss.Style) string {
	return fmt.Sprintf("  %s %s\n",
		labelStyle.Render(label),
		valueStyle.Render(value))
}

// truncateID is a wrapper for utils.TruncateID for backward compatibility
func truncateID(id string, maxLen int) string {
	return utils.TruncateID(id, maxLen)
}

// formatTimestamp formats a time with standard format
func formatTimestamp(t time.Time) string {
	return t.Format("Jan 2, 2006 15:04:05 MST")
}

// formatTime formats a time with short format (HH:MM:SS)
func formatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// activityStyle returns the appropriate style based on elapsed time
// Green: < 5 min (active), Yellow: 5-30 min (stale), Gray: > 30 min (inactive)
func activityStyle(elapsed time.Duration, styles OverviewStyles) lipgloss.Style {
	switch {
	case elapsed < 5*time.Minute:
		return styles.Success
	case elapsed < 30*time.Minute:
		return styles.Warning
	default:
		return styles.Subtle
	}
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetActionIcon returns an icon/emoji for a given action type
func GetActionIcon(actionType string) string {
	switch actionType {
	case "task_start":
		return "▶️"
	case "file_operation":
		return "📁"
	case "git_commit":
		return "📝"
	case "status_update":
		return "📊"
	case "log":
		return "📋"
	case "note":
		return "📌"
	default:
		return "⚡"
	}
}
