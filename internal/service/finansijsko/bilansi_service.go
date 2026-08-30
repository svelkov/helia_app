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
	GetZakljucniTableFields() []domain.Fields
	GetBilansStanjaTableFields() []domain.Fields
	GetBilansStanjaStampaTableFields() []domain.Fields
	GetBilansUspehaTableFields() []domain.Fields
	GetBilansUspehaStampaTableFields() []domain.Fields
	GetZakljucniList(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error
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

// GetZakljucniList retrieves data for Zakljucni list (closing account balance)
func (s *BilansiResource) GetZakljucniList(ctx context.Context, tbl *domain.TableData, params domain.ZakljucniParams, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "", "", false, false, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	// hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	// // =============================================================================
	// // STEP 1: Define query structure based on tipLista (analytical level)
	// // =============================================================================

	// var config domain.QueryConfig
	// switch params.TipLista {
	// case "1": // Analitička (detailed by konto + sifra)
	// 	config = domain.QueryConfig{
	// 		SelectCols:  "fpro.konto, COALESCE(fpro.sifra, '') as sifra, fpro.idfkpl,",
	// 		GroupByCols: "fpro.konto, fpro.sifra, fpro.idfkpl",
	// 		OrderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC, COALESCE(NULLIF(agg.sifra, '')::numeric, 0) ASC",
	// 	}
	// case "2": // Sintetička (summary by konto only)
	// 	config = domain.QueryConfig{
	// 		SelectCols:  "fpro.konto,",
	// 		GroupByCols: "fpro.konto",
	// 		OrderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC",
	// 	}
	// case "3": // Grupa (summary by truncated konto)
	// 	config = domain.QueryConfig{
	// 		SelectCols:  fmt.Sprintf("LEFT(fpro.konto, %d) as konto,", s.cfg.NDuzSint),
	// 		GroupByCols: fmt.Sprintf("LEFT(fpro.konto, %d)", s.cfg.NDuzSint),
	// 		OrderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC",
	// 	}
	// }

	// // =============================================================================
	// // STEP 2: Build INNER aggregation query (fpro transactions grouped by account)
	// // =============================================================================
	// aggregationSQL := fmt.Sprintf(`%s
	// 	COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjedug,
	// 	COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjepot,
	// 	COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as prometdug,
	// 	COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as prometpot
	// 	FROM fpro`, fmt.Sprintf("SELECT %s", config.SelectCols))

	// innerQb := common.NewQueryBuilder(aggregationSQL, true)

	// // Add period filters (god/kar)
	// if hasGod {
	// 	innerQb.AddEqual("fpro.god", session.SelectedGod)
	// }
	// if hasKar {
	// 	innerQb.AddEqual("fpro.kar", session.SelectedKar)
	// }

	// // Add date range filters
	// innerQb.AddCondition("fpro.danal", params.OdDatuma, ">=")
	// innerQb.AddCondition("fpro.danal", params.DoDatuma, "<=")

	// // Add konta (account) range filters
	// if params.OdKonta != "" {
	// 	innerQb.AddCondition("COALESCE(NULLIF(fpro.konto, '')::numeric, 0)", params.OdKonta, ">=")
	// }
	// if params.DoKonta != "" {
	// 	innerQb.AddCondition("COALESCE(NULLIF(fpro.konto, '')::numeric, 0)", params.DoKonta, "<=")
	// }

	// // // Add specific filters
	// // if params.Klasa9 == "true" {
	// // 	innerQb.AddLikeBegin("fpro.konto", "9") // Filter for class 9 accounts only
	// // }

	// // Add sifra (code) range filters only for tipLista "1"
	// if params.TipLista == "1" {
	// 	if params.OdSifre != "" {
	// 		innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.OdSifre, ">=")
	// 	}
	// 	if params.DoSifre != "" {
	// 		innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.DoSifre, "<=")
	// 	}
	// }

	// // Filter vkonta (account type): only 1 (assets) and 2 (liabilities) for tipLista 1 and 2
	// if params.TipLista == "1" || params.TipLista == "2" {
	// 	innerQb.AddIn("fpro.vkonta", []interface{}{"1", "2"})
	// }

	// // Apply grouping to aggregate transactions
	// innerQb.AddGroupBy(config.GroupByCols)

	// // Filter by samosaprometom if needed (exclude zero saldo rows)
	// if params.SamosaPrometom == "true" {
	// 	innerQb.AddHaving("((COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0)) != 0 OR (COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0)) != 0)")
	// }

	// // Build inner aggregation query
	// innerSql, innerArgs := innerQb.Build()

	// // =============================================================================
	// // STEP 3: Build OUTER query - join with account names from fkpl table
	// // =============================================================================
	// var outerSql string
	// var allArgs []interface{}

	// switch params.TipLista {
	// case "1": // Analitička - join directly to fkpl by idfkpl
	// 	// Build outer query builder with inner SQL embedded
	// 	outerQb := common.NewQueryBuilder(fmt.Sprintf(
	// 		`SELECT agg.*, COALESCE(fkpl.naziv, '') as naziv FROM (%s) agg`, innerSql), true)
	// 	outerQb.AddJoin("left join fkpl on fkpl.idfkpl = agg.idfkpl")
	// 	outerQb.AddOrderBy(config.OrderByCols)

	// 	// Add pagination using QueryBuilder
	// 	if !getTotalRecords {
	// 		outerQb.SetLimit(pageSize)
	// 		outerQb.SetOffset((currentPage - 1) * pageSize)
	// 	}
	// 	outerQb.AddArgs(innerArgs...) // Add inner query parameters
	// 	// Build outer query - it should not create new parameters since no WHERE conditions
	// 	outerSql, allArgs = outerQb.Build()

	// case "2": // Sintetička - get distinct naziv per konto
	// 	// Build distinct subquery for fkpl names
	// 	distinctQb := common.NewQueryBuilder(
	// 		`SELECT DISTINCT ON (konto) konto, naziv FROM fkpl`, true)
	// 	distinctQb.AddArgs(innerArgs...) // Add inner query parameters if needed for filtering
	// 	distinctQb.AddEqual("vkonta", 2)
	// 	distinctQb.AddOrderBy("konto, god DESC, kar DESC")
	// 	distinctSql, distinctArgs := distinctQb.Build()

	// 	// Build outer query builder with inner SQL and distinct subquery embedded
	// 	outerQb := common.NewQueryBuilder(fmt.Sprintf(
	// 		`SELECT agg.*, COALESCE(fkpl_data.naziv, '') as naziv FROM (%s) agg LEFT JOIN (%s) fkpl_data ON fkpl_data.konto = agg.konto`,
	// 		innerSql, distinctSql), true)
	// 	outerQb.AddOrderBy(config.OrderByCols)

	// 	// Add pagination using QueryBuilder
	// 	if !getTotalRecords {
	// 		outerQb.SetLimit(pageSize)
	// 		outerQb.SetOffset((currentPage - 1) * pageSize)
	// 	}
	// 	outerQb.AddArgs(distinctArgs...) // Add distinct query parameters (if any)
	// 	// Add inner query parameters
	// 	// Build outer query - combine parameters from inner and distinct queries only
	// 	outerSql, allArgs = outerQb.Build()

	// case "3": // Grupa - get distinct naziv per truncated konto group
	// 	// Build distinct subquery for fkpl names grouped by truncated konto
	// 	distinctQb := common.NewQueryBuilder(fmt.Sprintf(
	// 		`SELECT DISTINCT ON (LEFT(konto, %d)) LEFT(konto, %d) as konto_trunc, naziv FROM fkpl`,
	// 		s.cfg.NDuzSint, s.cfg.NDuzSint), true)
	// 	distinctQb.AddArgs(innerArgs...)
	// 	distinctQb.AddEqual("vkonta", 3)
	// 	distinctQb.AddOrderBy(fmt.Sprintf("LEFT(konto, %d), god DESC, kar DESC", s.cfg.NDuzSint))
	// 	distinctSql, distinctArgs := distinctQb.Build()

	// 	// Build outer query builder with inner SQL and distinct subquery embedded
	// 	outerQb := common.NewQueryBuilder(fmt.Sprintf(
	// 		`SELECT agg.*, COALESCE(fkpl_data.naziv, '') as naziv FROM (%s) agg LEFT JOIN (%s) fkpl_data ON fkpl_data.konto_trunc = agg.konto`,
	// 		innerSql, distinctSql), true)
	// 	outerQb.AddOrderBy(config.OrderByCols)

	// 	// Add pagination using QueryBuilder
	// 	if !getTotalRecords {
	// 		outerQb.SetLimit(pageSize)
	// 		outerQb.SetOffset((currentPage - 1) * pageSize)
	// 	}
	// 	outerQb.AddArgs(distinctArgs...) // Add distinct query parameters (if any)
	// 	// Build outer query - combine parameters from inner and distinct queries only
	// 	outerSql, allArgs = outerQb.Build()
	// }

	// // =============================================================================
	// // STEP 4: Execute query with pagination applied at SQL level
	// // =============================================================================
	// //fmt.Println("SQL Query for Zakljucni list:", outerSql, allArgs)
	// entities, err := s.fproRepo.GetAllCustom(ctx, outerSql, "", allArgs, "", "")
	// if err != nil {
	// 	return err
	// }
	entities, err := s.getZakljucniQuery(ctx, tbl, getTotalRecords, params, "O", pageSize, currentPage) // "O" for data version of the query (with pagination and total count)
	if err != nil {
		return err
	}

	// Populate table rows - pagination is already applied at SQL level
	start := (currentPage - 1) * pageSize
	rowNum := 1

	for _, entity := range *entities {
		// Calculate combined saldo
		saldoDug := entity.PocStanjeDug + entity.PrometDug
		saldoPot := entity.PocStanjePot + entity.PrometPot

		if saldoDug > saldoPot {
			saldoDug = saldoDug - saldoPot
			saldoPot = 0
		} else if saldoPot > saldoDug {
			saldoPot = saldoPot - saldoDug
			saldoDug = 0
		} else {
			saldoDug = 0
			saldoPot = 0
		}
		// Build fields based on tipLista
		fields := []string{
			fmt.Sprintf("%d", start+rowNum),
			entity.Konto,
		}
		sifra := ""
		// Add sifra only for tipLista "1"
		if params.TipLista == "1" {
			sifra = entity.Sifra
		}
		fields = append(fields, []string{
			sifra,
			entity.Naziv,
			common.FormatNumberWithSystemLocale(entity.PocStanjeDug, 2),
			common.FormatNumberWithSystemLocale(entity.PocStanjePot, 2),
			common.FormatNumberWithSystemLocale(entity.PrometDug, 2),
			common.FormatNumberWithSystemLocale(entity.PrometPot, 2),
			common.FormatNumberWithSystemLocale(saldoDug, 2),
			common.FormatNumberWithSystemLocale(saldoPot, 2),
		}...)

		tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)

		rowNum++
	}

	// Set table headers
	tbl.Headers = s.GetZakljucniTableFields()

	return nil
}

