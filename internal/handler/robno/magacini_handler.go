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
	magaciniContentTitle string = "MAGACINI"
	magaciniTableID      string = "magacini-table"
	magaciniURLPrefix    string = "/api/magacini"
	magaciniURLGetAll    string = "/api/magacini/all"
	magaciniURLPrint     string = "/api/magacini/stampa"
	magaciniURLNovi      string = "/api/magacini/confirm-add"
	magaciniURLSave      string = "/api/magacini/save"
	magaciniURLDelete    string = "/api/magacini/confirm-delete"
	magaciniURLUpdate    string = "/api/magacini/confirm-update"
)

func SetMagaciniFields() []domain.Fields {
	return []domain.Fields{
		{Name: "mag", Label: "Magacin", Width: "8"},
		{Name: "nadmag", Label: "Nadređeni", Width: "12"},
		{Name: "tipmag", Label: "Tip", Width: "8"},
		{Name: "opis", Label: "Opis magacina", Width: "30"},
		{Name: "magosoba", Label: "Osoba magacina", Width: "20"},
		{Name: "nacvodzal", Label: "Način vođenja", Width: "12"},
		{Name: "tipcene", Label: "Tip cena", Width: "10"},
	}
}

type MagaciniHandler struct {
	service         service.Service[domain.Magacini]
	magaciniService robnosvc.MagaciniService
	cfg             config.Config
	lm              *middleware.LockMiddleware
}

func NewMagaciniHandler(service *service.BaseService[domain.Magacini], magaciniService robnosvc.MagaciniService, cfg config.Config, lm *middleware.LockMiddleware) *MagaciniHandler {
	return &MagaciniHandler{
		service:         service,
		magaciniService: magaciniService,
		cfg:             cfg,
		lm:              lm,
	}
}

