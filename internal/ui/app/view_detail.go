package app

import (
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/agent_details"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/agent_details/tabs"
)

// renderDetail renders the agent detail view using new modular code
func (m *Model) renderDetail() string {
	if m.selectedAgent == nil {
		return "No agent selected"
	}

	// Create context for agent details
	ctx := &agent_details.DataContext{
		Agent:             m.selectedAgent,
		Commits:           m.commits,
		Notes:             m.notes,
		Tasks:             m.tasks,
		TaskNotes:         []domain.TaskNote{}, // TODO: Load based on selected task
		Actions:           m.actions,
		Logs:              m.logs,
		SelectedTaskIndex: m.tasksIndex,
		Width:             m.width,
	}

	// Create styles wrapper - convert app.OverviewStyles to tabs.OverviewStyles
	tabStyles := tabs.OverviewStyles{
		SectionTitle: m.overviewStyles.SectionTitle,
		Label:        m.overviewStyles.Label,
		Value:        m.overviewStyles.Value,
		Subtle:       m.overviewStyles.Subtle,
		Success:      m.overviewStyles.Success,
		Warning:      m.overviewStyles.Warning,
		Primary:      m.overviewStyles.Primary,
		Secondary:    m.overviewStyles.Secondary,
		Git:          m.overviewStyles.Git,
		Error:        m.overviewStyles.Error,
	}
	agentStyles := agent_details.Styles{
		OverviewStyles: tabStyles,
	}

	// Create agent details view (stateless rendering)
	detailView := agent_details.New(agentStyles, m.tabs)
	detailView.SetContext(ctx)
	detailView.SetSize(m.width, m.height)

	return detailView.View()
}
