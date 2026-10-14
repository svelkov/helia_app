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

// The "Porudžbenice" screen (the legacy rob_Porudzbenica with its stavke rob_PORUDZBENICE_DET, the vrste
// dokumenata of the group POR): its urls and the ids of its elements (the templates get them by robnoPorUI).
const (
	robnoPorURL = robnoDokumentaURLPrefix + "/porudzbenice"
	// robnoPorDokumentURL + "/:id" reloads the content of the screen with an existing document (the selected
	// broj), robnoPorNoviURL with a new one ("Novi dok.", a typed broj).
	robnoPorDokumentURL = robnoPorURL + "/dokument"
	robnoPorNoviURL     = robnoPorURL + "/novi"
	// robnoPorSaveURL saves the header (POST), robnoPorPotvrdaURL asks to confirm the deletion of the document
	// and robnoPorBrisiURL + "/:id" deletes it.
	robnoPorSaveURL    = robnoPorURL + "/save"
	robnoPorBrisiURL   = robnoPorURL + "/brisi"
	robnoPorPotvrdaURL = robnoPorBrisiURL + "/potvrda"
	// robnoPorPrikazURL fills the tab "Pregled porudžbenica" when it is shown, robnoPorPrikazDokumentaURL reloads
	// its grid, robnoPorPrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoPorPrikazURL          = robnoPorURL + "/prikaz"
	robnoPorPrikazDokumentaURL = robnoPorPrikazURL + "/dokumenta"
	robnoPorPrikazStavkeURL    = robnoPorPrikazURL + "/stavke"
	// The stavke: robnoPorArtikalURL returns the artikal of the stavka (GET rdokid, sifra, konto) and + "/stopa"
	// the stopa of a tarifa (GET rdokid, tarifa), robnoPorStavkaURL a saved stavka for the form (GET rdokid, id),
	// robnoPorStavkaSaveURL saves the stavka (POST), robnoPorStavkaPotvrdaURL asks to confirm the deletion of a
	// stavka and robnoPorStavkaBrisiURL + "/:id" deletes it.
	robnoPorArtikalURL       = robnoPorURL + "/artikal"
	robnoPorStavkaURL        = robnoPorURL + "/stavka"
	robnoPorStavkaSaveURL    = robnoPorStavkaURL + "/save"
	robnoPorStavkaPotvrdaURL = robnoPorStavkaURL + "/potvrda"
	robnoPorStavkaBrisiURL   = robnoPorStavkaURL + "/brisi"

	robnoPorDialogID              = "robno-por-dialog"
	robnoPorBodyID                = "robno-por-body"
	robnoPorInfoID                = "robno-por-info"
	robnoPorInfoDialogID          = "robno-por-info-dialog"
	robnoPorPotvrdaDialogID       = "robno-por-brisanje"
	robnoPorStavkaInfoID          = "robno-por-stavka-info"
	robnoPorStavkaInfoDialogID    = "robno-por-stavka-info-dialog"
	robnoPorStavkaPotvrdaDialogID = "robno-por-stavka-brisanje"
	robnoPorHeaderPanelID         = "robno-por-header-panel"
	robnoPorStavkePanelID         = "robno-por-stavke-panel"
	robnoPorPrikazID              = "robno-por-prikaz"
	robnoPorPrikazStavkeID        = "robno-por-prikaz-stavke"
	robnoPorPrikazRdokID          = "robno-por-prikaz-rdokid"
	robnoPorPrikazSearchID        = "robno-por-prikaz-search"
	robnoPorPrikazStampaID        = "robno-por-prikaz-stampa"

	// Ids of the grids: the stavke of the document, the documents and the stavke of the selected one of the
	// tab "Pregled porudžbenica".
	robnoPorStavkeTableID         = "robno-por-stavke-table"
	robnoPorDokumentaTableID      = "robno-por-dokumenta-table"
	robnoPorDokumentStavkeTableID = "robno-por-dokument-stavke-table"
)

// robnoPorUI returns the ids and the urls of the "Porudžbenice" screen.
func robnoPorUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoPorHeaderPanelID,
		StavkePanelID:     robnoPorStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoPorDialogID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoPorDokumentURL,
		ArtikalURL:        robnoPorArtikalURL,
		NoviURL:           robnoPorNoviURL,
		BodyID:            robnoPorBodyID,
		InfoID:            robnoPorInfoID,
		StavkaInfoID:      robnoPorStavkaInfoID,
		PrikazID:          robnoPorPrikazID,
		PrikazStavkeID:    robnoPorPrikazStavkeID,
		PrikazRdokID:      robnoPorPrikazRdokID,
		PrikazStampaID:    robnoPorPrikazStampaID,
		PrikazURL:         robnoPorPrikazURL,
	}
}

