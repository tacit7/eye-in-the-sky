package app

import (
	"database/sql"

	"github.com/tacit7/eye-in-the-sky/internal/data"
)

// DataClient aggregates all data stores
type DataClient struct {
	Agents   AgentStore
	Tasks    TaskStore
	Metrics  MetricsStore
	Notes    NotesStore
	Commits  CommitsStore
	Actions  ActionsStore
}

// NewDataClient creates a new data client with all stores initialized
func NewDataClient(db *sql.DB) *DataClient {
	return &DataClient{
		Agents:   data.NewAgentStore(db),
		Tasks:    data.NewTaskStore(),
		Metrics:  data.NewMetricsStore(db),
		Notes:    data.NewNotesStore(db),
		Commits:  data.NewCommitsStore(db),
		Actions:  data.NewActionsStore(db),
	}
}