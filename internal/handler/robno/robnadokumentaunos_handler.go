package robno

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"helia/config"
	tmpl "helia/frontend/templates"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	commonsvc "helia/internal/service/common"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RobnoDokumentaUnosHandler handles the entry of the robni dokumenti of a robni nalog: the screens the
// legacy "Snimi nalog" of the "Unos dokumenta" tab opens after the nalog is saved, one per group of the
// vrsta dokumenta (dokvrsta.grpdok, see robnoDokumentaUnosEkrani and OtvoriDokument). The screens are
// added one by one; the "Fakture veleprodaje" screen (Rob_Fakture of
// frontend/templates/robno/robnadokumentaunos.templ) is opened by otvoriFakturu.
type RobnoDokumentaUnosHandler struct {
	translator *i18n.Service
	service    robnosvc.RobnoDokumentaService
	cfg        config.Config
	lm         *middleware.LockMiddleware
	ls         *middleware.LockService
}

// NewRobnoDokumentaUnosHandler creates the handler of the entry of the robni dokumenti.
func NewRobnoDokumentaUnosHandler(s robnosvc.RobnoDokumentaService, cfg config.Config, lm *middleware.LockMiddleware, ls *middleware.LockService, translator *i18n.Service) *RobnoDokumentaUnosHandler {
	return &RobnoDokumentaUnosHandler{translator: translator, service: s, cfg: cfg, lm: lm, ls: ls}
}

const (
	// robnoDokumentaURLOtvoriDokument opens the screen of the entry of the documents of a nalog by the
	// group of its vrsta dokumenta (see OtvoriDokument).
	robnoDokumentaURLOtvoriDokument = robnoDokumentaURLPrefix + "/unos-dokumenta/otvori"
	// robnoDokumentaDokumentInfoDialogID is the id of the message dialogs of OtvoriDokument (e.g.
	// "Program za ovu opciju nije instaliran"): its own id, so its OK button closes this dialog and not
	// another message dialog of the tab that is still in the page.
	robnoDokumentaDokumentInfoDialogID = "robno-dokumenta-dokument-info"
	// robnoDokumentaDokumentDialogID is the container of the dialog of the screen of the entry of the
	// documents of a nalog in the tab "Unos dokumenta" (RobnoDokumentaMain; filled by
	// RobnoDokumentaOtvoriDokumentScript).
	robnoDokumentaDokumentDialogID = "robno-dokumenta-dokument-dialog"
)

// The declarations of the "Fakture veleprodaje" screen (the templates RobnoFakture and
// RobnoFaktureDialog of frontend/templates/robno/robnadokumentaunos.templ). The template package cannot
// import this package, so every url, element id and hx-vals string the markup of the screen uses is
// declared here: the urls and the ids reach the templates as domain.RobnoFaktureUI (see
// robnoFaktureUI) and the two hx-vals strings of the header of the nalog as plain arguments, while
// robnoFaktureButtonsFor builds the buttons of the screen with the same constants.
const (
	// URLs of the screen. TODO (temporary): the handler and the routes of the fakture are not written
	// yet; the urls are the ones the screen will use (the prefix of the other tabs of "Robna
	// dokumenta").
	robnoFaktureURL     = robnoDokumentaURLPrefix + "/fakture"
	robnoFaktureSaveURL = robnoFaktureURL + "/save"
	robnoFaktureNoviURL = robnoFaktureURL + "/novi"
	// robnoFaktureDeleteURL + "/:id" deletes the document (DELETE), robnoFaktureDokPotvrdaURL asks to confirm it.
	robnoFaktureDeleteURL     = robnoFaktureURL + "/brisi"
	robnoFaktureDokPotvrdaURL = robnoFaktureDeleteURL + "/potvrda"
	robnoFaktureStavkaSaveURL = robnoFaktureURL + "/stavka/save"
	robnoFaktureStavkaDelURL  = robnoFaktureURL + "/stavka/brisi"
	// robnoFaktureStavkaDelURL + "/:id" deletes the stavka of the id (DELETE: "Da" of the confirmation).
	// robnoFaktureStavkaURL returns a saved stavka for the form (GET rdokid, id: "Izmeni" of its row),
	// robnoFaktureArtikalURL the artikal of the stavka (GET rdokid, sifra, konto: the choice of the artikal),
	// robnoFaktureStavkaPotvrdaURL the confirmation of the deletion of a stavka (GET rdokid, id: "Obriši" of
	// its row).
	robnoFaktureStavkaURL        = robnoFaktureURL + "/stavka"
	robnoFaktureArtikalURL       = robnoFaktureURL + "/artikal"
	robnoFaktureStavkaPotvrdaURL = robnoFaktureURL + "/stavka/potvrda"
	// robnoFaktureStavkaPotvrdaDialogID is the dialog of the confirmation of the deletion of a stavka.
	robnoFaktureStavkaPotvrdaDialogID = "robno-fakture-stavka-brisanje"
	// robnoFaktureDokPotvrdaDialogID is the confirmation of the deletion of the document.
	robnoFaktureDokPotvrdaDialogID = "robno-fakture-dok-brisanje"
	robnoFaktureArtikalSearch      = "/api/promet/searchbutton"
	robnoFakturePartnerSearch      = "/api/partneri/searchbutton"
	// TODO: the endpoint of the search of the robni dokumenti is not written yet (neither is the one of
	// robnoFakturePartnerSearch); the button of the broj dokumenta already sends the request here.
	robnoFaktureDokumentSearch = "/api/robno-dokumenta/searchbutton"

	// Ids of the two collapsible controls of the screen. They are given to the script of the screen
	// (RobnoFaktureScript) and to the buttons (robnoFaktureButtonsFor), so that every id is written in
	// one place only.
	robnoFaktureHeaderPanelID = "robno-fakture-header-panel"
	robnoFaktureStavkePanelID = "robno-fakture-stavke-panel"

	// robnoFaktureDialogID is the id of the dialog of the screen (Rob_Fakture, the one closeDialog hides).
	robnoFaktureDialogID = "robno-fakture-dialog"
	// robnoFaktureDokumentURL + "/:id" reloads the content of the screen with an existing document
	// (FaktureDokument), robnoFaktureNoviURL with a new one (FaktureNovi); robnoFaktureBodyID is
	// that content (RobnoFaktureBody) and robnoFaktureInfoDialogID the dialog of its messages.
	robnoFaktureDokumentURL  = robnoFaktureURL + "/dokument"
	robnoFaktureInfoDialogID = "robno-fakture-info-dialog"
	// robnoFakturePrikazURL fills the tab "Prikaz faktura" (FakturePrikaz), robnoFakturePrikazStavkeURL +
	// "/:id" the grid of the stavke of its selected document (FakturePrikazStavke).
	robnoFakturePrikazURL       = robnoFaktureURL + "/prikaz"
	robnoFakturePrikazStavkeURL = robnoFakturePrikazURL + "/stavke"
	// robnoFakturePrikazDokumentaURL reloads the grid of the documents of the tab (the pagination, the
	// search and the sort; FakturePrikazDokumenta).
	robnoFakturePrikazDokumentaURL = robnoFakturePrikazURL + "/dokumenta"

	// Ids of the elements of the screen (the templates get them by robnoFaktureUI):
	// robnoFaktureBodyID is the content of the tab "Fakture" the screen reloads with the selected document,
	// robnoFaktureInfoID the element of the messages of the screen (RobnoFaktureInfo, robnoFaktureGreska),
	// robnoFaktureStavkaInfoID the container of the dialog robnoFaktureStavkaInfoDialogID of the messages
	// of the stavke (e.g. the error of the save of a stavka).
	robnoFaktureBodyID             = "robno-fakture-body"
	robnoFaktureInfoID             = "robno-fakture-info"
	robnoFaktureStavkaInfoID       = "robno-fakture-stavka-info"
	robnoFaktureStavkaInfoDialogID = "robno-fakture-stavka-info-dialog"
	// The tab "Prikaz faktura": its content (reloaded when the tab is shown), the container of the grid of
	// the stavke of the selected document, the search of the documents, the hidden rdokid of the selected
	// document and the print button that prints it.
	robnoFakturePrikazID       = "robno-fakture-prikaz"
	robnoFakturePrikazStavkeID = "robno-fakture-prikaz-stavke"
	robnoFakturePrikazSearchID = "robno-fakture-prikaz-search"
	robnoFakturePrikazRdokID   = "robno-fakture-prikaz-rdokid"
	robnoFakturePrikazStampaID = "robno-fakture-prikaz-stampa"

	// Ids of the grids of the screen: the stavke and the avansi of the tab "Fakture", the documents of
	// the nalog and the stavke of the selected one of the tab "Prikaz faktura".
	robnoFaktureStavkeTableID         = "robno-fakture-stavke-table"
	robnoFaktureAvansiTableID         = "robno-fakture-avansi-table"
	robnoFaktureDokumentaTableID      = "robno-fakture-dokumenta-table"
	robnoFaktureDokumentStavkeTableID = "robno-fakture-dokument-stavke-table"
)

