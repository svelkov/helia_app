package robno

import (
	"fmt"
	"strings"

	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RobnoDokumentaUnosHandler handles the entry of the robni dokumenti of a robni nalog: the screens the
// legacy "Snimi nalog" of the "Unos dokumenta" tab opens after the nalog is saved, one per group of the
// vrsta dokumenta (dokvrsta.grpdok, see robnoDokumentaUnosEkrani and OtvoriDokument). The screens are
// added one by one; the templates of the "Fakture veleprodaje" screen are in
// frontend/templates/robno/robnadokumentaunos.templ, its handler is not written yet.
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

	// robnoFaktureDialogID is the id of the dialog of the screen (RobnoFaktureDialog, the one closeDialog hides).
	robnoFaktureDialogID = "robno-fakture-dialog"
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
}

// robnoDokumentaUnosEkrani are the screens of the entry by the group of the vrsta dokumenta
// (dokvrsta.grpdok), like the SWITCH of the legacy "Snimi nalog". A group gets its open function when its
// screen is written: FAK, FRP and KPK the faktura veleprodaje (the legacy rob_Fakture; with
// fvr.etunosprod = 2 the legacy pr_FaktureOtp), IRS only from a magacin of the proizvodnja (tipmag P or
// L).
var robnoDokumentaUnosEkrani = map[string]robnoDokumentaUnosEkran{
	"FAK": {legacy: "rob_Fakture"},
	"FRP": {legacy: "rob_Fakture"},
	"KPK": {legacy: "rob_Fakture"},
	"FZR": {legacy: "rob_Fakture"},
	"DIR": {legacy: "rob_Fakture_maloprodaje"},
	"IRT": {legacy: "rob_Interni_racuni"},
	"IRP": {legacy: "rob_Interni_racuni"},
	"IRS": {legacy: "rob_Interni_racuni"},
	"PRE": {legacy: "rob_Fakture_Usluga"},
	"FUR": {legacy: "rob_Fakture_Usluga"},
	"FCO": {legacy: "rob_Fakture_Usluga"},
	"RCO": {legacy: "rob_Fakture_Usluga"},
	"KPF": {legacy: "rob_Fakture_Usluga"},
	"OPD": {legacy: "rob_opsti_dokumenti_opd"},
	"POP": {legacy: "rob_opsti_dokumenti_opd"},
	"POT": {legacy: "rob_opsti_dokumenti_opd"},
	"FIN": {legacy: "rob_opsti_dokumenti_FIN"},
	"KOL": {legacy: "rob_opsti_dokumenti_KOL"},
	"MNV": {legacy: "rob_opsti_dokumenti_MNV"},
	"NIV": {legacy: "rob_Nivelacije_Veleprodaja"},
	"ARA": {legacy: "rob_AvansniRacuni"},
	"ARU": {legacy: "rob_AvansniRacuni"},
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

// AddRoutes registers the routes of the entry of the robni dokumenti.
func (h *RobnoDokumentaUnosHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/unos-dokumenta/otvori", h.OtvoriDokument)
}
