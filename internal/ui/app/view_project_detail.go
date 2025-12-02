package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// renderProjectDetail renders the project detail view using new modular code
func (m *Model) renderProjectDetail() string {
	log.Printf("[RENDER-PROJECT-DETAIL] Called! projectTabsIndex=%d", m.projectTabsIndex)

	// For now, use a placeholder. We'll implement project selection later
	if m.projectInfo == nil {
		return "No project selected"
	}

	// Try to get the project from the database using the git path
	var selectedProject *database.Project
	dbProject, err := m.data.DB.GetProjectByPath(m.projectInfo.GitRootPath)
	if err != nil {
		// Project not in database yet, use runtime-detected info as fallback
		projectPath := m.projectInfo.GitRootPath
		projectBranch := m.projectInfo.Branch
		projectRemote := m.projectInfo.RemoteURL
		selectedProject = &database.Project{
			Name:      m.projectInfo.RepoName,
			Path:      &projectPath,
			Branch:    &projectBranch,
			RemoteURL: &projectRemote,
		}
	} else {
		selectedProject = dbProject
	}

	// Load active agents for this project using database query
	var agents []domain.Agent
	if selectedProject.ID != "" {
		// Use database query if we have a project ID
		dbAgents, err := m.data.DB.GetActiveAgentsByProject(selectedProject.ID)
		if err == nil && len(dbAgents) > 0 {
			// Convert database.Agent to domain.Agent
			for _, dbAgent := range dbAgents {
				agents = append(agents, convertToDomainAgent(dbAgent))
			}
		}
	}

	// Fallback to filtering from cached agents if no DB results
	if len(agents) == 0 {
		for _, agent := range m.agents {
			// Filter active agents (not completed or failed) for this project
			if agent.ProjectName == selectedProject.Name &&
				agent.Status != "completed" && agent.Status != "failed" {
				agents = append(agents, agent)
			}
		}
	}

	// Load project notes (stub for now - need to implement GetProjectNotes)
	var notes []domain.Note

	// Load project tasks (stub for now - use existing tasks)
	tasks := []domain.Task{}

	// Only recreate table if it doesn't exist, agent count changed, or manual refresh requested
	// This prevents unnecessary table recreation on every render
	shouldRecreateTable := len(agents) > 0 && (
		m.projectAgentsForceRefresh ||
		m.projectAgentsTable.Rows() == nil ||
		len(m.projectAgentsTable.Rows()) != len(agents))

	if shouldRecreateTable {
		reason := "initial"
		if m.projectAgentsForceRefresh {
			reason = "manual refresh"
			m.projectAgentsForceRefresh = false // Clear flag after recreation
		} else if len(m.projectAgentsTable.Rows()) != len(agents) {
			reason = "agent count changed"
		}
		log.Printf("[PROJECT-DETAIL] Recreating agents table (%s: agents=%d, existing_rows=%d)",
			reason, len(agents), len(m.projectAgentsTable.Rows()))
		m.projectAgentsTable = m.createProjectAgentsTable(agents)
	}

	// Render the view directly
	return m.renderProjectDetailView(selectedProject, agents, notes, tasks)
}

