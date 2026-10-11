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

// The "Interni prenosi - dostavnice" screen (the legacy rob_Interni_racuni, the vrste dokumenata of the
// groups IRT and IRP): its urls and the ids of its elements (the templates get them by robnoInterniUI).
const (
	robnoInterniURL = robnoDokumentaURLPrefix + "/interni"
	// robnoInterniDokumentURL + "/:id" reloads the content of the screen with an existing document (the
	// selected broj), robnoInterniNoviURL with a new one ("Novi dok.", a typed broj).
	robnoInterniDokumentURL = robnoInterniURL + "/dokument"
	robnoInterniNoviURL     = robnoInterniURL + "/novi"
	// robnoInterniSaveURL saves the header (POST), robnoInterniPotvrdaURL asks to confirm the deletion of the
	// document and robnoInterniBrisiURL + "/:id" deletes it.
	robnoInterniSaveURL    = robnoInterniURL + "/save"
	robnoInterniBrisiURL   = robnoInterniURL + "/brisi"
	robnoInterniPotvrdaURL = robnoInterniBrisiURL + "/potvrda"
	// robnoInterniPrikazURL fills the tab "Pregled internih prenosa" when it is shown, robnoInterniPrikazDokumentaURL
	// reloads its grid (the pagination, the search, the sort), robnoInterniPrikazStavkeURL + "/:id" the stavke
	// (DETALJI) of the selected document.
	robnoInterniPrikazURL          = robnoInterniURL + "/prikaz"
	robnoInterniPrikazDokumentaURL = robnoInterniPrikazURL + "/dokumenta"
	robnoInterniPrikazStavkeURL    = robnoInterniPrikazURL + "/stavke"
	// The stavke (the legacy rob_InterniRacuni_DET): robnoInterniArtikalURL returns the artikal of the stavka
	// (GET rdokid, sifra, konto), robnoInterniStavkaURL a saved stavka for the form (GET rdokid, id: "Izmeni" of
	// its row), robnoInterniStavkaSaveURL saves the stavka (POST), robnoInterniStavkaPotvrdaURL asks to confirm
	// the deletion of a stavka (GET rdokid, id: "Obriši" of its row) and robnoInterniStavkaBrisiURL + "/:id"
	// deletes it.
	robnoInterniArtikalURL       = robnoInterniURL + "/artikal"
	robnoInterniStavkaURL        = robnoInterniURL + "/stavka"
	robnoInterniStavkaSaveURL    = robnoInterniStavkaURL + "/save"
	robnoInterniStavkaPotvrdaURL = robnoInterniStavkaURL + "/potvrda"
	robnoInterniStavkaBrisiURL   = robnoInterniStavkaURL + "/brisi"

	robnoInterniDialogID         = "robno-interni-dialog"
	robnoInterniBodyID           = "robno-interni-body"
	robnoInterniInfoID           = "robno-interni-info"
	robnoInterniInfoDialogID     = "robno-interni-info-dialog"
	robnoInterniPotvrdaDialogID  = "robno-interni-brisanje"
	robnoInterniStavkaInfoID     = "robno-interni-stavka-info"
	// The dialogs of the stavke (the error of the save, the confirmation of the deletion) in robnoInterniStavkaInfoID.
	robnoInterniStavkaInfoDialogID    = "robno-interni-stavka-info-dialog"
	robnoInterniStavkaPotvrdaDialogID = "robno-interni-stavka-brisanje"
	robnoInterniHeaderPanelID    = "robno-interni-header-panel"
	robnoInterniStavkePanelID    = "robno-interni-stavke-panel"
	robnoInterniPrikazID         = "robno-interni-prikaz"
	robnoInterniPrikazStavkeID   = "robno-interni-prikaz-stavke"
	robnoInterniPrikazRdokID     = "robno-interni-prikaz-rdokid"
	robnoInterniPrikazSearchID   = "robno-interni-prikaz-search"
	robnoInterniPrikazStampaID   = "robno-interni-prikaz-stampa"
	robnoInterniMsgObrisano      = "Uspešno obrisani podaci..."
	robnoInterniMsgBrisanjeDokum = "Brisanje dokumenta?"

	// Ids of the grids: the stavke of the document, the documents and the stavke of the selected one of
	// the tab "Pregled internih prenosa".
	robnoInterniStavkeTableID         = "robno-interni-stavke-table"
	robnoInterniDokumentaTableID      = "robno-interni-dokumenta-table"
	robnoInterniDokumentStavkeTableID = "robno-interni-dokument-stavke-table"
)

