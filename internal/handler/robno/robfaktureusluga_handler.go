package robno

import (
	"fmt"
	"net/http"
	"strings"

	tmpl "helia/frontend/templates"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/internal/common"
	"helia/internal/domain"
	commonsvc "helia/internal/service/common"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// The "Fakture usluga" screen (the legacy rob_Fakture_Usluga and rob_FAKTURE_Usluga_DET2, the vrste dokumenata
// of the groups PRE, FUR, FCO, RCO and KPF): its urls and the ids of its elements (the templates get them by
// robnoUslugeUI). The header and the second tab are the ones of the faktura.
const (
	robnoUslugeURL = robnoDokumentaURLPrefix + "/usluge"
	// robnoUslugeDokumentURL + "/:id" reloads the content of the screen with an existing document (the
	// selected broj), robnoUslugeNoviURL with a new one ("Novi dok.", a typed broj).
	robnoUslugeDokumentURL = robnoUslugeURL + "/dokument"
	robnoUslugeNoviURL     = robnoUslugeURL + "/novi"
	// robnoUslugeSaveURL saves the header (POST), robnoUslugePotvrdaURL asks to confirm the deletion of the
	// document and robnoUslugeBrisiURL + "/:id" deletes it.
	robnoUslugeSaveURL    = robnoUslugeURL + "/save"
	robnoUslugeBrisiURL   = robnoUslugeURL + "/brisi"
	robnoUslugePotvrdaURL = robnoUslugeBrisiURL + "/potvrda"
	// The tab "Pregled faktura": robnoUslugePrikazURL fills it when it is shown, robnoUslugePrikazDokumentaURL
	// reloads its grid, robnoUslugePrikazStavkeURL + "/:id" the stavke of the selected document.
	robnoUslugePrikazURL          = robnoUslugeURL + "/prikaz"
	robnoUslugePrikazDokumentaURL = robnoUslugePrikazURL + "/dokumenta"
	robnoUslugePrikazStavkeURL    = robnoUslugePrikazURL + "/stavke"
	// The stavke (the legacy rob_FAKTURE_Usluga_DET2): robnoUslugeArtikalURL returns the artikal of the stavka
	// (GET rdokid, sifra) and robnoUslugeArtikalURL + "/stopa" the stopa of the tarifa (GET rdokid, tarifa),
	// robnoUslugeStavkaURL a saved stavka for the form (GET rdokid, id: "Izmeni" of its row),
	// robnoUslugeStavkaSaveURL saves the stavka (POST), robnoUslugeStavkaPotvrdaURL asks to confirm the
	// deletion of a stavka (GET rdokid, id: "Obriši" of its row) and robnoUslugeStavkaBrisiURL + "/:id" deletes it.
	robnoUslugeArtikalURL       = robnoUslugeURL + "/artikal"
	robnoUslugeStavkaURL        = robnoUslugeURL + "/stavka"
	robnoUslugeStavkaSaveURL    = robnoUslugeStavkaURL + "/save"
	robnoUslugeStavkaPotvrdaURL = robnoUslugeStavkaURL + "/potvrda"
	robnoUslugeStavkaBrisiURL   = robnoUslugeStavkaURL + "/brisi"

	robnoUslugeDialogID              = "robno-usluge-dialog"
	robnoUslugeBodyID                = "robno-usluge-body"
	robnoUslugeInfoID                = "robno-usluge-info"
	robnoUslugeInfoDialogID          = "robno-usluge-info-dialog"
	robnoUslugePotvrdaDialogID       = "robno-usluge-brisanje"
	robnoUslugeStavkaInfoID          = "robno-usluge-stavka-info"
	robnoUslugeStavkaInfoDialogID    = "robno-usluge-stavka-info-dialog"
	robnoUslugeStavkaPotvrdaDialogID = "robno-usluge-stavka-brisanje"
	robnoUslugeHeaderPanelID         = "robno-usluge-header-panel"
	robnoUslugeStavkePanelID         = "robno-usluge-stavke-panel"
	robnoUslugePrikazID              = "robno-usluge-prikaz"
	robnoUslugePrikazStavkeID        = "robno-usluge-prikaz-stavke"
	robnoUslugePrikazRdokID          = "robno-usluge-prikaz-rdokid"
	robnoUslugePrikazSearchID        = "robno-usluge-prikaz-search"
	robnoUslugePrikazStampaID        = "robno-usluge-prikaz-stampa"
	robnoUslugeStavkeTableID         = "robno-usluge-stavke-table"
	robnoUslugeAvansiTableID         = "robno-usluge-avansi-table"
	robnoUslugeDokumentaTableID      = "robno-usluge-dokumenta-table"
	robnoUslugeDokumentStavkeTableID = "robno-usluge-dokument-stavke-table"
	robnoUslugeMsgObrisano           = "Uspešno obrisani podaci..."
	robnoUslugeMsgBrisanjeDokumenta  = "Brisanje dokumenta?"
	robnoUslugeMsgBrisanjeStavke     = "Brisanje podataka? Da li ste sigurni?"
	robnoUslugeMsgStavkaNePostoji    = "Stavka koju birate za izmenu ne postoji."
	robnoUslugePrikazVrdOd           = 100
	robnoUslugePrikazVrdDo           = 299
	robnoUslugeInfoTitle             = "Info"
	robnoUslugeBrisanjeTitle         = "Brisanje"
	robnoUslugeStavkeNaslovGrida     = "Stavke"
	robnoUslugeDokumentaNaslovGrida  = "Dokumenta"
	robnoUslugeDokStavkeNaslovGrida  = "Stavke dokumenta"
	robnoUslugeAvansiNaslovGrida     = "Avansi"
	robnoUslugeDefaultNaslovEkrana   = "Fakture tekst"
)

// robnoUslugeUI returns the ids and the urls of the "Fakture usluga" screen.
func robnoUslugeUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID:      robnoUslugeHeaderPanelID,
		StavkePanelID:      robnoUslugeStavkePanelID,
		ContentID:          "#" + robnoDokumentaContentID,
		DialogID:           robnoUslugeDialogID,
		ArtikalSearchURL:   robnoFaktureArtikalSearch,
		PartnerSearchURL:   robnoFakturePartnerSearch,
		DokumentSearchURL:  robnoFaktureDokumentSearch,
		DokumentURL:        robnoUslugeDokumentURL,
		NoviURL:            robnoUslugeNoviURL,
		ArtikalURL:         robnoUslugeArtikalURL,
		BodyID:             robnoUslugeBodyID,
		InfoID:             robnoUslugeInfoID,
		StavkaInfoID:       robnoUslugeStavkaInfoID,
		PrikazID:           robnoUslugePrikazID,
		PrikazStavkeID:     robnoUslugePrikazStavkeID,
		PrikazRdokID:       robnoUslugePrikazRdokID,
		PrikazStampaID:     robnoUslugePrikazStampaID,
		PrikazURL:          robnoUslugePrikazURL,
		PrikazStavkeNaslov: robnoUslugeDokStavkeNaslovGrida,
	}
}