// robnoFaktureUI groups the urls and the element ids of the "Fakture veleprodaje" screen.
func robnoFaktureUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoFaktureHeaderPanelID,
		StavkePanelID:     robnoFaktureStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoFaktureDialogID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoFaktureDokumentURL,
		NoviURL:           robnoFaktureNoviURL,
		PrikazURL:         robnoFakturePrikazURL,
		ArtikalURL:        robnoFaktureArtikalURL,
		BodyID:            robnoFaktureBodyID,
		InfoID:            robnoFaktureInfoID,
		StavkaInfoID:      robnoFaktureStavkaInfoID,
		PrikazID:          robnoFakturePrikazID,
		PrikazStavkeID:    robnoFakturePrikazStavkeID,
		PrikazRdokID:      robnoFakturePrikazRdokID,
		PrikazStampaID:    robnoFakturePrikazStampaID,
	}
}

// robnoFaktureButtonsFor builds the buttons of the "Fakture veleprodaje" screen.
func robnoFaktureButtonsFor() domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:            "robno-fakture-sacuvaj",
			LabelText:     "Sačuvaj",
			Icon:          "save",
			BtnClass:      common.ClassSaveButton,
			HxActionURL:   robnoFaktureSaveURL,
			HxRequestType: "POST",
			// Only the fields of the header are sent (the read only values of the strips are disabled,
			// so the browser does not send them; the handler reads them from the dokument/kupac).
			HxInclude: "#" + robnoFaktureHeaderPanelID + " [data-panel-fields]",
			// The handler answers with the content of the tab with the saved document (the header
			// locked, the stavke open), or with the JSON of the field errors (status 422, not
			// swapped), shown by the script of the screen (robnoFaktureAfterHeaderSave).
			HxTarget:             "#" + robnoFaktureBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoFaktureHeaderPanelID, robnoFaktureStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-fakture-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoFaktureHeaderPanelID, robnoFaktureStavkePanelID},
		},
		New: domain.Button{
			Id:               "robno-fakture-novi",
			LabelText:        "Novi dok.",
			Icon:             "add",
			BtnClass:         common.ClassNewButton,
			HxActionURL:      robnoFaktureNoviURL,
			HxRequestType:    "GET",
			HxTarget:         "#" + robnoDokumentaContentID,
			HxSwap:           "innerHTML",
			HxOnAfterRequest: "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{
				robnoFaktureHeaderPanelID, robnoFaktureStavkePanelID},
		},
		// "Briši" asks to confirm the deletion of the document (its url gets the rdokid in robnoFaktureEkran).
		Delete: domain.Button{
			Id:            "robno-fakture-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   robnoFaktureDokPotvrdaURL,
			HxRequestType: "GET",
			HxTarget:      "#" + robnoFaktureInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:        "robno-fakture-nazad",
			LabelText: "Nazad",
			Icon:      "back",
			BtnClass:  common.ClassCloseButton,
			// TODO: the route of the "Fakture veleprodaje" menu entry
			HxActionURL:   "/api/robna-dokumenta",
			HxRequestType: "GET",
			HxTarget:      "#main-content",
			HxSwap:        "innerHTML",
		},
		// "Štampaj": the print of the document of the form, like the "Štampa" sub-tab of "Pregled dokumenta"
		// (the print of its vrsta dokumenta), with the rdokid of the form instead of the selected row (the
		// hidden rdokid of the stavke, id stavkardokid).
		Print: common.SetPrintButton("robno-fakture-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		// "Odustani" of the change of a stavka (like "Odustani" of the stavke of a financial nalog): disabled
		// until "Izmeni" of a row, it unlocks the grid of the stavke and disables itself again.
		CancelStavka: domain.Button{
			Id:         "robno-fakture-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robnoFaktureStavkaOdustani",
		},
		// "Sačuvaj" of the stavka: the content of the tab with the saved stavka (the grid, the iznos of the
		// document, the strip of the nalog), or the JSON of the field errors (status 422, not swapped).
		SaveStavka: domain.Button{
			Id:                   "robno-fakture-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoFaktureStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoFaktureStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoFaktureBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
	}
}

