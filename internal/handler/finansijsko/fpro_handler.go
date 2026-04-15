package finansijsko

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"helia/config"
	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	fproContentTitle string = "FPRO"
	fproTableID      string = "fpro-table"
	fproURLPrefix    string = "/api/fpro/"
	fproURLGetAll    string = "/api/fpro/nalog/:id"
	fproURLStavka    string = "/api/fpro/stavka"
	fproURLNextFpro  string = "/api/fpro/nextfpro"
	fproURLDelete    string = "/api/fpro/confirm-delete"
	fproURLUpdate    string = "/api/fpro/stavka/update"
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

func (h *FproHandler) DeleteFpro(c *gin.Context) {
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

	tblStavke := common.SetTableBasicData("STAVKE NALOGA", naloziStavkeTableID, h.fproService.GetTableStavkeFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.fproService.DeleteFpro(c.Request.Context(), id, &tblStavke, currentPage, pageSize)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgDeleteData+": "+err.Error())
		return
	}
	c.Header("HX-Trigger", "fpro:changed")
	common.WriteJSONResponse(c, http.StatusOK, true, nil, "Uspešno obrisani podaci...")
}

func (h *FproHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.fproService.GetTableStavkeFields(), "#dialog-delete-message")
}

func (h *FproHandler) confirmAddHandler(c *gin.Context) {
	utils.ConfirmAddHelper(c, strings.TrimSuffix(fproURLPrefix, "/"), h.fproService.GetTableStavkeFields(), "#dialog-form-message")
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

	searchQuery := c.Query("query")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), fproURLGetAll, fmt.Sprintf("#%s", naloziStavkeTableID), "")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tblStavke := common.SetTableBasicData("STAVKE NALOGA", naloziStavkeTableID, h.fproService.GetTableStavkeFields(), fproURLGetAll, fproURLGetAll, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblStavke, "", fproURLStavka, false, false, false)
	tblStavke.BtnDelete.HxActionURL = fproURLDelete
	tblStavke.BtnUpdate.HxActionURL = fproURLUpdate
	tblStavke.BtnUpdate.HxOnAfterRequest = "populateFproUpdateFormFromEvent(event)"
	tblStavke.BtnUpdate.HxSwap = "none"
	tblStavke.DetailURL = fproURLGetAll

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
	err = h.fproService.GetAllFproByFnalID(ctx, &fproPayload, &tblStavke, idFnal, page, pageSize, searchQuery)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	//tblStavke.URLGetAll = fproURLGetAll
	if requestSource == "btnpage" || requestSource == "searchinput" {
		tmpl.Table(tblStavke, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		return
	}
	err = tmpl_fin.NaloziDetail(tblStavke, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tblStavke := common.SetTableBasicData("STAVKE NALOGA", naloziStavkeTableID, h.fproService.GetTableStavkeFields(), "", "", 0, 0, 0, 0, h.cfg)
	tblStavke.URLGetAll = fproURLDelete
	tblStavke.URLPrefix = fproURLDelete
	tblStavke.BtnDelete.HxActionURL = fproURLDelete
	tblStavke.BtnUpdate.HxActionURL = fproURLUpdate
	tblStavke.BtnUpdate.HxOnAfterRequest = "populateFproUpdateFormFromEvent(event)"
	tblStavke.BtnUpdate.HxSwap = "none"
	err = h.fproService.GetAllFproByFnalID(c.Request.Context(), &fproStavke, &tblStavke, fnalID, currentPage, pageSize, "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	tblStavke.SearchEnabled = true
	tblStavke.BtnAdd.IsVisible = false
	tblStavke.BtnPrint.IsVisible = false
	tblStavke.ShowActions = true
	c.Header("HX-Trigger", "fpro:changed")
	common.WriteJSONResponse(c, http.StatusOK, true, nil, "Uspešno upisani podaci...")

	//tmpl.Table(tblStavke, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *FproHandler) UpdateFproStavke(c *gin.Context) {
	idFpro, err := utils.GetInt64FromQueryRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}

	fpro, err := h.fproService.GetByID(c.Request.Context(), common.IDfpro, idFpro)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	// return JSON response with fpro data for populating the update form
	c.JSON(http.StatusOK, setUpdateValues(fpro))
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

func (h *FproHandler) GetNalogTotalValues(c *gin.Context) {
	idFnal, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	urlGetAll := fmt.Sprintf("/api/fpro/nalog/total/%d", idFnal)
	nalogTotal := domain.NalogTotalValues{HxGetURL: urlGetAll}
	err = h.fproService.GetNalogTotalValues(c.Request.Context(), &nalogTotal, idFnal)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	tmpl_fin.NalogTotalValues(nalogTotal, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func setUpdateValues(fpro *domain.Fpro) map[string]interface{} {
	result := map[string]interface{}{
		"idfpro":     fmt.Sprintf("%d", fpro.IDFpro),
		"konto":      fpro.Konto,
		"kontonaziv": fpro.NazivKonta,
		"sifra":      fpro.Sifra,
		"sifranaziv": fpro.NazivAnalitike,
		"vrd":        fpro.Vrd,
		"opisknj":    fpro.Opis,
		"magaciniid": fpro.IDMagacini.Int16,
		"fispid":     fpro.IDFisp.Int16,
		"komid":      fpro.IDKom.Int16,
		"idorgjed":   fpro.IDOrgjed.Int64,
		"mestrotrid": fpro.IDMestotr.Int16,
		"dokum":      fpro.Dokum,
		"dadok":      fpro.Dadok.Time.Format(common.HtmlLayout),
		"tra":        fpro.Tra,
		"dokumv":     fpro.Dokumv.String,
		"travez":     fpro.Travez.Int16,
		"datdokv":    fpro.Dadokv.Time.Format(common.HtmlLayout),
		"sifval":     fpro.Sifval.Int16,
		"deviznos":   fmt.Sprintf("%.2f", fpro.Deviznos),
		"kurs":       fmt.Sprintf("%.4f", fpro.Kurs),
		"Iznos":      fmt.Sprintf("%.2f", fpro.Iznos),
		"IDFnal":     fmt.Sprintf("%d", fpro.IDFnal),
	}

	if fpro.Kat == 1 || fpro.Kat == 2 {
		result["duguje"] = fmt.Sprintf("%.2f", fpro.Iznos)
	}
	if fpro.Kat == 3 || fpro.Kat == 4 {
		result["potrazuje"] = fmt.Sprintf("%.2f", fpro.Iznos)
	}
	return result
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
	r.GET("/api/fpro/nalog/:id", h.GetNalogStavke)
	r.GET("/api/fpro/nalog/total/:id", h.GetNalogTotalValues)
	r.GET("/api/fpro/confirm-delete", h.confirmDeleteHandler)
	r.GET("/api/fpro/confirm-update", h.getFproStavkeHandler)
	r.GET("/api/fpro/confirm-add", h.confirmAddHandler)
	r.GET("/api/fpro/:id", h.GetFpro)
	r.DELETE("/api/fpro/stavka/:id", h.DeleteFpro)
	r.POST("/api/fpro/nalog/:id/stavke/save", h.SaveNalogStavke)
	r.PUT("/api/fpro/nalog/:id/stavke/save", h.SaveNalogStavke)
	r.GET("/api/fpro/stavka/update/:id", h.UpdateFproStavke)

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
