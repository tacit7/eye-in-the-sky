package presenters

import (
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// OverviewData contains preprocessed data for the overview tab renderer
// This separates presentation (rendering) from business logic (data preparation)
type OverviewData struct {
	AgentInfo  domain.Agent
	Commits    []domain.Commit
	Notes      []domain.Note
	Tasks      []domain.Task
	Actions    []domain.Action
	TaskCounts map[string]int
	LastActive string
}

// BuildOverviewData prepares data for the overview tab renderer
// Input: raw model state (agent, commits, notes, tasks, actions)
// Output: preprocessed data ready for rendering
// This layer handles all business logic calculations so renderers stay pure
func BuildOverviewData(
	agent *domain.Agent,
	commits []domain.Commit,
	notes []domain.Note,
	tasks []domain.Task,
	actions []domain.Action,
) OverviewData {
	data := OverviewData{}

	if agent != nil {
		data.AgentInfo = *agent
	}

	data.Commits = commits
	data.Notes = notes
	data.Tasks = tasks
	data.Actions = actions
	data.TaskCounts = CountTasksByState(tasks)

	// Calculate last activity
	if !data.AgentInfo.LastActivityAt.IsZero() {
		data.LastActive = FormatDuration(time.Since(data.AgentInfo.LastActivityAt))
	}

	return data
}

// CountTasksByState counts tasks grouped by their state
// Exported to be used by app package and other presenters
func CountTasksByState(tasks []domain.Task) map[string]int {
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

// FormatDuration formats a duration for display
// Exported for use in overview and other presenters
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if minutes > 0 {
		if minutes == 1 {
			return "1 min"
		}
		return fmt.Sprintf("%d min", minutes)
	}
	if seconds == 1 {
		return "1 sec"
	}
	return fmt.Sprintf("%d sec", seconds)
}
