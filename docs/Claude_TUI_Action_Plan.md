
# Eye-in-the-Sky TUI — Scoped Keybindings, Modals, and Forms
**Audience:** junior developer  
**Goal:** implement context-scoped keybindings, modal system, and form-driven actions for Sessions and Tickets; integrate AppleScript, Taskwarrior, YAML-configured keymaps; keep Elm-style Update loop pure and predictable.

---

## 0. Principles to follow
- Model is the single source of truth; no globals; no mutations in View.
- All side effects via `tea.Cmd`; never block Update.
- The app routes keys by scope: Global; Modal; View; Tab.
- Modals pause the app; only modal keys work until closed.
- Keep functions small; add unit tests for Update handlers.

---

## 1. Target user-facing behavior
### Top-level pages (ViewList)
- Overview: list of agents; `n` opens New Session form; `enter` or `right` shows Agent Detail.
- Project: tickets; `n` opens New Ticket form; `enter` opens selected ticket.
- Claude: browse `~/.claude`; `v` validate; `e` edit; `r` refresh; sticky validation status.
- Usage: read-only; `j/k` and arrows scroll; page up/down work.

### Agent Detail (ViewDetail)
- Tabs: Overview; Commits; Logs; Notes; Actions; Tasks; each with its own keymap; `esc/q` returns to Overview list with selection preserved.

### Modals
- Form modal for New Session; New Ticket; Help overlay; Error modal.
- While a modal is open: only `Tab`, `Shift+Tab`, `Enter`, `Esc`, `?`, `Shift+H`, `Ctrl+C`, `Shift+Q` work.
- Help always shows the current scope first; if opened from a modal, it shows modal keys.

---

## 2. Repo structure changes
```
internal/ui/
  keybindings/
    loader.go          # YAML load; normalized map
    resolver.go        # scope resolution helpers
  modal/
    modal.go           # base Modal model; routing gate
    form.go            # FormSpec; textinputs; validation; submit
    help.go            # context help renderer; reads key map
    error.go           # simple error modal
  views/
    list/
      overview.go      # handleOverviewKeys; list rendering
      project.go       # handleProjectKeys; tickets
      claude.go        # handleClaudeKeys; viewport update
      usage.go         # handleUsageKeys; viewport only
    detail/
      overview.go
      commits.go
      logs.go
      notes.go
      actions.go
      tasks.go
  cmds/
    applescript.go     # ExecProcess helpers; capture stdout/stderr
    taskwarrior.go     # task add; parse UUID
    sqlite.go          # optional chained inserts
```

---

## 3. Data types
```go
// Key scopes
type KeyScope int
const (
  ScopeGlobal KeyScope = iota
  ScopeModal
  ScopeView
  ScopeTab
)

// YAML model
type KeybindingsConfig struct {
  Global  map[string][]string            `yaml:"global"`
  List    map[string]map[string][]string `yaml:"list"`   // overview, project, claude, usage
  Detail  map[string]map[string][]string `yaml:"detail"` // overview, commits, logs, notes, actions, tasks
}

// Form spec
type FormField struct {
  Name, Label, Placeholder string
  Multiline bool
  Required  bool
}

type FormSpec struct {
  Title  string
  Fields []FormField
  Submit string
  Cancel string
  Source string // "new-session", "new-ticket"
}

type formSubmittedMsg struct {
  Source string
  Data   map[string]string
}

// Modal model
type ModalType int
const (
  ModalNone ModalType = iota
  ModalForm
  ModalHelp
  ModalError
)

type Modal struct {
  Active bool
  Type   ModalType
  Title  string
  // for form
  FormSpec FormSpec
  Inputs   []textinput.Model
  FocusIdx int
  // for help/error
  Content string
}
```

---

