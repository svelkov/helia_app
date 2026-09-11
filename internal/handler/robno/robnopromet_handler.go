package robno

import (
	"context"
	"net/http"

	"helia/config"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	robnoPrometURLPrefix = "/api/robno-promet"

	// Grupe 1
	robnoPrometGrupe1Title   = "Promet artikala po grupama za period"
	robnoPrometGrupe1TableID = "robnopromet-grupe"
	robnoPrometGrupe1URL     = robnoPrometURLPrefix + "/grupe-1"

	// Kupci
	robnoPrometKupciTitle   = "Promet artikala po kupcima za period"
	robnoPrometKupciTableID = "robnopromet-kupci"
	robnoPrometKupciURL     = robnoPrometURLPrefix + "/kupci"

	// Dobavljači
	robnoPrometDobavljaciTitle   = "Nabavka po dobavljačima"
	robnoPrometDobavljaciTableID = "robnopromet-dobavljaci"
	robnoPrometDobavljaciURL     = robnoPrometURLPrefix + "/dobavljaci"

	// RUC - Lager lista (default sub-tab of the RUC area)
	robnoPrometRucTitle   = "Lager lista ulaz/izlaz RUC"
	robnoPrometRucTableID = "robnopromet-ruc"
	robnoPrometRucURL     = robnoPrometURLPrefix + "/lager-ruc"

	// RUC remaining sub-tabs
	robnoPrometRucUlazIzlazTitle        = "Ulaz/izlaz za period"
	robnoPrometRucUlazIzlazTableID      = "robnopromet-ruc-ulaz-izlaz"
	robnoPrometRucUlazIzlazURL          = robnoPrometURLPrefix + "/ruc/ulaz-izlaz-period"
	robnoPrometRucMagacinimaTitle       = "RUC po magacinima"
	robnoPrometRucMagacinimaTableID     = "robnopromet-ruc-magacinima"
	robnoPrometRucMagacinimaURL         = robnoPrometURLPrefix + "/ruc/ruc-po-magacinima"
	robnoPrometRucIzlazneFaktureTitle   = "Izlazne fakture"
	robnoPrometRucIzlazneFaktureTableID = "robnopromet-ruc-izlazne-fakture"
	robnoPrometRucIzlazneFaktureURL     = robnoPrometURLPrefix + "/ruc/izlazne-fakture"

	// Gradilište
	robnoPrometGradilisteTitle   = "Izveštaj zaduženja gradilišta"
	robnoPrometGradilisteTableID = "robnopromet-gradiliste"
	robnoPrometGradilisteURL     = robnoPrometURLPrefix + "/gradiliste"

	// Gradilište VPC-NC
	robnoPrometGradVpcTitle   = "Izveštaj zaduženja gradilišta VPC-NC"
	robnoPrometGradVpcTableID = "robnopromet-gradiliste-vpc"
	robnoPrometGradVpcURL     = robnoPrometURLPrefix + "/gradiliste-vpc-nc"
)

type RobnoPrometHandler struct {
	service robnosvc.RobnoPrometService
	cfg     config.Config
	tabs    *domain.TabData
	subtabs *domain.TabData
}

func NewRobnoPrometHandler(s robnosvc.RobnoPrometService, cfg config.Config) *RobnoPrometHandler {
	return &RobnoPrometHandler{service: s, cfg: cfg, tabs: robnoPrometTabs(), subtabs: robnoPrometRucSubTabs()}
}

