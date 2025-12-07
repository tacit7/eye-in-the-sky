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

	// Initialize todo service and registry using main database
	todoService := todo.NewService(db)
	handler := todo_mcp.NewHandler(todoService)
	todoRegistry := todo_mcp.NewRegistry(handler)

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
	// - i-update-status (redundant with hook logging)
	// - i-action (redundant with hook logging)
	// - i-log (redundant with hook logging)
	// - i-log-session-cost (redundant with hook logging)
	// - i-update-description (redundant, hooks handle this)
	// All are still available via HandleTool() for backward compatibility

	// mcp.AddTool(s.mcp, &mcp.Tool{
	// 	Name:        "i-update-status",
	// 	Description: "Update agent status and current task",
	// }, s.handleUpdateStatus)

	// mcp.AddTool(s.mcp, &mcp.Tool{
	// 	Name:        "i-action",
	// 	Description: "Log agent actions (task_start, file_operation, git_commit, status_update)",
	// }, s.handleLogAction)

	// mcp.AddTool(s.mcp, &mcp.Tool{
	// 	Name:        "i-log",
	// 	Description: "Add log entry to session",
	// }, s.handleAddLog)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-commits",
		Description: "Track git commits",
	}, s.handleLogCommits)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-end-session",
		Description: "Complete agent session",
	}, s.handleEndSession)

	// Deprecated: Use i-save-session-context instead
	// mcp.AddTool(s.mcp, &mcp.Tool{
	// 	Name:        "i-save-context",
	// 	Description: "Save session state for resumption (DEPRECATED - use i-save-session-context)",
	// }, s.handleSaveSessionContextOld)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-save-agent-context",
		Description: "Save agent-specific context in markdown format to agent_context table",
	}, s.handleSaveAgentContext)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-save-session-context",
		Description: "Save session context in markdown format to session_context table",
	}, s.handleSaveSessionContext)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-load-agent-context",
		Description: "Load agent-specific context from agent_context table",
	}, s.handleLoadAgentContext)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-load-session-context",
		Description: "Load session context from session_context table",
	}, s.handleLoadSessionContext)

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
		Name:        "i-note-get",
		Description: "Retrieve note by ID",
	}, s.handleGetNote)

	// mcp.AddTool(s.mcp, &mcp.Tool{
	// 	Name:        "i-log-session-cost",
	// 	Description: "Log session token usage and cost metrics",
	// }, s.handleLogSessionCost)

	// mcp.AddTool(s.mcp, &mcp.Tool{
	// 	Name:        "i-update-description",
	// 	Description: "Update session feature description",
	// }, s.handleUpdateFeatureDescription)

	// Todo Management Tools
	if s.todoRegistry != nil {
		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-create",
			Description: "Create a new task with optional priority and tags",
		}, s.handleTodoCreate)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-annotate",
			Description: "Add a markdown note to a task",
		}, s.handleTodoAnnotate)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-start",
			Description: "Move a task to doing state",
		}, s.handleTodoStart)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-done",
			Description: "Move a task to done state",
		}, s.handleTodoDone)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-status",
			Description: "Move a task to any workflow state",
		}, s.handleTodoStatus)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-tag",
			Description: "Add or remove tags from a task",
		}, s.handleTodoTag)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-list",
			Description: "Retrieve tasks with optional filters",
		}, s.handleTodoList)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-list-agent",
			Description: "Retrieve tasks filtered by agent ID",
		}, s.handleTodoListAgent)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-list-session",
			Description: "Retrieve tasks filtered by session ID",
		}, s.handleTodoListSession)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-search",
			Description: "Perform full-text search on tasks",
		}, s.handleTodoSearch)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-delete",
			Description: "Permanently delete a task",
		}, s.handleTodoDelete)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-add-session",
			Description: "Add a task to a session",
		}, s.handleTodoAddSession)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-remove-session",
			Description: "Remove a task from a session",
		}, s.handleTodoRemoveSession)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-add-session-to-tasks",
			Description: "Add all tasks from source session to target session (bulk operation)",
		}, s.handleTodoBulkAddSession)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-reindex",
			Description: "Rebuild the FTS5 search index",
		}, s.handleTodoReindex)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-vacuum",
			Description: "Run database maintenance (VACUUM and ANALYZE)",
		}, s.handleTodoVacuum)

		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "i-todo-project-sync",
			Description: "Sync workflow states from YAML definition",
		}, s.handleTodoProjectSync)
	}

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-speak",
		Description: "Speak a message aloud using macOS text-to-speech with premium voices",
	}, s.handleISpeak)

	// Subagent Prompts Tools
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-prompt-create",
		Description: "Create a new subagent prompt template (global or project-scoped)",
	}, s.handlePromptCreate)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-prompt-get",
		Description: "Retrieve a subagent prompt by slug or ID (project-aware fallback)",
	}, s.handlePromptGet)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-prompt-list",
		Description: "List subagent prompts with filtering and deduplication options",
	}, s.handlePromptList)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-prompt-update",
		Description: "Update a subagent prompt with optimistic locking",
	}, s.handlePromptUpdate)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-prompt-delete",
		Description: "Delete a subagent prompt (soft delete by default)",
	}, s.handlePromptDelete)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-agent-import",
		Description: "Import Claude Code agent definitions from .claude/agents/ directory into subagent_prompts table",
	}, s.handleAgentImport)

	// NATS Messaging Tools
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "i-nats-send",
		Description: `Send a message via NATS JetStream messaging system

IMPORTANT: Before using, start JetStream with: nats-server -js

Parameters:
- sender_id (required): Your session_id (identifies who is sending)
- receiver_id (optional): Target session_id for targeted delivery. Empty = broadcast to all agents
- subject (optional): Message subject. Auto-prefixed with 'events.' if not present. Default: events.test
- message (required): The message content

Examples:
- Broadcast: {"sender_id": "my-session-id", "message": "Deploy starting", "receiver_id": ""}
- Targeted: {"sender_id": "my-session-id", "receiver_id": "target-session-id", "message": "Review auth module"}
- Custom subject: {"sender_id": "my-session-id", "subject": "critique.ui", "message": "..."} → becomes events.critique.ui`,
	}, s.handleNATSSend)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "i-nats-listen",
		Description: `Check for new NATS messages in the instruction queue

Polls for messages sent to your session_id or broadcast messages.

Parameters:
- session_id (required): Your session ID (used to filter targeted messages)
- last_sequence (optional): Last processed sequence number. Use 0 or omit to get all new messages
- max_messages (optional): Maximum messages to fetch (default: 10)

Returns messages where receiver_id matches your session_id or is empty (broadcast).

Example:
{"session_id": "abc123", "last_sequence": 5, "max_messages": 20}`,
	}, s.handleNATSListen)

	// Chat Messaging Tool
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "i-chat-send",
		Description: `Send a message directly to a chat channel

Allows agents to proactively send messages to channels without going through stdout.

Parameters:
- channel_id (required): Channel ID to send message to
- session_id (required): Session ID of the sender
- body (required): Message body text
- sender_role (optional): Sender role (default: 'agent')
- recipient_role (optional): Recipient role (default: 'user')
- provider (optional): Provider name (default: 'claude')

Example:
{"channel_id": "channel-uuid", "session_id": "session-uuid", "body": "Task completed successfully"}`,
	}, s.handleChatSend)

	// Project Management Tools
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-project-add",
		Description: "Create a new project in Eye in the Sky for tracking agents and tasks",
	}, s.handleProjectAdd)

	// Agent Spawning Tools
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-spawn-agent",
		Description: "Spawn a new Claude Code agent with Eye in the Sky integration. Supports foreground/background execution and parent tracking.",
	}, s.handleSpawnAgent)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "i-spawn-claude",
		Description: "Spawn a Claude Code process and capture its session ID from the init message.",
	}, s.handleSpawnClaude)

	// Note: ExecuteNew, ExecuteContinue, ExecuteResume are reusable Go functions
	// for the MCP server, not exposed as tools to Claude Code users.
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