## 4. YAML example (put under `~/.eye-in-the-sky/keybindings.yaml`)
```yaml
keybindings:
  global:
    quit: ["ctrl+c", "shift+q"]
    help: ["?", "shift+h"]

  list:
    overview:
      new_session: ["n"]
      select: ["enter", "right"]
      down: ["j", "down"]
      up: ["k", "up"]

    project:
      new_ticket: ["n"]
      select: ["enter"]
      down: ["j", "down"]
      up: ["k", "up"]

    claude:
      validate: ["v"]
      edit: ["e"]
      refresh: ["r"]
      open: ["enter"]
      down: ["j", "down"]
      up: ["k", "up"]
      page_down: ["pgdown"]
      page_up: ["pgup"]

    usage:
      down: ["j", "down"]
      up: ["k", "up"]
      page_down: ["pgdown"]
      page_up: ["pgup"]

  detail:
    overview:
      select: ["enter"]
    commits:
      refresh: ["r"]
    logs:
      refresh: ["r"]
    notes:
      new_note: ["n"]
    actions:
      rerun: ["r"]
    tasks:
      new_task: ["n"]
```

---

## 5. Loading keybindings
- Implement `keybindings/loader.go`:
  - Locate YAML; if missing, create defaults.
  - Unmarshal; normalize to lowercase; flatten into a resolver map:
    `map[KeyScope]map[string]map[string]bool` where `map[action][key]` is true.
- Implement `keybindings/resolver.go`:
  - `func Resolve(m *Model, msg tea.KeyMsg) (scope KeyScope, action string, ok bool)`
  - Order: if modal active → modal keys; else tab; else view; else global.
  - Provide helpers: `IsQuit`, `IsHelp`, etc.

---

## 6. Modal subsystem
### 6.1 Base routing gate
- In `Model.Update`, if `m.modal.Active` then route all messages to `modal.Update` and return.
- Modal recognizes: `tab`, `shift+tab`, `enter`, `esc`, `?`, `shift+h`, quit keys.

### 6.2 Form modal
- Use `bubbles/textinput` for each field.
- Focus management:
  ```go
  // Tab
  inputs[focus].Blur()
  focus = (focus + 1) % len(inputs)
  inputs[focus].Focus()
  // Shift+Tab
  inputs[focus].Blur()
  focus = (focus - 1 + len(inputs)) % len(inputs)
  inputs[focus].Focus()
  ```
- Enter behavior:
  - If focused field is multiline, insert newline; else submit.
  - Validate required fields; if missing, show inline error.
  - On submit: emit `formSubmittedMsg{Source: spec.Source, Data: values}`; close modal.

### 6.3 Help modal
- Reads current scope from resolver; lists actions and keys from YAML; render in two columns; scroll if long using `viewport`.

### 6.4 Error modal
- Simple; shows `Title` and `Content`; Esc closes.

---

## 7. Page handlers
### 7.1 List view
- Split per tab file: `views/list/overview.go`, `project.go`, `claude.go`, `usage.go`.
- Each provides:
  ```go
  func HandleKeys(m *Model, msg tea.KeyMsg) (tea.Model, tea.Cmd)
  func View(m *Model) string
  ```

### 7.2 Detail view
- Same pattern under `views/detail/*`.

### 7.3 Routing
- In root `handleKeyPress`:
  1) Check quit; 2) Check help; 3) If modal active → modal; else route to current view; then into current tab handler.

---

## 8. Commands layer
### 8.1 AppleScript
```go
func RunAppleScript(args ...string) tea.Cmd {
  var stdoutBuf, stderrBuf bytes.Buffer
  cmd := exec.Command("osascript", args...)
  cmd.Stdout = &stdoutBuf
  cmd.Stderr = &stderrBuf
  return tea.ExecProcess(cmd, func(err error) tea.Msg {
    return appleScriptDoneMsg{Out: stdoutBuf.String(), ErrText: stderrBuf.String(), Err: err}
  })
}
```
- On success: status footer “Session window opened”.
- On error: open `ModalError` with `ErrText`.

### 8.2 Taskwarrior
- `task add <desc> project:<proj> +tag1 +tag2` via ExecProcess.
- Parse stdout for UUID; post `ticketCreatedMsg{UUID: ...}`; show footer.

### 8.3 Optional chained ops
- Use `tea.Batch` to follow success with: SQLite insert; window focus script; UI refresh.

---

