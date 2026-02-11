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

// Get retrieves a single row
func (p *PostgresDB) Get(dest interface{}, query string, args ...interface{}) error {
	return p.conn.Get(dest, query, args...)
}

// Select retrieves multiple rows
func (p *PostgresDB) Select(dest interface{}, query string, args ...interface{}) error {
	return p.conn.Select(dest, query, args...)
}

// QueryRow executes a query that returns a single row
func (p *PostgresDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.conn.QueryRow(query, args...)
}

// QueryContext executes a query with context
func (p *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.conn.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row with context
func (p *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.conn.QueryRowContext(ctx, query, args...)
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

// Exec executes a query within a transaction
func (t *PostgresTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// QueryRow executes a query that returns a single row within a transaction
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
