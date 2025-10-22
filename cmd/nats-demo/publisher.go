package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Event represents a generic event structure
type Event struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Publisher handles message publishing to NATS JetStream
type Publisher struct {
	js nats.JetStreamContext
}

// NewPublisher creates a new publisher instance
func NewPublisher(nc *nats.Conn) (*Publisher, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}
	return &Publisher{js: js}, nil
}

// PublishEvent publishes a single event to the specified subject
func (p *Publisher) PublishEvent(subject string, event Event) (*nats.PubAck, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	ack, err := p.js.Publish(subject, data)
	if err != nil {
		return nil, fmt.Errorf("failed to publish to %s: %w", subject, err)
	}

	return ack, nil
}

// PublishBatch publishes multiple events to different subjects
func (p *Publisher) PublishBatch(events map[string][]Event) error {
	for subject, eventList := range events {
		fmt.Printf("\n📤 Publishing to subject: %s\n", subject)
		for _, event := range eventList {
			ack, err := p.PublishEvent(subject, event)
			if err != nil {
				return err
			}
			fmt.Printf("  ✓ Event ID:%d published (seq: %d)\n", event.ID, ack.Sequence)
		}
	}
	return nil
}

// RunPublisherDemo demonstrates various publishing scenarios
func RunPublisherDemo() error {
	fmt.Println("\n🚀 Starting Publisher Demo")
	fmt.Println("=" + string(make([]byte, 50)))

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("Publisher-Demo"))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer nc.Drain()

	// Create publisher
	publisher, err := NewPublisher(nc)
	if err != nil {
		return err
	}

	// Scenario 1: Publish user events
	fmt.Println("\n📌 Scenario 1: User Events")
	userEvents := []Event{
		{
			ID:        1,
			Type:      "user",
			UserID:    "user123",
			Action:    "login",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"ip":       "192.168.1.1",
				"device":   "mobile",
			},
		},
		{
			ID:        2,
			Type:      "user",
			UserID:    "user123",
			Action:    "profile_update",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"fields": []string{"email", "phone"},
			},
		},
		{
			ID:        3,
			Type:      "user",
			UserID:    "user456",
			Action:    "logout",
			Timestamp: time.Now(),
		},
	}

	for _, event := range userEvents {
		ack, err := publisher.PublishEvent("events.user", event)
		if err != nil {
			return err
		}
		fmt.Printf("  ✓ Published user event ID:%d (seq: %d) - %s\n",
			event.ID, ack.Sequence, event.Action)
		time.Sleep(100 * time.Millisecond)
	}

	// Scenario 2: Publish system events
	fmt.Println("\n📌 Scenario 2: System Events")
	systemEvents := []Event{
		{
			ID:        100,
			Type:      "system",
			UserID:    "system",
			Action:    "health_check",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"cpu_usage":    45.2,
				"memory_usage": 62.8,
				"status":       "healthy",
			},
		},
		{
			ID:        101,
			Type:      "system",
			UserID:    "system",
			Action:    "backup_started",
			Timestamp: time.Now(),
		},
	}

	for _, event := range systemEvents {
		ack, err := publisher.PublishEvent("events.system", event)
		if err != nil {
			return err
		}
		fmt.Printf("  ✓ Published system event ID:%d (seq: %d) - %s\n",
			event.ID, ack.Sequence, event.Action)
	}

	// Scenario 3: Batch publishing to multiple subjects
	fmt.Println("\n📌 Scenario 3: Batch Publishing")
	batchEvents := map[string][]Event{
		"events.order": {
			{ID: 200, Type: "order", UserID: "user789", Action: "order_created", Timestamp: time.Now()},
			{ID: 201, Type: "order", UserID: "user789", Action: "order_confirmed", Timestamp: time.Now()},
		},
		"events.payment": {
			{ID: 300, Type: "payment", UserID: "user789", Action: "payment_initiated", Timestamp: time.Now()},
			{ID: 301, Type: "payment", UserID: "user789", Action: "payment_completed", Timestamp: time.Now()},
		},
	}

	if err := publisher.PublishBatch(batchEvents); err != nil {
		return err
	}

	// Scenario 4: High-frequency publishing
	fmt.Println("\n📌 Scenario 4: High-Frequency Publishing (10 messages)")
	start := time.Now()
	for i := 1; i <= 10; i++ {
		event := Event{
			ID:        1000 + i,
			Type:      "metrics",
			UserID:    "monitor",
			Action:    "metric_recorded",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"metric_id": i,
				"value":     float64(i) * 10.5,
			},
		}

		ack, err := publisher.PublishEvent("events.metrics", event)
		if err != nil {
			return err
		}

		if i == 1 || i == 10 {
			fmt.Printf("  ✓ Metric %d published (seq: %d)\n", i, ack.Sequence)
		}
	}
	duration := time.Since(start)
	fmt.Printf("  ⚡ Published 10 messages in %v\n", duration)

	fmt.Println("\n✅ Publisher demo completed successfully!")
	return nil
}

// Main function for standalone execution
func init() {
	// This allows the file to be run standalone if needed
	// Usage: go run publisher.go
}

// Example of how to run standalone
func main() {
	if err := RunPublisherDemo(); err != nil {
		log.Fatal(err)
	}
}