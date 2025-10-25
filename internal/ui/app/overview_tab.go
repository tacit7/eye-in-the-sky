package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderOverviewTab renders the overview tab content
func (m *Model) renderOverviewTab() string {
	if m.selectedAgent == nil {
		return "No agent selected"
	}

	agent := *m.selectedAgent
	sections := []string{
		m.renderAgentInfo(agent),
		m.renderTiming(agent),
		m.renderCommitsSection(),
		m.renderNotesSection(),
		m.renderTasksSection(),
		m.renderActionsSection(),
	}

	return strings.Join(filterNonEmpty(sections), "\n")
}

// renderAgentInfo renders agent identification information
func (m *Model) renderAgentInfo(agent domain.Agent) string {
	var sb strings.Builder
	sb.WriteString(m.styles.SectionTitle.Render("📋 Agent Information"))
	sb.WriteString("\n\n")

	sb.WriteString(m.renderField("Agent ID:", string(agent.ID), m.styles.Value))
	sb.WriteString(m.renderField("Status:", agent.Status, GetStatusStyle(agent.Status, m.styles)))

	if agent.FeatureDesc != "" {
		sb.WriteString(m.renderField("Description:", agent.FeatureDesc, m.styles.Value))
	}
	if agent.ProjectName != "" {
		sb.WriteString(m.renderField("Project:", agent.ProjectName, m.styles.Value))
	}
	if agent.CurrentTask != "" {
		sb.WriteString(m.renderField("Current Task:", agent.CurrentTask, m.styles.Value))
	}

	sb.WriteString(m.renderField("Source:", agent.Source, m.styles.Value))

	if agent.GitWorktreePath != "" {
		sb.WriteString(m.renderField("Worktree:", agent.GitWorktreePath, m.styles.Value))
	}
	if agent.SessionID != "" {
		sb.WriteString(m.renderField("Session ID:", agent.SessionID, m.styles.Value))
	}
	if agent.ParentSessionID != "" {
		sb.WriteString(m.renderField("Parent Session:", agent.ParentSessionID, m.styles.Value))
	}
	if agent.ParentAgentID != "" {
		sb.WriteString(m.renderField("Parent Agent:", string(agent.ParentAgentID), m.styles.Value))
	}
	if agent.WindowID != "" {
		sb.WriteString(m.renderField("Window ID:", agent.WindowID, m.styles.Value))
	}

	return sb.String()
}

// renderTiming renders timing information section
func (m *Model) renderTiming(agent domain.Agent) string {
	var sb strings.Builder
	sb.WriteString(m.styles.SectionTitle.Render("⏱️ Timing"))
	sb.WriteString("\n\n")

	sb.WriteString(m.renderField("Created:", agent.CreatedAt.Format("Jan 2, 2006 15:04:05 MST"), m.styles.Value))
	sb.WriteString(m.renderField("Updated:", agent.UpdatedAt.Format("Jan 2, 2006 15:04:05 MST"), m.styles.Value))

	if !agent.LastActivityAt.IsZero() {
		elapsed := time.Since(agent.LastActivityAt)
		activityStr := fmt.Sprintf("%s (%s ago)",
			agent.LastActivityAt.Format("15:04:05"),
			FormatDuration(elapsed))
		style := m.getActivityStyle(elapsed)
		sb.WriteString(m.renderField("Last Activity:", activityStr, style))
	}

	sb.WriteString(m.renderDuration(agent))

	return sb.String()
}

// renderCommitsSection renders recent commits section
func (m *Model) renderCommitsSection() string {
	if len(m.commits) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(m.styles.SectionTitle.Render("📝 Recent Commits"))
	sb.WriteString("\n\n")

	maxCommits := min(len(m.commits), 5)
	for i := 0; i < maxCommits; i++ {
		commit := m.commits[i]
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

	return sb.String()
}

// renderNotesSection renders notes summary section
func (m *Model) renderNotesSection() string {
	if len(m.notes) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(m.styles.SectionTitle.Render("📌 Notes Summary"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("  %s %d notes\n",
		m.styles.Label.Render("Total:"),
		len(m.notes)))

	mostRecent := m.notes[0]
	preview := mostRecent.Content
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		m.styles.Label.Render("Latest:"),
		m.styles.Subtle.Render(preview)))

	return sb.String()
}

// renderTasksSection renders task summary section
func (m *Model) renderTasksSection() string {
	if len(m.tasks) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(m.styles.SectionTitle.Render("✅ Tasks Summary"))
	sb.WriteString("\n\n")

	counts := m.countTasksByState()

	if counts["todo"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			m.styles.Primary.Render("Todo:"),
			counts["todo"]))
	}
	if counts["inProgress"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			m.styles.Warning.Render("In Progress:"),
			counts["inProgress"]))
	}
	if counts["completed"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			m.styles.Success.Render("Done:"),
			counts["completed"]))
	}
	if counts["archived"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			m.styles.Subtle.Render("Archived:"),
			counts["archived"]))
	}

	return sb.String()
}

// renderActionsSection renders recent actions section
func (m *Model) renderActionsSection() string {
	if len(m.actions) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(m.styles.SectionTitle.Render("⚡ Recent Actions"))
	sb.WriteString("\n\n")

	maxActions := min(len(m.actions), 3)
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

	return sb.String()
}

// renderField renders a label-value pair
func (m *Model) renderField(label, value string, style lipgloss.Style) string {
	return fmt.Sprintf("  %s %s\n",
		m.styles.Label.Render(label),
		style.Render(value))
}

// getActivityStyle returns appropriate style based on elapsed time
func (m *Model) getActivityStyle(elapsed time.Duration) lipgloss.Style {
	if elapsed < 5*time.Minute {
		return m.styles.Success
	} else if elapsed < 30*time.Minute {
		return m.styles.Warning
	}
	return m.styles.Subtle
}

// renderDuration renders session or total duration
func (m *Model) renderDuration(agent domain.Agent) string {
	if agent.Status != "completed" && agent.Status != "failed" {
		duration := time.Since(agent.CreatedAt)
		return m.renderField("Session Duration:", FormatDuration(duration), m.styles.Value)
	} else if agent.CompletedAt != nil {
		duration := agent.CompletedAt.Sub(agent.CreatedAt)
		return m.renderField("Total Duration:", FormatDuration(duration), m.styles.Value)
	}
	return ""
}

// countTasksByState returns task counts by state
func (m *Model) countTasksByState() map[string]int {
	counts := map[string]int{
		"todo":       0,
		"inProgress": 0,
		"completed":  0,
		"archived":   0,
	}

	for _, task := range m.tasks {
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

// filterNonEmpty removes empty strings from slice
func filterNonEmpty(sections []string) []string {
	var result []string
	for _, s := range sections {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
