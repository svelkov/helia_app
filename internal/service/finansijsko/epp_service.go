package finansijsko

import (
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"

	"github.com/gin-gonic/gin"
)

// EppService defines the interface for operations related to EPP (Evidencija Prethodnog Poreza).
type EppService interface {
	GetSekcijeIzvori(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetEvidencija(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetSefKpr(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error
	GetSekcijeIzvoriTableFields() []domain.Fields
	GetEvidencijaTableFields() []domain.Fields
	GetSefKprTableFields() []domain.Fields
	GetFieldCache() map[string]reflect.StructField
}

// EppResource implements the EppService interface.
type EppResource struct {
	eppSekcijeIzvoriService  *service.BaseService[domain.EppSekcija]
	eppEvidencijaService     *service.BaseService[domain.EppEvidencija]
	eppSefKprService         *service.BaseService[domain.EppSefKpr]
	eppSekcijeIzvoriRepo     *repository.BaseRepository[domain.EppSekcija]
	eppEvidencijaRepo        *repository.BaseRepository[domain.EppEvidencija]
	eppSefKprRepo            *repository.BaseRepository[domain.EppSefKpr]
	sekcijeIzvoriTableFields []domain.Fields
	evidencijaTableFields    []domain.Fields
	sefKprTableFields        []domain.Fields
}

func NewEppService(
	eppSekcijeIzvoriService *service.BaseService[domain.EppSekcija],
	eppEvidencijaService *service.BaseService[domain.EppEvidencija],
	eppSefKprService *service.BaseService[domain.EppSefKpr],
	eppSekcijeIzvoriRepo *repository.BaseRepository[domain.EppSekcija],
	eppEvidencijaRepo *repository.BaseRepository[domain.EppEvidencija],
	eppSefKprRepo *repository.BaseRepository[domain.EppSefKpr],
) *EppResource {
	rs := &EppResource{
		eppSekcijeIzvoriService: eppSekcijeIzvoriService,
		eppEvidencijaService:    eppEvidencijaService,
		eppSefKprService:        eppSefKprService,
		eppSekcijeIzvoriRepo:    eppSekcijeIzvoriRepo,
		eppEvidencijaRepo:       eppEvidencijaRepo,
		eppSefKprRepo:           eppSefKprRepo,
	}
	rs.setServiceFieldValues()
	return rs
}

// GetSekcijeIzvoriTableFields returns the table field definitions for Sekcije i izvori
func (s *EppResource) GetSekcijeIzvoriTableFields() []domain.Fields {
	return s.sekcijeIzvoriTableFields
}

// GetEvidencijaTableFields returns the table field definitions for Evidencija PP
func (s *EppResource) GetEvidencijaTableFields() []domain.Fields {
	return s.evidencijaTableFields
}

// GetSefKprTableFields returns the table field definitions for SEF-KPR
func (s *EppResource) GetSefKprTableFields() []domain.Fields {
	return s.sefKprTableFields
}

// GetSekcije retrieves data for EPP Sekcije i izvori
func (s *EppResource) GetSekcijeIzvori(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Sekcije i izvori", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.eppSekcijeIzvoriRepo.GetHasGodHasKar()

	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")

	// Build query for Sekcije i izvori
	qb := common.NewQueryBuilder(`
		SELECT 
			fepp.nivo, fepp.sekcija, fepp.izvor, fepp.naziv,
			fepp.akt1, fepp.akt2, fepp.akt3, fepp.akt4,
			fepp.kprpoc, fepp.krpdodatnap, fepp.kprpdv,
			fepp.idfepp
		FROM fepp`)

	if hasGod {
		qb.AddEqual("fepp.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fepp.kar", session.SelectedKar)
	}
	if odDatuma != "" {
		qb.AddCondition("fepp.datumkreiranja", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("fepp.datumkreiranja", doDatuma, "<=")
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.EppSekcija{}))
		qb.AddSearchConditions(s.GetSekcijeIzvoriTableFields(), searchText)
	}

	qb.AddOrderBy("fepp.nivo ASC, fepp.sekcija ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.eppSekcijeIzvoriRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
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
			akt1Str := "Ne"
			akt2Str := "Ne"
			akt3Str := "Ne"
			akt4Str := "Ne"

			if entity.Akt1 {
				akt1Str = "Da"
			}
			if entity.Akt2 {
				akt2Str = "Da"
			}
			if entity.Akt3 {
				akt3Str = "Da"
			}
			if entity.Akt4 {
				akt4Str = "Da"
			}

			fields := []string{
				fmt.Sprintf("%v", entity.Nivo),
				entity.Sekcija,
				entity.Izvor,
				entity.Naziv,
				akt1Str,
				akt2Str,
				akt3Str,
				akt4Str,
				entity.KprPoc,
				entity.KprDodatnaPDV,
				entity.KprPDV,
				fmt.Sprintf("%v", entity.IDFepp),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetSekcijeIzvoriTableFields()

	return nil
}

// GetEvidencija retrieves data for EPP Evidencija PP
func (s *EppResource) GetEvidencija(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "Evidencija PP", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.eppEvidencijaRepo.GetHasGodHasKar()

	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")

	// Build query for Evidencija PP
	qb := common.NewQueryBuilder(`
		SELECT 
			feppevidencija.polje, feppevidencija.opis, feppevidencija.osn1,
			feppevidencija.pdv1, feppevidencija.osn2, feppevidencija.pdv2,
			feppevidencija.oddat, feppevidencija.dodat, feppevidencija.nipo,
			feppevidencija.fsepid
		FROM feppevidencija`)

	if hasGod {
		qb.AddEqual("feppevidencija.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("feppevidencija.kar", session.SelectedKar)
	}
	if odDatuma != "" {
		qb.AddCondition("feppevidencija.datum", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("feppevidencija.datum", doDatuma, "<=")
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.EppEvidencija{}))
		qb.AddSearchConditions(s.GetEvidencijaTableFields(), searchText)
	}

	qb.AddOrderBy("feppevidencija.datum DESC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.eppEvidencijaRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
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
				entity.Polje,
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Osn1, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv1, 2),
				common.FormatNumberWithSystemLocale(entity.Osn2, 2),
				common.FormatNumberWithSystemLocale(entity.Pdv2, 2),
				entity.Oddat.Format(common.DateLayout),
				entity.Dodat.Format(common.DateLayout),
				entity.Nipo,
				fmt.Sprintf("%v", entity.FsepID),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetEvidencijaTableFields()

	return nil
}

// GetSefKpr retrieves data for EPP SEF-KPR
func (s *EppResource) GetSefKpr(c *gin.Context, tbl *domain.TableData, getTotalRecords bool, pageSize, currentPage int) error {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetTableConfig(tbl, "SEF-KPR", "", true, true, false)
	tbl.SearchEnabled = true
	common.SetupTablePagination(tbl, currentPage, pageSize)

	hasGod, hasKar := s.eppSefKprRepo.GetHasGodHasKar()

	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	searchText := c.Query("query")

	// Build query for SEF-KPR
	qb := common.NewQueryBuilder(`
		SELECT 
			feppsef.redbroj, feppsef.dokumenttip, feppsef.brdokumenta,
			feppsef.datumdokumenta, feppsef.datumlicnog, feppsef.iznos,
			feppsef.pdv, feppsef.konto, feppsef.status,
			feppsef.idfeppsef
		FROM feppsef`)

	if hasGod {
		qb.AddEqual("feppsef.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("feppsef.kar", session.SelectedKar)
	}
	if odDatuma != "" {
		qb.AddCondition("feppsef.datumdokumenta", odDatuma, ">=")
	}
	if doDatuma != "" {
		qb.AddCondition("feppsef.datumdokumenta", doDatuma, "<=")
	}
	if searchText != "" {
		qb.SetEntityType(reflect.TypeOf(domain.EppSefKpr{}))
		qb.AddSearchConditions(s.GetSefKprTableFields(), searchText)
	}

	qb.AddOrderBy("feppsef.redbroj ASC")

	if !getTotalRecords {
		qb.SetLimit(pageSize)
		qb.SetOffset((currentPage - 1) * pageSize)
	}

	// Execute query and populate table
	sqlQuery, args := qb.Build()
	entities, err := s.eppSefKprRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
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
				fmt.Sprintf("%v", entity.RedBroj),
				entity.DokumentTip,
				entity.BrDokumenta,
				entity.DatumDokumenta.Format(common.DateLayout),
				entity.DatumLicnog.Format(common.DateLayout),
				common.FormatNumberWithSystemLocale(entity.Iznos, 2),
				common.FormatNumberWithSystemLocale(entity.PDV, 2),
				entity.Konto,
				entity.Status,
				fmt.Sprintf("%v", entity.IDFeppSef),
			}
			tblRow := domain.TableRow{Fields: fields, HasUpdate: true, HasDelete: true}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	// Set table headers
	tbl.Headers = s.GetSefKprTableFields()

	return nil
}

// GetFieldCache returns the cached field structure
func (s *EppResource) GetFieldCache() map[string]reflect.StructField {
	if s.eppSekcijeIzvoriService == nil {
		return make(map[string]reflect.StructField)
	}
	return s.eppSekcijeIzvoriService.GetFieldCache()
}

// setServiceFieldValues initializes table field definitions for EPP
func (s *EppResource) setServiceFieldValues() {
	// Fields for Sekcije i izvori
	s.sekcijeIzvoriTableFields = []domain.Fields{
		{Name: "nivo", Label: "Nivo", Width: "8", Field: "fepp.nivo", SkipInSearch: false},
		{Name: "sekcija", Label: "Sekcija", Width: "15", Field: "fepp.sekcija", SkipInSearch: false},
		{Name: "izvor", Label: "Izvor", Width: "15", Field: "fepp.izvor", SkipInSearch: false},
		{Name: "naziv", Label: "Naziv polja", Width: "25", Field: "fepp.naziv", SkipInSearch: false},
		{Name: "akt1", Label: "Aktivni 1", Width: "10", Field: "fepp.akt1", SkipInSearch: true},
		{Name: "akt2", Label: "Aktivni 2", Width: "10", Field: "fepp.akt2", SkipInSearch: true},
		{Name: "akt3", Label: "Aktivni 3", Width: "10", Field: "fepp.akt3", SkipInSearch: true},
		{Name: "akt4", Label: "Aktivni 4", Width: "10", Field: "fepp.akt4", SkipInSearch: true},
		{Name: "kprpoc", Label: "KPR početna", Width: "12", Field: "fepp.kprpoc", SkipInSearch: true},
		{Name: "krpdodatnap", Label: "KPR dodatna PDV", Width: "12", Field: "fepp.krpdodatnap", SkipInSearch: true},
		{Name: "kprpdv", Label: "KPR PDV račun", Width: "12", Field: "fepp.kprpdv", SkipInSearch: true},
		{Name: "idfepp", Label: "ID", Width: "8", Field: "fepp.idfepp", SkipInSearch: true},
	}

	// Fields for Evidencija PP
	s.evidencijaTableFields = []domain.Fields{
		{Name: "polje", Label: "Polje", Width: "15", Field: "feppevidencija.polje", SkipInSearch: false},
		{Name: "opis", Label: "Opis", Width: "25", Field: "feppevidencija.opis", SkipInSearch: false},
		{Name: "osn1", Label: "Osnova 1", Width: "12", Field: "feppevidencija.osn1", SkipInSearch: false},
		{Name: "pdv1", Label: "PDV 1", Width: "12", Field: "feppevidencija.pdv1", SkipInSearch: false},
		{Name: "osn2", Label: "Osnova 2", Width: "12", Field: "feppevidencija.osn2", SkipInSearch: false},
		{Name: "pdv2", Label: "PDV 2", Width: "12", Field: "feppevidencija.pdv2", SkipInSearch: false},
		{Name: "oddat", Label: "Od datuma", Width: "10", Field: "feppevidencija.oddat", SkipInSearch: false},
		{Name: "dodat", Label: "Do datuma", Width: "10", Field: "feppevidencija.dodat", SkipInSearch: false},
		{Name: "nipo", Label: "Nipo", Width: "10", Field: "feppevidencija.nipo", SkipInSearch: false},
		{Name: "fsepid", Label: "ID", Width: "8", Field: "feppevidencija.fsepid", SkipInSearch: true},
	}

	// Fields for SEF-KPR
	s.sefKprTableFields = []domain.Fields{
		{Name: "redbroj", Label: "Red. broj", Width: "8", Field: "feppsef.redbroj", SkipInSearch: false},
		{Name: "dokumenttip", Label: "Tip dokumenta", Width: "12", Field: "feppsef.dokumenttip", SkipInSearch: false},
		{Name: "brdokumenta", Label: "Broj dokumenta", Width: "12", Field: "feppsef.brdokumenta", SkipInSearch: false},
		{Name: "datumdokumenta", Label: "Datum dokumenta", Width: "10", Field: "feppsef.datumdokumenta", SkipInSearch: false},
		{Name: "datumlicnog", Label: "Datum ličnog", Width: "10", Field: "feppsef.datumlicnog", SkipInSearch: false},
		{Name: "iznos", Label: "Iznos", Width: "12", Field: "feppsef.iznos", SkipInSearch: false},
		{Name: "pdv", Label: "PDV", Width: "12", Field: "feppsef.pdv", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "10", Field: "feppsef.konto", SkipInSearch: false},
		{Name: "status", Label: "Status", Width: "12", Field: "feppsef.status", SkipInSearch: true},
		{Name: "idfeppsef", Label: "ID", Width: "8", Field: "feppsef.idfeppsef", SkipInSearch: true},
	}
}
