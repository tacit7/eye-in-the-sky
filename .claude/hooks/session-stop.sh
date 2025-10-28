#!/usr/bin/env bash
set -uo pipefail  # Removed -e to allow graceful error handling

# Read JSON input from stdin with timeout
if ! input_json=$(timeout 2 cat); then
  # If read fails or times out, exit gracefully
  exit 0
fi

# If input is empty, exit gracefully
if [ -z "$input_json" ]; then
  exit 0
fi

# Extract session_id from JSON stdin
session_id=$(jq -r '.session_id // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")

# Map session_id to agent_id (JSON file lookup - benchmarked as faster than SQLite)
agent_mapping_file="$CLAUDE_PROJECT_DIR/.claude/hooks/session_agent_map.json"
if [ -f "$agent_mapping_file" ]; then
  agent_id=$(jq -r --arg sid "$session_id" '.[$sid] // "unknown"' "$agent_mapping_file" 2>/dev/null || echo "unknown")
else
  agent_id="unknown"
fi

# Log stop action message
message="Session stopped | session=$session_id | agent=$agent_id"

# Call Eye in the Sky CLI to log stop action
if [ -x "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" ]; then
  "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" i-log \
    --session-id "$session_id" \
    --type "info" \
    --message "$message" 2>/dev/null
fi

# Update agent status to idle (waiting for user input)
if [ -x "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" ] && [ "$agent_id" != "unknown" ]; then
  "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" i-update-status \
    --agent-id "$agent_id" \
    --status "idle" \
    --current-task "Waiting for user input" 2>/dev/null
fi

# Always exit successfully so we don't block Claude Code
exit 0
