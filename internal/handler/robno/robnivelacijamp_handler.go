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

// The "Nivelacije maloprodaje" screen (the legacy rob_Nivelacija_maloprodaje with its stavke
// rob_Nivelacija_Maloprodaje_DET, the vrste dokumenata of the group MNI): its urls and the ids of its elements (the
// templates get them by robnoMniUI).
const (
	robnoMniURL = robnoDokumentaURLPrefix + "/nivelacije-mp"
	// robnoMniDokumentURL + "/:id" reloads the content of the screen with an existing document (the selected broj),
	// robnoMniNoviURL with a new one ("Novi dok.", a typed broj).
	robnoMniDokumentURL = robnoMniURL + "/dokument"
	robnoMniNoviURL     = robnoMniURL + "/novi"
	// robnoMniSaveURL saves the header (POST), robnoMniPotvrdaURL asks to confirm the deletion of the document and
	// robnoMniBrisiURL + "/:id" deletes it.
	robnoMniSaveURL    = robnoMniURL + "/save"
	robnoMniBrisiURL   = robnoMniURL + "/brisi"
	robnoMniPotvrdaURL = robnoMniBrisiURL + "/potvrda"
	// robnoMniPrikazURL fills the tab "Pregled nivelacija" when it is shown, robnoMniPrikazDokumentaURL reloads its
	// grid, robnoMniPrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoMniPrikazURL          = robnoMniURL + "/prikaz"
	robnoMniPrikazDokumentaURL = robnoMniPrikazURL + "/dokumenta"
	robnoMniPrikazStavkeURL    = robnoMniPrikazURL + "/stavke"
	// The stavke: robnoMniArtikalURL + "/stopa" returns the stopa of a tarifa (GET rdokid, tarifa), robnoMniStavkaURL
	// a saved stavka for the form (GET rdokid, id), robnoMniStavkaSaveURL saves the stavka (POST),
	// robnoMniStavkaPotvrdaURL asks to confirm the deletion of a stavka and robnoMniStavkaBrisiURL + "/:id" deletes
	// it.
	robnoMniArtikalURL       = robnoMniURL + "/artikal"
	robnoMniStavkaURL        = robnoMniURL + "/stavka"
	robnoMniStavkaSaveURL    = robnoMniStavkaURL + "/save"
	robnoMniStavkaPotvrdaURL = robnoMniStavkaURL + "/potvrda"
	robnoMniStavkaBrisiURL   = robnoMniStavkaURL + "/brisi"

	robnoMniDialogID              = "robno-mni-dialog"
	robnoMniBodyID                = "robno-mni-body"
	robnoMniInfoID                = "robno-mni-info"
	robnoMniInfoDialogID          = "robno-mni-info-dialog"
	robnoMniPotvrdaDialogID       = "robno-mni-brisanje"
	robnoMniStavkaInfoID          = "robno-mni-stavka-info"
	robnoMniStavkaInfoDialogID    = "robno-mni-stavka-info-dialog"
	robnoMniStavkaPotvrdaDialogID = "robno-mni-stavka-brisanje"
	robnoMniHeaderPanelID         = "robno-mni-header-panel"
	robnoMniStavkePanelID         = "robno-mni-stavke-panel"
	robnoMniPrikazID              = "robno-mni-prikaz"
	robnoMniPrikazStavkeID        = "robno-mni-prikaz-stavke"
	robnoMniPrikazRdokID          = "robno-mni-prikaz-rdokid"
	robnoMniPrikazSearchID        = "robno-mni-prikaz-search"
	robnoMniPrikazStampaID        = "robno-mni-prikaz-stampa"

	// Ids of the grids: the stavke of the document, the documents and the stavke of the selected one of the tab
	// "Pregled nivelacija".
	robnoMniStavkeTableID         = "robno-mni-stavke-table"
	robnoMniDokumentaTableID      = "robno-mni-dokumenta-table"
	robnoMniDokumentStavkeTableID = "robno-mni-dokument-stavke-table"

	// The tab "Pregled nivelacija" lists the documents of the vrste of the group between these vrste (the legacy
	// QRY_RDOK_TIP with 100 and 199).
	robnoMniPrikazVrdOd = 100
	robnoMniPrikazVrdDo = 199
)

