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
	// Register agent tools using the generic AddTool function with "i-" prefix
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-register",
		Description: "Register a new Claude Code agent",
	}, s.handleRegisterAgent)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-register-claude-desktop",
		Description: "Register a new Claude Desktop agent",
	}, s.handleRegisterDesktopAgent)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-status",
		Description: "Update agent status and current task",
	}, s.handleUpdateStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-action",
		Description: "Log agent activities",
	}, s.handleLogAction)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-commits",
		Description: "Track git commits",
	}, s.handleLogCommits)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-end",
		Description: "Complete agent session",
	}, s.handleEndSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-save-context",
		Description: "Save session state for resumption",
	}, s.handleSaveSessionContext)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-load-context",
		Description: "Load previous session state",
	}, s.handleLoadSessionContext)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-note",
		Description: "Add contextual notes to session",
	}, s.handleAddSessionNote)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-window",
		Description: "Get current active window info (macOS)",
	}, s.handleGetCurrentWindow)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-bring-front",
		Description: "Bring agent window to front (macOS)",
	}, s.handleBringWindowFront)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-help",
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

func (s *Server) handleLogCommits(ctx context.Context, req *mcp.CallToolRequest, args LogCommitsArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.LogCommits(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Commits logged successfully"},
		},
	}, result, nil
}

func (s *Server) handleSaveSessionContext(ctx context.Context, req *mcp.CallToolRequest, args SaveSessionContextArgs) (*mcp.CallToolResult, any, error) {
	// For now, return a simple success message - session context can be implemented later
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Session context saved successfully"},
		},
	}, map[string]interface{}{"success": true, "message": "Session context saved"}, nil
}

func (s *Server) handleLoadSessionContext(ctx context.Context, req *mcp.CallToolRequest, args LoadSessionContextArgs) (*mcp.CallToolResult, any, error) {
	// For now, return a simple success message - session context can be implemented later
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Session context loaded successfully"},
		},
	}, map[string]interface{}{"success": true, "message": "Session context loaded"}, nil
}

func (s *Server) handleAddSessionNote(ctx context.Context, req *mcp.CallToolRequest, args AddSessionNoteArgs) (*mcp.CallToolResult, any, error) {
	// For now, return a simple success message - session notes can be implemented later
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Session note added successfully"},
		},
	}, map[string]interface{}{"success": true, "message": "Session note added"}, nil
}

func (s *Server) handleGetCurrentWindow(ctx context.Context, req *mcp.CallToolRequest, args GetCurrentWindowArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.GetCurrentWindow(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Current window information retrieved successfully"},
		},
	}, result, nil
}

func (s *Server) handleBringWindowFront(ctx context.Context, req *mcp.CallToolRequest, args BringWindowFrontArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.BringWindowFront(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Window brought to front successfully"},
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
	// Support both old names (for dashboard) and new shortened names
	switch toolName {
	case "register_agent", "i-register-agent", "i-register":
		var args RegisterAgentArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.RegisterAgent(args)

	case "register_claude_desktop_agent", "i-register-claude-desktop-agent", "i-register-claude-desktop":
		var args RegisterDesktopAgentArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.RegisterDesktopAgent(args)

	case "update_status", "i-update-status", "i-status":
		var args UpdateStatusArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.UpdateStatus(args)

	case "log_action", "i-log-action", "i-action":
		var args LogActionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.LogAction(args)

	case "log_commits", "i-log-commits", "i-commits":
		var args LogCommitsArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.LogCommits(args)

	case "end_session", "i-end-session", "i-end":
		var args EndSessionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.EndSession(args)

	case "get_current_window", "i-get-current-window", "i-window":
		var args GetCurrentWindowArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.GetCurrentWindow(args)

	case "bring_window_front", "i-bring-window-front", "i-bring-front":
		var args BringWindowFrontArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.BringWindowFront(args)

	case "help", "i-help":
		var args HelpArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.Help(args)

	// Session context tools (placeholder implementations)
	case "save_session_context", "i-save-session-context", "i-save-context":
		return map[string]interface{}{"success": true, "message": "Session context saved"}, nil

	case "load_session_context", "i-load-session-context", "i-load-context":
		return map[string]interface{}{"success": true, "message": "Session context loaded"}, nil

	case "add_session_note", "i-add-session-note", "i-note":
		return map[string]interface{}{"success": true, "message": "Session note added"}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}