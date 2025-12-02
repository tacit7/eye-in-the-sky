# Action Plan: Breaking Apart TUI Views into Modular Structure

## Overview

This document provides a detailed action plan to refactor the existing monolithic Bubble Tea TUI architecture into the new modular structure using **views** and **subviews (tabs)**. The goal is to separate responsibilities per view, isolate tab logic, and align with standard Bubble Tea Model/Update/View conventions.

---

## 1. View Mapping

| Old ViewType | New Package / View | Tabs | Notes |
|---------------|--------------------|-------|-------|
| `ViewList` | `views/overview/` | 5 | Main navigation (Overview, Project, Claude, Token Usage, Keybindings) |
| `ViewDetail` | `views/agent_details/` | 7 | Agent-specific details and sub-tabs |

Each view is its own Bubble Tea model, managing internal state and update logic.

---

## 2. Directory Creation

```bash
mkdir -p internal/ui/app/views/overview
mkdir -p internal/ui/app/views/project
mkdir -p internal/ui/app/views/claude
mkdir -p internal/ui/app/views/usage
mkdir -p internal/ui/app/views/keybindings
mkdir -p internal/ui/app/views/agent_details/tabs
```

---

## 3. Split `ViewList` (Main Tabs)

1. **Create new view files for each main tab:**

| New File | Responsibility |
|-----------|----------------|
| `model_overview.go` | Holds agent list, selection state, and layout dimensions |
| `update_overview.go` | Handles input and data refresh for Overview tab |
| `view_overview.go` | Renders agent list and status indicators |

2. **Repeat for each top-level tab:**
   - `views/project/`
   - `views/claude/`
   - `views/usage/`
   - `views/keybindings/`

Each directory should contain:
```
model_<name>.go
update_<name>.go
view_<name>.go
```

3. **Update routing in `update_root.go`:**

```go
switch m.currentView {
case ViewOverview:
    return m.overview.Update(msg)
case ViewProject:
    return m.project.Update(msg)
case ViewClaude:
    return m.claude.Update(msg)
case ViewUsage:
    return m.usage.Update(msg)
case ViewKeybindings:
    return m.keybindings.Update(msg)
}
```

---

## 4. Split `ViewDetail` (Agent Details)

1. **Create the main container:**
```
internal/ui/app/views/agent_details/
├── model_agent_details.go
├── update_agent_details.go
├── view_agent_details.go
└── tabs/
```

2. **Create sub-tabs inside `/tabs`:**

| Tab | File | Resolver Context |
|------|------|------------------|
| ← Back | handled in `update_agent_details.go` | `agent_details.back` |
| Overview | `overview_tab.go` | `agent_details.overview` |
| Tasks | `tasks_tab.go` | `agent_details.tasks` |
| Actions | `actions_tab.go` | `agent_details.actions` |
| Logs | `logs_tab.go` | `agent_details.logs` |
| Commits | `commits_tab.go` | `agent_details.commits` |
| Notes | `notes_tab.go` | `agent_details.notes` |

Each tab should implement:
```go
func (t *TasksTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
func (t *TasksTab) View() string { ... }
```

---

## 5. Keybinding Update

Update YAML:

```yaml
overview:
  j: down
  k: up
  enter: select

agent_details:
  overview:
    j: down
  tasks:
    j: down
    r: refresh
  actions:
    j: down
  logs:
    j: down
  commits:
    j: down
  notes:
    j: down
```

---

## 6. Testing Realignment

- Move or rename tests accordingly:
  - `update_list_test.go` → `overview/update_overview_test.go`
  - `update_projects_test.go`, `update_claude_test.go`, etc., into their respective view folders.
- Add smoke tests per view verifying `Init()`, `Update()`, and `View()` run without errors.

---

## 7. Cleanup

- Delete old monolithic files:
  - `view_main.go`, `update_main.go`
  - `view_detail.go`, `update_detail.go`
  - `detail_tabs.go`
- Consolidate shared UI utilities under `components/`.
- Run:
  ```bash
  goimports -w .
  go test ./...
  go build ./cmd/eye-ui
  ```

---

## ✅ Resulting Layout

```
views/
├── overview/
├── project/
├── claude/
├── usage/
├── keybindings/
└── agent_details/
    ├── model_agent_details.go
    ├── update_agent_details.go
    ├── view_agent_details.go
    └── tabs/
        ├── overview_tab.go
        ├── tasks_tab.go
        ├── actions_tab.go
        ├── logs_tab.go
        ├── commits_tab.go
        └── notes_tab.go
```
