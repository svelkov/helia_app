package finansijsko

import (
	"fmt"
	"helia/config"
	"log"
	"net/http"

	tmpl_fin "helia/frontend/templates/finansijsko"
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
	saldaURLPartneriPrelomljeno       string = "/api/salda/partneriprelomljeno"
	saldaURLKomercijalisti            string = "/api/salda/komercijalisti"
	saldaURLrealizacijakomercijalisti string = "/api/salda/realizacijakomercijalisti"
	saldaURLKlase5i6Analitika         string = "/api/salda/klase56analitika"
	saldaURLKlase5i6MT                string = "/api/salda/klase56mt"
	saldaURLtotals                    string = "/api/salda/totalvalues"
	saldaURLKlase5i6Totals            string = "/api/salda/klase56totalvalues"
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
            "od_konta": document.getElementById("od_konta")?.value,
			"do_konta": document.getElementById("do_konta")?.value,
			"od_sifre": document.getElementById("od_sifre")?.value,
			"do_sifre": document.getElementById("do_sifre")?.value,
			"chk_salda_valute": document.querySelector('input[name="chk_salda_valute"]:checked')?.value,
			"cbx_tipizvestaja": document.getElementById("cbx_tipizvestaja")?.value,
			"cbx_klasa": document.getElementById("cbx_klasa")?.value,
			"cbx_odmeseca": document.getElementById("cbx_odmeseca")?.value,
			"cbx_domeseca": document.getElementById("cbx_domeseca")?.value,
			"chk_saldo_razl_nula": document.querySelector('input[name="chk_saldo_razl_nula"]:checked')?.value,
			"chk_saldo_vece_nula": document.querySelector('input[name="chk_saldo_vece_nula"]:checked')?.value,
			"chk_saldo_manje_nula": document.querySelector('input[name="chk_saldo_manje_nula"]:checked')?.value,
			"chk_saldo_nula": document.querySelector('input[name="chk_saldo_nula"]:checked')?.value,
			"od_salda": document.getElementById("od_salda")?.value,
			"do_salda": document.getElementById("do_salda")?.value,
        }`

	hxValsSaldaPartneriPrelomljeno = `js:{"sifra_od": document.getElementById("sifra_od")?.value,
			"sifra_do": document.getElementById("sifra_do")?.value}`

	hxValsSaldaKlase5i6Analitika = `js:{
			"odsifre": document.getElementById("odsifre")?.value,
			"dosifre": document.getElementById("dosifre")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
			"klasa": document.getElementById("klasa")?.value	
		}`
	hxValsSaldaKlase5i6MT = `js:{
			"odkonta": document.getElementById("odkonta")?.value,
			"dokonta": document.getElementById("dokonta")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
			"klasa": document.getElementById("klasa")?.value	
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
	btnPrint := common.SetButton("stampa", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

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
	btnPrint := common.SetButton("stampa", "Štampa", "fin_print", saldaURLPojedinacnihKonta+"/print", "#saldapojedinacnihkonta-table", "innerHTML", "GET", "", hxValsSaldaPojedinacnihKonta, true, common.ClassPrintButton, "")
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
		fieldsError := h.service.CheckSaldaParameters(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsSaldaPojedinacnihKonta
		err := h.service.GetSaldaPojedinacnihKonta(c, &tbl, false, pageSize, page)
		tbl.ShowPagination = false
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaGrupeKonta(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	tbl := common.SetTableBasicData(saldaContentTitle, saldaGrupeKontaTableID, h.service.GetGrupeKontaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLSaldaGrupeKonta, "#saldagrupekonta-table", "innerHTML", "GET", "", hxValsSaldaGrupeKonta, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", saldaURLSaldaGrupeKonta+"/print", "#saldagrupekonta-table", "innerHTML", "GET", "", hxValsSaldaGrupeKonta, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput(translator, saldaURLSaldaGrupeKonta, fmt.Sprintf("#%s", saldaGrupeKontaTableID), hxValsSaldaGrupeKonta)
	common.SetTableConfig(&tbl, "", saldaURLSaldaGrupeKonta, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		setActiveSaldaTab(&h.tabData, "saldagrupe")
		err := tmpl_fin.SaldaGrupeKonta(h.tabData, tbl, btnObrada, btnPrint, common.MonthComboItems, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"od_konta", "do_konta", "od_sifre", "do_sifre", "chk_salda_valute",
			"cbx_tipizvestaja", "cbx_klasa", "cbx_odmeseca", "cbx_domeseca", "chk_saldo_razl_nula",
			"chk_saldo_vece_nula", "chk_saldo_manje_nula", "chck_saldo_nula", "od_salda", "do_salda"}
		fieldsError := h.service.CheckSaldaGrupeParameters(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetSaldaGrupeKonta(c, &tbl, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSaldaGrupeKonta(c, &tbl, false, page, pageSize)
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
	searchInput := common.CreateSearchInput(translator, saldaURLSaldaPartneri, fmt.Sprintf("#%s", saldaPartneriTableID), "")
	common.SetTableConfig(&tblPartneri, "", saldaURLSaldaPartneri, false, false, false)

	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	err := h.service.GetSaldaPartneriList(c, &tblPartneri, true, currentPage, pageSize)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	err = h.service.GetSaldaPartneriList(c, &tblPartneri, false, currentPage, pageSize)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	tblPartneri.BtnAdd.IsVisible = false
	tblPartneri.BtnPrint.IsVisible = true
	tblPartneri.DetailTarget = "#saldapartneri-detalji"
	tblPartneri.DetailURL = "/api/salda/partneridetails?idpartneri="
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
	strPartnerID := c.Query("idpartneri")
	idPartneri := common.StringToInt64(strPartnerID)
	tblSalda := common.SetTableBasicData("", "saldapartneri-salda-table", h.service.GetSaldaPartneriHeaderTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblSalda, "", saldaURLSaldaPartneri, false, false, false)

	tblDetalji := common.SetTableBasicData("", "saldapartneri-detalji-table", h.service.GetSaldaPartneriDetailTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tblDetalji, "", saldaURLSaldaPartneri, false, false, false)
	err := h.service.ProcessSaldaPartneriDetails(c, idPartneri, &tblSalda, &tblDetalji)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	tmpl_fin.SaldaPartneriDetalji(tblSalda, tblDetalji, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *SaldaHandler) SaldaPartneriPrelomljeno(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTablePrelomljenoID, h.service.GetSaldaPartneriPrelomljenoTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	searchInput := common.CreateSearchInput(translator, saldaURLPartneriPrelomljeno, fmt.Sprintf("#%s", saldaTablePrelomljenoID), "")
	common.SetTableConfig(&tbl, "", saldaURLPartneriPrelomljeno, false, false, false)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/partneriprelomljeno", "#saldatable-prelomljeno", "innerHTML", "GET", "", hxValsSaldaPartneriPrelomljeno, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", "/api/salda/partneriprelomljeno/print", "#saldatable-prelomljeno", "innerHTML", "GET", "", hxValsSaldaPartneriPrelomljeno, true, common.ClassPrintButton, "")
	setActiveSaldaTab(&h.tabData, "saldapartneriprelomljeno")

	if requestSource == "menu" || requestSource == "tab" {
		err := tmpl_fin.SaldaPartneriPrelomljeno(h.tabData, tbl, btnObrada, btnPrint, searchInput, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetSaldaPartneriPrelomljeno(c, &tbl, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSaldaPartneriPrelomljeno(c, &tbl, false, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaKlase5i6Analitika(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	total := domain.SaldaDto{}
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "user session not found")
		return
	}
	gnGod := userSession.SelectedGod
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLKlase5i6Analitika, fmt.Sprintf("#%s", saldaTableKlase5i6AnalitikaID), "innerHTML", "GET", "", hxValsSaldaKlase5i6Analitika, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", saldaURLKlase5i6Analitika+"/print", fmt.Sprintf("#%s", saldaTableKlase5i6AnalitikaID), "innerHTML", "GET", "", hxValsSaldaKlase5i6Analitika, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput(i18n.GetInstance(), saldaURLKlase5i6Analitika, fmt.Sprintf("#%s", saldaTableKlase5i6AnalitikaID), "")

	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableKlase5i6AnalitikaID, h.service.GetSaldaKlase5i6AnalitikaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.Pagination.HxVals = hxValsSaldaKlase5i6Analitika
	common.SetTableConfig(&tbl, saldaContentTitle, "", false, false, false)

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
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetSaldaKlase5i6Analitika(c, &tbl, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetSaldaKlase5i6Analitika(c, &tbl, false, page, pageSize)
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
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, "user session not found")
		return
	}
	gnGod := userSession.SelectedGod
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLKlase5i6MT, fmt.Sprintf("#%s", saldaTableKlase5i6MTID), "innerHTML", "GET", "", hxValsSaldaKlase5i6MT, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", saldaURLKlase5i6MT+"/print", fmt.Sprintf("#%s", saldaTableKlase5i6MTID), "innerHTML", "GET", "", hxValsSaldaKlase5i6MT, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput(i18n.GetInstance(), saldaURLKlase5i6MT, fmt.Sprintf("#%s", saldaTableKlase5i6MTID), "")

	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableKlase5i6MTID, h.service.GetSaldaKlase5i6MTTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tbl.Pagination.HxVals = hxValsSaldaKlase5i6MT
	common.SetTableConfig(&tbl, saldaContentTitle, "", false, false, false)

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
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.SaldaKlase5i6MT(c, &tbl, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.SaldaKlase5i6MT(c, &tbl, false, page, pageSize)
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
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetKomercijalistiTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	searchInput := common.CreateSearchInput(translator, saldaURLKomercijalisti, fmt.Sprintf("#%s", saldaTableKomercijalistiTableID), "")
	common.SetTableConfig(&tbl, "", saldaURLKomercijalisti, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLKomercijalisti, "#saldatable-komercijalisti", "innerHTML", "GET", "", hxValsSaldaKomercijalisti, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampaj", "fin_print", saldaURLKomercijalisti+"/print", "#saldatable-komercijalisti", "innerHTML", "GET", "", hxValsSaldaKomercijalisti, true, common.ClassPrintButton, "")
		setActiveSaldaTab(&h.tabData, "saldakomercijalisti")
		err := tmpl_fin.SaldaPoKomercijalistima(h.tabData, tbl, btnObrada, btnPrint, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.SaldaPoKomercijalistima(c, &tbl, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.SaldaPoKomercijalistima(c, &tbl, false, page, pageSize)
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
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetRealizacijaKomercijalistiTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	searchInput := common.CreateSearchInput(translator, saldaURLrealizacijakomercijalisti, fmt.Sprintf("#%s", saldaTablePrelomljenoID), "")
	common.SetTableConfig(&tbl, "", saldaURLrealizacijakomercijalisti, false, false, false)

	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLrealizacijakomercijalisti, "#realizacija-komercijalisti-table", "innerHTML", "GET", "", hxValsSaldaRealizacijakomercijalisti, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampaj", "fin_print", saldaURLrealizacijakomercijalisti+"/print", "#realizacija-komercijalisti-table", "innerHTML", "GET", "", hxValsSaldaRealizacijakomercijalisti, true, common.ClassPrintButton, "")

		setActiveSaldaTab(&h.tabData, "realizacijakomercijalisti")
		err := tmpl_fin.RealizacijaKomercijalisti(h.tabData, tbl, btnObrada, btnPrint, searchInput, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.RealizacijaKomercijalisti(c, &tbl, true, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.RealizacijaKomercijalisti(c, &tbl, false, page, pageSize)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) Klase5i6TotalValues(c *gin.Context) {
	totalValues, err := h.service.GetSaldaKlase5i6TotalValues(c)
	if err != nil {
		log.Printf("Error getting total values for klase 5 i 6: %v", err)
	}
	tmpl_fin.SaldaTotalValues(totalValues, saldaURLKlase5i6Totals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *SaldaHandler) TotalValues(c *gin.Context) {
	totalValues, err := h.service.GetSaldaTotalValues(c)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	tmpl_fin.SaldaTotalValues(totalValues, saldaURLtotals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *SaldaHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for salda
	r.GET("/api/salda", h.SaldaMain)
	r.GET("/api/salda/pojedinacnihkonta", h.SaldaPojedinacnihKonta)
	r.GET("/api/salda/grupekonta", h.SaldaGrupeKonta)
	r.GET("/api/salda/partneri", h.SaldaPartneri)
	r.GET("/api/salda/partneridetails", h.SaldaPartneriDetalji)
	r.GET("/api/salda/partneriprelomljeno", h.SaldaPartneriPrelomljeno)
	r.GET("/api/salda/klase56analitika", h.SaldaKlase5i6Analitika)
	r.GET("/api/salda/klase56mt", h.SaldaKlase5i6MT)
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
