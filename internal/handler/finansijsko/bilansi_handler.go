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
	lm      *middleware.LockMiddleware
}

func NewBilansiHandler(service *finservice.BilansiResource, cfg config.Config, lm *middleware.LockMiddleware) *BilansiHandler {
	handler := &BilansiHandler{
		cfg: cfg,
		lm:  lm,
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
	common.SetActiveTab(&h.tabData, "zakljucnilist")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLZakljucni, "#bilansitable", "innerHTML", "GET", "", hxValsZakljucni, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), bilansiURLZakljucni, fmt.Sprintf("#%s", bilansiTableID), hxValsZakljucni)

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
		searchInput := common.CreateSearchInput("search-input", translator, bilansiURLZakljucni, fmt.Sprintf("#%s", bilansiTableID), hxValsZakljucni)

		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		common.SetTableConfig(&tbl, "ZAKLJUCNI LIST", bilansiURLZakljucni, false, false, false)

		common.SetActiveTab(&h.tabData, "zakljucnilist")
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
		tbl.HasTotals = true
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
	common.SetActiveTab(&h.tabData, "bilansstanja")
	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansStanjaTableFields(), bilansiURLStanja, bilansiURLStanja, 0, 0, 0, 0, h.cfg)
	tbl.BtnExportPDF.IsVisible = true
	tbl.BtnExportExcel.IsVisible = true
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		common.SetTableConfig(&tbl, "BILANS STANJA", bilansiURLStanja, true, true, false)
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), bilansiURLStanja, fmt.Sprintf("#%s", bilansiTableID), hxValsStanja)
		// get only totals after the get totals records call
		err := h.service.GetBilansStanja(c, &tbl)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		tbl.HxVals = hxValsStanja
		tbl.Pagination.HxVals = hxValsStanja
		tbl.URLGetAll = bilansiURLStanja
		common.SetTableButtons(&tbl, bilansiURLStanja)
		tbl.ShowActions = true
		tbl.BtnPrint.IsVisible = false
		tbl.ShowPagination = false
		tbl.HasTotals = true
		err = tmpl_fin.BilansStanja(h.tabData, tbl, searchInput, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		err := h.service.GetBilansStanja(c, &tbl)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		common.SetTableButtons(&tbl, bilansiURLStanja)
		tbl.BtnAdd.IsVisible = true
		tbl.HxVals = hxValsStanja
		tbl.URLGetAll = bilansiURLStanja
		tbl.BtnPrint.IsVisible = false
		tbl.ShowActions = true
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
	}
}
func (h *BilansiHandler) ConfirmAddBilansStanja(c *gin.Context) {
	dialog := domain.Dialog{
		Id:          "bilans-stanja-add-dialog",
		Title:       "Dodaj bilans stanja",
		HxActionURL: "/api/bilansi/stanja/create",
	}
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Sačuvaj",
		HxActionURL:      "/api/bilansi/stanja/create",
		HxRequestType:    "POST",
		IdDialog:         dialog.Id,
		BtnClass:         common.ClassSaveButton,
		HxOnAfterRequest: "handleFormResponse(event, 'POST')",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Odustani",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	csrfToken := common.GetCsrfToken(c)
	tmpl_fin.BilansStanjaForm(csrfToken, domain.Bils{}, dialog, btnSacuvaj, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *BilansiHandler) CreateBilansStanja(c *gin.Context) {
	entity := &domain.Bils{}
	// Parse request body
	if err := c.ShouldBind(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	// Validate required fields
	fieldErrors := h.service.ValidateBilansStanja(entity)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	tableFields, err := utils.GetFieldsFromCacheForUpdate(h.service.GetFieldCache())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	mappedTableFields := h.service.MapEntityToValues(entity, tableFields)
	_, err = h.service.Add(c, entity, common.IDbils, mappedTableFields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData)
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
	//c.Redirect(http.StatusSeeOther, utils.GetRedirectURL(c))
}
func (h *BilansiHandler) ConfirmUpdateBilansStanja(c *gin.Context) {
	id, err := utils.GetInt64FromQueryRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	dialog := domain.Dialog{
		Id:          "bilans-stanja-update-dialog",
		Title:       "Izmeni bilans stanja",
		HxActionURL: fmt.Sprintf("/api/bilansi/stanja/update/%d", id),
	}
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Sačuvaj",
		HxActionURL:      fmt.Sprintf("/api/bilansi/stanja/update/%d", id),
		HxRequestType:    "PUT",
		IdDialog:         dialog.Id,
		HxOnAfterRequest: "handleFormResponse(event, 'PUT')",
		BtnClass:         common.ClassSaveButton,
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Odustani",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	csrfToken := common.GetCsrfToken(c)

	entity, err := h.service.GetByID(c.Request.Context(), common.IDbils, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if entity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgNotFound)
		return
	}
	tmpl_fin.BilansStanjaForm(csrfToken, *entity, dialog, btnSacuvaj, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *BilansiHandler) UpdateBilansStanja(c *gin.Context) {
	entity := &domain.Bils{}
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	// Parse request body
	if err := c.ShouldBind(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	// Validate required fields
	fieldErrors := h.service.ValidateBilansStanja(entity)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	tableFields, err := utils.GetFieldsFromCacheForUpdate(h.service.GetFieldCache())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	mappedTableFields := h.service.MapEntityToValues(entity, tableFields)
	err = h.service.Update(c, entity, common.IDbils, id, mappedTableFields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData)
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
	c.Redirect(http.StatusSeeOther, utils.GetRedirectURL(c))
}
func (h *BilansiHandler) ConfirmDeleteBilansStanja(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.service.GetBilansStanjaTableFields(), "#info-message")
}
func (h *BilansiHandler) DeleteBilansStanja(c *gin.Context) {
	id := c.Param("id")
	h.service.DeleteBilansStanja(c.Request.Context(), id)
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

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", prometURLAnKonta, "#bilansitable", "innerHTML", "GET", "", hxValsZakljucni, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		btnExportXML := common.SetButton("exportxml-btn", "Export XML", "exportxml", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), prometURLAnKonta, fmt.Sprintf("#%s", prometTableID), hxValsZakljucni)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetZakljucniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, prometContentTitle, "", false, false, false)
		common.SetActiveTab(&h.tabData, "stampanjebilansstanja")
		common.SetTableConfig(&tbl, "ŠTANPANJE BILANSA STANJA", bilansiURLStanja, false, false, false)

		err := tmpl_fin.StampanjeBilansaStanja(h.tabData, tbl, btnObrada, btnPrint, btnExportXML, searchInput, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		//validacija input parametre:
		fieldParameters := []string{"bs_stanje_datum"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansStanjaTableFields(), "", bilansiURLZakljucni, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLZakljucni, false, false, false)
		err := h.service.GetBilansStanja(c, &tbl)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		utils.RenderContent(c, tbl)
	}
}

// Bilans Uspeha Functions
func (h *BilansiHandler) BilansUspeha(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	totals := domain.BilansiTotals{}
	translator := i18n.GetInstance()
	common.SetActiveTab(&h.tabData, "bilansuspeha")
	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansUspehaTableFields(), bilansiURLUspeha, bilansiURLUspeha, 0, 0, 0, 0, h.cfg)
	tbl.HxVals = hxValsUspeha
	tbl.Pagination.HxVals = hxValsUspeha
	tbl.URLGetAll = bilansiURLUspeha
	common.SetTableButtons(&tbl, bilansiURLUspeha)
	tbl.ShowActions = true
	tbl.BtnPrint.IsVisible = false
	tbl.HasTotals = true
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		common.SetTableConfig(&tbl, "BILANS USPEHA", bilansiURLUspeha, true, true, false)
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), bilansiURLUspeha, fmt.Sprintf("#%s", bilansiTableID), hxValsUspeha)

		err := h.service.GetBilansUspeha(c, &tbl)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = tmpl_fin.BilansUspeha(h.tabData, tbl, searchInput, &totals, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		err := h.service.GetBilansUspeha(c, &tbl)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
	}
}
func (h *BilansiHandler) ConfirmAddBilansUspeha(c *gin.Context) {
	dialog := domain.Dialog{
		Id:          "bilans-uspeha-add-dialog",
		Title:       "Dodaj bilans uspeha",
		HxActionURL: "/api/bilansi/uspeha/create",
	}
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Sačuvaj",
		HxActionURL:      "/api/bilansi/uspeha/create",
		HxRequestType:    "POST",
		IdDialog:         dialog.Id,
		BtnClass:         common.ClassSaveButton,
		HxOnAfterRequest: "handleFormResponse(event, 'POST')",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Odustani",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	csrfToken := common.GetCsrfToken(c)
	tmpl_fin.BilansUspehaForm(csrfToken, domain.Bilu{}, dialog, btnSacuvaj, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *BilansiHandler) CreateBilansUspeha(c *gin.Context) {
	entity := &domain.Bilu{}
	// Parse request body
	if err := c.ShouldBind(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	// Validate required fields
	fieldErrors := h.service.ValidateBilansUspeha(entity)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	tableFields, err := utils.GetFieldsFromCacheForUpdate(h.service.GetFieldCacheBilu())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	mappedTableFields := h.service.MapEntityToValuesBilu(entity, tableFields)
	_, err = h.service.AddBilu(c, entity, common.IDbilu, mappedTableFields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData)
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}
func (h *BilansiHandler) ConfirmUpdateBilansUspeha(c *gin.Context) {
	id, err := utils.GetInt64FromQueryRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	dialog := domain.Dialog{
		Id:          "bilans-uspeha-update-dialog",
		Title:       "Izmeni bilans uspeha",
		HxActionURL: fmt.Sprintf("/api/bilansi/uspeha/update/%d", id),
	}
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Sačuvaj",
		HxActionURL:      fmt.Sprintf("/api/bilansi/uspeha/update/%d", id),
		HxRequestType:    "PUT",
		IdDialog:         dialog.Id,
		HxOnAfterRequest: "handleFormResponse(event, 'PUT')",
		BtnClass:         common.ClassSaveButton,
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Odustani",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	csrfToken := common.GetCsrfToken(c)

	entity, err := h.service.GetByIDBilu(c.Request.Context(), common.IDbilu, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if entity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgNotFound)
		return
	}
	tmpl_fin.BilansUspehaForm(csrfToken, *entity, dialog, btnSacuvaj, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *BilansiHandler) UpdateBilansUspeha(c *gin.Context) {
	entity := &domain.Bilu{}
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	// Parse request body
	if err := c.ShouldBind(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	// Validate required fields
	fieldErrors := h.service.ValidateBilansUspeha(entity)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	tableFields, err := utils.GetFieldsFromCacheForUpdate(h.service.GetFieldCacheBilu())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	mappedTableFields := h.service.MapEntityToValuesBilu(entity, tableFields)
	err = h.service.UpdateBilu(c, entity, common.IDbilu, id, mappedTableFields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData)
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
	c.Redirect(http.StatusSeeOther, utils.GetRedirectURL(c))
}
func (h *BilansiHandler) ConfirmDeleteBilansUspeha(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.service.GetBilansUspehaTableFields(), "#info-message")
}
func (h *BilansiHandler) DeleteBilansUspeha(c *gin.Context) {
	id := c.Param("id")
	h.service.DeleteBilansUspeha(c.Request.Context(), id)
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
		common.SetActiveTab(&h.tabData, "stampanjebilansuspeha")

		err := tmpl_fin.StampanjeBilansaUspeha(h.tabData, tbl, btnObrada, btnPrint, btnExportXML, translator, gnGod).Render(c.Request.Context(), c.Writer)
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

		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansUspehaTableFields(), "", bilansiURLUspehaAlg, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLUspehaAlg, false, false, false)
		err := h.service.GetBilansUspeha(c, &tbl)
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

	// Bilans Stanja routes
	r.GET("api/bilansi/stanja", h.BilansStanja)
	r.GET("api/bilansi/stanja/confirm-add", h.ConfirmAddBilansStanja)
	r.POST("api/bilansi/stanja/create", h.CreateBilansStanja)
	r.GET("api/bilansi/stanja/confirm-update", h.ConfirmUpdateBilansStanja)
	r.PUT("api/bilansi/stanja/update/:id", h.UpdateBilansStanja)
	r.GET("api/bilansi/stanja/confirm-delete", h.ConfirmDeleteBilansStanja)
	r.DELETE("api/bilansi/stanja/:id", h.DeleteBilansStanja)
	r.GET("api/bilansi/stanja/stampanje", h.StampanjeBilansStanja)

	// Bilans Uspeha routes
	r.GET("api/bilansi/uspeha", h.BilansUspeha)
	r.GET("api/bilansi/uspeha/confirm-add", h.ConfirmAddBilansUspeha)
	r.POST("api/bilansi/uspeha/create", h.CreateBilansUspeha)
	r.GET("api/bilansi/uspeha/confirm-update", h.ConfirmUpdateBilansUspeha)
	r.PUT("api/bilansi/uspeha/update/:id", h.UpdateBilansUspeha)
	r.GET("api/bilansi/uspeha/confirm-delete", h.ConfirmDeleteBilansUspeha)
	r.DELETE("api/bilansi/uspeha/:id", h.DeleteBilansUspeha)
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
