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

// The "Profakture" screen (the legacy rob_ProFakture with its stavke rob_PROFAKTURE_DET, the vrste dokumenata of
// the group PRO): its urls and the ids of its elements (the templates get them by robnoProUI).
const (
	robnoProURL = robnoDokumentaURLPrefix + "/profakture"
	// robnoProDokumentURL + "/:id" reloads the content of the screen with an existing document (the selected
	// broj), robnoProNoviURL with a new one ("Novi dok.", a typed broj).
	robnoProDokumentURL = robnoProURL + "/dokument"
	robnoProNoviURL     = robnoProURL + "/novi"
	// robnoProSaveURL saves the header (POST), robnoProPotvrdaURL asks to confirm the deletion of the document
	// and robnoProBrisiURL + "/:id" deletes it.
	robnoProSaveURL    = robnoProURL + "/save"
	robnoProBrisiURL   = robnoProURL + "/brisi"
	robnoProPotvrdaURL = robnoProBrisiURL + "/potvrda"
	// robnoProPrikazURL fills the tab "Prikaz profaktura" when it is shown, robnoProPrikazDokumentaURL reloads
	// its grid, robnoProPrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoProPrikazURL          = robnoProURL + "/prikaz"
	robnoProPrikazDokumentaURL = robnoProPrikazURL + "/dokumenta"
	robnoProPrikazStavkeURL    = robnoProPrikazURL + "/stavke"
	// The stavke: robnoProArtikalURL returns the artikal of the stavka (GET rdokid, sifra, konto) and + "/stopa"
	// the stopa of a tarifa (GET rdokid, tarifa), robnoProStavkaURL a saved stavka for the form (GET rdokid, id),
	// robnoProStavkaSaveURL saves the stavka (POST), robnoProStavkaPotvrdaURL asks to confirm the deletion of a
	// stavka and robnoProStavkaBrisiURL + "/:id" deletes it.
	robnoProArtikalURL       = robnoProURL + "/artikal"
	robnoProStavkaURL        = robnoProURL + "/stavka"
	robnoProStavkaSaveURL    = robnoProStavkaURL + "/save"
	robnoProStavkaPotvrdaURL = robnoProStavkaURL + "/potvrda"
	robnoProStavkaBrisiURL   = robnoProStavkaURL + "/brisi"

	robnoProDialogID              = "robno-pro-dialog"
	robnoProBodyID                = "robno-pro-body"
	robnoProInfoID                = "robno-pro-info"
	robnoProInfoDialogID          = "robno-pro-info-dialog"
	robnoProPotvrdaDialogID       = "robno-pro-brisanje"
	robnoProStavkaInfoID          = "robno-pro-stavka-info"
	robnoProStavkaInfoDialogID    = "robno-pro-stavka-info-dialog"
	robnoProStavkaPotvrdaDialogID = "robno-pro-stavka-brisanje"
	robnoProHeaderPanelID         = "robno-pro-header-panel"
	robnoProStavkePanelID         = "robno-pro-stavke-panel"
	robnoProPrikazID              = "robno-pro-prikaz"
	robnoProPrikazStavkeID        = "robno-pro-prikaz-stavke"
	robnoProPrikazRdokID          = "robno-pro-prikaz-rdokid"
	robnoProPrikazSearchID        = "robno-pro-prikaz-search"
	robnoProPrikazStampaID        = "robno-pro-prikaz-stampa"

	// Ids of the grids: the stavke of the document, the documents and the stavke of the selected one of the tab
	// "Prikaz profaktura".
	robnoProStavkeTableID         = "robno-pro-stavke-table"
	robnoProDokumentaTableID      = "robno-pro-dokumenta-table"
	robnoProDokumentStavkeTableID = "robno-pro-dokument-stavke-table"
)

// robnoProUI returns the ids and the urls of the "Profakture" screen.
func robnoProUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoProHeaderPanelID,
		StavkePanelID:     robnoProStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoProDialogID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoProDokumentURL,
		ArtikalURL:        robnoProArtikalURL,
		NoviURL:           robnoProNoviURL,
		BodyID:            robnoProBodyID,
		InfoID:            robnoProInfoID,
		StavkaInfoID:      robnoProStavkaInfoID,
		PrikazID:          robnoProPrikazID,
		PrikazStavkeID:    robnoProPrikazStavkeID,
		PrikazRdokID:      robnoProPrikazRdokID,
		PrikazStampaID:    robnoProPrikazStampaID,
		PrikazURL:         robnoProPrikazURL,
	}
}

