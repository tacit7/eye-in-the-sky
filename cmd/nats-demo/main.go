package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	fmt.Println("🚀 NATS JetStream Demo")
	fmt.Println("=" + string(make([]byte, 50)))

	// Connect to NATS
	nc, err := connectToNATS()
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Drain()

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	// Setup stream
	if err := setupStream(js); err != nil {
		log.Fatalf("Failed to setup stream: %v", err)
	}

	// Demo scenarios
	fmt.Println("\n📤 Publishing Messages...")
	if err := publishMessages(js); err != nil {
		log.Fatalf("Failed to publish messages: %v", err)
	}

	fmt.Println("\n📥 Consuming Messages (Durable Consumer)...")
	if err := consumeMessages(js); err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	fmt.Println("\n🔄 Demonstrating Replay...")
	if err := replayMessages(js); err != nil {
		log.Fatalf("Failed to replay messages: %v", err)
	}

	fmt.Println("\n👥 Multiple Consumers Demo...")
	if err := multipleConsumersDemo(js); err != nil {
		log.Fatalf("Failed in multiple consumers demo: %v", err)
	}

	fmt.Println("\n✅ Demo completed successfully!")
}

func connectToNATS() (*nats.Conn, error) {
	fmt.Println("📡 Connecting to NATS server...")

	nc, err := nats.Connect(
		nats.DefaultURL,
		nats.Name("NATS-Demo-Client"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(5),
	)

	if err != nil {
		return nil, err
	}

	fmt.Printf("✓ Connected to %s\n", nc.ConnectedUrl())
	return nc, nil
}

func setupStream(js nats.JetStreamContext) error {
	streamName := "EVENTS"

	// Check if stream exists
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		// Create new stream
		fmt.Printf("📦 Creating new stream: %s\n", streamName)

		streamConfig := &nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"events.*"},
			Storage:  nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxMsgs:  10000,
			MaxBytes: 1048576, // 1MB
			MaxAge:   24 * time.Hour,
		}

		stream, err = js.AddStream(streamConfig)
		if err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}

		fmt.Printf("✓ Stream created: %s\n", streamName)
	} else {
		fmt.Printf("✓ Stream already exists: %s (Messages: %d)\n",
			stream.Config.Name, stream.State.Msgs)
	}

	return nil
}

func publishMessages(js nats.JetStreamContext) error {
	subject := "events.user"

	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf(`{"id": %d, "event": "user_action", "timestamp": "%s"}`,
			i, time.Now().Format(time.RFC3339))

		ack, err := js.Publish(subject, []byte(msg))
		if err != nil {
			return fmt.Errorf("failed to publish message %d: %w", i, err)
		}

		fmt.Printf("  ✓ Published message #%d (seq: %d)\n", i, ack.Sequence)
		time.Sleep(100 * time.Millisecond) // Small delay for demo purposes
	}

	return nil
}

func consumeMessages(js nats.JetStreamContext) error {
	subject := "events.user"
	consumerName := "USER_CONSUMER"

	// Create durable pull consumer
	sub, err := js.PullSubscribe(subject, consumerName)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	// Fetch messages
	msgs, err := sub.Fetch(5)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	for i, msg := range msgs {
		fmt.Printf("  📨 Received message #%d: %s\n", i+1, string(msg.Data))

		// Acknowledge message
		if err := msg.Ack(); err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}
	}

	fmt.Printf("  ✓ Consumed and acknowledged %d messages\n", len(msgs))
	return nil
}

func replayMessages(js nats.JetStreamContext) error {
	subject := "events.user"
	replayConsumer := "REPLAY_CONSUMER"

	// Create consumer that starts from beginning
	sub, err := js.PullSubscribe(subject, replayConsumer,
		nats.DeliverAll(),
		nats.AckExplicit(),
	)
	if err != nil {
		return fmt.Errorf("failed to create replay consumer: %w", err)
	}

	// Fetch all available messages
	msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
	if err != nil && err != nats.ErrTimeout {
		return fmt.Errorf("failed to fetch replay messages: %w", err)
	}

	fmt.Printf("  🔄 Replaying %d messages from stream start:\n", len(msgs))
	for i, msg := range msgs {
		meta, _ := msg.Metadata()
		fmt.Printf("    Message #%d (Seq: %d): %s\n",
			i+1, meta.Sequence.Stream, string(msg.Data))
		msg.Ack()
	}

	return nil
}

func multipleConsumersDemo(js nats.JetStreamContext) error {
	subject := "events.*"

	// Publish some test messages for multiple subjects
	fmt.Println("  📤 Publishing messages to multiple subjects...")
	for _, topic := range []string{"events.order", "events.payment", "events.user"} {
		for i := 1; i <= 2; i++ {
			msg := fmt.Sprintf(`{"topic": "%s", "id": %d}`, topic, i)
			if _, err := js.Publish(topic, []byte(msg)); err != nil {
				return err
			}
			fmt.Printf("    ✓ Published to %s\n", topic)
		}
	}

	// Create three different consumers with different configurations
	consumers := []struct {
		name   string
		config string
		opts   []nats.SubOpt
	}{
		{
			name:   "REALTIME_CONSUMER",
			config: "New messages only",
			opts:   []nats.SubOpt{nats.DeliverNew()},
		},
		{
			name:   "ALL_CONSUMER",
			config: "All messages from start",
			opts:   []nats.SubOpt{nats.DeliverAll()},
		},
		{
			name:   "LAST_CONSUMER",
			config: "Last message per subject",
			opts:   []nats.SubOpt{nats.DeliverLastPerSubject()},
		},
	}

	fmt.Println("\n  👥 Running multiple consumers:")
	for _, consumer := range consumers {
		fmt.Printf("\n  Consumer: %s (%s)\n", consumer.name, consumer.config)

		sub, err := js.PullSubscribe(subject, consumer.name, consumer.opts...)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", consumer.name, err)
		}

		msgs, err := sub.Fetch(10, nats.MaxWait(1*time.Second))
		if err != nil && err != nats.ErrTimeout {
			return fmt.Errorf("failed to fetch from %s: %w", consumer.name, err)
		}

		if len(msgs) == 0 {
			fmt.Println("    No messages (expected for REALTIME_CONSUMER)")
		} else {
			for _, msg := range msgs {
				fmt.Printf("    Received: %s\n", string(msg.Data))
				msg.Ack()
			}
		}
	}

	return nil
}