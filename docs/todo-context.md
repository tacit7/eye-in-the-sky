# Todo Backend Context

Complete documentation for the Todo Management Backend used by Eye-in-the-Sky for multi-agent task coordination.

## Overview

The Todo Backend is a SQLite-based task management system integrated with Eye-in-the-Sky MCP (Model Context Protocol) tools. It allows agents to create, manage, and track tasks with full session and agent provenance tracking.

**Key Features:**
- SQLite database at `~/.config/eye-in-the-sky/todo.db`
- 12 MCP commands for full CRUD operations (scoped as `i-todo-*`)
- FTS5 full-text search on tasks
- Multi-agent coordination via session/agent IDs
- Hierarchical tasks (subtasks with recursive delete)
- Workflow state management (todo, doing, review, done)
- Tags, priorities, weights, and due dates
- Automatic weekly reindex and maintenance
- Full integration with Eye-in-the-Sky agent system

## Architecture

### Directory Structure

```
internal/todo/
├── db/
│   ├── db.go           # Database connection and initialization
│   ├── migrate.go      # Schema migrations
│   └── maintenance.go  # FTS5 reindex and vacuum
├── models/
│   └── models.go       # Data structures (Task, Project, Note, Tag, etc.)
├── repository/
│   ├── project_repo.go # Project CRUD operations
│   ├── task_repo.go    # Task CRUD and search
│   └── note_repo.go    # Note operations
├── mcp/
│   ├── handlers.go     # MCP command handlers (12 commands)
│   └── registry.go     # Command routing and registration
├── util/
│   ├── validation.go   # Input validation
│   └── yaml_workflow.go # Workflow YAML parsing
└── service.go          # High-level service layer
```

### Technology Stack

- **Language:** Go 1.21+
- **Database:** SQLite3 with FTS5 full-text search
- **MCP Integration:** Go MCP SDK
- **Concurrency:** SQLite WAL mode + explicit transactions for writes
- **Serialization:** JSON for MCP commands and responses

## Database Schema

### Tables

#### projects
- `id` - Primary key
- `uuid` - Unique identifier
- `name` - Project name
- `repo_slug` - Optional git repo slug
- `created_at`, `updated_at`, `archived_at` - Timestamps

#### workflow_states
- `id` - Primary key
- `project_id` - Foreign key to projects
- `code` - State code (e.g., "todo", "doing")
- `display_name` - User-visible name
- `position` - Display order
- UNIQUE(project_id, code) - One code per project

#### tasks
- `id` - Primary key
- `project_id` - Foreign key to projects
- `description` - Task description (required)
- `state_code` - Current workflow state
- `parent_id` - Foreign key to parent task (for subtasks)
- `priority` - 1-5 (optional)
- `weight` - Integer weight (optional)
- `position` - Sort order
- `due_date` - Due date (optional)
- `session_id` - Eye-in-the-Sky session ID (optional)
- `agent_id` - Eye-in-the-Sky agent ID (optional)
- `created_at`, `updated_at`, `archived_at` - Timestamps

#### task_notes
- `id` - Primary key
- `task_id` - Foreign key to tasks
- `body_markdown` - Note content (markdown)
- `created_at` - Timestamp (append-only)

#### tags
- `id` - Primary key
- `name` - Tag name (unique)
- `created_at` - Timestamp

#### task_tags
- `id` - Primary key
- `task_id` - Foreign key to tasks
- `tag_id` - Foreign key to tags
- UNIQUE(task_id, tag_id) - One tag per task

#### task_events
- `id` - Primary key
- `task_id` - Foreign key to tasks
- `event_type` - Type of change
- `old_value`, `new_value` - Change details
- `created_at` - Timestamp

#### task_search (FTS5 virtual table)
- `rowid` - Links to tasks.id
- `description` - Full-text indexed
- `latest_note` - Most recent note (auto-synced)
- `tags` - Space-separated tag names (auto-synced)

#### meta
- `key` - Configuration key
- `value` - Configuration value
- Stores: `last_reindex_at` (Sunday reindex tracking)

### Triggers

