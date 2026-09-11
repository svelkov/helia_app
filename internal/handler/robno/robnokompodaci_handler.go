package robno

import (
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
	robnoKompodaciURLPrefix = "/api/robno-kompodaci"

	robnoKompodaciKupciArtTitle      = "Pregled realizacije po kupcima i artiklima"
	robnoKompodaciURLKupciArt        = robnoKompodaciURLPrefix + "/realizacija-po-kupcima-artiklima"
	robnoKompodaciURLKupciArtStampa  = robnoKompodaciURLPrefix + "/realizacija-po-kupcima-artiklima/stampa"
	robnoKompodaciKupciArtTableID    = "robno-kompodaci-kupci-artikli-table"
	robnoKompodaciKupciGrTitle       = "Pregled realizacije po kupcima i grupama"
	robnoKompodaciURLKupciGr         = robnoKompodaciURLPrefix + "/realizacija-po-kupcima-grupama"
	robnoKompodaciKupciGrTableID     = "robno-kompodaci-kupci-grupe-table"
	robnoKompodaciArtTitle           = "Pregled realizacije po artiklima"
	robnoKompodaciURLArt             = robnoKompodaciURLPrefix + "/realizacija-po-artiklima"
	robnoKompodaciURLArtStampa       = robnoKompodaciURLPrefix + "/realizacija-po-artiklima/stampa"
	robnoKompodaciArtTableID         = "robno-kompodaci-artikli-table"
	robnoKompodaciUcesceArtTitle     = "Pregled ucesca artikla u ukupnom prometu"
	robnoKompodaciURLUcesceArt       = robnoKompodaciURLPrefix + "/ucesce-artikla"
	robnoKompodaciURLUcesceArtStampa = robnoKompodaciURLPrefix + "/ucesce-artikla/stampa"
	robnoKompodaciUcesceArtTableID   = "robno-kompodaci-ucesce-artikla-table"
	robnoKompodaciUcesceGrTitle      = "Pregled ucesca grupe artikala u ukupnom prometu"
	robnoKompodaciURLUcesceGr        = robnoKompodaciURLPrefix + "/ucesce-grupe-artikala"
	robnoKompodaciURLUcesceGrStampa  = robnoKompodaciURLPrefix + "/ucesce-grupe-artikala/stampa"
	robnoKompodaciUcesceGrTableID    = "robno-kompodaci-ucesce-grupe-table"

	hxValsRobnoKompodaciRealizKupciArt = `js:{
            "tip_izvestaja": document.querySelector('input[name="tip_izvestaja"]:checked')?.value,
        	"trziste": document.querySelector('input[name="trziste"]:checked')?.value,
      		"odgrupe": document.getElementById("odgrupe")?.value,
			"dogrupe": document.getElementById("dogrupe")?.value,
			"odsifreartikla": document.getElementById("odsifreartikla")?.value,
			"dosifreartikla": document.getElementById("dosifreartikla")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
            "dodatuma": document.getElementById("dodatuma")?.value,
		}`
	hxValsRobnoKompodaciArt = `js:{
          	"trziste": document.querySelector('input[name="trziste"]:checked')?.value,
      		"odgrupe": document.getElementById("odgrupe")?.value,
			"dogrupe": document.getElementById("dogrupe")?.value,
			"odsifreartikla": document.getElementById("odsifreartikla")?.value,
			"dosifreartikla": document.getElementById("dosifreartikla")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
            "dodatuma": document.getElementById("dodatuma")?.value,
		}`
	hxValsRobnoKompodaciUcesceArt = `js:{
            "odsifreartikla": document.getElementById("odsifreartikla")?.value,
			"dosifreartikla": document.getElementById("dosifreartikla")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
            "dodatuma": document.getElementById("dodatuma")?.value,
		}`
	hxValsRobnoKompodaciUcesceGr = `js:{
            "odgrupe": document.getElementById("odgrupe")?.value,
			"dogrupe": document.getElementById("dogrupe")?.value,
			"oddatuma": document.getElementById("oddatuma")?.value,
            "dodatuma": document.getElementById("dodatuma")?.value,
		}`
)