// robnoPorPrikazOpis is the tab "Pregled porudžbenica": the documents of the vrsta in the magacin of the nalog
// (the legacy QRY_RDOK) and the stavke of the selected one (QRY_RPRO_ZADOK).
func (h *RobnoDokumentaUnosHandler) robnoPorPrikazOpis() robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoPorUI(),
		dokumentaTableID: robnoPorDokumentaTableID,
		stavkeTableID:    robnoPorDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetFaktureDokumentaTableFields(),
		stavkeFields:     h.service.GetPorStavkeTableFields(),
		dokumentaRows:    robnosvc.RobDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobPorStavkeTableRows,
		dokumentaURL:     robnoPorPrikazDokumentaURL,
		stavkeURL:        robnoPorPrikazStavkeURL,
		searchID:         robnoPorPrikazSearchID,
		vrste:            func(vrd int) []int { return []int{vrd} },
		greska:           robnoPorGreska,
	}
}

// PorPrikaz fills the tab "Pregled porudžbenica" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) PorPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoPorPrikazOpis())
}

// PorPrikazDokumenta returns the grid of the documents of the tab "Pregled porudžbenica" only (the
// pagination, the search and the sort of the grid).
func (h *RobnoDokumentaUnosHandler) PorPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoPorPrikazOpis())
}

// PorPrikazStavke returns the stavke of the document of the id (the click of a row of the documents of the
// tab "Pregled porudžbenica").
func (h *RobnoDokumentaUnosHandler) PorPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoPorPrikazOpis())
}

// robnoPorButtons builds the buttons of the "Porudžbenice" screen for the document of the nalog (rdokID 0: a
// new one), like the ones of the interni prenos: "Sačuvaj" saves the header and opens the stavke, "Izmeni"
// switches from the stavke back to the header, "Novi dok." a new document, "Briši" deletes the saved one
// (after its confirmation), "Nazad" releases the lock of the nalog and closes the dialog, "Štampaj" prints the
// document. "Sačuvaj" of the stavka saves it (the legacy "Sačuvaj" of rob_PORUDZBENICE_DET), "Odustani" gives
// up the change of a stavka (enabled by "Izmeni" of its row).
func robnoPorButtons(rnalID, rdokID int64, vrd int) domain.RobnoFaktureButtons {
	btns := domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-por-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoPorSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoPorHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoPorBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoPorHeaderPanelID, robnoPorStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-por-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoPorHeaderPanelID, robnoPorStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-por-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoPorNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoPorBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-por-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoPorPotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoPorInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-por-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoPorDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta).
		Print: common.SetPrintButton("robno-por-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		SaveStavka: domain.Button{
			Id:                   "robno-por-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoPorStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoPorStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoPorBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-por-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robPorStavkaOdustani",
		},
	}
	// A new document has nothing to delete.
	if rdokID <= 0 {
		btns.Delete = domain.Button{Id: "robno-por-brisi", LabelText: "Briši", Icon: "delete", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassDeleteButton + " opacity-50 cursor-not-allowed"}
	}
	return btns
}

// robnoPorEkranData is the content of the tab "Porudžbenica" (the header, the stavka, the grid of the stavke
// and the buttons) with the message of the screen.
type robnoPorEkranData struct {
	header domain.RobnoPorHeaderView
	stavka domain.RobnoPorStavka
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	poruka string
}

// robnoPorEkran builds the content of the tab "Porudžbenica" with the document of the params (a new one for an
// empty broj, see GetRobPorDokument): the strip of the nalog, the header, the stavka, the grid of the stavke
// and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoPorEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoPorEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoPorEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobPorDokument(ctx, params)
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
	ekran.tbl = h.robnoFaktureTable("Stavke porudžbenice", robnoPorStavkeTableID, h.service.GetPorStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, ekran.header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobPorStavkeTableRows(stavke)
	// The buttons of every row: "Izmeni" puts the stavka in the form (robPorStavkaUcitaj), "Obriši" asks to
	// confirm its deletion; both send the document with the id of the row.
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, ekran.header.RdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoPorStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robPorStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoPorStavkaPotvrdaURL,
		HxTarget: "#" + robnoPorStavkaInfoID, HxSwap: "innerHTML"}
	ekran.btns = robnoPorButtons(dok.RnalID, ekran.header.RdokID, vrd)
	return ekran, nil
}

