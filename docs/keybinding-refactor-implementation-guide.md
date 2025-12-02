# Keybinding Refactor - Implementation Guide

## Overview

This guide provides step-by-step instructions to complete the keybinding refactor. The core infrastructure is done. Remaining work is structured in phases with exact code patterns.

---

## Phase 2: Create Remaining Tag Files

### 2.1 Create `keymap_navigation.go`

This tag is shared by multiple views (main, details, logs, contexts, actions).

```go
package dashboard

import "github.com/jroimartin/gocui"

// RegisterNavigationKeys registers shared navigation keybindings
// Tag: "navigation" - Reusable scroll/cursor bindings for list-like views
func (a *App) RegisterNavigationKeys() {
	// Cursor/Scroll up - j
	a.keys.Register(KeySpec{
		View: "", // Will be bound per-view
		Key:  'j',
		Mod:  gocui.ModNone,
		Fn:   a.navigationUp, // Generic handler
		Tag:  "navigation",
	})

	// Cursor/Scroll down - k
	a.keys.Register(KeySpec{
		View: "",
		Key:  'k',
		Mod:  gocui.ModNone,
		Fn:   a.navigationDown,
		Tag:  "navigation",
	})

	// Arrow up
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyArrowUp,
		Mod:  gocui.ModNone,
		Fn:   a.navigationUp,
		Tag:  "navigation",
	})

	// Arrow down
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyArrowDown,
		Mod:  gocui.ModNone,
		Fn:   a.navigationDown,
		Tag:  "navigation",
	})

	// Page down - Space
	a.keys.Register(KeySpec{
		View: "",
		Key:  ' ',
		Mod:  gocui.ModNone,
		Fn:   a.pageDown,
		Tag:  "navigation",
	})

	// Page up - b
	a.keys.Register(KeySpec{
		View: "",
		Key:  'b',
		Mod:  gocui.ModNone,
		Fn:   a.pageUp,
		Tag:  "navigation",
	})

	// Page down - PgDn
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyPgdn,
		Mod:  gocui.ModNone,
		Fn:   a.pageDown,
		Tag:  "navigation",
	})

	// Page up - PgUp
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyPgup,
		Mod:  gocui.ModNone,
		Fn:   a.pageUp,
		Tag:  "navigation",
	})

	// Top - g (Vim-style)
	a.keys.Register(KeySpec{
		View: "",
		Key:  'g',
		Mod:  gocui.ModNone,
		Fn:   a.jumpToTop,
		Tag:  "navigation",
	})

	// Bottom - G (Vim-style)
	a.keys.Register(KeySpec{
		View: "",
		Key:  'G',
		Mod:  gocui.ModNone,
		Fn:   a.jumpToBottom,
		Tag:  "navigation",
	})

	// Home
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyHome,
		Mod:  gocui.ModNone,
		Fn:   a.jumpToTop,
		Tag:  "navigation",
	})

	// End
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyEnd,
		Mod:  gocui.ModNone,
		Fn:   a.jumpToBottom,
		Tag:  "navigation",
	})
}

// navigationUp handles up movement - cursor in main, scroll in others
func (a *App) navigationUp(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == viewMain {
		return a.cursorUp(g, v)
	}
	return a.scrollUp(g, v)
}

// navigationDown handles down movement - cursor in main, scroll in others
func (a *App) navigationDown(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == viewMain {
		return a.cursorDown(g, v)
	}
	return a.scrollDown(g, v)
}
```

**NOTE**: Navigation tag uses `View: ""` (empty) because it needs to work on whatever view is current. The key: handlers check `v.Name()` to determine behavior.

### 2.2 Create `keymap_details.go`

```go
package dashboard

import "github.com/jroimartin/gocui"

// RegisterDetailsKeys registers keybindings for the details view
// Tag: "details" - Agent detail view actions
func (a *App) RegisterDetailsKeys() {
	// Back to main - q
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'q',
		Mod:  gocui.ModNone,
		Fn:   a.closeDetails,
		Tag:  "details",
	})

	// Refresh details - R
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'R',
		Mod:  gocui.ModNone,
		Fn:   a.refreshDetails,
		Tag:  "details",
	})

	// View all contexts - c
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'c',
		Mod:  gocui.ModNone,
		Fn:   a.viewAllContexts,
		Tag:  "details",
	})

	// View all contexts - x (alternative)
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'x',
		Mod:  gocui.ModNone,
		Fn:   a.viewAllContexts,
		Tag:  "details",
	})

	// View all actions/logs - l
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'l',
		Mod:  gocui.ModNone,
		Fn:   a.viewAllActions,
		Tag:  "details",
	})

	// Start session - s
	a.keys.Register(KeySpec{
		View: "details",
		Key:  's',
		Mod:  gocui.ModNone,
		Fn:   a.startSession,
		Tag:  "details",
	})

	// Go to window - w
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'w',
		Mod:  gocui.ModNone,
		Fn:   a.goToWindow,
		Tag:  "details",
	})

	// Navigation keys will be added via "navigation" tag
}
```

