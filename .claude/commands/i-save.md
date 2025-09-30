# Save Session Context

Save your current session state for resumption later.

## Instructions

1. Use the `i-save-context` MCP tool to save your progress
2. Provide current phase and progress information
3. Include completed tasks, pending tasks, and next actions
4. Optionally include key decisions, important files, and notes

## Example

```
Use the i-save-context MCP tool with:
- agent_id: your agent id
- current_phase: "{{phase}}"
- pending_tasks: array of remaining tasks
- completed_tasks: array of finished tasks
- next_actions: array of next steps
- key_decisions: important decisions made
- important_files: critical files modified
- auto_save: true (to set status to idle)
```

## Usage

```
/i-save Working on authentication implementation
```

This saves your current session context and sets status to idle.
