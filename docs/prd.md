# Claude Code Multi-Agent Management System - Product Requirements Document

**Document Version:** 1.0  
**Date:** September 27, 2025  
**Author:** Uriel Maldonado  
**Project Code:** CCMAS-001  

---

## 1. Executive Summary

### 1.1 Project Overview

The Claude Code Multi-Agent Management System is a developer tool that provides real-time visibility and control over multiple concurrent Claude Code instances. The system solves the problem of losing track of AI agents working across different git worktrees and features by providing a centralized dashboard showing what each agent is currently doing.

### 1.2 Business Objectives

- **Primary Goal:** Eliminate the problem of forgetting about running Claude
  Code instances and losing track of their progress
- **Secondary Goal:** Provide centralized visibility into multiple concurrent
  AI development agents
- **Success Metrics:** 100% visibility into active Claude Code agents with clear status reporting

### 1.3 Target Users

- **Primary:** Senior developers running multiple Claude Code instances across different git worktrees
- **Secondary:** Development teams coordinating multiple AI-assisted coding sessions

---

## 2. Product Requirements

### 2.1 Functional Requirements

#### 2.1.1 Agent Management

- **REQ-001:** System shall generate unique 8-character hash IDs for each Claude Code agent instance (format: git-style hash like "a3f7d2e1")
- **REQ-002:** System shall track agent status: "active", "idle", "working", "completed", "failed"
- **REQ-003:** System shall record agent metadata: created_at, updated_at, git_worktree_path, feature_description, current_task
- **REQ-004:** System shall provide agent lifecycle management (register, update_status, log_action, end_session)

#### 2.1.2 Action Logging & Progress Tracking

- **REQ-005:** System shall log agent actions when Claude reports major work items (e.g., "Starting work on login controller")
- **REQ-006:** System shall track git commits made during agent sessions
- **REQ-007:** System shall categorize actions by type: "task_start", "file_operation", "git_commit", "status_update"
- **REQ-008:** System shall store current agent status and last known activity

#### 2.1.3 Multi-Agent Dashboard

- **REQ-009:** System shall provide web interface accessible at localhost:8080
- **REQ-010:** Dashboard shall display real-time status of all active agents across different worktrees
- **REQ-011:** Dashboard shall show current task, last activity, and time since last update for each agent
- **REQ-012:** Dashboard shall provide detailed agent history showing all actions and commits
- **REQ-013:** Dashboard shall highlight agents that haven't reported activity recently (potential forgotten instances)

#### 2.1.4 MCP Integration

- **REQ-014:** System shall expose MCP tools for Claude Code integration:
  - `register_agent(agent_id, description, worktree_path)` → confirms registration
  - `update_status(agent_id, status, current_task?)` → updates agent status
  - `log_action(agent_id, action_type, description, details?)` → logs agent activity
  - `log_commits(agent_id, commit_hashes[])` → tracks git commits made
  - `end_session(agent_id, summary?, final_status?)` → completes agent session

### 2.2 Non-Functional Requirements

#### 2.2.1 Performance

- **REQ-015:** Dashboard page loads shall complete within 2 seconds for datasets up to 100 active agents
- **REQ-016:** MCP tool calls shall respond within 500ms
- **REQ-017:** System shall handle concurrent agent updates without data corruption

#### 2.2.2 Reliability

- **REQ-018:** System shall maintain 99% uptime during development sessions
- **REQ-019:** Database operations shall be atomic to prevent data corruption
- **REQ-020:** System shall gracefully handle agent disconnections

#### 2.2.3 Usability

- **REQ-021:** Dashboard shall be accessible without authentication (localhost-only)
- **REQ-022:** Interface shall clearly distinguish between active and idle agents
- **REQ-023:** Dashboard shall provide visual alerts for agents with no recent activity

#### 2.2.4 Maintainability

- **REQ-024:** Codebase shall follow Python PEP 8 standards
- **REQ-025:** Database schema shall support future feature additions
- **REQ-026:** System shall provide clear error messages and logging

---

## 3. Technical Specifications

### 3.1 Architecture

#### 3.1.1 System Components

- **MCP Server:** Python-based server implementing Model Context Protocol
- **Web Dashboard:** Flask-based web application
- **Database:** SQLite for data persistence
- **Integration Layer:** MCP tools interface for Claude Code

#### 3.1.2 Technology Stack

- **Backend:** Go 1.21+ (single binary deployment)
- **Web Framework:** Standard library `net/http` + `html/template`
- **Database:** SQLite 3 with `github.com/mattn/go-sqlite3`
- **Frontend:** HTML5, Bootstrap 5, vanilla JavaScript
- **MCP Implementation:** Official Go SDK `github.com/modelcontextprotocol/go-sdk/mcp`

### 3.2 Database Design

#### 3.2.1 Agents Table

