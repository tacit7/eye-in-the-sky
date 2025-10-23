package services

import (
	"testing"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// MockClock implements Clock for testing
type MockClock struct {
	current time.Time
}

func (m *MockClock) Now() time.Time { return m.current }

func TestUsageService_Build_EmptyInputs(t *testing.T) {
	clock := &MockClock{current: time.Date(2025, 10, 22, 19, 4, 0, 0, time.UTC)}
	svc := NewUsageService(clock)

	summary, err := svc.Build(Inputs{})

	if err != nil {
		t.Errorf("Build() returned unexpected error: %v", err)
	}

	if summary.Sessions.HasData {
		t.Error("Expected Sessions.HasData = false for empty input")
	}

	if summary.Monthly.HasData {
		t.Error("Expected Monthly.HasData = false for empty input")
	}

	if summary.DailyCC.HasData {
		t.Error("Expected DailyCC.HasData = false for empty input")
	}

	if summary.MonthlyCC.HasData {
		t.Error("Expected MonthlyCC.HasData = false for empty input")
	}

	if summary.TotalsUSD != 0.0 {
		t.Errorf("Expected TotalsUSD = 0.0, got %f", summary.TotalsUSD)
	}

	if summary.TokensTotal != 0 {
		t.Errorf("Expected TokensTotal = 0, got %d", summary.TokensTotal)
	}

	if !summary.LastSync.Equal(clock.current) {
		t.Errorf("Expected LastSync = %v, got %v", clock.current, summary.LastSync)
	}
}

func TestUsageService_Build_SessionMetrics(t *testing.T) {
	clock := &MockClock{current: time.Date(2025, 10, 22, 19, 4, 0, 0, time.UTC)}
	svc := NewUsageService(clock)

	inputs := Inputs{
		Sessions: []domain.SessionMetric{
			{
				AgentID:          "agent-1",
				TokensUsed:       1000,
				TokensBudget:     2000,
				EstimatedCostUSD: 0.05,
				ModelName:        "claude-sonnet",
			},
			{
				AgentID:          "agent-2",
				TokensUsed:       500,
				TokensBudget:     1000,
				EstimatedCostUSD: 0.025,
				ModelName:        "claude-haiku",
			},
		},
	}

	summary, err := svc.Build(inputs)

	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}

	if !summary.Sessions.HasData {
		t.Fatal("Expected Sessions.HasData = true")
	}

	if len(summary.Sessions.Rows) != 2 {
		t.Fatalf("Expected 2 session rows, got %d", len(summary.Sessions.Rows))
	}

	// Check first row
	row1 := summary.Sessions.Rows[0]
	if row1.AgentID != "agent-1" {
		t.Errorf("Expected AgentID = agent-1, got %s", row1.AgentID)
	}
	if row1.TokensUsed != 1000 {
		t.Errorf("Expected TokensUsed = 1000, got %d", row1.TokensUsed)
	}
	if row1.Usage != 50.0 {
		t.Errorf("Expected Usage = 50.0, got %f", row1.Usage)
	}

	// Check totals
	expectedCost := 0.075
	if diff := summary.Sessions.Totals.CostUSD - expectedCost; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("Expected total cost = %.6f, got %.6f", expectedCost, summary.Sessions.Totals.CostUSD)
	}
	if summary.Sessions.Totals.Tokens != 1500 {
		t.Errorf("Expected total tokens = 1500, got %d", summary.Sessions.Totals.Tokens)
	}
	if summary.Sessions.Totals.Snapshots != 2 {
		t.Errorf("Expected snapshots = 2, got %d", summary.Sessions.Totals.Snapshots)
	}
}

