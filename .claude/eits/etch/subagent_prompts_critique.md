# Subagent Prompts Implementation Plan - Design Critique

**Reviewer**: Claude Code Agent b783b254-2944-4170-9919-8a0393c89d59
**Date**: 2025-12-01
**Documents Reviewed**:
- `/Users/urielmaldonado/.claude/plans/glittery-painting-bachman.md`
- `.claude/eits/etch/subagent_prompts_implementation_plan.md`

---

## Executive Summary

The implementation plan is comprehensive and well-structured, but has several critical design flaws and inconsistencies with existing EITS patterns. The plan correctly identifies the Agent.ProjectID type mismatch bug, but introduces new issues in schema design, MCP tool interfaces, and concurrency handling.

**Overall Assessment**: 6/10 - Solid foundation but needs significant refinement before implementation.

---

## Critical Issues (Must Fix Before Implementation)

### 1. Confirmed Type Mismatch Bug ⚠️

**Location**: `internal/database/models.go:20`
**Current**: `ProjectID *int`
**Schema**: `agents.project_id TEXT` (schema.sql:15)

This is a runtime bomb that will cause marshal/unmarshal errors when reading/writing agent project associations.

**Impact**:
- Agent registration with projects will fail
- Querying agents by project will fail
- JSON serialization/deserialization will fail

**Fix**: Change `models.go:20` from `ProjectID *int` to `ProjectID *string`

**Priority**: CRITICAL - Fix this BEFORE implementing subagent_prompts to avoid cascading issues.

---

### 2. Slug Uniqueness Constraint Too Restrictive ⚠️

**Current Design**: `slug TEXT UNIQUE NOT NULL`

**Problem**: Prevents having the same slug for both global and project-scoped prompts.

**Example Failure**:
```sql
-- This works
INSERT INTO subagent_prompts (slug, scope, project_id) VALUES ('bug-fixer', 'global', NULL);

-- This fails with UNIQUE constraint violation
INSERT INTO subagent_prompts (slug, scope, project_id) VALUES ('bug-fixer', 'project', 'abc123');
```

**Why This Matters**:
- Can't override global prompts at project level
- Defeats the purpose of project-scoped prompts
- The priority resolution logic already handles ambiguity

**The Plan's Justification** (line 879): "intentional to prevent ambiguity"
**Counter-argument**: `GetSubagentPromptBySlug()` already resolves ambiguity by checking project scope first (lines 216-225).

**Recommended Fix**:
```sql
-- Option 1: Composite unique constraint
UNIQUE(slug, COALESCE(project_id, ''))

-- Option 2: Partial unique index (SQLite 3.8.0+)
CREATE UNIQUE INDEX idx_unique_slug_global ON subagent_prompts(slug) WHERE scope = 'global';
CREATE UNIQUE INDEX idx_unique_slug_project ON subagent_prompts(slug, project_id) WHERE scope = 'project';
```

**Priority**: HIGH - Fundamentally affects the usability of the feature.

---

### 3. Version Auto-Increment Race Condition ⚠️

**Current Implementation** (lines 318-343):
```go
// Read current prompt (not in transaction)
current, err := db.GetSubagentPromptByID(id)

// Check if prompt_text changed
if *params.PromptText != current.PromptText {
    updates = append(updates, "version = version + 1")
}

// Execute UPDATE (separate transaction)
_, err = db.conn.Exec(query, args...)
```

**Race Condition**:
1. Thread A reads current version: 5
2. Thread B reads current version: 5
3. Thread A updates version to 6
4. Thread B updates version to 6 (lost update!)

**Impact**: Version numbers can be incorrect under concurrent modifications.

**Recommended Fix**:
```sql
-- Use database trigger instead (atomic)
CREATE TRIGGER version_on_text_change
AFTER UPDATE OF prompt_text ON subagent_prompts
FOR EACH ROW
WHEN NEW.prompt_text != OLD.prompt_text
BEGIN
    UPDATE subagent_prompts SET version = OLD.version + 1 WHERE id = NEW.id;
END;
```

**Additional Issue**: `UpdateSubagentPromptParams.Version *int` (line 131) allows manual version override but the implementation doesn't use it. Remove this field if versioning should be automatic.

**Priority**: MEDIUM - Low likelihood in single-agent scenarios, but violates correctness guarantees.

---

## Major Design Issues

### 4. Tags Field Violates Existing Pattern

**Current Design**: `tags TEXT` (comma-separated string)

**Existing EITS Pattern** (schema.sql:212-224):
```sql
CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT UNIQUE, color TEXT);
CREATE TABLE task_tags (task_id TEXT, tag_id INTEGER, PRIMARY KEY (task_id, tag_id));
```

