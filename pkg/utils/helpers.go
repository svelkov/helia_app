package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
)

// Error messages (should be moved to a separate file if used across packages)
const (
	ErrMsgInvalidID    = "invalid ID provided"
	ErrMsgGetIDFromURL = "failed to get ID from URL"
	ErrMsgParseForm    = "failed to parse form"
	ErrMsgFormDecode   = "failed to decode form"
	ErrMsgReadData     = "failed to read data"
	ErrMsgSaveData     = "failed to save data"
	ErrMsgDeleteData   = "failed to delete data"
	ErrMsgValidation   = "validation failed"
	ErrMsgDataOk       = "operation successful"
)

// WriteJSONResponse writes a JSON response with the given status, success, errors, and message.
func WriteJSONResponse(
	w http.ResponseWriter,
	status int,
	success bool,
	errors []domain.FieldError,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(domain.Response{
		Success: success,
		Errors:  errors,
		Message: message,
	})
}

// parseAndDecode parses the form and decodes it into the provided entity.
func parseAndDecode(r *http.Request, entity any) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("%s: %w", ErrMsgParseForm, err)
	}
	decoder := schema.NewDecoder()
	if err := decoder.Decode(entity, r.PostForm); err != nil {
		return fmt.Errorf("%s: %w", ErrMsgFormDecode, err)
	}
	return nil
}

// GetIDFromRequest extracts the ID from the request URL path.
func GetIDFromRequest(r *http.Request, key string) (int64, error) {
	idStr := r.URL.Query().Get(key)
	if idStr == "" {
		return 0, errors.New(ErrMsgGetIDFromURL)
	}
	return strconv.ParseInt(idStr, 10, 64)
}

// DeleteHelper handles HTTP DELETE requests for a resource.
// It expects the resource ID in the URL path and uses the provided service to delete the resource.
func DeleteHelper[T any](
	w http.ResponseWriter,
	r *http.Request,
	service service.Service[T],
	idType string,
) {
	id, err := GetIDFromRequest(r, "id")
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	if err := service.Delete(idType, id); err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, fmt.Sprintf("%s: %v", ErrMsgDeleteData, err))
		return
	}

	WriteJSONResponse(w, http.StatusOK, true, nil, ErrMsgDataOk)
}

// ConfirmDeleteHelper renders a confirmation dialog for resource deletion.
func ConfirmDeleteHelper(
	w http.ResponseWriter,
	r *http.Request,
	tableFields []domain.Fields,
) {
	idStr := r.URL.Query().Get("id")
	url := r.URL.Query().Get("url")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, ErrMsgInvalidID)
		return
	}

	url = fmt.Sprintf("%s/%d", url, id)
	dialog := SetDialogValues(idStr, url, "Brisanje podataka", "hx-delete")
	RenderDialogContent(w, r, dialog, tableFields, ActionDelete)
}

// ConfirmAddHelper renders a dialog for adding a new resource.
func ConfirmAddHelper(
	w http.ResponseWriter,
	r *http.Request,
	url string,
	tableFields []domain.Fields,
) {
	for i := range tableFields {
		tableFields[i].Value = ""
	}
	dialog := SetDialogValues("", url, "Unos novih podataka", "hx-post")
	RenderDialogContent(w, r, dialog, tableFields, ActionAdd)
}

// ConfirmUpdateHelper renders a dialog for updating a resource.
func ConfirmUpdateHelper[T any](
	w http.ResponseWriter,
	r *http.Request,
	service service.Service[T],
	tableFields []domain.Fields,
	idField string,
) {
	idStr := r.URL.Query().Get("id")
	url := r.URL.Query().Get("url")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, ErrMsgInvalidID)
		return
	}

	entity, err := service.GetByID(idField, int64(id))
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, ErrMsgReadData)
		return
	}

	url = fmt.Sprintf("%s/%d", url, id)
	fields := service.MapEntityToValues(entity, tableFields)
	dialog := SetDialogValues(idStr, url, "Izmena podataka", "hx-put")
	RenderDialogContent(w, r, dialog, fields, ActionUpdate)
}

// CreateHelper handles HTTP POST requests for creating a new resource.
func CreateHelper[T any](
	w http.ResponseWriter,
	r *http.Request,
	entity *T,
	service service.Service[T],
	idField string,
	tableFields []domain.Fields,
) (insertedId int64, err error) {
	if err := parseAndDecode(r, entity); err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, err.Error())
		return 0, err
	}

	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, lastInsertedID, err := service.Create(entity, idField, tableFields)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, false, fieldErrors, ErrMsgSaveData)
		return 0, err
	}
	if len(fieldErrors) > 0 {
		WriteJSONResponse(w, http.StatusUnprocessableEntity, false, fieldErrors, ErrMsgValidation)
		return 0, nil
	}

	WriteJSONResponse(w, http.StatusCreated, true, nil, ErrMsgDataOk)
	return lastInsertedID, nil
}

// UpdateHelper handles HTTP PUT requests for updating a resource.
func UpdateHelper[T any](
	w http.ResponseWriter,
	r *http.Request,
	entity *T,
	service service.Service[T],
	tableFields []domain.Fields,
	idField string,
) {
	redirectURL := fmt.Sprintf("%s/all", r.URL.Path[:strings.LastIndex(r.URL.Path, "/")])
	id, err := GetIDFromRequest(r, "id")
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	if err := parseAndDecode(r, entity); err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, err := service.Update(entity, idField, id, tableFields)
	if len(fieldErrors) > 0 {
		WriteJSONResponse(w, http.StatusUnprocessableEntity, false, fieldErrors, ErrMsgValidation)
		return
	}
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, false, nil, ErrMsgSaveData)
		return
	}

	WriteJSONResponse(w, http.StatusOK, true, nil, ErrMsgDataOk)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// GetAllEntityHelper fetches and returns paginated data for a resource.
func GetAllEntityHelper[T any](
	w http.ResponseWriter,
	r *http.Request,
	service service.Service[T],
	tableFields []domain.Fields,
	entityContentTitle, entityTableID, entityURLPrefix, entityURLGetall, idField string,
	hasUpdateDelete ...bool,
) *domain.TableData {
	searchValue := r.URL.Query().Get("query")

	totRecords, err := service.GetTotalRecords(tableFields, searchValue)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, false, nil, ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(r, totRecords)
	allEntities, err := service.GetAll(pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, false, nil, ErrMsgReadData)
		return nil
	}

	table := common.SetTableBasicData(
		entityContentTitle, entityTableID, tableFields,
		entityURLPrefix, entityURLGetall,
		pageSize, currentPage, totalPages, totRecords,
	)
	common.SetTableRows(&table, *allEntities, tableFields, idField, entityURLPrefix, service.GetFieldCache())
	return &table
}

// GetEntityHelper fetches and returns a single resource by ID.
func GetEntityHelper[T any](
	w http.ResponseWriter,
	r *http.Request,
	service service.Service[T],
	tableFields []domain.Fields,
	idField string,
) {
	id, err := GetIDFromRequest(r, "id")
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	entity, err := service.GetByID(idField, id)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, false, nil, ErrMsgReadData)
		return
	}

	json.NewEncoder(w).Encode(entity)
}
