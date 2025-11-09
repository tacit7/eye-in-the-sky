-- Migration: Convert tasks.session_id to task_sessions junction table
-- This migration creates a many-to-many relationship between tasks and sessions

-- Step 1: Create the task_sessions junction table
CREATE TABLE IF NOT EXISTS task_sessions (
    task_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (task_id, session_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Step 2: Create indexes for the junction table
CREATE INDEX IF NOT EXISTS idx_task_sessions_task ON task_sessions(task_id);
CREATE INDEX IF NOT EXISTS idx_task_sessions_session ON task_sessions(session_id);

-- Step 3: Migrate existing data from tasks.session_id to task_sessions
-- Only migrate rows where session_id is not NULL
INSERT INTO task_sessions (task_id, session_id, created_at)
SELECT id, session_id, created_at
FROM tasks
WHERE session_id IS NOT NULL;

-- Step 4: Drop the old session_id column from tasks table
-- Note: SQLite doesn't support DROP COLUMN directly in older versions
-- We need to recreate the table without the session_id column

-- Begin transaction for table recreation
BEGIN TRANSACTION;

-- Create new tasks table without session_id column
CREATE TABLE tasks_new (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    state_id INTEGER,
    project_id TEXT,
    priority INTEGER DEFAULT 0,
    due_at DATETIME,
    completed_at DATETIME,
    agent_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    archived BOOLEAN DEFAULT 0,
    FOREIGN KEY (state_id) REFERENCES workflow_states(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Copy data from old table to new table (excluding session_id)
INSERT INTO tasks_new (id, title, description, state_id, project_id, priority, due_at, completed_at, agent_id, created_at, updated_at, archived)
SELECT id, title, description, state_id, project_id, priority, due_at, completed_at, agent_id, created_at, updated_at, archived
FROM tasks;

-- Drop old table
DROP TABLE tasks;

-- Rename new table to tasks
ALTER TABLE tasks_new RENAME TO tasks;

-- Recreate indexes for tasks table (excluding idx_tasks_session)
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent_id);

-- Recreate update trigger for tasks
CREATE TRIGGER IF NOT EXISTS update_tasks_updated_at
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
    UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

COMMIT;
