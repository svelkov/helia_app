package robno

import (
	"fmt"
	"net/http"

	tmpl "helia/frontend/templates"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/internal/common"
	"helia/internal/domain"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// The "Knjižna pisma" screen (the legacy rob_KnjiznoPismoFIN with its stavke rob_KnjiznoPismoFINDET and the window
// rob_KnjPismoPR_Obracun, the vrste dokumenata of the groups KNO and KNZ): its urls and the ids of its elements (the
// templates get them by robnoKnpUI).
const (
	robnoKnpURL = robnoDokumentaURLPrefix + "/knjizna-pisma"
	// robnoKnpDokumentURL + "/:id" reloads the content of the screen with an existing document (the selected broj),
	// robnoKnpNoviURL with a new one ("Novi dok.", a typed broj).
	robnoKnpDokumentURL = robnoKnpURL + "/dokument"
	robnoKnpNoviURL     = robnoKnpURL + "/novi"
	// robnoKnpSaveURL saves the header (POST), robnoKnpPotvrdaURL asks to confirm the deletion of the document and
	// robnoKnpBrisiURL + "/:id" deletes it; robnoKnpVezniURL returns the vezni dokument (GET rdokid, JSON) and
	// robnoKnpObracunURL the window of the obračun (GET rdokid, "Pregl. Vr. Fakt.").
	robnoKnpSaveURL    = robnoKnpURL + "/save"
	robnoKnpBrisiURL   = robnoKnpURL + "/brisi"
	robnoKnpPotvrdaURL = robnoKnpBrisiURL + "/potvrda"
	robnoKnpVezniURL   = robnoKnpURL + "/vezni"
	robnoKnpObracunURL = robnoKnpURL + "/obracun"
	// robnoKnpPrikazURL fills the tab "Prikaz knjižnih pisama" when it is shown, robnoKnpPrikazDokumentaURL reloads
	// its grid, robnoKnpPrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoKnpPrikazURL          = robnoKnpURL + "/prikaz"
	robnoKnpPrikazDokumentaURL = robnoKnpPrikazURL + "/dokumenta"
	robnoKnpPrikazStavkeURL    = robnoKnpPrikazURL + "/stavke"
	// The stavke: robnoKnpStavkaURL returns a corrected stavka for the form (GET rdokid, id), robnoKnpStavkaSaveURL
	// saves it (POST), robnoKnpStavkaDodajURL + "/:id" copies the stavka of the faktura of the id (POST rdokid,
	// "Dodaj"), robnoKnpStavkaPotvrdaURL asks to confirm the deletion of a stavka and robnoKnpStavkaBrisiURL + "/:id"
	// deletes it ("Vrati jedan"); robnoKnpStavkeDodajURL copies all the stavke of the faktura (POST rdokid, "Dodaj
	// sve"), robnoKnpStavkePotvrdaURL asks to confirm the deletion of all of them and robnoKnpStavkeBrisiURL + "/:id"
	// (the document) deletes them ("Vrati sve").
	robnoKnpStavkaURL        = robnoKnpURL + "/stavka"
	robnoKnpStavkaSaveURL    = robnoKnpStavkaURL + "/save"
	robnoKnpStavkaDodajURL   = robnoKnpStavkaURL + "/dodaj"
	robnoKnpStavkaPotvrdaURL = robnoKnpStavkaURL + "/potvrda"
	robnoKnpStavkaBrisiURL   = robnoKnpStavkaURL + "/brisi"
	robnoKnpStavkeURL        = robnoKnpURL + "/stavke"
	robnoKnpStavkeDodajURL   = robnoKnpStavkeURL + "/dodaj"
	robnoKnpStavkePotvrdaURL = robnoKnpStavkeURL + "/potvrda"
	robnoKnpStavkeBrisiURL   = robnoKnpStavkeURL + "/brisi"

	robnoKnpDialogID               = "robno-knp-dialog"
	robnoKnpBodyID                 = "robno-knp-body"
	robnoKnpInfoID                 = "robno-knp-info"
	robnoKnpInfoDialogID           = "robno-knp-info-dialog"
	robnoKnpObracunDialogID        = "robno-knp-obracun"
	robnoKnpPotvrdaDialogID        = "robno-knp-brisanje"
	robnoKnpStavkaInfoID           = "robno-knp-stavka-info"
	robnoKnpStavkaInfoDialogID     = "robno-knp-stavka-info-dialog"
	robnoKnpStavkaPotvrdaDialogID  = "robno-knp-stavka-brisanje"
	robnoKnpStavkePotvrdaDialogID  = "robno-knp-stavke-brisanje"
	robnoKnpHeaderPanelID          = "robno-knp-header-panel"
	robnoKnpStavkePanelID          = "robno-knp-stavke-panel"
	robnoKnpPrikazID               = "robno-knp-prikaz"
	robnoKnpPrikazStavkeID         = "robno-knp-prikaz-stavke"
	robnoKnpPrikazRdokID           = "robno-knp-prikaz-rdokid"
	robnoKnpPrikazSearchID         = "robno-knp-prikaz-search"
	robnoKnpPrikazStampaID         = "robno-knp-prikaz-stampa"
	robnoKnpFakturaTableID         = "robno-knp-faktura-table"
	robnoKnpKorigovaneTableID      = "robno-knp-korigovane-table"
	robnoKnpDokumentaTableID       = "robno-knp-dokumenta-table"
	robnoKnpDokumentStavkeTableID  = "robno-knp-dokument-stavke-table"
	robnoKnpMsgBrisanjeStavki      = "Brisanje svih korigovanih stavki? Da li ste sigurni?"
	robnoKnpStavkaFormaPoljaFilter = "#stavkardokid, #knpstavkarproid, #kolicina, #prodajnacena, #rabatstavke"
)

// robnoKnpUI returns the ids and the urls of the "Knjižna pisma" screen.
func robnoKnpUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoKnpHeaderPanelID,
		StavkePanelID:     robnoKnpStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoKnpDialogID,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoKnpDokumentURL,
		NoviURL:           robnoKnpNoviURL,
		VezniURL:          robnoKnpVezniURL,
		BodyID:            robnoKnpBodyID,
		InfoID:            robnoKnpInfoID,
		StavkaInfoID:      robnoKnpStavkaInfoID,
		PrikazID:          robnoKnpPrikazID,
		PrikazStavkeID:    robnoKnpPrikazStavkeID,
		PrikazRdokID:      robnoKnpPrikazRdokID,
		PrikazStampaID:    robnoKnpPrikazStampaID,
		PrikazURL:         robnoKnpPrikazURL,
	}
}

