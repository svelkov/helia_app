package finansijsko

import (
	"fmt"
	"helia/config"
	"net/http"
	"strings"

	tmpl "helia/frontend/templates"
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
	izvodiContentTitle      string = "IZVODI"
	izvodiTableMasterID     string = "izvodi-master-table"
	izvodiTableDetailID     string = "izvodi-detail-table"
	izvodiURLPrefix         string = "/api/izvodi/"
	izvodiURLGetAll         string = "/api/izvodi/all"
	izvodiURLUcitavanje     string = "/api/izvodi/ucitavanje"
	izvodiURLKnjizenje      string = "/api/izvodi/knjizenje"
	izvodiURLPregled        string = "/api/izvodi/pregled"
	izvodiURLImport         string = "/api/izvodi/import"
	izvodiURLBrisi          string = "/api/izvodi/brisi"
	izvodiURLAzurirajKonta  string = "/api/izvodi/azurirajkonta"
	izvodiURLRaznoteza      string = "/api/izvodi/raznoteza"
	izvodiURLPregledDetalji string = "/api/izvodi/pregled/detalji"

	hxValsPregledIzvoda = ` js:{
		idbanke: document.getElementById("idbanke").value,
		odbrojaizvoda: document.getElementById("odbrojaizvoda").value,
		dobrojaizvoda: document.getElementById("dobrojaizvoda").value,
		oddatuma: document.getElementById("oddatuma").value,
		dodatuma: document.getElementById("dodatuma").value,
		status: document.querySelector('input[name="status"]:checked') ? document.querySelector('input[name="status"]:checked').value : "",
	}`
	hxValsKnjizenjeIzvoda = ` js:{
		idbanke: document.getElementById("idbanke").value,
		odbrojaizvoda: document.getElementById("odbrojaizvoda").value,
		dobrojaizvoda: document.getElementById("dobrojaizvoda").value,
		oddatuma: document.getElementById("oddatuma").value,
		dodatuma: document.getElementById("dodatuma").value,
		}`
)

type IzvodiHandler struct {
	tabData domain.TabData
	service finservice.IzvodiService
	cfg     config.Config
	lm      *middleware.LockMiddleware
}

