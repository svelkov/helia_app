package common

const (
	IDdrzave   = "iddrzave"
	IDbanke    = "idbanke"
	IDsifop    = "idsifop"
	IDpartneri = "idpartneri"
	IDtipdok   = "idtipdok"
	IDdokvrsta = "dokvrstaid"
	IDpopdv    = "popdvid"
	IDorgjed   = "idorgjed"
	IDmestotr  = "mestotrid"
	IDsifplizv = "sifplizvid"
	IDfvknjrac = "idfvknjrac"
	IDsifmesto = "sifm"
	IDbnkizv   = "bnkizvid"
	IDfvepdv   = "fvepdvid"
	IDfkpl     = "idfkpl"
	IDfnal     = "idfnal"
	IDfpro     = "idfpro"
	IDoamgrp   = "oamgrpid"
)
const (
	ActionDelete = "DELETE"
	ActionAdd    = "ADD"
	ActionUpdate = "UPDATE"
)

const (
	ErrMsgParseForm           = "Parsiranje forme nije uspelo."
	ErrMsgGetIdFromUrl        = "Neuspešno preuzimanje ID-a iz URL-a."
	ErrMsgFormDecode          = "Neuspešno dekodiranje forme."
	ErrMsgValidation          = "Greške prilkom validacije"
	ErrMsgSaveData            = "Greška prilikom upisa podataka"
	ErrMsgDeleteData          = "Greška prilikom brisanja podataka. Greska: %s"
	ErrMsgInvalidId           = "Invalid ID"
	ErrMsgReadData            = "Greška prilikom čitanja podataka"
	ErrMsgRenderTemplate      = "Error rendering template"
	ErrMsgInvalidID           = "Invalid ID provided in request"
	ErrMsgGetTotalRecords     = "Failed to get total records"
	ErrMsgGetIDFromURL        = "failed to get ID from URL"
	ErrMsgFailedToSetTableRow = "Failed to set table rows"
	OkMsgSaveData             = "Uspešno upisani podaci"
	OkMsgDeleteData           = "Uspešno obrisani podaci"
	OkMsgReadData             = "Uspešno učitani podaci"
	OkMsgOperationSuccessfull = "operation successful"
	ErrMsgObavezanPodatak     = "obavezan podatak..."
	ErrMsgGetData             = "greska prilikom preuzimanja podataka"
	ErrMsgGetKontoSifra       = "nepostojeci konto ili sifra"
)

// constants for styling of inputs
// const (
// 	ClassInputNumericDisabled = "h-6 px-1 py-1 text-sm text-right border border-blue-400 rounded focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-200 disabled:text-gray-500 disabled:cursor-not-allowed "
// 	ClassInputTextEnabled     = "h-6 text-sm rounded border border-blue-400 focus:bg-blue-100 focus:border-blue-500 focus:ring-blue-500 px-1 border text-left"
// 	ClassInputTextDisabled    = "h-6 px-1 text-sm min-w-0 text-left rounded border border-blue-400 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-200 disabled:text-gray-500 disabled:cursor-not-allowed"
// 	ClassLabel                = "h-6 text-sm text-gray-700 whitespace-nowrap flex-shrink-0"
// 	ClassSelect               = "h-6 text-sm text-black rounded border border-blue-400 focus:border-blue-500 focus:ring-blue-500 px-1"
// 	//ClassButton               = "h-6 text-sm px-1 bg-gray-200 hover:bg-gray-300 rounded border border-red-400 flex items-center justify-center flex-shrink-0"
// 	ClassButton              = "h-6 text-sm bg-gray-200 hover:bg-gray-300 px-1 rounded border border-blue-400 flex items-center justify-center flex-shrink-0"
// 	ClassSearchInput         = "search-input border rounded px-2 h-7 py-1 text-xs w-64 pl-8"
// 	PatternLettersAndNumbers = "[A-Za-z0-9]+"
// 	PatternNumbers           = "[0-9]+"
// 	ClassAddButton           = "bg-blue-600 rounded h-7 px-1 py-1 mr-1 rounded text-xs flex text-white items-center"
// 	ClassSaveButton          = "bg-green-600 rounded h-7 px-1 py-1 mr-1 rounded text-xs flex text-white items-center"
// 	ClassDeleteButton        = "bg-red-600 rounded h-7 px-1 py-1 rounded text-xs flex text-white items-center justify-center w-20"
// 	ClassOdustaniButton      = "bg-gray-600 rounded h-7 px-1 py-1 mr-1 rounded text-xs flex text-white items-center justify-center w-20"
// 	ClassCloseButton         = "bg-gray-600 rounded h-7 px-1 py-1 mr-1 rounded text-xs flex text-white items-center"
// 	ClassObradaButton        = "bg-green-600 rounded h-6 px-1 py-1 rounded text-xs flex text-white items-center justify-center w-20"
// 	ClassPrintButton         = "bg-blue-500 hover:bg-blue-700 text-white h-6 px-1 py-1 rounded text-xs flex items-center text-center w-20"
// 	ClassDialogCloseButton   = "text-white hover:text-gray-300 mr-4"
// 	ClassConfirmButton       = "bg-red-600 rounded h-7 px-1 py-1 mr-1 text-xs flex text-white items-center"
// )

