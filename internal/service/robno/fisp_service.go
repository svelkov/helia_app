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

type FispService interface {
	service.Service[domain.Fisp]
	GetAllFisp(context.Context, *domain.TableData, int, int, bool, string, string, string, string) error
	GetFispTableFields() []domain.Fields
	ValidateEntity(context.Context, *domain.Fisp) []domain.FieldError
	GetByID(context.Context, string, int64) (*domain.Fisp, error)
	GetFvrData(context.Context) (domain.Fvr, error)
}

type FispResource struct {
	service      *service.BaseService[domain.Fisp]
	fispRepo     *repository.BaseRepository[domain.Fisp]
	partneriRepo *repository.BaseRepository[domain.Partneri]
	fvrRepo      *repository.BaseRepository[domain.Fvr]
	fields       []domain.Fields
	cfg          config.Config
}

func NewFispResource(base *service.BaseService[domain.Fisp], repo *repository.BaseRepository[domain.Fisp], partneri *repository.BaseRepository[domain.Partneri], fvr *repository.BaseRepository[domain.Fvr], cfg config.Config) *FispResource {
	r := &FispResource{service: base, fispRepo: repo, partneriRepo: partneri, fvrRepo: fvr, cfg: cfg}
	r.fields = []domain.Fields{
		{Name: "konto", Label: "Konto partnera", Width: "15"},
		{Name: "sifra", Label: "Sifra partnera", Width: "15"},
		{Name: "partnernaziv", Label: "Naziv partnera", Width: "30"},
		{Name: "mi", Label: "Sifra MI", Width: "12"},
		{Name: "naziv", Label: "Naziv", Width: "30"},
		{Name: "adresa", Label: "Adresa", Width: "25"},
	}
	return r
}

