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
	"helia/internal/handler"
	"helia/internal/middleware"
	"helia/internal/service"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	komercijalistiContentTitle string = "KOMERCIJALISTI"
	komercijalistiTableID      string = "komercijalisti-table"
	komercijalistiURLPrefix    string = "/api/komercijalisti"
	komercijalistiURLGetAll    string = "/api/komercijalisti/all"
	komercijalistiURLPrint     string = "/api/komercijalisti/stampa"
	komercijalistiURLNovi      string = "/api/komercijalisti/confirm-add"
	komercijalistiURLSave      string = "/api/komercijalisti/save"
	komercijalistiURLDelete    string = "/api/komercijalisti/confirm-delete"
	komercijalistiURLUpdate    string = "/api/komercijalisti/confirm-update"
)

type KomercijalistiHandler struct {
	service               service.Service[domain.Komercijalisti]
	komercijalistiService robnosvc.KomercijalistiService
	cfg                   config.Config
	lm                    *middleware.LockMiddleware
}

func NewKomercijalistiHandler(service *service.BaseService[domain.Komercijalisti], komercijalistiService robnosvc.KomercijalistiService, cfg config.Config, lm *middleware.LockMiddleware) *KomercijalistiHandler {
	return &KomercijalistiHandler{
		service:               service,
		komercijalistiService: komercijalistiService,
		cfg:                   cfg,
		lm:                    lm,
	}
}

