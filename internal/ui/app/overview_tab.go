package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderOverviewTab renders the overview tab content
func (m *Model) renderOverviewTab() string {
	if m.selectedAgent == nil {
		return "No agent selected"
	}

	var sb strings.Builder
	agent := *m.selectedAgent

	// Basic Information Section
	sb.WriteString(m.styles.SectionTitle.Render("📋 Agent Information"))
	sb.WriteString("\n\n")

	// Agent ID and Status
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		m.styles.Label.Render("Agent ID:"),
		m.styles.Value.Render(string(agent.ID))))

	sb.WriteString(fmt.Sprintf("  %s %s\n",
		m.styles.Label.Render("Status:"),
		GetStatusStyle(agent.Status, m.styles).Render(string(agent.Status))))

	// Description
	if agent.FeatureDesc != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Description:"),
			m.styles.Value.Render(agent.FeatureDesc)))
	}

	// Project (for desktop agents)
	if agent.ProjectName != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Project:"),
			m.styles.Value.Render(agent.ProjectName)))
	}

	// Current Task
	if agent.CurrentTask != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Current Task:"),
			m.styles.Value.Render(agent.CurrentTask)))
	}

	// Source
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		m.styles.Label.Render("Source:"),
		m.styles.Value.Render(string(agent.Source))))

	// Worktree Path (if applicable)
	if agent.GitWorktreePath != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Worktree:"),
			m.styles.Value.Render(agent.GitWorktreePath)))
	}

	// Session ID
	if agent.SessionID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Session ID:"),
			m.styles.Value.Render(agent.SessionID)))
	}

	// Parent Session ID (if it's a subagent)
	if agent.ParentSessionID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Parent Session:"),
			m.styles.Value.Render(agent.ParentSessionID)))
	}

	// Parent Agent ID (if it's a subagent)
	if agent.ParentAgentID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Parent Agent:"),
			m.styles.Value.Render(string(agent.ParentAgentID))))
	}

	// Window ID (for desktop agents)
	if agent.WindowID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Window ID:"),
			m.styles.Value.Render(agent.WindowID)))
	}

	// Timing Information
	sb.WriteString("\n")
	sb.WriteString(m.styles.SectionTitle.Render("⏱️ Timing"))
	sb.WriteString("\n\n")

	// Created At
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		m.styles.Label.Render("Created:"),
		m.styles.Value.Render(agent.CreatedAt.Format("Jan 2, 2006 15:04:05 MST"))))

	// Updated At
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		m.styles.Label.Render("Updated:"),
		m.styles.Value.Render(agent.UpdatedAt.Format("Jan 2, 2006 15:04:05 MST"))))

	// Last Activity
	if !agent.LastActivityAt.IsZero() {
		elapsed := time.Since(agent.LastActivityAt)
		activityStr := fmt.Sprintf("%s (%s ago)",
			agent.LastActivityAt.Format("15:04:05"),
			FormatDuration(elapsed))

		var activityStyle = m.styles.Subtle
		if elapsed < 5*time.Minute {
			activityStyle = m.styles.Success
		} else if elapsed < 30*time.Minute {
			activityStyle = m.styles.Warning
		}

		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Last Activity:"),
			activityStyle.Render(activityStr)))
	}

	// Session Duration
	if agent.Status != "completed" && agent.Status != "failed" {
		duration := time.Since(agent.CreatedAt)
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Session Duration:"),
			m.styles.Value.Render(FormatDuration(duration))))
	} else if agent.CompletedAt != nil {
		duration := agent.CompletedAt.Sub(agent.CreatedAt)
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Total Duration:"),
			m.styles.Value.Render(FormatDuration(duration))))
	}

	// Recent Commits Section
	if len(m.commits) > 0 {
		sb.WriteString("\n")
		sb.WriteString(m.styles.SectionTitle.Render("📝 Recent Commits"))
		sb.WriteString("\n\n")

		// Show up to 5 most recent commits
		maxCommits := 5
		if len(m.commits) < maxCommits {
			maxCommits = len(m.commits)
		}

		for i := 0; i < maxCommits; i++ {
			commit := m.commits[i]
			// Format: hash (time ago) - message
			elapsed := time.Since(commit.Timestamp)
			commitLine := fmt.Sprintf("  %s %s - %s\n",
				m.styles.Git.Render(string(commit.Hash)[:8]),
				m.styles.Subtle.Render(fmt.Sprintf("(%s ago)", FormatDuration(elapsed))),
				commit.Message)
			sb.WriteString(commitLine)
		}

		if len(m.commits) > maxCommits {
			sb.WriteString(m.styles.Subtle.Render(
				fmt.Sprintf("  ... and %d more commits", len(m.commits)-maxCommits)))
			sb.WriteString("\n")
		}
	}

	// Notes Summary
	if len(m.notes) > 0 {
		sb.WriteString("\n")
		sb.WriteString(m.styles.SectionTitle.Render("📌 Notes Summary"))
		sb.WriteString("\n\n")

		// Show count and most recent note preview
		sb.WriteString(fmt.Sprintf("  %s %d notes\n",
			m.styles.Label.Render("Total:"),
			len(m.notes)))

		if len(m.notes) > 0 {
			mostRecent := m.notes[0]
			preview := mostRecent.Content
			if len(preview) > 100 {
				preview = preview[:97] + "..."
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n",
				m.styles.Label.Render("Latest:"),
				m.styles.Subtle.Render(preview)))
		}
	}

	// Task Summary
	if len(m.tasks) > 0 {
		sb.WriteString("\n")
		sb.WriteString(m.styles.SectionTitle.Render("✅ Tasks Summary"))
		sb.WriteString("\n\n")

		// Count by state and archived status
		var completed, inProgress, todo, archived int
		for _, task := range m.tasks {
			if task.Archived {
				archived++
			} else {
				switch task.StateID {
				case 3: // done
					completed++
				case 2: // in_progress
					inProgress++
				case 1: // todo
					todo++
				default:
					todo++
				}
			}
		}

		// Display counts
		if todo > 0 {
			sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
				m.styles.Primary.Render("Todo:"),
				todo))
		}
		if inProgress > 0 {
			sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
				m.styles.Warning.Render("In Progress:"),
				inProgress))
		}
		if completed > 0 {
			sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
				m.styles.Success.Render("Done:"),
				completed))
		}
		if archived > 0 {
			sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
				m.styles.Subtle.Render("Archived:"),
				archived))
		}
	}

	// Actions Summary
	if len(m.actions) > 0 {
		sb.WriteString("\n")
		sb.WriteString(m.styles.SectionTitle.Render("⚡ Recent Actions"))
		sb.WriteString("\n\n")

		// Show last 3 actions
		maxActions := 3
		if len(m.actions) < maxActions {
			maxActions = len(m.actions)
		}

		for i := 0; i < maxActions; i++ {
			action := m.actions[i]
			elapsed := time.Since(action.Timestamp)
			actionLine := fmt.Sprintf("  %s %s - %s\n",
				GetActionIcon(action.ActionType),
				m.styles.Subtle.Render(fmt.Sprintf("(%s ago)", FormatDuration(elapsed))),
				action.Description)
			sb.WriteString(actionLine)
		}

		if len(m.actions) > maxActions {
			sb.WriteString(m.styles.Subtle.Render(
				fmt.Sprintf("  ... and %d more actions", len(m.actions)-maxActions)))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// buildTaskSummary builds a task summary string
func buildTaskSummary(tasks []domain.Task, styles Styles) string {
	var sb strings.Builder

	// Count by state
	var completed, inProgress, todo int
	for _, task := range tasks {
		if !task.Archived {
			switch task.StateID {
			case 3: // done
				completed++
			case 2: // in_progress
				inProgress++
			case 1: // todo
				todo++
			default:
				todo++
			}
		}
	}

	// Display counts
	if todo > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			styles.Primary.Render("Todo:"),
			todo))
	}
	if inProgress > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			styles.Warning.Render("In Progress:"),
			inProgress))
	}
	if completed > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			styles.Success.Render("Done:"),
			completed))
	}

	return sb.String()
}

// buildNotesSummary builds a notes summary string
func buildNotesSummary(notes []domain.Note, styles Styles) string {
	var sb strings.Builder

	// Show count and most recent note preview
	sb.WriteString(fmt.Sprintf("  %s %d notes\n",
		styles.Label.Render("Total:"),
		len(notes)))

	if len(notes) > 0 {
		mostRecent := notes[0]
		preview := mostRecent.Content
		if len(preview) > 100 {
			preview = preview[:97] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			styles.Label.Render("Latest:"),
			styles.Subtle.Render(preview)))
	}

	return sb.String()
}
