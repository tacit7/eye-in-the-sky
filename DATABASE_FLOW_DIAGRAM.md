# Database Population Flow Diagram

## Complete Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    CLAUDE CODE JSONL FILES                      │
│                                                                  │
│  ~/.config/claude/projects/myproject/session-abc.jsonl         │
│  ~/.claude/projects/myproject/session-def.jsonl                │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│              FILE DISCOVERY (discovery.go)                      │
│                                                                  │
│  • Check CLAUDE_CONFIG_DIR env var                             │
│  • Fall back to ~/.config/claude/projects/                     │
│  • Fall back to ~/.claude/projects/                            │
│  • Walk directories for *.jsonl files                          │
│  • Extract project name from path                              │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│           SYNC MANAGER (sync.go) - Two Modes                    │
│                                                                  │
│  1. Full Sync: Parse ALL files                                 │
│  2. Incremental: Check mtime, only parse changed files         │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│        PARALLEL PARSER (parser.go) - Worker Pool                │
│                                                                  │
│     File Channel    →    Worker 1 ──→ ┐                        │
│                     →    Worker 2 ──→ │                        │
│                     →    Worker 3 ──→ ├→  Result Channel       │
│                     →    Worker 4 ──→ │                        │
│                                        ┴→                        │
│  Each Worker:                                                   │
│  • Open JSONL file                                             │
│  • Parse JSON lines                                            │
│  • Validate required fields                                    │
│  • Calculate hash (SHA256)                                     │
│  • Skip malformed entries                                      │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│          DEDUPLICATION CHECK (in-memory map)                    │
│                                                                  │
│  • For each entry: Create hash = SHA256(messageId + requestId) │
│  • Check if hash already seen in current parse                 │
│  • Skip if duplicate                                           │
│  • Add to dedup map if new                                     │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│           DATABASE INSERTION (ccusage.go)                       │
│                                                                  │
│  Insert into usage_entries (transaction mode)                  │
│  • Checks UNIQUE constraint on unique_hash                    │
│  • Silently skips duplicates from previous syncs               │
│  • Updates file_metadata with new mtime                        │
│  • Auto-set created_at timestamp                               │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│              SQLITE DATABASE (WAL Mode)                         │
│                                                                  │
│  ┌─────────────────────────────────────────────────────┐       │
│  │ usage_entries                                       │       │
│  ├─────────────────────────────────────────────────────┤       │
│  │ id | session_id | timestamp | project | model ...  │       │
│  │  1 | sess-123   | 2025-10...|  proj1  | claude-3.. │       │
│  │  2 | sess-456   | 2025-10...|  proj1  | claude-3.. │       │
│  │  3 | sess-789   | 2025-10...|  proj2  | claude-4.. │       │
│  │ ... (indexed on timestamp, project, model, session) │       │
│  └─────────────────────────────────────────────────────┘       │
│                                                                  │
│  ┌─────────────────────────────────────────────────────┐       │
│  │ file_metadata                                       │       │
│  ├─────────────────────────────────────────────────────┤       │
│  │ file_path | last_mtime | last_parsed_at           │       │
│  │ /home/.../session-abc.jsonl | 1729444400 | 2025... │       │
│  │ /home/.../session-def.jsonl | 1729444500 | 2025... │       │
│  └─────────────────────────────────────────────────────┘       │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│               QUERY API (api.go) - High-Level                   │
│                                                                  │
│  • GetDailyUsageReport(days)    → []DailyReport               │
│  • GetSessionUsageReport(limit) → []SessionReport             │
│  • GetMonthlyUsageReport(y, m)  → *MonthlyReport              │
│  • GetTotalCostSummary(days)    → *CostSummary               │
│  • GetActiveBlockReport()       → *ActiveBlockReport          │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
┌─────────────────────────────────────────────────────────────────┐
│                    TUI DISPLAY (view.go)                        │
│                                                                  │
│  ┌──────────────────────────────────────────────────────┐      │
│  │ Claude Code Daily Usage (Last 7 Days)              │      │
│  ├──────────────────────────────────────────────────────┤      │
│  │ Date         Project    Input   Output   Cost      │      │
│  │ 2025-10-20   myproject  1,500   2,300    $0.25     │      │
│  │ 2025-10-19   myproject  2,100   1,800    $0.22     │      │
│  └──────────────────────────────────────────────────────┘      │
│                                                                  │
│  ┌──────────────────────────────────────────────────────┐      │
│  │ Claude Code Sessions                                │      │
│  ├──────────────────────────────────────────────────────┤      │
│  │ SessionId    Project    Duration    Cost           │      │
│  │ session-abc  myproject  1h 30m      $0.25          │      │
│  └──────────────────────────────────────────────────────┘      │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ↓
                    ✅ USER SEES DATA IN TUI
```

## Sequence Diagram - Startup

```
Application Start
      │
      ├─→ Initialize Eye-in-the-Sky DB
      │       │
      │       └─→ Load agents, sessions, etc.
      │
      ├─→ Initialize CCUsage DB
      │       │
      │       └─→ Create database file
      │
      ├─→ Check if database exists
      │       │
      │       ├─ YES: Skip schema creation
      │       │
      │       └─ NO: Create tables and indexes
      │
      ├─→ Create Sync Manager
      │
      ├─→ Run Initial Sync
      │       │
      │       ├─→ Discover files
      │       │       │
      │       │       └─→ Find *.jsonl files
      │       │
      │       ├─→ Check file mtimes
      │       │       │
      │       │       └─→ Load from file_metadata
      │       │
      │       ├─→ Filter modified files
      │       │       │
      │       │       └─→ Only parse if mtime changed
      │       │
      │       ├─→ Parse in parallel
      │       │       │
      │       │       ├─→ Worker 1: Parse chunk
      │       │       ├─→ Worker 2: Parse chunk
      │       │       ├─→ Worker 3: Parse chunk
      │       │       └─→ Worker 4: Parse chunk
      │       │
      │       ├─→ Batch insert to database
      │       │
      │       └─→ Update file_metadata
      │
      ├─→ Display TUI
      │       │
      │       └─→ Query API for data
      │
      └─→ Start monitoring loop
              │
              └─→ Every 30 seconds: Run incremental sync
