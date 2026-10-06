package robno

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"helia/config"
	tmpl "helia/frontend/templates"
	tmpl_rep_rob "helia/frontend/templates/reports/robno"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Tab indexes of the "Robna dokumenta" option (same order as robnoDokumentaTabs).
const (
	robnoDokumentaTabUnos = iota
	robnoDokumentaTabPregled
	robnoDokumentaTabSpecifikacije
	robnoDokumentaTabKontiranje
	robnoDokumentaTabPrepis
	robnoDokumentaTabPrikazUkupneObrade
	robnoDokumentaTabPrikazNaloga
	robnoDokumentaTabPrikazDokumenataUNalogu
	robnoDokumentaTabPrikazDokumenataPooperateru
	// robnoDokumentaTabCount is the number of tabs; robnoDokumentaSubTabs must have this length.
	robnoDokumentaTabCount
)

const (
	robnoDokumentaURLPrefix = "/api/robno-dokumenta"
	robnoDokumentaTitle     = "Robna dokumenta"

	// Tab 1 - Unos dokumenta
	robnoDokumentaUnosTitle     = "Unos dokumenta"
	robnoDokumentaUnosTableID   = "robno-dokumenta-unos-table"
	robnoDokumentaURLUnos       = robnoDokumentaURLPrefix + "/unos"
	robnoDokumentaURLUnosStampa = robnoDokumentaURLUnos + "/stampa"
	// next nalog / header data of an existing nalog / confirmation of the save, like the "Nalozi"
	// (fnal) screen.
	robnoDokumentaURLNextNalog = robnoDokumentaURLPrefix + "/nextnalog"
	robnoDokumentaURLNalogData = robnoDokumentaURLPrefix + "/nalog-data"
	robnoDokumentaURLConfirm   = robnoDokumentaURLPrefix + "/confirm-addupdate"

	// TODO (temporary): the "Fakture veleprodaje" screen (the RobnoFakture template) has no handler
	// yet. The button "Fakture veleprodaje (preview)" of the "Unos dokumenta" tab and these two routes
	// only render the template so that it can be reviewed in the browser: the save route answers
	// success without saving anything, so that the screen switches from the header to the entry of the
	// stavke (robnoFaktureAfterHeaderSave). Remove the button, the routes and FakturePreview /
	// FakturePreviewSave together with the two table ids when the handler of the
	// fakture is written.
	robnoDokumentaURLFakturePreview     = robnoDokumentaURLPrefix + "/fakture-preview"
	robnoDokumentaURLFakturePreviewSave = robnoDokumentaURLFakturePreview + "/save"
	robnoDokumentaFaktureStavkeTableID  = "robno-fakture-stavke-table"
	robnoDokumentaFaktureAvansiTableID  = "robno-fakture-avansi-table"

	// Štampa fakture (the report RobnoStampaFaktura, the legacy ROB_RPT_STAMPA_FAKTURA).
	robnoDokumentaStampaFakturaTitle = "Štampa fakture"
	robnoDokumentaURLStampaFaktura   = robnoDokumentaURLPrefix + "/fakture/stampa"

	// Štampa izvozne fakture (the report RobnoStampaFakturaIzvoz, the fakture in a foreign valuta, the
	// legacy PR_RPT_FAKTURA_OTPIZV).
	robnoDokumentaStampaFakturaIzvozTitle = "Invoice"
	robnoDokumentaURLStampaFakturaIzvoz   = robnoDokumentaURLPrefix + "/fakture-izvoz/stampa"

	// Štampa maloprodajnog računa (the report RobnoStampaFakturaMP, the documents of the group DIR, the
	// legacy ROB_RPT_STAMPA_FAKTURA_MP).
	robnoDokumentaStampaFakturaMPTitle = "Štampa maloprodajnog računa"
	robnoDokumentaURLStampaFakturaMP   = robnoDokumentaURLPrefix + "/fakture-mp/stampa"

	// Štampa fakture usluga (the report RobnoStampaFakturaUsluge, the groups PRE and FUR, the legacy
	// ROB_RPT_STAMPA_FAKTURA_USLUGE2).
	robnoDokumentaStampaFakturaUslugeTitle = "Štampa fakture usluga"
	robnoDokumentaURLStampaFakturaUsluge   = robnoDokumentaURLPrefix + "/fakture-usluge/stampa"

	// Štampa avansnog računa (the report RobnoStampaFakturaAvansni, the groups ARA and ARU, the legacy
	// ROB_RPT_STAMPA_ARA).
	robnoDokumentaStampaFakturaAvansniTitle = "Štampa avansnog računa"
	robnoDokumentaURLStampaFakturaAvansni   = robnoDokumentaURLPrefix + "/fakture-avansne/stampa"

	// Štampa popisa (the report RobnoStampaPopis, the documents of the vrsta dokumenta 101).
	robnoDokumentaStampaPopisTitle = "Štampa popisa"
	robnoDokumentaURLStampaPopis   = robnoDokumentaURLPrefix + "/popis/stampa"

	// Štampa kalkulacije veleprodaje (the report RobnoStampaKalkulacija, the prijemni listovi of the
	// group PLT, the legacy RPT_ROB_KALKULACIJA).
	robnoDokumentaStampaKalkulacijaTitle = "Kalkulacija veleprodaje"
	robnoDokumentaURLStampaKalkulacija   = robnoDokumentaURLPrefix + "/kalkulacija/stampa"
	robnoDokumentaVrdPopis               = 101

	// Štampa opšteg dokumenta (the report RobnoStampaOpstiDokument, the groups OPD, POT, KOL and FIN,
	// the legacy RPT_OPDSTAMPA).
	robnoDokumentaStampaOpstiDokumentTitle = "Štampa opšteg dokumenta"
	robnoDokumentaURLStampaOpstiDokument   = robnoDokumentaURLPrefix + "/opsti-dokument/stampa"
	// robnoDokumentaStampaFakturaFlds are the fields the "Štampaj" button of the "Štampa" sub-tab of
	// "Pregled dokumenta" sends: the selected document (the hidden rdokid) and the filter of the
	// sub-tab.
	robnoDokumentaStampaFakturaFlds = "rdokid,magaciniid,vrd,oddanal,dodanal"

	// Tab 2 - Pregled dokumenta
	robnoDokumentaPregledTitle   = "Pregled dokumenta"
	robnoDokumentaPregledTableID = "robno-dokumenta-pregled-table"
	robnoDokumentaURLPregled     = robnoDokumentaURLPrefix + "/pregled"

	// Tab 2, sub-tab 1 - Štampa
	robnoDokumentaPregledStampaTitle    = "Pregled dokumenata - štampa"
	robnoDokumentaPregledStampaTableID  = "robno-dokumenta-pregled-stampa-table"
	robnoDokumentaURLPregledStampa      = robnoDokumentaURLPregled + "/stampa"
	robnoDokumentaURLPregledStampaPrint = robnoDokumentaURLPregledStampa + "/print"

	// Tab 2, sub-tab 2 - eFaktura
	robnoDokumentaEFakturaTitle          = "Pregled dokumenata - eFaktura"
	robnoDokumentaEFakturaTableID        = "robno-dokumenta-efaktura-table"
	robnoDokumentaURLEFaktura            = robnoDokumentaURLPregled + "/efaktura"
	robnoDokumentaURLEFakturaPrint       = robnoDokumentaURLEFaktura + "/print"
	robnoDokumentaURLEFakturaStatus      = robnoDokumentaURLEFaktura + "/status"
	robnoDokumentaURLEFakturaPosalji     = robnoDokumentaURLEFaktura + "/posalji"
	robnoDokumentaURLEFakturaProveri     = robnoDokumentaURLEFaktura + "/proveri-status"
	robnoDokumentaURLEFakturaOtkazi      = robnoDokumentaURLEFaktura + "/otkazi"
	robnoDokumentaURLEFakturaStorniraj   = robnoDokumentaURLEFaktura + "/storniraj"
	robnoDokumentaURLEFakturaStornirajPE = robnoDokumentaURLEFaktura + "/storniraj-pe"

	// robnoDokumentaGrupeEFaktura is the default of the "Grupe dokumenata" filter of the eFaktura
	// sub-tab: the groups of documents (dokvrsta.grpdok) that can be sent as an eFaktura (the legacy
	// screen shows this list in the filter).
	robnoDokumentaGrupeEFaktura = "FAK,FRP,PRE,FUR,FPR,FMA,FZR,ARA,ARU,KPK"

	// Tab 3 - Specifikacije dokumenta
	robnoDokumentaSpecifikacijeTitle     = "Specifikacije dokumenta"
	robnoDokumentaSpecifikacijeTableID   = "robno-dokumenta-specifikacije-table"
	robnoDokumentaURLSpecifikacije       = robnoDokumentaURLPrefix + "/specifikacije"
	robnoDokumentaURLSpecifikacijeStampa = robnoDokumentaURLSpecifikacije + "/stampa"
	// robnoDokumentaSpecifikacijaFaktureTitle is the title of the report the "Specifikacija" print of
	// the tab renders (the legacy RPT_ROB_SPECIFIKACIJA_fakture). rep.ReportTitle renders the report
	// name literally, so it is the title as the legacy report shows it.
	robnoDokumentaSpecifikacijaFaktureTitle = "SPECIFIKACIJA EKSTERNIH RACUNA"

	// Tab 4 - Kontiranje dokumenata
	robnoDokumentaKontiranjeTitle   = "Kontiranje dokumenata"
	robnoDokumentaURLKontiranje     = robnoDokumentaURLPrefix + "/kontiranje"
	robnoDokumentaKontiranjeBtnID   = "kontiranje-knjizi-btn"
	robnoDokumentaRavnotezaBtnID    = "kontiranje-ravnoteza-btn"
	robnoDokumentaOznaciBtnID       = "kontiranje-oznaci-btn"
	robnoDokumentaKontiranjeTableID = "robno-dokumenta-kontiranje-table"

	// Tab 4, sub-tab 1 - Knjiženje dokumenata
	robnoDokumentaKnjizenjeTitle     = "Kontiranje dokumenata - knjiženje"
	robnoDokumentaKnjizenjeTableID   = "robno-dokumenta-knjizenje-table"
	robnoDokumentaURLKnjizenje       = robnoDokumentaURLKontiranje + "/knjizenje"
	robnoDokumentaURLKnjizenjeKnjizi = robnoDokumentaURLKnjizenje + "/knjizi"
	robnoDokumentaURLKnjizenjeRavnot = robnoDokumentaURLKnjizenje + "/ravnoteza"

	// Tab 4, sub-tab 2 - Pregled proknjiženih / neproknjiženih dokumenata
	robnoDokumentaKontiranjePregledTitle   = "Kontiranje dokumenata - pregled"
	robnoDokumentaKontiranjePregledTableID = "robno-dokumenta-kontiranje-pregled-table"
	robnoDokumentaURLKontiranjePregled     = robnoDokumentaURLKontiranje + "/pregled"
	robnoDokumentaURLOznaciNeproknjizene   = robnoDokumentaURLKontiranjePregled + "/oznaci-neproknjizene"
	robnoDokumentaKontiranjePregledFlds    = "magaciniid,tipdok,odnaloga,donaloga,oddokum,dodokum,oddanal,dodanal,proknjizen"

	// Tab 4, sub-tab 3 - Pregled proknjiženih / neproknjiženih dokumenata po magacinima
	robnoDokumentaPoMagacinimaTitle   = "Kontiranje dokumenata - pregled po magacinima"
	robnoDokumentaPoMagacinimaTableID = "robno-dokumenta-kontiranje-magacini-table"
	robnoDokumentaURLPoMagacinima     = robnoDokumentaURLKontiranje + "/po-magacinima"
	robnoDokumentaPoMagacinimaFlds    = "magaciniid,odnaloga,donaloga,oddokum,dodokum,oddanal,dodanal,proknjizen"

	// Tab 4, printing of the sub-tabs (the print preview is built together with the reports).
	robnoDokumentaURLKontiranjeStampa = robnoDokumentaURLKontiranje + "/stampa"

	// robnoDokumentaOdNaloga / robnoDokumentaDoNaloga are the default range of the "broj naloga"
	// and "broj dokumenta" selections of "Kontiranje dokumenata" and robnoDokumentaDanaSelekcije
	// the number of days the default range of the "datum naloga" looks back (the defaults of the
	// legacy screen).
	robnoDokumentaOdNaloga      = "0"
	robnoDokumentaDoNaloga      = "999999"
	robnoDokumentaDanaSelekcije = 7

	// Tab 5 - Prepis dokumenta
	robnoDokumentaPrepisTitle     = "Prepis dokumenta"
	robnoDokumentaPrepisTableID   = "robno-dokumenta-prepis-table"
	robnoDokumentaURLPrepis       = robnoDokumentaURLPrefix + "/prepis"
	robnoDokumentaURLPrepisStampa = robnoDokumentaURLPrepis + "/stampa"

	// Tab 6 - Prikaz ukupne obrade
	robnoDokumentaUkupnaObradaTitle     = "Prikaz ukupne obrade"
	robnoDokumentaUkupnaObradaTableID   = "robno-dokumenta-ukupna-obrada-table"
	robnoDokumentaURLUkupnaObrada       = robnoDokumentaURLPrefix + "/ukupna-obrada"
	robnoDokumentaURLUkupnaObradaStampa = robnoDokumentaURLUkupnaObrada + "/stampa"

	// Tab 7 - Prikaz naloga
	robnoDokumentaPrikazNalogaTitle     = "Prikaz naloga"
	robnoDokumentaPrikazNalogaTableID   = "robno-dokumenta-prikaz-naloga-table"
	robnoDokumentaURLPrikazNaloga       = robnoDokumentaURLPrefix + "/prikaz-naloga"
	robnoDokumentaURLPrikazNalogaStampa = robnoDokumentaURLPrikazNaloga + "/stampa"

	// Tab 8 - Prikaz dokumenata u nalogu
	robnoDokumentaUNaloguTitle     = "Prikaz dokumenata u nalogu"
	robnoDokumentaUNaloguTableID   = "robno-dokumenta-u-nalogu-table"
	robnoDokumentaURLUNalogu       = robnoDokumentaURLPrefix + "/u-nalogu"
	robnoDokumentaURLUNaloguStampa = robnoDokumentaURLUNalogu + "/stampa"

	// Tab 9 - Prikaz dokumenata po operateru
	robnoDokumentaPooperateruTitle     = "Prikaz dokumenata po operateru"
	robnoDokumentaPooperateruTableID   = "robno-dokumenta-po-operateru-table"
	robnoDokumentaURLPooperateru       = robnoDokumentaURLPrefix + "/po-operateru"
	robnoDokumentaURLPooperateruStampa = robnoDokumentaURLPooperateru + "/stampa"

	// Common element ids.
	robnoDokumentaSearchInputID = "robno-dokumenta-search-input"
	robnoDokumentaObradaBtnID   = "obrada-btn"
	// robnoDokumentaPrintFlds is the print (štampa) contract of every tab of the option: the ids of
	// the fields of the parameter panels of the tabs (the panel of a tab renders only its own fields,
	// the ones of the other tabs are skipped by openPrintWithParams). The parameters of the print of
	// a tab are then read from the request into the single domain.RobnoDokumentaParams structure.
	robnoDokumentaPrintFlds = "tipdok,vrd,magaciniid,nalog,danal,datob,opis," +
		"grupedokumenata,datumstatusa," +
		"odvrd,dovrd,odnaloga,donaloga,oddokum,dodokum,oddanal,dodanal," +
		"oddatob,dodatob,chkpodatumunaloga,chkpodatumuobrade,chkpooperateru,oper," +
		"proknjizen,oznacineproknjizenim,stampajsamozbir,tipspecifikacije"
	// robnoDokumentaInfoMessageID is the staging element of the messages of the actions of the
	// tabs (it is emitted by the layout of the option).
	robnoDokumentaInfoMessageID = "info-message"
	// robnoDokumentaInfoMessageDialogID is the staging element the error dialogs of the tabs are
	// rendered into (it is emitted by RobnoDokumentaShell, so every tab of the option has it): the
	// missing session and every failed query of a tab answer with the dialog utils.RenderDialogOK
	// renders into it.
	robnoDokumentaInfoMessageDialogID = "robnodokumenta-infomessage"

	// Tab 1 - Unos dokumenta, element ids.
	robnoDokumentaContentID  = "robno-dokumenta-content"
	robnoDokumentaUnosFormID = "robno-dokumenta-unos-form"
	// The id of the save button ends with "btn-save" so that the double click on a row of the grid
	// (handleDblClickNalogSelection) can find and click it, like the "Nalozi" screen.
	robnoDokumentaSaveBtnID      = "robno-dokumenta-btn-save"
	robnoDokumentaNoviNalogBtnID = "robno-dokumenta-novi-nalog-btn"

	// Confirm dialog of the save of the header. The dialog is rendered into the #dialog-confirm
	// staging element (the same one the "Nalozi" screen uses) and the response of the dialog is
	// handled by handleDialogResponse.
	robnoDokumentaConfirmStagingID = "dialog-confirm"
	robnoDokumentaConfirmDialogID  = "dialog-robno-dokumenta-unos"
	robnoDokumentaDialogResponseID = "addupdate-dialog"

	// robnoDokumentaEntityType is the entitytype of the locks of the robni nalozi (rnal) and
	// robnoDokumentaSourceTipdok marks the reload of the grid after the change of the vrsta naloga
	// (only the grid is then rendered, the header of the form is kept).
	robnoDokumentaEntityType   = "rnal"
	robnoDokumentaSourceTipdok = "tipdok"

	// robnoDokumentaSourceBtn is the request source the buttons rendered by components.Button send
	// (hx-headers of the component). A request with that source renders the whole screen the button
	// opens, so it must not be treated as a data request of a grid (common.IsDataRequest).
	robnoDokumentaSourceBtn = "btn"

	// hxValsRobnoDokumentaUnos sends the header of the nalog and the grid filters of the tab.
	hxValsRobnoDokumentaUnos = `js:{
		"tipdok": document.getElementById("tipdok")?.value,
		"vrd": document.getElementById("vrd")?.value,
		"magaciniid": document.getElementById("magaciniid")?.value,
		"nalog": document.getElementById("nalog")?.value,
		"danal": document.getElementById("danal")?.value,
		"datob": document.getElementById("datob")?.value,
		"opis": document.getElementById("opis")?.value,
	}`
)