// RobnoPrometMain renders the robno promet main page (first tab - grupe 1).
func (h *RobnoPrometHandler) RobnoPrometMain(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 0)
	tbl := h.reportTable(robnoPrometGrupe1Title, robnoPrometGrupe1TableID, h.service.GetPrometArtiklaTableFields(), robnoPrometGrupe1URL)
	obrada, stampaj := h.reportButtons(robnoPrometGrupe1TableID, robnoPrometGrupe1TableID, robnoPrometGrupe1URL)
	magValues, grupeValues, err := h.comboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	if err := tmpl_robno.RobnoPrometMain(*h.tabs, tbl, magValues, grupeValues, obrada, stampaj, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// RobnoPrometGrupe1 renders the "grupe-1" report form when the tab is opened
// (X-Request-Source: tab/menu) or returns the report table for data requests.
func (h *RobnoPrometHandler) RobnoPrometArtikal(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 0)
	tbl := h.reportTable(robnoPrometGrupe1Title, robnoPrometGrupe1TableID, h.service.GetPrometArtiklaTableFields(), robnoPrometGrupe1URL)
	if !h.isDataRequest(c) {
		obrada, stampaj := h.reportButtons(robnoPrometGrupe1TableID, robnoPrometGrupe1TableID, robnoPrometGrupe1URL)
		magValues, grupeValues, err := h.comboValues(ctx)
		if err != nil {
			h.error(c, err)
			return
		}
		if err := tmpl_robno.RobnoPrometPoGrupiArtikala1(*h.tabs, tbl, magValues, grupeValues, obrada, stampaj, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
			h.error(c, err)
		}
		return
	}
	params := h.params(c)
	if err := h.service.GetPrometArtikala(ctx, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	utils.RenderContent(c, tbl)
}

// RobnoPrometKupci renders the "kupci" report form when the tab is opened or
// returns the report table for data requests.
func (h *RobnoPrometHandler) RobnoPrometKupci(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 1)
	tbl := h.reportTable(robnoPrometKupciTitle, robnoPrometKupciTableID, h.service.GetPrometKupcaTableFields(), robnoPrometKupciURL)
	if !h.isDataRequest(c) {
		obrada, stampaj := h.reportButtons(robnoPrometKupciTableID, robnoPrometKupciTableID, robnoPrometKupciURL)
		magValues, grupeValues, err := h.comboValues(ctx)
		if err != nil {
			h.error(c, err)
			return
		}
		tmpl_robno.RobnoPrometPoKupcima(*h.tabs, tbl, magValues, grupeValues, obrada, stampaj, i18n.GetInstance()).Render(ctx, c.Writer)
		return
	}
	// TODO: kupci ("Promet artikala po kupcima") data is not wired yet - add a
	// service method (e.g. GetKupci) and fetch/populate the table here.
	return
}

// RobnoPrometDobavljaci renders the "dobavljaci" report form when the tab is
// opened or returns the report table for data requests.
func (h *RobnoPrometHandler) RobnoPrometDobavljaci(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 2)
	tbl := h.reportTable(robnoPrometDobavljaciTitle, robnoPrometDobavljaciTableID, h.service.GetPrometOdDobavljacaTableFields(), robnoPrometDobavljaciURL)
	if !h.isDataRequest(c) {
		obrada, stampaj := h.reportButtons(robnoPrometDobavljaciTableID, robnoPrometDobavljaciTableID, robnoPrometDobavljaciURL)
		magValues, grupeValues, err := h.comboValues(ctx)
		if err != nil {
			h.error(c, err)
			return
		}
		if err := tmpl_robno.RobnoPrometPoDobavljacima(*h.tabs, tbl, magValues, grupeValues, obrada, stampaj, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
			h.error(c, err)
		}
		return
	}
	params := h.params(c)
	if err := h.service.GetNabavkeOdDobavljaca(ctx, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	utils.RenderContent(c, tbl)
}

// RobnoPrometRuc renders the RUC "Lager lista" sub-report form (the default
// sub-tab) when the tab is opened or returns the report table for data
// requests.
func (h *RobnoPrometHandler) RobnoPrometRucMain(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 3)
	common.SetActiveTab(h.subtabs, 0)
	tbl := h.reportTable(robnoPrometRucTitle, robnoPrometRucTableID, h.service.GetPrometRucTableFields(), robnoPrometRucURL)
	if !h.isDataRequest(c) {
		obrada, stampaj := h.reportButtons(robnoPrometRucTableID, robnoPrometRucTableID, robnoPrometRucURL)
		magValues, grupeValues, err := h.comboValues(ctx)
		if err != nil {
			h.error(c, err)
			return
		}
		if err := tmpl_robno.RobnoPrometRucLagerLista(*h.tabs, *h.subtabs, tbl, magValues, grupeValues, obrada, stampaj, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
			h.error(c, err)
		}
		return
	}
	params := h.params(c)
	if err := h.service.GetPrometRucLagerLista(ctx, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	utils.RenderContent(c, tbl)
}

// RobnoPrometRucUlazIzlaz renders the "Ulaz/izlaz za period" RUC sub-report.
func (h *RobnoPrometHandler) RobnoPrometRucUlazIzlaz(c *gin.Context) {
	h.renderRucSubReport(c, 1)
}

// RobnoPrometRucMagacinima renders the "RUC po magacinima" RUC sub-report.
func (h *RobnoPrometHandler) RobnoPrometRucMagacinima(c *gin.Context) {
	h.renderRucSubReport(c, 2)
}

