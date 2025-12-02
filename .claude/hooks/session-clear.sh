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

# Extract new session_id and source from JSON stdin
new_session_id=$(jq -r '.session_id // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")
source=$(jq -r '.source // "unknown"' <<<"$input_json" 2>/dev/null || echo "unknown")

# Only proceed if this is a clear event
if [ "$source" != "clear" ]; then
  exit 0
fi

# Log session clear with new session ID
message="New session started after clear | new_session=$new_session_id | Please re-register with i-start-session"
if [ -x "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" ]; then
  "$CLAUDE_PROJECT_DIR/bin/eye-in-the-sky" i-log \
    --session-id "$new_session_id" \
    --type "info" \
    --message "$message" 2>/dev/null
fi

# Display sticky macOS alert dialog (won't disappear until user dismisses)
osascript -e "display alert \"Eye-in-the-Sky: Session Cleared\" message \"New session ID: $new_session_id\n\nRe-register with i-start-session and update session_agent_map.json\" as critical" &

# Always exit successfully so we don't block Claude Code
exit 0
