package finansijsko

import (
	"fmt"
	"helia/config"
	"helia/pkg/utils"
	"net/http"

	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	fseppContentTitle    string = "EVIDENCIJA PRETHODNOG POREZA"
	fseppTableID         string = "fsepp_table"
	fseppURLPrefix       string = "/api/fsepp/"
	fseppURLSekcije      string = "/api/fsepp/sekcije"
	fseppURLSekcijeUnos  string = "/api/fsepp/sekcije/unos"
	fseppURLSekcijeSave  string = "/api/fsepp/sekcije/save"
	fseppURLEvidencija   string = "/api/fsepp/evidencija"
	fseppURLSefKpr       string = "/api/fsepp/sefkpr"
	fseppURLSefKprImport string = "/api/fsepp/sefkpr/import"
)

const (
	hxValsFseppEvidencija = `js:{
		oddatuma: document.getElementById("oddatuma")?.value,
		dodatuma: document.getElementById("dodatuma")?.value
	}`
	hxValsFseppKpr = `js:{
		oddatuma: document.getElementById("oddatuma")?.value,
		dodatuma: document.getElementById("dodatuma")?.value
	}`
	hxValsFseppImport = `js:{
		oddatuma: document.getElementById("oddatuma")?.value,
		dodatuma: document.getElementById("dodatuma")?.value,
		filter_type: document.querySelector('input[name="filter_type"]:checked')?.value,
		file: document.getElementById("file-input-data")?.value,
		file_type: document.getElementById("file-type-data")?.value
	}`
)

type FseppHandler struct {
	tabData domain.TabData
	service finservice.FseppService
	cfg     config.Config
	lm      *middleware.LockMiddleware
}

func NewFseppHandler(service finservice.FseppService, cfg config.Config, lm *middleware.LockMiddleware) *FseppHandler {
	handler := &FseppHandler{
		cfg: cfg,
		lm:  lm,
	}
	handler.tabData = GetFseppTabData()
	handler.service = service
	return handler
}

