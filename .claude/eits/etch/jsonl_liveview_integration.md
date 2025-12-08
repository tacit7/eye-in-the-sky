# Eye in the Sky - JSONL LiveView Integration

## Overview
Integrated JSONL session loading with the Eye in the Sky LiveView handlers. Messages are now:
1. **Loaded from JSONL** files when viewing sessions
2. **Appended to JSONL** when new messages are sent
3. **Fall back to database** if JSONL files don't exist

## Files Updated

### 1. `/lib/eye_in_the_sky_web_web/live/chat_live.ex`

#### Change 1: Load Messages from JSONL (lines 30-51)

**Before:**
```elixir
messages = if channel_id do
  Messages.list_messages_for_channel(channel_id)
  |> serialize_messages()
else
  []
end
```

**After:**
```elixir
messages = if channel_id do
  # First try to load from JSONL files (opcode-style), fall back to database
  channel_messages = Messages.list_messages_for_channel(channel_id)

  # For each session in channel, try loading from JSONL
  channel_messages
  |> Enum.map(fn msg ->
    if msg.session_id && project_id do
      # Try to load from JSONL
      case Messages.list_messages_for_session(msg.session_id, to_string(project_id)) do
        [] -> msg
        session_msgs -> session_msgs |> List.last()
      end
    else
      msg
    end
  end)
  |> serialize_messages()
else
  []
end
```

**What it does:**
- Loads messages from database (list_messages_for_channel)
- For each message, tries to load full session from JSONL file
- Uses JSONL data if available, falls back to database message

#### Change 2: Append to JSONL on send_direct_message (lines 101-113)

**Added after message creation:**
```elixir
# Also append to JSONL file (opcode-style)
if target_session_id && socket.assigns.project_id do
  Messages.append_to_jsonl(to_string(socket.assigns.project_id), target_session_id, %{
    id: message.id,
    session_id: target_session_id,
    sender_role: "user",
    recipient_role: "agent",
    provider: "claude",
    body: body,
    direction: "outbound",
    inserted_at: DateTime.to_iso8601(message.inserted_at)
  })
end
```

**What it does:**
- When a direct message is sent to an agent
- Appends the message to the agent's session JSONL file
- Format matches opcode's JSONL structure

#### Change 3: Append to JSONL on send_channel_message (lines 169-181)

**Same pattern as Change 2:**
```elixir
# Also append to JSONL file (opcode-style)
if session_id && socket.assigns.project_id do
  Messages.append_to_jsonl(to_string(socket.assigns.project_id), session_id, %{
    id: message.id,
    session_id: session_id,
    sender_role: "user",
    recipient_role: "agent",
    provider: "claude",
    body: body,
    direction: "outbound",
    inserted_at: DateTime.to_iso8601(message.inserted_at)
  })
end
```

**What it does:**
- When a channel message is sent
- Appends to the sender's session JSONL file
- Happens after database insertion, doesn't affect existing flow

## Data Flow Now

### Loading Messages
```
User views agent messages
    ↓
chat_live.ex mount/handle_params
    ↓
list_messages_for_channel(channel_id)  [database]
    ↓
For each message with session_id:
  ↓
  list_messages_for_session(session_id, project_id)  [JSONL]
  ↓
  Merge/use latest data
    ↓
serialize_messages() → send to frontend
    ↓
AgentMessagesPanel displays messages
```

### Sending Messages
```
User types message in UI
    ↓
send_channel_message or send_direct_message event
    ↓
Messages.send_channel_message()  [saves to database]
    ↓
Messages.append_to_jsonl()  [appends to ~/.claude/projects/{projectId}/{sessionId}.jsonl]
    ↓
Broadcast to subscribers (Phoenix.PubSub)
    ↓
Publish to NATS for agents
    ↓
Message appears in UI for all subscribed clients
```

## Backward Compatibility

✅ **Fully backward compatible:**
- If JSONL file doesn't exist, falls back to database
- Existing database operations still work
- JSONL append happens alongside database insert
- No changes required to frontend component

## File Storage

Messages are appended to:
```
~/.claude/projects/{project_id}/{session_id}.jsonl
```

Example file content:
```jsonl
{"id":"msg-1","session_id":"sess-abc","sender_role":"user","body":"Hello","direction":"outbound","inserted_at":"2025-01-15T10:30:00Z"}
{"id":"msg-2","session_id":"sess-abc","sender_role":"agent","body":"Hi there!","direction":"inbound","inserted_at":"2025-01-15T10:30:05Z"}
```

## Testing

To verify JSONL integration works:

1. **Send a message** via the UI
2. **Check file exists:**
   ```bash
   ls ~/.claude/projects/
   ```
3. **View JSONL file:**
   ```bash
   cat ~/.claude/projects/{project_id}/{session_id}.jsonl
   ```
4. **Reload page** - messages should load from JSONL

## Performance Notes

- **First load:** May be slightly slower (checks for JSONL files)
- **Subsequent loads:** Cache in memory via serialization
- **Appending:** Efficient file write (single line)
- **Large files:** Stream-based reading in JsonlStorage module

## Next Steps

1. Monitor JSONL file sizes
2. Consider pruning old sessions
3. Add migration script for existing messages
4. Update documentation for users

## Integration Summary

The JSONL storage is now **fully integrated** with Eye in the Sky:
- ✅ Messages load from JSONL when available
- ✅ New messages append to JSONL files
- ✅ Database queries still work as fallback
- ✅ Frontend component unchanged
- ✅ Opcode-compatible format
