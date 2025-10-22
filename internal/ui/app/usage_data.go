package app

import (
	"sort"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// UsageSummary holds aggregated usage data ready for rendering
type UsageSummary struct {
	Rows        []UsageRow
	Totals      UsageTotals
	HasData     bool
}

// UsageRow represents a single row of usage data
type UsageRow struct {
	Timestamp    string
	Usage        float64
	TokensUsed   int
	Cost         float64
	Model        string
	AgentID      string
	TokensBudget int
}

// UsageTotals holds aggregate totals for a usage section
type UsageTotals struct {
	CostUSD      float64
	Tokens       int
	Snapshots    int
}

// BuildSessionMetricsSummary processes Eye-in-the-Sky session metrics
func BuildSessionMetricsSummary(metrics []domain.SessionMetric) UsageSummary {
	if len(metrics) == 0 {
		return UsageSummary{HasData: false}
	}

	rows := make([]UsageRow, 0, len(metrics))
	totals := UsageTotals{}

	for _, metric := range metrics {
		usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

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
		HasData: true,
	}
}

// BuildMonthlyCostsSummary processes monthly cost data
func BuildMonthlyCostsSummary(metrics []*domain.SessionMetric) UsageSummary {
	if len(metrics) == 0 {
		return UsageSummary{HasData: false}
	}

	rows := make([]UsageRow, 0, len(metrics))
	totals := UsageTotals{}

	for _, metric := range metrics {
		if metric == nil {
			continue
		}
		usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

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
		HasData: true,
	}
}

// BuildClaudeDailySummary aggregates Claude Code daily usage data
func BuildClaudeDailySummary(reports []api.DailyReport) UsageSummary {
	if len(reports) == 0 {
		return UsageSummary{HasData: false}
	}

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

		// Store the date and individual token counts in a formatted way for the renderer
		rows = append(rows, UsageRow{
			Timestamp:  date,
			TokensUsed: totalTokens,
			Cost:       report.TotalCost,
			// Store individual token counts in the Model field for now (renderer will format properly)
			Model: formatDailyTokens(report),
		})

		totals.CostUSD += report.TotalCost
		totals.Tokens += totalTokens
	}

	totals.Snapshots = len(dailyTotals)

	return UsageSummary{
		Rows:    rows,
		Totals:  totals,
		HasData: true,
	}
}

// BuildClaudeMonthlyReport processes Claude Code monthly summary data
func BuildClaudeMonthlyReport(reports []api.MonthlyReport) UsageSummary {
	if len(reports) == 0 {
		return UsageSummary{HasData: false}
	}

	rows := make([]UsageRow, 0, len(reports))
	totals := UsageTotals{}

	for _, monthly := range reports {
		totalTokensMonth := monthly.TotalInputTokens + monthly.TotalOutputTokens + monthly.TotalCacheTokens

		rows = append(rows, UsageRow{
			Timestamp:  monthly.Month,
			TokensUsed: totalTokensMonth,
			Cost:       monthly.TotalCost,
			// Store days count in TokensBudget field (renderer will handle it)
			TokensBudget: monthly.Days,
			// Store token breakdown in Model field
			Model: formatMonthlyTokens(&monthly),
		})

		totals.CostUSD += monthly.TotalCost
		totals.Tokens += totalTokensMonth
	}

	totals.Snapshots = len(reports)

	return UsageSummary{
		Rows:    rows,
		Totals:  totals,
		HasData: true,
	}
}

// formatDailyTokens creates a formatted string for daily token breakdown
// This is temporary storage - the renderer will properly format this data
func formatDailyTokens(report *api.DailyReport) string {
	// The renderer will parse this and format it properly
	return "tokens"
}

// formatMonthlyTokens creates a formatted string for monthly token breakdown
func formatMonthlyTokens(report *api.MonthlyReport) string {
	// The renderer will parse this and format it properly
	return "tokens"
}