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
	service             service.Service[domain.Magkonto]
	magacinKontoService robnosvc.MagacinKontoService
	cfg                 config.Config
	lm                  *middleware.LockMiddleware
}

func NewMagacinKontoHandler(service *service.BaseService[domain.Magkonto], magacinKontoService robnosvc.MagacinKontoService, cfg config.Config, lm *middleware.LockMiddleware) *MagacinKontoHandler {
	return &MagacinKontoHandler{
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
	fieldErrors := h.magacinKontoService.ValidateEntity(ctx, &dto)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, "Validation errors")
		return
	}
	fields := []domain.Fields{
		{Name: "mag"},
		{Name: "konto"},
		{Name: "idfkpl"},
		{Name: "vkonta"},
		{Name: "kontoprih"},
		{Name: "kontotroska"},
		{Name: "kontoruc"},
		{Name: "kontorab"},
	}

	_, _, err := h.magacinKontoService.Create(ctx, &dto, common.IDmagacin, fields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

func (h *MagacinKontoHandler) UpdateMagacinKonto(c *gin.Context) {
	var magacinKonto domain.Magkonto
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	getEntity, err := h.magacinKontoService.GetByID(c.Request.Context(), common.IDmagacin, id)
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
	fields := []domain.Fields{
		{Name: "konto", Value: magacinKonto.Konto},
		{Name: "idfkpl", Value: fmt.Sprintf("%d", magacinKonto.IDFkpl)},
		{Name: "vkonta", Value: fmt.Sprintf("%d", magacinKonto.VKonta)},
		{Name: "kontoprih", Value: magacinKonto.KontoPrih},
		{Name: "kontotroska", Value: magacinKonto.KontoTroska},
		{Name: "kontoruc", Value: magacinKonto.KontoRuc},
		{Name: "kontorab", Value: magacinKonto.KontoRab},
	}
	fieldErrors, err := h.magacinKontoService.Update(c.Request.Context(), &magacinKonto, common.IDmagacin, id, fields)
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

func (h *MagacinKontoHandler) DeleteMagacinKonto(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	err = h.magacinKontoService.Delete(ctx, common.IDmagacin, id)
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

	model := domain.Magkonto{
		VKonta: 1,
	}

	btnSave := common.SetButton("btn-save-magacin-konto", "Sačuvaj", "save", magacinKontoURLSave, "", "", "POST", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)

	tmpl_robno.MagacinKontoDialog(dialog, common.ActionAdd, model, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
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
	entity, err := h.magacinKontoService.GetByID(ctx, common.IDmagacin, int64(id))
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

	btnSave := common.SetButton("btn-save-magacin-konto", "Sačuvaj", "save", updateURL, "", "", "PUT", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	common.SetUnlockButtonProperties(&btnCancel, fmt.Sprintf("/api/magacin-konto/unlock/%d", id))
	common.SetUnlockButtonProperties(&btnClose, fmt.Sprintf("/api/magacin-konto/unlock/%d", id))

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)
	tmpl_robno.MagacinKontoDialog(dialog, common.ActionUpdate, *entity, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *MagacinKontoHandler) GetMagacinKonto(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, SetMagacinKontoFields(), common.IDmagacin)
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
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), magacinKontoURLGetAll, fmt.Sprintf("#%s", magacinKontoTableID), "")
		btnPrint := common.SetPrintButton("btn-print-magacin-konto", "Štampa", "fin_print", magacinKontoURLPrint, "GET", true, common.ClassPrintButton, "")
		tmpl_robno.MagacinKonto(tbl, searchInput, btnPrint, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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

	translator := i18n.GetInstance()
	tmpl_robno_rep.MagacinKontoStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *MagacinKontoHandler) UnlockMagacinKonto(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Magacin konto unlocked..."})
}

func (h *MagacinKontoHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.POST("/api/magacin-konto", h.CreateMagacinKonto)
	r.GET("/api/magacin-konto/all", h.GetAllMagacinKonto)
	r.GET("/api/magacin-konto/stampa", h.MagacinKontoStampa)
	r.GET("/api/magacin-konto/confirm-delete", h.lm.WithEntityLockHold("magkonto", "magaciniid"), h.confirmDeleteHandler)
	r.GET("/api/magacin-konto/confirm-update", h.lm.WithEntityLockHold("magkonto", "magaciniid"), h.confirmUpdateHandler)
	r.GET("/api/magacin-konto/confirm-add", h.confirmAddHandler)
	r.GET("/api/magacin-konto/:id", h.GetMagacinKonto)
	r.POST("/api/magacin-konto/save", h.CreateMagacinKonto)
	r.PUT("/api/magacin-konto/:id", h.UpdateMagacinKonto)
	r.DELETE("/api/magacin-konto/:id", h.DeleteMagacinKonto)
	r.GET("/api/magacin-konto/unlock/:id", h.UnlockMagacinKonto)
}
