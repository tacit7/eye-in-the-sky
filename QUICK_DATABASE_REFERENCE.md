# Quick Database Reference

## TL;DR - Database Population

### Automatic (Recommended)
```go
import (
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
)

// 1. Initialize database
ccdb, _ := db.New(dbPath)
defer ccdb.Close()

// 2. Sync (auto-discovers JSONL files and populates)
syncMgr := parser.NewSyncManager(ccdb)
syncMgr.Sync()  // Done!
```

## Database Schema

### Table: `usage_entries`
```
id                      INTEGER PRIMARY KEY
session_id              TEXT NOT NULL
timestamp               TEXT NOT NULL          (ISO8601)
project                 TEXT NOT NULL          (from file path)
model                   TEXT NOT NULL          (e.g., "claude-3-sonnet...")
input_tokens            INTEGER NOT NULL
output_tokens           INTEGER NOT NULL
cache_creation_tokens   INTEGER NOT NULL
cache_read_tokens       INTEGER NOT NULL
total_cost              REAL NOT NULL          (USD)
message_id              TEXT NOT NULL
request_id              TEXT NOT NULL
unique_hash             TEXT UNIQUE NOT NULL   (SHA256 for dedup)
created_at              DATETIME               (auto-set)
```

### Table: `file_metadata`
```
file_path               TEXT PRIMARY KEY       (absolute path to JSONL)
last_mtime              INTEGER NOT NULL       (Unix timestamp)
last_parsed_at          TEXT NOT NULL          (ISO8601)
```

## Where Data Comes From

```
JSONL Files (Claude Code data)
    ↓
~/.config/claude/projects/{project}/{sessionId}.jsonl
~/.claude/projects/{project}/{sessionId}.jsonl
    ↓
Auto-discovered by parser
    ↓
Parsed in parallel (4 workers)
    ↓
Inserted into usage_entries table
    ↓
Dedup checked (unique_hash)
    ↓
Ready to query!
```

## File Format

Each line in JSONL file is JSON:
```json
{
  "sessionId": "session-abc123",
  "timestamp": "2025-10-20T14:30:00Z",
  "message": {
    "usage": {
      "input_tokens": 100,
      "output_tokens": 200,
      "cache_creation_input_tokens": 0,
      "cache_read_input_tokens": 0
    },
    "model": "claude-3-sonnet-20240229",
    "id": "msg-123"
  },
  "costUSD": 0.05,
  "requestId": "req-456"
}
```

## Quick Operations

### View recent entries
```bash
sqlite3 ~/.config/eye-in-the-sky/ccusage.sqlite \
  "SELECT timestamp, project, model, total_cost FROM usage_entries LIMIT 10;"
```

### Count entries
```bash
sqlite3 ~/.config/eye-in-the-sky/ccusage.sqlite \
  "SELECT COUNT(*) FROM usage_entries;"
```

### Daily totals
```bash
sqlite3 ~/.config/eye-in-the-sky/ccusage.sqlite \
  "SELECT DATE(timestamp), SUM(total_cost) FROM usage_entries GROUP BY DATE(timestamp);"
```

## Sync Strategies

### Initial Full Sync
```go
syncMgr.FullSync()  // Parse everything
```

### Incremental Sync
```go
syncMgr.Sync()      // Only changed files (default, faster)
```

### Periodic Background Sync
```go
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
	syncMgr.Sync()  // Runs every 30 seconds
}
```

## Deduplication

Automatic via `UNIQUE constraint` on `unique_hash`:
- Hash = SHA256(messageId + requestId)
- Duplicate inserts silently rejected
- Safe to re-run sync

## Common Issues

| Problem | Solution |
|---------|----------|
| No data in DB | Check `~/.config/claude/projects/` exists and has JSONL files |
| DB corrupted | `rm ~/.config/eye-in-the-sky/ccusage.sqlite` then restart |
| Slow performance | Run `sqlite3 ... "VACUUM;"` to optimize |
| Only old data | Run `FullSync()` to force reparse |

## In Code

```go
// Initialize
ccdb, err := db.New(filepath.Join(home, ".config/eye-in-the-sky/ccusage.sqlite"))

// Sync on startup
syncMgr := parser.NewSyncManager(ccdb)
syncMgr.Sync()

// Query
daily, _ := api.GetDailyUsageReport(ccdb, 7)
monthly, _ := api.GetMonthlyUsageReport(ccdb, 2025, 10)

// Use in TUI
for _, report := range daily {
    fmt.Printf("%s: $%.2f\n", report.Date, report.TotalCost)
}
```

## Database Path

```
~/.config/eye-in-the-sky/ccusage.sqlite
```

## That's It!

The system:
- ✅ Finds JSONL files automatically
- ✅ Creates database if needed
- ✅ Creates schema if needed
- ✅ Parses in parallel
- ✅ Deduplicates automatically
- ✅ Handles errors gracefully

No manual database setup needed!
