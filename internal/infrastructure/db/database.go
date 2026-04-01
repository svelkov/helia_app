package db

import (
	"context"
	"database/sql"
)

// Database defines the interface for database operations
// This abstraction allows for easier testing and potential database switching
type Database interface {
	// Query methods (context-aware - preferred)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Execution methods (for INSERT, UPDATE, DELETE)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Transaction methods
	Beginx() (Transaction, error)

	// Connection methods
	Ping() error
	Close() error
}

// Transaction defines the interface for database transactions
type Transaction interface {
	// Context-aware methods (preferred)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Commit and Rollback
	Commit() error
	Rollback() error
}
