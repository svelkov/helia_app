package finansijsko

import (
	"fmt"
	"helia/config"
	"helia/pkg/utils"
	"net/http"

	"helia/frontend/components"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	popdvContentTitle string = "PRIJAVA PDV OD NABAVKE DOBARA I USLUGA"
	popdvTableID      string = "popdv_table"
	popdvURLPrefix    string = "/api/popdv/"
	popdvURLPolja     string = "/api/popdv/polja"
	popdvURLPoljaUnos string = "/api/popdv/polja/unos"
	popdvURLPoljaSave string = "/api/popdv/polja/save"
	popdvURLPrijava   string = "/api/popdv/prijava"
	pppdvURLPrijava   string = "/api/popdv/pppdv/prijava"
	popdvURLStampa    string = "/api/popdv/stampa"
)

const (
	hxValsPopdvPrijava = `js:{
		oddatuma: document.getElementById("oddatuma")?.value,
		dodatuma: document.getElementById("dodatuma")?.value
	}`
)

type PopdvHandler struct {
	tabData domain.TabData
	service finservice.PopdvService
	cfg     config.Config
	lm      *middleware.LockMiddleware
}

func NewPopdvHandler(service finservice.PopdvService, cfg config.Config, lm *middleware.LockMiddleware) *PopdvHandler {
	handler := &PopdvHandler{
		cfg: cfg,
		lm:  lm,
	}
	handler.tabData = GetPopdvTabData()
	handler.service = service
	return handler
}

