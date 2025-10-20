package api

import (
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
)

// DailyReport represents daily usage data for display
type DailyReport struct {
	Date              string
	Project           string
	InputTokens       int
	OutputTokens      int
	CacheTokens       int
	TotalCost         float64
	CostPerInputToken  float64
	CostPerOutputToken float64
}

// SessionReport represents session usage data for display
type SessionReport struct {
	SessionID    string
	Project      string
	StartTime    string
	Duration     string
	InputTokens  int
	OutputTokens int
	CacheTokens  int
	TotalCost    float64
}

// MonthlyReport represents monthly aggregated usage
type MonthlyReport struct {
	Month              string
	Days               int
	TotalInputTokens   int
	TotalOutputTokens  int
	TotalCacheTokens   int
	TotalCost          float64
	AverageDailyCost   float64
	HighestDailyCost   float64
	ProjectBreakdown   map[string]ProjectMonthlyData
}

// ProjectMonthlyData represents monthly data per project
type ProjectMonthlyData struct {
	Project           string
	InputTokens       int
	OutputTokens      int
	CacheTokens       int
	TotalCost         float64
	DaysActive        int
}

// CostSummary represents overall cost metrics
type CostSummary struct {
	TimeRange          string
	Days               int
	TotalInputTokens   int
	TotalOutputTokens  int
	TotalCacheTokens   int
	TotalCost          float64
	AverageDailyCost   float64
	AverageSessionCost float64
	SessionCount       int
	ProjectCount       int
	ModelsUsed         []string
}

// ActiveBlockReport represents the current billing block
type ActiveBlockReport struct {
	StartTime        string
	EndTime          string
	TimeRemaining    string
	InputTokens      int
	OutputTokens     int
	CacheTokens      int
	TotalCost        float64
	ProjectBreakdown map[string]BlockProjectBreakdown
}

// BlockProjectBreakdown represents project breakdown in active block
type BlockProjectBreakdown struct {
	Project      string
	InputTokens  int
	OutputTokens int
	CacheTokens  int
	TotalCost    float64
}