type RobnoKompodaciHandler struct {
	service robnosvc.RobnoKompodaciService
	cfg     config.Config
	tabs    *domain.TabData
}

func NewRobnoKompodaciHandler(service robnosvc.RobnoKompodaciService, cfg config.Config) *RobnoKompodaciHandler {
	return &RobnoKompodaciHandler{service: service, cfg: cfg, tabs: robnoKompodaciTabs()}
}

func (h *RobnoKompodaciHandler) RobnoKompodaciMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	translator := i18n.GetInstance()
	common.SetActiveTab(h.tabs, 0)
	grpValues, _ := h.service.GetRobneGrupeComboValues(c.Request.Context())
	currentPage, pageSize, totalPages := common.GetPaginationData(c, 0, h.cfg)
	tbl := common.SetTableBasicData(robnoKompodaciKupciArtTitle, robnoKompodaciKupciArtTableID, h.service.GetPrikazKarticeKupcaDobavljacaTableFields(), robnoKompodaciURLKupciArt, robnoKompodaciURLKupciArt, pageSize, currentPage, totalPages, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKompodaciKupciArtTableID, robnoKompodaciURLKupciArt, false, false, false)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", robnoKompodaciURLKupciArt, "#"+robnoKompodaciURLKupciArt, "innerHTML", "GET", "", hxValsRobnoKompodaciRealizKupciArt, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetButton("print-btn", "Štampaj", "stampa", robnoKompodaciURLKupciArtStampa, "", "", "GET", "", hxValsRobnoKompodaciRealizKupciArt, true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", translator, robnoKompodaciURLKupciArt, fmt.Sprintf("#%s", robnoKompodaciURLKupciArt), hxValsRobnoKompodaciRealizKupciArt)

	tmpl_robno.RobnoKompodaciMain(*h.tabs, tbl, grpValues, btnObrada, btnPrint, searchInput, session.SelectedGod, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)

}

