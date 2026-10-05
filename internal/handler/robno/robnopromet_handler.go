package robno

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"helia/config"
	tmpl_rep_rob "helia/frontend/templates/reports/robno"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

const (
	robnoPrometURLPrefix                  = "/api/robno-promet"
	robnoPrometArtikalTitle               = "Promet artikala po grupama za period"
	robnoPrometArtikalTableID             = "robnopromet-grupe-table"
	robnoPrometArtikalURL                 = robnoPrometURLPrefix + "/grupe-1"
	robnoPrometArtikalURLStampa           = robnoPrometURLPrefix + "/grupe-1/stampa"
	robnoPrometKupciTitle                 = "PROMET ARTIKALA PO KUPCIMA"
	robnoPrometKupciTableID               = "robnopromet-kupci-table"
	robnoPrometKupciURL                   = robnoPrometURLPrefix + "/kupci"
	robnoPrometKupciURLStampa             = robnoPrometKupciURL + "/stampa"
	robnoPrometDobavljaciTitle            = "Nabavka po dobavljačima"
	robnoPrometDobavljaciTableID          = "robnopromet-dobavljaci-table"
	robnoPrometDobavljaciURL              = robnoPrometURLPrefix + "/dobavljaci"
	robnoPrometDobavljaciURLStampa        = robnoPrometDobavljaciURL + "/stampa"
	robnoPrometRucTitle                   = "Lager lista ulaz/izlaz RUC"
	robnoPrometRucTableID                 = "robnopromet-ruc-table"
	robnoPrometRucURL                     = robnoPrometURLPrefix + "/lager-ruc"
	robnoPrometRucURLStampa               = robnoPrometRucURL + "/stampa"
	robnoPrometRucUlazIzlazTitle          = "Ulaz/izlaz za period"
	robnoPrometRucUlazIzlazTableID        = "robnopromet-ruc-ulaz-izlaz-table"
	robnoPrometRucUlazIzlazURL            = robnoPrometURLPrefix + "/ruc/ulaz-izlaz-period"
	robnoPrometRucUlazIzlazURLStampa      = robnoPrometRucUlazIzlazURL + "/stampa"
	robnoPrometRucMagacinimaTitle         = "RUC po magacinima"
	robnoPrometRucMagacinimaTableID       = "robnopromet-ruc-magacinima-table"
	robnoPrometRucMagacinimaURL           = robnoPrometURLPrefix + "/ruc/ruc-po-magacinima"
	robnoPrometRucMagacinimaURLStampa     = robnoPrometRucMagacinimaURL + "/stampa"
	robnoPrometRucIzlazneFaktureTitle     = "Izlazne fakture"
	robnoPrometRucIzlazneFaktureTableID   = "robnopromet-ruc-izlazne-fakture-table"
	robnoPrometRucIzlazneFaktureURL       = robnoPrometURLPrefix + "/ruc/izlazne-fakture"
	robnoPrometRucIzlazneFaktureURLStampa = robnoPrometRucIzlazneFaktureURL + "/stampa"
	robnoPrometGradilisteTitle            = "Izveštaj zaduženja gradilišta"
	robnoPrometGradilisteTableID          = "robnopromet-gradiliste-table"
	robnoPrometGradilisteURL              = robnoPrometURLPrefix + "/gradiliste"
	robnoPrometGradilisteURLStampa        = robnoPrometGradilisteURL + "/stampa"
	robnoPrometGradVpcTitle               = "Izveštaj zaduženja gradilišta VPC-NC"
	robnoPrometGradVpcTableID             = "robnopromet-gradiliste-vpc-table"
	robnoPrometGradVpcURL                 = robnoPrometURLPrefix + "/gradiliste-vpc-nc"
	robnoPrometGradVpcURLStampa           = robnoPrometGradVpcURL + "/stampa"

	// robnoPrometPrintFields are the fields of the panel of the tabs "Promet po grupi artikala",
	// "Promet po kupcima", "Nabavka po dobavljačima" and of the two "Izveštaj zaduženja gradilišta"
	// tabs: the "Stampa" button sends them to the print (openPrintWithParams).
	robnoPrometPrintFields = "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma"
	// robnoPrometRucLagerPrintFields are the fields of the panel of the "Lager lista" sub-tab sent by
	// its "Stampa" button (tipcene is the radio of the price: the checked one is sent).
	// robnoPrometRucUlazIzlazPrintFields are the fields of the panel of the "Ulaz/izlaz za period"
	// sub-tab sent by its "Stampa" button (ulazizlaz is the radio of the promet: the checked one is sent).
	robnoPrometRucUlazIzlazPrintFields = "ulazizlaz,odmagacina,domagacina,oddatuma,dodatuma,zbirmagacina"
	// robnoPrometRucMagaciniPrintFields are the fields of the panel of the "RUC po magacinima" sub-tab
	// sent by its "Stampa" button.
	// robnoPrometRucFakturePrintFields are the fields of the panel of the "Izlazne fakture" sub-tab sent
	// by its "Stampa" button.
	robnoPrometRucFakturePrintFields  = "odmagacina,domagacina,oddatuma,dodatuma,stampajsamoZbir,ukljuceneusluge"
	robnoPrometRucMagaciniPrintFields = "odmagacina,domagacina,oddatuma,dodatuma,stampajsamoZbir"
	robnoPrometRucLagerPrintFields    = "tipcene,odmagacina,domagacina,stanjenadan,stampajgrupapodgrupa,zaliheodnule,azbucnired,odgrupe,dogrupe,odpodgrupe,dopodgrupe"

	hxValsRobnoPrometArtikal = `js:{
		"sourceTab": "",
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"odgrupe": document.getElementById("odgrupe")?.value,
		"dogrupe": document.getElementById("dogrupe")?.value,
		"oddatuma": document.getElementById("oddatuma")?.value,
		"dodatuma": document.getElementById("dodatuma")?.value,
	}`
	hxValsRobnoPrometKupci = `js:{
		"sourceTab": "",
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"odgrupe": document.getElementById("odgrupe")?.value,
		"dogrupe": document.getElementById("dogrupe")?.value,
		"oddatuma": document.getElementById("oddatuma")?.value,
		"dodatuma": document.getElementById("dodatuma")?.value,
	}`
	hxValsRobnoPrometDobavljaci = `js:{
		"sourceTab": "",
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"odgrupe": document.getElementById("odgrupe")?.value,
		"dogrupe": document.getElementById("dogrupe")?.value,
		"oddatuma": document.getElementById("oddatuma")?.value,
		"dodatuma": document.getElementById("dodatuma")?.value,
	}`
	hxValsRobnoPrometRucLager = `js:{
		"sourceTab": "ruc-lager",
		"tipcene": document.querySelector('input[name="tipcene"]:checked')?.value,
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"stanjenadan": document.getElementById("stanjenadan")?.value,
		"stampajgrupapodgrupa": document.getElementById("stampajgrupapodgrupa")?.checked,
		"zaliheodnule": document.getElementById("zaliheodnule")?.checked,
		"azbucnired": document.getElementById("azbucnired")?.checked,
		"odgrupe": document.getElementById("odgrupe")?.value,
		"dogrupe": document.getElementById("dogrupe")?.value,
		"odpodgrupe": document.getElementById("odpodgrupe")?.value,
		"dopodgrupe": document.getElementById("dopodgrupe")?.value,
	}`
	hxValsRobnoPrometRucUlazIzlaz = `js:{
		"sourceTab": "ruc-ulaz-izlaz",
		"ulazizlaz": document.querySelector('input[name="ulazizlaz"]:checked')?.value,
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"oddatuma": document.getElementById("oddatuma")?.value,
		"dodatuma": document.getElementById("dodatuma")?.value,
		"zbirmagacina": document.getElementById("zbirmagacina")?.checked,
	}`
	hxValsRobnoPrometRucMagacinima = `js:{
		"sourceTab": "ruc-magacinima",
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"oddatuma": document.getElementById("oddatuma")?.value,
		"dodatuma": document.getElementById("dodatuma")?.value,
		"stampajsamoZbir": document.getElementById("stampajsamoZbir")?.checked,
	}`
	hxValsRobnoPrometRucIzlazne = `js:{
		"sourceTab": "ruc-izlazne-fakture",
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"oddatuma": document.getElementById("oddatuma")?.value,
		"dodatuma": document.getElementById("dodatuma")?.value,
		"stampajsamoZbir": document.getElementById("stampajsamoZbir")?.checked,
		"ukljuceneusluge": document.getElementById("ukljuceneusluge")?.checked,
	}`
	hxValsRobnoPrometGradiliste = `js:{
		"sourceTab": "",
		"odmagacina": document.getElementById("odmagacina")?.value,
		"domagacina": document.getElementById("domagacina")?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"odgrupe": document.getElementById("odgrupe")?.value,
		"dogrupe": document.getElementById("dogrupe")?.value,
		"oddatuma": document.getElementById("oddatuma")?.value,
		"dodatuma": document.getElementById("dodatuma")?.value,
	}`
)

