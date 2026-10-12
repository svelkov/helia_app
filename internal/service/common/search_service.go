package common

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"helia/internal/domain"
	"helia/internal/repository"

	helcommon "helia/internal/common"
)

// SearchService collects the searches that are used by more than one screen (partneri, konta,
// artikli, ...), like CommonService does for the combos: every search is defined once (a
// searchDefinition in searchDefinitions) and answered by the same handler (handler.SearchHandler,
// GET /api/search/:entity) for components.TableCombo.
type SearchService interface {
	// Search returns the rows of the search entity that match text (and the filters the search
	// accepts), ready for components.TableComboResults.
	Search(ctx context.Context, entity, text string, filters map[string]string) (domain.TableComboResults, error)
	// Entities returns the names of the defined searches.
	Entities() []string
}

// SearchResource implements SearchService.
type SearchResource struct {
	repo repository.BaseRepository[domain.SearchRowDto]
}

// NewSearchService creates the shared search service. The repository only runs the queries of the
// searches (every search selects its columns as key, c1, c2, ... into domain.SearchRowDto).
func NewSearchService(repo repository.BaseRepository[domain.SearchRowDto]) *SearchResource {
	return &SearchResource{repo: repo}
}

// searchLimit is the number of rows a search returns at most.
const searchLimit = 50

// searchDefinition is one search: its query, the columns of its results and how it is filtered.
type searchDefinition struct {
	// headers are the columns of the results (their Label, Width, TextAlign); the query selects the
	// key of the row as "key" and the cells of the columns as c1, c2, ... (as text).
	headers []domain.Fields
	query   string
	// period is the alias of the table whose god and kar are the ones of the session ("" = none).
	period string
	// conditions are fixed conditions of the search (without parameters).
	conditions []string
	// searchIn are the columns the text is searched in (ILIKE %text%).
	searchIn []string
	// filters are the columns the search can be limited by (equal), each with the names of the
	// parameters of the request that give its value, in the order they are tried (the first one that
	// is not empty is used), e.g. "f.konto": {"konto", "fkto", "pkto"}: the screens send the konto of
	// the partner as konto, fkto (the kupac/dobavljač of the document) or pkto (the prodavnica).
	filters map[string][]string
	// requiredFilters are the columns of filters that must have a value (not empty and not "-", the text of
	// a combo without a selected row): without it the search gives only the row "-" (no selection), e.g. the
	// mesta troška without the org. jedinica.
	requiredFilters []string
	orderBy         string
	// textColumns and detailColumns: see domain.TableComboResults.
	textColumns, detailColumns []int
}

// The names of the parameters of the konto and of the šifra of a partner: the generic konto and
// sifra, the kupac/dobavljač of the robni dokument (rdok.fkto, rdok.fana) and the prodavnica (rdok.pkto,
// rdok.pana).
var (
	robnoSearchKonto = []string{"konto", "fkto", "pkto", "kupackonto", "tkonto"}
	robnoSearchSifra = []string{"sifra", "fana", "pana", "kupacsifra"}
)

