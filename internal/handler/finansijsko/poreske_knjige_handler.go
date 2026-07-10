package finansijsko

import (
	"fmt"
	"helia/config"
	"helia/pkg/utils"
	"net/http"
	"time"

	tmpl_fin "helia/frontend/templates/finansijsko"
	tmpl_rep_fin "helia/frontend/templates/reports/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	poreskeKnjigeContentTitle       string = "PORESKE KNJIGE"
	poreskeKnjigeTableID            string = "poreskeknjigetable"
	poreskeKnjigeURLPrefix          string = "/api/poreskeknjige/"
	poreskeKnjigeURLGetAll          string = "/api/poreskeknjige/all"
	poreskeKnjigeURLIzdatih         string = "/api/poreskeknjige/izdatih"
	poreskeKnjigeURLIzdatihPrint    string = "/api/poreskeknjige/izdatih/print"
	poreskeKnjigeURLPrimljenih      string = "/api/poreskeknjige/primljenih"
	poreskeKnjigeURLPrimljenihPrint string = "/api/poreskeknjige/primljenih/print"
	poreskeKnjigeURLPrijava         string = "/api/poreskeknjige/prijava"
)
const (
	hxValsPoreskaPrijava = `js:{
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value
        }`
	hxValsKirKpr = `js:{
			knjiga: document.getElementById("knjiga")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value
        }`
)

type PoreskeKnjigeHandler struct {
	tabData domain.TabData
	service finservice.PoreskeKnjigeService
	cfg     config.Config
	lm      *middleware.LockMiddleware
}

func NewPoreskeKnjigeHandler(service finservice.PoreskeKnjigeService, cfg config.Config, lm *middleware.LockMiddleware) *PoreskeKnjigeHandler {
	handler := &PoreskeKnjigeHandler{
		cfg: cfg,
		lm:  lm,
	}
	handler.tabData = GetPoreskeKnjigeTabData()
	handler.service = service
	return handler
}

