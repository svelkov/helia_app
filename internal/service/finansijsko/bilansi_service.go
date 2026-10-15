package finansijsko

import (
	"context"
	"fmt"
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	"helia/internal/service"
	"math"
	"reflect"
	"strings"
	"time"
)

// BilansiService defines the interface for operations related to Bilansi (Balance Sheets).
type BilansiService interface {
	GetZakljucniListAnalitika(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error
	GetZakljucniListSintetika(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error
	GetZakljucniListSubsintetika(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error
	GetZakljucniListZaStampu(ctx context.Context, tbl *domain.TableData, tblSummary *domain.TableData, params domain.ZakljucniParams, nDuzSint int) error
	GetBilansStanja(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, searchText string, skraceni bool) error
	GetBilansStanjaObrada(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, stanjeNaDan string, skraceni, lPGODizPS bool) error
	GetBilansStanjaZaStampu(ctx context.Context, tbl *domain.TableData, tipStampe string, skraceni bool) error
	GetBilansStanjaXML(ctx context.Context, skraceni bool) ([]byte, error)
	GetBilansUspeha(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, searchText string, skraceni bool) error
	GetBilansUspehaZaStampu(ctx context.Context, tbl *domain.TableData, tipStampe string) error
	GetBilansUspehaObrada(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, odDatuma, doDatuma string, skraceni, lPGODizPG bool) error
	GetByID(ctx context.Context, idField string, idValue int64) (*domain.Bils, error)
	Update(ctx context.Context, entity *domain.Bils, idField string, idValue interface{}, tableFields []domain.Fields) error
	Add(ctx context.Context, entity *domain.Bils, idField string, tableFields []domain.Fields) (int64, error)
	MapEntityToValues(entity *domain.Bils, tableFields []domain.Fields) []domain.Fields
	ValidateBilansStanja(entity *domain.Bils) []domain.FieldError
	DeleteBilansStanja(ctx context.Context, id int64) error
	GetFieldCache() map[string]reflect.StructField
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetZakljucniTableFields() []domain.Fields
	GetBilansStanjaTableFields() []domain.Fields
	GetBilansStanjaStampaTableFields() []domain.Fields
	GetBilansUspehaTableFields() []domain.Fields
	GetBilansUspehaStampaTableFields() []domain.Fields
	// Bilu (Bilans Uspeha) methods
	GetByIDBilu(ctx context.Context, idField string, idValue int64) (*domain.Bilu, error)
	UpdateBilu(ctx context.Context, entity *domain.Bilu, idField string, idValue interface{}, tableFields []domain.Fields) error
	AddBilu(ctx context.Context, entity *domain.Bilu, idField string, tableFields []domain.Fields) (int64, error)
	MapEntityToValuesBilu(entity *domain.Bilu, tableFields []domain.Fields) []domain.Fields
	ValidateBilansUspeha(entity *domain.Bilu) []domain.FieldError
	DeleteBilansUspeha(ctx context.Context, id int64) error
	GetFieldCacheBilu() map[string]reflect.StructField
}

// BilansiResource implements the BilansiService interface.
type BilansiResource struct {
	biluService                   *service.BaseService[domain.Bilu]
	biluRepo                      *repository.BaseRepository[domain.Bilu]
	bilsService                   *service.BaseService[domain.Bils]
	bilsRepo                      *repository.BaseRepository[domain.Bils]
	fproRepo                      *repository.BaseRepository[domain.FproDto]
	fvrRepo                       *repository.BaseRepository[domain.Fvr]
	fkplRepo                      *repository.BaseRepository[domain.Fkpl]
	fin                           *repository.FinansijskoRepository
	zakljucniTableFields          []domain.Fields
	bilansStanjaTableFields       []domain.Fields
	bilansUspehaTableFields       []domain.Fields
	bilansUspehaStampaTableFields []domain.Fields
	bilansStanjaStampaTableFields []domain.Fields
	cfg                           config.Config
}

func NewBilansiService(
	biluService *service.BaseService[domain.Bilu],
	biluRepo *repository.BaseRepository[domain.Bilu],
	bilsService *service.BaseService[domain.Bils],
	bilsRepo *repository.BaseRepository[domain.Bils],
	fproRepo *repository.BaseRepository[domain.FproDto],
	fvrRepo *repository.BaseRepository[domain.Fvr],
	fkplRepo *repository.BaseRepository[domain.Fkpl],
	cfg config.Config,
) *BilansiResource {
	rs := &BilansiResource{
		biluService: biluService,
		biluRepo:    biluRepo,
		bilsService: bilsService,
		bilsRepo:    bilsRepo,
		fproRepo:    fproRepo,
		fvrRepo:     fvrRepo,
		fkplRepo:    fkplRepo,
		fin:         repository.NewFinansijskoRepository(bilsRepo.DB),
		cfg:         cfg,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *BilansiResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return repository.GetFvrData(ctx, s.fvrRepo)
}

func (s *BilansiResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Bils, error) {
	return s.bilsRepo.GetByID(ctx, idField, idValue)
}
func (s *BilansiResource) Update(ctx context.Context, entity *domain.Bils, idField string, idValue interface{}, tableFields []domain.Fields) error {
	return s.bilsRepo.Update(ctx, entity, idField, idValue, tableFields)
}
func (s *BilansiResource) Add(ctx context.Context, entity *domain.Bils, idField string, tableFields []domain.Fields) (int64, error) {
	return s.bilsRepo.Create(ctx, entity, idField, tableFields)
}
func (s *BilansiResource) MapEntityToValues(entity *domain.Bils, tableFields []domain.Fields) []domain.Fields {
	return s.bilsService.MapEntityToValues(entity, tableFields)
}

// GetZakljucniTableFields returns the table field definitions for Zakljucni list
func (s *BilansiResource) GetZakljucniTableFields() []domain.Fields {
	return s.zakljucniTableFields
}

// GetBilansStanjaTableFields returns the table field definitions for Bilans stanja
func (s *BilansiResource) GetBilansStanjaTableFields() []domain.Fields {
	return s.bilansStanjaTableFields
}
func (s *BilansiResource) GetBilansStanjaStampaTableFields() []domain.Fields {
	return s.bilansStanjaStampaTableFields
}

// GetBilansUspehaTableFields returns the table field definitions for Bilans uspeha
func (s *BilansiResource) GetBilansUspehaTableFields() []domain.Fields {
	return s.bilansUspehaTableFields
}

// GetBilansUspehaStampaTableFields returns the table field definitions for Bilans uspeha stampa
func (s *BilansiResource) GetBilansUspehaStampaTableFields() []domain.Fields {
	return s.bilansUspehaStampaTableFields
}

// GetZakljucniListAnalitika retrieves the analitika Zakljucni list (tipLista "1") in a single
// pass: one query returns every hierarchy level (analitika, sintetika, grupa, klasa) and the
// rows are mapped into tbl with the level marker in Fields[0]:
//
//	"1", "2", ... analitika rows (konto + sifra), numbered
//	"G1"          sintetika subtotal (LEFT(konto, cfg.NDuzSint))
//	"G2"          grupa subtotal (LEFT(konto, 2))
//	"G3"          klasa subtotal (LEFT(konto, 1))
//
// On the totals pass (getTotalRecords) only the analitika leaves are counted and summed, so the
// subtotal levels are not counted twice.
func (s *BilansiResource) GetZakljucniListAnalitika(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error {

	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	b, err := s.buildZakljucniListAnalitikaQuery(ctx, session, params, common.TipStampePreview)
	if err != nil {
		return err
	}
	// Everything up to the end of the requested page is fetched (the totals pass fetches the whole
	// hierarchy without LIMIT): the pagination counters are the number of hierarchy rows - detail
	// rows and subtotals - and the rows before the page are the ones that continue the redni broj of
	// the accounts across the pages.
	if !getTotalRecords {
		b.Limit(currentPage * pageSize)
	}

	tbl.Headers = s.GetZakljucniTableFields()
	common.SetupTablePagination(tbl, currentPage, pageSize)

	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	// Totals pass: every hierarchy level is an aggregate of the analitika rows, so only the
	// leaves are counted and summed - adding the subtotal levels up would count twice. The
	// pagination counters however count the whole hierarchy (the number of rendered rows).
	if getTotalRecords {
		leaves := make([]domain.FproDto, 0, len(*entities))
		for _, entity := range *entities {
			if entity.NivoOrder == zakljucniNivoAnalitika {
				leaves = append(leaves, entity)
			}
		}
		setZakljucniTotals(tbl, leaves, pageSize, len(*entities))
		return nil
	}

	// The query returns the hierarchy already ordered level by level, so every subtotal row
	// follows the rows it aggregates. TableRow.ID carries the level marker: the running row
	// number of the analitika rows and G1/G2/G3 for the sintetika/grupa/klasa subtotals.
	//
	// The redni broj column (Fields[0]) numbers the analitika rows only: the subtotal rows
	// (sintetika, grupa, klasa) leave it empty and do not consume a number. The numbering continues
	// across the pages (see zakljucniPageRows), so the first account of a page follows the last
	// account of the page before it.
	pageRows, rowNum := zakljucniPageRows(*entities, currentPage, pageSize, zakljucniNivoAnalitika)
	// Every subtotal row shows the sum of the netted sides below it, so the saldo columns foot
	// at every level and the grid shows the same numbers as the printed report.
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(common.NDuzSintFromContext(ctx, s.cfg), params.TipLista, *entities)
	nazivCache := make(map[string]string)
	for _, entity := range pageRows {
		marker := zakljucniNivoMarker(entity.NivoOrder, rowNum)
		redniBroj := ""
		if entity.NivoOrder == zakljucniNivoAnalitika {
			redniBroj = fmt.Sprintf("%d", rowNum)
			rowNum++
		}
		if marker == "" {
			continue // unknown level: skip it rather than render a nameless row
		}

		// The account names are not part of the aggregation, so they are resolved per unique
		// hierarchy key (cached, so every key is looked up only once).
		naziv, cached := nazivCache[entity.Konto]
		if !cached {
			naziv = s.getKontoNaziv(ctx, entity.Konto)
			nazivCache[entity.Konto] = naziv
		}

		saldoDug, saldoPot := saldoAcc.saldo(entity)
		fields := []string{
			redniBroj,
			entity.Konto,
			entity.Sifra,
			naziv,
			common.FormatNumberWithSystemLocale(entity.PocStanjeDug, 2),
			common.FormatNumberWithSystemLocale(entity.PocStanjePot, 2),
			common.FormatNumberWithSystemLocale(entity.PrometDug, 2),
			common.FormatNumberWithSystemLocale(entity.PrometPot, 2),
			common.FormatNumberWithSystemLocale(saldoDug, 2),
			common.FormatNumberWithSystemLocale(saldoPot, 2),
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{ID: marker, Fields: fields, HasUpdate: false, HasDelete: false})
	}

	return nil
}

// Hierarchy levels of the analitika Zakljucni list (GROUPING SETS); the values double as the
// "nivo_order" of the query result.
const (
	zakljucniNivoKlasa     = 1 // LEFT(konto, 1)
	zakljucniNivoGrupa     = 2 // LEFT(konto, 2)
	zakljucniNivoSintetika = 3 // LEFT(konto, cfg.NDuzSint)
	zakljucniNivoAnalitika = 4 // konto + sifra
)

// zakljucniNivoMarker maps a hierarchy level to the marker used in Fields[0] of the analitika
// report rows: the running row number for the analitika (detail) rows and G1/G2/G3 for the
// sintetika/grupa/klasa subtotals. An unknown level returns "", and the caller skips that row.
func zakljucniNivoMarker(nivo, rowNum int) string {
	if nivo == zakljucniNivoAnalitika {
		return fmt.Sprintf("%d", rowNum) // detail rows are numbered, not grouped
	}
	return zakljucniNivoGroupMarker(nivo, zakljucniNivoSintetika)
}

// zakljucniNivoGroupMarker maps a hierarchy level to the G<n> marker of a hierarchy whose
// innermost subtotal level is topNivo: G1 is the innermost subtotal and the outermost level
// (klasa) gets the highest number, so every list numbers its subtotals from its own finest
// level. Levels outside klasa..topNivo return "", and the caller skips that row.
func zakljucniNivoGroupMarker(nivo, topNivo int) string {
	if nivo < zakljucniNivoKlasa || nivo > topNivo {
		return ""
	}
	return fmt.Sprintf("G%d", topNivo-nivo+1)
}

// zakljucniDetailNivo returns the finest level of the hierarchy of a tipLista, i.e. the level
// that the grid and the printed report show as detail rows: konto + sifra (analitika), konto
// (subsintetika) or LEFT(konto, NDuzSint) (sintetika).
func zakljucniDetailNivo(tipLista string) int {
	if tipLista == "3" {
		return zakljucniNivoSintetika
	}
	return zakljucniNivoAnalitika
}

// zakljucniPageRows returns the rows of the requested page together with the redni broj the first
// account of the page gets.
//
// The data pass fetches the hierarchy from its beginning up to the end of the requested page (see
// the queries below), so the accounts rendered on the pages before it are known: the redni broj
// continues across the pages instead of restarting with 1 on every page.
//
//	rows       the rows fetched for the request (the whole hierarchy up to the end of the page)
//	detailNivo the level that carries the redni broj (the accounts of the list)
func zakljucniPageRows(rows []domain.FproDto, currentPage, pageSize, detailNivo int) ([]domain.FproDto, int) {
	firstRowNum := 1
	pageStart := (currentPage - 1) * pageSize
	if pageStart <= 0 {
		return rows, firstRowNum
	}
	if pageStart >= len(rows) {
		return nil, firstRowNum
	}
	for _, entity := range rows[:pageStart] {
		if entity.NivoOrder == detailNivo {
			firstRowNum++
		}
	}
	return rows[pageStart:], firstRowNum
}

// zakljucniSaldoAccumulator accumulates the netted saldo of the detail rows per subtotal key, so
// every subtotal row shows the sum of the sides below it instead of a netting of its own level.
// The saldo columns then add up at every level and the grid and the printed report agree.
//
// The subtotal rows themselves come from the query (they carry the raw column sums); only their
// saldo columns are taken from this accumulator.
type zakljucniSaldoAccumulator struct {
	nduzSint   int
	detailNivo int
	byKey      map[string][2]float64 // "<nivo>|<konto>" -> netted duguje, potražuje
	dugPot     [2]float64            // grand total of the netted sides of every detail row
}

// newZakljucniSaldoAccumulator creates an accumulator for a list whose finest level is detailNivo.
func newZakljucniSaldoAccumulator(nduzSint, detailNivo int) *zakljucniSaldoAccumulator {
	return &zakljucniSaldoAccumulator{
		nduzSint:   nduzSint,
		detailNivo: detailNivo,
		byKey:      make(map[string][2]float64),
	}
}

// zakljucniSaldoAccumulatorFromDetails builds the accumulator of a list: every row of the finest
// (detail) level is accumulated into its subtotal levels.
func zakljucniSaldoAccumulatorFromDetails(nduzSint int, tipLista string, entities []domain.FproDto) *zakljucniSaldoAccumulator {
	acc := newZakljucniSaldoAccumulator(nduzSint, zakljucniDetailNivo(tipLista))
	for _, entity := range entities {
		if entity.NivoOrder == acc.detailNivo {
			acc.addDetail(entity)
		}
	}
	return acc
}

// zakljucniSaldoKey builds the accumulator key of one subtotal level.
func zakljucniSaldoKey(nivo int, konto string) string {
	return fmt.Sprintf("%d|%s", nivo, konto)
}

// addDetail accumulates one detail row into its subtotal levels and into the grand total. The
// levels above the detail level are the konto prefixes: LEFT(konto, NDuzSint), LEFT(konto, 2)
// and LEFT(konto, 1).
func (a *zakljucniSaldoAccumulator) addDetail(ent domain.FproDto) {
	dug, pot := netirajSaldo(ent.PocStanjeDug+ent.PrometDug, ent.PocStanjePot+ent.PrometPot)
	a.dugPot[0] += dug
	a.dugPot[1] += pot

	if a.detailNivo > zakljucniNivoSintetika && a.nduzSint < len(ent.Konto) {
		a.add(zakljucniNivoSintetika, ent.Konto[:a.nduzSint], dug, pot)
	}
	if a.detailNivo > zakljucniNivoGrupa && len(ent.Konto) > 2 {
		a.add(zakljucniNivoGrupa, ent.Konto[:2], dug, pot)
	}
	if a.detailNivo > zakljucniNivoKlasa && len(ent.Konto) > 1 {
		a.add(zakljucniNivoKlasa, ent.Konto[:1], dug, pot)
	}
}

func (a *zakljucniSaldoAccumulator) add(nivo int, konto string, dug, pot float64) {
	key := zakljucniSaldoKey(nivo, konto)
	cur := a.byKey[key]
	a.byKey[key] = [2]float64{cur[0] + dug, cur[1] + pot}
}

// saldo returns the saldo to display for a row: its own netted side for a detail row and the
// accumulated sides of the rows below it for a subtotal row.
func (a *zakljucniSaldoAccumulator) saldo(ent domain.FproDto) (float64, float64) {
	if ent.NivoOrder == a.detailNivo {
		return netirajSaldo(ent.PocStanjeDug+ent.PrometDug, ent.PocStanjePot+ent.PrometPot)
	}
	v := a.byKey[zakljucniSaldoKey(ent.NivoOrder, ent.Konto)]
	return v[0], v[1]
}

// total returns the grand total: the sum of the netted sides of every detail row.
func (a *zakljucniSaldoAccumulator) total() (float64, float64) {
	return a.dugPot[0], a.dugPot[1]
}

// zakljucniPrintMarker maps a hierarchy level to the marker used in Fields[0] of the printed
// rows: the finest level of the printed list (finestNivo) is printed as a numbered detail row
// and every level above it gets the G<n> marker of its distance from it, so G1 stays the
// innermost subtotal. A level that is neither the finest one nor one of its parents returns "".
func zakljucniPrintMarker(nivo, finestNivo, rowNum int) string {
	if nivo == finestNivo {
		return fmt.Sprintf("%d", rowNum)
	}
	return zakljucniNivoGroupMarker(nivo, finestNivo-1)
}

// buildZakljucniListAnalitikaQuery builds the single-pass hierarchy query of the analitika
// Zakljucni list with the QueryBuilder:
//
//	filtered  the fpro transactions of the report (period filters only)
//	agg       GROUPING SETS produce one row per level: analitika (konto + sifra),
//	          sintetika (LEFT(konto, cfg.NDuzSint)), grupa (LEFT(konto, 2)) and klasa
//	          (LEFT(konto, 1))
//	outer     the level is mapped to nivo_order/konto/sifra and the result is ordered by the
//	          level aware sort key, so every subtotal row follows the rows it aggregates
//
// Nothing is hardcoded: the period comes from params, god/kar from the session (only when the
// ledger carries them) and the sintetika prefix length from the configuration. The search text
// matches the hierarchy key (konto and sifra).
func (s *BilansiResource) buildZakljucniListAnalitikaQuery(ctx context.Context, session *domain.UserSession, params domain.ZakljucniParams, printType string) (*querybuilder.Builder, error) {
	duzSin := session.GetDuzSin(s.cfg.NDuzSint)
	filter := querybuilder.New(`SELECT konto, sifra, fnal.tipdok, kat, iznos FROM fpro`).
		AddJoin("inner join fnal on fnal.idfnal = fpro.idfnal").
		// Both kinds of booking lines belong to this list: vkonta 1 = konto iz kontnog plana and
		// vkonta 2 = konto analitike/partnera (the filter the flattened zakljucniInnerQuery used for the
		// analitika and subsintetika lists). With only vkonta = 1 the list stopped at konto 4654,
		// although the bookings of the period go up to 8941.
		AddIn("fpro.vkonta", []any{"1"})
	// The range of the konta is compared on the sintetika prefix (like the other two lists of the
	// report and like the legacy query did).
	whereRange(filter, fmt.Sprintf("left(fpro.konto, %d)", duzSin), params.OdKonta, params.DoKonta)
	filter.AddCondition("fnal.danal", params.OdDatuma, ">=").
		AddCondition("fnal.danal", params.DoDatuma, "<=")
	// The range of the sifre applies to this list only (it books konto + sifra).
	whereRange(filter, "COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.OdSifre, params.DoSifre)
	// Klasa 9 accounts are excluded for printing only, exactly like the flattened
	// zakljucniInnerQuery does it.
	if printType == common.TipStampePrint && params.Klasa9 == "false" {
		filter.Where("fpro.konto NOT LIKE '9%'")
	}

	nivoOrderCase := fmt.Sprintf("CASE WHEN k4 IS NOT NULL THEN %d WHEN k3 IS NOT NULL THEN %d WHEN k2 IS NOT NULL THEN %d ELSE %d END AS nivo_order",
		zakljucniNivoAnalitika, zakljucniNivoSintetika, zakljucniNivoGrupa, zakljucniNivoKlasa)
	agg := fmt.Sprintf(`
		SELECT
			LEFT(konto, 1) AS k1,
			LEFT(konto, 2) AS k2,
			LEFT(konto, %d) AS k3,
			konto AS k4,
			sifra AS s4,
			SUM(CASE WHEN tipdok = '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS pocstanjedug,
			SUM(CASE WHEN tipdok = '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS pocstanjepot,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS prometdug,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS prometpot
		FROM filtered
		GROUP BY GROUPING SETS (
			(konto, sifra),
			(LEFT(konto, %d)),
			(LEFT(konto, 2)),
			(LEFT(konto, 1))
		)
	`, duzSin, duzSin)
	b, err := s.zakljucniHierarchy(ctx, filter, agg, fmt.Sprintf(`SELECT
		nivo_order,
		konto_grp AS konto,
		sifra_grp AS sifra,
		pocstanjedug, pocstanjepot, prometdug, prometpot
	FROM (
		SELECT
			%s,
			COALESCE(k4, k3, k2, k1) AS konto_grp,
			COALESCE(s4, '') AS sifra_grp,
			RPAD(COALESCE(k4, k3, k2, k1), 6, '~') AS sort_key,
			pocstanjedug, pocstanjepot, prometdug, prometpot
		FROM agg
	) sub`, nivoOrderCase))
	if err != nil {
		return nil, err
	}

	// The search text matches the hierarchy key (konto and sifra).
	whereZakljucniSearch(b, params.SearchText, "sub.konto", "sub.sifra")
	b.OrderByExpr(`sort_key COLLATE "C", sifra_grp COLLATE "C" NULLS LAST`)
	return b, nil
}

// GetZakljucniListSintetika retrieves the sintetika Zakljucni list (tipLista "3") in a single
// pass: one query returns the three aggregation levels (sintetika, grupa, klasa) and the rows
// are mapped into tbl with the level marker in Fields[0]:
//
//	"G1" sintetika subtotal (LEFT(konto, cfg.NDuzSint))
//	"G2" grupa subtotal (LEFT(konto, 2))
//	"G3" klasa subtotal (LEFT(konto, 1))
//
// There are no detail rows - every row is a subtotal - and on the totals pass the finest level
// (sintetika) is counted and summed, so the outer levels are not counted twice.
func (s *BilansiResource) GetZakljucniListSintetika(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error {

	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	b, err := s.buildZakljucniListSintetikaQuery(ctx, session, params, common.TipStampePreview)
	if err != nil {
		return err
	}
	// Everything up to the end of the requested page is fetched (the totals pass fetches the whole
	// hierarchy without LIMIT): the pagination counters are the number of hierarchy rows - detail
	// rows and subtotals - and the rows before the page are the ones that continue the redni broj of
	// the accounts across the pages.
	if !getTotalRecords {
		b.Limit(currentPage * pageSize)
	}

	tbl.Headers = s.GetZakljucniTableFields()
	common.SetupTablePagination(tbl, currentPage, pageSize)

	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	// Totals pass: every level aggregates the same transactions, so only the finest level is
	// counted and summed - adding the grupa/klasa levels up would count twice.
	if getTotalRecords {
		sintetike := make([]domain.FproDto, 0, len(*entities))
		for _, entity := range *entities {
			if entity.NivoOrder == zakljucniNivoSintetika {
				sintetike = append(sintetike, entity)
			}
		}
		setZakljucniTotals(tbl, sintetike, pageSize, len(*entities))
		return nil
	}

	// The query returns the hierarchy already ordered level by level, so every subtotal row
	// follows the rows it aggregates. TableRow.ID carries the level marker (G1/G2/G3).
	//
	// The redni broj column (Fields[0]) numbers the sintetika rows only - in this list they are
	// the accounts: the grupa and klasa subtotal rows leave it empty and do not consume a number.
	// The numbering continues across the pages (see zakljucniPageRows).
	// Every subtotal row shows the sum of the netted sides below it, so the saldo columns foot
	// at every level and the grid shows the same numbers as the printed report.
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(common.NDuzSintFromContext(ctx, s.cfg), params.TipLista, *entities)
	nazivCache := make(map[string]string)
	pageRows, rowNum := zakljucniPageRows(*entities, currentPage, pageSize, zakljucniNivoSintetika)
	for _, entity := range pageRows {
		marker := zakljucniNivoMarker(entity.NivoOrder, rowNum)
		redniBroj := ""
		if entity.NivoOrder == zakljucniNivoSintetika {
			redniBroj = fmt.Sprintf("%d", rowNum)
			rowNum++
		}
		if marker == "" {
			continue // unknown level: skip it rather than render a nameless row
		}

		// The account names are not part of the aggregation, so they are resolved per unique
		// hierarchy key (cached, so every key is looked up only once).
		naziv, cached := nazivCache[entity.Konto]
		if !cached {
			naziv = s.getKontoNaziv(ctx, entity.Konto)
			nazivCache[entity.Konto] = naziv
		}

		saldoDug, saldoPot := saldoAcc.saldo(entity)
		fields := []string{
			redniBroj,
			entity.Konto,
			"", // sifra: this list aggregates by konto prefix only
			naziv,
			common.FormatNumberWithSystemLocale(entity.PocStanjeDug, 2),
			common.FormatNumberWithSystemLocale(entity.PocStanjePot, 2),
			common.FormatNumberWithSystemLocale(entity.PrometDug, 2),
			common.FormatNumberWithSystemLocale(entity.PrometPot, 2),
			common.FormatNumberWithSystemLocale(saldoDug, 2),
			common.FormatNumberWithSystemLocale(saldoPot, 2),
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{ID: marker, Fields: fields, HasUpdate: false, HasDelete: false})
	}

	return nil
}

// buildZakljucniListSintetikaQuery builds the single-pass sintetika hierarchy query with the
// QueryBuilder - the analitika query without the konto + sifra detail level:
//
//	filtered  the fpro transactions of the report (period filters only)
//	agg       GROUPING SETS produce one row per level: sintetika (LEFT(konto, cfg.NDuzSint)),
//	          grupa (LEFT(konto, 2)) and klasa (LEFT(konto, 1))
//	outer     the level is mapped to nivo_order/konto and the result is ordered by the level
//	          aware sort key, so every subtotal row follows the rows it aggregates
//
// Nothing is hardcoded: the period comes from params, god/kar from the session (only when the
// ledger carries them) and the sintetika prefix length from the configuration. The search text
// matches the hierarchy key (konto).
func (s *BilansiResource) buildZakljucniListSintetikaQuery(ctx context.Context, session *domain.UserSession, params domain.ZakljucniParams, printType string) (*querybuilder.Builder, error) {
	duzSin := session.GetDuzSin(s.cfg.NDuzSint)
	filter := querybuilder.New(`SELECT konto, fnal.tipdok, kat, iznos FROM fpro`).
		AddJoin("inner join fnal on fnal.idfnal = fpro.idfnal")
	whereRange(filter, fmt.Sprintf("left(fpro.konto, %d)", duzSin), params.OdKonta, params.DoKonta)
	filter.AddCondition("fnal.danal", params.OdDatuma, ">=").
		AddCondition("fnal.danal", params.DoDatuma, "<=")
	// Klasa 9 accounts are excluded for printing only, exactly like the flattened
	// zakljucniInnerQuery does it.
	if printType == common.TipStampePrint && params.Klasa9 == "false" {
		filter.Where("fpro.konto NOT LIKE '9%'")
	}

	nivoOrderCase := fmt.Sprintf("CASE WHEN k3 IS NOT NULL THEN %d WHEN k2 IS NOT NULL THEN %d ELSE %d END AS nivo_order",
		zakljucniNivoSintetika, zakljucniNivoGrupa, zakljucniNivoKlasa)
	agg := fmt.Sprintf(`
		SELECT
			LEFT(konto, 1) AS k1,
			LEFT(konto, 2) AS k2,
			LEFT(konto, %d) AS k3,
			SUM(CASE WHEN tipdok = '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS pocstanjedug,
			SUM(CASE WHEN tipdok = '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS pocstanjepot,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS prometdug,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS prometpot
		FROM filtered
		GROUP BY GROUPING SETS (
			(LEFT(konto, %d)),
			(LEFT(konto, 2)),
			(LEFT(konto, 1))
		)
	`, duzSin, duzSin)
	b, err := s.zakljucniHierarchy(ctx, filter, agg, fmt.Sprintf(`SELECT
		nivo_order,
		konto_grp AS konto,
		pocstanjedug, pocstanjepot, prometdug, prometpot
	FROM (
		SELECT
			%s,
			COALESCE(k3, k2, k1) AS konto_grp,
			RPAD(COALESCE(k3, k2, k1), 6, '~') AS sort_key,
			pocstanjedug, pocstanjepot, prometdug, prometpot
		FROM agg
	) sub`, nivoOrderCase))
	if err != nil {
		return nil, err
	}

	// The search text matches the hierarchy key (konto).
	whereZakljucniSearch(b, params.SearchText, "sub.konto")
	b.OrderByExpr(`sort_key COLLATE "C" NULLS LAST`)
	return b, nil
}

// GetZakljucniListSubsintetika retrieves the subsintetika Zakljucni list (tipLista "2") in a
// single pass: one query returns every hierarchy level (subsintetika, sintetika, grupa, klasa)
// and the rows are mapped into tbl with the level marker in Fields[0]:
//
//	"G1" subsintetika subtotal (konto)
//	"G2" sintetika subtotal (LEFT(konto, cfg.NDuzSint))
//	"G3" grupa subtotal (LEFT(konto, 2))
//	"G4" klasa subtotal (LEFT(konto, 1))
//
// There are no detail rows - every row is a subtotal - and on the totals pass the finest level
// (subsintetika) is counted and summed, so the outer levels are not counted twice.
func (s *BilansiResource) GetZakljucniListSubsintetika(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error {

	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	b, err := s.buildZakljucniListSubsintetikaQuery(ctx, session, params, common.TipStampePreview)
	if err != nil {
		return err
	}
	// Everything up to the end of the requested page is fetched (the totals pass fetches the whole
	// hierarchy without LIMIT): the pagination counters are the number of hierarchy rows - detail
	// rows and subtotals - and the rows before the page are the ones that continue the redni broj of
	// the accounts across the pages.
	if !getTotalRecords {
		b.Limit(currentPage * pageSize)
	}
	tbl.Headers = s.GetZakljucniTableFields()
	common.SetupTablePagination(tbl, currentPage, pageSize)

	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
	}

	// Totals pass: every level aggregates the same transactions, so only the finest level (the
	// konto level, nivo 4 in this list) is counted and summed - adding the outer levels up would
	// count twice.
	if getTotalRecords {
		konta := make([]domain.FproDto, 0, len(*entities))
		for _, entity := range *entities {
			if entity.NivoOrder == zakljucniNivoAnalitika {
				konta = append(konta, entity)
			}
		}
		setZakljucniTotals(tbl, konta, pageSize, len(*entities))
		return nil
	}

	// The query returns the hierarchy already ordered level by level, so every subtotal row
	// follows the rows it aggregates. TableRow.ID carries the level marker (G1..G4).
	//
	// The redni broj column (Fields[0]) numbers the konto rows only - in this list they are the
	// accounts: the sintetika, grupa and klasa subtotal rows leave it empty and do not consume a
	// number. The numbering continues across the pages (see zakljucniPageRows).
	// Every subtotal row shows the sum of the netted sides below it, so the saldo columns foot
	// at every level and the grid shows the same numbers as the printed report.
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(common.NDuzSintFromContext(ctx, s.cfg), params.TipLista, *entities)
	nazivCache := make(map[string]string)
	pageRows, rowNum := zakljucniPageRows(*entities, currentPage, pageSize, zakljucniNivoAnalitika)
	for _, entity := range pageRows {
		marker := zakljucniNivoGroupMarker(entity.NivoOrder, zakljucniNivoAnalitika)
		redniBroj := ""
		if entity.NivoOrder == zakljucniNivoAnalitika {
			redniBroj = fmt.Sprintf("%d", rowNum)
			rowNum++
		}
		if marker == "" {
			continue // unknown level: skip it rather than render a nameless row
		}

		// The account names are not part of the aggregation, so they are resolved per unique
		// hierarchy key (cached, so every key is looked up only once).
		naziv, cached := nazivCache[entity.Konto]
		if !cached {
			naziv = s.getKontoNaziv(ctx, entity.Konto)
			nazivCache[entity.Konto] = naziv
		}

		saldoDug, saldoPot := saldoAcc.saldo(entity)
		fields := []string{
			redniBroj,
			entity.Konto,
			"", // sifra: this list aggregates by konto only
			naziv,
			common.FormatNumberWithSystemLocale(entity.PocStanjeDug, 2),
			common.FormatNumberWithSystemLocale(entity.PocStanjePot, 2),
			common.FormatNumberWithSystemLocale(entity.PrometDug, 2),
			common.FormatNumberWithSystemLocale(entity.PrometPot, 2),
			common.FormatNumberWithSystemLocale(saldoDug, 2),
			common.FormatNumberWithSystemLocale(saldoPot, 2),
		}
		tbl.Rows = append(tbl.Rows, domain.TableRow{ID: marker, Fields: fields, HasUpdate: false, HasDelete: false})
	}

	return nil
}

// buildZakljucniListSubsintetikaQuery builds the single-pass subsintetika hierarchy query with
// the QueryBuilder - the analitika query without the sifra detail level:
//
//	filtered  the fpro transactions of the report (period and konta range filters)
//	agg       GROUPING SETS produce one row per level: subsintetika (konto), sintetika
//	          (LEFT(konto, cfg.NDuzSint)), grupa (LEFT(konto, 2)) and klasa (LEFT(konto, 1))
//	outer     the level is mapped to nivo_order/konto and the result is ordered by the level
//	          aware sort key, so every subtotal row follows the rows it aggregates
//
// Nothing is hardcoded: the period and the konta range come from params, god/kar from the
// session (only when the ledger carries them) and the sintetika prefix length from the
// configuration. The search text matches the hierarchy key (konto).
func (s *BilansiResource) buildZakljucniListSubsintetikaQuery(ctx context.Context, session *domain.UserSession, params domain.ZakljucniParams, printType string) (*querybuilder.Builder, error) {
	duzSin := session.GetDuzSin(s.cfg.NDuzSint)
	filter := querybuilder.New(`SELECT fpro.konto, fnal.tipdok, fpro.kat, fpro.iznos FROM fpro`).
		AddJoin("inner join fnal on fnal.idfnal = fpro.idfnal").
		AddCondition("fpro.konto", params.OdKonta, ">=").
		AddCondition("fpro.konto", params.DoKonta, "<=").
		AddCondition("fnal.danal", params.OdDatuma, ">=").
		AddCondition("fnal.danal", params.DoDatuma, "<=")
	// Klasa 9 accounts are excluded for printing only, exactly like the flattened
	// zakljucniInnerQuery does it.
	if printType == common.TipStampePrint && params.Klasa9 == "false" {
		filter.Where("fpro.konto NOT LIKE '9%'")
	}

	nivoOrderCase := fmt.Sprintf("CASE WHEN k4 IS NOT NULL THEN %d WHEN k3 IS NOT NULL THEN %d WHEN k2 IS NOT NULL THEN %d ELSE %d END AS nivo_order",
		zakljucniNivoAnalitika, zakljucniNivoSintetika, zakljucniNivoGrupa, zakljucniNivoKlasa)
	agg := fmt.Sprintf(`
		SELECT
			LEFT(konto, 1) AS k1,
			LEFT(konto, 2) AS k2,
			LEFT(konto, %d) AS k3,
			konto AS k4,
			SUM(CASE WHEN tipdok = '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS pocstanjedug,
			SUM(CASE WHEN tipdok = '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS pocstanjepot,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS prometdug,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS prometpot
		FROM filtered
		GROUP BY GROUPING SETS (
			(konto),
			(LEFT(konto, %d)),
			(LEFT(konto, 2)),
			(LEFT(konto, 1))
		)
	`, duzSin, duzSin)
	b, err := s.zakljucniHierarchy(ctx, filter, agg, fmt.Sprintf(`SELECT
		nivo_order,
		konto_grp AS konto,
		pocstanjedug, pocstanjepot, prometdug, prometpot
	FROM (
		SELECT
			%s,
			COALESCE(k4, k3, k2, k1) AS konto_grp,
			RPAD(COALESCE(k4, k3, k2, k1), 6, '~') AS sort_key,
			pocstanjedug, pocstanjepot, prometdug, prometpot
		FROM agg
	) sub`, nivoOrderCase))
	if err != nil {
		return nil, err
	}

	// The search text matches the hierarchy key (konto).
	whereZakljucniSearch(b, params.SearchText, "sub.konto")
	b.OrderByExpr(`sort_key COLLATE "C" NULLS LAST`)
	return b, nil
}

// GetZakljucniListZaStampu returns all zakljucni list rows for printing, with the hierarchy
// produced by the same single-pass query the preview uses, so printed and previewed levels can
// never drift apart. TipLista selects the hierarchy:
//
//	"1" analitika:    detail rows (konto + sifra) numbered, then G1=sintetika, G2=grupa, G3=klasa
//	"2" subsintetika: detail rows (konto) numbered, then G1=sintetika, G2=grupa, G3=klasa
//	"3" sintetika:    detail rows (sintetika) numbered, then G1=grupa, G2=klasa
//
// The grand total is the sum of the detail level and tblSummary gets one row per klasa.
func (s *BilansiResource) GetZakljucniListZaStampu(ctx context.Context, tbl, tblSummary *domain.TableData, params domain.ZakljucniParams, nDuzSint int) error {

	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	// nDuzSint (the prefix length the handler resolved for this request) is passed in for explicitness;
	// the hierarchy queries below resolve the same value from the request session
	// (common.NDuzSintFromContext).

	// detailNivo is the level printed as numbered detail rows; every level above it is printed
	// as a G<n> subtotal, exactly like the preview shows them.
	detailNivo := zakljucniDetailNivo(params.TipLista)
	var b *querybuilder.Builder
	var err error
	switch params.TipLista {
	case "1":
		b, err = s.buildZakljucniListAnalitikaQuery(ctx, session, params, common.TipStampePrint)
	case "2":
		b, err = s.buildZakljucniListSubsintetikaQuery(ctx, session, params, common.TipStampePrint)
	case "3":
		b, err = s.buildZakljucniListSintetikaQuery(ctx, session, params, common.TipStampePrint)
	default:
		return fmt.Errorf("invalid tip_lista: %s", params.TipLista)
	}
	if err != nil {
		return err
	}

	entities, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	// ── Row markers ──────────────────────────────
	// Row [0] encodes the row type: the detail rows carry their running number, the subtotal
	// rows carry G1..G<n> (G1 = the innermost subtotal) and "T" is the grand total. The levels
	// come from the query of the selected list, so they are the same ones the preview shows.
	formatRow := func(marker, konto, sifra, naziv string, pstDug, pstPot, promDug, promPot, saldoDug, saldoPot float64) domain.TableRow {
		ukupDug := pstDug + promDug
		ukupPot := pstPot + promPot
		return domain.TableRow{
			ID: marker,
			Fields: []string{
				marker, konto, sifra, naziv,
				common.FormatNumberWithSystemLocale(pstDug, 2),
				common.FormatNumberWithSystemLocale(pstPot, 2),
				common.FormatNumberWithSystemLocale(promDug, 2),
				common.FormatNumberWithSystemLocale(promPot, 2),
				common.FormatNumberWithSystemLocale(ukupDug, 2),
				common.FormatNumberWithSystemLocale(ukupPot, 2),
				common.FormatNumberWithSystemLocale(saldoDug, 2),
				common.FormatNumberWithSystemLocale(saldoPot, 2),
			},
		}
	}

	var grandPstDug, grandPstPot, grandPromDug, grandPromPot float64
	rbr := 1

	// Every subtotal row shows the sum of the netted sides below it, so the saldo columns foot
	// at every level and the printed report shows the same numbers as the grid.
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(common.NDuzSintFromContext(ctx, s.cfg), params.TipLista, *entities)

	nazivCache := make(map[string]string)
	for _, ent := range *entities {
		marker := zakljucniPrintMarker(ent.NivoOrder, detailNivo, rbr)
		if marker == "" {
			continue // unknown level: skip it rather than print a nameless row
		}
		if ent.NivoOrder == detailNivo {
			rbr++
			grandPstDug += ent.PocStanjeDug
			grandPstPot += ent.PocStanjePot
			grandPromDug += ent.PrometDug
			grandPromPot += ent.PrometPot
		}

		// The account names are not part of the aggregation, so they are resolved per unique
		// hierarchy key (cached, so every key is looked up only once).
		naziv, cached := nazivCache[ent.Konto]
		if !cached {
			naziv = s.getKontoNaziv(ctx, ent.Konto)
			nazivCache[ent.Konto] = naziv
		}

		saldoDug, saldoPot := saldoAcc.saldo(ent)
		tbl.Rows = append(tbl.Rows, formatRow(marker, ent.Konto, ent.Sifra, naziv,
			ent.PocStanjeDug, ent.PocStanjePot, ent.PrometDug, ent.PrometPot, saldoDug, saldoPot))

		// The klasa rows also feed the summary table (one line per klasa).
		if ent.NivoOrder == zakljucniNivoKlasa {
			tblSummary.Rows = append(tblSummary.Rows, formatRow("K", ent.Konto, "",
				i18n.GetInstance().Label("klasa")+": "+ent.Konto,
				ent.PocStanjeDug, ent.PocStanjePot, ent.PrometDug, ent.PrometPot, saldoDug, saldoPot))
		}
	}

	// Grand total row: the sum of the netted sides of every detail row, exactly like the total
	// row at the bottom of the grid.
	grandSaldoDug, grandSaldoPot := saldoAcc.total()
	tbl.Rows = append(tbl.Rows, formatRow("T", "", "", i18n.GetInstance().Label("Ukupno"),
		grandPstDug, grandPstPot, grandPromDug, grandPromPot, grandSaldoDug, grandSaldoPot))

	// Grand total for summary.
	tblSummary.Rows = append(tblSummary.Rows, formatRow("T", "", "", "TOTAL:",
		grandPstDug, grandPstPot, grandPromDug, grandPromPot, grandSaldoDug, grandSaldoPot))

	return nil
}

// netirajSaldo netira duguje/potražuje: veća strana se umanjuje za manju, a manja se
// postavlja na 0; ako su jednake, obe strane su 0 (WinDev ekvivalent:
// IF SALDODUG > SALDOPOT THEN ... / IF SALDODUG < SALDOPOT THEN ... / IF = THEN 0/0).
func netirajSaldo(dug, pot float64) (float64, float64) {
	switch {
	case dug > pot:
		return dug - pot, 0
	case pot > dug:
		return 0, pot - dug
	default:
		return 0, 0
	}
}

// zakljucniHierarchy builds the single-pass hierarchy query of the Zakljucni list:
// WITH filtered AS (filter, restricted to the period of the session), agg AS (agg) outer.
// agg and outer are static SQL; the sintetika prefix length is written into them as an int
// (the GROUPING SETS must repeat the very same expression as the SELECT, and two placeholders
// would not be the same expression).
func (s *BilansiResource) zakljucniHierarchy(ctx context.Context, filter *querybuilder.Builder, agg, outer string) (*querybuilder.Builder, error) {
	if err := s.fproRepo.ApplyFilters(ctx, filter, "fpro", nil, ""); err != nil {
		return nil, err
	}
	return querybuilder.New(outer).
		WithCTEFrom("filtered", filter).
		WithCTE("agg", agg), nil
}

// whereZakljucniSearch adds the OR-ed ILIKE filter of the Zakljucni list search box: one of
// the fields contains the text (nothing when the text is empty).
func whereZakljucniSearch(b *querybuilder.Builder, text string, fields ...string) {
	if text == "" {
		return
	}
	conds := make([]querybuilder.Cond, 0, len(fields))
	for _, f := range fields {
		conds = append(conds, querybuilder.C(f+" ILIKE '%' || ? || '%'", text))
	}
	b.WhereAny(conds...)
}

// setZakljucniTotals fills the pagination counters and the totals row (columns 4..9) of the
// Zakljucni list grid.
//
// entities are the rows of the level that is summed - the accounts of the list - so the totals
// count every transaction once. totalRows is the number of rows the whole hierarchy renders
// (accounts + subtotals) and is the number the pager counts, so every row of the list can be
// reached with it.
func setZakljucniTotals(tbl *domain.TableData, entities []domain.FproDto, pageSize, totalRows int) {
	common.SetTableTotalRecords(tbl, totalRows, pageSize)
	if len(tbl.Headers) <= 9 {
		return
	}

	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column

	var pstDug, pstPot, promDug, promPot, saldoDug, saldoPot float64
	for _, e := range entities {
		pstDug += e.PocStanjeDug
		pstPot += e.PocStanjePot
		promDug += e.PrometDug
		promPot += e.PrometPot
		d, p := netirajSaldo(e.PocStanjeDug+e.PrometDug, e.PocStanjePot+e.PrometPot)
		saldoDug += d
		saldoPot += p
	}

	tbl.Totals[4] = common.FormatNumberWithSystemLocale(pstDug, 2)             // Početno stanje duguje
	tbl.Totals[5] = common.FormatNumberWithSystemLocale(pstPot, 2)             // Početno stanje potražuje
	tbl.Totals[6] = common.FormatNumberWithSystemLocale(promDug, 2)            // Promet duguje
	tbl.Totals[7] = common.FormatNumberWithSystemLocale(promPot, 2)            // Promet potražuje
	tbl.Totals[8] = common.FormatNumberWithSystemLocale(math.Abs(saldoDug), 2) // Saldo duguje
	tbl.Totals[9] = common.FormatNumberWithSystemLocale(math.Abs(saldoPot), 2) // Saldo potražuje
}

// GetBilansStanja retrieves data for Bilans stanja (balance sheet)
func (s *BilansiResource) GetBilansStanja(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, searchText string, skraceni bool) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	// Get Totals for table
	bTotals := querybuilder.New(`SELECT
			COALESCE(SUM(bils.tgod), 0) as tgod,
			COALESCE(SUM(bils.tgodh), 0) as tgodh,
			COALESCE(SUM(bils.pgod), 0) as pgod,
			COALESCE(SUM(bils.pgodh), 0) as pgodh,
			COALESCE(SUM(bils.pgodps), 0) as pgodps,
			COALESCE(SUM(bils.pgodhps), 0) as pgodhps
			FROM bils`)

	b := querybuilder.New(`SELECT
			bils.bilsid, bils.rbr, bils.grac, bils.nazp, bils.aop,
			bils.konta, bils.tgod, bils.pgod,
			bils.nipo, bils.tgodh, bils.pgodh,
			bils.pozic_1, bils.pozic_2, bils.pozic_3, bils.pozic_4,
			bils.pozic_5, bils.pozic_6, bils.pozic_7, bils.pozic_8,
			bils.pozic_9, bils.pozic_10, bils.pozic_11, bils.pozic_12, bils.skraceni FROM bils`)

	// Add same filters to both queries
	for _, x := range []*querybuilder.Builder{b, bTotals} {
		if err := s.bilsRepo.ApplyFilters(ctx, x, "bils", nil, ""); err != nil {
			return err
		}
	}

	totalsResult, err := s.bilsRepo.Select(ctx, bTotals)
	if err != nil {
		return err
	}
	if totalsResult != nil && len(*totalsResult) > 0 {
		totals.TekGod = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGod, 2)
		totals.PrethGod = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGod, 2)
		totals.TekGodH = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGodH, 0)
		totals.PrethGodH = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodH, 0)
		totals.PocStanje = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodPS, 2)
		totals.PocStanjeH = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodHPS, 0)
	}

	// Add search filter if provided (also to the totals of the grid below)
	whereBilansSearch(b, "bils", searchText)
	whereBilansSearch(bTotals, "bils", searchText)

	entities, err := s.bilsRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			//Ensure non-negative values
			tgod := float64(0)
			if entity.TGod > 0 {
				tgod = entity.TGod
			}
			pgod := float64(0)
			if entity.PGod > 0 {
				pgod = entity.PGod
			}
			pgodps := float64(0)
			if entity.PGodPS > 0 {
				pgodps = entity.PGodPS
			}
			fields := []string{
				fmt.Sprintf("%d", entity.Rbr),
				entity.Grac,
				entity.NazP,
				fmt.Sprintf("%04d", entity.AOP),
				entity.Konta,
				common.FormatNumberWithSystemLocale(float64(tgod), 2),
				common.FormatNumberWithSystemLocale(float64(pgod), 2),
				common.FormatNumberWithSystemLocale(float64(pgodps), 2),
				fmt.Sprintf("%d", entity.NiPo),
				common.FormatNumberWithSystemLocale(float64(tgod/1000), 2),
				common.FormatNumberWithSystemLocale(float64(pgod/1000), 2),
				common.FormatNumberWithSystemLocale(float64(pgodps/1000), 2),
				fmt.Sprintf("%04d", entity.Pozic1),
				fmt.Sprintf("%04d", entity.Pozic2),
				fmt.Sprintf("%04d", entity.Pozic3),
				fmt.Sprintf("%04d", entity.Pozic4),
				fmt.Sprintf("%04d", entity.Pozic5),
				fmt.Sprintf("%04d", entity.Pozic6),
				fmt.Sprintf("%04d", entity.Pozic7),
				fmt.Sprintf("%04d", entity.Pozic8),
				fmt.Sprintf("%04d", entity.Pozic9),
				fmt.Sprintf("%04d", entity.Pozic10),
				fmt.Sprintf("%04d", entity.Pozic11),
				fmt.Sprintf("%04d", entity.Pozic12),
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", entity.BilsID), Fields: fields, HasUpdate: true, HasDelete: true})
		}
	}
	// Execute query and get entities
	entitiesTotal, err := s.bilsRepo.Select(ctx, bTotals)
	if err != nil {
		return err
	}
	if len(*entitiesTotal) > 0 {
		entity := (*entitiesTotal)[0]
		if len(tbl.Headers) > 10 {
			// Set totals in header if needed
			tbl.Totals = make([]string, len(tbl.Headers))
			tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")                          // Set label for totals column
			tbl.Totals[5] = common.FormatNumberWithSystemLocale(entity.TGod, 2)         // Tekuća godina
			tbl.Totals[6] = common.FormatNumberWithSystemLocale(entity.PGod, 2)         // Prethodna godina
			tbl.Totals[7] = common.FormatNumberWithSystemLocale(entity.PGodPS, 2)       // Prethodna godina - početno stanje
			tbl.Totals[9] = common.FormatNumberWithSystemLocale(entity.TGod/1000, 2)    // Tekuća godina u hiljadama
			tbl.Totals[10] = common.FormatNumberWithSystemLocale(entity.PGod/1000, 2)   // Prethodna godina u hiljadama
			tbl.Totals[11] = common.FormatNumberWithSystemLocale(entity.PGodPS/1000, 2) // Prethodna godina - početno stanje u hiljadama
		}
	}

	return nil
}

