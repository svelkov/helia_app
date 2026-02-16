package finansijsko

import (
	"database/sql"
	"errors"
	"fmt"
	"helia/config"
	"helia/global"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	finval "helia/internal/validation/finansijsko"
	"log"
	"strconv"

	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	naloziTableID string = "nalozi-table"
	ActionAdd     string = "add"
	ActionUpdate  string = "update"
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
	service.Service[domain.Fnal]
	//GetNalogViewData fetches all data required to render the Nalog list page.
	CreateNalog(c *gin.Context, fnal *domain.Fnal, idField string, fields []domain.Fields) (frpoPaylad *domain.FproPayload, fieldsError []domain.FieldError, lastInsertedId int64, err error)
	GetNalogViewData(c *gin.Context, searchQuery, selectedTipdok string, page, pageSize int, isInitialRequest bool) (NalogViewData, error)
	SetNalogIDFieldName(string)
	SetSfTableFields([]domain.Fields)
	GetNaloziTableFields() []domain.Fields
	GetNaloziStavkeTableFields() []domain.Fields
	GetNalogPrepisData(c *gin.Context, searchQuery string, page, pageSize int) (NalogViewData, error)
	GetNalogStornirajData(c *gin.Context, searchQuery string, page, pageSize int) (NalogViewData, error)
	GetNextNalog(c *gin.Context, tipdok string) (int64, error)
	GetTipdokOptions(c *gin.Context) ([]domain.Tipdok, error)
	Validation(entity domain.Fnal) ([]domain.FieldError, error)
	ValidationKopirajNalog(entity domain.Fnal) ([]domain.FieldError, error)
	GetByTipdokNalog(c *gin.Context, tipdok string, nalog int64) (domain.Fnal, error)
	GetIdTipdokByTipdok(c *gin.Context, tipdok string) (int64, error)
	GetOrgJedinice(c *gin.Context) ([]domain.ComboItem, error)
	GetMestoTroska(c *gin.Context, idorgjed int64) ([]domain.ComboItem, error)
	MapReqToEntity(c *gin.Context, req domain.FnalPayload, entity *domain.Fnal, action string)
	MapToFproPayload(fproPayload *domain.FproPayload, entity domain.Fnal, lastInsertedID int64)
	UpdateNalog(c *gin.Context) (fproPayload domain.FproPayload, tblStavke domain.TableData, fieldErrors []domain.FieldError, err error)
}

// NalogResource implements the NalogService interface.
type NalogResource struct {
	service                 *service.BaseService[domain.Fnal]
	fproService             FproService
	validator               *finval.FnalValidator
	fnalRepo                repository.BaseRepository[domain.Fnal]
	tipdokRepo              repository.BaseRepository[domain.Tipdok]
	sfRepo                  repository.BaseRepository[domain.Sf]
	ojRepo                  repository.BaseRepository[domain.Orgjed]
	mtroskaRepo             repository.BaseRepository[domain.Mestotr]
	nalogIDFieldName        string
	naloziTableFields       []domain.Fields
	naloziStavkeTableFields []domain.Fields
	sfTableFields           []domain.Fields
	cfg                     config.Config
}

