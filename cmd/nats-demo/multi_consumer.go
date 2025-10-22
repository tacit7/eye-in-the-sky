package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// MultiConsumerDemo demonstrates multiple concurrent consumers
type MultiConsumerDemo struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// NewMultiConsumerDemo creates a new multi-consumer demo instance
func NewMultiConsumerDemo(nc *nats.Conn) (*MultiConsumerDemo, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}
	return &MultiConsumerDemo{nc: nc, js: js}, nil
}

// ConsumerWorker represents a concurrent consumer worker
type ConsumerWorker struct {
	ID          int
	Name        string
	Subject     string
	DeliverMode string
	Opts        []nats.SubOpt
	Messages    []string
	mu          sync.Mutex
}

// Run processes messages for this consumer worker
func (cw *ConsumerWorker) Run(js nats.JetStreamContext, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("  🔧 Worker %d (%s) starting - Mode: %s\n", cw.ID, cw.Name, cw.DeliverMode)

	// Create subscription with specified options
	sub, err := js.PullSubscribe(cw.Subject, cw.Name, cw.Opts...)
	if err != nil {
		fmt.Printf("    ❌ Worker %d failed to subscribe: %v\n", cw.ID, err)
		return
	}

	// Fetch messages
	msgs, err := sub.Fetch(10, nats.MaxWait(3*time.Second))
	if err != nil && err != nats.ErrTimeout {
		fmt.Printf("    ❌ Worker %d fetch error: %v\n", cw.ID, err)
		return
	}

	// Process messages
	cw.mu.Lock()
	for _, msg := range msgs {
		meta, _ := msg.Metadata()
		msgInfo := fmt.Sprintf("Seq:%d - %s", meta.Sequence.Stream, string(msg.Data))
		cw.Messages = append(cw.Messages, msgInfo)
		msg.Ack()
	}
	messageCount := len(msgs)
	cw.mu.Unlock()

	fmt.Printf("    ✓ Worker %d processed %d messages\n", cw.ID, messageCount)
}

// RunConcurrentConsumers demonstrates multiple consumers running simultaneously
func (mcd *MultiConsumerDemo) RunConcurrentConsumers() error {
	fmt.Println("\n👥 Running Multiple Concurrent Consumers")
	fmt.Println("  This demonstrates how multiple consumers can work independently")
	fmt.Println("  on the same stream with different delivery modes.")

	// First, publish some test messages
	fmt.Println("\n  📤 Publishing test messages...")
	for i := 1; i <= 15; i++ {
		subject := "events.user"
		if i%3 == 0 {
			subject = "events.system"
		} else if i%5 == 0 {
			subject = "events.metrics"
		}

		msg := fmt.Sprintf(`{"id":%d,"type":"%s","time":"%s"}`,
			i, subject, time.Now().Format(time.RFC3339))

		if _, err := mcd.js.Publish(subject, []byte(msg)); err != nil {
			return err
		}
	}
	fmt.Println("  ✓ Published 15 test messages across different subjects")

	// Define consumer workers with different configurations
	workers := []*ConsumerWorker{
		{
			ID:          1,
			Name:        "REALTIME_WORKER",
			Subject:     "events.*",
			DeliverMode: "New messages only",
			Opts:        []nats.SubOpt{nats.DeliverNew()},
		},
		{
			ID:          2,
			Name:        "REPLAY_WORKER",
			Subject:     "events.*",
			DeliverMode: "All messages from start",
			Opts:        []nats.SubOpt{nats.DeliverAll()},
		},
		{
			ID:          3,
			Name:        "USER_ONLY_WORKER",
			Subject:     "events.user",
			DeliverMode: "User events only",
			Opts:        []nats.SubOpt{nats.DeliverAll()},
		},
		{
			ID:          4,
			Name:        "LAST_MSG_WORKER",
			Subject:     "events.*",
			DeliverMode: "Last message per subject",
			Opts:        []nats.SubOpt{nats.DeliverLastPerSubject()},
		},
		{
			ID:          5,
			Name:        "SYSTEM_MONITOR",
			Subject:     "events.system",
			DeliverMode: "System events only",
			Opts:        []nats.SubOpt{nats.DeliverAll()},
		},
	}

	// Run all workers concurrently
	fmt.Println("\n  🚀 Starting concurrent consumers...")
	var wg sync.WaitGroup
	startTime := time.Now()

	for _, worker := range workers {
		wg.Add(1)
		go worker.Run(mcd.js, &wg)
		time.Sleep(100 * time.Millisecond) // Small delay to avoid race conditions
	}

	// Wait for all workers to complete
	wg.Wait()
	duration := time.Since(startTime)

	// Display results
	fmt.Printf("\n  ⏱️  All workers completed in %v\n", duration)
	fmt.Println("\n  📊 Results Summary:")
	for _, worker := range workers {
		worker.mu.Lock()
		fmt.Printf("    Worker %d (%s): Processed %d messages\n",
			worker.ID, worker.Name, len(worker.Messages))
		if len(worker.Messages) > 0 && len(worker.Messages) <= 3 {
			// Show first few messages for small result sets
			for _, msg := range worker.Messages {
				fmt.Printf("      - %s\n", msg)
			}
		}
		worker.mu.Unlock()
	}

	return nil
}