```sql
CREATE TABLE agents (
    id TEXT PRIMARY KEY,              -- 8-char hash like "a3f7d2e1"
    status TEXT NOT NULL,             -- "active", "idle", "working", "completed", "failed"
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    git_worktree_path TEXT,           -- path to git worktree
    feature_description TEXT,         -- what the agent is working on
    current_task TEXT,                -- current specific task
    last_activity_at TIMESTAMP        -- when agent last reported activity
);
```

#### 3.2.2 Actions Table

```sql
CREATE TABLE actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    action_type TEXT NOT NULL,        -- "task_start", "file_operation", "git_commit", "status_update"
    description TEXT NOT NULL,        -- human-readable description of action
    details TEXT,                     -- JSON blob for additional context
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);
```

#### 3.2.3 Commits Table

```sql
CREATE TABLE commits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    commit_hash TEXT NOT NULL,
    commit_message TEXT,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);
```

### 3.3 API Specifications

#### 3.3.1 MCP Tools Interface

```go
// Tool: register_agent
type RegisterAgentArgs struct {
    AgentID     string `json:"agent_id" jsonschema:"description=8-character agent identifier"`
    Description string `json:"description" jsonschema:"description=What the agent is working on"`
    WorktreePath string `json:"worktree_path,omitempty" jsonschema:"description=Path to git worktree"`
}

type RegisterAgentResult struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

// Tool: update_status
type UpdateStatusArgs struct {
    AgentID     string `json:"agent_id" jsonschema:"description=Agent identifier"`
    Status      string `json:"status" jsonschema:"description=Agent status (active/idle/working/completed/failed)"`
    CurrentTask string `json:"current_task,omitempty" jsonschema:"description=Current task description"`
}

// Tool: log_action
type LogActionArgs struct {
    AgentID     string `json:"agent_id" jsonschema:"description=Agent identifier"`
    ActionType  string `json:"action_type" jsonschema:"description=Type of action (task_start/file_operation/git_commit/status_update)"`
    Description string `json:"description" jsonschema:"description=Human-readable action description"`
    Details     string `json:"details,omitempty" jsonschema:"description=Additional JSON details"`
}

// Tool: log_commits
type LogCommitsArgs struct {
    AgentID        string   `json:"agent_id" jsonschema:"description=Agent identifier"`
    CommitHashes   []string `json:"commit_hashes" jsonschema:"description=Array of git commit hashes"`
    CommitMessages []string `json:"commit_messages,omitempty" jsonschema:"description=Array of commit messages"`
}

// Tool: end_session
type EndSessionArgs struct {
    AgentID     string `json:"agent_id" jsonschema:"description=Agent identifier"`
    Summary     string `json:"summary,omitempty" jsonschema:"description=Session summary"`
    FinalStatus string `json:"final_status,omitempty" jsonschema:"description=Final agent status"`
}
```

### 3.4 File Structure

```
claude-code-mcp/
├── README.md
├── go.mod
├── go.sum
├── main.go                        # Entry point
├── cmd/
│   └── server/
│       └── main.go                # Main server executable
├── internal/
│   ├── mcp/
│   │   ├── server.go              # MCP server implementation
│   │   └── tools.go               # MCP tool handlers
│   ├── dashboard/
│   │   ├── server.go              # HTTP server for dashboard
│   │   └── handlers.go            # HTTP handlers
│   ├── database/
│   │   ├── db.go                  # Database connection and setup
│   │   ├── models.go              # Data models
│   │   └── queries.go             # Database queries
│   └── utils/
│       └── utils.go               # Utility functions
├── web/
│   ├── templates/
│   │   ├── base.html              # Base template
│   │   ├── index.html             # Agent overview page
│   │   └── agent.html             # Agent detail page
│   └── static/
│       ├── css/
│       │   └── styles.css
│       └── js/
│           └── dashboard.js
├── data/
│   └── agents.db                  # SQLite database file (created at runtime)
├── scripts/
│   └── build.sh                   # Build script
└── tests/
    ├── mcp_test.go
    ├── database_test.go
    └── dashboard_test.go
```

---

## 4. User Experience

### 4.1 User Flows

#### 4.1.1 Primary User Flow: Multi-Agent Management

1. Developer starts Claude Code instance in worktree #1: "You are agent abc123, register yourself for working on user authentication"
2. Claude calls `register_agent()` with ID and description
3. Developer starts second Claude Code instance in worktree #2: "You are agent def456, register yourself for API endpoint work"
4. Developer opens dashboard at localhost:8080 to see both active agents
5. Claude instances periodically call `update_status()` and `log_action()` as they work
6. Developer checks dashboard throughout day to see current status of all agents
7. When agents complete work, they call `log_commits()` and `end_session()`

#### 4.1.2 Secondary User Flow: Agent Status Monitoring

1. Developer has 3 active Claude instances running
2. Developer opens dashboard to check progress
3. Dashboard shows one agent hasn't reported activity in 2 hours (potential forgotten instance)
4. Developer investigates and finds Claude instance waiting for input
5. Developer provides input and agent resumes work

