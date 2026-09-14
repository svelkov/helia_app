package robno

import (
	"fmt"
	"net/http"

	"helia/config"
	tmpl_rep_rob "helia/frontend/templates/reports/robno"
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
	robnoKarticaArtikalTitle          = "Prikaz kartice artikla"
	robnoKarticaURLArtikal            = "/api/robno-kartica/artikla"
	robnoKarticaArtikalTableID        = "robnokartica-artikli-table"
	robnoKarticaURLArtikalStampa      = "/api/robno-kartica/artikla/stampa"
	robnoKarticaURLSubsintetika       = "/api/robno-kartica/subsintetickog-konta"
	robnoKarticaURLSubsintetikaStampa = "/api/robno-kartica/subsintetickog-konta/stampa"
	robnoKarticaSubsintetikaTableID   = "robnokartica-subsintetika-table"
	robnoKarticaSubsintetikaTitle     = "PRIKAZ KARTICE SUBSINTETIČKOG KONTA"

	hxValsRobnaKarticaSubsntetika = `js:{
			"sourceTab": "subsintetika",
            "magacin": document.getElementById("magacin")?.value,
			"konto": document.getElementById("konto")?.value,
			"brojnaloga": document.getElementById("brojnaloga")?.value,
            "oddatumanaloga": document.getElementById("oddatumanaloga")?.value,
            "dodatumanaloga": document.getElementById("dodatumanaloga")?.value,
			"odiznosa": document.getElementById("odiznosa")?.value,
			"doiznosa": document.getElementById("doiznosa")?.value,
			"chkpobrojunaloga": document.getElementById("chkpobrojunaloga")?.checked,
			"chkpodatumunaloga": document.getElementById("chkpodatumunaloga")?.checked,
			"chkpoiznosu": document.getElementById("chkpoiznosu")?.checked,
		}`
	hxValsRobnaKarticaArtikal = `js:{
			"sourceTab": "artikli",
            "magacin": document.getElementById("magacin")?.value,
			"konto": document.getElementById("konto")?.value,
            "brojnaloga": document.getElementById("brojnaloga")?.value,
            "oddatuma": document.getElementById("oddatuma")?.value,
            "dodatuma": document.getElementById("dodatuma")?.value,
			"oddatumanaloga": document.getElementById("oddatumanaloga")?.value,
			"dodatumanaloga": document.getElementById("dodatumanaloga")?.value,
			"oddatumaobrade": document.getElementById("oddatumaobrade")?.value,
			"dodatumaobrade": document.getElementById("dodatumaobrade")?.value,
			"sifvrstedokumenta": document.getElementById("sifvrstedokumenta")?.value,
			"brojdokumenta": document.getElementById("brojdokumenta")?.value,
			"oddatumadok": document.getElementById("oddatumadok")?.value,
			"dodatumaadok": document.getElementById("dodatumaadok")?.value,
			"odiznosa": document.getElementById("odiznosa")?.value,
			"doiznosa": document.getElementById("doiznosa")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"chkpobrojunaloga": document.getElementById("chkpobrojunaloga")?.checked,
			"chkpodatumunaloga": document.getElementById("chkpodatumunaloga")?.checked,
			"chkpodatumuobrade": document.getElementById("chkpodatumuobrade")?.checked,
			"chkpovrstidokumenta": document.getElementById("chkpovrstidokumenta")?.checked,
			"chkpobrojudokumenta": document.getElementById("chkpobrojudokumenta")?.checked,
			"chkpodatumdokumenta": document.getElementById("chkpodatumdokumenta")?.checked,
			"chkpoiznosu": document.getElementById("chkpoiznosu")?.checked,
			"chkpRPROID": document.getElementById("chkpRPROID")?.checked,
			"stampajpomesecima": document.getElementById("stampajpomesecima")?.checked,
			"sortiranje": document.querySelector('input[name="sortiranje"]:checked')?.value,
		}`
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
	translator := i18n.GetInstance()
	common.SetActiveTab(h.tabs, 0)
	magValues, err := h.service.GetMagacinComboValues(c.Request.Context())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tblData := common.SetTableBasicData(robnoKarticaArtikalTitle, robnoKarticaArtikalTableID, h.service.GetKarticaArtiklaTableFields(), robnoKarticaURLArtikal, robnoKarticaURLArtikal, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblData, robnoKarticaArtikalTableID, robnoKarticaURLArtikal, false, false, false)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoKarticaURLArtikal, "#"+robnoKarticaArtikalTableID, "innerHTML", "GET", "", hxValsRobnaKarticaArtikal, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", robnoKarticaURLArtikalStampa, "#"+robnoKarticaArtikalTableID, "innerHTML", "GET", "", hxValsRobnaKarticaArtikal, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", translator, robnoKarticaURLArtikal, fmt.Sprintf("#%s", robnoKarticaArtikalTableID), hxValsRobnaKarticaArtikal)

	tmpl_robno.RobnoKarticaMain(*h.tabs, tblData, magValues, btnObrada, btnPrint, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *RobnoKarticaHandler) GetPrikazKarticeArtikla(c *gin.Context) {
	ctx := c.Request.Context()
	requestSource := c.Request.Header.Get("X-Request-Source")
	common.SetActiveTab(h.tabs, 0)
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tbl := common.SetTableBasicData(robnoKarticaArtikalTitle, robnoKarticaURLArtikal, h.service.GetKarticaArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKarticaArtikalTitle, "", false, false, false)
	tbl.HasTotals = true
	if requestSource == "menu" || requestSource == "tab" {
		magValues, err := h.service.GetMagacinComboValues(ctx)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoKarticaURLArtikal, "#"+robnoKarticaArtikalTableID, "innerHTML", "GET", "", hxValsRobnaKarticaArtikal, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", robnoKarticaURLArtikalStampa, "#"+robnoKarticaArtikalTableID, "innerHTML", "GET", "", hxValsRobnaKarticaArtikal, true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput("search-input", translator, robnoKarticaURLArtikal, fmt.Sprintf("#%s", robnoKarticaArtikalTableID), hxValsRobnaKarticaArtikal)
		tmpl_robno.RobnoKarticaArtikla(*h.tabs, tbl, magValues, btnObrada, btnPrint, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation)
			return
		}
		odIznosa, _ := utils.GetFloat64FromQueryRequest(c, "odiznosa")
		doIznosa, _ := utils.GetFloat64FromQueryRequest(c, "doiznosa")
		params := domain.RobnoKarticaParams{
			Magacin:           magacin,
			Konto:             c.Query("konto"),
			Nalozi:            c.Query("brojnaloga"),
			OdDanal:           c.Query("oddatumanaloga"),
			DoDanal:           c.Query("dodatumanaloga"),
			OdDatumObrade:     c.Query("oddatumaobrade"),
			DoDatumObrade:     c.Query("dodatumaobrade"),
			SifVrsteDokumenta: c.Query("sifvrstedokumenta"),
			BrojDokumenta:     c.Query("brojdokumenta"),
			OdDatumDok:        c.Query("oddatumadok"),
			DoDatumDok:        c.Query("dodatumaadok"),
			OdIznosa:          odIznosa,
			DoIznosa:          doIznosa,
			OdSifre:           c.Query("odsifre"),
			DoSifre:           c.Query("dosifre"),
			CbxDatum:          c.Query("chkpodatumunaloga") == "true",
			CbxBrojNaloga:     c.Query("chkpobrojunaloga") == "true",
			CbxDatumObrade:    c.Query("chkpodatumuobrade") == "true",
			CbxVrstaDokumenta: c.Query("chkpovrstidokumenta") == "true",
			CbxBrojDokumenta:  c.Query("chkpobrojudokumenta") == "true",
			CbxDatumDokumenta: c.Query("chkpodatumdokumenta") == "true",
			CbxIznos:          c.Query("chkpoiznosu") == "true",
			CbxRPROID:         c.Query("chkpRPROID") == "true",
			StampajPoMesecima: c.Query("stampajpomesecima") == "true",
			Sortiranje:        c.Query("sortiranje"),
			SearchText:        c.Query("query"),
			ReportTip:         "karticaartikla",
		}
		//validacija input parametre:
		fieldsError := h.service.ValidacijaKarticaArtikla(params)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData(robnoKarticaArtikalTitle, robnoKarticaArtikalTableID, h.service.GetKarticaArtiklaTableFields(), "", robnoKarticaURLArtikal, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", robnoKarticaURLArtikal, false, false, false)
		tbl.Pagination.HxVals = hxValsRobnaKarticaArtikal

		err = h.service.GetKarticaArtikla(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetKarticaArtikla(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
		return
	}
}

func (h *RobnoKarticaHandler) GetKarticaSubsintetickogKonta(c *gin.Context) {
	ctx := c.Request.Context()
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	common.SetActiveTab(h.tabs, 1)
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tbl := common.SetTableBasicData(robnoKarticaSubsintetikaTitle, robnoKarticaSubsintetikaTableID, h.service.GetKarticaSubsintetickogKontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKarticaSubsintetikaTitle, "", false, false, false)
	tbl.HasTotals = true
	if requestSource == "menu" || requestSource == "tab" {
		magValues, err := h.service.GetMagacinComboValues(ctx)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		printFields := "magacin,konto,brojnaloga,oddatumanaloga,dodatumanaloga,odiznosa,doiznosa,chkpobrojunaloga,chkpodatumunaloga,chkpoiznosu"
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoKarticaURLSubsintetika, "#"+robnoKarticaSubsintetikaTableID, "innerHTML", "GET", "", hxValsRobnaKarticaSubsntetika, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoKarticaURLSubsintetikaStampa, "GET", true, common.ClassPrintButton, printFields)
		searchInput := common.CreateSearchInput("search-input", translator, robnoKarticaURLSubsintetika, fmt.Sprintf("#%s", robnoKarticaSubsintetikaTableID), hxValsRobnaKarticaSubsntetika)
		tmpl_robno.RobnoKarticaSubsintetickoKonto(*h.tabs, tbl, magValues, btnObrada, btnPrint, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		return
	}
	// If it's a POST request, the make obrada
	if requestSource == "btn" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation)
			return
		}
		odIznosa, _ := utils.GetFloat64FromQueryRequest(c, "odiznosa")
		doIznosa, _ := utils.GetFloat64FromQueryRequest(c, "doiznosa")
		params := domain.RobnoKarticaParams{
			Magacin:       magacin,
			Konto:         c.Query("konto"),
			OdDanal:       c.Query("oddatumanaloga"),
			DoDanal:       c.Query("dodatumanaloga"),
			OdIznosa:      odIznosa,
			DoIznosa:      doIznosa,
			CbxDatum:      c.Query("chkpodatumunaloga") == "true",
			CbxBrojNaloga: c.Query("chkpobrojunaloga") == "true",
			CbxIznos:      c.Query("chkpoiznosu") == "true",
			SearchText:    c.Query("query"),
			ReportTip:     "karticasubsintetickogkonta",
		}
		//validacija input parametre:
		fieldsError := h.service.ValidacijaSubsintetickogKonta(params)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData(robnoKarticaSubsintetikaTitle, robnoKarticaSubsintetikaTableID, h.service.GetKarticaSubsintetickogKontaTableFields(), "", robnoKarticaURLSubsintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, robnoKarticaSubsintetikaTableID, robnoKarticaURLSubsintetika, false, false, false)
		tbl.Pagination.HxVals = hxValsRobnaKarticaSubsntetika

		err = h.service.GetKarticaSubsintetickogKonta(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetKarticaSubsintetickogKonta(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
		return
	}
}

func (h *RobnoKarticaHandler) GetPrikazKarticeArtiklaStampa(c *gin.Context) {

}

func (h *RobnoKarticaHandler) GetKarticaSubsintetickogKontaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	//translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation)
		return
	}
	odIznosa, _ := utils.GetFloat64FromQueryRequest(c, "odiznosa")
	doIznosa, _ := utils.GetFloat64FromQueryRequest(c, "doiznosa")
	params := domain.RobnoKarticaParams{
		Magacin:       magacin,
		Konto:         c.Query("konto"),
		Nalozi:        c.Query("brojnaloga"),
		OdDanal:       c.Query("oddatumanaloga"),
		DoDanal:       c.Query("dodatumanaloga"),
		OdIznosa:      odIznosa,
		DoIznosa:      doIznosa,
		CbxDatum:      c.Query("chkpodatumunaloga") == "true",
		CbxBrojNaloga: c.Query("chkpobrojunaloga") == "true",
		CbxIznos:      c.Query("chkpoiznosu") == "true",
		SearchText:    c.Query("query"),
		ReportTip:     "karticasubsintetickogkonta",
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	tbl := common.SetTableBasicData(robnoKarticaSubsintetikaTitle, robnoKarticaSubsintetikaTableID, h.service.GetKarticaSubsintetickogKontaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetKarticaSubsintetickogKontaStampa(ctx, &tbl, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
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
		ReportName:  "Kartica subsintetičkog konta",
		ParameterItems: map[string]domain.ParameterItem{
			"Magacin":        {Name: "Magacin", Value: fmt.Sprintf("%d", magacin)},
			"Konto":          {Name: "Konto", Value: params.Konto},
			"BrojNaloga":     {Name: "Broj naloga", Value: params.Nalozi},
			"OdDatumaNaloga": {Name: "Od datuma naloga", Value: params.OdDanal},
			"DoDatumaNaloga": {Name: "Do datuma naloga", Value: params.DoDanal},
			"OdIznosa":       {Name: "Od iznosa", Value: common.FormatNumberWithSystemLocale(params.OdIznosa, 2)},
			"DoIznosa":       {Name: "Do iznosa", Value: common.FormatNumberWithSystemLocale(params.DoIznosa, 2)},
		},
	}

	tmpl_rep_rob.RobnoKarticaSubsintetickogKontaStampa(repParams, params, tbl, i18n.GetInstance()).Render(ctx, c.Writer)

}

func (h *RobnoKarticaHandler) AddRoutes(r *gin.Engine) {
	// Apply auth middleware to all Robno Promet routes.
	r.Use(middleware.Auth())

	// Define routes for Robno Promet.
	r.GET("/api/robno-kartica", h.RobnoKarticaMain)
	r.GET("/api/robno-kartica/artikla", h.GetPrikazKarticeArtikla)
	r.GET("/api/robno-kartica/artikla/stampa", h.GetPrikazKarticeArtiklaStampa)
	r.GET("/api/robno-kartica/subsintetickog-konta", h.GetKarticaSubsintetickogKonta)
	r.GET("/api/robno-kartica/subsintetickog-konta/stampa", h.GetKarticaSubsintetickogKontaStampa)
}

func robnoKarticaTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnokartica-artikli", Label: translator.Label("Prikaz kartice artikla"), HXRequestUrl: robnoKarticaURLArtikal, IsActive: true, Name: "artikli"},
		{ID: "robnokartica-subsintetika", Label: translator.Label("Prikaz kartice subsintetičkog konta"), HXRequestUrl: robnoKarticaURLSubsintetika, IsActive: false, Name: "subsintetika"},
	}}
}
