package overview

import (
	"context"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// AgentStore interface matches the app.AgentStore interface
type AgentStore interface {
	LoadAgents(ctx context.Context, showAll bool) ([]domain.Agent, error)
}

// DataClientAdapter adapts app.DataClient to overview.DataClient
type DataClientAdapter struct {
	agentStore AgentStore
}

// NewDataClientAdapter creates a new data client adapter
func NewDataClientAdapter(agentStore AgentStore) *DataClientAdapter {
	return &DataClientAdapter{
		agentStore: agentStore,
	}
}

// LoadAgents loads agents from the database
func (d *DataClientAdapter) LoadAgents(showAll bool) ([]domain.Agent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.agentStore.LoadAgents(ctx, showAll)
}
