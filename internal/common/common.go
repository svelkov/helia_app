package common

import (
	"database/sql"
	"fmt"
	"helia/internal/domain"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetPaginationData calculates pagination details (totalPages, etc.).
// This function can be used by both handler (initial page load) and service (query construction).
func GetPaginationData(r *http.Request, totalRecords int) (currentPage, pageSize, totalPages int) {
	// Default values
	pageSize = 10
	currentPage = 1

	if r != nil { // Allow calling without request for service side
		if psStr := r.URL.Query().Get("pageSize"); psStr != "" {
			if ps, err := strconv.Atoi(psStr); err == nil && ps > 0 {
				pageSize = ps
			}
		}
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				currentPage = p
			}
		}
	}

	totalPages = (totalRecords + pageSize - 1) / pageSize
	if totalPages == 0 { // Handle case with 0 records
		totalPages = 1
	}

	// Ensure current page is within valid range
	if currentPage > totalPages {
		currentPage = totalPages
	}
	if currentPage < 1 {
		currentPage = 1
	}

	return currentPage, pageSize, totalPages
}

// GetPageAndPageSizeFromRequest extracts "page" and "pageSize" query parameters.
func GetPageAndPageSizeFromRequest(r *http.Request) (page, pageSize int) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page = 1 // Default
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	pageSize = 10 // Default
	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
		pageSize = ps
	}
	return page, pageSize
}

// SetTableBasicData initializes a domain.TableData struct with common values.
// This is used by the service to build the TableData, and can be customized with options.
func SetTableBasicData(title, tableID string, headers []domain.Fields, urlPrefix, URLGetAll string, pageSize, currentPage, totalPages, totalRecords int, opts ...func(*domain.TableData)) domain.TableData {
	table := domain.TableData{
		ContentTitle: title,
		TableID:      tableID,
		Headers:      headers,
		URLPrefix:    urlPrefix,
		URLGetAll:    URLGetAll,
		Pagination: domain.PaginationData{
			PageSize:     pageSize,
			CurrentPage:  currentPage,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
			PageSizes:    []int{10, 15, 25, 50, 100},
			StartRecord:  currentPage*pageSize - pageSize + 1,
			EndRecord:    currentPage * pageSize,
		},
		ShowActions: true,                                                 // Default, can be overridden
		BtnAdd:      domain.Button{LabelText: "Dodaj", IsVisible: true},   // Default, can be overridden
		BtnUpdate:   domain.Button{LabelText: "Izmeni", IsVisible: true},  // Default, can be overridden
		BtnDelete:   domain.Button{LabelText: "Obriši", IsVisible: true},  // Default, can be overridden
		BtnPrint:    domain.Button{LabelText: "Stampaj", IsVisible: true}, // Default, can be overridden
	}
	table.ShowPagination = totalRecords > 0 // Show pagination only if there are records
	for _, opt := range opts {
		opt(&table)
	}
	if table.Pagination.EndRecord > totalRecords {
		table.Pagination.EndRecord = totalRecords
	}
	if table.Pagination.StartRecord > totalRecords {
		table.Pagination.StartRecord = totalRecords
	}
	return table
}

func SetTableRows[T any](table *domain.TableData, entities []T, tableFields []domain.Fields, idField, entityURLPrefix string, fieldCache map[string]reflect.StructField) (*domain.TableData, error) {
	for _, entity := range entities {
		val := reflect.ValueOf(entity)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		// Get ID field value from cache
		idFld, found := fieldCache[strings.ToLower(idField)]
		if !found {
			return nil, fmt.Errorf("ID field %s not found in cache", idField)
		}
		// Get ID field value dynamically
		idValue := val.FieldByName(idFld.Name)
		id := fmt.Sprintf("%v", idValue.Interface())

		// Extract specified fields using cache
		fields := []string{}
		for _, field := range tableFields {
			fieldInfo, found := fieldCache[strings.ToLower(field.Name)]
			if !found {
				continue // or return error if field is required
			}

			value := GetFormattedValue(fieldInfo, val.FieldByName(fieldInfo.Name))
			fields = append(fields, value)
		}
		table = SetTableButtons(table, entityURLPrefix) // Set buttons for the table
		// Create table row
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: true, HasDelete: true}
		table.Rows = append(table.Rows, row)
	}
	return table, nil
}

