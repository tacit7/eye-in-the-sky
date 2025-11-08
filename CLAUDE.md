# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **Eye in the Sky Claude Code MCP Server** - a Model Context Protocol (MCP) server that provides real-time visibility and control over multiple concurrent Claude Code instances. The system tracks AI agents working across different git worktrees and provides a centralized TUI dashboard showing what each agent is currently doing.

## Architecture

### Core Components
- **MCP Server**: Go-based server implementing Model Context Protocol for Claude Code integration
- **TUI Dashboard**: Terminal UI (Bubble Tea) providing real-time agent visibility and control
- **Database**: SQLite for tracking agents, actions, and git commits
- **Integration Layer**: MCP tools interface for seamless Claude Code integration

### Technology Stack
- **Backend**: Go 1.21+
- **TUI**: Bubble Tea framework with Lipgloss for styling
- **Database**: SQLite 3 with `github.com/mattn/go-sqlite3`
- **MCP**: Official Go SDK `github.com/modelcontextprotocol/go-sdk/mcp`

### Project Structure
```
eye-in-the-sky/
├── cmd/
│   ├── server/main.go             # MCP server executable
│   └── eye-ui/main.go             # TUI dashboard executable
├── internal/
│   ├── mcp/                       # MCP server implementation
│   ├── ui/                        # TUI dashboard (Bubble Tea)
│   ├── database/                  # SQLite connection and queries
│   ├── utils/                     # Utility functions
│   └── window/                    # Window management utilities
└── tests/                         # Go test files

### Database Location
SQLite database: `~/.config/eye-in-the-sky/agents.db` (created at runtime)
```

## Key Concepts

