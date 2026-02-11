package finansijsko

import (
	"fmt"
	"helia/config"
	"helia/pkg/utils"
	"net/http"

	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	bilansiContentTitle     string = "BILANSI"
	bilansiTableID          string = "bilansitable"
	bilansizakljucniTableID string = "zakljucnitable"
	bilansiURLPrefix        string = "/api/bilansi/"
	bilansiURLZakljucni     string = "/api/bilansi/zakljucni"
	bilansiURLStanja        string = "/api/bilansi/stanja"
	bilansiURLStanjaAlg     string = "/api/bilansi/stanja/stampanje"
	bilansiURLUspeha        string = "/api/bilansi/uspeha"
	bilansiURLUspehaAlg     string = "/api/bilansi/uspeha/stampanje"
)

const (
	hxValsZakljucni = `js:{
            odkonta: document.getElementById("odkonta")?.value,
            dokonta: document.getElementById("dokonta")?.value,
            odsifre: document.getElementById("odsifre")?.value,
            dosifre: document.getElementById("dosifre")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
            tip_zakljucni: document.querySelector('input[name="tip_zakljucni"]:checked')?.value,
			analitickakonta: document.getElementById("analitickakonta")?.checked,
    		klasa9: document.getElementById("klasa9")?.checked,
    		samosaprometom: document.getElementById("samosaprometom")?.checked,
    		zabanku: document.getElementById("zabanku")?.checked
        }`
	hxValsStanja = `js:{
            bs_stanje_datum: document.getElementById("bs_stanje_datum")?.value
        }`
	hxValsUspeha = `js:{
            bu_od_datum: document.getElementById("bu_od_datum")?.value,
            bu_do_datum: document.getElementById("bu_do_datum")?.value
        }`
)

type BilansiHandler struct {
	tabData domain.TabData
	service *finservice.BilansiResource
	cfg     config.Config
}

func NewBilansiHandler(service *finservice.BilansiResource, cfg config.Config) *BilansiHandler {
	handler := &BilansiHandler{
		cfg: cfg,
	}
	handler.tabData = GetBilansiTabData()
	handler.service = service
	return handler
}

