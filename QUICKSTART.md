# Eye in the Sky - Quick Start Guide

Get up and running with Eye in the Sky in under 5 minutes!

## 🚀 Quick Setup

### 1. Build & Start
```bash
# Build the integrated server
go build -o bin/eye-in-the-sky-integrated ./main.go

# Start the server
./bin/eye-in-the-sky-integrated
```

### 2. Access Dashboard
Open your browser to: **http://localhost:8080**

### 3. Register Your First Agent

In Claude Code, use the MCP tool:

**For Worktree Agents:**
```json
{
  "description": "Working on user authentication",
  "worktree_path": "/path/to/your/project"
}
```

**For Claude Desktop Agents:**
```json
{
  "description": "Building React dashboard",
  "project_name": "Dashboard Project"
}
```

## 📋 Essential MCP Tools

### Register Agent
```json
// Auto-generates agent ID (recommended)
{
  "description": "Your task description",
  "worktree_path": "/your/project/path"
}
```

### Update Status
```json
{
  "agent_id": "your-agent-id",
  "status": "working",
  "current_task": "Implementing login system"
}
```

### Log Activities
```json
{
  "agent_id": "your-agent-id",
  "action_type": "file_operation",
  "description": "Created authentication middleware"
}
```

### End Session
```json
{
  "agent_id": "your-agent-id",
  "summary": "Completed user authentication feature",
  "final_status": "completed"
}
```

## 🎯 Quick Workflow

1. **Start** Eye in the Sky server
2. **Register** your Claude Code agent
3. **Update** status as you work
4. **Log** major actions and milestones
5. **Track** git commits automatically
6. **End** session with summary
7. **Monitor** all agents via dashboard

## 📊 Dashboard Overview

- **Agent List**: See all active agents and their status
- **Activity Feed**: Real-time stream of agent actions
- **Agent Details**: Click any agent ID for detailed view
- **Status Badges**: 🟢 Active, 🟡 Idle, 🔴 Failed, ⚪ Completed

## 🛠️ Troubleshooting

**Server won't start?**
```bash
# Check if port is free
lsof -i :8080

# Use different port
./bin/eye-in-the-sky-integrated -port 8081
```

**Dashboard shows error?**
```bash
# Restart with fresh database
./bin/eye-in-the-sky-integrated -db ./fresh-agents.db
```

**MCP tools not working?**
- Verify server is running: `ps aux | grep eye-in-the-sky`
- Check server logs for errors
- Try the `help` tool for tool information

## 💡 Pro Tips

- **Agent IDs** are auto-generated as 8-character git-style hashes
- **Use descriptive task descriptions** for better tracking
- **Update status frequently** for accurate monitoring
- **End sessions with summaries** for better project history
- **Multiple agents** can work simultaneously on different projects

## 📖 Need More Help?

- Read the full **MANUAL.md** for comprehensive documentation
- Check server logs for detailed error information
- Use the MCP `help` tool for tool-specific guidance

---

**Your agent tracking hash: 534002f0** 🎯

*Happy multi-agent development!*