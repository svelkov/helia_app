package finansijsko

import (
	"fmt"
	"helia/config"
	"helia/pkg/utils"
	"net/http"
	"time"

	tmpl_fin "helia/frontend/templates/finansijsko"
	tmpl_rep "helia/frontend/templates/reports/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	bilansiContentTitle              string = "BILANSI"
	bilansiTableID                   string = "bilansitable"
	bilansizakljucniTableID          string = "zakljucnitable"
	bilansiURLPrefix                 string = "/api/bilansi/"
	bilansiURLZakljucni              string = "/api/bilansi/zakljucni"
	bilansiURLZakljucniObrazacStampa string = "/api/bilansi/zakljucni/obrazacstampa"
	bilansiURLStanja                 string = "/api/bilansi/stanja"
	bilansiURLStanjaStampanje        string = "/api/bilansi/stanja/stampanje"
	bilansiURLStanjaObrazacStampa    string = "/api/bilansi/stanja/obrazacstampa"
	bilansiURLUspeha                 string = "/api/bilansi/uspeha"
	bilansiURLUspehaStampanje        string = "/api/bilansi/uspeha/stampanje"
	bilansiURLUspehaObrazacStampa    string = "/api/bilansi/uspeha/obrazacstampa"
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
            stanjenadan: document.getElementById("stanjenadan")?.value,
			skraceni: document.getElementById("skraceni")?.checked
        }`
	hxValsUspeha = `js:{
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value,
			skraceni: document.getElementById("skraceni")?.checked
        }`
)

type BilansiHandler struct {
	tabData domain.TabData
	service finservice.BilansiService
	cfg     config.Config
	lm      *middleware.LockMiddleware
}

func NewBilansiHandler(service finservice.BilansiService, cfg config.Config, lm *middleware.LockMiddleware) *BilansiHandler {
	handler := &BilansiHandler{
		cfg: cfg,
		lm:  lm,
	}
	handler.tabData = GetBilansiTabData()
	handler.service = service
	return handler
}

func (h *BilansiHandler) BilansiMain(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}
	btnPrint := domain.Button{
		Id:            "btn-print-zakljucni",
		IsVisible:     true,
		LabelText:     "Štampa ",
		BtnClass:      common.ClassPrintButton,
		HxActionURL:   bilansiURLZakljucniObrazacStampa,
		DataFields:    "odkonta,dokonta,odsifre,dosifre,oddatuma,dodatuma,tip_zakljucni,analitickakonta,klasa9,samosaprometom,zabanku",
		HxRequestType: "GET",
	}

	common.SetActiveTab(&h.tabData, "zakljucnilist")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLZakljucni, "#bilansitable", "innerHTML", "GET", "", hxValsZakljucni, true, common.ClassSaveButton, "handleBackendResponse")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), bilansiURLZakljucni, fmt.Sprintf("#%s", bilansiTableID), hxValsZakljucni)

	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetZakljucniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "ZAKLJUCNI LIST", "", false, false, false)
	tbl.HxVals = hxValsZakljucni
	err := tmpl_fin.BilansiMain(h.tabData, tbl, btnObrada, btnPrint, searchInput, i18n.GetInstance(), gnGod).Render(ctx, c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *BilansiHandler) ZakljucniList(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	ctx := c.Request.Context()
	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetZakljucniTableFields(), bilansiURLZakljucni, bilansiURLZakljucni, 0, 0, 0, 0, h.cfg)
	if requestSource == "menu" || requestSource == "tab" {
		btnPrint := domain.Button{
			Id:            "btn-print-zakljucni",
			IsVisible:     true,
			LabelText:     "Štampa ",
			BtnClass:      common.ClassPrintButton,
			HxActionURL:   bilansiURLZakljucniObrazacStampa,
			DataFields:    "odkonta,dokonta,odsifre,dosifre,oddatuma,dodatuma,tip_zakljucni,analitickakonta,klasa9,samosaprometom,zabanku",
			HxRequestType: "GET",
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLZakljucni, "#bilansitable", "innerHTML", "GET", "", hxValsZakljucni, true, common.ClassSaveButton, "handleBackendResponse")
		searchInput := common.CreateSearchInput("search-input", translator, bilansiURLZakljucni, fmt.Sprintf("#%s", bilansiTableID), hxValsZakljucni)

		session := domain.GetSessionFromStdContext(ctx)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		common.SetTableConfig(&tbl, "ZAKLJUCNI LIST", bilansiURLZakljucni, false, false, false)

		common.SetActiveTab(&h.tabData, "zakljucnilist")
		err := tmpl_fin.ZakljucniList(h.tabData, tbl, btnObrada, btnPrint, searchInput, translator, gnGod).Render(ctx, c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		//validacija input parametre:
		params := domain.ZakljucniParams{
			OdKonta:        c.Query("odkonta"),
			DoKonta:        c.Query("dokonta"),
			OdSifre:        c.Query("odsifre"),
			DoSifre:        c.Query("dosifre"),
			OdDatuma:       c.Query("oddatuma"),
			DoDatuma:       c.Query("dodatuma"),
			TipLista:       c.Query("tip_zakljucni"),
			Klasa9:         c.Query("klasa9"),
			SamosaPrometom: c.Query("samosaprometom"),
		}
		fieldParameters := []string{}
		if params.TipLista == "1" {
			fieldParameters = []string{"odkonta", "dokonta", "odsifre", "dosifre", "oddatuma", "dodatuma", "tip_zakljucni", "analitickakonta", "klasa9", "samosaprometom", "zabanku"}
		} else {
			fieldParameters = []string{"oddatuma", "dodatuma", "tip_zakljucni", "analitickakonta", "klasa9", "samosaprometom", "zabanku"}
		}

		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsZakljucni
		err := h.service.GetZakljucniList(ctx, &tbl, params, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetZakljucniList(ctx, &tbl, params, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}
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

func (h *BilansiHandler) ZakljucniListObrazacStampa(c *gin.Context) {
	ctx := c.Request.Context()
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")

	// Format dates for display
	odFmt, doFmt := odDatuma, doDatuma
	if t, err := time.Parse("2006-01-02", odDatuma); err == nil {
		odFmt = t.Format(common.DateLayout)
	}
	if t, err := time.Parse("2006-01-02", doDatuma); err == nil {
		doFmt = t.Format(common.DateLayout)
	}

	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
	}

	params := domain.ZakljucniStampaParams{
		Naziv:           fvrData.Naziv,
		Adresa:          fvrData.Adresa,
		PostBroj:        fvrData.Pobro,
		Mesto:           fvrData.Mesto,
		MaticniBroj:     fvrData.Matbr,
		SifraDelatnosti: fvrData.SifDel,
		OdDatuma:        odFmt,
		DoDatuma:        doFmt,
		DatumObrade:     time.Now().Format(common.DateLayout),
		God:             session.SelectedGod,
		NDuzSint:        h.cfg.NDuzSint,
		TipLista:        c.Query("tip_zakljucni"),
		ZaBanku:         c.Query("zabanku"),
	}

	zakljucniParams := domain.ZakljucniParams{
		OdKonta:        c.Query("odkonta"),
		DoKonta:        c.Query("dokonta"),
		OdSifre:        c.Query("odsifre"),
		DoSifre:        c.Query("dosifre"),
		OdDatuma:       odDatuma,
		DoDatuma:       doDatuma,
		TipLista:       c.Query("tip_zakljucni"),
		Klasa9:         c.Query("klasa9"),
		SamosaPrometom: c.Query("samosaprometom"),
	}
	switch params.TipLista {
	case "1":
		params.ReportTitle = "Zaključni list analitičkih konta"
	case "2":
		params.ReportTitle = "Zaključni list subsintetičkih konta"
	case "3":
		params.ReportTitle = "Zaključni list sintetičkih konta"
	default:
		params.ReportTitle = "Zaključni list"
	}
	tbl := domain.TableData{}
	tblSummary := domain.TableData{}
	if err := h.service.GetZakljucniListZaStampu(ctx, &tbl, &tblSummary, zakljucniParams, h.cfg.NDuzSint); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	translator := i18n.GetInstance()
	tmpl_rep.ZakljucniListObrazacStampa(repParams, params, tbl, tblSummary, translator).Render(ctx, c.Writer)
}

func (h *BilansiHandler) BilansStanja(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	ctx := c.Request.Context()
	searchText := c.Query("query")
	skraceni := c.Query("skraceni") == "true" || c.Query("skraceni") == "1"
	common.SetActiveTab(&h.tabData, "bilansstanja")
	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansStanjaTableFields(), bilansiURLStanja, bilansiURLStanja, 0, 0, 0, 0, h.cfg)
	tbl.BtnExportPDF.IsVisible = true
	tbl.BtnExportExcel.IsVisible = true
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromStdContext(ctx)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		totals := domain.BilansiTotals{}
		common.SetTableConfig(&tbl, "BILANS STANJA", bilansiURLStanja, true, true, false)
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), bilansiURLStanja, fmt.Sprintf("#%s", bilansiTableID), hxValsStanja)
		// get only totals after the get totals records call
		err := h.service.GetBilansStanja(ctx, &tbl, &totals, searchText, skraceni)
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
		err = tmpl_fin.BilansStanja(h.tabData, tbl, totals, searchInput, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		totals := domain.BilansiTotals{}
		err := h.service.GetBilansStanja(ctx, &tbl, &totals, searchText, skraceni)
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
	//ctx := c.Request.Context()
	dialog := domain.Dialog{
		Id:            "bilans-stanja-add-dialog",
		Title:         "Dodaj bilans stanja",
		HxActionURL:   "/api/bilansi/stanja/create",
		HxRequestType: "POST",
	}
	btnSacuvaj := domain.Button{
		Id:        "btn-sacuvaj",
		IsVisible: true,
		LabelText: "Sačuvaj",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassSaveButton,
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
	ctx := c.Request.Context()
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
	_, err = h.service.Add(ctx, entity, common.IDbils, mappedTableFields)
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
		Id:            "bilans-stanja-update-dialog",
		Title:         "Izmeni bilans stanja",
		HxActionURL:   fmt.Sprintf("/api/bilansi/stanja/update/%d", id),
		HxRequestType: "PUT",
	}
	btnSacuvaj := domain.Button{
		Id:        "btn-sacuvaj",
		IsVisible: true,
		LabelText: "Sačuvaj",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassSaveButton,
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
	ctx := c.Request.Context()

	entity, err := h.service.GetByID(ctx, common.IDbils, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	if entity == nil {
		common.WriteJSONResponse(c, http.StatusNotFound, false, nil, common.ErrMsgNotFound)
		return
	}
	tmpl_fin.BilansStanjaForm(csrfToken, *entity, dialog, btnSacuvaj, btnCancel, btnClose, i18n.GetInstance()).Render(ctx, c.Writer)
}
func (h *BilansiHandler) UpdateBilansStanja(c *gin.Context) {
	entity := &domain.Bils{}
	ctx := c.Request.Context()
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
	err = h.service.Update(ctx, entity, common.IDbils, id, mappedTableFields)
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
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	err = h.service.DeleteBilansStanja(c.Request.Context(), id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}
func (h *BilansiHandler) ObradaStampanjeBilansaStanja(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	ctx := c.Request.Context()
	//stanjeNaDan := c.Query("stanjenadan")

	skraceni := c.Query("skraceni") == "true" || c.Query("skraceni") == "1"
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromStdContext(ctx)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		btnPrint := domain.Button{
			Id:            "btn-print-bilansstanja",
			IsVisible:     true,
			LabelText:     "Štampa ",
			BtnClass:      common.ClassPrintButton,
			HxActionURL:   bilansiURLStanjaObrazacStampa,
			DataFields:    "skraceni,pocstanjepg,stanjenadan",
			HxRequestType: "GET",
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLStanjaStampanje, "#bilansitable", "innerHTML", "GET", "", hxValsStanja, true, common.ClassSaveButton, "handleDialogResponse")
		btnExportXML := common.SetButton("exportxml-btn", "Export XML", "exportxml", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassButton, "")
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), bilansiURLStanja, fmt.Sprintf("#%s", bilansiTableID), hxValsStanja)

		//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansStanjaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, bilansiContentTitle, "", false, false, false)
		common.SetActiveTab(&h.tabData, "stampanjebilansstanja")
		common.SetTableConfig(&tbl, "ŠTANPANJE BILANSA STANJA", bilansiURLStanja, false, false, false)
		err := h.service.GetBilansStanjaZaStampu(ctx, &tbl, common.TipStampePreview, skraceni)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
			return
		}
		err = tmpl_fin.StampanjeBilansaStanja(h.tabData, tbl, btnObrada, btnPrint, btnExportXML, searchInput, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		//validacija input parametre:
		fieldParameters := []string{"stanjenadan"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		stanjeNaDan := c.Query("stanjenadan")
		skraceni := c.Query("skraceni") == "true" || c.Query("skraceni") == "1"
		lPGODizPS := c.Query("pocstanjepg") == "true" || c.Query("pocstanjepg") == "1"
		totals := domain.BilansiTotals{}

		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansStanjaTableFields(), "", bilansiURLZakljucni, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLZakljucni, false, false, false)
		err := h.service.GetBilansStanjaObrada(ctx, &tbl, &totals, stanjeNaDan, skraceni, lPGODizPS)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		utils.RenderContent(c, tbl)
	}
}
func (h *BilansiHandler) BilansStanjaObrazacStampa(c *gin.Context) {
	stanjeNaDan := c.Query("stanjenadan")
	skraceni := c.Query("skraceni") == "true" || c.Query("skraceni") == "1"
	ctx := c.Request.Context()
	stanjeNaDanFmt := stanjeNaDan
	if t, err2 := time.Parse("2006-01-02", stanjeNaDan); err2 == nil {
		stanjeNaDanFmt = t.Format(common.DateLayout)
	}
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "portrait",
	}
	params := domain.BilansiStampaParams{
		PIB:             fvrData.PIB,
		Naziv:           fvrData.Naziv,
		Sediste:         fvrData.Mesto,
		MaticniBroj:     fvrData.Matbr,
		SifraDelatnosti: fvrData.SifDel,
		Adresa:          fvrData.Adresa,
		PostBroj:        fvrData.Pobro,
		Mesto:           fvrData.Mesto,
		OdDatuma:        stanjeNaDanFmt,
		Orientation:     "portrait",
	}
	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansStanjaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "STAMPANJE BILANSA STANJA", bilansiURLStanjaStampanje, false, false, false)
	err = h.service.GetBilansStanjaZaStampu(ctx, &tbl, common.TipStampePrint, skraceni)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	translator := i18n.GetInstance()
	tmpl_rep.BilansStanjaObrazacStampa(repParams, params, tbl, translator).Render(ctx, c.Writer)
}

// Bilans Uspeha Functions
func (h *BilansiHandler) BilansUspeha(c *gin.Context) {
	totals := domain.BilansiTotals{}
	requestSource := c.Request.Header.Get("X-Request-Source")
	ctx := c.Request.Context()
	searchText := c.Query("query")
	skraceni := c.Query("skraceni") == "true" || c.Query("skraceni") == "1"
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
		session := domain.GetSessionFromStdContext(ctx)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		common.SetTableConfig(&tbl, "BILANS USPEHA", bilansiURLUspeha, true, true, false)
		searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), bilansiURLUspeha, fmt.Sprintf("#%s", bilansiTableID), hxValsUspeha)

		err := h.service.GetBilansUspeha(ctx, &tbl, &totals, searchText, skraceni)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		err = tmpl_fin.BilansUspeha(h.tabData, tbl, searchInput, &totals, translator, gnGod).Render(ctx, c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		err := h.service.GetBilansUspeha(ctx, &tbl, &totals, searchText, skraceni)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
	}
}
func (h *BilansiHandler) ConfirmAddBilansUspeha(c *gin.Context) {
	dialog := domain.Dialog{
		Id:            "bilans-uspeha-add-dialog",
		Title:         "Dodaj bilans uspeha",
		HxActionURL:   "/api/bilansi/uspeha/create",
		HxRequestType: "POST",
	}
	btnSacuvaj := domain.Button{
		Id:        "btn-sacuvaj",
		IsVisible: true,
		LabelText: "Sačuvaj",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassSaveButton,
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
		Id:            "bilans-uspeha-update-dialog",
		Title:         "Izmeni bilans uspeha",
		HxActionURL:   fmt.Sprintf("/api/bilansi/uspeha/update/%d", id),
		HxRequestType: "PUT",
	}
	btnSacuvaj := domain.Button{
		Id:        "btn-sacuvaj",
		IsVisible: true,
		LabelText: "Sačuvaj",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassSaveButton,
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
	ctx := c.Request.Context()
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
	err = h.service.UpdateBilu(ctx, entity, common.IDbilu, id, mappedTableFields)
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
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	err = h.service.DeleteBilansUspeha(c.Request.Context(), id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}
func (h *BilansiHandler) ObradaStampanjeBilansUspeha(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromStdContext(ctx)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}
		btnPrint := domain.Button{
			Id:            "btn-print-bilansuspeha",
			IsVisible:     true,
			LabelText:     "Štampa ",
			BtnClass:      common.ClassPrintButton,
			HxActionURL:   bilansiURLUspehaObrazacStampa,
			DataFields:    "skraceni,pocstanjepg,oddatuma,dodatuma",
			HxRequestType: "GET",
		}
		tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansUspehaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "STAMPANJE BILANSA USPEHA", bilansiURLUspehaStampanje, false, false, false)
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", bilansiURLUspehaStampanje, "#bilu-print-area", "innerHTML", "GET", "", hxValsUspeha, true, common.ClassSaveButton, "handleDialogResponse")
		btnExportXML := common.SetButton("exportxml-btn", "Export XML", "exportxml", "", "", "", "GET", "", hxValsUspeha, true, common.ClassButton, "handleExportXMLResponse")
		common.SetActiveTab(&h.tabData, "stampanjebilansuspeha")
		err := h.service.GetBilansUspehaZaStampu(ctx, &tbl, common.TipStampePreview)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tmpl_fin.StampanjeBilansaUspeha(h.tabData, tbl, btnObrada, btnPrint, btnExportXML, translator, gnGod).Render(c.Request.Context(), c.Writer)
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		//validacija input parametre:
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		skraceni := c.Query("skraceni") == "true" || c.Query("skraceni") == "1"
		lPGODizPG := c.Query("pocstanjepg") == "true" || c.Query("pocstanjepg") == "1"

		tbl := common.SetTableBasicData("", bilansiTableID, h.service.GetBilansUspehaStampaTableFields(), "", bilansiURLUspehaStampanje, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", bilansiURLUspehaStampanje, false, false, false)
		totals := domain.BilansiTotals{}
		ctx := c.Request.Context()
		err := h.service.GetBilansUspehaObrada(ctx, &tbl, &totals, odDatuma, doDatuma, skraceni, lPGODizPG)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}
func (h *BilansiHandler) BilansUspehaObrazacStampa(c *gin.Context) {
	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")
	ctx := c.Request.Context()
	// Build display-format dates (DD.MM.YYYY)
	odDatumaFmt, doDatumaFmt := odDatuma, doDatuma
	if t, err2 := time.Parse("2006-01-02", odDatuma); err2 == nil {
		odDatumaFmt = t.Format(common.DateLayout)
	}
	if t, err2 := time.Parse("2006-01-02", doDatuma); err2 == nil {
		doDatumaFmt = t.Format(common.DateLayout)
	}

	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "portrait",
	}
	params := domain.BilansiStampaParams{
		PIB:             fvrData.PIB,
		Naziv:           fvrData.Naziv,
		Sediste:         fvrData.Mesto,
		MaticniBroj:     fvrData.Matbr,
		SifraDelatnosti: fvrData.SifDel,
		Adresa:          fvrData.Adresa,
		PostBroj:        fvrData.Pobro,
		Mesto:           fvrData.Mesto,
		OdDatuma:        odDatumaFmt,
		DoDatuma:        doDatumaFmt,
	}
	tbl := common.SetTableBasicData(bilansiContentTitle, bilansiTableID, h.service.GetBilansUspehaStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "STAMPANJE BILANSA USPEHA", bilansiURLUspehaStampanje, false, false, false)
	err = h.service.GetBilansUspehaZaStampu(ctx, &tbl, common.TipStampePrint)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	translator := i18n.GetInstance()
	tmpl_rep.BilansUspehaObrazacStampa(repParams, params, tbl, translator).Render(ctx, c.Writer)
}
func (h *BilansiHandler) ExportXMLBilansUspeha(c *gin.Context) {

}

func (h *BilansiHandler) ExportXMLBilansStanja(c *gin.Context) {

}

// RegisterRoutes registers the routes for the Bilansi handler
func (h *BilansiHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/bilansi", h.BilansiMain)
	r.GET("api/bilansi/zakljucni", h.ZakljucniList)
	r.GET("api/bilansi/zakljucni/obrazacstampa", h.ZakljucniListObrazacStampa)

	// Bilans Stanja routes
	r.GET("api/bilansi/stanja", h.BilansStanja)
	r.GET("api/bilansi/stanja/confirm-add", h.ConfirmAddBilansStanja)
	r.POST("api/bilansi/stanja/create", h.CreateBilansStanja)
	r.GET("api/bilansi/stanja/confirm-update", h.ConfirmUpdateBilansStanja)
	r.PUT("api/bilansi/stanja/update/:id", h.UpdateBilansStanja)
	r.GET("api/bilansi/stanja/confirm-delete", h.ConfirmDeleteBilansStanja)
	r.DELETE("api/bilansi/stanja/:id", h.DeleteBilansStanja)
	r.GET("api/bilansi/stanja/stampanje", h.ObradaStampanjeBilansaStanja)
	r.GET("api/bilansi/stanja/obrazacstampa", h.BilansStanjaObrazacStampa)
	r.GET("api/bilansi/stanja/exportxml", h.ExportXMLBilansStanja)
	// Bilans Uspeha routes
	r.GET("api/bilansi/uspeha", h.BilansUspeha)
	r.GET("api/bilansi/uspeha/confirm-add", h.ConfirmAddBilansUspeha)
	r.POST("api/bilansi/uspeha/create", h.CreateBilansUspeha)
	r.GET("api/bilansi/uspeha/confirm-update", h.ConfirmUpdateBilansUspeha)
	r.PUT("api/bilansi/uspeha/update/:id", h.UpdateBilansUspeha)
	r.GET("api/bilansi/uspeha/confirm-delete", h.ConfirmDeleteBilansUspeha)
	r.DELETE("api/bilansi/uspeha/:id", h.DeleteBilansUspeha)
	r.GET("api/bilansi/uspeha/stampanje", h.ObradaStampanjeBilansUspeha)
	r.GET("api/bilansi/uspeha/obrazacstampa", h.BilansUspehaObrazacStampa)
	r.GET("api/bilansi/uspeha/exportxml", h.ExportXMLBilansUspeha)

}

func GetBilansiTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "zakljucnilist", Label: "Zakljucni list", HXRequestUrl: bilansiURLZakljucni, IsActive: true, Name: "zakljucnilist"},
			{ID: "bilansstanja", Label: "Bilans stanja", HXRequestUrl: bilansiURLStanja, IsActive: false, Name: "bilansstanja"},
			{ID: "stampanjebilansstanja", Label: "Stampanje bilansa stanja", HXRequestUrl: bilansiURLStanjaStampanje, IsActive: false, Name: "stampanjebilansstanja"},
			{ID: "bilansuspeha", Label: "Bilans uspeha", HXRequestUrl: bilansiURLUspeha, IsActive: false, Name: "bilansuspeha"},
			{ID: "stampanjebilansuspeha", Label: "Stampanje bilansa uspeha", HXRequestUrl: bilansiURLUspehaStampanje, IsActive: false, Name: "stampanjebilansuspeha"},
		},
	}
}
