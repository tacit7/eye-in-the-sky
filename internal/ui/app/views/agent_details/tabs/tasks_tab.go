package tabs

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// RenderTasks renders the tasks tab content
// TODO: State migration in progress - needs taskIndex, width, detail view state
func RenderTasks(ctx *DataContext, overviewStyles OverviewStyles) string {
	start := time.Now()
	now := start.Format("15:04:05.000")
	log.Printf("[PERF][%s] RenderTasks called with %d tasks", now, len(ctx.Tasks))

	if len(ctx.Tasks) == 0 {
		return overviewStyles.Subtle.Render("\n  No tasks found for this agent\n")
	}

	var sb strings.Builder

	// Sort tasks: priority (H>M>L), then by status (pending first), then by creation date
	sortedTasks := make([]domain.Task, len(ctx.Tasks))
	copy(sortedTasks, ctx.Tasks)

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

	// Simple header for now - TODO: Add proper column alignment
	sb.WriteString(overviewStyles.Primary.Render("ID       Priority  Status        Description"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// Render tasks - simplified version
	for _, task := range sortedTasks {
		// Format priority display
		var priDisplay string
		switch task.Priority {
		case 5:
			priDisplay = "[CRIT]"
		case 4, 3:
			priDisplay = "[HIGH]"
		case 2:
			priDisplay = "[MED]"
		case 1:
			priDisplay = "[LOW]"
		default:
			priDisplay = "[-]"
		}

		// Truncate description if needed
		desc := task.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}

		// Format task ID
		taskID := string(task.ID)
		if len(taskID) > 8 {
			taskID = taskID[:8]
		}

		line := fmt.Sprintf("%-8s %-9s %-13s %s\n",
			taskID, priDisplay, task.WorkflowStatus, desc)
		sb.WriteString(line)
	}

	result := sb.String()
	renderTime := time.Since(start)
	endTime := time.Now().Format("15:04:05.000")
	log.Printf("[PERF][%s] Task render complete (%d tasks) took %v", endTime, len(sortedTasks), renderTime)
	return result
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