func (h *PoreskeKnjigeHandler) PoreskeKnjigeMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}

	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", poreskeKnjigeURLIzdatih, "#"+poreskeKnjigeTableID, "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassSaveButton, "handleBackendResponse")
	btnPrint := common.SetPrintButton("btn-print-kir", "Štampa", "fin_print", poreskeKnjigeURLIzdatihPrint, "GET", true, common.ClassPrintButton, "knjiga,oddatuma,dodatuma,stampaponalozima,stampaponalozimazbirno")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), poreskeKnjigeURLIzdatih, fmt.Sprintf("#%s", poreskeKnjigeTableID), hxValsKirKpr)

	tbl := common.SetTableBasicData(poreskeKnjigeContentTitle, poreskeKnjigeTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.Pagination.HxVals = hxValsKirKpr
	tbl.URLGetAll = poreskeKnjigeURLIzdatih
	tbl.URLPrefix = poreskeKnjigeURLIzdatih
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxActionURL = poreskeKnjigeURLIzdatih + "/unos"
	tbl.BtnAdd.HxTarget = "#dialog-content"
	common.SetTableConfig(&tbl, "KNJIGA IZDATIH RAČUNA", poreskeKnjigeURLIzdatih, true, true, false)
	// Get knjiga values
	knjigaValues := []domain.ComboItem{}
	err := h.service.GetTipoveKnjigaValues(c.Request.Context(), &knjigaValues, "I")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	err = tmpl_fin.PoreskeKnjigeMain(h.tabData, tbl, btnObrada, btnPrint, searchInput, knjigaValues, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *PoreskeKnjigeHandler) KnjigaIzdatihRacuna(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl := common.SetTableBasicData("", poreskeKnjigeTableID, h.service.GetKirTableFields(), "", poreskeKnjigeURLIzdatih, 0, 0, 0, 0, h.cfg)
	tbl.Pagination.HxVals = hxValsKirKpr
	common.SetTableConfig(&tbl, "KNJIGA IZDATIH RAČUNA", poreskeKnjigeURLIzdatih, true, true, false)
	tbl.Pagination.HxVals = hxValsKirKpr
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxActionURL = poreskeKnjigeURLIzdatih + "/unos"
	tbl.BtnAdd.HxTarget = "#dialog-content"

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", poreskeKnjigeURLIzdatih, "#"+poreskeKnjigeTableID, "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassSaveButton, "handleBackendResponse")
		btnPrint := common.SetPrintButton("btn-print-kir", "Štampa", "fin_print", poreskeKnjigeURLIzdatihPrint, "GET", true, common.ClassPrintButton, "knjiga,oddatuma,dodatuma,stampaponalozima,stampaponalozimazbirno")
		searchInput := common.CreateSearchInput("search-input", translator, poreskeKnjigeURLIzdatih, fmt.Sprintf("#%s", poreskeKnjigeTableID), hxValsKirKpr)
		knjigaValues := []domain.ComboItem{}
		// Get knjiga values
		err := h.service.GetTipoveKnjigaValues(c.Request.Context(), &knjigaValues, "I")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
			return
		}

		h.tabData = setPoreskeKnjigeActiveTab(h.tabData, "izdatih")
		err = tmpl_fin.KnjigaIzdatihRacuna(h.tabData, tbl, btnObrada, btnPrint, searchInput, knjigaValues, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		//validacija input parametre:
		fieldParameters := []string{"knjiga", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.KnjigeParameters{
			Knjiga:     c.Query("knjiga"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			SearchText: c.Query("query"),
		}
		err := h.service.GetKnjigaIzdatihRacuna(c.Request.Context(), &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetKnjigaIzdatihRacuna(c.Request.Context(), &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}
func (h *PoreskeKnjigeHandler) KnjigaIzdatihRacunaUnos(c *gin.Context) {
	translator := i18n.GetInstance()
	dialog := domain.Dialog{
		Id: "dialog-kir-unos",
	}
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Sačuvaj",
		IsVisible:     true,
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   poreskeKnjigeURLIzdatih + "/save",
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

	tipdokValues := []domain.ComboItem{}
	knjigaValues := []domain.ComboItem{}
	err := h.service.GetTipoveKnjigaValues(c.Request.Context(), &knjigaValues, "I")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	err = h.service.GetTipdokValues(c.Request.Context(), &tipdokValues)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tmpl_fin.DialogKirUnos(dialog, knjigaValues, tipdokValues, btnSave, btnCancel, btnClose, common.GetCsrfToken(c), translator).Render(c.Request.Context(), c.Writer)
}

// KnjigaIzdatihRacunaSave handles both creation and update of Kir entries based on the presence of an ID in the URL.
func (h *PoreskeKnjigeHandler) KnjigaIzdatihRacunaSave(c *gin.Context) {
	var err error
	var kir domain.Kir
	kirID := int64(0)
	if c.Query("id") != "" {
		kirID, err = utils.GetInt64FromParameterRequest(c, "id")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID+": "+err.Error())
			return
		}
	}
	if err := c.ShouldBind(&kir); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode+": "+err.Error())
		return
	}
	cAction := common.ActionAdd
	if kirID != 0 {
		cAction = common.ActionUpdate
	}
	kir.IDKir = int(kirID)
	filedErrors, err := h.service.KirValidate(c.Request.Context(), &kir, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgValidation+", greska: "+err.Error())
		return
	}
	if len(filedErrors) != 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, filedErrors, common.ErrMsgValidation)
		return
	}
	err = h.service.SaveKnjigaIzdatihRacuna(c.Request.Context(), &kir, cAction)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData+": "+err.Error())
		return
	}
}

