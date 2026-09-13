package robno

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"helia/config"
	rep "helia/frontend/templates/reports/robno"
	tmpl "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"
	"net/http"
	"strconv"
)

const (
	fispTitle   = "MESTA ISPORUKE"
	fispTableID = "fisp-table"
	fispURL     = "/api/fisp"
	fispAll     = "/api/fisp/all"
	fispPrint   = "/api/fisp/stampa"
	fispAdd     = "/api/fisp/confirm-add"
	fispSave    = "/api/fisp/save"
	fispDelete  = "/api/fisp/confirm-delete"
	fispUpdate  = "/api/fisp/confirm-update"
)

type FispHandler struct {
	service     service.Service[domain.Fisp]
	fispService robnosvc.FispService
	cfg         config.Config
	lm          *middleware.LockMiddleware
}

func NewFispHandler(base *service.BaseService[domain.Fisp], resource robnosvc.FispService, cfg config.Config, lm *middleware.LockMiddleware) *FispHandler {
	return &FispHandler{service: base, fispService: resource, cfg: cfg, lm: lm}
}

func fispFields(e *domain.Fisp) []domain.Fields {
	return []domain.Fields{{Name: "idpartneri", Value: fmt.Sprint(e.IDPartneri)}, {Name: "konto", Value: e.Konto}, {Name: "sifra", Value: e.Sifra}, {Name: "mi", Value: fmt.Sprint(e.MI)}, {Name: "naziv", Value: e.Naziv}, {Name: "adresa", Value: e.Adresa}, {Name: "pobro", Value: fmt.Sprint(e.Pobro)}, {Name: "mesto", Value: e.Mesto}, {Name: "pib", Value: e.PIB}, {Name: "kontaktosb", Value: e.KontaktOsb}, {Name: "gln", Value: fmt.Sprint(e.GLN)}, {Name: "email", Value: e.Email}, {Name: "ter", Value: fmt.Sprint(e.Ter)}}
}
func (h *FispHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	var e domain.Fisp
	if err := c.ShouldBind(&e); err != nil {
		common.WriteJSONResponse(c, 400, false, nil, err.Error())
		return
	}
	if v := h.fispService.ValidateEntity(ctx, &e); len(v) > 0 {
		common.WriteJSONResponse(c, 400, false, v, common.ErrMsgValidation)
		return
	}
	if _, _, err := h.fispService.Create(ctx, &e, common.IDmi, fispFields(&e)); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, 200, true, nil, common.OkMsgSaveData)
}
func (h *FispHandler) Update(c *gin.Context) {
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, 400, false, nil, common.ErrMsgInvalidID)
		return
	}
	var e domain.Fisp
	if err = c.ShouldBind(&e); err != nil {
		common.WriteJSONResponse(c, 400, false, nil, err.Error())
		return
	}
	v, err := h.fispService.Update(c.Request.Context(), &e, common.IDmi, id, fispFields(&e))
	if err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	if len(v) > 0 {
		common.WriteJSONResponse(c, 400, false, v, common.ErrMsgValidation)
		return
	}
	common.WriteJSONResponse(c, 200, true, nil, common.OkMsgSaveData)
}
func (h *FispHandler) Delete(c *gin.Context) {
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, 400, false, nil, common.ErrMsgInvalidID)
		return
	}
	if err = h.fispService.Delete(c.Request.Context(), common.IDmi, id); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, fmt.Sprintf(common.ErrMsgDeleteData, err))
		return
	}
	common.WriteJSONResponse(c, 200, true, nil, common.OkMsgDeleteData)
}
func (h *FispHandler) confirmDelete(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.fispService.GetFispTableFields(), "#info-message")
}
func (h *FispHandler) confirmAdd(c *gin.Context) {
	d := domain.Dialog{Title: "Novo mesto isporuke", Id: "fisp-add-form", HxActionURL: fispSave, HxRequestType: "POST", OkText: "Sačuvaj", CancelText: "Odustani"}
	save := common.SetButton("btn-save-fisp", "Sačuvaj", "save", fispSave, "", "", "POST", "#info-message", "", true, common.ClassSaveButton, "")
	close := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	close.IdDialog = d.Id
	cancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	cancel.IdDialog = d.Id
	tmpl.FispDialog(d, common.ActionAdd, domain.Fisp{}, save, cancel, close, i18n.GetInstance(), common.GetCsrfToken(c)).Render(c.Request.Context(), c.Writer)
}
func (h *FispHandler) confirmUpdate(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		common.WriteJSONResponse(c, 400, false, nil, common.ErrMsgInvalidID)
		return
	}
	e, err := h.fispService.GetByID(c.Request.Context(), common.IDmi, int64(id))
	if err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	d := domain.Dialog{Title: "Izmena mesta isporuke", Id: "fisp-edit-form", HxActionURL: fmt.Sprintf("%s/%d", fispURL, id), HxRequestType: "PUT"}
	save := common.SetButton("btn-save-fisp", "Sačuvaj", "save", d.HxActionURL, "", "", "PUT", "#info-message", "", true, common.ClassSaveButton, "")
	close := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	close.IdDialog = d.Id
	cancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	cancel.IdDialog = d.Id
	common.SetUnlockButtonProperties(&close, fmt.Sprintf("%s/unlock/%d", fispURL, id))
	common.SetUnlockButtonProperties(&cancel, fmt.Sprintf("%s/unlock/%d", fispURL, id))
	tmpl.FispDialog(d, common.ActionUpdate, *e, save, cancel, close, i18n.GetInstance(), common.GetCsrfToken(c)).Render(c.Request.Context(), c.Writer)
}
func (h *FispHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromContext(c) == nil {
		common.WriteJSONResponse(c, 401, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	page, size := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl := common.SetTableBasicData(fispTitle, fispTableID, h.fispService.GetFispTableFields(), "", fispAll, size, page, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, fispTitle, fispAll, true, true, false)
	tbl.ShowPagination = true
	tbl.ShowActions = true
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = fispUpdate
	tbl.BtnDelete.HxActionURL = fispDelete
	tbl.BtnAdd.HxActionURL = fispAdd
	tbl.BtnAdd.IsVisible = true
	tbl.URLGetAll = fispAll
	tbl.URLPrefix = fispURL
	sort, order, search := c.Query("sortBy"), c.Query("sortOrder"), c.Query("query")
	if err := h.fispService.GetAllFisp(ctx, &tbl, page, size, true, sort, order, search, common.TipStampePreview); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	if err := h.fispService.GetAllFisp(ctx, &tbl, page, size, false, sort, order, search, common.TipStampePreview); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	if c.GetHeader("X-Request-Source") == "menu" || c.GetHeader("X-Request-Source") == "" {
		tmpl.Fisp(tbl, common.CreateSearchInput("search-input", i18n.GetInstance(), fispAll, "#"+fispTableID, ""), common.SetPrintButton("btn-print-fisp", "Štampa", "fin_print", fispPrint, "GET", true, common.ClassPrintButton, ""), i18n.GetInstance()).Render(ctx, c.Writer)
	} else {
		utils.RenderContent(c, tbl)
	}
}
func (h *FispHandler) Print(c *gin.Context) {
	ctx := c.Request.Context()
	f, err := h.fispService.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	page, size := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl := common.SetTableBasicData(fispTitle, fispTableID, h.fispService.GetFispTableFields(), "", fispAll, 0, 0, 0, 0, h.cfg)
	if err = h.fispService.GetAllFisp(ctx, &tbl, page, size, true, c.Query("sortBy"), c.Query("sortOrder"), c.Query("query"), common.TipStampePrint); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	rep.FispStampa(domain.ReportParameters{Orientation: "landscape", CompanyName: f.Naziv, Adress: f.Adresa, Postcode: f.Pobro, City: f.Mesto, PIB: f.PIB, MatBroj: f.Matbr, ReportName: "Mesta isporuke", ParameterItems: map[string]domain.ParameterItem{}}, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}
func (h *FispHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())
	r.POST("/api/fisp/", h.Create)
	r.POST("/api/fisp/save", h.Create)
	r.PUT("/api/fisp/:id", h.Update)
	r.DELETE("/api/fisp/:id", h.Delete)
	r.GET("/api/fisp/all", h.GetAll)
	r.GET("/api/fisp/print", h.Print)
	r.GET("/api/fisp/add", h.confirmAdd)
	r.GET("/api/fisp/update", h.confirmUpdate)
	r.GET("/api/fisp/delete", h.confirmDelete)
	r.GET("/api/fisp/unlock/:id", func(c *gin.Context) { c.JSON(200, gin.H{"success": true}) })
}