// robnoDokumentaUnosEkran is the screen of the entry of the documents of a group of vrste dokumenta:
// the legacy window (for the message while it is not written yet) and the handler that opens it (nil
// while the screen is not written).
type robnoDokumentaUnosEkran struct {
	legacy string
	open   func(h *RobnoDokumentaUnosHandler, c *gin.Context, dok robnoDokumentaUnosDokument)
}

// robnoDokumentaUnosDokument is what a screen of the entry needs: the nalog (rnal), the magacin and the
// vrsta dokumenta selected in the header of the "Unos dokumenta" tab.
type robnoDokumentaUnosDokument struct {
	RnalID     int64
	MagaciniID int
	Vrd        int
	Grpdok     string
	Tipdok     string
	Nalog      int
}

// robnoDokumentaUnosEkrani are the screens of the entry by the group of the vrsta dokumenta
// (dokvrsta.grpdok), like the SWITCH of the legacy "Snimi nalog". A group gets its open function when its
// screen is written: FAK, FRP and KPK the faktura veleprodaje (the legacy rob_Fakture; with
// fvr.etunosprod = 2 the legacy pr_FaktureOtp), IRS only from a magacin of the proizvodnja (tipmag P or
// L).
var robnoDokumentaUnosEkrani = map[string]robnoDokumentaUnosEkran{
	"FAK": {legacy: "rob_Fakture", open: (*RobnoDokumentaUnosHandler).otvoriFakturu},
	"FRP": {legacy: "rob_Fakture", open: (*RobnoDokumentaUnosHandler).otvoriFakturu},
	// The knjižna odobrenja and zaduženja get their own screen (and validation) later.
	"KPK": {legacy: "rob_Fakture"},
	"FZR": {legacy: "rob_Fakture", open: (*RobnoDokumentaUnosHandler).otvoriFakturu},
	"DIR": {legacy: "rob_Fakture_maloprodaje"},
	// The interni prenosi - dostavnice (the legacy rob_Interni_racuni).
	"IRT": {legacy: "rob_Interni_racuni", open: (*RobnoDokumentaUnosHandler).otvoriInterniRacun},
	"IRP": {legacy: "rob_Interni_racuni", open: (*RobnoDokumentaUnosHandler).otvoriInterniRacun},
	"IRS": {legacy: "rob_Interni_racuni"},
	"PRE": {legacy: "rob_Fakture_Usluga", open: (*RobnoDokumentaUnosHandler).otvoriFakturuUsluga},
	"FUR": {legacy: "rob_Fakture_Usluga", open: (*RobnoDokumentaUnosHandler).otvoriFakturuUsluga},
	"FCO": {legacy: "rob_Fakture_Usluga", open: (*RobnoDokumentaUnosHandler).otvoriFakturuUsluga},
	"RCO": {legacy: "rob_Fakture_Usluga", open: (*RobnoDokumentaUnosHandler).otvoriFakturuUsluga},
	"KPF": {legacy: "rob_Fakture_Usluga", open: (*RobnoDokumentaUnosHandler).otvoriFakturuUsluga},
	"OPD": {legacy: "rob_opsti_dokumenti_opd", open: (*RobnoDokumentaUnosHandler).otvoriOpstiDokument},
	"POP": {legacy: "rob_opsti_dokumenti_opd", open: (*RobnoDokumentaUnosHandler).otvoriOpstiDokument},
	"POT": {legacy: "rob_opsti_dokumenti_opd", open: (*RobnoDokumentaUnosHandler).otvoriOpstiDokument},
	"FIN": {legacy: "rob_opsti_dokumenti_FIN", open: (*RobnoDokumentaUnosHandler).otvoriOpstiDokument},
	"KOL": {legacy: "rob_opsti_dokumenti_KOL", open: (*RobnoDokumentaUnosHandler).otvoriOpstiDokument},
	"MNV": {legacy: "rob_opsti_dokumenti_MNV", open: (*RobnoDokumentaUnosHandler).otvoriOpstiDokument},
	"NIV": {legacy: "rob_Nivelacije_Veleprodaja", open: (*RobnoDokumentaUnosHandler).otvoriOpstiDokument},
	"ARA": {legacy: "rob_AvansniRacuni", open: (*RobnoDokumentaUnosHandler).otvoriFakturuUsluga},
	"ARU": {legacy: "rob_AvansniRacuni", open: (*RobnoDokumentaUnosHandler).otvoriFakturuUsluga},
	"PLT": {legacy: "rob_Kalkulacije_Veleprodaja"},
	"KAL": {legacy: "rob_Kalkulacije_Maloprodaja"},
	"PRI": {legacy: "rob_Prenosnice"},
	"PRU": {legacy: "rob_Prenosnice"},
	"GRD": {legacy: "rob_ZadauzGRADILISTA"},
	"PRO": {legacy: "rob_ProFakture"},
	"POR": {legacy: "rob_Porudzbenica"},
	"OTR": {legacy: "rob_Trebovanja"},
	"KNO": {legacy: "rob_KnjiznoPismoFIN"},
	"KNZ": {legacy: "rob_KnjiznoPismoFIN"},
	"PTG": {legacy: "rob_opsti_dokPOPIS"},
	"SIV": {legacy: "rob_ZaduzSitanInventar"},
	"MNI": {legacy: "rob_Nivelacija_maloprodaje"},
}

