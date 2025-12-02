# Eye in the Sky - Claude Code Hooks Integration

This document describes the hook system for integrating Eye in the Sky agent tracking with Claude Code sessions.

## Overview

Claude Code hooks allow external commands to run at specific events (SessionStart, SessionEnd, PreCompact, etc.). We use these hooks to automatically track agent activity in the Eye in the Sky database.

## CLI Commands (Completed)

All CLI commands are available at `bin/eits-cli`:

### 1. start-session
**Usage:** `eits-cli start-session <session-id> [options]`

**Returns:** Agent UUID (plain text by default, JSON with `--json`)

**Example:**
```bash
AGENT_ID=$(eits-cli start-session "$SESSION_ID")
```

**Options:**
- `--description` - Session description (default: "Auto-started session")
- `--worktree` - Git worktree path (default: current directory)
- `--project` - Project name
- `--parent-agent-id` - Parent agent ID for subagents

### 2. get-latest-agent
**Usage:** `eits-cli get-latest-agent <session-id>`

**Returns:** Most recent agent UUID for a session

**Example:**
```bash
AGENT_ID=$(eits-cli get-latest-agent "$SESSION_ID")
```

### 3. update-status
**Usage:** `eits-cli update-status --agent-id <id> --status <status> [--task <task>]`

**Status values:** `active`, `working`, `idle`, `completed`, `failed`

**Example:**
```bash
eits-cli update-status --agent-id "$AGENT_ID" --status working
```

### 4. log-action
**Usage:** `eits-cli log-action --agent-id <id> --type <type> --desc <description> [--details <json>]`

**Action types:** `task_start`, `file_operation`, `git_commit`, `status_update`

**Example:**
```bash
eits-cli log-action --agent-id "$AGENT_ID" --type task_start --desc "Starting authentication implementation"
```

### 5. log-commits
**Usage:** `eits-cli log-commits --agent-id <id> --hashes <hash1,hash2> [--messages <msg1,msg2>]`

**Example:**
```bash
eits-cli log-commits --agent-id "$AGENT_ID" --hashes "abc123,def456"
```

### 6. log-compaction
**Usage:** `eits-cli log-compaction --agent-id <id> --session-id <id> [--transcript <path>] [--trigger <manual|auto>] [--summary <text>]`

**Example:**
```bash
eits-cli log-compaction --agent-id "$AGENT_ID" --session-id "$SESSION_ID" --transcript "$TRANSCRIPT_PATH" --trigger "auto"
```

### 7. end-session
**Usage:** `eits-cli end-session --agent-id <id> [--status <completed|failed>] [--summary <text>]`

**Example:**
```bash
eits-cli end-session --agent-id "$AGENT_ID" --status completed
```

## Hook Configuration

### Hook Input Structure

