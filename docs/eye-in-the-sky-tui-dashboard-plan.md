# Eye in the Sky TUI Dashboard Action Plan

## Overview
This document describes the detailed plan for implementing the **Eye in the Sky TUI Dashboard** using Go and the `gocui` library.  
The dashboard monitors active agents, their sessions, and tasks in real time.

---

## 1. New Directory Structure

```
eye-in-the-sky/
├── cmd/
│   ├── server/              # existing MCP server
│   └── dashboard/           # NEW TUI CLI
│       ├── main.go
│       └── config/
│           ├── config.json
│           └── keys.json
│
├── internal/
│   ├── dashboard/           # NEW package for TUI logic
│   │   ├── app.go
│   │   ├── layout.go
│   │   ├── render.go
│   │   ├── keymap.go
│   │   ├── commands.go
│   │   ├── refresh.go
│   │   ├── detail.go
│   │   └── actions.go
│   ├── database/            # existing SQLite logic
│   └── window/              # existing AppleScript manager
```

---

## 2. Configuration

### config.json
```json
{
  "poll_interval": 5,
  "default_filter": "active",
  "colors": {
    "active": "green",
    "idle": "blue",
    "working": "yellow",
    "failed": "red",
    "archived": "gray"
  },
  "scroll_mode": "scroll",
  "pagination_mode": "page"
}
```

### keys.json
```json
{
  "quit": ["q"],
  "refresh": ["r"],
  "continue_session": ["c"],
  "go_to_window": ["w"],
  "toggle_all_agents": ["a"],
  "up": ["k", "<Up>"],
  "down": ["j", "<Down>"],
  "logs": ["L"],
  "command_mode": [":"]
}
```

---

## 3. Integration Points

### Database
- Use `internal/database` for all queries.
- Connect to `data/agents.db`.
- Queries include agent listing, logs by session, and detail lookups.

### Window Manager
- Use `internal/window/manager.go` for `w` keybinding.
- Calls `Manager.BringToFront()` for the selected agent’s `window_id`.

---

## 4. Core Components

### app.go
- Initialize `gocui.Gui`.
- Load config and keymap.
- Setup layout and event loop.
- Handle refresh ticker (poll every N seconds).

### layout.go
- Define main, logs, and command views.
- Resize dynamically based on terminal size.

### render.go
- Draw agent table with color-coded statuses.
- Highlight selected row.
- Use ANSI or `fatih/color` for status colors.

### keymap.go
- Load from `keys.json`.
- Bind keys dynamically based on config.
- Support reloading bindings via `:reload`.

### refresh.go
- Poll DB every N seconds (configurable).
- Re-render main list while preserving cursor position.

### detail.go
- Display selected agent details (status, project, task).
- Show collapsible logs pane.
- `L` opens full logs view; `q` returns to main list.

### commands.go
- Implements `:` mode.
- Supports commands: `:reload`, `:quit`, `:config-reload`, `:logs`.

### actions.go
- `c`: open new shell and run `claude -c <session_id>`.
- `w`: bring agent window to front via AppleScript.

---

## 5. Polling Loop

```go
ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
for {
  select {
  case <-ticker.C:
    agents := store.ListAgents(currentFilter)
    app.RenderAgents(agents)
  }
}
```

---

## 6. User Interaction Summary

| Key | Action | Description |
|-----|---------|-------------|
| `r` | Refresh | Reload agents from DB |
| `a` | Toggle | Show all or active agents |
| `c` | Continue | Run `claude -c` for session |
| `w` | Window | Bring agent window to front |
| `L` | Logs | Show all session logs |
| `:` | Command | Enter command mode |
| `q` | Quit | Exit dashboard |

---

## 7. Development Order

| Step | Task | Output |
|------|------|---------|
| 1 | Create `internal/dashboard` package | base files |
| 2 | Load config + keys | verify on startup |
| 3 | Render static list | mock agents |
| 4 | Wire real DB | query agent data |
| 5 | Add keybindings | movement + actions |
| 6 | Implement polling | auto-refresh |
| 7 | Add logs/detail view | session info |
| 8 | Add actions (c, w) | shell + AppleScript |
| 9 | Command mode | reload + quit |
| 10 | Polish | colors + error handling |

---

## 8. Entry Point

`cmd/dashboard/main.go`:
```go
package main

import (
  "eye-in-the-sky/internal/dashboard"
  "eye-in-the-sky/internal/database"
)

func main() {
  db := database.Connect("data/agents.db")
  store := database.NewStore(db)
  app := dashboard.NewApp(store)
  app.Run()
}
```

---

## 9. Run Command

```bash
go run ./cmd/dashboard
```

---

## 10. Deliverables

- Working TUI dashboard listing all agents.
- Configurable keymaps and colors.
- Polling refresh with detail + logs view.
- Fully isolated from MCP server logic.