// hxValsRobnoDokumentaTipdok is sent when the vrsta naloga changes. The source parameter tells the
// handler to render only the grid (the header of the form is kept) and the hidden #nalog-trigger
// element then asks for the next broj naloga of the new vrsta naloga.
const hxValsRobnoDokumentaTipdok = `js:{"tipdok": document.getElementById("tipdok")?.value, ` +
	`"vrd": document.getElementById("vrd")?.value, ` +
	`"magaciniid": document.getElementById("magaciniid")?.value, ` +
	`"source": "tipdok"}`

// hxValsRobnoDokumentaNalog sends the vrsta naloga and the broj naloga of the header to the check
// of an existing nalog (GET /api/robno-dokumenta/nalog-data).
const hxValsRobnoDokumentaNalog = `js:{"tipdok": document.getElementById("tipdok")?.value, ` +
	`"nalog": document.getElementById("nalog")?.value}`

// The declarations of the "Fakture veleprodaje" screen (the templates RobnoFakture and
// RobnoFaktureDialog of frontend/templates/robno/robnadokumenta.templ). The template package cannot
// import this package, so every url, element id and hx-vals string the markup of the screen uses is
// declared here: the urls and the ids reach the templates as domain.RobnoFaktureUI (see
// robnoFaktureUI) and the two hx-vals strings of the header of the nalog as plain arguments, while
// robnoFaktureButtonsFor builds the buttons of the screen with the same constants.
const (
	// URLs of the screen. TODO (temporary): the handler and the routes of the fakture are not written
	// yet; the urls are the ones the screen will use (the prefix of the other tabs of "Robna
	// dokumenta").
	robnoFaktureURL           = robnoDokumentaURLPrefix + "/fakture"
	robnoFaktureSaveURL       = robnoFaktureURL + "/save"
	robnoFaktureNoviURL       = robnoFaktureURL + "/novi"
	robnoFaktureDeleteURL     = robnoFaktureURL + "/brisi"
	robnoFaktureStavkaSaveURL = robnoFaktureURL + "/stavka/save"
	robnoFaktureStavkaDelURL  = robnoFaktureURL + "/stavka/brisi"
	robnoFaktureArtikalSearch = "/api/promet/searchbutton"
	robnoFakturePartnerSearch = "/api/partneri/searchbutton"
	// TODO: the endpoint of the search of the robni dokumenti is not written yet (neither is the one of
	// robnoFakturePartnerSearch); the button of the broj dokumenta already sends the request here.
	robnoFaktureDokumentSearch = "/api/robno-dokumenta/searchbutton"

	// Ids of the two collapsible controls of the screen. They are given to the script of the screen
	// (RobnoFaktureScript) and to the buttons (robnoFaktureButtonsFor), so that every id is written in
	// one place only.
	robnoFaktureHeaderPanelID = "robno-fakture-header-panel"
	robnoFaktureStavkePanelID = "robno-fakture-stavke-panel"

	// robnoDokumentaFaktureDialogStagingID is the element of the "Unos dokumenta" tab (the template
	// RobnoDokumentaMain) the dialog of the "Fakture veleprodaje" screen is rendered into and
	// robnoFaktureDialogID the id of the dialog itself (the one closeDialog hides). The staging element
	// is the hx-target of the button that opens the dialog and the dialog id is the IdDialog of the
	// "Zatvori" button of the dialog.
	robnoDokumentaFaktureDialogStagingID = "robno-fakture-dialog-staging"
	robnoFaktureDialogID                 = "robno-fakture-dialog"
)

// robnoFaktureUI groups the urls and the element ids of the "Fakture veleprodaje" screen.
func robnoFaktureUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoFaktureHeaderPanelID,
		StavkePanelID:     robnoFaktureStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoFaktureDialogID,
		DialogStagingID:   robnoDokumentaFaktureDialogStagingID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
	}
}

// robnoFaktureButtonsFor builds the buttons of the "Fakture veleprodaje" screen.
func robnoFaktureButtonsFor() domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:            "robno-fakture-sacuvaj",
			LabelText:     "Sačuvaj",
			Icon:          "save",
			BtnClass:      common.ClassButton,
			HxActionURL:   robnoFaktureSaveURL,
			HxRequestType: "POST",
			// Only the fields of the header are sent (the read only values of the strips are disabled,
			// so the browser does not send them; the handler reads them from the dokument/kupac).
			HxInclude: "#" + robnoFaktureHeaderPanelID + " [data-panel-fields]",
			// The handler answers with the standard JSON response; the state of the two controls is
			// switched by the script of the screen (RobnoFaktureScript) after the request. The same
			// script also works when the handler swaps the whole tab (hx-target = robnoFaktureUI's
			// ContentID), because the swapped markup carries the state of the saved dokument.
			HxSwap:               "none",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoFaktureHeaderPanelID, robnoFaktureStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-fakture-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoFaktureHeaderPanelID, robnoFaktureStavkePanelID},
		},
		New: domain.Button{
			Id:               "robno-fakture-novi",
			LabelText:        "Novi dok.",
			Icon:             "add",
			BtnClass:         common.ClassButton,
			HxActionURL:      robnoFaktureNoviURL,
			HxRequestType:    "GET",
			HxTarget:         "#" + robnoDokumentaContentID,
			HxSwap:           "innerHTML",
			HxOnAfterRequest: "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{
				robnoFaktureHeaderPanelID, robnoFaktureStavkePanelID},
		},
		Delete: domain.Button{
			Id:            "robno-fakture-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassButton,
			HxActionURL:   robnoFaktureDeleteURL,
			HxRequestType: "DELETE",
			HxInclude:     "#" + robnoFaktureHeaderPanelID + " [data-panel-fields]",
			HxSwap:        "none",
		},
		Back: domain.Button{
			Id:        "robno-fakture-nazad",
			LabelText: "Nazad",
			Icon:      "back",
			BtnClass:  common.ClassButton,
			// TODO: the route of the "Fakture veleprodaje" menu entry
			HxActionURL:   "/api/robna-dokumenta",
			HxRequestType: "GET",
			HxTarget:      "#main-content",
			HxSwap:        "innerHTML",
		},
		SaveStavka: domain.Button{
			Id:            "robno-fakture-stavka-sacuvaj",
			LabelText:     "Sačuvaj",
			Icon:          "save",
			BtnClass:      common.ClassButton,
			HxActionURL:   robnoFaktureStavkaSaveURL,
			HxRequestType: "POST",
			HxInclude:     "#" + robnoFaktureStavkePanelID + " [data-panel-fields]",
			HxTarget:      "#" + robnoFaktureStavkePanelID,
			HxSwap:        "none",
		},
		ModifyStavka: domain.Button{
			Id:        "robno-fakture-stavka-izmeni",
			LabelText: "Izmeni",
			Icon:      "refresh",
			BtnClass:  common.ClassButton,
		},
		DeleteStavka: domain.Button{
			Id:            "robno-fakture-stavka-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassButton,
			HxActionURL:   robnoFaktureStavkaDelURL,
			HxRequestType: "DELETE",
			HxInclude:     "#" + robnoFaktureStavkePanelID + " [data-panel-fields]",
			HxTarget:      "#" + robnoFaktureStavkePanelID,
			HxSwap:        "none",
		},
	}
}

// Sub-tab indexes of the "Pregled dokumenta" tab (the two sub-tabs that are implemented).
const (
	robnoDokumentaSubTabPregledStampa = iota
	robnoDokumentaSubTabPregledEFaktura
)

// Sub-tab indexes of the "Kontiranje dokumenata" tab.
const (
	robnoDokumentaSubTabKontiranjeKnjizenje = iota
	robnoDokumentaSubTabKontiranjePregled
	robnoDokumentaSubTabKontiranjePoMagacinima
)

// Values of the "Proknjiženi / Neproknjiženi dokumenti" radio buttons of the last two sub-tabs of
// "Kontiranje dokumenata" and the default values of the ranges of the tab, like the legacy screen
// shows them.
const (
	robnoDokumentaProknjizen   = "D"
	robnoDokumentaNeproknjizen = "N"
)

// hxValsRobnoDokumentaPregledStampa sends the filters of the "Štampa" sub-tab.
const hxValsRobnoDokumentaPregledStampa = `js:{
	"magaciniid": document.getElementById("magaciniid")?.value,
	"vrd": document.getElementById("vrd")?.value,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
}`

// hxValsRobnoDokumentaPregledEFaktura sends the filters of the "eFaktura" sub-tab.
const hxValsRobnoDokumentaPregledEFaktura = `js:{
	"grupedokumenata": document.getElementById("grupedokumenata")?.value,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
	"datumstatusa": document.getElementById("datumstatusa")?.value,
}`

// hxValsRobnoDokumentaKnjizenje sends the parameters of the "Knjiženje dokumenata" sub-tab.
const hxValsRobnoDokumentaKnjizenje = `js:{
	"tipdok": document.getElementById("tipdok")?.value,
	"vrd": document.getElementById("vrd")?.value,
	"magaciniid": document.getElementById("magaciniid")?.value,
	"odnaloga": document.getElementById("odnaloga")?.value,
	"donaloga": document.getElementById("donaloga")?.value,
	"oddokum": document.getElementById("oddokum")?.value,
	"dodokum": document.getElementById("dodokum")?.value,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
}`

// hxValsRobnoDokumentaKontiranjePregled sends the parameters and the state selected with the radio
// buttons of the second sub-tab of "Kontiranje dokumenata".
const hxValsRobnoDokumentaKontiranjePregled = `js:{
	"magaciniid": document.getElementById("magaciniid")?.value,
	"tipdok": document.getElementById("tipdok")?.value,
	"odnaloga": document.getElementById("odnaloga")?.value,
	"donaloga": document.getElementById("donaloga")?.value,
	"oddokum": document.getElementById("oddokum")?.value,
	"dodokum": document.getElementById("dodokum")?.value,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
	"proknjizen": document.querySelector("input[name='proknjizen']:checked")?.value,
	"oznacineproknjizenim": document.getElementById("oznacineproknjizenim")?.checked ? "true" : "",
}`

// hxValsRobnoDokumentaPoMagacinima sends the parameters of the third sub-tab of "Kontiranje
// dokumenata" (it has no vrsta naloga filter, its rows are grouped by the magacin).
const hxValsRobnoDokumentaPoMagacinima = `js:{
	"magaciniid": document.getElementById("magaciniid")?.value,
	"odnaloga": document.getElementById("odnaloga")?.value,
	"donaloga": document.getElementById("donaloga")?.value,
	"oddokum": document.getElementById("oddokum")?.value,
	"dodokum": document.getElementById("dodokum")?.value,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
	"proknjizen": document.querySelector("input[name='proknjizen']:checked")?.value,
}`

// hxValsRobnoDokumentaPrikazNaloga sends the parameters of the "Prikaz naloga" tab, including the
// state of the three checkboxes that enable the date and operator filters.
const hxValsRobnoDokumentaPrikazNaloga = `js:{
	"magaciniid": document.getElementById("magaciniid")?.value,
	"odvrd": document.getElementById("odvrd")?.value,
	"dovrd": document.getElementById("dovrd")?.value,
	"odnaloga": document.getElementById("odnaloga")?.value,
	"donaloga": document.getElementById("donaloga")?.value,
	"chkpodatumunaloga": document.getElementById("chkpodatumunaloga")?.checked,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
	"chkpodatumuobrade": document.getElementById("chkpodatumuobrade")?.checked,
	"oddatob": document.getElementById("oddatob")?.value,
	"dodatob": document.getElementById("dodatob")?.value,
	"chkpooperateru": document.getElementById("chkpooperateru")?.checked,
	"oper": document.getElementById("oper")?.value,
}`

