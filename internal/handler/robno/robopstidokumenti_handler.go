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

// The "Opšti dokumenti" screen (the legacy rob_opsti_dokumenti_opd, the vrste dokumenata of the groups OPD, POP
// and POT): its urls and the ids of its elements (the templates get them by robnoOpstiUI). The header and the
// stavke are on one screen: one of them is processed at a time.
const (
	robnoOpstiURL = robnoDokumentaURLPrefix + "/opsti"
	// robnoOpstiDokumentURL + "/:id" reloads the content of the screen with an existing document (the selected
	// broj), robnoOpstiNoviURL with a new one ("Novi dok.", a typed broj); robnoOpstiSaveURL saves the header.
	robnoOpstiDokumentURL = robnoOpstiURL + "/dokument"
	robnoOpstiNoviURL     = robnoOpstiURL + "/novi"
	robnoOpstiSaveURL     = robnoOpstiURL + "/save"
	// The tab "Pregled": robnoOpstiPrikazURL fills it when it is shown, robnoOpstiPrikazDokumentaURL reloads
	// its grid, robnoOpstiPrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoOpstiPrikazURL          = robnoOpstiURL + "/prikaz"
	robnoOpstiPrikazDokumentaURL = robnoOpstiPrikazURL + "/dokumenta"
	robnoOpstiPrikazStavkeURL    = robnoOpstiPrikazURL + "/stavke"
	// The stavke: robnoOpstiArtikalURL returns the artikal of the stavka (GET rdokid, sifra, konto, otk, serija,
	// rok), robnoOpstiStavkaSaveURL adds the stavka (POST), robnoOpstiStavkaPotvrdaURL asks to confirm the
	// deletion of a stavka (GET rdokid, id: "Obriši" of its row) and robnoOpstiStavkaBrisiURL + "/:id" deletes it.
	robnoOpstiArtikalURL = robnoOpstiURL + "/artikal"
	// robnoOpstiKontoMagacinaURL returns the first konto of a magacin (GET magaciniid: the konto zaliha where
	// the roba of a prenosnica goes when its magacin is selected).
	robnoOpstiKontoMagacinaURL  = robnoOpstiURL + "/konto-magacina"
	robnoOpstiStavkaSaveURL     = robnoOpstiURL + "/stavka/save"
	robnoOpstiStavkaPotvrdaURL  = robnoOpstiURL + "/stavka/potvrda"
	robnoOpstiStavkaBrisiURL    = robnoOpstiURL + "/stavka/brisi"
	robnoOpstiDialogID          = "robno-opsti-dialog"
	robnoOpstiBodyID            = "robno-opsti-body"
	robnoOpstiInfoID            = "robno-opsti-info"
	robnoOpstiInfoDialogID      = "robno-opsti-info-dialog"
	robnoOpstiStavkaInfoID      = "robno-opsti-stavka-info"
	robnoOpstiStavkaInfoDialog  = "robno-opsti-stavka-info-dialog"
	robnoOpstiStavkaPotvrdaDlg  = "robno-opsti-stavka-brisanje"
	robnoOpstiHeaderPanelID     = "robno-opsti-header-panel"
	robnoOpstiStavkePanelID     = "robno-opsti-stavke-panel"
	robnoOpstiPrikazID          = "robno-opsti-prikaz"
	robnoOpstiPrikazStavkeID    = "robno-opsti-prikaz-stavke"
	robnoOpstiPrikazRdokID      = "robno-opsti-prikaz-rdokid"
	robnoOpstiPrikazSearchID    = "robno-opsti-prikaz-search"
	robnoOpstiPrikazStampaID    = "robno-opsti-prikaz-stampa"
	robnoOpstiStavkeTableID     = "robno-opsti-stavke-table"
	robnoOpstiDokumentaTableID  = "robno-opsti-dokumenta-table"
	robnoOpstiDokStavkeTableID  = "robno-opsti-dokument-stavke-table"
	robnoOpstiMsgBrisanjeStavke = "Brisanje podataka? Da li ste sigurni?"
)

// robnoOpstiUI returns the ids and the urls of the "Opšti dokumenti" screen.
func robnoOpstiUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:      robnoOpstiHeaderPanelID,
		StavkePanelID:      robnoOpstiStavkePanelID,
		ContentID:          "#" + robnoDokumentaContentID,
		DialogID:           robnoOpstiDialogID,
		DokumentURL:        robnoOpstiDokumentURL,
		NoviURL:            robnoOpstiNoviURL,
		ArtikalURL:         robnoOpstiArtikalURL,
		KontoMagacinaURL:   robnoOpstiKontoMagacinaURL,
		BodyID:             robnoOpstiBodyID,
		InfoID:             robnoOpstiInfoID,
		StavkaInfoID:       robnoOpstiStavkaInfoID,
		PrikazID:           robnoOpstiPrikazID,
		PrikazStavkeID:     robnoOpstiPrikazStavkeID,
		PrikazRdokID:       robnoOpstiPrikazRdokID,
		PrikazStampaID:     robnoOpstiPrikazStampaID,
		PrikazURL:          robnoOpstiPrikazURL,
		PrikazStavkeNaslov: "Stavke dokumenta",
	}
}