func NewIzvodiHandler(service finservice.IzvodiService, cfg config.Config, lm *middleware.LockMiddleware) *IzvodiHandler {
	handler := &IzvodiHandler{
		cfg: cfg,
		lm:  lm,
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
	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)
	// Create configuration
	btnImport := common.SetButton("import-btn", translator.Button("Import"), "fin_import", izvodiURLUcitavanje, "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
	btnAzurKonta := common.SetButton("azur-konta-btn", translator.Button("Ažuriraj konta"), "fin_azurirajkonta", izvodiURLAzurirajKonta, "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
	btnBrisi := common.SetButton("brisi-btn", translator.Button("Briši"), "delete", izvodiURLBrisi, "#izvodi-master-table", "innerHTML", "DELETE", "", "", true, common.ClassDeleteButton, "")

	tblMaster := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblMaster, izvodiContentTitle, "", false, false, false)

	tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetail, "", "", false, false, false)
	h.tabData = setIzvodiActiveTab(h.tabData, "ucitavanje")
	bankValues, err := h.service.GetBanke(c.Request.Context()) // Get bank values for dropdowns
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	err = tmpl_fin.IzvodiMain(h.tabData, tblMaster, tblDetail, btnImport, btnAzurKonta, btnBrisi, bankValues, gnGod, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *IzvodiHandler) UcitavanjeIzvoda(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnImport := common.SetButton("import-btn", translator.Button("Import"), "fin_import", izvodiURLUcitavanje, "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
		btnAzurKonta := common.SetButton("azur-konta-btn", translator.Button("Ažuriraj konta"), "fin_azurirajkonta", izvodiURLAzurirajKonta, "#izvodi-master-table", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "")
		btnBrisi := common.SetButton("brisi-btn", translator.Button("Briši"), "delete", izvodiURLBrisi, "#izvodi-master-table", "innerHTML", "DELETE", "", "", true, common.ClassDeleteButton, "")

		tblMaster := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblMaster, izvodiContentTitle, "", false, false, false)

		tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblDetail, "", "", false, false, false)
		bankeValues, err := h.service.GetBanke(c.Request.Context()) // Get bank values for dropdowns
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}

		h.tabData = setIzvodiActiveTab(h.tabData, "ucitavanje")
		tmpl_fin.UcitavanjeIzvoda(h.tabData, tblMaster, tblDetail, btnImport, btnAzurKonta, btnBrisi, bankeValues, gnGod, translator, csrfToken).Render(c.Request.Context(), c.Writer)
	}

	if requestSource == "btn" {
		filePath := c.PostForm("filepath")

		softwareName := c.PostForm("software")
		if filePath == "" {
			utils.RenderDialogOK(c, "dialog-ucitavanjeizvoda", common.ErrMsgFileSelection)
			return
		}
		if softwareName == "" {
			utils.RenderDialogOK(c, "dialog-ucitavanjeizvoda", common.ErrMsgSoftwareSelection)
			return
		}
		fileData := c.PostForm("file")
		fileLines := strings.Split(fileData, "\n")
		err := h.service.ImportIzvod(c.Request.Context(), fileLines, softwareName)
		if err != nil {
			utils.RenderDialogOK(c, "dialog-ucitavanjeizvoda", fmt.Sprintf("Greška prilikom uvoza izvoda: %v", err))
			return
		}
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
		csrfToken := common.GetCsrfToken(c)
		btnObrada := common.SetButton("obrada-btn", translator.Button("Obrada"), "fin_obrada", izvodiURLKnjizenje, "#izvodi-master-table", "innerHTML", "GET", "", hxValsKnjizenjeIzvoda, true, common.ClassSaveButton, "")
		btnRavnoteza := common.SetButton("ravnoteza-btn", translator.Button("Proveri ravnotežu"), "fin_ravnoteza", izvodiURLRaznoteza, "#izvodi-detail-table", "innerHTML", "POST", "", hxValsKnjizenjeIzvoda, true, common.ClassSaveButton, "")
		btnKnjizenje := common.SetButton("knjizenje-btn", translator.Button("Knjiženje"), "fin_knjizenje", izvodiURLKnjizenje, "#izvodi-master-table", "innerHTML", "POST", "", hxValsKnjizenjeIzvoda, true, common.ClassSaveButton, "")

		tblHeader := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblHeader, izvodiContentTitle, "", false, false, false)
		tblHeader.URLGetAll = izvodiURLKnjizenje
		tblHeader.URLPrefix = izvodiURLKnjizenje
		tblHeader.Pagination.HxVals = hxValsKnjizenjeIzvoda
		tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblDetail, "", "", false, false, false)
		bankeValues, err := h.service.GetBanke(c.Request.Context()) // Get bank values for dropdowns
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}
		tipdokValues, err := h.service.GetTipdokOptions(c.Request.Context()) // Get tip dok values for dropdowns
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}
		nextNalog := int64(0)
		if tipdokValues != nil && len(tipdokValues) > 0 {
			nextNalog, err = h.service.GetNextNalog(c.Request.Context(), tipdokValues[0].Value) // Get next nalog for the first tipdok value
			if err != nil {
				common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
				return
			}
		}
		h.tabData = setIzvodiActiveTab(h.tabData, "knjizenje")
		tmpl_fin.KnjizenjeIzvoda(h.tabData, tblHeader, tblDetail, btnObrada, btnRavnoteza, btnKnjizenje, bankeValues, tipdokValues, gnGod, fmt.Sprintf("%d", nextNalog), translator, csrfToken).Render(c.Request.Context(), c.Writer)
		return
	}

	if requestSource == "btnprihvati" || requestSource == "btnpage" || requestSource == "searchInput" || requestSource == "btnobrada" || requestSource == "btn" {
		ctx := c.Request.Context()
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tblHeader := common.SetTableBasicData("", izvodiTableMasterID, h.service.GetMasterTableFields(), "", izvodiURLKnjizenje, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblHeader, izvodiContentTitle, izvodiURLKnjizenje, false, false, false)
		tblHeader.DetailTarget = fmt.Sprintf("#%s", izvodiTableDetailID)
		tblHeader.DetailURL = izvodiURLPregledDetalji
		tblHeader.DetailHxRequestType = "GET"
		tblHeader.DetailHxSwap = "innerHTML"
		tblHeader.DetailHxTrigger = "click, change delay:500ms"
		params := domain.IzvodiParams{
			IDbanke:        c.Query("idbanke"),
			OdBrojaIzvoda:  c.Query("odbrojaizvoda"),
			DoBrojaIzvoda:  c.Query("dobrojaizvoda"),
			OdDatumaIzvoda: c.Query("oddatuma"),
			DoDatumaIzvoda: c.Query("dodatuma"),
			StatusIzvoda:   "1", // samo neproknjizene izvode
			SearchText:     c.Query("search"),
		}

		err := h.service.GetIzvodiHeader(ctx, &tblHeader, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = h.service.GetIzvodiHeader(ctx, &tblHeader, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		tmpl.Table(tblHeader, translator).Render(c.Request.Context(), c.Writer)
		return
	}
}

func (h *IzvodiHandler) PregledIzvoda(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	tblHeader := common.SetTableBasicData(izvodiContentTitle, izvodiTableMasterID, h.service.GetMasterTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblHeader, izvodiContentTitle, "", false, false, false)
	tblHeader.URLGetAll = izvodiURLPregled
	tblHeader.URLPrefix = izvodiURLPregled
	tblHeader.Pagination.HxVals = hxValsPregledIzvoda
	tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "obrada", izvodiURLPregled, "#izvodi-master-table", "innerHTML", "GET", "", hxValsPregledIzvoda, true, common.ClassSaveButton, "")

		common.SetTableConfig(&tblDetail, "", "", false, false, false)
		bankeValues, err := h.service.GetBanke(c.Request.Context()) // Get bank values for dropdowns
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}

		h.tabData = setIzvodiActiveTab(h.tabData, "pregled")
		err = tmpl_fin.PregledIzvoda(h.tabData, tblHeader, tblDetail, btnObrada, bankeValues, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchInput" {
		ctx := c.Request.Context()
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)

		tblHeader.DetailTarget = fmt.Sprintf("#%s", izvodiTableDetailID)
		tblHeader.DetailURL = izvodiURLPregledDetalji
		tblHeader.DetailHxRequestType = "GET"
		tblHeader.DetailHxSwap = "innerHTML"
		tblHeader.DetailHxTrigger = "click, change delay:500ms"
		params := domain.IzvodiParams{
			OdBrojaIzvoda:  c.Query("odbrojaizvoda"),
			DoBrojaIzvoda:  c.Query("dobrojaizvoda"),
			OdDatumaIzvoda: c.Query("oddatuma"),
			DoDatumaIzvoda: c.Query("dodatuma"),
			StatusIzvoda:   c.Query("status"),
			IDbanke:        c.Query("idbanke"),
		}

		err := h.service.GetIzvodiHeader(ctx, &tblHeader, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetIzvodiHeader(ctx, &tblHeader, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		utils.RenderContent(c, tblHeader)
	}
}
func (h *IzvodiHandler) GetIzvodiDetalji(c *gin.Context) {
	idFizvzag, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, "ID izvoda je obavezan")
		return
	}
	ctx := c.Request.Context()
	tblDetail := common.SetTableBasicData("", izvodiTableDetailID, h.service.GetDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetail, "", "", false, false, false)
	tblDetail.URLPrefix = fmt.Sprintf("%s/%d", izvodiURLPregledDetalji, idFizvzag)
	tblDetail.URLGetAll = fmt.Sprintf("%s/%d", izvodiURLPregledDetalji, idFizvzag)

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	params := domain.IzvodiParams{
		IDfizvzag: fmt.Sprintf("%d", idFizvzag),
	}
	err = h.service.GetIzvodiDetail(ctx, &tblDetail, true, pageSize, page, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	err = h.service.GetIzvodiDetail(ctx, &tblDetail, false, pageSize, page, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}
	tmpl.Table(tblDetail, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *IzvodiHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("/api/izvodi", h.IzvodiMain)
	r.POST("/api/izvodi/ucitavanje", h.UcitavanjeIzvoda)
	r.GET("/api/izvodi/knjizenje", h.KnjizenjeIzvoda)
	r.GET("/api/izvodi/pregled", h.PregledIzvoda)
	r.GET("/api/izvodi/pregled/detalji/:id", h.GetIzvodiDetalji)
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
