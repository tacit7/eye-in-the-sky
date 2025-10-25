package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL embed.FS

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// NewFromConnection wraps an existing sql.DB connection in a DB struct.
// This is useful when you already have an open connection and want to use it with the database.DB interface.
// Note: No migrations are run, as they should have already been executed.
func NewFromConnection(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) initSchema() error {
	// Use migrations instead of embedded schema
	return db.RunMigrations()
}

func (db *DB) Health() error {
	return db.conn.Ping()
}

// ============================================================================
// Database Query Methods (expose underlying sql.DB methods)
// ============================================================================

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

// ============================================================================
// FTS5 Maintenance Functions
// ============================================================================

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

// MaybeReindex checks if a weekly reindex is needed.
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
