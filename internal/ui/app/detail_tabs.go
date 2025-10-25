package app

// Tab index constants (must match the order in tabs array in model.go)
// Visual order: [← Back] [O]verview [T]asks [A]ctions [L]ogs [C]ommits [N]otes
// After -1 adjustment for Back button:
const (
	TabOverview = iota // 0 -> visual index 1
	TabTasks           // 1 -> visual index 2
	TabActions         // 2 -> visual index 3
	TabLogs            // 3 -> visual index 4
	TabCommits         // 4 -> visual index 5
	TabNotes           // 5 -> visual index 6
)

// DetailTabRenderers is a registry mapping tab indices to their renderer functions
var DetailTabRenderers = map[int]func(*Model) string{
	TabOverview: (*Model).renderOverviewTab,
	TabTasks:    (*Model).renderTasksTabView,
	TabActions:  (*Model).renderActionsTabView,
	TabLogs:     (*Model).renderLogsTabView,
	TabCommits:  (*Model).renderCommitsTabView,
	TabNotes:    (*Model).renderNotesTabView,
}