// RunLoadBalancedConsumers demonstrates load-balanced queue groups
func (mcd *MultiConsumerDemo) RunLoadBalancedConsumers() error {
	fmt.Println("\n⚖️  Load-Balanced Queue Group Demo")
	fmt.Println("  Multiple consumers share work from the same queue")

	// Publish work items
	fmt.Println("\n  📤 Publishing 20 work items...")
	for i := 1; i <= 20; i++ {
		workItem := fmt.Sprintf(`{"task_id":%d,"priority":%d,"work":"Process order #%d"}`,
			i, (i%3)+1, i)
		if _, err := mcd.js.Publish("work.queue", []byte(workItem)); err != nil {
			return err
		}
	}

	// Create queue group consumers
	queueGroup := "WORK_QUEUE_GROUP"
	numWorkers := 3
	var wg sync.WaitGroup
	results := make([][]string, numWorkers)

	fmt.Printf("\n  👷 Starting %d queue workers...\n", numWorkers)

	for workerID := 0; workerID < numWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each worker joins the same queue group
			sub, err := mcd.js.QueueSubscribe("work.queue", queueGroup,
				nats.DeliverAll(),
				nats.AckExplicit(),
				nats.ManualAck(),
			)
			if err != nil {
				fmt.Printf("    ❌ Worker %d failed: %v\n", id, err)
				return
			}

			// Process messages
			processed := 0
			for processed < 10 { // Try to get up to 10 messages
				msgs, err := sub.Fetch(1, nats.MaxWait(1*time.Second))
				if err == nats.ErrTimeout {
					break
				}
				if err != nil {
					fmt.Printf("    ❌ Worker %d fetch error: %v\n", id, err)
					break
				}

				for _, msg := range msgs {
					results[id] = append(results[id], string(msg.Data))
					msg.Ack()
					processed++
					time.Sleep(50 * time.Millisecond) // Simulate work
				}
			}

			fmt.Printf("    ✓ Worker %d processed %d items\n", id, len(results[id]))
		}(workerID)

		time.Sleep(100 * time.Millisecond) // Stagger worker starts
	}

	wg.Wait()

	// Show distribution
	fmt.Println("\n  📊 Work Distribution:")
	totalProcessed := 0
	for i, workerResults := range results {
		totalProcessed += len(workerResults)
		fmt.Printf("    Worker %d: %d items (%.1f%%)\n",
			i, len(workerResults),
			float64(len(workerResults))*100.0/20.0)
	}
	fmt.Printf("    Total processed: %d/20 items\n", totalProcessed)

	return nil
}

