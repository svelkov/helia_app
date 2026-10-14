# Go style

## Formatting and checks
- Run `gofmt` and `go vet`. Fix every warning you introduce.
- The Go files have CRLF line endings, so `gofmt -l .` lists them all. Check the files you changed with
  `tr -d '\r' < file | gofmt -l`; format with `gofmt -w file`.
- Templ: edit the `.templ` source, never the generated `*_templ.go`. Run `templ fmt <file>` and
  `templ generate -f <file>` on the single file you changed, from the repo root. Never on a folder:
  `templ generate -path <dir>` rewrites every `*_templ.go` of the folder.

## Code
- Handle every error. Wrap with context: `fmt.Errorf("load partner %d: %w", id, err)`.
  Use `errors.Is` / `errors.As` to check.
- `context.Context` is the first parameter of anything that does I/O. Handlers pass `c.Request.Context()` to
  services, never the `*gin.Context`.
- Layers: handler → service → repository → infrastructure/db. `internal/domain` imports no other internal
  package; never import upwards. SQL only in `internal/repository` (see `database.md`).
- Accept interfaces, return structs. Define small interfaces where they are used.
- No globals and no `init()` for wiring. Inject dependencies through constructors; the wiring is in
  `factory.go` (`setEntities`).
- Package names are short and describe responsibility (`repository`, `querybuilder`), never `utils` / `helpers` /
  `common` for new code. `internal/common` keeps only tiny shared helpers and constants.
- No stutter: `partner.Service`, not `partner.PartnerService`.
- Use `any`, not `interface{}`, in new code.
- Keep functions small. Do not add abstractions without a second use.
- Names follow the domain: the Serbian terms of the legacy program (nalog, stavka, konto, šifra, magacin, rok …)
  in identifiers and comments; comments are English sentences around them.

## Project conventions
- CSS class constants (`Class...`) go in `internal/common/const.go`, never as local `const` blocks in `.templ`
  files. Element ids of one screen go in that screen's handler const block and reach the template through the UI
  struct or a template parameter (templates can't import the handler package).
- Routes: write the full path as a string literal in the route registration, e.g.
  `r.DELETE("/api/robno-dokumenta/interni/brisi/:id", h.DeleteRobInterni)`, never `const + "/:id"`. URL constants
  used by buttons and links must equal the route paths.
- A new robno document screen registers its dokvrsta groups (`grpdok`) in `robnoDokumentaUnosEkrani`
  (`internal/handler/robno/robnadokumentaunos_handler.go`) with its open function in the same change, so the
  nalog dialog opens it.

## Tests
- Table-driven with `t.Run`; handlers with `httptest` (assert on the returned HTML fragment or JSON).
- Run `go test -count=1 -skip TestZzPorTmp ./...`. `TestZzPorTmp` reaches the real database.
