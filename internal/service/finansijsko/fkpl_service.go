package finansijsko

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
	commonsvc "helia/internal/service/common"
	"reflect"
	"strings"
)

// FkplService defines the interface for operations related to Fkpl.
type FkplService interface {
	service.Service[domain.Fkpl]
	TraziKonto(ctx context.Context, konto, sifra, vkonta string) (entities *[]domain.Fkpl, err error)
	KontoSearchForTable(ctx context.Context, tbl *domain.TableData, searchValue, konto, vkonta, fieldName string, fieldColIndex int) error
	GetAllByVkonta(ctx context.Context, vkonta string, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, printType string) error
	GetFvrData(ctx context.Context) (domain.Fvr, error)
	GetFkplTableFields() []domain.Fields
	ValidateEntity(ctx context.Context, entity *domain.Fkpl) []domain.FieldError
	GetAnalitikaForSelect(ctx context.Context) ([]domain.ComboItem, error)
}

// FkplResource implements the FkplService interface.
type FkplResource struct {
	service          *service.BaseService[domain.Fkpl]
	fkplRepo         *repository.BaseRepository[domain.Fkpl]
	fvrRepo          *repository.BaseRepository[domain.Fvr]
	tipAnalitikeRepo *repository.BaseRepository[domain.Tipanalitike]
	partneriRepo     *repository.BaseRepository[domain.Partneri]
	fkplTableFields  []domain.Fields
	cfg              config.Config
	commonSvc        commonsvc.CommonService
}

func NewFkplResource(service *service.BaseService[domain.Fkpl], fkplRepo *repository.BaseRepository[domain.Fkpl], fvrRepo *repository.BaseRepository[domain.Fvr], tipAnalitikeRepo *repository.BaseRepository[domain.Tipanalitike], partneriRepo *repository.BaseRepository[domain.Partneri], cfg config.Config, commonSvc commonsvc.CommonService) *FkplResource {
	rs := &FkplResource{
		service:          service,
		fkplRepo:         fkplRepo,
		fvrRepo:          fvrRepo,
		tipAnalitikeRepo: tipAnalitikeRepo,
		partneriRepo:     partneriRepo,
		cfg:              cfg,
		commonSvc:        commonSvc,
	}
	rs.setKontniPlanTableFields()
	return rs
}

func (s *FkplResource) GetFvrData(ctx context.Context) (domain.Fvr, error) {
	if s.fvrRepo == nil {
		return domain.Fvr{}, fmt.Errorf("fvrRepo not initialized")
	}
	return repository.GetFvrData(ctx, s.fvrRepo)
}

// GetAnalitikaForSelect returns the tipovi analitike that can be selected for a konto
// (CommonService).
func (s *FkplResource) GetAnalitikaForSelect(ctx context.Context) ([]domain.ComboItem, error) {
	return s.commonSvc.GetTipoviAnalitikeComboValues(ctx)
}

func (s *FkplResource) GetFieldCache() map[string]reflect.StructField {
	if s.service.GetFieldCache() == nil {
		s.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return s.service.GetFieldCache()
}
func (s *FkplResource) GetFkplTableFields() []domain.Fields {
	return s.fkplTableFields
}

func (s *FkplResource) Create(ctx context.Context, entity *domain.Fkpl, idFkpl string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, 0, errors.New(common.ErrMsgUserSessionNotFound)
	}
	// The repository fills god/kar from the session and the audit columns, in its own transaction.
	// The handler sends only the column names, so the values come from the entity (before, the
	// fields went in without values and the konto was inserted empty).
	id, err := s.fkplRepo.Create(ctx, entity, common.IDfkpl, fkplFieldValues(entity, fields))
	if err != nil {
		return nil, 0, fmt.Errorf("insert fkpl failed: %w", err)
	}
	return nil, id, nil
}

