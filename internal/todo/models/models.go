package models

import (
	"time"
)

// Project represents a project with tasks and workflow states.
type Project struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Path          *string    `json:"path,omitempty"`
	RemoteURL     *string    `json:"remote_url,omitempty"`
	Subpath       *string    `json:"subpath,omitempty"`
	Module        *string    `json:"module,omitempty"`
	Salt          *string    `json:"salt,omitempty"`
	IDAlgorithm   string     `json:"id_algorithm"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastCommit    *string    `json:"last_commit,omitempty"`
	Active        bool       `json:"active"`
}

// WorkflowState represents a state in the workflow (e.g., "todo", "in_progress", "done").
type WorkflowState struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Position *int    `json:"position,omitempty"`
	Color    *string `json:"color,omitempty"`
}

// Task represents a single task.
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	StateID     *int       `json:"state_id,omitempty"`
	ProjectID   string     `json:"project_id"`
	Priority    int        `json:"priority"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	SessionID   *string    `json:"session_id,omitempty"` // Eye-in-the-Sky session ID
	AgentID     *string    `json:"agent_id,omitempty"`   // Eye-in-the-Sky agent ID
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Archived    bool       `json:"archived"`
	Notes       []Note     `json:"notes,omitempty"`
	Tags        []Tag      `json:"tags,omitempty"`
}

// Note represents a note attached to a task.
type Note struct {
	ID        int       `json:"id"`
	TaskID    string    `json:"task_id"`
	Author    *string   `json:"author,omitempty"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// Tag represents a global tag that can be applied to tasks.
type Tag struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color,omitempty"`
}

// TaskTag represents the association between a task and a tag.
type TaskTag struct {
	TaskID string `json:"task_id"`
	TagID  int    `json:"tag_id"`
}

// TaskEvent represents an audit event for task changes.
type TaskEvent struct {
	ID        int        `json:"id"`
	TaskID    string     `json:"task_id"`
	EventType string     `json:"event_type"`
	Payload   *string    `json:"payload,omitempty"` // JSON string
	Actor     *string    `json:"actor,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// SearchResult represents a task search result with ranking.
type SearchResult struct {
	Task Task    `json:"task"`
	Rank float64 `json:"rank"`
}

// Filters for task queries.
type Filters struct {
	StateID  *int
	Tags     []string
	DueBefore *time.Time
	DueAfter  *time.Time
	Priority  *int
	HasNote   bool
	IsActive  bool // archived = 0
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
	Title       string
	Description *string
	StateID     *int
	Priority    *int
	DueAt       *time.Time
	SessionID   *string // Eye-in-the-Sky session ID
	AgentID     *string // Eye-in-the-Sky agent ID
}

// UpdateTaskInput is the input for updating a task.
type UpdateTaskInput struct {
	Title       *string
	Description *string
	StateID     *int
	Priority    *int
	DueAt       *time.Time
	CompletedAt *time.Time
	Archived    *bool
}