// robnoUslugePrikazOpis is the tab "Pregled faktura": the documents of the vrste of the group of the vrsta of
// the screen between 100 and 299 (the legacy QRY_RDOK_TIP) in the magacin of the nalog and the stavke of the
// selected one.
func (h *RobnoDokumentaUnosHandler) robnoUslugePrikazOpis(c *gin.Context) robnoPrikaz {
	return robnoPrikaz{
		ui:               robnoUslugeUI(),
		dokumentaTableID: robnoUslugeDokumentaTableID,
		stavkeTableID:    robnoUslugeDokumentStavkeTableID,
		dokumentaNaslov:  robnoUslugeDokumentaNaslovGrida,
		stavkeNaslov:     robnoUslugeDokStavkeNaslovGrida,
		dokumentaFields:  h.service.GetFaktureDokumentaTableFields(),
		stavkeFields:     h.service.GetFaktureDokumentStavkeTableFields(),
		dokumentaRows:    robnosvc.RobDokumentaPrikazTableRows,
		stavkeRows:       robnosvc.RobStavkePrikazTableRows,
		dokumentaURL:     robnoUslugePrikazDokumentaURL,
		stavkeURL:        robnoUslugePrikazStavkeURL,
		searchID:         robnoUslugePrikazSearchID,
		vrste: func(vrd int) []int {
			vrste, err := h.service.GetVrsteGrupeDokumenta(c.Request.Context(), vrd, robnoUslugePrikazVrdOd, robnoUslugePrikazVrdDo)
			if err != nil || len(vrste) == 0 {
				return []int{vrd}
			}
			return vrste
		},
		greska: robnoUslugeGreska,
	}
}

