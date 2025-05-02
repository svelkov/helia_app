package finansijsko

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	naloziContentTitle = "NALOZI"
	naloziTableID      = "nalozi-table"
	naloziURLPrefix    = "/api/nalozi/"
)

var naloziTableFields = []domain.Fields{
	{Name: "rbr", Label: "R. Broj", Width: "4"},
	{Name: "tipdok", Label: "Vrsta Naloga", Width: "6"},
	{Name: "nalog", Label: "Br. Naloga", Width: "12"},
	{Name: "danal", Label: "Datum naloga", Width: "12"},
	{Name: "opis", Label: "opis ", Width: "60"},
	{Name: "dug", Label: "Duguje", Width: "14"},
	{Name: "pot", Label: "Potrazuje", Width: "14"},
	{Name: "datob", Label: "Datum obrade", Width: "12"},
	{Name: "brst", Label: "Br.Stavki", Width: "5"},
	{Name: "nalsts", Label: "Status naloga", Width: "10"},
}
var tabData = domain.TabData{
	Tabs: []domain.TabItem{
		{ID: "knjizenje", Label: "Knjiženje Naloga", HXRequestUrl: fmt.Sprintf("%sknjizenje", naloziURLPrefix), IsActive: true},
		{ID: "prepis", Label: "Prepis Naloga", HXRequestUrl: fmt.Sprintf("%sprepis", naloziURLPrefix), IsActive: false},
		{ID: "storniranje", Label: "Storniranje naloga", HXRequestUrl: fmt.Sprintf("%sstorniranje", naloziURLPrefix), IsActive: false},
		{ID: "prikaz", Label: "Prikaz/Štampa naloga", HXRequestUrl: fmt.Sprintf("%sprikaz", naloziURLPrefix), IsActive: false},
	},
}
var btnSave = domain.Button{
	Id:            "btn-save",
	LabelText:     "Snimi Nalog",
	HxActionURL:   naloziURLPrefix,
	HxTarget:      "this",
	HxSwap:        "innerHTML",
	HxOn:          "click: this.trigger('save')",
	HxRequestType: "POST",
}
var btnNoviNalog = domain.Button{
	Id:            "btn-novi-nalog",
	LabelText:     "Novi Nalog",
	HxActionURL:   naloziURLPrefix,
	HxTarget:      "this",
	HxSwap:        "innerHTML",
	HxOn:          "click: this.trigger('save')",
	HxRequestType: "POST",
}

type FnalHandler struct {
	Service        *service.BaseService[domain.Fnal]
	tipdok_service *service.BaseService[domain.Tipdok]
}

func NewFnalHandler(service *service.BaseService[domain.Fnal], tipdok_service *service.BaseService[domain.Tipdok]) *FnalHandler {
	return &FnalHandler{
		Service:        service,
		tipdok_service: tipdok_service,
	}
}

func (h *FnalHandler) CreateNalog(w http.ResponseWriter, r *http.Request) {
	var nalog domain.Fnal
	utils.CreateHelper(w, r, &nalog, h.Service, naloziTableFields)
}

func (h *FnalHandler) UpdateNalog(w http.ResponseWriter, r *http.Request) {
	var nalog domain.Fnal
	utils.UpdateHelper(w, r, &nalog, h.Service, naloziTableFields, utils.IDfnal)
}

func (h *FnalHandler) DeleteNalog(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDfnal)
}

func (h *FnalHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, naloziTableFields)
}

func (h *FnalHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(naloziURLPrefix, "/"), naloziTableFields)
}

func (h *FnalHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, naloziTableFields, utils.IDfnal)
}

func (h *FnalHandler) GetNalog(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, naloziTableFields, utils.IDfnal)
}

