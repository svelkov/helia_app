package handler

import (
	"fmt"
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	"helia/pkg/utils"
	"log"
	"net/http"

	tmpl1 "helia/frontend/templates/opstipodaci"
	tmpl2 "helia/frontend/templates/reports/opstipodaci"

	"github.com/gin-gonic/gin"
)

const (
	partneriContentTitle     string = "PARTNERI"
	partneriTableID          string = "partneri-table"
	tekruciRacuniTableID     string = "tekuciracuni-table"
	partneriURLPrefix        string = "/api/partneri"
	partneriURLGetAll        string = "/api/partneri/all"
	partneriURLCreate        string = "/api/partneri/create"
	partneriURLPrint         string = "/api/partneri/stampa/dialog"
	partneriURLConfirmAdd    string = "/api/partneri/confirm-add"
	partneriURLConfirmUpdate string = "/api/partneri/confirm-update"
	partneriURLConfirmDelete string = "/api/partneri/confirm-delete"
	hxValsPIB                string = `js:{
            			"pib": document.getElementById("pib")?.value
       					}`
)

type PartneriHandler struct {
	Service service.PartneriService
	cfg     config.Config
	lm      *middleware.LockMiddleware
}

func NewPartneriHandler(service service.PartneriService, cfg config.Config, lm *middleware.LockMiddleware) *PartneriHandler {
	return &PartneriHandler{Service: service, cfg: cfg, lm: lm}
}

func (h *PartneriHandler) GetAllPartneri(c *gin.Context) {
	tbl := common.SetTableBasicData(partneriContentTitle, partneriTableID, SetPartneriFields(), partneriURLPrefix, partneriURLGetAll, 0, 0, 0, 0, h.cfg)

	btnPrint := common.SetButton("stampa-btn", "Štampa", "fin_print", partneriURLPrint, "#dialog-partneri-stampa", "outerHTML", "GET", "", "", true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), partneriURLGetAll, fmt.Sprintf("#%s", partneriTableID), "")

	common.SetTableConfig(&tbl, partneriContentTitle, partneriTableID, true, true, false)
	tbl.URLGetAll = partneriURLGetAll
	tbl.URLPrefix = partneriURLPrefix
	tbl.BtnAdd.HxRequestType = "GET"
	tbl.BtnAdd.HxActionURL = partneriURLConfirmAdd
	tbl.BtnAdd.HxTarget = "#dialog-content"
	tbl.BtnUpdate.HxRequestType = "GET"
	tbl.BtnDelete.HxRequestType = "GET"
	tbl.BtnUpdate.HxActionURL = partneriURLConfirmUpdate
	tbl.BtnDelete.HxActionURL = partneriURLConfirmDelete
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	searchText := c.Query("search")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	err := h.Service.GetAllPartneri(c.Request.Context(), &tbl, true, page, pageSize, searchText, sortBy, sortOrder)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error fetching partneri: %v", err))
		return
	}
	err = h.Service.GetAllPartneri(c.Request.Context(), &tbl, false, page, pageSize, searchText, sortBy, sortOrder)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error fetching partneri: %v", err))
		return
	}
	tmpl1.PartneriMain(tbl, searchInput, btnPrint, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *PartneriHandler) PartneriConfirmAdd(c *gin.Context) {
	ctx := c.Request.Context()
	csrfToken := common.GetCsrfToken(c)
	dialog := domain.Dialog{
		Id:            "partneri-add-dialog",
		Title:         "Dodaj partnera",
		HxActionURL:   "/api/partneri/create",
		HxRequestType: "POST",
	}
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Sačuvaj",
		HxActionURL:      "/api/partneri/create",
		HxRequestType:    "POST",
		IdDialog:         dialog.Id,
		BtnClass:         common.ClassSaveButton,
		HxOnAfterRequest: fmt.Sprintf("handleDialogResponse('%s')", dialog.Id),
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
	btnProveriPIB := setButtonProveriPIB(dialog.Id)
	entity := domain.Partneri{}
	lastID, _ := h.Service.GetLastPartneriID(ctx)
	entity.Sifra = fmt.Sprintf("%d", lastID)
	tblTekRacuni := common.SetTableBasicData("Tekuci racuni", tekruciRacuniTableID, h.Service.GetTekuciRacuniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tblTekRacuni.ShowActions = true
	tblTekRacuni.BtnDelete.IsVisible = true
	tblTekRacuni.BtnUpdate.IsVisible = false
	tipoviAnalitike, err := h.Service.GetTipoveAnalitike(ctx)
	if err != nil {
		log.Printf("Error fetching tipove analitike: %v", err)
		tipoviAnalitike = []domain.ComboItem{} // fallback to empty list
	}
	tmpl1.PartneriFormMain(entity, tblTekRacuni, dialog, tipoviAnalitike, btnSacuvaj, btnCancel, btnClose, btnProveriPIB, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *PartneriHandler) PartneriCreate(c *gin.Context) {
	var entity domain.Partneri
	ctx := c.Request.Context()
	log.Println("PartneriCreate called, caller:", c.Request.RequestURI)

	// Parse request to get partner data
	if err := c.ShouldBind(&entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	filedsErrors, err := h.Service.ValidacijaPartneri(ctx, &entity, common.ActionAdd)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Error validating partner")
		return
	}
	if len(filedsErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, filedsErrors, "Validation errors")
		return
	}

	// Extract tekracuni data from form BEFORE creating partner
	tekRacuniList := extractTekRacuniFromForm(c)

	// Create partner and tekracuni together
	newID, err := h.Service.CreateWithTekRacuni(ctx, &entity, tekRacuniList)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error creating partner: %v", err))
		return
	}

	log.Printf("Partner created successfully with ID: %d", newID)
}
func (h *PartneriHandler) PartneriConfirmDelete(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, SetPartneriFields(), "#info-message")
}

