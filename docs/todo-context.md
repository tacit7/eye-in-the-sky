# Todo Backend Context

Complete documentation for the Todo Management Backend used by Eye-in-the-Sky for multi-agent task coordination.

## Overview

The Todo Backend is a SQLite-based task management system integrated with Eye-in-the-Sky MCP (Model Context Protocol) tools and the TUI Dashboard. It allows agents to create, manage, and track tasks with full session and agent provenance tracking. The todo system is fully integrated with the Eye-in-the-Sky TUI for real-time task management and visibility.

**Key Features:**
- Consolidated SQLite database at `~/.config/eye-in-the-sky/agents.db` (shared with Eye-in-the-Sky)
- 12 MCP commands for full CRUD operations (scoped as `i-todo-*`)
- TUI integration for task display, sorting, and management
- FTS5 full-text search on tasks (indexed and auto-synced)
- Multi-agent coordination via session/agent IDs
- Workflow state management (global: todo, in_progress, done)
- Integer priorities (0-5) for flexible priority levels
- Tags, due dates, and rich notes
- Automatic weekly reindex and maintenance
- Full integration with Eye-in-the-Sky agent system and dashboard
- Transaction-based writes with rollback on error

## Architecture

### Directory Structure

```
internal/
├── database/
│   ├── db.go                                    # Main DB connection and methods
│   ├── migrations/
│   │   ├── 0028_create_todo_tables.sql        # Todo schema (projects, tasks, FTS5)
│   │   └── *.sql                               # Other migrations
│   └── queries.go                              # Database queries
├── todo/
│   ├── models/
│   │   └── models.go                           # Data structures (Task, Project, Note, Tag)
│   ├── repository/
│   │   ├── project_repo.go                     # Project CRUD operations
│   │   ├── task_repo.go                        # Task CRUD and FTS5 search
│   │   └── note_repo.go                        # Note operations
│   ├── mcp/
│   │   ├── handlers.go                         # MCP command handlers (12 commands)
│   │   └── registry.go                         # Command routing and registration
│   ├── util/
│   │   └── validation.go                       # Input validation
│   └── service.go                              # High-level service layer
└── mcp/
    └── server.go                               # MCP server initialization
```

**Note:** The todo system is now consolidated into the main database. No separate `internal/todo/db/` package exists.

### TUI Integration Layer

The TUI Dashboard integrates with the todo backend through:

```
internal/ui/app/
├── client.go                   # DataClient with TodoStore integration
├── tasks_tab.go               # Task display and rendering (Tasks tab #6)
├── overview_tab.go            # Task summary in overview
└── update_tasks.go            # Task interactions and state updates

internal/data/
└── todo.go                    # TodoStore implementing TaskStore interface
```

**TUI Components:**
- **Tasks Tab**: Displays all tasks for current agent, sorted by priority/state/date
- **Task Details**: Shows full task info with notes, tags, project, state
- **Overview Summary**: Task counts by state (Todo, In Progress, Done)
- **Mark Done**: `d` key marks task complete via `MarkDone()` method

### Technology Stack

- **Language:** Go 1.21+
- **Database:** SQLite3 with FTS5 full-text search
- **MCP Integration:** Go MCP SDK
- **Concurrency:** SQLite WAL mode + explicit transactions for writes
- **Serialization:** JSON for MCP commands and responses

## Database Schema

### Tables

#### projects
- `id` TEXT PRIMARY KEY - UUID identifier
- `name` TEXT - Project name
- `path` TEXT - Optional file path (UNIQUE)
- `remote_url` TEXT - Optional git remote URL
- `subpath` TEXT - Optional subpath
- `module` TEXT - Optional module name
- `salt` TEXT - Optional salt for hashing
- `id_algorithm` TEXT - ID generation algorithm (e.g., 'uuidv5')
- `created_at`, `updated_at` DATETIME - Timestamps
- `last_commit` TEXT - Last commit hash
- `active` BOOLEAN - Active flag

#### workflow_states (GLOBAL)
- `id` INTEGER PRIMARY KEY - Auto-increment
- `name` TEXT UNIQUE - State name (e.g., "todo", "in_progress", "done")
- `position` INTEGER - Display order (optional)
- `color` TEXT - Display color (optional)
- `updated_at` TIMESTAMP - Last update
- **Note:** Global states, not per-project

#### tasks
- `id` TEXT PRIMARY KEY - UUID identifier
- `title` TEXT NOT NULL - Task title
- `description` TEXT - Optional description
- `project_id` TEXT - Foreign key to projects
- `state_id` INTEGER - Foreign key to workflow_states
- `priority` INTEGER - 0-5 priority level
- `due_at` DATETIME - Optional due date
- `completed_at` DATETIME - Optional completion timestamp
- `session_id` TEXT - Eye-in-the-Sky session ID (optional)
- `agent_id` TEXT - Eye-in-the-Sky agent ID (optional)
- `created_at`, `updated_at` DATETIME - Timestamps
- `archived` BOOLEAN - Soft delete flag

