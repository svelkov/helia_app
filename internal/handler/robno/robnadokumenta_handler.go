package robno

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"helia/config"
	tmpl "helia/frontend/templates"
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
	// FakturePreviewSave together with the two table id / field helpers when the handler of the
	// fakture is written.
	robnoDokumentaURLFakturePreview     = robnoDokumentaURLPrefix + "/fakture-preview"
	robnoDokumentaURLFakturePreviewSave = robnoDokumentaURLFakturePreview + "/save"
	robnoDokumentaFaktureStavkeTableID  = "robno-fakture-stavke-table"
	robnoDokumentaFaktureAvansiTableID  = "robno-fakture-avansi-table"

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
		"proknjizen,oznacineproknjizenim"
	// robnoDokumentaInfoMessageID is the staging element of the messages of the actions of the
	// tabs (it is emitted by the layout of the option).
	robnoDokumentaInfoMessageID = "info-message"

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

type RobnoDokumentaHandler struct {
	service robnosvc.RobnoDokumentaService
	cfg     config.Config
	lm      *middleware.LockMiddleware
	ls      *middleware.LockService
	tabs    domain.TabData
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

func NewRobnoDokumentaHandler(s robnosvc.RobnoDokumentaService, cfg config.Config, lm *middleware.LockMiddleware, ls *middleware.LockService) *RobnoDokumentaHandler {
	h := &RobnoDokumentaHandler{
		service: s,
		cfg:     cfg,
		lm:      lm,
		ls:      ls,
		tabs:    robnoDokumentaTabs(),
		subTabs: robnoDokumentaSubTabs(),
	}
	h.setHandlerFieldValues()
	return h
}

// setHandlerFieldValues defines the buttons of the header of the "Unos dokumenta" tab, like the
// "Nalozi" (fnal) screen: "Snimi nalog" opens the confirm dialog of the save (which decides
// between a new nalog and the continuation of an existing one) and "Novi nalog" asks the service
// for the next free broj naloga of the selected vrsta naloga.
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
		HxTarget:      "#" + tmpl_robno.RobnoDokumentaFaktureDialogStagingID,
		HxSwap:        "innerHTML",
		BtnClass:      common.ClassButton,
	}
}

// Tab 1 - Unos dokumenta
//
// The tab shows the header of a robni nalog (vrsta naloga za knjiženje, vrsta dokumenta, datum
// obrade, broj naloga, datum naloga, opis knjiženja i magacin), the "Prikaz ukupne obrade" panel
// and the grid of the nalozi of the selected vrsta naloga.
func (h *RobnoDokumentaHandler) RobnoDokumentaMain(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabUnos)
	if !ok {
		return
	}
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
		h.error(c, err)
		return
	}
	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	total := domain.RobnoDokumentaTotal{}
	if err := h.service.GetUnosDokumentaTotal(ctx, &total); err != nil {
		h.error(c, err)
		return
	}
	payload, err := h.unosPayload(ctx, params)
	if err != nil {
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, robnoDokumentaURLUnos, "#"+robnoDokumentaUnosTableID, hxValsRobnoDokumentaUnos)
	if err := tmpl_robno.RobnoDokumentaMain(h.tabs, h.subTabsFor(robnoDokumentaTabUnos), tbl, tipdokValues, vrstaDokumentaValues, magValues, total, payload, h.btnSave, h.btnNoviNalog, h.btnFakture, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// FakturePreview renders the new "Fakture veleprodaje" screen (the RobnoFaktureDialog template) as a
// dialog over the "Unos dokumenta" tab, with the empty data of a new faktura. The state of the two
// collapsible controls of the screen is selected with ?snimljen=true (the header is then locked and
// the stavke are open).
//
// TODO (temporary): the screen has no handler yet; this method and the two preview routes only serve
// to open the template from the "Unos dokumenta" tab (see robnoDokumentaURLFakturePreview).
func (h *RobnoDokumentaHandler) FakturePreview(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}

	// Both grids of the screen are empty: their queries belong to the handler of the fakture (TODO).
	tbl := common.SetTableBasicData("Stavke fakture", robnoDokumentaFaktureStavkeTableID, robnoDokumentaFaktureStavkeFields(), "", robnoDokumentaURLFakturePreview, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaFaktureStavkeTableID, robnoDokumentaURLFakturePreview, false, false, false)
	// The dialog is opened by the "Fakture veleprodaje (preview)" button, which sends the request
	// source "btn" (components.Button). common.IsDataRequest reports every source except menu/tab as
	// a data request, so the source of the button has to be excluded here: otherwise the button would
	// receive only the body of the grid instead of the whole dialog. Only the requests of the two
	// grids of the screen (search, paging, header of the table) answer with their table.
	if common.IsDataRequest(c) && c.Request.Header.Get("X-Request-Source") != robnoDokumentaSourceBtn {
		utils.RenderContent(c, tbl)
		return
	}
	avansiTbl := common.SetTableBasicData("Avansi", robnoDokumentaFaktureAvansiTableID, robnoDokumentaFaktureAvansiFields(), "", robnoDokumentaURLFakturePreview, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&avansiTbl, robnoDokumentaFaktureAvansiTableID, robnoDokumentaURLFakturePreview, false, false, false)

	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}

	// The header of the faktura is empty (the documents and the kupac are read by the handler of the
	// fakture); only the date of the fakturisanja defaults to today of the selected business year.
	header := tmpl_robno.RobnoFaktureHeader{
		Snimljen:          c.Query("snimljen") == "true",
		DatumFakturisanja: h.businessToday(ctx),
	}
	// The buttons of the screen point to the preview routes; "Nazad" closes the dialog instead of
	// navigating back to the tab (TODO: the real endpoints of the fakture).
	btns := tmpl_robno.RobnoFaktureButtonsFor()
	btns.Save.HxActionURL = robnoDokumentaURLFakturePreviewSave
	btns.Back.HxActionURL = ""
	btns.Back.HxRequestType = ""
	btns.Back.HxOnClick = "closeDialog"
	btns.Back.HxOnClickArg = []any{tmpl_robno.RobnoFaktureDialogID}
	// The "Zatvori" button of the title bar of the dialog (CloseButton calls closeDialog with the
	// dialog id).
	btnClose := domain.Button{
		Id:       "robno-fakture-dialog-close",
		IdDialog: tmpl_robno.RobnoFaktureDialogID,
		BtnClass: common.ClassDialogCloseButton,
	}

	translator := i18n.GetInstance()
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, robnoDokumentaURLFakturePreview, "#"+robnoDokumentaFaktureStavkeTableID, "")
	// TODO: the combos of the valute and of the sistemi PDV of the screen.
	if err := tmpl_robno.RobnoFaktureDialog(tbl, avansiTbl, header, tmpl_robno.RobnoFaktureStavka{}, vrstaDokumentaValues, nil, nil, btns, btnClose, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// FakturePreviewSave answers the save of the header of the preview of the "Fakture veleprodaje"
// screen.
//
// TODO (temporary): it does not save anything; it answers success only so that the screen switches
// from the header to the entry of the stavke (robnoFaktureAfterHeaderSave locks the header and opens
// the stavke when the response was successful).
func (h *RobnoDokumentaHandler) FakturePreviewSave(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusOK, true, nil, "Pregled: podaci fakture nisu sačuvani (handler faktura još nije implementiran)")
}

