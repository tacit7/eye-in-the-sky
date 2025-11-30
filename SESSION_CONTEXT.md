# Session Context: Project Detail View & Database Integration

**Date:** 2025-11-27
**Session ID:** 534002f0
**Current Agent:** dc6bbeef-6592-4ad2-9d03-5d7df8ae4426

## Overview

This session focused on implementing the Project Detail View with Bubble Tea table components and completing database integration for project-agent relationships.

---

## Completed Work

### 1. Database Migration: Project-Agent Linking

**Problem:** Agents table used `project_name` (string) instead of proper foreign key relationship.

**Solution:** Added `project_id` foreign key to agents table.

**Files Modified:**
- `internal/database/models.go` - Added `ProjectID *string` field to Agent model
- `internal/database/schema.sql` - Added `project_id TEXT REFERENCES projects(id)` column
- `internal/database/queries.go` - Updated all agent queries to include project_id

**Migration Created:**
```sql
-- internal/database/migrations/004_add_project_id_to_agents.sql
ALTER TABLE agents ADD COLUMN project_id TEXT REFERENCES projects(id);
CREATE INDEX IF NOT EXISTS idx_agents_project_id ON agents(project_id);
UPDATE agents SET project_id = (
    SELECT id FROM projects WHERE name = agents.project_name LIMIT 1
) WHERE project_name IS NOT NULL AND project_id IS NULL;
```

**New Database Queries:**
- `GetActiveAgentsByProject(projectID string)` - Fetch active agents filtered by project

**Query Updates:**
- `CreateAgent()` - Now inserts project_id
- `GetAgent()` - Now selects project_id
- `GetAgentBySessionID()` - Now selects project_id
- `ListAgents()` - Now selects project_id in all three query variants

---

### 2. Project Detail View with Bubble Tea Components

**Architecture Decision:** Use Bubble Tea `table.Model` components instead of pure functions

**Rationale:**
- User requested stateful Bubble Tea components
- Built-in features: selection, navigation, sorting
- Applied to all split-pane tabs (Agents, Notes, Tasks)

#### Model Layer Changes

**File:** `internal/ui/app/views/project/model_project_details.go`

**Added imports:**
```go
import "github.com/charmbracelet/bubbles/table"
```

**Added table model fields:**
```go
type Model struct {
    // ... existing fields ...
    agentsTable table.Model
    notesTable  table.Model
    tasksTable  table.Model
}
```

**New methods:**
- `createAgentsTable()` - Builds table with columns: ID, Status, Description, Session
- `createNotesTable()` - Placeholder for notes table
- `createTasksTable()` - Placeholder for tasks table
- `initializeTables()` - Initializes all tables from context data

**Table Configuration:**
```go
columns := []table.Column{
    {Title: "ID", Width: 10},
    {Title: "Status", Width: 10},
    {Title: "Description", Width: 30},
    {Title: "Session", Width: 12},
}
```

**Lifecycle Hooks:**
- Updated `SetContext()` to call `initializeTables()`
- Updated `SetSize()` to reinitialize tables with new dimensions

#### Update Handler Changes

**File:** `internal/ui/app/views/project/update_project_details.go`

**Key change:** Delegates navigation to `table.Update()`
```go
case "j", "down", "k", "up":
    switch m.tabs.ActiveIndex {
    case tabAgents:
        var cmd tea.Cmd
        m.agentsTable, cmd = m.agentsTable.Update(msg)
        // Sync selection index with table cursor
        m.ctx.SelectedAgentIndex = m.agentsTable.Cursor()
        m.ctx.RightPaneOffset = 0
        return m, cmd
    // ... similar for notes and tasks ...
    }
```

#### View Renderer Changes

**File:** `internal/ui/app/views/project/view_project_details.go`

**Updated signature:**
```go
func (m *Model) renderAgentsTab() string {
    return tabs.RenderAgents(m.ctx, m.styles.OverviewStyles, m.agentsTable, m.width, m.height)
}
```

#### Agents Tab Implementation

**File:** `internal/ui/app/views/project/tabs/agents_tab.go`

**Complete rewrite** with split-pane layout:

**Left pane:** `table.Model.View()` showing agent list table
**Right pane:** Agent details (ID, status, description, session, paths, timestamps)
**Separator:** `theme.PanelSidebar` vertical divider

**Layout:**
```go
result := lipgloss.JoinHorizontal(
    lipgloss.Top,
    leftPane,                                           // table.View()
    theme.PanelSidebar.Copy().Height(height).Render(""),  // divider
    rightPane,                                          // agent details
)
```

**Components used:**
- `lipgloss.JoinHorizontal()` for split-pane layout
- `theme.PanelNormal` for left pane container
- `theme.PanelNoBorder` for right pane container
- `theme.TextMuted` for empty states
- `renderField()` helper for label-value pairs

**Agent details shown:**
- ID, Status, Description
- Session ID, Project Name
- Git Worktree Path
- Current Task
- Last Activity timestamp
- Created timestamp

