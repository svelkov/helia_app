package common

import (
	"regexp"
	"strings"
)

const (
	IDdrzave         = "iddrzave"
	IDbanke          = "idbanke"
	IDsifop          = "idsifop"
	IDpartneri       = "idpartneri"
	IDtipdok         = "idtipdok"
	IDdokvrsta       = "dokvrstaid"
	IDpopdv          = "popdvid"
	IDorgjed         = "idorgjed"
	IDmestotr        = "mestotrid"
	IDsifplizv       = "sifplizvid"
	IDfvknjrac       = "idfvknjrac"
	IDsifmesto       = "sifm"
	IDbnkizv         = "bnkizvid"
	IDfvepdv         = "fvepdvid"
	IDfkpl           = "idfkpl"
	IDfnal           = "idfnal"
	IDfpro           = "idfpro"
	IDoamgrp         = "oamgrpid"
	IDbils           = "bilsid"
	IDbilu           = "biluid"
	IDtekracuni      = "tekracuniid"
	IDtipanalitike   = "tipanalitikeid"
	IDentityLocks    = "entitylocksid"
	IDvalute         = "idvalute"
	IDkomercijalista = "komid"
	IDmagacin        = "magaciniid"
	IDmi             = "fispid"
	IDjedmere        = "jedmereid"
	IDtipknjid       = "tipknjid"
	IDKir            = "idkir"
	IDkpr            = "idkpr"
	IDfsepp          = "fseppid"
	IDfizvzag        = "idfizvzag"
	IDkam            = "idkam"
	IDtkam           = "idtkam"
	IDrgru           = "rgruid"
	IDrpgru          = "rpgruid"
	IDrpor           = "rporid"
	IDrsif           = "rsifid"
)

const (
	ActionDelete            = "DELETE"
	ActionAdd               = "ADD"
	ActionUpdate            = "UPDATE"
	TipStampePreview string = "PREVIEW"
	TipStampePrint   string = "PRINT"
)

// Date format constants
const (
	DateLayout  = "02.01.2006" // Serbian date format (DD.MM.YYYY)
	HtmlLayout  = "2006-01-02" // HTML5 date input format (YYYY-MM-DD)
	HtmlLayout1 = "2006-02-01" // HTML5 datetime input format (YYYY-DD-MM HH:MM:SS)
	DateLayouts = "02.01.2006, 2006-01-02, 02/01/2006, 02.01.2006, 02-01-2006"
)

