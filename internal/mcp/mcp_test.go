package mcp

import (
	"os"
	"testing"

	"github.com/tacit7/eye-in-the-sky/internal/database"
)

func TestMCPTools(t *testing.T) {
	// Create temporary database
	dbPath := "./test_mcp.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create tools instance
	tools := NewTools(db)

	// Test register_agent
	registerArgs := RegisterAgentArgs{
		AgentID:      "test456a",
		Description:  "Testing MCP tools",
		WorktreePath: stringPtr("/test/path"),
	}

	registerResult, err := tools.RegisterAgent(registerArgs)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	if !registerResult.Success {
		t.Errorf("Expected registration success, got: %s", registerResult.Message)
	}

	// Test duplicate registration (should fail)
	duplicateResult, err := tools.RegisterAgent(registerArgs)
	if err != nil {
		t.Fatalf("Unexpected error on duplicate registration: %v", err)
	}

	if duplicateResult.Success {
		t.Error("Expected duplicate registration to fail")
	}

	// Test update_status
	updateArgs := UpdateStatusArgs{
		AgentID:     "test456a",
		Status:      "working",
		CurrentTask: stringPtr("Running tests"),
	}

	updateResult, err := tools.UpdateStatus(updateArgs)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	if !updateResult.Success {
		t.Errorf("Expected status update success, got: %s", updateResult.Message)
	}

	// Test log_action
	actionArgs := LogActionArgs{
		AgentID:     "test456a",
		ActionType:  "task_start",
		Description: "Started testing MCP tools",
		Details:     stringPtr(`{"test": true}`),
	}

	actionResult, err := tools.LogAction(actionArgs)
	if err != nil {
		t.Fatalf("Failed to log action: %v", err)
	}

	if !actionResult.Success {
		t.Errorf("Expected action log success, got: %s", actionResult.Message)
	}

	// Test log_commits
	commitsArgs := LogCommitsArgs{
		AgentID:        "test456a",
		CommitHashes:   []string{"abc123", "def456"},
		CommitMessages: []string{"Test commit 1", "Test commit 2"},
	}

	commitsResult, err := tools.LogCommits(commitsArgs)
	if err != nil {
		t.Fatalf("Failed to log commits: %v", err)
	}

	if !commitsResult.Success {
		t.Errorf("Expected commits log success, got: %s", commitsResult.Message)
	}

	// Test end_session
	endArgs := EndSessionArgs{
		AgentID:     "test456a",
		Summary:     stringPtr("Testing completed successfully"),
		FinalStatus: stringPtr("completed"),
	}

	endResult, err := tools.EndSession(endArgs)
	if err != nil {
		t.Fatalf("Failed to end session: %v", err)
	}

	if !endResult.Success {
		t.Errorf("Expected session end success, got: %s", endResult.Message)
	}

	// Verify agent is completed
	agent, err := db.GetAgent("test456a")
	if err != nil {
		t.Fatalf("Failed to get agent after session end: %v", err)
	}

	if agent.Status != "completed" {
		t.Errorf("Expected agent status 'completed', got: %s", agent.Status)
	}
}

func TestMCPServer(t *testing.T) {
	// Create temporary database
	dbPath := "./test_server.db"
	defer os.Remove(dbPath)

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create server
	server := NewServer(db)

	// Test tool list
	tools := server.GetToolList()
	if len(tools) != 5 {
		t.Errorf("Expected 5 tools, got %d", len(tools))
	}

	// Test tool names
	expectedTools := map[string]bool{
		"register_agent": false,
		"update_status":  false,
		"log_action":     false,
		"log_commits":    false,
		"end_session":    false,
	}

	for _, tool := range tools {
		if _, exists := expectedTools[tool.Name]; exists {
			expectedTools[tool.Name] = true
		} else {
			t.Errorf("Unexpected tool: %s", tool.Name)
		}
	}

	for tool, found := range expectedTools {
		if !found {
			t.Errorf("Missing tool: %s", tool)
		}
	}

	// Test HandleTool with register_agent
	registerJSON := `{"agent_id":"test789b","description":"Server test","worktree_path":"/test"}`
	result, err := server.HandleTool("register_agent", []byte(registerJSON))
	if err != nil {
		t.Fatalf("Failed to handle register_agent tool: %v", err)
	}

	registerResult, ok := result.(RegisterAgentResult)
	if !ok {
		t.Fatalf("Expected RegisterAgentResult, got %T", result)
	}

	if !registerResult.Success {
		t.Errorf("Expected tool success, got: %s", registerResult.Message)
	}

	// Test unknown tool
	_, err = server.HandleTool("unknown_tool", []byte("{}"))
	if err == nil {
		t.Error("Expected error for unknown tool")
	}
}
