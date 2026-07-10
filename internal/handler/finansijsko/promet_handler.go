package finansijsko

import (
	"fmt"
	"helia/config"
	"net/http"
	"reflect"
	"strings"

	tmpl_fin "helia/frontend/templates/finansijsko"
	tmpl_rep_fin "helia/frontend/templates/reports/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	prometContentTitle                         string = "PROMET"
	prometTableID                              string = "promettable"
	prometURLPrefix                            string = "/api/promet/"
	prometURLGetAll                            string = "/api/promet/all"
	prometURLAnKonta                           string = "/api/promet/analitickakonta"
	prometKontaAnalitickiDodatniParametriURL   string = "/api/promet/analitickakonta/dodatniparametri"
	prometURLAnKontaMi                         string = "/api/promet/analitickakontami"
	prometURLDeviznaKonta                      string = "/api/promet/deviznihanalitickihkonta"
	prometURLSubsintetika                      string = "/api/promet/subsintetickakonta"
	prometURLSintetika                         string = "/api/promet/sintetickakonta"
	prometURLKarticaSintetika                  string = "/api/promet/karticasintetickihkonta"
	prometURLSubsintetikaVrd                   string = "/api/promet/subsintetickakontapovrd"
	prometURLKontaAnaliticki                   string = "/api/promet/kontaanaliticki"
	prometURLAnalitickaKarticaStampaDialog     string = "/api/promet/analitickakonta/stampadialog"
	prometURLAnalitickaKarticaStampa           string = "/api/promet/analitickakonta/stampa"
	prometURLAnalitickaKarticaPoMIStampaDialog string = "/api/promet/analitickakontami/stampadialog"
	prometURLAnalitickaKarticaPoMIStampa       string = "/api/promet/analitickakontami/stampa"
	prometURLSubsintetikaStampaDialog          string = "/api/promet/subsintetickakonta/stampadialog"
	prometURLSubsintetikaStampa                string = "/api/promet/subsintetickakonta/stampa"
	prometURLSintetikaStampaDialog             string = "/api/promet/sintetickihkonta/stampadialog"
	prometURLSintetikaStampa                   string = "/api/promet/sintetickihkonta/stampa"
	prometURLKarticaSintetikaStampaDialog      string = "/api/promet/karticasintetickihkonta/stampadialog"
	prometURLKarticaSintetikaStampa            string = "/api/promet/karticasintetickihkonta/stampa"
	prometURLSubsintetikaVrdStampa             string = "/api/promet/subsintetickakontapovrd/stampa"
	stampaSintetikaFields                      string = "saldapomesecima,ukupansaldo,konto,datumstampe"
	stampaAnKarticaFileds                      string = "saldapomesecima,ukupansaldo,odkonta,dokonta,odsifre,dosifre,datumstampe,karticadeviznihkonta,karticasakolicinom,chkpogrupinaloga,odvrstenaloga,dovrstenaloga,chkpobrojunaloga,odbrojanaloga,dobrojanaloga,chkpodatumunaloga,oddatumanaloga,dodatumanaloga,chkpodatumuobrade,oddatumaobrade,dodatumaobrade,chkpovrstidokumenta,vrstidokumenta,chkpobrojudokumenta,odbrojedokumenta,dobrojedokumenta,chkpodatumdokumenta,oddatumadokumenta,dodatumadokumenta,chkpoiznosu,odiznosa,doiznosa,chkponacinknjizenja,nacinknjizenja"
	stampaSubsKarticaFields                    string = "saldapomesecima,ukupansaldo,stampadokumenta,odkonta,dokonta,datumstampe,oddatuma,dodatuma"
	stampaKarticaSintKontaFields               string = "konto,oddatuma,dodatuma,analitika"
	stampaSubsintetikaVrdFields                string = "konto,oddatuma,dodatuma"
	stampaAnKarticaMIFields                    string = "saldapomesecima,konto,sifra,oddatuma,dodatuma,odmi,domi"
)

type PrometHandler struct {
	tabData domain.TabData
	service finservice.PrometService
	cfg     config.Config
}

const (
	hxValsAnalitickihKonta = `js:{
            konto: document.getElementById("konto")?.value,
            sifra: document.getElementById("sifra")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
			dodatniparametri: document.getElementById("dodatniparametri")?.checked
        }`
	hxValsMI = `js:{
            konto: document.getElementById("konto")?.value,
            sifra: document.getElementById("sifra")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
			odmi: document.getElementById("odmi")?.value,
			domi: document.getElementById("domi")?.value,
        }`
	hxValsDeviznaKonta = `js:{
            konto: document.getElementById("konto")?.value,
            sifra: document.getElementById("sifra")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
			odmi: document.getElementById("odmi")?.value,
			domi: document.getElementById("domi")?.value,
        }`
	hxValsSubsintetika = `js:{
            odkonta: document.getElementById("odkonta")?.value,
			dokonta: document.getElementById("dokonta")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value
			}`
	hxValsSintetika = `js:{
            konto: document.getElementById("konto")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value
			}`
	hxValsKarticaSintetika = `js:{
            konto: document.getElementById("konto")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
			analitika: document.getElementById("analitika")?.checked
			}`
	hxValsSubsintetikaVrd = `js:{
            konto: document.getElementById("konto")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value
			}`
	hxValsKontaAnaliticki = `js:{
            konto: document.getElementById("konto")?.value,
			odsifre: document.getElementById("odsifre")?.value,
			dosifre: document.getElementById("dosifre")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
			}`
)