// RunCompetingConsumers demonstrates competing consumers pattern
func (mcd *MultiConsumerDemo) RunCompetingConsumers() error {
	fmt.Println("\n🏁 Competing Consumers Pattern")
	fmt.Println("  Multiple consumers compete for messages from the same stream")

	// Create stream with explicit config
	streamName := "COMPETITION"
	stream, err := mcd.js.StreamInfo(streamName)
	if err != nil {
		// Create new stream
		stream, err = mcd.js.AddStream(&nats.StreamConfig{
			Name:      streamName,
			Subjects:  []string{"compete.*"},
			Storage:   nats.FileStorage,
			Retention: nats.WorkQueuePolicy, // Messages removed after ack
		})
		if err != nil {
			return fmt.Errorf("failed to create competition stream: %w", err)
		}
	}
	fmt.Printf("  ✓ Using stream: %s\n", stream.Config.Name)

	// Publish messages
	fmt.Println("\n  📤 Publishing competition messages...")
	for i := 1; i <= 10; i++ {
		msg := fmt.Sprintf(`{"contest_id":%d,"prize":%d}`, i, i*100)
		if _, err := mcd.js.Publish("compete.race", []byte(msg)); err != nil {
			return err
		}
	}

	// Create competing consumers
	competitors := []string{"FAST_CONSUMER", "SLOW_CONSUMER", "MEDIUM_CONSUMER"}
	var wg sync.WaitGroup
	winnerCounts := make(map[string]int)
	var mu sync.Mutex

	fmt.Println("\n  🏃 Starting competition...")
	for _, name := range competitors {
		wg.Add(1)
		go func(consumerName string) {
			defer wg.Done()

			// Different processing speeds
			var delay time.Duration
			switch consumerName {
			case "FAST_CONSUMER":
				delay = 10 * time.Millisecond
			case "SLOW_CONSUMER":
				delay = 100 * time.Millisecond
			default:
				delay = 50 * time.Millisecond
			}

			// Create consumer
			sub, err := mcd.js.PullSubscribe("compete.*", consumerName,
				nats.DeliverAll(),
				nats.AckExplicit(),
			)
			if err != nil {
				fmt.Printf("    ❌ %s failed to join: %v\n", consumerName, err)
				return
			}

			// Compete for messages
			for {
				msgs, err := sub.Fetch(1, nats.MaxWait(500*time.Millisecond))
				if err == nats.ErrTimeout || len(msgs) == 0 {
					break
				}
				if err != nil {
					break
				}

				for _, msg := range msgs {
					time.Sleep(delay) // Processing time varies by consumer
					msg.Ack()

					mu.Lock()
					winnerCounts[consumerName]++
					mu.Unlock()
				}
			}
		}(name)

		time.Sleep(50 * time.Millisecond) // Stagger starts
	}

	wg.Wait()

	// Display results
	fmt.Println("\n  🏆 Competition Results:")
	for name, count := range winnerCounts {
		speed := "Medium"
		if name == "FAST_CONSUMER" {
			speed = "Fast"
		} else if name == "SLOW_CONSUMER" {
			speed = "Slow"
		}
		fmt.Printf("    %s (%s): Won %d messages\n", name, speed, count)
	}

	return nil
}

// RunMultiConsumerDemo executes all multi-consumer scenarios
func RunMultiConsumerDemo() error {
	fmt.Println("\n🚀 Starting Multi-Consumer Demo")
	fmt.Println("=" + string(make([]byte, 50)))

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("MultiConsumer-Demo"))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer nc.Drain()

	demo, err := NewMultiConsumerDemo(nc)
	if err != nil {
		return err
	}

	// Run different scenarios
	fmt.Println("\n📌 Scenario 1: Concurrent Consumers with Different Modes")
	if err := demo.RunConcurrentConsumers(); err != nil {
		return err
	}

	fmt.Println("\n📌 Scenario 2: Load-Balanced Queue Group")
	if err := demo.RunLoadBalancedConsumers(); err != nil {
		return err
	}

	fmt.Println("\n📌 Scenario 3: Competing Consumers Pattern")
	if err := demo.RunCompetingConsumers(); err != nil {
		return err
	}

	fmt.Println("\n✅ Multi-Consumer demo completed successfully!")
	return nil
}

// Main function for standalone execution
func main() {
	if err := RunMultiConsumerDemo(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}