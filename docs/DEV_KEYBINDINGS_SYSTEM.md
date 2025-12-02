# Eye-in-the-Sky: Keybinding System Architecture

**Last Updated:** October 2025  
**Audience:** TUI / UI developers  
**Scope:** Applies to the Bubble Tea-based TUI (`internal/ui/app`)

---

## 1. Overview

The keybinding system lets users customize keyboard shortcuts through YAML configuration.  
Instead of hard-coding key handling inside `Update()` functions, the app uses a **resolver** that maps keypresses (`tea.KeyMsg`) to **semantic actions** (e.g., `"down"`, `"new_session"`, `"refresh"`).

### Why it exists
- To **centralize input mapping** logic.  
- To **support per-view and per-tab scopes** (Overview, Project, Claude, etc.).  
- To let the **help modal render dynamically** based on YAML contents.  
- To allow **live reload** of keybindings without recompilation.

---

## 2. Data Flow Summary

```
Key press (tea.KeyMsg)
     ↓
m.Update(msg)
     ↓
handleKeyPress() → handleListKeys() / handleDetailKeys()
     ↓
m.keybindResolver.Resolve(msg, false)
     ↓
returns (action, found)
     ↓
switch action { … }  → run logic or open modal
```

If the resolver doesn’t find a mapping or the mapped action isn’t handled,  
fallback logic (hard-coded j/k navigation, etc.) runs afterward.

---

## 3. Components

### 3.1 YAML Configuration

The user keybindings live in:
```
~/.eye-in-the-sky/keybindings.yaml
```

**Example:**

```yaml
keybindings:
  global:
    help: ["ctrl+h", "?"]
    quit: ["ctrl+c", "shift+q"]
    refresh: ["r"]

  list:
    overview:
      j: down
      k: up
      enter: select
      n: new_session
    project:
      j: down
      k: up
      n: new_ticket
    claude:
      j: down
      k: up
      e: edit
      v: validate
    usage:
      j: down
      k: up
      pgdn: page_down
      pgup: page_up
    config:
      j: down
      k: up
      e: edit
      ctrl+s: save
      esc: cancel

  agent_details:
    overview:
      j: down
      k: up
      enter: select
    commits:
      j: down
      k: up
      h: scroll_left
      l: scroll_right
      r: refresh
    logs:
      j: down
      k: up
      r: refresh
    notes:
      j: down
      k: up
      n: new_note
    actions:
      j: down
      k: up
      r: rerun
    tasks:
      j: down
      k: up
      n: new_task
    projects:
      j: down
      k: up
```

Each top-level key (`global`, `list`, `agent_details`) defines **scope groups**.
Scopes use underscore format internally (e.g., `list_overview`, `agent_details_commits`).

**Note:** `detail:` is deprecated but still supported for backward compatibility. Use `agent_details:` in new configurations.

---

### 3.2 Resolver (`keybindResolver`)

Responsible for:
- Tracking **current scope** (`list_overview`, `detail_commits`, etc.).
- Looking up key-to-action mappings.
- Returning `(action, found)`.

Typical call:
```go
action, found := m.keybindResolver.Resolve(msg, m.modalManager.IsActive())
```

The resolver maintains internal state:
```go
type Resolver struct {
    config      KeybindingsConfig
    resolved    ResolvedKeybindings        // map of scope → key → action
    currentView string                      // e.g., "list" or "detail"
    currentTab  string                      // e.g., "overview" or "commits"
}
```

Scope is constructed as `view_tab` (e.g., `list_overview`, `detail_commits`).

#### Core methods
| Method | Purpose |
|--------|----------|
| `SetContext(view, tab string)` | Set current scope (e.g., `"list"`, `"overview"` → scope `"list_overview"`) |
| `Resolve(msg tea.KeyMsg, modalActive bool)` | Match keypress within current context; modalActive checks modal scope first |
| `Reload()` | Reload YAML and rebuild binding maps |
| `GetScopeKeybindings(scope string)` | For Help modal rendering |

---

### 3.3 Scopes

Each **view** defines its own logical scope using underscore format: `view_tab`.