#### task_notes
- `id` INTEGER PRIMARY KEY - Auto-increment
- `task_id` TEXT - Foreign key to tasks
- `author` TEXT - Optional author name
- `body` TEXT - Note content (plain text or markdown)
- `created_at` DATETIME - Timestamp (append-only)

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
- `task_id` TEXT UNINDEXED - Links to tasks.id
- `title` TEXT - Task title (indexed)
- `description` TEXT - Task description (indexed)
- **Porter tokenizer** for intelligent word stemming
- **Auto-synced** via triggers on task insert/update/delete

#### meta
- `key` - Configuration key
- `value` - Configuration value
- Stores: `last_reindex_at` (Sunday reindex tracking)

### Triggers

**Automatic Timestamp Updates:**
- `update_tasks_updated_at` - Set updated_at on task changes
- `update_projects_updated_at` - Set updated_at on project changes
- `update_workflow_states_updated_at` - Set updated_at on state changes

**FTS5 Synchronization:**
- `sync_task_search_insert` - Insert task_id, title, description into FTS5 on task creation
- `sync_task_search_update` - Update title and description in FTS5 on task changes
- `sync_task_search_delete` - Remove from FTS5 on task deletion

## MCP Commands

All 12 commands are exposed as MCP tools following Eye-in-the-Sky naming convention with the `i-todo-` prefix. This follows the same pattern as other Eye-in-the-Sky commands like `i-start-session`, `i-action`, `i-log`, etc.

**Response Format:** Most commands return standard format with `task_id` (UUID), `description` (task title), and `uuid_short` (first 8 characters of UUID). List and search commands return arrays of tasks or results.

### 1. i-todo-create

**Purpose:** Create a new task

**Input:**
```json
{
  "project_id": "test-project-001",
  "title": "Task description",
  "description": "Detailed description (optional)",
  "priority": 3,
  "tags": ["feature", "backend"],
  "session_id": "a745298e-2081-498a-85db-62954f4dbc05",
  "agent_id": "50cea2e9-3049-4ae1-a861-e8547132f9c8"
}
```

**Output:**
```json
{
  "task_id": "3ff1d4c2-ebdc-4bfe-9f01-2eb8a67eb590",
  "description": "Task description",
  "uuid_short": "3ff1d4c2"
}
```

**Behavior:**
- Creates task with default state "todo" (state_id = 1)
- Generates UUID for task ID
- Adds all specified tags (auto-creates tags if needed)
- Tracks session and agent IDs for provenance
- Auto-syncs to FTS5 search index
- Wrapped in transaction with rollback on error
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
  "project_id": "test-project-001",
  "filters": {
    "state_id": 2,
    "tags": ["feature"],
    "priority": 3,
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
      "id": "3ff1d4c2-ebdc-4bfe-9f01-2eb8a67eb590",
      "title": "Task description",
      "priority": 3,
      "state_id": 2,
      "tags": ["feature", "urgent"]
    }
  ]
}
```

**Filters (all optional):**
- `state_id` - Workflow state ID (1=todo, 2=in_progress, 3=done)
- `tags` - Array of tags (AND logic)
- `priority` - Integer 0-5
- `active` - Only non-archived (default: false)

### 8. i-todo-search

**Purpose:** Full-text search on tasks

**Input:**
```json
{
  "project_id": "test-project-001",
  "query": "feature implementation",
  "limit": 10
}
```

**Output:**
```json
{
  "results": [
    {
      "task_id": "3ff1d4c2-ebdc-4bfe-9f01-2eb8a67eb590",
      "title": "Implement new feature",
      "rank": 0.75
    }
  ]
}
```

**Behavior:**
- Uses FTS5 indexed search with Porter tokenizer
- Searches title and description fields
- Results ranked by FTS5 relevance score
- Project-scoped
- Pagination support via limit and offset

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

1. `database.New(dbPath)` opens/creates `~/.config/eye-in-the-sky/agents.db`
2. Enables WAL mode for concurrent access
3. Enables foreign key constraints
4. Runs all migrations in order (including `0028_create_todo_tables.sql`)
5. Sets up triggers for automatic FTS5 synchronization
6. Checks if weekly FTS5 reindex is needed

The todo system is fully consolidated into the main database, with no separate initialization needed.

## Performance Considerations

- **WAL Mode:** Enables concurrent reads while writes happen
- **FTS5:** Fast full-text search (indexed)
- **Indexes:** All foreign keys and common filters are indexed
- **Transactions:** Write operations wrapped in explicit transactions
- **Weekly Maintenance:** Automatic reindex on Sundays (7+ days)

## Development Commands

### Build (with FTS5 support)
```bash
go build -tags "sqlite_fts5" -o bin/eye-in-the-sky ./cmd/server
```

### Test
```bash
go test -tags "sqlite_fts5" ./internal/todo/...
```

### Database
The consolidated database file is automatically created at:
```
~/.config/eye-in-the-sky/agents.db
```

To reset (clears all Eye-in-the-Sky and todo data):
```bash
rm ~/.config/eye-in-the-sky/agents.db
```

### View Database
```bash
sqlite3 ~/.config/eye-in-the-sky/agents.db
sqlite> SELECT * FROM tasks;
sqlite> SELECT * FROM task_search WHERE task_search MATCH 'search term';
```

## Domain Model

The internal todo system uses a unified domain model in `internal/domain/task.go`:

```go
type Task struct {
    ID              TaskID        // UUID
    Title           string        // Task title (required)
    Description     string        // Optional detailed description
    ProjectID       string        // Project association
    StateID         int           // 1=todo, 2=in_progress, 3=done
    Priority        int           // 0-5 scale (5=critical, 0=none)
    DueAt           time.Time     // Optional due date
    CompletedAt     time.Time     // Completion timestamp
    SessionID       string        // Eye-in-the-Sky session ID
    AgentID         string        // Eye-in-the-Sky agent ID
    CreatedAt       time.Time     // Creation timestamp
    UpdatedAt       time.Time     // Last update timestamp
    Archived        bool          // Soft delete flag
    Notes           []TaskNote    // Append-only notes
    Tags            []string      // Associated tags
    WorkflowStatus  string        // Human-readable state ("todo", "in_progress", "done")
}

