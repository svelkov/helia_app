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

// The "Kalkulacije maloprodaje" screen (the legacy rob_Kalkulacije_Maloprodaja with its stavke
// rob_Kalkulacije_Maloprodaja_Stavke, the vrste dokumenata of the group KAL): its urls and the ids of its elements (the
// templates get them by robnoKmpUI and robnoKmpEkranUI).
const (
	robnoKmpURL = robnoDokumentaURLPrefix + "/kalkulacije-mp"
	// robnoKmpDokumentURL + "/:id" reloads the content of the screen with an existing document (the selected broj),
	// robnoKmpNoviURL with a new one ("Novi dok.", a typed broj).
	robnoKmpDokumentURL = robnoKmpURL + "/dokument"
	robnoKmpNoviURL     = robnoKmpURL + "/novi"
	// robnoKmpSaveURL saves the header (POST), robnoKmpPotvrdaURL asks to confirm the deletion of the document and
	// robnoKmpBrisiURL + "/:id" deletes it; robnoKmpStopaURL returns the stopa of a poreska oznaka on a datum (GET po,
	// datum: the grid of the plaćen porez and the tarifa of a stavka of the šifra 0).
	robnoKmpSaveURL    = robnoKmpURL + "/save"
	robnoKmpBrisiURL   = robnoKmpURL + "/brisi"
	robnoKmpPotvrdaURL = robnoKmpBrisiURL + "/potvrda"
	robnoKmpStopaURL   = robnoKmpURL + "/stopa"
	// robnoKmpPrikazURL fills the tab "Pregled kalkulacija" when it is shown, robnoKmpPrikazDokumentaURL reloads its
	// grid, robnoKmpPrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoKmpPrikazURL          = robnoKmpURL + "/prikaz"
	robnoKmpPrikazDokumentaURL = robnoKmpPrikazURL + "/dokumenta"
	robnoKmpPrikazStavkeURL    = robnoKmpPrikazURL + "/stavke"
	// The stavke: robnoKmpArtikalURL returns the artikal of the stavka (GET rdokid, sifra, konto), robnoKmpStavkaURL a
	// saved stavka for the form ("Izmeni" of its row, GET rdokid, id), robnoKmpStavkaSaveURL saves the stavka (POST),
	// robnoKmpStavkaPotvrdaURL asks to confirm the deletion of a stavka and robnoKmpStavkaBrisiURL + "/:id" deletes it.
	robnoKmpArtikalURL       = robnoKmpURL + "/artikal"
	robnoKmpStavkaURL        = robnoKmpURL + "/stavka"
	robnoKmpStavkaSaveURL    = robnoKmpStavkaURL + "/save"
	robnoKmpStavkaPotvrdaURL = robnoKmpStavkaURL + "/potvrda"
	robnoKmpStavkaBrisiURL   = robnoKmpStavkaURL + "/brisi"

	robnoKmpDialogID              = "robno-kmp-dialog"
	robnoKmpBodyID                = "robno-kmp-body"
	robnoKmpInfoID                = "robno-kmp-info"
	robnoKmpInfoDialogID          = "robno-kmp-info-dialog"
	robnoKmpPotvrdaDialogID       = "robno-kmp-brisanje"
	robnoKmpStavkaInfoID          = "robno-kmp-stavka-info"
	robnoKmpStavkaInfoDialogID    = "robno-kmp-stavka-info-dialog"
	robnoKmpStavkaPotvrdaDialogID = "robno-kmp-stavka-brisanje"
	robnoKmpHeaderPanelID         = "robno-kmp-header-panel"
	robnoKmpStavkePanelID         = "robno-kmp-stavke-panel"
	robnoKmpPrikazID              = "robno-kmp-prikaz"
	robnoKmpPrikazStavkeID        = "robno-kmp-prikaz-stavke"
	robnoKmpPrikazRdokID          = "robno-kmp-prikaz-rdokid"
	robnoKmpPrikazSearchID        = "robno-kmp-prikaz-search"
	robnoKmpPrikazStampaID        = "robno-kmp-prikaz-stampa"
	robnoKmpStavkeTableID         = "robno-kmp-stavke-table"
	robnoKmpDokumentaTableID      = "robno-kmp-dokumenta-table"
	robnoKmpDokumentStavkeTableID = "robno-kmp-dokument-stavke-table"
)