func NewNalogService(
	service *service.BaseService[domain.Fnal],
	fproService FproService,
	validator *finval.FnalValidator,
	fnalRepo repository.BaseRepository[domain.Fnal],
	tipdokRepo repository.BaseRepository[domain.Tipdok],
	sfRepo repository.BaseRepository[domain.Sf],
	ojRepo repository.BaseRepository[domain.Orgjed],
	mtroskaRepo repository.BaseRepository[domain.Mestotr],
	nalogIDFieldName string,
	naloziTableFields []domain.Fields,
	sfTableFields []domain.Fields,
	cfg config.Config,
) *NalogResource {
	rs := &NalogResource{
		service:           service,
		fproService:       fproService,
		fnalRepo:          fnalRepo,
		tipdokRepo:        tipdokRepo,
		sfRepo:            sfRepo,
		ojRepo:            ojRepo,
		mtroskaRepo:       mtroskaRepo,
		nalogIDFieldName:  nalogIDFieldName,
		naloziTableFields: naloziTableFields,
		sfTableFields:     sfTableFields,
		validator:         validator,
		cfg:               cfg,
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
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}
func (s *NalogResource) Validation(entity domain.Fnal) ([]domain.FieldError, error) {
	fieldErrors, err := s.validator.Validate(&entity)
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
func (s *NalogResource) Create(c *gin.Context, fnal *domain.Fnal, idField string, fields []domain.Fields) (fieldsError []domain.FieldError, lastInsertedID int64, err error) {
	return s.service.Create(c, fnal, idField, fields)
}

// Delete implements NalogService.
func (s *NalogResource) Delete(idField string, id int64) error {
	return s.service.Delete(idField, id)
}

// GetAll implements NalogService.
func (s *NalogResource) GetAll(c *gin.Context, page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]domain.Fnal, error) {
	return s.service.GetAll(c, page, offset, tableFields, idField, searchParams...)
}

// GetAllCustom implements NalogService.
func (s *NalogResource) GetAllCustom(c *gin.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fnal, error) {
	return s.service.GetAllCustom(c, queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements NalogService.
func (s *NalogResource) GetByID(idField string, idValue int64) (*domain.Fnal, error) {
	return s.service.GetByID(idField, idValue)
}

// GetTotalRecords implements NalogService.
func (s *NalogResource) GetTotalRecords(c *gin.Context, tableFields []domain.Fields, searchParams ...string) (int, error) {
	return s.service.GetTotalRecords(c, tableFields, searchParams...)
}

// GetTotalRecordsCustom implements NalogService.
func (s *NalogResource) GetTotalRecordsCustom(c *gin.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(c, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements NalogService.
func (s *NalogResource) MapEntityToValues(entity *domain.Fnal, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

// Update implements NalogService.
func (s *NalogResource) Update(c *gin.Context, entity *domain.Fnal, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	return s.service.Update(c, entity, idField, idValue, tableFields)
}

// Create implements NalogService.
func (s *NalogResource) CreateNalog(c *gin.Context, fnal *domain.Fnal, idField string, fields []domain.Fields) (frpoPaylad *domain.FproPayload, fieldsError []domain.FieldError, lastInsertedID int64, err error) {

	var nalog domain.FnalPayload
	var fproPayload domain.FproPayload
	if err := c.ShouldBind(&nalog); err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}

	s.MapReqToEntity(c, nalog, fnal, ActionAdd)
	fieldErrors, err := s.validator.Validate(fnal)
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	if len(fieldErrors) > 0 {
		return &fproPayload, fieldErrors, 0, nil
	}

	// map request to entity
	id, err := s.GetIdTipdokByTipdok(c, nalog.Tipdok)
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	fnal.IDTipdok = id
	log.Println("Creating Nalog with Tipdok ID:", id, " Tipdok:", nalog.Tipdok, " Nalog:", fnal.Nalog, " Danal:", fnal.Danal.Format("2006-01-02"))
	fieldErrors, lastInsertedID, err = s.service.Create(c, fnal, common.IDfnal, s.MapEntityToValues(fnal, s.GetNaloziTableFields()))
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	if len(fieldErrors) > 0 {
		return &fproPayload, fieldErrors, lastInsertedID, nil
	}
	// Lock the header using global resource lock
	mutex := global.GetHeaderLock(lastInsertedID)
	// Try to lock without blocking
	if !mutex.TryLock() {
		return &fproPayload, []domain.FieldError{}, 0, errors.New(common.ErrMsgStatusConflict)
	}
	defer mutex.Unlock() // Ensure mutex is always unlocked, even if an error occurs
	s.MapToFproPayload(&fproPayload, *fnal, lastInsertedID)

	orgjed, err := s.GetOrgJedinice(c)
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	mtroska, err := s.GetMestoTroska(c, 0) // No orgjed yet for new nalog
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	fproPayload.Orgjed = orgjed
	fproPayload.Mtroska = mtroska
	return &fproPayload, []domain.FieldError{}, lastInsertedID, nil
}

func (s *NalogResource) UpdateNalog(c *gin.Context) (fproPayload domain.FproPayload, tblStavke domain.TableData, fieldErrors []domain.FieldError, err error) {
	var entity *domain.Fnal
	var nalog domain.FnalPayload

	fnalStr := c.Param("id")
	if err := c.ShouldBind(&nalog); err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}

	fnalID, err := strconv.ParseInt(fnalStr, 10, 64)
	if err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}

	// Acquire lock using client ID (unique per browser tab)
	clientID := c.GetHeader("X-Client-ID")
	if clientID == "" {
		clientID = c.Query("client_id")
	}

	// Get username from claims or session
	username := "unknown"
	userClaims := domain.GetCurrentUserClaims(c)
	if userClaims != nil && userClaims.Username != "" {
		username = userClaims.Username
	} else {
		session := domain.GetSessionFromContext(c)
		if session != nil && session.UserName != "" {
			username = session.UserName
		}
	}

	if clientID == "" {
		// Fallback: generate unique ID
		session := domain.GetSessionFromContext(c)
		clientID = fmt.Sprintf("%s_%d_%d_%d", username, session.SelectedGod, session.SelectedKar, time.Now().UnixNano())
	}

	// Try to acquire lock
	acquired, existingLock, err := global.TryLockNalog(fnalID, username, clientID)
	if err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}
	if !acquired {
		errorMsg := global.FormatLockError(existingLock)
		return fproPayload, tblStavke, []domain.FieldError{}, fmt.Errorf(errorMsg)
	}

	entity, err = s.service.GetByID(common.IDfnal, fnalID)
	if err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}
	// map request to entity
	s.MapReqToEntity(c, nalog, entity, ActionUpdate)

	fieldErrors, err = s.Update(c, entity, common.IDfnal, fnalID, s.MapEntityToValues(entity, s.GetNaloziTableFields()))
	if err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}
	if len(fieldErrors) > 0 {
		return fproPayload, tblStavke, fieldErrors, fmt.Errorf(common.ErrMsgValidation)
	}

	s.MapToFproPayload(&fproPayload, *entity, fnalID)

	fproEntities, err := s.fproService.GetAllByFnalID(c, fnalID)
	if err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}

	// // Get session for god/kar values
	// session = domain.GetSessionFromContext(c)
	// if session == nil {
	// 	return fproPayload, tblStavke, []domain.FieldError{}, fmt.Errorf("user session not found")
	// }

	arrOrgjed := []domain.ComboItem{{Key: "", Value: ""}} //add empty element in combo box
	orgjed, err := s.GetOrgJedinice(c)
	if err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}
	arrOrgjed = append(arrOrgjed, orgjed...)
	mtroska := []domain.ComboItem{{Key: "", Value: ""}}
	fproPayload.Orgjed = arrOrgjed
	fproPayload.Mtroska = mtroska
	// because the stavke table is specific we could not use template helper for table and table rows
	tblStavke = common.SetTableBasicData("Stavke Naloga", naloziTableID+"_stavke", s.GetNaloziStavkeTableFields(), "", "", 10, 0, 0, 0, s.cfg)
	tblStavke.ShowPagination = true
	tblStavke.SearchEnabled = true

	tblRows, err := common.SetTableRows(&tblStavke, *fproEntities, s.GetNaloziStavkeTableFields(), "idfpro", "", s.fproService.GetFieldCache())
	if err != nil {
		return fproPayload, tblStavke, []domain.FieldError{}, err
	}
	tblStavke.Rows = tblRows.Rows
	return fproPayload, tblStavke, []domain.FieldError{}, nil
}

