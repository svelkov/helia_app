package finansijsko

import (
	"fmt"
	"net/http"

	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/i18n"
	"helia/internal/middleware"
	"helia/internal/service"
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
	saldaURLPrefix                    string = "/api/salda/"
	saldaURLSaldaGrupeKonta           string = "/api/salda/grupekonta"
	saldaURLSaldaPartneri             string = "/api/salda/partneri"
	saldaURLPartneriPrelomljeno       string = "/api/salda/partneriprelomljeno"
	saldaURLKomercijalisti            string = "/api/salda/komercijalisti"
	saldaURLrealizacijakomercijalisti string = "/api/salda/realizacijakomercijalisti"
)

type SaldaHandler struct {
	tabData domain.TabData
	service *service.SaldaResource
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
	hxValsSaldaKomercijalisti = `js:{"pod_datumom": document.getElementById("pod_datumom")?.value,
			"od_komercijaliste": document.getElementById("od_komercijaliste")?.value,
			"do_komercijaliste": document.getElementById("do_komercijaliste")?.value}`
	hxValsSaldaRealizacijakomercijalisti = `js:{"od_datuma": document.getElementById("od_datuma")?.value,
			"do_datuma": document.getElementById("do_datuma")?.value,
			"od_komercijaliste": document.getElementById("od_komercijaliste")?.value,
			"do_komercijaliste": document.getElementById("do_komercijaliste")?.value}`
)

func NewSaldaHandler(service *service.SaldaResource) *SaldaHandler {
	handler := &SaldaHandler{}
	handler.tabData = GetSaldaTabData()
	handler.service = service
	return handler
}

