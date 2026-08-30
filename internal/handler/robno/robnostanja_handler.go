package robno

import (
	"net/http"

	"helia/config"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	robnosvc "helia/internal/service/robno"

	"github.com/gin-gonic/gin"
)

const (
	robnoStanjaURLPrefix    = "/api/robno-stanja"
	robnoStanjaArticleTab   = robnoStanjaURLPrefix + "/artikl"
	robnoStanjaArticle      = robnoStanjaArticleTab + "/table"
	robnoStanjaMultipleTab  = robnoStanjaURLPrefix + "/vise-artikala"
	robnoStanjaMultiple     = robnoStanjaMultipleTab + "/table"
	robnoStanjaSubTab       = robnoStanjaURLPrefix + "/saldo-subsintetickog-konta"
	robnoStanjaSub          = robnoStanjaSubTab + "/table"
	robnoStanjaReconcileTab = robnoStanjaURLPrefix + "/svodjenje-zalihe"
	robnoStanjaReconcile    = robnoStanjaReconcileTab + "/table"
)

type RobnoStanjaHandler struct {
	service robnosvc.RobnoStanjaService
	cfg     config.Config
	tabs    *domain.TabData
}

func NewRobnoStanjaHandler(s robnosvc.RobnoStanjaService, cfg config.Config) *RobnoStanjaHandler {
	return &RobnoStanjaHandler{service: s, cfg: cfg, tabs: robnoStanjaTabs()}
}

