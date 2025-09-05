package domain

import (
	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
)

type MenuResponse struct {
	Submenu []string `json:"submenu"`
}
type SubMenuItem struct {
	Name string        `json:"name"`
	URL  templ.SafeURL `json:"url,omitempty"`
	Icon string        `json:"icon,omitempty"`
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
	Id            string
	LabelText     string
	HxActionURL   string
	HxTarget      string
	HxSwap        string
	HxOn          string
	HxInclude     string
	HxRequestType string
	IdDialog      string
	ActionMethod  string
	Icon          string
	IsVisible     bool
}
type Fields struct {
	Name           string
	Label          string
	Type           string
	ValidationText string
	Value          string
	Width          string
	TabIndex       string
	SkipInSearch   bool
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
type UserClaims struct {
	Username string `json:"username"`
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
type TipdokComboItem struct {
	TipDok string `json:"tip_dok"`
	Opis   string `json:"opis"`
}