// robnoKnpPrikazOpis is the tab "Prikaz knjižnih pisama": the documents of the vrsta in the magacin of the nalog
// (the legacy QRY_RDOK) and the stavke of the selected one (QRY_RPRO_ZADOK).
func (h *RobnoDokumentaUnosHandler) robnoKnpPrikazOpis() robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoKnpUI(),
		dokumentaTableID: robnoKnpDokumentaTableID,
		stavkeTableID:    robnoKnpDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetKnpDokumentaTableFields(),
		stavkeFields:     h.service.GetKnpStavkeTableFields(),
		dokumentaRows:    robnosvc.RobKnpDokumentaPrikazTableRows,
		stavkeRows: func(stavke []domain.RobStavkaRow) []domain.TableRow {
			return robnosvc.RobKnpStavkeTableRows(stavke, true)
		},
		dokumentaURL: robnoKnpPrikazDokumentaURL,
		stavkeURL:    robnoKnpPrikazStavkeURL,
		searchID:     robnoKnpPrikazSearchID,
		vrste:        func(vrd int) []int { return []int{vrd} },
		greska:       robnoKnpGreska,
	}
}

// KnpPrikaz fills the tab "Prikaz knjižnih pisama" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) KnpPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoKnpPrikazOpis())
}

// KnpPrikazDokumenta returns the grid of the documents of the tab "Prikaz knjižnih pisama" only (the pagination,
// the search and the sort of the grid).
func (h *RobnoDokumentaUnosHandler) KnpPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoKnpPrikazOpis())
}

