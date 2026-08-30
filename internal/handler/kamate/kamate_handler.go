package kamate

import (
	"fmt"
	"helia/config"
	"helia/pkg/utils"
	"net/http"
	"time"

	tmpl_kam "helia/frontend/templates/kamate"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	kamateContentTitle           string = "KAMATE"
	kamatneStopeContentTitle     string = "KAMATNE STOPE"
	kamateTableID                string = "kamatetable"
	kamatePartneriTableID        string = "kamate-partneri-table"
	kamateDetaljiTableID         string = "kamate-detalji-table"
	kamateURLPrefix              string = "/api/kamate/"
	kamateURLTipovikamate        string = "/api/kamate/tipovikamate"
	kamateURLStope               string = "/api/kamate/stope"
	kamateURLFormiranje          string = "/api/kamate/formiranje"
	kamateURLFormiranjeDetalji   string = "/api/kamate/formiranje/detalji"
	kamateURLObracun             string = "/api/kamate/obracun"
	tipoviKamateContentTitle     string = "TIPOVI KAMATE"
	tipoviKamateTableID          string = "tipovikamatetable"
	tipoviKamateURLStope         string = "/api/kamate/tipovikamate/stope"
	tipoviKamateURLUnos          string = "/api/kamate/tipovikamate/confirm-add"
	tipoviKamateURLUpdate        string = "/api/kamate/tipovikamate/confirm-update"
	tipoviKamateURLDelete        string = "/api/kamate/tipovikamate/confirm-delete"
	tipoviKamateURLSave          string = "/api/kamate/tipovikamate/save"
	kamatneStopeURLUnos          string = "/api/kamate/stope/confirm-add"
	kamatneStopeURLConfirmUpdate string = "/api/kamate/stope/confirm-update"
	kamatneStopeURLUpdate        string = "/api/kamate/stope/update"
	kamatneStopeURLSave          string = "/api/kamate/stope/save"
	kamatneStopeURLDelete        string = "/api/kamate/stope/delete"
)

const (
	hxValsStope = `js:{
            tip_kamate: document.getElementById("tip_kamate")?.value,
            stopa_od: document.getElementById("stopa_od")?.value,
            stopa_do: document.getElementById("stopa_do")?.value
        }`
	hxValsFormiranjeKamListova = `js:{
            konto: document.getElementById("konto")?.value,
            odsifre: document.getElementById("odsifre")?.value,
            dosifre: document.getElementById("dosifre")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
			prikaz_otvorene: document.getElementById("prikaz_otvorene")?.checked,
			prikaz_zatvorene: document.getElementById("prikaz_zatvorene")?.checked
        }`

	hxValsObracun = `js:{
            od_broja_liste: document.getElementById("od_broja_liste")?.value,
            do_broja_liste: document.getElementById("do_broja_liste")?.value,
            pod_datumom: document.getElementById("pod_datumom")?.value
        }`
)

type KamateHandler struct {
	tabData domain.TabData
	service finservice.KamateService
	cfg     config.Config
	lm      *middleware.LockMiddleware
}

func NewKamateHandler(service finservice.KamateService, cfg config.Config, lm *middleware.LockMiddleware) *KamateHandler {
	handler := &KamateHandler{
		service: service,
		cfg:     cfg,
		lm:      lm,
	}
	handler.tabData = GetKamateTabData()
	return handler
}

