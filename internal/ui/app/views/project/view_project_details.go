package project

import (
	"log"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/project/tabs"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// View renders the project details view
func (m *Model) View() string {
	if m.ctx == nil || m.ctx.Project == nil {
		return "No project selected"
	}

	// DEBUG: Log tab state
	log.Printf("[PROJECT-VIEW] tabs.ActiveIndex=%d", m.tabs.ActiveIndex)

	// Render tabs bar using theme styles
	tabNames := []string{"← Back", "[O]verview", "[A]gents", "[N]otes", "[T]asks", "[F]iles"}
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
	case tabAgents:
		content = m.renderAgentsTab()
	case tabNotes:
		content = m.renderNotesTab()
	case tabTasks:
		content = m.renderTasksTab()
	case tabFiles:
		content = m.renderFilesTab()
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

// renderAgentsTab renders the agents tab with table.Model
func (m *Model) renderAgentsTab() string {
	return tabs.RenderAgents(m.ctx, m.styles.OverviewStyles, m.agentsTable, m.width, m.height)
}

// renderNotesTab renders the notes tab with table.Model
func (m *Model) renderNotesTab() string {
	return tabs.RenderNotes(m.ctx, m.styles.OverviewStyles, m.notesTable, m.width, m.height)
}

// renderTasksTab renders the tasks tab with table.Model
func (m *Model) renderTasksTab() string {
	return tabs.RenderTasks(m.ctx, m.styles.OverviewStyles, m.tasksTable, m.width, m.height)
}

// renderFilesTab renders the files tab
func (m *Model) renderFilesTab() string {
	return tabs.RenderFiles(m.ctx, m.styles.OverviewStyles)
}