// robnoInterniUI returns the ids and the urls of the "Interni prenosi - dostavnice" screen. The tab
// "Pregled internih prenosa" has no url yet (its grids are filled later).
func robnoInterniUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoInterniHeaderPanelID,
		StavkePanelID:     robnoInterniStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoInterniDialogID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoInterniDokumentURL,
		ArtikalURL:        robnoInterniArtikalURL,
		NoviURL:           robnoInterniNoviURL,
		BodyID:            robnoInterniBodyID,
		InfoID:            robnoInterniInfoID,
		StavkaInfoID:      robnoInterniStavkaInfoID,
		PrikazID:          robnoInterniPrikazID,
		PrikazStavkeID:    robnoInterniPrikazStavkeID,
		PrikazRdokID:      robnoInterniPrikazRdokID,
		PrikazStampaID:    robnoInterniPrikazStampaID,
		PrikazURL:         robnoInterniPrikazURL,
		// The legacy title of the grid of the stavke of the second tab.
		PrikazStavkeNaslov: "Detalji",
	}
}

// robnoInterniPrikazOpis is the tab "Pregled internih prenosa": the documents of the vrste of the legacy
// window (robnosvc.RobInterniVrste) in the magacin of the nalog and the stavke (DETALJI) of the selected one.
func (h *RobnoDokumentaUnosHandler) robnoInterniPrikazOpis() robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoInterniUI(),
		dokumentaTableID: robnoInterniDokumentaTableID,
		stavkeTableID:    robnoInterniDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Detalji",
		dokumentaFields:  h.service.GetInterniDokumentaTableFields(),
		stavkeFields:     h.service.GetInterniDokumentStavkeTableFields(),
		dokumentaRows:    robnosvc.RobInterniDokumentaTableRows,
		stavkeRows:       robnosvc.RobInterniDokumentStavkeTableRows,
		dokumentaURL:     robnoInterniPrikazDokumentaURL,
		stavkeURL:        robnoInterniPrikazStavkeURL,
		searchID:         robnoInterniPrikazSearchID,
		vrste:            func(int) []int { return robnosvc.RobInterniVrste },
		greska:           robnoInterniGreska,
	}
}

// InterniPrikaz fills the tab "Pregled internih prenosa" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) InterniPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoInterniPrikazOpis())
}

// InterniPrikazDokumenta returns the grid of the documents of the tab "Pregled internih prenosa" only (the
// pagination, the search and the sort of the grid).
func (h *RobnoDokumentaUnosHandler) InterniPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoInterniPrikazOpis())
}

// InterniPrikazStavke returns the stavke (DETALJI) of the document of the id (the click of a row of the
// documents of the tab "Pregled internih prenosa").
func (h *RobnoDokumentaUnosHandler) InterniPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoInterniPrikazOpis())
}

