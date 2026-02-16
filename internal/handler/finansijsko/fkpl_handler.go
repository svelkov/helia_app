package finansijsko

import (
	"net/http"
	"strings"

	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	fkplContentTitle   string = "KONTNI PLAN"
	fkplTableID        string = "fkpl-table"
	searchKontoTableID string = "searchkonto-table"
	fkplURLPrefix      string = "/api/fkpl/"
	fkplURLGetAll      string = "/api/fkpl/all"
)

func SetFkplFields() []domain.Fields {
	return []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Sifra", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "120"},
		{Name: "vkonta", Label: "Vrsta konta", Width: "4"},
	}
}

var fkplSearchTableFields = []domain.Fields{
	{Name: "konto", Label: "Konto", Width: "10"},
	{Name: "sifra", Label: "Sifra", Width: "10"},
	{Name: "naziv", Label: "Naziv", Width: "120"},
}

type FkplHandler struct {
	Service *service.BaseService[domain.Fkpl]
	cfg     config.Config
}

func NewFkplHandler(service *service.BaseService[domain.Fkpl], cfg config.Config) *FkplHandler {
	return &FkplHandler{Service: service, cfg: cfg}
}

func (h *FkplHandler) CreateFkpl(c *gin.Context) {
	var fkpl domain.Fkpl
	utils.CreateHelper(c, &fkpl, h.Service, common.IDfkpl, SetFkplFields())
}

func (h *FkplHandler) UpdateFkpl(c *gin.Context) {
	var fkpl domain.Fkpl
	utils.UpdateHelper(c, &fkpl, h.Service, SetFkplFields(), common.IDfkpl)
}

func (h *FkplHandler) DeleteFkpl(c *gin.Context) {
	utils.DeleteHelper(c, h.Service, common.IDfkpl)
}

func (h *FkplHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, SetFkplFields())
}

func (h *FkplHandler) confirmAddHandler(c *gin.Context) {
	utils.ConfirmAddHelper(c, strings.TrimSuffix(fkplURLPrefix, "/"), SetFkplFields())
}

func (h *FkplHandler) confirmUpdateHandler(c *gin.Context) {
	utils.ConfirmUpdateHelper(c, h.Service, SetFkplFields(), common.IDfkpl)
}

func (h *FkplHandler) GetFkpl(c *gin.Context) {
	utils.GetEntityHelper(c, h.Service, SetFkplFields(), common.IDfkpl)
}

func (h *FkplHandler) GetAllFkpl(c *gin.Context) {

	tbl := utils.GetAllEntityHelper(c, h.Service, SetFkplFields(), fkplContentTitle, fkplTableID, fkplURLPrefix, fkplURLGetAll, common.IDfkpl, h.cfg)
	utils.RenderContent(c, *tbl)
}

func (h *FkplHandler) TraziKonto(c *gin.Context) {
	args := []interface{}{}
	// Parse query parameters from the URL

	konto := c.Query("konto")
	sifra := c.Query("sifra")
	vkonta := c.Query("vkonta")
	if vkonta == "" && konto == "" && sifra == "" {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Nedostaje parametar konto ili vkonta")
		return
	}

	// Custom SQL query for searching konto, sifra, or naziv
	qb := common.NewQueryBuilder(`SELECT f.naziv FROM baza.fkpl as f`)

	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "User session not found")
		return
	}

	hasGod, hasKar := h.Service.Repo.GetHasGodHasKar()
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
		c.Writer.Header().Set("Content-Type", "text/plain")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Nije pronađena šifra")
		return
	}

	sqlQuery, args := qb.Build()
	entities, err := h.Service.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Greška prilikom pretrage konta")
		return
	}
	// Check if pointer is not nil and slice is not empty
	if entities != nil && len(*entities) > 0 {
		firstElement := (*entities)[0].Naziv
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.Write([]byte(firstElement))
		return
	}
	c.Writer.Header().Set("Content-Type", "text/plain")
	if vkonta == "2" {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Nije pronađen konto")
		return
	}
	if vkonta == "1" {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Nije pronađena šifra")
		return
	}

}
func (h *FkplHandler) TraziKontoSearchTable(c *gin.Context) {
	// Parse query parameters
	searchValue := c.Query("query")
	konto := c.Query("konto")
	vkonta := c.Query("vkonta")
	fieldName := c.DefaultQuery("fieldName", "konto")

	if searchValue == "" {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, "Nedostaje parametar za pretrazivanje")
		return
	}

	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "User session not found")
		return
	}

	// Build query
	qb := common.NewQueryBuilder(`SELECT f.idfkpl, f.konto, f.sifra, f.naziv FROM baza.fkpl as f`)

	hasGod, hasKar := h.Service.Repo.GetHasGodHasKar()
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
	qb.AddCustomCondition(`(f.konto ILIKE '%' || $%d || '%' OR f.sifra ILIKE '%' || $%d || '%' OR f.naziv ILIKE '%' || $%d || '%')`, nbrArgs+1, nbrArgs+2, nbrArgs+3)
	qb.AddCustomCondition("ORDER BY f.konto LIMIT 20")

	sqlQuery, args := qb.Build()
	entities, err := h.Service.GetAllCustom(c, sqlQuery, "", args, "", "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}

	// Prepare table data
	tbl := common.SetTableBasicData("", searchKontoTableID, fkplSearchTableFields, "", "", 0, 0, 0, 0, h.cfg)
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

	tblRows, err := common.SetTableRows(&tbl, *entities, fkplSearchTableFields, "idfkpl", "", h.Service.GetFieldCache())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to set table rows")
		return
	}

	tbl.Rows = tblRows.Rows
	tbl.BtnAdd = domain.Button{IsVisible: false}
	tbl.BtnPrint = domain.Button{IsVisible: false}

	utils.RenderContent(c, tbl)
}
func (h *FkplHandler) AddRoutes(r *gin.Engine) {
	// Create API group with prefix
	//api := r.Group(fkplURLPrefix)
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for fkpl
	r.POST("/api/fkpl", h.CreateFkpl)
	r.GET("/api/fkpl/all", h.GetAllFkpl)
	r.GET("/api/fkpl/confirm-delete", h.confirmDeleteHandler)
	r.GET("/api/fkpl/confirm-update", h.confirmUpdateHandler)
	r.GET("/api/fkpl/confirm-add", h.confirmAddHandler)
	r.GET("/api/fkpl/:id", h.GetFkpl)
	r.PUT("/api/fkpl/:id", h.UpdateFkpl)
	r.DELETE("/api/fkpl/:id", h.DeleteFkpl)
	r.GET("/api/fkpl/trazikonto", h.TraziKonto)
	r.GET("/api/fkpl/trazikontosearchtable", h.TraziKontoSearchTable)
}