**Problems**:
- No referential integrity
- Can't reuse tags across tasks and prompts
- Inefficient filtering (requires LIKE queries)
- Inconsistent with established codebase patterns

**Recommended Fix**:
```sql
CREATE TABLE prompt_tags (
    prompt_id TEXT,
    tag_id INTEGER,
    PRIMARY KEY (prompt_id, tag_id),
    FOREIGN KEY (prompt_id) REFERENCES subagent_prompts(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

**Benefits**:
- Reuse existing tags infrastructure
- Efficient filtering by tag
- Referential integrity
- Consistent with codebase patterns

**Priority**: MEDIUM - Affects query performance and data integrity.

---

### 5. CASCADE Deletion Inconsistency with EITS Pattern

**Current Design**: `project_id TEXT REFERENCES projects(id) ON DELETE CASCADE`

**EITS Pattern Analysis**:
```sql
agents.project_id     → NO CASCADE (defaults to RESTRICT)
tasks.project_id      → NO CASCADE (defaults to RESTRICT)
commits.project_id    → NO CASCADE (defaults to RESTRICT)
session_metrics.*     → ON DELETE CASCADE (dependent data)
task_sessions.*       → ON DELETE CASCADE (junction table)
```

**Pattern**: EITS uses CASCADE for junction tables and dependent data, but RESTRICT for entity references.

**Question**: Are prompts "dependent data" or "entity references"?

**Argument for CASCADE** (current design):
- Project deleted → prompts orphaned → delete them
- Matches migration 003 project cleanup pattern

**Argument for RESTRICT** (EITS pattern):
- Project deleted → prevent if prompts exist
- Preserves valuable prompts
- Matches agent/task/commit pattern

**Edge Case**: What if a project-scoped prompt is actively being used by a running session when the project gets deleted? CASCADE would delete it mid-session, potentially causing errors.

**Recommendation**:
1. Follow EITS pattern and use RESTRICT (safer, more explicit)
2. OR add soft-delete to projects table and never hard-delete
3. OR document CASCADE behavior prominently in CLAUDE.md

**Priority**: MEDIUM - Affects data safety and system behavior.

---

### 6. created_by Field Lacks Foreign Key

**Current Schema** (line 27): `created_by TEXT`

**Issues**:
- No referential integrity
- No cascade behavior if agent deleted
- Unclear what it represents (agent ID, user, session?)

**EITS Pattern**:
```sql
tasks.agent_id    → FOREIGN KEY to agents(id)
actions.agent_id  → FOREIGN KEY to agents(id)
commits.agent_id  → FOREIGN KEY to agents(id)
```

**The Plan Says** (line 65): "could FK to agents table if needed"
**Problem**: This is wishy-washy. Either it's an agent reference or it's not.

**Recommendations**:
1. **If agent ID**: Add `FOREIGN KEY (created_by) REFERENCES agents(id)`
2. **If human user**: Rename to `created_by_user` and document format
3. **If unused**: Remove the field entirely
4. **If both possible**: Use two fields: `created_by_agent_id`, `created_by_user`

**Priority**: LOW - Doesn't break functionality, but affects data integrity.

---

## MCP Tool Design Issues

### 7. i-prompt-get Parameter Ambiguity

**Current Schema** (lines 530-548):
```json
{
  "slug": "string (provide slug or id)",
  "id": "string (provide slug or id)",
  "project_id": "string (optional)"
}
```

**Problems**:
1. What if both slug AND id provided? Plan doesn't specify priority.
2. What if NEITHER provided? Schema doesn't enforce "at least one".
3. Handler pseudocode lacks validation for these cases.

**Recommended Fix**:
```json
{
  "oneOf": [
    {"required": ["slug"]},
    {"required": ["id"]}
  ]
}
```

**Handler Validation**:
```go
if (slug != "" && id != "") || (slug == "" && id == "") {
    return error("provide exactly one of: slug or id")
}
```

**Priority**: MEDIUM - Will cause user confusion and potential bugs.

---

### 8. i-prompt-list Response Redundancy

**Current Response** (lines 602-618):
```json
{
  "success": true,
  "prompts": [...],
  "count": 1,
  "total": 10
}
```

**Questions**:
- What's the difference between `count` and `total`?
- Is `count` the length of returned array? (redundant - caller can do `prompts.length`)
- Is `total` the total matching prompts before LIMIT? (requires separate COUNT(*) query)

**Missing Implementation**:
- No COUNT(*) query shown in `ListSubagentPrompts()` (lines 256-310)
- No documentation of count vs total semantics

**Recommended Response**:
```json
{
  "prompts": [...],
  "total": 100,        // Total matching prompts
  "limit": 50,
  "offset": 0,
  "has_more": true
}
```

**Priority**: LOW - Cosmetic issue, but affects API clarity.

---

### 9. Missing Filter by Tags

**Current i-prompt-list filters**:
- project_id ✓
- scope ✓
- active ✓
- limit/offset ✓
- **tags ✗**

**Use Case**: "Show me all prompts tagged with 'code-review'"

**Current Workaround**:
1. Fetch ALL prompts
2. Parse comma-separated tags client-side
3. Filter in memory (inefficient)

**If Using Junction Table**:
```sql
SELECT p.* FROM subagent_prompts p
JOIN prompt_tags pt ON p.id = pt.prompt_id
JOIN tags t ON pt.tag_id = t.tag_id
WHERE t.name IN ('code-review', 'security') AND p.active = 1
```

**Recommendation**: Add `tags` array parameter to `i-prompt-list`.

**Priority**: MEDIUM - Affects usability and query performance.

---

## Go Implementation Issues

### 10. Missing context.Context Parameter

**Current Signatures**:
```go
func (db *DB) CreateSubagentPrompt(params CreateSubagentPromptParams) (*SubagentPrompt, error)
func (db *DB) GetSubagentPromptByID(id string) (*SubagentPrompt, error)
```

**Go Best Practice**: All database operations should accept `context.Context` for:
- Request cancellation
- Timeout enforcement
- Trace/span propagation
- Graceful shutdown

**The Plan Mentions This** (line 421): "add ctx parameter"
**But**: Function signatures aren't updated

**Corrected Signatures**:
```go
func (db *DB) CreateSubagentPrompt(ctx context.Context, params CreateSubagentPromptParams) (*SubagentPrompt, error)
func (db *DB) GetSubagentPromptByID(ctx context.Context, id string) (*SubagentPrompt, error)
```

**Priority**: MEDIUM - Affects production-readiness.

---

### 11. Dynamic UPDATE Query Lacks Validation

**Current Code** (lines 324-359):
```go
updates := []string{}
args := []interface{}{}