func (h *RobnoKompodaciHandler) PregledRealizacijePoKupcimaArtiklima(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	common.SetActiveTab(h.tabs, 0)
	tbl := common.SetTableBasicData(robnoKompodaciKupciArtTitle, robnoKompodaciKupciArtTableID, h.service.GetPregledRealizacijePoKupcimaArtiklimaTableFields(), robnoKompodaciURLKupciArt, "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKompodaciKupciArtTableID, robnoKompodaciURLKupciArt, false, false, false)
	if common.IsDataRequest(c) {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPregledRealizacijePoKupcimaArtiklima(ctx, &tbl, true, pageSize, page, h.params(c), common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Greska prilikom dobavljanja resultata iz beckend-a, greska: %s", err.Error()))
			return
		}
		if err := h.service.GetPregledRealizacijePoKupcimaArtiklima(ctx, &tbl, false, pageSize, page, h.params(c), common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Greska prilikom dobavljanja resultata iz beckend-a, greska: %s", err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	grpValues, _ := h.service.GetRobneGrupeComboValues(ctx)
	btnObrada := common.SetButton("obrada-btn", "Obradi", "fin_obrada", robnoKompodaciURLKupciArt, "#"+robnoKompodaciKupciArtTableID, "innerHTML", "GET", "", hxValsRobnoKompodaciRealizKupciArt, true, common.ClassSaveButton, "")
	btnPrint := common.SetPrintButton("print-btn", "Štampa", "stampa", robnoKompodaciURLArtStampa, "GET", true, common.ClassPrintButton, "trziste,tip_izvestaja,odgrupe,dogrupe,odsifreartikla,dosifreartikla,oddatuma,dodatuma")
	searchInput := common.CreateSearchInput("search-input", translator, robnoKompodaciURLKupciArt, fmt.Sprintf("#%s", robnoKompodaciURLKupciArt), hxValsRobnoKompodaciRealizKupciArt)

	tmpl_robno.RobnoKompodaciPregledRealizacijePoKupcimaArtiklima(*h.tabs, tbl, grpValues, btnObrada, btnPrint, searchInput, session.SelectedGod, translator).Render(ctx, c.Writer)

}

func (h *RobnoKompodaciHandler) GetPregledRealizacijePoKupcimaArtiklimaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return
	}

	params := h.params(c)

	// Get company info
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgGetData)
		return
	}

	// Get table data
	tbl := common.SetTableBasicData(robnoKompodaciKupciArtTitle, robnoKompodaciKupciArtTableID, h.service.GetPregledRealizacijePoKupcimaArtiklimaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetPregledRealizacijePoKupcimaArtiklima(ctx, &tbl, true, 0, 0, params, common.TipStampePrint); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error retrieving data: %s", err.Error()))
		return
	}

	// Map trziste value
	trziste := c.Query("trziste")
	trzisteName := "-"
	switch trziste {
	case "1":
		trzisteName = "DOMAĆE"
	case "2":
		trzisteName = "IZVOZ"
	case "3":
		trzisteName = "UKUPNO"
	}

	// Prepare report parameters
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoKompodaciKupciArtTitle,
		ParameterItems: map[string]domain.ParameterItem{
			"Trziste":   {Name: "Tržište", Value: trzisteName},
			"OdGrupe":   {Name: "Od grupe", Value: params.OdGrupe},
			"DoGrupe":   {Name: "Do grupe", Value: params.DoGrupe},
			"Konto":     {Name: "Konto", Value: c.Query("konto")},
			"OdArtikla": {Name: "Od artikla", Value: params.OdArtikla},
			"DoArtikla": {Name: "Do artikla", Value: params.DoArtikla},
			"OdDatuma":  {Name: "Od datuma", Value: params.OdDatuma},
			"DoDatuma":  {Name: "Do datuma", Value: params.DoDatuma},
		},
	}

	tmpl_rep_rob.RobnoKompodaciPregledRealizacijePoKupcimaArtiklimaStampa(repParams, params, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func (h *RobnoKompodaciHandler) PregledRealizacijePoArtiklima(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	common.SetActiveTab(h.tabs, 1)
	tbl := common.SetTableBasicData(robnoKompodaciArtTitle, robnoKompodaciArtTableID, h.service.GetPregledRealizacijePoArtiklimaTableFields(), "", robnoKompodaciURLArt, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKompodaciArtTableID, robnoKompodaciURLArt, false, false, false)
	if common.IsDataRequest(c) {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPregledRealizacijePoArtiklima(ctx, &tbl, true, pageSize, page, h.params(c), common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Greska prilikom dobavljanja resultata iz beckend-a, greska: %s", err.Error()))
			return
		}
		if err := h.service.GetPregledRealizacijePoArtiklima(ctx, &tbl, false, pageSize, page, h.params(c), common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Greska prilikom dobavljanja resultata iz beckend-a, greska: %s", err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	grpValues, _ := h.service.GetRobneGrupeComboValues(ctx)
	btnObrada := common.SetButton("obrada-btn", "Obradi", "fin_obrada", robnoKompodaciURLArt, "#"+robnoKompodaciArtTableID, "innerHTML", "GET", "", hxValsRobnoKompodaciArt, true, common.ClassSaveButton, "")
	btnPrint := common.SetPrintButton("print-btn", "Štampa", "stampa", robnoKompodaciURLArtStampa, "GET", true, common.ClassPrintButton, "trziste,odgrupe,dogrupe,odsifreartikla,dosifreartikla,oddatuma,dodatuma")
	searchInput := common.CreateSearchInput("search-input", translator, robnoKompodaciURLArt, fmt.Sprintf("#%s", robnoKompodaciURLArt), hxValsRobnoKompodaciArt)

	tmpl_robno.RobnoKompodaciPregledRealizacijePoArtiklima(*h.tabs, tbl, grpValues, btnObrada, btnPrint, searchInput, session.SelectedGod, translator).Render(ctx, c.Writer)

}

func (h *RobnoKompodaciHandler) GetPregledRealizacijePoArtiklimaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return
	}

	params := h.params(c)

	// Get company info
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgGetData)
		return
	}

	// Get table data
	tbl := common.SetTableBasicData(robnoKompodaciArtTitle, robnoKompodaciArtTableID, h.service.GetPregledRealizacijePoArtiklimaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetPregledRealizacijePoArtiklima(ctx, &tbl, true, 0, 0, params, common.TipStampePrint); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error retrieving data: %s", err.Error()))
		return
	}

	// Map trziste value
	trziste := c.Query("trziste")
	trzisteName := "-"
	switch trziste {
	case "1":
		trzisteName = "DOMAĆE"
	case "2":
		trzisteName = "IZVOZ"
	case "3":
		trzisteName = "UKUPNO"
	}

	// Prepare report parameters
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoKompodaciArtTitle,
		ParameterItems: map[string]domain.ParameterItem{
			"OdGrupe":   {Name: "Od grupe", Value: params.OdGrupe},
			"DoGrupe":   {Name: "Do grupe", Value: params.DoGrupe},
			"OdArtikla": {Name: "Od artikla", Value: params.OdArtikla},
			"DoArtikla": {Name: "Do artikla", Value: params.DoArtikla},
			"OdDatuma":  {Name: "Od datuma", Value: params.OdDatuma},
			"DoDatuma":  {Name: "Do datuma", Value: params.DoDatuma},
			"Trziste":   {Name: "Tržište", Value: trzisteName},
		},
	}

	tmpl_rep_rob.RobnoKompodaciPregledRealizacijePoArtiklimaStampa(repParams, params, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func (h *RobnoKompodaciHandler) PregledUcescaArtikla(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	translator := i18n.GetInstance()
	common.SetActiveTab(h.tabs, 2)
	tbl := common.SetTableBasicData(robnoKompodaciUcesceArtTitle, robnoKompodaciUcesceArtTableID, h.service.GetPregledUcescaArtiklaTableFields(), "", robnoKompodaciURLUcesceArt, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKompodaciUcesceArtTableID, robnoKompodaciURLUcesceArt, false, false, false)
	tbl.URLGetAll = robnoKompodaciURLUcesceArt
	tbl.URLPrefix = robnoKompodaciURLUcesceArt
	tbl.Pagination.HxVals = hxValsRobnoKompodaciUcesceArt
	if common.IsDataRequest(c) {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		filedsError := common.ValidateRequiredParams(c, []string{"odsifreartikla", "dosifreartikla", "oddatuma", "dodatuma"})
		if len(filedsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, filedsError, fmt.Sprintf("Greska prilikom dobavljanja resultata iz beckend-a, greska: %s", fmt.Errorf("missing required parameters").Error()))
			return
		}
		if err := h.service.GetPregledUcescaArtikla(ctx, &tbl, true, pageSize, page, h.params(c), common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Greska prilikom preuzimanja podataka, greska: %s", err.Error()))
			return
		}
		if err := h.service.GetPregledUcescaArtikla(ctx, &tbl, false, pageSize, page, h.params(c), common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Greska prilikom preuzimanja podataka, greska: %s", err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	btnObrada := common.SetButton("obrada-btn", "Obradi", "fin_obrada", robnoKompodaciURLUcesceArt, "#"+robnoKompodaciUcesceArtTableID, "innerHTML", "GET", "", hxValsRobnoKompodaciUcesceArt, true, common.ClassSaveButton, "")
	btnPrint := common.SetPrintButton("print-btn", "Štampa", "stampa", robnoKompodaciURLUcesceArtStampa, "GET", true, common.ClassPrintButton, "odsifreartikla,dosifreartikla,oddatuma,dodatuma")
	searchInput := common.CreateSearchInput("search-input", translator, robnoKompodaciURLUcesceArt, fmt.Sprintf("#%s", robnoKompodaciUcesceArtTableID), hxValsRobnoKompodaciUcesceArt)

	tmpl_robno.RobnoKompodaciPregledUcescaArtikla(*h.tabs, tbl, btnObrada, btnPrint, searchInput, session.SelectedGod, translator).Render(ctx, c.Writer)
}

func (h *RobnoKompodaciHandler) GetPregledUcescaArtiklaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return
	}

	params := h.params(c)

	// Get company info
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgGetData)
		return
	}

	// Get table data
	tbl := common.SetTableBasicData(robnoKompodaciUcesceArtTitle, robnoKompodaciUcesceArtTableID, h.service.GetPregledUcescaArtiklaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetPregledUcescaArtikla(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error retrieving data: %s", err.Error()))
		return
	}

	// Prepare report parameters
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoKompodaciUcesceArtTitle,
		ParameterItems: map[string]domain.ParameterItem{
			"OdArtikla": {Name: "Od artikla", Value: params.OdArtikla},
			"DoArtikla": {Name: "Do artikla", Value: params.DoArtikla},
			"OdDatuma":  {Name: "Od datuma", Value: params.OdDatuma},
			"DoDatuma":  {Name: "Do datuma", Value: params.DoDatuma},
		},
	}

	tmpl_rep_rob.RobnoKompodaciUcescaArtiklaStampa(repParams, params, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func (h *RobnoKompodaciHandler) PregledUcescaGrupeArtikala(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	translator := i18n.GetInstance()
	common.SetActiveTab(h.tabs, 3)
	tbl := common.SetTableBasicData(robnoKompodaciUcesceGrTitle, robnoKompodaciUcesceGrTableID, h.service.GetPregledUcescaGrupeArtikalaTableFields(), "", robnoKompodaciURLUcesceGr, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoKompodaciUcesceGrTableID, robnoKompodaciURLUcesceGr, false, false, false)
	if common.IsDataRequest(c) {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		filedsError := common.ValidateRequiredParams(c, []string{"odgrupe", "dogrupe", "oddatuma", "dodatuma"})
		if len(filedsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, filedsError, fmt.Sprintf("Greska prilikom dobavljanja resultata iz beckend-a, greska: %s", fmt.Errorf("missing required parameters").Error()))
			return
		}
		if err := h.service.GetPregledUcescaGrupeArtikala(ctx, &tbl, true, pageSize, page, h.params(c), common.TipStampePreview); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Greska prilikom dobavljanja resultata iz beckend-a, greska: %s", err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	grpValues, _ := h.service.GetRobneGrupeComboValues(ctx)
	btnObrada := common.SetButton("obrada-btn", "Obradi", "fin_obrada", robnoKompodaciURLUcesceGr, "#"+robnoKompodaciUcesceGrTableID, "innerHTML", "GET", "", hxValsRobnoKompodaciUcesceGr, true, common.ClassSaveButton, "handleDialogResponse")
	btnPrint := common.SetPrintButton("print-btn", "Štampa", "stampa", robnoKompodaciURLUcesceGrStampa, "GET", true, common.ClassPrintButton, "odgrupe,dogrupe,oddatuma,dodatuma")

	searchInput := common.CreateSearchInput("search-input", translator, robnoKompodaciURLUcesceGr, fmt.Sprintf("#%s", robnoKompodaciURLUcesceGr), hxValsRobnoKompodaciUcesceGr)
	tmpl_robno.RobnoKompodaciPregledUcescaGrupeArtikala(*h.tabs, tbl, grpValues, btnObrada, btnPrint, searchInput, session.SelectedGod, translator).Render(ctx, c.Writer)
}

func (h *RobnoKompodaciHandler) params(c *gin.Context) domain.RobnoKomPodaciParams {
	return domain.RobnoKomPodaciParams{OdGrupe: c.Query("odgrupe"), DoGrupe: c.Query("dogrupe"), OdArtikla: c.Query("odsifreartikla"), DoArtikla: c.Query("dosifreartikla"), OdDatuma: c.Query("oddatuma"), DoDatuma: c.Query("dodatuma"), Trziste: c.Query("trziste"), SearchText: c.Query("query")}
}

func (h *RobnoKompodaciHandler) GetPregledUcescaGrupeArtikalaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return
	}

	params := h.params(c)

	// Validate required parameters
	filedsError := common.ValidateRequiredParams(c, []string{"odgrupe", "dogrupe", "oddatuma", "dodatuma"})
	if len(filedsError) > 0 {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, filedsError, "Missing required parameters")
		return
	}

	// Get company info
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgGetData)
		return
	}

	// Get table data
	tbl := common.SetTableBasicData(robnoKompodaciUcesceGrTitle, robnoKompodaciUcesceGrTableID, h.service.GetPregledUcescaGrupeArtikalaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetPregledUcescaGrupeArtikala(ctx, &tbl, true, 0, 0, params, common.TipStampePrint); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error retrieving data: %s", err.Error()))
		return
	}

	// Prepare report parameters
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoKompodaciUcesceGrTitle,
		ParameterItems: map[string]domain.ParameterItem{
			"OdGrupe":  {Name: "Od grupe", Value: params.OdGrupe},
			"DoGrupe":  {Name: "Do grupe", Value: params.DoGrupe},
			"OdDatuma": {Name: "Od datuma", Value: params.OdDatuma},
			"DoDatuma": {Name: "Do datuma", Value: params.DoDatuma},
		},
	}

	tmpl_rep_rob.RobnoKompodaciUcescaGrupeArtikalaStampa(repParams, params, tbl, i18n.GetInstance()).Render(ctx, c.Writer)
}