// GetZakljucniListZaStampu returns all zakljucni list rows for printing, including embedded
// subgroup (sintetika) and group (group prefix) total rows with row-type markers.
// Subgroup = same konto value; group = konto[0:nDuzSint].
func (s *BilansiResource) GetZakljucniListZaStampu(ctx context.Context, tbl, tblSummary *domain.TableData, params domain.ZakljucniParams, nDuzSint int) error {

	entities, err := s.getZakljucniQuery(ctx, nil, false, params, "S", 0, 0) // "S" for stampa/print version of the query (no pagination, no total count)
	if err != nil {
		return err
	}
	// ── Build rows with embedded group total rows ──────────────────────────────
	// Row [0] encodes type: "1","2",... = data, "G1".."G4" = group totals, "T" = grand total.
	//
	// Group levels per tipLista (innermost → outermost):
	//   tipLista=1 (analitika):    G1=subsintetika(full konto), G2=sintetika(nDuzSint), G3=grupa(2), G4=klasa(1)
	//   tipLista=2 (subsintetika): G1=sintetika(nDuzSint), G2=grupa(2), G3=klasa(1)
	//   tipLista=3 (sintetika):    G1=grupa(2), G2=klasa(1)
	formatRow := func(marker, konto, sifra, naziv string, pstDug, pstPot, promDug, promPot float64) domain.TableRow {
		ukupDug := pstDug + promDug
		ukupPot := pstPot + promPot
		saldoDug := ukupDug
		saldoPot := ukupPot
		if saldoDug > saldoPot {
			saldoDug -= saldoPot
			saldoPot = 0
		} else if saldoPot > saldoDug {
			saldoPot -= saldoDug
			saldoDug = 0
		} else {
			saldoDug, saldoPot = 0, 0
		}
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

	// groupLevel describes one aggregation tier.
	type groupLevel struct {
		marker string // "G1", "G2", "G3", or "G4"
		keyLen int    // 0 = full konto; >0 = konto[:keyLen]
	}
	keyOf := func(konto string, keyLen int) string {
		if keyLen <= 0 || keyLen >= len(konto) {
			return konto
		}
		return konto[:keyLen]
	}

	var levels []groupLevel
	switch params.TipLista {
	case "1": // analitika: 4 levels
		levels = []groupLevel{
			{"G1", 4},        // subsintetika = full konto (e.g. "0112")
			{"G2", nDuzSint}, // sintetika     (e.g. "011", nDuzSint=3)
			{"G3", 2},        // grupa         (e.g. "01")
			{"G4", 1},        // klasa         (e.g. "0")
		}
	case "2": // subsintetika: 3 levels
		levels = []groupLevel{
			{"G1", nDuzSint}, // sintetika     (e.g. "011")
			{"G2", 2},        // grupa         (e.g. "01")
			{"G3", 1},        // klasa         (e.g. "0")
		}
	default: // "3" sintetika: 2 levels
		levels = []groupLevel{
			{"G1", 2}, // grupa  (e.g. "01")
			{"G2", 1}, // klasa  (e.g. "0")
		}
	}

	// Per-level running state.
	type levelState struct {
		pstDug, pstPot   float64
		promDug, promPot float64
		prevKey          string
		naziv            string
	}
	states := make([]levelState, len(levels))

	// Per-klasa summary state (key = first digit of konto).
	type klasaState struct {
		pstDug, pstPot   float64
		promDug, promPot float64
	}
	klasaMap := make(map[string]*klasaState)
	var klasaKeys []string // preserve insertion order

	var grandPstDug, grandPstPot, grandPromDug, grandPromPot float64
	rbr := 1

	for _, ent := range *entities {
		if params.SamosaPrometom == "true" && ent.PrometDug == 0 && ent.PrometPot == 0 {
			continue // skip zero-promet rows if samosaprometom filter is on
		}
		// Compute key for each level.
		keys := make([]string, len(levels))
		for i, lv := range levels {
			keys[i] = keyOf(ent.Konto, lv.keyLen)
		}

		// Find the outermost (highest index) level whose key changed.
		changedAt := -1
		for i := len(levels) - 1; i >= 0; i-- {
			if keys[i] != states[i].prevKey {
				changedAt = i
				break
			}
		}

		// Flush levels 0..changedAt (innermost first).
		if changedAt >= 0 {
			for i := 0; i <= changedAt; i++ {
				if states[i].prevKey != "" {
					tbl.Rows = append(tbl.Rows, formatRow(
						levels[i].marker, states[i].prevKey, "", states[i].naziv,
						states[i].pstDug, states[i].pstPot, states[i].promDug, states[i].promPot,
					))
				}
				states[i] = levelState{}
			}
		}

		// Data row.
		sifra := ""
		if params.TipLista == "1" {
			sifra = ent.Sifra
		}

		tbl.Rows = append(tbl.Rows, formatRow(
			fmt.Sprintf("%d", rbr), ent.Konto, sifra, ent.Naziv,
			ent.PocStanjeDug, ent.PocStanjePot, ent.PrometDug, ent.PrometPot,
		))
		rbr++

		// Accumulate into all levels.
		for i := range levels {
			states[i].pstDug += ent.PocStanjeDug
			states[i].pstPot += ent.PocStanjePot
			states[i].promDug += ent.PrometDug
			states[i].promPot += ent.PrometPot
			states[i].prevKey = keys[i]

			if params.TipLista == "1" {

				switch i {
				case 0:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto)
				case 1:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:nDuzSint])
				case 2:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:2])
				case 3:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:1])
				}
			}

			if params.TipLista == "2" {
				switch i {
				case 0:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:nDuzSint])
				case 1:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:2])
				case 2:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:1])
				}
			}
			if params.TipLista == "3" {
				switch i {
				case 0:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:2])
				case 1:
					states[i].naziv = s.getKontoNaziv(ctx, ent.Konto[:1])
				}
			}
			//states[i].naziv = ent.Naziv

		}
		grandPstDug += ent.PocStanjeDug
		grandPstPot += ent.PocStanjePot
		grandPromDug += ent.PrometDug
		grandPromPot += ent.PrometPot

		// Accumulate per-klasa (first digit of konto).
		if len(ent.Konto) > 0 {
			k := ent.Konto[:1]
			if _, ok := klasaMap[k]; !ok {
				klasaMap[k] = &klasaState{}
				klasaKeys = append(klasaKeys, k)
			}
			klasaMap[k].pstDug += ent.PocStanjeDug
			klasaMap[k].pstPot += ent.PocStanjePot
			klasaMap[k].promDug += ent.PrometDug
			klasaMap[k].promPot += ent.PrometPot
		}
	}

	// Flush remaining levels (innermost first).
	for i := 0; i < len(levels); i++ {
		if states[i].prevKey != "" {
			tbl.Rows = append(tbl.Rows, formatRow(
				levels[i].marker, states[i].prevKey, "", states[i].naziv,
				states[i].pstDug, states[i].pstPot, states[i].promDug, states[i].promPot,
			))
		}
	}

	// Grand total row.
	tbl.Rows = append(tbl.Rows, formatRow("T", "", "", i18n.GetInstance().Label("Ukupno"),
		grandPstDug, grandPstPot, grandPromDug, grandPromPot))

	// Populate tblSummary: one row per klasa (key = first digit of konto).
	for _, k := range klasaKeys {
		st := klasaMap[k]
		tblSummary.Rows = append(tblSummary.Rows, formatRow(
			"K", k, "", i18n.GetInstance().Label("klasa")+": "+k,
			st.pstDug, st.pstPot, st.promDug, st.promPot,
		))
	}
	// Grand total for summary.
	tblSummary.Rows = append(tblSummary.Rows, formatRow("T", "", "", "TOTAL:",
		grandPstDug, grandPstPot, grandPromDug, grandPromPot))

	return nil
}

