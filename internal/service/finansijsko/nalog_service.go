package finansijsko

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	finval "helia/internal/validation/finansijsko"
	"log"
	"strings"

	"reflect"
	"time"
)

const (
	naloziStavkeTableID string = "nalog-stavke-table"
	ActionAdd           string = "add"
	ActionUpdate        string = "update"
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
	CreateNalog(ctx context.Context, fnal *domain.Fnal, idField string, fields []domain.Fields, currentPage, pageSize int, searchText string) (frpoPaylad *domain.FproPayload, fieldsError []domain.FieldError, lastInsertedId int64, err error)
	GetNalogViewData(ctx context.Context, tbl *domain.TableData, searchQuery string, page, pageSize int, isInitialRequest bool, sortBy, sortOrder, tipdok string) (NalogViewData, error)
	GetNalogPrepisData(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, searchText, sortBy, sortOrder string) error
	GetNalogStornirajData(ctx context.Context, tbl *domain.TableData, page, pageSize int, searchText, sortBy, sortOrder string) (NalogViewData, error)
	GetNextNalog(ctx context.Context, tipdok string) (int64, error)
	GetTipdokOptions(ctx context.Context) ([]domain.Tipdok, error)
	NalogValidation(ctx context.Context, entity domain.Fnal, action string) ([]domain.FieldError, error)
	ValidateCopyNalog(ctx context.Context, req domain.FnalPayload, entity domain.Fnal) ([]domain.FieldError, error)
	GetByTipdokNalog(ctx context.Context, tipdok string, nalog int64) (domain.Fnal, error)
	GetIdTipdokByTipdok(ctx context.Context, tipdok string) (int64, error)
	MapReqToEntity(ctx context.Context, req domain.FnalPayload, entity *domain.Fnal, action string)
	MapToFproPayload(fproPayload *domain.FproPayload, entity domain.Fnal, lastInsertedID int64)
	UpdateNalog(ctx context.Context, fnalID int64, payload *domain.FnalPayload, tblStavke *domain.TableData, currentPage, pageSize int, searchText string) (fproPayload domain.FproPayload, fieldErrors []domain.FieldError, err error)
	KopirajNalog(ctx context.Context, idFnal int64, entity domain.Fnal) error
	CheckNalogExistForCopy(ctx context.Context, entity domain.Fnal) (bool, int64, error)
	KopirajNalogToExisting(ctx context.Context, req domain.FnalPayload, target_IdFnal int64) error
	StornirajNalog(ctx context.Context, idFnal int64, entity domain.Fnal) error
	GetNalogStampanjeData(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, page, pageSize int, searchText, sortBy, sortOrder string) error
	GetNalogStampanjeDataDetalji(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, page, pageSize int, idFnal int64, searchText, sortBy, sortOrder string) error
	SetNalogIDFieldName(string)
	GetNaloziTableFields() []domain.Fields
	GetNaloziStavkeTableFields() []domain.Fields
	GetNaloziStampaTableFields() []domain.Fields
	GetNaloziStavkeStampaTableFields() []domain.Fields
}

// NalogResource implements the NalogService interface.
type NalogResource struct {
	service                       *service.BaseService[domain.Fnal]
	fproService                   FproService
	validator                     *finval.FnalValidator
	fnalRepo                      repository.BaseRepository[domain.Fnal]
	fproRepo                      repository.BaseRepository[domain.Fpro]
	tipdokRepo                    repository.BaseRepository[domain.Tipdok]
	sfRepo                        repository.BaseRepository[domain.Sf]
	ojRepo                        repository.BaseRepository[domain.Orgjed]
	mtroskaRepo                   repository.BaseRepository[domain.Mestotr]
	fvrRepo                       repository.BaseRepository[domain.Fvr]
	nalogIDFieldName              string
	naloziTableFields             []domain.Fields
	naloziStampaTableFields       []domain.Fields
	naloziStavkeTableFields       []domain.Fields
	naloziStavkeStampaTableFields []domain.Fields
	cfg                           config.Config
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
	fvrRepo repository.BaseRepository[domain.Fvr],
	fproRepo repository.BaseRepository[domain.Fpro],
	nalogIDFieldName string,

	cfg config.Config,
) *NalogResource {
	rs := &NalogResource{
		service:          service,
		fproService:      fproService,
		fnalRepo:         fnalRepo,
		tipdokRepo:       tipdokRepo,
		sfRepo:           sfRepo,
		ojRepo:           ojRepo,
		mtroskaRepo:      mtroskaRepo,
		nalogIDFieldName: nalogIDFieldName,
		validator:        validator,
		fvrRepo:          fvrRepo,
		fproRepo:         fproRepo,
		cfg:              cfg,
	}
	rs.setServiceFieldValues()
	return rs
}

func (s *NalogResource) SetNalogIDFieldName(nalogIDFieldName string) {
	s.nalogIDFieldName = nalogIDFieldName
}

func (s *NalogResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}
func (s *NalogResource) checkGodinaZatvorena(ctx context.Context) (bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return true, fmt.Errorf("user session not found")
	}
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()

	qb := common.NewQueryBuilder(`SELECT godzatv FROM fvr `, true)
	if hasGod {
		qb.AddEqual("fvr.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fvr.kar", userSession.SelectedKar)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.fvrRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return true, fmt.Errorf("failed to get Fvr entities: %w", err)
	}
	if len(*entities) > 0 {
		return (*entities)[0].GodZatv, nil
	}
	return false, nil
}

