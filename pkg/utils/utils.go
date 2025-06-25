// Shared utilities (e.g., date parsing)
package utils

import (
	"encoding/json"
	"fmt"
	"helia/internal/domain"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	tmpl "helia/frontend/templates"
	tmpl_opsti "helia/frontend/templates/opstipodaci"

	"github.com/a-h/templ"
)

var God = 2024
var Kar = 1

const DefaultPageSize = 10

func SetGodKar(god, kar int) {
	God = god
	Kar = kar
}

// add returns the sum of two integers
func Add(a, b int) int {
	return a + b
}

// sub returns the difference between two integers
func Sub(a, b int) int {
	return a - b
}

func Le(a, b int) bool {
	return a <= b
}

func Ge(a, b int) bool {
	return a >= b
}

func Lt(a, b int) bool {
	return a < b
}

func Gt(a, b int) bool {
	return a > b
}

// seq generates a slice of integers from 1 to n
func Seq(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i + 1
	}
	return result
}

// Create funcmap
func CreateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"seq": func(n int) []int {
			seq := make([]int, n)
			for i := 0; i < n; i++ {
				seq[i] = i + 1
			}
			return seq
		},
		"gt":  func(a, b int) bool { return a > b },
		"lt":  func(a, b int) bool { return a < b },
		"ge":  func(a, b int) bool { return a >= b },
		"le":  func(a, b int) bool { return a <= b },
		"con": func(a, b string) bool { return strings.Contains(a, b) },
	}
}

func RenderContent(w http.ResponseWriter, r *http.Request, table domain.TableData, tmplName ...string) {
	reqURI := r.URL.RequestURI()
	templateName := ""

	if len(tmplName) == 0 {
		if !strings.Contains(reqURI, "?") {
			templateName = "ContentContainer"
		} else {
			templateName = "Table"
		}
	} else {
		templateName = tmplName[0]
	}

	w.Header().Set("Content-Type", "text/html")

	var err error
	switch templateName {
	case "Table":
		err = tmpl.Table(table).Render(r.Context(), w)
	case "ContentContainer":
		searchControl := domain.InputControl{
			ID:           "search-control",
			Label:        "Pretraži",
			Type:         "search",
			Placeholder:  "Unesite tekst za pretragu",
			HxActionURL:  table.URLGetAll,
			HxTarget:     "#table-body",
			HxSwap:       "innerHTML",
			HxInclude:    "#tipdokSelect",
			Autocomplete: "off",
		}
		err = tmpl.ContentContainer(table, searchControl).Render(r.Context(), w)
	default:
		http.Error(w, fmt.Sprintf("Template '%s' not found", templateName), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func RenderDialogContent(w http.ResponseWriter, r *http.Request, dialog domain.Dialog, fields []domain.Fields, actionType string) {
	if actionType == "DELETE" {
		component := tmpl.DeleteDialog(dialog) //use the component
		component.Render(r.Context(), w)
	}
	if actionType == "ADD" || actionType == "UPDATE" {
		component := tmpl_opsti.AddUpdateForm(dialog, fields) //use the component
		component.Render(r.Context(), w)
	}
}

func SetDialogValues(id string, actionURL, title, requestType string) domain.Dialog {
	return domain.Dialog{
		Id:            id,
		Title:         title,
		OkText:        "Potvrdi",
		CancelText:    "Odustani",
		SaveText:      "Sacuvaj",
		HxActionURL:   actionURL,
		HxTarget:      "#info-message",
		HxSwap:        "innerHTML",
		HxRequestType: requestType,
	}
}

// Helper function to extract HTML from a rendered templ.Component.
func ExtractHTML(component string, targetID string, tagType string) string {
	startTag := fmt.Sprintf(`<%s id="%s">`, tagType, targetID) // Adjust if using a different tag
	endTag := fmt.Sprintf("</%s>", tagType)                    // Adjust if using a different tag

	startIndex := -1
	endIndex := -1

	startIndex = strings.Index(component, startTag)
	if startIndex != -1 {
		startIndex += len(startTag)
		endIndex = strings.Index(component[startIndex:], endTag)
		if endIndex != -1 {
			endIndex += startIndex
			return component[startIndex:endIndex]
		}
	}
	return "" // Or handle the error as you see fit.
}

// GetIDFromRequest extracts ID from URL path parameters.
func GetIDFromRequest(r *http.Request, idParamName string) (int, error) {
	idStr := r.PathValue(idParamName) // Assuming chi router for URL params
	if idStr == "" {
		return 0, fmt.Errorf("ID parameter '%s' is missing in URL", idParamName)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format for '%s': %w", idParamName, err)
	}
	return id, nil
}

// CreateResponse creates a standardized JSON response.
func CreateResponse(w http.ResponseWriter, success bool, errors []domain.FieldError, message string, statusCode int) domain.Response {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	return domain.Response{
		Success:   success,
		Errors:    errors,
		Message:   message,
		HxTrigger: "showMessage",
	}
}
func SendResponse(w http.ResponseWriter, statusCode int, response domain.Response) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)

}

// Helper function to dynamically get the template rendering function by name
func getTemplateFunc(tmpl *domain.MyTemplates, tmplName string) (func(domain.TableData) templ.Component, bool) {
	switch tmplName {
	case "Table":
		return tmpl.Table, true
	case "ContentContainer":
		return tmpl.ContentContainer, true
	case "Nalozi":
		return tmpl.Nalozi, true
		// Add more cases for your other template names
	default:
		return nil, false
	}
}

// WithHxInclude is an option for SetTableBasicData to set HxInclude.
func WithHxInclude(hxInclude string) func(*domain.TableData) {
	return func(t *domain.TableData) {
		t.HxInclude = hxInclude
	}
}

// WithPagination is an option for SetTableBasicData to enable/configure pagination.
func WithPagination(currentPage, pageSize, totalPages int) func(*domain.TableData) {
	return func(t *domain.TableData) {
		//t.ShowPagination = true
		t.Pagination.CurrentPage = currentPage
		t.Pagination.PageSize = pageSize
		t.Pagination.TotalPages = totalPages
	}
}

// WithUpdate is an option for SetTableBasicData to set show update actions.
func WithUpdate() func(*domain.TableData) {
	return func(t *domain.TableData) {
		t.BtnUpdate.IsVisible = true
	}
}

// WithDelete is an option for SetTableBasicData to set show delete actions.
func WithDelete() func(*domain.TableData) {
	return func(t *domain.TableData) {
		t.BtnDelete.IsVisible = true
	}
}