**Automatic Updates:**
- `update_tasks_updated_at` - Set updated_at on task changes
- `update_projects_updated_at` - Set updated_at on project changes
- `update_workflow_states_updated_at` - Set updated_at on state changes

**FTS5 Synchronization:**
- `sync_task_search_insert` - Add to index on task creation
- `sync_task_search_update` - Update index on task changes
- `sync_task_search_delete` - Remove from index on deletion
- `sync_latest_note_on_insert` - Update search snapshot on new note

## MCP Commands

All 12 commands are exposed as MCP tools following Eye-in-the-Sky naming convention with the `i-todo-` prefix. This follows the same pattern as other Eye-in-the-Sky commands like `i-start-session`, `i-action`, `i-log`, etc.

**Response Format:** All commands return standard format with `task_id`, `description`, and `uuid_short` (6-digit zero-padded identifier derived from task ID, e.g., "000042").

### 1. i-todo-create

**Purpose:** Create a new task

**Input:**
```json
{
  "project_id": 1,
  "description": "Task description",
  "priority": 3,
  "tags": ["feature", "backend"],
  "parent_id": null,
  "session_id": "580AD2D7-6385-449C-B960-2C40DC726ACD",
  "agent_id": "50cea2e9-3049-4ae1-a861-e8547132f9c8"
}
```

**Output:**
```json
{
  "task_id": 42,
  "description": "Task description",
  "uuid_short": "000042"
}
```

**Behavior:**
- Creates task with default state "todo"
- Adds all specified tags
- Tracks session and agent IDs for provenance
- Returns immediately with task info

### 2. i-todo-annotate

**Purpose:** Add a markdown note to a task

**Input:**
```json
{
  "task_id": 42,
  "body": "# Progress\n\nCompleted first draft of feature."
}
```

**Output:** Task info (task_id, description, uuid_short)

**Behavior:**
- Appends note to task (never overwrites)
- Auto-syncs latest note to FTS index
- Updates task's updated_at timestamp

### 3. i-todo-start

**Purpose:** Move task to "doing" state

**Input:**
```json
{
  "task_id": 42
}
```

**Output:** Task info with new state

### 4. i-todo-done

**Purpose:** Move task to "done" state

**Input:**
```json
{
  "task_id": 42
}
```

**Output:** Task info with new state

### 5. i-todo-status

**Purpose:** Move task to any workflow state

**Input:**
```json
{
  "task_id": 42,
  "state": "review"
}
```

**Output:** Task info with new state

**Note:** State code must exist in project's workflow definition

### 6. i-todo-tag

**Purpose:** Add or remove tags from a task

**Input:**
```json
{
  "task_id": 42,
  "add": ["urgent", "review"],
  "remove": ["draft"]
}
```

**Output:** Task info with updated tags

### 7. i-todo-list

**Purpose:** Retrieve tasks with optional filters

**Input:**
```json
{
  "project_id": 1,
  "filters": {
    "state": "doing",
    "tags": ["feature"],
    "priority": 5,
    "active": true
  },
  "limit": 50
}
```

**Output:**
```json
{
  "tasks": [
    {
      "id": 42,
      "description": "Task description",
      "priority": 5,
      "state": "doing",
      "tags": ["feature", "urgent"]
    }
  ]
}
```

**Filters (all optional):**
- `state` - Workflow state code
- `tags` - Array of tags (AND logic)
- `priority` - Integer 1-5
- `active` - Only non-archived (default: false)

### 8. i-todo-search

**Purpose:** Full-text search on tasks

**Input:**
```json
{
  "project_id": 1,
  "query": "feature implementation",
  "limit": 10
}
```

**Output:**
```json
{
  "results": [
    {
      "task_id": 42,
      "description": "Implement new feature",
      "rank": 0.95
    }
  ]
}
```

**Behavior:**
- Uses FTS5 indexed search
- Searches description, notes, and tags
- Results ranked by relevance
- Project-scoped

### 9. i-todo-delete

**Purpose:** Permanently delete a task

**Input:**
```json
{
  "task_id": 42
}
```

**Output:**
```json
{
  "ok": true
}
```

