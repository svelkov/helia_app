package repository

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestLockRepositorySQL(t *testing.T) {
	exp := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	ctx := context.Background()
	tests := []struct {
		name     string
		respond  drvResult
		run      func(r *LockRepository) error
		wantSQL  string
		wantArgs []any
	}{
		{
			name:    "insert",
			respond: drvResult{affected: 1},
			run: func(r *LockRepository) error {
				tx, _ := r.BeginTx()
				ok, err := r.InsertTx(ctx, tx, "fnal", 7, "ana", exp)
				if err == nil && !ok {
					t.Error("one affected row means inserted")
				}
				return err
			},
			wantSQL:  "INSERT INTO entitylocks (entitytype, entityid, lockedby, expiresat) VALUES ($1, $2, $3, $4) ON CONFLICT (entitytype, entityid) DO NOTHING",
			wantArgs: []any{"fnal", int64(7), "ana", exp},
		},
		{
			name:    "holder for update",
			respond: drvRow([]string{"lockedby", "expiresat"}, "ana", exp),
			run: func(r *LockRepository) error {
				tx, _ := r.BeginTx()
				who, until, found, err := r.HolderTx(ctx, tx, "fnal", 7, true)
				if err == nil && (!found || who != "ana" || !until.Equal(exp)) {
					t.Errorf("got %q %v %v", who, until, found)
				}
				return err
			},
			wantSQL:  "SELECT lockedby, expiresat FROM entitylocks WHERE (entitytype = $1) AND (entityid = $2) FOR UPDATE",
			wantArgs: []any{"fnal", int64(7)},
		},
		{
			name:    "holder not found",
			respond: drvResult{cols: []string{"lockedby", "expiresat"}},
			run: func(r *LockRepository) error {
				tx, _ := r.BeginTx()
				_, _, found, err := r.HolderTx(ctx, tx, "fnal", 7, false)
				if found {
					t.Error("no row must mean not found")
				}
				return err
			},
			wantSQL:  "SELECT lockedby, expiresat FROM entitylocks WHERE (entitytype = $1) AND (entityid = $2)",
			wantArgs: []any{"fnal", int64(7)},
		},
		{
			name: "delete only expired",
			run: func(r *LockRepository) error {
				tx, _ := r.BeginTx()
				return r.DeleteTx(ctx, tx, "fnal", 7, true)
			},
			wantSQL:  "DELETE FROM entitylocks WHERE (entitytype = $1) AND (entityid = $2) AND (expiresat < NOW())",
			wantArgs: []any{"fnal", int64(7)},
		},
		{
			name: "refresh keeps lockedby",
			run: func(r *LockRepository) error {
				tx, _ := r.BeginTx()
				return r.RefreshTx(ctx, tx, "fnal", 7, "ana", exp)
			},
			wantSQL:  `UPDATE "entitylocks" SET expiresat = $1 WHERE (entitytype = $2) AND (entityid = $3) AND (lockedby = $4)`,
			wantArgs: []any{exp, "fnal", int64(7), "ana"},
		},
		{
			name:     "release with an empty user still filters by lockedby",
			run:      func(r *LockRepository) error { return r.Release(ctx, "fnal", 7, "") },
			wantSQL:  "DELETE FROM entitylocks WHERE (entitytype = $1) AND (entityid = $2) AND (lockedby = $3)",
			wantArgs: []any{"fnal", int64(7), ""},
		},
		{
			name:     "delete expired",
			run:      func(r *LockRepository) error { return r.DeleteExpired(ctx) },
			wantSQL:  "DELETE FROM entitylocks WHERE (expiresat < NOW())",
			wantArgs: []any{},
		},
		{
			name:    "active lock",
			respond: drvRow([]string{"id", "entitytype", "entityid", "lockedby", "lockedat", "expiresat"}, int64(1), "fnal", int64(7), "ana", exp, exp),
			run: func(r *LockRepository) error {
				l, err := r.Active(ctx, "fnal", 7, exp)
				if err == nil && (l == nil || l.LockedBy != "ana") {
					t.Errorf("got %+v", l)
				}
				return err
			},
			wantSQL:  "SELECT entitylocksid AS id, entitytype, entityid, lockedby, lockedat, expiresat FROM entitylocks WHERE (entitytype = $1) AND (entityid = $2) AND (expiresat > $3)",
			wantArgs: []any{"fnal", int64(7), exp},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, f := newDrvDB(t, func(string, []any) drvResult { return tt.respond })
			if err := tt.run(NewLockRepository(d)); err != nil {
				t.Fatal(err)
			}
			if len(f.calls) != 1 {
				t.Fatalf("%d statements, want 1", len(f.calls))
			}
			if got := squash(f.calls[0].query); got != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", got, tt.wantSQL)
			}
			if !reflect.DeepEqual(f.calls[0].args, tt.wantArgs) {
				t.Errorf("args\n got: %#v\nwant: %#v", f.calls[0].args, tt.wantArgs)
			}
		})
	}
}
