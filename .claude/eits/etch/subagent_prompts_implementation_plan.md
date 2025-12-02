# Subagent Prompts Implementation Plan

## Executive Summary

This document provides a detailed implementation plan for adding a `subagent_prompts` table to the Eye in the Sky database. The table will store reusable prompt templates that can be scoped globally or to specific projects, enabling consistent subagent behavior across sessions.

---

## 1. Database Schema Design

### Table: `subagent_prompts`

```sql
CREATE TABLE IF NOT EXISTS subagent_prompts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT,
    prompt_text TEXT NOT NULL,
    project_id TEXT,
    scope TEXT NOT NULL DEFAULT 'global',
    active BOOLEAN DEFAULT 1,
    version INTEGER DEFAULT 1,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CHECK (scope IN ('global', 'project')),
    CHECK (scope = 'global' OR project_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_subagent_prompts_project ON subagent_prompts(project_id);
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_scope ON subagent_prompts(scope);
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_active ON subagent_prompts(active);
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_slug ON subagent_prompts(slug);

CREATE TRIGGER IF NOT EXISTS update_subagent_prompts_updated_at
AFTER UPDATE ON subagent_prompts
FOR EACH ROW
BEGIN
    UPDATE subagent_prompts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
```

### Schema Design Decisions

**Primary Key:**
- `id TEXT PRIMARY KEY` - UUID-based, follows EITS pattern for entity tables

**Scoping Strategy:**
- `scope` enum: 'global' or 'project'
- `project_id` nullable FK to `projects(id)` with `ON DELETE CASCADE`
- CHECK constraint ensures project-scoped prompts have `project_id`
- Global prompts have `NULL project_id`

**Core Fields:**
- `name`: Human-readable name for the prompt
- `slug`: URL-safe unique identifier (used for referencing in code/CLI)
- `description`: Optional documentation about the prompt's purpose
- `prompt_text`: The actual prompt content (TEXT type for large content)
- `active`: Soft delete flag (follows `projects.active` pattern)
- `version`: Allows versioning of prompts (incremented on prompt_text changes)
- `tags`: Comma-separated tags for categorization/filtering
- `created_by`: Optional attribution (could FK to agents table if needed)

**Foreign Keys:**
- `project_id REFERENCES projects(id) ON DELETE CASCADE`
  - When project is deleted, project-scoped prompts are removed
  - Global prompts remain unaffected
  - Aligns with existing project scoping patterns in tasks/commits

**Indexes:**
- `project_id`: For filtering prompts by project
- `scope`: For quick global vs project filtering
- `active`: For filtering active vs inactive prompts
- `slug`: For fast lookups by slug

**Trigger:**
- Auto-update `updated_at` on modifications (standard EITS pattern)

---

## 2. Go Database Functions (CRUD Operations)

### File: `internal/database/prompts.go` (new file)

### Data Structures

```go
package database

import "time"

// SubagentPrompt represents a reusable prompt template
type SubagentPrompt struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Slug        string    `json:"slug"`
    Description *string   `json:"description,omitempty"`
    PromptText  string    `json:"prompt_text"`
    ProjectID   *string   `json:"project_id,omitempty"`
    Scope       string    `json:"scope"`
    Active      bool      `json:"active"`
    Version     int       `json:"version"`
    Tags        *string   `json:"tags,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    CreatedBy   *string   `json:"created_by,omitempty"`
}

// CreateSubagentPromptParams contains parameters for creating a prompt
type CreateSubagentPromptParams struct {
    Name        string
    Slug        string
    Description *string
    PromptText  string
    ProjectID   *string
    Scope       string
    Tags        *string
    CreatedBy   *string
}

// UpdateSubagentPromptParams contains parameters for updating a prompt
type UpdateSubagentPromptParams struct {
    Name        *string
    Description *string
    PromptText  *string
    Active      *bool
    Tags        *string
    Version     *int
}

