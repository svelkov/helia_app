package repository

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// ---------------------------------------------------------------- fakes

type call struct {
	kind  string
	query string
	args  []any
}

type fakeResult struct{ n int64 }

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.n, nil }

type fakeDB struct {
	calls []call
	rows  int64
	tx    *fakeTx
}

func (f *fakeDB) rec(kind, q string, a []any) { f.calls = append(f.calls, call{kind, q, a}) }
func (f *fakeDB) last() call                  { return f.calls[len(f.calls)-1] }

func (f *fakeDB) GetContext(_ context.Context, _ interface{}, q string, a ...interface{}) error {
	f.rec("get", q, a)
	return nil
}
func (f *fakeDB) SelectContext(_ context.Context, _ interface{}, q string, a ...interface{}) error {
	f.rec("select", q, a)
	return nil
}
func (f *fakeDB) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (f *fakeDB) QueryRowContext(context.Context, string, ...interface{}) *sql.Row { return nil }
func (f *fakeDB) ExecContext(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
	f.rec("exec", q, a)
	return fakeResult{f.rows}, nil
}
func (f *fakeDB) Beginx() (db.Transaction, error) {
	f.tx = &fakeTx{f: f}
	return f.tx, nil
}
func (f *fakeDB) Ping() error  { return nil }
func (f *fakeDB) Close() error { return nil }

type fakeTx struct {
	f                     *fakeDB
	committed, rolledBack bool
}

func (t *fakeTx) ExecContext(_ context.Context, q string, a ...interface{}) (sql.Result, error) {
	t.f.rec("tx-exec", q, a)
	return fakeResult{t.f.rows}, nil
}
func (t *fakeTx) QueryRowContext(context.Context, string, ...interface{}) *sql.Row { return nil }
func (t *fakeTx) Commit() error                                                    { t.committed = true; return nil }
func (t *fakeTx) Rollback() error                                                  { t.rolledBack = true; return nil }

// ---------------------------------------------------------------- entities

type tenantCols struct {
	God int `db:"god"`
	Kar int `db:"kar"`
}

type partner struct {
	tenantCols
	ID      int       `db:"id"`
	Name    string    `db:"name"`
	Note    *string   `db:"note"`
	Active  bool      `db:"active"`
	Created time.Time `db:"created"`
	Hidden  string    `db:"hidden" addupdate:"false"`
	Skip    string    `db:"-"`
}

type plain struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

var fixed = time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC)

func setup(t *testing.T, session bool) (*BaseRepository[partner], *fakeDB, context.Context) {
	t.Helper()
	now = func() time.Time { return fixed }
	sessionFromContext = func(context.Context) (entitySession, error) {
		if !session {
			return entitySession{}, ErrNoSession
		}
		return entitySession{UserName: "ana", God: 2026, Kar: 7}, nil
	}
	t.Cleanup(func() { now = time.Now; sessionFromContext = defaultSessionFromContext })
	f := &fakeDB{rows: 1}
	return NewBaseRepository[partner](f, "partneri"), f, context.Background()
}

var defaultSessionFromContext = sessionFromContext

func eq(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s\n got: %s\nwant: %s", name, got, want)
	}
}

func eqArgs(t *testing.T, got, want []any) {
	t.Helper()
	if len(got) != len(want) || (len(got) > 0 && !reflect.DeepEqual(got, want)) {
		t.Errorf("args got %v want %v", got, want)
	}
}

var fields = []domain.Fields{
	{Name: "Name"}, {Name: "Note"}, {Name: "Active", SkipInSearch: true}, {Name: "Created"},
}

// ---------------------------------------------------------------- tests

func TestNewBaseRepositoryValidates(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("invalid table name must panic at startup")
		}
	}()
	NewBaseRepository[plain](&fakeDB{}, "bad name; drop")
}

func TestNewBaseRepositoryBadTag(t *testing.T) {
	type bad struct {
		A string `db:"a;b"`
	}
	defer func() {
		if recover() == nil {
			t.Error("invalid db tag must panic at startup")
		}
	}()
	NewBaseRepository[bad](&fakeDB{}, "t")
}

func TestStructLiteralRepositoryWorks(t *testing.T) {
	_, f, ctx := setup(t, true)
	r := &BaseRepository[plain]{DB: f, TableName: "things"} // not created via NewBaseRepository
	if _, err := r.GetByID(ctx, "id", 1); err != nil {
		t.Fatal(err)
	}
	eq(t, "sql", f.last().query, `SELECT COALESCE("id", 0) AS "id", COALESCE("name", '') AS "name" FROM "things" WHERE ("id" = $1)`)
}

