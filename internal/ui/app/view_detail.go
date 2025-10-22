package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderDetailView renders the agent detail view
func (m *Model) renderDetailView() string {
	if m.selectedAgent == nil {
		return "No agent selected"
	}

	var view strings.Builder

	// Header
	view.WriteString(m.renderHeader())
	view.WriteString("\n\n")

	// Tabs (with back arrow)
	tabsDisplay := m.tabs.Render()
	view.WriteString(tabsDisplay)
	view.WriteString("\n")

	// Agent info header
	view.WriteString(m.renderAgentInfoHeader())
	view.WriteString("\n")

	// Content based on active tab
	var content string
	switch m.tabs.ActiveIndex {
	case 0: // Back arrow - shouldn't render content
		content = m.styles.Subtle.Render("Press Enter or Esc to go back")
	case 1: // Overview
		content = m.renderAgentDetails()
	case 2: // Commits
		content = m.renderCommitsTab()
	case 3: // Logs
		content = m.renderLogsTab()
	case 4: // Notes
		content = m.renderNotesTab()
	case 5: // Actions
		content = m.renderActionsTab()
	case 6: // Tasks
		content = m.renderTasksTab()
	default:
		content = "Unknown tab"
	}

	// Calculate content height
	contentHeight := m.height - 10 // Account for header, tabs, agent info, footer
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Create scrollable content box
	contentBox := m.styles.ContentBox.
		Width(m.width - 4).
		Height(contentHeight).
		Render(content)
	view.WriteString(contentBox)
	view.WriteString("\n")

	// Footer
	view.WriteString(m.renderFooter())

	return view.String()
}

// renderAgentInfoHeader renders the agent information header in detail view
func (m *Model) renderAgentInfoHeader() string {
	if m.selectedAgent == nil {
		return ""
	}

	agent := *m.selectedAgent

	// Build status with color
	var statusStyle lipgloss.Style
	switch agent.Status {
	case "active":
		statusStyle = m.styles.Success
	case "working":
		statusStyle = m.styles.Warning
	case "idle":
		statusStyle = m.styles.Subtle
	case "failed":
		statusStyle = m.styles.Error
	case "completed":
		statusStyle = m.styles.Primary
	default:
		statusStyle = m.styles.Subtle
	}

	// Format task/feature info
	taskInfo := agent.CurrentTask
	if taskInfo == "" {
		taskInfo = agent.FeatureDesc
	}
	if taskInfo == "" {
		taskInfo = m.styles.Subtle.Render("No task description")
	}

	// Build header content
	header := fmt.Sprintf(" %s %s  %s  %s",
		m.styles.Primary.Bold(true).Render("Agent:"),
		m.styles.Bold.Render(string(agent.ID)),
		statusStyle.Render(fmt.Sprintf("[%s]", agent.Status)),
		taskInfo,
	)

	return m.styles.InfoBox.
		Width(m.width - 4).
		Render(header)
}

// renderAgentDetails renders the overview tab content
func (m *Model) renderAgentDetails() string {
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
		m.getStatusStyle(agent.Status).Render(string(agent.Status))))

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
			m.formatDuration(elapsed))

		var activityStyle lipgloss.Style
		if elapsed < 5*time.Minute {
			activityStyle = m.styles.Success
		} else if elapsed < 30*time.Minute {
			activityStyle = m.styles.Warning
		} else {
			activityStyle = m.styles.Subtle
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
			m.styles.Value.Render(m.formatDuration(duration))))
	} else if agent.CompletedAt != nil {
		duration := agent.CompletedAt.Sub(agent.CreatedAt)
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Total Duration:"),
			m.styles.Value.Render(m.formatDuration(duration))))
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
				m.styles.Subtle.Render(fmt.Sprintf("(%s ago)", m.formatDuration(elapsed))),
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

		// Count by status
		statusCounts := make(map[string]int)
		for _, task := range m.tasks {
			status := task.Status
			if status == "" {
				status = "pending"
			}
			statusCounts[status]++
		}

		// Display counts
		if count, ok := statusCounts["completed"]; ok && count > 0 {
			sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
				m.styles.Success.Render("Completed:"),
				count))
		}
		if count, ok := statusCounts["pending"]; ok && count > 0 {
			sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
				m.styles.Warning.Render("Pending:"),
				count))
		}
		if count, ok := statusCounts["deleted"]; ok && count > 0 {
			sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
				m.styles.Subtle.Render("Deleted:"),
				count))
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
				m.getActionIcon(action.ActionType),
				m.styles.Subtle.Render(fmt.Sprintf("(%s ago)", m.formatDuration(elapsed))),
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

