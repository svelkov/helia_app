package finansijsko

import (
	"fmt"
	"net/http"
	"strconv"

	"helia/config"
	tmpl_fin "helia/frontend/templates/finansijsko"
	tmpl_fin_rep "helia/frontend/templates/reports/finansijsko"
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
	fkplContentTitle   string = "KONTNI PLAN"
	fkplTableID        string = "fkpl-table"
	searchKontoTableID string = "searchkonto-table"
	fkplURLPrefix      string = "/api/fkpl"
	fkplURLGetAll      string = "/api/fkpl/all"
	fkplURLPrint       string = "/api/fkpl/stampa"
	fkplURLNovi        string = "/api/fkpl/confirm-add"
	fkplURLSave        string = "/api/fkpl/save"
	fkplURLDelete      string = "/api/fkpl/confirm-delete"
	fkplURLUpdate      string = "/api/fkpl/confirm-update"
)

const hxValsFkpl = `js:{
	"vkonta": document.querySelector('input[name="vkonta"]:checked')?.value || "",
	"query": document.getElementById("search-input")?.value || ""
}`

func SetFkplFields() []domain.Fields {
	return []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Sifra", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "120"},
		{Name: "vkonta", Label: "Vrsta konta", Width: "4"},
	}
}

var fkplSearchTableFields = []domain.Fields{
	{Name: "konto", Label: "Konto", Width: "10"},
	{Name: "sifra", Label: "Sifra", Width: "10"},
	{Name: "naziv", Label: "Naziv", Width: "120"},
}

type FkplHandler struct {
	service     service.Service[domain.Fkpl]
	fkplService finservice.FkplService
	cfg         config.Config
	lm          *middleware.LockMiddleware
}

func NewFkplHandler(service *service.BaseService[domain.Fkpl], fkplService finservice.FkplService, cfg config.Config, lm *middleware.LockMiddleware) *FkplHandler {
	return &FkplHandler{service: service, fkplService: fkplService, cfg: cfg, lm: lm}
}

func (h *FkplHandler) CreateFkpl(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	dto := domain.Fkpl{}
	if err := c.ShouldBind(&dto); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	filedErrors := h.fkplService.ValidateEntity(ctx, &dto)
	if len(filedErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, filedErrors, "Validation errors")
		return
	}
	fields := []domain.Fields{
		{Name: "konto"},
		{Name: "tipanalitikeid"},
		{Name: "vkonta"},
		{Name: "naziv"},
		{Name: "devizni"},
		{Name: "kolicinski"},
	}

	_, _, err := h.fkplService.Create(ctx, &dto, common.IDfkpl, fields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

func (h *FkplHandler) UpdateFkpl(c *gin.Context) {
	var fkpl domain.Fkpl
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}
	getEntity, err := h.fkplService.GetByID(c.Request.Context(), common.IDfkpl, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if getEntity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgGetData)
		return
	}
	err = c.ShouldBind(&fkpl)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid request body: "+err.Error())
		return
	}
	fields := []domain.Fields{
		{Name: "naziv", Value: fkpl.Naziv},
		{Name: "devizni", Value: fmt.Sprintf("%t", fkpl.Devizni)},
		{Name: "kolicinski", Value: fmt.Sprintf("%t", fkpl.Kolicinski)},
	}
	fieldErrors, err := h.fkplService.Update(c.Request.Context(), &fkpl, common.IDfkpl, id, fields)
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