// Responsive Tailwind CSS Classes
// Mobile-first approach: base styles for mobile, then tablet (md:), then desktop (lg:)

const (
	// Input - Numeric (Disabled)
	ClassInputNumericDisabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 py-1 text-xs sm:text-sm border border-blue-400 rounded focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-200 disabled:text-gray-500 disabled:cursor-not-allowed text-right w-full"
	// Input - Text (Enabled)
	ClassInputTextEnabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm rounded border border-blue-400 focus:bg-blue-100 focus:border-blue-500 focus:ring-blue-500 text-left w-full"
	// Input - Text (Disabled)
	ClassInputTextDisabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm min-w-0 text-left rounded border border-blue-400 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-200 disabled:text-gray-500 disabled:cursor-not-allowed w-full"
	// Label
	ClassLabel = "h-8 sm:h-7 md:h-6 text-xs sm:text-sm text-gray-700 whitespace-nowrap flex-shrink-0 flex items-center"

	// Select/Dropdown
	ClassSelect = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm text-black rounded border border-blue-400 focus:border-blue-500 focus:ring-blue-500 w-full"
	// Generic Button
	ClassButton = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm bg-gray-200 hover:bg-gray-300 rounded border border-blue-400 flex items-center justify-center flex-shrink-0 whitespace-nowrap"
	// Search Input
	ClassSearchInput = "search-input border rounded px-3 sm:px-2 h-9 sm:h-8 md:h-7 py-1 text-xs sm:text-sm w-full sm:w-72 md:w-64 pl-9 sm:pl-8"
	// Action Buttons - Mobile optimized
	ClassAddButton = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max sm:w-24 md:w-24"
	ClassNewButton = "bg-blue-600 hover:bg-blue-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max sm:w-24 md:w-24"
	//	ClassSaveButton        = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max"
	ClassSaveButton        = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-8 md:h-7 md:px-1 sm:px py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap min-w-max sm:w-24 md:w-24"
	ClassDeleteButton      = "bg-red-600 hover:bg-red-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassOdustaniButton    = "bg-gray-600 hover:bg-gray-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassCloseButton       = "bg-gray-600 hover:bg-gray-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-2 min-w-max"
	ClassObradaButton      = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassPrintButton       = "bg-blue-500 hover:bg-blue-900 text-white h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-xs sm:text-sm flex items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassDialogCloseButton = "text-white hover:text-gray-300 p-2 sm:p-1 transition-colors h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-xs sm:text-sm flex items-center justify-center"
	ClassConfirmButton     = "bg-red-600 hover:bg-red-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max"
	ClassErrorField        = "text-red-600 text-xs mt-1"
	ClassIcon              = "h-5 w-5 mr-2 inline-block"
	ClassPdfButton         = "bg-blue-500 hover:bg-blue-900 text-white h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-sm flex items-center whitespace-nowrap"
	ClassExcelButton       = "bg-blue-500 hover:bg-blue-700 text-white h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-xs sm:text-sm flex items-center justify-center whitespace-nowrap w-full sm:w-12 md:w-12"

	// Patterns (unchanged)
	PatternLettersAndNumbers = "[A-Za-z0-9]+"
	PatternNumbers           = "[0-9]+"
	// Classes for table paginator
	ClassTablePageButtonSelected = "block px-3 py-1 leading-tight text-blue-700 bg-blue-100 border border-gray-300"
	ClassTablePageButton         = "block px-3 py-1 leading-tight text-gray-500 bg-white border border-gray-300 rounded-r-lg hover:bg-gray-100 hover:text-gray-700"
)

