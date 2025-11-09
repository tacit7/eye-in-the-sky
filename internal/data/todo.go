package data

import (
	"context"
	"fmt"
	"log"
	"time"

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

			// Count tasks belonging to this agent (session_ids contains agent.SessionID or agent_id match)
			for _, task := range tasks {
				sessionMatches := false
				for _, sid := range task.SessionIDs {
					if sid == agent.SessionID {
						sessionMatches = true
						break
					}
				}
				if sessionMatches || (task.AgentID != nil && *task.AgentID == string(agent.ID)) {
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
// Filters tasks by agent_id in a single SQL query (no N+1 pattern)
func (s *todoStore) LoadByAgent(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
	// Return empty if service not initialized
	if s.svc == nil {
		return []domain.Task{}, nil
	}

	start := time.Now()
	startTime := start.Format("15:04:05.000")
	log.Printf("[PERF][%s] LoadByAgent starting for agent %s", startTime, agentID)

	// Single SQL query: join tasks, workflow_states, and projects in one round-trip
	query := `
		SELECT
			t.id,
			t.title,
			t.description,
			t.state_id,
			w.name AS workflow_name,
			w.color AS workflow_color,
			p.id AS project_id,
			p.name AS project_name,
			t.priority,
			t.created_at,
			t.updated_at,
			t.archived,
			t.session_id,
			t.agent_id,
			t.due_at,
			t.completed_at
		FROM tasks t
		LEFT JOIN workflow_states w ON w.id = t.state_id
		LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.agent_id = ?
		AND t.archived = 0
		ORDER BY t.priority DESC, t.state_id ASC, t.created_at ASC
		LIMIT ? OFFSET ?
	`

	rows, err := s.svc.GetDB().Query(query, string(agentID), limit, offset)
	queryTime := time.Since(start)
	queryEndTime := time.Now().Format("15:04:05.000")
	log.Printf("[PERF][%s] Task query for agent %s took %v", queryEndTime, agentID, queryTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var (
			id            string
			title         string
			description   *string
			stateID       int
			workflowName  *string
			workflowColor *string
			projectID     *int
			projectName   *string
			priority      int
			createdAt     string
			updatedAt     *string
			archived      bool
			sessionID     *string
			agentIDStr    *string
			dueAt         *string
			completedAt   *string
		)

		if err := rows.Scan(&id, &title, &description, &stateID, &workflowName, &workflowColor,
			&projectID, &projectName, &priority, &createdAt, &updatedAt, &archived, &sessionID, &agentIDStr, &dueAt, &completedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		// Parse timestamps
		createdTime, _ := time.Parse(time.RFC3339, createdAt)
		var updatedTime time.Time
		if updatedAt != nil {
			updatedTime, _ = time.Parse(time.RFC3339, *updatedAt)
		} else {
			updatedTime = createdTime
		}

		projectIDVal := 0
		if projectID != nil {
			projectIDVal = *projectID
		}

		task := domain.Task{
			ID:               domain.TaskID(id),
			Title:            title,
			Description:      derefString(description),
			StateID:          stateID,
			WorkflowStatus:   derefString(workflowName),
			ProjectID:        projectIDVal,
			Priority:         priority,
			CreatedAt:        createdTime,
			UpdatedAt:        updatedTime,
			Archived:         archived,
			SessionID:        derefString(sessionID),
			AgentID:          derefString(agentIDStr),
		}

		if dueAt != nil {
			t, _ := time.Parse(time.RFC3339, *dueAt)
			task.DueAt = t
		}

		if completedAt != nil {
			t, _ := time.Parse(time.RFC3339, *completedAt)
			task.CompletedAt = t
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading task rows: %w", err)
	}

	return tasks, nil
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

func (s *todoStore) MarkTodo(ctx context.Context, taskID domain.TaskID) error {
	// Get the task first to find the "todo" state ID
	repo := s.svc.GetTasksRepo()
	_, err := repo.FindByID(string(taskID))
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get workflow states to find "todo" state
	states, err := s.svc.GetProjectsRepo().GetWorkflowStates()
	if err != nil {
		return fmt.Errorf("failed to get workflow states: %w", err)
	}

	var todoStateID *int
	for _, state := range states {
		if state.Name == "todo" {
			todoStateID = &state.ID
			break
		}
	}

	if todoStateID == nil {
		// Default to state_id 1 if "todo" state not found
		id := 1
		todoStateID = &id
	}

	// Update task to todo state
	_, err = repo.MoveToState(string(taskID), *todoStateID)
	if err != nil {
		return fmt.Errorf("failed to mark task todo: %w", err)
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

	var projectID int
	for _, p := range projects {
		if p.Name == projectName {
			projectID = p.ID
			break
		}
	}

	if projectID == 0 {
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

	// Take first session ID if available (domain.Task uses single SessionID)
	if len(t.SessionIDs) > 0 {
		task.SessionID = t.SessionIDs[0]
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

// LoadTaskNotes loads task notes (annotations) for a specific task
func (s *todoStore) LoadTaskNotes(ctx context.Context, taskID domain.TaskID) ([]domain.TaskNote, error) {
	if s.svc == nil {
		return []domain.TaskNote{}, nil
	}

	query := `
		SELECT id, task_id, author, body, created_at
		FROM task_notes
		WHERE task_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.svc.GetDB().Query(query, string(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to query task notes: %w", err)
	}
	defer rows.Close()

	var notes []domain.TaskNote
	for rows.Next() {
		var (
			id        int
			taskIDStr string
			author    *string
			body      string
			createdAt string
		)

		if err := rows.Scan(&id, &taskIDStr, &author, &body, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan task note row: %w", err)
		}

		createdTime, _ := time.Parse(time.RFC3339, createdAt)

		note := domain.TaskNote{
			ID:        id,
			TaskID:    domain.TaskID(taskIDStr),
			Author:    derefString(author),
			Body:      body,
			CreatedAt: createdTime,
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading task note rows: %w", err)
	}

	return notes, nil
}

// AddTaskNote adds an annotation to a task
func (s *todoStore) AddTaskNote(ctx context.Context, taskID domain.TaskID, body string) error {
	if s.svc == nil {
		return fmt.Errorf("todo service not available")
	}

	_, err := s.svc.AddTaskNote(string(taskID), body)
	return err
}

// Helper function to dereference string pointers
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