func TestGetByIDTenantScoped(t *testing.T) {
	r, f, ctx := setup(t, true)
	if _, err := r.GetByID(ctx, "id", 5); err != nil {
		t.Fatal(err)
	}
	eq(t, "sql", f.last().query, `SELECT COALESCE("god", 0) AS "god", COALESCE("kar", 0) AS "kar", COALESCE("id", 0) AS "id", `+
		`COALESCE("name", '') AS "name", "note", COALESCE("active", false) AS "active", "created" `+
		`FROM "partneri" WHERE ("id" = $1) AND (god = $2) AND (kar = $3)`)
	eqArgs(t, f.last().args, []any{5, 2026, 7})
}

func TestGetByIDEmptyIDIsNeverDropped(t *testing.T) {
	r, f, ctx := setup(t, true)
	if _, err := r.GetByID(ctx, "id", ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.last().query, `("id" = $1)`) {
		t.Errorf("empty id must stay a condition, got %s", f.last().query)
	}
	eqArgs(t, f.last().args, []any{"", 2026, 7})
}

func TestTenantScopeNeedsSessionButNotForPlainEntities(t *testing.T) {
	r, _, ctx := setup(t, false)
	if _, err := r.GetByID(ctx, "id", 1); !errors.Is(err, ErrNoSession) {
		t.Errorf("tenant entity without session must fail, got %v", err)
	}
	if err := r.Delete(ctx, "id", 1); !errors.Is(err, ErrNoSession) {
		t.Errorf("delete without session must fail, got %v", err)
	}
	if _, err := r.WithoutTenantScope().GetByID(ctx, "id", 1); err != nil {
		t.Errorf("opt-out must work without session: %v", err)
	}
	p := NewBaseRepository[plain](&fakeDB{}, "things")
	if _, err := p.GetByID(ctx, "id", 1); err != nil {
		t.Errorf("entity without god/kar needs no session: %v", err)
	}
}

func TestWithoutTenantScopeDoesNotChangeTheShared(t *testing.T) {
	r, f, ctx := setup(t, true)
	tests := []struct {
		name      string
		repo      *BaseRepository[partner]
		wantWhere string
		wantArgs  []any
	}{
		{"opt-out copy is not scoped", r.WithoutTenantScope(), `WHERE ("id" = $1)`, []any{5}},
		{"original stays scoped", r, `WHERE ("id" = $1) AND (god = $2) AND (kar = $3)`, []any{5, 2026, 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.repo.GetByID(ctx, "id", 5); err != nil {
				t.Fatal(err)
			}
			if !strings.HasSuffix(f.last().query, tt.wantWhere) {
				t.Errorf("got %s\nwant suffix %s", f.last().query, tt.wantWhere)
			}
			eqArgs(t, f.last().args, tt.wantArgs)
		})
	}
}

func TestValueCopySharesMetadata(t *testing.T) {
	r, f, ctx := setup(t, true)
	v := *r // services hold repositories by value
	if v.table != r.table {
		t.Error("a value copy must share the entity metadata")
	}
	if _, err := v.GetByID(ctx, "id", 5); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(f.last().query, `AND (god = $2) AND (kar = $3)`) {
		t.Errorf("a value copy must stay tenant-scoped, got %s", f.last().query)
	}
}

func TestGetAll(t *testing.T) {
	r, f, ctx := setup(t, true)
	tests := []struct {
		name, sortBy, sortOrder, wantOrder string
	}{
		{"column asc", "Name", "asc", "ORDER BY name ASC"},
		{"column desc", "name", "DESC", "ORDER BY name DESC"},
		{"unknown column falls back to id", "name; DROP TABLE partneri", "desc", "ORDER BY id DESC"},
		{"empty sort", "", "", "ORDER BY id ASC"},
		{"order direction is normalized", "name", "asc; DROP", "ORDER BY name ASC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := r.GetAll(ctx, 10, 20, fields, "id", "acme", tt.sortBy, tt.sortOrder); err != nil {
				t.Fatal(err)
			}
			want := `SELECT "id", "god", "kar", "name", "note", "active", "created" FROM "partneri" ` +
				`WHERE ("partneri".god = $1) AND ("partneri".kar = $2) AND (name::text ILIKE $3 OR note::text ILIKE $4) ` +
				tt.wantOrder + ` LIMIT 10 OFFSET 20`
			eq(t, "sql", f.last().query, want)
			eqArgs(t, f.last().args, []any{2026, 7, "%acme%", "%acme%"})
		})
	}
}

