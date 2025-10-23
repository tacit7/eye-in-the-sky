package db

import (
	"fmt"
	"time"
)

// SetupTriggers creates all database triggers for automatic updates and FTS sync.
func (db *DB) SetupTriggers() error {
	triggers := []string{
		updateTasksUpdatedAtOnUpdate,
		syncTaskSearchOnInsert,
		syncTaskSearchOnUpdate,
		syncTaskSearchOnDelete,
		syncLatestNoteToSearchOnNoteInsert,
		updateProjectsUpdatedAtOnUpdate,
		syncWorkflowStatesUpdatedAtOnUpdate,
	}

	for _, trigger := range triggers {
		if _, err := db.conn.Exec(trigger); err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}
	}

	return nil
}

// Update updated_at on tasks table
const updateTasksUpdatedAtOnUpdate = `
CREATE TRIGGER IF NOT EXISTS update_tasks_updated_at
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
	UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`

// Update updated_at on projects table
const updateProjectsUpdatedAtOnUpdate = `
CREATE TRIGGER IF NOT EXISTS update_projects_updated_at
AFTER UPDATE ON projects
FOR EACH ROW
BEGIN
	UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`

// Update updated_at on workflow_states table
const syncWorkflowStatesUpdatedAtOnUpdate = `
CREATE TRIGGER IF NOT EXISTS update_workflow_states_updated_at
AFTER UPDATE ON workflow_states
FOR EACH ROW
BEGIN
	UPDATE workflow_states SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`

// Sync task_search on task insert
const syncTaskSearchOnInsert = `
CREATE TRIGGER IF NOT EXISTS sync_task_search_insert
AFTER INSERT ON tasks
FOR EACH ROW
BEGIN
	INSERT INTO task_search(rowid, description, latest_note, tags)
	VALUES (
		NEW.id,
		NEW.description,
		(SELECT body_markdown FROM task_notes WHERE task_id = NEW.id ORDER BY created_at DESC LIMIT 1),
		(SELECT group_concat(t.name, ' ') FROM task_tags tt JOIN tags t ON tt.tag_id = t.id WHERE tt.task_id = NEW.id)
	);
END;
`

// Sync task_search on task update
const syncTaskSearchOnUpdate = `
CREATE TRIGGER IF NOT EXISTS sync_task_search_update
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
	UPDATE task_search
	SET
		description = NEW.description,
		latest_note = (SELECT body_markdown FROM task_notes WHERE task_id = NEW.id ORDER BY created_at DESC LIMIT 1),
		tags = (SELECT group_concat(t.name, ' ') FROM task_tags tt JOIN tags t ON tt.tag_id = t.id WHERE tt.task_id = NEW.id)
	WHERE rowid = NEW.id;
END;
`

// Sync task_search on task delete
const syncTaskSearchOnDelete = `
CREATE TRIGGER IF NOT EXISTS sync_task_search_delete
AFTER DELETE ON tasks
FOR EACH ROW
BEGIN
	DELETE FROM task_search WHERE rowid = OLD.id;
END;
`

// Sync latest note when a note is inserted
const syncLatestNoteToSearchOnNoteInsert = `
CREATE TRIGGER IF NOT EXISTS sync_latest_note_on_insert
AFTER INSERT ON task_notes
FOR EACH ROW
BEGIN
	UPDATE task_search
	SET latest_note = NEW.body_markdown
	WHERE rowid = NEW.task_id;
END;
`

// MaybeReindex checks if it's Sunday and 7 days have passed since last reindex.
func (db *DB) MaybeReindex() error {
	now := time.Now()
	if now.Weekday() != time.Sunday {
		return nil
	}

	// Get last reindex date from meta table
	var lastReindexStr string
	err := db.conn.QueryRow("SELECT value FROM meta WHERE key = 'last_reindex_at'").Scan(&lastReindexStr)
	if err != nil {
		// No record yet, set it and reindex
		if _, err := db.conn.Exec("INSERT OR REPLACE INTO meta (key, value) VALUES (?, ?)", "last_reindex_at", now.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("failed to set last_reindex_at: %w", err)
		}
		return db.Reindex()
	}

	lastReindex, err := time.Parse(time.RFC3339, lastReindexStr)
	if err != nil {
		return fmt.Errorf("failed to parse last_reindex_at: %w", err)
	}

	// If 7 days have passed, reindex
	if now.Sub(lastReindex) >= 7*24*time.Hour {
		if err := db.Reindex(); err != nil {
			return err
		}
		// Update meta
		if _, err := db.conn.Exec("UPDATE meta SET value = ? WHERE key = 'last_reindex_at'", now.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("failed to update last_reindex_at: %w", err)
		}
	}

	return nil
}

// Reindex rebuilds the FTS5 task_search index.
func (db *DB) Reindex() error {
	if _, err := db.conn.Exec("REINDEX task_search;"); err != nil {
		return fmt.Errorf("failed to reindex task_search: %w", err)
	}
	return nil
}

// Vacuum performs VACUUM and ANALYZE on the database.
func (db *DB) Vacuum() error {
	if _, err := db.conn.Exec("VACUUM;"); err != nil {
		return fmt.Errorf("failed to vacuum: %w", err)
	}
	if _, err := db.conn.Exec("ANALYZE;"); err != nil {
		return fmt.Errorf("failed to analyze: %w", err)
	}
	return nil
}
