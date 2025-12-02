package main

import (
	"fmt"
	"log"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/mcp"
)

func main() {
	fmt.Println("=== Testing Eye in the Sky Session ===\n")

	// Connect to production database
	db, err := database.New("./data/agents.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tools := mcp.NewTools(db)

	// Start a new session
	projectName := "Eye in the Sky"
	worktreePath := "/Users/urielmaldonado/projects/eye-in-the-sky"

	fmt.Println("1. Starting session...")
	result, err := tools.StartSession(mcp.StartSessionArgs{
		Description:  "Testing Eye in the Sky session tracking",
		ProjectName:  &projectName,
		WorktreePath: &worktreePath,
	})
	if err != nil {
		log.Fatalf("StartSession failed: %v", err)
	}

	fmt.Printf("   ✓ Session created: %s\n", result.SessionID)
	fmt.Printf("   ✓ Agent ID: %s\n", result.AgentID)

	sessionID := result.SessionID
	agentID := result.AgentID

	// Add some logs
	fmt.Println("\n2. Adding logs...")
	tools.AddLog(mcp.AddLogArgs{
		SessionID: sessionID,
		Type:      "action",
		Message:   "Started testing session tracking",
	})
	tools.AddLog(mcp.AddLogArgs{
		SessionID: sessionID,
		Type:      "info",
		Message:   "Database connection established",
	})
	fmt.Println("   ✓ Logs added")

	// Add a note
	fmt.Println("\n3. Adding note...")
	tools.AddNote(mcp.AddNoteArgs{
		SessionID: sessionID,
		Content:   "Testing complete session workflow with database verification",
	})
	fmt.Println("   ✓ Note added")

	// Set context
	fmt.Println("\n4. Setting context...")
	tools.SetContext(mcp.SetContextArgs{
		SessionID: sessionID,
		Key:       "test_type",
		Value:     "integration",
	})
	tools.SetContext(mcp.SetContextArgs{
		SessionID: sessionID,
		Key:       "database",
		Value:     "production",
	})
	fmt.Println("   ✓ Context set")

	// Update status
	fmt.Println("\n5. Updating agent status...")
	currentTask := "Verifying database writes"
	tools.UpdateStatus(mcp.UpdateStatusArgs{
		AgentID:     agentID,
		Status:      "working",
		CurrentTask: &currentTask,
	})
	fmt.Println("   ✓ Status updated")

	// Log action
	fmt.Println("\n6. Logging action...")
	details := `{"test": "session_verification", "database": "agents.db"}`
	tools.LogAction(mcp.LogActionArgs{
		AgentID:     agentID,
		ActionType:  "task_start",
		Description: "Started session verification test",
		Details:     &details,
	})
	fmt.Println("   ✓ Action logged")

	// Verify database writes
	fmt.Println("\n7. Verifying database...")

	// Check agent exists
	agent, err := db.GetAgent(agentID)
	if err != nil {
		log.Fatalf("Failed to get agent: %v", err)
	}
	fmt.Printf("   ✓ Agent found: %s (status: %s)\n", agent.ID, agent.Status)

	// Check session exists
	session, err := db.GetSession(sessionID)
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}
	fmt.Printf("   ✓ Session found: %s\n", session.ID)

	// Check logs
	logs, err := db.GetLogs(sessionID)
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}
	fmt.Printf("   ✓ Logs count: %d\n", len(logs))

	// Check notes
	notes, err := db.GetNotes(sessionID)
	if err != nil {
		log.Fatalf("Failed to get notes: %v", err)
	}
	fmt.Printf("   ✓ Notes count: %d\n", len(notes))

	// Check context
	context, err := db.GetContext(sessionID)
	if err != nil {
		log.Fatalf("Failed to get context: %v", err)
	}
	fmt.Printf("   ✓ Context entries: %d\n", len(context))

	// Check actions
	actions, err := db.GetActionsForAgent(agentID, 10)
	if err != nil {
		log.Fatalf("Failed to get actions: %v", err)
	}
	fmt.Printf("   ✓ Actions count: %d\n", len(actions))

	// Get full session data
	fmt.Println("\n8. Retrieving full session data...")
	sessionData, err := tools.GetSession(mcp.GetSessionArgs{
		SessionID: sessionID,
	})
	if err != nil {
		log.Fatalf("GetSession failed: %v", err)
	}
	fmt.Printf("   ✓ Session ID: %s\n", sessionData.SessionID)
	fmt.Printf("   ✓ Agent ID: %s\n", sessionData.AgentID)
	fmt.Printf("   ✓ Logs: %d\n", len(sessionData.Logs))
	fmt.Printf("   ✓ Notes: %d\n", len(sessionData.Notes))
	fmt.Printf("   ✓ Context: %d\n", len(sessionData.Context))

	fmt.Println("\n=== Session Test Complete ===")
	fmt.Printf("\n✅ Active session: %s\n", sessionID)
	fmt.Printf("✅ Agent ID: %s\n", agentID)
	fmt.Printf("✅ View at: http://localhost:8080/agent/%s\n", agentID)
	fmt.Println("\n⚠️  Session left active for dashboard viewing")
}