func (h *FnalHandler) GetAllFnalTipdok(w http.ResponseWriter, r *http.Request) {
	//var nalog domain.Fnal
	var err error
	qryText := ""
	whereText := ""
	orderBy := ""
	limitOffset := ""
	tipdokValues := &[]domain.Tipdok{}
	isSearch := true
	args := []interface{}{}
	//check if the tipdok is selected from the dropdown
	// If not, set the default value to the first item in the list
	hasGod, hasKar := h.tipdok_service.Repo.CheckGogKar()
	basicWhere := h.tipdok_service.Repo.CreateBasicWhere(naloziTableFields, &args, hasGod, hasKar)

	tipdok := r.URL.Query().Get("tipdok")
	if tipdok == "" {
		isSearch = false
		qryText = "SELECT idtipdok, tipdok, opis FROM tipdok "
		whereText = fmt.Sprintf(" %s AND (grpdok = 'FIN' OR grpdok = 'SVI') ", basicWhere)
		orderBy = " ORDER BY tipdok"
		limitOffset = ""
		tipdokValues, err = h.tipdok_service.GetAllCustom(qryText, whereText, args, limitOffset, orderBy)
		if err != nil {
			response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.ReadDataErrMsg, http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		if len(*tipdokValues) > 0 {
			tipdok = (*tipdokValues)[0].TipDok
		}
	}
	// Fetch all "fnal" entities from the service layer

	qryText = `SELECT count(*) from fnal `
	// fnal.idfnal, fnal.tipdok, fnal.nalog, fnal.danal, fnal.opis, fnal.dug, fnal.pot, fnal.datob, fnal.brst, fnal.nalsts FROM fnal`
	whereText = fmt.Sprintf("%s AND (tipdok = '%s') ", basicWhere, tipdok)
	orderBy = ""
	limitOffset = ""
	totRecords, err := h.Service.GetTotalRecordsCustom(qryText, whereText, args, "", "")
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	currentPage, pageSize, totalPages := utils.GetPaginationData(r, totRecords)
	pageSize = 12
	qryText = `SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal `
	whereText = fmt.Sprintf(" %s AND (tipdok = '%s') ", basicWhere, tipdok)
	limitOffset = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, (currentPage-1)*pageSize)
	orderBy = " ORDER BY danal desc"
	allEntities, err := h.Service.GetAllCustom(qryText, whereText, args, limitOffset, orderBy)
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	table := utils.SetTableBasicData(naloziContentTitle, naloziTableID, naloziTableFields, naloziURLPrefix, pageSize, currentPage, totalPages, totRecords)

	for _, entity := range *allEntities {
		val := reflect.ValueOf(entity)
		if val.Kind() == reflect.Ptr {
			val = val.Elem() // Handle pointer case
		}

		// Get ID field value dynamically
		idValue, _, found := utils.GetFieldByNameCaseInsensitive(val, utils.IDfnal)
		id := ""
		if found {
			id = fmt.Sprintf("%v", idValue)
		}

		// Extract specified fields dynamically
		var fields []string
		for _, fieldName := range naloziTableFields {
			fieldValue, filedType, found := utils.GetFieldByNameCaseInsensitive(val, fieldName.Name)
			if found {
				value, err := utils.GetFormattedValue(filedType, fieldValue)
				if err != nil {
					FormatValueErrMsg := fmt.Sprintf("Error formatting value: %s", fieldValue)
					response := utils.CreateResponse(w, false, []domain.FieldError{}, fmt.Sprintf(FormatValueErrMsg, err.Error()), http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}
				fields = append(fields, value)
				continue
			}
		}
		// Create table row
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: false, HasDelete: false}
		table.Rows = append(table.Rows, row)
	}
	table.ShowActions = false
	table.HXInclude = "#tipdokSelect"
	if isSearch {
		// Render only the table
		err = tmpl.Table(*table).Render(r.Context(), w)
		if err != nil {
			response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.RenderTemplateErr, http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

	} else {
		ukObrada := domain.UkupnaObrada{}
		// Render the template for nalozi with the data
		err = tmpl_fin.NaloziContent(tabData, tipdokValues, ukObrada, btnSave, btnNoviNalog, *table).Render(r.Context(), w)
		if err != nil {
			response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.RenderTemplateErr, http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}
}
func (h *FnalHandler) GetAllFnal(w http.ResponseWriter, r *http.Request) {
	//var nalog domain.Fnal
	searchValue := r.URL.Query().Get("query")
	tipdok := r.URL.Query().Get("tipdok")

	fmt.Println("searchValue", searchValue)
	args := []interface{}{}
	hasGod, hasKar := h.Service.Repo.CheckGogKar()
	basicWhere := h.Service.Repo.CreateBasicWhere(naloziTableFields, &args, hasGod, hasKar, searchValue)
	// Fetch all "fnal" entities from the service layer
	qryText := `SELECT count(*) from fnal `
	// fnal.idfnal, fnal.tipdok, fnal.nalog, fnal.danal, fnal.opis, fnal.dug, fnal.pot, fnal.datob, fnal.brst, fnal.nalsts FROM fnal`
	whereText := fmt.Sprintf("%s AND (tipdok = '%s') ", basicWhere, tipdok)
	orderBy := ""
	limitOffset := ""
	totRecords, err := h.Service.GetTotalRecordsCustom(qryText, whereText, args, "", "")
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	currentPage, pageSize, totalPages := utils.GetPaginationData(r, totRecords)
	pageSize = 12
	qryText = `SELECT idfnal, tipdok, nalog, danal, opis, dug, pot, datob, brst, nalsts FROM fnal `
	whereText = fmt.Sprintf("%s AND (tipdok = '%s') ", basicWhere, tipdok)
	limitOffset = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, (currentPage-1)*pageSize)
	orderBy = " ORDER BY danal desc"
	allEntities, err := h.Service.GetAllCustom(qryText, whereText, args, limitOffset, orderBy)
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	table := utils.SetTableBasicData(naloziContentTitle, naloziTableID, naloziTableFields, naloziURLPrefix, pageSize, currentPage, totalPages, totRecords)

	for _, entity := range *allEntities {
		val := reflect.ValueOf(entity)
		if val.Kind() == reflect.Ptr {
			val = val.Elem() // Handle pointer case
		}

		// Get ID field value dynamically
		idValue, _, found := utils.GetFieldByNameCaseInsensitive(val, utils.IDfnal)
		id := ""
		if found {
			id = fmt.Sprintf("%v", idValue)
		}

		// Extract specified fields dynamically
		var fields []string
		for _, fieldName := range naloziTableFields {
			fieldValue, filedType, found := utils.GetFieldByNameCaseInsensitive(val, fieldName.Name)
			if found {
				value, err := utils.GetFormattedValue(filedType, fieldValue)
				if err != nil {
					FormatValueErrMsg := fmt.Sprintf("Error formatting value: %s", fieldValue)
					response := utils.CreateResponse(w, false, []domain.FieldError{}, fmt.Sprintf(FormatValueErrMsg, err.Error()), http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}
				fields = append(fields, value)
				continue
			}
		}
		// Create table row
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: false, HasDelete: false}
		table.Rows = append(table.Rows, row)
	}
	table.ShowActions = false

	// Render the template with the data
	// Render only the table
	err = tmpl.Table(*table).Render(r.Context(), w)
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.RenderTemplateErr, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

}

func (h *FnalHandler) AddRoutes(r *http.ServeMux) {
	// Define routes for nalozi
	r.HandleFunc("POST /api/nalozi", infrastructure.AuthMiddleware(h.CreateNalog))
	r.HandleFunc("GET /api/nalozi/all", infrastructure.AuthMiddleware(h.GetAllFnal))
	r.HandleFunc("GET /api/nalozi/all/tipdok", infrastructure.AuthMiddleware(h.GetAllFnalTipdok))
	r.HandleFunc("GET /api/nalozi/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/nalozi/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/nalozi/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/nalozi/{id}", infrastructure.AuthMiddleware(h.GetNalog))
	r.HandleFunc("PUT /api/nalozi/{id}", infrastructure.AuthMiddleware(h.UpdateNalog))
	r.HandleFunc("DELETE /api/nalozi/{id}", infrastructure.AuthMiddleware(h.DeleteNalog))
	r.HandleFunc("GET /api/nalozi/knjizenje", infrastructure.AuthMiddleware(h.GetAllFnal))
	r.HandleFunc("GET /api/nalozi/prepis", infrastructure.AuthMiddleware(h.GetAllFnal))
	r.HandleFunc("GET /api/nalozi/storniranje", infrastructure.AuthMiddleware(h.GetAllFnal))
	r.HandleFunc("GET /api/nalozi/prikaz", infrastructure.AuthMiddleware(h.GetAllFnal))
}