// robnoMniUI returns the ids and the urls of the "Nivelacije maloprodaje" screen.
func robnoMniUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:     robnoMniHeaderPanelID,
		StavkePanelID:     robnoMniStavkePanelID,
		ContentID:         "#" + robnoDokumentaContentID,
		DialogID:          robnoMniDialogID,
		ArtikalSearchURL:  robnoFaktureArtikalSearch,
		PartnerSearchURL:  robnoFakturePartnerSearch,
		DokumentSearchURL: robnoFaktureDokumentSearch,
		DokumentURL:       robnoMniDokumentURL,
		ArtikalURL:        robnoMniArtikalURL,
		NoviURL:           robnoMniNoviURL,
		BodyID:            robnoMniBodyID,
		InfoID:            robnoMniInfoID,
		StavkaInfoID:      robnoMniStavkaInfoID,
		PrikazID:          robnoMniPrikazID,
		PrikazStavkeID:    robnoMniPrikazStavkeID,
		PrikazRdokID:      robnoMniPrikazRdokID,
		PrikazStampaID:    robnoMniPrikazStampaID,
		PrikazURL:         robnoMniPrikazURL,
	}
}

// robnoMniPrikazOpis is the tab "Pregled nivelacija": the documents of the vrste of the group of the vrsta of the
// screen between 100 and 199 in the magacin of the nalog (the legacy QRY_RDOK_TIP) and the stavke of the selected
// one (QRY_RPRO_ZADOK, with the columns of the stavke of the screen).
func (h *RobnoDokumentaUnosHandler) robnoMniPrikazOpis(c *gin.Context) robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoMniUI(),
		dokumentaTableID: robnoMniDokumentaTableID,
		stavkeTableID:    robnoMniDokumentStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetFaktureDokumentaTableFields(),
		stavkeFields:     h.service.GetMniStavkeTableFields(),
		dokumentaRows:    robnosvc.RobDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobMniStavkeTableRows,
		dokumentaURL:     robnoMniPrikazDokumentaURL,
		stavkeURL:        robnoMniPrikazStavkeURL,
		searchID:         robnoMniPrikazSearchID,
		vrste: func(vrd int) []int {
			vrste, err := h.service.GetVrsteGrupeDokumenta(c.Request.Context(), vrd, robnoMniPrikazVrdOd, robnoMniPrikazVrdDo)
			if err != nil || len(vrste) == 0 {
				return []int{vrd}
			}
			return vrste
		},
		greska: robnoMniGreska,
	}
}

// MniPrikaz fills the tab "Pregled nivelacija" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) MniPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoMniPrikazOpis(c))
}

// MniPrikazDokumenta returns the grid of the documents of the tab "Pregled nivelacija" only (the pagination, the
// search and the sort of the grid).
func (h *RobnoDokumentaUnosHandler) MniPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoMniPrikazOpis(c))
}

// MniPrikazStavke returns the stavke of the document of the id (the click of a row of the documents of the tab
// "Pregled nivelacija").
func (h *RobnoDokumentaUnosHandler) MniPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoMniPrikazOpis(c))
}

// robnoMniButtons builds the buttons of the "Nivelacije maloprodaje" screen for the document of the nalog (rdokID
// 0: a new one), like the ones of the porudžbenica: "Sačuvaj" saves the header and opens the stavke, "Izmeni"
// switches from the stavke back to the header, "Novi dok." a new document, "Briši" deletes the saved one (after its
// confirmation), "Nazad" releases the lock of the nalog and closes the dialog, "Štampaj" prints the document (the
// legacy RPT_NIVELACIJA_MALOPRODAJE). "Sačuvaj" of the stavka saves it, "Odustani" gives up the change of a stavka
// (enabled by "Izmeni" of its row).
func robnoMniButtons(rnalID, rdokID int64, vrd int) domain.RobnoFaktureButtons {
	btns := domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-mni-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoMniSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoMniHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoMniBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoMniHeaderPanelID, robnoMniStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-mni-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoMniHeaderPanelID, robnoMniStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-mni-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoMniNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoMniBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-mni-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoMniPotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoMniInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-mni-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoMniDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta, the group MNI).
		Print: common.SetPrintButton("robno-mni-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		SaveStavka: domain.Button{
			Id:                   "robno-mni-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoMniStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoMniStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoMniBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-mni-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robMniStavkaOdustani",
		},
	}
	// A new document has nothing to delete.
	if rdokID <= 0 {
		btns.Delete = domain.Button{Id: "robno-mni-brisi", LabelText: "Briši", Icon: "delete", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassDeleteButton + " opacity-50 cursor-not-allowed"}
	}
	return btns
}

// robnoMniEkranData is the content of the tab "Nivelacije maloprodaje" (the header, the stavka, the grid of the
// stavke and the buttons) with the message of the screen.
type robnoMniEkranData struct {
	header domain.RobnoMniHeaderView
	stavka domain.RobnoMniStavka
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	poruka string
}

