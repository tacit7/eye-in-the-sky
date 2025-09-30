# CLAUDE_CONTEXT.md

This file provides context to Claude Code about the current state of the Eye in the Sky MCP Server project.

## Current Project State

### Last Updated: 2025-09-29 06:45:00 UTC

### Recent Changes
- ✅ **MCP Tool Naming Convention Updated**: All 12 MCP tools now use "i-" prefix as requested
- ✅ **Handler Functions Complete**: Added missing handler functions for all tools
- ✅ **Backward Compatibility**: Dashboard server can still use old tool names
- ✅ **Testing Verified**: All tools tested and working with new naming convention

### MCP Tools Current State
All tools now use "i-" prefix for Claude Desktop integration:

#### Core Agent Management
- `i-register-agent` - Register Claude Code worktree agents
- `i-register-claude-desktop-agent` - Register Claude Desktop agents
- `i-update-status` - Update agent status and current task
- `i-end-session` - Complete agent sessions

#### Activity Tracking
- `i-log-action` - Log agent activities with timestamps
- `i-log-commits` - Track git commits by agents

#### Session Management
- `i-save-session-context` - Save session state for resumption
- `i-load-session-context` - Load previous session state
- `i-add-session-note` - Add contextual notes to sessions

#### Window Management (macOS)
- `i-get-current-window` - Get active window information
- `i-bring-window-front` - Bring agent window to front

#### Help & Support
- `i-help` - Get detailed help and usage instructions

### Technical Implementation Details

#### File Changes Made
1. **`internal/mcp/server.go`**:
   - Updated `registerTools()` function with "i-" prefix for all tools
   - Added missing handler functions for session context and window management
   - Enhanced `HandleTool()` for dual compatibility (old/new names)

2. **Tool Registration Pattern**:
   ```go
   mcp.AddTool(s.mcp, &mcp.Tool{
       Name:        "i-[tool-name]",
       Description: "[tool description]",
   }, s.handle[ToolName])
   ```

#### Compatibility Layers
- **MCP Protocol**: Uses new "i-" prefixed names
- **Dashboard Server**: Supports both old and new names via HandleTool function
- **Claude Desktop**: Will see and use new "i-" prefixed tool names

### Server Configuration
- **Build Target**: `./bin/eye-in-the-sky`
- **Database**: `./data/agents.db` (SQLite)
- **Dashboard**: `http://localhost:8080`
- **MCP Transport**: stdio (JSON-RPC over stdin/stdout)

### Dependencies Status
- **Go Version**: 1.23.0
- **MCP SDK**: v0.8.0 (`github.com/modelcontextprotocol/go-sdk/mcp`)
- **Database**: `github.com/mattn/go-sqlite3`
- **All Dependencies**: ✅ Resolved and working

### Testing Status
- ✅ **Build**: Successful compilation
- ✅ **MCP Protocol**: All 12 tools registered and responding
- ✅ **Tool Names**: Verified all tools have "i-" prefix
- ✅ **Server Startup**: Clean startup with dashboard and MCP server
- ✅ **Backward Compatibility**: Dashboard can call tools with old names

### Current Issues
- **None**: Implementation completed successfully
- **Claude Desktop Connection**: Previous connection issues may be resolved with new naming

### Next Development Areas
1. **Session Context Implementation**: Placeholder handlers can be enhanced with full functionality
2. **Window Management**: macOS-specific features can be expanded
3. **Performance Monitoring**: Add metrics for tool usage
4. **Error Handling**: Enhanced error reporting and recovery

### Development Workflow
1. **Build**: `go build -o bin/eye-in-the-sky ./cmd/server`
2. **Test MCP**: `python3 test_i_prefix.py` (validates all tool names)
3. **Run Server**: `./bin/eye-in-the-sky` (starts both MCP and dashboard)
4. **Dashboard**: Access at `http://localhost:8080`

### Key Architecture Points
- **Multi-Mode Operation**: Detects if running as MCP server or dashboard server
- **Intelligent Mode Detection**: Uses stdin characteristics to determine runtime mode
- **Tool Registration**: Centralized in `registerTools()` function
- **Handler Pattern**: Consistent handler functions for all MCP tools
- **Database Integration**: SQLite for persistent agent and session data

---
**Agent ID**: 6d09ae9e (Current session)
**Project Path**: `/Users/urielmaldonado/projects/eye-in-the-sky`
**Documentation**: See STATUS.md for latest session summary
**FAQ**: See `/Users/urielmaldonado/projects/FAQS/projects/eye-in-the-sky-faq.md`