package common

import (
	"context"
	"database/sql"
	"fmt"
	"helia/config"
	"helia/i18n"
	"helia/internal/domain"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var MonthComboItems = []domain.ComboItem{
	{Key: "1", Value: "Januar"},
	{Key: "2", Value: "Februar"},
	{Key: "3", Value: "Mart"},
	{Key: "4", Value: "April"},
	{Key: "5", Value: "Maj"},
	{Key: "6", Value: "Jun"},
	{Key: "7", Value: "Jul"},
	{Key: "8", Value: "Avgust"},
	{Key: "9", Value: "Septembar"},
	{Key: "10", Value: "Oktobar"},
	{Key: "11", Value: "Novembar"},
	{Key: "12", Value: "Decembar"},
}

// GetPaginationData calculates pagination details (totalPages, etc.).
// This function can be used by both handler (initial page load) and service (query construction).
func GetPaginationData(c *gin.Context, totalRecords int, cfg config.Config) (currentPage, pageSize, totalPages int) {
	// Default values
	pageSize = cfg.PageSize
	currentPage = 1

	if c != nil { // Allow calling without request for service side
		if psStr := c.Query("pageSize"); psStr != "" {
			if ps, err := strconv.Atoi(psStr); err == nil && ps > 0 {
				pageSize = ps
			}
		}
		if pageStr := c.Query("page"); pageStr != "" {
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
func GetPageAndPageSizeFromRequest(c *gin.Context, cfg config.Config) (page, pageSize int) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")

	page = 1 // Default
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
		pageSize = ps
	}
	if pageSize == 0 {
		pageSize = cfg.PageSize // Default
	}
	return page, pageSize
}

// SetTableBasicData initializes a domain.TableData struct with common values.
// This is used by the service to build the TableData, and can be customized with options.
func SetTableBasicData(title, tableID string, headers []domain.Fields, urlPrefix, URLGetAll string, pageSize, currentPage, totalPages, totalRecords int, cfg config.Config, opts ...func(*domain.TableData)) domain.TableData {
	if pageSize == 0 {
		pageSize = cfg.PageSize
	}
	translator := i18n.GetInstance()
	table := domain.TableData{
		ContentTitle: translator.Title(title),
		TableID:      tableID,
		Headers:      headers,
		URLPrefix:    urlPrefix,
		URLGetAll:    URLGetAll,
		Pagination: domain.PaginationData{
			PageSize:     pageSize,
			CurrentPage:  currentPage,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
			PageSizes:    cfg.PageSizes,
			StartRecord:  currentPage*pageSize - pageSize + 1,
			EndRecord:    currentPage * pageSize,
		},
		ShowActions: true,                                                                                                         // Default, can be overridden
		BtnAdd:      domain.Button{LabelText: translator.Button("Dodaj"), BtnClass: ClassAddButton, IsVisible: true},              // Default, can be overridden
		BtnUpdate:   domain.Button{LabelText: "Izmeni", BtnClass: ClassConfirmButton, IdDialog: "dialog-update", IsVisible: true}, // Default, can be overridden
		BtnDelete:   domain.Button{LabelText: "Obriši", BtnClass: ClassDeleteButton, IdDialog: "dialog-delete", IsVisible: true},  // Default, can be overridden
		BtnPrint:    domain.Button{LabelText: translator.Button("Stampaj"), BtnClass: ClassPrintButton, IsVisible: true},          // Default, can be overridden
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
		// Create table row
		row := domain.TableRow{ID: id, Fields: fields, HasUpdate: true, HasDelete: true}
		table.Rows = append(table.Rows, row)
	}
	table = SetTableButtons(table, entityURLPrefix) // Set buttons for the table
	return table, nil
}

func SetTableButtons(table *domain.TableData, entityURLPrefix string) *domain.TableData {
	table.BtnAdd.IsVisible = true                                                   // Show Add button in the table header
	table.BtnAdd.HxActionURL = fmt.Sprintf("%s/confirm-add", entityURLPrefix)       // Set the URL for Add button
	table.BtnUpdate.HxRequestType = "GET"                                           // Set Update button to use GET request
	table.BtnUpdate.IsVisible = true                                                // Show Update button in the table header
	table.BtnUpdate.HxActionURL = fmt.Sprintf("%s/confirm-update", entityURLPrefix) // Set the URL for Update button
	table.BtnDelete.HxRequestType = "GET"                                           // Set Delete button to use DELETE request
	table.BtnDelete.IsVisible = true                                                // Show Delete button in the table header
	table.BtnDelete.HxActionURL = fmt.Sprintf("%s/confirm-delete", entityURLPrefix) // Set the URL for Delete button
	table.BtnPrint.IsVisible = true                                                 // Show Print button in the table header
	table.BtnPrint.HxActionURL = fmt.Sprintf("%s/print", entityURLPrefix)           // Show Print button in the table header
	return table
}

func SetButton(Id, LabelText, Icon, HxActionURL, HxTarget, HxSwap, HxRequestType, HxInclude, HxVals string, IsVisible bool, class string, afterRequest string) domain.Button {
	return domain.Button{
		Id:               Id,
		LabelText:        LabelText,
		Icon:             Icon,
		HxActionURL:      HxActionURL,
		HxTarget:         HxTarget,
		HxSwap:           HxSwap,
		HxVals:           HxVals,
		HxRequestType:    HxRequestType,
		HxInclude:        HxInclude,
		IsVisible:        IsVisible,
		BtnClass:         class,
		HxOnAfterRequest: afterRequest,
	}
}
func SetPrintButton(Id, LabelText, Icon, HxActionURL, HxRequestType string, IsVisible bool, class string, dataFields string) domain.Button {
	return domain.Button{
		Id:            Id,
		LabelText:     LabelText,
		Icon:          Icon,
		HxActionURL:   HxActionURL,
		HxRequestType: HxRequestType,
		IsVisible:     IsVisible,
		BtnClass:      class,
		DataFields:    dataFields,
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
			return t.Time.Format(DateLayout)
		}

		if fieldInfo.Type == reflect.TypeOf(time.Time{}) {
			t, ok := fieldValue.Interface().(time.Time)
			if !ok {
				return fmt.Sprintf("%v", fieldValue.Interface())
			}
			return t.Format(DateLayout)
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
func FormatFloatNumber64WithSystemLocale(val interface{}, precision int) float64 {
	switch v := val.(type) {
	case float64:
		return float64(v)
	}
	return val.(float64)
}

// FormatNullTime concisely in a helper:
func FormatNullTime(nt sql.NullTime, layout string) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format(layout)
}

// AddDaysToNullTime  helper function:
func AddDaysToNullTime(nt sql.NullTime, days int, layout string) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.AddDate(0, 0, days).Format(layout)
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

//***************************************

// GetTranslatedSubMenus returns submenus with translated names
func GetTranslatedSubMenus(menuData domain.MenuDataItems, menuName string, subMenus []domain.SubMenuItem, lang string) []domain.SubMenuItem {

	return translateSubMenus(subMenus, menuName, lang)
}

// translateSubMenus translates submenu names
func translateSubMenus(subMenus []domain.SubMenuItem, menuName, lang string) []domain.SubMenuItem {
	translated := make([]domain.SubMenuItem, len(subMenus))
	for i, subMenu := range subMenus {
		translated[i] = domain.SubMenuItem{
			SubMenuName: getSubmenuKey(menuName, subMenu.SubMenuName),
			Url:         subMenu.Url,
			Icon:        subMenu.Icon,
		}
	}
	return translated
}

// getSubmenuKey generates the translation key for a submenu item
func getSubmenuKey(menuName, submenuName string) string {
	// Convert submenu name to key format (same as in JSON files)
	key := "menu." + menuName + ".submenu." + submenuName
	itemName := i18n.GetInstance().T(key)
	return itemName
}

// **********************************************
// WriteJSONResponse writes a JSON response with the given status, success, errors, and message.
func WriteJSONResponse(
	c *gin.Context,
	status int,
	success bool,
	errors []domain.FieldError,
	message string,
) {
	c.JSON(status, domain.Response{
		StatusCode: status,
		Success:    success,
		Errors:     errors,
		Message:    message,
	})
}

// GetIconSVG returns the SVG path for a given icon name
func GetIconSVG(iconName string) string {
	if svg, ok := IconSVG[iconName]; ok {
		return svg
	}
	// Return a default icon if not found
	return `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />`
}

func ValidateRequiredParams(c *gin.Context, paramNames []string) []domain.FieldError {
	fieldsError := []domain.FieldError{}

	for _, paramName := range paramNames {
		if c.Query(paramName) == "" {
			fieldsError = append(fieldsError, domain.FieldError{
				Field:        paramName,
				ErrorMessage: ErrMsgObavezanPodatak,
			})
		}
	}

	return fieldsError
}

// GetCsrfToken retrieves the CSRF token from the Gin context
func GetCsrfToken(c *gin.Context) string {
	token, exists := c.Get("csrf_token")
	if !exists {
		return ""
	}
	if tokenStr, ok := token.(string); ok {
		return tokenStr
	}
	return ""
}

// GetCsrfTokenFromSession retrieves the CSRF token from session first, then fallback to context
func GetCsrfTokenFromSession(c *gin.Context) string {
	session := sessions.Default(c)
	csrfToken := ""
	if token := session.Get("csrf_token"); token != nil {
		if tokenStr, ok := token.(string); ok {
			csrfToken = tokenStr
		}
	}
	if csrfToken == "" {
		csrfToken = GetCsrfToken(c)
	}
	return csrfToken
}

func ExtractInnerXML(s, startTag, endTag string) string {
	start := strings.Index(s, startTag)
	end := strings.Index(s, endTag)
	if start != -1 && end != -1 && end > start {
		return s[start+len(startTag) : end]
	}
	return ""
}

func ExtractTag(xmlStr, tag string) string {
	openTag := fmt.Sprintf("<%s>", tag)
	closeTag := fmt.Sprintf("</%s>", tag)
	start := strings.Index(xmlStr, openTag)
	end := strings.Index(xmlStr, closeTag)
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return strings.TrimSpace(xmlStr[start+len(openTag) : end])
}

func GetMontshName() []string {
	translator := i18n.GetInstance()
	months := []string{
		translator.Label("januar"),
		translator.Label("februar"),
		translator.Label("mart"),
		translator.Label("april"),
		translator.Label("maj"),
		translator.Label("jun"),
		translator.Label("jul"),
		translator.Label("avgust"),
		translator.Label("septembar"),
		translator.Label("oktobar"),
		translator.Label("novembar"),
		translator.Label("decembar"),
	}
	return months
}

// SetTableConfig configures common table properties
func SetTableConfig(tbl *domain.TableData, contentTitle, urlPrefix string, showActions, showAdd, showPrint bool) {
	tbl.ShowActions = showActions
	tbl.ContentTitle = contentTitle
	tbl.URLPrefix = urlPrefix
	tbl.URLGetAll = urlPrefix
	tbl.BtnAdd.IsVisible = showAdd
	tbl.BtnPrint.IsVisible = showPrint
}

// CreateSearchInput creates a search input control with common configuration
func CreateSearchInput(id string, translator *i18n.Service, hxActionURL, hxTarget, hxVals string) domain.InputControl {
	return domain.InputControl{
		ID:           id,
		Label:        translator.Label("Pretrazi"),
		Type:         "search",
		Placeholder:  "Unesite tekst za pretragu",
		HxActionURL:  hxActionURL,
		HxTarget:     hxTarget,
		HxSwap:       "innerHTML",
		HxTrigger:    "keyup changed delay:500ms",
		Autocomplete: "off",
		Class:        ClassSearchInput,
		HxVals:       hxVals,
	}
}

// SetupTablePagination configures common table pagination and display settings
// This function sets up search, pagination, actions visibility, and button visibility
func SetupTablePagination(tbl *domain.TableData, currentPage, pageSize int) {
	tbl.ShowPagination = true
	tbl.Pagination.CurrentPage = currentPage
	tbl.Pagination.StartRecord = (currentPage-1)*pageSize + 1
	tbl.Pagination.EndRecord = tbl.Pagination.StartRecord + pageSize - 1
	tbl.Pagination.PageSize = pageSize
	if tbl.Pagination.EndRecord > tbl.Pagination.TotalRecords && tbl.Pagination.TotalRecords > 0 {
		tbl.Pagination.EndRecord = tbl.Pagination.TotalRecords
	}
	if tbl.Pagination.StartRecord > tbl.Pagination.TotalRecords && tbl.Pagination.TotalRecords > 0 {
		tbl.Pagination.StartRecord = tbl.Pagination.TotalRecords
	}
}

// SetTableTotalRecords sets the total records and calculates pagination values
// Returns true if the operation was to get total records only (no data processing needed)
func SetTableTotalRecords(tbl *domain.TableData, totalRecords, pageSize int) {
	tbl.ShowPagination = true
	tbl.Pagination.TotalRecords = totalRecords
	tbl.Pagination.StartRecord = (tbl.Pagination.CurrentPage-1)*pageSize + 1
	tbl.Pagination.EndRecord = tbl.Pagination.StartRecord + pageSize - 1
	tbl.Pagination.TotalPages = (totalRecords + pageSize - 1) / pageSize
	if tbl.Pagination.EndRecord > totalRecords {
		tbl.Pagination.EndRecord = totalRecords
	}
	if tbl.Pagination.StartRecord > totalRecords {
		tbl.Pagination.StartRecord = totalRecords
	}
}

func SetActiveTab(tabs *domain.TabData, index int) {
	for i := range tabs.Tabs {
		if i == index {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
}

// BilansCharAt returns the rune at position i in s as a string, or " " if out of range.
func BilansCharAt(s string, i int) string {
	r := []rune(s)
	if i < len(r) {
		return string(r[i])
	}
	return " "
}

func SetUnlockButtonProperties(btn *domain.Button, url string) {
	btn.HxActionURL = url
	btn.HxInclude = "input[name='_csrf']"
	btn.HxRequestType = "POST"
	btn.HxOnAfterRequest = "closeDialog"
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// FvrRepository is the minimal interface needed to fetch FVR (company) data.
type FvrRepository interface {
	GetHasGodHasKar() (bool, bool)
	GetAllCustom(ctx context.Context, queryText, whereText string, args []interface{}, limitOffset, sortBy string) (*[]domain.Fvr, error)
}

// GetFvrData retrieves company (FVR) data filtered by the current user session (god, kar, firma).
func GetFvrData(ctx context.Context, repo FvrRepository) (domain.Fvr, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.Fvr{}, fmt.Errorf("user session not found")
	}
	qb := NewQueryBuilder(`SELECT naziv, adresa, pobro, mesto, pib, matbr, sifdel FROM fvr`, true)
	hasGod, hasKar := repo.GetHasGodHasKar()
	if hasGod {
		qb.AddEqual("god", session.SelectedGod)
	}
	if hasKar {
		qb.AddEqual("kar", session.SelectedKar)
	}
	qb.AddEqual("fvr.naziv", session.Firma)
	sqlQuery, args := qb.Build()
	entities, err := repo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return domain.Fvr{}, err
	}
	if len(*entities) > 0 {
		return (*entities)[0], nil
	}
	return domain.Fvr{}, nil
}
