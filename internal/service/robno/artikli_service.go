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

// ArtikliService defines the interface for operations related to Artikli.
type ArtikliService interface {
	service.Service[domain.Rsif]
	GetAllArtikli(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, tipStampe string) error
	GetArtikliTableFields() []domain.Fields
	ValidateEntity(ctx context.Context, entity *domain.Rsif) []domain.FieldError
	GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Rsif, error)
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// ArtikliResource implements the ArtikliService interface.
type ArtikliResource struct {
	service            *service.BaseService[domain.Rsif]
	artikliRepo        *repository.BaseRepository[domain.Rsif]
	fvrRepo            *repository.BaseRepository[domain.Fvr]
	artikliTableFields []domain.Fields
	cfg                config.Config
}

func NewArtikliResource(service *service.BaseService[domain.Rsif], artikliRepo *repository.BaseRepository[domain.Rsif], fvrRepo *repository.BaseRepository[domain.Fvr], cfg config.Config) *ArtikliResource {
	rs := &ArtikliResource{
		service:     service,
		artikliRepo: artikliRepo,
		fvrRepo:     fvrRepo,
		cfg:         cfg,
	}
	rs.setArtikliTableFields()
	return rs
}

func (s *ArtikliResource) setArtikliTableFields() {
	s.artikliTableFields = []domain.Fields{
		{Name: "sifra", Label: "Šifra", Width: "10"},
		{Name: "naziv", Label: "Naziv artikla", Width: "25"},
		{Name: "komercopis", Label: "Komercialni opis", Width: "25"},
		{Name: "jm", Label: "JM", Width: "8"},
		{Name: "pro", Label: "Proizvođač", Width: "15"},
		{Name: "tip", Label: "Tip", Width: "10"},
		{Name: "barkod", Label: "Barkod", Width: "15"},
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "model", Label: "Model", Width: "10"},
		{Name: "kvalitet", Label: "Kvalitet", Width: "10"},
		{Name: "serbr", Label: "Serijski broj", Width: "12"},
		{Name: "zemljaproizv", Label: "Zemlja proiz.", Width: "12"},
	}
}

func (s *ArtikliResource) GetArtikliTableFields() []domain.Fields {
	return s.artikliTableFields
}

func (s *ArtikliResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}

func (s *ArtikliResource) Create(ctx context.Context, entity *domain.Rsif, idField string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, 0, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "rsif",
		EntityType:    reflect.TypeOf(entity).Elem(),
		TableFields:   s.GetArtikliTableFields(),
		IncludeGodKar: true,
	}

	qb := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.BuildInsert(ctx, fields, idField)

	tx, err := s.artikliRepo.BeginTx()
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
		return nil, 0, fmt.Errorf("insert artikli failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, 0, nil
}

func (s *ArtikliResource) Update(ctx context.Context, entity *domain.Rsif, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "rsif",
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

	tx, err := s.artikliRepo.BeginTx()
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
		return nil, fmt.Errorf("update artikli failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, nil
}

func (s *ArtikliResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *ArtikliResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Rsif, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *ArtikliResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Rsif, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *ArtikliResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Rsif, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *ArtikliResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *ArtikliResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *ArtikliResource) MapEntityToValues(entity *domain.Rsif, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *ArtikliResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

func (s *ArtikliResource) GetAllArtikli(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	hasGod, hasKar := s.artikliRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("SELECT rsifid, sifra, naziv, komercopis, jm, pro, tip, barkod, konto, model, kvalitet, serbr, zemljaproizv FROM rsif", true)

	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" && printType == common.TipStampePreview {
		qb.AddCustomSearchCondition([]string{"naziv", "komercopis", "pro", "serbr"}, searchText)
	}
	if !getTotalRecords {
		if sortBy != "" {
			qb.AddOrderBy(fmt.Sprintf("rsif.%s", sortBy))
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
	entities, err := s.artikliRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	for _, entity := range *entities {
		fields := []string{
			fmt.Sprintf("%d", entity.Sifra),
			entity.Naziv,
			entity.KomercOpis,
			entity.JM,
			entity.Pro,
			entity.Tip,
			entity.Barkod,
			entity.Konto,
			entity.Model,
			entity.Kvalitet,
			entity.Serbr,
			entity.ZemljaProizv,
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.RsifID), Fields: fields, HasUpdate: true, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

func (s *ArtikliResource) ValidateEntity(ctx context.Context, entity *domain.Rsif) []domain.FieldError {
	var errors []domain.FieldError

	if entity.Naziv == "" {
		errors = append(errors, domain.FieldError{
			Field:        "naziv",
			ErrorMessage: "Naziv artikla je obavezan",
		})
	}

	if entity.Sifra == 0 {
		errors = append(errors, domain.FieldError{
			Field:        "sifra",
			ErrorMessage: "Šifra artikla je obavezna",
		})
	}

	return errors
}

func (s *ArtikliResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return common.GetFvrData(ctx, s.fvrRepo)
}