// RobnoPrometRucIzlazneFakture renders the "Izlazne fakture" RUC sub-report.
func (h *RobnoPrometHandler) RobnoPrometRucIzlazneFakture(c *gin.Context) {
	h.renderRucSubReport(c, 3)
}

// renderRucSubReport renders the active RUC sub-report form (sub-tabs 1..3).
// The report tables/columns and data (service) are wired later when their
// specs are provided.
func (h *RobnoPrometHandler) renderRucSubReport(c *gin.Context, subIndex int) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 3)
	common.SetActiveTab(h.subtabs, subIndex)
	var title, tableID, url string
	switch subIndex {
	case 1:
		title, tableID, url = robnoPrometRucUlazIzlazTitle, robnoPrometRucUlazIzlazTableID, robnoPrometRucUlazIzlazURL
	case 2:
		title, tableID, url = robnoPrometRucMagacinimaTitle, robnoPrometRucMagacinimaTableID, robnoPrometRucMagacinimaURL
	default:
		title, tableID, url = robnoPrometRucIzlazneFaktureTitle, robnoPrometRucIzlazneFaktureTableID, robnoPrometRucIzlazneFaktureURL
	}
	tbl := h.reportTable(title, tableID, []domain.Fields{}, url)
	if !h.isDataRequest(c) {
		obrada, stampaj := h.reportButtons(tableID, tableID, url)
		magValues, err := h.service.GetMagacinComboValues(ctx)
		if err != nil {
			h.error(c, err)
			return
		}
		translator := i18n.GetInstance()
		var renderErr error
		switch subIndex {
		case 1:
			renderErr = tmpl_robno.RobnoPrometRucUlazIzlaz(*h.tabs, *h.subtabs, tbl, magValues, obrada, stampaj, translator).Render(ctx, c.Writer)
		case 2:
			renderErr = tmpl_robno.RobnoPrometRucMagacinima(*h.tabs, *h.subtabs, tbl, magValues, obrada, stampaj, translator).Render(ctx, c.Writer)
		default:
			renderErr = tmpl_robno.RobnoPrometRucIzlazneFakture(*h.tabs, *h.subtabs, tbl, magValues, obrada, stampaj, translator).Render(ctx, c.Writer)
		}
		if renderErr != nil {
			h.error(c, renderErr)
		}
		return
	}
	// TODO: implement the data request (service) for this sub-report.
}

// RobnoPrometGradiliste renders the "gradiliste" report form when the tab is
// opened or returns the report table for data requests.
func (h *RobnoPrometHandler) RobnoPrometGradiliste(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 4)
	tbl := h.reportTable(robnoPrometGradilisteTitle, robnoPrometGradilisteTableID, h.service.GetPrometGradilistaTableFields(), robnoPrometGradilisteURL)
	if !h.isDataRequest(c) {
		obrada, stampaj := h.reportButtons(robnoPrometGradilisteTableID, robnoPrometGradilisteTableID, robnoPrometGradilisteURL)
		magValues, grupeValues, err := h.comboValues(ctx)
		if err != nil {
			h.error(c, err)
			return
		}
		if err := tmpl_robno.RobnoPrometGradiliste(*h.tabs, tbl, magValues, grupeValues, obrada, stampaj, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
			h.error(c, err)
		}
		return
	}
	params := h.params(c)
	if err := h.service.GetPrometGradilista(ctx, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	utils.RenderContent(c, tbl)
}

// RobnoPrometGradilisteVpcNc renders the "gradiliste-vpc-nc" report form when
// the tab is opened or returns the report table for data requests.
func (h *RobnoPrometHandler) RobnoPrometGradilisteVpcNc(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 5)
	tbl := h.reportTable(robnoPrometGradVpcTitle, robnoPrometGradVpcTableID, h.service.GetPrometGradilisteVpcNcTableFields(), robnoPrometGradVpcURL)
	if !h.isDataRequest(c) {
		obrada, stampaj := h.reportButtons(robnoPrometGradVpcTableID, robnoPrometGradVpcTableID, robnoPrometGradVpcURL)
		magValues, grupeValues, err := h.comboValues(ctx)
		if err != nil {
			h.error(c, err)
			return
		}
		if err := tmpl_robno.RobnoPrometGradilisteVpcNc(*h.tabs, tbl, magValues, grupeValues, obrada, stampaj, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
			h.error(c, err)
		}
		return
	}
	params := h.params(c)
	if err := h.service.GetPrometGradilisteVpcNc(ctx, &tbl, true, 0, 0, params); err != nil {
		h.error(c, err)
		return
	}
	utils.RenderContent(c, tbl)
}

