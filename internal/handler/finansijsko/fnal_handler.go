package finansijsko

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"helia/config"
	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	finservice "helia/internal/service/finansijsko"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	naloziContentTitle     string = "NALOZI"
	naloziTableID          string = "nalozi-table"
	naloziStavkeTableID    string = "nalog-stavke-table"
	naloziURLPrefix        string = "/api/nalozi/"
	naloziURLGetAllSearch  string = "/api/nalozi/all/search"
	naloziURLNextNalog     string = "/api/nalozi/nextnalog"
	naloziURLCreate        string = "/api/nalozi/new"
	naloziURLUpdate        string = "/api/nalozi/update"
	naloziURLUpdateCopy    string = "/api/nalozi/update/copy"
	naloziURLSaveStavke    string = "/api/fpro/nalog/%d/stavke/save"
	naloziURLPrintStavke   string = "/api/print/nalog/stavke"
	naloziURLStorniraj     string = "/api/nalozi/storniraj"
	naloziURLStampa        string = "/api/nalozi/prikaz"
	naloziURLStampaDetalji string = "/api/nalozi/prikaz/detalji"
	naloziURLStavkeNaloga  string = "/api/fpro/nalog/"

	hxValsKopirajNalog = `js:{
			"query": document.getElementById("search-prepishdr")?.value, 
			"kopirajceonalog": document.getElementById("kopirajceonalog")?.checked ? "1" : "0"
		}`
	hxValsKnjizenje = `js:{
            "tipdok": document.getElementById("tipdok")?.value,
            "query": document.getElementById("search-input")?.value
        }`
	hxValsStampa = `js:{
			"query": document.getElementById("search-control")?.value,	
		}`
	hxValsStampaDetalji = `js:{
			"idfnal": document.getElementById("idfnal")?.value,
			"query": document.getElementById("search-stampa-detalji")?.value,
		}`
)

// FnalHandler handles requests related to "fnal" entities.
type FnalHandler struct {
	naloziService finservice.NalogService
	service       service.BaseService[domain.Fnal]
	tabData       domain.TabData
	btnSave       domain.Button
	btnNoviNalog  domain.Button
	typeView      string // Type of view, e.g. "knjizenje", "prepis", etc.
	cfg           config.Config
	lm            *middleware.LockMiddleware
	ls            *middleware.LockService
}

func NewFnalHandler(nalogService finservice.NalogService, baseService *service.BaseService[domain.Fnal], cfg config.Config, lm *middleware.LockMiddleware, ls *middleware.LockService) *FnalHandler {
	handler := &FnalHandler{
		naloziService: nalogService,
		service:       *baseService,
		cfg:           cfg,
		lm:            lm,
		ls:            ls,
	}
	handler.setHandlerFieldValues()
	return handler
}

func (h *FnalHandler) GetNalog(c *gin.Context) {
	utils.GetEntityHelper(c, h.naloziService, h.naloziService.GetNaloziTableFields(), common.IDfnal)
}

// GetNextNalog returns the next available Nalog number for a given tipdok.
func (h *FnalHandler) GetNextNalog(c *gin.Context) {
	tipdok := c.Query("tipdok")

	ctx := c.Request.Context()
	nextNalog, err := h.naloziService.GetNextNalog(ctx, tipdok)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get next Nalog")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"nalog":   nextNalog,
	})
}

// GetUpdateDataNalog retrieves data needed for updating a Nalog.
func (h *FnalHandler) GetUpdateDataNalog(c *gin.Context) {
	tipdok := c.Query("tipdok")
	nalog := c.Query("nalog")
	nalogNbr := common.StringToInt64(nalog)

	ctx := c.Request.Context()
	nalogEntity, err := h.naloziService.GetByTipdokNalog(ctx, tipdok, int64(nalogNbr))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get next Nalog")
		return
	}
	userSesion := domain.GetSessionFromContext(c)
	if nalogEntity.Danal.IsZero() {
		now := time.Now()
		nalogEntity.Danal = time.Date(userSesion.SelectedGod, now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	}
	if nalogEntity.Datob.IsZero() {
		now := time.Now()
		nalogEntity.Datob = time.Date(userSesion.SelectedGod, now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"datob":   nalogEntity.Datob.Format(common.HtmlLayout),
		"danal":   nalogEntity.Danal.Format(common.HtmlLayout),
		"opis":    nalogEntity.Opis,
	})
}