// hxValsRobnoDokumentaUNalogu sends the parameters of the "Prikaz dokumenata u nalogu" tab. The
// parameters of the tab are the same selection of nalozi as the "Prikaz naloga" tab, including the
// state of the three checkboxes that enable the date and operator filters.
const hxValsRobnoDokumentaUNalogu = hxValsRobnoDokumentaPrikazNaloga

// hxValsRobnoDokumentaPooperateru sends the parameters of the "Prikaz dokumenata po operateru" tab.
// The panel of the tab has no magacin and no vrsta naloga selection (the legacy screen has none), so
// it sends the range of the broj naloga and the state of the three checkboxes with their fields.
const hxValsRobnoDokumentaPooperateru = `js:{
	"odnaloga": document.getElementById("odnaloga")?.value,
	"donaloga": document.getElementById("donaloga")?.value,
	"chkpodatumunaloga": document.getElementById("chkpodatumunaloga")?.checked,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
	"chkpodatumuobrade": document.getElementById("chkpodatumuobrade")?.checked,
	"oddatob": document.getElementById("oddatob")?.value,
	"dodatob": document.getElementById("dodatob")?.value,
	"chkpooperateru": document.getElementById("chkpooperateru")?.checked,
	"oper": document.getElementById("oper")?.value,
}`

// hxValsRobnoDokumentaSpecifikacije sends the parameters of the "Specifikacije dokumenta" tab: the
// magacin, the ranges of the vrste naloga za knjiženje, of the broj naloga and of the broj dokumenta,
// the vrsta dokumenta, the range of the dates and the state of the print.
const hxValsRobnoDokumentaSpecifikacije = `js:{
	"magaciniid": document.getElementById("magaciniid")?.value,
	"odvrd": document.getElementById("odvrd")?.value,
	"dovrd": document.getElementById("dovrd")?.value,
	"vrd": document.getElementById("vrd")?.value,
	"odnaloga": document.getElementById("odnaloga")?.value,
	"donaloga": document.getElementById("donaloga")?.value,
	"oddokum": document.getElementById("oddokum")?.value,
	"dodokum": document.getElementById("dodokum")?.value,
	"oddanal": document.getElementById("oddanal")?.value,
	"dodanal": document.getElementById("dodanal")?.value,
	"stampajsamozbir": document.getElementById("stampajsamozbir")?.checked,
	"tipspecifikacije": document.querySelector("input[name='tipspecifikacije']:checked")?.value,
}`

type RobnoDokumentaHandler struct {
	translator *i18n.Service
	service    robnosvc.RobnoDokumentaService
	cfg        config.Config
	lm         *middleware.LockMiddleware
	ls         *middleware.LockService
	tabs       domain.TabData
	// subTabs holds the sub-tabs of every tab, indexed like robnoDokumentaTabs (a tab without
	// sub-tabs has an empty set and its sub-tab bar is then not rendered).
	subTabs []domain.TabData
	// btnSave and btnNoviNalog are the buttons of the header of the "Unos dokumenta" tab (they are
	// defined once, like the "Nalozi" screen does).
	btnSave      domain.Button
	btnNoviNalog domain.Button
	// TODO (temporary): btnFakture opens the new "Fakture veleprodaje" screen (RobnoFakture).
	btnFakture domain.Button
}

// NewRobnoDokumentaHandler creates the handler of the "Robna dokumenta" option.
func NewRobnoDokumentaHandler(s robnosvc.RobnoDokumentaService, cfg config.Config, lm *middleware.LockMiddleware, ls *middleware.LockService, translator *i18n.Service) *RobnoDokumentaHandler {
	h := &RobnoDokumentaHandler{
		translator: translator,
		service:    s,
		cfg:        cfg,
		lm:         lm,
		ls:         ls,
		tabs:       robnoDokumentaTabs(translator),
		subTabs:    robnoDokumentaSubTabs(translator),
	}
	h.setHandlerFieldValues()
	return h
}

// setHandlerFieldValues defines the buttons of the header of the "Unos dokumenta" tab.
func (h *RobnoDokumentaHandler) setHandlerFieldValues() {
	h.btnSave = domain.Button{
		Id:        robnoDokumentaSaveBtnID,
		IdDialog:  robnoDokumentaDialogResponseID,
		IsVisible: true,
		LabelText: "Snimi nalog",
		// The response of the confirm is handled by handleDialogResponse (the SaveButton component
		// emits the call with the IdDialog as the argument).
		HxActionURL:      robnoDokumentaURLConfirm,
		HxTarget:         "#" + robnoDokumentaConfirmStagingID,
		HxSwap:           "innerHTML",
		HxOnAfterRequest: "handleDialogResponse",
		HxRequestType:    "POST",
		BtnClass:         common.ClassSaveButton,
		HxInclude:        "#" + robnoDokumentaUnosFormID,
	}
	h.btnNoviNalog = domain.Button{
		Id:               robnoDokumentaNoviNalogBtnID,
		IsVisible:        true,
		LabelText:        "Novi nalog",
		HxActionURL:      robnoDokumentaURLNextNalog,
		HxRequestType:    "GET",
		HxInclude:        "#tipdok",
		HxOnAfterRequest: "handleNextNalogResponse",
		HxSwap:           "none",
		BtnClass:         common.ClassNewButton,
	}
	// TODO (temporary): opens the new "Fakture veleprodaje" screen (RobnoFakture) for review; see
	// robnoDokumentaURLFakturePreview. The screen is rendered as a dialog into the staging element of
	// the "Unos dokumenta" tab (tmpl_robno.RobnoDokumentaFaktureDialogStagingID).
	h.btnFakture = domain.Button{
		Id:            "robno-dokumenta-fakture-btn",
		IsVisible:     true,
		LabelText:     "Fakture veleprodaje (preview)",
		HxActionURL:   robnoDokumentaURLFakturePreview,
		HxRequestType: "GET",
		HxTarget:      "#" + robnoDokumentaFaktureDialogStagingID,
		HxSwap:        "innerHTML",
		BtnClass:      common.ClassButton,
	}
}

// RobnoDokumentaMain renders the "Unos dokumenta" tab (header of the nalog and grid of the nalozi).
func (h *RobnoDokumentaHandler) RobnoDokumentaMain(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabUnos)
	subTabs := h.subTabsFor(robnoDokumentaTabUnos)
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	params := h.unosParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaUnosTitle, robnoDokumentaUnosTableID, h.service.GetUnosDokumentaTableFields(), robnoDokumentaURLUnos, robnoDokumentaURLUnos, pageSize, page, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaUnosTitle, robnoDokumentaURLUnos, false, false, false)
	tbl.HxVals = hxValsRobnoDokumentaUnos
	tbl.Pagination.HxVals = hxValsRobnoDokumentaUnos
	// Double click on a row of the grid loads the nalog into the header of the tab (like the
	// "Nalozi" screen does with its "knjiženje" tab).
	tbl.FuncDblClick = "handleDblClickNalogSelection(this)"
	tbl.FuncClick = "selectRow(this)"
	if !h.getUnosDokumenta(c, &tbl, params, page, pageSize) {
		return
	}
	// Data request from the grid (search / paging) and the reload on the change of the vrsta
	// naloga (source=tipdok): only the grid is rendered, the header of the form is kept.
	if common.IsDataRequest(c) || c.Query("source") == robnoDokumentaSourceTipdok {
		utils.RenderContent(c, tbl)
		return
	}

	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	total := domain.RobnoDokumentaTotal{}
	if err := h.service.GetUnosDokumentaTotal(ctx, &total); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	payload, err := h.unosPayload(ctx, params)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLUnos, "#"+robnoDokumentaUnosTableID, hxValsRobnoDokumentaUnos)
	if err := tmpl_robno.RobnoDokumentaMain(robnoFaktureUI(), hxValsRobnoDokumentaTipdok, hxValsRobnoDokumentaNalog, tabs, subTabs, tbl, tipdokValues, vrstaDokumentaValues, magValues, total, payload, h.btnSave, h.btnNoviNalog, h.btnFakture, search, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// FakturePreview renders the "Fakture veleprodaje" screen as a dialog (TODO: temporary preview).
func (h *RobnoDokumentaHandler) FakturePreview(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}

	// Both grids of the screen are empty: their queries belong to the handler of the fakture (TODO).
	tbl := common.SetTableBasicData("Stavke fakture", robnoDokumentaFaktureStavkeTableID, h.service.GetFaktureStavkeTableFields(), "", robnoDokumentaURLFakturePreview, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaFaktureStavkeTableID, robnoDokumentaURLFakturePreview, false, false, false)
	if common.IsDataRequest(c) && c.Request.Header.Get("X-Request-Source") != robnoDokumentaSourceBtn {
		utils.RenderContent(c, tbl)
		return
	}
	avansiTbl := common.SetTableBasicData("Avansi", robnoDokumentaFaktureAvansiTableID, h.service.GetFaktureAvansiTableFields(), "", robnoDokumentaURLFakturePreview, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&avansiTbl, robnoDokumentaFaktureAvansiTableID, robnoDokumentaURLFakturePreview, false, false, false)

	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	header := domain.RobnoFaktureHeaderView{}
	header.Snimljen = c.Query("snimljen") == "true"
	btns := robnoFaktureButtonsFor()
	btns.Save.HxActionURL = robnoDokumentaURLFakturePreviewSave
	btns.Back.HxActionURL = ""
	btns.Back.HxRequestType = ""
	btns.Back.HxOnClick = "closeDialog"
	btns.Back.HxOnClickArg = []any{robnoFaktureDialogID}
	// The "Zatvori" button of the title bar of the dialog (CloseButton calls closeDialog with the dialog id).
	btnClose := domain.Button{
		Id:       "robno-fakture-dialog-close",
		IdDialog: robnoFaktureDialogID,
		BtnClass: common.ClassDialogCloseButton,
	}

	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLFakturePreview, "#"+robnoDokumentaFaktureStavkeTableID, "")
	// TODO: the combos of the valute, of the sistemi PDV and of the ZIRP računa of the screen (the
	// last two are combos in the template).
	if err := tmpl_robno.RobnoFaktureDialog(robnoFaktureUI(), tbl, avansiTbl, header, domain.RobnoFaktureStavka{}, vrstaDokumentaValues, nil, nil, nil, btns, btnClose, search, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// FakturePreviewSave answers the preview save without saving anything (TODO: temporary).
func (h *RobnoDokumentaHandler) FakturePreviewSave(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusOK, true, nil, "Pregled: podaci fakture nisu sačuvani (handler faktura još nije implementiran)")
}

// robnoDokumentaStampaParams reads the selection of the print of the robni dokumenti (common to the
// prints of every vrsta dokumenta): the one of the source query of the legacy reports - the vrsta
// naloga (tipdok), the groups of the vrste dokumenta (grupedokumenata), the ranges of the broj naloga
// (odnaloga/donaloga), of the broj dokumenta (oddokum/dodokum) and of the vrsta dokumenta (odvrd/dovrd)
// and the magacin (magaciniid) - and the one of the "Štampa" sub-tab of "Pregled dokumenta": the
// selected document (rdokid), the vrsta dokumenta (vrd) and the range of the datum naloga.
func robnoDokumentaStampaParams(c *gin.Context) domain.RobnoStampaFakturaParams {
	return domain.RobnoStampaFakturaParams{
		Tipdok:          c.Query("tipdok"),
		GrupeDokumenata: c.Query("grupedokumenata"),
		OdNaloga:        c.Query("odnaloga"),
		DoNaloga:        c.Query("donaloga"),
		OdDokum:         c.Query("oddokum"),
		DoDokum:         c.Query("dodokum"),
		MagaciniID:      common.StringToInt(c.Query("magaciniid")),
		OdVrd:           c.Query("odvrd"),
		DoVrd:           c.Query("dovrd"),
		RdokID:          int64(common.StringToInt(c.Query("rdokid"))),
		Vrd:             c.Query("vrd"),
		OdDanal:         c.Query("oddanal"),
		DoDanal:         c.Query("dodanal"),
	}
}

// robnoDokumentaStampaReportParams returns the parameters of the report of a printed robni dokument
// (common to the prints of every vrsta dokumenta): the izdavalac with its logo, the user and the title.
func robnoDokumentaStampaReportParams(firma domain.RobnoStampaFakturaFirmaDto, userSession *domain.UserSession, title string) domain.ReportParameters {
	return domain.ReportParameters{
		Orientation: "portrait",
		CompanyName: firma.Naziv,
		Adress:      firma.Adresa,
		Postcode:    firma.Pobro,
		City:        firma.Mesto,
		PIB:         firma.Pib,
		MatBroj:     firma.Matbr,
		SifDel:      firma.Sifdel,
		TekRac:      firma.Tekrac,
		Telefon:     firma.Tel,
		CompanyLogo: robnosvc.RobnoStampaLogo(firma.Logo),
		UserName:    userSession.UserName,
		ReportName:  title,
		God:         userSession.SelectedGod,
	}
}