### 2.3 Create `keymap_edit.go`

```go
package dashboard

import "github.com/jroimartin/gocui"

// RegisterEditKeys registers keybindings for edit mode
// Tag: "edit" - Text input mode with readline support
func (a *App) RegisterEditKeys() {
	// Submit - Enter
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyEnter,
		Mod:  gocui.ModNone,
		Fn:   a.saveDescription,
		Tag:  "edit",
	})

	// Cancel - Esc
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyEsc,
		Mod:  gocui.ModNone,
		Fn:   a.cancelEdit,
		Tag:  "edit",
	})

	// Readline: Clear line - Ctrl+U
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlU,
		Mod:  gocui.ModNone,
		Fn:   a.readlineClearLine,
		Tag:  "edit",
	})

	// Readline: Delete word - Ctrl+W
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlW,
		Mod:  gocui.ModNone,
		Fn:   a.readlineDeleteWord,
		Tag:  "edit",
	})

	// Readline: Jump to start - Ctrl+A
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlA,
		Mod:  gocui.ModNone,
		Fn:   a.readlineJumpStart,
		Tag:  "edit",
	})

	// Readline: Jump to end - Ctrl+E
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlE,
		Mod:  gocui.ModNone,
		Fn:   a.readlineJumpEnd,
		Tag:  "edit",
	})

	// Allow arrow keys, backspace, delete - handled by gocui.DefaultEditor
	// No need to bind these; DefaultEditor handles them when view is editable
}

// Readline helper functions

func (a *App) readlineClearLine(g *gocui.Gui, v *gocui.View) error {
	v.Clear()
	v.SetCursor(0, 0)
	return nil
}

func (a *App) readlineDeleteWord(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	if cx == 0 {
		return nil
	}

	line := v.Buffer()
	if len(line) == 0 {
		return nil
	}

	// Find start of current word (scan backward to space or start)
	start := cx - 1
	for start > 0 && line[start] != ' ' {
		start--
	}
	if line[start] == ' ' {
		start++
	}

	// Delete from start to cursor
	newLine := line[:start] + line[cx:]
	v.Clear()
	v.Write([]byte(newLine))
	v.SetCursor(start, cy)

	return nil
}

func (a *App) readlineJumpStart(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	v.SetCursor(0, cy)
	return nil
}

func (a *App) readlineJumpEnd(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	line := v.Buffer()
	v.SetCursor(len(line), cy)
	return nil
}
```

### 2.4 Create `keymap_global.go`

```go
package dashboard

import "github.com/jroimartin/gocui"

// RegisterGlobalKeys registers global keybindings (active in all views)
// Tag: "global" - Only Ctrl+C and ?
func (a *App) RegisterGlobalKeys() {
	// Quit - Ctrl+C (global emergency exit)
	a.keys.Register(KeySpec{
		View: "", // Global
		Key:  gocui.KeyCtrlC,
		Mod:  gocui.ModNone,
		Fn:   a.quit,
		Tag:  "global",
	})

	// Help - ? (show keybinding help modal)
	a.keys.Register(KeySpec{
		View: "", // Global
		Key:  '?',
		Mod:  gocui.ModNone,
		Fn:   a.showHelp,
		Tag:  "global",
	})
}

// showHelp displays a help modal with current keybindings
func (a *App) showHelp(g *gocui.Gui, v *gocui.View) error {
	// TODO: Implement help modal
	// For now, just show a message
	return a.showMessage("Press Ctrl+C to quit, q to go back")
}
```

---

## Phase 3: Update app.go

### 3.1 Add Fields to App Struct

In `internal/dashboard/app.go`, add these fields:

```go
type App struct {
	gui                *gocui.Gui
	db                 *database.DB
	config             *Config
	keymap             *Keymap  // Old config-based keymap (keep for now)
	keys               *KeyRegistry  // NEW
	modes              *ModeManager  // NEW
	views              *ViewManager  // NEW
	agents             []*database.Agent
	selectedIdx        int
	filter             string
	ticker             *time.Ticker
	quitChan           chan bool
	currentAgentID     string
	currentCompactions []*database.Compaction
	compactionLineMap  map[int]int
	currentContexts    []*database.SessionContext
	contextLineMap     map[int]int
	sectionLines       map[string]int
	showRefreshDot     bool
}
```

### 3.2 Update NewApp Function