// robnoPrometRequiredParams are the parameters the grid and the print of the tabs with the panel of
// robnoPrometPrintFields require.
var robnoPrometRequiredParams = []string{"odmagacina", "domagacina", "odsifre", "dosifre", "odgrupe", "dogrupe", "oddatuma", "dodatuma"}

// robnoPrometRucUlazIzlazRequiredParams are the parameters the grid of the "Ulaz/izlaz za period"
// sub-tab requires.
var robnoPrometRucUlazIzlazRequiredParams = []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"}

// robnoPrometRucMagaciniRequiredParams are the parameters the grid and the print of the "RUC po
// magacinima" sub-tab require.
var robnoPrometRucMagaciniRequiredParams = []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"}

// robnoPrometRucLagerRequiredParams are the parameters the grid and the print of the "Lager lista"
// sub-tab require.
var robnoPrometRucLagerRequiredParams = []string{"odmagacina", "domagacina", "stanjenadan"}

func hxValsRobnoPrometGrupeForTab(tabName string) string {
	return strings.Replace(hxValsRobnoPrometArtikal, `"sourceTab": ""`, `"sourceTab": "`+tabName+`"`, 1)
}

// RobnoPrometHandler serves the robno promet report pages. It is created once and
// shared across concurrent requests, so it must not hold mutable per-request state.
//
// tabs/subtabs are immutable tab prototypes: every handler activates a tab through
// common.SetActiveTab, which returns a request-scoped copy. Mutating the prototypes
// directly (or sharing them with a template) would race between concurrent requests.
type RobnoPrometHandler struct {
	translator *i18n.Service
	service    robnosvc.RobnoPrometService
	cfg        config.Config
	tabs       domain.TabData
	subtabs    domain.TabData
}

func NewRobnoPrometHandler(service robnosvc.RobnoPrometService, cfg config.Config, translator *i18n.Service) *RobnoPrometHandler {
	return &RobnoPrometHandler{translator: translator, service: service,
		cfg:     cfg,
		tabs:    robnoPrometTabs(translator),
		subtabs: robnoPrometRucSubTabs(translator),
	}
}

