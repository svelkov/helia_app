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
	rs.SetMagaciniTableFields()
	return rs
}
func (s *MagaciniResource) SetMagaciniTableFields() {
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
	// The repository fills god/kar from the session and the audit columns, in its own transaction.
	if _, err := s.magaciniRepo.Create(ctx, entity, idField, fields); err != nil {
		return nil, 0, fmt.Errorf("insert magacini failed: %w", err)
	}
	return nil, 0, nil
}

func (s *MagaciniResource) Update(ctx context.Context, entity *domain.Magacini, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	// Restricted to the tenant of the session; ErrNotFound when the magacin is not there.
	if err := s.magaciniRepo.Update(ctx, entity, idField, idFieldValue, fields); err != nil {
		return nil, fmt.Errorf("update magacini failed: %w", err)
	}
	return nil, nil
}

func (s *MagaciniResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *MagaciniResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Magacini, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *MagaciniResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Magacini, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *MagaciniResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *MagaciniResource) MapEntityToValues(entity *domain.Magacini, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *MagaciniResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

// magaciniSortColumns maps the sortBy keys of the magacini table (the column names) to columns.
var magaciniSortColumns = map[string]string{
	"magaciniid": "magacini.magaciniid", "mag": "magacini.mag", "opis": "magacini.opis", "tipmag": "magacini.tipmag",
	"adresa": "magacini.adresa", "pobro": "magacini.pobro", "mesto": "magacini.mesto", "nadmag": "magacini.nadmag",
	"magosoba": "magacini.magosoba", "tel": "magacini.tel", "fax": "magacini.fax", "tipzal": "magacini.tipzal",
	"tipcene": "magacini.tipcene", "nacvodzal": "magacini.nacvodzal", "analiza": "magacini.analiza",
	"email": "magacini.email", "tipart": "magacini.tipart",
}

func (s *MagaciniResource) GetAllMagacini(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	b := querybuilder.New("SELECT magaciniid, mag, opis, tipmag, adresa, pobro, mesto, nadmag, magosoba, tel, fax, tipzal, tipcene, nacvodzal, analiza, email, tipart FROM magacini")
	if err := s.magaciniRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return err
	}
	if printType == common.TipStampePreview {
		b.AddSearch([]string{"opis", "magosoba", "email"}, searchText)
	}
	if getTotalRecords && printType == common.TipStampePreview {
		total, err := s.magaciniRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	if !getTotalRecords {
		b.OrderByKey(strings.ToLower(sortBy), sortOrder, magaciniSortColumns)
		if printType == common.TipStampePreview {
			b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
		}
	}

	entities, err := s.magaciniRepo.Select(ctx, b)
	if err != nil {
		return err
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
	return repository.GetFvrData(ctx, s.fvrRepo)
}