// StampaFakturaPrint prints the fakture of the selection (the report RobnoStampaFaktura, the legacy
// ROB_RPT_STAMPA_FAKTURA), one faktura per page. The selection is read by robnoDokumentaStampaParams;
// the groups of the vrste dokumenta are by default the group of the vrsta dokumenta odvrd, like the
// legacy report. A single faktura is printed with odnaloga = donaloga and oddokum = dodokum (or with
// the rdokid of the document).
func (h *RobnoDokumentaHandler) StampaFakturaPrint(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	params := robnoDokumentaStampaParams(c)
	// The print of the "Štampa" sub-tab without a selected document and without a vrsta dokumenta
	// prints the documents of the groups of the fakture (the groups of the eFaktura), not every robni
	// dokument of the period.
	if params.RdokID == 0 && params.GrupeDokumenata == "" && params.OdVrd == "" && params.Vrd == "" {
		params.GrupeDokumenata = robnoDokumentaGrupeEFaktura
	}
	fakture, firma, err := h.service.GetStampaFaktura(ctx, params)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaFakturaTitle)
	if err := tmpl_rep_rob.RobnoStampaFaktura(repParams, fakture, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// StampaFakturaMP prints the maloprodajni računi of the selection (the report RobnoStampaFakturaMP, the
// legacy ROB_RPT_STAMPA_FAKTURA_MP), one račun per page. The selection is read by
// robnoDokumentaStampaParams; without a selected document the računi of the group DIR are printed.
func (h *RobnoDokumentaHandler) StampaFakturaMP(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	fakture, firma, err := h.service.GetStampaFakturaMP(ctx, robnoDokumentaStampaParams(c))
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaFakturaMPTitle)
	if err := tmpl_rep_rob.RobnoStampaFakturaMP(repParams, fakture, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// StampaFakturaUsluge prints the fakture usluga of the selection (the report RobnoStampaFakturaUsluge,
// the legacy ROB_RPT_STAMPA_FAKTURA_USLUGE2), one faktura per page. The selection is read by
// robnoDokumentaStampaParams.
func (h *RobnoDokumentaHandler) StampaFakturaUsluge(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	fakture, firma, err := h.service.GetStampaFakturaUsluge(ctx, robnoDokumentaStampaParams(c))
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaFakturaUslugeTitle)
	if err := tmpl_rep_rob.RobnoStampaFakturaUsluge(repParams, fakture, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// StampaFakturaAvansni prints the avansni računi of the selection (the report
// RobnoStampaFakturaAvansni, the legacy ROB_RPT_STAMPA_ARA), one račun per page. The selection is read
// by robnoDokumentaStampaParams.
func (h *RobnoDokumentaHandler) StampaFakturaAvansni(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	fakture, firma, err := h.service.GetStampaFakturaAvansni(ctx, robnoDokumentaStampaParams(c))
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaFakturaAvansniTitle)
	if err := tmpl_rep_rob.RobnoStampaFakturaAvansni(repParams, fakture, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// StampaFakturaIzvoz prints the izvozne fakture of the selection (the report RobnoStampaFakturaIzvoz,
// the legacy PR_RPT_FAKTURA_OTPIZV), one faktura per page. The selection is read by
// robnoDokumentaStampaParams; "komercopis" (the legacy ipCBOX_KOMERCOPIS) adds the komercijalni opis
// of the artikli to their naziv.
func (h *RobnoDokumentaHandler) StampaFakturaIzvoz(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	fakture, firma, err := h.service.GetStampaFakturaIzvoz(ctx, robnoDokumentaStampaParams(c), c.Query("komercopis") == "true")
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaFakturaIzvozTitle)
	if err := tmpl_rep_rob.RobnoStampaFakturaIzvoz(repParams, fakture, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// StampaPopis prints the popisi of the selection (the report RobnoStampaPopis, the documents of the
// vrsta dokumenta 101), one popis per page. The selection is read by robnoDokumentaStampaParams;
// without a selected document and without a vrsta dokumenta it is the popisi of the selection.
func (h *RobnoDokumentaHandler) StampaPopis(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	params := robnoDokumentaStampaParams(c)
	if params.RdokID == 0 && params.Vrd == "" {
		params.Vrd = fmt.Sprintf("%d", robnoDokumentaVrdPopis)
	}
	popisi, firma, err := h.service.GetStampaPopis(ctx, params)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaPopisTitle)
	if err := tmpl_rep_rob.RobnoStampaPopis(repParams, popisi, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// StampaKalkulacija prints the kalkulacije veleprodaje (prijemni listovi) of the selection (the report
// RobnoStampaKalkulacija, the legacy RPT_ROB_KALKULACIJA), one kalkulacija per page in landscape: a
// domestic document with the porezi of the document of the dobavljač, a document in a foreign valuta
// with the valuta and the kurs. The selection is read by robnoDokumentaStampaParams.
func (h *RobnoDokumentaHandler) StampaKalkulacija(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	kalkulacije, firma, err := h.service.GetStampaKalkulacija(ctx, robnoDokumentaStampaParams(c))
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaKalkulacijaTitle)
	repParams.Orientation = "landscape"
	if err := tmpl_rep_rob.RobnoStampaKalkulacija(repParams, kalkulacije, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// StampaOpstiDokument prints the opšti dokumenti of the selection (the report RobnoStampaOpstiDokument,
// the legacy RPT_OPDSTAMPA), one document per page, each by the group of its vrsta dokumenta. The
// selection is read by robnoDokumentaStampaParams.
func (h *RobnoDokumentaHandler) StampaOpstiDokument(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	dokumenti, firma, err := h.service.GetStampaOpstiDokument(ctx, robnoDokumentaStampaParams(c))
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := robnoDokumentaStampaReportParams(firma, userSession, robnoDokumentaStampaOpstiDokumentTitle)
	if err := tmpl_rep_rob.RobnoStampaOpstiDokument(repParams, dokumenti, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// GetNextNalog returns the next free broj naloga of a vrsta naloga as JSON.
func (h *RobnoDokumentaHandler) GetNextNalog(c *gin.Context) {
	nextNalog, err := h.service.GetNextNalog(c.Request.Context(), c.Query("tipdok"))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgGetData)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "nalog": nextNalog})
}

// GetNalogData returns the header of an existing robni nalog as JSON.
func (h *RobnoDokumentaHandler) GetNalogData(c *gin.Context) {
	ctx := c.Request.Context()
	tipdok := c.Query("tipdok")
	nalog := common.StringToInt(c.Query("nalog"))
	rnal, err := h.service.GetByTipdokNalog(ctx, tipdok, nalog)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgGetData)
		return
	}
	if rnal.RnalID == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"nalog":      rnal.Nalog,
		"danal":      nullDateHtml(rnal.Danal),
		"datob":      nullDateHtml(rnal.Datob),
		"opis":       rnal.Opis,
		"magaciniid": rnal.MagaciniID.Int64,
	})
}

// ConfirmUnosDokumenta opens the confirm dialog of the save of the header.
func (h *RobnoDokumentaHandler) ConfirmUnosDokumenta(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobnoDokumentaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	action := common.ActionAdd
	existing, err := h.service.GetByTipdokNalog(ctx, params.Tipdok, common.StringToInt(params.Nalog))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgReadData)
		return
	}
	if existing.RnalID != 0 {
		// The nalog exists: it must not be held by another user before it is continued.
		if err := h.lm.VerifyLock(c, robnoDokumentaEntityType, existing.RnalID); err != nil {
			common.WriteJSONResponse(c, http.StatusConflict, false, []domain.FieldError{}, common.ErrMsgStatusConflict)
			return
		}
		action = common.ActionUpdate
	}
	if fieldErrors := h.service.ValidateUnosDokumenta(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}

	dialogTitle := "Novi nalog"
	msg := []string{h.translator.Message(`Otvaranje novog naloga?`)}
	if action == common.ActionUpdate {
		dialogTitle = "Nastavak knjiženja naloga"
		msg = []string{
			h.translator.Message(`Nastavak knjiženja naloga?`),
			fmt.Sprintf("%s: %s", h.translator.Message(`Vrsta naloga`), params.Tipdok),
			fmt.Sprintf("%s: %s", h.translator.Message(`Broj naloga`), params.Nalog),
		}
	}

	dialog := domain.Dialog{
		Id:            robnoDokumentaConfirmDialogID,
		Title:         dialogTitle,
		OkText:        "Da",
		CancelText:    "Ne",
		SaveText:      "Da",
		HxTarget:      "#" + robnoDokumentaConfirmStagingID,
		HxSwap:        "innerHTML",
		HxRequestType: "POST",
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	// The confirm button posts the header of the form to the create or to the update of the nalog.
	btnSacuvaj := domain.Button{
		Id:               "btn-sacuvaj",
		IsVisible:        true,
		LabelText:        "Da",
		IdDialog:         dialog.Id,
		HxInclude:        "#" + robnoDokumentaUnosFormID,
		HxTarget:         "#" + robnoDokumentaContentID,
		HxSwap:           "none",
		BtnClass:         common.ClassSaveButton,
		HxRequestType:    "POST",
		HxActionURL:      robnoDokumentaURLUnos,
		HxOnAfterRequest: "closeDialog",
	}
	if action == common.ActionUpdate {
		btnSacuvaj.HxRequestType = "PUT"
		btnSacuvaj.HxActionURL = fmt.Sprintf("%s/%d", robnoDokumentaURLUnos, existing.RnalID)
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Ne",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	tmpl.DialogConfirm(msg, dialog, btnClose, btnSacuvaj, btnCancel, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// SaveUnosDokumenta inserts the header of a new robni nalog (rnal) and locks it.
func (h *RobnoDokumentaHandler) SaveUnosDokumenta(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return
	}
	var params domain.RobnoDokumentaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	if fieldErrors := h.service.ValidateUnosDokumenta(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rnalID, err := h.service.CreateUnosDokumenta(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgSaveData+" error:"+err.Error())
		return
	}
	// The header is locked for the user (the documents of the nalog are entered in the following
	// requests, like the "stavke" of a financial nalog).
	if err := h.ls.Lock(ctx, robnoDokumentaEntityType, rnalID, userSession.UserName); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgLockFailed)
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

// UpdateUnosDokumenta saves the header of an existing robni nalog (rnal).
func (h *RobnoDokumentaHandler) UpdateUnosDokumenta(c *gin.Context) {
	ctx := c.Request.Context()
	rnalID, err := utils.GetInt64FromParameterRequest(c, "id")
	if err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgInvalidID)
		return
	}
	var params domain.RobnoDokumentaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, []domain.FieldError{}, common.ErrMsgFormDecode)
		return
	}
	if fieldErrors := h.service.ValidateUnosDokumenta(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.UpdateUnosDokumenta(ctx, rnalID, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgUpdate+err.Error())
		return
	}
	common.WriteJSONResponse(c, http.StatusOK, true, nil, common.OkMsgSaveData)
}

// nullDateHtml renders a nullable date of the header for an HTML date input.
func nullDateHtml(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(common.HtmlLayout)
}

// unosParams reads the header of the nalog and the filters of the tab from the request.
func (h *RobnoDokumentaHandler) unosParams(c *gin.Context) domain.RobnoDokumentaParams {
	return domain.RobnoDokumentaParams{
		Tipdok:     c.Query("tipdok"),
		Vrd:        c.Query("vrd"),
		Nalog:      c.Query("nalog"),
		Danal:      c.Query("danal"),
		Datob:      c.Query("datob"),
		Opis:       c.Query("opis"),
		MagaciniID: common.StringToInt(c.Query("magaciniid")),
		SearchText: c.Query("query"),
	}
}

// unosPayload prepares the form of the tab (next broj naloga and today's dates).
func (h *RobnoDokumentaHandler) unosPayload(ctx context.Context, params domain.RobnoDokumentaParams) (domain.RobnoDokumentaParams, error) {
	payload := params
	if payload.Nalog == "" {
		nextNalog, err := h.service.GetNextNalog(ctx, params.Tipdok)
		if err != nil {
			return payload, err
		}
		payload.Nalog = fmt.Sprintf("%d", nextNalog)
	}
	today := h.businessToday(ctx)
	if payload.Danal == "" {
		payload.Danal = today
	}
	if payload.Datob == "" {
		payload.Datob = today
	}
	return payload, nil
}

// getUnosDokumenta fills the grid of the "Unos dokumenta" tab (totals, then the rows).
func (h *RobnoDokumentaHandler) getUnosDokumenta(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams, page, pageSize int) bool {
	for _, total := range []bool{true, false} {
		if err := h.service.GetUnosDokumenta(c.Request.Context(), tbl, total, page, pageSize, params); err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return false
		}
	}
	return true
}

// PregledDokumenata renders the "Pregled dokumenta" tab (its first sub-tab "Štampa").
func (h *RobnoDokumentaHandler) StampaDokumentaObrada(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabPregled)
	subTabs := common.SetActiveTab(h.subTabsFor(robnoDokumentaTabPregled), robnoDokumentaSubTabPregledStampa)
	params := h.pregledParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaPregledStampaTitle, robnoDokumentaPregledStampaTableID, h.service.GetDokumentaPreviewTableFields(), "", robnoDokumentaURLPregledStampa, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaPregledStampaTitle, robnoDokumentaURLPregledStampa, false, false, false)
	tbl.HasTotals = true
	// A click on a row selects the document whose faktura the "Štampaj" button prints (a second click
	// on it clears the selection: the button then prints the fakture of the whole filter).
	tbl.FuncClick = "robnoPregledSelectDokument(this)"
	if common.IsDataRequest(c) {
		currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetDokumentaPreview(ctx, &tbl, true, currentPage, pageSize, params, common.TipStampePreview); err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		if err := h.service.GetDokumentaPreview(ctx, &tbl, false, currentPage, pageSize, params, common.TipStampePreview); err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnObrada := h.obradaButton(robnoDokumentaURLPregledStampa, robnoDokumentaPregledStampaTableID, hxValsRobnoDokumentaPregledStampa)
	// The print of the sub-tab is the štampa fakture: the faktura of the selected row or the fakture of
	// the filter of the sub-tab.
	btnPrint := common.SetPrintButton("stampa-btn", "Štampaj", "fin_print", robnoDokumentaURLPregledStampaPrint, "GET", true, common.ClassPrintButton, robnoDokumentaStampaFakturaFlds)
	btnPrint.HxTarget = "#info-message"
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	params.OdDanal = time.Date(userSession.SelectedGod, time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(common.HtmlLayout)
	params.DoDanal = time.Date(userSession.SelectedGod, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC).Format(common.HtmlLayout)

	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLPregledStampa, "#"+robnoDokumentaPregledStampaTableID, hxValsRobnoDokumentaPregledStampa)
	tmpl_robno.RobnoDokumentaPregled(tabs, subTabs, tbl, magValues, vrstaDokumentaValues, params, btnObrada, btnPrint, search, h.translator).Render(ctx, c.Writer)
}

// StampaDokumentaPrint is the print of the "Štampa" sub-tab of "Pregled dokumenta": the print of the
func (h *RobnoDokumentaHandler) StampaDokumentaPrint(c *gin.Context) {
	h.stampaDokumenta(c)
}

// stampaDokumenta prints the robni dokumenti of the selection with the print of their vrsta dokumenta,
func (h *RobnoDokumentaHandler) stampaDokumenta(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dok, err := h.service.GetStampaDokument(c.Request.Context(), rdokID, c.Query("vrd"))
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if dok.Grpdok == "" {
		if rdokID == 0 && dok.Vrd == 0 {
			h.StampaFakturaPrint(c)
			return
		}
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, "Vrsta dokumenta nema grupu dokumenta: štampa nije moguća")
		return
	}
	stampa, found := h.robnoDokumentaStampe()[strings.ToUpper(strings.TrimSpace(dok.Grpdok))]
	if !found {
		// The legacy OTHER CASE: the group has no print.
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf("Za grupu dokumenta %s nema štampe", dok.Grpdok))
		return
	}
	stampa(c, dok)
}

// robnoDokumentaStampaFunc prints a robni dokument (or the documents of the selection) of a group of
// vrste dokumenta.
type robnoDokumentaStampaFunc func(c *gin.Context, dok domain.RobnoStampaDokumentDto)

