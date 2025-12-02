package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// Get session_id from environment or use default
	sessionID := os.Getenv("SESSION_ID")
	if sessionID == "" {
		sessionID = "unknown"
	}

	// Get receiver session_id from environment (optional - for targeted messages)
	receiverID := os.Getenv("RECEIVER_ID")

	// Get message from command line or use default
	message := "Hello from NATS!"
	if len(os.Args) > 1 {
		message = os.Args[1]
	}

	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Get JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to get JetStream context: %v", err)
	}

	// Create message payload
	payload := map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
		"sender":    sessionID,
		"receiver":  receiverID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Publish to events.test subject
	subject := "events.test"
	ack, err := js.Publish(subject, data)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	fmt.Printf("✅ Message published\n")
	fmt.Printf("   Subject:  %s\n", subject)
	fmt.Printf("   Sequence: %d\n", ack.Sequence)
	fmt.Printf("   Message:  %s\n", message)
}
