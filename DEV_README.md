# Eye in the Sky - Developer Guide

This guide covers setup, architecture, and development workflows for contributors.

## Project Structure

```
eye-in-the-sky/
├── cmd/
│   ├── eye-ui/main.go             # TUI dashboard executable
│   └── server/main.go             # MCP server executable
├── internal/
│   ├── mcp/                       # MCP server implementation
│   ├── ui/                        # TUI dashboard (Bubble Tea)
│   │   ├── app/                   # Main application model and logic
│   │   ├── components/            # Reusable UI components
│   │   ├── keybindings/           # Keyboard shortcuts
│   │   ├── modal/                 # Modal dialogs
│   │   ├── services/              # Business logic services
│   │   ├── util/                  # Utility functions
│   │   └── viewmodel/             # View model representations
│   ├── database/                  # SQLite connection and queries
│   ├── data/                      # Data access layer (stores)
│   ├── domain/                    # Domain models (Agent, Task, etc.)
│   ├── ccusage/                   # Claude Code usage data
│   ├── utils/                     # Shared utilities
│   └── window/                    # Window management utilities
├── tests/                         # Go test files
├── CLAUDE.md                      # Technical guide for Claude Code
├── DEV_README.md                  # This file
└── README.md                      # User-facing documentation
```

## Build Instructions

### Building the TUI Dashboard

```bash
# Build the TUI
go build -o bin/eye-ui ./cmd/eye-ui

# Run the TUI
./bin/eye-ui
```

### Building the MCP Server

```bash
# Build the MCP server
go build -o bin/eye-in-the-sky ./cmd/server

# Run the MCP server
./bin/eye-in-the-sky
```

## Database

### Location
SQLite database is stored at: `~/.config/eye-in-the-sky/agents.db`

### Schema Initialization
The database schema is automatically created on first run via `internal/database/schema.go`.

### Reset Database
```bash
rm ~/.config/eye-in-the-sky/agents.db
```

The database will be recreated with fresh schema on next startup.

### Core Tables

| Table | Purpose |
|-------|---------|
| `agents` | Track Claude Code instances (worktrees, desktop) |
| `actions` | Log all agent activities |
| `commits` | Track git commits made by agents |
| `notes` | Session notes created by agents |
| `logs` | Session logs |
| `metrics` | Session metrics and performance data |

## Logging

### TUI Logging

Logs are written to `~/.config/eye-in-the-sky/tui.log`

**Configuration** (`cmd/eye-ui/main.go`):
- File is truncated on startup to avoid log bloat
- Uses `SyncWriter` to ensure immediate flushing
- Includes timestamps and source file names

**Example log entry:**
```
2025/10/25 13:05:30 main.go:51: === TUI Dashboard Started ===
```

**Important**: Ensure only a single writer is used to avoid duplicate log entries. The `SyncWriter` wraps the file and handles both writing and flushing.

### Viewing Logs

```bash
# Watch logs in real-time
tail -f ~/.config/eye-in-the-sky/tui.log

# Search logs for specific errors
grep -i error ~/.config/eye-in-the-sky/tui.log
```

## Testing

### Run All Tests
```bash
go test ./...
```

### Run with Coverage
```bash
go test -cover ./...
```

### Run Specific Package Tests
```bash
go test ./internal/database
go test ./internal/ui/app
go test ./internal/mcp
```

### Run with Verbose Output
```bash
go test -v ./...
```

### Test Tagging
Tests use `build` tags for platform-specific functionality:
```bash
go test -tags=darwin ./internal/window
go test -tags=linux ./internal/window
```

## Data Loading & Refresh

### Startup Flow

1. **Initialization** (`cmd/eye-ui/main.go`)
   - Sets up logging
   - Connects to SQLite database
   - Creates application model

2. **Model Init** (`internal/ui/app/model.go::Init()`)
   - Returns async commands via `tea.Batch()`:
     - `loadAgentsCmd()` - Load agents from database
     - `loadClaudeFilesCmd()` - Load Claude config files
     - `tickCmd()` - Start refresh ticker

3. **Async Command Pattern**
   - All data loads use Bubble Tea's command system
   - Commands emit messages that update state
   - UI stays responsive during I/O operations

### Refresh Mechanisms

**Time-based** (configurable interval, default 5-30 seconds)
- `tickCmd()` triggers periodic `loadAgentsCmd()`
- CCUsage data syncs every 30 seconds if available

**Manual** (user-triggered)
- Press `r` to refresh agent list
- Toggle "show all" to reload agents

**Event-based**
- Creating/archiving agents triggers `loadAgentsCmd()`
- Tab switching triggers appropriate loaders

### Async Command Examples

All defined in `internal/ui/app/cmd_patterns.go`:

