# Eye in the Sky – Keybinding Refactor Action Plan (v1)

## Overview
This plan replaces the current hybrid global/view keybinding system with a **scoped keybinding architecture** for `gocui`.  
The goal is to eliminate conflicts, simplify mode handling, and allow configuration-driven keymaps from JSON.

---

## 1. Problem Summary

**Current issues:**
- Global bindings fire during input/edit modes.
- Duplicate handlers between main and detail views.
- No centralized registry of keys, making cleanup error-prone.

**Root cause:**
`gocui` global keybindings apply to all views. This leads to collisions when users type in editable fields or modals.

---

## 2. Objective

Implement a **Scoped Keybinding Manager** that registers keymaps per-view (e.g., `main`, `details`, `edit`) and deactivates them on view switch.  
Only critical actions like `Ctrl+C` (quit) and `?` (help) remain global.

---

## 3. Architecture

### Core Components
- `internal/ui/keymap.go`: Registry for all bindings
- `internal/ui/modes.go`: Manages active mode stack
- `internal/ui/views.go`: Handles creation/destruction of views with scoped bindings

### Data Flow
```
App.Start()
  -> loadConfig("keys.json")
  -> keyRegistry.RegisterTag("main", bindings.Main)
  -> gui.SetCurrentView("main")

User presses 'd' in main view
  -> KeyRegistry dispatches bound handler for tag "main"
  -> handler runs (no globals interfere)

User enters detail view
  -> keyRegistry.UnbindTag("main")
  -> keyRegistry.BindTag("details")
```

---

## 4. Implementation Steps

### Step 1 — Create KeyRegistry
```go
type KeySpec struct {
    View string
    Key interface{}
    Mod gocui.Modifier
    Fn func(*gocui.Gui, *gocui.View) error
    Tag string
}

type KeyRegistry struct {
    gui *gocui.Gui
    specs []KeySpec
    active map[string][]KeySpec
}
```

**Methods:**
- `Bind(spec KeySpec)` → Register a single binding.
- `UnbindTag(tag string)` → Remove all bindings under tag.
- `RebindTag(tag string)` → Reactivate stored bindings for tag.

---

### Step 2 — Replace Ad-hoc Bindings
Move all `gui.SetKeybinding` calls from `main.go` and `actions.go` into centralized setup files:

```
internal/ui/keymap_main.go
internal/ui/keymap_details.go
internal/ui/keymap_edit.go
```

Each defines:
```go
func RegisterMainKeys(r *KeyRegistry) {
  r.Bind(KeySpec{View: "main", Key: 'q', Fn: quit, Tag: "main"})
  r.Bind(KeySpec{View: "main", Key: 'r', Fn: refresh, Tag: "main"})
}
```

---

### Step 3 — Add Mode Stack
```go
type Mode int
const (
    ModeMain Mode = iota
    ModeDetails
    ModeEdit
)

type ModeManager struct {
    stack []Mode
}

func (m *ModeManager) Push(mode Mode) { m.stack = append(m.stack, mode) }
func (m *ModeManager) Pop() Mode { ... }
func (m *ModeManager) Current() Mode { ... }
```

Used in `App.SwitchView()`:
```go
func (a *App) SwitchView(tag string) {
    a.keys.UnbindAll()
    a.keys.RebindTag(tag)
    a.modes.Push(tagToMode(tag))
}
```

---

### Step 4 — Modify Views
Every `SetView()` call should:
- Unbind previous tag
- Register bindings for the new one
- Set focus

```go
func (a *App) showDetails() error {
    a.keys.UnbindTag("main")
    a.keys.RebindTag("details")
    g.SetCurrentView("details")
}
```

---

### Step 5 — Configurable Keymaps
Keys read from `config/eye-in-the-sky/keys.json`.

**Example:**
```json
{
  "main": {
    "refresh": ["r", "R"],
    "quit": ["q", "Ctrl+C"],
    "details": ["Enter"],
    "tasks": ["t"]
  },
  "details": {
    "back": ["q", "Esc"],
    "scroll_down": ["j"],
    "scroll_up": ["k"],
    "view_logs": ["l"]
  }
}
```

A JSON parser converts string keys (`"Ctrl+C"`, `"<Up>"`) to `gocui.Key` constants.

---

### Step 6 — Remove Global Bindings
Only retain:
- `Ctrl+C` → quit
- `?` → show help modal

Everything else is registered under scoped tags (`main`, `details`, `edit`, etc.).

---

### Step 7 — Cleanup Helpers
Add safe destroy methods:
```go
func (r *KeyRegistry) UnbindAll() {
    for tag := range r.active {
        r.UnbindTag(tag)
    }
}
```

Ensure all modal and edit views call this before closing.

---

### Step 8 — Testing
- Enter edit mode → Type letters → No refresh or toggle triggered.
- Switch views → Keys change context instantly.
- JSON rebind test → Modify `keys.json`, restart, verify dynamic changes.
- Press `Ctrl+C` → Works anywhere.

---

## 5. Benefits
✅ No more keybinding collisions  
✅ Dynamically configurable keys via JSON  
✅ Clean separation per view  
✅ Easier testing and extensibility  
✅ Compatible with both Vim and non-Vim keymaps

---

## 6. Deliverables
- [ ] `internal/ui/keymap.go` (registry + parser)
- [ ] `internal/ui/keymap_main.go`
- [ ] `internal/ui/keymap_details.go`
- [ ] `internal/ui/keymap_edit.go`
- [ ] `config/eye-in-the-sky/keys.json` default map
- [ ] Tests for bind/unbind/view switching

---

## 7. Future Enhancements
- Optional mode indicators in the status bar.
- Support for user-defined macros.
- Runtime reloading of keybindings (`:reload-keys` command).

---

## 8. Summary
The new system removes all but two global bindings, replaces them with scoped, tag-based registration, and adds a JSON-driven configuration model.  
This gives predictable behavior across modes while keeping the flexibility to redefine controls without recompiling.