func (h *PoreskeKnjigeHandler) KnjigaPrimljenihRacuna(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", poreskeKnjigeURLPrimljenih, "#"+poreskeKnjigeTableID, "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassSaveButton, "handleBackendResponse")
		btnPrint := common.SetPrintButton("btn-print-kpr", "Štampa", "fin_print", poreskeKnjigeURLPrimljenihPrint, "GET", true, common.ClassPrintButton, "knjiga,oddatuma,dodatuma,stampaponalozima,stampaponalozimazbirno")
		searchInput := common.CreateSearchInput("search-input", translator, poreskeKnjigeURLPrimljenih, fmt.Sprintf("#%s", poreskeKnjigeTableID), hxValsKirKpr)

		tbl := common.SetTableBasicData(poreskeKnjigeContentTitle, poreskeKnjigeTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsKirKpr
		common.SetTableConfig(&tbl, "KNJIGA PRIMLJENIH RAČUNA", poreskeKnjigeURLPrimljenih, true, true, false)
		tbl.URLGetAll = poreskeKnjigeURLPrimljenih
		tbl.URLPrefix = poreskeKnjigeURLPrimljenih
		tbl.Pagination.HxVals = hxValsKirKpr
		tbl.BtnAdd.HxRequestType = "GET"
		tbl.BtnAdd.HxActionURL = poreskeKnjigeURLPrimljenih + "/unos"
		tbl.BtnAdd.HxTarget = "#dialog-content"
		knjigaValues := []domain.ComboItem{}
		// Get knjiga values
		err := h.service.GetTipoveKnjigaValues(c.Request.Context(), &knjigaValues, "U")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
			return
		}

		h.tabData = setPoreskeKnjigeActiveTab(h.tabData, "primljenih")
		err = tmpl_fin.KnjigaPrimljenihRacuna(h.tabData, tbl, btnObrada, btnPrint, searchInput, knjigaValues, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		//validacija input parametre:
		fieldParameters := []string{"knjiga", "oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		params := domain.KnjigeParameters{
			Knjiga:     c.Query("knjiga"),
			OdDatuma:   c.Query("oddatuma"),
			DoDatuma:   c.Query("dodatuma"),
			SearchText: c.Query("query"),
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", poreskeKnjigeTableID, h.service.GetKprTableFields(), "", poreskeKnjigeURLPrimljenih, 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsKirKpr
		common.SetTableConfig(&tbl, "", poreskeKnjigeURLPrimljenih, true, true, false)
		err := h.service.GetKnjigaPrimljenihRacuna(c.Request.Context(), &tbl, true, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetKnjigaPrimljenihRacuna(c.Request.Context(), &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *PoreskeKnjigeHandler) PoreskaPrijava(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	poreskaPrijavaData := &domain.PoreskaPrijavaData{}
	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", poreskeKnjigeURLPrijava, "#poreska-prijava-form", "innerHTML", "GET", "", hxValsPoreskaPrijava, true, common.ClassSaveButton, "")
		btnPrint := common.SetButton("stampa", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
		btnObrada.HxOnClick = "handlePoreskaPrijavaResponse(evt)"
		tbl := common.SetTableBasicData(poreskeKnjigeContentTitle, poreskeKnjigeTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, poreskeKnjigeContentTitle, "", false, false, false)

		h.tabData = setPoreskeKnjigeActiveTab(h.tabData, "prijava")

		err := tmpl_fin.PoreskaPrijava(h.tabData, *poreskaPrijavaData, btnObrada, btnPrint, gnGod, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		err := h.service.GetPoreskaPrijava(c.Request.Context(), poreskaPrijavaData)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
			return
		}
		err = tmpl_fin.PoreskaPrijavaForm(*poreskaPrijavaData, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
}

// KnjigaIzdatihRacunaStampa renders the full printable KIR report.
func (h *PoreskeKnjigeHandler) KnjigaIzdatihRacunaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	fieldParameters := []string{"knjiga", "oddatuma", "dodatuma"}
	if fieldsError := common.ValidateRequiredParams(c, fieldParameters); len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
		return
	}

	params := domain.KnjigeParameters{
		Knjiga:                  c.Query("knjiga"),
		OdDatuma:                c.Query("oddatuma"),
		DoDatuma:                c.Query("dodatuma"),
		StampajPoNalozima:       c.Query("stampaponalozima") == "true",
		StampajPoNalozimaZbirno: c.Query("stampaponalozimazbirno") == "true",
	}

	tbl := common.SetTableBasicData("", poreskeKnjigeTableID, h.service.GetKirStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetKnjigaIzdatihRacunaStampa(ctx, &tbl, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	fmtDate := func(s string) string {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t.Format(common.DateLayout)
		}
		return s
	}

	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  "KNJIGA IZDATIH RAČUNA",
		ParameterItems: map[string]domain.ParameterItem{
			"OdDatuma": {Name: "Za period od", Value: fmtDate(params.OdDatuma)},
			"DoDatuma": {Name: "Do", Value: fmtDate(params.DoDatuma)},
			"Knjiga":   {Name: "Knjiga", Value: params.Knjiga},
		},
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.KnjigaIzlaznihRacunaStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

// KnjigaPrimljenihRacunaStampa renders the full printable KPR report.
func (h *PoreskeKnjigeHandler) KnjigaPrimljenihRacunaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	fieldParameters := []string{"knjiga", "oddatuma", "dodatuma"}
	if fieldsError := common.ValidateRequiredParams(c, fieldParameters); len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
		return
	}

	params := domain.KnjigeParameters{
		Knjiga:                  c.Query("knjiga"),
		OdDatuma:                c.Query("oddatuma"),
		DoDatuma:                c.Query("dodatuma"),
		StampajPoNalozima:       c.Query("stampaponalozima") == "true",
		StampajPoNalozimaZbirno: c.Query("stampaponalozimazbirno") == "true",
	}

	tbl := common.SetTableBasicData("", poreskeKnjigeTableID, h.service.GetKprStampaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetKnjigaPrimljenihRacunaStampa(ctx, &tbl, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetData)
		return
	}

	fmtDate := func(s string) string {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t.Format(common.DateLayout)
		}
		return s
	}

	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  "KNJIGA PRIMLJENIH RAČUNA",
		ParameterItems: map[string]domain.ParameterItem{
			"OdDatuma": {Name: "Za period od", Value: fmtDate(params.OdDatuma)},
			"DoDatuma": {Name: "Do", Value: fmtDate(params.DoDatuma)},
			"Knjiga":   {Name: "Knjiga", Value: params.Knjiga},
		},
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.KnjigaPrimljenihRacunaStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

// RegisterRoutes registers the routes for the PoreskeKnjige handler
func (h *PoreskeKnjigeHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/poreskeknjige", h.PoreskeKnjigeMain)
	r.GET("api/poreskeknjige/izdatih", h.KnjigaIzdatihRacuna)
	r.GET("api/poreskeknjige/izdatih/print", h.KnjigaIzdatihRacunaStampa)
	r.GET("api/poreskeknjige/izdatih/unos", h.KnjigaIzdatihRacunaUnos)
	r.POST("api/poreskeknjige/izdatih/save", h.KnjigaIzdatihRacunaSave)
	r.PUT("api/poreskeknjige/izdatih/save/:id", h.KnjigaIzdatihRacunaSave)
	r.GET("api/poreskeknjige/primljenih", h.KnjigaPrimljenihRacuna)
	r.GET("api/poreskeknjige/primljenih/print", h.KnjigaPrimljenihRacunaStampa)
	r.GET("api/poreskeknjige/prijava", h.PoreskaPrijava)
}

func GetPoreskeKnjigeTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "izdatih", Label: "Knjiga izdatih računa", HXRequestUrl: poreskeKnjigeURLIzdatih, IsActive: true, Name: "izdatih"},
			{ID: "primljenih", Label: "Knjiga primljenih računa", HXRequestUrl: poreskeKnjigeURLPrimljenih, IsActive: false, Name: "primljenih"},
			{ID: "prijava", Label: "Poreska prijava", HXRequestUrl: poreskeKnjigeURLPrijava, IsActive: false, Name: "prijava"},
		},
	}
}

func setPoreskeKnjigeActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
