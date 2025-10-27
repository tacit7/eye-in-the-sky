package agent_details

import (
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/agent_details/tabs"
)

// View renders the agent details view
func (m *Model) View() string {
	if m.ctx == nil || m.ctx.Agent == nil {
		return "No agent selected"
	}

	// TODO: Add header + tabs bar rendering

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
	case tabProjects:
		content = m.renderProjectsTab()
	default:
		content = ""
	}

	// TODO: Add footer

	return content
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