func NewPrometHandler(service finservice.PrometService, cfg config.Config) *PrometHandler {
	handler := &PrometHandler{
		service: service,
		cfg:     cfg,
	}
	handler.tabData = GetTabData()
	handler.service = service
	return handler
}

func (h *PrometHandler) PrometMain(c *gin.Context) {
	// Create configuration
	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	common.SetActiveTab(&h.tabData, "analitickakonta")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLAnKonta, "#promettable", "innerHTML", "GET", "#konto, #sifra, #oddatuma, #dodatuma", hxValsAnalitickihKonta, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", prometURLAnalitickaKarticaStampaDialog, "#dialog-proment-analitika-stampa", "innerHTML", "GET", "", hxValsAnalitickihKonta, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), prometURLAnKonta, fmt.Sprintf("#%s", prometTableID), hxValsAnalitickihKonta)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, prometContentTitle, "", false, false, false)
	tbl.FuncClick = "selectRow"                             // naziv js function for Click
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)" // naziv js function for dblClick
	err := tmpl_fin.PrometMain(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, prometKontaAnalitickiDodatniParametriURL, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *PrometHandler) PrometAnalitickihKonta(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	//if the call come from menu click or tab click then render the page with parameters and empty table
	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, prometContentTitle, "", false, false, false)
	common.SetActiveTab(&h.tabData, "analitickakonta")
	tbl.HasTotals = true
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLAnKonta, "#promettable", "innerHTML", "GET", "", hxValsAnalitickihKonta, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", prometURLAnalitickaKarticaStampaDialog, "#dialog-proment-analitika-stampa", "innerHTML", "GET", "", hxValsAnalitickihKonta, true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput("search-input", translator, prometURLAnKonta, fmt.Sprintf("#%s", prometTableID), hxValsAnalitickihKonta)
		err := tmpl_fin.PrometAnalitickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, prometKontaAnalitickiDodatniParametriURL, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			Konto:      c.Query("konto"),
			Sifra:      c.Query("sifra"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			SearchText: c.Query("query"),
			ReportTip:  "prometankonta",
		}
		//validacija input parametre:
		fieldParameters := []string{"konto", "sifra", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetAnkontaTableFields(), "", prometURLAnKonta, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLAnKonta, false, false, false)
		tbl.Pagination.HxVals = hxValsAnalitickihKonta

		err := h.service.GetPrometAnalitickihKonta(ctx, &tbl, true, pageSize, page, false, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometAnalitickihKonta(ctx, &tbl, false, pageSize, page, false, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}

func (h *PrometHandler) PrometAnalitickihKontaStampaDialog(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	konto := c.Query("konto")
	sifra := c.Query("sifra")
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")

	dialog := domain.Dialog{
		Id:          "dialog-proment-analitika-stampa-dlg",
		Title:       "Štampa - izbor parametara",
		OkText:      "Štampaj",
		CancelText:  "Odustani",
		HxActionURL: prometKontaAnalitickiDodatniParametriURL,
		HxTarget:    "#dialog-proment-analitika-stampa",
		HxSwap:      "innerHTML",
	}
	prometParams := domain.PrometParam{
		Konto:    konto,
		Sifra:    sifra,
		OdDatuma: odDatuma,
		DoDatuma: doDatuma,
	}
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "print", prometURLAnalitickaKarticaStampa, "GET", true, common.ClassPrintButton, stampaAnKarticaFileds)
	btnClose.IdDialog = "dialog-proment-analitika-stampa-dlg"
	btnCancel.IdDialog = "dialog-proment-analitika-stampa-dlg"
	tmpl_fin.PrometAnalitickihKontaDialog(dialog, btnPrint, btnCancel, btnClose, prometParams, userSession.SelectedGod, translator).Render(ctx, c.Writer)
}