// robnoInterniButtons builds the buttons of the "Interni prenosi - dostavnice" screen for the document of the
// nalog (rdokID 0: a new one), like the ones of the faktura: "Sačuvaj" saves the header and opens the stavke,
// "Izmeni" switches from the stavke back to the header, "Novi dok." a new document, "Briši" deletes the saved
// one (after its confirmation), "Nazad" releases the lock of the nalog and closes the dialog, "Štampaj" prints
// the document. "Sačuvaj" of the stavka saves it (the legacy "Snimi" of rob_InterniRacuni_DET), "Odustani"
// gives up the change of a stavka (enabled by "Izmeni" of its row).
func robnoInterniButtons(rnalID, rdokID int64, vrd int) domain.RobnoFaktureButtons {
	disabled := func(id, label, icon, class string) domain.Button {
		return domain.Button{Id: id, LabelText: label, Icon: icon, IsVisible: true, IsDisabled: true,
			BtnClass: class + " opacity-50 cursor-not-allowed"}
	}
	btns := domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:            "robno-interni-sacuvaj",
			LabelText:     "Sačuvaj",
			Icon:          "save",
			BtnClass:      common.ClassSaveButton,
			HxActionURL:   robnoInterniSaveURL,
			HxRequestType: "POST",
			// Only the fields of the header; the answer is the content of the tab with the saved document (the
			// header locked, the stavke open) or the JSON of the field errors (status 422, not swapped).
			HxInclude:            "#" + robnoInterniHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoInterniBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoInterniHeaderPanelID, robnoInterniStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-interni-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoInterniHeaderPanelID, robnoInterniStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-interni-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoInterniNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoInterniBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-interni-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoInterniPotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoInterniInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-interni-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoInterniDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta), with the
		// rdokid of the form (the hidden rdokid of the stavke, id stavkardokid), like the faktura.
		Print: common.SetPrintButton("robno-interni-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		// The content of the tab with the saved stavka, or the JSON of the field errors (status 422, not swapped).
		SaveStavka: domain.Button{
			Id:                   "robno-interni-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoInterniStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoInterniStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoInterniBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-interni-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robInterniStavkaOdustani",
		},
	}
	// A new document has nothing to delete.
	if rdokID <= 0 {
		btns.Delete = disabled("robno-interni-brisi", "Briši", "delete", common.ClassDeleteButton)
	}
	return btns
}

// robnoInterniEkranData is the content of the tab "Interni prenosi" (the header, the stavka, the grid of the
// stavke and the buttons) with the message of the screen.
type robnoInterniEkranData struct {
	header domain.RobnoInterniHeaderView
	stavka domain.RobnoInterniStavka
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	poruka string
}

// robnoInterniEkran builds the content of the tab "Interni prenosi" with the document of the params (a new
// one for an empty broj, see GetRobInterniDokument): the strip of the nalog, the header with the prodavnica,
// the stavka, the grid of the stavke and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoInterniEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoInterniEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoInterniEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobInterniDokument(ctx, params)
	if err != nil {
		return ekran, err
	}
	ekran.header, ekran.stavka, ekran.poruka = dokument.Header, dokument.Stavka, dokument.Poruka
	robnoNalogStrip(&ekran.header.RobnoFaktureHeaderView, rnal, magacin)
	// The vrsta of a document opened by its id is the one of the document.
	vrd := dok.Vrd
	if v := common.StringToInt(ekran.header.Vrd); v > 0 {
		vrd = v
	}
	ekran.tbl = h.robnoFaktureTable("Stavke internog prenosa", robnoInterniStavkeTableID, h.service.GetInterniStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, ekran.header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobInterniStavkeTableRows(stavke)
	// The buttons of every row: "Izmeni" puts the stavka in the form (robInterniStavkaUcitaj), "Obriši" asks to
	// confirm its deletion; both send the document with the id of the row.
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, ekran.header.RdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoInterniStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robInterniStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoInterniStavkaPotvrdaURL,
		HxTarget: "#" + robnoInterniStavkaInfoID, HxSwap: "innerHTML"}
	ekran.btns = robnoInterniButtons(dok.RnalID, ekran.header.RdokID, vrd)
	return ekran, nil
}

// robnoInterniBody renders the content of the tab "Interni prenosi" (the answer of the save, of a selected
// or new document and of the deletion), with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoInterniBody(c *gin.Context, ekran robnoInterniEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobInterniBody(robnoInterniUI(), ekran.header, ekran.stavka, ekran.tbl, robnosvc.RobSistemPdvCombo(), ekran.btns,
		h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoInterniInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-interni-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-interni-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoInterniInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoInterniGreska shows an error in the dialog of the messages of the screen instead of the target of
// the request.
func robnoInterniGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoInterniInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoInterniInfoDialogID, message, true)
}