func SetTableButtons(table *domain.TableData, entityURLPrefix string) *domain.TableData {
	table.BtnAdd.IsVisible = true                                                  // Show Add button in the table header
	table.BtnAdd.HxActionURL = fmt.Sprintf("%s/confirm-add", entityURLPrefix)       // Set the URL for Add button
	table.BtnUpdate.IsVisible = true                                               // Show Update button in the table header
	table.BtnUpdate.HxActionURL = fmt.Sprintf("%s/confirm-update", entityURLPrefix) // Set the URL for Update button
	table.BtnDelete.IsVisible = true                                               // Show Delete button in the table header
	table.BtnDelete.HxActionURL = fmt.Sprintf("%s/confirm-delete", entityURLPrefix) // Set the URL for Delete button
	table.BtnPrint.IsVisible = true                                                // Show Print button in the table header
	table.BtnPrint.HxActionURL = fmt.Sprintf("%s/report", entityURLPrefix)          // Show Print button in the table header
	return table
}

func SetButton(Id, LabelText, Icon, HxActionURL, HxTarget, HxSwap, HxRequestType, HxInclude, HxVals string, IsVisible bool) domain.Button {
	return domain.Button{
		Id:            Id,
		LabelText:     LabelText,
		Icon:          Icon,
		HxActionURL:   HxActionURL,
		HxTarget:      HxTarget,
		HxSwap:        HxSwap,
		HxVals:        HxVals,
		HxRequestType: HxRequestType,
		HxInclude:     HxInclude,
		IsVisible:     IsVisible,
	}
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
func GetFormattedValue(fieldInfo reflect.StructField, fieldValue reflect.Value) string {
	// Handle special types based on cached field type
	//fmt.Println("Field Type:", fieldInfo.Type.Kind(), "Value:", fieldValue.Interface())
	switch fieldInfo.Type.Kind() {
	case reflect.Struct:
		// Special handling for time.Time
		if fieldInfo.Type == reflect.TypeOf(sql.NullTime{}) {
			fmt.Println(fieldInfo.Type)
			t, ok := fieldValue.Interface().(sql.NullTime)
			if !ok {
				return fmt.Sprintf("%v", fieldValue.Interface())
			}
			return t.Time.Format("02.01.2006")
		}

		if fieldInfo.Type == reflect.TypeOf(time.Time{}) {
			t, ok := fieldValue.Interface().(time.Time)
			if !ok {
				return fmt.Sprintf("%v", fieldValue.Interface())
			}
			return t.Format("02.01.2006")
		}

	case reflect.Float32, reflect.Float64:
		// Format numbers with 2 decimal places
		return FormatNumberWithSystemLocale(fieldValue.Float(), 2)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return FormatNumberWithSystemLocale(fieldValue.Int(), 0)

	case reflect.Bool:
		if fieldValue.Bool() {
			return "true" // Or "Yes" for English
		}
		return "false" // Or "No" for English

	case reflect.String:
		// Handle empty strings
		str := fieldValue.String()
		if str == "" {
			return ""
		}
		return str

	case reflect.Ptr:
		// Handle nil pointers
		if fieldValue.IsNil() {
			return ""
		}
		// Create a new StructField for the dereferenced type
		derefFieldInfo := reflect.StructField{
			Type: fieldInfo.Type.Elem(), // Get the type the pointer points to
		}
		return GetFormattedValue(derefFieldInfo, fieldValue.Elem())
	}

	// Default formatting for other types
	return fmt.Sprintf("%v", fieldValue.Interface())
}

func FormatNumberWithSystemLocale(val interface{}, precision int) string {
	systemLang := detectSystemLanguage()
	p := message.NewPrinter(systemLang)

	switch v := val.(type) {
	case float64:
		return p.Sprintf("%.*f", precision, v)
	case float32:
		return p.Sprintf("%.*f", precision, float64(v))
	case int:
		return p.Sprintf("%d", v)
	default:
		return p.Sprintf("%v", val)
	}
}

// detectSystemLanguage reads $LANG/$LC_ALL and returns a language.Tag.
func detectSystemLanguage() language.Tag {
	// Priority: LC_ALL > LANG > Default (English)
	langEnv := os.Getenv("LC_ALL")
	if langEnv == "" {
		langEnv = os.Getenv("LANG")
	}

	if langEnv != "" {
		// Parse language tag (e.g., "en_US.UTF-8" → "en-US")
		tag, err := language.Parse(langEnv)
		if err == nil {
			return tag
		}
	}

	return language.English // Fallback
}

func StringToFloat64(str string, defaultValue float64) float64 {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultValue
	}
	return val
}
func StringToInt(str string, defaultValue int) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return val
}
func StringToInt64(str string, defaultValue int64) int64 {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultValue
	}
	return val
}
func StringToBool(str string, defaultValue bool) bool {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return defaultValue
	}
	return val
}

func GetSubMenus(menuData domain.MenuDataItems, targetMenu string) []domain.SubMenuItem {
	for _, menuItem := range menuData.MenuItems {
		if menuItem.Name == targetMenu {
			return menuItem.SubMenus
		}
	}
	return nil // or empty slice if menu not found
}
