package mcp

// MCP Tool Request/Response types as defined in the PRD

// RegisterAgentArgs represents the arguments for register_agent tool
type RegisterAgentArgs struct {
	AgentID      string  `json:"agent_id" jsonschema:"description=8-character agent identifier"`
	Description  string  `json:"description" jsonschema:"description=What the agent is working on"`
	WorktreePath *string `json:"worktree_path,omitempty" jsonschema:"description=Path to git worktree"`
}

// RegisterAgentResult represents the response from register_agent tool
type RegisterAgentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateStatusArgs represents the arguments for update_status tool
type UpdateStatusArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description=Agent identifier"`
	Status      string  `json:"status" jsonschema:"description=Agent status (active/idle/working/completed/failed)"`
	CurrentTask *string `json:"current_task,omitempty" jsonschema:"description=Current task description"`
}

// UpdateStatusResult represents the response from update_status tool
type UpdateStatusResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LogActionArgs represents the arguments for log_action tool
type LogActionArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description=Agent identifier"`
	ActionType  string  `json:"action_type" jsonschema:"description=Type of action (task_start/file_operation/git_commit/status_update)"`
	Description string  `json:"description" jsonschema:"description=Human-readable action description"`
	Details     *string `json:"details,omitempty" jsonschema:"description=Additional JSON details"`
}

// LogActionResult represents the response from log_action tool
type LogActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LogCommitsArgs represents the arguments for log_commits tool
type LogCommitsArgs struct {
	AgentID        string   `json:"agent_id" jsonschema:"description=Agent identifier"`
	CommitHashes   []string `json:"commit_hashes" jsonschema:"description=Array of git commit hashes"`
	CommitMessages []string `json:"commit_messages,omitempty" jsonschema:"description=Array of commit messages"`
}

// LogCommitsResult represents the response from log_commits tool
type LogCommitsResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// EndSessionArgs represents the arguments for end_session tool
type EndSessionArgs struct {
	AgentID     string  `json:"agent_id" jsonschema:"description=Agent identifier"`
	Summary     *string `json:"summary,omitempty" jsonschema:"description=Session summary"`
	FinalStatus *string `json:"final_status,omitempty" jsonschema:"description=Final agent status"`
}

// EndSessionResult represents the response from end_session tool
type EndSessionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}