// otvoriInterniRacun opens the "Interni prenosi - dostavnice" screen (the legacy rob_Interni_racuni) of the
// nalog as a dialog with its two tabs, with a new document (the next broj of the vrsta in the magacin,
// NoviBrojRDOK). The nalog stays locked while the dialog is open: "Zatvori" and "Nazad" release it.
func (h *RobnoDokumentaUnosHandler) otvoriInterniRacun(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoInterniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	// The grids of the tab "Pregled internih prenosa" (filled when the tab is shown: InterniPrikaz).
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoInterniPrikazOpis(), nil)

	// "Zatvori" of the title bar releases the lock of the nalog, like the one of the faktura.
	btnClose := domain.Button{
		Id:            "robno-interni-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoInterniDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_InterniRacuni(robnoInterniUI(), ekran.header, ekran.stavka, ekran.tbl, dokumentaTbl, dokumentStavkeTbl, robnosvc.RobSistemPdvCombo(),
		ekran.btns, btnClose, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// interniDokument renders the content of the tab "Interni prenosi" with the document of the params: the
// nalog (its vrsta, broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) interniDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoInterniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoInterniEkran(c, dok, params)
	if err != nil {
		robnoInterniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoInterniBody(c, ekran)
}

// InterniDokument reloads the content of the tab "Interni prenosi" with the existing document of the id (GET
// /dokument/:id with the rnalid of the screen: the selected broj). A document that cannot be changed (of
// another nalog, posted, fiscalized) gives a new document and the message why.
func (h *RobnoDokumentaUnosHandler) InterniDokument(c *gin.Context) {
	h.interniDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// InterniNovi reloads the content of the tab "Interni prenosi" with a new document (GET rnalid, vrd and brdok):
// the typed broj (its document when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) InterniNovi(c *gin.Context) {
	h.interniDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: c.Query("brdok")})
}

// SaveRobInterni saves the header of the interni prenos ("Sačuvaj" of the header, POST of the fields of the
// header), like the legacy "Sačuvaj" of rob_Interni_racuni: the validation (ValidateRobInterni, the legacy
// ValidacijaRDOK), the save (SaveRobInterni, SnimiRDOK) and the content of the tab with the saved document:
// the header saved and locked, the stavke open (the legacy opens rob_InterniRacuni_DET). A validation error
// is the JSON of the field errors (status 422, the content is not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobInterni(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoInterniParams
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
	if fieldErrors := h.service.ValidateRobInterni(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if params.RdokID, err = h.service.SaveRobInterni(ctx, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoInterniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	// The saved header is locked, the stavke are entered.
	ekran.header.Snimljen = true
	h.robnoInterniBody(c, ekran)
}

// PotvrdaBrisanjaInterni asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the
// legacy Dialog("Brisanje dokumenta?")): "Da" deletes it (DeleteRobInterni), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaInterni(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dialog := domain.Dialog{Id: robnoInterniPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-interni-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-interni-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoInterniBrisiURL, rdokID),
		HxTarget:             "#" + robnoInterniBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-interni-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{robnoInterniMsgBrisanjeDokum}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobInterni deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy "Briši" of
// rob_Interni_racuni, DeleteRobInterni) and answers with the content of the tab with a new document of the
// vrsta (the legacy Reset and NoviBrojRDOK) and the message of the deletion. A document that cannot be deleted
// (posted, fiscalized) keeps the screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobInterni(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoInterniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobInterni(ctx, rdokID)
	if err != nil {
		robnoInterniGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoInterniGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoInterniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoInterniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoInterniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.poruka = robnoInterniMsgObrisano
	h.robnoInterniBody(c, ekran)
}

// InterniArtikal returns the artikal of a stavka (GET rdokid, sifra - the šifra or the barkod - and konto) as
// JSON for the form of the stavke (the legacy exit of EDT_SIFRA and EDT_KONTO): see GetRobInterniArtikal;
// found false when the artikal does not exist.
func (h *RobnoDokumentaUnosHandler) InterniArtikal(c *gin.Context) {
	ctx := c.Request.Context()
	artikal, found, err := h.service.GetRobInterniArtikal(ctx, int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"), c.Query("konto"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// InterniStavka returns a saved stavka (GET rdokid, id) as JSON for the form of the stavke (the legacy
// "Izmene"): see GetRobInterniStavkaForma.
func (h *RobnoDokumentaUnosHandler) InterniStavka(c *gin.Context) {
	ctx := c.Request.Context()
	stavka, found, err := h.service.GetRobInterniStavkaForma(ctx, int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
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

// SaveRobInterniStavka saves a stavka of the interni prenos (POST of the fields of the stavke, the legacy
// "Snimi" of rob_InterniRacuni_DET): the validation (ValidateRobInterniStavka, the legacy Validacija), the save
// (SaveRobInterniStavka: CreateRPRO or UpdateRPRO) and the content of the tab with the document; the konto of
// the stavka stays for the next one (the legacy sOldKonto). A validation error is the JSON of the field errors
// (status 422, the content is not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobInterniStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobInterniStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobInterniStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobInterniStavka(ctx, params); err != nil {
		// The error dialog goes into its container of the stavke instead of the content of the tab.
		c.Header("HX-Retarget", "#"+robnoInterniStavkaInfoID)
		c.Header("HX-Reswap", "innerHTML")
		utils.RenderDialogOK(c, robnoInterniStavkaInfoDialogID, common.ErrMsgSaveData+": "+err.Error(), true)
		return
	}
	h.robnoInterniStavkeOdgovor(c, params.RdokID, params.Konto)
}

// PotvrdaBrisanjaInterniStavke asks to confirm the deletion of a stavka ("Obriši" of its row in the grid, GET
// rdokid and id): "Da" deletes it (DeleteRobInterniStavka) and the content of the tab comes back.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaInterniStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rproID := robnoFaktureRproID(c)
	dialog := domain.Dialog{Id: robnoInterniStavkaPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-interni-stavka-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-interni-stavka-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoInterniStavkaBrisiURL, rproID),
		HxTarget:             "#" + robnoInterniBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-interni-stavka-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message("Da li želite da obrišete stavku?")}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobInterniStavka deletes a stavka of the interni prenos (DELETE …/stavka/brisi/:id, the legacy
// BrisiRPRO: the stavke after it are numbered again; its document is read from the stavka) and answers with
// the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobInterniStavka(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID, err := h.service.DeleteRobInterniStavka(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		c.Header("HX-Retarget", "#"+robnoInterniStavkaInfoID)
		c.Header("HX-Reswap", "innerHTML")
		utils.RenderDialogOK(c, robnoInterniStavkaInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), true)
		return
	}
	h.robnoInterniStavkeOdgovor(c, rdokID, "")
}