// NalogValidation checks if the given Fnal entity is valid for creation or update.
func (s *NalogResource) NalogValidation(ctx context.Context, entity domain.Fnal, action string) ([]domain.FieldError, error) {
	checkGodinaZatvorena, err := s.checkGodinaZatvorena(ctx)
	if err != nil {
		return nil, err
	}
	if checkGodinaZatvorena {
		return []domain.FieldError{}, errors.New("godina je zatvorena, nije moguće izvršiti izmene na nalogu")
	}
	fieldErrors := []domain.FieldError{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fieldErrors, fmt.Errorf("user session not found")
	}
	if entity.Tipdok == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "tipdok",
			ErrorMessage: "obavezan podatak",
		})
	}
	if entity.Nalog == 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nalog",
			ErrorMessage: "obavezan podatak",
		})
	}
	if entity.Danal.Year() != userSession.SelectedGod {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "danal",
			ErrorMessage: fmt.Sprintf("godina naloga (%d) ne odgovara selektovanoj godini (%d)", entity.Danal.Year(), userSession.SelectedGod),
		})
	}
	if entity.Datob.Year() != userSession.SelectedGod {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "datob",
			ErrorMessage: fmt.Sprintf("godina naloga (%d) ne odgovara selektovanoj godini (%d)", entity.Datob.Year(), userSession.SelectedGod),
		})
	}
	if action == common.ActionAdd {
		exists, _, err := s.CheckNalogExistForCopy(ctx, entity)
		if err != nil {
			return fieldErrors, err
		}
		if exists {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:        "nalog",
				ErrorMessage: fmt.Sprintf("nalog sa brojem %d već postoji", entity.Nalog),
			})
		}
	}

	return fieldErrors, nil
}

// Create implements NalogService.
func (s *NalogResource) Create(ctx context.Context, fnal *domain.Fnal, idField string, fields []domain.Fields) (fieldsError []domain.FieldError, lastInsertedID int64, err error) {
	return s.service.Create(ctx, fnal, idField, fields)
}

// Delete implements NalogService.
func (s *NalogResource) Delete(ctx context.Context, idField string, id int64) error {
	return s.service.Delete(ctx, idField, id)
}

// GetAll implements NalogService.
func (s *NalogResource) GetAll(ctx context.Context, page int, offset int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]domain.Fnal, error) {
	return s.service.GetAll(ctx, page, offset, tableFields, idField, searchParams...)
}

// GetAllCustom implements NalogService.
func (s *NalogResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fnal, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements NalogService.
func (s *NalogResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Fnal, error) {
	return s.service.GetByID(ctx, idField, idValue)
}

// GetTotalRecords implements NalogService.
func (s *NalogResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchParams ...string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchParams...)
}

// GetTotalRecordsCustom implements NalogService.
func (s *NalogResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements NalogService.
func (s *NalogResource) MapEntityToValues(entity *domain.Fnal, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

// Update implements NalogService.
func (s *NalogResource) Update(ctx context.Context, entity *domain.Fnal, idField string, idValue interface{}, tableFields []domain.Fields) ([]domain.FieldError, error) {
	return s.service.Update(ctx, entity, idField, idValue, tableFields)
}

// CreateNalog implements NalogService.
func (s *NalogResource) CreateNalog(ctx context.Context, fnal *domain.Fnal, idField string, fields []domain.Fields, currentPage, pageSize int, searchText string) (frpoPaylad *domain.FproPayload, fieldsError []domain.FieldError, lastInsertedID int64, err error) {
	// This method should only handle business logic, not HTTP binding
	// HTTP binding should be done in the handler
	var fproPayload domain.FproPayload

	fieldErrors, err := s.validator.Validate(fnal)
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	if len(fieldErrors) > 0 {
		return &fproPayload, fieldErrors, 0, nil
	}

	// map request to entity
	id, err := s.GetIdTipdokByTipdok(ctx, fnal.Tipdok)
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	fnal.IDTipdok = id
	log.Println("Creating Nalog with Tipdok ID:", id, " Tipdok:", fnal.Tipdok, " Nalog:", fnal.Nalog, " Danal:", fnal.Danal.Format("2006-01-02"))
	fieldErrors, lastInsertedID, err = s.service.Create(ctx, fnal, common.IDfnal, s.MapEntityToValues(fnal, s.GetNaloziTableFields()))
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}
	if len(fieldErrors) > 0 {
		return &fproPayload, fieldErrors, lastInsertedID, nil
	}
	s.MapToFproPayload(&fproPayload, *fnal, lastInsertedID)

	tblStavke := common.SetTableBasicData("Stavke Naloga", naloziStavkeTableID, s.fproService.GetTableStavkeFields(), "", "", pageSize, 0, 0, 0, s.cfg)
	tblStavke.ShowActions = true
	err = s.fproService.GetAllFproByFnalID(ctx, &fproPayload, &tblStavke, lastInsertedID, currentPage, pageSize, searchText)
	if err != nil {
		return &fproPayload, []domain.FieldError{}, 0, err
	}

	return &fproPayload, []domain.FieldError{}, lastInsertedID, nil
}

// UpdateNalog implements NalogService.
func (s *NalogResource) UpdateNalog(ctx context.Context, fnalID int64, payload *domain.FnalPayload, tblStavke *domain.TableData, currentPage, pageSize int, searchText string) (fproPayload domain.FproPayload, fieldErrors []domain.FieldError, err error) {
	var entity *domain.Fnal
	*tblStavke = common.SetTableBasicData("Stavke Naloga", naloziStavkeTableID, s.fproService.GetTableStavkeFields(), "", "", pageSize, 0, 0, 0, s.cfg)
	tblStavke.ShowActions = true
	// Get username from claims or session
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fproPayload, []domain.FieldError{}, fmt.Errorf("user session not found")
	}
	entity, err = s.service.GetByID(ctx, common.IDfnal, fnalID)
	if err != nil {
		return fproPayload, []domain.FieldError{}, err
	}
	// map request to entity
	s.MapReqToEntity(ctx, *payload, entity, ActionUpdate)

	fieldErrors, err = s.Update(ctx, entity, common.IDfnal, fnalID, s.MapEntityToValues(entity, s.GetNaloziTableFields()))
	if err != nil {
		return fproPayload, []domain.FieldError{}, err
	}
	if len(fieldErrors) > 0 {
		return fproPayload, []domain.FieldError{}, fmt.Errorf("%s", common.ErrMsgValidation)
	}

	s.MapToFproPayload(&fproPayload, *entity, fnalID)

	err = s.fproService.GetAllFproByFnalID(ctx, &fproPayload, tblStavke, fnalID, currentPage, pageSize, searchText)
	if err != nil {
		return fproPayload, []domain.FieldError{}, err
	}

	return fproPayload, []domain.FieldError{}, nil
}

