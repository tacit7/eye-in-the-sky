# Claude Config Browser Tab - Implementation Guide

## Overview

The **Claude tab** is a Bubble Tea TUI component that lets users browse, view, validate, and edit Claude Code configuration files from `~/.claude` directory. It's integrated into the Eye in the Sky dashboard as the third tab alongside Overview and Project tabs.

**User experience:**
1. Press `c` from Overview tab to open Claude browser
2. Navigate files with j/k keys
3. Press Enter to view file content
4. Press j/k to scroll content, v to validate JSON, e to edit in $EDITOR
5. Press r to refresh, esc/q to go back

## Architecture Overview

### Bubble Tea Pattern

The implementation follows **Elm Architecture** (pure functional patterns):

```
User Input → tea.Msg
    ↓
Update(msg) → consumes message, returns (Model, tea.Cmd)
    ↓
Cmd executes → emits new tea.Msg
    ↓
View(model) → renders to string (pure, no mutations)
```

**Key principle:** All state changes happen via messages. Update() is the only place that mutates model. View() is pure and never changes state.

### File Organization

```
internal/ui/app/
├── claude_helpers.go          # Business logic (file I/O, validation)
├── claude_messages.go         # Message types and command factories
├── view_claude.go             # Rendering (two-pane layout)
├── model.go                   # State fields added
├── view_list.go               # Tab routing (case 2)
├── update_list.go             # Key handlers
└── update_root.go             # Message handlers, window resize
```

### State Model

```go
type Model struct {
    // Claude tab state
    claudeFiles         []ClaudeFile              // Files in ~/.claude
    claudeSelectedIndex int                       // Currently selected file (0-based)
    claudeContent       string                    // File content being viewed
    claudeViewport      viewport.Model            // Scrollable content viewer
    claudeValidStatus   string                    // JSON validation result
    claudeShowingContent bool                     // Toggle: file list vs content

    // ... other fields (agents, width, height, etc.)
}

type ClaudeFile struct {
    Name    string        // e.g., "settings.json"
    Path    string        // Full path e.g., "/Users/foo/.claude/settings.json"
    IsDir   bool          // true for directories like "hooks/"
    ModTime time.Time     // Last modified
}
```

### Message Types

```go
// File operations
type claudeFilesLoadedMsg struct {
    files []ClaudeFile
    err   error
}

type claudeFileContentLoadedMsg struct {
    content string
    path    string
    err     error
}

// Editor and validation
type claudeEditorClosedMsg struct {
    err error
}

type claudeValidationResultMsg struct {
    valid  bool
    status string
}
```

## Navigation Flow

### Two States

The tab toggles between two states:

**State 1: File List** (`claudeShowingContent == false`)
- Left pane: File list with selection
- Right pane: Help/instructions
- j/k: Navigate file list
- Enter: Load selected file → switch to State 2

**State 2: Content View** (`claudeShowingContent == true`)
- Left pane: File list (read-only)
- Right pane: File content in viewport
- j/k: Scroll content
- Esc/q: Back to State 1
- v: Validate JSON
- e: Open in editor

### Viewport Scrolling Pattern

**CRITICAL:** When scrolling viewport, use this pattern:

```go
// Correct - captures viewport update
var cmd tea.Cmd
m.claudeViewport, cmd = m.claudeViewport.Update(msg)
return m, cmd

// Wrong - viewport won't re-render
m.claudeViewport.LineDown(1)
return m, nil  // No cmd returned!
```

The viewport returns a command that tells Bubble Tea to repaint. Without it, scrolling looks frozen.

## Input Handling

### Key Handler Routing

```
handleListKeys() (update_list.go)
    ↓
if m.listTabs.ActiveIndex == 2 {  // Claude tab only
    switch msg.String() {
    case "j", "down":              // Navigate/scroll
    case "k", "up":
    case "enter":                  // Load file
    case "v":                       // Validate JSON
    case "e":                       // Edit in $EDITOR
    case "r":                       // Refresh
    case "esc", "q":                // Back
    case "pgup", "pgdown":          // Page scroll
    }
}
```

