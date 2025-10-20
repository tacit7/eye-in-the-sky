# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **Claude Code Multi-Agent Management System** (Eye in the Sky) - a developer tool that provides real-time visibility and control over multiple concurrent Claude Code instances. The system tracks AI agents working across different git worktrees and provides a centralized dashboard showing what each agent is currently doing.

## Architecture

### Core Components
- **MCP Server**: Go-based server implementing Model Context Protocol for Claude Code integration
- **Web Dashboard**: HTTP server providing real-time agent visibility at localhost:8080
- **Database**: SQLite for tracking agents, actions, and git commits
- **Integration Layer**: MCP tools interface for seamless Claude Code integration

### Technology Stack
- **Backend**: Go 1.21+ with standard library `net/http` and `html/template`
- **Database**: SQLite 3 with `github.com/mattn/go-sqlite3`
- **Frontend**: HTML5, Bootstrap 5, vanilla JavaScript
- **MCP**: Official Go SDK `github.com/modelcontextprotocol/go-sdk/mcp`

### Project Structure
```
claude-code-mcp/
├── main.go                        # Entry point
├── cmd/server/main.go             # Main server executable
├── internal/
│   ├── mcp/                       # MCP server implementation
│   ├── dashboard/                 # HTTP server for web dashboard
│   ├── database/                  # SQLite connection and queries
│   └── utils/                     # Utility functions
├── web/
│   ├── templates/                 # HTML templates
│   └── static/                    # CSS and JavaScript
└── tests/                         # Go test files

### Database Location
SQLite database: `~/.config/eye-in-the-sky/agents.db` (created at runtime)
```

## Key Concepts

### Agent Management
- Each Claude Code instance gets a unique 8-character hash ID (git-style like "a3f7d2e1")
- Agent IDs are auto-generated if not provided using SHA1-based git-style hashes
- Two agent types: "worktree" (git-based) and "desktop" (Claude Desktop)
- Agent states: "active", "idle", "working", "completed", "failed"
- Metadata tracking: creation time, git worktree path, feature description, current task, window ID (for desktop agents)

### Action Logging
- All major Claude Code activities are logged with timestamps
- Action types: "task_start", "file_operation", "git_commit", "status_update"
- Git commits are tracked separately with hashes and messages

### MCP Integration
The system exposes these MCP tools for Claude Code integration:
- `register_agent(agent_id?, description, worktree_path?)` - Register new worktree agent
- `register_claude_desktop_agent(agent_id?, description, project_name, window_id?)` - Register new desktop agent
- `update_status(agent_id, status, current_task?)` - Update agent status
- `log_action(agent_id, action_type, description, details?)` - Log agent activity
- `log_commits(agent_id, commit_hashes[], commit_messages?)` - Track git commits
- `end_session(agent_id, summary?, final_status?)` - Complete agent session
- `help(tool?)` - Get detailed help and usage instructions

## Development Commands

### Building and Running
```bash
# Build the application
go build -o bin/eye-in-the-sky ./cmd/server

# Run with default settings (database at ~/.config/eye-in-the-sky/agents.db)
./bin/eye-in-the-sky
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/database
go test ./internal/mcp
go test ./internal/dashboard

# Run with verbose output
go test -v ./...
```

### Database Management
```bash
# The SQLite database is created automatically at runtime
# Location: ~/.config/eye-in-the-sky/agents.db

# To reset the database, simply delete the file
rm ~/.config/eye-in-the-sky/agents.db

# Database schema is initialized on first run
```

## Multi-Agent Workflow

### Agent Registration

#### For Git Worktree Agents (Claude Code):
```
"Register yourself for working on user authentication in /path/to/worktree"
```
The system will auto-generate a git-style hash ID like `a3f7d2e1`.

#### For Claude Desktop Agents:
```
"Register as Claude Desktop agent working on MyApp project"
```
Optionally include window ID for window management:
```
"Register as Claude Desktop agent for MyApp project with window ID win_12345"
```

### Status Updates
Claude instances should periodically update their status:
- When starting new tasks
- When completing file operations
- When making git commits
- When changing status (active/idle/working)

### Session Completion
When finishing work:
```
"End session, log these commits: abc123f, def456a with summary: Completed user authentication feature"
```

## Database Schema

### Agents Table
- `id`: 8-character hash identifier (auto-generated if not provided)
- `status`: Current agent status
- `source`: Agent type ("worktree" or "desktop")
- `created_at/updated_at`: Timestamps
- `git_worktree_path`: Path to git worktree (worktree agents only)
- `feature_description`: High-level feature being worked on
- `current_task`: Specific current task
- `last_activity_at`: When agent last reported activity
- `window_id`: Claude Desktop window identifier (desktop agents only)

### Actions Table
- Links to agents via `agent_id`
- Tracks all agent activities with timestamps
- Categorizes actions by type
- Stores human-readable descriptions and JSON details

### Commits Table
- Links to agents via `agent_id`
- Tracks git commit hashes and messages
- Associates commits with agent sessions

## Dashboard Features

### Agent Overview (localhost:8080)
- Real-time status of all active agents
- Visual indicators for agent health (Green/Yellow/Red)
- Quick access to agent details
- Summary statistics for active sessions

### Agent Details (localhost:8080/agent/<id>)
- Complete activity timeline
- All git commits made during session
- Current status and task information
- Session management actions

## Development Guidelines

### Code Standards
- Follow Go conventions and PEP 8 standards where applicable
- Use structured logging for debugging
- Implement proper error handling with context
- Write tests for all public functions
- Use meaningful commit messages

### MCP Tool Development
- All MCP tools should respond within 500ms
- Validate input parameters thoroughly
- Provide clear success/error messages
- Handle concurrent agent updates gracefully
- Maintain atomic database operations

### Database Operations
- Use transactions for multi-table operations
- Implement proper error handling and rollback
- Validate foreign key constraints
- Use prepared statements to prevent SQL injection
- Handle database locking appropriately for concurrent access

## Performance Considerations

- Dashboard should load within 2 seconds for up to 100 active agents
- MCP tool calls must respond within 500ms
- Database operations should be atomic to prevent corruption
- System should maintain 99% uptime during development sessions
- Gracefully handle agent disconnections and network issues