func (h *RobnoPrometHandler) RobnoPrometMain(c *gin.Context) {
	h.RobnoPrometArtikal(c)
}

func (h *RobnoPrometHandler) RobnoPrometArtikal(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 0)

	tbl := common.SetTableBasicData(robnoPrometArtikalTitle, robnoPrometArtikalTableID, h.service.GetPrometArtiklaTableFields(), robnoPrometArtikalURL, robnoPrometArtikalURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometArtikalTableID, robnoPrometArtikalURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometArtikal
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometArtikala(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometArtikala(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometArtikalURL, "#"+robnoPrometArtikalTableID, "innerHTML", "GET", "", hxValsRobnoPrometArtikal, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("print-btn", "Stampa", "stampa", robnoPrometArtikalURLStampa, "GET", true, common.ClassPrintButton, robnoPrometPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometArtikalURL, "#"+robnoPrometArtikalTableID, hxValsRobnoPrometArtikal)
	if err := tmpl_robno.RobnoPrometMain(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometKupci(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 1)
	tbl := common.SetTableBasicData(robnoPrometKupciTitle, robnoPrometKupciTableID, h.service.GetPrometKupcaTableFields(), robnoPrometKupciURL, robnoPrometKupciURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometKupciTitle, robnoPrometKupciURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometKupci
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometPoKupcima(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometPoKupcima(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometKupciURL, "#"+robnoPrometKupciTableID, "innerHTML", "GET", "", hxValsRobnoPrometKupci, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometKupciTableID+"-stampa", "Stampa", "stampa", robnoPrometKupciURLStampa, "GET", true, common.ClassPrintButton, robnoPrometPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometKupciURL, "#"+robnoPrometKupciTableID, hxValsRobnoPrometKupci)
	if err := tmpl_robno.RobnoPrometPoKupcima(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometDobavljaci(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 2)
	tbl := common.SetTableBasicData(robnoPrometDobavljaciTitle, robnoPrometDobavljaciTableID, h.service.GetPrometOdDobavljacaTableFields(), robnoPrometDobavljaciURL, robnoPrometDobavljaciURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometDobavljaciTableID, robnoPrometDobavljaciURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometDobavljaci
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetNabavkeOdDobavljaca(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetNabavkeOdDobavljaca(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometDobavljaciURL, "#"+robnoPrometDobavljaciTableID, "innerHTML", "GET", "", hxValsRobnoPrometDobavljaci, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometDobavljaciTableID+"-stampa", "Stampa", "stampa", robnoPrometDobavljaciURLStampa, "GET", true, common.ClassPrintButton, robnoPrometPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometDobavljaciURL, "#"+robnoPrometDobavljaciTableID, hxValsRobnoPrometDobavljaci)
	if err := tmpl_robno.RobnoPrometPoDobavljacima(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometRucMain(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 3)
	subtabs := common.SetActiveTab(h.subtabs, 0)
	tbl := common.SetTableBasicData(robnoPrometRucTitle, robnoPrometRucTableID, h.service.GetPrometRucLagerListaTableFields(), robnoPrometRucURL, robnoPrometRucURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucTableID, robnoPrometRucURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometRucLager
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRucLagerRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometRucLagerParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometRucLagerLista(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometRucLagerLista(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucURL, "#"+robnoPrometRucTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucLager, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucTableID+"-stampa", "Stampa", "stampa", robnoPrometRucURLStampa, "GET", true, common.ClassPrintButton, robnoPrometRucLagerPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometRucURL, "#"+robnoPrometRucTableID, hxValsRobnoPrometRucLager)
	if err := tmpl_robno.RobnoPrometRucLagerLista(tabs, subtabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometRucUlazIzlaz(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 3)
	subtabs := common.SetActiveTab(h.subtabs, 1)

	tbl := common.SetTableBasicData(robnoPrometRucUlazIzlazTitle, robnoPrometRucUlazIzlazTableID, h.service.GetPrometRucUlazIzlazTableFields(), robnoPrometRucUlazIzlazURL, robnoPrometRucUlazIzlazURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucUlazIzlazTableID, robnoPrometRucUlazIzlazURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometRucUlazIzlaz
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRucUlazIzlazRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometRucUlazIzlazParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometRucUlazIzlaz(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometRucUlazIzlaz(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucUlazIzlazURL, "#"+robnoPrometRucUlazIzlazTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucUlazIzlaz, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucUlazIzlazTableID+"-stampa", "Stampa", "stampa", robnoPrometRucUlazIzlazURLStampa, "GET", true, common.ClassPrintButton, robnoPrometRucUlazIzlazPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometRucUlazIzlazURL, "#"+robnoPrometRucUlazIzlazTableID, hxValsRobnoPrometRucUlazIzlaz)
	if err := tmpl_robno.RobnoPrometRucUlazIzlaz(tabs, subtabs, tbl, magValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometRucMagacinima(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 3)
	subtabs := common.SetActiveTab(h.subtabs, 2)
	tbl := common.SetTableBasicData(robnoPrometRucMagacinimaTitle, robnoPrometRucMagacinimaTableID, h.service.GetPrometRucMagaciniTableFields(), robnoPrometRucMagacinimaURL, robnoPrometRucMagacinimaURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucMagacinimaTableID, robnoPrometRucMagacinimaURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometRucMagacinima
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRucMagaciniRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometRucMagaciniParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometRucMagacini(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometRucMagacini(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucMagacinimaURL, "#"+robnoPrometRucMagacinimaTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucMagacinima, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucMagacinimaTableID+"-stampa", "Stampa", "stampa", robnoPrometRucMagacinimaURLStampa, "GET", true, common.ClassPrintButton, robnoPrometRucMagaciniPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometRucMagacinimaURL, "#"+robnoPrometRucMagacinimaTableID, hxValsRobnoPrometRucMagacinima)
	if err := tmpl_robno.RobnoPrometRucMagacinima(tabs, subtabs, tbl, magValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometRucIzlazneFakture(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 3)
	subtabs := common.SetActiveTab(h.subtabs, 3)
	tbl := common.SetTableBasicData(robnoPrometRucIzlazneFaktureTitle, robnoPrometRucIzlazneFaktureTableID, h.service.GetPrometRucFaktureTableFields(), robnoPrometRucIzlazneFaktureURL, robnoPrometRucIzlazneFaktureURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucIzlazneFaktureTableID, robnoPrometRucIzlazneFaktureURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometRucIzlazne
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRucMagaciniRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometRucFaktureParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometRucFakture(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometRucFakture(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucIzlazneFaktureURL, "#"+robnoPrometRucIzlazneFaktureTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucIzlazne, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucIzlazneFaktureTableID+"-stampa", "Stampa", "stampa", robnoPrometRucIzlazneFaktureURLStampa, "GET", true, common.ClassPrintButton, robnoPrometRucFakturePrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometRucIzlazneFaktureURL, "#"+robnoPrometRucIzlazneFaktureTableID, hxValsRobnoPrometRucIzlazne)
	if err := tmpl_robno.RobnoPrometRucIzlazneFakture(tabs, subtabs, tbl, magValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometGradiliste(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 4)

	tbl := common.SetTableBasicData(robnoPrometGradilisteTitle, robnoPrometGradilisteTableID, h.service.GetPrometGradilistaTableFields(), robnoPrometGradilisteURL, robnoPrometGradilisteURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometGradilisteTableID, robnoPrometGradilisteURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometGradiliste
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometGradilista(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometGradilista(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometGradilisteURL, "#"+robnoPrometGradilisteTableID, "innerHTML", "GET", "", hxValsRobnoPrometGradiliste, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometGradilisteTableID+"-stampa", "Stampa", "stampa", robnoPrometGradilisteURLStampa, "GET", true, common.ClassPrintButton, robnoPrometPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometGradilisteURL, "#"+robnoPrometGradilisteTableID, hxValsRobnoPrometGradiliste)
	if err := tmpl_robno.RobnoPrometGradiliste(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoPrometHandler) RobnoPrometGradilisteVpcNc(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tabs := common.SetActiveTab(h.tabs, 5)
	tbl := common.SetTableBasicData(robnoPrometGradVpcTitle, robnoPrometGradVpcTableID, h.service.GetPrometGradilisteVpcNcTableFields(), robnoPrometGradVpcURL, robnoPrometGradVpcURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometGradVpcTableID, robnoPrometGradVpcURL, false, false, false)
	tbl.Pagination.HxVals = hxValsRobnoPrometGradiliste
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, robnoPrometRequiredParams)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := robnoPrometParamsFromRequest(c)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometGradilisteVpcNc(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometGradilisteVpcNc(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometGradVpcURL, "#"+robnoPrometGradVpcTableID, "innerHTML", "GET", "", hxValsRobnoPrometGradiliste, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometGradVpcTableID+"-stampa", "Stampa", "stampa", robnoPrometGradVpcURLStampa, "GET", true, common.ClassPrintButton, robnoPrometPrintFields)
	search := common.CreateSearchInput("search-input", h.translator, robnoPrometGradVpcURL, "#"+robnoPrometGradVpcTableID, hxValsRobnoPrometGradiliste)
	if err := tmpl_robno.RobnoPrometGradilisteVpcNc(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

// RobnoPrometArtikalStampa prints "Promet artikala po grupama za period" (tab 1): the same columns as
// the grid of the tab without the detail column, the whole selection with its totals.
func (h *RobnoPrometHandler) RobnoPrometArtikalStampa(c *gin.Context) {
	h.prometStampa(c, robnoPrometArtikalTitle, robnoPrometArtikalTableID, "landscape", h.service.GetPrometArtiklaTableFields(), h.prometSelection(), h.service.GetPrometArtikala, tmpl_rep_rob.RobnoPrometArtikalStampa)
}

// RobnoPrometKupciStampa prints "Promet artikala po kupcima" (tab 2).
func (h *RobnoPrometHandler) RobnoPrometKupciStampa(c *gin.Context) {
	h.prometStampa(c, robnoPrometKupciTitle, robnoPrometKupciTableID, "landscape", h.service.GetPrometKupcaTableFields(), h.prometSelection(), h.service.GetPrometPoKupcima, tmpl_rep_rob.RobnoPrometKupciStampa)
}

// RobnoPrometDobavljaciStampa prints "Nabavka po dobavljačima" (tab 3).
func (h *RobnoPrometHandler) RobnoPrometDobavljaciStampa(c *gin.Context) {
	h.prometStampa(c, robnoPrometDobavljaciTitle, robnoPrometDobavljaciTableID, "landscape", h.service.GetPrometOdDobavljacaTableFields(), h.prometSelection(), h.service.GetNabavkeOdDobavljaca, tmpl_rep_rob.RobnoPrometDobavljaciStampa)
}

// RobnoPrometGradilisteStampa prints "Izveštaj zaduženja gradilišta" (tab 5).
func (h *RobnoPrometHandler) RobnoPrometGradilisteStampa(c *gin.Context) {
	h.prometStampa(c, robnoPrometGradilisteTitle, robnoPrometGradilisteTableID, "portrait", h.service.GetPrometGradilistaTableFields(), h.prometSelection(), h.service.GetPrometGradilista, tmpl_rep_rob.RobnoPrometGradilisteStampa)
}

// RobnoPrometGradilisteVpcNcStampa prints "Izveštaj zaduženja gradilišta VPC-NC" (tab 6).
func (h *RobnoPrometHandler) RobnoPrometGradilisteVpcNcStampa(c *gin.Context) {
	h.prometStampa(c, robnoPrometGradVpcTitle, robnoPrometGradVpcTableID, "portrait", h.service.GetPrometGradilisteVpcNcTableFields(), h.prometSelection(), h.service.GetPrometGradilisteVpcNc, tmpl_rep_rob.RobnoPrometGradilisteVpcNcStampa)
}

// RobnoPrometRucLagerStampa prints the "Lager lista" (tab 4, sub-tab "Lager lista"): the stanje of the
// articles on the date of the selection valued with the chosen price, the same columns as the grid of
// the sub-tab without the detail column and the total vrednost.
func (h *RobnoPrometHandler) RobnoPrometRucLagerStampa(c *gin.Context) {
	h.prometStampa(c, robnoPrometRucTitle, robnoPrometRucTableID, "portrait", h.service.GetPrometRucLagerListaTableFields(), h.prometRucLagerSelection(c.Request.Context()), h.service.GetPrometRucLagerLista, tmpl_rep_rob.RobnoPrometRucLagerStampa)
}

// RobnoPrometRucUlazIzlazStampa prints the "Ulaz/izlaz za period" (tab 4, sub-tab "Ulaz/izlaz za
// period"): the columns of the grid of the sub-tab for the chosen promet, without the detail column and
// the columns of the producer, all the rows and the totals.
func (h *RobnoPrometHandler) RobnoPrometRucUlazIzlazStampa(c *gin.Context) {
	h.prometStampa(c, robnoPrometRucUlazIzlazTitle, robnoPrometRucUlazIzlazTableID, "landscape", h.service.GetPrometRucUlazIzlazTableFields(), h.prometRucUlazIzlazSelection(c.Request.Context()), h.service.GetPrometRucUlazIzlaz, tmpl_rep_rob.RobnoPrometRucUlazIzlazStampa)
}

// prometRucUlazIzlazSelection is the selection of the "Ulaz/izlaz za period" sub-tab.
func (h *RobnoPrometHandler) prometRucUlazIzlazSelection(ctx context.Context) robnoPrometSelection {
	return robnoPrometSelection{
		required: robnoPrometRucUlazIzlazRequiredParams,
		read:     robnoPrometRucUlazIzlazParamsFromRequest,
		reportParams: func(params domain.RobnoPrometParams) []domain.ParameterItem {
			return h.prometRucUlazIzlazReportParams(ctx, params)
		},
	}
}

// RobnoPrometRucMagaciniStampa prints the "RUC po magacinima" (tab 4, sub-tab "RUC po magacinima") like
// the legacy "PREGLED RUC-a PO MAGACINIMA": the articles under the header of their magacin with the total
// of every magacin (only the totals with "Štampaj samo zbir po magacinima") and the total of the list.
func (h *RobnoPrometHandler) RobnoPrometRucMagaciniStampa(c *gin.Context) {
	ctx := c.Request.Context()
	selection := robnoPrometSelection{
		required: robnoPrometRucMagaciniRequiredParams,
		read:     robnoPrometRucMagaciniParamsFromRequest,
		reportParams: func(params domain.RobnoPrometParams) []domain.ParameterItem {
			return h.prometRucMagaciniReportParams(ctx, params)
		},
		title: func(domain.RobnoPrometParams) string { return "Pregled RUC-a po magacinima" },
	}
	h.prometStampa(c, robnoPrometRucMagacinimaTitle, robnoPrometRucMagacinimaTableID, "landscape", h.service.GetPrometRucMagaciniTableFields(), selection, h.service.GetPrometRucMagacini, tmpl_rep_rob.RobnoPrometRucMagaciniStampa)
}

// RobnoPrometRucFaktureStampa prints the "Izlazne fakture" (tab 4, sub-tab "Izlazne fakture") like the
// legacy "PREGLED FAKTURA": the fakture under the header of their magacin with the total of every
// magacin (only the totals with "Štampaj samo zbir po magacinima") and the total of the list.
func (h *RobnoPrometHandler) RobnoPrometRucFaktureStampa(c *gin.Context) {
	ctx := c.Request.Context()
	selection := robnoPrometSelection{
		required: robnoPrometRucMagaciniRequiredParams,
		read:     robnoPrometRucFaktureParamsFromRequest,
		reportParams: func(params domain.RobnoPrometParams) []domain.ParameterItem {
			return h.prometRucFaktureReportParams(ctx, params)
		},
		title: func(domain.RobnoPrometParams) string { return "Pregled faktura" },
	}
	h.prometStampa(c, robnoPrometRucIzlazneFaktureTitle, robnoPrometRucIzlazneFaktureTableID, "landscape", h.service.GetPrometRucFaktureTableFields(), selection, h.service.GetPrometRucFakture, tmpl_rep_rob.RobnoPrometRucFaktureStampa)
}

// robnoPrometSelection is how the print of a robno promet tab reads its selection: the parameters it
// requires, the reader of the parameters, the builder of the parameters of the report header and,
// optionally, the title of the report when it depends on the selection (nil: the title of the tab).
type robnoPrometSelection struct {
	required     []string
	read         func(*gin.Context) domain.RobnoPrometParams
	reportParams func(domain.RobnoPrometParams) []domain.ParameterItem
	title        func(domain.RobnoPrometParams) string
}

// prometSelection is the selection of the tabs with the panel of robnoPrometPrintFields (magacini,
// šifre artikla, grupe and dates).
func (h *RobnoPrometHandler) prometSelection() robnoPrometSelection {
	return robnoPrometSelection{required: robnoPrometRequiredParams, read: robnoPrometParamsFromRequest, reportParams: h.prometReportParams}
}

// prometRucLagerSelection is the selection of the "Lager lista" sub-tab: the header shows the magacini
// with their naziv, like the legacy print, and the title the chosen tip cene.
func (h *RobnoPrometHandler) prometRucLagerSelection(ctx context.Context) robnoPrometSelection {
	return robnoPrometSelection{
		required: robnoPrometRucLagerRequiredParams,
		read:     robnoPrometRucLagerParamsFromRequest,
		reportParams: func(params domain.RobnoPrometParams) []domain.ParameterItem {
			return h.prometRucLagerReportParams(ctx, params)
		},
		title: robnoPrometRucLagerTitle,
	}
}

// robnoPrometRucLagerTitle is the title of the print of the "Lager lista": the legacy print names the
// tip cene in the title.
func robnoPrometRucLagerTitle(params domain.RobnoPrometParams) string {
	switch params.TipCene {
	case "prosecnanabavna":
		return "Lager lista po pros. nab. cenama"
	case "prodajna":
		return "Lager lista po prodajnim cenama"
	default:
		return "Lager lista po neto fakt. cenama"
	}
}

// robnoPrometFetch is the service function of a robno promet tab: the same function fills the grid
// (common.TipStampePreview) and the print (common.TipStampePrint).
type robnoPrometFetch func(context.Context, *domain.TableData, bool, int, int, domain.RobnoPrometParams, string) error

// robnoPrometReport is the report template of the print of a robno promet tab.
type robnoPrometReport func(domain.ReportParameters, domain.TableData, []domain.ParameterItem, *i18n.Service) templ.Component

// prometStampa renders the print of a robno promet tab: the selection of the panel of the tab is read
// from the request (selection), the service fills the table of the tab in the print mode (the columns of the
// grid without the detail column, all the rows and the totals) and the report renders it under the
// header with the firm and the selection.
func (h *RobnoPrometHandler) prometStampa(c *gin.Context, title, tableID, orientation string, fields []domain.Fields, selection robnoPrometSelection, fetch robnoPrometFetch, report robnoPrometReport) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	if fieldsError := common.ValidateRequiredParams(c, selection.required); len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
		return
	}
	params := selection.read(c)
	if selection.title != nil {
		title = selection.title(params)
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tbl := common.SetTableBasicData(title, tableID, fields, "", "", 0, 0, 0, 0, h.cfg)
	if err := fetch(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := domain.ReportParameters{
		Orientation: orientation,
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  title,
	}
	if err := report(repParams, tbl, selection.reportParams(params), h.translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

// prometReportParams builds the parameters of the report header of a robno promet print in a fixed
// order: the ranges of the magacini, of the šifre artikla, of the grupe and of the dates.
func (h *RobnoPrometHandler) prometReportParams(params domain.RobnoPrometParams) []domain.ParameterItem {
	items := []domain.ParameterItem{}
	add := func(label, value string) {
		if value == "" {
			return
		}
		items = append(items, domain.ParameterItem{Name: h.translator.Label(label), Value: value})
	}
	add("Od magacina", params.OdMagacina)
	add("Do magacina", params.DoMagacina)
	add("Od šifre artikla", params.OdSifreArtikla)
	add("Do šifre artikla", params.DoSifreArtikla)
	add("Od grupe", params.OdGrupe)
	add("Do grupe", params.DoGrupe)
	add("Od datuma", robnoPrometDatum(params.OdDatuma))
	add("Do datuma", robnoPrometDatum(params.DoDatuma))
	return items
}

// prometRucLagerReportParams builds the parameters of the report header of the "Lager lista" like the
// legacy print: od magacina and do magacina (with the naziv of the magacin) and the date of the stanje;
// then the options of the list that change its content (the grupe and podgrupe of the selection and
// "Zalihe <> 0").
func (h *RobnoPrometHandler) prometRucLagerReportParams(ctx context.Context, params domain.RobnoPrometParams) []domain.ParameterItem {
	magacin := h.prometMagacinNaziv(ctx)
	items := []domain.ParameterItem{}
	add := func(label, value string) {
		if value == "" {
			return
		}
		items = append(items, domain.ParameterItem{Name: h.translator.Label(label), Value: value})
	}
	add("Od magacina", magacin(params.OdMagacina))
	add("Do magacina", magacin(params.DoMagacina))
	add("Stanje na dan", robnoPrometDatum(params.StanjeNaDan))
	if params.ZaliheOdNule {
		items = append(items, domain.ParameterItem{Value: h.translator.Label("Zalihe <> 0")})
	}
	return items
}

// prometMagacinNaziv returns the naziv of a magacin for the headers of the prints ("1 - veleprodaja"):
// the magacin itself when it is not found.
func (h *RobnoPrometHandler) prometMagacinNaziv(ctx context.Context) func(string) string {
	magacini := map[string]string{}
	if values, err := h.service.GetMagacinComboValues(ctx); err == nil {
		for _, item := range values {
			magacini[item.Key] = item.Value
		}
	}
	return func(mag string) string {
		if naziv, found := magacini[mag]; found {
			return naziv
		}
		return mag
	}
}

// prometRucUlazIzlazReportParams builds the parameters of the report header of the "Ulaz/izlaz za
// period": the promet, the magacini (with their naziv), the period and the sum of the magacini.
func (h *RobnoPrometHandler) prometRucUlazIzlazReportParams(ctx context.Context, params domain.RobnoPrometParams) []domain.ParameterItem {
	magacin := h.prometMagacinNaziv(ctx)
	items := []domain.ParameterItem{}
	add := func(label, value string) {
		if value == "" {
			return
		}
		items = append(items, domain.ParameterItem{Name: h.translator.Label(label), Value: value})
	}
	promet := "Ulaz i izlaz robe"
	switch params.UlazIzlaz {
	case "ulaz":
		promet = "Ulaz robe"
	case "izlaz":
		promet = "Izlaz robe"
	}
	items = append(items, domain.ParameterItem{Value: h.translator.Label(promet)})
	add("Od magacina", magacin(params.OdMagacina))
	add("Do magacina", magacin(params.DoMagacina))
	add("Od datuma", robnoPrometDatum(params.OdDatuma))
	add("Do datuma", robnoPrometDatum(params.DoDatuma))
	if params.ZbirMagacina {
		items = append(items, domain.ParameterItem{Value: h.translator.Label("Zbir izabranih magacina")})
	}
	return items
}

// robnoPrometParamsFromRequest reads the selection of the panel of the tabs with the fields of
// robnoPrometPrintFields (the grid and the print of a tab read the same parameters).
func robnoPrometParamsFromRequest(c *gin.Context) domain.RobnoPrometParams {
	return domain.RobnoPrometParams{
		OdMagacina:     c.Query("odmagacina"),
		DoMagacina:     c.Query("domagacina"),
		OdSifreArtikla: c.Query("odsifre"),
		DoSifreArtikla: c.Query("dosifre"),
		OdGrupe:        c.Query("odgrupe"),
		DoGrupe:        c.Query("dogrupe"),
		OdDatuma:       c.Query("oddatuma"),
		DoDatuma:       c.Query("dodatuma"),
	}
}

// robnoPrometRucLagerParamsFromRequest reads the selection of the "Lager lista" sub-tab: the price
// of the lager (the radio "tipcene"), the magacini, the date of the stanje, the grupe and podgrupe and
// the options of the list.
func robnoPrometRucLagerParamsFromRequest(c *gin.Context) domain.RobnoPrometParams {
	return domain.RobnoPrometParams{
		TipCene:              c.Query("tipcene"),
		OdMagacina:           c.Query("odmagacina"),
		DoMagacina:           c.Query("domagacina"),
		StanjeNaDan:          c.Query("stanjenadan"),
		OdGrupe:              c.Query("odgrupe"),
		DoGrupe:              c.Query("dogrupe"),
		OdPodgrupe:           c.Query("odpodgrupe"),
		DoPodgrupe:           c.Query("dopodgrupe"),
		ZaliheOdNule:         c.Query("zaliheodnule") == "true",
		AzbucniRed:           c.Query("azbucnired") == "true",
		StampajGrupaPodgrupa: c.Query("stampajgrupapodgrupa") == "true",
	}
}

// robnoPrometRucUlazIzlazParamsFromRequest reads the selection of the "Ulaz/izlaz za period" sub-tab:
// the promet (the radio "ulazizlaz": ulaz, izlaz or ulazizlaz), the magacini, the period and the sum of
// the magacini.
func robnoPrometRucUlazIzlazParamsFromRequest(c *gin.Context) domain.RobnoPrometParams {
	return domain.RobnoPrometParams{
		UlazIzlaz:    c.Query("ulazizlaz"),
		OdMagacina:   c.Query("odmagacina"),
		DoMagacina:   c.Query("domagacina"),
		OdDatuma:     c.Query("oddatuma"),
		DoDatuma:     c.Query("dodatuma"),
		ZbirMagacina: c.Query("zbirmagacina") == "true",
	}
}

// robnoPrometRucMagaciniParamsFromRequest reads the selection of the "RUC po magacinima" sub-tab: the
// magacini, the period and "Štampaj samo zbir po magacinima".
func robnoPrometRucMagaciniParamsFromRequest(c *gin.Context) domain.RobnoPrometParams {
	return domain.RobnoPrometParams{
		OdMagacina:      c.Query("odmagacina"),
		DoMagacina:      c.Query("domagacina"),
		OdDatuma:        c.Query("oddatuma"),
		DoDatuma:        c.Query("dodatuma"),
		StampajSamoZbir: c.Query("stampajsamoZbir") == "true",
	}
}

// robnoPrometRucFaktureParamsFromRequest reads the selection of the "Izlazne fakture" sub-tab: the
// magacini, the period, "Štampaj samo zbir po magacinima" and "Uključene usluge".
func robnoPrometRucFaktureParamsFromRequest(c *gin.Context) domain.RobnoPrometParams {
	params := robnoPrometRucMagaciniParamsFromRequest(c)
	params.UkljuceneUsluge = c.Query("ukljuceneusluge") == "true"
	return params
}

// prometRucFaktureReportParams builds the parameters of the report header of the "Izlazne fakture"
// like the legacy print: the magacini (with their naziv), the period and the state of "Uključene
// usluge".
func (h *RobnoPrometHandler) prometRucFaktureReportParams(ctx context.Context, params domain.RobnoPrometParams) []domain.ParameterItem {
	items := h.prometRucMagaciniReportParams(ctx, params)
	usluge := "☐"
	if params.UkljuceneUsluge {
		usluge = "☑"
	}
	return append(items, domain.ParameterItem{Name: h.translator.Label("Uključene usluge"), Value: usluge})
}

// prometRucMagaciniReportParams builds the parameters of the report header of the "RUC po magacinima"
// like the legacy print: the magacini (with their naziv) and the period.
func (h *RobnoPrometHandler) prometRucMagaciniReportParams(ctx context.Context, params domain.RobnoPrometParams) []domain.ParameterItem {
	magacin := h.prometMagacinNaziv(ctx)
	items := []domain.ParameterItem{}
	add := func(label, value string) {
		if value == "" {
			return
		}
		items = append(items, domain.ParameterItem{Name: h.translator.Label(label), Value: value})
	}
	add("Od magacina", magacin(params.OdMagacina))
	add("Do magacina", magacin(params.DoMagacina))
	add("Od datuma", robnoPrometDatum(params.OdDatuma))
	add("Do datuma", robnoPrometDatum(params.DoDatuma))
	if params.StampajSamoZbir {
		items = append(items, domain.ParameterItem{Value: h.translator.Label("Štampaj samo zbir po magacinima")})
	}
	return items
}

// robnoPrometDatum renders a date of the selection (yyyy-mm-dd) in the report header (dd.mm.yyyy).
func robnoPrometDatum(value string) string {
	parsed, err := time.Parse(common.HtmlLayout, value)
	if err != nil {
		return value
	}
	return parsed.Format(common.DateLayout)
}

// AddRoutes registers the robno promet routes.
//
// Auth middleware is registered once, globally, in setupRouter. Do NOT call r.Use
// here: gin appends to the engine-wide chain, so every feature doing this makes the
// middleware run once per previously registered feature on every request.
func (h *RobnoPrometHandler) AddRoutes(r *gin.Engine) {
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
	r.GET(robnoPrometArtikalURLStampa, h.RobnoPrometArtikalStampa)
	r.GET(robnoPrometKupciURLStampa, h.RobnoPrometKupciStampa)
	r.GET(robnoPrometDobavljaciURLStampa, h.RobnoPrometDobavljaciStampa)
	r.GET(robnoPrometGradilisteURLStampa, h.RobnoPrometGradilisteStampa)
	r.GET(robnoPrometGradVpcURLStampa, h.RobnoPrometGradilisteVpcNcStampa)
	r.GET(robnoPrometRucURLStampa, h.RobnoPrometRucLagerStampa)
	r.GET(robnoPrometRucUlazIzlazURLStampa, h.RobnoPrometRucUlazIzlazStampa)
	r.GET(robnoPrometRucMagacinimaURLStampa, h.RobnoPrometRucMagaciniStampa)
	r.GET(robnoPrometRucIzlazneFaktureURLStampa, h.RobnoPrometRucFaktureStampa)
}

func (h *RobnoPrometHandler) printNotImplemented(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Robno promet stampa jos nije implementirana")
}

func robnoPrometRucSubTabs(translator *i18n.Service) domain.TabData {
	return domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnopromet-ruc-lager", Label: translator.Label("Lager lista"), HXRequestUrl: robnoPrometRucURL, IsActive: true, Name: "ruc-lager"},
		{ID: "robnopromet-ruc-ulaz-izlaz", Label: translator.Label("Ulaz/izlaz za period"), HXRequestUrl: robnoPrometRucUlazIzlazURL, Name: "ruc-ulaz-izlaz"},
		{ID: "robnopromet-ruc-magacinima", Label: translator.Label("RUC po magacinima"), HXRequestUrl: robnoPrometRucMagacinimaURL, Name: "ruc-magacinima"},
		{ID: "robnopromet-ruc-izlazne-fakture", Label: translator.Label("Izlazne fakture"), HXRequestUrl: robnoPrometRucIzlazneFaktureURL, Name: "ruc-izlazne-fakture"}},
	}
}
func robnoPrometTabs(translator *i18n.Service) domain.TabData {
	return domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnopromet-grupe1", Label: translator.Label("Promet po grupi artikala 1"), HXRequestUrl: robnoPrometArtikalURL, IsActive: true, Name: "grupe1"},
		{ID: "robnopromet-kupci", Label: translator.Label("Promet po kupcima"), HXRequestUrl: robnoPrometKupciURL, Name: "kupci"},
		{ID: "robnopromet-dobavljaci", Label: translator.Label("Nabavka po dobavljačima"), HXRequestUrl: robnoPrometDobavljaciURL, Name: "dobavljaci"},
		{ID: "robnopromet-ruc", Label: translator.Label("Lager lista ulaz/izlaz RUC"), HXRequestUrl: robnoPrometRucURL, Name: "ruc"},
		{ID: "robnopromet-gradiliste", Label: translator.Label("Izveštaj zaduženja gradilišta"), HXRequestUrl: robnoPrometGradilisteURL, Name: "gradiliste"},
		{ID: "robnopromet-gradiliste-vpc", Label: translator.Label("Izveštaj zaduženja gradilišta VPC-NC"), HXRequestUrl: robnoPrometGradVpcURL, Name: "gradiliste-vpc"}},
	}
}
