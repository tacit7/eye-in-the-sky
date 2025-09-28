package database

import "time"

// Agent represents a Claude Code instance
type Agent struct {
	ID                 string     `json:"id"`
	Status             string     `json:"status"`
	Source             string     `json:"source"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	GitWorktreePath    *string    `json:"git_worktree_path,omitempty"`
	FeatureDescription *string    `json:"feature_description,omitempty"`
	CurrentTask        *string    `json:"current_task,omitempty"`
	LastActivityAt     *time.Time `json:"last_activity_at,omitempty"`
	WindowID           *string    `json:"window_id,omitempty"`
}

// Action represents an activity performed by an agent
type Action struct {
	ID          int       `json:"id"`
	AgentID     string    `json:"agent_id"`
	Timestamp   time.Time `json:"timestamp"`
	ActionType  string    `json:"action_type"`
	Description string    `json:"description"`
	Details     *string   `json:"details,omitempty"`
}

// Commit represents a git commit made by an agent
type Commit struct {
	ID            int       `json:"id"`
	AgentID       string    `json:"agent_id"`
	CommitHash    string    `json:"commit_hash"`
	CommitMessage *string   `json:"commit_message,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// Valid agent statuses
const (
	StatusActive    = "active"
	StatusIdle      = "idle"
	StatusWorking   = "working"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// Valid agent sources
const (
	SourceWorktree = "worktree"
	SourceDesktop  = "desktop"
)

// Valid action types
const (
	ActionTaskStart     = "task_start"
	ActionFileOperation = "file_operation"
	ActionGitCommit     = "git_commit"
	ActionStatusUpdate  = "status_update"
)