// isDataRequest reports whether the request is a report data request
// (Obradi, sorting, paging, search), as opposed to opening the report tab.
func (h *RobnoPrometHandler) isDataRequest(c *gin.Context) bool {
	switch c.Request.Header.Get("X-Request-Source") {
	case "", "menu", "tab":
		return false
	default:
		return true
	}
}

// reportTable builds an empty table for a report tab.
func (h *RobnoPrometHandler) reportTable(title, tableID string, fields []domain.Fields, url string) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, url, false, false, false)
	tbl.HasTotals = true
	return tbl
}

// reportButtons returns the "Obradi" and "Štampaj" buttons of a report.
func (h *RobnoPrometHandler) reportButtons(idPrefix, tableID, url string) (domain.Button, domain.Button) {
	obrada := common.SetButton(idPrefix+"-obrada", "Obradi", "obrada", url, "#"+tableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	stampaj := common.SetButton(idPrefix+"-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	return obrada, stampaj
}

// comboValues returns the lists of magacini and robne grupe used to populate
// the "od/do magacina" and "od/do grupe" comboboxes of the report forms.
func (h *RobnoPrometHandler) comboValues(ctx context.Context) ([]domain.ComboItem, []domain.ComboItem, error) {
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		return nil, nil, err
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	if err != nil {
		return nil, nil, err
	}
	return magValues, grupeValues, nil
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
	r.GET("/api/robno-promet/grupe-1", h.RobnoPrometArtikal)
	r.GET("/api/robno-promet/kupci", h.RobnoPrometKupci)
	r.GET("/api/robno-promet/dobavljaci", h.RobnoPrometDobavljaci)
	r.GET("/api/robno-promet/lager-ruc", h.RobnoPrometRucMain)
	r.GET("/api/robno-promet/ruc/ulaz-izlaz-period", h.RobnoPrometRucUlazIzlaz)
	r.GET("/api/robno-promet/ruc/ruc-po-magacinima", h.RobnoPrometRucMagacinima)
	r.GET("/api/robno-promet/ruc/izlazne-fakture", h.RobnoPrometRucIzlazneFakture)
	r.GET("/api/robno-promet/gradiliste", h.RobnoPrometGradiliste)
	r.GET("/api/robno-promet/gradiliste-vpc-nc", h.RobnoPrometGradilisteVpcNc)
}

func robnoPrometRucSubTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnopromet-ruc-lager", Label: translator.Label("Lager lista"), HXRequestUrl: robnoPrometRucURL, IsActive: true, Name: "ruc-lager"},
		{ID: "robnopromet-ruc-ulaz-izlaz", Label: translator.Label("Ulaz/izlaz za period"), HXRequestUrl: robnoPrometRucUlazIzlazURL, Name: "ruc-ulaz-izlaz"},
		{ID: "robnopromet-ruc-magacinima", Label: translator.Label("RUC po magacinima"), HXRequestUrl: robnoPrometRucMagacinimaURL, Name: "ruc-magacinima"},
		{ID: "robnopromet-ruc-izlazne-fakture", Label: translator.Label("Izlazne fakture"), HXRequestUrl: robnoPrometRucIzlazneFaktureURL, Name: "ruc-izlazne-fakture"},
	}}
}

func robnoPrometTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnopromet-grupe1", Label: translator.Label("Promet po grupi artikala 1"), HXRequestUrl: robnoPrometGrupe1URL, IsActive: true, Name: "grupe1"},
		{ID: "robnopromet-kupci", Label: translator.Label("Promet po kupcima"), HXRequestUrl: robnoPrometKupciURL, Name: "kupci"},
		{ID: "robnopromet-dobavljaci", Label: translator.Label("Nabavka po dobavljačima"), HXRequestUrl: robnoPrometDobavljaciURL, Name: "dobavljaci"},
		{ID: "robnopromet-ruc", Label: translator.Label("Lager lista ulaz/izlaz RUC"), HXRequestUrl: robnoPrometRucURL, Name: "ruc"},
		{ID: "robnopromet-gradiliste", Label: translator.Label("Izveštaj zaduženja gradilišta"), HXRequestUrl: robnoPrometGradilisteURL, Name: "gradiliste"},
		{ID: "robnopromet-gradiliste-vpc", Label: translator.Label("Izveštaj zaduženja gradilišta VPC-NC"), HXRequestUrl: robnoPrometGradVpcURL, Name: "gradiliste-vpc"},
	}}
}
