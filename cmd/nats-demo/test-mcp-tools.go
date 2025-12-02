package main

import (
	"encoding/json"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/mcp"
)

func main() {
	tools := mcp.NewTools(nil) // nil DB for testing messaging only

	// Test sending a message
	fmt.Println("=== Testing NATS Send ===")
	sendArgs := mcp.NATSSendArgs{
		SenderID:   "test-session",
		ReceiverID: "cb16a4e0-ec93-49df-8627-d477659758b0",
		Message:    "Hello from MCP test tool!",
	}

	sendResult, err := tools.NATSSend(sendArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		resultJSON, _ := json.MarshalIndent(sendResult, "", "  ")
		fmt.Printf("Result: %s\n\n", resultJSON)
	}

	// Test listening for messages
	fmt.Println("=== Testing NATS Listen ===")
	listenArgs := mcp.NATSListenArgs{
		SessionID: "cb16a4e0-ec93-49df-8627-d477659758b0",
	}

	listenResult, err := tools.NATSListen(listenArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		resultJSON, _ := json.MarshalIndent(listenResult, "", "  ")
		fmt.Printf("Result: %s\n", resultJSON)
	}
}