// Update implements NalogService.
func (s *NalogResource) GetNextNalog(c *gin.Context, tipdok string) (int64, error) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return 0, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(` SELECT COALESCE(MAX(nalog), 0) + 1 as nalog FROM fnal`)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	sqlQuery, args := qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return 0, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) == 0 {
		return 1, nil
	}

	return (*entities)[0].Nalog, nil

}
func (s *NalogResource) GetByTipdokNalog(c *gin.Context, tipdok string, nalog int64) (domain.Fnal, error) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return domain.Fnal{}, fmt.Errorf("user session not found")
	}

	entity := domain.Fnal{}
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(` SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	qb.AddEqual("nalog", nalog)
	sqlQuery, args := qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return entity, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) > 0 {
		entity = (*entities)[0]
	}
	return entity, nil
}
func (s *NalogResource) GetIdTipdokByTipdok(c *gin.Context, tipdok string) (int64, error) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return 0, fmt.Errorf("user session not found")
	}

	entity := domain.Tipdok{}
	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(` SELECT idtipdok, tipdok, opis FROM tipdok`)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	sqlQuery, args := qb.Build()
	tipdokEntities, err := s.tipdokRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return 0, errors.New(common.ErrMsgGetData)
	}
	if len(*tipdokEntities) > 0 {
		entity = (*tipdokEntities)[0]
		return int64(entity.IDTipDok), nil
	}
	return 0, errors.New(common.ErrNoDataFound)
}

// GetNalogViewData fetches all data required to render the Nalog list page.
func (s *NalogResource) GetNalogViewData(c *gin.Context, searchQuery, selectedTipdok string, page, pageSize int, isInitialRequest bool) (NalogViewData, error) {
	viewData := NalogViewData{
		IsInitialLoad: isInitialRequest,
	}

	// Get session for god/kar values at the start
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return viewData, fmt.Errorf("user session not found")
	}

	// 1. Fetch Tipdok Options on initial load or if selectedTipdok is empty
	// to determine the default.
	if isInitialRequest || selectedTipdok == "" {
		tipdokOptions, err := s.GetTipdokOptions(c)
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
	noviNalogBr, _ := s.GetNextNalog(c, viewData.DefaultTipdok)
	viewData.NextNalog = fmt.Sprintf("%d", noviNalogBr)

	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	// Get total records
	qbCount := common.NewQueryBuilder(` SELECT count(*) FROM fnal`)
	qbCount.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qbCount.AddEqual("tipdok", viewData.DefaultTipdok)
	sqlQuery, args := qbCount.Build()
	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}

	// Calculate pagination details
	currentPage, calculatedPageSize, totalPages := common.GetPaginationData(c, totRecords, s.cfg) // Pass nil for req
	if page > 0 {                                                                                 // Override current page if provided from handler
		currentPage = page
	}
	if pageSize > 0 { // Override page size if provided from handler
		calculatedPageSize = pageSize
	}

	qb := common.NewQueryBuilder(` SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", viewData.DefaultTipdok)
	qb.AddOrderBy(`danal DESC`)
	qb.SetLimit(calculatedPageSize)
	qb.SetOffset((currentPage - 1) * calculatedPageSize)
	sqlQuery, args = qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	viewData.FnalEntities = *entities

	// Prepare TableData for UI
	table := common.SetTableBasicData("NALOZI", "nalozi-table", s.naloziTableFields, "/api/nalozi/", "/api/nalozi/", calculatedPageSize, currentPage, totalPages, totRecords, s.cfg)
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
		ukObrada, err := s.GetNalogTotals(c)
		if err != nil {
			return viewData, fmt.Errorf("failed to get Nalog totals: %w", err)
		}
		viewData.UkupnaObrada = &ukObrada
	}

	return viewData, nil
}