// fkplFieldValues sets the Value (or IsNull) of every field from the matching column of the konto.
// An unknown column is an error of the caller and keeps its value.
func fkplFieldValues(entity *domain.Fkpl, fields []domain.Fields) []domain.Fields {
	out := make([]domain.Fields, len(fields))
	copy(out, fields)
	for i := range out {
		switch strings.ToLower(out[i].Name) {
		case "konto":
			out[i].Value = entity.Konto
		case "sifra":
			out[i].Value = entity.Sifra
		case "naziv":
			out[i].Value = entity.Naziv
		case "vkonta":
			out[i].Value = fmt.Sprintf("%d", entity.Vkonta)
		case "devizni":
			out[i].Value = fmt.Sprintf("%t", entity.Devizni)
		case "kolicinski":
			out[i].Value = fmt.Sprintf("%t", entity.Kolicinski)
		case "kursirati":
			out[i].Value = fmt.Sprintf("%t", entity.Kursirati)
		case "tipanalitikeid":
			if entity.TipanalitikeID == nil {
				out[i].IsNull = true
			} else {
				out[i].Value = fmt.Sprintf("%d", *entity.TipanalitikeID)
			}
		}
	}
	return out
}
func (s *FkplResource) Update(ctx context.Context, entity *domain.Fkpl, idField string, idFiledValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	// Restricted to the tenant of the session; ErrNotFound when the konto is not there.
	if err := s.fkplRepo.Update(ctx, entity, idField, idFiledValue, fields); err != nil {
		return nil, fmt.Errorf("update fkpl failed: %w", err)
	}
	return nil, nil
}
func (s *FkplResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return s.service.Delete(ctx, idField, idFieldValue)
}
func (s *FkplResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Fkpl, error) {
	return s.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

// fkplSortColumns maps the sortBy keys of the kontni plan table (the column names) to columns.
var fkplSortColumns = map[string]string{
	"idfkpl": "fkpl.idfkpl", "konto": "fkpl.konto", "naziv": "fkpl.naziv", "devizni": "fkpl.devizni",
	"kursirati": "fkpl.kursirati", "vkonta": "fkpl.vkonta", "kolicinski": "fkpl.kolicinski",
}

func (s *FkplResource) GetAllByVkonta(ctx context.Context, vkonta string, tbl *domain.TableData, currentPage, pageSize int, getTotalRecords bool, sortBy, sortOrder, searchText, printType string) error {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	b := querybuilder.New("SELECT idfkpl, konto, coalesce(a.naziv, '') as annaziv, fkpl.naziv, devizni, kursirati, vkonta, kolicinski FROM fkpl").
		AddJoin("left join tipanalitike a on a.tipanalitikeid = fkpl.tipanalitikeid").
		AddEqual("fkpl.vkonta", vkonta) // optional: an empty vrsta konta lists every konto
	if err := s.fkplRepo.ApplyFilters(ctx, b, "fkpl", nil, ""); err != nil {
		return err
	}
	if printType == common.TipStampePreview {
		b.AddSearch([]string{"konto", "a.naziv", "fkpl.naziv"}, searchText)
	}
	if getTotalRecords && printType == common.TipStampePreview {
		total, err := s.fkplRepo.Count(ctx, b)
		if err != nil {
			return err
		}
		common.SetTableTotalRecords(tbl, total, pageSize)
		return nil
	}
	if !getTotalRecords {
		b.OrderByKey(strings.ToLower(sortBy), sortOrder, fkplSortColumns)
		if printType == common.TipStampePreview {
			b.Limit(pageSize).Offset((currentPage - 1) * pageSize)
		}
	}

	entities, err := s.fkplRepo.Select(ctx, b)
	if err != nil {
		return err
	}
	for _, entity := range *entities {
		fields := []string{
			entity.Konto,
			entity.AnalitikaNaziv,
			entity.Naziv,
			fmt.Sprintf("%t", entity.Devizni),
			fmt.Sprintf("%t", entity.Kursirati),
			fmt.Sprintf("%t", entity.Kolicinski),
			fmt.Sprintf("%d", entity.Vkonta),
		}
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: true, HasDelete: true}
		tbl.Rows = append(tbl.Rows, tblRow)
	}

	return nil
}

// GetByID implements FkplService.
func (s *FkplResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Fkpl, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	if !querybuilder.IsIdent(idField) || strings.Contains(idField, ".") {
		return nil, fmt.Errorf("invalid id column %q", idField)
	}
	// The id is a key: Where, never skipped. The period is qualified with fkpl, because the joined
	// tipanalitike has its own god/kar.
	b := querybuilder.New("SELECT idfkpl, konto, coalesce(a.naziv, '') as annaziv, fkpl.naziv, devizni, kolicinski, vkonta, fkpl.tipanalitikeid FROM fkpl").
		AddJoin("left join tipanalitike a on a.tipanalitikeid = fkpl.tipanalitikeid").
		Where("fkpl."+idField+" = ?", idValue)
	if err := s.fkplRepo.ApplyFilters(ctx, b, "fkpl", nil, ""); err != nil {
		return nil, err
	}
	entities, err := s.fkplRepo.Select(ctx, b)
	if err != nil {
		return nil, err
	}
	if len(*entities) == 0 {
		return nil, errors.New("Entity not found")
	}
	return &(*entities)[0], nil
}

// GetTotalRecords implements FkplService.
func (s *FkplResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return s.service.GetTotalRecords(ctx, tableFields, searchText)
}