func TestUsageService_Build_MonthlyCosts(t *testing.T) {
	clock := &MockClock{current: time.Date(2025, 10, 22, 19, 4, 0, 0, time.UTC)}
	svc := NewUsageService(clock)

	ts1 := time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC)
	ts2 := time.Date(2025, 10, 16, 10, 0, 0, 0, time.UTC)

	inputs := Inputs{
		Monthly: []domain.SessionMetric{
			{
				Timestamp:        ts1,
				TokensUsed:       2000,
				TokensBudget:     10000,
				EstimatedCostUSD: 0.10,
				ModelName:        "claude-opus",
			},
			{
				Timestamp:        ts2,
				TokensUsed:       1500,
				TokensBudget:     10000,
				EstimatedCostUSD: 0.075,
				ModelName:        "claude-sonnet",
			},
		},
	}

	summary, err := svc.Build(inputs)

	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}

	if !summary.Monthly.HasData {
		t.Fatal("Expected Monthly.HasData = true")
	}

	if len(summary.Monthly.Rows) != 2 {
		t.Fatalf("Expected 2 monthly rows, got %d", len(summary.Monthly.Rows))
	}

	// Check timestamps are formatted
	row1 := summary.Monthly.Rows[0]
	if row1.Timestamp != "2025-10-15 14:00" {
		t.Errorf("Expected timestamp = '2025-10-15 14:00', got '%s'", row1.Timestamp)
	}

	// Check totals
	if summary.Monthly.Totals.CostUSD != 0.175 {
		t.Errorf("Expected total cost = 0.175, got %f", summary.Monthly.Totals.CostUSD)
	}
	if summary.Monthly.Totals.Tokens != 3500 {
		t.Errorf("Expected total tokens = 3500, got %d", summary.Monthly.Totals.Tokens)
	}
}

func TestUsageService_Build_DailyCC(t *testing.T) {
	clock := &MockClock{current: time.Date(2025, 10, 22, 19, 4, 0, 0, time.UTC)}
	svc := NewUsageService(clock)

	inputs := Inputs{
		DailyCC: []api.DailyReport{
			{
				Date:                 "2025-10-22",
				InputTokens:          5000,
				OutputTokens:         2000,
				CacheCreationTokens:  1000,
				CacheReadTokens:      500,
				TotalCost:            0.50,
			},
			{
				Date:                 "2025-10-21",
				InputTokens:          3000,
				OutputTokens:         1500,
				CacheCreationTokens:  800,
				CacheReadTokens:      400,
				TotalCost:            0.35,
			},
			{
				// Duplicate date - should aggregate
				Date:                 "2025-10-22",
				InputTokens:          1000,
				OutputTokens:         500,
				CacheCreationTokens:  200,
				CacheReadTokens:      100,
				TotalCost:            0.10,
			},
		},
	}

	summary, err := svc.Build(inputs)

	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}

	if !summary.DailyCC.HasData {
		t.Fatal("Expected DailyCC.HasData = true")
	}

	// Should have 2 rows (2025-10-22 aggregated, 2025-10-21 separate)
	if len(summary.DailyCC.Rows) != 2 {
		t.Fatalf("Expected 2 daily rows (aggregated), got %d", len(summary.DailyCC.Rows))
	}

	// First row should be most recent date (2025-10-22)
	row1 := summary.DailyCC.Rows[0]
	if row1.Timestamp != "2025-10-22" {
		t.Errorf("Expected first row date = 2025-10-22, got %s", row1.Timestamp)
	}

	// Check aggregation
	if row1.InputTokens != 6000 {
		t.Errorf("Expected aggregated input tokens = 6000, got %d", row1.InputTokens)
	}
	if row1.OutputTokens != 2500 {
		t.Errorf("Expected aggregated output tokens = 2500, got %d", row1.OutputTokens)
	}
	if row1.CacheCreate != 1200 {
		t.Errorf("Expected aggregated cache creation = 1200, got %d", row1.CacheCreate)
	}
	if row1.CacheRead != 600 {
		t.Errorf("Expected aggregated cache read = 600, got %d", row1.CacheRead)
	}
	if row1.Cost != 0.60 {
		t.Errorf("Expected aggregated cost = 0.60, got %f", row1.Cost)
	}

	totalExpected := 6000 + 2500 + 1200 + 600
	if row1.TokensUsed != totalExpected {
		t.Errorf("Expected total tokens = %d, got %d", totalExpected, row1.TokensUsed)
	}

	// Check totals
	if summary.DailyCC.Totals.CostUSD != 0.95 {
		t.Errorf("Expected total cost = 0.95, got %f", summary.DailyCC.Totals.CostUSD)
	}
}

