# EITS.DB Migration Context

**Session:** a6afd4f3-f2c7-4ebd-b599-3e23aba05888
**Agent:** f6be0d33-5971-47f5-9217-2a440fd22110
**Date:** 2025-11-05
**Commit:** a2a9333

## Migration Overview

Migrating Eye in the Sky from `agents.db` to `eits.db` with simplified schema. This is a **major refactor** changing project IDs from TEXT (UUIDs) to INTEGER (autoincrement).

## What Was Completed (Phase 1 & 2)

### ✅ Database Path Migration
Changed all database paths from `agents.db` → `eits.db` in:
- `cmd/server/main.go` (line 26, 43)
- `cmd/todo/main.go` (line 26)
- `cmd/eits-cli/main.go` (line 32, 93)
- `main.go` (line 19)
- `internal/ui/app/config.go` (line 61)

### ✅ Project Model Refactor
**File:** `internal/todo/models/models.go`

**Before:**
```go
type Project struct {
    ID            string     `json:"id"`           // TEXT UUID
    Name          string     `json:"name"`
    Path          *string    `json:"path,omitempty"`
    RemoteURL     *string    `json:"remote_url,omitempty"`
    Subpath       *string    `json:"subpath,omitempty"`
    Module        *string    `json:"module,omitempty"`
    Salt          *string    `json:"salt,omitempty"`
    IDAlgorithm   string     `json:"id_algorithm"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
    LastCommit    *string    `json:"last_commit,omitempty"`
    Active        bool       `json:"active"`
}
```

**After:**
```go
type Project struct {
    ID        int        `json:"id"`           // INTEGER autoincrement
    Name      string     `json:"name"`
    Path      *string    `json:"path,omitempty"`
    RemoteURL *string    `json:"remote_url,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}
```

**Also updated:**
- `Task.ProjectID` from `string` → `int`

### ✅ Project Repository Refactor
**File:** `internal/todo/repository/project_repo.go`

**Changed function signatures:**
- `CreateProject(name string, path *string, remoteURL *string)` - No longer takes ID param, uses autoincrement
- `GetProjectByID(id int)` - Changed from `string` to `int`
- `GetProjectByPath(path string)` - Updated schema fields
- `GetProjectByRemoteURL(remoteURL string)` - Updated schema fields
- `UpdateProject(id int, updates map[string]interface{})` - Changed from `string` to `int`
- `ListProjects()` - Removed `active` filter, updated schema

**SQL Changes:**
- Removed fields: `subpath`, `module`, `salt`, `id_algorithm`, `last_commit`, `active`
- SELECT queries now use: `id, name, path, remote_url, created_at, updated_at`

### ✅ Service Layer Additions
**File:** `internal/todo/service.go`

Added `DetectCurrentProject()` function:
- Uses `gitinfo.RemoteURL()` for primary matching (more reliable)
- Falls back to `gitinfo.RepoPath()` for path matching
- Returns `*models.Project` with integer ID

### ✅ MCP Handler Additions
**File:** `internal/todo/mcp/handlers.go`

Added `HandleGetProject()`:
- Detects current project using git info
- Returns `GetProjectResponse` with integer project_id

### ✅ Database Schema Update
**Added to eits.db:**
```sql
CREATE TABLE agent_context (
  agent_id TEXT NOT NULL,
  project_id INTEGER NOT NULL,
  context TEXT NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (agent_id, project_id),
  FOREIGN KEY (agent_id) REFERENCES agents(id),
  FOREIGN KEY (project_id) REFERENCES projects(id)
);
```

## ✅ Phase 2 Complete: Repository, Service, and MCP Layers (Session 2)

### ✅ Repository Layer
**File:** `internal/todo/repository/task_repo.go`
- Changed all `projectID string` params → `projectID int`
- Updated functions:
  - `CreateTask(projectID int, ...)`
  - `List(projectID int, ...)`
  - `Search(projectID int, ...)`

**File:** `internal/todo/repository/note_repo.go`
- Updated functions:
  - `GetProjectNoteStats(projectID int)`
  - `GetNotesByProjectID(projectID int)`
  - `SearchNotes(projectID int, ...)`

### ✅ Service Layer
**File:** `internal/todo/service.go`
- Updated all project operations:
  - `CreateProject(name string, path *string, remoteURL *string)` - Removed uuid param
  - `GetProject(projectID int)` - Changed param type
  - `UpdateProject(projectID int, ...)` - Changed param type
  - `CreateTask(projectID int, ...)` - Changed param type
  - `ListTasks(projectID int, ...)` - Changed param type
  - `ListTasksWithSort(projectID int, ...)` - Changed param type
  - `SearchTasks(projectID int, ...)` - Changed param type
  - `GetProjectNoteStats(projectID int)` - Changed param type

### ✅ MCP Handlers
**File:** `internal/todo/mcp/handlers.go`
- Updated all request/response types:
  - `CreateRequest.ProjectID` - Changed from `string` to `int`
  - `ListRequest.ProjectID` - Changed to `int`
  - `SearchRequest.ProjectID` - Changed to `int`
  - `GetProjectResponse.ProjectID` - Changed to `int`
  - `ProjectSyncRequest.ProjectID` - Changed to `int`

### ✅ Data Layer
**File:** `internal/data/todo.go`
- Updated `projectID` variable from `string` to `int` in `LoadByProject()`
- Fixed `projectID` scanning from `*string` to `*int` in task query

### ✅ Domain Layer
**File:** `internal/domain/task.go`
- Updated `Task.ProjectID` from `string` to `int`

### ✅ CLI Layer
**File:** `cmd/todo/main.go`
- Fixed `CreateProject()` calls to use new signature (removed uuid param)
- Fixed `CreateTask()` calls to remove extra nil parameter
- Fixed task ID handling (kept as string for UUIDs)
- Fixed `Note.BodyMarkdown` → `Note.Body`

## ✅ Phase 3 Complete: Remove Unused Features (Session 2)

**Personas System** - Removed completely:
- ✅ `internal/mcp/server.go` - Removed `i-snapshot-expertise`, `i-persona-get`, `i-persona-list` tool registrations
- ✅ `internal/mcp/server.go` - Removed handler functions: `handleSnapshotExpertise`, `handleGetPersona`, `handleListPersonas`
- ✅ `internal/mcp/tools.go` - Removed tool implementations: `SnapshotExpertise`, `GetPersona`, `ListPersonas`, `CreatePersona`
- ✅ `internal/mcp/tools.go` - Removed persona loading from `StartSession()` function
- ✅ `internal/mcp/tools.go` - Updated instructions to remove persona tool documentation
- ✅ `internal/mcp/types.go` - Removed types: `GetPersonaArgs`, `GetPersonaResult`, `ListPersonasArgs`, `ListPersonasResult`, `PersonaSummary`, `SnapshotExpertiseArgs`, `SnapshotExpertiseResult`
- ✅ `internal/database/...` - Persona CRUD functions already removed in earlier commit

**Compaction Tracking** - Removed completely:
- ✅ `internal/mcp/server.go` - Removed `i-log-compaction` tool registration
- ✅ `internal/mcp/server.go` - Removed handler function: `handleLogCompaction`
- ✅ `internal/mcp/tools.go` - Removed tool implementation: `LogCompaction()` and helper functions (`copyFile`, `countJSONLLines`)
- ✅ `internal/mcp/tools.go` - Updated instructions to remove compaction documentation
- ✅ `internal/mcp/types.go` - Removed types: `LogCompactionArgs`, `LogCompactionResult`
- ✅ `internal/mcp/tools.go` - Removed unused imports: `bufio`, `io`, `path/filepath`
- ✅ `internal/database/...` - Compaction CRUD functions already removed in earlier commit

**Session Context** - Kept (still useful for state persistence)

## ✅ Phase 4 Complete: Schema Fixes and Testing (Session 2)

### Schema Issues Fixed

**Problems Identified:**
1. Tasks table had `id INTEGER` but code expected `id TEXT` (UUID)
2. Missing columns: `due_at`, `completed_at`
3. `session_id` was `INTEGER` but code expected `TEXT`
4. Missing supporting tables: `task_notes`, `tags`, `task_tags`, `task_search`
5. Missing default workflow states
6. NULL handling issues for `priority` and `archived` fields

**Schema Changes Applied:**

**Tasks Table Recreated:**
```sql
CREATE TABLE tasks (
  id TEXT PRIMARY KEY,                    -- Changed from INTEGER to TEXT (UUID)
  title TEXT NOT NULL,
  description TEXT,
  project_id INTEGER,                     -- Kept as INTEGER (correct)
  state_id INTEGER NOT NULL,
  agent_id TEXT,
  session_id TEXT,                        -- Changed from INTEGER to TEXT
  priority INTEGER DEFAULT 0,
  due_at TIMESTAMP,                       -- ADDED
  completed_at TIMESTAMP,                 -- ADDED
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP,
  archived INTEGER DEFAULT 0,
  FOREIGN KEY (project_id) REFERENCES projects(id),
  FOREIGN KEY (state_id) REFERENCES workflow_states(id)
);
```

**Supporting Tables Created:**
- `task_notes` - Task annotations with TEXT task_id FK
- `tags` - Global tag definitions
- `task_tags` - Junction table for task-tag relationships
- `task_search` - FTS5 virtual table for full-text search

**Default Data Inserted:**
```sql
INSERT INTO workflow_states (id, name, position, color) VALUES
  (1, 'todo', 0, '#gray'),
  (2, 'in_progress', 1, '#blue'),
  (3, 'done', 2, '#green');
```

### Code Fixes

**File:** `internal/todo/repository/task_repo.go`

1. **CreateTask function:**
   - Added default `state_id = 1` when not provided
   - Ensures all tasks have valid workflow state

2. **FindByID function:**
   - Added `COALESCE(priority, 0)` to handle NULL priorities
   - Added `COALESCE(archived, 0)` to handle NULL archived flags

3. **List function:**
   - Updated SELECT query with COALESCE for nullable fields

4. **Search function:**
   - Updated SELECT query with COALESCE for nullable fields

### Testing Results

✅ **Project creation:** Autoincrement IDs working correctly
```bash
$ go run cmd/todo/main.go project create "test-migration-project"
Created project: &{1 test-migration-project <nil> <nil> 2025-11-07...}
```

✅ **Task creation:** UUID-based task IDs working correctly
```bash
$ go run cmd/todo/main.go task create 1 "Test task after all schema fixes"
Created task: &{f0f25941-fd53-4b91-939b-2907514c23bc Test task after all schema fixes...}
```

✅ **Task listing:** Tasks retrieved successfully
```bash
$ go run cmd/todo/main.go task list 1
Tasks for project 1:
  [f0f25941-fd53-4b91-939b-2907514c23bc] ...
  [740d046b-41c4-4ccf-8557-cd0672d854a6] ...
```

✅ **Database verification:** Schema and data correct
```sql
SELECT id, title, project_id, state_id, priority FROM tasks;
-- Results: 2 tasks with UUID IDs, integer project_id, default state_id=1
```

### Files Modified (Session 2, Phase 4)
- `internal/todo/repository/task_repo.go` - Query fixes and default state handling

### Commits
1. `25bc8f1` - Complete eits.db migration Phase 2 & 3
2. `fcd4d47` - Fix eits.db schema compatibility and task creation

## Migration Complete! 🎉

All phases successfully completed:
- ✅ Phase 1: Database path migration (agents.db → eits.db)
- ✅ Phase 2: Project ID migration (TEXT UUID → INTEGER autoincrement)
- ✅ Phase 3: Remove unused features (personas, compaction tracking)
- ✅ Phase 4: Schema fixes and end-to-end testing

The eits.db schema is now fully functional and tested.

## Schema Comparison: agents.db vs eits.db

### Projects Table

**agents.db:**
```sql
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT UNIQUE,
    remote_url TEXT,
    subpath TEXT,
    module TEXT,
    salt TEXT,
    id_algorithm TEXT DEFAULT 'uuidv5',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_commit TEXT,
    active BOOLEAN DEFAULT 1
);
```

**eits.db:**
```sql
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    path TEXT UNIQUE,
    remote_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Key Differences:**
- `id`: TEXT → INTEGER (breaking change)
- Removed: subpath, module, salt, id_algorithm, last_commit, active
- Simpler, cleaner schema

## Breaking Changes

### API Changes
- **project_id type**: All MCP tools now expect INTEGER instead of TEXT/UUID
- **Project creation**: No longer accepts ID parameter (autoincrement)
- **Features removed**: Personas, compaction tracking

### Migration Notes
- No data migration performed (starting fresh with eits.db)
- Old agents.db remains intact at ~/.config/eye-in-the-sky/agents.db
- New eits.db at ~/.config/eye-in-the-sky/eits.db

## Files Modified (This Session)

1. `cmd/server/main.go` - DB path
2. `cmd/todo/main.go` - DB path
3. `cmd/eits-cli/main.go` - DB path
4. `main.go` - DB path
5. `internal/ui/app/config.go` - DB path
6. `internal/todo/models/models.go` - Project struct refactor
7. `internal/todo/repository/project_repo.go` - All functions updated
8. `internal/todo/service.go` - Added DetectCurrentProject()
9. `internal/todo/mcp/handlers.go` - Added HandleGetProject()

## Next Steps

1. **Continue Phase 2**: Update remaining repository/service/MCP layers for int project IDs
2. **Phase 3**: Remove persona and compaction code
3. **Phase 4**: Update data layer
4. **Testing**: Verify all changes work with eits.db
5. **Documentation**: Update CLAUDE.md with new schema

## Decision Log

- **Database choice**: Using eits.db (cleaner schema)
- **Migration approach**: Refactor code to match eits.db (no data migration)
- **Project ID**: INTEGER autoincrement (more standard than UUIDs)
- **Features**: Remove personas and compaction (not needed now)
- **Session context**: Keep (still useful for agent state)
