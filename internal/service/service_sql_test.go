package service

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

type sqlCall struct {
	query string
	args  []any
}

type okResult struct{}

func (okResult) LastInsertId() (int64, error) { return 0, nil }
func (okResult) RowsAffected() (int64, error) { return 1, nil }

// recordingDB returns a mock that records every query and returns no rows.
func recordingDB(calls *[]sqlCall) *db.MockDatabase {
	rec := func(q string, a []any) { *calls = append(*calls, sqlCall{q, a}) }
	return &db.MockDatabase{
		SelectContextFunc: func(_ context.Context, _ interface{}, q string, a ...interface{}) error {
			rec(q, a)
			return nil
		},
		GetContextFunc: func(_ context.Context, _ interface{}, q string, a ...interface{}) error {
			rec(q, a)
			return nil
		},
		ExecContextFunc: func(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
			rec(q, a)
			return okResult{}, nil
		},
		BeginxFunc: func() (db.Transaction, error) {
			return &db.MockTransaction{
				ExecContextFunc: func(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
					rec(q, a)
					return okResult{}, nil
				},
				CommitFunc:   func() error { return nil },
				RollbackFunc: func() error { return nil },
			}, nil
		},
	}
}

func sessionCtx() context.Context {
	return domain.SetSessionInStdContext(context.Background(),
		&domain.UserSession{UserName: "ana", SelectedGod: 2026, SelectedKar: 7})
}

// oneLine collapses the whitespace of a multi-line query so expectations stay readable.
func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

const partneriSelect = "SELECT sifra, naziv, adresa, pobro, mesto, pib, jmbg, bpg, index, gln, jib, " +
	"coalesce(ziro, '') as ziro, matbr, konta, tippdv, email, telefon, kontaktosb, budzetski, jbkjs, " +
	"napomena, idpartneri FROM partneri"

func TestPartneriServiceSQL(t *testing.T) {
	tests := []struct {
		name     string
		run      func(s *PartneriResource) error
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "list page sorted by a whitelisted column",
			run: func(s *PartneriResource) error {
				return s.GetAllPartneri(sessionCtx(), &domain.TableData{}, false, 2, 10, "", "Naziv", "desc")
			},
			wantSQL:  partneriSelect + " WHERE (partneri.god = $1) AND (partneri.kar = $2) ORDER BY naziv DESC LIMIT 10 OFFSET 10",
			wantArgs: []any{2026, 7},
		},
		{
			name: "unknown sort key is ignored",
			run: func(s *PartneriResource) error {
				return s.GetAllPartneri(sessionCtx(), &domain.TableData{}, false, 1, 10, "", "naziv; DROP TABLE partneri", "asc")
			},
			wantSQL:  partneriSelect + " WHERE (partneri.god = $1) AND (partneri.kar = $2) LIMIT 10",
			wantArgs: []any{2026, 7},
		},
		{
			name: "search covers every column and counts with the same filters",
			run: func(s *PartneriResource) error {
				return s.GetAllPartneri(sessionCtx(), &domain.TableData{}, true, 1, 10, "ab", "", "")
			},
			wantSQL: "SELECT count(*) FROM (" + partneriSelect + " WHERE (partneri.god = $1) AND (partneri.kar = $2) AND (" +
				"LOWER(partneri.sifra) LIKE LOWER($3) OR LOWER(partneri.naziv) LIKE LOWER($4) OR " +
				"LOWER(partneri.adresa) LIKE LOWER($5) OR LOWER(partneri.pobro::text) LIKE LOWER($6) OR " +
				"LOWER(partneri.mesto) LIKE LOWER($7) OR LOWER(partneri.pib) LIKE LOWER($8) OR " +
				"LOWER(partneri.jmbg) LIKE LOWER($9) OR LOWER(partneri.bpg) LIKE LOWER($10) OR " +
				"LOWER(partneri.index) LIKE LOWER($11))) AS cnt",
			wantArgs: []any{2026, 7, "ab", "ab", "ab", "ab", "ab", "ab", "ab", "ab", "ab"},
		},
		{
			name: "exists by field keeps an empty value as a condition",
			run: func(s *PartneriResource) error {
				_, err := s.ExistsByField(sessionCtx(), "pib", "")
				return err
			},
			wantSQL:  "SELECT sifra FROM partneri WHERE (partneri.pib = $1) AND (partneri.god = $2) AND (partneri.kar = $3) LIMIT 1",
			wantArgs: []any{"", 2026, 7},
		},
		{
			name: "tekuci racuni are scoped to the partner and the tenant",
			run: func(s *PartneriResource) error {
				return s.GetTekuciRacuni(sessionCtx(), 42, &domain.TableData{})
			},
			wantSQL: "SELECT t.tekrac, b.banka FROM tekracuni as t left join banke as b on b.idbanke = t.idbanke " +
				"WHERE (t.idpartneri = $1) AND (t.god = $2) AND (t.kar = $3)",
			wantArgs: []any{int64(42), 2026, 7},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			mock := recordingDB(&calls)
			s := &PartneriResource{
				partneriRepo:  repository.NewBaseRepository[domain.Partneri](mock, "partneri"),
				tekracuniRepo: repository.NewBaseRepository[domain.TekRacuni](mock, "tekracuni"),
			}
			if err := tt.run(s); err != nil {
				t.Fatal(err)
			}
			if len(calls) != 1 {
				t.Fatalf("got %d queries, want 1", len(calls))
			}
			if got := oneLine(calls[0].query); got != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", got, tt.wantSQL)
			}
			if !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %v want %v", calls[0].args, tt.wantArgs)
			}
		})
	}
}

