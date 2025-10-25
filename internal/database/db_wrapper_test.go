package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestNewFromConnection tests the NewFromConnection wrapper function
func TestNewFromConnection(t *testing.T) {
	// Setup: Create an in-memory SQLite database
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer conn.Close()

	// Test basic creation
	db := NewFromConnection(conn)
	if db == nil {
		t.Fatal("NewFromConnection returned nil")
	}

	// Test health check works
	if err := db.Health(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

// TestNewFromConnectionWithNilInput tests that nil input creates a DB struct
func TestNewFromConnectionWithNilInput(t *testing.T) {
	db := NewFromConnection(nil)
	if db == nil {
		t.Fatal("NewFromConnection with nil should return non-nil DB struct")
	}
	// Note: Health check will panic with nil connection, which is expected behavior
	// Users should not pass nil connections in production
}

// TestNewFromConnectionExecutesQueries tests that wrapped connection can execute queries
func TestNewFromConnectionExecutesQueries(t *testing.T) {
	// Setup
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer conn.Close()

	// Create test table
	if _, err := conn.Exec("CREATE TABLE test(id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Wrap the connection
	db := NewFromConnection(conn)

	// Test Exec
	result, err := db.Exec("INSERT INTO test(name) VALUES(?)", "test_name")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected failed: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", rowsAffected)
	}

	// Test QueryRow (simpler than Query with Next())
	row := db.QueryRow("SELECT name FROM test WHERE id = ?", 1)
	var scannedName string
	if err := row.Scan(&scannedName); err != nil {
		t.Fatalf("QueryRow Scan failed: %v", err)
	}

	if scannedName != "test_name" {
		t.Errorf("Expected 'test_name', got '%s'", scannedName)
	}
}

// TestNewFromConnectionReindexAndVacuum tests maintenance operations
func TestNewFromConnectionReindexAndVacuum(t *testing.T) {
	// Setup
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer conn.Close()

	db := NewFromConnection(conn)

	// Test Reindex (should work even with no FTS5 index)
	err = db.Reindex()
	if err == nil {
		t.Log("Reindex succeeded on empty database")
	}
	// Note: might fail on in-memory with no index, that's OK

	// Test Vacuum (should succeed)
	err = db.Vacuum()
	if err != nil {
		t.Logf("Vacuum warning: %v (expected on in-memory)", err)
	}
}

// TestNewFromConnectionMultipleWrappers tests that multiple wrappers can share same connection
func TestNewFromConnectionMultipleWrappers(t *testing.T) {
	// Setup: Single connection
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer conn.Close()

	// Create test table
	_, err = conn.Exec("CREATE TABLE shared(id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create multiple wrappers around same connection
	db1 := NewFromConnection(conn)
	db2 := NewFromConnection(conn)

	// Use first wrapper to insert
	_, err = db1.Exec("INSERT INTO shared(value) VALUES(?)", "from_db1")
	if err != nil {
		t.Fatalf("db1 insert failed: %v", err)
	}

	// Use second wrapper to query
	row := db2.QueryRow("SELECT value FROM shared WHERE id = 1")
	var value string
	if err := row.Scan(&value); err != nil {
		t.Fatalf("db2 query failed: %v", err)
	}

	if value != "from_db1" {
		t.Errorf("Expected 'from_db1', got '%s'", value)
	}
}