// ListSubagentPromptsParams contains parameters for listing prompts
type ListSubagentPromptsParams struct {
    ProjectID *string
    Scope     *string
    Active    *bool
    Limit     int
    Offset    int
}
```

### CRUD Functions

#### Create

```go
// CreateSubagentPrompt inserts a new prompt into the database
func (db *DB) CreateSubagentPrompt(params CreateSubagentPromptParams) (*SubagentPrompt, error) {
    // Generate UUID for id
    id := uuid.New().String()

    // Validate scope
    if params.Scope != "global" && params.Scope != "project" {
        return nil, fmt.Errorf("invalid scope: must be 'global' or 'project'")
    }

    // Validate scope constraints
    if params.Scope == "project" && params.ProjectID == nil {
        return nil, fmt.Errorf("project_id required for project-scoped prompts")
    }
    if params.Scope == "global" && params.ProjectID != nil {
        return nil, fmt.Errorf("project_id must be null for global prompts")
    }

    query := `
        INSERT INTO subagent_prompts (id, name, slug, description, prompt_text, project_id, scope, tags, created_by)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

    _, err := db.conn.Exec(query, id, params.Name, params.Slug, params.Description,
        params.PromptText, params.ProjectID, params.Scope, params.Tags, params.CreatedBy)
    if err != nil {
        return nil, fmt.Errorf("failed to create subagent prompt: %w", err)
    }

    return db.GetSubagentPromptByID(id)
}
```

#### Read (Single)

```go
// GetSubagentPromptByID retrieves a prompt by its ID
func (db *DB) GetSubagentPromptByID(id string) (*SubagentPrompt, error) {
    query := `
        SELECT id, name, slug, description, prompt_text, project_id, scope, active, version, tags, created_at, updated_at, created_by
        FROM subagent_prompts
        WHERE id = ?
    `

    var prompt SubagentPrompt
    row := db.conn.QueryRow(query, id)
    err := row.Scan(&prompt.ID, &prompt.Name, &prompt.Slug, &prompt.Description, &prompt.PromptText,
        &prompt.ProjectID, &prompt.Scope, &prompt.Active, &prompt.Version, &prompt.Tags,
        &prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("prompt not found: %s", id)
        }
        return nil, fmt.Errorf("failed to get prompt: %w", err)
    }

    return &prompt, nil
}

// GetSubagentPromptBySlug retrieves a prompt by slug with project scope priority
// If projectID is provided, project-scoped prompts take priority over global
func (db *DB) GetSubagentPromptBySlug(slug string, projectID *string) (*SubagentPrompt, error) {
    var query string
    var args []interface{}

    if projectID != nil {
        // Check project-scoped first, then fall back to global
        query = `
            SELECT id, name, slug, description, prompt_text, project_id, scope, active, version, tags, created_at, updated_at, created_by
            FROM subagent_prompts
            WHERE slug = ? AND active = 1
              AND (project_id = ? OR scope = 'global')
            ORDER BY (CASE WHEN project_id = ? THEN 0 ELSE 1 END)
            LIMIT 1
        `
        args = []interface{}{slug, *projectID, *projectID}
    } else {
        // Only global prompts
        query = `
            SELECT id, name, slug, description, prompt_text, project_id, scope, active, version, tags, created_at, updated_at, created_by
            FROM subagent_prompts
            WHERE slug = ? AND active = 1 AND scope = 'global'
            LIMIT 1
        `
        args = []interface{}{slug}
    }

    var prompt SubagentPrompt
    row := db.conn.QueryRow(query, args...)
    err := row.Scan(&prompt.ID, &prompt.Name, &prompt.Slug, &prompt.Description, &prompt.PromptText,
        &prompt.ProjectID, &prompt.Scope, &prompt.Active, &prompt.Version, &prompt.Tags,
        &prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("prompt not found: %s", slug)
        }
        return nil, fmt.Errorf("failed to get prompt by slug: %w", err)
    }

    return &prompt, nil
}
```

#### Read (List)

```go
// ListSubagentPrompts retrieves prompts with filtering and pagination
func (db *DB) ListSubagentPrompts(params ListSubagentPromptsParams) ([]*SubagentPrompt, error) {
    query := `
        SELECT id, name, slug, description, prompt_text, project_id, scope, active, version, tags, created_at, updated_at, created_by
        FROM subagent_prompts
        WHERE 1=1
    `
    args := []interface{}{}

    // Build WHERE clause
    if params.ProjectID != nil {
        query += " AND (project_id = ? OR scope = 'global')"
        args = append(args, *params.ProjectID)
    }
    if params.Scope != nil {
        query += " AND scope = ?"
        args = append(args, *params.Scope)
    }
    if params.Active != nil {
        query += " AND active = ?"
        args = append(args, *params.Active)
    }

    // Order and pagination
    query += " ORDER BY updated_at DESC"
    if params.Limit > 0 {
        query += " LIMIT ?"
        args = append(args, params.Limit)
    }
    if params.Offset > 0 {
        query += " OFFSET ?"
        args = append(args, params.Offset)
    }

    rows, err := db.conn.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to list prompts: %w", err)
    }
    defer rows.Close()

    prompts := []*SubagentPrompt{}
    for rows.Next() {
        var prompt SubagentPrompt
        err := rows.Scan(&prompt.ID, &prompt.Name, &prompt.Slug, &prompt.Description, &prompt.PromptText,
            &prompt.ProjectID, &prompt.Scope, &prompt.Active, &prompt.Version, &prompt.Tags,
            &prompt.CreatedAt, &prompt.UpdatedAt, &prompt.CreatedBy)
        if err != nil {
            return nil, fmt.Errorf("failed to scan prompt: %w", err)
        }
        prompts = append(prompts, &prompt)
    }

    return prompts, nil
}
```

#### Update

```go
// UpdateSubagentPrompt updates a prompt's fields
func (db *DB) UpdateSubagentPrompt(id string, params UpdateSubagentPromptParams) (*SubagentPrompt, error) {
    // Get current prompt to check if prompt_text changed
    current, err := db.GetSubagentPromptByID(id)
    if err != nil {
        return nil, err
    }

    updates := []string{}
    args := []interface{}{}

    // Build dynamic UPDATE
    if params.Name != nil {
        updates = append(updates, "name = ?")
        args = append(args, *params.Name)
    }
    if params.Description != nil {
        updates = append(updates, "description = ?")
        args = append(args, *params.Description)
    }
    if params.PromptText != nil {
        updates = append(updates, "prompt_text = ?")
        args = append(args, *params.PromptText)

        // Auto-increment version if prompt_text changed
        if *params.PromptText != current.PromptText {
            updates = append(updates, "version = version + 1")
        }
    }
    if params.Active != nil {
        updates = append(updates, "active = ?")
        args = append(args, *params.Active)
    }
    if params.Tags != nil {
        updates = append(updates, "tags = ?")
        args = append(args, *params.Tags)
    }

    if len(updates) == 0 {
        return current, nil // Nothing to update
    }

    args = append(args, id)
    query := fmt.Sprintf("UPDATE subagent_prompts SET %s WHERE id = ?", strings.Join(updates, ", "))

    _, err = db.conn.Exec(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to update prompt: %w", err)
    }

    return db.GetSubagentPromptByID(id)
}
```

#### Delete (Soft)

```go
// DeactivateSubagentPrompt soft-deletes a prompt by setting active = 0
func (db *DB) DeactivateSubagentPrompt(id string) error {
    query := "UPDATE subagent_prompts SET active = 0 WHERE id = ?"
    result, err := db.conn.Exec(query, id)
    if err != nil {
        return fmt.Errorf("failed to deactivate prompt: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("prompt not found: %s", id)
    }

    return nil
}
```

#### Delete (Hard)

```go
// DeleteSubagentPrompt permanently removes a prompt
func (db *DB) DeleteSubagentPrompt(id string) error {
    query := "DELETE FROM subagent_prompts WHERE id = ?"
    result, err := db.conn.Exec(query, id)
    if err != nil {
        return fmt.Errorf("failed to delete prompt: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("prompt not found: %s", id)
    }

    return nil
}
```

### Error Handling

- Use `sql.ErrNoRows` for not-found conditions
- Return wrapped errors with context using `fmt.Errorf`
- Validate constraints before database operations
- Use context for cancellation/timeout support (add ctx parameter)

---

## 3. MCP Tool Interface Design

### File: `internal/mcp/tools.go` (add to existing)

### Tool 1: `i-prompt-create`

**Description:** Create a new subagent prompt scoped globally or to a project

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Human-readable name for the prompt"
    },
    "slug": {
      "type": "string",
      "description": "URL-safe unique identifier (e.g., 'code-reviewer', 'bug-fixer')"
    },
    "prompt_text": {
      "type": "string",
      "description": "The actual prompt content"
    },
    "description": {
      "type": "string",
      "description": "Optional description of the prompt's purpose"
    },
    "scope": {
      "type": "string",
      "enum": ["global", "project"],
      "description": "Scope: 'global' or 'project'"
    },
    "project_id": {
      "type": "string",
      "description": "Project ID (required if scope=project)"
    },
    "tags": {
      "type": "string",
      "description": "Comma-separated tags for categorization"
    }
  },
  "required": ["name", "slug", "prompt_text", "scope"]
}
```

**Response:**
```json
{
  "success": true,
  "prompt_id": "uuid",
  "slug": "prompt-slug",
  "message": "Prompt 'Code Reviewer' created successfully"
}
```

**Handler:**
```go
func handlePromptCreate(db *database.DB, args map[string]interface{}) (interface{}, error) {
    // Parse and validate arguments
    name := args["name"].(string)
    slug := args["slug"].(string)
    promptText := args["prompt_text"].(string)
    scope := args["scope"].(string)

    var description, projectID, tags *string
    if val, ok := args["description"].(string); ok {
        description = &val
    }
    if val, ok := args["project_id"].(string); ok {
        projectID = &val
    }
    if val, ok := args["tags"].(string); ok {
        tags = &val
    }

    // Create prompt
    prompt, err := db.CreateSubagentPrompt(database.CreateSubagentPromptParams{
        Name:        name,
        Slug:        slug,
        Description: description,
        PromptText:  promptText,
        ProjectID:   projectID,
        Scope:       scope,
        Tags:        tags,
    })
    if err != nil {
        return map[string]interface{}{"success": false, "error": err.Error()}, nil
    }

    return map[string]interface{}{
        "success":   true,
        "prompt_id": prompt.ID,
        "slug":      prompt.Slug,
        "message":   fmt.Sprintf("Prompt '%s' created successfully", name),
    }, nil
}
```

### Tool 2: `i-prompt-get`

**Description:** Retrieve a prompt by slug or ID

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "slug": {
      "type": "string",
      "description": "Prompt slug (provide slug or id)"
    },
    "id": {
      "type": "string",
      "description": "Prompt ID (provide slug or id)"
    },
    "project_id": {
      "type": "string",
      "description": "Project ID for scoped slug lookup"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "prompt": {
    "id": "uuid",
    "name": "Code Reviewer",
    "slug": "code-reviewer",
    "prompt_text": "You are an expert code reviewer...",
    "scope": "global",
    "project_id": null,
    "active": true,
    "version": 1
  }
}
```