// MapEntityToValues implements FkplService.
func (s *FkplResource) MapEntityToValues(entity *domain.Fkpl, tableFields []domain.Fields) []domain.Fields {
	return s.service.MapEntityToValues(entity, tableFields)
}
func (s *FkplResource) ValidateEntity(ctx context.Context, entity *domain.Fkpl) []domain.FieldError {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return []domain.FieldError{{Field: "session", ErrorMessage: "Korisnička sesija nije pronađena"}}
	}

	// nDuzSIN depends on the bookkeeping type of the selected firma (FVR.KNJIGOVOD)
	duzSin := userSession.GetDuzSin(s.cfg.NDuzSint)

	var fieldErrors []domain.FieldError
	if entity.Konto == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Obavezan podatak..."})
	}
	if entity.Naziv == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "naziv", ErrorMessage: "Obavezan podatak..."})
	}
	if entity.Vkonta == 1 && entity.TipanalitikeID == nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "tipanalitikeid", ErrorMessage: "Obavezan podatak..."})
	}
	if entity.Vkonta == 2 && len(entity.Konto) <= duzSin {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: fmt.Sprintf("Subsintetički konto mora imati minimum %d karaktera", duzSin+1)})
	}
	if entity.Vkonta == 3 && len(entity.Konto) != duzSin {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: fmt.Sprintf("Sintetički konto mora imati tačno %d karaktera", duzSin)})
	}
	if entity.Vkonta == 4 && len(entity.Konto) != 2 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Grupa konto mora imati tačno 2 karaktera"})
	}
	if entity.Vkonta == 5 && len(entity.Konto) != 1 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Klasa konto mora imati tačno 1 karakter"})
	}
	// The konto must be unique in the period of the session. (The old builder had no entity type
	// here, so it found no god/kar and checked every period.)
	// The pair konto + vrsta analitike is the key; a konto without analitika is compared with the
	// konta without analitika (IS NULL - "= NULL" would never match).
	b := querybuilder.New("SELECT idfkpl FROM fkpl").
		Where("konto = ?", entity.Konto)
	if entity.TipanalitikeID == nil {
		b.Where("tipanalitikeid IS NULL")
	} else {
		b.Where("tipanalitikeid = ?", *entity.TipanalitikeID)
	}
	var existingEntity *[]domain.Fkpl
	err := s.fkplRepo.ApplyFilters(ctx, b, "", nil, "")
	if err == nil {
		existingEntity, err = s.fkplRepo.Select(ctx, b)
	}
	if err != nil {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Greška prilikom provere jedinstvenosti konta"})
	} else if len(*existingEntity) > 0 {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: "konto", ErrorMessage: "Konto sa ovom šifrom i vrstom analitike već postoji"})
	}
	return fieldErrors
}

