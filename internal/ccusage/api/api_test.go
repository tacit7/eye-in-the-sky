package api

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
)

func setupTestDB(t *testing.T) *db.CCUsageDB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	testDB, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Insert test data
	now := time.Now()
	entries := []db.UsageEntryRow{
		{
			SessionID:           "session1",
			Timestamp:           now.AddDate(0, 0, -5).Format(time.RFC3339),
			Project:             "project1",
			Model:               "claude-3-sonnet-20240229",
			InputTokens:         100,
			OutputTokens:        200,
			CacheCreationTokens: 10,
			CacheReadTokens:     5,
			TotalCost:           0.05,
			MessageID:           "msg1",
			RequestID:           "req1",
			UniqueHash:          "hash1",
		},
		{
			SessionID:           "session2",
			Timestamp:           now.AddDate(0, 0, -3).Format(time.RFC3339),
			Project:             "project1",
			Model:               "claude-3-opus-20240229",
			InputTokens:         150,
			OutputTokens:        250,
			CacheCreationTokens: 15,
			CacheReadTokens:     8,
			TotalCost:           0.08,
			MessageID:           "msg2",
			RequestID:           "req2",
			UniqueHash:          "hash2",
		},
		{
			SessionID:           "session3",
			Timestamp:           now.Format(time.RFC3339),
			Project:             "project2",
			Model:               "claude-3-sonnet-20240229",
			InputTokens:         200,
			OutputTokens:        300,
			CacheCreationTokens: 20,
			CacheReadTokens:     10,
			TotalCost:           0.10,
			MessageID:           "msg3",
			RequestID:           "req3",
			UniqueHash:          "hash3",
		},
	}

	if err := testDB.BatchInsertUsageEntries(entries); err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	return testDB
}

func TestGetDailyUsageReport(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	reports, err := GetDailyUsageReport(testDB, 7)
	if err != nil {
		t.Fatalf("Failed to get daily usage report: %v", err)
	}

	if len(reports) == 0 {
		t.Errorf("Expected at least one daily report, got %d", len(reports))
	}

	for _, report := range reports {
		if report.Date == "" {
			t.Error("Expected date to be non-empty")
		}
		if report.Project == "" {
			t.Error("Expected project to be non-empty")
		}
		if report.TotalCost <= 0 {
			t.Error("Expected positive cost")
		}
	}
}

func TestGetSessionUsageReport(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	reports, err := GetSessionUsageReport(testDB, 10)
	if err != nil {
		t.Fatalf("Failed to get session usage report: %v", err)
	}

	if len(reports) != 3 {
		t.Errorf("Expected 3 session reports, got %d", len(reports))
	}

	for _, report := range reports {
		if report.SessionID == "" {
			t.Error("Expected session ID to be non-empty")
		}
		if report.StartTime == "" {
			t.Error("Expected start time to be non-empty")
		}
		if report.Duration == "" {
			t.Error("Expected duration to be non-empty")
		}
	}
}

func TestGetMonthlyUsageReport(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	now := time.Now()
	report, err := GetMonthlyUsageReport(testDB, now.Year(), int(now.Month()))
	if err != nil {
		t.Fatalf("Failed to get monthly usage report: %v", err)
	}

	if report == nil {
		t.Error("Expected non-nil report")
		return
	}

	if report.TotalInputTokens == 0 {
		t.Error("Expected non-zero input tokens")
	}

	if report.TotalCost == 0 {
		t.Error("Expected non-zero cost")
	}

	if len(report.ProjectBreakdown) == 0 {
		t.Error("Expected project breakdown")
	}
}

func TestGetTotalCostSummary(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	summary, err := GetTotalCostSummary(testDB, 30)
	if err != nil {
		t.Fatalf("Failed to get cost summary: %v", err)
	}

	if summary == nil {
		t.Error("Expected non-nil summary")
		return
	}

	if summary.TotalCost == 0 {
		t.Error("Expected non-zero total cost")
	}

	if summary.SessionCount != 3 {
		t.Errorf("Expected 3 sessions, got %d", summary.SessionCount)
	}

	if summary.ProjectCount == 0 {
		t.Error("Expected at least 1 project")
	}

	if summary.AverageDailyCost <= 0 {
		t.Error("Expected positive average daily cost")
	}
}

func TestGetActiveBlockReport(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	report, err := GetActiveBlockReport(testDB)
	if err != nil {
		t.Fatalf("Failed to get active block report: %v", err)
	}

	if report == nil {
		t.Error("Expected non-nil report")
		return
	}

	if report.StartTime == "" {
		t.Error("Expected start time to be non-empty")
	}

	if report.EndTime == "" {
		t.Error("Expected end time to be non-empty")
	}

	if report.TimeRemaining == "" {
		t.Error("Expected time remaining to be non-empty")
	}
}

func TestGetDailyUsageReportWithDefaultDays(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	reports, err := GetDailyUsageReport(testDB, 0)
	if err != nil {
		t.Fatalf("Failed to get daily usage report with default days: %v", err)
	}

	if len(reports) == 0 {
		t.Errorf("Expected reports with default days")
	}
}

func TestGetSessionUsageReportWithDefaultLimit(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	reports, err := GetSessionUsageReport(testDB, 0)
	if err != nil {
		t.Fatalf("Failed to get session usage report with default limit: %v", err)
	}

	if len(reports) != 3 {
		t.Errorf("Expected 3 session reports with default limit, got %d", len(reports))
	}
}

func TestGetMonthlyUsageReportWithDefaultMonth(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	report, err := GetMonthlyUsageReport(testDB, 0, 0)
	if err != nil {
		t.Fatalf("Failed to get monthly usage report with default month: %v", err)
	}

	if report == nil {
		t.Error("Expected non-nil report with default month")
	}
}