// KnpPrikazStavke returns the stavke of the document of the id (the click of a row of the documents of the tab
// "Prikaz knjižnih pisama").
func (h *RobnoDokumentaUnosHandler) KnpPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoKnpPrikazOpis())
}

// robnoKnpButtons builds the buttons of the "Knjižna pisma" screen for the document of the nalog (rdokID 0: a new
// one), like the ones of the profaktura: "Sačuvaj" saves the header (and opens the stavke of a knjižno pismo with a
// vezni dokument), "Izmeni" switches from the stavke back to the header, "Novi dok." a new document, "Briši"
// deletes the saved one (after its confirmation), "Pregl. Vr. Fakt." opens the window of the obračun, "Nazad"
// releases the lock of the nalog and closes the dialog, "Štampaj" prints the document. The stavke: "Dodaj sve"
// copies all the stavke of the faktura, "Vrati sve" deletes all the corrected stavke (after the confirmation),
// "Sačuvaj" of the stavka saves the corrected one and "Odustani" gives up its change (both enabled by "Izmeni" of
// its row).
func robnoKnpButtons(rnalID, rdokID int64, vrd int) domain.RobnoKnpButtons {
	btns := domain.RobnoKnpButtons{RobnoFaktureButtons: domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-knp-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoKnpSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoKnpHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoKnpBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoKnpHeaderPanelID, robnoKnpStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-knp-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoKnpHeaderPanelID, robnoKnpStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-knp-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoKnpNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoKnpBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-knp-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoKnpPotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoKnpInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-knp-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoKnpDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta, the groups KNO and KNZ:
		// with stavke ROB_RPT_KNJPISMO, without them ROB_RPT_KNJPISMOFIN).
		Print: common.SetPrintButton("robno-knp-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		SaveStavka: domain.Button{
			Id:                   "robno-knp-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			IsDisabled:           true,
			BtnClass:             common.ClassSaveButton + " opacity-50",
			HxActionURL:          robnoKnpStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            robnoKnpStavkaFormaPoljaFilter,
			HxTarget:             "#" + robnoKnpBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-knp-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50",
			HxOnClick:  "robKnpStavkaOdustani",
		},
	}}
	pregled := func(id string) domain.Button {
		return domain.Button{
			Id:            id,
			LabelText:     "Pregl. Vr. Fakt.",
			Icon:          "stampa",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoKnpObracunURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoKnpInfoID,
			HxSwap:        "innerHTML",
		}
	}
	btns.Pregled, btns.PregledStavke = pregled("robno-knp-pregled"), pregled("robno-knp-stavke-pregled")
	btns.DodajSve = domain.Button{
		Id:            "robno-knp-dodaj-sve",
		LabelText:     "Dodaj sve",
		Icon:          "add",
		BtnClass:      common.ClassNewButton,
		HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoKnpStavkeDodajURL, rdokID),
		HxRequestType: "POST",
		HxTarget:      "#" + robnoKnpBodyID,
		HxSwap:        "outerHTML",
	}
	btns.VratiSve = domain.Button{
		Id:            "robno-knp-vrati-sve",
		LabelText:     "Vrati sve",
		Icon:          "delete",
		BtnClass:      common.ClassDeleteButton,
		HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoKnpStavkePotvrdaURL, rdokID),
		HxRequestType: "GET",
		HxTarget:      "#" + robnoKnpStavkaInfoID,
		HxSwap:        "innerHTML",
	}
	// A new document has nothing to delete or to show.
	if rdokID <= 0 {
		nedostupno := func(b domain.Button) domain.Button {
			return domain.Button{Id: b.Id, LabelText: b.LabelText, Icon: b.Icon, IsVisible: true, IsDisabled: true,
				BtnClass: b.BtnClass + " opacity-50 cursor-not-allowed"}
		}
		btns.Delete, btns.Pregled = nedostupno(btns.Delete), nedostupno(btns.Pregled)
	}
	return btns
}