| View / Context | Scope Name | Source File | Description |
|----------------|------------------------|--------------|--------------|
| Overview Tab (List) | `list_overview` | `update_main.go` | Shows agents list |
| Project Tab (List) | `list_project` | `update_main.go` | Project tasks, CLAUDE.md, etc. |
| Claude Tab (List) | `list_claude` | `update_main.go` | File explorer for Claude configs |
| Usage Tab (List) | `list_usage` | `update_main.go` | Token usage + viewport |
| Config Tab (List) | `list_config` | `update_main.go` | Keybindings YAML viewer/editor |
| Agent Details: Overview | `agent_details_overview` | `update_detail.go` | Agent info, recent commits |
| Agent Details: Commits | `agent_details_commits` | `update_detail.go` | Commit split-pane view |
| Agent Details: Logs | `agent_details_logs` | `update_detail.go` | Session logs with filtering |
| Agent Details: Notes | `agent_details_notes` | `update_detail.go` | Session notes |
| Agent Details: Actions | `agent_details_actions` | `update_detail.go` | All agent actions |
| Agent Details: Tasks | `agent_details_tasks` | `update_detail.go` | Agent tasks sorted by priority |
| Agent Details: Projects | `agent_details_projects` | `update_detail.go` | Project-specific tasks |

**Note:** `detail_*` aliases are maintained for backward compatibility with existing YAML files.

---

## 4. Input Handling Lifecycle

### 4.1 Keyboard Flow

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    }
}
```

### 4.2 Resolver Integration (List View)

```go
func (m *Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    m.keybindResolver.SetContext("list", activeTabName)

    action, found := m.keybindResolver.Resolve(msg, m.modalManager.IsActive())
    if found {
        switch action {
        case "new_session": ...
        case "new_ticket": ...
        case "down": ...
        case "up": ...
        case "select": ...
        case "refresh": ...
        }
    }

    // Static fallback for j/k if resolver didn't match
    if !m.modalManager.IsActive() {
        switch msg.String() {
        case "j", "down": m.selectedIndex++
        case "k", "up":   m.selectedIndex--
        }
    }
    return m, nil
}
```

---

## 5. Global vs Scoped Keybindings

| Scope Type | Description | Examples |
|-------------|--------------|-----------|
| **Global** | Always active, regardless of view. Hardcoded or resolver-based. | `Ctrl+C` (hardcoded), `Ctrl+H` (hardcoded), `r` (resolver-based). |
| **View-Scoped** | Active within a major view like List or Detail. | `tab`/`shift+tab` for changing tabs. |
| **Tab-Scoped** | Active within a specific tab of a view. | `j/k` navigation in `list_overview` only. |
| **Modal-Scoped** | Active when a modal form is open. | Currently modal components handle their own keys; resolver modal scope not yet populated. |

**Important:** Some global keys are **hardcoded and bypass the resolver entirely** in `update_root.go`:
- `Ctrl+C` and `Shift+Q` → Quit (always works)
- `Ctrl+H` → Help (always works, cannot be remapped via YAML)

Other global keys (`quit`, `refresh`, etc.) **flow through the resolver** and can be customized in YAML:
```go
action, found := m.keybindResolver.Resolve(msg, m.modalManager.IsActive())
if !found {
    return m, nil
}

