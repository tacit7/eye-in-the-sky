package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
	"github.com/tacit7/eye-in-the-sky/internal/todo/util"
)

// TaskResponse is the standard response format for task operations.
type TaskResponse struct {
	TaskID      int    `json:"task_id"`
	Description string `json:"description"`
	UUIDShort   string `json:"uuid_short"`
}

// ListTaskResponse includes task details in list operations.
type ListTaskResponse struct {
	ID          int      `json:"id"`
	Description string   `json:"description"`
	Priority    *int     `json:"priority,omitempty"`
	State       *string  `json:"state,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// SearchTaskResponse includes ranking for search results.
type SearchTaskResponse struct {
	TaskID      int     `json:"task_id"`
	Description string  `json:"description"`
	Rank        float64 `json:"rank"`
}

// ============================================================================
// todo.create - Create a new task
// ============================================================================

type CreateRequest struct {
	ProjectID   int      `json:"project_id"`
	Description string   `json:"description"`
	Priority    *int     `json:"priority,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ParentID    *int     `json:"parent_id,omitempty"`
	SessionID   *string  `json:"session_id,omitempty"`
	AgentID     *string  `json:"agent_id,omitempty"`
}

func (h *Handler) HandleCreate(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req CreateRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid create request: %w", err)
	}

	// Validate inputs
	if err := util.ValidateDescription(req.Description); err != nil {
		return nil, err
	}
	if err := util.ValidatePriority(req.Priority); err != nil {
		return nil, err
	}

	// Start transaction for write operation
	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Create task
	input := models.CreateTaskInput{
		Description: req.Description,
		ParentID:    req.ParentID,
		Priority:    req.Priority,
		StateCode:   stringPtr("todo"),
		SessionID:   req.SessionID,
		AgentID:     req.AgentID,
	}

	task, err := h.svc.GetTasksRepo().CreateTask(req.ProjectID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Add tags if provided
	for _, tagName := range req.Tags {
		if err := util.ValidateTagName(tagName); err != nil {
			return nil, err
		}
		if _, err := h.svc.GetTasksRepo().AddTag(task.ID, tagName); err != nil {
			return nil, fmt.Errorf("failed to add tag: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return TaskResponse{
		TaskID:      task.ID,
		Description: task.Description,
		UUIDShort:   generateUUIDShort(task.ID),
	}, nil
}

// ============================================================================
// todo.annotate - Add a note to a task
// ============================================================================

type AnnotateRequest struct {
	TaskID int    `json:"task_id"`
	Body   string `json:"body"`
}

func (h *Handler) HandleAnnotate(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req AnnotateRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid annotate request: %w", err)
	}

	if err := util.ValidateNoteContent(req.Body); err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Add note
	_, err = h.svc.GetTasksRepo().AddNote(req.TaskID, req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to add note: %w", err)
	}

	// Get task for response
	task, err := h.svc.GetTasksRepo().FindByID(req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return TaskResponse{
		TaskID:      task.ID,
		Description: task.Description,
		UUIDShort:   generateUUIDShort(task.ID),
	}, nil
}

// ============================================================================
// todo.start - Move task to doing state
// ============================================================================

type StateChangeRequest struct {
	TaskID int `json:"task_id"`
}

func (h *Handler) HandleStart(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req StateChangeRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid start request: %w", err)
	}

	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	task, err := h.svc.GetTasksRepo().MoveToState(req.TaskID, "doing")
	if err != nil {
		return nil, fmt.Errorf("failed to move task to doing: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return TaskResponse{
		TaskID:      task.ID,
		Description: task.Description,
		UUIDShort:   generateUUIDShort(task.ID),
	}, nil
}

// ============================================================================
// todo.done - Move task to done state
// ============================================================================

func (h *Handler) HandleDone(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req StateChangeRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid done request: %w", err)
	}

	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	task, err := h.svc.GetTasksRepo().MoveToState(req.TaskID, "done")
	if err != nil {
		return nil, fmt.Errorf("failed to move task to done: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return TaskResponse{
		TaskID:      task.ID,
		Description: task.Description,
		UUIDShort:   generateUUIDShort(task.ID),
	}, nil
}

// ============================================================================
// todo.status - Move task to any workflow state
// ============================================================================

type StatusRequest struct {
	TaskID int    `json:"task_id"`
	State  string `json:"state"`
}

func (h *Handler) HandleStatus(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req StatusRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid status request: %w", err)
	}

	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	task, err := h.svc.GetTasksRepo().MoveToState(req.TaskID, req.State)
	if err != nil {
		return nil, fmt.Errorf("failed to move task to state %s: %w", req.State, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return TaskResponse{
		TaskID:      task.ID,
		Description: task.Description,
		UUIDShort:   generateUUIDShort(task.ID),
	}, nil
}

// ============================================================================
// todo.tag - Add or remove tags from a task
// ============================================================================

type TagRequest struct {
	TaskID int      `json:"task_id"`
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

func (h *Handler) HandleTag(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req TagRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid tag request: %w", err)
	}

	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Add tags
	for _, tagName := range req.Add {
		if err := util.ValidateTagName(tagName); err != nil {
			return nil, err
		}
		if _, err := h.svc.GetTasksRepo().AddTag(req.TaskID, tagName); err != nil {
			return nil, fmt.Errorf("failed to add tag: %w", err)
		}
	}

	// Remove tags
	for _, tagName := range req.Remove {
		if err := h.svc.GetTasksRepo().RemoveTag(req.TaskID, tagName); err != nil {
			return nil, fmt.Errorf("failed to remove tag: %w", err)
		}
	}

	task, err := h.svc.GetTasksRepo().FindByID(req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return TaskResponse{
		TaskID:      task.ID,
		Description: task.Description,
		UUIDShort:   generateUUIDShort(task.ID),
	}, nil
}

// ============================================================================
// todo.list - Retrieve tasks with filters
// ============================================================================

type ListRequest struct {
	ProjectID int       `json:"project_id"`
	Filters   *Filters  `json:"filters,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

type Filters struct {
	State    *string  `json:"state,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Priority *int     `json:"priority,omitempty"`
	Active   bool     `json:"active,omitempty"`
}

type ListResponse struct {
	Tasks []ListTaskResponse `json:"tasks"`
}

func (h *Handler) HandleList(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req ListRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid list request: %w", err)
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	// Build filters
	filters := models.Filters{
		IsActive: req.Filters != nil && req.Filters.Active,
	}

	if req.Filters != nil {
		filters.StateCode = req.Filters.State
		filters.Tags = req.Filters.Tags
		filters.Priority = req.Filters.Priority
	}

	tasks, err := h.svc.GetTasksRepo().List(req.ProjectID, filters, models.SortByPosition)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	response := ListResponse{Tasks: make([]ListTaskResponse, 0, len(tasks))}
	for _, task := range tasks {
		tagNames := make([]string, len(task.Tags))
		for i, tag := range task.Tags {
			tagNames[i] = tag.Name
		}
		response.Tasks = append(response.Tasks, ListTaskResponse{
			ID:          task.ID,
			Description: task.Description,
			Priority:    task.Priority,
			State:       task.StateCode,
			Tags:        tagNames,
		})
	}

	return response, nil
}

// ============================================================================
// todo.search - Full-text search on tasks
// ============================================================================

type SearchRequest struct {
	ProjectID int    `json:"project_id"`
	Query     string `json:"query"`
	Limit     int    `json:"limit,omitempty"`
}

type SearchResponse struct {
	Results []SearchTaskResponse `json:"results"`
}

func (h *Handler) HandleSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req SearchRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	if err := util.ValidateSearchQuery(req.Query); err != nil {
		return nil, err
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	results, err := h.svc.GetTasksRepo().Search(req.ProjectID, req.Query, req.Limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	response := SearchResponse{Results: make([]SearchTaskResponse, 0, len(results))}
	for _, result := range results {
		response.Results = append(response.Results, SearchTaskResponse{
			TaskID:      result.Task.ID,
			Description: result.Task.Description,
			Rank:        result.Rank,
		})
	}

	return response, nil
}

// ============================================================================
// todo.delete - Permanently delete a task
// ============================================================================

type DeleteRequest struct {
	TaskID int `json:"task_id"`
}

type OKResponse struct {
	OK bool `json:"ok"`
}

func (h *Handler) HandleDelete(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req DeleteRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid delete request: %w", err)
	}

	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := h.svc.HardDeleteTask(req.TaskID); err != nil {
		return nil, fmt.Errorf("failed to delete task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return OKResponse{OK: true}, nil
}

// ============================================================================
// todo.reindex - Rebuild FTS5 index
// ============================================================================

func (h *Handler) HandleReindex(ctx context.Context, args json.RawMessage) (interface{}, error) {
	if err := h.svc.Reindex(); err != nil {
		return nil, fmt.Errorf("failed to reindex: %w", err)
	}
	return OKResponse{OK: true}, nil
}

// ============================================================================
// todo.vacuum - Database maintenance
// ============================================================================

func (h *Handler) HandleVacuum(ctx context.Context, args json.RawMessage) (interface{}, error) {
	if err := h.svc.Vacuum(); err != nil {
		return nil, fmt.Errorf("failed to vacuum: %w", err)
	}
	return OKResponse{OK: true}, nil
}

// ============================================================================
// todo.project.sync - Sync workflow states from YAML
// ============================================================================

type ProjectSyncRequest struct {
	ProjectID int    `json:"project_id"`
	YAML      string `json:"yaml"`
}

func (h *Handler) HandleProjectSync(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req ProjectSyncRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, fmt.Errorf("invalid project sync request: %w", err)
	}

	tx, err := h.svc.GetDB().BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := h.svc.SyncWorkflowFromYAML(req.ProjectID, []byte(req.YAML)); err != nil {
		return nil, fmt.Errorf("failed to sync workflow: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return OKResponse{OK: true}, nil
}

// ============================================================================
// Helper functions
// ============================================================================

// generateUUIDShort generates a short UUID-like string from a task ID.
func generateUUIDShort(taskID int) string {
	// Generate a deterministic 6-character string from task ID
	// This is a simple hash for demonstration; in production, store actual UUIDs
	hash := fmt.Sprintf("%06d", taskID%1000000)
	return hash
}

// stringPtr creates a pointer to a string.
func stringPtr(s string) *string {
	return &s
}