func (h *FnalHandler) CreateNalog(c *gin.Context) {
	var entity domain.Fnal
	var req domain.FnalPayload

	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return
	}
	if err := c.ShouldBind(&req); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}

	// Check if nalog already exists - if so, redirect to update
	existingNalog, err := h.naloziService.GetByTipdokNalog(ctx, req.Tipdok, common.StringToInt64(req.Nalog))
	if err == nil && existingNalog.IDFnal != 0 {
		log.Printf("Nalog already exists (IDFnal=%d), redirecting to update instead of create", existingNalog.IDFnal)
		c.Request.Method = "PUT"
		c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%d", existingNalog.IDFnal)})
		h.UpdateNalog(c)
		return
	}
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

	h.naloziService.MapReqToEntity(ctx, req, &entity, common.ActionAdd)
	fproPayload, fieldsErrors, lastInsertedID, err := h.naloziService.CreateNalog(ctx, &entity, common.IDfnal, h.naloziService.GetNaloziTableFields(), currentPage, pageSize, "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsErrors, common.ErrMsgSaveData+" error:"+err.Error())
		return
	}
	if len(fieldsErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldsErrors, common.ErrMsgValidation)
		return
	}

	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), fmt.Sprintf("%s%d", naloziURLStavkeNaloga, lastInsertedID), fmt.Sprintf("#%s", naloziStavkeTableID), "")
	tblStavke := common.SetTableBasicData("Stavke Naloga", naloziStavkeTableID, h.service.MapEntityToValues(&entity, h.naloziService.GetNaloziTableFields()), "", "", 10, 0, 0, 0, h.cfg)
	btnSave, btnPrint := setStavkeButtons("POST", lastInsertedID)
	btnClose := domain.Button{
		Id:            "btn-close",
		IsVisible:     true,
		IdDialog:      "nalog-stavke-dialog",
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   fmt.Sprintf("/api/nalozi/unlock/%d", lastInsertedID),
		HxRequestType: "POST",
	}
	btnNazad := domain.Button{
		Id:               "btn-nazad",
		IsVisible:        true,
		IdDialog:         "nalog-stavke-dialog",
		LabelText:        "Back",
		Icon:             "back",
		BtnClass:         common.ClassCloseButton,
		HxActionURL:      fmt.Sprintf("/api/nalozi/unlock/%d", lastInsertedID),
		HxRequestType:    "POST",
		HxOnAfterRequest: "closeDialog",
	}
	// Lock the header for the newly created nalog
	err = h.ls.Lock(c, "fnal", lastInsertedID, userSession.UserName)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgLockFailed)
		return
	}
	err = tmpl_fin.NalogKnjizenjeStavke(*fproPayload, tblStavke, btnSave, btnPrint, btnClose, btnNazad, searchInput, i18n.GetInstance(), common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
}

// UpdateNalog handles the update of an existing "fnal" entity.
func (h *FnalHandler) UpdateNalog(c *gin.Context) {
	ctx := c.Request.Context()
	var tblStavke domain.TableData
	idParam := c.Param("id")
	searchText := c.Query("query")
	fnalID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}
	urlGetAll := fmt.Sprintf("/api/fpro/nalog/%d", fnalID)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), urlGetAll, fmt.Sprintf("#%s", naloziStavkeTableID), "")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	var payload domain.FnalPayload
	if err := c.ShouldBind(&payload); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	err = h.lm.VerifyLock(c, "fnal", fnalID)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusConflict, false, []domain.FieldError{}, common.ErrMsgStatusConflict)
		return
	}
	fproPayload, fieldErrors, err := h.naloziService.UpdateNalog(ctx, fnalID, &payload, &tblStavke, page, pageSize, searchText)
	if err != nil {
		if strings.Contains(err.Error(), common.ErrMsgUpdate) {
			common.WriteJSONResponse(c, http.StatusConflict, false, []domain.FieldError{}, common.ErrMsgUpdate)
			return
		}
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldErrors, common.ErrMsgReadData+" error:"+err.Error())
		return
	}
	btnSave, btnPrint := setStavkeButtons("POST", fnalID)
	btnClose := domain.Button{
		Id:            "btn-close",
		IsVisible:     true,
		IdDialog:      "nalog-stavke-dialog",
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   fmt.Sprintf("/api/nalozi/unlock/%d", fnalID),
		HxRequestType: "POST",
		HxOn:          "closeDialog('nalog-stavke-dialog')",
	}
	btnNazad := domain.Button{
		Id:               "btn-nazad",
		IsVisible:        true,
		IdDialog:         "nalog-stavke-dialog",
		LabelText:        "Back",
		Icon:             "back",
		BtnClass:         common.ClassCloseButton,
		HxActionURL:      fmt.Sprintf("/api/nalozi/unlock/%d", fnalID),
		HxRequestType:    "POST",
		HxOnAfterRequest: "closeDialog",
	}
	tblStavke.URLGetAll = urlGetAll
	tblStavke.URLPrefix = urlGetAll
	err = tmpl_fin.NalogKnjizenjeStavke(fproPayload, tblStavke, btnSave, btnPrint, btnClose, btnNazad, searchInput, i18n.GetInstance(), common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
}

func (h *FnalHandler) DeleteNalog(c *gin.Context) {
	utils.DeleteHelper(c, h.naloziService, common.IDfnal)
}

// UnlockNalog releases the lock on a nalog
func (h *FnalHandler) UnlockNalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Nalog unlocked..."})
}
func (h *FnalHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.naloziService.GetNaloziTableFields())
}

