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

// KomercijalistiService defines the interface for operations related to Komercijalisti.
type KomercijalistiService interface {
	service.Service[domain.Komercijalisti]
	GetAllKomercijalisti(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, tipStampe string) error
	GetKomercijalistiTableFields() []domain.Fields
	ValidateEntity(ctx context.Context, entity *domain.Komercijalisti) []domain.FieldError
	GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Komercijalisti, error)
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// KomercijalistiResource implements the KomercijalistiService interface.
type KomercijalistiResource struct {
	service                   *service.BaseService[domain.Komercijalisti]
	komercijalistiRepo        *repository.BaseRepository[domain.Komercijalisti]
	fvrRepo                   *repository.BaseRepository[domain.Fvr]
	komercijalistiTableFields []domain.Fields
	cfg                       config.Config
}

func NewKomercijalistiResource(service *service.BaseService[domain.Komercijalisti], komercijalistiRepo *repository.BaseRepository[domain.Komercijalisti], fvrRepo *repository.BaseRepository[domain.Fvr], cfg config.Config) *KomercijalistiResource {
	rs := &KomercijalistiResource{
		service:            service,
		komercijalistiRepo: komercijalistiRepo,
		fvrRepo:            fvrRepo,
		cfg:                cfg,
	}
	rs.setKomercijalistiTableFields()
	return rs
}

func (s *KomercijalistiResource) setKomercijalistiTableFields() {
	s.komercijalistiTableFields = []domain.Fields{
		{Name: "sifkom", Label: "Šifra", Width: "10"},
		{Name: "sifnadred", Label: "Nadređeni", Width: "12"},
		{Name: "imeprezime", Label: "Ime i Prezime", Width: "25"},
		{Name: "adresa", Label: "Adresa", Width: "20"},
		{Name: "mesto", Label: "Mesto", Width: "15"},
		{Name: "telposao", Label: "Tel. Posao", Width: "15"},
		{Name: "telmob", Label: "Tel. Mobilni", Width: "15"},
		{Name: "totprod", Label: "Ukupna Prodaja", Width: "15"},
		{Name: "totprofit", Label: "Ukupan Profit", Width: "15"},
		{Name: "zaddatprod", Label: "Poslednji Datum", Width: "15"},
		{Name: "totnaplaceno", Label: "Ukupno Naplaćeno", Width: "15"},
	}
}

func (s *KomercijalistiResource) GetKomercijalistiTableFields() []domain.Fields {
	return s.komercijalistiTableFields
}

func (s *KomercijalistiResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}

func (s *KomercijalistiResource) Create(ctx context.Context, entity *domain.Komercijalisti, idField string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, 0, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "komercijalisti",
		EntityType:    reflect.TypeOf(entity).Elem(),
		TableFields:   s.GetKomercijalistiTableFields(),
		IncludeGodKar: true,
	}

	qb := common.NewRepositoryQueryBuilder(config)
	hasGod, hasKar := qb.CheckGodKarFields()
	if hasGod || hasKar {
		qb.AddGodKarConditions(hasGod, hasKar, userSession.SelectedGod, userSession.SelectedKar)
	}

	sqlQuery, args := qb.BuildInsert(ctx, fields, idField)

	tx, err := s.komercijalistiRepo.BeginTx()
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
		return nil, 0, fmt.Errorf("insert komercijalisti failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, 0, nil
}

func (s *KomercijalistiResource) Update(ctx context.Context, entity *domain.Komercijalisti, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	config := common.RepositoryConfig{
		TableName:     "komercijalisti",
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

	tx, err := s.komercijalistiRepo.BeginTx()
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
		return nil, fmt.Errorf("update komercijalisti failed: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return nil, nil
}

func (s *KomercijalistiResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *KomercijalistiResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Komercijalisti, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *KomercijalistiResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Komercijalisti, error) {
	return s.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *KomercijalistiResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Komercijalisti, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *KomercijalistiResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *KomercijalistiResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

func (s *KomercijalistiResource) MapEntityToValues(entity *domain.Komercijalisti, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *KomercijalistiResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

func (s *KomercijalistiResource) GetAllKomercijalisti(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	hasGod, hasKar := s.komercijalistiRepo.GetHasGodHasKar()
	qb := common.NewQueryBuilder("SELECT komid, sifkom, sifnadred, imeprezime, adresa, mesto, telposao, telmob, totprod, totprofit, zaddatprod, totnaplaceno, loginname FROM komercijalisti", true)

	if hasGod {
		qb.AddEqual("god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", userSession.SelectedKar)
	}
	if searchText != "" && printType == common.TipStampePreview {
		qb.AddCustomSearchCondition([]string{"imeprezime", "adresa", "telposao", "telmob"}, searchText)
	}
	if !getTotalRecords {
		if sortBy != "" {
			qb.AddOrderBy(fmt.Sprintf("komercijalisti.%s", sortBy))
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
	entities, err := s.komercijalistiRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}

	if getTotalRecords && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*entities), pageSize)
		return nil
	}

	for _, entity := range *entities {
		fields := []string{
			fmt.Sprintf("%d", entity.Sifkom),
			fmt.Sprintf("%d", entity.SifNadred),
			entity.ImePrezime,
			entity.Adresa,
			entity.Mesto,
			entity.TelPosao,
			entity.TelMob,
			fmt.Sprintf("%.2f", entity.TotProd),
			fmt.Sprintf("%.2f", entity.TotProfit),
			entity.ZadDatProd.Format("02.01.2006"),
			fmt.Sprintf("%.2f", entity.TotNaplaceno),
			entity.LoginName,
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.KomID), Fields: fields, HasUpdate: true, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

func (s *KomercijalistiResource) ValidateEntity(ctx context.Context, entity *domain.Komercijalisti) []domain.FieldError {
	var errors []domain.FieldError

	if entity.ImePrezime == "" {
		errors = append(errors, domain.FieldError{
			Field:        "imeprezime",
			ErrorMessage: "Ime i prezime je obavezno",
		})
	}

	if entity.Sifkom == 0 {
		errors = append(errors, domain.FieldError{
			Field:        "sifkom",
			ErrorMessage: "Šifra komercijaliste je obavezna",
		})
	}

	return errors
}

func (s *KomercijalistiResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return common.GetFvrData(ctx, s.fvrRepo)
}
