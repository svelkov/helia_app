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
	robnoKarticaURLPrefix  = "/api/robno-kartica"
	robnoKarticaArticleTab = robnoKarticaURLPrefix + "/artikla"
	robnoKarticaArticle    = robnoKarticaArticleTab + "/table"
	robnoKarticaSubTab     = robnoKarticaURLPrefix + "/subsintetickog-konta"
	robnoKarticaSub        = robnoKarticaSubTab + "/table"
)

type RobnoKarticaHandler struct {
	service robnosvc.RobnoKarticaService
	cfg     config.Config
	tabs    *domain.TabData
}

func NewRobnoKarticaHandler(service robnosvc.RobnoKarticaService, cfg config.Config) *RobnoKarticaHandler {
	return &RobnoKarticaHandler{service: service, cfg: cfg, tabs: robnoKarticaTabs()}
}

func (h *RobnoKarticaHandler) RobnoKarticaMain(c *gin.Context) {
	articleTable := h.newTable("Prikaz kartice artikla", "robnokartica-artikli", h.service.GetKarticaArtiklaTableFields(), robnoKarticaArticle)
	subTable := h.newTable("Prikaz kartice subsintetičkog konta", "robnokartica-subsintetika", h.service.GetKarticaSubsintetickogKontaTableFields(), robnoKarticaSub)
	articleButton := common.SetButton("robno-kartica-article-process", "Obradi", "obrada", robnoKarticaArticle, "#robnokartica-artikli", "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	articlePrintButton := common.SetButton("robno-kartica-article-print", "Štampa", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	subButton := common.SetButton("robno-kartica-sub-process", "Obradi", "obrada", robnoKarticaSub, "#robnokartica-subsintetika", "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	subPrintButton := common.SetButton("robno-kartica-sub-print", "Štampa", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	magacini := []domain.ComboItem{}
	if err := tmpl_robno.RobnoKarticaMain(*h.tabs, articleTable, subTable, articleButton, articlePrintButton, subButton, subPrintButton, magacini, i18n.GetInstance()).Render(c.Request.Context(), c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoKarticaHandler) GetPrikazKarticeArtikla(c *gin.Context) {
	tbl := h.newTable("Prikaz kartice artikla", "robnokartica-artikli", h.service.GetKarticaArtiklaTableFields(), robnoKarticaArticle)
	params := domain.PrometParam{Sifra: c.Query("sifra"), OdDatuma: c.Query("oddatuma"), DoDatuma: c.Query("dodatuma"), SearchText: c.Query("query"), ReportTip: "karticaankonta"}
	if err := h.service.GetKarticaArtikla(c.Request.Context(), &tbl, true, 0, 0, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	_ = tmpl_robno.RobnoKarticaArticleTable(tbl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *RobnoKarticaHandler) GetPrikazKarticeArtiklaTab(c *gin.Context) {
	common.SetActiveTab(h.tabs, 0)
	tbl := h.newTable("Prikaz kartice artikla", "robnokartica-artikli", h.service.GetKarticaArtiklaTableFields(), robnoKarticaArticle)
	button := common.SetButton("robno-kartica-article-process", "Obradi", "obrada", robnoKarticaArticle, "#robnokartica-artikli", "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	printButton := common.SetButton("robno-kartica-article-print", "Štampa", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	magacini := []domain.ComboItem{}
	_ = tmpl_robno.RobnoKarticaArticle(*h.tabs, tbl, button, printButton, magacini, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *RobnoKarticaHandler) GetKarticaSubsintetickogKonta(c *gin.Context) {
	common.SetActiveTab(h.tabs, 1)
	tbl := h.newTable("Prikaz kartice subsintetičkog konta", "robnokartica-subsintetika", h.service.GetKarticaSubsintetickogKontaTableFields(), robnoKarticaSub)
	params := domain.PrometParam{OdKonta: c.Query("odkonta"), DoKonta: c.Query("dokonta"), OdDatuma: c.Query("oddatuma"), DoDatuma: c.Query("dodatuma"), SearchText: c.Query("query"), ReportTip: "subsintetickakonta"}
	if err := h.service.GetKarticaSubsintetickogKonta(c.Request.Context(), &tbl, true, 0, 0, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	_ = tmpl_robno.RobnoKarticaSubsyntheticTable(tbl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *RobnoKarticaHandler) GetKarticaSubsintetickogKontaTab(c *gin.Context) {
	common.SetActiveTab(h.tabs, 1)
	tbl := h.newTable("Prikaz kartice subsintetičkog konta", "robnokartica-subsintetika", h.service.GetKarticaSubsintetickogKontaTableFields(), robnoKarticaSub)
	button := common.SetButton("robno-kartica-sub-process", "Obradi", "obrada", robnoKarticaSub, "#robnokartica-subsintetika", "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	printButton := common.SetButton("robno-kartica-sub-print", "Štampa", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	magacini := []domain.ComboItem{}
	_ = tmpl_robno.RobnoKarticaSubsynthetic(*h.tabs, tbl, button, printButton, magacini, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *RobnoKarticaHandler) newTable(title, id string, fields []domain.Fields, url string) domain.TableData {
	tbl := common.SetTableBasicData(title, id, fields, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, url, false, false, false)
	tbl.HasTotals = true
	return tbl
}

func (h *RobnoKarticaHandler) AddRoutes(r *gin.Engine) {
	// Apply auth middleware to all Robno Promet routes.
	r.Use(middleware.Auth())

	// Define routes for Robno Promet.
	r.GET("/api/robno-kartica", h.RobnoKarticaMain)
	r.GET("/api/robno-kartica/artikla", h.GetPrikazKarticeArtiklaTab)
	r.GET("/api/robno-kartica/artikla/table", h.GetPrikazKarticeArtikla)
	r.GET("/api/robno-kartica/subsintetickog-konta", h.GetKarticaSubsintetickogKontaTab)
	r.GET("/api/robno-kartica/subsintetickog-konta/table", h.GetKarticaSubsintetickogKonta)
}

func robnoKarticaTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnokartica-artikli", Label: translator.Label("Prikaz kartice artikla"), HXRequestUrl: robnoKarticaArticleTab, IsActive: true, Name: "artikli"},
		{ID: "robnokartica-subsintetika", Label: translator.Label("Prikaz kartice subsintetičkog konta"), HXRequestUrl: robnoKarticaSubTab, IsActive: false, Name: "subsintetika"},
	}}
}