**Behavior:**
- Hard delete (not soft delete)
- Cascades to all child tasks
- Removes all notes and tags
- Removes from FTS index
- Wrapped in transaction with rollback on error

### 10. i-todo-reindex

**Purpose:** Rebuild FTS5 search index

**Input:**
```json
{}
```

**Output:**
```json
{
  "ok": true
}
```

**Behavior:**
- Runs `REINDEX task_search`
- Updates last_reindex_at in meta table
- Called automatically on Sundays (7+ days since last reindex)

### 11. i-todo-vacuum

**Purpose:** Database maintenance (cleanup and analysis)

**Input:**
```json
{}
```

**Output:**
```json
{
  "ok": true
}
```

**Behavior:**
- Runs `VACUUM` (reclaims disk space)
- Runs `ANALYZE` (updates statistics)
- Can be run manually or scheduled

### 12. i-todo-project-sync

**Purpose:** Sync workflow states from YAML definition

**Input:**
```json
{
  "project_id": 1,
  "yaml": "workflow:\n  - code: todo\n    label: To Do\n  - code: doing\n    label: In Progress\n  - code: done\n    label: Done"
}
```

**Output:**
```json
{
  "ok": true
}
```

**YAML Format:**
```yaml
workflow:
  - code: todo
    label: To Do
  - code: doing
    label: In Progress
  - code: review
    label: Under Review
  - code: done
    label: Done
```

## Session and Agent ID Tracking

Every task can be created with session and agent IDs for multi-agent coordination.

### Why Track Session and Agent IDs?

- **Provenance:** Know which session/agent created the task
- **Ownership:** Track which agent is responsible
- **Coordination:** Agents can query tasks they created or are responsible for
- **Audit Trail:** Complete history of who created what

### How to Use

When creating a task via MCP, include the IDs:

```json
{
  "command": "todo.create",
  "args": {
    "project_id": 1,
    "description": "Feature implementation",
    "session_id": "580AD2D7-6385-449C-B960-2C40DC726ACD",
    "agent_id": "50cea2e9-3049-4ae1-a861-e8547132f9c8"
  }
}
```

### Querying by Session/Agent

Filter tasks to find those created by a specific session or agent:

```json
{
  "command": "todo.search",
  "args": {
    "project_id": 1,
    "query": "session_id:580AD2D7-6385-449C-B960-2C40DC726ACD"
  }
}
```

Or retrieve all tasks and filter in application logic based on returned session/agent IDs.

## Validation Rules

All inputs are validated before database operations.

### Description
- Cannot be empty
- Trimmed before validation

### Priority
- Optional
- Must be 1-5 if provided

### Weight
- Optional
- Must be non-negative
- Maximum 1000

### Tag Name
- Cannot be empty
- Maximum 64 characters
- Auto-trimmed

### Due Date
- Optional
- Stored as TIMESTAMP

### Parent Task
- Optional
- Cannot equal child task (cycle prevention)
- Maximum nesting depth: 32 levels

### Workflow State
- Must be valid for the project
- Use project's workflow definition

## Service Layer API

The `todo.Service` provides high-level operations:

```go
// Project operations
CreateProject(uuid, name)
GetProject(projectID)
ListProjects()
UpdateProject(projectID, name)
DeleteProject(projectID)
GetOrCreateProjectFromRepo(repoSlug, uuid)

// Workflow operations
GetWorkflow(projectID)
SyncWorkflowFromYAML(projectID, yamlData)

// Task operations
CreateTask(projectID, description, parentID)
GetTask(taskID)
ListTasks(projectID, filters)
ListTasksWithSort(projectID, filters, sortBy)
UpdateTaskDescription(taskID, description)
SetTaskState(taskID, stateCode)
SetTaskPriority(taskID, priority)
SetTaskWeight(taskID, weight)
SetTaskDueDate(taskID, dueDate)
SetTaskParent(taskID, parentID)
ReorderTask(taskID, newPosition)
DeleteTask(taskID) // soft delete
HardDeleteTask(taskID) // permanent delete
SearchTasks(projectID, query, limit, offset)

// Tag operations
AddTaskTag(taskID, tagName)
RemoveTaskTag(taskID, tagName)

// Note operations
AddTaskNote(taskID, bodyMarkdown)
GetNotesByTask(taskID)
GetLatestNoteForTask(taskID)
UpdateNote(noteID, bodyMarkdown)
DeleteNote(noteID)
GetProjectNoteStats(projectID)

// Database operations
Reindex()
Vacuum()
CheckAndMaybeReindex()
```