func (h *FnalHandler) confirmAddHandler(c *gin.Context) {
	var entity domain.Fnal
	var req domain.FnalPayload
	var action string = common.ActionAdd
	ctx := c.Request.Context()

	if err := c.ShouldBind(&req); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	// map request to entity
	h.naloziService.MapReqToEntity(ctx, req, &entity, action)

	nalogEntity, err := h.naloziService.GetByTipdokNalog(ctx, entity.Tipdok, entity.Nalog)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}

	// If nalog exists, check if it's locked by someone else (but don't acquire yet)
	if nalogEntity.IDFnal != 0 {
		err = h.lm.VerifyLock(c, "fnal", nalogEntity.IDFnal)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusConflict, false, []domain.FieldError{}, common.ErrMsgStatusConflict)
			return
		}
	}
	dlgTitle := ""
	hxVals := ""
	translator := i18n.GetInstance()
	msg := []string{}
	hxVals = fmt.Sprintf(`{"tipdok": "%s", "nalog": "%s", "danal": "%s", "datob": "%s", "opis": "%s"}`, req.Tipdok, req.Nalog, req.Danal, req.Datob, req.Opis)
	if nalogEntity.IDFnal == 0 {
		dlgTitle = "Novi Nalog"
		msg = append(msg, translator.Message(`Otvaranje novog naloga?`))
	} else {
		action = common.ActionUpdate
		dlgTitle = "Nastavak knjiženja naloga"
		msg = append(msg, translator.Message(`Nastavak knjiženja naloga?`))
		msg = append(msg, fmt.Sprintf("%s: %s", translator.Message(`Vrsta naloga`), req.Tipdok))
		msg = append(msg, fmt.Sprintf("%s: %s", translator.Message(`Broj naloga`), req.Nalog))
	}
	fieldErrors, err := h.naloziService.NalogValidation(ctx, entity, action)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}

	dialog := domain.Dialog{
		Title:         dlgTitle,
		OkText:        "Da",
		CancelText:    "Ne",
		SaveText:      "Da",
		HxTarget:      "#dialog-stavkenaloga",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}

	btnSacuvaj := domain.Button{
		Id:        "btn-sacuvaj",
		IsVisible: true,
		LabelText: "Da",
		HxVals:    hxVals,
		IdDialog:  dialog.Id,
		HxTarget:  "#dialog-stavkenaloga",
		BtnClass:  common.ClassSaveButton,
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Ne",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	if action == common.ActionAdd {
		dialog.Id = "dlg_nalog_create"
		btnSacuvaj.HxRequestType = "POST"
		btnSacuvaj.HxActionURL = naloziURLCreate
	}
	if action == common.ActionUpdate {
		dialog.Id = "dlg_nalog_update"
		btnSacuvaj.HxRequestType = "PUT"
		btnSacuvaj.HxActionURL = fmt.Sprintf("%s/%d", naloziURLUpdate, nalogEntity.IDFnal)
	}
	btnCancel.IdDialog = dialog.Id
	btnClose.IdDialog = dialog.Id
	tmpl.DialogConfirm(msg, dialog, btnClose, btnSacuvaj, btnCancel, i18n.GetInstance(), common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

func (h *FnalHandler) GetIdTipdokByTipdok(c *gin.Context, tipdok string) (int64, error) {
	return h.naloziService.GetIdTipdokByTipdok(c, tipdok)
}

// --- CONSOLIDATED VIEW DATA HANDLER ---
func (h *FnalHandler) GetNalogMainView(c *gin.Context) {
	searchQuery := c.Query("query")
	selectedTipdok := c.Query("tipdok")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	// Determine if this is an initial full page request (not an HTMX swap)
	isInitialRequest := selectedTipdok == ""
	if isInitialRequest {
		//h.naloziService.SetSfTableFields(sfTableFields)
		h.naloziService.SetNalogIDFieldName(common.IDfnal)
	}
	// Prepare TableData for UI
	//tbl := common.SetTableBasicData("", prometTableID, h.service.GetKontaAnalitickiTableFields(), "", prometURLKontaAnaliticki, 0, 0, 0, 0, h.cfg)

	tbl := common.SetTableBasicData("NALOZI", "nalozi-table", h.naloziService.GetNaloziTableFields(), "/api/nalozi/", "/api/nalozi/", pageSize, page, 0, 0, h.cfg)
	// Call the service to get ALL necessary data for the view
	ctx := c.Request.Context()

	viewData, err := h.naloziService.GetNalogViewData(ctx, &tbl, searchQuery, page, pageSize, isInitialRequest, sortBy, sortOrder, selectedTipdok)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}

	// Populate table rows based on entities from service
	tbl.Rows = h.populateTableRows(tbl, viewData.FnalEntities, h.naloziService.GetNaloziTableFields())
	tbl.FuncDblClick = "handleDblClickNalogSelection(this)"
	tbl.FuncClick = "selectRow(this)"
	// Update tab active state based on current view (usually default tab is active for main view)
	currentTabData := h.tabData // Make a copy to modify active states
	currentTabData.Tabs[0].IsActive = true
	currentTabData.Tabs[1].IsActive = false
	currentTabData.Tabs[2].IsActive = false
	currentTabData.Tabs[3].IsActive = false
	searchControl := domain.InputControl{
		ID:           "search-input",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite podatak za pretragu",
		HxActionURL:  naloziURLGetAllSearch,
		HxTarget:     "#nalozi-table",
		HxSwap:       "outerHTML",
		HxInclude:    "#tipdok",
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	// Render the appropriate template based on whether it's an initial load or HTMX swap
	userSession := domain.GetSessionFromContext(c)

	nalogPayload := domain.FnalPayload{
		Nalog: viewData.NextNalog,
		Danal: time.Date(userSession.SelectedGod, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Format(common.HtmlLayout),
		Datob: time.Date(userSession.SelectedGod, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Format(common.HtmlLayout),
	}

	// Get CSRF token from context (set by middleware)
	csrfToken := common.GetCsrfTokenFromSession(c)
	tbl.HxVals = hxValsKnjizenje
	tbl.Pagination.HxVals = hxValsKnjizenje

	if selectedTipdok == "" {
		err = tmpl_fin.NaloziContent(
			currentTabData,
			viewData.TipdokComboItems,
			*viewData.UkupnaObrada, // Dereference as template expects value
			h.btnSave,
			h.btnNoviNalog,
			tbl,
			searchControl,
			nalogPayload, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
	} else {
		// HTMX request, just render the table component
		c.Header("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Table(tbl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	}

	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
		return
	}
}

// FnalPrepis
func (h *FnalHandler) FnalPrepis(c *gin.Context) {
	requestSource := c.GetHeader("Hx-Trigger")
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	// Prepare TableData for UI
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl := common.SetTableBasicData("NALOZI ZAGLAVLJE", "nalozi-table", h.naloziService.GetNaloziTableFields(), "/api/nalozi/", "/api/nalozi/", 0, 0, 0, 0, h.cfg)

	// Get data for the header table
	ctx := c.Request.Context()
	searchText := c.Query("query")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	err := h.naloziService.GetNalogPrepisData(ctx, &tbl, page, pageSize, searchText, sortBy, sortOrder) // Not initial load for prepis tab
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for prepis")
		return
	}
	tbl.DetailTarget = "#nalozi_kopiranje"
	tbl.DetailURL = "/api/fpro/nalog/" // URL for fetching details
	// Set active tab for prepis
	currentTabData := h.tabData
	currentTabData.Tabs[0].IsActive = false
	currentTabData.Tabs[1].IsActive = true
	currentTabData.Tabs[2].IsActive = false
	currentTabData.Tabs[3].IsActive = false
	searchControl := domain.InputControl{
		ID:           "search-prepishdr",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  "/api/nalozi/prepis",
		HxTarget:     "#nalozi-table",
		HxSwap:       "innerHTML",
		HxInclude:    "#search-prepishdr, #idfnal",
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	tbl.HxVals = hxValsKopirajNalog
	tbl.Pagination.HxVals = hxValsKopirajNalog
	if requestSource == "prepis" {
		err = tmpl_fin.NaloziKopiranje(currentTabData, tbl, domain.TableData{}, searchControl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	} else {
		// If this is an HTMX request, we just render the table component
		err = tmpl.Table(tbl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	}
}

// FnalPrepis Save
func (h *FnalHandler) FnalPrepisSave(c *gin.Context) {
	var entity domain.Fnal
	var req domain.FnalPayload
	if err := c.ShouldBind(&req); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	ctx := c.Request.Context()
	h.naloziService.MapReqToEntity(ctx, req, &entity, common.ActionAdd)
	if c.Request.Method == http.MethodPost {
		fieldErrors, err := h.naloziService.ValidateCopyNalog(ctx, req, entity)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
			return
		}
		if len(fieldErrors) > 0 {
			common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
			return
		}
		exists, idfnal, err := h.naloziService.CheckNalogExistForCopy(ctx, entity)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
			return
		}
		if exists {
			copyToExistingNalog(c, req, idfnal)
			common.WriteJSONResponse(c, http.StatusOK, true, []domain.FieldError{}, "Nalog je uspešno kopiran...")
			return
		}

		ctx := c.Request.Context()
		err = h.naloziService.KopirajNalog(ctx, idfnal, entity)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
			return
		}
	}
	if c.Request.Method == http.MethodPut {
		target_IdFnal, err := utils.GetInt64FromParameterRequest(c, "id")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
			return
		}
		ctx := c.Request.Context()
		err = h.naloziService.KopirajNalogToExisting(ctx, req, target_IdFnal)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
			return
		}
	}
	common.WriteJSONResponse(c, http.StatusOK, true, []domain.FieldError{}, "Nalog je uspešno kopiran...")

}

// --- PRIVATE HELPER FUNCTIONS ---
func copyToExistingNalog(c *gin.Context, req domain.FnalPayload, idfnal int64) {
	dlgTitle := ""
	hxVals := ""
	translator := i18n.GetInstance()
	msg := []string{}
	hxVals = fmt.Sprintf(`{"idfnal": "%s", "tipdok": "%s", "nalog": "%s", "danal": "%s", "datob": "%s", "opis": "%s"}`, req.IDFnal, req.Tipdok, req.Nalog, req.Danal, req.Datob, req.Opis)

	dlgTitle = "Kopiranje u postojeci nalog"
	msg = append(msg, translator.Message(`Kopiranje u postojeci nalog?`))
	msg = append(msg, fmt.Sprintf("%s: %s", translator.Message(`Vrsta naloga`), req.OldTipdok))
	msg = append(msg, fmt.Sprintf("%s: %s", translator.Message(`Broj naloga`), req.OldNalog))

	dialog := domain.Dialog{
		Id:            "dialogconfirm",
		Title:         dlgTitle,
		OkText:        "Da",
		CancelText:    "Ne",
		SaveText:      "Da",
		HxTarget:      "#dialog-confirm",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}

	btnSacuvaj := domain.Button{
		Id:        "btn-sacuvaj",
		IsVisible: true,
		LabelText: "Da",
		HxVals:    hxVals,
		IdDialog:  dialog.Id,
		HxTarget:  "#dialog-confirm",
		BtnClass:  common.ClassSaveButton,
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Ne",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	btnSacuvaj.HxRequestType = "PUT"
	btnSacuvaj.HxActionURL = fmt.Sprintf("%s/%d", naloziURLUpdateCopy, idfnal)

	btnCancel.IdDialog = dialog.Id
	btnClose.IdDialog = dialog.Id

	csrfToken := common.GetCsrfTokenFromSession(c)
	tmpl.DialogConfirm(msg, dialog, btnClose, btnSacuvaj, btnCancel, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
}

// FnalPrepis, FnalStorniraj, FnalPrikaz:
func (h *FnalHandler) FnalPrepisDialog(c *gin.Context) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	idFnalParam := c.Query("id") //
	idFnal, err := strconv.ParseInt(idFnalParam, 10, 64)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}
	kopirajCeoNalog := c.Query("kopirajceonalog") == "1"
	// Get CSRF token from session first, then fallback to context
	csrfToken := common.GetCsrfTokenFromSession(c)
	// Get data for the header table
	result, err := h.naloziService.GetByID(c.Request.Context(), common.IDfnal, idFnal)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for prepis")
		return
	}
	dialog := domain.Dialog{
		Id:            "dialog_kopiraj",
		Title:         "Kopiraj Nalog",
		OkText:        "Kopiraj",
		CancelText:    "Otkaži",
		SaveText:      "Snimi",
		HxTarget:      "#nalozi_kopiranje",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
		HxActionURL:   fmt.Sprintf("/api/nalozi/prepis/%d", idFnal),
	}
	now := time.Now()
	currentBusinessDate := time.Date(userSession.SelectedGod, now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Format(common.HtmlLayout)
	modelView := domain.KopirajNalog{
		IDFnal:     idFnal,
		NalogOld:   fmt.Sprintf("%d", result.Nalog),
		DanalOld:   result.Danal.Format(common.DateLayout),
		DatKnjOld:  result.Datob.Format(common.DateLayout),
		Tipdok_Old: result.Tipdok,
		OpisOld:    result.Opis,
		DanalNew:   currentBusinessDate,
		DatknjNew:  currentBusinessDate,
	}

	ctx := c.Request.Context()
	tipdokValues, err := h.naloziService.GetTipdokOptions(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Tipdok options")
		return
	}
	if len(tipdokValues) > 0 {
		tipdok := tipdokValues[0].TipDok
		nextNalog, _ := h.naloziService.GetNextNalog(ctx, tipdok)
		modelView.NalogNew = fmt.Sprintf("%d", nextNalog)
	}
	for _, item := range tipdokValues {
		if strings.Trim(strings.ToLower(item.TipDok), " ") == strings.Trim(strings.ToLower(result.Tipdok), " ") {
			modelView.TipdokOld = fmt.Sprintf("%s-%s", result.Tipdok, item.Opis)
		}
		modelView.TipdokValues = append(modelView.TipdokValues, domain.ComboItem{Key: item.TipDok, Value: item.TipDok + "-" + item.Opis})
	}
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Kopiraj nalog",
		IsVisible:     true,
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   fmt.Sprintf("/api/nalozi/prepis/%d", idFnal),
		HxVals:        fmt.Sprintf(`{"idfnal": "%d"}`, idFnal),
		HxRequestType: "POST",
		HxTarget:      "#dialog-confirm",
		//HxOnAfterRequest: "handleFormResponse",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		LabelText: "Odustani ",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassCloseButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	if kopirajCeoNalog {
		err = tmpl_fin.NaloziKopiranjeDialog(dialog, modelView, btnSave, btnCancel, btnClose, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	} else {
		tblOriginal := common.SetTableBasicData("Stavke originalnog naloga", "stavke-originalnog-naloga", h.naloziService.GetNaloziTableFields(), "", "", 10, 0, 0, 0, h.cfg)
		tblCopy := common.SetTableBasicData("Stavke novog naloga", "stavke-novog-naloga", h.naloziService.GetNaloziTableFields(), "", "", 10, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblOriginal, "", naloziURLStorniraj, false, false, false)
		common.SetTableConfig(&tblCopy, "", naloziURLStorniraj, false, false, false)
		tblOriginal.ShowActions = false
		tblCopy.ShowActions = false

		btnPrikazi := domain.Button{
			Id:          "btn-prikazi",
			LabelText:   "Prikaži",
			IsVisible:   true,
			IdDialog:    dialog.Id,
			BtnClass:    common.ClassSaveButton,
			Icon:        "refresh",
			HxActionURL: fmt.Sprintf("/api/nalozi/prepis/stavke/%d", idFnal),
		}
		btnNazad := btnCancel
		btnNazad.LabelText = "Back"
		btnNazad.Icon = "back"
		tmpl_fin.NaloziKopiranjeDialogDeo(dialog, modelView, btnSave, btnClose, btnPrikazi, btnNazad, tblOriginal, tblCopy, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
	}

}

// FnalStorniranje
func (h *FnalHandler) FnalStorniraj(c *gin.Context) {
	requestSource := c.GetHeader("Hx-Trigger")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl := common.SetTableBasicData("NALOZI STORNIRANJE", "nalozi-storniranje-table", h.naloziService.GetNaloziTableFields(), naloziURLStorniraj, naloziURLStorniraj, pageSize, page, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "PREGLED NALOGA", naloziURLStorniraj, true, false, false)
	// Get data for the header table
	ctx := c.Request.Context()
	viewData, err := h.naloziService.GetNalogStornirajData(ctx, &tbl, page, pageSize, "", "", "") // Not initial load for prepis tab
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for storniranje")
		return
	}
	viewData.TableData.Rows = h.populateTableRows(viewData.TableData, viewData.FnalEntities, h.naloziService.GetNaloziTableFields())

	// Set active tab for storniraj
	currentTabData := h.tabData
	currentTabData.Tabs[0].IsActive = false
	currentTabData.Tabs[1].IsActive = false
	currentTabData.Tabs[2].IsActive = true
	currentTabData.Tabs[3].IsActive = false

	searchControl := domain.InputControl{
		ID:           "search-control",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  naloziURLGetAllSearch,
		HxTarget:     "#table-body",
		HxSwap:       "innerHTML",
		HxInclude:    "#tipdok",
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}

	if requestSource != "" && requestSource == "storniraj" {
		err = tmpl_fin.NaloziStorniranje(currentTabData, tbl, domain.TableData{}, searchControl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	} else {
		// If this is an HTMX request, we just render the table component
		err = tmpl.Table(tbl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	}
}

// FnalStorniranje Dialog
func (h *FnalHandler) FnalStornirajDialog(c *gin.Context) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	idFnal, err := utils.GetInt64FromQueryRequest(c, "id")
	if err != nil || idFnal == 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}
	// Get CSRF token from session first, then fallback to context
	csrfToken := common.GetCsrfTokenFromSession(c)
	// Get data for the header table
	result, err := h.naloziService.GetByID(c.Request.Context(), common.IDfnal, idFnal)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for prepis")
		return
	}
	dialog := domain.Dialog{
		Id:            "dialog_storniraj",
		Title:         "Storniraj Nalog",
		OkText:        "Storniraj",
		CancelText:    "Otkaži",
		SaveText:      "Snimi",
		HxTarget:      "#nalozi_storniranje",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
		HxActionURL:   fmt.Sprintf("%s/%d", naloziURLStorniraj, idFnal),
	}
	now := time.Now()
	currentBusinessDate := time.Date(userSession.SelectedGod, now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Format(common.HtmlLayout)

	modelView := domain.KopirajNalog{
		IDFnal:    idFnal,
		NalogOld:  fmt.Sprintf("%d", result.Nalog),
		DanalOld:  result.Danal.Format(common.HtmlLayout),
		DatKnjOld: result.Datob.Format(common.HtmlLayout),
		TipdokOld: result.Tipdok,
		OpisOld:   result.Opis,
		DanalNew:  currentBusinessDate,
		DatknjNew: currentBusinessDate,
	}

	ctx := c.Request.Context()
	tipdokValues, err := h.naloziService.GetTipdokOptions(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Tipdok options")
		return
	}
	nextNalog, _ := h.naloziService.GetNextNalog(ctx, result.Tipdok)
	modelView.NalogNew = fmt.Sprintf("%d", nextNalog)

	for _, item := range tipdokValues {
		if strings.Trim(strings.ToLower(item.TipDok), " ") == strings.Trim(strings.ToLower(result.Tipdok), " ") {
			modelView.TipdokOld = fmt.Sprintf("%s-%s", result.Tipdok, item.Opis)
		}
		modelView.TipdokValues = append(modelView.TipdokValues, domain.ComboItem{Key: item.TipDok, Value: item.TipDok + "-" + item.Opis})
	}
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Snimi nalog",
		IsVisible:     true,
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   fmt.Sprintf("%s/%d", naloziURLStorniraj, idFnal),
		HxVals:        fmt.Sprintf(`{"idfnal": "%d", "tipdok": "%s"}`, idFnal, result.Tipdok),
		HxRequestType: "POST",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		LabelText: "Odustani ",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassCloseButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	content := tmpl_fin.NaloziStorniranjeDialog(dialog, modelView, btnSave, btnClose, btnCancel, i18n.GetInstance(), csrfToken)
	err = tmpl.Dialog(dialog.Id, csrfToken, content, dialog, btnSave, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *FnalHandler) FnalStornirajSave(c *gin.Context) {
	idFnalParam := c.Param("id")
	idFnal, err := strconv.ParseInt(idFnalParam, 10, 64)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}
	var entity domain.Fnal
	var req domain.FnalPayload
	if err := c.ShouldBind(&req); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	ctx := c.Request.Context()
	// map request to entity
	h.naloziService.MapReqToEntity(ctx, req, &entity, common.ActionAdd)
	fieldErrors, err := h.naloziService.NalogValidation(ctx, entity, common.ActionAdd)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	err = h.naloziService.StornirajNalog(ctx, idFnal, entity)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, []domain.FieldError{}, "Nalog je uspešno storniran")
}

// ValidacijaNalogStorniranje validates the nalozi copy form data and returns an array of errors
func (h *FnalHandler) ValidacijaNalogStorniranje(c *gin.Context, danalStr, datobStr string, brNaloga int) []domain.FieldError {
	var errors []domain.FieldError

	// Parse and validate datal (naloga date)
	dPomDate, err := time.Parse(common.DateLayout, danalStr)
	if err != nil {
		errors = append(errors, domain.FieldError{Field: "danal", ErrorMessage: "morate uneti korektan datum naloga"})
	}
	session := domain.GetSessionFromContext(c)
	if session == nil {
		errors = append(errors, domain.FieldError{Field: "session", ErrorMessage: "User session not found"})
	} else if dPomDate.Year() != session.SelectedGod {
		errors = append(errors, domain.FieldError{Field: "danal", ErrorMessage: fmt.Sprintf("nekorektan datum naloga, godina mora biti jednaka poslovnoj %d", session.SelectedGod)})
	}

	// Parse and validate datob (obrade date)
	dPomDate, err = time.Parse(common.DateLayout, datobStr)
	if err != nil {
		errors = append(errors, domain.FieldError{Field: "datob", ErrorMessage: "morate uneti korektan datum obrade naloga"})
	}
	session = domain.GetSessionFromContext(c)
	if session == nil {
		errors = append(errors, domain.FieldError{Field: "session", ErrorMessage: "User session not found"})
	} else if dPomDate.Year() != session.SelectedGod {
		errors = append(errors, domain.FieldError{Field: "datob", ErrorMessage: fmt.Sprintf("nekorektan datum obrade, godina mora biti jednaka poslovnoj %d", session.SelectedGod)})
	}

	// Validate broj naloga (voucher number)
	if brNaloga == 0 {
		errors = append(errors, domain.FieldError{Field: "brNaloga", ErrorMessage: "morate uneti broj naloga"})
	}

	return errors
}

// FnalPrikazStampa renders the print view for nalozi. This is a full page render, not just a table update, so it stays in the handler.
func (h *FnalHandler) FnalPrikazStampa(c *gin.Context) {
	requestSource := c.GetHeader("Hx-Trigger")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tblHdr := common.SetTableBasicData("NALOZI STAMPANJE", "nalozi-stampanje-table", h.naloziService.GetNaloziStampaTableFields(), naloziURLStampa, naloziURLStampa, pageSize, page, 0, 0, h.cfg)
	common.SetTableConfig(&tblHdr, "PREGLED NALOGA", naloziURLStampa, true, false, false)
	tblHdr.HxVals = hxValsStampa
	tblHdr.Pagination.HxVals = hxValsStampa

	tblDet := common.SetTableBasicData("STAVKE NALOGA", "stavke-naloga-stampa", h.naloziService.GetNaloziStavkeStampaTableFields(), "", "", pageSize, page, 0, 0, h.cfg)
	common.SetTableConfig(&tblDet, "STAVKE NALOGA", naloziURLStampaDetalji, false, false, false)
	tblDet.ShowActions = false
	tblDet.URLGetAll = naloziURLStampaDetalji
	tblDet.URLPrefix = naloziURLStampaDetalji
	tblDet.HxVals = hxValsStampaDetalji
	tblDet.HxTarget = "#stavke-naloga"

	tblHdr.DetailTarget = "#stavke-naloga"
	tblHdr.DetailURL = naloziURLStampaDetalji
	tblHdr.DetailHxRequestType = "GET"
	tblHdr.DetailHxSwap = "innerHTML"
	tblHdr.DetailHxTrigger = "click, change delay:500ms"

	// Get data for the header table
	ctx := c.Request.Context()
	searchText := c.Query("query")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	err := h.naloziService.GetNalogStampanjeData(ctx, &tblHdr, true, page, pageSize, searchText, sortBy, sortOrder) // Not initial load for prepis tab
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for storniranje")
		return
	}
	err = h.naloziService.GetNalogStampanjeData(ctx, &tblHdr, false, page, pageSize, searchText, sortBy, sortOrder) // Not initial load for prepis tab
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for storniranje")
		return
	}
	// Set active tab for storniraj
	currentTabData := h.tabData
	currentTabData.Tabs[0].IsActive = false
	currentTabData.Tabs[1].IsActive = false
	currentTabData.Tabs[2].IsActive = false
	currentTabData.Tabs[3].IsActive = true

	searchControl := domain.InputControl{
		ID:           "search-control",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  naloziURLStampa,
		HxTarget:     "#nalozi-stampanje-table",
		HxSwap:       "innerHTML",
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	searchControlDetalji := domain.InputControl{
		ID:           "search-control-detalji",
		Label:        "Pretraži Stavke Naloga",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  naloziURLStampaDetalji,
		HxTarget:     "#stavke-naloga-stampa",
		HxSwap:       "innerHTML",
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	btnPrint := domain.Button{
		Id:            "btn-print-stavke",
		IsVisible:     true,
		LabelText:     "Štampa ",
		BtnClass:      common.ClassPrintButton,
		HxActionURL:   naloziURLPrintStavke,
		HxRequestType: "GET",
	}
	if requestSource != "" && requestSource == "prikaz" {
		err = tmpl_fin.NaloziStampanje(currentTabData, tblHdr, tblDet, btnPrint, searchControl, searchControlDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	} else {
		// If this is an HTMX request, we just render the table component
		err = tmpl.Table(tblHdr, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	}

}

// FnalPrikazStampaDetalji renders the details table for the print view. This is a full page render, not just a table update, so it stays in the handler.
func (h *FnalHandler) FnalPrikazStampaDetalji(c *gin.Context) {
	requestSource := c.GetHeader("X-Request-Source")
	idFnal, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, "Invalid ID parameter")
		return
	}
	url := fmt.Sprintf("%s/%d", naloziURLStampaDetalji, idFnal)
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tblDet := common.SetTableBasicData("STAVKE NALOGA", "stavke-naloga-stampa", h.naloziService.GetNaloziStavkeStampaTableFields(), "", "", pageSize, page, 0, 0, h.cfg)
	common.SetTableConfig(&tblDet, "STAVKE NALOGA", url, false, false, false)
	tblDet.ShowActions = false

	tblDet.URLGetAll = url
	tblDet.URLPrefix = url

	// Get data for the header table
	ctx := c.Request.Context()
	searchText := c.Query("query")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	err = h.naloziService.GetNalogStampanjeDataDetalji(ctx, &tblDet, true, page, pageSize, idFnal, searchText, sortBy, sortOrder) // Not initial load for prepis tab
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for storniranje")
		return
	}
	err = h.naloziService.GetNalogStampanjeDataDetalji(ctx, &tblDet, false, page, pageSize, idFnal, searchText, sortBy, sortOrder) // Not initial load for prepis tab
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for storniranje")
		return
	}
	// If this is an HTMX request, we just render the table component
	if requestSource == "searchinput" || requestSource == "btnpage" || requestSource == "tblheader" {
		err = tmpl.Table(tblDet, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	} else {
		searchControlDetalji := domain.InputControl{
			ID:           "search-control-detalji",
			Label:        "Pretraži Stavke Naloga",
			Type:         "search",
			Placeholder:  "Unesite broj naloga ili druge podatke",
			HxActionURL:  url,
			HxTarget:     "#stavke-naloga-stampa",
			HxSwap:       "innerHTML",
			Autocomplete: "off",
			HxVals:       fmt.Sprintf(`{"idfnal": "%d"}`, idFnal),
			Class:        common.ClassSearchInput,
		}
		tmpl_fin.NaloziStampanjeStavke(tblDet, searchControlDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	}

}

// populateTableRows extracts and formats data for table display.
// This logic can stay in handler as it's UI formatting specific.
func (h *FnalHandler) populateTableRows(tableData domain.TableData, entities []domain.Fnal, fieldsDef []domain.Fields) []domain.TableRow {
	var tableRows []domain.TableRow

	fieldCache := h.naloziService.GetFieldCache()

	for _, entity := range entities {
		val := reflect.ValueOf(entity)
		// No need for val.Kind() == reflect.Ptr check if service returns value types

		// Get ID field value dynamically
		idValue, _, found := common.GetFieldByNameCaseInsensitive(val, common.IDfnal)
		id := ""
		if found {
			id = fmt.Sprintf("%v", idValue)
		}

		// Extract specified fields dynamically
		var fields []string
		for _, fieldName := range fieldsDef {
			fieldInfo, found := fieldCache[strings.ToLower(fieldName.Name)]
			if !found {
				continue // or return error if field is required
			}

			value := common.GetFormattedValue(fieldInfo, val.FieldByName(fieldInfo.Name))
			fields = append(fields, value)
		}
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: tableData.BtnUpdate.IsVisible, HasDelete: tableData.BtnDelete.IsVisible}
		tableRows = append(tableRows, row)
	}
	return tableRows
}
func setStavkeButtons(requestType string, idFnal int64) (domain.Button, domain.Button) {
	btnSave := domain.Button{
		Id:               "btn-save-stavke",
		IsVisible:        true,
		LabelText:        "Sačuvaj ",
		BtnClass:         common.ClassSaveButton,
		HxTarget:         "#nalog-stavke-table",
		HxActionURL:      fmt.Sprintf(naloziURLSaveStavke, idFnal),
		HxRequestType:    requestType,
		HxOnAfterRequest: "",
	}

	btnPrint := domain.Button{
		Id:            "btn-print-stavke",
		IsVisible:     true,
		LabelText:     "Štampa ",
		BtnClass:      common.ClassPrintButton,
		HxTarget:      "#nalog-stavke-table",
		HxActionURL:   naloziURLPrintStavke,
		HxRequestType: "GET",
	}
	return btnSave, btnPrint
}