func (h *FseppHandler) FseppMain(c *gin.Context) {
	session := domain.GetSessionFromStdContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	h.tabData = setFseppActiveTab(h.tabData, "sekcije-izvori")
	translator := i18n.GetInstance()
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), fseppURLEvidencija, fmt.Sprintf("#%s", fseppTableID), hxValsFseppEvidencija)

	tbl := common.SetTableBasicData(fseppContentTitle, fseppTableID, h.service.GetSekcijeIzvoriTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.Pagination.HxVals = hxValsFseppEvidencija
	tbl.URLGetAll = fseppURLEvidencija
	tbl.URLPrefix = fseppURLEvidencija
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxActionURL = fseppURLSekcijeUnos
	tbl.BtnAdd.HxTarget = "#dialog-content"
	common.SetTableConfig(&tbl, "EVIDENCIJA PRETHODNOG POREZA", fseppURLEvidencija, true, true, false)
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search-input")
	ctx := c.Request.Context()

	err := h.service.GetSekcijeIzvori(ctx, &tbl, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetSekcijeIzvori(ctx, &tbl, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tbl.URLGetAll = fseppURLSekcije
	tbl.URLPrefix = fseppURLSekcije
	tbl.BtnAdd.IsVisible = true
	tbl.BtnAdd.HxActionURL = fseppURLSekcijeUnos
	tmpl_fin.FseppMain(h.tabData, tbl, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
}

func (h *FseppHandler) FseppSekcijeIzvori(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	h.tabData = setFseppActiveTab(h.tabData, "sekcije-izvori")

	session := domain.GetSessionFromStdContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), fseppURLSekcije, fmt.Sprintf("#%s", fseppTableID), "")

	tbl := common.SetTableBasicData(fseppContentTitle, fseppTableID, h.service.GetSekcijeIzvoriTableFields(), "", "", 0, 0, 0, 0, h.cfg)

	common.SetTableConfig(&tbl, "SEKCIJE I IZVORI", fseppURLSekcije, true, true, false)

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search-input")
	ctx := c.Request.Context()

	err := h.service.GetSekcijeIzvori(ctx, &tbl, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetSekcijeIzvori(ctx, &tbl, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tbl.Pagination.HxVals = hxValsFseppEvidencija
	tbl.URLGetAll = fseppURLSekcije
	tbl.URLPrefix = fseppURLSekcije
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.IsVisible = true
	tbl.BtnAdd.HxActionURL = fseppURLSekcijeUnos
	tbl.BtnAdd.HxTarget = "#dialog-content"
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		tmpl.Table(tbl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	} else {
		tmpl_fin.FseppSekcijeIzvori(h.tabData, tbl, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
	}
}

func (h *FseppHandler) FseppSekcijeUnos(c *gin.Context) {
	translator := i18n.GetInstance()
	dialog := domain.Dialog{
		Id: "dialog-fsepp-unos",
	}
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Sačuvaj",
		IsVisible:     true,
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   fseppURLSekcijeSave,
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
	model := domain.Fsepp{}
	csrfToken := common.GetCsrfToken(c)
	tmpl_fin.DialogFseppSekcije(model, dialog, btnSave, btnCancel, btnClose, translator, csrfToken).Render(c.Request.Context(), c.Writer)

}

// FseppSekcijeSave handles the saving of Sekcije and Izvori data
func (h *FseppHandler) FseppSekcijeSave(c *gin.Context) {

}

func (h *FseppHandler) FseppEvidencija(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		btnObrada := common.SetButton("obrada-btn", translator.Button("Obrada"), "fin_obrada", fseppURLEvidencija, "#"+fseppTableID, "innerHTML", "GET", "", hxValsFseppEvidencija, true, common.ClassSaveButton, "")
		btnDelete := common.SetButton("delete-btn", "Obriši", "fin_delete", "", "", "innerHTML", "GET", "", hxValsFseppEvidencija, true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), fseppURLEvidencija, fmt.Sprintf("#%s", fseppTableID), hxValsFseppEvidencija)

		tbl := common.SetTableBasicData(fseppContentTitle, fseppTableID, h.service.GetEvidencijaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsFseppEvidencija
		tbl.URLGetAll = fseppURLEvidencija
		tbl.URLPrefix = fseppURLEvidencija
		common.SetTableConfig(&tbl, "EVIDENCIJA PP", fseppURLEvidencija, false, false, false)

		h.tabData = setFseppActiveTab(h.tabData, "evidencija")
		err := tmpl_fin.FseppEvidencija(h.tabData, tbl, btnObrada, btnDelete, searchInput, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" || requestSource == "btn" {
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			utils.RenderDialogOK(c, "dialog-fsepp-validation-error", common.ErrMsgValidation)
			//common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		searchText := c.Query("query")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", fseppTableID, h.service.GetEvidencijaTableFields(), "", fseppURLEvidencija, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", fseppURLEvidencija, false, false, false)
		tbl.Pagination.HxVals = hxValsFseppEvidencija
		err := h.service.GetEvidencija(ctx, &tbl, true, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			utils.RenderDialogOK(c, "dialog-fsepp-get-total-records", common.ErrMsgGetTotalRecords)
			//common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetEvidencija(ctx, &tbl, false, pageSize, page, odDatuma, doDatuma, searchText)
		if err != nil {
			utils.RenderDialogOK(c, "dialog-fsepp-render-template", common.ErrMsgGetData)
			//common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl.URLGetAll = fseppURLEvidencija
		tbl.URLPrefix = fseppURLEvidencija
		utils.RenderContent(c, tbl)
	}
}

func (h *FseppHandler) FseppSefKpr(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	csrftoken := common.GetCsrfToken(c)
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnImport := common.SetButton("import-btn", "Import", "fin_import", fseppURLSefKprImport, "#"+fseppTableID, "innerHTML", "POST", "", hxValsFseppImport, true, common.ClassSaveButton, "handleDialogResponse")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), fseppURLSefKpr, fmt.Sprintf("#%s", fseppTableID), hxValsFseppImport)

		tbl := common.SetTableBasicData(fseppContentTitle, fseppTableID, h.service.GetSefKprTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsFseppImport
		tbl.URLGetAll = fseppURLSefKpr
		tbl.URLPrefix = fseppURLSefKpr
		common.SetTableConfig(&tbl, "SEF-KPR", fseppURLSefKpr, false, false, false)

		h.tabData = setFseppActiveTab(h.tabData, "sef-kpr")
		err := tmpl_fin.FseppSefKpr(h.tabData, tbl, btnImport, searchInput, gnGod, csrftoken, translator).Render(c.Request.Context(), c.Writer)
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

		tbl := common.SetTableBasicData("", fseppTableID, h.service.GetSefKprTableFields(), "", fseppURLSefKpr, 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsFseppImport
		common.SetTableConfig(&tbl, "", fseppURLSefKpr, true, true, false)
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

func (h *FseppHandler) FseppSefKprImport(c *gin.Context) {
	translator := i18n.GetInstance()
	dialog := domain.Dialog{
		Id:    "dialog-fsepp-import",
		Title: "Import rezultat",
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	btnOk := domain.Button{
		Id:           "btn-ok",
		LabelText:    "OK",
		IsVisible:    true,
		IdDialog:     dialog.Id,
		BtnClass:     common.ClassSaveButton,
		HxOnClick:    "closeDialog",
		HxOnClickArg: dialog.Id,
	}

	// Get CSV content from form data
	fileContent := c.PostForm("file")
	fileType := c.PostForm("file_type")
	filterType := c.PostForm("filter_type")
	odDatuma := c.PostForm("oddatuma")
	doDatuma := c.PostForm("dodatuma")

	if fileContent == "" {
		tmpl.DialogOk("Niste izabrali fajl za import ili je fajl prazan. Molimo izaberite fajl i pokušajte ponovo.", dialog, btnClose, btnOk, translator).Render(c.Request.Context(), c.Writer)
		return
	}
	if fileType != "csv" {
		tmpl.DialogOk("Nepodržani tip fajla. Molimo izaberite CSV fajl i pokušajte ponovo.", dialog, btnClose, btnOk, translator).Render(c.Request.Context(), c.Writer)
		return
	}
	tbl := common.SetTableBasicData("", fseppTableID, h.service.GetSefKprTableFields(), "", fseppURLSefKpr, 0, 0, 0, 0, h.cfg)

	// Default pagination for import results: 50 rows per page, first page
	pageSize := 50
	currentPage := 1

	err := h.service.FseppSefKprImport(c.Request.Context(), &tbl, fileContent, filterType, odDatuma, doDatuma, true, pageSize, currentPage)
	if err != nil {
		tmpl.DialogOk("Došlo je do greške prilikom obrade fajla: "+err.Error(), dialog, btnClose, btnOk, translator).Render(c.Request.Context(), c.Writer)
		return
	}
}

// RegisterRoutes registers the routes for the EPP handler
func (h *FseppHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/fsepp", h.FseppMain)
	r.GET("api/fsepp/sekcije", h.FseppSekcijeIzvori)
	r.POST("api/fsepp/sekcije/save", h.FseppSekcijeSave)
	r.GET("api/fsepp/sekcije/unos", h.FseppSekcijeUnos)
	r.GET("api/fsepp/evidencija", h.FseppEvidencija)
	r.GET("api/fsepp/sefkpr", h.FseppSefKpr)
	r.POST("api/fsepp/sefkpr/import", h.FseppSefKprImport)
}

func GetFseppTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "sekcije-izvori", Label: "Sekcije i izvori", HXRequestUrl: fseppURLSekcije, IsActive: true, Name: "sekcije-izvori"},
			{ID: "evidencija", Label: "Evidencija PP", HXRequestUrl: fseppURLEvidencija, IsActive: false, Name: "evidencija"},
			{ID: "sef-kpr", Label: "SEF-KPR", HXRequestUrl: fseppURLSefKpr, IsActive: false, Name: "sef-kpr"},
		},
	}
}

func setFseppActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
