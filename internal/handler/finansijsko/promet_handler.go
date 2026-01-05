package finansijsko

import (
	"fmt"
	"net/http"

	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/i18n"
	"helia/internal/middleware"
	"helia/internal/service"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	prometContentTitle string = "PROMET"
	prometTableID      string = "promettable"
	prometURLPrefix    string = "/api/promet/"
	prometURLGetAll    string = "/api/promet/all"
)

type PrometHandler struct {
	tabData domain.TabData
	service *service.PrometResource
}

const (
	hxVals = `js:{
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
)

func NewPrometHandler(service *service.PrometResource) *PrometHandler {
	handler := &PrometHandler{}
	handler.tabData = GetTabData()
	handler.service = service
	return handler
}

func (h *PrometHandler) PrometMain(c *gin.Context) {
	// Create configuration

	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/analitickakonta", "#promettable", "innerHTML", "GET", "#konto, #sifra, #oddatuma, #dodatuma", hxVals, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	tbl.FuncClick = "selectRow"                             // naziv js function for Click
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)" // naziv js function for dblClick
	err := tmpl_fin.PrometMain(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *PrometHandler) PrometAnalitickihKonta(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/analitickakonta", "#promettable", "innerHTML", "GET", "", hxVals, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0)
		tbl.ShowActions = false
		h.tabData = setActiveTab(h.tabData, "analitickakonta")
		err := tmpl_fin.PrometAnalitickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		//validacija input parametre:
		fieldParameters := []string{"konto", "sifra", "oddatuma", "dodatuma"}
		fieldsError := h.service.CheckPrometParameters(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c)
		response, err := h.service.GetPrometAnalitickihKonta(c, true, 0, 0)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		totalRecords := response.TotalRecords
		totalPages := (totalRecords + pageSize - 1) / pageSize
		// Get paginated data
		response, err = h.service.GetPrometAnalitickihKonta(c, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetAnkontaTableFields(), "", "/api/promet/analitickakonta", pageSize, page, totalPages, totalRecords)
		tbl.ShowActions = false
		tbl.ShowPagination = true
		tbl.Pagination.HxVals = hxVals
		// Prepare TableData for UI
		tblRows, err := common.SetTableRows(&tbl, response.Data, h.service.GetAnkontaTableFields(), "idfpro", "", h.service.GetFieldCache())
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgFailedToSetTableRow)
			return
		}
		tbl.Rows = tblRows.Rows
		tbl.BtnAdd = domain.Button{IsVisible: false}   // Hide Add button in this view
		tbl.BtnPrint = domain.Button{IsVisible: false} // Hide Print button in this view

		utils.RenderContent(c, tbl)
	}
}

func (h *PrometHandler) PrometAnalitickihKontaPoMI(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/analitickakontami", "#promettable", "innerHTML", "GET", "", hxValsMI, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnKontaMiTableFields(), "", "", 0, 0, 0, 0)
		tbl.ShowActions = false
		h.tabData = setActiveTab(h.tabData, "analitickakontami")
		err := tmpl_fin.AnalitickaKarticaPoMI(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		//validacija input parametre:
		fieldParameters := []string{"konto", "sifra", "oddatuma", "dodatuma", "odmi", "domi"}
		fieldsError := h.service.CheckPrometParameters(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c)
		response, err := h.service.GetPrometAnalitickihKontaMi(c, true, 0, 0)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		totalRecords := response.TotalRecords
		totalPages := (totalRecords + pageSize - 1) / pageSize
		// Get paginated data
		response, err = h.service.GetPrometAnalitickihKontaMi(c, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetAnkontaTableFields(), "", "/api/promet/analitickakontami", pageSize, page, totalPages, totalRecords)
		tbl.ShowActions = false
		tbl.ShowPagination = true
		tbl.Pagination.HxVals = hxVals
		// Prepare TableData for UI
		tblRows, err := common.SetTableRows(&tbl, response.Data, h.service.GetAnkontaTableFields(), "idfpro", "", h.service.GetFieldCache())
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgFailedToSetTableRow)
			return
		}
		tbl.Rows = tblRows.Rows
		tbl.BtnAdd = domain.Button{IsVisible: false}   // Hide Add button in this view
		tbl.BtnPrint = domain.Button{IsVisible: false} // Hide Print button in this view

		utils.RenderContent(c, tbl)
	}
}
func (h *PrometHandler) PrometDeviznihAnalitickihKonta(c *gin.Context) {
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/deviznihanalitickihkonta", "#promettable", "innerHTML", "GET", "", hxVals, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnDeviznaKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false

	h.tabData = setActiveTab(h.tabData, "deviznihanalitickihkonta")
	err := tmpl_fin.PrometDeviznihAnalitickihKonta(h.tabData, tbl, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	//validacija input parametre:
	fieldParameters := []string{"konto", "sifra", "oddatuma", "dodatuma"}
	fieldsError := h.service.CheckPrometParameters(c, fieldParameters)
	if len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
		return
	}

	page, pageSize := common.GetPageAndPageSizeFromRequest(c)
	response, err := h.service.GetPrometAnalitickihKonta(c, true, 0, 0)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	totalRecords := response.TotalRecords
	totalPages := (totalRecords + pageSize - 1) / pageSize
	// Get paginated data
	response, err = h.service.GetPrometAnalitickihKonta(c, false, pageSize, page)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tbl = common.SetTableBasicData("", prometTableID, h.service.GetAnkontaTableFields(), "", "/api/promet/deviznihanalitickihkonta", pageSize, page, totalPages, totalRecords)
	tbl.ShowActions = false
	tbl.ShowPagination = true
	tbl.Pagination.HxVals = hxVals
	// Prepare TableData for UI
	tblRows, err := common.SetTableRows(&tbl, response.Data, h.service.GetAnkontaTableFields(), "idfpro", "", h.service.GetFieldCache())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgFailedToSetTableRow)
		return
	}
	tbl.Rows = tblRows.Rows
	tbl.BtnAdd = domain.Button{IsVisible: false}   // Hide Add button in this view
	tbl.BtnPrint = domain.Button{IsVisible: false} // Hide Print button in this view

	utils.RenderContent(c, tbl)

}

func (h *PrometHandler) PrometSubsintetickihKonta(c *gin.Context) {
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/subsintetickakonta", "#promettable", "innerHTML", "POST", "", "", true, common.ClassObradaButton, "handleDialogResponse('tab4')")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSubsintetickihKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false

	h.tabData = setActiveTab(h.tabData, "subsintetickakonta")
	err := tmpl_fin.PrometSubsintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}
func (h *PrometHandler) PrometSintetickihKonta(c *gin.Context) {
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/sintetickakonta", "#promet-table", "innerHTML", "POST", "", "", true, common.ClassObradaButton, "handleDialogResponse('tab5')")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSintetickihKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false

	h.tabData = setActiveTab(h.tabData, "sintetickakonta")
	err := tmpl_fin.PrometSintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}
func (h *PrometHandler) KarticaSintetickihKonta(c *gin.Context) {
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/karticasintetickihkonta", "#promettable", "innerHTML", "POST", "", "", true, common.ClassObradaButton, "handleDialogResponse('tab6')")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKarticaSintetickihKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false

	h.tabData = setActiveTab(h.tabData, "karticasintetickihkonta")
	err := tmpl_fin.KarticaSintetickiKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *PrometHandler) PrometKontaAnaliticki(c *gin.Context) {
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/kontaanaliticki", "#promettable", "innerHTML", "POST", "", "", true, common.ClassObradaButton, "handleDialogResponse('tab7')")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKontaAnalitickiTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false

	h.tabData = setActiveTab(h.tabData, "kontaanaliticki") // Activate the "Promet konta analitički" tab
	err := tmpl_fin.PrometKontaAnaliticki(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *PrometHandler) PrometTotalValues(c *gin.Context) {

	// Get totals data
	response, err := h.service.GetPrometTotals(c)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}

	err = tmpl_fin.TotalValues(response.Totals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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
	r.GET("/api/promet/karticasintetickihkonta", h.KarticaSintetickihKonta)
	r.GET("/api/promet/kontaanaliticki", h.PrometKontaAnaliticki)
	r.GET("/api/promet/totalvalues", h.PrometTotalValues)
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
			{ID: "kontapovrd", Label: "Konta po VRD", HXRequestUrl: fmt.Sprintf("%skontapovrd", prometURLPrefix), IsActive: false, Name: "kontapovrd"},
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