| Command | Triggers | Returns |
|---------|----------|---------|
| `loadAgentsCmd()` | Init, refresh, filters | `AgentsLoadedMsg` |
| `loadAgentDetailsCmd()` | User selects agent | `AgentDetailsLoadedMsg` |
| `loadTasksCmd()` | Switch to tasks tab | `TasksLoadedMsg` |
| `loadCommitsCmd()` | Switch to commits tab | `CommitsLoadedMsg` |
| `loadActionsCmd()` | Switch to actions tab | `ActionsLoadedMsg` |
| `loadNotesCmd()` | Switch to notes tab | `NotesLoadedMsg` |

## Development Workflow

### Task Tracking
This project uses Taskwarrior for persistent task management:

```bash
# Create a task
task add "Fix login bug" project:eye-in-the-sky priority:H +bug

# Start working on it
task 1 start

# Add progress notes
task 1 annotate "Fixed password validation in auth.go"

# Mark as done
task 1 done
```

See CLAUDE.md for complete Taskwarrior integration guide.

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/my-feature

# Make changes and test
go test ./...

# Commit with descriptive message
git commit -m "feat: Add new feature

Detailed explanation of what changed and why."

# Push to remote
git push origin feature/my-feature
```

### Common Development Tasks

#### Adding a New Tab to TUI

1. Define tab in `internal/ui/app/model.go::NewModel()`
2. Create load command in `internal/ui/app/cmd_patterns.go`
3. Add message type in `internal/ui/app/model.go`
4. Handle message in `Update()` method
5. Add render function for the tab
6. Update keyboard navigation in `Update()`

#### Adding a New Database Store

1. Define domain model in `internal/domain/`
2. Create store interface in `internal/data/`
3. Implement store with database queries in `internal/data/`
4. Add to `DataClient` in `internal/ui/app/client.go`
5. Create async command in `internal/ui/app/cmd_patterns.go`
6. Wire into UI via `Update()` handler

#### Debugging Data Loading

Enable debug logging:
```bash
# Add debug output to cmd_patterns.go
log.Printf("DEBUG: Loading agents...")
log.Printf("DEBUG: Loaded %d agents", len(agents))

# Rebuild and run
go build -o bin/eye-ui ./cmd/eye-ui
./bin/eye-ui

# Check logs
tail -f ~/.config/eye-in-the-sky/tui.log
```

## Performance Considerations

- **TUI Rendering**: Should complete within 100ms
- **MCP Tools**: Must respond within 500ms
- **Database Operations**: Use timeouts (default 5 seconds)
- **Refresh Interval**: Configurable, typically 5-30 seconds

## Common Issues & Solutions

### Duplicate Logs in Startup
**Cause**: Using `io.MultiWriter(f, syncWriter)` where `syncWriter` already wraps `f`

**Solution**: Use only `syncWriter`:
```go
log.SetOutput(syncWriter)  // ✅ Correct
log.SetOutput(io.MultiWriter(f, syncWriter))  // ❌ Causes duplicates
```

### Database Lock Errors
The SQLite database may lock if multiple processes access it simultaneously.

**Solution**: Ensure only one TUI instance runs at a time, or use WAL mode (Write-Ahead Logging).

### Memory Leaks in Long Sessions
Large data loads without cleanup can accumulate in memory.

**Solution**: Implement pagination for large datasets and clean up old data periodically.

## Architecture Notes

### MCP Integration
The system uses the Model Context Protocol (MCP) to integrate with Claude Code. MCP tools are defined in `internal/mcp/tools.go` and are called directly from Claude Code.

### Event-Driven Updates
The TUI uses Bubble Tea's message-based event system. All updates flow through the `Update()` method, making the code predictable and testable.

### Data Consistency
All database operations use transactions where multiple tables are modified, ensuring atomic updates and consistency.

### Caching & Dirty Flags
The model uses dirty flags (`overviewDirty`, `usageDirty`) to optimize re-rendering by avoiding unnecessary rebuilds.

## Useful Commands

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run ./...

# Check for issues
go vet ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Find TODOs and FIXMEs
grep -r "TODO\|FIXME" internal/ cmd/

# Count lines of code
wc -l $(find . -name "*.go" -type f)
```

## IDE Setup

### VS Code
Install extensions:
- Go (golang.go)
- Gofmt (golang.gofmt)

Settings:
```json
{
  "go.lintOnSave": "package",
  "go.useLanguageServer": true,
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### GoLand / IntelliJ
Built-in Go support is excellent. Enable:
- Go code inspections
- Run tests from editor
- Debug mode integration

## References

- **[Go Documentation](https://golang.org/doc/)**
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling
- **[SQLite Driver](https://modernc.org/sqlite)** - Database driver
- **[MCP Specification](https://modelcontextprotocol.io/)** - Protocol spec
- **[Taskwarrior](https://taskwarrior.org/)** - Task management

---

**Last Updated**: October 25, 2025

For user documentation, see [README.md](README.md). For Claude Code integration details, see [CLAUDE.md](CLAUDE.md).