func (h *KamateHandler) KamateMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	translator := i18n.GetInstance()
	h.tabData = setKamateActiveTab(h.tabData, "tipovikamate")

	searchInput := common.CreateSearchInput("search-input", translator, kamateURLTipovikamate, fmt.Sprintf("#%s", tipoviKamateTableID), "")
	tbl := common.SetTableBasicData(tipoviKamateContentTitle, tipoviKamateTableID, h.service.GetTipoviKamateTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, tipoviKamateContentTitle, kamateURLTipovikamate, true, true, true)
	tbl.BtnAdd.IsVisible = true
	tbl.BtnAdd.HxActionURL = tipoviKamateURLUnos
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxTarget = "#dialog-tipovi-kamate"
	tbl.BtnAdd.HxSwap = "innerHTML"
	tbl.BtnDelete.IsVisible = true
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnDelete.HxActionURL = tipoviKamateURLDelete
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = tipoviKamateURLUpdate
	tbl.BtnUpdate.IsVisible = true
	tbl.BtnUpdate.HxTarget = "#dialog-tipovi-kamate"
	tbl.BtnUpdate.HxSwap = "innerHTML"
	tbl.URLGetAll = kamateURLTipovikamate
	tbl.URLPrefix = kamateURLTipovikamate

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search")
	ctx := c.Request.Context()

	err := h.service.GetTipoviKamate(ctx, &tbl, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetTipoviKamate(ctx, &tbl, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tmpl_kam.KamateMain(h.tabData, tbl, searchInput, translator).Render(c.Request.Context(), c.Writer)

}
func (h *KamateHandler) TipoveKamate(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	translator := i18n.GetInstance()
	h.tabData = setKamateActiveTab(h.tabData, "tipovikamate")

	searchInput := common.CreateSearchInput("search-input", translator, kamateURLTipovikamate, fmt.Sprintf("#%s", tipoviKamateTableID), "")
	tbl := common.SetTableBasicData(tipoviKamateContentTitle, tipoviKamateTableID, h.service.GetTipoviKamateTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, tipoviKamateContentTitle, kamateURLTipovikamate, true, true, true)
	tbl.BtnAdd.IsVisible = true
	tbl.BtnAdd.HxActionURL = tipoviKamateURLUnos
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxTarget = "#dialog-tipovi-kamate"
	tbl.BtnAdd.HxSwap = "innerHTML"
	tbl.BtnDelete.IsVisible = true
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnDelete.HxActionURL = tipoviKamateURLDelete
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = tipoviKamateURLUpdate
	tbl.BtnUpdate.IsVisible = true
	tbl.BtnUpdate.HxTarget = "#dialog-tipovi-kamate"
	tbl.BtnUpdate.HxSwap = "innerHTML"
	tbl.URLGetAll = kamateURLTipovikamate
	tbl.URLPrefix = kamateURLTipovikamate

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search")
	ctx := c.Request.Context()

	err := h.service.GetTipoviKamate(ctx, &tbl, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetTipoviKamate(ctx, &tbl, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tmpl_kam.TipoviKamate(h.tabData, tbl, searchInput, translator).Render(c.Request.Context(), c.Writer)

}

// TipoviKamateUnos displays the dialog for adding or editing a Tipovi Kamate record
func (h *KamateHandler) TipoviKamateAddUpdate(c *gin.Context) {
	var err error
	translator := i18n.GetInstance()
	dialog := domain.Dialog{
		Id: "dialog-tipovikamate-unos",
	}
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Sačuvaj",
		IsVisible:     true,
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   tipoviKamateURLSave,
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

	id, _ := utils.GetInt64FromQueryRequest(c, "id")
	emptyString := ""
	dto := &domain.Kam{Opis: &emptyString, Model: &emptyString}
	if id > 0 {
		// Load existing record for editing
		dto, err = h.service.GetTipoviKamateByID(c.Request.Context(), id)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData+": "+err.Error())
			return
		}
	}

	tmpl_kam.TipoviKamateDialog(*dto, dialog, btnSave, btnCancel, btnClose, common.GetCsrfToken(c), translator).Render(c.Request.Context(), c.Writer)
}

// TipoviKamateSave handles both creation and update of Tipovi Kamate entries based on the presence of an ID in the URL.
func (h *KamateHandler) TipoviKamateSave(c *gin.Context) {
	var err error
	var kam domain.Kam
	kamID := int64(0)
	if c.PostForm("id") != "" {
		kamID = common.StringToInt64(c.PostForm("id"))
	}
	if err := c.ShouldBind(&kam); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode+": "+err.Error())
		return
	}
	cAction := common.ActionAdd
	if kamID != 0 {
		cAction = common.ActionUpdate
	}

	fieldErrors, err := h.service.ValidateTipoviKamate(c.Request.Context(), &kam, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation+", greska: "+err.Error())
		return
	}
	if len(fieldErrors) != 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	err = h.service.SaveTipoviKamate(c.Request.Context(), &kam, kamID, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)

}

// TipoviKamateUpdate handles both creation and update of Tipovi Kamate entries based on the presence of an ID in the URL.
func (h *KamateHandler) TipoviKamateUpdate(c *gin.Context) {
	var err error
	var kam domain.Kam
	kamID, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID+": "+err.Error())
		return
	}
	if err := c.ShouldBind(&kam); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode+": "+err.Error())
		return
	}

	cAction := common.ActionUpdate

	fieldErrors, err := h.service.ValidateTipoviKamate(c.Request.Context(), &kam, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation+", greska: "+err.Error())
		return
	}
	if len(fieldErrors) != 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	err = h.service.SaveTipoviKamate(c.Request.Context(), &kam, kamID, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)

}