---

### 3. Project Detail View Integration

**File:** `internal/ui/app/view_project_detail.go`

**Updates:**
- Uses `GetProjectByPath()` to look up project in database
- Falls back to runtime-detected `projectInfo` if project not in DB
- Uses `GetActiveAgentsByProject()` to fetch agents with proper FK filtering
- Converts database.Agent → domain.Agent via `convertToDomainAgent()`

**Conversion Helper:**
```go
func convertToDomainAgent(dbAgent *database.Agent) domain.Agent {
    agent := domain.Agent{
        ID:         domain.AgentID(dbAgent.ID),
        Status:     dbAgent.Status,
        // ... etc ...
    }
    // Convert pointer fields to non-pointer with nil checks
    if dbAgent.GitWorktreePath != nil {
        agent.GitWorktreePath = *dbAgent.GitWorktreePath
    }
    // ... etc ...
    return agent
}
```

---

### 4. Dependency Addition

**Added:** `github.com/charmbracelet/bubbles/table`

**Impact:** First stateful Bubble Tea component in the codebase (introduces hybrid architecture)

**Installation:**
```bash
go get github.com/charmbracelet/bubbles/table
```

---

## Architecture Patterns

### Hybrid Architecture
- **Stateful components** (`table.Model`) for left pane lists
- **Pure rendering** (Lipgloss) for right pane details
- **Parent model** manages table state and delegates updates

### Table Model Lifecycle
1. **Initialization:** `createAgentsTable()` called in `SetContext()` and `SetSize()`
2. **Update:** Parent delegates navigation keys to `table.Update()`
3. **Render:** `table.View()` embedded in split-pane layout
4. **Selection sync:** Table cursor synced with context selection indices

### Data Flow
```
Database Query:
  GetActiveAgentsByProject(projectID)
    ↓
  []database.Agent
    ↓
  convertToDomainAgent() for each
    ↓
  []domain.Agent → DataContext.Agents

Table Creation:
  DataContext.Agents
    ↓
  createAgentsTable()
    ↓
  table.Model

Rendering:
  table.Model.View()
    ↓
  Split-pane layout (lipgloss.JoinHorizontal)
    ↓
  Final string output
```

---

## Current Issues

### ⚠️ Schema Mismatch

**Problem:** Type inconsistency between agents and projects tables
- `agents.project_id` is **TEXT** (from migration 004)
- `projects.id` is **INTEGER** (existing schema)

**Current Database State:**
```
Latest Agent:
  id: dc6bbeef-6592-4ad2-9d03-5d7df8ae4426
  project_name: eye-in-the-sky
  project_id: eits-project-1 (TEXT - no matching project!)

Projects Table:
  id: 1 (INTEGER)
  name: test-migration-project
  path: (null)
```

**Impact:**
- Agent references non-existent project (foreign key not enforced)
- `GetActiveAgentsByProject()` won't work correctly
- Schema inconsistency blocks proper relationships

**Resolution Options:**
1. Create "eye-in-the-sky" project with TEXT id = "eits-project-1"
2. Update agent to use INTEGER project_id = 1
3. Migrate projects.id to TEXT (requires migration + schema changes)
4. Migrate agents.project_id to INTEGER (requires migration + schema changes)

---

## Files Created/Modified Summary

### Created (1 file)
- `internal/database/migrations/004_add_project_id_to_agents.sql`

### Modified (8 files)
1. `internal/database/models.go` - Added ProjectID field
2. `internal/database/schema.sql` - Added project_id column to agents
3. `internal/database/queries.go` - Updated all agent queries
4. `internal/ui/app/views/project/model_project_details.go` - Added table models
5. `internal/ui/app/views/project/update_project_details.go` - Table update delegation
6. `internal/ui/app/views/project/view_project_details.go` - Pass table to renderer
7. `internal/ui/app/views/project/tabs/agents_tab.go` - Complete rewrite with table
8. `internal/ui/app/view_project_detail.go` - DB integration + conversion helper

### Dependencies Added
- `github.com/charmbracelet/bubbles/table`

---

## Build Status

✅ **Compiles successfully** with no errors

```bash
go build -o bin/eye-ui ./cmd/eye-ui
# Success - no errors
```

---

## Key Decisions

### 1. Use Bubble Tea Components (Not Pure Functions)
**Decision:** Integrate `table.Model` from bubbles library
**Rationale:** User explicitly requested Bubble Tea components for built-in features
**Trade-off:** Breaks pure function pattern, introduces stateful complexity
**Impact:** Hybrid architecture - stateful tables + pure detail rendering
**Reversible:** Yes, but would require rewriting tabs

### 2. Apply Table Pattern to All Split-Pane Tabs
**Decision:** Use table.Model for Agents, Notes, and Tasks tabs
**Rationale:** Consistency across similar UI patterns
**Trade-off:** More state to manage in parent model
**Impact:** Uniform UX, but increased complexity
**Reversible:** Yes, can implement each tab differently