// robnoKmpUI returns the ids and the urls of the "Kalkulacije maloprodaje" screen.
func robnoKmpUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoKmpHeaderPanelID,
		StavkePanelID:     robnoKmpStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoKmpDialogID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoKmpDokumentURL,
		ArtikalURL:        robnoKmpArtikalURL,
		NoviURL:           robnoKmpNoviURL,
		BodyID:            robnoKmpBodyID,
		InfoID:            robnoKmpInfoID,
		StavkaInfoID:      robnoKmpStavkaInfoID,
		PrikazID:          robnoKmpPrikazID,
		PrikazStavkeID:    robnoKmpPrikazStavkeID,
		PrikazRdokID:      robnoKmpPrikazRdokID,
		PrikazStampaID:    robnoKmpPrikazStampaID,
		PrikazURL:         robnoKmpPrikazURL,
	}
}

// robnoKmpEkranUI are the urls of the screen the templates need besides domain.RobnoFaktureUI: the stopa of a poreska
// oznaka on a datum and a saved stavka.
func robnoKmpEkranUI() domain.RobnoKalUI {
	return domain.RobnoKalUI{StopaURL: robnoKmpStopaURL, StavkaURL: robnoKmpStavkaURL}
}

// robnoKmpPrikazOpis is the tab "Pregled kalkulacija": the documents of the vrsta in the magacin of the nalog (the
// legacy QRY_RDOK) and the stavke of the selected one (QRY_RPRO_ZADOK; the stavke of the šifra 0 with their naziv), with
// the columns of the kalkulacija veleprodaje.
func (h *RobnoDokumentaUnosHandler) robnoKmpPrikazOpis() robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoKmpUI(),
		dokumentaTableID: robnoKmpDokumentaTableID,
		stavkeTableID:    robnoKmpDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetKalDokumentaTableFields(),
		stavkeFields:     h.service.GetKalPrikazStavkeTableFields(),
		dokumentaRows:    robnosvc.RobKalDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobKalPrikazStavkeTableRows,
		dokumentaURL:     robnoKmpPrikazDokumentaURL,
		stavkeURL:        robnoKmpPrikazStavkeURL,
		searchID:         robnoKmpPrikazSearchID,
		vrste:            func(vrd int) []int { return []int{vrd} },
		greska:           robnoKmpGreska,
	}
}

// KmpPrikaz fills the tab "Pregled kalkulacija" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) KmpPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoKmpPrikazOpis())
}

// KmpPrikazDokumenta returns the grid of the documents of the tab "Pregled kalkulacija" only.
func (h *RobnoDokumentaUnosHandler) KmpPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoKmpPrikazOpis())
}

// KmpPrikazStavke returns the stavke of the document of the id (the click of a row of the tab "Pregled kalkulacija").
func (h *RobnoDokumentaUnosHandler) KmpPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoKmpPrikazOpis())
}

