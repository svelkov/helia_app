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
	artikliContentTitle string = "ARTIKLI"
	artikliTableID      string = "artikli-table"
	artikliURLPrefix    string = "/api/artikli"
	artikliURLGetAll    string = "/api/artikli/all"
	artikliURLPrint     string = "/api/artikli/stampa"
	artikliURLNovi      string = "/api/artikli/confirm-add"
	artikliURLSave      string = "/api/artikli/save"
	artikliURLDelete    string = "/api/artikli/confirm-delete"
	artikliURLUpdate    string = "/api/artikli/confirm-update"
)

type ArtikliHandler struct {
	service        service.Service[domain.Rsif]
	artikliService robnosvc.ArtikliService
	cfg            config.Config
	lm             *middleware.LockMiddleware
}

func NewArtikliHandler(
	service *service.BaseService[domain.Rsif],
	artikliService robnosvc.ArtikliService,
	cfg config.Config,
	lm *middleware.LockMiddleware,
) *ArtikliHandler {
	return &ArtikliHandler{
		service:        service,
		artikliService: artikliService,
		cfg:            cfg,
		lm:             lm,
	}
}

func (h *ArtikliHandler) CreateArtikli(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	dto := domain.Rsif{}
	if err := c.ShouldBind(&dto); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	fieldErrors := h.artikliService.ValidateEntity(ctx, &dto)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, "Validation errors")
		return
	}
	fields := []domain.Fields{
		{Name: "sifra"},
		{Name: "naziv"},
		{Name: "komercopis"},
		{Name: "pro"},
		{Name: "jm"},
		{Name: "konto"},
		{Name: "model"},
		{Name: "tip"},
		{Name: "barkod"},
		{Name: "kvalitet"},
		{Name: "miza"},
		{Name: "maza"},
		{Name: "serbr"},
		{Name: "zemljaproizv"},
	}

	_, _, err := h.artikliService.Create(ctx, &dto, common.IDrsif, fields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

func (h *ArtikliHandler) UpdateArtikli(c *gin.Context) {
	var artikli domain.Rsif
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	getEntity, err := h.artikliService.GetByID(c.Request.Context(), common.IDrsif, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if getEntity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgGetData)
		return
	}
	err = c.ShouldBind(&artikli)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	fields := []domain.Fields{
		{Name: "sifra", Value: fmt.Sprintf("%d", artikli.Sifra)},
		{Name: "naziv", Value: artikli.Naziv},
		{Name: "komercopis", Value: artikli.KomercOpis},
		{Name: "pro", Value: artikli.Pro},
		{Name: "jm", Value: artikli.JM},
		{Name: "konto", Value: artikli.Konto},
		{Name: "model", Value: artikli.Model},
		{Name: "tip", Value: artikli.Tip},
		{Name: "barkod", Value: artikli.Barkod},
		{Name: "kvalitet", Value: artikli.Kvalitet},
		{Name: "miza", Value: fmt.Sprintf("%f", artikli.Miza)},
		{Name: "maza", Value: fmt.Sprintf("%f", artikli.Maza)},
		{Name: "serbr", Value: artikli.Serbr},
		{Name: "zemljaproizv", Value: artikli.ZemljaProizv},
	}
	fieldErrors, err := h.artikliService.Update(c.Request.Context(), &artikli, common.IDrsif, id, fields)
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

func (h *ArtikliHandler) DeleteArtikli(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	err = h.artikliService.Delete(ctx, common.IDrsif, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

func (h *ArtikliHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, handler.SetArtikliFields(), "#info-message")
}

func (h *ArtikliHandler) confirmAddHandler(c *gin.Context) {
	dialog := domain.Dialog{
		Title:         "Novi artikal",
		Id:            "artikli-add-form",
		HxActionURL:   artikliURLSave,
		HxRequestType: "POST",
		OkText:        "Sačuvaj",
		CancelText:    "Odustani",
	}

	model := domain.Rsif{
		Miza: 0,
		Maza: 0,
	}

	btnSave := common.SetButton("btn-save-artikli", "Sačuvaj", "save", artikliURLSave, "", "", "POST", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)

	// Empty combo options for now
	modelCombo := []domain.ComboItem{}
	jmCombo := []domain.ComboItem{}
	gruCombo := []domain.ComboItem{}
	pgruCombo := []domain.ComboItem{}

	tmpl_robno.ArtikliDialog(dialog, common.ActionAdd, model, btnSave, btnCancel, btnClose, modelCombo, jmCombo, gruCombo, pgruCombo, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *ArtikliHandler) confirmUpdateHandler(c *gin.Context) {
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
	entity, err := h.artikliService.GetByID(ctx, common.IDrsif, int64(id))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	updateURL := fmt.Sprintf("/api/artikli/%d", id)
	dialog := domain.Dialog{
		Title:         "Izmena artikla",
		Id:            "artikli-edit-form",
		HxActionURL:   updateURL,
		HxRequestType: "PUT",
	}

	btnSave := common.SetButton("btn-save-artikli", "Sačuvaj", "save", updateURL, "", "", "PUT", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	common.SetUnlockButtonProperties(&btnCancel, fmt.Sprintf("/api/artikli/unlock/%d", id))
	common.SetUnlockButtonProperties(&btnClose, fmt.Sprintf("/api/artikli/unlock/%d", id))

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)

	// Empty combo options for now
	modelCombo := []domain.ComboItem{}
	jmCombo := []domain.ComboItem{}
	gruCombo := []domain.ComboItem{}
	pgruCombo := []domain.ComboItem{}

	tmpl_robno.ArtikliDialog(dialog, common.ActionUpdate, *entity, btnSave, btnCancel, btnClose, modelCombo, jmCombo, gruCombo, pgruCombo, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *ArtikliHandler) GetArtikli(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, handler.SetArtikliFields(), common.IDrsif)
}

func (h *ArtikliHandler) GetAllArtikli(c *gin.Context) {
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

	tbl := common.SetTableBasicData(artikliContentTitle, artikliTableID, h.artikliService.GetArtikliTableFields(), "", artikliURLGetAll, pageSize, currentPage, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "ARTIKLI", artikliURLGetAll, true, true, false)
	tbl.ShowPagination = true
	tbl.ShowActions = true
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = artikliURLUpdate
	tbl.BtnDelete.HxActionURL = artikliURLDelete
	tbl.BtnAdd.HxActionURL = artikliURLNovi
	tbl.BtnAdd.IsVisible = true
	tbl.URLGetAll = artikliURLGetAll
	tbl.URLPrefix = artikliURLPrefix

	err := h.artikliService.GetAllArtikli(ctx, &tbl, currentPage, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	err = h.artikliService.GetAllArtikli(ctx, &tbl, currentPage, pageSize, false, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	tbl.BtnDelete.IsVisible = false

	if requestSource == "menu" || requestSource == "" {
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), artikliURLGetAll, fmt.Sprintf("#%s", artikliTableID), "")
		btnPrint := common.SetPrintButton("btn-print-artikli", "Štampa", "fin_print", artikliURLPrint, "GET", true, common.ClassPrintButton, "")
		tmpl_robno.ArtikliMain(tbl, searchInput, btnPrint, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	} else {
		utils.RenderContent(c, tbl)
	}
}

func (h *ArtikliHandler) ArtikliStampa(c *gin.Context) {
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

	fvrData, err := h.artikliService.GetFvrData(ctx)
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
		ReportName:     "ARTIKLI - DETALJNI PRIKAZ",
		ParameterItems: map[string]domain.ParameterItem{},
	}

	tbl := common.SetTableBasicData(artikliContentTitle, artikliTableID, h.artikliService.GetArtikliTableFields(), "", artikliURLGetAll, 0, 0, 0, 0, h.cfg)
	err = h.artikliService.GetAllArtikli(ctx, &tbl, page, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePrint)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}

	translator := i18n.GetInstance()
	tmpl_robno_rep.ArtikliStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *ArtikliHandler) UnlockArtikli(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Artikli unlocked..."})
}

func (h *ArtikliHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.POST("/api/artikli", h.CreateArtikli)
	r.GET("/api/artikli/all", h.GetAllArtikli)
	r.GET("/api/artikli/stampa", h.ArtikliStampa)
	r.GET("/api/artikli/confirm-add", h.confirmAddHandler)
	r.GET("/api/artikli/confirm-update", h.confirmUpdateHandler)
	r.GET("/api/artikli/confirm-delete", h.confirmDeleteHandler)
	r.PUT("/api/artikli/:id", h.UpdateArtikli)
	r.POST("/api/artikli/save", h.CreateArtikli)
	r.DELETE("/api/artikli/:id", h.DeleteArtikli)
	r.GET("/api/artikli/:id", h.GetArtikli)
	r.GET("/api/artikli/unlock/:id", h.UnlockArtikli)
}
