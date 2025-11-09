package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/todo"
)

// Handler wraps the todo service for MCP command handling.
type Handler struct {
	svc *todo.Service
}

// NewHandler creates a new MCP handler for todo commands.
func NewHandler(service *todo.Service) *Handler {
	return &Handler{svc: service}
}

// CommandHandler is the signature for all MCP command handlers.
type CommandHandler func(ctx context.Context, args json.RawMessage) (interface{}, error)

// Registry maps command names to their handlers.
type Registry struct {
	handlers map[string]CommandHandler
}

// NewRegistry creates a new command registry with all handlers registered.
func NewRegistry(handler *Handler) *Registry {
	return &Registry{
		handlers: map[string]CommandHandler{
			"i-todo-create":       handler.HandleCreate,
			"i-todo-annotate":     handler.HandleAnnotate,
			"i-todo-start":        handler.HandleStart,
			"i-todo-done":         handler.HandleDone,
			"i-todo-status":       handler.HandleStatus,
			"i-todo-tag":          handler.HandleTag,
			"i-todo-list":         handler.HandleList,
			"i-todo-list-agent":   handler.HandleListAgent,
			"i-todo-list-session": handler.HandleListSession,
			"i-todo-search":       handler.HandleSearch,
			"i-todo-delete":       handler.HandleDelete,
			"i-todo-reindex":      handler.HandleReindex,
			"i-todo-vacuum":       handler.HandleVacuum,
			"i-todo-project-sync": handler.HandleProjectSync,
		},
	}
}

// Execute runs a command by name with the given arguments.
func (r *Registry) Execute(ctx context.Context, command string, args json.RawMessage) (interface{}, error) {
	handler, ok := r.handlers[command]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", command)
	}
	return handler(ctx, args)
}

// GetHandlers returns the registered handlers map.
func (r *Registry) GetHandlers() map[string]CommandHandler {
	return r.handlers
}