### Agent Management
- Each Claude Code instance gets a unique UUID (e.g., "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
- Agent IDs are auto-generated as UUIDs when you call i-start-session
- All agents are worktree-based (git repository tracking)
- Agent states and lifecycle:
  - `active`: Session is ongoing, ready for work
  - `idle`: Session paused, waiting for next task
  - `working`: Currently executing a task
  - `completed`: Session is FULLY OVER (use only when ending entire session, not for completing individual features)
  - `failed`: Session ended with error
- Metadata tracking: creation time, git worktree path, feature description, current task, session ID

### Action Logging
- Claude Code hooks automatically log all tool execution via PreToolUse/PostToolUse events
- Hook system manages agent status updates (idle on Stop, completed on SessionEnd)
- Git commits are tracked separately with hashes and messages

### MCP Integration
The system exposes 22 MCP tools for Claude Code integration:

**Session Lifecycle (4 tools):**
- `i-start-session` - Start new session and receive UUID agent_id
- `i-end-session` - Complete agent session
- `i-save-context` - Save session state for resumption
- `i-instructions` - Get detailed help and usage instructions

**Git & Window (2 tools):**
- `i-commits` - Track git commits
- `i-window` - Get current active window info (macOS)

**Session Data (2 tools):**
- `i-note-add` - Add note to session
- `i-log-compaction` - Log conversation compaction and backup JSONL file

**Persona Management (1 tool):**
- `i-snapshot-expertise` - Save agent expertise as reusable persona

**Todo/Task Management (12 tools):**
- `i-todo-create/annotate/start/done/status/tag/list/search/delete/reindex/vacuum/project-sync`

**Utility (1 tool):**
- `i-speak` - Text-to-speech with premium macOS voices

**Note:** The following tools are handled automatically by hooks and are NOT exposed to Claude:
- `i-update-status` - Hooks update status on Stop/SessionEnd events
- `i-log` - Hooks log all tool calls via PreToolUse/PostToolUse
- `i-action` - Redundant with hook logging
- `i-log-session-cost` - Hook logging captures token usage
- `i-update-description` - Hooks handle this automatically

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
go test ./internal/ui

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

### Logging

Log files are written to `~/.config/eye-in-the-sky/`:

- **`tui.log`**: TUI Dashboard debug output
  - Contains all debug messages from the TUI application
  - Truncated on startup to clear old logs
  - Configured with timestamps and source file names

- **`cc_usage.log`**: Claude Code usage data (for future use)
  - Reserved for CloudCode usage tracking and metrics

## Multi-Agent Workflow

### Agent Registration

Start a new session with your Claude Code session ID:
```
i-start-session({
  "session_id": "your-claude-code-session-id",
  "description": "Working on user authentication",
  "worktree_path": "/path/to/worktree" // optional
})
```
The system will return a UUID agent ID for you to use in subsequent calls.

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
- `id`: UUID identifier (auto-generated)
- `status`: Current agent status
- `source`: Always "worktree"
- `created_at/updated_at`: Timestamps
- `git_worktree_path`: Path to git worktree
- `feature_description`: High-level feature being worked on
- `current_task`: Specific current task
- `last_activity_at`: When agent last reported activity
- `session_id`: Associated session identifier
- `parent_agent_id`: Parent agent if this is a subagent

### Actions Table
- Links to agents via `agent_id`
- Tracks all agent activities with timestamps
- Categorizes actions by type
- Stores human-readable descriptions and JSON details

### Commits Table
- Links to agents via `agent_id`
- Tracks git commit hashes and messages
- Associates commits with agent sessions

### Logs Table
- Links to sessions via `session_id`
- Simple log storage for hook execution and MCP calls
- Fields: `id`, `session_id`, `type`, `message`, `timestamp`
- Type values: "action" (PreToolUse), "info" (PostToolUse), "debug", "error"
- Used by `i-log` CLI command and Claude Code hooks
- Separate from structured `session_logs` table (which has more fields)

## TUI Dashboard Features

### Agent List View (Overview Page)
- Real-time status of all active agents
- Visual indicators for agent health (color-coded status)
- Hierarchical display with green │ for subagents
- Grouped by parent-child relationships
- j/k navigation, enter to view details
- This is the main overview page showing all agents

### Agent Detail View (Individual Agent Page - Tabbed Interface)
- **[O]verview**: Agent info, recent commits, and notes summary
- **[C]ommits**: Split-pane view of commits with diff details
- **[L]ogs**: Session logs with timestamp and type filtering
- **[N]otes**: Session notes with creation timestamps
- **[A]ctions**: All agent actions with descriptions
- **[T]asks**: Agent tasks filtered by session, sorted by priority (H>M>L), deleted tasks at bottom
- **[P]roject tickets**: Project-specific tasks (Note: Known issue - may not respond to keypress)
- Keyboard navigation: o/c/l/n/a/t/p for tabs, j/k for items, h/l for scrolling

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

- TUI should render within 100ms and update smoothly at 60fps
- MCP tool calls must respond within 500ms
- Database operations should be atomic to prevent corruption
- System should maintain 99% uptime during development sessions
- Gracefully handle agent disconnections and database lock contention

## Development Workflow

### Task Management
- **Eye in the Sky MCP tools** (`i-todo-create`, `i-todo-start`, `i-todo-done`, etc.) are the primary task tracking system
- **TodoWrite** mirrors Eye in the Sky tasks for current session visibility in Claude Code
- Annotate tasks with progress using `i-todo-annotate` as work happens
- Tasks remain in their current state until explicitly moved to the next state
- Only the user decides when work is complete

### Information Tracking
- Use **logs** for tracking session information by default
- Use **notes** (`i-note-add`) when the user explicitly requests them
- Both logs and notes are persistent and searchable across sessions

### Eye-in-the-Sky MCP Integration
- Use `i-start-session` to begin work and receive an agent_id
- Use `i-todo-create`, `i-todo-start`, `i-todo-done` for task lifecycle management
- Use `i-commits` to track git commits
- Use `i-end-session` when work is complete

### Work Flow Pattern
1. Start session with `i-start-session`
2. Create tasks with `i-todo-create`
3. Work on code changes
4. Annotate tasks with progress using `i-todo-annotate`
5. Update TodoWrite to reflect task state
6. Commit code with meaningful messages
7. Track commits with `i-commits`
8. Mark tasks done with `i-todo-done`
9. End session with `i-end-session`