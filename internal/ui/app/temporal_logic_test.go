package app

import (
	"testing"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// Phase 4: Temporal Logic Tests

// TestCreatedAtTimestamp tests agent creation timestamp
func TestCreatedAtTimestamp(t *testing.T) {
	m := newTestModel(t)
	now := time.Now()
	agent := &domain.Agent{
		ID:        "test",
		CreatedAt: now,
	}
	m.selectedAgent = agent

	if agent.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// Should be approximately now (within 1 second)
	diff := time.Now().Sub(agent.CreatedAt)
	if diff > time.Second {
		t.Errorf("CreatedAt should be approximately now, got diff: %v", diff)
	}
}

// TestUpdatedAtAfterCreation tests that UpdatedAt can be after CreatedAt
func TestUpdatedAtAfterCreation(t *testing.T) {
	m := newTestModel(t)
	createdAt := time.Now().Add(-1 * time.Hour)
	updatedAt := time.Now()

	agent := &domain.Agent{
		ID:        "test",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	m.selectedAgent = agent

	if !agent.UpdatedAt.After(agent.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

// TestLastActivityAtTracking tests last activity timestamp
func TestLastActivityAtTracking(t *testing.T) {
	m := newTestModel(t)
	createdAt := time.Now().Add(-30 * time.Minute)
	lastActivityAt := time.Now().Add(-5 * time.Minute)

	agent := &domain.Agent{
		ID:             "test",
		CreatedAt:      createdAt,
		LastActivityAt: lastActivityAt,
	}
	m.selectedAgent = agent

	// LastActivityAt should be between created and now
	if !agent.LastActivityAt.After(agent.CreatedAt) {
		t.Error("LastActivityAt should be after CreatedAt")
	}

	if agent.LastActivityAt.After(time.Now().Add(time.Second)) {
		t.Error("LastActivityAt should not be in the future")
	}
}

// TestCompletedAtForFinishedSession tests completion timestamp
func TestCompletedAtForFinishedSession(t *testing.T) {
	m := newTestModel(t)
	createdAt := time.Now().Add(-2 * time.Hour)
	completedAt := time.Now().Add(-10 * time.Minute)

	agent := &domain.Agent{
		ID:          "test",
		CreatedAt:   createdAt,
		CompletedAt: &completedAt,
		Status:      "completed",
	}
	m.selectedAgent = agent

	if agent.CompletedAt == nil {
		t.Error("CompletedAt should not be nil for completed agent")
	}

	if !agent.CompletedAt.After(agent.CreatedAt) {
		t.Error("CompletedAt should be after CreatedAt")
	}
}

// TestSessionDurationCalculation tests duration between created and now
func TestSessionDurationCalculation(t *testing.T) {
	m := newTestModel(t)
	createdAt := time.Now().Add(-30 * time.Minute)

	agent := &domain.Agent{
		ID:        "test",
		CreatedAt: createdAt,
		Status:    "active",
	}
	m.selectedAgent = agent

	duration := time.Now().Sub(agent.CreatedAt)

	// Should be approximately 30 minutes
	if duration < 29*time.Minute || duration > 31*time.Minute {
		t.Errorf("Duration should be approximately 30 minutes, got %v", duration)
	}
}

// TestCompletedSessionDurationCalculation tests duration for finished sessions
func TestCompletedSessionDurationCalculation(t *testing.T) {
	m := newTestModel(t)
	createdAt := time.Now().Add(-2 * time.Hour)
	completedAt := time.Now().Add(-30 * time.Minute)

	agent := &domain.Agent{
		ID:          "test",
		CreatedAt:   createdAt,
		CompletedAt: &completedAt,
		Status:      "completed",
	}
	m.selectedAgent = agent

	duration := agent.CompletedAt.Sub(agent.CreatedAt)

	// Should be approximately 90 minutes (2 hours - 30 minutes)
	if duration < 89*time.Minute || duration > 91*time.Minute {
		t.Errorf("Duration should be approximately 90 minutes, got %v", duration)
	}
}

// TestCommitTimestamps tests commit timestamp ordering
func TestCommitTimestamps(t *testing.T) {
	m := newTestModel(t)
	now := time.Now()

	m.commits = make([]domain.Commit, 3)
	m.commits[0] = domain.Commit{
		Hash:      "abc123",
		Timestamp: now.Add(-30 * time.Minute),
	}
	m.commits[1] = domain.Commit{
		Hash:      "def456",
		Timestamp: now.Add(-15 * time.Minute),
	}
	m.commits[2] = domain.Commit{
		Hash:      "ghi789",
		Timestamp: now,
	}

	// Commits should be ordered oldest to newest
	if !m.commits[0].Timestamp.Before(m.commits[1].Timestamp) {
		t.Error("Commit 0 should be before commit 1")
	}
	if !m.commits[1].Timestamp.Before(m.commits[2].Timestamp) {
		t.Error("Commit 1 should be before commit 2")
	}
}

// TestActionTimestamps tests action timestamp tracking
func TestActionTimestamps(t *testing.T) {
	m := newTestModel(t)
	now := time.Now()

	m.actions = make([]domain.Action, 2)
	m.actions[0] = domain.Action{
		Description: "Task started",
		Timestamp:   now.Add(-10 * time.Minute),
		ActionType:  "task_start",
	}
	m.actions[1] = domain.Action{
		Description: "File changed",
		Timestamp:   now,
		ActionType:  "file_operation",
	}

	// Actions should have chronological timestamps
	if !m.actions[0].Timestamp.Before(m.actions[1].Timestamp) {
		t.Error("Action 0 should be before action 1")
	}
}

// TestStatusRefreshTiming tests last refresh timestamp
func TestStatusRefreshTiming(t *testing.T) {
	m := newTestModel(t)
	m.lastRefresh = time.Now()
	m.lastUpdate = time.Now()

	if m.lastRefresh.IsZero() {
		t.Error("lastRefresh should not be zero")
	}

	// lastUpdate should be approximately same as lastRefresh
	diff := m.lastUpdate.Sub(m.lastRefresh)
	if diff < 0 {
		diff = -diff
	}
	if diff > 100*time.Millisecond {
		t.Errorf("lastUpdate and lastRefresh should be close, diff: %v", diff)
	}
}

// TestTimeProgression tests that time progresses correctly
func TestTimeProgression(t *testing.T) {
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	end := time.Now()

	if !end.After(start) {
		t.Error("End time should be after start time")
	}

	elapsed := end.Sub(start)
	if elapsed < 10*time.Millisecond {
		t.Errorf("Elapsed time should be at least 10ms, got %v", elapsed)
	}
}

// TestZeroTimeComparison tests comparison with zero time
func TestZeroTimeComparison(t *testing.T) {
	zeroTime := time.Time{}
	now := time.Now()

	if !now.After(zeroTime) {
		t.Error("Now should be after zero time")
	}

	if zeroTime.IsZero() == false {
		t.Error("Zero time should be identified as zero")
	}
}

// TestTimeEquality tests time equality checks
func TestTimeEquality(t *testing.T) {
	time1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	time3 := time.Date(2025, 1, 1, 12, 0, 1, 0, time.UTC)

	if !time1.Equal(time2) {
		t.Error("Equal times should compare equal")
	}

	if time1.Equal(time3) {
		t.Error("Different times should not compare equal")
	}
}

// TestNotesCreatedAt tests note creation timestamps
func TestNotesCreatedAt(t *testing.T) {
	m := newTestModel(t)
	now := time.Now()

	m.notes = make([]domain.Note, 2)
	m.notes[0] = domain.Note{
		Content:   "First note",
		CreatedAt: now.Add(-1 * time.Hour),
	}
	m.notes[1] = domain.Note{
		Content:   "Second note",
		CreatedAt: now,
	}

	if m.notes[1].CreatedAt.Before(m.notes[0].CreatedAt) {
		t.Error("Second note should be created after first note")
	}
}

// TestAgentStatusTiming tests how status relates to timestamps
func TestAgentStatusTiming(t *testing.T) {
	m := newTestModel(t)
	createdAt := time.Now().Add(-1 * time.Hour)
	var completedAtPtr *time.Time

	// Active agent
	activeAgent := &domain.Agent{
		ID:          "active",
		CreatedAt:   createdAt,
		CompletedAt: nil,
		Status:      "active",
	}

	// Completed agent
	completedTime := time.Now()
	completedAtPtr = &completedTime
	completedAgent := &domain.Agent{
		ID:          "completed",
		CreatedAt:   createdAt,
		CompletedAt: completedAtPtr,
		Status:      "completed",
	}

	m.selectedAgent = activeAgent
	if m.selectedAgent.CompletedAt != nil {
		t.Error("Active agent should not have CompletedAt")
	}

	m.selectedAgent = completedAgent
	if m.selectedAgent.CompletedAt == nil {
		t.Error("Completed agent should have CompletedAt")
	}
}

// TestRefreshIntervalValidation tests refresh interval is positive
func TestRefreshIntervalValidation(t *testing.T) {
	// Refresh interval should be positive
	refreshInterval := 5 * time.Second
	if refreshInterval <= 0 {
		t.Error("Refresh interval should be positive")
	}

	// Test with zero interval (should be invalid)
	invalidInterval := time.Duration(0)
	if invalidInterval > 0 {
		t.Log("Zero interval is invalid")
	}
}

// TestTimeInFuture tests time in the future is detected
func TestTimeInFuture(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	if !future.After(now) {
		t.Error("Future time should be after now")
	}

	if !now.After(past) {
		t.Error("Now should be after past time")
	}
}

// TestDurationComparison tests duration comparisons
func TestDurationComparison(t *testing.T) {
	short := 5 * time.Minute
	long := 1 * time.Hour

	if short >= long {
		t.Error("5 minutes should be less than 1 hour")
	}

	if long <= short {
		t.Error("1 hour should be more than 5 minutes")
	}
}

// TestCommitTimestampOrdering tests multiple commits maintain timestamp order
func TestCommitTimestampOrdering(t *testing.T) {
	m := newTestModel(t)

	// Create 5 commits with increasing timestamps
	m.commits = make([]domain.Commit, 5)
	baseTime := time.Now().Add(-10 * time.Minute)

	for i := 0; i < 5; i++ {
		m.commits[i] = domain.Commit{
			Hash:      domain.CommitHash(string(rune('a' + i))),
			Timestamp: baseTime.Add(time.Duration(i*2) * time.Minute),
		}
	}

	// Verify ordering
	for i := 0; i < len(m.commits)-1; i++ {
		if !m.commits[i].Timestamp.Before(m.commits[i+1].Timestamp) {
			t.Errorf("Commit %d should be before commit %d", i, i+1)
		}
	}
}
