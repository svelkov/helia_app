package finansijsko

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	tmpl "helia/frontend/templates/finansijsko"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	fproContentTitle    string = "FPRO"
	fproTableID         string = "fpro-table"
	fproURLPrefix       string = "/api/fpro/"
	fproURLGetAll       string = "/api/fpro/all/tipdok"
	fproURLGetAllSearch string = "/api/nalozi/stavke/{idfnalog}"
	fproURLNextFpro     string = "/api/fpro/nextfpro"
)

// FproHandler handles requests related to "fpro" entities.
type FproHandler struct {
	fproService service.FproService // Use the interface
	btnSave     domain.Button
	btnNazad    domain.Button
}

func NewFproHandler(service service.FproService) *FproHandler {
	handler := &FproHandler{fproService: service}
	handler.setHandlerFieldValues()
	return handler
}

func (h *FproHandler) GetFpro(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.fproService, h.fproService.GetTableStavkeFields(), utils.IDfpro)
}

func (h *FproHandler) CreateFpro(w http.ResponseWriter, r *http.Request) {
	var fpro domain.Fpro
	lastInsertedID, err := utils.CreateHelper(w, r, &fpro, h.fproService, utils.IDfpro, h.fproService.GetTableStavkeFields())
	if err != nil {
		return
	}
	// Lock the header
	mu, _ := headerLocks.LoadOrStore(lastInsertedID, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
}

func (h *FproHandler) UpdateFpro(w http.ResponseWriter, r *http.Request) {
	var fpro domain.Fpro
	fproID := r.URL.Query().Get("id")
	// Lock the header
	mu, _ := headerLocks.LoadOrStore(fproID, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	utils.UpdateHelper(w, r, &fpro, h.fproService, h.fproService.GetTableStavkeFields(), utils.IDfpro)
}

func (h *FproHandler) DeleteFpro(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper[domain.Fpro](w, r, h.fproService, utils.IDfpro)
}

func (h *FproHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, h.fproService.GetTableStavkeFields())
}

func (h *FproHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(fproURLPrefix, "/"), h.fproService.GetTableStavkeFields())
}

func (h *FproHandler) getFproStavkeHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper[domain.Fpro](w, r, nil, h.fproService.GetTableStavkeFields(), utils.IDfpro)
}

// --- CONSOLIDATED VIEW DATA HANDLER ---
func (h *FproHandler) GetNalogStavke(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")
	idFnalParam := parts[len(parts)-1] //
	searchQuery := r.URL.Query().Get("query")
	searchControl := domain.InputControl{
		ID:           "search-controldetail",
		Label:        "Pretraži stavke naloga",
		Type:         "search",
		Placeholder:  "Unesite broj naloga ili druge podatke",
		HxActionURL:  fmt.Sprintf("/api/fpro/nalog/%s", idFnalParam),
		HxTarget:     "#nalozi_kopiranje_stavke",
		HxSwap:       "innerHTML",
		HxInclude:    "#search-controldetail",
		Class:        utils.ClassSearchInput,
		Autocomplete: "off",
	}
	if idFnalParam == "" && searchQuery == "" {
		err := tmpl.NaloziDetail(domain.TableData{}, searchControl).Render(r.Context(), w)
		if err != nil {
			respondWithError(w, utils.RenderTemplateErr, err, http.StatusInternalServerError)
			return
		}
	}

	idFnal, err := strconv.ParseInt(idFnalParam, 10, 64)
	if err != nil {
		respondWithError(w, utils.InvalidIDErrMsg, fmt.Errorf("invalid idfnal: %s", idFnalParam), http.StatusBadRequest)
		return
	}

	page, pageSize := common.GetPageAndPageSizeFromRequest(r)
	// Call the service to get ALL necessary data for the view
	tbl, err := h.fproService.GetNaloziStavke(r, idFnal, searchQuery, page, pageSize, h.fproService.GetTableStavkeFields())
	if err != nil {
		respondWithError(w, utils.ReadDataErrMsg, err, http.StatusInternalServerError)
		return
	}
	tbl.TableID = "nalozi_kopiranje_stavke"
	tbl.URLGetAll = "/api/fpro/nalog/" + idFnalParam

	err = tmpl.NaloziDetail(tbl, searchControl).Render(r.Context(), w)
	if err != nil {
		respondWithError(w, utils.ReadDataErrMsg, err, http.StatusInternalServerError)
		return
	}
}

