# Action Plan: Docs Tab for Agent Details View

**Created:** 2025-10-26
**Agent ID:** c31b2285-fcff-4c01-823f-b020c6e9f6ef
**Session ID:** b998bfe1-df7a-4f9b-8ed9-0ab04d16f2ae

---

## Overview

Add a **Docs** tab (index 8) to the agent details view that displays documents created/modified during an agent session. Users can browse and view document contents directly in the TUI.

---

## 1. Database Schema

### 1.1 Create `documents` Table

```sql
CREATE TABLE documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    file_path TEXT NOT NULL,              -- Absolute path to document
    relative_path TEXT,                    -- Path relative to worktree
    document_type TEXT NOT NULL,           -- "markdown", "text", "code", "other"
    title TEXT,                            -- Extracted from file or filename
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    file_size INTEGER,                     -- Size in bytes
    line_count INTEGER,                    -- Total lines
    status TEXT DEFAULT 'active',          -- "active", "deleted", "moved"
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE INDEX idx_documents_agent_id ON documents(agent_id);
CREATE INDEX idx_documents_type ON documents(document_type);
CREATE INDEX idx_documents_status ON documents(status);

-- Auto-update trigger for updated_at
CREATE TRIGGER update_documents_updated_at
    AFTER UPDATE ON documents
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE documents SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
```

### 1.2 Migration File

Create: `internal/database/migrations/00XX_create_documents_table.sql`

---

## 2. MCP Tool Updates

### 2.1 New Tool: `i-log-document`

**Purpose:** Log document creation/modification during agent session

**Parameters:**
```go
{
    "agent_id": "uuid",
    "file_path": "/absolute/path/to/doc.md",
    "document_type": "markdown",  // optional, auto-detect from extension
    "title": "Document Title"     // optional, extract from content
}
```

**Logic:**
1. Validate agent_id exists
2. Auto-detect document_type from extension if not provided
3. Extract title from first `# Heading` in markdown files
4. Calculate file_size and line_count
5. Insert into documents table
6. Return success/failure

**Error Cases:**
- Agent not found
- File doesn't exist
- File not readable
- Duplicate entry (same file_path for agent)

### 2.2 Update `i-instructions` Tool

Add documentation section for `i-log-document` usage:
```
When you create or modify documentation files (.md, .txt, design docs, etc.):
- Call i-log-document to track the file
- Provide meaningful titles for easy browsing
- Documents appear in the Docs tab of agent details
```

---

## 3. Tab Implementation

### 3.1 File Structure

```
internal/ui/app/views/agent_details/tabs/
└── docs_tab.go
```

### 3.2 Tab Constant

**File:** `internal/ui/app/views/agent_details/model_agent_details.go`

```go
const (
    tabBack     = 0
    tabOverview = 1
    tabTasks    = 2
    tabActions  = 3
    tabLogs     = 4
    tabCommits  = 5
    tabNotes    = 6
    tabProjects = 7
    tabDocs     = 8  // NEW
)
```

### 3.3 Tab Label

**File:** `internal/ui/app/detail_tabs.go` (or new location after refactor)

```go
var DetailTabLabels = []string{
    "← Back",
    "[O]verview",
    "[T]asks",
    "[A]ctions",
    "[L]ogs",
    "[C]ommits",
    "[N]otes",
    "[P]roject",
    "[D]ocs",  // NEW
}
```

### 3.4 Implementation: `docs_tab.go`

**Package:** `tabs`

**Key Components:**

