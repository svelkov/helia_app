package domain

import (
	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
)

// TrustedJS holds a JavaScript expression that is emitted verbatim into an
// HTML event-handler attribute (e.g. hx-on::after-request).
// SECURITY: assign only from compile-time string constants in server-side Go code.
// NEVER assign from user-supplied input (query params, form fields, headers, etc.)
// as it would result in a reflected XSS vulnerability.
type TrustedJS = string

// Config holds handler configuration
type HandlerConfig struct {
	ContentTitle string
	TableID      string
	APIPrefix    string
	IDField      string
}
type TableOptions struct {
	HasUpdate  bool
	HasDelete  bool
	Pagination []int // [page, pageSize]
	HxInclude  string
}

type TableOption func(*TableOptions)

type Dialog struct {
	Id            string
	Title         string
	OkText        string
	CancelText    string
	SaveText      string
	HxActionURL   string
	HxTarget      string
	HxSwap        string
	HxOn          string
	HxRequestType string
}

// SearchPopupConfig holds configuration for the search popup
type SearchPopupConfig struct {
	FieldID         string
	SearchURL       string
	Placeholder     string
	InitialValue    string
	DisplayFieldID  string // Optional: ID of the display field (e.g., "naziv")
	Width           string // Optional: custom width class
	MaxHeight       string // Optional: custom max height for results
	ClearButtonText string // Optional: custom clear button text
	Position        string // Optional: positioning class (default: "left-0 top-full")
}

// NewSearchPopupConfig creates a new config with sensible defaults
func NewSearchPopupConfig(fieldID, searchURL, placeholder string) SearchPopupConfig {
	return SearchPopupConfig{
		FieldID:         fieldID,
		SearchURL:       searchURL,
		Placeholder:     placeholder,
		DisplayFieldID:  fieldID + "naziv", // Default naming convention
		Width:           "w-80",            // Default width
		MaxHeight:       "max-h-60",        // Default max height
		ClearButtonText: "Clear ❌",         // Default clear text
		Position:        "left-0 top-full", // Default position
	}
}

// SearchResultRow represents a single row of data in the search results table
type SearchResultRow struct {
	Value        string   // The actual value to be stored in the field
	DisplayValue string   // The main display value (usually the first column)
	Cells        []string // All cell values including the display value
	IsClickable  bool     // Whether this row should be selectable
}

type Button struct {
	Id               string
	LabelText        string
	HxActionURL      string
	HxTarget         string
	HxSwap           string
	HxOnClick        string
	HxOnClickArg     string // Optional argument for HxOnClick function
	HxInclude        string
	HxVals           string
	HxRequestType    string
	HxOnAfterRequest TrustedJS
	IdDialog         string
	ActionMethod     string
	Icon             string
	IsVisible        bool
	IsDisabled       bool
	BtnClass         string
	DataFields       string // Parameters for print comma separated, e.g. `param1,param2...
	HxHeaders        string // JSON string for custom headers, e.g. `{"X-Custom-Header": "value"}`
	OpenDialog       bool   // Whether this button opens a dialog
}
type Fields struct {
	Name            string
	Label           string
	Type            string
	ValidationText  string
	Value           string
	Width           string
	TabIndex        string
	SkipInSearch    bool
	Field           string
	Sortable        bool
	Params          map[string]string
	IncludeInTotals bool
	TextAlign       string
}
type FieldError struct {
	Field        string `json:"field"`
	ErrorMessage string `json:"message"`
}

type Response struct {
	StatusCode int          `json:"statusCode"`
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Errors     []FieldError `json:"errors,omitempty"`
	HxTrigger  string       `json:"hxTrigger,omitempty"`
	HxLocation string       `json:"hxLocation,omitempty"`
}

// UserClaims defines the claims (data) stored in the JWT.
// For access tokens: Contains username and CSRF hash
// For refresh tokens: Contains username only with type="refresh" claim
type UserClaims struct {
	Username    string `json:"username"`
	UserID      int    `json:"userid"`
	Email       string `json:"email"`
	Firma       string `json:"firma"`        // Company identifier
	SelectedGod int    `json:"selected_god"` // Fiscal year
	SelectedKar int    `json:"selected_kar"` // Accounting period
	DuzSin      int    `json:"duz_sin"`      // Length of synthetic account
	Language    string `json:"language"`     // UI language
	TokenType   string `json:"token_type"`   // "access" or "refresh"
	CSRFHash    string `json:"csrf_hash"`    // SHA256 hash of CSRF token (access tokens only)
	Mesto       string `json:"mesto"`        // User's city (for report headers)
	jwt.RegisteredClaims
}