// Update implements NalogService.
func (s *NalogResource) GetNextNalog(ctx context.Context, tipdok string) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(` SELECT COALESCE(MAX(nalog), 0) + 1 as nalog FROM fnal`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	sqlQuery, args := qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return 0, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) == 0 {
		return 1, nil
	}

	return (*entities)[0].Nalog, nil

}

// GetByTipdokNalog retrieves a Fnal entity based on tipdok and nalog values.
func (s *NalogResource) GetByTipdokNalog(ctx context.Context, tipdok string, nalog int64) (domain.Fnal, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.Fnal{}, fmt.Errorf("user session not found")
	}

	entity := domain.Fnal{}
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(` SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	qb.AddEqual("nalog", nalog)
	sqlQuery, args := qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return entity, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) > 0 {
		entity = (*entities)[0]
	}
	return entity, nil
}

// GetIdTipdokByTipdok retrieves the ID of a Tipdok based on its name.
func (s *NalogResource) GetIdTipdokByTipdok(ctx context.Context, tipdok string) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("user session not found")
	}

	entity := domain.Tipdok{}
	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(` SELECT idtipdok, tipdok, opis FROM tipdok`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	sqlQuery, args := qb.Build()
	tipdokEntities, err := s.tipdokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
func (s *NalogResource) GetNalogViewData(ctx context.Context, tbl *domain.TableData, searchQuery string, currentPage, pageSize int, isInitialRequest bool, sortBy, sortOrder, tipdok string) (NalogViewData, error) {
	// Initialize viewData
	var viewData NalogViewData

	// Get session for god/kar values at the start
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return viewData, fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	// 1. Fetch Tipdok Options on initial load or if selectedTipdok is empty
	// to determine the default.
	if isInitialRequest || tipdok == "" {
		tipdokOptions, err := s.GetTipdokOptions(ctx)
		if err != nil {
			return viewData, fmt.Errorf("failed to get tipdok options: %w", err)
		}
		viewData.TipdokOptions = &tipdokOptions
		for _, td := range tipdokOptions {
			viewData.TipdokComboItems = append(viewData.TipdokComboItems, domain.ComboItem{Key: td.TipDok, Value: td.TipDok + " - " + td.Opis})
		}

		if tipdok == "" && len(tipdokOptions) > 0 {
			viewData.DefaultTipdok = tipdokOptions[0].TipDok
			tipdok = tipdokOptions[0].TipDok
		} else {
			viewData.DefaultTipdok = tipdok // Use the provided one if available
		}
	} else {
		viewData.DefaultTipdok = tipdok // Use the provided one directly
	}
	noviNalogBr, _ := s.GetNextNalog(ctx, viewData.DefaultTipdok)
	viewData.NextNalog = fmt.Sprintf("%d", noviNalogBr)

	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	// Get total records
	qbCount := common.NewQueryBuilder(` SELECT count(*) FROM fnal`, true)
	qbCount.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qbCount.AddEqual("tipdok", viewData.DefaultTipdok)
	if searchQuery != "" {
		qbCount.AddLike("CONCAT(tipdok, ' ', nalog, ' ', opis, ' ', dug, ' ', pot, ' ', datob, ' ', brst, ' ', nalsts)", searchQuery)
	}

	sqlQuery, args := qbCount.Build()
	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}

	// Calculate pagination details
	common.SetTableTotalRecords(tbl, totRecords, pageSize)

	qb := common.NewQueryBuilder(` SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddEqual("tipdok", tipdok)
	if searchQuery != "" {
		qb.AddLike("CONCAT(tipdok, ' ', nalog, ' ', opis, ' ', dug, ' ', pot, ' ', datob, ' ', brst, ' ', nalsts)", searchQuery)
	}
	if sortBy != "" {
		qb.AddOrderBy(fmt.Sprintf("fnal.%s", sortBy))
	}
	if sortOrder != "" && (sortOrder == "ASC" || sortOrder == "DESC") {
		qb.AddSortOrder(sortOrder)
	}
	qb.SetLimit(pageSize)
	qb.SetOffset((currentPage - 1) * pageSize)
	sqlQuery, args = qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	viewData.FnalEntities = *entities

	// Additional table data configuration can happen here
	tbl.URLGetAll = "/api/nalozi/all/tipdok" // For tipdok specific updates
	tbl.Pagination.HxInclude = "#tipdok, #search-input"
	tbl.HxInclude = "#tipdok, #search-input"
	tbl.ShowActions = false // Default, can be overridden by specific handler needs
	tbl.BtnAdd.IsVisible = false
	tbl.BtnUpdate.IsVisible = false
	tbl.BtnDelete.IsVisible = false
	tbl.BtnPrint.IsVisible = false

	viewData.TableData = *tbl

	// 3. Fetch UkupnaObrada totals if it's an initial load
	if isInitialRequest {
		ukObrada, err := s.GetNalogTotals(ctx, tipdok)
		if err != nil {
			return viewData, fmt.Errorf("failed to get Nalog totals: %w", err)
		}
		viewData.UkupnaObrada = &ukObrada
	}

	return viewData, nil
}

