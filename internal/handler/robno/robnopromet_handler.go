package robno

import (
	"net/http"
	"strings"

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
	robnoPrometURLPrefix  = "/api/robno-promet"
	robnoPrometGrupe1Tab  = robnoPrometURLPrefix + "/grupe-1"
	robnoPrometGrupe1     = robnoPrometGrupe1Tab + "/table"
	robnoPrometGrupe2Tab  = robnoPrometURLPrefix + "/grupe-2"
	robnoPrometGrupe2     = robnoPrometGrupe2Tab + "/table"
	robnoPrometDobTab     = robnoPrometURLPrefix + "/dobavljaci"
	robnoPrometDob        = robnoPrometDobTab + "/table"
	robnoPrometRucTab     = robnoPrometURLPrefix + "/lager-ruc"
	robnoPrometRuc        = robnoPrometRucTab + "/table"
	robnoPrometGradTab    = robnoPrometURLPrefix + "/gradiliste"
	robnoPrometGrad       = robnoPrometGradTab + "/table"
	robnoPrometGradVpcTab = robnoPrometURLPrefix + "/gradiliste-vpc-nc"
	robnoPrometGradVpc    = robnoPrometGradVpcTab + "/table"
)

type RobnoPrometHandler struct {
	service robnosvc.RobnoPrometService
	cfg     config.Config
	tabs    *domain.TabData
}

func NewRobnoPrometHandler(s robnosvc.RobnoPrometService, cfg config.Config) *RobnoPrometHandler {
	return &RobnoPrometHandler{service: s, cfg: cfg, tabs: robnoPrometTabs()}
}

