# NATS JetStream Demo

A comprehensive demonstration of NATS JetStream capabilities, including message persistence, replay, and multiple consumer patterns. This demo is a proof-of-concept for integrating durable message passing into the Eye-in-the-Sky multi-agent management system.

## 🚀 Quick Start

```bash
# 1. Install and start NATS with JetStream
./scripts/setup.sh

# 2. Run the complete demo
./scripts/test.sh

# 3. Clean up when done
./scripts/cleanup.sh
```

## 📋 Prerequisites

- Go 1.21+ installed
- macOS or Linux (tested on macOS)
- Homebrew (for macOS) or ability to install binaries (Linux)

## 📦 What's Included

### Demo Programs

1. **`main.go`** - Complete orchestrated demo showing all features
2. **`publisher.go`** - Advanced publishing patterns and scenarios
3. **`consumer.go`** - Various consumption and replay strategies
4. **`multi_consumer.go`** - Concurrent consumers, queue groups, and competing patterns

### Scripts

- **`setup.sh`** - Installs NATS server/CLI and starts JetStream
- **`test.sh`** - Interactive test menu and demo runner
- **`cleanup.sh`** - Cleans up streams, consumers, and data

## 🎯 Key Features Demonstrated

### Message Persistence & Durability
- Messages survive server restarts
- Configurable retention policies (time, size, count)
- File-based storage for production use

### Replay Capabilities
- Replay from beginning (`DeliverAll`)
- Replay from specific sequence number
- Replay from timestamp
- Get last message per subject

### Consumer Patterns
- **Durable Consumers**: Resume from last acknowledged message
- **Queue Groups**: Load-balanced message distribution
- **Competing Consumers**: Multiple consumers competing for messages
- **Filtered Consumers**: Subject-based message filtering

### Production Features
- Acknowledgment tracking
- Batch processing
- Work queue patterns
- Real-time monitoring

## 💻 Installation

### Automated Setup

Run the setup script which handles everything:
```bash
cd cmd/nats-demo
./scripts/setup.sh
```

This will:
1. Check for NATS server and CLI
2. Install missing components via Homebrew (macOS) or direct download (Linux)
3. Start NATS server with JetStream enabled
4. Verify JetStream is running

### Manual Setup

If you prefer manual installation:

```bash
# Install NATS
brew install nats-server nats  # macOS
# OR download from https://github.com/nats-io/nats-server/releases

# Start NATS with JetStream
nats-server -js

# Install Go dependencies
go get github.com/nats-io/nats.go
```

## 🔧 Running the Demo

### Interactive Mode (Recommended)

```bash
./scripts/test.sh
```

This provides an interactive menu with options to:
- Run individual demo programs
- Manage streams and consumers
- Monitor server status
- Publish test messages
- View real-time metrics

### Run Individual Components

```bash
# Run complete demo
go run main.go

# Run publisher scenarios
go run publisher.go

# Run consumer patterns
go run consumer.go

# Run multi-consumer demo
go run multi_consumer.go
```

### Automatic Mode

For CI/CD or automated testing:
```bash
./scripts/test.sh --auto
```

## 📚 Usage Examples

### Basic Publishing

```go
js, _ := nc.JetStream()
ack, err := js.Publish("events.user", []byte(`{"action":"login"}`))
fmt.Printf("Message published, sequence: %d\n", ack.Sequence)
```

### Creating a Durable Consumer

```go
sub, _ := js.PullSubscribe("events.*", "MY_CONSUMER",
    nats.DeliverAll(),
    nats.AckExplicit(),
    nats.Durable("MY_CONSUMER"),
)
msgs, _ := sub.Fetch(10)
```

### Replay from Beginning

```go
sub, _ := js.PullSubscribe("events.*", "REPLAY_CONSUMER",
    nats.DeliverAll(),  // Start from beginning
    nats.AckExplicit(),
)
```

### Queue Group for Load Balancing

```go
sub, _ := js.QueueSubscribe("work.queue", "WORKERS",
    nats.DeliverAll(),
    nats.ManualAck(),
)
```

## 🛠️ CLI Commands Reference

### Stream Management

```bash
# List all streams
nats stream ls

# Create EVENTS stream
nats stream add EVENTS --subjects="events.*" --storage=file

# View stream info
nats stream info EVENTS

# View messages in stream
nats stream view EVENTS

# Monitor stream in real-time
nats stream view EVENTS --follow

# Delete stream
nats stream delete EVENTS -f
```

### Consumer Management

```bash
# List consumers
nats consumer ls EVENTS

# Create consumer
nats consumer add EVENTS MY_CONSUMER --deliver=all --ack=explicit

# Get next message
nats consumer next EVENTS MY_CONSUMER

# Delete consumer
nats consumer delete EVENTS MY_CONSUMER -f
```