if params.Name != nil {
    updates = append(updates, "name = ?")
    args = append(args, *params.Name)
}
```

**Problem**: No validation that `*params.Name` isn't empty string.

**Impact**: Will violate implicit NOT NULL constraint expectations.

**Example**:
```go
UpdateSubagentPrompt(id, UpdateSubagentPromptParams{
    Name: ptr(""),  // Empty string - should error but doesn't
})
```

**Recommended Fix**:
```go
if params.Name != nil {
    if *params.Name == "" {
        return nil, fmt.Errorf("name cannot be empty")
    }
    updates = append(updates, "name = ?")
    args = append(args, *params.Name)
}
```

**Priority**: LOW - Edge case but violates data integrity.

---

## Testing & Documentation Issues

### 12. Incomplete Test Coverage

**Current Test Plan** (lines 813-834) covers:
- CRUD operations ✓
- Constraint validation ✓
- Slug uniqueness ✓
- Project scope priority ✓
- Soft vs hard delete ✓

**Missing Critical Tests**:
1. Concurrent modifications / version race conditions
2. Trigger behavior (updated_at auto-update)
3. CASCADE deletion (delete project → prompts deleted)
4. NULL handling for all nullable fields
5. CHECK constraint violations
6. Slug case sensitivity (is "Bug-Fixer" ≠ "bug-fixer"?)
7. Large prompt_text (MB+ sized content)
8. Unicode in slugs/names
9. SQL injection / XSS in prompt_text
10. MCP tool error responses (all return success:false appropriately)

**Missing Performance Tests**:
- Query performance with 1K, 10K, 100K prompts
- Index effectiveness (EXPLAIN QUERY PLAN)
- Concurrent access under load

**Priority**: MEDIUM - Incomplete testing risks production bugs.

---

### 13. Missing Critical Usage Documentation

**Current Examples** (lines 897-960) show:
- Creating prompts ✓
- Retrieving prompts ✓
- Listing prompts ✓

**Missing Critical Patterns**:
1. **How does an agent actually USE a prompt?**
   - Manual fetch with i-prompt-get?
   - Auto-injection by EITS?
   - Integration with Task tool?

2. **Prompt lifecycle management**:
   - Version vs create new?
   - Deprecation strategy?
   - Migration across projects?

3. **Global vs Project decision tree**:
   - When to use which scope?
   - Promote project → global?
   - Override global for project?

4. **Performance considerations**:
   - Should prompts be cached?
   - Limits (how many prompts is too many)?
   - Optimization for large prompt_text?

5. **Security considerations**:
   - Can prompts contain secrets?
   - Access control (who modifies globals)?
   - Audit logging?

**Priority**: HIGH - Poor documentation = poor adoption.

---

### 14. No Migration Rollback Strategy

**Current Plan**: Create migration `005_add_subagent_prompts_table.sql`

**Missing**:
1. Rollback script (what if migration fails?)
2. Data migration (are there existing prompts to migrate?)
3. Version compatibility (old code + new schema?)
4. Migration testing procedure
5. Schema version tracking in `meta` table

**SQLite Limitation**: Not all DDL is transactional, so partial failures can leave schema corrupted.

**Recommendation**:
1. Create `005_rollback_subagent_prompts.sql`
2. Test migration on database copy
3. Document rollback procedure
4. Update `meta.schema_version`

**Priority**: MEDIUM - Affects deployment safety.

---

## Future Enhancements Critique

### 15. Prompt Variables Design Flaw

**Proposed** (lines 843-846):
> Support template variables in prompt_text (e.g., `{{project_name}}`)
> Implement variable substitution at retrieval time

**Problems**:
1. Mixes data storage with presentation logic
2. Where do variables come from? (project context? parameters? system?)
3. Does version track template or rendered output?
4. How to escape literal `{{var}}`?

**Better Design**:
```sql
-- Add is_template flag
ALTER TABLE subagent_prompts ADD COLUMN is_template BOOLEAN DEFAULT 0;
ALTER TABLE subagent_prompts ADD COLUMN variable_schema TEXT; -- JSON schema
```

**Separate MCP Tool**:
```json
// i-prompt-render
{
  "prompt_id": "uuid",
  "variables": {
    "project_name": "eye-in-the-sky",
    "project_path": "/path/to/repo"
  }
}
→ Returns rendered prompt
```

**Priority**: LOW - Future feature, but worth designing correctly now.

---

## Positive Aspects

The plan does several things well:

1. **Comprehensive Documentation**: 976 lines covering schema, Go code, MCP tools, and examples
2. **Pattern Awareness**: Correctly identifies and attempts to follow EITS patterns
3. **Explicit Scoping**: The scope field + CHECK constraints are good design
4. **Soft Delete Pattern**: Follows EITS convention with active flag
5. **Indexing Strategy**: Proper indexes on FKs and frequently queried fields
6. **Migration Structure**: Clear phases and implementation checklist
7. **Bug Discovery**: Identified the Agent.ProjectID type mismatch

---

## Summary of Recommendations

### Critical (Must Fix)
1. ✅ Fix Agent.ProjectID type mismatch (int → string)
2. ✅ Fix slug uniqueness constraint to allow project overrides
3. ✅ Fix version auto-increment race condition with database trigger

### High Priority
4. ✅ Follow EITS pattern for tags (junction table, not TEXT)
5. ✅ Add comprehensive usage documentation

### Medium Priority
6. ✅ Reconsider CASCADE vs RESTRICT for project_id FK
7. ✅ Add context.Context to all database functions
8. ✅ Fix i-prompt-get parameter validation (oneOf constraint)
9. ✅ Add filter by tags to i-prompt-list
10. ✅ Expand test coverage (concurrency, triggers, cascades)
11. ✅ Add migration rollback strategy

### Low Priority
12. ✅ Clarify created_by field purpose or remove
13. ✅ Fix i-prompt-list response redundancy (count vs total)
14. ✅ Add empty string validation in update
15. ✅ Design template variables properly if implementing

---

## Overall Verdict

**Rating**: 6/10 - Solid foundation, significant refinement needed

**Strengths**:
- Comprehensive planning and documentation
- Good understanding of EITS patterns
- Identifies existing bugs
- Clear implementation phases

**Weaknesses**:
- Critical race condition in version management
- Slug uniqueness too restrictive
- Inconsistent with EITS patterns (tags, cascade)
- Missing edge case testing
- Incomplete usage documentation

**Recommendation**:
- Fix critical issues (1-3) before implementation
- Address high priority issues (4-5) during implementation
- Defer low priority items to post-MVP iteration

**Confidence**: HIGH - The plan is thorough enough to identify most issues through code review. Actual implementation will surface additional edge cases.

---

## Next Steps

1. **Immediate**: Fix Agent.ProjectID type mismatch
2. **Before Implementation**: Address critical and high priority issues
3. **During Implementation**: Write comprehensive tests
4. **Post-Implementation**: Document usage patterns and best practices
5. **Follow-up**: Consider template variables and advanced features

---

**End of Critique**