// GetDailyUsageReport returns daily usage data for the last N days
func GetDailyUsageReport(ccdb *db.CCUsageDB, days int) ([]DailyReport, error) {
	if days <= 0 {
		days = 7
	}

	now := time.Now()
	since := now.AddDate(0, 0, -days)

	dailyUsage, err := ccdb.GetDailyUsage(since, now, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage: %w", err)
	}

	var reports []DailyReport
	for _, d := range dailyUsage {
		report := DailyReport{
			Date:        d.Date,
			Project:     d.Project,
			InputTokens: d.InputTokens,
			OutputTokens: d.OutputTokens,
			CacheTokens: d.CacheCreationTokens + d.CacheReadTokens,
			TotalCost:   d.TotalCost,
		}

		// Calculate per-token costs
		totalTokens := d.InputTokens + d.OutputTokens
		if totalTokens > 0 {
			report.CostPerInputToken = d.TotalCost / float64(d.InputTokens)
			report.CostPerOutputToken = d.TotalCost / float64(d.OutputTokens)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// GetSessionUsageReport returns session usage data
func GetSessionUsageReport(ccdb *db.CCUsageDB, limit int) ([]SessionReport, error) {
	if limit <= 0 {
		limit = 10
	}

	sessions, err := ccdb.GetSessionData("", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	var reports []SessionReport
	for _, s := range sessions {
		duration := s.EndTime.Sub(s.StartTime)
		durationStr := fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)

		report := SessionReport{
			SessionID:    s.SessionID,
			Project:      s.Project,
			StartTime:    s.StartTime.Format("2006-01-02 15:04"),
			Duration:     durationStr,
			InputTokens:  s.InputTokens,
			OutputTokens: s.OutputTokens,
			CacheTokens:  s.CacheTokens,
			TotalCost:    s.TotalCost,
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// GetMonthlyUsageReport returns monthly usage data
func GetMonthlyUsageReport(ccdb *db.CCUsageDB, year, month int) (*MonthlyReport, error) {
	if month <= 0 || month > 12 {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}

	usage, err := ccdb.GetMonthlyUsage(year, month, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly usage: %w", err)
	}

	// Get daily data for this month to calculate additional metrics
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	now := time.Now()
	if endDate.After(now) {
		endDate = now
	}

	dailyUsage, err := ccdb.GetDailyUsage(startDate, endDate, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage for monthly report: %w", err)
	}

	totalInputTokens := usage.InputTokens
	totalOutputTokens := usage.OutputTokens
	totalCacheTokens := usage.CacheCreationTokens + usage.CacheReadTokens
	totalCost := usage.TotalCost

	dayCount := len(dailyUsage)
	if dayCount == 0 {
		dayCount = 1 // Avoid division by zero
	}

	avgDailyCost := totalCost / float64(dayCount)
	highestDailyCost := 0.0
	for _, d := range dailyUsage {
		if d.TotalCost > highestDailyCost {
			highestDailyCost = d.TotalCost
		}
	}

	report := &MonthlyReport{
		Month:              startDate.Format("2006-01"),
		Days:               dayCount,
		TotalInputTokens:   totalInputTokens,
		TotalOutputTokens:  totalOutputTokens,
		TotalCacheTokens:   totalCacheTokens,
		TotalCost:          totalCost,
		AverageDailyCost:   avgDailyCost,
		HighestDailyCost:   highestDailyCost,
		ProjectBreakdown:   make(map[string]ProjectMonthlyData),
	}

	// Build project breakdown
	for _, d := range dailyUsage {
		if proj, exists := report.ProjectBreakdown[d.Project]; exists {
			proj.InputTokens += d.InputTokens
			proj.OutputTokens += d.OutputTokens
			proj.CacheTokens += d.CacheCreationTokens + d.CacheReadTokens
			proj.TotalCost += d.TotalCost
			proj.DaysActive++
			report.ProjectBreakdown[d.Project] = proj
		} else {
			report.ProjectBreakdown[d.Project] = ProjectMonthlyData{
				Project:      d.Project,
				InputTokens:  d.InputTokens,
				OutputTokens: d.OutputTokens,
				CacheTokens:  d.CacheCreationTokens + d.CacheReadTokens,
				TotalCost:    d.TotalCost,
				DaysActive:   1,
			}
		}
	}

	return report, nil
}

// GetTotalCostSummary returns overall cost metrics
func GetTotalCostSummary(ccdb *db.CCUsageDB, days int) (*CostSummary, error) {
	if days <= 0 {
		days = 30
	}

	now := time.Now()
	since := now.AddDate(0, 0, -days)

	// Get daily usage
	dailyUsage, err := ccdb.GetDailyUsage(since, now, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage: %w", err)
	}

	// Get session data
	sessions, err := ccdb.GetSessionData("", 1000) // Get many sessions
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	// Calculate totals
	summary := &CostSummary{
		TimeRange:      fmt.Sprintf("Last %d days", days),
		Days:           days,
		ModelsUsed:     []string{},
		ProjectCount:   0,
		SessionCount:   len(sessions),
	}

	// Aggregate from daily data
	projectMap := make(map[string]bool)
	modelMap := make(map[string]bool)

	for _, d := range dailyUsage {
		summary.TotalInputTokens += d.InputTokens
		summary.TotalOutputTokens += d.OutputTokens
		summary.TotalCacheTokens += d.CacheCreationTokens + d.CacheReadTokens
		summary.TotalCost += d.TotalCost
		projectMap[d.Project] = true
	}

	// Extract unique projects
	for range projectMap {
		summary.ProjectCount++
	}

	for model := range modelMap {
		summary.ModelsUsed = append(summary.ModelsUsed, model)
	}

	// Calculate averages
	if len(dailyUsage) > 0 {
		summary.AverageDailyCost = summary.TotalCost / float64(len(dailyUsage))
	}

	if len(sessions) > 0 {
		summary.AverageSessionCost = summary.TotalCost / float64(len(sessions))
	}

	return summary, nil
}

// GetActiveBlockReport returns the current 5-hour billing block
func GetActiveBlockReport(ccdb *db.CCUsageDB) (*ActiveBlockReport, error) {
	block, err := ccdb.GetActiveBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get active block: %w", err)
	}

	now := time.Now()
	timeRemaining := block.EndTime.Sub(now)
	timeRemainingStr := fmt.Sprintf("%dh %dm %ds",
		int(timeRemaining.Hours()),
		int(timeRemaining.Minutes())%60,
		int(timeRemaining.Seconds())%60,
	)

	report := &ActiveBlockReport{
		StartTime:        block.StartTime.Format("2006-01-02 15:04"),
		EndTime:          block.EndTime.Format("2006-01-02 15:04"),
		TimeRemaining:    timeRemainingStr,
		InputTokens:      block.InputTokens,
		OutputTokens:     block.OutputTokens,
		CacheTokens:      block.CacheTokens,
		TotalCost:        block.TotalCost,
		ProjectBreakdown: make(map[string]BlockProjectBreakdown),
	}

	// Build project breakdown
	for project, data := range block.ProjectBreakdown {
		report.ProjectBreakdown[project] = BlockProjectBreakdown{
			Project:      project,
			InputTokens:  data.InputTokens,
			OutputTokens: data.OutputTokens,
			CacheTokens:  data.CacheTokens,
			TotalCost:    data.TotalCost,
		}
	}

	return report, nil
}
