package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
)

// formatNumber formats an integer with comma separators
func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

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
	default:
		return "Unknown view"
	}
}

// renderListView renders the agent list view
func (m *Model) renderListView() string {
	// Check which tab is active
	var contentBuilder strings.Builder

	// Reserve space for header (3 lines) + title bar + footer + borders
	visibleHeight := m.height - 8
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	// Calculate and store layout for click hit testing
	m.listLayout = ListLayout{
		HeaderH: 3,                    // renderHeader uses ~3 lines
		TabsH:   2,                    // tabs box uses ~2 lines
		FooterH: 2,                    // renderFooter uses ~2 lines
		Content: Bounds{X: 2, Y: 5, W: m.width - 4, H: visibleHeight + 2},
		RowH:    1, // one line per agent
	}

	// Render content based on active tab
	switch m.listTabs.ActiveIndex {
	case 1: // Project tab
		contentBuilder.WriteString(m.renderProjectTab())
	case 3: // Usage tab
		contentBuilder.WriteString(m.renderUsageTab())
	default: // Overview, Claude tabs
		if len(m.agents) == 0 {
			contentBuilder.WriteString(m.styles.Subtle.Render("  No agents found"))
			contentBuilder.WriteString("\n")
		} else {
			// Add header row
			headerLine := fmt.Sprintf("%-2s %-12s  %-40s  %-40s  %-30s  %s",
				"",
				"Status",
				"Agent ID",
				"Task",
				"Source",
				"Session",
			)
			contentBuilder.WriteString(m.styles.Primary.Render(headerLine))
			contentBuilder.WriteString("\n")
			contentBuilder.WriteString(m.styles.Border.Render(strings.Repeat("─", m.width-6)))
			contentBuilder.WriteString("\n")

			endIndex := m.listOffset + visibleHeight
			if endIndex > len(m.agents) {
				endIndex = len(m.agents)
			}

			for i := m.listOffset; i < endIndex; i++ {
				agent := m.agents[i]
				line := m.renderAgentLine(agent, i == m.selectedIndex)
				contentBuilder.WriteString(line)
				contentBuilder.WriteString("\n")
			}
		}
	}

	// Create bordered content box
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Width(m.width - 4).
		Height(visibleHeight + 2).
		Padding(0, 1).
		Render(contentBuilder.String())

	// Assemble full view
	var b strings.Builder

	// Top header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Tabs with bottom border
	tabsBox := lipgloss.NewStyle().
		BorderBottom(true).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Width(m.width).
		Render(m.listTabs.View())
	b.WriteString(tabsBox)
	b.WriteString("\n")

	// Bordered content
	b.WriteString(contentBox)

	// Footer
	b.WriteString("\n")
	footer := m.renderFooter()
	b.WriteString(footer)

	return b.String()
}

// renderDetailView renders the agent detail view with tabs
func (m *Model) renderDetailView() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Build detail content based on active tab
	var detailContent string
	switch m.tabs.ActiveIndex {
	case 0: // Back arrow - should not be shown, handled by update logic
		detailContent = m.renderAgentDetails()
	case 1: // Agent View
		detailContent = m.renderAgentDetails()
	case 2: // Commits
		detailContent = m.renderCommitsTab()
	case 3: // Logs
		detailContent = m.renderLogsTab()
	case 4: // Notes
		detailContent = m.renderNotesTab()
	case 5: // Actions
		detailContent = m.renderActionsTab()
	case 6: // Tasks
		detailContent = m.renderTasksTab()
	default:
		detailContent = m.renderAgentDetails()
	}

	// Reserve space for header, agent info header, tabs, footer, and borders
	visibleHeight := m.height - 13
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	// Create bordered content box
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Width(m.width - 4).
		Height(visibleHeight + 2).
		Padding(0, 1).
		Render(detailContent)

	// Assemble full view
	var b strings.Builder

	// Top header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Agent info header
	agentHeader := m.renderAgentInfoHeader()
	b.WriteString(agentHeader)
	b.WriteString("\n")

	// Tabs with bottom border
	tabsBox := lipgloss.NewStyle().
		BorderBottom(true).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Width(m.width).
		Render(m.tabs.View())
	b.WriteString(tabsBox)
	b.WriteString("\n")

	// Bordered content
	b.WriteString(contentBox)

	// Footer
	b.WriteString("\n")
	footer := m.renderFooter()
	b.WriteString(footer)

	return b.String()
}