// Handler for additional parameters for printing Analiticka Kartica
func (h *PrometHandler) PrometAnalitickihKontaDodatniParametri(c *gin.Context) {
	dialog := domain.Dialog{
		Id:          "dialog-proment-analitika-stampa-dlg",
		Title:       "Izbor parametara",
		OkText:      "Štampaj",
		CancelText:  "Odustani",
		HxActionURL: prometKontaAnalitickiDodatniParametriURL,
		HxTarget:    "#dialog-proment-analitika-stampa",
		HxSwap:      "innerHTML",
	}
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnSelect := common.SetPrintButton("stampa-btn", "Izaberi", "save", prometURLAnalitickaKarticaStampa, "GET", true, common.ClassPrintButton, stampaAnKarticaFileds)
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnClose.IdDialog = "dialog-proment-analitika-stampa-dlg"
	btnCancel.IdDialog = "dialog-proment-analitika-stampa-dlg"
	tmpl_fin.PrometAnalitickihKontaDodatniParametri(dialog, btnClose, btnSelect, btnCancel, domain.PrometParam{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

// Handler for printing Analiticka Kartica with selected parameters
func (h *PrometHandler) PrometAnalitickihKontaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	//translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	prometStampaParams := getAnKarticaPrameterValues(c)

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetPrometAnalitickaKarticaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetPrometAnalitickaKarticaStampa(ctx, &tbl, prometStampaParams)
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
		ReportName:  "Analitička kartica",
		ParameterItems: map[string]domain.ParameterItem{
			"OdKonta":     {Name: "Od konta", Value: prometStampaParams.OdKonta},
			"DoKonta":     {Name: "Do konta", Value: prometStampaParams.DoKonta},
			"OdSifre":     {Name: "Od šifre", Value: prometStampaParams.OdSifre},
			"DoSifre":     {Name: "Do šifre", Value: prometStampaParams.DoSifre},
			"OdDatuma":    {Name: "Od datuma", Value: prometStampaParams.OdDatuma},
			"DoDatuma":    {Name: "Do datuma", Value: prometStampaParams.DoDatuma},
			"StanjeNaDan": {Name: "Stanje na dan", Value: prometStampaParams.DatumStampe},
		},
	}

	tmpl_rep_fin.PrometAnalitickaKarticaStampa(repParams, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

// Handler for printing Analiticka Kartica Po MI
func (h *PrometHandler) PrometAnalitickihKontaPoMIStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	prometStampaParams := getAnKarticaMIParameterValues(c)

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetPrometAnalitickaKarticaPoMIStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetPrometAnalitickaKarticaPoMIStampa(ctx, &tbl, prometStampaParams)
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
		ReportName:  "Analitička kartica po mestima isporuke",
		ParameterItems: map[string]domain.ParameterItem{
			"Konto":    {Name: "Konto", Value: c.Query("konto")},
			"Sifra":    {Name: "Šifra", Value: c.Query("sifra")},
			"OdDatuma": {Name: "Od datuma", Value: c.Query("oddatuma")},
			"DoDatuma": {Name: "Do datuma", Value: c.Query("dodatuma")},
			"OdMI":     {Name: "Od mesta isporuke", Value: c.Query("odmi")},
			"DoMI":     {Name: "Do mesta isporuke", Value: c.Query("domi")},
		},
	}
	tmpl_rep_fin.PrometAnalitickaKarticaPoMIStampa(repParams, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func getAnKarticaMIParameterValues(c *gin.Context) domain.PrometStampaParam {
	prometStampaParams := domain.PrometStampaParam{}
	for field := range strings.SplitSeq(stampaAnKarticaMIFields, ",") {
		if strings.HasPrefix(field, "chk") || strings.HasPrefix(field, "cbx") {
			value := c.Query(field) == "true"
			setFieldValue(&prometStampaParams, field, value)
		} else {
			value := c.Query(field)
			setFieldValue(&prometStampaParams, field, value)
		}
	}
	// konto from browse tab maps to both OdKonta and DoKonta
	if prometStampaParams.DoKonta == "" && prometStampaParams.OdKonta != "" {
		prometStampaParams.DoKonta = prometStampaParams.OdKonta
	}
	// sifra from browse tab maps to both OdSifre and DoSifre
	if sifra := c.Query("sifra"); sifra != "" {
		prometStampaParams.OdSifre = sifra
		prometStampaParams.DoSifre = sifra
	}
	return prometStampaParams
}

func (h *PrometHandler) PrometAnalitickihKontaPoMI(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	//if the call come from menu click or tab click then render the page with parameters and empty table
	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnKontaMiTableFields(), "", prometURLAnKontaMi, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, prometContentTitle, prometURLAnKontaMi, false, false, false)
	common.SetActiveTab(&h.tabData, "analitickakontami")
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLAnKontaMi, "#promettable", "innerHTML", "GET", "", hxValsMI, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetPrintButton("print-btn", "Štampa", "print", prometURLAnalitickaKarticaPoMIStampa, "GET", true, common.ClassPrintButton, stampaAnKarticaMIFields)
		searchInput := common.CreateSearchInput("search-input", translator, prometURLAnKontaMi, fmt.Sprintf("#%s", prometTableID), hxValsMI)
		err := tmpl_fin.AnalitickaKarticaPoMI(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			Konto:      c.Query("konto"),
			Sifra:      c.Query("sifra"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			OdMI:       c.Query("odmi"),
			DoMI:       c.Query("domi"),
			SearchText: c.Query("query"),
			ReportTip:  "prometankontami",
		}
		//validacija input parametre:
		fieldParameters := []string{"konto", "sifra", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetAnKontaMiTableFields(), "", prometURLAnKontaMi, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLAnKontaMi, false, false, false)
		tbl.Pagination.HxVals = hxValsMI

		err := h.service.GetPrometAnalitickihKonta(ctx, &tbl, true, pageSize, page, true, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometAnalitickihKonta(ctx, &tbl, false, pageSize, page, true, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}
func (h *PrometHandler) PrometDeviznihAnalitickihKonta(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLDeviznaKonta, "#promettable", "innerHTML", "GET", "", hxValsDeviznaKonta, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnDeviznaKontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", "", false, false, false)

		common.SetActiveTab(&h.tabData, "deviznihanalitickihkonta")
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		err := tmpl_fin.PrometDeviznihAnalitickihKonta(h.tabData, tbl, tbl, btnPrint, btnObrada, domain.TotalValues{}, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			Konto:      c.Query("konto"),
			Sifra:      c.Query("sifra"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			SearchText: c.Query("query"),
			ReportTip:  "deviznihanalitickihkonta",
		}
		//validacija input parametre:
		fieldParameters := []string{"konto", "sifra", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetAnkontaTableFields(), "", prometURLDeviznaKonta, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLDeviznaKonta, false, false, false)
		tbl.Pagination.HxVals = hxValsDeviznaKonta

		err := h.service.GetPrometAnalitickihKonta(ctx, &tbl, true, pageSize, page, false, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometAnalitickihKonta(ctx, &tbl, false, pageSize, page, false, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}
func (h *PrometHandler) PrometSubsintetickihKonta(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLSubsintetika, "#promettable", "innerHTML", "GET", "", hxValsSubsintetika, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", prometURLSubsintetikaStampaDialog, "#dialog-promet-subsint-stampa", "innerHTML", "GET", "", hxValsSubsintetika, true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput("search-input", translator, prometURLSubsintetika, fmt.Sprintf("#%s", prometTableID), hxValsSubsintetika)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSubsintetickihKontaTableFields(), "", prometURLSubsintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLSubsintetika, false, false, false)
		common.SetActiveTab(&h.tabData, "subsintetickakonta")
		err := tmpl_fin.PrometSubsintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			OdKonta:    c.Query("odkonta"),
			DoKonta:    c.Query("dokonta"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			SearchText: c.Query("query"),
			ReportTip:  "subsintetickihkonta",
		}
		//validacija input parametre:
		fieldParameters := []string{"odkonta", "dokonta", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetSubsintetickihKontaTableFields(), "", prometURLSubsintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLSubsintetika, false, false, false)
		tbl.Pagination.HxVals = hxValsSubsintetika

		err := h.service.GetPrometSubsintetickihKonta(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometSubsintetickihKonta(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}
func (h *PrometHandler) PrometSintetickihKonta(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLSintetika, "#promettable", "innerHTML", "GET", "", hxValsSintetika, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", prometURLSintetikaStampaDialog, "#dialog-promet-sint-stampa", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput("search-input", translator, prometURLSintetika, fmt.Sprintf("#%s", prometTableID), hxValsSintetika)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSintetickihKontaTableFields(), "", prometURLSintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLSintetika, false, false, false)
		common.SetActiveTab(&h.tabData, "sintetickakonta")
		err := tmpl_fin.PrometSintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			Konto:      c.Query("konto"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			SearchText: c.Query("query"),
			ReportTip:  "sintetickakonta",
		}
		//validacija input parametre:
		fieldParameters := []string{"konto", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetSintetickihKontaTableFields(), "", prometURLSintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLSintetika, false, false, false)
		tbl.Pagination.HxVals = hxValsSintetika

		err := h.service.GetPrometSintetickihKonta(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometSintetickihKonta(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}
func (h *PrometHandler) PrometSintetickihKontaStampaDialog(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	translator := i18n.GetInstance()
	konto := c.Query("konto")
	prometParams := domain.PrometParam{Konto: konto}
	dialog := domain.Dialog{Id: "dialog-sint-stampa-dlg", Title: "Sintetička kartica - parametri štampe"}
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnCancel := common.SetButton("cancel-btn", "Nazad", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampaj", "print", prometURLSintetikaStampa, "GET", true, common.ClassPrintButton, stampaSintetikaFields)
	btnClose.IdDialog = "dialog-sint-stampa-dlg"
	btnCancel.IdDialog = "dialog-sint-stampa-dlg"
	tmpl_fin.PrometSintetickihKontaDialog(dialog, btnPrint, btnCancel, btnClose, prometParams, userSession.SelectedGod, translator).Render(ctx, c.Writer)
}

func (h *PrometHandler) PrometSintetickihKontaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	stampParams := getKarticaSintParameterValues(c)

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetPrometSintetickihKontaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetPrometSintetickihKontaStampa(ctx, &tbl, stampParams)
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
		ReportName:  "Sintetička kartica",
		ParameterItems: map[string]domain.ParameterItem{
			"OdKonta":     {Name: "Od konta", Value: stampParams.OdKonta},
			"DoKonta":     {Name: "Do konta", Value: stampParams.DoKonta},
			"StanjeNaDan": {Name: "Stanje kartica na dan", Value: stampParams.DatumStampe},
		},
	}
	tmpl_rep_fin.KarticaSintetickihKontaStampa(repParams, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

// Handler for Promet Kartica Sintetickih Konta
func (h *PrometHandler) PrometKarticaSintetickihKonta(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLKarticaSintetika, "#promettable", "innerHTML", "GET", "", hxValsKarticaSintetika, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetPrintButton("print-btn", "Štampa", "stampa", prometURLKarticaSintetikaStampa, "GET", true, common.ClassPrintButton, stampaKarticaSintKontaFields)
		searchInput := common.CreateSearchInput("search-input", translator, prometURLKarticaSintetika, fmt.Sprintf("#%s", prometTableID), hxValsKarticaSintetika)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKarticaSintetikaTableFields(), "", prometURLKarticaSintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLKarticaSintetika, false, false, false)
		common.SetActiveTab(&h.tabData, "karticasintetickihkonta")
		err := tmpl_fin.KarticaSintetickiKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			Konto:      c.Query("konto"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			Analitika:  c.Query("analitika"),
			SearchText: c.Query("query"),
			ReportTip:  "karticasintetickihkonta",
		}
		//validacija input parametre:
		fieldParameters := []string{"konto", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetKarticaSintetikaTableFields(), "", prometURLKarticaSintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLKarticaSintetika, false, false, false)
		tbl.Pagination.HxVals = hxValsKarticaSintetika

		err := h.service.GetPrometKarticaSintetickihKonta(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometKarticaSintetickihKonta(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}

func (h *PrometHandler) PrometKarticaSintetickihKontaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	stampParams := getKarticaSintKontaParameterValues(c)

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetPrometKarticaSintKontaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetPrometKarticaSintKontaStampa(ctx, &tbl, stampParams)
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
		ReportName:  "Kartica sintetičkog konta",
		ParameterItems: map[string]domain.ParameterItem{
			"Konto":    {Name: "Konto", Value: stampParams.OdKonta + "  " + tbl.ContentTitle},
			"OdDatuma": {Name: "Od datuma", Value: stampParams.OdDatuma},
			"DoDatuma": {Name: "Do datuma", Value: stampParams.DoDatuma},
		},
	}
	tmpl_rep_fin.KarticaSintKontaStampa(repParams, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func getKarticaSintKontaParameterValues(c *gin.Context) domain.PrometStampaParam {
	return domain.PrometStampaParam{
		OdKonta:   c.Query("konto"),
		DoKonta:   c.Query("konto"),
		OdDatuma:  c.Query("oddatuma"),
		DoDatuma:  c.Query("dodatuma"),
		Analitika: c.Query("analitika") == "true",
	}
}

func getKarticaSintParameterValues(c *gin.Context) domain.PrometStampaParam {
	params := domain.PrometStampaParam{}
	for field := range strings.SplitSeq(stampaSintetikaFields, ",") {
		if strings.HasPrefix(field, "chk") || strings.HasPrefix(field, "cbx") {
			value := c.Query(field) == "true"
			setFieldValue(&params, field, value)
		} else {
			value := c.Query(field)
			setFieldValue(&params, field, value)
		}
	}
	// single konto field maps to both OdKonta and DoKonta
	params.DoKonta = params.OdKonta
	return params
}

func (h *PrometHandler) PrometSubsintetickaKontaPoVRD(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLSubsintetikaVrd, "#promettable", "innerHTML", "GET", "", hxValsSubsintetikaVrd, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetPrintButton("print-btn", "Štampa", "stampa", prometURLSubsintetikaVrdStampa, "GET", true, common.ClassPrintButton, stampaSubsintetikaVrdFields)
		searchInput := common.CreateSearchInput("search-input", translator, prometURLSubsintetikaVrd, fmt.Sprintf("#%s", prometTableID), hxValsSubsintetikaVrd)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSubsintetikaVrdTableFields(), "", prometURLSubsintetikaVrd, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLSubsintetikaVrd, false, false, false)
		common.SetActiveTab(&h.tabData, "subsintetickakontapovrd")
		err := tmpl_fin.PrometKontaPoVRD(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			Konto:      c.Query("konto"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			SearchText: c.Query("query"),
			ReportTip:  "subsintetickihkontapovrd",
		}

		//validacija input parametre:
		fieldParameters := []string{"konto", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetSubsintetikaVrdTableFields(), "", prometURLSubsintetikaVrd, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLSubsintetikaVrd, false, false, false)
		tbl.Pagination.HxVals = hxValsSubsintetikaVrd

		err := h.service.GetPrometSubsintetikaVrd(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometSubsintetikaVrd(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}

}

func (h *PrometHandler) PrometSubsintetickaKontaPoVRDStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	params := domain.PrometParam{
		Konto:    c.Query("konto"),
		OdDatuma: c.Query("oddatuma"),
		DoDatuma: c.Query("dodatuma"),
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetPrometSubsintetikaVrdStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetPrometSubsintetikaVrd(ctx, &tbl, false, 0, 0, params, common.TipStampePrint)
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
		ReportName:  "Promet subsintetike po VRD",
		ParameterItems: map[string]domain.ParameterItem{
			"Konto":    {Name: "Konto", Value: params.Konto + "  " + tbl.ContentTitle},
			"OdDatuma": {Name: "Od datuma", Value: params.OdDatuma},
			"DoDatuma": {Name: "Do datuma", Value: params.DoDatuma},
		},
	}
	tmpl_rep_fin.PrometSubsintetikaVrdStampa(repParams, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func (h *PrometHandler) PrometKontaAnaliticki(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLKontaAnaliticki, "#promettable", "innerHTML", "GET", "", hxValsKontaAnaliticki, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#dialog-proment-analitika-stampa", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput("search-input", translator, prometURLKontaAnaliticki, fmt.Sprintf("#%s", prometTableID), hxValsKontaAnaliticki)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKontaAnalitickiTableFields(), "", prometURLKontaAnaliticki, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLKontaAnaliticki, false, false, false)
		common.SetActiveTab(&h.tabData, "kontaanaliticki")
		err := tmpl_fin.PrometKontaAnaliticki(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		params := domain.PrometParam{
			Konto:      c.Query("konto"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			OdSifre:    c.Query("odsifre"),
			DoSifre:    c.Query("dosifre"),
			SearchText: c.Query("query"),
			ReportTip:  "kontaanaliticki",
		}
		//validacija input parametre:
		fieldParameters := []string{"konto", "oddatuma", "dodatuma", "odsifre", "dosifre"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetKontaAnalitickiTableFields(), "", prometURLKontaAnaliticki, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", prometURLKontaAnaliticki, false, false, false)
		tbl.Pagination.HxVals = hxValsKontaAnaliticki

		err := h.service.GetPrometKontaAnaliticki(ctx, &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometKontaAnaliticki(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}

func (h *PrometHandler) TotalValues(c *gin.Context) {

	// Get totals data
	params := domain.PrometParam{
		Konto:      c.Query("konto"),
		Sifra:      c.Query("sifra"),
		OdKonta:    c.Query("odkonta"),
		DoKonta:    c.Query("dokonta"),
		OdDatuma:   c.Query("oddatuma"),
		DoDatuma:   c.Query("dodatuma"),
		OdSifre:    c.Query("odsifre"),
		DoSifre:    c.Query("dosifre"),
		OdMI:       c.Query("odmi"),
		DoMI:       c.Query("domi"),
		SearchText: c.Query("query"),
		ReportTip:  c.Query("tabname"),
		Analitika:  c.Query("analitika"),
	}
	response, err := h.service.GetPrometTotals(c.Request.Context(), params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}

	err = tmpl_fin.PrometTotalValues(response.Totals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}

}

func (h *PrometHandler) PrometSubsintetickihKontaStampaDialog(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	odKonta := c.Query("odkonta")
	doKonta := c.Query("dokonta")

	dialog := domain.Dialog{
		Id:    "dialog-promet-subsint-stampa-dlg",
		Title: "Štampa - izbor parametara",
	}
	prometParams := domain.PrometParam{
		OdKonta: odKonta,
		DoKonta: doKonta,
	}
	btnClose := common.SetButton("close-btn", "", "close", "", "", "", "", "", "", true, common.ClassDialogCloseButton, "")
	btnCancel := common.SetButton("cancel-btn", "Odustani", "cancel", "", "", "", "", "", "", true, common.ClassOdustaniButton, "")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampaj", "print", prometURLSubsintetikaStampa, "GET", true, common.ClassPrintButton, stampaSubsKarticaFields)
	btnClose.IdDialog = "dialog-promet-subsint-stampa-dlg"
	btnCancel.IdDialog = "dialog-promet-subsint-stampa-dlg"
	tmpl_fin.PrometSubsintetickihKontaDialog(dialog, btnPrint, btnCancel, btnClose, prometParams, userSession.SelectedGod, translator).Render(ctx, c.Writer)
}

func (h *PrometHandler) PrometSubsintetickihKontaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	stampParams := getSubsKarticaParameterValues(c)

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetPrometSubsintetickihKontaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	err = h.service.GetPrometSubsintetickihKontaStampa(ctx, &tbl, stampParams)
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
		ReportName:  "Promet subsintetičkih konta",
		ParameterItems: map[string]domain.ParameterItem{
			"OdKonta":     {Name: "Od konta", Value: stampParams.OdKonta},
			"DoKonta":     {Name: "Do konta", Value: stampParams.DoKonta},
			"OdDatuma":    {Name: "Od datuma", Value: stampParams.OdDatuma},
			"DoDatuma":    {Name: "Do datuma", Value: stampParams.DoDatuma},
			"StanjeNaDan": {Name: "Stanje na dan", Value: stampParams.DatumStampe},
		},
	}

	tmpl_rep_fin.PrometSubsintetickihKontaKarticaStampa(repParams, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func getSubsKarticaParameterValues(c *gin.Context) domain.PrometStampaParam {
	params := domain.PrometStampaParam{}
	for field := range strings.SplitSeq(stampaSubsKarticaFields, ",") {
		if strings.HasPrefix(field, "chk") || strings.HasPrefix(field, "cbx") {
			value := c.Query(field) == "true"
			setFieldValue(&params, field, value)
		} else {
			value := c.Query(field)
			setFieldValue(&params, field, value)
		}
	}
	return params
}

func getAnKarticaPrameterValues(c *gin.Context) domain.PrometStampaParam {
	prometStampaParams := domain.PrometStampaParam{}
	for field := range strings.SplitSeq(stampaAnKarticaFileds, ",") {
		// Process each field
		if strings.HasPrefix(field, "chk") || strings.HasPrefix(field, "cbx") {
			value := c.Query(field) == "true"
			setFieldValue(&prometStampaParams, field, value)
		} else {
			value := c.Query(field)
			setFieldValue(&prometStampaParams, field, value)
		}
	}
	return prometStampaParams
}

// paramFieldMap maps lowercase query param names to PrometStampaParam struct field names.
var paramFieldMap = map[string]string{
	"saldapomesecima":      "SaldaPoMesecima",
	"ukupansaldo":          "UkupanSaldo",
	"odkonta":              "OdKonta",
	"dokonta":              "DoKonta",
	"odsifre":              "OdSifre",
	"dosifre":              "DoSifre",
	"datumstampe":          "DatumStampe",
	"karticadeviznihkonta": "KarticaDeviznihKonta",
	"karticasakolicinom":   "KarticaSaKolicinom",
	"chkpogrupianaloga":    "ChkPoGrupiNaloga",
	"odvrstenaloga":        "OdVrsteNaloga",
	"dovrstenaloga":        "DoVrsteNaloga",
	"chkpobrojunaloga":     "ChkPoBrojuNaloga",
	"odbrojenaloga":        "OdBrojaNaloga",
	"dobrojenaloga":        "DoBrojaNaloga",
	"chkpodatumunaloga":    "ChkPoDatumuNaloga",
	"oddatumanaloga":       "OdDatumaNaloga",
	"dodatumanaloga":       "DoDatumaNaloga",
	"chkpodatumuobrade":    "ChkPoDatumuObrade",
	"oddatumaobrade":       "OdDatumaObrade",
	"dodatumaobrade":       "DoDatumaObrade",
	"chkpovrstidokumenta":  "ChkPoVrstiDokumenta",
	"vrstidokumenta":       "VrsteDokumenta",
	"chkpobrojudokumenta":  "ChkPoBrojuDokumenta",
	"odbrojedokumenta":     "OdBrojaDokumenta",
	"dobrojedokumenta":     "DoBrojaDokumenta",
	"chkpodatumdokumenta":  "ChkPoDatumuDokumenta",
	"oddatumadokumenta":    "OdDatumaDokumenta",
	"dodatumadokumenta":    "DoDatumaDokumenta",
	"chkpoiznosu":          "ChkPoIznosu",
	"odiznosa":             "OdIznosa",
	"doiznosa":             "DoIznosa",
	"chkponacinknjizenja":  "ChkPoNacinuKnjizenja",
	"nacinknjizenja":       "NacinKnjizenja",
	"stampadokumenta":      "StampaDokumenta",
	"oddatuma":             "OdDatuma",
	"dodatuma":             "DoDatuma",
	"konto":                "OdKonta",
	"odmi":                 "OdMI",
	"domi":                 "DoMI",
}

// setFieldValue sets the value of a field in PrometStampaParam based on its type (bool or string)
func setFieldValue(prometStampaParams *domain.PrometStampaParam, field string, value any) {
	structField, ok := paramFieldMap[field]
	if !ok {
		return
	}
	v := reflect.ValueOf(prometStampaParams).Elem().FieldByName(structField)
	if !v.IsValid() || !v.CanSet() {
		return
	}
	switch val := value.(type) {
	case bool:
		if v.Kind() == reflect.Bool {
			v.SetBool(val)
		}
	case string:
		if v.Kind() == reflect.String {
			v.SetString(val)
		} else if v.Kind() == reflect.Bool {
			v.SetBool(val == "true" || val == "on")
		}
	}
}

func (h *PrometHandler) AddRoutes(r *gin.Engine) {
	// Create API group with prefix
	//api := r.Group(prometURLPrefix)
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for promet
	r.GET("/api/promet", h.PrometMain)
	r.GET("/api/promet/analitickakonta", h.PrometAnalitickihKonta)
	r.GET("/api/promet/analitickakonta/stampadialog", h.PrometAnalitickihKontaStampaDialog)
	r.GET("/api/promet/analitickakonta/stampa", h.PrometAnalitickihKontaStampa)
	r.GET("/api/promet/analitickakonta/dodatniparametri", h.PrometAnalitickihKontaDodatniParametri)
	r.GET("/api/promet/analitickakontami", h.PrometAnalitickihKontaPoMI)
	r.GET("/api/promet/analitickakontami/stampa", h.PrometAnalitickihKontaPoMIStampa)
	r.GET("/api/promet/deviznihanalitickihkonta", h.PrometDeviznihAnalitickihKonta)
	r.GET("/api/promet/subsintetickakonta", h.PrometSubsintetickihKonta)
	r.GET("/api/promet/subsintetickakonta/stampadialog", h.PrometSubsintetickihKontaStampaDialog)
	r.GET("/api/promet/subsintetickakonta/stampa", h.PrometSubsintetickihKontaStampa)
	r.GET("/api/promet/sintetickakonta", h.PrometSintetickihKonta)
	r.GET("/api/promet/sintetickihkonta/stampadialog", h.PrometSintetickihKontaStampaDialog)
	r.GET("/api/promet/sintetickihkonta/stampa", h.PrometSintetickihKontaStampa)
	r.GET("/api/promet/karticasintetickihkonta", h.PrometKarticaSintetickihKonta)
	r.GET("/api/promet/karticasintetickihkonta/stampa", h.PrometKarticaSintetickihKontaStampa)
	r.GET("/api/promet/subsintetickakontapovrd/stampa", h.PrometSubsintetickaKontaPoVRDStampa)
	r.GET("/api/promet/subsintetickakontapovrd", h.PrometSubsintetickaKontaPoVRD)
	r.GET("/api/promet/kontaanaliticki", h.PrometKontaAnaliticki)
	r.GET("/api/promet/totalvalues", h.TotalValues)
	r.GET("/api/promet/searchbutton", utils.SearchButtonDialog)

}

func GetTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "prometankonta", Label: "Analitičkih konta", HXRequestUrl: fmt.Sprintf("%sanalitickakonta", prometURLPrefix), IsActive: true, Name: "analitickakonta"},
			{ID: "prometankontami", Label: "An. konta po MI", HXRequestUrl: fmt.Sprintf("%sanalitickakontami", prometURLPrefix), IsActive: false, Name: "analitickakontami"},
			{ID: "deviznihanalitickihkonta", Label: "Deviznih an. konta", HXRequestUrl: fmt.Sprintf("%sdeviznihanalitickihkonta", prometURLPrefix), IsActive: false, Name: "deviznihanalitickihkonta"},
			{ID: "subsintetickakonta", Label: "Subsintetičkih konta", HXRequestUrl: fmt.Sprintf("%ssubsintetickakonta", prometURLPrefix), IsActive: false, Name: "subsintetickakonta"},
			{ID: "sintetickakonta", Label: "Sintetičkih konta", HXRequestUrl: fmt.Sprintf("%ssintetickakonta", prometURLPrefix), IsActive: false, Name: "sintetickakonta"},
			{ID: "karticasintetickihkonta", Label: "Kartica sintetičkih konta", HXRequestUrl: fmt.Sprintf("%skarticasintetickihkonta", prometURLPrefix), IsActive: false, Name: "karticasintetickihkonta"},
			{ID: "subsintetickakontapovrd", Label: "Konta po VRD", HXRequestUrl: fmt.Sprintf("%ssubsintetickakontapovrd", prometURLPrefix), IsActive: false, Name: "subsintetickakontapovrd"},
			{ID: "kontaanaliticki", Label: "Konta analitički", HXRequestUrl: fmt.Sprintf("%skontaanaliticki", prometURLPrefix), IsActive: false, Name: "kontaanaliticki"},
		},
	}
}
