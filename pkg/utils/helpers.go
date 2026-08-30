package utils

import (
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

// DeleteHelper handles HTTP DELETE requests for a resource.
// It expects the resource ID in the URL path and uses the provided service to delete the resource.
func DeleteHelper[T any](
	c *gin.Context,
	service service.Service[T],
	idType string,
) {
	id, err := GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	if err := service.Delete(c.Request.Context(), idType, id); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, fmt.Sprintf("%s: %s", common.ErrMsgDeleteData, err.Error()))
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgDeleteData)
}

// ConfirmDeleteHelper renders a confirmation dialog for resource deletion.
func ConfirmDeleteHelper(c *gin.Context, tableFields []domain.Fields, hxTarget string) {
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
		Id:                   "btn-delete",
		LabelText:            "Obriši",
		IsVisible:            true,
		BtnClass:             common.ClassDeleteButton,
		HxActionURL:          url,
		HxRequestType:        "DELETE",
		IdDialog:             "dialog-delete",
		HxTarget:             hxTarget,
		HxSwap:               "innerHTML",
		HxOnAfterRequest:     "handleDialogResponse",
		HxOnAfterRequestArgs: []any{"dialog-delete"},
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
func ConfirmAddHelper(c *gin.Context, url string, tableFields []domain.Fields, hxTarget string) {
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
		HxTarget:      hxTarget,
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

	entity, err := service.GetByID(c.Request.Context(), idField, int64(id))
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
	ctx := c.Request.Context()
	tableFields = service.MapEntityToValues(entity, tableFields)
	fieldErrors, _, err := service.Create(ctx, entity, idField, tableFields)
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
	id, err := GetInt64FromParameterRequest(c, "id")
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
	fieldErrors, err := service.Update(c.Request.Context(), entity, idField, id, tableFields)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData)
		return
	}

	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
	c.Redirect(http.StatusSeeOther, GetRedirectURL(c))
}

// GetRedirectURL calculates redirect URL by removing the last path segment
func GetRedirectURL(c *gin.Context) string {
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
	searchText := c.DefaultQuery("query", "")
	sortBy := c.DefaultQuery("sortBy", "")
	sortOrder := c.DefaultQuery("sortOrder", "")
	ctx := c.Request.Context()
	totRecords, err := service.GetTotalRecords(ctx, tableFields, searchText)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(c, totRecords, cfg)
	allEntities, err := service.GetAll(ctx, pageSize, (currentPage-1)*pageSize, tableFields, idField, searchText, sortBy, sortOrder)
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
	table.BtnPrint.DataFields = "sortBy,sortOrder"
	table.BtnExportPDF.IsVisible = true
	table.BtnExportExcel.IsVisible = true
	common.SetTableRows(&table, *allEntities, tableFields, idField, entityURLPrefix, service.GetFieldCache())
	return &table
}

// GetAllPrintEntityHelper fetches and returns paginated data for a resource.
func GetAllPrintEntityHelper[T any](
	c *gin.Context,
	service service.Service[T],
	tableFields []domain.Fields,
	entityContentTitle, entityTableID, entityURLPrefix, entityURLGetall, idField string,
	cfg config.Config,
) *domain.TableData {
	sortBy := c.DefaultQuery("sortBy", "")
	sortOrder := c.DefaultQuery("sortOrder", "")
	ctx := c.Request.Context()
	allEntities, err := service.GetAll(ctx, 0, 0, tableFields, idField, "", sortBy, sortOrder)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	tbl := common.SetTableBasicData(
		entityContentTitle, entityTableID, tableFields,
		entityURLPrefix, entityURLGetall,
		0, 0, 0, 0,
		cfg,
	)
	tbl.ShowPagination = false
	common.SetTableRows(&tbl, *allEntities, tableFields, idField, entityURLPrefix, service.GetFieldCache())
	return &tbl
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
	sortBy := c.DefaultQuery("sortBy", "")
	sortOrder := c.DefaultQuery("sortOrder", "")

	ctx := c.Request.Context()
	totRecords, err := service.GetTotalRecords(ctx, tableFields, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(c, totRecords, cfg)
	allEntities, err := service.GetAll(ctx, pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue, sortBy, sortOrder)
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
	sortBy := c.DefaultQuery("sortBy", "")
	sortOrder := c.DefaultQuery("sortOrder", "")
	ctx := c.Request.Context()
	totRecords, err := service.GetTotalRecords(ctx, tableFields, searchValue)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return nil
	}

	currentPage, pageSize, totalPages := common.GetPaginationData(c, totRecords, cfg)
	allEntities, err := service.GetAll(ctx, pageSize, (currentPage-1)*pageSize, tableFields, idField, searchValue, sortBy, sortOrder)
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
	id, err := GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	ctx := c.Request.Context()
	entity, err := service.GetByID(ctx, idField, id)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	c.JSON(http.StatusOK, entity)
}

