# Claude Code Hooks: Complete Guide

## Overview

Claude Code hooks are shell commands that execute in response to specific events during a Claude Code session. They enable custom automation, logging, and integration with external systems.

## Hook Architecture

### Hook Execution Model

- Hooks are **synchronous** and block the Claude Code workflow until completion
- Each hook receives **JSON data via stdin** containing session and event information
- Hooks can access **environment variables** set by Claude Code
- Hook output to stdout/stderr is captured but not shown to the user by default
- Exit code 0 = success, non-zero = failure (may block operations depending on hook type)

### Hook Configuration Location

Hooks are configured in `.claude/settings.local.json`:

```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "/absolute/path/to/hook-script.sh"
          }
        ]
      }
    ]
  }
}
```

## JSON Input Format

### Base Structure

All hooks receive JSON via stdin with this base structure:

```json
{
  "session_id": "string",
  "transcript_path": "string",
  "cwd": "string",
  "permission_mode": "string",
  "hook_event_name": "string"
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Unique identifier for the Claude Code session |
| `transcript_path` | string | Absolute path to the conversation JSON file |
| `cwd` | string | Current working directory |
| `permission_mode` | string | One of: "default", "plan", "acceptEdits", "bypassPermissions" |
| `hook_event_name` | string | Name of the hook event (e.g., "PreToolUse", "PostToolUse") |

### Event-Specific Fields

#### PreToolUse / PostToolUse

Additional fields for tool execution hooks:

```json
{
  "tool_name": "string",
  "tool_use": {
    "name": "string",
    "id": "string",
    "input": { }
  }
}
```

#### UserPromptSubmit

Additional fields for user prompt submission:

```json
{
  "prompt": "string"
}
```

## Environment Variables

### Always Available

| Variable | Description | Example |
|----------|-------------|---------|
| `CLAUDE_PROJECT_DIR` | Absolute path to project root | `/Users/user/projects/my-app` |
| `CLAUDE_CODE_REMOTE` | Remote execution indicator | `"true"` or empty/unset |

### Hook-Specific

| Variable | Hook | Description |
|----------|------|-------------|
| `CLAUDE_ENV_FILE` | SessionStart | File path for persisting environment variables |

### NOT Provided by Claude Code

These must be managed by your hook scripts:

- Agent IDs (Eye-in-the-Sky specific)
- Custom tracking identifiers
- User-defined metadata

## Hook Types

### Available Hook Events

1. **SessionStart**: Fires when a new Claude Code session begins
2. **PreToolUse**: Fires before any tool is executed
3. **PostToolUse**: Fires after any tool completes
4. **UserPromptSubmit**: Fires when user submits a prompt

### Hook Matchers

The `matcher` field supports glob patterns:

```json
{
  "matcher": "*",           // Match all tools
  "matcher": "Read",        // Match specific tool
  "matcher": "mcp__*",      // Match all MCP tools
  "matcher": "Bash"         // Match Bash tool only
}
```

## Best Practices

### CRITICAL: Resilient Hook Design

**Hook scripts MUST NOT block Claude Code execution.** Hooks that crash or hang will prevent users from working.

**Required resilience patterns:**
1. **Never use `set -e`** - causes immediate exit on errors
2. **Timeout stdin reads** - stdin may be unavailable or empty
3. **Suppress errors** on external commands (jq, curl, osascript)
4. **Always exit 0** - never return non-zero exit codes
5. **Provide fallback values** for all extracted fields

### 1. Use Resilient Error Handling

```bash
#!/usr/bin/env bash
# RECOMMENDED: Allows graceful error handling
set -uo pipefail  # Undefined vars and pipe failures only

# AVOID in production hooks (crashes on any error):
# set -euo pipefail
```

**Why avoid `-e`?**
- Hooks block Claude Code if they crash
- stdin may be empty/unavailable (causes read to fail)
- External commands can fail unpredictably
- Better to log errors and continue than block the user

### 2. Read JSON Input with Timeout

**CRITICAL:** stdin is not guaranteed. Handle gracefully:

```bash
# BAD: Blocks forever if stdin unavailable
input_json=$(cat)

