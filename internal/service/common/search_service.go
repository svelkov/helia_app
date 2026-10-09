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
	orderBy string
	// textColumns and detailColumns: see domain.TableComboResults.
	textColumns, detailColumns []int
}

// The names of the parameters of the konto and of the šifra of a partner: the generic konto and
// sifra, the kupac/dobavljač of the robni dokument (rdok.fkto, rdok.fana) and the prodavnica (rdok.pkto,
// rdok.pana).
var (
	robnoSearchKonto = []string{"konto", "fkto", "pkto"}
	robnoSearchSifra = []string{"sifra", "fana", "pana"}
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
	for _, column := range columns {
		for _, name := range def.filters[column] {
			if value := strings.TrimSpace(filters[name]); value != "" {
				qb.AddEqual(column, value)
				break
			}
		}
	}
	if text = strings.TrimSpace(text); text != "" && len(def.searchIn) > 0 {
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