func (h *PartneriHandler) PartneriConfirmUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	csrfToken := common.GetCsrfToken(c)
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	id, err := utils.GetInt64FromQueryRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, "Invalid ID")
		return
	}
	dialog := domain.Dialog{
		Id:            "partneri-update-dialog",
		Title:         "Izmeni partnera",
		HxActionURL:   fmt.Sprintf("/api/partneri/update/%d", id),
		HxRequestType: "PUT",
	}
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Sačuvaj",
		HxActionURL:      fmt.Sprintf("/api/partneri/update/%d", id),
		HxRequestType:    "PUT",
		IdDialog:         dialog.Id,
		BtnClass:         common.ClassSaveButton,
		HxOnAfterRequest: fmt.Sprintf("handleDialogResponse('%s')", dialog.Id),
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
	btnProveriPIB := setButtonProveriPIB(dialog.Id)
	entity, err := h.Service.GetByID(ctx, common.IDpartneri, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Error fetching partner")
		return
	}

	tblTekRacuni := common.SetTableBasicData("Tekuci racuni", tekruciRacuniTableID, h.Service.GetTekuciRacuniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tblTekRacuni.ShowActions = true
	tblTekRacuni.BtnDelete.IsVisible = true
	tblTekRacuni.BtnUpdate.IsVisible = false
	err = h.Service.GetTekuciRacuni(ctx, id, &tblTekRacuni)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Error fetching tekući računi")
		return
	}
	tipoviAnalitike, err := h.Service.GetTipoveAnalitike(ctx)
	if err != nil {
		log.Printf("Error fetching tipove analitike: %v", err)
		tipoviAnalitike = []domain.ComboItem{}
	}
	tmpl1.PartneriFormMain(*entity, tblTekRacuni, dialog, tipoviAnalitike, btnSacuvaj, btnCancel, btnClose, btnProveriPIB, i18n.GetInstance(), csrfToken).Render(c.Request.Context(), c.Writer)
}

