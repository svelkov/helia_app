package middleware

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"helia/internal/infrastructure/db"
)

type rowsResult int64

func (r rowsResult) LastInsertId() (int64, error) { return 0, nil }
func (r rowsResult) RowsAffected() (int64, error) { return int64(r), nil }

type execCall struct {
	query string
	args  []any
}

// recordExec returns an ExecContext func that records every statement and affects one row.
func recordExec(calls *[]execCall) func(context.Context, string, ...interface{}) (sql.Result, error) {
	return func(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
		*calls = append(*calls, execCall{q, a})
		return rowsResult(1), nil
	}
}

func TestLockServiceSQL(t *testing.T) {
	tests := []struct {
		name       string
		run        func(*LockService) error
		wantSQL    string
		wantArgs   func(args []any) bool
		wantCommit bool
	}{
		{
			name:    "unlock keeps every condition",
			run:     func(s *LockService) error { return s.Unlock(context.Background(), "fnal", 7, "ana") },
			wantSQL: "DELETE FROM entitylocks WHERE (entitytype = $1) AND (entityid = $2) AND (lockedby = $3)",
			wantArgs: func(a []any) bool {
				return reflect.DeepEqual(a, []any{"fnal", int64(7), "ana"})
			},
		},
		{
			name:    "unlock with empty user still filters by lockedby",
			run:     func(s *LockService) error { return s.Unlock(context.Background(), "fnal", 7, "") },
			wantSQL: "DELETE FROM entitylocks WHERE (entitytype = $1) AND (entityid = $2) AND (lockedby = $3)",
			wantArgs: func(a []any) bool {
				return reflect.DeepEqual(a, []any{"fnal", int64(7), ""})
			},
		},
		{
			name:     "cleanup expired",
			run:      func(s *LockService) error { return s.CleanupExpired(context.Background()) },
			wantSQL:  "DELETE FROM entitylocks WHERE (expiresat < NOW())",
			wantArgs: func(a []any) bool { return len(a) == 0 },
		},
		{
			name: "lock inserts a new lock",
			run:  func(s *LockService) error { return s.Lock(context.Background(), "fnal", 7, "ana") },
			wantSQL: "INSERT INTO entitylocks (entitytype, entityid, lockedby, expiresat) VALUES ($1, $2, $3, $4)\n" +
				"\t\tON CONFLICT (entitytype, entityid) DO NOTHING",
			wantArgs: func(a []any) bool {
				if len(a) != 4 || !reflect.DeepEqual(a[:3], []any{"fnal", int64(7), "ana"}) {
					return false
				}
				_, ok := a[3].(time.Time)
				return ok
			},
			wantCommit: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []execCall
			committed := false
			tx := &db.MockTransaction{
				ExecContextFunc: recordExec(&calls),
				CommitFunc:      func() error { committed = true; return nil },
				RollbackFunc:    func() error { return nil },
			}
			mock := &db.MockDatabase{
				ExecContextFunc: recordExec(&calls),
				BeginxFunc:      func() (db.Transaction, error) { return tx, nil },
			}
			if err := tt.run(NewLockService(mock)); err != nil {
				t.Fatal(err)
			}
			if len(calls) != 1 {
				t.Fatalf("got %d statements, want 1", len(calls))
			}
			if calls[0].query != tt.wantSQL {
				t.Errorf("sql\n got: %q\nwant: %q", calls[0].query, tt.wantSQL)
			}
			if !tt.wantArgs(calls[0].args) {
				t.Errorf("unexpected args %v", calls[0].args)
			}
			if committed != tt.wantCommit {
				t.Errorf("committed = %v, want %v", committed, tt.wantCommit)
			}
		})
	}
}
