# Database rules

PostgreSQL through `github.com/jmoiron/sqlx` and `github.com/lib/pq`, behind the interfaces of
`internal/infrastructure/db` (`Database`, `Transaction`). Every statement is built with
`helia/internal/infrastructure/db/querybuilder`.

## Where SQL may live
- Statements run only in `internal/repository` and `internal/infrastructure/db`.
- Services, handlers and `common` never call `DB.Get/Select/Query/QueryRow/Exec` or `tx.QueryRowContext/ExecContext`,
  never build SQL strings (`fmt.Sprintf("SELECT ...")`), and never import `database/sql` for queries (types such as
  `sql.NullTime` are fine).
- A service may build a read query with `querybuilder` (static fragments) and run it through its repository:
  `repo.Select(ctx, b)` (builder or union), `repo.Count(ctx, b)`, `repo.Exec(ctx, b)`, `repo.ExecWithTx(ctx, tx, b)`.
  Never pass built SQL strings around.
- Everything else goes into a repository method: statements in a transaction, single-row or scalar reads, writes
  beyond plain CRUD, statements that join several modules.

## Repositories
- `BaseRepository[T]` (`NewBaseRepository[T](db, table)`): CRUD of one entity: `GetByID`, `GetAll`,
  `GetTotalRecords`, `Create`, `Update`, `Delete`, the `*WithTx` forms, and `Select` / `Count` / `Exec` /
  `ApplyFilters` for custom queries. It panics at startup on an invalid table name or db tag
  (`TestStartupWiring` in the root package catches that).
- `RobnoRepository` (`robno_*_repository.go`, one file per table or concern) and `FinansijskoRepository`
  (`finansijsko_repository.go`): the statements of the modules. Add a new statement as a method there, next to the
  ones on the same table. Use the helpers: `inPeriod(ctx, b, prefix)` for the tenant, `row` (single row, a missing
  row wraps `sql.ErrNoRows`), `exec` (rows affected), `execOne` (ErrNotFound when no row matched, for updates by
  id), `selectRows`. A method that runs in a transaction takes `tx db.Transaction` (nil = the database).
- `LockRepository` (entitylocks), `user_repository.go` (appusers), `fvr.go` (`GetFvrData` and the lists of the
  choice of the period).
- Methods take `ctx` first, wrap errors with context and `%w`, and check `errors.Is(err, sql.ErrNoRows)`
  (never `==`).

## Using querybuilder
- Placeholders in raw fragments are always `?`, never `$1`. Numbering happens in `Build()`.
- `Build()` and `BuildCount()` return `(sql, args, error)`. Always check the error.
- The base query passed to `querybuilder.New` has no top-level `WHERE`; the builder adds it (subqueries inside the
  base may have their own).
- Values are always arguments. Never concatenate a value into SQL.
- Raw fragments (`Where`, `AddJoin`, `Having`, `Set`, `From`, `WithCTE`, `OrderByExpr`, `GroupByExpr`) are static SQL
  written in code. Request input never goes into them.
- Identifiers chosen at runtime (a table, a column) come from code or an allow-list and are quoted with
  `querybuilder.QuoteIdent`; names read from the catalog are checked with `querybuilder.IsIdent` first (see
  `StanjeDodaj` / `StanjeIzmeni`, `RproPostaviPartnera`, `RsifKreirajNulu`).
- Sorting from a request: `OrderByKey(key, dir, allowedMap)`. Never pass request input to `OrderBy`,
  `OrderByExpr`, `GroupBy`, `GroupByExpr`, column names or operators.
- Static ORDER BY / GROUP BY: plain columns with `OrderBy` / `GroupBy`; expressions (`::numeric`, `COLLATE`,
  functions, `nulls last`) with `OrderByExpr` / `GroupByExpr`, values as `?` arguments.
- OR logic: `WhereAny(querybuilder.C(...), querybuilder.C(...))`. It is grouped, so it can not bypass the tenant
  filter.
- `AddEqual`, `AddLike`, `AddIn`, `AddCondition` skip empty values on purpose (optional filters). For ids and
  tenant conditions use `Where("id = ?", id)`, which is never skipped. Never use `AddEqual` for an id. A list of
  ids that must not be empty: build `in (?, ?, ...)` with `querybuilder.Args(list)` and return an error for an
  empty list.
- Count for a page query: `b.BuildCount()` (or `repo.Count(ctx, b)`) on the same builder. It does not modify the
  builder.
- A builder used as a CTE of another query: `outer.WithCTEFrom(name, sub)`. Never copy a built SQL string and its
  args into another builder by hand.
- `UPDATE ... FROM (subquery)`: `NewUpdate(table).Set(...).From("(...) x", args...)`. `INSERT ... RETURNING`:
  `.Returning("id")`. `SELECT ... FOR UPDATE`: `.ForUpdate()` (last).
- JSON operators containing `?` are written `??`.

## Tenant (god / kar)
- Every statement on a table with `god` / `kar` is restricted to the session's god/kar, joins and subqueries
  included (e.g. partneri joined by šifra needs `p.god = x.god AND p.kar = x.kar`; joins by a key such as
  `idpartneri` do not).
- `BaseRepository` does it for GetByID, GetAll, GetTotalRecords, Update and Delete, and Create fills god/kar from
  the session. Custom queries: `repo.ApplyFilters(ctx, b, alias, fields, search)`; repository methods: `inPeriod`.
- The session is in the request context: handlers pass `c.Request.Context()` to services, never the
  `*gin.Context` (the session is not visible through it, and the repository returns `ErrNoSession`).
- `WithoutTenantScope()` is allowed only for repositories that truly cross tenants. Every use needs a comment that
  says why, and the user's approval.
- Approved exceptions: the nalog functions `fn_kopiraj_*` / `fn_storniraj_*` (primary keys, no tenant check),
  `FvrFirme` / `FvrGodine` / `FvrKarovi` (read before a period is chosen), `RdokIzvorniGodine` and
  `DokvrstaAvansnaGodine` (an explicit year with the session's kar), `FizvdetBrisi` (the period of the izvod
  header). Tables without a period: valute, porkat, rpor, initable (kar only), entitylocks, appusers.

## Repository behavior to keep
- `Update`, `Delete`, their `*WithTx` forms and `execOne` return `repository.ErrNotFound` when no row matched
  (e.g. an id of another period). Handle it; do not ignore it. Treat it as success only where "nothing to delete"
  is normal, and say so in a comment.
- Wrap errors with `%w`. Return errors instead of empty queries. Never discard an error with `_ =` (except
  `_ = tx.Rollback()` in a deferred cleanup).

## Tests
- Every repository method and every change of the builder or the SQL generation has a table-driven test that
  asserts the exact SQL string and args. Never connect to a database in a test.
- Repository tests use the fake driver of `internal/repository/fakedb_test.go` (`newDrvDB`, `drvResult`, `drvRow`,
  `squash`). Methods of `RobnoRepository` and `FinansijskoRepository` get a `goldenCase` in the list of their
  test file; run it with `-update`, review the diff of `testdata/*.golden`, then run it without `-update`.
- Run `go test -count=1 ./internal/infrastructure/db/... ./internal/repository/...`.
- Checking statements against the local database is allowed only when the user asks, and only read-only:
  `HELIA_EXPLAIN_DB=1 go test -count=1 -run TestRealDBExplain ./internal/repository/` (EXPLAIN in a read-only
  transaction). Never run writes, `EXPLAIN ANALYZE`, or the `fn_kopiraj_*` / `fn_storniraj_*` functions.