// whereGodKar restricts the query to the given period when the table carries god/kar. The
// bilansi obrada passes the period explicitly (it also reads the previous year), so this is
// not always the period of the session. prefix is "" or an alias with a dot ("fpro.").
func whereGodKar(b *querybuilder.Builder, prefix string, hasGod, hasKar bool, god, kar int) {
	if hasGod {
		b.Where(prefix+"god = ?", god)
	}
	if hasKar {
		b.Where(prefix+"kar = ?", kar)
	}
}

// whereMesecOdDo restricts danal to the months od-do (each bound only when > 0).
func whereMesecOdDo(b *querybuilder.Builder, odMes, doMes int) {
	if odMes > 0 {
		b.Where("extract(month from danal)::int >= ?", odMes)
	}
	if doMes > 0 {
		b.Where("extract(month from danal)::int <= ?", doMes)
	}
}

// whereBilansSearch adds the search of the bilans grids: naziv pozicije, grupa racuna, konta or
// AOP contains the text (table is "bils" or "bilu"; nothing when the text is empty). The old
// condition was invalid SQL (it used the bilu alias on bils, stray quotes and ILIKE on the
// integer aop), so a search always failed.
func whereBilansSearch(b *querybuilder.Builder, table, text string) {
	if text == "" {
		return
	}
	pattern := "%" + text + "%"
	b.WhereAny(
		querybuilder.C(table+".nazp ILIKE ?", pattern),
		querybuilder.C(table+".grac ILIKE ?", pattern),
		querybuilder.C(table+".konta ILIKE ?", pattern),
		querybuilder.C(table+".aop::text ILIKE ?", pattern),
	)
}

