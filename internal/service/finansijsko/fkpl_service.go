package finansijsko

import (
	"errors"
	"fmt"
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/repository"
	"helia/internal/service"
	"reflect"

	"github.com/gin-gonic/gin"
)

// FkplService defines the interface for operations related to Fkpl.
type FkplService interface {
	service.Service[domain.Fkpl]
	TraziKonto(c *gin.Context) (entities *[]domain.Fkpl, err error)
	KontoSearchForTable(c *gin.Context, tbl *domain.TableData) error
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

func (r *FkplResource) Create(c *gin.Context, entity *domain.Fkpl, idFkpl string, fields []domain.Fields) ([]domain.FieldError, int64, error) {
	return r.service.Create(c, entity, idFkpl, fields)
}
func (r *FkplResource) Update(c *gin.Context, entity *domain.Fkpl, idField string, idFiledValue interface{}, fields []domain.Fields) ([]domain.FieldError, error) {
	return r.service.Update(c, entity, idField, idFiledValue, fields)

}
func (r *FkplResource) Delete(idField string, idFieldValue int64) error {
	return r.service.Delete(idField, idFieldValue)
}
func (r *FkplResource) GetAll(c *gin.Context, page int, pageSize int, tableFields []domain.Fields, idField string, searchParams ...string) (*[]domain.Fkpl, error) {
	return r.service.GetAll(c, page, pageSize, tableFields, idField, searchParams...)
}

// GetAllCustom implements FkplService.
func (r *FkplResource) GetAllCustom(c *gin.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (*[]domain.Fkpl, error) {
	return r.service.GetAllCustom(c, queryText, whereText, args, limitOffset, orderBy)
}

// GetByID implements FkplService.
func (r *FkplResource) GetByID(idField string, idValue int64) (*domain.Fkpl, error) {
	return r.service.GetByID(idField, idValue)
}

// GetTotalRecords implements FkplService.
func (r *FkplResource) GetTotalRecords(c *gin.Context, tableFields []domain.Fields, searchParams ...string) (int, error) {
	return r.service.GetTotalRecords(c, tableFields, searchParams...)
}

// GetTotalRecordsCustom implements FkplService.
func (r *FkplResource) GetTotalRecordsCustom(c *gin.Context, queryText string, whereText string, args []interface{}, limitOffset string, orderBy string) (int, error) {
	return r.service.GetTotalRecordsCustom(c, queryText, whereText, args, limitOffset, orderBy)
}

// MapEntityToValues implements FkplService.
func (r *FkplResource) MapEntityToValues(entity *domain.Fkpl, tableFields []domain.Fields) []domain.Fields {
	return r.service.MapEntityToValues(entity, tableFields)
}

func (r *FkplResource) TraziKonto(c *gin.Context) (entities *[]domain.Fkpl, err error) {

	konto := c.Query("konto")
	sifra := c.Query("sifra")
	vkonta := c.Query("vkonta")
	if vkonta == "" && konto == "" && sifra == "" {
		return nil, errors.New(common.ErrMsgMissingParameter)
	}

	// Custom SQL query for searching konto, sifra, or naziv
	qb := common.NewQueryBuilder(`SELECT f.naziv FROM baza.fkpl as f`)

	session := domain.GetSessionFromContext(c)
	if session == nil {
		return nil, errors.New(common.ErrMsgUserSessionNotFound)
	}

	hasGod, hasKar := r.fkplRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	if konto != "" {
		qb.AddEqual("konto", konto)
	}
	if sifra != "" {
		qb.AddEqual("sifra", sifra)
	}
	if vkonta != "" {
		qb.AddEqual("vkonta", vkonta)
	}

	if vkonta == "1" && sifra == "" {
		return nil, errors.New(common.ErrMsgMissingParameter)
	}

	sqlQuery, args := qb.Build()
	entities, err = r.fkplRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return nil, errors.New(common.ErrMsgSearchKonto)
	}
	if len(*entities) == 0 {
		return nil, errors.New(common.ErrMsgNotFound)
	}
	return entities, nil
}
func (r *FkplResource) KontoSearchForTable(c *gin.Context, tbl *domain.TableData) error {
	searchValue := c.Query("query")
	konto := c.Query("konto")
	vkonta := c.Query("vkonta")
	fieldName := c.DefaultQuery("fieldName", "konto")

	if searchValue == "" {
		return errors.New(common.ErrMsgMissingParameter)
	}

	session := domain.GetSessionFromContext(c)
	if session == nil {
		return errors.New(common.ErrMsgUserSessionNotFound)
	}

	// Build query
	qb := common.NewQueryBuilder(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM baza.fkpl as f`)

	hasGod, hasKar := r.fkplRepo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("f.god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("f.kar", session.SelectedKar)
	}

	if konto != "" {
		qb.AddEqual("f.konto", konto)
	}

	if vkonta != "" {
		qb.AddEqual("f.vkonta", vkonta)
	}

	nbrArgs := qb.GetArgsCount()
	qb.AddCustomCondition(fmt.Sprintf(`(f.konto ILIKE '%%' || $%d || '%%' OR f.sifra ILIKE '%%' || $%d || '%%' OR f.naziv ILIKE '%%' || $%d || '%%')`, nbrArgs+1, nbrArgs+2, nbrArgs+3))
	qb.AddArgs(searchValue, searchValue, searchValue)
	qb.AddOrderBy("f.konto")
	qb.SetLimit(20)

	sqlQuery, args := qb.Build()
	entities, err := r.fkplRepo.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		return errors.New(common.ErrMsgReadData)
	}

	// Prepare table data
	tbl.ShowActions = false
	tbl.ShowPagination = false
	tbl.FuncClick = "selectRow"
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)"

	// Determine destination field based on vkonta or fieldName
	destFieldMap := map[string]string{
		"2": "konto",
		"1": "sifra",
	}
	if destField, exists := destFieldMap[vkonta]; exists {
		tbl.DestField = destField
	} else {
		tbl.DestField = fieldName
	}
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
