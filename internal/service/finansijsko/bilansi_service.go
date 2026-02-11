package finansijsko

import (
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// BilansiService defines the interface for operations related to Bilansi (Balance Sheets).
type BilansiService interface {
	GetTableFields() []domain.Fields
	GetZakljucniTableFields() []domain.Fields
	GetBilansStanjaTableFields() []domain.Fields
	GetBilansUspehaTableFields() []domain.Fields
	GetZakljucniList(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetBilansStanja(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetBilansUspeha(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetFieldCache() map[string]reflect.StructField
}

// BilansiResource implements the BilansiService interface.
type BilansiResource struct {
	biluService             *service.BaseService[domain.BiluPayload]
	biluRepo                *repository.BaseRepository[domain.BiluPayload]
	bilsService             *service.BaseService[domain.BilsPayload]
	bilsRepo                *repository.BaseRepository[domain.BilsPayload]
	fproRepo                *repository.BaseRepository[domain.FproDto]
	zakljucniTableFields    []domain.Fields
	bilansStanjaTableFields []domain.Fields
	bilansUspehaTableFields []domain.Fields
	cfg                     config.Config
}

func NewBilansiService(
	biluService *service.BaseService[domain.BiluPayload],
	biluRepo *repository.BaseRepository[domain.BiluPayload],
	bilsService *service.BaseService[domain.BilsPayload],
	bilsRepo *repository.BaseRepository[domain.BilsPayload],
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

// ZakljucniListItem represents a closing list item
type ZakljucniListItem struct {
	RB        int
	Konto     string
	Sifra     string
	Naziv     string
	Klasa     string
	Grupa     string
	Sint      string
	PstDug    float64 // Beginning balance debit
	PstPot    float64 // Beginning balance credit
	PrometDug float64 // Debit movement
	PrometPot float64 // Credit movement
	SaldoDug  float64 // Debit balance
	SaldoPot  float64 // Credit balance
}

// GetZakljucniList retrieves data for Zakljucni list (closing account balance)
func (s *BilansiResource) GetZakljucniList(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "", "", false, false, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()

	// Get parameters
	odKonta := c.Query("odkonta")
	doKonta := c.Query("dokonta")
	odSifre := c.Query("odsifre")
	doSifre := c.Query("dosifre")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	tipLista := c.Query("tip_zakljucni")
	//analitickakonta := c.Query("analitickakonta")
	klasa9 := c.Query("klasa9")
	samosaprometom := c.Query("samosaprometom")
	//zabanku := c.Query("zabanku")

	// Build map to store closing list items
	zakljucniMap := make(map[string]*ZakljucniListItem)

	// Build optimized query with only needed columns
	qb := common.NewQueryBuilder(`SELECT fpro.konto, fpro.sifra, coalesce(fkpl.naziv, '') as naziv,
								  fpro.tipdok, fpro.iznos, fpro.kat FROM fpro`)
	// Conditional joins only
	if tipLista == "1" {
		qb.AddJoin("left join fkpl on fkpl.idfkpl = fpro.idfkpl")
	}
	if tipLista == "2" {
		qb.AddJoin("left join fkpl on fkpl.god = fpro.god and fkpl.kar = fpro.kar and fkpl.konto = fpro.konto and fkpl.vkonta = 2")
	}
	if tipLista == "3" {
		qb.AddJoin("left join fkpl on fkpl.god = fpro.god and fkpl.kar = fpro.kar and fkpl.konto = left(fpro.konto, length(fpro.konto) -1)  and fkpl.vkonta = 3")
	}
	// Add god and kar filters
	if hasGod {
		qb.AddEqual("fpro.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fpro.kar", session.SelectedKar)
	}

	// Add filters
	if odKonta != "" {
		qb.AddCondition("fpro.konto::numeric", odKonta, ">=")
	}
	if doKonta != "" {
		qb.AddCondition("fpro.konto::numeric", doKonta, "<=")
	}
	qb.AddCondition("fpro.danal", odDatuma, ">=")
	qb.AddCondition("fpro.danal", doDatuma, "<=")
	if tipLista == "1" {
		// Analytic: by account and code
		if odSifre != "" {
			qb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", odSifre, ">=")
		}
		if doSifre != "" {
			qb.AddCondition("COALESCE(NULLIF(fpro.sifra, '')::numeric, 0)", doSifre, "<=")
		}
	}

	qb.AddOrderBy("fpro.konto::numeric ASC, COALESCE(NULLIF(fpro.sifra, '')::numeric, 0) ASC")
	// Execute query
	sqlQuery, args := qb.Build()
	//fmt.Println("SQL Query for Zakljucni list:", sqlQuery, args)
	entities, err := s.fproRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	// Read records and build map
	redBr := 1
	for _, entity := range *entities {
		// Determine key based on tip lista
		var sKey string
		switch tipLista {
		case "1":
			sKey = fmt.Sprintf("%s-%s", entity.Konto, entity.Sifra)
		case "2", "3":
			sKey = entity.Konto
		}

		// Get or create item in map
		item, exists := zakljucniMap[sKey]
		if !exists {
			klasa := ""
			grupa := ""
			sint := ""
			if len(entity.Konto) > 0 {
				klasa = entity.Konto[0:1]
			}
			if len(entity.Konto) >= s.cfg.NDuzSint {
				grupa = entity.Konto[0 : s.cfg.NDuzSint-1]
				sint = entity.Konto[0:s.cfg.NDuzSint]
			}
			item = &ZakljucniListItem{
				RB:    redBr,
				Konto: entity.Konto,
				Sifra: entity.Sifra,
				Naziv: entity.Naziv,
				Klasa: klasa,
				Grupa: grupa,
				Sint:  sint,
			}
			if tipLista == "2" || tipLista == "3" {
				item.Sifra = ""
			}
			if tipLista == "3" {
				item.Konto = item.Konto[0 : len(item.Konto)-1]
			}
			zakljucniMap[sKey] = item
			redBr++
		}

		// Accumulate values
		if entity.Tipdok == "00" {
			// Beginning balance
			if entity.Kat == 1 || entity.Kat == 2 {
				item.PstDug += entity.Iznos
				item.SaldoDug += entity.Iznos
			} else {
				item.PstPot += entity.Iznos
				item.SaldoPot += entity.Iznos
			}
		} else {
			// Movement
			if entity.Kat == 1 || entity.Kat == 2 {
				item.PrometDug += entity.Iznos
				item.SaldoDug += entity.Iznos
			} else {
				item.PrometPot += entity.Iznos
				item.SaldoPot += entity.Iznos
			}
		}
	}

	// Convert map to slice and apply filters
	items := make([]*ZakljucniListItem, 0, len(zakljucniMap))
	for _, item := range zakljucniMap {
		// Filter: skip items with no movement if checkbox is set
		if samosaprometom == "true" && item.SaldoDug == 0 && item.SaldoPot == 0 {
			continue
		}

		// Filter: skip class 9 if checkbox is not set
		if klasa9 != "true" && item.Klasa == "9" {
			continue
		}

		// Calculate net balance
		if item.SaldoDug > item.SaldoPot {
			item.SaldoDug = item.SaldoDug - item.SaldoPot
			item.SaldoPot = 0
		} else if item.SaldoDug < item.SaldoPot {
			item.SaldoPot = item.SaldoPot - item.SaldoDug
			item.SaldoDug = 0
		} else {
			item.SaldoDug = 0
			item.SaldoPot = 0
		}

		items = append(items, item)
	}

	// Handle pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(items), pageSize)
		return nil
	}

	// Apply pagination
	start := (currentPage - 1) * pageSize
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	var paginatedItems []*ZakljucniListItem
	if start < len(items) {
		paginatedItems = items[start:end]
	}

	// Populate table rows
	for i, item := range paginatedItems {
		fields := []string{
			fmt.Sprintf("%d", start+i+1),
			item.Konto,
			item.Sifra,
			item.Naziv,
			common.FormatNumberWithSystemLocale(item.PstDug, 2),
			common.FormatNumberWithSystemLocale(item.PstPot, 2),
			common.FormatNumberWithSystemLocale(item.PrometDug, 2),
			common.FormatNumberWithSystemLocale(item.PrometPot, 2),
			common.FormatNumberWithSystemLocale(item.SaldoDug, 2),
			common.FormatNumberWithSystemLocale(item.SaldoPot, 2),
		}
		tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	// Set table headers
	tbl.Headers = s.GetZakljucniTableFields()

	return nil
}

// GetBilansStanja retrieves data for Bilans stanja (balance sheet)
func (s *BilansiResource) GetBilansStanja(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Bilans stanja", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	searchText := c.Query("query")
	skraceni := c.Query("skraceni") == "true"
	searchLower := strings.ToLower(searchText)

	// Build optimized query for BILS data
	qb := common.NewQueryBuilder(`SELECT bils.idbils, bils.vkonta, bils.naziv, bils.aop, 
		bils.konta, bils.tgod, bils.pgod, bils.nipo, bils.tgodh, bils.pgodh, bils.pozicije, bils.skraceni FROM bils`)

	// Add filters directly in query
	qb.AddEqual("bils.god", session.SelectedGod)
	qb.AddEqual("bils.kar", session.SelectedKar)

	// Filter by skraceni flag if set
	if skraceni {
		qb.AddEqual("bils.skraceni", 1)
	}

	// Add search filter if provided
	if searchText != "" {
		qb.AddCondition("LOWER(bils.naziv)", searchLower, "LIKE")
	}

	qb.AddOrderBy("bils.nipo ASC, bils.vkonta ASC")

	// Only apply pagination if not counting total records
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and get entities
	sqlQuery, args := qb.Build()
	entities, err := s.bilsRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Handle total records request
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		start := (currentPage - 1) * pageSize
		for i, entity := range *entities {
			// Apply filtering logic on processed data
			if skraceni && entity.Skraceni != 1 {
				continue
			}

			// Ensure non-negative values
			// tgod := int64(0)
			// if entity.TGOD > 0 {
			// 	tgod = entity.TGOD
			// }
			// pgod := int64(0)
			// if entity.PGod > 0 {
			// 	pgod = entity.PGod
			// }

			fields := []string{
				fmt.Sprintf("%d", start+i+1),
				fmt.Sprintf("%d", entity.Vkonta),
				entity.Naziv,
				fmt.Sprintf("%d", entity.AOP),
				entity.Konta,
				// common.FormatNumberWithSystemLocale(float64(tgod), 2),
				// common.FormatNumberWithSystemLocale(float64(pgod), 2),
				// fmt.Sprintf("%d", entity.NiPo),
				// common.FormatNumberWithSystemLocale(float64(tgod/1000), 2),
				// common.FormatNumberWithSystemLocale(float64(pgod/1000), 2),
				//entity.Pozic1,
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: fields, HasUpdate: true, HasDelete: true})
		}
	}

	// Set table headers
	tbl.Headers = s.GetBilansStanjaTableFields()

	return nil
}

// GetBilansUspeha retrieves data for Bilans uspeha (income statement)
func (s *BilansiResource) GetBilansUspeha(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Bilans uspeha", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	searchText := c.Query("query")
	zeroVal := common.FormatNumberWithSystemLocale(0, 2)

	// Build optimized query for Bilans uspeha - fetches income/expense account data
	// Only select necessary columns
	qb := common.NewQueryBuilder(`SELECT bilu.idbilu, bilu.konto, bilu.sifra, bilu.naziv, bilu.vkonta FROM bilu`)

	// Add filters
	qb.AddEqual("bilu.god", session.SelectedGod)
	qb.AddEqual("bilu.kar", session.SelectedKar)
	// Filter for income/expense accounts (typically classes 4, 5, 6, 7)
	qb.AddCondition("bilu.vkonta", "4", ">=")

	// Add search filter if provided
	if searchText != "" {
		searchLower := strings.ToLower(searchText)
		qb.AddCondition("LOWER(bilu.naziv)", searchLower, "LIKE")
	}

	qb.AddOrderBy("bilu.vkonta ASC, bilu.konto ASC")

	// Only apply pagination if not counting total records
	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and get entities
	sqlQuery, args := qb.Build()
	entities, err := s.biluRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Handle total records request
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// Populate table rows efficiently
	if entities != nil && len(*entities) > 0 {
		start := (currentPage - 1) * pageSize
		for i, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%d", start+i+1),
				fmt.Sprintf("%d", entity.Vkonta),
				entity.Naziv,
				"",
				"",
				zeroVal,
				zeroVal,
				zeroVal,
			}
			tbl.Rows = append(tbl.Rows, domain.TableRow{Fields: fields, HasUpdate: true, HasDelete: true})
		}
	}

	// Set table headers
	tbl.Headers = s.GetBilansUspehaTableFields()

	return nil
}

// GetFieldCache returns the cached field structure
func (s *BilansiResource) GetFieldCache() map[string]reflect.StructField {
	if s.bilsService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.bilsService.GetFieldCache()
}

// BilansStanjaItem represents a processed balance sheet item
type BilansStanjaItem struct {
	VKonta           int64
	Naziv            string
	AOP              string
	KontaList        string
	TekucaGodina     int64
	PrethodnаGodina  int64
	NivoPodataka     int64
	TekucaHiljada    int64
	PrethodnаHiljada int64
	Pozicije         string
}

// processBilsData processes BILS records and calculates totals from fpro transactions
// This implements the logic from the WinDev ObradaBILS procedure
func (s *BilansiResource) processBilsData(god, kar int, skraceni bool) ([]*BilansStanjaItem, error) {
	// Map to store processed BILS records by ID
	bilsMap := make(map[int64]*domain.BilsPayload)

	// Step 1: Reset all totals to 0
	if err := s.resetBilsTotalsForYear(god, kar, bilsMap); err != nil {
		return nil, fmt.Errorf("error resetting totals: %w", err)
	}

	// Step 2: Get all BILS records
	// allBils, err := s.bilsRepo.GetAll(nil)
	// if err != nil {
	// 	return nil, fmt.Errorf("error fetching BILS records: %w", err)
	// }

	// Populate map and filter by year/month
	// for _, bils := range *allBils {
	// 	if bils.God == int64(god) && bils.Kar == int64(kar) {
	// 		bilsMap[bils.IDBils] = &bils
	// 	}
	// }

	// Step 4: Convert to BilansStanjaItem and apply filtering
	result := make([]*BilansStanjaItem, 0)
	for _, bils := range bilsMap {
		// Apply skraceni filter if needed
		if skraceni && bils.Skraceni != 1 {
			continue
		}

		// Ensure no negative values
		/* 	tgod := int64(0)
		if bils.TGOD > 0 {
			tgod = bils.TGOD
		}
		pgod := int64(0)
		if bils.PGod > 0 {
			pgod = bils.PGod
		}

		item := &BilansStanjaItem{
			VKonta:           bils.VKonta,
			Naziv:            bils.Naziv,
			AOP:              fmt.Sprintf("%d", bils.AOP),
			KontaList:        bils.Konta,
			TekucaGodina:     tgod,
			PrethodnаGodina:  pgod,
			NivoPodataka:     bils.NIPO,
			TekucaHiljada:    tgod / 1000,
			PrethodnаHiljada: pgod / 1000,
			Pozicije:         "",
		} */
		//result = append(result, item)
	}

	return result, nil
}

// resetBilsTotalsForYear resets TGOD, PGOD values for all BILS records in the period
func (s *BilansiResource) resetBilsTotalsForYear(god, kar int, bilsMap map[int64]*domain.BilsPayload) error {
	// This would query the database and reset values
	// For now, we'll handle this in the processBilsLevel function
	return nil
}

// setServiceFieldValues initializes table field definitions for Bilansi
func (s *BilansiResource) setServiceFieldValues() {
	// Fields for Zakljucni list
	s.zakljucniTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "8", Field: "", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "12", Field: "bils.konto", SkipInSearch: false},
		{Name: "sifra", Label: "Šifra", Width: "10", Field: "bils.sifra", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv", Width: "25", Field: "bils.naziv", SkipInSearch: false},
		{Name: "rocstanje_duguje", Label: "Roc. stanje duguje", Width: "15", Field: "", SkipInSearch: true},
		{Name: "rocstanje_potrazuje", Label: "Roc. stanje potražuje", Width: "15", Field: "", SkipInSearch: true},
		{Name: "promet_duguje", Label: "Promet duguje", Width: "15", Field: "", SkipInSearch: true},
		{Name: "promet_potrazuje", Label: "Promet potražuje", Width: "15", Field: "", SkipInSearch: true},
		{Name: "saldo_duguje", Label: "Saldo duguje", Width: "15", Field: "", SkipInSearch: true},
		{Name: "saldo_potrazuje", Label: "Saldo potražuje", Width: "15", Field: "", SkipInSearch: true},
	}

	// Fields for Bilans stanja
	s.bilansStanjaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "8", Field: "", SkipInSearch: true},
		{Name: "grupa_racuna", Label: "Grupa računa", Width: "15", Field: "bils.vkonta", SkipInSearch: false},
		{Name: "naziv_pozicije", Label: "Naziv pozicije", Width: "25", Field: "bils.naziv", SkipInSearch: false},
		{Name: "oznaka_aop", Label: "Oznaka za AOP", Width: "12", Field: "", SkipInSearch: true},
		{Name: "spisak_konta", Label: "Spisak konta", Width: "15", Field: "", SkipInSearch: true},
		{Name: "tekuca_godina", Label: "Tekuća godina", Width: "15", Field: "", SkipInSearch: true},
		{Name: "prethodna_godina", Label: "Prethodna godina", Width: "15", Field: "", SkipInSearch: true},
		{Name: "nivo_podataka", Label: "Nivo podataka", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tekuca_hiljada", Label: "Tekuća u hiljadama", Width: "15", Field: "", SkipInSearch: true},
		{Name: "prethodna_hiljada", Label: "Prethodna u hiljadama", Width: "15", Field: "", SkipInSearch: true},
		{Name: "pozicije", Label: "Pozicije", Width: "12", Field: "", SkipInSearch: true},
	}

	// Fields for Bilans uspeha
	s.bilansUspehaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "8", Field: "", SkipInSearch: true},
		{Name: "grupa_racuna", Label: "Grupa računa", Width: "15", Field: "bils.vkonta", SkipInSearch: false},
		{Name: "naziv_pozicije", Label: "Naziv pozicije", Width: "25", Field: "bils.naziv", SkipInSearch: false},
		{Name: "aop", Label: "AOP", Width: "12", Field: "", SkipInSearch: true},
		{Name: "spisak_konta", Label: "Spisak konta", Width: "15", Field: "", SkipInSearch: true},
		{Name: "tekuca_godina", Label: "Tekuća godina", Width: "15", Field: "", SkipInSearch: true},
		{Name: "prethodna_godina", Label: "Prethodna godina", Width: "15", Field: "", SkipInSearch: true},
		{Name: "nivo_podataka", Label: "Nivo podataka", Width: "12", Field: "", SkipInSearch: true},
	}
}