// OtvoriDokument opens the screen of the entry of the documents of a robni nalog (GET with rnalid,
// magaciniid and vrd), like the legacy "Snimi nalog" after the nalog is saved: the screen of the group
// of the vrsta dokumenta (robnoDokumentaUnosEkrani) as a dialog over the "Unos dokumenta" tab, or the
// message "Program za ovu opciju nije instaliran" for a group without a screen. The page calls it after
// the save of the nalog (see robnoDokumentaOtvoriDokumentTrigger). A screen keeps the nalog locked and
// releases it with robnoDokumentaURLUnlockNalog when it is closed; without a screen the lock is released
// here.
func (h *RobnoDokumentaUnosHandler) OtvoriDokument(c *gin.Context) {
	ctx := c.Request.Context()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, common.ErrMsgSessionNotFound)
		return
	}
	dok := robnoDokumentaUnosDokument{
		RnalID:     int64(common.StringToInt(c.Query("rnalid"))),
		MagaciniID: common.StringToInt(c.Query("magaciniid")),
		Vrd:        common.StringToInt(c.Query("vrd")),
		Tipdok:     strings.TrimSpace(c.Query("tipdok")),
		Nalog:      common.StringToInt(c.Query("nalog")),
	}
	grpdok, err := h.service.GetGrupaDokumenta(ctx, dok.Vrd)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	dok.Grpdok = strings.ToUpper(strings.TrimSpace(grpdok))
	ekran, found := robnoDokumentaUnosEkrani[dok.Grpdok]
	if !found || ekran.open == nil {
		// No screen holds the nalog: its lock (taken by the save of the nalog) is released.
		if dok.RnalID > 0 && h.ls != nil {
			_ = h.ls.Unlock(ctx, robnoDokumentaEntityType, dok.RnalID, userSession.UserName)
		}
		msg := h.translator.Message("Program za ovu opciju nije instaliran")
		if found {
			msg += " (" + ekran.legacy + ")"
		}
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, msg)
		return
	}
	ekran.open(h, c, dok)
}

// otvoriFakturu opens the "Fakture veleprodaje" screen (the legacy rob_Fakture) of the nalog as a
// dialog with its two tabs, with a new document (the next broj of the vrsta in the magacin). The nalog
// stays locked while the dialog is open: "Zatvori" and "Nazad" release it (robnoDokumentaURLUnlockNalog)
// and close the dialog.
func (h *RobnoDokumentaUnosHandler) otvoriFakturu(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoFaktureEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	// The grids of the tab "Prikaz faktura" are filled when the tab is shown (FakturePrikaz).
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoFakturePrikazOpis(), nil)

	// "Zatvori" of the title bar releases the lock of the nalog, like the buttons of the dialog of the
	// stavke of a financial nalog.
	btnClose := domain.Button{
		Id:            "robno-fakture-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoFaktureDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_Fakture(robnoFaktureUI(), ekran.tbl, ekran.avansiTbl, dokumentaTbl, dokumentStavkeTbl, ekran.header, ekran.stavka,
		ekran.vrstaDokumentaValues, nil, robnosvc.RobSistemPdvCombo(), ekran.ziroRacuni, ekran.btns, btnClose, ekran.search, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
	}
}

// robnoPrikaz describes the second tab of a screen of the robni dokumenti (the faktura: "Prikaz faktura", the
// interni prenos: "Pregled internih prenosa"): its ids and urls (ui, the grids, the search, the routes of the
// grid and of the stavke), its columns and rows, the vrste dokumenata of its grid (by the vrsta of the
// screen) and how an error is shown. The tab is filled when it is shown (robnoPrikazTab), its grid is
// reloaded by the pagination, the search and the sort (robnoPrikazGrid), and a click of a row loads the
// stavke of the document (robnoPrikazStavke).
type robnoPrikaz struct {
	ui                              domain.RobnoFaktureUI
	dokumentaTableID, stavkeTableID string
	dokumentaNaslov, stavkeNaslov   string
	dokumentaFields, stavkeFields   []domain.Fields
	dokumentaRows                   func([]domain.RobDokumentPrikazRow) []domain.TableRow
	stavkeRows                      func([]domain.RobStavkaRow) []domain.TableRow
	dokumentaURL, stavkeURL         string
	searchID                        string
	vrste                           func(vrd int) []int
	greska                          func(c *gin.Context, message string)
}

// robnoFakturePrikazOpis is the tab "Prikaz faktura": the documents of the vrsta of the faktura.
func (h *RobnoDokumentaUnosHandler) robnoFakturePrikazOpis() robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoFaktureUI(),
		dokumentaTableID: robnoFaktureDokumentaTableID,
		stavkeTableID:    robnoFaktureDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetFaktureDokumentaTableFields(),
		stavkeFields:     h.service.GetFaktureDokumentStavkeTableFields(),
		dokumentaRows:    robnosvc.RobDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobStavkePrikazTableRows,
		dokumentaURL:     robnoFakturePrikazDokumentaURL,
		stavkeURL:        robnoFakturePrikazStavkeURL,
		searchID:         robnoFakturePrikazSearchID,
		vrste:            func(vrd int) []int { return []int{vrd} },
		greska:           robnoFaktureGreska,
	}
}