### Message Handling (update_root.go)

Messages come back from async operations:

```go
case claudeFilesLoadedMsg:
    // Update file list
    m.claudeFiles = msg.files
    // Clamp selection to valid range
    if m.claudeSelectedIndex >= len(m.claudeFiles) {
        m.claudeSelectedIndex = max(0, len(m.claudeFiles)-1)
    }

case claudeFileContentLoadedMsg:
    // Display file content
    m.claudeContent = msg.content
    m.claudeViewport.SetContent(msg.content)
    m.claudeViewport.GotoTop()

case claudeValidationResultMsg:
    m.claudeValidStatus = msg.status  // Show in right pane
```

## Rendering

### Two-Pane Layout

```go
// view_claude.go: renderClaudeTab()

leftPane := m.renderClaudeFileList()     // File list
rightPane := m.renderClaudeContentViewer() // Content or help

// Join horizontally with separator
return lipgloss.JoinHorizontal(lipgloss.Top,
    leftStyle.Render(leftPane),
    separator,
    rightStyle.Render(rightPane),
)
```

**Parent wrapper** (view_list.go): Adds the ContentBox border around entire Claude tab.

### Left Pane: File List

```go
// renderClaudeFileList() → string

// Title
"📁 ~/.claude"

// Files with selection indicators
"  📄 settings.json"     // Unselected
" >📄 mcp.json"          // Selected (reversed style)
"  📁 hooks"             // Directory

// Return joined string
```

### Right Pane: Content or Help

```go
if !claudeShowingContent {
    // Show instructions
    "Claude Config Browser

     j/k: Navigate
     Enter: View file
     v: Validate JSON
     ..."
} else {
    // Show viewport with file content
    header := "✓ Valid JSON"  // or validation error
    return header + "\n\n" + m.claudeViewport.View()
}
```

## Business Logic

### File Scanning (claude_helpers.go)

```go
func scanClaudeDir() ([]ClaudeFile, error) {
    // Open ~/.claude
    // Read entries
    // Filter: .json, .md files + directories
    // Sort: directories first, then by name
    // Return ClaudeFile slice
}
```

### JSON Validation

```go
func validateJSON(content string) (bool, string) {
    // Try unmarshaling
    // Return (valid, status message)
    // Example: (false, "Invalid JSON: unexpected EOF")
}
```

### File Loading

```go
// In update_list.go when Enter pressed:
return m, loadClaudeFileContentCmd(filePath)

// In claude_messages.go:
func loadClaudeFileContentCmd(path string) tea.Cmd {
    return func() tea.Msg {
        content := readFile(path)
        prettified := prettifyJSON(content)  // Auto-format if JSON
        return claudeFileContentLoadedMsg{
            content: prettified,
            path: path,
        }
    }
}
```

## Window Resize Handling

When terminal is resized (tea.WindowSizeMsg):

```go
// update_root.go: Update()
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height

    // Claude tab viewport resize
    if m.currentView == ViewList && m.listTabs.ActiveIndex == 2 {
        rightPaneWidth := m.width - 40  // Left pane ≈ 30 cols + separator
        available := m.height - 8       // Header, tabs, footer

        m.claudeViewport.Width = rightPaneWidth
        m.claudeViewport.Height = available
    }
```

## Current Implementation Status

### ✅ Implemented

- [x] File list rendering with selection
- [x] Two-pane Lipgloss layout
- [x] File content viewing with viewport
- [x] j/k navigation (file list and content)
- [x] Enter to load file
- [x] JSON validation (v key)
- [x] External editor launch (e key) with tea.ExecProcess
- [x] File refresh (r key)
- [x] Escape to go back (esc/q)
- [x] Page scrolling (pgup/pgdn)
- [x] Window resize handling
- [x] Viewport re-rendering (fixed in commit 67839a3)
- [x] Selection bounds checking (fixed in commit 67839a3)

### ❌ Not Implemented

