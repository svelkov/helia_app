package robno

import (
	"context"
	"fmt"
	"net/http"

	"helia/config"
	tmpl_rep_rob "helia/frontend/templates/reports/robno"
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
	robnoStanjaArtikalTitle                 = "Prikaz stanja artikla"
	robnoStanjaArtikalTableID               = "robnostanja-artikal-table"
	robnoStanjaSubsintetickoKontoTitle      = "Saldo subsintetičkog konta"
	robnoStanjaSubsintetickoKontoTableID    = "robnostanja-subsinteticko-konto-table"
	robnoStanjaViseArtikalaTitle            = "Prikaz stanja više artikala"
	robnoStanjaViseArtikalaSifraTableID     = "robnostanja-vise-artikala-sifra-table"
	robnoStanjaViseArtikalaGrupaTableID     = "robnostanja-vise-artikala-grupa-table"
	robnoStanjaSvodjenjeTitle               = "Svođenje stanja zalihe"
	robnoStanjaSvodjenjeTableID             = "robnostanja-svodjenje-table"
	robnoStanjaURLPrefix                    = "/api/robno-stanja"
	robnoStanjaURLArtikal                   = robnoStanjaURLPrefix + "/artikal"
	robnoStanjaURLArtikalStampa             = robnoStanjaURLArtikal + "/stampa"
	robnoStanjaURLViseArtikala              = robnoStanjaURLPrefix + "/vise-artikala"
	robnoStanjaURLViseArtikalaSifra         = robnoStanjaURLViseArtikala + "/sifra"
	robnoStanjaURLViseArtikalaSifraStampa   = robnoStanjaURLViseArtikalaSifra + "/stampa"
	robnoStanjaURLViseArtikalaGrupa         = robnoStanjaURLViseArtikala + "/grupa"
	robnoStanjaURLViseArtikalaGrupaStampa   = robnoStanjaURLViseArtikalaGrupa + "/stampa"
	robnoStanjaURLSubsintetickogKonta       = robnoStanjaURLPrefix + "/subsintetickog-konta"
	robnoStanjaURLSubsintetickogKontaStampa = robnoStanjaURLSubsintetickogKonta + "/stampa"
	robnoStanjaURLSvodjenjeZaliha           = robnoStanjaURLPrefix + "/svodjenje-zalihe"
	robnoStanjaURLMestoTroska               = robnoStanjaURLPrefix + "/mesto-troska"
	robnoStanjaURLtotals                    = robnoStanjaURLPrefix + "/totalvalues"
	robnoStanjaInfoMessageDialogID          = "info-message-dialog"

	robnoStanjaPrintFieldsArtikal            = "magacin,konto,sifra"
	robnoStanjaPrintFieldsViseArtikalaSifra  = "magacin,odkonta,dokonta,odsifre,dosifre,finansijskiiznos,artiklisastanjem,artiklibezstanja,prosecnacenastanje,prosecnacenaulaz,zadobavljaca,odmeseca,domeseca"
	robnoStanjaPrintFieldsViseArtikalaGrupa  = "magacin,odgrupe,dogrupe,odsifre,dosifre,finansijskiiznos,novastranapogrupi,odmeseca,domeseca"
	robnoStanjaPrintFieldsSubsintetickoKonto = "magacin,konto"

	hxValsRobnoStanjaArtikal = `js:{
		"sourceTab": "artikl",
		"magacin": document.getElementById("magacin")?.value,
		"konto": document.getElementById("konto")?.value,
		"sifra": document.getElementById("sifra")?.value,
	}`
	hxValsRobnoStanjaViseArtikalaSifra = `js:{
		"sourceTab": "vise-artikala-sifra",
		"magacin": document.getElementById("magacin")?.value,
		"odkonta": document.getElementById("odkonta")?.value,
		"dokonta": document.getElementById("dokonta")?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"finansijskiiznos": document.getElementById("finansijskiiznos")?.checked,
		"artiklisastanjem": document.getElementById("artiklisastanjem")?.checked,
		"artiklibezstanja": document.getElementById("artiklibezstanja")?.checked,
		"prosecnacenastanje": document.getElementById("prosecnacenastanje")?.checked,
		"prosecnacenaulaz": document.getElementById("prosecnacenaulaz")?.checked,
		"zadobavljaca": document.getElementById("zadobavljaca")?.checked,
		"odmeseca": document.getElementById("odmeseca")?.value,
		"domeseca": document.getElementById("domeseca")?.value,
	}`
	hxValsRobnoStanjaViseArtikalaGrupa = `js:{
		"sourceTab": "vise-artikala-grupa",
		"magacin": document.getElementById("magacin")?.value,
		"odgrupe": document.getElementById("odgrupe")?.value,
		"dogrupe": document.getElementById("dogrupe")?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"finansijskiiznos": document.getElementById("finansijskiiznos")?.checked,
		"novastranapogrupi": document.getElementById("novastranapogrupi")?.checked,
		"odmeseca": document.getElementById("odmeseca")?.value,
		"domeseca": document.getElementById("domeseca")?.value,
	}`
	hxValsRobnoStanjaSubsintetickoKonto = `js:{
		"sourceTab": "subsinteticki-konto",
		"magacin": document.getElementById("magacin")?.value,
		"konto": document.getElementById("konto")?.value,
	}`
	hxValsRobnoSvodjenjeZaliha = `js:{
		"sourceTab": "svodjenje-zalihe",
		"magacin": document.getElementById("magacin")?.value,
		"svodjenjenacin": document.querySelector('input[name="svodjenjenacin"]:checked')?.value,
		"odsifre": document.getElementById("odsifre")?.value,
		"dosifre": document.getElementById("dosifre")?.value,
		"obradiartiklesastanjem": document.getElementById("obradiartiklesastanjem")?.checked,
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
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	common.SetActiveTab(h.tabs, 0)
	total := domain.RobnoStanjaTotal{}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tbl := common.SetTableBasicData(robnoStanjaArtikalTitle, robnoStanjaArtikalTableID, h.service.GetPojedinacnogArtiklaTableFields(), robnoStanjaURLArtikal, robnoStanjaURLArtikal, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaArtikalTableID, robnoStanjaURLArtikal, false, false, false)
	tbl.HasTotals = true
	btnObrada := h.obradaButton(robnoStanjaURLArtikal, robnoStanjaArtikalTableID, hxValsRobnoStanjaArtikal)
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoStanjaURLArtikalStampa, "GET", true, common.ClassPrintButton, robnoStanjaPrintFieldsArtikal)
	if err := tmpl_robno.RobnoStanjaMain(*h.tabs, tbl, magValues, btnObrada, btnPrint, total, robnoStanjaURLtotals, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoStanjaHandler) PrikazStanjaPojedinacnogArtikla(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	common.SetActiveTab(h.tabs, 0)
	translator := i18n.GetInstance()
	total := domain.RobnoStanjaTotal{}
	tbl := common.SetTableBasicData(robnoStanjaArtikalTitle, robnoStanjaArtikalTableID, h.service.GetPojedinacnogArtiklaTableFields(), "", robnoStanjaURLArtikal, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaArtikalTableID, robnoStanjaURLArtikal, false, false, false)
	tbl.HasTotals = true
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"magacin", "konto", "sifra"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataConversion, err.Error()))
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:      magacin,
			Konto:        c.Query("konto"),
			SifraArtikla: c.Query("sifra"),
			ReportTip:    "robnostanjaartikal",
			SearchText:   c.Query("query"),
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoStanjaArtikal
		if err := h.service.GetStanjePojedinacnogArtikla(ctx, &tbl, true, pageSize, page, params, common.TipStampePreview); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetStanjePojedinacnogArtikla(ctx, &tbl, false, pageSize, page, params, common.TipStampePreview); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		tbl.HasTotals = true
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	btnObrada := h.obradaButton(robnoStanjaURLArtikal, robnoStanjaArtikalTableID, hxValsRobnoStanjaArtikal)
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoStanjaURLArtikalStampa, "GET", true, common.ClassPrintButton, robnoStanjaPrintFieldsArtikal)
	tmpl_robno.RobnoStanjePojedinacnogArtikla(*h.tabs, tbl, magValues, btnObrada, btnPrint, total, robnoStanjaURLtotals, translator).Render(ctx, c.Writer)
}

func (h *RobnoStanjaHandler) PrikazStanjaArtikalaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, common.ErrMsgUnauthorized)
		return
	}
	magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, err.Error())
		return
	}
	params := domain.RobnoStanjaParams{
		Magacin:      magacin,
		Konto:        c.Query("konto"),
		SifraArtikla: c.Query("sifra"),
		ReportTip:    "robnostanjaartikal",
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tbl := common.SetTableBasicData(robnoStanjaArtikalTitle, robnoStanjaArtikalTableID, h.service.GetPojedinacnogArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetStanjePojedinacnogArtikla(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	totalValues := domain.RobnoStanjaTotal{}
	if err := h.service.GetUkupnaObrada(ctx, &totalValues, params); err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoStanjaArtikalTitle,
		ParameterItems: map[string]domain.ParameterItem{
			"Magacin": {Name: "Magacin", Value: fmt.Sprintf("%d", params.Magacin)},
			"Konto":   {Name: "Konto", Value: params.Konto},
			"Artikal": {Name: "Šifra artikla", Value: params.SifraArtikla},
		},
	}
	tmpl_rep_rob.RobnoStanjaArtikalStampa(repParams, params, totalValues, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}
func (h *RobnoStanjaHandler) RobnoStanjeArtikalUkupnaObrada(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, err.Error())
		return
	}
	totalValues := domain.RobnoStanjaTotal{}

	params := domain.RobnoStanjaParams{
		Magacin:      magacin,
		Konto:        c.Query("konto"),
		SifraArtikla: c.Query("sifra"),
		ReportTip:    "robnostanjaartikal",
	}
	err = h.service.GetUkupnaObrada(ctx, &totalValues, params)
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tmpl_robno.RobnoStanjeArtikalUkupnaObrada(totalValues, i18n.GetInstance()).Render(ctx, c.Writer)
	// Implement the handler logic for RobnoStanjeArtikalUkupnaObrada here
}

func (h *RobnoStanjaHandler) PrikazStanjaViseArtikalaMain(c *gin.Context) {
	h.PrikazStanjaViseArtikalaSifra(c)
}

func (h *RobnoStanjaHandler) PrikazStanjaViseArtikalaSifra(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	common.SetActiveTab(h.tabs, 1)
	common.SetActiveTab(h.subtabs, 0)
	translator := i18n.GetInstance()
	tbl := common.SetTableBasicData(robnoStanjaViseArtikalaTitle, robnoStanjaViseArtikalaSifraTableID, h.service.GetViseArtikalaTableFields(), "", robnoStanjaURLViseArtikalaSifra, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaViseArtikalaSifraTableID, robnoStanjaURLViseArtikalaSifra, false, false, false)
	h.service.SetDefaultTableData(&tbl)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"magacin"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataConversion, err.Error()))
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:            magacin,
			OdKonta:            c.Query("odkonta"),
			DoKonta:            c.Query("dokonta"),
			OdSifre:            c.Query("odsifre"),
			DoSifre:            c.Query("dosifre"),
			OdMeseca:           c.Query("odmeseca"),
			DoMeseca:           c.Query("domeseca"),
			FinansijskiIznos:   c.Query("finansijskiiznos") == "true",
			ArtikliSaStanjem:   c.Query("artiklisastanjem") == "true",
			ArtikliBezStanja:   c.Query("artiklibezstanja") == "true",
			ProsecnaCenaStanje: c.Query("prosecnacenastanje") == "true",
			ProsecnaCenaUlaz:   c.Query("prosecnacenaulaz") == "true",
			ZaDobavljaca:       c.Query("zadobavljaca") == "true",
			ReportTip:          "robnostanjaviseartikalasifra",
			SearchText:         c.Query("query"),
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoStanjaViseArtikalaSifra
		if err := h.service.GetStanjaViseArtikalaSifra(ctx, &tbl, true, pageSize, page, params); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetStanjaViseArtikalaSifra(ctx, &tbl, false, pageSize, page, params); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	btnObrada := h.obradaButton(robnoStanjaURLViseArtikalaSifra, robnoStanjaViseArtikalaSifraTableID, hxValsRobnoStanjaViseArtikalaSifra)
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoStanjaURLViseArtikalaSifraStampa, "GET", true, common.ClassPrintButton, robnoStanjaPrintFieldsViseArtikalaSifra)
	btnEan13 := common.SetButton("robnostanja-vise-sifra-ean13", "Šifra -> EAN13", "sifraean13", "", "", "", "GET", "", "", true, common.ClassButton, "")
	btnNalepnice := common.SetButton("robnostanja-vise-sifra-nalepnice", "Nalepnice", "nalepnice", "", "", "", "GET", "", "", true, common.ClassButton, "")
	if err := tmpl_robno.RobnoStanjeViseArtikalaSifra(*h.tabs, *h.subtabs, "vise-artikala", robnoStanjaViseArtikalaTitle, tbl, magValues, btnObrada, btnPrint, btnEan13, btnNalepnice, translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoStanjaHandler) PrikazStanjaViseArtikalaGrupa(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	common.SetActiveTab(h.tabs, 1)
	common.SetActiveTab(h.subtabs, 1)
	translator := i18n.GetInstance()
	tbl := common.SetTableBasicData(robnoStanjaViseArtikalaTitle, robnoStanjaViseArtikalaGrupaTableID, h.service.GetViseArtikalaTableFields(), "", robnoStanjaURLViseArtikalaGrupa, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaViseArtikalaGrupaTableID, robnoStanjaURLViseArtikalaGrupa, false, false, false)
	h.service.SetDefaultTableData(&tbl)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"magacin"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataConversion, err.Error()))
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:           magacin,
			OdGrupe:           c.Query("odgrupe"),
			DoGrupe:           c.Query("dogrupe"),
			OdSifre:           c.Query("odsifre"),
			DoSifre:           c.Query("dosifre"),
			OdMeseca:          c.Query("odmeseca"),
			DoMeseca:          c.Query("domeseca"),
			FinansijskiIznos:  c.Query("finansijskiiznos") == "true",
			NovaStranaPoGrupi: c.Query("novastranapogrupi") == "true",
			ReportTip:         "robnostanjaviseartikalagrupa",
			SearchText:        c.Query("query"),
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoStanjaViseArtikalaGrupa
		if err := h.service.GetStanjaViseArtikalaGrupa(ctx, &tbl, true, pageSize, page, params); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetStanjaViseArtikalaGrupa(ctx, &tbl, false, pageSize, page, params); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	btnObrada := h.obradaButton(robnoStanjaURLViseArtikalaGrupa, robnoStanjaViseArtikalaGrupaTableID, hxValsRobnoStanjaViseArtikalaGrupa)
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoStanjaURLViseArtikalaGrupaStampa, "GET", true, common.ClassPrintButton, robnoStanjaPrintFieldsViseArtikalaGrupa)
	btnEan13 := common.SetButton("robnostanja-vise-grupa-ean13", "Šifra -> EAN13", "sifraean13", "", "", "", "GET", "", "", true, common.ClassButton, "")
	btnNalepnice := common.SetButton("robnostanja-vise-grupa-nalepnice", "Nalepnice", "nalepnice", "", "", "", "GET", "", "", true, common.ClassButton, "")
	if err := tmpl_robno.RobnoStanjeViseArtikalGrupa(*h.tabs, *h.subtabs, "vise-artikala", robnoStanjaViseArtikalaTitle, tbl, magValues, btnObrada, btnPrint, btnEan13, btnNalepnice, translator).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoStanjaHandler) PrikazSaldaSubsintetickogKonta(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	common.SetActiveTab(h.tabs, 2)
	total := domain.SaldaDto{}
	tbl := common.SetTableBasicData(robnoStanjaSubsintetickoKontoTitle, robnoStanjaSubsintetickoKontoTableID, h.service.GetSubsintetiskogKontaTableFields(), "", robnoStanjaURLSubsintetickogKonta, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaSubsintetickoKontoTableID, robnoStanjaURLSubsintetickogKonta, false, false, false)
	h.service.SetDefaultTableData(&tbl)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"magacin", "konto"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataConversion, err.Error()))
			return
		}
		params := domain.RobnoStanjaParams{
			Magacin:    magacin,
			Konto:      c.Query("konto"),
			ReportTip:  "robnostanjasubsintetickogkonta",
			SearchText: c.Query("query"),
		}

		if err := h.service.GetStanjaSubsintetickogKonta(ctx, &tbl, params); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}

		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	btnObrada := h.obradaButton(robnoStanjaURLSubsintetickogKonta, robnoStanjaSubsintetickoKontoTableID, hxValsRobnoStanjaSubsintetickoKonto)
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", robnoStanjaURLSubsintetickogKontaStampa, "GET", true, common.ClassPrintButton, robnoStanjaPrintFieldsSubsintetickoKonto)
	if err := tmpl_robno.RobnoStanjeSubsintetickogKonta(*h.tabs, tbl, magValues, btnObrada, btnPrint, total, robnoStanjaURLtotals, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *RobnoStanjaHandler) SvodjenjeStanjaZaliha(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	common.SetActiveTab(h.tabs, 3)
	tbl := common.SetTableBasicData(robnoStanjaSvodjenjeTitle, robnoStanjaSvodjenjeTableID, h.service.GetSvodjenjeZalihaTableFields(), "", robnoStanjaURLSvodjenjeZaliha, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoStanjaSvodjenjeTableID, robnoStanjaURLSvodjenjeZaliha, false, false, false)
	if common.IsDataRequest(c) {
		fieldsError := common.ValidateRequiredParams(c, []string{"magacin"})
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		magacin, err := utils.GetIntFromQueryRequest(c, "magacin")
		if err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataConversion, err.Error()))
			return
		}
		idOrgJed, _ := utils.GetIntFromQueryRequest(c, "idorgjed")
		mestoTroskaID, _ := utils.GetIntFromQueryRequest(c, "mestotrid")
		params := domain.RobnoStanjaParams{
			Magacin:                magacin,
			OdSifre:                c.Query("odsifre"),
			DoSifre:                c.Query("dosifre"),
			NacinSvodjenja:         c.Query("svodjenjenacin"),
			ObradiArtikleSaStanjem: c.Query("obradiartiklesastanjem") == "true",
			VrstaNaloga:            c.Query("vrstanaloga"),
			IdOrgJed:               idOrgJed,
			MestoTroskaID:          mestoTroskaID,
			BrojNaloga:             c.Query("brojnaloga"),
			DatumNaloga:            c.Query("datumnaloga"),
			DatumObradeNaloga:      c.Query("datumobradenaloga"),
			OpisKnjizenja:          c.Query("opisknjizenja"),
			ReportTip:              "robnostanjasvodjenjezalihe",
			SearchText:             c.Query("query"),
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoSvodjenjeZaliha
		if err := h.service.GetSvodjenjeZaliha(ctx, &tbl, true, pageSize, page, params); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, common.ErrMsgGetTotalRecords+": "+err.Error())
			return
		}
		if err := h.service.GetSvodjenjeZaliha(ctx, &tbl, false, pageSize, page, params); err != nil {
			utils.RenderDialogOK(c, robnoStanjaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	tipDokValues, ojValues, err := h.GetComboValues(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	idOrgjed := 0
	if len(ojValues) > 0 {
		idOrgjed = common.StringToInt(ojValues[0].Key)
	}
	mestoTroskaValues, err := h.service.GetMestoTroskaComboValues(ctx, idOrgjed)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	btnObrada := h.obradaButton(robnoStanjaURLSvodjenjeZaliha, robnoStanjaSvodjenjeTableID, hxValsRobnoSvodjenjeZaliha)
	btnPrint := common.SetButton("robnostanja-svodjenje-stampa", "Štampaj", "stampa", "", "", "", "GET", "", "", true, common.ClassPrintButton, "")
	if err := tmpl_robno.RobnoSvodjenjeZaliha(*h.tabs, tbl, magValues, tipDokValues, ojValues, mestoTroskaValues, btnObrada, btnPrint, i18n.GetInstance()).Render(ctx, c.Writer); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
	}
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

func (h *RobnoStanjaHandler) obradaButton(url, tableID, vals string) domain.Button {
	return common.SetButton("obrada-btn", "Obrada", "obrada", url, "#"+tableID, "innerHTML", "GET", "", vals, true, common.ClassSaveButton, "handleDialogResponse")
}

func (h *RobnoStanjaHandler) stampaNotImplemented(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Robno stanja stampa jos nije implementirana")
}

func (h *RobnoStanjaHandler) AddRoutes(r *gin.Engine) {
	// Apply auth middleware to all Robno Stanja routes.
	r.Use(middleware.Auth())

	// Define routes for Robno Stanja.
	r.GET("/api/robno-stanja", h.RobnoStanjaMain)
	r.GET("/api/robno-stanja/artikal", h.PrikazStanjaPojedinacnogArtikla)
	r.GET("/api/robno-stanja/artikal/stampa", h.PrikazStanjaArtikalaStampa)
	r.GET("/api/robno-stanja/vise-artikala", h.PrikazStanjaViseArtikalaMain)
	r.GET("/api/robno-stanja/vise-artikala/sifra", h.PrikazStanjaViseArtikalaSifra)
	r.GET("/api/robno-stanja/vise-artikala/sifra/stampa", h.stampaNotImplemented)
	r.GET("/api/robno-stanja/vise-artikala/grupa", h.PrikazStanjaViseArtikalaGrupa)
	r.GET("/api/robno-stanja/vise-artikala/grupa/stampa", h.stampaNotImplemented)
	r.GET("/api/robno-stanja/subsintetickog-konta", h.PrikazSaldaSubsintetickogKonta)
	r.GET("/api/robno-stanja/subsinteticko-konta/stampa", h.stampaNotImplemented)
	r.GET("/api/robno-stanja/svodjenje-zaliha", h.SvodjenjeStanjaZaliha)
	r.GET("/api/robno-stanja/mesto-troska", h.GetMestoTroskaComboValues)
	r.GET("/api/robno-stanja/ukupna-obrada", h.RobnoStanjeArtikalUkupnaObrada)
}

func robnoStanjaTabs() *domain.TabData {
	translator := i18n.GetInstance()
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robnostanja-artikl", Label: translator.T("Prikaz stanja pojedinačnog artikla"), HXRequestUrl: robnoStanjaURLArtikal, IsActive: true, Name: "artikl"},
		{ID: "robnostanja-vise", Label: translator.T("Prikaz stanja više artikala"), HXRequestUrl: robnoStanjaURLViseArtikala, Name: "vise-artikala"},
		{ID: "robnostanja-sub", Label: translator.T("Prikaz salda subsintetičkog konta"), HXRequestUrl: robnoStanjaURLSubsintetickogKonta, Name: "subsinteticki-konto"},
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
