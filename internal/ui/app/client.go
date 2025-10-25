package app

import (
	"database/sql"

	"github.com/tacit7/eye-in-the-sky/internal/data"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/todo"
)

// DataClient aggregates all data stores
type DataClient struct {
	Agents   AgentStore
	Tasks    TaskStore
	Metrics  MetricsStore
	Notes    NotesStore
	Commits  CommitsStore
	Actions  ActionsStore
	Logs     LogsStore
}

// NewDataClient creates a new data client with all stores initialized
func NewDataClient(db *sql.DB) *DataClient {
	// Create database wrapper for todo service
	// Note: We pass the wrapper directly. The todo service will use its methods
	// which are compatible with database.DB interface
	dbWrapper := NewDBWrapper(db)
	// Create todo service - it expects a database-like interface
	// Our DBWrapper implements the necessary methods
	todoService := createTodoServiceFromWrapper(dbWrapper)

	return &DataClient{
		Agents:   data.NewAgentStore(db),
		Tasks:    data.NewTodoStore(todoService),
		Metrics:  data.NewMetricsStore(db),
		Notes:    data.NewNotesStore(db),
		Commits:  data.NewCommitsStore(db),
		Actions:  data.NewActionsStore(db),
		Logs:     data.NewLogsStore(db),
	}
}

// DBWrapper wraps sql.DB to implement the database.DB interface for todo service
type DBWrapper struct {
	conn *sql.DB
}

// NewDBWrapper creates a new database wrapper
func NewDBWrapper(conn *sql.DB) *DBWrapper {
	return &DBWrapper{conn: conn}
}

// Exec wraps sql.DB.Exec
func (db *DBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Query wraps sql.DB.Query
func (db *DBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow wraps sql.DB.QueryRow
func (db *DBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// Close closes the database connection
func (db *DBWrapper) Close() error {
	return db.conn.Close()
}

// Health checks database connectivity
func (db *DBWrapper) Health() error {
	return db.conn.Ping()
}

// RunMigrations is a no-op for wrapped sql.DB (migrations already run)
func (db *DBWrapper) RunMigrations() error {
	return nil
}

// Reindex is a no-op for wrapped sql.DB
func (db *DBWrapper) Reindex() error {
	_, err := db.conn.Exec("REINDEX task_search;")
	return err
}

// Vacuum is a no-op for wrapped sql.DB
func (db *DBWrapper) Vacuum() error {
	_, err := db.conn.Exec("VACUUM; ANALYZE;")
	return err
}

// BeginTx begins a transaction
func (db *DBWrapper) BeginTx(ctx interface{}, opts interface{}) (interface{}, error) {
	return db.conn.Begin()
}

// createTodoServiceFromWrapper creates a todo service using the wrapper
// It wraps the existing sql.DB connection and creates a todo service from it
func createTodoServiceFromWrapper(dbWrapper *DBWrapper) *todo.Service {
	if dbWrapper == nil || dbWrapper.conn == nil {
		return nil
	}

	// Wrap the existing sql.DB connection in a database.DB
	db := database.NewFromConnection(dbWrapper.conn)

	// Create and return the todo service
	return todo.NewService(db)
}