func SearchButtonDialog(c *gin.Context) {
	hxVals := ""
	konto := c.Query("konto")
	vkonta := c.Query("vkonta")
	queryParams := c.Request.URL.Query()
	id := ""
	placeholder := "trazi podatak..."

	fieldName := queryParams.Get("destfield")

	hxVals = fmt.Sprintf(`{"konto": "%s", "vkonta": "%s", "fieldName": "%s"}`, konto, vkonta, fieldName)
	tmpl.SearchButtonDialog(id, id, placeholder, "/api/fkpl/trazikontosearchtable", "#search-results", "innerHTML", hxVals, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)

}

// PaginateTableData splits table rows into pages for print-friendly reports
// rowsPerPage: number of rows per page (typically 20-30)
// Returns a slice of TableData objects, one for each page
func PaginateTableData(tableData *domain.TableData, rowsPerPage int) []*domain.TableData {
	if rowsPerPage <= 0 {
		rowsPerPage = 25
	}

	totalRows := len(tableData.Rows)
	if totalRows == 0 {
		return []*domain.TableData{tableData}
	}

	// Calculate number of pages needed
	totalPages := (totalRows + rowsPerPage - 1) / rowsPerPage

	// Create a slice to hold paginated tables
	paginatedTables := make([]*domain.TableData, 0, totalPages)

	// Split rows into pages
	for page := 0; page < totalPages; page++ {
		startIdx := page * rowsPerPage
		endIdx := startIdx + rowsPerPage
		if endIdx > totalRows {
			endIdx = totalRows
		}

		// Create a copy of the table data for this page
		pageTable := &domain.TableData{
			ContentTitle:        tableData.ContentTitle,
			TableID:             tableData.TableID,
			Headers:             tableData.Headers,
			Rows:                tableData.Rows[startIdx:endIdx],
			Pagination:          tableData.Pagination,
			URLPrefix:           tableData.URLPrefix,
			URLGetAll:           tableData.URLGetAll,
			HxInclude:           tableData.HxInclude,
			HxTarget:            tableData.HxTarget,
			BtnAdd:              tableData.BtnAdd,
			BtnUpdate:           tableData.BtnUpdate,
			BtnDelete:           tableData.BtnDelete,
			BtnPrint:            tableData.BtnPrint,
			BtnExportExcel:      tableData.BtnExportExcel,
			BtnExportPDF:        tableData.BtnExportPDF,
			SearchEnabled:       tableData.SearchEnabled,
			ShowActions:         tableData.ShowActions,
			ShowPagination:      false,
			HxVals:              tableData.HxVals,
			DetailTarget:        tableData.DetailTarget,
			DetailURL:           tableData.DetailURL,
			DetailHxRequestType: tableData.DetailHxRequestType,
			DetailHxTrigger:     tableData.DetailHxTrigger,
			DetailHxSwap:        tableData.DetailHxSwap,
			DetailHxHeaders:     tableData.DetailHxHeaders,
			ExportFilename:      tableData.ExportFilename,
			HasExportExcel:      tableData.HasExportExcel,
			HasExportPdf:        tableData.HasExportPdf,
			FuncClick:           tableData.FuncClick,
			FuncDblClick:        tableData.FuncDblClick,
			DestField:           tableData.DestField,
			Totals:              tableData.Totals,
			HasTotals:           tableData.HasTotals && page == totalPages-1, // Show totals only on last page
			TotalsCalculated:    tableData.TotalsCalculated,
		}

		paginatedTables = append(paginatedTables, pageTable)
	}

	return paginatedTables
}