// robnoPorBody renders the content of the tab "Porudžbenica" (the answer of the save, of a selected or new
// document and of the deletion), with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoPorBody(c *gin.Context, ekran robnoPorEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobPorBody(robnoPorUI(), ekran.header, ekran.stavka, ekran.tbl, robnosvc.RobSistemPdvCombo(), ekran.btns,
		h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoPorInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-por-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-por-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoPorInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoPorGreska shows an error in the dialog of the messages of the screen instead of the target of the
// request.
func robnoPorGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoPorInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoPorInfoDialogID, message, true)
}

// robnoPorStavkaGreska shows an error of the stavke in the container of the dialogs of the stavke.
func robnoPorStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoPorStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoPorStavkaInfoDialogID, message, true)
}

// otvoriPorudzbenicu opens the "Porudžbenice" screen (the legacy rob_Porudzbenica) of the nalog as a dialog
// with its two tabs, with a new document (the next broj of the vrsta in the magacin, NoviBrojRDOK). The nalog
// stays locked while the dialog is open: "Zatvori" and "Nazad" release it (the legacy BrisiRNAL on close).
func (h *RobnoDokumentaUnosHandler) otvoriPorudzbenicu(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoPorEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoPorPrikazOpis(), nil)
	btnClose := domain.Button{
		Id:            "robno-por-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoPorDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_Porudzbenica(robnoPorUI(), ekran.header, ekran.stavka, ekran.tbl, dokumentaTbl, dokumentStavkeTbl, robnosvc.RobSistemPdvCombo(),
		ekran.btns, btnClose, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// porDokument renders the content of the tab "Porudžbenica" with the document of the params: the nalog (its
// vrsta, broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) porDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoPorGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoPorEkran(c, dok, params)
	if err != nil {
		robnoPorGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoPorBody(c, ekran)
}

// PorDokument reloads the content of the tab "Porudžbenica" with the existing document of the id (GET
// /dokument/:id with the rnalid of the screen: the selected broj). A document that cannot be changed (of
// another nalog, posted, fiscalized) gives a new document and the message why (the legacy exit of EDT_DOKUM).
func (h *RobnoDokumentaUnosHandler) PorDokument(c *gin.Context) {
	h.porDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// PorNovi reloads the content of the tab "Porudžbenica" with a new document (GET rnalid, vrd and brdok): the
// typed broj (its document when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) PorNovi(c *gin.Context) {
	h.porDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: c.Query("brdok")})
}

// SaveRobPor saves the header of the porudžbenica ("Sačuvaj" of the header, the legacy BTN_Snimi1): the
// validation (ValidateRobPor, the legacy ValidacijaRDOK), the save (SaveRobPor, SnimiRDOK) and the content of
// the tab with the saved document: the header saved and locked, the stavke open (the legacy opens
// rob_PORUDZBENICE_DET). A validation error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobPor(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoPorParams
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
	if fieldErrors := h.service.ValidateRobPor(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if params.RdokID, err = h.service.SaveRobPor(ctx, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoPorEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = true
	h.robnoPorBody(c, ekran)
}

// PotvrdaBrisanjaPor asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the
// legacy Dialog("Brisanje dokumenta?")): "Da" deletes it (DeleteRobPor), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaPor(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dialog := domain.Dialog{Id: robnoPorPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-por-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-por-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoPorBrisiURL, rdokID),
		HxTarget:             "#" + robnoPorBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-por-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{robnoInterniMsgBrisanjeDokum}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobPor deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy "Brisanje") and
// answers with the content of the tab with a new document of the vrsta (the legacy Reset and NoviBrojRDOK) and
// the message of the deletion. A document that cannot be deleted (posted, fiscalized) keeps the screen and
// shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobPor(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoPorGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobPor(ctx, rdokID)
	if err != nil {
		robnoPorGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoPorGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoPorGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoPorEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoPorGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.poruka = robnoInterniMsgObrisano
	h.robnoPorBody(c, ekran)
}