// renderLogsView renders the logs view with split panes
func (m *Model) renderLogsView() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
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

	// Footer
	footer := m.renderFooterWithKeys("[j/k] Move  [r] Refresh  [q/esc] Back")

	return m.renderSplitPaneView(
		"Logs",
		items,
		m.logsIndex,
		len(m.logs),
		detailContent,
		footer,
	)
}

// renderLogDetails renders the full log message with markdown syntax highlighting
func (m *Model) renderLogDetails(log Log) string {
	var b strings.Builder

	// Timestamp
	b.WriteString(m.styles.Primary.Render("Time: "))
	b.WriteString(m.styles.Text.Render(log.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Type
	b.WriteString(m.styles.Primary.Render("Type: "))
	typeStyle := m.styles.Text
	switch log.Type {
	case "error":
		typeStyle = m.styles.Failed
	case "warning":
		typeStyle = m.styles.Idle
	case "info":
		typeStyle = m.styles.Active
	}
	b.WriteString(typeStyle.Render(strings.ToUpper(log.Type)))
	b.WriteString("\n\n")

	// Message with markdown rendering
	b.WriteString(m.styles.Title.Render("Message"))
	b.WriteString("\n")

	// Try to render as markdown, fall back to plain text if it fails
	rendered, err := m.renderMarkdown(log.Message)
	if err != nil {
		b.WriteString(m.styles.Text.Render(log.Message))
	} else {
		b.WriteString(rendered)
	}
	b.WriteString("\n")

	return b.String()
}

// renderMarkdown renders markdown content with syntax highlighting
func (m *Model) renderMarkdown(content string) (string, error) {
	// If renderer is not available, return error to fall back to plain text
	if m.mdRenderer == nil {
		return "", fmt.Errorf("markdown renderer not available")
	}

	// Render the markdown using cached renderer
	rendered, err := m.mdRenderer.Render(content)
	if err != nil {
		return "", err
	}

	return rendered, nil
}

// renderAgentInfoHeader renders the agent information header (project, ID, description)
func (m *Model) renderAgentInfoHeader() string {
	if m.selectedAgent == nil {
		return ""
	}

	agent := m.selectedAgent
	var parts []string

	// Project name
	if agent.ProjectName != "" {
		projectText := m.styles.Primary.Render("Project: ") + m.styles.Text.Render(agent.ProjectName)
		parts = append(parts, projectText)
	}

	// Agent ID
	idText := m.styles.Primary.Render("Agent: ") + m.styles.Active.Render(agent.ID)
	parts = append(parts, idText)

	// Agent description
	if agent.AgentDescription != "" {
		descText := m.styles.Subtle.Render(agent.AgentDescription)
		parts = append(parts, descText)
	}

	// Join parts with separator
	headerContent := strings.Join(parts, " │ ")

	// Create bordered header box
	headerBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Width(m.width - 2).
		Padding(0, 1).
		Render(headerContent)

	return headerBox
}

// renderHeader renders the application header
func (m *Model) renderHeader() string {
	// App name and icon
	appName := m.styles.Title.Render("■ eye-in-the-sky")

	// Current path/view indicator
	var viewPath string
	switch m.currentView {
	case ViewList:
		viewPath = "/agents"
	case ViewDetail:
		if m.selectedAgent != nil {
			viewPath = fmt.Sprintf("/agents/%s", truncateID(m.selectedAgent.ID, 8))
		} else {
			viewPath = "/agents/detail"
		}
	}

	pathStyle := m.styles.Subtle.Render(viewPath)

	// Top line: app name + path
	topLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		appName,
		"  ",
		pathStyle,
	)

	// Stats bar
	filterStatus := "Active"
	if m.showAll {
		filterStatus = "All"
	}

	stats := fmt.Sprintf("Agents: %d │ Filter: %s │ Last refresh: %s",
		len(m.agents),
		filterStatus,
		m.lastRefresh.Format("15:04:05"),
	)

	statsLine := m.styles.Subtle.Render(stats)

	// Status message or error (if present)
	var statusLine string
	if m.err != nil {
		statusLine = m.styles.Failed.Render("✗ Error: " + m.err.Error())
	} else if m.statusMsg != "" {
		statusLine = m.styles.Active.Render("● " + m.statusMsg)
	}

	// Create bordered header box
	var headerContent string
	if statusLine != "" {
		headerContent = lipgloss.JoinVertical(
			lipgloss.Left,
			topLine,
			statsLine,
			statusLine,
		)
	} else {
		headerContent = lipgloss.JoinVertical(
			lipgloss.Left,
			topLine,
			statsLine,
		)
	}

	headerBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Width(m.width - 2).
		Padding(0, 1).
		Render(headerContent)

	return headerBox
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
			"[c] Continue",
			"[s] Start",
			"[n] New",
		}
	case ViewDetail:
		keys = []string{
			"[q/esc] Back",
			"[o/c/l/n/a] Tabs",
			"[j/k] Navigate",
			"[h/l] Scroll",
			"[ctrl+u/d] Page",
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
	// Subagent indicator - green fullwidth vertical for child agents
	var prefix string
	var status string
	statusStyle := m.getStatusStyle(agent.Status, agent.LastActivityAt)

	if agent.ParentAgentID != "" {
		// Subagent: fullwidth vertical, no circle
		prefix = m.styles.Active.Render("｜") + " "
		status = strings.ToUpper(agent.Status)
	} else {
		// Parent agent: normal prefix, with circle
		prefix = "  "
		statusIndicator := m.getStatusIndicator(agent.Status)
		status = statusIndicator + " " + strings.ToUpper(agent.Status)
	}

	statusText := statusStyle.Render(status)

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

	// Session ID or fallback (only first 8 chars)
	var sessionIDDisplay string
	fullSessionID := agent.CurrentSessionID
	if fullSessionID == "" {
		sessionIDDisplay = "—"
	} else {
		if len(fullSessionID) > 8 {
			sessionIDDisplay = fullSessionID[:8]
		} else {
			sessionIDDisplay = fullSessionID
		}
		sessionIDDisplay = "📋 " + sessionIDDisplay
	}

	// Combine into line with prefix - session ID on the right
	line := fmt.Sprintf("%s%-12s  %-40s  %-40s  %-30s  %s",
		prefix,
		statusText,
		truncate(info, 40),
		truncate(task, 40),
		truncate(source, 30),
		sessionIDDisplay,
	)

	if selected {
		return m.styles.Primary.Reverse(true).Render(line)
	}
	return m.styles.Text.Render(line)
}

// renderAgentDetails renders the agent view (detailed information about an agent)
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

	if agent.CurrentSessionID != "" {
		b.WriteString(m.styles.Primary.Render("Session ID: "))
		b.WriteString(m.styles.Text.Render(agent.CurrentSessionID))
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

	// Session costs summary
	if len(m.sessionMetrics) > 0 {
		latest := m.sessionMetrics[0]
		b.WriteString("\n")
		b.WriteString(m.styles.Title.Render("Current Session Costs"))
		b.WriteString("\n")

		// Token usage
		usagePercent := float64(latest.TokensUsed) / float64(latest.TokensBudget) * 100
		b.WriteString(m.styles.Primary.Render("  Tokens: "))
		tokenInfo := fmt.Sprintf("%d / %d (%.1f%%)", latest.TokensUsed, latest.TokensBudget, usagePercent)
		b.WriteString(m.styles.Text.Render(tokenInfo))
		b.WriteString("\n")

		// Cost estimate
		if latest.EstimatedCostUSD > 0 {
			b.WriteString(m.styles.Primary.Render("  Cost: "))
			costInfo := fmt.Sprintf("$%.4f", latest.EstimatedCostUSD)
			b.WriteString(m.styles.Text.Render(costInfo))
			b.WriteString("\n")
		}

		// Model name
		if latest.ModelName != "" {
			b.WriteString(m.styles.Primary.Render("  Model: "))
			b.WriteString(m.styles.Text.Render(latest.ModelName))
			b.WriteString("\n")
		}

		// Last updated
		b.WriteString(m.styles.Primary.Render("  Updated: "))
		b.WriteString(m.styles.Subtle.Render(formatTimestamp(latest.Timestamp)))
		b.WriteString("\n")
	}

	// Recent commits (last 5 only)
	if len(m.commits) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Title.Render("Recent Commits (last 5)"))
		b.WriteString("\n")

		maxCommits := 5
		if len(m.commits) < maxCommits {
			maxCommits = len(m.commits)
		}

		for i := 0; i < maxCommits; i++ {
			commit := m.commits[i]
			line := fmt.Sprintf("  %s  %s  %s",
				commit.Timestamp.Format("15:04:05"),
				truncateCommitHash(commit.CommitHash, 8),
				truncate(commit.CommitMessage, 60),
			)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}

		if len(m.commits) > 5 {
			b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("  ... %d more (see Commits tab)", len(m.commits)-5)))
			b.WriteString("\n")
		}
	}

	// Session notes (first line only, last 5)
	if len(m.notes) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Title.Render("Session Notes (last 5)"))
		b.WriteString("\n")

		maxNotes := 5
		if len(m.notes) < maxNotes {
			maxNotes = len(m.notes)
		}

		for i := 0; i < maxNotes; i++ {
			note := m.notes[i]
			// Get first line only
			firstLine := getFirstLine(note.Content)
			line := fmt.Sprintf("  %s  %s",
				note.Timestamp.Format("15:04:05"),
				truncate(firstLine, 80),
			)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}

		if len(m.notes) > 5 {
			b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("  ... %d more (see Notes tab)", len(m.notes)-5)))
			b.WriteString("\n")
		}
	}

	// Recent actions (last 5 only)
	if len(m.actions) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.Title.Render("Recent Actions (last 5)"))
		b.WriteString("\n")

		maxActions := 5
		if len(m.actions) < maxActions {
			maxActions = len(m.actions)
		}

		for i := 0; i < maxActions; i++ {
			action := m.actions[i]
			line := fmt.Sprintf("  %s  %-15s  %s",
				action.Timestamp.Format("15:04:05"),
				truncate(action.ActionType, 15),
				truncate(action.Description, 50),
			)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}

		if len(m.actions) > 5 {
			b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("  ... %d more (see Actions tab)", len(m.actions)-5)))
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

