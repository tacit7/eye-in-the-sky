-- Migration 002: Simplify context tables to use markdown format
-- This migration converts session_context to use a single TEXT field like agent_context

-- Create new simplified session_context table
CREATE TABLE IF NOT EXISTS session_context_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    context TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Copy existing data to new table (if any exists)
-- Note: This will lose the structured data and convert to a simple text format
INSERT INTO session_context_new (id, agent_id, session_id, context, created_at, updated_at)
SELECT
    id,
    agent_id,
    session_id,
    'Phase: ' || COALESCE(current_phase, 'N/A') || '\n\n' ||
    'Completed Tasks:\n' || COALESCE(completed_tasks, '[]') || '\n\n' ||
    'Pending Tasks:\n' || COALESCE(pending_tasks, '[]') || '\n\n' ||
    'Next Actions:\n' || COALESCE(next_actions, '[]') || '\n\n' ||
    'Learned Context:\n' || COALESCE(learned_context, 'N/A'),
    created_at,
    updated_at
FROM session_context;

-- Drop old table
DROP TABLE session_context;

-- Rename new table to session_context
ALTER TABLE session_context_new RENAME TO session_context;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_session_context_agent_id ON session_context(agent_id);
CREATE INDEX IF NOT EXISTS idx_session_context_session_id ON session_context(session_id);
CREATE INDEX IF NOT EXISTS idx_session_context_updated_at ON session_context(updated_at);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_session_context_updated_at
AFTER UPDATE ON session_context
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE session_context SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Ensure agent_context table exists with proper structure
CREATE TABLE IF NOT EXISTS agent_context (
    agent_id TEXT NOT NULL,
    project_id INTEGER NOT NULL,
    context TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (agent_id, project_id),
    FOREIGN KEY (agent_id) REFERENCES agents(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- Create trigger for agent_context
CREATE TRIGGER IF NOT EXISTS agent_context_update
AFTER UPDATE ON agent_context
FOR EACH ROW
BEGIN
    UPDATE agent_context
    SET updated_at = CURRENT_TIMESTAMP
    WHERE agent_id = NEW.agent_id
      AND project_id = NEW.project_id;
END;
