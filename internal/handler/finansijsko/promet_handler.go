package finansijsko

import (
	"fmt"
	"helia/config"
	"net/http"

	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	prometContentTitle        string = "PROMET"
	prometTableID             string = "promettable"
	prometURLPrefix           string = "/api/promet/"
	prometURLGetAll           string = "/api/promet/all"
	prometURLAnKonta          string = "/api/promet/analitickakonta"
	prometURLAnKontaMi        string = "/api/promet/analitickakontami"
	prometURLDeviznaKonta     string = "/api/promet/deviznihanalitickihkonta"
	prometURLSubsintetika     string = "/api/promet/subsintetickakonta"
	prometURLSintetika        string = "/api/promet/sintetickakonta"
	prometURLKarticaSintetika string = "/api/promet/karticasintetickihkonta"
	prometURLSubsintetikaVrd  string = "/api/promet/subsintetickakontapovrd"
	prometURLKontaAnaliticki  string = "/api/promet/kontaanaliticki"
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
            dodatuma: document.getElementById("dodatuma")?.value
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

	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/analitickakonta", "#promettable", "innerHTML", "GET", "#konto, #sifra, #oddatuma, #dodatuma", hxValsAnalitickihKonta, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput(i18n.GetInstance(), prometURLAnKonta, fmt.Sprintf("#%s", prometTableID), hxValsAnalitickihKonta)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, prometContentTitle, "", false, false, false)
	tbl.FuncClick = "selectRow"                             // naziv js function for Click
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)" // naziv js function for dblClick
	err := tmpl_fin.PrometMain(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *PrometHandler) PrometAnalitickihKonta(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLAnKonta, "#promettable", "innerHTML", "GET", "", hxValsAnalitickihKonta, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, prometURLAnKonta, fmt.Sprintf("#%s", prometTableID), hxValsAnalitickihKonta)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, "", false, false, false)
		h.tabData = setActiveTab(h.tabData, "analitickakonta")
		err := tmpl_fin.PrometAnalitickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
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

		err := h.service.GetPrometAnalitickihKonta(c, &tbl, true, pageSize, page, false)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometAnalitickihKonta(c, &tbl, false, pageSize, page, false)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *PrometHandler) PrometAnalitickihKontaPoMI(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnKontaMiTableFields(), "", prometURLAnKontaMi, 0, 0, 0, 0, h.cfg)

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLAnKontaMi, "#promettable", "innerHTML", "GET", "", hxValsMI, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, prometURLAnKontaMi, fmt.Sprintf("#%s", prometTableID), hxValsMI)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		common.SetTableConfig(&tbl, prometContentTitle, prometURLAnKontaMi, false, false, false)
		h.tabData = setActiveTab(h.tabData, "analitickakontami")
		err := tmpl_fin.AnalitickaKarticaPoMI(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
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

		err := h.service.GetPrometAnalitickihKonta(c, &tbl, true, pageSize, page, true)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometAnalitickihKonta(c, &tbl, false, pageSize, page, true)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

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

		h.tabData = setActiveTab(h.tabData, "deviznihanalitickihkonta")
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

		err := h.service.GetPrometAnalitickihKonta(c, &tbl, true, pageSize, page, false)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometAnalitickihKonta(c, &tbl, false, pageSize, page, false)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

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
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, prometURLSubsintetika, fmt.Sprintf("#%s", prometTableID), hxValsSubsintetika)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSubsintetickihKontaTableFields(), "", prometURLSubsintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLSubsintetika, false, false, false)
		h.tabData = setActiveTab(h.tabData, "subsintetickakonta")
		err := tmpl_fin.PrometSubsintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
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

		err := h.service.GetPrometSubsintetickihKonta(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometSubsintetickihKonta(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

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
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, prometURLSintetika, fmt.Sprintf("#%s", prometTableID), hxValsSintetika)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSintetickihKontaTableFields(), "", prometURLSintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLSintetika, false, false, false)
		h.tabData = setActiveTab(h.tabData, "sintetickakonta")
		err := tmpl_fin.PrometSintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
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

		err := h.service.GetPrometSintetickihKonta(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometSintetickihKonta(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}
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
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, prometURLKarticaSintetika, fmt.Sprintf("#%s", prometTableID), hxValsKarticaSintetika)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKarticaSintetikaTableFields(), "", prometURLKarticaSintetika, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLKarticaSintetika, false, false, false)
		h.tabData = setActiveTab(h.tabData, "karticasintetickihkonta")
		err := tmpl_fin.KarticaSintetickiKonta(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
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

		err := h.service.GetPrometKarticaSintetickihKonta(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometKarticaSintetickihKonta(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
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
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, prometURLSubsintetikaVrd, fmt.Sprintf("#%s", prometTableID), hxValsSubsintetikaVrd)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSubsintetikaVrdTableFields(), "", prometURLSubsintetikaVrd, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLSubsintetikaVrd, false, false, false)
		h.tabData = setActiveTab(h.tabData, "subsintetickakontapovrd")
		err := tmpl_fin.PrometKontaPoVRD(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
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

		err := h.service.GetPrometSubsintetikaVrd(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometSubsintetikaVrd(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}

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
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, prometURLKontaAnaliticki, fmt.Sprintf("#%s", prometTableID), hxValsKontaAnaliticki)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKontaAnalitickiTableFields(), "", prometURLKontaAnaliticki, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, prometURLKontaAnaliticki, false, false, false)
		h.tabData = setActiveTab(h.tabData, "kontaanaliticki")
		err := tmpl_fin.PrometKontaAnaliticki(h.tabData, tbl, btnPrint, btnObrada, domain.TotalValues{}, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
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

		err := h.service.GetPrometKontaAnaliticki(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetPrometKontaAnaliticki(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *PrometHandler) TotalValues(c *gin.Context) {

	// Get totals data
	response, err := h.service.GetPrometTotals(c)
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

func (h *PrometHandler) AddRoutes(r *gin.Engine) {
	// Create API group with prefix
	//api := r.Group(prometURLPrefix)
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for promet
	r.GET("/api/promet", h.PrometMain)
	r.GET("/api/promet/analitickakonta", h.PrometAnalitickihKonta)
	r.GET("/api/promet/analitickakontami", h.PrometAnalitickihKontaPoMI)
	r.GET("/api/promet/deviznihanalitickihkonta", h.PrometDeviznihAnalitickihKonta)
	r.GET("/api/promet/subsintetickakonta", h.PrometSubsintetickihKonta)
	r.GET("/api/promet/sintetickakonta", h.PrometSintetickihKonta)
	r.GET("/api/promet/karticasintetickihkonta", h.PrometKarticaSintetickihKonta)
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

func setActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
