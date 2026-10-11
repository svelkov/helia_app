package robno

import (
	"fmt"
	"net/http"
	"strconv"

	"helia/config"
	tmpl_robno_rep "helia/frontend/templates/reports/robno"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	magacinKontoContentTitle string = "MAGACIN KONTO"
	magacinKontoTableID      string = "magacin-konto-table"
	magacinKontoURLPrefix    string = "/api/magacin-konto"
	magacinKontoURLGetAll    string = "/api/magacin-konto/all"
	magacinKontoURLPrint     string = "/api/magacin-konto/stampa"
	magacinKontoURLNovi      string = "/api/magacin-konto/confirm-add"
	magacinKontoURLSave      string = "/api/magacin-konto/save"
	magacinKontoURLDelete    string = "/api/magacin-konto/confirm-delete"
	magacinKontoURLUpdate    string = "/api/magacin-konto/confirm-update"
)

func SetMagacinKontoFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "8"},
		{Name: "konto", Label: "Konto zaliha", Width: "15"},
		{Name: "opis_konta", Label: "Opis konta zalihe", Width: "25"},
		{Name: "opis_mag", Label: "Opis magacina", Width: "25"},
		{Name: "kontoprih", Label: "Konto prihoda", Width: "15"},
		{Name: "kontotroska", Label: "Konto troška", Width: "15"},
		{Name: "kontoruc", Label: "Konto RUC", Width: "15"},
	}
}

type MagacinKontoHandler struct {
	translator          *i18n.Service
	service             service.Service[domain.Magkonto]
	magacinKontoService robnosvc.MagacinKontoService
	cfg                 config.Config
	lm                  *middleware.LockMiddleware
}

func NewMagacinKontoHandler(service *service.BaseService[domain.Magkonto], magacinKontoService robnosvc.MagacinKontoService, cfg config.Config, lm *middleware.LockMiddleware, translator *i18n.Service) *MagacinKontoHandler {
	return &MagacinKontoHandler{
		translator:          translator,
		service:             service,
		magacinKontoService: magacinKontoService,
		cfg:                 cfg,
		lm:                  lm,
	}
}