# GOOD: Timeout with graceful exit
if ! input_json=$(timeout 2 cat); then
  # stdin read failed or timed out
  exit 0  # Exit gracefully
fi

# Exit if input is empty
if [ -z "$input_json" ]; then
  exit 0
fi

# Extract fields with error suppression and fallbacks
session_id=$(jq -r '.session_id // "unknown"' <<< "$input_json" 2>/dev/null || echo "unknown")
tool_name=$(jq -r '.tool_name // "unknown"' <<< "$input_json" 2>/dev/null || echo "unknown")
```

**Key points:**
- `timeout 2 cat` prevents hanging (2 second max)
- Check for empty input before parsing
- `2>/dev/null` suppresses jq errors
- `|| echo "fallback"` provides defaults
- Always exit 0 at end of script

### 3. Handle Missing Fields Gracefully

```bash
# Use fallback values
session_id=$(jq -r '.session_id // "unknown"' <<< "$input_json")
event_type=$(jq -r '.hook_event_name // .type // "unknown"' <<< "$input_json")
```

### 4. Log to Files, Not Stdout

```bash
# Don't pollute stdout - Claude Code may capture it
LOG_FILE="/tmp/my-hook.log"
echo "Hook executed at $(date)" >> "$LOG_FILE"
```

### 5. Use Absolute Paths

```bash
# Relative to project root
"$CLAUDE_PROJECT_DIR/bin/my-tool"

# Or use absolute paths
/usr/local/bin/my-tool
```

### 6. Make Scripts Executable

```bash
chmod +x .claude/hooks/my-hook.sh
```

## Common Patterns

### Pattern 1: Logging Tool Usage

```bash
#!/usr/bin/env bash
set -euo pipefail

input_json=$(cat)
session_id=$(jq -r '.session_id' <<< "$input_json")
tool_name=$(jq -r '.tool_name // "unknown"' <<< "$input_json")
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LOG_FILE="$CLAUDE_PROJECT_DIR/logs/tool-usage.log"
mkdir -p "$(dirname "$LOG_FILE")"

echo "[$timestamp] session=$session_id tool=$tool_name" >> "$LOG_FILE"
```

### Pattern 2: Session-to-Agent Mapping

When you need to map Claude Code session IDs to your own identifiers (e.g., Eye-in-the-Sky agent IDs):

**Why JSON mapping over database queries:**
- Benchmarked as **6.9% faster** than SQLite (15.1ms vs 16.2ms avg)
- More consistent performance (lower variance)
- No database connection overhead
- No locking concerns
- Simpler to update
- Works even if database is busy

```bash
#!/usr/bin/env bash
set -uo pipefail

# Read JSON with timeout
if ! input_json=$(timeout 2 cat); then
  exit 0
fi

session_id=$(jq -r '.session_id // "unknown"' <<< "$input_json" 2>/dev/null || echo "unknown")

# Look up agent ID from mapping file (faster than SQLite query)
MAPPING_FILE="$CLAUDE_PROJECT_DIR/.claude/hooks/session_agent_map.json"
if [ -f "$MAPPING_FILE" ]; then
  agent_id=$(jq -r --arg sid "$session_id" '.[$sid] // "unknown"' "$MAPPING_FILE" 2>/dev/null || echo "unknown")
else
  agent_id="unknown"
fi

