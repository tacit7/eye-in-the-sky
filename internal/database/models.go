package database

import "time"

// Agent represents a Claude Code instance
type Agent struct {
	ID                  string     `json:"id"`
	Status              string     `json:"status"`
	Source              string     `json:"source"`
	Description         *string    `json:"description,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	GitWorktreePath     *string    `json:"git_worktree_path,omitempty"`
	FeatureDescription  *string    `json:"feature_description,omitempty"`
	CurrentTask         *string    `json:"current_task,omitempty"`
	LastActivityAt      *time.Time `json:"last_activity_at,omitempty"`
	WindowID            *string    `json:"window_id,omitempty"`
	TerminalApplication *string    `json:"terminal_application,omitempty"`
	ProjectName         *string    `json:"project_name,omitempty"`
	SessionID           *string    `json:"session_id,omitempty"`
	PersonaID           *string    `json:"persona_id,omitempty"`
	ParentAgentID       *string    `json:"parent_agent_id,omitempty"`
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
	// Removed SourceDesktop - only supporting worktree agents now
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
	Title     string    `json:"title"`
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

// Compaction represents a conversation compaction event (backup/snapshot)
type Compaction struct {
	ID            int       `json:"id"`
	AgentID       string    `json:"agent_id"`
	SessionID     string    `json:"session_id"` // The session that was compacted
	CompactedAt   time.Time `json:"compacted_at"`
	Summary       *string   `json:"summary,omitempty"`
	JsonlFilePath *string   `json:"jsonl_file_path,omitempty"`
	JsonlFileSize *int64    `json:"jsonl_file_size,omitempty"`
	MessageCount  *int      `json:"message_count,omitempty"`
}

// SessionContext represents a saved session checkpoint
type SessionContext struct {
	ID              int       `json:"id"`
	AgentID         string    `json:"agent_id"`
	SessionID       string    `json:"session_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CurrentPhase    *string   `json:"current_phase,omitempty"`
	OverallProgress *float32  `json:"overall_progress,omitempty"`
	PendingTasks    *string   `json:"pending_tasks,omitempty"`
	CompletedTasks  *string   `json:"completed_tasks,omitempty"`
	NextActions     *string   `json:"next_actions,omitempty"`
	Dependencies    *string   `json:"dependencies,omitempty"`
	ImportantFiles  *string   `json:"important_files,omitempty"`
	Milestones      *string   `json:"milestones,omitempty"`
	CurrentGoals    *string   `json:"current_goals,omitempty"`
	Blockers        *string   `json:"blockers,omitempty"`
	KeyDecisions    *string   `json:"key_decisions,omitempty"`
	Environment     *string   `json:"environment,omitempty"`
	Metrics         *string   `json:"metrics,omitempty"`
	AutoSave        bool      `json:"auto_save"`
	LearnedContext  *string   `json:"learned_context,omitempty"`
}

// SessionMetrics represents token usage and cost tracking for a session
type SessionMetrics struct {
	ID               int       `json:"id"`
	AgentID          string    `json:"agent_id"`
	SessionID        *string   `json:"session_id,omitempty"`
	TokensUsed       int       `json:"tokens_used"`
	TokensBudget     int       `json:"tokens_budget"`
	TokensRemaining  int       `json:"tokens_remaining"`
	InputTokens      *int      `json:"input_tokens,omitempty"`
	OutputTokens     *int      `json:"output_tokens,omitempty"`
	EstimatedCostUSD *float64  `json:"estimated_cost_usd,omitempty"`
	ModelName        *string   `json:"model_name,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	CreatedAt        time.Time `json:"created_at"`
	Notes            *string   `json:"notes,omitempty"`
}
