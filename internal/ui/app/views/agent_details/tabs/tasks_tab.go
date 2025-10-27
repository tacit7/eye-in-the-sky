package tabs

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// RenderTasks renders the tasks tab with split-pane view
// Left pane: task list, Right pane: description + annotations
func RenderTasks(ctx *DataContext, overviewStyles OverviewStyles) string {
	return RenderTasksSplitPane(ctx, overviewStyles, ctx.SelectedTaskIndex, ctx.Width)
}

// RenderTasksSplitPane renders tasks in split-pane layout
func RenderTasksSplitPane(ctx *DataContext, overviewStyles OverviewStyles, selectedIndex int, width int) string {
	start := time.Now()
	now := start.Format("15:04:05.000")
	log.Printf("[PERF][%s] RenderTasks called with %d tasks", now, len(ctx.Tasks))

	if len(ctx.Tasks) == 0 {
		return theme.TextMuted.Render("\n  No tasks found for this agent\n")
	}

	// Sort tasks
	sortedTasks := sortTasks(ctx.Tasks)

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	// Render left pane (task list)
	leftPane := renderTaskList(sortedTasks, selectedIndex, leftWidth)

	// Render right pane (selected task details)
	rightPane := ""
	if selectedIndex >= 0 && selectedIndex < len(sortedTasks) {
		rightPane = renderTaskDetails(sortedTasks[selectedIndex], ctx.TaskNotes, rightWidth)
	} else {
		rightPane = theme.TextMuted.Copy().
			Width(rightWidth).
			Render("Select a task to view details")
	}

	// Join panes horizontally
	result := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		theme.PanelSidebar.Copy().Height(30).Render(""),
		rightPane,
	)

	renderTime := time.Since(start)
	endTime := time.Now().Format("15:04:05.000")
	log.Printf("[PERF][%s] Task render complete (%d tasks) took %v", endTime, len(sortedTasks), renderTime)
	return result
}

// sortTasks sorts tasks by priority, status, and creation date
func sortTasks(tasks []domain.Task) []domain.Task {
	sorted := make([]domain.Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		// Archived tasks to bottom
		if sorted[i].Archived != sorted[j].Archived {
			return !sorted[i].Archived
		}
		// Higher priority first
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority > sorted[j].Priority
		}
		// Earlier state first (todo < in_progress < done)
		if sorted[i].StateID != sorted[j].StateID {
			return sorted[i].StateID < sorted[j].StateID
		}
		// Newer first
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	return sorted
}