func (s *FispResource) GetFispTableFields() []domain.Fields { return s.fields }
func (s *FispResource) GetFieldCache() map[string]reflect.StructField {
	return s.service.GetFieldCache()
}
func (s *FispResource) SetFieldCache(v map[string]reflect.StructField) { s.service.SetFieldCache(v) }
func (s *FispResource) Create(ctx context.Context, e *domain.Fisp, id string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return nil, 0, errors.New(common.ErrMsgUserSessionNotFound)
	}
	if err := s.setPartnerID(ctx, e); err != nil {
		return nil, 0, err
	}
	fields = replaceField(fields, "idpartneri", fmt.Sprint(e.IDPartneri))
	q := common.NewRepositoryQueryBuilder(common.RepositoryConfig{TableName: "fisp", EntityType: reflect.TypeOf(e).Elem(), TableFields: s.fields, IncludeGodKar: true})
	hasGod, hasKar := q.CheckGodKarFields()
	if hasGod || hasKar {
		u := domain.GetSessionFromStdContext(ctx)
		q.AddGodKarConditions(hasGod, hasKar, u.SelectedGod, u.SelectedKar)
	}
	sqlQuery, args := q.BuildInsert(ctx, fields, id)
	tx, err := s.fispRepo.BeginTx()
	if err != nil {
		return nil, 0, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = tx.QueryRowContext(ctx, sqlQuery, args...).Err(); err != nil {
		return nil, 0, fmt.Errorf("insert fisp failed: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("error committing transaction: %w", err)
	}
	return nil, 0, nil
}
func (s *FispResource) Update(ctx context.Context, e *domain.Fisp, id string, idValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	if err := s.setPartnerID(ctx, e); err != nil {
		return nil, err
	}
	fields = replaceField(fields, "idpartneri", fmt.Sprint(e.IDPartneri))
	q := common.NewRepositoryQueryBuilder(common.RepositoryConfig{TableName: "fisp", EntityType: reflect.TypeOf(e).Elem(), TableFields: fields, IncludeGodKar: true})
	hasGod, hasKar := q.CheckGodKarFields()
	if hasGod || hasKar {
		u := domain.GetSessionFromStdContext(ctx)
		q.AddGodKarConditions(hasGod, hasKar, u.SelectedGod, u.SelectedKar)
	}
	sqlQuery, args := q.BuildUpdate(ctx, fields, id, idValue)
	tx, err := s.fispRepo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = tx.QueryRowContext(ctx, sqlQuery, args...).Err(); err != nil {
		return nil, fmt.Errorf("update fisp failed: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}
	return nil, nil
}
func (s *FispResource) Delete(ctx context.Context, id string, value int64) error {
	return s.service.Delete(ctx, id, value)
}
func (s *FispResource) GetAll(ctx context.Context, p, size int, f []domain.Fields, id, search, sort, order string) (*[]domain.Fisp, error) {
	return s.service.GetAll(ctx, p, size, f, id, search, sort, order)
}
func (s *FispResource) GetAllCustom(ctx context.Context, q, w string, a []interface{}, l, o string) (*[]domain.Fisp, error) {
	return s.service.GetAllCustom(ctx, q, w, a, l, o)
}
func (s *FispResource) GetByID(ctx context.Context, id string, value int64) (*domain.Fisp, error) {
	return s.service.GetByID(ctx, id, value)
}
func (s *FispResource) GetTotalRecords(ctx context.Context, f []domain.Fields, search string) (int, error) {
	return s.service.GetTotalRecords(ctx, f, search)
}
func (s *FispResource) GetTotalRecordsCustom(ctx context.Context, q, w string, a []interface{}, l, o string) (int, error) {
	return s.service.GetTotalRecordsCustom(ctx, q, w, a, l, o)
}
func (s *FispResource) MapEntityToValues(e *domain.Fisp, f []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(e, f)
}
func (s *FispResource) ValidateEntity(ctx context.Context, e *domain.Fisp) []domain.FieldError {
	var r []domain.FieldError
	if e.Konto == "" {
		r = append(r, domain.FieldError{Field: "konto", ErrorMessage: "Konto je obavezan"})
	}
	if e.Sifra == "" {
		r = append(r, domain.FieldError{Field: "sifra", ErrorMessage: "Šifra partnera je obavezna"})
	}
	if e.Naziv == "" {
		r = append(r, domain.FieldError{Field: "naziv", ErrorMessage: "Naziv mesta isporuke je obavezan"})
	}
	return r
}
func (s *FispResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	return common.GetFvrData(ctx, s.fvrRepo)
}

func replaceField(fields []domain.Fields, name, value string) []domain.Fields {
	for i := range fields {
		if fields[i].Name == name {
			fields[i].Value = value
			return fields
		}
	}
	return append(fields, domain.Fields{Name: name, Value: value})
}

func (s *FispResource) setPartnerID(ctx context.Context, e *domain.Fisp) error {
	if s.partneriRepo == nil {
		return errors.New("partneri repository is not initialized")
	}
	qb := common.NewQueryBuilder("SELECT idpartneri, sifra FROM partneri", true)
	qb.AddEqual("sifra", e.Sifra)
	qb.AddOrderBy("idpartneri")
	qb.SetLimit(1)
	sqlQuery, args := qb.Build()
	partners, err := s.partneriRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return fmt.Errorf("find partner by code failed: %w", err)
	}
	if len(*partners) == 0 {
		return fmt.Errorf("partner sa sifrom %q ne postoji", e.Sifra)
	}
	e.IDPartneri = (*partners)[0].IDPartneri
	return nil
}

func (s *FispResource) GetAllFisp(ctx context.Context, tbl *domain.TableData, page, size int, total bool, sort, order, search, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}
	qb := common.NewQueryBuilder(`SELECT fisp.fispid, fisp.idpartneri, fisp.god, fisp.kar, fisp.konto, fisp.sifra, fisp.mi, fisp.naziv, fisp.adresa, fisp.mesto, fisp.pobro, fisp.pib, fisp.kontaktosb, fisp.gln, fisp.email, COALESCE(p.naziv, '') AS partnernaziv FROM fisp`, true)
	qb.AddJoin("LEFT JOIN partneri p ON p.idpartneri = fisp.idpartneri")
	hasGod, hasKar := s.fispRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("fisp.god", userSession.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("fisp.kar", userSession.SelectedKar)
	}
	if search != "" && printType == common.TipStampePreview {
		qb.SetEntityType(reflect.TypeOf(domain.Fisp{}))
		qb.AddCustomSearchCondition([]string{"fisp.konto", "fisp.sifra", "p.naziv", "fisp.naziv", "fisp.adresa"}, search)
	}
	if !total {
		sortMap := map[string]string{"konto": "fisp.konto", "sifra": "fisp.sifra", "partnernaziv": "p.naziv", "mi": "fisp.mi", "naziv": "fisp.naziv", "adresa": "fisp.adresa"}
		if v, ok := sortMap[sort]; ok {
			qb.AddOrderBy(v)
			if order == "DESC" {
				qb.AddSortOrder("DESC")
			} else {
				qb.AddSortOrder("ASC")
			}
		}
		if printType == common.TipStampePreview {
			qb.SetLimit(size)
			qb.SetOffset((page - 1) * size)
		}
	}
	sqlQuery, args := qb.Build()
	rows, err := s.fispRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return err
	}
	if total && printType == common.TipStampePreview {
		common.SetTableTotalRecords(tbl, len(*rows), size)
		return nil
	}
	for _, e := range *rows {
		tbl.Rows = append(tbl.Rows, domain.TableRow{ID: fmt.Sprint(e.FispID), Fields: []string{e.Konto, e.Sifra, e.PartnerNaziv, fmt.Sprint(e.MI), e.Naziv, e.Adresa}, HasUpdate: true, HasDelete: true})
	}
	return nil
}
