# CCUsage Integration - Complete Implementation Checklist

## Phase 1: Database Setup ✅
- [x] Schema created (2 tables)
- [x] Parser built (parallel JSONL processing)
- [x] API wrapper created (5 high-level functions)
- [x] Tests passing (20+ tests)

---

## Phase 2: Initialize Database Button ⏳

### Model Updates (`internal/ui/app/model.go`)
- [ ] Add ccusageDB field
- [ ] Add ccusageDaily field
- [ ] Add ccusageSessions field
- [ ] Add ccusageMonthly field
- [ ] Add ccusageBlock field
- [ ] Add ccusageCosts field
- [ ] Add ccusageSyncing field (bool)
- [ ] Add ccusageSyncStatus field (string)
- [ ] Add ccusageEntryCount field (int)
- [ ] Add lastCCUsageSync field (time.Time)

### Database Query (`internal/ccusage/db/queries.go`)
- [ ] Add GetEntryCount() function

### Model Loading (`internal/ui/app/model.go`)
- [ ] Add loadCCUsageData() method
- [ ] Track entry count in NewModel()
- [ ] Load initial data if available

### Command Handler (`internal/ui/app/update.go`)
- [ ] Add initCCUsageMsg type
- [ ] Add case for 'I' key press
- [ ] Add initCCUsageCmd() function
- [ ] Add case for initCCUsageMsg in Update()
- [ ] Add import for parser package

### View Rendering (`internal/ui/app/view.go`)
- [ ] Update renderUsageTab() to show empty state
- [ ] Add "Press 'I' to initialize" message
- [ ] Show initialization status
- [ ] Add renderClaudeCodeMetrics() method
- [ ] Add renderCombinedSummary() method
- [ ] Call new methods after data loads

### Help Text (`internal/ui/app/keymap.go`)
- [ ] Add "i" keybinding for initialize

### Main Initialization (`cmd/server/main.go`)
- [ ] Import ccusage packages
- [ ] Initialize ccusageDB
- [ ] Create sync manager
- [ ] Run initial sync
- [ ] Pass ccusageDB to model
- [ ] Close ccusageDB on shutdown

---

## Phase 3: UI Enhancements (Optional)
- [ ] Add spinner during sync
- [ ] Add progress bar for parsing
- [ ] Add error notifications
- [ ] Add success message with entry count
- [ ] Add manual refresh option

---

## Testing Checklist

### Unit Tests
- [ ] Test GetEntryCount() with empty database
- [ ] Test GetEntryCount() with data
- [ ] Test initCCUsageCmd() with no files
- [ ] Test initCCUsageCmd() with files
- [ ] Test renderUsageTab() empty state
- [ ] Test renderUsageTab() with data

### Integration Tests
- [ ] Test full initialization flow
- [ ] Test with real JSONL files
- [ ] Test error scenarios
- [ ] Test concurrent updates
- [ ] Test UI responsiveness during sync

### Manual Testing
- [ ] [ ] Start Eye-in-the-Sky with empty database
- [ ] [ ] Verify "Press 'I'" message appears
- [ ] [ ] Press 'I' to initialize
- [ ] [ ] Verify sync progress shows
- [ ] [ ] Verify data appears after sync
- [ ] [ ] Verify entry count displays
- [ ] [ ] Test with multiple projects
- [ ] [ ] Test with large JSONL files

---

## Documentation Updates
- [ ] Update README.md with initialization instructions
- [ ] Add keyboard shortcut to help menu
- [ ] Document database location
- [ ] Add troubleshooting section

---

## Code Files to Modify

### Files to Edit:
1. `internal/ui/app/model.go` - Add fields, loading logic
2. `internal/ui/app/view.go` - Add rendering, UI
3. `internal/ui/app/update.go` - Add command handler
4. `internal/ui/app/keymap.go` - Add help text
5. `internal/ccusage/db/queries.go` - Add GetEntryCount()
6. `cmd/server/main.go` - Initialize database

### Total Lines of Code:
- Model updates: ~40 lines
- View updates: ~150 lines
- Command handler: ~50 lines
- Query function: ~15 lines
- Main initialization: ~30 lines
- **Total: ~285 lines**