// robnoProPrikazOpis is the tab "Prikaz profaktura": the documents of the vrsta in the magacin of the nalog (the
// legacy QRY_RDOK) and the stavke of the selected one (QRY_RPRO_ZADOK).
func (h *RobnoDokumentaUnosHandler) robnoProPrikazOpis() robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoProUI(),
		dokumentaTableID: robnoProDokumentaTableID,
		stavkeTableID:    robnoProDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetFaktureDokumentaTableFields(),
		stavkeFields:     h.service.GetProStavkeTableFields(),
		dokumentaRows:    robnosvc.RobDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobProStavkeTableRows,
		dokumentaURL:     robnoProPrikazDokumentaURL,
		stavkeURL:        robnoProPrikazStavkeURL,
		searchID:         robnoProPrikazSearchID,
		vrste:            func(vrd int) []int { return []int{vrd} },
		greska:           robnoProGreska,
	}
}

// ProPrikaz fills the tab "Prikaz profaktura" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) ProPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoProPrikazOpis())
}

// ProPrikazDokumenta returns the grid of the documents of the tab "Prikaz profaktura" only (the pagination, the
// search and the sort of the grid).
func (h *RobnoDokumentaUnosHandler) ProPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoProPrikazOpis())
}

// ProPrikazStavke returns the stavke of the document of the id (the click of a row of the documents of the tab
// "Prikaz profaktura").
func (h *RobnoDokumentaUnosHandler) ProPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoProPrikazOpis())
}

// robnoProButtons builds the buttons of the "Profakture" screen for the document of the nalog (rdokID 0: a new
// one), like the ones of the porudžbenica: "Sačuvaj" saves the header and opens the stavke, "Izmeni" switches from
// the stavke back to the header, "Novi dok." a new document, "Briši" deletes the saved one (after its
// confirmation), "Nazad" releases the lock of the nalog and closes the dialog, "Štampaj" prints the document as a
// profaktura or a ponuda (the radio tipprof). "Sačuvaj" of the stavka saves it (the legacy "Sačuvaj" of
// rob_PROFAKTURE_DET), "Odustani" gives up the change of a stavka (enabled by "Izmeni" of its row).
func robnoProButtons(rnalID, rdokID int64, vrd int) domain.RobnoFaktureButtons {
	btns := domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-pro-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoProSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoProHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoProBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoProHeaderPanelID, robnoProStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-pro-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoProHeaderPanelID, robnoProStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-pro-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoProNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoProBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-pro-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoProPotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoProInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-pro-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoProDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta, the group PRO) as a
		// profaktura or a ponuda (the radio tipprof of the strip).
		Print: common.SetPrintButton("robno-pro-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid,tipprof"),
		SaveStavka: domain.Button{
			Id:                   "robno-pro-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoProStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoProStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoProBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-pro-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robProStavkaOdustani",
		},
	}
	// A new document has nothing to delete.
	if rdokID <= 0 {
		btns.Delete = domain.Button{Id: "robno-pro-brisi", LabelText: "Briši", Icon: "delete", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassDeleteButton + " opacity-50 cursor-not-allowed"}
	}
	return btns
}

// robnoProEkranData is the content of the tab "Profakture" (the header, the stavka, the grid of the stavke and the
// buttons) with the message of the screen.
type robnoProEkranData struct {
	header domain.RobnoProHeaderView
	stavka domain.RobnoProStavka
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	poruka string
}