## 9. Forms to implement
### New Session (Overview tab)
- `FormSpec`: Title “New Session”; one field `description` required; Submit “Create”.
- Submit flow:
  1) Close modal; set `m.loading = true` and show spinner next to status.
  2) `RunAppleScript("create_session.scpt", description)`
  3) On success: `m.loading = false`; footer “Session window opened”.
  4) Optional: `tea.Batch(insertSessionRowCmd(), focusWindowCmd())`.

### New Ticket (Project tab)
- Fields: `description` required; `tags` optional; project auto-populated but editable.
- Submit: `task add` command; parse UUID; footer “Ticket created: <uuid>”.

---

## 10. Claude tab fixes
- Use `m.claudeViewport, cmd = m.claudeViewport.Update(msg)` for j/k; capture `cmd`.
- Bounds checks in key handlers only; clamp selection after refresh.
- Validation status persists in `m.claudeValidStatus` until next validation or file load.

---

## 11. Help modal content
- Show: Scope name; actions; keys; short descriptions.
- Source of truth: YAML; never hardcode bindings into help text.
- If opened from inside a form, show modal keybindings first.

---

## 12. Testing checklist
- Unit tests for: resolver precedence; modal Tab/Shift+Tab wrapping; form submit messages; Claude bounds checks; detail Esc returns to list with selection intact.
- Golden tests for Help rendering using a small YAML.
- Manual test scripts:
  - Create session; force AppleScript failure; verify error modal.
  - Create ticket; confirm UUID parse.
  - Toggle help from within a form; Esc twice returns to page.

---

## 13. Acceptance criteria
- Keys are scoped; `n` does different things in Overview vs Project; Usage is read-only.
- Modals block page keys; Tab and Shift+Tab move focus; Enter submits; Esc cancels.
- Help modal shows the current scope; opened by `?` or `Shift+H` anywhere; Help over a form shows modal keys.
- YAML changes are persisted only when user saves in Config tab; atomic write; errors are visible.
- AppleScript and Taskwarrior run async; success or failure is visible to the user; UI never freezes.
- Claude tab scrolls properly; validation status is sticky.

---

## 14. Milestones
1. **Scaffolding**: folders; types; loader; resolver; modal base; wire modal gate.  
2. **Help modal**: render keys from YAML for each scope.  
3. **Overview page**: `n` opens New Session form; AppleScript integration; spinner; error modal.  
4. **Project page**: `n` opens New Ticket form; Taskwarrior integration; UUID parse.  
5. **Claude fixes**: viewport Update pattern; sticky validation.  
6. **Detail view**: independent keymaps for subtabs; Esc/q return behavior.  
7. **Config tab**: view and edit YAML; atomic save.  
8. **Tests**: resolver; modal; forms; Claude; detail navigation.

---

## 15. Gotchas to avoid
- Do not mutate state in `View()`.  
- Do not call `exec.Command(...).Run()` in Update; always `tea.ExecProcess`.  
- Do not log errors only; show visible error modals for user-facing failures.  
- Do not let modals leak focus; always Blur previous input before focusing next.  
- Do not forget WindowSizeMsg; resize viewports and modal boxes on change.

---

## 16. Dev ready tasks (copy into issues)
- [ ] Create keybindings loader and resolver; defaults if YAML missing.
- [ ] Add modal gate; implement Modal, FormSpec, form update; help and error modals.
- [ ] Wire global keys: Ctrl+C; Shift+Q; `?`; `Shift+H`.
- [ ] Implement Overview.HandleKeys; open New Session form; AppleScript Cmd; spinner; error modal.
- [ ] Implement Project.HandleKeys; open New Ticket form; Taskwarrior Cmd; UUID parse.
- [ ] Fix Claude scrolling using viewport.Update; sticky validation.
- [ ] Implement Detail handlers for tabs; independent keymaps; Esc returns to Overview.
- [ ] Add Config tab to view and edit YAML; atomic save.
- [ ] Unit tests for resolver precedence; modal focus; form submit; Claude bounds checks.
- [ ] Manual test plan; record GIFs.
