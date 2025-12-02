# Migration Guide: Clean Architecture Refactoring

## Overview
This guide explains how to migrate the Eye in the Sky codebase from embedded SQL to clean architecture with separated data layers.

## Architecture Changes

### Before (Embedded SQL)
```
internal/ui/app/
  model.go         (1000+ lines: Model + SQL + types)
  update.go        (SQL queries mixed with UI logic)
  view_tasks.go    (TaskWarrior calls + rendering)
```

### After (Clean Architecture)
```
internal/
  domain/          (Pure domain types)
    agent.go
    task.go
    ...
  ui/app/
    stores.go      (Interface definitions)
    client.go      (DataClient aggregator)
    model.go       (UI state only)
    cmd_patterns.go (Bubble Tea commands)
  data/
    agents.go      (SQL implementation)
    tasks.go       (TaskWarrior implementation)
    ...
```

## Migration Steps

### Step 1: Replace Type References
Replace local types with domain types throughout the codebase:

```go
// Before
type Agent struct { ... }  // in model.go

// After
import "github.com/tacit7/eye-in-the-sky/internal/domain"
// Use domain.Agent
```

### Step 2: Update Model Initialization
Replace database connection with DataClient:

```go
// Before
func NewModel(db *sql.DB, config Config) *Model {
    return &Model{
        db: db,
        // ...
    }
}

// After
func NewModel(db *sql.DB, config Config) *Model {
    return &Model{
        data: NewDataClient(db),
        // ...
    }
}
```

### Step 3: Replace SQL Queries
Convert direct SQL queries to store method calls:

```go
// Before (in loadAgents)
rows, err := m.db.Query(query)
// ... manual scanning ...

// After
return loadAgentsCmd(m.data.Agents)
```

### Step 4: Update Update() Method
Replace blocking operations with commands:

```go
// Before
if err := m.loadAgents(); err != nil {
    m.err = err
}

// After
case "r":
    return m, loadAgentsCmd(m.data.Agents)

case AgentsLoadedMsg:
    m.agents = msg.Agents
    return m, nil
```

### Step 5: Fix Import Cycles
If you encounter import cycles:
1. Ensure domain package has no dependencies
2. Interfaces stay in ui/app
3. Data layer doesn't import ui

### Step 6: Update TaskWarrior Integration
The new TaskStore properly streams tasks:

```go
// Before (in loadTasks)
cmd := exec.Command("task", "export")
output, _ := cmd.Output()  // Loads everything into memory

// After (in data/tasks.go)
stdout, _ := cmd.StdoutPipe()
dec := json.NewDecoder(stdout)  // Streams results
```

## Common Patterns

### Loading Data
```go
// Command pattern for async loading
func loadSomethingCmd(store SomeStore) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := withTimeout()
        defer cancel()

        data, err := store.LoadSomething(ctx)
        if err != nil {
            return ErrMsg{Error: err}
        }
        return SomethingLoadedMsg{Data: data}
    }
}
```

### Handling Messages
```go
case SomethingLoadedMsg:
    m.something = msg.Data
    // Optionally trigger next load
    return m, nextCommand()
```

### Error Handling
```go
case ErrMsg:
    // UI decides how to present errors
    if errors.Is(msg.Error, context.DeadlineExceeded) {
        m.statusMsg = "Request timed out"
    } else {
        m.statusMsg = msg.Error.Error()
    }
    return m, nil
```

## Testing

### Mocking Stores
```go
mock := &MockTaskStore{
    LoadByAgentFunc: func(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
        return testTasks, nil
    },
}

// Use mock in tests
client := &DataClient{Tasks: mock}
```

### Table-Driven Tests
See `internal/data/tasks_test.go` for examples.

## Checklist

- [ ] Replace all local type definitions with domain types
- [ ] Update Model struct to use DataClient
- [ ] Convert loadAgents() to use AgentStore
- [ ] Convert loadTasks() to use TaskStore
- [ ] Convert loadCommits() to use CommitsStore
- [ ] Convert loadNotes() to use NotesStore
- [ ] Convert loadMetrics() to use MetricsStore
- [ ] Update all Update() cases to handle new message types
- [ ] Fix any import cycles
- [ ] Add tests for critical paths
- [ ] Remove old embedded SQL code
- [ ] Update views to handle domain types

## Files to Update

1. `internal/ui/app/model.go` - Use model_refactored.go as reference
2. `internal/ui/app/update.go` - Replace SQL with commands
3. `internal/ui/app/update_*.go` - Update message handling
4. `internal/ui/app/view.go` - Use domain types
5. `internal/ui/app/view_*.go` - Use domain types

## Validation

After migration:
1. Run `go build ./...` - Should compile without errors
2. Run `go test ./...` - Tests should pass
3. Start TUI - Should display agents and tasks correctly
4. Test refresh (r key) - Should update data
5. Test detail view - Should load all tabs

## Benefits After Migration

✅ Testable - Easy to mock stores
✅ Maintainable - Clear separation of concerns
✅ Performant - Streaming for large datasets
✅ Extensible - Easy to add new stores
✅ Type-safe - Domain types prevent errors
✅ Clean - No 1000+ line files