## Integration with Eye-in-the-Sky

The todo backend is integrated into the Eye-in-the-Sky MCP server as a set of 12 tools.

### In MCP Server

Tools are registered in `internal/mcp/server.go`:

```go
mcp.AddTool(s.mcp, &mcp.Tool{
    Name:        "i-todo-create",
    Description: "Create a new task with optional priority and tags",
}, s.handleTodoCreate)
```

### Usage by Agents

Agents call todo MCP tools the same way as other Eye-in-the-Sky tools:

```go
// In agent context, use the MCP tools
result, err := callMCPTool("i-todo-create", map[string]interface{}{
    "project_id": 1,
    "description": "Implement feature X",
    "session_id": currentSessionID,
    "agent_id": currentAgentID,
})
```

## Database Initialization

The database is created automatically on first use:

1. `OpenDB()` opens/creates `~/.config/eye-in-the-sky/todo.db`
2. Enables WAL mode for concurrent access
3. Enables foreign key constraints
4. Runs all migrations (CREATE TABLE statements)
5. Sets up triggers for automatic synchronization
6. Checks if weekly reindex is needed

Existing databases are updated with new columns via ALTER TABLE in migrations.

## Performance Considerations

- **WAL Mode:** Enables concurrent reads while writes happen
- **FTS5:** Fast full-text search (indexed)
- **Indexes:** All foreign keys and common filters are indexed
- **Transactions:** Write operations wrapped in explicit transactions
- **Weekly Maintenance:** Automatic reindex on Sundays (7+ days)

## Development Commands

### Build
```bash
go build -o bin/eye-in-the-sky ./cmd/server
```

### Test
```bash
go test ./internal/todo/...
```

### Database
The database file is automatically created at:
```
~/.config/eye-in-the-sky/todo.db
```

To reset:
```bash
rm ~/.config/eye-in-the-sky/todo.db
```

## Examples

### Create a Project
```json
{
  "command": "i-todo-create",
  "args": {
    "project_id": 1,
    "description": "Setup CI/CD pipeline",
    "priority": 5,
    "tags": ["infra", "critical"],
    "session_id": "580AD2D7-6385-449C-B960-2C40DC726ACD",
    "agent_id": "50cea2e9-3049-4ae1-a861-e8547132f9c8"
  }
}
```

### Create a Subtask
```json
{
  "command": "todo.create",
  "args": {
    "project_id": 1,
    "description": "Configure GitHub Actions",
    "parent_id": 5,
    "priority": 4
  }
}
```

### Search for Urgent Tasks
```json
{
  "command": "i-todo-search",
  "args": {
    "project_id": 1,
    "query": "urgent critical"
  }
}
```

### Move Task to Review
```json
{
  "command": "i-todo-status",
  "args": {
    "task_id": 42,
    "state": "review"
  }
}
```

### Add Notes and Complete
```json
{
  "command": "i-todo-annotate",
  "args": {
    "task_id": 42,
    "body": "# Completed\n\nFeature is ready for review. All tests pass."
  }
}
```

Then:
```json
{
  "command": "i-todo-done",
  "args": {
    "task_id": 42
  }
}
```

## Troubleshooting

### Database Lock Errors
- Ensure only one writer is active at a time
- Use transactions properly
- Check disk space

### Search Not Working
- Run `todo.reindex` to rebuild FTS index
- Check that tasks have descriptions

### Missing Tasks
- Verify project_id is correct
- Check if tasks are archived (use `active: true` filter)
- Run `todo.vacuum` to clean up database

## Future Enhancements

Potential improvements:
- Batch operations (create multiple tasks)
- Advanced filtering (date ranges, complex queries)
- Task templates
- Notifications and webhooks
- Task dependencies and blocking
- Time tracking integration
- Export/import functionality
