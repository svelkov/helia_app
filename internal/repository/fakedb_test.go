package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	"helia/internal/infrastructure/db"
)

// The fake driver records every statement with its arguments and answers it with a scripted
// result, behind the real database/sql + sqlx + db.PostgresDB stack (so QueryRow, Scan,
// transactions and RowsAffected work as in production). It never talks to a database.

type drvCall struct {
	query string
	args  []any
	inTx  bool
}

// drvResult answers one statement: rows for a query, affected for an exec, or err.
type drvResult struct {
	cols     []string
	rows     [][]any
	affected int64
	err      error
}

type drvDB struct {
	mu        sync.Mutex
	calls     []drvCall
	respond   func(query string, args []any) drvResult
	commits   int
	rollbacks int
}

// record records a statement; inTx tells whether it ran in a transaction of its connection.
func (f *drvDB) record(q string, nv []driver.NamedValue, inTx bool) drvResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	args := make([]any, len(nv))
	for i, v := range nv {
		args[i] = v.Value
	}
	f.calls = append(f.calls, drvCall{query: q, args: args, inTx: inTx})
	if f.respond == nil {
		return drvResult{}
	}
	return f.respond(q, args)
}

// newDrvDB returns a db.Database backed by the fake driver.
func newDrvDB(t *testing.T, respond func(query string, args []any) drvResult) (db.Database, *drvDB) {
	t.Helper()
	f := &drvDB{respond: respond}
	sqldb := sql.OpenDB(drvConnector{f})
	t.Cleanup(func() { _ = sqldb.Close() })
	return db.NewPostgresDB(sqlx.NewDb(sqldb, "postgres")), f
}

// drvRow answers with a single row.
func drvRow(cols []string, vals ...any) drvResult {
	return drvResult{cols: cols, rows: [][]any{vals}}
}

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

type drvConnector struct{ f *drvDB }

func (c drvConnector) Connect(context.Context) (driver.Conn, error) { return &drvConn{f: c.f}, nil }
func (c drvConnector) Driver() driver.Driver                        { return drvDriver{} }

type drvDriver struct{}

func (drvDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use the connector") }

// drvConn is one connection; inTx is set while a transaction is open on it.
type drvConn struct {
	f    *drvDB
	inTx bool
}

func (c *drvConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}
func (c *drvConn) Close() error { return nil }
func (c *drvConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *drvConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.inTx = true
	return &drvTx{c}, nil
}

func (c *drvConn) QueryContext(_ context.Context, q string, nv []driver.NamedValue) (driver.Rows, error) {
	r := c.f.record(q, nv, c.inTx)
	if r.err != nil {
		return nil, r.err
	}
	return &drvRows{cols: r.cols, rows: r.rows}, nil
}

func (c *drvConn) ExecContext(_ context.Context, q string, nv []driver.NamedValue) (driver.Result, error) {
	r := c.f.record(q, nv, c.inTx)
	if r.err != nil {
		return nil, r.err
	}
	return driver.RowsAffected(r.affected), nil
}

// CheckNamedValue accepts every value as is (time.Time, []byte, ...), like the pq driver does.
func (c *drvConn) CheckNamedValue(*driver.NamedValue) error { return nil }

type drvTx struct{ c *drvConn }

func (t *drvTx) Commit() error {
	t.c.inTx = false
	t.c.f.mu.Lock()
	defer t.c.f.mu.Unlock()
	t.c.f.commits++
	return nil
}

func (t *drvTx) Rollback() error {
	t.c.inTx = false
	t.c.f.mu.Lock()
	defer t.c.f.mu.Unlock()
	t.c.f.rollbacks++
	return nil
}

type drvRows struct {
	cols []string
	rows [][]any
	i    int
}

func (r *drvRows) Columns() []string { return r.cols }
func (r *drvRows) Close() error      { return nil }
func (r *drvRows) Next(dest []driver.Value) error {
	if r.i >= len(r.rows) {
		return io.EOF
	}
	for j, v := range r.rows[r.i] {
		dest[j] = v
	}
	r.i++
	return nil
}