### Tool 3: `i-prompt-list`

**Description:** List available prompts with filtering

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "project_id": {
      "type": "string",
      "description": "Filter by project (includes global prompts)"
    },
    "scope": {
      "type": "string",
      "enum": ["global", "project"],
      "description": "Filter by scope"
    },
    "active": {
      "type": "boolean",
      "description": "Filter by active status (default: true)"
    },
    "limit": {
      "type": "integer",
      "description": "Maximum results (default: 50)"
    },
    "offset": {
      "type": "integer",
      "description": "Pagination offset (default: 0)"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "prompts": [
    {
      "id": "uuid",
      "name": "Code Reviewer",
      "slug": "code-reviewer",
      "description": "Reviews code for best practices",
      "scope": "global",
      "version": 1
    }
  ],
  "count": 1,
  "total": 10
}
```

### Tool 4: `i-prompt-update`

**Description:** Update an existing prompt

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "description": "Prompt ID to update"
    },
    "name": {
      "type": "string",
      "description": "Updated name"
    },
    "description": {
      "type": "string",
      "description": "Updated description"
    },
    "prompt_text": {
      "type": "string",
      "description": "Updated prompt text (increments version)"
    },
    "tags": {
      "type": "string",
      "description": "Updated tags"
    },
    "active": {
      "type": "boolean",
      "description": "Active status"
    }
  },
  "required": ["id"]
}
```