const (
	ErrMsgParseForm             = "Parsiranje forme nije uspelo."
	ErrMsgGetIdFromUrl          = "Neuspešno preuzimanje ID-a iz URL-a."
	ErrMsgFormDecode            = "Neuspešno dekodiranje forme."
	ErrMsgValidation            = "Greška prilikom validacije"
	ErrMsgSaveData              = "Greška prilikom upisa podataka"
	ErrMsgDeleteData            = "Greška prilikom brisanja podataka. Greska: %s"
	ErrMsgInvalidId             = "Nevažeći ID"
	ErrMsgReadData              = "Greška prilikom čitanja podataka"
	ErrMsgRenderTemplate        = "Greška prilikom renderovanja šablona"
	ErrMsgInvalidID             = "Nevažeći ID prosleđen u zahtevu"
	ErrMsgGetTotalRecords       = "Neuspešno preuzimanje ukupnog broja zapisa"
	ErrMsgGetIDFromURL          = "Neuspešno preuzimanje ID-a iz URL-a"
	ErrMsgFailedToSetTableRow   = "Neuspešno postavljanje redova tabele"
	OkMsgSaveData               = "Uspešno upisani podaci"
	OkMsgDeleteData             = "Uspešno obrisani podaci"
	OkMsgReadData               = "Uspešno učitani podaci"
	OkMsgOperationSuccessfull   = "Uspešno izvršena operacija"
	ErrMsgObavezanPodatak       = "obavezan podatak..."
	ErrMsgGetData               = "greska prilikom preuzimanja podataka"
	ErrMsgDataFetch             = "Greška prilikom preuzimanja podataka, greška: %s"
	ErrMsgMissingRequiredParams = "Nedostaju obavezni parametri"
	ErrMsgGetKontoSifra         = "nepostojeci konto ili sifra"
	ErrNoDataFound              = "nema pronadjenih podataka"
	ErrMsgStatusConflict        = "Podaci trenutno obrađuje drugi korisnik, pokušajte ponovo kasnije"
	ErrMsgUnauthorized          = "Niste autorizovani za ovu akciju"
	ErrMsgMissingParameter      = "Nedostaje parametar konto ili vkonta"
	ErrMsgUserSessionNotFound   = "User session not found"
	ErrMsgNotFound              = "Nije pronađena šifra"
	ErrMsgSessionNotFound       = "Korisnička sesija nije dostupna"
	ErrMsgSearchKonto           = "Greška prilikom pretrage konta"
	ErrMsgLockFailed            = "Neuspešno zaključavanje zapisa, pokušajte ponovo"
	ErrMsgUpdate                = "Greska prilikom ažuriranja. "
	ErrMsgFileSelection         = "Niste odabrali fajl za upload ili software"
	ErrMsgSoftwareSelection     = "Niste odabrali kom softwaru pripada fajl za import"
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
	// Input - Numeric (Enabled)
	ClassInputNumericEnabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 py-1 text-xs sm:text-sm border border-blue-400 rounded focus:bg-green-50 focus:border-green-500 focus:ring-green-500 text-right w-full"
	// Input - Numeric (Disabled)
	ClassInputNumericDisabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 py-1 text-xs sm:text-sm border border-blue-400 rounded focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-200 disabled:text-gray-500 disabled:cursor-not-allowed text-right w-full"
	// Input - Text (Enabled)
	ClassInputTextEnabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm rounded border border-blue-400 focus:bg-green-50 focus:border-green-500 focus:ring-green-500 text-left w-full"
	// Input - Text (Disabled)
	ClassInputTextDisabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm min-w-0 text-left rounded border border-blue-400 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-300 disabled:text-gray-600 disabled:cursor-not-allowed w-full"
	// TextArea (Enabled)
	ClassTextAreaEnabled = "px-2 sm:px-1.5 md:px-1 py-2 text-xs sm:text-sm rounded border border-blue-400 focus:bg-green-50 focus:border-green-500 focus:ring-green-500 text-left w-full resize-none"
	// Label
	ClassLabel = "h-8 sm:h-7 md:h-6 text-xs sm:text-sm text-black whitespace-nowrap flex-shrink-0 flex items-center"

	// Select/Dropdown
	ClassSelect         = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm text-black rounded border border-blue-400 focus:border-blue-500 focus:ring-blue-500 w-full"
	ClassSelectDisabled = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm text-black rounded border border-blue-400 focus:border-blue-500 focus:ring-blue-500 disabled:bg-gray-300 disabled:text-gray-600 disabled:cursor-not-allowed w-full"
	ClassSelectEnabled  = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm text-black rounded border border-blue-400 focus:border-blue-500 focus:ring-blue-500 w-full"

	ClassRadioSet = "w-4 h-4 rounded border-blue-400 text-blue-600 focus:ring-blue-500"
	//checkbox
	ClassCheckbox = "h-4 w-4 rounded border-blue-400 text-blue-600 focus:ring-blue-500"
	//Class span for checkbox label
	ClassCheckboxSpan = "text-sm text-black"
	//	Class checkbox label container
	ClassCheckboxLabel = "flex items-center gap-1"
	// Generic Button
	ClassButton = "h-8 sm:h-7 md:h-6 px-2 sm:px-1.5 md:px-1 text-xs sm:text-sm bg-gray-200 hover:bg-gray-300 rounded border border-blue-400 flex items-center justify-center flex-shrink-0 whitespace-nowrap"
	// Search Input
	ClassSearchInput = "search-input border rounded px-3 sm:px-2 h-9 sm:h-8 md:h-7 py-1 text-xs sm:text-sm w-full sm:w-72 md:w-64 pl-9 sm:pl-8"
	// Action Buttons - Mobile optimized
	ClassAddButton = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max sm:w-24 md:w-24"
	ClassNewButton = "bg-blue-600 hover:bg-blue-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max sm:w-24 md:w-24"
	//	ClassSaveButton        = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max"
	ClassSaveButton          = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-8 md:h-7 md:px-1 sm:px py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap min-w-max sm:w-24 md:w-24"
	ClassDeleteButton        = "bg-red-600 hover:bg-red-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassOdustaniButton      = "bg-gray-600 hover:bg-gray-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassButtonDisabled      = "bg-gray-600 hover:bg-gray-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24 opacity-50 cursor-not-allowed"
	ClassCloseButton         = "bg-gray-600 hover:bg-gray-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-2 min-w-max"
	ClassObradaButton        = "bg-green-600 hover:bg-green-700 rounded h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassPrintButton         = "bg-blue-500 hover:bg-blue-900 text-white h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-xs sm:text-sm flex items-center justify-center whitespace-nowrap w-full sm:w-24 md:w-24"
	ClassStampaOpomenaButton = "bg-blue-500 hover:bg-blue-900 text-white h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-xs sm:text-sm flex items-center justify-center whitespace-nowrap w-full sm:w-36 md:w-36"
	ClassDialogCloseButton   = "text-white hover:text-gray-300 p-2 sm:p-1 transition-colors h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-xs sm:text-sm flex items-center justify-center"
	ClassConfirmButton       = "bg-red-600 hover:bg-red-700 rounded h-9 sm:h-8 md:h-7 px-2 sm:px-2 md:px-1 py-1 text-xs sm:text-sm flex text-white items-center justify-center whitespace-nowrap gap-1 min-w-max"
	ClassErrorField          = "text-red-600 text-xs mt-1"
	ClassIcon                = "h-5 w-5 mr-2 inline-block"
	ClassPdfButton           = "bg-blue-500 hover:bg-blue-900 text-white h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-sm flex items-center whitespace-nowrap"
	ClassExcelButton         = "bg-blue-500 hover:bg-blue-700 text-white h-9 sm:h-7 md:h-7 px-2 sm:px-2 md:px-1 py-1 rounded text-xs sm:text-sm flex items-center justify-center whitespace-nowrap w-full sm:w-12 md:w-12"

	// Classes for dialog forms
	ClassMainDialogDiv = "fixed inset-0 bg-black flex items-center justify-center p-4 bg-opacity-40 z-[60] "
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

// iconSVG contains all SVG path definitions for inline rendering
var IconSVG = map[string]string{
	// Glavni meni ikonice
	// Šifarnici — Bookmark (kolekcija šifara)
	"sifarnici": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.593 3.322c1.1.128 1.907 1.077 1.907 2.185V21L12 17.25 4.5 21V5.507c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0Z" />`,
	// Finansijsko — Calculator (obračun)
	"finansijsko": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 15.75V18m-7.5-6.75h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V13.5Zm0 2.25h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V18Zm2.498-6.75h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V13.5Zm0 2.25h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V18Zm2.504-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5Zm0 2.25h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V18Zm2.498-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5ZM8.25 6h7.5v2.25h-7.5V6ZM12 2.25c-1.892 0-3.758.11-5.593.322C5.307 2.7 4.5 3.65 4.5 4.757V19.5a2.25 2.25 0 0 0 2.25 2.25h10.5a2.25 2.25 0 0 0 2.25-2.25V4.757c0-1.108-.806-2.057-1.907-2.185A48.507 48.507 0 0 0 12 2.25Z" />`,
	// 3. Komercijalno poslovanje — Shopping Cart (korpa/prodaja)
	"komercijalno": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 0 0-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 0 0-16.536-1.84M7.5 14.25 5.106 5.272M6 20.25a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Zm12.75 0a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Z" />`,
	// 4. Robno — Cube (roba/artikli)
	"robno": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />`,
	// 5. Blagajna — Cash Register / Building Store (prodavnica)
	"blagajna": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />`,
	// 6. Proizvodnja — Cog/Wrench (proizvodnja)
	"proizvodnja": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.42 15.17 17.25 21A2.652 2.652 0 0 0 21 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 1 1-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 0 0 4.486-6.336l-3.276 3.277a3.004 3.004 0 0 1-2.25-2.25l3.276-3.276a4.5 4.5 0 0 0-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437 1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008Z" />`,
	// 7. Osnovna sredstva — Building Office (zgrade/oprema)
	"osnovnasredstva": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />`,


		// =========================================================
	// SET 1: KLASIČNI POSLOVNI PRISTUP (Opšti podaci / Šifarnici)
	// =========================================================
	// Partneri — Users (grupa korisnika)
	"sifarnici_partneri": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />`,
	// Vrste naloga — DocumentText (dokument)
	"sifarnici_vrstenaloga": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	// Vrste dokumenata — DocumentDuplicate (duplikat)
	"sifarnici_vrstedokumenta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 0 1-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 0 1 1.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.243-8.161-7.5-8.876a9.06 9.06 0 0 0-1.5-.124H9.375c-.621 0-1.125.504-1.125 1.125v3.5m7.5 10.375H9.375a1.125 1.125 0 0 1-1.125-1.125v-9.25m12 6.625v-1.875a3.375 3.375 0 0 0-3.375-3.375h-1.5a1.125 1.125 0 0 1-1.125-1.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H9.75" />`,
	// Vrste por. knjige — BookOpen (otvorena knjiga)
	"sifarnici_vrsteporknjige": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25" />`,
	// Vrste PDV — ReceiptPercent (procenat na računu)
	"sifarnici_vrsteevpdv": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 14.25l6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0c1.1.128 1.907 1.077 1.907 2.185ZM9.75 9h.008v.008H9.75V9Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm4.125 4.5h.008v.008h-.008V13.5Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />`,
	// OJ — BuildingOffice (zgrada)
	"sifarnici_oj": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />`,
	// Mesta troška — Banknotes (novac)
	"sifarnici_mtroska": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 0 0-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 0 1-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 0 0 3 15h-.75M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm3 0h.008v.008H18V10.5Zm-12 0h.008v.008H6V10.5Z" />`,
	// Države — Globe (globus)
	"sifarnici_drzave": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />`,
	// Opštine — MapPin (pin lokacije)
	"sifarnici_opstine": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" />`,
	// Mesta — Map (mapa)
	"sifarnici_mesta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 6.75V15m6-6v8.25m.503 3.498 4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 0 0-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0Z" />`,
	// Banke — BuildingLibrary (institucija)
	"sifarnici_banke": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 21v-8.25M15.75 21v-8.25M8.25 21v-8.25M3 9l9-6 9 6m-1.5 12V10.332A48.36 48.36 0 0 0 12 9.75c-2.551 0-5.056.2-7.5.582V21M3 21h18M12 6.75h.008v.008H12V6.75Z" />`,
	// Banke izvodi — DocumentChartBar (dokument sa grafikonom)
	"sifarnici_bankeizv": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25M9 16.5v.75m3-3v3M15 12v5.25m-4.5-15H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	// Šif pl. izvoda — ListBullet (lista)
	"sifarnici_sifplizv": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0ZM3.75 12h.007v.008H3.75V12Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm-.375 5.25h.007v.008H3.75v-.008Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />`,
	// Opsti podaci submenu
	// "sifarnici_partneri":       `<path fill-rule="evenodd"  d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Zm6-10.125a1.875 1.875 0 1 1-3.75 0 1.875 1.875 0 0 1 3.75 0Zm1.294 6.336a6.721 6.721 0 0 1-3.17.789 6.721 6.721 0 0 1-3.168-.789 3.376 3.376 0 0 1 6.338 0Z" />`,
	// "sifarnici_vrstenaloga":    `<path fill-rule="evenodd" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 7.5h1.5m-1.5 3h1.5m-7.5 3h7.5m-7.5 3h7.5m3-9h3.375c.621 0 1.125.504 1.125 1.125V18a2.25 2.25 0 0 1-2.25 2.25M16.5 7.5V18a2.25 2.25 0 0 0 2.25 2.25M16.5 7.5V4.875c0-.621-.504-1.125-1.125-1.125H4.125C3.504 3.75 3 4.254 3 4.875V18a2.25 2.25 0 0 0 2.25 2.25h13.5M6 7.5h3v3H6v-3Z" />`,
	// "sifarnici_vrstedokumenta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z" />`,
	// "sifarnici_vrsteporknjige": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6.429 9.75 2.25 12l4.179 2.25m0-4.5 5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0 4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0-5.571 3-5.571-3" />`,
	// "sifarnici_vrsteevpdv":     `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 14.25 6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0c1.1.128 1.907 1.077 1.907 2.185ZM9.75 9h.008v.008H9.75V9Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm4.125 4.5h.008v.008h-.008V13.5Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />`,
	// "sifarnici_oj":             `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.375 19.5h17.25m-17.25 0a1.125 1.125 0 0 1-1.125-1.125M3.375 19.5h7.5c.621 0 1.125-.504 1.125-1.125m-9.75 0V5.625m0 12.75v-1.5c0-.621.504-1.125 1.125-1.125m18.375 2.625V5.625m0 12.75c0 .621-.504 1.125-1.125 1.125m1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125m0 3.75h-7.5A1.125 1.125 0 0 1 12 18.375m9.75-12.75c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125m19.5 0v1.5c0 .621-.504 1.125-1.125 1.125M2.25 5.625v1.5c0 .621.504 1.125 1.125 1.125m0 0h17.25m-17.25 0h7.5c.621 0 1.125.504 1.125 1.125M3.375 8.25c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125m17.25-3.75h-7.5c-.621 0-1.125.504-1.125 1.125m8.625-1.125c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-17.25 0h7.5m-7.5 0c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125M12 10.875v-1.5m0 1.5c0 .621-.504 1.125-1.125 1.125M12 10.875c0 .621.504 1.125 1.125 1.125m-2.25 0c.621 0 1.125.504 1.125 1.125M13.125 12h7.5m-7.5 0c-.621 0-1.125.504-1.125 1.125M20.625 12c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-17.25 0h7.5M12 14.625v-1.5m0 1.5c0 .621-.504 1.125-1.125 1.125M12 14.625c0 .621.504 1.125 1.125 1.125m-2.25 0c.621 0 1.125.504 1.125 1.125m0 1.5v-1.5m0 0c0-.621.504-1.125 1.125-1.125m0 0h7.5" />`,
	// "sifarnici_mtroska":        `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 11.625h4.5m-4.5 2.25h4.5m2.121 1.527c-1.171 1.464-3.07 1.464-4.242 0-1.172-1.465-1.172-3.84 0-5.304 1.171-1.464 3.07-1.464 4.242 0M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	// "sifarnici_drzave":         `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 6.75V15m6-6v8.25m.503 3.498 4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 0 0-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0Z" />`,
	// "sifarnici_opstine":        `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 21v-8.25M15.75 21v-8.25M8.25 21v-8.25M3 9l9-6 9 6m-1.5 12V10.332A48.36 48.36 0 0 0 12 9.75c-2.551 0-5.056.2-7.5.582V21M3 21h18M12 6.75h.008v.008H12V6.75Z" />`,
	// "sifarnici_mesta":          `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />`,
	// "sifarnici_banke":          `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />`,
	// "sifarnici_bankeizv":       `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Z" />`,
	// "sifarnici_sifplizv":       `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.242 5.992h12m-12 6.003H20.24m-12 5.999h12M4.117 7.495v-3.75H2.99m1.125 3.75H2.99m1.125 0H5.24m-1.92 2.577a1.125 1.125 0 1 1 1.591 1.59l-1.83 1.83h2.16M2.99 15.745h1.125a1.125 1.125 0 0 1 0 2.25H3.74m0-.002h.375a1.125 1.125 0 0 1 0 2.25H2.99" />`,
	// finansijko submenu
	//  1. Tipovi analitike — Chart with trend line
	"fin_tipovianalitike": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 3v11.25A2.25 2.25 0 0 0 6 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0 1 18 16.5h-2.25m-7.5 0h7.5m-7.5 0-1 3m8.5-3 1 3m0 0 .5 1.5m-.5-1.5h-9.5m0 0-.5 1.5m.75-9 3-3 2.148 2.148A12.061 12.061 0 0 1 16.5 7.605" />`,
	// 2 Kontni plan
	"fin_kontniplan": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 6.878V6a2.25 2.25 0 0 1 2.25-2.25h7.5A2.25 2.25 0 0 1 18 6v.878m-12 0c.235-.083.487-.128.75-.128h10.5c.263 0 .515.045.75.128m-12 0A2.25 2.25 0 0 0 4.5 9v.878m13.5-3A2.25 2.25 0 0 1 19.5 9v.878m0 0a2.246 2.246 0 0 0-.75-.128H5.25c-.263 0-.515.045-.75.128m15 0A2.25 2.25 0 0 1 21 12v6a2.25 2.25 0 0 1-2.25 2.25H5.25A2.25 2.25 0 0 1 3 18v-6c0-.98.626-1.813 1.5-2.122" />`,
	// 3. Nalozi za knjizenje
	"fin_nalozi": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 19V4a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v13H7a2 2 0 0 0-2 2Zm0 0a2 2 0 0 0 2 2h12M9 3v14m7 0v4"/>`,
	// 4. Dnevnik knjiženja — Open Book
	"fin_dnevnikknjizenja": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25" />`,
	// 5. Promet
	"fin_promet": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 3v4a1 1 0 0 1-1 1H5m4 10v-2m3 2v-6m3 6v-3m4-11v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V7.914a1 1 0 0 1 .293-.707l3.914-3.914A1 1 0 0 1 9.914 3H18a1 1 0 0 1 1 1Z" />`,
	// 6. Salda konta
	"fin_saldakonta": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7h1v12a1 1 0 0 1-1 1h-2a1 1 0 0 1-1-1V5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v14a1 1 0 0 0 1 1h11.5M7 14h6m-6 3h6m0-10h.5m-.5 3h.5M7 7h3v3H7V7Z" />`,
	// 7. kompenzacije
	"fin_kompenzacije":  `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 18v2h6V4H4v2m16 12v2h-6V4h6v2M6.49545 14.4954 4.00003 12m0 0 2.49542-2.49543M4.00003 12h5.94809m7.49798 2.5539L20 12m0 0-2.5539-2.55392M20 12h-5.8319" />`,
	"fin_kompenzacije3": `<path fill-rule="evenodd" d="m-0.25,23.94l22.28,0c1.06,0 1.93,-0.75 1.93,-1.68l0,-16.74l-24.21,0l0,18.42zm9.81,-4.19c0.33,0.29 0.33,0.75 0,1.04c-0.17,0.14 -0.38,0.21 -0.6,0.21s-0.43,-0.07 -0.6,-0.21l-1.71,-1.49l-1.71,1.49c-0.17,0.14 -0.38,0.21 -0.6,0.21c-0.22,0 -0.43,-0.07 -0.6,-0.21c-0.33,-0.29 -0.33,-0.75 0,-1.04l1.71,-1.49l-1.71,-1.49c-0.33,-0.29 -0.33,-0.75 0,-1.04c0.33,-0.29 0.87,-0.29 1.19,0l1.71,1.49l1.71,-1.49c0.33,-0.29 0.87,-0.29 1.2,0c0.33,0.29 0.33,0.75 0,1.04l-1.71,1.49l1.71,1.49zm8,1.35l-0.59,0c-0.47,0 -0.84,-0.33 -0.84,-0.73s0.38,-0.73 0.84,-0.73l0.59,0c0.47,0 0.85,0.33 0.85,0.73s-0.38,0.73 -0.85,0.73zm-2.99,-11.27l5.4,0c0.47,0 0.85,0.33 0.85,0.73s-0.38,0.73 -0.85,0.73l-5.4,0c-0.47,0 -0.84,-0.33 -0.84,-0.73s0.38,-0.73 0.84,-0.73zm3.84,6.33c0,0.41 -0.38,0.73 -0.85,0.73l-0.59,0c-0.47,0 -0.84,-0.33 -0.84,-0.73c0,-0.41 0.38,-0.73 0.84,-0.73l0.59,0c0.47,0 0.85,0.33 0.85,0.73zm-3.84,1.37l5.4,0c0.47,0 0.85,0.33 0.85,0.73s-0.38,0.73 -0.85,0.73l-5.4,0c-0.47,0 -0.84,-0.33 -0.84,-0.73s0.38,-0.73 0.84,-0.73zm-10.61,-7.7l1.85,0l0,-1.61c0,-0.41 0.38,-0.73 0.84,-0.73c0.47,0 0.84,0.33 0.84,0.73l0,1.61l1.86,0c0.47,0 0.84,0.33 0.84,0.73s-0.38,0.73 -0.84,0.73l-1.86,0l0,1.61c0,0.41 -0.38,0.74 -0.84,0.74c-0.46,0 -0.84,-0.33 -0.84,-0.74l0,-1.61l-1.85,0c-0.47,0 -0.85,-0.33 -0.85,-0.73s0.38,-0.73 0.85,-0.73z" />`,
	// 8. otvorene stavke
	"fin_otvorenestavke": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.03v13m0-13c-2.819-.831-4.715-1.076-8.029-1.023A.99.99 0 0 0 3 6v11c0 .563.466 1.014 1.03 1.007 3.122-.043 5.018.212 7.97 1.023m0-13c2.819-.831 4.715-1.076 8.029-1.023A.99.99 0 0 1 21 6v11c0 .563-.466 1.014-1.03 1.007-3.122-.043-5.018.212-7.97 1.023" />`,
	// 9. Bilansi
	"fin_bilansi": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 0 0-2 2v4m5-6h8M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m0 0h3a2 2 0 0 1 2 2v4m0 0v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-6m18 0s-4 2-9 2-9-2-9-2m9-2h.01"/>`,
	// 10. Poreske knjige i prijava — Shield Check
	"fin_poreskeknjige": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />`,
	// 11. POPDV, PPPDV i IPV prijave — Document + Arrow Up
	"fin_popdvprijave": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m6.75 12-3-3m0 0-3 3m3-3v6m-1.5-15H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	// 12. Evidencija predhodnog poreza EPP — Archive Box
	"fin_evidencijaepp": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m20.25 7.5-.625 10.632a2.25 2.25 0 0 1-2.247 2.118H6.622a2.25 2.25 0 0 1-2.247-2.118L3.75 7.5m8.25 3v6.75m0 0-3-3m3 3 3-3M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125Z" />`,
	// 13. Učitavanje i knjiženje izvoda — Arrow Up Tray
	"fin_ucitavanjeizvoda": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 7.5 7.5 12M12 7.5V16.5" />`,
	// 14. Obracun kamate
	"fin_obracunkamate": `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m11.0001 18-.8536-.8536c-.0937-.0937-.1464-.2209-.1464-.3535v-4.4172c0-.2422-.08794-.4762-.24744-.6585L4.45127 5.6585C3.88551 5.01192 4.34469 4 5.20385 4H18.7547c.8658 0 1.3225 1.02544.7433 1.66896L16.5001 9m-2.5 9.3754c.3347.3615.7824.6134 1.2788.7195.4771.1584 1.0002.1405 1.464-.05.4638-.1906.8338-.5396 1.0356-.977.2462-.8286-.6363-1.7337-1.7735-1.9948-1.1372-.2611-2.016-1.1604-1.7735-1.9948.2016-.4375.5716-.7868 1.0354-.9774.4639-.1905.9871-.2082 1.4643-.0496.491.1045.9348.3517 1.2689.7067m-1.9397 5.41V20m0-8v.9771"/>`,
	//
	// Podmeni: Komercijalno poslovanje
	// Zahtev za nabavku — DocumentPlus (Dokument sa plusom - novi zahtev)
	"komercijala_zahtevzanabavku": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m3.75 9v6m3-3h-6m-1.5-9H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	// Zahtev za ponudu — DocumentText (Dokument sa tekstom)
	"komercijala_zahtevzaponudu": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	// Porudžbenice dobavljačima — ClipboardDocumentList (Klipbord sa listom)
	"komercijala_porudzbenicedobavljacima": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z" />`,
	// Prijemnica materijala/sirovine — ArrowDownTray (Ulaz robe)
	"komercijala_prijemnicamaterijala": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />`,
	// Profakture — DocumentDuplicate (Duplikat dokumenta)
	"komercijala_profakture": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 0 1-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 0 1 1.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.243-8.161-7.5-8.876a9.06 9.06 0 0 0-1.5-.124H9.375c-.621 0-1.125.504-1.125 1.125v3.5m7.5 10.375H9.375a1.125 1.125 0 0 1-1.125-1.125v-9.25m12 6.625v-1.875a3.375 3.375 0 0 0-3.375-3.375h-1.5a1.125 1.125 0 0 1-1.125-1.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H9.75" />`,
	// Nalozi za isporuku — Truck (Kamion)
	"komercijala_nalozizaisporuku": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.25 18.75a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 0 1-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 0 0-3.213-9.193 2.056 2.056 0 0 0-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 0 0-10.026 0 1.106 1.106 0 0 0-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12" />`,
	// Otpremanje robe — ArrowUpTray (Izlaz robe / otprema)
	"komercijala_otpremanjerobe": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />`,

	// ikone za buttone
	"fin_ravnoteza":     `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m16 10 3-3m0 0-3-3m3 3H5v3m3 4-3 3m0 0 3 3m-3-3h14v-3"/>`,
	"fin_knjizenje":     `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 5V4a1 1 0 0 0-1-1H8.914a1 1 0 0 0-.707.293L4.293 7.207A1 1 0 0 0 4 7.914V20a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-5M9 3v4a1 1 0 0 1-1 1H4m11.383.772 2.745 2.746m1.215-3.906a2.089 2.089 0 0 1 0 2.953l-6.65 6.646L9 17.95l.739-3.692 6.646-6.646a2.087 2.087 0 0 1 2.958 0Z"/>`,
	"fin_import":        `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12V7.914a1 1 0 0 1 .293-.707l3.914-3.914A1 1 0 0 1 9.914 3H18a1 1 0 0 1 1 1v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1v-4m5-13v4a1 1 0 0 1-1 1H5m0 6h9m0 0-2-2m2 2-2 2"/>`,
	"fin_azurirajkonta": `<path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M6 4v10m0 0a2 2 0 1 0 0 4m0-4a2 2 0 1 1 0 4m0 0v2m6-16v2m0 0a2 2 0 1 0 0 4m0-4a2 2 0 1 1 0 4m0 0v10m6-16v10m0 0a2 2 0 1 0 0 4m0-4a2 2 0 1 1 0 4m0 0v2"/>`,
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
	"refresh":     `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.651 7.65a7.131 7.131 0 0 0-12.68 3.15M18.001 4v4h-4m-7.652 8.35a7.13 7.13 0 0 0 12.68-3.15M6 20v-4h4"/>`,
	"back":        `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.5 8.046H11V6.119c0-.921-.9-1.446-1.524-.894l-5.108 4.49a1.2 1.2 0 0 0 0 1.739l5.108 4.49c.624.556 1.524.027 1.524-.893v-1.928h2a3.023 3.023 0 0 1 3 3.046V19a5.593 5.593 0 0 0-1.5-10.954Z"/>`,
	"cancel":      `<path fill-rule="evenodd" d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25Zm-1.72 6.97a.75.75 0 1 0-1.06 1.06L10.94 12l-1.72 1.72a.75.75 0 1 0 1.06 1.06L12 13.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L13.06 12l1.72-1.72a.75.75 0 1 0-1.06-1.06L12 10.94l-1.72-1.72Z" clip-rule="evenodd"></path>`,
	"close":       `<path fill-rule="evenodd" d="M2 12C2 6.477 6.477 2 12 2s10 4.477 10 10-4.477 10-10 10S2 17.523 2 12Zm7.707-3.707a1 1 0 0 0-1.414 1.414L10.586 12l-2.293 2.293a1 1 0 1 0 1.414 1.414L12 13.414l2.293 2.293a1 1 0 0 0 1.414-1.414L13.414 12l2.293-2.293a1 1 0 0 0-1.414-1.414L12 10.586 9.707 8.293Z" clip-rule="evenodd"/>`,
	"save":        `<path fill-rule="evenodd" d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12Zm13.36-1.814a.75.75 0 1 0-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 0 0-1.06 1.06l2.25 2.25a.75.75 0 0 0 1.14-.094l3.75-5.25Z" clip-rule="evenodd"></path>`,
	"delete":      `<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 7h14m-9 3v8m4-8v8M10 3h4a1 1 0 0 1 1 1v3H9V4a1 1 0 0 1 1-1ZM6 7h12v13a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V7Z"/>`,
	"file-select": `<path d="m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />`,
	// robno submenu
	"robno_robnadokumenta":       `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	"robno_prometartikli":        `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />`,
	"robno_stanjeartikli":        `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />`,
	"robno_izvestajiulazizlaz":   `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 0 1 3 19.875v-6.75ZM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V8.625ZM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V4.125Z" />`,
	"robno_komercijalnipodaci":   `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.25 14.15v4.25c0 1.094-.787 2.036-1.872 2.18-2.087.277-4.216.42-6.378.42s-4.291-.143-6.378-.42c-1.085-.144-1.872-1.086-1.872-2.18v-4.25m16.5 0a2.18 2.18 0 0 0 .75-1.661V8.706c0-1.081-.768-2.015-1.837-2.175a48.114 48.114 0 0 0-3.413-.387m4.5 8.006c-.194.165-.42.295-.673.38A23.978 23.978 0 0 1 12 15.75c-2.648 0-5.195-.429-7.577-1.22a2.016 2.016 0 0 1-.673-.38m0 0A2.18 2.18 0 0 1 3 12.489V8.706c0-1.081.768-2.015 1.837-2.175a48.111 48.111 0 0 1 3.413-.387m7.5 0V5.25A2.25 2.25 0 0 0 13.5 3h-3a2.25 2.25 0 0 0-2.25 2.25v.894m7.5 0a48.667 48.667 0 0 0-7.5 0M12 12.75h.008v.008H12v-.008Z" />`,
	"robno_artikli":              `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-2.25-1.313M21 7.5v2.25m0-2.25-2.25 1.313M3 7.5l2.25-1.313M3 7.5l2.25 1.313M3 7.5v2.25m9 3 2.25-1.313M12 12.75l-2.25-1.313M12 12.75V15m0 6.75 2.25-1.313M12 21.75V19.5m0 2.25-2.25-1.313m0-16.875L12 2.25l2.25 1.313M21 14.25v2.25l-2.25 1.313m-13.5 0L3 16.5v-2.25" />`,
	"robno_jedinicemere":         `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8.25h18M3 8.25v7.5a2.25 2.25 0 0 0 2.25 2.25h13.5A2.25 2.25 0 0 0 21 15.75v-7.5M6.75 8.25v3m3-3v4.5m3-4.5v3m3-3v4.5m3-4.5v3" />`,
	"robno_robnegrupe":           `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6.429 9.75 2.25 12l4.179 2.25m0-4.5 5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0 4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0-5.571 3-5.571-3" />`,
	"robno_robnepodgrupe":        `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 12.75V12A2.25 2.25 0 0 1 4.5 9.75h15A2.25 2.25 0 0 1 21.75 12v.75m-8.69-6.44-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z" />`,
	"robno_mestaisporuke":        `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" />`,
	"robno_komercijalisti":       `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />`,
	"robno_poreskestope":         `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 14.25l6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0c1.1.128 1.907 1.077 1.907 2.185ZM9.75 9h.008v.008H9.75V9Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm4.125 4.5h.008v.008h-.008V13.5Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />`,
	"robno_vrstedokumenata":      `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 9.776c.112-.017.227-.026.344-.026h15.812c.117 0 .232.009.344.026m-16.5 0a2.25 2.25 0 0 0-1.883 2.542l.857 6a2.25 2.25 0 0 0 2.227 1.932H19.05a2.25 2.25 0 0 0 2.227-1.932l.857-6a2.25 2.25 0 0 0-1.883-2.542m-16.5 0V6A2.25 2.25 0 0 1 6 3.75h3.879a1.5 1.5 0 0 1 1.06.44l2.122 2.12a1.5 1.5 0 0 0 1.06.44H18A2.25 2.25 0 0 1 20.25 9v.776" />`,
	"robno_cenovnik":             `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 0 0 5.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 0 0 9.568 3Z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 6h.008v.008H6V6Z" />`,
	"robno_tipoviknjiznihpisama": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25" />`,
	"robno_magacini":             `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />`,
	"robno_magacinkonto":         `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.25 21v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21m0 0h4.5V3.545M12.75 21h7.5V10.75M2.25 21h1.5m18 0h-18M2.25 9l4.5-1.636M18.75 3l-1.5.545m0 6.205 3 1m1.5.5-1.5-.5M6.75 7.364V3h-3v18m3-13.636 10.5-3.819" />`, // Kraj poslovne godine — Calendar
	"robno_krajposlovnegodine":   `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />`,

	// ---------------------------------------------------------
	// BLAGAJNA
	// 1. Dnevnik blagajne — Calendar + Note (dnevnik zapis)
	"blagajna_dnevnikblagajne": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />`,
	// 2. Kontiranje blagajne — Calculator (kalkulacija)
	"blagajna_kontiranjeblagajne": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 15.75V18m-7.5-6.75h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V13.5Zm0 2.25h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V18Zm2.498-6.75h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V13.5Zm0 2.25h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V18Zm2.504-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5Zm0 2.25h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V18Zm2.498-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5ZM8.25 6h7.5v2.25h-7.5V6ZM12 2.25c-1.892 0-3.758.11-5.593.322C5.307 2.7 4.5 3.65 4.5 4.757V19.5a2.25 2.25 0 0 0 2.25 2.25h10.5a2.25 2.25 0 0 0 2.25-2.25V4.757c0-1.108-.806-2.057-1.907-2.185A48.507 48.507 0 0 0 12 2.25Z" />`,
	// 3. Knjiženje blagajne — Book + Arrow right (knjiženje)
	"blagajna_knjizenjeblagajne": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25M15.75 9.75 18 12m0 0-2.25 2.25M18 12H9" />`,
	// 4. Specifikacija apoena — Clipboard list (specifikacija)
	"blagajna_specifikacijaapoena": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z" />`,
	// 5. Valute — Currency Dollar (valuta)
	"blagajna_valute": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />`,
	// 6. Kursna lista — Arrows right-left (kursna lista)
	"blagajna_kursnalista": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18 9 11.25l4.306 4.306a11.95 11.95 0 0 1 5.814-5.518l2.74-1.22m0 0-5.94-2.281m5.94 2.28-2.28 5.941" />`,
	// 7. Apoeni — Banknotes (novčanice)
	"blagajna_apoeni": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 0 0-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 0 1-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 0 0 3 15h-.75M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm3 0h.008v.008H18V10.5Zm-12 0h.008v.008H6V10.5Z" />`,
	// 8. Tipovi blagajne — Building/Store (POS lokacija)
	"blagajna_tipoviblagajne": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />`,
	// 9. Vrste stavki blagajne — Tag (oznake/šifre)
	"blagajna_vrstestavkiblagajne": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 0 0 5.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 0 0 9.568 3Z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 6h.008v.008H6V6Z" />`,
	// 10. Dodatna knjiženja — Document + Plus (dodavanje)
	"blagajna_dodatnaknjizenja": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m3.75 9v6m3-3H9m1.5-12H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />`,
	// 11. Putni nalozi — Paper airplane / Map (putovanje)
	"blagajna_putninalozi": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5" />`,
	// 12. Dnevnici — Calendar days (kalendar)
	"blagajna_dnevnice": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />`,
	// 13. Tipovi troškova — Receipt percent (troškovi)
	"blagajna_tipovitroskova": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 14.25l6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0c1.1.128 1.907 1.077 1.907 2.185ZM9.75 9h.008v.008H9.75V9Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm4.125 4.5h.008v.008h-.008V13.5Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />`,
	// 14. Prevozna sredstva — Truck (vozilo)
	"blagajna_prevoznasredstva": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.25 18.75a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 0 1-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 0 0-3.213-9.193 2.056 2.056 0 0 0-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 0 0-10.026 0 1.106 1.106 0 0 0-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12" />`,

	// Podmeni: Proizvodnja
	// Matični podaci normativa — DocumentText
	"proizvodnja_maticnipodacinormativa": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />`,
	// Konta troškova — Calculator (isti kao finansijsko)
	"proizvodnja_kontatroskova": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 15.75V18m-7.5-6.75h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V13.5Zm0 2.25h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V18Zm2.498-6.75h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V13.5Zm0 2.25h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V18Zm2.504-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5Zm0 2.25h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V18Zm2.498-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5ZM8.25 6h7.5v2.25h-7.5V6ZM12 2.25c-1.892 0-3.758.11-5.593.322C5.307 2.7 4.5 3.65 4.5 4.757V19.5a2.25 2.25 0 0 0 2.25 2.25h10.5a2.25 2.25 0 0 0 2.25-2.25V4.757c0-1.108-.806-2.057-1.907-2.185A48.507 48.507 0 0 0 12 2.25Z" />`,
	// Radni nalog za proizvodnju — ClipboardDocumentList
	"proizvodnja_radninalog": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />`,
	// Unos dokumenta — PencilSquare
	"proizvodnja_unosdokumenta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />`,
	// Pregled dokumenta — Eye
	"proizvodnja_pregleddokumenta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />`,
	// Specifikacija dokumenta — Bars3 (Lista)
	"proizvodnja_specifikacijadokumenta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />`,
	// Knjiženje dokumenta — BookOpen
	"proizvodnja_knjizenjedokumenta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />`,
	// Promet po artiklima i kontima — ChartBar
	"proizvodnja_prometartiklikonta": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />`,
	// Stanje poluproizvoda — Cube
	"proizvodnja_stanjepoluproizvoda": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />`,
	// Trenutno stanje zaliha — ArchiveBox
	"proizvodnja_trenutnostanjezaliha": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />`,
	// Pregled stanja i cena artikala — Tag
	"proizvodnja_pregledstanjacena": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />`,
	// Obračunska kalkulacija — Calculator
	"proizvodnja_obracunskakalkulacija": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 15.75V18m-7.5-6.75h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V13.5Zm0 2.25h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V18Zm2.498-6.75h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V13.5Zm0 2.25h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V18Zm2.504-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5Zm0 2.25h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V18Zm2.498-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5ZM8.25 6h7.5v2.25h-7.5V6ZM12 2.25c-1.892 0-3.758.11-5.593.322C5.307 2.7 4.5 3.65 4.5 4.757V19.5a2.25 2.25 0 0 0 2.25 2.25h10.5a2.25 2.25 0 0 0 2.25-2.25V4.757c0-1.108-.806-2.057-1.907-2.185A48.507 48.507 0 0 0 12 2.25Z" />`,
	// Kalkulacija proizvoda — Cog
	"proizvodnja_kalkulacijaproizvoda": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />`,
	// Sitan inventar — ClipboardDocumentCheck
	"proizvodnja_sitaninventar": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />`,
	// Izveštaji — ChartBar
	"proizvodnja_izvestaji": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />`,
	// Kraj poslovne godine — Calendar
	"proizvodnja_krajposlovnegodine": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />`,
	// ---------------------------------------------------------
	// Podmeni: Osnovna sredstva
	// Matični podaci osnovnih sredstava — ClipboardDocumentList (klipbord - matični podaci)
	"os_maticnipodaci": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z" />`,
	// Uknjižavanje, promene — ArrowsRightLeft (promene)
	"os_uknjizavanjepromene": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />`,
	// Obračun amortizacije — Calculator
	"os_obracunamortizacije": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 15.75V18m-7.5-6.75h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V13.5Zm0 2.25h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V18Zm2.498-6.75h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V13.5Zm0 2.25h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V18Zm2.504-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5Zm0 2.25h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V18Zm2.498-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5ZM8.25 6h7.5v2.25h-7.5V6ZM12 2.25c-1.892 0-3.758.11-5.593.322C5.307 2.7 4.5 3.65 4.5 4.757V19.5a2.25 2.25 0 0 0 2.25 2.25h10.5a2.25 2.25 0 0 0 2.25-2.25V4.757c0-1.108-.806-2.057-1.907-2.185A48.507 48.507 0 0 0 12 2.25Z" />`,
	// Obračun poreske amortizacije — ReceiptPercent (procenat/porez)
	"os_obracunporeskeamortizacije": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 14.25l6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0c1.1.128 1.907 1.077 1.907 2.185ZM9.75 9h.008v.008H9.75V9Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm4.125 4.5h.008v.008h-.008V13.5Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />`,
	// Amortizacione grupe — Folder (grupa)
	"os_amortizacionegrupe": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 12.75V12A2.25 2.25 0 0 1 4.5 9.75h15A2.25 2.25 0 0 1 21.75 12v.75m-8.69-6.44-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z" />`,
	// Amortizacione podgrupe — FolderOpen (otvorena grupa)
	"os_amortizacionepodgrupe": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 9.776c.112-.017.227-.026.344-.026h15.812c.117 0 .232.009.344.026m-16.5 0a2.25 2.25 0 0 0-1.883 2.542l.857 6a2.25 2.25 0 0 0 2.227 1.932H19.05a2.25 2.25 0 0 0 2.227-1.932l.857-6a2.25 2.25 0 0 0-1.883-2.542m-16.5 0V6A2.25 2.25 0 0 1 6 3.75h3.879a1.5 1.5 0 0 1 1.06.44l2.122 2.12a1.5 1.5 0 0 0 1.06.44H18A2.25 2.25 0 0 1 20.25 9v.776" />`,
	// Organizacione jedinice — BuildingOffice2
	"os_organizacionejedinice": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Z" />`,
	// Rukovaoci — User (korisnik)
	"os_rukovaoci": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />`,
	// Objekti — MapPin (lokacija)
	"os_objekti": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" />`,
	// Konta nabavke i ispravke — Banknotes (novac/konta)
	"os_kontanabavke": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 0 0-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 0 1-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 0 0 3 15h-.75M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm3 0h.008v.008H18V10.5Zm-12 0h.008v.008H6V10.5Z" />`,
	// Koeficijenti revalorizacije — Calculator (kalkulacija)
	"os_koeficijentirevalorizacije": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 15.75V18m-7.5-6.75h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V13.5Zm0 2.25h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V18Zm2.498-6.75h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V13.5Zm0 2.25h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V18Zm2.504-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5Zm0 2.25h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V18Zm2.498-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5ZM8.25 6h7.5v2.25h-7.5V6ZM12 2.25c-1.892 0-3.758.11-5.593.322C5.307 2.7 4.5 3.65 4.5 4.757V19.5a2.25 2.25 0 0 0 2.25 2.25h10.5a2.25 2.25 0 0 0 2.25-2.25V4.757c0-1.108-.806-2.057-1.907-2.185A48.507 48.507 0 0 0 12 2.25Z" />`,
	// Stanje osnovnih sredstava — Cube (roba/imovina)
	"os_stanjeosnovnihsredstava": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />`,
	// Kartica osnovnog sredstva — Identification (identifikaciona kartica)
	"os_karticaosnovnogsredstva": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Zm6-10.125a1.875 1.875 0 1 1-3.75 0 1.875 1.875 0 0 1 3.75 0Zm1.294 6.336a6.721 6.721 0 0 1-3.17.789 6.721 6.721 0 0 1-3.168-.789 3.376 3.376 0 0 1 6.338 0Z" />`,
	// Izveštaji — ChartBar
	"os_izvestaji": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />`,
	// Kraj poslovne godine — Calendar
	"os_krajposlovnegodine": `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />`,
}
var MainMenuIconsSVG = map[string]string{

	// 🔵 1. ŠIFARNICI — Blue → Indigo
	"sifarnici": `<defs>
		<linearGradient id="grad-sifarnici" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#60a5fa"/>
			<stop offset="100%" stop-color="#6366f1"/>
		</linearGradient>
	</defs>
	<path stroke="url(#grad-sifarnici)" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.593 3.322c1.1.128 1.907 1.077 1.907 2.185V21L12 17.25 4.5 21V5.507c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0Z" />`,

	// 🟢 2. FINANSIJSKO — Emerald → Teal
	"finansijsko": `<defs>
		<linearGradient id="grad-finansijsko" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#34d399"/>
			<stop offset="100%" stop-color="#14b8a6"/>
		</linearGradient>
	</defs>
	<path stroke="url(#grad-finansijsko)" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.75 15.75V18m-7.5-6.75h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V13.5Zm0 2.25h.008v.008H8.25v-.008Zm0 2.25h.008v.008H8.25V18Zm2.498-6.75h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V13.5Zm0 2.25h.007v.008h-.007v-.008Zm0 2.25h.007v.008h-.007V18Zm2.504-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5Zm0 2.25h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V18Zm2.498-6.75h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V13.5ZM8.25 6h7.5v2.25h-7.5V6ZM12 2.25c-1.892 0-3.758.11-5.593.322C5.307 2.7 4.5 3.65 4.5 4.757V19.5a2.25 2.25 0 0 0 2.25 2.25h10.5a2.25 2.25 0 0 0 2.25-2.25V4.757c0-1.108-.806-2.057-1.907-2.185A48.507 48.507 0 0 0 12 2.25Z" />`,

	// 🟡 3. KOMERCIJALNO — Amber → Orange
	"komercijalno": `<defs>
		<linearGradient id="grad-komercijalno" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#fbbf24"/>
			<stop offset="100%" stop-color="#f97316"/>
		</linearGradient>
	</defs>
	<path stroke="url(#grad-komercijalno)" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 0 0-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 0 0-16.536-1.84M7.5 14.25 5.106 5.272M6 20.25a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Zm12.75 0a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Z" />`,

	// 🟠 4. ROBNO — Orange → Pink
	"robno": `<defs>
		<linearGradient id="grad-robno" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#fb923c"/>
			<stop offset="100%" stop-color="#ec4899"/>
		</linearGradient>
	</defs>
	<path stroke="url(#grad-robno)" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />`,

	// 🟣 5. BLAGAJNA — Violet → Fuchsia
	"blagajna": `<defs>
		<linearGradient id="grad-blagajna" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#a78bfa"/>
			<stop offset="100%" stop-color="#d946ef"/>
		</linearGradient>
	</defs>
	<path stroke="url(#grad-blagajna)" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />`,

	// 🔴 6. PROIZVODNJA — Red → Rose
	"proizvodnja": `<defs>
		<linearGradient id="grad-proizvodnja" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#ea8e8e"/>
			<stop offset="100%" stop-color="#f43f5e"/>
		</linearGradient>
	</defs>
	<path stroke="url(#grad-proizvodnja)" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 21h19.5M3 21V9l6-3v3l6-3v3l6-3v15M6 21v-3h2.25v3M10.5 21v-3h2.25v3M15 21v-3h2.25v3M12 3v3M15 3v3M9 3v3" />`,

	// 🩵 7. OSNOVNA SREDSTVA — Cyan → Blue
	"osnovnasredstva": `<defs>
		<linearGradient id="grad-osnovnasredstva" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#22d3ee"/>
			<stop offset="100%" stop-color="#3b82f6"/>
		</linearGradient>
	</defs>
	<path stroke="url(#grad-osnovnasredstva)" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />`,
}

// IconGradientGroups — mapiranje ikonica u grupe
var IconGradientGroups = map[string]string{
	// Šifarnici
	"sifarnici_partneri":       "sifarnici",
	"sifarnici_vrstenaloga":    "sifarnici",
	"sifarnici_vrstedokumenta": "sifarnici",
	"sifarnici_vrsteporknjige": "sifarnici",
	"sifarnici_vrsteevpdv":     "sifarnici",
	"sifarnici_oj":             "sifarnici",
	"sifarnici_mtroska":        "sifarnici",
	"sifarnici_drzave":         "sifarnici",
	"sifarnici_opstine":        "sifarnici",
	"sifarnici_mesta":          "sifarnici",
	"sifarnici_banke":          "sifarnici",
	"sifarnici_bankeizv":       "sifarnici",
	"sifarnici_sifplizv":       "sifarnici",

	// Finansijsko
	"fin_kontniplan":       "finansijsko",
	"fin_nalozi":           "finansijsko",
	"fin_promet":           "finansijsko",
	"fin_saldakonta":       "finansijsko",
	"fin_kompenzacije":     "finansijsko",
	"fin_otvorenestavke":   "finansijsko",
	"fin_kompenzacije3":    "finansijsko",
	"fin_obracunkamate":    "finansijsko",
	"fin_bilansi":          "finansijsko",
	"fin_ravnoteza":        "finansijsko",
	"fin_knjizenje":        "finansijsko",
	"fin_import":           "finansijsko",
	"fin_azurirajkonta":    "finansijsko",
	"fin_obrada":           "finansijsko",
	"fin_poreskeknjige":    "finansijsko",
	"fin_popdvprijave":     "finansijsko",
	"fin_evidencijaepp":    "finansijsko",
	"fin_ucitavanjeizvoda": "finansijsko",
	"fin_dnevnikknjizenja": "finansijsko",
	"fin_tipovianalitike":  "finansijsko",

	// Komercijalno
	"komercijala_zahtevzanabavku":          "komercijalno",
	"komercijala_zahtevzaponudu":           "komercijalno",
	"komercijala_porudzbenicedobavljacima": "komercijalno",
	"komercijala_prijemnicamaterijala":     "komercijalno",
	"komercijala_profakture":               "komercijalno",
	"komercijala_nalozizaisporuku":         "komercijalno",
	"komercijala_otpremanjerobe":           "komercijalno",
	// Robno
	"robno_robnadokumenta":       "robno",
	"robno_artikli":              "robno",
	"robno_stanjeartikli":        "robno",
	"robno_izvestajiulazizlaz":   "robno",
	"robno_prometartikli":        "robno",
	"robno_jedinicemere":         "robno",
	"robno_robnegrupe":           "robno",
	"robno_robnepodgrupe":        "robno",
	"robno_krajposlovnegodine":   "robno",
	"robno_mestaisporuke":        "robno",
	"robno_komercijalisti":       "robno",
	"robno_poreskestope":         "robno",
	"robno_vrstedokumenata":      "robno",
	"robno_cenovnik":             "robno",
	"robno_tipoviknjiznihpisama": "robno",
	"robno_magacini":             "robno",
	"robno_magacinkonto":         "robno",
	"robno_komercijalnipodaci":   "robno",

	// Blagajna
	"blagajna_dnevnikblagajne":     "blagajna",
	"blagajna_kontiranjeblagajne":  "blagajna",
	"blagajna_knjizenjeblagajne":   "blagajna",
	"blagajna_valute":              "blagajna",
	"blagajna_kursnalista":         "blagajna",
	"blagajna_apoeni":              "blagajna",
	"blagajna_tipoviblagajne":      "blagajna",
	"blagajna_vrstestavkiblagajne": "blagajna",
	"blagajna_dodatnaknjizenja":    "blagajna",
	"blagajna_putninalozi":         "blagajna",
	"blagajna_dnevnice":            "blagajna",
	"blagajna_tipovitroskova":      "blagajna",
	"blagajna_prevoznasredstva":    "blagajna",
	"blagajna_specifikacijaapoena": "blagajna",

	// Proizvodnja
	"proizvodnja_maticnipodacinormativa": "proizvodnja",
	"proizvodnja_kontatroskova":          "proizvodnja",
	"proizvodnja_radninalog":             "proizvodnja",
	"proizvodnja_unosdokumenta":          "proizvodnja",
	"proizvodnja_pregleddokumenta":       "proizvodnja",
	"proizvodnja_specifikacijadokumenta": "proizvodnja",
	"proizvodnja_knjizenjedokumenta":     "proizvodnja",
	"proizvodnja_prometartiklikonta":     "proizvodnja",
	"proizvodnja_stanjepoluproizvoda":    "proizvodnja",
	"proizvodnja_trenutnostanjezaliha":   "proizvodnja",
	"proizvodnja_pregledstanjacena":      "proizvodnja",
	"proizvodnja_obracunskakalkulacija":  "proizvodnja",
	"proizvodnja_kalkulacijaproizvoda":   "proizvodnja",
	"proizvodnja_sitaninventar":          "proizvodnja",
	"proizvodnja_izvestaji":              "proizvodnja",
	"proizvodnja_krajposlovnegodine":     "proizvodnja",

	// Osnovna sredstva
	"os_maticnipodaci":              "osnovnasredstva",
	"os_uknjizavanjepromene":        "osnovnasredstva",
	"os_obracunamortizacije":        "osnovnasredstva",
	"os_obracunporeskeamortizacije": "osnovnasredstva",
	"os_amortizacionegrupe":         "osnovnasredstva",
	"os_amortizacionepodgrupe":      "osnovnasredstva",
	"os_organizacionejedinice":      "osnovnasredstva",
	"os_rukovaoci":                  "osnovnasredstva",
	"os_objekti":                    "osnovnasredstva",
	"os_kontanabavke":               "osnovnasredstva",
	"os_koeficijentirevalorizacije": "osnovnasredstva",
	"os_stanjeosnovnihsredstava":    "osnovnasredstva",
	"os_karticaosnovnogsredstva":    "osnovnasredstva",
	"os_izvestaji":                  "osnovnasredstva",
	"os_krajposlovnegodine":         "osnovnasredstva",
}

// GradientDefs — definicije gradienta
var GradientDefs = map[string]string{
	"sifarnici": `<linearGradient id="grad-sifarnici" x1="0%" y1="0%" x2="100%" y2="100%">
		<stop offset="0%" stop-color="#acdded"/>
		<stop offset="100%" stop-color="#279ac7"/>
	</linearGradient>`,

	"finansijsko": `<linearGradient id="grad-finansijsko" x1="0%" y1="0%" x2="100%" y2="100%">
		<stop offset="0%" stop-color="#34d399"/>
		<stop offset="100%" stop-color="#14b8a6"/>
	</linearGradient>`,

	"komercijalno": `<linearGradient id="grad-komercijalno" x1="0%" y1="0%" x2="100%" y2="100%">
		<stop offset="0%" stop-color="#fbbf24"/>
		<stop offset="100%" stop-color="#f97316"/>
	</linearGradient>`,

	"robno": `<linearGradient id="grad-robno" x1="0%" y1="0%" x2="100%" y2="100%">
		<stop offset="0%" stop-color="#fb923c"/>
		<stop offset="100%" stop-color="#ec4899"/>
	</linearGradient>`,

	"blagajna": `<linearGradient id="grad-blagajna" x1="0%" y1="0%" x2="100%" y2="100%">
		<stop offset="0%" stop-color="#a78bfa"/>
		<stop offset="100%" stop-color="#d946ef"/>
	</linearGradient>`,
	"proizvodnja": `<linearGradient id="grad-proizvodnja" x1="0%" y1="0%" x2="100%" y2="100%">
		<stop offset="0%" stop-color="#ea8e8e"/>
		<stop offset="100%" stop-color="#f43f5e"/>
	</linearGradient>`,
	"osnovnasredstva": `<linearGradient id="grad-osnovnasredstva" x1="0%" y1="0%" x2="100%" y2="100%">
		<stop offset="0%" stop-color="#22d3ee"/>
		<stop offset="100%" stop-color="#3b82f6"/>
	</linearGradient>`,
}

// GetIconGradient vraća SVG ikonicu sa gradient-om (ako pripada grupi)
func GetIconGradient(name string) string {
	svg, ok := IconSVG[name]
	if !ok {
		return ""
	}

	group, hasGroup := IconGradientGroups[name]
	if !hasGroup {
		return svg // vrati original
	}

	defs, ok := GradientDefs[group]
	if !ok {
		return svg
	}

	// Zameni ili dodaj stroke
	gradientStroke := `stroke="url(#grad-` + group + `)"`
	svg = replaceStroke(svg, gradientStroke)

	return `<defs>` + defs + `</defs>` + svg
}

// replaceStroke zamenjuje stroke="..." ili dodaje stroke u path
func replaceStroke(svg, newStroke string) string {
	if strings.Contains(svg, `stroke="`) {
		re := regexp.MustCompile(`stroke="[^"]*"`)
		return re.ReplaceAllString(svg, newStroke)
	}
	return strings.Replace(svg, `<path `, `<path `+newStroke+` `, 1)
}