func (h *KamateHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.service.GetTipoviKamateTableFields(), "#info-message")
}

// TipoviKamateDelete handles deletion of a Tipovi Kamate record
func (h *KamateHandler) TipoviKamateDelete(c *gin.Context) {
	kamID, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID+": "+err.Error())
		return
	}

	err = h.service.DeleteTipoviKamate(c.Request.Context(), kamID)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgDeleteData+": "+err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

func (h *KamateHandler) KamatneStope(c *gin.Context) {
	translator := i18n.GetInstance()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	h.tabData = setKamateActiveTab(h.tabData, "kamatnestope")

	searchInput := common.CreateSearchInput("search-input", translator, kamateURLStope, fmt.Sprintf("#%s", kamateTableID), "")
	tbl := common.SetTableBasicData(kamatneStopeContentTitle, kamateTableID, h.service.GetKamatneStopeTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, kamatneStopeContentTitle, kamateURLStope, true, true, true)
	tbl.BtnAdd.IsVisible = true
	tbl.BtnAdd.HxActionURL = kamatneStopeURLUnos
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxTarget = "#dialog-tipovi-kamate"
	tbl.BtnAdd.HxSwap = "innerHTML"
	tbl.BtnDelete.IsVisible = true
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnDelete.HxActionURL = kamatneStopeURLDelete
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = kamatneStopeURLConfirmUpdate
	tbl.BtnUpdate.IsVisible = true
	tbl.BtnUpdate.HxTarget = "#dialog-tipovi-kamate"
	tbl.BtnUpdate.HxSwap = "innerHTML"
	tbl.URLGetAll = kamateURLStope
	tbl.URLPrefix = kamateURLStope

	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search")
	ctx := c.Request.Context()

	err := h.service.GetKamatneStope(ctx, &tbl, true, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetKamatneStope(ctx, &tbl, false, pageSize, page, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	tmpl_kam.KamatneStope(h.tabData, tbl, searchInput, translator).Render(c.Request.Context(), c.Writer)
}

// KamatneStopeUnos displays the dialog for adding or editing a Kamatne stope record
func (h *KamateHandler) confirmKamatneStopeAddUpdate(c *gin.Context) {
	translator := i18n.GetInstance()
	dialog := domain.Dialog{
		Id: "dialog-kamatnestope-unos",
	}
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Sačuvaj",
		IsVisible:     true,
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   kamatneStopeURLSave,
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

	model := &domain.Tkam{}
	id, _ := utils.GetInt64FromQueryRequest(c, "id")
	if id > 0 {
		// Load existing record for editing
		var err error
		model, err = h.service.GetKamatneStopeByID(c.Request.Context(), id)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData+": "+err.Error())
			return
		}
		btnSave.HxActionURL = kamatneStopeURLUpdate + fmt.Sprintf("/%d", id)
		btnSave.HxRequestType = "PUT"
	}

	err := h.service.GetTipKamateOptions(c.Request.Context(), model)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	now := time.Now()
	model.Odd = &now
	model.Dod = &now
	err = tmpl_kam.KamatneStopeDialog(*model, dialog, btnSave, btnCancel, btnClose, common.GetCsrfToken(c), translator).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

// KamatneStopeSave handles both creation and update of Kamatne stope entries
func (h *KamateHandler) KamatneStopeSave(c *gin.Context) {
	var err error
	var dto domain.TkamDTO
	if err := c.ShouldBind(&dto); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode+": "+err.Error())
		return
	}
	// Convert DTO to Tkam struct and parse dates manually
	tkam := domain.Tkam{
		Idkam: &dto.IDkam,
		Kst:   dto.Kst,
	}
	dteOdd := common.StringToDate(dto.Odd, common.HtmlLayout)
	dteDod := common.StringToDate(dto.Dod, common.HtmlLayout)
	tkam.Odd = &dteOdd
	tkam.Dod = &dteDod

	cAction := common.ActionAdd
	if dto.ID != 0 {
		tkam.Idtkam = dto.ID
		cAction = common.ActionUpdate
	}

	filedErrors, err := h.service.ValidateKamatneStope(c.Request.Context(), &tkam, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation+", greska: "+err.Error())
		return
	}
	if len(filedErrors) != 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, filedErrors, common.ErrMsgValidation)
		return
	}
	err = h.service.SaveKamatneStope(c.Request.Context(), &tkam, dto.ID, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

// KamatneStopeUpdate handles both creation and update of Kamatne stope entries based on the presence of an ID in the URL.
func (h *KamateHandler) KamatneStopeUpdate(c *gin.Context) {
	var err error
	var tkam domain.Tkam
	var dto domain.TkamDTO
	tkamID, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID+": "+err.Error())
		return
	}
	if err := c.ShouldBind(&dto); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode+": "+err.Error())
		return
	}
	// Convert DTO to Tkam struct and parse dates manually
	tkam = domain.Tkam{
		Idkam: &dto.IDkam,
		Kst:   dto.Kst,
	}
	dteOdd := common.StringToDate(dto.Odd, common.HtmlLayout)
	dteDod := common.StringToDate(dto.Dod, common.HtmlLayout)
	tkam.Odd = &dteOdd
	tkam.Dod = &dteDod

	if dto.ID != 0 {
		tkam.Idtkam = dto.ID
	}
	cAction := common.ActionUpdate

	fieldErrors, err := h.service.ValidateKamatneStope(c.Request.Context(), &tkam, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation+", greska: "+err.Error())
		return
	}
	if len(fieldErrors) != 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	err = h.service.SaveKamatneStope(c.Request.Context(), &tkam, tkamID, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)

}

