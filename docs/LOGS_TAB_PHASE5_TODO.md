# Logs Tab Auto-Refresh - Phase 5 Implementation Notes

**Status**: Infrastructure complete, integration pending
**Date**: 2025-10-27
**Agent**: 5f8088fe-9d8a-469e-814b-ad3215519a81

---

## Completed (Phases 1-4)

✅ **Phase 1**: Centralized theme package
  - `internal/ui/theme/` already existed
  - Added `OverviewStyles` compatibility layer to `theme.go`
  - Provides bridge between old and new style systems

✅ **Phase 2**: Updated DataContext
  - Added `LogsListIndex int` for viewport scroll position
  - Added `LastFetchedAt time.Time` for incremental fetching
  - Location: `internal/ui/app/views/agent_details/tabs/overview_tab.go:19-31`

✅ **Phase 3**: Rewrote logs_tab.go
  - Color-coded log types (error/warning/info/commit)
  - Full-width messages (no 50-char truncation)
  - Responsive to terminal width via `ctx.Width`
  - Helper functions: `truncateString()`, `padRight()`
  - Location: `internal/ui/app/views/agent_details/tabs/logs_tab.go`

✅ **Phase 4**: Database query
  - Added `GetLogsAfter(sessionID, after time.Time)`
  - Returns logs created after timestamp (for incremental fetching)
  - Location: `internal/ui/app/database/queries.go:599-622`

---

## Pending: Phase 5 - Auto-Refresh Integration

### Current Architecture

**Tick System** (already exists):
```
internal/ui/app/model.go:596-600
- tickCmd() creates tick every m.config.RefreshDuration()
- Sends tickMsg on interval

internal/ui/app/update_root.go:158-176
- case tickMsg: refreshes agent list only
- Does NOT refresh detail data (logs, commits, etc.)
```

**Data Loading**:
```
internal/ui/app/model.go:697-719
- loadLogs() loads all logs for selectedAgent.SessionID
- Called once when agent is selected
- No incremental update mechanism
```

### Implementation Path

**Option A: Add to existing tick handler** (Recommended)
```go
// In internal/ui/app/update_root.go, case tickMsg:

case tickMsg:
    cmds := []tea.Cmd{
        loadAgentsCmd(m.data.Agents),
        m.tickCmd(),
    }

    // NEW: Refresh logs if on logs tab
    if m.currentView == ViewDetail && m.tabs.ActiveIndex == tabLogs {
        if m.selectedAgent != nil {
            // Option A.1: Full refresh (simple)
            if err := m.loadLogs(); err != nil {
                log.Printf("Warning: Failed to reload logs: %v\n", err)
            }

            // Option A.2: Incremental refresh (optimal)
            // if err := m.loadLogsIncremental(); err != nil {
            //     log.Printf("Warning: Failed to reload logs: %v\n", err)
            // }
        }
    }

    return m, tea.Batch(cmds...)
```

**Option B: Create dedicated logs refresh message**
```go
// In model.go
type refreshLogsMsg struct{}

// In update_root.go, case tickMsg:
if m.currentView == ViewDetail && m.tabs.ActiveIndex == tabLogs {
    cmds = append(cmds, func() tea.Msg { return refreshLogsMsg{} })
}

// New case in update_root.go:
case refreshLogsMsg:
    if err := m.loadLogsIncremental(); err != nil {
        log.Printf("Warning: Failed to reload logs: %v\n", err)
    }
    return m, nil
```

### New Method Needed: loadLogsIncremental()

```go
// In internal/ui/app/model.go (after loadLogs)

// loadLogsIncremental loads new logs since last fetch
func (m *Model) loadLogsIncremental() error {
    if m.selectedAgent == nil || m.selectedAgent.SessionID == "" {
        return nil
    }

    // Get last timestamp from existing logs
    var after time.Time
    if len(m.logs) > 0 {
        after = m.logs[len(m.logs)-1].Timestamp
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Use new GetLogsAfter query
    newLogs, err := m.data.DB.GetLogsAfter(m.selectedAgent.SessionID, after)
    if err != nil {
        return err
    }

    // Append new logs
    if len(newLogs) > 0 {
        m.logs = append(m.logs, newLogs...)

        // Auto-scroll if near bottom (viewport behavior)
        // This requires knowing viewport state - may need refactoring
    }

    return nil
}
```

### DataClient Interface Update

Need to add `GetLogsAfter` to DataClient interface:

```go
// In internal/ui/app/data_client.go

type LogsRepository interface {
    LoadBySession(ctx context.Context, sessionID string, limit int) ([]Log, error)
    GetLogsAfter(sessionID string, after time.Time) ([]*database.Log, error) // NEW
}
```

### Challenges

1. **Type Mismatch**: `m.logs` is `[]Log` (local type) but `GetLogsAfter` returns `[]*database.Log`
   - Need type conversion helper
   - Or change m.logs to use database.Log

2. **Viewport State**: Auto-scroll "if near bottom" requires viewport awareness
   - Current architecture: tabs are stateless renderers
   - May need to track scroll position in Model

3. **Performance**: Refreshing every tick might be too frequent
   - Consider separate refresh interval for logs (e.g., every 5 seconds vs general 2 seconds)
   - Could use conditional: `if time.Since(m.lastLogsRefresh) > 5*time.Second`

---

## Testing Checklist (Post-Implementation)

- [ ] Logs auto-refresh when on logs tab
- [ ] No refresh when on other tabs
- [ ] New logs append to existing logs
- [ ] Scroll position maintained when new logs arrive
- [ ] Auto-scroll if at bottom when new logs arrive
- [ ] Performance acceptable (no UI lag)
- [ ] Works with rapid log generation
- [ ] Handles empty log updates gracefully

---

## Alternative: Full Viewport-Based Approach

If you want full tail -f behavior per LOG_VIEWER_ACTION_PLAN.md:

1. Convert logs tab to use `bubbles/viewport` component
2. Store viewport model in parent Model
3. Implement append-based content updates
4. Add scroll position tracking

This is more invasive but provides better UX. Current approach (Option A above) is simpler and works with existing architecture.

---

## Next Steps

1. Decide between Option A (tick handler) vs Option B (dedicated message)
2. Implement `loadLogsIncremental()` method
3. Update DataClient interface
4. Handle type conversions
5. Test with active logging sessions
6. Consider viewport refactor if better UX needed