// robnoKnpEkranData is the content of the tab "Knjižna pisma" (the header, the stavke with their two grids and the
// buttons) with the message of the screen.
type robnoKnpEkranData struct {
	header                    domain.RobnoKnpHeaderView
	stavke                    domain.RobnoKnpStavke
	fakturaTbl, korigovaneTbl domain.TableData
	btns                      domain.RobnoKnpButtons
	poruka                    string
}

// robnoKnpEkran builds the content of the tab "Knjižna pisma" with the document of the params (a new one for an
// empty broj, see GetRobKnpDokument): the strip of the nalog, the header, the grids of the stavke of the faktura
// ("Dodaj" of a row copies it) and of the corrected stavke ("Izmeni" puts the stavka in the form, "Obriši" asks to
// confirm its deletion) and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoKnpEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoKnpEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoKnpEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobKnpDokument(ctx, params)
	if err != nil {
		return ekran, err
	}
	ekran.header, ekran.stavke, ekran.poruka = dokument.Header, dokument.Stavke, dokument.Poruka
	robnoNalogStrip(&ekran.header.RobnoFaktureHeaderView, rnal, magacin)
	vrd := dok.Vrd
	if v := common.StringToInt(ekran.header.Vrd); v > 0 {
		vrd = v
	}
	fields := h.service.GetKnpStavkeTableFields()
	ekran.fakturaTbl = h.robnoFaktureTable("Stavke po izdatoj fakturi", robnoKnpFakturaTableID, fields)
	ekran.korigovaneTbl = h.robnoFaktureTable("Korigovane stavke", robnoKnpKorigovaneTableID, fields)
	if rdokID := ekran.header.RdokID; rdokID > 0 && ekran.header.Vezan() {
		faktura, korigovane, err := h.service.GetRobKnpStavke(ctx, rdokID)
		if err != nil {
			return ekran, err
		}
		vals := fmt.Sprintf(`{"rdokid": "%d"}`, rdokID)
		ekran.fakturaTbl.Rows = robnosvc.RobKnpStavkeTableRows(faktura, true)
		ekran.fakturaTbl.ShowActions = len(faktura) > 0
		ekran.fakturaTbl.HxVals = vals
		ekran.fakturaTbl.BtnUpdate = domain.Button{LabelText: "Dodaj", IsVisible: true, HxRequestType: "POST", HxActionURL: robnoKnpStavkaDodajURL,
			HxTarget: "#" + robnoKnpBodyID, HxSwap: "outerHTML"}
		ekran.korigovaneTbl.Rows = robnosvc.RobKnpStavkeTableRows(korigovane, false)
		ekran.korigovaneTbl.ShowActions = len(korigovane) > 0
		ekran.korigovaneTbl.HxVals = vals
		ekran.korigovaneTbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoKnpStavkaURL,
			HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robKnpStavkaUcitaj(event)"}
		ekran.korigovaneTbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoKnpStavkaPotvrdaURL,
			HxTarget: "#" + robnoKnpStavkaInfoID, HxSwap: "innerHTML"}
	}
	ekran.btns = robnoKnpButtons(dok.RnalID, ekran.header.RdokID, vrd)
	return ekran, nil
}

// robnoKnpBody renders the content of the tab "Knjižna pisma" (the answer of the save, of a selected or new document,
// of the stavke and of the deletion), with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoKnpBody(c *gin.Context, ekran robnoKnpEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobKnpBody(robnoKnpUI(), ekran.header, ekran.stavke, ekran.fakturaTbl, ekran.korigovaneTbl, robnosvc.RobSistemPdvCombo(),
		ekran.btns, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoKnpInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-knp-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-knp-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoKnpInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoKnpGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func robnoKnpGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoKnpInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoKnpInfoDialogID, message, true)
}

// robnoKnpStavkaGreska shows an error of the stavke in the container of the dialogs of the stavke.
func robnoKnpStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoKnpStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoKnpStavkaInfoDialogID, message, true)
}