// robnoKmpButtons builds the buttons of the "Kalkulacije maloprodaje" screen for the document of the nalog (rdokID 0: a
// new one), like the ones of the faktura: "Sačuvaj" saves the header and opens the stavke, "Izmeni" switches back to the
// header, "Novi dok." a new document, "Briši" deletes the saved one (after its confirmation), "Nazad" releases the lock
// of the nalog and closes the dialog, "Štampaj" prints the document (the legacy RPT_ROB_KALKULACIJA_MP). "Sačuvaj" of
// the stavka saves it (a new one, or the one of "Izmeni" of its row), "Odustani" gives up the change of a stavka
// (nothing is saved).
func robnoKmpButtons(rnalID, rdokID int64, vrd int) domain.RobnoFaktureButtons {
	btns := domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-kmp-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoKmpSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoKmpHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoKmpBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoKmpHeaderPanelID, robnoKmpStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-kmp-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoKmpHeaderPanelID, robnoKmpStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-kmp-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoKmpNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoKmpBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-kmp-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoKmpPotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoKmpInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-kmp-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoKmpDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta, the group KAL).
		Print: common.SetPrintButton("robno-kmp-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		SaveStavka: domain.Button{
			Id:                   "robno-kmp-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoKmpStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoKmpStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoKmpBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-kmp-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robKmpStavkaOdustani",
		},
	}
	// A new document has nothing to delete.
	if rdokID <= 0 {
		btns.Delete = domain.Button{Id: "robno-kmp-brisi", LabelText: "Briši", Icon: "delete", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassDeleteButton + " opacity-50 cursor-not-allowed"}
	}
	return btns
}

// robnoKmpEkranData is the content of the tab "Kalkulacije maloprodaja" (the header, the entry of a stavka, the grid of
// the stavke and the buttons).
type robnoKmpEkranData struct {
	dokument domain.RobnoKmpDokument
	tbl      domain.TableData
	btns     domain.RobnoFaktureButtons
}

// robnoKmpEkran builds the content of the tab with the document of the params (a new one for an empty broj, see
// GetRobKmpDokument): the strip of the nalog, the header, the stavka with the grid of the stavke ("Izmeni" puts the
// stavka of a row in the form, "Obriši" asks to confirm its deletion) and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoKmpEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoKmpEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoKmpEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	if ekran.dokument, err = h.service.GetRobKmpDokument(ctx, params); err != nil {
		return ekran, err
	}
	header := &ekran.dokument.Header
	robnoNalogStrip(&header.RobnoFaktureHeaderView, rnal, magacin)
	vrd := dok.Vrd
	if v := common.StringToInt(header.Vrd); v > 0 {
		vrd = v
	}
	ekran.tbl = h.robnoFaktureTable("Stavke kalkulacije", robnoKmpStavkeTableID, h.service.GetKmpStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobKmpStavkeTableRows(stavke)
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoKmpStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robKmpStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoKmpStavkaPotvrdaURL,
		HxTarget: "#" + robnoKmpStavkaInfoID, HxSwap: "innerHTML"}
	ekran.btns = robnoKmpButtons(dok.RnalID, header.RdokID, vrd)
	return ekran, nil
}

// robnoKmpBody renders the content of the tab (the answer of the save, of a selected or new document and of the
// deletions) with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoKmpBody(c *gin.Context, ekran robnoKmpEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobKmpBody(robnoKmpUI(), robnoKmpEkranUI(), ekran.dokument, ekran.tbl, robnosvc.RobSistemPdvCombo(), ekran.btns,
		h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if poruka := ekran.dokument.Poruka; poruka != "" {
		dialog := domain.Dialog{Id: robnoKmpInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-kmp-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-kmp-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoKmpInfoID, poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoKmpGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func robnoKmpGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoKmpInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoKmpInfoDialogID, message, true)
}

// robnoKmpStavkaGreska shows an error of the stavke in the container of their dialogs.
func robnoKmpStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoKmpStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoKmpStavkaInfoDialogID, message, true)
}