// renderTaskList renders the left pane task list
func renderTaskList(tasks []domain.Task, selectedIndex int, width int) string {
	var sb strings.Builder

	// Header
	sb.WriteString(theme.TextSubtitle.Render("Tasks"))
	sb.WriteString("\n")
	sb.WriteString(theme.TextMuted.Render(strings.Repeat("─", width-2)))
	sb.WriteString("\n\n")

	// Task items
	for i, task := range tasks {
		isSelected := i == selectedIndex

		// Priority indicator
		priDisplay := formatPriority(task.Priority)

		// Title with truncation
		title := task.Title
		if len(title) > width-20 {
			title = title[:width-23] + "..."
		}

		// Build line
		line := fmt.Sprintf("%s %s", priDisplay, title)

		// Apply style
		if isSelected {
			line = theme.ListItemSelected.Copy().Width(width - 2).Render(line)
		} else {
			line = theme.ListItem.Copy().Width(width - 2).Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return theme.List.Copy().Width(width).Render(sb.String())
}

// renderTaskDetails renders the right pane with description and annotations
func renderTaskDetails(task domain.Task, taskNotes []domain.TaskNote, width int) string {
	var sb strings.Builder

	// Task header
	sb.WriteString(theme.TextTitle.Copy().Width(width-4).Render(task.Title))
	sb.WriteString("\n\n")

	// Status and priority
	statusLine := fmt.Sprintf("%s  %s",
		theme.TextLabel.Render("Status:"),
		theme.TextValue.Render(task.WorkflowStatus))
	sb.WriteString(statusLine)
	sb.WriteString("\n")

	priLine := fmt.Sprintf("%s  %s",
		theme.TextLabel.Render("Priority:"),
		theme.TextValue.Render(formatPriorityText(task.Priority)))
	sb.WriteString(priLine)
	sb.WriteString("\n\n")

	// Description section
	sb.WriteString(theme.TextSubtitle.Render("Description"))
	sb.WriteString("\n")
	sb.WriteString(theme.TextMuted.Render(strings.Repeat("─", width-4)))
	sb.WriteString("\n")
	descStyle := theme.TextNormal.Copy().Width(width - 4)
	sb.WriteString(descStyle.Render(task.Description))
	sb.WriteString("\n\n")

	// Annotations section
	sb.WriteString(theme.TextSubtitle.Render("Annotations"))
	sb.WriteString("\n")
	sb.WriteString(theme.TextMuted.Render(strings.Repeat("─", width-4)))
	sb.WriteString("\n")

	if len(taskNotes) == 0 {
		sb.WriteString(theme.TextMuted.Render("  No annotations"))
	} else {
		for _, note := range taskNotes {
			timestamp := note.CreatedAt.Format("15:04:05")
			author := "system"
			if note.Author != nil && *note.Author != "" {
				author = *note.Author
			}
			noteHeader := fmt.Sprintf("[%s] %s:", timestamp, author)
			sb.WriteString(theme.TextTimestamp.Render(noteHeader))
			sb.WriteString("\n")
			sb.WriteString(theme.TextNormal.Copy().Width(width-4).Render(note.Body))
			sb.WriteString("\n\n")
		}
	}

	return theme.PanelNoBorder.Copy().Width(width).Render(sb.String())
}

// formatPriority returns a visual priority indicator
func formatPriority(priority int) string {
	switch priority {
	case 5:
		return theme.TextError.Render("█████")
	case 4:
		return theme.TextWarning.Render("████░")
	case 3:
		return theme.TextHighlight.Render("███░░")
	case 2:
		return theme.TextValue.Render("██░░░")
	case 1:
		return theme.TextMuted.Render("█░░░░")
	default:
		return theme.TextMuted.Render("░░░░░")
	}
}

// formatPriorityText returns text description of priority
func formatPriorityText(priority int) string {
	switch priority {
	case 5:
		return "CRITICAL"
	case 4:
		return "HIGH"
	case 3:
		return "HIGH"
	case 2:
		return "MEDIUM"
	case 1:
		return "LOW"
	default:
		return "NONE"
	}
}

// TODO: The following methods require Model state migration (taskIndex, width, detail view state)
// They will be re-implemented once state is properly migrated to the Model layer

/*
// renderTaskDetails renders detailed view of a single task
func renderTaskDetails(task domain.Task, styles OverviewStyles) string {
	// TODO: Implement after state migration
	return ""
}

// renderTaskDetailSplitPane renders the split-pane detail view for a task
// TODO: Needs activeTaskDetailView state migration
func renderTaskDetailSplitPane(task domain.Task, styles OverviewStyles) string {
	// TODO: Implement after state migration
	return ""
}

// renderTaskCommitsPane renders the commits pane
// TODO: Needs selectedCommitIdx state migration
func renderTaskCommitsPane(task domain.Task, styles OverviewStyles) string {
	// TODO: Implement after state migration
	return ""
}

// renderTaskNotesPane renders the notes pane
// TODO: Needs selectedNoteIdx state migration
func renderTaskNotesPane(task domain.Task, styles OverviewStyles) string {
	// TODO: Implement after state migration
	return ""
}

// renderCommitDetailView renders the full detail view of a selected commit
// TODO: Needs activeTaskDetailView state migration
func renderCommitDetailView(task domain.Task, styles OverviewStyles) string {
	// TODO: Implement after state migration
	return ""
}
*/