// GetNalogPrepisData fetches all data required to render the Nalog list page.
func (s *NalogResource) GetNalogPrepisData(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, searchText, sortBy, sortOrder string) error {

	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()

	qbCount := common.NewQueryBuilder(`SELECT count(*) FROM fnal`, true)
	if hasGod {
		qbCount.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qbCount.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" {
		qbCount.AddLike("CONCAT(tipdok, ' ', nalog, ' ', opis, ' ', dug, ' ', pot, ' ', datob, ' ', brst, ' ', nalsts)", searchText)
	}
	sqlQuery, args := qbCount.Build()
	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get total records for Fnal: %w", err)
	}
	common.SetTableTotalRecords(tbl, totRecords, pageSize)

	qb := common.NewQueryBuilder(`SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts, idtipdok FROM fnal`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" {
		qb.AddLike("CONCAT(tipdok, ' ', nalog, ' ', opis, ' ', dug, ' ', pot, ' ', datob, ' ', brst, ' ', nalsts)", searchText)
	}
	if sortBy != "" {
		qb.AddOrderBy(fmt.Sprintf("fnal.%s", sortBy))
	}
	if sortOrder != "" && (sortOrder == "ASC" || sortOrder == "DESC") {
		qb.AddSortOrder(sortOrder)
	}
	qb.SetLimit(pageSize)
	qb.SetOffset((currentPage - 1) * pageSize)

	sqlQuery, args = qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) == 0 {
		return fmt.Errorf("no Fnal entities found")
	}
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		rbr := 1 + (currentPage-1)*pageSize // Calculate RBR based on current page and page size
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				fmt.Sprintf("%d", rbr),
				entity.Tipdok,
				fmt.Sprintf("%d", entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				entity.Datob.Format(common.DateLayout),
				fmt.Sprintf("%d", entity.Brst),
				entity.Nalsts,
				fmt.Sprintf("%d", entity.IDTipdok),
			)
			rbr++
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFnal), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	//get tipdok options for validation
	_, err = s.GetTipdokOptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tipdok options: %w", err)
	}
	// Configure table
	// Note: viewData related properties could be added to TableData struct if needed
	// Additional table data configuration can happen here
	tbl.URLGetAll = "/api/nalozi/prepis"
	tbl.Pagination.HxInclude = "#search-input"
	tbl.HxInclude = "#search-input"
	tbl.DetailTarget = "#nalozi_kopiranje_stavke"
	tbl.ShowActions = true // Default, can be overridden by specific handler needs
	tbl.BtnAdd.IsVisible = false
	tbl.BtnUpdate.IsVisible = true
	tbl.BtnDelete.IsVisible = false
	tbl.BtnPrint.IsVisible = false
	tbl.BtnUpdate.LabelText = "Kopiraj"
	tbl.BtnUpdate.HxActionURL = "/api/nalozi/confirm-prepis" // Set the HxActionURL for the update button

	return nil
}

// KopirajNalog handles copying a nalog to a new nalog (create scenario).
func (s *NalogResource) KopirajNalog(ctx context.Context, idFnal int64, entity domain.Fnal) error {
	// Get source nalog

	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	username := userSession.UserName

	var copyFnExists bool
	err := s.fnalRepo.DB.QueryRowContext(ctx,
		`SELECT to_regprocedure('fn_kopiraj_nalog_sa_stavkama(bigint,bigint,text,date,date,text,integer,integer,integer,text)') IS NOT NULL`,
	).Scan(&copyFnExists)
	if err != nil {
		return fmt.Errorf("failed to verify copy function existence: %w", err)
	}
	if !copyFnExists {
		return fmt.Errorf("required database function fn_kopiraj_nalog_sa_stavkama is missing")
	}
	// get the idtipdok for the provided tipdok in the request body to pass to the copy function
	var newFnalID, newIdTipdok int64
	hasgod, haskar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idtipdok FROM tipdok `, true)
	if hasgod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if haskar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("tipdok", entity.Tipdok)
	sqlQueryTipdok, argsTipdok := qb.Build()
	entites, err := s.tipdokRepo.GetAllCustom(ctx, sqlQueryTipdok, "", argsTipdok, "", "")
	if err != nil {
		return fmt.Errorf("failed to get tipdok ID: %w", err)
	}
	if len(*entites) == 0 {
		return fmt.Errorf("tipdok not found: %s", entity.Tipdok)
	}
	newIdTipdok = int64((*entites)[0].IDTipDok)

	err = s.fnalRepo.DB.QueryRowContext(ctx,
		`SELECT fn_kopiraj_nalog_sa_stavkama($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		idFnal,
		entity.Nalog,
		entity.Tipdok,
		entity.Danal.Format(common.HtmlLayout),
		entity.Datob.Format(common.HtmlLayout),
		entity.Opis,
		newIdTipdok,
		userSession.SelectedGod,
		userSession.SelectedKar,
		username,
	).Scan(&newFnalID)
	if err != nil {
		return fmt.Errorf("Greska prilikom kopiranja naloga preko DB funkcije: %w", err)
	}

	return nil
}

