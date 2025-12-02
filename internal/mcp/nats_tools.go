package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSSendArgs represents arguments for sending a NATS message
type NATSSendArgs struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id,omitempty"` // Optional - empty means broadcast
	Message    string `json:"message"`
	Subject    string `json:"subject,omitempty"` // Optional - defaults to events.test
}

// NATSSendResult represents the result of sending a NATS message
type NATSSendResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Sequence uint64 `json:"sequence"`
	Subject  string `json:"subject"`
}

// NATSListenArgs represents arguments for checking NATS messages
type NATSListenArgs struct {
	SessionID    string `json:"session_id"`               // Current session ID to filter messages
	LastSequence uint64 `json:"last_sequence,omitempty"`  // Last processed sequence (0 = get all new)
	MaxMessages  int    `json:"max_messages,omitempty"`   // Max messages to fetch (default 10)
}

// NATSMessage represents a message received from NATS
type NATSMessage struct {
	Sequence     uint64 `json:"sequence"`
	Timestamp    string `json:"timestamp"`
	Sender       string `json:"sender"`
	Receiver     string `json:"receiver"`
	Message      string `json:"message"`
	Subject      string `json:"subject"`
}

// NATSListenResult represents the result of checking for NATS messages
type NATSListenResult struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	Messages     []NATSMessage `json:"messages,omitempty"`
	Count        int           `json:"count"`
	LastSequence uint64        `json:"last_sequence"` // Track for next call
}

// NATSSend sends a message via NATS JetStream
func (t *Tools) NATSSend(args NATSSendArgs) (NATSSendResult, error) {
	// Normalize subject to events.* pattern
	subject := args.Subject
	if subject == "" {
		subject = "events.test"
	} else if !strings.HasPrefix(subject, "events.") {
		// Auto-prefix with events. if not already present
		subject = "events." + subject
	}

	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return NATSSendResult{Success: false, Message: fmt.Sprintf("Failed to connect to NATS: %v", err)}, nil
	}
	defer nc.Close()

	// Get JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return NATSSendResult{Success: false, Message: fmt.Sprintf("Failed to get JetStream: %v", err)}, nil
	}

	// Ensure EVENTS stream exists
	streamName := "EVENTS"
	_, err = js.StreamInfo(streamName)
	if err != nil {
		// Stream doesn't exist, create it
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"events.*"},
			Storage:  nats.FileStorage,
			MaxAge:   24 * time.Hour,
		})
		if err != nil {
			return NATSSendResult{Success: false, Message: fmt.Sprintf("Failed to create stream: %v", err)}, nil
		}
	}

	// Create message payload
	payload := map[string]interface{}{
		"message":   args.Message,
		"timestamp": time.Now().Format(time.RFC3339),
		"sender":    args.SenderID,
		"receiver":  args.ReceiverID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return NATSSendResult{Success: false, Message: fmt.Sprintf("Failed to marshal JSON: %v", err)}, nil
	}

	// Publish to NATS
	ack, err := js.Publish(subject, data)
	if err != nil {
		return NATSSendResult{Success: false, Message: fmt.Sprintf("Failed to publish: %v", err)}, nil
	}

	messageType := "broadcast"
	if args.ReceiverID != "" {
		messageType = fmt.Sprintf("targeted to %s", args.ReceiverID)
	}

	return NATSSendResult{
		Success:  true,
		Message:  fmt.Sprintf("Message sent (%s)", messageType),
		Sequence: ack.Sequence,
		Subject:  subject,
	}, nil
}

// NATSListen queries NATS directly for new messages
func (t *Tools) NATSListen(args NATSListenArgs) (NATSListenResult, error) {
	// Default max messages
	maxMessages := 10
	if args.MaxMessages > 0 {
		maxMessages = args.MaxMessages
	}

	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return NATSListenResult{Success: false, Message: fmt.Sprintf("Failed to connect to NATS: %v", err)}, nil
	}
	defer nc.Close()

	// Get JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return NATSListenResult{Success: false, Message: fmt.Sprintf("Failed to get JetStream: %v", err)}, nil
	}

	// Check if stream exists
	streamName := "EVENTS"
	streamInfo, err := js.StreamInfo(streamName)
	if err != nil {
		return NATSListenResult{
			Success: true,
			Message: "No EVENTS stream found - no messages available",
			Count:   0,
		}, nil
	}

	// Calculate starting sequence
	startSeq := args.LastSequence + 1
	if startSeq == 0 {
		startSeq = 1
	}

	// Don't try to fetch beyond what exists
	if startSeq > streamInfo.State.LastSeq {
		return NATSListenResult{
			Success:      true,
			Message:      "No new messages",
			Count:        0,
			LastSequence: args.LastSequence,
		}, nil
	}

	// Create a temporary pull consumer to fetch messages
	consumerConfig := &nats.ConsumerConfig{
		DeliverPolicy: nats.DeliverByStartSequencePolicy,
		OptStartSeq:   startSeq,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: "events.*",
	}

	consumer, err := js.AddConsumer(streamName, consumerConfig)
	if err != nil {
		return NATSListenResult{Success: false, Message: fmt.Sprintf("Failed to create consumer: %v", err)}, nil
	}

	// Fetch messages
	sub, err := js.PullSubscribe("events.*", "", nats.BindStream(streamName))
	if err != nil {
		return NATSListenResult{Success: false, Message: fmt.Sprintf("Failed to subscribe: %v", err)}, nil
	}
	defer sub.Unsubscribe()

	// Pull messages
	msgs, err := sub.Fetch(maxMessages, nats.MaxWait(2*time.Second))
	if err != nil && err != nats.ErrTimeout {
		return NATSListenResult{Success: false, Message: fmt.Sprintf("Failed to fetch messages: %v", err)}, nil
	}

	// Parse and filter messages
	var messages []NATSMessage
	var lastSeq uint64 = args.LastSequence

	for _, msg := range msgs {
		meta, err := msg.Metadata()
		if err != nil {
			continue
		}

		// Parse message payload
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			continue
		}

		// Extract fields
		sender, _ := payload["sender"].(string)
		receiver, _ := payload["receiver"].(string)
		message, _ := payload["message"].(string)
		timestamp, _ := payload["timestamp"].(string)

		// Filter by receiver (targeted or broadcast)
		isForMe := receiver == "" || receiver == args.SessionID
		if !isForMe {
			msg.Ack()
			lastSeq = meta.Sequence.Stream
			continue
		}

		// Add to results
		messages = append(messages, NATSMessage{
			Sequence:  meta.Sequence.Stream,
			Timestamp: timestamp,
			Sender:    sender,
			Receiver:  receiver,
			Message:   message,
			Subject:   msg.Subject,
		})

		lastSeq = meta.Sequence.Stream
		msg.Ack()
	}

	// Clean up temporary consumer
	js.DeleteConsumer(streamName, consumer.Name)

	if len(messages) == 0 {
		return NATSListenResult{
			Success:      true,
			Message:      "No new messages for this session",
			Count:        0,
			LastSequence: lastSeq,
		}, nil
	}

	return NATSListenResult{
		Success:      true,
		Message:      fmt.Sprintf("Found %d new message(s)", len(messages)),
		Messages:     messages,
		Count:        len(messages),
		LastSequence: lastSeq,
	}, nil
}

