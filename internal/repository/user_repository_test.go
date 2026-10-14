package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"helia/internal/domain"
)

func TestUserRepositorySQL(t *testing.T) {
	u := &domain.User{Id: 9, Username: "ana", Email: "a@x", Name: "Ana", PasswordHash: "h", TwoFAEnabled: true, TwoFASecret: "s", BackupCodes: "b"}
	tests := []struct {
		name     string
		respond  drvResult
		run      func(r *BaseRepository[domain.User]) error
		wantSQL  string
		wantArgs []any
		wantErr  error
	}{
		{
			name:    "insert returns the id",
			respond: drvRow([]string{"id"}, int64(31)),
			run: func(r *BaseRepository[domain.User]) error {
				id, err := InsertUser(context.Background(), r, u)
				if err == nil && id != 31 {
					t.Errorf("id %d", id)
				}
				return err
			},
			wantSQL:  "INSERT INTO appusers (username, email, name, password_hash, two_fa_enabled, two_fa_secret, backup_codes) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			wantArgs: []any{"ana", "a@x", "Ana", "h", true, "s", "b"},
		},
		{
			name:    "update",
			respond: drvResult{affected: 1},
			run:     func(r *BaseRepository[domain.User]) error { return UpdateUser(context.Background(), r, u) },
			wantSQL: `UPDATE "appusers" SET username = $1, email = $2, name = $3, password_hash = $4, two_fa_enabled = $5, ` +
				`two_fa_secret = $6, backup_codes = $7 WHERE (id = $8)`,
			wantArgs: []any{"ana", "a@x", "Ana", "h", true, "s", "b", int64(9)},
		},
		{
			name:    "update of a missing user is ErrNotFound",
			respond: drvResult{affected: 0},
			run:     func(r *BaseRepository[domain.User]) error { return UpdateUser(context.Background(), r, u) },
			wantSQL: `UPDATE "appusers" SET username = $1, email = $2, name = $3, password_hash = $4, two_fa_enabled = $5, ` +
				`two_fa_secret = $6, backup_codes = $7 WHERE (id = $8)`,
			wantArgs: []any{"ana", "a@x", "Ana", "h", true, "s", "b", int64(9)},
			wantErr:  ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, f := newDrvDB(t, func(string, []any) drvResult { return tt.respond })
			err := tt.run(NewBaseRepository[domain.User](d, "appusers"))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err %v, want %v", err, tt.wantErr)
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
