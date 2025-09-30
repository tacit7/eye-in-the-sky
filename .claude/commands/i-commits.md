# Log Commits

Track git commits made during your session.

## Instructions

1. Use the `i-commits` MCP tool to log commits
2. Provide commit hashes and optional messages
3. Or let it auto-detect recent commits from git log

## Example

**Manual logging:**
```
Use the i-commits MCP tool with:
- agent_id: your agent id
- commit_hashes: ["abc123", "def456"]
- commit_messages: ["Add JWT middleware", "Fix login bug"]
```

**Auto-detect:**
```
Use the i-commits MCP tool with:
- agent_id: your agent id
(leave commit_hashes empty to auto-detect last 3 commits)
```

## Usage

```
/i-commits
```

This auto-detects and logs your recent commits.