switch action {
case "quit":
    return m, tea.Quit
case "refresh":
    m.statusMsg = "Refreshing..."
    return m, loadAgentsCmd(m.data.Agents)
}
```

---

## 6. The Help System

The **Help Modal** is generated dynamically from resolver data.

### When user presses `Ctrl+H` or `?`:
1. `handleGlobalKeys()` detects the action `"help"`.
2. Calls `m.openContextualHelp()`.
3. `openContextualHelp()` fetches keybindings for current context:
   ```go
   bindings := m.keybindResolver.GetScopeKeybindings(m.currentScope)
   ```
4. ModalManager opens a Help modal and populates a scrollable `viewport.Model`.

### Footer Help
- Static text only: `"Ctrl-H: Help"`.
- No YAML dependency.
- Displayed via `renderFooter()`.

---

## 7. Config Tab and Validation

When user edits keybindings:
- The `Config` tab loads YAML into `keybindingsYAML`.
- If invalid, parser throws error → stored in `m.keybindingsError`.
- UI displays:
  ```
  ⚠️ Invalid keybindings.yaml: <error>
  ```
- Invalid YAML is **not applied**, resolver keeps last valid state.

---

## 8. Fallback Hierarchy

Order of evaluation for every keypress in `handleKeyPress()`:

1. **Modal active?** → Modal handler gets first priority, processes key or passes through.
2. **Hardcoded global keys** (in `handleGlobalKeys()`) → `Ctrl+C`, `Shift+Q`, `Ctrl+H`.
3. **Resolver lookup** (for global, view-scoped, and tab-scoped actions):
   - Checks current resolver context (set via `SetContext()`)
   - Returns `(action, found)` tuple
   - If found, switch on action name
   - If not found, return early
4. **Component-level fallback** (in individual handlers like `handleListKeys()`):
   - Hardcoded navigation (`j/k` for up/down, `h/l` for horizontal scroll)
   - Only applies when resolver didn't match the key
5. **No match:** → Ignore key, return unchanged model.

---

## 9. Adding New Actions

To add a new action (e.g. `"archive_agent"`):

1. **Add YAML entry** in appropriate scope:
   ```yaml
   list:
     overview:
       a: archive_agent
   ```
2. **Handle it** in code:
   ```go
   case "archive_agent":
       agent := m.SelectedAgent()
       if agent != nil {
           return m, m.archiveAgentCmd(agent.ID)
       }
   ```
3. **Update Help Modal**: Automatically included.

---

## 10. Common Failure Modes

| Symptom | Cause | Fix |
|----------|--------|-----|
| Keypress ignored | YAML missing mapping or scope wrong | Add binding to correct scope |
| Action found but nothing happens | No case in `switch action` | Add handling logic |
| “j/k” dead in overview | Resolver matches but doesn’t handle; fallback missing | Keep fallback block below resolver |
| Help modal empty | YAML invalid or resolver context unset | Verify `SetContext()` before Resolve() |
| Config tab error | YAML parse failure | Fix syntax or spacing |

---

## 11. Debugging Tips

- Use `debugf("Key pressed: %s", msg.String())` to confirm input.
- To trace resolver matches:
  ```go
  action, found := m.keybindResolver.Resolve(msg, modalActive)
  debugf("Resolver: action=%s, found=%v", action, found)
  ```
- To verify current scope before resolving:
  ```go
  m.keybindResolver.SetContext("list", "overview")
  debugf("Scope set to: list_overview")
  ```
- Check keybindings YAML validity in the Config tab — invalid YAML is rejected with error message.
- Use the Config tab to reload YAML on the fly; press `r` to refresh resolver with new bindings.

---

## 12. Future Enhancements

| Idea | Benefit |
|------|----------|
| Modal-specific scopes | Different keymaps for forms/help views |
| Keybinding categories | Group keys visually in Help modal |
| Schema validation | Detect duplicates or unmapped actions at load |
| Mouse binding support | Integrate `tea.MouseMsg` into resolver |

---

## 13. Known Issues and Limitations

| Issue | Impact | Workaround |
|-------|--------|-----------|
| **Modal scope not populated** | Modal keybindings cannot be configured via YAML; modals handle keys internally | Modal components implement their own key handling |
| **Ctrl+H hardcoded** | Help key cannot be remapped via YAML | Always available, but cannot customize trigger key |
| **No per-action validation** | Invalid action names in YAML are silently ignored (no warning at load time) | Test keybindings in Config tab to verify actions are recognized |
| **Scope naming transition** | Code now uses `agent_details_*` but YAML still supports `detail.*` for backward compatibility | Both naming schemes work; prefer `agent_details` in new configs |

---

### TL;DR

- `Resolver` maps keys → actions from YAML based on context (underscore format: `view_tab`).
- Agent details view uses `agent_details_*` scopes; `detail_*` aliases maintained for backward compatibility.
- Some global keys (`Ctrl+C`, `Ctrl+H`) are hardcoded and bypass the resolver.
- Each tab/view sets its own resolver context via `SetContext(view, tab)`.
- Fallbacks keep navigation reliable when resolver doesn't match.
- Help modal shows real mappings dynamically from resolver data.
