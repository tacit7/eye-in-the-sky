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
		// Deleted tasks go to the bottom
		if sortedTasks[i].Status == "deleted" && sortedTasks[j].Status != "deleted" {
			return false
		}
		if sortedTasks[i].Status != "deleted" && sortedTasks[j].Status == "deleted" {
			return true
		}

		// Priority order: H > M > L > (no priority)
		iPri := GetPriorityWeight(sortedTasks[i].Priority)
		jPri := GetPriorityWeight(sortedTasks[j].Priority)
		if iPri != jPri {
			return iPri > jPri
		}

		// Then by status (pending before completed)
		if sortedTasks[i].Status != sortedTasks[j].Status {
			if sortedTasks[i].Status == "pending" {
				return true
			}
			if sortedTasks[j].Status == "pending" {
				return false
			}
		}

		// Finally by creation date (newer first)
		return sortedTasks[i].Entry.After(sortedTasks[j].Entry)
	})

	// Header
	sb.WriteString(m.styles.Primary.Render(fmt.Sprintf("%-5s %-8s %-10s %-60s %s\n",
		"ID", "Priority", "Status", "Description", "Created")))
	sb.WriteString(m.styles.Border.Render(strings.Repeat("─", m.width-10)))
	sb.WriteString("\n")

	// Render tasks
	for i, task := range sortedTasks {
		selected := i == m.taskIndex

		// Format priority with color
		var priStyle lipgloss.Style
		var priDisplay string
		switch task.Priority {
		case "H":
			priStyle = m.styles.Error
			priDisplay = "[HIGH]"
		case "M":
			priStyle = m.styles.Warning
			priDisplay = "[MED]"
		case "L":
			priStyle = m.styles.Subtle
			priDisplay = "[LOW]"
		default:
			priStyle = m.styles.Subtle
			priDisplay = "[-]"
		}

		// Format status with color
		var statusStyle lipgloss.Style
		switch task.Status {
		case "completed":
			statusStyle = m.styles.Success
		case "pending":
			statusStyle = m.styles.Warning
		case "deleted":
			statusStyle = m.styles.Subtle.Strikethrough(true)
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
			fmt.Sprintf("%d", task.ID),
			priStyle.Render(priDisplay),
			statusStyle.Render(task.Status),
			desc,
			task.Entry.Format("Jan 2 15:04"),
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

	// Task UUID
	if task.UUID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			styles.Label.Render("UUID:"),
			styles.Value.Render(task.UUID)))
	}

	// Tags
	if len(task.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			styles.Label.Render("Tags:"),
			styles.Value.Render(strings.Join(task.Tags, ", "))))
	}

	// Annotations
	if len(task.Annotations) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styles.Label.Render("Annotations:"))
		sb.WriteString("\n")
		for _, ann := range task.Annotations {
			sb.WriteString(fmt.Sprintf("  • %s\n", ann.Description))
		}
	}

	// Virtual tags (computed)
	if len(task.VirtualTags) > 0 {
		sb.WriteString(fmt.Sprintf("\n  %s %s\n",
			styles.Label.Render("Virtual Tags:"),
			styles.Subtle.Render(strings.Join(task.VirtualTags, ", "))))
	}

	return sb.String()
}
