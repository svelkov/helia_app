package finansijsko

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"helia/config"
	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/global"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	fproContentTitle    string = "FPRO"
	fproTableID         string = "fpro-table"
	fproURLPrefix       string = "/api/fpro/"
	fproURLGetAll       string = "/api/fpro/all/tipdok"
	fproURLGetAllSearch string = "/api/nalozi/stavke/:id"
	fproURLNextFpro     string = "/api/fpro/nextfpro"
)

// FproHandler handles requests related to "fpro" entities.
type FproHandler struct {
	fproService finservice.FproService // Use the interface
	btnSave     domain.Button
	btnNazad    domain.Button
	cfg         config.Config
	lm          *middleware.LockMiddleware
}

func NewFproHandler(service finservice.FproService, cfg config.Config, lm *middleware.LockMiddleware) *FproHandler {
	handler := &FproHandler{fproService: service, cfg: cfg, lm: lm}
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

// GetNalogStavke handles the request to get the stavke (items) of a nalog (order) for a given ID.
func (h *FproHandler) GetNalogStavke(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	idFnal, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	urlGetAll := fmt.Sprintf("/api/fpro/nalog/%d", idFnal)
	searchQuery := c.Query("query")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), urlGetAll, fmt.Sprintf("#%s", naloziStavkeTableID), "")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl := common.SetTableBasicData("STAVKE NALOGA", naloziStavkeTableID, h.fproService.GetTableStavkeFields(), urlGetAll, urlGetAll, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "", urlGetAll, false, false, false)

	if idFnal == 0 && searchQuery == "" {
		err := tmpl_fin.NaloziDetail(domain.TableData{}, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// Call the service to get ALL necessary data for the view
	fproPayload := domain.FproPayload{}
	ctx := c.Request.Context()
	err = h.fproService.GetAllFproByFnalID(ctx, &fproPayload, &tbl, idFnal, page, pageSize, searchQuery)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	tbl.URLGetAll = urlGetAll
	if requestSource == "btnpage" || requestSource == "searchinput" {
		utils.RenderContent(c, tbl)
		return
	}
	err = tmpl_fin.NaloziDetail(tbl, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
}

// SaveNalogStavke handles the request to save the stavke (items) of a nalog (order) for a given ID.
func (h *FproHandler) SaveNalogStavke(c *gin.Context) {
	var fproStavke domain.FproPayload
	fnalID, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}

	if err := c.ShouldBind(&fproStavke); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	
	fproStavke.IDFnal = fnalID
	filedErrors := h.fproService.FproValidate(c.Request.Context(), &fproStavke)
	if len(filedErrors) != 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, filedErrors, common.ErrMsgValidation)
		return
	}
	err = h.fproService.SaveNalogStavke(c.Request.Context(), &fproStavke)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData)
		return
	}
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	urlGetAll := fmt.Sprintf("/api/fpro/nalog/%d", fnalID)
	tblStavke := common.SetTableBasicData("STAVKE NALOGA", naloziStavkeTableID, h.fproService.GetTableStavkeFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.fproService.GetAllFproByFnalID(c.Request.Context(), &fproStavke, &tblStavke, fnalID, currentPage, pageSize, "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	tblStavke.URLGetAll = urlGetAll
	tblStavke.URLPrefix = urlGetAll
	tblStavke.SearchEnabled = false
	tblStavke.BtnAdd.IsVisible = false
	tblStavke.BtnPrint.IsVisible = false
	tblStavke.ShowActions = false
	tmpl.Table(tblStavke, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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

	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	r.POST("/api/fpro", h.CreateFpro)
	r.GET("/api/fpro/nalog/:id", h.GetNalogStavke)
	r.GET("/api/fpro/confirm-delete", h.confirmDeleteHandler)
	r.GET("/api/fpro/confirm-update", h.getFproStavkeHandler)
	r.GET("/api/fpro/confirm-add", h.confirmAddHandler)
	r.GET("/api/fpro/:id", h.GetFpro)
	r.PUT("/api/fpro/:id", h.UpdateFpro)
	r.DELETE("/api/fpro/:id", h.DeleteFpro)
	r.POST("/api/fpro/nalog/:id/stavke/save", h.SaveNalogStavke)
	r.PUT("/api/fpro/nalog/:id/stavke/save", h.SaveNalogStavke)

	// r.GET /api/fpro/prepis", h.FproPrepis))
	// r.GET /api/fpro/storniranje", h.FproStorniranje))
	// r.GET /api/fpro/prikaz", h.FproPrikaz))
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