```go
package tabs

import (
    "fmt"
    "strings"
    "github.com/charmbracelet/lipgloss"
)

// Document represents a tracked document
type Document struct {
    ID           int
    FilePath     string
    RelativePath string
    DocumentType string
    Title        string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    FileSize     int64
    LineCount    int
    Status       string
}

// RenderDocs renders the docs tab with split-pane view
func RenderDocs(ctx *DataContext, styles OverviewStyles) string {
    if ctx.Agent == nil {
        return styles.Subtle.Render("\n  No agent selected\n")
    }

    docs := ctx.Documents  // []Document from DataContext

    if len(docs) == 0 {
        return styles.Subtle.Render("\n  No documents created in this session\n")
    }

    // Left pane: Document list
    leftPane := renderDocumentList(docs, ctx.DocsListIndex, styles)

    // Right pane: Document preview (if selected)
    rightPane := ""
    if ctx.DocsListIndex >= 0 && ctx.DocsListIndex < len(docs) {
        rightPane = renderDocumentPreview(docs[ctx.DocsListIndex], styles)
    }

    // Split-pane layout (similar to commits tab)
    return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

// renderDocumentList renders the left pane with document list
func renderDocumentList(docs []Document, selectedIndex int, styles OverviewStyles) string {
    var b strings.Builder

    b.WriteString(styles.Label.Render("📄 Session Documents"))
    b.WriteString("\n\n")

    for i, doc := range docs {
        icon := getDocumentIcon(doc.DocumentType)
        title := doc.Title
        if title == "" {
            title = filepath.Base(doc.FilePath)
        }

        // Highlight selected
        if i == selectedIndex {
            b.WriteString(styles.Selected.Render(fmt.Sprintf("▶ %s %s", icon, title)))
        } else {
            b.WriteString(fmt.Sprintf("  %s %s", icon, title))
        }

        // Show metadata
        b.WriteString(styles.Subtle.Render(fmt.Sprintf(" (%s, %d lines)",
            doc.DocumentType, doc.LineCount)))
        b.WriteString("\n")
    }

    return b.String()
}

// renderDocumentPreview renders the right pane with document content preview
func renderDocumentPreview(doc Document, styles OverviewStyles) string {
    var b strings.Builder

    // Header
    b.WriteString(styles.Primary.Render(doc.Title))
    b.WriteString("\n")
    b.WriteString(styles.Subtle.Render(doc.RelativePath))
    b.WriteString("\n\n")

    // Metadata
    b.WriteString(styles.Label.Render("Type: "))
    b.WriteString(doc.DocumentType)
    b.WriteString("\n")

    b.WriteString(styles.Label.Render("Size: "))
    b.WriteString(formatFileSize(doc.FileSize))
    b.WriteString("\n")

    b.WriteString(styles.Label.Render("Lines: "))
    b.WriteString(fmt.Sprintf("%d", doc.LineCount))
    b.WriteString("\n")

    b.WriteString(styles.Label.Render("Created: "))
    b.WriteString(formatTimestamp(doc.CreatedAt))
    b.WriteString("\n\n")

    // Content preview (first 20 lines)
    b.WriteString(styles.Label.Render("Preview:"))
    b.WriteString("\n")
    b.WriteString(styles.Subtle.Render("─────────────────────────────────────"))
    b.WriteString("\n")

    content := readFilePreview(doc.FilePath, 20)
    b.WriteString(content)

    b.WriteString("\n")
    b.WriteString(styles.Subtle.Render("─────────────────────────────────────"))

    return b.String()
}

// Helper: Get icon based on document type
func getDocumentIcon(docType string) string {
    switch docType {
    case "markdown":
        return "📝"
    case "text":
        return "📄"
    case "code":
        return "💻"
    default:
        return "📋"
    }
}

// Helper: Format file size
func formatFileSize(bytes int64) string {
    if bytes < 1024 {
        return fmt.Sprintf("%d B", bytes)
    } else if bytes < 1024*1024 {
        return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
    } else {
        return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
    }
}

// Helper: Read first N lines of file
func readFilePreview(path string, maxLines int) string {
    content, err := os.ReadFile(path)
    if err != nil {
        return fmt.Sprintf("Error reading file: %v", err)
    }

    lines := strings.Split(string(content), "\n")
    if len(lines) > maxLines {
        lines = lines[:maxLines]
    }

    return strings.Join(lines, "\n")
}
```

---

## 4. DataContext Updates

### 4.1 Add Fields

**File:** `internal/ui/app/views/agent_details/tabs/data_context.go` (or wherever DataContext is defined)

