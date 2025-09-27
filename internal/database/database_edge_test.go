package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAgentIDValidation(t *testing.T) {
	dbPath := "./test_validation.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name    string
		agentID string
		wantErr bool
	}{
		{"Valid 8 char ID", "abc12345", false},
		{"Too short", "abc123", true},
		{"Too long", "abc123456", true},
		{"Empty", "", true},
		{"Exactly 8 chars", "12345678", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.GetAgent(tt.agentID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for agent ID %q, got none", tt.agentID)
				}
				// For invalid IDs, we should get validation errors not "not found" errors
				if err != nil && IsAgentNotFoundError(err) {
					t.Errorf("Expected validation error for agent ID %q, got 'not found' error", tt.agentID)
				}
			} else {
				// For valid IDs that don't exist, we should get "not found" not validation errors
				if err != nil && !IsAgentNotFoundError(err) {
					t.Errorf("Unexpected error for valid agent ID %q: %v", tt.agentID, err)
				}
			}
		})
	}
}

func TestConcurrentAgentCreation(t *testing.T) {
	dbPath := "./test_concurrent.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Try to create the same agent concurrently
	agent := &Agent{
		ID:                 "test567a",
		Status:             StatusActive,
		GitWorktreePath:    stringPtr("/test/path"),
		FeatureDescription: stringPtr("Concurrent test"),
		LastActivityAt:     timePtr(time.Now()),
	}

	// Create channels to coordinate goroutines
	done := make(chan error, 2)

	// Launch two goroutines trying to create the same agent
	for i := 0; i < 2; i++ {
		go func() {
			done <- db.CreateAgent(agent)
		}()
	}

	// Collect results
	var successCount, errorCount int
	for i := 0; i < 2; i++ {
		err := <-done
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Exactly one should succeed, one should fail
	if successCount != 1 || errorCount != 1 {
		t.Errorf("Expected 1 success and 1 error, got %d successes and %d errors", successCount, errorCount)
	}
}

func TestTransactionRollback(t *testing.T) {
	dbPath := "./test_rollback.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create an agent first
	agent := &Agent{
		ID:                 "test890a",
		Status:             StatusActive,
		GitWorktreePath:    stringPtr("/test/path"),
		FeatureDescription: stringPtr("Rollback test"),
		LastActivityAt:     timePtr(time.Now()),
	}

	if err := db.CreateAgent(agent); err != nil {
		t.Fatalf("Failed to create test agent: %v", err)
	}

	// Test transaction rollback by creating an invalid scenario
	err = db.WithTransaction(context.Background(), func(tx *Tx) error {
		// Update agent status successfully
		if err := tx.UpdateAgentStatusTx("test890a", StatusWorking, stringPtr("Testing rollback")); err != nil {
			return err
		}

		// Try to create a duplicate agent (this should fail and rollback the status update)
		duplicateAgent := &Agent{
			ID:                 "test890a", // Same ID - should fail
			Status:             StatusIdle,
			GitWorktreePath:    stringPtr("/other/path"),
			FeatureDescription: stringPtr("Duplicate"),
			LastActivityAt:     timePtr(time.Now()),
		}
		return tx.CreateAgentTx(duplicateAgent)
	})

	if err == nil {
		t.Error("Expected transaction to fail due to duplicate agent")
	}

	// Verify that the status update was rolled back
	retrievedAgent, err := db.GetAgent("test890a")
	if err != nil {
		t.Fatalf("Failed to retrieve agent: %v", err)
	}

	if retrievedAgent.Status != StatusActive {
		t.Errorf("Expected status to remain %s after rollback, got %s", StatusActive, retrievedAgent.Status)
	}
}

func TestInvalidStatusValues(t *testing.T) {
	dbPath := "./test_invalid_status.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create an agent first
	agent := &Agent{
		ID:                 "test456b",
		Status:             StatusActive,
		GitWorktreePath:    stringPtr("/test/path"),
		FeatureDescription: stringPtr("Status test"),
		LastActivityAt:     timePtr(time.Now()),
	}

	if err := db.CreateAgent(agent); err != nil {
		t.Fatalf("Failed to create test agent: %v", err)
	}

	// Test updating with invalid status - should succeed at DB level but fail at MCP tool level
	invalidStatuses := []string{"invalid", "unknown", ""}

	for _, status := range invalidStatuses {
		err := db.UpdateAgentStatus("test456b", status, nil)
		if err != nil {
			t.Logf("Database rejected invalid status %q: %v", status, err)
		} else {
			t.Logf("Database accepted status %q (validation happens at MCP level)", status)
		}
	}
}

func TestLargeCommitBatch(t *testing.T) {
	dbPath := "./test_large_batch.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create an agent first
	agent := &Agent{
		ID:                 "testbigb",
		Status:             StatusActive,
		GitWorktreePath:    stringPtr("/test/path"),
		FeatureDescription: stringPtr("Large batch test"),
		LastActivityAt:     timePtr(time.Now()),
	}

	if err := db.CreateAgent(agent); err != nil {
		t.Fatalf("Failed to create test agent: %v", err)
	}

	// Create large batch of commits
	commitCount := 1000
	commitHashes := make([]string, commitCount)
	commitMessages := make([]string, commitCount)

	for i := 0; i < commitCount; i++ {
		commitHashes[i] = generateCommitHash(i)
		commitMessages[i] = generateCommitMessage(i)
	}

	start := time.Now()
	err = db.CreateCommits("testbigb", commitHashes, commitMessages)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to create large commit batch: %v", err)
	}

	t.Logf("Created %d commits in %v", commitCount, duration)

	// Verify commits were created
	commits, err := db.GetCommitsForAgent("testbigb")
	if err != nil {
		t.Fatalf("Failed to retrieve commits: %v", err)
	}

	if len(commits) != commitCount {
		t.Errorf("Expected %d commits, got %d", commitCount, len(commits))
	}
}

// Helper functions
func generateCommitHash(i int) string {
	return fmt.Sprintf("commit%03d", i)
}

func generateCommitMessage(i int) string {
	return fmt.Sprintf("Commit message %d for testing", i)
}

func IsAgentNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if it's one of our custom errors for agent not found
	return strings.Contains(err.Error(), "agent not found")
}