```go
func NewApp(db *database.DB) (*App, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}

	config, keymap, err := loadConfig()
	if err != nil {
		return nil, err
	}

	app := &App{
		gui:      g,
		db:       db,
		config:   config,
		keymap:   keymap,
		keys:     NewKeyRegistry(g),      // NEW
		modes:    NewModeManager(),       // NEW
		agents:   []*database.Agent{},
		filter:   config.DefaultFilter,
		quitChan: make(chan bool),
	}

	app.views = NewViewManager(app)  // NEW

	// Register all keybinding tags
	app.RegisterGlobalKeys()
	app.RegisterMainKeys()
	app.RegisterDetailsKeys()
	app.RegisterNavigationKeys()
	app.RegisterEditKeys()

	// Activate initial tags
	app.keys.BindTag("global")     // Always active
	app.keys.BindTag("main")       // Start in main view
	app.keys.BindTag("navigation") // Navigation active in main

	g.SetManagerFunc(app.layout)
	g.Cursor = true

	return app, nil
}
```

### 3.3 Remove Old setupKeybindings Call

Find and remove the call to `app.setupKeybindings()` in `Run()` method (if it exists).

---

## Phase 4: Update View Transitions

### 4.1 Update `viewAgentDetails` in actions.go

Replace the keybinding setup section with:

```go
func (a *App) viewAgentDetails(agentID string) error {
	// ... existing code to fetch agent data ...

	// Switch to details view with proper tag management
	err := a.views.SwitchToView(
		[]string{"main", "navigation"},      // Unbind these
		[]string{"details", "navigation"},   // Bind these
		ModeDetails,                          // Push this mode
		"details",                            // Set this view current
	)
	if err != nil {
		return err
	}

	// ... rest of existing rendering code ...

	return nil
}
```

### 4.2 Update `closeDetails` in actions.go

Replace the entire function:

```go
func (a *App) closeDetails(g *gocui.Gui, v *gocui.View) error {
	// Clear current agent ID
	a.currentAgentID = ""

	// Delete the details view
	if err := g.DeleteView("details"); err != nil && err != gocui.ErrUnknownView {
		return err
	}

	// Use view manager to restore main view
	err := a.views.PopView(
		[]string{"details", "navigation"},  // Unbind these
		[]string{"main", "navigation"},     // Rebind these
		viewMain,                            // Return to this view
	)
	if err != nil {
		return err
	}

	// Restore status bar
	statusView, err := g.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [q] Quit | [R] Refresh | [a] Toggle All | [e] Edit | [d] Done | [D] Archive | [n] New | [r] Resume | [s] Start | [w] Window | [L] Logs")

	return a.refreshAgents()
}
```

### 4.3 Update `editDescription` in keymap.go

Replace the entire function:

```go
func (a *App) editDescription(g *gocui.Gui, v *gocui.View) error {
	// Get selected agent
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}
	agent := a.agents[a.selectedIdx]

	// Get current description
	currentDesc := ""
	if agent.Description != nil {
		currentDesc = *agent.Description
	}

	// Create input view
	maxX, maxY := g.Size()
	inputWidth := 60
	inputHeight := 3
	x0 := (maxX - inputWidth) / 2
	y0 := (maxY - inputHeight) / 2

	inputView, err := g.SetView("edit-input", x0, y0, x0+inputWidth, y0+inputHeight)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	inputView.Title = " Edit Description (Enter to save, Esc to cancel) "
	inputView.Editable = true
	inputView.Editor = gocui.DefaultEditor
	inputView.Clear()
	inputView.Write([]byte(currentDesc))
	inputView.SetCursor(len(currentDesc), 0)

	// Switch to edit mode with proper tag management
	return a.views.SwitchToView(
		[]string{"main", "navigation"},  // Unbind main view tags
		[]string{"edit"},                 // Bind edit mode tags
		ModeEdit,                         // Push edit mode
		"edit-input",                     // Set edit-input as current view
	)
}
```

### 4.4 Update `cancelEdit` in keymap.go

Replace with:

```go
func (a *App) cancelEdit(g *gocui.Gui, v *gocui.View) error {
	return a.views.CloseModalView(
		"edit-input",                     // View to delete
		[]string{"edit"},                 // Tags to unbind
		viewMain,                         // Return to main view
		[]string{"main", "navigation"},   // Rebind these tags
	)
}
```

### 4.5 Update Modal Closes

Update `closeModalView` in `view_contexts.go`:

```go
func (a *App) closeModalView(viewName string) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return a.views.CloseModalView(
			viewName,                         // Modal view name
			[]string{"navigation"},           // Modal tags (just navigation for scrolling)
			"details",                        // Return to details
			[]string{"details", "navigation"}, // Restore details tags
		)
	}
}
```

---

## Phase 5: Config System

### 5.1 Create Default keys.json