// GetBilansStanjaObrada processes and retrieves Bilans stanja (balance sheet) data.
// Translated from WinDev ObradaBILS: resets BILS totals, recalculates from fpro via
// obradaKontaBilsCached for leaf rows (Konta != ""), aggregates via POZIC[1..12] for
// summary rows, then populates tbl with final results.
//
// Konta entry format for BILS: "[+/-][D/P/S][konto]" e.g. "+D10", "-S204"
//
//	cZnak (±)    = add or subtract the account
//	cDugPot(D/P/S) = debit-only / credit-only / net balance side
//
// dVred   = Abs(openingBalance + monthlyMovements) → TGOD
// dVredPS = Abs(openingBalance)                    → PGOD when lPGODizPS=true
//
// Performance: all BILS rows loaded into memory once (keyed by AOP); the salda of all the konto
// prefixes read from fpro in one query (fproBilsSalda) and all the computed rows written by one
// statement (BilsUpisiIznose), instead of two fpro scans a prefix and one UPDATE a row.
func (s *BilansiResource) GetBilansStanjaObrada(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, stanjeNaDan string, skraceni bool, lPGODizPS bool) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.bilsRepo.GetHasGodHasKar()
	god := userSession.SelectedGod
	kar := userSession.SelectedKar

	// OdMES is always 1 per WinDev ObradaKONTA logic; DoMES = month of stanjeNaDan
	odMes := 1
	doMes := 12
	if t, err := time.Parse("2006-01-02", stanjeNaDan); err == nil {
		doMes = int(t.Month())
	}

	// ── STEP 1: Reset totals ────────────────────────────────────────────────────
	// All rows: tgod=0, tgodh=0
	resetAll := querybuilder.NewUpdate("bils").Set("tgod = 0, tgodh = 0")
	whereGodKar(resetAll, "", hasGod, hasKar, god, kar)
	if err := s.bilsRepo.Exec(ctx, resetAll); err != nil {
		return err
	}

	// Summary rows (konta=''): also reset pgod, pgodh, pgodps, pgodhps
	resetSummary := querybuilder.NewUpdate("bils").Set("pgod = 0, pgodh = 0, pgodps = 0, pgodhps = 0")
	whereGodKar(resetSummary, "", hasGod, hasKar, god, kar)
	resetSummary.Where("konta = ''")
	if err := s.bilsRepo.Exec(ctx, resetSummary); err != nil {
		return err
	}

	// Leaf rows + lPGODizPS=true: also reset pgod, pgodh
	if lPGODizPS {
		resetLeaf := querybuilder.NewUpdate("bils").Set("pgod = 0, pgodh = 0")
		whereGodKar(resetLeaf, "", hasGod, hasKar, god, kar)
		resetLeaf.Where("konta != ''")
		if err := s.bilsRepo.Exec(ctx, resetLeaf); err != nil {
			return err
		}
	}

	// ── STEP 2: Get max NIPO ────────────────────────────────────────────────────
	maxNipoB := querybuilder.New(`select coalesce(max(nipo), 1) as nipo from bils`)
	whereGodKar(maxNipoB, "", hasGod, hasKar, god, kar)
	maxNipoRecords, err := s.bilsRepo.Select(ctx, maxNipoB)
	if err != nil {
		return err
	}
	maxk := int16(1)
	if maxNipoRecords != nil && len(*maxNipoRecords) > 0 && (*maxNipoRecords)[0].NiPo > 1 {
		maxk = (*maxNipoRecords)[0].NiPo
	}

	// ── STEP 3: Load ALL bils rows into memory once, keyed by AOP ──────────────
	allBilsB := querybuilder.New(`select bilsid, aop, konta, tgod, pgod, pgodps, tgodh, pgodh, pgodhps, nipo, skraceni,
		pozic_1, pozic_2, pozic_3, pozic_4, pozic_5, pozic_6,
		pozic_7, pozic_8, pozic_9, pozic_10, pozic_11, pozic_12 from bils`)
	whereGodKar(allBilsB, "", hasGod, hasKar, god, kar)
	allBilsRecords, err := s.bilsRepo.Select(ctx, allBilsB)
	if err != nil {
		return err
	}

	bilsByAop := make(map[int]*domain.Bils, len(*allBilsRecords))
	for i := range *allBilsRecords {
		row := &(*allBilsRecords)[i]
		bilsByAop[row.AOP] = row
	}

	// The salda of every konto prefix of the konta of the rows, read from fpro in one query (the legacy read FSAL
	// once for every prefix of every row). The aggregates read fpro, so they are scoped by fpro's own god/kar columns
	// (not by bils').
	prefiksi := []string{}
	vidjeni := map[string]bool{}
	for i := range *allBilsRecords {
		s.obradaKontaBilsCached((*allBilsRecords)[i].Konta, god, kar, odMes, doMes, func(_, _, _, _ int, sKonto string) (float64, float64, float64, float64) {
			if !vidjeni[sKonto] {
				vidjeni[sKonto] = true
				prefiksi = append(prefiksi, sKonto)
			}
			return 0, 0, 0, 0
		})
	}
	fproGod, fproKar := s.fproRepo.GetHasGodHasKar()
	salda, err := s.fproBilsSalda(ctx, fproGod, fproKar, god, kar, odMes, doMes, prefiksi)
	if err != nil {
		return fmt.Errorf("obrada bilansa stanja: %w", err)
	}
	saldoPrefiksa := func(_, _, _, _ int, sKonto string) (float64, float64, float64, float64) {
		v := salda[sKonto]
		return v.pDug, v.pPot, v.mDug, v.mPot
	}

	// ── STEP 4: Process each NIPO level ────────────────────────────────────────
	// The computed iznosi of the rows, written in one statement after the last level.
	var iznosi []domain.BilsIznosi
	for k := int16(1); k <= maxk; k++ {
		var levelRows []*domain.Bils
		for i := range *allBilsRecords {
			if (*allBilsRecords)[i].NiPo == k {
				levelRows = append(levelRows, &(*allBilsRecords)[i])
			}
		}

		for _, bils := range levelRows {
			if bils.Konta != "" {
				// Leaf row: calculate from fpro via cached aggregate
				dVred, dVredPS := s.obradaKontaBilsCached(bils.Konta, god, kar, odMes, doMes, saldoPrefiksa)
				bils.TGod = math.Round(dVred)
				bils.TGodH = int64(math.Round(dVred / 1000))
				if lPGODizPS {
					bils.PGod = math.Round(dVredPS)
					bils.PGodH = int64(math.Round(dVredPS / 1000))
				}
			} else {
				// Summary row: aggregate from referenced BILS rows via POZIC[1..12]
				pozici := [12]int16{
					bils.Pozic1, bils.Pozic2, bils.Pozic3, bils.Pozic4,
					bils.Pozic5, bils.Pozic6, bils.Pozic7, bils.Pozic8,
					bils.Pozic9, bils.Pozic10, bils.Pozic11, bils.Pozic12,
				}
				for _, pozic := range pozici {
					if pozic == 0 {
						continue
					}
					nAOP := int(pozic)
					if nAOP < 0 {
						nAOP = -nAOP
					}
					rel, found := bilsByAop[nAOP]
					if !found {
						continue
					}
					if pozic > 0 {
						bils.TGod += rel.TGod
						bils.TGodH += rel.TGodH
						bils.PGod += rel.PGod
						bils.PGodH += rel.PGodH
						bils.PGodPS += rel.PGodPS
						bils.PGodHPS += rel.PGodHPS
					} else {
						bils.TGod -= rel.TGod
						bils.TGodH -= rel.TGodH
						bils.PGod -= rel.PGod
						bils.PGodH -= rel.PGodH
						bils.PGodPS -= rel.PGodPS
						bils.PGodHPS -= rel.PGodHPS
					}
				}
			}

			// Clamp all fields to non-negative
			if bils.TGod < 0 {
				bils.TGod = 0
			}
			if bils.TGodH < 0 {
				bils.TGodH = 0
			}
			if bils.PGod < 0 {
				bils.PGod = 0
			}
			if bils.PGodH < 0 {
				bils.PGodH = 0
			}
			if bils.PGodPS < 0 {
				bils.PGodPS = 0
			}
			if bils.PGodHPS < 0 {
				bils.PGodHPS = 0
			}

			iznosi = append(iznosi, domain.BilsIznosi{BilsID: bils.BilsID, TGod: bils.TGod, TGodH: bils.TGodH, PGod: bils.PGod,
				PGodH: bils.PGodH, PGodPS: bils.PGodPS, PGodHPS: bils.PGodHPS})
		}
	}
	if err := s.fin.BilsUpisiIznose(ctx, nil, iznosi); err != nil {
		return fmt.Errorf("obrada bilansa stanja: %w", err)
	}

	// ── STEP 5: Fetch final results and populate table ──────────────────────────
	finalB := querybuilder.New(`select bilsid, rbr, grac, nazp, aop, konta, napomena,
		tgod, pgod, pgodps, tgodh, pgodh, pgodhps, nipo, skraceni from bils`)
	whereGodKar(finalB, "", hasGod, hasKar, god, kar)
	if skraceni {
		finalB.AddEqual("skraceni", 1)
	}
	finalB.OrderBy("aop asc")

	finalRecords, err := s.bilsRepo.Select(ctx, finalB)
	if err != nil {
		return err
	}

	var sumTGod, sumPGod, sumPGodPS float64
	var sumTGodH, sumPGodH, sumPGodHPS int64
	if finalRecords != nil {
		for _, bils := range *finalRecords {
			tgod, pgod, pgodps := bils.TGod, bils.PGod, bils.PGodPS
			if tgod < 0 {
				tgod = 0
			}
			if pgod < 0 {
				pgod = 0
			}
			if pgodps < 0 {
				pgodps = 0
			}
			sumTGod += tgod
			sumPGod += pgod
			sumPGodPS += pgodps
			sumTGodH += bils.TGodH
			sumPGodH += bils.PGodH
			sumPGodHPS += bils.PGodHPS
			fields := []string{
				bils.Grac,
				bils.NazP,
				fmt.Sprintf("%04d", bils.AOP),
				bils.Napomena,
				common.FormatNumberWithSystemLocale(tgod, 2),
				common.FormatNumberWithSystemLocale(pgod, 2),
				common.FormatNumberWithSystemLocale(pgodps, 2),
				common.FormatNumberWithSystemLocale(float64(bils.TGodH), 0),
				common.FormatNumberWithSystemLocale(float64(bils.PGodH), 0),
				common.FormatNumberWithSystemLocale(float64(bils.PGodHPS), 0),
				fmt.Sprintf("%d", bils.NiPo),
				fmt.Sprintf("%d", bils.Skraceni),
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", bils.BilsID), Fields: fields, HasUpdate: false, HasDelete: false})
		}
	}

	totals.TekGod = common.FormatNumberWithSystemLocale(sumTGod, 2)
	totals.PrethGod = common.FormatNumberWithSystemLocale(sumPGod, 2)
	totals.TekGodH = common.FormatNumberWithSystemLocale(float64(sumTGodH), 0)
	totals.PrethGodH = common.FormatNumberWithSystemLocale(float64(sumPGodH), 0)
	totals.PocStanje = common.FormatNumberWithSystemLocale(sumPGodPS, 2)
	totals.PocStanjeH = common.FormatNumberWithSystemLocale(float64(sumPGodHPS), 0)
	bilansiUkupno(tbl, bilansiSuma{4, sumTGod, 2}, bilansiSuma{5, sumPGod, 2}, bilansiSuma{6, sumPGodPS, 2},
		bilansiSuma{7, float64(sumTGodH), 0}, bilansiSuma{8, float64(sumPGodH), 0}, bilansiSuma{9, float64(sumPGodHPS), 0})

	return nil
}