// robnoMniEkran builds the content of the tab "Nivelacije maloprodaje" with the document of the params (a new one
// for an empty broj, see GetRobMniDokument): the strip of the nalog, the header, the stavka, the grid of the stavke
// and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoMniEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoMniEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoMniEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobMniDokument(ctx, params)
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
	ekran.tbl = h.robnoFaktureTable("Stavke nivelacije", robnoMniStavkeTableID, h.service.GetMniStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, ekran.header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobMniStavkeTableRows(stavke)
	// The buttons of every row: "Izmeni" puts the stavka in the form (robMniStavkaUcitaj), "Obriši" asks to confirm
	// its deletion; both send the document with the id of the row.
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, ekran.header.RdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoMniStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robMniStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoMniStavkaPotvrdaURL,
		HxTarget: "#" + robnoMniStavkaInfoID, HxSwap: "innerHTML"}
	ekran.btns = robnoMniButtons(dok.RnalID, ekran.header.RdokID, vrd)
	return ekran, nil
}

// robnoMniBody renders the content of the tab "Nivelacije maloprodaje" (the answer of the save, of a selected or
// new document and of the deletion), with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoMniBody(c *gin.Context, ekran robnoMniEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobMniBody(robnoMniUI(), ekran.header, ekran.stavka, ekran.tbl, ekran.btns, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoMniInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-mni-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-mni-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoMniInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoMniGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func robnoMniGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoMniInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoMniInfoDialogID, message, true)
}

// robnoMniStavkaGreska shows an error of the stavke in the container of the dialogs of the stavke.
func robnoMniStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoMniStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoMniStavkaInfoDialogID, message, true)
}

// otvoriNivelacijuMP opens the "Nivelacije maloprodaje" screen (the legacy rob_Nivelacija_maloprodaje) of the nalog
// as a dialog with its two tabs, with a new document (the next broj of the vrsta in the magacin, NoviBrojRDOK). The
// nalog stays locked while the dialog is open: "Zatvori" and "Nazad" release it.
func (h *RobnoDokumentaUnosHandler) otvoriNivelacijuMP(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoMniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoMniPrikazOpis(c), nil)
	btnClose := domain.Button{
		Id:            "robno-mni-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoMniDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_NivelacijaMP(robnoMniUI(), ekran.header, ekran.stavka, ekran.tbl, dokumentaTbl, dokumentStavkeTbl, ekran.btns, btnClose,
		h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// mniDokument renders the content of the tab "Nivelacije maloprodaje" with the document of the params: the nalog
// (its vrsta, broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) mniDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoMniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoMniEkran(c, dok, params)
	if err != nil {
		robnoMniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoMniBody(c, ekran)
}

// MniDokument reloads the content of the tab "Nivelacije maloprodaje" with the existing document of the id (GET
// /dokument/:id with the rnalid of the screen: the selected broj). A document that cannot be changed (of another
// nalog, posted, fiscalized) gives a new document and the message why (the legacy exit of EDT_DOKUM).
func (h *RobnoDokumentaUnosHandler) MniDokument(c *gin.Context) {
	h.mniDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// MniNovi reloads the content of the tab "Nivelacije maloprodaje" with a new document (GET rnalid, vrd and brdok):
// the typed broj (its document when it exists), else the next broj of the vrsta ("Novi dok.", the legacy
// BTN_NoviDOK).
func (h *RobnoDokumentaUnosHandler) MniNovi(c *gin.Context) {
	h.mniDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: c.Query("brdok")})
}

// SaveRobMni saves the header of the nivelacija maloprodaje ("Sačuvaj" of the header, the legacy BTN_Snimi1): the
// validation (ValidateRobMni, the legacy ValidacijaRDOK), the save (SaveRobMni, SnimiRDOK) and the content of the
// tab with the saved document: the header saved and locked, the stavke open (the legacy opens
// rob_Nivelacija_Maloprodaje_DET). A validation error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobMni(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoMniParams
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
	if fieldErrors := h.service.ValidateRobMni(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if params.RdokID, err = h.service.SaveRobMni(ctx, params); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoMniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = true
	h.robnoMniBody(c, ekran)
}

// PotvrdaBrisanjaMni asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the legacy
// Dialog("Brisanje dokumenta?")): "Da" deletes it (DeleteRobMni), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaMni(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dialog := domain.Dialog{Id: robnoMniPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-mni-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-mni-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoMniBrisiURL, rdokID),
		HxTarget:             "#" + robnoMniBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-mni-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{robnoInterniMsgBrisanjeDokum}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobMni deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy BTN_Brisanje) and
// answers with the content of the tab with a new document of the vrsta (the legacy Reset and NoviBrojRDOK) and the
// message of the deletion. A document that cannot be deleted (posted, fiscalized) keeps the screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobMni(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoMniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobMni(ctx, rdokID)
	if err != nil {
		robnoMniGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoMniGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoMniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoMniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoMniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.poruka = robnoInterniMsgObrisano
	h.robnoMniBody(c, ekran)
}

