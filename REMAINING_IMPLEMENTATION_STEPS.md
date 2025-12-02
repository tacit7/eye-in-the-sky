# Remaining Implementation Steps

## ✅ COMPLETED (Commit: caa6654)

1. ✅ Added ccusageDB fields to Model struct
2. ✅ Added imports for ccusage packages
3. ✅ Updated NewModel() to accept ccusageDB parameter
4. ✅ Added loadCCUsageData() method
5. ✅ Added GetEntryCount() database function

---

## ⏳ REMAINING STEPS (Next to Implement)

### Step 1: Add Initialize Command Handler
**File**: `internal/ui/app/update.go`

Add this at the top (imports):
```go
import (
	// ... existing imports ...
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
)
```

Add this new message type:
```go
// initCCUsageMsg triggers CCUsage database initialization
type initCCUsageMsg struct{}
```

In Update() method, add case for key press 'I':
```go
case tea.KeyMsg:
	switch msg.String() {
	// ... existing cases ...

	case "i", "I":
		// Only in Usage tab when empty
		if m.currentView == ViewList && m.listTabs.ActiveIndex == 3 {
			if m.ccusageDB != nil && m.ccusageEntryCount == 0 {
				m.ccusageSyncing = true
				m.ccusageSyncStatus = "Initializing database..."
				return m, tea.Batch(
					m.initCCUsageCmd(),
					m.tickCmd(),
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

Handle the message in Update():
```go
case initCCUsageMsg:
	// Sync is complete
	return m, m.tickCmd()
```

---

### Step 2: Update Usage Tab View
**File**: `internal/ui/app/view.go`

Replace the entire `renderUsageTab()` function (around line 915):

See TUI_INTEGRATION_CODE_SNIPPETS.md for the complete function.

Add these two new methods:

```go
// renderClaudeCodeMetrics renders Claude Code usage data
func (m *Model) renderClaudeCodeMetrics() string {
	// See TUI_INTEGRATION_CODE_SNIPPETS.md for full implementation
}

// renderCombinedSummary renders both data sources combined
func (m *Model) renderCombinedSummary() string {
	// See TUI_INTEGRATION_CODE_SNIPPETS.md for full implementation
}
```

---

### Step 3: Initialize Database in Main
**File**: `cmd/server/main.go`

Add imports:
```go
import (
	// ... existing imports ...
	ccdb "github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	ccparser "github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
)
```

Add initialization before creating model:
```go
// Initialize CCUsage database
home, _ := os.UserHomeDir()
ccusageDBPath := filepath.Join(home, ".config/eye-in-the-sky/ccusage.sqlite")

var ccusageDB *ccdb.CCUsageDB
ccusageDB, err = ccdb.New(ccusageDBPath)
if err != nil {
	log.Printf("Warning: CCUsage database unavailable: %v", err)
	ccusageDB = nil
} else {
	defer ccusageDB.Close()

	// Perform initial sync
	syncMgr := ccparser.NewSyncManager(ccusageDB)
	if err := syncMgr.Sync(); err != nil {
		log.Printf("Warning: Initial CCUsage sync failed: %v", err)
	}
}
```

Update model creation:
```go
// OLD:
// tui, err := app.NewModel(agentDB)

// NEW:
tui, err := app.NewModel(agentDB, ccusageDB)
```

---

### Step 4: Add Initial CCUsage Load
**File**: `internal/ui/app/model.go`

In NewModel(), after loading agents, add:
```go
// Load initial ccusage data
if ccusageDB != nil {
	if err := m.loadCCUsageData(); err != nil {
		m.ccusageErr = err
	}
}
```

---

### Step 5: Add Periodic Sync Refresh
**File**: `internal/ui/app/update.go`

In the `tickMsg` case in Update(), add:
```go
// Refresh CCUsage data if needed (every 30 seconds)
if m.ccusageDB != nil && time.Since(m.lastCCUsageSync) > 30*time.Second {
	if err := m.loadCCUsageData(); err != nil {
		// Log but continue
		fmt.Fprintf(os.Stderr, "Warning: Failed to reload ccusage data: %v\n", err)
	}
}
```

---

## Implementation Order

Follow this order for cleanest implementation:

1. **Step 1** - Add Update logic (command handler)
2. **Step 2** - Update view rendering
3. **Step 3** - Initialize in main
4. **Step 4** - Load initial data in model
5. **Step 5** - Add periodic refresh

---

## Testing After Each Step

### After Step 1 (Command Handler)
```bash
go build ./cmd/server
# Verify no compile errors
```

### After Step 2 (View Updates)
```bash
go build ./cmd/server
# Verify rendering compiles
```

### After Step 3 (Main Init)
```bash
go build ./cmd/server
# Run server and verify no panic on startup
```

### After All Steps
```bash
go build ./cmd/server
./server

# In another terminal:
# Trigger initialization by pressing 'I' in Usage tab
# Verify:
# - "Initializing..." message appears
# - Database is populated
# - Data displays after sync
```

---

## Expected User Experience

### First Launch (Empty DB)
```
Press tab to navigate
Navigate to Usage tab ('U')
See: "No Claude Code usage data found. Press 'I' to initialize"
Press 'I'
See: "⏳ Initializing database..."
Wait for completion...
See: "✓ Initialized! Found X entries"
Data displays automatically
```

### Subsequent Launches
```
Launch app
Usage tab auto-loads Claude Code data
Every 30 seconds: auto-sync checks for new files
Tab updates silently if new data found
```

---

## File Locations for Reference

- **View code snippets**: `TUI_INTEGRATION_CODE_SNIPPETS.md`
- **Initialize button details**: `INITIALIZE_DATABASE_BUTTON.md`
- **Implementation checklist**: `IMPLEMENTATION_CHECKLIST.md`
- **Database schema**: `DATABASE_SCHEMA_AND_POPULATION.md`

---

## Estimated Time

- Step 1: 20 min
- Step 2: 30 min
- Step 3: 15 min
- Step 4: 5 min
- Step 5: 10 min
- Testing: 30 min

**Total: ~2 hours**

---

## Current Status

**Completed**: 60%
**Remaining**: 40%

Backend fully functional. Need to:
- Add command handler
- Update UI rendering
- Initialize in main

---

## Next Commit Will Include

After completing steps 1-5, the next commit should include:
- update.go (command handler)
- view.go (rendering)
- cmd/server/main.go (initialization)

This will complete the full CCUsage integration!

---

**Ready to continue? Follow the remaining steps above!**