func (h *RobnoKompodaciHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())
	r.GET("/api/robno-kompodaci", h.RobnoKompodaciMain)
	r.GET("/api/robno-kompodaci/realizacija-po-kupcima-artiklima", h.PregledRealizacijePoKupcimaArtiklima)
	r.GET("/api/robno-kompodaci/realizacija-po-kupcima-artiklima/stampa", h.GetPregledRealizacijePoKupcimaArtiklimaStampa)
	r.GET("/api/robno-kompodaci/realizacija-po-artiklima", h.PregledRealizacijePoArtiklima)
	r.GET("/api/robno-kompodaci/realizacija-po-artiklima/stampa", h.GetPregledRealizacijePoArtiklimaStampa)
	r.GET("/api/robno-kompodaci/ucesce-artikla", h.PregledUcescaArtikla)
	r.GET("/api/robno-kompodaci/ucesce-artikla/stampa", h.GetPregledUcescaArtiklaStampa)
	r.GET("/api/robno-kompodaci/ucesce-grupe-artikala", h.PregledUcescaGrupeArtikala)
	r.GET("/api/robno-kompodaci/ucesce-grupe-artikala/stampa", h.GetPregledUcescaGrupeArtikalaStampa)
}

func robnoKompodaciTabs() *domain.TabData {
	return &domain.TabData{Tabs: []domain.TabItem{
		{ID: "robno-kompodaci-real-kupci-artikli", Label: "Realizacija po artiklima i kupcima", HXRequestUrl: robnoKompodaciURLKupciArt, Name: "real-kupci-artikli"},
		{ID: "robno-kompodaci-real-artikli", Label: "Realizacija po artiklima", HXRequestUrl: robnoKompodaciURLArt, Name: "real-artikli"},
		{ID: "robno-kompodaci-ucesce-artikal", Label: "Ucesce artikla u ukupnom prometu", HXRequestUrl: robnoKompodaciURLUcesceArt, Name: "ucesce-artikal"},
		{ID: "robno-kompodaci-ucesce-grupa", Label: "Ucesce grupe artikala u ukupnom prometu", HXRequestUrl: robnoKompodaciURLUcesceGr, Name: "ucesce-grupa"},
	}}
}