1. **Directory Navigation** - Pressing Enter on directories (hooks/, agents/) doesn't open them
   - Would need: Separate state for directory path, modify scanClaudeDir to accept parent path
   - Complexity: Medium

2. **Mouse Support** - Can't click files to select/open
   - Would need: Click handler in handleListClick(), coordinate mapping
   - Complexity: Low-Medium

3. **File Filtering/Search** - No way to filter file list by name
   - Would need: Input field state, filter logic in renderClaudeFileList
   - Complexity: Medium-High (requires text input handling)

4. **Auto-Reload on File Change** - Files edited externally don't auto-refresh
   - Would need: fsnotify watcher, watch command
   - Complexity: Medium

5. **Non-JSON Syntax Highlighting** - Only JSON gets colorized
   - Would need: Extend colorizeJSON to handle .md, .yaml, etc.
   - Complexity: Low

### 🐛 Known Issues

None currently. Previous viewport and bounds-checking bugs were fixed in commit 67839a3.

## Testing Guide

### Build and Run

```bash
go build -o bin/eye-ui ./cmd/eye-ui
./bin/eye-ui
```

### Manual Test Checklist

- [ ] Press `c` from Overview → see Claude tab with file list
- [ ] Press `j` → selection moves down, visual feedback immediate
- [ ] Press `k` → selection moves up
- [ ] Press `Enter` on file → right pane shows content
- [ ] Press `j` while viewing → content scrolls smoothly
- [ ] Press `v` → validation status appears (green or red)
- [ ] Press `e` → editor opens, edit file, save and exit → content reloads
- [ ] Press `r` → file list refreshes
- [ ] Press `esc` → back to file list
- [ ] Resize terminal → layout adjusts properly
- [ ] No crashes or panic when navigating

### Debug Commands

```bash
# Check files exist
ls ~/.claude/*.json ~/.claude/*.md

# Check database populated
sqlite3 ~/.config/eye-in-the-sky/agents.db 'select count(*) from agents;'

# Run with debug output
DEBUG=1 ./bin/eye-ui 2>&1 | grep -i claude
```

## Code Patterns to Follow

### 1. Message-Driven Updates

```go
// Update() - only place where state changes
case claudeFileContentLoadedMsg:
    m.claudeContent = msg.content
    m.claudeViewport.SetContent(msg.content)
    return m, nil

// NOT in View() or anywhere else
```

### 2. Async Operations Return Commands

```go
// Correct
return m, loadClaudeFileContentCmd(path)

// Wrong
content := readFile(path)  // Blocks!
m.claudeContent = content
return m, nil
```

### 3. Viewport Update Pattern

```go
// Always capture returned command
var cmd tea.Cmd
m.claudeViewport, cmd = m.claudeViewport.Update(msg)
return m, cmd

// Not just:
m.claudeViewport.LineDown(1)
return m, nil
```

### 4. View Functions Are Pure

```go
// Good - reads state only
func (m *Model) renderClaudeFileList() string {
    return strings.Join(lines, "\n")
}

// Bad - mutates state
func (m *Model) renderClaudeTab() string {
    m.renderCount++  // NO!
    return ...
}
```

### 5. Boundary Checks in Handlers

```go
// In Update() handlers - check bounds BEFORE using index
if m.claudeSelectedIndex >= 0 && m.claudeSelectedIndex < len(m.claudeFiles) {
    file := m.claudeFiles[m.claudeSelectedIndex]
    // safe to use file
}

// In View() - trust state invariants are maintained by Update()
```

## Common Pitfalls

1. **Forgetting viewport.Update() cmd** - Scrolling appears frozen
   - Fix: Always return the cmd from viewport.Update()

2. **State mutations in View()** - Causes non-deterministic rendering
   - Fix: Only read state in View(), mutate only in Update()

3. **No boundary checks in handlers** - Panic on out-of-bounds access
   - Fix: Check `claudeSelectedIndex < len(claudeFiles)` before access

