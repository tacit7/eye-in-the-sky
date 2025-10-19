-- Create compactions table for tracking conversation compaction events
CREATE TABLE compactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    old_session_id TEXT,  -- Session that was compacted (if trackable)
    new_session_id TEXT NOT NULL,  -- New session after compaction
    compacted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    summary TEXT,  -- Extracted summary from compaction
    jsonl_file_path TEXT,  -- Path to backed-up JSONL file
    jsonl_file_size INTEGER,  -- File size in bytes
    message_count INTEGER,  -- Number of messages in conversation
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Indexes for efficient queries
CREATE INDEX idx_compactions_agent_id ON compactions(agent_id);
CREATE INDEX idx_compactions_new_session_id ON compactions(new_session_id);
CREATE INDEX idx_compactions_compacted_at ON compactions(compacted_at);
