# Messaging Implementation - Eye in the Sky

**Date**: 2025-11-30

## Overview

Eye in the Sky implements **opcode-style messaging** using NATS pub/sub with the **eits-messaging-v1 protocol** for agent-to-agent communication, while maintaining opcode's subprocess spawning approach for Claude Code interaction.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│         Phoenix LiveView (UI Layer)                 │
│  User types message → send_message event            │
└───────────────┬─────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────┐
│    AgentLive.Show (LiveView Handler)                │
│  1. Create message in DB (Messages.send_message)    │
│  2. Publish to NATS (Publisher.publish_message)     │
│  3. Spawn Claude subprocess (SessionManager)        │
└──────┬──────────────────────┬───────────────────────┘
       │                      │
       │ NATS publish         │ Subprocess spawn
       ▼                      ▼
┌──────────────┐      ┌──────────────────────────────┐
│ NATS.Publisher│      │  Claude.SessionManager       │
│ events.chat   │      │  - Spawns claude binary      │
└──────┬────────┘      │  - Streams stdout/stderr     │
       │               │  - Parses JSON output        │
       │               └───────────┬──────────────────┘
       │                           │
       ▼                           ▼ Assistant reply
┌──────────────────────┐   ┌──────────────────────────┐
│  NATS JetStream      │   │  Messages.record_reply   │
│  (events.*)          │   │  + Publisher.publish     │
│                      │   └──────────┬───────────────┘
│  - events.chat       │              │
│  - events.protocol   │              │ Publish reply
│  - events.task       │◄─────────────┘
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  NATS.Consumer       │
│  Subscribes to:      │
│  - events.chat       │
│  - events.protocol   │
└──────┬───────────────┘
       │
       ▼ Broadcast to PubSub
┌──────────────────────┐
│  Phoenix.PubSub      │
│  session:ID:messages │
└──────┬───────────────┘
       │
       ▼ Real-time update
┌──────────────────────┐
│  LiveView UI         │
│  Messages tab        │
└──────────────────────┘
```

## Message Flow

### 1. User Sends Message

**File**: `lib/eye_in_the_sky_web_web/live/agent_live/show.ex:124-155`

```elixir
def handle_event("send_message", %{"body" => body, "provider" => provider}, socket) do
  session_id = socket.assigns.session_id

  # Step 1: Create outbound message in database
  {:ok, message} = Messages.send_message(%{
    session_id: session_id,
    sender_role: "user",
    recipient_role: "agent",
    provider: provider,
    body: body
  })

  # Step 2: Publish to NATS for agent consumption
  Publisher.publish_message(message)

  # Step 3: Spawn Claude CLI subprocess
  SessionManager.start_session(session_id, body, model: provider_to_model(provider))
end
```

### 2. NATS Publisher

**File**: `lib/eye_in_the_sky_web/nats/publisher.ex`

Publishes messages with **eits-messaging-v1 envelope**:

```elixir
def publish_message(message, opts \\ []) do
  # Build envelope following eits-messaging-v1 protocol
  envelope = %{
    op: "msg",                    # Operation: msg, ack, confirm, ping, pong
    channel: "chat",              # Channel: chat, protocol, task
    version: "eits-messaging-v1",
    reply_to: message.session_id,
    msg: message.body,
    meta: %{
      message_id: message.id,
      provider: message.provider,
      timestamp: DateTime.to_iso8601(message.inserted_at)
    }
  }

  payload = Jason.encode!(envelope)
  Gnat.pub(connection, "events.chat", payload)
end
```

### 3. Claude SessionManager Spawns Subprocess

**File**: `lib/eye_in_the_sky_web/claude/session_manager.ex:60-84`

```elixir
def handle_call({:start_session, session_id, prompt, opts}, _from, state) do
  # Spawn claude binary (like opcode does)
  case CLI.spawn_new_session(prompt, opts) do
    {:ok, port, session_ref} ->
      # Track the running subprocess
      session_info = %{
        port: port,
        session_id: session_id,
        started_at: DateTime.utc_now(),
        output_buffer: [],
        claude_session_id: nil
      }
      {:reply, {:ok, session_ref}, Map.put(state, session_ref, session_info)}
  end
