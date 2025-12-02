package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

func TestRenderAgentInfo(t *testing.T) {
	styles := defaultTestOverviewStyles()

	agent := domain.Agent{
		ID:              "test-agent-123",
		Status:          "active",
		FeatureDesc:     "Test feature",
		ProjectName:     "test-project",
		CurrentTask:     "Implement feature",
		Source:          "worktree",
		GitWorktreePath: "/path/to/worktree",
		SessionID:       "session-123",
		ParentSessionID: "parent-session-456",
		ParentAgentID:   "parent-agent-789",
		WindowID:        "window-001",
	}

	output := renderAgentInfo(agent, styles)

	// Verify section header
	if !strings.Contains(output, "Agent Information") {
		t.Error("Missing 'Agent Information' header")
	}

	// Verify all fields are present
	expectedFields := []string{
		"Agent ID:",
		"Status:",
		"Description:",
		"Project:",
		"Current Task:",
		"Source:",
		"Worktree:",
		"Session ID:",
		"Parent Session:",
		"Parent Agent:",
		"Window ID:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Missing field: %s", field)
		}
	}

	// Verify values are present
	if !strings.Contains(output, "test-agent-123") {
		t.Error("Agent ID value missing")
	}
	if !strings.Contains(output, "Test feature") {
		t.Error("Feature description missing")
	}
}

func TestRenderAgentInfoOptionalFields(t *testing.T) {
	styles := defaultTestOverviewStyles()

	agent := domain.Agent{
		ID:     "test-agent",
		Status: "active",
		Source: "worktree",
	}

	output := renderAgentInfo(agent, styles)

	// Verify optional fields are not included
	if strings.Contains(output, "Description:") {
		t.Error("Optional description should not appear when empty")
	}
	if strings.Contains(output, "Project:") {
		t.Error("Optional project should not appear when empty")
	}
}

func TestRenderTiming(t *testing.T) {
	styles := defaultTestOverviewStyles()

	now := time.Now()
	agent := domain.Agent{
		CreatedAt:      now.Add(-1 * time.Hour),
		UpdatedAt:      now.Add(-10 * time.Minute),
		LastActivityAt: now.Add(-2 * time.Minute),
		Status:         "active",
	}

	output := renderTiming(agent, styles)

	// Verify section header
	if !strings.Contains(output, "Timing") {
		t.Error("Missing 'Timing' header")
	}

	// Verify timing fields
	if !strings.Contains(output, "Created:") {
		t.Error("Missing Created field")
	}
	if !strings.Contains(output, "Updated:") {
		t.Error("Missing Updated field")
	}
	if !strings.Contains(output, "Last Activity:") {
		t.Error("Missing Last Activity field")
	}
	if !strings.Contains(output, "Session Duration:") {
		t.Error("Missing Session Duration field")
	}
}

func TestRenderTimingCompleted(t *testing.T) {
	styles := defaultTestOverviewStyles()

	completedTime := time.Now()
	agent := domain.Agent{
		CreatedAt:   completedTime.Add(-2 * time.Hour),
		UpdatedAt:   completedTime,
		CompletedAt: &completedTime,
		Status:      "completed",
	}

	output := renderTiming(agent, styles)

	// Completed agents should show total duration, not session duration
	if !strings.Contains(output, "Total Duration:") {
		t.Error("Completed agents should show 'Total Duration'")
	}
	if strings.Contains(output, "Session Duration:") {
		t.Error("Completed agents should not show 'Session Duration'")
	}
}