func (h *KamateHandler) confirmDeleteKamatneStope(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.service.GetKamatneStopeTableFields(), "#info-message")
}

// KamatneStopeDelete handles deletion of a Kamatne stope record
func (h *KamateHandler) KamatneStopeDelete(c *gin.Context) {
	tkamID, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID+": "+err.Error())
		return
	}

	err = h.service.DeleteKamatneStope(c.Request.Context(), tkamID)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgDeleteData+": "+err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

func (h *KamateHandler) FormiranjeKamatnihListova(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", kamateURLFormiranje, "#"+kamatePartneriTableID, "innerHTML", "GET", "", hxValsFormiranjeKamListova, true, common.ClassSaveButton, "handleBackendResponse")
		btnFormiraj := common.SetButton("form-kamlistova-btn", "Formiraj Kam. Listova", "fin_save", kamateURLFormiranje, "#kamate-detalji", "innerHTML", "POST", "", hxValsFormiranjeKamListova, true, common.ClassAddButton, "")

		searchInput := common.CreateSearchInput("search-input", translator, kamateURLFormiranje, fmt.Sprintf("#%s", kamatePartneriTableID), hxValsFormiranjeKamListova)
		searchInputDok := common.CreateSearchInput("search-input-dok", translator, kamateURLFormiranjeDetalji, fmt.Sprintf("#%s", kamateDetaljiTableID), hxValsFormiranjeKamListova)
		tblPartneri := common.SetTableBasicData(kamateContentTitle, kamatePartneriTableID, h.service.GetFormiranjeListovaPartneriTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblPartneri, "PREGLED PARTNERA", kamateURLFormiranje, false, false, false)
		tblPartneri.URLGetAll = kamateURLFormiranje
		tblPartneri.URLPrefix = kamateURLFormiranje
		tblPartneri.Pagination.HxVals = hxValsFormiranjeKamListova

		tblDokumenta := common.SetTableBasicData(kamateContentTitle, kamateDetaljiTableID, h.service.GetFormiranjeListovaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblDokumenta, "PREGLED DOKUMENATA", kamateURLFormiranje, false, false, false)

		h.tabData = setKamateActiveTab(h.tabData, "formiranje")
		err := tmpl_kam.FormiranjeKamatnihListova(h.tabData, tblPartneri, tblDokumenta, btnObrada, btnFormiraj, searchInput, searchInputDok, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}
	if requestSource == "btn" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"konto", "odsifre", "dosifre", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		params := domain.KamateParameters{
			Konto:           c.Query("konto"),
			OdSifre:         c.Query("odsifre"),
			DoSifre:         c.Query("dosifre"),
			OdDatuma:        c.Query("oddatuma"),
			DoDatuma:        c.Query("dodatuma"),
			PrikazOtvorene:  c.Query("prikaz_otvorene") == "true",
			PrikazZatvorene: c.Query("prikaz_zatvorene") == "true",
		}
		searchText := c.Query("query")
		ctx := c.Request.Context()

		tblPartneri := common.SetTableBasicData("", kamateTableID, h.service.GetFormiranjeListovaPartneriTableFields(), "", kamateURLFormiranje, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tblPartneri, "", kamateURLFormiranje, false, false, false)
		tblPartneri.URLGetAll = kamateURLFormiranje
		tblPartneri.URLPrefix = kamateURLFormiranje
		tblPartneri.Pagination.HxVals = hxValsFormiranjeKamListova
		err := h.service.GetFormiranjeKamatnihListova(ctx, &tblPartneri, true, pageSize, page, params, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetFormiranjeKamatnihListova(ctx, &tblPartneri, false, pageSize, page, params, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tblPartneri)
		return
	}
}