func (h *BilansiHandler) BilansiMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	h.tabData = setBilansiActiveTab(h.tabData, "zakljucnilist")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLZakljucni, "#bilansitable", "innerHTML", "GET", "", hxValsZakljucni, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput(i18n.GetInstance(), bilansiURLZakljucni, fmt.Sprintf("#%s", bilansiTableID), hxValsZakljucni)

	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetZakljucniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "ZAKLJUCNI LIST", "", false, false, false)
	tbl.HxVals = hxValsZakljucni
	err := tmpl_fin.BilansiMain(h.tabData, tbl, btnObrada, btnPrint, searchInput, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *BilansiHandler) ZakljucniList(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetZakljucniTableFields(), bilansiURLZakljucni, bilansiURLZakljucni, 0, 0, 0, 0, h.cfg)
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLZakljucni, "#bilansitable", "innerHTML", "GET", "", hxValsZakljucni, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput(translator, bilansiURLZakljucni, fmt.Sprintf("#%s", bilansiTableID), hxValsZakljucni)

		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		common.SetTableConfig(&tbl, "ZAKLJUCNI LIST", bilansiURLZakljucni, false, false, false)

		h.tabData = setBilansiActiveTab(h.tabData, "zakljucnilist")
		err := tmpl_fin.ZakljucniList(h.tabData, tbl, btnObrada, btnPrint, searchInput, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		//validacija input parametre:
		fieldParameters := []string{"odkonta", "dokonta", "odsifre", "dosifre", "oddatuma", "dodatuma", "tip_zakljucni", "analitickakonta", "klasa9", "samosaprometom", "zabanku"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsZakljucni
		err := h.service.GetZakljucniList(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetZakljucniList(c, &tbl, false, pageSize, page)
		tbl.HxVals = hxValsZakljucni
		tbl.Pagination.HxVals = hxValsZakljucni
		tbl.URLGetAll = bilansiURLZakljucni
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *BilansiHandler) BilansStanja(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansStanjaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "BILANS STANJA", bilansiURLStanja, true, true, false)
		searchInput := common.CreateSearchInput(i18n.GetInstance(), bilansiURLZakljucni, fmt.Sprintf("#%s", bilansiTableID), hxValsZakljucni)
		tbl.BtnAdd.IsVisible = true
		tbl.BtnPrint.IsVisible = true
		h.tabData = setBilansiActiveTab(h.tabData, "bilansstanja")
		err := tmpl_fin.BilansStanja(h.tabData, tbl, searchInput, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansStanjaTableFields(), "", bilansiURLStanja, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLStanja, true, true, false)
		err := h.service.GetBilansStanja(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetBilansStanja(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.HxVals = hxValsStanja
		tbl.Pagination.HxVals = hxValsStanja
		tbl.URLGetAll = bilansiURLStanja

		utils.RenderContent(c, tbl)
	}
}

func (h *BilansiHandler) StampanjeBilansStanja(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		// btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLAnKonta, "#bilansitable", "innerHTML", "GET", "", hxValsZakljucni, true, common.ClassSaveButton, "handleDialogResponse")
		// btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		// searchInput := common.CreateSearchInput(translator, prometURLAnKonta, fmt.Sprintf("#%s", prometTableID), hxValsZakljucni)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetZakljucniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, "", false, false, false)
		h.tabData = setActiveTab(h.tabData, "zakljucnilist")
		common.SetTableConfig(&tbl, "BILANS STANJA", bilansiURLStanja, false, false, false)

		h.tabData = setBilansiActiveTab(h.tabData, "stampanjebilansstanja")
		err := tmpl_fin.StampanjeBilansStanja(h.tabData, tbl, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		//validacija input parametre:
		fieldParameters := []string{"bs_stanje_datum"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansStanjaTableFields(), "", bilansiURLZakljucni, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLZakljucni, false, false, false)
		err := h.service.GetBilansStanja(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetBilansStanja(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *BilansiHandler) BilansUspeha(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansUspehaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "BILANS USPEHA", bilansiURLUspeha, true, true, false)

		h.tabData = setBilansiActiveTab(h.tabData, "bilansuspeha")
		err := tmpl_fin.BilansUspeha(h.tabData, tbl, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansUspehaTableFields(), "", bilansiURLUspeha, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLUspeha, true, true, false)
		err := h.service.GetBilansUspeha(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetBilansUspeha(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *BilansiHandler) StampanjeBilansUspeha(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansUspehaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "STAMPANJE BILANSA USPEHA", bilansiURLUspehaAlg, false, false, false)
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLUspehaAlg, "#bilansitable", "innerHTML", "GET", "", hxValsUspeha, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		btnExportXML := common.SetButton("exportxml-btn", "Export XML", "exportxml", "", "", "", "GET", "", hxValsUspeha, true, common.ClassButton, "handleExportXMLResponse")
		h.tabData = setBilansiActiveTab(h.tabData, "stampanjebilansuspeha")

		err := tmpl_fin.StampanjeBilansUspeha(h.tabData, tbl, btnObrada, btnPrint, btnExportXML, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		//validacija input parametre:
		fieldParameters := []string{"bu_od_datum", "bu_do_datum"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansUspehaTableFields(), "", bilansiURLUspehaAlg, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLUspehaAlg, false, false, false)
		err := h.service.GetBilansUspeha(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetBilansUspeha(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// RegisterRoutes registers the routes for the Bilansi handler
func (h *BilansiHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/bilansi", h.BilansiMain)
	r.GET("api/bilansi/zakljucni", h.ZakljucniList)
	r.GET("api/bilansi/stanja", h.BilansStanja)
	r.GET("api/bilansi/stanja/stampanje", h.StampanjeBilansStanja)
	r.GET("api/bilansi/uspeha", h.BilansUspeha)
	r.GET("api/bilansi/uspeha/stampanje", h.StampanjeBilansUspeha)
}

func GetBilansiTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "zakljucnilist", Label: "Zakljucni list", HXRequestUrl: bilansiURLZakljucni, IsActive: true, Name: "zakljucnilist"},
			{ID: "bilansstanja", Label: "Bilans stanja", HXRequestUrl: bilansiURLStanja, IsActive: false, Name: "bilansstanja"},
			{ID: "stampanjebilansstanja", Label: "Stampanje bilansa stanja", HXRequestUrl: bilansiURLStanjaAlg, IsActive: false, Name: "stampanjebilansstanja"},
			{ID: "bilansuspeha", Label: "Bilans uspeha", HXRequestUrl: bilansiURLUspeha, IsActive: false, Name: "bilansuspeha"},
			{ID: "stampanjebilansuspeha", Label: "Stampanje bilansa uspeha", HXRequestUrl: bilansiURLUspehaAlg, IsActive: false, Name: "stampanjebilansuspeha"},
		},
	}
}

func setBilansiActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
