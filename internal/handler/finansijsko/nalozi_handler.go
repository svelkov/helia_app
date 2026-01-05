package finansijsko

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
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
	naloziContentTitle    string = "NALOZI"
	naloziTableID         string = "nalozi-table"
	naloziURLPrefix       string = "/api/nalozi/"
	naloziURLGetAllSearch string = "/api/nalozi/all/search"
	naloziURLNextNalog    string = "/api/nalozi/nextnalog"
	naloziURLCreate       string = "/api/nalozi/new"
	naloziURLUpdate       string = "/api/nalozi/update"
	ActionAdd             string = "add"
	ActionUpdate          string = "update"
	naloziURLSaveStavke   string = "/api/fpro/nalog/stavke/save"
	naloziURLPrintStavke  string = "/api/print/nalog/stavke"
)

// FnalHandler handles requests related to "fnal" entities.
type FnalHandler struct {
	naloziService service.NalogService
	service       service.BaseService[domain.Fnal]
	tabData       domain.TabData
	btnSave       domain.Button
	btnNoviNalog  domain.Button
	typeView      string // Type of view, e.g. "knjizenje", "prepis", etc.
}

func NewFnalHandler(nalogService service.NalogService, baseService *service.BaseService[domain.Fnal]) *FnalHandler {
	handler := &FnalHandler{
		naloziService: nalogService,
		service:       *baseService,
	}
	handler.setHanlderFieldValues()
	return handler
}

func (h *FnalHandler) GetNalog(c *gin.Context) {
	utils.GetEntityHelper(c, h.naloziService, h.naloziService.GetNaloziTableFields(), common.IDfnal)
}