func TestGetAllInvalidIDField(t *testing.T) {
	r, _, ctx := setup(t, true)
	if _, err := r.GetAll(ctx, 10, 0, fields, "id; drop", "", "", ""); err == nil {
		t.Error("invalid id column must fail")
	}
}

func TestGetTotalRecords(t *testing.T) {
	r, f, ctx := setup(t, true)
	if _, err := r.GetTotalRecords(ctx, fields, "acme"); err != nil {
		t.Fatal(err)
	}
	eq(t, "sql", f.last().query, `SELECT count(*) FROM "partneri" WHERE ("partneri".god = $1) AND ("partneri".kar = $2) `+
		`AND (name::text ILIKE $3 OR note::text ILIKE $4)`)
	eqArgs(t, f.last().args, []any{2026, 7, "%acme%", "%acme%"})
}

func TestInsertStatement(t *testing.T) {
	r, _, ctx := setup(t, true)
	q, args, err := r.CreateInsertStatement(ctx, nil, []domain.Fields{
		{Name: "Name", Value: "Acme"}, {Name: "note", IsNull: true, Value: "ignored"},
		{Name: "god", Value: "1"}, {Name: "XDATUNOSA", Value: "x"}, // reserved: ignored
	}, "id")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "sql", q, `INSERT INTO "partneri" ("god", "kar", "xopunos", "xdatunosa", "name", "note") VALUES ($1, $2, $3, $4, $5, $6) RETURNING "id"`)
	eqArgs(t, args, []any{2026, 7, "ana", fixed, "Acme", nil})

	rNo, _, ctxNo := setup(t, false)
	if _, _, err := rNo.CreateInsertStatement(ctxNo, nil, nil, "id"); !errors.Is(err, ErrNoSession) {
		t.Errorf("insert without session must fail, got %v", err)
	}
	if _, _, err := r.CreateInsertStatement(ctx, nil, []domain.Fields{{Name: "a;b"}}, "id"); err == nil {
		t.Error("bad column must fail")
	}
}

func TestUpdate(t *testing.T) {
	r, f, ctx := setup(t, true)
	if err := r.Update(ctx, nil, "id", 42, []domain.Fields{{Name: "Name", Value: "New"}, {Name: "xopizmene", Value: "evil"}}); err != nil {
		t.Fatal(err)
	}
	c := f.last()
	eq(t, "sql", c.query, `UPDATE "partneri" SET "name" = $1, "xopizmene" = $2, "xdatizmene" = $3 WHERE ("id" = $4) AND (god = $5) AND (kar = $6)`)
	eqArgs(t, c.args, []any{"New", "ana", fixed, 42, 2026, 7})
	if !f.tx.committed {
		t.Error("update must commit")
	}
}

func TestUpdateAndDeleteReturnNotFoundWhenNothingMatched(t *testing.T) {
	r, f, ctx := setup(t, true)
	f.rows = 0 // e.g. the id belongs to another tenant
	err := r.Update(ctx, nil, "id", 42, []domain.Fields{{Name: "Name", Value: "x"}})
	if !errors.Is(err, ErrNotFound) || !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("update: got %v", err)
	}
	if f.tx.committed {
		t.Error("must not commit when nothing matched")
	}
	if err := r.Delete(ctx, "id", 9); !errors.Is(err, ErrNotFound) {
		t.Errorf("delete: got %v", err)
	}
	tx, _ := f.Beginx()
	if err := r.DeleteWithTx(ctx, tx, "id", 9); !errors.Is(err, ErrNotFound) {
		t.Errorf("delete with tx: got %v", err)
	}
	if err := r.UpdateWithTx(ctx, tx, nil, "id", 9, nil); !errors.Is(err, ErrNotFound) {
		t.Errorf("update with tx: got %v", err)
	}
}

func TestDelete(t *testing.T) {
	r, f, ctx := setup(t, true)
	if err := r.Delete(ctx, "id", 9); err != nil {
		t.Fatal(err)
	}
	eq(t, "sql", f.last().query, `DELETE FROM "partneri" WHERE ("id" = $1) AND (god = $2) AND (kar = $3)`)
	eqArgs(t, f.last().args, []any{int64(9), 2026, 7})
	if _, _, err := r.CreateDeleteStatement(ctx, "id; x", 1); err == nil {
		t.Error("bad id column must fail")
	}
}