### 3. Keep project_name Alongside project_id
**Decision:** Maintain backward compatibility with project_name column
**Rationale:** Existing agents rely on project_name
**Trade-off:** Data duplication, potential inconsistency
**Impact:** Migration can backfill without breaking existing code
**Reversible:** Yes, can drop column in future migration

### 4. Store Table Models in Parent Model
**Decision:** Add table fields to `Model` struct instead of tabs
**Rationale:** Tables need state persistence across renders
**Trade-off:** Tabs no longer fully stateless
**Impact:** Parent model grows, but tables work correctly
**Reversible:** Could move to separate tab models

---

## Testing Checklist

### Database
- [ ] Verify migration 004 creates project_id column
- [ ] Verify index on agents(project_id) exists
- [ ] Test GetActiveAgentsByProject() returns correct agents
- [ ] Test with agents that have project_id set
- [ ] Test with agents that have NULL project_id

### UI - Agents Tab
- [ ] Navigate to Project Detail view (press P)
- [ ] Switch to Agents tab (press A)
- [ ] Verify table displays with correct columns
- [ ] Test j/k navigation in agent list
- [ ] Verify right pane updates with selected agent details
- [ ] Test with empty agent list
- [ ] Test with single agent
- [ ] Test with multiple agents

### Integration
- [ ] Test full flow: Overview → Project Detail → Agents tab → Back
- [ ] Verify tab switching (O, A, N, T, F keys)
- [ ] Test window resize updates table dimensions
- [ ] Verify context updates when switching tabs

---

## Next Steps

### Immediate Priority
1. **Resolve schema mismatch** - Decide on INTEGER vs TEXT for project IDs
2. **Create eye-in-the-sky project** - Add proper project entry to database
3. **Test agents tab** - Run TUI and verify functionality

### Pending Implementation
4. Implement `notes_tab.go` with table.Model (similar to agents)
5. Implement `tasks_tab.go` with table.Model (similar to agents)
6. Implement `files_tab.go` with ranger/lf integration
7. Test complete navigation flow
8. Build and verify all features

---

## Code Snippets Reference

### Creating a Table Model
```go
func (m *Model) createAgentsTable() table.Model {
    columns := []table.Column{
        {Title: "ID", Width: 10},
        {Title: "Status", Width: 10},
        {Title: "Description", Width: 30},
        {Title: "Session", Width: 12},
    }

    rows := []table.Row{}
    for _, agent := range m.ctx.Agents {
        rows = append(rows, table.Row{
            string(agent.ID),
            agent.Status,
            agent.AgentDescription,
            agent.SessionID,
        })
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(m.height - 5),
    )

    // Apply theme styles
    s := table.DefaultStyles()
    s.Header = s.Header.Bold(true)
    s.Selected = s.Selected.Bold(true).Foreground(primaryColor)
    t.SetStyles(s)

    return t
}
```

### Delegating Updates to Table
```go
case "j", "down", "k", "up":
    var cmd tea.Cmd
    m.agentsTable, cmd = m.agentsTable.Update(msg)
    m.ctx.SelectedAgentIndex = m.agentsTable.Cursor()
    m.ctx.RightPaneOffset = 0
    return m, cmd
```

### Split-Pane Rendering
```go
leftPane := theme.PanelNormal.Copy().
    Width(width/2).
    Height(height).
    Render(agentsTable.View())

rightPane := theme.PanelNoBorder.Copy().
    Width(width/2).
    Height(height).
    Render(agentDetailsString)

divider := theme.PanelSidebar.Copy().Height(height).Render("")

return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, divider, rightPane)
```

---

## References

### Documentation
- **Plan file:** `/Users/urielmaldonado/.claude/plans/velvety-squishing-wirth.md`
- **Database:** `~/.config/eye-in-the-sky/eits.db`
- **CLAUDE.md:** Project guidelines and architecture

### Key Code Locations
- Agent model: `internal/database/models.go:5-25`
- Project queries: `internal/database/queries.go:1247-1337`
- Agents tab: `internal/ui/app/views/project/tabs/agents_tab.go`
- Table creation: `internal/ui/app/views/project/model_project_details.go:101-152`
- Update delegation: `internal/ui/app/views/project/update_project_details.go:37-64`

### Bubble Tea Components
- **Table:** https://github.com/charmbracelet/bubbles/tree/master/table
- **Lipgloss:** https://github.com/charmbracelet/lipgloss

---

## Session Metrics

- **Files Modified:** 8
- **Lines Added:** ~450
- **Lines Removed:** ~30
- **Migration Files:** 1
- **Database Queries:** 5 (1 new, 4 updated)
- **Bubble Tea Components:** 3 tables added
- **Build Status:** ✅ Success
- **Tests Written:** 0 (manual testing pending)

---

*Generated by Claude Code - Session 534002f0*
*Last Updated: 2025-11-27*