func (h *RobnoPrometHandler) RobnoPrometMain(c *gin.Context) {
	tbl := h.table("Promet artikala po grupama za period", "robnopromet-grupe", h.service.GetGrupeTableFields(), robnoPrometGrupe1)
	obrada := common.SetButton("robnopromet-grupe-obrada", "Obradi", "obrada", robnoPrometGrupe1, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	stampaj := common.SetButton("robnopromet-grupe-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	if err := tmpl_robno.RobnoPrometMain(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer); err != nil {
		h.error(c, err)
	}
}

func (h *RobnoPrometHandler) grupe1Tab(c *gin.Context) { h.renderGrupe(c, true, false) }
func (h *RobnoPrometHandler) grupe2Tab(c *gin.Context) { h.renderGrupe(c, false, true) }
func (h *RobnoPrometHandler) dobTab(c *gin.Context) {
	h.renderSimple(c, "Nabavka po dobavljačima", "robnopromet-dobavljaci", h.service.GetDobavljaciTableFields(), robnoPrometDob, "dob")
}
func (h *RobnoPrometHandler) rucTab(c *gin.Context) {
	h.renderSimple(c, "Lager lista ulaz/izlaz RUC", "robnopromet-ruc", h.service.GetRucTableFields(), robnoPrometRuc, "ruc")
}
func (h *RobnoPrometHandler) gradTab(c *gin.Context) {
	h.renderSimple(c, "Izveštaj zaduženja gradilišta", "robnopromet-gradiliste", h.service.GetGradilisteTableFields(), robnoPrometGrad, "grad")
}
func (h *RobnoPrometHandler) gradVpcTab(c *gin.Context) {
	h.renderSimple(c, "Izveštaj zaduženja gradilišta VPC-NC", "robnopromet-gradiliste-vpc", h.service.GetGradilisteTableFields(), robnoPrometGradVpc, "gradvpc")
}

func (h *RobnoPrometHandler) grupe1(c *gin.Context) {
	h.renderGrupe(c, true, false)
}
func (h *RobnoPrometHandler) grupe2(c *gin.Context) {
	h.renderGrupe(c, false, true)
}
func (h *RobnoPrometHandler) dobavljaci(c *gin.Context) {
	h.renderSimple(c, "Nabavka po dobavljačima", "robnopromet-dobavljaci", h.service.GetDobavljaciTableFields(), robnoPrometDob, "dob")
}
func (h *RobnoPrometHandler) ruc(c *gin.Context) {
	h.renderSimple(c, "Lager lista ulaz/izlaz RUC", "robnopromet-ruc", h.service.GetRucTableFields(), robnoPrometRuc, "ruc")
}
func (h *RobnoPrometHandler) gradiliste(c *gin.Context) {
	h.renderSimple(c, "Izveštaj zaduženja gradilišta", "robnopromet-gradiliste", h.service.GetGradilisteTableFields(), robnoPrometGrad, "grad")
}
func (h *RobnoPrometHandler) gradilisteVpcNc(c *gin.Context) {
	h.renderSimple(c, "Izveštaj zaduženja gradilišta VPC-NC", "robnopromet-gradiliste-vpc", h.service.GetGradilisteTableFields(), robnoPrometGradVpc, "gradvpc")
}

func (h *RobnoPrometHandler) renderGrupe(c *gin.Context, first, second bool) {
	url, id, method := robnoPrometGrupe1, "robnopromet-grupe", h.service.GetGrupe1
	if second {
		url, id, method = robnoPrometGrupe2, "robnopromet-grupe-2", h.service.GetGrupe2
	}
	tbl := h.table("Promet artikala po grupama za period", id, h.service.GetGrupeTableFields(), url)
	if c.Request.URL.Path == robnoPrometGrupe1Tab || c.Request.URL.Path == robnoPrometGrupe2Tab {
		var renderErr error
		if first {
			obrada := common.SetButton("robnopromet-grupe-obrada", "Obradi", "obrada", url, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
			stampaj := common.SetButton("robnopromet-grupe-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
			renderErr = tmpl_robno.RobnoPrometGrupe1(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer)
		} else {
			obrada := common.SetButton("robnopromet-grupe-2-obrada", "Obradi", "obrada", url, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
			stampaj := common.SetButton("robnopromet-grupe-2-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
			renderErr = tmpl_robno.RobnoPrometGrupe2(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer)
		}
		if renderErr != nil {
			h.error(c, renderErr)
		}
		return
	}
	params := h.params(c)
	if err := method(c, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	_ = tmpl_robno.RobnoPrometGrupe1Table( tbl, i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoPrometHandler) renderSimple(c *gin.Context, title, id string, fields []domain.Fields, url, report string) {
	tbl := h.table(title, id, fields, url)
	if c.Request.URL.Path == strings.TrimSuffix(url, "/table") {
		obrada := common.SetButton(id+"-obrada", "Obradi", "obrada", url, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
		stampaj := common.SetButton(id+"-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
		var renderErr error
		switch report {
		case "dob":
			common.SetActiveTab(h.tabs, 0)
			renderErr = tmpl_robno.RobnoPrometDobavljaci(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer)
		case "ruc":
			common.SetActiveTab(h.tabs, 1)
			renderErr = tmpl_robno.RobnoPrometRuc(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer)
		case "grad":
			common.SetActiveTab(h.tabs, 2)
			renderErr = tmpl_robno.RobnoPrometGradiliste(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer)
		case "gradvpc":
			common.SetActiveTab(h.tabs, 3)
			renderErr = tmpl_robno.RobnoPrometGradilisteVpcNc(*h.tabs, tbl, obrada, stampaj, i18n.GetInstance()).Render(c, c.Writer)
		}
		if renderErr != nil {
			h.error(c, renderErr)
		}
		return
	}
	params := h.params(c)
	params.ReportTip = report
	var err error
	switch report {
	case "dob":
		err = h.service.GetDobavljaci(c, &tbl, true, 0, 0, params)
	case "ruc":
		err = h.service.GetRuc(c, &tbl, true, 0, 0, params)
	case "grad":
		err = h.service.GetGradiliste(c, &tbl, true, 0, 0, params)
	case "gradvpc":
		err = h.service.GetGradilisteVpcNc(c, &tbl, true, 0, 0, params)
	}
	if err != nil {
		h.error(c, err)
		return
	}
	_ = tmpl_robno.RobnoPrometTable(tbl, i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoPrometHandler) table(title, id string, fields []domain.Fields, url string) domain.TableData {
	tbl := common.SetTableBasicData(title, id, fields, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, url, false, false, false)
	tbl.HasTotals = true
	return tbl
}
func (h *RobnoPrometHandler) params(c *gin.Context) domain.PrometParam {
	return domain.PrometParam{OdKonta: c.Query("odmagacina"), DoKonta: c.Query("domagacina"), OdSifre: c.Query("odsifre"), DoSifre: c.Query("dosifre"), OdDatuma: c.Query("oddatuma"), DoDatuma: c.Query("dodatuma"), SearchText: c.Query("query")}
}
func (h *RobnoPrometHandler) error(c *gin.Context, err error) {
	common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
}

func (h *RobnoPrometHandler) AddRoutes(r *gin.Engine) {
	// Apply auth middleware to all Robno Promet routes.
	r.Use(middleware.Auth())

	// Define routes for Robno Promet.
	r.GET("/api/robno-promet", h.RobnoPrometMain)
	r.GET("/api/robno-promet/grupe-1", h.grupe1Tab)
	r.GET("/api/robno-promet/grupe-1/table", h.grupe1)
	r.GET("/api/robno-promet/grupe-2", h.grupe2Tab)
	r.GET("/api/robno-promet/grupe-2/table", h.grupe2)
	r.GET("/api/robno-promet/dobavljaci", h.dobTab)
	r.GET("/api/robno-promet/dobavljaci/table", h.dobavljaci)
	r.GET("/api/robno-promet/lager-ruc", h.rucTab)
	r.GET("/api/robno-promet/lager-ruc/table", h.ruc)
	r.GET("/api/robno-promet/gradiliste", h.gradTab)
	r.GET("/api/robno-promet/gradiliste/table", h.gradiliste)
	r.GET("/api/robno-promet/gradiliste-vpc-nc", h.gradVpcTab)
	r.GET("/api/robno-promet/gradiliste-vpc-nc/table", h.gradilisteVpcNc)
}

func robnoPrometTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnopromet-grupe1", Label: translator.Label("Promet po grupi artikala 1"), HXRequestUrl: robnoPrometGrupe1Tab, IsActive: true, Name: "grupe1"},
		{ID: "robnopromet-grupe2", Label: translator.Label("Promet po grupi artikala 2"), HXRequestUrl: robnoPrometGrupe2Tab, Name: "grupe2"},
		{ID: "robnopromet-dobavljaci", Label: translator.Label("Nabavka po dobavljačima"), HXRequestUrl: robnoPrometDobTab, Name: "dobavljaci"},
		{ID: "robnopromet-ruc", Label: translator.Label("Lager lista ulaz/izlaz RUC"), HXRequestUrl: robnoPrometRucTab, Name: "ruc"},
		{ID: "robnopromet-gradiliste", Label: translator.Label("Izveštaj zaduženja gradilišta"), HXRequestUrl: robnoPrometGradTab, Name: "gradiliste"},
		{ID: "robnopromet-gradiliste-vpc", Label: translator.Label("Izveštaj zaduženja gradilišta VPC-NC"), HXRequestUrl: robnoPrometGradVpcTab, Name: "gradiliste-vpc"},
	}}
}
