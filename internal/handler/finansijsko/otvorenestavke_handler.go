package finansijsko

import (
	"fmt"
	"helia/config"
	"time"

	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	tmpl_rep_fin "helia/frontend/templates/reports/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service/finansijsko"
	"helia/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	otvorenestavkeContentTitle             string = "OTVORENE STAVKE"
	zatvoreneStavkeContentTitle            string = "ZATVORENE STAVKE"
	otvorenestavkeTableID                  string = "otvorenestavketable"
	otvorenestavkeDetaljiTableID           string = "otvorenestavketabledetalji"
	zatvoreneStavkeTableID                 string = "zatvorenestavketable"
	zatvoreneStavkeDetaljiTableID          string = "zatvorenestavketabledetalji"
	otvoreneStavkeURL                      string = "/api/otvorenestavke"
	otvoreneStavkeURLPPartneri             string = "/api/otvorenestavke/partneri"
	otvoreneStavkeURLPartneriDetalji       string = "/api/otvorenestavke/partneridetails"
	otvoreneStavkeURLPartneriStampa        string = "/api/otvorenestavke/partneri/stampa"
	otvoreneStavkeURLPartneriStampaOpomene string = "/api/otvorenestavke/partneri/stampa/opomene"
	zatvoreneStavkeURLPartneri             string = "/api/otvorenestavke/zatvorene/partneri"
	zatvoreneStavkeURLPartneriStampa       string = "/api/otvorenestavke/zatvorene/partneri/stampa"
	zatvoreneStavkeURLPartneriDetalji      string = "/api/otvorenestavke/zatvorene/partneridetails"
	iosContentTitle                        string = "IOS"
	iosTableID                             string = "iostable"
	iosDetaljiTableID                      string = "iostabledetalji"
	iosURLPartneri                         string = "/api/otvorenestavke/ios/partneri"
	iosURLPartneriDetalji                  string = "/api/otvorenestavke/ios/partneridetails"
	iosURLPartneriStampa                   string = "/api/otvorenestavke/ios/partneri/stampa"
	dospelaContentTitle                    string = "DOSPELA POTRAŽIVANJA/DUGOVANJA"
	dospelaDetaljiTableID                  string = "dospeladetalji"
	dospelaTableID                         string = "dospelatable"
	dospelaURLPartneri                     string = "/api/otvorenestavke/dospela/partneri"
	dospelaURLPartneriDetalji              string = "/api/otvorenestavke/dospela/partneridetails"
	dospelaStampaAnalitickiURL             string = "/api/otvorenestavke/dospela/stampa/analiticki"
	dugovanjaContentTitle                  string = "PREGLED POTRAZIVANJA/OBAVEZE"
	dugovanjaTableID                       string = "dugovanjatable"
	dugovanjaURL                           string = "/api/otvorenestavke/pregleddugovanja"
	dugovanjaURLStampa                     string = "/api/otvorenestavke/pregleddugovanja/stampa"
	dospelaStarostiContentTitle            string = "PREGLED DOSPELOG DUGA PO STAROSTI"
	dospelaStarostiTableID                 string = "dospelastarostitable"
	dospelaStarostiURL                     string = "/api/otvorenestavke/dospelidugpostarosti"
	dospelaStarostiStampaURL               string = "/api/otvorenestavke/dospelidugpostarosti/stampa"
	povezivanjePartneriTableID             string = "pov-partneri-table"
	povezivanjeUplateTableID               string = "pov-uplate-table"
	povezivanjeFaktureTableID              string = "pov-fakture-table"

	hxValsOtvoreneStavke string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"poddatumom": document.getElementById("poddatumom")?.value,
			"otvstavkedana": document.getElementById("otvstavkedana")?.value,
            "query": document.getElementById("search-input")?.value
        }`
	hxValsOtvoreneStavkeDetalji string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"poddatumom": document.getElementById("poddatumom")?.value,
			"otvstavkedana": document.getElementById("otvstavkedana")?.value,
            "query": document.getElementById("search-input-detalji")?.value
        }`
	hxValsZatvoreneStavke string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
            "query": document.getElementById("search-input")?.value
        }`
	hxValsZatvoreneStavkeDetalji string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
            "query": document.getElementById("search-input-detalji")?.value
        }`
	hxValsIOS string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"poddatumom": document.getElementById("poddatumom")?.value,
			"otvstavkedana": document.getElementById("otvstavkedana")?.value,
            "query": document.getElementById("search-input")?.value
        }`
	hxValsIOSDetalji string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"poddatumom": document.getElementById("poddatumom")?.value,
			"otvstavkedana": document.getElementById("otvstavkedana")?.value,
            "query": document.getElementById("search-input-detalji")?.value
        }`
	hxValsDospela string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"poddatumom": document.getElementById("poddatumom")?.value,
			"brojdana": document.getElementById("brojdana")?.value,
			"tip_pregleda": document.querySelector('input[name="tip_pregleda"]:checked')?.value,
			"tip_potrazivanja": document.querySelector('input[name="tip_potrazivanja"]:checked')?.value,
			"query": document.getElementById("search-input")?.value
			}`
	hxValsDospelaDetalji string = `js:{
			"konto": document.getElementById("konto")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"poddatumom": document.getElementById("poddatumom")?.value,
			"brojdana": document.getElementById("brojdana")?.value,
			"tip_pregleda": document.querySelector('input[name="tip_pregleda"]:checked')?.value,
			"tip_potrazivanja": document.querySelector('input[name="tip_potrazivanja"]:checked')?.value,
            "query": document.getElementById("search-input-detalji")?.value
        }`
	hxValsDugovanja string = `js:{
			"odkonta": document.getElementById("odkonta")?.value,
			"dokonta": document.getElementById("dokonta")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"stanjenadan": document.getElementById("stanjenadan")?.value,
			"tip_pregleda": document.querySelector('input[name="tip_pregleda"]:checked')?.value,
			"dospece15": document.getElementById("dospece15")?.value,
			"dospece30": document.getElementById("dospece30")?.value,
			"dospece60": document.getElementById("dospece60")?.value,
			"dospece90": document.getElementById("dospece90")?.value,	
			"dospece120": document.getElementById("dospece120")?.value,
			"query": document.getElementById("search-input")?.value
		}`

	hxValsPovezivanje = `js:{
		"konto": document.getElementById("konto")?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"query": document.getElementById("search-partneri")?.value
		}`
)

// OtvoreneStavkeHandler handles all requests for Otvorene Stavke module
type OtvoreneStavkeHandler struct {
	tabData domain.TabData
	service finansijsko.OtvoreneStavkeService
	cfg     config.Config
}

// NewOtvoreneStavkeHandler creates a new handler instance
func NewOtvoreneStavkeHandler(
	service finansijsko.OtvoreneStavkeService,
	cfg config.Config,
) *OtvoreneStavkeHandler {
	handler := &OtvoreneStavkeHandler{
		service: service,
		cfg:     cfg,
	}
	handler.tabData = GetOtvoreneStavkeTabData()
	return handler
}

// OtvoreneStavkeMain - Main entry point for Otvorene Stavke
func (h *OtvoreneStavkeHandler) OtvoreneStavkeMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	translator := i18n.GetInstance()
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	h.setOtvoreneStavkeActiveTab("otvorenestavke")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), otvoreneStavkeURLPPartneri, fmt.Sprintf("#%s", otvorenestavkeTableID), hxValsOtvoreneStavke)
	searchInputDetalji := common.CreateSearchInput("search-input-detalji", i18n.GetInstance(), otvoreneStavkeURLPartneriDetalji, fmt.Sprintf("#%s", otvorenestavkeDetaljiTableID), hxValsOtvoreneStavkeDetalji)

	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", otvoreneStavkeURLPPartneri, fmt.Sprintf("#%s", otvorenestavkeTableID), "innerHTML", "GET", "", hxValsOtvoreneStavke, true, common.ClassSaveButton, "handleBackendResponse")
	btnPrint := common.SetPrintButton("btn-print-otvorene", translator.Button("Štampa"), "fin_stampa", otvoreneStavkeURLPartneriStampa, "GET", true, common.ClassPrintButton, "konto,odsifre,dosifre,poddatumom,otvstavkedana,otvorene-selected-id,stampaotvstavke")
	btnOpomene := common.SetPrintButton("btn-print-opomene", translator.Button("Štampa Opomene"), "fin_stampa_opomene", otvoreneStavkeURLPartneriStampa, "GET", true, common.ClassStampaOpomenaButton, "konto,odsifre,dosifre,poddatumom,otvstavkedana,otvorene-selected-id,stampaopomena")

	tblPartneri := common.SetTableBasicData(otvorenestavkeContentTitle, otvorenestavkeTableID, h.service.GetPartneriFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblPartneri, otvorenestavkeContentTitle, "", false, false, false)
	tblDetalji := common.SetTableBasicData("Otvorene stavke - detalji", otvorenestavkeDetaljiTableID, h.service.GetOtvoreneStavkeDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "OTVORENE STAVKE - DETALJI", "", false, false, false)
	err := tmpl_fin.OtvoreneStavkeMain(h.tabData, tblPartneri, tblDetalji, btnObrada, btnPrint, btnOpomene, searchInput, searchInputDetalji, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

// OtvoreneStavke - Tab 1: Otvorene stavke (Open Items)
func (h *OtvoreneStavkeHandler) OtvoreneStavke(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}

	gnGod := userSession.SelectedGod
	tblPartneri := common.SetTableBasicData(otvorenestavkeContentTitle, otvorenestavkeTableID, h.service.GetPartneriFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblPartneri, otvorenestavkeContentTitle, "", false, false, false)
	tblDetalji := common.SetTableBasicData("Otvorene stavke - detalji", otvorenestavkeDetaljiTableID, h.service.GetOtvoreneStavkeDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "OTVORENE STAVKE - DETALJI", "", false, false, false)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), otvoreneStavkeURLPPartneri, fmt.Sprintf("#%s", otvorenestavkeTableID), hxValsOtvoreneStavke)
	searchInputDetalji := common.CreateSearchInput("search-input-detalji", i18n.GetInstance(), otvoreneStavkeURLPartneriDetalji, fmt.Sprintf("#%s", otvorenestavkeDetaljiTableID), hxValsOtvoreneStavkeDetalji)
	tblPartneri.Pagination.HxVals = hxValsOtvoreneStavke
	tblPartneri.HasTotals = true
	tblDetalji.Pagination.HxVals = hxValsOtvoreneStavkeDetalji
	tblPartneri.URLGetAll = otvoreneStavkeURLPPartneri
	tblPartneri.URLPrefix = otvoreneStavkeURLPPartneri
	tblDetalji.URLGetAll = otvoreneStavkeURLPartneriDetalji
	tblDetalji.URLPrefix = otvoreneStavkeURLPartneriDetalji
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", otvoreneStavkeURLPPartneri, fmt.Sprintf("#%s", otvorenestavkeTableID), "innerHTML", "GET", "", hxValsOtvoreneStavke, true, common.ClassSaveButton, "handleBackendResponse")

	btnPrint := common.SetPrintButton("btn-print-otvorene", "Štampa", "fin_stampa", otvoreneStavkeURLPartneriStampa, "GET", true, common.ClassPrintButton, "konto,odsifre,dosifre,poddatumom,otvstavkedana,otvorene-selected-id,stampaotvstavke")
	btnOpomene := common.SetPrintButton("btn-print-opomene", "Štampa Opomene", "fin_stampa_opomene", otvoreneStavkeURLPartneriStampa, "GET", true, common.ClassStampaOpomenaButton, "konto,odsifre,dosifre,poddatumom,otvstavkedana,otvorene-selected-id,stampaopomena")
	h.setOtvoreneStavkeActiveTab("otvorenestavke")

	tblPartneri.DetailTarget = fmt.Sprintf("#%s", otvorenestavkeDetaljiTableID)
	tblPartneri.DetailURL = "/api/otvorenestavke/partneridetails"
	tblPartneri.DetailHxRequestType = "GET"
	tblPartneri.DetailHxSwap = "innerHTML"
	tblPartneri.DetailHxTrigger = "click, change delay:500ms"

	if requestSource == "menu" || requestSource == "tab" {
		tblPartneri.ShowPagination = true
		err := tmpl_fin.OtvoreneStavke(h.tabData, tblPartneri, tblDetalji, btnObrada, btnPrint, btnOpomene, searchInput, searchInputDetalji, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.OtvStavkeParam{}
		params.SearchText = c.Query("query")
		params.Konto = c.Query("konto")
		params.OdSifre = c.Query("odsifre")
		params.DoSifre = c.Query("dosifre")
		params.PodDatumom = c.Query("poddatumom")
		params.OtvStavkeDana = c.Query("otvstavkedana")
		fieldsError := common.ValidateRequiredParams(c, []string{"konto", "odsifre", "dosifre", "poddatumom", "otvstavkedana"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetOtvoreneStavkePartneri(ctx, &tblPartneri, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		err = h.service.GetOtvoreneStavkePartneri(ctx, &tblPartneri, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tblPartneri)
		return
	}
}
func (h *OtvoreneStavkeHandler) OtvoreneStavkeDetalji(c *gin.Context) {
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgGetIDFromURL)
		return
	}
	ctx := c.Request.Context()
	searchText := c.Query("query")
	tblDetalji := common.SetTableBasicData("Otvorene stavke - detalji", otvorenestavkeDetaljiTableID, h.service.GetOtvoreneStavkeDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "OTVORENE STAVKE - DETALJI", "", false, false, false)

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	repParams := domain.ReportParameters{}
	err = h.service.GetOtvoreneStavkeDetalji(ctx, id, &tblDetalji, true, pageSize, page, searchText, "", "", &repParams)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	err = h.service.GetOtvoreneStavkeDetalji(ctx, id, &tblDetalji, false, pageSize, page, searchText, "", "", &repParams)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tblDetalji.Pagination.HxVals = hxValsOtvoreneStavkeDetalji
	tblDetalji.URLGetAll = otvoreneStavkeURLPartneriDetalji
	tblDetalji.URLPrefix = otvoreneStavkeURLPartneriDetalji
	tmpl.Table(tblDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

// PregledOtvorenihStavkiStampa renders a printable open items overview report for all partners.
func (h *OtvoreneStavkeHandler) PregledOtvorenihStavkiStampa(c *gin.Context) {
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	tipStampeOtvStavke := c.Query("stampaotvstavke")
	tipStampeOpomena := c.Query("stampaopomena")
	id, err := utils.GetInt64FromQueryRequest(c, "otvorene-selected-id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid ID")
		return
	}
	tipObrade := ""
	if tipStampeOtvStavke == "otvstavke" {
		tipObrade = "otvstavke"
	}
	if tipStampeOpomena == "opomena" {
		tipObrade = "opomena"
	}
	podDatumom := c.Query("poddatumom")

	// Format dates for display (input: 2006-01-02, output: DD.MM.YYYY)
	podDatumomFmt := podDatumom
	if t, err := time.Parse("2006-01-02", podDatumom); err == nil {
		podDatumomFmt = t.Format(common.DateLayout)
	}
	ctx := c.Request.Context()
	tbl := common.SetTableBasicData("Otvorene stavke - detalji", "otvorene-stavke-detalji-table", h.service.GetOpomenaStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "OTVORENE STAVKE - DETALJI", "", false, false, false)
	fvrData, err := h.service.GetFvrData(ctx)
	repParams := domain.ReportParameters{
		ReportName:  translator.Title("PREGLED OTVORENIH STAVKI"),
		Orientation: "portrait",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		SifDel:      fvrData.SifDel,
		ParameterItems: map[string]domain.ParameterItem{
			"Naziv":       {Name: translator.Label("Naziv"), Value: "KUPAC"},
			"Adresa":      {Name: translator.Label("Adresa"), Value: "Adresa"},
			"Postcode":    {Name: translator.Label("Postcode"), Value: "Postbroj"},
			"Mesto":       {Name: translator.Label("Mesto"), Value: "Mesto"},
			"ObvPDV":      {Name: translator.Label("Obveznik PDV Br."), Value: "Obveznik PDV"},
			"PIB":         {Name: translator.Label("PIB"), Value: "PIB"},
			"Telefon":     {Name: translator.Label("Telefon"), Value: "Telefon"},
			"Konto":       {Name: translator.Label("Konto"), Value: "Konto"},
			"Sifra":       {Name: translator.Label("Sifra"), Value: "Šifra"},
			"Stanjenadan": {Name: translator.Label("Stanjenadan"), Value: podDatumomFmt},
		},
	}
	tbl.HasTotals = true
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	err = h.service.GetOtvoreneStavkeDetalji(ctx, id, &tbl, false, 0, 0, "", common.TipStampePrint, tipObrade, &repParams)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tipStampe := "otvstavke"
	if tipStampeOtvStavke == "otvstavke" {
		repParams.ReportName = translator.Title("PREGLED OTVORENIH STAVKI")
		tmpl_rep_fin.OtvoreneStavkePregledStampa(repParams, tbl, tipStampe, translator).Render(ctx, c.Writer)
	}
	if tipStampeOpomena == "opomena" {
		tipStampe = "opomena"
		repParams.ReportName = translator.Title("OPOMENA ZA NEPLAĆENE RAČUNE")
		repParams.ParameterItems["otvstavkedana"] = domain.ParameterItem{Name: translator.Label("Otvorene stavke dana"), Value: c.Query("otvstavkedana")}
		tmpl_rep_fin.OtvoreneStavkePregledStampa(repParams, tbl, tipStampe, translator).Render(ctx, c.Writer)
	}
}

// ZatvoreneStavke - Tab 2: Zatvorene stavke (Closed Items)
func (h *OtvoreneStavkeHandler) ZatvoreneStavke(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	gnGod := userSession.SelectedGod
	tblPartneri := common.SetTableBasicData(zatvoreneStavkeContentTitle, zatvoreneStavkeTableID, h.service.GetPartneriFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblPartneri, zatvoreneStavkeContentTitle, "", false, false, false)
	tblDetalji := common.SetTableBasicData("Zatvorene stavke - detalji", zatvoreneStavkeDetaljiTableID, h.service.GetZatvoreneStavkeDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "ZATVORENE STAVKE - DETALJI", "", false, false, false)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), zatvoreneStavkeURLPartneri, fmt.Sprintf("#%s", zatvoreneStavkeTableID), hxValsZatvoreneStavke)
	searchInputDetalji := common.CreateSearchInput("search-input-detalji", i18n.GetInstance(), zatvoreneStavkeURLPartneriDetalji, fmt.Sprintf("#%s", zatvoreneStavkeDetaljiTableID), hxValsZatvoreneStavkeDetalji)
	tblPartneri.Pagination.HxVals = hxValsZatvoreneStavke
	tblDetalji.Pagination.HxVals = hxValsZatvoreneStavkeDetalji
	tblPartneri.URLGetAll = zatvoreneStavkeURLPartneri
	tblPartneri.URLPrefix = zatvoreneStavkeURLPartneri
	tblDetalji.URLGetAll = zatvoreneStavkeURLPartneriDetalji
	tblDetalji.URLPrefix = zatvoreneStavkeURLPartneriDetalji
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", zatvoreneStavkeURLPartneri, fmt.Sprintf("#%s", zatvoreneStavkeTableID), "innerHTML", "GET", "", hxValsZatvoreneStavke, true, common.ClassSaveButton, "handleBackendResponse")
	btnPrint := common.SetPrintButton("btn-print-zatvorene", translator.Button("Štampa"), "fin_print", zatvoreneStavkeURLPartneriStampa, "GET", true, common.ClassPrintButton, "konto,odsifre,dosifre,oddatuma,dodatuma,zatvorene-selected-id")
	h.setOtvoreneStavkeActiveTab("zatvorenestavke")
	tblPartneri.DetailTarget = fmt.Sprintf("#%s", zatvoreneStavkeDetaljiTableID)
	tblPartneri.DetailURL = "/api/otvorenestavke/zatvorene/partneridetails"
	tblPartneri.DetailHxRequestType = "GET"
	tblPartneri.DetailHxSwap = "innerHTML"
	tblPartneri.DetailHxTrigger = "click, change delay:500ms"

	if requestSource == "menu" || requestSource == "tab" {
		tblPartneri.ShowPagination = true
		err := tmpl_fin.ZatvoreneStavke(h.tabData, tblPartneri, tblDetalji, btnObrada, btnPrint, searchInput, searchInputDetalji, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.OtvStavkeParam{}
		params.SearchText = c.Query("query")
		params.Konto = c.Query("konto")
		params.OdSifre = c.Query("odsifre")
		params.DoSifre = c.Query("dosifre")
		params.OdDatuma = c.Query("oddatuma")
		params.DoDatuma = c.Query("dodatuma")
		fieldError := common.ValidateRequiredParams(c, []string{"konto", "odsifre", "dosifre", "oddatuma", "dodatuma"})
		if len(fieldError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetZatvoreneStavkePartneri(ctx, &tblPartneri, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		err = h.service.GetZatvoreneStavkePartneri(ctx, &tblPartneri, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tblPartneri)
		return
	}
}

// ZatvoreneStavkeDetalji - Tab 2 details: Zatvorene stavke (Closed Items)
func (h *OtvoreneStavkeHandler) ZatvoreneStavkeDetalji(c *gin.Context) {
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgGetIDFromURL)
		return
	}
	ctx := c.Request.Context()
	searchText := c.Query("query")
	tblDetalji := common.SetTableBasicData("Zatvorene stavke - detalji", zatvoreneStavkeDetaljiTableID, h.service.GetZatvoreneStavkeDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "ZATVORENE STAVKE - DETALJI", "", false, false, false)

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	err = h.service.GetZatvoreneStavkeDetalji(ctx, id, &tblDetalji, true, pageSize, page, searchText, "", nil)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	err = h.service.GetZatvoreneStavkeDetalji(ctx, id, &tblDetalji, false, pageSize, page, searchText, "", nil)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	tblDetalji.Pagination.HxVals = hxValsZatvoreneStavkeDetalji
	tblDetalji.URLGetAll = zatvoreneStavkeURLPartneriDetalji
	tblDetalji.URLPrefix = zatvoreneStavkeURLPartneriDetalji
	tmpl.Table(tblDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

// PregledZatvorenihStavkiStampa renders a printable closed items overview report grouped by partner.
func (h *OtvoreneStavkeHandler) PregledZatvorenihStavkiStampa(c *gin.Context) {
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	id, err := utils.GetInt64FromQueryRequest(c, "zatvorene-selected-id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid ID")
		return
	}
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")

	// Format dates for display (input: 2006-01-02, output: DD.MM.YYYY)
	odDatumaFmt := odDatuma
	doDatumaFmt := doDatuma
	if t, err := time.Parse("2006-01-02", odDatuma); err == nil {
		odDatumaFmt = t.Format(common.DateLayout)
	}
	if t, err := time.Parse("2006-01-02", doDatuma); err == nil {
		doDatumaFmt = t.Format(common.DateLayout)
	}
	ctx := c.Request.Context()
	tbl := common.SetTableBasicData("Zatvorene stavke - detalji", "zatvorene-stavke-detalji-table", h.service.GetZatvoreneStavkeStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "ZATVORENE STAVKE - DETALJI", "", false, false, false)
	fvrData, err := h.service.GetFvrData(ctx)
	repParams := domain.ReportParameters{
		ReportName:  translator.Title("PREGLED ZATVORENIH STAVKI"),
		Orientation: "portrait",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		SifDel:      fvrData.SifDel,
		ParameterItems: map[string]domain.ParameterItem{
			"Konto":    {Name: translator.Label("Konto"), Value: c.Query("konto")},
			"empty":    {Name: "", Value: ""},
			"Od Šifre": {Name: translator.Label("Od Šifre"), Value: c.Query("odsifre")},
			"Do Šifre": {Name: translator.Label("Do Šifre"), Value: c.Query("dosifre")},
			"OdDatuma": {Name: translator.Label("Od Datuma"), Value: odDatumaFmt},
			"DoDatuma": {Name: translator.Label("Do Datuma"), Value: doDatumaFmt},
			"Partner":  {Name: translator.Label("Partner"), Value: ""},
		},
	}

	tbl.HasTotals = true
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	err = h.service.GetZatvoreneStavkeDetalji(ctx, id, &tbl, false, 0, 0, "", common.TipStampePrint, &repParams)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.ZatvoreneStavkeStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

// IOS - Tab 3: IOS (Izvod otvorenih stavki)
func (h *OtvoreneStavkeHandler) IOS(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}

	h.setOtvoreneStavkeActiveTab("ios")
	gnGod := userSession.SelectedGod
	tblPartneri := common.SetTableBasicData(iosContentTitle, iosTableID, h.service.GetPartneriFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblPartneri, iosContentTitle, "", false, false, false)
	tblDetalji := common.SetTableBasicData("IOS - detalji", iosDetaljiTableID, h.service.GetOtvoreneStavkeDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "IOS - DETALJI", "", false, false, false)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), iosURLPartneri, fmt.Sprintf("#%s", iosTableID), hxValsIOS)
	searchInputDetalji := common.CreateSearchInput("search-input-detalji", i18n.GetInstance(), iosURLPartneriDetalji, fmt.Sprintf("#%s", iosDetaljiTableID), hxValsIOSDetalji)
	tblPartneri.Pagination.HxVals = hxValsIOS
	tblDetalji.Pagination.HxVals = hxValsIOSDetalji
	tblPartneri.URLGetAll = iosURLPartneri
	tblPartneri.URLPrefix = iosURLPartneri
	tblDetalji.URLGetAll = iosURLPartneriDetalji
	tblDetalji.URLPrefix = iosURLPartneriDetalji
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", iosURLPartneri, fmt.Sprintf("#%s", iosTableID), "innerHTML", "GET", "", hxValsIOS, true, common.ClassSaveButton, "handleBackendResponse")
	btnPrint := common.SetPrintButton("btn-print-ios", translator.Button("Štampa"), "fin_stampa", iosURLPartneriStampa, "GET", true, common.ClassPrintButton, "konto,odsifre,dosifre,poddatumom,otvstavkedana,ios-selected-id")

	tblPartneri.DetailTarget = fmt.Sprintf("#%s", iosDetaljiTableID)
	tblPartneri.DetailURL = "/api/otvorenestavke/ios/partneridetails"
	tblPartneri.DetailHxRequestType = "GET"
	tblPartneri.DetailHxSwap = "innerHTML"
	tblPartneri.DetailHxTrigger = "click, change delay:500ms"

	if requestSource == "menu" || requestSource == "tab" {
		tblPartneri.ShowPagination = true
		err := tmpl_fin.IOS(h.tabData, tblPartneri, tblDetalji, btnObrada, btnPrint, searchInput, searchInputDetalji, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.OtvStavkeParam{}
		params.SearchText = c.Query("query")
		params.Konto = c.Query("konto")
		params.OdSifre = c.Query("odsifre")
		params.DoSifre = c.Query("dosifre")
		params.PodDatumom = c.Query("poddatumom")
		params.OtvStavkeDana = c.Query("otvstavkedana")
		fieldsError := common.ValidateRequiredParams(c, []string{"konto", "odsifre", "dosifre", "poddatumom", "otvstavkedana"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetIOSPartneri(ctx, &tblPartneri, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		err = h.service.GetIOSPartneri(ctx, &tblPartneri, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tblPartneri)
		return
	}
}

// IOSDetalji - Renders the details of open items for a selected partner in the IOS tab.
func (h *OtvoreneStavkeHandler) IOSDetalji(c *gin.Context) {
	tblDetalji := common.SetTableBasicData("IOS - detalji", iosDetaljiTableID, h.service.GetIOSDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "IOS - DETALJI", "", false, false, false)

	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgGetIDFromURL)
		return
	}
	searchText := c.Query("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	err = h.service.GetIOSDetalji(ctx, id, &tblDetalji, true, pageSize, page, searchText, "", nil)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	err = h.service.GetIOSDetalji(ctx, id, &tblDetalji, false, pageSize, page, searchText, "", nil)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tblDetalji.Pagination.HxVals = hxValsIOSDetalji
	tblDetalji.URLGetAll = iosURLPartneriDetalji
	tblDetalji.URLPrefix = iosURLPartneriDetalji
	tmpl.Table(tblDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

// IOSDetaljiStampa renders a printable IOS (Izvod otvorenih stavki) report for a selected partner.
func (h *OtvoreneStavkeHandler) IOSDetaljiStampa(c *gin.Context) {
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	id, err := utils.GetInt64FromQueryRequest(c, "ios-selected-id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "Invalid ID")
		return
	}
	podDatumom := c.Query("poddatumom")
	podDatumomFmt := podDatumom
	if t, err := time.Parse("2006-01-02", podDatumom); err == nil {
		podDatumomFmt = t.Format(common.DateLayout)
	}
	ctx := c.Request.Context()
	tbl := common.SetTableBasicData("IOS - detalji", "ios-detalji-table", h.service.GetIOSStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "IOS - DETALJI", "", false, false, false)
	fvrData, err := h.service.GetFvrData(ctx)
	repParams := domain.ReportParameters{
		ReportName:  translator.Title("IZVOD OTVORENIH STAVKI"),
		Orientation: "portrait",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		SifDel:      fvrData.SifDel,
		ParameterItems: map[string]domain.ParameterItem{
			"Naziv":       {Name: translator.Label("Naziv"), Value: ""},
			"Adresa":      {Name: translator.Label("Adresa"), Value: ""},
			"Postcode":    {Name: translator.Label("Postcode"), Value: ""},
			"Mesto":       {Name: translator.Label("Mesto"), Value: ""},
			"ObvPDV":      {Name: translator.Label("Obveznik PDV Br."), Value: ""},
			"PIB":         {Name: translator.Label("PIB"), Value: ""},
			"Telefon":     {Name: translator.Label("Telefon"), Value: ""},
			"Konto":       {Name: translator.Label("Konto"), Value: c.Query("konto")},
			"Sifra":       {Name: translator.Label("Sifra"), Value: ""},
			"Stanjenadan": {Name: translator.Label("Stanjenadan"), Value: podDatumomFmt},
		},
	}
	tbl.HasTotals = true
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	err = h.service.GetIOSDetalji(ctx, id, &tbl, false, 0, 0, "", common.TipStampePrint, &repParams)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.IOSStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

// DospelaPotraživanja - Tab 4: Dospela potraživanja/dugovanja
func (h *OtvoreneStavkeHandler) DospelaPotrazivanja(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	partneriFields := h.service.GetPartneriFields()
	tipPregleda := c.Query("tip_pregleda")
	if tipPregleda == "S" {
		partneriFields = h.service.GetPartneriFieldsSintetika()
	}
	h.setOtvoreneStavkeActiveTab("dospelapotrazivanja")
	gnGod := userSession.SelectedGod
	tblPartneri := common.SetTableBasicData(dospelaContentTitle, iosTableID, partneriFields, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblPartneri, dospelaContentTitle, "", false, false, false)
	tblDetalji := common.SetTableBasicData("Dospela potraživanja - detalji", dospelaDetaljiTableID, h.service.GetDospelaDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "DOSPELA POTRAŽIVANJA - DETALJI", "", false, false, false)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), dospelaURLPartneri, fmt.Sprintf("#%s", dospelaTableID), hxValsDospela)
	searchInputDetalji := common.CreateSearchInput("search-input-detalji", i18n.GetInstance(), dospelaURLPartneriDetalji, fmt.Sprintf("#%s", dospelaDetaljiTableID), hxValsDospelaDetalji)
	tblPartneri.Pagination.HxVals = hxValsDospela
	tblDetalji.Pagination.HxVals = hxValsDospelaDetalji
	tblPartneri.URLGetAll = dospelaURLPartneri
	tblPartneri.URLPrefix = dospelaURLPartneri
	tblDetalji.URLGetAll = dospelaURLPartneriDetalji
	tblDetalji.URLPrefix = dospelaURLPartneriDetalji
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", dospelaURLPartneri, fmt.Sprintf("#%s", dospelaTableID), "innerHTML", "GET", "", hxValsDospela, true, common.ClassSaveButton, "handleBackendResponse")
	btnPrint := common.SetPrintButton("btn-print-dospela", translator.Button("Štampa"), "fin_stampa", dospelaStampaAnalitickiURL, "GET", true, common.ClassPrintButton, "konto,odsifre,dosifre,poddatumom,brojdana,tip_pregleda,tip_potrazivanja")

	tblPartneri.DetailTarget = fmt.Sprintf("#%s", dospelaDetaljiTableID)
	tblPartneri.DetailURL = "/api/otvorenestavke/dospela/partneridetails"
	tblPartneri.DetailHxRequestType = "GET"
	tblPartneri.DetailHxSwap = "innerHTML"
	tblPartneri.DetailHxTrigger = "click, change delay:500ms"

	if requestSource == "menu" || requestSource == "tab" {
		tblPartneri.ShowPagination = true
		err := tmpl_fin.DospelaPotrazivanja(h.tabData, tblPartneri, tblDetalji, btnObrada, btnPrint, searchInput, searchInputDetalji, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.OtvStavkeParam{}
		params.SearchText = c.Query("query")
		params.Konto = c.Query("konto")
		params.OdSifre = c.Query("odsifre")
		params.DoSifre = c.Query("dosifre")
		params.PodDatumom = c.Query("poddatumom")
		params.BrojDana = c.Query("brojdana")
		params.TipPregleda = c.Query("tip_pregleda")
		params.TipPotrazivanja = c.Query("tip_potrazivanja")

		fieldsError := common.ValidateRequiredParams(c, []string{"konto", "odsifre", "dosifre", "poddatumom", "brojdana", "tip_pregleda", "tip_potrazivanja"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetDospelaPotrazivanjaPartneri(ctx, &tblPartneri, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		h.service.GetDospelaPotrazivanjaPartneri(ctx, &tblPartneri, false, pageSize, page, params)

		utils.RenderContent(c, tblPartneri)
	}
}

// DospelaPotrazivanjaDetalji renders the details of due receivables/liabilities for a selected partner.
func (h *OtvoreneStavkeHandler) DospelaPotrazivanjaDetalji(c *gin.Context) {

	tblDetalji := common.SetTableBasicData("Dospela potraživanja - detalji", dospelaDetaljiTableID, h.service.GetDospelaDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "DOSPELA POTRAŽIVANJA - DETALJI", "", false, false, false)

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	ctx := c.Request.Context()
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgGetIDFromURL)
		return
	}
	searchText := c.Query("query")
	err = h.service.GetDospelaPotrazivanjaDetalji(ctx, id, &tblDetalji, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	err = h.service.GetDospelaPotrazivanjaDetalji(ctx, id, &tblDetalji, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tblDetalji.Pagination.HxVals = hxValsDospelaDetalji
	tblDetalji.URLGetAll = dospelaURLPartneriDetalji
	tblDetalji.URLPrefix = dospelaURLPartneriDetalji
	tmpl.Table(tblDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

// DugovanjaObavezeStampa renders the analytical print report for dospela potrazivanja grouped by partner.
func (h *OtvoreneStavkeHandler) DugovanjaObavezeStampa(c *gin.Context) {
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	ctx := c.Request.Context()
	params := domain.OtvStavkeParam{
		Konto:           c.Query("konto"),
		OdSifre:         c.Query("odsifre"),
		DoSifre:         c.Query("dosifre"),
		PodDatumom:      c.Query("poddatumom"),
		BrojDana:        c.Query("brojdana"),
		TipPregleda:     c.Query("tip_pregleda"),
		TipPotrazivanja: c.Query("tip_potrazivanja"),
	}
	podDatumomFmt := params.PodDatumom
	if t, err := time.Parse("2006-01-02", params.PodDatumom); err == nil {
		podDatumomFmt = t.Format(common.DateLayout)
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	brojDanaLabel := translator.Label("Potrazivanja koja dospevaju u narednih")
	repParams := domain.ReportParameters{
		ReportName:  translator.Title("PREGLED DOSPELIH POTRAZIVANJA"),
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		SifDel:      fvrData.SifDel,
		ParameterItems: map[string]domain.ParameterItem{
			"Stanjenadan": {Name: translator.Label("Stanje na dan"), Value: podDatumomFmt},
			"Brojdana":    {Name: brojDanaLabel, Value: params.BrojDana},
			"TipPregleda": {Name: "", Value: params.TipPregleda},
		},
	}
	partners, err := h.service.GetDospelaPotrazivanjaStampa(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.DugovanjaObavezeStampa(repParams, partners, translator).Render(ctx, c.Writer)
}

// PregledPotrazivanjaObaveze - Tab 5: Pregled potraživanja po obavezama
func (h *OtvoreneStavkeHandler) PregledPotrazivanjaObaveze(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	h.setOtvoreneStavkeActiveTab("pregleddugovanja")
	gnGod := userSession.SelectedGod
	headerFileds := h.service.GetDugovanjaFields()

	tbl := common.SetTableBasicData(dospelaContentTitle, dospelaTableID, headerFileds, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, dospelaContentTitle, "", false, false, false)
	tblDetalji := common.SetTableBasicData("Dospela potraživanja - detalji", dospelaDetaljiTableID, h.service.GetDospelaDetaljiFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "PREGLED POTRAŽIVANJA/OBAVEZE", "", false, false, false)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), dugovanjaURL, fmt.Sprintf("#%s", dugovanjaTableID), hxValsDugovanja)
	tbl.Pagination.HxVals = hxValsDugovanja
	tbl.URLGetAll = dugovanjaURL
	tbl.URLPrefix = dugovanjaURL
	tbl.HasTotals = true
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", dugovanjaURL, fmt.Sprintf("#%s", dugovanjaTableID), "innerHTML", "GET", "", hxValsDugovanja, true, common.ClassSaveButton, "handleBackendResponse")
	btnPrint := common.SetPrintButton("btn-print-dugovanja", translator.Button("Štampa"), "fin_stampa", dugovanjaURLStampa, "GET", true, common.ClassPrintButton, "odkonta,dokonta,odsifre,dosifre,stanjenadan,dospece15,dospece30,dospece60,dospece90,dospece120,tip_pregleda,stampaj_samo_zbir")

	if requestSource == "menu" || requestSource == "tab" {
		tbl.ShowPagination = true
		err := tmpl_fin.PregledPotrazivanjaObaveze(h.tabData, tbl, btnObrada, btnPrint, searchInput, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.OtvStavkeParam{}
		params.SearchText = c.Query("query")
		params.OdKonta = c.Query("odkonta")
		params.DoKonta = c.Query("dokonta")
		params.OdSifre = c.Query("odsifre")
		params.DoSifre = c.Query("dosifre")
		params.StanjeNaDan = c.Query("stanjenadan")
		params.TipPregleda = c.Query("tip_pregleda")
		params.Dospece15 = c.Query("dospece15")
		params.Dospece30 = c.Query("dospece30")
		params.Dospece60 = c.Query("dospece60")
		params.Dospece90 = c.Query("dospece90")
		params.Dospece120 = c.Query("dospece120")

		fieldsError := common.ValidateRequiredParams(c, []string{"odkonta", "dokonta", "odsifre", "dosifre", "stanjenadan", "tip_pregleda", "dospece15", "dospece30", "dospece60", "dospece90", "dospece120"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		setHeaderTextfromRequest(c, &headerFileds)
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetPregledPotrazivanjaObaveze(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		err = h.service.GetPregledPotrazivanjaObaveze(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
}

// PotrazivanjaDugovanjaStampa renders the printable pregled potrazivanja/obaveze report grouped by partner.
func (h *OtvoreneStavkeHandler) PotrazivanjaDugovanjaStampa(c *gin.Context) {
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	ctx := c.Request.Context()
	params := domain.OtvStavkeParam{
		OdKonta:     c.Query("odkonta"),
		DoKonta:     c.Query("dokonta"),
		OdSifre:     c.Query("odsifre"),
		DoSifre:     c.Query("dosifre"),
		StanjeNaDan: c.Query("stanjenadan"),
		TipPregleda: c.Query("tip_pregleda"),
		Dospece15:   c.Query("dospece15"),
		Dospece30:   c.Query("dospece30"),
		Dospece60:   c.Query("dospece60"),
		Dospece90:   c.Query("dospece90"),
		Dospece120:  c.Query("dospece120"),
	}
	standeNaDanFmt := params.StanjeNaDan
	if t, err := time.Parse("2006-01-02", params.StanjeNaDan); err == nil {
		standeNaDanFmt = t.Format(common.DateLayout)
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	val15, _ := strconv.Atoi(params.Dospece15)
	val30, _ := strconv.Atoi(params.Dospece30)
	val60, _ := strconv.Atoi(params.Dospece60)
	val90, _ := strconv.Atoi(params.Dospece90)
	repParams := domain.ReportParameters{
		ReportName:  translator.Title("PREGLED POTRAZIVANJA/DUGOVANJA"),
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		SifDel:      fvrData.SifDel,
		ParameterItems: map[string]domain.ParameterItem{
			"StanjeNaDan":     {Name: translator.Label("Stanje na dan"), Value: standeNaDanFmt},
			"TipPregleda":     {Name: "", Value: params.TipPregleda},
			"DLabel15":        {Name: "", Value: "0-" + params.Dospece15 + " " + translator.Label("Dana")},
			"DLabel30":        {Name: "", Value: strconv.Itoa(val15+1) + "-" + params.Dospece30 + " " + translator.Label("Dana")},
			"DLabel60":        {Name: "", Value: strconv.Itoa(val30+1) + "-" + params.Dospece60 + " " + translator.Label("Dana")},
			"DLabel90":        {Name: "", Value: strconv.Itoa(val60+1) + "-" + params.Dospece90 + " " + translator.Label("Dana")},
			"DLabel120":       {Name: "", Value: strconv.Itoa(val90+1) + "-" + params.Dospece120 + " " + translator.Label("Dana")},
			"DLabel120Plus":   {Name: "", Value: ">" + params.Dospece120 + " " + translator.Label("Dana")},
			"StampajSamoZbir": {Name: "", Value: c.Query("stampaj_samo_zbir")},
		},
	}
	partners, err := h.service.GetPregledPotrazivanjaObavezeStampa(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.PotrazivanjaDugovanjaStampa(repParams, partners, translator).Render(ctx, c.Writer)
}

// PregledDospelogDugaPoStarosti - Tab 7: Pregled dospelog duga po starosti
func (h *OtvoreneStavkeHandler) PregledDospelogDugaPoStarosti(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	h.setOtvoreneStavkeActiveTab("pregleddospelogdugapostarosti")
	gnGod := userSession.SelectedGod
	headerFileds := h.service.GetDugovanjaFields()

	tbl := common.SetTableBasicData(dospelaStarostiContentTitle, dospelaStarostiTableID, headerFileds, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, dospelaStarostiContentTitle, "", false, false, false)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), dospelaStarostiURL, fmt.Sprintf("#%s", dospelaStarostiTableID), hxValsDugovanja)
	tbl.Pagination.HxVals = hxValsDugovanja
	tbl.URLGetAll = dospelaStarostiURL
	tbl.URLPrefix = dospelaStarostiURL
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", dospelaStarostiURL, fmt.Sprintf("#%s", dospelaStarostiTableID), "innerHTML", "GET", "", hxValsDugovanja, true, common.ClassSaveButton, "handleBackendResponse")
	btnPrint := common.SetPrintButton("stampa-btn", translator.Button("Štampa"), "fin_stampa", dospelaStarostiStampaURL, "GET", true, common.ClassPrintButton, "odkonta,dokonta,odsifre,dosifre,stanjenadan,dospece15,dospece30,dospece60,dospece90,dospece120,tip_pregleda")

	if requestSource == "menu" || requestSource == "tab" {
		tbl.ShowPagination = true
		err := tmpl_fin.PregledDospelogDugaPoStarosti(h.tabData, tbl, btnObrada, btnPrint, searchInput, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.OtvStavkeParam{}
		params.SearchText = c.Query("query")
		params.OdKonta = c.Query("odkonta")
		params.DoKonta = c.Query("dokonta")
		params.OdSifre = c.Query("odsifre")
		params.DoSifre = c.Query("dosifre")
		params.StanjeNaDan = c.Query("stanjenadan")
		params.TipPregleda = c.Query("tip_pregleda")
		params.Dospece15 = c.Query("dospece15")
		params.Dospece30 = c.Query("dospece30")
		params.Dospece60 = c.Query("dospece60")
		params.Dospece90 = c.Query("dospece90")
		params.Dospece120 = c.Query("dospece120")

		fieldsError := common.ValidateRequiredParams(c, []string{"odkonta", "dokonta", "odsifre", "dosifre", "stanjenadan", "tip_pregleda", "dospece15", "dospece30", "dospece60", "dospece90", "dospece120"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		setHeaderTextfromRequest(c, &headerFileds)
		tbl := common.SetTableBasicData(dospelaStarostiContentTitle, dospelaStarostiTableID, headerFileds, "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, dospelaStarostiContentTitle, "", false, false, false)
		tbl.Pagination.HxVals = hxValsDugovanja
		tbl.URLGetAll = dospelaStarostiURL
		tbl.URLPrefix = dospelaStarostiURL
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetPregledDugovanjaPoStarosti(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		err = h.service.GetPregledDugovanjaPoStarosti(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
}

// PregledDospelogDugaPoStarostiStampa - Tab 7: Print report for dospeli dug po starosti
func (h *OtvoreneStavkeHandler) PregledDospelogDugaPoStarostiStampa(c *gin.Context) {
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	ctx := c.Request.Context()
	params := domain.OtvStavkeParam{
		OdKonta:     c.Query("odkonta"),
		DoKonta:     c.Query("dokonta"),
		OdSifre:     c.Query("odsifre"),
		DoSifre:     c.Query("dosifre"),
		StanjeNaDan: c.Query("stanjenadan"),
		TipPregleda: c.Query("tip_pregleda"),
		Dospece15:   c.Query("dospece15"),
		Dospece30:   c.Query("dospece30"),
		Dospece60:   c.Query("dospece60"),
		Dospece90:   c.Query("dospece90"),
		Dospece120:  c.Query("dospece120"),
	}
	standeNaDanFmt := params.StanjeNaDan
	if t, err := time.Parse("2006-01-02", params.StanjeNaDan); err == nil {
		standeNaDanFmt = t.Format(common.DateLayout)
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	repParams := domain.ReportParameters{
		ReportName:  translator.Title("PREGLED DUGOVANJA KUPACA / DOSPELI DUG PO STAROSTI"),
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		SifDel:      fvrData.SifDel,
		ParameterItems: map[string]domain.ParameterItem{
			"StanjeNaDan": {Name: translator.Label("Stanje na dan"), Value: standeNaDanFmt},
		},
	}
	tbl := common.SetTableBasicData(dospelaStarostiContentTitle, dospelaStarostiTableID, h.service.GetDugovanjaStarostiStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.HasTotals = true
	err = h.service.GetPregledDugovanjaPoStarostiStampa(ctx, &tbl, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.DospelogDugaPoStarostiStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

// PovezivanjeRacunaUplata - Tab 8: Povezivanje računa i uplata
func (h *OtvoreneStavkeHandler) PovezivanjeRacunaUplata(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "Unauthorized")
		return
	}
	gnGod := userSession.SelectedGod

	tblPartneri := common.SetTableBasicData("PARTNERI", povezivanjePartneriTableID, h.service.GetPovezivanjePartneriFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblPartneri, "PARTNERI", "", false, false, false)
	tblUplate := common.SetTableBasicData("UPLATE", povezivanjeUplateTableID, h.service.GetPovezivanjeUplateFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblUplate, "UPLATE", "", false, false, false)
	tblUplate.HasTotals = true
	tblFakture := common.SetTableBasicData("RAČUNI", povezivanjeFaktureTableID, h.service.GetPovezivanjeFaktureFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblFakture, "RAČUNI", "", false, false, false)
	tblFakture.HasTotals = true

	btnObrada := common.SetButton("btn-obrada", "Obradi", "fin_obrada", "api/otvorenestavke/povezivanje", "#"+povezivanjePartneriTableID, "innerHTML", "GET", "", hxValsPovezivanje, true, common.ClassSaveButton, "handleBackendResponse")
	btnZatvaranje := common.SetButton("btn-zatvori-racun", "Zatv. račun", "fin_obrada", "api/otvorenestavke/povezivanje", "#pov-dialog-content", "innerHTML", "POST", "", hxValsPovezivanje, true, common.ClassPrintButton, "handleBackendResponse")
	btnNazad := common.SetButton("btn-nazad", "Nazad", "", "", "", "", "", "", "", true, common.ClassButton, "")

	searchPartneri := common.CreateSearchInput("search-partneri", i18n.GetInstance(), "api/otvorenestavke/povezivanje", "#"+povezivanjePartneriTableID, hxValsPovezivanje)
	tblPartneri.URLGetAll = "api/otvorenestavke/povezivanje"
	tblPartneri.URLPrefix = "api/otvorenestavke/povezivanje"
	tblPartneri.Pagination.HxVals = hxValsPovezivanje

	h.setOtvoreneStavkeActiveTab("povezivanje")

	if requestSource == "menu" || requestSource == "tab" {
		err := tmpl_fin.PovezivanjeRacunaUplata(h.tabData, tblPartneri, tblUplate, tblFakture, btnObrada, btnZatvaranje, btnNazad, searchPartneri, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		}
		return
	}

	// For btnobrada: render the template with data (service is TODO)
	err := tmpl_fin.PovezivanjeRacunaUplata(h.tabData, tblPartneri, tblUplate, tblFakture, btnObrada, btnZatvaranje, btnNazad, searchPartneri, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}
func setHeaderTextfromRequest(c *gin.Context, headerFields *[]domain.Fields) {
	// Get the parameter values
	dospece15 := c.Query("dospece15")
	dospece30 := c.Query("dospece30")
	dospece60 := c.Query("dospece60")
	dospece90 := c.Query("dospece90")
	dospece120 := c.Query("dospece120")

	val15, _ := strconv.Atoi(dospece15)
	val30, _ := strconv.Atoi(dospece30)
	val60, _ := strconv.Atoi(dospece60)
	val90, _ := strconv.Atoi(dospece90)

	// Create ranges based on the values
	ranges := map[string]string{
		"dospece15":      "0-" + dospece15,
		"dospece30":      strconv.Itoa(val15+1) + "-" + dospece30,
		"dospece60":      strconv.Itoa(val30+1) + "-" + dospece60,
		"dospece90":      strconv.Itoa(val60+1) + "-" + dospece90,
		"dospece120":     strconv.Itoa(val90+1) + "-" + dospece120,
		"dospece120plus": ">" + dospece120,
	}

	for i, field := range *headerFields {
		if field.Params != nil {
			if daysRange, ok := ranges[field.Name]; ok {
				(*headerFields)[i].Params["days"] = daysRange
			}
		}
	}
}

// RegisterRoutes registers the routes for the Otvorene Stavke handler
func (h *OtvoreneStavkeHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/otvorenestavke", h.OtvoreneStavkeMain)
	r.GET("api/otvorenestavke/partneri", h.OtvoreneStavke)
	r.GET("api/otvorenestavke/partneri/stampa", h.PregledOtvorenihStavkiStampa)
	r.GET("api/otvorenestavke/partneri/pregled", h.PregledOtvorenihStavkiStampa)
	r.GET("api/otvorenestavke/partneridetails/:id", h.OtvoreneStavkeDetalji)
	r.GET("api/otvorenestavke/zatvorene/partneri", h.ZatvoreneStavke)
	r.GET("api/otvorenestavke/zatvorene/partneri/stampa", h.PregledZatvorenihStavkiStampa)
	r.GET("api/otvorenestavke/zatvorene/partneridetails/:id", h.ZatvoreneStavkeDetalji)
	r.GET("api/otvorenestavke/ios/partneri", h.IOS)
	r.GET("api/otvorenestavke/ios/partneri/stampa", h.IOSDetaljiStampa)
	r.GET("api/otvorenestavke/ios/partneridetails/:id", h.IOSDetalji)
	r.GET("api/otvorenestavke/dospela/partneri", h.DospelaPotrazivanja)
	r.GET("api/otvorenestavke/dospela/partneridetails/:id", h.DospelaPotrazivanjaDetalji)
	r.GET("api/otvorenestavke/dospela/stampa/analiticki", h.DugovanjaObavezeStampa)
	r.GET("api/otvorenestavke/pregleddugovanja", h.PregledPotrazivanjaObaveze)
	r.GET("api/otvorenestavke/pregleddugovanja/stampa", h.PotrazivanjaDugovanjaStampa)
	r.GET("api/otvorenestavke/dospelidugpostarosti", h.PregledDospelogDugaPoStarosti)
	r.GET("api/otvorenestavke/dospelidugpostarosti/stampa", h.PregledDospelogDugaPoStarostiStampa)
	r.GET("api/otvorenestavke/povezivanje", h.PovezivanjeRacunaUplata)
	r.POST("api/otvorenestavke/povezivanje", h.PovezivanjeRacunaUplata)
}

// Helper functions

func GetOtvoreneStavkeTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "otvorenestavke", Label: "Otvorene stavke", HXRequestUrl: "/api/otvorenestavke/partneri", IsActive: true, Name: "otvorenestavke"},
			{ID: "zatvorenestavke", Label: "Zatvorene stavke", HXRequestUrl: "/api/otvorenestavke/zatvorene/partneri", IsActive: false, Name: "zatvorenestavke"},
			{ID: "ios", Label: "IOS", HXRequestUrl: "/api/otvorenestavke/ios/partneri", IsActive: false, Name: "ios"},
			{ID: "dospelapotrazivanja", Label: "Dospela potraživanja/dugovanja", HXRequestUrl: "/api/otvorenestavke/dospela/partneri", IsActive: false, Name: "dospelapotrazivanja"},
			{ID: "pregleddugovanja", Label: "Pregled potraživanja/dugovanja", HXRequestUrl: "/api/otvorenestavke/pregleddugovanja", IsActive: false, Name: "pregleddugovanja"},
			{ID: "pregleddospelogdugapostarosti", Label: "Pregled dospelog duga po starosti", HXRequestUrl: "/api/otvorenestavke/dospelidugpostarosti", IsActive: false, Name: "pregleddospelogdugapostarosti"},
			{ID: "povezivanje", Label: "Povezivanje računa i uplata", HXRequestUrl: "/api/otvorenestavke/povezivanje", IsActive: false, Name: "povezivanje"},
		},
	}
}

func (h *OtvoreneStavkeHandler) setOtvoreneStavkeActiveTab(tab string) {
	for i, t := range h.tabData.Tabs {
		if t.Name == tab {
			h.tabData.Tabs[i].IsActive = true
		} else {
			h.tabData.Tabs[i].IsActive = false
		}
	}
}
