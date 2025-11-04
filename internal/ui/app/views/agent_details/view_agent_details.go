package agent_details

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/agent_details/tabs"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// View renders the agent details view
func (m *Model) View() string {
	if m.ctx == nil || m.ctx.Agent == nil {
		return "No agent selected"
	}

	// Render tabs bar using theme styles
	tabNames := []string{"← Back", "[O]verview", "[T]asks", "[A]ctions", "[L]ogs", "[C]ommits", "[N]otes", "[S]ession Context"}
	var renderedTabs []string
	for i, name := range tabNames {
		if i == m.tabs.ActiveIndex {
			renderedTabs = append(renderedTabs, theme.TabActive.Render(name))
		} else {
			renderedTabs = append(renderedTabs, theme.TabInactive.Render(name))
		}
	}
	tabsBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Content based on active tab - explicit switch, no registry
	var content string
	switch m.tabs.ActiveIndex {
	case tabBack:
		content = "← Back"
	case tabOverview:
		content = m.renderOverviewTab()
	case tabTasks:
		content = m.renderTasksTab()
	case tabActions:
		content = m.renderActionsTab()
	case tabLogs:
		content = m.renderLogsTab()
	case tabCommits:
		content = m.renderCommitsTab()
	case tabNotes:
		content = m.renderNotesTab()
	case tabSessionContext:
		content = m.renderSessionContextTab()
	default:
		content = ""
	}

	// Combine tabs and content
	return tabsBar + "\n\n" + content
}

// renderOverviewTab renders the overview tab with caching
func (m *Model) renderOverviewTab() string {
	if !m.overviewDirty && m.overviewCache != "" {
		return m.overviewCache
	}

	m.overviewCache = tabs.RenderOverview(m.ctx, m.styles.OverviewStyles)
	m.overviewDirty = false
	return m.overviewCache
}

// renderTasksTab renders the tasks tab
func (m *Model) renderTasksTab() string {
	return tabs.RenderTasks(m.ctx, m.styles.OverviewStyles)
}

// renderActionsTab renders the actions tab
func (m *Model) renderActionsTab() string {
	return tabs.RenderActions(m.ctx, m.styles.OverviewStyles)
}

// renderLogsTab renders the logs tab
func (m *Model) renderLogsTab() string {
	return tabs.RenderLogs(m.ctx, m.styles.OverviewStyles)
}

// renderCommitsTab renders the commits tab
func (m *Model) renderCommitsTab() string {
	return tabs.RenderCommits(m.ctx, m.styles.OverviewStyles)
}

// renderNotesTab renders the notes tab
func (m *Model) renderNotesTab() string {
	return tabs.RenderNotes(m.ctx, m.styles.OverviewStyles)
}

// renderProjectsTab renders the projects tab
func (m *Model) renderProjectsTab() string {
	return tabs.RenderProjects(m.ctx, m.styles.OverviewStyles)
}

// renderSessionContextTab renders the session context tab
func (m *Model) renderSessionContextTab() string {
	return tabs.RenderSessionContext(m.ctx, m.styles.OverviewStyles)
}