// otvoriKnjiznoPismo opens the "Knjižna pisma" screen (the legacy rob_KnjiznoPismoFIN) of the nalog as a dialog with
// its two tabs, with a new document (the next broj of the vrsta in the magacin, NoviBrojRDOK). The nalog stays
// locked while the dialog is open: "Zatvori" and "Nazad" release it (the legacy BrisiRNAL on close).
func (h *RobnoDokumentaUnosHandler) otvoriKnjiznoPismo(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoKnpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoKnpPrikazOpis(), nil)
	btnClose := domain.Button{
		Id:            "robno-knp-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoKnpDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_KnjiznoPismo(robnoKnpUI(), ekran.header, ekran.stavke, ekran.fakturaTbl, ekran.korigovaneTbl, dokumentaTbl, dokumentStavkeTbl,
		robnosvc.RobSistemPdvCombo(), ekran.btns, btnClose, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// knpDokument renders the content of the tab "Knjižna pisma" with the document of the params: the nalog (its vrsta,
// broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) knpDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoKnpEkran(c, dok, params)
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	// A saved knjižno pismo with a vezni dokument opens with its stavke.
	ekran.header.Snimljen = ekran.header.RdokID > 0 && ekran.header.Vezan()
	h.robnoKnpBody(c, ekran)
}

// KnpDokument reloads the content of the tab "Knjižna pisma" with the existing document of the id (GET
// /dokument/:id with the rnalid of the screen: the selected broj). A document that cannot be changed (of another
// nalog, posted, fiscalized) gives a new document and the message why (the legacy BTN_TraziDOK).
func (h *RobnoDokumentaUnosHandler) KnpDokument(c *gin.Context) {
	h.knpDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// KnpNovi reloads the content of the tab "Knjižna pisma" with a new document (GET rnalid, vrd and brdok): the typed
// broj (its document when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) KnpNovi(c *gin.Context) {
	h.knpDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: c.Query("brdok")})
}

// SaveRobKnp saves the header of the knjižno pismo ("Sačuvaj" of the header, the legacy BTN_Snimi): the validation
// (ValidateRobKnp, the legacy ValidacijaRDOK), the save (SaveRobKnp, SnimiRDOK) and the content of the tab with the
// saved document: with a vezni dokument the header locked and the stavke open (the legacy opens
// rob_KnjiznoPismoFINDET), without it the header stays open. A validation error is the JSON of the field errors
// (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobKnp(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoKnpParams
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
	if fieldErrors := h.service.ValidateRobKnp(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if params.RdokID, err = h.service.SaveRobKnp(ctx, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoKnpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = ekran.header.Vezan()
	h.robnoKnpBody(c, ekran)
}

// robnoKnpPotvrda renders a confirmation of the screen: "Da" sends the DELETE of the url into the content of the tab,
// "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) robnoKnpPotvrda(c *gin.Context, dialogID, poruka, url string) {
	dialog := domain.Dialog{Id: dialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: dialogID + "-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   dialogID + "-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          url,
		HxTarget:             "#" + robnoKnpBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: dialogID + "-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{poruka}, dialog, btnClose, btnDa, btnNe, h.translator, common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// PotvrdaBrisanjaKnp asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the legacy
// Dialog("Brisanje dokumenta?")): "Da" deletes it (DeleteRobKnp).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKnp(c *gin.Context) {
	h.robnoKnpPotvrda(c, robnoKnpPotvrdaDialogID, robnoInterniMsgBrisanjeDokum,
		fmt.Sprintf("%s/%d", robnoKnpBrisiURL, common.StringToInt(c.Query("rdokid"))))
}

// DeleteRobKnp deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy "Brisanje") and answers
// with the content of the tab with a new document of the vrsta (the legacy Reset and NoviBrojRDOK) and the message
// of the deletion. A document that cannot be deleted (posted, fiscalized) keeps the screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobKnp(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobKnp(ctx, rdokID)
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoKnpGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoKnpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.poruka = robnoInterniMsgObrisano
	h.robnoKnpBody(c, ekran)
}

