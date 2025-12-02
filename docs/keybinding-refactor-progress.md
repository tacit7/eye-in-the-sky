# Keybinding Refactor Progress

## Completed (Phase 1 - Core Infrastructure)

### ✅ 1. modes.go
- Full mode stack implementation
- Modes: Main, Details, Logs, Edit, Cmdline, Tasks
- Push/Pop/Current/Previous methods
- Reset and Depth tracking

### ✅ 2. keymap.go
- KeyRegistry with tag-based binding system
- Register/BindTag/UnbindTag/RebindTag methods
- UnbindAll for cleanup
- ParseKey function supporting:
  - Literals (a, R, :)
  - Specials (<Up>, <Down>, <Enter>, <Esc>, <PageUp>, <PageDown>, <Tab>, <Backspace>, <Delete>, <Home>, <End>, <F1>-<F12>)
  - Modifiers (Ctrl+X, Alt+X)
  - Vim aliases (<CR>, <C-c>, <M-x>)

### ✅ 3. views.go
- ViewManager with lifecycle helpers
- SwitchToView for tag transitions
- PopView for back navigation
- CloseModalView centralized helper
- CurrentTag mapper

### ✅ 4. keymap_main.go
- Main view keybindings extracted
- All bindings use "main" tag
- Quit, View details, Edit, Refresh, Toggle, Archive, Done, New, Resume, Start, Window, Logs

## Remaining Work

### Phase 2: Tag Definitions
- [ ] keymap_details.go - details + navigation tags
- [ ] keymap_edit.go - edit mode + readline bindings
- [ ] keymap_navigation.go - shared scroll/cursor (j/k, arrows, PgUp/PgDn, g/G, Home/End)
- [ ] keymap_global.go - Only Ctrl+C and ?

### Phase 3: Config
- [ ] cmd/dashboard/config/keys.json - default key mappings
- [ ] Config loader with user override (~/.config/eye-in-the-sky/keys.json)
- [ ] External editor config (vim/vscode/nano)

### Phase 4: Integration
- [ ] Update app.go to add:
  ```go
  keys  *KeyRegistry
  modes *ModeManager
  views *ViewManager
  ```
- [ ] Replace setupKeybindings() with RegisterXXXKeys() calls
- [ ] Update all view creation (details, logs, contexts, actions modals)
- [ ] Update all modal close handlers to use views.CloseModalView()

### Phase 5: Edit Mode
- [ ] Implement readline keybindings (Ctrl+U, Ctrl+W, Ctrl+A, Ctrl+E)
- [ ] External editor integration (spawn vim/vscode)
- [ ] Test: Type in edit field without triggering app actions

### Phase 6: Testing & Cleanup
- [ ] Test edit mode isolation
- [ ] Test view switching (no zombie bindings)
- [ ] Test config loading with user overrides
- [ ] Regression pass on all features
- [ ] Remove old keybinding code from old keymap.go functions
- [ ] Delete old setupKeybindings function

## Key Design Decisions

1. **Two Independent Systems**: Mode stack (UI state) + Tags (input handling)
2. **Hard Switches**: Unbind old → Bind new (no priority overlay)
3. **Minimal Globals**: Only Ctrl+C (quit) and ? (help)
4. **Tag Reusability**: "navigation" tag used by main, details, logs
5. **Explicit Lifecycle**: Modal close handlers call centralized helper

## Next Steps

The immediate priority is to fix the edit mode bug. Options:

1. **Quick Fix**: Add inEditMode flag, check in all global handlers
2. **Proper Fix**: Complete Phase 2-4 above to use tag system

Recommendation: Complete the refactor properly to avoid future issues.

## File Structure

```
internal/dashboard/
├── modes.go              ✅ Mode stack
├── keymap.go             ✅ KeyRegistry + parser
├── views.go              ✅ Lifecycle helpers
├── keymap_main.go        ✅ Main view bindings
├── keymap_details.go     ⏳ Details + navigation
├── keymap_edit.go        ⏳ Edit mode + readline
├── keymap_navigation.go  ⏳ Shared nav bindings
├── keymap_global.go      ⏳ Ctrl+C, ?
└── app.go                ⏳ Needs updates
```

## Testing Strategy

1. Edit mode: Type letters → No refresh/toggle triggered
2. View switch: main → details → no zombie 'R' binding
3. Modal close: contexts → details → bindings restored
4. Config load: User keys override defaults
5. Regression: All existing functionality works
