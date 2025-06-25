package utils

import (
	"encoding/json"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/service"

	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
)

func DeleteHelper[T any](w http.ResponseWriter, r *http.Request, service service.Service[T], idType string) {
	// Get the `id` from the URL path
	id, err := GetIDFromRequest(r, "id")
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, GetIdFromUrlErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Call the delete method
	err = service.Delete(idType, id)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, fmt.Sprintf(DeleteDataErrMsg, err.Error()), http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Success response
	response := CreateResponse(w, true, []domain.FieldError{}, DeleteDataOkMsg, http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func ConfirmDeleteHelper(w http.ResponseWriter, r *http.Request, tableFields []domain.Fields) {
	idStr := r.URL.Query().Get("id")
	url := r.URL.Query().Get("url")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, InvalidIdErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	url = fmt.Sprintf("%s%d", url, id)
	dialog := SetDialogValues(idStr, url, "Brisanje podataka", "hx-delete")
	RenderDialogContent(w, r, dialog, tableFields, ActionDelete)
}

func ConfirmAddHelper(w http.ResponseWriter, r *http.Request, url string, tableFields []domain.Fields) {
	for i, field := range tableFields {
		field.Value = ""
		tableFields[i] = field
	}
	dialog := SetDialogValues("", url, "Unos novih podataka", "hx-post")
	RenderDialogContent(w, r, dialog, tableFields, ActionAdd)

}

func ConfirmUpdateHelper[T any](w http.ResponseWriter, r *http.Request, service service.Service[T], tableFields []domain.Fields, idField string) {

	idStr := r.URL.Query().Get("id")
	url := r.URL.Query().Get("url")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, InvalidIdErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	entity, err := service.GetByID(idField, int64(id))
	if err != nil {
		http.Error(w, "No record for update is available", http.StatusBadRequest)
		return
	}
	url = fmt.Sprintf("%s%d", url, id)
	fields := service.MapEntityToValues(entity, tableFields)
	dialog := SetDialogValues(idStr, url, "Izmena podataka", "hx-put")
	RenderDialogContent(w, r, dialog, fields, ActionUpdate)
}

func CreateHelper[T any](w http.ResponseWriter, r *http.Request, entity *T, service service.Service[T], idField string, tableFields []domain.Fields) (insertedId int64, err error) {
	// Parse the form data
	err = r.ParseForm()
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ParseFormErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create a decoder
	decoder := schema.NewDecoder()

	// Decode the form data into the struct
	err = decoder.Decode(entity, r.PostForm)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, FormDecodeErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, lastInsertedID, err := service.Create(entity, idField, tableFields)
	if err != nil {
		response := CreateResponse(w, false, fieldErrors, SaveDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	if len(fieldErrors) > 0 {
		response := CreateResponse(w, false, fieldErrors, ValidationErrMsg, http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(response)
		return
	}
	response := CreateResponse(w, true, fieldErrors, SaveDataOkMsg, http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	return lastInsertedID, err
}

func UpdateHelper[T any](w http.ResponseWriter, r *http.Request, entity *T, service service.Service[T], tableFields []domain.Fields, idField string) {
	// Get the `id` from the URL path
	redirectURL := fmt.Sprintf("%s/all", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")])
	id, err := GetIDFromRequest(r, "id")
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, GetIdFromUrlErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	// Parse the form data
	err = r.ParseForm()
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ParseFormErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create a decoder
	decoder := schema.NewDecoder()

	// Decode the form data into the struct
	err = decoder.Decode(entity, r.PostForm)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, FormDecodeErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	tableFields = service.MapEntityToValues(entity, tableFields)

	fieldErrors, err := service.Update(entity, idField, id, tableFields)
	if len(fieldErrors) > 0 {
		response := CreateResponse(w, false, fieldErrors, ValidationErrMsg, http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(response)
		return
	}
	if err != nil {
		response := CreateResponse(w, false, fieldErrors, SaveDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreateResponse(w, true, fieldErrors, SaveDataOkMsg, http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	// redirect to the preview site
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func GetAllEntityHelper[T any](w http.ResponseWriter, r *http.Request, entity *T, service service.Service[T], tableFields []domain.Fields, entityContentTitle, entityTableID, entityURLPrefix, entityURLGetall, idField string, hasUpdateDelete ...bool) *domain.TableData {

	// Parse query parameters from the URL
	searchValue := r.URL.Query().Get("query")
	// Fetch all "drzava" entities from the service layer
	totRecords, err := service.GetTotalRecords(tableFields, searchValue)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(r, totRecords)
	allEntities, err := service.GetAll(pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return nil
	}

	// Convert the fetched data into the format expected by the template
	table := common.SetTableBasicData(entityContentTitle, entityTableID, tableFields, entityURLPrefix, entityURLGetall, pageSize, currentPage, totalPages, totRecords)
	common.SetTableRows(&table, *allEntities, tableFields, idField, entityURLPrefix, service.GetFieldCache())

	return &table
	// RenderContent(w, r, *table)
}

func GetEntityHelper[T any](w http.ResponseWriter, r *http.Request, service service.Service[T], tableFields []domain.Fields, idField string) {
	// Get the `id` from the URL path
	id, err := GetIDFromRequest(r, "id")
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, GetIdFromUrlErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	entity, err := service.GetByID(idField, int64(id))
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	json.NewEncoder(w).Encode(entity)
}