// GetNalogPrepisData fetches all data required to render the Nalog list page.
func (s *NalogResource) GetNalogPrepisData(c *gin.Context, searchQuery string, page, pageSize int) (NalogViewData, error) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return NalogViewData{}, fmt.Errorf("user session not found")
	}
	viewData := NalogViewData{
		DefaultTipdok: "",
	}

	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()

	qbCount := common.NewQueryBuilder(`SELECT count(*) FROM fnal`)
	qbCount.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	sqlQuery, args := qbCount.Build()
	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}

	// Calculate pagination details
	currentPage, calculatedPageSize, totalPages := common.GetPaginationData(c, totRecords, s.cfg) // Pass nil for req
	if page > 0 {                                                                                 // Override current page if provided from handler
		currentPage = page
	}
	if pageSize > 0 { // Override page size if provided from handler
		calculatedPageSize = pageSize
	}

	qb := common.NewQueryBuilder(`SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddOrderBy(`danal DESC`)
	qb.SetLimit(calculatedPageSize)
	qb.SetOffset((currentPage - 1) * calculatedPageSize)

	sqlQuery, args = qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	viewData.FnalEntities = *entities

	//get tipdok options
	tipdokOptions, err := s.GetTipdokOptions(c)
	if err != nil {
		return viewData, fmt.Errorf("failed to get tipdok options: %w", err)
	}
	viewData.TipdokOptions = &tipdokOptions
	for _, td := range tipdokOptions {
		viewData.TipdokComboItems = append(viewData.TipdokComboItems, domain.ComboItem{Key: td.TipDok, Value: td.TipDok + " - " + td.Opis})
	}

	// Prepare TableData for UI
	table := common.SetTableBasicData("NALOZI ZAGLAVLJE", "nalozi-table", s.naloziTableFields, "/api/nalozi/", "/api/nalozi/", calculatedPageSize, currentPage, totalPages, totRecords, s.cfg)
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

	// Get session for god/kar values
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return viewData, fmt.Errorf("user session not found")
	}

	qbCount := common.NewQueryBuilder(`SELECT count(*) FROM fnal`)
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qbCount.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	sqlQuery, args := qbCount.Build()
	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}

	// Calculate pagination details
	currentPage, calculatedPageSize, totalPages := common.GetPaginationData(c, totRecords, s.cfg) // Pass nil for req
	if page > 0 {                                                                                 // Override current page if provided from handler
		currentPage = page
	}
	if pageSize > 0 { // Override page size if provided from handler
		calculatedPageSize = pageSize
	}

	qb := common.NewQueryBuilder(`SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddOrderBy(`danal DESC`)
	qb.SetLimit(calculatedPageSize)
	qb.SetOffset((currentPage - 1) * calculatedPageSize)
	sqlQuery, args = qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	viewData.FnalEntities = *entities

	// Prepare TableData for UI
	table := common.SetTableBasicData("NALOZI ZAGLAVLJE", "nalozi-table", s.naloziTableFields, "/api/nalozi/", "/api/nalozi/", calculatedPageSize, currentPage, totalPages, totRecords, s.cfg)
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
func (s *NalogResource) GetNalogTotals(c *gin.Context) (domain.UkupnaObrada, error) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return domain.UkupnaObrada{}, fmt.Errorf("user session not found")
	}

	qb := common.NewQueryBuilder(`SELECT COALESCE(SUM(dug), 0) as dug, COALESCE(SUM(pot), 0) as pot, COALESCE(SUM(brst), 0) as brst, COALESCE(SUM(brna), 0) as brna FROM sf`)
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	sqlQuery, args := qb.Build()
	totalsSf, err := s.sfRepo.GetAllCustom(c, sqlQuery, "", args, "", "") // Assuming a GetCustom method for single row
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
func (s *NalogResource) GetTipdokOptions(c *gin.Context) ([]domain.Tipdok, error) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return nil, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idtipdok, tipdok, opis FROM tipdok`)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddCustomCondition("(grpdok = 'FIN' OR grpdok = 'SVI')")
	qb.AddOrderBy("tipdok::NUMERIC ASC")
	sqlQuery, args := qb.Build()
	tipdokValues, err := s.tipdokRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tipdok options: %w", err)
	}
	return *tipdokValues, nil
}

func (s *NalogResource) GetOrgJedinice(c *gin.Context) ([]domain.ComboItem, error) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	args := []interface{}{}
	hasGod, hasKar := s.ojRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idorgjed, ojozn, naziv FROM orgjed`)
	qb.AddGodKarConditions(hasGod, hasKar, session.SelectedGod, session.SelectedKar)
	qb.AddOrderBy("ojozn ASC")
	sqlQuery, args := qb.Build()
	ojEntites, err := s.ojRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, errors.New(common.ErrMsgGetData)
	}
	for _, oj := range *ojEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", oj.IDOrgjed),
			Value: fmt.Sprintf("%s - %s", oj.OjOzn, oj.Naziv),
		})
	}
	return comboItems, nil
}

