package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	todo_mcp "github.com/tacit7/eye-in-the-sky/internal/todo/mcp"
)

// Callback-style handlers for MCP SDK compatibility
// These match the signature expected by mcp.AddTool with the old callback pattern

func (s *Server) handleTodoCreate(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.CreateRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-create", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoAnnotate(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.AnnotateRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-annotate", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoStart(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.StateChangeRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-start", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoDone(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.StateChangeRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-done", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoStatus(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.StatusRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-status", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoTag(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.TagRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-tag", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoList(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.ListRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-list", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoSearch(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.SearchRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-search", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoDelete(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.DeleteRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-delete", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoReindex(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.ReindexRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-reindex", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoVacuum(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.VacuumRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-vacuum", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}

func (s *Server) handleTodoProjectSync(ctx context.Context, req *mcp.CallToolRequest, args todo_mcp.ProjectSyncRequest) (*mcp.CallToolResult, any, error) {
	jsonArgs, _ := json.Marshal(args)
	result, err := s.todoRegistry.Execute(ctx, "i-todo-project-sync", jsonArgs)
	if err != nil {
		return nil, nil, err
	}
	resultJSON, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, result, nil
}