// robnoDokumentaStampe returns the prints of the robni dokumenti by the group of the vrsta dokumenta
// (DOKVRSTA.GRPDOK), the cases of the legacy print button. The name in a case is the legacy (WinDev)
// report of the group: a group whose print is not translated yet answers with a message
// (robnoDokumentaStampaNijeImplementirana); to add a print, write its handler and put it in its case.
func (h *RobnoDokumentaHandler) robnoDokumentaStampe() map[string]robnoDokumentaStampaFunc {
	faktura := h.stampaFakturaDokumenta
	intrac := h.stampaInterniPrenos
	// ROB_RPT_STAMPA_FAKTURA_USLUGE2 (its variants of DODOZNFAK print with the same report).
	usluge := func(c *gin.Context, _ domain.RobnoStampaDokumentDto) { h.StampaFakturaUsluge(c) }
	// ROB_RPT_STAMPA_ARA (its variants of DODOZNFAK print with the same report).
	avansni := func(c *gin.Context, _ domain.RobnoStampaDokumentDto) { h.StampaFakturaAvansni(c) }
	opsti := func(c *gin.Context, _ domain.RobnoStampaDokumentDto) { h.StampaOpstiDokument(c) }
	return map[string]robnoDokumentaStampaFunc{
		// Kalkulacija veleprodaje, the prijemni listovi (RPT_ROB_KALKULACIJA; its variants of DODOZNFAK
		// print with the same report): the domestic and the foreign dobavljači.
		"PLT": func(c *gin.Context, _ domain.RobnoStampaDokumentDto) { h.StampaKalkulacija(c) },
		// Fakture.
		"FAK": faktura,
		"FRP": faktura,
		// Fakture maloprodaje.
		"DIR": func(c *gin.Context, _ domain.RobnoStampaDokumentDto) { h.StampaFakturaMP(c) },
		// Interni prenosi.
		"IRT": intrac,
		"IRS": intrac,
		"IRP": intrac,
		// Interni prenosi proizvodnje.
		"PPR": robnoDokumentaStampaNijeImplementirana("ROB_RPT_INTRAC_PROIZV", true),
		// Fakture usluga.
		"PRE": usluge,
		"FUR": usluge,
		// Zaduženje i razduženje CO.
		"FCO": robnoDokumentaStampaNijeImplementirana("ROB_RPT_STAMPA_ZADUZENJA_CO", false),
		"RCO": h.stampaRazduzenjaCO,
		// Popis (početno stanje).
		"POP": func(c *gin.Context, _ domain.RobnoStampaDokumentDto) { h.StampaPopis(c) },
		// Opšti dokumenti.
		"OPD": opsti,
		"POT": opsti,
		"KOL": opsti,
		"FIN": opsti,
		// Nivelacije veleprodaje.
		"NIV": robnoDokumentaStampaNijeImplementirana("RPT_ROB_NIVELACIJA", false),
		// Avansni računi.
		"ARA": avansni,
		"ARU": avansni,
		// Kalkulacije maloprodaje (the selected document).
		"KAL": robnoDokumentaStampaNijeImplementirana("RPT_ROB_KALKULACIJA_MP", true),
		// Prenosnice izlaz i ulaz.
		"PRI": robnoDokumentaStampaNijeImplementirana("RPT_ROB_PRENOSNICE", false),
		"PRU": robnoDokumentaStampaNijeImplementirana("RPT_ROB_PRENOSNICEULAZ", false),
		// Zaduženje gradilišta.
		"GRD": robnoDokumentaStampaNijeImplementirana("RPT_ROB_ZADGRADILISTA", false),
		// Profakture.
		"PRO": h.stampaProfaktura,
		// Popis tekuće godine (the selected document).
		"PTG": robnoDokumentaStampaNijeImplementirana("RPT_POPIS_TEKGOD", false),
		// Sitan inventar.
		"SIV": robnoDokumentaStampaNijeImplementirana("RPT_ROB_ZADUZ_SI", false),
		// Knjižna pisma.
		"KNO": h.stampaKnjiznoPismo,
		"KNZ": h.stampaKnjiznoPismo,
		// Nivelacije maloprodaje.
		"MNI": h.stampaNivelacijaMaloprodaje,
		// Fakture za robu (the selected document).
		"FZR": h.stampaFakturaZaRobu,
		// Knjižna odobrenja.
		"KPK": h.stampaKnjiznoOdobrenje,
		"KPF": h.stampaKnjiznoOdobrenje,
	}
}

// robnoDokumentaStampaIzvestaj returns the name of the legacy report of a document: with the custom
// report of the vrsta dokumenta (DOKVRSTA.DODOZNFAK) the variant "<izvestaj>_<dodoznfak>" (the legacy
// NoSpace(DOKVRSTA.DODOZNFAK)), else the report itself.
func robnoDokumentaStampaIzvestaj(izvestaj string, dok domain.RobnoStampaDokumentDto) string {
	if varijanta := strings.ReplaceAll(dok.Dodoznfak, " ", ""); varijanta != "" {
		return izvestaj + "_" + varijanta
	}
	return izvestaj
}

// robnoDokumentaStampaNijeImplementirana returns the print of a group whose legacy report is not
// translated yet: it answers with the name of the report (with its variant of the vrsta dokumenta when
// the legacy chooses the report by DODOZNFAK).
func robnoDokumentaStampaNijeImplementirana(izvestaj string, saVarijantom bool) robnoDokumentaStampaFunc {
	return func(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
		naziv := izvestaj
		if saVarijantom {
			naziv = robnoDokumentaStampaIzvestaj(izvestaj, dok)
		}
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf("Štampa dokumenta (%s, vrsta %d) još nije implementirana", naziv, dok.Vrd))
	}
}

// stampaFakturaDokumenta is the print of the fakture (groups FAK and FRP), the legacy case of Srbija
// with the tip proizvodnje 1: a faktura in the domestic valuta is ROB_RPT_STAMPA_FAKTURA (or its
// variant of DODOZNFAK, printed with the base report until the variant is translated), a faktura in a
// foreign valuta is PR_RPT_FAKTURA_OTPIZV. The legacy prints of the tip proizvodnje 2
// (PR_RPT_FAKTURA_OTP) and of Republika Srpska (ROB_RPT_STAMPA_FAKTURA_MED) are not translated.
func (h *RobnoDokumentaHandler) stampaFakturaDokumenta(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if dok.Devizni() {
		h.StampaFakturaIzvoz(c)
		return
	}
	h.StampaFakturaPrint(c)
}

// stampaInterniPrenos is the print of the interni prenosi (groups IRT, IRS and IRP): with the option
// "irrn" (the legacy CBOX_IR_RN) the računi ROB_RPT_STAMPA_INTRACFKT, else ROB_RPT_STAMPA_INTRAC (or
// its variant of DODOZNFAK).
func (h *RobnoDokumentaHandler) stampaInterniPrenos(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if c.Query("irrn") == "true" {
		robnoDokumentaStampaNijeImplementirana("ROB_RPT_STAMPA_INTRACFKT", false)(c, dok)
		return
	}
	robnoDokumentaStampaNijeImplementirana("ROB_RPT_STAMPA_INTRAC", true)(c, dok)
}

// stampaRazduzenjaCO is the print of the razduženje CO (group RCO): only a kontiran document is printed
// (ROB_RPT_STAMPA_RAZDUZENJA_CO).
func (h *RobnoDokumentaHandler) stampaRazduzenjaCO(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if dok.RdokID != 0 && !dok.Kontiran() {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, "Dokument nije kontiran!")
		return
	}
	robnoDokumentaStampaNijeImplementirana("ROB_RPT_STAMPA_RAZDUZENJA_CO", false)(c, dok)
}

// stampaProfaktura is the print of the profakture (group PRO): in the domestic valuta
// ROB_RPT_STAMPA_PROFAKTURA, in a foreign valuta ROB_RPT_PROFAKTURAIZV (of the kupac of the kontni plan).
func (h *RobnoDokumentaHandler) stampaProfaktura(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if dok.Devizni() {
		robnoDokumentaStampaNijeImplementirana("ROB_RPT_PROFAKTURAIZV", false)(c, dok)
		return
	}
	robnoDokumentaStampaNijeImplementirana("ROB_RPT_STAMPA_PROFAKTURA", false)(c, dok)
}

// stampaKnjiznoPismo is the print of the knjižna pisma (groups KNO and KNZ): a document with stavke is
// ROB_RPT_KNJPISMO, a document without stavke (finansijsko) ROB_RPT_KNJPISMOFIN.
func (h *RobnoDokumentaHandler) stampaKnjiznoPismo(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if dok.ImaStavke {
		robnoDokumentaStampaNijeImplementirana("ROB_RPT_KNJPISMO", false)(c, dok)
		return
	}
	robnoDokumentaStampaNijeImplementirana("ROB_RPT_KNJPISMOFIN", false)(c, dok)
}

// stampaNivelacijaMaloprodaje is the print of the nivelacije maloprodaje (group MNI): only a document
// with stavke is printed (RPT_NIVELACIJA_MALOPRODAJE).
func (h *RobnoDokumentaHandler) stampaNivelacijaMaloprodaje(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if dok.RdokID != 0 && !dok.ImaStavke {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, "Dokument nema stavki")
		return
	}
	robnoDokumentaStampaNijeImplementirana("RPT_NIVELACIJA_MALOPRODAJE", false)(c, dok)
}

// stampaFakturaZaRobu is the print of the group FZR (Srbija): a document in the domestic valuta is
// PR_RPT_FAKTURA_OTP (or its variant of DODOZNFAK); the ino faktura has no print.
func (h *RobnoDokumentaHandler) stampaFakturaZaRobu(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if dok.Devizni() {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, "Ino faktura nije implementirana!")
		return
	}
	robnoDokumentaStampaNijeImplementirana("PR_RPT_FAKTURA_OTP", true)(c, dok)
}

// stampaKnjiznoOdobrenje is the print of the knjižna odobrenja (groups KPK and KPF): only a document in
// the domestic valuta is printed (PR_RPT_KNJODOBRENJE).
func (h *RobnoDokumentaHandler) stampaKnjiznoOdobrenje(c *gin.Context, dok domain.RobnoStampaDokumentDto) {
	if dok.Devizni() {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, "Štampa knjižnog odobrenja u stranoj valuti ne postoji")
		return
	}
	robnoDokumentaStampaNijeImplementirana("PR_RPT_KNJODOBRENJE", false)(c, dok)
}

