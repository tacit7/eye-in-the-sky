package app

import (
	"context"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/agent_details"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/agent_details/tabs"
)

// renderDetail renders the agent detail view using new modular code
func (m *Model) renderDetail() string {
	if m.selectedAgent == nil {
		return "No agent selected"
	}

	// Load task notes for the selected task
	var taskNotes []domain.TaskNote
	if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
		selectedTask := m.tasks[m.tasksIndex]
		notes, err := m.data.Tasks.LoadTaskNotes(context.Background(), selectedTask.ID)
		if err == nil {
			taskNotes = notes
		}
	}

	// Create context for agent details
	ctx := &agent_details.DataContext{
		Agent:             m.selectedAgent,
		Commits:           m.commits,
		Notes:             m.notes,
		Tasks:             m.tasks,
		TaskNotes:         taskNotes,
		Actions:           m.actions,
		Logs:              m.logs,
		SessionContexts:   m.sessionContexts,
		SelectedTaskIndex: m.tasksIndex,
		NotesIndex:        m.notesIndex,
		CommitsIndex:      m.commitsIndex,
		Width:             m.width,
		MarkdownRenderer:  m.mdRenderer,
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