func (s *NalogResource) GetMestoTroska(c *gin.Context, idorgjed int64) ([]domain.ComboItem, error) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		return nil, fmt.Errorf("user session not found")
	}

	comboItems := []domain.ComboItem{}
	qb := common.NewQueryBuilder(`SELECT mestotrid, mtroska, opis, idorgjed FROM mestotr`)
	hasGod, hasKar := s.mtroskaRepo.GetHasGodHasKar()
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddCondition("idorgjed", idorgjed, "=")
	qb.AddOrderBy("mtroska ASC")
	sqlQuery, args := qb.Build()
	mtroskaEntites, err := s.mtroskaRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return comboItems, errors.New(common.ErrMsgGetData)
	}
	for _, mtroska := range *mtroskaEntites {
		comboItems = append(comboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", mtroska.MestoTrID),
			Value: fmt.Sprintf("%s - %s", mtroska.Mtroska, mtroska.Opis),
		})
	}
	return comboItems, nil
}

func (s *NalogResource) GetNaloziTableFields() []domain.Fields {
	if len(s.naloziTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziTableFields
}
func (s *NalogResource) GetNaloziStavkeTableFields() []domain.Fields {
	if len(s.naloziStavkeTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziStavkeTableFields
}

// ValidateCopyNalog validates the data for copying a nalog (voucher)
func (s *NalogResource) ValidateCopyNalog(entity domain.Fnal, danalStr, datobStr string, brNaloga int64) []domain.FieldError {
	var fieldErrors []domain.FieldError

	// Parse and validate danal (naloga date)
	var dPomDate time.Time
	var err error

	if strings.Contains(danalStr, ".") {
		dPomDate, err = time.Parse(common.DateLayout, danalStr)
	} else {
		dPomDate, err = time.Parse(common.HtmlLayout, danalStr)
	}

	if err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "danal",
			ErrorMessage: "morate uneti korektan datum naloga",
		})
	} else {
		// Check if year matches current business year
		if dPomDate.Year() != entity.God {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "danal",
				ErrorMessage: fmt.Sprintf("nekorektan datum naloga, godina mora biti jednaka poslovnoj %d", entity.God),
			})
		}
	}

	// Parse and validate datob (obrade date)
	if strings.Contains(datobStr, ".") {
		dPomDate, err = time.Parse(common.DateLayout, datobStr)
	} else {
		dPomDate, err = time.Parse(common.HtmlLayout, datobStr)
	}

	if err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "datob",
			ErrorMessage: "morate uneti korektan datum obrade naloga",
		})
	} else {
		// Check if year matches current business year
		if dPomDate.Year() != entity.God {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "datob",
				ErrorMessage: fmt.Sprintf("nekorektan datum obrade, godina mora biti jednaka poslovnoj %d", entity.God),
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
		{Name: "opis", Label: "opis ", Width: "60", Field: "opis", Sortable: true},
		{Name: "dug", Label: "Duguje", Width: "14", Field: "dug", Sortable: true},
		{Name: "pot", Label: "Potrazuje", Width: "14", Field: "pot", Sortable: true},
		{Name: "datob", Label: "Datum obrade", Width: "12", Field: "datob", Sortable: true},
		{Name: "brst", Label: "Br.Stavki", Width: "5", Field: "brst", Sortable: true},
		{Name: "nalsts", Label: "Status naloga", Width: "10", Field: "nalsts", Sortable: true},
		{Name: "idtipdok", Label: "Id tipdok", Width: "10", Field: "idtipdok"},
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
		{Name: "idorgjed", Label: "Id orgjed", Width: "10", Field: "idorgjed"},
	}
}

