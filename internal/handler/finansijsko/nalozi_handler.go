package finansijsko

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	naloziContentTitle    string = "NALOZI"
	naloziTableID         string = "nalozi-table"
	naloziURLPrefix       string = "/api/nalozi/"
	naloziURLGetAll       string = "/api/nalozi/all/tipdok"
	naloziURLGetAllSearch string = "/api/nalozi/all/search"
	naloziURLNextNalog    string = "/api/nalozi/nextnalog"
)

var headerLocks = sync.Map{}

// FnalHandler handles requests related to "fnal" entities.
type FnalHandler struct {
	naloziService service.NalogService
	tabData       domain.TabData
	btnSave       domain.Button
	btnNoviNalog  domain.Button
	typeView      string // Type of view, e.g. "knjizenje", "prepis", etc.
}

func NewFnalHandler(nalogService service.NalogService) *FnalHandler {
	handler := &FnalHandler{
		naloziService: nalogService,
	}
	handler.setHanlderFieldValues()
	return handler
}

func (h *FnalHandler) GetNalog(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.naloziService, h.naloziService.GetNaloziTableFields(), utils.IDfnal)
}

func (h *FnalHandler) GetNextNalog(w http.ResponseWriter, r *http.Request) {
	tipdok := r.URL.Query().Get("tipdok")

	nextNalog, err := h.naloziService.GetNextNalog(tipdok)
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, "Failed to get next Nalog", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	// Set headers and return JSON
	tmpl := `<input class="mt-1 block p-1 w-[10ch] border border-gray-300 rounded-md shadow-sm focus:ring-blue-100 focus:border-blue-100 w-full min-w-0" 
                  type="text" 
                  id="brojNaloga" 
                  maxlength="10" 
                  tabindex="3" 
                  name="brojNaloga" 
                  value="%d">`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, tmpl, nextNalog)
}

func (h *FnalHandler) CreateNalog(w http.ResponseWriter, r *http.Request) {
	// var nalog domain.Fnal
	// lastInsertedID, err := utils.CreateHelper(w, r, &nalog, h.naloziService, utils.IDfnal, h.naloziService.GetNaloziTableFields())
	// if err != nil {
	// 	return
	// }
	// // Lock the header
	// mu, _ := headerLocks.LoadOrStore(lastInsertedID, &sync.Mutex{})
	// mutex := mu.(*sync.Mutex)
	// mutex.Lock()
	err := tmpl_fin.NalogKnjizenjeStavke().Render(r.Context(), w)
	if err != nil {
		respondWithError(w, utils.ReadDataErrMsg, err, http.StatusInternalServerError)
		return
	}
}

func (h *FnalHandler) UpdateNalog(w http.ResponseWriter, r *http.Request) {
	var nalog domain.Fnal
	fnalID := r.URL.Query().Get("id")
	// Lock the header
	mu, _ := headerLocks.LoadOrStore(fnalID, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	utils.UpdateHelper(w, r, &nalog, h.naloziService, h.naloziService.GetNaloziTableFields(), utils.IDfnal)

}

func (h *FnalHandler) DeleteNalog(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.naloziService, utils.IDfnal)
}

func (h *FnalHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, h.naloziService.GetNaloziTableFields())
}

func (h *FnalHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(naloziURLPrefix, "/"), h.naloziService.GetNaloziTableFields())
}

func (h *FnalHandler) getNalogStavkeHandler(w http.ResponseWriter, r *http.Request) {
	// Render update form with entity data.
	utils.ConfirmUpdateHelper[domain.Fnal](w, r, nil, h.naloziService.GetNaloziTableFields(), utils.IDfnal) // Placeholder
}

// --- CONSOLIDATED VIEW DATA HANDLER ---