// robnoPrikazTabele returns the grids of the second tab: the documents (a click of a row loads its stavke
// into the grid of the stavke and selects the document "Štampaj" prints) and the empty grid of the stavke.
func (h *RobnoDokumentaUnosHandler) robnoPrikazTabele(p robnoPrikaz, dokumenta []domain.RobDokumentPrikazRow) (domain.TableData, domain.TableData) {
	dokumentaTbl := h.robnoFaktureTable(p.dokumentaNaslov, p.dokumentaTableID, p.dokumentaFields)
	dokumentaTbl.Rows = p.dokumentaRows(dokumenta)
	dokumentaTbl.DetailURL = p.stavkeURL
	dokumentaTbl.DetailTarget = "#" + p.ui.PrikazStavkeID
	dokumentaTbl.DetailHxRequestType = "GET"
	dokumentaTbl.DetailHxSwap = "innerHTML"
	dokumentaTbl.DetailHxTrigger = "click"
	// The click of a row also selects the document to print (robnoFakturePrikazIzbor: its rdokid in the
	// hidden ui.PrikazRdokID). "Štampaj" after the search prints it like "Štampaj" of the first tab; it is
	// disabled until a document is selected.
	dokumentaTbl.FuncClick = "robnoFakturePrikazIzbor(this)"
	dokumentaTbl.BtnPrint = common.SetPrintButton(p.ui.PrikazStampaID, "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint,
		"GET", true, common.ClassPrintButton+" opacity-50 pointer-events-none", p.ui.PrikazRdokID)
	return dokumentaTbl, h.robnoPrikazStavkeTabela(p, nil)
}

// robnoPrikazDokumenta returns the grid of the documents of the second tab of the request (GET rnalid and
// vrd; page, pageSize, query, sortBy and sortOrder of the pagination, the search and the sort of the grid):
// the documents of the vrste of the screen in the magacin of the nalog (the legacy query of the grid: by the
// vrsta and the magacin, not by the nalog), with the empty grid of the stavke and the search.
func (h *RobnoDokumentaUnosHandler) robnoPrikazDokumenta(c *gin.Context, p robnoPrikaz) (domain.TableData, domain.TableData, domain.InputControl, error) {
	ctx := c.Request.Context()
	dok, err := h.robnoFaktureNalog(ctx, int64(common.StringToInt(c.Query("rnalid"))), common.StringToInt(c.Query("vrd")))
	if err != nil {
		return domain.TableData{}, domain.TableData{}, domain.InputControl{}, err
	}
	page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
	dokumenta, total, err := h.service.GetRobDokumentaPrikaz(ctx, dok.MagaciniID, p.vrste(dok.Vrd), domain.RobDokumentaPrikazParams{
		Page: page, PageSize: pageSize, Search: c.Query("query"), SortBy: c.Query("sortBy"), SortOrder: c.Query("sortOrder")})
	if err != nil {
		return domain.TableData{}, domain.TableData{}, domain.InputControl{}, err
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(p, dokumenta)
	// The pagination, the search and the sort reload the grid with the nalog and the vrsta; the pagination
	// keeps the search text.
	vals := fmt.Sprintf(`{"rnalid": "%d", "vrd": "%d"}`, dok.RnalID, dok.Vrd)
	dokumentaTbl.URLGetAll = p.dokumentaURL
	dokumentaTbl.HxVals = vals
	dokumentaTbl.Pagination.HxVals = vals
	dokumentaTbl.Pagination.HxInclude = "#" + p.searchID
	dokumentaTbl.Pagination.CurrentPage = page
	dokumentaTbl.Pagination.PageSize = pageSize
	common.SetTableTotalRecords(&dokumentaTbl, total, pageSize)
	search := common.CreateSearchInput(p.searchID, h.translator, p.dokumentaURL, "#"+p.dokumentaTableID, vals)
	return dokumentaTbl, dokumentStavkeTbl, search, nil
}

// robnoPrikazStavkeTabela returns the grid of the stavke of the selected document of the second tab.
func (h *RobnoDokumentaUnosHandler) robnoPrikazStavkeTabela(p robnoPrikaz, stavke []domain.RobStavkaRow) domain.TableData {
	tbl := h.robnoFaktureTable(p.stavkeNaslov, p.stavkeTableID, p.stavkeFields)
	tbl.Rows = p.stavkeRows(stavke)
	return tbl
}

// robnoPrikazTab fills the second tab when it is shown (GET rnalid and vrd): the documents and the empty
// grid of the stavke. An error is shown in the dialog of the messages of the screen.
func (h *RobnoDokumentaUnosHandler) robnoPrikazTab(c *gin.Context, p robnoPrikaz) {
	dokumentaTbl, dokumentStavkeTbl, search, err := h.robnoPrikazDokumenta(c, p)
	if err != nil {
		p.greska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	_ = tmpl_robno.RobnoFakturePrikaz(p.ui, dokumentaTbl, dokumentStavkeTbl, search, h.translator).Render(c.Request.Context(), c.Writer)
}

// robnoPrikazGrid returns the grid of the documents of the second tab only (the pagination, the search and
// the sort of the grid).
func (h *RobnoDokumentaUnosHandler) robnoPrikazGrid(c *gin.Context, p robnoPrikaz) {
	dokumentaTbl, _, _, err := h.robnoPrikazDokumenta(c, p)
	if err != nil {
		p.greska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	utils.RenderContent(c, dokumentaTbl, "Table")
}

// robnoPrikazStavke returns the grid of the stavke of the document of the id (GET .../stavke/:id, the click
// of a row of the documents of the second tab).
func (h *RobnoDokumentaUnosHandler) robnoPrikazStavke(c *gin.Context, p robnoPrikaz) {
	ctx := c.Request.Context()
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		p.greska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	_ = tmpl_robno.RobnoFakturePrikazStavke(h.robnoPrikazStavkeTabela(p, stavke), h.translator).Render(ctx, c.Writer)
}

// FakturePrikaz fills the tab "Prikaz faktura" when it is shown (GET rnalid and vrd; see robnoPrikazTab).
func (h *RobnoDokumentaUnosHandler) FakturePrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoFakturePrikazOpis())
}

// FakturePrikazDokumenta returns the grid of the documents of the tab "Prikaz faktura" only (the
// pagination, the search and the sort of the grid).
func (h *RobnoDokumentaUnosHandler) FakturePrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoFakturePrikazOpis())
}

