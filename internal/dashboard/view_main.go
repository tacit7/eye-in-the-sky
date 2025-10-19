package dashboard

import (
	"fmt"
	"time"

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
	fmt.Fprintf(v, "%-10s %-8s %-10s %-12s %-22s %-16s %-13s %-26s\n", "STATUS", "ACTIVITY", "AGENT", "SESSION", "WINDOW", "DESCRIPTION", "PROJECT", "TASK")
	fmt.Fprintln(v, "────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")

	// Agents
	for _, agent := range a.agents {
		statusIcon := getStatusIcon(agent)
		agentID := agent.ID[:8]

		description := "-"
		if agent.Description != nil {
			description = *agent.Description
			if len(description) > 14 {
				description = description[:11] + "..."
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

		windowID := "-"
		if agent.WindowID != nil && *agent.WindowID != "" {
			windowID = *agent.WindowID
			if len(windowID) > 20 {
				windowID = windowID[:20]
			}
		}

		projectName := "N/A"
		if agent.ProjectName != nil {
			projectName = *agent.ProjectName
			if len(projectName) > 11 {
				projectName = projectName[:8] + "..."
			}
		}

		task := "No task"
		if agent.CurrentTask != nil {
			task = *agent.CurrentTask
		} else if agent.FeatureDescription != nil {
			task = *agent.FeatureDescription
		}
		if len(task) > 24 {
			task = task[:21] + "..."
		}

		// Get last action timestamp
		var lastActionTime *time.Time
		actions, err := a.db.GetActionsForAgent(agent.ID, 1)
		if err == nil && len(actions) > 0 {
			lastActionTime = &actions[0].Timestamp
		}

		// Format last activity time from last action
		activityTime := formatLastActivity(lastActionTime)

		fmt.Fprintf(v, "%s %-8s %-10s %-12s %-22s %-16s %-13s %-26s\n", statusIcon, activityTime, agentID, sessionName, windowID, description, projectName, task)
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

// formatLastActivity formats the last activity time
// Shows time (HH:MM) if less than 24 hours, otherwise shows date (Mon DD)
func formatLastActivity(lastActivity *time.Time) string {
	if lastActivity == nil {
		return "-"
	}

	now := time.Now()
	duration := now.Sub(*lastActivity)

	// Less than 24 hours: show time
	if duration < 24*time.Hour {
		return lastActivity.Format("15:04")
	}

	// More than 24 hours: show month and day
	return lastActivity.Format("Jan 02")
}

// getStatusIcon returns a colored icon for the status
func getStatusIcon(agent *database.Agent) string {
	// Check if agent is stale or unknown based on last activity
	if agent.LastActivityAt != nil {
		timeSinceActivity := time.Since(*agent.LastActivityAt)

		// For active/working/idle agents, check activity time
		if agent.Status == database.StatusActive || agent.Status == database.StatusWorking || agent.Status == database.StatusIdle {
			// >1 hour inactive = unknown (might be dead)
			if timeSinceActivity > 1*time.Hour {
				return "\033[1;90m?\033[0m UNKNOWN " // Gray question mark
			}
			// >30 minutes inactive = stale
			if timeSinceActivity > 30*time.Minute {
				return "\033[1;90m●\033[0m STALE   " // Gray/dim
			}
		}
	}

	switch agent.Status {
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
