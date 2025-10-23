# CLAUDE TAB — TECHNICAL SPEC & ACTION PLAN

## OVERVIEW
The **Claude tab** provides a structured, readable view into Anthropic Claude’s local configuration directory (`~/.claude`).  
It enables:
- Browsing JSON config files (`settings.json`, `mcp.json`, `hooks/`, etc.)
- Viewing prettified, syntax-colored JSON
- Validating files with detailed error feedback
- Opening configs in the user’s `$EDITOR`
- Basic status summary (active model, hook count, last modified)

---

## 1. STACK REQUIREMENTS

| Purpose | Library | Reason |
|----------|----------|--------|
| Core TUI Framework | **charmbracelet/bubbletea** | Main event loop, state management |
| Layout & Styling | **charmbracelet/lipgloss** | Layouts, borders, padding, color styles |
| Lists / Navigation | **charmbracelet/bubbles/list** | File list sidebar with key navigation |
| Scrollable Views | **charmbracelet/bubbles/viewport** | Right-side content window (JSON viewer) |
| Keybinding Display | **charmbracelet/bubbles/help** | Optional hints bar (bottom) |
| Syntax Highlighting | **alecthomas/chroma** | JSON colorization without manual parsing |
| File I/O | **os**, **io/ioutil**, **encoding/json** | Read/validate files |
| External Editor | **os/exec** | Launch `$EDITOR` (`nvim` fallback) |

---

## 2. FILE STRUCTURE

```
internal/ui/tabs/claude/
│
├── model.go          # ClaudeTabModel struct + init
├── update.go         # Handles messages, keypress, editor
├── view.go           # Layout + Lipgloss rendering
├── validate.go       # JSON validation and schema check
└── fileutils.go      # Read, prettify, colorize functions
```

---

## 3. DATA MODEL

```go
type ClaudeTabModel struct {
    files   list.Model
    viewer  viewport.Model
    content string
    active  string
    status  string
    width, height int
}
```

### State Transitions
- **Idle** → **FileSelected** → **ContentLoaded**
- **ValidateRequested** → **Validated**
- **EditorOpen** → **ReloadAfterEdit**

---

## 4. COMMANDS & MESSAGES

| Message | Source | Purpose |
|----------|---------|----------|
| `loadFileMsg` | On `Enter` | Reads + prettifies JSON |
| `validateFileMsg` | On `v` | Validates JSON + updates status |
| `editorClosedMsg` | After `e` | Re-reads file, revalidates |
| `reloadFilesMsg` | On `r` | Refreshes file list |
| `errorMsg` | Any | Displays red status footer |

---

## 5. UI LAYOUT

### Layout Engine: **Lipgloss**
- Two-column grid layout:
  - Left: list.Model (files)
  - Right: viewport.Model (content)
- Border between panes
- Footer status line with color-coded validation message

```go
layout := lipgloss.JoinHorizontal(
    lipgloss.Top,
    leftPaneStyle.Render(m.files.View()),
    rightPaneStyle.Render(m.viewer.View()),
)
```

---

## 6. KEYBINDINGS

| Key | Action | Implementation |
|-----|---------|----------------|
| `j/k` | Navigate file list | `bubbles/list` |
| `Enter` | Load file into viewer | Sends `loadFileMsg` |
| `v` | Validate JSON | Runs `validateFileMsg` |
| `e` | Open `$EDITOR` | `exec.Command(os.Getenv("EDITOR"))` |
| `r` | Reload directory | Rescan `~/.claude` |
| `q` | Exit tab | Returns control to root view |

---

## 7. FILE RENDERING

### Prettify JSON
```go
var out bytes.Buffer
err := json.Indent(&out, rawBytes, "", "  ")
if err != nil {
    return string(rawBytes), err
}
return out.String(), nil
```

### Colorize JSON (Optional)
```go
lexer := lexers.Get("json")
iterator, _ := lexer.Tokenise(nil, content)
style := styles.Get("dracula")
formatter := chromahtml.New(chromahtml.WithClasses())
var buf bytes.Buffer
formatter.Format(&buf, style, iterator)
```

### Non-JSON Files
Just read and display plain text.

---

## 8. VALIDATION LOGIC

```go
func validateJSON(content []byte) (bool, string) {
    var js map[string]interface{}
    err := json.Unmarshal(content, &js)
    if err != nil {
        return false, fmt.Sprintf("Invalid JSON: %v", err)
    }
    return true, "Valid JSON"
}
```

Optional schema checks:
- Missing keys (`model`, `api_key`, `hooks`)
- Empty arrays (`hooks: []`)
- Show warnings in yellow

---

## 9. EXTERNAL EDITOR SUPPORT

```go
func openInEditor(path string) tea.Cmd {
    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "nvim"
    }
    cmd := exec.Command(editor, path)
    cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
    return tea.ExecProcess(cmd, func(err error) tea.Msg {
        return editorClosedMsg{err: err}
    })
}
```

After editor exit:
- Re-read file
- Validate again
- Show result in footer

---

## 10. IMPLEMENTATION ORDER (for devs)

1. **Setup Bubble Tea boilerplate** under `internal/ui/tabs/claude`.
2. **Implement file scanning**:
   ```go
   func scanClaudeDir() []list.Item
   ```
   Filter `~/.claude` for `.json`, `.md`, `hooks/`, `agents/`.
3. **Wire up list + viewport** in `Init()` and `Update()`.
4. **Add key handling** (j/k/Enter/v/e/r/q).
5. **Render layout** with Lipgloss grid.
6. **Add validation output** footer under viewer.
7. **Implement editor launch + post-validate**.
8. **Colorize JSON** using `chroma` (optional but recommended).
9. **Polish visuals** (Lipgloss borders, headers).
10. **Integrate tab switcher** with root model.

---

## 11. FUTURE EXTENSIONS
- Add **“Sessions” subtab** (token counts, last message timestamp).
- Integrate `claude` CLI for live status via `exec.Command("claude", "status")`.
- Add **Codex engine** support via `~/.codex` folder mirroring this layout.
- Implement **fsnotify** to auto-refresh on config change.

---

### Implementation Philosophy
Keep it *deterministic*:
- All reads from disk.
- No async network calls.
- All state transitions handled via `tea.Cmd`.
- Each action produces one message and one state update.  

No “magic.” If it’s missing, render an empty view; don’t guess.
