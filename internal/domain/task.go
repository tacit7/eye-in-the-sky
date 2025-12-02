package domain

import "time"

// TaskID is a unique identifier for a task (UUID)
type TaskID string

// Task represents an internal todo system task
type Task struct {
	ID          TaskID
	Title       string
	Description string
	ProjectID   int
	StateID     int       // Workflow state: 1=todo, 2=in_progress, 3=done
	Priority    int       // 0-5 priority level
	DueAt       time.Time // Optional due date
	CompletedAt time.Time // Optional completion timestamp
	SessionID   string    // Eye-in-the-Sky session ID (optional)
	AgentID     string    // Eye-in-the-Sky agent ID (optional)
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Archived    bool
	Notes       []TaskNote
	Tags        []string
	WorkflowStatus string // Human-readable state (todo, in_progress, done, etc.)
}

// TaskNote represents a note attached to a task
type TaskNote struct {
	ID          int
	TaskID      TaskID
	Author      string
	Body        string    // Markdown content
	CreatedAt   time.Time
}