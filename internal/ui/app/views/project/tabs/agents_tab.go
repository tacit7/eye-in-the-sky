package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// RenderAgents renders the agents tab with split-pane: table (left) + details (right)
func RenderAgents(ctx *DataContext, styles OverviewStyles, agentsTable table.Model, width, height int) string {
	if ctx == nil || len(ctx.Agents) == 0 {
		return theme.TextMuted.Render("\n  No active agents for this project\n")
	}

	// Calculate split widths
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	// Left pane: table.Model view
	leftPane := renderAgentsTable(agentsTable, leftWidth, height)

	// Right pane: selected agent details
	rightPane := renderAgentDetails(ctx, rightWidth, height, styles)

	// Join panes horizontally with separator
	result := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		theme.PanelSidebar.Copy().Height(height).Render(""),
		rightPane,
	)

	return result
}

// renderAgentsTable renders the left pane table
func renderAgentsTable(t table.Model, width, height int) string {
	return theme.PanelNormal.Copy().
		Width(width).
		Height(height).
		Render(t.View())
}

// renderAgentDetails renders the right pane with selected agent info
func renderAgentDetails(ctx *DataContext, width, height int, styles OverviewStyles) string {
	if ctx.SelectedAgentIndex < 0 || ctx.SelectedAgentIndex >= len(ctx.Agents) {
		return theme.PanelNoBorder.Copy().
			Width(width).
			Height(height).
			Render(theme.TextMuted.Render("Select an agent to view details"))
	}

	agent := ctx.Agents[ctx.SelectedAgentIndex]

	var details strings.Builder

	// Header
	details.WriteString(styles.SectionTitle.Render("Agent Details") + "\n\n")

	// Agent info
	details.WriteString(renderField("ID", string(agent.ID), styles))
	details.WriteString(renderField("Status", agent.Status, styles))
	details.WriteString(renderField("Description", agent.AgentDescription, styles))
	details.WriteString(renderField("Session ID", agent.SessionID, styles))
	details.WriteString(renderField("Project", agent.ProjectName, styles))
	details.WriteString(renderField("Git Path", agent.GitWorktreePath, styles))
	details.WriteString(renderField("Current Task", agent.CurrentTask, styles))

	if !agent.LastActivityAt.IsZero() {
		details.WriteString(renderField("Last Activity", agent.LastActivityAt.Format("2006-01-02 15:04:05"), styles))
	}

	if !agent.CreatedAt.IsZero() {
		details.WriteString(renderField("Created", agent.CreatedAt.Format("2006-01-02 15:04:05"), styles))
	}

	return theme.PanelNoBorder.Copy().
		Width(width).
		Height(height).
		Render(details.String())
}

// renderField renders a label-value pair
func renderField(label, value string, styles OverviewStyles) string {
	if value == "" {
		value = theme.TextMuted.Render("(none)")
	}
	return fmt.Sprintf("%s %s\n", styles.Label.Render(label+":"), styles.Value.Render(value))
}
