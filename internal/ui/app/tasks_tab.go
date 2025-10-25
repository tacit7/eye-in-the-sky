package app

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderTasksTabView renders the tasks tab content
func (m *Model) renderTasksTabView() string {
	start := time.Now()
	now := start.Format("15:04:05.000")
	log.Printf("[PERF][%s] renderTasksTabView called with %d tasks", now, len(m.tasks))

	if len(m.tasks) == 0 {
		return m.styles.Subtle.Render("\n  No tasks found for this agent\n")
	}

	var sb strings.Builder

	// Sort tasks: priority (H>M>L), then by status (pending first), then by creation date
	sortedTasks := make([]domain.Task, len(m.tasks))
	copy(sortedTasks, m.tasks)

	sort.Slice(sortedTasks, func(i, j int) bool {
		// Archived tasks go to the bottom
		if sortedTasks[i].Archived && !sortedTasks[j].Archived {
			return false
		}
		if !sortedTasks[i].Archived && sortedTasks[j].Archived {
			return true
		}

		// Priority order: higher numbers first (5 > 4 > ... > 0)
		iPri := sortedTasks[i].Priority
		jPri := sortedTasks[j].Priority
		if iPri != jPri {
			return iPri > jPri
		}

		// Then by state (todo before in_progress before done)
		iState := sortedTasks[i].StateID
		jState := sortedTasks[j].StateID
		if iState != jState {
			return iState < jState // Lower state IDs first (1=todo < 2=in_progress < 3=done)
		}

		// Finally by creation date (newer first)
		return sortedTasks[i].CreatedAt.After(sortedTasks[j].CreatedAt)
	})

	// Define column widths for alignment
	idWidth := 8
	priorityWidth := 8
	statusWidth := 12
	descWidth := 50
	createdWidth := 17

	// Helper to pad right without extra spacing
	padRightStyled := func(text string, width int) string {
		return fmt.Sprintf("%-*s", width, text)
	}

	// Header - no trailing spaces after each label
	headerLine := padRightStyled("ID", idWidth) +
		padRightStyled("Priority", priorityWidth) +
		padRightStyled("Status", statusWidth) +
		padRightStyled("Description", descWidth) +
		fmt.Sprintf("%*s", createdWidth, "Created")

	sb.WriteString(m.styles.Primary.Render(headerLine))
	sb.WriteString("\n")
	sb.WriteString(m.styles.Border.Render(strings.Repeat("─", m.width-10)))
	sb.WriteString("\n")

	// Render tasks
	for i, task := range sortedTasks {
		selected := i == m.taskIndex

		// Format priority with color (0-5 scale)
		var priStyle lipgloss.Style
		var priDisplay string
		switch task.Priority {
		case 5:
			priStyle = m.styles.Error
			priDisplay = "[CRIT]"
		case 4, 3:
			priStyle = m.styles.Error
			priDisplay = "[HIGH]"
		case 2:
			priStyle = m.styles.Warning
			priDisplay = "[MED]"
		case 1:
			priStyle = m.styles.Subtle
			priDisplay = "[LOW]"
		default:
			priStyle = m.styles.Subtle
			priDisplay = "[-]"
		}

		// Format state with color
		var statusStyle lipgloss.Style
		switch task.WorkflowStatus {
		case "done":
			statusStyle = m.styles.Success
		case "in_progress":
			statusStyle = m.styles.Warning
		case "todo":
			statusStyle = m.styles.Primary
		default:
			statusStyle = m.styles.Subtle
		}

		// Truncate description if needed
		desc := task.Description
		if len(desc) > descWidth-3 {
			desc = desc[:descWidth-6] + "..."
		}

		// Format line with aligned columns (matching header)
		// Pad text before applying styles to keep visible width correct
		taskID := fmt.Sprintf("%s", string(task.ID)[:8])
		createdStr := task.CreatedAt.Format("Jan 2 15:04")

		line := padRightStyled(taskID, idWidth) +
			priStyle.Render(fmt.Sprintf("%-*s", priorityWidth, priDisplay)) +
			statusStyle.Render(fmt.Sprintf("%-*s", statusWidth, task.WorkflowStatus)) +
			padRightStyled(desc, descWidth) +
			fmt.Sprintf("%*s", createdWidth, createdStr)

		if selected {
			sb.WriteString(m.styles.Selected.Render(line))
		} else {
			sb.WriteString(line)
		}
		sb.WriteString("\n")
	}

	// Show task details if one is selected
	if m.taskIndex >= 0 && m.taskIndex < len(sortedTasks) {
		task := sortedTasks[m.taskIndex]
		sb.WriteString("\n")
		sb.WriteString(m.styles.Border.Render(strings.Repeat("─", m.width-10)))
		sb.WriteString("\n")
		sb.WriteString(renderTaskDetails(task, m.styles))
	}

	result := sb.String()
	renderTime := time.Since(start)
	endTime := time.Now().Format("15:04:05.000")
	log.Printf("[PERF][%s] Task render complete (%d tasks) took %v", endTime, len(sortedTasks), renderTime)
	return result
}

// renderTaskDetails renders detailed view of a single task
func renderTaskDetails(task domain.Task, styles Styles) string {
	var sb strings.Builder

	sb.WriteString(styles.SectionTitle.Render("Task Details"))
	sb.WriteString("\n\n")

	// Task ID
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		styles.Label.Render("ID:"),
		styles.Value.Render(string(task.ID))))

	// Priority
	sb.WriteString(fmt.Sprintf("  %s %d\n",
		styles.Label.Render("Priority:"),
		task.Priority))

	// State
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		styles.Label.Render("State:"),
		styles.Value.Render(task.WorkflowStatus)))

	// Project
	if task.ProjectID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			styles.Label.Render("Project:"),
			styles.Value.Render(task.ProjectID)))
	}

	// Tags
	if len(task.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			styles.Label.Render("Tags:"),
			styles.Value.Render(strings.Join(task.Tags, ", "))))
	}

	// Notes
	if len(task.Notes) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styles.Label.Render("Notes:"))
		sb.WriteString("\n")
		for _, note := range task.Notes {
			sb.WriteString(fmt.Sprintf("  • %s\n", note.Body))
		}
	}

	// Session/Agent info
	if task.SessionID != "" {
		sb.WriteString(fmt.Sprintf("\n  %s %s\n",
			styles.Subtle.Render("Session:"),
			styles.Subtle.Render(truncateID(task.SessionID, 8))))
	}
	if task.AgentID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			styles.Subtle.Render("Agent:"),
			styles.Subtle.Render(truncateID(task.AgentID, 8))))
	}

	return sb.String()
}
