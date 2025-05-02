package utils

import (
	"encoding/json"
	"fmt"
	"helia/internal/domain"
	"helia/internal/service"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
)

func DeleteHelper[T any](w http.ResponseWriter, r *http.Request, service service.Service[T], idType string) {
	// Get the `id` from the URL path
	id, err := GetIdFromUrl(r)
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
	entity, err := service.GetByID(idField, id)
	if err != nil {
		http.Error(w, "No record for update is available", http.StatusBadRequest)
		return
	}
	url = fmt.Sprintf("%s%d", url, id)
	fields := service.MapEntityToValues(entity, tableFields)
	dialog := SetDialogValues(idStr, url, "Izmena podataka", "hx-put")
	RenderDialogContent(w, r, dialog, fields, ActionUpdate)
}

func CreateHelper[T any](w http.ResponseWriter, r *http.Request, entity *T, service service.Service[T], tableFields []domain.Fields) {
	// Parse the form data
	err := r.ParseForm()
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
	fieldErrors, err := service.Create(entity, tableFields)
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
}

func UpdateHelper[T any](w http.ResponseWriter, r *http.Request, entity *T, service service.Service[T], tableFields []domain.Fields, idField string) {
	// Get the `id` from the URL path
	redirectURL := fmt.Sprintf("%s/all", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")])
	id, err := GetIdFromUrl(r)
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

func GetAllEntityHelper[T any](w http.ResponseWriter, r *http.Request, entity *T, service service.Service[T], tableFields []domain.Fields, entityContentTitle, entityTableID, entityURLPrefix, idField string, hasUpdateDelete ...bool) {

	// Parse query parameters from the URL
	searchValue := r.URL.Query().Get("query")
	// Fetch all "drzava" entities from the service layer
	totRecords, err := service.GetTotalRecords(tableFields, searchValue)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	currentPage, pageSize, totalPages := GetPaginationData(r, totRecords)
	allEntities, err := service.GetAll(pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert the fetched data into the format expected by the template
	table := SetTableBasicData(entityContentTitle, entityTableID, tableFields, entityURLPrefix, pageSize, currentPage, totalPages, totRecords)

	for _, entity := range *allEntities {
		val := reflect.ValueOf(entity)
		if val.Kind() == reflect.Ptr {
			val = val.Elem() // Handle pointer case
		}

		// Get ID field value dynamically
		idValue, _, found := GetFieldByNameCaseInsensitive(val, idField)
		id := ""
		if found {

			id = fmt.Sprintf("%v", idValue)
		}

		// Extract specified fields dynamically
		var fields []string
		for _, fieldName := range tableFields {
			fieldValue, filedType, found := GetFieldByNameCaseInsensitive(val, fieldName.Name)
			if found {
				value, err := GetFormattedValue(filedType, fieldValue)
				if err != nil {
					FormatValueErrMsg := fmt.Sprintf("Error formatting value: %s", fieldValue)
					response := CreateResponse(w, false, []domain.FieldError{}, fmt.Sprintf(FormatValueErrMsg, err.Error()), http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}
				fields = append(fields, value)
				continue
			}
		}
		// Create table row
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: true, HasDelete: true}
		if len(hasUpdateDelete) > 0 {
			row.HasUpdate = hasUpdateDelete[0]
		}
		if len(hasUpdateDelete) > 1 {
			row.HasDelete = hasUpdateDelete[1]
		}
		table.Rows = append(table.Rows, row)
	}

	RenderContent(w, r, *table)
}

func GetFormattedValue(fieldType string, fieldValue reflect.Value) (string, error) {
	switch fieldType {
	case "int", "int64", "int32":
		return fmt.Sprintf("%d", fieldValue.Int()), nil
	case "float64":
		return fmt.Sprintf("%.2f", fieldValue.Float()), nil
	case "string":
		return fmt.Sprintf("%s", fieldValue.String()), nil
	case "bool":
		return strconv.FormatBool(fieldValue.Bool()), nil
	case "time", "date", "datetime", "timestamp":
		t, ok := fieldValue.Interface().(time.Time)
		if !ok {
			return "", fmt.Errorf("fieldValue is not of type time.Time")
		}
		return t.Format("2006-01-02"), nil
	default:
		return fmt.Sprintf("%v", fieldValue), nil
	}
}

func GetEntityHelper[T any](w http.ResponseWriter, r *http.Request, service service.Service[T], tableFields []domain.Fields, idField string) {
	// Get the `id` from the URL path
	id, err := GetIdFromUrl(r)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, GetIdFromUrlErrMsg, http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	entity, err := service.GetByID(idField, id)
	if err != nil {
		response := CreateResponse(w, false, []domain.FieldError{}, ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	json.NewEncoder(w).Encode(entity)
}

// getFieldByNameCaseInsensitive searches for a field name case-insensitively
func GetFieldByNameCaseInsensitive(val reflect.Value, fieldName string) (reflect.Value, string, bool) {
	fieldNameLower := strings.ToLower(fieldName)

	// Iterate over struct fields
	for i := 0; i < val.NumField(); i++ {
		structField := val.Type().Field(i)
		if strings.ToLower(structField.Name) == fieldNameLower {
			return val.Field(i), strings.ToLower(val.Field(i).Type().Name()), true
		}
	}
	return reflect.Value{}, "", false
}