// createProjectAgentsTable creates a table.Model for project agents
func (m *Model) createProjectAgentsTable(agents []domain.Agent) table.Model {
	// Calculate available width for table (accounting for split pane and borders)
	// leftWidth is m.width/2, minus panel borders (2 chars each side = 4 total)
	availableWidth := (m.width / 2) - 4

	// Distribute widths proportionally, ensuring minimum widths
	idWidth := max(8, availableWidth*12/82)      // ~15% of total
	statusWidth := max(8, availableWidth*10/82)  // ~12% of total
	descWidth := max(15, availableWidth*35/82)   // ~43% of total
	sessionWidth := max(12, availableWidth*25/82) // ~30% of total

	columns := []table.Column{
		{Title: "ID", Width: idWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Description", Width: descWidth},
		{Title: "Session Name", Width: sessionWidth},
	}

	log.Printf("[TABLE] availableWidth=%d, columns: ID=%d, Status=%d, Desc=%d, Session=%d",
		availableWidth, idWidth, statusWidth, descWidth, sessionWidth)

	// Build session name map
	sessionNames := make(map[string]string)
	for _, agent := range agents {
		if agent.SessionID != "" {
			if session, err := m.data.DB.GetSession(agent.SessionID); err == nil && session.Name != nil {
				sessionNames[agent.SessionID] = *session.Name
			}
		}
	}

	rows := []table.Row{}
	for _, agent := range agents {
		sessionDisplay := ""
		if name, ok := sessionNames[agent.SessionID]; ok && name != "" {
			sessionDisplay = name
			if len(sessionDisplay) > sessionWidth-3 {
				sessionDisplay = sessionDisplay[:sessionWidth-6] + "..."
			}
		} else if agent.SessionID != "" {
			sessionDisplay = agent.SessionID[:min(12, len(agent.SessionID))]
		}

		rows = append(rows, table.Row{
			string(agent.ID)[:min(idWidth, len(string(agent.ID)))],
			agent.Status,
			agent.AgentDescription,
			sessionDisplay,
		})
	}

	// Table height: Use a reasonable default since we'll control via panel height
	// The panel will constrain the table's visible area
	tableHeight := max(5, len(rows)+3) // Rows + header + padding
	if tableHeight > m.height-5 {
		tableHeight = m.height - 5
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Apply styling
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	activeColor := lipgloss.Color(m.theme.Colors.Active)
	s.Selected = s.Selected.Bold(true).Foreground(activeColor)
	t.SetStyles(s)

	// Restore cursor position
	if m.projectAgentsIndex < len(rows) {
		t.SetCursor(m.projectAgentsIndex)
	}

	return t
}

// renderProjectDetailView renders the complete project detail view
func (m *Model) renderProjectDetailView(proj *database.Project, agents []domain.Agent, notes []domain.Note, tasks []domain.Task) string {
	// DEBUG
	log.Printf("[PROJECT-DETAIL-VIEW] renderProjectDetailView called, projectTabsIndex=%d", m.projectTabsIndex)

	// Use NavBar component for tabs
	navbar := components.NewNavBar(
		[]string{"← Back", "[O]verview", "[A]gents", "[N]otes", "[T]asks", "[F]iles"},
		m.theme.Colors.Active,
		m.theme.Colors.Text,
	)
	navbar.Set(m.projectTabsIndex)
	tabsBar := navbar.View()

	log.Printf("[PROJECT-DETAIL-VIEW] tabsBar length: %d chars", len(tabsBar))

	// Calculate available height for content using actual tabs height
	tabsH := lipgloss.Height(tabsBar)
	contentHeight := m.height - tabsH - 2 // Account for \n\n separator
	if contentHeight < 5 {
		contentHeight = 5
	}

	log.Printf("[PROJECT-DETAIL-VIEW] tabsH=%d, m.height=%d, contentHeight=%d", tabsH, m.height, contentHeight)

	// Render content based on active tab
	var content string
	switch m.projectTabsIndex {
	case 0: // Back
		content = "Press esc or q to go back"
	case 1: // Overview
		content = m.renderProjectOverview(proj, agents)
	case 2: // Agents
		content = m.renderProjectAgentsTab(agents, contentHeight)
	case 3: // Notes
		content = "Notes tab (not yet implemented)"
	case 4: // Tasks
		content = "Tasks tab (not yet implemented)"
	case 5: // Files
		content = "Files tab (not yet implemented)"
	default:
		content = ""
	}

	result := tabsBar + "\n\n" + content
	log.Printf("[PROJECT-DETAIL-VIEW] Final result length: %d chars (tabs=%d, content=%d)", len(result), len(tabsBar), len(content))
	log.Printf("[PROJECT-DETAIL-VIEW] First 200 chars: %s", result[:min(200, len(result))])
	return result
}

// renderProjectOverview renders the overview tab
func (m *Model) renderProjectOverview(proj *database.Project, agents []domain.Agent) string {
	var b strings.Builder
	b.WriteString(m.overviewStyles.SectionTitle.Render("Project Details") + "\n\n")

	if proj != nil {
		b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Name:"), proj.Name))
		if proj.Path != nil {
			b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Path:"), *proj.Path))
		}
		if proj.Branch != nil {
			b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Branch:"), *proj.Branch))
		}
		if proj.RemoteURL != nil {
			b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Remote:"), *proj.RemoteURL))
		}
	}

	b.WriteString(fmt.Sprintf("\n%s %d\n", m.overviewStyles.Label.Render("Active Agents:"), len(agents)))

	return b.String()
}

// renderProjectAgentsTab renders the agents tab with split pane
func (m *Model) renderProjectAgentsTab(agents []domain.Agent, contentHeight int) string {
	if len(agents) == 0 {
		return theme.TextMuted.Render("\n  No active agents for this project\n")
	}

	// Calculate split widths
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth - 1

	// Account for panel borders/padding (2 lines for top+bottom border, 2 chars for left+right)
	panelHeight := contentHeight

	log.Printf("[AGENTS-TAB] contentHeight=%d, panelHeight=%d, leftWidth=%d, rightWidth=%d",
		contentHeight, panelHeight, leftWidth, rightWidth)

	// Left pane: agents table
	leftPane := theme.PanelNormal.Copy().
		Width(leftWidth).
		Height(panelHeight).
		Render(m.projectAgentsTable.View())

	// Right pane: selected agent details
	rightPane := m.renderProjectAgentDetails(agents, rightWidth, panelHeight)

	// Join with separator (no border, just vertical line)
	separator := theme.PanelSidebar.Copy().Height(panelHeight).Render("")

	result := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, separator, rightPane)

	// Verify the result doesn't exceed contentHeight
	resultH := lipgloss.Height(result)
	log.Printf("[AGENTS-TAB] result height=%d (should be <=%d)", resultH, contentHeight)

	return result
}

// renderProjectAgentDetails renders the details pane for selected agent
func (m *Model) renderProjectAgentDetails(agents []domain.Agent, width int, contentHeight int) string {
	if m.projectAgentsIndex < 0 || m.projectAgentsIndex >= len(agents) {
		return theme.PanelNoBorder.Copy().
			Width(width).
			Height(contentHeight).
			Render(theme.TextMuted.Render("Select an agent to view details"))
	}

	agent := agents[m.projectAgentsIndex]
	var b strings.Builder

	b.WriteString(m.overviewStyles.SectionTitle.Render("Agent Details") + "\n\n")
	b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("ID:"), string(agent.ID)))
	b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Status:"), agent.Status))
	b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Description:"), agent.AgentDescription))
	b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Session ID:"), agent.SessionID))
	b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Project:"), agent.ProjectName))
	b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Git Path:"), agent.GitWorktreePath))
	b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Current Task:"), agent.CurrentTask))

	if !agent.LastActivityAt.IsZero() {
		b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Last Activity:"), agent.LastActivityAt.Format("2006-01-02 15:04:05")))
	}

	if !agent.CreatedAt.IsZero() {
		b.WriteString(fmt.Sprintf("%s %s\n", m.overviewStyles.Label.Render("Created:"), agent.CreatedAt.Format("2006-01-02 15:04:05")))
	}

	return theme.PanelNoBorder.Copy().
		Width(width).
		Height(contentHeight).
		Render(b.String())
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// convertToDomainAgent converts a database.Agent to a domain.Agent
func convertToDomainAgent(dbAgent *database.Agent) domain.Agent {
	agent := domain.Agent{
		ID:         domain.AgentID(dbAgent.ID),
		Status:     dbAgent.Status,
		Source:     dbAgent.Source,
		CreatedAt:  dbAgent.CreatedAt,
		UpdatedAt:  dbAgent.UpdatedAt,
		Bookmarked: dbAgent.Bookmarked,
	}

	// Convert pointer fields to non-pointer fields (using empty string/zero time as default)
	if dbAgent.GitWorktreePath != nil {
		agent.GitWorktreePath = *dbAgent.GitWorktreePath
	}
	if dbAgent.FeatureDescription != nil {
		agent.FeatureDesc = *dbAgent.FeatureDescription
	}
	if dbAgent.CurrentTask != nil {
		agent.CurrentTask = *dbAgent.CurrentTask
	}
	if dbAgent.LastActivityAt != nil {
		agent.LastActivityAt = *dbAgent.LastActivityAt
	}
	if dbAgent.WindowID != nil {
		agent.WindowID = *dbAgent.WindowID
	}
	if dbAgent.TerminalApplication != nil {
		agent.TerminalApplication = *dbAgent.TerminalApplication
	}
	if dbAgent.Description != nil {
		agent.AgentDescription = *dbAgent.Description
	}
	if dbAgent.ProjectName != nil {
		agent.ProjectName = *dbAgent.ProjectName
	}
	if dbAgent.SessionID != nil {
		agent.SessionID = *dbAgent.SessionID
	}
	if dbAgent.ParentAgentID != nil {
		agent.ParentAgentID = *dbAgent.ParentAgentID
	}

	return agent
}
