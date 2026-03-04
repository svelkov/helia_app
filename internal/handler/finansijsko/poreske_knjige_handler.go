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
	poreskeKnjigeContentTitle  string = "PORESKE KNJIGE"
	poreskeKnjigeTableID       string = "poreskeknjigetable"
	poreskeKnjigeURLPrefix     string = "/api/poreskeknjige/"
	poreskeKnjigeURLGetAll     string = "/api/poreskeknjige/all"
	poreskeKnjigeURLIzdatih    string = "/api/poreskeknjige/izdatih"
	poreskeKnjigeURLPrimljenih string = "/api/poreskeknjige/primljenih"
	poreskeKnjigeURLPrijava    string = "/api/poreskeknjige/prijava"
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
	service *finservice.PoreskeKnjigeResource
	cfg     config.Config
}

func NewPoreskeKnjigeHandler(service *finservice.PoreskeKnjigeResource, cfg config.Config) *PoreskeKnjigeHandler {
	handler := &PoreskeKnjigeHandler{
		cfg: cfg,
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

	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", poreskeKnjigeURLIzdatih, "#"+poreskeKnjigeTableID, "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("stampa-btn", "Štampa", "fin_print", poreskeKnjigeURLPrimljenih+"/print", "", "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), poreskeKnjigeURLIzdatih, fmt.Sprintf("#%s", poreskeKnjigeTableID), hxValsKirKpr)

	tbl := common.SetTableBasicData(poreskeKnjigeContentTitle, poreskeKnjigeTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, poreskeKnjigeContentTitle, "", false, false, false)
	tbl.Pagination.HxVals = hxValsKirKpr
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxActionURL = poreskeKnjigeURLIzdatih + "/unos"
	tbl.BtnAdd.HxTarget = "#dialog-content"
	common.SetTableConfig(&tbl, "KNJIGA IZDATIH RAČUNA", poreskeKnjigeURLIzdatih, true, true, false)
	// Get knjiga values
	knjigaValues := []domain.ComboItem{}
	err := h.service.GetTipoveKnjigaValues(c, &knjigaValues, "I")
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

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", poreskeKnjigeURLIzdatih, "#"+poreskeKnjigeTableID, "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa-btn", "Štampa", "fin_print", poreskeKnjigeURLIzdatih+"/print", "", "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput("search-input", translator, poreskeKnjigeURLIzdatih, fmt.Sprintf("#%s", poreskeKnjigeTableID), hxValsKirKpr)

		tbl := common.SetTableBasicData(poreskeKnjigeContentTitle, poreskeKnjigeTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsKirKpr
		tbl.BtnAdd.HxRequestType = "GET"
		tbl.BtnAdd.HxActionURL = poreskeKnjigeURLIzdatih + "/unos"
		tbl.BtnAdd.HxTarget = "#dialog-content"

		common.SetTableConfig(&tbl, "KNJIGA IZDATIH RAČUNA", poreskeKnjigeURLIzdatih, true, true, false)

		knjigaValues := []domain.ComboItem{}
		// Get knjiga values
		err := h.service.GetTipoveKnjigaValues(c, &knjigaValues, "I")
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

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", poreskeKnjigeTableID, h.service.GetKirTableFields(), "", poreskeKnjigeURLIzdatih, 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsKirKpr
		common.SetTableConfig(&tbl, "", poreskeKnjigeURLIzdatih, true, true, false)
		err := h.service.GetKnjigaIzdatihRacuna(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetKnjigaIzdatihRacuna(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}
func (h *PoreskeKnjigeHandler) KnjigaIzdatihRacunaUnos(c *gin.Context) {
	translator := i18n.GetInstance()
	btnSave := common.SetButton("save-btn", "Sačuvaj", "save", "", "#dialog-content", "innerHTML", "POST", "", "", true, common.ClassSaveButton, "handleKIRUnosResponse")
	btnCancel := common.SetButton("cancel-btn", "Odustani", "odustani", "", "#dialog-content", "innerHTML", "GET", "", "", true, common.ClassButton, "closeDialog()")
	tipdokValues := []domain.ComboItem{}
	knjigaValues := []domain.ComboItem{}
	h.service.GetTipoveKnjigaValues(c, &knjigaValues, "I")

	err := tmpl_fin.DialogKirUnos(knjigaValues, tipdokValues, btnSave, btnCancel, translator).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
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

		btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", poreskeKnjigeURLPrimljenih, "#"+poreskeKnjigeTableID, "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa-btn", "Štampa", "fin_print", poreskeKnjigeURLPrimljenih+"/print", "", "innerHTML", "GET", "", hxValsKirKpr, true, common.ClassPrintButton, "")
		searchInput := common.CreateSearchInput("search-input", translator, poreskeKnjigeURLPrimljenih, fmt.Sprintf("#%s", poreskeKnjigeTableID), hxValsKirKpr)

		tbl := common.SetTableBasicData(poreskeKnjigeContentTitle, poreskeKnjigeTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsKirKpr
		common.SetTableConfig(&tbl, "KNJIGA PRIMLJENIH RAČUNA", poreskeKnjigeURLPrimljenih, true, true, false)

		knjigaValues := []domain.ComboItem{}
		// Get knjiga values
		err := h.service.GetTipoveKnjigaValues(c, &knjigaValues, "U")
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

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", poreskeKnjigeTableID, h.service.GetKprTableFields(), "", poreskeKnjigeURLPrimljenih, 0, 0, 0, 0, h.cfg)
		tbl.Pagination.HxVals = hxValsKirKpr
		common.SetTableConfig(&tbl, "", poreskeKnjigeURLPrimljenih, true, true, false)
		err := h.service.GetKnjigaPrimljenihRacuna(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetKnjigaPrimljenihRacuna(c, &tbl, false, pageSize, page)
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
		btnObrada.HxOn = "handlePoreskaPrijavaResponse(evt)"
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
		err := h.service.GetPoreskaPrijava(c, poreskaPrijavaData)
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

// RegisterRoutes registers the routes for the PoreskeKnjige handler
func (h *PoreskeKnjigeHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/poreskeknjige", h.PoreskeKnjigeMain)
	r.GET("api/poreskeknjige/izdatih", h.KnjigaIzdatihRacuna)
	r.GET("api/poreskeknjige/izdatih/unos", h.KnjigaIzdatihRacunaUnos)
	r.GET("api/poreskeknjige/primljenih", h.KnjigaPrimljenihRacuna)
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
