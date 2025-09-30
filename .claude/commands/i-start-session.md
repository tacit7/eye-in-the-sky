# Start Session

Register yourself as an agent and begin tracking your work session.

## Instructions

1. Use the `i-register` MCP tool to register yourself
2. Provide a description of what you'll be working on
3. Optionally include project name and worktree path
4. Store the returned agent_id for use in subsequent calls

## Example

```
Use the i-register MCP tool with:
- description: "{{description}}"
- project_name: "eye-in-the-sky" (optional)
- worktree_path: current directory (optional)

After registration, confirm your agent_id and set status to "working".
```

## Usage

```
/i-start-session Working on authentication system
```
