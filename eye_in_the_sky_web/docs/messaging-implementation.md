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

## Message Grouping & WhatsApp-Style UI

**Date**: 2025-12-01

### Overview

Messages are now grouped by sender to create a clean, WhatsApp-style interface. Consecutive messages from the same sender appear as a single group with one sender label, improving readability and reducing visual clutter.

### Backend Grouping

**File**: `lib/eye_in_the_sky_web_web/live/agent_live/show.ex:336-373`

Messages are grouped server-side using `Enum.chunk_by/2`:

```elixir
defp group_and_serialize_messages(messages) when is_list(messages) do
  messages
  |> Enum.chunk_by(&{&1.sender_role, &1.direction})
  |> Enum.map(fn group ->
    first_message = List.first(group)
    last_message = List.last(group)

    %{
      sender_role: first_message.sender_role,
      direction: first_message.direction,
      provider: first_message.provider,
      timestamp: first_message.inserted_at,
      date: NaiveDateTime.to_date(first_message.inserted_at),
      status: last_message.status,  # Status of last message in group
      messages: Enum.map(group, fn msg ->
        %{
          id: msg.id,
          body: msg.body,
          inserted_at: msg.inserted_at
        }
      end)
    }
  end)
  |> add_date_separators()
end
```

**Key Features:**
- Groups by `{sender_role, direction}` tuple
- Stores sender metadata at group level (not per-message)
- Status represents the last message in the group
- Individual messages keep only: `id`, `body`, `inserted_at`

### Date Separators

**File**: `lib/eye_in_the_sky_web_web/live/agent_live/show.ex:364-373`

Automatically adds date separators when messages cross day boundaries:

```elixir
defp add_date_separators(groups) do
  groups
  |> Enum.with_index()
  |> Enum.map(fn {group, idx} ->
    prev_date = if idx > 0, do: Enum.at(groups, idx - 1).date, else: nil
    show_date = prev_date && group.date != prev_date

    Map.put(group, :show_date_separator, show_date)
  end)
end
```

Shows "Today", "Yesterday", or formatted dates between different days.

### Frontend Rendering

**File**: `assets/svelte/components/tabs/MessagesTab.svelte:252-304`

Messages render in grouped structure:

```svelte
{#each messages as messageGroup}
  <!-- Date separator -->
  {#if messageGroup.show_date_separator}
    <div class="date-separator">
      <span>{formatDate(messageGroup.date)}</span>
    </div>
  {/if}

  <!-- Message group -->
  <div class="message-group {messageGroup.direction}">
    <!-- Sender label (once per group) -->
    {#if messageGroup.direction === 'inbound'}
      <div class="group-sender-label">
        {messageGroup.sender_role === 'user' ? 'You' : 'Agent'}
        {#if messageGroup.provider}
          · {messageGroup.provider}
        {/if}
      </div>
    {/if}

    <!-- Multiple message bubbles in tight group -->
    <div class="message-bubbles-container">
      {#each messageGroup.messages as message, idx}
        <div class="message-bubble {messageGroup.direction}">
          <div class="message-text">{message.body}</div>

          <!-- Show time/status only on last message -->
          {#if idx === messageGroup.messages.length - 1}
            <div class="message-meta">
              <span class="message-time">{formatTime(message.inserted_at)}</span>
              <!-- Status icons for outbound messages -->
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </div>
{/each}
```

### Visual Design

**Spacing:**
- **Between groups** (different senders): 1.5rem (24px)
- **Within groups** (same sender): 0.25rem (4px)
- **Date separators**: 1.5rem margin

**Colors:**
- **Sent messages** (outbound): `#d9fdd3` (light green)
- **Received messages** (inbound): `white`
- **Dark mode sent**: `#005c4b` (dark green)
- **Dark mode received**: `#202c33` (dark gray)
- **Background**: `#efeae2` (WhatsApp beige)

**Status Indicators:**
- Pending: Clock icon
- Failed: X icon
- Delivered: Double checkmarks (WhatsApp-style)

### Data Structure Transformation