func (s *FkplResource) TraziKonto(ctx context.Context, konto, sifra, vkonta string) (entities *[]domain.Fkpl, err error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	var b *querybuilder.Builder
	// if vkonta == "1"  then search ibn table partneri for the tipanalitke for that konto and use that tipanalitikeid to search in fkpl, if vkonta == "2" then search directly in fkpl with vkonta = 2 and konto
	if vkonta == "1" {
		b = querybuilder.New(`SELECT ta.tipanalitikeid, p.naziv as naziv, p.sifra as sifra FROM partneri as p`).
			AddJoin("inner join tipanalitike ta on ta.tipanalitikeid = p.tipanalitikeid").
			AddJoin("inner join fkpl f on f.tipanalitikeid = ta.tipanalitikeid")
	}
	if vkonta == "2" {
		b = querybuilder.New(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM fkpl as f`)
	}
	if b == nil {
		return nil, errors.New("Invalid vkonta value")
	}
	if err := s.fkplRepo.ApplyFilters(ctx, b, "f", nil, ""); err != nil {
		return nil, err
	}
	if vkonta == "1" {
		// The partneri of the period too (like KontoSearchForTable).
		if err := s.partneriRepo.ApplyFilters(ctx, b, "p", nil, ""); err != nil {
			return nil, err
		}
	}

	// The optional filters skip empty values by themselves.
	b.AddEqual("f.konto", konto)
	// only search in fkpl if vkonta is 2, if vkonta is 1 then search in partneri and join with fkpl through tipanalitikeid
	if vkonta == "2" {
		b.Where("f.vkonta = ?", vkonta)
	}
	if vkonta == "1" {
		b.AddEqual("p.sifra", sifra)
	}

	entities, err = s.fkplRepo.Select(ctx, b)
	if err != nil {
		return nil, errors.New(common.ErrMsgSearchKonto)
	}
	if len(*entities) == 0 {
		return nil, errors.New(common.ErrMsgNotFound)
	}
	return entities, nil
}
func (s *FkplResource) KontoSearchForTable(ctx context.Context, tbl *domain.TableData, searchValue, konto, vkonta, fieldName string, fieldColIndex int) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}
	var b *querybuilder.Builder

	// if vkonta == "1"  then search ibn table partneri for the tipanalitke for that konto and use that tipanalitikeid to search in fkpl, if vkonta == "2" then search directly in fkpl with vkonta = 2 and konto
	if vkonta == "1" {
		// The join of the fkpl of the period (and of the search text): static fragments, every value
		// is an argument.
		hasGod, hasKar := s.fkplRepo.GetHasGodHasKar()
		fkplJoin := "left join fkpl f on f.tipanalitikeid = ta.tipanalitikeid"
		var joinArgs []any
		if hasGod {
			fkplJoin += " and f.god = ?"
			joinArgs = append(joinArgs, session.SelectedGod)
		}
		if hasKar {
			fkplJoin += " and f.kar = ?"
			joinArgs = append(joinArgs, session.SelectedKar)
		}
		if searchValue != "" {
			fkplJoin += " and (f.sifra ILIKE '%' || ? || '%' OR f.naziv ILIKE '%' || ? || '%')"
			joinArgs = append(joinArgs, searchValue, searchValue)
		}
		b = querybuilder.New(`SELECT distinct ta.tipanalitikeid, p.naziv as naziv, p.sifra as sifra, f.konto FROM partneri as p`).
			AddJoin("left join tipanalitike ta on ta.tipanalitikeid = p.tipanalitikeid").
			AddJoin(fkplJoin, joinArgs...)
		if err := s.partneriRepo.ApplyFilters(ctx, b, "p", nil, ""); err != nil {
			return err
		}
		if searchValue != "" {
			b.WhereAny(
				querybuilder.C("p.sifra ILIKE '%' || ? || '%'", searchValue),
				querybuilder.C("p.naziv ILIKE '%' || ? || '%'", searchValue),
			)
		}
		b.OrderBy("p.sifra")
	}
	if vkonta == "2" {
		b = querybuilder.New(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM fkpl as f`).
			Where("f.vkonta = ?", "2")
		if err := s.fkplRepo.ApplyFilters(ctx, b, "f", nil, ""); err != nil {
			return err
		}
		if searchValue != "" {
			b.WhereAny(
				querybuilder.C("f.konto ILIKE '%' || ? || '%'", searchValue),
				querybuilder.C("f.sifra ILIKE '%' || ? || '%'", searchValue),
				querybuilder.C("f.naziv ILIKE '%' || ? || '%'", searchValue),
			)
		}
		b.OrderBy("f.konto")
	}
	if b == nil {
		return errors.New("Invalid vkonta value")
	}

	b.Limit(20)

	entities, err := s.fkplRepo.Select(ctx, b)
	if err != nil {
		return errors.New(common.ErrMsgReadData)
	}

	// Prepare table data
	tbl.ShowActions = false
	tbl.ShowPagination = false
	tbl.FuncClick = "selectRow"
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)"

	tbl.DestField = fieldName
	tbl.DestFieldColIndex = fieldColIndex
	for _, entity := range *entities {
		fields := []string{}
		// Add common fields
		fields = append(fields,
			entity.Konto,
			entity.Sifra,
			entity.Naziv)
		tblRow := domain.TableRow{ID: fmt.Sprintf("%d", entity.IDFkpl), Fields: fields, HasUpdate: false, HasDelete: false}
		tbl.Rows = append(tbl.Rows, tblRow)
	}
	tbl.BtnAdd = domain.Button{IsVisible: false}
	tbl.BtnPrint = domain.Button{IsVisible: false}

	return nil
}

func (s *FkplResource) GetTableFileds() []domain.Fields {
	return s.fkplTableFields
}

func (s *FkplResource) setKontniPlanTableFields() {
	s.fkplTableFields = []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10", Sortable: true},
		{Name: "annnaziv", Label: "Naziv Analitike", Width: "10"},
		{Name: "naziv", Label: "Naziv konta", Width: "80", Sortable: true},
		{Name: "devizni", Label: "Devizno konto", Type: "checkbox", Width: "5", TextAlign: "center"},
		{Name: "kursirati", Label: "Kursirati konto", Type: "checkbox", Width: "5", TextAlign: "center"},
		{Name: "kolicinski", Label: "Kolicinski", Type: "checkbox", Width: "5", TextAlign: "center"},
		{Name: "vkonta", Label: "Vrsta konta", Width: "4", TextAlign: "center"},
	}

}
