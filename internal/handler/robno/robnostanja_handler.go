package robno

import (
	"context"
	"net/http"

	"helia/config"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	robnoStanjaArtikalTitle               = "Prikaz stanja artikla"
	robnoStanjaArtikalTableID             = "artikal-table"
	robnoStanjaSubsintetickoKontoTitle    = "Saldo subsintetičkog konta"
	robnoStanjaSubsintetickoKontoTableID  = "subsinteticko-konto-table"
	robnoStanjaArtikalGrupaTitle          = "Prikaz stanja artikla po grupi"
	robnoStanjaArtikalGrupaTableID        = "artikal-grupa-table"
	robnoStanjaURLArtikal                 = "/api/robno-stanja/artikal"
	robnoStanjaURLViseArtikala            = "/api/robno-stanja/vise-artikala"
	robnoStanjaURLViseArtikalaSifra       = "/api/robno-stanja/vise-artikala/sifra"
	robnoStanjaURLViseArtikalaSifraStampa = "/api/robno-stanja/vise-artikala/sifra/stampa"
	robnoStanjaURLViseArtikalaGrupa       = "/api/robno-stanja/vise-artikala/grupa"
	robnoStanjaURLViseArtikalaGrupaStampa = "/api/robno-stanja/vise-artikala/grupa/stampa"
	robnoStanjaURLSubsintetickoKonto      = "/api/robno-stanja/saldo-subsintetickog-konta"
	robnoStanjaURLSvodjenjeZaliha         = "/api/robno-stanja/svodjenje-zalihe"
	robnoStanjaURLtotals                  = "/api/robno-stanja/totalvalues"
	robnoStanjaURLArtikalStampa           = "/api/robno-stanja/artikal/stampa"

	hxValsRobnoStanjaArtikal = `js:{
        "magacin": document.getElementById("magacin")?.value,
		"konto": document.getElementById("konto")?.value
    }`
	hxValsRobnoStanjaSubsintetickoKonto = `js:{
        "magacin": document.getElementById("magacin")?.value,
		"konto": document.getElementById("konto")?.value
    }`
)

type RobnoStanjaHandler struct {
	service robnosvc.RobnoStanjaService
	cfg     config.Config
	tabs    *domain.TabData
	subtabs *domain.TabData
}

func NewRobnoStanjaHandler(s robnosvc.RobnoStanjaService, cfg config.Config) *RobnoStanjaHandler {
	return &RobnoStanjaHandler{service: s, cfg: cfg, tabs: robnoStanjaTabs(), subtabs: robnoStanjaSubTabs()}
}

