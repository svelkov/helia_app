package service

import (
	"fmt"
	"helia/global"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/validation"

	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// NalogViewData encapsulates all data needed for the Nalog display page.
type NalogViewData struct {
	FnalEntities     []domain.Fnal
	TableData        domain.TableData
	TipdokOptions    *[]domain.Tipdok     // Pointer to allow nil if not needed
	UkupnaObrada     *domain.UkupnaObrada // Pointer to allow nil if not needed
	IsInitialLoad    bool                 // True if it's the first full page load, false for HTMX partials
	DefaultTipdok    string               // The default tipdok if none is selected
	TypeView         string               // Type of view, e.g., "knjizenje", "izvrsenje", etc.
	TipdokComboItems []domain.ComboItem   // Map of tipdok values for combo box
	NextNalog        string               // Next available nalog number as string
}

// NalogService defines the interface for operations related to Fnal (Nalogs).
type NalogService interface {
	Service[domain.Fnal]
	//GetNalogViewData fetches all data required to render the Nalog list page.
	GetNalogViewData(c *gin.Context, searchQuery, selectedTipdok string, page, pageSize int, isInitialRequest bool) (NalogViewData, error)
	SetNalogIDFieldName(string)
	SetSfTableFields([]domain.Fields)
	GetNaloziTableFields() []domain.Fields
	GetNalogPrepisData(c *gin.Context, searchQuery string, page, pageSize int) (NalogViewData, error)
	GetNalogStornirajData(c *gin.Context, searchQuery string, page, pageSize int) (NalogViewData, error)
	GetNextNalog(tipdok string) (int64, error)
	GetTipdokOptions() ([]domain.Tipdok, error)
	Validation(entity domain.Fnal) ([]domain.FieldError, error)
	ValidationKopirajNalog(entity domain.Fnal) ([]domain.FieldError, error)
	GetByTipdokNalog(tipdok string, nalog int64) (domain.Fnal, error)
}

// NalogResource implements the NalogService interface.
type NalogResource struct {
	service                 *BaseService[domain.Fnal]
	validator               validation.RuleBasedValidator[domain.Fnal]
	fnalRepo                repository.BaseRepository[domain.Fnal]
	tipdokRepo              repository.BaseRepository[domain.Tipdok]
	sfRepo                  repository.BaseRepository[domain.Sf]
	nalogIDFieldName        string
	naloziTableFields       []domain.Fields
	naloziStavkeTableFields []domain.Fields
	sfTableFields           []domain.Fields
}

func NewNalogService(
	service *BaseService[domain.Fnal],
	fnalRepo repository.BaseRepository[domain.Fnal],
	tipdokRepo repository.BaseRepository[domain.Tipdok],
	sfRepo repository.BaseRepository[domain.Sf],
	nalogIDFieldName string,
	naloziTableFields []domain.Fields,
	sfTableFields []domain.Fields,
	validator *validation.RuleBasedValidator[domain.Fnal],
) *NalogResource {
	rs := &NalogResource{
		service:           service,
		fnalRepo:          fnalRepo,
		tipdokRepo:        tipdokRepo,
		sfRepo:            sfRepo,
		nalogIDFieldName:  nalogIDFieldName,
		naloziTableFields: naloziTableFields,
		sfTableFields:     sfTableFields,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *NalogResource) SetNalogIDFieldName(nalogIDFieldName string) {
	s.nalogIDFieldName = nalogIDFieldName
}
func (s *NalogResource) SetNaloziTableFields(naloziTableFields []domain.Fields) {
	s.naloziTableFields = naloziTableFields
}
func (s *NalogResource) SetSfTableFields(sfTableFields []domain.Fields) {
	s.sfTableFields = sfTableFields
}

func (s *NalogResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.fieldCache == nil {
		s.service.fieldCache = make(map[string]reflect.StructField)
	}
	return s.service.fieldCache
}
func (s *NalogResource) Validation(entity domain.Fnal) ([]domain.FieldError, error) {
	fieldErrors, err := s.service.Validator.Validate(&entity)
	if err != nil {
		return fieldErrors, err
	}
	return fieldErrors, nil
}
func (s *NalogResource) ValidationKopirajNalog(entity domain.Fnal) ([]domain.FieldError, error) {
	fieldErrors, err := s.Validation(entity)
	if err != nil {
		return fieldErrors, err
	}

	return fieldErrors, nil
}

// Create implements NalogService.
func (s *NalogResource) Create(fnal *domain.Fnal, idField string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	// args := []interface{}{}
	// hasGod, hasKar := s.tipdokRepo.CheckGogKar()

	// qryText := `INSERT INTO fnal (god, kar, nalog, danal, datob, opis, xdatunosa, xopunosa)
	// 			VALUES () `
	// s.fnalRepo.CreateInsertStatement()
	return []domain.FieldError{}, 0, nil

}

// Delete implements NalogService.
func (s *NalogResource) Delete(idField string, id int64) error {
	return s.service.Delete(idField, id)
}

// GetAll implements NalogService.
func (s *NalogResource) GetAll(page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]domain.Fnal, error) {
	return s.service.GetAll(page, offset, tableFields, idField, searchParams...)
}

// GetAllCustom implements NalogService.
func (s *NalogResource) GetAllCustom(queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fnal, error) {
	return s.service.GetAllCustom(queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements NalogService.
func (s *NalogResource) GetByID(idField string, idValue int64) (*domain.Fnal, error) {
	return s.service.GetByID(idField, idValue)
}

// GetTotalRecords implements NalogService.
func (s *NalogResource) GetTotalRecords(tableFields []domain.Fields, searchParams ...string) (int, error) {
	return s.service.GetTotalRecords(tableFields, searchParams...)
}

// GetTotalRecordsCustom implements NalogService.
func (s *NalogResource) GetTotalRecordsCustom(queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements NalogService.
func (s *NalogResource) MapEntityToValues(entity *domain.Fnal, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

// Update implements NalogService.
func (s *NalogResource) Update(entity *domain.Fnal, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	return s.service.Update(entity, idField, idValue, tableFields)
}

// Update implements NalogService.
func (s *NalogResource) GetNextNalog(tipdok string) (int64, error) {
	// 2. Fetch Fnal Entities
	args := []interface{}{}
	// Use viewData.DefaultTipdok for the actual Fnal query
	whereText := s.buildFnalWhere("", "", &args)
	paramNbr := len(args)
	selectQuery := `SELECT COALESCE(MAX(nalog), 0) + 1 as nalog FROM fnal `

	whereText = fmt.Sprintf(`%s AND fnal.tipdok = $%d `, whereText, paramNbr+1)
	args = append(args, tipdok)
	entities, err := s.fnalRepo.GetAllCustom(selectQuery, whereText, args, "", "")
	if err != nil {
		return 0, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) == 0 {
		return 1, nil
	}

	return (*entities)[0].Nalog, nil

}
func (s *NalogResource) GetByTipdokNalog(tipdok string, nalog int64) (domain.Fnal, error) {
	entity := domain.Fnal{}
	args := []interface{}{}
	hasGod, hasKar := s.fnalRepo.CheckGogKar()
	basicWhere := s.fnalRepo.CreateBasicWhere(s.naloziTableFields, &args, hasGod, hasKar, "")
	selectQuery := `SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`
	whereText := fmt.Sprintf("%s AND tipdok = $%d AND nalog = $%d", basicWhere, len(args)+1, len(args)+2)
	args = append(args, tipdok, nalog)
	entities, err := s.fnalRepo.GetAllCustom(selectQuery, whereText, args, "", "")
	if err != nil {
		return entity, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) > 0 {
		entity = (*entities)[0]
	}
	return entity, nil
}

// Helper to construct common WHERE clauses and arguments for Fnal queries
func (s *NalogResource) buildFnalWhere(searchQuery, tipdok string, args *[]interface{}) string {
	hasGod, hasKar := s.fnalRepo.CheckGogKar()
	basicWhere := s.fnalRepo.CreateBasicWhere(s.naloziTableFields, args, hasGod, hasKar, searchQuery)

	var conditions []string
	if basicWhere != "" {
		conditions = append(conditions, basicWhere)
	}
	if tipdok != "" {
		// Ensure tipdok is properly escaped for SQL if it comes from user input and is directly inserted
		// For now, assuming it's safe or will be parameterized if db.Query takes it.
		// It's already in the basicWhere if it's in searchParams.
		conditions = append(conditions, fmt.Sprintf("(tipdok = '%s')", tipdok))
	}

	if len(conditions) == 0 {
		return ""
	}
	return strings.Join(conditions, " AND ")
}

// GetNalogViewData fetches all data required to render the Nalog list page.
func (s *NalogResource) GetNalogViewData(c *gin.Context, searchQuery, selectedTipdok string, page, pageSize int, isInitialRequest bool) (NalogViewData, error) {
	viewData := NalogViewData{
		IsInitialLoad: isInitialRequest,
	}

	// 1. Fetch Tipdok Options if it's an initial load or if selectedTipdok is empty
	// to determine the default.
	if isInitialRequest || selectedTipdok == "" {
		tipdokOptions, err := s.GetTipdokOptions()
		if err != nil {
			return viewData, fmt.Errorf("failed to get tipdok options: %w", err)
		}
		viewData.TipdokOptions = &tipdokOptions
		for _, td := range tipdokOptions {
			viewData.TipdokComboItems = append(viewData.TipdokComboItems, domain.ComboItem{Key: td.TipDok, Value: td.TipDok + " - " + td.Opis})
		}

		if selectedTipdok == "" && len(tipdokOptions) > 0 {
			viewData.DefaultTipdok = tipdokOptions[0].TipDok
		} else {
			viewData.DefaultTipdok = selectedTipdok // Use the provided one if available
		}
	} else {
		viewData.DefaultTipdok = selectedTipdok // Use the provided one directly
	}
	noviNalogBr, _ := s.GetNextNalog(viewData.DefaultTipdok)
	viewData.NextNalog = fmt.Sprintf("%d", noviNalogBr)
	// 2. Fetch Fnal Entities
	args := []interface{}{}
	// Use viewData.DefaultTipdok for the actual Fnal query
	whereText := s.buildFnalWhere(searchQuery, viewData.DefaultTipdok, &args)

	// Get total records
	totalRecordsQuery := `SELECT count(*) FROM fnal `

	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(totalRecordsQuery, whereText, args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}

	// Calculate pagination details
	currentPage, calculatedPageSize, totalPages := common.GetPaginationData(c, totRecords) // Pass nil for req
	if page > 0 {                                                                          // Override current page if provided from handler
		currentPage = page
	}
	if pageSize > 0 { // Override page size if provided from handler
		calculatedPageSize = pageSize
	}

	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", calculatedPageSize, (currentPage-1)*calculatedPageSize)
	orderBy := " ORDER BY danal DESC"
	selectQuery := `SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`

	entities, err := s.fnalRepo.GetAllCustom(selectQuery, whereText, args, limitOffset, orderBy)
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	viewData.FnalEntities = *entities

	// Prepare TableData for UI
	table := common.SetTableBasicData("NALOZI", "nalozi-table", s.naloziTableFields, "/api/nalozi/", "/api/nalozi/", calculatedPageSize, currentPage, totalPages, totRecords)
	// Additional table data configuration can happen here
	if searchQuery != "" { // If it's a search, the URL might change for HTMX
		table.URLGetAll = "/api/nalozi/all/search"
	} else {
		table.URLGetAll = "/api/nalozi/all/tipdok" // For tipdok specific updates
	}
	table.Pagination.HxInclude = "#tipdok, #search-input"
	table.HxInclude = "#tipdok, #search-input"
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
	viewData.TableData = table

	// 3. Fetch UkupnaObrada totals if it's an initial load
	if isInitialRequest {
		ukObrada, err := s.GetNalogTotals()
		if err != nil {
			return viewData, fmt.Errorf("failed to get Nalog totals: %w", err)
		}
		viewData.UkupnaObrada = &ukObrada
	}

	return viewData, nil
}

// GetNalogPrepisData fetches all data required to render the Nalog list page.
func (s *NalogResource) GetNalogPrepisData(c *gin.Context, searchQuery string, page, pageSize int) (NalogViewData, error) {
	viewData := NalogViewData{
		DefaultTipdok: "",
	}
	args := []interface{}{}
	// Use viewData.DefaultTipdok for the actual Fnal query
	whereText := s.buildFnalWhere(searchQuery, "", &args)

	// Get total records
	totalRecordsQuery := `SELECT count(*) FROM fnal `

	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(totalRecordsQuery, whereText, args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}

	// Calculate pagination details
	currentPage, calculatedPageSize, totalPages := common.GetPaginationData(c, totRecords) // Pass nil for req
	if page > 0 {                                                                          // Override current page if provided from handler
		currentPage = page
	}
	if pageSize > 0 { // Override page size if provided from handler
		calculatedPageSize = pageSize
	}

	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", calculatedPageSize, (currentPage-1)*calculatedPageSize)
	orderBy := " ORDER BY danal DESC"
	selectQuery := `SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`

	entities, err := s.fnalRepo.GetAllCustom(selectQuery, whereText, args, limitOffset, orderBy)
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	viewData.FnalEntities = *entities

	//get tipdok options
	tipdokOptions, err := s.GetTipdokOptions()
	if err != nil {
		return viewData, fmt.Errorf("failed to get tipdok options: %w", err)
	}
	viewData.TipdokOptions = &tipdokOptions
	for _, td := range tipdokOptions {
		viewData.TipdokComboItems = append(viewData.TipdokComboItems, domain.ComboItem{Key: td.TipDok, Value: td.TipDok + " - " + td.Opis})
	}

	// Prepare TableData for UI
	table := common.SetTableBasicData("NALOZI ZAGLAVLJE", "nalozi-table", s.naloziTableFields, "/api/nalozi/", "/api/nalozi/", calculatedPageSize, currentPage, totalPages, totRecords)
	// Additional table data configuration can happen here
	table.URLGetAll = "/api/nalozi/prepis"
	table.Pagination.HxInclude = "#search-input"
	table.HxInclude = "#search-input"
	table.DetailTarget = "#nalozi_kopiranje_stavke"
	table.ShowActions = true // Default, can be overridden by specific handler needs
	table.BtnAdd.IsVisible = false
	table.BtnUpdate.IsVisible = true
	table.BtnDelete.IsVisible = false
	table.BtnPrint.IsVisible = false
	table.BtnUpdate.LabelText = "Kopiraj"
	table.BtnUpdate.HxActionURL = "/api/nalozi/confirm-prepis" // Set the HxActionURL for the update button
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
	viewData.TableData = table

	return viewData, nil
}

// GetNalogStornirajData fetches all data required to render the Nalog list page.
func (s *NalogResource) GetNalogStornirajData(c *gin.Context, searchQuery string, page, pageSize int) (NalogViewData, error) {
	viewData := NalogViewData{
		DefaultTipdok: "",
	}
	args := []interface{}{}
	// Use viewData.DefaultTipdok for the actual Fnal query
	whereText := s.buildFnalWhere(searchQuery, "", &args)

	// Get total records
	totalRecordsQuery := `SELECT count(*) FROM fnal `

	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(totalRecordsQuery, whereText, args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}

	// Calculate pagination details
	currentPage, calculatedPageSize, totalPages := common.GetPaginationData(c, totRecords) // Pass nil for req
	if page > 0 {                                                                          // Override current page if provided from handler
		currentPage = page
	}
	if pageSize > 0 { // Override page size if provided from handler
		calculatedPageSize = pageSize
	}

	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", calculatedPageSize, (currentPage-1)*calculatedPageSize)
	orderBy := " ORDER BY danal DESC"
	selectQuery := `SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`

	entities, err := s.fnalRepo.GetAllCustom(selectQuery, whereText, args, limitOffset, orderBy)
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	viewData.FnalEntities = *entities

	// Prepare TableData for UI
	table := common.SetTableBasicData("NALOZI ZAGLAVLJE", "nalozi-table", s.naloziTableFields, "/api/nalozi/", "/api/nalozi/", calculatedPageSize, currentPage, totalPages, totRecords)
	btnStorniraj := domain.Button{
		LabelText:     "Storniraj",
		HxActionURL:   "/api/nalozi/confirm-storniraj",
		IsVisible:     true,
		Id:            "btn-storniraj",
		HxRequestType: "POST",
	}
	// Additional table data configuration can happen here
	table.URLGetAll = "/api/nalozi/storniranje"
	table.Pagination.HxInclude = "#search-input"
	table.HxInclude = "#search-input"
	table.DetailTarget = "#nalozi_storniraj_stavke"
	table.ShowActions = true // Default, can be overridden by specific handler needs
	table.BtnAdd.IsVisible = false
	table.BtnUpdate.IsVisible = true
	table.BtnDelete.IsVisible = true
	table.BtnPrint.IsVisible = false
	table.BtnUpdate = btnStorniraj
	table.BtnDelete.LabelText = "Obriši"
	table.BtnUpdate.HxActionURL = "/api/nalozi/confirm-storniraj" // Set the HxActionURL for the update button
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
	viewData.TableData = table

	return viewData, nil
}

// GetNalogTotals fetches aggregated data from the 'sf' table.
// This method stays the same, as it's a specific business operation.
func (s *NalogResource) GetNalogTotals() (domain.UkupnaObrada, error) {
	args := []interface{}{}

	hasGod, hasKar := s.sfRepo.CheckGogKar()
	basicWhere := s.sfRepo.CreateBasicWhere(s.sfTableFields, &args, hasGod, hasKar)
	qryText := `SELECT COALESCE(SUM(brst), 0) as brst, COALESCE(SUM(brna), 0) as brna, COALESCE(SUM(dug), 0) as dug, COALESCE(SUM(pot), 0) as pot FROM sf` // Ensure these columns exist in domain.Sf
	// Use sqlx.Get to directly map the aggregated row to Sf
	totalsSf, err := s.sfRepo.GetAllCustom(qryText, basicWhere, args, "", "") // Assuming a GetCustom method for single row
	if err != nil {
		return domain.UkupnaObrada{}, fmt.Errorf("failed to get SF totals: %w", err)
	}

	if totalsSf != nil && len(*totalsSf) > 0 {
		ukObrada := domain.UkupnaObrada{
			Duguje:    common.FormatNumberWithSystemLocale((*totalsSf)[0].Dug, 2),
			Potrazuje: common.FormatNumberWithSystemLocale((*totalsSf)[0].Pot, 2),
			UkStavki:  common.FormatNumberWithSystemLocale((*totalsSf)[0].Brst, 0),
			UkNaloga:  common.FormatNumberWithSystemLocale((*totalsSf)[0].Brna, 0),
		}
		return ukObrada, nil
	}
	return domain.UkupnaObrada{}, nil
}

// GetTipdokOptions fetches the list of tipdok options for filtering.
// This method stays the same.
func (s *NalogResource) GetTipdokOptions() ([]domain.Tipdok, error) {
	args := []interface{}{}
	hasGod, hasKar := s.tipdokRepo.CheckGogKar()
	basicWhere := s.tipdokRepo.CreateBasicWhere(nil, &args, hasGod, hasKar) // No specific fields for tipdok

	qryText := `SELECT idtipdok, tipdok, opis FROM tipdok`
	whereText := fmt.Sprintf(" %s AND (grpdok = 'FIN' OR grpdok = 'SVI') ", basicWhere)
	orderBy := " ORDER BY tipdok"

	tipdokValues, err := s.tipdokRepo.GetAllCustom(qryText, whereText, args, "", orderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get tipdok options: %w", err)
	}
	return *tipdokValues, nil
}

func (s *NalogResource) GetNaloziTableFields() []domain.Fields {
	if len(s.naloziTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziTableFields
}

// ValidateCopyNalog validates the data for copying a nalog (voucher)
func (s *NalogResource) ValidateCopyNalog(danalStr, datobStr string, brNaloga int64) []domain.FieldError {
	var fieldErrors []domain.FieldError

	// Parse and validate danal (naloga date)
	var dPomDate time.Time
	var err error

	if strings.Contains(danalStr, ".") {
		dPomDate, err = time.Parse("02.01.2006", danalStr)
	} else {
		dPomDate, err = time.Parse("2006-01-02", danalStr)
	}

	if err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "danal",
			ErrorMessage: "morate uneti korektan datum naloga",
		})
	} else {
		// Check if year matches current business year
		if dPomDate.Year() != global.GetGnGod() {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "danal",
				ErrorMessage: fmt.Sprintf("nekorektan datum naloga, godina mora biti jednaka poslovnoj %d", global.GetGnGod()),
			})
		}
	}

	// Parse and validate datob (obrade date)
	if strings.Contains(datobStr, ".") {
		dPomDate, err = time.Parse("02.01.2006", datobStr)
	} else {
		dPomDate, err = time.Parse("2006-01-02", datobStr)
	}

	if err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "datob",
			ErrorMessage: "morate uneti korektan datum obrade naloga",
		})
	} else {
		// Check if year matches current business year
		if dPomDate.Year() != global.GetGnGod() {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "datob",
				ErrorMessage: fmt.Sprintf("nekorektan datum obrade, godina mora biti jednaka poslovnoj %d", global.GetGnGod()),
			})
		}
	}

	// Validate broj naloga (voucher number)
	if brNaloga == 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nalog",
			ErrorMessage: "morate uneti broj naloga",
		})
	}

	return fieldErrors
}

