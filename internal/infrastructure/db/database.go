package db

import (
	"context"
	"database/sql"
)

// Database defines the interface for database operations
// This abstraction allows for easier testing and potential database switching
type Database interface {
	// Query methods
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Transaction methods
	Beginx() (Transaction, error)

	// Connection methods
	Ping() error
	Close() error
}

// Transaction defines the interface for database transactions
type Transaction interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Commit() error
	Rollback() error
}
