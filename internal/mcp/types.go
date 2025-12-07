package mcp

import "github.com/tacit7/eye-in-the-sky/internal/database"

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
	ParentID   string `json:"parent_id" jsonschema:"description:Parent entity identifier (session, agent, context, etc)"`
	ParentType string `json:"parent_type" jsonschema:"description:Parent entity type (sessions, agents, contexts, etc)"`
	Body       string `json:"body" jsonschema:"description:Note content"`
}

type AddNoteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetNoteArgs represents the arguments for i-note-get tool
type GetNoteArgs struct {
	NoteID string `json:"note_id" jsonschema:"description:Note ID to retrieve"`
}

type GetNoteResult struct {
	NoteID     string `json:"note_id"`
	ParentID   string `json:"parent_id"`
	ParentType string `json:"parent_type"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
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

// UpdateFeatureDescriptionArgs represents the arguments for i-update-description tool
type UpdateFeatureDescriptionArgs struct {
	AgentID            string `json:"agent_id" jsonschema:"description:Agent UUID identifier"`
	FeatureDescription string `json:"feature_description" jsonschema:"description:New feature description for the session"`
}

type UpdateFeatureDescriptionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}


// ISpeakArgs represents the arguments for i-speak tool
type ISpeakArgs struct {
	Message string  `json:"message" jsonschema:"description:Message to speak aloud"`
	Voice   *string `json:"voice,omitempty" jsonschema:"description:Premium voice to use (Ava, Isha, Lee, Jamie, Serena). Defaults to Ava"`
	Rate    *int    `json:"rate,omitempty" jsonschema:"description:Speaking rate in words per minute (90-450). Defaults to 200"`
}

type ISpeakResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	VoiceUsed string `json:"voice_used"`
}

// Subagent Prompts Tool Types

// CreatePromptArgs represents the arguments for i-prompt-create tool
type CreatePromptArgs struct {
	Name        string `json:"name" jsonschema:"description:Human-readable name"`
	Slug        string `json:"slug" jsonschema:"description:URL-friendly identifier (kebab-case)"`
	Description string `json:"description,omitempty" jsonschema:"description:What this prompt does"`
	PromptText  string `json:"prompt_text" jsonschema:"description:The actual prompt template"`
	ProjectID   string `json:"project_id,omitempty" jsonschema:"description:Project ID for project-scoped prompt (omit for global)"`
	Tags        string `json:"tags,omitempty" jsonschema:"description:Comma-separated tags"`
	CreatedBy   string `json:"created_by,omitempty" jsonschema:"description:Who created this prompt"`
}

// GetPromptArgs represents the arguments for i-prompt-get tool
type GetPromptArgs struct {
	ID          string `json:"id,omitempty" jsonschema:"description:Prompt ID"`
	Slug        string `json:"slug,omitempty" jsonschema:"description:Prompt slug"`
	ProjectID   string `json:"project_id,omitempty" jsonschema:"description:Project context for slug lookup (checks project-scoped first, falls back to global)"`
	IncludeText bool   `json:"include_text,omitempty" jsonschema:"description:Include prompt_text in response (default: true)"`
}

// ListPromptsArgs represents the arguments for i-prompt-list tool
type ListPromptsArgs struct {
	ProjectID   string   `json:"project_id,omitempty" jsonschema:"description:Filter by project ID"`
	Active      *bool    `json:"active,omitempty" jsonschema:"description:Filter by active status"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description:Filter by tags"`
	Resolve     bool     `json:"resolve,omitempty" jsonschema:"description:Deduplicate by slug (project overrides global)"`
	IncludeText bool     `json:"include_text,omitempty" jsonschema:"description:Include prompt_text in results (default: false for list)"`
	Limit       int      `json:"limit,omitempty" jsonschema:"description:Maximum number of results"`
	Offset      int      `json:"offset,omitempty" jsonschema:"description:Offset for pagination"`
}

// UpdatePromptArgs represents the arguments for i-prompt-update tool
type UpdatePromptArgs struct {
	ID              string  `json:"id" jsonschema:"description:Prompt ID to update"`
	ExpectedVersion int     `json:"expected_version" jsonschema:"description:Expected version for optimistic locking"`
	Name            *string `json:"name,omitempty" jsonschema:"description:New name"`
	Slug            *string `json:"slug,omitempty" jsonschema:"description:New slug (kebab-case)"`
	Description     *string `json:"description,omitempty" jsonschema:"description:New description"`
	PromptText      *string `json:"prompt_text,omitempty" jsonschema:"description:New prompt text (auto-increments version)"`
	ProjectID       *string `json:"project_id,omitempty" jsonschema:"description:New project ID"`
	Tags            *string `json:"tags,omitempty" jsonschema:"description:New tags"`
	Active          *bool   `json:"active,omitempty" jsonschema:"description:New active status"`
}

