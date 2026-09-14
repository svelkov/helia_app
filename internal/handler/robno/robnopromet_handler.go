package robno

import (
	"context"
	"net/http"
	"strings"

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
	robnoPrometURLPrefix                  = "/api/robno-promet"
	robnoPrometGrupe1Title                = "Promet artikala po grupama za period"
	robnoPrometGrupe1TableID              = "robnopromet-grupe"
	robnoPrometGrupe1URL                  = robnoPrometURLPrefix + "/grupe-1"
	robnoPrometGrupe1URLStampa            = robnoPrometURLPrefix + "/grupe-1/stampa"
	robnoPrometKupciTitle                 = "Promet artikala po kupcima za period"
	robnoPrometKupciTableID               = "robnopromet-kupci"
	robnoPrometKupciURL                   = robnoPrometURLPrefix + "/kupci"
	robnoPrometKupciURLStampa             = robnoPrometKupciURL + "/stampa"
	robnoPrometDobavljaciTitle            = "Nabavka po dobavljačima"
	robnoPrometDobavljaciTableID          = "robnopromet-dobavljaci"
	robnoPrometDobavljaciURL              = robnoPrometURLPrefix + "/dobavljaci"
	robnoPrometDobavljaciURLStampa        = robnoPrometDobavljaciURL + "/stampa"
	robnoPrometRucTitle                   = "Lager lista ulaz/izlaz RUC"
	robnoPrometRucTableID                 = "robnopromet-ruc"
	robnoPrometRucURL                     = robnoPrometURLPrefix + "/lager-ruc"
	robnoPrometRucURLStampa               = robnoPrometRucURL + "/stampa"
	robnoPrometRucUlazIzlazTitle          = "Ulaz/izlaz za period"
	robnoPrometRucUlazIzlazTableID        = "robnopromet-ruc-ulaz-izlaz"
	robnoPrometRucUlazIzlazURL            = robnoPrometURLPrefix + "/ruc/ulaz-izlaz-period"
	robnoPrometRucUlazIzlazURLStampa      = robnoPrometRucUlazIzlazURL + "/stampa"
	robnoPrometRucMagacinimaTitle         = "RUC po magacinima"
	robnoPrometRucMagacinimaTableID       = "robnopromet-ruc-magacinima"
	robnoPrometRucMagacinimaURL           = robnoPrometURLPrefix + "/ruc/ruc-po-magacinima"
	robnoPrometRucMagacinimaURLStampa     = robnoPrometRucMagacinimaURL + "/stampa"
	robnoPrometRucIzlazneFaktureTitle     = "Izlazne fakture"
	robnoPrometRucIzlazneFaktureTableID   = "robnopromet-ruc-izlazne-fakture"
	robnoPrometRucIzlazneFaktureURL       = robnoPrometURLPrefix + "/ruc/izlazne-fakture"
	robnoPrometRucIzlazneFaktureURLStampa = robnoPrometRucIzlazneFaktureURL + "/stampa"
	robnoPrometGradilisteTitle            = "Izveštaj zaduženja gradilišta"
	robnoPrometGradilisteTableID          = "robnopromet-gradiliste"
	robnoPrometGradilisteURL              = robnoPrometURLPrefix + "/gradiliste"
	robnoPrometGradilisteURLStampa        = robnoPrometGradilisteURL + "/stampa"
	robnoPrometGradVpcTitle               = "Izveštaj zaduženja gradilišta VPC-NC"
	robnoPrometGradVpcTableID             = "robnopromet-gradiliste-vpc"
	robnoPrometGradVpcURL                 = robnoPrometURLPrefix + "/gradiliste-vpc-nc"
	robnoPrometGradVpcURLStampa           = robnoPrometGradVpcURL + "/stampa"

	hxValsRobnoPrometGrupe = `js:{
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
)

func hxValsRobnoPrometGrupeForTab(tabName string) string {
	return strings.Replace(hxValsRobnoPrometGrupe, `"sourceTab": ""`, `"sourceTab": "`+tabName+`"`, 1)
}

type RobnoPrometHandler struct {
	service robnosvc.RobnoPrometService
	cfg     config.Config
	tabs    *domain.TabData
	subtabs *domain.TabData
}

func NewRobnoPrometHandler(service robnosvc.RobnoPrometService, cfg config.Config) *RobnoPrometHandler {
	return &RobnoPrometHandler{service: service, cfg: cfg, tabs: robnoPrometTabs(), subtabs: robnoPrometRucSubTabs()}
}

func (h *RobnoPrometHandler) RobnoPrometMain(c *gin.Context) {
	h.RobnoPrometArtikal(c)
}

func (h *RobnoPrometHandler) RobnoPrometArtikal(c *gin.Context) {
	ctx, _, ok := h.context(c, 0)
	if !ok {
		return
	}
	tbl := common.SetTableBasicData(robnoPrometGrupe1Title, robnoPrometGrupe1TableID, h.service.GetPrometArtiklaTableFields(), robnoPrometGrupe1URL, robnoPrometGrupe1URL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoPrometGrupe1TableID, robnoPrometGrupe1URL, false, false, false)
	if common.IsDataRequest(c) {
		if !h.validate(c, []string{"odmagacina", "domagacina", "odsifre", "dosifre", "odgrupe", "dogrupe", "oddatuma", "dodatuma"}) {
			return
		}
		if !h.getPaginatedReport(c, &tbl, h.service.GetPrometArtikala) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, grupeValues, err := h.comboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	translator := i18n.GetInstance()
	groupVals := hxValsRobnoPrometGrupeForTab("grupe1")
	btnObrada := h.obradaButton(robnoPrometGrupe1URL, robnoPrometGrupe1TableID, groupVals)
	btnPrint := common.SetPrintButton("print-btn", "Stampa", "stampa", robnoPrometGrupe1URLStampa, "GET", true, common.ClassPrintButton, "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma")
	search := common.CreateSearchInput("search-input", translator, robnoPrometGrupe1URL, "#"+robnoPrometGrupe1TableID, groupVals)
	if err := tmpl_robno.RobnoPrometMain(*h.tabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

func (h *RobnoPrometHandler) RobnoPrometKupci(c *gin.Context) {
	ctx, _, ok := h.context(c, 1)
	if !ok {
		return
	}
	h.renderStandard(c, ctx, robnoPrometKupciTitle, robnoPrometKupciTableID, robnoPrometKupciURL, h.service.GetPrometKupcaTableFields(), h.service.GetPrometPoKupcima, robnoPrometKupciURLStampa, hxValsRobnoPrometGrupeForTab("kupci"), []string{"odmagacina", "domagacina", "odsifre", "dosifre", "odgrupe", "dogrupe", "oddatuma", "dodatuma"}, func(tbl domain.TableData, magValues, grupeValues []domain.ComboItem, b, p domain.Button, s domain.InputControl) error {
		return tmpl_robno.RobnoPrometPoKupcima(*h.tabs, tbl, magValues, grupeValues, b, p, s, i18n.GetInstance()).Render(ctx, c.Writer)
	})
}

func (h *RobnoPrometHandler) RobnoPrometDobavljaci(c *gin.Context) {
	ctx, _, ok := h.context(c, 2)
	if !ok {
		return
	}
	h.renderStandard(c, ctx, robnoPrometDobavljaciTitle, robnoPrometDobavljaciTableID, robnoPrometDobavljaciURL, h.service.GetPrometOdDobavljacaTableFields(), h.service.GetNabavkeOdDobavljaca, robnoPrometDobavljaciURLStampa, hxValsRobnoPrometGrupeForTab("dobavljaci"), []string{"odmagacina", "domagacina", "odsifre", "dosifre", "odgrupe", "dogrupe", "oddatuma", "dodatuma"}, func(tbl domain.TableData, magValues, grupeValues []domain.ComboItem, b, p domain.Button, s domain.InputControl) error {
		return tmpl_robno.RobnoPrometPoDobavljacima(*h.tabs, tbl, magValues, grupeValues, b, p, s, i18n.GetInstance()).Render(ctx, c.Writer)
	})
}

func (h *RobnoPrometHandler) RobnoPrometRucMain(c *gin.Context) {
	h.renderRuc(c, 0)
}
func (h *RobnoPrometHandler) RobnoPrometRucUlazIzlaz(c *gin.Context)      { h.renderRuc(c, 1) }
func (h *RobnoPrometHandler) RobnoPrometRucMagacinima(c *gin.Context)     { h.renderRuc(c, 2) }
func (h *RobnoPrometHandler) RobnoPrometRucIzlazneFakture(c *gin.Context) { h.renderRuc(c, 3) }

func (h *RobnoPrometHandler) renderRuc(c *gin.Context, subIndex int) {
	ctx, _, ok := h.context(c, 3)
	if !ok {
		return
	}
	common.SetActiveTab(h.subtabs, subIndex)
	var title, tableID, url, printURL, vals string
	fields := h.service.GetPrometRucTableFields()
	var required []string
	switch subIndex {
	case 1:
		title, tableID, url, printURL, vals = robnoPrometRucUlazIzlazTitle, robnoPrometRucUlazIzlazTableID, robnoPrometRucUlazIzlazURL, robnoPrometRucUlazIzlazURLStampa, hxValsRobnoPrometRucUlazIzlaz
		required = []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"}
	case 2:
		title, tableID, url, printURL, vals = robnoPrometRucMagacinimaTitle, robnoPrometRucMagacinimaTableID, robnoPrometRucMagacinimaURL, robnoPrometRucMagacinimaURLStampa, hxValsRobnoPrometRucMagacinima
		required = []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"}
	case 3:
		title, tableID, url, printURL, vals = robnoPrometRucIzlazneFaktureTitle, robnoPrometRucIzlazneFaktureTableID, robnoPrometRucIzlazneFaktureURL, robnoPrometRucIzlazneFaktureURLStampa, hxValsRobnoPrometRucIzlazne
		required = []string{"odmagacina", "domagacina", "oddatuma", "dodatuma"}
	default:
		title, tableID, url, printURL, vals = robnoPrometRucTitle, robnoPrometRucTableID, robnoPrometRucURL, robnoPrometRucURLStampa, hxValsRobnoPrometRucLager
		required = []string{"odmagacina", "domagacina", "stanjenadan"}
	}
	tbl := common.SetTableBasicData(title, tableID, fields, url, url, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, tableID, url, false, false, false)
	if common.IsDataRequest(c) {
		if !h.validate(c, required) {
			return
		}
		if subIndex != 0 {
			common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "RUC podizveštaj još nije implementiran")
			return
		}
		if subIndex == 0 && !h.getPaginatedReport(c, &tbl, h.service.GetPrometRucLagerLista) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(url, tableID, vals)
	btnPrint := common.SetPrintButton(tableID+"-stampa", "Stampa", "stampa", printURL, "GET", true, common.ClassPrintButton, h.printFields(subIndex))
	search := common.CreateSearchInput("search-input", translator, url, "#"+tableID, vals)
	var renderErr error
	switch subIndex {
	case 1:
		renderErr = tmpl_robno.RobnoPrometRucUlazIzlaz(*h.tabs, *h.subtabs, tbl, magValues, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer)
	case 2:
		renderErr = tmpl_robno.RobnoPrometRucMagacinima(*h.tabs, *h.subtabs, tbl, magValues, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer)
	case 3:
		renderErr = tmpl_robno.RobnoPrometRucIzlazneFakture(*h.tabs, *h.subtabs, tbl, magValues, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer)
	default:
		grupeValues, comboErr := h.service.GetRobneGrupeComboValues(ctx)
		if comboErr != nil {
			h.error(c, comboErr)
			return
		}
		renderErr = tmpl_robno.RobnoPrometRucLagerLista(*h.tabs, *h.subtabs, tbl, magValues, grupeValues, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer)
	}
	if renderErr != nil {
		h.error(c, renderErr)
	}
}

func (h *RobnoPrometHandler) RobnoPrometGradiliste(c *gin.Context) {
	ctx, _, ok := h.context(c, 4)
	if !ok {
		return
	}
	h.renderStandard(c, ctx, robnoPrometGradilisteTitle, robnoPrometGradilisteTableID, robnoPrometGradilisteURL, h.service.GetPrometGradilistaTableFields(), h.service.GetPrometGradilista, robnoPrometGradilisteURLStampa, hxValsRobnoPrometGrupeForTab("gradiliste"), []string{"odmagacina", "domagacina", "odsifre", "dosifre", "oddatuma", "dodatuma", "odgrupe", "dogrupe"}, func(tbl domain.TableData, magValues, grupeValues []domain.ComboItem, b, p domain.Button, s domain.InputControl) error {
		return tmpl_robno.RobnoPrometGradiliste(*h.tabs, tbl, magValues, grupeValues, b, p, s, i18n.GetInstance()).Render(ctx, c.Writer)
	})
}

func (h *RobnoPrometHandler) RobnoPrometGradilisteVpcNc(c *gin.Context) {
	ctx, _, ok := h.context(c, 5)
	if !ok {
		return
	}
	h.renderStandard(c, ctx, robnoPrometGradVpcTitle, robnoPrometGradVpcTableID, robnoPrometGradVpcURL, h.service.GetPrometGradilisteVpcNcTableFields(), h.service.GetPrometGradilisteVpcNc, robnoPrometGradVpcURLStampa, hxValsRobnoPrometGrupeForTab("gradiliste-vpc"), []string{"odmagacina", "domagacina", "odsifre", "dosifre", "oddatuma", "dodatuma", "odgrupe", "dogrupe"}, func(tbl domain.TableData, magValues, grupeValues []domain.ComboItem, b, p domain.Button, s domain.InputControl) error {
		return tmpl_robno.RobnoPrometGradilisteVpcNc(*h.tabs, tbl, magValues, grupeValues, b, p, s, i18n.GetInstance()).Render(ctx, c.Writer)
	})
}

type robnoPrometFetchFunc func(context.Context, *domain.TableData, bool, int, int, domain.PrometParam) error
type robnoPrometRenderFunc func(domain.TableData, []domain.ComboItem, []domain.ComboItem, domain.Button, domain.Button, domain.InputControl) error

func (h *RobnoPrometHandler) renderStandard(c *gin.Context, ctx context.Context, title, tableID, url string, fields []domain.Fields, fetch robnoPrometFetchFunc, printURL, vals string, required []string, render robnoPrometRenderFunc) {
	tbl := common.SetTableBasicData(title, tableID, fields, url, url, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, tableID, url, false, false, false)
	if common.IsDataRequest(c) {
		if !h.validate(c, required) || !h.getPaginatedReport(c, &tbl, fetch) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, grupeValues, err := h.comboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(url, tableID, vals)
	btnPrint := common.SetPrintButton(tableID+"-stampa", "Stampa", "stampa", printURL, "GET", true, common.ClassPrintButton, h.standardPrintFields())
	search := common.CreateSearchInput("search-input", translator, url, "#"+tableID, vals)
	if err := render(tbl, magValues, grupeValues, btnObrada, btnPrint, search); err != nil {
		h.error(c, err)
	}
}

func (h *RobnoPrometHandler) context(c *gin.Context, tab int) (context.Context, domain.UserSession, bool) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return nil, domain.UserSession{}, false
	}
	common.SetActiveTab(h.tabs, tab)
	return c.Request.Context(), *session, true
}

func (h *RobnoPrometHandler) validate(c *gin.Context, required []string) bool {
	if errors := common.ValidateRequiredParams(c, required); len(errors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, errors, common.ErrMsgValidation)
		return false
	}
	return true
}

func (h *RobnoPrometHandler) obradaButton(url, tableID, vals string) domain.Button {
	return common.SetButton(tableID+"-obrada", "Obradi", "fin_obrada", url, "#"+tableID, "innerHTML", "GET", "", vals, true, common.ClassSaveButton, "handleDialogResponse")
}

func (h *RobnoPrometHandler) comboValues(ctx context.Context) ([]domain.ComboItem, []domain.ComboItem, error) {
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		return nil, nil, err
	}
	grupeValues, err := h.service.GetRobneGrupeComboValues(ctx)
	return magValues, grupeValues, err
}

func (h *RobnoPrometHandler) params(c *gin.Context) domain.PrometParam {
	return domain.PrometParam{
		OdKonta: c.Query("odmagacina"), DoKonta: c.Query("domagacina"), OdSifre: c.Query("odsifre"), DoSifre: c.Query("dosifre"),
		OdDatuma: c.Query("oddatuma"), DoDatuma: c.Query("dodatuma"), StanjeNaDan: c.Query("stanjenadan"), TipCene: c.Query("tipcene"),
		UlazIzlaz: c.Query("ulazizlaz"), OdPodgrupe: c.Query("odpodgrupe"), DoPodgrupe: c.Query("dopodgrupe"), SearchText: c.Query("query"),
		StampajGrupaPodgrupa: c.Query("stampajgrupapodgrupa") == "true", ZaliheOdNule: c.Query("zaliheodnule") == "true", AzbucniRed: c.Query("azbucnired") == "true",
		ZbirMagacina: c.Query("zbirmagacina") == "true", StampajSamoZbir: c.Query("stampajsamoZbir") == "true", UkljuceneUsluge: c.Query("ukljuceneusluge") == "true",
	}
}

func (h *RobnoPrometHandler) getPaginatedReport(c *gin.Context, tbl *domain.TableData, fetch robnoPrometFetchFunc) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	params := h.params(c)
	for _, total := range []bool{true, false} {
		if err := fetch(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

func (h *RobnoPrometHandler) standardPrintFields() string {
	return "odmagacina,domagacina,odsifre,dosifre,odgrupe,dogrupe,oddatuma,dodatuma"
}
func (h *RobnoPrometHandler) printFields(subIndex int) string {
	switch subIndex {
	case 1:
		return "ulazizlaz,odmagacina,domagacina,oddatuma,dodatuma,zbirmagacina"
	case 2:
		return "odmagacina,domagacina,oddatuma,dodatuma,stampajsamoZbir"
	case 3:
		return "odmagacina,domagacina,oddatuma,dodatuma,stampajsamoZbir,ukljuceneusluge"
	default:
		return "tipcene,odmagacina,domagacina,stanjenadan,stampajgrupapodgrupa,zaliheodnule,azbucnired,odgrupe,dogrupe,odpodgrupe,dopodgrupe"
	}
}

func (h *RobnoPrometHandler) error(c *gin.Context, err error) {
	common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
}

func (h *RobnoPrometHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())
	r.GET(robnoPrometURLPrefix, h.RobnoPrometMain)
	r.GET(robnoPrometGrupe1URL, h.RobnoPrometArtikal)
	r.GET(robnoPrometKupciURL, h.RobnoPrometKupci)
	r.GET(robnoPrometDobavljaciURL, h.RobnoPrometDobavljaci)
	r.GET(robnoPrometRucURL, h.RobnoPrometRucMain)
	r.GET(robnoPrometRucUlazIzlazURL, h.RobnoPrometRucUlazIzlaz)
	r.GET(robnoPrometRucMagacinimaURL, h.RobnoPrometRucMagacinima)
	r.GET(robnoPrometRucIzlazneFaktureURL, h.RobnoPrometRucIzlazneFakture)
	r.GET(robnoPrometGradilisteURL, h.RobnoPrometGradiliste)
	r.GET(robnoPrometGradVpcURL, h.RobnoPrometGradilisteVpcNc)
	for _, printURL := range []string{
		robnoPrometGrupe1URLStampa,
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

func robnoPrometRucSubTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{{ID: "robnopromet-ruc-lager", Label: translator.Label("Lager lista"), HXRequestUrl: robnoPrometRucURL, IsActive: true, Name: "ruc-lager"}, {ID: "robnopromet-ruc-ulaz-izlaz", Label: translator.Label("Ulaz/izlaz za period"), HXRequestUrl: robnoPrometRucUlazIzlazURL, Name: "ruc-ulaz-izlaz"}, {ID: "robnopromet-ruc-magacinima", Label: translator.Label("RUC po magacinima"), HXRequestUrl: robnoPrometRucMagacinimaURL, Name: "ruc-magacinima"}, {ID: "robnopromet-ruc-izlazne-fakture", Label: translator.Label("Izlazne fakture"), HXRequestUrl: robnoPrometRucIzlazneFaktureURL, Name: "ruc-izlazne-fakture"}}}
}
func robnoPrometTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{{ID: "robnopromet-grupe1", Label: translator.Label("Promet po grupi artikala 1"), HXRequestUrl: robnoPrometGrupe1URL, IsActive: true, Name: "grupe1"}, {ID: "robnopromet-kupci", Label: translator.Label("Promet po kupcima"), HXRequestUrl: robnoPrometKupciURL, Name: "kupci"}, {ID: "robnopromet-dobavljaci", Label: translator.Label("Nabavka po dobavljačima"), HXRequestUrl: robnoPrometDobavljaciURL, Name: "dobavljaci"}, {ID: "robnopromet-ruc", Label: translator.Label("Lager lista ulaz/izlaz RUC"), HXRequestUrl: robnoPrometRucURL, Name: "ruc"}, {ID: "robnopromet-gradiliste", Label: translator.Label("Izveštaj zaduženja gradilišta"), HXRequestUrl: robnoPrometGradilisteURL, Name: "gradiliste"}, {ID: "robnopromet-gradiliste-vpc", Label: translator.Label("Izveštaj zaduženja gradilišta VPC-NC"), HXRequestUrl: robnoPrometGradVpcURL, Name: "gradiliste-vpc"}}}
}
