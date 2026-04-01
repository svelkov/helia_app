package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	conn *sqlx.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(conn *sqlx.DB) *PostgresDB {
	return &PostgresDB{conn: conn}
}

// GetContext retrieves a single row with context
func (p *PostgresDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return p.conn.GetContext(ctx, dest, query, args...)
}

// Get retrieves a single row (legacy, prefer GetContext)
func (p *PostgresDB) Get(dest interface{}, query string, args ...interface{}) error {
	return p.conn.Get(dest, query, args...)
}

// SelectContext retrieves multiple rows with context
func (p *PostgresDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return p.conn.SelectContext(ctx, dest, query, args...)
}

// Select retrieves multiple rows (legacy, prefer SelectContext)
func (p *PostgresDB) Select(dest interface{}, query string, args ...interface{}) error {
	return p.conn.Select(dest, query, args...)
}

// QueryRowContext executes a query that returns a single row with context
func (p *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.conn.QueryRowContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row (legacy, prefer QueryRowContext)
func (p *PostgresDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.conn.QueryRow(query, args...)
}

// QueryContext executes a query with context
func (p *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.conn.QueryContext(ctx, query, args...)
}

// ExecContext executes a query (INSERT, UPDATE, DELETE) with context
func (p *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.conn.ExecContext(ctx, query, args...)
}

// Beginx starts a new transaction
func (p *PostgresDB) Beginx() (Transaction, error) {
	tx, err := p.conn.Beginx()
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %v", err)
	}
	return &PostgresTx{tx: tx}, nil
}

// Ping verifies the database connection
func (p *PostgresDB) Ping() error {
	return p.conn.Ping()
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.conn.Close()
}

// PostgresTx implements the Transaction interface for PostgreSQL
type PostgresTx struct {
	tx *sqlx.Tx
}

// ExecContext executes a query within a transaction with context
func (t *PostgresTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// Exec executes a query within a transaction (legacy, prefer ExecContext)
func (t *PostgresTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// QueryRowContext executes a query that returns a single row within a transaction with context
func (t *PostgresTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row within a transaction (legacy, prefer QueryRowContext)
func (t *PostgresTx) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// Commit commits the transaction
func (t *PostgresTx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PostgresTx) Rollback() error {
	return t.tx.Rollback()
}