func (h *FnalHandler) GetNalogMainView(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("query")
	selectedTipdok := r.URL.Query().Get("tipdok")

	page, pageSize := common.GetPageAndPageSizeFromRequest(r)

	// Determine if this is an initial full page request (not an HTMX swap)
	isInitialRequest := selectedTipdok == ""
	if isInitialRequest {
		h.naloziService.SetSfTableFields(sfTableFields)
		h.naloziService.SetNalogIDFieldName(utils.IDfnal)
	}
	// Call the service to get ALL necessary data for the view
	viewData, err := h.naloziService.GetNalogsViewData(r, searchQuery, selectedTipdok, page, pageSize, isInitialRequest)
	if err != nil {
		respondWithError(w, utils.ReadDataErrMsg, err, http.StatusInternalServerError)
		return
	}

	// Populate table rows based on entities from service
	viewData.TableData.Rows = h.populateTableRows(viewData.TableData, viewData.FnalEntities, h.naloziService.GetNaloziTableFields())

	// Update tab active state based on current view (usually default tab is active for main view)
	currentTabData := h.tabData // Make a copy to modify active states
	currentTabData.Tabs[0].IsActive = true
	currentTabData.Tabs[1].IsActive = false
	currentTabData.Tabs[2].IsActive = false
	currentTabData.Tabs[3].IsActive = false
	searchControl := domain.InputControl{
		ID:           "search-control",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  naloziURLGetAllSearch,
		HxTarget:     "#table-body",
		HxSwap:       "innerHTML",
		HxInclude:    "#tipdokSelect",
		Autocomplete: "off",
	}
	// Render the appropriate template based on whether it's an initial load or HTMX swap
	if viewData.IsInitialLoad {
		err = tmpl_fin.NaloziContent(
			currentTabData,
			viewData.TipdokOptions,
			*viewData.UkupnaObrada, // Dereference as template expects value
			h.btnSave,
			h.btnNoviNalog,
			viewData.TableData,
			searchControl).Render(r.Context(), w)
	} else {
		// HTMX request, just render the table component
		err = tmpl.Table(viewData.TableData).Render(r.Context(), w)
	}

	if err != nil {
		respondWithError(w, utils.RenderTemplateErr, err, http.StatusInternalServerError)
	}
}

// FnalPrepis
func (h *FnalHandler) FnalPrepis(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(r)

	// Get data for the header table
	viewData, err := h.naloziService.GetNalogsPrepisData(r, searchQuery, page, pageSize) // Not initial load for prepis tab
	if err != nil {
		respondWithError(w, "Failed to get Nalog header data for prepis", err, http.StatusInternalServerError)
		return
	}
	viewData.TableData.Rows = h.populateTableRows(viewData.TableData, viewData.FnalEntities, h.naloziService.GetNaloziTableFields())
	viewData.TableData.DetailTarget = "#nalozi_kopiranje_stavke"
	viewData.TableData.DetailURL = "/api/fpro/nalog/" // URL for fetching details
	// Set active tab for prepis
	currentTabData := h.tabData
	currentTabData.Tabs[0].IsActive = false
	currentTabData.Tabs[1].IsActive = true
	currentTabData.Tabs[2].IsActive = false
	currentTabData.Tabs[3].IsActive = false
	searchControl := domain.InputControl{
		ID:           "search-prepishdr",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  "/api/nalozi/prepis",
		HxTarget:     "#table-body",
		HxSwap:       "innerHTML",
		HxInclude:    "#search-prepishdr, #idfnal",
		Autocomplete: "off",
	}
	if r.Header["Hx-Trigger"] != nil && r.Header["Hx-Trigger"][0] == "prepis" {
		err = tmpl_fin.NaloziKopiranje(currentTabData, viewData.TableData, domain.TableData{}, searchControl).Render(r.Context(), w)
		if err != nil {
			respondWithError(w, utils.RenderTemplateErr, err, http.StatusInternalServerError)
			return
		}
	} else {
		// If this is an HTMX request, we just render the table component
		err = tmpl.Table(viewData.TableData).Render(r.Context(), w)
		if err != nil {
			respondWithError(w, utils.RenderTemplateErr, err, http.StatusInternalServerError)
			return
		}
	}
}

// FnalPrepis, FnalStorniranje, FnalPrikaz:
func (h *FnalHandler) FnalPrepisDialog(w http.ResponseWriter, r *http.Request) {
	idFnalParam := r.URL.Query().Get("idfnal") //
	idFnal, err := strconv.ParseInt(idFnalParam, 10, 64)
	if err != nil {
		respondWithError(w, utils.InvalidIDErrMsg, fmt.Errorf("invalid idfnal: %s", idFnalParam), http.StatusBadRequest)
		return
	}

	// Get data for the header table
	result, err := h.naloziService.GetByID(utils.IDfnal, idFnal)
	if err != nil {
		respondWithError(w, "Failed to get Nalog header data for prepis", err, http.StatusInternalServerError)
		return
	}
	dialog := domain.Dialog{
		Id:            "nalog_kopiraj",
		Title:         "Kopiraj Nalog",
		OkText:        "Kopiraj",
		CancelText:    "Otkaži",
		SaveText:      "Snimi",
		HxTarget:      "#nalozi_kopiranje_stavke",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
	}
	modelView := domain.KopirajNalog{
		IDFnal:    idFnal,
		NalogOld:  fmt.Sprintf("%d", result.Nalog),
		TipdokOld: result.Tipdok,
		DanalOld:  result.Danal,
		OpisOld:   result.Opis,
	}
	content := tmpl_fin.NaloziKopiranjeDialog(modelView)

	err = tmpl.Dialog("nalog_kopiraj", content, dialog).Render(r.Context(), w)
	if err != nil {
		respondWithError(w, utils.RenderTemplateErr, err, http.StatusInternalServerError)
		return
	}
}

