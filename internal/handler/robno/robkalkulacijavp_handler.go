package robno

import (
	"fmt"
	"net/http"
	"strings"

	tmpl "helia/frontend/templates"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/internal/common"
	"helia/internal/domain"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// The "Kalkulacije veleprodaje" screen (the legacy rob_Kalkulacije_Veleprodaja with its stavke
// rob_Kalkulacije_Veleprodaja_Stavke and its zavisni troškovi rob_Zavisni_Troskovi, the vrste dokumenata of the group
// PLT): its urls and the ids of its elements (the templates get them by robnoKalUI and domain.RobnoKalUI).
const (
	robnoKalURL = robnoDokumentaURLPrefix + "/kalkulacije-vp"
	// robnoKalDokumentURL + "/:id" reloads the content of the screen with an existing document (the selected broj),
	// robnoKalNoviURL with a new one ("Novi dok.", a typed broj).
	robnoKalDokumentURL = robnoKalURL + "/dokument"
	robnoKalNoviURL     = robnoKalURL + "/novi"
	// robnoKalSaveURL saves the header (POST), robnoKalPotvrdaURL asks to confirm the deletion of the document and
	// robnoKalBrisiURL + "/:id" deletes it; robnoKalStopaURL returns the stopa of a poreska oznaka on a datum (GET po,
	// datum: the grids of the porezi).
	robnoKalSaveURL    = robnoKalURL + "/save"
	robnoKalBrisiURL   = robnoKalURL + "/brisi"
	robnoKalPotvrdaURL = robnoKalBrisiURL + "/potvrda"
	robnoKalStopaURL   = robnoKalURL + "/stopa"
	// robnoKalPrikazURL fills the tab "Pregled kalkulacija" when it is shown, robnoKalPrikazDokumentaURL reloads its
	// grid, robnoKalPrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoKalPrikazURL          = robnoKalURL + "/prikaz"
	robnoKalPrikazDokumentaURL = robnoKalPrikazURL + "/dokumenta"
	robnoKalPrikazStavkeURL    = robnoKalPrikazURL + "/stavke"
	// The stavke: robnoKalArtikalURL returns the artikal of the stavka (GET rdokid, sifra, konto), robnoKalStavkaURL a
	// saved stavka for the form (GET rdokid, id), robnoKalStavkaSaveURL saves the stavka (POST),
	// robnoKalStavkaPotvrdaURL asks to confirm the deletion of a stavka and robnoKalStavkaBrisiURL + "/:id" deletes it.
	robnoKalArtikalURL       = robnoKalURL + "/artikal"
	robnoKalStavkaURL        = robnoKalURL + "/stavka"
	robnoKalStavkaSaveURL    = robnoKalStavkaURL + "/save"
	robnoKalStavkaPotvrdaURL = robnoKalStavkaURL + "/potvrda"
	robnoKalStavkaBrisiURL   = robnoKalStavkaURL + "/brisi"
	// The zavisni troškovi: robnoKalTrosakURL returns a saved trošak for the form (GET rdokid, id),
	// robnoKalTrosakSaveURL saves the trošak (POST), robnoKalTrosakPotvrdaURL asks to confirm the deletion of a trošak
	// and robnoKalTrosakBrisiURL + "/:id" deletes it.
	robnoKalTrosakURL        = robnoKalURL + "/trosak"
	robnoKalTrosakSaveURL    = robnoKalTrosakURL + "/save"
	robnoKalTrosakPotvrdaURL = robnoKalTrosakURL + "/potvrda"
	robnoKalTrosakBrisiURL   = robnoKalTrosakURL + "/brisi"

	robnoKalDialogID               = "robno-kal-dialog"
	robnoKalBodyID                 = "robno-kal-body"
	robnoKalInfoID                 = "robno-kal-info"
	robnoKalInfoDialogID           = "robno-kal-info-dialog"
	robnoKalPotvrdaDialogID        = "robno-kal-brisanje"
	robnoKalPitanjeDialogID        = "robno-kal-pitanje"
	robnoKalStavkaInfoID           = "robno-kal-stavka-info"
	robnoKalStavkaInfoDialogID     = "robno-kal-stavka-info-dialog"
	robnoKalStavkaPotvrdaDialogID  = "robno-kal-stavka-brisanje"
	robnoKalTrosakPotvrdaDialogID  = "robno-kal-trosak-brisanje"
	robnoKalHeaderPanelID          = "robno-kal-header-panel"
	robnoKalTroskoviPanelID        = "robno-kal-troskovi-panel"
	robnoKalStavkePanelID          = "robno-kal-stavke-panel"
	robnoKalPrikazID               = "robno-kal-prikaz"
	robnoKalPrikazStavkeID         = "robno-kal-prikaz-stavke"
	robnoKalPrikazRdokID           = "robno-kal-prikaz-rdokid"
	robnoKalPrikazSearchID         = "robno-kal-prikaz-search"
	robnoKalPrikazStampaID         = "robno-kal-prikaz-stampa"
	robnoKalStavkeTableID          = "robno-kal-stavke-table"
	robnoKalTroskoviTableID        = "robno-kal-troskovi-table"
	robnoKalDokumentaTableID       = "robno-kal-dokumenta-table"
	robnoKalDokumentStavkeTableID  = "robno-kal-dokument-stavke-table"
	robnoKalTroskoviPitanjeDialog  = "robno-kal-troskovi-pitanje"
	robnoKalMsgTroskoviPostoje     = "Da li postoje eksterni zavisni troškovi za ovaj dokument?"
	robnoKalMsgBrisanjeStavkePotvr = "Brisanje stavke? Da li ste sigurni?"
)

// The open part of the stavke and of the zavisni troškovi after the save of the header (both are enabled).
const (
	robnoKalOtvoreneStavke   = "stavke"
	robnoKalOtvoreniTroskovi = "troskovi"
)

// robnoKalUI returns the ids and the urls of the "Kalkulacije veleprodaje" screen.
func robnoKalUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoKalHeaderPanelID,
		StavkePanelID:     robnoKalStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoKalDialogID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoKalDokumentURL,
		ArtikalURL:        robnoKalArtikalURL,
		NoviURL:           robnoKalNoviURL,
		BodyID:            robnoKalBodyID,
		InfoID:            robnoKalInfoID,
		StavkaInfoID:      robnoKalStavkaInfoID,
		PrikazID:          robnoKalPrikazID,
		PrikazStavkeID:    robnoKalPrikazStavkeID,
		PrikazRdokID:      robnoKalPrikazRdokID,
		PrikazStampaID:    robnoKalPrikazStampaID,
		PrikazURL:         robnoKalPrikazURL,
	}
}

