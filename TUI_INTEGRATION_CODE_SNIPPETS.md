# TUI Integration - Code Snippets

Copy-paste ready code snippets for integrating CCUsage into the Eye-in-the-Sky Usage tab.

## 1. Model Struct Updates

**File**: `internal/ui/app/model.go`

Add these fields to the `Model` struct (around line 26):

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

	// CCUsage timing
	lastCCUsageSync time.Time
	ccusageErr      error
}
```

Add imports (at the top of model.go):

```go
import (
	// ... existing imports ...
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
)
```

## 2. Update NewModel() Constructor

**File**: `internal/ui/app/model.go`

Update the function signature and initialization (around line 214):

```go
// NewModel creates a new application model
func NewModel(db *sql.DB, ccusageDB *db.CCUsageDB) (*Model, error) {
	// ... existing code up to Model creation ...

	m := &Model{
		db:            db,
		ccusageDB:     ccusageDB,  // ADD THIS LINE
		config:        config,
		keys:          keys,
		theme:         theme,
		styles:        styles,
		help:          helpModel,
		showHelp:      false,
		windowFocuser: windowFocuser,
		claudePath:    claudePath,
		mdRenderer:    mdRenderer,
		tabs:          tabs,
		listTabs:      listTabs,
		currentView:   ViewList,
		showAll:       config.ShowAllAgents,
		agents:        []Agent{},
	}

	// Load initial agent list
	if err := m.loadAgents(); err != nil {
		m.err = err
	}

	// ADD THESE LINES - Load initial CCUsage data
	if ccusageDB != nil {
		if err := m.loadCCUsageData(); err != nil {
			m.ccusageErr = err
		}
	}

	return m, nil
}
```

## 3. Add loadCCUsageData() Method

**File**: `internal/ui/app/model.go`

Add this new method anywhere in the Model struct implementation (good place after `loadSessionMetrics()`):

```go
// loadCCUsageData loads Claude Code usage data from ccusage database
func (m *Model) loadCCUsageData() error {
	if m.ccusageDB == nil {
		return nil
	}

	// Load daily usage (last 7 days)
	daily, err := api.GetDailyUsageReport(m.ccusageDB, 7)
	if err != nil {
		return fmt.Errorf("failed to load daily usage: %w", err)
	}
	m.ccusageDaily = daily

	// Load sessions
	sessions, err := api.GetSessionUsageReport(m.ccusageDB, 10)
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}
	m.ccusageSessions = sessions

	// Load monthly summary
	now := time.Now()
	monthly, err := api.GetMonthlyUsageReport(m.ccusageDB, now.Year(), int(now.Month()))
	if err != nil {
		return fmt.Errorf("failed to load monthly: %w", err)
	}
	m.ccusageMonthly = monthly

	// Load active block
	block, err := api.GetActiveBlockReport(m.ccusageDB)
	if err != nil {
		return fmt.Errorf("failed to load active block: %w", err)
	}
	m.ccusageBlock = block

	// Load cost summary
	costs, err := api.GetTotalCostSummary(m.ccusageDB, 30)
	if err != nil {
		return fmt.Errorf("failed to load cost summary: %w", err)
	}
	m.ccusageCosts = costs

	m.lastCCUsageSync = time.Now()
	return nil
}
```

## 4. Update renderUsageTab()

**File**: `internal/ui/app/view.go`

Replace the entire `renderUsageTab()` function (starting around line 915):

```go
// renderUsageTab renders session costs for all agents with monthly breakdown
func (m *Model) renderUsageTab() string {
	var b strings.Builder

	// Render Eye-in-the-Sky metrics
	if len(m.allSessionMetrics) == 0 && len(m.monthlyCosts) == 0 && m.ccusageDB == nil {
		return m.styles.Subtle.Render("  No usage data available")
	}

	// Eye-in-the-Sky Section (existing)
	if len(m.allSessionMetrics) > 0 {
		b.WriteString(m.styles.Title.Render("Eye-in-the-Sky Session Metrics"))
		b.WriteString("\n\n")

		// Column headers
		headerLine := fmt.Sprintf("%-10s %-8s %-12s %-10s %-12s",
			"Agent", "Usage %", "Tokens", "Cost USD", "Model")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		// Data rows
		totalCost := 0.0
		totalTokens := 0

		for _, metric := range m.allSessionMetrics {
			usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

			agentID := truncate(metric.AgentID, 8)
			usage := fmt.Sprintf("%.1f%%", usagePercent)
			tokens := fmt.Sprintf("%d", metric.TokensUsed)
			cost := fmt.Sprintf("$%.4f", metric.EstimatedCostUSD)
			model := truncate(metric.ModelName, 12)

			line := fmt.Sprintf("%-10s %-8s %-12s %-10s %-12s",
				agentID, usage, tokens, cost, model)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")

			totalCost += metric.EstimatedCostUSD
			totalTokens += metric.TokensUsed
		}

		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Total: $%.4f | %d tokens\n", totalCost, totalTokens)))
	}

	// Monthly Cost Breakdown (existing)
	if len(m.monthlyCosts) > 0 {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Monthly Cost Breakdown"))
		b.WriteString("\n\n")

		// Column headers
		headerLine := fmt.Sprintf("%-20s %-8s %-12s %-10s %-8s",
			"Timestamp", "Usage %", "Tokens Used", "Cost USD", "Model")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		// Data rows
		totalMonthlyCost := 0.0
		totalMonthlyTokens := 0

		for _, metric := range m.monthlyCosts {
			usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

			timestamp := metric.Timestamp.Format("2006-01-02 15:04")
			usage := fmt.Sprintf("%.1f%%", usagePercent)
			tokens := fmt.Sprintf("%d", metric.TokensUsed)
			cost := fmt.Sprintf("$%.4f", metric.EstimatedCostUSD)
			model := truncate(metric.ModelName, 8)

			line := fmt.Sprintf("%-20s %-8s %-12s %-10s %-8s",
				timestamp, usage, tokens, cost, model)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")

			totalMonthlyCost += metric.EstimatedCostUSD
			totalMonthlyTokens += metric.TokensUsed
		}

		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Monthly Total: $%.4f | %d tokens across %d snapshots\n",
			totalMonthlyCost, totalMonthlyTokens, len(m.monthlyCosts))))
	}

	// Claude Code Section (NEW)
	if m.ccusageDB != nil {
		b.WriteString(m.renderClaudeCodeMetrics())
		b.WriteString("\n\n")
		b.WriteString(m.renderCombinedSummary())
	}

	return b.String()
}
```

## 5. Add renderClaudeCodeMetrics() Method

**File**: `internal/ui/app/view.go`

Add this new method (good place right after `renderUsageTab()`):

```go
// renderClaudeCodeMetrics renders Claude Code usage data
func (m *Model) renderClaudeCodeMetrics() string {
	var b strings.Builder

	// Daily Usage Section
	if len(m.ccusageDaily) > 0 {
		b.WriteString(m.styles.Title.Render("Claude Code Daily Usage (Last 7 Days)"))
		b.WriteString("\n\n")

		headerLine := fmt.Sprintf("%-12s %-15s %-10s %-10s %-10s",
			"Date", "Project", "Input", "Output", "Cost USD")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		totalCost := 0.0
		for _, report := range m.ccusageDaily {
			line := fmt.Sprintf("%-12s %-15s %-10d %-10d %-10s",
				report.Date,
				truncate(report.Project, 12),
				report.InputTokens,
				report.OutputTokens,
				fmt.Sprintf("$%.4f", report.TotalCost))
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
			totalCost += report.TotalCost
		}

		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Daily Total: $%.4f\n", totalCost)))
	}

	// Session Usage Section
	if len(m.ccusageSessions) > 0 {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Claude Code Sessions"))
		b.WriteString("\n\n")

		headerLine := fmt.Sprintf("%-20s %-15s %-12s %-10s %-10s",
			"SessionId", "Project", "Duration", "Tokens", "Cost USD")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		for _, report := range m.ccusageSessions {
			tokens := report.InputTokens + report.OutputTokens
			line := fmt.Sprintf("%-20s %-15s %-12s %-10d %-10s",
				truncate(report.SessionID, 18),
				truncate(report.Project, 13),
				report.Duration,
				tokens,
				fmt.Sprintf("$%.4f", report.TotalCost))
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}
	}

	// Monthly Summary Section
	if m.ccusageMonthly != nil {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Claude Code Monthly Summary"))
		b.WriteString("\n\n")

		summary := fmt.Sprintf("Month: %s | Days: %d | Total Tokens: %d | Total Cost: $%.4f\n",
			m.ccusageMonthly.Month,
			m.ccusageMonthly.Days,
			m.ccusageMonthly.TotalInputTokens+m.ccusageMonthly.TotalOutputTokens,
			m.ccusageMonthly.TotalCost)
		b.WriteString(m.styles.Text.Render(summary))

		if m.ccusageMonthly.AverageDailyCost > 0 {
			summary2 := fmt.Sprintf("Avg Daily: $%.4f | Highest Day: $%.4f\n",
				m.ccusageMonthly.AverageDailyCost,
				m.ccusageMonthly.HighestDailyCost)
			b.WriteString(m.styles.Subtle.Render(summary2))
		}
	}

	// Active Block Section
	if m.ccusageBlock != nil {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Current Billing Block"))
		b.WriteString("\n\n")

		blockInfo := fmt.Sprintf("Block: %s to %s | Time Remaining: %s\n",
			m.ccusageBlock.StartTime,
			m.ccusageBlock.EndTime,
			m.ccusageBlock.TimeRemaining)
		b.WriteString(m.styles.Text.Render(blockInfo))

		blockUsage := fmt.Sprintf("Usage: %d input | %d output | $%.4f total\n",
			m.ccusageBlock.InputTokens,
			m.ccusageBlock.OutputTokens,
			m.ccusageBlock.TotalCost)
		b.WriteString(m.styles.Text.Render(blockUsage))
	}

	return b.String()
}