// robnoOpstiPrikazOpis is the tab "Pregled": the documents of the vrsta of the screen in the magacin of the
// nalog (the legacy QRY_RDOK_GRPDOK; a finansijski or a količinski opšti dokument and a nivelacija: the vrste of
// its group, FIN, KOL or MNV, from 1 to 199) and the stavke of the selected one. A zaduženje sitnog inventara
// (SIV) has the columns of its legacy grids (robnoOpstiPrikazSiv).
func (h *RobnoDokumentaUnosHandler) robnoOpstiPrikazOpis(c *gin.Context) robnoPrikaz {
	p := h.robnoOpstiPrikazOsnovni(c)
	if h.robnoOpstiPrikazGrupa(c) == "SIV" {
		p.dokumentaFields, p.dokumentaRows = h.service.GetSivDokumentaTableFields(), robnosvc.RobSivDokumentaPrikazTableRows
		p.stavkeFields, p.stavkeRows = h.service.GetSivPrikazStavkeTableFields(), robnosvc.RobSivPrikazStavkeTableRows
	}
	return p
}

// robnoOpstiPrikazGrupa is the group of the vrsta of a request of the tab "Pregled" (upper case; "" when it cannot
// be read): the one of the vrd of the query, else of the document of the id (the stavke of a selected document).
func (h *RobnoDokumentaUnosHandler) robnoOpstiPrikazGrupa(c *gin.Context) string {
	ctx := c.Request.Context()
	vrd := common.StringToInt(c.Query("vrd"))
	if vrd <= 0 {
		if row, found, err := h.service.GetRobniDokumentByID(ctx, int64(common.StringToInt(c.Param("id")))); err == nil && found {
			vrd = row.Vrd
		}
	}
	if vrd <= 0 {
		return ""
	}
	grupa, err := h.service.GetGrupaDokumenta(ctx, vrd)
	if err != nil {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(grupa))
}

// robnoOpstiPrikazOsnovni is the tab "Pregled" of the opšti dokumenti with the columns of the fakture.
func (h *RobnoDokumentaUnosHandler) robnoOpstiPrikazOsnovni(c *gin.Context) robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoOpstiUI(),
		dokumentaTableID: robnoOpstiDokumentaTableID,
		stavkeTableID:    robnoOpstiDokStavkeTableID,
		dokumentaNaslov:  "Dokumenta",
		stavkeNaslov:     "Stavke dokumenta",
		dokumentaFields:  h.service.GetFaktureDokumentaTableFields(),
		stavkeFields:     h.service.GetFaktureDokumentStavkeTableFields(),
		dokumentaRows:    robnosvc.RobDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobStavkePrikazTableRows,
		dokumentaURL:     robnoOpstiPrikazDokumentaURL,
		stavkeURL:        robnoOpstiPrikazStavkeURL,
		searchID:         robnoOpstiPrikazSearchID,
		vrste: func(vrd int) []int {
			ctx := c.Request.Context()
			grupa, err := h.service.GetGrupaDokumenta(ctx, vrd)
			if grupa = strings.ToUpper(strings.TrimSpace(grupa)); err == nil && (grupa == "FIN" || grupa == "KOL" || grupa == "MNV") {
				if vrste, err := h.service.GetVrsteGrupeDokumenta(ctx, vrd, 1, 199); err == nil && len(vrste) > 0 {
					return vrste
				}
			}
			return []int{vrd}
		},
		greska: robnoOpstiGreska,
	}
}

// OpstiPrikaz fills the tab "Pregled" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) OpstiPrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoOpstiPrikazOpis(c))
}

// OpstiPrikazDokumenta returns the grid of the documents of the tab "Pregled" only.
func (h *RobnoDokumentaUnosHandler) OpstiPrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoOpstiPrikazOpis(c))
}

// OpstiPrikazStavke returns the stavke of the document of the id (the click of a row of the tab "Pregled").
func (h *RobnoDokumentaUnosHandler) OpstiPrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoOpstiPrikazOpis(c))
}