// KopirajNalogToExisting handles copying a nalog to an existing nalog (update scenario).
func (s *NalogResource) KopirajNalogToExisting(ctx context.Context, req domain.FnalPayload, target_IdFnal int64) error {
	// Implementation for copying to existing nalog
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}
	username := userSession.UserName
	// Username is retrieved from user session in context

	var copyFnExists bool
	err := s.fnalRepo.DB.QueryRowContext(ctx,
		`SELECT to_regprocedure('fn_kopiraj_stavke_u_postojeci_nalog(bigint,bigint,text)') IS NOT NULL`,
	).Scan(&copyFnExists)
	if err != nil {
		return fmt.Errorf("failed to verify copy function existence: %w", err)
	}
	if !copyFnExists {
		return fmt.Errorf("required database function fn_kopiraj_stavke_u_postojeci_nalog is missing")
	}

	var source_IdFnal int64
	source_IdFnal = common.StringToInt64(req.IDFnal)
	if source_IdFnal == 0 {
		return fmt.Errorf("invalid source idfnal: %s", req.IDFnal)
	}
	var fnResult sql.NullString
	err = s.fnalRepo.DB.QueryRowContext(ctx,
		`SELECT fn_kopiraj_stavke_u_postojeci_nalog($1,$2,$3)`,
		source_IdFnal,
		target_IdFnal,
		username,
	).Scan(&fnResult)
	if err != nil {
		return fmt.Errorf("Greska prilikom kopiranja naloga preko DB funkcije: %w", err)
	}

	return nil
}

// GetNalogStornirajData fetches all data required to render the Nalog list page.
func (s *NalogResource) GetNalogStornirajData(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, searchText, sortBy, sortOrder string) (NalogViewData, error) {
	viewData := NalogViewData{
		DefaultTipdok: "",
	}
	// Get session for god/kar values
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return viewData, fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, currentPage, pageSize)
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qbCount := common.NewQueryBuilder(`SELECT count(*) FROM fnal`, true)
	if hasGod {
		qbCount.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qbCount.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" {
		qbCount.AddLike("CONCAT(tipdok, ' ', nalog, ' ', opis, ' ', dug, ' ', pot, ' ', datob, ' ', brst, ' ', nalsts)", searchText)
	}

	sqlQuery, args := qbCount.Build()
	totRecords, err := s.fnalRepo.GetTotalRecordsCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get total records for Fnal: %w", err)
	}
	common.SetTableTotalRecords(tbl, totRecords, pageSize)
	qb := common.NewQueryBuilder(`SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts, idtipdok FROM fnal`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" {
		qb.AddLike("CONCAT(tipdok, ' ', nalog, ' ', opis, ' ', dug, ' ', pot, ' ', datob, ' ', brst, ' ', nalsts)", searchText)
	}
	if sortBy != "" {
		qb.AddOrderBy(fmt.Sprintf("fnal.%s", sortBy))
	}
	if sortOrder != "" && (sortOrder == "ASC" || sortOrder == "DESC") {
		qb.AddSortOrder(sortOrder)
	}
	qb.SetLimit(pageSize)
	qb.SetOffset((currentPage - 1) * pageSize)

	sqlQuery, args = qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return viewData, fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if len(*entities) == 0 {
		return viewData, fmt.Errorf("no Fnal entities found")
	}
	viewData.FnalEntities = *entities
	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		rbr := 1 + (currentPage-1)*pageSize // Calculate RBR based on current page and page size
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				fmt.Sprintf("%d", rbr),
				entity.Tipdok,
				fmt.Sprintf("%d", entity.Nalog),
				entity.Danal.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				entity.Datob.Format(common.DateLayout),
				fmt.Sprintf("%d", entity.Brst),
				entity.Nalsts,
				fmt.Sprintf("%d", entity.IDTipdok),
			)
			rbr++
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFnal), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}
	btnStorniraj := domain.Button{
		LabelText:     "Storniraj",
		HxActionURL:   "/api/nalozi/confirm-storniraj",
		IsVisible:     true,
		Id:            "btn-storniraj",
		HxRequestType: "POST",
	}
	// Additional table data configuration can happen here
	tbl.URLGetAll = "/api/nalozi/storniraj"
	tbl.Pagination.HxInclude = "#search-input"
	tbl.HxInclude = "#search-input"
	tbl.DetailTarget = "#nalozi_storniraj_stavke"
	tbl.BtnUpdate = btnStorniraj
	tbl.BtnDelete.LabelText = "Obriši"
	tbl.ShowActions = true // Default, can be overridden by specific handler needs
	tbl.BtnUpdate.IsVisible = true
	tbl.BtnDelete.IsVisible = true
	tbl.BtnUpdate.HxActionURL = "/api/nalozi/confirm-storniraj" // Set the HxActionURL for the update button

	viewData.TableData = *tbl

	return viewData, nil
}

