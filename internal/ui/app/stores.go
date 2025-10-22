package app

import (
	"context"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// AgentStore handles agent data operations
type AgentStore interface {
	LoadAgents(ctx context.Context) ([]domain.Agent, error)
	LoadAgent(ctx context.Context, agentID domain.AgentID) (*domain.Agent, error)
	UpdateStatus(ctx context.Context, agentID domain.AgentID, status string) error
}

// TaskStore handles TaskWarrior operations
type TaskStore interface {
	// LoadCountsByAgent returns task counts for multiple agents
	LoadCountsByAgent(ctx context.Context, agents []domain.Agent) ([]struct {
		AgentID domain.AgentID
		Count   int
	}, error)

	// LoadByAgent returns tasks for a specific agent with pagination
	LoadByAgent(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error)

	// LoadRecentByAgent returns the most recent tasks for an agent
	LoadRecentByAgent(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.Task, error)

	// MarkDone marks a task as completed
	MarkDone(ctx context.Context, taskID domain.TaskID) error
}

// MetricsStore handles session metrics operations
type MetricsStore interface {
	LoadAll(ctx context.Context) ([]domain.SessionMetric, error)
	LoadByAgent(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.SessionMetric, error)
	LoadByTimeRange(ctx context.Context, start, end time.Time) ([]domain.SessionMetric, error)
	LoadMonthly(ctx context.Context, year int, month time.Month) ([]domain.SessionMetric, error)
}

// NotesStore handles session notes operations
type NotesStore interface {
	LoadByAgent(ctx context.Context, agentID domain.AgentID) ([]domain.Note, error)
	Create(ctx context.Context, agentID domain.AgentID, content string) error
	Delete(ctx context.Context, noteID domain.NoteID) error
}

// CommitsStore handles git commit operations
type CommitsStore interface {
	LoadByAgent(ctx context.Context, agentID domain.AgentID) ([]domain.Commit, error)
	LoadRecent(ctx context.Context, limit int) ([]domain.Commit, error)
}

// ActionsStore handles agent action operations
type ActionsStore interface {
	LoadByAgent(ctx context.Context, agentID domain.AgentID) ([]domain.Action, error)
	Create(ctx context.Context, action domain.Action) error
}