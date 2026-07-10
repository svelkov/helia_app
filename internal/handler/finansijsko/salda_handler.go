package finansijsko

import (
	"fmt"
	"helia/config"
	"log"
	"net/http"
	"time"

	tmpl_fin "helia/frontend/templates/finansijsko"
	tmpl_rep_fin "helia/frontend/templates/reports/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	saldaContentTitle                 string = "SALDA"
	saldaTableID                      string = "saldatable"
	saldaGrupeKontaTableID            string = "saldagrupekonta-table"
	saldaTablePrelomljenoID           string = "saldatable-prelomljeno"
	saldaPartneriTableID              string = "saldapartneri-table"
	saldaTableKomercijalistiTableID   string = "saldatable-komercijalisti"
	saldaTablerealizacijakomTableID   string = "saldatable-komrealizacija"
	saldaTableKlase5i6AnalitikaID     string = "saldatable-klase5i6analitika"
	saldaTableKlase5i6MTID            string = "saldatable-klase5i6mt"
	saldaURLPojedinacnihKonta         string = "/api/salda/pojedinacnihkonta"
	saldaURLPrefix                    string = "/api/salda/"
	saldaURLSaldaGrupeKonta           string = "/api/salda/grupekonta"
	saldaURLSaldaPartneri             string = "/api/salda/partneri"
	saldaURLpartneriIzborStampe       string = "/api/salda/partneri/izborparametara"
	saldaURLPartneriPrelomljeno       string = "/api/salda/partneriprelomljeno"
	saldaURLPartneriPrelomljenoStampa string = "/api/salda/partneriprelomljeno/stampa"
	saldaURLPartneraPoKontimaStampa   string = "/api/salda/partneri/pokontima/stampa"
	saldaURLKomercijalisti            string = "/api/salda/komercijalisti"
	saldaURLrealizacijakomercijalisti string = "/api/salda/realizacijakomercijalisti"
	saldaURLKlase5i6Analitika         string = "/api/salda/klase56analitika"
	saldaURLKlase5i6AnalitikaStampa   string = "/api/salda/klase56analitika/stampa"
	saldaURLKlase5i6MT                string = "/api/salda/klase56mt"
	saldaURLKlase5i6MTStampa          string = "/api/salda/klase56mt/stampa"
	saldaURLtotals                    string = "/api/salda/totalvalues"
	saldaURLKlase5i6Totals            string = "/api/salda/klase56totalvalues"
	saldaURLPojedinacnihKontaStampa   string = "/api/salda/pojedinacnihkonta/stampa"
)

type SaldaHandler struct {
	tabData domain.TabData
	service finservice.SaldaService
	cfg     config.Config
}

const (
	hxValsSaldaPojedinacnihKonta = `js:{
            "konto": document.getElementById("konto")?.value,
			"sifra": document.getElementById("sifra")?.value,
			"tipkonta": document.querySelector('input[name="tipkonta"]:checked')?.value
        }`
	hxValsSaldaGrupeKonta = `js:{
            "odkonta": document.getElementById("odkonta")?.value,
			"dokonta": document.getElementById("dokonta")?.value,
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"chk_salda_valute": document.querySelector('input[name="chk_salda_valute"]:checked')?.value,
			"cbx_tipizvestaja": document.getElementById("cbx_tipizvestaja")?.value,
			"cbx_klasa": document.getElementById("cbx_klasa")?.value,
			"cbx_odmeseca": document.getElementById("cbx_odmeseca")?.value,
			"cbx_domeseca": document.getElementById("cbx_domeseca")?.value,
			"saldo_filter": document.querySelector('input[name="saldo_filter"]:checked')?.value,
			"odsalda": document.getElementById("odsalda")?.value,
			"dosalda": document.getElementById("dosalda")?.value,
        }`

	hxValsSaldaPartneriPrelomljeno = `js:{"sifra_od": document.getElementById("sifra_od")?.value,
			"sifra_do": document.getElementById("sifra_do")?.value}`

	hxValsSaldaKlase5i6Analitika = `js:{
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
			"klasa": document.querySelector('input[name="klasa"]:checked')?.value	
		}`
	hxValsSaldaKlase5i6MT = `js:{
			"odkonta": document.getElementById("odkonta")?.value,
			"dokonta": document.getElementById("dokonta")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
			"klasa": document.querySelector('input[name="klasa"]:checked')?.value	
		}`
	hxValsSaldaKomercijalisti = `js:{"pod_datumom": document.getElementById("pod_datumom")?.value,
			"od_komercijaliste": document.getElementById("od_komercijaliste")?.value,
			"do_komercijaliste": document.getElementById("do_komercijaliste")?.value}`
	hxValsSaldaRealizacijakomercijalisti = `js:{"od_datuma": document.getElementById("od_datuma")?.value,
			"do_datuma": document.getElementById("do_datuma")?.value,
			"od_komercijaliste": document.getElementById("od_komercijaliste")?.value,
			"do_komercijaliste": document.getElementById("do_komercijaliste")?.value}`
)