// searchDefinitions are the searches of the application. A new search is a new entry here.
var searchDefinitions = map[string]searchDefinition{
	// Partneri: the šifarnik of the partneri (the field Šifra of the partner).
	"partneri": {
		headers: []domain.Fields{
			{Label: "Šifra", Width: "10%"},
			{Label: "Naziv"},
			{Label: "Adresa"},
			{Label: "Mesto"},
			{Label: "PIB", TextAlign: "right"},
			{Label: "Matični broj", TextAlign: "right"},
		},
		query: `select coalesce(p.sifra, '') as key, coalesce(p.sifra, '') as c1, coalesce(p.naziv, '') as c2,
			coalesce(p.adresa, '') as c3, trim(coalesce(nullif(p.pobro, 0)::text, '') || ' ' || coalesce(p.mesto, '')) as c4,
			coalesce(p.pib, '') as c5, coalesce(p.matbr, '') as c6
			from partneri p`,
		period:        "p",
		searchIn:      []string{"p.sifra", "p.naziv", "p.pib", "p.matbr"},
		orderBy:       "p.sifra",
		textColumns:   []int{0, 1},
		detailColumns: []int{2, 3},
	},
	// Analitika: the analitika (fkpl.vkonta 1) of a konto with the data of the partner; the filter
	// "konto" limits it to one konto (e.g. the kupci of 2040).
	"analitika": {
		headers: []domain.Fields{
			{Label: "Konto", Width: "10%"},
			{Label: "Šifra", Width: "10%"},
			{Label: "Naziv"},
			{Label: "Adresa"},
			{Label: "Mesto"},
			{Label: "PIB", TextAlign: "right"},
		},
		query: `select f.sifra as key, f.konto as c1, f.sifra as c2, coalesce(p.naziv, f.naziv, '') as c3,
			coalesce(p.adresa, '') as c4, trim(coalesce(nullif(p.pobro, 0)::text, '') || ' ' || coalesce(p.mesto, '')) as c5,
			coalesce(p.pib, '') as c6
			from fkpl f left join partneri p on p.idpartneri = f.idpartneri`,
		period:        "f",
		conditions:    []string{"f.vkonta = 1"},
		searchIn:      []string{"f.sifra", "f.naziv", "p.naziv", "p.pib"},
		filters:       map[string][]string{"f.konto": robnoSearchKonto},
		orderBy:       "f.konto, f.sifra",
		textColumns:   []int{1, 2},
		detailColumns: []int{3, 4},
	},
	// Konta: the konta of the kontni plan (fkpl.vkonta 2, the field Konto).
	"konta": {
		headers: []domain.Fields{
			{Label: "Konto", Width: "20%"},
			{Label: "Naziv"},
		},
		query:       `select f.konto as key, f.konto as c1, coalesce(f.naziv, '') as c2 from fkpl f`,
		period:      "f",
		conditions:  []string{"f.vkonta = 2"},
		searchIn:    []string{"f.konto", "f.naziv"},
		orderBy:     "f.konto",
		textColumns: []int{0, 1},
	},
	// Kontni plan by id: the konta (fkpl.vkonta 2) and the analitike of the konta (vkonta 1) of the
	// kontni plan by their id (idfkpl), e.g. the konto zaliha of a magacin; a konto with analitike is
	// there more times (once per šifra), so the key is the id of the row. The last column is the vkonta
	// of the row (the screen shows it).
	"kontni-plan-id": {
		headers: []domain.Fields{
			{Label: "Konto", Width: "15%"},
			{Label: "Šifra", Width: "12%"},
			{Label: "Naziv"},
			{Label: "V. konta", Width: "10%", TextAlign: "center"},
		},
		query: `select f.idfkpl::text as key, f.konto as c1, coalesce(f.sifra, '') as c2, coalesce(f.naziv, '') as c3,
			f.vkonta::text as c4 from fkpl f`,
		period:      "f",
		conditions:  []string{"f.vkonta in (1, 2)"},
		searchIn:    []string{"f.konto", "f.sifra", "f.naziv"},
		orderBy:     "f.konto, f.vkonta desc, f.sifra",
		textColumns: []int{0, 1, 2},
	},
	// Magacini: the magacini of the period (key the magaciniid).
	"magacini": {
		headers: []domain.Fields{
			{Label: "Magacin", Width: "15%", TextAlign: "right"},
			{Label: "Opis"},
		},
		query:       `select m.magaciniid::text as key, m.mag::text as c1, coalesce(m.opis, '') as c2 from magacini m`,
		period:      "m",
		searchIn:    []string{"m.mag::text", "m.opis"},
		orderBy:     "m.mag",
		textColumns: []int{0, 1},
	},
	// Kontni plan: the konta of the kontni plan of the kind the screen sends as "vkonta" (2 the konta,
	// 3 the sintetička konta; without it all the konta, not the analitika).
	"kontni-plan": {
		headers: []domain.Fields{
			{Label: "Konto", Width: "20%"},
			{Label: "Naziv"},
		},
		query:       `select f.konto as key, f.konto as c1, coalesce(f.naziv, '') as c2 from fkpl f`,
		period:      "f",
		conditions:  []string{"f.vkonta <> 1"},
		searchIn:    []string{"f.konto", "f.naziv"},
		filters:     map[string][]string{"f.vkonta::text": {"vkonta"}},
		orderBy:     "f.konto",
		textColumns: []int{0, 1},
	},
	// Robni dokumenti: the robni dokumenti (rdok) of the period with their kupac and iznos; the filters
	// "vrd" and "rnalid" limit them to one vrsta dokumenta and one nalog (e.g. the broj dokumenta of the
	// fakture veleprodaje: the documents of the nalog).
	"robni-dokumenti": {
		headers: []domain.Fields{
			{Label: "Broj dokumenta", Width: "12%", TextAlign: "right"},
			{Label: "Datum dokumenta", Width: "12%"},
			{Label: "Konto/Šifra", Width: "14%"},
			{Label: "Naziv partnera"},
			{Label: "Iznos", TextAlign: "right"},
		},
		// The key is the id of the document (rdokid): the screen opens the selected document by it.
		query: `select r.rdokid::text as key, r.dokum::text as c1, coalesce(to_char(r.dadok, 'DD.MM.YYYY'), '') as c2,
			trim(coalesce(r.fkto, '') || ' ' || coalesce(r.fana, '')) as c3, coalesce(kup.naziv, '') as c4,
			coalesce(to_char(r.iznos, 'FM999999999990.00'), '') as c5
			from rdok r
			left join lateral (select f.naziv from fkpl f
				where f.god = r.god and f.kar = r.kar and f.konto = r.fkto and f.sifra = r.fana
				order by f.vkonta desc, f.idfkpl limit 1) kup on true`,
		period:        "r",
		searchIn:      []string{"r.dokum::text", "r.fana", "kup.naziv"},
		filters:       map[string][]string{"r.vrd::text": {"vrd"}, "r.rnalid::text": {"rnalid"}},
		orderBy:       "r.dokum desc",
		textColumns:   []int{0, 3},
		detailColumns: []int{1},
	},
	// Sintetička konta: the sintetička konta of the kontni plan (fkpl.vkonta 3).
	"sinteticka-konta": {
		headers: []domain.Fields{
			{Label: "Konto", Width: "20%"},
			{Label: "Naziv"},
		},
		query:       `select f.konto as key, f.konto as c1, coalesce(f.naziv, '') as c2 from fkpl f`,
		period:      "f",
		conditions:  []string{"f.vkonta = 3"},
		searchIn:    []string{"f.konto", "f.naziv"},
		orderBy:     "f.konto",
		textColumns: []int{0, 1},
	},
	// Mesta isporuke: the mesta isporuke (fisp) of the partneri; the konto and the šifra of the
	// partner (konto/fkto/pkto and sifra/fana/pana) limit them to the ones of one partner (the kupac of
	// the document).
	"mesta-isporuke": {
		headers: []domain.Fields{
			{Label: "MI", Width: "8%", TextAlign: "right"},
			{Label: "Naziv"},
			{Label: "Adresa"},
			{Label: "Mesto"},
			{Label: "GLN", TextAlign: "right"},
		},
		query: `select coalesce(f.mi, 0)::bigint::text as key, coalesce(f.mi, 0)::bigint::text as c1,
			coalesce(f.naziv, '') as c2, coalesce(f.adresa, '') as c3,
			trim(coalesce(nullif(f.pobro, 0)::text, '') || ' ' || coalesce(f.mesto, '')) as c4,
			coalesce(nullif(f.gln, 0)::text, '') as c5
			from fisp f`,
		period:        "f",
		searchIn:      []string{"f.mi::text", "f.naziv", "f.adresa", "f.mesto"},
		filters:       map[string][]string{"f.konto": robnoSearchKonto, "f.sifra": robnoSearchSifra},
		orderBy:       "f.konto, f.sifra, f.mi",
		textColumns:   []int{0, 1},
		detailColumns: []int{2, 3},
	},
	// Komercijalisti: the šifarnik of the komercijalisti.
	"komercijalisti": {
		headers: []domain.Fields{
			{Label: "Šifra", Width: "10%", TextAlign: "right"},
			{Label: "Ime i prezime"},
			{Label: "Mesto"},
			{Label: "Telefon"},
		},
		query: `select coalesce(k.sifkom, 0)::text as key, coalesce(k.sifkom, 0)::text as c1,
			coalesce(k.imeprezime, '') as c2, coalesce(k.mesto, '') as c3,
			coalesce(nullif(k.telmob, ''), k.telposao, '') as c4
			from komercijalisti k`,
		period:      "k",
		searchIn:    []string{"k.sifkom::text", "k.imeprezime", "k.mesto"},
		orderBy:     "k.sifkom",
		textColumns: []int{0, 1},
	},
	// The šifarnici of the combos of the documents, keyed by the primary key of the šifarnik and shown
	// as šifra and naziv only.
	// Org. jedinice (orgjed.idorgjed).
	"org-jedinice": {
		headers: []domain.Fields{
			{Label: "Šifra", Width: "20%"},
			{Label: "Naziv"},
		},
		query:       `select o.idorgjed::text as key, coalesce(o.ojozn, '') as c1, coalesce(o.naziv, '') as c2 from orgjed o`,
		period:      "o",
		searchIn:    []string{"o.ojozn", "o.naziv"},
		orderBy:     "o.ojozn",
		textColumns: []int{0, 1},
	},
	// Mesta troška (mestotr.mestotrid); the filter "idorgjed" limits them to one org. jedinica.
	"mesta-troska": {
		headers: []domain.Fields{
			{Label: "Šifra", Width: "20%"},
			{Label: "Naziv"},
		},
		query:    `select m.mestotrid::text as key, coalesce(m.mtroska, '') as c1, coalesce(m.opis, '') as c2 from mestotr m`,
		period:   "m",
		searchIn: []string{"m.mtroska", "m.opis"},
		filters:  map[string][]string{"m.idorgjed::text": {"idorgjed", "idorgjed1"}},
		// Without the org. jedinica only "-" (the mesta troška of the selected org. jedinica, like the faktura).
		requiredFilters: []string{"m.idorgjed::text"},
		orderBy:         "m.mtroska",
		textColumns:     []int{0, 1},
	},
	// Mesta isporuke by fisp.fispid (the search "mesta-isporuke" is keyed by the MI); the konto and
	// the šifra of the partner limit them to the ones of the kupac.
	"mesta-isporuke-id": {
		headers: []domain.Fields{
			{Label: "MI", Width: "15%", TextAlign: "right"},
			{Label: "Naziv"},
		},
		query:       `select f.fispid::text as key, coalesce(f.mi, 0)::bigint::text as c1, coalesce(f.naziv, '') as c2 from fisp f`,
		period:      "f",
		searchIn:    []string{"f.mi::text", "f.naziv"},
		filters:     map[string][]string{"f.konto": robnoSearchKonto, "f.sifra": robnoSearchSifra},
		orderBy:     "f.mi",
		textColumns: []int{0, 1},
	},
	// Komercijalisti by komercijalisti.komid (the search "komercijalisti" is keyed by the šifra).
	"komercijalisti-id": {
		headers: []domain.Fields{
			{Label: "Šifra", Width: "15%", TextAlign: "right"},
			{Label: "Ime i prezime"},
		},
		query:       `select k.komid::text as key, coalesce(k.sifkom, 0)::text as c1, coalesce(k.imeprezime, '') as c2 from komercijalisti k`,
		period:      "k",
		searchIn:    []string{"k.sifkom::text", "k.imeprezime"},
		orderBy:     "k.sifkom",
		textColumns: []int{0, 1},
	},
	// Banke za uplatu - izvoz (bnkizv.bnkizvid) with their SWIFT, broj računa and beneficiary.
	"banke-izvoz": {
		headers: []domain.Fields{
			{Label: "Šifra", TextAlign: "right"},
			{Label: "Naziv"},
			{Label: "SWIFT"},
			{Label: "Broj računa"},
			{Label: "Beneficiary"},
		},
		query: `select b.bnkizvid::text as key, coalesce(b.sifbank, 0)::text as c1, coalesce(b.bnkdes, '') as c2,
			coalesce(b.swiftadr, '') as c3, coalesce(b.brojrac, '') as c4, coalesce(b.beneficiary, '') as c5 from bnkizv b`,
		period:      "b",
		searchIn:    []string{"b.sifbank::text", "b.bnkdes", "b.swiftadr", "b.brojrac", "b.beneficiary"},
		orderBy:     "b.sifbank",
		textColumns: []int{0, 1},
	},
	// POPDV polja of the evidencije PDV (fvepdv) of the type the screen sends as "vktip" (e.g. "I" for
	// the izlazni računi), by the order of the evidencija and the polje; the key is the polje (the
	// documents store it, e.g. rdok.polje).
	"popdv": {
		headers: []domain.Fields{
			{Label: "Polje"},
			{Label: "Opis"},
			{Label: "Tip", TextAlign: "right"},
			{Label: "Deo", TextAlign: "right"},
		},
		query: `select coalesce(p.polje, '') as key, coalesce(p.polje, '') as c1, coalesce(p.opis, '') as c2,
			coalesce(p.tip, 0)::text as c3, coalesce(p.deo, 0)::text as c4
			from popdv p join fvepdv e on e.fvepdvid = p.fvepdvid`,
		period:      "e",
		searchIn:    []string{"p.polje", "p.opis"},
		filters:     map[string][]string{"e.vktip": {"vktip"}},
		orderBy:     "e.vkrbr, p.polje",
		textColumns: []int{0, 1},
	},
	// Poreske kategorije (porkat.porkatid) of the PDV obveznici or of the firms out of the PDV (the
	// filter "obveznik", "true" or "false"), with the članovi of the zakon the napomena o poreskom
	// oslobođenju is made of (the šifarnik is common to all the periods).
	"poreske-kategorije": {
		headers: []domain.Fields{
			{Label: "Šifra"},
			{Label: "Kategorija"},
			{Label: "Tekst"},
			{Label: "Zakon"},
			{Label: "Član", TextAlign: "right"},
			{Label: "Stav", TextAlign: "right"},
			{Label: "Tačka", TextAlign: "right"},
			{Label: "Podtačka", TextAlign: "right"},
		},
		query: `select p.porkatid::text as key, coalesce(p.keykat, '') as c1, coalesce(p.category, '') as c2,
			coalesce(p.text, '') as c3, coalesce(p.law, '') as c4, coalesce(p.article, '') as c5,
			coalesce(p.paragraph, '') as c6, coalesce(p.point, '') as c7, coalesce(p.subpoint, '') as c8
			from porkat p`,
		searchIn:    []string{"p.keykat", "p.category", "p.text"},
		filters:     map[string][]string{"p.obveznik::text": {"obveznik"}},
		orderBy:     "p.keykat, p.category",
		textColumns: []int{0, 1},
	},
	// Valute (valute.idvalute; the šifarnik is common to all the periods).
	"valute": {
		headers: []domain.Fields{
			{Label: "Šifra", Width: "15%", TextAlign: "right"},
			{Label: "Naziv"},
		},
		query:       `select v.idvalute::text as key, coalesce(v.sifval, 0)::text as c1, coalesce(v.naziv, '') as c2 from valute v`,
		searchIn:    []string{"v.sifval::text", "v.naziv", "v.oznval"},
		orderBy:     "v.sifval",
		textColumns: []int{0, 1},
	},
	// Artikli: the šifarnik of the artikli (rsif, the field Šifra of the robno).
	"artikli": {
		headers: []domain.Fields{
			{Label: "Šifra", Width: "12%", TextAlign: "right"},
			{Label: "Naziv"},
			{Label: "JM", Width: "8%"},
			{Label: "Barkod", Width: "20%"},
		},
		query: `select r.sifra::text as key, r.sifra::text as c1, coalesce(r.naziv, '') as c2,
			coalesce(r.jm, '') as c3, coalesce(r.barkod, '') as c4 from rsif r`,
		period:      "r",
		searchIn:    []string{"r.sifra::text", "r.naziv", "r.barkod"},
		orderBy:     "r.sifra",
		textColumns: []int{0, 1},
	},
}

