package robno

import (
	"context"
	"errors"
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
)

// MagacinKontoService defines the interface for operations related to Magkonto.
type MagacinKontoService interface {
	service.Service[domain.Magkonto]
	GetAllMagacinKonto(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, typePrint string) error
	GetMagacinKontoTableFields() []domain.Fields
	ValidateEntity(ctx context.Context, entity *domain.Magkonto) []domain.FieldError
	GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Magkonto, error)
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// MagacinKontoResource implements the MagacinKontoService interface.
type MagacinKontoResource struct {
	service                 *service.BaseService[domain.Magkonto]
	magacinKontoRepo        *repository.BaseRepository[domain.Magkonto]
	fvrRepo                 *repository.BaseRepository[domain.Fvr]
	magacinKontoTableFields []domain.Fields
	cfg                     config.Config
}

func NewMagacinKontoResource(service *service.BaseService[domain.Magkonto], magacinKontoRepo *repository.BaseRepository[domain.Magkonto], fvrRepo *repository.BaseRepository[domain.Fvr], cfg config.Config) *MagacinKontoResource {
	rs := &MagacinKontoResource{
		service:          service,
		magacinKontoRepo: magacinKontoRepo,
		fvrRepo:          fvrRepo,
		cfg:              cfg,
	}
	rs.setMagacinKontoTableFields()
	return rs
}

func (s *MagacinKontoResource) setMagacinKontoTableFields() {
	s.magacinKontoTableFields = []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "8"},
		{Name: "konto", Label: "Konto zaliha", Width: "15"},
		{Name: "opis_konta", Label: "Opis konta zalihe", Width: "25"},
		{Name: "opis_mag", Label: "Opis magacina", Width: "25"},
		{Name: "kontoprih", Label: "Konto prihoda", Width: "15"},
		{Name: "kontotroska", Label: "Konto troška", Width: "15"},
		{Name: "kontoruc", Label: "Konto RUC", Width: "15"},
	}
}

func (s *MagacinKontoResource) GetMagacinKontoTableFields() []domain.Fields {
	return s.magacinKontoTableFields
}

func (s *MagacinKontoResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}

func (s *MagacinKontoResource) Create(ctx context.Context, entity *domain.Magkonto, idField string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, 0, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "magkonto",
		EntityType:    reflect.TypeOf(entity).Elem(),
		TableFields:   s.GetMagacinKontoTableFields(),
		IncludeGodKar: true,
	}

	qb := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.BuildInsert(ctx, fields, idField)

	tx, err := s.magacinKontoRepo.BeginTx()
	if err != nil {
		return nil, 0, fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(ctx, sqlQuery, args...).Err()
	if err != nil {
		return nil, 0, fmt.Errorf("insert magkonto failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, 0, nil
}

func (s *MagacinKontoResource) Update(ctx context.Context, entity *domain.Magkonto, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "magkonto",
		EntityType:    reflect.TypeOf(entity).Elem(),
		TableFields:   fields,
		IncludeGodKar: true,
	}

	qb := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.BuildUpdate(ctx, fields, idField, idFieldValue)

	tx, err := s.magacinKontoRepo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(ctx, sqlQuery, args...).Err()
	if err != nil {
		return nil, fmt.Errorf("update magkonto failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, nil
}

func (s *MagacinKontoResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *MagacinKontoResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Magkonto, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *MagacinKontoResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Magkonto, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *MagacinKontoResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Magkonto, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *MagacinKontoResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *MagacinKontoResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *MagacinKontoResource) MapEntityToValues(entity *domain.Magkonto, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *MagacinKontoResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

func (s *MagacinKontoResource) GetAllMagacinKonto(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	hasGod, hasKar := s.magacinKontoRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("SELECT magaciniid, mag, god, kar, xdatunosa, xdatizmene, xopunos, xopizmene, konto, idfkpl, vkonta, kontoprih, kontotroska, kontoruc, kontorab FROM magkonto", true)

	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" && printType == common.TipStampePreview {
		qb.AddCustomSearchCondition([]string{"konto", "kontoprih", "kontotroska"}, searchText)
	}
	if !getTotalRecords {
		if sortBy != "" {
			qb.AddOrderBy(fmt.Sprintf("magkonto.%s", sortBy))
		}
		if sortOrder != "" && (sortOrder == "ASC" || sortOrder == "DESC") {
			qb.AddSortOrder(sortOrder)
		}
		if printType == common.TipStampePreview {
			qb.SetLimit(pageSize)
			qb.SetOffset((currentPage - 1) * pageSize)
		}
	}

	sqlQuery, args := qb.Build()
	entities, err := s.magacinKontoRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	for _, entity := range *entities {
		fields := []string{
			fmt.Sprintf("%d", entity.Mag),
			entity.Konto,
			"", // opis_konta - placeholder
			"", // opis_mag - placeholder
			entity.KontoPrih,
			entity.KontoTroska,
			entity.KontoRuc,
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.MagaciniID), Fields: fields, HasUpdate: true, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

func (s *MagacinKontoResource) ValidateEntity(ctx context.Context, entity *domain.Magkonto) []domain.FieldError {
	var errors []domain.FieldError

	if entity.Konto == "" {
		errors = append(errors, domain.FieldError{
			Field:        "konto",
			ErrorMessage: "Konto zaliha je obavezan",
		})
	}

	return errors
}

func (s *MagacinKontoResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return common.GetFvrData(ctx, s.fvrRepo)
}