func TestExistsByFieldRejectsUnknownColumns(t *testing.T) {
	var calls []sqlCall
	s := &PartneriResource{partneriRepo: repository.NewBaseRepository[domain.Partneri](recordingDB(&calls), "partneri")}
	if _, err := s.ExistsByField(sessionCtx(), "naziv OR 1=1", "x"); err == nil {
		t.Error("unknown field must be rejected")
	}
	if len(calls) != 0 {
		t.Errorf("nothing may reach the database, got %d queries", len(calls))
	}
}

func TestUserServiceSQL(t *testing.T) {
	tests := []struct {
		name     string
		run      func(s UserService) error
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "empty username is still a condition",
			run: func(s UserService) error {
				_, err := s.GetUserByUsername(context.Background(), "")
				return err
			},
			wantSQL:  "SELECT id, username, email, name, password_hash, two_fa_enabled, two_fa_secret, backup_codes FROM appusers WHERE (username = $1) LIMIT 1",
			wantArgs: []any{""},
		},
		{
			name: "update user",
			run: func(s UserService) error {
				return s.UpdateUser(context.Background(), &domain.User{Id: 9, Username: "ana"})
			},
			wantSQL: `UPDATE "appusers" SET username = $1, email = $2, name = $3, password_hash = $4, ` +
				`two_fa_enabled = $5, two_fa_secret = $6, backup_codes = $7 WHERE (id = $8)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			s := NewUserService(repository.NewBaseRepository[domain.User](recordingDB(&calls), "appusers"), nil)
			if err := tt.run(s); err != nil {
				t.Fatal(err)
			}
			if len(calls) != 1 {
				t.Fatalf("got %d queries, want 1", len(calls))
			}
			if got := oneLine(calls[0].query); got != tt.wantSQL {
				t.Errorf("sql\n got: %s\nwant: %s", got, tt.wantSQL)
			}
			if tt.wantArgs != nil && !reflect.DeepEqual(calls[0].args, tt.wantArgs) {
				t.Errorf("args got %v want %v", calls[0].args, tt.wantArgs)
			}
			if tt.wantArgs == nil && len(calls[0].args) != 8 {
				t.Errorf("want 8 args, got %d", len(calls[0].args))
			}
		})
	}
}

// TestSubmenuNameIsAKey checks that the menu name is always a condition: an empty name must not
// list the submenus of every menu.
func TestSubmenuNameIsAKey(t *testing.T) {
	var calls []sqlCall
	s := &MenuResource{subMenuRepo: repository.NewBaseRepository[domain.SubMenuItem](recordingDB(&calls), "submenuitems")}
	if _, err := s.GetSubmenuData(sessionCtx(), ""); err != nil {
		t.Fatal(err)
	}
	want := "FROM menuitems m inner join submenuitems sb on sb.menuid = m.id WHERE (m.menuname = $1) ORDER BY sortorder"
	if got := oneLine(calls[0].query); !strings.Contains(got, want) {
		t.Errorf("sql\n got: %s\nwant to contain: %s", got, want)
	}
	if !reflect.DeepEqual(calls[0].args, []any{""}) {
		t.Errorf("args %#v", calls[0].args)
	}
}
