package agent_details

import "github.com/tacit7/eye-in-the-sky/internal/domain"

// DataContext wraps all data needed by agent details tabs
// Container and tabs import it; nobody imports root app
type DataContext struct {
	Agent   *domain.Agent
	Commits []domain.Commit
	Notes   []domain.Note
	Tasks   []domain.Task
	Actions []domain.Action
}
