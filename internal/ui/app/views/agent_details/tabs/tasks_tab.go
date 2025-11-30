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

// RenderTasksSplitPane renders tasks in split-pane layout with viewport
func RenderTasksSplitPane(ctx *DataContext, overviewStyles OverviewStyles, selectedIndex int, width int) string {
	start := time.Now()
	now := start.Format("15:04:05.000")
	log.Printf("[PERF][%s] RenderTasks called with %d tasks", now, len(ctx.Tasks))

	if len(ctx.Tasks) == 0 {
		return theme.TextMuted.Render("\n  No tasks found for this agent\n")
	}

	// Tasks are already sorted in model.loadTasks()
	allTasks := ctx.Tasks
	totalCount := len(allTasks)

	// Limit to last 5 tasks for display
	tasks := allTasks
	if len(allTasks) > 5 {
		tasks = allTasks[:5]
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1
	viewportHeight := 30

	// Render left pane (task list with count header)
	leftPane := renderTaskList(tasks, selectedIndex, leftWidth, totalCount)

	// Render right pane (selected task details with viewport)
	rightPane := ""
	if selectedIndex >= 0 && selectedIndex < len(tasks) {
		// Get task details content
		content := renderTaskDetailsContent(tasks[selectedIndex], ctx.TaskNotes, rightWidth, ctx.MarkdownRenderer)

		// Update viewport size and content
		ctx.TaskDetailViewport.Width = rightWidth
		ctx.TaskDetailViewport.Height = viewportHeight
		ctx.TaskDetailViewport.SetContent(content)

		// Render viewport
		rightPane = ctx.TaskDetailViewport.View()
	} else {
		rightPane = theme.TextMuted.Copy().
			Width(rightWidth).
			Render("Select a task to view details")
	}

	// Join panes horizontally
	result := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		theme.PanelSidebar.Copy().Height(viewportHeight).Render(""),
		rightPane,
	)

	renderTime := time.Since(start)
	endTime := time.Now().Format("15:04:05.000")
	log.Printf("[PERF][%s] Task render complete (%d tasks) took %v", endTime, len(tasks), renderTime)
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
func renderTaskList(tasks []domain.Task, selectedIndex int, width int, totalCount int) string {
	var sb strings.Builder

	// Add count header
	countHeader := fmt.Sprintf("Showing %d of %d tasks", len(tasks), totalCount)
	sb.WriteString(theme.TextSubtitle.Render(countHeader))
	sb.WriteString("\n")
	sb.WriteString(theme.TextMuted.Render(strings.Repeat("─", width-4)))
	sb.WriteString("\n\n")

	// Task items
	for i, task := range tasks {
		isSelected := i == selectedIndex

		// Cursor indicator
		cursor := "  "
		if isSelected {
			cursor = "> "
		}

		// Status icon
		statusIcon := getTaskStateIcon(task.StateID)

		// Priority indicator
		priDisplay := formatPriority(task.Priority)

		// Title with truncation
		title := task.Title
		if len(title) > width-12 {
			title = title[:width-15] + "..."
		}

		// Build line: cursor status priority title
		line := fmt.Sprintf("%s%s %s %s", cursor, statusIcon, priDisplay, title)

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

// renderTaskDetailsContent renders task details content for viewport
func renderTaskDetailsContent(task domain.Task, taskNotes []domain.TaskNote, width int, mdRenderer MarkdownRenderer) string {
	var sb strings.Builder

	// Task header
	sb.WriteString(theme.TextTitle.Render(task.Title))
	sb.WriteString("\n\n")

	// Description
	sb.WriteString(task.Description)
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
			if note.Author != "" {
				author = note.Author
			}
			noteHeader := fmt.Sprintf("[%s] %s:", timestamp, author)
			sb.WriteString(theme.TextTimestamp.Render(noteHeader))
			sb.WriteString("\n")

			// Render markdown if renderer available, otherwise plain text
			renderedBody := note.Body
			if mdRenderer != nil {
				rendered, err := mdRenderer.Render(note.Body)
				if err == nil {
					renderedBody = rendered
				}
				// On error, fall back to plain text
			}

			sb.WriteString(renderedBody)
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// formatPriority returns a visual priority indicator
func formatPriority(priority int) string {
	switch priority {
	case 5:
		return theme.TextError.Render("C") // Critical
	case 4:
		return theme.TextWarning.Render("H") // High
	case 3:
		return theme.TextHighlight.Render("H") // High
	case 2:
		return theme.TextValue.Render("M") // Medium
	case 1:
		return theme.TextMuted.Render("L") // Low
	default:
		return theme.TextMuted.Render("-") // None
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

// getTaskStateIcon returns an icon for the task state
func getTaskStateIcon(stateID int) string {
	switch stateID {
	case 1:
		return "\uf096" // nf-fa-square_o (todo)
	case 2:
		return "\uf04b" // nf-fa-play (in progress)
	case 3:
		return "\uf00c" // nf-fa-check (done)
	default:
		return "\uf128" // nf-fa-question
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
