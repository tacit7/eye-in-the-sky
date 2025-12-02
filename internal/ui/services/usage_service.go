package services

import (
	"sort"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// Clock provides current time (injected for testability)
type Clock interface {
	Now() time.Time
}

// SystemClock implements Clock using system time
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

// UsageService handles data aggregation for usage metrics
type UsageService struct {
	clock Clock
}

// NewUsageService creates a new usage service
func NewUsageService(c Clock) *UsageService {
	return &UsageService{clock: c}
}

// Inputs contains all data needed to build usage summary
type Inputs struct {
	Sessions  []domain.SessionMetric  // latest per agent
	Monthly   []domain.SessionMetric  // time-stamped snapshots
	DailyCC   []api.DailyReport
	MonthlyCC []api.MonthlyReport
	ActiveBlock *api.ActiveBlockReport
	LastSync  time.Time
}

// UsageRow represents a single row of usage data
type UsageRow struct {
	Timestamp    string
	AgentID      string
	Model        string
	Usage        float64
	TokensUsed   int
	TokensBudget int
	Cost         float64

	// Optional breakdowns for Claude Code data
	InputTokens  int
	OutputTokens int
	CacheCreate  int
	CacheRead    int
	Days         int // for monthly reports
}

// UsageTotals holds aggregate totals for a usage section
type UsageTotals struct {
	CostUSD   float64
	Tokens    int
	Snapshots int
}

// UsageSummary holds aggregated usage data ready for rendering
type UsageSummary struct {
	Rows    []UsageRow
	Totals  UsageTotals
	HasData bool
}

// Summary is the complete output from Build()
type Summary struct {
	Sessions   UsageSummary
	Monthly    UsageSummary
	DailyCC    UsageSummary
	MonthlyCC  UsageSummary
	TotalsUSD  float64
	TokensTotal int
	LastSync   time.Time
}

// Build performs pure aggregation and sorting; no formatting
func (s *UsageService) Build(in Inputs) (Summary, error) {
	summary := Summary{
		LastSync: s.clock.Now(),
	}

	// Aggregate session metrics
	if len(in.Sessions) > 0 {
		summary.Sessions = s.buildSessionMetrics(in.Sessions)
	}

	// Aggregate monthly costs
	if len(in.Monthly) > 0 {
		summary.Monthly = s.buildMonthlyCosts(in.Monthly)
	}

	// Aggregate Claude Code daily
	if len(in.DailyCC) > 0 {
		summary.DailyCC = s.buildDailySummary(in.DailyCC)
	}

	// Aggregate Claude Code monthly
	if len(in.MonthlyCC) > 0 {
		summary.MonthlyCC = s.buildMonthlyReport(in.MonthlyCC)
	}

	// Calculate totals (use DailyCC since it's more granular; MonthlyCC is just a different aggregation of the same data)
	summary.TotalsUSD = summary.Sessions.Totals.CostUSD +
		summary.Monthly.Totals.CostUSD +
		summary.DailyCC.Totals.CostUSD

	summary.TokensTotal = summary.Sessions.Totals.Tokens +
		summary.Monthly.Totals.Tokens +
		summary.DailyCC.Totals.Tokens

	return summary, nil
}

func (s *UsageService) buildSessionMetrics(metrics []domain.SessionMetric) UsageSummary {
	rows := make([]UsageRow, 0, len(metrics))
	totals := UsageTotals{}

	for _, metric := range metrics {
		usagePercent := 0.0
		if metric.TokensBudget > 0 {
			usagePercent = float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100
		}

		rows = append(rows, UsageRow{
			AgentID:      string(metric.AgentID),
			Usage:        usagePercent,
			TokensUsed:   metric.TokensUsed,
			Cost:         metric.EstimatedCostUSD,
			Model:        metric.ModelName,
			TokensBudget: metric.TokensBudget,
		})

		totals.CostUSD += metric.EstimatedCostUSD
		totals.Tokens += metric.TokensUsed
	}

	totals.Snapshots = len(metrics)

	return UsageSummary{
		Rows:    rows,
		Totals:  totals,
		HasData: len(rows) > 0,
	}
}

func (s *UsageService) buildMonthlyCosts(metrics []domain.SessionMetric) UsageSummary {
	rows := make([]UsageRow, 0, len(metrics))
	totals := UsageTotals{}

	for _, metric := range metrics {
		usagePercent := 0.0
		if metric.TokensBudget > 0 {
			usagePercent = float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100
		}

		rows = append(rows, UsageRow{
			Timestamp:    metric.Timestamp.Format("2006-01-02 15:04"),
			Usage:        usagePercent,
			TokensUsed:   metric.TokensUsed,
			Cost:         metric.EstimatedCostUSD,
			Model:        metric.ModelName,
			TokensBudget: metric.TokensBudget,
		})

		totals.CostUSD += metric.EstimatedCostUSD
		totals.Tokens += metric.TokensUsed
	}

	totals.Snapshots = len(metrics)

	return UsageSummary{
		Rows:    rows,
		Totals:  totals,
		HasData: len(rows) > 0,
	}
}

func (s *UsageService) buildDailySummary(reports []api.DailyReport) UsageSummary {
	// Aggregate by date only
	dailyTotals := make(map[string]*api.DailyReport)
	for _, report := range reports {
		if daily, exists := dailyTotals[report.Date]; exists {
			daily.InputTokens += report.InputTokens
			daily.OutputTokens += report.OutputTokens
			daily.CacheCreationTokens += report.CacheCreationTokens
			daily.CacheReadTokens += report.CacheReadTokens
			daily.TotalCost += report.TotalCost
		} else {
			newReport := report
			dailyTotals[report.Date] = &newReport
		}
	}

	// Sort dates in reverse order (most recent first)
	dates := make([]string, 0, len(dailyTotals))
	for date := range dailyTotals {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	// Build rows and calculate totals
	rows := make([]UsageRow, 0, len(dailyTotals))
	totals := UsageTotals{}

	for _, date := range dates {
		report := dailyTotals[date]
		totalTokens := report.InputTokens + report.OutputTokens +
			report.CacheCreationTokens + report.CacheReadTokens

		// Parse date and format with weekday (e.g., "Mon 2025-10-22")
		formattedDate := date // fallback to original
		if parsedDate, err := time.Parse("2006-01-02", date); err == nil {
			formattedDate = parsedDate.Format("Mon 2006-01-02")
		}

		rows = append(rows, UsageRow{
			Timestamp:    formattedDate,
			TokensUsed:   totalTokens,
			Cost:         report.TotalCost,
			InputTokens:  report.InputTokens,
			OutputTokens: report.OutputTokens,
			CacheCreate:  report.CacheCreationTokens,
			CacheRead:    report.CacheReadTokens,
		})

		totals.CostUSD += report.TotalCost
		totals.Tokens += totalTokens
	}

	totals.Snapshots = len(dailyTotals)

	return UsageSummary{
		Rows:    rows,
		Totals:  totals,
		HasData: len(rows) > 0,
	}
}

func (s *UsageService) buildMonthlyReport(reports []api.MonthlyReport) UsageSummary {
	rows := make([]UsageRow, 0, len(reports))
	totals := UsageTotals{}

	for _, monthly := range reports {
		totalTokensMonth := monthly.TotalInputTokens + monthly.TotalOutputTokens + monthly.TotalCacheTokens

		rows = append(rows, UsageRow{
			Timestamp:    monthly.Month,
			TokensUsed:   totalTokensMonth,
			Cost:         monthly.TotalCost,
			InputTokens:  monthly.TotalInputTokens,
			OutputTokens: monthly.TotalOutputTokens,
			CacheCreate:  monthly.TotalCacheTokens,
			Days:         monthly.Days,
		})

		totals.CostUSD += monthly.TotalCost
		totals.Tokens += totalTokensMonth
	}

	totals.Snapshots = len(reports)

	return UsageSummary{
		Rows:    rows,
		Totals:  totals,
		HasData: len(rows) > 0,
	}
}