// obradaKontaBilsCached computes (dVred, dVredPS) for a BILS konta list.
// Konta entry format: "[+/-][D/P/S][konto]" e.g. "+D10", "-S204", "+P12"
//
//	cZnak   (±)    = add or subtract the account
//	cDugPot (D/P/S) = debit-only / credit-only / net balance
//	sKonto         = account prefix (2, 3, or >3 chars)
//
// Returns:
//
//	dVred   = Abs(openingSaldo + monthlySaldo)  → used for TGOD
//	dVredPS = Abs(openingSaldo)                 → used for PGOD when lPGODizPS=true
func (s *BilansiResource) obradaKontaBilsCached(
	konta string, god, kar, odMes, doMes int,
	cachedAggregate func(god, kar, odMes, doMes int, sKonto string) (pDug, pPot, mDug, mPot float64),
) (dVred, dVredPS float64) {
	if konta == "" {
		return 0, 0
	}

	var xMSaldo, xPSaldo float64

	for _, entry := range strings.Split(konta, ";") {
		entry = strings.TrimSpace(entry)
		if len(entry) < 3 {
			continue
		}
		cZnak := string(entry[0])
		cDugPot := strings.ToLower(string(entry[1]))
		sKonto := entry[2:]

		if (cZnak != "+" && cZnak != "-") || (cDugPot != "d" && cDugPot != "p" && cDugPot != "s") {
			continue
		}

		nPlusMinus := 1.0
		if cZnak == "-" {
			nPlusMinus = -1.0
		}

		pDug, pPot, mDug, mPot := cachedAggregate(god, kar, odMes, doMes, sKonto)

		// Apply sign multiplier (mirrors WinDev nPlusMinus * FSAL.MDUG[i])
		xMDug := mDug * nPlusMinus
		xMPot := mPot * nPlusMinus
		xPDug := pDug * nPlusMinus
		xPPot := pPot * nPlusMinus

		// Apply cDugPot: D=debit side, P=credit side, S=net balance
		switch cDugPot {
		case "d":
			xMSaldo += xMDug
			xPSaldo += xPDug
		case "p":
			xMSaldo += xMPot
			xPSaldo += xPPot
		case "s":
			xMSaldo += xMDug - xMPot
			xPSaldo += xPDug - xPPot
		}
	}

	// dVred   = Abs(opening + movements) — total for current year column
	// dVredPS = Abs(opening)             — opening balance only (previous year / PS column)
	dVred = math.Abs(xPSaldo + xMSaldo)
	dVredPS = math.Abs(xPSaldo)
	return
}

