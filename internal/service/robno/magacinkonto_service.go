package robno

import (
	"context"
	"database/sql"
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

// MagacinKontoService defines the interface for operations related to Magkonto.
type MagacinKontoService interface {
	service.Service[domain.Magkonto]
	GetAllMagacinKonto(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, typePrint string) error
	GetMagacinKontoTableFields() []domain.Fields
	ValidateEntity(ctx context.Context, entity *domain.Magkonto) []domain.FieldError
	GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Magkonto, error)
	GetMagacinKontoView(ctx context.Context, magkontoID int64) (domain.MagkontoView, bool, error)
	GetFvrData(ctx context.Context) (domain.Fvr, error)
}

// MagacinKontoResource implements the MagacinKontoService interface.
type MagacinKontoResource struct {
	service                 *service.BaseService[domain.Magkonto]
	magacinKontoRepo        *repository.BaseRepository[domain.Magkonto]
	magacinKontoViewRepo    *repository.BaseRepository[domain.MagkontoView]
	fvrRepo                 *repository.BaseRepository[domain.Fvr]
	robno                   *repository.RobnoRepository
	magacinKontoTableFields []domain.Fields
	cfg                     config.Config
}

func NewMagacinKontoResource(service *service.BaseService[domain.Magkonto], magacinKontoRepo *repository.BaseRepository[domain.Magkonto], magacinKontoViewRepo *repository.BaseRepository[domain.MagkontoView], fvrRepo *repository.BaseRepository[domain.Fvr], cfg config.Config) *MagacinKontoResource {
	rs := &MagacinKontoResource{
		service:              service,
		magacinKontoRepo:     magacinKontoRepo,
		magacinKontoViewRepo: magacinKontoViewRepo,
		fvrRepo:              fvrRepo,
		robno:                repository.NewRobnoRepository(magacinKontoRepo.DB),
		cfg:                  cfg,
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
	// The repository fills god/kar from the session and the audit columns, in its own transaction.
	id, err := s.magacinKontoRepo.Create(ctx, entity, idField, fields)
	if err != nil {
		return nil, 0, fmt.Errorf("insert magkonto failed: %w", err)
	}
	return nil, id, nil
}

func (s *MagacinKontoResource) Update(ctx context.Context, entity *domain.Magkonto, idField string, idFieldValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	// Restricted to the tenant of the session; ErrNotFound when the konto zaliha is not there.
	if err := s.magacinKontoRepo.Update(ctx, entity, idField, idFieldValue, fields); err != nil {
		return nil, fmt.Errorf("update magkonto failed: %w", err)
	}
	return nil, nil
}

func (s *MagacinKontoResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}

func (s *MagacinKontoResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Magkonto, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

func (s *MagacinKontoResource) GetByID(ctx context.Context, idField string, idFieldValue int64) (*domain.Magkonto, error) {
	return s.service.GetByID(ctx, idField, idFieldValue)
}

func (s *MagacinKontoResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

func (s *MagacinKontoResource) MapEntityToValues(entity *domain.Magkonto, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}

func (s *MagacinKontoResource) SetFieldCache(cache map[string]reflect.StructField) {
	s.service.SetFieldCache(cache)
}

// magacinKontoViewFrom is the from of the queries of MagkontoView: the konto zaliha with the opis of its
// magacin, the šifra and the naziv of its konto (fkpl of idfkpl) and the nazivi of the other konta (the
// konto of the kontni plan, fkpl.vkonta 2, of the period).
const magacinKontoViewFrom = ` from magkonto m
		left join magacini mg on mg.magaciniid = m.magaciniid
		left join fkpl f on f.idfkpl = m.idfkpl
		left join lateral (select x.naziv from fkpl x where x.god = m.god and x.kar = m.kar and x.vkonta = 2
			and x.konto = m.kontoprih order by x.idfkpl limit 1) kp on true
		left join lateral (select x.naziv from fkpl x where x.god = m.god and x.kar = m.kar and x.vkonta = 2
			and x.konto = m.kontotroska order by x.idfkpl limit 1) kt on true
		left join lateral (select x.naziv from fkpl x where x.god = m.god and x.kar = m.kar and x.vkonta = 2
			and x.konto = m.kontoruc order by x.idfkpl limit 1) kr on true
		left join lateral (select x.naziv from fkpl x where x.god = m.god and x.kar = m.kar and x.vkonta = 2
			and x.konto = m.kontorab order by x.idfkpl limit 1) kb on true`

// magacinKontoViewSelect selects a MagkontoView (see magacinKontoViewFrom).
const magacinKontoViewSelect = `select m.magkontoid, m.magaciniid, m.mag, m.god, m.kar, m.xdatunosa, m.xdatizmene,
		coalesce(m.xopunos, '') as xopunos, m.xopizmene, coalesce(m.konto, '') as konto, coalesce(m.idfkpl, 0) as idfkpl,
		coalesce(m.vkonta, 0) as vkonta, coalesce(m.kontoprih, '') as kontoprih, coalesce(m.kontotroska, '') as kontotroska,
		coalesce(m.kontoruc, '') as kontoruc, coalesce(m.kontorab, '') as kontorab,
		coalesce(mg.opis, '') as magacinopis, coalesce(f.sifra, '') as kontosifra, coalesce(f.naziv, '') as kontonaziv,
		coalesce(kp.naziv, '') as kontoprihnaziv, coalesce(kt.naziv, '') as kontotroskanaziv,
		coalesce(kr.naziv, '') as kontorucnaziv, coalesce(kb.naziv, '') as kontorabnaziv` + magacinKontoViewFrom

// magacinKontoSort are the columns of the grid it can be sorted by.
var magacinKontoSort = map[string]string{"mag": "m.mag", "konto": "m.konto", "kontoprih": "m.kontoprih",
	"kontotroska": "m.kontotroska", "kontoruc": "m.kontoruc"}

// GetAllMagacinKonto fills the grid of the konta zaliha of the magacini of the period (a page of them;
// with getTotalRecords only their number), with the naziv of the konto zaliha and the opis of the
// magacin. The key of a row is magkontoid.
func (s *MagacinKontoResource) GetAllMagacinKonto(ctx context.Context, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}
	b := querybuilder.New(magacinKontoViewSelect).
		Where("m.god = ?", userSession.SelectedGod).
		Where("m.kar = ?", userSession.SelectedKar)
	if printType == common.TipStampePreview {
		b.AddSearch([]string{"m.mag", "m.konto", "f.naziv", "mg.opis", "m.kontoprih", "m.kontotroska", "m.kontoruc"}, searchText)
	}

	if getTotalRecords && printType == common.TipStampePreview {
		total, err := s.magacinKontoViewRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}

	if _, ok := magacinKontoSort[sortBy]; ok {
		b.OrderByKey(sortBy, sortOrder, magacinKontoSort).OrderBy("m.magkontoid")
	} else {
		b.OrderBy("m.mag, m.konto, m.magkontoid")
	}
	if printType == common.TipStampePreview {
		b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
	}
	entities, err := s.magacinKontoViewRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	for _, entity := range *entities {
		fields := []string{
			fmt.Sprintf("%d", entity.Mag),
			strings.TrimSpace(entity.Konto + " " + entity.KontoSifra),
			entity.KontoNaziv,
			entity.MagacinOpis,
			entity.KontoPrih,
			entity.KontoTroska,
			entity.KontoRuc,
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.MagkontoID), Fields: fields, HasUpdate: true, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

// GetMagacinKontoView returns the konto zaliha of the id (magkontoid) of the period with the texts of its
// keys (found is false without it).
func (s *MagacinKontoResource) GetMagacinKontoView(ctx context.Context, magkontoID int64) (domain.MagkontoView, bool, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return domain.MagkontoView{}, false, errors.New(common.ErrMsgUserSessionNotFound)
	}
	b := querybuilder.New(magacinKontoViewSelect).
		Where("m.god = ?", userSession.SelectedGod).
		Where("m.kar = ?", userSession.SelectedKar).
		Where("m.magkontoid = ?", magkontoID)
	entities, err := s.magacinKontoViewRepo.Select(ctx, b)
	if err != nil {
		return domain.MagkontoView{}, false, err
	}
	if entities == nil || len(*entities) == 0 {
		return domain.MagkontoView{}, false, nil
	}
	return (*entities)[0], true, nil
}

// ValidateEntity validates a konto zaliha of a magacin before it is saved and fills what is read from
// its keys: the magacin is required and of the period (its broj mag), the konto zaliha is required and a
// konto (fkpl.vkonta 2) or an analitika (vkonta 1) of the kontni plan of the period (its konto and
// vkonta), the magacin has it only once (entity.MagkontoID is the row that is changed, 0 for a new one),
// and the konta of prihod, trošak, RUC and rabat, when given, are konta of the kontni plan (vkonta 2).
func (s *MagacinKontoResource) ValidateEntity(ctx context.Context, entity *domain.Magkonto) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("magaciniid", common.ErrMsgUserSessionNotFound)
		return fieldErrors
	}
	// The magacin.
	if entity.MagaciniID <= 0 {
		add("magaciniid", common.ErrMsgObavezanPodatak)
	} else if mag, err := s.robno.MagacinMag(ctx, entity.MagaciniID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			add("magaciniid", robnoUnosMsgNePostoji)
		} else {
			add("magaciniid", common.ErrMsgReadData)
		}
	} else {
		entity.Mag = mag
	}

	// The konto zaliha: a konto or an analitika of the kontni plan.
	if entity.IDFkpl <= 0 {
		add("idfkpl", common.ErrMsgObavezanPodatak)
	} else if konto, vkonta, err := s.robno.KontoZaliha(ctx, entity.IDFkpl); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			add("idfkpl", robnoUnosMsgNePostoji)
		} else {
			add("idfkpl", common.ErrMsgReadData)
		}
	} else {
		entity.Konto, entity.VKonta = konto, vkonta
	}

	// The magacin has the konto zaliha only once.
	if len(fieldErrors) == 0 {
		if count, err := s.robno.MagkontoDrugi(ctx, entity.MagaciniID, entity.IDFkpl, entity.MagkontoID); err != nil {
			add("idfkpl", common.ErrMsgReadData)
		} else if count > 0 {
			add("idfkpl", "Magacin već ima ovaj konto.")
		}
	}

	// The konta of prihod, trošak, RUC and rabat (optional): konta of the kontni plan.
	for _, k := range []struct {
		field string
		value *string
	}{{"kontoprih", &entity.KontoPrih}, {"kontotroska", &entity.KontoTroska}, {"kontoruc", &entity.KontoRuc}, {"kontorab", &entity.KontoRab}} {
		*k.value = strings.TrimSpace(*k.value)
		if *k.value == "" {
			continue
		}
		if count, err := s.robno.SintetickiKontoPostoji(ctx, *k.value); err != nil {
			add(k.field, common.ErrMsgReadData)
		} else if count == 0 {
			add(k.field, robnoUnosMsgNePostoji)
		}
	}
	return fieldErrors
}

func (s *MagacinKontoResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return repository.GetFvrData(ctx, s.fvrRepo)
}