func (h *PartneriHandler) PartneriUpdate(c *gin.Context) {
	var entity domain.Partneri
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, "Unauthorized")
		return
	}
	id, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, "Invalid ID")
		return
	}
	// Parse request to get partner data
	if err := c.ShouldBind(&entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	filedsErrors, err := h.Service.ValidacijaPartneri(ctx, &entity, common.ActionUpdate)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Error validating partner")
		return
	}
	if len(filedsErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, filedsErrors, "Validation errors")
		return
	}

	// Extract tekracuni data from form BEFORE updating partner
	tekRacuniList := extractTekRacuniFromForm(c)

	// Update partner and tekracuni together
	err = h.Service.UpdateWithTekRacuni(ctx, &entity, id, tekRacuniList)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, fmt.Sprintf("Error updating partner: %v", err))
		return
	}

	log.Printf("Partner updated successfully with ID: %d", id)
}

// Get PIB from NBS
func (h *PartneriHandler) CheckPIBForPartner(c *gin.Context) {
	ctx := c.Request.Context()
	pib := c.Query("pib")
	if pib == "" {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "PIB not provided")
		return
	}
	companyInfo, err := h.Service.GetCompanyByPIB(pib)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "PIB not provided")
		return
	}

	// Get tekRacuni data
	type TekRacuniRow struct {
		Redbroj    string `json:"redbroj"`
		Brojracuna string `json:"brojracuna"`
		Banka      string `json:"banka"`
	}

	tbl := common.SetTableBasicData("Tekuci racuni", tekruciRacuniTableID, h.Service.GetTekuciRacuniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	h.Service.GetCompanyAccountsByPIB(&tbl, pib)

	// Convert table rows to tekRacuni rows
	tekRacuniRows := make([]TekRacuniRow, 0)
	for i, row := range tbl.Rows {
		if len(row.Fields) >= 3 {
			tekRacuniRows = append(tekRacuniRows, TekRacuniRow{
				Redbroj:    fmt.Sprintf("%d", i+1),
				Brojracuna: row.Fields[1],
				Banka:      row.Fields[2],
			})
		}
	}

	lastID, err := h.Service.GetLastPartneriID(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, "Failed to get last partner ID")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"naziv":     companyInfo.Naziv,
		"naziv1":    companyInfo.Naziv1,
		"mesto":     companyInfo.Mesto,
		"pobro":     companyInfo.PBroj,
		"adresa":    companyInfo.Adresa,
		"matbr":     companyInfo.MatBr,
		"tekRacuni": tekRacuniRows,
		"sifra":     lastID,
	})
}

func setButtonProveriPIB(dialogID string) domain.Button {
	btnProveriPIB := domain.Button{
		Id:               "btnProveriPIB",
		IsVisible:        true,
		LabelText:        "Proveri PIB",
		HxActionURL:      "/api/partneri/proveripib",
		HxRequestType:    "GET",
		IdDialog:         dialogID,
		Icon:             "snimi",
		HxVals:           hxValsPIB,
		HxOnAfterRequest: "handleProveriPIBResponse(event)",
		BtnClass:         common.ClassPrintButton,
	}
	return btnProveriPIB
}