func NewSaldaHandler(service finservice.SaldaService, cfg config.Config) *SaldaHandler {
	handler := &SaldaHandler{
		cfg: cfg,
	}
	handler.tabData = GetSaldaTabData()
	handler.service = service
	return handler
}

func (h *SaldaHandler) SaldaMain(c *gin.Context) {
	// Create configuration
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/pojedinacnihkonta", "#saldapojedinacnihkonta-table", "innerHTML", "GET", "#konto", hxValsSaldaPojedinacnihKonta, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := domain.Button{
		Id:            "btn-print-salda",
		IsVisible:     true,
		LabelText:     "Štampa",
		BtnClass:      common.ClassPrintButton,
		HxActionURL:   saldaURLPojedinacnihKontaStampa,
		DataFields:    "konto,sifra,tipkonta",
		HxRequestType: "GET",
	}

	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetPojedKontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	h.service.SetDefaultTableData(&tbl)
	setActiveSaldaTab(&h.tabData, "saldapojedinacnihkonta")
	tbl.ShowActions = false
	tbl.FuncClick = "selectRow"                             // naziv js function for Click
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)" // naziv js function for dblClick
	err := tmpl_fin.SaldaMain(h.tabData, tbl, btnPrint, btnObrada, domain.SaldaDto{}, saldaURLtotals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *SaldaHandler) SaldaPojedinacnihKonta(c *gin.Context) {
	// Get our custom header
	total := domain.SaldaDto{}
	requestSource := c.Request.Header.Get("X-Request-Source")
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLPojedinacnihKonta, "#saldapojedinacnihkonta-table", "innerHTML", "GET", "", hxValsSaldaPojedinacnihKonta, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := domain.Button{
		Id:            "btn-print-salda",
		IsVisible:     true,
		LabelText:     "Štampa",
		BtnClass:      common.ClassPrintButton,
		HxActionURL:   saldaURLPojedinacnihKontaStampa,
		DataFields:    "konto,sifra,tipkonta",
		HxRequestType: "GET",
	}
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetPojedKontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, saldaContentTitle, saldaURLPojedinacnihKonta, false, false, false)
	h.service.SetDefaultTableData(&tbl)
	setActiveSaldaTab(&h.tabData, "saldapojedinacnihkonta")

	if requestSource == "menu" || requestSource == "tab" {
		//if the call come from menu click or tab click then render the page with parameters and empty table
		err := tmpl_fin.SaldaPojedinacnihKonta(h.tabData, tbl, btnObrada, btnPrint, total, saldaURLtotals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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
		switch tipkonta {
		case "1":
			fieldParameters = []string{"konto", "sifra", "tipkonta"}
		case "2", "3":
			fieldParameters = []string{"konto", "tipkonta"}
		}
		// Extract parameters from query
		konto := c.Query("konto")
		sifra := c.Query("sifra")
		ctx := c.Request.Context()
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsSaldaPojedinacnihKonta
		err := h.service.GetSaldaPojedinacnihKonta(ctx, &tbl, false, pageSize, page, konto, sifra, tipkonta)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.ShowPagination = false
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaGrupeKonta(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	tbl := common.SetTableBasicData(saldaContentTitle, saldaGrupeKontaTableID, h.service.GetGrupeKontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLSaldaGrupeKonta, "#saldagrupekonta-table", "innerHTML", "GET", "", hxValsSaldaGrupeKonta, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", saldaURLSaldaGrupeKonta+"/print", "#saldagrupekonta-table", "innerHTML", "GET", "", hxValsSaldaGrupeKonta, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", translator, saldaURLSaldaGrupeKonta, fmt.Sprintf("#%s", saldaGrupeKontaTableID), hxValsSaldaGrupeKonta)
	common.SetTableConfig(&tbl, "", saldaURLSaldaGrupeKonta, false, false, false)
	tbl.Pagination.HxVals = hxValsSaldaGrupeKonta
	if requestSource == "menu" || requestSource == "tab" {
		setActiveSaldaTab(&h.tabData, "saldagrupe")
		err := tmpl_fin.SaldaGrupeKonta(h.tabData, tbl, btnObrada, btnPrint, common.MonthComboItems, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		var fieldParameters []string
		if c.Query("cbx_tipizvestaja") == "analitika" {
			fieldParameters = []string{"odkonta", "dokonta", "odsifre", "dosifre", "chk_salda_valute",
				"cbx_tipizvestaja", "cbx_klasa", "cbx_odmeseca", "cbx_domeseca", "saldo_filter", "odsalda", "dosalda"}
		}
		if c.Query("cbx_tipizvestaja") == "subsintetika" || c.Query("cbx_tipizvestaja") == "sintetika" {
			fieldParameters = []string{"odkonta", "dokonta", "chk_salda_valute",
				"cbx_tipizvestaja", "cbx_klasa", "cbx_odmeseca", "cbx_domeseca", "saldo_filter", "odsalda", "dosalda"}
		}
		// Extract parameters from query
		ctx := c.Request.Context()
		params := domain.SaldaParam{
			OdKonta:         c.Query("odkonta"),
			DoKonta:         c.Query("dokonta"),
			OdSifre:         c.Query("odsifre"),
			DoSifre:         c.Query("dosifre"),
			ChkSaldaValute:  c.Query("chk_salda_valute"),
			CbxTipIzvestaja: c.Query("cbx_tipizvestaja"),
			CbxKlasa:        c.Query("cbx_klasa"),
			CbxOdMeseca:     c.Query("cbx_odmeseca"),
			CbxDoMeseca:     c.Query("cbx_domeseca"),
			SaldoFilter:     c.Query("saldo_filter"),
			OdSalda:         c.Query("odsalda"),
			DoSalda:         c.Query("dosalda"),
		}

		fieldsError := h.service.CheckSaldaGrupeParameters(ctx, fieldParameters, params)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetSaldaGrupeKonta(ctx, &tbl, true, page, pageSize, params, h.cfg.NDuzSint)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSaldaGrupeKonta(ctx, &tbl, false, page, pageSize, params, h.cfg.NDuzSint)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaPartneri(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	setActiveSaldaTab(&h.tabData, "saldapartneri")
	tblPartneri := common.SetTableBasicData("", saldaPartneriTableID, h.service.GetSaldaPartneriTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	translator := i18n.GetInstance()
	searchInput := common.CreateSearchInput("search-input", translator, saldaURLSaldaPartneri, fmt.Sprintf("#%s", saldaPartneriTableID), "")
	common.SetTableConfig(&tblPartneri, "", saldaURLSaldaPartneri, false, false, false)

	ctx := c.Request.Context()
	searchText := c.Query("query")
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	err := h.service.GetSaldaPartneriList(ctx, &tblPartneri, true, currentPage, pageSize, searchText, "", "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetSaldaPartneriList(ctx, &tblPartneri, false, currentPage, pageSize, searchText, "", "")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	btnPrint := common.SetButton("stampa-btn", "Štampa", "stampa", saldaURLpartneriIzborStampe, "#dialog-salda-partneri-stampa", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")
	btnPrint.OpenDialog = true
	tblPartneri.BtnAdd.IsVisible = false
	tblPartneri.BtnPrint = btnPrint
	tblPartneri.BtnPrint.IsVisible = true
	tblPartneri.DetailTarget = "#saldapartneri-detalji"
	tblPartneri.DetailURL = "/api/salda/partneridetails"
	tblPartneri.DetailHxRequestType = "GET"
	tblPartneri.DetailHxSwap = "innerHTML"
	tblPartneri.DetailHxTrigger = "click, change delay:500ms"
	if requestSource == "menu" || requestSource == "tab" {
		err = tmpl_fin.SaldaPartneri(h.tabData, tblPartneri, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// if the request comes from search input or pagination button then render only table content
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		utils.RenderContent(c, tblPartneri)
	}
}

func (h *SaldaHandler) SaldaPartneriDetalji(c *gin.Context) {
	strPartnerID := c.Param("id")
	idPartneri := common.StringToInt64(strPartnerID)
	tblSalda := common.SetTableBasicData("", "saldapartneri-salda-table", h.service.GetSaldaPartneriHeaderTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblSalda, "", saldaURLSaldaPartneri, false, false, false)

	tblDetalji := common.SetTableBasicData("", "saldapartneri-detalji-table", h.service.GetSaldaPartneriDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "", saldaURLSaldaPartneri, false, false, false)
	ctx := c.Request.Context()
	Err := h.service.ProcessSaldaPartneriDetails(ctx, idPartneri, &tblSalda, &tblDetalji, "", "", "")
	if Err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	tmpl_fin.SaldaPartneriDetalji(tblSalda, tblDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *SaldaHandler) SaldaPartneriIzborParametara(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	dialog := domain.Dialog{
		Id:          "dialog-salda-partneri-stampa-dlg",
		Title:       "Štampa - izbor parametara",
		OkText:      "Štampaj",
		CancelText:  "Odustani",
		HxActionURL: saldaURLPartneriPrelomljenoStampa,
		HxTarget:    "#dialog-salda-partneri-stampa",
		HxSwap:      "innerHTML",
	}
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "print", saldaURLPartneraPoKontimaStampa, "GET", true, common.ClassPrintButton, "stampaj-detalje,novi-partner-nova-strana,odsifre,dosifre")
	btnCancel := common.SetButton("btn-cancel", "Odustani", "cancel", "", "#dialog-salda-partneri-stampa", "innerHTML", "GET", "", "", true, common.ClassOdustaniButton, "")
	btnClose := common.SetButton("btn-close", "", "close", "", "#dialog-salda-partneri-stampa", "innerHTML", "GET", "", "", true, common.ClassDialogCloseButton, "")
	btnClose.IdDialog = "dialog-salda-partneri-stampa-dlg"
	btnCancel.IdDialog = "dialog-salda-partneri-stampa-dlg"
	tmpl_fin.SaldaPartneriStampaDialog(dialog, btnPrint, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *SaldaHandler) SaldaPartneriPrelomljeno(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTablePrelomljenoID, h.service.GetSaldaPartneriPrelomljenoTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	searchInput := common.CreateSearchInput("search-input", translator, saldaURLPartneriPrelomljeno, fmt.Sprintf("#%s", saldaTablePrelomljenoID), "")
	common.SetTableConfig(&tbl, "PREGLED SALDA PARTNERA", saldaURLPartneriPrelomljeno, false, false, false)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/partneriprelomljeno", "#saldatable-prelomljeno", "innerHTML", "GET", "", hxValsSaldaPartneriPrelomljeno, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", saldaURLPartneriPrelomljenoStampa, "GET", true, common.ClassPrintButton, "sifra_od,sifra_do")
	setActiveSaldaTab(&h.tabData, "saldapartneriprelomljeno")

	if requestSource == "menu" || requestSource == "tab" {
		err := tmpl_fin.SaldaPartneriPrelomljeno(h.tabData, tbl, btnObrada, btnPrint, searchInput, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		ctx := c.Request.Context()
		searchText := c.Query("query")
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetSaldaPartneriPrelomljeno(ctx, &tbl, true, page, pageSize, searchText, "", "")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSaldaPartneriPrelomljeno(ctx, &tbl, false, page, pageSize, searchText, "", "")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// SaldaPartneriPrelomljenoStampa renders a full-page printable Salda Partnera Prelomljeno report.
func (h *SaldaHandler) SaldaPartneriPrelomljenoStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	sifrOd := c.Query("sifra_od")
	sifraDo := c.Query("sifra_do")

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	repParams := domain.ReportParameters{
		ReportName:  translator.Title("PREGLED STANJA PARTNERA - PRELOMLJENO"),
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		ParameterItems: map[string]domain.ParameterItem{
			"OdPartnera":  {Name: translator.Label("Od partnera"), Value: sifrOd},
			"DoPartnera":  {Name: translator.Label("Do partnera"), Value: sifraDo},
			"StanjeNaDan": {Name: translator.Label("Stanje na dan"), Value: time.Now().Format(common.DateLayout)},
		},
	}

	tbl := common.SetTableBasicData("", "salda-partneri-prelomljeno-stampa", h.service.GetSaldaPartneriPrelomljenoStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetSaldaPartneriPrelomljenoStampa(ctx, &tbl, sifrOd, sifraDo); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.SaldaPartneriPrelomljenoStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

// SaldaPartneraPoKontimaStampa renders a full-page printable Salda Partnera po kontima report.
func (h *SaldaHandler) SaldaPartneraPoKontimaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	sifrOd := c.Query("odsifre")
	sifraDo := c.Query("dosifre")
	stampajDetalje := c.Query("stampajdetalje") == "true"
	noviPartnerNovaSt := c.Query("novipartnernova") == "true"

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	repParams := domain.ReportParameters{
		ReportName:  translator.Title("PREGLED STANJA PARTNERA PO KONTIMA"),
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		ParameterItems: map[string]domain.ParameterItem{
			"OdPartnera":  {Name: translator.Label("Od Partnera"), Value: sifrOd},
			"DoPartnera":  {Name: translator.Label("Do Partnera"), Value: sifraDo},
			"StanjeNaDan": {Name: translator.Label("Stanje na dan"), Value: time.Now().Format(common.DateLayout)},
		},
	}

	tbl := common.SetTableBasicData("", "salda-partnera-pokontima-stampa", h.service.GetSaldaPartneraPoKontimaStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetSaldaPartneraPoKontimaStampa(ctx, &tbl, sifrOd, sifraDo, stampajDetalje); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.SaldaPartneraPoKontimaStampa(repParams, tbl, stampajDetalje, noviPartnerNovaSt, translator).Render(ctx, c.Writer)
}

func (h *SaldaHandler) SaldaKlase5i6Analitika(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	total := domain.SaldaDto{}
	userSession := domain.GetSessionFromStdContext(c.Request.Context())
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "user session not found")
		return
	}
	gnGod := userSession.SelectedGod
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLKlase5i6Analitika, fmt.Sprintf("#%s", saldaTableKlase5i6AnalitikaID), "innerHTML", "GET", "", hxValsSaldaKlase5i6Analitika, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", saldaURLKlase5i6AnalitikaStampa, "GET", true, common.ClassPrintButton, "odsifre,dosifre,oddatuma,dodatuma,klasa")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), saldaURLKlase5i6Analitika, fmt.Sprintf("#%s", saldaTableKlase5i6AnalitikaID), "")

	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableKlase5i6AnalitikaID, h.service.GetSaldaKlase5i6AnalitikaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.Pagination.HxVals = hxValsSaldaKlase5i6Analitika
	common.SetTableConfig(&tbl, "SALDA KLASA 5 i 6 ANALITIKA", saldaURLKlase5i6Analitika, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		setActiveSaldaTab(&h.tabData, "saldaklase56analitika")
		err := tmpl_fin.SaldaKlase5i6analiticki(h.tabData, tbl, btnObrada, btnPrint, searchInput, total, gnGod, saldaURLKlase5i6Totals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"odsifre", "dosifre", "oddatuma", "dodatuma", "klasa"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		ctx := c.Request.Context()
		saldaParam := domain.SaldaParam{
			OdSifre:  c.Query("odsifre"),
			DoSifre:  c.Query("dosifre"),
			OdDatuma: c.Query("oddatuma"),
			DoDatuma: c.Query("dodatuma"),
			Klasa:    c.Query("klasa"),
		}
		searchText := c.Query("query")
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetSaldaKlase5i6Analitika(ctx, &tbl, true, page, pageSize, saldaParam, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSaldaKlase5i6Analitika(ctx, &tbl, false, page, pageSize, saldaParam, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaKlase5i6MT(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	total := domain.SaldaDto{}
	userSession := domain.GetSessionFromStdContext(c.Request.Context())
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "user session not found")
		return
	}
	gnGod := userSession.SelectedGod
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLKlase5i6MT, fmt.Sprintf("#%s", saldaTableKlase5i6MTID), "innerHTML", "GET", "", hxValsSaldaKlase5i6MT, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("stampa-btn", "Štampa", "fin_print", saldaURLKlase5i6MTStampa, "GET", true, common.ClassPrintButton, "odkonta,dokonta,oddatuma,dodatuma,klasa")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), saldaURLKlase5i6MT, fmt.Sprintf("#%s", saldaTableKlase5i6MTID), "")

	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableKlase5i6MTID, h.service.GetSaldaKlase5i6MTTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.Pagination.HxVals = hxValsSaldaKlase5i6MT
	common.SetTableConfig(&tbl, "SALDA KLASA 5 i 6 MT", saldaURLKlase5i6MT, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		setActiveSaldaTab(&h.tabData, "saldaklase56mt")
		err := tmpl_fin.SaldaKlase5i6MT(h.tabData, tbl, btnObrada, btnPrint, searchInput, total, gnGod, saldaURLKlase5i6Totals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"odkonta", "dokonta", "oddatuma", "dodatuma", "klasa"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		ctx := c.Request.Context()
		searchText := c.Query("query")
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		saldaParam := domain.SaldaParam{
			OdSifre:  c.Query("odsifre"),
			DoSifre:  c.Query("dosifre"),
			OdDatuma: c.Query("oddatuma"),
			DoDatuma: c.Query("dodatuma"),
			Klasa:    c.Query("klasa"),
		}
		err := h.service.SaldaKlase5i6MT(ctx, &tbl, true, page, pageSize, saldaParam, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.SaldaKlase5i6MT(ctx, &tbl, false, page, pageSize, saldaParam, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaKomercijalisti(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(c.Request.Context())
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "user session not found")
		return
	}
	gnGod := userSession.SelectedGod
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetKomercijalistiTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	searchInput := common.CreateSearchInput("search-input", translator, saldaURLKomercijalisti, fmt.Sprintf("#%s", saldaTableKomercijalistiTableID), "")
	common.SetTableConfig(&tbl, "SALDA PO KOMERCIJALISTIMA", saldaURLKomercijalisti, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLKomercijalisti, "#saldatable-komercijalisti", "innerHTML", "GET", "", hxValsSaldaKomercijalisti, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampaj", "fin_print", saldaURLKomercijalisti+"/print", "#saldatable-komercijalisti", "innerHTML", "GET", "", hxValsSaldaKomercijalisti, true, common.ClassPrintButton, "")
		setActiveSaldaTab(&h.tabData, "saldakomercijalisti")
		err := tmpl_fin.SaldaPoKomercijalistima(h.tabData, tbl, btnObrada, btnPrint, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"pod_datumom", "od_komercijaliste", "do_komercijaliste"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		ctx := c.Request.Context()
		searchText := c.Query("query")
		sortBy := c.Query("sortBy")
		sortOrder := c.Query("sortOrder")
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.SaldaPoKomercijalistima(ctx, &tbl, true, page, pageSize, searchText, sortBy, sortOrder)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.SaldaPoKomercijalistima(ctx, &tbl, false, page, pageSize, searchText, sortBy, sortOrder)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) RealizacijaKomercijalisti(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(c.Request.Context())
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "user session not found")
		return
	}
	gnGod := userSession.SelectedGod
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetRealizacijaKomercijalistiTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	searchInput := common.CreateSearchInput("search-input", translator, saldaURLrealizacijakomercijalisti, fmt.Sprintf("#%s", saldaTablePrelomljenoID), "")
	common.SetTableConfig(&tbl, "REALIZACIJA PO KOMERCIJALISTIMA", saldaURLrealizacijakomercijalisti, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLrealizacijakomercijalisti, "#realizacija-komercijalisti-table", "innerHTML", "GET", "", hxValsSaldaRealizacijakomercijalisti, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampaj", "fin_print", saldaURLrealizacijakomercijalisti+"/print", "#realizacija-komercijalisti-table", "innerHTML", "GET", "", hxValsSaldaRealizacijakomercijalisti, true, common.ClassPrintButton, "")

		setActiveSaldaTab(&h.tabData, "realizacijakomercijalisti")
		err := tmpl_fin.RealizacijaKomercijalisti(h.tabData, tbl, btnObrada, btnPrint, searchInput, gnGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"od_datuma", "do_datuma", "od_komercijaliste", "do_komercijaliste"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		ctx := c.Request.Context()
		searchText := c.Query("query")
		sortBy := c.Query("sortBy")
		sortOrder := c.Query("sortOrder")
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.RealizacijaKomercijalisti(ctx, &tbl, true, page, pageSize, searchText, sortBy, sortOrder)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.RealizacijaKomercijalisti(ctx, &tbl, false, page, pageSize, searchText, sortBy, sortOrder)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) Klase5i6TotalValues(c *gin.Context) {
	totalValues, err := h.service.GetSaldaKlase5i6TotalValues(c.Request.Context())
	if err != nil {
		log.Printf("Error getting total values for klase 5 i 6: %v", err)
	}
	tmpl_fin.SaldaTotalValues(totalValues, saldaURLKlase5i6Totals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *SaldaHandler) TotalValues(c *gin.Context) {
	konto := c.Query("konto")
	sifra := c.Query("sifra")
	tipKonta := c.Query("tipkonta")
	if konto == "" || tipKonta == "" {
		tmpl_fin.SaldaTotalValues(domain.SaldaDto{}, saldaURLtotals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		return
	}
	totalValues, err := h.service.GetSaldaTotalValues(c.Request.Context(), konto, sifra, tipKonta)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}

	tmpl_fin.SaldaTotalValues(totalValues, saldaURLtotals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

// SaldaPojedinacnihKontaStampa renders a full-page printable Salda pojedinačnih konta report.
func (h *SaldaHandler) SaldaPojedinacnihKontaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	konto := c.Query("konto")
	sifra := c.Query("sifra")
	tipkonta := c.Query("tipkonta")

	tipkontaLabel := ""
	switch tipkonta {
	case "1":
		tipkontaLabel = translator.Label("Analitika")
	case "2":
		tipkontaLabel = translator.Label("Subsintetika")
	case "3":
		tipkontaLabel = translator.Label("Sintetika")
	}

	// Fetch totals
	total, err := h.service.GetSaldaTotalValues(ctx, konto, sifra, tipkonta)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	// Fetch monthly table data (all records, no pagination)
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetPojedKontaTableFields(), "", "", 999999, 1, 1, 0, h.cfg)
	if err := h.service.GetSaldaPojedinacnihKonta(ctx, &tbl, false, 999999, 1, konto, sifra, tipkonta); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
	repParams := domain.ReportParameters{
		ReportName:     translator.Title("SALDA POJEDINAČNIH KONTA"),
		Orientation:    "portrait",
		CompanyName:    fvrData.Naziv,
		Adress:         fvrData.Adresa,
		Postcode:       fvrData.Pobro,
		City:           fvrData.Mesto,
		ParameterItems: make(map[string]domain.ParameterItem),
	}
	if konto != "" {
		repParams.ParameterItems["Konto"] = domain.ParameterItem{Name: translator.Label("Konto"), Value: konto}
	}
	if sifra != "" {
		repParams.ParameterItems["Sifra"] = domain.ParameterItem{Name: translator.Label("Šifra"), Value: sifra}
	}
	if tipkontaLabel != "" {
		repParams.ParameterItems["TipKonta"] = domain.ParameterItem{Name: translator.Label("Tip konta"), Value: tipkontaLabel}
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.SaldaPojedinacnihKontaStampa(tbl, total, repParams, translator).Render(ctx, c.Writer)
}

func (h *SaldaHandler) SaldaKlase5i6AnalitikaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	saldaParam := domain.SaldaParam{
		OdSifre:  c.Query("odsifre"),
		DoSifre:  c.Query("dosifre"),
		OdDatuma: c.Query("oddatuma"),
		DoDatuma: c.Query("dodatuma"),
		Klasa:    c.Query("klasa"),
	}
	odDatFmt := saldaParam.OdDatuma
	if t, err := time.Parse("2006-01-02", saldaParam.OdDatuma); err == nil {
		odDatFmt = t.Format(common.DateLayout)
	}
	doDatFmt := saldaParam.DoDatuma
	if t, err := time.Parse("2006-01-02", saldaParam.DoDatuma); err == nil {
		doDatFmt = t.Format(common.DateLayout)
	}

	reportName := translator.Title("PREGLED SALDA KONTA KLASE 5 I 6")
	switch saldaParam.Klasa {
	case "5":
		reportName = translator.Title("PREGLED SALDA KONTA KLASE 5")
	case "6":
		reportName = translator.Title("PREGLED SALDA KONTA KLASE 6")
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	repParams := domain.ReportParameters{
		ReportName:  reportName,
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		ParameterItems: map[string]domain.ParameterItem{
			"OdSifre":  {Name: translator.Label("Počev od šifre"), Value: saldaParam.OdSifre},
			"DoSifre":  {Name: translator.Label("Zaključno sa šifrom"), Value: saldaParam.DoSifre},
			"OdDatuma": {Name: translator.Label("Počev od datuma"), Value: odDatFmt},
			"DoDatuma": {Name: translator.Label("Zaključno sa datumom"), Value: doDatFmt},
		},
	}

	tbl := common.SetTableBasicData("", "salda-klase56-analitika-stampa", h.service.GetSaldaKlase5i6AnalitikaStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetSaldaKlase5i6AnalitikaStampa(ctx, &tbl, saldaParam); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.SaldaKlase56AnalitikaStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *SaldaHandler) SaldaKlase5i6MTStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	saldaParam := domain.SaldaParam{
		OdKonta:  c.Query("odkonta"),
		DoKonta:  c.Query("dokonta"),
		OdDatuma: c.Query("oddatuma"),
		DoDatuma: c.Query("dodatuma"),
		Klasa:    c.Query("klasa"),
	}

	odDatFmt := saldaParam.OdDatuma
	if t, err := time.Parse("2006-01-02", saldaParam.OdDatuma); err == nil {
		odDatFmt = t.Format(common.DateLayout)
	}
	doDatFmt := saldaParam.DoDatuma
	if t, err := time.Parse("2006-01-02", saldaParam.DoDatuma); err == nil {
		doDatFmt = t.Format(common.DateLayout)
	}

	reportName := translator.Title("PREGLED SALDA KONTA KLASE 5 I 6")
	switch saldaParam.Klasa {
	case "5":
		reportName = translator.Title("PREGLED SALDA KONTA KLASE 5")
	case "6":
		reportName = translator.Title("PREGLED SALDA KONTA KLASE 6")
	}

	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	repParams := domain.ReportParameters{
		ReportName:  reportName,
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		ParameterItems: map[string]domain.ParameterItem{
			"OdKonta":  {Name: translator.Label("Počev od konta"), Value: saldaParam.OdKonta},
			"DoKonta":  {Name: translator.Label("Zaključno sa kontom"), Value: saldaParam.DoKonta},
			"OdDatuma": {Name: translator.Label("Počev od datuma"), Value: odDatFmt},
			"DoDatuma": {Name: translator.Label("Zaključno sa datumom"), Value: doDatFmt},
		},
	}

	tbl := common.SetTableBasicData("", "salda-klase56-mt-stampa", h.service.GetSaldaKlase5i6MTStampaFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetSaldaKlase5i6MTStampa(ctx, &tbl, saldaParam); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.SaldaKlase56MTStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *SaldaHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for salda
	r.GET("/api/salda", h.SaldaMain)
	r.GET("/api/salda/pojedinacnihkonta", h.SaldaPojedinacnihKonta)
	r.GET("/api/salda/pojedinacnihkonta/stampa", h.SaldaPojedinacnihKontaStampa)
	r.GET("/api/salda/grupekonta", h.SaldaGrupeKonta)
	r.GET("/api/salda/partneri", h.SaldaPartneri)
	r.GET("/api/salda/partneridetails/:id", h.SaldaPartneriDetalji)
	r.GET("/api/salda/partneri/izborparametara", h.SaldaPartneriIzborParametara)
	r.GET("/api/salda/partneri/pokontima/stampa", h.SaldaPartneraPoKontimaStampa)
	r.GET("/api/salda/partneriprelomljeno", h.SaldaPartneriPrelomljeno)
	r.GET("/api/salda/partneriprelomljeno/stampa", h.SaldaPartneriPrelomljenoStampa)
	r.GET("/api/salda/klase56analitika", h.SaldaKlase5i6Analitika)
	r.GET("/api/salda/klase56analitika/stampa", h.SaldaKlase5i6AnalitikaStampa)
	r.GET("/api/salda/klase56mt", h.SaldaKlase5i6MT)
	r.GET("/api/salda/klase56mt/stampa", h.SaldaKlase5i6MTStampa)
	r.GET("/api/salda/komercijalisti", h.SaldaKomercijalisti)
	r.GET("/api/salda/realizacijakomercijalisti", h.RealizacijaKomercijalisti)
	r.GET("/api/salda/searchbutton", utils.SearchButtonDialog)
	r.GET("/api/salda/totalvalues", h.TotalValues)
	r.GET("/api/salda/klase56totalvalues", h.Klase5i6TotalValues)
}

func GetSaldaTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "saldapojedinacnihkonta", Label: "Pojedinačnih konta", HXRequestUrl: fmt.Sprintf("%spojedinacnihkonta", saldaURLPrefix), IsActive: true, Name: "saldapojedinacnihkonta"},
			{ID: "saldagrupe", Label: "Grupe konta", HXRequestUrl: fmt.Sprintf("%sgrupekonta", saldaURLPrefix), IsActive: false, Name: "saldagrupe"},
			{ID: "saldapartneri", Label: "Partnera po kontima", HXRequestUrl: fmt.Sprintf("%spartneri", saldaURLPrefix), IsActive: false, Name: "saldapartneri"},
			{ID: "saldapartneriprelomljeno", Label: "Partnera prelomljeno", HXRequestUrl: fmt.Sprintf("%spartneriprelomljeno", saldaURLPrefix), IsActive: false, Name: "saldapartneriprelomljeno"},
			{ID: "saldaklase56analitika", Label: "Klase 5 i 6 sa analitikom", HXRequestUrl: fmt.Sprintf("%sklase56analitika", saldaURLPrefix), IsActive: false, Name: "saldaklase56analitika"},
			{ID: "saldaklase56mt", Label: "Klase 5 i 6 po MT", HXRequestUrl: fmt.Sprintf("%sklase56mt", saldaURLPrefix), IsActive: false, Name: "saldaklase56mt"},
			{ID: "saldakomercijalisti", Label: "Po komercijalistima", HXRequestUrl: fmt.Sprintf("%skomercijalisti", saldaURLPrefix), IsActive: false, Name: "saldakomercijalisti"},
			{ID: "realizacijakomercijalisti", Label: "Realizacija po komercijalistima", HXRequestUrl: fmt.Sprintf("%srealizacijakomercijalisti", saldaURLPrefix), IsActive: false, Name: "realizacijakomercijalisti"},
		},
	}
}

func setActiveSaldaTab(tabs *domain.TabData, tabName string) {
	for i := range tabs.Tabs {
		tabs.Tabs[i].IsActive = tabs.Tabs[i].Name == tabName
	}
}
