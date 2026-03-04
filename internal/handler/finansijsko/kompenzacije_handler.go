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
	kompenzacijeContentTitle               string = "KOMPENZACIJE"
	kompenzacijePregledTableID             string = "kompenzacije-pregled-table"
	kompenzacijeFormiranjeTableID          string = "kompenzacije-formiranje-table"
	kompenzacijePregledKompenzacijaTableID string = "kompenzacije-pregled-kompenzacija-table"
	kompenzacijeKnjizenjeTableID           string = "kompenzacije-knjizenje-table"
	kompenzacijeURLPrefix                  string = "/api/kompenzacije/"
	kompenzacijeURLPregledPartnera         string = "/api/kompenzacije/pregledpartnera"
	kompenzacijeURLFormiranje              string = "/api/kompenzacije/formiranje"
	kompenzacijeURLPregled                 string = "/api/kompenzacije/pregled"
	kompenzacijeURLKnjizenje               string = "/api/kompenzacije/knjizenje"
)

type KompenzacijeHandler struct {
	tabData domain.TabData
	cfg     config.Config
	service *finservice.KompenzacijeResource
}

const (
	hxValsKompenzacijePregledPartnera = `js:{
            "query": document.getElementById("search-input")?.value
        }`
	hxValsKompenzacijeFormiranje = `js:{
            "konto_duznika": document.getElementById("konto-duznika")?.value,
			"sifra_duznika": document.getElementById("sifra-duznika")?.value,
			"konto_poverioca": document.getElementById("konto-poverioca")?.value,
			"sifra_poverioca": document.getElementById("sifra-poverioca")?.value
        }`
	hxValsKompenzacijePregled = `js:{
            "query": document.getElementById("search-input")?.value,
			"status_filter": document.querySelector('input[name="status-filter"]:checked')?.value
        }`
	hxValsKompenzacijeKnjizenje = `js:{
            "od_datuma_kompenz": document.getElementById("od-datuma-kompenz")?.value,
			"do_datuma_kompenz": document.getElementById("do-datuma-kompenz")?.value,
			"od_broja_komp": document.getElementById("od-broja-komp")?.value,
			"do_broja_komp": document.getElementById("do-broja-komp")?.value,
			"tipdok_knjizenje": document.getElementById("tipdok-knjizenje")?.value,
			"opis_knjizenja": document.getElementById("opis-knjizenja")?.value
        }`
)

func NewKompenzacijeHandler(service *finservice.KompenzacijeResource, cfg config.Config) *KompenzacijeHandler {
	handler := &KompenzacijeHandler{
		cfg:     cfg,
		service: service,
	}
	handler.tabData = GetKompenzacijeTabData()
	return handler
}