// FakturePrikazStavke returns the grid of the stavke of the document of the id (GET /prikaz/stavke/:id,
// the click of a row of the documents of the tab "Prikaz faktura").
func (h *RobnoDokumentaUnosHandler) FakturePrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoFakturePrikazOpis())
}

// robnoFaktureGreska shows an error in the dialog of the messages of the screen (robnoFaktureInfoID)
// instead of the target of the request.
func robnoFaktureGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoFaktureInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoFaktureInfoDialogID, message, true)
}

// FaktureDokument reloads the content of the tab "Fakture" of the "Fakture veleprodaje" screen with the
// existing document of the id (GET /dokument/:id with the rnalid of the screen), like the legacy exit of
// EDT_DOKUM with a selected broj. A document that cannot be changed (of another nalog, posted,
// fiscalized, sent to the SEF) gives a new document with the next broj and the message why
// (RobnoFaktureInfo).
func (h *RobnoDokumentaUnosHandler) FaktureDokument(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Param("id")))
	h.faktureDokument(c, domain.RobnoFaktureDokumentParams{RdokID: rdokID})
}

// FaktureNovi reloads the content of the tab "Fakture" with a new document of the vrsta (GET /novi with
// rnalid, vrd and brdok): the typed broj (its document when it exists, like FaktureDokument), else the
// next broj of the vrsta in the magacin of the nalog ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) FaktureNovi(c *gin.Context) {
	h.faktureDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: strings.TrimSpace(c.Query("brdok"))})
}

// faktureDokument renders the content of the tab "Fakture" with the document of the params: the nalog
// (its vrsta, broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) faktureDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, common.ErrMsgSessionNotFound)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		c.Header("HX-Reswap", "none")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{{Field: "brdok", ErrorMessage: common.ErrMsgReadData}}, err.Error())
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoFaktureEkran(c, dok, params)
	if err != nil {
		c.Header("HX-Reswap", "none")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{{Field: "brdok", ErrorMessage: common.ErrMsgReadData}}, err.Error())
		return
	}
	h.robnoFaktureBodyOdgovor(c, ekran)
}

// robnoFaktureBodyOdgovor renders the content of the tab "Fakture" (the answer of a selected or new
// document and of the deletion of a document), with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoFaktureBodyOdgovor(c *gin.Context, ekran robnoFaktureEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobnoFaktureBody(robnoFaktureUI(), ekran.tbl, ekran.avansiTbl, ekran.header, ekran.stavka, ekran.vrstaDokumentaValues,
		nil, robnosvc.RobSistemPdvCombo(), ekran.ziroRacuni, ekran.btns, ekran.search, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoFaktureInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-fakture-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-fakture-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoFaktureInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoFaktureNalog returns the document of the screen of the nalog of the id: its vrsta, broj and magacin
// are read from the nalog (the screen sends only the rnalid), the vrsta dokumenta is vrd.
func (h *RobnoDokumentaUnosHandler) robnoFaktureNalog(ctx context.Context, rnalID int64, vrd int) (robnoDokumentaUnosDokument, error) {
	rnal, found, err := h.service.GetRobnoUnosNalog(ctx, rnalID)
	if err != nil {
		return robnoDokumentaUnosDokument{}, err
	}
	if !found {
		return robnoDokumentaUnosDokument{}, fmt.Errorf("nalog ne postoji")
	}
	dok := robnoDokumentaUnosDokument{RnalID: rnalID, Vrd: vrd, Tipdok: strings.TrimSpace(rnal.Tipdok), Nalog: rnal.Nalog}
	if rnal.MagaciniID.Valid {
		dok.MagaciniID = int(rnal.MagaciniID.Int64)
	}
	return dok, nil
}

// SaveRobFakture saves the header of the faktura ("Sačuvaj" of the header, POST of the fields of the
// header), like the legacy "Sačuvaj" of rob_Fakture: the validation (ValidateRobFakture), the save of the
// robni dokument, a new one or the existing one of the broj (SaveRobFaktura: SnimiRDOK and ObracunFAKT),
// and the content of the tab with the saved document: the header saved and locked, the stavke open
// (the legacy opens the window of the stavke). A validation error is the JSON of the field errors
// (status 422, the content is not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobFakture(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoFaktureParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	// The nalog of the rnalid gives the vrsta, the broj and the magacin of the document.
	dok, err := h.robnoFaktureNalog(ctx, params.RnalID, common.StringToInt(params.Vrd))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.MagaciniID = dok.MagaciniID
	if fieldErrors := h.service.ValidateRobFakture(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	// The broj of the saved document is opened by its id.
	params.RdokID, err = h.service.SaveRobFaktura(ctx, params)
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, common.ErrMsgSaveData+": "+err.Error(), true)
		return
	}
	ekran, err := h.robnoFaktureEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	// The saved header is locked, the stavke are entered (the legacy rob_FAKTURE_DET).
	ekran.header.Snimljen = true
	if err := tmpl_robno.RobnoFaktureBody(robnoFaktureUI(), ekran.tbl, ekran.avansiTbl, ekran.header, ekran.stavka, ekran.vrstaDokumentaValues,
		nil, robnosvc.RobSistemPdvCombo(), ekran.ziroRacuni, ekran.btns, ekran.search, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
}