// fproBilsSaldo are the sums of the stavke of fpro of a konto prefix: the početno stanje (tipdok '00') and the promet
// of the months (the other tipdok), duguje (kat 1, 2) and potražuje (kat 3, 4).
type fproBilsSaldo struct{ pDug, pPot, mDug, mPot float64 }

// fproBilsSalda returns the salda of the konto prefixes (the konta of the rows of the bilans stanja) from fpro in one
// query: the stavke of the period summed by konto, then the konta of every prefix summed (a konto matches when it
// begins with the prefix, ignoring the case, like the former ILIKE 'prefix%'); the promet only of the months
// [odMes..doMes] of danal (a bound <= 0 is not applied). A prefix without stavke is not in the map (its salda are 0).
// The former two queries a prefix each scanned fpro (an ILIKE prefix can not use an index).
func (s *BilansiResource) fproBilsSalda(ctx context.Context, hasGod, hasKar bool, god, kar, odMes, doMes int, prefiksi []string) (map[string]fproBilsSaldo, error) {
	salda := make(map[string]fproBilsSaldo, len(prefiksi))
	if len(prefiksi) == 0 {
		return salda, nil
	}
	// The months of the promet (the conditions of whereMesecOdDo).
	mesec, mesecArgs := "true", []any{}
	if odMes > 0 {
		mesec += " and extract(month from fpro.danal)::int >= ?"
		mesecArgs = append(mesecArgs, odMes)
	}
	if doMes > 0 {
		mesec += " and extract(month from fpro.danal)::int <= ?"
		mesecArgs = append(mesecArgs, doMes)
	}
	args := append(append([]any{}, mesecArgs...), mesecArgs...)
	konta := querybuilder.New(`select fpro.konto,
		sum(case when fpro.tipdok = '00' and fpro.kat in (1,2) then fpro.iznos else 0 end) as pdug,
		sum(case when fpro.tipdok = '00' and fpro.kat in (3,4) then fpro.iznos else 0 end) as ppot,
		sum(case when fpro.tipdok != '00' and fpro.kat in (1,2) and `+mesec+` then fpro.iznos else 0 end) as mdug,
		sum(case when fpro.tipdok != '00' and fpro.kat in (3,4) and `+mesec+` then fpro.iznos else 0 end) as mpot
		from fpro`, args...)
	whereGodKar(konta, "fpro.", hasGod, hasKar, god, kar)
	konta.Where("fpro.konto is not null").GroupBy("fpro.konto")

	pArgs := make([]any, len(prefiksi))
	for i, p := range prefiksi {
		pArgs[i] = p
	}
	b := querybuilder.New(`select p.prefiks as konto,
		coalesce(sum(k.pdug), 0) as pocstanjedug, coalesce(sum(k.ppot), 0) as pocstanjepot,
		coalesce(sum(k.mdug), 0) as prometdug, coalesce(sum(k.mpot), 0) as prometpot
		from (values `+strings.TrimSuffix(strings.Repeat("(?::text), ", len(prefiksi)), ", ")+`) p(prefiks)
		join k on left(lower(k.konto), length(p.prefiks)) = lower(p.prefiks)`, pArgs...).
		WithCTEFrom("k", konta).
		GroupBy("p.prefiks")
	rows, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("salda konta bilansa stanja: %w", err)
	}
	if rows != nil {
		for _, r := range *rows {
			salda[r.Konto] = fproBilsSaldo{pDug: r.PocStanjeDug, pPot: r.PocStanjePot, mDug: r.PrometDug, mPot: r.PrometPot}
		}
	}
	return salda, nil
}