func (s *NalogResource) MapReqToEntity(c *gin.Context, req domain.FnalPayload, entity *domain.Fnal, action string) {
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		// This should not happen as middleware ensures session exists
		return
	}

	userClaims := domain.GetCurrentUserClaims(c)
	username := ""
	if userClaims != nil {
		username = userClaims.Username
	}

	if action == "add" {
		entity.God = userSession.SelectedGod
		entity.Kar = userSession.SelectedKar
		entity.Xdatunosa = sql.NullTime{Time: time.Now(), Valid: true}
		entity.Xopunos = sql.NullString{String: username, Valid: true}
	}
	if action == "update" {
		entity.Xdatizmene = sql.NullTime{Time: time.Now(), Valid: true}
		entity.Xopizmene = sql.NullString{String: username, Valid: true}
	}
	entity.Nalog = common.StringToInt64(req.Nalog)
	entity.Danal = common.StringToDate(req.Danal)
	entity.Datob = common.StringToDate(req.Datob)
	entity.Tipdok = req.Tipdok

}
func (s *NalogResource) MapToFproPayload(fproPayload *domain.FproPayload, entity domain.Fnal, lastInsertedID int64) {

	fproPayload.IDFnal = entity.IDFnal
	fproPayload.Tipdok = entity.Tipdok
	fproPayload.Nalog = fmt.Sprintf("%d", entity.Nalog)
	fproPayload.Danal = entity.Danal.Format(common.HtmlLayout)
	fproPayload.Datob = entity.Datob.Format(common.HtmlLayout)
	fproPayload.Opis = entity.Opis
	fproPayload.Duguje = common.FormatNumberWithSystemLocale(entity.Dug, 2)
	fproPayload.Potrazuje = common.FormatNumberWithSystemLocale(entity.Pot, 2)
	fproPayload.Saldo = common.FormatNumberWithSystemLocale(entity.Dug-entity.Pot, 2)
}