// FaktureArtikal returns the artikal of a stavka (GET rdokid, sifra - the šifra or the barkod - and konto) as
// JSON for the form of the stavke (the legacy exit of EDT_SIFRA): see GetRobStavkaArtikal; found false when
// the artikal does not exist.
func (h *RobnoDokumentaUnosHandler) FaktureArtikal(c *gin.Context) {
	ctx := c.Request.Context()
	artikal, found, err := h.service.GetRobStavkaArtikal(ctx, int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"), c.Query("konto"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// FaktureStavka returns a saved stavka (GET rdokid, rproid) as JSON for the form of the stavke (the legacy
// "Izmene"): see GetRobStavkaForma.
func (h *RobnoDokumentaUnosHandler) FaktureStavka(c *gin.Context) {
	ctx := c.Request.Context()
	stavka, found, err := h.service.GetRobStavkaForma(ctx, int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": "Stavka koju birate za izmenu ne postoji."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SaveRobFaktureStavka saves a stavka of the faktura (POST of the fields of the stavke, the legacy
// "Snimi" of rob_FAKTURE_DET): the validation (ValidateRobStavka), the save (SaveRobStavka: a new stavka
// or the changed one) and the content of the tab with the document (the grid, the iznos of the document,
// the strip of the nalog; the focus to the šifra for the next stavka). A validation error is the JSON of
// the field errors (status 422, the content is not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobFaktureStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobStavka(ctx, params); err != nil {
		// The error dialog goes into its container of the stavke instead of the content of the tab.
		c.Header("HX-Retarget", "#"+robnoFaktureStavkaInfoID)
		c.Header("HX-Reswap", "innerHTML")
		utils.RenderDialogOK(c, robnoFaktureStavkaInfoDialogID, common.ErrMsgSaveData+": "+err.Error(), true)
		return
	}
	h.robnoFaktureStavkeOdgovor(c, params.RdokID)
}

// DeleteRobFaktureStavka deletes a stavka of the faktura (DELETE …/stavka/brisi/:id, the id of the stavka:
// the legacy "Brisanje"; its document is read from the stavka) and answers with the content of the tab
// with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobFaktureStavka(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID, err := h.service.DeleteRobStavka(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		c.Header("HX-Reswap", "none")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoFaktureStavkeOdgovor(c, rdokID)
}

// PotvrdaBrisanjaStavke asks to confirm the deletion of a stavka ("Obriši" of its row in the grid, GET
// rdokid and id): "Da" deletes it (DeleteRobFaktureStavka) and the content of the tab comes back.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rproID := robnoFaktureRproID(c)
	dialog := domain.Dialog{Id: robnoFaktureStavkaPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-fakture-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-fakture-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoFaktureStavkaDelURL, rproID),
		HxTarget:             "#" + robnoFaktureBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-fakture-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message("Da li želite da obrišete stavku?")}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// PotvrdaBrisanjaFakture asks to confirm the deletion of the document of the screen (GET rdokid: "Briši" of
// the faktura, the legacy Dialog("Brisanje dokumenta?")): "Da" deletes it (DeleteRobFakture), "Ne" closes the
// dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaFakture(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dialog := domain.Dialog{Id: robnoFaktureDokPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-fakture-dok-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-fakture-dok-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoFaktureDeleteURL, rdokID),
		HxTarget:             "#" + robnoFaktureBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-fakture-dok-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{robnoInterniMsgBrisanjeDokum}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobFakture deletes the faktura of the id with its stavke, avansi and rate (DELETE /brisi/:id, the
// legacy "Briši" of rob_Fakture, DeleteRobFaktura) and answers with the content of the tab with a new
// document of the vrsta (the legacy Reset and NoviBrojRDOK) and the message of the deletion. A document that
// cannot be deleted (sent to the SEF or to the Pojedinačna evidencija PDV, fiscalized, posted) keeps the
// screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobFakture(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoFaktureGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobFaktura(ctx, rdokID)
	if err != nil {
		robnoFaktureGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoFaktureGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoFaktureGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoFaktureEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoFaktureGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.poruka = robnoInterniMsgObrisano
	h.robnoFaktureBodyOdgovor(c, ekran)
}

// robnoFaktureRproID returns the stavka of a request: the id of its row in the grid (id) or rproid.
func robnoFaktureRproID(c *gin.Context) int64 {
	if id := common.StringToInt(c.Query("id")); id > 0 {
		return int64(id)
	}
	return int64(common.StringToInt(c.Query("rproid")))
}

// robnoFaktureStavkeOdgovor renders the content of the tab "Fakture" with the document of the stavke (its
// header saved, the stavke open): the answer of the save and of the deletion of a stavka.
func (h *RobnoDokumentaUnosHandler) robnoFaktureStavkeOdgovor(c *gin.Context, rdokID int64) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoFaktureEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = true
	_ = tmpl_robno.RobnoFaktureBody(robnoFaktureUI(), ekran.tbl, ekran.avansiTbl, ekran.header, ekran.stavka, ekran.vrstaDokumentaValues,
		nil, robnosvc.RobSistemPdvCombo(), ekran.ziroRacuni, ekran.btns, ekran.search, h.translator).Render(ctx, c.Writer)
}

// robnoFaktureEkranData is the content of the tab "Fakture" of the screen with its document.
type robnoFaktureEkranData struct {
	header               domain.RobnoFaktureHeaderView
	stavka               domain.RobnoFaktureStavka
	tbl, avansiTbl       domain.TableData
	vrstaDokumentaValues []domain.ComboItem
	ziroRacuni           []domain.ComboItem
	btns                 domain.RobnoFaktureButtons
	search               domain.InputControl
	poruka               string
}