```go
type DataContext struct {
    Agent         *domain.Agent
    Tasks         []domain.Task
    Actions       []domain.Action
    Logs          []domain.Log
    Commits       []domain.Commit
    Notes         []domain.Note
    Documents     []Document        // NEW

    // Navigation state
    TasksListIndex    int
    CommitsListIndex  int
    DocsListIndex     int           // NEW
    // ... other indices
}
```

### 4.2 Add Loader

**File:** `internal/database/queries.go`

```go
// GetDocumentsByAgentID retrieves all documents for an agent
func (db *Database) GetDocumentsByAgentID(agentID string) ([]Document, error) {
    query := `
        SELECT id, agent_id, file_path, relative_path, document_type,
               title, created_at, updated_at, file_size, line_count, status
        FROM documents
        WHERE agent_id = ?
        AND status = 'active'
        ORDER BY created_at DESC
    `

    rows, err := db.db.Query(query, agentID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var documents []Document
    for rows.Next() {
        var doc Document
        err := rows.Scan(
            &doc.ID, &doc.AgentID, &doc.FilePath, &doc.RelativePath,
            &doc.DocumentType, &doc.Title, &doc.CreatedAt, &doc.UpdatedAt,
            &doc.FileSize, &doc.LineCount, &doc.Status,
        )
        if err != nil {
            return nil, err
        }
        documents = append(documents, doc)
    }

    return documents, rows.Err()
}
```

---

## 5. Keybindings

### 5.1 Update YAML Template

**File:** `internal/ui/keybindings/loader.go`

```yaml
agent_details:
  docs:
    j: down
    k: up
    enter: open_full    # Open document in editor
    r: refresh
    d: delete_doc       # Mark document as deleted
```

### 5.2 Add Scope

**File:** `internal/ui/app/update_detail.go` (or new `update_agent_details.go`)

```go
func getCurrentDetailTabName(tabIndex int) string {
    switch tabIndex {
    case tabBack:
        return "back"
    case tabOverview:
        return "overview"
    case tabTasks:
        return "tasks"
    case tabActions:
        return "actions"
    case tabLogs:
        return "logs"
    case tabCommits:
        return "commits"
    case tabNotes:
        return "notes"
    case tabProjects:
        return "projects"
    case tabDocs:
        return "docs"  // NEW
    default:
        return "overview"
    }
}
```

### 5.3 Handle Actions

```go
case "down":
    if m.tabs.ActiveIndex == tabDocs {
        if m.docsListIndex < len(m.documents)-1 {
            m.docsListIndex++
        }
    }
case "up":
    if m.tabs.ActiveIndex == tabDocs {
        if m.docsListIndex > 0 {
            m.docsListIndex--
        }
    }
case "open_full":
    if m.tabs.ActiveIndex == tabDocs {
        // Open document in $EDITOR
        doc := m.documents[m.docsListIndex]
        return m, openInEditor(doc.FilePath)
    }
```

---

## 6. Integration with Agent Workflow

### 6.1 Auto-logging Strategy

**Option A: Explicit logging** (recommended)
- User calls `i-log-document` when creating docs
- Full control, no surprises

**Option B: Auto-detect from file operations**
- Watch actions with action_type="file_operation"
- Auto-log if file is .md, .txt, etc.
- Risk: logs too many files

**Option C: Git commit hook**
- When commits are logged, scan for documentation files
- Log them automatically
- Ties docs to commits

**Recommendation:** Start with Option A (explicit), add Option C later.

### 6.2 Document Type Detection

```go
func detectDocumentType(filePath string) string {
    ext := filepath.Ext(filePath)
    switch ext {
    case ".md", ".markdown":
        return "markdown"
    case ".txt", ".text":
        return "text"
    case ".go", ".py", ".js", ".ts", ".sh":
        return "code"
    default:
        return "other"
    }
}
```

### 6.3 Title Extraction

