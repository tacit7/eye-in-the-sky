package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

// CCUsageDB manages the ccusage SQLite database
type CCUsageDB struct {
	db       *sql.DB
	mutex    sync.RWMutex
	ownsConn bool // Track if we own the connection (for backwards compatibility)
}

// New creates a new CCUsageDB instance and initializes the schema
// Deprecated: Use NewWithConnection instead to share the main database connection
func New(dbPath string) (*CCUsageDB, error) {
	db, err := sql.Open("sqlite", "file:"+dbPath+"?cache=shared&mode=rwc&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings for concurrency
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	ccdb := &CCUsageDB{
		db:       db,
		ownsConn: true, // We created the connection, we own it
	}

	// Initialize schema
	if err := ccdb.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return ccdb, nil
}

// NewWithConnection creates a new CCUsageDB instance using an existing database connection
// This is preferred for sharing the main eits.db connection
func NewWithConnection(db *sql.DB) (*CCUsageDB, error) {
	ccdb := &CCUsageDB{
		db:       db,
		ownsConn: false, // We don't own the connection
	}

	// Initialize schema (tables will be created if they don't exist)
	if err := ccdb.initSchema(); err != nil {
		return nil, err
	}

	return ccdb, nil
}

// initSchema creates the database schema if it doesn't exist
func (c *CCUsageDB) initSchema() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.db.Exec(InitDB())
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// Close closes the database connection
// Only closes the connection if we own it (created via New)
func (c *CCUsageDB) Close() error {
	if c.ownsConn {
		return c.db.Close()
	}
	// Don't close shared connections
	return nil
}

// InsertUsageEntry inserts a usage entry into the database
func (c *CCUsageDB) InsertUsageEntry(entry UsageEntryRow) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	query := `
		INSERT INTO usage_entries (
			session_id, timestamp, project, model,
			input_tokens, output_tokens,
			cache_creation_tokens, cache_read_tokens,
			total_cost, message_id, request_id, unique_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := c.db.Exec(query,
		entry.SessionID,
		entry.Timestamp,
		entry.Project,
		entry.Model,
		entry.InputTokens,
		entry.OutputTokens,
		entry.CacheCreationTokens,
		entry.CacheReadTokens,
		entry.TotalCost,
		entry.MessageID,
		entry.RequestID,
		entry.UniqueHash,
	)

	if err != nil {
		return fmt.Errorf("failed to insert usage entry: %w", err)
	}

	return nil
}

// BatchInsertUsageEntries inserts multiple usage entries in a transaction
func (c *CCUsageDB) BatchInsertUsageEntries(entries []UsageEntryRow) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO usage_entries (
			session_id, timestamp, project, model,
			input_tokens, output_tokens,
			cache_creation_tokens, cache_read_tokens,
			total_cost, message_id, request_id, unique_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err := stmt.Exec(
			entry.SessionID,
			entry.Timestamp,
			entry.Project,
			entry.Model,
			entry.InputTokens,
			entry.OutputTokens,
			entry.CacheCreationTokens,
			entry.CacheReadTokens,
			entry.TotalCost,
			entry.MessageID,
			entry.RequestID,
			entry.UniqueHash,
		)
		if err != nil {
			return fmt.Errorf("failed to insert entry: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateFileMetadata updates the file metadata for tracking changes
func (c *CCUsageDB) UpdateFileMetadata(filePath string, mtime int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	query := `
		INSERT INTO file_metadata (file_path, last_mtime, last_parsed_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(file_path) DO UPDATE SET
			last_mtime = excluded.last_mtime,
			last_parsed_at = datetime('now')
	`

	_, err := c.db.Exec(query, filePath, mtime)
	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	return nil
}

// GetFileMetadata retrieves metadata for a file
func (c *CCUsageDB) GetFileMetadata(filePath string) (int64, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var mtime int64
	err := c.db.QueryRow(
		"SELECT last_mtime FROM file_metadata WHERE file_path = ?",
		filePath,
	).Scan(&mtime)

	if err == sql.ErrNoRows {
		return 0, nil // File not yet tracked
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return mtime, nil
}

// UsageEntryRow represents a row to be inserted into the database
type UsageEntryRow struct {
	SessionID            string
	Timestamp            string
	Project              string
	Model                string
	InputTokens          int
	OutputTokens         int
	CacheCreationTokens  int
	CacheReadTokens      int
	TotalCost            float64
	MessageID            string
	RequestID            string
	UniqueHash           string
}
