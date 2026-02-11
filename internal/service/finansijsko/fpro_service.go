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

// FproViewData encapsulates all data needed for the Nalog display page.
type FproViewData struct {
	FproEntities []domain.Fpro
	TableData    domain.TableData
}

// NalogService defines the interface for operations related to Fpro (Nalogs).
type FproService interface {
	service.Service[domain.Fpro]
	//GetNalogsViewData fetches all data required to render the Nalog list page.
	SetNalogIDFieldName(string)
	GetTableStavkeFields() []domain.Fields
	GetTableNalogFields() []domain.Fields
	GetNaloziStavke(c *gin.Context, nalogID int64, searchQuery string, page int, offset int, tableFields []domain.Fields) (domain.TableData, error)
	GetFieldCache() map[string]reflect.StructField
	GetAllByFnalID(c *gin.Context, fnalID int64) (*[]domain.Fpro, error)
}

// FproResource implements the NalogService interface.
type FproResource struct {
	service                 *service.BaseService[domain.Fpro]
	fproRepo                repository.BaseRepository[domain.Fpro]
	fproIDFieldName         string
	naloziTableFields       []domain.Fields
	naloziStavkeTableFields []domain.Fields
	cfg                     config.Config
}

func NewFproService(
	Service *service.BaseService[domain.Fpro],
	fproRepo repository.BaseRepository[domain.Fpro],
	fproIDFieldName string,
	naloziTableFields []domain.Fields,
	naloziStavkeTableFields []domain.Fields,
	cfg config.Config,
) *FproResource {
	rs := &FproResource{
		service:                 Service,
		fproRepo:                fproRepo,
		fproIDFieldName:         fproIDFieldName,
		naloziTableFields:       naloziTableFields,
		naloziStavkeTableFields: naloziStavkeTableFields,
		cfg:                     cfg,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *FproResource) SetNalogIDFieldName(fproIDFieldName string) {
	s.fproIDFieldName = fproIDFieldName
}
func (s *FproResource) SetNaloziTableFields(naloziTableFields []domain.Fields) {
	s.naloziTableFields = naloziTableFields
}

func (s *FproResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}

// Create implements NalogService.
func (s *FproResource) Create(c *gin.Context, Fpro *domain.Fpro, idField string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	return s.service.Create(c, Fpro, idField, fields)
}

// Delete implements NalogService.
func (s *FproResource) Delete(idField string, id int64) error {
	return s.service.Delete(idField, id)
}

// GetAll implements NalogService.
func (s *FproResource) GetAll(c *gin.Context, page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]domain.Fpro, error) {
	return s.service.GetAll(c, page, offset, tableFields, idField, searchParams...)
}

