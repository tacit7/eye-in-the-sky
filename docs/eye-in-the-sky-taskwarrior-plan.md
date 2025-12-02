# Eye in the Sky – Taskwarrior Integration Plan (Final v2)

## Summary
Integrate Taskwarrior (3.4.1) into the Eye in the Sky TUI, replacing the external `ftask` script. Tasks become a first-class view alongside Agents. Supports reading tasks, continuing sessions, opening agents or commits, marking done, annotating, and creating new tasks—all backed by Taskwarrior’s SQLite database.

---

## 1. Data Model

### Source
Taskwarrior now uses **Taskchampion** with SQLite backend (`~/.task/taskchampion.sqlite3`).  
You’ll query it via the **Taskwarrior CLI** (`task export`) for now; native SQLite reading can be added later if needed.

### Task JSON structure
```json
{
  "id": 123,
  "uuid": "f8a9e7b4-0000-4a91-8170-9b63b9b31d76",
  "description": "Fix refresh bug",
  "tags": ["session:a3d9f8c1", "agent:e71bc233", "commit:8e34cd1"],
  "status": "pending",
  "entry": "2025-10-18T10:22:00Z",
  "modified": "2025-10-18T11:03:00Z",
  "annotations": [{"entry": "2025-10-18T10:30:00Z", "description": "Started debugging"}]
}
```

### Go struct
```go
type TwTask struct {
    ID           int
    UUID         string
    Description  string
    Status       string
    Entry        time.Time
    Modified     time.Time
    Tags         []string
    Annotations  []TwAnnotation
    SessionID    string
    AgentID      string
    CommitHashes []string
}
type TwAnnotation struct {
    Entry       time.Time
    Description string
}
```
Tag extraction logic:
```go
for _, t := range task.Tags {
    switch {
    case strings.HasPrefix(t, "session:"):
        task.SessionID = strings.TrimPrefix(t, "session:")
    case strings.HasPrefix(t, "agent:"):
        task.AgentID = strings.TrimPrefix(t, "agent:")
    case strings.HasPrefix(t, "commit:"):
        task.CommitHashes = append(task.CommitHashes, strings.TrimPrefix(t, "commit:"))
    }
}
```

---

## 2. UI Design

### Mode Switching
- `F2` or `t` → Tasks mode  
- `F1` or `g` → Agents mode  
- Status bar shows `Mode: Tasks` when active.

### Task List View
Columns: `UUID | Description | Status | Session | Agent | Commit Count | Modified`

Supports filtering (`/`), paging (`PgUp/PgDn`), and toggling All / Pending / Completed (`a`, `p`, `x`).

---

## 3. Detail View
Triggered by `Enter`.

```
Task: Fix refresh bug
UUID: f8a9e7b4-0000-4a91-8170-9b63b9b31d76
Status: pending
Created: 2025-10-18 10:22
Modified: 2025-10-18 11:03

Session: a3d9f8c1
Agent:   e71bc233
Commits: 8e34cd1, 7bd2cc9, d3f11b2

Annotations:
- 10:30  Started debugging
```

Footer:
```
cs continue session • ca continue agent • enter view commits • d done • A annotate • n new • q back
```

---

## 4. Actions

### cs – Continue session
Runs:
```bash
claude -c <session-id>
```

### ca – Continue agent
Runs:
```bash
CLAUDE_AGENT=<agent-id> claude -c
```

### Enter on commit line
Opens popup with:
```bash
git --no-pager show <commit> --oneline --no-color -n 1
```

### d – Done
```bash
task <uuid> done
```

### A – Annotate
```bash
task <uuid> annotate "<text>"
```

### n – New Task
Interactive form:
```bash
task add "<desc>" <tags...>
```

---

## 5. Config

### keys.json
```json
"tasks": {
  "continue_session": ["c", "cs"],
  "continue_agent": ["a", "ca"],
  "view_commits": ["<Enter>"],
  "done": ["d"],
  "annotate": ["A"],
  "new_task": ["n"],
  "filter_pending": ["p"],
  "filter_completed": ["x"],
  "filter_all": ["a"],
  "back": ["<Esc>"]
}
```

### config.json
```json
"taskwarrior": {
  "enabled": true,
  "bin": "task",
  "default_filter": "status:pending",
  "poll_interval": 5,
  "database": "~/.task/taskchampion.sqlite3"
}
```

---

## 6. Implementation Layout

```
internal/dashboard/tasks.go
internal/dashboard/tasks_view.go
internal/dashboard/tasks_actions.go
internal/dashboard/tasks_keymap.go
```

### tasks.go
Handles data load via `task export`.

### tasks_view.go
Draws list/detail panes.

### tasks_actions.go
Implements `claude`, `git`, `task` commands.

### tasks_keymap.go
Maps JSON keybindings to actions.

---

## 7. Installer Updates
During setup:
- Detect `task` binary.
- Ask user: *Enable Taskwarrior integration (y/N)?*
- If yes, set `"taskwarrior.enabled": true`.

---

## 8. Edge Handling
- Missing binary → warning banner.  
- Missing DB → disable Tasks mode.  
- Multi-commit tasks show expandable count.  
- Git errors handled with popup once.

---

## 9. Tests
Mock JSON from `task export` and test actions:
- Parsing tags
- Continue session
- Annotate
- New task
- Commit popup close with `q`

---

## 10. Outstanding Choice
For commit popup, choose one:
- Summary: `git show -s --oneline`
- Short diff (default): first 20 lines of patch.