// robnoOpstiButtons builds the buttons of the "Opšti dokumenti" screen for the document of the nalog: "Sačuvaj"
// saves the header and switches to the stavke, "Izmeni" switches from the stavke back to the header, "Novi dok."
// a new document, "Štampaj" prints the saved document (RPT_STAMPAPOPIS / RPT_OPDSTAMPA of the vrsta), "Nazad"
// releases the lock of the nalog and closes the dialog; "Sačuvaj" of the stavka adds it (the legacy has no change
// of a stavka).
func robnoOpstiButtons(rnalID int64, vrd int) domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-opsti-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoOpstiSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoOpstiHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoOpstiBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoOpstiHeaderPanelID, robnoOpstiStavkePanelID},
		},
		Modify: domain.Button{
			Id:        "robno-opsti-izmeni",
			LabelText: "Izmeni",
			Icon:      "refresh",
			BtnClass:  common.ClassNewButton,
			HxOnClick: "robOpstiIzmeni",
		},
		New: domain.Button{
			Id:            "robno-opsti-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoOpstiNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoOpstiBodyID,
			HxSwap:        "outerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-opsti-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoOpstiDialogID},
		},
		// The prenosnica prints by its checkboxes (naziv artikla, komercijalni opis).
		Print: common.SetPrintButton("robno-opsti-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid,nazivustampi,komercopis"),
		SaveStavka: domain.Button{
			Id:                   "robno-opsti-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoOpstiStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoOpstiStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoOpstiBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
	}
}

// robnoOpstiEkranData is the content of the first tab with the message of the screen.
type robnoOpstiEkranData struct {
	header domain.RobnoOpstiHeaderView
	stavka domain.RobnoOpstiStavka
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	poruka string
}

// robnoOpstiEkran builds the content of the first tab with the document of the params (a new one for an empty
// broj): the strip of the nalog, the header, the stavka, the grid of the stavke and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoOpstiEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoOpstiEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoOpstiEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobOpstiDokument(ctx, params)
	if err != nil {
		return ekran, err
	}
	ekran.header, ekran.stavka, ekran.poruka = dokument.Header, dokument.Stavka, dokument.Poruka
	robnoNalogStrip(&ekran.header.RobnoFaktureHeaderView, rnal, magacin)
	vrd := dok.Vrd
	if v := common.StringToInt(ekran.header.Vrd); v > 0 {
		vrd = v
	}
	// The grid of the stavke: "Obriši" of a row asks to confirm its deletion.
	// A finansijski opšti dokument has the iznos of the stavke only, a količinski one the količina and the lot.
	fields, rows := h.service.GetOpstiStavkeTableFields(), robnosvc.RobOpstiStavkeTableRows
	switch {
	case ekran.header.Fin:
		fields, rows = h.service.GetOpstiFinStavkeTableFields(), robnosvc.RobOpstiFinStavkeTableRows
	case ekran.header.Kol:
		fields, rows = h.service.GetOpstiKolStavkeTableFields(), robnosvc.RobOpstiKolStavkeTableRows
	case ekran.header.Mnv:
		fields, rows = h.service.GetOpstiMnvStavkeTableFields(), robnosvc.RobOpstiMnvStavkeTableRows
	case ekran.header.Niv:
		// The OTK broj, the serija and the rok of a magacin of the tip zaliha 2 only.
		lot := ekran.stavka.Tipzal == 2
		fields, rows = h.service.GetOpstiNivStavkeTableFields(lot), robnosvc.RobNivStavkeTableRows(lot)
	case ekran.header.Pre:
		// A prenosnica: the nabavna cena of the stavke too.
		fields, rows = h.service.GetOpstiPreStavkeTableFields(), robnosvc.RobPreStavkeTableRows
	case ekran.header.Siv:
		// A zaduženje sitnog inventara: the stavke with the nabavna cena and the lot (the legacy TABLE_rpro of
		// rob_ZaduzSitanInventar: the columns of the prenosnica).
		fields, rows = h.service.GetOpstiPreStavkeTableFields(), robnosvc.RobPreStavkeTableRows
	case ekran.header.Otr:
		// A trebovanje: the stavke with the lot (the vrednost of the stavka).
		fields, rows = h.service.GetOpstiOtrStavkeTableFields(), robnosvc.RobOpstiStavkeTableRows
	case ekran.header.Grd:
		// A zaduženje gradilišta: the stavke without the lot (the vrednost of the stavka).
		fields, rows = h.service.GetOpstiGrdStavkeTableFields(), robnosvc.RobGrdStavkeTableRows
	}
	ekran.tbl = h.robnoFaktureTable("Stavke dokumenta", robnoOpstiStavkeTableID, fields)
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, ekran.header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = rows(stavke)
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, ekran.header.RdokID)
	// The legacy has no change of a stavka ("Izmeni" of the stavke is commented out).
	ekran.tbl.BtnUpdate.IsVisible = false
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoOpstiStavkaPotvrdaURL,
		HxTarget: "#" + robnoOpstiStavkaInfoID, HxSwap: "innerHTML"}
	ekran.btns = robnoOpstiButtons(dok.RnalID, vrd)
	return ekran, nil
}

