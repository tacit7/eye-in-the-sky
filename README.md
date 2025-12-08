# Eye in the Sky - Claude Code Multi-Agent Management System

**Real-time visibility and control over multiple concurrent Claude Code instances**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Dashboard](https://img.shields.io/badge/Dashboard-Live-brightgreen)](http://localhost:8080)

---

## 🎯 Overview

Eye in the Sky is a comprehensive multi-agent management system designed for Claude Code users. It provides real-time tracking, monitoring, and coordination of multiple AI agents working across different git worktrees and Claude Desktop projects.

### Key Features

- **🔍 Real-time Agent Tracking**: Monitor multiple Claude Code instances simultaneously
- **📊 Web Dashboard**: Intuitive interface at `http://localhost:8080`
- **🔧 MCP Integration**: Seamless Claude Code integration via Model Context Protocol
- **📁 Git Worktree Support**: Track agents working in different git repositories
- **🖥️ Claude Desktop Support**: Manage desktop-based agents with window management
- **📝 Activity Logging**: Complete audit trail of agent actions and git commits
- **🎨 Responsive UI**: Modern Bootstrap 5 interface with real-time updates

---

## 🚀 Quick Start

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
```json
{
  "description": "Working on user authentication",
  "worktree_path": "/path/to/your/project"
}
```

**📖 [Read the Quick Start Guide →](QUICKSTART.md)**

### 4. MCP Setup (Claude CLI)

If Claude Code says there’s no MCP server available, add the Eye in the Sky server for this project using the Claude CLI:

```bash
claude mcp add --transport stdio eye-in-the-sky -- /absolute/path/to/eye-in-the-sky/bin/eye-in-the-sky
claude mcp list
```

You should see a ✓ Connected status and “Scope: Local config (private to you in this project)”. See MCP_SETUP_GUIDE.md for details and troubleshooting.

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| **[📖 MANUAL.md](MANUAL.md)** | Complete user manual with detailed documentation |
| **[🚀 QUICKSTART.md](QUICKSTART.md)** | Get up and running in under 5 minutes |
| **[🛠️ CLAUDE.md](CLAUDE.md)** | Technical guide for Claude Code integration |

---

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Claude Code   │    │  Eye in the Sky │    │  Web Dashboard  │
│    Instances    │◄──►│   MCP Server    │◄──►│  (localhost:8080)│
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │ SQLite Database │
                       │  (agents.db)    │
                       └─────────────────┘
```

### Agent Types

- **📁 Worktree Agents**: Claude Code instances working in git worktrees
- **🖥️ Desktop Agents**: Claude Desktop instances working on projects

---

## 🛠️ Installation & Setup

### Prerequisites
- **Go 1.21+** (for building from source)
- **Git** (for version control)
- **Modern web browser** (for dashboard access)

### Build Options

```bash
# Integrated server (MCP + Dashboard) - Recommended
go build -o bin/eye-in-the-sky-integrated ./main.go

# MCP-only server
go build -o bin/eye-in-the-sky ./cmd/server

# Dashboard-only server
go build -o bin/dashboard ./wt-dashboard/cmd/server
```

### Command Line Options
```bash
./bin/eye-in-the-sky-integrated [options]

Options:
  -help           Show help information

Database location: ~/.config/eye-in-the-sky/agents.db
```

---

## 🎮 MCP Tools Reference

The system provides 7 MCP tools for Claude Code integration:

| Tool | Purpose | Auto-Generated ID |
|------|---------|-------------------|
| `register_agent` | Register worktree-based agents | ✅ |
| `register_claude_desktop_agent` | Register Claude Desktop agents | ✅ |
| `update_status` | Update agent status and current task | - |
| `log_action` | Log agent activities and operations | - |
| `log_commits` | Track git commits made by agents | - |
| `end_session` | Complete agent session with summary | - |
| `help` | Get detailed tool help and usage | - |

**Agent IDs**: Automatically generated as 8-character git-style hashes (e.g., `a3f7d2e1`)

---

## 📊 Dashboard Features

### Main Dashboard (`/`)
- **Agent Overview**: Real-time status of all agents
- **Activity Feed**: Live stream of agent actions
- **Status Badges**: 🟢 Active, 🟡 Idle, 🔴 Failed, ⚪ Completed
- **Source Indicators**: 📁 Worktree, 🖥️ Desktop

### Agent Details (`/agent/{id}`)
- Complete activity timeline
- Git commit tracking
- Session management
- Performance metrics

### Features
- **Auto-refresh**: Updates every 30 seconds
- **Responsive Design**: Mobile and desktop optimized
- **Keyboard Shortcuts**: Ctrl+R to refresh, Escape to navigate
- **Toast Notifications**: Real-time status updates

---

## 🔧 Development

### Project Structure
```
eye-in-the-sky/
├── bin/                          # Built executables
├── cmd/server/                   # MCP-only server
├── internal/
│   ├── mcp/                      # MCP server implementation
│   ├── dashboard/                # HTTP server for web dashboard
│   ├── database/                 # SQLite connection and queries
│   └── utils/                    # Utility functions
├── web/
│   ├── templates/                # HTML templates
│   └── static/                   # CSS and JavaScript
├── data/                         # Database files
│   └── agents.db                 # SQLite database (auto-created)
├── MANUAL.md                     # Complete documentation
├── QUICKSTART.md                 # Quick start guide
└── main.go                       # Integrated server entry point
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/database
go test ./internal/mcp
```

### Database Management
```bash
# The SQLite database is created automatically
# Location: ~/.config/eye-in-the-sky/agents.db

# Reset database
rm ~/.config/eye-in-the-sky/agents.db
```

---

## 🔍 Monitoring & Troubleshooting

### Health Checks
```bash
# Check if server is running
ps aux | grep eye-in-the-sky

# Test dashboard
curl -I http://localhost:8080

# Check database
sqlite3 ~/.config/eye-in-the-sky/agents.db "SELECT COUNT(*) FROM agents;"
```

### Common Issues
- **Port conflicts**: Use `-port` flag to specify different port
- **Permission errors**: Check database file permissions
- **Template errors**: Verify `web/templates/` directory exists

### Debug Mode
```bash
# Enable verbose logging
./bin/eye-in-the-sky-integrated -debug 2> debug.log
```

---

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit changes**: `git commit -m 'Add amazing feature'`
4. **Push to branch**: `git push origin feature/amazing-feature`
5. **Open a Pull Request**

### Development Guidelines
- Follow Go conventions and best practices
- Add tests for new features
- Update documentation for changes
- Use meaningful commit messages

---

## 📈 Version History

- **v1.0**: Initial release with full MCP integration and dashboard
- **v0.9**: Beta with Claude Desktop agent support and window management
- **v0.8**: Alpha with git worktree agent tracking
- **v0.7**: Dashboard prototype with mock data

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🆘 Support

- **📖 Documentation**: [MANUAL.md](MANUAL.md) for comprehensive guide
- **🚀 Quick Help**: [QUICKSTART.md](QUICKSTART.md) for immediate setup
- **🐛 Issues**: Check existing issues in the repository
- **💬 Discussions**: Use GitHub Discussions for questions

---

**🎯 Your Agent Hash: 534002f0**

*Making multi-agent development visible and manageable.*
