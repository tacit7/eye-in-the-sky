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

func main() {
	// Generate unique consumer name with timestamp
	consumerName := fmt.Sprintf("LISTENER-%d", time.Now().Unix())

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

	fmt.Printf("🎧 NATS Message Listener Daemon (PID: %d)\n", os.Getpid())
	fmt.Println("========================")
	fmt.Println("Listening for messages on events.* subjects...")
	fmt.Printf("Consumer: %s\n", consumerName)
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

	// Subscribe using push consumer for real-time delivery
	sub, err := js.Subscribe("events.*", func(msg *nats.Msg) {
		// Get message metadata
		meta, err := msg.Metadata()
		if err != nil {
			fmt.Printf("⚠️  Error getting metadata: %v\n", err)
		}

		// Print message details
		fmt.Printf("\n📨 Message Received [%s]\n", time.Now().Format("15:04:05"))
		fmt.Printf("   Subject:  %s\n", msg.Subject)
		fmt.Printf("   Sequence: %d\n", meta.Sequence.Stream)
		fmt.Printf("   Time:     %s\n", meta.Timestamp.Format("2006-01-02 15:04:05"))

		// Try to parse as JSON for pretty printing
		var data interface{}
		if err := json.Unmarshal(msg.Data, &data); err == nil {
			prettyJSON, _ := json.MarshalIndent(data, "   ", "  ")
			fmt.Printf("   Data:     %s\n", string(prettyJSON))
		} else {
			// Not JSON, print raw
			fmt.Printf("   Data:     %s\n", string(msg.Data))
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

	fmt.Printf("\n✅ Subscribed to events.* (consumer: %s)\n", consumerName)
	fmt.Println("Waiting for messages...\n")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	fmt.Println("\n\n👋 Shutting down listener...")
}
