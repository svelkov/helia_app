package utils

import (
	"errors"
	"fmt"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/schema"
)

// parseAndDecode parses the form and decodes it into the provided entity.
func parseAndDecode(r *http.Request, entity any) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("%s: %w", common.ErrMsgParseForm, err)
	}
	decoder := schema.NewDecoder()
	if err := decoder.Decode(entity, r.PostForm); err != nil {
		return fmt.Errorf("%s: %w", common.ErrMsgFormDecode, err)
	}
	return nil
}

// GetIDFromRequest extracts the ID from the request URL path.
func GetIDFromRequest(c *gin.Context, key string) (int64, error) {
	idStr := c.Param(key)
	if idStr == "" {
		return 0, errors.New(common.ErrMsgGetIDFromURL)
	}
	return strconv.ParseInt(idStr, 10, 64)
}

// DeleteHelper handles HTTP DELETE requests for a resource.
// It expects the resource ID in the URL path and uses the provided service to delete the resource.
func DeleteHelper[T any](
	c *gin.Context,
	service service.Service[T],
	idType string,
) {
	id, err := GetIDFromRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	if err := service.Delete(idType, id); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, fmt.Sprintf("%s: %v", common.ErrMsgDeleteData, err))
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

// ConfirmDeleteHelper renders a confirmation dialog for resource deletion.
func ConfirmDeleteHelper(
	c *gin.Context,
	tableFields []domain.Fields,
) {
	idStr := c.Query("id")
	url := c.Query("url")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	url = fmt.Sprintf("%s/%d", url, id)
	dialog := SetDialogValues("dialog-delete", url, "Brisanje podataka", "DELETE")
	btnConfirm := domain.Button{
		Id:            "btn-delete",
		LabelText:     "Obriši",
		IsVisible:     true,
		BtnClass:      common.ClassDeleteButton,
		HxActionURL:   url,
		HxRequestType: "DELETE",
		IdDialog:      "dialog-delete",
		HxTarget:      "#info-message",
		HxSwap:        "innerHTML",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		LabelText: "Odustani",
		IsVisible: true,
		IdDialog:  "dialog-delete",
		BtnClass:  common.ClassCloseButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  "dialog-delete",
		BtnClass:  common.ClassDialogCloseButton,
	}

	RenderDialogContent(c, dialog, tableFields, common.ActionDelete, btnConfirm, btnCancel, btnClose)
}

// ConfirmAddHelper renders a dialog for adding a new resource.
func ConfirmAddHelper(c *gin.Context, url string, tableFields []domain.Fields) {
	for i := range tableFields {
		tableFields[i].Value = ""
	}
	dialog := SetDialogValues("dialog-save", url, "Unos novih podataka", "POST")
	btnSave := domain.Button{
		Id:            "btn-save",
		LabelText:     "Sačuvaj",
		IsVisible:     true,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   url,
		HxRequestType: "POST",
		IdDialog:      "dialog-save",
		HxTarget:      "#info-message",
		HxSwap:        "innerHTML",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		LabelText: "Odustani",
		IsVisible: true,
		IdDialog:  "dialog-save",
		BtnClass:  common.ClassCloseButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  "dialog-save",
		BtnClass:  common.ClassDialogCloseButton,
	}
	RenderDialogContent(c, dialog, tableFields, common.ActionAdd, btnSave, btnCancel, btnClose)
}

// ConfirmUpdateHelper renders a dialog for updating a resource.
func ConfirmUpdateHelper[T any](c *gin.Context, service service.Service[T], tableFields []domain.Fields, idField string) {
	idStr := c.Query("id")
	url := c.Query("url")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgInvalidID)
		return
	}

	entity, err := service.GetByID(idField, int64(id))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgReadData)
		return
	}

	url = fmt.Sprintf("%s/%d", url, id)
	fields := service.MapEntityToValues(entity, tableFields)
	dialog := SetDialogValues("dialog-izmeni", url, "Izmena podataka", "PUT")
	btnSave := domain.Button{
		Id:               "btn-save",
		LabelText:        "Sačuvaj",
		IsVisible:        true,
		BtnClass:         common.ClassSaveButton,
		HxActionURL:      url,
		HxRequestType:    "PUT",
		IdDialog:         "dialog-izmeni",
		HxTarget:         "#info-message",
		HxOnAfterRequest: "handleDialogResponse('dialog-izmeni')",
		HxSwap:           "innerHTML",
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		LabelText: "Odustani",
		IsVisible: true,
		IdDialog:  "dialog-izmeni",
		BtnClass:  common.ClassCloseButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  "dialog-izmeni",
		BtnClass:  common.ClassDialogCloseButton,
	}
	RenderDialogContent(c, dialog, fields, common.ActionUpdate, btnSave, btnCancel, btnClose)
}

// CreateHelper handles HTTP POST requests for creating a new resource.
func CreateHelper[T any](
	c *gin.Context,
	entity *T,
	service service.Service[T],
	idField string,
	tableFields []domain.Fields,
) (insertedId int64, err error) {

	// Use Gin's JSON binding instead of custom parseAndDecode
	if err := c.ShouldBindJSON(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return 0, err
	}

	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, lastInsertedID, err := service.Create(entity, idField, tableFields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldErrors, common.ErrMsgSaveData)
		return 0, err
	}

	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return 0, nil
	}

	common.WriteJSONResponse(c, http.StatusCreated, true, nil, common.OkMsgSaveData)
	return lastInsertedID, nil
}

// UpdateHelper handles HTTP PUT requests for updating a resource.
func UpdateHelper[T any](
	c *gin.Context,
	entity *T,
	service service.Service[T],
	tableFields []domain.Fields,
	idField string,
) {
	id, err := GetIDFromRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}
	// Parse request body
	if err := c.ShouldBindJSON(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, err := service.Update(entity, idField, id, tableFields)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData)
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
	c.Redirect(http.StatusSeeOther, getRedirectURL(c))
}

// getRedirectURL calculates redirect URL by removing the last path segment
func getRedirectURL(c *gin.Context) string {
	path := c.Request.URL.Path
	lastSlashIndex := strings.LastIndex(path, "/")

	if lastSlashIndex > 0 {
		return path[:lastSlashIndex] + "/all"
	}

	return "/all"
}

// GetAllEntityHelper fetches and returns paginated data for a resource.
func GetAllEntityHelper[T any](
	c *gin.Context,
	service service.Service[T],
	tableFields []domain.Fields,
	entityContentTitle, entityTableID, entityURLPrefix, entityURLGetall, idField string,
	hasUpdateDelete ...bool,
) *domain.TableData {
	searchValue := c.DefaultQuery("query", "")

	totRecords, err := service.GetTotalRecords(tableFields, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(c, totRecords)
	allEntities, err := service.GetAll(pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
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
	c *gin.Context,
	service service.Service[T],
	tableFields []domain.Fields,
	idField string,
) {
	id, err := GetIDFromRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	entity, err := service.GetByID(idField, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	c.JSON(http.StatusOK, entity)
}
