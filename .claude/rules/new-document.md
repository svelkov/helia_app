# New documents and options (from the legacy WinDev program)

Use this when the user asks for a new document or option of the legacy program (e.g. the robno document
`rob_ProFakture`) and gives screenshots of its windows and a PDF of its WinDev code.

## 1. Read the legacy first
- Read every screenshot and the whole PDF before writing code. Map:
  - the windows: the header (e.g. `rob_ProFakture`) and the stavke (the `_DET` window), tabs, dialogs;
  - the controls (`EDT_...`, `COMBO_...`, `TABLE_...`) to fields, with their labels, order, defaults and which are
    read-only;
  - the procedures to functions: initialization (`PuniCOMBO`, `NoviDOK`), the exits of the fields (`exit of
    EDT_x`), validation (`Validacija`, `ValidacijaRDOK`), saving (`SnimiRDOK`, `CreateRpro`, `EditRpro`),
    deleting (`BrisiRPRO`, "Briši"), the posting (`AzurirajRNAL`, `AzurirajRSTA`, `AzurirajStanja`), prints.
- Find the group of the document (`dokvrsta.grpdok`, e.g. `PRO`) and its entry in `robnoDokumentaUnosEkrani`
  (`internal/handler/robno/robnadokumentaunos_handler.go`): it names the legacy window of every group.
- If the PDF is missing a procedure the screens need, or a screenshot is unreadable, list it and ask.

## 2. Find what already exists
- The closest existing document: the screens in `robnoDokumentaUnosEkrani` with an `open` function (fakture,
  fakture usluga and avansni računi, interni računi, porudžbenica, opšti dokumenti with prenosnice, gradilište,
  popis, nivelacije). A document that differs only in a few fields or rules is a variant of that screen (flags in
  its domain types, service and template, like the groups of the opšti dokumenti); a document with its own
  windows gets its own screen (like the porudžbenica).
- Reuse the shared functions of `internal/service/robno` instead of writing new ones, e.g.: `GetRobStavkeKontekst`,
  `SaveRobniDokument`, `robFaktureSnimanje`, `ValidateRobFakture`, `GetRobKupac`, `robnoUnosObavezanID`,
  `robBrisiDokument` / `robBrisiDokumentSa`, `AzurirajRNAL`, `AzurirajRSTA`, `AzurirajDRSTA`,
  `robAzurirajStanjeStavke`, `AzurirajStanja`, `PreracunRataDokumenta`, `ObracunFakture`, and the existing
  `RobnoRepository` methods (`internal/repository/robno_*_repository.go`). Search before adding.

## 3. Plan and wait
Before editing, present:
- the legacy → new mapping (windows, fields, procedures);
- the files to create and to change (see 4), and what is reused;
- the behavior to decide: the questions and messages of the legacy (Da/Ne dialogs), what is not ported yet,
  differences from the legacy you propose.
Wait for approval. Port the legacy behavior faithfully; do not invent features.

## 4. Files, in this order
1. **Domain** `internal/domain/model_robno<name>.go`: the params of the header and the stavka, the document view,
   the stavka form. Reuse existing types (e.g. `RobnoFaktureDokumentParams`, `RobStavkaRow`) where they fit.
2. **Repository**: new statements only as methods of `RobnoRepository` in the matching `robno_*_repository.go`
   (or a new `robno_<name>_repository.go`), following `database.md` (tenant via `inPeriod`, `execOne` for updates
   by id), each with a golden case in its test file.
3. **Service** `internal/service/robno/robnadokumenta_<name>_service.go`: methods of `RobnoDokumentaResource`
   (document, validation, save, delete, artikal of a stavka, stavka validation/save/delete, stavka form, list).
   Add them to the `RobnoDokumentaService` interface and the grid columns (`domain.Fields`) to the constructor in
   `robnadokumenta_service.go`. Doc comments name the legacy procedure each function ports.
4. **Handler** `internal/handler/robno/rob<name>_handler.go`: methods of `RobnoDokumentaUnosHandler`: the open
   function (`otvori<Name>`), the routes in a `robno<Name>Routes(r)` called from `AddRoutes` (full literal paths,
   see `go-style.md`), element-id constants of the screen. Register the group(s) in `robnoDokumentaUnosEkrani`
   with `open` in the same change. Handlers pass `c.Request.Context()` to the service.
5. **Template** `frontend/templates/robno/rob<name>.templ` built from `frontend/components` (input, combobox,
   table, button, …), layout like the existing screens; CSS class constants in `internal/common/const.go`; then
   `templ fmt` and `templ generate -f` on that file only.
6. **Search / combos**: a new lookup of a combobox goes into `internal/service/common/search_service.go` (or
   reuses an existing search).
7. **Translations**: every new label in all three files `i18n/translations/en.json`, `sr.json` and `ср.json`.
8. **Print** (if the legacy prints the document): `robnadokumenta_stampa_service.go` and
   `frontend/templates/reports/robno/`.
Other modules (finansijsko, kamate, os) follow the same order with their own packages; a simple šifarnik
(one table, list + form) uses `registerGenericEntity` in `factory.go` instead of its own handler.

## 5. Finish
- Build, vet, the formatting check and `go test -count=1 -skip TestZzPorTmp ./...` (including
  `TestStartupWiring`, which checks the routes register); add tests for the new repository methods and the
  service logic.
- Report: files created and changed, the legacy features ported and the ones left out, behavior differences, and
  the steps for the user to try the screen (which nalog / vrsta opens it). You can't open the screen yourself:
  say so.
