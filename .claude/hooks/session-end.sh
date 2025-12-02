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

# Extract session_id and reason from JSON stdin
session_id=$(jq -r '.session_id // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")
reason=$(jq -r '.reason // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")

# Map session_id to agent_id (JSON file lookup)
agent_mapping_file="$CLAUDE_PROJECT_DIR/.claude/hooks/session_agent_map.json"
if [ -f "$agent_mapping_file" ]; then
  agent_id=$(jq -r --arg sid "$session_id" '.[$sid] // "unknown"' "$agent_mapping_file" 2>/dev/null || echo "unknown")
else
  agent_id="unknown"
fi

# Log session end
message="Session ended | session=$session_id | agent=$agent_id | reason=$reason"
if [ -x "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" ]; then
  "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" i-log \
    --session-id "$session_id" \
    --type "info" \
    --message "$message" 2>/dev/null
fi

# Update agent status to completed
if [ -x "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" ] && [ "$agent_id" != "unknown" ]; then
  "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" i-update-status \
    --agent-id "$agent_id" \
    --status "completed" \
    --current-task "Session ended: $reason" 2>/dev/null
fi

# Remove mapping from session_agent_map.json
if [ -f "$agent_mapping_file" ] && [ "$session_id" != "unknown" ]; then
  # Create temp file with mapping removed
  jq --arg sid "$session_id" 'del(.[$sid])' "$agent_mapping_file" > "${agent_mapping_file}.tmp" 2>/dev/null
  # Replace original file if jq succeeded
  if [ -f "${agent_mapping_file}.tmp" ]; then
    mv "${agent_mapping_file}.tmp" "$agent_mapping_file" 2>/dev/null || rm -f "${agent_mapping_file}.tmp"
  fi
fi

# Send desktop notification
short_session_id="${session_id:0:8}"
osascript -e "display notification \"Session ${short_session_id} ended: ${reason}\" with title \"Eye in the Sky\" sound name \"Glass\"" 2>/dev/null &

# Always exit successfully so we don't block Claude Code
exit 0
