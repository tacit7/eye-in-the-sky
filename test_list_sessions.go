package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/mcp"
)

func main() {
	// Open database
	db, err := database.New("./data/agents.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tools instance
	tools := mcp.NewTools(db)

	// Test 1: List all sessions
	fmt.Println("=== Test 1: List all sessions ===")
	args1 := mcp.ListSessionsArgs{}
	result1, err := tools.ListSessions(args1)
	if err != nil {
		log.Fatalf("Test 1 failed: %v", err)
	}
	jsonData1, _ := json.MarshalIndent(result1, "", "  ")
	fmt.Printf("%s\n\n", jsonData1)

	// Test 2: List only active sessions
	fmt.Println("=== Test 2: List active sessions only ===")
	activeOnly := true
	args2 := mcp.ListSessionsArgs{
		ActiveOnly: &activeOnly,
	}
	result2, err := tools.ListSessions(args2)
	if err != nil {
		log.Fatalf("Test 2 failed: %v", err)
	}
	jsonData2, _ := json.MarshalIndent(result2, "", "  ")
	fmt.Printf("%s\n\n", jsonData2)

	// Test 3: List sessions for specific agent (use first agent from result1 if available)
	if len(result1.Sessions) > 0 {
		agentID := result1.Sessions[0].AgentID
		fmt.Printf("=== Test 3: List sessions for agent %s ===\n", agentID)
		args3 := mcp.ListSessionsArgs{
			AgentID: &agentID,
		}
		result3, err := tools.ListSessions(args3)
		if err != nil {
			log.Fatalf("Test 3 failed: %v", err)
		}
		jsonData3, _ := json.MarshalIndent(result3, "", "  ")
		fmt.Printf("%s\n\n", jsonData3)
	} else {
		fmt.Println("=== Test 3: Skipped (no sessions available) ===\n")
	}

	fmt.Println("All tests passed!")
}