// robnoDokumentaFaktureStavkeFields is the (temporary) set of columns the preview of the "Fakture
// veleprodaje" screen shows in the grid of the stavke.
//
// TODO (temporary): the real columns are defined by the handler of the fakture (the fields of the
// second print screen of the option).
func robnoDokumentaFaktureStavkeFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Redni broj", Width: "6", TextAlign: "right", SkipInSearch: true},
		{Name: "konto", Label: "Konto", Width: "7", SkipInSearch: true},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", SkipInSearch: true},
		{Name: "naziv", Label: "Naziv artikla", Width: "24"},
		{Name: "jm", Label: "JM", Width: "4", TextAlign: "center"},
		{Name: "kolicina", Label: "Količina", Width: "8", TextAlign: "right"},
		{Name: "magacinskacena", Label: "Magacinska cena", Width: "9", TextAlign: "right"},
		{Name: "iznos", Label: "Iznos", Width: "10", TextAlign: "right", IncludeInTotals: true},
		{Name: "prodajnacena", Label: "Prodajna cena", Width: "9", TextAlign: "right"},
		{Name: "rabat", Label: "Rabat", Width: "6", TextAlign: "right"},
	}
}

// robnoDokumentaFaktureAvansiFields is the (temporary) set of columns the preview of the "Fakture
// veleprodaje" screen shows in the grid of the avansi (TODO: the real ones).
func robnoDokumentaFaktureAvansiFields() []domain.Fields {
	return []domain.Fields{
		{Name: "trazi", Label: "Traži", Width: "5", SkipInSearch: true},
		{Name: "brdok", Label: "Broj dok.", Width: "8", SkipInSearch: true},
		{Name: "avansa", Label: "Avansa", Width: "8", TextAlign: "right", SkipInSearch: true},
		{Name: "datumavansa", Label: "Datum avansa", Width: "10", SkipInSearch: true},
		{Name: "iznosavansa", Label: "Iznos avansa", Width: "10", TextAlign: "right", SkipInSearch: true},
		{Name: "ostatakavansa", Label: "Ostatak avansa", Width: "10", TextAlign: "right", SkipInSearch: true},
		{Name: "zatvorenona", Label: "Iznos koji je zatv. na fakt.", Width: "12", TextAlign: "right", SkipInSearch: true},
	}
}

// GetNextNalog returns the next free broj naloga of a vrsta naloga as JSON. It is used by the
// "Novi nalog" button and, after the change of the vrsta naloga, by the hidden #nalog-trigger
// element of the tab (the response is handled by handleNextNalogResponse).
func (h *RobnoDokumentaHandler) GetNextNalog(c *gin.Context) {
	nextNalog, err := h.service.GetNextNalog(c.Request.Context(), c.Query("tipdok"))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, common.ErrMsgGetData)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "nalog": nextNalog})
}

// GetNalogData returns the dates, the opis and the magacin of an existing robni nalog as JSON. It is
// called when the user leaves the field "Broj naloga" (the response is handled by
// handleUpdateNalogDataResponse), like the "Nalozi" screen does. When the broj naloga is still
// free the response has success=false and the form keeps its values.
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

