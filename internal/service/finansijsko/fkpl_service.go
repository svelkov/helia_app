package finansijsko

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

// FkplService defines the interface for operations related to Fkpl.
type FkplService interface {
	service.Service[domain.Fkpl]
	TraziKonto(ctx context.Context, konto, sifra, vkonta string) (entities *[]domain.Fkpl, err error)
	KontoSearchForTable(ctx context.Context, tbl *domain.TableData, searchValue, konto, vkonta, fieldName string, fieldColIndex int) error
}

// FkplResource implements the FkplService interface.
type FkplResource struct {
	service         *service.BaseService[domain.Fkpl]
	fkplRepo        repository.BaseRepository[domain.Fkpl]
	fkplTableFields []domain.Fields
	cfg             config.Config
}

func NewFkplResource(service *service.BaseService[domain.Fkpl], fkplRepo repository.BaseRepository[domain.Fkpl], cfg config.Config) *FkplResource {
	return &FkplResource{
		service:         service,
		fkplRepo:        fkplRepo,
		fkplTableFields: SetFkplTableFields(),
		cfg:             cfg,
	}
}

func (r *FkplResource) GetFieldCache() map[string]reflect.StructField {
	if r.service.GetFieldCache() == nil {
		r.service.SetFieldCache(make(map[string]reflect.StructField))
	}
	return r.service.GetFieldCache()
}

func (r *FkplResource) Create(ctx context.Context, entity *domain.Fkpl, idFkpl string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	return r.service.Create(ctx, entity, idFkpl, fields)
}
func (r *FkplResource) Update(ctx context.Context, entity *domain.Fkpl, idField string, idFiledValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	return r.service.Update(ctx, entity, idField, idFiledValue, fields)

}
func (r *FkplResource) Delete(ctx context.Context, idField string, idFieldValue int64) error {
	return r.service.Delete(ctx, idField, idFieldValue)
}
func (r *FkplResource) GetAll(ctx context.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchText, sortBy, sortOrder string) (*[]domain.Fkpl, error) {
	return r.service.GetAll(ctx, page, pageSize, tableFields, idField, searchText, sortBy, sortOrder)
}

// GetAllCustom implements FkplService.
func (r *FkplResource) GetAllCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fkpl, error) {
	return r.service.GetAllCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements FkplService.
func (r *FkplResource) GetByID(ctx context.Context, idField string, idValue int64) (*domain.Fkpl, error) {
	return r.service.GetByID(ctx, idField, idValue)
}

// GetTotalRecords implements FkplService.
func (r *FkplResource) GetTotalRecords(ctx context.Context, tableFields []domain.Fields, searchText string) (int, error) {
	return r.service.GetTotalRecords(ctx, tableFields, searchText)
}

// GetTotalRecordsCustom implements FkplService.
func (r *FkplResource) GetTotalRecordsCustom(ctx context.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return r.service.GetTotalRecordsCustom(ctx, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements FkplService.
func (r *FkplResource) MapEntityToValues(entity *domain.Fkpl, tableFields []domain.Fields) []domain.Fields {
	return r.service.MapEntityToValues(entity, tableFields)
}

func (r *FkplResource) TraziKonto(ctx context.Context, konto, sifra, vkonta string) (entities *[]domain.Fkpl, err error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}
	var qb *common.QueryBuilder
	hasGod, hasKar := r.fkplRepo.GetHasGodHasKar()
	// if vkonta == "1"  then search ibn table partneri for the tipanalitke for that konto and use that tipanalitikeid to search in fkpl, if vkonta == "2" then search directly in fkpl with vkonta = 2 and konto
	if vkonta == "1" {
		qb = common.NewQueryBuilder(`SELECT ta.tipanalitikeid, p.naziv as naziv, p.sifra as sifra FROM partneri as p `, true)
		qb.AddJoin("inner join tipanalitike ta on ta.tipanalitikeid = p.tipanalitikeid ")
		qb.AddJoin(" inner join fkpl f on f.tipanalitikeid = ta.tipanalitikeid ")
	}
	if vkonta == "2" {
		qb = common.NewQueryBuilder(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM fkpl as f`, true)
	}
	if hasGod {
		qb.AddEqual("f.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", session.SelectedKar)
	}

	if konto != "" {
		qb.AddEqual("f.konto", konto)
	}
	// only search in fkpl if vkonta is 2, if vkonta is 1 then search in partneri and join with fkpl through tipanalitikeid
	if vkonta == "2" {
		qb.AddEqual("f.vkonta", vkonta)
	}
	if vkonta == "1" && sifra != "" {
		qb.AddEqual("p.sifra", sifra)
	}

	sqlQuery, args := qb.Build()
	entities, err = r.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, errors.New(common.ErrMsgSearchKonto)
	}
	if len(*entities) == 0 {
		return nil, errors.New(common.ErrMsgNotFound)
	}
	return entities, nil
}
func (r *FkplResource) KontoSearchForTable(ctx context.Context, tbl *domain.TableData, searchValue, konto, vkonta, fieldName string, fieldColIndex int) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}
	var qb *common.QueryBuilder
	hasGod, hasKar := r.fkplRepo.GetHasGodHasKar()
	// if vkonta == "1"  then search ibn table partneri for the tipanalitke for that konto and use that tipanalitikeid to search in fkpl, if vkonta == "2" then search directly in fkpl with vkonta = 2 and konto
	if vkonta == "1" {
		qb = common.NewQueryBuilder(`SELECT ta.tipanalitikeid, p.naziv as naziv, p.sifra as sifra, f.konto FROM partneri as p `, true)
		qb.AddJoin("left join tipanalitike ta on ta.tipanalitikeid = p.tipanalitikeid ")
		qb.AddJoin(" left join fkpl f on f.tipanalitikeid = ta.tipanalitikeid ")
	}
	if vkonta == "2" {
		qb = common.NewQueryBuilder(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM fkpl as f`, true)
	}
	if hasGod {
		qb.AddEqual("f.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", session.SelectedKar)
	}
	if vkonta == "2" && konto != "" {
		qb.AddEqual("f.vkonta", vkonta)
		qb.AddEqual("f.konto", konto)
	}

	nbrArgs := qb.GetArgsCount()
	qb.AddCustomCondition(fmt.Sprintf(`(f.konto ILIKE '%%' || $%d || '%%' OR f.sifra ILIKE '%%' || $%d || '%%' OR f.naziv ILIKE '%%' || $%d || '%%' OR p.naziv ILIKE '%%' || $%d || '%%')`, nbrArgs+1, nbrArgs+1, nbrArgs+1, nbrArgs+1))
	qb.AddArgs(searchValue)
	if vkonta == "2" {
		qb.AddOrderBy("f.konto")
	}
	if vkonta == "1" {
		qb.AddOrderBy("p.sifra")
	}
	qb.SetLimit(20)

	sqlQuery, args := qb.Build()
	entities, err := r.fkplRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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

func (r FkplResource) GetTableFileds() []domain.Fields {
	return r.fkplTableFields
}
func SetFkplTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Sifra", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "120"},
		{Name: "vkonta", Label: "Vrsta konta", Width: "4"},
	}
}
