package finansijsko

import (
	"fmt"
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
	sqlQuery := `SELECT f.naziv
				FROM baza.fkpl as f`
	whereText := `WHERE 1 = 1  `
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "User session not found")
		return
	}

	hasGod, hasKAr := h.Service.Repo.GetHasGodHasKar()
	param := 1
	if hasGod {
		whereText += fmt.Sprintf(" AND f.god = $%d ", param)
		param++
		args = append(args, session.SelectedGod)
	}
	if hasKAr {
		whereText += fmt.Sprintf(" AND f.kar = $%d ", param)
		param++
		args = append(args, session.SelectedKar)
	}
	if konto != "" {
		whereText += fmt.Sprintf(" AND f.konto = $%d ", param)
		param++
		args = append(args, konto)
	}

	if sifra != "" {
		whereText += fmt.Sprintf(" AND f.sifra = $%d ", param)
		param++
		args = append(args, sifra)
	}
	if vkonta != "" {
		whereText += fmt.Sprintf(" AND f.vkonta = $%d", param)
		param++
		args = append(args, vkonta)
	}
	if vkonta == "1" && sifra == "" {
		c.Writer.Header().Set("Content-Type", "text/plain")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Nije pronađena šifra")
		return
	}
	entities, err := h.Service.GetAllCustom(c, sqlQuery, whereText, args, "", "")
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

	args := []interface{}{}
	// Parse query parameters from the URL
	searchValue := c.Query("query")
	konto := c.Query("konto")
	vkonta := c.Query("vkonta")
	if searchValue == "" {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Nedostaje parametar za pretrazivanje")
		return

	}
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "User session not found")
		return
	}

	whereKonto := ""
	// Custom SQL query for searching konto, sifra, or naziv
	args = append(args, session.SelectedGod, session.SelectedKar)
	sqlQuery := `SELECT idfkpl, f.konto, f.sifra, f.naziv
				FROM baza.fkpl as f`
	if konto != "" {
		whereKonto = ` AND konto = $3 AND f.vkonta = $4 AND (f.konto ILIKE '%' || $5 || '%' 
										OR f.sifra ILIKE '%' || $6 || '%' 
										OR f.naziv ILIKE '%' || $7 || '%' ) ORDER BY konto LIMIT 20 `
		args = append(args, konto)
	} else {
		whereKonto = ` AND f.vkonta = $3 AND (f.konto ILIKE '%' || $4 || '%' 
										OR f.sifra ILIKE '%' || $5 || '%' 
										OR f.naziv ILIKE '%' || $6 || '%' ) ORDER BY konto LIMIT 20 `

	}
	whereText := fmt.Sprintf(`WHERE f.god = $1 AND f.kar = $2  %s`, whereKonto)
	args = append(args, vkonta, searchValue, searchValue, searchValue)
	entities, err := h.Service.GetAllCustom(c, sqlQuery, whereText, args, "", "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	// Convert the fetched data into the format expected by the template
	tbl := common.SetTableBasicData("", searchKontoTableID, fkplSearchTableFields, "", "", 0, 0, 0, 0, h.cfg)
	tbl.ShowActions = false
	tbl.ShowPagination = false
	tbl.FuncClick = "selectRow"                             // naziv js function for Click
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)" // naziv js function for dblClick
	if vkonta == "2" {
		tbl.DestField = "konto"
	}
	if vkonta == "1" {
		tbl.DestField = "sifra"
	}

	// Prepare TableData for UI
	tblRows, err := common.SetTableRows(&tbl, *entities, fkplSearchTableFields, "idfkpl", "", h.Service.GetFieldCache())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to set table rows")
		return
	}
	tbl.Rows = tblRows.Rows
	tbl.BtnAdd = domain.Button{IsVisible: false}   // Hide Add button in this view
	tbl.BtnPrint = domain.Button{IsVisible: false} // Hide Print button in this view

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
