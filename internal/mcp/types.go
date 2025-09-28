package mcp

// RegisterAgentArgs represents the arguments for register_agent tool
type RegisterAgentArgs struct {
	AgentID      *string `json:"agent_id,omitempty"`
	Description  string  `json:"description"`
	WorktreePath *string `json:"worktree_path,omitempty"`
	ProjectName  *string `json:"project_name,omitempty"`
}

type RegisterAgentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RegisterDesktopAgentArgs represents the arguments for register_claude_desktop_agent tool
type RegisterDesktopAgentArgs struct {
	AgentID     *string `json:"agent_id,omitempty"`
	Description string  `json:"description"`
	ProjectName string  `json:"project_name"`
	WindowID    *string `json:"window_id,omitempty"`
}

type RegisterDesktopAgentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateStatusArgs represents the arguments for update_status tool
type UpdateStatusArgs struct {
	AgentID     string  `json:"agent_id"`
	Status      string  `json:"status"`
	CurrentTask *string `json:"current_task,omitempty"`
}

type UpdateStatusResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LogActionArgs represents the arguments for log_action tool
type LogActionArgs struct {
	AgentID     string  `json:"agent_id"`
	ActionType  string  `json:"action_type"`
	Description string  `json:"description"`
	Details     *string `json:"details,omitempty"`
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
	AgentID     string  `json:"agent_id"`
	Summary     *string `json:"summary,omitempty"`
	FinalStatus *string `json:"final_status,omitempty"`
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

