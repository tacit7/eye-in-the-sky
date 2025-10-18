package dashboard

import (
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// renderAgents renders the agent list in the main view
func (a *App) renderAgents() error {
	v, err := a.gui.View(viewMain)
	if err != nil {
		return err
	}

	v.Clear()

	if len(a.agents) == 0 {
		fmt.Fprintln(v, "\n  No active agents")
		return nil
	}

	// Header
	fmt.Fprintf(v, "%-10s %-10s %-20s %-15s %-25s\n", "STATUS", "AGENT", "SESSION", "PROJECT", "TASK")
	fmt.Fprintln(v, "────────────────────────────────────────────────────────────────────────────────")

	// Agents
	for _, agent := range a.agents {
		statusIcon := getStatusIcon(agent.Status)
		agentID := agent.ID[:8]
		sessionName := "-"

		// Get session name if available
		if agent.CurrentSessionID != nil && *agent.CurrentSessionID != "" {
			session, err := a.db.GetSession(*agent.CurrentSessionID)
			if err == nil && session != nil && session.Name != nil {
				sessionName = *session.Name
				if len(sessionName) > 18 {
					sessionName = sessionName[:15] + "..."
				}
			}
		}

		projectName := "N/A"
		if agent.ProjectName != nil {
			projectName = *agent.ProjectName
			if len(projectName) > 13 {
				projectName = projectName[:10] + "..."
			}
		}

		task := "No task"
		if agent.CurrentTask != nil {
			task = *agent.CurrentTask
		} else if agent.FeatureDescription != nil {
			task = *agent.FeatureDescription
		}
		if len(task) > 23 {
			task = task[:20] + "..."
		}

		fmt.Fprintf(v, "%s %-10s %-20s %-15s %-25s\n", statusIcon, agentID, sessionName, projectName, task)
	}

	return nil
}

// getStatusIcon returns a colored icon for the status
func getStatusIcon(status string) string {
	switch status {
	case database.StatusActive:
		return "\033[1;32m●\033[0m ACTIVE  " // Green
	case database.StatusWorking:
		return "\033[1;34m●\033[0m WORKING " // Blue
	case database.StatusIdle:
		return "\033[1;33m●\033[0m IDLE    " // Yellow
	case database.StatusCompleted:
		return "\033[1;36m✓\033[0m COMPLETE" // Cyan
	case database.StatusFailed:
		return "\033[1;31m✗\033[0m FAILED  " // Red
	default:
		return "\033[1;37m?\033[0m UNKNOWN " // White
	}
}