// UslugePrikaz fills the tab "Pregled faktura" when it is shown (GET rnalid and vrd).
func (h *RobnoDokumentaUnosHandler) UslugePrikaz(c *gin.Context) {
	h.robnoPrikazTab(c, h.robnoUslugePrikazOpis(c))
}

// UslugePrikazDokumenta returns the grid of the documents of the tab "Pregled faktura" only (the pagination,
// the search and the sort of the grid).
func (h *RobnoDokumentaUnosHandler) UslugePrikazDokumenta(c *gin.Context) {
	h.robnoPrikazGrid(c, h.robnoUslugePrikazOpis(c))
}

// UslugePrikazStavke returns the stavke of the document of the id (the click of a row of the documents of the
// tab "Pregled faktura").
func (h *RobnoDokumentaUnosHandler) UslugePrikazStavke(c *gin.Context) {
	h.robnoPrikazStavke(c, h.robnoUslugePrikazOpis(c))
}

// robnoUslugeButtons builds the buttons of the "Fakture usluga" screen for the document of the nalog (rdokID
// 0: a new one), like the ones of the faktura: "Sačuvaj" saves the header and opens the stavke, "Izmeni"
// switches from the stavke back to the header, "Novi dok." a new document, "Briši" deletes the saved one
// (after its confirmation), "Nazad" releases the lock of the nalog and closes the dialog, "Štampaj" prints the
// document; "Sačuvaj" of the stavka saves it (the legacy "Snimi" of rob_FAKTURE_Usluga_DET2), "Odustani"
// gives up the change of a stavka (enabled by "Izmeni" of its row).
func robnoUslugeButtons(rnalID, rdokID int64, vrd int) domain.RobnoFaktureButtons {
	btns := domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "robno-usluge-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoUslugeSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoUslugeHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoUslugeBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{robnoUslugeHeaderPanelID, robnoUslugeStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "robno-usluge-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{robnoUslugeHeaderPanelID, robnoUslugeStavkePanelID},
		},
		New: domain.Button{
			Id:            "robno-usluge-novi",
			LabelText:     "Novi dok.",
			Icon:          "add",
			BtnClass:      common.ClassNewButton,
			HxActionURL:   fmt.Sprintf("%s?rnalid=%d&vrd=%d", robnoUslugeNoviURL, rnalID, vrd),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoUslugeBodyID,
			HxSwap:        "outerHTML",
		},
		Delete: domain.Button{
			Id:            "robno-usluge-brisi",
			LabelText:     "Briši",
			Icon:          "delete",
			BtnClass:      common.ClassDeleteButton,
			HxActionURL:   fmt.Sprintf("%s?rdokid=%d", robnoUslugePotvrdaURL, rdokID),
			HxRequestType: "GET",
			HxTarget:      "#" + robnoUslugeInfoID,
			HxSwap:        "innerHTML",
		},
		Back: domain.Button{
			Id:                   "robno-usluge-nazad",
			LabelText:            "Nazad",
			Icon:                 "back",
			BtnClass:             common.ClassCloseButton,
			HxActionURL:          robnoFaktureUnlockURL(rnalID),
			HxRequestType:        "POST",
			HxSwap:               "none",
			HxOnAfterRequest:     "closeDialog",
			HxOnAfterRequestArgs: []any{robnoUslugeDialogID},
		},
		// "Štampaj": the print of the document of the form (the print of its vrsta dokumenta), with the rdokid
		// of the form (the hidden rdokid of the stavke, id stavkardokid), like the faktura.
		Print: common.SetPrintButton("robno-usluge-stampa", "Štampaj", "stampa", robnoDokumentaURLPregledStampaPrint, "GET", true,
			common.ClassPrintButton, "stavkardokid"),
		SaveStavka: domain.Button{
			Id:                   "robno-usluge-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          robnoUslugeStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + robnoUslugeStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + robnoUslugeBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "robno-usluge-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "robUslugeStavkaOdustani",
		},
	}
	// A new document has nothing to delete.
	if rdokID <= 0 {
		btns.Delete.IsDisabled = true
		btns.Delete.BtnClass += " opacity-50 cursor-not-allowed"
		btns.Delete.HxActionURL = ""
	}
	return btns
}

