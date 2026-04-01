package finansijsko

import (
	"context"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
)

// KamateService defines the interface for operations related to Kamate (Interest Rates and Calculations).
type KamateService interface {
	GetTableFields() []domain.Fields
	GetFormiranjeLisovaTableFields() []domain.Fields
	GetObracunTableFields() []domain.Fields
	GetKamatneStope(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error
	GetFormiranjeLista(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, konto, odSifre, doSifre, odDatuma, doDatuma, searchText string) error
	GetObracunKamate(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odBrojaListe, doBrojaListe, podDatumom, searchText string) error
	GetFieldCache() map[string]reflect.StructField
}

// KamateResource implements the KamateService interface.
type KamateResource struct {
	fkplService            *service.BaseService[domain.Fkpl]
	fkplRepo               *repository.BaseRepository[domain.Fkpl]
	kamateStopTableFields  []domain.Fields
	formiranjeLisovaFields []domain.Fields
	obracunTableFields     []domain.Fields
}

func NewKamateService(
	fkplService *service.BaseService[domain.Fkpl],
	fkplRepo *repository.BaseRepository[domain.Fkpl],
) *KamateResource {
	rs := &KamateResource{
		fkplService: fkplService,
		fkplRepo:    fkplRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// GetTableFields returns the table field definitions for Kamate (default Kamatne stope)
func (s *KamateResource) GetTableFields() []domain.Fields {
	return s.kamateStopTableFields
}

// GetFormiranjeLisovaTableFields returns the table field definitions for Formiranje listova
func (s *KamateResource) GetFormiranjeLisovaTableFields() []domain.Fields {
	return s.formiranjeLisovaFields
}

// GetObracunTableFields returns the table field definitions for Obracun kamate
func (s *KamateResource) GetObracunTableFields() []domain.Fields {
	return s.obracunTableFields
}

// GetKamatneStope retrieves data for Kamatne stope (Interest Rates)
func (s *KamateResource) GetKamatneStope(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Kamatne stope", "", false, false, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()

	// Build query for Kamatne stope
	qb := common.NewQueryBuilder(`
		SELECT 
			fkpl.idfkpl, fkpl.konto, fkpl.sifra, fkpl.naziv,
			fkpl.god, fkpl.kar,
			'1900-01-01'::date as vrsta_od_datuma,
			'2099-12-31'::date as vrsta_do_datuma,
			0 as kamatna_stopa,
			0 as redni_broj
		FROM fkpl `, true)

	if hasGod {
		qb.AddEqual("fkpl.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fkpl.kar", session.SelectedKar)
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Fkpl{}))
		qb.AddSearchConditions(s.GetTableFields(), searchText)
	}

	qb.AddOrderBy("fkpl.konto ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				entity.Konto,
				"31/12/2099",
				common.FormatNumberWithSystemLocale(0, 2),
				"1",
				"",
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetTableFields()

	return nil
}

// GetFormiranjeLista retrieves data for Formiranje kamatnih listova (Forming Interest Lists)
func (s *KamateResource) GetFormiranjeLista(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, konto, odSifre, doSifre, odDatuma, doDatuma, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Pregled dokumenata za formiranje kamatnih listova", "", false, false, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()

	// Build query for Formiranje listova - fetches document data
	qb := common.NewQueryBuilder(`
		SELECT 
			fkpl.idfkpl, fkpl.konto, fkpl.sifra, fkpl.naziv,
			fkpl.god, fkpl.kar,
			0 as broj_dokumenta,
			'1900-01-01'::date as datdok,
			0 as rok,
			'1900-01-01'::date as datum_rospesca,
			0 as zaduzenje,
			0 as uplata,
			0 as saldo
		FROM fkpl `, true)

	if hasGod {
		qb.AddEqual("fkpl.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fkpl.kar", session.SelectedKar)
	}
	if konto != "" {
		qb.AddCondition("fkpl.konto", konto, "=")
	}
	if odSifre != "" {
		qb.AddCondition("fkpl.sifra", odSifre, ">=")
	}
	if doSifre != "" {
		qb.AddCondition("fkpl.sifra", doSifre, "<=")
	}
	if odDatuma != "" {
		qb.AddCondition("datdok", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("datdok", doDatuma, "<=")
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Fkpl{}))
		qb.AddSearchConditions(s.GetFormiranjeLisovaTableFields(), searchText)
	}

	qb.AddOrderBy("fkpl.konto ASC, fkpl.sifra ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for i, entity := range *entities {
			fields := []string{
				fmt.Sprintf("%d", i+1),
				entity.Konto,
				"01/01/1900",
				"0",
				"01/01/1900",
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				"",
				"",
				"",
				"",
				"",
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetFormiranjeLisovaTableFields()

	return nil
}

// GetObracunKamate retrieves data for Obracun kamate (Interest Calculation)
func (s *KamateResource) GetObracunKamate(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int, odBrojaListe, doBrojaListe, podDatumom, searchText string) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Obracun kamate", "", false, false, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()

	// Build query for Obracun kamate
	qb := common.NewQueryBuilder(`
		SELECT 
			fkpl.idfkpl, fkpl.konto, fkpl.sifra, fkpl.naziv,
			fkpl.god, fkpl.kar,
			0 as br_kam_lista,
			0 as broj_dokumenta,
			'1900-01-01'::date as datum_dokuma,
			0 as rok,
			'1900-01-01'::date as datum_rospesca,
			'1900-01-01'::date as od_datuma,
			'1900-01-01'::date as do_datuma,
			0 as osnova,
			0 as duguje,
			0 as potrazuje,
			0 as kamatna_stopa,
			0 as model_kamate,
			0 as koeficijent,
			0 as iznos_dana,
			0 as broj,
			0 as iznos
		FROM fkpl `, true)

	if hasGod {
		qb.AddEqual("fkpl.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fkpl.kar", session.SelectedKar)
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.Fkpl{}))
		qb.AddSearchConditions(s.GetObracunTableFields(), searchText)
	}
	qb.AddCondition("br_kam_lista", odBrojaListe, ">=")
	qb.AddCondition("br_kam_lista", doBrojaListe, "<=")
	qb.AddCondition("datum_dokuma", podDatumom, "<=")

	qb.AddOrderBy("fkpl.konto ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	// Set total records and pagination
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		for _, entity := range *entities {
			fields := []string{
				"",
				"",
				"01/01/1900",
				"0",
				"01/01/1900",
				"01/01/1900",
				"01/01/1900",
				common.FormatNumberWithSystemLocale(entity.Kolicinski, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				"0",
				common.FormatNumberWithSystemLocale(0, 2),
				common.FormatNumberWithSystemLocale(0, 2),
				"0",
				common.FormatNumberWithSystemLocale(0, 2),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetObracunTableFields()

	return nil
}

// GetFieldCache returns the cached field structure
func (s *KamateResource) GetFieldCache() map[string]reflect.StructField {
	if s.fkplService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.fkplService.GetFieldCache()
}

// setServiceFieldValues initializes table field definitions for Kamate
func (s *KamateResource) setServiceFieldValues() {
	// Fields for Kamatne stope (Interest Rates)
	s.kamateStopTableFields = []domain.Fields{
		{Name: "vrsta_od_datuma", Label: "Vrst od datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "vrsta_do_datuma", Label: "Vrst do datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "kamatna_stopa", Label: "Kamatna stopa", Width: "12", Field: "", SkipInSearch: true},
		{Name: "redni_broj", Label: "Redni broj", Width: "10", Field: "", SkipInSearch: true},
		{Name: "opis", Label: "OPIS", Width: "25", Field: "", SkipInSearch: true},
	}

	// Fields for Formiranje listova (Forming Lists)
	s.formiranjeLisovaFields = []domain.Fields{
		{Name: "broj_dokumenta", Label: "Broj Dokumenta", Width: "12", Field: "", SkipInSearch: true},
		{Name: "datum_dokuma", Label: "Datum Dokum.", Width: "12", Field: "", SkipInSearch: true},
		{Name: "rok", Label: "Rok.", Width: "10", Field: "", SkipInSearch: true},
		{Name: "datum_rospesca", Label: "Datum rospesca", Width: "12", Field: "", SkipInSearch: true},
		{Name: "zaduzenje", Label: "Zaduzenje", Width: "12", Field: "", SkipInSearch: true},
		{Name: "uplata", Label: "Uplata", Width: "12", Field: "", SkipInSearch: true},
		{Name: "saldo", Label: "Saldo", Width: "12", Field: "", SkipInSearch: true},
		{Name: "broj_naloga", Label: "Broj Naloga", Width: "12", Field: "", SkipInSearch: true},
		{Name: "dnal", Label: "DNAL", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tzatv", Label: "TZATV", Width: "12", Field: "", SkipInSearch: true},
		{Name: "totv", Label: "TOTV", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tbrsi", Label: "TBRSI", Width: "12", Field: "", SkipInSearch: true},
		{Name: "tip_dokum", Label: "Tip Dokum.", Width: "12", Field: "", SkipInSearch: true},
	}

	// Fields for Obracun kamate (Interest Calculation)
	s.obracunTableFields = []domain.Fields{
		{Name: "br_kam_lista", Label: "Br. kam. lista", Width: "12", Field: "", SkipInSearch: true},
		{Name: "broj_dokumenta", Label: "Broj Dokumenta", Width: "12", Field: "", SkipInSearch: true},
		{Name: "datum_dokuma", Label: "Datum Dokum.", Width: "12", Field: "", SkipInSearch: true},
		{Name: "rok", Label: "Rok.", Width: "10", Field: "", SkipInSearch: true},
		{Name: "datum_rospesca", Label: "Datum rospesca", Width: "12", Field: "", SkipInSearch: true},
		{Name: "od_datuma", Label: "Od Datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "do_datuma", Label: "Do Datuma", Width: "12", Field: "", SkipInSearch: true},
		{Name: "osnova", Label: "Osnova", Width: "12", Field: "", SkipInSearch: true},
		{Name: "duguje", Label: "Duguje", Width: "12", Field: "", SkipInSearch: true},
		{Name: "potrazuje", Label: "Potrazuje", Width: "12", Field: "", SkipInSearch: true},
		{Name: "kamatna_stopa", Label: "Kamatna Stopa", Width: "12", Field: "", SkipInSearch: true},
		{Name: "model_kamate", Label: "Model Kamate", Width: "12", Field: "", SkipInSearch: true},
		{Name: "koeficijent", Label: "Koeficijent", Width: "12", Field: "", SkipInSearch: true},
		{Name: "iznos_dana", Label: "Iznos Dana", Width: "12", Field: "", SkipInSearch: true},
		{Name: "broj", Label: "Broj", Width: "12", Field: "", SkipInSearch: true},
		{Name: "iznos", Label: "Iznos", Width: "12", Field: "", SkipInSearch: true},
	}
}
