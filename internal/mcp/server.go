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

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	log.Println("🔧 MCP Server starting...")

	// For now, this is a placeholder for actual MCP SDK integration
	// In a real implementation, this would set up the MCP protocol handlers

	log.Println("✅ MCP Server started successfully")
	log.Println("📋 Available tools:")
	log.Println("  - register_agent: Register a new Claude Code agent")
	log.Println("  - update_status: Update agent status and current task")
	log.Println("  - log_action: Log agent activities")
	log.Println("  - log_commits: Track git commits")
	log.Println("  - end_session: Complete agent session")
	log.Println("  - help: Get detailed help and usage instructions")
	log.Println("")
	log.Println("💡 Use the 'help' tool for detailed instructions:")
	log.Println("   - Get all help: {}")
	log.Println("   - Specific tool: {\"tool\": \"register_agent\"}")
	log.Println("📊 Dashboard available at: http://localhost:8080")

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

	case "sync_commits":
		var args SyncCommitsArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for sync_commits: %w", err)
		}
		return s.tools.SyncCommits(args)

	case "end_session":
		var args EndSessionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for end_session: %w", err)
		}
		return s.tools.EndSession(args)

	case "help":
		var args HelpArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments for help: %w", err)
		}
		return s.tools.Help(args)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetToolList returns a list of available MCP tools with basic info
func (s *Server) GetToolList() []Tool {
	return []Tool{
		{
			Name:        "register_agent",
			Description: "Register a new Claude Code agent",
		},
		{
			Name:        "update_status",
			Description: "Update agent status and current task",
		},
		{
			Name:        "log_action",
			Description: "Log agent activities",
		},
		{
			Name:        "log_commits",
			Description: "Track git commits",
		},
		{
			Name:        "end_session",
			Description: "Complete agent session",
		},
		{
			Name:        "help",
			Description: "Get detailed help and usage instructions",
		},
	}
}

// GetDetailedToolList returns tools with comprehensive documentation
func (s *Server) GetDetailedToolList() []Tool {
	return s.tools.getAllToolsWithHelp()
}
