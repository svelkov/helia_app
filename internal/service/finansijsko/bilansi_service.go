package finansijsko

import (
	"context"
	"fmt"
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
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
		cfg:         cfg,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *BilansiResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
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
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	sqlQuery, queryArgs := s.buildZakljucniListAnalitikaQuery(session, hasGod, hasKar, params, common.TipStampePreview)
	if !getTotalRecords {
		sqlQuery = fmt.Sprintf("%s LIMIT %d OFFSET %d", sqlQuery, pageSize, (currentPage-1)*pageSize)
	}

	tbl.Headers = s.GetZakljucniTableFields()
	common.SetupTablePagination(tbl, currentPage, pageSize)

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", queryArgs, "", "")
	if err != nil {
		return err
	}

	// Totals pass: every hierarchy level is an aggregate of the analitika rows, so only the
	// leaves are counted and summed - adding the subtotal levels up would count twice.
	if getTotalRecords {
		leaves := make([]domain.FproDto, 0, len(*entities))
		for _, entity := range *entities {
			if entity.NivoOrder == zakljucniNivoAnalitika {
				leaves = append(leaves, entity)
			}
		}
		setZakljucniTotals(tbl, leaves, pageSize)
		return nil
	}

	// The query returns the hierarchy already ordered level by level, so every subtotal row
	// follows the rows it aggregates. Fields[0] carries the level marker: the running row
	// number for analitika rows, G1/G2/G3 for the sintetika/grupa/klasa subtotals.
	//
	// The hierarchy is always returned complete - slicing it with LIMIT/OFFSET would separate
	// subtotal rows from their children - so pageSize/currentPage only drive the pagination
	// counters (the total is the number of analitika rows).
	rowNum := 1
	// Every subtotal row shows the sum of the netted sides below it, so the saldo columns foot
	// at every level and the grid shows the same numbers as the printed report.
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(s.cfg, params.TipLista, *entities)
	nazivCache := make(map[string]string)
	for i, entity := range *entities {
		marker := zakljucniNivoMarker(entity.NivoOrder, rowNum)
		if entity.NivoOrder == zakljucniNivoAnalitika {
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
			fmt.Sprintf("%d", i+1),
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
func zakljucniSaldoAccumulatorFromDetails(cfg config.Config, tipLista string, entities []domain.FproDto) *zakljucniSaldoAccumulator {
	acc := newZakljucniSaldoAccumulator(cfg.NDuzSint, zakljucniDetailNivo(tipLista))
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
func (s *BilansiResource) buildZakljucniListAnalitikaQuery(session *domain.UserSession, hasGod, hasKar bool, params domain.ZakljucniParams, printType string) (string, []any) {
	filterQb := common.NewQueryBuilder(`SELECT konto, sifra, fnal.tipdok, kat, iznos FROM fpro`, true)
	filterQb.AddJoin(" inner join fnal on fnal.idfnal = fpro.idfnal ")
	if hasGod {
		filterQb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		filterQb.AddEqual("fpro.kar", session.SelectedKar)
	}
	filterQb.AddEqual("fpro.vkonta", 1)
	filterQb.AddCondition("fnal.danal", params.OdDatuma, ">=")
	filterQb.AddCondition("fnal.danal", params.DoDatuma, "<=")
	// Klasa 9 accounts are excluded for printing only, exactly like the flattened
	// zakljucniInnerQuery does it.
	if printType == common.TipStampePrint && params.Klasa9 == "false" {
		filterQb.AddCustomCondition("fpro.konto NOT LIKE '9%'")
	}
	filterSql, filterArgs := filterQb.Build()

	// The sintetika prefix length is a parameter as well, so its placeholder has to continue
	// the numbering of the placeholders already used by the "filtered" CTE above.
	sintParam := len(filterArgs) + 1
	nivoOrderCase := fmt.Sprintf("CASE WHEN k4 IS NOT NULL THEN %d WHEN k3 IS NOT NULL THEN %d WHEN k2 IS NOT NULL THEN %d ELSE %d END AS nivo_order",
		zakljucniNivoAnalitika, zakljucniNivoSintetika, zakljucniNivoGrupa, zakljucniNivoKlasa)

	baseSql := fmt.Sprintf(`WITH filtered AS (%s),
	agg AS (
		SELECT
			LEFT(konto, 1) AS k1,
			LEFT(konto, 2) AS k2,
			LEFT(konto, $%d) AS k3,
			konto AS k4,
			sifra AS s4,
			SUM(CASE WHEN tipdok = '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS pocstanjedug,
			SUM(CASE WHEN tipdok = '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS pocstanjepot,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS prometdug,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS prometpot
		FROM filtered
		GROUP BY GROUPING SETS (
			(konto, sifra),
			(LEFT(konto, $%d)),
			(LEFT(konto, 2)),
			(LEFT(konto, 1))
		)
	)
	SELECT
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
	) sub`, filterSql, sintParam, sintParam, nivoOrderCase)

	qb := common.NewQueryBuilder(baseSql, false)
	qb.AddArgs(filterArgs...)  // $1..$n : the "filtered" CTE
	qb.AddArgs(s.cfg.NDuzSint) // $n+1   : the sintetika prefix length

	// The search placeholder follows the CTE arguments and the prefix length.
	if params.SearchText != "" {
		qb.Where(zakljucniSearchCondition(len(filterArgs)+2, "sub.konto", "sub.sifra"), params.SearchText)
	}
	qb.AddOrderBy(`sort_key COLLATE "C", sifra_grp COLLATE "C" NULLS LAST`)

	return qb.Build()
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
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	sqlQuery, queryArgs := s.buildZakljucniListSintetikaQuery(session, hasGod, hasKar, params, common.TipStampePreview)
	if !getTotalRecords {
		sqlQuery = fmt.Sprintf("%s LIMIT %d OFFSET %d", sqlQuery, pageSize, (currentPage-1)*pageSize)
	}

	tbl.Headers = s.GetZakljucniTableFields()
	common.SetupTablePagination(tbl, currentPage, pageSize)

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", queryArgs, "", "")
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
		setZakljucniTotals(tbl, sintetike, pageSize)
		return nil
	}

	// The query returns the hierarchy already ordered level by level, so every subtotal row
	// follows the rows it aggregates. Fields[0] carries the level marker (G1/G2/G3): there are
	// no detail rows, so the marker is never a row number.
	//
	// The hierarchy is always returned complete - slicing it with LIMIT/OFFSET would separate
	// subtotal rows from their children - so pageSize/currentPage only drive the pagination
	// counters (the total is the number of sintetika rows).
	// Every subtotal row shows the sum of the netted sides below it, so the saldo columns foot
	// at every level and the grid shows the same numbers as the printed report.
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(s.cfg, params.TipLista, *entities)
	nazivCache := make(map[string]string)
	for i, entity := range *entities {
		marker := zakljucniNivoMarker(entity.NivoOrder, 0) // no detail rows, so no row numbering
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
			fmt.Sprintf("%d", i+1),
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
func (s *BilansiResource) buildZakljucniListSintetikaQuery(session *domain.UserSession, hasGod, hasKar bool, params domain.ZakljucniParams, printType string) (string, []any) {
	filterQb := common.NewQueryBuilder(`SELECT konto, fnal.tipdok, kat, iznos FROM fpro`, true)
	filterQb.AddJoin(" inner join fnal on fnal.idfnal = fpro.idfnal ")
	if hasGod {
		filterQb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		filterQb.AddEqual("fpro.kar", session.SelectedKar)
	}
	filterQb.AddCondition(fmt.Sprintf("left(fpro.konto, %d)", s.cfg.NDuzSint), params.OdKonta, ">=")
	filterQb.AddCondition(fmt.Sprintf("left(fpro.konto, %d)", s.cfg.NDuzSint), params.DoKonta, "<=")
	filterQb.AddCondition("fnal.danal", params.OdDatuma, ">=")
	filterQb.AddCondition("fnal.danal", params.DoDatuma, "<=")
	// Klasa 9 accounts are excluded for printing only, exactly like the flattened
	// zakljucniInnerQuery does it.
	if printType == common.TipStampePrint && params.Klasa9 == "false" {
		filterQb.AddCustomCondition("fpro.konto NOT LIKE '9%'")
	}
	filterSql, filterArgs := filterQb.Build()

	// The sintetika prefix length is a parameter as well, so its placeholder has to continue
	// the numbering of the placeholders already used by the "filtered" CTE above.
	sintParam := len(filterArgs) + 1
	nivoOrderCase := fmt.Sprintf("CASE WHEN k3 IS NOT NULL THEN %d WHEN k2 IS NOT NULL THEN %d ELSE %d END AS nivo_order",
		zakljucniNivoSintetika, zakljucniNivoGrupa, zakljucniNivoKlasa)

	baseSql := fmt.Sprintf(`WITH filtered AS (%s),
	agg AS (
		SELECT
			LEFT(konto, 1) AS k1,
			LEFT(konto, 2) AS k2,
			LEFT(konto, $%d) AS k3,
			SUM(CASE WHEN tipdok = '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS pocstanjedug,
			SUM(CASE WHEN tipdok = '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS pocstanjepot,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS prometdug,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS prometpot
		FROM filtered
		GROUP BY GROUPING SETS (
			(LEFT(konto, $%d)),
			(LEFT(konto, 2)),
			(LEFT(konto, 1))
		)
	)
	SELECT
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
	) sub`, filterSql, sintParam, sintParam, nivoOrderCase)

	qb := common.NewQueryBuilder(baseSql, false)
	qb.AddArgs(filterArgs...)  // $1..$n : the "filtered" CTE
	qb.AddArgs(s.cfg.NDuzSint) // $n+1   : the sintetika prefix length

	// The search placeholder follows the CTE arguments and the prefix length.
	if params.SearchText != "" {
		qb.Where(zakljucniSearchCondition(len(filterArgs)+2, "sub.konto"), params.SearchText)
	}
	qb.AddOrderBy(`sort_key COLLATE "C" NULLS LAST`)

	return qb.Build()
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
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	sqlQuery, queryArgs := s.buildZakljucniListSubsintetikaQuery(session, hasGod, hasKar, params, common.TipStampePreview)
	if !getTotalRecords {
		sqlQuery = fmt.Sprintf("%s LIMIT %d OFFSET %d", sqlQuery, pageSize, (currentPage-1)*pageSize)
	}
	tbl.Headers = s.GetZakljucniTableFields()
	common.SetupTablePagination(tbl, currentPage, pageSize)

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", queryArgs, "", "")
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
		setZakljucniTotals(tbl, konta, pageSize)
		return nil
	}

	// The query returns the hierarchy already ordered level by level, so every subtotal row
	// follows the rows it aggregates. Fields[0] carries the level marker (G1..G4): there are no
	// detail rows, so the marker is never a row number.
	//
	// The hierarchy is always returned complete - slicing it with LIMIT/OFFSET would separate
	// subtotal rows from their children - so pageSize/currentPage only drive the pagination
	// counters (the total is the number of subsintetika rows).
	// Every subtotal row shows the sum of the netted sides below it, so the saldo columns foot
	// at every level and the grid shows the same numbers as the printed report.
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(s.cfg, params.TipLista, *entities)
	nazivCache := make(map[string]string)
	for i, entity := range *entities {
		marker := zakljucniNivoGroupMarker(entity.NivoOrder, zakljucniNivoAnalitika)
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
			fmt.Sprintf("%d", i+1),
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
		tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", i), Fields: fields, HasUpdate: false, HasDelete: false})
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
func (s *BilansiResource) buildZakljucniListSubsintetikaQuery(session *domain.UserSession, hasGod, hasKar bool, params domain.ZakljucniParams, printType string) (string, []any) {
	filterQb := common.NewQueryBuilder(`SELECT fpro.konto, fnal.tipdok, fpro.kat, fpro.iznos FROM fpro`, true)
	filterQb.AddJoin(" inner join fnal on fnal.idfnal = fpro.idfnal ")
	if hasGod {
		filterQb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		filterQb.AddEqual("fpro.kar", session.SelectedKar)
	}
	filterQb.AddCondition("fpro.konto", params.OdKonta, ">=")
	filterQb.AddCondition("fpro.konto", params.DoKonta, "<=")
	filterQb.AddCondition("fnal.danal", params.OdDatuma, ">=")
	filterQb.AddCondition("fnal.danal", params.DoDatuma, "<=")
	// Klasa 9 accounts are excluded for printing only, exactly like the flattened
	// zakljucniInnerQuery does it.
	if printType == common.TipStampePrint && params.Klasa9 == "false" {
		filterQb.AddCustomCondition("fpro.konto NOT LIKE '9%'")
	}
	filterSql, filterArgs := filterQb.Build()

	// The sintetika prefix length is a parameter as well, so its placeholder has to continue
	// the numbering of the placeholders already used by the "filtered" CTE above.
	sintParam := len(filterArgs) + 1
	nivoOrderCase := fmt.Sprintf("CASE WHEN k4 IS NOT NULL THEN %d WHEN k3 IS NOT NULL THEN %d WHEN k2 IS NOT NULL THEN %d ELSE %d END AS nivo_order",
		zakljucniNivoAnalitika, zakljucniNivoSintetika, zakljucniNivoGrupa, zakljucniNivoKlasa)

	baseSql := fmt.Sprintf(`WITH filtered AS (%s),
	agg AS (
		SELECT
			LEFT(konto, 1) AS k1,
			LEFT(konto, 2) AS k2,
			LEFT(konto, $%d) AS k3,
			konto AS k4,
			SUM(CASE WHEN tipdok = '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS pocstanjedug,
			SUM(CASE WHEN tipdok = '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS pocstanjepot,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (1,2) THEN iznos ELSE 0 END) AS prometdug,
			SUM(CASE WHEN tipdok <> '00' AND kat IN (3,4) THEN iznos ELSE 0 END) AS prometpot
		FROM filtered
		GROUP BY GROUPING SETS (
			(konto),
			(LEFT(konto, $%d)),
			(LEFT(konto, 2)),
			(LEFT(konto, 1))
		)
	)
	SELECT
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
	) sub`, filterSql, sintParam, sintParam, nivoOrderCase)

	qb := common.NewQueryBuilder(baseSql, false)
	qb.AddArgs(filterArgs...)  // $1..$n : the "filtered" CTE
	qb.AddArgs(s.cfg.NDuzSint) // $n+1   : the sintetika prefix length

	// The search placeholder follows the CTE arguments and the prefix length.
	if params.SearchText != "" {
		qb.Where(zakljucniSearchCondition(len(filterArgs)+2, "sub.konto"), params.SearchText)
	}
	qb.AddOrderBy(`sort_key COLLATE "C" NULLS LAST`)

	return qb.Build()
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
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	// nDuzSint is not used any more: the hierarchy queries take the prefix length from the
	// configuration (s.cfg.NDuzSint), which is the value the handler passes in.

	// detailNivo is the level printed as numbered detail rows; every level above it is printed
	// as a G<n> subtotal, exactly like the preview shows them.
	detailNivo := zakljucniDetailNivo(params.TipLista)
	var sqlQuery string
	var queryArgs []any
	switch params.TipLista {
	case "1":
		sqlQuery, queryArgs = s.buildZakljucniListAnalitikaQuery(session, hasGod, hasKar, params, common.TipStampePrint)
	case "2":
		sqlQuery, queryArgs = s.buildZakljucniListSubsintetikaQuery(session, hasGod, hasKar, params, common.TipStampePrint)
	case "3":
		sqlQuery, queryArgs = s.buildZakljucniListSintetikaQuery(session, hasGod, hasKar, params, common.TipStampePrint)
	default:
		return fmt.Errorf("invalid tip_lista: %s", params.TipLista)
	}

	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", queryArgs, "", "")
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
	saldoAcc := zakljucniSaldoAccumulatorFromDetails(s.cfg, params.TipLista, *entities)

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

// zakljucniSearchCondition builds the OR-ed ILIKE filter of the Zakljucni list search box.
// Every field reuses the same placeholder, so the search term is passed only once.
func zakljucniSearchCondition(placeholder int, fields ...string) string {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		parts = append(parts, fmt.Sprintf("%s ILIKE '%%' || $%d || '%%'", f, placeholder))
	}
	return " ( " + strings.Join(parts, " OR ") + " ) "
}

// ---- Zakljucni list queries -------------------------------------------------
//
// One query per tipLista, because the aggregation grain and the way the account name is
// resolved differ:
//
//	tipLista 1 (analitika)    one row per konto+sifra, name via fkpl.idfkpl of the row
//	tipLista 2 (subsintetika) one row per konto,        name = newest fkpl row of that konto
//	tipLista 3 (sintetika)    one row per konto prefix, name = newest fkpl row of that prefix
//
// Everything the three share (period/konta/sifra filters, search, ordering, pagination and
// the totals pass) lives in the helpers below.

// zakljucniQueryOptions carries the call parameters plus the per-variant SQL bits.
type zakljucniQueryOptions struct {
	params          domain.ZakljucniParams
	printType       string
	getTotalRecords bool
	pageSize        int
	currentPage     int

	// filled in by the tipLista query that is running:
	orderByCols string // ORDER BY of the outer query (agg.* aliases)
	nameExpr    string // expression with the account name (SELECT + search filter)
	withSifra   bool   // tipLista 1 aggregates and searches by sifra as well
}

// getZakljucniQuery builds and executes the query for the requested tipLista of the
// Zakljucni list (1 = analitika, 2 = subsintetika, 3 = sintetika).
func (s *BilansiResource) getZakljucniQuery(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, params domain.ZakljucniParams, printType string, pageSize, currentPage int) (*[]domain.FproDto, error) {
	o := zakljucniQueryOptions{
		params:          params,
		printType:       printType,
		getTotalRecords: getTotalRecords,
		pageSize:        pageSize,
		currentPage:     currentPage,
	}
	switch params.TipLista {
	case "1":
		return s.getZakljucniAnalitikaQuery(ctx, tbl, o)
	case "2":
		return s.getZakljucniSubsintetikaQuery(ctx, tbl, o)
	case "3":
		return s.getZakljucniSintetikaQuery(ctx, tbl, o)
	default:
		return nil, fmt.Errorf("invalid tip_lista: %s", params.TipLista)
	}
}

// getZakljucniAnalitikaQuery (tipLista "1"): one row per konto + sifra. The name comes from
// the fkpl row the transaction itself points to (fkpl.idfkpl), so it joins directly.
func (s *BilansiResource) getZakljucniAnalitikaQuery(ctx context.Context, tbl *domain.TableData, o zakljucniQueryOptions) (*[]domain.FproDto, error) {
	innerSql, innerArgs, err := s.zakljucniInnerQuery(ctx, o.params, o.printType,
		"fpro.konto, COALESCE(fpro.sifra, '') as sifra, fpro.idfkpl,",
		"fpro.konto, fpro.sifra, fpro.idfkpl", true, true)
	if err != nil {
		return nil, err
	}

	o.orderByCols = "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC, COALESCE(NULLIF(agg.sifra, '')::numeric, 0) ASC"
	o.nameExpr = "fkpl.naziv"
	o.withSifra = true

	needName := zakljucniNeedsName(o.getTotalRecords, o.params.SearchText)
	qb := zakljucniOuterQueryBuilder(innerSql, "left join fkpl on fkpl.idfkpl = agg.idfkpl", o.nameExpr, innerArgs, needName)
	return s.executeZakljucniQuery(ctx, tbl, qb, innerArgs, o)
}

// getZakljucniSubsintetikaQuery (tipLista "2"): one row per konto. The name is the naziv of
// the newest fkpl row of that konto (DISTINCT ON lookup, joined on konto).
func (s *BilansiResource) getZakljucniSubsintetikaQuery(ctx context.Context, tbl *domain.TableData, o zakljucniQueryOptions) (*[]domain.FproDto, error) {
	innerSql, innerArgs, err := s.zakljucniInnerQuery(ctx, o.params, o.printType,
		"fpro.konto,", "fpro.konto", false, true)
	if err != nil {
		return nil, err
	}

	o.orderByCols = "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC"
	o.nameExpr = "fkpl_data.naziv"

	needName := zakljucniNeedsName(o.getTotalRecords, o.params.SearchText)
	nameJoin, args := "", innerArgs
	if needName {
		nameJoin, args = zakljucniNameLookupJoin("konto", 2, innerArgs)
	}
	qb := zakljucniOuterQueryBuilder(innerSql, nameJoin, o.nameExpr, args, needName)
	return s.executeZakljucniQuery(ctx, tbl, qb, args, o)
}

// getZakljucniSintetikaQuery (tipLista "3"): one row per konto prefix
// (LEFT(konto, NDuzSint)). The name is the naziv of the newest fkpl row of that prefix.
func (s *BilansiResource) getZakljucniSintetikaQuery(ctx context.Context, tbl *domain.TableData, o zakljucniQueryOptions) (*[]domain.FproDto, error) {
	innerSql, innerArgs, err := s.zakljucniInnerQuery(ctx, o.params, o.printType,
		fmt.Sprintf("LEFT(fpro.konto, %d) as konto,", s.cfg.NDuzSint),
		fmt.Sprintf("LEFT(fpro.konto, %d)", s.cfg.NDuzSint), false, false)
	if err != nil {
		return nil, err
	}

	o.orderByCols = "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC"
	o.nameExpr = "fkpl_data.naziv"

	needName := zakljucniNeedsName(o.getTotalRecords, o.params.SearchText)
	nameJoin, args := "", innerArgs
	if needName {
		nameJoin, args = zakljucniNameLookupJoin(fmt.Sprintf("LEFT(konto, %d)", s.cfg.NDuzSint), 3, innerArgs)
	}
	qb := zakljucniOuterQueryBuilder(innerSql, nameJoin, o.nameExpr, args, needName)
	return s.executeZakljucniQuery(ctx, tbl, qb, args, o)
}

// zakljucniInnerQuery builds the INNER aggregation of the Zakljucni list - one row per
// tipLista grain - with all the report filters: period, konta (and sifra) range, klasa 9,
// vkonta and samosaprometom.
//
//	selectCols   the grouping columns (must end with a comma)
//	groupByCols  the matching GROUP BY expression
//	filterSifra  apply OdSifre/DoSifre (analitika only)
//	filterVkonta only vkonta 1 and 2 (analitika and subsintetika)
func (s *BilansiResource) zakljucniInnerQuery(ctx context.Context, params domain.ZakljucniParams, printType, selectCols, groupByCols string, filterSifra, filterVkonta bool) (string, []any, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return "", nil, fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	aggregationSQL := fmt.Sprintf(`%s
		COALESCE(SUM(CASE WHEN fnal.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjedug,
		COALESCE(SUM(CASE WHEN fnal.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjepot,
		COALESCE(SUM(CASE WHEN fnal.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as prometdug,
		COALESCE(SUM(CASE WHEN fnal.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as prometpot
		FROM fpro
		inner join fnal on fnal.idfnal = fpro.idfnal `, "SELECT "+selectCols)

	innerQb := common.NewQueryBuilder(aggregationSQL, true)
	if hasGod {
		innerQb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		innerQb.AddEqual("fpro.kar", session.SelectedKar)
	}
	innerQb.AddCondition("fnal.danal", params.OdDatuma, ">=")
	innerQb.AddCondition("fnal.danal", params.DoDatuma, "<=")
	if params.OdKonta != "" {
		innerQb.AddCondition(fmt.Sprintf(" left(fpro.konto, %d)", s.cfg.NDuzSint), params.OdKonta, ">=")
	}
	if params.DoKonta != "" {
		innerQb.AddCondition(fmt.Sprintf(" left(fpro.konto, %d)", s.cfg.NDuzSint), params.DoKonta, "<=")
	}
	// Apply Klasa 9 filter only for printing, not for data retrieval for processing
	if printType == common.TipStampePrint {
		if params.Klasa9 == "false" {
			innerQb.AddCustomCondition("fpro.konto NOT LIKE '9%'")
		}
	}
	if filterSifra {
		if params.OdSifre != "" {
			innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.OdSifre, ">=")
		}
		if params.DoSifre != "" {
			innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.DoSifre, "<=")
		}
	}
	if filterVkonta {
		innerQb.AddIn("fpro.vkonta", []interface{}{"1", "2"})
	}
	// Filter by samosaprometom if needed (exclude zero saldo rows)
	if params.SamosaPrometom == "true" {
		innerQb.AddHaving("((COALESCE(SUM(CASE WHEN fnal.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fnal.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0)) != 0 OR (COALESCE(SUM(CASE WHEN fnal.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fnal.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0)) != 0)")
	}

	innerQb.AddGroupBy(groupByCols)
	innerSql, innerArgs := innerQb.Build()
	return innerSql, innerArgs, nil
}

// zakljucniNeedsName reports whether the account name (join / lookup subquery) is needed:
// always for the data and print pass, and on the totals pass only when the search matches
// on the name.
func zakljucniNeedsName(getTotalRecords bool, searchText string) bool {
	return !getTotalRecords || searchText != ""
}

// zakljucniNameLookupJoin builds the DISTINCT ON subquery that resolves a konto (or konto
// prefix) to the naziv of the newest fkpl row, wrapped in the join clause of the outer query.
//
// The subquery is embedded AFTER the inner query, so its $n placeholders must continue after
// the inner ones: innerArgs seed the builder, and the returned args are therefore
// innerArgs + the subquery's own values - in SQL text order.
func zakljucniNameLookupJoin(nameKey string, vkonta int, innerArgs []any) (string, []any) {
	subQb := common.NewQueryBuilder(fmt.Sprintf(`SELECT DISTINCT ON (%s) %s as konto_key, naziv FROM fkpl`, nameKey, nameKey), true)
	subQb.AddArgs(innerArgs...)
	subQb.AddEqual("vkonta", vkonta)
	subQb.AddOrderBy(fmt.Sprintf("%s, god DESC, kar DESC", nameKey))
	subSql, args := subQb.Build()
	return fmt.Sprintf(`left join (%s) fkpl_data on fkpl_data.konto_key = agg.konto`, subSql), args
}

// zakljucniOuterQueryBuilder wraps the inner aggregation in the OUTER query that adds the
// account name. Without a join the name stays empty (”), which is fine for the totals pass -
// those rows are only counted and summed.
func zakljucniOuterQueryBuilder(innerSql, nameJoin, nameExpr string, args []any, needName bool) *common.QueryBuilder {
	outerSelect := "agg.*, '' as naziv"
	if needName {
		outerSelect = fmt.Sprintf("agg.*, COALESCE(%s, '') as naziv", nameExpr)
	} else {
		nameJoin = ""
	}

	qb := common.NewQueryBuilder(fmt.Sprintf(`SELECT %s FROM (%s) agg`, outerSelect, innerSql), true)
	if nameJoin != "" {
		qb.AddJoin(nameJoin)
	}
	qb.AddArgs(args...) // keeps the $n numbering aligned with the embedded inner SQL
	return qb
}

// executeZakljucniQuery adds the search filter, ordering and pagination to the outer query,
// runs it and - on the totals pass - fills the pagination counters and the totals row.
func (s *BilansiResource) executeZakljucniQuery(ctx context.Context, tbl *domain.TableData, qb *common.QueryBuilder, args []any, o zakljucniQueryOptions) (*[]domain.FproDto, error) {
	// Add search filter if provided (all fields share one placeholder)
	if o.params.SearchText != "" {
		searchFields := []string{o.nameExpr, "agg.konto"}
		if o.withSifra {
			searchFields = append(searchFields, "agg.sifra")
		}
		qb.AddCustomCondition(zakljucniSearchCondition(len(args)+1, searchFields...), o.params.SearchText)
	}

	// The totals pass needs neither ORDER BY nor LIMIT/OFFSET (rows are only summed).
	if !o.getTotalRecords {
		qb.AddOrderBy(o.orderByCols)
		if o.printType == common.TipStampePreview { // printing/processing fetches the whole report
			qb.SetLimit(o.pageSize)
			qb.SetOffset((o.currentPage - 1) * o.pageSize)
		}
	}

	outerSql, queryArgs := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, outerSql, "", queryArgs, "", "")
	if err != nil {
		return nil, err
	}

	// Count filtered items if needed for total records. The rows were fetched by the totals
	// pass (no ORDER BY / name lookup), so they are only counted and summed.
	if o.printType == common.TipStampePreview && o.getTotalRecords {
		setZakljucniTotals(tbl, *entities, o.pageSize)
	}
	return entities, nil
}

// setZakljucniTotals fills the pagination counters and the totals row (columns 4..9) of the
// Zakljucni list grid from the unpaged aggregate rows.
func setZakljucniTotals(tbl *domain.TableData, entities []domain.FproDto, pageSize int) {
	common.SetTableTotalRecords(tbl, len(entities), pageSize)
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
	hasGod, hasKar := s.bilsRepo.GetHasGodHasKar()

	// Get Totals for table
	qbTotals := common.NewQueryBuilder(`SELECT
			COALESCE(SUM(bils.tgod), 0) as tgod,
			COALESCE(SUM(bils.tgodh), 0) as tgodh,
			COALESCE(SUM(bils.pgod), 0) as pgod,
			COALESCE(SUM(bils.pgodh), 0) as pgodh,
			COALESCE(SUM(bils.pgodps), 0) as pgodps,
			COALESCE(SUM(bils.pgodhps), 0) as pgodhps
			FROM bils`, true)

	qb := common.NewQueryBuilder(`SELECT 
			bils.bilsid, bils.rbr, bils.grac, bils.nazp, bils.aop, 
			bils.konta, bils.tgod, bils.pgod,
			bils.nipo, bils.tgodh, bils.pgodh,
			bils.pozic_1, bils.pozic_2, bils.pozic_3, bils.pozic_4, 
			bils.pozic_5, bils.pozic_6, bils.pozic_7, bils.pozic_8,
			bils.pozic_9, bils.pozic_10, bils.pozic_11, bils.pozic_12, bils.skraceni FROM bils`, true)

	// Add same filters to both queries
	if hasGod {
		qb.AddEqual("bils.god", session.SelectedGod)
		qbTotals.AddEqual("bils.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("bils.kar", session.SelectedKar)
		qbTotals.AddEqual("bils.kar", session.SelectedKar)
	}

	qbTotalQry, qbTotalArgs := qbTotals.Build()
	totalsResult, err := s.bilsRepo.GetAllCustom(ctx, qbTotalQry, "", qbTotalArgs, "", "")
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

	// Add search filter if provided
	if searchText != "" {
		nbrParam := len(qb.GetArgs()) + 1
		customCondition := fmt.Sprintf(`bilu.nazp ilike '%%' || $%d || '%%'" 
		OR bilu.grac ilike '%%' || $%d || '%%'"
		OR bilu.konta ilike '%%' || $%d || '%%'"
		OR bilu.aop ilike '%%' || $%d || '%%'`, nbrParam, nbrParam, nbrParam, nbrParam)
		qb.AddCustomCondition(customCondition, searchText)
		qbTotals.AddCustomCondition(customCondition, searchText)
	}

	sqlQuery, args := qb.Build()
	entities, err := s.bilsRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
	sqlQuery, args = qbTotals.Build()
	entitiesTotal, err := s.bilsRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
// Performance: all BILS rows loaded into memory once (keyed by AOP); fpro aggregates
// cached per unique (sKonto, god, kar, odMes, doMes) to avoid repeated DB round-trips.
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
	resetAllQb := common.NewQueryBuilder("update bils set tgod = 0, tgodh = 0", true)
	if hasGod {
		resetAllQb.AddEqual("god", god)
	}
	if hasKar {
		resetAllQb.AddEqual("kar", kar)
	}
	resetAllSql, resetAllArgs := resetAllQb.Build()
	if _, err := s.bilsRepo.DB.ExecContext(ctx, resetAllSql, resetAllArgs...); err != nil {
		return err
	}

	// Summary rows (konta=''): also reset pgod, pgodh, pgodps, pgodhps
	resetSummaryQb := common.NewQueryBuilder("update bils set pgod = 0, pgodh = 0, pgodps = 0, pgodhps = 0", true)
	if hasGod {
		resetSummaryQb.AddEqual("god", god)
	}
	if hasKar {
		resetSummaryQb.AddEqual("kar", kar)
	}
	resetSummaryQb.AddCustomCondition("konta = ''")
	resetSummarySql, resetSummaryArgs := resetSummaryQb.Build()
	if _, err := s.bilsRepo.DB.ExecContext(ctx, resetSummarySql, resetSummaryArgs...); err != nil {
		return err
	}

	// Leaf rows + lPGODizPS=true: also reset pgod, pgodh
	if lPGODizPS {
		resetLeafQb := common.NewQueryBuilder("update bils set pgod = 0, pgodh = 0", true)
		if hasGod {
			resetLeafQb.AddEqual("god", god)
		}
		if hasKar {
			resetLeafQb.AddEqual("kar", kar)
		}
		resetLeafQb.AddCustomCondition("konta != ''")
		resetLeafSql, resetLeafArgs := resetLeafQb.Build()
		if _, err := s.bilsRepo.DB.ExecContext(ctx, resetLeafSql, resetLeafArgs...); err != nil {
			return err
		}
	}

	// ── STEP 2: Get max NIPO ────────────────────────────────────────────────────
	maxNipoQb := common.NewQueryBuilder(`select coalesce(max(nipo), 1) as nipo from bils`, true)
	if hasGod {
		maxNipoQb.AddEqual("god", god)
	}
	if hasKar {
		maxNipoQb.AddEqual("kar", kar)
	}
	maxNipoSql, maxNipoArgs := maxNipoQb.Build()
	maxNipoRecords, err := s.bilsRepo.GetAllCustom(ctx, maxNipoSql, "", maxNipoArgs, "", "")
	if err != nil {
		return err
	}
	maxk := int16(1)
	if maxNipoRecords != nil && len(*maxNipoRecords) > 0 && (*maxNipoRecords)[0].NiPo > 1 {
		maxk = (*maxNipoRecords)[0].NiPo
	}

	// ── STEP 3: Load ALL bils rows into memory once, keyed by AOP ──────────────
	allBilsQb := common.NewQueryBuilder(`select bilsid, aop, konta, tgod, pgod, pgodps, tgodh, pgodh, pgodhps, nipo, skraceni,
		pozic_1, pozic_2, pozic_3, pozic_4, pozic_5, pozic_6,
		pozic_7, pozic_8, pozic_9, pozic_10, pozic_11, pozic_12 from bils`, true)
	if hasGod {
		allBilsQb.AddEqual("god", god)
	}
	if hasKar {
		allBilsQb.AddEqual("kar", kar)
	}
	allBilsSql, allBilsArgs := allBilsQb.Build()
	allBilsRecords, err := s.bilsRepo.GetAllCustom(ctx, allBilsSql, "", allBilsArgs, "", "")
	if err != nil {
		return err
	}

	bilsByAop := make(map[int]*domain.Bils, len(*allBilsRecords))
	for i := range *allBilsRecords {
		row := &(*allBilsRecords)[i]
		bilsByAop[row.AOP] = row
	}

	// fpro aggregate cache: key = "sKonto:god:kar:odMes:doMes"
	// Returns (openingDug, openingPot, monthlyDug, monthlyPot) — avoids repeated DB round-trips
	// for rows sharing the same account prefix.
	type fsalAggResult struct{ pDug, pPot, mDug, mPot float64 }
	fsalCache := make(map[string]fsalAggResult)

	cachedFsalAggregate := func(g, k, od, do int, sKonto string) (float64, float64, float64, float64) {
		key := fmt.Sprintf("%s:%d:%d:%d:%d", sKonto, g, k, od, do)
		if v, ok := fsalCache[key]; ok {
			return v.pDug, v.pPot, v.mDug, v.mPot
		}
		pDug, pPot, mDug, mPot := s.queryFproBilsAggregate(ctx, hasGod, hasKar, g, k, od, do, sKonto)
		fsalCache[key] = fsalAggResult{pDug, pPot, mDug, mPot}
		return pDug, pPot, mDug, mPot
	}

	// ── STEP 4: Process each NIPO level ────────────────────────────────────────
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
				dVred, dVredPS := s.obradaKontaBilsCached(bils.Konta, god, kar, odMes, doMes, cachedFsalAggregate)
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

			// Single UPDATE per row
			fields := []domain.Fields{
				{Name: "tgod", Value: fmt.Sprintf("%v", bils.TGod)},
				{Name: "tgodh", Value: fmt.Sprintf("%v", bils.TGodH)},
				{Name: "pgod", Value: fmt.Sprintf("%v", bils.PGod)},
				{Name: "pgodh", Value: fmt.Sprintf("%v", bils.PGodH)},
				{Name: "pgodps", Value: fmt.Sprintf("%v", bils.PGodPS)},
				{Name: "pgodhps", Value: fmt.Sprintf("%v", bils.PGodHPS)},
			}
			if err := s.bilsRepo.Update(ctx, bils, "bilsid", bils.BilsID, fields); err != nil {
				return err
			}
		}
	}

	// ── STEP 5: Fetch final results and populate table ──────────────────────────
	finalQb := common.NewQueryBuilder(`select bilsid, rbr, grac, nazp, aop, konta, napomena,
		tgod, pgod, pgodps, tgodh, pgodh, pgodhps, nipo, skraceni from bils`, true)
	if hasGod {
		finalQb.AddEqual("god", god)
	}
	if hasKar {
		finalQb.AddEqual("kar", kar)
	}
	if skraceni {
		finalQb.AddEqual("skraceni", 1)
	}
	finalQb.AddOrderBy("aop asc")
	finalSql, finalArgs := finalQb.Build()

	finalRecords, err := s.bilsRepo.GetAllCustom(ctx, finalSql, "", finalArgs, "", "")
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

// queryFproBilsAggregate returns (openingDug, openingPot, monthlyDug, monthlyPot) from fpro
// for the given account prefix.
//
//	Opening balance = tipdok = '00'
//	Monthly movements = tipdok != '00', month-filtered by [odMes..doMes]; odMes=0 → full year.
//
// This is the BILS equivalent of queryFproAggregate (which only fetches movements for BILU).
func (s *BilansiResource) queryFproBilsAggregate(ctx context.Context, hasGod, hasKar bool, god, kar, odMes, doMes int, sKonto string) (pDug, pPot, mDug, mPot float64) {
	// Opening balance (tipdok = '00')
	qbP := common.NewQueryBuilder(`select
		coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) as dug,
		coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as pot
		from fpro`, true)
	if hasGod {
		qbP.AddEqual("god", god)
	}
	if hasKar {
		qbP.AddEqual("kar", kar)
	}
	qbP.AddCustomCondition("fnal.tipdok = '00'")
	qbP.AddLikeBegin("fnal.konto", sKonto)
	sqlP, argsP := qbP.Build()
	rowsP, err := s.fproRepo.GetAllCustom(ctx, sqlP, "", argsP, "", "")
	if err == nil && rowsP != nil && len(*rowsP) > 0 {
		pDug = (*rowsP)[0].Dug.Float64
		pPot = (*rowsP)[0].Pot.Float64
	}

	// Monthly movements (tipdok != '00', month range)
	qbM := common.NewQueryBuilder(`select
		coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) as dug,
		coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as pot
		from fpro fnal`, true)
	if hasGod {
		qbM.AddEqual("god", god)
	}
	if hasKar {
		qbM.AddEqual("kar", kar)
	}
	qbM.AddCustomCondition("fnal.tipdok != '00'")
	qbM.AddLikeBegin("fnal.konto", sKonto)
	if odMes > 0 {
		qbM.AddCondition("extract(month from danal)::int", odMes, ">=")
	}
	if doMes > 0 {
		qbM.AddCondition("extract(month from danal)::int", doMes, "<=")
	}
	sqlM, argsM := qbM.Build()
	rowsM, err := s.fproRepo.GetAllCustom(ctx, sqlM, "", argsM, "", "")
	if err == nil && rowsM != nil && len(*rowsM) > 0 {
		mDug = (*rowsM)[0].Dug.Float64
		mPot = (*rowsM)[0].Pot.Float64
	}
	return
}

// GetBilansStanjaZaStampu retrieves data for Bilans stanja (balance sheet) for printing
func (s *BilansiResource) GetBilansStanjaZaStampu(ctx context.Context, tbl *domain.TableData, tipStampe string, skraceni bool) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.bilsRepo.GetHasGodHasKar()
	// Get Totals for table
	qbTotals := common.NewQueryBuilder(`SELECT
			COALESCE(SUM(bils.tgod), 0) as tgod,
			COALESCE(SUM(bils.tgodh), 0) as tgodh,
			COALESCE(SUM(bils.pgod), 0) as pgod,
			COALESCE(SUM(bils.pgodh), 0) as pgodh,
			COALESCE(SUM(bils.pgodps), 0) as pgodps,
			COALESCE(SUM(bils.pgodhps), 0) as pgodhps
			FROM bils`, true)

	qb := common.NewQueryBuilder(`SELECT 
			bils.bilsid, bils.rbr, bils.grac, bils.nazp, bils.aop, 
			bils.napomena, bils.tgodh, bils.pgodh, bils.pgodhps,
			bils.tgod, bils.pgod, bils.pgodps, bils.nipo, bils.skraceni FROM bils`, true)

	// Add same filters to both queries
	if hasGod {
		qb.AddEqual("bils.god", session.SelectedGod)
		qbTotals.AddEqual("bils.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("bils.kar", session.SelectedKar)
		qbTotals.AddEqual("bils.kar", session.SelectedKar)
	}
	if skraceni {
		qb.AddEqual("bils.skraceni", 1)
		qbTotals.AddEqual("bils.skraceni", 1)
	}

	qbTotalQry, qbTotalArgs := qbTotals.Build()
	totalsResult, err := s.bilsRepo.GetAllCustom(ctx, qbTotalQry, "", qbTotalArgs, "", "")
	if err != nil {
		return err
	}
	tbl.HasTotals = true
	if totalsResult != nil && len(*totalsResult) > 0 {
		tbl.Totals = make([]string, len(tbl.Headers))
		tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
		tbl.Totals[5] = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGod, 2)
		tbl.Totals[6] = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGod, 2)
		tbl.Totals[7] = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodPS, 2)
		tbl.Totals[9] = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGodH, 2)
		tbl.Totals[10] = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodH, 2)
		tbl.Totals[11] = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodHPS, 2)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.bilsRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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

	return nil
}
func (s *BilansiResource) GetBilansUspeha(ctx context.Context, tbl *domain.TableData, totals *domain.BilansiTotals, searchText string, skraceni bool) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.bilsRepo.GetHasGodHasKar()

	// If only totals are needed, we can sum directly in SQL without fetching all records
	qbTotals := common.NewQueryBuilder(`SELECT
			COALESCE(SUM(bilu.tgod), 0) as tgod,
			COALESCE(SUM(bilu.tgodh), 0) as tgodh,
			COALESCE(SUM(bilu.pgod), 0) as pgod,
			COALESCE(SUM(bilu.pgodh), 0) as pgodh
			FROM bilu`, true)

	qb := common.NewQueryBuilder(`SELECT 
			bilu.biluid, bilu.rbr, bilu.grac, bilu.nazp, bilu.aop, 
			bilu.konta, bilu.tgod, bilu.pgod,
			bilu.nipo, bilu.tgodh, bilu.pgodh,
			bilu.pozic_1, bilu.pozic_2, bilu.pozic_3, bilu.pozic_4, 
			bilu.pozic_5, bilu.pozic_6, bilu.pozic_7, bilu.pozic_8,
			bilu.pozic_9, bilu.pozic_10, bilu.pozic_11, bilu.pozic_12, bilu.skraceni FROM bilu`, true)

	// Add filters directly in query
	if hasGod {
		qb.AddEqual("bilu.god", session.SelectedGod)
		qbTotals.AddEqual("bilu.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("bilu.kar", session.SelectedKar)
		qbTotals.AddEqual("bilu.kar", session.SelectedKar)
	}
	// Filter by skraceni flag if set
	if skraceni {
		qb.AddEqual("bilu.skraceni", 1)
		qbTotals.AddEqual("bilu.skraceni", 1)
	}
	qbTotalQry, qbTotalArgs := qbTotals.Build()
	totalsResult, err := s.biluRepo.GetAllCustom(ctx, qbTotalQry, "", qbTotalArgs, "", "")
	if err != nil {
		return err
	}
	if totalsResult != nil && len(*totalsResult) > 0 {
		totals.TekGod = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGod, 2)
		totals.PrethGod = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGod, 2)
		totals.TekGodH = common.FormatNumberWithSystemLocale((*totalsResult)[0].TGodH, 0)
		totals.PrethGodH = common.FormatNumberWithSystemLocale((*totalsResult)[0].PGodH, 0)
	}

	// Add search filter if provided
	if searchText != "" {
		nbrParam := len(qb.GetArgs()) + 1
		customCondition := fmt.Sprintf(`bilu.nazp ilike '%%' || $%d || '%%'" 
		OR bilu.grac ilike '%%' || $%d || '%%'"
		OR bilu.konta ilike '%%' || $%d || '%%'"
		OR bilu.aop ilike '%%' || $%d || '%%'`, nbrParam, nbrParam, nbrParam, nbrParam)
		qb.AddCustomCondition(customCondition, searchText)
		qbTotals.AddCustomCondition(customCondition, searchText)
	}

	qb.AddOrderBy("bilu.rbr ASC")

	// Execute query and get entities
	sqlQuery, args := qb.Build()
	entities, err := s.biluRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
	sqlQuery, args = qbTotals.Build()
	entitiesTotal, err := s.bilsRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
	resetAllQb := common.NewQueryBuilder("update bilu set tgod = 0, tgodh = 0", true)
	if hasGod {
		resetAllQb.AddEqual("god", god)
	}
	if hasKar {
		resetAllQb.AddEqual("kar", kar)
	}
	resetAllSql, resetAllArgs := resetAllQb.Build()
	if _, err := s.biluRepo.DB.ExecContext(ctx, resetAllSql, resetAllArgs...); err != nil {
		return err
	}

	resetSummaryQb := common.NewQueryBuilder("update bilu set pgod = 0, pgodh = 0", true)
	if hasGod {
		resetSummaryQb.AddEqual("god", god)
	}
	if hasKar {
		resetSummaryQb.AddEqual("kar", kar)
	}
	resetSummaryQb.AddCustomCondition("konta = ''")
	resetSummarySql, resetSummaryArgs := resetSummaryQb.Build()
	if _, err := s.biluRepo.DB.ExecContext(ctx, resetSummarySql, resetSummaryArgs...); err != nil {
		return err
	}

	// STEP 2: Get max NIPO value
	maxNipoQb := common.NewQueryBuilder(`select coalesce(max(nipo), 1) as nipo from bilu`, true)
	if hasGod {
		maxNipoQb.AddEqual("god", god)
	}
	if hasKar {
		maxNipoQb.AddEqual("kar", kar)
	}
	maxNipoSql, maxNipoArgs := maxNipoQb.Build()
	maxNipoRecords, err := s.biluRepo.GetAllCustom(ctx, maxNipoSql, "", maxNipoArgs, "", "")
	if err != nil {
		return err
	}
	maxk := int16(1)
	if maxNipoRecords != nil && len(*maxNipoRecords) > 0 && (*maxNipoRecords)[0].NiPo > 1 {
		maxk = (*maxNipoRecords)[0].NiPo
	}

	// STEP 3: Load ALL bilu rows into memory once, keyed by AOP.
	// This eliminates up to N×12 per-row DB lookups during summary aggregation.
	allBiluQb := common.NewQueryBuilder(`select biluid, aop, konta, tgod, pgod, tgodh, pgodh, nipo, skraceni,
			pozic_1, pozic_2, pozic_3, pozic_4, pozic_5, pozic_6,
			pozic_7, pozic_8, pozic_9, pozic_10, pozic_11, pozic_12 from bilu`, true)
	if hasGod {
		allBiluQb.AddEqual("god", god)
	}
	if hasKar {
		allBiluQb.AddEqual("kar", kar)
	}
	allBiluSql, allBiluArgs := allBiluQb.Build()
	allBiluRecords, err := s.biluRepo.GetAllCustom(ctx, allBiluSql, "", allBiluArgs, "", "")
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

	cachedFproAggregate := func(k, r, od, do int, sKonto string) (float64, float64) {
		key := fmt.Sprintf("%s:%d:%d:%d:%d", sKonto, k, r, od, do)
		if v, ok := fproCache[key]; ok {
			return v[0], v[1]
		}
		dug, pot := s.queryFproAggregate(ctx, hasGod, hasKar, k, r, od, do, sKonto)
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
			if err := s.biluRepo.Update(ctx, bilu, "biluid", bilu.BiluID, fields); err != nil {
				return err
			}
		}
	}

	// STEP 5: Fetch final results and populate table
	finalQb := common.NewQueryBuilder(`select biluid, rbr, grac, nazp, aop, konta, tgod, pgod, tgodh, pgodh, nipo, skraceni from bilu`, true)
	if hasGod {
		finalQb.AddEqual("god", god)
	}
	if hasKar {
		finalQb.AddEqual("kar", kar)
	}
	if skraceni {
		finalQb.AddEqual("skraceni", 1)
	}
	finalQb.AddOrderBy("aop asc")
	finalSql, finalArgs := finalQb.Build()

	finalRecords, err := s.biluRepo.GetAllCustom(ctx, finalSql, "", finalArgs, "", "")
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

	return nil
}

