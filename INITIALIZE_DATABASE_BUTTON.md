# Initialize Database Button - Implementation Guide

## Problem
If there's no data in the CCUsage database, the Usage tab shows "No data available". We need a way for users to initialize the database by discovering and parsing JSONL files.

## Solution
Add an interactive button/command in the Usage tab that triggers database initialization.

---

## Step 1: Add State to Model

**File**: `internal/ui/app/model.go`

Add these fields to the `Model` struct:

```go
type Model struct {
	// ... existing fields ...

	// CCUsage database connection
	ccusageDB *db.CCUsageDB

	// Cached ccusage data
	ccusageDaily    []api.DailyReport
	ccusageSessions []api.SessionReport
	ccusageMonthly  *api.MonthlyReport
	ccusageBlock    *api.ActiveBlockReport
	ccusageCosts    *api.CostSummary

	// CCUsage sync state
	lastCCUsageSync   time.Time
	ccusageSyncing    bool           // NEW: Track if sync is running
	ccusageSyncStatus string         // NEW: Status message
	ccusageEntryCount int            // NEW: Number of entries in DB
}
```

---

## Step 2: Add Count Function to DB

**File**: `internal/ccusage/db/queries.go`

Add this function to count entries:

```go
// GetEntryCount returns the number of entries in the database
func (c *CCUsageDB) GetEntryCount() (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM usage_entries").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count entries: %w", err)
	}

	return count, nil
}
```

---

## Step 3: Update Model Loading

**File**: `internal/ui/app/model.go`

Update the `loadCCUsageData()` method to track entry count:

```go
// loadCCUsageData loads Claude Code usage data from ccusage database
func (m *Model) loadCCUsageData() error {
	if m.ccusageDB == nil {
		return nil
	}

	// Get entry count
	count, err := m.ccusageDB.GetEntryCount()
	if err != nil {
		m.ccusageEntryCount = 0
	} else {
		m.ccusageEntryCount = count
	}

	// If no entries, don't bother querying
	if m.ccusageEntryCount == 0 {
		m.ccusageDaily = []api.DailyReport{}
		m.ccusageSessions = []api.SessionReport{}
		m.ccusageMonthly = nil
		m.ccusageBlock = nil
		m.ccusageCosts = nil
		return nil
	}

	// Load daily usage (last 7 days)
	daily, err := api.GetDailyUsageReport(m.ccusageDB, 7)
	if err != nil {
		return fmt.Errorf("failed to load daily usage: %w", err)
	}
	m.ccusageDaily = daily

	// ... rest of loading ...

	m.lastCCUsageSync = time.Now()
	return nil
}
```

---

## Step 4: Add Initialize Command

**File**: `internal/ui/app/update.go`

Add a new message type and handler:

```go
// In update.go, add this message type:

// initCCUsageMsg triggers CCUsage database initialization
type initCCUsageMsg struct{}

// In the Update() method, add a case for key press (like 'I' for Initialize):

case tea.KeyMsg:
	switch msg.String() {
	// ... existing key handlers ...

	case "i", "I":
		// Only in Usage tab
		if m.currentView == ViewList && m.listTabs.ActiveIndex == 3 {
			if m.ccusageDB != nil && m.ccusageEntryCount == 0 {
				m.ccusageSyncing = true
				m.ccusageSyncStatus = "Initializing database..."
				return m, tea.Batch(
					m.initCCUsageCmd(),
					m.tickCmd(), // Keep UI responsive
				)
			}
		}
	}
```

Add the command function:

```go
// initCCUsageCmd runs database initialization
func (m *Model) initCCUsageCmd() tea.Cmd {
	return func() tea.Msg {
		if m.ccusageDB == nil {
			return initCCUsageMsg{}
		}

		syncMgr := parser.NewSyncManager(m.ccusageDB)
		if err := syncMgr.Sync(); err != nil {
			m.ccusageSyncStatus = fmt.Sprintf("Error: %v", err)
			m.ccusageSyncing = false
			return initCCUsageMsg{}
		}

		// Reload data
		if err := m.loadCCUsageData(); err != nil {
			m.ccusageSyncStatus = fmt.Sprintf("Error loading data: %v", err)
		} else {
			m.ccusageSyncStatus = fmt.Sprintf("Initialized! Found %d entries", m.ccusageEntryCount)
		}

		m.ccusageSyncing = false
		return initCCUsageMsg{}
	}
}
```

Handle the message:

```go
// In Update() method, add case for initCCUsageMsg:

case initCCUsageMsg:
	// Sync is complete, data reloaded
	// UI will refresh automatically
	return m, m.tickCmd()
```

---

## Step 5: Update Usage Tab Rendering

**File**: `internal/ui/app/view.go`

Update `renderUsageTab()` to show initialization UI when empty:

```go
// renderUsageTab renders session costs for all agents with monthly breakdown
func (m *Model) renderUsageTab() string {
	var b strings.Builder

	// Check if CCUsage database is empty
	if m.ccusageDB != nil && m.ccusageEntryCount == 0 {
		b.WriteString(m.styles.Title.Render("Claude Code Usage"))
		b.WriteString("\n\n")

		if m.ccusageSyncing {
			b.WriteString(m.styles.Working.Render("  ⏳ " + m.ccusageSyncStatus))
		} else {
			b.WriteString(m.styles.Subtle.Render("  No Claude Code usage data found"))
			b.WriteString("\n\n")
			b.WriteString(m.styles.Primary.Render("  Press 'I' to initialize database"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  This will discover and parse JSONL files from:"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  ~/.config/claude/projects/"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  ~/.claude/projects/"))
		}

		b.WriteString("\n\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n\n")
	}

	// ... rest of existing Usage tab rendering ...

	// Only show Eye-in-the-Sky and Claude Code data if we have data
	if len(m.allSessionMetrics) > 0 || m.ccusageEntryCount > 0 {
		// Existing Eye-in-the-Sky section
		if len(m.allSessionMetrics) > 0 {
			b.WriteString(m.styles.Title.Render("Eye-in-the-Sky Session Metrics"))
			// ... existing rendering code ...
		}

		// New Claude Code section
		if m.ccusageDB != nil && m.ccusageEntryCount > 0 {
			b.WriteString(m.renderClaudeCodeMetrics())
			b.WriteString("\n\n")
			b.WriteString(m.renderCombinedSummary())
		}
	}

	return b.String()
}
```

---

## Step 6: Update Help Text

**File**: `internal/ui/app/keymap.go`

Add the new keybinding to help text:

```go
// In your keybindings, add:

"i": "Initialize CCUsage database",
```

---

## Step 7: Add to Model Initialization

**File**: `internal/ui/app/model.go`

In `NewModel()`, initialize the count:

```go
m := &Model{
	// ... existing fields ...
	ccusageDB: ccusageDB,
	ccusageSyncing: false,
	ccusageSyncStatus: "",
	ccusageEntryCount: 0,
}

// Check initial entry count
if ccusageDB != nil {
	if count, err := ccusageDB.GetEntryCount(); err == nil {
		m.ccusageEntryCount = count
	}
}
```

---

## User Experience Flow

### When Database is Empty:

```
┌──────────────────────────────────────────────────────┐
│ Claude Code Usage                                    │
│                                                      │
│ No Claude Code usage data found                      │
│                                                      │
│ Press 'I' to initialize database                    │
│                                                      │
│ This will discover and parse JSONL files from:      │
│ ~/.config/claude/projects/                          │
│ ~/.claude/projects/                                 │
│                                                      │
├──────────────────────────────────────────────────────┤
│ Eye-in-the-Sky Session Metrics                      │
│ [existing data]                                      │
└──────────────────────────────────────────────────────┘
```

### When User Presses 'I':

```
┌──────────────────────────────────────────────────────┐
│ Claude Code Usage                                    │
│                                                      │
│ ⏳ Initializing database...                         │
│                                                      │
├──────────────────────────────────────────────────────┤
│ Eye-in-the-Sky Session Metrics                      │
│ [existing data]                                      │
└──────────────────────────────────────────────────────┘
```

### After Initialization Completes:

```
┌──────────────────────────────────────────────────────┐
│ Claude Code Usage                                    │
│                                                      │
│ ✓ Initialized! Found 1250 entries                   │
│                                                      │
├──────────────────────────────────────────────────────┤
│ Claude Code Daily Usage (Last 7 Days)               │
│ Date         Project        Input    Output   Cost  │
│ 2025-10-20   myproject      1,500    2,300   $0.25  │
│ 2025-10-19   myproject      2,100    1,800   $0.22  │
│                                                      │
│ Claude Code Monthly Summary                         │
│ Month: 2025-10 | Days: 20 | Total: $4.32           │
│                                                      │
│ Combined Cost Summary                               │
│ Eye-in-the-Sky:  $12.45                            │
│ Claude Code:     $4.32                             │
│ ─────────────────────                              │
│ Total:          $16.77                             │
└──────────────────────────────────────────────────────┘
```

---

## Error Handling

If initialization fails:

```
⚠️ Error: Failed to discover files
```

The system will:
1. Log the error
2. Display message to user
3. Allow user to retry by pressing 'I' again
4. Keep app running normally

---

## Import Statements Needed

**File**: `internal/ui/app/update.go`

Add these imports:

```go
import (
	// ... existing imports ...
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
)
```

---

## Summary

The initialize button provides:

✅ One-click database initialization
✅ User feedback during processing
✅ Entry count tracking
✅ Graceful error handling
✅ Non-blocking UI (uses Bubble Tea commands)
✅ Clear instructions for users

Just press 'I' to initialize when database is empty!
