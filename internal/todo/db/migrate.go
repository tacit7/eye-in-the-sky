package db

import (
	"fmt"
)

// RunMigrations creates all required tables and indexes if they don't exist.
func (db *DB) RunMigrations() error {
	// Create tables in dependency order
	migrations := []string{
		createProjectsTable,
		createWorkflowStatesTable,
		createTasksTable,
		createTaskNotesTable,
		createTagsTable,
		createTaskTagsTable,
		createTaskEventsTable,
		createMetaTable,
		createTaskSearchTable,
		createIndexes,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Add columns for session and agent tracking (if not already present)
	if err := db.addSessionAgentColumns(); err != nil {
		return fmt.Errorf("failed to add session/agent columns: %w", err)
	}

	return nil
}

// addSessionAgentColumns adds session_id and agent_id columns to existing databases
func (db *DB) addSessionAgentColumns() error {
	// Add session_id column if it doesn't exist
	if _, err := db.conn.Exec(`
		ALTER TABLE tasks ADD COLUMN session_id TEXT;
	`); err != nil {
		// Column likely already exists; ignore error
		_ = err
	}

	// Add agent_id column if it doesn't exist
	if _, err := db.conn.Exec(`
		ALTER TABLE tasks ADD COLUMN agent_id TEXT;
	`); err != nil {
		// Column likely already exists; ignore error
		_ = err
	}

	return nil
}

const createProjectsTable = `
CREATE TABLE IF NOT EXISTS projects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uuid TEXT UNIQUE NOT NULL,
	name TEXT NOT NULL,
	repo_slug TEXT UNIQUE,
	archived_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createWorkflowStatesTable = `
CREATE TABLE IF NOT EXISTS workflow_states (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL,
	code TEXT NOT NULL,
	display_name TEXT NOT NULL,
	position INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
	UNIQUE(project_id, code)
);
`

const createTasksTable = `
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL,
	description TEXT NOT NULL,
	state_code TEXT,
	parent_id INTEGER,
	priority INTEGER,
	weight INTEGER,
	position INTEGER DEFAULT 0,
	due_date TIMESTAMP,
	session_id TEXT,
	agent_id TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	archived_at TIMESTAMP,
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
	FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
	CHECK (priority IS NULL OR (priority >= 1 AND priority <= 5))
);
`

const createTaskNotesTable = `
CREATE TABLE IF NOT EXISTS task_notes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	body_markdown TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
`

const createTagsTable = `
CREATE TABLE IF NOT EXISTS tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createTaskTagsTable = `
CREATE TABLE IF NOT EXISTS task_tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
	FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
	UNIQUE(task_id, tag_id)
);
`

const createTaskEventsTable = `
CREATE TABLE IF NOT EXISTS task_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	event_type TEXT NOT NULL,
	old_value TEXT,
	new_value TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
`

const createMetaTable = `
CREATE TABLE IF NOT EXISTS meta (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createTaskSearchTable = `
CREATE VIRTUAL TABLE IF NOT EXISTS task_search USING fts5(
	description,
	latest_note,
	tags,
	content=tasks,
	content_rowid=id
);
`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state_code);
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_archived ON tasks(archived_at);
CREATE INDEX IF NOT EXISTS idx_task_notes_task ON task_notes(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_task ON task_tags(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_task_events_task ON task_events(task_id);
CREATE INDEX IF NOT EXISTS idx_workflow_states_project ON workflow_states(project_id);
`
