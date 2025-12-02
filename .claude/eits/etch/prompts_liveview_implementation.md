# Prompts LiveView Implementation

## Overview
Created a complete LiveView interface for viewing and managing subagent prompts in the Phoenix web application.

## Files Created

### 1. Context Module
**File**: `eye_in_the_sky_web/lib/eye_in_the_sky_web/prompts.ex`
- `list_prompts/1` - List all prompts with optional filters (project_id, include_inactive)
- `get_prompt!/1` - Get prompt by ID (raises on not found)
- `get_prompt_by_slug/2` - Get prompt by slug with project-aware fallback
- `create_prompt/1` - Create new prompt
- `update_prompt/2` - Update existing prompt
- `deactivate_prompt/1` - Soft delete (set active=false)
- `delete_prompt/1` - Hard delete
- `list_global_prompts/0` - List global prompts only
- `list_project_prompts/1` - List project-specific prompts

### 2. Schema Module
**File**: `eye_in_the_sky_web/lib/eye_in_the_sky_web/prompts/prompt.ex`
- Ecto schema for `subagent_prompts` table
- Primary key: UUID string
- Validations:
  - Required: name, slug, prompt_text
  - Slug format: kebab-case (`^[a-z][a-z0-9-]*$`)
  - Unique constraints on slug (global and per-project)
- Auto-generates UUID if not provided

### 3. LiveView Module
**File**: `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/live/prompt_live/index.ex`
- Full-featured prompt list page with:
  - Search functionality (name, slug, description)
  - Scope filtering (all, global, project)
  - Table view with columns: Scope, Name, Slug, Description, Version, Updated
  - Action buttons: View, Edit, Delete (soft delete)
  - DaisyUI styling with dark mode support
  - Real-time filtering with debounce

### 4. Router Update
**File**: `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/router.ex`
- Added route: `live "/prompts", PromptLive.Index, :index`

## Features Implemented

### Search & Filtering
- **Search**: Real-time search across name, slug, and description (300ms debounce)
- **Scope Filter**: Toggle between All, Global, and Project-scoped prompts
- Empty state with helpful message when no results found

### Display
- Badge indicators for scope (Global=primary, Project=secondary)
- Version badges showing current version
- Relative timestamps with full datetime on hover
- Line-clamped descriptions with full text on hover
- Monospace slug display for clarity

### Actions
- **View**: Preview prompt details (button ready, modal not implemented)
- **Edit**: Edit prompt (button ready, form not implemented)
- **Delete**: Soft delete with confirmation dialog

### Responsive Design
- Mobile-friendly layout
- Collapsible search and filters on small screens
- Theme toggle (light/dark mode) in header

## Database Test Data

Three test prompts were added:

1. **Code Review** (global)
   - Slug: `code-review`
   - Description: Review code for bugs, security issues, and best practices

2. **Test Generator** (project-scoped, project_id=1)
   - Slug: `test-gen`
   - Description: Generate comprehensive unit tests for code

3. **Documentation Writer** (global)
   - Slug: `docs-writer`
   - Description: Generate clear and comprehensive documentation

## How to Access

1. Start Phoenix server: `cd eye_in_the_sky_web && mix phx.server`
2. Navigate to: `http://localhost:4000/prompts`

## Next Steps (Not Implemented)

### High Priority
1. **Create/Edit Modal** - Form for creating and editing prompts
2. **View Modal** - Full prompt text preview with syntax highlighting
3. **Project Selector** - Dropdown to filter by specific project
4. **Pagination** - Handle large numbers of prompts efficiently

### Medium Priority
5. **Bulk Actions** - Select multiple prompts for batch operations
6. **Export/Import** - Share prompts between projects or instances
7. **Tags Support** - Display and filter by tags
8. **Version History** - View previous versions of prompts
9. **Duplicate Prompt** - Clone existing prompt as starting point

### Low Priority
10. **Usage Stats** - Track how often prompts are used
11. **Favorites** - Bookmark frequently used prompts
12. **Categories** - Organize prompts into categories

## Technical Notes

- Uses Ecto for database operations (not raw SQL)
- Follows Phoenix LiveView patterns from existing pages
- DaisyUI components for consistent styling
- No global navigation bar (matches existing app design)
- Compilation successful with only deprecation warnings in unrelated code

## Testing Checklist

- [x] Create context module with CRUD operations
- [x] Create Ecto schema with validations
- [x] Create LiveView with search and filters
- [x] Add route to router
- [x] Compile successfully
- [x] Insert test data
- [ ] Manual browser test (requires Phoenix server running)
- [ ] Test search functionality
- [ ] Test scope filtering
- [ ] Test soft delete
- [ ] Test on mobile viewport
- [ ] Test dark mode toggle

## Known Limitations

1. **No Create/Edit UI** - Can only view and delete prompts (create via MCP tools or SQL)
2. **No View Modal** - Can't preview full prompt text in UI
3. **No Project Filtering** - Shows all projects, no dropdown to filter specific project
4. **Soft Delete Only** - Hard delete requires SQL or MCP tools
5. **No Undo** - Deactivated prompts can be reactivated via SQL but not UI
6. **No Navigation Bar** - Direct URL access only (`/prompts`)

## MCP Integration

The LiveView complements the existing MCP tools:
- `i-prompt-list` - Programmatic access to prompts
- `i-prompt-get` - Retrieve specific prompt
- `i-prompt-create` - Create new prompts
- `i-prompt-update` - Update existing prompts
- `i-prompt-delete` - Soft/hard delete prompts

The Phoenix UI provides a visual interface while MCP tools enable Claude Code integration.