// renderTasksTab renders the tasks tab content
func (m *Model) renderTasksTab() string {
	if len(m.tasks) == 0 {
		return m.styles.Subtle.Render("\n  No tasks found for this agent\n")
	}

	var sb strings.Builder

	// Sort tasks: priority (H>M>L), then by status (pending first), then by creation date
	sortedTasks := make([]domain.Task, len(m.tasks))
	copy(sortedTasks, m.tasks)

	sort.Slice(sortedTasks, func(i, j int) bool {
		// Deleted tasks go to the bottom
		if sortedTasks[i].Status == "deleted" && sortedTasks[j].Status != "deleted" {
			return false
		}
		if sortedTasks[i].Status != "deleted" && sortedTasks[j].Status == "deleted" {
			return true
		}

		// Priority order: H > M > L > (no priority)
		iPri := m.getPriorityWeight(sortedTasks[i].Priority)
		jPri := m.getPriorityWeight(sortedTasks[j].Priority)
		if iPri != jPri {
			return iPri > jPri
		}

		// Then by status (pending before completed)
		if sortedTasks[i].Status != sortedTasks[j].Status {
			if sortedTasks[i].Status == "pending" {
				return true
			}
			if sortedTasks[j].Status == "pending" {
				return false
			}
		}

		// Finally by creation date (newer first)
		return sortedTasks[i].Entry.After(sortedTasks[j].Entry)
	})

	// Header
	sb.WriteString(m.styles.Primary.Render(fmt.Sprintf("%-5s %-8s %-10s %-60s %s\n",
		"ID", "Priority", "Status", "Description", "Created")))
	sb.WriteString(m.styles.Border.Render(strings.Repeat("─", m.width-10)))
	sb.WriteString("\n")

	// Render tasks
	for i, task := range sortedTasks {
		selected := i == m.taskIndex

		// Format priority with color
		var priStyle lipgloss.Style
		var priDisplay string
		switch task.Priority {
		case "H":
			priStyle = m.styles.Error
			priDisplay = "[HIGH]"
		case "M":
			priStyle = m.styles.Warning
			priDisplay = "[MED]"
		case "L":
			priStyle = m.styles.Subtle
			priDisplay = "[LOW]"
		default:
			priStyle = m.styles.Subtle
			priDisplay = "[-]"
		}

		// Format status with color
		var statusStyle lipgloss.Style
		switch task.Status {
		case "completed":
			statusStyle = m.styles.Success
		case "pending":
			statusStyle = m.styles.Warning
		case "deleted":
			statusStyle = m.styles.Subtle.Strikethrough(true)
		default:
			statusStyle = m.styles.Subtle
		}

		// Truncate description if needed
		desc := task.Description
		if len(desc) > 58 {
			desc = desc[:55] + "..."
		}

		// Format line
		line := fmt.Sprintf("%-5s %s %-10s %-60s %s",
			fmt.Sprintf("%d", task.ID),
			priStyle.Render(priDisplay),
			statusStyle.Render(task.Status),
			desc,
			task.Entry.Format("Jan 2 15:04"),
		)

		if selected {
			sb.WriteString(m.styles.Selected.Render(line))
		} else {
			sb.WriteString(line)
		}
		sb.WriteString("\n")
	}

	// Show task details if one is selected
	if m.taskIndex >= 0 && m.taskIndex < len(sortedTasks) {
		task := sortedTasks[m.taskIndex]
		sb.WriteString("\n")
		sb.WriteString(m.styles.Border.Render(strings.Repeat("─", m.width-10)))
		sb.WriteString("\n")
		sb.WriteString(m.renderTaskDetails(task))
	}

	return sb.String()
}

