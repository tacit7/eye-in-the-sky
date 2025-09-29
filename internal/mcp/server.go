package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// Server represents the MCP server
type Server struct {
	db    *database.DB
	tools *Tools
	mcp   *mcp.Server
}

// NewServer creates a new MCP server instance
func NewServer(db *database.DB) *Server {
	// Configure log to output to stderr for MCP compatibility
	log.SetOutput(os.Stderr)

	tools := NewTools(db)

	// Create MCP server with implementation
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "eye-in-the-sky",
		Version: "1.0.0",
	}, nil)

	s := &Server{
		db:    db,
		tools: tools,
		mcp:   mcpServer,
	}

	// Register all tools
	s.registerTools()

	return s
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	log.Println("🔧 MCP Server starting...")
	log.Println("📊 Dashboard available at: http://localhost:8080")

	// Run the MCP server on stdio transport (this handles JSON-RPC over stdin/stdout)
	err := s.mcp.Run(ctx, &mcp.StdioTransport{})
	if err != nil {
		log.Printf("🔧 MCP Server shutting down: %v", err)
	} else {
		log.Println("🔧 MCP Server shutting down...")
	}

	return err
}

// registerTools registers all available tools with the MCP server
func (s *Server) registerTools() {
	// Register agent tools using the generic AddTool function
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "register_agent",
		Description: "Register a new Claude Code agent",
	}, s.handleRegisterAgent)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "register_claude_desktop_agent",
		Description: "Register a new Claude Desktop agent",
	}, s.handleRegisterDesktopAgent)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "update_status",
		Description: "Update agent status and current task",
	}, s.handleUpdateStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "log_action",
		Description: "Log agent activities",
	}, s.handleLogAction)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "end_session",
		Description: "Complete agent session",
	}, s.handleEndSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "help",
		Description: "Get detailed help and usage instructions",
	}, s.handleHelp)
}

// Tool handlers using the generic AddTool pattern
func (s *Server) handleRegisterAgent(ctx context.Context, req *mcp.CallToolRequest, args RegisterAgentArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.RegisterAgent(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Agent registered successfully"},
		},
	}, result, nil
}

func (s *Server) handleRegisterDesktopAgent(ctx context.Context, req *mcp.CallToolRequest, args RegisterDesktopAgentArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.RegisterDesktopAgent(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Claude Desktop agent registered successfully"},
		},
	}, result, nil
}

func (s *Server) handleUpdateStatus(ctx context.Context, req *mcp.CallToolRequest, args UpdateStatusArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.UpdateStatus(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Status updated successfully"},
		},
	}, result, nil
}

func (s *Server) handleLogAction(ctx context.Context, req *mcp.CallToolRequest, args LogActionArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.LogAction(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Action logged successfully"},
		},
	}, result, nil
}

func (s *Server) handleEndSession(ctx context.Context, req *mcp.CallToolRequest, args EndSessionArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.EndSession(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Session ended successfully"},
		},
	}, result, nil
}

func (s *Server) handleHelp(ctx context.Context, req *mcp.CallToolRequest, args HelpArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.Help(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Help information retrieved successfully"},
		},
	}, result, nil
}

// HandleTool provides compatibility for dashboard server to call tools directly
func (s *Server) HandleTool(toolName string, argsJSON []byte) (interface{}, error) {
	// This method provides backward compatibility for the dashboard server
	// It routes tool calls to the appropriate handlers
	switch toolName {
	case "register_agent":
		var args RegisterAgentArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.RegisterAgent(args)

	case "register_claude_desktop_agent":
		var args RegisterDesktopAgentArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.RegisterDesktopAgent(args)

	case "update_status":
		var args UpdateStatusArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.UpdateStatus(args)

	case "log_action":
		var args LogActionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.LogAction(args)

	case "end_session":
		var args EndSessionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.EndSession(args)

	case "bring_window_front":
		var args BringWindowFrontArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.BringWindowFront(args)

	case "help":
		var args HelpArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.Help(args)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}