package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Specialized handlers for each todo command (for MCP tool routing)

func (s *Server) handleTodoCreate(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.create", args)
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

func (s *Server) handleTodoAnnotate(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.annotate", args)
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

func (s *Server) handleTodoStart(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.start", args)
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

func (s *Server) handleTodoDone(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.done", args)
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

func (s *Server) handleTodoStatus(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.status", args)
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

func (s *Server) handleTodoTag(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.tag", args)
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

func (s *Server) handleTodoList(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.list", args)
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

func (s *Server) handleTodoSearch(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.search", args)
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

func (s *Server) handleTodoDelete(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.delete", args)
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

func (s *Server) handleTodoReindex(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.reindex", args)
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

func (s *Server) handleTodoVacuum(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.vacuum", args)
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

func (s *Server) handleTodoProjectSync(ctx context.Context, req *mcp.CallToolRequest, args json.RawMessage) (*mcp.CallToolResult, any, error) {
	result, err := s.todoRegistry.Execute(ctx, "todo.project.sync", args)
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