func (h *PopdvHandler) PopdvMain(c *gin.Context) {
	session := domain.GetSessionFromStdContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	h.tabData = setPopdvActiveTab(h.tabData, "polja-prijave")
	translator := i18n.GetInstance()
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), popdvURLPolja, fmt.Sprintf("#%s", popdvTableID), "")

	tbl := common.SetTableBasicData(popdvContentTitle, popdvTableID, h.service.GetPoljaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.URLGetAll = popdvURLPolja
	tbl.URLPrefix = popdvURLPolja
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxActionURL = popdvURLPoljaUnos
	tbl.BtnAdd.HxTarget = "#dialog-content"
	common.SetTableConfig(&tbl, popdvContentTitle, popdvURLPolja, true, true, false)
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search-input")
	ctx := c.Request.Context()

	err := h.service.GetPolja(ctx, &tbl, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetPolja(ctx, &tbl, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tmpl_fin.PopdvMain(h.tabData, tbl, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
}

func (h *PopdvHandler) PopdvPolja(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	h.tabData = setPopdvActiveTab(h.tabData, "polja-prijave")

	session := domain.GetSessionFromStdContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), popdvURLPolja, fmt.Sprintf("#%s", popdvTableID), "")

	tbl := common.SetTableBasicData(popdvContentTitle, popdvTableID, h.service.GetPoljaTableFields(), "", "", 0, 0, 0, 0, h.cfg)

	common.SetTableConfig(&tbl, "POLJA POPDV PRIJAVE", popdvURLPolja, true, true, false)

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search-input")
	ctx := c.Request.Context()

	err := h.service.GetPolja(ctx, &tbl, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetPolja(ctx, &tbl, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tbl.URLGetAll = popdvURLPolja
	tbl.URLPrefix = popdvURLPolja
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.IsVisible = true
	tbl.BtnAdd.HxActionURL = popdvURLPoljaUnos
	tbl.BtnAdd.HxTarget = "#dialog-content"
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		components.Table(tbl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	} else {
		tmpl_fin.PopdvPolja(h.tabData, tbl, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
	}
}

func (h *PopdvHandler) PopdvPoljaUnos(c *gin.Context) {
	translator := i18n.GetInstance()
	dialog := domain.Dialog{
		Id: "dialog-popdv-unos",
	}
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Sačuvaj",
		IsVisible:     true,
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   popdvURLPoljaSave,
		HxRequestType: "POST",
		HxSwap:        "none",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		LabelText: "Odustani",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassCloseButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	model := domain.Popdv{}
	csrfToken := common.GetCsrfToken(c)
	tmpl_fin.DialogPopdvPolja(model, dialog, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *PopdvHandler) PopdvPoljaSave(c *gin.Context) {
	// Implementation for saving POPDV fields
}

func (h *PopdvHandler) PopdvPrijava(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	h.tabData = setPopdvActiveTab(h.tabData, "popprijava")
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		btnObrada := common.SetButton("obrada-btn", translator.Button("Obrada"), "fin_obrada", popdvURLPrijava, "#"+popdvTableID, "innerHTML", "GET", "", hxValsPopdvPrijava, true, common.ClassSaveButton, "")
		btnDelete := common.SetButton("delete-btn", "Obriši", "fin_delete", "", "", "innerHTML", "GET", "", hxValsPopdvPrijava, true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), popdvURLPrijava, fmt.Sprintf("#%s", popdvTableID), hxValsPopdvPrijava)

		tbl := common.SetTableBasicData(popdvContentTitle, popdvTableID, h.service.GetPrijavaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsPopdvPrijava
		tbl.URLGetAll = popdvURLPrijava
		tbl.URLPrefix = popdvURLPrijava
		common.SetTableConfig(&tbl, "POPDV PRIJAVA", popdvURLPrijava, false, false, false)
		err := tmpl_fin.PopdvPrijava(h.tabData, tbl, btnObrada, btnDelete, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" || requestSource == "btn" {
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			utils.RenderDialogOK(c, "dialog-popdv-validation-error", common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		searchText := c.Query("query")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", popdvTableID, h.service.GetPrijavaTableFields(), "", popdvURLPrijava, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", popdvURLPrijava, false, false, false)
		tbl.Pagination.HxVals = hxValsPopdvPrijava
		err := h.service.GetPrijava(ctx, &tbl, true, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			utils.RenderDialogOK(c, "dialog-popdv-get-total-records", common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetPrijava(ctx, &tbl, false, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			utils.RenderDialogOK(c, "dialog-popdv-render-template", common.ErrMsgGetData)
			return
		}
		tbl.URLGetAll = popdvURLPrijava
		tbl.URLPrefix = popdvURLPrijava
		utils.RenderContent(c, tbl)
	}
}

func (h *PopdvHandler) PppdvPrijava(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	h.tabData = setPopdvActiveTab(h.tabData, "ppprijava")
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		btnObrada := common.SetButton("obrada-btn", translator.Button("Obrada"), "fin_obrada", popdvURLPrijava, "#"+popdvTableID, "innerHTML", "GET", "", hxValsPopdvPrijava, true, common.ClassSaveButton, "")
		btnDelete := common.SetButton("delete-btn", "Obriši", "fin_delete", "", "", "innerHTML", "GET", "", hxValsPopdvPrijava, true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), popdvURLPrijava, fmt.Sprintf("#%s", popdvTableID), hxValsPopdvPrijava)

		tbl := common.SetTableBasicData(popdvContentTitle, popdvTableID, h.service.GetPrijavaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsPopdvPrijava
		tbl.URLGetAll = popdvURLPrijava
		tbl.URLPrefix = popdvURLPrijava
		common.SetTableConfig(&tbl, "POPDV PRIJAVA", popdvURLPrijava, false, false, false)
		err := tmpl_fin.PopdvPrijava(h.tabData, tbl, btnObrada, btnDelete, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" || requestSource == "btn" {
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			utils.RenderDialogOK(c, "dialog-popdv-validation-error", common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		searchText := c.Query("query")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", popdvTableID, h.service.GetPrijavaTableFields(), "", popdvURLPrijava, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", popdvURLPrijava, false, false, false)
		tbl.Pagination.HxVals = hxValsPopdvPrijava
		err := h.service.GetPrijava(ctx, &tbl, true, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			utils.RenderDialogOK(c, "dialog-popdv-get-total-records", common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetPrijava(ctx, &tbl, false, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			utils.RenderDialogOK(c, "dialog-popdv-render-template", common.ErrMsgGetData)
			return
		}
		tbl.URLGetAll = popdvURLPrijava
		tbl.URLPrefix = popdvURLPrijava
		utils.RenderContent(c, tbl)
	}
}

func (h *PopdvHandler) PopdvStampa(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	h.tabData = setPopdvActiveTab(h.tabData, "stampa")
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnPrint := common.SetButton("print-btn", translator.Button("Štampa"), "fin_print", "", "", "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), popdvURLStampa, fmt.Sprintf("#%s", popdvTableID), "")

		tbl := common.SetTableBasicData(popdvContentTitle, popdvTableID, h.service.GetStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.URLGetAll = popdvURLStampa
		tbl.URLPrefix = popdvURLStampa
		common.SetTableConfig(&tbl, "STAMPA EVIDENCIJA", popdvURLStampa, false, false, false)

		h.tabData = setPopdvActiveTab(h.tabData, "stampa")
		err := tmpl_fin.PopdvStampa(h.tabData, tbl, btnPrint, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		searchText := c.Query("search-input")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", popdvTableID, h.service.GetStampaTableFields(), "", popdvURLStampa, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", popdvURLStampa, false, false, false)
		err := h.service.GetStampa(ctx, &tbl, true, pageSize, page, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetStampa(ctx, &tbl, false, pageSize, page, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}
		tbl.URLGetAll = popdvURLStampa
		tbl.URLPrefix = popdvURLStampa
		utils.RenderContent(c, tbl)
	}
}

// RegisterRoutes registers the routes for the EPP handler
func (h *PopdvHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/popdv", h.PopdvMain)
	r.GET("api/popdv/polja", h.PopdvPolja)
	r.POST("api/popdv/polja/save", h.PopdvPoljaSave)
	r.GET("api/popdv/polja/unos", h.PopdvPoljaUnos)
	r.GET("api/popdv/prijava", h.PopdvPrijava)
	r.GET("api/popdv/pppdv/prijava", h.PppdvPrijava)
	r.GET("api/popdv/stampa", h.PopdvStampa)
}

// Tab data helpers
func GetPopdvTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "polja-prijave", Label: "Polja POPDV prijave", HXRequestUrl: popdvURLPolja, IsActive: true, Icon: "home", Name: "polja-prijave"},
			{ID: "popprijava", Label: "POPDV prijava", HXRequestUrl: popdvURLPrijava, IsActive: false, Icon: "report", Name: "popprijava"},
			{ID: "ppprijava", Label: "PPPDV prijava", HXRequestUrl: pppdvURLPrijava, IsActive: false, Icon: "report", Name: "ppprijava"},
			{ID: "stampa", Label: "Stampa evidencija", HXRequestUrl: popdvURLStampa, IsActive: false, Icon: "print", Name: "stampa"},
		},
	}
}

func setPopdvActiveTab(tabData domain.TabData, activeTabId string) domain.TabData {
	for i := range tabData.Tabs {
		tabData.Tabs[i].IsActive = tabData.Tabs[i].ID == activeTabId
	}
	return tabData
}