// extractTekRacuniFromForm extracts tekracuni data from form inputs
// Form sends data as tekracuni[0].redbroj, tekracuni[0].brojracuna, tekracuni[0].banka, etc.
func extractTekRacuniFromForm(c *gin.Context) []domain.TekRacuni {
	result := make([]domain.TekRacuni, 0)

	// Get all form values
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		return result
	}

	// Debug: Log all form keys
	log.Printf("DEBUG: All form keys received:")
	for key, values := range c.Request.PostForm {
		log.Printf("  Key: '%s', Values: %v", key, values)
	}

	// Extract tekracuni entries - they come as tekracuni[0].redbroj, tekracuni[0].brojracuna, tekracuni[0].banka
	tekRacuniMap := make(map[int]domain.TekRacuni)

	for key, values := range c.Request.PostForm {
		if len(values) == 0 {
			continue
		}
		value := values[0]

		// Parse keys like "tekracuni[0].redbroj"
		if len(key) > len("tekracuni[") && key[:len("tekracuni[")] == "tekracuni[" {
			// Find the index
			endBracket := len("tekracuni[")
			for endBracket < len(key) && key[endBracket] != ']' {
				endBracket++
			}
			if endBracket < len(key) {
				indexStr := key[len("tekracuni["):endBracket]
				var index int
				fmt.Sscanf(indexStr, "%d", &index)

				log.Printf("DEBUG: Parsing tekracuni[%d], key: '%s', value: '%s'", index, key, value)

				// Get or create the entry
				tr := tekRacuniMap[index]

				// Parse the field name
				fieldStart := endBracket + 2 // skip "].
				if fieldStart < len(key) {
					fieldName := key[fieldStart:]

					switch fieldName {
					case "redbroj":
						// Ignore redbroj as it's just a display field
					case "brojracuna":
						tr.TekRac = value
						log.Printf("  Set TekRac to: %s", value)
					case "banka":
						tr.BnkCod = value
						log.Printf("  Set BnkCod to: %s", value)
					}

					tekRacuniMap[index] = tr
				}
			}
		}
	}

	log.Printf("DEBUG: tekRacuniMap has %d entries", len(tekRacuniMap))

	// Convert map to slice, only include non-empty entries
	for i := 0; i < len(tekRacuniMap); i++ {
		if tr, ok := tekRacuniMap[i]; ok {
			if tr.TekRac != "" || tr.BnkCod != "" {
				log.Printf("DEBUG: Adding tekracuni[%d] - TekRac: '%s', BnkCod: '%s'", i, tr.TekRac, tr.BnkCod)
				result = append(result, tr)
			}
		}
	}

	log.Printf("DEBUG: Extracted %d tekracuni total", len(result))
	return result
}

// GetPartneriForm returns the appropriate form fragment based on tipanalitike.
// Called by HTMX when the "Tip analitike" combo changes.
func (h *PartneriHandler) GetPartneriForm(c *gin.Context) {
	ctx := c.Request.Context()
	csrfToken := common.GetCsrfToken(c)
	tipAnalitike := c.Query("tipanalitike")

	dialog := domain.Dialog{
		Id:            "partneri-add-dialog",
		Title:         "Dodaj partnera",
		HxActionURL:   "/api/partneri/create",
		HxRequestType: "POST",
	}
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Sačuvaj",
		HxActionURL:      "/api/partneri/create",
		HxRequestType:    "POST",
		IdDialog:         dialog.Id,
		BtnClass:         common.ClassSaveButton,
		HxOnAfterRequest: fmt.Sprintf("handleDialogResponse('%s')", dialog.Id),
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
	btnProveriPIB := setButtonProveriPIB(dialog.Id)
	entity := domain.Partneri{}
	tblTekRacuni := common.SetTableBasicData("Tekuci racuni", tekruciRacuniTableID, h.Service.GetTekuciRacuniTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	tblTekRacuni.ShowActions = true
	tblTekRacuni.BtnDelete.IsVisible = true
	tblTekRacuni.BtnUpdate.IsVisible = false

	switch tipAnalitike {
	// Add cases here as you create new form templates, e.g.:
	case "5": // Fizicka lica
		tmpl1.PartneriFormFizickaLica(entity, tblTekRacuni, dialog, []domain.ComboItem{}, btnSacuvaj, btnCancel, btnClose, btnProveriPIB, i18n.GetInstance(), csrfToken).Render(ctx, c.Writer)
	default:
		tmpl1.PartneriFormKomintenti(entity, tblTekRacuni, dialog, []domain.ComboItem{}, btnSacuvaj, btnCancel, btnClose, btnProveriPIB, i18n.GetInstance(), csrfToken).Render(ctx, c.Writer)
	}
}
func (h *PartneriHandler) PartneriStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")
	searchText := c.Query("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	fvrData, err := h.Service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	repParams := domain.ReportParameters{
		Orientation:    "landscape",
		CompanyName:    fvrData.Naziv,
		Adress:         fvrData.Adresa,
		Postcode:       fvrData.Pobro,
		City:           fvrData.Mesto,
		PIB:            fvrData.PIB,
		MatBroj:        fvrData.Matbr,
		ReportName:     "Kontni Plan",
		ParameterItems: map[string]domain.ParameterItem{},
	}
	fmt.Println(repParams)
	tbl := common.SetTableBasicData(partneriContentTitle, partneriTableID, h.Service.GetPartneriTableFields(), "", partneriURLGetAll, 0, 0, 0, 0, h.cfg)
	err = h.Service.GetAllPartneri(c.Request.Context(), &tbl, true, page, pageSize, searchText, sortBy, sortOrder)
	//err = h.Service.GetAllPartneri(ctx, &tbl, page, pageSize, true, sortBy, sortOrder, searchText, common.TipStampePrint)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	//translator := i18n.GetInstance()
	//tmpl_fin_rep.PartneriStampa(repParams, tbl, translator).Render(ctx, c.Writer)
}

