# CCUsage TUI Integration Plan

## Current State

The Eye-in-the-Sky TUI already has a Usage tab that displays:
- **Current Session Summary**: Per-agent token usage and costs from `session_metrics` table
- **Monthly Cost Breakdown**: Monthly costs from `session_metrics` aggregations

## Integration Strategy

We will add Claude Code usage data (from ccusage) alongside the existing Eye-in-the-Sky session metrics. The new sections will include:

1. **Daily Usage** - Daily breakdown of Claude tokens and costs
2. **Session Usage** - Claude Code session aggregation
3. **Monthly Usage** - Monthly summary of all Claude usage
4. **Total Cost** - Overall spending metrics

## Implementation Steps

### Step 1: Create Go Wrapper Functions
**File**: `internal/ccusage/api.go` (new)

Expose the ccusage database queries as simple, callable functions:
- `GetDailyUsageData(db *sql.DB, days int) ([]DailyReport, error)`
- `GetSessionUsageData(db *sql.DB, limit int) ([]SessionReport, error)`
- `GetMonthlyUsageData(db *sql.DB, year, month int) (*MonthlyReport, error)`
- `GetTotalCostData(db *sql.DB, days int) (*CostSummary, error)`

### Step 2: Integrate CCUsage Database with TUI
**Files**: `internal/ui/app/model.go`

Add to Model struct:
```go
ccusageDB *db.CCUsageDB  // CCUsage database connection
ccusageData CCUsageData   // Cache of ccusage query results
```

Initialize in `NewModel()`:
```go
ccusageDB, err := db.New(dbPath)
if err != nil {
    // Log warning but continue - ccusage is optional
}
```

### Step 3: Load CCUsage Data
**File**: `internal/ui/app/model.go`

Add method `loadCCUsageData()`:
- Query daily usage (last 7 days)
- Query session usage
- Query monthly usage
- Calculate total cost

Call during:
- Application startup
- Tab refresh (every refresh cycle)
- On demand when Usage tab becomes active

### Step 4: Enhance renderUsageTab()
**File**: `internal/ui/app/view.go`

Modify to show new sections:
1. Eye-in-the-Sky metrics (keep existing)
2. **Claude Code Daily Usage** - Last 7 days
3. **Claude Code Session Summary** - Top sessions
4. **Claude Code Monthly Summary** - Current month total
5. **Combined Cost Summary** - Both Eye-in-the-Sky + Claude Code

### Step 5: Handle Data Synchronization
**File**: `internal/ui/app/model.go`

Add auto-sync:
- On each tick, check if ccusage files have changed
- Run incremental sync if needed
- Update UI asynchronously

## Data Structures

### CCUsageData
```go
type CCUsageData struct {
    DailyUsage   []models.DailyUsage
    SessionUsage []models.SessionData
    MonthlyUsage *models.DailyUsage
    ActiveBlock  *models.BlockUsage
    LastSync     time.Time
    SyncError    error
}
```

### UI Display Format

```
═══════════════════════════════════════════════════════════════════════
Eye-in-the-Sky Session Metrics
─────────────────────────────────────────────────────────────────────
[Existing content]

═══════════════════════════════════════════════════════════════════════
Claude Code Daily Usage (Last 7 Days)
─────────────────────────────────────────────────────────────────────
Date         Project        Input    Output   Cost USD
2025-10-20   myproject      1,500    2,300    $0.25
2025-10-19   myproject      2,100    1,800    $0.22
...

═══════════════════════════════════════════════════════════════════════
Claude Code Session Summary
─────────────────────────────────────────────────────────────────────
SessionId    Project        Start              Duration    Cost USD
session-123  myproject      2025-10-20 14:30   45m         $0.18
session-122  myproject      2025-10-20 13:00   30m         $0.12
...

═══════════════════════════════════════════════════════════════════════
Claude Code Monthly Summary
─────────────────────────────────────────────────────────────────────
Month        Days    Total Tokens    Total Cost
2025-10      20      156,800         $4.32

═══════════════════════════════════════════════════════════════════════
Combined Cost Summary
─────────────────────────────────────────────────────────────────────
Eye-in-the-Sky Sessions:    $12.45
Claude Code Usage:          $4.32
                           ─────────
Total Monthly Cost:         $16.77
```

## Database Configuration

The ccusage database needs to be initialized on application startup:

```go
// In main.go or server initialization
dbPath := filepath.Join(home, ".config/eye-in-the-sky/ccusage.sqlite")
ccusageDB, err := db.New(dbPath)
if err != nil {
    log.Printf("Warning: CCUsage database unavailable: %v", err)
} else {
    // Perform initial sync
    syncMgr := parser.NewSyncManager(ccusageDB)
    if err := syncMgr.Sync(); err != nil {
        log.Printf("Warning: Initial CCUsage sync failed: %v", err)
    }
}
```

## Error Handling

- If CCUsage database is unavailable, Usage tab shows only Eye-in-the-Sky metrics
- If sync fails, continue with cached data
- Errors logged but don't crash application
- Display "Loading..." if data is being fetched asynchronously

## Performance Considerations

1. **Query Performance**: All queries targeted to return in <50ms
2. **Refresh Rate**: Load CCUsage data on each refresh cycle (configurable)
3. **Caching**: Keep data in memory, update periodically
4. **Background Sync**: Can optionally run sync in background goroutine
5. **Memory**: Typical memory footprint <10MB for daily data

## Future Enhancements

1. **Filtering**: Filter by project, date range, model
2. **Export**: Export usage data to CSV/JSON
3. **Charts**: ASCII charts for visual representation
4. **Alerts**: Alert if daily/monthly costs exceed threshold
5. **Comparison**: Compare costs across different time periods
6. **Detailed View**: Click through to see model breakdown

## Testing Strategy

1. **Unit Tests**: Test wrapper functions with sample data
2. **Integration Tests**: Test TUI rendering with both data sources
3. **Mock Data**: Create test fixtures with sample JSONL files
4. **Display Tests**: Verify proper formatting and wrapping

## Files to Modify

| File | Changes |
|------|---------|
| `internal/ccusage/api.go` | NEW - Wrapper functions |
| `internal/ui/app/model.go` | Add CCUsage DB + data loading |
| `internal/ui/app/view.go` | Enhance renderUsageTab() |
| `internal/ui/app/update.go` | Add CCUsage refresh logic |
| `cmd/server/main.go` | Initialize CCUsage database |

## Implementation Order

1. ✅ Create API wrapper functions (Phase 1)
2. Integrate CCUsage DB into TUI model (Phase 2)
3. Add data loading functions (Phase 3)
4. Enhance Usage tab rendering (Phase 4)
5. Add refresh/sync logic (Phase 5)
6. Test and optimize (Phase 6)