func (h *FnalHandler) FnalStorniranje(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("query")
	page, pageSize := common.GetPageAndPageSizeFromRequest(r)

	// Get data for the header table
	viewData, err := h.naloziService.GetNalogsStorniranjeData(r, searchQuery, page, pageSize) // Not initial load for prepis tab
	if err != nil {
		respondWithError(w, "Failed to get Nalog header data for storniranje", err, http.StatusInternalServerError)
		return
	}
	viewData.TableData.Rows = h.populateTableRows(viewData.TableData, viewData.FnalEntities, h.naloziService.GetNaloziTableFields())

	// Set active tab for prepis
	currentTabData := h.tabData
	currentTabData.Tabs[0].IsActive = false
	currentTabData.Tabs[1].IsActive = false
	currentTabData.Tabs[2].IsActive = true
	currentTabData.Tabs[3].IsActive = false

	searchControl := domain.InputControl{
		ID:           "search-control",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  naloziURLGetAllSearch,
		HxTarget:     "#table-body",
		HxSwap:       "innerHTML",
		HxInclude:    "#tipdokSelect",
		Autocomplete: "off",
	}

	if r.Header["Hx-Trigger"] != nil && r.Header["Hx-Trigger"][0] == "storniranje" {
		err = tmpl_fin.NaloziStorniranje(currentTabData, viewData.TableData, domain.TableData{}, searchControl).Render(r.Context(), w)
		if err != nil {
			respondWithError(w, utils.RenderTemplateErr, err, http.StatusInternalServerError)
			return
		}
	} else {
		// If this is an HTMX request, we just render the table component
		err = tmpl.Table(viewData.TableData).Render(r.Context(), w)
		if err != nil {
			respondWithError(w, utils.RenderTemplateErr, err, http.StatusInternalServerError)
			return
		}
	}
}

func (h *FnalHandler) FnalPrikaz(w http.ResponseWriter, r *http.Request) {
	// Render the template for nalozi with the data
	searchControl := domain.InputControl{
		ID:           "search-control",
		Label:        "Pretraži Naloge",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  naloziURLGetAllSearch,
		HxTarget:     "#table-body",
		HxSwap:       "innerHTML",
		HxInclude:    "#tipdokSelect",
		Autocomplete: "off",
	}
	err := tmpl_fin.NaloziContent(h.tabData, nil, domain.UkupnaObrada{}, h.btnSave, h.btnNoviNalog, domain.TableData{}, searchControl).Render(r.Context(), w)
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.RenderTemplateErr, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
}

// populateTableRows extracts and formats data for table display.
// This logic can stay in handler as it's UI formatting specific.
func (h *FnalHandler) populateTableRows(tableData domain.TableData, entities []domain.Fnal, fieldsDef []domain.Fields) []domain.TableRow {
	var tableRows []domain.TableRow

	fieldCache := h.naloziService.GetFieldCache()

	for _, entity := range entities {
		val := reflect.ValueOf(entity)
		// No need for val.Kind() == reflect.Ptr check if service returns value types

		// Get ID field value dynamically
		idValue, _, found := common.GetFieldByNameCaseInsensitive(val, utils.IDfnal)
		id := ""
		if found {
			id = fmt.Sprintf("%v", idValue)
		}

		// Extract specified fields dynamically
		var fields []string
		for _, fieldName := range fieldsDef {
			fieldInfo, found := fieldCache[strings.ToLower(fieldName.Name)]
			if !found {
				continue // or return error if field is required
			}

			value := common.GetFormattedValue(fieldInfo, val.FieldByName(fieldInfo.Name))
			fields = append(fields, value)
		}
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: tableData.BtnUpdate.IsVisible, HasDelete: tableData.BtnDelete.IsVisible}
		tableRows = append(tableRows, row)
	}
	return tableRows
}