func (h *MagaciniHandler) CreateMagacini(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	dto := domain.Magacini{}
	if err := c.ShouldBind(&dto); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	fieldErrors := h.magaciniService.ValidateEntity(ctx, &dto)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, "Validation errors")
		return
	}
	fields := []domain.Fields{
		{Name: "mag"},
		{Name: "opis"},
		{Name: "tipmag"},
		{Name: "adresa"},
		{Name: "pobro"},
		{Name: "mesto"},
		{Name: "nadmag"},
		{Name: "magosoba"},
		{Name: "tel"},
		{Name: "fax"},
		{Name: "tipzal"},
		{Name: "tipcene"},
		{Name: "nacvodzal"},
		{Name: "analiza"},
		{Name: "email"},
		{Name: "tipart"},
	}

	_, _, err := h.magaciniService.Create(ctx, &dto, common.IDmagacin, fields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

func (h *MagaciniHandler) UpdateMagacini(c *gin.Context) {
	var magacini domain.Magacini
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	getEntity, err := h.magaciniService.GetByID(c.Request.Context(), common.IDmagacin, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if getEntity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgGetData)
		return
	}
	err = c.ShouldBind(&magacini)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	fields := []domain.Fields{
		{Name: "opis", Value: magacini.Opis},
		{Name: "tipmag", Value: magacini.Tipmag},
		{Name: "adresa", Value: magacini.Adresa},
		{Name: "pobro", Value: fmt.Sprintf("%d", magacini.Pobro)},
		{Name: "mesto", Value: magacini.Mesto},
		{Name: "nadmag", Value: fmt.Sprintf("%d", magacini.Nadmag)},
		{Name: "magosoba", Value: magacini.Magosoba},
		{Name: "tel", Value: magacini.Tel},
		{Name: "fax", Value: magacini.Fax},
		{Name: "tipzal", Value: fmt.Sprintf("%d", magacini.Tipzal)},
		{Name: "tipcene", Value: fmt.Sprintf("%d", magacini.Tipcene)},
		{Name: "nacvodzal", Value: fmt.Sprintf("%d", magacini.Nacvodzal)},
		{Name: "analiza", Value: fmt.Sprintf("%d", magacini.Analiza)},
		{Name: "email", Value: magacini.Email},
		{Name: "tipart", Value: magacini.Tipart},
	}
	fieldErrors, err := h.magaciniService.Update(c.Request.Context(), &magacini, common.IDmagacin, id, fields)
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

func (h *MagaciniHandler) DeleteMagacini(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	err = h.magaciniService.Delete(ctx, common.IDmagacin, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

func (h *MagaciniHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, SetMagaciniFields(), "#info-message")
}

func (h *MagaciniHandler) confirmAddHandler(c *gin.Context) {
	dialog := domain.Dialog{
		Title:         "Novi magacin",
		Id:            "magacini-add-form",
		HxActionURL:   magaciniURLSave,
		HxRequestType: "POST",
		OkText:        "Sačuvaj",
		CancelText:    "Odustani",
	}

	model := domain.Magacini{
		Tipcene:   0,
		Nacvodzal: 1,
		Tipzal:    1,
		Analiza:   0,
	}

	btnSave := common.SetButton("btn-save-magacini", "Sačuvaj", "save", magaciniURLSave, "", "", "POST", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)

	tmpl_robno.MagaciniDialog(dialog, common.ActionAdd, model, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *MagaciniHandler) confirmUpdateHandler(c *gin.Context) {
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
	entity, err := h.magaciniService.GetByID(ctx, common.IDmagacin, int64(id))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	updateURL := fmt.Sprintf("/api/magacini/%d", id)
	dialog := domain.Dialog{
		Title:         "Izmena magacina",
		Id:            "magacini-edit-form",
		HxActionURL:   updateURL,
		HxRequestType: "PUT",
	}

	btnSave := common.SetButton("btn-save-magacini", "Sačuvaj", "save", updateURL, "", "", "PUT", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	common.SetUnlockButtonProperties(&btnCancel, fmt.Sprintf("/api/magacini/unlock/%d", id))
	common.SetUnlockButtonProperties(&btnClose, fmt.Sprintf("/api/magacini/unlock/%d", id))

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)
	tmpl_robno.MagaciniDialog(dialog, common.ActionUpdate, *entity, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *MagaciniHandler) GetMagacini(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, SetMagaciniFields(), common.IDmagacin)
}

func (h *MagaciniHandler) GetAllMagacini(c *gin.Context) {
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

	// btnPrint := common.SetButton("stampa-btn", "Štampa", "fin_print", magaciniURLPrint, "#dialog-magacini-stampa", "outerHTML", "GET", "", "", true, common.ClassPrintButton, "")
	// searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), magaciniURLGetAll, fmt.Sprintf("#%s", magaciniTableID), "")

	tbl := common.SetTableBasicData(magaciniContentTitle, magaciniTableID, h.magaciniService.GetMagaciniTableFields(), "", magaciniURLGetAll, pageSize, currentPage, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "MAGACINI", magaciniURLGetAll, true, true, false)
	tbl.ShowPagination = true
	tbl.ShowActions = true
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = magaciniURLUpdate
	tbl.BtnDelete.HxActionURL = magaciniURLDelete
	tbl.BtnAdd.HxActionURL = magaciniURLNovi
	tbl.BtnAdd.IsVisible = true
	tbl.URLGetAll = magaciniURLGetAll
	tbl.URLPrefix = magaciniURLPrefix

	err := h.magaciniService.GetAllMagacini(ctx, &tbl, currentPage, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	err = h.magaciniService.GetAllMagacini(ctx, &tbl, currentPage, pageSize, false, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	tbl.BtnDelete.IsVisible = false

	if requestSource == "menu" || requestSource == "" {
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), magaciniURLGetAll, fmt.Sprintf("#%s", magaciniTableID), "")
		btnPrint := common.SetPrintButton("btn-print-fkpl", "Štampa", "fin_print", magaciniURLPrint, "GET", true, common.ClassPrintButton, "")
		tmpl_robno.Magacini(tbl, searchInput, btnPrint, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	} else {
		utils.RenderContent(c, tbl)
	}
}

func (h *MagaciniHandler) MagaciniStampa(c *gin.Context) {
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

	fvrData, err := h.magaciniService.GetFvrData(ctx)
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
		ReportName:     "Magacini",
		ParameterItems: map[string]domain.ParameterItem{},
	}

	tbl := common.SetTableBasicData(magaciniContentTitle, magaciniTableID, h.magaciniService.GetMagaciniTableFields(), "", magaciniURLGetAll, 0, 0, 0, 0, h.cfg)
	err = h.magaciniService.GetAllMagacini(ctx, &tbl, page, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePrint)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}

	translator := i18n.GetInstance()
	tmpl_robno_rep.MagaciniStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *MagaciniHandler) UnlockMagacini(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Magacini unlocked..."})
}

func (h *MagaciniHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.POST("/api/magacini", h.CreateMagacini)
	r.GET("/api/magacini/all", h.GetAllMagacini)
	r.GET("/api/magacini/stampa", h.MagaciniStampa)
	r.GET("/api/magacini/confirm-delete", h.lm.WithEntityLockHold("magacini", "id"), h.confirmDeleteHandler)
	r.GET("/api/magacini/confirm-update", h.lm.WithEntityLockHold("magacini", "id"), h.confirmUpdateHandler)
	r.GET("/api/magacini/confirm-add", h.confirmAddHandler)
	r.GET("/api/magacini/:id", h.GetMagacini)
	r.POST("/api/magacini/save", h.CreateMagacini)
	r.PUT("/api/magacini/:id", h.UpdateMagacini)
	r.DELETE("/api/magacini/:id", h.DeleteMagacini)
	r.GET("/api/magacini/unlock/:id", h.UnlockMagacini)
}