// MniStavka returns a saved stavka (GET rdokid, id) as JSON for the form of the stavke ("Izmeni" of its row): see
// GetRobMniStavkaForma.
func (h *RobnoDokumentaUnosHandler) MniStavka(c *gin.Context) {
	stavka, found, err := h.service.GetRobMniStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
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

// SaveRobMniStavka saves a stavka of the nivelacija maloprodaje (POST of the fields of the stavke, the legacy
// BTN_Snimi of rob_Nivelacija_Maloprodaje_DET): the validation (ValidateRobMniStavka, the legacy Validacija), the
// save (SaveRobMniStavka: CreateRpro or EditRpro) and the content of the tab with the document; "Unesi procenat"
// with the procenat and "Promeni tarifu" stay for the next stavka. A validation error is the JSON of the field
// errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobMniStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobMniStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobMniStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobMniStavka(ctx, params); err != nil {
		robnoMniStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	stanje := domain.RobnoMniStavka{UnesiProcenat: params.UnesiProcenat, PromeniTarifu: params.PromeniTarifu}
	if params.UnesiProcenat {
		stanje.Procenat = params.Procenat
	}
	h.robnoMniStavkeOdgovor(c, params.RdokID, stanje)
}

// PotvrdaBrisanjaMniStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the
// legacy YesNo "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it (DeleteRobMniStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaMniStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rproID := robnoFaktureRproID(c)
	dialog := domain.Dialog{Id: robnoMniStavkaPotvrdaDialogID, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-mni-stavka-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-mni-stavka-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoMniStavkaBrisiURL, rproID),
		HxTarget:             "#" + robnoMniBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-mni-stavka-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message("Brisanje podataka? Da li ste sigurni?")}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer)
}

// DeleteRobMniStavka deletes a stavka of the nivelacija maloprodaje (DELETE …/stavka/brisi/:id, the legacy
// BrisiRPRO) and answers with the content of the tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobMniStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobMniStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoMniStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoMniStavkeOdgovor(c, rdokID, domain.RobnoMniStavka{})
}

// robnoMniStavkeOdgovor renders the content of the tab "Nivelacije maloprodaje" with the document of the stavke (its
// header saved, the stavke open): the answer of the save and of the deletion of a stavka. stanje is the state of the
// form kept for the next stavka (the checkboxes and the procenat).
func (h *RobnoDokumentaUnosHandler) robnoMniStavkeOdgovor(c *gin.Context, rdokID int64, stanje domain.RobnoMniStavka) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoMniGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoMniEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoMniGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.stavka.UnesiProcenat, ekran.stavka.Procenat, ekran.stavka.PromeniTarifu = stanje.UnesiProcenat, stanje.Procenat, stanje.PromeniTarifu
	ekran.header.Snimljen = true
	h.robnoMniBody(c, ekran)
}

// robnoMniRoutes registers the routes of the "Nivelacije maloprodaje" screen.
func (h *RobnoDokumentaUnosHandler) robnoMniRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/nivelacije-mp/dokument/:id", h.MniDokument)
	r.GET("/api/robno-dokumenta/nivelacije-mp/novi", h.MniNovi)
	r.POST("/api/robno-dokumenta/nivelacije-mp/save", h.SaveRobMni)
	r.GET("/api/robno-dokumenta/nivelacije-mp/brisi/potvrda", h.PotvrdaBrisanjaMni)
	r.DELETE("/api/robno-dokumenta/nivelacije-mp/brisi/:id", h.DeleteRobMni)
	// The stopa of a tarifa: the one of the stavke of the faktura usluga (the legacy exits of EDT_po and EDT_po1,
	// VratiPorez2 on the datum of the document).
	r.GET("/api/robno-dokumenta/nivelacije-mp/artikal/stopa", h.UslugeStopa)
	r.GET("/api/robno-dokumenta/nivelacije-mp/stavka", h.MniStavka)
	r.POST("/api/robno-dokumenta/nivelacije-mp/stavka/save", h.SaveRobMniStavka)
	r.GET("/api/robno-dokumenta/nivelacije-mp/stavka/potvrda", h.PotvrdaBrisanjaMniStavke)
	r.DELETE("/api/robno-dokumenta/nivelacije-mp/stavka/brisi/:id", h.DeleteRobMniStavka)
	r.GET("/api/robno-dokumenta/nivelacije-mp/prikaz", h.MniPrikaz)
	r.GET("/api/robno-dokumenta/nivelacije-mp/prikaz/dokumenta", h.MniPrikazDokumenta)
	r.GET("/api/robno-dokumenta/nivelacije-mp/prikaz/stavke/:id", h.MniPrikazStavke)
}