// renderTaskDetails renders detailed view of a single task
func (m *Model) renderTaskDetails(task domain.Task) string {
	var sb strings.Builder

	sb.WriteString(m.styles.SectionTitle.Render("Task Details"))
	sb.WriteString("\n\n")

	// Task UUID
	if task.UUID != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("UUID:"),
			m.styles.Value.Render(task.UUID)))
	}

	// Tags
	if len(task.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("  %s %s\n",
			m.styles.Label.Render("Tags:"),
			m.styles.Value.Render(strings.Join(task.Tags, ", "))))
	}

	// Annotations
	if len(task.Annotations) > 0 {
		sb.WriteString("\n")
		sb.WriteString(m.styles.Label.Render("Annotations:"))
		sb.WriteString("\n")
		for _, ann := range task.Annotations {
			sb.WriteString(fmt.Sprintf("  • %s\n", ann.Description))
		}
	}

	// Virtual tags (computed)
	if len(task.VirtualTags) > 0 {
		sb.WriteString(fmt.Sprintf("\n  %s %s\n",
			m.styles.Label.Render("Virtual Tags:"),
			m.styles.Subtle.Render(strings.Join(task.VirtualTags, ", "))))
	}

	return sb.String()
}

// renderCommitsTab renders the commits tab content with split-pane view
func (m *Model) renderCommitsTab() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Build item list for left pane
	items := make([]string, len(m.commits))
	for i, commit := range m.commits {
		date := commit.Timestamp.Format("2006-01-02")
		hash := truncateCommitHash(string(commit.Hash), 8)
		subject := truncate(commit.Message, 40)
		items[i] = fmt.Sprintf("%s  %s  %s", date, hash, subject)
	}

	// Build detail content for right pane
	var detailContent string
	if m.commitsIndex >= 0 && m.commitsIndex < len(m.commits) {
		detailContent = m.renderCommitDetails(m.commits[m.commitsIndex])
	} else if len(m.commits) == 0 {
		return m.styles.Subtle.Render("No commits found")
	}

	return m.renderSplitPaneContent(items, m.commitsIndex, detailContent)
}

// renderCommitDetails renders details for a single commit
func (m *Model) renderCommitDetails(commit domain.Commit) string {
	var b strings.Builder

	// Commit hash
	b.WriteString(m.styles.Git.Render("commit "))
	b.WriteString(m.styles.Value.Render(string(commit.Hash)))
	b.WriteString("\n")

	// Author info (if available)
	if commit.Author != "" {
		b.WriteString(m.styles.Label.Render("Author: "))
		b.WriteString(m.styles.Value.Render(commit.Author))
		b.WriteString("\n")
	}

	// Date
	b.WriteString(m.styles.Label.Render("Date:   "))
	b.WriteString(m.styles.Value.Render(commit.Timestamp.Format("Mon Jan 2 15:04:05 2006 MST")))
	b.WriteString("\n\n")

	// Message
	b.WriteString(m.styles.Text.Render(commit.Message))
	b.WriteString("\n")

	// Files changed (if available)
	if len(commit.FilesChanged) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.SectionTitle.Render("Files Changed"))
		b.WriteString("\n\n")
		for _, file := range commit.FilesChanged {
			b.WriteString(fmt.Sprintf("  %s\n", file))
		}
	}

	// Diff preview (if available)
	if commit.Diff != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.SectionTitle.Render("Diff Preview"))
		b.WriteString("\n\n")
		// Show first 20 lines of diff
		lines := strings.Split(commit.Diff, "\n")
		maxLines := 20
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		for i := 0; i < maxLines; i++ {
			line := lines[i]
			if strings.HasPrefix(line, "+") {
				b.WriteString(m.styles.Success.Render(line))
			} else if strings.HasPrefix(line, "-") {
				b.WriteString(m.styles.Error.Render(line))
			} else {
				b.WriteString(m.styles.Subtle.Render(line))
			}
			b.WriteString("\n")
		}
		if len(lines) > maxLines {
			b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("... +%d more lines", len(lines)-maxLines)))
		}
	}

	return b.String()
}

