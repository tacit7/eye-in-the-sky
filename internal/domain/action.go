package domain

import "time"

// ActionID represents a unique action identifier
type ActionID int

// Action represents an agent action
type Action struct {
	ID          ActionID
	AgentID     AgentID
	ActionType  string // task_start, file_operation, git_commit, status_update
	Description string
	Details     string // JSON data
	Timestamp   time.Time
}