func TestUsageService_Build_MonthlyCC(t *testing.T) {
	clock := &MockClock{current: time.Date(2025, 10, 22, 19, 4, 0, 0, time.UTC)}
	svc := NewUsageService(clock)

	inputs := Inputs{
		MonthlyCC: []api.MonthlyReport{
			{
				Month:             "2025-10",
				Days:              22,
				TotalInputTokens:  50000,
				TotalOutputTokens: 25000,
				TotalCacheTokens:  10000,
				TotalCost:         5.00,
			},
			{
				Month:             "2025-09",
				Days:              30,
				TotalInputTokens:  60000,
				TotalOutputTokens: 30000,
				TotalCacheTokens:  12000,
				TotalCost:         6.00,
			},
		},
	}

	summary, err := svc.Build(inputs)

	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}

	if !summary.MonthlyCC.HasData {
		t.Fatal("Expected MonthlyCC.HasData = true")
	}

	if len(summary.MonthlyCC.Rows) != 2 {
		t.Fatalf("Expected 2 monthly CC rows, got %d", len(summary.MonthlyCC.Rows))
	}

	// Check first row
	row1 := summary.MonthlyCC.Rows[0]
	if row1.Timestamp != "2025-10" {
		t.Errorf("Expected month = 2025-10, got %s", row1.Timestamp)
	}
	if row1.Days != 22 {
		t.Errorf("Expected days = 22, got %d", row1.Days)
	}
	if row1.InputTokens != 50000 {
		t.Errorf("Expected input tokens = 50000, got %d", row1.InputTokens)
	}

	totalExpected := 50000 + 25000 + 10000
	if row1.TokensUsed != totalExpected {
		t.Errorf("Expected total tokens = %d, got %d", totalExpected, row1.TokensUsed)
	}

	// Check totals
	if summary.MonthlyCC.Totals.CostUSD != 11.00 {
		t.Errorf("Expected total cost = 11.00, got %f", summary.MonthlyCC.Totals.CostUSD)
	}
	if summary.MonthlyCC.Totals.Snapshots != 2 {
		t.Errorf("Expected snapshots = 2, got %d", summary.MonthlyCC.Totals.Snapshots)
	}
}

func TestUsageService_Build_GrandTotals(t *testing.T) {
	clock := &MockClock{current: time.Date(2025, 10, 22, 19, 4, 0, 0, time.UTC)}
	svc := NewUsageService(clock)

	inputs := Inputs{
		Sessions: []domain.SessionMetric{
			{TokensUsed: 1000, EstimatedCostUSD: 0.05},
		},
		Monthly: []domain.SessionMetric{
			{TokensUsed: 2000, EstimatedCostUSD: 0.10},
		},
		DailyCC: []api.DailyReport{
			{
				Date:         "2025-10-22",
				InputTokens:  3000,
				OutputTokens: 1000,
				TotalCost:    0.20,
			},
		},
		MonthlyCC: []api.MonthlyReport{
			{
				Month:             "2025-10",
				TotalInputTokens:  4000,
				TotalOutputTokens: 2000,
				TotalCost:         0.30,
			},
		},
	}

	summary, err := svc.Build(inputs)

	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}

	// Check grand totals (MonthlyCC is NOT included to avoid double-counting with DailyCC)
	expectedTotalCost := 0.05 + 0.10 + 0.20  // Sessions + Monthly + DailyCC
	const tolerance = 0.0001
	if diff := summary.TotalsUSD - expectedTotalCost; diff > tolerance || diff < -tolerance {
		t.Errorf("Expected grand total cost = %.6f, got %.6f (diff = %.6f)", expectedTotalCost, summary.TotalsUSD, diff)
	}

	expectedTokens := 1000 + 2000 + (3000 + 1000)  // Sessions + Monthly + DailyCC
	if summary.TokensTotal != expectedTokens {
		t.Errorf("Expected grand total tokens = %d, got %d", expectedTokens, summary.TokensTotal)
	}
}

func TestUsageService_Build_ZeroBudget(t *testing.T) {
	clock := &MockClock{current: time.Date(2025, 10, 22, 19, 4, 0, 0, time.UTC)}
	svc := NewUsageService(clock)

	inputs := Inputs{
		Sessions: []domain.SessionMetric{
			{
				TokensUsed:       1000,
				TokensBudget:     0, // Zero budget should not cause panic
				EstimatedCostUSD: 0.05,
			},
		},
	}

	summary, err := svc.Build(inputs)

	if err != nil {
		t.Fatalf("Build() returned unexpected error: %v", err)
	}

	if len(summary.Sessions.Rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(summary.Sessions.Rows))
	}

	row := summary.Sessions.Rows[0]
	if row.Usage != 0.0 {
		t.Errorf("Expected Usage = 0.0 for zero budget, got %f", row.Usage)
	}
}
