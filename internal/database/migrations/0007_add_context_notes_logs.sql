-- Add session_context table for storing agent session state and progress
CREATE TABLE IF NOT EXISTS session_context (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Session state
    current_phase TEXT,
    overall_progress REAL DEFAULT 0.0,

    -- Tasks
    pending_tasks TEXT,      -- JSON array of pending tasks
    completed_tasks TEXT,    -- JSON array of completed tasks
    next_actions TEXT,       -- JSON array of next planned actions

    -- Dependencies and files
    dependencies TEXT,       -- JSON array of dependency information
    important_files TEXT,    -- JSON array of critical file paths

    -- Progress tracking
    milestones TEXT,         -- JSON array of milestone objects
    current_goals TEXT,      -- JSON array of current goals
    blockers TEXT,           -- JSON array of blocker objects

    -- Key decisions
    key_decisions TEXT,      -- JSON array of decision objects

    -- Environment context
    environment TEXT,        -- JSON object for environment info

    -- Metrics
    metrics TEXT,            -- JSON object for quantitative metrics

    -- Session metadata
    auto_save BOOLEAN DEFAULT 0,

    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Add session_notes table for contextual notes
CREATE TABLE IF NOT EXISTS session_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    note_type TEXT NOT NULL,         -- "observation", "decision", "blocker", "reminder", "learning"
    content TEXT NOT NULL,
    priority TEXT NOT NULL,          -- "low", "medium", "high", "critical"
    tags TEXT,                       -- JSON array of tags

    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Add session_logs table for detailed activity logs
CREATE TABLE IF NOT EXISTS session_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    log_level TEXT NOT NULL,         -- "debug", "info", "warning", "error", "critical"
    category TEXT NOT NULL,          -- "file_operation", "git", "mcp_call", "tool_usage", "decision", "error"
    message TEXT NOT NULL,
    details TEXT,                    -- JSON object for additional context

    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Create indexes for performance
CREATE INDEX idx_session_context_agent_id ON session_context(agent_id);
CREATE INDEX idx_session_context_session_id ON session_context(session_id);
CREATE INDEX idx_session_context_updated_at ON session_context(updated_at);

CREATE INDEX idx_session_notes_agent_id ON session_notes(agent_id);
CREATE INDEX idx_session_notes_timestamp ON session_notes(timestamp);
CREATE INDEX idx_session_notes_priority ON session_notes(priority);
CREATE INDEX idx_session_notes_note_type ON session_notes(note_type);

CREATE INDEX idx_session_logs_agent_id ON session_logs(agent_id);
CREATE INDEX idx_session_logs_timestamp ON session_logs(timestamp);
CREATE INDEX idx_session_logs_log_level ON session_logs(log_level);
CREATE INDEX idx_session_logs_category ON session_logs(category);

-- Add trigger to auto-update session_context updated_at
CREATE TRIGGER update_session_context_updated_at
    AFTER UPDATE ON session_context
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE session_context SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