// otvoriKalkulacijuMP opens the "Kalkulacije maloprodaje" screen (the legacy rob_Kalkulacije_Maloprodaja) of the nalog
// as a dialog with its two tabs, with a new document (the next broj of the vrsta in the magacin, NoviBrojRDOK). The
// nalog stays locked while the dialog is open: "Zatvori" and "Nazad" release it.
func (h *RobnoDokumentaUnosHandler) otvoriKalkulacijuMP(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoKmpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoKmpPrikazOpis(), nil)
	btnClose := domain.Button{
		Id:            "robno-kmp-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoKmpDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_KalkulacijaMP(robnoKmpUI(), robnoKmpEkranUI(), ekran.dokument, ekran.tbl, dokumentaTbl, dokumentStavkeTbl,
		robnosvc.RobSistemPdvCombo(), ekran.btns, btnClose, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// kmpDokument renders the content of the tab with the document of the params: the nalog (its vrsta, broj and magacin)
// is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) kmpDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoKmpEkran(c, dok, params)
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoKmpBody(c, ekran)
}

// KmpDokument reloads the content of the tab with the existing document of the id (GET /dokument/:id with the rnalid
// of the screen: the selected broj). A document that cannot be changed (of another nalog, posted, fiscalized) gives a
// new document and the message why (the legacy exit of EDT_DOKUM).
func (h *RobnoDokumentaUnosHandler) KmpDokument(c *gin.Context) {
	h.kmpDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// KmpNovi reloads the content of the tab with a new document (GET rnalid, vrd and brdok): the typed broj (its document
// when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) KmpNovi(c *gin.Context) {
	h.kmpDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: c.Query("brdok")})
}

// SaveRobKmp saves the header of the kalkulacija maloprodaje ("Sačuvaj" of the header, the legacy BTN_Snimi1): the
// validation (ValidateRobKmp, the legacy ValidacijaRDOK: field errors as JSON with status 422, its question as a dialog
// whose "Da" sends the header again with potvrda), the save (SaveRobKmp, NoviSLOGRDOK) and the content of the tab with
// the saved document: the header saved and locked, the stavke open (the legacy opens
// rob_Kalkulacije_Maloprodaja_Stavke).
func (h *RobnoDokumentaUnosHandler) SaveRobKmp(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoKmpParams
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
	fieldErrors, pitanje := h.service.ValidateRobKmp(ctx, params)
	if len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if pitanje != "" {
		h.robnoKalPitanje(c, robnoKmpInfoID, robnoKmpBodyID, pitanje, robnoKmpSaveURL, "#"+robnoKmpHeaderPanelID+" [data-panel-fields]")
		return
	}
	if params.RdokID, err = h.service.SaveRobKmp(ctx, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoKmpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.dokument.Header.Snimljen = true
	h.robnoKmpBody(c, ekran)
}

// PotvrdaBrisanjaKmp asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the legacy
// BTN_Brisanje): the artikli whose stanje would be negative after the deletion (the legacy warning) and "Brisanje
// podataka? Da li ste sigurni?"; "Da" deletes it (DeleteRobKmp), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKmp(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	zalihe, err := h.service.GetRobKalZaliheBrisanja(ctx, rdokID)
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruke := []string{}
	if len(zalihe) > 0 {
		poruke = append(poruke, h.translator.Message("Posle brisanja, stanje sledećih artikala biće negativno:"))
		poruke = append(poruke, zalihe...)
	}
	poruke = append(poruke, h.translator.Message("Brisanje podataka? Da li ste sigurni?"))
	dialog := domain.Dialog{Id: robnoKmpPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-kmp-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-kmp-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoKmpBrisiURL, rdokID),
		HxTarget:             "#" + robnoKmpBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-kmp-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm(poruke, dialog, btnClose, btnDa, btnNe, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobKmp deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy BTN_Brisanje) and answers
// with the content of the tab with a new document of the vrsta (the legacy BTN_NoviDOK) and the message of the
// deletion. A document that cannot be deleted (posted, fiscalized) keeps the screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobKmp(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobKmp(ctx, rdokID)
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoKmpGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoKmpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.dokument.Poruka = robnoInterniMsgObrisano
	h.robnoKmpBody(c, ekran)
}