#### 4.1.3 Workflow: Session Completion

1. Claude completes feature work
2. Developer runs: "end-session, log these commits: abc123f, def456a"
3. Claude calls `log_commits()` then `end_session()`
4. Dashboard updates to show agent as completed with commit summary

### 4.2 Interface Design

#### 4.2.1 Agent Overview Page (/)

- **Header:** "Active Claude Code Agents"
- **Active Agents Section:**
  - **Table Columns:** Agent ID, Status, Worktree, Current Task, Last Activity, Actions
  - **Status Indicators:** Green (active), Yellow (idle), Red (no activity >2hrs)
  - **Quick Actions:** View Details, End Session
- **Recently Completed Section:**
  - **Table Columns:** Agent ID, Completed, Feature, Commits Made, Duration
- **Summary Stats:** Total Active, Total Today, Avg Session Duration

#### 4.2.2 Agent Detail Page (/agent/<id>)

- **Header:** Agent ID, Status, and Worktree Path
- **Current Status:** Feature Description, Current Task, Last Activity Time
- **Activity Timeline:** Chronological list of all actions and status updates
- **Commits Made:** List of git commits with hashes and messages
- **Actions:** End Session, Update Status, View Worktree
- **Navigation:** Back to agent overview

---

## 5. Implementation Plan

### 5.1 Development Phases

#### 5.1.1 Phase 1: Core Infrastructure (Week 1)

- **Deliverables:**
  - SQLite database setup with agents, actions, and commits tables
  - Basic MCP server implementation
  - Core MCP tools (register_agent, update_status, log_action, log_commits, end_session)
- **Acceptance Criteria:**
  - Claude Code can successfully call all MCP tools
  - Agent data persists correctly in SQLite
  - Multiple agents can be tracked simultaneously

#### 5.1.2 Phase 2: Web Dashboard (Week 2)

- **Deliverables:**
  - Flask application with agent overview page
  - Agent detail view with activity timeline
  - Bootstrap styling with status indicators
- **Acceptance Criteria:**
  - Dashboard displays all active agents correctly
  - Agent detail page shows complete activity history
  - Visual indicators clearly show agent status

#### 5.1.3 Phase 3: Multi-Agent Integration & Testing (Week 3)

- **Deliverables:**
  - End-to-end testing with multiple concurrent Claude Code instances
  - Forgotten agent detection and highlighting
  - Documentation and setup instructions
- **Acceptance Criteria:**
  - Multiple agents can be managed simultaneously without conflicts
  - Dashboard clearly identifies inactive agents
  - Complete multi-agent workflow functions correctly

### 5.2 Success Criteria

- **Technical:** All functional requirements implemented and tested with multiple concurrent agents
- **Usability:** Dashboard provides clear visibility into all active Claude Code instances
- **Performance:** System handles at least 10 concurrent agents without performance degradation
- **Integration:** Seamless multi-agent workflow with proper status tracking

### 5.3 Future Enhancements (Post-MVP)

- Real-time dashboard updates (WebSocket integration)
- Agent interaction capabilities (send commands, kill agents)
- Advanced filtering and search across agents
- Agent performance metrics and analytics
- Multi-user support with team agent visibility
- Agent template and automation features

---

## 6. Risk Assessment

### 6.1 Technical Risks

- **Risk:** MCP protocol integration complexity with Claude Code
- **Mitigation:** Start with simple tool calls, build incrementally

- **Risk:** SQLite database conflicts with concurrent agent updates
- **Mitigation:** Implement proper database locking and transaction handling

### 6.2 User Experience Risks

- **Risk:** Dashboard becomes cluttered with many active agents
- **Mitigation:** Implement grouping, filtering, and priority indicators

- **Risk:** Agents failing to report status causing inaccurate dashboard
- **Mitigation:** Implement timeout detection and manual status override

### 6.3 Integration Risks

- **Risk:** Claude Code instances not reliably calling MCP tools
- **Mitigation:** Clear documentation and simple tool interface design

---

## 7. Appendices

### 7.1 Glossary

- **Agent:** A single Claude Code instance working on a specific feature or task
- **MCP:** Model Context Protocol - Communication standard for AI tools
- **Worktree:** Git worktree - separate working directory for different features
- **Session:** Complete lifecycle of an agent from registration to completion

### 7.2 References

- Model Context Protocol Documentation
- Claude Code Documentation
- Flask Documentation
- SQLite Documentation

---

**Document Approval:**

- [ ] Technical Requirements Review
- [ ] Multi-Agent Workflow Validation
- [ ] User Experience Review
- [ ] Final Approval

**Next Steps:**

1. Review and approve this PRD
2. Set up development environment with MCP SDK
3. Begin Phase 1 implementation
4. Test with multiple concurrent Claude Code instances
