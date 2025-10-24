-- Create todo system tables
-- Projects for organizing tasks
CREATE TABLE IF NOT EXISTS projects (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	path TEXT UNIQUE,
	remote_url TEXT,
	subpath TEXT,
	module TEXT,
	salt TEXT,
	id_algorithm TEXT DEFAULT 'uuidv5',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	last_commit TEXT,
	active BOOLEAN DEFAULT 1
);

-- Global workflow states (e.g., todo, in_progress, done)
CREATE TABLE IF NOT EXISTS workflow_states (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL,
	position INTEGER,
	color TEXT,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT,
	state_id INTEGER,
	project_id TEXT,
	priority INTEGER DEFAULT 0,
	due_at DATETIME,
	completed_at DATETIME,
	session_id TEXT,
	agent_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME,
	archived BOOLEAN DEFAULT 0,
	FOREIGN KEY (state_id) REFERENCES workflow_states(id),
	FOREIGN KEY (project_id) REFERENCES projects(id),
	FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Task notes
CREATE TABLE IF NOT EXISTS task_notes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id TEXT,
	author TEXT,
	body TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL,
	color TEXT
);

-- Task-Tag association
CREATE TABLE IF NOT EXISTS task_tags (
	task_id TEXT,
	tag_id INTEGER,
	PRIMARY KEY (task_id, tag_id),
	FOREIGN KEY (task_id) REFERENCES tasks(id),
	FOREIGN KEY (tag_id) REFERENCES tags(id)
);

-- Task events for audit trail
CREATE TABLE IF NOT EXISTS task_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id TEXT,
	event_type TEXT,
	payload TEXT,
	actor TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- Meta table for storing configuration and last-update timestamps
CREATE TABLE IF NOT EXISTS meta (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- FTS5 virtual table for full-text search on tasks
CREATE VIRTUAL TABLE IF NOT EXISTS task_search USING fts5(
	task_id UNINDEXED,
	title,
	description,
	tokenize='porter'
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_tasks_session ON tasks(session_id);
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent_id);
CREATE INDEX IF NOT EXISTS idx_task_notes_task_id ON task_notes(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_task ON task_tags(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_task_events_task ON task_events(task_id);

-- Trigger to auto-update tasks.updated_at
CREATE TRIGGER IF NOT EXISTS update_tasks_updated_at
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
	UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to auto-update projects.updated_at
CREATE TRIGGER IF NOT EXISTS update_projects_updated_at
AFTER UPDATE ON projects
FOR EACH ROW
BEGIN
	UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to auto-update workflow_states.updated_at
CREATE TRIGGER IF NOT EXISTS update_workflow_states_updated_at
AFTER UPDATE ON workflow_states
FOR EACH ROW
BEGIN
	UPDATE workflow_states SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Triggers for FTS5 sync
-- Insert into FTS5 when a task is created
CREATE TRIGGER IF NOT EXISTS sync_task_search_insert
AFTER INSERT ON tasks
FOR EACH ROW
BEGIN
	INSERT INTO task_search(task_id, title, description)
	VALUES (
		NEW.id,
		NEW.title,
		NEW.description
	);
END;

-- Update FTS5 when a task is updated
CREATE TRIGGER IF NOT EXISTS sync_task_search_update
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
	UPDATE task_search
	SET
		title = NEW.title,
		description = NEW.description
	WHERE task_id = NEW.id;
END;

-- Delete from FTS5 when a task is deleted
CREATE TRIGGER IF NOT EXISTS sync_task_search_delete
AFTER DELETE ON tasks
FOR EACH ROW
BEGIN
	DELETE FROM task_search WHERE task_id = OLD.id;
END;

-- Initialize default workflow states if not present
INSERT OR IGNORE INTO workflow_states (id, name, position) VALUES
	(1, 'todo', 1),
	(2, 'in_progress', 2),
	(3, 'done', 3);