func (h *KomercijalistiHandler) CreateKomercijalisti(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	dto := domain.Komercijalisti{}
	if err := c.ShouldBind(&dto); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	fieldErrors := h.komercijalistiService.ValidateEntity(ctx, &dto)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, "Validation errors")
		return
	}
	fields := []domain.Fields{
		{Name: "sifkom"},
		{Name: "sifnadred"},
		{Name: "imeprezime"},
		{Name: "adresa"},
		{Name: "mesto"},
		{Name: "telposao"},
		{Name: "telmob"},
		{Name: "totprod"},
		{Name: "totprofit"},
		{Name: "zaddatprod"},
		{Name: "totnaplaceno"},
		{Name: "loginname"},
	}

	_, _, err := h.komercijalistiService.Create(ctx, &dto, common.IDkomercijalista, fields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

func (h *KomercijalistiHandler) UpdateKomercijalisti(c *gin.Context) {
	var komercijalisti domain.Komercijalisti
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	getEntity, err := h.komercijalistiService.GetByID(c.Request.Context(), common.IDkomercijalista, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if getEntity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgGetData)
		return
	}
	err = c.ShouldBind(&komercijalisti)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	fields := []domain.Fields{
		{Name: "sifkom", Value: fmt.Sprintf("%d", komercijalisti.Sifkom)},
		{Name: "sifnadred", Value: fmt.Sprintf("%d", komercijalisti.SifNadred)},
		{Name: "imeprezime", Value: komercijalisti.ImePrezime},
		{Name: "adresa", Value: komercijalisti.Adresa},
		{Name: "mesto", Value: komercijalisti.Mesto},
		{Name: "telposao", Value: komercijalisti.TelPosao},
		{Name: "telmob", Value: komercijalisti.TelMob},
		{Name: "totprod", Value: fmt.Sprintf("%f", komercijalisti.TotProd)},
		{Name: "totprofit", Value: fmt.Sprintf("%f", komercijalisti.TotProfit)},
		{Name: "zaddatprod", Value: komercijalisti.ZadDatProd.Format("2006-01-02")},
		{Name: "totnaplaceno", Value: fmt.Sprintf("%f", komercijalisti.TotNaplaceno)},
		{Name: "loginname", Value: komercijalisti.LoginName},
	}
	fieldErrors, err := h.komercijalistiService.Update(c.Request.Context(), &komercijalisti, common.IDkomercijalista, id, fields)
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

func (h *KomercijalistiHandler) DeleteKomercijalisti(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	err = h.komercijalistiService.Delete(ctx, common.IDkomercijalista, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

func (h *KomercijalistiHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, handler.SetKomercijalistiFileds(), "#info-message")
}

func (h *KomercijalistiHandler) confirmAddHandler(c *gin.Context) {
	dialog := domain.Dialog{
		Title:         "Novi komercijalista",
		Id:            "komercijalisti-add-form",
		HxActionURL:   komercijalistiURLSave,
		HxRequestType: "POST",
		OkText:        "Sačuvaj",
		CancelText:    "Odustani",
	}

	model := domain.Komercijalisti{
		TotProd:      0,
		TotProfit:    0,
		TotNaplaceno: 0,
	}

	btnSave := common.SetButton("btn-save-komercijalisti", "Sačuvaj", "save", komercijalistiURLSave, "", "", "POST", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)

	tmpl_robno.KomercijalistiDialog(dialog, common.ActionAdd, model, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *KomercijalistiHandler) confirmUpdateHandler(c *gin.Context) {
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
	entity, err := h.komercijalistiService.GetByID(ctx, common.IDkomercijalista, int64(id))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	updateURL := fmt.Sprintf("/api/komercijalisti/%d", id)
	dialog := domain.Dialog{
		Title:         "Izmena komercijaliste",
		Id:            "komercijalisti-edit-form",
		HxActionURL:   updateURL,
		HxRequestType: "PUT",
	}

	btnSave := common.SetButton("btn-save-komercijalisti", "Sačuvaj", "save", updateURL, "", "", "PUT", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	common.SetUnlockButtonProperties(&btnCancel, fmt.Sprintf("/api/komercijalisti/unlock/%d", id))
	common.SetUnlockButtonProperties(&btnClose, fmt.Sprintf("/api/komercijalisti/unlock/%d", id))

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)
	tmpl_robno.KomercijalistiDialog(dialog, common.ActionUpdate, *entity, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *KomercijalistiHandler) GetKomercijalisti(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, handler.SetKomercijalistiFileds(), common.IDkomercijalista)
}

func (h *KomercijalistiHandler) GetAllKomercijalisti(c *gin.Context) {
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

	tbl := common.SetTableBasicData(komercijalistiContentTitle, komercijalistiTableID, h.komercijalistiService.GetKomercijalistiTableFields(), "", komercijalistiURLGetAll, pageSize, currentPage, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "KOMERCIJALISTI", komercijalistiURLGetAll, true, true, false)
	tbl.ShowPagination = true
	tbl.ShowActions = true
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = komercijalistiURLUpdate
	tbl.BtnDelete.HxActionURL = komercijalistiURLDelete
	tbl.BtnAdd.HxActionURL = komercijalistiURLNovi
	tbl.BtnAdd.IsVisible = true
	tbl.URLGetAll = komercijalistiURLGetAll
	tbl.URLPrefix = komercijalistiURLPrefix

	err := h.komercijalistiService.GetAllKomercijalisti(ctx, &tbl, currentPage, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	err = h.komercijalistiService.GetAllKomercijalisti(ctx, &tbl, currentPage, pageSize, false, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	tbl.BtnDelete.IsVisible = false

	if requestSource == "menu" || requestSource == "" {
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), komercijalistiURLGetAll, fmt.Sprintf("#%s", komercijalistiTableID), "")
		btnPrint := common.SetPrintButton("btn-print-komercijalisti", "Štampa", "fin_print", komercijalistiURLPrint, "GET", true, common.ClassPrintButton, "")
		tmpl_robno.KomercijalistiMain(tbl, searchInput, btnPrint, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	} else {
		utils.RenderContent(c, tbl)
	}
}

func (h *KomercijalistiHandler) KomercijalistiStampa(c *gin.Context) {
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

	fvrData, err := h.komercijalistiService.GetFvrData(ctx)
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
		ReportName:     "Komercijalisti",
		ParameterItems: map[string]domain.ParameterItem{},
	}

	tbl := common.SetTableBasicData(komercijalistiContentTitle, komercijalistiTableID, h.komercijalistiService.GetKomercijalistiTableFields(), "", komercijalistiURLGetAll, 0, 0, 0, 0, h.cfg)
	err = h.komercijalistiService.GetAllKomercijalisti(ctx, &tbl, page, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePrint)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}

	translator := i18n.GetInstance()
	tmpl_robno_rep.KomercijalistiStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *KomercijalistiHandler) UnlockKomercijalisti(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Komercijalisti unlocked..."})
}

func (h *KomercijalistiHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.POST("/api/komercijalisti", h.CreateKomercijalisti)
	r.GET("/api/komercijalisti/all", h.GetAllKomercijalisti)
	r.GET("/api/komercijalisti/stampa", h.KomercijalistiStampa)
	r.GET("/api/komercijalisti/confirm-add", h.confirmAddHandler)
	r.GET("/api/komercijalisti/confirm-update", h.confirmUpdateHandler)
	r.GET("/api/komercijalisti/confirm-delete", h.confirmDeleteHandler)
	r.PUT("/api/komercijalisti/:id", h.UpdateKomercijalisti)
	r.POST("/api/komercijalisti/save", h.CreateKomercijalisti)
	r.DELETE("/api/komercijalisti/:id", h.DeleteKomercijalisti)
	r.GET("/api/komercijalisti/:id", h.GetKomercijalisti)
	r.GET("/api/komercijalisti/unlock/:id", h.UnlockKomercijalisti)
}
