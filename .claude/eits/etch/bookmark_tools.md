# Bookmark MCP Tools

## Proposed Tools

### `i-bookmark-add`
Add a new bookmark.

**Parameters:**
- `bookmark_type` (required): "file" | "note" | "agent" | "session" | "task" | "url"
- `bookmark_id` (optional): ID for note/agent/session/task bookmarks
- `file_path` (optional): Path for file bookmarks
- `line_number` (optional): Line number for file bookmarks
- `url` (optional): URL for url bookmarks
- `title` (optional): Display title
- `description` (optional): Additional context
- `category` (optional): Grouping category
- `priority` (optional): Importance (0-100)
- `project_id` (optional): Project context
- `agent_id` (optional): Creating agent

**Examples:**
```json
// Bookmark a file
{
  "bookmark_type": "file",
  "file_path": "/Users/urielmaldonado/projects/eye-in-the-sky/internal/mcp/tools.go",
  "line_number": 431,
  "title": "i-instructions text",
  "category": "important",
  "priority": 80
}

// Bookmark a note
{
  "bookmark_type": "note",
  "bookmark_id": "123",
  "title": "Critical bug findings",
  "category": "bugs"
}
```

### `i-bookmark-list`
List bookmarks with optional filtering.

**Parameters:**
- `bookmark_type` (optional): Filter by type
- `category` (optional): Filter by category
- `project_id` (optional): Filter by project
- `agent_id` (optional): Filter by agent
- `limit` (optional): Max results (default: 50)
- `order_by` (optional): "priority" | "created_at" | "accessed_at" | "position"

**Returns:** Array of bookmark objects with expanded references

### `i-bookmark-update`
Update an existing bookmark.

**Parameters:**
- `bookmark_id` (required): Bookmark to update
- `title` (optional): New title
- `description` (optional): New description
- `category` (optional): New category
- `priority` (optional): New priority
- `position` (optional): New position

### `i-bookmark-delete`
Remove a bookmark.

**Parameters:**
- `bookmark_id` (required): Bookmark to delete

### `i-bookmark-access`
Record bookmark access (updates accessed_at timestamp).

**Parameters:**
- `bookmark_id` (required): Bookmark being accessed

**Returns:** Full bookmark details with expanded reference (e.g., file contents, note body, agent info)

## UI Integration

### Where Bookmarks Appear

1. **Global Bookmarks Page** (`/bookmarks`)
   - Grouped by category
   - Sortable by priority/recency
   - Quick jump to bookmarked entities

2. **Project Bookmarks Tab** (`/projects/:id/bookmarks`)
   - Project-specific bookmarks
   - File bookmarks with line numbers
   - One-click navigation

3. **Agent Detail View** (`/agents/:id`)
   - Bookmarks created by that agent
   - Bookmark the agent itself

4. **File Tree View** (`/projects/:id/files`)
   - Bookmark icon next to files
   - Show bookmarked line numbers

5. **TUI Dashboard**
   - New 'B' key to view bookmarks
   - Quick bookmark current agent/file

## Implementation Steps

1. Add `bookmarks` table to database migration
2. Implement MCP tools in `internal/mcp/tools.go`
3. Add bookmark repository in `internal/database/`
4. Add Phoenix LiveView pages:
   - `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/live/bookmark_live/index.ex`
   - `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/live/bookmark_live/show.ex`
5. Add bookmark component to project nav
6. Add TUI bookmark view in `internal/ui/`
7. Update documentation in CLAUDE.md
