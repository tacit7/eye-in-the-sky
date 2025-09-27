package database

import (
	"os"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	// Create temporary database
	dbPath := "./test_agents.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test health check
	if err := db.Health(); err != nil {
		t.Fatalf("Database health check failed: %v", err)
	}

	// Test agent operations
	testAgent := &Agent{
		ID:                 "test123",
		Status:             StatusActive,
		GitWorktreePath:    stringPtr("/path/to/worktree"),
		FeatureDescription: stringPtr("Test feature"),
		CurrentTask:        stringPtr("Testing database"),
		LastActivityAt:     timePtr(time.Now()),
	}

	// Create agent
	if err := db.CreateAgent(testAgent); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Get agent
	retrieved, err := db.GetAgent("test123")
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if retrieved.ID != "test123" || retrieved.Status != StatusActive {
		t.Errorf("Agent data mismatch: %+v", retrieved)
	}

	// Update agent status
	if err := db.UpdateAgentStatus("test123", StatusWorking, stringPtr("New task")); err != nil {
		t.Fatalf("Failed to update agent status: %v", err)
	}

	// Test action operations
	testAction := &Action{
		AgentID:     "test123",
		ActionType:  ActionTaskStart,
		Description: "Started testing",
		Details:     stringPtr(`{"test": true}`),
	}

	if err := db.CreateAction(testAction); err != nil {
		t.Fatalf("Failed to create action: %v", err)
	}

	// Get actions
	actions, err := db.GetActionsForAgent("test123", 10)
	if err != nil {
		t.Fatalf("Failed to get actions: %v", err)
	}

	if len(actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(actions))
	}

	// Test commit operations
	commitHashes := []string{"abc123", "def456"}
	commitMessages := []string{"First commit", "Second commit"}

	if err := db.CreateCommits("test123", commitHashes, commitMessages); err != nil {
		t.Fatalf("Failed to create commits: %v", err)
	}

	// Get commits
	commits, err := db.GetCommitsForAgent("test123")
	if err != nil {
		t.Fatalf("Failed to get commits: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(commits))
	}

	// Test session end
	if err := db.EndAgentSession("test123", "Test completed", StatusCompleted); err != nil {
		t.Fatalf("Failed to end session: %v", err)
	}

	// Verify agent status updated
	final, err := db.GetAgent("test123")
	if err != nil {
		t.Fatalf("Failed to get final agent: %v", err)
	}

	if final.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, final.Status)
	}

	// Test stats
	stats, err := db.GetAgentStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats[StatusCompleted] != 1 {
		t.Errorf("Expected 1 completed agent, got %d", stats[StatusCompleted])
	}
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}