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
	ProjectName        *string    `json:"project_name,omitempty"`
	CurrentSessionID   *string    `json:"current_session_id,omitempty"`
	PersonaID          *string    `json:"persona_id,omitempty"`
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
	StatusArchived  = "archived"
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

// Session represents a tracked agent session
type Session struct {
	ID        string     `json:"id"`
	AgentID   string     `json:"agent_id"`
	Name      *string    `json:"name,omitempty"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// Log represents a session log entry
type Log struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Note represents a session note
type Note struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Context represents a session context key-value pair
type Context struct {
	ID        int    `json:"id"`
	SessionID string `json:"session_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

// Persona represents an expert agent template with predefined expertise
type Persona struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Expertise      string    `json:"expertise"`       // JSON array of expertise areas
	InitialContext string    `json:"initial_context"` // The persona's initial instructions
	PreferredTools *string   `json:"preferred_tools,omitempty"` // JSON array of preferred MCP tools
	Specialization *string   `json:"specialization,omitempty"`  // Primary domain
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SessionWithAgent represents a session with its associated agent information
type SessionWithAgent struct {
	ID                 string     `json:"id"`
	AgentID            string     `json:"agent_id"`
	Name               *string    `json:"name,omitempty"`
	StartedAt          time.Time  `json:"started_at"`
	EndedAt            *time.Time `json:"ended_at,omitempty"`
	AgentStatus        string     `json:"agent_status"`
	FeatureDescription *string    `json:"feature_description,omitempty"`
	CurrentTask        *string    `json:"current_task,omitempty"`
	ProjectName        *string    `json:"project_name,omitempty"`
}

// Compaction represents a conversation compaction event
type Compaction struct {
	ID            int       `json:"id"`
	AgentID       string    `json:"agent_id"`
	OldSessionID  *string   `json:"old_session_id,omitempty"`
	NewSessionID  string    `json:"new_session_id"`
	CompactedAt   time.Time `json:"compacted_at"`
	Summary       *string   `json:"summary,omitempty"`
	JsonlFilePath *string   `json:"jsonl_file_path,omitempty"`
	JsonlFileSize *int64    `json:"jsonl_file_size,omitempty"`
	MessageCount  *int      `json:"message_count,omitempty"`
}
