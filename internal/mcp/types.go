package mcp

// RegisterAgentArgs represents the arguments for register_agent tool
type RegisterAgentArgs struct {
	AgentID      *string `json:"agent_id,omitempty" jsonschema:"description:Unique 8-character hex identifier (optional - auto-generated if not provided)"`
	Description  string  `json:"description" jsonschema:"description:Brief description of what the agent will work on"`
	WorktreePath *string `json:"worktree_path,omitempty" jsonschema:"description:Path to the git repository (optional)"`
	ProjectName  *string `json:"project_name,omitempty" jsonschema:"description:Name of the project being worked on (optional)"`
}

type RegisterAgentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RegisterDesktopAgentArgs represents the arguments for register_claude_desktop_agent tool
type RegisterDesktopAgentArgs struct {
	AgentID     *string `json:"agent_id,omitempty" jsonschema:"description:Unique 8-character hex identifier (optional - auto-generated if not provided)"`
	Description string  `json:"description" jsonschema:"description:Brief description of what the agent will work on"`
	ProjectName string  `json:"project_name" jsonschema:"description:Name of the project being worked on"`
	WindowID    *string `json:"window_id,omitempty" jsonschema:"description:Claude Desktop window identifier for window management (optional)"`
}

type RegisterDesktopAgentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateStatusArgs represents the arguments for update_status tool
type UpdateStatusArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description:8-character agent identifier"`
	Status      string  `json:"status" jsonschema:"description:One of: active, working, idle, completed, failed"`
	CurrentTask *string `json:"current_task,omitempty" jsonschema:"description:Description of current task (optional)"`
}

type UpdateStatusResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LogActionArgs represents the arguments for log_action tool
type LogActionArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description:8-character agent identifier"`
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
	AgentID     string  `json:"agent_id" jsonschema:"description:8-character agent identifier"`
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

// Tool represents an MCP tool descriptor
type Tool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Instructions string           `json:"instructions,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	Examples    []string          `json:"examples,omitempty"`
}

// HelpArgs represents the arguments for help tool
type HelpArgs struct {
	Tool *string `json:"tool,omitempty"`
}

type HelpResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Tools   []Tool `json:"tools,omitempty"`
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
	AgentID string `json:"agent_id"`
}

type BringWindowFrontResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

