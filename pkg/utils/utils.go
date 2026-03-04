// Shared utilities (e.g., date parsing)
package utils

import (
	"encoding/json"
	"fmt"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"html/template"
	"net/http"
	"reflect"
	"strings"

	tmpl "helia/frontend/templates"
	tmpl_opsti "helia/frontend/templates/opstipodaci"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

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
		"icon": func(name string) template.HTML {
			if svg, exists := common.IconSVG[name]; exists {
				return template.HTML(svg)
			}
			return template.HTML("")
		},
	}
}

func RenderContent(c *gin.Context, table domain.TableData, tmplName ...string) {
	reqURI := c.Request.URL.RequestURI()
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

	var err error
	switch templateName {
	case "Table":
		err = tmpl.Table(table, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	case "ContentContainer":
		translaotr := i18n.GetInstance()
		searchControl := domain.InputControl{
			ID:           "search-control",
			Label:        translaotr.Label("Pretraži"),
			Type:         "search",
			Placeholder:  "Unesite tekst za pretragu",
			HxActionURL:  table.URLGetAll,
			HxTarget:     fmt.Sprintf("#%s", table.TableID),
			HxSwap:       "innerHTML",
			HxTrigger:    "keyup changed delay:500ms",
			Autocomplete: "off",
			Class:        common.ClassSearchInput,
		}
		err = tmpl.ContentContainer(table, searchControl, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	default:
		c.JSON(400, gin.H{"error": fmt.Sprintf("Template '%s' not found", templateName)})
		return
	}

}

func RenderDialogContent(c *gin.Context, dialog domain.Dialog, fields []domain.Fields, actionType string, btnSave, btnCancel, btnClose domain.Button, rowID string) error {
	var component templ.Component
	translator := i18n.GetInstance()
	csrfToken := common.GetCsrfToken(c)
	switch actionType {
	case "DELETE":
		btnSave.HxOnAfterRequest = "handleDeleteResponse"
		component = tmpl.DeleteDialog(csrfToken, dialog, btnSave, btnCancel, btnClose, rowID, translator)
	case "ADD", "UPDATE":
		component = tmpl_opsti.AddUpdateForm(csrfToken, dialog, fields, btnSave, btnCancel, btnClose, translator)
	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}

	return component.Render(c.Request.Context(), c.Writer)
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
func SetButton(id, labelText, actionUrl, hxTarget, hxSwap, hxOn, hxInclude, hxVals, hxRequestType, hxOnAfterRequest, idDialog, actionMethod, icon, class string, isVisible bool) domain.Button {
	return domain.Button{
		Id:               id,
		LabelText:        labelText,
		HxActionURL:      actionUrl,
		HxTarget:         hxTarget,
		HxSwap:           hxSwap,
		HxOn:             hxOn,
		HxInclude:        hxInclude,
		HxVals:           hxVals,
		HxRequestType:    hxRequestType,
		HxOnAfterRequest: hxOnAfterRequest,
		IdDialog:         idDialog,
		ActionMethod:     actionMethod,
		Icon:             icon,
		BtnClass:         class,
		IsVisible:        isVisible,
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

func GetFieldsFromCacheForUpdate(fieldCache map[string]reflect.StructField) ([]domain.Fields, error) {
	var fields []domain.Fields

	for _, field := range fieldCache {
		dbTag, dbOk := field.Tag.Lookup("db")
		addUpdateTag, addUpdateOk := field.Tag.Lookup("addupdate")
		_, formOk := field.Tag.Lookup("form")

		if dbOk && addUpdateOk && addUpdateTag == "true" && formOk {
			fieldInfo := domain.Fields{
				Name: dbTag,
				Type: field.Type.String(),
			}
			fields = append(fields, fieldInfo)
		}
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields found for update")
	}

	return fields, nil
}
