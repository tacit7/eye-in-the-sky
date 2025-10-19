package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current view
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render help overlay if shown
	if m.showHelp {
		return m.renderHelp()
	}

	switch m.currentView {
	case ViewList:
		return m.renderListView()
	case ViewDetail:
		return m.renderDetailView()
	case ViewLogs:
		return m.renderLogsView()
	default:
		return "Unknown view"
	}
}

// renderListView renders the agent list view
func (m *Model) renderListView() string {
	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n\n")

	// Agent list
	visibleHeight := m.height - 5 // Reserve for header and footer
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	if len(m.agents) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No agents found"))
		b.WriteString("\n")
	} else {
		endIndex := m.listOffset + visibleHeight
		if endIndex > len(m.agents) {
			endIndex = len(m.agents)
		}

		for i := m.listOffset; i < endIndex; i++ {
			agent := m.agents[i]
			line := m.renderAgentLine(agent, i == m.selectedIndex)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Footer
	footer := m.renderFooter()
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

// renderDetailView renders the agent detail view
func (m *Model) renderDetailView() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n\n")

	// Agent details
	b.WriteString(m.renderAgentDetails())

	// Footer
	footer := m.renderFooter()
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

// renderLogsView renders the logs view
func (m *Model) renderLogsView() string {
	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n\n")

	// Logs content
	b.WriteString(m.styles.Title.Render("All Actions Log"))
	b.WriteString("\n\n")

	if len(m.actions) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No actions found"))
	} else {
		visibleHeight := m.height - 8
		if visibleHeight < 1 {
			visibleHeight = 10
		}

		endIndex := m.detailOffset + visibleHeight
		if endIndex > len(m.actions) {
			endIndex = len(m.actions)
		}

		for i := m.detailOffset; i < endIndex; i++ {
			action := m.actions[i]
			line := fmt.Sprintf("  %s | %s | %s",
				action.Timestamp.Format("15:04:05"),
				action.ActionType,
				action.Description,
			)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}
	}

	// Footer
	footer := m.renderFooter()
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

// renderHeader renders the application header
func (m *Model) renderHeader() string {
	title := "Eye in the Sky - Agent Dashboard"

	filterStatus := "Active"
	if m.showAll {
		filterStatus = "All"
	}

	info := fmt.Sprintf("Agents: %d | Filter: %s | Last refresh: %s",
		len(m.agents),
		filterStatus,
		m.lastRefresh.Format("15:04:05"),
	)

	titleStyle := m.styles.Title.Width(m.width)
	infoStyle := m.styles.Subtle.Width(m.width)

	return titleStyle.Render(title) + "\n" + infoStyle.Render(info)
}

// renderFooter renders the application footer with key bindings
func (m *Model) renderFooter() string {
	var keys []string

	switch m.currentView {
	case ViewList:
		keys = []string{
			"[q] Quit",
			"[r] Refresh",
			"[a] Toggle Filter",
			"[j/k] Navigate",
			"[enter] Details",
			"[L] Logs",
		}
	case ViewDetail, ViewLogs:
		keys = []string{
			"[q/esc] Back",
			"[j/k] Scroll",
			"[ctrl+u/d] Page",
			"[g/G] Top/Bottom",
		}
	}

	help := strings.Join(keys, "  ")

	footerStyle := m.styles.Subtle.
		Width(m.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border))

	return footerStyle.Render(help)
}

// renderAgentLine renders a single agent line in the list
func (m *Model) renderAgentLine(agent Agent, selected bool) string {
	// Status indicator
	status := m.getStatusIndicator(agent.Status)
	statusStyle := m.getStatusStyle(agent.Status, agent.LastActivityAt)
	statusText := statusStyle.Render(status + " " + strings.ToUpper(agent.Status))

	// Agent info
	info := fmt.Sprintf("%s", agent.ID)
	if agent.AgentDescription != "" {
		info = fmt.Sprintf("%s - %s", agent.ID, agent.AgentDescription)
	}

	// Current task
	task := agent.CurrentTask
	if task == "" {
		task = agent.FeatureDesc
	}
	if task == "" {
		task = "No task"
	}

	// Project/source
	source := agent.ProjectName
	if source == "" && agent.GitWorktreePath != "" {
		source = agent.GitWorktreePath
	}
	if source == "" {
		source = agent.Source
	}

	// Combine into line
	line := fmt.Sprintf("  %-12s  %-40s  %-40s  %s",
		statusText,
		truncate(info, 40),
		truncate(task, 40),
		truncate(source, 30),
	)

	if selected {
		return m.styles.Primary.Reverse(true).Render(line)
	}
	return m.styles.Text.Render(line)
}

// renderAgentDetails renders detailed information about an agent
func (m *Model) renderAgentDetails() string {
	if m.selectedAgent == nil {
		return ""
	}

	var b strings.Builder
	agent := m.selectedAgent

	// Agent header
	b.WriteString(m.styles.Title.Render(fmt.Sprintf("Agent: %s", agent.ID)))
	b.WriteString("\n\n")

	// Agent info
	b.WriteString(m.styles.Primary.Render("Status: "))
	statusStyle := m.getStatusStyle(agent.Status, agent.LastActivityAt)
	b.WriteString(statusStyle.Render(strings.ToUpper(agent.Status)))
	b.WriteString("\n")

	if agent.AgentDescription != "" {
		b.WriteString(m.styles.Primary.Render("Description: "))
		b.WriteString(m.styles.Text.Render(agent.AgentDescription))
		b.WriteString("\n")
	}

	if agent.ProjectName != "" {
		b.WriteString(m.styles.Primary.Render("Project: "))
		b.WriteString(m.styles.Text.Render(agent.ProjectName))
		b.WriteString("\n")
	}

	if agent.CurrentTask != "" {
		b.WriteString(m.styles.Primary.Render("Current Task: "))
		b.WriteString(m.styles.Text.Render(agent.CurrentTask))
		b.WriteString("\n")
	}

	if agent.FeatureDesc != "" {
		b.WriteString(m.styles.Primary.Render("Feature: "))
		b.WriteString(m.styles.Text.Render(agent.FeatureDesc))
		b.WriteString("\n")
	}

	if agent.GitWorktreePath != "" {
		b.WriteString(m.styles.Primary.Render("Worktree: "))
		b.WriteString(m.styles.Text.Render(agent.GitWorktreePath))
		b.WriteString("\n")
	}

	b.WriteString(m.styles.Primary.Render("Created: "))
	b.WriteString(m.styles.Text.Render(agent.CreatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString("\n")

	if !agent.LastActivityAt.IsZero() {
		b.WriteString(m.styles.Primary.Render("Last Activity: "))
		b.WriteString(m.styles.Text.Render(formatTimestamp(agent.LastActivityAt)))
		b.WriteString("\n")
	}

	// Recent commits
	if len(m.commits) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Title.Render("Recent Commits"))
		b.WriteString("\n")
		for _, commit := range m.commits {
			line := fmt.Sprintf("  %s  %s  %s",
				commit.Timestamp.Format("15:04:05"),
				commit.CommitHash[:8],
				truncate(commit.CommitMessage, 60),
			)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}
	}

	// Recent actions
	if len(m.actions) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Title.Render("Recent Actions"))
		b.WriteString("\n")

		visibleHeight := m.height - 25
		if visibleHeight < 5 {
			visibleHeight = 5
		}

		endIndex := m.detailOffset + visibleHeight
		if endIndex > len(m.actions) {
			endIndex = len(m.actions)
		}

		for i := m.detailOffset; i < endIndex; i++ {
			action := m.actions[i]
			line := fmt.Sprintf("  %s  %-15s  %s",
				action.Timestamp.Format("15:04:05"),
				action.ActionType,
				action.Description,
			)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}

		if endIndex < len(m.actions) {
			remaining := len(m.actions) - endIndex
			b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("  ... %d more actions", remaining)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// getStatusIndicator returns the status indicator symbol
func (m *Model) getStatusIndicator(status string) string {
	switch status {
	case "active":
		return "●"
	case "working":
		return "●"
	case "idle":
		return "●"
	case "completed":
		return "✓"
	case "failed":
		return "✗"
	case "stale":
		return "●"
	default:
		return "?"
	}
}

// getStatusStyle returns the style for a given status
func (m *Model) getStatusStyle(status string, lastActivity time.Time) lipgloss.Style {
	// Check for stale status
	if !lastActivity.IsZero() {
		inactiveDuration := time.Since(lastActivity)
		if inactiveDuration > time.Hour {
			return m.styles.Unknown
		} else if inactiveDuration > 30*time.Minute {
			return m.styles.Stale
		}
	}

	switch status {
	case "active":
		return m.styles.Active
	case "working":
		return m.styles.Working
	case "idle":
		return m.styles.Idle
	case "completed":
		return m.styles.Completed
	case "failed":
		return m.styles.Failed
	default:
		return m.styles.Unknown
	}
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatTimestamp formats a timestamp in a human-readable way
func formatTimestamp(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d min ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		return t.Format("2006-01-02 15:04")
	}
}

// renderHelp renders the help overlay
func (m *Model) renderHelp() string {
	configDir, _ := ConfigDir()
	
	title := m.styles.Title.Render("Eye in the Sky - Help")
	
	// Get help content from help model
	helpView := m.help.View(m.keys)
	
	// Footer with config info
	configPath := fmt.Sprintf("Config: %s", configDir)
	version := "Version: 0.1.0"
	footer := m.styles.Subtle.Render(fmt.Sprintf("%s | %s", configPath, version))
	
	// Create bordered box
	helpContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		helpView,
		"",
		footer,
	)
	
	// Style the help box
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Padding(1, 2).
		Width(m.width - 4).
		Render(helpContent)
	
	// Center the help box
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		helpBox,
	)
}