// robnoKalEkranUI are the ids and urls of the screen the templates need besides domain.RobnoFaktureUI: the panel of the
// zavisni troškovi, the urls of the stopa, of a saved trošak and of a saved stavka.
func robnoKalEkranUI() domain.RobnoKalUI {
	return domain.RobnoKalUI{
		TroskoviPanelID: robnoKalTroskoviPanelID,
		StopaURL:        robnoKalStopaURL,
		TrosakURL:       robnoKalTrosakURL,
		StavkaURL:       robnoKalStavkaURL,
	}
}

// robnoKalPrikazOpis is the tab "Pregled kalkulacija": the documents of the vrsta in the magacin of the nalog (the
// legacy QRY_RDOK) and the stavke of the selected one (QRY_RPRO_ZADOK).
func (h *RobnoDokumentaUnosHandler) robnoKalPrikazOpis() robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoKalUI(),
		dokumentaTableID: robnoKalDokumentaTableID,
		stavkeTableID:    robnoKalDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetKalDokumentaTableFields(),
		stavkeFields:     h.service.GetKalPrikazStavkeTableFields(),
		dokumentaRows:    robnosvc.RobKalDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobKalPrikazStavkeTableRows,
		dokumentaURL:     robnoKalPrikazDokumentaURL,
		stavkeURL:        robnoKalPrikazStavkeURL,
		searchID:         robnoKalPrikazSearchID,
		vrste:            func(vrd int) []int { return []int{vrd} },
		greska:           robnoKalGreska,
	}
}