// renderLogsTab renders the logs tab content with split-pane view
func (m *Model) renderLogsTab() string {
	if len(m.logs) == 0 {
		return m.styles.Subtle.Render("No logs found")
	}

	// Build item list for left pane
	items := make([]string, len(m.logs))
	for i, log := range m.logs {
		timestamp := log.Timestamp.Format("15:04:05")
		logType := truncate(log.Type, 10)
		items[i] = fmt.Sprintf("%s  %-10s", timestamp, logType)
	}

	// Build detail content for right pane
	var detailContent string
	if m.logsIndex >= 0 && m.logsIndex < len(m.logs) {
		detailContent = m.renderLogDetails(m.logs[m.logsIndex])
	}

	return m.renderSplitPaneContent(items, m.logsIndex, detailContent)
}

// renderLogDetails renders details for a single log entry
func (m *Model) renderLogDetails(log domain.Log) string {
	var b strings.Builder

	// Log header
	b.WriteString(m.styles.Primary.Render("Type: "))
	b.WriteString(m.styles.Value.Render(log.Type))
	b.WriteString("\n")

	b.WriteString(m.styles.Primary.Render("Time: "))
	b.WriteString(m.styles.Value.Render(log.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Content (with markdown rendering if needed)
	if strings.Contains(log.Message, "```") || strings.Contains(log.Message, "#") {
		// Try to render as markdown
		if rendered, err := m.renderMarkdown(log.Message); err == nil {
			b.WriteString(rendered)
		} else {
			b.WriteString(m.styles.Text.Render(log.Message))
		}
	} else {
		b.WriteString(m.styles.Text.Render(log.Message))
	}

	return b.String()
}

// renderNotesTab renders the notes tab content with split-pane view
func (m *Model) renderNotesTab() string {
	if len(m.notes) == 0 {
		return m.styles.Subtle.Render("No notes available")
	}

	// Build item list for left pane
	items := make([]string, len(m.notes))
	for i, note := range m.notes {
		timestamp := note.CreatedAt.Format("15:04:05")
		preview := truncate(note.Content, 30)
		items[i] = fmt.Sprintf("%s  %s", timestamp, preview)
	}

	// Build detail content for right pane
	var detailContent string
	if m.notesIndex >= 0 && m.notesIndex < len(m.notes) {
		detailContent = m.renderNoteDetails(m.notes[m.notesIndex])
	}

	return m.renderSplitPaneContent(items, m.notesIndex, detailContent)
}

// renderNoteDetails renders details for a single note
func (m *Model) renderNoteDetails(note domain.Note) string {
	var b strings.Builder

	// Note header
	b.WriteString(m.styles.Primary.Render("Created: "))
	b.WriteString(m.styles.Value.Render(note.CreatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Content
	b.WriteString(m.styles.Text.Render(note.Content))

	return b.String()
}

// renderActionsTab renders the actions tab content with split-pane view
func (m *Model) renderActionsTab() string {
	if len(m.actions) == 0 {
		return m.styles.Subtle.Render("No actions recorded")
	}

	// Build item list for left pane
	items := make([]string, len(m.actions))
	for i, action := range m.actions {
		timestamp := action.Timestamp.Format("15:04:05")
		actionType := truncate(action.ActionType, 15)
		items[i] = fmt.Sprintf("%s  %-15s", timestamp, actionType)
	}

	// Build detail content for right pane
	var detailContent string
	if m.actionsIndex >= 0 && m.actionsIndex < len(m.actions) {
		detailContent = m.renderActionDetails(m.actions[m.actionsIndex])
	}

	return m.renderSplitPaneContent(items, m.actionsIndex, detailContent)
}

// renderActionDetails renders the full action details
func (m *Model) renderActionDetails(action domain.Action) string {
	var b strings.Builder

	// Action type
	b.WriteString(m.styles.Primary.Render("Action: "))
	b.WriteString(m.styles.Text.Render(action.ActionType))
	b.WriteString("\n")

	// Timestamp
	b.WriteString(m.styles.Primary.Render("Time: "))
	b.WriteString(m.styles.Text.Render(action.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Description
	b.WriteString(m.styles.Title.Render("Description"))
	b.WriteString("\n")
	b.WriteString(m.styles.Text.Render(action.Description))
	b.WriteString("\n\n")

	// Details (if available)
	if action.Details != "" {
		b.WriteString(m.styles.Title.Render("Details"))
		b.WriteString("\n")
		b.WriteString(m.styles.Text.Render(action.Details))
		b.WriteString("\n")
	}

	return b.String()
}

// Helper methods for detail view
func (m *Model) getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "active":
		return m.styles.Success
	case "working":
		return m.styles.Warning
	case "idle":
		return m.styles.Subtle
	case "failed":
		return m.styles.Error
	case "completed":
		return m.styles.Primary
	default:
		return m.styles.Subtle
	}
}

func (m *Model) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

func (m *Model) getActionIcon(actionType string) string {
	switch actionType {
	case "task_start":
		return "▶️"
	case "file_operation":
		return "📁"
	case "git_commit":
		return "📝"
	case "status_update":
		return "📊"
	case "log":
		return "📋"
	case "note":
		return "📌"
	default:
		return "⚡"
	}
}

func (m *Model) getPriorityWeight(priority string) int {
	switch priority {
	case "H":
		return 3
	case "M":
		return 2
	case "L":
		return 1
	default:
		return 0
	}
}

// renderSplitPaneContent renders a split-pane layout with list on left and details on right
func (m *Model) renderSplitPaneContent(items []string, selectedIndex int, detailContent string) string {
	// Calculate dimensions
	totalWidth := m.width - 6
	leftPaneWidth := totalWidth / 3
	// rightPaneWidth := totalWidth - leftPaneWidth - 1 // -1 for separator (unused)

	// Build left pane (list)
	var leftPane strings.Builder
	for i, item := range items {
		if i == selectedIndex {
			leftPane.WriteString(m.styles.Selected.Render(truncate(item, leftPaneWidth-2)))
		} else {
			leftPane.WriteString(truncate(item, leftPaneWidth-2))
		}
		if i < len(items)-1 {
			leftPane.WriteString("\n")
		}
	}

	// Build right pane (details)
	rightPane := detailContent
	if rightPane == "" {
		rightPane = m.styles.Subtle.Render("Select an item to view details")
	}

	// Combine panes side by side
	leftLines := strings.Split(leftPane.String(), "\n")
	rightLines := strings.Split(rightPane, "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var combined strings.Builder
	for i := 0; i < maxLines; i++ {
		// Left pane line
		if i < len(leftLines) {
			line := leftLines[i]
			combined.WriteString(line)
			// Pad to fill width
			padding := leftPaneWidth - lipgloss.Width(line)
			if padding > 0 {
				combined.WriteString(strings.Repeat(" ", padding))
			}
		} else {
			combined.WriteString(strings.Repeat(" ", leftPaneWidth))
		}

		// Separator
		combined.WriteString(m.styles.Border.Render("│"))

		// Right pane line
		if i < len(rightLines) {
			combined.WriteString(" " + rightLines[i])
		}

		if i < maxLines-1 {
			combined.WriteString("\n")
		}
	}

	return combined.String()
}

// renderMarkdown attempts to render markdown content with styling
func (m *Model) renderMarkdown(content string) (string, error) {
	// Simple markdown rendering
	lines := strings.Split(content, "\n")
	var rendered strings.Builder

	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			rendered.WriteString(m.styles.Code.Render(line))
		} else if strings.HasPrefix(line, "# ") {
			rendered.WriteString(m.styles.Title.Render(strings.TrimPrefix(line, "# ")))
		} else if strings.HasPrefix(line, "## ") {
			rendered.WriteString(m.styles.SectionTitle.Render(strings.TrimPrefix(line, "## ")))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			rendered.WriteString("  • " + strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		} else {
			rendered.WriteString(line)
		}
		rendered.WriteString("\n")
	}

	return rendered.String(), nil
}