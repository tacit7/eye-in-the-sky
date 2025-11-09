package mcp

import (
	"fmt"
	"time"
)

// SaveAgentContextArgs for saving agent-specific context in markdown format
type SaveAgentContextArgs struct {
	AgentID   string `json:"agent_id"`
	ProjectID int    `json:"project_id"`
	Context   string `json:"context"` // Markdown formatted context
}

// SaveAgentContextResult response
type SaveAgentContextResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SaveSessionContextArgs for saving session context in markdown format
type SaveSessionContextArgs struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id,omitempty"` // Optional, auto-generated if empty
	Context   string `json:"context"`              // Markdown formatted context
}

// SaveSessionContextResult response
type SaveSessionContextResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// SaveAgentContext saves markdown-formatted context to agent_context table
// This is agent-specific context per project, gets overwritten on each save
func (t *Tools) SaveAgentContext(args SaveAgentContextArgs) (SaveAgentContextResult, error) {
	if args.AgentID == "" {
		return SaveAgentContextResult{
			Success: false,
			Message: "agent_id is required",
		}, nil
	}

	if args.ProjectID == 0 {
		return SaveAgentContextResult{
			Success: false,
			Message: "project_id is required",
		}, nil
	}

	if args.Context == "" {
		return SaveAgentContextResult{
			Success: false,
			Message: "context cannot be empty",
		}, nil
	}

	// Validate agent exists
	_, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return SaveAgentContextResult{
			Success: false,
			Message: fmt.Sprintf("Agent not found: %s", args.AgentID),
		}, nil
	}

	// Save or update agent context
	query := `
		INSERT INTO agent_context (agent_id, project_id, context, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(agent_id, project_id) DO UPDATE SET
			context = excluded.context,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = t.db.Exec(query, args.AgentID, args.ProjectID, args.Context)
	if err != nil {
		return SaveAgentContextResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save agent context: %v", err),
		}, nil
	}

	return SaveAgentContextResult{
		Success: true,
		Message: fmt.Sprintf("Agent context saved successfully for agent %s, project %d", args.AgentID, args.ProjectID),
	}, nil
}

// SaveSessionContext saves markdown-formatted context to session_context table
// Creates a new record each time, maintaining history with timestamps
func (t *Tools) SaveSessionContext(args SaveSessionContextArgs) (SaveSessionContextResult, error) {
	if args.AgentID == "" {
		return SaveSessionContextResult{
			Success: false,
			Message: "agent_id is required",
		}, nil
	}

	if args.Context == "" {
		return SaveSessionContextResult{
			Success: false,
			Message: "context cannot be empty",
		}, nil
	}

	// Validate agent exists
	_, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return SaveSessionContextResult{
			Success: false,
			Message: fmt.Sprintf("Agent not found: %s", args.AgentID),
		}, nil
	}

	// Generate session_id if not provided
	sessionID := args.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s_%d", args.AgentID, time.Now().Unix())
	}

	// Insert new session context record
	query := `
		INSERT INTO session_context (agent_id, session_id, context, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err = t.db.Exec(query, args.AgentID, sessionID, args.Context)
	if err != nil {
		return SaveSessionContextResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save session context: %v", err),
		}, nil
	}

	return SaveSessionContextResult{
		Success:   true,
		Message:   fmt.Sprintf("Session context saved successfully for session %s", sessionID),
		SessionID: sessionID,
	}, nil
}

// LoadAgentContextArgs for loading agent context
type LoadAgentContextArgs struct {
	AgentID   string `json:"agent_id"`
	ProjectID int    `json:"project_id"`
}

// LoadAgentContextResult response
type LoadAgentContextResult struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Context   string    `json:"context,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// LoadAgentContext loads the agent context for a specific agent-project pair
func (t *Tools) LoadAgentContext(args LoadAgentContextArgs) (LoadAgentContextResult, error) {
	if args.AgentID == "" {
		return LoadAgentContextResult{
			Success: false,
			Message: "agent_id is required",
		}, nil
	}

	if args.ProjectID == 0 {
		return LoadAgentContextResult{
			Success: false,
			Message: "project_id is required",
		}, nil
	}

	query := `
		SELECT context, updated_at
		FROM agent_context
		WHERE agent_id = ? AND project_id = ?
	`

	var context string
	var updatedAt time.Time
	err := t.db.QueryRow(query, args.AgentID, args.ProjectID).Scan(&context, &updatedAt)
	if err != nil {
		return LoadAgentContextResult{
			Success: false,
			Message: fmt.Sprintf("No agent context found for agent %s, project %d", args.AgentID, args.ProjectID),
		}, nil
	}

	return LoadAgentContextResult{
		Success:   true,
		Message:   "Agent context loaded successfully",
		Context:   context,
		UpdatedAt: updatedAt,
	}, nil
}

// LoadSessionContextArgs for loading session context
type LoadSessionContextArgs struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id,omitempty"` // If empty, loads most recent
}

// LoadSessionContextResult response
type LoadSessionContextResult struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Context   string    `json:"context,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// LoadSessionContext loads session context - either specific session or most recent
func (t *Tools) LoadSessionContext(args LoadSessionContextArgs) (LoadSessionContextResult, error) {
	if args.AgentID == "" {
		return LoadSessionContextResult{
			Success: false,
			Message: "agent_id is required",
		}, nil
	}

	var query string
	var queryArgs []interface{}

	if args.SessionID != "" {
		// Load specific session
		query = `
			SELECT session_id, context, created_at, updated_at
			FROM session_context
			WHERE agent_id = ? AND session_id = ?
		`
		queryArgs = []interface{}{args.AgentID, args.SessionID}
	} else {
		// Load most recent session
		query = `
			SELECT session_id, context, created_at, updated_at
			FROM session_context
			WHERE agent_id = ?
			ORDER BY updated_at DESC
			LIMIT 1
		`
		queryArgs = []interface{}{args.AgentID}
	}

	var sessionID, context string
	var createdAt, updatedAt time.Time
	err := t.db.QueryRow(query, queryArgs...).Scan(&sessionID, &context, &createdAt, &updatedAt)
	if err != nil {
		return LoadSessionContextResult{
			Success: false,
			Message: fmt.Sprintf("No session context found for agent %s", args.AgentID),
		}, nil
	}

	return LoadSessionContextResult{
		Success:   true,
		Message:   "Session context loaded successfully",
		Context:   context,
		SessionID: sessionID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
