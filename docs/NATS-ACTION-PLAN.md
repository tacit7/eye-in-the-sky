

# **NATS JetStream Integration Plan**

## **1. Overview**
JetStream is NATS’ persistence and replay system. It adds durability, acknowledgment tracking, and replay capability on top of NATS Core’s lightweight pub/sub model.
Use JetStream when you need:
- Durable message storage and replay.
- Consumers that can reconnect and resume from sequence numbers.
- Kafka-like reliability without the operational overhead.

---

## **2. Setup**

### **Install NATS Server**
```bash
brew install nats-server
```

### **Enable JetStream**
Run NATS with JetStream enabled:
```bash
nats-server -js
```

Check that JetStream is active:
```bash
nats server report jetstream
```

---

## **3. Go Dependencies**
Install the Go client library:
```bash
go get github.com/nats-io/nats.go
```

---

## **4. Implementation Example**

### **Stream Setup**
You can create a stream manually or from code.
Using the CLI:
```bash
nats stream add EVENTS --subjects="events.*" --storage=file --retention=limits
```

### **Go Example**
```go
package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to local NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	// Ensure stream exists
	streamName := "EVENTS"
	_, err = js.StreamInfo(streamName)
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"events.*"},
			Storage:  nats.FileStorage,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Stream created:", streamName)
	}

	// Publish messages
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("Event #%d", i)
		_, err := js.Publish("events.user", []byte(msg))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Published:", msg)
	}

	// Create a durable consumer
	sub, err := js.PullSubscribe("events.user", "USER_CONSUMER")
	if err != nil {
		log.Fatal(err)
	}

	// Fetch messages
	msgs, err := sub.Fetch(5)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range msgs {
		fmt.Printf("Received: %s\n", string(msg.Data))
		msg.Ack()
	}
}
```

---

## **5. Replaying Messages**

You can replay by sequence number or timestamp.

**CLI:**
```bash
nats consumer add EVENTS myreplayer --filter "events.*" --deliver all
nats consumer next EVENTS myreplayer
```

**Go:**
```go
sub, _ := js.PullSubscribe("events.user", "REPLAYER",
	nats.DeliverAll(),
)
```

---

## **6. Testing & Verification**

**List streams:**
```bash
nats stream ls
```

**View stored messages:**
```bash
nats stream view EVENTS
```

**Replay all messages:**
```bash
nats consumer add EVENTS replay --deliver all
nats consumer next EVENTS replay
```

Restart NATS or your Go consumer and confirm that JetStream resumes correctly and replays from the last acknowledged sequence.

---

## **7. Action Plan for Junior Developer**

1. **Install and run NATS server with JetStream**
   ```bash
   brew install nats-server
   nats-server -js
   ```

2. **Verify JetStream is running**
   ```bash
   nats server report jetstream
   ```

3. **Create project directory**
   ```bash
   mkdir nats-demo && cd nats-demo
   go mod init nats-demo
   go get github.com/nats-io/nats.go
   ```

4. **Copy the Go code above into `main.go`.**

5. **Run it to create the stream and publish 5 messages.**

6. **Stop and rerun the program.**
   - Messages will replay if JetStream is enabled.
   - Confirm durable subscription works by restarting.

7. **Experiment:**
   - Change subjects (`events.user`, `events.order`).
   - Add `DeliverLast()` or `DeliverAll()` for replay modes.
   - Observe message order and persistence.

8. **Document the results in `docs/NATS-ACTION-PLAN.md`.**

---

With this plan in place, the team can quickly deploy a reliable pub/sub system with replay capability using JetStream instead of Kafka.