// KnpVezni returns the vezni dokument of the id (GET rdokid: the faktura selected in the header, the legacy
// BTN_TraziOTP) as JSON: see GetRobKnpVezni; found false when it is not a document of the period.
func (h *RobnoDokumentaUnosHandler) KnpVezni(c *gin.Context) {
	vezni, found, err := h.service.GetRobKnpVezni(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "vezni": vezni})
}

// KnpObracun opens the window "Pregled vrednosti fakture po starim i novim uslovima prodaje" of the knjižno pismo
// (GET rdokid: "Pregl. Vr. Fakt.", the legacy rob_KnjPismoPR_Obracun); a document without a vezni dokument or without
// stavke shows why.
func (h *RobnoDokumentaUnosHandler) KnpObracun(c *gin.Context) {
	ctx := c.Request.Context()
	redovi, poruka, err := h.service.GetRobKnpObracun(ctx, int64(common.StringToInt(c.Query("rdokid"))))
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		robnoKnpGreska(c, poruka)
		return
	}
	dialog := domain.Dialog{Id: robnoKnpObracunDialogID, Title: "Pregled vrednosti fakture"}
	btnClose := domain.Button{Id: "robno-knp-obracun-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	_ = tmpl_robno.RobKnpObracun(dialog, redovi, btnClose, h.translator).Render(ctx, c.Writer)
}

// KnpStavka returns a corrected stavka (GET rdokid, id) as JSON for the form of the stavke (the legacy row of
// TABLE_rpro1): see GetRobKnpStavkaForma.
func (h *RobnoDokumentaUnosHandler) KnpStavka(c *gin.Context) {
	stavka, found, err := h.service.GetRobKnpStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
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

// SaveRobKnpStavka saves the corrected values of a stavka (POST rdokid, rproid, kolicina, prodajnacena, rabatstavke:
// the legacy exit of a row of TABLE_rpro1) and answers with the content of the tab with the document. Without a
// stavka of "Izmeni" the error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobKnpStavka(c *gin.Context) {
	var params domain.RobKnpStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if params.RproID <= 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false,
			[]domain.FieldError{{Field: "kolicina", ErrorMessage: "Izaberite stavku (Izmeni)."}}, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveRobKnpStavka(c.Request.Context(), params); err != nil {
		robnoKnpStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoKnpStavkeOdgovor(c, params.RdokID, "")
}

// DodajKnpStavku copies the stavka of the faktura of the id into the knjižno pismo (POST /stavka/dodaj/:id with
// rdokid, "Dodaj" of its row, the legacy BTN_Dodaj): a stavka already copied gives the message why.
func (h *RobnoDokumentaUnosHandler) DodajKnpStavku(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.PostForm("rdokid")))
	poruka, err := h.service.DodajRobKnpStavke(c.Request.Context(), rdokID, []int64{int64(common.StringToInt(c.Param("id")))})
	if err != nil {
		robnoKnpStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoKnpStavkeOdgovor(c, rdokID, poruka)
}

// DodajKnpStavke copies all the stavke of the faktura into the knjižno pismo (POST rdokid, "Dodaj sve", the legacy
// BTN_DodajSve) but the ones already copied.
func (h *RobnoDokumentaUnosHandler) DodajKnpStavke(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.DodajRobKnpStavke(c.Request.Context(), rdokID, nil)
	if err != nil {
		robnoKnpStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoKnpStavkeOdgovor(c, rdokID, poruka)
}

// PotvrdaBrisanjaKnpStavke asks to confirm the deletion of a corrected stavka ("Obriši" of its row, GET rdokid and
// id, the legacy "Vrati jedan"): "Da" deletes it (DeleteRobKnpStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKnpStavke(c *gin.Context) {
	h.robnoKnpPotvrda(c, robnoKnpStavkaPotvrdaDialogID, h.translator.Message("Da li želite da obrišete stavku?"),
		fmt.Sprintf("%s/%d", robnoKnpStavkaBrisiURL, robnoFaktureRproID(c)))
}

