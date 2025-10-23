package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFileName = "todo.db"
	configDir  = ".config/eye-in-the-sky"
)

type DB struct {
	conn *sql.DB
}

// OpenDB opens or creates the todo database with FTS5 enabled.
// Database location: ~/.config/eye-in-the-sky/todo.db
func OpenDB() (*DB, error) {
	// Ensure config directory exists
	configPath := filepath.Join(os.Getenv("HOME"), configDir)
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Database file path
	dbPath := filepath.Join(configPath, dbFileName)

	// Open connection with FTS5 enabled
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL", dbPath)
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	db := &DB{conn: conn}

	// Enable FTS5 extension
	if err := db.enableFTS5(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable FTS5: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Setup triggers
	if err := db.SetupTriggers(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to setup triggers: %w", err)
	}

	// Check if weekly reindex is needed
	if err := db.MaybeReindex(); err != nil {
		// Log but don't fail on reindex errors
		fmt.Printf("warning: reindex check failed: %v\n", err)
	}

	return db, nil
}

// enableFTS5 enables the FTS5 extension for full-text search.
func (db *DB) enableFTS5() error {
	// Try to load FTS5; SQLite3 might have it compiled in
	_, err := db.conn.Exec("SELECT fts5(?1)", "version")
	if err != nil {
		// FTS5 not available; this is a limitation of the SQLite3 build
		return fmt.Errorf("FTS5 not available in this SQLite3 build: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying SQL database connection.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Exec executes a statement without returning rows.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Query executes a query that returns rows.
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow executes a query that returns a single row.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// BeginTx starts a transaction.
func (db *DB) BeginTx() (*sql.Tx, error) {
	return db.conn.Begin()
}