type TaskNote struct {
    ID        int
    TaskID    TaskID
    Author    string
    Body      string        // Markdown content
    CreatedAt time.Time
}
```

**Priority Scale:**
- 5: Critical (high urgency)
- 4-3: High
- 2: Medium
- 1: Low
- 0: None/Unset

**State IDs:**
- 1: todo
- 2: in_progress
- 3: done

## Examples

### Create a Task
```json
{
  "command": "i-todo-create",
  "args": {
    "project_id": "eye-in-the-sky",
    "title": "Setup CI/CD pipeline",
    "priority": 5,
    "tags": ["infra", "critical"],
    "session_id": "a745298e-2081-498a-85db-62954f4dbc05",
    "agent_id": "50cea2e9-3049-4ae1-a861-e8547132f9c8"
  }
}
```

### Search for Tasks
```json
{
  "command": "i-todo-search",
  "args": {
    "project_id": "eye-in-the-sky",
    "query": "setup ci/cd"
  }
}
```

### Move Task to In-Progress
```json
{
  "command": "i-todo-start",
  "args": {
    "task_id": "3ff1d4c2-ebdc-4bfe-9f01-2eb8a67eb590"
  }
}
```

### Change Task State (by ID)
```json
{
  "command": "i-todo-status",
  "args": {
    "task_id": "3ff1d4c2-ebdc-4bfe-9f01-2eb8a67eb590",
    "state_id": 2
  }
}
```

### Add Notes to Task
```json
{
  "command": "i-todo-annotate",
  "args": {
    "task_id": "3ff1d4c2-ebdc-4bfe-9f01-2eb8a67eb590",
    "body": "Progress: Completed initial setup and GitHub Actions configuration.\nNext: Add tests and security scanning."
  }
}
```

### Complete a Task
```json
{
  "command": "i-todo-done",
  "args": {
    "task_id": "3ff1d4c2-ebdc-4bfe-9f01-2eb8a67eb590"
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

## TUI Task Management

The Eye-in-the-Sky TUI Dashboard displays and manages tasks in the Tasks tab:

### Display Features

- **Task List**: Shows all tasks for the current agent
- **Sorting**: Priority (5→0) → State (1→3) → CreatedAt (newest first)
- **Color Coding**:
  - Priority: [CRIT] (red), [HIGH] (red), [MED] (yellow), [LOW] (gray), [-] (gray)
  - State: done (green), in_progress (yellow), todo (primary color)
- **Pagination**: Loads up to 100 tasks per agent

### Task Details Panel

When a task is selected, shows:
- Task ID (first 8 chars of UUID)
- Priority (0-5 integer)
- State (human-readable: "todo", "in_progress", "done")
- Project ID
- Tags
- Notes with timestamps
- Session/Agent IDs (for provenance)

### Keyboard Navigation

- `j`/`k`: Navigate task list up/down
- `g`/`G`: Go to start/end of list
- `d`: Mark current task as done (triggers MCP MarkDone)
- `Page Up/Down`: Scroll task list
- `T`: Switch to Tasks tab

### Integration Points

1. **DataClient**: Initializes TodoStore for task loading
2. **TaskStore Interface**: Implemented by TodoStore
3. **Task Model**: Uses unified `internal/domain/task.go`
4. **MCP Integration**: Mark done calls `Tasks.MarkDone()` which updates database

## Future Enhancements

Potential improvements:
- Batch operations (create multiple tasks)
- Advanced filtering (date ranges, complex queries)
- Task templates
- Notifications and webhooks
- Task dependencies and blocking
- Time tracking integration
- Export/import functionality
- Edit task state/priority from TUI
- Create new tasks from TUI
- Annotate tasks from TUI