// Main kompenzacije handler - displays the initial page with first tab
func (h *KompenzacijeHandler) KompenzacijeMain(c *gin.Context) {
	translator := i18n.GetInstance()
	csrfToken, _ := c.Cookie("csrf_token")
	searchInput := common.CreateSearchInput("search-input", translator, kompenzacijeURLPregledPartnera, fmt.Sprintf("#%s", kompenzacijePregledTableID), hxValsKompenzacijePregledPartnera)

	// Create configuration
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/kompenzacije/pregledpartnera", fmt.Sprintf("#%s", kompenzacijePregledTableID), "innerHTML", "GET", "", hxValsKompenzacijePregledPartnera, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(kompenzacijeContentTitle, kompenzacijePregledTableID, h.service.GetPregledPartneraTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "PREGLED PARTNERA ZA FORMIRANJE KOMPENZACIJE", kompenzacijeURLPregledPartnera, false, false, false)

	setActiveKompenzacijeTab(&h.tabData, "pregledpartnera")
	err := tmpl_fin.KompenzacijePregledPartnera(h.tabData, tbl, searchInput, btnObrada, btnPrint, translator, csrfToken).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

// Tab 1: Pregled partnera za kompenzacije
func (h *KompenzacijeHandler) KompenzacijePregledPartnera(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	csrfToken, _ := c.Cookie("csrf_token")
	translator := i18n.GetInstance()

	searchInput := common.CreateSearchInput("search-input", translator, kompenzacijeURLPregledPartnera, fmt.Sprintf("#%s", kompenzacijePregledTableID), hxValsKompenzacijePregledPartnera)

	tbl := common.SetTableBasicData(kompenzacijeContentTitle, kompenzacijePregledTableID, h.service.GetPregledPartneraTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "PREGLED PARTNERA ZA FORMIRANJE KOMPENZACIJE", kompenzacijeURLPregledPartnera, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", kompenzacijeURLPregledPartnera, fmt.Sprintf("#%s", kompenzacijePregledTableID), "innerHTML", "GET", "", hxValsKompenzacijePregledPartnera, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

		setActiveKompenzacijeTab(&h.tabData, "pregledpartnera")
		err := tmpl_fin.KompenzacijePregledPartnera(h.tabData, tbl, searchInput, btnObrada, btnPrint, translator, csrfToken).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

		err := h.service.ObradaPredlogKompenzacije(c, &tbl, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.Pagination.HxVals = hxValsKompenzacijePregledPartnera
		err = h.service.ObradaPredlogKompenzacije(c, &tbl, false, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// Tab 2: Formiranje kompenzacije
func (h *KompenzacijeHandler) KompenzacijeFormiranje(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	csrfToken, _ := c.Cookie("csrf_token")

	duznikData := common.SetTableBasicData("", "duznik-table", h.service.GetFormiranjeTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&duznikData, "", "", false, false, false)
	duznikData.ShowPagination = false

	poverilacData := common.SetTableBasicData("", "poverilac-table", h.service.GetFormiranjeTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&poverilacData, "", "", false, false, false)
	poverilacData.ShowPagination = false

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", kompenzacijeURLFormiranje, "#tab2", "innerHTML", "GET", "", hxValsKompenzacijeFormiranje, true, common.ClassSaveButton, "handleDialogResponse")
		btnFormKomp := common.SetButton("form-komp-btn", "Formiraj kompenzaciju", "fin_save", kompenzacijeURLFormiranje+"/formiraj", "#tab2", "innerHTML", "POST", "", hxValsKompenzacijeFormiranje, true, common.ClassAddButton, "")

		setActiveKompenzacijeTab(&h.tabData, "formiranje")
		err := tmpl_fin.KompenzacijeFormiranje(h.tabData, duznikData, poverilacData, btnObrada, btnFormKomp, translator, csrfToken).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		// TODO: Implement data fetching when service is ready
		err := h.service.FormiranjeKompenzacije(c, &duznikData, &poverilacData, 1)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = tmpl_fin.KompenzacijeFormiranje(h.tabData, duznikData, poverilacData, domain.Button{}, domain.Button{}, translator, csrfToken).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		}
	}
}

// Tab 3: Pregled kompenzacija
func (h *KompenzacijeHandler) KompenzacijePregled(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	searchInput := common.CreateSearchInput("search-input", translator, kompenzacijeURLPregled, fmt.Sprintf("#%s", kompenzacijePregledKompenzacijaTableID), hxValsKompenzacijePregled)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", kompenzacijeURLPregled, fmt.Sprintf("#%s", kompenzacijePregledKompenzacijaTableID), "innerHTML", "GET", "", hxValsKompenzacijePregled, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa-btn", "Štampa", "fin_print", kompenzacijeURLPregled+"/print", "", "innerHTML", "GET", "", hxValsKompenzacijePregled, true, common.ClassPrintButton, "")

	tblHdr := common.SetTableBasicData("", kompenzacijePregledKompenzacijaTableID, h.service.GetKompenzacijeTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblHdr, "KOMPENZACIJE", kompenzacijeURLPregled, false, false, false)

	tblDet := common.SetTableBasicData("", "kompenzacije-det-table", h.service.GetDokumentaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDet, "", "", false, false, false)
	tblDet.ShowPagination = false

	if requestSource == "menu" || requestSource == "tab" {

		setActiveKompenzacijeTab(&h.tabData, "pregled")
		err := tmpl_fin.KompenzacijePregled(h.tabData, tblHdr, tblDet, searchInput, btnObrada, btnPrint, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		// TODO: Implement data fetching when service is ready
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

		err := h.service.PregledKompenzacije(c, &tblHdr, &tblDet, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.PregledKompenzacije(c, &tblHdr, &tblDet, false, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tblHdr)
	}
}

// Tab 4: Knjiženje kompenzacija
func (h *KompenzacijeHandler) KompenzacijeKnjizenje(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	csrfToken, _ := c.Cookie("csrf_token")

	// TODO: Replace with actual combo values from service
	tipdokValues := []domain.ComboItem{
		{Key: "1", Value: "Finansijski nalog"},
		{Key: "2", Value: "Kompenzacija"},
		{Key: "3", Value: "Drugi tip"},
	}

	tbl := common.SetTableBasicData("", kompenzacijeKnjizenjeTableID, h.service.GetKompenzacijeTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "", "", false, false, true)

	dokumentaData := common.SetTableBasicData("", "kompenzacije-knjizenje-dokumenta-table", h.service.GetDokumentaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&dokumentaData, "", "", false, false, false)
	dokumentaData.ShowPagination = false

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", kompenzacijeURLKnjizenje, fmt.Sprintf("#%s", kompenzacijeKnjizenjeTableID), "innerHTML", "GET", "", hxValsKompenzacijeKnjizenje, true, common.ClassSaveButton, "handleDialogResponse")
		btnRavnoteza := common.SetButton("ravnoteza-btn", "Pr. Ravnotezu", "fin_ravnoteza", kompenzacijeURLKnjizenje+"/ravnoteza", "", "innerHTML", "POST", "", hxValsKompenzacijeKnjizenje, true, common.ClassAddButton, "")
		btnKnjizi := common.SetButton("knjizi-btn", "Knjiži", "fin_save", kompenzacijeURLKnjizenje+"/knjizi", "", "innerHTML", "POST", "", hxValsKompenzacijeKnjizenje, true, common.ClassSaveButton, "")

		setActiveKompenzacijeTab(&h.tabData, "knjizenje")
		err := tmpl_fin.KompenzacijeKnjizenje(h.tabData, tbl, dokumentaData, tipdokValues, btnObrada, btnRavnoteza, btnKnjizi, translator, csrfToken).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" {
		// TODO: Implement data fetching when service is ready
		// err := h.service.GetKnjizenjeData(c, &tbl, &dokumentaData)
		// if err != nil {
		// 	common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		// 	return
		// }

		utils.RenderContent(c, tbl)
	}
}

// AddRoutes registers all kompenzacije routes
func (h *KompenzacijeHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for kompenzacije
	r.GET("/api/kompenzacije", h.KompenzacijeMain)
	r.GET("/api/kompenzacije/pregledpartnera", h.KompenzacijePregledPartnera)
	r.GET("/api/kompenzacije/formiranje", h.KompenzacijeFormiranje)
	r.POST("/api/kompenzacije/formiranje/formiraj", h.KompenzacijeFormiranje) // TODO: Create separate handler for POST
	r.GET("/api/kompenzacije/pregled", h.KompenzacijePregled)
	r.GET("/api/kompenzacije/pregled/print", h.KompenzacijePregled) // TODO: Create print handler
	r.GET("/api/kompenzacije/knjizenje", h.KompenzacijeKnjizenje)
	r.POST("/api/kompenzacije/knjizenje/ravnoteza", h.KompenzacijeKnjizenje) // TODO: Create separate handler
	r.POST("/api/kompenzacije/knjizenje/knjizi", h.KompenzacijeKnjizenje)    // TODO: Create separate handler
}

// GetKompenzacijeTabData returns the tab configuration for kompenzacije module
func GetKompenzacijeTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "pregledpartnera", Label: "Pregled partnera", HXRequestUrl: fmt.Sprintf("%spregledpartnera", kompenzacijeURLPrefix), IsActive: true, Name: "pregledpartnera"},
			{ID: "formiranje", Label: "Formiranje", HXRequestUrl: fmt.Sprintf("%sformiranje", kompenzacijeURLPrefix), IsActive: false, Name: "formiranje"},
			{ID: "pregled", Label: "Pregled", HXRequestUrl: fmt.Sprintf("%spregled", kompenzacijeURLPrefix), IsActive: false, Name: "pregled"},
			{ID: "knjizenje", Label: "Knjiženje", HXRequestUrl: fmt.Sprintf("%sknjizenje", kompenzacijeURLPrefix), IsActive: false, Name: "knjizenje"},
		},
	}
}

// setActiveKompenzacijeTab sets the active tab in the tab data
func setActiveKompenzacijeTab(tabs *domain.TabData, tabName string) {
	for i := range tabs.Tabs {
		tabs.Tabs[i].IsActive = tabs.Tabs[i].Name == tabName
	}
}
