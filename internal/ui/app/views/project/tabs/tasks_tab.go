package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// RenderTasks renders the tasks tab with split-pane: table (left) + details (right)
func RenderTasks(ctx *DataContext, styles OverviewStyles, tasksTable table.Model, width, height int) string {
	if ctx == nil || len(ctx.Tasks) == 0 {
		return theme.TextMuted.Render("\n  No tasks for this project\n")
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	// Left pane: table.Model view
	leftPane := renderTasksTable(tasksTable, leftWidth, height)

	// Right pane: selected task details
	rightPane := renderTaskDetails(ctx, rightWidth, height, styles)

	// Join panes horizontally with separator
	result := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		theme.PanelSidebar.Copy().Height(height).Render(""),
		rightPane,
	)

	return result
}

// renderTasksTable renders the left pane table
func renderTasksTable(t table.Model, width, height int) string {
	return theme.PanelNormal.Copy().
		Width(width).
		Height(height).
		Render(t.View())
}

// renderTaskDetails renders the right pane with selected task info
func renderTaskDetails(ctx *DataContext, width, height int, styles OverviewStyles) string {
	if ctx.TasksIndex < 0 || ctx.TasksIndex >= len(ctx.Tasks) {
		return theme.PanelNoBorder.Copy().
			Width(width).
			Height(height).
			Render(theme.TextMuted.Render("Select a task to view details"))
	}

	task := ctx.Tasks[ctx.TasksIndex]

	var details strings.Builder

	// Header
	details.WriteString(styles.SectionTitle.Render("Task Details") + "\n\n")

	// Task info
	details.WriteString(renderTaskField("ID", string(task.ID), styles))
	details.WriteString(renderTaskField("Title", task.Title, styles))

	// State with color coding
	stateStr := formatTaskState(task.StateID)
	details.WriteString(renderTaskField("State", stateStr, styles))

	// Priority with color coding
	priorityStr := formatPriority(task.Priority)
	details.WriteString(renderTaskField("Priority", priorityStr, styles))

	// Description
	if task.Description != "" {
		details.WriteString("\n" + styles.Label.Render("Description:") + "\n")
		details.WriteString(styles.Value.Render(task.Description) + "\n")
	}

	// Tags
	if len(task.Tags) > 0 {
		details.WriteString(renderTaskField("Tags", strings.Join(task.Tags, ", "), styles))
	}

	// Session and Agent
	if task.SessionID != "" {
		details.WriteString(renderTaskField("Session", task.SessionID, styles))
	}
	if task.AgentID != "" {
		details.WriteString(renderTaskField("Agent", task.AgentID, styles))
	}

	// Timestamps
	if !task.CreatedAt.IsZero() {
		details.WriteString(renderTaskField("Created", task.CreatedAt.Format("2006-01-02 15:04:05"), styles))
	}
	if !task.UpdatedAt.IsZero() {
		details.WriteString(renderTaskField("Updated", task.UpdatedAt.Format("2006-01-02 15:04:05"), styles))
	}
	if !task.DueAt.IsZero() {
		details.WriteString(renderTaskField("Due", task.DueAt.Format("2006-01-02"), styles))
	}
	if !task.CompletedAt.IsZero() {
		details.WriteString(renderTaskField("Completed", task.CompletedAt.Format("2006-01-02 15:04:05"), styles))
	}

	// Annotations
	if len(ctx.TaskNotes) > 0 {
		details.WriteString("\n" + styles.Label.Render("Annotations:") + "\n")
		for _, note := range ctx.TaskNotes {
			timestamp := note.CreatedAt.Format("2006-01-02 15:04")
			details.WriteString(fmt.Sprintf("  [%s] %s\n", timestamp, note.Body))
		}
	}

	return theme.PanelNoBorder.Copy().
		Width(width).
		Height(height).
		Render(details.String())
}

// renderTaskField renders a label-value pair for tasks
func renderTaskField(label, value string, styles OverviewStyles) string {
	if value == "" {
		value = theme.TextMuted.Render("(none)")
	}
	return fmt.Sprintf("%s %s\n", styles.Label.Render(label+":"), styles.Value.Render(value))
}

// formatTaskState converts state ID to readable string
func formatTaskState(stateID int) string {
	switch stateID {
	case 1:
		return "Todo"
	case 2:
		return "In Progress"
	case 3:
		return "Done"
	default:
		return fmt.Sprintf("State %d", stateID)
	}
}

// formatPriority converts priority to readable string
func formatPriority(priority int) string {
	switch {
	case priority >= 4:
		return "High"
	case priority >= 2:
		return "Medium"
	default:
		return "Low"
	}
}