// robnoUslugeEkranData is the content of the first tab (the header, the stavka, the grids of the stavke and of
// the avansi, the combos and the buttons) with the message and the title of the screen.
type robnoUslugeEkranData struct {
	header               domain.RobnoFaktureHeaderView
	stavka               domain.RobnoUslugeStavka
	tbl, avansiTbl       domain.TableData
	vrstaDokumentaValues []domain.ComboItem
	ziroRacuni           []domain.ComboItem
	btns                 domain.RobnoFaktureButtons
	poruka, naslov       string
}

// robnoUslugeEkran builds the content of the first tab with the document of the params (a new one for an
// empty broj, see GetRobUslugeDokument): the strip of the nalog, the header, the stavka, the grids of the
// stavke and of the avansi and the buttons.
func (h *RobnoDokumentaUnosHandler) robnoUslugeEkran(c *gin.Context, dok robnoDokumentaUnosDokument, params domain.RobnoFaktureDokumentParams) (robnoUslugeEkranData, error) {
	ctx := c.Request.Context()
	ekran := robnoUslugeEkranData{}
	rnal, magacin, err := h.service.GetByTipdokNalogMagacin(ctx, dok.Tipdok, dok.Nalog, dok.MagaciniID)
	if err != nil {
		return ekran, err
	}
	if ekran.vrstaDokumentaValues, err = h.service.GetVrstaDokumentaComboValues(ctx, commonsvc.WithIn("modul", h.cfg.GrupeDokumenata.GetDokRob()...)); err != nil {
		return ekran, err
	}
	dokument, err := h.service.GetRobUslugeDokument(ctx, params)
	if err != nil {
		return ekran, err
	}
	ekran.header, ekran.stavka, ekran.ziroRacuni, ekran.poruka, ekran.naslov = dokument.Header, dokument.Stavka, dokument.ZiroRacuni,
		dokument.Poruka, dokument.Naslov
	robnoNalogStrip(&ekran.header, rnal, magacin)
	// The vrsta of a document opened by its id is the one of the document.
	vrd := dok.Vrd
	if v := common.StringToInt(ekran.header.Vrd); v > 0 {
		vrd = v
	}

	// The grids: the stavke (the buttons of every row: "Izmeni" puts the stavka in the form, "Obriši" asks to
	// confirm its deletion; both send the document with the id of the row) and the avansi of the document.
	ekran.tbl = h.robnoFaktureTable(robnoUslugeStavkeNaslovGrida, robnoUslugeStavkeTableID, h.service.GetUslugeStavkeTableFields())
	stavke, err := h.service.GetRobStavkeDokumenta(ctx, ekran.header.RdokID)
	if err != nil {
		return ekran, err
	}
	ekran.tbl.Rows = robnosvc.RobUslugeStavkeTableRows(stavke)
	ekran.tbl.ShowActions = len(stavke) > 0
	ekran.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, ekran.header.RdokID)
	ekran.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: robnoUslugeStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "robUslugeStavkaUcitaj(event)"}
	ekran.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: robnoUslugeStavkaPotvrdaURL,
		HxTarget: "#" + robnoUslugeStavkaInfoID, HxSwap: "innerHTML"}
	ekran.avansiTbl = h.robnoFaktureTable(robnoUslugeAvansiNaslovGrida, robnoUslugeAvansiTableID, h.service.GetFaktureAvansiTableFields())
	ekran.avansiTbl.Rows = dokument.Avansi
	ekran.btns = robnoUslugeButtons(dok.RnalID, ekran.header.RdokID, vrd)
	return ekran, nil
}