// StornirajNalog handles copying a nalog to a new nalog with storniraj flag (create scenario).
func (s *NalogResource) StornirajNalog(ctx context.Context, idFnal int64, entity domain.Fnal) error {
	// Get source nalog

	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	username := userSession.UserName

	var copyFnExists bool
	err := s.fnalRepo.DB.QueryRowContext(ctx,
		`SELECT to_regprocedure('fn_storniraj_nalog_sa_stavkama(bigint,bigint,text,date,date,text,integer,integer,text)') IS NOT NULL`,
	).Scan(&copyFnExists)
	if err != nil {
		return fmt.Errorf("failed to verify storniraj function existence: %w", err)
	}
	if !copyFnExists {
		return fmt.Errorf("required database function fn_storniraj_nalog_sa_stavkama is missing")
	}
	var newFnalID int64
	err = s.fnalRepo.DB.QueryRowContext(ctx,
		`SELECT fn_storniraj_nalog_sa_stavkama($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		idFnal,
		entity.Nalog,
		entity.Tipdok,
		entity.Danal.Format(common.HtmlLayout),
		entity.Datob.Format(common.HtmlLayout),
		entity.Opis,
		userSession.SelectedGod,
		userSession.SelectedKar,
		username,
	).Scan(&newFnalID)
	if err != nil {
		return fmt.Errorf("Greska prilikom storniranja naloga preko DB funkcije: %w", err)
	}

	return nil
}

func (s *NalogResource) GetNalogStampanjeData(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, page, pageSize int, searchText, sortBy, sortOrder string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, page, pageSize)

	qb := common.NewQueryBuilder(`SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts, idtipdok FROM fnal`, true)
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" {
		qb.AddLike("CONCAT(tipdok, ' ', nalog, ' ', opis, ' ', iznos, ' ', datob, ' ', brst, ' ', nalsts, ' ', idtipdok)", searchText)
	}
	if !getTotalRecords {
		if sortBy != "" {
			qb.AddOrderBy(fmt.Sprintf("fnal.%s", sortBy))
		}
		if sortOrder != "" && (sortOrder == "ASC" || sortOrder == "DESC") {
			qb.AddSortOrder(sortOrder)
		}
		qb.SetLimit(pageSize)
		qb.SetOffset((page - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get Fnal entities: %w", err)
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil // Return early if we only needed to get total records
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		rbr := 1 + (page-1)*pageSize // Calculate RBR based on current page and page size
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				entity.Tipdok,
				fmt.Sprintf("%d", entity.Nalog),
				entity.Datob.Format(common.DateLayout),
				entity.Danal.Format(common.DateLayout),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				entity.Nalsts,
				fmt.Sprintf("%d", entity.Brst),
			)
			rbr++
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFnal), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}
func (s *NalogResource) GetNalogStampanjeDataDetalji(ctx context.Context, tbl *domain.TableData, getTotalRecords bool, page, pageSize int, idFnal int64, searchText, sortBy, sortOrder string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("user session not found")
	}

	common.SetupTablePagination(tbl, page, pageSize)

	qb := common.NewQueryBuilder(`SELECT fpro.rbr, fpro.konto, fpro.sifra, fkpl.naziv, fpro.vrd, fpro.opis, 
	case when fpro.kat in (1, 2) then fpro.iznos else 0 end as dug,
	case when fpro.kat in (3, 4) then fpro.iznos else 0 end as pot,
	fpro.dokum, fpro.dadok, fpro.vkonta FROM fpro`, true)
	qb.AddJoin("left join  fkpl ON fpro.idfkpl = fkpl.idfkpl")

	qb.AddEqual("idfnal", idFnal)
	if searchText != "" {
		qb.AddLike("CONCAT(fpro.rbr, ' ', fpro.konto, ' ', fpro.sifra, ' ', fkpl.naziv, ' ', fpro.vrd, ' ', fpro.opis, ' ', fpro.iznos, ' ', fpro.dokum, ' ', fpro.dadok, ' ', fpro.vkonta)", searchText)
	}
	if !getTotalRecords {
		if sortBy != "" {
			qb.AddOrderBy(fmt.Sprintf("%s", sortBy))
		}
		if sortOrder != "" && (sortOrder == "ASC" || sortOrder == "DESC") {
			qb.AddSortOrder(sortOrder)
		}
		qb.SetLimit(pageSize)
		qb.SetOffset((page - 1) * pageSize)
	}
	sqlQuery, args := qb.Build()
	entities, err := s.fproRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("failed to get Fpro entities: %w", err)
	}
	if getTotalRecords {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil // Return early if we only needed to get total records
	}

	// Populate table rows
	if entities != nil && len(*entities) > 0 {
		rbr := 1 + (page-1)*pageSize // Calculate RBR based on current page and page size
		for _, entity := range *entities {
			fields := []string{}
			// Add common fields
			fields = append(fields,
				fmt.Sprintf("%d", rbr),
				entity.Konto,
				entity.Sifra,
				entity.Naziv,
				fmt.Sprintf("%d", entity.Vrd),
				entity.Opis,
				common.FormatNumberWithSystemLocale(entity.Dug, 2),
				common.FormatNumberWithSystemLocale(entity.Pot, 2),
				entity.Dokum,
				common.FormatNullTime(entity.Dadok, common.DateLayout),
				fmt.Sprintf("%d", entity.Vkonta),
			)
			rbr++
			tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFpro), Fields: fields, HasUpdate: false, HasDelete: false}
			tbl.Rows = append(tbl.Rows, tblRow)
		}
	}

	return nil
}

// GetNalogTotals calculates the total dug, pot, brst, and brna for the current user's selected god and kar.
func (s *NalogResource) GetNalogTotals(ctx context.Context, tipdok string) (domain.UkupnaObrada, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.UkupnaObrada{}, fmt.Errorf("user session not found")
	}

	qb := common.NewQueryBuilder(`SELECT COALESCE(SUM(dug), 0) as dug, 
				COALESCE(SUM(pot), 0) as pot,
				COALESCE(SUM(brst), 0) as brst, 
				COALESCE(count(*), 0) as brna FROM fnal`, true)
	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	// if tipdok != "" {
	// 	qb.AddEqual("tipdok", tipdok)
	// }
	sqlQuery, args := qb.Build()
	totalsFnal, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "") // Assuming a GetCustom method for single row
	if err != nil {
		return domain.UkupnaObrada{}, fmt.Errorf("failed to get Fnal totals: %w", err)
	}

	if totalsFnal != nil && len(*totalsFnal) > 0 {
		ukObrada := domain.UkupnaObrada{
			Duguje:    common.FormatNumberWithSystemLocale((*totalsFnal)[0].Dug, 2),
			Potrazuje: common.FormatNumberWithSystemLocale((*totalsFnal)[0].Pot, 2),
			UkStavki:  common.FormatNumberWithSystemLocale((*totalsFnal)[0].Brst, 0),
			UkNaloga:  common.FormatNumberWithSystemLocale((*totalsFnal)[0].Brna, 0),
		}
		return ukObrada, nil
	}
	return domain.UkupnaObrada{}, nil
}

// GetTipdokOptions fetches the list of tipdok options for filtering. This method stays the same.
func (s *NalogResource) GetTipdokOptions(ctx context.Context) ([]domain.Tipdok, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.tipdokRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idtipdok, tipdok, opis FROM tipdok`, true)
	qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	qb.AddCustomCondition("(grpdok = 'FIN' OR grpdok = 'SVI')")
	qb.AddOrderBy("tipdok::NUMERIC ASC")
	sqlQuery, args := qb.Build()
	tipdokValues, err := s.tipdokRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tipdok options: %w", err)
	}
	return *tipdokValues, nil
}