**Response:**
```json
{
  "success": true,
  "prompt_id": "uuid",
  "version": 2,
  "message": "Prompt updated successfully"
}
```

### Tool 5: `i-prompt-delete`

**Description:** Soft delete (deactivate) a prompt

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "description": "Prompt ID to deactivate"
    }
  },
  "required": ["id"]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Prompt deactivated successfully"
}
```

---

## 4. Integration with Existing Project Scoping Patterns

### Current EITS Project Scoping

**Project Model:**
- Primary Key: `id TEXT PRIMARY KEY` (UUID)
- Unique constraint on `path`
- Contains: name, slug, path, remote_url, branch, active flag

**Tables Using Project Scoping:**
1. `agents.project_id TEXT REFERENCES projects(id)` - nullable
2. `tasks.project_id TEXT REFERENCES projects(id)` - nullable
3. `commits.project_id TEXT REFERENCES projects(id)` - nullable

**Pattern:**
- Project references are **nullable** - allows entities to be global or project-scoped
- No explicit CASCADE on most project_id FKs (defaults to RESTRICT)
- Indexes created for project_id columns for query performance

### Subagent Prompts Alignment

**Matches Existing Patterns:**
1. **Data Type:** TEXT for project_id FK (UUID-based)
2. **Nullable FK:** Allows global (NULL) and project-scoped (NOT NULL) prompts
3. **Indexing:** `idx_subagent_prompts_project` follows naming convention
4. **Soft Delete:** `active BOOLEAN DEFAULT 1` (like projects table)

**Enhancements:**
1. **Explicit Scope Field:** `scope` enum ('global' or 'project') for clarity
2. **CHECK Constraints:** Ensures data integrity at database level
3. **CASCADE Behavior:** `ON DELETE CASCADE` for project_id FK
   - When project deleted, project-scoped prompts are removed
   - Global prompts remain unaffected

### Query Patterns for Scoped Prompts

**Get prompts for a project (including global):**
```sql
SELECT * FROM subagent_prompts
WHERE active = 1 AND (project_id = ? OR scope = 'global')
ORDER BY scope DESC, updated_at DESC
```

**Priority Resolution:**
- Project-scoped prompts take priority over global prompts with same slug
- `GetSubagentPromptBySlug()` checks project scope first, falls back to global

### Migration Path

**File:** `internal/database/migrations/005_add_subagent_prompts_table.sql`

```sql
-- Migration 005: Add subagent_prompts table for storing reusable prompt templates