// robnoUslugeBody renders the content of the first tab (the answer of the save, of a selected or new document,
// of the stavke and of the deletion), with the message of the screen in its dialog.
func (h *RobnoDokumentaUnosHandler) robnoUslugeBody(c *gin.Context, ekran robnoUslugeEkranData) {
	ctx := c.Request.Context()
	if err := tmpl_robno.RobUslugeBody(robnoUslugeUI(), ekran.header, ekran.stavka, ekran.tbl, ekran.avansiTbl, ekran.vrstaDokumentaValues,
		robnosvc.RobSistemPdvCombo(), ekran.ziroRacuni, ekran.btns, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if ekran.poruka != "" {
		dialog := domain.Dialog{Id: robnoUslugeInfoDialogID, Title: robnoUslugeInfoTitle}
		btnClose := domain.Button{Id: "robno-usluge-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "robno-usluge-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(robnoUslugeInfoID, ekran.poruka, dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// robnoUslugeGreska shows an error in the dialog of the messages of the screen instead of the target of the
// request.
func robnoUslugeGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoUslugeInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoUslugeInfoDialogID, message, true)
}

// robnoUslugeStavkaGreska shows an error of the stavke in the container of the dialogs of the stavke.
func robnoUslugeStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+robnoUslugeStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, robnoUslugeStavkaInfoDialogID, message, true)
}

// otvoriFakturuUsluga opens the "Fakture usluga" screen (the legacy rob_Fakture_Usluga) of the nalog as a
// dialog with its two tabs, with a new document (the next broj of the vrsta in the magacin, NoviBrojRDOK). The
// nalog stays locked while the dialog is open: "Zatvori" and "Nazad" release it.
func (h *RobnoDokumentaUnosHandler) otvoriFakturuUsluga(c *gin.Context, dok robnoDokumentaUnosDokument) {
	ctx := c.Request.Context()
	ekran, err := h.robnoUslugeEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	// The grids of the tab "Pregled faktura" (filled when the tab is shown: UslugePrikaz).
	dokumentaTbl, dokumentStavkeTbl := h.robnoPrikazTabele(h.robnoUslugePrikazOpis(c), nil)
	btnClose := domain.Button{
		Id:            "robno-usluge-dialog-close",
		IsVisible:     true,
		IdDialog:      robnoUslugeDialogID,
		BtnClass:      common.ClassDialogCloseButton,
		HxActionURL:   robnoFaktureUnlockURL(dok.RnalID),
		HxRequestType: "POST",
	}
	naslov := ekran.naslov
	if naslov == "" {
		naslov = robnoUslugeDefaultNaslovEkrana
	}
	if err := tmpl_robno.Rob_FaktureUsluga(robnoUslugeUI(), naslov, ekran.header, ekran.stavka, ekran.tbl, ekran.avansiTbl, dokumentaTbl,
		dokumentStavkeTbl, ekran.vrstaDokumentaValues, robnosvc.RobSistemPdvCombo(), ekran.ziroRacuni, ekran.btns, btnClose, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, robnoDokumentaDokumentInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// uslugeDokument renders the content of the first tab with the document of the params: the nalog (its vrsta,
// broj and magacin) is the one of the rnalid of the query.
func (h *RobnoDokumentaUnosHandler) uslugeDokument(c *gin.Context, params domain.RobnoFaktureDokumentParams) {
	dok, err := h.robnoFaktureNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rnalid"))), params.Vrd)
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.RnalID, params.MagaciniID = dok.RnalID, dok.MagaciniID
	ekran, err := h.robnoUslugeEkran(c, dok, params)
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	h.robnoUslugeBody(c, ekran)
}

// UslugeDokument reloads the first tab with the existing document of the id (GET /dokument/:id with the rnalid
// of the screen: the selected broj). A document that cannot be changed (of another nalog, posted, fiscalized,
// sent to the SEF) gives a new document and the message why.
func (h *RobnoDokumentaUnosHandler) UslugeDokument(c *gin.Context) {
	h.uslugeDokument(c, domain.RobnoFaktureDokumentParams{RdokID: int64(common.StringToInt(c.Param("id")))})
}

// UslugeNovi reloads the first tab with a new document (GET rnalid, vrd and brdok): the typed broj (its
// document when it exists), else the next broj of the vrsta ("Novi dok.").
func (h *RobnoDokumentaUnosHandler) UslugeNovi(c *gin.Context) {
	h.uslugeDokument(c, domain.RobnoFaktureDokumentParams{Vrd: common.StringToInt(c.Query("vrd")), Dokum: strings.TrimSpace(c.Query("brdok"))})
}

// SaveRobUsluge saves the header of the faktura usluga ("Sačuvaj" of the header, POST of the fields of the
// header), like the legacy "Sačuvaj" of rob_Fakture_Usluga: the validation (ValidateRobUsluge, the legacy
// ValidacijaRDOK), the save (SaveRobUsluge: SnimiRDOK, PreracunRATA, ObracunFAKT) and the first tab with the
// saved document: the header saved and locked, the stavke open (the legacy opens rob_FAKTURE_Usluga_DET2). A
// validation error is the JSON of the field errors (status 422, the content is not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobUsluge(c *gin.Context) {
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
	dok, err := h.robnoFaktureNalog(ctx, params.RnalID, common.StringToInt(params.Vrd))
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	params.MagaciniID = dok.MagaciniID
	if fieldErrors := h.service.ValidateRobUsluge(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if params.RdokID, err = h.service.SaveRobUsluge(ctx, params); err != nil {
		robnoUslugeGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	ekran, err := h.robnoUslugeEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: params.RdokID, Dokum: params.BrojDokum})
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.header.Snimljen = true
	h.robnoUslugeBody(c, ekran)
}

// PotvrdaBrisanjaUsluge asks to confirm the deletion of the document of the screen (GET rdokid: "Briši", the
// legacy Dialog("Brisanje dokumenta?")): "Da" deletes it (DeleteRobUsluge), "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaUsluge(c *gin.Context) {
	h.robnoUslugePotvrda(c, robnoUslugePotvrdaDialogID, "robno-usluge-brisanje", robnoUslugeMsgBrisanjeDokumenta,
		fmt.Sprintf("%s/%d", robnoUslugeBrisiURL, int64(common.StringToInt(c.Query("rdokid")))))
}

// PotvrdaBrisanjaUslugeStavke asks to confirm the deletion of a stavka ("Obriši" of its row in the grid, GET
// rdokid and id, the legacy YesNo of BrisiRPRO): "Da" deletes it (DeleteRobUslugeStavka).
func (h *RobnoDokumentaUnosHandler) PotvrdaBrisanjaUslugeStavke(c *gin.Context) {
	h.robnoUslugePotvrda(c, robnoUslugeStavkaPotvrdaDialogID, "robno-usluge-stavka-brisanje", robnoUslugeMsgBrisanjeStavke,
		fmt.Sprintf("%s/%d", robnoUslugeStavkaBrisiURL, robnoFaktureRproID(c)))
}

// robnoUslugePotvrda renders a confirmation of a deletion: "Da" sends the DELETE of the url and gets the first
// tab, "Ne" closes the dialog.
func (h *RobnoDokumentaUnosHandler) robnoUslugePotvrda(c *gin.Context, dialogID, idPrefix, poruka, url string) {
	dialog := domain.Dialog{Id: dialogID, Title: robnoUslugeBrisanjeTitle}
	btnClose := domain.Button{Id: idPrefix + "-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   idPrefix + "-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        "DELETE",
		HxActionURL:          url,
		HxTarget:             "#" + robnoUslugeBodyID,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: idPrefix + "-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message(poruka)}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// DeleteRobUsluge deletes the document of the id with its stavke (DELETE /brisi/:id, the legacy "Briši" of
// rob_Fakture_Usluga, DeleteRobFakturaUsluge) and answers with the first tab with a new document of the vrsta
// and the message of the deletion. A document that cannot be deleted (sent to the SEF or to the Pojedinačna
// evidencija PDV, fiscalized, posted) keeps the screen and shows why.
func (h *RobnoDokumentaUnosHandler) DeleteRobUsluge(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	row, found, err := h.service.GetRobniDokumentByID(ctx, rdokID)
	if err == nil && !found {
		err = fmt.Errorf("dokument ne postoji")
	}
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	poruka, err := h.service.DeleteRobFakturaUsluge(ctx, rdokID)
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		robnoUslugeGreska(c, poruka)
		return
	}
	dok, err := h.robnoFaktureNalog(ctx, row.RnalID, row.Vrd)
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran, err := h.robnoUslugeEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd})
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	ekran.poruka = robnoUslugeMsgObrisano
	h.robnoUslugeBody(c, ekran)
}