// robnoProEkran builds the content of the tab "Profakture" with the document of the params (a new one for an empty
// broj, see GetRobProDokument): the strip of the nalog, the header, the stavka, the grid of the stavke and the
// buttons.
func (h *RobnoDokumentaUnosHandler) robnoProEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoProEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoProEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobProDokument(ctx, params)
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
	ekran.tbl = h.robnoFaktureTable("Stavke profakture", robnoProStavkeTableID, h.service.GetProStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, ekran.header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobProStavkeTableRows(stavke)
	// The buttons of every row: "Izmeni" puts the stavka in the form (robProStavkaUcitaj), "Obriši" asks to
	// confirm its deletion; both send the document with the id of the row.
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, ekran.header.RdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoProStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robProStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoProStavkaPotvrdaURL,
		HxTarget: "#" + robnoProStavkaInfoID, HxSwap: "innerHTML"}
	ekran.btns = robnoProButtons(dok.RnalID, ekran.header.RdokID, vrd)
	return ekran, nil
}

// robnoProBody renders the content of the tab "Profakture" (the answer of the save, of a selected or new document
// and of the deletion), with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoProBody(c *gin.Context, ekran robnoProEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobProBody(robnoProUI(), ekran.header, ekran.stavka, ekran.tbl, robnosvc.RobSistemPdvCombo(), ekran.btns,
		h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoProInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-pro-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-pro-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoProInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoProGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func robnoProGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoProInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoProInfoDialogID, message, true)
}

// robnoProStavkaGreska shows an error of the stavke in the container of the dialogs of the stavke.
func robnoProStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoProStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoProStavkaInfoDialogID, message, true)
}

// otvoriProfakturu opens the "Profakture" screen (the legacy rob_ProFakture) of the nalog as a dialog with its two
// tabs, with a new document (the next broj of the vrsta in the magacin, NoviBrojRDOK). The nalog stays locked
// while the dialog is open: "Zatvori" and "Nazad" release it (the legacy BrisiRNAL on close).
func (h *RobnoDokumentaUnosHandler) otvoriProfakturu(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoProEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoProPrikazOpis(), nil)
	btnClose := domain.Button{
		Id:            "robno-pro-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoProDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_ProFakture(robnoProUI(), ekran.header, ekran.stavka, ekran.tbl, dokumentaTbl, dokumentStavkeTbl, robnosvc.RobSistemPdvCombo(),
		ekran.btns, btnClose, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// proDokument renders the content of the tab "Profakture" with the document of the params: the nalog (its vrsta,
// broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) proDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoProGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoProEkran(c, dok, params)
	if err != nil {
		robnoProGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoProBody(c, ekran)
}

// ProDokument reloads the content of the tab "Profakture" with the existing document of the id (GET /dokument/:id
// with the rnalid of the screen: the selected broj). A document that cannot be changed (of another nalog, posted,
// fiscalized) gives a new document and the message why (the legacy exit of EDT_DOKUM).
func (h *RobnoDokumentaUnosHandler) ProDokument(c *gin.Context) {
	h.proDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// ProNovi reloads the content of the tab "Profakture" with a new document (GET rnalid, vrd and brdok): the typed
// broj (its document when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) ProNovi(c *gin.Context) {
	h.proDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: c.Query("brdok")})
}

// SaveRobPro saves the header of the profaktura ("Sačuvaj" of the header, the legacy BTN_Snimi1): the validation
// (ValidateRobPro, the legacy ValidacijaRDOK), the save (SaveRobPro, SnimiRDOK) and the content of the tab with
// the saved document: the header saved and locked, the stavke open (the legacy opens rob_PROFAKTURE_DET). A
// validation error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobPro(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoProParams
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
	if fieldErrors := h.service.ValidateRobPro(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if params.RdokID, err = h.service.SaveRobPro(ctx, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoProEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = true
	h.robnoProBody(c, ekran)
}

// PotvrdaBrisanjaPro asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the legacy
// Dialog("Brisanje dokumenta?")): "Da" deletes it (DeleteRobPro), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaPro(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dialog := domain.Dialog{Id: robnoProPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-pro-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-pro-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoProBrisiURL, rdokID),
		HxTarget:             "#" + robnoProBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-pro-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{robnoInterniMsgBrisanjeDokum}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobPro deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy "Brisanje") and
// answers with the content of the tab with a new document of the vrsta (the legacy Reset and NoviBrojRDOK) and the
// message of the deletion. A document that cannot be deleted (posted, fiscalized) keeps the screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobPro(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoProGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobPro(ctx, rdokID)
	if err != nil {
		robnoProGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoProGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoProGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoProEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoProGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.poruka = robnoInterniMsgObrisano
	h.robnoProBody(c, ekran)
}

