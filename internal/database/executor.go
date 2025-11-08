package database

import "database/sql"

// Executor is an interface that can execute database queries.
// Both DB and Tx implement this interface, allowing repositories
// to work with either a direct connection or within a transaction.
type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