---

## How It Works

### 1. User Starts App (First Time)
```
App starts
  ↓
Initialize CCUsage database
  ↓
Check if database has data
  ↓
NO: Show empty state + "Press 'I'"
YES: Load and display data
```

### 2. User Presses 'I'
```
User presses 'I'
  ↓
Show "Initializing..." status
  ↓
Run sync in background command
  ↓
Discover JSONL files
  ↓
Parse in parallel (4 workers)
  ↓
Insert into database
  ↓
Reload data and display
  ↓
Show "Initialized! Found X entries"
```

### 3. Subsequent Uses
```
App starts
  ↓
Check database entry count
  ↓
Load and display data
  ↓
Run incremental sync every 30 seconds
  ↓
Update display if new data found
```

---

## User Experience

### First Launch (No Data)
```
┌─────────────────────────────────────────┐
│ Usage Tab                               │
├─────────────────────────────────────────┤
│ Claude Code Usage                       │
│                                         │
│ No Claude Code usage data found         │
│                                         │
│ Press 'I' to initialize database       │
│                                         │
│ This will discover and parse JSONL...   │
│                                         │
└─────────────────────────────────────────┘
```

### During Initialization
```
┌─────────────────────────────────────────┐
│ Usage Tab                               │
├─────────────────────────────────────────┤
│ Claude Code Usage                       │
│                                         │
│ ⏳ Initializing database...             │
│    Scanning files...                    │
│    Parsing 1,250 entries...             │
│                                         │
└─────────────────────────────────────────┘
```

### After Initialization
```
┌─────────────────────────────────────────┐
│ Usage Tab                               │
├─────────────────────────────────────────┤
│ Claude Code Usage                       │
│                                         │
│ ✓ Initialized! Found 1,250 entries     │
│                                         │
│ Claude Code Daily Usage (Last 7 Days)  │
│ Date      Project    Input  Output Cost │
│ 2025-10   proj1      1,500  2,300  0.25│
│ 2025-10   proj1      2,100  1,800  0.22│
│                                         │
│ Claude Code Monthly Summary             │
│ Month: 2025-10 | Total: $4.32          │
│                                         │
│ Combined Cost Summary                   │
│ Eye-in-the-Sky:  $12.45                │
│ Claude Code:     $4.32                 │
│ Total:           $16.77                │
│                                         │
└─────────────────────────────────────────┘
```

---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `I` | Initialize CCUsage database (when empty) |
| `?` | Show help menu |
| `U` | Switch to Usage tab |
| `q` | Quit |

---

## Error Scenarios Handled

1. **No JSONL files found**
   - Message: "No files found in Claude directories"
   - User can check paths and try again

2. **Malformed JSONL**
   - Message: "Skipped N invalid entries"
   - Valid entries still inserted

3. **Database errors**
   - Message: "Database error: [error details]"
   - Suggests deleting database and retrying

4. **Permission denied**
   - Message: "Permission denied: Check ~/.config/eye-in-the-sky/"
   - User can fix permissions and retry

5. **Disk space**
   - Message: "Not enough disk space"
   - User needs to free space

---

## Testing Commands

```bash
# Check database has data
sqlite3 ~/.config/eye-in-the-sky/ccusage.sqlite "SELECT COUNT(*) FROM usage_entries;"

# Check file metadata
sqlite3 ~/.config/eye-in-the-sky/ccusage.sqlite "SELECT COUNT(*) FROM file_metadata;"

# View recent entries
sqlite3 ~/.config/eye-in-the-sky/ccusage.sqlite \
  "SELECT timestamp, project, model, total_cost FROM usage_entries ORDER BY timestamp DESC LIMIT 5;"
```

---

## Estimated Effort

| Task | Time |
|------|------|
| Model updates | 30 min |
| View rendering | 45 min |
| Command handler | 30 min |
| Testing | 1 hour |
| Documentation | 30 min |
| **Total** | **~3 hours** |

---

## Next Steps

1. Start with model.go updates (add fields)
2. Add database query function
3. Implement command handler in update.go
4. Update view rendering
5. Test with real data
6. Document and commit

Ready to implement? Follow the detailed guides in the documentation files!