// ProArtikal returns the artikal of a stavka (GET rdokid, sifra - the šifra or the barkod - and konto) as JSON for
// the form of the stavke (the legacy exit of EDT_SIFRA and EDT_KONTO): see GetRobProArtikal; found false when the
// artikal does not exist.
func (h *RobnoDokumentaUnosHandler) ProArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetRobProArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"), c.Query("konto"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// SaveRobProStavka saves a stavka of the profaktura (POST of the fields of the stavke, the legacy "Sačuvaj" of
// rob_PROFAKTURE_DET): the validation (ValidateRobProStavka, the legacy Validacija), the save (SaveRobProStavka:
// CreateRpro or EditRpro) and the content of the tab with the document; the konto of the stavka stays for the next
// one. A validation error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobProStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobProStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobProStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobProStavka(ctx, params); err != nil {
		robnoProStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoProStavkeOdgovor(c, params.RdokID, params.Konto)
}

// PotvrdaBrisanjaProStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the
// legacy YesNo "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it (DeleteRobProStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaProStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rproID := robnoFaktureRproID(c)
	dialog := domain.Dialog{Id: robnoProStavkaPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-pro-stavka-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-pro-stavka-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoProStavkaBrisiURL, rproID),
		HxTarget:             "#" + robnoProBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-pro-stavka-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message("Da li želite da obrišete stavku?")}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobProStavka deletes a stavka of the profaktura (DELETE …/stavka/brisi/:id, the legacy BrisiRPRO) and
// answers with the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobProStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobProStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoProStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoProStavkeOdgovor(c, rdokID, "")
}

// robnoProStavkeOdgovor renders the content of the tab "Profakture" with the document of the stavke (its header
// saved, the stavke open): the answer of the save and of the deletion of a stavka. A konto keeps the konto of the
// form for the next stavka.
func (h *RobnoDokumentaUnosHandler) robnoProStavkeOdgovor(c *gin.Context, rdokID int64, konto string) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoProGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoProEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoProGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if konto = strings.TrimSpace(konto); konto != "" {
		ekran.stavka.Konto = konto
	}
	ekran.header.Snimljen = true
	h.robnoProBody(c, ekran)
}

// robnoProRoutes registers the routes of the "Profakture" screen.
func (h *RobnoDokumentaUnosHandler) robnoProRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/profakture/dokument/:id", h.ProDokument)
	r.GET("/api/robno-dokumenta/profakture/novi", h.ProNovi)
	r.POST("/api/robno-dokumenta/profakture/save", h.SaveRobPro)
	r.GET("/api/robno-dokumenta/profakture/brisi/potvrda", h.PotvrdaBrisanjaPro)
	r.DELETE("/api/robno-dokumenta/profakture/brisi/:id", h.DeleteRobPro)
	r.GET("/api/robno-dokumenta/profakture/artikal", h.ProArtikal)
	// The stopa of a tarifa: the one of the stavke of the faktura usluga (the legacy exit of EDT_po).
	r.GET("/api/robno-dokumenta/profakture/artikal/stopa", h.UslugeStopa)
	// A saved stavka for the form: the one of the faktura (GetRobStavkaForma; its naziv is rpro.naz1 when it has one).
	r.GET("/api/robno-dokumenta/profakture/stavka", h.FaktureStavka)
	r.POST("/api/robno-dokumenta/profakture/stavka/save", h.SaveRobProStavka)
	r.GET("/api/robno-dokumenta/profakture/stavka/potvrda", h.PotvrdaBrisanjaProStavke)
	r.DELETE("/api/robno-dokumenta/profakture/stavka/brisi/:id", h.DeleteRobProStavka)
	r.GET("/api/robno-dokumenta/profakture/prikaz", h.ProPrikaz)
	r.GET("/api/robno-dokumenta/profakture/prikaz/dokumenta", h.ProPrikazDokumenta)
	r.GET("/api/robno-dokumenta/profakture/prikaz/stavke/:id", h.ProPrikazStavke)
}