// GetNextNalog returns the next available Nalog number for a given tipdok.
func (h *FnalHandler) GetNextNalog(c *gin.Context) {
	tipdok := c.Query("tipdok")

	nextNalog, err := h.naloziService.GetNextNalog(tipdok)
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
	nalogEntity, err := h.naloziService.GetByTipdokNalog(tipdok, int64(nalogNbr))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get next Nalog")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"datob":   nalogEntity.Datob.Format("2006-01-02"),
		"danal":   nalogEntity.Danal.Format("2006-01-02"),
		"opis":    nalogEntity.Opis,
	})
}
func (h *FnalHandler) CreateNalog(c *gin.Context) {
	var entity domain.Fnal
	var nalog domain.FnalPayload
	var fproPayload domain.FproPayload
	if err := c.ShouldBind(&nalog); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	// map request to entity
	h.naloziService.MapReqToEntity(c, nalog, &entity, ActionAdd)
	id, err := h.naloziService.GetIdTipdokByTipdok(nalog.Tipdok)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	entity.IDTipdok = id
	fieldErrors, lastInsertedID, err := h.naloziService.Create(&entity, common.IDfnal, h.service.MapEntityToValues(&entity, h.naloziService.GetNaloziTableFields()))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	// Lock the header using global resource lock
	mutex := global.GetHeaderLock(lastInsertedID)
	// Try to lock without blocking
	if !mutex.TryLock() {
		common.WriteJSONResponse(c, http.StatusConflict, false, []domain.FieldError{}, common.ErrMsgStatusConflict)
		return
	}
	defer mutex.Unlock() // Ensure mutex is always unlocked, even if an error occurs
	h.naloziService.MapToFproPayload(&fproPayload, entity, lastInsertedID)
	orgjed, err := h.naloziService.GetOrgJedinice()
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	mtroska, err := h.naloziService.GetMestoTroska(0) // No orgjed yet for new nalog
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	fproPayload.Orgjed = orgjed
	fproPayload.Mtroska = mtroska

	tblStavke := common.SetTableBasicData("Stavke Naloga", naloziTableID+"_stavke", h.service.MapEntityToValues(&entity, h.naloziService.GetNaloziTableFields()), "", "", 10, 0, 0, 0)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	btnSave, btnPrint := setStavkeButtons("POST")
	err = tmpl_fin.NalogKnjizenjeStavke(fproPayload, tblStavke, btnSave, btnPrint, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
}

// UpdateNalog handles the update of an existing "fnal" entity.
func (h *FnalHandler) UpdateNalog(c *gin.Context) {
	fproPayload, tblStavke, err, errorCode, errMessage, fieldErrors := h.naloziService.UpdateNalog(c)
	if err != nil {
		common.WriteJSONResponse(c, errorCode, false, fieldErrors, errMessage)
		return
	}
	btnSave, btnPrint := setStavkeButtons("PUT")
	err = tmpl_fin.NalogKnjizenjeStavke(fproPayload, tblStavke, btnSave, btnPrint, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
}

func setStavkeButtons(requestType string) (domain.Button, domain.Button) {
	btnSave := domain.Button{
		Id:            "btn-save-stavke",
		IsVisible:     true,
		LabelText:     "Sačuvaj ",
		BtnClass:      common.ClassSaveButton,
		HxTarget:      "#nalozi-table-stavke",
		HxActionURL:   naloziURLSaveStavke,
		HxRequestType: requestType,
	}

	btnPrint := domain.Button{
		Id:            "btn-print-stavke",
		IsVisible:     true,
		LabelText:     "Štampa ",
		BtnClass:      common.ClassPrintButton,
		HxTarget:      "#nalozi-table-stavke",
		HxActionURL:   naloziURLPrintStavke,
		HxRequestType: "GET",
	}
	return btnSave, btnPrint
}

func (h *FnalHandler) DeleteNalog(c *gin.Context) {
	utils.DeleteHelper(c, h.naloziService, common.IDfnal)
}

func (h *FnalHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.naloziService.GetNaloziTableFields())
}

func (h *FnalHandler) confirmAddHandler(c *gin.Context) {
	var entity domain.Fnal
	var req domain.FnalPayload
	var action string = ActionAdd

	if err := c.ShouldBind(&req); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	// map request to entity
	h.naloziService.MapReqToEntity(c, req, &entity, ActionAdd)
	fieldErrors, err := h.naloziService.Validation(entity)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	nalogEntity, err := h.naloziService.GetByTipdokNalog(entity.Tipdok, entity.Nalog)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldErrors, common.ErrMsgReadData)
		return
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
		action = ActionUpdate
		dlgTitle = "Nastavak knjiženja naloga"
		msg = append(msg, translator.Message(`Nastavak knjiženja naloga?`))
		msg = append(msg, fmt.Sprintf("%s: %s", translator.Message(`Vrsta naloga`), req.Tipdok))
		msg = append(msg, fmt.Sprintf("%s: %s", translator.Message(`Broj naloga`), req.Nalog))
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
	if action == ActionAdd {
		dialog.Id = "dlg_nalog_create"
		btnSacuvaj.HxRequestType = "POST"
		btnSacuvaj.HxActionURL = naloziURLCreate
	}
	if action == ActionUpdate {
		dialog.Id = "dlg_nalog_update"
		btnSacuvaj.HxRequestType = "PUT"
		btnSacuvaj.HxActionURL = fmt.Sprintf("%s/%d", naloziURLUpdate, nalogEntity.IDFnal)
	}
	btnCancel.HxOn = fmt.Sprintf("handleDialogResponse('%s')", dialog.Id)
	btnClose.HxOnAfterRequest = fmt.Sprintf("handleDialogResponse('%s')", dialog.Id)
	btnCancel.IdDialog = dialog.Id
	btnClose.IdDialog = dialog.Id

	csrfToken := common.GetCsrfTokenFromSession(c)
	tmpl.DialogConfirm(msg, dialog, btnClose, btnSacuvaj, btnCancel, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *FnalHandler) GetIdTipdokByTipdok(tipdok string) (int64, error) {
	return h.naloziService.GetIdTipdokByTipdok(tipdok)
}
func (h *FnalHandler) getNalogStavkeHandler(c *gin.Context) {
	// Render update form with entity data.
	utils.ConfirmUpdateHelper[domain.Fnal](c, nil, h.naloziService.GetNaloziTableFields(), common.IDfnal) // Placeholder
}

// --- CONSOLIDATED VIEW DATA HANDLER ---
func (h *FnalHandler) GetNalogMainView(c *gin.Context) {
	searchQuery := c.Query("query")
	selectedTipdok := c.Query("tipdok")

	page, pageSize := common.GetPageAndPageSizeFromRequest(c)

	// Determine if this is an initial full page request (not an HTMX swap)
	isInitialRequest := selectedTipdok == ""
	if isInitialRequest {
		h.naloziService.SetSfTableFields(sfTableFields)
		h.naloziService.SetNalogIDFieldName(common.IDfnal)
	}
	// Call the service to get ALL necessary data for the view
	viewData, err := h.naloziService.GetNalogViewData(c, searchQuery, selectedTipdok, page, pageSize, isInitialRequest)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}

	// Populate table rows based on entities from service
	viewData.TableData.Rows = h.populateTableRows(viewData.TableData, viewData.FnalEntities, h.naloziService.GetNaloziTableFields())
	viewData.TableData.FuncDblClick = "handleDblClickNalogSelection(this)"
	viewData.TableData.FuncClick = "selectRow(this)"
	// Update tab active state based on current view (usually default tab is active for main view)
	currentTabData := h.tabData // Make a copy to modify active states
	currentTabData.Tabs[0].IsActive = true
	currentTabData.Tabs[1].IsActive = false
	currentTabData.Tabs[2].IsActive = false
	currentTabData.Tabs[3].IsActive = false
	searchControl := domain.InputControl{
		ID:           "search-control",
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
	nalogPayload := domain.FnalPayload{
		Nalog: viewData.NextNalog,
		Danal: time.Now().Format("2006-01-02"),
		Datob: time.Now().Format("2006-01-02"),
	}

	// Get CSRF token from context (set by middleware)
	csrfToken := common.GetCsrfTokenFromSession(c)

	if viewData.IsInitialLoad {
		err = tmpl_fin.NaloziContent(
			currentTabData,
			viewData.TipdokComboItems,
			*viewData.UkupnaObrada, // Dereference as template expects value
			h.btnSave,
			h.btnNoviNalog,
			viewData.TableData,
			searchControl,
			nalogPayload, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
	} else {
		// HTMX request, just render the table component
		c.Header("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Table(viewData.TableData, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	}

	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
		return
	}
}

// FnalPrepis
func (h *FnalHandler) FnalPrepis(c *gin.Context) {
	searchQuery := c.Query("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c)

	// Get data for the header table
	viewData, err := h.naloziService.GetNalogPrepisData(c, searchQuery, page, pageSize) // Not initial load for prepis tab
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Nalog header data for prepis")
		return
	}
	viewData.TableData.Rows = h.populateTableRows(viewData.TableData, viewData.FnalEntities, h.naloziService.GetNaloziTableFields())
	viewData.TableData.DetailTarget = "#nalozi_kopiranje"
	viewData.TableData.DetailURL = "/api/fpro/nalog/" // URL for fetching details
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
	if c.Request.Header["Hx-Trigger"] != nil && c.Request.Header["Hx-Trigger"][0] == "prepis" {
		err = tmpl_fin.NaloziKopiranje(currentTabData, viewData.TableData, domain.TableData{}, searchControl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	} else {
		// If this is an HTMX request, we just render the table component
		err = tmpl.Table(viewData.TableData, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	}
}

// FnalPrepis, FnalStorniranje, FnalPrikaz:
func (h *FnalHandler) FnalPrepisDialog(c *gin.Context) {
	idFnalParam := c.Query("idfnal") //
	idFnal, err := strconv.ParseInt(idFnalParam, 10, 64)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}
	// Get CSRF token from session first, then fallback to context
	csrfToken := common.GetCsrfTokenFromSession(c)
	// Get data for the header table
	result, err := h.naloziService.GetByID(common.IDfnal, idFnal)
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
	modelView := domain.KopirajNalog{
		IDFnal:    idFnal,
		NalogOld:  fmt.Sprintf("%d", result.Nalog),
		DanalOld:  result.Danal.Format("01.01.2006"),
		DatKnjOld: result.Datob.Format("02.01.2006"),
		OpisOld:   result.Opis,
		DanalNew:  time.Now().Format("2006-01-02"), // it is important for frontend, format should be YYYY-MM-DD
		DatknjNew: time.Now().Format("2006-01-02"),
	}

	tipdokValues, err := h.naloziService.GetTipdokOptions()
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Tipdok options")
		return
	}
	if len(tipdokValues) > 0 {
		tipdok := tipdokValues[0].TipDok
		nextNalog, _ := h.naloziService.GetNextNalog(tipdok)
		modelView.NalogNew = fmt.Sprintf("%d", nextNalog)
	}
	for _, item := range tipdokValues {
		if strings.Trim(strings.ToLower(item.TipDok), " ") == strings.Trim(strings.ToLower(result.Tipdok), " ") {
			modelView.TipdokOld = fmt.Sprintf("%s-%s", result.Tipdok, item.Opis)
		}
		modelView.TipdokValues = append(modelView.TipdokValues, domain.ComboItem{Key: item.TipDok, Value: item.TipDok + "-" + item.Opis})
	}
	btnSave := domain.Button{
		Id:               "btn-save",
		LabelText:        "Snimi nalog",
		IsVisible:        true,
		IdDialog:         dialog.Id,
		HxInclude:        "#fin-nalozi",
		BtnClass:         common.ClassSaveButton,
		HxActionURL:      fmt.Sprintf("/api/nalozi/prepis/%d", idFnal),
		HxVals:           fmt.Sprintf(`{"idfnal": '"%d"}`, idFnal),
		HxRequestType:    "POST",
		HxOnAfterRequest: "closeDialog",
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
	content := tmpl_fin.NaloziKopiranjeDialog(modelView, btnSave, i18n.GetInstance(), csrfToken)
	err = tmpl.Dialog(dialog.Id, csrfToken, content, dialog, btnSave, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
		return
	}
}

// FnalPrepis Save
func (h *FnalHandler) FnalPrepisSave(c *gin.Context) {
	// modelView := domain.KopirajNalog{}
	idFnalParam := c.Param("idfnal")
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
	// map request to entity
	h.naloziService.MapReqToEntity(c, req, &entity, ActionAdd)
	fieldErrors, err := h.naloziService.ValidationKopirajNalog(entity)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}

	fmt.Println(idFnal)

}

func (h *FnalHandler) FnalStorniraj(c *gin.Context) {
	searchQuery := c.Query("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c)

	// Get data for the header table
	viewData, err := h.naloziService.GetNalogStornirajData(c, searchQuery, page, pageSize) // Not initial load for prepis tab
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

	if c.Request.Header["Hx-Trigger"] != nil && c.Request.Header["Hx-Trigger"][0] == "storniranje" {
		err = tmpl_fin.NaloziStorniranje(currentTabData, viewData.TableData, domain.TableData{}, searchControl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	} else {
		// If this is an HTMX request, we just render the table component
		err = tmpl.Table(viewData.TableData, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
			return
		}
	}
}

// FnalPrepis, FnalStorniranje, FnalPrikaz:
func (h *FnalHandler) FnalStornirajDialog(c *gin.Context) {
	idFnalParam := c.Query("id") //
	idFnal, err := strconv.ParseInt(idFnalParam, 10, 64)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}
	// Get CSRF token from session first, then fallback to context
	csrfToken := common.GetCsrfTokenFromSession(c)
	// Get data for the header table
	result, err := h.naloziService.GetByID(common.IDfnal, idFnal)
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
		HxActionURL:   fmt.Sprintf("/api/nalozi/storniranje/%d", idFnal),
	}
	modelView := domain.KopirajNalog{
		IDFnal:    idFnal,
		NalogOld:  fmt.Sprintf("%d", result.Nalog),
		DanalOld:  result.Danal.Format("02.01.2006"),
		DatKnjOld: result.Datob.Format("02.01.2006"),
		OpisOld:   result.Opis,
		DanalNew:  time.Now().Format("02.01.2006"), // it is important for frontend, format should be YYYY-MM-DD
		DatknjNew: time.Now().Format("02.01.2006"),
	}

	tipdokValues, err := h.naloziService.GetTipdokOptions()
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get Tipdok options")
		return
	}
	if len(tipdokValues) > 0 {
		tipdok := tipdokValues[0].TipDok
		nextNalog, _ := h.naloziService.GetNextNalog(tipdok)
		modelView.NalogNew = fmt.Sprintf("%d", nextNalog)
	}
	for _, item := range tipdokValues {
		if strings.Trim(strings.ToLower(item.TipDok), " ") == strings.Trim(strings.ToLower(result.Tipdok), " ") {
			modelView.TipdokOld = fmt.Sprintf("%s-%s", result.Tipdok, item.Opis)
		}
		modelView.TipdokValues = append(modelView.TipdokValues, domain.ComboItem{Key: item.TipDok, Value: item.TipDok + "-" + item.Opis})
	}
	btnSave := domain.Button{
		Id:               "btn-save",
		LabelText:        "Snimi nalog",
		IsVisible:        true,
		IdDialog:         dialog.Id,
		BtnClass:         common.ClassSaveButton,
		HxActionURL:      fmt.Sprintf("/api/nalozi/storniraj/%d", idFnal),
		HxVals:           fmt.Sprintf(`{"idfnal": "%d", "tipdok": "%s"}`, idFnal, result.Tipdok),
		HxRequestType:    "POST",
		HxOnAfterRequest: "closeDialog",
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
	content := tmpl_fin.NaloziStorniranjeDialog(modelView, btnSave, i18n.GetInstance(), csrfToken)
	err = tmpl.Dialog(dialog.Id, csrfToken, content, dialog, btnSave, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *FnalHandler) FnalStornirajSave(c *gin.Context) {
	idFnalParam := c.Param("idfnal")
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
	// map request to entity
	h.naloziService.MapReqToEntity(c, req, &entity, ActionAdd)
	fieldErrors, err := h.naloziService.ValidationKopirajNalog(entity)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}

	fmt.Println(idFnal)

}

// ValidacijaNalogStorniranje validates the nalozi copy form data and returns an array of errors
func (h *FnalHandler) ValidacijaNalogStorniranje(danalStr, datobStr string, brNaloga int) []domain.FieldError {
	var errors []domain.FieldError

	// Parse and validate datal (naloga date)
	dPomDate, err := time.Parse("02.01.2006", danalStr)
	if err != nil {
		errors = append(errors, domain.FieldError{Field: "danal", ErrorMessage: "morate uneti korektan datum naloga"})
	} else if dPomDate.Year() != global.GetGnGod() {
		errors = append(errors, domain.FieldError{Field: "danal", ErrorMessage: fmt.Sprintf("nekorektan datum naloga, godina mora biti jednaka poslovnoj %d", global.GetGnGod())})
	}

	// Parse and validate datob (obrade date)
	dPomDate, err = time.Parse("02.01.2006", datobStr)
	if err != nil {
		errors = append(errors, domain.FieldError{Field: "datob", ErrorMessage: "morate uneti korektan datum obrade naloga"})
	} else if dPomDate.Year() != global.GetGnGod() {
		errors = append(errors, domain.FieldError{Field: "datob", ErrorMessage: fmt.Sprintf("nekorektan datum obrade, godina mora biti jednaka poslovnoj %d", global.GetGnGod())})
	}

	// Validate broj naloga (voucher number)
	if brNaloga == 0 {
		errors = append(errors, domain.FieldError{Field: "brNaloga", ErrorMessage: "morate uneti broj naloga"})
	}

	return errors
}

func (h *FnalHandler) FnalPrikaz(c *gin.Context) {
	// Render the template for nalozi with the data
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
	// Get CSRF token from context
	csrfToken, _ := c.Get("csrf_token")
	csrfTokenStr := ""
	if token, ok := csrfToken.(string); ok {
		csrfTokenStr = token
	}
	err := tmpl_fin.NaloziContent(h.tabData, nil, domain.UkupnaObrada{}, h.btnSave, h.btnNoviNalog, domain.TableData{}, searchControl, domain.FnalPayload{}, i18n.GetInstance(), csrfTokenStr).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
		return
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
	r.PUT("/api/nalozi/update/:id", h.UpdateNalog)
	r.DELETE("/api/nalozi/:id", h.DeleteNalog)
	r.GET("/api/nalozi/prepis", h.FnalPrepis)
	r.POST("/api/nalozi/prepis/:idfnal", h.FnalPrepisSave)
	r.GET("/api/nalozi/confirm-prepis", h.FnalPrepisDialog)
	r.GET("/api/nalozi/confirm-storniraj", h.FnalStornirajDialog)
	r.GET("/api/nalozi/storniranje", h.FnalStorniraj)
	r.POST("/api/nalozi/storniraj/:idfnal", h.FnalStornirajSave)
	r.GET("/api/nalozi/prikaz", h.FnalPrikaz)
}

func (h *FnalHandler) setHanlderFieldValues() {
	h.typeView = "knjizenje" // Default view type, can be changed based on context
	h.tabData = domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "knjizenje", Label: "Knjiženje Naloga", HXRequestUrl: fmt.Sprintf("%sknjizenje", naloziURLPrefix), IsActive: true, Name: "knjizenje"},
			{ID: "prepis", Label: "Prepis Naloga", HXRequestUrl: fmt.Sprintf("%sprepis", naloziURLPrefix), IsActive: false, Name: "prepis"},
			{ID: "storniranje", Label: "Storniranje naloga", HXRequestUrl: fmt.Sprintf("%sstorniranje", naloziURLPrefix), IsActive: false, Name: "storniranje"},
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
	case "storniranje":
		return "tabstorniranje"
	case "prikaz":
		return "tabprikaz"
	default:
		return "" // Default to knjizenje if no match
	}
}