// KmpArtikal returns the artikal of a stavka (GET rdokid, sifra - the šifra or the barkod - and konto) as JSON for the
// form of the stavke (the legacy exit of EDT_SIFRA): see GetRobKmpArtikal; found false when the artikal does not exist.
func (h *RobnoDokumentaUnosHandler) KmpArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetRobKmpArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"),
		c.Query("konto"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// KmpStavka returns a saved stavka (GET rdokid, id) as JSON for the form of the stavke ("Izmeni" of its row): see
// GetRobKmpStavkaForma.
func (h *RobnoDokumentaUnosHandler) KmpStavka(c *gin.Context) {
	stavka, found, err := h.service.GetRobKmpStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
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

// SaveRobKmpStavka saves a stavka of the kalkulacija maloprodaje (POST of the fields of the stavke, the legacy
// BTN_Snimi1 of rob_Kalkulacije_Maloprodaja_Stavke; a stavka of "Izmeni" as a change): the validation
// (ValidateRobKmpStavka: field errors as JSON with status 422), the save (SaveRobKmpStavka, NoviSLOGRPRO) and the content
// of the tab with the document (the stavke open).
func (h *RobnoDokumentaUnosHandler) SaveRobKmpStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobKmpStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobKmpStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobKmpStavka(ctx, params); err != nil {
		robnoKmpStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoKmpOdgovor(c, params.RdokID)
}

// PotvrdaBrisanjaKmpStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the legacy
// YesNo "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it (DeleteRobKmpStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaKmpStavke(c *gin.Context) {
	ctx := c.Request.Context()
	dialog := domain.Dialog{Id: robnoKmpStavkaPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-kmp-stavka-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-kmp-stavka-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoKmpStavkaBrisiURL, robnoFaktureRproID(c)),
		HxTarget:             "#" + robnoKmpBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-kmp-stavka-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message("Brisanje podataka? Da li ste sigurni?")}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobKmpStavka deletes a stavka of the kalkulacija maloprodaje (DELETE …/stavka/brisi/:id, the legacy
// BrisiSLOGRPRO) and answers with the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobKmpStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobKmpStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoKmpStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoKmpOdgovor(c, rdokID)
}

// robnoKmpOdgovor renders the content of the tab with the document of the stavke (its header saved, the stavke open):
// the answer of the save and of the deletion of a stavka.
func (h *RobnoDokumentaUnosHandler) robnoKmpOdgovor(c *gin.Context, rdokID int64) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoKmpGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoKmpEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoKmpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.dokument.Header.Snimljen = true
	h.robnoKmpBody(c, ekran)
}

// robnoKmpRoutes registers the routes of the "Kalkulacije maloprodaje" screen.
func (h *RobnoDokumentaUnosHandler) robnoKmpRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/kalkulacije-mp/dokument/:id", h.KmpDokument)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/novi", h.KmpNovi)
	r.POST("/api/robno-dokumenta/kalkulacije-mp/save", h.SaveRobKmp)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/brisi/potvrda", h.PotvrdaBrisanjaKmp)
	r.DELETE("/api/robno-dokumenta/kalkulacije-mp/brisi/:id", h.DeleteRobKmp)
	// The stopa of a poreska oznaka on a datum: the one of the kalkulacija veleprodaje (the legacy VratiPorez2).
	r.GET("/api/robno-dokumenta/kalkulacije-mp/stopa", h.KalStopa)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/artikal", h.KmpArtikal)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/stavka", h.KmpStavka)
	r.POST("/api/robno-dokumenta/kalkulacije-mp/stavka/save", h.SaveRobKmpStavka)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/stavka/potvrda", h.PotvrdaBrisanjaKmpStavke)
	r.DELETE("/api/robno-dokumenta/kalkulacije-mp/stavka/brisi/:id", h.DeleteRobKmpStavka)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/prikaz", h.KmpPrikaz)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/prikaz/dokumenta", h.KmpPrikazDokumenta)
	r.GET("/api/robno-dokumenta/kalkulacije-mp/prikaz/stavke/:id", h.KmpPrikazStavke)
}