```

## Data Entry Lifecycle

```
JSONL File Created
      │
      ├─→ Claude Code writes JSON line to file
      │       │
      │       └─→ JSONL line: {"sessionId": "...", ...}
      │
User Starts Eye-in-the-Sky TUI
      │
      ├─→ CCUsage subsystem initializes
      │
      ├─→ File discovery finds JSONL file
      │       │
      │       └─→ Path: ~/.config/claude/projects/myproject/session-abc.jsonl
      │
      ├─→ Check file_metadata table
      │       │
      │       └─→ Is this file new or modified?
      │
      ├─→ YES: Add to parse queue
      │
      ├─→ Parser processes file
      │       │
      │       ├─→ Read file content
      │       │
      │       ├─→ Parse JSON lines
      │       │
      │       ├─→ For each line:
      │       │   ├─→ Validate required fields
      │       │   ├─→ Create unique_hash
      │       │   ├─→ Check dedup map
      │       │   └─→ Create UsageEntryRow
      │       │
      │       └─→ Collect all valid entries
      │
      ├─→ Batch insert to database (transaction)
      │       │
      │       ├─→ BEGIN TRANSACTION
      │       │
      │       ├─→ INSERT multiple rows
      │       │   └─→ If unique_hash conflict: skip (dupe)
      │       │
      │       └─→ COMMIT TRANSACTION
      │
      ├─→ Update file_metadata
      │       │
      │       └─→ Store new mtime and parse timestamp
      │
      ├─→ Data now queryable
      │       │
      │       └─→ GetDailyUsageReport() returns data
      │
      └─→ TUI renders Usage tab with data
```

## Database States

```
STATE 1: Fresh Installation
┌──────────────────────────────┐
│ No database file exists      │
│ No JSONL files processed     │
└──────────────┬───────────────┘
              │ User starts app
              ↓
┌──────────────────────────────┐
│ Database created             │
│ Schema initialized           │
│ usage_entries: EMPTY         │
│ file_metadata: EMPTY         │
└──────────────┬───────────────┘
              │ Sync runs
              ↓


STATE 2: After First Sync
┌──────────────────────────────┐
│ Database exists              │
│ Schema complete              │
│ usage_entries: 1000+ rows    │
│ file_metadata: 5 entries     │
└──────────────┬───────────────┘
              │ App running
              ↓


STATE 3: Incremental Update (file changed)
┌──────────────────────────────┐
│ Database exists              │
│ file_metadata shows old mtime│
│ New JSONL entries written    │
└──────────────┬───────────────┘
              │ Next sync (30s)
              ↓
┌──────────────────────────────┐
│ Check file_metadata          │
│ Detect mtime changed         │
│ Parse only modified file     │
│ Merge new entries            │
│ Update file_metadata         │
└──────────────┬───────────────┘
              │ Query returns updated data
              ↓
        ✅ TUI shows new entries
```

## Error Recovery

```
Error Occurs
      │
      ├─ File read error
      │   └─→ Skip file, continue with others
      │
      ├─ JSON parse error
      │   └─→ Skip line, continue parsing
      │
      ├─ Database insert error
      │   └─→ Rollback transaction, log error
      │
      ├─ Sync fails completely
      │   └─→ Log error, use cached data
      │
      ├─ Network/permissions
      │   └─→ Report warning, continue
      │
      └─ Data validation fails
          └─→ Skip entry, continue

Result: Application continues working
        Uses cached/partial data
        Logs errors for debugging
```

## Performance Profile

```
OPERATION                           TIME        NOTES
─────────────────────────────────────────────────────────────
Initialize database                ~10ms       First time only
Create schema                       ~5ms        If needed
Discover files (100 files)          ~50ms       File I/O
Full sync (10k entries)             ~400ms      4 workers
  ├─ Parse time                     ~350ms      Parallel
  └─ Insert time                    ~50ms       Transaction
Incremental sync (1 file changed)   ~50ms       Only changed file
Query daily usage                   <10ms       With index
Query monthly usage                 <20ms       With index
Query active block                  <5ms        Simple calc
Refresh every 30 seconds            ~50ms       Only changed files
─────────────────────────────────────────────────────────────

Total for initial startup: ~500ms (including parse)
Ongoing overhead: ~50ms every 30 seconds
Memory usage: <10MB typical
```

## Deduplication Mechanism

```
Entry from JSONL:
  messageId: "msg-123"
  requestId: "req-456"
         │
         ├─→ Hash = SHA256("msg-123req-456")
         │
         ├─→ First parse:
         │   └─→ unique_hash NOT in database
         │   └─→ INSERT successful
         │
         ├─→ Second parse (rerun sync):
         │   └─→ unique_hash already exists
         │   └─→ UNIQUE constraint violation
         │   └─→ INSERT rejected (no error)
         │
         └─→ Data never duplicated
```