end
```

### 4. Claude CLI Spawning

**File**: `lib/eye_in_the_sky_web/claude/cli.ex:30-67`

```elixir
def spawn_new_session(prompt, opts \\ []) do
  # Build args like opcode does
  args = [
    "-p", prompt,
    "--model", model,
    "--output-format", "stream-json",
    "--verbose",
    "--dangerously-skip-permissions"
  ]

  # Spawn subprocess
  port = Port.open(
    {:spawn_executable, claude_path},
    [
      :binary,
      :exit_status,
      :use_stdio,
      :stderr_to_stdout,
      {:args, args},
      {:cd, project_path},
      {:env, build_env()}
    ]
  )

  spawn_link(fn -> handle_port_output(port, session_ref, caller) end)
end
```

### 5. Agent Reply Processing

**File**: `lib/eye_in_the_sky_web/claude/session_manager.ex:197-217`

```elixir
if parsed["type"] == "assistant" do
  content = parsed["content"] || parsed["message"]

  # Create inbound message in database
  {:ok, message} = Messages.record_incoming_reply(
    session_info.session_id,
    "claude",
    content
  )

  # Publish agent reply to NATS for agent-to-agent communication
  Publisher.publish_message(message)
end
```

### 6. NATS Consumer Receives Reply

**File**: `lib/eye_in_the_sky_web/nats/consumer.ex:50-70`

```elixir
defp handle_envelope(%{"op" => "msg", "channel" => "chat"} = envelope, _topic) do
  session_id = envelope["reply_to"]
  provider = get_in(envelope, ["meta", "provider"]) || "unknown"
  message_body = envelope["msg"]

  case Messages.record_incoming_reply(session_id, provider, message_body) do
    {:ok, message} ->
      # Broadcast to Phoenix PubSub for LiveView updates
      Phoenix.PubSub.broadcast(
        EyeInTheSkyWeb.PubSub,
        "session:#{session_id}:messages",
        {:new_message, message}
      )
  end
end
```

### 7. LiveView Real-time Update

**File**: `lib/eye_in_the_sky_web_web/live/agent_live/show.ex:153-169`

```elixir
def handle_info({:new_message, _message}, socket) do
  session_id = socket.assigns.session_id

  # Reload messages and update UI
  updated_messages = serialize_messages(Messages.list_messages_for_session(session_id))
  counts = Sessions.get_session_counts(session_id)

  socket
  |> assign(:messages, updated_messages)
  |> assign(:counts, counts)