// GetAllCustom implements NalogService.
func (s *FproResource) GetAllCustom(c *gin.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fpro, error) {
	return s.service.GetAllCustom(c, queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements NalogService.
func (s *FproResource) GetByID(idField string, idValue int64) (*domain.Fpro, error) {
	return s.service.GetByID(idField, idValue)
}

// GetTotalRecords implements NalogService.
func (s *FproResource) GetTotalRecords(c *gin.Context, tableFields []domain.Fields, searchParams ...string) (int, error) {
	return s.service.GetTotalRecords(c, tableFields, searchParams...)
}

// GetTotalRecordsCustom implements NalogService.
func (s *FproResource) GetTotalRecordsCustom(c *gin.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(c, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements NalogService.
func (s *FproResource) MapEntityToValues(entity *domain.Fpro, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

// Update implements NalogService.
func (s *FproResource) Update(c *gin.Context, entity *domain.Fpro, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	return s.service.Update(c, entity, idField, idValue, tableFields)
}

func (s *FproResource) GetAllByFnalID(c *gin.Context, fnalID int64) (*[]domain.Fpro, error) {
	queryText := `select fp.*,  fk.naziv as naziv,
	case when fp.kat = 1 then fp.iznos
		 when fp.kat = 2 then fp.iznos
		 else 0 end as dug,
	case when fp.kat = 3 then fp.iznos
		 when fp.kat = 4 then fp.iznos
		 else 0 end as pot
				  from fpro as fp
			 	  left join fkpl fk on fk.idfkpl = fp.idfkpl`
	whereText := " where idfnal = $1"
	args := []interface{}{}
	args = append(args, fnalID)

	entities, err := s.fproRepo.GetAllCustom(c, queryText, whereText, args, "", " order by rbr desc ")
	return entities, err
}

// Helper to construct common WHERE clauses and arguments for Fpro queries
func (s *FproResource) buildFproWhere(idfnal int64, god, kar int, args *[]interface{}, searchQuery string) string {
	hasGod, hasKar := s.fproRepo.GetHasGodHasKar()
	basicWhere := s.fproRepo.CreateBasicWhere(s.naloziStavkeTableFields, args, hasGod, hasKar, god, kar, searchQuery)

	var conditions []string
	if basicWhere != "" {
		conditions = append(conditions, basicWhere)
	}
	if idfnal > 0 {
		conditions = append(conditions, fmt.Sprintf("(idfnal = %d)", idfnal))
	}

	if len(conditions) == 0 {
		return ""
	}
	return strings.Join(conditions, " AND ")
}

// GetNalogsViewData fetches all data required to render the Nalog list page.
func (s *FproResource) GetNaloziStavke(c *gin.Context, idFnal int64, searchQuery string, page int, offset int, tableFields []domain.Fields) (domain.TableData, error) {
	table := domain.TableData{}
	// 2. Fetch Fpro Entities
	args := []interface{}{}

	// Get session for god/kar values
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return table, fmt.Errorf("user session not found")
	}

	// Use viewData.DefaultTipdok for the actual Fpro query
	whereText := s.buildFproWhere(idFnal, session.SelectedGod, session.SelectedKar, &args, searchQuery)

	// Get total records
	totalRecordsQuery := `SELECT count(*) FROM Fpro `

	totRecords, err := s.fproRepo.GetTotalRecordsCustom(c, totalRecordsQuery, whereText, args, "", "")
	if err != nil {
		return table, fmt.Errorf("failed to get total records for Fpro: %w", err)
	}

	// Calculate pagination details
	currentPage, calculatedPageSize, totalPages := common.GetPaginationData(c, totRecords, s.cfg) // Pass nil for req
	if page > 0 {                                                                                 // Override current page if provided from handler
		currentPage = page
	}

	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", calculatedPageSize, (currentPage-1)*calculatedPageSize)
	orderBy := " ORDER BY rbr "
	selectQuery := `SELECT idFpro, rbr, fpro.konto, fpro.sifra, fkpl.naziv as nazivkonta, vrd, opis, dokum, dadok, rok, fpro.vkonta,
	 CASE 
        WHEN fpro.kat = 1 THEN fpro.iznos
        WHEN fpro.kat = 2 THEN  fpro.iznos
        ELSE 0
    END AS dug,
    CASE 
        WHEN fpro.kat = 3 THEN  fpro.iznos
        WHEN fpro.kat = 4 THEN  fpro.iznos
        ELSE 0
    END AS pot FROM Fpro
	LEFT join fkpl ON fkpl.idfkpl = fpro.idfkpl`

	entities, err := s.fproRepo.GetAllCustom(c, selectQuery, whereText, args, limitOffset, orderBy)
	if err != nil {
		return table, fmt.Errorf("failed to get Fpro entities: %w", err)
	}

	//viewData = *entities
	// Prepare TableData for UI
	table = common.SetTableBasicData("STAVKE NALOGA", "nalozi-tablestavke", s.naloziStavkeTableFields, "", "", calculatedPageSize, currentPage, totalPages, totRecords, s.cfg)

	// Prepare TableData for UI
	tbl, err := common.SetTableRows(&table, *entities, s.GetTableStavkeFields(), s.fproIDFieldName, "", s.service.GetFieldCache())
	if err != nil {
		return table, fmt.Errorf("failed to set table rows: %w", err)
	}
	table = *tbl
	// Additional table data configuration can happen here
	// if searchQuery != "" { // If it's a search, the URL might change for HTMX
	// 	table.URLGetAll = "/api/nalozi/all/search"
	// } else {
	// 	table.URLGetAll = "/api/nalozi/all/tipdok" // For tipdok specific updates
	// }
	table.Pagination.HxInclude = "#tipdokSelect, #search-input"
	table.HxInclude = "#tipdokSelect, #search-input"
	table.ShowActions = false // Default, can be overridden by specific handler needs
	table.BtnAdd.IsVisible = false
	table.BtnUpdate.IsVisible = false
	table.BtnDelete.IsVisible = false
	table.BtnPrint.IsVisible = false
	//table.Pagination.PageSizes = []int{5, 10, 20, 30, 50} // Example sizes
	table.Pagination.PageSize = calculatedPageSize
	table.Pagination.CurrentPage = currentPage
	table.Pagination.TotalPages = totalPages
	table.Pagination.TotalRecords = totRecords
	table.Pagination.StartRecord = (currentPage-1)*calculatedPageSize + 1
	table.Pagination.EndRecord = currentPage * calculatedPageSize
	// Ensure EndRecord is not greater than TotalRecords
	if table.Pagination.EndRecord > totRecords {
		table.Pagination.EndRecord = totRecords
	}
	// viewData.TableData = table

	// 3. Fetch UkupnaObrada totals if it's an initial load

	return table, nil
}

// GetTableStavkeFields  fetches all data required to render the Nalog Stavke list page.
func (s *FproResource) GetTableStavkeFields() []domain.Fields {
	if len(s.naloziStavkeTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziStavkeTableFields
}

// GetTableNalogFields fetches all data required to render the Nalog list page.
func (s *FproResource) GetTableNalogFields() []domain.Fields {
	if len(s.naloziTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziStavkeTableFields
}

func (s *FproResource) setServiceFieldValues() {
	s.naloziTableFields = []domain.Fields{
		{Name: "rbr", Label: "R. Broj", Width: "4"},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "6"},
		{Name: "nalog", Label: "Br. Naloga", Width: "12"},
		{Name: "danal", Label: "Datum naloga", Width: "12"},
		{Name: "opis", Label: "opis ", Width: "60"},
		{Name: "dug", Label: "Duguje", Width: "14"},
		{Name: "pot", Label: "Potrazuje", Width: "14"},
		{Name: "datob", Label: "Datum obrade", Width: "12"},
		{Name: "brst", Label: "Br.Stavki", Width: "5"},
		{Name: "nalsts", Label: "Status naloga", Width: "10"},
	}
	s.naloziStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "R. Broj", Width: "4"},
		{Name: "konto", Label: "Konto", Width: "6"},
		{Name: "sifra", Label: "Sifra", Width: "6"},
		{Name: "nazivkonta", Label: "Naziv Konta", Width: "60"},
		{Name: "vrd", Label: "Vrsta Dok.", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "60"},
		{Name: "dug", Label: "Duguje", Width: "14", SkipInSearch: true},
		{Name: "pot", Label: "Potrazuje", Width: "14", SkipInSearch: true},
		{Name: "dokum", Label: "Br. Dokum", Width: "12"},
		{Name: "dadok", Label: "Datum Dok.", Width: "12"},
	}
}