// KalPrikaz fills the tab "Pregled kalkulacija" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) KalPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoKalPrikazOpis())
}

// KalPrikazDokumenta returns the grid of the documents of the tab "Pregled kalkulacija" only.
func (h *RobnoDokumentaUnosHandler) KalPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoKalPrikazOpis())
}

// KalPrikazStavke returns the stavke of the document of the id (the click of a row of the tab "Pregled kalkulacija").
func (h *RobnoDokumentaUnosHandler) KalPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoKalPrikazOpis())
}

// robnoKalButtons builds the buttons of the "Kalkulacije veleprodaje" screen for the document of the nalog (rdokID 0:
// a new one), like the ones of the porudžbenica: "Sačuvaj" saves the header and opens the stavke, "Izmeni" switches
// back to the header, "Novi dok." a new document, "Briši" deletes the saved one (after its confirmation), "Nazad"
// releases the lock of the nalog and closes the dialog, "Štampaj" prints the document (the legacy
// RPT_ROB_KALKULACIJA). "Sačuvaj" of the stavka saves it, "Odustani" gives up the change of a stavka.
func robnoKalButtons(rnalID, rdokID int64, vrd int) domain.RobnoFaktureButtons {
	btns := domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-kal-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoKalSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoKalHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoKalBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoKalHeaderPanelID, robnoKalStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-kal-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClickArg: []any{robnoKalHeaderPanelID, robnoKalTroskoviPanelID, robnoKalStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-kal-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoKalNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoKalBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-kal-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoKalPotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoKalInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-kal-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoKalDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta).
		Print: common.SetPrintButton("robno-kal-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		SaveStavka: domain.Button{
			Id:                   "robno-kal-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoKalStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoKalStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoKalBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-kal-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robKalStavkaOdustani",
		},
	}
	// A new document has nothing to delete.
	if rdokID <= 0 {
		btns.Delete = domain.Button{Id: "robno-kal-brisi", LabelText: "Briši", Icon: "delete", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassDeleteButton + " opacity-50 cursor-not-allowed"}
	}
	return btns
}

// robnoKalTrosakButtons are the buttons of the form of a zavisni trošak: "Sačuvaj" saves it (a new one or the one
// loaded by "Izmeni" of its row), "Odustani" gives up the change (a next "Sačuvaj" adds a trošak).
func robnoKalTrosakButtons() (save, cancel domain.Button) {
	save = domain.Button{
		Id:                   "robno-kal-trosak-sacuvaj",
		LabelText:            "Sačuvaj",
		Icon:                 "save",
		BtnClass:             common.ClassSaveButton,
		HxActionURL:          robnoKalTrosakSaveURL,
		HxRequestType:        "POST",
		HxInclude:            "#" + robnoKalTroskoviPanelID + " [data-panel-fields]",
		HxTarget:             "#" + robnoKalBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
		HxOnAfterRequestArgs: []any{},
	}
	cancel = domain.Button{
		Id:         "robno-kal-trosak-odustani",
		LabelText:  "Odustani",
		IsVisible:  true,
		IsDisabled: true,
		BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
		HxOnClick:  "robKalTrosakOdustani",
	}
	return save, cancel
}

