package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// Server represents the MCP server
type Server struct {
	db    *database.DB
	tools *Tools
}

// NewServer creates a new MCP server instance
func NewServer(db *database.DB) *Server {
	return &Server{
		db:    db,
		tools: NewTools(db),
	}
}

// Start starts the MCP server (placeholder for actual MCP SDK integration)
func (s *Server) Start(ctx context.Context) error {
	log.Println("🔧 MCP Server starting...")

	// TODO: This is where we would integrate with the actual MCP Go SDK
	// For now, we'll implement a basic JSON-RPC like interface

	log.Println("✅ MCP Server started successfully")
	log.Println("📋 Available tools:")
	log.Println("  - register_agent")
	log.Println("  - update_status")
	log.Println("  - log_action")
	log.Println("  - log_commits")
	log.Println("  - end_session")

	// Keep running until context is cancelled
	<-ctx.Done()
	log.Println("🔧 MCP Server shutting down...")

	return nil
}

// HandleTool processes MCP tool calls (simplified implementation)
func (s *Server) HandleTool(toolName string, argsJSON []byte) (interface{}, error) {
	switch toolName {
	case "register_agent":
		var args RegisterAgentArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for register_agent: %w", err)
		}
		return s.tools.RegisterAgent(args)

	case "update_status":
		var args UpdateStatusArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for update_status: %w", err)
		}
		return s.tools.UpdateStatus(args)

	case "log_action":
		var args LogActionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for log_action: %w", err)
		}
		return s.tools.LogAction(args)

	case "log_commits":
		var args LogCommitsArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for log_commits: %w", err)
		}
		return s.tools.LogCommits(args)

	case "end_session":
		var args EndSessionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for end_session: %w", err)
		}
		return s.tools.EndSession(args)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetToolList returns the list of available tools with their schemas
func (s *Server) GetToolList() []ToolSchema {
	return []ToolSchema{
		{
			Name:        "register_agent",
			Description: "Register a new Claude Code agent instance",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "8-character agent identifier",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "What the agent is working on",
					},
					"worktree_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to git worktree",
					},
				},
				"required": []string{"agent_id", "description"},
			},
		},
		{
			Name:        "update_status",
			Description: "Update an agent's status and current task",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "Agent identifier",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Agent status (active/idle/working/completed/failed)",
						"enum":        []string{"active", "idle", "working", "completed", "failed"},
					},
					"current_task": map[string]interface{}{
						"type":        "string",
						"description": "Current task description",
					},
				},
				"required": []string{"agent_id", "status"},
			},
		},
		{
			Name:        "log_action",
			Description: "Log an action performed by the agent",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "Agent identifier",
					},
					"action_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of action",
						"enum":        []string{"task_start", "file_operation", "git_commit", "status_update"},
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Human-readable action description",
					},
					"details": map[string]interface{}{
						"type":        "string",
						"description": "Additional JSON details",
					},
				},
				"required": []string{"agent_id", "action_type", "description"},
			},
		},
		{
			Name:        "log_commits",
			Description: "Log git commits made by the agent",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "Agent identifier",
					},
					"commit_hashes": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Array of git commit hashes",
					},
					"commit_messages": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Array of commit messages",
					},
				},
				"required": []string{"agent_id", "commit_hashes"},
			},
		},
		{
			Name:        "end_session",
			Description: "End the agent session with optional summary",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "Agent identifier",
					},
					"summary": map[string]interface{}{
						"type":        "string",
						"description": "Session summary",
					},
					"final_status": map[string]interface{}{
						"type":        "string",
						"description": "Final agent status",
						"enum":        []string{"completed", "failed"},
					},
				},
				"required": []string{"agent_id"},
			},
		},
	}
}

// ToolSchema represents the schema for an MCP tool
type ToolSchema struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}