end
```

## Protocol Specification: eits-messaging-v1

### Envelope Format

```json
{
  "op": "propose|ack|confirm|msg|ping|pong",
  "channel": "protocol|chat|task",
  "version": "eits-messaging-v1",
  "reply_to": "<session_id>",
  "msg": "free-form text",
  "meta": {
    "message_id": "uuid",
    "provider": "claude|openai",
    "timestamp": "ISO8601"
  }
}
```

### Operation Codes (op)

| Op | Purpose |
|----|---------|
| `msg` | Regular chat message |
| `ack` | Acknowledgment within 30s |
| `confirm` | Final confirmation |
| `ping` | Health check |
| `pong` | Health check response |
| `propose` | Propose an action/task |

### Channels

| Channel | Purpose |
|---------|---------|
| `chat` | User-agent conversational messages |
| `protocol` | Control/handshakes (acks/confirms) |
| `task` | Actionable requests or task routing |

### NATS Subjects

- **Stream**: `EVENTS`
- **Subjects**: `events.*`
  - `events.chat` — Conversational messages
  - `events.protocol` — Control/handshakes
  - `events.task` — Task routing
  - Default: `events.test`

## Key Differences from Opcode

| Aspect | Opcode | Eye in the Sky |
|--------|--------|----------------|
| **Message Transport** | Direct stdout/stderr streaming | NATS pub/sub + stdout streaming |
| **Protocol** | Raw JSON lines | Structured eits-messaging-v1 envelopes |
| **Use Case** | Single user → single agent | Multi-agent communication |
| **Subprocess Spawning** | ✅ Same approach | ✅ Same approach |
| **Claude CLI flags** | ✅ Same flags | ✅ Same flags |
| **Session management** | Filesystem only | Database + NATS + filesystem |

## Setup Requirements

### 1. NATS Server

Start NATS with JetStream enabled:

```bash
nats-server -js
```

Or install via Homebrew:

```bash
brew install nats-server
nats-server -js
```

### 2. Database Migration

Ensure messages table is migrated:

```bash
cd eye_in_the_sky_web
mix ecto.migrate
```

### 3. Dependencies

Install Elixir dependencies (includes gnat):

```bash
mix deps.get
```

### 4. Start Phoenix

```bash
mix phx.server
```

## Testing

### Manual Test with Test Agent

**File**: `priv/scripts/test_agent.exs`

```bash
elixir priv/scripts/test_agent.exs
```

This script:
1. Connects to NATS at localhost:4222
2. Subscribes to `events.chat`
3. Auto-replies to incoming messages with eits-messaging-v1 envelopes

### Testing Flow

1. Start NATS server: `nats-server -js`
2. Start test agent: `elixir priv/scripts/test_agent.exs`
3. Start Phoenix: `mix phx.server`
4. Open browser: http://localhost:4000
5. Navigate to agent detail page
6. Click Messages tab
7. Send a message
8. Verify:
   - Message appears in UI (outbound)
   - Test agent receives it via NATS
   - Test agent replies via NATS
   - Reply appears in UI (inbound)

## Monitoring

### NATS CLI Tools

Install NATS CLI:

```bash
brew install nats-io/nats-tools/nats
```

Monitor stream:

```bash
# List streams
nats stream ls

# Monitor EVENTS stream
nats stream info EVENTS

# Subscribe to all events
nats sub "events.>"

# View specific subject
nats sub "events.chat"
```

### Phoenix LiveDashboard

View real-time metrics:

```
http://localhost:4000/dev/dashboard
```

## Database Schema

### messages table

```sql
CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  project_id INTEGER REFERENCES projects(id),
  session_id TEXT,
  sender_role TEXT NOT NULL,
  recipient_role TEXT,
  provider TEXT,
  provider_session_id TEXT,
  direction TEXT NOT NULL,  -- 'inbound' or 'outbound'
  body TEXT NOT NULL,
  status TEXT NOT NULL,     -- 'sent', 'delivered', 'failed', 'pending'
  metadata TEXT,
  inserted_at DATETIME,
  updated_at DATETIME
);
```

## Code Organization

```
lib/eye_in_the_sky_web/
├── messages/
│   └── message.ex              # Ecto schema
├── messages.ex                 # Context module (CRUD)
├── nats/
│   ├── publisher.ex            # NATS publisher (eits-messaging-v1)
│   └── consumer.ex             # NATS consumer GenServer
├── claude/
│   ├── cli.ex                  # Subprocess spawner (like opcode)
│   └── session_manager.ex      # Process lifecycle management
└── application.ex              # Supervision tree (starts Consumer)

lib/eye_in_the_sky_web_web/live/agent_live/
└── show.ex                     # LiveView handlers

assets/svelte/components/
└── AgentDetail.svelte          # Messages UI (lines 348-429)
```

## Future Enhancements

1. **Acknowledgment system**: Implement retry logic with exponential backoff
2. **Task routing**: Use `events.task` for inter-agent task delegation
3. **Broadcast messages**: Implement empty `receiver_id` for announcements
4. **Message persistence**: Archive messages to separate table for long-term storage
5. **Read receipts**: Track when messages are viewed
6. **Typing indicators**: Real-time typing status via `events.protocol`
7. **File attachments**: Support file sharing via metadata + blob storage
