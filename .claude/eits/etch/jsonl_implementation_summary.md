# Eye in the Sky - JSONL Session Loading Implementation

## Overview
Implemented opcode-style JSONL-based session loading in Eye in the Sky. Sessions can now be loaded from files stored in `~/.claude/projects/{projectId}/{sessionId}.jsonl` instead of relying solely on database queries.

## Files Created

### 1. `/lib/eye_in_the_sky_web/messages/jsonl_storage.ex`
Elixir module handling all JSONL file operations.

**Key Functions:**
- `read_session_messages(project_id, session_id)` - Reads all messages from JSONL file
- `append_message(project_id, session_id, message_data)` - Appends single message to file
- `write_session_messages(project_id, session_id, messages)` - Bulk writes messages to file
- `get_session_file_path(project_id, session_id)` - Returns file path

**Features:**
- Parses JSON lines into Message structs
- Handles timestamp parsing (ISO 8601 and Unix timestamps)
- Auto-creates directories if needed
- Logging at each step
- Graceful fallback when files don't exist

## Files Modified

### 2. `/lib/eye_in_the_sky_web/messages.ex`
Updated Messages context to support JSONL loading.

**New Functions:**
- `list_messages_for_session(session_id, project_id)` - Loads from JSONL if project_id provided, falls back to database
- `list_recent_messages(session_id, limit, project_id)` - Loads recent messages from JSONL
- `append_to_jsonl(project_id, session_id, message_attrs)` - Appends message to JSONL file
- `write_session_to_jsonl(project_id, session_id)` - Bulk migrates session from database to JSONL
- `get_session_jsonl_path(project_id, session_id)` - Gets file path for a session

**Backward Compatibility:**
- Existing database functions remain unchanged
- New JSONL functions are additive
- Falls back to database if JSONL file not found

## Data Flow

### Loading Messages (JSONL)
```
Frontend (AgentMessagesPanel.svelte)
    ↓
Backend (Messages context)
    ↓
list_messages_for_session(session_id, project_id)
    ↓
JsonlStorage.read_session_messages(project_id, session_id)
    ↓
File.stream!(path)
    ↓
Parse each JSON line → Message struct
    ↓
Sort by inserted_at timestamp
    ↓
Return to frontend
```

### Appending Messages (JSONL)
```
Backend event handler (send_direct_message, etc.)
    ↓
Messages.append_to_jsonl(project_id, session_id, attrs)
    ↓
JsonlStorage.append_message(...)
    ↓
File.write(..., [:append])
    ↓
New line appended to JSONL file
```

## File Storage Format

**Location:** `~/.claude/projects/{projectId}/{sessionId}.jsonl`

**Example:**
```jsonl
{"id":"abc-123","session_id":"def-456","sender_role":"user","body":"Hello","direction":"outbound","inserted_at":"2025-01-15T10:30:00Z"}
{"id":"xyz-789","session_id":"def-456","sender_role":"agent","body":"Hi there!","direction":"inbound","inserted_at":"2025-01-15T10:30:05Z"}
```

## Integration with Eye in the Sky Frontend

The AgentMessagesPanel component can now load messages from JSONL files by passing `project_id` when calling the backend:

```elixir
# Old way (database only)
Messages.list_messages_for_session(session_id)

# New way (JSONL with fallback)
Messages.list_messages_for_session(session_id, project_id)
```

The frontend doesn't need to change - it receives Message structs either way.

## Migration Path

To migrate existing sessions from database to JSONL:

```elixir
# Write single session to JSONL
Messages.write_session_to_jsonl(project_id, session_id)

# Verify JSONL file exists
Path.expand("~/.claude/projects/#{project_id}/#{session_id}.jsonl")
|> File.exists?()
```

## Benefits

1. **Opcode Compatibility** - Uses same storage format as opcode
2. **Simple Format** - JSONL is human-readable and easy to debug
3. **Append-only** - Fast writes, perfect for streaming
4. **No Database Overhead** - Direct file I/O
5. **Backward Compatible** - Database queries still work
6. **Fallback Support** - Automatic fallback if file not found

## Next Steps

1. Update AgentMessagesPanel.svelte to pass `project_id` when loading messages
2. Update LiveView handlers to append to JSONL when messages are created
3. Add migration script to convert existing database messages to JSONL
4. Monitor file I/O performance with large session files

## Testing

To test the implementation:

```elixir
# Load from JSONL
messages = Messages.list_messages_for_session("session-id", "project-id")

# Append new message
Messages.append_to_jsonl("project-id", "session-id", %{
  id: Ecto.UUID.generate(),
  sender_role: "user",
  body: "Test message",
  direction: "outbound",
  inserted_at: DateTime.utc_now()
})

# Verify file
path = Messages.get_session_jsonl_path("project-id", "session-id")
File.exists?(path)
```

## Implementation Notes

- Uses Elixir's `File.stream!` for memory-efficient line-by-line reading
- Handles both ISO 8601 and Unix timestamp formats
- Automatically generates UUIDs for messages without IDs
- Maintains insertion order by sorting on `inserted_at`
- Logs all operations for debugging
