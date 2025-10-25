package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderTasksTabView renders the tasks tab content
func (m *Model) renderTasksTabView() string {
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

	// Header
	sb.WriteString(m.styles.Primary.Render(fmt.Sprintf("%-5s %-8s %-10s %-60s %s\n",
		"ID", "Priority", "Status", "Description", "Created")))
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
		if len(desc) > 58 {
			desc = desc[:55] + "..."
		}

		// Format line
		line := fmt.Sprintf("%-5s %s %-10s %-60s %s",
			fmt.Sprintf("%s", string(task.ID)[:8]), // First 8 chars of UUID
			priStyle.Render(priDisplay),
			statusStyle.Render(task.WorkflowStatus),
			desc,
			task.CreatedAt.Format("Jan 2 15:04"),
		)

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

	return sb.String()
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
			styles.Subtle.Render(task.SessionID)))
	}
	if task.AgentID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			styles.Subtle.Render("Agent:"),
			styles.Subtle.Render(task.AgentID)))
	}

	return sb.String()
}
