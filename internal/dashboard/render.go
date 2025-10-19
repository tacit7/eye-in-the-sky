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
	fmt.Fprintf(v, "%-10s %-10s %-12s %-20s %-15s %-30s\n", "STATUS", "AGENT", "SESSION", "DESCRIPTION", "PROJECT", "TASK")
	fmt.Fprintln(v, "──────────────────────────────────────────────────────────────────────────────────────────────────────")

	// Agents
	for _, agent := range a.agents {
		statusIcon := getStatusIcon(agent.Status)
		agentID := agent.ID[:8]

		description := "-"
		if agent.Description != nil {
			description = *agent.Description
			if len(description) > 18 {
				description = description[:15] + "..."
			}
		}

		sessionName := "-"
		// Get session name or ID (first part only) if available
		if agent.CurrentSessionID != nil && *agent.CurrentSessionID != "" {
			session, err := a.db.GetSession(*agent.CurrentSessionID)
			if err == nil && session != nil {
				if session.Name != nil && *session.Name != "" {
					sessionName = *session.Name
				} else {
					// Show first part of session ID (before - or _)
					sessionName = getFirstPart(session.ID)
				}
				if len(sessionName) > 10 {
					sessionName = sessionName[:10]
				}
			} else {
				// Session not found, show first part of current session ID
				sessionName = getFirstPart(*agent.CurrentSessionID)
				if len(sessionName) > 10 {
					sessionName = sessionName[:10]
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
		if len(task) > 28 {
			task = task[:25] + "..."
		}

		fmt.Fprintf(v, "%s %-10s %-12s %-20s %-15s %-30s\n", statusIcon, agentID, sessionName, description, projectName, task)
	}

	return nil
}

// getFirstPart extracts the first part of a session ID (before - or _)
func getFirstPart(s string) string {
	// Split on - or _
	for i, c := range s {
		if c == '-' || c == '_' {
			return s[:i]
		}
	}
	return s
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
