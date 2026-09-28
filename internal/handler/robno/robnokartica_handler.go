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
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	robnoKarticaArtikalTitle          = "PRIKAZ KARTICE ARTIKLA"
	robnoKarticaURLArtikal            = "/api/robno-kartica/artikla"
	robnoKarticaArtikalTableID        = "robnokartica-artikli-table"
	robnoKarticaURLArtikalStampa      = "/api/robno-kartica/artikla/stampa"
	robnoKarticaURLSubsintetika       = "/api/robno-kartica/subsintetickog-konta"
	robnoKarticaURLSubsintetikaStampa = "/api/robno-kartica/subsintetickog-konta/stampa"
	robnoKarticaSubsintetikaTableID   = "robnokartica-subsintetika-table"
	robnoKarticaSubsintetikaTitle     = "PRIKAZ KARTICE SUBSINTETIČKOG KONTA"
	robnoKarticaInfoMessageDialogID   = "info-message-dialog"

	printFieldsSubsintetika = "magacin,konto,brojnaloga,oddatumanaloga,dodatumanaloga,odiznosa,doiznosa,chkpobrojunaloga,chkpodatumunaloga,chkpoiznosu"
	printFieldsArtikal      = "magacin,konto,brojnaloga,oddatumanaloga,dodatumanaloga,oddatumaobrade,dodatumaobrade,sifvrstedokumenta,brojdokumenta,oddatumadok,dodatumaadok,odiznosa,doiznosa,odsifreartikla,dosifreartikla,chkpobrojunaloga,chkpodatumunaloga,chkpodatumuobrade,chkpovrstidokumenta,chkpobrojudokumenta,chkpodatumdokumenta,chkpoiznosu,chkpRPROID,stampajpomesecima,sortiranje"

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
			"chkpobrojunaloga": document.getElementById("chkpobrojunaloga")?.checked,
            "brojnaloga": document.getElementById("brojnaloga")?.value,
			"chkpodatumunaloga": document.getElementById("chkpodatumunaloga")?.checked,
			"oddatumanaloga": document.getElementById("oddatumanaloga")?.value,
			"dodatumanaloga": document.getElementById("dodatumanaloga")?.value,
			"chkpodatumuobrade": document.getElementById("chkpodatumuobrade")?.checked,
			"oddatumaobrade": document.getElementById("oddatumaobrade")?.value,
			"dodatumaobrade": document.getElementById("dodatumaobrade")?.value,
			"chkpovrstidokumenta": document.getElementById("chkpovrstidokumenta")?.checked,
			"sifvrstedokumenta": document.getElementById("sifvrstedokumenta")?.value,
			"chkpobrojudokumenta": document.getElementById("chkpobrojudokumenta")?.checked,
			"brojdokumenta": document.getElementById("brojdokumenta")?.value,
			"chkpodatumdokumenta": document.getElementById("chkpodatumdokumenta")?.checked,
		    "dodatumadok": document.getElementById("dodatumadok")?.value,
            "oddatumadok": document.getElementById("oddatumadok")?.value,
			"odiznosa": document.getElementById("odiznosa")?.value,
			"chkpoiznosu": document.getElementById("chkpoiznosu")?.checked,
			"doiznosa": document.getElementById("doiznosa")?.value,
			"odsifreartikla": document.getElementById("odsifreartikla")?.value,
			"dosifreartikla": document.getElementById("dosifreartikla")?.value,
			"chkpRPROID": document.getElementById("chkpRPROID")?.checked,
			"stampajpomesecima": document.getElementById("stampajpomesecima")?.checked,
			"sortiranje": document.querySelector('input[name="sortiranje"]:checked')?.value,
		}`
)

type RobnoKarticaHandler struct {
	service robnosvc.RobnoKarticaService
	cfg     config.Config
	tabs    domain.TabData
}

func NewRobnoKarticaHandler(service robnosvc.RobnoKarticaService, cfg config.Config) *RobnoKarticaHandler {
	return &RobnoKarticaHandler{service: service, cfg: cfg, tabs: robnoKarticaTabs()}
}

func (h *RobnoKarticaHandler) RobnoKarticaMain(c *gin.Context) {
	translator := i18n.GetInstance()
	tabs := common.SetActiveTab(h.tabs, 0)
	magValues, err := h.service.GetMagacinComboValues(c.Request.Context())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tblData := common.SetTableBasicData(robnoKarticaArtikalTitle, robnoKarticaArtikalTableID, h.service.GetKarticaArtiklaTableFields(), robnoKarticaURLArtikal, robnoKarticaURLArtikal, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblData, robnoKarticaArtikalTableID, robnoKarticaURLArtikal, false, false, false)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoKarticaURLArtikal, "#"+robnoKarticaArtikalTableID, "innerHTML", "GET", "", hxValsRobnaKarticaArtikal, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoKarticaURLArtikalStampa, "GET", true, common.ClassPrintButton, printFieldsArtikal)
	searchInput := common.CreateSearchInput("search-input", translator, robnoKarticaURLArtikal, fmt.Sprintf("#%s", robnoKarticaArtikalTableID), hxValsRobnaKarticaArtikal)

	tmpl_robno.RobnoKarticaMain(tabs, tblData, magValues, btnObrada, btnPrint, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *RobnoKarticaHandler) GetPrikazKarticeArtikla(c *gin.Context) {
	ctx := c.Request.Context()
	tabs := common.SetActiveTab(h.tabs, 0)
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	tbl := common.SetTableBasicData(robnoKarticaArtikalTitle, robnoKarticaArtikalTableID, h.service.GetKarticaArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKarticaArtikalTitle, robnoKarticaURLArtikal, false, false, false)
	tbl.HasTotals = true
	if common.IsDataRequest(c) {
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataConversion, err.Error()))
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
			OdDatumaDok:       c.Query("oddatumadok"),
			DoDatumaDok:       c.Query("dodatumadok"),
			OdIznosa:          odIznosa,
			DoIznosa:          doIznosa,
			OdSifre:           c.Query("odsifreartikla"),
			DoSifre:           c.Query("dosifreartikla"),
			CbxDatumNaloga:    c.Query("chkpodatumunaloga") == "true",
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
		common.SetTableConfig(&tbl, "PRIKAZ KARTICE ARTIKLA", robnoKarticaURLArtikal, false, false, false)
		tbl.Pagination.HxVals = hxValsRobnaKarticaArtikal

		err = h.service.GetKarticaArtikla(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}

		err = h.service.GetKarticaArtikla(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
		return
	}
	// Handle data request here
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoKarticaURLArtikal, "#"+robnoKarticaArtikalTableID, "innerHTML", "GET", "", hxValsRobnaKarticaArtikal, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoKarticaURLArtikalStampa, "GET", true, common.ClassPrintButton, printFieldsArtikal)
	searchInput := common.CreateSearchInput("search-input", translator, robnoKarticaURLArtikal, fmt.Sprintf("#%s", robnoKarticaArtikalTableID), hxValsRobnaKarticaArtikal)
	tmpl_robno.RobnoKarticaArtikla(tabs, tbl, magValues, btnObrada, btnPrint, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *RobnoKarticaHandler) GetPrikazKarticeArtiklaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgValidation)
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
		OdDatumaDok:       c.Query("oddatumadok"),
		DoDatumaDok:       c.Query("dodatumaadok"),
		OdIznosa:          odIznosa,
		DoIznosa:          doIznosa,
		OdSifre:           c.Query("odsifreartikla"),
		DoSifre:           c.Query("dosifreartikla"),
		CbxBrojNaloga:     c.Query("chkpobrojunaloga") == "true",
		CbxDatumNaloga:    c.Query("chkpodatumunaloga") == "true",
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
	if fieldsError := h.service.ValidacijaKarticaArtikla(params); len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldsError, common.ErrMsgValidation)
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	tbl := common.SetTableBasicData(robnoKarticaArtikalTitle, robnoKarticaArtikalTableID, h.service.GetKarticaArtiklaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetKarticaArtikla(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
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
		ReportName:  "KARTICA ARTIKALA",
		ParameterItems: map[string]domain.ParameterItem{
			"Magacin":           {Name: "Magacin", Value: fmt.Sprintf("%d", params.Magacin)},
			"OdSifre":           {Name: "Od artikla", Value: params.OdSifre},
			"DoSifre":           {Name: "Do artikla", Value: params.DoSifre},
			"BrojNaloga":        {Name: "Broj naloga", Value: params.Nalozi},
			"OdDatumaNaloga":    {Name: "Od datuma naloga", Value: params.OdDanal},
			"DoDatumaNaloga":    {Name: "Do datuma naloga", Value: params.DoDanal},
			"OdDatumaObrade":    {Name: "Od datuma obrade", Value: params.OdDatumObrade},
			"DoDatumaObrade":    {Name: "Do datuma obrade", Value: params.DoDatumObrade},
			"SifVrsteDokumenta": {Name: "Šifra vrste dokumenta", Value: params.SifVrsteDokumenta},
			"BrojDokumenta":     {Name: "Broj dokumenta", Value: params.BrojDokumenta},
			"OdDatumaDokumenta": {Name: "Od datuma dokumenta", Value: params.OdDatumaDok},
			"DoDatumaDokumenta": {Name: "Do datuma dokumenta", Value: params.DoDatumaDok},
			"OdIznosa":          {Name: "Od iznosa", Value: common.FormatNumberWithSystemLocale(params.OdIznosa, 2)},
			"DoIznosa":          {Name: "Do iznosa", Value: common.FormatNumberWithSystemLocale(params.DoIznosa, 2)},
			"Sortiranje":        {Name: "Sortiranje", Value: sortiranjeNaziv(params.Sortiranje)},
		},
	}
	tmpl_rep_rob.RobnoKarticaArtikalStampa(repParams, params, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}
func (h *RobnoKarticaHandler) GetKarticaSubsintetickogKonta(c *gin.Context) {
	ctx := c.Request.Context()
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
	if common.IsDataRequest(c) {
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, err.Error())
			return
		}
		odIznosa, _ := utils.GetFloat64FromQueryRequest(c, "odiznosa")
		doIznosa, _ := utils.GetFloat64FromQueryRequest(c, "doiznosa")
		params := domain.RobnoKarticaParams{
			Magacin:        magacin,
			Konto:          c.Query("konto"),
			OdDanal:        c.Query("oddatumanaloga"),
			DoDanal:        c.Query("dodatumanaloga"),
			Nalozi:         c.Query("brojnaloga"),
			OdIznosa:       odIznosa,
			DoIznosa:       doIznosa,
			CbxDatumNaloga: c.Query("chkpodatumunaloga") == "true",
			CbxBrojNaloga:  c.Query("chkpobrojunaloga") == "true",
			CbxIznos:       c.Query("chkpoiznosu") == "true",
			SearchText:     c.Query("query"),
			ReportTip:      "karticasubsintetickogkonta",
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
		err = h.service.GetKarticaSubsintetickogKonta(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetKarticaSubsintetickogKonta(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, err.Error())
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", robnoKarticaURLSubsintetika, "#"+robnoKarticaSubsintetikaTableID, "innerHTML", "GET", "", hxValsRobnaKarticaSubsntetika, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoKarticaURLSubsintetikaStampa, "GET", true, common.ClassPrintButton, printFieldsSubsintetika)
	searchInput := common.CreateSearchInput("search-input", translator, robnoKarticaURLSubsintetika, fmt.Sprintf("#%s", robnoKarticaSubsintetikaTableID), hxValsRobnaKarticaSubsntetika)
	tmpl_robno.RobnoKarticaSubsintetickoKonto(h.tabs, tbl, magValues, btnObrada, btnPrint, searchInput, userSession.SelectedGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *RobnoKarticaHandler) GetKarticaSubsintetickogKontaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	//translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, common.ErrMsgUnauthorized)
		return
	}
	magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
	if err != nil {
		utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, err.Error())
		return
	}
	odIznosa, _ := utils.GetFloat64FromQueryRequest(c, "odiznosa")
	doIznosa, _ := utils.GetFloat64FromQueryRequest(c, "doiznosa")
	params := domain.RobnoKarticaParams{
		Magacin:        magacin,
		Konto:          c.Query("konto"),
		Nalozi:         c.Query("brojnaloga"),
		OdDanal:        c.Query("oddatumanaloga"),
		DoDanal:        c.Query("dodatumanaloga"),
		OdIznosa:       odIznosa,
		DoIznosa:       doIznosa,
		CbxDatumNaloga: c.Query("chkpodatumunaloga") == "true",
		CbxBrojNaloga:  c.Query("chkpobrojunaloga") == "true",
		CbxIznos:       c.Query("chkpoiznosu") == "true",
		SearchText:     c.Query("query"),
		ReportTip:      "karticasubsintetickogkonta",
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tbl := common.SetTableBasicData(robnoKarticaSubsintetikaTitle, robnoKarticaSubsintetikaTableID, h.service.GetKarticaSubsintetickogKontaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetKarticaSubsintetickogKonta(ctx, &tbl, false, 0, 0, params, common.TipStampePrint)
	if err != nil {
		utils.RenderDialogOK(c, robnoKarticaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
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
	// Define routes for Robno Promet.
	r.GET("/api/robno-kartica", h.RobnoKarticaMain)
	r.GET("/api/robno-kartica/artikla", h.GetPrikazKarticeArtikla)
	r.GET("/api/robno-kartica/artikla/stampa", h.GetPrikazKarticeArtiklaStampa)
	r.GET("/api/robno-kartica/subsintetickog-konta", h.GetKarticaSubsintetickogKonta)
	r.GET("/api/robno-kartica/subsintetickog-konta/stampa", h.GetKarticaSubsintetickogKontaStampa)
}

func robnoKarticaTabs() domain.TabData {
	translator := i18n.GetInstance()
	return domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnokartica-artikli", Label: translator.Label("Prikaz kartice artikla"), HXRequestUrl: robnoKarticaURLArtikal, IsActive: true, Name: "artikli"},
		{ID: "robnokartica-subsintetika", Label: translator.Label("Prikaz kartice subsintetičkog konta"), HXRequestUrl: robnoKarticaURLSubsintetika, IsActive: false, Name: "subsintetika"},
	}}
}

func sortiranjeNaziv(sortiranje string) string {
	if sortiranje == "dokument" {
		return "Sortiranje po datumu dokumenta"
	}
	return "Sortiranje po datumu naloga"
}