func (h *MagacinKontoHandler) CreateMagacinKonto(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	dto := domain.Magkonto{}
	if err := c.ShouldBind(&dto); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	// The validation also reads the broj of the magacin and the konto and vkonta of the konto zaliha.
	fieldErrors := h.magacinKontoService.ValidateEntity(ctx, &dto)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	_, _, err := h.magacinKontoService.Create(ctx, &dto, common.IDmagkonto, magacinKontoSaveFields(dto))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

func (h *MagacinKontoHandler) UpdateMagacinKonto(c *gin.Context) {
	ctx := c.Request.Context()
	var magacinKonto domain.Magkonto
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	getEntity, err := h.magacinKontoService.GetByID(ctx, common.IDmagkonto, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if getEntity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgGetData)
		return
	}
	err = c.ShouldBind(&magacinKonto)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	// The row that is changed (the validation does not count it as a second konto of the magacin); the
	// validation also reads the broj of the magacin and the konto and vkonta of the konto zaliha.
	magacinKonto.MagkontoID = id
	if fieldErrors := h.magacinKontoService.ValidateEntity(ctx, &magacinKonto); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	fieldErrors, err := h.magacinKontoService.Update(ctx, &magacinKonto, common.IDmagkonto, id, magacinKontoSaveFields(magacinKonto))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

// magacinKontoSaveFields returns the columns of magkonto a save writes, with their values (the magacin
// and its broj, the konto zaliha with its konto and vkonta, the konta of prihod, trošak, RUC and rabat).
func magacinKontoSaveFields(m domain.Magkonto) []domain.Fields {
	return []domain.Fields{
		{Name: "magaciniid", Value: strconv.Itoa(m.MagaciniID)},
		{Name: "mag", Value: strconv.Itoa(m.Mag)},
		{Name: "konto", Value: m.Konto},
		{Name: "idfkpl", Value: strconv.Itoa(m.IDFkpl)},
		{Name: "vkonta", Value: strconv.Itoa(m.VKonta)},
		{Name: "kontoprih", Value: m.KontoPrih},
		{Name: "kontotroska", Value: m.KontoTroska},
		{Name: "kontoruc", Value: m.KontoRuc},
		{Name: "kontorab", Value: m.KontoRab},
	}
}

func (h *MagacinKontoHandler) DeleteMagacinKonto(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	err = h.magacinKontoService.Delete(ctx, common.IDmagkonto, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

func (h *MagacinKontoHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, SetMagacinKontoFields(), "#info-message")
}

func (h *MagacinKontoHandler) confirmAddHandler(c *gin.Context) {
	dialog := domain.Dialog{
		Title:         "Novi magacin konto",
		Id:            "magacin-konto-add-form",
		HxActionURL:   magacinKontoURLSave,
		HxRequestType: "POST",
		OkText:        "Sačuvaj",
		CancelText:    "Odustani",
	}

	model := domain.MagkontoView{}

	// A new one: after the save the form is emptied for the next one (magacinKontoNoviResponse).
	btnSave := magacinKontoSaveButton(magacinKontoURLSave, "POST", dialog.Id, "magacinKontoNoviResponse")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	csrfToken := common.GetCsrfToken(c)

	tmpl_robno.MagacinKontoDialog(dialog, common.ActionAdd, model, btnSave, btnCancel, btnClose, h.translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

// magacinKontoSaveButton is "Sačuvaj" of the form (POST of a new konto zaliha, PUT of a changed one): its
// target is the message area at the bottom of tmpl.Dialog (dialog-form-message), but the answer (JSON) is
// not swapped into it (hx-swap none); afterRequest (called with the id of the dialog) shows its
// message at the bottom of the dialog or the errors under the fields: magacinKontoIzmenaResponse closes
// the dialog after a successful save, magacinKontoNoviResponse empties the form for the next one.
func magacinKontoSaveButton(url, requestType, dialogID, afterRequest string) domain.Button {
	btn := common.SetButton("btn-save-magacin-konto", "Sačuvaj", "save", url, "#dialog-form-message", "none", requestType, "", "", true, common.ClassSaveButton, afterRequest)
	btn.IdDialog = dialogID
	return btn
}

func (h *MagacinKontoHandler) confirmUpdateHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	rowID := c.Query("id")
	id, err := strconv.Atoi(rowID)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	// The row of the grid (its key is magkontoid) with the texts of its magacin and konta.
	entity, found, err := h.magacinKontoService.GetMagacinKontoView(ctx, int64(id))
	if err == nil && !found {
		err = fmt.Errorf("%s", common.ErrMsgGetData)
	}
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	updateURL := fmt.Sprintf("/api/magacin-konto/%d", id)
	dialog := domain.Dialog{
		Title:         "Izmena magacin konto",
		Id:            "magacin-konto-edit-form",
		HxActionURL:   updateURL,
		HxRequestType: "PUT",
	}

	// A changed one: after the save the dialog is closed (magacinKontoIzmenaResponse).
	btnSave := magacinKontoSaveButton(updateURL, "PUT", dialog.Id, "magacinKontoIzmenaResponse")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	common.SetUnlockButtonProperties(&btnCancel, fmt.Sprintf("/api/magacin-konto/unlock/%d", id))
	common.SetUnlockButtonProperties(&btnClose, fmt.Sprintf("/api/magacin-konto/unlock/%d", id))

	csrfToken := common.GetCsrfToken(c)
	tmpl_robno.MagacinKontoDialog(dialog, common.ActionUpdate, entity, btnSave, btnCancel, btnClose, h.translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *MagacinKontoHandler) GetMagacinKonto(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, SetMagacinKontoFields(), common.IDmagkonto)
}

func (h *MagacinKontoHandler) GetAllMagacinKonto(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	searchText := c.Query("query")
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

	tbl := common.SetTableBasicData(magacinKontoContentTitle, magacinKontoTableID, h.magacinKontoService.GetMagacinKontoTableFields(), "", magacinKontoURLGetAll, pageSize, currentPage, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "", magacinKontoURLGetAll, false, false, false)
	tbl.ShowPagination = true
	tbl.ShowActions = true
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = magacinKontoURLUpdate
	tbl.BtnDelete.HxActionURL = magacinKontoURLDelete
	tbl.BtnAdd.HxActionURL = magacinKontoURLNovi
	tbl.BtnAdd.IsVisible = true
	tbl.URLGetAll = magacinKontoURLGetAll
	tbl.URLPrefix = magacinKontoURLPrefix

	err := h.magacinKontoService.GetAllMagacinKonto(ctx, &tbl, currentPage, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	err = h.magacinKontoService.GetAllMagacinKonto(ctx, &tbl, currentPage, pageSize, false, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	tbl.BtnDelete.IsVisible = false

	if requestSource == "menu" || requestSource == "" {
		searchInput := common.CreateSearchInput("search-input", h.translator, magacinKontoURLGetAll, fmt.Sprintf("#%s", magacinKontoTableID), "")
		btnPrint := common.SetPrintButton("btn-print-magacin-konto", "Štampa", "fin_print", magacinKontoURLPrint, "GET", true, common.ClassPrintButton, "")
		tmpl_robno.MagacinKonto(tbl, searchInput, btnPrint, h.translator).Render(c.Request.Context(), c.Writer)
	} else {
		utils.RenderContent(c, tbl)
	}
}

func (h *MagacinKontoHandler) MagacinKontoStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	searchText := c.Query("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

	fvrData, err := h.magacinKontoService.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}

	repParams := domain.ReportParameters{
		Orientation:    "landscape",
		CompanyName:    fvrData.Naziv,
		Adress:         fvrData.Adresa,
		Postcode:       fvrData.Pobro,
		City:           fvrData.Mesto,
		PIB:            fvrData.PIB,
		MatBroj:        fvrData.Matbr,
		ReportName:     "Magacin Konto",
		ParameterItems: map[string]domain.ParameterItem{},
	}

	tbl := common.SetTableBasicData(magacinKontoContentTitle, magacinKontoTableID, h.magacinKontoService.GetMagacinKontoTableFields(), "", magacinKontoURLGetAll, 0, 0, 0, 0, h.cfg)
	err = h.magacinKontoService.GetAllMagacinKonto(ctx, &tbl, page, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePrint)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}

	tmpl_robno_rep.MagacinKontoStampa(repParams, tbl, h.translator).Render(ctx, c.Writer)
}

func (h *MagacinKontoHandler) UnlockMagacinKonto(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Magacin konto unlocked..."})
}

func (h *MagacinKontoHandler) AddRoutes(r *gin.Engine) {
	r.POST("/api/magacin-konto", h.CreateMagacinKonto)
	r.GET("/api/magacin-konto/all", h.GetAllMagacinKonto)
	r.GET("/api/magacin-konto/stampa", h.MagacinKontoStampa)
	// The grid sends the key of the row (magkontoid) as "id".
	r.GET("/api/magacin-konto/confirm-delete", h.lm.WithEntityLockHold("magkonto", "id"), h.confirmDeleteHandler)
	r.GET("/api/magacin-konto/confirm-update", h.lm.WithEntityLockHold("magkonto", "id"), h.confirmUpdateHandler)
	r.GET("/api/magacin-konto/confirm-add", h.confirmAddHandler)
	r.GET("/api/magacin-konto/:id", h.GetMagacinKonto)
	r.POST("/api/magacin-konto/save", h.CreateMagacinKonto)
	r.PUT("/api/magacin-konto/:id", h.UpdateMagacinKonto)
	r.DELETE("/api/magacin-konto/:id", h.DeleteMagacinKonto)
	r.GET("/api/magacin-konto/unlock/:id", h.UnlockMagacinKonto)
}
