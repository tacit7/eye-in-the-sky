# Eye in the Sky - User Manual

**Version 1.0 | Claude Code Multi-Agent Management System**

---

## Table of Contents

1. [Overview](#overview)
2. [Installation & Setup](#installation--setup)
3. [Getting Started](#getting-started)
4. [MCP Tools Reference](#mcp-tools-reference)
5. [Dashboard Guide](#dashboard-guide)
6. [Agent Management](#agent-management)
7. [Troubleshooting](#troubleshooting)
8. [Advanced Usage](#advanced-usage)
9. [API Reference](#api-reference)

---

## Overview

**Eye in the Sky** is a comprehensive multi-agent management system designed for Claude Code users. It provides real-time visibility and control over multiple concurrent Claude Code instances, tracking AI agents working across different git worktrees and projects.

### Key Features

- **Real-time Agent Tracking**: Monitor multiple Claude Code instances simultaneously
- **Git Integration**: Track agents working in different git worktrees
- **Claude Desktop Support**: Manage both worktree and desktop-based agents
- **Activity Logging**: Complete audit trail of agent actions and git commits
- **Web Dashboard**: Intuitive web interface for agent monitoring
- **MCP Integration**: Seamless integration with Claude Code via Model Context Protocol

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Claude Code   │    │  Eye in the Sky │    │  Web Dashboard  │
│    Instances    │◄──►│   MCP Server    │◄──►│  (localhost:8080)│
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │ SQLite Database │
                       │  (agents.db)    │
                       └─────────────────┘
```

### Agent Types

**Worktree Agents**: Claude Code instances working in git worktrees
- Tracked by git worktree path
- Full git integration
- Automatic commit tracking

**Desktop Agents**: Claude Desktop instances working on projects
- Tracked by project name
- Window management support
- Session-based tracking

---

## Installation & Setup

### Prerequisites

- **Go 1.21+** (for building from source)
- **Git** (for version control)
- **Modern web browser** (for dashboard access)

### Quick Start

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd eye-in-the-sky
   ```

2. **Build the application**:
   ```bash
   # Build the integrated server (MCP + Dashboard)
   go build -o bin/eye-in-the-sky-integrated ./main.go

   # Or build MCP-only server
   go build -o bin/eye-in-the-sky ./cmd/server
   ```

3. **Start the server**:
   ```bash
   # Start integrated server (recommended)
   ./bin/eye-in-the-sky-integrated

   # Or start with custom settings
   ./bin/eye-in-the-sky-integrated -port 8081 -db ./custom/path/agents.db
   ```

4. **Access the dashboard**:
   Open your web browser to `http://localhost:8080`

### Command Line Options

```bash
./bin/eye-in-the-sky-integrated [options]

Options:
  -port string    Port for dashboard server (default "8080")
  -db string      SQLite database path (default "./data/agents.db")
  -help           Show help information
```

### Directory Structure

After setup, your directory structure will look like:

```
eye-in-the-sky/
├── bin/
│   ├── eye-in-the-sky-integrated    # Main executable
│   └── eye-in-the-sky              # MCP-only server
├── data/
│   └── agents.db                   # SQLite database (auto-created)
├── web/
│   ├── templates/                  # Dashboard HTML templates
│   └── static/                     # CSS and JavaScript files
└── internal/                       # Go source code
```

---

## Getting Started

### Starting Your First Session

1. **Start the Eye in the Sky server**:
   ```bash
   ./bin/eye-in-the-sky-integrated
   ```

2. **Register a Claude Code agent** (in Claude Code):
   ```json
   {
     "description": "Working on user authentication feature",
     "worktree_path": "/path/to/your/project"
   }
   ```

3. **Monitor agent activity** via the dashboard at `http://localhost:8080`

### Basic Workflow

1. **Agent Registration**: Register each Claude Code instance when starting work
2. **Status Updates**: Agents update their status as they work (active/working/idle)
3. **Action Logging**: All significant activities are logged automatically
4. **Commit Tracking**: Git commits are tracked and associated with agents
5. **Session Completion**: End sessions with summaries when work is finished

### Agent Lifecycle

```
Registration → Active → Working → Idle → Completed
     ↓            ↓         ↓        ↓         ↓
  Assigned ID   Tasks    Actions   Paused   Summary
```

---

## MCP Tools Reference

The Eye in the Sky system provides 7 MCP tools for Claude Code integration:

### 1. register_agent

Register a new worktree-based Claude Code agent.

**Parameters**:
```json
{
  "agent_id": "optional - auto-generates 8-char hash if not provided",
  "description": "required - what the agent will work on",
  "worktree_path": "optional - path to git worktree"
}
```

**Example**:
```json
{
  "description": "Implementing user authentication system",
  "worktree_path": "/Users/dev/projects/myapp"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Agent registered successfully",
  "agent_id": "a3f7d2e1"
}
```

### 2. register_claude_desktop_agent

Register a new Claude Desktop agent.

**Parameters**:
```json
{
  "agent_id": "optional - auto-generates 8-char hash if not provided",
  "description": "required - what the agent will work on",
  "project_name": "required - name of the project",
  "window_id": "optional - for window management"
}
```

**Example**:
```json
{
  "description": "Building React components for dashboard",
  "project_name": "Dashboard Redesign",
  "window_id": "win_12345"
}
```

### 3. update_status

Update an agent's current status and task.

**Parameters**:
```json
{
  "agent_id": "required - 8-character agent hash",
  "status": "required - active|idle|working|completed|failed",
  "current_task": "optional - specific task description"
}
```

**Example**:
```json
{
  "agent_id": "a3f7d2e1",
  "status": "working",
  "current_task": "Implementing JWT token validation"
}
```

### 4. log_action

Log agent activities and operations.

**Parameters**:
```json
{
  "agent_id": "required - 8-character agent hash",
  "action_type": "required - task_start|file_operation|git_commit|status_update",
  "description": "required - human readable description",
  "details": "optional - JSON string with additional details"
}
```

**Example**:
```json
{
  "agent_id": "a3f7d2e1",
  "action_type": "file_operation",
  "description": "Created authentication middleware",
  "details": "{\"file\": \"middleware/auth.js\", \"lines_added\": 45}"
}
```

### 5. log_commits

Track git commits made by agents.

**Parameters**:
```json
{
  "agent_id": "required - 8-character agent hash",
  "commit_hashes": "required - array of commit hashes",
  "commit_messages": "optional - array of commit messages"
}
```

**Example**:
```json
{
  "agent_id": "a3f7d2e1",
  "commit_hashes": ["abc123ef", "def456gh"],
  "commit_messages": ["Add JWT middleware", "Fix token expiration bug"]
}
```

### 6. end_session

Complete an agent session with summary.

**Parameters**:
```json
{
  "agent_id": "required - 8-character agent hash",
  "summary": "optional - session summary",
  "final_status": "optional - completed|failed (defaults to completed)"
}
```

**Example**:
```json
{
  "agent_id": "a3f7d2e1",
  "summary": "Successfully implemented user authentication with JWT tokens. Added middleware, tests, and documentation.",
  "final_status": "completed"
}
```

### 7. help

Get detailed help information about available tools.

**Parameters**:
```json
{
  "tool": "optional - specific tool name for detailed help"
}
```

**Examples**:
```json
{}                              // Get general help
{"tool": "register_agent"}      // Get specific tool help
```

---

## Dashboard Guide

The web dashboard provides a real-time view of all agent activity.

### Main Dashboard (/)

**Agent Overview Section**:
- **Active Agents**: Count of currently active agents
- **Total Sessions**: All-time session count
- **Recent Activity**: Latest agent actions

**Agent List**:
- **Agent ID**: 8-character git-style hash
- **Status Badge**: Color-coded status indicator
  - 🟢 Green: Active/Working
  - 🟡 Yellow: Idle
  - 🔴 Red: Failed
  - ⚪ Gray: Completed
- **Source Type**: Worktree 📁 or Desktop 🖥️
- **Description**: Agent's current task
- **Last Activity**: Time since last action

### Agent Details (/agent/{id})

**Agent Information**:
- Current status and task
- Registration details
- Session duration
- Git worktree path (if applicable)

**Activity Timeline**:
- Chronological list of all actions
- Timestamps and descriptions
- Action type categorization

**Git Commits**:
- List of commits made during session
- Commit hashes and messages
- Links to commit details

**Quick Actions**:
- Update agent status
- End session
- Export activity log

### Navigation

- **Home**: Return to main dashboard
- **Agents**: Filter by agent type or status
- **Activity**: System-wide activity log
- **Settings**: Configuration options

---

## Agent Management

### Best Practices

**Agent Registration**:
- Register agents at the start of each work session
- Use descriptive names that explain the work being done
- Include worktree paths for better tracking

**Status Updates**:
- Update status when switching between tasks
- Use "idle" when taking breaks
- Use "working" for active development

**Action Logging**:
- Log significant milestones and decisions
- Include relevant details in the details field
- Use consistent action types

**Session Management**:
- End sessions with descriptive summaries
- Include key accomplishments and next steps
- Mark failed sessions appropriately

### Agent States

| Status | Description | When to Use |
|--------|-------------|-------------|
| `active` | Agent is available and ready | Just registered, ready for work |
| `working` | Agent is actively working | Currently coding, debugging, testing |
| `idle` | Agent is paused | Taking a break, waiting for input |
| `completed` | Session finished successfully | Work completed, session ended |
| `failed` | Session ended with issues | Encountered blocking issues |

### Multi-Agent Coordination

**Worktree Management**:
- Each agent should work in a separate git worktree
- Coordinate merge conflicts through the dashboard
- Track cross-agent dependencies

**Task Distribution**:
- Assign clear, non-overlapping tasks
- Use action logging to communicate progress
- Monitor dashboard for coordination needs

---

## Troubleshooting

### Common Issues

**Dashboard not accessible**:
- Verify server is running: `ps aux | grep eye-in-the-sky`
- Check port availability: `curl -I http://localhost:8080`
- Restart server: `./bin/eye-in-the-sky-integrated`

**MCP tools not responding**:
- Check server logs for errors
- Verify database connection
- Restart with fresh database if needed

**Agent registration failing**:
- Verify agent_id format (8 hex characters)
- Check for duplicate registrations
- Validate required parameters

**Database issues**:
- Check file permissions on `./data/agents.db`
- Verify disk space availability
- Run with different database path if needed

### Error Codes

| Error | Cause | Solution |
|-------|-------|----------|
| `agent_not_found` | Invalid agent ID | Verify agent is registered |
| `duplicate_agent` | Agent ID already exists | Use different ID or update existing |
| `invalid_status` | Unknown status value | Use: active, idle, working, completed, failed |
| `database_error` | SQLite connection issue | Check database file and permissions |

### Debug Mode

Start server with debug logging:
```bash
./bin/eye-in-the-sky-integrated -debug
```

### Log Files

Server logs are written to stderr. Capture with:
```bash
./bin/eye-in-the-sky-integrated 2> server.log
```

---

## Advanced Usage

### Custom Database Paths

```bash
# Use custom database location
./bin/eye-in-the-sky-integrated -db /custom/path/agents.db

# Use temporary database for testing
./bin/eye-in-the-sky-integrated -db ./test-agents.db
```

### Multiple Server Instances

Run multiple instances on different ports:
```bash
# Production instance
./bin/eye-in-the-sky-integrated -port 8080 -db ./prod-agents.db

# Development instance
./bin/eye-in-the-sky-integrated -port 8081 -db ./dev-agents.db
```

### Database Queries

Direct SQLite access for reporting:
```bash
sqlite3 ./data/agents.db

# Example queries
SELECT * FROM agents WHERE status = 'active';
SELECT COUNT(*) FROM actions WHERE agent_id = 'a3f7d2e1';
SELECT * FROM commits WHERE created_at > datetime('now', '-1 day');
```

### Backup and Restore

**Backup**:
```bash
cp ./data/agents.db ./backup/agents-$(date +%Y%m%d).db
```

**Restore**:
```bash
cp ./backup/agents-20240101.db ./data/agents.db
```

---

## API Reference

### HTTP Endpoints

The dashboard server provides REST API endpoints:

**GET /api/agents**
- Returns: List of all agents
- Query params: `status`, `source`, `limit`

**GET /api/agent/{id}**
- Returns: Agent details with activity timeline

**GET /api/actions/{agent_id}**
- Returns: Actions for specific agent
- Query params: `limit`, `action_type`

**GET /api/commits/{agent_id}**
- Returns: Git commits for specific agent

**GET /api/stats**
- Returns: System statistics and metrics

### Response Format

All API responses follow this format:
```json
{
  "success": true,
  "data": { ... },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

Error responses:
```json
{
  "success": false,
  "error": "Error description",
  "code": "error_code"
}
```

---

## Support & Resources

### Getting Help

- **Documentation**: This manual and inline code comments
- **Issues**: Check existing issues in the repository
- **Logs**: Enable debug mode for detailed troubleshooting

### Contributing

- Follow Go conventions and best practices
- Add tests for new features
- Update documentation for changes
- Use meaningful commit messages

### Version History

- **v1.0**: Initial release with MCP integration and dashboard
- **v0.9**: Beta with Claude Desktop agent support
- **v0.8**: Alpha with basic worktree agent tracking

---

*Eye in the Sky - Making multi-agent development visible and manageable.*