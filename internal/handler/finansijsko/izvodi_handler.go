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
	izvodiContentTitle  string = "IZVODI"
	izvodiTableMasterID string = "izvodi-master-table"
	izvodiTableDetailID string = "izvodi-detail-table"
	izvodiURLPrefix     string = "/api/izvodi/"
	izvodiURLGetAll     string = "/api/izvodi/all"
	izvodiURLUcitavanje string = "/api/izvodi/ucitavanje"
	izvodiURLKnjizenje  string = "/api/izvodi/knjizenje"
	izvodiURLPregled    string = "/api/izvodi/pregled"
)

type IzvodiHandler struct {
	tabData domain.TabData
	service finservice.IzvodiService
	cfg     config.Config
}

func NewIzvodiHandler(service finservice.IzvodiService, cfg config.Config) *IzvodiHandler {
	handler := &IzvodiHandler{
		cfg: cfg,
	}
	handler.tabData = GetIzvodiTabData()
	handler.service = service
	return handler
}

func (h *IzvodiHandler) IzvodiMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}

	// Create configuration
	btnImport := common.SetButton("import-btn", "Import", "import", "/api/izvodi/import", "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
	btnAzurKonta := common.SetButton("azur-konta-btn", "Ažur. konta", "azur_konta", "/api/izvodi/azurirajkonta", "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
	btnBrisi := common.SetButton("brisi-btn", "Briši", "brisi", "/api/izvodi/brisi", "#izvodi-master-table", "innerHTML", "DELETE", "", "", true, common.ClassDeleteButton, "")

	tblMaster := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblMaster, izvodiContentTitle, "", false, false, false)

	tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetail, "", "", false, false, false)

	err := tmpl_fin.IzvodiMain(h.tabData, tblMaster, tblDetail, btnImport, btnAzurKonta, btnBrisi, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *IzvodiHandler) UcitavanjeIzvoda(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnImport := common.SetButton("import-btn", "Import", "import", izvodiURLUcitavanje, "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
		btnAzurKonta := common.SetButton("azur-konta-btn", "Ažur. konta", "azur_konta", "/api/izvodi/azurirajkonta", "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
		btnBrisi := common.SetButton("brisi-btn", "Briši", "brisi", "/api/izvodi/brisi", "#izvodi-master-table", "innerHTML", "DELETE", "", "", true, common.ClassDeleteButton, "")

		tblMaster := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblMaster, izvodiContentTitle, "", false, false, false)

		tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblDetail, "", "", false, false, false)

		h.tabData = setIzvodiActiveTab(h.tabData, "ucitavanje")
		err := tmpl_fin.UcitavanjeIzvoda(h.tabData, tblMaster, tblDetail, btnImport, btnAzurKonta, btnBrisi, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}

	if requestSource == "btnimport" {
		// TODO: Implement import logic
		common.WriteJSONResponse(c, http.StatusOK, true, nil, "Import completed successfully")
	}
}

func (h *IzvodiHandler) KnjizenjeIzvoda(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnPrihvati := common.SetButton("obrada-btn", "Obrada", "obrada", izvodiURLKnjizenje, "#izvodi-master-table", "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
		btnRaznoteza := common.SetButton("raznoteza-btn", "Proveri raznotežu", "raznoteza", "/api/izvodi/raznoteza", "#izvodi-detail-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
		btnKnjizenje := common.SetButton("knjizenje-btn", "Knjiženje", "knjizenje", "/api/izvodi/knjizenje", "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")

		tblMaster := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblMaster, izvodiContentTitle, "", false, false, false)

		tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblDetail, "", "", false, false, false)

		h.tabData = setIzvodiActiveTab(h.tabData, "knjizenje")
		err := tmpl_fin.KnjizenjeIzvoda(h.tabData, tblMaster, tblDetail, btnPrihvati, btnRaznoteza, btnKnjizenje, []domain.ComboItem{}, []domain.ComboItem{}, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}

	if requestSource == "btnprihvati" || requestSource == "btnpage" {
		// TODO: Implement data retrieval logic
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tblMaster := common.SetTableBasicData("", izvodiTableMasterID, h.service.GetMasterTableFields(), "", izvodiURLKnjizenje, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblMaster, "", izvodiURLKnjizenje, false, false, false)

		err := h.service.GetIzvodiHeader(c, &tblMaster, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetIzvodiHeader(c, &tblMaster, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tblMaster)
	}
}

func (h *IzvodiHandler) PregledIzvoda(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", izvodiURLPregled, "#izvodi-master-table", "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")

		tblMaster := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblMaster, izvodiContentTitle, "", false, false, false)

		tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblDetail, "", "", false, false, false)

		h.tabData = setIzvodiActiveTab(h.tabData, "pregled")
		err := tmpl_fin.PregledIzvoda(h.tabData, tblMaster, tblDetail, btnObrada, []domain.ComboItem{}, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchInput" {
		// TODO: Implement data retrieval logic
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tblMaster := common.SetTableBasicData("", izvodiTableMasterID, h.service.GetMasterTableFields(), "", izvodiURLPregled, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblMaster, "", izvodiURLPregled, false, false, false)

		err := h.service.GetIzvodiHeader(c, &tblMaster, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetIzvodiHeader(c, &tblMaster, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tblMaster)
	}
}

func (h *IzvodiHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("/api/izvodi", h.IzvodiMain)
	r.GET("/api/izvodi/ucitavanje", h.UcitavanjeIzvoda)
	r.GET("/api/izvodi/knjizenje", h.KnjizenjeIzvoda)
	r.GET("/api/izvodi/pregled", h.PregledIzvoda)
	r.POST("/api/izvodi/import", h.UcitavanjeIzvoda)
	r.POST("/api/izvodi/azurirajkonta", h.UcitavanjeIzvoda)
	r.DELETE("/api/izvodi/brisi", h.UcitavanjeIzvoda)
	r.POST("/api/izvodi/raznoteza", h.KnjizenjeIzvoda)
	r.POST("/api/izvodi/knjizenje", h.KnjizenjeIzvoda)
}

func GetIzvodiTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "ucitavanje", Label: "Učitavanje izvoda", HXRequestUrl: fmt.Sprintf("%sucitavanje", izvodiURLPrefix), IsActive: true, Name: "ucitavanje"},
			{ID: "knjizenje", Label: "Knjiženje izvoda", HXRequestUrl: fmt.Sprintf("%sknjizenje", izvodiURLPrefix), IsActive: false, Name: "knjizenje"},
			{ID: "pregled", Label: "Pregled izvoda", HXRequestUrl: fmt.Sprintf("%spregled", izvodiURLPrefix), IsActive: false, Name: "pregled"},
		},
	}
}

func setIzvodiActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