// renderCombinedSummary renders both data sources combined
func (m *Model) renderCombinedSummary() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Combined Cost Summary"))
	b.WriteString("\n\n")

	eyeInTheSkyTotal := 0.0
	for _, metric := range m.monthlyCosts {
		eyeInTheSkyTotal += metric.EstimatedCostUSD
	}

	claudeCodeTotal := 0.0
	if m.ccusageMonthly != nil {
		claudeCodeTotal = m.ccusageMonthly.TotalCost
	}

	grandTotal := eyeInTheSkyTotal + claudeCodeTotal

	line1 := fmt.Sprintf("Eye-in-the-Sky Sessions:  $%.4f", eyeInTheSkyTotal)
	b.WriteString(m.styles.Text.Render(line1))
	b.WriteString("\n")

	line2 := fmt.Sprintf("Claude Code Usage:        $%.4f", claudeCodeTotal)
	b.WriteString(m.styles.Text.Render(line2))
	b.WriteString("\n")

	line3 := "─────────────────────────────────"
	b.WriteString(m.styles.Border.Render(line3))
	b.WriteString("\n")

	line4 := fmt.Sprintf("Total Monthly Cost:       $%.4f", grandTotal)
	b.WriteString(m.styles.Primary.Render(line4))
	b.WriteString("\n")

	return b.String()
}
```

## 6. Add Refresh Logic

**File**: `internal/ui/app/update.go`

Find the `Update()` method and locate the `tickMsg` case handler. Add this logic:

```go
case tickMsg:
	// ... existing refresh logic ...

	// Refresh CCUsage data if needed (every 30 seconds)
	if m.ccusageDB != nil && time.Since(m.lastCCUsageSync) > 30*time.Second {
		if err := m.loadCCUsageData(); err != nil {
			m.ccusageErr = err
			// Log but don't crash
			fmt.Fprintf(os.Stderr, "Warning: Failed to load ccusage data: %v\n", err)
		}
	}

	return m, m.tickCmd()
```

## 7. Initialize Database in main()

**File**: `cmd/server/main.go`

Add imports:

```go
import (
	// ... existing imports ...
	ccdb "github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	ccparser "github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
)
```

Add database initialization (before creating TUI model):

```go
// Initialize CCUsage database
var ccusageDB *ccdb.CCUsageDB
ccusageDBPath := filepath.Join(home, ".config/eye-in-the-sky/ccusage.sqlite")
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

// Pass CCUsage database to TUI model
tui, err := app.NewModel(agentDB, ccusageDB)
if err != nil {
	log.Fatal(err)
}
```

## Summary of Changes

| File | Changes |
|------|---------|
| `internal/ui/app/model.go` | Add ccusage fields, imports, loader method |
| `internal/ui/app/view.go` | Add two render methods, update renderUsageTab |
| `internal/ui/app/update.go` | Add 30-second refresh logic |
| `cmd/server/main.go` | Initialize ccusageDB and pass to model |

Total: ~4 files, ~150 lines of new code