func (h *KamateHandler) FormiranjeKamatnihListovaObrada(c *gin.Context) {
	fieldParameters := []string{"konto", "odsifre", "dosifre", "oddatuma", "dodatuma"}
	fieldsError := common.ValidateRequiredParams(c, fieldParameters)

	if len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
		return
	}
}

func (h *KamateHandler) ObracunKamate(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		tbl := common.SetTableBasicData(kamateContentTitle, kamateTableID, h.service.GetObracunTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "OBRACUN KAMATE", kamateURLObracun, false, false, false)

		h.tabData = setKamateActiveTab(h.tabData, "obracun")
		err := tmpl_kam.ObracunKamate(h.tabData, tbl, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		fieldParameters := []string{"pod_datumom"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		odBrojaListe := c.Query("od_broja_liste")
		doBrojaListe := c.Query("do_broja_liste")
		podDatumom := c.Query("pod_datumom")
		searchText := c.Query("search")
		ctx := c.Request.Context()

		tbl := common.SetTableBasicData("", kamateTableID, h.service.GetObracunTableFields(), "", kamateURLObracun, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", kamateURLObracun, false, false, false)
		err := h.service.GetObracunKamate(ctx, &tbl, true, pageSize, page, odBrojaListe, doBrojaListe, podDatumom, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetObracunKamate(ctx, &tbl, false, pageSize, page, odBrojaListe, doBrojaListe, podDatumom, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// RegisterRoutes registers the routes for the Kamate handler
func (h *KamateHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/kamate", h.KamateMain)
	r.GET("api/kamate/tipovikamate", h.TipoveKamate)
	r.GET("api/kamate/tipovikamate/confirm-add", h.TipoviKamateAddUpdate)
	r.GET("api/kamate/tipovikamate/confirm-update", h.TipoviKamateAddUpdate)
	r.POST("api/kamate/tipovikamate/save", h.TipoviKamateSave)
	r.PUT("api/kamate/tipovikamate/update/:id", h.lm.WithEntityLockHold("kam", "id"), h.TipoviKamateUpdate)
	r.GET("api/kamate/tipovikamate/confirm-delete", h.confirmDeleteHandler)
	r.DELETE("api/kamate/tipovikamate/:id", h.lm.WithEntityLockHold("kam", "id"), h.TipoviKamateDelete)
	r.GET("api/kamate/stope", h.KamatneStope)
	r.GET("api/kamate/stope/confirm-add", h.confirmKamatneStopeAddUpdate)
	r.GET("api/kamate/stope/confirm-update", h.confirmKamatneStopeAddUpdate)
	r.POST("api/kamate/stope/save", h.KamatneStopeSave)
	r.PUT("api/kamate/stope/update/:id", h.lm.WithEntityLockHold("tkam", "id"), h.KamatneStopeUpdate)
	r.GET("api/kamate/stope/confirm-delete", h.confirmDeleteKamatneStope)
	r.GET("api/kamate/formiranje", h.FormiranjeKamatnihListova)
	r.GET("api/kamate/obracun", h.ObracunKamate)
}

func GetKamateTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "tipovikamate", Label: "Tipove kamate", HXRequestUrl: kamateURLTipovikamate, IsActive: true, Name: "tipovikamate"},
			{ID: "kamatnestope", Label: "Kamatne stope", HXRequestUrl: kamateURLStope, IsActive: false, Name: "kamatnestope"},
			{ID: "formiranje", Label: "Formiranje kamatnih listova", HXRequestUrl: kamateURLFormiranje, IsActive: false, Name: "formiranje"},
			{ID: "obracun", Label: "Obracun kamate", HXRequestUrl: kamateURLObracun, IsActive: false, Name: "obracun"},
		},
	}
}

func setKamateActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
