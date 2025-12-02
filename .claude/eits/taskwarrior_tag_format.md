# TaskWarrior Tag Format Requirements

## Critical Issues: Hyphens and Colons in Tags

TaskWarrior does not properly handle:
1. **Hyphens (`-`)** in tag names - causes filtering to fail silently
2. **Colons (`:`)** in tag queries - accepts them when creating but can't query them

## Required Format Conversion

All UUID-based identifiers must be converted from hyphen format to underscore format when used as TaskWarrior tags:

### Session IDs
- **Wrong**: `+session:da0df968-a580-45fc-9ed8-3f06740beb53` (has hyphens and colon)
- **Correct**: `+session_da0df968_a580_45fc_9ed8_3f06740beb53` (all underscores)

### Agent IDs
- **Wrong**: `+agent:cd9e1db3-c7aa-43d7-b67c-3e76e5b2c4b3` (has hyphens and colon)
- **Correct**: `+agent_cd9e1db3_c7aa_43d7_b67c_3e76e5b2c4b3` (all underscores)

### Subagent IDs
- **Wrong**: `+subagent:489fb01c-6860-40f1-8f37-b29cfcad2590` (has hyphens and colon)
- **Correct**: `+subagent_489fb01c_6860_40f1_8f37_b29cfcad2590` (all underscores)

## Implementation Details

### TUI Code (view_tasks.go)
The TUI converts UUIDs to underscore format AND removes colons:
```go
searchTag = fmt.Sprintf("+subagent_%s", strings.ReplaceAll(m.selectedAgent.ID, "-", "_"))
searchTag = fmt.Sprintf("+session_%s", strings.ReplaceAll(m.selectedAgent.SessionID, "-", "_"))
agentSearchStr = fmt.Sprintf("+agent_%s", strings.ReplaceAll(m.selectedAgent.ID, "-", "_"))
```

### Creating Tasks
When creating tasks programmatically:
1. Always convert UUID hyphens to underscores
2. Add tags to the task description (TaskWarrior's tag parsing is inconsistent)
3. Example: `task add "Fix bug +session:${SESSION_ID//-/_} +agent:${AGENT_ID//-/_}"`

### Subagent Responsibilities
Subagents launched via the Task tool must:
1. Convert their agent ID from UUID format to underscore format
2. Tag tasks with `+subagent:` using underscores
3. Never overwrite task descriptions with just the tag

## Common Mistakes to Avoid

1. **Don't use raw UUIDs as tags** - Always convert hyphens to underscores
2. **Don't rely on TaskWarrior's tag: syntax** - Include tags in the description
3. **Don't overwrite descriptions** - Use `task modify` to append tags, not replace

## Testing Tag Format

To verify tags are correctly formatted:
```bash
# Check for incorrect hyphenated tags
task export | grep -E "\\+(session|agent|subagent):[^\"]*-"

# List all session tags
task export | grep -o "+session:[a-zA-Z0-9_]*" | sort | uniq -c

# List all agent/subagent tags
task export | grep -o "+\\(agent\\|subagent\\):[a-zA-Z0-9_]*" | sort | uniq -c
```

## Migration Script

To fix existing tasks with hyphenated tags:
```bash
# Fix session tags
task export | jq -r '.[] | select(.description | contains("+session:") and contains("-")) | .uuid' | \
while read uuid; do
  desc=$(task export | jq -r --arg u "$uuid" '.[] | select(.uuid == $u) | .description')
  fixed=$(echo "$desc" | sed 's/\(+session:\)\([^[:space:]]*\)-/\1\2_/g')
  task "$uuid" modify "$fixed"
done
```