// robnoInterniStavkeOdgovor renders the content of the tab "Interni prenosi" with the document of the stavke
// (its header saved, the stavke open): the answer of the save and of the deletion of a stavka. A konto keeps
// the konto of the form for the next stavka.
func (h *RobnoDokumentaUnosHandler) robnoInterniStavkeOdgovor(c *gin.Context, rdokID int64, konto string) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoInterniGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoInterniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoInterniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if konto = strings.TrimSpace(konto); konto != "" {
		ekran.stavka.Konto = konto
	}
	ekran.header.Snimljen = true
	h.robnoInterniBody(c, ekran)
}

// robnoInterniRoutes registers the routes of the "Interni prenosi - dostavnice" screen.
func (h *RobnoDokumentaUnosHandler) robnoInterniRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/interni/dokument/:id", h.InterniDokument)
	r.GET("/api/robno-dokumenta/interni/novi", h.InterniNovi)
	r.POST("/api/robno-dokumenta/interni/save", h.SaveRobInterni)
	r.GET("/api/robno-dokumenta/interni/brisi/potvrda", h.PotvrdaBrisanjaInterni)
	r.DELETE("/api/robno-dokumenta/interni/brisi/:id", h.DeleteRobInterni)
	r.GET("/api/robno-dokumenta/interni/artikal", h.InterniArtikal)
	r.GET("/api/robno-dokumenta/interni/stavka", h.InterniStavka)
	r.POST("/api/robno-dokumenta/interni/stavka/save", h.SaveRobInterniStavka)
	r.GET("/api/robno-dokumenta/interni/stavka/potvrda", h.PotvrdaBrisanjaInterniStavke)
	r.DELETE("/api/robno-dokumenta/interni/stavka/brisi/:id", h.DeleteRobInterniStavka)
	r.GET("/api/robno-dokumenta/interni/prikaz", h.InterniPrikaz)
	r.GET("/api/robno-dokumenta/interni/prikaz/dokumenta", h.InterniPrikazDokumenta)
	r.GET("/api/robno-dokumenta/interni/prikaz/stavke/:id", h.InterniPrikazStavke)
}
