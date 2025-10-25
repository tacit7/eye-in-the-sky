package data

import (
	"context"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/todo"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
)

// todoStore implements TaskStore interface using the internal todo system
type todoStore struct {
	svc *todo.Service
}

// NewTodoStore creates a new todo store
func NewTodoStore(svc *todo.Service) *todoStore {
	if svc == nil {
		// Return empty store - will return no tasks but won't crash
		return &todoStore{svc: nil}
	}
	return &todoStore{svc: svc}
}

// LoadCountsByAgent returns task counts for multiple agents
// Queries all tasks for agents filtered by session_id or agent_id tags
func (s *todoStore) LoadCountsByAgent(ctx context.Context, agents []domain.Agent) ([]struct {
	AgentID domain.AgentID
	Count   int
}, error) {
	// Return empty counts if service is not initialized
	if s.svc == nil {
		result := make([]struct {
			AgentID domain.AgentID
			Count   int
		}, 0, len(agents))
		return result, nil
	}

	result := make([]struct {
		AgentID domain.AgentID
		Count   int
	}, 0, len(agents))

	for _, agent := range agents {
		// Get all projects for counting
		projects, err := s.svc.ListProjects()
		if err != nil {
			continue // Skip agent if error
		}

		count := 0
		for _, project := range projects {
			// Query tasks for this project filtered by agent
			filters := models.Filters{
				IsActive: true, // Only non-archived
			}

			tasks, err := s.svc.GetTasksRepo().List(project.ID, filters, models.SortByCreated)
			if err != nil {
				continue
			}

			// Count tasks belonging to this agent (session_id or agent_id match)
			for _, task := range tasks {
				if (task.SessionID != nil && *task.SessionID == agent.SessionID) ||
					(task.AgentID != nil && *task.AgentID == string(agent.ID)) {
					count++
				}
			}
		}

		result = append(result, struct {
			AgentID domain.AgentID
			Count   int
		}{
			AgentID: agent.ID,
			Count:   count,
		})
	}

	return result, nil
}

// LoadByAgent returns tasks for a specific agent with pagination
// Filters tasks by session_id or agent_id
func (s *todoStore) LoadByAgent(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
	// Return empty if service not initialized
	if s.svc == nil {
		return []domain.Task{}, nil
	}

	// Get all projects
	projects, err := s.svc.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var allTasks []domain.Task
	for _, project := range projects {
		// Query tasks for this project
		filters := models.Filters{
			IsActive: true, // Only non-archived
		}

		tasks, err := s.svc.GetTasksRepo().List(project.ID, filters, models.SortByCreated)
		if err != nil {
			continue // Skip project on error
		}

		// Filter tasks by agent
		for _, task := range tasks {
			// Match by session_id or agent_id
			if (task.SessionID != nil && *task.SessionID == string(agentID)) ||
				(task.AgentID != nil && *task.AgentID == string(agentID)) {
				allTasks = append(allTasks, s.toDomainTask(task))
			}
		}
	}

	// Apply pagination
	if offset >= len(allTasks) {
		return []domain.Task{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(allTasks) {
		end = len(allTasks)
	}

	return allTasks[offset:end], nil
}

// LoadRecentByAgent returns the most recent tasks for an agent
func (s *todoStore) LoadRecentByAgent(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.Task, error) {
	tasks, err := s.LoadByAgent(ctx, agentID, 0, 0) // Get all tasks first
	if err != nil {
		return nil, err
	}

	// Return only the requested limit (already in created_at order from query)
	if limit > 0 && len(tasks) > limit {
		return tasks[:limit], nil
	}

	return tasks, nil
}

// MarkDone marks a task as completed by moving it to the "done" state
func (s *todoStore) MarkDone(ctx context.Context, taskID domain.TaskID) error {
	// Get the task first to find the "done" state ID
	repo := s.svc.GetTasksRepo()
	_, err := repo.FindByID(string(taskID))
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get workflow states to find "done" state
	states, err := s.svc.GetProjectsRepo().GetWorkflowStates()
	if err != nil {
		return fmt.Errorf("failed to get workflow states: %w", err)
	}

	var doneStateID *int
	for _, state := range states {
		if state.Name == "done" {
			doneStateID = &state.ID
			break
		}
	}

	if doneStateID == nil {
		// Default to state_id 3 if "done" state not found
		id := 3
		doneStateID = &id
	}

	// Update task to done state
	_, err = repo.MoveToState(string(taskID), *doneStateID)
	if err != nil {
		return fmt.Errorf("failed to mark task done: %w", err)
	}

	return nil
}

// LoadByProject loads tasks for a specific project
func (s *todoStore) LoadByProject(ctx context.Context, projectName string, limit int) ([]domain.Task, error) {
	// Find project by name
	projects, err := s.svc.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var projectID string
	for _, p := range projects {
		if p.Name == projectName {
			projectID = p.ID
			break
		}
	}

	if projectID == "" {
		return []domain.Task{}, nil // Project not found, return empty
	}

	// Query tasks for this project
	filters := models.Filters{
		IsActive: true, // Only non-archived
	}

	tasks, err := s.svc.GetTasksRepo().List(projectID, filters, models.SortByCreated)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Convert to domain tasks
	var domainTasks []domain.Task
	for _, task := range tasks {
		domainTasks = append(domainTasks, s.toDomainTask(task))
		if limit > 0 && len(domainTasks) >= limit {
			break
		}
	}

	return domainTasks, nil
}

// toDomainTask converts a todo model task to domain task
func (s *todoStore) toDomainTask(t models.Task) domain.Task {
	task := domain.Task{
		ID:          domain.TaskID(t.ID),
		Title:       t.Title,
		ProjectID:   t.ProjectID,
		StateID:     *t.StateID,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt,
		Archived:    t.Archived,
	}

	// Optional fields
	if t.Description != nil {
		task.Description = *t.Description
	}

	if t.DueAt != nil {
		task.DueAt = *t.DueAt
	}

	if t.CompletedAt != nil {
		task.CompletedAt = *t.CompletedAt
	}

	if t.UpdatedAt != nil {
		task.UpdatedAt = *t.UpdatedAt
	} else {
		task.UpdatedAt = t.CreatedAt // Fallback to created_at
	}

	if t.SessionID != nil {
		task.SessionID = *t.SessionID
	}

	if t.AgentID != nil {
		task.AgentID = *t.AgentID
	}

	// Convert notes
	for _, note := range t.Notes {
		task.Notes = append(task.Notes, domain.TaskNote{
			ID:        note.ID,
			TaskID:    domain.TaskID(note.TaskID),
			Author:    derefString(note.Author),
			Body:      note.Body,
			CreatedAt: note.CreatedAt,
		})
	}

	// Extract tags
	for _, tag := range t.Tags {
		task.Tags = append(task.Tags, tag.Name)
	}

	// Set workflow status based on state_id
	task.WorkflowStatus = s.getWorkflowStatusForState(*t.StateID)

	return task
}

// getWorkflowStatusForState returns the human-readable workflow status for a state ID
func (s *todoStore) getWorkflowStatusForState(stateID int) string {
	states, err := s.svc.GetProjectsRepo().GetWorkflowStates()
	if err != nil {
		return ""
	}

	for _, state := range states {
		if state.ID == stateID {
			return state.Name
		}
	}

	return ""
}

// Helper function to dereference string pointers
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
