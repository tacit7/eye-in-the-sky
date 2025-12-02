# Scripts

## seed_agents.go

Generate test agents with mock data for development and testing.

### Usage

```bash
# Generate 5 agents (default)
go run scripts/seed_agents.go

# Generate 10 agents
go run scripts/seed_agents.go -count 10

# Generate 20 agents to custom database
go run scripts/seed_agents.go -count 20 -db /path/to/agents.db
```

### What it creates

For each agent, the script generates:
- Agent record with realistic status, feature description, and current task
- Session record linked to the agent
- 0-5 random notes with auto-generated titles
- 1-10 random actions (task_start, file_operation, git_commit, status_update)
- 0-5 random git commits with realistic commit messages

### Mock Data Sources

The script uses realistic mock data including:
- **Statuses**: active, working, idle, completed, failed
- **Features**: "user authentication system", "payment processing", etc.
- **Tasks**: "implementing database schema", "writing unit tests", etc.
- **Projects**: web-app, api-server, mobile-backend, etc.
- **UUIDs**: Generated using `github.com/google/uuid`

All timestamps are randomized within the last 72 hours to simulate real activity.
