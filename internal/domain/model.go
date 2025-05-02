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

type Button struct {
	Id            string
	LabelText     string
	HxActionURL   string
	HxTarget      string
	HxSwap        string
	HxOn          string
	HxRequestType string
	IdDialog      string
}
type Fields struct {
	Name           string
	Label          string
	Type           string
	ValidationText string
	Value          string
	Width          string
	TabIndex       string
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
	ID           string `json:"id"`            // Unique ID for the tab (used in the HTML)
	Label        string `json:"label"`         // Text displayed on the tab button
	HXRequestUrl string `json:"hx_requestUrl"` // The HTMX GET endpoint for this tab's content
	IsActive     bool   `json:"is_active"`     // Optional: To mark the initially active tab
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