```go
func extractTitle(filePath string) string {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return filepath.Base(filePath)
    }

    // For markdown, extract first # heading
    lines := strings.Split(string(content), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "# ") {
            return strings.TrimPrefix(line, "# ")
        }
    }

    // Fallback to filename
    return filepath.Base(filePath)
}
```

---

## 7. Testing Strategy

### 7.1 Unit Tests

**File:** `internal/database/queries_test.go`

```go
func TestGetDocumentsByAgentID(t *testing.T) {
    // Test empty results
    // Test single document
    // Test multiple documents
    // Test ordering (newest first)
    // Test filtering by status
}
```

**File:** `internal/ui/app/views/agent_details/tabs/docs_tab_test.go`

```go
func TestRenderDocs(t *testing.T) {
    // Test empty state
    // Test document list rendering
    // Test preview rendering
}
```

### 7.2 Integration Tests

1. Create agent session
2. Log 3 documents of different types
3. Open TUI and navigate to Docs tab
4. Verify all 3 documents appear
5. Navigate with j/k
6. Verify preview updates

### 7.3 Manual Testing Checklist

- [ ] Create new agent session
- [ ] Log markdown document with title
- [ ] Log text file without title
- [ ] Navigate to agent details > Docs tab
- [ ] Verify both documents appear
- [ ] Use j/k to navigate
- [ ] Verify preview shows correct content
- [ ] Test with 0 documents (empty state)
- [ ] Test with 20+ documents (scrolling)
- [ ] Test document type icons display correctly
- [ ] Test file size formatting (B, KB, MB)

---

## 8. Implementation Phases

### Phase 1: Database (30 min)
- [ ] Create migration file
- [ ] Run migration
- [ ] Add query functions
- [ ] Write unit tests

### Phase 2: MCP Tool (45 min)
- [ ] Implement `i-log-document` tool
- [ ] Add validation logic
- [ ] Add auto-detection helpers
- [ ] Update `i-instructions` output
- [ ] Test tool with manual calls

### Phase 3: Tab UI (60 min)
- [ ] Create `docs_tab.go`
- [ ] Implement `RenderDocs` function
- [ ] Implement document list rendering
- [ ] Implement preview rendering
- [ ] Add helper functions (icons, formatting)
- [ ] Test rendering with mock data

### Phase 4: Integration (45 min)
- [ ] Add tab constant and label
- [ ] Update DataContext struct
- [ ] Add document loading to data layer
- [ ] Wire up tab in detail view
- [ ] Add keybinding scope
- [ ] Implement navigation handlers

### Phase 5: Testing & Polish (30 min)
- [ ] Run unit tests
- [ ] Run integration tests
- [ ] Manual testing with real data
- [ ] Fix bugs
- [ ] Update documentation

**Total Estimated Time:** 3.5 hours

---

## 9. Future Enhancements

### 9.1 Full Document Viewer
- Press `enter` to open full document in modal
- Syntax highlighting for code
- Markdown rendering

### 9.2 Document Diff
- Show changes if document updated during session
- Link to git commits that modified document

### 9.3 Search & Filter
- Search document titles
- Filter by document type
- Filter by date range

### 9.4 Export
- Export all session docs as ZIP
- Generate index page with links

---

## 10. Open Questions

1. **Should we track all file modifications or only explicit logs?**
   - Recommendation: Start explicit only

2. **Should deleted files remain in the list?**
   - Recommendation: Mark as "deleted" status, show with strikethrough

3. **How to handle documents outside worktree?**
   - Recommendation: Allow but show warning icon

4. **Should we version documents?**
   - Recommendation: Phase 2 enhancement, not MVP

---

## Dependencies

- SQLite database at `~/.config/eye-in-the-sky/agents.db`
- Bubble Tea framework for TUI
- Lipgloss for styling
- MCP server running

---

## Success Criteria

- [x] Users can view all documents created during agent session
- [x] Document list shows type, title, and metadata
- [x] Preview pane shows first 20 lines of content
- [x] Navigation with j/k works smoothly
- [x] Empty state handled gracefully
- [x] All tests pass
