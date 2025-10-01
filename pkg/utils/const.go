package utils

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
	ParseFormErrMsg    = "Parsiranje forme nije uspelo."
	GetIdFromUrlErrMsg = "Neuspešno preuzimanje ID-a iz URL-a."
	FormDecodeErrMsg   = "Neuspešno dekodiranje forme."
	ValidationErrMsg   = "Greške prilkom validacije"
	SaveDataErrMsg     = "Greška prilikom upisa podataka"
	SaveDataOkMsg      = "Uspešno upisani podaci"
	DeleteDataErrMsg   = "Greška prilikom brisanja podataka. Greska: %s"
	DeleteDataOkMsg    = "Uspešno obrisani podaci"
	InvalidIdErrMsg    = "Invalid ID"
	ReadDataErrMsg     = "Greška prilikom čitanja podataka"
	ReadDataOkMsg      = "Uspešno učitani podaci"
	RenderTemplateErr  = "Error rendering template"
	InvalidIDErrMsg    = "Invalid ID provided in request"
)

// constants for styling of inputs
const (
	ClassInputNumericDisabled = "h-6 px-1 py-1 text-sm text-right border border-blue-400 rounded focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-200 disabled:text-gray-500 disabled:cursor-not-allowed "
	ClassInputTextEnabled     = "h-6 text-sm rounded border border-blue-400 focus:bg-blue-100 focus:border-blue-500 focus:ring-blue-500 px-1 border text-left"
	ClassInputTextDisabled    = "h-6 px-1 text-sm flex-1 min-w-0 text-left rounded border border-blue-400 focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-200 disabled:text-gray-500 disabled:cursor-not-allowed"
	ClassLabel                = "h-6 text-sm text-gray-700 whitespace-nowrap flex-shrink-0"
	ClassSelect               = "h-6 text-sm text-black rounded border border-blue-400 focus:border-blue-500 focus:ring-blue-500 px-1"
	//ClassButton               = "h-6 text-sm px-1 bg-gray-200 hover:bg-gray-300 rounded border border-red-400 flex items-center justify-center flex-shrink-0"
	ClassButton              = "h-6 text-sm bg-gray-200 hover:bg-gray-300 px-1 rounded border border-blue-400 flex items-center justify-center flex-shrink-0"
	ClassSearchInput         = "search-input border rounded px-2 h-7 py-1 text-xs w-64 pl-8"
	PatternLettersAndNumbers = "[A-Za-z0-9]+"
	PatternNumbers           = "[0-9]+"
)
