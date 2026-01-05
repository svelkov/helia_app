package finansijsko

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	tmpl "helia/frontend/templates/finansijsko"
	"helia/global"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/i18n"
	"helia/internal/middleware"
	"helia/internal/service"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	fproContentTitle    string = "FPRO"
	fproTableID         string = "fpro-table"
	fproURLPrefix       string = "/api/fpro/"
	fproURLGetAll       string = "/api/fpro/all/tipdok"
	fproURLGetAllSearch string = "/api/nalozi/stavke/{idfnalog}"
	fproURLNextFpro     string = "/api/fpro/nextfpro"
)

// FproHandler handles requests related to "fpro" entities.
type FproHandler struct {
	fproService service.FproService // Use the interface
	btnSave     domain.Button
	btnNazad    domain.Button
}

func NewFproHandler(service service.FproService) *FproHandler {
	handler := &FproHandler{fproService: service}
	handler.setHandlerFieldValues()
	return handler
}

func (h *FproHandler) GetFpro(c *gin.Context) {
	utils.GetEntityHelper(c, h.fproService, h.fproService.GetTableStavkeFields(), common.IDfpro)
}

func (h *FproHandler) CreateFpro(c *gin.Context) {
	var fpro domain.Fpro
	lastInsertedID, err := utils.CreateHelper(c, &fpro, h.fproService, common.IDfpro, h.fproService.GetTableStavkeFields())
	if err != nil {
		return
	}
	// Lock the header
	mutex := global.GetHeaderLock(lastInsertedID)
	mutex.Lock()
}

func (h *FproHandler) UpdateFpro(c *gin.Context) {
	var fpro domain.Fpro
	fproID := c.Query("id")
	// Lock the header
	mutex := global.GetHeaderLock(fproID)
	mutex.Lock()
	utils.UpdateHelper(c, &fpro, h.fproService, h.fproService.GetTableStavkeFields(), common.IDfpro)
}

func (h *FproHandler) DeleteFpro(c *gin.Context) {
	utils.DeleteHelper(c, h.fproService, common.IDfpro)
}

func (h *FproHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.fproService.GetTableStavkeFields())
}

func (h *FproHandler) confirmAddHandler(c *gin.Context) {
	utils.ConfirmAddHelper(c, strings.TrimSuffix(fproURLPrefix, "/"), h.fproService.GetTableStavkeFields())
}

func (h *FproHandler) getFproStavkeHandler(c *gin.Context) {
	utils.ConfirmUpdateHelper[domain.Fpro](c, nil, h.fproService.GetTableStavkeFields(), common.IDfpro)
}

// --- CONSOLIDATED VIEW DATA HANDLER ---
func (h *FproHandler) GetNalogStavke(c *gin.Context) {
	idFnalParam := c.PostForm("id")
	searchQuery := c.Query("query")
	searchControl := domain.InputControl{
		ID:           "search-controldetail",
		Label:        "Pretraži stavke naloga",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  fmt.Sprintf("/api/fpro/nalog/%s", idFnalParam),
		HxTarget:     "#nalozi_kopiranje_stavke",
		HxSwap:       "innerHTML",
		HxInclude:    "#search-controldetail",
		Class:        common.ClassSearchInput,
		Autocomplete: "off",
	}
	if idFnalParam == "" && searchQuery == "" {
		err := tmpl.NaloziDetail(domain.TableData{}, searchControl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}

	idFnal, err := strconv.ParseInt(idFnalParam, 10, 64)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	page, pageSize := common.GetPageAndPageSizeFromRequest(c)
	// Call the service to get ALL necessary data for the view
	tbl, err := h.fproService.GetNaloziStavke(c, idFnal, searchQuery, page, pageSize, h.fproService.GetTableStavkeFields())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	tbl.TableID = "nalozi_kopiranje_stavke"
	tbl.URLGetAll = "/api/fpro/nalog/" + idFnalParam

	err = tmpl.NaloziDetail(tbl, searchControl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
}

func (h *FproHandler) populateTableRows(tableData domain.TableData, entities []domain.Fpro, fieldsDef []domain.Fields) []domain.TableRow {
	var tableRows []domain.TableRow
	fieldCache := h.fproService.GetFieldCache()

	for _, entity := range entities {
		val := reflect.ValueOf(entity)
		idValue, _, found := common.GetFieldByNameCaseInsensitive(val, common.IDfpro)
		id := ""
		if found {
			id = fmt.Sprintf("%v", idValue)
		}
		var fields []string
		for _, fieldName := range fieldsDef {
			fieldInfo, found := fieldCache[strings.ToLower(fieldName.Name)]
			if !found {
				continue // or return error if field is required
			}

			value := common.GetFormattedValue(fieldInfo, val.FieldByName(fieldInfo.Name))
			fields = append(fields, value)
			fields = append(fields, "")
		}
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: tableData.BtnUpdate.IsVisible, HasDelete: tableData.BtnDelete.IsVisible}
		tableRows = append(tableRows, row)
	}
	return tableRows
}

/*
	 func (h *FproHandler) FproPrikaz(c *gin.Context) {
		err := tmpl_fin.FproContent(h.tabData, nil, domain.UkupnaObrada{}, h.btnSave, h.btnNoviFpro, domain.TableData{}).Render(r.Context(), w)
		if err != nil {
			response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.RenderTemplateErr, http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}
*/
func (h *FproHandler) AddRoutes(r *gin.Engine) {

	// Create API group with prefix
	api := r.Group(fproURLPrefix)
	api.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	api.POST("/api/fpro", h.CreateFpro)
	api.GET("/api/fpro/nalog/:idfnalog", h.GetNalogStavke)
	api.GET("/api/fpro/confirm-delete", h.confirmDeleteHandler)
	api.GET("/api/fpro/confirm-update", h.getFproStavkeHandler)
	api.GET("/api/fpro/confirm-add", h.confirmAddHandler)
	api.GET("/api/fpro/:id", h.GetFpro)
	api.PUT("/api/fpro/:id", h.UpdateFpro)
	api.DELETE(" /api/fpro/:id", h.DeleteFpro)
	// api.GET /api/fpro/prepis", h.FproPrepis))
	// api.GET /api/fpro/storniranje", h.FproStorniranje))
	// api.GET /api/fpro/prikaz", h.FproPrikaz))
}

func (h *FproHandler) setHandlerFieldValues() {
	h.btnSave = domain.Button{
		Id:            "btn-save",
		LabelText:     "Snimi Fpro",
		HxActionURL:   fproURLPrefix,
		HxTarget:      "this",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
	}
	h.btnNazad = domain.Button{
		Id:            "btn-nazad-fpro",
		LabelText:     "Nazad",
		HxActionURL:   fproURLNextFpro,
		HxRequestType: "GET",
	}
}
