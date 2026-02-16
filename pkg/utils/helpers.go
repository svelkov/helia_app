package utils

import (
	"errors"
	"fmt"
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/service"
	"net/http"
	"strconv"
	"strings"

	tmpl "helia/frontend/templates"

	"github.com/gin-gonic/gin"
)

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
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, fmt.Sprintf("%s: %s", common.ErrMsgDeleteData, err.Error()))
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

// ConfirmDeleteHelper renders a confirmation dialog for resource deletion.
func ConfirmDeleteHelper(
	c *gin.Context,
	tableFields []domain.Fields,
) {
	rowID := c.Query("id")
	url := c.Query("url")

	id, err := strconv.Atoi(rowID)
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

	RenderDialogContent(c, dialog, tableFields, common.ActionDelete, btnConfirm, btnCancel, btnClose, rowID)
}

// ConfirmAddHelper renders a dialog for adding a new resource.
func ConfirmAddHelper(c *gin.Context, url string, tableFields []domain.Fields) {
	for i := range tableFields {
		tableFields[i].Value = ""
	}
	rowID := ""
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
	RenderDialogContent(c, dialog, tableFields, common.ActionAdd, btnSave, btnCancel, btnClose, rowID)
}

// ConfirmUpdateHelper renders a dialog for updating a resource.
func ConfirmUpdateHelper[T any](c *gin.Context, service service.Service[T], tableFields []domain.Fields, idField string) {
	rowID := c.Query("id")
	url := c.Query("url")

	id, err := strconv.Atoi(rowID)
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
	RenderDialogContent(c, dialog, fields, common.ActionUpdate, btnSave, btnCancel, btnClose, rowID)
}

// CreateHelper handles HTTP POST requests for creating a new resource.
func CreateHelper[T any](
	c *gin.Context,
	entity *T,
	service service.Service[T],
	idField string,
	tableFields []domain.Fields,
) (fieldsError []domain.FieldError, err error) {
	fieldsError = []domain.FieldError{}
	// Use Gin's JSON binding instead of custom parseAndDecode
	if err := c.ShouldBind(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return nil, err
	}

	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, _, err := service.Create(c, entity, idField, tableFields)
	if err != nil || len(fieldErrors) > 0 {
		return fieldErrors, err
	}

	return nil, nil
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
	if err := c.ShouldBind(entity); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, err := service.Update(c, entity, idField, id, tableFields)
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
	cfg config.Config,
	hasUpdateDelete ...bool,
) *domain.TableData {
	searchValue := c.DefaultQuery("query", "")

	totRecords, err := service.GetTotalRecords(c, tableFields, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(c, totRecords, cfg)
	allEntities, err := service.GetAll(c, pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	table := common.SetTableBasicData(
		entityContentTitle, entityTableID, tableFields,
		entityURLPrefix, entityURLGetall,
		pageSize, currentPage, totalPages, totRecords,
		cfg,
	)
	common.SetTableRows(&table, *allEntities, tableFields, idField, entityURLPrefix, service.GetFieldCache())
	return &table
}

func GetAllPdfEntityHelper[T any](
	c *gin.Context,
	service service.Service[T],
	tableFields []domain.Fields,
	entityContentTitle, entityTableID, entityURLPrefix, entityURLGetall, idField string,
	cfg config.Config,
	hasUpdateDelete ...bool,
) *domain.TableData {
	searchValue := c.DefaultQuery("query", "")

	totRecords, err := service.GetTotalRecords(c, tableFields, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(c, totRecords, cfg)
	allEntities, err := service.GetAll(c, pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	table := common.SetTableBasicData(
		entityContentTitle, entityTableID, tableFields,
		entityURLPrefix, entityURLGetall,
		pageSize, currentPage, totalPages, totRecords, cfg,
	)
	common.SetTableRows(&table, *allEntities, tableFields, idField, entityURLPrefix, service.GetFieldCache())
	return &table
}
func GetAllExcelEntityHelper[T any](
	c *gin.Context,
	service service.Service[T],
	tableFields []domain.Fields,
	entityContentTitle, entityTableID, entityURLPrefix, entityURLGetall, idField string,
	cfg config.Config,
	hasUpdateDelete ...bool,
) *domain.TableData {
	searchValue := c.DefaultQuery("query", "")

	totRecords, err := service.GetTotalRecords(c, tableFields, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(c, totRecords, cfg)
	allEntities, err := service.GetAll(c, pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	table := common.SetTableBasicData(
		entityContentTitle, entityTableID, tableFields,
		entityURLPrefix, entityURLGetall,
		pageSize, currentPage, totalPages, totRecords, cfg,
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

func SearchButtonDialog(c *gin.Context) {
	hxVals := ""
	vkonta := c.Query("vkonta")
	queryParams := c.Request.URL.Query()
	id := ""
	fieldName := "konto" // default
	placeholder := "trazi konto..."

	// Detect which field is being searched by checking query parameters
	searchFields := []string{"search-konto", "search-odkonta", "search-dokonta", "search-odmi", "search-domi", "search-sifra", "search-odsifre", "search-dosifre"}
	for _, searchField := range searchFields {
		if _, exists := queryParams[searchField]; exists {
			// Extract field name from search-* parameter
			fieldName = strings.TrimPrefix(searchField, "search-")
			id = fieldName
			placeholder = fmt.Sprintf("trazi %s...", fieldName)
			break
		}
	}

	// Get the value for the detected field
	fieldValue := c.Query(fieldName)

	hxVals = fmt.Sprintf(`{"konto": "%s", "vkonta": "%s", "fieldName": "%s"}`, fieldValue, vkonta, fieldName)
	tmpl.SearchButtonDialog(id, id, placeholder, "/api/fkpl/trazikontosearchtable", "#search-results", "innerHTML", hxVals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)

}