// bilansiSuma is a total of a column of a bilans table: the index of the column and the decimals of its rows.
type bilansiSuma struct {
	kolona   int
	iznos    float64
	decimale int
}

// bilansiUkupno sets the totals row at the end of a bilans table: "Ukupno" under the first column and the sums under
// their columns (a column the table has not is skipped).
func bilansiUkupno(tbl *domain.TableData, sume ...bilansiSuma) {
	tbl.HasTotals = true
	tbl.Totals = make([]string, len(tbl.Headers))
	if len(tbl.Totals) == 0 {
		return
	}
	tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")
	for _, suma := range sume {
		if suma.kolona < len(tbl.Totals) {
			tbl.Totals[suma.kolona] = common.FormatNumberWithSystemLocale(suma.iznos, suma.decimale)
		}
	}
}

// GetBilansStanjaZaStampu retrieves data for Bilans stanja (balance sheet) for printing
func (s *BilansiResource) GetBilansStanjaZaStampu(ctx context.Context, tbl *domain.TableData, tipStampe string, skraceni bool) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	b := querybuilder.New(`SELECT
			bils.bilsid, bils.rbr, bils.grac, bils.nazp, bils.aop,
			bils.napomena, bils.tgodh, bils.pgodh, bils.pgodhps,
			bils.tgod, bils.pgod, bils.pgodps, bils.nipo, bils.skraceni FROM bils`)

	if err := s.bilsRepo.ApplyFilters(ctx, b, "bils", nil, ""); err != nil {
		return err
	}
	if skraceni {
		b.AddEqual("bils.skraceni", 1)
	}
	// The rows of the print by AOP (like the obrada and the legacy, ORDER BY AOP).
	b.OrderBy("bils.aop ASC")

	entities, err := s.bilsRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	// The totals at the end of the table: the sums of the shown values.
	var sumTGod, sumPGod, sumPGodPS, sumTGodH, sumPGodH, sumPGodHPS float64
	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			//Ensure non-negative values
			tgod := float64(0)
			if entity.TGod > 0 {
				tgod = entity.TGod
			}
			pgod := float64(0)
			if entity.PGod > 0 {
				pgod = entity.PGod
			}
			pgodps := float64(0)
			if entity.PGodPS > 0 {
				pgodps = entity.PGodPS
			}
			sumTGod, sumPGod, sumPGodPS = sumTGod+tgod, sumPGod+pgod, sumPGodPS+pgodps
			sumTGodH, sumPGodH, sumPGodHPS = sumTGodH+float64(entity.TGodH), sumPGodH+float64(entity.PGodH), sumPGodHPS+float64(entity.PGodHPS)
			fields := []string{}
			if tipStampe == common.TipStampePreview {
				fields = []string{
					entity.Grac,
					entity.NazP,
					fmt.Sprintf("%04d", entity.AOP),
					entity.Napomena,
					common.FormatNumberWithSystemLocale(tgod, 2),
					common.FormatNumberWithSystemLocale(pgod, 2),
					common.FormatNumberWithSystemLocale(pgodps, 2),
					common.FormatNumberWithSystemLocale(entity.TGodH, 2),
					common.FormatNumberWithSystemLocale(entity.PGodH, 2),
					common.FormatNumberWithSystemLocale(entity.PGodHPS, 2),
					fmt.Sprintf("%d", entity.NiPo),
					fmt.Sprintf("%d", entity.Skraceni),
				}
			}
			if tipStampe == common.TipStampePrint {
				fields = []string{
					entity.Grac,
					entity.NazP,
					fmt.Sprintf("%04d", entity.AOP),
					entity.Napomena,
					common.FormatNumberWithSystemLocale(entity.TGodH, 2),
					common.FormatNumberWithSystemLocale(entity.PGodH, 2),
					common.FormatNumberWithSystemLocale(entity.PGodHPS, 2),
				}
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", entity.BilsID), Fields: fields, HasUpdate: false, HasDelete: false})
		}
	}
	if tipStampe == common.TipStampePrint {
		bilansiUkupno(tbl, bilansiSuma{4, sumTGodH, 2}, bilansiSuma{5, sumPGodH, 2}, bilansiSuma{6, sumPGodHPS, 2})
	} else {
		bilansiUkupno(tbl, bilansiSuma{4, sumTGod, 2}, bilansiSuma{5, sumPGod, 2}, bilansiSuma{6, sumPGodPS, 2},
			bilansiSuma{7, sumTGodH, 2}, bilansiSuma{8, sumPGodH, 2}, bilansiSuma{9, sumPGodHPS, 2})
	}
	return nil
}
func (s *BilansiResource) GetBilansUspeha(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, searchText string, skraceni bool) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	// If only totals are needed, we can sum directly in SQL without fetching all records
	bTotals := querybuilder.New(`SELECT
			COALESCE(SUM(bilu.tgod), 0) as tgod,
			COALESCE(SUM(bilu.tgodh), 0) as tgodh,
			COALESCE(SUM(bilu.pgod), 0) as pgod,
			COALESCE(SUM(bilu.pgodh), 0) as pgodh
			FROM bilu`)

	b := querybuilder.New(`SELECT
			bilu.biluid, bilu.rbr, bilu.grac, bilu.nazp, bilu.aop,
			bilu.konta, bilu.tgod, bilu.pgod,
			bilu.nipo, bilu.tgodh, bilu.pgodh,
			bilu.pozic_1, bilu.pozic_2, bilu.pozic_3, bilu.pozic_4,
			bilu.pozic_5, bilu.pozic_6, bilu.pozic_7, bilu.pozic_8,
			bilu.pozic_9, bilu.pozic_10, bilu.pozic_11, bilu.pozic_12, bilu.skraceni FROM bilu`)

	// Add filters directly in query
	for _, x := range []*querybuilder.Builder{b, bTotals} {
		if err := s.biluRepo.ApplyFilters(ctx, x, "bilu", nil, ""); err != nil {
			return err
		}
		// Filter by skraceni flag if set
		if skraceni {
			x.AddEqual("bilu.skraceni", 1)
		}
	}
	totalsResult, err := s.biluRepo.Select(ctx, bTotals)
	if err != nil {
		return err
	}
	if totalsResult != nil && len(*totalsResult) > 0 {
		totals.TekGod = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGod, 2)
		totals.PrethGod = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGod, 2)
		totals.TekGodH = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGodH, 0)
		totals.PrethGodH = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodH, 0)
	}

	// Add search filter if provided (also to the totals of the grid below)
	whereBilansSearch(b, "bilu", searchText)
	whereBilansSearch(bTotals, "bilu", searchText)

	// The rows by AOP.
	b.OrderBy("bilu.aop ASC")

	// Execute query and get entities
	entities, err := s.biluRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			//Ensure non-negative values
			tgod := float64(0)
			if entity.TGod > 0 {
				tgod = entity.TGod
			}
			pgod := float64(0)
			if entity.PGod > 0 {
				pgod = entity.PGod
			}
			fields := []string{
				fmt.Sprintf("%d", entity.Rbr),
				entity.Grac,
				entity.NazP,
				fmt.Sprintf("%04d", entity.AOP),
				entity.Konta,
				common.FormatNumberWithSystemLocale(float64(tgod), 2),
				common.FormatNumberWithSystemLocale(float64(pgod), 2),
				fmt.Sprintf("%d", entity.NiPo),
				common.FormatNumberWithSystemLocale(float64(tgod/1000), 2),
				common.FormatNumberWithSystemLocale(float64(pgod/1000), 2),
				fmt.Sprintf("%04d", entity.Pozic1),
				fmt.Sprintf("%04d", entity.Pozic2),
				fmt.Sprintf("%04d", entity.Pozic3),
				fmt.Sprintf("%04d", entity.Pozic4),
				fmt.Sprintf("%04d", entity.Pozic5),
				fmt.Sprintf("%04d", entity.Pozic6),
				fmt.Sprintf("%04d", entity.Pozic7),
				fmt.Sprintf("%04d", entity.Pozic8),
				fmt.Sprintf("%04d", entity.Pozic9),
				fmt.Sprintf("%04d", entity.Pozic10),
				fmt.Sprintf("%04d", entity.Pozic11),
				fmt.Sprintf("%04d", entity.Pozic12),
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", entity.BiluID), Fields: fields, HasUpdate: true, HasDelete: true})
		}
	}
	// Execute query and get entities
	entitiesTotal, err := s.bilsRepo.Select(ctx, bTotals)
	if err != nil {
		return err
	}
	if len(*entitiesTotal) > 0 {
		entity := (*entitiesTotal)[0]
		if len(tbl.Headers) > 10 {
			// Set totals in header if needed
			tbl.Totals = make([]string, len(tbl.Headers))
			tbl.Totals[0] = i18n.GetInstance().Label("Ukupno")                       // Set label for totals column
			tbl.Totals[5] = common.FormatNumberWithSystemLocale(entity.TGod, 2)      // Tekuća godina
			tbl.Totals[6] = common.FormatNumberWithSystemLocale(entity.PGod, 2)      // Prethodna godina
			tbl.Totals[8] = common.FormatNumberWithSystemLocale(entity.TGod/1000, 2) // Tekuća godina u hiljadama
			tbl.Totals[9] = common.FormatNumberWithSystemLocale(entity.PGod/1000, 2) // Prethodna godina u hiljadama
		}
	}

	return nil
}

