package mcp

// Removed RegisterAgentArgs and RegisterDesktopAgentArgs - using only StartSession now

// UpdateStatusArgs represents the arguments for update_status tool
type UpdateStatusArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description:Agent UUID identifier"`
	Status      string  `json:"status" jsonschema:"description:One of: active, working, idle, completed, failed"`
	CurrentTask *string `json:"current_task,omitempty" jsonschema:"description:Description of current task (optional)"`
}

type UpdateStatusResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LogActionArgs represents the arguments for log_action tool
type LogActionArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description:Agent UUID identifier"`
	ActionType  string  `json:"action_type" jsonschema:"description:One of: task_start, file_operation, git_commit, status_update"`
	Description string  `json:"description" jsonschema:"description:Human-readable description of the action"`
	Details     *string `json:"details,omitempty" jsonschema:"description:Additional structured information as JSON string (optional)"`
}

type LogActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LogCommitsArgs represents the arguments for log_commits tool
type LogCommitsArgs struct {
	AgentID        string   `json:"agent_id"`
	CommitHashes   []string `json:"commit_hashes"`
	CommitMessages []string `json:"commit_messages,omitempty"`
}

type LogCommitsResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// EndSessionArgs represents the arguments for end_session tool
type EndSessionArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description:Agent UUID identifier"`
	Summary     *string `json:"summary,omitempty" jsonschema:"description:Summary of work completed (optional but recommended)"`
	FinalStatus *string `json:"final_status,omitempty" jsonschema:"description:Either 'completed' or 'failed' (optional, defaults to 'completed')"`
}

type EndSessionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SyncCommitsArgs represents the arguments for sync_commits tool
type SyncCommitsArgs struct {
	AgentID string `json:"agent_id"`
	Count   *int   `json:"count,omitempty"` // Number of recent commits to sync (default: 5)
}

type SyncCommitsResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// InstructionsArgs represents the arguments for i-instructions tool
type InstructionsArgs struct {
	// No arguments needed
}

type InstructionsResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	Instructions string `json:"instructions"`
}

// GetCurrentWindowArgs represents the arguments for get_current_window tool
type GetCurrentWindowArgs struct {
	// No arguments needed - detects current window automatically
}

type GetCurrentWindowResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Application   string `json:"application,omitempty"`
	WindowTitle   string `json:"window_title,omitempty"`
	WindowID      string `json:"window_id,omitempty"`
	Position      string `json:"position,omitempty"`
	Size          string `json:"size,omitempty"`
}

// BringWindowFrontArgs represents the arguments for bring_window_front tool
type BringWindowFrontArgs struct {
	AgentID string `json:"agent_id" jsonschema:"description:Agent UUID identifier"`
}

type BringWindowFrontResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// StartSessionArgs represents the arguments for i-start-session tool
type StartSessionArgs struct {
	SessionID        string  `json:"session_id" jsonschema:"description:Claude Code session ID (mandatory)"`
	AgentDescription *string `json:"agent_description,omitempty" jsonschema:"description:Agent name/label (e.g., 'Frontend Dev Agent') (optional)"`
	Name             *string `json:"name,omitempty" jsonschema:"description:Human-readable session name (optional)"`
	Description      string  `json:"description" jsonschema:"description:What you'll be working on"`
	ProjectName      *string `json:"project_name,omitempty" jsonschema:"description:Project name (optional)"`
	WorktreePath     *string `json:"worktree_path,omitempty" jsonschema:"description:Path to git repository (optional)"`
	PersonaID        *string `json:"persona_id,omitempty" jsonschema:"description:Persona ID to load initial context from (optional)"`
	ParentAgentID    *string `json:"parent_agent_id,omitempty" jsonschema:"description:Parent agent ID if this is a subagent (optional)"`
	ParentSessionID  *string `json:"parent_session_id,omitempty" jsonschema:"description:Parent session ID if this is a subsession (optional)"`
}

type StartSessionResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	AgentID        string `json:"agent_id"`
	SessionID      string `json:"session_id"`
	InitialContext string `json:"initial_context,omitempty"` // Loaded from persona if provided
}

// AddLogArgs represents the arguments for i-log tool
type AddLogArgs struct {
	SessionID string `json:"session_id" jsonschema:"description:Session identifier"`
	Type      string `json:"type" jsonschema:"description:Log type: action, commit, error, info"`
	Message   string `json:"message" jsonschema:"description:Log message"`
}

type AddLogResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AddNoteArgs represents the arguments for i-note-add tool
type AddNoteArgs struct {
	SessionID string `json:"session_id" jsonschema:"description:Session identifier"`
	Content   string `json:"content" jsonschema:"description:Note content"`
}

type AddNoteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SetContextArgs represents the arguments for i-context-set tool
type SetContextArgs struct {
	SessionID string `json:"session_id" jsonschema:"description:Session identifier"`
	Key       string `json:"key" jsonschema:"description:Context key"`
	Value     string `json:"value" jsonschema:"description:Context value"`
}

type SetContextResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetSessionArgs represents the arguments for i-session-get tool
type GetSessionArgs struct {
	SessionID string `json:"session_id" jsonschema:"description:Session identifier"`
}

type GetSessionResult struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	AgentID   string            `json:"agent_id"`
	SessionID string            `json:"session_id"`
	Logs      []SessionLog      `json:"logs"`
	Notes     []SessionNote     `json:"notes"`
	Context   map[string]string `json:"context"`
}

type SessionLog struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type SessionNote struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// CreatePersonaArgs represents the arguments for i-persona-create tool
type CreatePersonaArgs struct {
	ID             string  `json:"id" jsonschema:"description:Unique persona identifier (e.g., 'frontend-specialist')"`
	Name           string  `json:"name" jsonschema:"description:Human-readable persona name"`
	Description    string  `json:"description" jsonschema:"description:Brief description of the persona"`
	Expertise      string  `json:"expertise" jsonschema:"description:JSON array of expertise areas"`
	InitialContext string  `json:"initial_context" jsonschema:"description:The persona's initial instructions and context"`
	PreferredTools *string `json:"preferred_tools,omitempty" jsonschema:"description:JSON array of preferred MCP tools (optional)"`
	Specialization *string `json:"specialization,omitempty" jsonschema:"description:Primary domain (e.g., 'frontend', 'backend', 'security') (optional)"`
}

type CreatePersonaResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetPersonaArgs represents the arguments for i-persona-get tool
type GetPersonaArgs struct {
	ID string `json:"id" jsonschema:"description:Persona identifier"`
}

type GetPersonaResult struct {
	Success        bool    `json:"success"`
	Message        string  `json:"message"`
	ID             string  `json:"id,omitempty"`
	Name           string  `json:"name,omitempty"`
	Description    string  `json:"description,omitempty"`
	Expertise      string  `json:"expertise,omitempty"`
	InitialContext string  `json:"initial_context,omitempty"`
	PreferredTools *string `json:"preferred_tools,omitempty"`
	Specialization *string `json:"specialization,omitempty"`
}

// ListPersonasArgs represents the arguments for i-persona-list tool
type ListPersonasArgs struct {
	Specialization *string `json:"specialization,omitempty" jsonschema:"description:Filter by specialization (optional)"`
}

type ListPersonasResult struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Personas []PersonaSummary  `json:"personas,omitempty"`
}

type PersonaSummary struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Specialization *string `json:"specialization,omitempty"`
}

