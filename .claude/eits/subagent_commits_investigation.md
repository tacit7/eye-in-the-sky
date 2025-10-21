# Subagent Commits Visibility Investigation - Ticket #27

## Executive Summary

Subagent commits are not visible in the Eye in the Sky TUI dashboard because the commit query logic only retrieves commits for the currently viewed agent, not for any child agents in the parent-child hierarchy.

## Root Cause

The visibility issue stems from TWO locations where commits are queried:

### 1. Database Layer Issue
**File**: `internal/database/queries.go`

Two functions have identical single-agent-only queries:

- **GetCommitsForAgent()** (lines 190-221)
- **ListCommits()** (lines 348-378)

Both use:
```sql
SELECT id, agent_id, commit_hash, commit_message, timestamp
FROM commits
WHERE agent_id = ?
ORDER BY timestamp DESC
```

This query ONLY returns commits for the specified agent ID, completely ignoring the `parent_agent_id` relationship.

### 2. UI Layer Issue
**File**: `internal/ui/app/model.go`

The commit loading code (lines 572-578) directly queries the database with the same limitation:
```sql
commitsQuery := `
    SELECT id, agent_id, commit_hash, commit_message, timestamp
    FROM commits
    WHERE agent_id = ?
    ORDER BY timestamp DESC
    LIMIT 20
`

commitRows, err := m.db.Query(commitsQuery, agent.ID)
```

This bypasses the database layer function and uses a raw query that also only fetches commits for the current agent.

## Database Structure Verification

The database schema IS correctly configured to support parent-child relationships:

- **agents table** has `parent_agent_id TEXT` column (added in migration 0027_add_parent_session_id.sql)
- **agents table** has index `idx_agents_parent_session_id` for efficient lookups
- **agents table** stores `parent_session_id TEXT` for session isolation

Example data in database:
```
Parent agent: cd9d55fd-2f0b-450a-b526-afb8531f0e81
  └─ Child agent: 489fb01c-6860-40f1-8f37-b29cfcad2590 (no commits logged)

Parent agent: 70391860-E5B4-48C2-A9E9-6AF77BC7D04E
  └─ 7 child agents (none have commits)

Parent agent: 68F3F2EA
  └─ Child agent: a1b2c3d4 (has 18 commits)
    └─ Grandchild agent: e5f6g7h8 (has 0 commits)
```

## Impact Analysis

When a parent agent with child agents is viewed in the TUI dashboard:

1. User navigates to parent agent
2. Presses 'c' or 'l' to view Commits tab
3. Only parent's commits are displayed
4. **Child agent commits are completely invisible** even though:
   - They exist in the database
   - The parent-child relationship is correctly stored
   - The database has proper indexes for queries

This breaks the core feature of the Eye in the Sky system: providing visibility into multi-agent hierarchies.

## Missing Implementation

The system lacks the following functionality:

### Missing Database Functions

1. **GetChildAgentIDs(agentID string) -> []string**
   - Recursively retrieves all descendant agent IDs
   - Must handle multiple generations (grandchildren, etc.)
   - Should be efficient with caching or CTE for deep hierarchies

2. **GetCommitsForAgentHierarchy(agentID string) -> []*Commit**
   - Retrieves commits for the agent AND all descendants
   - Should maintain chronological ordering across all commits
   - Optional: filter by parent_session_id for session isolation
   - Optional: limit results (LIMIT clause)

### Missing UI Changes

The UI needs to:
1. Use the new database function instead of direct queries
2. Display agent source in commit list (e.g., "(child)" tag)
3. Consider visual hierarchy or grouping of child commits
4. Handle the case where parent has no commits but children do

## Considerations for Implementation

### 1. Session Isolation
Should child agent commits only be displayed if they belong to the same parent_session_id?
- Current schema stores parent_session_id on agents
- This could prevent displaying commits from unrelated subagent sessions
- Needs clarification on desired behavior

### 2. Performance
For agents with many descendants:
- Recursive queries could be slow
- Might need indexes on (parent_agent_id, timestamp)
- Could implement caching layer

### 3. UI Presentation
- Should child commits be visually distinguished?
- Should there be a hierarchy indicator showing which agent made the commit?
- Should commits be grouped by agent or chronological?

## Files Requiring Modification

1. **internal/database/queries.go**
   - Add GetChildAgentIDs() function
   - Add GetCommitsForAgentHierarchy() function
   - Consider updating GetCommitsForAgent() and ListCommits() to handle hierarchies

2. **internal/ui/app/model.go** (lines 572-578)
   - Replace direct query with call to GetCommitsForAgentHierarchy()
   - Consider updating commit rendering to show agent source

3. **internal/database/models.go** (optional)
   - May need to add AgentID field to Commit model for display purposes if not already present

## Database Queries Verified

- Schema supports parent-child relationships: ✓
- parent_agent_id column exists: ✓
- parent_session_id column exists: ✓
- Indexes on both columns exist: ✓
- Sample hierarchies exist in database: ✓
- Commits table has agent_id foreign key: ✓

## Next Steps

1. Implement GetChildAgentIDs() in database layer
2. Implement GetCommitsForAgentHierarchy() in database layer
3. Update UI query in model.go:572-578 to use new function
4. Add child agent indicators in commit display
5. Test with parent agents that have child agents with commits
6. Consider writing database tests for the new hierarchy functions

## Related Artifacts

- Ticket #27: "Investigate why subagent commits are not visible in Eye in the Sky"
- Tag investigation in Taskwarrior
- Parent agent from previous investigation: cd9d55fd-2f0b-450a-b526-afb8531f0e81
- Child agent example: 489fb01c-6860-40f1-8f37-b29cfcad2590