// ConfirmUnosDokumenta opens the confirm dialog of the save of the header: "Novi nalog" when the
// broj naloga is still free, "Nastavak knjiženja naloga" when it already exists (then the nalog is
// first checked to be free of locks). The dialog posts to the create or to the update of the
// header, depending on the answer.
func (h *RobnoDokumentaHandler) ConfirmUnosDokumenta(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
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
	msg := []string{translator.Message(`Otvaranje novog naloga?`)}
	if action == common.ActionUpdate {
		dialogTitle = "Nastavak knjiženja naloga"
		msg = []string{
			translator.Message(`Nastavak knjiženja naloga?`),
			fmt.Sprintf("%s: %s", translator.Message(`Vrsta naloga`), params.Tipdok),
			fmt.Sprintf("%s: %s", translator.Message(`Broj naloga`), params.Nalog),
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
	tmpl.DialogConfirm(msg, dialog, btnClose, btnSacuvaj, btnCancel, translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// SaveUnosDokumenta inserts the header of a new robni nalog (rnal) and locks it for the user, like
// the create of a financial nalog (fnal) does.
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

// UpdateUnosDokumenta saves the header of an existing robni nalog (rnal). The route keeps the lock
// of the nalog (the middleware refreshes it).
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

// nullDateHtml renders a nullable date of the header for an HTML date input (yyyy-mm-dd, "" when
// the date is not set).
func nullDateHtml(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(common.HtmlLayout)
}

// context checks the session and marks the given tab as active.
func (h *RobnoDokumentaHandler) context(c *gin.Context, tabIndex int) (context.Context, domain.UserSession, bool) {
	session := domain.GetSessionFromContext(c)
	if session == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, []domain.FieldError{}, common.ErrMsgUnauthorized)
		return nil, domain.UserSession{}, false
	}
	common.SetActiveTab(h.tabs, tabIndex)
	return c.Request.Context(), *session, true
}

// unosParams reads the header of the nalog and the grid filters of the tab from the request into the
// single params structure of the option (only the fields of this tab are filled).
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

// unosPayload prepares the form of the tab: the broj naloga is the next free number of the
// selected vrsta naloga and both dates default to today of the selected business year.
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

// getUnosDokumenta fills the grid of the tab (first the total records, then the rows of the page).
func (h *RobnoDokumentaHandler) getUnosDokumenta(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams, page, pageSize int) bool {
	for _, total := range []bool{true, false} {
		if err := h.service.GetUnosDokumenta(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// Tab 2 - Pregled dokumenta
//
// The tab shows the robne dokumente (rdok) of the current period. It has sub-tabs; implemented so
// far are the first two: "Štampa" (the print preview of the documents) and "eFaktura" (the status
// of the documents in the eFaktura system).
func (h *RobnoDokumentaHandler) PregledDokumenata(c *gin.Context) {
	h.pregledStampa(c)
}

// PregledStampa renders the "Štampa" sub-tab of "Pregled dokumenta".
func (h *RobnoDokumentaHandler) PregledStampa(c *gin.Context) {
	h.pregledStampa(c)
}

// pregledStampa renders the "Štampa" sub-tab: the filters (magacin, vrsta dokumenta and the range
// of the dates of the nalog) and the grid of the documents. The "Obradi" button sends a data
// request, in which case only the grid is rendered.
func (h *RobnoDokumentaHandler) pregledStampa(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabPregled)
	if !ok {
		return
	}
	subTabs := h.subTabsFor(robnoDokumentaTabPregled)
	common.SetActiveTab(subTabs, robnoDokumentaSubTabPregledStampa)
	params := h.pregledParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaPregledStampaTitle, robnoDokumentaPregledStampaTableID, h.service.GetPregledStampaTableFields(), "", robnoDokumentaURLPregledStampa, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, robnoDokumentaPregledStampaTitle, robnoDokumentaURLPregledStampa, false, false, false)
	tbl.HasTotals = true
	if common.IsDataRequest(c) {
		if !h.getPregledStampa(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(robnoDokumentaURLPregledStampa, robnoDokumentaPregledStampaTableID, hxValsRobnoDokumentaPregledStampa)
	btnPrint := common.SetPrintButton("stampa-btn", "Štampaj", "fin_print", robnoDokumentaURLPregledStampaPrint, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#info-message"
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, robnoDokumentaURLPregledStampa, "#"+robnoDokumentaPregledStampaTableID, hxValsRobnoDokumentaPregledStampa)
	if err := tmpl_robno.RobnoDokumentaPregled(h.tabs, subTabs, tbl, magValues, vrstaDokumentaValues, params, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// getPregledStampa fills the grid of the "Štampa" sub-tab (first the total records, then the rows).
func (h *RobnoDokumentaHandler) getPregledStampa(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaPregledStampa
	for _, total := range []bool{true, false} {
		if err := h.service.GetPregledStampa(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// PregledEFaktura renders the "eFaktura" sub-tab of "Pregled dokumenta": the filters (grupe
// dokumenata and the range of the dates of the documents), the actions of the eFaktura (slanje,
// provera statusa, otkazivanje, storniranje i ažuriranje statusa) and the grid with the status of
// every document. The "Obradi" button sends a data request, in which case only the grid is
// rendered.
func (h *RobnoDokumentaHandler) PregledEFaktura(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabPregled)
	if !ok {
		return
	}
	subTabs := h.subTabsFor(robnoDokumentaTabPregled)
	common.SetActiveTab(subTabs, robnoDokumentaSubTabPregledEFaktura)
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
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(robnoDokumentaURLEFaktura, robnoDokumentaEFakturaTableID, hxValsRobnoDokumentaPregledEFaktura)
	btnPosalji := h.efakturaActionButton("efaktura-posalji-btn", "Pošalji eFakturu", "komercijala_otpremanjerobe", robnoDokumentaURLEFakturaPosalji)
	btnProveri := h.efakturaActionButton("efaktura-proveri-status-btn", "Proveri status EF", "refresh", robnoDokumentaURLEFakturaProveri)
	btnOtkazi := h.efakturaActionButton("efaktura-otkazi-btn", "Otkaži eFakturu", "cancel", robnoDokumentaURLEFakturaOtkazi)
	btnStorniraj := h.efakturaActionButton("efaktura-storniraj-btn", "Storniraj eFakturu", "back", robnoDokumentaURLEFakturaStorniraj)
	btnStornirajPE := h.efakturaActionButton("efaktura-storniraj-pe-btn", "Storniraj PE", "back", robnoDokumentaURLEFakturaStornirajPE)
	btnAzurirajStatus := h.efakturaActionButton("efaktura-status-btn", "Ažuriranje statusa eFaktura", "refresh", robnoDokumentaURLEFakturaStatus)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, robnoDokumentaURLEFaktura, "#"+robnoDokumentaEFakturaTableID, hxValsRobnoDokumentaPregledEFaktura)
	if err := tmpl_robno.RobnoDokumentaEFaktura(h.tabs, subTabs, tbl, magValues, params, btnObrada, btnPosalji, btnProveri, btnOtkazi, btnStorniraj, btnStornirajPE, btnAzurirajStatus, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// getPregledEFaktura fills the grid of the "eFaktura" sub-tab (first the total records, then the
// rows).
func (h *RobnoDokumentaHandler) getPregledEFaktura(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaPregledEFaktura
	for _, total := range []bool{true, false} {
		if err := h.service.GetPregledEFaktura(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// pregledParams reads the filters of the "Pregled dokumenta" tab from the request into the single
// params structure of the option (only the fields of the two sub-tabs of the tab are filled). The
// grupe dokumenata of the eFaktura sub-tab default to robnoDokumentaGrupeEFaktura and the date of
// the ažuriranje statusa to the current date of the business year, like the legacy screen shows
// them.
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

// efakturaActionButton builds one of the action buttons of the eFaktura sub-tab: the actions are
// posted without a request source, so the handler knows the request is not a data request of the
// grid, and every action keeps its own icon. The label is translated here because the generic
// button component (which renders the icon) does not translate it.
func (h *RobnoDokumentaHandler) efakturaActionButton(id, label, icon, url string) domain.Button {
	return common.SetButton(id, i18n.GetInstance().Button(label), icon, url, "#"+robnoDokumentaInfoMessageID, "innerHTML", "POST", "", hxValsRobnoDokumentaPregledEFaktura, true, common.ClassSaveButton, "handleDialogResponse")
}

// businessToday returns the current date of the selected business year as the form of the tab
// expects it (yyyy-mm-dd).
func (h *RobnoDokumentaHandler) businessToday(ctx context.Context) string {
	today := time.Now()
	if session := domain.GetSessionFromStdContext(ctx); session != nil {
		today = time.Date(session.SelectedGod, today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	}
	return today.Format(common.HtmlLayout)
}

// PregledStampaPrint is the print of the "Štampa" sub-tab.
//
// TODO: not implemented yet - the print preview has to be built (the report template with the
// columns of the grid and the header of the company, like the other robno reports).
func (h *RobnoDokumentaHandler) PregledStampaPrint(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Štampa pregleda dokumenata još nije implementirana")
}

// PregledEFakturaPrint is the print of the "eFaktura" sub-tab.
//
// TODO: not implemented yet (see PregledStampaPrint).
func (h *RobnoDokumentaHandler) PregledEFakturaPrint(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Štampa pregleda eFaktura još nije implementirana")
}

// PregledEFakturaAkcija handles the actions of the eFaktura sub-tab (slanje, provera statusa,
// otkazivanje, storniranje i ažuriranje statusa).
//
// The buttons send the request source "btn", so the implementation of these actions must not use
// common.IsDataRequest (that check belongs to the grid only).
//
// TODO: not implemented yet - the actions have to be sent to the eFaktura (SEF) system and the
// status of the documents (status_salinv, datum_stat_salinv, cirinvoiceid, vatrecordingstatus,
// datum_stat_indvat) has to be stored back into rdok. Which of the actions was requested can be
// read from the route (c.Request.URL.Path).
func (h *RobnoDokumentaHandler) PregledEFakturaAkcija(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "eFaktura operacija još nije implementirana")
}

// Tab 3 - Specifikacije dokumenta
func (h *RobnoDokumentaHandler) SpecifikacijeDokumenta(c *gin.Context) {
	h.render(c, robnoDokumentaTabSpecifikacije)
}

// Tab 4 - Kontiranje dokumenata
//
// The tab has three sub-tabs, like the legacy screen: "Knjiženje dokumenata" (the documents of the
// selection with the actions of the posting), "Pregled proknjiženih / neproknjiženih dokumenata"
// and "Pregled proknjiženih / neproknjiženih dokumenata po magacinima" (the documents of the posted
// or of the not posted state).
func (h *RobnoDokumentaHandler) KontiranjeDokumenata(c *gin.Context) {
	h.kontiranjeKnjizenje(c)
}

// KontiranjeKnjizenje renders the "Knjiženje dokumenata" sub-tab of "Kontiranje dokumenata".
func (h *RobnoDokumentaHandler) KontiranjeKnjizenje(c *gin.Context) {
	h.kontiranjeKnjizenje(c)
}

// kontiranjeKnjizenje renders the "Knjiženje dokumenata" sub-tab: the parameters of the selection
// (vrsta naloga za knjiženje, vrsta dokumenta, magacin, ranges of the broj naloga, the broj
// dokumenta and the datum naloga), the buttons of the posting and the grid of the robni dokumenti.
// The "Obrada" button sends a data request, in which case only the grid is rendered.
func (h *RobnoDokumentaHandler) kontiranjeKnjizenje(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabKontiranje)
	if !ok {
		return
	}
	subTabs := h.subTabsFor(robnoDokumentaTabKontiranje)
	common.SetActiveTab(subTabs, robnoDokumentaSubTabKontiranjeKnjizenje)
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
		h.error(c, err)
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	vrstaDokumentaValues, err := h.service.GetVrstaDokumentaComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(robnoDokumentaURLKnjizenje, robnoDokumentaKnjizenjeTableID, hxValsRobnoDokumentaKnjizenje)
	btnRavnoteza := h.kontiranjeActionButton(robnoDokumentaRavnotezaBtnID, "Proveri ravnotežu", "fin_ravnoteza", robnoDokumentaURLKnjizenjeRavnot, hxValsRobnoDokumentaKnjizenje)
	btnKnjizi := h.kontiranjeActionButton(robnoDokumentaKontiranjeBtnID, "Knjiži", "fin_knjizenje", robnoDokumentaURLKnjizenjeKnjizi, hxValsRobnoDokumentaKnjizenje)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, robnoDokumentaURLKnjizenje, "#"+robnoDokumentaKnjizenjeTableID, hxValsRobnoDokumentaKnjizenje)
	if err := tmpl_robno.RobnoDokumentaKontiranjeKnjizenje(h.tabs, subTabs, tbl, tipdokValues, vrstaDokumentaValues, magValues, params, btnObrada, btnRavnoteza, btnKnjizi, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// getKontiranjeKnjizenje fills the grid of the "Knjiženje dokumenata" sub-tab (first the total
// records, then the rows of the page).
func (h *RobnoDokumentaHandler) getKontiranjeKnjizenje(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaKnjizenje
	for _, total := range []bool{true, false} {
		if err := h.service.GetKontiranjeKnjizenje(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// KontiranjePregled renders the "Pregled proknjiženih / neproknjiženih dokumenata" sub-tab of
// "Kontiranje dokumenata": the parameters of the selection, the state of the posting (the radio
// buttons "Proknjiženi dokumenti" / "Neproknjiženi dokumenti" and the checkbox that marks the
// displayed documents as not posted) and the grid of the robni dokumenti.
func (h *RobnoDokumentaHandler) KontiranjePregled(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabKontiranje)
	if !ok {
		return
	}
	subTabs := h.subTabsFor(robnoDokumentaTabKontiranje)
	common.SetActiveTab(subTabs, robnoDokumentaSubTabKontiranjePregled)
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
		h.error(c, err)
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(robnoDokumentaURLKontiranjePregled, robnoDokumentaKontiranjePregledTableID, hxValsRobnoDokumentaKontiranjePregled)
	// The action of the checkbox "Označi prikazana dokumenta kao neproknjižena..." of the legacy
	// screen: the checkbox is sent with the parameters of the grid, the marking of the displayed
	// documents is then executed by the action (TODO: not implemented yet).
	btnOznaci := h.kontiranjeActionButton(robnoDokumentaOznaciBtnID, "Označi kao neproknjižena", "cancel", robnoDokumentaURLOznaciNeproknjizene, hxValsRobnoDokumentaKontiranjePregled)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, robnoDokumentaURLKontiranjePregled, "#"+robnoDokumentaKontiranjePregledTableID, hxValsRobnoDokumentaKontiranjePregled)
	if err := tmpl_robno.RobnoDokumentaKontiranjePregled(h.tabs, subTabs, tbl, tipdokValues, magValues, params, btnObrada, btnOznaci, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// getKontiranjePregled fills the grid of the "Pregled proknjiženih / neproknjiženih dokumenata"
// sub-tab (first the total records, then the rows of the page).
func (h *RobnoDokumentaHandler) getKontiranjePregled(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaKontiranjePregled
	for _, total := range []bool{true, false} {
		if err := h.service.GetKontiranjePregled(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// KontiranjePoMagacinima renders the "Pregled proknjiženih / neproknjiženih dokumenata po
// magacinima" sub-tab of "Kontiranje dokumenata": the same filters as the previous sub-tab without
// the vrsta naloga (the rows are grouped by the magacin of the document).
func (h *RobnoDokumentaHandler) KontiranjePoMagacinima(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabKontiranje)
	if !ok {
		return
	}
	subTabs := h.subTabsFor(robnoDokumentaTabKontiranje)
	common.SetActiveTab(subTabs, robnoDokumentaSubTabKontiranjePoMagacinima)
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
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(robnoDokumentaURLPoMagacinima, robnoDokumentaPoMagacinimaTableID, hxValsRobnoDokumentaPoMagacinima)
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, robnoDokumentaURLPoMagacinima, "#"+robnoDokumentaPoMagacinimaTableID, hxValsRobnoDokumentaPoMagacinima)
	if err := tmpl_robno.RobnoDokumentaKontiranjePoMagacinima(h.tabs, subTabs, tbl, magValues, params, btnObrada, search, translator).Render(ctx, c.Writer); err != nil {
		h.error(c, err)
	}
}

// getKontiranjePoMagacinima fills the grid of the "Pregled ... po magacinima" sub-tab (first the
// total records, then the rows of the page).
func (h *RobnoDokumentaHandler) getKontiranjePoMagacinima(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaPoMagacinima
	for _, total := range []bool{true, false} {
		if err := h.service.GetKontiranjePoMagacinima(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// kontiranjeParams reads the parameters of the "Kontiranje dokumenata" tab from the request into the
// single params structure of the option (only the fields of the tab are filled). The ranges of the
// selection keep the defaults of the legacy screen (0 - 999999 for the broj naloga and the broj
// dokumenta and the last week for the datum naloga) and the state of the posting defaults to the
// posted documents of the radio buttons.
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

// kontiranjeActionButton builds one of the action buttons of the "Knjiženje dokumenata" sub-tab.
// The actions are posted without a request source (so the handler knows the request is not a data
// request of the grid) and the label is translated here because the generic button component
// (which renders the icon) does not translate it.
func (h *RobnoDokumentaHandler) kontiranjeActionButton(id, label, icon, url, hxVals string) domain.Button {
	return common.SetButton(id, i18n.GetInstance().Button(label), icon, url, "#"+robnoDokumentaInfoMessageID, "innerHTML", "POST", "", hxVals, true, common.ClassSaveButton, "handleDialogResponse")
}

// businessDaysBefore returns the date of the selected business year that lies the given number of
// days before its today (yyyy-mm-dd).
func (h *RobnoDokumentaHandler) businessDaysBefore(ctx context.Context, days int) string {
	today := time.Now()
	if session := domain.GetSessionFromStdContext(ctx); session != nil {
		today = time.Date(session.SelectedGod, today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	}
	return today.AddDate(0, 0, -days).Format(common.HtmlLayout)
}

// The actions of the "Knjiženje dokumenata" sub-tab.
//
// The buttons send the request source "btn", so the implementation of these actions must not use
// common.IsDataRequest (that check belongs to the grid only).
//
// TODO: not implemented yet - the posting of the robne dokumente creates the financial postings
// (fpro) of the kontiranje of the vrsta dokumenta (dokvrsta.kodknj) and sets the state of the
// document (rdok.knjige_1) and "Proveri ravnotežu" compares the duguje and the potražuje of those
// postings before the posting. The action that was requested can be read from the route
// (c.Request.URL.Path).
func (h *RobnoDokumentaHandler) KontiranjeKnjizi(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Knjiženje robnih dokumenata još nije implementirano")
}

// KontiranjeRavnoteza is the "Pr. ravnotežu" action of the "Knjiženje dokumenata" sub-tab.
//
// TODO: not implemented yet (see KontiranjeKnjizi).
func (h *RobnoDokumentaHandler) KontiranjeRavnoteza(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Provera ravnoteže još nije implementirana")
}

// KontiranjeOznaciNeproknjizene is the "Označi prikazana dokumenta kao neproknjižena..." action of
// the "Pregled proknjiženih / neproknjiženih dokumenata" sub-tab: it marks the displayed documents
// as not posted (rdok.knjige_1 = 'N').
//
// TODO: not implemented yet - the action has to unpost the displayed documents (and, like the
// legacy screen, remove the financial postings of their kontiranje).
func (h *RobnoDokumentaHandler) KontiranjeOznaciNeproknjizene(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Označavanje dokumenata kao neproknjiženih još nije implementirano")
}

// Tab 5 - Prepis dokumenta
func (h *RobnoDokumentaHandler) PrepisDokumenta(c *gin.Context) { h.render(c, robnoDokumentaTabPrepis) }

// Tab 6 - Prikaz ukupne obrade
//
// The tab has no parameters and no "Obrada" button: it shows the totals of the robni nalozi of
// every magacin of the current period and the print button of the list.
func (h *RobnoDokumentaHandler) PrikazUkupneObrade(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabPrikazUkupneObrade)
	if !ok {
		return
	}
	tbl := common.SetTableBasicData(robnoDokumentaUkupnaObradaTitle, robnoDokumentaUkupnaObradaTableID, h.service.GetPrikazUkupneObradeTableFields(), "", robnoDokumentaURLUkupnaObrada, 0, 0, 0, 0, h.cfg)
	// The toolbar of the grid holds the print button (there is no "Obrada" button on this tab).
	common.SetTableConfig(&tbl, robnoDokumentaUkupnaObradaTitle, robnoDokumentaURLUkupnaObrada, false, false, true)
	tbl.HasTotals = true

	translator := i18n.GetInstance()
	tbl.BtnPrint = common.SetPrintButton("ukupna-obrada-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLUkupnaObradaStampa, "GET", true, common.ClassPrintButton, "")
	tbl.BtnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	tbl.BtnPrint.HxSwap = "innerHTML"
	tbl.BtnPrint.HxOnAfterRequest = "handleDialogResponse"

	_, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	if common.IsDataRequest(c) {
		if err := h.service.GetPrikazUkupneObrade(ctx, &tbl, pageSize); err != nil {
			h.error(c, err)
			return
		}
		utils.RenderContent(c, tbl)
		return
	}
	if err := h.service.GetPrikazUkupneObrade(ctx, &tbl, pageSize); err != nil {
		h.error(c, err)
		return
	}
	if renderErr := tmpl_robno.RobnoDokumentaPrikazUkupneObrade(h.tabs, h.subTabsFor(robnoDokumentaTabPrikazUkupneObrade), tbl, translator).Render(ctx, c.Writer); renderErr != nil {
		h.error(c, renderErr)
	}
}

// PrikazUkupneObradePrint is the print of the "Prikaz ukupne obrade" tab.
//
// TODO: not implemented yet - the print preview has to be built (the report template with the
// columns of the grid and the header of the company, like the other robno reports).
func (h *RobnoDokumentaHandler) PrikazUkupneObradePrint(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Štampa prikaza ukupne obrade još nije implementirana")
}

// Tab 7 - Prikaz naloga
//
// The tab shows the robni nalozi of the current period filtered by magacin, the range of the vrste
// naloga and the range of the broj naloga. The filters "Po datumu naloga", "Po datumu obrade" and
// "Po operateru" are applied only when their checkbox is checked (their fields are disabled until
// then, like the legacy screen: the checkboxes toggle them with controlToggle).
func (h *RobnoDokumentaHandler) PrikazNaloga(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabPrikazNaloga)
	if !ok {
		return
	}
	params := h.prikazNalogaParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaPrikazNalogaTitle, robnoDokumentaPrikazNalogaTableID, h.service.GetPrikazNalogaTableFields(), "", robnoDokumentaURLPrikazNaloga, 0, 0, 0, 0, h.cfg)
	// The buttons of the tab ("Obrada" and "Štampaj") are in the row of the "Po operateru" checkbox,
	// so the toolbar of the grid holds only the title.
	common.SetTableConfig(&tbl, robnoDokumentaPrikazNalogaTitle, robnoDokumentaURLPrikazNaloga, false, false, false)
	if common.IsDataRequest(c) {
		if !h.getPrikazNaloga(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnPrint := common.SetPrintButton("prikaz-naloga-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLPrikazNalogaStampa, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	btnPrint.HxVals = hxValsRobnoDokumentaPrikazNaloga

	btnObrada := h.obradaButton(robnoDokumentaURLPrikazNaloga, robnoDokumentaPrikazNalogaTableID, hxValsRobnoDokumentaPrikazNaloga)
	if renderErr := tmpl_robno.RobnoDokumentaPrikazNaloga(h.tabs, h.subTabsFor(robnoDokumentaTabPrikazNaloga), tbl, magValues, tipdokValues, params, btnObrada, btnPrint, translator).Render(ctx, c.Writer); renderErr != nil {
		h.error(c, renderErr)
	}
}

// getPrikazNaloga fills the grid of the "Prikaz naloga" tab (first the total records, then the rows
// of the page).
func (h *RobnoDokumentaHandler) getPrikazNaloga(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaPrikazNaloga
	for _, total := range []bool{true, false} {
		if err := h.service.GetPrikazNaloga(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// prikazNalogaParams reads the parameters of the "Prikaz naloga" tab from the request into the
// single params structure of the option (only the fields of the tab are filled; the same selection
// of nalozi is read by the "Prikaz dokumenata u nalogu" and the "Prikaz dokumenata po operateru"
// tabs). The ranges of the broj naloga keep the defaults of the legacy screen (0 - 999999) and the
// dates of the two date filters the last week of the business year, like the other tabs of the
// option.
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

// PrikazNalogaPrint is the print of the "Prikaz naloga" tab.
//
// TODO: not implemented yet - the print preview has to be built (the report template with the
// columns of the grid and the header of the company, like the other robno reports).
func (h *RobnoDokumentaHandler) PrikazNalogaPrint(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Štampa prikaza naloga još nije implementirana")
}

// Tab 8 - Prikaz dokumenata u nalogu
//
// The tab shows the robni dokumenti (rdok) of the nalozi selected with the parameters of the tab.
// The parameters are the same selection of nalozi as the "Prikaz naloga" tab (they are read with
// prikazNalogaParams) and its filters "Po datumu naloga", "Po datumu obrade" and "Po operateru"
// are applied only when their checkbox is checked.
func (h *RobnoDokumentaHandler) PrikazDokumenataUNalogu(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabPrikazDokumenataUNalogu)
	if !ok {
		return
	}
	params := h.prikazNalogaParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaUNaloguTitle, robnoDokumentaUNaloguTableID, h.service.GetPrikazDokumenataUNaloguTableFields(), "", robnoDokumentaURLUNalogu, 0, 0, 0, 0, h.cfg)
	// The buttons of the tab ("Obrada" and "Štampaj") are in the parameter panel, so the toolbar of
	// the grid holds only the title.
	common.SetTableConfig(&tbl, robnoDokumentaUNaloguTitle, robnoDokumentaURLUNalogu, false, false, false)
	if common.IsDataRequest(c) {
		if !h.getPrikazDokumenataUNalogu(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	magValues, err := h.service.GetMagacinComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}
	tipdokValues, err := h.service.GetTipdokComboValues(ctx)
	if err != nil {
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnPrint := common.SetPrintButton("u-nalogu-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLUNaloguStampa, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	btnPrint.HxVals = hxValsRobnoDokumentaUNalogu

	btnObrada := h.obradaButton(robnoDokumentaURLUNalogu, robnoDokumentaUNaloguTableID, hxValsRobnoDokumentaUNalogu)
	if renderErr := tmpl_robno.RobnoDokumentaPrikazDokumenataUNalogu(h.tabs, h.subTabsFor(robnoDokumentaTabPrikazDokumenataUNalogu), tbl, magValues, tipdokValues, params, btnObrada, btnPrint, translator).Render(ctx, c.Writer); renderErr != nil {
		h.error(c, renderErr)
	}
}

// getPrikazDokumenataUNalogu fills the grid of the "Prikaz dokumenata u nalogu" tab (first the total
// records, then the rows of the page).
func (h *RobnoDokumentaHandler) getPrikazDokumenataUNalogu(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaUNalogu
	for _, total := range []bool{true, false} {
		if err := h.service.GetPrikazDokumenataUNalogu(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// PrikazDokumenataUNaloguPrint is the print of the "Prikaz dokumenata u nalogu" tab.
//
// TODO: not implemented yet - the print preview has to be built (the report template with the
// columns of the grid and the header of the company, like the other robno reports).
func (h *RobnoDokumentaHandler) PrikazDokumenataUNaloguPrint(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Štampa prikaza dokumenata u nalogu još nije implementirana")
}

// Tab 9 - Prikaz dokumenata po operateru
//
// The tab shows the same robni dokumenti as the "Prikaz dokumenata u nalogu" tab, grouped by the
// operater of the document. Its panel has no magacin and no vrsta naloga selection (the legacy screen
// of the tab has none either): the parameters are the range of the broj naloga and the three optional
// filters "Po datumu naloga", "Po datumu obrade" and "Po operateru".
func (h *RobnoDokumentaHandler) PrikazDokumenataPooperateru(c *gin.Context) {
	ctx, _, ok := h.context(c, robnoDokumentaTabPrikazDokumenataPooperateru)
	if !ok {
		return
	}
	params := h.prikazNalogaParams(c)

	tbl := common.SetTableBasicData(robnoDokumentaPooperateruTitle, robnoDokumentaPooperateruTableID, h.service.GetPrikazDokumenataPooperateruTableFields(), "", robnoDokumentaURLPooperateru, 0, 0, 0, 0, h.cfg)
	// The buttons of the tab ("Obrada" and "Štampaj") are in the parameter panel, so the toolbar of
	// the grid holds only the title.
	common.SetTableConfig(&tbl, robnoDokumentaPooperateruTitle, robnoDokumentaURLPooperateru, false, false, false)
	if common.IsDataRequest(c) {
		if !h.getPrikazDokumenataPooperateru(c, &tbl, params) {
			return
		}
		utils.RenderContent(c, tbl)
		return
	}

	translator := i18n.GetInstance()
	btnPrint := common.SetPrintButton("po-operateru-stampa-btn", "Štampaj", "stampa", robnoDokumentaURLPooperateruStampa, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	btnPrint.HxTarget = "#" + robnoDokumentaInfoMessageID
	btnPrint.HxSwap = "innerHTML"
	btnPrint.HxOnAfterRequest = "handleDialogResponse"
	btnPrint.HxVals = hxValsRobnoDokumentaPooperateru

	btnObrada := h.obradaButton(robnoDokumentaURLPooperateru, robnoDokumentaPooperateruTableID, hxValsRobnoDokumentaPooperateru)
	if renderErr := tmpl_robno.RobnoDokumentaPrikazDokumenataPooperateru(h.tabs, h.subTabsFor(robnoDokumentaTabPrikazDokumenataPooperateru), tbl, params, btnObrada, btnPrint, translator).Render(ctx, c.Writer); renderErr != nil {
		h.error(c, renderErr)
	}
}

// getPrikazDokumenataPooperateru fills the grid of the "Prikaz dokumenata po operateru" tab (first
// the total records, then the rows of the page).
func (h *RobnoDokumentaHandler) getPrikazDokumenataPooperateru(c *gin.Context, tbl *domain.TableData, params domain.RobnoDokumentaParams) bool {
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	tbl.Pagination.HxVals = hxValsRobnoDokumentaPooperateru
	for _, total := range []bool{true, false} {
		if err := h.service.GetPrikazDokumenataPooperateru(c.Request.Context(), tbl, total, pageSize, page, params); err != nil {
			h.error(c, err)
			return false
		}
	}
	return true
}

// PrikazDokumenataPooperateruPrint is the print of the "Prikaz dokumenata po operateru" tab.
//
// TODO: not implemented yet - the print preview has to be built (the report template with the
// columns of the grid and the header of the company, like the other robno reports).
func (h *RobnoDokumentaHandler) PrikazDokumenataPooperateruPrint(c *gin.Context) {
	common.WriteJSONResponse(c, http.StatusNotImplemented, false, nil, "Štampa prikaza dokumenata po operateru još nije implementirana")
}

// render builds the (still empty) table of one tab and renders its template.
//
// TODO: as every single tab is implemented, its data query is added to the service and called
// here (both for the initial render and for the HTMX data request), together with the parameters
// and the print field list of the tab.
func (h *RobnoDokumentaHandler) render(c *gin.Context, tabIndex int) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "no user session found")
		return
	}
	common.SetActiveTab(h.tabs, tabIndex)

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
		h.error(c, err)
		return
	}

	translator := i18n.GetInstance()
	btnObrada := h.obradaButton(url, tableID, "")
	// The print of every tab receives the whole selection of the option (the same print contract).
	btnPrint := common.SetPrintButton(tableID+"-stampa", "Štampa", "stampa", printURL, "GET", true, common.ClassPrintButton, robnoDokumentaPrintFlds)
	// TODO: pass the hxVals of the tab instead of "".
	search := common.CreateSearchInput(robnoDokumentaSearchInputID, translator, url, "#"+tableID, "")

	if renderErr := h.renderTab(c, tabIndex, tbl, magValues, btnObrada, btnPrint, search); renderErr != nil {
		h.error(c, renderErr)
	}
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
//
// TODO: when a tab gets its own content (parameters, master/detail tables, ...) render the
// template with the additional data it needs.
func (h *RobnoDokumentaHandler) renderTab(c *gin.Context, tabIndex int, tbl domain.TableData, magValues []domain.ComboItem, btnObrada, btnPrint domain.Button, search domain.InputControl) error {
	ctx := c.Request.Context()
	tabs := h.tabs
	subTabs := h.subTabsFor(tabIndex)
	translator := i18n.GetInstance()
	switch tabIndex {
	case robnoDokumentaTabSpecifikacije:
		return tmpl_robno.RobnoDokumentaSpecifikacije(tabs, subTabs, tbl, magValues, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer)
	case robnoDokumentaTabPrepis:
		return tmpl_robno.RobnoDokumentaPrepis(tabs, subTabs, tbl, magValues, btnObrada, btnPrint, search, translator).Render(ctx, c.Writer)
	default:
		return fmt.Errorf("unknown tab %d of the Robna dokumenta option", tabIndex)
	}
}

// subTabsFor returns the sub-tabs of the given tab (an empty set when the tab has no sub-tabs).
func (h *RobnoDokumentaHandler) subTabsFor(tabIndex int) domain.TabData {
	if tabIndex >= 0 && tabIndex < len(h.subTabs) {
		return h.subTabs[tabIndex]
	}
	return domain.TabData{}
}

// obradaButton builds the "Obrada" button of a tab. The button sends the parameters of the tab
// (hxVals) and the request source "btnobrada", so the handler renders only the grid.
func (h *RobnoDokumentaHandler) obradaButton(url, tableID, hxVals string) domain.Button {
	return common.SetButton(robnoDokumentaObradaBtnID, "Obrada", "obrada", url, "#"+tableID, "innerHTML", "GET", "", hxVals, true, common.ClassSaveButton, "handleDialogResponse")
}

func (h *RobnoDokumentaHandler) error(c *gin.Context, err error) {
	common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, err.Error())
}

func (h *RobnoDokumentaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta", h.RobnoDokumentaMain)
	r.GET("/api/robno-dokumenta/unos", h.RobnoDokumentaMain)
	r.GET("/api/robno-dokumenta/pregled", h.PregledDokumenata)
	r.GET("/api/robno-dokumenta/pregled/stampa", h.PregledStampa)
	r.GET("/api/robno-dokumenta/pregled/stampa/print", h.PregledStampaPrint)
	r.GET("/api/robno-dokumenta/pregled/efaktura", h.PregledEFaktura)
	r.GET("/api/robno-dokumenta/pregled/efaktura/print", h.PregledEFakturaPrint)
	r.POST("/api/robno-dokumenta/pregled/efaktura/posalji", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/proveri-status", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/otkazi", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/storniraj", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/storniraj-pe", h.PregledEFakturaAkcija)
	r.POST("/api/robno-dokumenta/pregled/efaktura/status", h.PregledEFakturaAkcija)
	r.GET("/api/robno-dokumenta/specifikacije", h.SpecifikacijeDokumenta)
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
	r.GET("/api/robno-dokumenta/po-operateru", h.PrikazDokumenataPooperateru)
	r.GET("/api/robno-dokumenta/po-operateru/stampa", h.PrikazDokumenataPooperateruPrint)
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

	// TODO: add the stampa (print) routes of the tabs together with their print templates.
}

// robnoDokumentaTabs defines the tabs of the "Robna dokumenta" option, in the order of the
// legacy menu: Unos dokumenta, Pregled dokumenta, Specifikacije dokumenta, Kontiranje dokumenata,
// Prepis dokumenta, Prikaz ukupne obrade, Prikaz naloga, Prikaz dokumenata u nalogu and
// Prikaz dokumenata po operateru.
func robnoDokumentaTabs() domain.TabData {
	translator := i18n.GetInstance()
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
// A tab without sub-tabs keeps an empty set, in which case the sub-tab bar is not rendered.
//
// TODO: fill the sets of the remaining tabs that have a sub-menu in the legacy menu
// ("Unos dokumenta", "Prepis dokumenta" and "Prikaz naloga"), e.g.:
//
//	{ID: "robno-dokumenta-unos-ulaz", Label: translator.T("..."), HXRequestUrl: "...", Name: "..."}
func robnoDokumentaSubTabs() []domain.TabData {
	translator := i18n.GetInstance()
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
