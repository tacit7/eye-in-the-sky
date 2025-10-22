package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Consumer handles message consumption from NATS JetStream
type Consumer struct {
	js   nats.JetStreamContext
	name string
}

// NewConsumer creates a new consumer instance
func NewConsumer(nc *nats.Conn, name string) (*Consumer, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}
	return &Consumer{js: js, name: name}, nil
}

// ConsumeWithOptions demonstrates various consumption patterns
func (c *Consumer) ConsumeWithOptions(subject string, opts ...nats.SubOpt) error {
	sub, err := c.js.PullSubscribe(subject, c.name, opts...)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Fetch messages with timeout
	msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
	if err != nil && err != nats.ErrTimeout {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	fmt.Printf("  📥 Consumer '%s' received %d messages:\n", c.name, len(msgs))
	for i, msg := range msgs {
		// Get message metadata
		meta, _ := msg.Metadata()

		// Try to parse as JSON
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err == nil {
			fmt.Printf("    [%d] Seq:%d - Event ID:%d, Action:%s, User:%s\n",
				i+1, meta.Sequence.Stream, event.ID, event.Action, event.UserID)
		} else {
			fmt.Printf("    [%d] Seq:%d - Raw: %s\n",
				i+1, meta.Sequence.Stream, string(msg.Data))
		}

		// Acknowledge message
		if err := msg.Ack(); err != nil {
			fmt.Printf("    ⚠️  Failed to ack message: %v\n", err)
		}
	}

	return nil
}

// ReplayFromBeginning replays all messages from the stream start
func (c *Consumer) ReplayFromBeginning(subject string) error {
	fmt.Printf("\n🔄 Replaying all messages from beginning for consumer '%s'\n", c.name)

	return c.ConsumeWithOptions(subject,
		nats.DeliverAll(),
		nats.AckExplicit(),
		nats.Description("Replay consumer - delivers all messages"),
	)
}

// ReplayFromSequence replays messages starting from a specific sequence
func (c *Consumer) ReplayFromSequence(subject string, seq uint64) error {
	fmt.Printf("\n🔄 Replaying from sequence %d for consumer '%s'\n", seq, c.name)

	// Start from specific sequence
	startSeq := nats.StartSequence(seq)
	return c.ConsumeWithOptions(subject, startSeq, nats.AckExplicit())
}

// ReplayFromTime replays messages from a specific time
func (c *Consumer) ReplayFromTime(subject string, startTime time.Time) error {
	fmt.Printf("\n🔄 Replaying from %s for consumer '%s'\n",
		startTime.Format(time.RFC3339), c.name)

	return c.ConsumeWithOptions(subject,
		nats.StartTime(startTime),
		nats.AckExplicit(),
	)
}

// ConsumeNewOnly consumes only new messages (after consumer creation)
func (c *Consumer) ConsumeNewOnly(subject string) error {
	fmt.Printf("\n📨 Consuming new messages only for '%s'\n", c.name)

	return c.ConsumeWithOptions(subject,
		nats.DeliverNew(),
		nats.AckExplicit(),
	)
}

// ConsumeLastPerSubject gets the last message for each subject
func (c *Consumer) ConsumeLastPerSubject(subject string) error {
	fmt.Printf("\n📨 Getting last message per subject for '%s'\n", c.name)

	return c.ConsumeWithOptions(subject,
		nats.DeliverLastPerSubject(),
		nats.AckExplicit(),
	)
}

// BatchConsume demonstrates batch consumption
func (c *Consumer) BatchConsume(subject string, batchSize int) error {
	fmt.Printf("\n📦 Batch consuming %d messages for '%s'\n", batchSize, c.name)

	sub, err := c.js.PullSubscribe(subject, c.name+"_BATCH",
		nats.DeliverAll(),
		nats.AckExplicit(),
		nats.MaxAckPending(batchSize),
	)
	if err != nil {
		return err
	}

	// Process in batches
	totalProcessed := 0
	for {
		msgs, err := sub.Fetch(batchSize, nats.MaxWait(1*time.Second))
		if err == nats.ErrTimeout {
			break
		}
		if err != nil {
			return err
		}

		if len(msgs) == 0 {
			break
		}

		fmt.Printf("  📦 Processing batch of %d messages\n", len(msgs))
		for _, msg := range msgs {
			msg.Ack()
			totalProcessed++
		}

		// Simulate processing time
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("  ✓ Total processed: %d messages\n", totalProcessed)
	return nil
}

// RunConsumerDemo demonstrates various consumption scenarios
func RunConsumerDemo() error {
	fmt.Println("\n🚀 Starting Consumer Demo")
	fmt.Println("=" + string(make([]byte, 50)))

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("Consumer-Demo"))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer nc.Drain()

	// Scenario 1: Basic consumption
	fmt.Println("\n📌 Scenario 1: Basic Consumption")
	basicConsumer, err := NewConsumer(nc, "BASIC_CONSUMER")
	if err != nil {
		return err
	}
	if err := basicConsumer.ConsumeNewOnly("events.user"); err != nil {
		fmt.Printf("  ℹ️  No new messages (expected if running after publisher)\n")
	}

	// Scenario 2: Replay from beginning
	fmt.Println("\n📌 Scenario 2: Replay All Messages")
	replayConsumer, err := NewConsumer(nc, "REPLAY_ALL")
	if err != nil {
		return err
	}
	if err := replayConsumer.ReplayFromBeginning("events.*"); err != nil {
		return err
	}

	// Scenario 3: Replay from specific sequence
	fmt.Println("\n📌 Scenario 3: Replay from Sequence 3")
	seqConsumer, err := NewConsumer(nc, "REPLAY_SEQ")
	if err != nil {
		return err
	}
	if err := seqConsumer.ReplayFromSequence("events.*", 3); err != nil {
		return err
	}

	// Scenario 4: Replay from time (1 minute ago)
	fmt.Println("\n📌 Scenario 4: Replay from 1 minute ago")
	timeConsumer, err := NewConsumer(nc, "REPLAY_TIME")
	if err != nil {
		return err
	}
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	if err := timeConsumer.ReplayFromTime("events.*", oneMinuteAgo); err != nil {
		return err
	}

	// Scenario 5: Get last message per subject
	fmt.Println("\n📌 Scenario 5: Last Message Per Subject")
	lastConsumer, err := NewConsumer(nc, "LAST_PER_SUBJECT")
	if err != nil {
		return err
	}
	if err := lastConsumer.ConsumeLastPerSubject("events.*"); err != nil {
		return err
	}

	// Scenario 6: Batch consumption
	fmt.Println("\n📌 Scenario 6: Batch Consumption")
	batchConsumer, err := NewConsumer(nc, "BATCH_CONSUMER")
	if err != nil {
		return err
	}
	if err := batchConsumer.BatchConsume("events.*", 5); err != nil {
		return err
	}

	// Scenario 7: Demonstrate durability
	fmt.Println("\n📌 Scenario 7: Durability Test")
	fmt.Println("  Creating durable consumer 'DURABLE_TEST'...")
	durableConsumer, err := NewConsumer(nc, "DURABLE_TEST")
	if err != nil {
		return err
	}

	// First fetch - get some messages
	sub, err := durableConsumer.js.PullSubscribe("events.*", "DURABLE_TEST",
		nats.DeliverAll(),
		nats.AckExplicit(),
		nats.Durable("DURABLE_TEST"),
	)
	if err != nil {
		return err
	}

	msgs, _ := sub.Fetch(3, nats.MaxWait(1*time.Second))
	fmt.Printf("  First fetch: Got %d messages\n", len(msgs))
	for _, msg := range msgs {
		msg.Ack()
	}

	// Simulate disconnect and reconnect
	fmt.Println("  Simulating disconnect...")
	sub.Unsubscribe()
	time.Sleep(1 * time.Second)

	// Reconnect to same durable consumer
	fmt.Println("  Reconnecting to durable consumer...")
	sub2, err := durableConsumer.js.PullSubscribe("events.*", "DURABLE_TEST",
		nats.Durable("DURABLE_TEST"),
	)
	if err != nil {
		return err
	}

	msgs2, _ := sub2.Fetch(5, nats.MaxWait(1*time.Second))
	fmt.Printf("  Second fetch: Got %d messages (should continue from where left off)\n", len(msgs2))
	for i, msg := range msgs2 {
		meta, _ := msg.Metadata()
		fmt.Printf("    Message %d: Seq %d\n", i+1, meta.Sequence.Stream)
		msg.Ack()
	}

	fmt.Println("\n✅ Consumer demo completed successfully!")
	return nil
}

// Allow standalone execution
func init() {
	// Usage: go run consumer.go
}