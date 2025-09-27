# Eye in the Sky - Dashboard

A web dashboard for monitoring Claude Code multi-agent sessions.

## Quick Start

1. **Run the dashboard:**
   ```bash
   go run main.go
   ```

2. **Open your browser:**
   ```
   http://localhost:8080
   ```

3. **Custom port:**
   ```bash
   go run main.go -port 8081
   ```

## Features Implemented

✅ **Basic Dashboard Structure**
- Responsive Bootstrap 5 UI
- Agent overview page with mock data
- Agent detail page with activity timeline
- Status indicators and badges

✅ **Frontend Features**
- Auto-refresh every 30 seconds
- Interactive JavaScript functionality
- Responsive design for mobile/desktop
- Toast notifications
- Keyboard shortcuts (Ctrl+R to refresh, Escape to go back)

✅ **Backend Features**
- Go HTTP server with template rendering
- Static file serving for CSS/JS
- Mock API endpoints for agent management
- Clean project structure following the PRD

## Current Status

This is the initial dashboard implementation with **mock data**. The dashboard is fully functional for UI testing and development, but needs to be connected to:

1. **Database layer** (SQLite with agent, actions, commits tables)
2. **MCP server integration** (for receiving agent updates)
3. **Real-time data** (replacing mock data with actual agent information)

## Project Structure

```
wt-dashboard/
├── main.go                           # Entry point
├── internal/dashboard/server.go      # HTTP server and handlers
├── web/
│   ├── templates/
│   │   ├── base.html                 # Base template with Bootstrap 5
│   │   ├── index.html                # Agent overview page
│   │   └── agent.html                # Agent detail page
│   └── static/
│       ├── css/styles.css            # Custom dashboard styles
│       └── js/dashboard.js           # Interactive JavaScript
└── README.md                         # This file
```

## Next Steps (Low-Hanging Fruit)

1. **Add database models** - Create SQLite schema and basic CRUD operations
2. **Connect mock data to database** - Replace hardcoded agent data
3. **Add more interactive features** - Real-time updates, filtering, search
4. **Improve error handling** - Better error pages and API responses
5. **Add configuration options** - Environment variables, config files

## Mock Data

The dashboard currently shows:
- 3 sample agents with different statuses (active, working, idle)
- Activity timeline with various action types
- Git commits with timestamps
- Summary statistics

Perfect for development and UI testing!