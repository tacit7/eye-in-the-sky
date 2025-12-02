# Action Plan: Incremental Log Viewer with Append-Based Refresh

**Date:** 2025-10-28

---

## 🎯 Goal
Implement a live-updating **plain log viewer** in the `logs` tab that:
- Appends new logs fetched since the last `created_at` timestamp.
- Refreshes automatically every 5 seconds.
- Keeps scroll position stable unless near bottom.
- Avoids re-rendering the entire viewport.

---

## ⚙️ Step-by-Step Implementation

### 1. Extend the Logs Model

Add these new fields to `Model`:

```go
type Model struct {
    viewport        viewport.Model
    logs            []domain.Log
    width, height   int
    refreshInterval time.Duration
    loadLogs        func(after time.Time) ([]domain.Log, error)
    lastTimestamp   time.Time
}
```

---

### 2. Modify the Constructor

```go
func New(loadLogs func(after time.Time) ([]domain.Log, error), width, height int) Model {
    m := Model{
        viewport:        viewport.New(width, height),
        refreshInterval: 5 * time.Second,
        loadLogs:        loadLogs,
        width:           width,
        height:          height,
    }
    return m
}
```

---

### 3. Add Message Types

```go
type RefreshMsg struct{}
type LogsLoadedMsg struct{ Logs []domain.Log }
type ErrMsg struct{ error }
```

---

### 4. Add Tick Command

```go
func tickCmd(d time.Duration) tea.Cmd {
    return tea.Tick(d, func(t time.Time) tea.Msg {
        return RefreshMsg{}
    })
}
```

---

### 5. Load Command (Incremental Fetch)

```go
func (m *Model) loadCmd() tea.Cmd {
    return func() tea.Msg {
        logs, err := m.loadLogs(m.lastTimestamp)
        if err != nil {
            return ErrMsg{err}
        }
        return LogsLoadedMsg{Logs: logs}
    }
}
```

---

### 6. Update Logic (Append Instead of Replace)

```go
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            m.viewport.LineUp(1)
        case "down", "j":
            m.viewport.LineDown(1)
        }
    case LogsLoadedMsg:
        if len(msg.Logs) > 0 {
            m.logs = append(m.logs, msg.Logs...)
            m.appendToViewport(msg.Logs)
            m.lastTimestamp = msg.Logs[len(msg.Logs)-1].CreatedAt
        }
        return m, tickCmd(m.refreshInterval)
    case RefreshMsg:
        return m, tea.Batch(m.loadCmd(), tickCmd(m.refreshInterval))
    }
    return m, nil
}
```

---

### 7. Append Helper

```go
func (m *Model) appendToViewport(newLogs []domain.Log) {
    var sb strings.Builder
    for _, log := range newLogs {
        sb.WriteString(fmt.Sprintf("%s  %s  %s\n",
            log.Timestamp.Format("15:04:05"),
            log.Type,
            log.Message,
        ))
    }
    m.viewport.SetContent(m.viewport.View() + sb.String())

    // Auto-scroll if near bottom
    if m.viewport.YOffset >= m.viewport.ScrollHeight()-m.viewport.Height {
        m.viewport.GotoBottom()
    }
}
```

---

### 8. Data Layer Query

In your DB adapter:

```go
func (db *DB) GetLogsAfter(agentID string, after time.Time) ([]domain.Log, error) {
    return db.Query(`
        SELECT * FROM logs
        WHERE agent_id = ? AND created_at > ?
        ORDER BY created_at ASC
    `, agentID, after)
}
```

---

### 9. Integration Example

```go
m.logsTab = logs.New(
    func(after time.Time) ([]domain.Log, error) {
        return m.data.DB.GetLogsAfter(m.currentAgentID, after)
    },
    width,
    height,
)
```

---

## 🧪 Testing Checklist

| Test Case | Expected Result |
|------------|----------------|
| Initial load | Loads all logs once |
| 5s refresh | Appends new logs only |
| Scroll position | Stays stable during refresh |
| Auto-scroll | Scrolls to bottom if near end |
| No new logs | View remains unchanged |
| Restart view | Rebuilds viewport cleanly |

---

## ⚡ Optional Enhancements

- **Backoff refresh** after several idle cycles.  
- **Color-coded severity** (INFO, WARN, ERROR).  
- **Retention window**: trim oldest logs beyond 1000 lines.  
- **Manual refresh keybinding** (`r` for reload).  

---

✅ **End Result:**  
A lightweight, auto-refreshing, append-only log viewer that behaves like `tail -f` and integrates seamlessly with Bubble Tea's viewport.
