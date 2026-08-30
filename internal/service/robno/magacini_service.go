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

// MagaciniService defines the interface for operations related to Magacini.
type MagaciniService interface {
	service.Service[domain.Magacini]
	GetAllMagacini(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, tipStampe string) error
	GetMagaciniTableFields() []domain.Fields
	ValidateEntity(ctx context.Context, entity *domain.Magacini) []domain.FieldError
	GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Magacini, error)
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// MagaciniResource implements the MagaciniService interface.
type MagaciniResource struct {
	service             *service.BaseService[domain.Magacini]
	magaciniRepo        *repository.BaseRepository[domain.Magacini]
	fvrRepo             *repository.BaseRepository[domain.Fvr]
	magaciniTableFields []domain.Fields
	cfg                 config.Config
}

func NewMagaciniResource(service *service.BaseService[domain.Magacini], magaciniRepo *repository.BaseRepository[domain.Magacini], fvrRepo *repository.BaseRepository[domain.Fvr], cfg config.Config) *MagaciniResource {
	rs := &MagaciniResource{
		service:      service,
		magaciniRepo: magaciniRepo,
		fvrRepo:      fvrRepo,
		cfg:          cfg,
	}
	rs.setMagaciniTableFields()
	return rs
}

func (s *MagaciniResource) setMagaciniTableFields() {
	s.magaciniTableFields = []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "8"},
		{Name: "nadmag", Label: "Nadređeni", Width: "12"},
		{Name: "tipmag", Label: "Tip", Width: "8"},
		{Name: "opis", Label: "Opis magacina", Width: "30"},
		{Name: "magosoba", Label: "Osoba magacina", Width: "20"},
		{Name: "nacvodzal", Label: "Način vođenja", Width: "12"},
		{Name: "tipcene", Label: "Tip cena", Width: "10"},
	}
}

func (s *MagaciniResource) GetMagaciniTableFields() []domain.Fields {
	return s.magaciniTableFields
}

func (s *MagaciniResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}

func (s *MagaciniResource) Create(ctx context.Context, entity *domain.Magacini, idField string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, 0, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "magacini",
		EntityType:    reflect.TypeOf(entity).Elem(),
		TableFields:   s.GetMagaciniTableFields(),
		IncludeGodKar: true,
	}

	qb := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.BuildInsert(ctx, fields, idField)

	tx, err := s.magaciniRepo.BeginTx()
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
		return nil, 0, fmt.Errorf("insert magacini failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, 0, nil
}

func (s *MagaciniResource) Update(ctx context.Context, entity *domain.Magacini, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "magacini",
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

	tx, err := s.magaciniRepo.BeginTx()
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
		return nil, fmt.Errorf("update magacini failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, nil
}

func (s *MagaciniResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *MagaciniResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Magacini, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *MagaciniResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Magacini, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *MagaciniResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Magacini, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *MagaciniResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *MagaciniResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *MagaciniResource) MapEntityToValues(entity *domain.Magacini, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *MagaciniResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

func (s *MagaciniResource) GetAllMagacini(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	hasGod, hasKar := s.magaciniRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("SELECT magaciniid, mag, opis, tipmag, adresa, pobro, mesto, nadmag, magosoba, tel, fax, tipzal, tipcene, nacvodzal, analiza, email, tipart FROM magacini", true)

	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" && printType == common.TipStampePreview {
		qb.AddCustomSearchCondition([]string{"opis", "magosoba", "email"}, searchText)
	}
	if !getTotalRecords {
		if sortBy != "" {
			qb.AddOrderBy(fmt.Sprintf("magacini.%s", sortBy))
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
	entities, err := s.magaciniRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
			fmt.Sprintf("%d", entity.Nadmag),
			entity.Tipmag,
			entity.Opis,
			entity.Magosoba,
			fmt.Sprintf("%d", entity.Nacvodzal),
			fmt.Sprintf("%d", entity.Tipcene),
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.MagaciniID), Fields: fields, HasUpdate: true, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

func (s *MagaciniResource) ValidateEntity(ctx context.Context, entity *domain.Magacini) []domain.FieldError {
	var errors []domain.FieldError

	if entity.Opis == "" {
		errors = append(errors, domain.FieldError{
			Field:        "opis",
			ErrorMessage: "Opis magacina je obavezan",
		})
	}

	return errors
}

func (s *MagaciniResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return common.GetFvrData(ctx, s.fvrRepo)
}