// DeleteRobKnpStavka deletes a corrected stavka (DELETE …/stavka/brisi/:id) and answers with the content of the tab
// with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobKnpStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobKnpStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoKnpStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoKnpStavkeOdgovor(c, rdokID, "")
}

// PotvrdaBrisanjaKnpStavki asks to confirm the deletion of all the corrected stavke (GET rdokid: "Vrati sve", the
// legacy BTN_VratiSve): "Da" deletes them (DeleteRobKnpStavke).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKnpStavki(c *gin.Context) {
	h.robnoKnpPotvrda(c, robnoKnpStavkePotvrdaDialogID, robnoKnpMsgBrisanjeStavki,
		fmt.Sprintf("%s/%d", robnoKnpStavkeBrisiURL, common.StringToInt(c.Query("rdokid"))))
}

// DeleteRobKnpStavke deletes all the corrected stavke of the knjižno pismo of the id (DELETE …/stavke/brisi/:id) and
// answers with the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobKnpStavke(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Param("id")))
	if err := h.service.DeleteRobKnpStavke(c.Request.Context(), rdokID); err != nil {
		robnoKnpStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoKnpStavkeOdgovor(c, rdokID, "")
}

// robnoKnpStavkeOdgovor renders the content of the tab "Knjižna pisma" with the document of the stavke (its header
// saved, the stavke open) and the message poruka (e.g. a stavka already copied): the answer of the changes of the
// stavke.
func (h *RobnoDokumentaUnosHandler) robnoKnpStavkeOdgovor(c *gin.Context, rdokID int64, poruka string) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoKnpGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoKnpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoKnpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = true
	if poruka != "" {
		ekran.poruka = poruka
	}
	h.robnoKnpBody(c, ekran)
}

// robnoKnpRoutes registers the routes of the "Knjižna pisma" screen.
func (h *RobnoDokumentaUnosHandler) robnoKnpRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/knjizna-pisma/dokument/:id", h.KnpDokument)
	r.GET("/api/robno-dokumenta/knjizna-pisma/novi", h.KnpNovi)
	r.POST("/api/robno-dokumenta/knjizna-pisma/save", h.SaveRobKnp)
	r.GET("/api/robno-dokumenta/knjizna-pisma/brisi/potvrda", h.PotvrdaBrisanjaKnp)
	r.DELETE("/api/robno-dokumenta/knjizna-pisma/brisi/:id", h.DeleteRobKnp)
	r.GET("/api/robno-dokumenta/knjizna-pisma/vezni", h.KnpVezni)
	r.GET("/api/robno-dokumenta/knjizna-pisma/obracun", h.KnpObracun)
	r.GET("/api/robno-dokumenta/knjizna-pisma/stavka", h.KnpStavka)
	r.POST("/api/robno-dokumenta/knjizna-pisma/stavka/save", h.SaveRobKnpStavka)
	r.POST("/api/robno-dokumenta/knjizna-pisma/stavka/dodaj/:id", h.DodajKnpStavku)
	r.GET("/api/robno-dokumenta/knjizna-pisma/stavka/potvrda", h.PotvrdaBrisanjaKnpStavke)
	r.DELETE("/api/robno-dokumenta/knjizna-pisma/stavka/brisi/:id", h.DeleteRobKnpStavka)
	r.POST("/api/robno-dokumenta/knjizna-pisma/stavke/dodaj", h.DodajKnpStavke)
	r.GET("/api/robno-dokumenta/knjizna-pisma/stavke/potvrda", h.PotvrdaBrisanjaKnpStavki)
	r.DELETE("/api/robno-dokumenta/knjizna-pisma/stavke/brisi/:id", h.DeleteRobKnpStavke)
	r.GET("/api/robno-dokumenta/knjizna-pisma/prikaz", h.KnpPrikaz)
	r.GET("/api/robno-dokumenta/knjizna-pisma/prikaz/dokumenta", h.KnpPrikazDokumenta)
	r.GET("/api/robno-dokumenta/knjizna-pisma/prikaz/stavke/:id", h.KnpPrikazStavke)
}
