package mcp

import (
	"os"
	"testing"

	"github.com/tacit7/eye-in-the-sky/internal/database"
)

func TestMCPToolsErrorHandling(t *testing.T) {
	dbPath := "./test_mcp_errors.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	tools := NewTools(db)

	t.Run("RegisterAgent with invalid ID", func(t *testing.T) {
		args := RegisterAgentArgs{
			AgentID:      stringPtr("invalid1"), // Invalid character
			Description:  "Test agent",
			WorktreePath: stringPtr("/test/path"),
		}

		result, err := tools.RegisterAgent(args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Success {
			t.Error("Expected registration to fail with invalid agent ID")
		}
	})

	t.Run("UpdateStatus for non-existent agent", func(t *testing.T) {
		args := UpdateStatusArgs{
			AgentID:     "notfound",
			Status:      database.StatusWorking,
			CurrentTask: stringPtr("Testing"),
		}

		result, err := tools.UpdateStatus(args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Success {
			t.Error("Expected status update to fail for non-existent agent")
		}
	})

	t.Run("UpdateStatus with invalid status", func(t *testing.T) {
		// First create a valid agent
		registerArgs := RegisterAgentArgs{
			AgentID:      stringPtr("1e51a9e1"),
			Description:  "Test agent",
			WorktreePath: stringPtr("/test/path"),
		}

		_, err := tools.RegisterAgent(registerArgs)
		if err != nil {
			t.Fatalf("Failed to register test agent: %v", err)
		}

		// Now try to update with invalid status
		updateArgs := UpdateStatusArgs{
			AgentID:     "1e51a9e1",
			Status:      "invalid_status",
			CurrentTask: stringPtr("Testing"),
		}

		result, err := tools.UpdateStatus(updateArgs)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Success {
			t.Error("Expected status update to fail with invalid status")
		}
	})

	t.Run("LogAction with invalid action type", func(t *testing.T) {
		args := LogActionArgs{
			AgentID:     "1e51a9e1",
			ActionType:  "invalid_action",
			Description: "Testing invalid action",
			Details:     stringPtr(`{"test": true}`),
		}

		result, err := tools.LogAction(args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Success {
			t.Error("Expected action log to fail with invalid action type")
		}
	})

	t.Run("LogCommits with empty hashes", func(t *testing.T) {
		args := LogCommitsArgs{
			AgentID:        "1e51a9e1",
			CommitHashes:   []string{},
			CommitMessages: []string{},
		}

		result, err := tools.LogCommits(args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Success {
			t.Error("Expected commit log to fail with empty hashes")
		}
	})

	t.Run("EndSession with invalid final status", func(t *testing.T) {
		args := EndSessionArgs{
			AgentID:     "1e51a9e1",
			Summary:     stringPtr("Test completed"),
			FinalStatus: stringPtr("invalid_final"),
		}

		result, err := tools.EndSession(args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Success {
			t.Error("Expected session end to fail with invalid final status")
		}
	})
}

func TestMCPWorkflowIntegration(t *testing.T) {
	dbPath := "./test_mcp_workflow.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	tools := NewTools(db)
	agentID := "abc12345"

	// Step 1: Register agent
	t.Run("1. Register Agent", func(t *testing.T) {
		args := RegisterAgentArgs{
			AgentID:      stringPtr(agentID),
			Description:  "Full workflow test agent",
			WorktreePath: stringPtr("/test/workflow"),
		}

		result, err := tools.RegisterAgent(args)
		if err != nil {
			t.Fatalf("Failed to register agent: %v", err)
		}

		if !result.Success {
			t.Errorf("Agent registration failed: %s", result.Message)
		}
	})

	// Step 2: Update status to working
	t.Run("2. Update Status to Working", func(t *testing.T) {
		args := UpdateStatusArgs{
			AgentID:     agentID,
			Status:      database.StatusWorking,
			CurrentTask: stringPtr("Starting workflow test"),
		}

		result, err := tools.UpdateStatus(args)
		if err != nil {
			t.Fatalf("Failed to update status: %v", err)
		}

		if !result.Success {
			t.Errorf("Status update failed: %s", result.Message)
		}
	})

	// Step 3: Log some actions
	t.Run("3. Log Actions", func(t *testing.T) {
		actions := []LogActionArgs{
			{
				AgentID:     agentID,
				ActionType:  database.ActionTaskStart,
				Description: "Started workflow test",
				Details:     stringPtr(`{"step": 1}`),
			},
			{
				AgentID:     agentID,
				ActionType:  database.ActionFileOperation,
				Description: "Created test file",
				Details:     stringPtr(`{"file": "test.go", "operation": "create"}`),
			},
		}

		for i, action := range actions {
			result, err := tools.LogAction(action)
			if err != nil {
				t.Fatalf("Failed to log action %d: %v", i+1, err)
			}

			if !result.Success {
				t.Errorf("Action %d log failed: %s", i+1, result.Message)
			}
		}
	})

	// Step 4: Log commits
	t.Run("4. Log Commits", func(t *testing.T) {
		args := LogCommitsArgs{
			AgentID:        agentID,
			CommitHashes:   []string{"abc123workflow", "def456workflow"},
			CommitMessages: []string{"Initial workflow commit", "Workflow test improvements"},
		}

		result, err := tools.LogCommits(args)
		if err != nil {
			t.Fatalf("Failed to log commits: %v", err)
		}

		if !result.Success {
			t.Errorf("Commit log failed: %s", result.Message)
		}
	})

	// Step 5: End session
	t.Run("5. End Session", func(t *testing.T) {
		args := EndSessionArgs{
			AgentID:     agentID,
			Summary:     stringPtr("Workflow test completed successfully"),
			FinalStatus: stringPtr(database.StatusCompleted),
		}

		result, err := tools.EndSession(args)
		if err != nil {
			t.Fatalf("Failed to end session: %v", err)
		}

		if !result.Success {
			t.Errorf("Session end failed: %s", result.Message)
		}
	})

	// Step 6: Verify final state
	t.Run("6. Verify Final State", func(t *testing.T) {
		agent, err := db.GetAgent(agentID)
		if err != nil {
			t.Fatalf("Failed to get final agent state: %v", err)
		}

		if agent.Status != database.StatusCompleted {
			t.Errorf("Expected final status %s, got %s", database.StatusCompleted, agent.Status)
		}

		// Verify actions were logged
		actions, err := db.GetActionsForAgent(agentID, 10)
		if err != nil {
			t.Fatalf("Failed to get actions: %v", err)
		}

		if len(actions) < 2 {
			t.Errorf("Expected at least 2 actions, got %d", len(actions))
		}

		// Verify commits were logged
		commits, err := db.GetCommitsForAgent(agentID)
		if err != nil {
			t.Fatalf("Failed to get commits: %v", err)
		}

		if len(commits) != 2 {
			t.Errorf("Expected 2 commits, got %d", len(commits))
		}
	})
}

func TestMCPServerJSONHandling(t *testing.T) {
	dbPath := "./test_mcp_json.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	server := NewServer(db)

	t.Run("Valid JSON for register_agent", func(t *testing.T) {
		jsonData := `{
			"agent_id": "12345678",
			"description": "JSON handling test",
			"worktree_path": "/test/json"
		}`

		result, err := server.HandleTool("register_agent", []byte(jsonData))
		if err != nil {
			t.Fatalf("Failed to handle tool: %v", err)
		}

		registerResult, ok := result.(RegisterAgentResult)
		if !ok {
			t.Fatalf("Expected RegisterAgentResult, got %T", result)
		}

		if !registerResult.Success {
			t.Errorf("Expected success, got: %s", registerResult.Message)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := `{"agent_id": "12345678", invalid json`

		_, err := server.HandleTool("register_agent", []byte(invalidJSON))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("Auto-generated agent ID", func(t *testing.T) {
		autoGenJSON := `{"description": "Auto-generated ID test"}`

		result, err := server.HandleTool("register_agent", []byte(autoGenJSON))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		registerResult, ok := result.(RegisterAgentResult)
		if !ok {
			t.Fatalf("Expected RegisterAgentResult, got %T", result)
		}

		if !registerResult.Success {
			t.Errorf("Expected success with auto-generated ID, got: %s", registerResult.Message)
		}
	})

	t.Run("Unknown tool", func(t *testing.T) {
		_, err := server.HandleTool("unknown_tool", []byte(`{}`))
		if err == nil {
			t.Error("Expected error for unknown tool")
		}
	})
}