// UslugeArtikal returns the artikal of a stavka (GET rdokid, sifra - the šifra or the barkod) as JSON for the
// form of the stavke (the legacy exit of EDT_SIFRA): see GetRobUslugaArtikal; found false when the artikal does
// not exist.
func (h *RobnoDokumentaUnosHandler) UslugeArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetRobUslugaArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "artikal": artikal})
}

// UslugeStopa returns the stopa of the tarifa of a stavka (GET rdokid, tarifa) with its poreska kategorija as
// JSON (the legacy exit of EDT_po): see GetRobUslugaStopa.
func (h *RobnoDokumentaUnosHandler) UslugeStopa(c *gin.Context) {
	stopa, err := h.service.GetRobUslugaStopa(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), common.StringToInt(c.Query("tarifa")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stopa)
}

// UslugeStavka returns a saved stavka (GET rdokid, id) as JSON for the form of the stavke (the legacy
// "Izmene"): see GetRobUslugaStavkaForma.
func (h *RobnoDokumentaUnosHandler) UslugeStavka(c *gin.Context) {
	stavka, found, err := h.service.GetRobUslugaStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), robnoFaktureRproID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": robnoUslugeMsgStavkaNePostoji})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SaveRobUslugeStavka saves a stavka of the faktura usluga (POST of the fields of the stavke, the legacy "Snimi"
