package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/mcp"
)

func main() {
	tools := mcp.NewTools(nil)
	sessionID := "cb16a4e0-ec93-49df-8627-d477659758b0"

	// Send a message
	fmt.Println("=== Sending Message ===")
	sendArgs := mcp.NATSSendArgs{
		SenderID:   "test-mcp",
		ReceiverID: sessionID,
		Message:    "execute: list all files in current directory",
	}

	sendResult, err := tools.NATSSend(sendArgs)
	if err != nil {
		fmt.Printf("Error sending: %v\n", err)
		return
	}
	fmt.Printf("✅ Sent message (sequence %d)\n\n", sendResult.Sequence)

	// Wait briefly for message to be stored
	time.Sleep(1 * time.Second)

	// Listen for messages (starting from beginning)
	fmt.Println("=== Listening for All Messages ===")
	listenArgs := mcp.NATSListenArgs{
		SessionID:    sessionID,
		LastSequence: 0, // Get all messages
		MaxMessages:  20,
	}

	listenResult, err := tools.NATSListen(listenArgs)
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
		return
	}

	resultJSON, _ := json.MarshalIndent(listenResult, "", "  ")
	fmt.Printf("%s\n\n", resultJSON)

	if listenResult.Count > 0 {
		fmt.Println("📨 Messages Received:")
		for _, msg := range listenResult.Messages {
			msgType := "BROADCAST"
			if msg.Receiver != "" {
				msgType = "TARGETED"
			}
			fmt.Printf("  [%s] Seq:%d From:%s - %s\n", msgType, msg.Sequence, msg.Sender, msg.Message)
		}
		fmt.Printf("\n✅ Last sequence: %d\n", listenResult.LastSequence)

		// Now try listening again with last sequence
		fmt.Println("\n=== Listening for New Messages Only ===")
		listenArgs.LastSequence = listenResult.LastSequence

		newResult, _ := tools.NATSListen(listenArgs)
		fmt.Printf("Result: %s (count: %d)\n", newResult.Message, newResult.Count)
	}
}