// PregledEFaktura renders the "eFaktura" sub-tab (filters, actions and grid).
func (h *RobnoDokumentaHandler) PregledEFaktura(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabPregled)
	subTabs := common.SetActiveTab(h.subTabsFor(robnoDokumentaTabPregled), robnoDokumentaSubTabPregledEFaktura)
	params := h.pregledParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaEFakturaTitle, robnoDokumentaEFakturaTableID, h.service.GetPregledEFakturaTableFields(), "", robnoDokumentaURLEFaktura, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaEFakturaTitle, robnoDokumentaURLEFaktura, false, false, false)
	tbl.HasTotals = true
	if common.IsDataRequest(c) {
		if !h.getPregledEFaktura(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnObrada := h.obradaButton(robnoDokumentaURLEFaktura, robnoDokumentaEFakturaTableID, hxValsRobnoDokumentaPregledEFaktura)
	btnPosalji := h.efakturaActionButton("efaktura-posalji-btn", "Pošalji eFakturu", "komercijala_otpremanjerobe", robnoDokumentaURLEFakturaPosalji)
	btnProveri := h.efakturaActionButton("efaktura-proveri-status-btn", "Proveri status EF", "refresh", robnoDokumentaURLEFakturaProveri)
	btnOtkazi := h.efakturaActionButton("efaktura-otkazi-btn", "Otkaži eFakturu", "cancel", robnoDokumentaURLEFakturaOtkazi)
	btnStorniraj := h.efakturaActionButton("efaktura-storniraj-btn", "Storniraj eFakturu", "back", robnoDokumentaURLEFakturaStorniraj)
	btnStornirajPE := h.efakturaActionButton("efaktura-storniraj-pe-btn", "Storniraj PE", "back", robnoDokumentaURLEFakturaStornirajPE)
	btnAzurirajStatus := h.efakturaActionButton("efaktura-status-btn", "Ažuriranje statusa eFaktura", "refresh", robnoDokumentaURLEFakturaStatus)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLEFaktura, "#"+robnoDokumentaEFakturaTableID, hxValsRobnoDokumentaPregledEFaktura)
	if err := tmpl_robno.RobnoDokumentaEFaktura(tabs, subTabs, tbl, magValues, params, btnObrada, btnPosalji, btnProveri, btnOtkazi, btnStorniraj, btnStornirajPE, btnAzurirajStatus, search, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// getPregledEFaktura fills the grid of the "eFaktura" sub-tab (totals, then the rows).
func (h *RobnoDokumentaHandler) getPregledEFaktura(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaPregledEFaktura
	for _, total := range []bool{true, false} {
		if err := h.service.GetPregledEFaktura(c.Request.Context(), tbl, total, page, pageSize, params); err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return false
		}
	}
	return true
}

// pregledParams reads the filters of the "Pregled dokumenta" tab from the request.
func (h *RobnoDokumentaHandler) pregledParams(c *gin.Context) domain.RobnoDokumentaParams {
	params := domain.RobnoDokumentaParams{
		MagaciniID:      common.StringToInt(c.Query("magaciniid")),
		Vrd:             c.Query("vrd"),
		OdDanal:         c.Query("oddanal"),
		DoDanal:         c.Query("dodanal"),
		GrupeDokumenata: c.Query("grupedokumenata"),
		DatumStatusa:    c.Query("datumstatusa"),
		SearchText:      c.Query("query"),
	}
	if params.GrupeDokumenata == "" {
		params.GrupeDokumenata = robnoDokumentaGrupeEFaktura
	}
	if params.DatumStatusa == "" {
		params.DatumStatusa = h.businessToday(c.Request.Context())
	}
	return params
}

// pregledReportParams builds the parameters of the report header of the print of the "Štampa"
// sub-tab of "Pregled dokumenta" in a fixed order: the filters of its panel (magacin, range of the
// datuma naloga and vrsta dokumenta), shown only when set.
func (h *RobnoDokumentaHandler) pregledReportParams(params domain.RobnoDokumentaParams) []domain.ParameterItem {
	items := []domain.ParameterItem{}
	add := func(label, value string) {
		if value == "" {
			return
		}
		items = append(items, domain.ParameterItem{Name: h.translator.Label(label), Value: value})
	}
	if params.MagaciniID != 0 {
		add("Magacin", fmt.Sprintf("%d", params.MagaciniID))
	}
	add("Od datuma naloga", params.OdDanal)
	add("Do datuma naloga", params.DoDanal)
	add("Vrsta dokumenta", params.Vrd)
	return items
}

// efakturaActionButton builds one of the action buttons of the eFaktura sub-tab.
func (h *RobnoDokumentaHandler) efakturaActionButton(id, label, icon, url string) domain.Button {
	return common.SetButton(id, h.translator.Button(label), icon, url, "#"+robnoDokumentaInfoMessageID, "innerHTML", "POST", "", hxValsRobnoDokumentaPregledEFaktura, true, common.ClassSaveButton, "handleDialogResponse")
}

// businessToday returns the current date of the selected business year (yyyy-mm-dd).
func (h *RobnoDokumentaHandler) businessToday(ctx context.Context) string {
	today := time.Now()
	if session := domain.GetSessionFromStdContext(ctx); session != nil {
		today = time.Date(session.SelectedGod, today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	}
	return today.Format(common.HtmlLayout)
}

// PregledEFakturaPrint is the print of the "eFaktura" sub-tab (TODO: not implemented).
func (h *RobnoDokumentaHandler) PregledEFakturaPrint(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Štampa pregleda eFaktura još nije implementirana")
}

// PregledEFakturaAkcija handles the eFaktura actions (TODO: not implemented).
func (h *RobnoDokumentaHandler) PregledEFakturaAkcija(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "eFaktura operacija još nije implementirana")
}

// SpecifikacijeDokumenta renders the "Specifikacije dokumenta" tab (the print parameters and the
// grid of the robni dokumenti of the selection).
func (h *RobnoDokumentaHandler) SpecifikacijeDokumenta(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabSpecifikacije)
	subTabs := h.subTabsFor(robnoDokumentaTabSpecifikacije)
	params := h.specifikacijeParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaSpecifikacijeTitle, robnoDokumentaSpecifikacijeTableID, h.service.GetSpecifikacijeDokumentaTableFields(), "", robnoDokumentaURLSpecifikacije, 0, 0, 0, 0, h.cfg)
	// The buttons of the tab ("Obrada" and "Štampaj") are in the parameter panel, so the toolbar of
	// the grid holds only the title.
	common.SetTableConfig(&tbl, robnoDokumentaSpecifikacijeTitle, robnoDokumentaURLSpecifikacije, false, false, false)
	if common.IsDataRequest(c) {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoDokumentaSpecifikacije
		err := h.service.GetSpecifikacijeDokumenta(c.Request.Context(), &tbl, true, page, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		err = h.service.GetSpecifikacijeDokumenta(c.Request.Context(), &tbl, false, page, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnObrada := h.obradaButton(robnoDokumentaURLSpecifikacije, robnoDokumentaSpecifikacijeTableID, hxValsRobnoDokumentaSpecifikacije)
	btnPrint := common.SetPrintButton("specifikacije-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLSpecifikacijeStampa, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	btnPrint.HxVals = hxValsRobnoDokumentaSpecifikacije

	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLSpecifikacije, "#"+robnoDokumentaSpecifikacijeTableID, hxValsRobnoDokumentaSpecifikacije)
	tmpl_robno.RobnoDokumentaSpecifikacije(tabs, subTabs, tbl, magValues, tipdokValues, vrstaDokumentaValues, params, btnObrada, btnPrint, search, h.translator).Render(ctx, c.Writer)
}

// specifikacijeParams reads the parameters of the "Specifikacije dokumenta" tab from the request. The
// default selection is the whole range of the broj naloga and of the broj dokumenta (robnoDokumentaOdNaloga
// - robnoDokumentaDoNaloga) and the period from the first day of the current month of the business year
// until today.
func (h *RobnoDokumentaHandler) specifikacijeParams(c *gin.Context) domain.RobnoDokumentaParams {
	params := domain.RobnoDokumentaParams{
		MagaciniID:       common.StringToInt(c.Query("magaciniid")),
		OdVrd:            c.Query("odvrd"),
		DoVrd:            c.Query("dovrd"),
		Vrd:              c.Query("vrd"),
		OdNaloga:         c.Query("odnaloga"),
		DoNaloga:         c.Query("donaloga"),
		OdDokum:          c.Query("oddokum"),
		DoDokum:          c.Query("dodokum"),
		OdDanal:          c.Query("oddanal"),
		DoDanal:          c.Query("dodanal"),
		StampajSamoZbir:  c.Query("stampajsamozbir") == "true",
		TipSpecifikacije: c.Query("tipspecifikacije"),
		SearchText:       c.Query("query"),
	}
	if params.OdNaloga == "" {
		params.OdNaloga = robnoDokumentaOdNaloga
	}
	if params.DoNaloga == "" {
		params.DoNaloga = robnoDokumentaDoNaloga
	}
	if params.OdDokum == "" {
		params.OdDokum = robnoDokumentaOdNaloga
	}
	if params.DoDokum == "" {
		params.DoDokum = robnoDokumentaDoNaloga
	}
	if params.OdDanal == "" {
		params.OdDanal = h.businessMonthStart(c.Request.Context())
	}
	if params.DoDanal == "" {
		params.DoDanal = h.businessToday(c.Request.Context())
	}
	return params
}

// specifikacijeReportParams builds the parameters of the report header of the print of the
// "Specifikacije dokumenta" tab in the order of the panel of the tab: the ranges of the vrste naloga
// za knjiženje, of the broj naloga and of the broj dokumenta, the vrsta dokumenta, the range of the
// dates and the magacin.
func (h *RobnoDokumentaHandler) specifikacijeReportParams(params domain.RobnoDokumentaParams) []domain.ParameterItem {
	items := []domain.ParameterItem{}
	add := func(label, value string) {
		if value == "" {
			return
		}
		items = append(items, domain.ParameterItem{Name: h.translator.Label(label), Value: value})
	}
	add("Od vrste naloga za knjiženje", params.OdVrd)
	add("Do vrste naloga za knjiženje", params.DoVrd)
	add("Vrsta dokumenta", params.Vrd)
	add("Od naloga", params.OdNaloga)
	add("Do naloga", params.DoNaloga)
	add("Od dokumenta", params.OdDokum)
	add("Do dokumenta", params.DoDokum)
	add("Od datuma", params.OdDanal)
	add("Do datuma", params.DoDanal)
	if params.MagaciniID != 0 {
		add("Magacin", fmt.Sprintf("%d", params.MagaciniID))
	}
	return items
}

// SpecifikacijeDokumentaPrint is the print of the "Specifikacije dokumenta" tab. The legacy screen
// chooses the report by the group of the selected vrsta dokumenta (dokvrsta.grpdok):
//
//	FAK, PRE, DIR, ARU, FZR -> RPT_ROB_SPECIFIKACIJA_fakture (the radio "Specifikacija", implemented
//	                           as the report RobSpecifikacijaFakture) or RPT_ROB_FAKTURNA_KNJIGA
//	                           (the radio "Fakturna knjiga", TODO),
//	IRT, IRS, IRP      -> RPT_ROB_IRTSPECIF (TODO),
//	NIV                -> RPT_ROB_SPECIFIKACIJA_NIV (TODO),
//	ARA, PLT           -> RPT_ROB_SPECIF_KALK (TODO),
//	KAL                -> (no report),
//	otherwise          -> "Program za ovu opciju nije instaliran".
//
// TODO: once the remaining reports are written, the print must choose the report by the grpdok of the
// selected vrsta dokumenta (and by the state of the radio buttons "Specifikacija"/"Fakturna knjiga").
func (h *RobnoDokumentaHandler) SpecifikacijeDokumentaPrint(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	params := h.specifikacijeParams(c)
	tbl := common.SetTableBasicData(robnoDokumentaSpecifikacijeTitle, robnoDokumentaSpecifikacijeTableID, h.service.GetSpecifikacijeDokumentaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	// One fetch with TipStampe-like unpaginated settings (page 0 / size 0 are no-ops in the
	// QueryBuilder): the whole selection is printed, like the prints of the other tabs.
	if err := h.service.GetSpecifikacijeDokumenta(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoDokumentaSpecifikacijaFaktureTitle,
	}
	tmpl_rep_rob.RobSpecifikacijaFakture(repParams, tbl, h.specifikacijeReportParams(params), h.translator).Render(ctx, c.Writer)
}

// KontiranjeDokumenata renders the "Kontiranje dokumenata" tab (first of its sub-tabs).
func (h *RobnoDokumentaHandler) KontiranjeDokumenata(c *gin.Context) {
	h.kontiranjeKnjizenje(c)
}

// KontiranjeKnjizenje renders the "Knjiženje dokumenata" sub-tab.
func (h *RobnoDokumentaHandler) KontiranjeKnjizenje(c *gin.Context) {
	h.kontiranjeKnjizenje(c)
}

// kontiranjeKnjizenje renders the "Knjiženje dokumenata" sub-tab (actions and grid).
func (h *RobnoDokumentaHandler) kontiranjeKnjizenje(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabKontiranje)
	subTabs := common.SetActiveTab(h.subTabsFor(robnoDokumentaTabKontiranje), robnoDokumentaSubTabKontiranjeKnjizenje)
	params := h.kontiranjeParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaKnjizenjeTitle, robnoDokumentaKnjizenjeTableID, h.service.GetKontiranjeKnjizenjeTableFields(), "", robnoDokumentaURLKnjizenje, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaKnjizenjeTitle, robnoDokumentaURLKnjizenje, false, false, false)
	tbl.HasTotals = true
	if common.IsDataRequest(c) {
		if !h.getKontiranjeKnjizenje(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnObrada := h.obradaButton(robnoDokumentaURLKnjizenje, robnoDokumentaKnjizenjeTableID, hxValsRobnoDokumentaKnjizenje)
	btnRavnoteza := h.kontiranjeActionButton(robnoDokumentaRavnotezaBtnID, "Proveri ravnotežu", "fin_ravnoteza", robnoDokumentaURLKnjizenjeRavnot, hxValsRobnoDokumentaKnjizenje)
	btnKnjizi := h.kontiranjeActionButton(robnoDokumentaKontiranjeBtnID, "Knjiži", "fin_knjizenje", robnoDokumentaURLKnjizenjeKnjizi, hxValsRobnoDokumentaKnjizenje)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLKnjizenje, "#"+robnoDokumentaKnjizenjeTableID, hxValsRobnoDokumentaKnjizenje)
	if err := tmpl_robno.RobnoDokumentaKontiranjeKnjizenje(tabs, subTabs, tbl, tipdokValues, vrstaDokumentaValues, magValues, params, btnObrada, btnRavnoteza, btnKnjizi, search, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// getKontiranjeKnjizenje fills the grid of the "Knjiženje dokumenata" sub-tab.
func (h *RobnoDokumentaHandler) getKontiranjeKnjizenje(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaKnjizenje
	for _, total := range []bool{true, false} {
		if err := h.service.GetKontiranjeKnjizenje(c.Request.Context(), tbl, total, page, pageSize, params); err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return false
		}
	}
	return true
}

// KontiranjePregled renders the "Pregled proknjiženih / neproknjiženih dokumenata" sub-tab.
func (h *RobnoDokumentaHandler) KontiranjePregled(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabKontiranje)
	subTabs := common.SetActiveTab(h.subTabsFor(robnoDokumentaTabKontiranje), robnoDokumentaSubTabKontiranjePregled)
	params := h.kontiranjeParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaKontiranjePregledTitle, robnoDokumentaKontiranjePregledTableID, h.service.GetKontiranjePregledTableFields(), "", robnoDokumentaURLKontiranjePregled, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaKontiranjePregledTitle, robnoDokumentaURLKontiranjePregled, false, false, false)
	tbl.HasTotals = true
	if common.IsDataRequest(c) {
		if !h.getKontiranjePregled(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnObrada := h.obradaButton(robnoDokumentaURLKontiranjePregled, robnoDokumentaKontiranjePregledTableID, hxValsRobnoDokumentaKontiranjePregled)
	// The action of the checkbox "Označi prikazana dokumenta kao neproknjižena..." of the legacy
	// screen: the checkbox is sent with the parameters of the grid, the marking of the displayed
	// documents is then executed by the action (TODO: not implemented yet).
	btnOznaci := h.kontiranjeActionButton(robnoDokumentaOznaciBtnID, "Označi kao neproknjižena", "cancel", robnoDokumentaURLOznaciNeproknjizene, hxValsRobnoDokumentaKontiranjePregled)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLKontiranjePregled, "#"+robnoDokumentaKontiranjePregledTableID, hxValsRobnoDokumentaKontiranjePregled)
	if err := tmpl_robno.RobnoDokumentaKontiranjePregled(tabs, subTabs, tbl, tipdokValues, magValues, params, btnObrada, btnOznaci, search, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// getKontiranjePregled fills the grid of the "Pregled proknjiženih / neproknjiženih" sub-tab.
func (h *RobnoDokumentaHandler) getKontiranjePregled(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaKontiranjePregled
	for _, total := range []bool{true, false} {
		if err := h.service.GetKontiranjePregled(c.Request.Context(), tbl, total, page, pageSize, params); err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return false
		}
	}
	return true
}

// KontiranjePoMagacinima renders the "Pregled ... po magacinima" sub-tab.
func (h *RobnoDokumentaHandler) KontiranjePoMagacinima(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabKontiranje)
	subTabs := common.SetActiveTab(h.subTabsFor(robnoDokumentaTabKontiranje), robnoDokumentaSubTabKontiranjePoMagacinima)
	params := h.kontiranjeParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaPoMagacinimaTitle, robnoDokumentaPoMagacinimaTableID, h.service.GetKontiranjePoMagacinimaTableFields(), "", robnoDokumentaURLPoMagacinima, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaPoMagacinimaTitle, robnoDokumentaURLPoMagacinima, false, false, false)
	tbl.HasTotals = true
	if common.IsDataRequest(c) {
		if !h.getKontiranjePoMagacinima(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnObrada := h.obradaButton(robnoDokumentaURLPoMagacinima, robnoDokumentaPoMagacinimaTableID, hxValsRobnoDokumentaPoMagacinima)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLPoMagacinima, "#"+robnoDokumentaPoMagacinimaTableID, hxValsRobnoDokumentaPoMagacinima)
	if err := tmpl_robno.RobnoDokumentaKontiranjePoMagacinima(tabs, subTabs, tbl, magValues, params, btnObrada, search, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// getKontiranjePoMagacinima fills the grid of the "Pregled ... po magacinima" sub-tab.
func (h *RobnoDokumentaHandler) getKontiranjePoMagacinima(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaPoMagacinima
	for _, total := range []bool{true, false} {
		if err := h.service.GetKontiranjePoMagacinima(c.Request.Context(), tbl, total, page, pageSize, params); err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return false
		}
	}
	return true
}

// kontiranjeParams reads the parameters of the "Kontiranje dokumenata" tab from the request.
func (h *RobnoDokumentaHandler) kontiranjeParams(c *gin.Context) domain.RobnoDokumentaParams {
	params := domain.RobnoDokumentaParams{
		MagaciniID:           common.StringToInt(c.Query("magaciniid")),
		Tipdok:               c.Query("tipdok"),
		Vrd:                  c.Query("vrd"),
		OdNaloga:             c.Query("odnaloga"),
		DoNaloga:             c.Query("donaloga"),
		OdDokum:              c.Query("oddokum"),
		DoDokum:              c.Query("dodokum"),
		OdDanal:              c.Query("oddanal"),
		DoDanal:              c.Query("dodanal"),
		Proknjizen:           c.Query("proknjizen"),
		OznaciNeproknjizenim: c.Query("oznacineproknjizenim") != "",
		SearchText:           c.Query("query"),
	}
	if params.OdNaloga == "" {
		params.OdNaloga = robnoDokumentaOdNaloga
	}
	if params.DoNaloga == "" {
		params.DoNaloga = robnoDokumentaDoNaloga
	}
	if params.OdDokum == "" {
		params.OdDokum = robnoDokumentaOdNaloga
	}
	if params.DoDokum == "" {
		params.DoDokum = robnoDokumentaDoNaloga
	}
	if params.Proknjizen == "" {
		params.Proknjizen = robnoDokumentaProknjizen
	}
	today := h.businessToday(c.Request.Context())
	if params.DoDanal == "" {
		params.DoDanal = today
	}
	if params.OdDanal == "" {
		params.OdDanal = h.businessDaysBefore(c.Request.Context(), robnoDokumentaDanaSelekcije)
	}
	return params
}

// kontiranjeActionButton builds one of the action buttons of the "Knjiženje" sub-tab.
func (h *RobnoDokumentaHandler) kontiranjeActionButton(id, label, icon, url, hxVals string) domain.Button {
	return common.SetButton(id, h.translator.Button(label), icon, url, "#"+robnoDokumentaInfoMessageID, "innerHTML", "POST", "", hxVals, true, common.ClassSaveButton, "handleDialogResponse")
}

// businessDaysBefore returns the date of the business year N days before today (yyyy-mm-dd).
func (h *RobnoDokumentaHandler) businessDaysBefore(ctx context.Context, days int) string {
	today := time.Now()
	if session := domain.GetSessionFromStdContext(ctx); session != nil {
		today = time.Date(session.SelectedGod, today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	}
	return today.AddDate(0, 0, -days).Format(common.HtmlLayout)
}

// businessMonthStart returns the first day of the current month of the selected business year
// (yyyy-mm-dd).
func (h *RobnoDokumentaHandler) businessMonthStart(ctx context.Context) string {
	today := time.Now()
	if session := domain.GetSessionFromStdContext(ctx); session != nil {
		today = time.Date(session.SelectedGod, today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	}
	first := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	return first.Format(common.HtmlLayout)
}

// KontiranjeKnjizi is the "Knjiži" action of the sub-tab (TODO: not implemented).
func (h *RobnoDokumentaHandler) KontiranjeKnjizi(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Knjiženje robnih dokumenata još nije implementirano")
}

// KontiranjeRavnoteza is the "Pr. ravnotežu" action of the sub-tab (TODO: not implemented).
func (h *RobnoDokumentaHandler) KontiranjeRavnoteza(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Provera ravnoteže još nije implementirana")
}

// KontiranjeOznaciNeproknjizene marks the displayed documents as not posted (TODO).
func (h *RobnoDokumentaHandler) KontiranjeOznaciNeproknjizene(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Označavanje dokumenata kao neproknjiženih još nije implementirano")
}

// PrepisDokumenta renders the "Prepis dokumenta" tab (still empty).
func (h *RobnoDokumentaHandler) PrepisDokumenta(c *gin.Context) {
	h.render(c, robnoDokumentaTabPrepis)
}

// tab 6
// PrikazUkupneObrade renders the "Prikaz ukupne obrade" tab (totals per magacin).
func (h *RobnoDokumentaHandler) PrikazUkupneObrade(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabPrikazUkupneObrade)
	tbl := common.SetTableBasicData(robnoDokumentaUkupnaObradaTitle, robnoDokumentaUkupnaObradaTableID, h.service.GetPrikazUkupneObradeTableFields(), "", robnoDokumentaURLUkupnaObrada, 0, 0, 0, 0, h.cfg)
	// The toolbar of the grid holds the print button (there is no "Obrada" button on this tab).
	common.SetTableConfig(&tbl, robnoDokumentaUkupnaObradaTitle, robnoDokumentaURLUkupnaObrada, false, false, false)
	btnPrint := common.SetPrintButton("ukupna-obrada-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLUkupnaObradaStampa, "GET", true, common.ClassPrintButton, "")
	searchInput := common.CreateSearchInput("search-input", h.translator, robnoDokumentaURLUkupnaObrada, fmt.Sprintf("#%s", robnoDokumentaUkupnaObradaTableID), "")

	currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	if err := h.service.GetPrikazUkupneObrade(ctx, &tbl, true, currentPage, pageSize, common.TipStampePreview); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if err := h.service.GetPrikazUkupneObrade(ctx, &tbl, false, currentPage, pageSize, common.TipStampePreview); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tmpl_robno.RobnoDokumentaPrikazUkupneObrade(tabs, domain.TabData{}, tbl, btnPrint, searchInput, h.translator).Render(ctx, c.Writer)
}

// PrikazUkupneObradePrint is the print of the "Prikaz ukupne obrade" tab.
func (h *RobnoDokumentaHandler) PrikazUkupneObradePrint(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tbl := common.SetTableBasicData(robnoDokumentaUkupnaObradaTitle, robnoDokumentaUkupnaObradaTableID, h.service.GetPrikazUkupneObradeTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	if err := h.service.GetPrikazUkupneObrade(ctx, &tbl, false, 0, 0, common.TipStampePrint); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := domain.ReportParameters{
		Orientation:    "landscape",
		CompanyName:    fvrData.Naziv,
		Adress:         fvrData.Adresa,
		Postcode:       fvrData.Pobro,
		City:           fvrData.Mesto,
		PIB:            fvrData.PIB,
		MatBroj:        fvrData.Matbr,
		ReportName:     robnoDokumentaUkupnaObradaTitle,
		ParameterItems: map[string]domain.ParameterItem{},
	}
	if err := tmpl_rep_rob.RobnoDokumentaPrikazUkupneObradeStampa(repParams, tbl, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// tab 7
// PrikazNaloga renders the "Prikaz naloga" tab (robni nalozi of the selection).
func (h *RobnoDokumentaHandler) PrikazNaloga(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabPrikazNaloga)
	subTabs := domain.TabData{}
	params := h.prikazNalogaParams(c)
	tbl := common.SetTableBasicData(robnoDokumentaPrikazNalogaTitle, robnoDokumentaPrikazNalogaTableID, h.service.GetPrikazNalogaTableFields(), "", robnoDokumentaURLPrikazNaloga, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaPrikazNalogaTitle, robnoDokumentaURLPrikazNaloga, false, false, false)
	searchInput := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLPrikazNaloga, "#"+robnoDokumentaPrikazNalogaTableID, hxValsRobnoDokumentaPrikazNaloga)

	if common.IsDataRequest(c) {
		currentPage, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		err := h.service.GetPrikazNaloga(c.Request.Context(), &tbl, true, currentPage, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		err = h.service.GetPrikazNaloga(c.Request.Context(), &tbl, false, currentPage, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnPrint := common.SetPrintButton("prikaz-naloga-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLPrikazNalogaStampa, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	btnPrint.HxVals = hxValsRobnoDokumentaPrikazNaloga

	btnObrada := h.obradaButton(robnoDokumentaURLPrikazNaloga, robnoDokumentaPrikazNalogaTableID, hxValsRobnoDokumentaPrikazNaloga)
	tmpl_robno.RobnoDokumentaPrikazNaloga(tabs, subTabs, tbl, magValues, tipdokValues, params, btnObrada, btnPrint, searchInput, h.translator).Render(ctx, c.Writer)
}

// prikazNalogaParams reads the parameters of the "Prikaz naloga" tab from the request.
func (h *RobnoDokumentaHandler) prikazNalogaParams(c *gin.Context) domain.RobnoDokumentaParams {
	params := domain.RobnoDokumentaParams{
		MagaciniID:     common.StringToInt(c.Query("magaciniid")),
		OdVrd:          c.Query("odvrd"),
		DoVrd:          c.Query("dovrd"),
		OdNaloga:       c.Query("odnaloga"),
		DoNaloga:       c.Query("donaloga"),
		ChkDatumNaloga: c.Query("chkpodatumunaloga") == "true",
		OdDanal:        c.Query("oddanal"),
		DoDanal:        c.Query("dodanal"),
		ChkDatumObrade: c.Query("chkpodatumuobrade") == "true",
		OdDatob:        c.Query("oddatob"),
		DoDatob:        c.Query("dodatob"),
		ChkOperator:    c.Query("chkpooperateru") == "true",
		Oper:           c.Query("oper"),
		SearchText:     c.Query("query"),
	}
	if params.OdNaloga == "" {
		params.OdNaloga = robnoDokumentaOdNaloga
	}
	if params.DoNaloga == "" {
		params.DoNaloga = robnoDokumentaDoNaloga
	}
	today := h.businessToday(c.Request.Context())
	if params.DoDanal == "" {
		params.DoDanal = today
	}
	if params.OdDanal == "" {
		params.OdDanal = h.businessDaysBefore(c.Request.Context(), robnoDokumentaDanaSelekcije)
	}
	if params.DoDatob == "" {
		params.DoDatob = today
	}
	if params.OdDatob == "" {
		params.OdDatob = h.businessDaysBefore(c.Request.Context(), robnoDokumentaDanaSelekcije)
	}
	return params
}

// prikazNalogaReportParams builds the parameters of the report header of the prints of the tabs
// "Prikaz naloga" (7) and "Prikaz dokumenata u nalogu" (8) in a fixed order: the selection of their
// common panel (magacin, range of the vrste naloga and of the broj naloga) plus the three optional
// filters, which are shown only when their checkbox is on.
func (h *RobnoDokumentaHandler) prikazNalogaReportParams(params domain.RobnoDokumentaParams) []domain.ParameterItem {
	items := []domain.ParameterItem{}
	add := func(label, value string) {
		if value == "" {
			return
		}
		items = append(items, domain.ParameterItem{Name: h.translator.Label(label), Value: value})
	}
	if params.MagaciniID != 0 {
		add("Magacin", fmt.Sprintf("%d", params.MagaciniID))
	}
	add("Od vrste naloga", params.OdVrd)
	add("Do vrste naloga", params.DoVrd)
	add("Od broja naloga", params.OdNaloga)
	add("Do broja naloga", params.DoNaloga)
	if params.ChkDatumNaloga {
		add("Od datuma naloga", params.OdDanal)
		add("Do datuma naloga", params.DoDanal)
	}
	if params.ChkDatumObrade {
		add("Od datuma obrade", params.OdDatob)
		add("Do datuma obrade", params.DoDatob)
	}
	if params.ChkOperator {
		add("Operater", params.Oper)
	}
	return items
}

// PrikazNalogaPrint is the print of the "Prikaz naloga" tab.
func (h *RobnoDokumentaHandler) PrikazNalogaPrint(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	params := h.prikazNalogaParams(c)
	tbl := common.SetTableBasicData(robnoDokumentaPrikazNalogaTitle, robnoDokumentaPrikazNalogaTableID, h.service.GetPrikazNalogaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	// One fetch with TipStampe-like unpaginated settings (page 0 / size 0 are no-ops in the
	// QueryBuilder): the whole selection is printed, like the "Prikaz ukupne obrade" print.
	if err := h.service.GetPrikazNaloga(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoDokumentaPrikazNalogaTitle,
	}
	tmpl_rep_rob.RobnoDokumentaPrikazNalogaStampa(repParams, tbl, h.prikazNalogaReportParams(params), h.translator).Render(ctx, c.Writer)
}

// PrikazDokumenataUNalogu renders the "Prikaz dokumenata u nalogu" tab.
func (h *RobnoDokumentaHandler) PrikazDokumenataUNalogu(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabPrikazDokumenataUNalogu)
	subTabs := h.subTabsFor(robnoDokumentaTabPrikazDokumenataUNalogu)
	params := h.prikazNalogaParams(c)
	tbl := common.SetTableBasicData(robnoDokumentaUNaloguTitle, robnoDokumentaUNaloguTableID, h.service.GetPrikazDokumenataUNaloguTableFields(), "", robnoDokumentaURLUNalogu, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaUNaloguTitle, robnoDokumentaURLUNalogu, false, false, false)
	searchInput := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLUNalogu, "#"+robnoDokumentaUNaloguTableID, hxValsRobnoDokumentaUNalogu)

	if common.IsDataRequest(c) {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoDokumentaUNalogu
		err := h.service.GetPrikazDokumenataUNalogu(c.Request.Context(), &tbl, true, page, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		err = h.service.GetPrikazDokumenataUNalogu(c.Request.Context(), &tbl, false, page, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnPrint := common.SetPrintButton("u-nalogu-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLUNaloguStampa, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	btnPrint.HxVals = hxValsRobnoDokumentaUNalogu

	btnObrada := h.obradaButton(robnoDokumentaURLUNalogu, robnoDokumentaUNaloguTableID, hxValsRobnoDokumentaUNalogu)
	tmpl_robno.RobnoDokumentaPrikazDokumenataUNalogu(tabs, subTabs, tbl, magValues, tipdokValues, params, btnObrada, btnPrint, searchInput, h.translator).Render(ctx, c.Writer)
}

// PrikazDokumenataUNaloguPrint is the print of the "Prikaz dokumenata u nalogu" tab.
func (h *RobnoDokumentaHandler) PrikazDokumenataUNaloguPrint(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	params := h.prikazNalogaParams(c)
	tbl := common.SetTableBasicData(robnoDokumentaUNaloguTitle, robnoDokumentaUNaloguTableID, h.service.GetPrikazDokumenataUNaloguTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	// One fetch with TipStampe-like unpaginated settings (page 0 / size 0 are no-ops in the
	// QueryBuilder): the whole selection is printed, like the "Prikaz naloga" print.
	if err := h.service.GetPrikazDokumenataUNalogu(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoDokumentaUNaloguTitle,
	}
	tmpl_rep_rob.RobnoDokumentaPrikazDokumenataUNaloguStampa(repParams, tbl, h.prikazNalogaReportParams(params), h.translator).Render(ctx, c.Writer)
}

// PrikazDokumenataPoOperateru renders the "Prikaz dokumenata po operateru" tab.
func (h *RobnoDokumentaHandler) PrikazDokumenataPoOperateru(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, robnoDokumentaTabPrikazDokumenataPooperateru)
	subTabs := domain.TabData{}
	params := h.prikazNalogaParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaPooperateruTitle, robnoDokumentaPooperateruTableID, h.service.GetPrikazDokumenataPooperateruTableFields(), "", robnoDokumentaURLPooperateru, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaPooperateruTitle, robnoDokumentaURLPooperateru, false, false, false)
	searchInput := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoDokumentaURLPooperateru, "#"+robnoDokumentaPooperateruTableID, hxValsRobnoDokumentaPooperateru)
	if common.IsDataRequest(c) {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl.Pagination.HxVals = hxValsRobnoDokumentaPooperateru
		err := h.service.GetPrikazDokumenataPoOperateru(c.Request.Context(), &tbl, true, page, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		err = h.service.GetPrikazDokumenataPoOperateru(c.Request.Context(), &tbl, false, page, pageSize, params, common.TipStampePreview)
		if err != nil {
			utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	btnPrint := common.SetPrintButton("po-operateru-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLPooperateruStampa, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	btnPrint.HxVals = hxValsRobnoDokumentaPooperateru
	btnObrada := h.obradaButton(robnoDokumentaURLPooperateru, robnoDokumentaPooperateruTableID, hxValsRobnoDokumentaPooperateru)
	tmpl_robno.RobnoDokumentaPrikazDokumenataPoOperateru(tabs, subTabs, tbl, params, btnObrada, btnPrint, searchInput, h.translator).Render(ctx, c.Writer)
}

// PrikazDokumenataPoOperateruPrint is the print of the "Prikaz dokumenata po operateru" tab.
func (h *RobnoDokumentaHandler) PrikazDokumenataPoOperateruPrint(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	params := h.prikazNalogaParams(c)
	tbl := common.SetTableBasicData(robnoDokumentaPooperateruTitle, robnoDokumentaPooperateruTableID, h.service.GetPrikazDokumenataPooperateruTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	// One fetch with TipStampe-like unpaginated settings (page 0 / size 0 are no-ops in the
	// QueryBuilder): the whole per-operater summary is printed.
	if err := h.service.GetPrikazDokumenataPoOperateru(ctx, &tbl, false, 0, 0, params, common.TipStampePrint); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	repParams := domain.ReportParameters{
		Orientation: "landscape",
		CompanyName: fvrData.Naziv,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		ReportName:  robnoDokumentaPooperateruTitle,
	}
	if err := tmpl_rep_rob.RobnoDokumentaPrikazDokumenataPoOperateruStampa(repParams, tbl, h.prikazNalogaReportParams(params), h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// render builds the (still empty) table of a tab and renders its template.
func (h *RobnoDokumentaHandler) render(c *gin.Context, tabIndex int) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, common.ErrMsgSessionNotFound)
		return
	}
	tabs := common.SetActiveTab(h.tabs, tabIndex)
	subTabs := h.subTabsFor(tabIndex)

	title, tableID, url, printURL, fields := h.tabData(tabIndex)
	tbl := common.SetTableBasicData(title, tableID, fields, url, url, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, tableID, url, false, false, false)

	// HTMX data request (search / paging) - the tab query is run here once it exists.
	if common.IsDataRequest(c) {
		// TODO: if !h.getPaginatedReport(c, &tbl) { return }
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaInfoMessageDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}

	btnObrada := h.obradaButton(url, tableID, "")
	// The print of every tab receives the whole selection of the option (the same print contract).
	btnPrint := common.SetPrintButton(tableID+"-stampa", "Štampa", "stampa", printURL, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, url, "#"+tableID, "")

	h.renderTab(c, tabIndex, tabs, subTabs, tbl, magValues, btnObrada, btnPrint, search)
}

// tabData returns the title, table id, url, print url and grid columns of a tab.
func (h *RobnoDokumentaHandler) tabData(tabIndex int) (title, tableID, url, printURL string, fields []domain.Fields) {
	switch tabIndex {
	case robnoDokumentaTabPregled:
		return robnoDokumentaPregledTitle, robnoDokumentaPregledTableID, robnoDokumentaURLPregled, robnoDokumentaURLPregledStampa, h.service.GetPregledDokumentaTableFields()
	case robnoDokumentaTabSpecifikacije:
		return robnoDokumentaSpecifikacijeTitle, robnoDokumentaSpecifikacijeTableID, robnoDokumentaURLSpecifikacije, robnoDokumentaURLSpecifikacijeStampa, h.service.GetSpecifikacijeDokumentaTableFields()
	case robnoDokumentaTabKontiranje:
		// The tab renders the first of its three sub-tabs ("Knjiženje dokumenata").
		return robnoDokumentaKnjizenjeTitle, robnoDokumentaKnjizenjeTableID, robnoDokumentaURLKnjizenje, robnoDokumentaURLKontiranjeStampa, h.service.GetKontiranjeKnjizenjeTableFields()
	case robnoDokumentaTabPrepis:
		return robnoDokumentaPrepisTitle, robnoDokumentaPrepisTableID, robnoDokumentaURLPrepis, robnoDokumentaURLPrepisStampa, h.service.GetPrepisDokumentaTableFields()
	case robnoDokumentaTabPrikazUkupneObrade:
		return robnoDokumentaUkupnaObradaTitle, robnoDokumentaUkupnaObradaTableID, robnoDokumentaURLUkupnaObrada, robnoDokumentaURLUkupnaObradaStampa, h.service.GetPrikazUkupneObradeTableFields()
	case robnoDokumentaTabPrikazNaloga:
		return robnoDokumentaPrikazNalogaTitle, robnoDokumentaPrikazNalogaTableID, robnoDokumentaURLPrikazNaloga, robnoDokumentaURLPrikazNalogaStampa, h.service.GetPrikazNalogaTableFields()
	case robnoDokumentaTabPrikazDokumenataUNalogu:
		return robnoDokumentaUNaloguTitle, robnoDokumentaUNaloguTableID, robnoDokumentaURLUNalogu, robnoDokumentaURLUNaloguStampa, h.service.GetPrikazDokumenataUNaloguTableFields()
	case robnoDokumentaTabPrikazDokumenataPooperateru:
		return robnoDokumentaPooperateruTitle, robnoDokumentaPooperateruTableID, robnoDokumentaURLPooperateru, robnoDokumentaURLPooperateruStampa, h.service.GetPrikazDokumenataPooperateruTableFields()
	default:
		return robnoDokumentaUnosTitle, robnoDokumentaUnosTableID, robnoDokumentaURLUnos, robnoDokumentaURLUnosStampa, h.service.GetUnosDokumentaTableFields()
	}
}

// renderTab renders the template of one tab.
func (h *RobnoDokumentaHandler) renderTab(c *gin.Context, tabIndex int, tabs, subTabs domain.TabData, tbl domain.TableData, magValues []domain.ComboItem, btnObrada, btnPrint domain.Button, search domain.InputControl) error {
	ctx := c.Request.Context()
	switch tabIndex {
	case robnoDokumentaTabPrepis:
		return tmpl_robno.RobnoDokumentaPrepis(tabs, subTabs, tbl, magValues, btnObrada, btnPrint, search, h.translator).Render(ctx, c.Writer)
	default:
		return fmt.Errorf("unknown tab %d of the Robna dokumenta option", tabIndex)
	}
}

// subTabsFor returns the sub-tabs of the given tab.
func (h *RobnoDokumentaHandler) subTabsFor(tabIndex int) domain.TabData {
	if tabIndex >= 0 && tabIndex < len(h.subTabs) {
		return h.subTabs[tabIndex]
	}
	return domain.TabData{}
}

// obradaButton builds the "Obrada" button of a tab.
func (h *RobnoDokumentaHandler) obradaButton(url, tableID, hxVals string) domain.Button {
	return common.SetButton(robnoDokumentaObradaBtnID, "Obrada", "obrada", url, "#"+tableID, "innerHTML", "GET", "", hxVals, true, common.ClassSaveButton, "handleDialogResponse")
}

// AddRoutes registers the routes of the "Robna dokumenta" option.
func (h *RobnoDokumentaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta", h.RobnoDokumentaMain)
	r.GET("/api/robno-dokumenta/unos", h.RobnoDokumentaMain)
	r.GET("/api/robno-dokumenta/pregled", h.StampaDokumentaObrada)
	r.GET("/api/robno-dokumenta/pregled/stampa", h.StampaDokumentaObrada)
	r.GET("/api/robno-dokumenta/pregled/stampa/print", h.StampaDokumentaPrint)
	r.GET("/api/robno-dokumenta/pregled/efaktura", h.PregledEFaktura)
	r.GET("/api/robno-dokumenta/pregled/efaktura/print", h.PregledEFakturaPrint)
	r.POST("/api/robno-dokumenta/pregled/efaktura/posalji", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/proveri-status", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/otkazi", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/storniraj", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/storniraj-pe", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/status", h.PregledEFakturaAkcija)
	r.GET("/api/robno-dokumenta/specifikacije", h.SpecifikacijeDokumenta)
	r.GET("/api/robno-dokumenta/specifikacije/stampa", h.SpecifikacijeDokumentaPrint)
	r.GET("/api/robno-dokumenta/kontiranje", h.KontiranjeDokumenata)
	r.GET("/api/robno-dokumenta/kontiranje/knjizenje", h.KontiranjeKnjizenje)
	r.GET("/api/robno-dokumenta/kontiranje/knjizenje/ravnoteza", h.KontiranjeRavnoteza)
	r.POST("/api/robno-dokumenta/kontiranje/knjizenje/ravnoteza", h.KontiranjeRavnoteza)
	r.POST("/api/robno-dokumenta/kontiranje/knjizenje/knjizi", h.KontiranjeKnjizi)
	r.GET("/api/robno-dokumenta/kontiranje/pregled", h.KontiranjePregled)
	r.POST("/api/robno-dokumenta/kontiranje/pregled/oznaci-neproknjizene", h.KontiranjeOznaciNeproknjizene)
	r.GET("/api/robno-dokumenta/kontiranje/po-magacinima", h.KontiranjePoMagacinima)
	r.GET("/api/robno-dokumenta/prepis", h.PrepisDokumenta)
	r.GET("/api/robno-dokumenta/ukupna-obrada", h.PrikazUkupneObrade)
	r.GET("/api/robno-dokumenta/ukupna-obrada/stampa", h.PrikazUkupneObradePrint)
	r.GET("/api/robno-dokumenta/prikaz-naloga", h.PrikazNaloga)
	r.GET("/api/robno-dokumenta/prikaz-naloga/stampa", h.PrikazNalogaPrint)
	r.GET("/api/robno-dokumenta/u-nalogu", h.PrikazDokumenataUNalogu)
	r.GET("/api/robno-dokumenta/u-nalogu/stampa", h.PrikazDokumenataUNaloguPrint)
	r.GET("/api/robno-dokumenta/po-operateru", h.PrikazDokumenataPoOperateru)
	r.GET("/api/robno-dokumenta/po-operateru/stampa", h.PrikazDokumenataPoOperateruPrint)
	r.GET("/api/robno-dokumenta/nextnalog", h.GetNextNalog)
	r.GET("/api/robno-dokumenta/nalog-data", h.GetNalogData)
	r.POST("/api/robno-dokumenta/confirm-addupdate", h.ConfirmUnosDokumenta)
	r.GET("/api/robno-dokumenta/confirm-addupdate", h.ConfirmUnosDokumenta)
	r.POST("/api/robno-dokumenta/unos", h.SaveUnosDokumenta)
	r.PUT("/api/robno-dokumenta/unos/:id", h.lm.WithEntityLockHold(robnoDokumentaEntityType, "id"), h.UpdateUnosDokumenta)

	// TODO (temporary): the two routes of the preview of the "Fakture veleprodaje" screen
	// (RobnoFakture). Remove them together with the "Fakture veleprodaje (preview)" button.
	r.GET("/api/robno-dokumenta/fakture-preview", h.FakturePreview)
	r.POST("/api/robno-dokumenta/fakture-preview/save", h.FakturePreviewSave)

	// Štampa fakture (RobnoStampaFaktura).
	r.GET(robnoDokumentaURLStampaFaktura, h.StampaFakturaPrint)
	r.GET(robnoDokumentaURLStampaPopis, h.StampaPopis)
	r.GET(robnoDokumentaURLStampaKalkulacija, h.StampaKalkulacija)
	r.GET(robnoDokumentaURLStampaFakturaMP, h.StampaFakturaMP)
	r.GET(robnoDokumentaURLStampaFakturaIzvoz, h.StampaFakturaIzvoz)
	r.GET(robnoDokumentaURLStampaFakturaUsluge, h.StampaFakturaUsluge)
	r.GET(robnoDokumentaURLStampaFakturaAvansni, h.StampaFakturaAvansni)
	r.GET(robnoDokumentaURLStampaOpstiDokument, h.StampaOpstiDokument)

	// TODO: add the stampa (print) routes of the tabs together with their print templates.
}

// robnoDokumentaTabs defines the tabs of the "Robna dokumenta" option.
func robnoDokumentaTabs(translator *i18n.Service) domain.TabData {
	return domain.TabData{Tabs: []domain.TabItem{
		{ID: "robno-dokumenta-unos", Label: translator.T("Unos dokumenta"), HXRequestUrl: robnoDokumentaURLUnos, IsActive: true, Name: "unos"},
		{ID: "robno-dokumenta-pregled", Label: translator.T("Pregled dokumenta"), HXRequestUrl: robnoDokumentaURLPregled, Name: "pregled"},
		{ID: "robno-dokumenta-specifikacije", Label: translator.T("Specifikacije dokumenta"), HXRequestUrl: robnoDokumentaURLSpecifikacije, Name: "specifikacije"},
		{ID: "robno-dokumenta-kontiranje", Label: translator.T("Kontiranje dokumenata"), HXRequestUrl: robnoDokumentaURLKontiranje, Name: "kontiranje"},
		{ID: "robno-dokumenta-prepis", Label: translator.T("Prepis dokumenta"), HXRequestUrl: robnoDokumentaURLPrepis, Name: "prepis"},
		{ID: "robno-dokumenta-ukupna-obrada", Label: translator.T("Prikaz ukupne obrade"), HXRequestUrl: robnoDokumentaURLUkupnaObrada, Name: "ukupna-obrada"},
		{ID: "robno-dokumenta-prikaz-naloga", Label: translator.T("Prikaz naloga"), HXRequestUrl: robnoDokumentaURLPrikazNaloga, Name: "prikaz-naloga"},
		{ID: "robno-dokumenta-u-nalogu", Label: translator.T("Prikaz dokumenata u nalogu"), HXRequestUrl: robnoDokumentaURLUNalogu, Name: "u-nalogu"},
		{ID: "robno-dokumenta-po-operateru", Label: translator.T("Prikaz dokumenata po operateru"), HXRequestUrl: robnoDokumentaURLPooperateru, Name: "po-operateru"},
	}}
}

// robnoDokumentaSubTabs holds the sub-tabs of every tab (indexed like robnoDokumentaTabs).
func robnoDokumentaSubTabs(translator *i18n.Service) []domain.TabData {
	empty := func() domain.TabData { return domain.TabData{Tabs: []domain.TabItem{}} }
	return []domain.TabData{
		empty(), // 0 - Unos dokumenta
		{
			// 1 - Pregled dokumenta: implemented are the first two sub-tabs; the others of the legacy
			// menu (Import Progress, Import Provizija, Import Poronizacija and Export TXT roba,usluge)
			// are added as they are built.
			Tabs: []domain.TabItem{
				{ID: "robno-dokumenta-pregled-stampa", Label: translator.T("Štampa"), HXRequestUrl: robnoDokumentaURLPregledStampa, IsActive: true, Name: "stampa"},
				{ID: "robno-dokumenta-pregled-efaktura", Label: translator.T("eFaktura"), HXRequestUrl: robnoDokumentaURLEFaktura, Name: "efaktura"},
			},
		},
		empty(), // 2 - Specifikacije dokumenta
		{
			// 3 - Kontiranje dokumenata: the three sub-tabs of the legacy screen.
			Tabs: []domain.TabItem{
				{ID: "robno-dokumenta-kontiranje-knjizenje", Label: translator.T("Knjiženje dokumenata"), HXRequestUrl: robnoDokumentaURLKnjizenje, IsActive: true, Name: "knjizenje"},
				{ID: "robno-dokumenta-kontiranje-pregled", Label: translator.T("Pregled proknjiženih / neproknjiženih dokumenata"), HXRequestUrl: robnoDokumentaURLKontiranjePregled, Name: "pregled"},
				{ID: "robno-dokumenta-kontiranje-magacini", Label: translator.T("Pregled proknjiženih / neproknjiženih dokumenata po magacinima"), HXRequestUrl: robnoDokumentaURLPoMagacinima, Name: "po-magacinima"},
			},
		},
		empty(), // 4 - Prepis dokumenta
		empty(), // 5 - Prikaz ukupne obrade
		empty(), // 6 - Prikaz naloga
		empty(), // 7 - Prikaz dokumenata u nalogu
		empty(), // 8 - Prikaz dokumenata po operateru
	}
}