// SnapshotExpertiseArgs represents arguments for capturing current agent expertise
type SnapshotExpertiseArgs struct {
	PersonaID      string  `json:"persona_id" jsonschema:"description:Unique ID for the persona (e.g., 'payment-flow-expert')"`
	PersonaName    string  `json:"persona_name" jsonschema:"description:Human-readable name"`
	ExpertiseAreas string  `json:"expertise_areas" jsonschema:"description:JSON array of expertise areas the agent has learned"`
	CurrentContext string  `json:"current_context" jsonschema:"description:The agent's current understanding and knowledge - provide detailed context about what you've learned"`
	Specialization *string `json:"specialization,omitempty" jsonschema:"description:Primary domain (optional)"`
	PreferredTools *string `json:"preferred_tools,omitempty" jsonschema:"description:JSON array of tools used (optional)"`
}

type SnapshotExpertiseResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	PersonaID string `json:"persona_id,omitempty"`
}

// GetContextArgs represents arguments for retrieving stored context
type GetContextArgs struct {
	AgentID string `json:"agent_id" jsonschema:"description:Agent ID to get context from"`
}

type GetContextResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	LearnedContext string `json:"learned_context,omitempty"` // The stored learned context
	CurrentPhase   string `json:"current_phase,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
}

// ListSessionsArgs represents the arguments for i-list-sessions tool
type ListSessionsArgs struct {
	AgentID    *string `json:"agent_id,omitempty" jsonschema:"description:Filter by agent ID (optional)"`
	ActiveOnly *bool   `json:"active_only,omitempty" jsonschema:"description:Show only active sessions (optional, default: false)"`
}

type ListSessionsResult struct {
	Success  bool             `json:"success"`
	Message  string           `json:"message"`
	Sessions []SessionSummary `json:"sessions,omitempty"`
}

type SessionSummary struct {
	ID                 string  `json:"id"`
	AgentID            string  `json:"agent_id"`
	Name               *string `json:"name,omitempty"`
	StartedAt          string  `json:"started_at"`
	EndedAt            *string `json:"ended_at,omitempty"`
	AgentStatus        string  `json:"agent_status"`
	FeatureDescription *string `json:"feature_description,omitempty"`
	CurrentTask        *string `json:"current_task,omitempty"`
	ProjectName        *string `json:"project_name,omitempty"`
	IsActive           bool    `json:"is_active"`
}

// LogCompactionArgs represents the arguments for i-log-compaction tool
type LogCompactionArgs struct {
	AgentID   string  `json:"agent_id" jsonschema:"description:Agent UUID identifier"`
	SessionID string  `json:"session_id" jsonschema:"description:Session ID that was compacted"`
	Summary   *string `json:"summary,omitempty" jsonschema:"description:Compaction summary text (optional)"`
}

type LogCompactionResult struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	CompactionID    int    `json:"compaction_id,omitempty"`
	JsonlBackupPath string `json:"jsonl_backup_path,omitempty"`
}

// LogSessionCostArgs represents the arguments for i-log-session-cost tool
type LogSessionCostArgs struct {
	AgentID          string   `json:"agent_id" jsonschema:"description:Agent UUID identifier"`
	SessionID        *string  `json:"session_id,omitempty" jsonschema:"description:Session ID (optional)"`
	TokensUsed       int      `json:"tokens_used" jsonschema:"description:Total tokens used in session"`
	TokensBudget     int      `json:"tokens_budget" jsonschema:"description:Total token budget for session"`
	TokensRemaining  int      `json:"tokens_remaining" jsonschema:"description:Remaining tokens in budget"`
	InputTokens      *int     `json:"input_tokens,omitempty" jsonschema:"description:Input tokens used (optional)"`
	OutputTokens     *int     `json:"output_tokens,omitempty" jsonschema:"description:Output tokens used (optional)"`
	EstimatedCostUSD *float64 `json:"estimated_cost_usd,omitempty" jsonschema:"description:Estimated cost in USD (optional)"`
	ModelName        *string  `json:"model_name,omitempty" jsonschema:"description:Model name (e.g., 'claude-sonnet-4-5') (optional)"`
	Notes            *string  `json:"notes,omitempty" jsonschema:"description:Additional notes about the session (optional)"`
}

type LogSessionCostResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

