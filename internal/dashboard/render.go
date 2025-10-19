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
	fmt.Fprintf(v, "%-10s %-10s %-20s %-15s %-18s %-15s %-25s\n", "STATUS", "AGENT", "DESCRIPTION", "SOURCE", "SESSION", "PROJECT", "TASK")
	fmt.Fprintln(v, "───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")

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

		// Source with window ID for desktop agents
		source := agent.Source
		if agent.Source == database.SourceDesktop && agent.WindowID != nil && *agent.WindowID != "" {
			windowID := *agent.WindowID
			if len(windowID) > 8 {
				windowID = windowID[:5] + "..."
			}
			source = fmt.Sprintf("desktop:%s", windowID)
		}
		if len(source) > 13 {
			source = source[:10] + "..."
		}

		sessionName := "-"
		// Get session name or ID if available
		if agent.CurrentSessionID != nil && *agent.CurrentSessionID != "" {
			session, err := a.db.GetSession(*agent.CurrentSessionID)
			if err == nil && session != nil {
				if session.Name != nil && *session.Name != "" {
					sessionName = *session.Name
				} else {
					// Show truncated session ID if no name
					sessionName = session.ID
				}
				if len(sessionName) > 16 {
					sessionName = sessionName[:13] + "..."
				}
			} else {
				// Session not found, show truncated current session ID
				sessionName = *agent.CurrentSessionID
				if len(sessionName) > 16 {
					sessionName = sessionName[:13] + "..."
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

		fmt.Fprintf(v, "%s %-10s %-20s %-15s %-18s %-15s %-25s\n", statusIcon, agentID, description, source, sessionName, projectName, task)
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
