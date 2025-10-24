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
    e: edit
    ctrl+s: save
    esc: cancel
```

Each top-level key (`global`, `list`, etc.) defines **scope groups**.  
Scopes may nest further — e.g. `"list.overview"`.

---

### 3.2 Resolver (`keybindResolver`)

Responsible for:
- Tracking **current scope** (`list`, `detail`, etc.).
- Looking up key-to-action mappings.
- Returning `(action, found)`.

Typical call:
```go
action, found := m.keybindResolver.Resolve(msg, false)
```

The resolver maintains internal state:
```go
type Resolver struct {
    Bindings map[string]map[string]string // scope → key → action
    Context  string                       // "list.overview"
}
```

#### Core methods
| Method | Purpose |
|--------|----------|
| `SetContext(view, tab string)` | e.g., `"list"`, `"overview"` → context `"list.overview"` |
| `Resolve(msg tea.KeyMsg, allowFallback bool)` | Match keypress within current context |
| `Reload()` | Reload YAML and rebuild binding maps |
| `GetScopeKeybindings(scope string)` | For Help modal rendering |

---

### 3.3 Scopes

Each **view** defines its own logical scope.  

| View / Context | Example Context String | Source File | Description |
|----------------|------------------------|--------------|--------------|
| Overview Tab | `list.overview` | `update_list.go` | Shows agents list |
| Project Tab | `list.project` | `update_list.go` | Project tasks, CLAUDE.md, etc. |
| Claude Tab | `list.claude` | `update_list.go` | File explorer for Claude configs |
| Usage Tab | `list.usage` | `update_list.go` | Token usage + viewport |
| Config Tab | `list.config` | `update_list.go` | Keybindings YAML viewer/editor |
| Detail View | `detail.overview`, etc. | `update_detail.go` | Per-agent detail tabs |

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

    action, found := m.keybindResolver.Resolve(msg, false)
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

    // Static fallback for j/k if resolver swallowed input
    if m.listTabs.ActiveIndex == 0 {
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
| **Global** | Always active, regardless of view. Handled in `handleGlobalKeys()`. | Quit (`Ctrl+C`), Help (`Ctrl+H`), Refresh (`r`). |
| **View-Scoped** | Active within a major view like List or Detail. | `tab`/`shift+tab` for changing tabs. |
| **Tab-Scoped** | Active within a specific tab of a view. | `j/k` navigation in Overview tab only. |
| **Modal-Scoped** | Active when a modal form is open. | `tab` to cycle fields, `esc` to cancel. |

Global bindings are hardcoded fallbacks in `update_root.go`:
```go
case "ctrl+c", "shift+q":
    return m, tea.Quit
case "ctrl+h", "?":
    return m.openContextualHelp()
```

Everything else flows through the resolver.

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

Order of evaluation for every keypress:

1. **Modal active?** → Modal handles all keys.  
2. **Global keys?** → Quit, Help, Refresh.  
3. **Resolver lookup:**  
   - If found and handled → stop.  
   - If found but unhandled → continue.  
4. **Static fallback** (hardcoded j/k navigation, etc.).  
5. **No match:** → Ignore key.

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
  debugf("Resolver: action=%s, found=%v", action, found)
  ```
- To verify scope:
  ```go
  debugf("Resolver context: %s", m.keybindResolver.CurrentScope())
  ```
- Use the Config tab to reload YAML on the fly.

---

## 12. Future Enhancements

| Idea | Benefit |
|------|----------|
| Modal-specific scopes | Different keymaps for forms/help views |
| Keybinding categories | Group keys visually in Help modal |
| Schema validation | Detect duplicates or unmapped actions at load |
| Mouse binding support | Integrate `tea.MouseMsg` into resolver |

---

### TL;DR

- `Resolver` maps keys → actions from YAML based on context.  
- Global keys (quit/help/refresh) are hardcoded.  
- Each tab/view sets its own resolver context.  
- Fallbacks keep navigation reliable when YAML is missing or mismatched.  
- Help modal shows real mappings dynamically; footer is static.