func TestRenderCommitsSection(t *testing.T) {
	styles := defaultTestOverviewStyles()
	commits := []domain.Commit{
		{
			Hash:      "abc123def456",
			Message:   "First commit",
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
		{
			Hash:      "xyz789uvw012",
			Message:   "Second commit",
			Timestamp: time.Now().Add(-30 * time.Minute),
		},
	}

	output := renderCommitsSection(commits, styles)

	// Verify section header
	if !strings.Contains(output, "Recent Commits") {
		t.Error("Missing 'Recent Commits' header")
	}

	// Verify commits are shown
	if !strings.Contains(output, "abc123de") {
		t.Error("First commit hash not shown")
	}
	if !strings.Contains(output, "First commit") {
		t.Error("First commit message not shown")
	}
}

func TestRenderCommitsSectionEmpty(t *testing.T) {
	styles := defaultTestOverviewStyles()
	commits := []domain.Commit{}

	output := renderCommitsSection(commits, styles)

	if strings.TrimSpace(output) != "" {
		t.Error("Empty commits should return empty string")
	}
}

func TestRenderNotesSection(t *testing.T) {
	styles := defaultTestOverviewStyles()
	notes := []domain.Note{
		{
			Content:   "This is a test note",
			CreatedAt: time.Now(),
		},
	}

	output := renderNotesSection(notes, styles)

	// Verify section header
	if !strings.Contains(output, "Notes Summary") {
		t.Error("Missing 'Notes Summary' header")
	}

	// Verify note content
	if !strings.Contains(output, "This is a test note") {
		t.Error("Note content not shown")
	}
}

func TestRenderNotesSectionEmpty(t *testing.T) {
	styles := defaultTestOverviewStyles()
	notes := []domain.Note{}

	output := renderNotesSection(notes, styles)

	if strings.TrimSpace(output) != "" {
		t.Error("Empty notes should return empty string")
	}
}

func TestRenderTasksSection(t *testing.T) {
	styles := defaultTestOverviewStyles()
	taskCounts := map[string]int{
		"todo":       1,
		"inProgress": 1,
		"completed":  1,
		"archived":   1,
	}

	output := renderTasksSection(taskCounts, styles)

	// Verify section header
	if !strings.Contains(output, "Tasks Summary") {
		t.Error("Missing 'Tasks Summary' header")
	}

	// Verify task counts appear
	if !strings.Contains(output, "Todo:") {
		t.Error("Missing Todo count")
	}
}

func TestRenderTasksSectionEmpty(t *testing.T) {
	styles := defaultTestOverviewStyles()
	taskCounts := map[string]int{}

	output := renderTasksSection(taskCounts, styles)

	if strings.TrimSpace(output) != "" {
		t.Error("Empty tasks should return empty string")
	}
}

func TestRenderActionsSection(t *testing.T) {
	styles := defaultTestOverviewStyles()
	actions := []domain.Action{
		{
			Description: "Started session",
			ActionType:  "task_start",
			Timestamp:   time.Now().Add(-5 * time.Minute),
		},
		{
			Description: "Committed changes",
			ActionType:  "git_commit",
			Timestamp:   time.Now().Add(-3 * time.Minute),
		},
	}

	output := renderActionsSection(actions, styles)

	// Verify section header
	if !strings.Contains(output, "Recent Actions") {
		t.Error("Missing 'Recent Actions' header")
	}

	// Verify actions are shown
	if !strings.Contains(output, "Started session") {
		t.Error("First action not shown")
	}
}

func TestRenderActionsSectionEmpty(t *testing.T) {
	styles := defaultTestOverviewStyles()
	actions := []domain.Action{}

	output := renderActionsSection(actions, styles)

	if strings.TrimSpace(output) != "" {
		t.Error("Empty actions should return empty string")
	}
}

func TestRenderDuration(t *testing.T) {
	styles := defaultTestOverviewStyles()

	// Active agent should show session duration
	activeAgent := domain.Agent{
		CreatedAt: time.Now().Add(-2 * time.Hour),
		Status:    "active",
	}

	output := renderDuration(activeAgent, styles)
	if !strings.Contains(output, "Session Duration:") {
		t.Error("Active agent should show session duration")
	}

	// Completed agent should show total duration
	completedTime := time.Now()
	completedAgent := domain.Agent{
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		CompletedAt: &completedTime,
		Status:      "completed",
	}

	output = renderDuration(completedAgent, styles)
	if !strings.Contains(output, "Total Duration:") {
		t.Error("Completed agent should show total duration")
	}
}

func TestGetStatusStyleFromOverview(t *testing.T) {
	styles := defaultTestOverviewStyles()

	tests := []struct {
		status   string
		expected lipgloss.Style
	}{
		{"active", styles.Success},
		{"working", styles.Warning},
		{"idle", styles.Subtle},
		{"failed", styles.Error},
		{"completed", styles.Primary},
		{"unknown", styles.Subtle},
	}

	for _, test := range tests {
		result := getStatusStyleFromOverview(test.status, styles)
		// Compare the string representation since Style objects don't have equality
		if result.String() != test.expected.String() {
			t.Errorf("Status %s returned wrong style", test.status)
		}
	}
}

// Benchmark tests to ensure performance
func BenchmarkRenderAgentInfo(b *testing.B) {
	styles := defaultTestOverviewStyles()
	agent := domain.Agent{
		ID:              "test-agent-123",
		Status:          "active",
		FeatureDesc:     "Test feature",
		ProjectName:     "test-project",
		CurrentTask:     "Implement feature",
		Source:          "worktree",
		GitWorktreePath: "/path/to/worktree",
		SessionID:       "session-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderAgentInfo(agent, styles)
	}
}

func BenchmarkRenderCommitsSection(b *testing.B) {
	styles := defaultTestOverviewStyles()
	commits := []domain.Commit{
		{Hash: "abc123def456", Message: "Commit 1", Timestamp: time.Now().Add(-1 * time.Hour)},
		{Hash: "def456ghi789", Message: "Commit 2", Timestamp: time.Now().Add(-30 * time.Minute)},
		{Hash: "ghi789jkl012", Message: "Commit 3", Timestamp: time.Now().Add(-15 * time.Minute)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderCommitsSection(commits, styles)
	}
}

func BenchmarkRenderTiming(b *testing.B) {
	styles := defaultTestOverviewStyles()
	agent := domain.Agent{
		CreatedAt:      time.Now().Add(-2 * time.Hour),
		UpdatedAt:      time.Now().Add(-10 * time.Minute),
		LastActivityAt: time.Now().Add(-2 * time.Minute),
		Status:         "active",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderTiming(agent, styles)
	}
}

// Note: Test helpers are in test_helpers.go
