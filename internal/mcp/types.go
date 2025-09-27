package mcp

// RegisterAgentArgs represents the arguments for register_agent tool
type RegisterAgentArgs struct {
	AgentID      string  `json:"agent_id"`
	Description  string  `json:"description"`
	WorktreePath *string `json:"worktree_path,omitempty"`
}

type RegisterAgentResult struct {
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

// Tool represents an MCP tool descriptor
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