// Helper function to convert SaldaDto to TableRow with month names
func saldaDtoToTableRows(saldaData []domain.SaldaDto) []domain.TableRow {
	monthNames := []string{
		"Početno stanje",
		"Januar",
		"Februar",
		"Mart",
		"April",
		"Maj",
		"Jun",
		"Jul",
		"Avgust",
		"Septembar",
		"Oktobar",
		"Novembar",
		"Decembar",
	}

	rows := make([]domain.TableRow, len(saldaData))
	for i, salda := range saldaData {
		monthName := monthNames[i]
		if monthName == "" {
			monthName = fmt.Sprintf("Mesec %d", salda.Mesec)
		}

		rows[i] = domain.TableRow{
			Fields: []string{
				monthName,
				common.FormatNumberWithSystemLocale(salda.Duguje, 2),
				common.FormatNumberWithSystemLocale(salda.Potrazuje, 2),
				common.FormatNumberWithSystemLocale(salda.Saldo, 2),
				common.FormatNumberWithSystemLocale(salda.SaldoKumul, 2),
			},
		}
	}
	return rows
}
func (h *SaldaHandler) SaldaMain(c *gin.Context) {
	// Create configuration
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/pojedinacnihkonta", "#saldapojedinacnihkonta-table", "innerHTML", "GET", "#konto", hxValsSaldaPojedinacnihKonta, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetPojedKontaTableFields(), "", "", 0, 0, 0, 0)
	h.service.SetDefaultTableData(&tbl)
	setActiveSaldaTab(&h.tabData, "saldapojedinacnihkonta")
	tbl.ShowActions = false
	tbl.FuncClick = "selectRow"                             // naziv js function for Click
	tbl.FuncDblClick = "handleDblClickKontoSelection(this)" // naziv js function for dblClick
	err := tmpl_fin.SaldaMain(h.tabData, tbl, btnPrint, btnObrada, domain.SaldaDto{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *SaldaHandler) SaldaPojedinacnihKonta(c *gin.Context) {
	// Get our custom header
	requestSource := c.Request.Header.Get("X-Request-Source")
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/pojedinacnihkonta", "#saldapojedinacnihkonta-table", "innerHTML", "GET", "", hxValsSaldaPojedinacnihKonta, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampa", "fin_print", "/api/salda/pojedinacnihkonta/print", "#saldapojedinacnihkonta-table", "innerHTML", "GET", "", hxValsSaldaPojedinacnihKonta, true, common.ClassPrintButton, "")
		// if the call come from menu click or tab click then render the page with parameters and empty table
		total := domain.SaldaDto{}
		tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetPojedKontaTableFields(), "", "", 0, 0, 0, 0)
		tbl.ShowActions = false
		h.service.SetDefaultTableData(&tbl)
		setActiveSaldaTab(&h.tabData, "saldapojedinacnihkonta")
		err := tmpl_fin.SaldaPojedinacnihKonta(h.tabData, tbl, btnObrada, btnPrint, total, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	// If it's a POST request, then make obrada
	if requestSource == "btnobrada" {
		// validacija input parametre:
		tipkonta := c.Query("tipkonta")
		if tipkonta == "" {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "Niste izabrali tip konta (sintetika, subsintetika, analitika)")
			return
		}
		fieldParameters := []string{}
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

		response, err := h.service.GetSaldaPojedinacnihKonta(c)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}

		tbl := common.SetTableBasicData("", saldaTableID, h.service.GetPojedKontaTableFields(), "", "/api/salda/pojedinacnihkonta", 0, 0, 0, 0)
		tbl.ShowActions = false
		tbl.ShowPagination = false
		tbl.Pagination.HxVals = hxValsSaldaPojedinacnihKonta
		// Convert SaldaDto to TableRow with month names
		tbl.Rows = saldaDtoToTableRows(response.Data)
		tbl.BtnAdd = domain.Button{IsVisible: false}   // Hide Add button in this view
		tbl.BtnPrint = domain.Button{IsVisible: false} // Hide Print button in this view

		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaGrupeKonta(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	searchInput := domain.InputControl{
		ID:           "search-control",
		Label:        translator.Label("Pretraži"),
		Type:         "search",
		Placeholder:  "Unesite tekst za pretragu",
		HxActionURL:  saldaURLSaldaGrupeKonta,
		HxTarget:     fmt.Sprintf("#%s", saldaGrupeKontaTableID),
		HxSwap:       "innerHTML",
		HxTrigger:    "keyup changed delay:500ms",
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
		HxVals:       hxValsSaldaGrupeKonta,
	}
	tbl := common.SetTableBasicData(saldaContentTitle, saldaGrupeKontaTableID, h.service.GetGrupeKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	tbl.ContentTitle = ""
	tbl.URLPrefix = saldaURLSaldaGrupeKonta
	tbl.URLGetAll = saldaURLSaldaGrupeKonta
	tbl.BtnAdd.IsVisible = false
	tbl.BtnPrint.IsVisible = false
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/grupekonta", "#saldagrupekonta-table", "innerHTML", "GET", "", hxValsSaldaGrupeKonta, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", "/api/salda/grupekonta/print", "#saldagrupekonta-table", "innerHTML", "GET", "", hxValsSaldaGrupeKonta, true, common.ClassPrintButton, "")

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
		page, pageSize := common.GetPageAndPageSizeFromRequest(c)
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

	tblPartneri := common.SetTableBasicData("", saldaPartneriTableID, h.service.GetSaldaPartneriTableFields(), "", "", 0, 0, 0, 0)
	translator := i18n.GetInstance()
	searchInput := domain.InputControl{
		ID:           "search-control",
		Label:        translator.Label("Pretraži"),
		Type:         "search",
		Placeholder:  "Unesite tekst za pretragu",
		HxActionURL:  saldaURLSaldaPartneri,
		HxTarget:     fmt.Sprintf("#%s", saldaPartneriTableID),
		HxSwap:       "innerHTML",
		HxTrigger:    "keyup changed delay:500ms",
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c)

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
	tblPartneri.URLPrefix = saldaURLSaldaPartneri
	tblPartneri.URLGetAll = saldaURLSaldaPartneri
	tblPartneri.BtnAdd.IsVisible = false
	tblPartneri.BtnPrint.IsVisible = true
	tblPartneri.ContentTitle = ""
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
	if requestSource == "btnpage" || requestSource == "searchinput" {
		utils.RenderContent(c, tblPartneri)
	}
}

func (h *SaldaHandler) SaldaPartneriDetalji(c *gin.Context) {
	strPartnerID := c.Query("idpartneri")
	idPartneri := common.StringToInt64(strPartnerID)
	tblSalda := common.SetTableBasicData("", "saldapartneri-salda-table", h.service.GetSaldaPartneriHeaderTableFields(), "", "", 0, 0, 0, 0)
	tblSalda.ShowActions = false
	tblSalda.ShowPagination = false

	tblDetalji := common.SetTableBasicData("", "saldapartneri-detalji-table", h.service.GetSaldaPartneriDetailTableFields(), "", "", 0, 0, 0, 0)
	tblDetalji.ShowActions = false
	tblDetalji.ShowPagination = false

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
	searchInput := domain.InputControl{
		ID:           "search-control",
		Label:        translator.Placeholder("input_search"),
		Type:         "search",
		Placeholder:  "Unesite tekst za pretragu",
		HxActionURL:  saldaURLPartneriPrelomljeno,
		HxTarget:     fmt.Sprintf("#%s", saldaTablePrelomljenoID),
		HxSwap:       "innerHTML",
		HxTrigger:    "keyup changed delay:500ms",
		HxVals:       hxValsSaldaPartneriPrelomljeno,
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTablePrelomljenoID, h.service.GetSaldaPartneriPrelomljenoTableFields(), "", "", 0, 0, 0, 0)
	tbl.URLPrefix = saldaURLPartneriPrelomljeno
	tbl.URLGetAll = saldaURLPartneriPrelomljeno
	tbl.BtnAdd.IsVisible = false
	tbl.BtnPrint.IsVisible = false

	tbl.ContentTitle = ""
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/partneriprelomljeno", "#saldatable-prelomljeno", "innerHTML", "GET", "", hxValsSaldaPartneriPrelomljeno, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", "/api/salda/partneriprelomljeno/print", "#saldatable-prelomljeno", "innerHTML", "GET", "", hxValsSaldaPartneriPrelomljeno, true, common.ClassPrintButton, "")

		tbl.ShowActions = false
		setActiveSaldaTab(&h.tabData, "saldapartneriprelomljeno")
		err := tmpl_fin.SaldaPartneriPrelomljeno(h.tabData, tbl, btnObrada, btnPrint, searchInput, translator).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c)
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
		tbl.ShowActions = false
		tbl.ShowPagination = true

		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaKlase5i6Analitika(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/salda/saldaklase56analitika", "#saldatable", "innerHTML", "GET", "", hxValsSaldaPojedinacnihKonta, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Stampaj", "fin_print", "/api/salda/saldaklase56analitika/print", "#saldatable", "innerHTML", "GET", "", hxValsSaldaPojedinacnihKonta, true, common.ClassPrintButton, "")
		total := domain.SaldaDto{}
		tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetPojedKontaTableFields(), "", "", 0, 0, 0, 0)
		tbl.ShowActions = false
		setActiveSaldaTab(&h.tabData, "saldaklase56")
		err := tmpl_fin.SaldaPojedinacnihKonta(h.tabData, tbl, btnObrada, btnPrint, total, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		fieldParameters := []string{"konto"}
		fieldsError := h.service.CheckSaldaParameters(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c)
		response, err := h.service.GetSaldaPojedinacnihKonta(c)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		totalRecords := response.TotalRecords
		totalPages := (totalRecords + pageSize - 1) / pageSize
		response, err = h.service.GetSaldaPojedinacnihKonta(c)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		tbl := common.SetTableBasicData("", saldaTableID, h.service.GetPojedKontaTableFields(), "", "/api/salda/saldaklase56analitika", pageSize, page, totalPages, totalRecords)
		tbl.ShowActions = false
		tbl.ShowPagination = true
		tbl.Pagination.HxVals = hxValsSaldaPojedinacnihKonta
		tblRows, err := common.SetTableRows(&tbl, response.Data, h.service.GetPojedKontaTableFields(), "idfpro", "", h.service.GetFieldCache())
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgFailedToSetTableRow)
			return
		}
		tbl.Rows = tblRows.Rows
		tbl.BtnAdd = domain.Button{IsVisible: false}
		tbl.BtnPrint = domain.Button{IsVisible: false}
		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) SaldaKlase5i6Mi(c *gin.Context) {

}
func (h *SaldaHandler) SaldaKomercijalisti(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	searchInput := domain.InputControl{
		ID:           "search-control",
		Label:        translator.Placeholder("input_search"),
		Type:         "search",
		Placeholder:  "Unesite tekst za pretragu",
		HxActionURL:  saldaURLKomercijalisti,
		HxTarget:     fmt.Sprintf("#%s", saldaTableKomercijalistiTableID),
		HxSwap:       "innerHTML",
		HxTrigger:    "keyup changed delay:500ms",
		HxVals:       hxValsSaldaKomercijalisti,
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetKomercijalistiTableFields(), "", "", 0, 0, 0, 0)
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLKomercijalisti, "#saldatable-komercijalisti", "innerHTML", "GET", "", hxValsSaldaKomercijalisti, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampaj", "fin_print", saldaURLKomercijalisti+"/print", "#saldatable-komercijalisti", "innerHTML", "GET", "", hxValsSaldaKomercijalisti, true, common.ClassPrintButton, "")

		tbl.ShowActions = false
		tbl.BtnAdd.IsVisible = false
		tbl.BtnPrint.IsVisible = false
		tbl.ContentTitle = ""
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
		fieldsError := h.service.CheckSaldaParameters(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c)
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
		tbl.ShowActions = false
		tbl.ShowPagination = true

		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) RealizacijaKomercijalisti(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()
	searchInput := domain.InputControl{
		ID:           "search-control",
		Label:        translator.Placeholder("input_search"),
		Type:         "search",
		Placeholder:  "Unesite tekst za pretragu",
		HxActionURL:  saldaURLrealizacijakomercijalisti,
		HxTarget:     fmt.Sprintf("#%s", saldaTablerealizacijakomTableID),
		HxSwap:       "innerHTML",
		HxTrigger:    "keyup changed delay:500ms",
		HxVals:       hxValsSaldaRealizacijakomercijalisti,
		Autocomplete: "off",
		Class:        common.ClassSearchInput,
	}
	tbl := common.SetTableBasicData(saldaContentTitle, saldaTableID, h.service.GetRealizacijaKomercijalistiTableFields(), "", "", 0, 0, 0, 0)
	if requestSource == "menu" || requestSource == "tab" {
		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", saldaURLrealizacijakomercijalisti, "#realizacija-komercijalisti-table", "innerHTML", "GET", "", hxValsSaldaRealizacijakomercijalisti, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampaj", "fin_print", saldaURLrealizacijakomercijalisti+"/print", "#realizacija-komercijalisti-table", "innerHTML", "GET", "", hxValsSaldaRealizacijakomercijalisti, true, common.ClassPrintButton, "")

		tbl.ShowActions = false
		tbl.BtnAdd.IsVisible = false
		tbl.BtnPrint.IsVisible = false
		tbl.ContentTitle = ""
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
		fieldsError := h.service.CheckSaldaParameters(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}
		page, pageSize := common.GetPageAndPageSizeFromRequest(c)
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
		tbl.ShowActions = false
		tbl.ShowPagination = true

		utils.RenderContent(c, tbl)
	}
}

func (h *SaldaHandler) TotalValues(c *gin.Context) {
	fieldParameters := []string{"konto"}
	fieldsError := h.service.CheckSaldaParameters(c, fieldParameters)
	if len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
		return
	}

	totalValues, err := h.service.GetSaldaTotalValues(c)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
		return
	}
	tmpl_fin.SaldaTotalValues(totalValues, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
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
	r.GET("/api/salda/klase56mi", h.SaldaKlase5i6Mi)
	r.GET("/api/salda/komercijalisti", h.SaldaKomercijalisti)
	r.GET("/api/salda/realizacijakomercijalisti", h.RealizacijaKomercijalisti)
	r.GET("/api/salda/searchbutton", utils.SearchButtonDialog)
	r.GET("/api/salda/totalvalues", h.TotalValues)
}

func GetSaldaTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "saldapojedinacnihkonta", Label: "Pojedinačnih konta", HXRequestUrl: fmt.Sprintf("%spojedinacnihkonta", saldaURLPrefix), IsActive: true, Name: "saldapojedinacnihkonta"},
			{ID: "saldagrupe", Label: "Grupe konta", HXRequestUrl: fmt.Sprintf("%sgrupekonta", saldaURLPrefix), IsActive: false, Name: "saldagrupe"},
			{ID: "saldapartneri", Label: "Partnera po kontima", HXRequestUrl: fmt.Sprintf("%spartneri", saldaURLPrefix), IsActive: false, Name: "saldapartneri"},
			{ID: "saldapartneriprelomljeno", Label: "Partnera prelomljeno", HXRequestUrl: fmt.Sprintf("%spartneriprelomljeno", saldaURLPrefix), IsActive: false, Name: "saldapartneriprelomljeno"},
			{ID: "saldaklase56analitika", Label: "Klase 5 i 6 sa analitikom", HXRequestUrl: fmt.Sprintf("%sklase56analitika", saldaURLPrefix), IsActive: false, Name: "saldaklase56analitika"},
			{ID: "saldaklase56mi", Label: "Klase 5 i 6 po MI", HXRequestUrl: fmt.Sprintf("%sklase56mi", saldaURLPrefix), IsActive: false, Name: "saldaklase56mi"},
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