// robnoFaktureEkran builds the content of the tab "Fakture" with the document of the broj (a new one
// for an empty broj, see GetRobFaktureDokument): the strip of the nalog, the header, the avansi and the
// buttons of the screen.
func (h *RobnoDokumentaUnosHandler) robnoFaktureEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoFaktureEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoFaktureEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	if ekran.vrstaDokumentaValues, err = h.service.GetVrstaDokumentaComboValues(ctx, commonsvc.WithIn("modul", h.cfg.GrupeDokumenata.GetDokRob()...)); err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobFaktureDokument(ctx, params)
	if err != nil {
		return ekran, err
	}
	// The vrsta of a document opened by its id is the one of the document.
	if vrd := common.StringToInt(dokument.Header.Vrd); vrd > 0 {
		dok.Vrd = vrd
	}
	ekran.header, ekran.stavka, ekran.ziroRacuni, ekran.poruka = dokument.Header, dokument.Stavka, dokument.ZiroRacuni, dokument.Poruka

	// The strip of the nalog.
	robnoNalogStrip(&ekran.header, rnal, magacin)

	// The grids: the stavke and the avansi of the document.
	ekran.tbl = h.robnoFaktureTable("Stavke fakture", robnoFaktureStavkeTableID, h.service.GetFaktureStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, ekran.header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobStavkeTableRows(stavke)
	// The buttons of every row: "Izmeni" puts the stavka in the form (robnoFaktureStavkaUcitaj), "Obriši"
	// asks to confirm its deletion; both send the document with the id of the row.
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, ekran.header.RdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoFaktureStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robnoFaktureStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoFaktureStavkaPotvrdaURL,
		HxTarget: "#robno-fakture-potvrda", HxSwap: "innerHTML"}
	ekran.avansiTbl = h.robnoFaktureTable("Avansi", robnoFaktureAvansiTableID, h.service.GetFaktureAvansiTableFields())
	ekran.avansiTbl.Rows = dokument.Avansi

	// The buttons: "Nazad" releases the lock of the nalog and closes the dialog, "Novi dok." reloads
	// the content with a new document.
	ekran.btns = robnoFaktureButtonsFor()
	ekran.btns.Back.HxActionURL = robnoFaktureUnlockURL(dok.RnalID)
	ekran.btns.Back.HxRequestType = "POST"
	ekran.btns.Back.HxTarget = ""
	ekran.btns.Back.HxSwap = "none"
	ekran.btns.Back.HxOnAfterRequest = "closeDialog"
	ekran.btns.Back.HxOnAfterRequestArgs = []any{robnoFaktureDialogID}
	ekran.btns.New.HxActionURL = fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoFaktureNoviURL, dok.RnalID, dok.Vrd)
	ekran.btns.New.HxTarget = "#" + robnoFaktureBodyID
	ekran.btns.New.HxSwap = "outerHTML"
	ekran.btns.New.HxOnAfterRequest = ""
	ekran.btns.New.HxOnAfterRequestArgs = nil
	// "Briši" of the saved document (a new one has nothing to delete).
	if ekran.header.RdokID > 0 {
		ekran.btns.Delete.HxActionURL = fmt.Sprintf("%s?rdokid=%d", robnoFaktureDokPotvrdaURL, ekran.header.RdokID)
	} else {
		ekran.btns.Delete.IsDisabled = true
		ekran.btns.Delete.BtnClass += " opacity-50 cursor-not-allowed"
		ekran.btns.Delete.HxActionURL = ""
	}

	ekran.search = common.CreateSearchInput(robnoDokumentaSearchInputID, h.translator, robnoFaktureURL, "#"+robnoFaktureStavkeTableID, "")
	return ekran, nil
}

// robnoNalogStrip fills the strip of the nalog of the header of a document (the screens of the entry of the
// robni dokumenti: the faktura, the interni prenos): the vrsta, the broj and the datum of the nalog, its
// totals and the magacin.
func robnoNalogStrip(header *domain.RobnoFaktureHeaderView, rnal domain.Rnal, magacin domain.Magacini) {
	header.VrNal = rnal.Tipdok
	if rnal.Nalog > 0 {
		header.Nalog = strconv.Itoa(rnal.Nalog)
	}
	if rnal.Danal.Valid {
		header.DatumNaloga = rnal.Danal.Time.Format(common.DateLayout)
	}
	header.UkDokum = strconv.Itoa(rnal.Brdo)
	header.UkStavki = strconv.Itoa(rnal.Brst)
	header.Duguje = fmt.Sprintf("%.2f", rnal.Dug)
	header.Potrazuje = fmt.Sprintf("%.2f", rnal.Pot)
	if magacin.MagaciniID > 0 {
		header.Magacin = strings.TrimSpace(fmt.Sprintf("%d %s", magacin.Mag, magacin.Opis))
	}
}

// robnoFaktureUnlockURL releases the lock of the nalog (the "Zatvori" and "Nazad" of the screen).
func robnoFaktureUnlockURL(rnalID int64) string {
	return fmt.Sprintf("%s/%d", robnoDokumentaURLUnlockNalog, rnalID)
}

// robnoFaktureTable builds an (empty) grid of the "Fakture veleprodaje" screen.
func (h *RobnoDokumentaUnosHandler) robnoFaktureTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", robnoFaktureURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, robnoFaktureURL, false, false, false)
	return tbl
}

// AddRoutes registers the routes of the entry of the robni dokumenti.
func (h *RobnoDokumentaUnosHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/unos-dokumenta/otvori", h.OtvoriDokument)
	r.GET("/api/robno-dokumenta/fakture/dokument/:id", h.FaktureDokument)
	r.GET("/api/robno-dokumenta/fakture/novi", h.FaktureNovi)
	r.POST("/api/robno-dokumenta/fakture/save", h.SaveRobFakture)
	r.GET("/api/robno-dokumenta/fakture/artikal", h.FaktureArtikal)
	r.GET("/api/robno-dokumenta/fakture/stavka", h.FaktureStavka)
	r.POST("/api/robno-dokumenta/fakture/stavka/save", h.SaveRobFaktureStavka)
	r.DELETE("/api/robno-dokumenta/fakture/stavka/brisi/:id", h.DeleteRobFaktureStavka)
	r.GET("/api/robno-dokumenta/fakture/stavka/potvrda", h.PotvrdaBrisanjaStavke)
	r.GET("/api/robno-dokumenta/fakture/brisi/potvrda", h.PotvrdaBrisanjaFakture)
	r.DELETE("/api/robno-dokumenta/fakture/brisi/:id", h.DeleteRobFakture)
	// The interni prenosi - dostavnice.
	h.robnoInterniRoutes(r)
	// The fakture usluga.
	h.robnoUslugeRoutes(r)
	// The opšti dokumenti.
	h.robnoOpstiRoutes(r)
	r.GET("/api/robno-dokumenta/fakture/prikaz", h.FakturePrikaz)
	r.GET("/api/robno-dokumenta/fakture/prikaz/dokumenta", h.FakturePrikazDokumenta)
	r.GET("/api/robno-dokumenta/fakture/prikaz/stavke/:id", h.FakturePrikazStavke)
}