func (s *NalogResource) setServiceFieldValues() {
	s.naloziTableFields = []domain.Fields{
		{Name: "rbr", Label: "R. Broj", Width: "4", Field: "rbr", Sortable: true},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "6", Field: "tipdok", Sortable: true},
		{Name: "nalog", Label: "Br. Naloga", Width: "12", Field: "nalog", Sortable: true},
		{Name: "danal", Label: "Datum naloga", Width: "12", Field: "danal", Sortable: true},
		{Name: "opis", Label: "opis ", Width: "60", Field: "opis"},
		{Name: "dug", Label: "Duguje", Width: "14", Field: "dug"},
		{Name: "pot", Label: "Potrazuje", Width: "14", Field: "pot"},
		{Name: "datob", Label: "Datum obrade", Width: "12", Field: "datob"},
		{Name: "brst", Label: "Br.Stavki", Width: "5", Field: "brst"},
		{Name: "nalsts", Label: "Status naloga", Width: "10", Field: "nalsts"},
	}
	s.naloziStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "R. Broj", Width: "4"},
		{Name: "konto", Label: "Konto", Width: "6"},
		{Name: "sifra", Label: "Sifra", Width: "6"},
		{Name: "naziv", Label: "Naziv Konta", Width: "60"},
		{Name: "vrd", Label: "Vrsta Dok.", Width: "10"},
		{Name: "opis", Label: "Opis", Width: "60"},
		{Name: "dug", Label: "Duguje", Width: "14"},
		{Name: "pot", Label: "Potrazuje", Width: "14"},
		{Name: "dokum", Label: "Br. Dokum", Width: "12"},
		{Name: "dadok", Label: "Datum Dok.", Width: "12"},
	}
}