func (h *FproHandler) populateTableRows(tableData domain.TableData, entities []domain.Fpro, fieldsDef []domain.Fields) []domain.TableRow {
	var tableRows []domain.TableRow
	fieldCache := h.fproService.GetFieldCache()

	for _, entity := range entities {
		val := reflect.ValueOf(entity)
		idValue, _, found := common.GetFieldByNameCaseInsensitive(val, utils.IDfpro)
		id := ""
		if found {
			id = fmt.Sprintf("%v", idValue)
		}
		var fields []string
		for _, fieldName := range fieldsDef {
			fieldInfo, found := fieldCache[strings.ToLower(fieldName.Name)]
			if !found {
				continue // or return error if field is required
			}

			value := common.GetFormattedValue(fieldInfo, val.FieldByName(fieldInfo.Name))
			fields = append(fields, value)
			fields = append(fields, "")
		}
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: tableData.BtnUpdate.IsVisible, HasDelete: tableData.BtnDelete.IsVisible}
		tableRows = append(tableRows, row)
	}
	return tableRows
}

/*
	 func (h *FproHandler) FproPrikaz(w http.ResponseWriter, r *http.Request) {
		err := tmpl_fin.FproContent(h.tabData, nil, domain.UkupnaObrada{}, h.btnSave, h.btnNoviFpro, domain.TableData{}).Render(r.Context(), w)
		if err != nil {
			response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.RenderTemplateErr, http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}
*/
func (h *FproHandler) AddRoutes(r *http.ServeMux) {
	// r.HandleFunc("GET /api/fpro/all", infrastructure.AuthMiddleware(h.GetFproMainView))
	// r.HandleFunc("GET /api/fpro/all/tipdok", infrastructure.AuthMiddleware(h.GetFproMainView))
	// r.HandleFunc("GET /api/fpro/all/search", infrastructure.AuthMiddleware(h.GetFproMainView))
	// r.HandleFunc("GET /api/fpro/knjizenje", infrastructure.AuthMiddleware(h.GetFproMainView))

	r.HandleFunc("POST /api/fpro", infrastructure.AuthMiddleware(h.CreateFpro))
	r.HandleFunc("GET /api/fpro/nalog/{idfnalog}", infrastructure.AuthMiddleware(h.GetNalogStavke))
	r.HandleFunc("GET /api/fpro/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/fpro/confirm-update", infrastructure.AuthMiddleware(h.getFproStavkeHandler))
	r.HandleFunc("GET /api/fpro/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/fpro/{id}", infrastructure.AuthMiddleware(h.GetFpro))
	r.HandleFunc("PUT /api/fpro/{id}", infrastructure.AuthMiddleware(h.UpdateFpro))
	r.HandleFunc("DELETE /api/fpro/{id}", infrastructure.AuthMiddleware(h.DeleteFpro))
	// r.HandleFunc("GET /api/fpro/prepis", infrastructure.AuthMiddleware(h.FproPrepis))
	// r.HandleFunc("GET /api/fpro/storniranje", infrastructure.AuthMiddleware(h.FproStorniranje))
	// r.HandleFunc("GET /api/fpro/prikaz", infrastructure.AuthMiddleware(h.FproPrikaz))
}

func (h *FproHandler) setHandlerFieldValues() {
	h.btnSave = domain.Button{
		Id:            "btn-save",
		LabelText:     "Snimi Fpro",
		HxActionURL:   fproURLPrefix,
		HxTarget:      "this",
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
	}
	h.btnNazad = domain.Button{
		Id:            "btn-nazad-fpro",
		LabelText:     "Nazad",
		HxActionURL:   fproURLNextFpro,
		HxRequestType: "GET",
	}
}