Create `cmd/dashboard/config/keys.json`:

```json
{
  "global": {
    "quit": ["Ctrl+C"],
    "help": ["?"]
  },
  "main": {
    "quit": ["q"],
    "refresh": ["R"],
    "toggle_filter": ["a"],
    "edit_description": ["e"],
    "mark_done": ["d"],
    "archive": ["D"],
    "new_session": ["n"],
    "resume_session": ["r"],
    "start_session": ["s"],
    "go_to_window": ["w"],
    "show_logs": ["L"],
    "view_details": ["<Enter>"]
  },
  "details": {
    "back": ["q"],
    "refresh": ["R"],
    "view_contexts": ["c", "x"],
    "view_logs": ["l"],
    "start_session": ["s"],
    "go_to_window": ["w"]
  },
  "navigation": {
    "up": ["j", "<Up>"],
    "down": ["k", "<Down>"],
    "page_up": ["b", "<PgUp>"],
    "page_down": ["<Space>", "<PgDn>"],
    "top": ["g", "<Home>"],
    "bottom": ["G", "<End>"]
  },
  "edit": {
    "submit": ["<Enter>"],
    "cancel": ["<Esc>"],
    "clear_line": ["Ctrl+U"],
    "delete_word": ["Ctrl+W"],
    "jump_start": ["Ctrl+A"],
    "jump_end": ["Ctrl+E"]
  }
}
```

### 5.2 Add Config Loader (Optional)

If you want to support loading keys from config:

```go
// In keymap.go or new config.go file

func (a *App) LoadKeysFromConfig(configPath string) error {
	// Read JSON file
	// Parse keys
	// For each key string, call ParseKey()
	// Register KeySpec
	// This is optional - hardcoded keys work fine
	return nil
}
```

---

## Phase 6: Testing

### Test 1: Edit Mode Isolation

```bash
go build -o bin/dashboard ./cmd/dashboard
./bin/dashboard
# Press 'e' on an agent
# Type: "Testing R a n stuff"
# Expected: Characters appear, NO refresh (R), NO toggle (a), NO new session (n)
# Press Enter to save
# Expected: Description updated
```

### Test 2: View Switching

```bash
./bin/dashboard
# Press Enter on an agent
# Press 'R' - Expected: Details refresh (not main refresh)
# Press 'q' - Expected: Return to main
# Press 'R' - Expected: Main refresh
```

### Test 3: Modal Lifecycle

```bash
./bin/dashboard
# Enter → details
# Press 'x' → contexts modal
# Press 'q' → back to details
# Press 'q' → back to main
# Expected: All keybindings work correctly at each level
```

### Test 4: Navigation Tag Reuse

```bash
./bin/dashboard
# Main view: Press 'j'/'k' - Expected: Cursor moves
# Enter details: Press 'j'/'k' - Expected: Scroll
# Press 'x' for contexts: Press 'j'/'k' - Expected: Scroll
```

---

## Cleanup

### Remove Old Code

1. Delete old `setupKeybindings()` function from old keymap.go
2. Delete old keybinding setup in `viewAgentDetails`
3. Delete old keybinding cleanup in `closeDetails`
4. Delete old edit mode keybinding code

### Verify Build

```bash
go build -o bin/dashboard ./cmd/dashboard
# Should compile without errors
```

---

## Summary Checklist

- [ ] Created keymap_navigation.go
- [ ] Created keymap_details.go
- [ ] Created keymap_edit.go with readline
- [ ] Created keymap_global.go
- [ ] Updated app.go with keys/modes/views fields
- [ ] Updated NewApp to register and bind tags
- [ ] Updated viewAgentDetails to use SwitchToView
- [ ] Updated closeDetails to use PopView
- [ ] Updated editDescription to use SwitchToView
- [ ] Updated cancelEdit to use CloseModalView
- [ ] Updated modal closes to use CloseModalView
- [ ] Created default keys.json (optional)
- [ ] Tested edit mode isolation
- [ ] Tested view switching
- [ ] Tested modal lifecycle
- [ ] Tested navigation tag reuse
- [ ] Removed old keybinding code
- [ ] Verified build succeeds

---

## Troubleshooting

**Problem**: "undefined: a.navigationUp"
**Solution**: Add the navigationUp/navigationDown functions from keymap_navigation.go

**Problem**: "undefined: a.readlineClearLine"
**Solution**: Add the readline helper functions from keymap_edit.go

**Problem**: Keys still fire in edit mode
**Solution**: Verify edit tag is bound and main tag is unbound in editDescription

**Problem**: Zombie bindings after view close
**Solution**: Ensure CloseModalView is called with correct tag lists

**Problem**: Navigation doesn't work
**Solution**: Verify "navigation" tag is in the bind list for each view