// robnoOpstiBody renders the content of the first tab with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoOpstiBody(c *gin.Context, ekran robnoOpstiEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobOpstiBody(robnoOpstiUI(), ekran.header, ekran.stavka, ekran.tbl, ekran.btns, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoOpstiInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-opsti-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-opsti-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoOpstiInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoOpstiGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func robnoOpstiGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoOpstiInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoOpstiInfoDialogID, message, true)
}

// robnoOpstiStavkaGreska shows an error of the stavke in the container of the dialogs of the stavke.
func robnoOpstiStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoOpstiStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoOpstiStavkaInfoDialog, message, true)
}

// otvoriOpstiDokument opens the "Opšti dokumenti" screen (the legacy rob_opsti_dokumenti_opd) of the nalog as a
// dialog with its two tabs, with a new document (the next broj of the vrsta in the magacin). The nalog stays
// locked while the dialog is open: "Zatvori" and "Nazad" release it.
func (h *RobnoDokumentaUnosHandler) otvoriOpstiDokument(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoOpstiEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoOpstiPrikazOpis(c), nil)
	btnClose := domain.Button{
		Id:            "robno-opsti-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoOpstiDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	if err := tmpl_robno.Rob_OpstiDokumenti(robnoOpstiUI(), ekran.header, ekran.stavka, ekran.tbl, dokumentaTbl, dokumentStavkeTbl, ekran.btns, btnClose,
		h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	// The message of the screen (e.g. the magacin of a nivelacija veleprodaje whose zalihe are not kept at the
	// prodajna cena, the legacy Info of PodaciZAGLAVLJE) over the opened screen.
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoOpstiInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "robno-opsti-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-opsti-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl.DialogOk(ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// opstiDokument renders the first tab with the document of the params: the nalog is the one of the rnalid of the
// query.
func (h *RobnoDokumentaUnosHandler) opstiDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoOpstiGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoOpstiEkran(c, dok, params)
	if err != nil {
		robnoOpstiGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoOpstiBody(c, ekran)
}

// OpstiDokument reloads the first tab with the existing document of the id (GET /dokument/:id with the rnalid of
// the screen). A document that cannot be changed (of another nalog, posted, fiscalized) gives a new document and
// the message why.
func (h *RobnoDokumentaUnosHandler) OpstiDokument(c *gin.Context) {
	h.opstiDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// OpstiNovi reloads the first tab with a new document (GET rnalid, vrd and brdok): the typed broj (its document
// when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) OpstiNovi(c *gin.Context) {
	h.opstiDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: strings.TrimSpace(c.Query("brdok"))})
}

// SaveRobOpsti saves the header of the opšti dokument ("Sačuvaj" of the header, the legacy "Sačuvaj" of
// rob_opsti_dokumenti_opd): the validation (ValidateRobOpsti, the legacy ValidacijaRDOK), the save (SaveRobOpsti,
// SnimiRDOK) and the first tab with the saved document: the header disabled, the stavke enabled. A validation
// error is the JSON of the field errors (status 422, the content is not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobOpsti(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.RobnoOpstiParams
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
	if fieldErrors := h.service.ValidateRobOpsti(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rdokID, err := h.service.SaveRobOpsti(ctx, params)
	if err != nil {
		robnoOpstiGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoOpstiEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd, RdokID: rdokID})
	if err != nil {
		robnoOpstiGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = true
	h.robnoOpstiBody(c, ekran)
}

