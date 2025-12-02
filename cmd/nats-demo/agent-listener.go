package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

type Message struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
}

func main() {
	// Get this agent's session ID from environment
	mySessionID := os.Getenv("SESSION_ID")
	if mySessionID == "" {
		log.Fatal("SESSION_ID environment variable is required")
	}

	// Generate unique consumer name
	consumerName := fmt.Sprintf("AGENT-%s-%d", mySessionID[:8], time.Now().Unix())

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

	fmt.Printf("🤖 Agent Listener (PID: %d)\n", os.Getpid())
	fmt.Println("========================")
	fmt.Printf("Session ID: %s\n", mySessionID)
	fmt.Printf("Consumer:   %s\n", consumerName)
	fmt.Println("Listening for targeted messages...")
	fmt.Println("Press Ctrl+C to stop\n")

	// Create or get existing stream
	streamName := "EVENTS"
	_, err = js.StreamInfo(streamName)
	if err != nil {
		// Stream doesn't exist, create it
		fmt.Printf("Creating stream: %s\n", streamName)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"events.*"},
			Storage:  nats.FileStorage,
			MaxAge:   24 * time.Hour,
		})
		if err != nil {
			log.Fatalf("Failed to create stream: %v", err)
		}
	}

	// Instructions queue file
	instructionsFile := "/tmp/agent-instructions.log"

	// Subscribe using push consumer for real-time delivery
	sub, err := js.Subscribe("events.*", func(msg *nats.Msg) {
		// Parse message
		var data Message
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			fmt.Printf("⚠️  Failed to parse message: %v\n", err)
			msg.Ack()
			return
		}

		// Check if message is for us (targeted or broadcast)
		isForMe := data.Receiver == "" || data.Receiver == mySessionID
		if !isForMe {
			// Not for us, skip silently
			msg.Ack()
			return
		}

		// Get message metadata
		meta, err := msg.Metadata()
		if err != nil {
			fmt.Printf("⚠️  Error getting metadata: %v\n", err)
		}

		// Print message details
		messageType := "📢 BROADCAST"
		if data.Receiver != "" {
			messageType = "🎯 TARGETED"
		}

		fmt.Printf("\n%s Message [%s]\n", messageType, time.Now().Format("15:04:05"))
		fmt.Printf("   From:     %s\n", data.Sender)
		fmt.Printf("   Subject:  %s\n", msg.Subject)
		fmt.Printf("   Sequence: %d\n", meta.Sequence.Stream)
		fmt.Printf("   Time:     %s\n", meta.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Instructions: %s\n", data.Message)

		// Write instruction to file for processing
		instruction := fmt.Sprintf("[%s] FROM=%s INSTR=%s\n",
			time.Now().Format(time.RFC3339),
			data.Sender,
			data.Message)

		f, err := os.OpenFile(instructionsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("⚠️  Failed to write instruction: %v\n", err)
		} else {
			f.WriteString(instruction)
			f.Close()
			fmt.Printf("   ✅ Instruction logged to %s\n", instructionsFile)
		}

		// Acknowledge the message
		msg.Ack()
	},
		nats.Durable(consumerName),
		nats.DeliverNew(), // Only new messages (not historical)
		nats.AckExplicit(),
		nats.ManualAck(),
	)

	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Printf("\n✅ Subscribed to events.* (filtering for session: %s)\n", mySessionID)
	fmt.Printf("📝 Instructions will be logged to: %s\n", instructionsFile)
	fmt.Println("Waiting for messages...\n")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\n\n👋 Shutting down agent listener...")
}