func TestApplyFiltersForCustomQueries(t *testing.T) {
	r, _, ctx := setup(t, true)
	b := queryFor("SELECT p.id FROM partneri p JOIN grad g ON g.id = p.grad_id")
	if err := r.ApplyFilters(ctx, b, "p", fields, "acme"); err != nil {
		t.Fatal(err)
	}
	b.OrderBy("p.id").Limit(5)
	q, args, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "sql", q, `SELECT p.id FROM partneri p JOIN grad g ON g.id = p.grad_id WHERE (p.god = $1) AND (p.kar = $2) `+
		`AND (name::text ILIKE $3 OR note::text ILIKE $4) ORDER BY p.id LIMIT 5`)
	eqArgs(t, args, []any{2026, 7, "%acme%", "%acme%"})
	if err := r.ApplyFilters(ctx, queryFor("SELECT 1"), "p; x", nil, ""); err == nil {
		t.Error("bad alias must fail")
	}
}

func TestBuilderMethods(t *testing.T) {
	page := func() *querybuilder.Builder {
		return querybuilder.New("SELECT p.id FROM partneri p").Where("p.id > ?", 5).OrderBy("p.id").Limit(10)
	}
	tests := []struct {
		name     string
		run      func(r *BaseRepository[partner], tx db.Transaction, ctx context.Context) error
		wantKind string
		wantSQL  string
		wantArgs []any
	}{
		{"select builder", func(r *BaseRepository[partner], _ db.Transaction, ctx context.Context) error {
			_, err := r.Select(ctx, page())
			return err
		}, "select", "SELECT p.id FROM partneri p WHERE (p.id > $1) ORDER BY p.id LIMIT 10", []any{5}},
		{"select union", func(r *BaseRepository[partner], _ db.Transaction, ctx context.Context) error {
			u := querybuilder.NewUnionAll().Add(querybuilder.New("SELECT id FROM a").Where("x = ?", 1)).
				Add(querybuilder.New("SELECT id FROM b").Where("x = ?", 2)).OrderBy("id")
			_, err := r.Select(ctx, u)
			return err
		}, "select", "(SELECT id FROM a WHERE (x = $1)) UNION ALL (SELECT id FROM b WHERE (x = $2)) ORDER BY id", []any{1, 2}},
		{"count ignores order and limit", func(r *BaseRepository[partner], _ db.Transaction, ctx context.Context) error {
			_, err := r.Count(ctx, page())
			return err
		}, "get", "SELECT count(*) FROM (SELECT p.id FROM partneri p WHERE (p.id > $1)) AS cnt", []any{5}},
		{"exec", func(r *BaseRepository[partner], _ db.Transaction, ctx context.Context) error {
			return r.Exec(ctx, querybuilder.NewUpdate("partneri").Set("naziv = ?", "x").Where("id = ?", 3))
		}, "exec", `UPDATE "partneri" SET naziv = $1 WHERE (id = $2)`, []any{"x", 3}},
		{"exec with tx", func(r *BaseRepository[partner], tx db.Transaction, ctx context.Context) error {
			return r.ExecWithTx(ctx, tx, querybuilder.New("DELETE FROM partneri").Where("id = ?", 3))
		}, "tx-exec", "DELETE FROM partneri WHERE (id = $1)", []any{3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, f, ctx := setup(t, true)
			f.rows = 0 // Exec must not require a matched row
			tx, _ := f.Beginx()
			if err := tt.run(r, tx, ctx); err != nil {
				t.Fatal(err)
			}
			got := f.last()
			if got.kind != tt.wantKind {
				t.Errorf("kind %q, want %q", got.kind, tt.wantKind)
			}
			eq(t, "sql", got.query, tt.wantSQL)
			eqArgs(t, got.args, tt.wantArgs)
		})
	}
}

func TestBuilderMethodsReturnBuildErrors(t *testing.T) {
	r, f, ctx := setup(t, true)
	bad := querybuilder.New("SELECT 1").Where("a = ?") // missing argument
	tx, _ := f.Beginx()
	checks := map[string]error{}
	_, checks["select"] = r.Select(ctx, bad)
	_, checks["count"] = r.Count(ctx, bad)
	checks["exec"] = r.Exec(ctx, bad)
	checks["exec with tx"] = r.ExecWithTx(ctx, tx, bad)
	for name, err := range checks {
		if err == nil || !strings.Contains(err.Error(), "placeholders") {
			t.Errorf("%s: want build error, got %v", name, err)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing may reach the database, got %d calls", len(f.calls))
	}
}

func TestHasGodKar(t *testing.T) {
	r, _, _ := setup(t, true)
	if g, k := r.GetHasGodHasKar(); !g || !k {
		t.Errorf("god/kar in embedded struct not detected: %v %v", g, k)
	}
	if g, k := NewBaseRepository[plain](&fakeDB{}, "t").GetHasGodHasKar(); g || k {
		t.Error("plain entity has no god/kar")
	}
}
