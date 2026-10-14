# helia (Go backend)

Accounting and inventory web app (finansijsko, robno, kamate, osnovna sredstva) that replaces a legacy WinDev
program. Go + gin, server-rendered templ + htmx + Tailwind, PostgreSQL through sqlx and lib/pq.

## Layout
- `main.go`, `factory.go`: startup and wiring (`setupRouter`, `setEntities`; all repositories, services and
  handlers are built there). `startup_test.go` runs the wiring without a database.
- `internal/handler` (`finansijsko`, `robno`, `kamate`, `os`, generic handlers): gin handlers and routes.
- `internal/service` (same modules, plus `common` for combos and searches): business logic.
- `internal/repository`: the ONLY place that talks to the database. `BaseRepository[T]` for CRUD,
  `RobnoRepository` (`robno_*_repository.go`), `FinansijskoRepository`, `LockRepository`.
- `internal/infrastructure/db`: `Database` / `Transaction` interfaces; `.../db/querybuilder`: the PostgreSQL
  builder (imports only the standard library).
- `internal/domain`: entities, DTOs and the user session. Imports nothing from other internal packages.
- `internal/middleware`: JWT, session (puts the session into the request context), CSRF, entity locks.
- `internal/validation`: validation rules of the entities.
- `internal/common`: tiny shared helpers and constants only (CSS classes in `const.go`). No SQL, no builder.
- `frontend/components` (reusable templ components), `frontend/templates` (pages, dialogs, reports),
  `frontend/static` (css, js, Tailwind input `static/src/input.css`).
- `config/config.json` (port, database, languages), `i18n/translations`, `migrations/`.

## Layers and import direction
handler → service → repository → infrastructure/db → standard library
- Never import "upwards" (repository must not import service or handler; querybuilder must not import domain or
  repository).
- Handlers pass `c.Request.Context()` to services, never the `*gin.Context`: the session (firma, god, kar) is only
  in the request context, and the repository needs it for the tenant scope.

## Tenant
Every period is a firma + `god` (year) + `kar`. Tables with `god` / `kar` are always read and written in the
session's period; the repository enforces it. Details and the approved exceptions: `.claude/rules/database.md`.

## Commands
- Build: `go build ./...`
- Static check: `go vet ./...`
- Format check: the files have CRLF line endings, so `gofmt -l .` lists them all; check the changed files with
  `tr -d '\r' < file | gofmt -l` (must print nothing).
- Tests: `go test -count=1 -skip TestZzPorTmp ./...` (`TestZzPorTmp` reaches the real database).
- Templ: `templ fmt <file>` and `templ generate -f <file>` on the single changed file, from the repo root.
- CSS: `./tailwindcss.exe -i frontend/static/src/input.css -o frontend/static/css/output.css`
  (`tailwind_build.sh` runs it with `--watch`).

## Agents (`.claude/agents/`)
go-analyst (read-only review), go-improver (small refactors), go-tester (tests), go-db-expert (SQL, read-only),
sql-safety-reviewer (SQL and tenant safety, read-only), go-security-reviewer (read-only), go-perf-profiler
(benchmarks), go-ci-release (Makefile, CI, Docker).

## Always
- Finish every change with build, vet, gofmt and tests, and report the result.
- Follow the rules in `.claude/rules/` (database, Go style, safe changes, new documents). They load automatically.
- All SQL goes through `querybuilder` and `repository`. No raw SQL anywhere else.
- Ask before changing behavior, not just structure. List the behavior change and wait.

## Never
- Edit existing files in `migrations/` (add a new migration instead).
- Run SQL against a real database (read-only checks only when the user asks), commit or push to git, or delete
  files outside the repo.
- Add dependencies without asking.