CREATE TABLE IF NOT EXISTS subagent_prompts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT,
    prompt_text TEXT NOT NULL,
    project_id TEXT,
    scope TEXT NOT NULL DEFAULT 'global',
    active BOOLEAN DEFAULT 1,
    version INTEGER DEFAULT 1,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CHECK (scope IN ('global', 'project')),
    CHECK (scope = 'global' OR project_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_subagent_prompts_project ON subagent_prompts(project_id);
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_scope ON subagent_prompts(scope);
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_active ON subagent_prompts(active);
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_slug ON subagent_prompts(slug);

CREATE TRIGGER IF NOT EXISTS update_subagent_prompts_updated_at
AFTER UPDATE ON subagent_prompts
FOR EACH ROW
BEGIN
    UPDATE subagent_prompts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
```

---

## 5. Implementation Checklist

### Phase 1: Database Layer
- [ ] Create migration file: `005_add_subagent_prompts_table.sql`
- [ ] Add table schema to `internal/database/schema.sql`
- [ ] Create `internal/database/prompts.go` with data structures
- [ ] Implement CRUD functions in `prompts.go`
- [ ] Write unit tests for database functions
- [ ] Run migration and verify schema

### Phase 2: MCP Tools
- [ ] Add tool handlers to `internal/mcp/tools.go`
- [ ] Register tools with MCP server in `internal/mcp/server.go`
- [ ] Test tools via Claude Code integration
- [ ] Document tools in CLAUDE.md

### Phase 3: TUI Integration (Optional)
- [ ] Add "Prompts" tab to TUI dashboard
- [ ] Implement list view for prompts
- [ ] Add create/edit/delete UI flows
- [ ] Add filtering by project/scope/active

### Phase 4: Documentation
- [ ] Update CLAUDE.md with new MCP tools
- [ ] Add usage examples to documentation
- [ ] Update API reference if applicable

---

## 6. Testing Strategy

### Unit Tests
- Test CRUD operations for all functions
- Test constraint validation (scope, project_id)
- Test slug uniqueness
- Test project-scoped vs global priority
- Test soft delete vs hard delete

### Integration Tests
- Test MCP tool handlers end-to-end
- Test project cascade deletion
- Test version incrementing on prompt_text changes

### Manual Testing
- Create global prompt via MCP tool
- Create project-scoped prompt
- Retrieve prompts by slug with project context
- List prompts filtered by project
- Update prompt and verify version increment
- Soft delete and verify it's excluded from active queries

---

## 7. Future Enhancements

### Full-Text Search
- Add FTS5 virtual table for prompt search (similar to `task_search`)
- Enable search across name, description, prompt_text, tags

### Prompt Variables
- Support template variables in prompt_text (e.g., `{{project_name}}`)
- Implement variable substitution at retrieval time

### Usage Tracking
- Track which prompts are used by which sessions
- Add analytics for prompt effectiveness

### Prompt Chaining
- Allow prompts to reference other prompts
- Build composite prompts from smaller reusable pieces

### Import/Export
- Export prompts to YAML/JSON files
- Import prompts from files or repositories
- Share prompts across EITS instances

---

## 8. Known Issues and Considerations

### Type Mismatch Bug
**Issue:** `Agent.ProjectID` in `models.go` is defined as `*int` but `schema.sql` defines `project_id` as `TEXT`.

**Impact:** Runtime errors when reading/writing agent project associations.

**Recommendation:** Change `models.go` line 20 from:
```go
ProjectID *int `json:"project_id,omitempty"` // Foreign key to projects.id
```
to:
```go
ProjectID *string `json:"project_id,omitempty"` // Foreign key to projects.id
```

### Slug Uniqueness
The `slug` field has a UNIQUE constraint, which means you cannot have the same slug for both a global and project-scoped prompt. This is intentional to prevent ambiguity, but worth documenting.

### CASCADE Considerations
The `ON DELETE CASCADE` on `project_id` means deleting a project will permanently remove all project-scoped prompts. Ensure this behavior is acceptable or implement a soft-delete project pattern.

---

## 9. Success Metrics

- [ ] All CRUD operations work correctly
- [ ] MCP tools respond within 500ms
- [ ] Schema migration runs without errors
- [ ] Unit tests pass with >90% coverage
- [ ] Documentation is clear and complete
- [ ] Integration with existing project scoping is seamless

---

## Appendix: Example Usage

### Creating a Global Code Reviewer Prompt

```
i-prompt-create({
  "name": "Code Reviewer",
  "slug": "code-reviewer",
  "prompt_text": "You are an expert code reviewer. Analyze the provided code for:\n- Best practices\n- Security vulnerabilities\n- Performance issues\n- Code style consistency\n\nProvide specific, actionable feedback.",
  "description": "Reviews code for quality, security, and performance",
  "scope": "global",
  "tags": "code-review,quality,security"
})
```

### Creating a Project-Specific Bug Fixer

```
i-prompt-create({
  "name": "EITS Bug Fixer",
  "slug": "eits-bug-fixer",
  "prompt_text": "You are debugging the Eye in the Sky codebase. Focus on:\n- Go error handling patterns\n- SQLite query issues\n- MCP tool integration bugs\n\nRefer to CLAUDE.md for project-specific context.",
  "description": "Specialized bug fixer for EITS codebase",
  "scope": "project",
  "project_id": "project-uuid-here",
  "tags": "debugging,eits,go"
})
```

### Retrieving a Prompt by Slug

```
i-prompt-get({
  "slug": "code-reviewer",
  "project_id": "project-uuid-here"
})
```

Response:
```json
{
  "success": true,
  "prompt": {
    "id": "uuid",
    "name": "Code Reviewer",
    "slug": "code-reviewer",
    "prompt_text": "You are an expert code reviewer...",
    "scope": "global",
    "project_id": null,
    "active": true,
    "version": 1
  }
}
```

### Listing All Active Prompts for a Project

```
i-prompt-list({
  "project_id": "project-uuid-here",
  "active": true,
  "limit": 50
})
```

---

## Summary

This implementation plan provides a comprehensive, production-ready design for adding subagent prompts to Eye in the Sky. The design:

1. **Follows EITS patterns:** UUID PKs, nullable project FKs, soft deletes, auto-update triggers
2. **Supports flexible scoping:** Global and project-scoped prompts with priority resolution
3. **Provides complete CRUD:** Create, Read, Update, Delete with proper validation
4. **Integrates seamlessly:** MCP tools expose functionality to Claude Code
5. **Maintains data integrity:** CHECK constraints, foreign keys, cascading deletes
6. **Enables future growth:** Versioning, tagging, extensibility

The schema is ready for implementation and testing.