### Server Monitoring

```bash
# Server information
nats server info

# JetStream status
nats server report jetstream

# Stream statistics
nats stream report

# Connection details
nats server report connections
```

### Publishing Messages

```bash
# Publish single message
echo '{"test":"data"}' | nats pub events.test

# Publish with reply subject
nats pub events.user --reply=responses.user '{"id":1}'

# Request-reply pattern
nats request help.request '{"need":"assistance"}'
```

## 🏗️ Architecture Decisions

### Why NATS JetStream?

1. **Lightweight**: Minimal operational overhead compared to Kafka
2. **Built-in Persistence**: No additional storage layer needed
3. **Flexible Replay**: Multiple replay strategies out of the box
4. **Native Go Support**: Excellent Go client library
5. **Low Latency**: Designed for real-time messaging

### Stream Configuration

```go
StreamConfig{
    Name:      "EVENTS",
    Subjects:  []string{"events.*"},  // Wildcard subjects
    Storage:   nats.FileStorage,      // Persistent storage
    Retention: nats.LimitsPolicy,     // Keep until limits
    MaxMsgs:   10000,                 // Message count limit
    MaxBytes:  1048576,               // 1MB size limit
    MaxAge:    24 * time.Hour,        // Time-based retention
}
```

### Consumer Types

- **Pull Consumers**: Application controls message flow
- **Push Consumers**: Server pushes messages to application
- **Queue Groups**: Multiple consumers share workload
- **Durable**: Survives restarts, tracks progress
- **Ephemeral**: Temporary, deleted when disconnected

## 📊 Retention Policies

### Limits Policy (Default)
- Keep messages until limits are reached
- Oldest messages deleted when limits exceeded
- Good for: Event sourcing, audit logs

### Work Queue Policy
- Messages deleted after acknowledgment
- Each message consumed exactly once
- Good for: Task distribution, job queues

### Interest Policy
- Keep messages while consumers exist
- Delete when no consumers interested
- Good for: Temporary data streams

## 🧪 Testing

### Run All Tests

```bash
cd cmd/nats-demo
go test -v ./...
```

### Verify Durability

1. Start the demo and publish messages:
   ```bash
   go run publisher.go
   ```

2. Stop NATS server:
   ```bash
   killall nats-server
   ```

3. Restart NATS server:
   ```bash
   nats-server -js
   ```

4. Run consumer to verify message replay:
   ```bash
   go run consumer.go
   ```

### Load Testing

```bash
# Publish 1000 messages
for i in {1..1000}; do
    echo "{\"id\":$i}" | nats pub test.load
done

# Monitor performance
nats stream report
```

## 🚨 Troubleshooting

### NATS Server Won't Start

```bash
# Check if already running
pgrep nats-server

# Kill existing process
killall nats-server

# Start with verbose logging
nats-server -js -V
```

### JetStream Not Enabled

```bash
# Verify JetStream status
nats server report jetstream

# If not enabled, restart with -js flag
nats-server -js
```

### Consumer Not Receiving Messages

```bash
# Check consumer info
nats consumer info EVENTS MY_CONSUMER

# Check for pending messages
nats stream view EVENTS

# Reset consumer position
nats consumer delete EVENTS MY_CONSUMER -f
nats consumer add EVENTS MY_CONSUMER --deliver=all
```

### Permission Denied on Scripts

```bash
# Make scripts executable
chmod +x scripts/*.sh
```

## 🔗 Integration with Eye-in-the-Sky

This demo serves as a proof-of-concept for integrating NATS JetStream into the Eye-in-the-Sky system. Future integration points:

1. **Agent Communication**: Durable message passing between agents
2. **Event Streaming**: Real-time agent status updates
3. **Task Distribution**: Work queue for agent tasks
4. **Audit Trail**: Persistent event log with replay
5. **Failure Recovery**: Agents can recover and replay missed messages

### Next Steps

1. Design event schema for agent messages
2. Implement NATS integration in MCP server
3. Add stream per agent or shared streams
4. Configure appropriate retention policies
5. Implement consumer groups for load balancing

## 📖 Resources

- [NATS Documentation](https://docs.nats.io/)
- [JetStream Concepts](https://docs.nats.io/jetstream/concepts)
- [Go Client Documentation](https://pkg.go.dev/github.com/nats-io/nats.go)
- [NATS CLI Cheat Sheet](https://docs.nats.io/using-nats/nats-tools/nats_cli)

## 📝 License

Part of the Eye-in-the-Sky project - see parent repository for license details.