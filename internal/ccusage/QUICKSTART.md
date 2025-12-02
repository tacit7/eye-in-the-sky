# CCUsage Package - Quick Start

## Installation

The ccusage package is part of the Eye-in-the-Sky project and is automatically available when building the project.

## Basic Usage

### 1. Initialize Database

```go
import "github.com/tacit7/eye-in-the-sky/internal/ccusage/db"

dbPath := "/path/to/ccusage.sqlite"
database, err := db.New(dbPath)
if err != nil {
    log.Fatal(err)
}
defer database.Close()
```

### 2. Sync Data from JSONL Files

```go
import "github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"

syncMgr := parser.NewSyncManager(database)

// Incremental sync (only changed files)
if err := syncMgr.Sync(); err != nil {
    log.Fatal(err)
}

// Or full sync (all files)
if err := syncMgr.FullSync(); err != nil {
    log.Fatal(err)
}
```

### 3. Query Usage Data

```go
import "time"

// Daily usage for last 7 days
since := time.Now().AddDate(0, 0, -7)
daily, err := database.GetDailyUsage(since, time.Now(), "")

// Session data
sessions, err := database.GetSessionData("myproject", 10)

// Monthly usage
monthly, err := database.GetMonthlyUsage(2025, 10, "")

// Current active billing block
block, err := database.GetActiveBlock()

// Total cost for time range
totalCost, err := database.GetTotalCost(since, time.Now(), "")
```

## Environment Variables

- `CLAUDE_CONFIG_DIR` - Custom Claude data directory paths
  - Single path: `export CLAUDE_CONFIG_DIR=/path/to/claude`
  - Multiple paths: `export CLAUDE_CONFIG_DIR=/path1,/path2`
  - If not set, defaults to `~/.config/claude/projects` and `~/.claude/projects`

## File Discovery

The parser automatically discovers JSONL files from Claude data directories:

```
~/.config/claude/projects/
  ├── project1/
  │   ├── sessionId1.jsonl
  │   └── sessionId2.jsonl
  └── project2/
      └── sessionId3.jsonl
```

## Testing

Run tests to verify installation:

```bash
# All ccusage tests
go test ./internal/ccusage/... -v

# Only parser tests
go test ./internal/ccusage/parser -v

# Only database tests
go test ./internal/ccusage/db -v
```

## Common Patterns

### Daily Report with Model Breakdown

```go
since := time.Now().AddDate(0, 0, -7)
daily, _ := database.GetDailyUsage(since, time.Now(), "")

for _, d := range daily {
    fmt.Printf("%s: %d input, %d output, $%.2f\n",
        d.Date,
        d.InputTokens,
        d.OutputTokens,
        d.TotalCost,
    )
}
```

### Project-Specific Query

```go
// Get sessions for specific project
sessions, _ := database.GetSessionData("my-project", 0)

// Get daily usage for specific project
daily, _ := database.GetDailyUsage(since, until, "my-project")

// Get monthly usage for specific project
monthly, _ := database.GetMonthlyUsage(2025, 10, "my-project")
```

### Active Block with Projections

```go
block, _ := database.GetActiveBlock()

fmt.Printf("Active block: %s to %s\n",
    block.StartTime.Format(time.RFC3339),
    block.EndTime.Format(time.RFC3339),
)

fmt.Printf("Current usage: %d input, %d output, $%.2f\n",
    block.InputTokens,
    block.OutputTokens,
    block.TotalCost,
)

for project, data := range block.ProjectBreakdown {
    fmt.Printf("  %s: $%.2f\n", project, data.TotalCost)
}
```

## Performance Tips

1. **Use Incremental Sync**: Call `Sync()` instead of `FullSync()` for better performance
2. **Batch Queries**: Query multiple days in one call instead of individual queries
3. **Project Filtering**: Use project parameter to narrow query scope
4. **Connection Reuse**: Keep database connection open for multiple queries

## Troubleshooting

### No Files Found

Ensure JSONL files exist in Claude data directories:
- Check `~/.config/claude/projects/`
- Check `~/.claude/projects/`
- Or set `CLAUDE_CONFIG_DIR` environment variable

### Duplicate Entry Error

This shouldn't happen in normal operation. If it occurs:
1. The entry is already in database (checked via unique_hash)
2. This is intentional - prevents duplicate charges
3. Safe to retry sync operation

### Slow Queries

1. Check database is using indexes (verify with `EXPLAIN QUERY PLAN`)
2. Run incremental sync instead of full sync
3. Narrow time range in queries
4. Use project filtering where possible

## API Reference

See `/internal/ccusage/README.md` for detailed API documentation.