func (h *FnalHandler) AddRoutes(r *gin.Engine) {
	// Create API group with prefix
	//api := r.Group(naloziURLPrefix)
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Consolidate main view routes
	r.GET("/api/nalozi/all", h.GetNalogMainView)
	r.GET("/api/nalozi/all/tipdok", h.GetNalogMainView) // Both can hit same handler
	r.GET("/api/nalozi/all/search", h.GetNalogMainView) // Both can hit same handler
	r.GET("/api/nalozi/knjizenje", h.GetNalogMainView)  // Also points to main view

	// ... (other CRUD and confirm routes)
	r.POST("/api/nalozi/new", h.CreateNalog)
	r.GET("/api/nalozi/nextnalog", h.GetNextNalog)
	r.GET("/api/nalozi/update-nalog-data", h.GetUpdateDataNalog)
	r.GET("/api/nalozi/confirm-delete", h.confirmDeleteHandler)
	r.POST("/api/nalozi/confirm-addupdate", h.confirmAddHandler)
	r.GET("/api/nalozi/confirm-add", h.confirmAddHandler)
	r.GET("/api/nalozi/:id", h.GetNalog)
	r.PUT("/api/nalozi/update/:id", h.lm.WithEntityLockHold("fnal", "id"), h.UpdateNalog)
	r.DELETE("/api/nalozi/:id", h.DeleteNalog)
	r.POST("/api/nalozi/unlock/:id", h.lm.WithEntityLockVerifyAndRelease("fnal", "id"), h.UnlockNalog)
	r.GET("/api/nalozi/prepis", h.FnalPrepis)
	r.POST("/api/nalozi/prepis/:id", h.FnalPrepisSave)
	r.GET("/api/nalozi/confirm-prepis", h.FnalPrepisDialog)
	r.PUT("/api/nalozi/update/copy/:id", h.FnalPrepisSave)
	r.GET("/api/nalozi/confirm-storniraj", h.FnalStornirajDialog)
	r.GET("/api/nalozi/storniraj", h.FnalStorniraj)
	r.POST("/api/nalozi/storniraj/:id", h.FnalStornirajSave)
	r.GET("/api/nalozi/prikaz", h.FnalPrikazStampa)
	r.GET("/api/nalozi/prikaz/detalji/:id", h.FnalPrikazStampaDetalji)
}

