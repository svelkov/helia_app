package robno

import (
	"context"
	"errors"
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"
	"strings"
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
	// The repository fills god/kar from the session and the audit columns, in its own transaction.
	if _, err := s.artikliRepo.Create(ctx, entity, idField, fields); err != nil {
		return nil, 0, fmt.Errorf("insert artikli failed: %w", err)
	}
	return nil, 0, nil
}

func (s *ArtikliResource) Update(ctx context.Context, entity *domain.Rsif, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	// Restricted to the tenant of the session; ErrNotFound when the artikal is not there.
	if err := s.artikliRepo.Update(ctx, entity, idField, idFieldValue, fields); err != nil {
		return nil, fmt.Errorf("update artikli failed: %w", err)
	}
	return nil, nil
}

func (s *ArtikliResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *ArtikliResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Rsif, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *ArtikliResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Rsif, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *ArtikliResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *ArtikliResource) MapEntityToValues(entity *domain.Rsif, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *ArtikliResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

// artikliSortColumns maps the sortBy keys of the artikli table (the column names) to columns.
var artikliSortColumns = map[string]string{
	"rsifid": "rsif.rsifid", "sifra": "rsif.sifra", "naziv": "rsif.naziv", "komercopis": "rsif.komercopis",
	"jm": "rsif.jm", "pro": "rsif.pro", "tip": "rsif.tip", "barkod": "rsif.barkod", "konto": "rsif.konto",
	"model": "rsif.model", "kvalitet": "rsif.kvalitet", "serbr": "rsif.serbr", "zemljaproizv": "rsif.zemljaproizv",
}

func (s *ArtikliResource) GetAllArtikli(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	b := querybuilder.New("SELECT rsifid, sifra, naziv, komercopis, jm, pro, tip, barkod, konto, model, kvalitet, serbr, zemljaproizv FROM rsif")
	if err := s.artikliRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return err
	}
	if printType == common.TipStampePreview {
		b.AddSearch([]string{"naziv", "komercopis", "pro", "serbr"}, searchText)
	}
	if getTotalRecords && printType == common.TipStampePreview {
		total, err := s.artikliRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	if !getTotalRecords {
		b.OrderByKey(strings.ToLower(sortBy), sortOrder, artikliSortColumns)
		if printType == common.TipStampePreview {
			b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
		}
	}

	entities, err := s.artikliRepo.Select(ctx, b)
	if err != nil {
		return err
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
	return repository.GetFvrData(ctx, s.fvrRepo)
}