// GetNaloziTableFields returns the field definitions for the nalozi table.
func (s *NalogResource) GetNaloziTableFields() []domain.Fields {
	if len(s.naloziTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziTableFields
}
func (s *NalogResource) GetNaloziStampaTableFields() []domain.Fields {
	if len(s.naloziStampaTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziStampaTableFields
}
func (s *NalogResource) GetNaloziStavkeStampaTableFields() []domain.Fields {
	if len(s.naloziStavkeStampaTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziStavkeStampaTableFields
}
func (s *NalogResource) GetNaloziStavkeTableFields() []domain.Fields {
	if len(s.naloziStavkeTableFields) == 0 {
		s.setServiceFieldValues()
	}
	return s.naloziStavkeTableFields
}

// ValidateCopyNalog validates the data for copying a nalog (voucher)
func (s *NalogResource) ValidateCopyNalog(ctx context.Context, req domain.FnalPayload, entity domain.Fnal) ([]domain.FieldError, error) {
	fieldErrors := []domain.FieldError{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fieldErrors, fmt.Errorf("user session not found")
	}
	checkGodinaZatvorena, err := s.checkGodinaZatvorena(ctx)
	if err != nil {
		return nil, err
	}
	if checkGodinaZatvorena {
		return []domain.FieldError{}, errors.New("godina je zatvorena, nije moguće izvršiti izmene na nalogu")
	}

	nalogExists, _, err := s.CheckNalogExistForCopy(ctx, entity)
	if err != nil {
		return fieldErrors, fmt.Errorf("error checking existing nalog: %w", err)
	}
	if nalogExists {
		return fieldErrors, fmt.Errorf("nalog sa tipom '%s' i brojem '%d' već postoji, nije dozvoljeno kopiranje sa istim brojem i tipom naloga", entity.Tipdok, entity.Nalog)
	}
	if entity.Nalog == 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nalog",
			ErrorMessage: "obavezan podatak",
		})
	}
	if entity.Danal.Year() != userSession.SelectedGod {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "danal",
			ErrorMessage: fmt.Sprintf("godina naloga (%d) ne odgovara selektovanoj godini (%d)", entity.Danal.Year(), userSession.SelectedGod),
		})
	}
	if entity.Datob.Year() != userSession.SelectedGod {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "datob",
			ErrorMessage: fmt.Sprintf("godina naloga (%d) ne odgovara selektovanoj godini (%d)", entity.Datob.Year(), userSession.SelectedGod),
		})
	}
	if req.Tipdok == req.OldTipdok && req.Nalog == req.OldNalog {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:        "nalog",
			ErrorMessage: "broj naloga i tip naloga su isti kao kod izvornog naloga, nije dozvoljeno kopiranje sa istim brojem i tipom naloga",
		})
	}
	return fieldErrors, nil
}

