-- Migration 0001: Initial schema
-- Eye in the Sky Database Schema

-- Agents table: tracks Claude Code instances
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,              -- 8-char hash like "a3f7d2e1"
    status TEXT NOT NULL,             -- "active", "idle", "working", "completed", "failed"
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    git_worktree_path TEXT,           -- path to git worktree
    feature_description TEXT,         -- what the agent is working on
    current_task TEXT,                -- current specific task
    last_activity_at TIMESTAMP        -- when agent last reported activity
);

-- Actions table: logs all agent activities
CREATE TABLE IF NOT EXISTS actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    action_type TEXT NOT NULL,        -- "task_start", "file_operation", "git_commit", "status_update"
    description TEXT NOT NULL,        -- human-readable description of action
    details TEXT,                     -- JSON blob for additional context
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Commits table: tracks git commits made by agents
CREATE TABLE IF NOT EXISTS commits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    commit_hash TEXT NOT NULL,
    commit_message TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_updated_at ON agents(updated_at);
CREATE INDEX IF NOT EXISTS idx_actions_agent_id ON actions(agent_id);
CREATE INDEX IF NOT EXISTS idx_actions_timestamp ON actions(timestamp);
CREATE INDEX IF NOT EXISTS idx_commits_agent_id ON commits(agent_id);

-- Migration tracking table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Record this migration
INSERT OR IGNORE INTO schema_migrations (version) VALUES ('0001_init');