# Use both IDs for your custom logic
echo "Session: $session_id, Agent: $agent_id"
```

**Mapping file format** (`.claude/hooks/session_agent_map.json`):

```json
{
  "97c212de-24e0-4c8d-b16a-876e406b7c14": "2743c649-18ab-422d-bdbe-bb3b59d2c431",
  "1a398965-1f97-4329-802d-0bbe0d923564": "354ab567-5a8b-4259-8f25-9ed68d96e7d8"
}
```

**When to update:**
- Add new entry when starting a new session with your tracking system
- Remove entries for completed sessions to keep file small
- File permissions: `chmod 600` (only owner can read/write)

**Alternative considered (not recommended):**
```bash
# SQLite query - works but slower and more complex
agent_id=$(sqlite3 "$DB_PATH" "SELECT id FROM agents WHERE session_id = '$session_id' LIMIT 1")
```
Benchmark showed this adds 1.1ms overhead per lookup and requires database access.

### Pattern 3: macOS Notifications

```bash
#!/usr/bin/env bash
set -euo pipefail

input_json=$(cat)
tool_name=$(jq -r '.tool_name // "unknown"' <<< "$input_json")

# Show notification
osascript -e "display notification \"Tool: $tool_name\" with title \"Claude Hook\""
```

### Pattern 4: Conditional Execution

Only run hook for specific tools:

```bash
#!/usr/bin/env bash
set -euo pipefail

input_json=$(cat)
tool_name=$(jq -r '.tool_name // "unknown"' <<< "$input_json")

# Only process git-related tools
if [[ "$tool_name" == "Bash" ]]; then
  tool_input=$(jq -r '.tool_use.input.command // ""' <<< "$input_json")
  if [[ "$tool_input" =~ ^git ]]; then
    # Log git commands
    echo "Git command: $tool_input" >> "$CLAUDE_PROJECT_DIR/logs/git.log"
  fi
fi
```

## Debugging Hooks

### Debug Hook Template

Create a debug hook to see exactly what data is passed:

```bash
#!/usr/bin/env bash

LOG_FILE="/tmp/claude-hook-debug.log"

echo "=== Hook execution at $(date) ===" >> "$LOG_FILE"
echo "Working directory: $(pwd)" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

echo "Environment variables:" >> "$LOG_FILE"
env | sort >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

echo "STDIN content:" >> "$LOG_FILE"
cat >> "$LOG_FILE"
echo "" >> "$LOG_FILE"
echo "=== End ===" >> "$LOG_FILE"
```

### Common Issues

#### 1. Hooks Not Running

**Symptoms**: Hook scripts don't execute, no logs generated

**Causes**:
- Script not executable (`chmod +x` required)
- Invalid path in settings.local.json
- Syntax error in JSON configuration
- Settings not reloaded (requires Claude Code restart)

**Solution**:
```bash
# Make executable
chmod +x .claude/hooks/my-hook.sh

# Verify path is absolute
ls -la /absolute/path/to/hook.sh

# Restart Claude Code to reload settings
```

#### 2. JSON Parsing Failures

**Symptoms**: `jq` errors, variables show "unknown"

**Causes**:
- Reading stdin incorrectly (use `cat` or `read -r`)
- Wrong JSON path in `jq` expression
- Multi-line JSON not handled

**Solution**:
```bash
# Read all input
input_json=$(cat)

# Debug: log raw input
echo "$input_json" >> /tmp/hook-raw.log

# Use fallbacks
field=$(jq -r '.field // "fallback"' <<< "$input_json")
```

#### 3. Environment Variables Missing

**Symptoms**: Variables are empty or undefined

**Causes**:
- Trying to use variables Claude Code doesn't provide
- Misspelled variable names
- Variables not exported in parent process

**Solution**:
```bash
# Check what's actually available
env | grep CLAUDE >> /tmp/hook-env.log

# Use JSON input instead of env vars
session_id=$(jq -r '.session_id' <<< "$input_json")
```

#### 4. Hook Blocks Claude Code

**Symptoms**: Claude Code freezes, operations timeout

**Causes**:
- Hook script hangs or runs too long
- Waiting for user input (stdin already consumed)
- Network calls timing out
- Deadlock with external process

**Solution**:
```bash
# Set timeouts
timeout 5s my-external-command

# Run expensive operations in background
my-slow-operation &