// OpstiArtikal returns the artikal of a stavka (GET rdokid, sifra, konto, otk, serija, rok) as JSON for the form of
// the stavke (the legacy exit of EDT_SIFRA): see GetRobOpstiArtikal.
func (h *RobnoDokumentaUnosHandler) OpstiArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetRobOpstiArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"),
		c.Query("konto"), c.Query("otk"), c.Query("serija"), c.Query("rok"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// OpstiKontoMagacina returns the first konto of the magacin with its naziv as JSON (GET magaciniid): the konto
// zaliha where the roba of a prenosnica goes when its magacin is selected (the legacy ListSelectPlus of
// Combo_MAGKONTO).
func (h *RobnoDokumentaUnosHandler) OpstiKontoMagacina(c *gin.Context) {
	konto, naziv, err := h.service.RobPreKontoMagacina(c.Request.Context(), common.StringToInt(c.Query("magaciniid")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"konto": konto, "naziv": naziv})
}

// SaveRobOpstiStavka adds a stavka to the opšti dokument (POST of the fields of the stavke, the legacy "Sačuvaj" of
// the stavke): the validation (ValidateRobOpstiStavka, the legacy Validacija), the save (SaveRobOpstiStavka,
// CreateRpro) and the first tab with the document (the stavke enabled; the konto stays for the next stavka). A
// validation error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobOpstiStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobOpstiStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobOpstiStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobOpstiStavka(ctx, params); err != nil {
		robnoOpstiStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoOpstiStavkeOdgovor(c, params.RdokID, params.Konto)
}

// PotvrdaBrisanjaOpstiStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the
// legacy YesNo of BrisiRPRO; a nivelacija veleprodaje first warns that the cena stays, the legacy YesNo of
// BTN_Brisanje): "Da" deletes it (DeleteRobOpstiStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaOpstiStavke(c *gin.Context) {
	poruke := []string{h.translator.Message(robnoOpstiMsgBrisanjeStavke)}
	if k, found, err := h.service.GetRobStavkeKontekst(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid")))); err == nil && found {
		if grupa, err := h.service.GetGrupaDokumenta(c.Request.Context(), k.Vrd); err == nil && strings.EqualFold(strings.TrimSpace(grupa), "NIV") {
			poruke = append([]string{robnosvc.RobNivMsgBrisanje()}, poruke...)
		}
	}
	dialog := domain.Dialog{Id: robnoOpstiStavkaPotvrdaDlg, Title: "Brisanje"}
	btnClose := domain.Button{Id: "robno-opsti-stavka-brisanje-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   "robno-opsti-stavka-brisanje-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          fmt.Sprintf("%s/%d", robnoOpstiStavkaBrisiURL, robnoFaktureRproID(c)),
		HxTarget:             "#" + robnoOpstiBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: "robno-opsti-stavka-brisanje-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm(poruke, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// DeleteRobOpstiStavka deletes a stavka of the opšti dokument (DELETE …/stavka/brisi/:id, the legacy BrisiRPRO; its
// document is read from the stavka) and answers with the first tab with the document.
func (h *RobnoDokumentaUnosHandler) DeleteRobOpstiStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobOpstiStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoOpstiStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoOpstiStavkeOdgovor(c, rdokID, "")
}

// robnoOpstiStavkeOdgovor renders the first tab with the document of the stavke (the stavke enabled): the answer
// of the save and of the deletion of a stavka. A konto keeps the konto of the form for the next stavka.
func (h *RobnoDokumentaUnosHandler) robnoOpstiStavkeOdgovor(c *gin.Context, rdokID int64, konto string) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoOpstiGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoOpstiEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd, RdokID: rdokID})
	if err != nil {
		robnoOpstiGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if konto = strings.ReplaceAll(strings.TrimSpace(konto), " ", ""); konto != "" {
		ekran.stavka.Konto = konto
	}
	ekran.header.Snimljen = true
	h.robnoOpstiBody(c, ekran)
}

// robnoOpstiRoutes registers the routes of the "Opšti dokumenti" screen.
func (h *RobnoDokumentaUnosHandler) robnoOpstiRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/opsti/dokument/:id", h.OpstiDokument)
	r.GET("/api/robno-dokumenta/opsti/novi", h.OpstiNovi)
	r.POST("/api/robno-dokumenta/opsti/save", h.SaveRobOpsti)
	r.GET("/api/robno-dokumenta/opsti/artikal", h.OpstiArtikal)
	r.GET("/api/robno-dokumenta/opsti/konto-magacina", h.OpstiKontoMagacina)
	r.POST("/api/robno-dokumenta/opsti/stavka/save", h.SaveRobOpstiStavka)
	r.GET("/api/robno-dokumenta/opsti/stavka/potvrda", h.PotvrdaBrisanjaOpstiStavke)
	r.DELETE("/api/robno-dokumenta/opsti/stavka/brisi/:id", h.DeleteRobOpstiStavka)
	r.GET("/api/robno-dokumenta/opsti/prikaz", h.OpstiPrikaz)
	r.GET("/api/robno-dokumenta/opsti/prikaz/dokumenta", h.OpstiPrikazDokumenta)
	r.GET("/api/robno-dokumenta/opsti/prikaz/stavke/:id", h.OpstiPrikazStavke)
}
