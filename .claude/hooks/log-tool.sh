#!/usr/bin/env bash
set -uo pipefail  # Removed -e to allow graceful error handling

# Make sure logs directory exists
mkdir -p "$CLAUDE_PROJECT_DIR/logs/hooks" 2>/dev/null || true

# Log environment variables for debugging (don't fail on error)
env | sort >"$CLAUDE_PROJECT_DIR/logs/hooks/env_$(date +%Y%m%d_%H%M%S).log" 2>/dev/null || true

# Read JSON input from stdin with timeout
if ! input_json=$(timeout 2 cat); then
  # If read fails or times out, exit gracefully
  exit 0
fi

# If input is empty, exit gracefully
if [ -z "$input_json" ]; then
  exit 0
fi

# Extract fields from JSON stdin (with fallbacks)
event_type=$(jq -r '.hook_event_name // .type // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")
tool_name=$(jq -r '.tool_name // .tool_use.name // .message.content[].name // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")
tool_id=$(jq -r '.tool_use.id // .message.content[].id // "no-id"' <<<"$input_json" 2>/dev/null || echo "no-id")
session_id=$(jq -r '.session_id // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")

# Map session_id to agent_id (JSON file lookup - benchmarked as faster than SQLite)
agent_mapping_file="$CLAUDE_PROJECT_DIR/.claude/hooks/session_agent_map.json"
if [ -f "$agent_mapping_file" ]; then
  agent_id=$(jq -r --arg sid "$session_id" '.[$sid] // "unknown"' "$agent_mapping_file" 2>/dev/null || echo "unknown")
else
  agent_id="unknown"
fi

# Decide message type
case "$event_type" in
PreToolUse) log_type="action" ;;
PostToolUse) log_type="info" ;;
*) log_type="debug" ;;
esac

# Log message content
message="Tool event: $event_type | name=$tool_name | id=$tool_id | session=$session_id | agent=$agent_id"

# Call Eye in the Sky CLI
if [ -x "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" ]; then
  if "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" i-log \
    --session-id "$session_id" \
    --type "$log_type" \
    --message "$message" 2>/dev/null; then
    # macOS alert on successful log (don't block if it fails)
    osascript -e "display notification \"Tool: $tool_name ($log_type)\" with title \"Claude Hook\" subtitle \"Logged successfully\"" 2>/dev/null &
  else
    # Log failed notification (don't block if it fails)
    osascript -e "display notification \"Tool: $tool_name\" with title \"Claude Hook\" subtitle \"Log failed\" sound name \"Basso\"" 2>/dev/null &
  fi
fi

# Always exit successfully so we don't block Claude Code
exit 0
