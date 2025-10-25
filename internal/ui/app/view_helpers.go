package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// formatNumber formats an integer with comma separators
func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// truncate shortens a string to maxLen with ellipsis if needed
func truncate(s string, maxLen int) string {
	if lipgloss.Width(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// truncateCommitHash formats a commit hash to a specific length
func truncateCommitHash(hash string, length int) string {
	if len(hash) <= length {
		return hash
	}
	return hash[:length]
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// padRight pads a string with spaces to reach the specified width
func padRight(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// centerText centers text within the specified width
func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}

	leftPad := (width - textWidth) / 2
	rightPad := width - textWidth - leftPad

	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatPercent formats a float as a percentage
func formatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(str string) string {
	// Simple ANSI escape code removal
	result := str
	for {
		start := strings.Index(result, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

// getRelativeTime formats a relative time string
func getRelativeTime(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds ago", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm ago", seconds/60)
	} else if seconds < 86400 {
		return fmt.Sprintf("%dh ago", seconds/3600)
	} else {
		return fmt.Sprintf("%dd ago", seconds/86400)
	}
}

// splitLines splits text into lines, handling different line endings
func splitLines(text string) []string {
	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

// joinLines joins lines with newline characters
func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// indentText indents each line of text by the specified number of spaces
func indentText(text string, spaces int) string {
	indent := strings.Repeat(" ", spaces)
	lines := splitLines(text)
	for i := range lines {
		if lines[i] != "" {
			lines[i] = indent + lines[i]
		}
	}
	return joinLines(lines)
}

// removeEmptyLines removes empty lines from text
func removeEmptyLines(text string) string {
	lines := splitLines(text)
	var nonEmpty []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	return joinLines(nonEmpty)
}

// highlightMatch highlights search matches in text
func (m *Model) highlightMatch(text, pattern string) string {
	if pattern == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerPattern := strings.ToLower(pattern)

	index := strings.Index(lowerText, lowerPattern)
	if index == -1 {
		return text
	}

	before := text[:index]
	match := text[index : index+len(pattern)]
	after := text[index+len(pattern):]

	return before + m.styles.Highlight.Render(match) + after
}

// getIconForStatus returns an icon for the given status
func getIconForStatus(status string) string {
	switch status {
	case "active":
		return "●"
	case "working":
		return "◉"
	case "idle":
		return "○"
	case "completed":
		return "✓"
	case "failed":
		return "✗"
	default:
		return "·"
	}
}

// getColorForStatus returns a color for the given status
func getColorForStatus(status string) string {
	switch status {
	case "active":
		return "green"
	case "working":
		return "yellow"
	case "idle":
		return "gray"
	case "completed":
		return "blue"
	case "failed":
		return "red"
	default:
		return "white"
	}
}

// renderLabelValue renders a label-value pair with consistent formatting
func renderLabelValue(label, value string, valueStyle lipgloss.Style, labelStyle lipgloss.Style) string {
	return fmt.Sprintf("  %s %s\n",
		labelStyle.Render(label),
		valueStyle.Render(value))
}

// countTasksByState counts tasks grouped by their state
func countTasksByState(tasks []domain.Task) map[string]int {
	counts := map[string]int{
		"todo":       0,
		"inProgress": 0,
		"completed":  0,
		"archived":   0,
	}

	for _, task := range tasks {
		if task.Archived {
			counts["archived"]++
		} else {
			switch task.StateID {
			case 3: // done
				counts["completed"]++
			case 2: // in_progress
				counts["inProgress"]++
			default: // todo
				counts["todo"]++
			}
		}
	}

	return counts
}

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

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// formatTimestamp formats a time with standard format
func formatTimestamp(t time.Time) string {
	return t.Format("Jan 2, 2006 15:04:05 MST")
}

// formatTime formats a time with short format (HH:MM:SS)
func formatTime(t time.Time) string {
	return t.Format("15:04:05")
}