# Never read from stdin again after initial read
```

#### 5. "PreToolUse hook error" / "PostToolUse hook error"

**Symptoms**: Claude Code shows "PreToolUse:ToolName hook error" for every tool call

**Causes**:
- Hook script exits with non-zero status
- `set -e` causes exit on any error (stdin read fails, jq parse error, etc.)
- stdin unavailable or empty when hook runs
- External command failures (jq, osascript, CLI tools)

**Root cause**: The hook script is using `set -euo pipefail` which causes immediate exit on ANY error. When stdin is unavailable or empty, `read -r input_json` or `cat` fails, triggering an immediate exit with non-zero status.

**Solution**:
```bash
#!/usr/bin/env bash
set -uo pipefail  # Removed -e flag

# Timeout stdin read (don't hang forever)
if ! input_json=$(timeout 2 cat); then
  exit 0  # Graceful exit if stdin unavailable
fi

# Check for empty input
if [ -z "$input_json" ]; then
  exit 0
fi

# Suppress errors on all external calls
session_id=$(jq -r '.session_id' <<< "$input_json" 2>/dev/null || echo "unknown")

# Background non-critical operations
osascript -e "display notification 'Done'" 2>/dev/null &

# ALWAYS exit successfully
exit 0
```

**Before/After:**
```bash
# BEFORE (causes errors)
set -euo pipefail              # Exits on any error
read -r input_json             # Fails if stdin empty
jq -r '.field'                 # Crashes if parse fails

# AFTER (resilient)
set -uo pipefail               # Removed -e
timeout 2 cat                  # Won't hang
jq ... 2>/dev/null || echo     # Fallback on error
exit 0                         # Never block
```

## Schema Limitations

The Claude Code settings schema does **NOT** support:

- ❌ `env` property on hook definitions (cannot set custom env vars)
- ❌ `timeout` configuration
- ❌ `async` execution
- ❌ Conditional hooks based on session state

Workarounds:
- Set env vars inside the hook script
- Implement timeouts within the script using `timeout` command
- Run background processes with `&` if needed
- Store state in files and check within hook

## Integration with Eye-in-the-Sky

### Hook Script for EITS

The `log-tool.sh` hook integrates Claude Code with Eye-in-the-Sky:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Read JSON from stdin
input_json=$(cat)

# Extract Claude Code data
session_id=$(jq -r '.session_id // "unknown"' <<< "$input_json")
tool_name=$(jq -r '.tool_name // "unknown"' <<< "$input_json")
event_type=$(jq -r '.hook_event_name // "unknown"' <<< "$input_json")

# Map session to agent
MAPPING_FILE="$CLAUDE_PROJECT_DIR/.claude/hooks/session_agent_map.json"
if [ -f "$MAPPING_FILE" ]; then
  agent_id=$(jq -r --arg sid "$session_id" '.[$sid] // "unknown"' "$MAPPING_FILE")
else
  agent_id="unknown"
fi

# Determine log type
case "$event_type" in
  PreToolUse) log_type="action" ;;
  PostToolUse) log_type="info" ;;
  *) log_type="debug" ;;
esac

# Call EITS CLI
if [ -x "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" ]; then
  "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" i-log \
    --session-id "$session_id" \
    --type "$log_type" \
    --message "Tool: $tool_name | Event: $event_type | Agent: $agent_id"
fi
```

### Where Logs Are Written

Hook execution writes logs to two destinations:

**1. Filesystem logs** (`$CLAUDE_PROJECT_DIR/logs/hooks/`):
- `env_*.log` - Environment variables snapshot (created on every hook execution)
- `benchmark.log` - Performance benchmarks (if enabled)
- These are debug logs for troubleshooting hook issues

**2. Eye-in-the-Sky database** (`~/.config/eye-in-the-sky/agents.db`):
- Table: `logs`
- Schema:
  ```sql
  CREATE TABLE logs (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL,
      type TEXT NOT NULL,           -- "action", "info", "debug"
      message TEXT NOT NULL,         -- Tool event details
      timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (session_id) REFERENCES sessions(id)
  );
  ```