func (h *RobnoStanjaHandler) RobnoStanjaMain(c *gin.Context) {
	common.SetActiveTab(h.tabs, 0)
	tbl := h.table("Prikaz stanja artikla", "robnostanja-artikl", h.service.GetPojedinacnogArtiklaTableFields(), robnoStanjaArticle)
	obrada := common.SetButton("robnostanja-artikl-obrada", "Obradi", "obrada", robnoStanjaArticle, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	stampaj := common.SetButton("robnostanja-artikl-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	if err := tmpl_robno.RobnoStanjaMain(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}
func (h *RobnoStanjaHandler) PrikazStanjaPojedincnogArtikla(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		common.SetActiveTab(h.tabs, 0)
		tbl := h.table("Prikaz stanja pojedinačnog artikla", "robnostanja-artikl", h.service.GetPojedinacnogArtiklaTableFields(), robnoStanjaArticle)
		button := common.SetButton("robnostanja-artikal-obrada", "Obradi", "obrada", robnoStanjaArticle, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
		printButton := common.SetButton("robnostanja-artikl-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
		_ = tmpl_robno.RobnoStanjePojedinacnogArtikla(*h.tabs, tbl, button, printButton, translator).Render(c, c.Writer)
		return
	}
	if requestSource != "menu" && requestSource != "tab" {
		//params := domain.RobnoStanjeParams{}
		//h.service.GetStanjePojedinacnogArtikla(ctx, tbl, true, 0, 0, params)
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgGetData)
		return
	}

}
func (h *RobnoStanjaHandler) PrikazStanjaViseArtikala(c *gin.Context) {

}
func (h *RobnoStanjaHandler) PrikazSaldaSubsintetickogKonta(c *gin.Context) {

}
func (h *RobnoStanjaHandler) SvodjenjeStanjaZaliha(c *gin.Context) {

}

func (h *RobnoStanjaHandler) article(c *gin.Context)   { h.renderArticle(c, robnoStanjaArticle) }
func (h *RobnoStanjaHandler) multiple(c *gin.Context)  { h.renderMultiple(c, robnoStanjaMultiple) }
func (h *RobnoStanjaHandler) sub(c *gin.Context)       { h.renderSub(c, robnoStanjaSub) }
func (h *RobnoStanjaHandler) reconcile(c *gin.Context) { h.renderReconcile(c, robnoStanjaReconcile) }

func (h *RobnoStanjaHandler) articleTab(c *gin.Context)   { h.renderArticleTab(c) }
func (h *RobnoStanjaHandler) multipleTab(c *gin.Context)  { h.renderMultipleTab(c) }
func (h *RobnoStanjaHandler) subTab(c *gin.Context)       { h.renderSubTab(c) }
func (h *RobnoStanjaHandler) reconcileTab(c *gin.Context) { h.renderReconcileTab(c) }
func (h *RobnoStanjaHandler) multipleSubtab(c *gin.Context) {
	_ = tmpl_robno.RobnoStanjeMultipleSubtab(i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoStanjaHandler) renderArticleTab(c *gin.Context) {
}

func (h *RobnoStanjaHandler) renderMultipleTab(c *gin.Context) {
	tbl := h.table("Prikaz stanja više artikala", "robnostanja-vise-artikala", h.service.GetViseArtikalaTableFields(), robnoStanjaMultiple)
	button := common.SetButton("robnostanja-vise-obrada", "Obradi", "obrada", robnoStanjaMultiple, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	printButton := common.SetButton("robnostanja-vise-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	_ = tmpl_robno.RobnoStanjeViseArtikala(*h.tabs, tbl, button, printButton, i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoStanjaHandler) renderSubTab(c *gin.Context) {
	common.SetActiveTab(h.tabs, 2)
	tbl := h.table("Saldo subsintetičkog konta", "robnostanja-sub", h.service.GetSubsintetiskogKontaTableFields(), robnoStanjaSub)
	button := common.SetButton("robnostanja-sub-obrada", "Obradi", "obrada", robnoStanjaSub, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	printButton := common.SetButton("robnostanja-sub-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	_ = tmpl_robno.RobnoStanjeSubsintetickogKonta(*h.tabs, tbl, button, printButton, i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoStanjaHandler) renderReconcileTab(c *gin.Context) {
	common.SetActiveTab(h.tabs, 3)
	tbl := h.table("Svođenje stanja zalihe", "robnostanja-svodjenje", h.service.GetSvodjenjeZalihaTableFields(), robnoStanjaReconcile)
	button := common.SetButton("robnostanja-svodjenje-obrada", "Obradi", "obrada", robnoStanjaReconcile, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	printButton := common.SetButton("robnostanja-svodjenje-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	_ = tmpl_robno.RobnoSvodjenjeZaliha(*h.tabs, tbl, button, printButton, i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoStanjaHandler) renderArticle(c *gin.Context, url string) {
	tbl := h.table("Prikaz stanja artikla", "robnostanja-artikl", h.service.GetPojedinacnogArtiklaTableFields(), url)
	params := domain.PrometParam{Sifra: c.Query("sifra"), Konto: c.Query("konto"), ReportTip: "prometankonta"}
	if err := h.service.GetStanjePojedinacnogArtikla(c, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	_ = tmpl_robno.RobnoStanjePojedinacnogArtikla(*h.tabs, tbl, common.SetButton("robnostanja-artikl-obrada", "Obradi", "obrada", url, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, ""), common.SetButton("robnostanja-artikl-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, ""), i18n.GetInstance()).Render(c, c.Writer)
}
func (h *RobnoStanjaHandler) renderMultiple(c *gin.Context, url string) {
	tbl := h.table("Prikaz stanja više artikala", "robnostanja-vise-artikala", h.service.GetViseArtikalaTableFields(), url)
	params := domain.PrometParam{OdKonta: c.Query("odkonta"), DoKonta: c.Query("dokonta"), OdSifre: c.Query("odsifre"), DoSifre: c.Query("dosifre"), ReportTip: "karticasintetickihkonta"}
	if err := h.service.GetStanjaViseArtikala(c, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	_ = tmpl_robno.RobnoStanjeViseArtikala(*h.tabs, tbl, common.SetButton("robnostanja-vise-artikala-obrada", "Obradi", "obrada", url, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, ""), common.SetButton("robnostanja-vise-artikala-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, ""), i18n.GetInstance()).Render(c, c.Writer)
}
func (h *RobnoStanjaHandler) renderSub(c *gin.Context, url string) {
	tbl := h.table("Saldo subsintetičkog konta", "robnostanja-sub", h.service.GetSubsintetiskogKontaTableFields(), url)
	params := domain.PrometParam{Konto: c.Query("konto"), ReportTip: "subsintetickakonta"}
	if err := h.service.GetStanjaSubsintetickogKonta(c, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	_ = tmpl_robno.RobnoStanjeSubsintetickogKonta(*h.tabs, tbl, common.SetButton("robnostanja-sub-obrada", "Obradi", "obrada", url, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, ""), common.SetButton("robnostanja-sub-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, ""), i18n.GetInstance()).Render(c, c.Writer)
}
func (h *RobnoStanjaHandler) renderReconcile(c *gin.Context, url string) {
	tbl := h.table("Svođenje stanja zalihe", "robnostanja-svodjenje", h.service.GetSvodjenjeZalihaTableFields(), url)
	params := domain.PrometParam{OdSifre: c.Query("odsifre"), DoSifre: c.Query("dosifre"), ReportTip: "kontaanaliticki"}
	if err := h.service.GetSvodjenjeZaliha(c, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	_ = tmpl_robno.RobnoSvodjenjeZaliha(*h.tabs, tbl, common.SetButton("robnostanja-svodjenje-obrada", "Obradi", "obrada", url, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, ""), common.SetButton("robnostanja-svodjenje-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, ""), i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoStanjaHandler) table(title, id string, fields []domain.Fields, url string) domain.TableData {
	tbl := common.SetTableBasicData(title, id, fields, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, url, false, false, false)
	tbl.HasTotals = true
	return tbl
}
func (h *RobnoStanjaHandler) error(c *gin.Context, err error) {
	common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
}
func (h *RobnoStanjaHandler) AddRoutes(r *gin.Engine) {
	// Apply auth middleware to all Robno Stanja routes.
	r.Use(middleware.Auth())

	// Define routes for Robno Stanja.
	r.GET("/api/robno-stanja", h.RobnoStanjaMain)
	r.GET("/api/robno-stanja/artikl", h.articleTab)
	r.GET("/api/robno-stanja/artikl/table", h.article)
	r.GET("/api/robno-stanja/vise-artikala", h.multipleTab)
	r.GET("/api/robno-stanja/vise-artikala/po-sifri", h.multipleSubtab)
	r.GET("/api/robno-stanja/vise-artikala/po-grupi", h.multipleSubtab)
	r.GET("/api/robno-stanja/vise-artikala/table", h.multiple)
	r.GET("/api/robno-stanja/saldo-subsintetickog-konta", h.subTab)
	r.GET("/api/robno-stanja/saldo-subsintetickog-konta/table", h.sub)
	r.GET("/api/robno-stanja/svodjenje-zalihe", h.reconcileTab)
	r.GET("/api/robno-stanja/svodjenje-zalihe/table", h.reconcile)
}
func robnoStanjaTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnostanja-artikl", Label: translator.T("Prikaz stanja pojedinačnog artikla"), HXRequestUrl: robnoStanjaArticleTab, IsActive: true, Name: "artikl"},
		{ID: "robnostanja-vise", Label: translator.T("Prikaz stanja više artikala"), HXRequestUrl: robnoStanjaMultipleTab, Name: "vise-artikala"},
		{ID: "robnostanja-sub", Label: translator.T("Prikaz salda subsintetičkog konta"), HXRequestUrl: robnoStanjaSubTab, Name: "saldo-subsintetickog-konta"},
		{ID: "robnostanja-svodjenje", Label: translator.T("Svodjenje stanja zalihe"), HXRequestUrl: robnoStanjaReconcileTab, Name: "svodjenje-zalihe"},
	}}
}