// GetBilansUspehaObrada processes and retrieves Bilans uspeha (income statement) data for printing.
// Translated from WinDev: resets BILU totals, recalculates from fpro via obradaKonta for leaf rows,
// aggregates via POZIC[1..12] for summary rows, then populates tbl with final results.
//
// Performance: all BILU rows are loaded once into an in-memory map keyed by AOP,
// eliminating the N×12 per-row DB lookups that caused 20-30s response times.
// fpro aggregation results are cached per unique konta prefix within a single call.
func (s *BilansiResource) GetBilansUspehaObrada(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, odDatuma, doDatuma string, skraceni, lPGODizPG bool) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	// Parse date range into months for obradaKonta
	odMes, doMes := 1, 12
	if t, err := time.Parse("2006-01-02", odDatuma); err == nil {
		odMes = int(t.Month())
	}
	if t, err := time.Parse("2006-01-02", doDatuma); err == nil {
		doMes = int(t.Month())
	}

	hasGod, hasKar := s.biluRepo.GetHasGodHasKar()
	god := userSession.SelectedGod
	kar := userSession.SelectedKar

	// STEP 1: Reset TGOD, TGODH for all BILU records; also reset PGOD, PGODH for summary rows.
	resetAll := querybuilder.NewUpdate("bilu").Set("tgod = 0, tgodh = 0")
	whereGodKar(resetAll, "", hasGod, hasKar, god, kar)
	if err := s.biluRepo.Exec(ctx, resetAll); err != nil {
		return err
	}

	resetSummary := querybuilder.NewUpdate("bilu").Set("pgod = 0, pgodh = 0")
	whereGodKar(resetSummary, "", hasGod, hasKar, god, kar)
	resetSummary.Where("konta = ''")
	if err := s.biluRepo.Exec(ctx, resetSummary); err != nil {
		return err
	}

	// STEP 2: Get max NIPO value
	maxNipoB := querybuilder.New(`select coalesce(max(nipo), 1) as nipo from bilu`)
	whereGodKar(maxNipoB, "", hasGod, hasKar, god, kar)
	maxNipoRecords, err := s.biluRepo.Select(ctx, maxNipoB)
	if err != nil {
		return err
	}
	maxk := int16(1)
	if maxNipoRecords != nil && len(*maxNipoRecords) > 0 && (*maxNipoRecords)[0].NiPo > 1 {
		maxk = (*maxNipoRecords)[0].NiPo
	}

	// STEP 3: Load ALL bilu rows into memory once, keyed by AOP.
	// This eliminates up to N×12 per-row DB lookups during summary aggregation.
	allBiluB := querybuilder.New(`select biluid, aop, konta, tgod, pgod, tgodh, pgodh, nipo, skraceni,
			pozic_1, pozic_2, pozic_3, pozic_4, pozic_5, pozic_6,
			pozic_7, pozic_8, pozic_9, pozic_10, pozic_11, pozic_12 from bilu`)
	whereGodKar(allBiluB, "", hasGod, hasKar, god, kar)
	allBiluRecords, err := s.biluRepo.Select(ctx, allBiluB)
	if err != nil {
		return err
	}

	// Map by AOP for O(1) lookup; pointer so we can mutate values in-place
	biluByAop := make(map[int]*domain.Bilu, len(*allBiluRecords))
	for i := range *allBiluRecords {
		row := &(*allBiluRecords)[i]
		biluByAop[row.AOP] = row
	}

	// fpro aggregate cache: key = "konta:god:kar:odMes:doMes"
	// Avoids re-querying the same konto prefix if multiple BILU rows share it.
	fproCache := make(map[string][2]float64)
	// The first failed aggregate; the rows are not written once it is set (see STEP 4).
	var aggErr error
	// The aggregates read fpro, so they are scoped by fpro's own god/kar columns (not by bilu's).
	fproGod, fproKar := s.fproRepo.GetHasGodHasKar()

	cachedFproAggregate := func(k, r, od, do int, sKonto string) (float64, float64) {
		key := fmt.Sprintf("%s:%d:%d:%d:%d", sKonto, k, r, od, do)
		if v, ok := fproCache[key]; ok {
			return v[0], v[1]
		}
		dug, pot, err := s.queryFproAggregate(ctx, fproGod, fproKar, k, r, od, do, sKonto)
		if err != nil {
			if aggErr == nil {
				aggErr = err
			}
			return 0, 0
		}
		fproCache[key] = [2]float64{dug, pot}
		return dug, pot
	}

	// STEP 4: Process each NIPO level (k = 1 to maxk)
	for k := int16(1); k <= maxk; k++ {
		// Collect rows at this level from the already-loaded map
		var levelRows []*domain.Bilu
		for i := range *allBiluRecords {
			if (*allBiluRecords)[i].NiPo == k {
				levelRows = append(levelRows, &(*allBiluRecords)[i])
			}
		}

		for _, bilu := range levelRows {
			if bilu.Konta != "" {
				// Leaf row: calculate from fpro using cached aggregate
				dVred, dVred1 := s.obradaKontaCached(bilu.Konta, god, kar, odMes, doMes, lPGODizPG, hasGod, hasKar, cachedFproAggregate)
				bilu.TGod = dVred
				bilu.TGodH = int64(math.Round(dVred / 1000))
				bilu.PGod = dVred1
				bilu.PGodH = int64(math.Round(dVred1 / 1000))
			} else {
				// Summary row: aggregate from referenced BILU rows via POZIC[1..12] using in-memory map
				pozici := [12]int16{
					bilu.Pozic1, bilu.Pozic2, bilu.Pozic3, bilu.Pozic4,
					bilu.Pozic5, bilu.Pozic6, bilu.Pozic7, bilu.Pozic8,
					bilu.Pozic9, bilu.Pozic10, bilu.Pozic11, bilu.Pozic12,
				}
				for _, pozic := range pozici {
					if pozic == 0 {
						continue
					}
					nAOP := int(pozic)
					if nAOP < 0 {
						nAOP = -nAOP
					}
					rel, found := biluByAop[nAOP]
					if !found {
						continue
					}
					if pozic > 0 {
						bilu.TGod += rel.TGod
						bilu.TGodH += rel.TGodH
						bilu.PGod += rel.PGod
						bilu.PGodH += rel.PGodH
					} else {
						bilu.TGod -= rel.TGod
						bilu.TGodH -= rel.TGodH
						bilu.PGod -= rel.PGod
						bilu.PGodH -= rel.PGodH
					}
				}
			}

			// Clamp negatives to zero
			if bilu.TGod < 0 {
				bilu.TGod = 0
			}
			if bilu.TGodH < 0 {
				bilu.TGodH = 0
			}
			if bilu.PGod < 0 {
				bilu.PGod = 0
			}
			if bilu.PGodH < 0 {
				bilu.PGodH = 0
			}

			// Single UPDATE per row (was 3× before)
			fields := []domain.Fields{
				{Name: "tgod", Value: fmt.Sprintf("%v", bilu.TGod)},
				{Name: "tgodh", Value: fmt.Sprintf("%v", bilu.TGodH)},
				{Name: "pgod", Value: fmt.Sprintf("%v", bilu.PGod)},
				{Name: "pgodh", Value: fmt.Sprintf("%v", bilu.PGodH)},
			}
			if aggErr != nil {
				return fmt.Errorf("obrada bilansa uspeha: %w", aggErr)
			}
			if err := s.biluRepo.Update(ctx, bilu, "biluid", bilu.BiluID, fields); err != nil {
				return err
			}
		}
	}

	// STEP 5: Fetch final results and populate table
	finalB := querybuilder.New(`select biluid, rbr, grac, nazp, aop, konta, tgod, pgod, tgodh, pgodh, nipo, skraceni from bilu`)
	whereGodKar(finalB, "", hasGod, hasKar, god, kar)
	if skraceni {
		finalB.AddEqual("skraceni", 1)
	}
	finalB.OrderBy("aop asc")

	finalRecords, err := s.biluRepo.Select(ctx, finalB)
	if err != nil {
		return err
	}

	var sumTGod, sumPGod float64
	var sumTGodH, sumPGodH int64
	if finalRecords != nil {
		for _, bilu := range *finalRecords {
			tgod := bilu.TGod
			pgod := bilu.PGod
			if tgod < 0 {
				tgod = 0
			}
			if pgod < 0 {
				pgod = 0
			}
			sumTGod += tgod
			sumPGod += pgod
			sumTGodH += bilu.TGodH
			sumPGodH += bilu.PGodH
			fields := []string{
				bilu.Grac,
				bilu.NazP,
				fmt.Sprintf("%d", bilu.AOP),
				bilu.Konta,
				common.FormatNumberWithSystemLocale(float64(bilu.TGodH), 0),
				common.FormatNumberWithSystemLocale(float64(bilu.PGodH), 0),
				common.FormatNumberWithSystemLocale(tgod, 2),
				common.FormatNumberWithSystemLocale(pgod, 2),
				fmt.Sprintf("%d", bilu.NiPo),
				fmt.Sprintf("%d", bilu.Skraceni),
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", bilu.BiluID), Fields: fields, HasUpdate: false, HasDelete: false})
		}
	}

	totals.TekGod = common.FormatNumberWithSystemLocale(sumTGod, 2)
	totals.PrethGod = common.FormatNumberWithSystemLocale(sumPGod, 2)
	totals.TekGodH = common.FormatNumberWithSystemLocale(float64(sumTGodH), 0)
	totals.PrethGodH = common.FormatNumberWithSystemLocale(float64(sumPGodH), 0)
	bilansiUkupno(tbl, bilansiSuma{4, float64(sumTGodH), 0}, bilansiSuma{5, float64(sumPGodH), 0},
		bilansiSuma{6, sumTGod, 2}, bilansiSuma{7, sumPGod, 2})

	return nil
}

// GetBilansUspehaZaStampu retrieves data for Bilans uspeha (income statement) for printing
func (s *BilansiResource) GetBilansUspehaZaStampu(ctx context.Context, tbl *domain.TableData, tipStampe string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	var b *querybuilder.Builder
	if tipStampe == common.TipStampePrint {
		b = querybuilder.New(`SELECT
			bilu.grac, bilu.nazp, bilu.aop,
			bilu.napomena, bilu.tgodh, bilu.pgodh FROM bilu`)
	}
	if tipStampe == common.TipStampePreview {
		b = querybuilder.New(`SELECT
			bilu.grac, bilu.nazp, bilu.aop,
			bilu.napomena, bilu.tgodh, bilu.pgodh, bilu.tgod, bilu.pgod, bilu.nipo, bilu.skraceni FROM bilu`)
	}
	if b == nil {
		// (The old code dereferenced a nil builder here and panicked.)
		return fmt.Errorf("invalid tip stampe: %s", tipStampe)
	}
	// Add filters directly in query
	if err := s.biluRepo.ApplyFilters(ctx, b, "bilu", nil, ""); err != nil {
		return err
	}

	// The rows of the print by AOP.
	b.OrderBy("bilu.aop ASC")

	// Execute query and get entities
	entities, err := s.biluRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	// The totals at the end of the table: the sums of the shown values.
	var sumTGodH, sumPGodH, sumTGod, sumPGod float64
	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			sumTGodH, sumPGodH = sumTGodH+float64(entity.TGodH), sumPGodH+float64(entity.PGodH)
			sumTGod, sumPGod = sumTGod+entity.TGod, sumPGod+entity.PGod
			fields := []string{}
			if tipStampe == common.TipStampePrint {
				fields = []string{
					entity.Grac,
					entity.NazP,
					fmt.Sprintf("%04d", entity.AOP),
					entity.Napomena,
					common.FormatNumberWithSystemLocale(entity.TGodH, 2),
					common.FormatNumberWithSystemLocale(entity.PGodH, 2),
				}
			}
			if tipStampe == common.TipStampePreview {
				fields = []string{
					entity.Grac,
					entity.NazP,
					fmt.Sprintf("%04d", entity.AOP),
					entity.Napomena,
					common.FormatNumberWithSystemLocale(entity.TGodH, 2),
					common.FormatNumberWithSystemLocale(entity.PGodH, 2),
					common.FormatNumberWithSystemLocale(entity.TGod, 2),
					common.FormatNumberWithSystemLocale(entity.PGod, 2),
					fmt.Sprintf("%d", entity.NiPo),
					fmt.Sprintf("%d", entity.Skraceni),
				}
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", entity.BiluID), Fields: fields, HasUpdate: true, HasDelete: true})
		}
	}
	if tipStampe == common.TipStampePreview {
		bilansiUkupno(tbl, bilansiSuma{4, sumTGodH, 2}, bilansiSuma{5, sumPGodH, 2}, bilansiSuma{6, sumTGod, 2}, bilansiSuma{7, sumPGod, 2})
	} else {
		bilansiUkupno(tbl, bilansiSuma{4, sumTGodH, 2}, bilansiSuma{5, sumPGodH, 2})
	}
	return nil
}

func (s *BilansiResource) ValidateBilansStanja(entity *domain.Bils) []domain.FieldError {
	var fieldErrors []domain.FieldError
	if entity.Rbr <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "rbr",
			ErrorMessage: "Red. Broj mora biti > 0",
		})
	}
	if entity.AOP <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "aop",
			ErrorMessage: "AOP mora biti > 0",
		})
	}
	if entity.NazP == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nazp",
			ErrorMessage: "Obavezan podatak...",
		})
	}
	if entity.NiPo < 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nipo",
			ErrorMessage: "Nivo podataka mora biti >= 0",
		})
	}

	return fieldErrors
}

func (s *BilansiResource) DeleteBilansStanja(ctx context.Context, id int64) error {
	return s.bilsRepo.Delete(ctx, common.IDbils, id)
}

// Bilu (Bilans Uspeha) Methods
func (s *BilansiResource) GetByIDBilu(ctx context.Context, idField string, idValue int64) (*domain.Bilu, error) {
	return s.biluRepo.GetByID(ctx, idField, idValue)
}

func (s *BilansiResource) UpdateBilu(ctx context.Context, entity *domain.Bilu, idField string, idValue interface{}, tableFields []domain.Fields) error {
	return s.biluRepo.Update(ctx, entity, idField, idValue, tableFields)
}

func (s *BilansiResource) AddBilu(ctx context.Context, entity *domain.Bilu, idField string, tableFields []domain.Fields) (int64, error) {
	return s.biluRepo.Create(ctx, entity, idField, tableFields)
}

func (s *BilansiResource) MapEntityToValuesBilu(entity *domain.Bilu, tableFields []domain.Fields) []domain.Fields {
	return s.biluService.MapEntityToValues(entity, tableFields)
}

func (s *BilansiResource) ValidateBilansUspeha(entity *domain.Bilu) []domain.FieldError {
	var fieldErrors []domain.FieldError
	if entity.Rbr <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "rbr",
			ErrorMessage: "Red. Broj mora biti > 0",
		})
	}
	if entity.AOP <= 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "aop",
			ErrorMessage: "AOP mora biti > 0",
		})
	}
	if entity.NazP == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nazp",
			ErrorMessage: "Obavezan podatak...",
		})
	}
	if entity.NiPo < 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nipo",
			ErrorMessage: "Nivo podataka mora biti >= 0",
		})
	}

	return fieldErrors
}

func (s *BilansiResource) DeleteBilansUspeha(ctx context.Context, id int64) error {
	return s.biluRepo.Delete(ctx, common.IDbilu, id)
}

func (s *BilansiResource) GetFieldCacheBilu() map[string]reflect.StructField {
	if s.biluService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.biluService.GetFieldCache()
}

// obradaKontaCached is the cached variant used by GetBilansUspehaObradaStampa.
// Instead of calling queryFproAggregate directly (which hits the DB each time),
// it accepts a closure that caches results by konto+god+kar+month-range key.
// This eliminates repeated DB round-trips when multiple BILU rows share a konto prefix.
func (s *BilansiResource) obradaKontaCached(konta string, god, kar, odMes, doMes int, lPGODizPG bool, hasGod, hasKar bool,
	cachedAggregate func(god, kar, odMes, doMes int, sKonto string) (float64, float64),
) (float64, float64) {
	if konta == "" {
		return 0, 0
	}

	var xMSaldo, xPSaldo float64
	var xTrosak, xPrihod float64
	var xTrosakPG, xPrihodPG float64

	for _, entry := range strings.Split(konta, ";") {
		entry = strings.TrimSpace(entry)
		if len(entry) < 2 {
			continue
		}
		cZnak := string(entry[0])
		if cZnak != "+" && cZnak != "-" {
			continue
		}
		sKonto := entry[1:]
		nPlusMinus := 1.0
		if cZnak == "-" {
			nPlusMinus = -1.0
		}

		dug, pot := cachedAggregate(god, kar, odMes, doMes, sKonto)
		diff := (dug - pot) * nPlusMinus
		if sKonto == "59" {
			xTrosak += diff
		}
		if sKonto == "69" {
			xPrihod += diff
		}
		xMSaldo += math.Abs(diff)

		if lPGODizPG {
			dugPG, potPG := cachedAggregate(god-1, kar, 0, 0, sKonto)
			diffPG := (dugPG - potPG) * nPlusMinus
			if sKonto == "59" {
				xTrosakPG += diffPG
			}
			if sKonto == "69" {
				xPrihodPG += diffPG
			}
			xPSaldo += math.Abs(diffPG)
		}
	}

	xTrosak = math.Abs(xTrosak)
	xPrihod = math.Abs(xPrihod)
	xTrosakPG = math.Abs(xTrosakPG)
	xPrihodPG = math.Abs(xPrihodPG)

	if strings.HasPrefix(konta, "+59") {
		if xTrosak <= xPrihod {
			xMSaldo = 0
		} else {
			xMSaldo = xTrosak - xPrihod
		}
		if xTrosakPG <= xPrihodPG {
			xPSaldo = 0
		} else {
			xPSaldo = xTrosakPG - xPrihodPG
		}
	}
	if strings.HasPrefix(konta, "+69") {
		if xTrosak > xPrihod {
			xMSaldo = 0
		} else {
			xMSaldo = xPrihod - xTrosak
		}
		if xTrosakPG > xPrihodPG {
			xPSaldo = 0
		} else {
			xPSaldo = xPrihodPG - xTrosakPG
		}
	}

	return xMSaldo, xPSaldo
}

