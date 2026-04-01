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
	"strconv"
)

// BilansiService defines the interface for operations related to Bilansi (Balance Sheets).
type BilansiService interface {
	GetTableFields() []domain.Fields
	GetZakljucniTableFields() []domain.Fields
	GetBilansStanjaTableFields() []domain.Fields
	GetBilansUspehaTableFields() []domain.Fields
	GetZakljucniList(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetBilansStanja(ctx context.Context, tbl *domain.TableData) error
	StampaBilansStanja(ctx context.Context, tbl *domain.TableData, getTotalRecords, getOnlyTotals bool, pageSize, currentPage int, totals *domain.BilansiTotals) error
	GetBilansUspeha(ctx context.Context, tbl *domain.TableData) error
	GetByID(ctx context.Context, idField string, idValue int64) (*domain.Bils, error)
	Update(ctx context.Context, entity *domain.Bils, idField string, idValue interface{}, tableFields []domain.Fields) error
	Add(ctx context.Context, entity *domain.Bils, idField string, tableFields []domain.Fields) (int64, error)
	MapEntityToValues(entity *domain.Bils, tableFields []domain.Fields) []domain.Fields
	ValidateBilansStanja(entity *domain.Bils) ([]domain.FieldError, error)
	DeleteBilansStanja(ctx context.Context) error
	GetFieldCache() map[string]reflect.StructField
	// Bilu (Bilans Uspeha) methods
	GetByIDBilu(ctx context.Context, idField string, idValue int64) (*domain.Bilu, error)
	UpdateBilu(ctx context.Context, entity *domain.Bilu, idField string, idValue interface{}, tableFields []domain.Fields) error
	AddBilu(ctx context.Context, entity *domain.Bilu, idField string, tableFields []domain.Fields) (int64, error)
	MapEntityToValuesBilu(entity *domain.Bilu, tableFields []domain.Fields) []domain.Fields
	ValidateBilansUspeha(entity *domain.Bilu) []domain.FieldError
	DeleteBilansUspeha(ctx context.Context) error
	GetFieldCacheBilu() map[string]reflect.StructField
}

// BilansiResource implements the BilansiService interface.
type BilansiResource struct {
	biluService             *service.BaseService[domain.Bilu]
	biluRepo                *repository.BaseRepository[domain.Bilu]
	bilsService             *service.BaseService[domain.Bils]
	bilsRepo                *repository.BaseRepository[domain.Bils]
	fproRepo                *repository.BaseRepository[domain.FproDto]
	zakljucniTableFields    []domain.Fields
	bilansStanjaTableFields []domain.Fields
	bilansUspehaTableFields []domain.Fields
	cfg                     config.Config
}

func NewBilansiService(
	biluService *service.BaseService[domain.Bilu],
	biluRepo *repository.BaseRepository[domain.Bilu],
	bilsService *service.BaseService[domain.Bils],
	bilsRepo *repository.BaseRepository[domain.Bils],
	fproRepo *repository.BaseRepository[domain.FproDto],
	cfg config.Config,
) *BilansiResource {
	rs := &BilansiResource{
		biluService: biluService,
		biluRepo:    biluRepo,
		bilsService: bilsService,
		bilsRepo:    bilsRepo,
		fproRepo:    fproRepo,
		cfg:         cfg,
	}
	rs.setServiceFieldValues()
	return rs
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

// GetBilansUspehaTableFields returns the table field definitions for Bilans uspeha
func (s *BilansiResource) GetBilansUspehaTableFields() []domain.Fields {
	return s.bilansUspehaTableFields
}

// GetZakljucniList retrieves data for Zakljucni list (closing account balance)
func (s *BilansiResource) GetZakljucniList(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "", "", false, false, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Get parameters - from context (these would normally come from handler, but keeping simple for now)
	odKonta := ""
	doKonta := ""
	odSifre := ""
	doSifre := ""
	odDatuma := ""
	doDatuma := ""
	tipLista := "1" // default to analitička
	klasa9 := ""
	samosaprometom := ""

	// =============================================================================
	// STEP 1: Define query structure based on tipLista (analytical level)
	// =============================================================================
	type QueryConfig struct {
		selectCols  string // Columns to select in inner query
		groupByCols string // Columns to group by
		orderByCols string // Columns to order by
	}

	var config QueryConfig
	switch tipLista {
	case "1": // Analitička (detailed by konto + sifra)
		config = QueryConfig{
			selectCols:  "fpro.konto, fpro.sifra, fpro.idfkpl,",
			groupByCols: "fpro.konto, fpro.sifra, fpro.idfkpl",
			orderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC, COALESCE(NULLIF(agg.sifra, '')::numeric, 0) ASC",
		}
	case "2": // Sintetička (summary by konto only)
		config = QueryConfig{
			selectCols:  "fpro.konto,",
			groupByCols: "fpro.konto",
			orderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC",
		}
	case "3": // Grupa (summary by truncated konto)
		config = QueryConfig{
			selectCols:  fmt.Sprintf("LEFT(fpro.konto, %d) as konto,", s.cfg.NDuzSint),
			groupByCols: fmt.Sprintf("LEFT(fpro.konto, %d)", s.cfg.NDuzSint),
			orderByCols: "COALESCE(NULLIF(agg.konto, '')::numeric, 0) ASC",
		}
	}

	// =============================================================================
	// STEP 2: Build INNER aggregation query (fpro transactions grouped by account)
	// =============================================================================
	aggregationSQL := fmt.Sprintf(`%s
		COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjedug,
		COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as pocstanjepot,
		COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) as prometdug,
		COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) as prometpot
		FROM fpro`, fmt.Sprintf("SELECT %s", config.selectCols))

	innerQb := common.NewQueryBuilder(aggregationSQL, true)

	// Add period filters (god/kar)
	if hasGod {
		innerQb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		innerQb.AddEqual("fpro.kar", session.SelectedKar)
	}

	// Add date range filters
	innerQb.AddCondition("fpro.danal", odDatuma, ">=")
	innerQb.AddCondition("fpro.danal", doDatuma, "<=")

	// Add konta (account) range filters
	if odKonta != "" {
		innerQb.AddCondition("COALESCE(NULLIF(fpro.konto, '')::numeric, 0)", odKonta, ">=")
	}
	if doKonta != "" {
		innerQb.AddCondition("COALESCE(NULLIF(fpro.konto, '')::numeric, 0)", doKonta, "<=")
	}

	// Add specific filters
	if klasa9 == "true" {
		innerQb.AddLikeBegin("fpro.konto", "9") // Filter for class 9 accounts only
	}

	// Add sifra (code) range filters only for tipLista "1"
	if tipLista == "1" {
		if odSifre != "" {
			innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", odSifre, ">=")
		}
		if doSifre != "" {
			innerQb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", doSifre, "<=")
		}
	}

	// Filter vkonta (account type): only 1 (assets) and 2 (liabilities) for tipLista 1 and 2
	if tipLista == "1" || tipLista == "2" {
		innerQb.AddIn("fpro.vkonta", []interface{}{"1", "2"})
	}

	// Apply grouping to aggregate transactions
	innerQb.AddGroupBy(config.groupByCols)

	// Filter by samosaprometom if needed (exclude zero saldo rows)
	if samosaprometom == "true" {
		innerQb.AddHaving("((COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND (fpro.kat = 1 OR fpro.kat = 2) THEN fpro.iznos ELSE 0 END), 0)) != 0 OR (COALESCE(SUM(CASE WHEN fpro.tipdok = '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN fpro.tipdok != '00' AND fpro.kat NOT IN (1,2) THEN fpro.iznos ELSE 0 END), 0)) != 0)")
	}

	// Build inner aggregation query
	innerSql, innerArgs := innerQb.Build()

	// =============================================================================
	// STEP 3: Build OUTER query - join with account names from fkpl table
	// =============================================================================
	var outerSql string
	var allArgs []interface{}

	switch tipLista {
	case "1": // Analitička - join directly to fkpl by idfkpl
		// Build outer query builder with inner SQL embedded
		outerQb := common.NewQueryBuilder(fmt.Sprintf(
			`SELECT agg.*, COALESCE(fkpl.naziv, '') as naziv FROM (%s) agg`, innerSql), true)
		outerQb.AddJoin("left join fkpl on fkpl.idfkpl = agg.idfkpl")
		outerQb.AddOrderBy(config.orderByCols)

		// Add pagination using QueryBuilder
		if !getTotalRecords {
			outerQb.SetLimit(pageSize)
			outerQb.SetOffset((currentPage - 1) * pageSize)
		}
		outerQb.AddArgs(innerArgs...) // Add inner query parameters
		// Build outer query - it should not create new parameters since no WHERE conditions
		outerSql, allArgs = outerQb.Build()

	case "2": // Sintetička - get distinct naziv per konto
		// Build distinct subquery for fkpl names
		distinctQb := common.NewQueryBuilder(
			`SELECT DISTINCT ON (konto) konto, naziv FROM fkpl`, true)
		distinctQb.AddArgs(innerArgs...) // Add inner query parameters if needed for filtering
		distinctQb.AddEqual("vkonta", 2)
		distinctQb.AddOrderBy("konto, god DESC, kar DESC")
		distinctSql, distinctArgs := distinctQb.Build()

		// Build outer query builder with inner SQL and distinct subquery embedded
		outerQb := common.NewQueryBuilder(fmt.Sprintf(
			`SELECT agg.*, COALESCE(fkpl_data.naziv, '') as naziv FROM (%s) agg LEFT JOIN (%s) fkpl_data ON fkpl_data.konto = agg.konto`,
			innerSql, distinctSql), true)
		outerQb.AddOrderBy(config.orderByCols)

		// Add pagination using QueryBuilder
		if !getTotalRecords {
			outerQb.SetLimit(pageSize)
			outerQb.SetOffset((currentPage - 1) * pageSize)
		}
		outerQb.AddArgs(distinctArgs...) // Add distinct query parameters (if any)
		// Add inner query parameters
		// Build outer query - combine parameters from inner and distinct queries only
		outerSql, allArgs = outerQb.Build()

	case "3": // Grupa - get distinct naziv per truncated konto group
		// Build distinct subquery for fkpl names grouped by truncated konto
		distinctQb := common.NewQueryBuilder(fmt.Sprintf(
			`SELECT DISTINCT ON (LEFT(konto, %d)) LEFT(konto, %d) as konto_trunc, naziv FROM fkpl`,
			s.cfg.NDuzSint, s.cfg.NDuzSint), true)
		distinctQb.AddArgs(innerArgs...)
		distinctQb.AddEqual("vkonta", 3)
		distinctQb.AddOrderBy(fmt.Sprintf("LEFT(konto, %d), god DESC, kar DESC", s.cfg.NDuzSint))
		distinctSql, distinctArgs := distinctQb.Build()

		// Build outer query builder with inner SQL and distinct subquery embedded
		outerQb := common.NewQueryBuilder(fmt.Sprintf(
			`SELECT agg.*, COALESCE(fkpl_data.naziv, '') as naziv FROM (%s) agg LEFT JOIN (%s) fkpl_data ON fkpl_data.konto_trunc = agg.konto`,
			innerSql, distinctSql), true)
		outerQb.AddOrderBy(config.orderByCols)

		// Add pagination using QueryBuilder
		if !getTotalRecords {
			outerQb.SetLimit(pageSize)
			outerQb.SetOffset((currentPage - 1) * pageSize)
		}
		outerQb.AddArgs(distinctArgs...) // Add distinct query parameters (if any)
		// Build outer query - combine parameters from inner and distinct queries only
		outerSql, allArgs = outerQb.Build()
	}

	// =============================================================================
	// STEP 4: Execute query with pagination applied at SQL level
	// =============================================================================
	//fmt.Println("SQL Query for Zakljucni list:", outerSql, allArgs)
	entities, err := s.fproRepo.GetAllCustom(ctx, outerSql, "", allArgs, "", "")
	if err != nil {
		return err
	}

	// Count filtered items if needed for total records
	if getTotalRecords {
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
		return nil
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
		if tipLista == "1" {
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

// GetBilansStanja retrieves data for Bilans stanja (balance sheet)
func (s *BilansiResource) GetBilansStanja(ctx context.Context, tbl *domain.TableData) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.bilsRepo.GetHasGodHasKar()

	searchText := ""
	skraceni := false

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

	// Filter by skraceni flag if set
	if skraceni {
		qb.AddEqual("bils.skraceni", skraceni)
		qbTotals.AddEqual("bils.skraceni", skraceni)
	}

	// Add search filter if provided
	if searchText != "" {
		qb.AddLike("bils.nazp", interface{}(searchText))
		qbTotals.AddLike("bils.nazp", interface{}(searchText))
	}

	sqlQuery, args := qb.Build()
	entities, err := s.bilsRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			// Apply filtering logic on processed data
			if skraceni && entity.Skraceni != 1 {
				continue
			}

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
				fmt.Sprintf("%d", entity.AOP),
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

// GetBilansUspeha retrieves data for Bilans uspeha (income statement)
func (s *BilansiResource) GetBilansUspeha(ctx context.Context, tbl *domain.TableData) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.bilsRepo.GetHasGodHasKar()

	searchText := ""
	skraceni := false

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
		qb.AddEqual("bilu.skraceni", skraceni)
		qbTotals.AddEqual("bilu.skraceni", skraceni)
	}

	// Add search filter if provided
	if searchText != "" {
		qb.AddLike("bilu.nazp", interface{}(searchText))
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
				fmt.Sprintf("%d", entity.AOP),
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

func (s *BilansiResource) DeleteBilansStanja(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id parameter is required")
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id parameter: %v", err)
	}
	return s.bilsRepo.Delete(ctx, common.IDbils, idInt)
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

func (s *BilansiResource) DeleteBilansUspeha(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id parameter is required")
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id parameter: %v", err)
	}
	return s.biluRepo.Delete(ctx, common.IDbilu, idInt)
}

func (s *BilansiResource) GetFieldCacheBilu() map[string]reflect.StructField {
	if s.biluService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.biluService.GetFieldCache()
}

func (s *BilansiResource) StampaBilansStanja(ctx context.Context, tbl *domain.TableData, getTotalRecords, getOnlyTotals bool, pageSize, currentPage int, totals *domain.BilansiTotals) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	god := session.SelectedGod
	kar := session.SelectedKar
	lPGODizPS := true // TODO: Get this from business logic

	// STEP 1: Reset TGOD, TGODH for all BILS records with given GOD and KAR
	resetQb := common.NewQueryBuilder(`SELECT bilsid, rbr, grac, nazp, aop, konta, tgod, pgod, tgodh, pgodh, pgodps, pgodhps, nipo, skraceni FROM bils`, true)
	resetQb.AddEqual("god", god)
	resetQb.AddEqual("kar", kar)
	resetSql, resetArgs := resetQb.Build()

	bilsRecords, err := s.bilsRepo.GetAllCustom(ctx, resetSql, "", resetArgs, "", "")
	if err != nil {
		return err
	}

	for _, bils := range *bilsRecords {
		bils.TGod = 0
		bils.TGodH = 0
		if bils.Konta == "" {
			bils.PGod = 0
			bils.PGodH = 0
			bils.PGodPS = 0
			bils.PGodHPS = 0
		} else {
			if lPGODizPS {
				bils.PGod = 0
				bils.PGodH = 0
			}
		}
		// Update the record
		s.bilsRepo.Update(ctx, &bils, "bilsid", bils.BilsID, []domain.Fields{})
	}

	// STEP 2: Get max NIPO value
	maxNipoQb := common.NewQueryBuilder(`SELECT COALESCE(MAX(nipo), 1) as max_nipo FROM bils`, true)
	maxNipoQb.AddEqual("god", god)
	maxNipoQb.AddEqual("kar", kar)
	maxNipoSql, maxNipoArgs := maxNipoQb.Build()

	maxNipoRecords, err := s.bilsRepo.GetAllCustom(ctx, maxNipoSql, "", maxNipoArgs, "", "")
	if err != nil {
		return err
	}

	maxk := int64(1)
	if maxNipoRecords != nil && len(*maxNipoRecords) > 0 {
		// Assuming the result has a field for max_nipo
		// You may need to adjust based on actual struct field names
		maxk = int64(1) // Placeholder - adjust based on actual query result
	}

	// STEP 3: Process each NIPO level
	bilsMapByAop := make(map[int]*domain.Bils)

	for k := int64(1); k <= maxk; k++ {
		levelQb := common.NewQueryBuilder(`SELECT bilsid, rbr, grac, nazp, aop, konta, tgod, pgod, tgodh, pgodh, pgodps, pgodhps, nipo, skraceni, pozic_1, pozic_2, pozic_3, pozic_4, pozic_5, pozic_6, pozic_7, pozic_8, pozic_9, pozic_10, pozic_11, pozic_12 FROM bils`, true)
		levelQb.AddEqual("god", god)
		levelQb.AddEqual("kar", kar)
		levelQb.AddEqual("nipo", k)
		levelSql, levelArgs := levelQb.Build()

		levelRecords, err := s.bilsRepo.GetAllCustom(ctx, levelSql, "", levelArgs, "", "")
		if err != nil {
			return err
		}

		if levelRecords == nil {
			continue
		}

		for _, bils := range *levelRecords {
			if bils.Konta != "" {
				// Call ObradaKONTA equivalent - you may need to implement this
				dVred, dVredPS := s.obradaKonta(ctx, bils.AOP, bils.Konta, god, kar)
				bils.TGod = dVred
				bils.TGodH = int64(dVred / 1000)
				if lPGODizPS {
					bils.PGod = dVredPS
					bils.PGodH = int64(dVredPS / 1000)
				}
				s.bilsRepo.Update(ctx, &bils, "bilsid", bils.BilsID, []domain.Fields{})
			} else {
				// KONTA is empty - aggregate values from related BILS records
				for i := 1; i <= 12; i++ {
					// Get position value from struct field (need to handle dynamically)
					var nAOP int64
					switch i {
					case 1:
						nAOP = int64(bils.Pozic1)
					case 2:
						nAOP = int64(bils.Pozic2)
					case 3:
						nAOP = int64(bils.Pozic3)
					case 4:
						nAOP = int64(bils.Pozic4)
					case 5:
						nAOP = int64(bils.Pozic5)
					case 6:
						nAOP = int64(bils.Pozic6)
					case 7:
						nAOP = int64(bils.Pozic7)
					case 8:
						nAOP = int64(bils.Pozic8)
					case 9:
						nAOP = int64(bils.Pozic9)
					case 10:
						nAOP = int64(bils.Pozic10)
					case 11:
						nAOP = int64(bils.Pozic11)
					case 12:
						nAOP = int64(bils.Pozic12)
					}

					if nAOP == 0 {
						continue
					}

					// Get absolute value
					if nAOP < 0 {
						nAOP = -nAOP
					}

					// Find BILS record by AOP
					aopQb := common.NewQueryBuilder(`SELECT bilsid, rbr, grac, nazp, aop, konta, tgod, pgod, tgodh, pgodh, pgodps, pgodhps, nipo, skraceni FROM bils`, true)
					aopQb.AddEqual("god", god)
					aopQb.AddEqual("kar", kar)
					aopQb.AddEqual("aop", nAOP)
					aopSql, aopArgs := aopQb.Build()

					aopRecords, err := s.bilsRepo.GetAllCustom(ctx, aopSql, "", aopArgs, "", "")
					if err != nil {
						continue
					}

					if aopRecords != nil && len(*aopRecords) > 0 {
						relatedBils := (*aopRecords)[0]

						// Get original pozic value to determine add or subtract
						var origPozic int64
						switch i {
						case 1:
							origPozic = int64(bils.Pozic1)
						case 2:
							origPozic = int64(bils.Pozic2)
						case 3:
							origPozic = int64(bils.Pozic3)
						case 4:
							origPozic = int64(bils.Pozic4)
						case 5:
							origPozic = int64(bils.Pozic5)
						case 6:
							origPozic = int64(bils.Pozic6)
						case 7:
							origPozic = int64(bils.Pozic7)
						case 8:
							origPozic = int64(bils.Pozic8)
						case 9:
							origPozic = int64(bils.Pozic9)
						case 10:
							origPozic = int64(bils.Pozic10)
						case 11:
							origPozic = int64(bils.Pozic11)
						case 12:
							origPozic = int64(bils.Pozic12)
						}

						if origPozic > 0 {
							bils.TGod += relatedBils.TGod
							bils.TGodH += relatedBils.TGodH
							bils.PGod += relatedBils.PGod
							bils.PGodH += relatedBils.PGodH
							bils.PGodPS += relatedBils.PGodPS
							bils.PGodHPS += relatedBils.PGodHPS
						} else if origPozic < 0 {
							bils.TGod -= relatedBils.TGod
							bils.TGodH -= relatedBils.TGodH
							bils.PGod -= relatedBils.PGod
							bils.PGodH -= relatedBils.PGodH
							bils.PGodPS -= relatedBils.PGodPS
							bils.PGodHPS -= relatedBils.PGodHPS
						}
					}
				}

				s.bilsRepo.Update(ctx, &bils, "bilsid", bils.BilsID, []domain.Fields{})
			}

			// Ensure non-negative values
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

			s.bilsRepo.Update(ctx, &bils, "bilsid", bils.BilsID, []domain.Fields{})
			bilsMapByAop[bils.AOP] = &bils
		}
	}

	// STEP 4: Fetch final results and populate table
	skraceni := false
	finalQb := common.NewQueryBuilder(`SELECT bilsid, rbr, grac, nazp, aop, napomena, tgodh, pgodh, pgodhps, tgod, pgod, pgodps, nipo, skraceni FROM bils`, true)
	finalQb.AddEqual("god", god)
	finalQb.AddEqual("kar", kar)
	if skraceni {
		finalQb.AddEqual("skraceni", 1)
	}
	finalQb.AddOrderBy("aop ASC")

	if !getTotalRecords && !getOnlyTotals {
		finalQb.SetLimit(pageSize)
		finalQb.SetOffset((currentPage - 1) * pageSize)
	}

	finalSql, finalArgs := finalQb.Build()
	finalRecords, err := s.bilsRepo.GetAllCustom(ctx, finalSql, "", finalArgs, "", "")
	if err != nil {
		return err
	}

	// Handle total records request
	if getTotalRecords && finalRecords != nil {
		common.SetTableTotalRecords(tbl, len(*finalRecords), pageSize)
		return nil
	}

	// Populate table rows
	if finalRecords != nil && len(*finalRecords) > 0 {
		for _, bils := range *finalRecords {
			fields := []string{
				fmt.Sprintf("%d", bils.Rbr),
				bils.Grac,
				bils.NazP,
				fmt.Sprintf("%d", bils.AOP),
				bils.Napomena,
				common.FormatNumberWithSystemLocale(float64(bils.TGodH), 0),
				common.FormatNumberWithSystemLocale(float64(bils.PGodH), 0),
				common.FormatNumberWithSystemLocale(float64(bils.PGodHPS), 0),
				fmt.Sprintf("%d", bils.BilsID),
				common.FormatNumberWithSystemLocale(float64(bils.TGod), 0),
				common.FormatNumberWithSystemLocale(float64(bils.PGod), 0),
				common.FormatNumberWithSystemLocale(float64(bils.PGodPS), 0),
				fmt.Sprintf("%d", bils.NiPo),
				fmt.Sprintf("%d", bils.Skraceni),
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprintf("%d", bils.BilsID), Fields: fields, HasUpdate: false, HasDelete: false})
		}
	}

	return nil
}

// obradaKonta processes account data and returns calculated values
// This is a placeholder - implement based on your business logic
func (s *BilansiResource) obradaKonta(ctx context.Context, aop int, konta string, god, kar int) (float64, float64) {
	// TODO: Implement the actual ObradaKONTA logic from WinDev
	// This should calculate dVred (current year value) and dVredPS (previous year value)
	return 0, 0
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
		{Name: "pstduguje", Label: "Poc. stanje duguje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
		{Name: "pstpotrazuje", Label: "Poc. stanje potražuje", Width: "15", Field: "", SkipInSearch: true, IncludeInTotals: true, TextAlign: "right"},
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
}
