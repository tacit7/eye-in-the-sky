package app

// Tab index constants
const (
	TabOverview = iota
	TabCommits
	TabLogs
	TabNotes
	TabActions
	TabTasks
)

// DetailTabRenderers is a registry mapping tab indices to their renderer functions
var DetailTabRenderers = map[int]func(*Model) string{
	TabOverview: (*Model).renderOverviewTab,
	TabCommits:  (*Model).renderCommitsTabView,
	TabLogs:     (*Model).renderLogsTabView,
	TabNotes:    (*Model).renderNotesTabView,
	TabActions:  (*Model).renderActionsTabView,
	TabTasks:    (*Model).renderTasksTabView,
}