// DeletePromptArgs represents the arguments for i-prompt-delete tool
type DeletePromptArgs struct {
	ID         string `json:"id" jsonschema:"description:Prompt ID to delete"`
	HardDelete bool   `json:"hard_delete,omitempty" jsonschema:"description:Permanently delete (default: soft delete by setting active=0)"`
}

// PromptResult represents the result for prompt operations
type PromptResult struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Prompt  *database.SubagentPrompt  `json:"prompt,omitempty"`
}

// ListPromptsResult represents the result for list operation
type ListPromptsResult struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Prompts []database.SubagentPrompt `json:"prompts,omitempty"`
	Count   int              `json:"count"`
}

// ImportAgentsArgs represents the arguments for i-agent-import tool
type ImportAgentsArgs struct {
	AgentsDir  string `json:"agents_dir,omitempty" jsonschema:"description:Path to agents directory (default: .claude/agents)"`
	ProjectID  string `json:"project_id,omitempty" jsonschema:"description:Project ID to scope imported agents (optional)"`
	Overwrite  bool   `json:"overwrite,omitempty" jsonschema:"description:Overwrite existing prompts with same slug (default: false)"`
	CreatedBy  string `json:"created_by,omitempty" jsonschema:"description:Who is importing (optional)"`
}

// ImportAgentsResult represents the result of agent import operation
type ImportAgentsResult struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	Imported []string `json:"imported,omitempty"`
	Skipped  []string `json:"skipped,omitempty"`
	Errors   []string `json:"errors,omitempty"`
	Count    int      `json:"count"`
}

// ChatSendArgs represents the arguments for i-chat-send tool
type ChatSendArgs struct {
	ChannelID    string  `json:"channel_id" jsonschema:"description:Channel ID to send message to"`
	SessionID    string  `json:"session_id" jsonschema:"description:Session ID of the sender"`
	Body         string  `json:"body" jsonschema:"description:Message body text"`
	SenderRole   *string `json:"sender_role,omitempty" jsonschema:"description:Sender role (default: 'agent')"`
	RecipientRole *string `json:"recipient_role,omitempty" jsonschema:"description:Recipient role (default: 'user')"`
	Provider     *string `json:"provider,omitempty" jsonschema:"description:Provider name (default: 'claude')"`
}

// ChatSendResult represents the result for i-chat-send tool
type ChatSendResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	MessageID string `json:"message_id,omitempty"`
}

// ProjectAddArgs represents the arguments for i-project-add tool
type ProjectAddArgs struct {
	Name      string  `json:"name" jsonschema:"description:Project name (required)"`
	Slug      *string `json:"slug,omitempty" jsonschema:"description:URL-friendly slug"`
	Path      *string `json:"path,omitempty" jsonschema:"description:Local filesystem path"`
	RemoteURL *string `json:"remote_url,omitempty" jsonschema:"description:Git remote URL"`
	GitRemote *string `json:"git_remote,omitempty" jsonschema:"description:Git remote name (e.g., origin)"`
	RepoURL   *string `json:"repo_url,omitempty" jsonschema:"description:Repository URL"`
	Branch    *string `json:"branch,omitempty" jsonschema:"description:Git branch name"`
	Active    *bool   `json:"active,omitempty" jsonschema:"description:Active status (default: true)"`
}

// ProjectAddResult represents the result for i-project-add tool
type ProjectAddResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ProjectID string `json:"project_id,omitempty"`
}

// SpawnAgentArgs represents the arguments for i-spawn-agent tool
type SpawnAgentArgs struct {
	Instructions    string  `json:"instructions" jsonschema:"description:Task instructions for the agent (required)"`
	Model           *string `json:"model,omitempty" jsonschema:"description:Model to use (haiku, sonnet, opus). Default: haiku"`
	ProjectPath     *string `json:"project_path,omitempty" jsonschema:"description:Working directory. Default: current directory"`
	SkipPermissions *bool   `json:"skip_permissions,omitempty" jsonschema:"description:Skip permission prompts (default: true)"`
	Background      *bool   `json:"background,omitempty" jsonschema:"description:Run agent in background (default: false)"`
	ParentAgentID   *string `json:"parent_agent_id,omitempty" jsonschema:"description:Parent agent ID for tracking hierarchy"`
	ParentSessionID *string `json:"parent_session_id,omitempty" jsonschema:"description:Parent session ID for tracking hierarchy"`
}

// SpawnAgentResult represents the result for i-spawn-agent tool
type SpawnAgentResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	CommandFile string `json:"command_file,omitempty"`
}

// SpawnClaudeArgs represents the arguments for spawning a Claude Code process
type SpawnClaudeArgs struct {
	Prompt      string  `json:"prompt" jsonschema:"description:Prompt to send to Claude Code"`
	Model       string  `json:"model" jsonschema:"description:Model to use (haiku, sonnet, opus)"`
	ProjectPath *string `json:"project_path,omitempty" jsonschema:"description:Working directory for Claude Code"`
}

// SpawnClaudeResult represents the result of spawning a Claude Code process
type SpawnClaudeResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	PID       int    `json:"pid,omitempty"`
}
