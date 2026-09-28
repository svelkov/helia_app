package robno

import (
	"fmt"
	"net/http"
	"strings"

	"helia/config"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

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
	robnoPrometRucTableID                 = "robnopromet-ruc"
	robnoPrometRucURL                     = robnoPrometURLPrefix + "/lager-ruc"
	robnoPrometRucURLStampa               = robnoPrometRucURL + "/stampa"
	robnoPrometRucUlazIzlazTitle          = "Ulaz/izlaz za period"
	robnoPrometRucUlazIzlazTableID        = "robnopromet-ruc-ulaz-izlaz-table"
	robnoPrometRucUlazIzlazURL            = robnoPrometURLPrefix + "/ruc/ulaz-izlaz-period"
	robnoPrometRucUlazIzlazURLStampa      = robnoPrometRucUlazIzlazURL + "/stampa"
	robnoPrometRucMagacinimaTitle         = "RUC po magacinima"
	robnoPrometRucMagacinimaTableID       = "robnopromet-ruc-magacinima"
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
	service robnosvc.RobnoPrometService
	cfg     config.Config
	tabs    domain.TabData
	subtabs domain.TabData
}

func NewRobnoPrometHandler(service robnosvc.RobnoPrometService, cfg config.Config) *RobnoPrometHandler {
	return &RobnoPrometHandler{service: service,
		cfg:     cfg,
		tabs:    robnoPrometTabs(),
		subtabs: robnoPrometRucSubTabs(),
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
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "odsifre", "dosifre", "odgrupe", "dogrupe", "oddatuma", "dodatuma"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoPrometParams{
			OdMagacina:     c.Query("odmagacina"),
			DoMagacina:     c.Query("domagacina"),
			OdSifreArtikla: c.Query("odsifre"),
			DoSifreArtikla: c.Query("dosifre"),
			OdGrupe:        c.Query("odgrupe"),
			DoGrupe:        c.Query("dogrupe"),
			OdDatuma:       c.Query("oddatuma"),
			DoDatuma:       c.Query("dodatuma"),
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometArtikala(ctx, &tbl, true, pageSize, page, params); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometArtikala(ctx, &tbl, false, pageSize, page, params); err != nil {
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
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometArtikalURL, "#"+robnoPrometArtikalTableID, "innerHTML", "GET", "", hxValsRobnoPrometArtikal, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("print-btn", "Stampa", "stampa", robnoPrometArtikalURLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma")
	search := common.CreateSearchInput("search-input", translator, robnoPrometArtikalURL, "#"+robnoPrometArtikalTableID, hxValsRobnoPrometArtikal)
	if err := tmpl_robno.RobnoPrometMain(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "odsifre", "dosifre", "odgrupe", "dogrupe", "oddatuma", "dodatuma"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoPrometParams{
			OdMagacina:     c.Query("odmagacina"),
			DoMagacina:     c.Query("domagacina"),
			OdSifreArtikla: c.Query("odsifre"),
			DoSifreArtikla: c.Query("dosifre"),
			OdGrupe:        c.Query("odgrupe"),
			DoGrupe:        c.Query("dogrupe"),
			OdDatuma:       c.Query("oddatuma"),
			DoDatuma:       c.Query("dodatuma"),
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometPoKupcima(ctx, &tbl, true, pageSize, page, params); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometPoKupcima(ctx, &tbl, false, pageSize, page, params); err != nil {
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
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometKupciURL, "#"+robnoPrometKupciTableID, "innerHTML", "GET", "", hxValsRobnoPrometKupci, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometKupciTableID+"-stampa", "Stampa", "stampa", robnoPrometKupciURLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma")
	search := common.CreateSearchInput("search-input", translator, robnoPrometKupciURL, "#"+robnoPrometKupciTableID, hxValsRobnoPrometKupci)
	if err := tmpl_robno.RobnoPrometPoKupcima(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "odsifre", "dosifre", "odgrupe", "dogrupe", "oddatuma", "dodatuma"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoPrometParams{
			OdMagacina:     c.Query("odmagacina"),
			DoMagacina:     c.Query("domagacina"),
			OdSifreArtikla: c.Query("odsifre"),
			DoSifreArtikla: c.Query("dosifre"),
			OdGrupe:        c.Query("odgrupe"),
			DoGrupe:        c.Query("dogrupe"),
			OdDatuma:       c.Query("oddatuma"),
			DoDatuma:       c.Query("dodatuma"),
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetNabavkeOdDobavljaca(ctx, &tbl, true, pageSize, page, params); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetNabavkeOdDobavljaca(ctx, &tbl, false, pageSize, page, params); err != nil {
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
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometDobavljaciURL, "#"+robnoPrometDobavljaciTableID, "innerHTML", "GET", "", hxValsRobnoPrometDobavljaci, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometDobavljaciTableID+"-stampa", "Stampa", "stampa", robnoPrometDobavljaciURLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma")
	search := common.CreateSearchInput("search-input", translator, robnoPrometDobavljaciURL, "#"+robnoPrometDobavljaciTableID, hxValsRobnoPrometDobavljaci)
	if err := tmpl_robno.RobnoPrometPoDobavljacima(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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
	tbl := common.SetTableBasicData(robnoPrometRucTitle, robnoPrometRucTableID, h.service.GetPrometRucTableFields(), robnoPrometRucURL, robnoPrometRucURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucTableID, robnoPrometRucURL, false, false, false)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "stanjenadan"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoPrometParams{
			OdMagacina:     c.Query("odmagacina"),
			DoMagacina:     c.Query("domagacina"),
			OdSifreArtikla: c.Query("odsifre"),
			DoSifreArtikla: c.Query("dosifre"),
			OdGrupe:        c.Query("odgrupe"),
			DoGrupe:        c.Query("dogrupe"),
			OdDatuma:       c.Query("oddatuma"),
			DoDatuma:       c.Query("dodatuma"),
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometRucLagerLista(ctx, &tbl, true, pageSize, page, params); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometRucLagerLista(ctx, &tbl, false, pageSize, page, params); err != nil {
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
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucURL, "#"+robnoPrometRucTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucLager, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucTableID+"-stampa", "Stampa", "stampa", robnoPrometRucURLStampa, "GET", true, common.ClassPrintButton, "tipcene,odmagacina,domagacina,stanjenadan,stampajgrupapodgrupa,zaliheodnule,azbucnired,odgrupe,dogrupe,odpodgrupe,dopodgrupe")
	search := common.CreateSearchInput("search-input", translator, robnoPrometRucURL, "#"+robnoPrometRucTableID, hxValsRobnoPrometRucLager)
	if err := tmpl_robno.RobnoPrometRucLagerLista(tabs, subtabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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

	tbl := common.SetTableBasicData(robnoPrometRucUlazIzlazTitle, robnoPrometRucUlazIzlazTableID, h.service.GetPrometRucTableFields(), robnoPrometRucUlazIzlazURL, robnoPrometRucUlazIzlazURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucUlazIzlazTableID, robnoPrometRucUlazIzlazURL, false, false, false)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "RUC podizveštaj još nije implementiran")
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucUlazIzlazURL, "#"+robnoPrometRucUlazIzlazTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucLager, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucUlazIzlazTableID+"-stampa", "Stampa", "stampa", robnoPrometRucUlazIzlazURLStampa, "GET", true, common.ClassPrintButton, "ulazizlaz,odmagacina,domagacina,oddatuma,dodatuma,zbirmagacina")
	search := common.CreateSearchInput("search-input", translator, robnoPrometRucUlazIzlazURL, "#"+robnoPrometRucUlazIzlazTableID, hxValsRobnoPrometRucUlazIzlaz)
	if err := tmpl_robno.RobnoPrometRucUlazIzlaz(tabs, subtabs, tbl, magValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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
	tbl := common.SetTableBasicData(robnoPrometRucMagacinimaTitle, robnoPrometRucMagacinimaTableID, h.service.GetPrometRucTableFields(), robnoPrometRucMagacinimaURL, robnoPrometRucMagacinimaURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucMagacinimaTableID, robnoPrometRucMagacinimaURL, false, false, false)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "RUC podizveštaj još nije implementiran")
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucMagacinimaURL, "#"+robnoPrometRucMagacinimaTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucLager, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucMagacinimaTableID+"-stampa", "Stampa", "stampa", robnoPrometRucMagacinimaURLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,oddatuma,dodatuma,stampajsamoZbir")
	search := common.CreateSearchInput("search-input", translator, robnoPrometRucMagacinimaURL, "#"+robnoPrometRucMagacinimaTableID, hxValsRobnoPrometRucMagacinima)
	if err := tmpl_robno.RobnoPrometRucMagacinima(tabs, subtabs, tbl, magValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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
	tbl := common.SetTableBasicData(robnoPrometRucIzlazneFaktureTitle, robnoPrometRucIzlazneFaktureTableID, h.service.GetPrometRucTableFields(), robnoPrometRucIzlazneFaktureURL, robnoPrometRucIzlazneFaktureURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometRucIzlazneFaktureTableID, robnoPrometRucIzlazneFaktureURL, false, false, false)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "RUC podizveštaj još nije implementiran")
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometRucIzlazneFaktureURL, "#"+robnoPrometRucIzlazneFaktureTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucLager, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometRucIzlazneFaktureTableID+"-stampa", "Stampa", "stampa", robnoPrometRucIzlazneFaktureURLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,oddatuma,dodatuma,stampajsamoZbir,ukljuceneusluge")
	search := common.CreateSearchInput("search-input", translator, robnoPrometRucIzlazneFaktureURL, "#"+robnoPrometRucIzlazneFaktureTableID, hxValsRobnoPrometRucIzlazne)
	if err := tmpl_robno.RobnoPrometRucIzlazneFakture(tabs, subtabs, tbl, magValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "odsifre", "dosifre", "oddatuma", "dodatuma", "odgrupe", "dogrupe"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoPrometParams{
			OdMagacina:     c.Query("odmagacina"),
			DoMagacina:     c.Query("domagacina"),
			OdSifreArtikla: c.Query("odsifre"),
			DoSifreArtikla: c.Query("dosifre"),
			OdGrupe:        c.Query("odgrupe"),
			DoGrupe:        c.Query("dogrupe"),
			OdDatuma:       c.Query("oddatuma"),
			DoDatuma:       c.Query("dodatuma"),
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometGradilista(ctx, &tbl, true, pageSize, page, params); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometGradilista(ctx, &tbl, false, pageSize, page, params); err != nil {
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
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometGradilisteURL, "#"+robnoPrometGradilisteTableID, "innerHTML", "GET", "", hxValsRobnoPrometGradiliste, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometGradilisteTableID+"-stampa", "Stampa", "stampa", robnoPrometGradilisteURLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma")
	search := common.CreateSearchInput("search-input", translator, robnoPrometGradilisteURL, "#"+robnoPrometGradilisteTableID, hxValsRobnoPrometGradiliste)
	if err := tmpl_robno.RobnoPrometGradiliste(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
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
		fieldsError := common.ValidateRequiredParams(c, []string{"odmagacina", "domagacina", "odsifre", "dosifre", "oddatuma", "dodatuma", "odgrupe", "dogrupe"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoPrometParams{
			OdMagacina:     c.Query("odmagacina"),
			DoMagacina:     c.Query("domagacina"),
			OdSifreArtikla: c.Query("odsifre"),
			DoSifreArtikla: c.Query("dosifre"),
			OdGrupe:        c.Query("odgrupe"),
			DoGrupe:        c.Query("dogrupe"),
			OdDatuma:       c.Query("oddatuma"),
			DoDatuma:       c.Query("dodatuma"),
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPrometGradilisteVpcNc(ctx, &tbl, true, pageSize, page, params); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetPrometGradilisteVpcNc(ctx, &tbl, false, pageSize, page, params); err != nil {
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
	translator := i18n.GetInstance()
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoPrometGradVpcURL, "#"+robnoPrometGradVpcTableID, "innerHTML", "GET", "", hxValsRobnoPrometRucLager, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton(robnoPrometGradVpcTableID+"-stampa", "Stampa", "stampa", robnoPrometGradVpcURLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma")
	search := common.CreateSearchInput("search-input", translator, robnoPrometGradVpcURL, "#"+robnoPrometGradVpcTableID, hxValsRobnoPrometGradiliste)
	if err := tmpl_robno.RobnoPrometGradilisteVpcNc(tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, session.SelectedGod, translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
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
	for _, printURL := range []string{
		robnoPrometArtikalURLStampa,
		robnoPrometKupciURLStampa,
		robnoPrometDobavljaciURLStampa,
		robnoPrometRucURLStampa,
		robnoPrometRucUlazIzlazURLStampa,
		robnoPrometRucMagacinimaURLStampa,
		robnoPrometRucIzlazneFaktureURLStampa,
		robnoPrometGradilisteURLStampa,
		robnoPrometGradVpcURLStampa,
	} {
		r.GET(printURL, h.printNotImplemented)
	}
}

func (h *RobnoPrometHandler) printNotImplemented(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Robno promet stampa jos nije implementirana")
}

func robnoPrometRucSubTabs() domain.TabData {
	translator := i18n.GetInstance()
	return domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnopromet-ruc-lager", Label: translator.Label("Lager lista"), HXRequestUrl: robnoPrometRucURL, IsActive: true, Name: "ruc-lager"},
		{ID: "robnopromet-ruc-ulaz-izlaz", Label: translator.Label("Ulaz/izlaz za period"), HXRequestUrl: robnoPrometRucUlazIzlazURL, Name: "ruc-ulaz-izlaz"},
		{ID: "robnopromet-ruc-magacinima", Label: translator.Label("RUC po magacinima"), HXRequestUrl: robnoPrometRucMagacinimaURL, Name: "ruc-magacinima"},
		{ID: "robnopromet-ruc-izlazne-fakture", Label: translator.Label("Izlazne fakture"), HXRequestUrl: robnoPrometRucIzlazneFaktureURL, Name: "ruc-izlazne-fakture"}},
	}
}
func robnoPrometTabs() domain.TabData {
	translator := i18n.GetInstance()
	return domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnopromet-grupe1", Label: translator.Label("Promet po grupi artikala 1"), HXRequestUrl: robnoPrometArtikalURL, IsActive: true, Name: "grupe1"},
		{ID: "robnopromet-kupci", Label: translator.Label("Promet po kupcima"), HXRequestUrl: robnoPrometKupciURL, Name: "kupci"},
		{ID: "robnopromet-dobavljaci", Label: translator.Label("Nabavka po dobavljačima"), HXRequestUrl: robnoPrometDobavljaciURL, Name: "dobavljaci"},
		{ID: "robnopromet-ruc", Label: translator.Label("Lager lista ulaz/izlaz RUC"), HXRequestUrl: robnoPrometRucURL, Name: "ruc"},
		{ID: "robnopromet-gradiliste", Label: translator.Label("Izveštaj zaduženja gradilišta"), HXRequestUrl: robnoPrometGradilisteURL, Name: "gradiliste"},
		{ID: "robnopromet-gradiliste-vpc", Label: translator.Label("Izveštaj zaduženja gradilišta VPC-NC"), HXRequestUrl: robnoPrometGradVpcURL, Name: "gradiliste-vpc"}},
	}
}