4. **Not setting claudeShowingContent before rendering** - Right pane shows help instead of content
   - Fix: Set `m.claudeShowingContent = true` before calling cmd

5. **Forgetting selection clamp after refresh** - Old index out of bounds
   - Fix: When file list changes, clamp index to new range

## Future Enhancement Approaches

### Directory Navigation

```go
// Add to Model:
claudeCurrentPath string  // e.g., "/Users/foo/.claude/hooks"

// Modify scanClaudeDir to take path parameter:
func scanClaudeDir(path string) ([]ClaudeFile, error)

// In renderClaudeFileList, show breadcrumb:
"📁 ~/.claude / hooks"

// On Enter on directory:
m.claudeCurrentPath = newPath
return m, m.loadClaudeFilesCmd()
```

### Mouse Support

```go
// In update_mouse.go: handleListClick()

// Calculate if click is in left pane (file list)
if msg.X < leftPaneWidth {
    // Map Y coordinate to file index
    rowClicked := (msg.Y - contentTop) / rowHeight
    m.claudeSelectedIndex = rowClicked
}

// Double-click to open
if time.Since(lastClickTime) < 500ms && lastClickX == msg.X {
    // Treat as Enter
}
```

### File Filtering

```go
// Add to Model:
claudeFilter string

// In view render:
renderClaudeFileList() {
    filtered := []ClaudeFile{}
    for _, f := range m.claudeFiles {
        if strings.Contains(f.Name, m.claudeFilter) {
            filtered = append(filtered, f)
        }
    }
}

// Key handler for '/' to enter filter mode
case "/":
    m.filterMode = true
    m.claudeFilter = ""
```

## Quick Reference

### State Fields (model.go)

```go
claudeFiles         []ClaudeFile    // Current file list
claudeSelectedIndex int             // 0-based index into claudeFiles
claudeContent       string          // Full file content
claudeViewport      viewport.Model  // Scroll position and height
claudeValidStatus   string          // "✓ Valid JSON" or error
claudeShowingContent bool           // false=list, true=content
```

### Message Types (claude_messages.go)

| Message | Emitted By | Fields |
|---------|-----------|--------|
| `claudeFilesLoadedMsg` | scanClaudeDir cmd | files, err |
| `claudeFileContentLoadedMsg` | loadClaudeFileContentCmd | content, path, err |
| `claudeEditorClosedMsg` | openClaudeFileInEditor | err |
| `claudeValidationResultMsg` | validateClaudeFileCmd | valid, status |

### Command Factories (claude_messages.go)

```go
m.loadClaudeFilesCmd()                    // Scan ~/.claude
loadClaudeFileContentCmd(path)             // Read and prettify file
validateClaudeFileCmd(content)             // Validate JSON
openClaudeFileInEditor(path)               // Launch $EDITOR
```

### Key Handlers (update_list.go, lines 145-224)

| Key | Action | State |
|-----|--------|-------|
| j/down | Navigate file list OR scroll content | Both |
| k/up | Navigate file list OR scroll content | Both |
| Enter | Load selected file | File list only |
| v | Validate JSON | Content only |
| e | Edit in $EDITOR | Both (uses selected file) |
| r | Refresh file list | Both |
| esc/q | Back to file list | Content only |
| pgup/pgdn | Page scroll | Content only |

## Getting Started

1. Read this file top-to-bottom for context
2. Review CLAUDE_TAB_SPEC.md for original requirements
3. Look at `internal/ui/app/claude_*.go` files
4. Study the Bubble Tea patterns in Usage tab (similar viewport pattern)
5. Run `./bin/eye-ui` and test manually
6. Make changes following patterns in "Code Patterns" section
7. Test thoroughly before committing

## Resources

- **Bubble Tea docs**: https://github.com/charmbracelet/bubbletea
- **Lipgloss docs**: https://github.com/charmbracelet/lipgloss
- **Viewport pattern**: See `internal/ui/app/update_list.go:84-99` (Usage tab)
- **Similar component**: Project tab in `view_project.go`
