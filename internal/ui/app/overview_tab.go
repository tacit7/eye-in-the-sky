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

	sb.WriteString(renderLabelValue("Agent ID:", string(agent.ID), m.styles.Value, m.styles.Label))
	sb.WriteString(renderLabelValue("Status:", agent.Status, GetStatusStyle(agent.Status, m.styles), m.styles.Label))

	if agent.FeatureDesc != "" {
		sb.WriteString(renderLabelValue("Description:", agent.FeatureDesc, m.styles.Value, m.styles.Label))
	}
	if agent.ProjectName != "" {
		sb.WriteString(renderLabelValue("Project:", agent.ProjectName, m.styles.Value, m.styles.Label))
	}
	if agent.CurrentTask != "" {
		sb.WriteString(renderLabelValue("Current Task:", agent.CurrentTask, m.styles.Value, m.styles.Label))
	}

	sb.WriteString(renderLabelValue("Source:", agent.Source, m.styles.Value, m.styles.Label))

	if agent.GitWorktreePath != "" {
		sb.WriteString(renderLabelValue("Worktree:", agent.GitWorktreePath, m.styles.Value, m.styles.Label))
	}
	if agent.SessionID != "" {
		sb.WriteString(renderLabelValue("Session ID:", agent.SessionID, m.styles.Value, m.styles.Label))
	}
	if agent.ParentSessionID != "" {
		sb.WriteString(renderLabelValue("Parent Session:", agent.ParentSessionID, m.styles.Value, m.styles.Label))
	}
	if agent.ParentAgentID != "" {
		sb.WriteString(renderLabelValue("Parent Agent:", string(agent.ParentAgentID), m.styles.Value, m.styles.Label))
	}
	if agent.WindowID != "" {
		sb.WriteString(renderLabelValue("Window ID:", agent.WindowID, m.styles.Value, m.styles.Label))
	}

	return sb.String()
}

// renderTiming renders timing information section
func (m *Model) renderTiming(agent domain.Agent) string {
	var sb strings.Builder
	sb.WriteString(m.styles.SectionTitle.Render("⏱️ Timing"))
	sb.WriteString("\n\n")

	sb.WriteString(renderLabelValue("Created:", agent.CreatedAt.Format("Jan 2, 2006 15:04:05 MST"), m.styles.Value, m.styles.Label))
	sb.WriteString(renderLabelValue("Updated:", agent.UpdatedAt.Format("Jan 2, 2006 15:04:05 MST"), m.styles.Value, m.styles.Label))

	if !agent.LastActivityAt.IsZero() {
		elapsed := time.Since(agent.LastActivityAt)
		activityStr := fmt.Sprintf("%s (%s ago)",
			agent.LastActivityAt.Format("15:04:05"),
			FormatDuration(elapsed))
		style := m.getActivityStyle(elapsed)
		sb.WriteString(renderLabelValue("Last Activity:", activityStr, style, m.styles.Label))
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

	maxCommits := minInt(len(m.commits), 5)
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

	counts := countTasksByState(m.tasks)

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

	maxActions := minInt(len(m.actions), 3)
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
		return renderLabelValue("Session Duration:", FormatDuration(duration), m.styles.Value, m.styles.Label)
	} else if agent.CompletedAt != nil {
		duration := agent.CompletedAt.Sub(agent.CreatedAt)
		return renderLabelValue("Total Duration:", FormatDuration(duration), m.styles.Value, m.styles.Label)
	}
	return ""
}
