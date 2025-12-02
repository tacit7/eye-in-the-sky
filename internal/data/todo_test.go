package data

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// setupTestDB creates an in-memory database with all necessary schema
func setupTestDB(t *testing.T) *database.DB {
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Wrap it in our DB type
	db := database.NewFromConnection(conn)

	// Create minimal schema matching actual database
	schema := `
	CREATE TABLE workflow_states (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		position INTEGER,
		color TEXT,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE projects (
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

	CREATE TABLE tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		project_id TEXT,
		state_id INTEGER DEFAULT 1,
		priority INTEGER DEFAULT 0,
		due_at DATETIME,
		completed_at DATETIME,
		session_id TEXT,
		agent_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		archived BOOLEAN DEFAULT 0,
		FOREIGN KEY (project_id) REFERENCES projects(id),
		FOREIGN KEY (state_id) REFERENCES workflow_states(id)
	);

	CREATE TABLE task_notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		author TEXT,
		body TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	);

	CREATE TABLE tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		color TEXT
	);

	CREATE TABLE task_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT NOT NULL,
		tag_id INTEGER NOT NULL,
		UNIQUE(task_id, tag_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id),
		FOREIGN KEY (tag_id) REFERENCES tags(id)
	);

	-- Insert default workflow states
	INSERT INTO workflow_states (id, name, position) VALUES
		(1, 'todo', 1),
		(2, 'in_progress', 2),
		(3, 'done', 3);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// TestNewTodoStore tests basic TodoStore creation and null handling
func TestNewTodoStore(t *testing.T) {
	// Test with nil service
	store := NewTodoStore(nil)
	if store == nil {
		t.Fatal("NewTodoStore with nil returned nil")
	}

	// Test basic operations don't crash
	ctx := context.Background()
	tasks, err := store.LoadByAgent(ctx, domain.AgentID("test"), 10, 0)
	if err != nil {
		t.Fatalf("LoadByAgent failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks with nil service, got %d", len(tasks))
	}
}

// TestTodoStoreNilHandling tests that nil service is handled gracefully
func TestTodoStoreNilHandling(t *testing.T) {
	store := NewTodoStore(nil)

	ctx := context.Background()

	// Test LoadCountsByAgent with nil service
	agents := []domain.Agent{
		{ID: domain.AgentID("test-agent"), SessionID: "test-session"},
	}
	counts, err := store.LoadCountsByAgent(ctx, agents)
	if err != nil {
		t.Fatalf("LoadCountsByAgent failed: %v", err)
	}

	if len(counts) == 0 {
		t.Log("LoadCountsByAgent correctly returns empty with nil service")
	}

	// Test LoadRecentByAgent with nil service
	recent, err := store.LoadRecentByAgent(ctx, domain.AgentID("test"), 10)
	if err != nil {
		t.Fatalf("LoadRecentByAgent failed: %v", err)
	}

	if len(recent) != 0 {
		t.Errorf("Expected 0 recent tasks, got %d", len(recent))
	}
}