// obradaKonta computes current-year (dVred) and previous-year (dVred1) saldo values
// for a konta list (ipLISTA) in the form "+59;-69;+1230" — semicolon-separated
// signed account prefixes. Translated from WinDev ObradaKONTA procedure.
//
// odMes/doMes: month range for current-year aggregation (1–12).
// lPGODizPG:   when true, also calculates the previous-year saldo (god-1, full year).
func (s *BilansiResource) obradaKonta(ctx context.Context, konta string, god, kar, odMes, doMes int, lPGODizPG bool) (float64, float64, error) {
	if konta == "" {
		return 0, 0, nil
	}
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	var xMSaldo, xPSaldo float64
	var xTrosak, xPrihod float64
	var xTrosakPG, xPrihodPG float64

	for _, entry := range strings.Split(konta, ";") {
		entry = strings.TrimSpace(entry)
		if len(entry) < 2 {
			continue
		}
		cZnak := string(entry[0])
		if cZnak != "+" && cZnak != "-" {
			continue
		}
		sKonto := entry[1:]
		nPlusMinus := 1.0
		if cZnak == "-" {
			nPlusMinus = -1.0
		}

		// Current year: sum fpro movements for the given month range
		dug, pot, err := s.queryFproAggregate(ctx, hasGod, hasKar, god, kar, odMes, doMes, sKonto)
		if err != nil {
			return 0, 0, err
		}
		diff := (dug - pot) * nPlusMinus
		if sKonto == "59" {
			xTrosak += diff
		}
		if sKonto == "69" {
			xPrihod += diff
		}
		xMSaldo += math.Abs(diff)

		// Previous year: same query on god-1, full year (no month filter)
		if lPGODizPG {
			dugPG, potPG, err := s.queryFproAggregate(ctx, hasGod, hasKar, god-1, kar, 0, 0, sKonto)
			if err != nil {
				return 0, 0, err
			}
			diffPG := (dugPG - potPG) * nPlusMinus
			if sKonto == "59" {
				xTrosakPG += diffPG
			}
			if sKonto == "69" {
				xPrihodPG += diffPG
			}
			xPSaldo += math.Abs(diffPG)
		}
	}

	// Finalise absolute values for expense/revenue tracking
	xTrosak = math.Abs(xTrosak)
	xPrihod = math.Abs(xPrihod)
	xTrosakPG = math.Abs(xTrosakPG)
	xPrihodPG = math.Abs(xPrihodPG)

	// Special handling for expense accounts (+59): profit = expenses - revenues (if positive)
	if strings.HasPrefix(konta, "+59") {
		if xTrosak <= xPrihod {
			xMSaldo = 0
		} else {
			xMSaldo = xTrosak - xPrihod
		}
		if xTrosakPG <= xPrihodPG {
			xPSaldo = 0
		} else {
			xPSaldo = xTrosakPG - xPrihodPG
		}
	}
	// Special handling for revenue accounts (+69): profit = revenues - expenses (if positive)
	if strings.HasPrefix(konta, "+69") {
		if xTrosak > xPrihod {
			xMSaldo = 0
		} else {
			xMSaldo = xPrihod - xTrosak
		}
		if xTrosakPG > xPrihodPG {
			xPSaldo = 0
		} else {
			xPSaldo = xPrihodPG - xTrosakPG
		}
	}

	return xMSaldo, xPSaldo, nil
}

// queryFproAggregate returns (dug, pot) totals from fpro for the given account prefix/code.
// queryFproAggregate returns (dug, pot) totals from fpro for the given account prefix.
// (e.g. a booking on "2042" was rolled up into FSAL records for "204" and "20").
// Here we replicate that by using prefix matching directly on fpro, so a query for
// sKonto="204" matches fpro rows with konto "2042", "20420", etc.
// odMes/doMes = 0 means full-year (no month filter).
func (s *BilansiResource) queryFproAggregate(ctx context.Context, hasGod, hasKar bool, god, kar, odMes, doMes int, sKonto string) (float64, float64, error) {
	// The tipdok is the column of fpro itself. The old query read fnal.tipdok without fnal in
	// FROM, failed, and the error was ignored, so the result was always 0, 0.
	b := querybuilder.New(`select
			coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) as dug,
			coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as pot
			from fpro`)
	whereGodKar(b, "fpro.", hasGod, hasKar, god, kar)

	// Only movements, not opening balance postings
	b.Where("fpro.tipdok != '00'")

	// Prefix match covers all sub-accounts at every depth:
	// sKonto="20"  matches fpro konto "20","204","2042","20420", ...
	// sKonto="204" matches fpro konto "204","2042","20420", ...
	// sKonto="2042" matches fpro konto "2042","20420", ...
	// No vkonta filter — fpro transactions exist at the actual booking level only,
	b.AddLikeBegin("fpro.konto", sKonto)

	// Exclude closing/transfer accounts at all levels
	b.Where("fpro.konto != '599'").
		Where("fpro.konto != '699'").
		Where("left(fpro.konto, 4) != '5999'").
		Where("left(fpro.konto, 4) != '6999'")

	// Month filter: omitted when odMes == 0 (full-year query)
	whereMesecOdDo(b, odMes, doMes)

	rows, err := s.fproRepo.Select(ctx, b)
	if err != nil {
		return 0, 0, fmt.Errorf("promet konta %s: %w", sKonto, err)
	}
	if len(*rows) == 0 {
		return 0, 0, nil
	}
	row := (*rows)[0]
	return row.Dug.Float64, row.Pot.Float64, nil
}

func (s *BilansiResource) getKontoNaziv(ctx context.Context, konto string) string {
	// The konto is the key of the lookup; the kontni plan of the period of the session.
	b := querybuilder.New("select naziv from fkpl").
		Where("konto = ?", konto).
		Where("vkonta > ?", 1)
	if err := s.fkplRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return ""
	}
	rows, err := s.fkplRepo.Select(ctx, b)
	if err != nil || rows == nil || len(*rows) == 0 {
		return ""
	}
	return (*rows)[0].Naziv
}

// GetFieldCache returns the cached field structure
func (s *BilansiResource) GetFieldCache() map[string]reflect.StructField {
	if s.bilsService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.bilsService.GetFieldCache()
}

// resetBilsTotalsForYear resets TGOD, PGOD values for all BILS records in the period
func (s *BilansiResource) resetBilsTotalsForYear(god, kar int, bilsMap map[int64]*domain.Bils) error {
	// This would query the database and reset values
	// For now, we'll handle this in the processBilsLevel function
	return nil
}

// setServiceFieldValues initializes table field definitions for Bilansi
func (s *BilansiResource) setServiceFieldValues() {
	// Fields for Zakljucni list
	s.zakljucniTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "8", Field: "", SkipInSearch: true, IncludeInTotals: true},
		{Name: "konto", Label: "Konto", Width: "12", Field: "bils.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "10", Field: "bils.sifra", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv", Width: "25", Field: "bils.naziv", SkipInSearch: false},
		{Name: "pstduguje", Label: "Početno stanje duguje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pstpotrazuje", Label: "Početno stanje potražuje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "prometduguje", Label: "Promet duguje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "prometpotrazuje", Label: "Promet potražuje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "saldoduguje", Label: "Saldo duguje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "saldopotrazuje", Label: "Saldo potražuje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
	}

	// Fields for Bilans stanja
	s.bilansStanjaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "8", Field: "", SkipInSearch: true, IncludeInTotals: true},
		{Name: "grupa_racuna", Label: "Grupa računa", Width: "15", Field: "bils.vkonta", SkipInSearch: false},
		{Name: "naziv_pozicije", Label: "Naziv pozicije", Width: "25", Field: "bils.naziv", SkipInSearch: false},
		{Name: "oznaka_aop", Label: "Oznaka za AOP", Width: "12", Field: "", SkipInSearch: true},
		{Name: "spisak_konta", Label: "Spisak konta", Width: "15", Field: "", SkipInSearch: true},
		{Name: "tekuca_godina", Label: "Tekuća godina", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "prethodna_godina", Label: "Prethodna godina", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pocetno_stanje", Label: "Prethodna godina PS", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "nivo_podataka", Label: "Nivo podataka", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tekuca_hiljada", Label: "Tekuća u hiljadama", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "prethodna_hiljada", Label: "Prethodna u hiljadama", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pocetno_hiljada", Label: "Prethodna u hiljadama PS", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pozic_1", Label: "Pozicija 1", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_2", Label: "Pozicija 2", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_3", Label: "Pozicija 3", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_4", Label: "Pozicija 4", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_5", Label: "Pozicija 5", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_6", Label: "Pozicija 6", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_7", Label: "Pozicija 7", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_8", Label: "Pozicija 8", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_9", Label: "Pozicija 9", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_10", Label: "Pozicija 10", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_11", Label: "Pozicija 11", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_12", Label: "Pozicija 12", Width: "12", Field: "", SkipInSearch: true},
	}
	// Fields for Bilans stanja
	s.bilansStanjaStampaTableFields = []domain.Fields{
		{Name: "grupa_racuna", Label: "Grupa računa", Width: "15", Field: "bils.grac", SkipInSearch: false, IncludeInTotals: true},
		{Name: "nazivp", Label: "Naziv pozicije", Width: "25", Field: "bils.nazivp", SkipInSearch: false, TextAlign: "left"},
		{Name: "oznaka_aop", Label: "Oznaka za AOP", Width: "12", Field: "bils.aop", SkipInSearch: true},
		{Name: "napomena", Label: "Napomena broj", Width: "25", Field: "bils.napomena", SkipInSearch: false, TextAlign: "left"},
		{Name: "tekuca_godina", Label: "Tekuća godina", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "prethodna_godina", Label: "Prethodna godina", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pocetno_stanje", Label: "Prethodna godina PS", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "tekuca_hiljada", Label: "Tekuća u hiljadama", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "prethodna_hiljada", Label: "Prethodna u hiljadama", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pocetno_hiljada", Label: "Prethodna u hiljadama PS", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "nivo_podataka", Label: "Nivo podataka", Width: "12", Field: "", SkipInSearch: true},
		{Name: "skraceni", Label: "Skraceni bilans", Width: "15", Field: "bilu.skraceni", SkipInSearch: true},
	}
	// Fields for Bilans uspeha
	s.bilansUspehaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "8", Field: "", SkipInSearch: true, IncludeInTotals: true},
		{Name: "grac", Label: "Grupa računa", Width: "15", Field: "bils.vkonta", SkipInSearch: false},
		{Name: "nazp", Label: "Naziv pozicije", Width: "25", Field: "bils.naziv", SkipInSearch: false},
		{Name: "aop", Label: "Oznaka za AOP", Width: "12", Field: "", SkipInSearch: true},
		{Name: "konta", Label: "Spisak konta", Width: "15", Field: "", SkipInSearch: true},
		{Name: "tgod", Label: "Tekuća godina", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pgod", Label: "Prethodna godina", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "nipo", Label: "Nivo podataka", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tgoh", Label: "Tekuća u hiljadama", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pgoh", Label: "Prethodna u hiljadama", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pozic_1", Label: "Pozicija 1", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_2", Label: "Pozicija 2", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_3", Label: "Pozicija 3", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_4", Label: "Pozicija 4", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_5", Label: "Pozicija 5", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_6", Label: "Pozicija 6", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_7", Label: "Pozicija 7", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_8", Label: "Pozicija 8", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_9", Label: "Pozicija 9", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_10", Label: "Pozicija 10", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_11", Label: "Pozicija 11", Width: "12", Field: "", SkipInSearch: true},
		{Name: "pozic_12", Label: "Pozicija 12", Width: "12", Field: "", SkipInSearch: true},
	}

	// Fields for Bilans uspeha
	s.bilansUspehaStampaTableFields = []domain.Fields{
		{Name: "grac", Label: "Grupa računa, račun", Width: "15", Field: "bilu.grac", SkipInSearch: false, IncludeInTotals: true},
		{Name: "nazp", Label: "Naziv pozicije", Width: "25", Field: "bilu.nazp", SkipInSearch: false, TextAlign: "left"},
		{Name: "aop", Label: "AOP ", Width: "12", Field: "bilu.aop", SkipInSearch: true},
		{Name: "napomena", Label: "Napomena, broj", Width: "15", Field: "bilu.napomena", SkipInSearch: true},
		{Name: "tgodh", Label: "Iznos u hiljadama \n Tekuća godina", Width: "15", Field: "bilu.tgodh", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pgodh", Label: "Iznos u hiljadama \n Prethodna godina", Width: "15", Field: "bilu.pgodh", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "tgod", Label: "Tekuca godina", Width: "15", Field: "bilu.tgod", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pgod", Label: "Prethodna godina", Width: "15", Field: "bilu.pgod", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "nipo", Label: "Nivo podataka", Width: "12", Field: "bilu.nipo", SkipInSearch: true},
		{Name: "skraceni", Label: "Skraceni bilans", Width: "15", Field: "bilu.skraceni", SkipInSearch: true},
	}
}