**Before** (flat array):
```json
[
  {"id": "1", "sender_role": "user", "direction": "outbound", "body": "How?", "status": "delivered"},
  {"id": "2", "sender_role": "user", "direction": "outbound", "body": "Why?", "status": "delivered"},
  {"id": "3", "sender_role": "agent", "direction": "inbound", "body": "Because...", "status": "delivered"}
]
```

**After** (grouped):
```json
[
  {
    "sender_role": "user",
    "direction": "outbound",
    "provider": "claude",
    "timestamp": "2025-12-01T10:00:00Z",
    "date": "2025-12-01",
    "show_date_separator": false,
    "status": "delivered",
    "messages": [
      {"id": "1", "body": "How?", "inserted_at": "..."},
      {"id": "2", "body": "Why?", "inserted_at": "..."}
    ]
  },
  {
    "sender_role": "agent",
    "direction": "inbound",
    "provider": "claude",
    "timestamp": "2025-12-01T10:00:45Z",
    "date": "2025-12-01",
    "show_date_separator": false,
    "status": "delivered",
    "messages": [
      {"id": "3", "body": "Because...", "inserted_at": "..."}
    ]
  }
]
```

### Testing Message Grouping

**IEx Example:**

```elixir
# Open IEx and spawn Claude CLI
port = Port.open(
  {:spawn_executable, "/opt/homebrew/bin/claude"},
  [
    :binary,
    :exit_status,
    :use_stdio,
    :stderr_to_stdout,
    {:args, ["--resume", "session-id", "-p", "test message", "--output-format", "stream-json"]},
    {:cd, "/path/to/project"}
  ]
)

# Check output
flush()
# Should see JSON stream with message exchanges
```

**UI Testing:**

1. Navigate to agent detail page → Messages tab
2. Send 3 consecutive messages → Should appear as **one group** with **one label**
3. Agent replies → New group created with agent label
4. Send messages on different days → Date separator appears
5. Verify spacing: tight within groups, loose between groups

### Provider Selector Enhancement

**File**: `assets/svelte/components/tabs/MessagesTab.svelte:360-395`

Provider selector now shows a badge with current selection:

```svelte
<div class="provider-selector">
  <div class="dropdown dropdown-top">
    <label class="btn btn-ghost btn-sm gap-2" title="Select AI Provider">
      <svg><!-- Monitor icon --></svg>
      <span class="badge badge-sm badge-primary">{selectedProvider}</span>
    </label>
    <ul class="dropdown-content menu">
      <!-- Dropdown items with checkmarks for selected option -->
    </ul>
  </div>
</div>
```

**Features:**
- Badge displays: "claude" or "openai"
- Checkmark appears next to selected provider in dropdown
- Placeholder updated to: "Send instruction to agent..."

### Performance Benefits

**Backend grouping advantages:**
- Groups calculated once server-side (not on every render)
- Reduced data transfer (metadata only sent once per group)
- Testable with ExUnit independently
- Consistent across all clients

**Before:** 100 messages = 100 sender labels
**After:** 100 messages = ~10-20 groups = 10-20 sender labels

### Call Sites Updated

All message loading uses `group_and_serialize_messages/1`:

1. **Line 257**: `load_tab_data(:messages, session_id)`
2. **Line 184**: `handle_event("send_message", ...)` - After sending
3. **Line 199**: `handle_info({:new_message, ...})` - Real-time updates
4. **Line 218**: `handle_info({:claude_output, ...})` - Claude output streaming

## Future Enhancements

1. **Acknowledgment system**: Implement retry logic with exponential backoff
2. **Task routing**: Use `events.task` for inter-agent task delegation
3. **Broadcast messages**: Implement empty `receiver_id` for announcements
4. **Message persistence**: Archive messages to separate table for long-term storage
5. **Read receipts**: Track when messages are viewed (per group)
6. **Typing indicators**: Real-time typing status via `events.protocol`
7. **File attachments**: Support file sharing via metadata + blob storage
8. **Message reactions**: Add emoji reactions to specific messages
9. **Message search**: Full-text search across grouped messages
10. **Collapsible groups**: Allow expanding/collapsing message groups for long conversations