// Entities returns the names of the defined searches, sorted.
func (s *SearchResource) Entities() []string {
	names := make([]string, 0, len(searchDefinitions))
	for name := range searchDefinitions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Search runs the search entity: the rows of the period of the session that match text in any of the
// columns of the search and the accepted filters, at most searchLimit of them.
func (s *SearchResource) Search(ctx context.Context, entity, text string, filters map[string]string) (domain.TableComboResults, error) {
	def, found := searchDefinitions[entity]
	if !found {
		return domain.TableComboResults{}, fmt.Errorf("unknown search %q", entity)
	}
	session, err := sessionFrom(ctx)
	if err != nil {
		return domain.TableComboResults{}, err
	}
	qb := helcommon.NewQueryBuilder(def.query, true)
	if def.period != "" {
		qb.AddEqual(def.period+".god", session.SelectedGod)
		qb.AddEqual(def.period+".kar", session.SelectedKar)
	}
	for _, condition := range def.conditions {
		qb.AddCustomCondition(condition)
	}
	// The filters in a fixed order (of the column), so the same request builds the same query; the
	// value of a column is the first of its parameters that is not empty.
	columns := make([]string, 0, len(def.filters))
	for column := range def.filters {
		columns = append(columns, column)
	}
	sort.Strings(columns)
	zadato := map[string]bool{}
	for _, column := range columns {
		for _, name := range def.filters[column] {
			if value := strings.TrimSpace(filters[name]); value != "" && value != "-" {
				qb.AddEqual(column, value)
				zadato[column] = true
				break
			}
		}
	}
	// A required filter without a value: only the row "-" (the empty key).
	for _, column := range def.requiredFilters {
		if !zadato[column] {
			prazan := make([]string, len(def.headers))
			if len(prazan) > 0 {
				prazan[0] = "-"
			}
			return domain.TableComboResults{
				Table:         domain.TableData{Headers: def.headers, Rows: []domain.TableRow{{ID: "", Fields: prazan}}},
				TextColumns:   def.textColumns,
				DetailColumns: def.detailColumns,
			}, nil
		}
	}
	// "-" is the text of a combo without a selected row (TableComboConfig.AllowEmpty): no text.
	if text = strings.TrimSpace(text); text == "-" {
		text = ""
	}
	if text != "" && len(def.searchIn) > 0 {
		n := qb.GetArgsCount() + 1
		parts := make([]string, 0, len(def.searchIn))
		for _, column := range def.searchIn {
			parts = append(parts, fmt.Sprintf("%s ILIKE '%%' || $%d || '%%'", column, n))
		}
		qb.AddCustomCondition("(" + strings.Join(parts, " OR ") + ")")
		qb.AddArgs(text)
	}
	qb.AddOrderBy(def.orderBy)
	qb.SetLimit(searchLimit)
	query, args := qb.Build()
	rows, err := s.repo.GetAllCustom(ctx, query, "", args, "", "")
	if err != nil {
		return domain.TableComboResults{}, err
	}
	res := domain.TableComboResults{
		Table:         domain.TableData{Headers: def.headers},
		TextColumns:   def.textColumns,
		DetailColumns: def.detailColumns,
	}
	for _, row := range *rows {
		res.Table.Rows = append(res.Table.Rows, domain.TableRow{ID: row.Key, Fields: row.Cells(len(def.headers))})
	}
	return res, nil
}