// GetBilansUspehaZaStampu retrieves data for Bilans uspeha (income statement) for printing
func (s *BilansiResource) GetBilansUspehaZaStampu(ctx context.Context, tbl *domain.TableData, tipStampe string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.bilsRepo.GetHasGodHasKar()
	var qb *common.QueryBuilder
	if tipStampe == common.TipStampePrint {
		qb = common.NewQueryBuilder(`SELECT 
			bilu.grac, bilu.nazp, bilu.aop, 
			bilu.napomena, bilu.tgodh, bilu.pgodh FROM bilu`, true)
	}
	if tipStampe == common.TipStampePreview {
		qb = common.NewQueryBuilder(`SELECT 
			bilu.grac, bilu.nazp, bilu.aop,   
			bilu.napomena, bilu.tgodh, bilu.pgodh, bilu.tgod, bilu.pgod, bilu.nipo, bilu.skraceni FROM bilu`, true)
	}
	// Add filters directly in query
	if hasGod {
		qb.AddEqual("bilu.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("bilu.kar", session.SelectedKar)
	}

	qb.AddOrderBy("bilu.rbr ASC")

	// Execute query and get entities
	sqlQuery, args := qb.Build()
	entities, err := s.biluRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
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
func (s *BilansiResource) obradaKonta(ctx context.Context, konta string, god, kar, odMes, doMes int, lPGODizPG bool) (float64, float64) {
	if konta == "" {
		return 0, 0
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
		dug, pot := s.queryFproAggregate(ctx, hasGod, hasKar, god, kar, odMes, doMes, sKonto)
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
			dugPG, potPG := s.queryFproAggregate(ctx, hasGod, hasKar, god-1, kar, 0, 0, sKonto)
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

	return xMSaldo, xPSaldo
}

// queryFproAggregate returns (dug, pot) totals from fpro for the given account prefix/code.
// queryFproAggregate returns (dug, pot) totals from fpro for the given account prefix.
// (e.g. a booking on "2042" was rolled up into FSAL records for "204" and "20").
// Here we replicate that by using prefix matching directly on fpro, so a query for
// sKonto="204" matches fpro rows with konto "2042", "20420", etc.
// odMes/doMes = 0 means full-year (no month filter).
func (s *BilansiResource) queryFproAggregate(ctx context.Context, hasGod, hasKar bool, god, kar, odMes, doMes int, sKonto string) (float64, float64) {
	qb := common.NewQueryBuilder(`select
			coalesce(sum(case when kat in (1,2) then iznos else 0 end), 0) as dug,
			coalesce(sum(case when kat in (3,4) then iznos else 0 end), 0) as pot
			from fpro`, true)
	if hasGod {
		qb.AddEqual("god", god)
	}
	if hasKar {
		qb.AddEqual("kar", kar)
	}

	// Only movements, not opening balance postings
	qb.AddCustomCondition("fnal.tipdok != '00'")

	// Prefix match covers all sub-accounts at every depth:
	// sKonto="20"  matches fpro konto "20","204","2042","20420", ...
	// sKonto="204" matches fpro konto "204","2042","20420", ...
	// sKonto="2042" matches fpro konto "2042","20420", ...
	// No vkonta filter — fpro transactions exist at the actual booking level only,
	qb.AddLikeBegin("konto", sKonto)

	// Exclude closing/transfer accounts at all levels
	qb.AddCustomCondition("konto != '599'")
	qb.AddCustomCondition("konto != '699'")
	qb.AddCustomCondition("left(konto, 4) != '5999'")
	qb.AddCustomCondition("left(konto, 4) != '6999'")

	// Month filter: omitted when odMes == 0 (full-year query)
	if odMes > 0 {
		qb.AddCondition("extract(month from danal)::int", odMes, ">=")
	}
	if doMes > 0 {
		qb.AddCondition("extract(month from danal)::int", doMes, "<=")
	}

	sql, args := qb.Build()
	rows, err := s.fproRepo.GetAllCustom(ctx, sql, "", args, "", "")
	if err != nil || rows == nil || len(*rows) == 0 {
		return 0, 0
	}
	row := (*rows)[0]
	return row.Dug.Float64, row.Pot.Float64
}

func (s *BilansiResource) getKontoNaziv(ctx context.Context, konto string) string {
	userSesion := domain.GetSessionFromStdContext(ctx)
	if userSesion == nil {
		return ""
	}
	qb := common.NewQueryBuilder("select naziv from fkpl", true)
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", userSesion.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSesion.SelectedKar)
	}
	qb.AddEqual("konto", konto)
	qb.AddCondition("vkonta", 1, ">")
	sql, args := qb.Build()
	rows, err := s.fkplRepo.GetAllCustom(ctx, sql, "", args, "", "")
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
		{Name: "grac", Label: "Grupa računa, račun", Width: "15", Field: "bilu.grac", SkipInSearch: false},
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