**Query logs from database:**
```bash
sqlite3 ~/.config/eye-in-the-sky/agents.db \
  "SELECT type, message, timestamp FROM logs \
   WHERE session_id = 'your-session-id' \
   ORDER BY timestamp DESC LIMIT 20"
```

**Note:** The filesystem `env_*.log` files are created on EVERY hook execution (PreToolUse + PostToolUse). Consider disabling this line in production to reduce disk I/O:
```bash
# Comment out this line in .claude/hooks/log-tool.sh
# env | sort >"$CLAUDE_PROJECT_DIR/logs/hooks/env_$(date +%Y%m%d_%H%M%S).log"
```

### Session-Agent Mapping

Create `.claude/hooks/session_agent_map.json`:

```json
{
  "97c212de-24e0-4c8d-b16a-876e406b7c14": "2743c649-18ab-422d-bdbe-bb3b59d2c431"
}
```

Update this file when starting new EITS sessions to maintain the mapping.

## Performance Considerations

### Hook Execution Time

- Hooks are **synchronous** and block tool execution
- Target: < 100ms per hook
- Avoid: Network calls, heavy computation, large file I/O

### Optimization Strategies

1. **Use background processes** for non-critical tasks:
   ```bash
   my-slow-task &
   ```

2. **Cache lookups** instead of repeated queries:
   ```bash
   # Bad: query every time
   agent_id=$(curl http://api/agent)

   # Good: cache in file
   CACHE="/tmp/agent-cache"
   if [ -f "$CACHE" ]; then
     agent_id=$(cat "$CACHE")
   fi
   ```

3. **Batch operations** instead of per-hook:
   ```bash
   # Append to buffer, flush periodically
   echo "event" >> /tmp/buffer.log
   ```

## Security Considerations

### Input Validation

Always validate JSON input before use:

```bash
# Validate JSON structure
if ! jq empty <<< "$input_json" 2>/dev/null; then
  echo "Invalid JSON" >&2
  exit 1
fi

# Sanitize for shell use
session_id=$(jq -r '.session_id' <<< "$input_json" | tr -cd '[:alnum:]-')
```

### File Permissions

Protect sensitive mapping files:

```bash
# Session-agent mapping should not be world-readable
chmod 600 .claude/hooks/session_agent_map.json
```

### Script Permissions

Only allow owner to execute hooks:

```bash
chmod 700 .claude/hooks/*.sh
```

## Testing Hooks

### Manual Testing

1. Create a test hook that logs everything:
   ```bash
   #!/usr/bin/env bash
   cat > /tmp/test-hook.log
   ```

2. Run Claude Code and trigger tools

3. Inspect `/tmp/test-hook.log` to see actual JSON format

### Automated Testing

```bash
#!/usr/bin/env bash
# test-hook.sh - Unit test for hooks

# Simulate Claude Code JSON input
test_json='{
  "session_id": "test-session-123",
  "tool_name": "Read",
  "hook_event_name": "PreToolUse"
}'

# Run hook with test input
result=$(echo "$test_json" | ./log-tool.sh 2>&1)

# Verify output
if [[ $result == *"test-session-123"* ]]; then
  echo "✓ Test passed"
else
  echo "✗ Test failed"
  exit 1
fi
```

## Resources

- [Claude Code Hooks Documentation](https://docs.claude.com/en/docs/claude-code/hooks)
- [Claude Code Settings Schema](https://json.schemastore.org/claude-code-settings.json)
- [jq Manual](https://stedolan.github.io/jq/manual/)

## Appendix: Complete Working Example

See `.claude/hooks/log-tool.sh` in this repository for a complete, production-ready hook that:
- Extracts session and tool data from JSON
- Maps sessions to agent IDs
- Calls Eye-in-the-Sky CLI for logging
- Shows macOS notifications on success/failure
- Handles errors gracefully
