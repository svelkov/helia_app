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
	// The repository fills god/kar from the session and the audit columns, in its own transaction.
	if _, err := s.komercijalistiRepo.Create(ctx, entity, idField, fields); err != nil {
		return nil, 0, fmt.Errorf("insert komercijalisti failed: %w", err)
	}
	return nil, 0, nil
}

func (s *KomercijalistiResource) Update(ctx context.Context, entity *domain.Komercijalisti, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	// Restricted to the tenant of the session; ErrNotFound when the komercijalista is not there.
	if err := s.komercijalistiRepo.Update(ctx, entity, idField, idFieldValue, fields); err != nil {
		return nil, fmt.Errorf("update komercijalisti failed: %w", err)
	}
	return nil, nil
}

func (s *KomercijalistiResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *KomercijalistiResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Komercijalisti, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *KomercijalistiResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Komercijalisti, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *KomercijalistiResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *KomercijalistiResource) MapEntityToValues(entity *domain.Komercijalisti, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *KomercijalistiResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

// komercijalistiSortColumns maps the sortBy keys of the komercijalisti table (the column names) to columns.
var komercijalistiSortColumns = map[string]string{
	"komid": "komercijalisti.komid", "sifkom": "komercijalisti.sifkom", "sifnadred": "komercijalisti.sifnadred",
	"imeprezime": "komercijalisti.imeprezime", "adresa": "komercijalisti.adresa", "mesto": "komercijalisti.mesto",
	"telposao": "komercijalisti.telposao", "telmob": "komercijalisti.telmob", "totprod": "komercijalisti.totprod",
	"totprofit": "komercijalisti.totprofit", "zaddatprod": "komercijalisti.zaddatprod",
	"totnaplaceno": "komercijalisti.totnaplaceno", "loginname": "komercijalisti.loginname",
}

func (s *KomercijalistiResource) GetAllKomercijalisti(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText string, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	b := querybuilder.New("SELECT komid, sifkom, sifnadred, imeprezime, adresa, mesto, telposao, telmob, totprod, totprofit, zaddatprod, totnaplaceno, loginname FROM komercijalisti")
	if err := s.komercijalistiRepo.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return err
	}
	if printType == common.TipStampePreview {
		b.AddSearch([]string{"imeprezime", "adresa", "telposao", "telmob"}, searchText)
	}
	if getTotalRecords && printType == common.TipStampePreview {
		total, err := s.komercijalistiRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	if !getTotalRecords {
		b.OrderByKey(strings.ToLower(sortBy), sortOrder, komercijalistiSortColumns)
		if printType == common.TipStampePreview {
			b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
		}
	}

	entities, err := s.komercijalistiRepo.Select(ctx, b)
	if err != nil {
		return err
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
	return repository.GetFvrData(ctx, s.fvrRepo)
}