// Handler for new simplified agent context
func (s *Server) handleSaveAgentContext(ctx context.Context, req *mcp.CallToolRequest, args SaveAgentContextArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.SaveAgentContext(args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save agent context: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

// Handler for new simplified session context
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

// Handler for loading agent context
func (s *Server) handleLoadAgentContext(ctx context.Context, req *mcp.CallToolRequest, args LoadAgentContextArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.LoadAgentContext(args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load agent context: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

// Handler for loading session context
func (s *Server) handleLoadSessionContext(ctx context.Context, req *mcp.CallToolRequest, args LoadSessionContextArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.LoadSessionContext(args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load session context: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

// Commented out - AddSessionNoteArgs was in old session_context.go
// func (s *Server) handleAddSessionNote(ctx context.Context, req *mcp.CallToolRequest, args AddSessionNoteArgs) (*mcp.CallToolResult, any, error) {
// 	// For now, return a simple success message - session notes can be implemented later
// 	return &mcp.CallToolResult{
// 		Content: []mcp.Content{
// 			&mcp.TextContent{Text: "Session note added successfully"},
// 		},
// 	}, map[string]interface{}{"success": true, "message": "Session note added"}, nil
// }

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

	case "update_feature_description", "i-update-feature-description", "i-update-description":
		var args UpdateFeatureDescriptionArgs
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			return nil, err
		}
		return s.tools.UpdateFeatureDescription(args)

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

func (s *Server) handleGetNote(ctx context.Context, req *mcp.CallToolRequest, args GetNoteArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.GetNote(args)
	if err != nil {
		return nil, nil, err
	}

	message := fmt.Sprintf("Note ID: %s\nParent: %s (%s)\nCreated: %s\n\n%s",
		result.NoteID, result.ParentID, result.ParentType, result.CreatedAt, result.Body)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
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

func (s *Server) handleUpdateFeatureDescription(ctx context.Context, req *mcp.CallToolRequest, args UpdateFeatureDescriptionArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.UpdateFeatureDescription(args)
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

func (s *Server) handleISpeak(ctx context.Context, req *mcp.CallToolRequest, args ISpeakArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ISpeak(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

// Subagent Prompts Handlers

func (s *Server) handlePromptCreate(ctx context.Context, req *mcp.CallToolRequest, args CreatePromptArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.CreateSubagentPrompt(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handlePromptGet(ctx context.Context, req *mcp.CallToolRequest, args GetPromptArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.GetSubagentPrompt(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handlePromptList(ctx context.Context, req *mcp.CallToolRequest, args ListPromptsArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ListSubagentPrompts(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handlePromptUpdate(ctx context.Context, req *mcp.CallToolRequest, args UpdatePromptArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.UpdateSubagentPrompt(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handlePromptDelete(ctx context.Context, req *mcp.CallToolRequest, args DeletePromptArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.DeleteSubagentPrompt(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleNATSSend(ctx context.Context, req *mcp.CallToolRequest, args NATSSendArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.NATSSend(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleAgentImport(ctx context.Context, req *mcp.CallToolRequest, args ImportAgentsArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ImportAgents(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleNATSListen(ctx context.Context, req *mcp.CallToolRequest, args NATSListenArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.NATSListen(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleChatSend(ctx context.Context, req *mcp.CallToolRequest, args ChatSendArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ChatSend(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleProjectAdd(ctx context.Context, req *mcp.CallToolRequest, args ProjectAddArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.AddProject(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleSpawnAgent(ctx context.Context, req *mcp.CallToolRequest, args SpawnAgentArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.SpawnAgent(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleSpawnClaude(ctx context.Context, req *mcp.CallToolRequest, args SpawnClaudeArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.SpawnClaude(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleExecuteNew(ctx context.Context, req *mcp.CallToolRequest, args ExecuteNewArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ExecuteNew(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleExecuteContinue(ctx context.Context, req *mcp.CallToolRequest, args ExecuteContinueArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ExecuteContinue(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}

func (s *Server) handleExecuteResume(ctx context.Context, req *mcp.CallToolRequest, args ExecuteResumeArgs) (*mcp.CallToolResult, any, error) {
	result, err := s.tools.ExecuteResume(args)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Message},
		},
	}, result, nil
}