// robnoKalEkranData is the content of the tab "Kalkulacije veleprodaje" (the header, the entry of a zavisni trošak and
// of a stavka, their grids and the buttons), the part open after the save of the header and the message of the screen.
type robnoKalEkranData struct {
	dokument   domain.RobnoKalDokument
	tbl        domain.TableData
	troskovi   domain.TableData
	btns       domain.RobnoFaktureButtons
	otvoreni   string
	pitajTrosk bool
}

// robnoKalEkran builds the content of the tab "Kalkulacije veleprodaje" with the document of the params (a new one for
// an empty broj, see GetRobKalDokument): the strip of the nalog, the header, the zavisni troškovi with their grid, the
// stavka with the grid of the stavke and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoKalEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoKalEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoKalEkranData{otvoreni: robnoKalOtvoreneStavke}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	if ekran.dokument, err = h.service.GetRobKalDokument(ctx, params); err != nil {
		return ekran, err
	}
	header := &ekran.dokument.Header
	robnoNalogStrip(&header.RobnoFaktureHeaderView, rnal, magacin)
	vrd := dok.Vrd
	if v := common.StringToInt(header.Vrd); v > 0 {
		vrd = v
	}
	rdokID := header.RdokID

	// The stavke: "Izmeni" of a row puts the stavka in the form (robKalStavkaUcitaj), "Obriši" asks to confirm its
	// deletion.
	ekran.tbl = h.robnoFaktureTable("Stavke kalkulacije", robnoKalStavkeTableID, h.service.GetKalStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, rdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobKalStavkeTableRows(stavke)
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, rdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoKalStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robKalStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoKalStavkaPotvrdaURL,
		HxTarget: "#" + robnoKalStavkaInfoID, HxSwap: "innerHTML"}

	// The zavisni troškovi: "Izmeni" of a row puts the trošak in its form (robKalTrosakUcitaj), "Obriši" asks to confirm
	// its deletion.
	ekran.troskovi = h.robnoFaktureTable("Zavisni troškovi", robnoKalTroskoviTableID, h.service.GetKalTroskoviTableFields())
	troskovi, err := h.service.GetRobKalTroskovi(ctx, rdokID)
	if err != nil {
		return ekran, err
	}
	ekran.troskovi.Rows = robnosvc.RobKalTroskoviTableRows(troskovi)
	ekran.troskovi.ShowActions = len(troskovi) > 0
	ekran.troskovi.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, rdokID)
	ekran.troskovi.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoKalTrosakURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robKalTrosakUcitaj(event)"}
	ekran.troskovi.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoKalTrosakPotvrdaURL,
		HxTarget: "#" + robnoKalStavkaInfoID, HxSwap: "innerHTML"}

	ekran.btns = robnoKalButtons(dok.RnalID, rdokID, vrd)
	return ekran, nil
}

