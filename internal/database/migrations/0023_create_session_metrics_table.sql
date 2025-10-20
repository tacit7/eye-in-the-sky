-- Create session_metrics table to track Claude Code session costs and token usage
CREATE TABLE IF NOT EXISTS session_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    session_id TEXT,
    tokens_used INTEGER NOT NULL,
    tokens_budget INTEGER NOT NULL,
    tokens_remaining INTEGER NOT NULL,
    input_tokens INTEGER,
    output_tokens INTEGER,
    estimated_cost_usd REAL,
    model_name TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_session_metrics_agent_id ON session_metrics(agent_id);
CREATE INDEX IF NOT EXISTS idx_session_metrics_session_id ON session_metrics(session_id);
CREATE INDEX IF NOT EXISTS idx_session_metrics_timestamp ON session_metrics(timestamp);
