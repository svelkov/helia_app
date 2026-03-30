package db

import (
	"context"
	"database/sql"
)

// MockDatabase is a mock implementation of the Database interface for testing
type MockDatabase struct {
	GetContextFunc      func(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetFunc             func(dest interface{}, query string, args ...interface{}) error
	SelectContextFunc   func(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectFunc          func(dest interface{}, query string, args ...interface{}) error
	QueryRowContextFunc func(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRowFunc        func(query string, args ...interface{}) *sql.Row
	QueryContextFunc    func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContextFunc     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	BeginxFunc          func() (Transaction, error)
	PingFunc            func() error
	CloseFunc           func() error
}

func (m *MockDatabase) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if m.GetContextFunc != nil {
		return m.GetContextFunc(ctx, dest, query, args...)
	}
	return nil
}

func (m *MockDatabase) Get(dest interface{}, query string, args ...interface{}) error {
	if m.GetFunc != nil {
		return m.GetFunc(dest, query, args...)
	}
	return nil
}

func (m *MockDatabase) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if m.SelectContextFunc != nil {
		return m.SelectContextFunc(ctx, dest, query, args...)
	}
	return nil
}

func (m *MockDatabase) Select(dest interface{}, query string, args ...interface{}) error {
	if m.SelectFunc != nil {
		return m.SelectFunc(dest, query, args...)
	}
	return nil
}

func (m *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.QueryRowContextFunc != nil {
		return m.QueryRowContextFunc(ctx, query, args...)
	}
	return nil
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(query, args...)
	}
	return nil
}

func (m *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.QueryContextFunc != nil {
		return m.QueryContextFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.ExecContextFunc != nil {
		return m.ExecContextFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockDatabase) Beginx() (Transaction, error) {
	if m.BeginxFunc != nil {
		return m.BeginxFunc()
	}
	return &MockTransaction{}, nil
}

func (m *MockDatabase) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}

func (m *MockDatabase) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockTransaction is a mock implementation of the Transaction interface for testing
type MockTransaction struct {
	ExecContextFunc     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	ExecFunc            func(query string, args ...interface{}) (sql.Result, error)
	QueryRowContextFunc func(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRowFunc        func(query string, args ...interface{}) *sql.Row
	CommitFunc          func() error
	RollbackFunc        func() error
}

func (m *MockTransaction) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.ExecContextFunc != nil {
		return m.ExecContextFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(query, args...)
	}
	return nil, nil
}

func (m *MockTransaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.QueryRowContextFunc != nil {
		return m.QueryRowContextFunc(ctx, query, args...)
	}
	return nil
}

func (m *MockTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(query, args...)
	}
	return nil
}

func (m *MockTransaction) Commit() error {
	if m.CommitFunc != nil {
		return m.CommitFunc()
	}
	return nil
}

func (m *MockTransaction) Rollback() error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc()
	}
	return nil
}
