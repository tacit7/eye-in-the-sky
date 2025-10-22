# TaskWarrior Agent Workflow Guide

## Task Status Workflow for Agents

Use tags to track task workflow status without modifying TaskWarrior configuration:

### Workflow States

1. **Pending** (default) - Task needs to be worked on
   - No special tag needed (default state)

2. **Working** - Agent is actively working on the task
   - Add tag: `+working`
   - Command: `task <id> modify +working`

3. **Review** - Agent has finished, needs review
   - Remove working tag, add review tag
   - Command: `task <id> modify -working +review`

4. **Completed** - Task passed review and is done
   - Command: `task <id> done`

5. **Blocked** - Task can't proceed
   - Add tag: `+blocked`
   - Command: `task <id> modify +blocked`

## Agent Commands

### When starting work on a task:
```bash
task <id> start                    # Mark as started (adds start timestamp)
task <id> modify +working          # Add working tag
task <id> annotate "Starting implementation"
```

### When finishing a task (ready for review):
```bash
task <id> modify -working +review  # Switch from working to review
task <id> annotate "Ready for review - implementation complete"
```

### After review approval:
```bash
task <id> done                     # Mark as completed
```

### If blocked:
```bash
task <id> modify +blocked
task <id> annotate "Blocked: <reason>"
```

## Viewing Tasks by Status

### See all working tasks:
```bash
task +working list
```

### See all tasks ready for review:
```bash
task +review list
```

### See blocked tasks:
```bash
task +blocked list
```

### See tasks for a specific agent that are working:
```bash
task +agent_<agent_id> +working list
```

### See tasks for a subagent ready for review:
```bash
task +subagent_<agent_id> +review list
```

## Example Workflow

```bash
# 1. Agent starts working on task 29
task 29 start
task 29 modify +working
task 29 annotate "Implementing UI fix for subagent display"

# 2. Agent finishes implementation
task 29 modify -working +review
task 29 annotate "Implementation complete - changed 'Agent:' to 'Subagent:' in detail view"

# 3. After review (by human or parent agent)
task 29 done
```

## Integration with Eye in the Sky

When agents update task status, they should also:
1. Update their Eye in the Sky status accordingly
2. Log the task status change as an action
3. Include task status in commit messages

Example:
```
i-update-status: working (on task #29)
i-log-action: Started working on TaskWarrior task #29
git commit -m "feat: Start implementation of task #29 [working]"
```

## Benefits of This Approach

- No TaskWarrior configuration needed
- Works immediately with existing setup
- Compatible with existing tag system
- Easy to query and filter
- Visible in task descriptions
- Can be combined with other tags (priority, project, etc.)