// getZakljucniQuery builds and executes the SQL query for Zakljucni list based on the provided parameters and print type.
func (s *BilansiResource) getZakljucniQuery(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, params domain.ZakljucniParams, printType string, pageSize, currentPage int) (*[]domain.FproDto, error) {
	// printType: "S" - stampa , ili "O" - obrada
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	var config domain.QueryConfig
	switch params.TipLista {
	case "1":
		config = domain.QueryConfig{
			SelectCols:  "fpro.konto, COALESCE(fpro.sifra, '') as sifra, fpro.idfkpl,",
			GroupByCols: "fpro.konto, fpro.sifra, fpro.idfkpl",
			OrderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC, COALESCE(NULLIF(agg.sifra, '')::numeric, 0) ASC",
		}
	case "2":
		config = domain.QueryConfig{
			SelectCols:  "fpro.konto,",
			GroupByCols: "fpro.konto",
			OrderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC",
		}
	case "3":
		config = domain.QueryConfig{
			SelectCols:  fmt.Sprintf("LEFT(fpro.konto, %d) as konto,", s.cfg.NDuzSint),
			GroupByCols: fmt.Sprintf("LEFT(fpro.konto, %d)", s.cfg.NDuzSint),
			OrderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC",
		}
	default:
		return nil, fmt.Errorf("invalid tip_lista: %s", params.TipLista)
	}

	aggregationSQL := fmt.Sprintf(`%s
		COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjedug,
		COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjepot,
		COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as prometdug,
		COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as prometpot
		FROM fpro`, fmt.Sprintf("SELECT %s", config.SelectCols))

	innerQb := common.NewQueryBuilder(aggregationSQL, true)
	if hasGod {
		innerQb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		innerQb.AddEqual("fpro.kar", session.SelectedKar)
	}
	innerQb.AddCondition("fpro.danal", params.OdDatuma, ">=")
	innerQb.AddCondition("fpro.danal", params.DoDatuma, "<=")
	if params.OdKonta != "" {
		innerQb.AddCondition("COALESCE(NULLIF(fpro.konto, '')::numeric, 0)", params.OdKonta, ">=")
	}
	if params.DoKonta != "" {
		innerQb.AddCondition("COALESCE(NULLIF(fpro.konto, '')::numeric, 0)", params.DoKonta, "<=")
	}
	// Apply Klasa 9 filter only for printing, not for data retrieval for processing
	if printType == "S" {
		if params.Klasa9 == "false" {
			innerQb.AddCustomCondition("fpro.konto NOT LIKE '9%'")
		}
	}
	if params.TipLista == "1" {
		if params.OdSifre != "" {
			innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.OdSifre, ">=")
		}
		if params.DoSifre != "" {
			innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", params.DoSifre, "<=")
		}
	}
	if params.TipLista == "1" || params.TipLista == "2" {
		innerQb.AddIn("fpro.vkonta", []interface{}{"1", "2"})
	}
	// Filter by samosaprometom if needed (exclude zero saldo rows)
	if params.SamosaPrometom == "true" {
		innerQb.AddHaving("((COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0)) != 0 OR (COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0)) != 0)")
	}

	innerQb.AddGroupBy(config.GroupByCols)
	innerSql, innerArgs := innerQb.Build()

	var outerSql string
	var allArgs []interface{}
	switch params.TipLista {
	case "1":
		outerQb := common.NewQueryBuilder(fmt.Sprintf(`SELECT agg.*, COALESCE(fkpl.naziv, '') as naziv FROM (%s) agg`, innerSql), true)
		outerQb.AddJoin("left join fkpl on fkpl.idfkpl = agg.idfkpl")
		outerQb.AddOrderBy(config.OrderByCols)
		outerQb.AddArgs(innerArgs...)
		// Add pagination using QueryBuilder
		if !getTotalRecords && printType == "O" { // Apply pagination only for data retrieval for processing, not for printing
			outerQb.SetLimit(pageSize)
			outerQb.SetOffset((currentPage - 1) * pageSize)
		}
		outerSql, allArgs = outerQb.Build()
	case "2":
		distinctQb := common.NewQueryBuilder(`SELECT DISTINCT ON (konto) konto, naziv FROM fkpl`, true)
		distinctQb.AddArgs(innerArgs...)
		distinctQb.AddEqual("vkonta", 2)
		distinctQb.AddOrderBy("konto, god DESC, kar DESC")
		distinctSql, distinctArgs := distinctQb.Build()
		outerQb := common.NewQueryBuilder(fmt.Sprintf(`SELECT agg.*, COALESCE(fkpl_data.naziv, '') as naziv FROM (%s) agg LEFT JOIN (%s) fkpl_data ON fkpl_data.konto = agg.konto`, innerSql, distinctSql), true)
		outerQb.AddOrderBy(config.OrderByCols)
		outerQb.AddArgs(distinctArgs...)
		// Add pagination using QueryBuilder
		if !getTotalRecords && printType == "O" { // Apply pagination only for data retrieval for processing, not for printing
			outerQb.SetLimit(pageSize)
			outerQb.SetOffset((currentPage - 1) * pageSize)
		}
		outerSql, allArgs = outerQb.Build()
	case "3":
		distinctQb := common.NewQueryBuilder(fmt.Sprintf(`SELECT DISTINCT ON (LEFT(konto, %d)) LEFT(konto, %d) as konto_trunc, naziv FROM fkpl`, s.cfg.NDuzSint, s.cfg.NDuzSint), true)
		distinctQb.AddArgs(innerArgs...)
		distinctQb.AddEqual("vkonta", 3)
		distinctQb.AddOrderBy(fmt.Sprintf("LEFT(konto, %d), god DESC, kar DESC", s.cfg.NDuzSint))
		distinctSql, distinctArgs := distinctQb.Build()
		outerQb := common.NewQueryBuilder(fmt.Sprintf(`SELECT agg.*, COALESCE(fkpl_data.naziv, '') as naziv FROM (%s) agg LEFT JOIN (%s) fkpl_data ON fkpl_data.konto_trunc = agg.konto`, innerSql, distinctSql), true)
		outerQb.AddOrderBy(config.OrderByCols)
		outerQb.AddArgs(distinctArgs...)
		// Add pagination using QueryBuilder
		if !getTotalRecords && printType == "O" { // Apply pagination only for data retrieval for processing, not for printing
			outerQb.SetLimit(pageSize)
			outerQb.SetOffset((currentPage - 1) * pageSize)
		}
		outerSql, allArgs = outerQb.Build()
	}

	entities, err := s.fproRepo.GetAllCustom(ctx, outerSql, "", allArgs, "", "")
	if err != nil {
		return nil, err
	}
	// Count filtered items if needed for total records
	if printType == "O" && getTotalRecords {
		count := len(*entities)
		common.SetTableTotalRecords(tbl, count, pageSize)
		tbl.Totals = make([]string, len(tbl.Headers))
		tbl.Totals[0] = i18n.GetInstance().Label("Ukupno") // Set label for totals column
		var pstPotTotal, pstDugTotal, dugTotal, potTotal float64

		for _, entity := range *entities {
			// Calculate combined saldo for counting total records
			pstDugTotal += entity.PocStanjeDug
			pstPotTotal += entity.PocStanjePot
			dugTotal += entity.PrometDug
			potTotal += entity.PrometPot
		}
		tbl.Totals[4] = common.FormatNumberWithSystemLocale(pstDugTotal, 2)                    // Total for Tekuća godina (can be calculated if needed)
		tbl.Totals[5] = common.FormatNumberWithSystemLocale(pstDugTotal, 2)                    // Total for Prethodna godina (can be calculated if needed)
		tbl.Totals[6] = common.FormatNumberWithSystemLocale(dugTotal, 2)                       // Total for Prethodna godina - početno stanje (can be calculated if needed)
		tbl.Totals[7] = common.FormatNumberWithSystemLocale(potTotal, 2)                       // Total for Promet potražuje (can be calculated if needed)
		tbl.Totals[8] = common.FormatNumberWithSystemLocale(math.Abs(pstDugTotal+dugTotal), 2) // Total for Saldo duguje (can be calculated if needed)
		tbl.Totals[9] = common.FormatNumberWithSystemLocale(math.Abs(pstPotTotal+potTotal), 2) // Total for Saldo potražuje (can be calculated if needed)
		return entities, nil
	}

	return entities, nil
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
	qbP.AddCustomCondition("tipdok = '00'")
	qbP.AddLikeBegin("konto", sKonto)
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
		from fpro`, true)
	if hasGod {
		qbM.AddEqual("god", god)
	}
	if hasKar {
		qbM.AddEqual("kar", kar)
	}
	qbM.AddCustomCondition("tipdok != '00'")
	qbM.AddLikeBegin("konto", sKonto)
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
	qb.AddCustomCondition("tipdok != '00'")

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