// Additional helper classes for responsive layouts
const (
	// Container for form fields
	ClassFormRow = "flex flex-col sm:flex-row gap-2 sm:gap-3 md:gap-4 items-stretch sm:items-center mb-3 sm:mb-2"
	// Container for button groups
	ClassButtonGroup = "flex flex-col sm:flex-row gap-2 sm:gap-2 md:gap-1 items-stretch sm:items-center"
	// Grid layouts
	ClassGridForm2Col = "grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4"
	ClassGridForm3Col = "grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4"
	// Card/Dialog container
	ClassDialog = "bg-white rounded-lg shadow-xl w-full max-w-full sm:max-w-lg md:max-w-2xl lg:max-w-4xl mx-auto p-4 sm:p-6 md:p-8"
	// Table container for mobile
	ClassTableContainer = "overflow-x-auto -mx-4 sm:mx-0"
	// Mobile menu button (hamburger)
	ClassMobileMenuButton = "inline-flex items-center justify-center p-2 rounded-md text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500 sm:hidden"
)

// <svg class="h-5 w-5 mr-2" fill="none" stroke="currentColor">
//
//		<path
//			fill-rule="evenodd"
//			d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Zm6-10.125a1.875 1.875 0 1 1-3.75 0 1.875 1.875 0 0 1 3.75 0Zm1.294 6.336a6.721 6.721 0 0 1-3.17.789 6.721 6.721 0 0 1-3.168-.789 3.376 3.376 0 0 1 6.338 0Z"
//		></path>
//	</svg>
//
// iconSVG contains all SVG path definitions for inline rendering
var IconSVG = map[string]string{
	"partneri":           `<path fill-rule="evenodd"  d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Zm6-10.125a1.875 1.875 0 1 1-3.75 0 1.875 1.875 0 0 1 3.75 0Zm1.294 6.336a6.721 6.721 0 0 1-3.17.789 6.721 6.721 0 0 1-3.168-.789 3.376 3.376 0 0 1 6.338 0Z" />`,
	"vrstenaloga":        `<path fill-rule="evenodd" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 7.5h1.5m-1.5 3h1.5m-7.5 3h7.5m-7.5 3h7.5m3-9h3.375c.621 0 1.125.504 1.125 1.125V18a2.25 2.25 0 0 1-2.25 2.25M16.5 7.5V18a2.25 2.25 0 0 0 2.25 2.25M16.5 7.5V4.875c0-.621-.504-1.125-1.125-1.125H4.125C3.504 3.75 3 4.254 3 4.875V18a2.25 2.25 0 0 0 2.25 2.25h13.5M6 7.5h3v3H6v-3Z" />`,
	"vrstedokumenta":     `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z" />`,
	"vrsteporknjige":     `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6.429 9.75 2.25 12l4.179 2.25m0-4.5 5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0 4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0-5.571 3-5.571-3" />`,
	"vrsteevpdv":         `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 14.25 6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0c1.1.128 1.907 1.077 1.907 2.185ZM9.75 9h.008v.008H9.75V9Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm4.125 4.5h.008v.008h-.008V13.5Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />`,
	"oj":                 `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.375 19.5h17.25m-17.25 0a1.125 1.125 0 0 1-1.125-1.125M3.375 19.5h7.5c.621 0 1.125-.504 1.125-1.125m-9.75 0V5.625m0 12.75v-1.5c0-.621.504-1.125 1.125-1.125m18.375 2.625V5.625m0 12.75c0 .621-.504 1.125-1.125 1.125m1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125m0 3.75h-7.5A1.125 1.125 0 0 1 12 18.375m9.75-12.75c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125m19.5 0v1.5c0 .621-.504 1.125-1.125 1.125M2.25 5.625v1.5c0 .621.504 1.125 1.125 1.125m0 0h17.25m-17.25 0h7.5c.621 0 1.125.504 1.125 1.125M3.375 8.25c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125m17.25-3.75h-7.5c-.621 0-1.125.504-1.125 1.125m8.625-1.125c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-17.25 0h7.5m-7.5 0c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125M12 10.875v-1.5m0 1.5c0 .621-.504 1.125-1.125 1.125M12 10.875c0 .621.504 1.125 1.125 1.125m-2.25 0c.621 0 1.125.504 1.125 1.125M13.125 12h7.5m-7.5 0c-.621 0-1.125.504-1.125 1.125M20.625 12c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-17.25 0h7.5M12 14.625v-1.5m0 1.5c0 .621-.504 1.125-1.125 1.125M12 14.625c0 .621.504 1.125 1.125 1.125m-2.25 0c.621 0 1.125.504 1.125 1.125m0 1.5v-1.5m0 0c0-.621.504-1.125 1.125-1.125m0 0h7.5" />`,
	"mtroska":            `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 11.625h4.5m-4.5 2.25h4.5m2.121 1.527c-1.171 1.464-3.07 1.464-4.242 0-1.172-1.465-1.172-3.84 0-5.304 1.171-1.464 3.07-1.464 4.242 0M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	"drzave":             `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 6.75V15m6-6v8.25m.503 3.498 4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 0 0-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0Z" />`,
	"opstine":            `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 21v-8.25M15.75 21v-8.25M8.25 21v-8.25M3 9l9-6 9 6m-1.5 12V10.332A48.36 48.36 0 0 0 12 9.75c-2.551 0-5.056.2-7.5.582V21M3 21h18M12 6.75h.008v.008H12V6.75Z" />`,
	"mesta":              `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />`,
	"banke":              `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />`,
	"bankeizv":           `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Z" />`,
	"sifplizv":           `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.242 5.992h12m-12 6.003H20.24m-12 5.999h12M4.117 7.495v-3.75H2.99m1.125 3.75H2.99m1.125 0H5.24m-1.92 2.577a1.125 1.125 0 1 1 1.591 1.59l-1.83 1.83h2.16M2.99 15.745h1.125a1.125 1.125 0 0 1 0 2.25H3.74m0-.002h.375a1.125 1.125 0 0 1 0 2.25H2.99" />`,
	"kontniplan1":        `<path d="M7 3v18"/><path d="M20.4 18.9c.2.5-.1 1.1-.6 1.3l-1.9.7c-.5.2-1.1-.1-1.3-.6L11.1 5.1c-.2-.5.1-1.1.6-1.3l1.9-.7c.5-.2 1.1.1 1.3.6Z"/>`,
	"kontniplan":         `<path d="M7 3v18"/><path d="M20.4 18.9c.2.5-.1 1.1-.6 1.3l-1.9.7c-.5.2-1.1-.1-1.3-.6L11.1 5.1c-.2-.5.1-1.1.6-1.3l1.9-.7c.5-.2 1.1.1 1.3.6Z"/>`,
	"fin_nalozi":         `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 19V4a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v13H7a2 2 0 0 0-2 2Zm0 0a2 2 0 0 0 2 2h12M9 3v14m7 0v4"/>`,
	"fin_promet":         `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 3v4a1 1 0 0 1-1 1H5m4 10v-2m3 2v-6m3 6v-3m4-11v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V7.914a1 1 0 0 1 .293-.707l3.914-3.914A1 1 0 0 1 9.914 3H18a1 1 0 0 1 1 1Z" />`,
	"fin_saldakonta":     `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7h1v12a1 1 0 0 1-1 1h-2a1 1 0 0 1-1-1V5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v14a1 1 0 0 0 1 1h11.5M7 14h6m-6 3h6m0-10h.5m-.5 3h.5M7 7h3v3H7V7Z" />`,
	"fin_kompenzacije":   `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 18v2h6V4H4v2m16 12v2h-6V4h6v2M6.49545 14.4954 4.00003 12m0 0 2.49542-2.49543M4.00003 12h5.94809m7.49798 2.5539L20 12m0 0-2.5539-2.55392M20 12h-5.8319" />`,
	"fin_otvorenestavke": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.03v13m0-13c-2.819-.831-4.715-1.076-8.029-1.023A.99.99 0 0 0 3 6v11c0 .563.466 1.014 1.03 1.007 3.122-.043 5.018.212 7.97 1.023m0-13c2.819-.831 4.715-1.076 8.029-1.023A.99.99 0 0 1 21 6v11c0 .563-.466 1.014-1.03 1.007-3.122-.043-5.018.212-7.97 1.023" />`,
	"fin_kompenzacije3":  `<path fill-rule="evenodd" d="m-0.25,23.94l22.28,0c1.06,0 1.93,-0.75 1.93,-1.68l0,-16.74l-24.21,0l0,18.42zm9.81,-4.19c0.33,0.29 0.33,0.75 0,1.04c-0.17,0.14 -0.38,0.21 -0.6,0.21s-0.43,-0.07 -0.6,-0.21l-1.71,-1.49l-1.71,1.49c-0.17,0.14 -0.38,0.21 -0.6,0.21c-0.22,0 -0.43,-0.07 -0.6,-0.21c-0.33,-0.29 -0.33,-0.75 0,-1.04l1.71,-1.49l-1.71,-1.49c-0.33,-0.29 -0.33,-0.75 0,-1.04c0.33,-0.29 0.87,-0.29 1.19,0l1.71,1.49l1.71,-1.49c0.33,-0.29 0.87,-0.29 1.2,0c0.33,0.29 0.33,0.75 0,1.04l-1.71,1.49l1.71,1.49zm8,1.35l-0.59,0c-0.47,0 -0.84,-0.33 -0.84,-0.73s0.38,-0.73 0.84,-0.73l0.59,0c0.47,0 0.85,0.33 0.85,0.73s-0.38,0.73 -0.85,0.73zm-2.99,-11.27l5.4,0c0.47,0 0.85,0.33 0.85,0.73s-0.38,0.73 -0.85,0.73l-5.4,0c-0.47,0 -0.84,-0.33 -0.84,-0.73s0.38,-0.73 0.84,-0.73zm3.84,6.33c0,0.41 -0.38,0.73 -0.85,0.73l-0.59,0c-0.47,0 -0.84,-0.33 -0.84,-0.73c0,-0.41 0.38,-0.73 0.84,-0.73l0.59,0c0.47,0 0.85,0.33 0.85,0.73zm-3.84,1.37l5.4,0c0.47,0 0.85,0.33 0.85,0.73s-0.38,0.73 -0.85,0.73l-5.4,0c-0.47,0 -0.84,-0.33 -0.84,-0.73s0.38,-0.73 0.84,-0.73zm-10.61,-7.7l1.85,0l0,-1.61c0,-0.41 0.38,-0.73 0.84,-0.73c0.47,0 0.84,0.33 0.84,0.73l0,1.61l1.86,0c0.47,0 0.84,0.33 0.84,0.73s-0.38,0.73 -0.84,0.73l-1.86,0l0,1.61c0,0.41 -0.38,0.74 -0.84,0.74c-0.46,0 -0.84,-0.33 -0.84,-0.74l0,-1.61l-1.85,0c-0.47,0 -0.85,-0.33 -0.85,-0.73s0.38,-0.73 0.85,-0.73z" />`,
	"fin_obracunkamate":  `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m11.0001 18-.8536-.8536c-.0937-.0937-.1464-.2209-.1464-.3535v-4.4172c0-.2422-.08794-.4762-.24744-.6585L4.45127 5.6585C3.88551 5.01192 4.34469 4 5.20385 4H18.7547c.8658 0 1.3225 1.02544.7433 1.66896L16.5001 9m-2.5 9.3754c.3347.3615.7824.6134 1.2788.7195.4771.1584 1.0002.1405 1.464-.05.4638-.1906.8338-.5396 1.0356-.977.2462-.8286-.6363-1.7337-1.7735-1.9948-1.1372-.2611-2.016-1.1604-1.7735-1.9948.2016-.4375.5716-.7868 1.0354-.9774.4639-.1905.9871-.2082 1.4643-.0496.491.1045.9348.3517 1.2689.7067m-1.9397 5.41V20m0-8v.9771"/>`,
	"fin_bilansi":        `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 0 0-2 2v4m5-6h8M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m0 0h3a2 2 0 0 1 2 2v4m0 0v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-6m18 0s-4 2-9 2-9-2-9-2m9-2h.01"/>`,
	"fin_obrada": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13v-2a1 1 0 0 0-1-1h-.757l-.707-1.707.535-.536a1 1 0 0 0 0-1.414l-1.414-1.414a1 1 0 0 0-1.414 0l-.536.535L14 4.757V4a1 1 0 0 0-1-1h-2a1 1 0 0 0-1 1v.757l-1.707.707-.536-.535a1 1 0 0 0-1.414 0L4.929 6.343a1 1 0 0 0 0 1.414l.536.536L4.757 10H4a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h.757l.707 1.707-.535.536a1 1 0 0 0 0 1.414l1.414 1.414a1 1 0 0 0 1.414 0l.536-.535 1.707.707V20a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-.757l1.707-.708.536.536a1 1 0 0 0 1.414 0l1.414-1.414a1 1 0 0 0 0-1.414l-.535-.536.707-1.707H20a1 1 0 0 0 1-1Z"/>
  					<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z"/>`,
	"stampa":      `<path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M16.444 18H19a1 1 0 0 0 1-1v-5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v5a1 1 0 0 0 1 1h2.556M17 11V5a1 1 0 0 0-1-1H8a1 1 0 0 0-1 1v6h10ZM7 15h10v4a1 1 0 0 1-1 1H8a1 1 0 0 1-1-1v-4Z"/>`,
	"odustani":    `<path fill-rule="evenodd" d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25ZM12.75 9a.75.75 0 0 0-1.5 0v2.25H9a.75.75 0 0 0 0 1.5h2.25V15a.75.75 0 0 0 1.5 0v-2.25H15a.75.75 0 0 0 0-1.5h-2.25V9Z" clip-rule="evenodd"></path>`,
	"file_export": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 10V4a1 1 0 0 0-1-1H9.914a1 1 0 0 0-.707.293L5.293 7.207A1 1 0 0 0 5 7.914V20a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-2M10 3v4a1 1 0 0 1-1 1H5m5 6h9m0 0-2-2m2 2-2 2"/>`,
	"file_csv":    `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10V7.914a1 1 0 0 1 .293-.707l3.914-3.914A1 1 0 0 1 9.914 3H18a1 1 0 0 1 1 1v6M5 19v1a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-1M10 3v4a1 1 0 0 1-1 1H5m2.665 9H6.647A1.647 1.647 0 0 1 5 15.353v-1.706A1.647 1.647 0 0 1 6.647 12h1.018M16 12l1.443 4.773L19 12m-6.057-.152-.943-.02a1.34 1.34 0 0 0-1.359 1.22 1.32 1.32 0 0 0 1.172 1.421l.536.059a1.273 1.273 0 0 1 1.226 1.718c-.2.571-.636.754-1.337.754h-1.13"/>`,
	"sort_down":   `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7.119 8h9.762a1 1 0 0 1 .772 1.636l-4.881 5.927a1 1 0 0 1-1.544 0l-4.88-5.927A1 1 0 0 1 7.118 8Z"/>`,
	"sort_up":     `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16.881 16H7.119a1 1 0 0 1-.772-1.636l4.881-5.927a1 1 0 0 1 1.544 0l4.88 5.927a1 1 0 0 1-.77 1.636Z"/>`,
	"sort":        `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m8 10 4-6 4 6H8Zm8 4-4 6-4-6h8Z"/>`,
	"add":         `<path fill-rule="evenodd" d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25ZM12.75 9a.75.75 0 0 0-1.5 0v2.25H9a.75.75 0 0 0 0 1.5h2.25V15a.75.75 0 0 0 1.5 0v-2.25H15a.75.75 0 0 0 0-1.5h-2.25V9Z" clip-rule="evenodd"></path>`,
	//"kontniplan":         `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 19V4a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v13H7a2 2 0 0 0-2 2Zm0 0a2 2 0 0 0 2 2h12M9 3v14m7 0v4" /><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 9h6m-6 3h6m-6 3h6M6.996 9h.01m-.01 3h.01m-.01 3h.01M4 5h16a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1Z" />`,
}
