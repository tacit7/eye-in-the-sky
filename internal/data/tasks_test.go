package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// MockTaskStore implements TaskStore interface for testing
type MockTaskStore struct {
	LoadCountsByAgentFunc func(ctx context.Context, agents []domain.Agent) ([]struct {
		AgentID domain.AgentID
		Count   int
	}, error)
	LoadByAgentFunc func(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error)
	LoadRecentByAgentFunc func(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.Task, error)
	MarkDoneFunc func(ctx context.Context, taskID domain.TaskID) error
}

func (m *MockTaskStore) LoadCountsByAgent(ctx context.Context, agents []domain.Agent) ([]struct {
	AgentID domain.AgentID
	Count   int
}, error) {
	if m.LoadCountsByAgentFunc != nil {
		return m.LoadCountsByAgentFunc(ctx, agents)
	}
	return nil, nil
}

func (m *MockTaskStore) LoadByAgent(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
	if m.LoadByAgentFunc != nil {
		return m.LoadByAgentFunc(ctx, agentID, limit, offset)
	}
	return nil, nil
}

func (m *MockTaskStore) LoadRecentByAgent(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.Task, error) {
	if m.LoadRecentByAgentFunc != nil {
		return m.LoadRecentByAgentFunc(ctx, agentID, limit)
	}
	return nil, nil
}

func (m *MockTaskStore) MarkDone(ctx context.Context, taskID domain.TaskID) error {
	if m.MarkDoneFunc != nil {
		return m.MarkDoneFunc(ctx, taskID)
	}
	return nil
}

func TestTaskStore_LoadCountsByAgent(t *testing.T) {
	// Create test agents
	agents := []domain.Agent{
		{
			ID:        domain.AgentID("agent1"),
			SessionID: "session_123",
			Status:    "active",
		},
		{
			ID:            domain.AgentID("subagent1"),
			ParentAgentID: "agent1",
			Status:        "working",
		},
	}

	t.Run("returns counts for agents", func(t *testing.T) {
		// This would normally test against real TaskWarrior
		// For unit tests, we use the mock
		mock := &MockTaskStore{
			LoadCountsByAgentFunc: func(ctx context.Context, agents []domain.Agent) ([]struct {
				AgentID domain.AgentID
				Count   int
			}, error) {
				return []struct {
					AgentID domain.AgentID
					Count   int
				}{
					{AgentID: "agent1", Count: 3},
					{AgentID: "subagent1", Count: 2},
				}, nil
			},
		}

		ctx := context.Background()
		counts, err := mock.LoadCountsByAgent(ctx, agents)

		assert.NoError(t, err)
		assert.Len(t, counts, 2)
		assert.Equal(t, 3, counts[0].Count)
		assert.Equal(t, 2, counts[1].Count)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		mock := &MockTaskStore{
			LoadCountsByAgentFunc: func(ctx context.Context, agents []domain.Agent) ([]struct {
				AgentID domain.AgentID
				Count   int
			}, error) {
				// Simulate long-running operation
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(100 * time.Millisecond):
					return nil, nil
				}
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := mock.LoadCountsByAgent(ctx, agents)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestTaskStore_LoadByAgent(t *testing.T) {
	agentID := domain.AgentID("test-agent")

	t.Run("returns tasks for agent", func(t *testing.T) {
		expectedTasks := []domain.Task{
			{
				ID:          domain.TaskID("task1"),
				Description: "Test task 1",
				Status:      "pending",
				Priority:    "H",
			},
			{
				ID:          domain.TaskID("task2"),
				Description: "Test task 2",
				Status:      "pending",
				Priority:    "M",
			},
		}

		mock := &MockTaskStore{
			LoadByAgentFunc: func(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
				return expectedTasks, nil
			},
		}

		ctx := context.Background()
		tasks, err := mock.LoadByAgent(ctx, agentID, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "Test task 1", tasks[0].Description)
		assert.Equal(t, "H", tasks[0].Priority)
	})

	t.Run("respects limit and offset", func(t *testing.T) {
		allTasks := []domain.Task{
			{ID: domain.TaskID("task1")},
			{ID: domain.TaskID("task2")},
			{ID: domain.TaskID("task3")},
			{ID: domain.TaskID("task4")},
			{ID: domain.TaskID("task5")},
		}

		mock := &MockTaskStore{
			LoadByAgentFunc: func(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
				end := offset + limit
				if end > len(allTasks) {
					end = len(allTasks)
				}
				if offset >= len(allTasks) {
					return []domain.Task{}, nil
				}
				return allTasks[offset:end], nil
			},
		}

		ctx := context.Background()

		// Get first 2 tasks
		tasks, err := mock.LoadByAgent(ctx, agentID, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, domain.TaskID("task1"), tasks[0].ID)

		// Get next 2 tasks
		tasks, err = mock.LoadByAgent(ctx, agentID, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, domain.TaskID("task3"), tasks[0].ID)

		// Get last task
		tasks, err = mock.LoadByAgent(ctx, agentID, 2, 4)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, domain.TaskID("task5"), tasks[0].ID)
	})
}

func TestTaskStore_MarkDone(t *testing.T) {
	t.Run("marks task as done", func(t *testing.T) {
		taskID := domain.TaskID("test-task-uuid")
		called := false

		mock := &MockTaskStore{
			MarkDoneFunc: func(ctx context.Context, tid domain.TaskID) error {
				called = true
				assert.Equal(t, taskID, tid)
				return nil
			},
		}

		ctx := context.Background()
		err := mock.MarkDone(ctx, taskID)

		assert.NoError(t, err)
		assert.True(t, called, "MarkDone should have been called")
	})

	t.Run("returns error on failure", func(t *testing.T) {
		taskID := domain.TaskID("test-task-uuid")
		expectedErr := fmt.Errorf("task not found")

		mock := &MockTaskStore{
			MarkDoneFunc: func(ctx context.Context, tid domain.TaskID) error {
				return expectedErr
			},
		}

		ctx := context.Background()
		err := mock.MarkDone(ctx, taskID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

// Example of table-driven tests
func TestTaskStore_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		agentID   domain.AgentID
		limit     int
		offset    int
		wantCount int
		wantErr   bool
	}{
		{"get all tasks", "agent1", 0, 0, 5, false},
		{"get first 3", "agent1", 3, 0, 3, false},
		{"get with offset", "agent1", 2, 2, 2, false},
		{"empty result", "agent-no-tasks", 10, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockTaskStore{
				LoadByAgentFunc: func(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
					if tt.wantErr {
						return nil, fmt.Errorf("error")
					}
					tasks := make([]domain.Task, tt.wantCount)
					return tasks, nil
				},
			}

			ctx := context.Background()
			tasks, err := mock.LoadByAgent(ctx, tt.agentID, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, tt.wantCount)
			}
		})
	}
}