// CheckNalogExistForCopy checks if a nalog with the same tipdok and nalog already exists to prevent duplicates during copy.
func (s *NalogResource) CheckNalogExistForCopy(ctx context.Context, entity domain.Fnal) (bool, int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return false, 0, fmt.Errorf("user session not found")
	}

	hasGod, hasKar := s.fnalRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder(`SELECT idfnal FROM fnal`, true)
	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	qb.AddEqual("tipdok", entity.Tipdok)
	qb.AddEqual("nalog", entity.Nalog)
	sqlQuery, args := qb.Build()
	entities, err := s.fnalRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return false, 0, err
	}
	if len(*entities) == 0 {
		return false, 0, nil
	}
	return len(*entities) > 0, (*entities)[0].IDFnal, nil

}

// setServiceFieldValues initializes the field definitions for nalozi and nalozi stavke tables.
func (s *NalogResource) setServiceFieldValues() {
	s.naloziTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni Broj", Width: "4", Field: "rbr", Sortable: true, TextAlign: "right"},
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "6", Field: "tipdok", Sortable: true},
		{Name: "nalog", Label: "Broj Naloga", Width: "12", Field: "nalog", Sortable: true},
		{Name: "danal", Label: "Datum Naloga", Width: "12", Field: "danal", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "60", Field: "opis", Sortable: true},
		{Name: "dug", Label: "Duguje", Width: "14", Field: "dug", Sortable: true, TextAlign: "right"},
		{Name: "pot", Label: "Potrazuje", Width: "14", Field: "pot", Sortable: true, TextAlign: "right"},
		{Name: "datob", Label: "Datum Obrade", Width: "12", Field: "datob", Sortable: true},
		{Name: "brst", Label: "Broj Stavki", Width: "5", Field: "brst", Sortable: true},
		{Name: "nalsts", Label: "Status Naloga", Width: "10", Field: "nalsts", Sortable: true},
		{Name: "idtipdok", Label: "Id tipdok", Width: "10", Field: "idtipdok"},
	}
	s.naloziStampaTableFields = []domain.Fields{
		{Name: "tipdok", Label: "Vrsta Naloga", Width: "6", Field: "tipdok", Sortable: true},
		{Name: "nalog", Label: "Broj Naloga", Width: "12", Field: "nalog", Sortable: true},
		{Name: "datob", Label: "Datum Obrade", Width: "12", Field: "datob", Sortable: true},
		{Name: "danal", Label: "Datum Naloga", Width: "12", Field: "danal", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "60", Field: "opis", Sortable: true},
		{Name: "dug", Label: "Duguje", Width: "14", Field: "dug", TextAlign: "right", Sortable: true},
		{Name: "pot", Label: "Potrazuje", Width: "14", Field: "pot", TextAlign: "right", Sortable: true},
		{Name: "nalsts", Label: "Status Naloga", Width: "10", Field: "nalsts", Sortable: true},
		{Name: "brst", Label: "Broj Stavki", Width: "5", Field: "brst", Sortable: true},
	}
	s.naloziStavkeStampaTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni Broj", Width: "4", Sortable: true},
		{Name: "konto", Label: "Konto", Width: "6", Sortable: true},
		{Name: "sifra", Label: "Sifra", Width: "6", Sortable: true},
		{Name: "naziv", Label: "Naziv Konta", Width: "60", Sortable: true},
		{Name: "vrd", Label: "Vrsta Dok.", Width: "10", Sortable: true},
		{Name: "opis", Label: "Opis", Width: "60", Sortable: true},
		{Name: "dug", Label: "Duguje", Width: "14", Sortable: true, TextAlign: "right"},
		{Name: "pot", Label: "Potrazuje", Width: "14", Sortable: true, TextAlign: "right"},
		{Name: "dokum", Label: "Broj Dokum", Width: "12", Sortable: true},
		{Name: "dadok", Label: "Datum Dokumenta", Width: "12", Sortable: true},
		{Name: "vkonta", Label: "Vrsta Konta", Width: "14", Sortable: true},
	}
	s.naloziStavkeTableFields = []domain.Fields{
		{Name: "rbr", Label: "Redni Broj", Width: "4"},
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

// MapReqToEntity maps the incoming request data to the Fnal entity for both add and update operations.
func (s *NalogResource) MapReqToEntity(ctx context.Context, req domain.FnalPayload, entity *domain.Fnal, action string) {
	// Get user session from context
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		// This should not happen as middleware ensures session exists
		return
	}

	if strings.ToLower(action) == "add" {
		entity.God = userSession.SelectedGod
		entity.Kar = userSession.SelectedKar
		entity.Xdatunosa = sql.NullTime{Time: time.Now(), Valid: true}
		entity.Xopunos = sql.NullString{String: userSession.UserName, Valid: true}
	}
	if strings.ToLower(action) == "update" {
		entity.Xdatizmene = sql.NullTime{Time: time.Now(), Valid: true}
		entity.Xopizmene = sql.NullString{String: userSession.UserName, Valid: true}
	}
	entity.Nalog = common.StringToInt64(req.Nalog)
	entity.Danal = common.StringToDate(req.Danal)
	entity.Datob = common.StringToDate(req.Datob)
	entity.Tipdok = req.Tipdok

}

// MapToFproPayload maps the Fnal entity data to the FproPayload structure for printing or external processing.
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