func (h *FnalHandler) setHandlerFieldValues() {
	h.tabData = domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "knjizenje", Label: "Knjiženje Naloga", HXRequestUrl: fmt.Sprintf("%sknjizenje", naloziURLPrefix), IsActive: true, Name: "knjizenje"},
			{ID: "prepis", Label: "Prepis Naloga", HXRequestUrl: fmt.Sprintf("%sprepis", naloziURLPrefix), IsActive: false, Name: "prepis"},
			{ID: "storniraj", Label: "Storniranje naloga", HXRequestUrl: fmt.Sprintf("%sstorniraj", naloziURLPrefix), IsActive: false, Name: "storniraj"},
			{ID: "prikaz", Label: "Prikaz/Štampa naloga", HXRequestUrl: fmt.Sprintf("%sprikaz", naloziURLPrefix), IsActive: false, Name: "stampa"},
		},
	}

	h.btnSave = domain.Button{
		Id:               "btn-save",
		IdDialog:         "addupdate-dialog",
		IsVisible:        true,
		LabelText:        "Snimi Nalog",
		HxActionURL:      "/api/nalozi/confirm-addupdate",
		HxTarget:         "#dialog-confirm",
		HxSwap:           "innerHTML",
		HxOnAfterRequest: "handleDialogResponse('addupdate-dialog')", //hx-on::after-request="handleDialogResponse('addupdate-dialog')"
		HxRequestType:    "POST",
		BtnClass:         common.ClassSaveButton + " w-24",
		HxInclude:        "#fin-nalozi",
	}
	h.btnNoviNalog = domain.Button{
		Id:               "btn-novi-nalog",
		LabelText:        "Novi Nalog",
		HxActionURL:      naloziURLNextNalog,
		HxInclude:        "#tipdok",
		HxRequestType:    "GET",
		HxOnAfterRequest: "handleNextNalogResponse",
		BtnClass:         common.ClassNewButton + " w-28",
	}
}

func getIdActiveTab(sourceTab string) string {
	switch sourceTab {
	case "knjizenje":
		return "tabknjizenje"
	case "prepis":
		return "tabprepis"
	case "storniraj":
		return "tabstorniraj"
	case "prikaz":
		return "tabprikaz"
	default:
		return "" // Default to knjizenje if no match
	}
}