// of rob_FAKTURE_Usluga_DET2): the validation (ValidateRobUslugaStavka, the legacy Validacija), the save
// (SaveRobUslugaStavka: CreateRpro or UpdateRPRO) and the first tab with the document; the konto of the stavka
// stays for the next one. A validation error is the JSON of the field errors (status 422, not swapped).
func (h *RobnoDokumentaUnosHandler) SaveRobUslugeStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.RobUslugaStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateRobUslugaStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if _, err := h.service.SaveRobUslugaStavka(ctx, params); err != nil {
		robnoUslugeStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.robnoUslugeStavkeOdgovor(c, params.RdokID, params.Konto)
}

// DeleteRobUslugeStavka deletes a stavka of the faktura usluga (DELETE …/stavka/brisi/:id, the legacy
// BrisiDOKUSLDET and BrisiRPRO; its document is read from the stavka) and answers with the first tab with the
// document.
func (h *RobnoDokumentaUnosHandler) DeleteRobUslugeStavka(c *gin.Context) {
	rdokID, err := h.service.DeleteRobUslugaStavka(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		robnoUslugeStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.robnoUslugeStavkeOdgovor(c, rdokID, "")
}

// robnoUslugeStavkeOdgovor renders the first tab with the document of the stavke (its header saved, the stavke
// open): the answer of the save and of the deletion of a stavka. A konto keeps the konto of the form for the
// next stavka.
func (h *RobnoDokumentaUnosHandler) robnoUslugeStavkeOdgovor(c *gin.Context, rdokID int64, konto string) {
	ctx := c.Request.Context()
	k, found, err := h.service.GetRobStavkeKontekst(ctx, rdokID)
	if err != nil || !found {
		robnoUslugeGreska(c, common.ErrMsgReadData)
		return
	}
	dok := robnoDokumentaUnosDokument{RnalID: k.RnalID, MagaciniID: k.MagaciniID, Vrd: k.Vrd, Tipdok: k.Tipdok, Nalog: k.Nalog}
	ekran, err := h.robnoUslugeEkran(c, dok, domain.RobnoFaktureDokumentParams{RnalID: dok.RnalID, MagaciniID: dok.MagaciniID, Vrd: dok.Vrd,
		RdokID: rdokID})
	if err != nil {
		robnoUslugeGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if konto = strings.TrimSpace(konto); konto != "" {
		ekran.stavka.Konto = konto
	}
	ekran.header.Snimljen = true
	h.robnoUslugeBody(c, ekran)
}

// robnoUslugeRoutes registers the routes of the "Fakture usluga" screen.
func (h *RobnoDokumentaUnosHandler) robnoUslugeRoutes(r *gin.Engine) {
	r.GET("/api/robno-dokumenta/usluge/dokument/:id", h.UslugeDokument)
	r.GET("/api/robno-dokumenta/usluge/novi", h.UslugeNovi)
	r.POST("/api/robno-dokumenta/usluge/save", h.SaveRobUsluge)
	r.GET("/api/robno-dokumenta/usluge/brisi/potvrda", h.PotvrdaBrisanjaUsluge)
	r.DELETE("/api/robno-dokumenta/usluge/brisi/:id", h.DeleteRobUsluge)
	r.GET("/api/robno-dokumenta/usluge/artikal", h.UslugeArtikal)
	r.GET("/api/robno-dokumenta/usluge/artikal/stopa", h.UslugeStopa)
	r.GET("/api/robno-dokumenta/usluge/stavka", h.UslugeStavka)
	r.POST("/api/robno-dokumenta/usluge/stavka/save", h.SaveRobUslugeStavka)
	r.GET("/api/robno-dokumenta/usluge/stavka/potvrda", h.PotvrdaBrisanjaUslugeStavke)
	r.DELETE("/api/robno-dokumenta/usluge/stavka/brisi/:id", h.DeleteRobUslugeStavka)
	r.GET("/api/robno-dokumenta/usluge/prikaz", h.UslugePrikaz)
	r.GET("/api/robno-dokumenta/usluge/prikaz/dokumenta", h.UslugePrikazDokumenta)
	r.GET("/api/robno-dokumenta/usluge/prikaz/stavke/:id", h.UslugePrikazStavke)
}