func (h *FkplHandler) DeleteFkpl(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	err = h.fkplService.Delete(ctx, common.IDfkpl, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

func (h *FkplHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, SetFkplFields(), "#info-message")
}

func (h *FkplHandler) confirmAddHandler(c *gin.Context) {
	ctx := c.Request.Context()
	dialog := domain.Dialog{
		Title:         "Novi konto",
		Id:            "fkpl-add-form",
		HxActionURL:   fkplURLSave,
		HxRequestType: "POST",
		OkText:        "Sačuvaj",
		CancelText:    "Odustani",
	}
	tipAnalitike, err := h.fkplService.GetAnalitikaForSelect(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	model := domain.Fkpl{} // Set default value for TipAnalitikeID
	if len(tipAnalitike) > 0 {
		id := common.StringToInt64(tipAnalitike[0].Key)
		model.TipanalitikeID = &id
	}
	model.Vkonta = 1 // Set default value for Vkonta to 1 (analiticki konto)

	btnSave := common.SetButton("btn-save-fkpl", "Sačuvaj", "save", fkplURLSave, "", "", "POST", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)

	tmpl_fin.KontniPlanDialog(dialog, common.ActionAdd, tipAnalitike, model, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *FkplHandler) confirmUpdateHandler(c *gin.Context) {
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
	// check lock
	entity, err := h.fkplService.GetByID(ctx, common.IDfkpl, int64(id))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	tipAnalitike, err := h.fkplService.GetAnalitikaForSelect(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	updateURL := fmt.Sprintf("/api/fkpl/%d", id)
	dialog := domain.Dialog{
		Title:         "Izmena konta",
		Id:            "fkpl-edit-form",
		HxActionURL:   updateURL,
		HxRequestType: "PUT",
	}

	btnSave := common.SetButton("btn-save-fkpl", "Sačuvaj", "save", updateURL, "", "", "PUT", "#info-message", "", true, common.ClassSaveButton, "")
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = dialog.Id
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnCancel.IdDialog = dialog.Id

	common.SetUnlockButtonProperties(&btnCancel, fmt.Sprintf("/api/fkpl/unlock/%d", id))
	common.SetUnlockButtonProperties(&btnClose, fmt.Sprintf("/api/fkpl/unlock/%d", id))

	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)
	tmpl_fin.KontniPlanDialog(dialog, common.ActionUpdate, tipAnalitike, *entity, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *FkplHandler) GetFkpl(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, SetFkplFields(), common.IDfkpl)
}

func (h *FkplHandler) GetAllFkpl(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	vkonta := c.Query("vkonta")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	searchText := c.Query("query")
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

	tbl := common.SetTableBasicData(fkplContentTitle, fkplTableID, h.fkplService.GetFkplTableFields(), "", fkplURLGetAll, pageSize, currentPage, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "", fkplURLGetAll, false, false, false)
	tbl.ShowPagination = true
	tbl.ShowActions = true
	tbl.Pagination.HxVals = hxValsFkpl
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = fkplURLUpdate
	tbl.BtnDelete.HxActionURL = fkplURLDelete
	tbl.BtnAdd.HxActionURL = fkplURLNovi
	tbl.BtnAdd.IsVisible = true
	tbl.HxVals = hxValsFkpl
	tbl.URLGetAll = fkplURLGetAll
	tbl.URLPrefix = fkplURLPrefix

	err := h.fkplService.GetAllByVkonta(ctx, vkonta, &tbl, currentPage, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	err = h.fkplService.GetAllByVkonta(ctx, vkonta, &tbl, currentPage, pageSize, false, sortBy, sortOrder, searchText, common.TipStampePreview)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	tbl.BtnDelete.IsVisible = false // Hide Delete button in the table header
	
	if requestSource == "menu" || requestSource == "" {
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), fkplURLGetAll, fmt.Sprintf("#%s", fkplTableID), hxValsFkpl)
		btnPrint := common.SetPrintButton("btn-print-fkpl", "Štampa", "fin_print", fkplURLPrint, "GET", true, common.ClassPrintButton, "vkonta")
		tmpl_fin.KontniPlan(tbl, searchInput, btnPrint, vkonta, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	} else {
		utils.RenderContent(c, tbl)
	}

}

// KontniPlanStampa renders a printable view of the kontni plan.
func (h *FkplHandler) KontniPlanStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	vkonta := c.Query("vkonta")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	searchText := c.Query("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	fvrData, err := h.fkplService.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  "Kontni Plan",
		ParameterItems: map[string]domain.ParameterItem{
			"VKonta": {Name: "Vrsta konta", Value: vkonta},
		},
	}

	tbl := common.SetTableBasicData(fkplContentTitle, fkplTableID, h.fkplService.GetFkplTableFields(), "", fkplURLGetAll, 0, 0, 0, 0, h.cfg)
	err = h.fkplService.GetAllByVkonta(ctx, vkonta, &tbl, page, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePrint)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	translator := i18n.GetInstance()
	tmpl_fin_rep.KontniPlanStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *FkplHandler) TraziKonto(c *gin.Context) {
	ctx := c.Request.Context()
	konto := c.Query("konto")
	sifra := c.Query("sifra")
	vkonta := c.Query("vkonta")
	entities, err := h.fkplService.TraziKonto(ctx, konto, sifra, vkonta)
	if err != nil {
		c.Writer.Header().Set("Content-Type", "text/plain")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	// Check if pointer is not nil and slice is not empty
	if entities != nil && len(*entities) > 0 {
		firstElement := (*entities)[0].Naziv
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.Write([]byte(firstElement))
		return
	}

}
func (h *FkplHandler) TraziKontoSearchTable(c *gin.Context) {
	// Parse query parameters
	tbl := common.SetTableBasicData("", searchKontoTableID, fkplSearchTableFields, "", "", 0, 0, 0, 0, h.cfg)
	searchValue := c.Query("query")
	konto := c.Query("konto")
	vkonta := c.Query("vkonta")
	fieldName := c.Query("fieldName")
	fieldColIndex := 0 // Default to first column if not provided
	if vkonta == "2" {
		konto = ""
	}
	if vkonta == "1" {
		fieldColIndex = 1 // If searching by sifra, the code value is in the second column
	}
	err := h.fkplService.KontoSearchForTable(c.Request.Context(), &tbl, searchValue, konto, vkonta, fieldName, fieldColIndex)
	if err != nil {
		c.Writer.Header().Set("Content-Type", "text/plain")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	utils.RenderContent(c, tbl)
}

// UnlockFkpl releases the lock on a fkpl
func (h *FkplHandler) UnlockFkpl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Fkpl unlocked..."})
}
func (h *FkplHandler) AddRoutes(r *gin.Engine) {
	// Create API group with prefix
	//api := r.Group(fkplURLPrefix)
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for fkpl
	r.POST("/api/fkpl", h.CreateFkpl)
	r.GET("/api/fkpl/all", h.GetAllFkpl)
	r.GET("/api/fkpl/stampa", h.KontniPlanStampa)
	r.GET("/api/fkpl/confirm-delete", h.lm.WithEntityLockHold("fkpl", "id"), h.confirmDeleteHandler)
	r.GET("/api/fkpl/confirm-update", h.lm.WithEntityLockHold("fkpl", "id"), h.confirmUpdateHandler)
	r.GET("/api/fkpl/confirm-add", h.confirmAddHandler)
	r.GET("/api/fkpl/:id", h.GetFkpl)
	r.POST("/api/fkpl/save", h.CreateFkpl)
	r.PUT("/api/fkpl/:id", h.UpdateFkpl)
	r.DELETE("/api/fkpl/:id", h.DeleteFkpl)
	r.GET("/api/fkpl/trazikonto", h.TraziKonto)
	r.GET("/api/fkpl/trazikontosearchtable", h.TraziKontoSearchTable)
	r.POST("/api/fkpl/unlock/:id", h.lm.WithEntityLockVerifyAndRelease("fkpl", "id"), h.UnlockFkpl)
}