func (h *PartneriHandler) PartnerStampaDialog(c *gin.Context) {
	dialog := domain.Dialog{
		Id:    "dialog-partneri-stampa",
		Title: "Štampanje partnera",
	}
	btnClose := domain.Button{
		Id:        "btn-close-partneri-stampa",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel-partneri-stampa",
		LabelText: "Odustani",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassCloseButton,
	}
	err := tmpl1.PartnerStampaDialog(dialog, btnClose, btnCancel, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgRenderTemplate)
	}
}
func (h *PartneriHandler) PartnerStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUserSessionNotFound)
		return
	}
	odSifre := c.Query("odsifre")
	doSifre := c.Query("dosifre")
	odMesta := c.Query("odmesta")
	doMesta := c.Query("domesta")
	partnerNaziv := c.Query("partnernaziv")
	konto := c.Query("konto")
	sortirajStampu := c.Query("sortirajstampu")

	fvrData, err := h.Service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
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
		ReportName:  "POSLOVNI PARTNERI",
		ParameterItems: map[string]domain.ParameterItem{
			"OdSifre":            {Name: "Od šifre", Value: odSifre},
			"DoSifre":            {Name: "Do šifre", Value: doSifre},
			"OdMesta":            {Name: "Od Mesta", Value: odMesta},
			"DoMesta":            {Name: "Do Mesta", Value: doMesta},
			"PartnerNazivSadrzi": {Name: "Partner Naziv sadrži", Value: partnerNaziv},
			"Konto":              {Name: "Konto", Value: konto},
			"SortirajStampu":     {Name: "Sortiraj stampu", Value: sortirajStampu},
		},
	}
	partneriParams := domain.PartneriParameters{
		OdSifre:        odSifre,
		DoSifre:        doSifre,
		OdMesta:        odMesta,
		DoMesta:        doMesta,
		PartnerNaziv:   partnerNaziv,
		Konto:          konto,
		SortirajStampu: sortirajStampu,
	}
	tbl := common.SetTableBasicData(partneriContentTitle, partneriTableID, h.Service.GetPartneriStampaFields(), "", partneriURLGetAll, 0, 0, 0, 0, h.cfg)
	err = h.Service.GetAllStampa(ctx, &tbl, partneriParams)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	translator := i18n.GetInstance()
	tmpl2.PartneriStampa(tbl, repParams, translator).Render(ctx, c.Writer)

}

func (h *PartneriHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	r.GET("/api/partneri/all", h.GetAllPartneri)
	r.POST("/api/partneri/create", h.PartneriCreate)
	r.PUT("/api/partneri/update/:id", h.PartneriUpdate)
	r.GET("/api/partneri/proveripib", h.CheckPIBForPartner)
	r.GET("/api/partneri/form", h.GetPartneriForm)
	r.GET("/api/partneri/confirm-add", h.PartneriConfirmAdd)
	r.GET("/api/partneri/confirm-delete", h.lm.WithEntityLockHold("fkpl", "id"), h.PartneriConfirmDelete)
	r.GET("/api/partneri/confirm-update", h.lm.WithEntityLockHold("fkpl", "id"), h.PartneriConfirmUpdate)
	r.GET("/api/partneri/stampa/dialog", h.PartnerStampaDialog)
	r.GET("/api/partneri/stampa", h.PartnerStampa)

}
