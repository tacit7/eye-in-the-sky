package models

import (
	"time"
)

// Project represents a project with tasks and workflow states.
type Project struct {
	ID          int        `json:"id"`
	UUID        string     `json:"uuid"`
	Name        string     `json:"name"`
	RepoSlug    *string    `json:"repo_slug,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
}

// WorkflowState represents a state in a project's workflow (e.g., "todo", "doing", "done").
type WorkflowState struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"project_id"`
	Code        string    `json:"code"`
	DisplayName string    `json:"display_name"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Task represents a single task or subtask.
type Task struct {
	ID          int        `json:"id"`
	ProjectID   int        `json:"project_id"`
	Description string     `json:"description"`
	StateCode   *string    `json:"state_code,omitempty"`
	ParentID    *int       `json:"parent_id,omitempty"`
	Priority    *int       `json:"priority,omitempty"` // 1-5
	Weight      *int       `json:"weight,omitempty"`
	Position    int        `json:"position"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	SessionID   *string    `json:"session_id,omitempty"` // Eye-in-the-Sky session ID
	AgentID     *string    `json:"agent_id,omitempty"`   // Eye-in-the-Sky agent ID
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	Notes       []Note     `json:"notes,omitempty"`
	Tags        []Tag      `json:"tags,omitempty"`
}

// Note represents a markdown note attached to a task.
type Note struct {
	ID            int       `json:"id"`
	TaskID        int       `json:"task_id"`
	BodyMarkdown  string    `json:"body_markdown"`
	CreatedAt     time.Time `json:"created_at"`
}

// Tag represents a global tag that can be applied to tasks.
type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskTag represents the association between a task and a tag.
type TaskTag struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	TagID     int       `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskEvent represents an audit event for task changes.
type TaskEvent struct {
	ID        int        `json:"id"`
	TaskID    int        `json:"task_id"`
	EventType string     `json:"event_type"`
	OldValue  *string    `json:"old_value,omitempty"`
	NewValue  *string    `json:"new_value,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// SearchResult represents a task search result with ranking.
type SearchResult struct {
	Task Task    `json:"task"`
	Rank float64 `json:"rank"`
}

// Filters for task queries.
type Filters struct {
	StateCode *string
	Tags      []string
	DueBefore *time.Time
	DueAfter  *time.Time
	Priority  *int
	HasNote   bool
	IsActive  bool // archived_at IS NULL
}

// SortOrder for task queries.
type SortOrder string

const (
	SortByPosition SortOrder = "position"
	SortByDue      SortOrder = "due_date"
	SortByPriority SortOrder = "priority"
	SortByCreated  SortOrder = "created_at"
	SortByUpdated  SortOrder = "updated_at"
)

// CreateTaskInput is the input for creating a new task.
type CreateTaskInput struct {
	Description string
	ParentID    *int
	StateCode   *string
	Priority    *int
	Weight      *int
	SessionID   *string // Eye-in-the-Sky session ID
	AgentID     *string // Eye-in-the-Sky agent ID
}

// UpdateTaskInput is the input for updating a task.
type UpdateTaskInput struct {
	Description *string
	StateCode   *string
	Priority    *int
	Weight      *int
	DueDate     *time.Time
}
