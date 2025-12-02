# CCUsage Parser Package

This package provides high-performance JSONL parsing for Claude Code usage data with SQLite persistence and auto-sync capabilities.

## Architecture

### Components

1. **models** - Data structures for usage entries, daily/monthly reports, and session data
2. **parser** - Parallel JSONL file parser with worker pool and deduplication
3. **db** - SQLite database operations with thread-safe concurrent access

### Key Features

- **Parallel Parsing**: Worker pool processes multiple JSONL files concurrently
- **Auto-Sync**: Tracks file modification times, only re-parses changed files
- **Deduplication**: SHA256 hashing prevents duplicate entries across re-parses
- **Thread-Safe**: Mutex-protected database operations with WAL mode for concurrent access
- **Performance**: ~5-10x faster than TypeScript re-parsing approach

## Usage Example

```go
package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
)

func main() {
	// Initialize database
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".config/eye-in-the-sky/ccusage.sqlite")

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Create sync manager and perform initial sync
	syncMgr := parser.NewSyncManager(database)
	if err := syncMgr.Sync(); err != nil {
		log.Fatal(err)
	}

	// Query daily usage
	since := time.Now().AddDate(0, 0, -7) // Last 7 days
	daily, err := database.GetDailyUsage(since, time.Now(), "")
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range daily {
		println(d.Date, d.Project, d.TotalCost)
	}

	// Query current active block
	activeBlock, err := database.GetActiveBlock()
	if err != nil {
		log.Fatal(err)
	}

	println("Active block cost:", activeBlock.TotalCost)
}
```

## Database Schema

### usage_entries
- `id` - Primary key
- `session_id` - Claude Code session ID
- `timestamp` - ISO8601 timestamp
- `project` - Project name
- `model` - Claude model used
- `input_tokens` - Input token count
- `output_tokens` - Output token count
- `cache_creation_tokens` - Prompt cache creation tokens
- `cache_read_tokens` - Prompt cache read tokens
- `total_cost` - USD cost
- `message_id` - Message ID
- `request_id` - Request ID
- `unique_hash` - SHA256(messageId + requestId) for deduplication

### file_metadata
- `file_path` - Absolute path to JSONL file
- `last_mtime` - Last modification time
- `last_parsed_at` - Last parse timestamp

## File Discovery

Automatically discovers JSONL files from:
1. `$CLAUDE_CONFIG_DIR` environment variable (if set, supports comma-separated paths)
2. `~/.config/claude/projects/`
3. `~/.claude/projects/`

File structure expected: `{base}/{project}/{sessionId}.jsonl`

## Performance Targets

- Parse 10k+ JSONL entries: <500ms with 4 workers
- Query for daily/monthly reports: <50ms
- Active block query: <10ms

## Thread Safety

- Database uses `sync.RWMutex` for thread-safe access
- SQLite configured with WAL mode for concurrent reads
- Connection pool: max 25 open, 5 idle connections
- Deduplication map protected during parsing

## Implementation Notes

- Malformed JSON lines are silently skipped (matches TypeScript behavior)
- File discovery continues even if some directories don't exist
- Errors during individual file parsing don't stop overall sync
- Cost calculations use pre-calculated `costUSD` from entries