// PorArtikal returns the artikal of a stavka (GET rdokid, sifra - the šifra or the barkod - and konto) as JSON
// for the form of the stavke (the legacy exit of EDT_SIFRA and EDT_KONTO): see GetRobPorArtikal; found false
// when the artikal does not exist.
func (h *RobnoDokumentaUnosHandler) PorArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetRobPorArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"), c.Query("konto"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// PorStavka returns a saved stavka (GET rdokid, id) as JSON for the form of the stavke (the legacy "Izmene"):
// see GetRobPorStavkaForma.
func (h *RobnoDokumentaUnosHandler) PorStavka(c *gin.Context) {
	stavka, found, err := h.service.GetRobPorStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
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

// SaveRobPorStavka saves a stavka of the porudžbenica (POST of the fields of the stavke, the legacy "Sačuvaj"
// of rob_PORUDZBENICE_DET): the validation (ValidateRobPorStavka, the legacy Validacija), the save
// (SaveRobPorStavka: CreateRpro or EditRpro) and the content of the tab with the document; the konto of the
// stavka stays for the next one. A validation error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobPorStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobPorStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobPorStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobPorStavka(ctx, params); err != nil {
		robnoPorStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoPorStavkeOdgovor(c, params.RdokID, params.Konto)
}

// PotvrdaBrisanjaPorStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id,
// the legacy YesNo "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it (DeleteRobPorStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaPorStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rproID := robnoFaktureRproID(c)
	dialog := domain.Dialog{Id: robnoPorStavkaPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-por-stavka-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-por-stavka-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoPorStavkaBrisiURL, rproID),
		HxTarget:             "#" + robnoPorBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-por-stavka-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message("Da li želite da obrišete stavku?")}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobPorStavka deletes a stavka of the porudžbenica (DELETE …/stavka/brisi/:id, the legacy BrisiRPRO)
// and answers with the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobPorStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobPorStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoPorStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoPorStavkeOdgovor(c, rdokID, "")
}

// robnoPorStavkeOdgovor renders the content of the tab "Porudžbenica" with the document of the stavke (its
// header saved, the stavke open): the answer of the save and of the deletion of a stavka. A konto keeps the
// konto of the form for the next stavka.
func (h *RobnoDokumentaUnosHandler) robnoPorStavkeOdgovor(c *gin.Context, rdokID int64, konto string) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoPorGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoPorEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoPorGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if konto = strings.TrimSpace(konto); konto != "" {
		ekran.stavka.Konto = konto
	}
	ekran.header.Snimljen = true
	h.robnoPorBody(c, ekran)
}

// robnoPorRoutes registers the routes of the "Porudžbenice" screen.
func (h *RobnoDokumentaUnosHandler) robnoPorRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/porudzbenice/dokument/:id", h.PorDokument)
	r.GET("/api/robno-dokumenta/porudzbenice/novi", h.PorNovi)
	r.POST("/api/robno-dokumenta/porudzbenice/save", h.SaveRobPor)
	r.GET("/api/robno-dokumenta/porudzbenice/brisi/potvrda", h.PotvrdaBrisanjaPor)
	r.DELETE("/api/robno-dokumenta/porudzbenice/brisi/:id", h.DeleteRobPor)
	r.GET("/api/robno-dokumenta/porudzbenice/artikal", h.PorArtikal)
	// The stopa of a tarifa: the one of the stavke of the faktura usluga (the legacy exit of EDT_po).
	r.GET("/api/robno-dokumenta/porudzbenice/artikal/stopa", h.UslugeStopa)
	r.GET("/api/robno-dokumenta/porudzbenice/stavka", h.PorStavka)
	r.POST("/api/robno-dokumenta/porudzbenice/stavka/save", h.SaveRobPorStavka)
	r.GET("/api/robno-dokumenta/porudzbenice/stavka/potvrda", h.PotvrdaBrisanjaPorStavke)
	r.DELETE("/api/robno-dokumenta/porudzbenice/stavka/brisi/:id", h.DeleteRobPorStavka)
	r.GET("/api/robno-dokumenta/porudzbenice/prikaz", h.PorPrikaz)
	r.GET("/api/robno-dokumenta/porudzbenice/prikaz/dokumenta", h.PorPrikazDokumenta)
	r.GET("/api/robno-dokumenta/porudzbenice/prikaz/stavke/:id", h.PorPrikazStavke)
}