// robnoKalBody renders the content of the tab "Kalkulacije veleprodaje" (the answer of the save, of a selected or new
// document and of the deletions) with the message of the screen in its dialog and, after the first save of a header
// without zavisni troškovi, the question of the legacy whether the document has them ("Da" opens their part).
func (h *RobnoDokumentaUnosHandler) robnoKalBody(c *gin.Context, ekran robnoKalEkranData) {
	ctx := c.Request.Context()
	trosakSave, trosakCancel := robnoKalTrosakButtons()
	if err := tmpl_robno.RobKalBody(robnoKalUI(), robnoKalEkranUI(), ekran.dokument, ekran.tbl, ekran.troskovi, robnosvc.RobSistemPdvCombo(), ekran.btns, trosakSave,
		trosakCancel, ekran.otvoreni, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if poruka := ekran.dokument.Poruka; poruka != "" {
		dialog := domain.Dialog{Id: robnoKalInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-kal-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-kal-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoKalInfoID, poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
		return
	}
	if ekran.pitajTrosk {
		dialog := domain.Dialog{Id: robnoKalTroskoviPitanjeDialog, Title: "Zavisni troškovi"}
		btnClose := domain.Button{Id: "robno-kal-troskovi-pitanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnDa := domain.Button{Id: "robno-kal-troskovi-pitanje-da", IsVisible: true, LabelText: "Da", IdDialog: dialog.Id,
			BtnClass: common.ClassSaveButton, HxOnClick: "robKalOtvoriTroskove",
			HxOnClickArg: []any{dialog.Id, robnoKalTroskoviPanelID, robnoKalStavkePanelID}}
		btnNe := domain.Button{Id: "robno-kal-troskovi-pitanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id,
			BtnClass: common.ClassOdustaniButton, HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl.DialogConfirm([]string{h.translator.Message(robnoKalMsgTroskoviPostoje)}, dialog, btnClose, btnDa, btnNe, h.translator,
			common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
	}
}

// robnoKalGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func robnoKalGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoKalInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoKalInfoDialogID, message, true)
}

// robnoKalStavkaGreska shows an error of the stavke or of the zavisni troškovi in the container of their dialogs.
func robnoKalStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoKalStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoKalStavkaInfoDialogID, message, true)
}

// robnoKalPitanje answers a save with the question of its validation (the legacy Dialog / YesNo before the save): "Da"
// sends the fields of the part (include) again with potvrda (url, the content of the screen bodyID is the answer), "Ne"
// closes the dialog. The dialog is shown in the container infoID instead of the target of the request.
func (h *RobnoDokumentaUnosHandler) robnoKalPitanje(c *gin.Context, infoID, bodyID, pitanje, url, include string) {
	c.Header("HX-Retarget", "#"+infoID)
	c.Header("HX-Reswap", "innerHTML")
	dialog := domain.Dialog{Id: robnoKalPitanjeDialogID, Title: "Pitanje"}
	btnClose := domain.Button{Id: "robno-kal-pitanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-kal-pitanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassSaveButton,
		HxRequestType:        "POST",
		HxActionURL:          url,
		HxInclude:            include,
		HxVals:               `{"potvrda": "true"}`,
		HxTarget:             "#" + bodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "robKalPosleOdgovora",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-kal-pitanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton,
		HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
	// The question has the values of the document in it (e.g. the broj of the izvorni dokument): its text, not a key.
	_ = tmpl.DialogConfirm(strings.Split(pitanje, "\n"), dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// otvoriKalkulaciju opens the "Kalkulacije veleprodaje" screen (the legacy rob_Kalkulacije_Veleprodaja) of the nalog as
// a dialog with its two tabs, with a new document (the next broj of the vrsta in the magacin, NoviBrojRDOK). The nalog
// stays locked while the dialog is open: "Zatvori" and "Nazad" release it.
func (h *RobnoDokumentaUnosHandler) otvoriKalkulaciju(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoKalEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoKalPrikazOpis(), nil)
	btnClose := domain.Button{
		Id:            "robno-kal-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoKalDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	trosakSave, trosakCancel := robnoKalTrosakButtons()
	if err := tmpl_robno.Rob_KalkulacijaVP(robnoKalUI(), robnoKalEkranUI(), ekran.dokument, ekran.tbl, ekran.troskovi, dokumentaTbl,
		dokumentStavkeTbl, robnosvc.RobSistemPdvCombo(), ekran.btns, trosakSave, trosakCancel, btnClose, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// kalDokument renders the content of the tab "Kalkulacije veleprodaje" with the document of the params: the nalog (its
// vrsta, broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) kalDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoKalEkran(c, dok, params)
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoKalBody(c, ekran)
}

// KalDokument reloads the content of the tab with the existing document of the id (GET /dokument/:id with the rnalid
// of the screen: the selected broj). A document that cannot be changed (of another nalog, posted, fiscalized) gives a
// new document and the message why (the legacy exit of EDT_DOKUM).
func (h *RobnoDokumentaUnosHandler) KalDokument(c *gin.Context) {
	h.kalDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// KalNovi reloads the content of the tab with a new document (GET rnalid, vrd and brdok): the typed broj (its document
// when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) KalNovi(c *gin.Context) {
	h.kalDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: c.Query("brdok")})
}

// KalStopa returns the stopa of a poreska oznaka on a datum (GET po, datum) as JSON for the grids of the porezi (the
// legacy exit of COL_PO: VratiPorez2).
func (h *RobnoDokumentaUnosHandler) KalStopa(c *gin.Context) {
	stopa, err := h.service.GetRobKalStopa(c.Request.Context(), common.StringToInt(c.Query("po")), c.Query("datum"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stopa": stopa})
}

// SaveRobKal saves the header of the kalkulacija ("Sačuvaj" of the header, the legacy BTN_Snimi1): the validation
// (ValidateRobKal, the legacy ValidacijaRDOK: field errors as JSON with status 422, its question as a dialog whose "Da"
// sends the header again with potvrda), the save (SaveRobKal, NoviSLOGRDOK) and the content of the tab with the saved
// document: the header saved and locked, the zavisni troškovi and the stavke enabled. A header saved the first time
// without zavisni troškovi asks whether the document has them (the legacy Dialog of BTN_Snimi1).
func (h *RobnoDokumentaUnosHandler) SaveRobKal(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoKalParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, params.RnalID, common.StringToInt(params.Vrd))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.MagaciniID = dok.MagaciniID
	fieldErrors, pitanje := h.service.ValidateRobKal(ctx, params)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if pitanje != "" {
		h.robnoKalPitanje(c, robnoKalInfoID, robnoKalBodyID, pitanje, robnoKalSaveURL, "#"+robnoKalHeaderPanelID+" [data-panel-fields]")
		return
	}
	if params.RdokID, err = h.service.SaveRobKal(ctx, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoKalEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.dokument.Header.Snimljen = true
	ekran.pitajTrosk = !ekran.dokument.Header.ImaTroskove
	h.robnoKalBody(c, ekran)
}

// PotvrdaBrisanjaKal asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the legacy
// BTN_Brisanje): the artikli whose stanje would be negative after the deletion (the legacy warning "Posle brisanja,
// stanje sledećih artikala biće negativno") and "Brisanje podataka? Da li ste sigurni?"; "Da" deletes it
// (DeleteRobKal), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKal(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	zalihe, err := h.service.GetRobKalZaliheBrisanja(ctx, rdokID)
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruke := []string{}
	if len(zalihe) > 0 {
		poruke = append(poruke, h.translator.Message("Posle brisanja, stanje sledećih artikala biće negativno:"))
		poruke = append(poruke, zalihe...)
	}
	poruke = append(poruke, h.translator.Message("Brisanje podataka? Da li ste sigurni?"))
	dialog := domain.Dialog{Id: robnoKalPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-kal-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-kal-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoKalBrisiURL, rdokID),
		HxTarget:             "#" + robnoKalBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-kal-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm(poruke, dialog, btnClose, btnDa, btnNe, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobKal deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy BTN_Brisanje) and answers
// with the content of the tab with a new document of the vrsta (the legacy BTN_NoviDOK) and the message of the
// deletion. A document that cannot be deleted (posted, fiscalized) keeps the screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobKal(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobKal(ctx, rdokID)
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoKalGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoKalEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.dokument.Poruka = robnoInterniMsgObrisano
	h.robnoKalBody(c, ekran)
}

// KalArtikal returns the artikal of a stavka (GET rdokid, sifra - the šifra or the barkod - and konto) as JSON for the
// form of the stavke (the legacy exit of EDT_SIFRA): see GetRobKalArtikal; found false when the artikal does not exist.
func (h *RobnoDokumentaUnosHandler) KalArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetRobKalArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"),
		c.Query("konto"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// KalStavka returns a saved stavka (GET rdokid, id) as JSON for the form of the stavke (the legacy BTN_Izmeni): see
// GetRobKalStavkaForma.
func (h *RobnoDokumentaUnosHandler) KalStavka(c *gin.Context) {
	stavka, found, err := h.service.GetRobKalStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
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

// SaveRobKalStavka saves a stavka of the kalkulacija (POST of the fields of the stavke, the legacy BTN_Snimi1 of
// rob_Kalkulacije_Veleprodaja_Stavke): the validation (ValidateRobKalStavka: field errors as JSON with status 422, its
// question as a dialog whose "Da" sends the stavka again with potvrda), the save (SaveRobKalStavka, NoviSLOGRPRO) and the
// content of the tab with the document (the stavke open; the konto of the stavka stays for the next one).
func (h *RobnoDokumentaUnosHandler) SaveRobKalStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobKalStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	fieldErrors, pitanje := h.service.ValidateRobKalStavka(ctx, params)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if pitanje != "" {
		h.robnoKalPitanje(c, robnoKalStavkaInfoID, robnoKalBodyID, pitanje, robnoKalStavkaSaveURL, "#"+robnoKalStavkePanelID+" [data-panel-fields]")
		return
	}
	if _, err := h.service.SaveRobKalStavka(ctx, params); err != nil {
		robnoKalStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoKalOdgovor(c, params.RdokID, robnoKalOtvoreneStavke, func(ekran *robnoKalEkranData) {
		if konto := strings.TrimSpace(params.Konto); konto != "" {
			ekran.dokument.Stavka.Konto = konto
		}
	})
}

// PotvrdaBrisanjaKalStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the legacy
// YesNo "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it (DeleteRobKalStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKalStavke(c *gin.Context) {
	h.robnoKalPotvrdaBrisanja(c, robnoKalStavkaPotvrdaDialogID, "robno-kal-stavka-brisanje",
		fmt.Sprintf("%s/%d", robnoKalStavkaBrisiURL, robnoFaktureRproID(c)), "Brisanje podataka? Da li ste sigurni?")
}

// robnoKalPotvrdaBrisanja renders the confirmation of the deletion of a row of the stavke or of the zavisni troškovi:
// "Da" sends the deletion (DELETE url, the content of the screen is the answer), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) robnoKalPotvrdaBrisanja(c *gin.Context, dialogID, prefix, url, poruka string) {
	dialog := domain.Dialog{Id: dialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: prefix + "-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   prefix + "-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          url,
		HxTarget:             "#" + robnoKalBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: prefix + "-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message(poruka)}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// DeleteRobKalStavka deletes a stavka of the kalkulacija (DELETE …/stavka/brisi/:id, the legacy BrisiRPRO) and answers
// with the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobKalStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobKalStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoKalStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoKalOdgovor(c, rdokID, robnoKalOtvoreneStavke, nil)
}

// KalTrosak returns a saved zavisni trošak (GET rdokid, id) as JSON for its form ("Izmeni" of its row, the legacy
// BTN_Izmeni of rob_Zavisni_Troskovi): see GetRobKalTrosakForma.
func (h *RobnoDokumentaUnosHandler) KalTrosak(c *gin.Context) {
	trosak, found, err := h.service.GetRobKalTrosakForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": "Zavisni trošak koji birate za izmenu ne postoji."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "trosak": trosak})
}

// SaveRobKalTrosak saves a zavisni trošak of the kalkulacija (POST of the fields of the zavisni troškovi, the legacy
// BTN_Snimi of rob_Zavisni_Troskovi): the validation (ValidateRobKalTrosak: field errors as JSON with status 422), the
// save (SaveRobKalTrosak, NoviSLOGRZTR) and the content of the tab with the document (the zavisni troškovi open; the
// konto of the dobavljač stays for the next one, the legacy sOldKonto).
func (h *RobnoDokumentaUnosHandler) SaveRobKalTrosak(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobKalTrosakParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobKalTrosak(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobKalTrosak(ctx, params); err != nil {
		robnoKalStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoKalOdgovor(c, params.RdokID, robnoKalOtvoreniTroskovi, func(ekran *robnoKalEkranData) {
		if konto := strings.TrimSpace(params.Konto); konto != "" {
			ekran.dokument.Trosak.KontoDobavljaca = konto
		}
	})
}

// PotvrdaBrisanjaKalTroska asks to confirm the deletion of a zavisni trošak ("Obriši" of its row, GET rdokid and id, the
// legacy Dialog "Brisanje stavke? Da li ste sigurni?"): "Da" deletes it (DeleteRobKalTrosak).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKalTroska(c *gin.Context) {
	h.robnoKalPotvrdaBrisanja(c, robnoKalTrosakPotvrdaDialogID, "robno-kal-trosak-brisanje",
		fmt.Sprintf("%s/%d?rdokid=%d", robnoKalTrosakBrisiURL, robnoFaktureRproID(c), common.StringToInt(c.Query("rdokid"))),
		robnoKalMsgBrisanjeStavkePotvr)
}

// DeleteRobKalTrosak deletes a zavisni trošak of the kalkulacija with its plaćen porez (DELETE …/trosak/brisi/:id with
// the rdokid of the document) and answers with the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobKalTrosak(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	if err := h.service.DeleteRobKalTrosak(c.Request.Context(), rdokID, int64(common.StringToInt(c.Param("id")))); err != nil {
		robnoKalStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoKalOdgovor(c, rdokID, robnoKalOtvoreniTroskovi, nil)
}

// robnoKalOdgovor renders the content of the tab with the document of the stavke or of the zavisni troškovi (its header
// saved, the part otvoreni open): the answer of their saves and deletions; izmena changes the content before it is
// rendered (e.g. keeps the konto of the form for the next stavka).
func (h *RobnoDokumentaUnosHandler) robnoKalOdgovor(c *gin.Context, rdokID int64, otvoreni string, izmena func(ekran *robnoKalEkranData)) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoKalGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoKalEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoKalGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if izmena != nil {
		izmena(&ekran)
	}
	ekran.dokument.Header.Snimljen = true
	ekran.otvoreni = otvoreni
	h.robnoKalBody(c, ekran)
}

// robnoKalRoutes registers the routes of the "Kalkulacije veleprodaje" screen.
func (h *RobnoDokumentaUnosHandler) robnoKalRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/kalkulacije-vp/dokument/:id", h.KalDokument)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/novi", h.KalNovi)
	r.POST("/api/robno-dokumenta/kalkulacije-vp/save", h.SaveRobKal)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/brisi/potvrda", h.PotvrdaBrisanjaKal)
	r.DELETE("/api/robno-dokumenta/kalkulacije-vp/brisi/:id", h.DeleteRobKal)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/stopa", h.KalStopa)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/artikal", h.KalArtikal)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/stavka", h.KalStavka)
	r.POST("/api/robno-dokumenta/kalkulacije-vp/stavka/save", h.SaveRobKalStavka)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/stavka/potvrda", h.PotvrdaBrisanjaKalStavke)
	r.DELETE("/api/robno-dokumenta/kalkulacije-vp/stavka/brisi/:id", h.DeleteRobKalStavka)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/trosak", h.KalTrosak)
	r.POST("/api/robno-dokumenta/kalkulacije-vp/trosak/save", h.SaveRobKalTrosak)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/trosak/potvrda", h.PotvrdaBrisanjaKalTroska)
	r.DELETE("/api/robno-dokumenta/kalkulacije-vp/trosak/brisi/:id", h.DeleteRobKalTrosak)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/prikaz", h.KalPrikaz)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/prikaz/dokumenta", h.KalPrikazDokumenta)
	r.GET("/api/robno-dokumenta/kalkulacije-vp/prikaz/stavke/:id", h.KalPrikazStavke)
}