// truncateCommitHash safely truncates a commit hash to the specified length
func truncateCommitHash(hash string, maxLen int) string {
	if len(hash) <= maxLen {
		return hash
	}
	return hash[:maxLen]
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

// getFirstLine extracts the first line from multi-line text
func getFirstLine(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return text
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

// renderCommitsTab renders the commits tab content with split-pane view
func (m *Model) renderCommitsTab() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Build item list for left pane
	items := make([]string, len(m.commits))
	for i, commit := range m.commits {
		date := commit.Timestamp.Format("2006-01-02")
		hash := truncateCommitHash(commit.CommitHash, 8)
		subject := truncate(commit.CommitMessage, 40)
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

// renderNotesTab renders the notes tab content with split-pane view
func (m *Model) renderNotesTab() string {
	if len(m.notes) == 0 {
		return m.styles.Subtle.Render("No notes available")
	}

	// Build item list for left pane
	items := make([]string, len(m.notes))
	for i, note := range m.notes {
		timestamp := note.Timestamp.Format("15:04:05")
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
func (m *Model) renderActionDetails(action Action) string {
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

// renderTasksTab renders the tasks tab content with split-pane view
func (m *Model) renderTasksTab() string {
	if len(m.tasks) == 0 {
		return m.styles.Subtle.Render("No tasks available for this agent")
	}

	// Build item list for left pane
	items := make([]string, len(m.tasks))
	for i, task := range m.tasks {
		status := task.Status
		if len(status) > 10 {
			status = status[:10]
		}
		items[i] = fmt.Sprintf("[%-10s] %s", status, truncate(task.Description, 50))
	}

	// Build detail content for right pane
	var detailContent string
	if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
		detailContent = m.renderTaskDetails(m.tasks[m.tasksIndex])
	}

	return m.renderSplitPaneContent(items, m.tasksIndex, detailContent)
}

// renderProjectTab renders the detected project information with sections
func (m *Model) renderProjectTab() string {
	var b strings.Builder

	if m.projectInfo == nil {
		b.WriteString(m.styles.Subtle.Render("  Not in a git repository"))
		b.WriteString("\n")
		return b.String()
	}

	// Project header
	b.WriteString(m.styles.Title.Render("Project: "))
	if m.projectInfo.Owner != "" && m.projectInfo.RepoName != "" {
		b.WriteString(m.styles.Text.Render(fmt.Sprintf("%s/%s", m.projectInfo.Owner, m.projectInfo.RepoName)))
	} else if m.projectInfo.RepoName != "" {
		b.WriteString(m.styles.Text.Render(m.projectInfo.RepoName))
	} else {
		b.WriteString(m.styles.Text.Render("(unknown)"))
	}
	b.WriteString("\n")

	// Project metadata
	if m.projectInfo.Branch != "" {
		b.WriteString(m.styles.Primary.Render("  Branch: "))
		b.WriteString(m.styles.Text.Render(m.projectInfo.Branch))
		b.WriteString("  ")
	}
	if m.projectInfo.Commit != "" {
		b.WriteString(m.styles.Primary.Render("Commit: "))
		b.WriteString(m.styles.Text.Render(m.projectInfo.Commit))
	}
	b.WriteString("\n")

	// Root path
	if m.projectInfo.GitRootPath != "" {
		b.WriteString(m.styles.Subtle.Render("  " + truncate(m.projectInfo.GitRootPath, 70)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Section navigation
	b.WriteString(m.renderProjectSectionNav())
	b.WriteString("\n")

	// Active section content
	switch m.projectSelectedSection {
	case 0: // TaskWarrior tasks
		b.WriteString(m.renderProjectTasks())
	case 1: // CLAUDE.md
		b.WriteString(m.renderProjectClaudeMD())
	case 2: // .md files
		b.WriteString(m.renderProjectMarkdownFiles())
	}

	return b.String()
}

// renderProjectSectionNav renders the section navigation indicator
func (m *Model) renderProjectSectionNav() string {
	sections := []string{"[1] Tasks", "[2] CLAUDE.md", "[3] Markdown"}
	var nav string

	for i, sec := range sections {
		if i == m.projectSelectedSection {
			nav += m.styles.Primary.Reverse(true).Render(" " + sec + " ")
		} else {
			nav += m.styles.Text.Render(" " + sec + " ")
		}
	}

	return "  " + nav
}

// renderProjectTasks renders the open TaskWarrior tasks section
func (m *Model) renderProjectTasks() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Open Tasks"))
	b.WriteString("\n")

	if len(m.projectTasks) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No open tasks"))
		b.WriteString("\n")
		return b.String()
	}

	for i, task := range m.projectTasks {
		selected := i == m.projectTasksIndex
		line := fmt.Sprintf("  %s  %s", task.Priority, truncate(task.Description, 60))

		if selected {
			b.WriteString(m.styles.Primary.Reverse(true).Render(line))
		} else {
			b.WriteString(m.styles.Text.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderProjectClaudeMD renders the CLAUDE.md content
func (m *Model) renderProjectClaudeMD() string {
	var b strings.Builder

	if m.claudeMDContent == "" {
		b.WriteString(m.styles.Subtle.Render("  CLAUDE.md not found"))
		b.WriteString("\n")
		return b.String()
	}

	b.WriteString(m.styles.Title.Render("CLAUDE.md"))
	b.WriteString("\n")

	// Render first ~20 lines
	lines := strings.Split(m.claudeMDContent, "\n")
	maxLines := 20
	if len(lines) < maxLines {
		maxLines = len(lines)
	}

	for i := 0; i < maxLines; i++ {
		line := lines[i]
		if len(line) > 70 {
			line = truncate(line, 70)
		}
		b.WriteString(m.styles.Text.Render(line))
		b.WriteString("\n")
	}

	if len(lines) > maxLines {
		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("  ... (%d more lines)", len(lines)-maxLines)))
		b.WriteString("\n")
	}

	return b.String()
}

// renderProjectMarkdownFiles renders the list of markdown files
func (m *Model) renderProjectMarkdownFiles() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Markdown Files"))
	b.WriteString("\n")

	if len(m.projectMDFiles) == 0 {
		b.WriteString(m.styles.Subtle.Render("  No markdown files found"))
		b.WriteString("\n")
		return b.String()
	}

	for i, file := range m.projectMDFiles {
		selected := i == m.projectMDFilesIndex
		line := fmt.Sprintf("  %s", file.Name)

		if selected {
			b.WriteString(m.styles.Primary.Reverse(true).Render(line))
			b.WriteString("\n")

			// Show preview of selected file
			preview := strings.Split(file.Content, "\n")
			previewLines := 5
			if len(preview) < previewLines {
				previewLines = len(preview)
			}

			for j := 0; j < previewLines; j++ {
				pline := preview[j]
				if len(pline) > 68 {
					pline = truncate(pline, 68)
				}
				b.WriteString(m.styles.Subtle.Render("    " + pline))
				b.WriteString("\n")
			}

			if len(preview) > previewLines {
				b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("    ... (%d more lines)", len(preview)-previewLines)))
				b.WriteString("\n")
			}
		} else {
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderUsageTab renders session costs for all agents with monthly breakdown
func (m *Model) renderUsageTab() string {
	var b strings.Builder

	// Show Claude Code empty state if database is empty
	if m.ccusageDB != nil && m.ccusageEntryCount == 0 {
		b.WriteString(m.styles.Title.Render("Claude Code Usage"))
		b.WriteString("\n\n")

		if m.ccusageSyncing {
			b.WriteString(m.styles.Working.Render("  ⏳ " + m.ccusageSyncStatus))
		} else {
			b.WriteString(m.styles.Subtle.Render("  No Claude Code usage data found"))
			b.WriteString("\n\n")
			b.WriteString(m.styles.Primary.Render("  Press 'I' to initialize database"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  This will discover and parse JSONL files from:"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  ~/.config/claude/projects/"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  ~/.claude/projects/"))
		}

		b.WriteString("\n\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n\n")
	}

	if len(m.allSessionMetrics) == 0 && len(m.monthlyCosts) == 0 && m.ccusageEntryCount == 0 {
		return m.styles.Subtle.Render("  No usage data available")
	}

	// Current Session Summary
	if len(m.allSessionMetrics) > 0 {
		b.WriteString(m.styles.Title.Render("Eye-in-the-Sky Session Metrics"))
		b.WriteString("\n\n")

		// Column headers
		headerLine := fmt.Sprintf("%-10s %-8s %-12s %-10s %-12s",
			"Agent", "Usage %", "Tokens", "Cost USD", "Model")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		// Data rows
		totalCost := 0.0
		totalTokens := 0

		for _, metric := range m.allSessionMetrics {
			usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

			agentID := truncate(metric.AgentID, 8)
			usage := fmt.Sprintf("%.1f%%", usagePercent)
			tokens := fmt.Sprintf("%d", metric.TokensUsed)
			cost := fmt.Sprintf("$%.4f", metric.EstimatedCostUSD)
			model := truncate(metric.ModelName, 12)

			line := fmt.Sprintf("%-10s %-8s %-12s %-10s %-12s",
				agentID, usage, tokens, cost, model)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")

			totalCost += metric.EstimatedCostUSD
			totalTokens += metric.TokensUsed
		}

		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Total: $%.4f | %d tokens\n", totalCost, totalTokens)))
	}

	// Monthly Cost Breakdown
	if len(m.monthlyCosts) > 0 {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Monthly Cost Breakdown"))
		b.WriteString("\n\n")

		// Column headers
		headerLine := fmt.Sprintf("%-20s %-8s %-12s %-10s %-8s",
			"Timestamp", "Usage %", "Tokens Used", "Cost USD", "Model")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		// Data rows
		totalMonthlyCost := 0.0
		totalMonthlyTokens := 0

		for _, metric := range m.monthlyCosts {
			usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

			timestamp := metric.Timestamp.Format("2006-01-02 15:04")
			usage := fmt.Sprintf("%.1f%%", usagePercent)
			tokens := fmt.Sprintf("%d", metric.TokensUsed)
			cost := fmt.Sprintf("$%.4f", metric.EstimatedCostUSD)
			model := truncate(metric.ModelName, 8)

			line := fmt.Sprintf("%-20s %-8s %-12s %-10s %-8s",
				timestamp, usage, tokens, cost, model)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")

			totalMonthlyCost += metric.EstimatedCostUSD
			totalMonthlyTokens += metric.TokensUsed
		}

		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Monthly Total: $%.4f | %d tokens across %d snapshots\n",
			totalMonthlyCost, totalMonthlyTokens, len(m.monthlyCosts))))
	}

	// Claude Code metrics (if available)
	if m.ccusageDB != nil && m.ccusageEntryCount > 0 {
		b.WriteString(m.renderClaudeCodeMetrics())
		b.WriteString("\n\n")
		b.WriteString(m.renderCombinedSummary())
	}

	return b.String()
}

// renderClaudeCodeMetrics renders Claude Code usage data
func (m *Model) renderClaudeCodeMetrics() string {
	var b strings.Builder

	// If no data, try to sync on-demand
	if m.ccusageEntryCount == 0 && m.ccusageDB != nil {
		b.WriteString(m.styles.Subtle.Render("No data available\n\n"))
		b.WriteString(m.styles.Text.Render("Press 'I' to sync and load Claude Code usage data from your local projects.\n"))
		b.WriteString(m.styles.Subtle.Render("This will scan ~/.config/claude/projects/ and ~/.claude/projects/ for usage logs.\n"))
		return b.String()
	}

	// Daily Usage Section
	if len(m.ccusageDaily) > 0 {
		b.WriteString(m.styles.Title.Render("Claude Code Daily Usage"))
		b.WriteString("\n\n")

		// Aggregate by date only
		dailyTotals := make(map[string]*api.DailyReport)
		for _, report := range m.ccusageDaily {
			if daily, exists := dailyTotals[report.Date]; exists {
				daily.InputTokens += report.InputTokens
				daily.OutputTokens += report.OutputTokens
				daily.CacheCreationTokens += report.CacheCreationTokens
				daily.CacheReadTokens += report.CacheReadTokens
				daily.TotalCost += report.TotalCost
			} else {
				newReport := report
				dailyTotals[report.Date] = &newReport
			}
		}

		// Sort dates in reverse order (most recent first)
		dates := make([]string, 0, len(dailyTotals))
		for date := range dailyTotals {
			dates = append(dates, date)
		}
		sort.Strings(dates)
		sort.Sort(sort.Reverse(sort.StringSlice(dates)))

		// Table header with borders
		tableWidth := 110
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", tableWidth)))
		b.WriteString("\n")

		headerLine := fmt.Sprintf("│ %-12s │ %-14s │ %-14s │ %-14s │ %-14s │ %-16s │ %-12s │",
			"Date", "Input", "Output", "CacheWr", "CacheRd", "Total Tokens", "Cost USD")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", tableWidth)))
		b.WriteString("\n")

		totalCost := 0.0
		for i, date := range dates {
			report := dailyTotals[date]
			totalTokens := report.InputTokens + report.OutputTokens + report.CacheCreationTokens + report.CacheReadTokens
			line := fmt.Sprintf("│ %-12s │ %-14s │ %-14s │ %-14s │ %-14s │ %-16s │ %-12s │",
				date,
				formatNumber(report.InputTokens),
				formatNumber(report.OutputTokens),
				formatNumber(report.CacheCreationTokens),
				formatNumber(report.CacheReadTokens),
				formatNumber(totalTokens),
				fmt.Sprintf("$%.4f", report.TotalCost))

			// Alternate row colors for better readability (only within table width)
			if i%2 == 0 {
				b.WriteString(m.styles.Text.Render(line))
			} else {
				// Darker background for alternating rows - only applies to text width
				altStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("250")).
					Background(lipgloss.Color("233")).
					Padding(0, 0)
				b.WriteString(altStyle.Render(line))
			}
			b.WriteString("\n")
			totalCost += report.TotalCost
		}

		// Table bottom border
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", tableWidth)))
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Total Cost: $%.4f\n", totalCost)))
	}

	// Monthly Summary Section
	if len(m.ccusageMonthly) > 0 {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Claude Code Monthly Summary"))
		b.WriteString("\n\n")

		// Monthly stats table
		monthlyTableWidth := 95
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", monthlyTableWidth)))
		b.WriteString("\n")

		monthHeader := fmt.Sprintf("│ %-12s │ %-12s │ %-16s │ %-16s │ %-16s │ %-12s │",
			"Month", "Days", "Input Tokens", "Output Tokens", "Total Tokens", "Cost USD")
		b.WriteString(m.styles.Primary.Render(monthHeader))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", monthlyTableWidth)))
		b.WriteString("\n")

		// Display each month's data
		for i, monthly := range m.ccusageMonthly {
			totalTokensMonth := monthly.TotalInputTokens + monthly.TotalOutputTokens + monthly.TotalCacheTokens

			monthLine := fmt.Sprintf("│ %-12s │ %-12d │ %-16s │ %-16s │ %-16s │ %-12s │",
				monthly.Month,
				monthly.Days,
				formatNumber(monthly.TotalInputTokens),
				formatNumber(monthly.TotalOutputTokens),
				formatNumber(totalTokensMonth),
				fmt.Sprintf("$%.4f", monthly.TotalCost))

			// Alternate row colors for better readability
			if i%2 == 0 {
				b.WriteString(m.styles.Text.Render(monthLine))
			} else {
				altStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("250")).
					Background(lipgloss.Color("233")).
					Padding(0, 0)
				b.WriteString(altStyle.Render(monthLine))
			}
			b.WriteString("\n")
		}

		b.WriteString(m.styles.Border.Render(strings.Repeat("─", monthlyTableWidth)))
		b.WriteString("\n")
	}

	// Active Block Section
	if m.ccusageBlock != nil {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Current Billing Block"))
		b.WriteString("\n\n")

		blockInfo := fmt.Sprintf("Block: %s to %s | Time Remaining: %s",
			m.ccusageBlock.StartTime,
			m.ccusageBlock.EndTime,
			m.ccusageBlock.TimeRemaining)
		b.WriteString(m.styles.Text.Render(blockInfo))
		b.WriteString("\n")

		blockUsage := fmt.Sprintf("Usage: %s input | %s output | $%.4f total",
			formatNumber(m.ccusageBlock.InputTokens),
			formatNumber(m.ccusageBlock.OutputTokens),
			m.ccusageBlock.TotalCost)
		b.WriteString(m.styles.Text.Render(blockUsage))
		b.WriteString("\n")
	}

	return b.String()
}

// renderCombinedSummary renders Claude Code usage summary
func (m *Model) renderCombinedSummary() string {
	var b strings.Builder

	if len(m.ccusageMonthly) == 0 {
		return ""
	}

	b.WriteString(m.styles.Title.Render("Total All-Time Cost"))
	b.WriteString("\n\n")

	// Calculate totals across all months
	totalCost := 0.0
	for _, monthly := range m.ccusageMonthly {
		totalCost += monthly.TotalCost
	}

	line := fmt.Sprintf("Claude Code Usage: $%.4f", totalCost)
	b.WriteString(m.styles.Primary.Render(line))
	b.WriteString("\n")

	return b.String()
}
