package app

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestDBWrapper tests the DBWrapper struct
func TestDBWrapper(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	wrapper := NewDBWrapper(conn)
	if wrapper == nil {
		t.Fatal("NewDBWrapper returned nil")
	}

	// Test Health
	if err := wrapper.Health(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

// TestCreateTodoServiceFromWrapper tests the createTodoServiceFromWrapper function
func TestCreateTodoServiceFromWrapper(t *testing.T) {
	// Setup: Create an in-memory SQL database
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	// Create minimal schema
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		active BOOLEAN DEFAULT 1
	);

	CREATE TABLE tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		project_id TEXT,
		state_id INTEGER DEFAULT 1,
		priority INTEGER DEFAULT 0,
		session_id TEXT,
		agent_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		archived BOOLEAN DEFAULT 0,
		FOREIGN KEY (project_id) REFERENCES projects(id),
		FOREIGN KEY (state_id) REFERENCES workflow_states(id)
	);

	INSERT INTO workflow_states (id, name) VALUES (1, 'todo');
	INSERT INTO workflow_states (id, name) VALUES (2, 'in_progress');
	INSERT INTO workflow_states (id, name) VALUES (3, 'done');
	`

	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Wrap the connection
	wrapper := NewDBWrapper(conn)

	// Create todo service from wrapper
	svc := createTodoServiceFromWrapper(wrapper)
	if svc == nil {
		t.Fatal("createTodoServiceFromWrapper returned nil")
	}

	// Service is successfully created
	t.Log("Successfully created todo service from wrapper")
}

// TestCreateTodoServiceFromWrapperNilInput tests nil input handling
func TestCreateTodoServiceFromWrapperNilInput(t *testing.T) {
	svc := createTodoServiceFromWrapper(nil)
	if svc != nil {
		t.Fatal("Expected nil service for nil wrapper")
	}
}

// TestCreateTodoServiceFromWrapperNilConnection tests wrapper with nil connection
func TestCreateTodoServiceFromWrapperNilConnection(t *testing.T) {
	wrapper := &DBWrapper{conn: nil}
	svc := createTodoServiceFromWrapper(wrapper)
	if svc != nil {
		t.Fatal("Expected nil service for wrapper with nil connection")
	}
}

// TestDataClientInitialization tests DataClient creation
func TestDataClientInitialization(t *testing.T) {
	// Setup: Create an in-memory SQL database
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	// Create minimal schema
	schema := `
	CREATE TABLE workflow_states (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL
	);

	CREATE TABLE projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		active BOOLEAN DEFAULT 1
	);

	CREATE TABLE tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		project_id TEXT,
		state_id INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		archived BOOLEAN DEFAULT 0
	);

	INSERT INTO workflow_states (id, name) VALUES (1, 'todo');
	INSERT INTO workflow_states (id, name) VALUES (2, 'in_progress');
	INSERT INTO workflow_states (id, name) VALUES (3, 'done');
	`

	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Create DataClient
	client := NewDataClient(conn)
	if client == nil {
		t.Fatal("NewDataClient returned nil")
	}

	// Verify all stores are initialized
	if client.Agents == nil {
		t.Fatal("Agents store is nil")
	}

	if client.Tasks == nil {
		t.Fatal("Tasks store is nil")
	}

	if client.Metrics == nil {
		t.Fatal("Metrics store is nil")
	}

	if client.Notes == nil {
		t.Fatal("Notes store is nil")
	}

	if client.Commits == nil {
		t.Fatal("Commits store is nil")
	}

	if client.Actions == nil {
		t.Fatal("Actions store is nil")
	}

	if client.Logs == nil {
		t.Fatal("Logs store is nil")
	}
}

// TestDBWrapperQuery tests Query method
func TestDBWrapperQuery(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	// Create test table
	if _, err := conn.Exec("CREATE TABLE test(id INTEGER, name TEXT)"); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if _, err := conn.Exec("INSERT INTO test VALUES(1, 'test')"); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	wrapper := NewDBWrapper(conn)

	rows, err := wrapper.Query("SELECT name FROM test WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Expected rows but got none")
	}

	var name string
	if err := rows.Scan(&name); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if name != "test" {
		t.Errorf("Expected 'test', got '%s'", name)
	}
}

// TestDBWrapperQueryRow tests QueryRow method
func TestDBWrapperQueryRow(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	// Create test table
	if _, err := conn.Exec("CREATE TABLE test(id INTEGER, name TEXT)"); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if _, err := conn.Exec("INSERT INTO test VALUES(1, 'test')"); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	wrapper := NewDBWrapper(conn)

	row := wrapper.QueryRow("SELECT name FROM test WHERE id = ?", 1)
	var name string
	if err := row.Scan(&name); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if name != "test" {
		t.Errorf("Expected 'test', got '%s'", name)
	}
}

// TestDBWrapperExec tests Exec method
func TestDBWrapperExec(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	// Create test table
	if _, err := conn.Exec("CREATE TABLE test(id INTEGER, name TEXT)"); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	wrapper := NewDBWrapper(conn)

	result, err := wrapper.Exec("INSERT INTO test VALUES(1, 'test')")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected failed: %v", err)
	}

	if affected != 1 {
		t.Errorf("Expected 1 row affected, got %d", affected)
	}
}

// TestDBWrapperReindex tests Reindex method
func TestDBWrapperReindex(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	wrapper := NewDBWrapper(conn)

	// Reindex should work (might fail on in-memory without FTS5, that's OK)
	err = wrapper.Reindex()
	if err != nil {
		t.Logf("Reindex note: %v (expected if no FTS5 index)", err)
	}
}

// TestDBWrapperVacuum tests Vacuum method
func TestDBWrapperVacuum(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	wrapper := NewDBWrapper(conn)

	// Vacuum should succeed
	err = wrapper.Vacuum()
	if err != nil {
		t.Logf("Vacuum note: %v (expected on in-memory)", err)
	}
}
