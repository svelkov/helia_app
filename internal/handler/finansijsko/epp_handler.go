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
	eppContentTitle  string = "EVIDENCIJA PRETHODNOG POREZA"
	eppTableID       string = "epp_table"
	eppURLPrefix     string = "/api/epp/"
	eppURLSekcije    string = "/api/epp/sekcije"
	eppURLEvidencija string = "/api/epp/evidencija"
	eppURLSefKpr     string = "/api/epp/sef-kpr"
)

const (
	hxValsEpp = `js:{
		oddatuma: document.getElementById("oddatuma")?.value,
		dodatuma: document.getElementById("dodatuma")?.value
	}`
)

type EppHandler struct {
	tabData domain.TabData
	service finservice.EppService
	cfg     config.Config
	lm *middleware.LockMiddleware
}

func NewEppHandler(service finservice.EppService, cfg config.Config, lm *middleware.LockMiddleware) *EppHandler {
	handler := &EppHandler{
		cfg: cfg,
		lm:  lm,
	}
	handler.tabData = GetEppTabData()
	handler.service = service
	return handler
}

func (h *EppHandler) EppMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}

	btnNew := common.SetButton("new-btn", "Novi", "fin_new", eppURLSekcije, "#"+eppTableID, "innerHTML", "GET", "", hxValsEpp, true, common.ClassSaveButton, "handleDialogResponse")
	btnDelete := common.SetButton("delete-btn", "Obriši", "fin_delete", "", "", "innerHTML", "GET", "", hxValsEpp, true, common.ClassButton, "")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), eppURLSekcije, fmt.Sprintf("#%s", eppTableID), hxValsEpp)

	tbl := common.SetTableBasicData(eppContentTitle, eppTableID, h.service.GetSekcijeIzvoriTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, eppContentTitle, "", false, false, false)
	tbl.Pagination.HxVals = hxValsEpp
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxActionURL = eppURLSekcije + "/unos"
	tbl.BtnAdd.HxTarget = "#dialog-content"
	common.SetTableConfig(&tbl, "EVIDENCIJA PRETHODNOG POREZA", eppURLSekcije, true, true, false)

	err := tmpl_fin.EppMain(h.tabData, tbl, btnNew, btnDelete, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *EppHandler) EppSekcijeIzvori(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnNew := common.SetButton("new-btn", "Novi", "fin_new", eppURLSekcije, "#"+eppTableID, "innerHTML", "GET", "", hxValsEpp, true, common.ClassSaveButton, "handleDialogResponse")
		btnDelete := common.SetButton("delete-btn", "Obriši", "fin_delete", "", "", "innerHTML", "GET", "", hxValsEpp, true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), eppURLSekcije, fmt.Sprintf("#%s", eppTableID), hxValsEpp)

		tbl := common.SetTableBasicData(eppContentTitle, eppTableID, h.service.GetSekcijeIzvoriTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsEpp
		tbl.BtnAdd.HxRequestType = "GET"
		tbl.BtnAdd.HxActionURL = eppURLSekcije + "/unos"
		tbl.BtnAdd.HxTarget = "#dialog-content"

		common.SetTableConfig(&tbl, "SEKCIJE I IZVORI", eppURLSekcije, true, true, false)

		h.tabData = setEppActiveTab(h.tabData, "sekcije")
		err := tmpl_fin.EppSekcijeIzvori(h.tabData, tbl, btnNew, btnDelete, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		searchText := c.Query("search-input")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", eppTableID, h.service.GetSekcijeIzvoriTableFields(), "", eppURLSekcije, 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsEpp
		common.SetTableConfig(&tbl, "", eppURLSekcije, true, true, false)
		err := h.service.GetSekcijeIzvori(ctx, &tbl, true, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSekcijeIzvori(ctx, &tbl, false, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *EppHandler) EppSekcijeUnos(c *gin.Context) {
	translator := i18n.GetInstance()
	btnSave := common.SetButton("save-btn", "Sačuvaj", "save", "", "#dialog-content", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "handleEppUnosResponse")
	btnCancel := common.SetButton("cancel-btn", "Odustani", "odustani", "", "#dialog-content", "innerHTML", "GET", "", "", true, common.ClassButton, "closeDialog()")

	err := tmpl_fin.DialogEppSekcije(btnSave, btnCancel, translator).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *EppHandler) EppEvidencija(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnNew := common.SetButton("new-btn", "Novi", "fin_new", eppURLEvidencija, "#"+eppTableID, "innerHTML", "GET", "", hxValsEpp, true, common.ClassSaveButton, "handleDialogResponse")
		btnDelete := common.SetButton("delete-btn", "Obriši", "fin_delete", "", "", "innerHTML", "GET", "", hxValsEpp, true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), eppURLEvidencija, fmt.Sprintf("#%s", eppTableID), hxValsEpp)

		tbl := common.SetTableBasicData(eppContentTitle, eppTableID, h.service.GetEvidencijaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsEpp
		common.SetTableConfig(&tbl, "EVIDENCIJA PP", eppURLEvidencija, true, true, false)

		h.tabData = setEppActiveTab(h.tabData, "evidencija")
		err := tmpl_fin.EppEvidencija(h.tabData, tbl, btnNew, btnDelete, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		searchText := c.Query("search-input")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", eppTableID, h.service.GetEvidencijaTableFields(), "", eppURLEvidencija, 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsEpp
		common.SetTableConfig(&tbl, "", eppURLEvidencija, true, true, false)
		err := h.service.GetEvidencija(ctx, &tbl, true, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetEvidencija(ctx, &tbl, false, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *EppHandler) EppSefKpr(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnNew := common.SetButton("new-btn", "Novi", "fin_new", eppURLSefKpr, "#"+eppTableID, "innerHTML", "GET", "", hxValsEpp, true, common.ClassSaveButton, "handleDialogResponse")
		btnDelete := common.SetButton("delete-btn", "Obriši", "fin_delete", "", "", "innerHTML", "GET", "", hxValsEpp, true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), eppURLSefKpr, fmt.Sprintf("#%s", eppTableID), hxValsEpp)

		tbl := common.SetTableBasicData(eppContentTitle, eppTableID, h.service.GetSefKprTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsEpp
		common.SetTableConfig(&tbl, "SEF-KPR", eppURLSefKpr, true, true, false)

		h.tabData = setEppActiveTab(h.tabData, "sef-kpr")
		err := tmpl_fin.EppSefKpr(h.tabData, tbl, btnNew, btnDelete, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		searchText := c.Query("search-input")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", eppTableID, h.service.GetSefKprTableFields(), "", eppURLSefKpr, 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsEpp
		common.SetTableConfig(&tbl, "", eppURLSefKpr, true, true, false)
		err := h.service.GetSefKpr(ctx, &tbl, true, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSefKpr(ctx, &tbl, false, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// RegisterRoutes registers the routes for the EPP handler
func (h *EppHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/epp", h.EppMain)
	r.GET("api/epp/sekcije", h.EppSekcijeIzvori)
	r.GET("api/epp/sekcije/unos", h.EppSekcijeUnos)
	r.GET("api/epp/evidencija", h.EppEvidencija)
	r.GET("api/epp/sefkpr", h.EppSefKpr)
}

func GetEppTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "sekcije-izvori", Label: "Sekcije i izvori", HXRequestUrl: eppURLSekcije, IsActive: true, Name: "sekcije-izvori"},
			{ID: "evidencija", Label: "Evidencija PP", HXRequestUrl: eppURLEvidencija, IsActive: false, Name: "evidencija"},
			{ID: "sef-kpr", Label: "SEF-KPR", HXRequestUrl: eppURLSefKpr, IsActive: false, Name: "sef-kpr"},
		},
	}
}

func setEppActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
