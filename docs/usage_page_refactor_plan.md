
# Usage Page Architecture Refactoring Plan

## Overview
Refactor the monolithic Model architecture to separate data aggregation from rendering, improving testability, maintainability, and performance.

## Phase 1: Extract Service Layer (Priority: Architecture)

### 1.1 Create internal/ui/services/usage_service.go
- Extract all data aggregation logic from Model
- New `UsageService` struct with methods:
  - `GetSessionMetricsSummary([]SessionMetric) UsageSummary`
  - `GetMonthlyCostsSummary([]*SessionMetric) UsageSummary`
  - `GetClaudeDailySummary([]api.DailyReport) UsageSummary`
  - `GetClaudeMonthlyReport([]api.MonthlyReport) UsageSummary`
  - `GetTotalCost([]SessionMetric, []api.MonthlyReport) float64`

### 1.2 Create internal/ui/models/usage_view_model.go
Define `UsageViewModel` with precomputed view data.

```go
type UsageViewModel struct {
    SummaryBar string
    Sections   []UsageSection
    LastSync   time.Time
    HasData    bool
}

type UsageSection struct {
    Title   string
    Content string
    HasData bool
}
```

### 1.3 Migrate Build Functions
- Move BuildSessionMetricsSummary, BuildMonthlyCostsSummary, BuildClaudeDailySummary, and BuildClaudeMonthlyReport from `usage_summary.go` to `usage_service.go`.
- Keep `UsageSummary`, `UsageRow`, and `UsageTotals` in `usage_data.go` (data only).

### 1.4 Update Model.renderUsageTab()
- Call `m.usageService.Build...()` instead of local `Build...()` functions.
- Build `UsageViewModel` once per refresh cycle.
- Pass `UsageViewModel` to pure rendering functions.

## File Structure After Refactoring

```
internal/ui/
├── services/
│   └── usage_service.go       # All data aggregation logic
├── models/
│   └── usage_view_model.go    # Precomputed view state
└── app/
    ├── usage_data.go          # Data structures only
    ├── usage_renderer.go      # Pure rendering (takes ViewModel)
    ├── view_usage.go          # Orchestration layer
    └── model.go               # Inject UsageService
```

## Benefits
1. **Testability:** Service layer can be unit tested without UI dependencies.
2. **Maintainability:** Clear separation of concerns (data → processing → rendering).
3. **Performance:** Foundation for caching (Phase 2).
4. **Extensibility:** New data sources can be added without touching rendering.

## Implementation Notes
- Maintain backward compatibility during migration.
- Each step should compile and pass tests.
- No behavioral changes in this phase.
- Set up foundation for caching and UX improvements (Phases 2 and 3).