func (h *RobnoStanjaHandler) RobnoStanjaMain(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 0)
	total := domain.SaldaDto{}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tbl := common.SetTableBasicData(robnoStanjaArtikalTitle, robnoStanjaArtikalTableID, h.service.GetPojedinacnogArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaArtikalTitle, robnoStanjaURLArtikal, false, false, false)
	tbl.HasTotals = true
	obrada := common.SetButton("robnostanja-artikl-obrada", "Obradi", "obrada", robnoStanjaURLArtikal, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	stampaj := common.SetButton("robnostanja-artikl-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	if err := tmpl_robno.RobnoStanjaMain(*h.tabs, tbl, magValues, obrada, stampaj, total, robnoStanjaURLtotals, i18n.GetInstance()).Render(c, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoStanjaHandler) PrikazStanjaPojedinacnogArtikla(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	total := domain.SaldaDto{}
	common.SetActiveTab(h.tabs, 0)

	tbl := common.SetTableBasicData(robnoStanjaArtikalTitle, robnoStanjaArtikalTableID, h.service.GetPojedinacnogArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaArtikalTitle, robnoStanjaURLArtikal, false, false, false)
	tbl.HasTotals = true
	btnPrint := common.SetPrintButton("btn-print-artikal", "Štampa", "fin_print", robnoStanjaURLArtikalStampa, "GET", true, common.ClassPrintButton, "magacin,konto")

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	if requestSource == "menu" || requestSource == "tab" {
		//if the call come from menu click or tab click then render the page with parameters and empty table
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoStanjaURLArtikal, "#"+robnoStanjaArtikalTableID, "innerHTML", "GET", "", hxValsRobnoStanjaArtikal, true, common.ClassSaveButton, "handleDialogResponse")
		err := tmpl_robno.RobnoStanjePojedinacnogArtikla(*h.tabs, tbl, magValues, btnObrada, btnPrint, total, robnoStanjaURLtotals, translator).Render(ctx, c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		//validacija input parametre:
		fieldsError := common.ValidateRequiredParams(c, []string{"magacin", "konto"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Invalid magacin value")
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:   magacin,
			Konto:     c.Query("konto"),
			ReportTip: "robnostanjaartikal",
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoStanjaArtikal
		if err := h.service.GetStanjePojedinacnogArtikla(ctx, &tbl, true, pageSize, page, params); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.ShowPagination = false
		utils.RenderContent(c, tbl)
		return
	}
}
func (h *RobnoStanjaHandler) PrikazStanjaViseArtikalaMain(c *gin.Context) {
	ctx := c.Request.Context()
	common.SetActiveTab(h.tabs, 1)
	common.SetActiveTab(h.subtabs, 0)
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tbl := h.table("Prikaz stanja više artikala", "robnostanja-vise-artikala", h.service.GetViseArtikalaTableFields(), robnoStanjaURLViseArtikala)
	button := common.SetButton("robnostanja-vise-obrada", "Obradi", "obrada", robnoStanjaURLViseArtikala, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	printButton := common.SetButton("robnostanja-vise-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	btnEan13 := domain.Button{Id: "robnostanja-vise-ean13", IsVisible: true, LabelText: "Šifra -> EAN13", BtnClass: common.ClassButton}
	btnNalepnice := domain.Button{Id: "robnostanja-vise-nalepnice", IsVisible: true, LabelText: "Nalepnice", BtnClass: common.ClassButton}
	tmpl_robno.RobnoStanjeViseArtikalaMain(*h.tabs, *h.subtabs, tbl, magValues, button, printButton, btnEan13, btnNalepnice, i18n.GetInstance()).Render(ctx, c.Writer)
}

func (h *RobnoStanjaHandler) PrikazStanjaViseArtikalaSifra(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoStanjaURLSubsintetickoKonto, "#"+robnoStanjaSubsintetickoKontoTableID, "innerHTML", "GET", "", hxValsRobnoStanjaSubsintetickoKonto, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := domain.Button{
		Id:            "btn-print-viseartikalasifra",
		IsVisible:     true,
		LabelText:     "Štampa",
		BtnClass:      common.ClassPrintButton,
		HxActionURL:   robnoStanjaURLArtikalStampa,
		DataFields:    "konto,sifra,tipkonta",
		HxRequestType: "GET",
	}

	tbl := common.SetTableBasicData(robnoStanjaSubsintetickoKontoTitle, robnoStanjaSubsintetickoKontoTableID, h.service.GetPojedinacnogArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaSubsintetickoKontoTitle, robnoStanjaURLSubsintetickoKonto, false, false, false)
	h.service.SetDefaultTableData(&tbl)
	common.SetActiveTab(h.subtabs, 0)

	if requestSource == "menu" || requestSource == "tab" {
		translator := i18n.GetInstance()
		//if the call come from menu click or tab click then render the page with parameters and empty table
		magValues, err := h.service.GetMagacinComboValues(c.Request.Context())
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
			return
		}
		btnEan13 := domain.Button{Id: "robnostanja-vise-sifra-ean13", IsVisible: true, LabelText: "Šifra -> EAN13", BtnClass: common.ClassButton}
		btnNalepnice := domain.Button{Id: "robnostanja-vise-sifra-nalepnice", IsVisible: true, LabelText: "Nalepnice", BtnClass: common.ClassButton}
		err = tmpl_robno.RobnoStanjeViseArtikalaSifra(*h.tabs, *h.subtabs, "vise-artikala", "Prikaz stanja više artikala", tbl, magValues, btnObrada, btnPrint, btnEan13, btnNalepnice, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		tipkonta := c.Query("tipkonta")
		fieldParameters := []string{}
		if tipkonta == "" {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Niste izabrali tip konta (sintetika, subsintetika, analitika)")
			return
		}
		//validacija input parametre:

		fieldParameters = []string{"magacin", "konto", "sifra"}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Invalid magacin value")
			return
		}
		konto := c.Query("konto")
		sifra := c.Query("sifra")
		ctx := c.Request.Context()
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:      magacin,
			Konto:        konto,
			SifraArtikla: sifra,
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoStanjaArtikal
		err = h.service.GetStanjaViseArtikalaSifra(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.ShowPagination = false
		utils.RenderContent(c, tbl)
		return
	}
}
func (h *RobnoStanjaHandler) PrikazStanjaViseArtikalaGrupa(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoStanjaURLViseArtikalaGrupa, "#"+robnoStanjaArtikalGrupaTableID, "innerHTML", "GET", "", hxValsRobnoStanjaArtikal, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := domain.Button{
		Id:            "btn-print-viseartikalasifra",
		IsVisible:     true,
		LabelText:     "Štampa",
		BtnClass:      common.ClassPrintButton,
		HxActionURL:   robnoStanjaURLViseArtikalaGrupaStampa,
		DataFields:    "konto,sifra,tipkonta",
		HxRequestType: "GET",
	}

	tbl := common.SetTableBasicData(robnoStanjaArtikalGrupaTitle, robnoStanjaArtikalGrupaTableID, h.service.GetViseArtikalaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaArtikalGrupaTitle, robnoStanjaURLViseArtikalaGrupa, false, false, false)
	h.service.SetDefaultTableData(&tbl)
	common.SetActiveTab(h.subtabs, 1)

	if requestSource == "menu" || requestSource == "tab" {
		translator := i18n.GetInstance()
		//if the call come from menu click or tab click then render the page with parameters and empty table
		magValues, err := h.service.GetMagacinComboValues(c.Request.Context())
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
			return
		}
		btnEan13 := domain.Button{Id: "robnostanja-vise-grupa-ean13", IsVisible: true, LabelText: "Šifra -> EAN13", BtnClass: common.ClassButton}
		btnNalepnice := domain.Button{Id: "robnostanja-vise-grupa-nalepnice", IsVisible: true, LabelText: "Nalepnice", BtnClass: common.ClassButton}
		tmpl_robno.RobnoStanjeViseArtikalGrupa(*h.tabs, *h.subtabs, "vise-artikala", "Prikaz stanja više artikala", tbl, magValues, btnObrada, btnPrint, btnEan13, btnNalepnice, translator).Render(c.Request.Context(), c.Writer)
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		tipkonta := c.Query("tipkonta")
		fieldParameters := []string{}
		if tipkonta == "" {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Niste izabrali tip konta (sintetika, subsintetika, analitika)")
			return
		}
		//validacija input parametre:

		fieldParameters = []string{"magacin", "konto", "sifra"}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Invalid magacin value")
			return
		}
		konto := c.Query("konto")
		sifra := c.Query("sifra")
		ctx := c.Request.Context()
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:      magacin,
			Konto:        konto,
			SifraArtikla: sifra,
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoStanjaArtikal
		err = h.service.GetStanjaViseArtikalaSifra(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.ShowPagination = false
		utils.RenderContent(c, tbl)
		return
	}
}

func (h *RobnoStanjaHandler) PrikazSaldaSubsintetickogKonta(c *gin.Context) {
	// Get our custom header
	total := domain.SaldaDto{}
	requestSource := c.Request.Header.Get("X-Request-Source")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoStanjaURLSubsintetickoKonto, "#"+robnoStanjaSubsintetickoKontoTableID, "innerHTML", "GET", "", hxValsRobnoStanjaSubsintetickoKonto, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := domain.Button{
		Id:            "btn-print-salda",
		IsVisible:     true,
		LabelText:     "Štampa",
		BtnClass:      common.ClassPrintButton,
		HxActionURL:   robnoStanjaURLArtikalStampa,
		DataFields:    "konto,sifra,tipkonta",
		HxRequestType: "GET",
	}
	magValues, err := h.service.GetMagacinComboValues(c.Request.Context())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tbl := common.SetTableBasicData(robnoStanjaSubsintetickoKontoTitle, robnoStanjaSubsintetickoKontoTableID, h.service.GetPojedinacnogArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaSubsintetickoKontoTitle, robnoStanjaURLSubsintetickoKonto, false, false, false)
	h.service.SetDefaultTableData(&tbl)
	common.SetActiveTab(h.tabs, 2)

	if requestSource == "menu" || requestSource == "tab" {
		//if the call come from menu click or tab click then render the page with parameters and empty table
		err := tmpl_robno.RobnoStanjeSubsintetickogKonta(*h.tabs, tbl, magValues, btnObrada, btnPrint, total, robnoStanjaURLtotals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		tipkonta := c.Query("tipkonta")
		fieldParameters := []string{}
		if tipkonta == "" {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Niste izabrali tip konta (sintetika, subsintetika, analitika)")
			return
		}
		//validacija input parametre:

		fieldParameters = []string{"magacin", "konto", "sifra"}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Invalid magacin value")
			return
		}
		konto := c.Query("konto")
		sifra := c.Query("sifra")
		ctx := c.Request.Context()
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:      magacin,
			Konto:        konto,
			SifraArtikla: sifra,
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoStanjaArtikal
		err = h.service.GetStanjaSubsintetickogKonta(ctx, &tbl, false, pageSize, page, params)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.ShowPagination = false
		utils.RenderContent(c, tbl)
		return
	}
}
func (h *RobnoStanjaHandler) SvodjenjeStanjaZaliha(c *gin.Context) {
	common.SetActiveTab(h.tabs, 3)
	magValues, err := h.service.GetMagacinComboValues(c.Request.Context())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tipDokValues, ojValues, err := h.GetComboValues(c.Request.Context())
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	idOrgjed := 0
	if len(ojValues) > 0 {
		idOrgjed = common.StringToInt(ojValues[0].Key)
	}
	mestoTroskaValues, err := h.service.GetMestoTroskaComboValues(c.Request.Context(), idOrgjed)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tbl := h.table("Svođenje stanja zalihe", "robnostanja-svodjenje", h.service.GetSvodjenjeZalihaTableFields(), robnoStanjaURLSvodjenjeZaliha)
	button := common.SetButton("robnostanja-svodjenje-obrada", "Obradi", "obrada", robnoStanjaURLSvodjenjeZaliha, "#"+tbl.TableID, "innerHTML", "GET", "", "", true, common.ClassSaveButton, "")
	printButton := common.SetButton("robnostanja-svodjenje-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	tmpl_robno.RobnoSvodjenjeZaliha(*h.tabs, tbl, magValues, tipDokValues, ojValues, mestoTroskaValues, button, printButton, i18n.GetInstance()).Render(c, c.Writer)
}

func (h *RobnoStanjaHandler) GetComboValues(ctx context.Context) ([]domain.ComboItem, []domain.ComboItem, error) {

	tipDokValues, err := h.service.GetTipDokComboValues(ctx)
	if err != nil {
		return nil, nil, err
	}
	ojValues, err := h.service.GetOjComboValues(ctx)
	if err != nil {
		return nil, nil, err
	}
	return tipDokValues, ojValues, nil
}
func (h *RobnoStanjaHandler) GetMestoTroskaComboValues(c *gin.Context) {
	idOj, _ := utils.GetIntFromQueryRequest(c, "idorgjed")
	mestoTroskaValues, err := h.service.GetMestoTroskaComboValues(c.Request.Context(), idOj)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	c.JSON(http.StatusOK, mestoTroskaValues)
}

func (h *RobnoStanjaHandler) table(title, id string, fields []domain.Fields, url string) domain.TableData {
	tbl := common.SetTableBasicData(title, id, fields, "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, url, false, false, false)
	tbl.HasTotals = true
	return tbl
}

func (h *RobnoStanjaHandler) AddRoutes(r *gin.Engine) {
	// Apply auth middleware to all Robno Stanja routes.
	r.Use(middleware.Auth())

	// Define routes for Robno Stanja.
	r.GET("/api/robno-stanja", h.RobnoStanjaMain)
	r.GET("/api/robno-stanja/artikal", h.PrikazStanjaPojedinacnogArtikla)
	r.GET("/api/robno-stanja/saldo-subsintetickog-konta", h.PrikazSaldaSubsintetickogKonta)
	r.GET("/api/robno-stanja/vise-artikala", h.PrikazStanjaViseArtikalaMain)
	r.GET("/api/robno-stanja/vise-artikala/sifra", h.PrikazStanjaViseArtikalaSifra)
	r.GET("/api/robno-stanja/vise-artikala/grupa", h.PrikazStanjaViseArtikalaGrupa)
	r.GET("/api/robno-stanja/svodjenje-zalihe", h.SvodjenjeStanjaZaliha)
	r.GET("/api/robno-stanja/mesto-troska", h.GetMestoTroskaComboValues)
}

func robnoStanjaTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnostanja-artikl", Label: translator.T("Prikaz stanja pojedinačnog artikla"), HXRequestUrl: robnoStanjaURLArtikal, IsActive: true, Name: "artikl"},
		{ID: "robnostanja-vise", Label: translator.T("Prikaz stanja više artikala"), HXRequestUrl: robnoStanjaURLViseArtikala, Name: "vise-artikala"},
		{ID: "robnostanja-sub", Label: translator.T("Prikaz salda subsintetičkog konta"), HXRequestUrl: robnoStanjaURLSubsintetickoKonto, Name: "subsinteticki-konto"},
		{ID: "robnostanja-svodjenje", Label: translator.T("Svodjenje stanja zalihe"), HXRequestUrl: robnoStanjaURLSvodjenjeZaliha, Name: "svodjenje-zalihe"},
	}}
}
func robnoStanjaSubTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnostanja-vise-sifra", Label: translator.T("Po Sifri"), HXRequestUrl: robnoStanjaURLViseArtikalaSifra, Name: "vise-artikala-sifra"},
		{ID: "robnostanja-vise-grupa", Label: translator.T("Po Grupi"), HXRequestUrl: robnoStanjaURLViseArtikalaGrupa, Name: "vise-artikala-grupa"},
	}}
}