All hooks receive JSON input via `$INPUT` environment variable:

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../session.jsonl",
  "cwd": "/Users/.../project",
  "hook_event_name": "SessionStart",
  "trigger": "manual|auto",
  "custom_instructions": "<string>"
}
```

### Extracting Agent ID

**Option 1: Database lookup (always works)**
```bash
AGENT_ID=$(eits-cli get-latest-agent "$(jq -r '.session_id' <<<\"$INPUT\")")
```

**Option 2: From transcript (after i-start-session MCP call)**
```bash
AGENT_ID=$(jq -r 'select(.message?.content[]?.type == "tool_result") | .message.content[]?.content? // empty | fromjson? | select(.agent_id != null) | [.agent_id] | @tsv' <"$(jq -r '.transcript_path' <<<\"$INPUT\")" | tail -1)
```

**Option 3: Hybrid (try transcript, fallback to database)**
```bash
AGENT_ID=$(jq -r 'select(.message?.content[]?.type == "tool_result") | .message.content[]?.content? // empty | fromjson? | select(.agent_id != null) | [.agent_id] | @tsv' <"$(jq -r '.transcript_path' <<<\"$INPUT\")" 2>/dev/null | tail -1)
[ -z "$AGENT_ID" ] && AGENT_ID=$(eits-cli get-latest-agent "$(jq -r '.session_id' <<<\"$INPUT\")")
```

## Hook Examples

### SessionStart Hook (✅ Completed)

Creates agent and shows macOS notification:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(/Users/urielmaldonado/projects/eye-in-the-sky/bin/eits-cli start-session \"$(jq -r '.session_id' <<<\"$INPUT\")\") && SESSION_ID=$(jq -r '.session_id' <<<\"$INPUT\") && osascript -e \"display notification \\\"Agent: ${AGENT_ID:0:8}\\nSession: ${SESSION_ID:0:8}\\\" with title \\\"Eye in the Sky\\\" subtitle \\\"Session Started\\\" sound name \\\"Glass\\\"\" &",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

**Triggers on:**
- `startup` - Initial session launch
- `resume` - Resuming with `--resume`, `--continue`, or `/resume`
- `clear` - After `/clear` command
- `compact` - After conversation compaction

**What it does:**
- Creates new agent with auto-generated UUID
- Shows notification with truncated agent/session IDs
- Plays "Glass" sound

### SessionEnd Hook (🔨 TODO: Test)

Marks agent as completed:

```json
{
  "hooks": {
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(eits-cli get-latest-agent \"$(jq -r '.session_id' <<<\"$INPUT\")\") && [ -n \"$AGENT_ID\" ] && eits-cli end-session --agent-id \"$AGENT_ID\" --status completed",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

**What it does:**
- Looks up agent by session_id
- Marks agent status as "completed"
- Updates last_activity timestamp

### PreCompact Hook (🔨 TODO: Test)

Logs conversation compaction event:

```json
{
  "hooks": {
    "PreCompact": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(eits-cli get-latest-agent \"$(jq -r '.session_id' <<<\"$INPUT\")\") && [ -n \"$AGENT_ID\" ] && eits-cli log-compaction --agent-id \"$AGENT_ID\" --session-id \"$(jq -r '.session_id' <<<\"$INPUT\")\" --transcript \"$(jq -r '.transcript_path' <<<\"$INPUT\")\" --trigger \"$(jq -r '.trigger' <<<\"$INPUT\")\" --summary \"$(jq -r '.custom_instructions // \"\"' <<<\"$INPUT\")\"",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

**What it does:**
- Logs when conversation gets compacted
- Records transcript file path and size
- Counts messages in JSONL file
- Tracks manual vs auto compaction

### Stop Hook (🔨 TODO: Implement & Test)

Mark agent as idle when Claude stops responding:

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(eits-cli get-latest-agent \"$(jq -r '.session_id' <<<\"$INPUT\")\") && [ -n \"$AGENT_ID\" ] && eits-cli update-status --agent-id \"$AGENT_ID\" --status idle",
            "timeout": 3
          }
        ]
      }
    ]
  }
}
```

### UserPromptSubmit Hook (🔨 TODO: Implement & Test)

Mark agent as working when user sends a message:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(eits-cli get-latest-agent \"$(jq -r '.session_id' <<<\"$INPUT\")\") && [ -n \"$AGENT_ID\" ] && eits-cli update-status --agent-id \"$AGENT_ID\" --status working",
            "timeout": 3
          }
        ]
      }
    ]
  }
}
```

### PostToolUse Hook - Git Commits (🔨 TODO: Implement & Test)

Automatically log git commits:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "if echo \"$(jq -r '.tool_input.command' <<<\"$INPUT\")\" | grep -q '^git commit'; then AGENT_ID=$(eits-cli get-latest-agent \"$(jq -r '.session_id' <<<\"$INPUT\")\") && [ -n \"$AGENT_ID\" ] && HASH=$(git log -1 --format=%h 2>/dev/null) && [ -n \"$HASH\" ] && eits-cli log-commits --agent-id \"$AGENT_ID\" --hashes \"$HASH\"; fi",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

## Complete Hook Configuration

Add this to `.claude/settings.json` or `.claude/settings.local.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(/Users/urielmaldonado/projects/eye-in-the-sky/bin/eits-cli start-session \"$(jq -r '.session_id' <<<\"$INPUT\")\") && SESSION_ID=$(jq -r '.session_id' <<<\"$INPUT\") && osascript -e \"display notification \\\"Agent: ${AGENT_ID:0:8}\\nSession: ${SESSION_ID:0:8}\\\" with title \\\"Eye in the Sky\\\" subtitle \\\"Session Started\\\" sound name \\\"Glass\\\"\" &",
            "timeout": 5
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(eits-cli get-latest-agent \"$(jq -r '.session_id' <<<\"$INPUT\")\") && [ -n \"$AGENT_ID\" ] && eits-cli end-session --agent-id \"$AGENT_ID\" --status completed",
            "timeout": 5
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "AGENT_ID=$(eits-cli get-latest-agent \"$(jq -r '.session_id' <<<\"$INPUT\")\") && [ -n \"$AGENT_ID\" ] && eits-cli log-compaction --agent-id \"$AGENT_ID\" --session-id \"$(jq -r '.session_id' <<<\"$INPUT\")\" --transcript \"$(jq -r '.transcript_path' <<<\"$INPUT\")\" --trigger \"$(jq -r '.trigger' <<<\"$INPUT\")\" --summary \"$(jq -r '.custom_instructions // \"\"' <<<\"$INPUT\")\"",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

## TODO: Remaining Work

### 1. Testing (HIGH PRIORITY)
- [ ] Test SessionStart hook in actual Claude Code session
- [ ] Verify agent creation and notification display
- [ ] Test SessionEnd hook
- [ ] Test PreCompact hook with manual and auto compaction
- [ ] Test agent_id extraction from transcript
- [ ] Test database lookup fallback

### 2. Optional Hooks (MEDIUM PRIORITY)
- [ ] Implement and test Stop hook (agent goes idle)
- [ ] Implement and test UserPromptSubmit hook (agent goes working)
- [ ] Implement and test PostToolUse git commit hook

### 3. Error Handling (MEDIUM PRIORITY)
- [ ] Add error handling for failed eits-cli commands
- [ ] Add logging for hook failures
- [ ] Add retry logic for transient failures

### 4. Notifications (LOW PRIORITY)
- [ ] Improve notification formatting
- [ ] Add notification for session end
- [ ] Add notification for compaction events
- [ ] Make notification sound configurable

### 5. Documentation (LOW PRIORITY)
- [ ] Add troubleshooting section
- [ ] Document common failure modes
- [ ] Add examples of viewing tracked data
- [ ] Create video/screenshots of hooks in action

## Database Schema

Hooks populate these database tables:

**agents**
- Tracks each Claude Code session
- Stores agent_id (UUID), session_id, status, worktree path
- Updated by: start-session, update-status, end-session

**actions**
- Logs agent activities
- Stores action_type, description, timestamp
- Updated by: log-action

**commits**
- Tracks git commits
- Stores commit_hash, commit_message, timestamp
- Updated by: log-commits

**compactions**
- Tracks conversation compactions
- Stores transcript path, file size, message count, trigger type
- Updated by: log-compaction

## Viewing Tracked Data

### Using the TUI Dashboard
```bash
./bin/eye-ui
```

### Direct Database Queries
```bash
# List all agents
sqlite3 ~/.config/eye-in-the-sky/agents.db "SELECT id, status, session_id, created_at FROM agents ORDER BY created_at DESC LIMIT 10;"

# View agent actions
sqlite3 ~/.config/eye-in-the-sky/agents.db "SELECT timestamp, action_type, description FROM actions WHERE agent_id = '<agent-id>' ORDER BY timestamp DESC;"

# View compaction history
sqlite3 ~/.config/eye-in-the-sky/agents.db "SELECT compacted_at, trigger, summary, message_count FROM compactions WHERE agent_id = '<agent-id>';"
```

## Troubleshooting

### Agent ID not found
**Problem:** `get-latest-agent` returns empty
**Solution:** Check if session_id matches what's in database
```bash
sqlite3 ~/.config/eye-in-the-sky/agents.db "SELECT id, session_id FROM agents WHERE session_id LIKE '%<partial-id>%';"
```

### Hook timeouts
**Problem:** Hook exceeds timeout and gets killed
**Solution:** Increase timeout in hook configuration or optimize command

### Database locked
**Problem:** `database is locked` error
**Solution:** Close TUI dashboard or other processes accessing the database

### Notification not showing
**Problem:** osascript notification doesn't appear
**Solution:** Check System Preferences → Notifications → Script Editor permissions

## Future Enhancements

- [ ] Add `save-context` CLI command (tracked in TaskWarrior #41)
- [ ] Implement context auto-save in PreCompact hook
- [ ] Add token usage tracking (log-session-cost command)
- [ ] Create dashboard view filtered by session_id
- [ ] Add webhook support for external integrations
- [ ] Support for multiple concurrent sessions