// Helper function to render responses (can be in utils, or helper here)
func respondWithError(w http.ResponseWriter, msg string, err error, statusCode int) {
	response := utils.CreateResponse(w, false, []domain.FieldError{}, fmt.Sprintf("%s: %v", msg, err), statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *FnalHandler) AddRoutes(r *http.ServeMux) {
	// Consolidate main view routes
	r.HandleFunc("GET /api/nalozi/all", infrastructure.AuthMiddleware(h.GetNalogMainView))
	r.HandleFunc("GET /api/nalozi/all/tipdok", infrastructure.AuthMiddleware(h.GetNalogMainView)) // Both can hit same handler
	r.HandleFunc("GET /api/nalozi/all/search", infrastructure.AuthMiddleware(h.GetNalogMainView)) // Both can hit same handler
	r.HandleFunc("GET /api/nalozi/knjizenje", infrastructure.AuthMiddleware(h.GetNalogMainView))  // Also points to main view

	// ... (other CRUD and confirm routes)
	r.HandleFunc("POST /api/nalozi", infrastructure.AuthMiddleware(h.CreateNalog))
	r.HandleFunc("GET /api/nalozi/nextnalog", infrastructure.AuthMiddleware(h.GetNextNalog))
	r.HandleFunc("GET /api/nalozi/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/nalozi/confirm-update", infrastructure.AuthMiddleware(h.getNalogStavkeHandler))
	r.HandleFunc("GET /api/nalozi/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/nalozi/{id}", infrastructure.AuthMiddleware(h.GetNalog))
	r.HandleFunc("PUT /api/nalozi/{id}", infrastructure.AuthMiddleware(h.UpdateNalog))
	r.HandleFunc("DELETE /api/nalozi/{id}", infrastructure.AuthMiddleware(h.DeleteNalog))
	r.HandleFunc("GET /api/nalozi/prepis", infrastructure.AuthMiddleware(h.FnalPrepis))
	r.HandleFunc("GET /api/nalozi/confirm-prepis", infrastructure.AuthMiddleware(h.FnalPrepisDialog))
	r.HandleFunc("GET /api/nalozi/storniranje", infrastructure.AuthMiddleware(h.FnalStorniranje))
	r.HandleFunc("GET /api/nalozi/prikaz", infrastructure.AuthMiddleware(h.FnalPrikaz))
}

func (h *FnalHandler) setHanlderFieldValues() {
	h.typeView = "knjizenje" // Default view type, can be changed based on context
	h.tabData = domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "knjizenje", Label: "Knjiženje Naloga", HXRequestUrl: fmt.Sprintf("%sknjizenje", naloziURLPrefix), IsActive: true, Name: "knjizenje"},
			{ID: "prepis", Label: "Prepis Naloga", HXRequestUrl: fmt.Sprintf("%sprepis", naloziURLPrefix), IsActive: false, Name: "prepis"},
			{ID: "storniranje", Label: "Storniranje naloga", HXRequestUrl: fmt.Sprintf("%sstorniranje", naloziURLPrefix), IsActive: false, Name: "storniranje"},
			{ID: "prikaz", Label: "Prikaz/Štampa naloga", HXRequestUrl: fmt.Sprintf("%sprikaz", naloziURLPrefix), IsActive: false, Name: "stampa"},
		},
	}
	h.btnSave = domain.Button{
		Id:            "btn-save",
		LabelText:     "Snimi Nalog",
		HxActionURL:   "/api/nalozi",
		HxTarget:      "#dialog-stavkenaloga",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
	}
	h.btnNoviNalog = domain.Button{
		Id:            "btn-novi-nalog",
		LabelText:     "Novi Nalog",
		HxActionURL:   naloziURLNextNalog,
		HxInclude:     "#tipdokSelect",
		HxTarget:      "#brojNaloga",
		HxSwap:        "outerHTML",
		HxRequestType: "GET",
	}
}

func getActiveTab(r *http.Request) string {
	activeTab := r.URL.Query().Get("sourceTab")
	if activeTab != "" {
		return activeTab
	}
	return ""
}

func getIdActiveTab(sourceTab string) string {
	switch sourceTab {
	case "knjizenje":
		return "tabknjizenje"
	case "prepis":
		return "tabprepis"
	case "storniranje":
		return "tabstorniranje"
	case "prikaz":
		return "tabprikaz"
	default:
		return "" // Default to knjizenje if no match
	}
}
