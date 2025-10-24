package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/todo"
	todo_mcp "github.com/tacit7/eye-in-the-sky/internal/todo/mcp"
)

// Server represents the MCP server
type Server struct {
	db          *database.DB
	tools       *Tools
	mcp         *mcp.Server
	todoService *todo.Service
	todoRegistry *todo_mcp.Registry
}

// NewServer creates a new MCP server instance
func NewServer(db *database.DB) *Server {
	// Configure log to output to stderr for MCP compatibility
	log.SetOutput(os.Stderr)

	tools := NewTools(db)

	// Initialize todo service and registry
	var todoService *todo.Service
	var todoRegistry *todo_mcp.Registry

	todoDb, err := db.OpenTodoDB()
	if err != nil {
		log.Printf("Warning: Failed to initialize todo database: %v\n", err)
	} else {
		todoService = todo.NewService(todoDb)
		handler := todo_mcp.NewHandler(todoService)
		todoRegistry = todo_mcp.NewRegistry(handler)
	}

	// Create MCP server with implementation
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "eits",
		Version: "1.0.0",
	}, nil)

	s := &Server{
		db:           db,
		tools:        tools,
		mcp:          mcpServer,
		todoService:  todoService,
		todoRegistry: todoRegistry,
	}

	// Register all tools
	s.registerTools()

	return s
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	log.Println("🔧 MCP Server starting...")

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
	// NOTE: The following tools are kept internal-only (not exposed to Claude):
	// - i-register, i-register-claude-desktop (use i-start-session instead)
	// - i-session-get, i-list-sessions (internal query tools)
	// - i-bring-front (window management - internal use)
	// - i-persona-get, i-persona-list (internal persona queries)
	// - i-load-context (internal context loading)
	// - i-context-set (removed, use i-save-context instead)
	// - i-note (deprecated stub, use i-note-add)
	// All are still available via HandleTool() for backward compatibility

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-update-status",
		Description: "Update agent status and current task",
	}, s.handleUpdateStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-action",
		Description: "Log agent actions (task_start, file_operation, git_commit, status_update)",
	}, s.handleLogAction)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-log",
		Description: "Add log entry to session",
	}, s.handleAddLog)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-commits",
		Description: "Track git commits",
	}, s.handleLogCommits)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-end-session",
		Description: "Complete agent session",
	}, s.handleEndSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-save-context",
		Description: "Save session state for resumption",
	}, s.handleSaveSessionContext)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-log-compaction",
		Description: "Log conversation compaction and backup JSONL file",
	}, s.handleLogCompaction)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-window",
		Description: "Get current active window info (macOS)",
	}, s.handleGetCurrentWindow)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-instructions",
		Description: "Get complete Eye in the Sky workflow and initialization instructions. Call this FIRST before starting a session to learn how to use the system.",
	}, s.handleInstructions)

	// POA Spec Tools - Session Management
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-start-session",
		Description: "Start a new session with agent registration. Pass your session_id (provided at start of conversation) and a description of what you'll be working on. The system will return an auto-generated agent_id to use for all subsequent calls. Call i-instructions for complete initialization workflow.",
	}, s.handleStartSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-note-add",
		Description: "Add note to session",
	}, s.handleAddNote)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-log-session-cost",
		Description: "Log session token usage and cost metrics",
	}, s.handleLogSessionCost)

	// Persona Management Tools
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-snapshot-expertise",
		Description: "Save current agent expertise as a reusable persona",
	}, s.handleSnapshotExpertise)

	// NOTE: i-persona-get and i-persona-list kept internal-only
	// Available via HandleTool() for dashboard/internal use

	// Todo Management Tools
	if s.todoRegistry != nil {
		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.create",
			Description: "Create a new task with optional priority and tags",
		}, s.handleTodoCreate)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.annotate",
			Description: "Add a markdown note to a task",
		}, s.handleTodoAnnotate)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.start",
			Description: "Move a task to doing state",
		}, s.handleTodoStart)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.done",
			Description: "Move a task to done state",
		}, s.handleTodoDone)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.status",
			Description: "Move a task to any workflow state",
		}, s.handleTodoStatus)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.tag",
			Description: "Add or remove tags from a task",
		}, s.handleTodoTag)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.list",
			Description: "Retrieve tasks with optional filters",
		}, s.handleTodoList)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.search",
			Description: "Perform full-text search on tasks",
		}, s.handleTodoSearch)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.delete",
			Description: "Permanently delete a task",
		}, s.handleTodoDelete)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.reindex",
			Description: "Rebuild the FTS5 search index",
		}, s.handleTodoReindex)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.vacuum",
			Description: "Run database maintenance (VACUUM and ANALYZE)",
		}, s.handleTodoVacuum)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "todo.project.sync",
			Description: "Sync workflow states from YAML definition",
		}, s.handleTodoProjectSync)
	}
}

// Tool handlers using the generic AddTool pattern
// Removed handleRegisterAgent and handleRegisterDesktopAgent - now using only handleStartSession

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
	result, err := s.tools.SaveSessionContext(args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save session context: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
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

func (s *Server) handleInstructions(ctx context.Context, req *mcp.CallToolRequest, args InstructionsArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.Instructions(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Instructions},
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
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleAddLog(ctx context.Context, req *mcp.CallToolRequest, args AddLogArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.AddLog(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

// HandleTool provides compatibility for dashboard server to call tools directly
func (s *Server) HandleTool(toolName string, argsJSON []byte) (interface{}, error) {
	// This method provides backward compatibility for the dashboard server
	// It routes tool calls to the appropriate handlers
	// Support both old names (for dashboard) and new shortened names
	switch toolName {
	// Removed register_agent and register_claude_desktop_agent - now using only i-start-session

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

	case "add_log", "i-add-log", "i-log":
		var args AddLogArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.AddLog(args)

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

	case "instructions", "i-instructions":
		var args InstructionsArgs
		return s.tools.Instructions(args)

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
// POA Spec Tool Handlers

func (s *Server) handleStartSession(ctx context.Context, req *mcp.CallToolRequest, args StartSessionArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.StartSession(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleAddNote(ctx context.Context, req *mcp.CallToolRequest, args AddNoteArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.AddNote(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleLogSessionCost(ctx context.Context, req *mcp.CallToolRequest, args LogSessionCostArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.LogSessionCost(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleSetContext(ctx context.Context, req *mcp.CallToolRequest, args SetContextArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.SetContext(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleGetSession(ctx context.Context, req *mcp.CallToolRequest, args GetSessionArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.GetSession(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleSnapshotExpertise(ctx context.Context, req *mcp.CallToolRequest, args SnapshotExpertiseArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.SnapshotExpertise(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleGetPersona(ctx context.Context, req *mcp.CallToolRequest, args GetPersonaArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.GetPersona(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleListPersonas(ctx context.Context, req *mcp.CallToolRequest, args ListPersonasArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ListPersonas(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleListSessions(ctx context.Context, req *mcp.CallToolRequest, args ListSessionsArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ListSessions(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleLogCompaction(ctx context.Context, req *mcp.CallToolRequest, args LogCompactionArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.LogCompaction(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}