// UserClaims defines the claims (data) stored in the JWT.
type User struct {
	Username string `json:"username"`
	Password string
	Email    string `json:"email"`
}

// Assuming 'tmpl' is of type '*MyTemplates'
type MyTemplates struct {
	// Your template rendering functions will be fields here
	Table            func(TableData) templ.Component
	ContentContainer func(TableData) templ.Component
	Nalozi           func(TableData) templ.Component
	// ... other template functions ...
}

type TabItem struct {
	ID           string `json:"id"`             // Unique ID for the tab (used in the HTML)
	Label        string `json:"label"`          // Text displayed on the tab button
	HXRequestUrl string `json:"hx_requestUrl"`  // The HTMX GET endpoint for this tab's content
	IsActive     bool   `json:"is_active"`      // Optional: To mark the initially active tab
	Name         string `json:"name"`           // Optional: Name of the tab, if needed for identification
	Icon         string `json:"icon,omitempty"` // Optional: Icon for the tab, if needed
}

type TabData struct {
	Tabs []TabItem `json:"tabs"`
	// You might have other data related to the overall page here
}

// New struct containing only the desired fields for the JSON response
type ComboItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LabelFieldConfig struct {
	ID         string
	LabelText  string
	ClassLabel string
}

type InputFieldConfig struct {
	ID               string
	Name             string
	Required         bool
	Placeholder      string
	Value            string
	FieldType        string
	Disabled         bool
	ClassInput       string
	BlurEndpoint     string
	HxTarget         string
	HxGet            string
	HxTrigger        string
	HxSwap           string
	HxVals           string
	HxInclude        string
	MinLength        string
	MaxLength        string
	Pattern          string
	TabIndex         string
	HxOnAfterRequest TrustedJS
	OnInput          string
	OnFocus          string
	DecimalPlaces    int
	DataDestField    string // Optional: ID of the field where the value should be stored, used for search popups
}

type ComboFieldConfig struct {
	ID               string
	Name             string
	Required         bool
	Placeholder      string
	Pattern          string
	Value            string
	LabelText        string
	FieldType        string
	HasLabel         bool
	Disabled         bool
	ClassSelect      string
	ClassLabel       string
	BlurEndpoint     string
	OptionsEndpoint  string
	OptionValues     []ComboItem
	SelectedValue    string
	ChangeEndpoint   string
	HxTarget         string
	HxChangeTarget   string
	HxVals           string
	HxInclude        string
	HxOnAfterRequest TrustedJS
	HxSwap           string
	HxParams         string
	HxGet            string
	HxTrigger        string
	MinLength        string
	MaxLength        string
	TabIndex         string
	OnInput          string
	OnFocus          string
}
type CheckboxFieldConfig struct {
	ID                string
	Name              string
	LabelText         string
	ClassLabel        string
	ClassCheckbox     string
	ClassCheckboxSpan string
	IsChecked         bool
	Disabled          bool
	OnChange          string
	OnChangeEndpoint  string
	HxTarget          string
	HxSwap            string
	HxVals            string
	TabIndex          string
	Value             string
	Checked           bool
}
type RadioFieldConfig struct {
	ID               string
	Name             string
	LabelText        string
	ClassLabel       string
	ClassSelect      string
	IsSelected       bool
	Disabled         bool
	OnChange         string
	OnChangeEndpoint string
	HxTarget         string
	HxSwap           string
	HxVals           string
	TabIndex         string
	Value            string
}

type SearchButtonConfig struct {
	ID          string
	Name        string
	HxUrl       string
	HxTarget    string
	HxSwap      string
	HxVals      string
	ClassButton string
	Icon        string
	Disabled    bool
}

type ReportParameters struct {
	ReportName     string
	CompanyName    string
	PIB            string
	MatBroj        string
	SifDel         string
	Adress         string
	Postcode       string
	City           string
	TekRac         string
	Telefon        string
	CompanyLogo    string
	UserName       string
	ParameterItems map[string]ParameterItem
	Orientation    string
}

type Komintent struct {
	Kar   int
	Naziv string
}

type ParameterItem struct {
	Name  string
	Value string
}
