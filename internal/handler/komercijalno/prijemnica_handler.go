package komercijalno

import (
	"fmt"
	"net/http"

	"helia/config"
	tmpl "helia/frontend/templates"
	tmpl_kom "helia/frontend/templates/komercijalno"
	tmpl_rep_kom "helia/frontend/templates/reports/komercijalno"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	komsvc "helia/internal/service/komercijalno"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// The "Prijemnica materijala po porudžbenici" screen of the komercijalno (the legacy pr_PrijemnicaSIROVINE, opened from
// the menu): its urls and the ids of its elements (the templates get them by prijUI, prijKomUI and prijPregledUI).
const (
	prijURL = "/api/komercijalno/prijemnice-sirovina"
	// The main screen: prijListaURL returns the list of the prijemnice (the filters of the main screen), prijDialogURL
	// the entry dialog of a prijemnica (GET id: a saved one, "Izmeni" of its row; pmagaciniid: a new one, "Dodaj"),
	// prijNoviURL reloads the content of the dialog with a new one (GET vrd, magaciniid and vezaid: another vrsta or
	// magacin, a porudžbenica of the "Broj porudžbenice").
	prijListaURL  = prijURL + "/lista"
	prijDialogURL = prijURL + "/dialog"
	prijNoviURL   = prijURL + "/novi"
	// prijSaveURL saves the header (POST); prijBrisiPotvrdaURL asks to confirm the deletion and prijBrisiURL + "/:id"
	// deletes it; prijPotvrdaURL asks to confirm the potvrda and prijPotvrdiURL + "/:id" confirms it.
	prijSaveURL         = prijURL + "/save"
	prijBrisiURL        = prijURL + "/brisi"
	prijBrisiPotvrdaURL = prijBrisiURL + "/potvrda"
	prijPotvrdaURL      = prijURL + "/potvrda"
	prijPotvrdiURL      = prijURL + "/potvrdi"
	// The stavke: prijArtikalURL returns the artikal of a šifra with its OTK (GET rdokid and sifra), prijStavkaURL a
	// saved stavka for the form ("Izmeni" of its row, GET rdokid and id), prijStavkaSaveURL saves a stavka (POST),
	// prijStavkaPotvrdaURL asks to confirm the deletion of a stavka and prijStavkaBrisiURL + "/:id" deletes it;
	// prijPreuzmiURL takes the stavke of the porudžbenica (POST rdokid and the selected one, or sve), prijVratiPotvrdaURL
	// asks to confirm the deletion of the last stavka or of all of them (GET rdokid and sve) and prijVratiURL + "/:id"
	// deletes them.
	prijArtikalURL       = prijURL + "/artikal"
	prijStavkaURL        = prijURL + "/stavka"
	prijStavkaSaveURL    = prijStavkaURL + "/save"
	prijStavkaPotvrdaURL = prijStavkaURL + "/potvrda"
	prijStavkaBrisiURL   = prijStavkaURL + "/brisi"
	prijPreuzmiURL       = prijStavkaURL + "/preuzmi"
	prijVratiURL         = prijStavkaURL + "/vrati"
	prijVratiPotvrdaURL  = prijVratiURL + "/potvrda"
	// prijPakovanjaURL opens the dialog "Pakovanje" (GET rdokid), prijPakovanjeSaveURL saves the pakovanje of a stavka
	// (POST).
	prijPakovanjaURL     = prijURL + "/pakovanja"
	prijPakovanjeSaveURL = prijPakovanjaURL + "/save"
	// prijPrikazStavkeURL + "/:id" returns the stavke of the selected prijemnica of the list; prijStampaURL prints a
	// prijemnica (GET rdokid), prijKartonURL the karton identifikacije of a stavka (GET rdokid and kirproid).
	prijPrikazStavkeURL = prijURL + "/prikaz/stavke"
	prijStampaURL       = prijURL + "/stampa"
	prijKartonURL       = prijURL + "/karton"

	prijDialogID            = "kom-prij-dialog"
	prijDialogCloseID       = "kom-prij-dialog-close"
	prijNazadID             = "kom-prij-nazad"
	prijBodyID              = "kom-prij-body"
	prijInfoID              = "kom-prij-info"
	prijInfoDialogID        = "kom-prij-info-dialog"
	prijStavkaInfoID        = "kom-prij-stavka-info"
	prijStavkaInfoDialogID  = "kom-prij-stavka-info-dialog"
	prijBrisanjeDialogID    = "kom-prij-brisanje"
	prijPotvrdaDialogID     = "kom-prij-potvrda"
	prijStavkaBrisanjeID    = "kom-prij-stavka-brisanje"
	prijVratiDialogID       = "kom-prij-vrati"
	prijPakovanjaDialogID   = "kom-prij-pakovanja"
	prijHeaderPanelID       = "kom-prij-header-panel"
	prijStavkePanelID       = "kom-prij-stavke-panel"
	prijStavkeTableID       = "kom-prij-stavke-table"
	prijIzvoriTableID       = "kom-prij-izvori-table"
	prijPakovanjaTableID    = "kom-prij-pakovanja-table"
	prijIzvorID             = "prijizvorid"
	prijKartonID            = "kirproid"
	prijPregledID           = "kom-prij-pregled"
	prijListaID             = "kom-prij-lista"
	prijDetaljiID           = "kom-prij-detalji"
	prijPregledInfoID       = "kom-prij-pregled-info"
	prijPregledInfoDialogID = "kom-prij-pregled-info-dialog"
	prijPregledRdokID       = "kom-prij-pregled-rdokid"
	prijPregledStampaID     = "kom-prij-pregled-stampa"
	prijPretragaID          = "kom-prij-pretraga"
	prijListaTableID        = "kom-prij-lista-table"
	prijDetaljiTableID      = "kom-prij-detalji-table"
	prijMsgPotvrdaPitanje   = "Da li želite da potvrdite dokument?"
	prijMsgBrisanjePitanje  = "Brisanje dokumenta?"
	prijMsgBrisanjeStavke   = "Brisanje podataka? Da li ste sigurni?"
	prijMsgStavkaNePostoji  = "Stavka koju birate za izmenu ne postoji."
	prijMsgArtikalNePostoji = "Artikal sa izabranom šifrom ne postoji!!!"
	prijMsgNijednaIzabrana  = "Niste selektovali nijednu stavku. Dodavanje stavke nije moguće"
)

// PrijemnicaHandler is the handler of the "Prijemnica materijala po porudžbenici" screen.
type PrijemnicaHandler struct {
	service    komsvc.PrijemnicaService
	translator *i18n.Service
	cfg        config.Config
}

// NewPrijemnicaHandler creates the handler of the "Prijemnica materijala po porudžbenici" screen.
func NewPrijemnicaHandler(s komsvc.PrijemnicaService, cfg config.Config, translator *i18n.Service) *PrijemnicaHandler {
	return &PrijemnicaHandler{service: s, translator: translator, cfg: cfg}
}

// prijUI returns the ids and the urls of the entry dialog.
func prijUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID: prijHeaderPanelID,
		StavkePanelID: prijStavkePanelID,
		ContentID:     "#" + prijBodyID,
		DialogID:      prijDialogID,
		NoviURL:       prijNoviURL,
		ArtikalURL:    prijArtikalURL,
		BodyID:        prijBodyID,
		InfoID:        prijInfoID,
		StavkaInfoID:  prijStavkaInfoID,
	}
}

// prijKomUI returns the ids of the entry dialog the shared RobnoFaktureUI does not have.
func prijKomUI() domain.KomPrijUI {
	return domain.KomPrijUI{IzvorID: prijIzvorID, KartonID: prijKartonID}
}

// prijPregledUI returns the ids and the urls of the main screen.
func prijPregledUI() domain.KomPregledUI {
	return domain.KomPregledUI{URL: prijURL, ListaURL: prijListaURL, DialogURL: prijDialogURL, RootID: prijPregledID, ListaID: prijListaID,
		DetaljiID: prijDetaljiID, InfoID: prijPregledInfoID, RdokID: prijPregledRdokID, SearchID: prijPretragaID}
}

// prijButtons builds the buttons of the entry dialog of a prijemnica, like the ones of the porudžbenica: "Sačuvaj"
// saves the header and opens the stavke, "Izmeni" switches back to the header, "Nazad" closes the dialog (the list of
// the main screen is reloaded); "Sačuvaj" of the stavka saves it (a new one, or the one of "Izmeni" of its row),
// "Odustani" gives up the change of a stavka. A new prijemnica and the deletion are "Dodaj" and "Briši" of the main
// screen.
func prijButtons() domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "kom-prij-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          prijSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + prijHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + prijBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{prijHeaderPanelID, prijStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "kom-prij-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{prijHeaderPanelID, prijStavkePanelID},
		},
		Back: domain.Button{
			Id:           prijNazadID,
			LabelText:    "Nazad",
			Icon:         "back",
			BtnClass:     common.ClassCloseButton,
			HxOnClick:    "closeDialog",
			HxOnClickArg: []any{prijDialogID},
		},
		SaveStavka: domain.Button{
			Id:                   "kom-prij-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          prijStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + prijStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + prijBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "kom-prij-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "komPrijStavkaOdustani",
		},
	}
}

// prijDugmadZa builds the potvrda ("Potvrda dok.", the legacy BTN_NoviDOK1), the print of the prijemnica (the legacy
// BTN_Print4), the buttons of the stavke of the porudžbenica ("Dodaj" the selected one, BTN_Dodaj; "Dodaj sve",
// BTN_DodajSve; "Vrati poslednju", BTN_VratiJedan; "Vrati sve", BTN_VratiSve), "Pakovanje" (BTN_Pakovanja) and "Karton
// identifikacije" of the selected stavka (BTN_KI); a new prijemnica has the potvrda and "Štampa" disabled (shown like
// the printer of the legacy window, enabled after the save).
func prijDugmadZa(rdokID int64) domain.KomPrijButtons {
	if rdokID <= 0 {
		return domain.KomPrijButtons{
			Potvrda: domain.Button{Id: "kom-prij-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, IsDisabled: true,
				BtnClass: common.ClassSaveButton + " opacity-50 cursor-not-allowed"},
			Stampa: common.SetPrintButton("kom-prij-stampa", "Štampa", "stampa", prijStampaURL, "GET", true,
				common.ClassPrintButton+" opacity-50 pointer-events-none", "stavkardokid"),
		}
	}
	stavke := func(id, label, vals string) domain.Button {
		return domain.Button{Id: id, LabelText: label, Icon: "add", IsVisible: true, BtnClass: common.ClassNewButton, HxActionURL: prijPreuzmiURL,
			HxRequestType: "POST", HxVals: vals, HxInclude: "#" + prijIzvorID, HxTarget: "#" + prijBodyID, HxSwap: "outerHTML"}
	}
	vrati := func(id, label string, sve bool) domain.Button {
		return domain.Button{Id: id, LabelText: label, Icon: "delete", IsVisible: true, BtnClass: common.ClassDeleteButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d&sve=%t", prijVratiPotvrdaURL, rdokID, sve), HxRequestType: "GET",
			HxTarget: "#" + prijStavkaInfoID, HxSwap: "innerHTML"}
	}
	return domain.KomPrijButtons{
		Potvrda: domain.Button{Id: "kom-prij-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, BtnClass: common.ClassSaveButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d", prijPotvrdaURL, rdokID), HxRequestType: "GET", HxTarget: "#" + prijInfoID, HxSwap: "innerHTML"},
		Stampa:         common.SetPrintButton("kom-prij-stampa", "Štampa", "stampa", prijStampaURL, "GET", true, common.ClassPrintButton, "stavkardokid"),
		Dodaj:          stavke("kom-prij-dodaj", "Dodaj", fmt.Sprintf(`{"rdokid": "%d"}`, rdokID)),
		DodajSve:       stavke("kom-prij-dodaj-sve", "Dodaj sve", fmt.Sprintf(`{"rdokid": "%d", "sve": "true"}`, rdokID)),
		VratiPoslednju: vrati("kom-prij-vrati-poslednju", "Vrati poslednju", false),
		VratiSve:       vrati("kom-prij-vrati-sve", "Vrati sve", true),
		Pakovanje: domain.Button{Id: "kom-prij-pakovanje", LabelText: "Pakovanje", Icon: "obrada", IsVisible: true, BtnClass: common.ClassNewButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d", prijPakovanjaURL, rdokID), HxRequestType: "GET", HxTarget: "#" + prijStavkaInfoID,
			HxSwap: "innerHTML"},
		Karton: common.SetPrintButton("kom-prij-karton", "Karton identifikacije", "stampa", prijKartonURL, "GET", true, common.ClassPrintButton,
			"stavkardokid,"+prijKartonID),
	}
}

// prijTable returns an empty grid of the screen.
func (h *PrijemnicaHandler) prijTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", prijURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, prijURL, false, false, false)
	return tbl
}

// prijEkran is the content of the entry dialog: the prijemnica, the grids and the buttons.
type prijEkran struct {
	dok    domain.KomPrijDokument
	izvori domain.TableData
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	dugmad domain.KomPrijButtons
}

// prijEkranZa builds the content of the entry dialog with the prijemnica of the id (0: a new one of the vrsta, the
// magacin and the porudžbenica): the header, the stavke of its porudžbenica (a click of a row selects it for "Dodaj"),
// the stavke with "Izmeni" and "Obriši" of every row (a click of a row selects it for "Karton identifikacije") and the
// buttons.
func (h *PrijemnicaHandler) prijEkranZa(c *gin.Context, rdokID int64, vrd, magaciniID int, vezaID int64) (prijEkran, error) {
	ctx := c.Request.Context()
	var e prijEkran
	var err error
	if e.dok, err = h.service.GetPrijemnica(ctx, rdokID, vrd, magaciniID, vezaID); err != nil {
		return e, err
	}
	header := e.dok.Header
	e.izvori = h.prijTable("", prijIzvoriTableID, h.service.IzvoriTableFields())
	e.izvori.Rows = komsvc.PrijIzvoriTableRows(e.dok.Izvori)
	e.izvori.FuncClick = fmt.Sprintf("komPrijIzbor(this, '%s')", prijIzvorID)
	e.tbl = h.prijTable("Stavke prijemnice", prijStavkeTableID, h.service.StavkeTableFields())
	stavke, err := h.service.GetStavke(ctx, header.RdokID)
	if err != nil {
		return e, err
	}
	e.tbl.Rows = komsvc.PrijStavkeTableRows(stavke, true)
	e.tbl.FuncClick = fmt.Sprintf("komPrijIzbor(this, '%s')", prijKartonID)
	e.tbl.ShowActions = len(stavke) > 0
	e.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	e.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: prijStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "komPrijStavkaUcitaj(event)"}
	e.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: prijStavkaPotvrdaURL,
		HxTarget: "#" + prijStavkaInfoID, HxSwap: "innerHTML"}
	e.btns = prijButtons()
	e.dugmad = prijDugmadZa(header.RdokID)
	return e, nil
}

// prijBody renders the content of the entry dialog (the answer of the save, of a new prijemnica, of the stavke) with
// the message of the screen (poruka, else the one of the prijemnica) in its dialog.
func (h *PrijemnicaHandler) prijBody(c *gin.Context, e prijEkran, poruka string) {
	ctx := c.Request.Context()
	if err := tmpl_kom.PrijemnicaBody(prijUI(), prijKomUI(), e.dok, e.izvori, e.tbl, e.btns, e.dugmad, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if poruka == "" {
		poruka = e.dok.Poruka
	}
	if poruka != "" {
		h.prijInfo(c, poruka)
	}
}

// prijInfo renders a message of the entry dialog in its container of the messages.
func (h *PrijemnicaHandler) prijInfo(c *gin.Context, poruka string) {
	dialog := domain.Dialog{Id: prijInfoDialogID, Title: "Info"}
	btnClose := domain.Button{Id: "kom-prij-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnOk := domain.Button{Id: "kom-prij-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
		HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
	_ = tmpl_robno.RobnoFaktureInfo(prijInfoID, h.translator.Message(poruka), dialog, btnClose, btnOk, h.translator).Render(c.Request.Context(), c.Writer)
}

// prijOdgovor renders the content of the entry dialog with the prijemnica of the id (0: a new one of the vrsta, the
// magacin and the porudžbenica), the header saved (snimljen: the stavke open) or not, with a message; an error is shown
// in the dialog of the screen.
func (h *PrijemnicaHandler) prijOdgovor(c *gin.Context, rdokID int64, vrd, magaciniID int, vezaID int64, snimljen bool, poruka string) {
	e, err := h.prijEkranZa(c, rdokID, vrd, magaciniID, vezaID)
	if err != nil {
		prijGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	e.dok.Header.Snimljen = snimljen && e.dok.Header.RdokID > 0
	h.prijBody(c, e, poruka)
}

// prijGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func prijGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+prijInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, prijInfoDialogID, message, true)
}

// prijStavkaGreska shows an error of the stavke in the container of their dialogs.
func prijStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+prijStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, prijStavkaInfoDialogID, message, true)
}

// Prijemnica opens the "Prijemnica materijala po porudžbenici" screen from the menu (the legacy pr_PrijemnicaSIROVINE,
// its tab "Pregled prijemnica" as the main screen): the filters (the first magacin of the user, the nepotvrđene
// prijemnice; all the vrste of the prijemnice of the magacin, like the legacy), the search with "Dodaj" and "Štampa",
// the list of the prijemnice and the stavke of the selected one.
func (h *PrijemnicaHandler) Prijemnica(c *gin.Context) {
	ctx := c.Request.Context()
	magacini, err := h.service.GetMagacini(ctx)
	if err != nil {
		utils.RenderDialogOK(c, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	f := komPregledFilteri{Doksts: domain.KomZahtevNepotvrdjen}
	if len(magacini) > 0 {
		f.MagaciniID = common.StringToInt(magacini[0].Key)
	}
	tbl, err := h.prijLista(c, f)
	if err != nil {
		utils.RenderDialogOK(c, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	ui := prijPregledUI()
	komDodajStampa(&tbl, ui, prijPregledStampaID, prijStampaURL, prijPregledRdokID)
	stavkeTbl := h.prijTable("", prijDetaljiTableID, h.service.PrikazStavkeTableFields())
	if err := tmpl_kom.PrijemnicaPregled(ui, magacini, f.MagaciniID, tbl, komPretraga(ui, h.translator), stavkeTbl,
		h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// prijLista builds the list of the prijemnice of the filters (the legacy PR_QRY_EDOK): "Izmeni" and "Briši" of every
// nepotvrđena prijemnica.
func (h *PrijemnicaHandler) prijLista(c *gin.Context, f komPregledFilteri) (domain.TableData, error) {
	prijemnice, err := h.service.GetPrikaz(c.Request.Context(), domain.KomPrijPrikazParams{MagaciniID: f.MagaciniID, Doksts: f.Doksts,
		SviMagacini: f.SviMagacini, Tekst: f.Tekst})
	if err != nil {
		return domain.TableData{}, err
	}
	return komLista(h.cfg, prijPregledUI(), prijListaTableID, h.service.PrikazTableFields(), komsvc.PrijPrikazTableRows(prijemnice),
		f.Doksts == domain.KomZahtevNepotvrdjen, prijBrisiPotvrdaURL, prijPrikazStavkeURL), nil
}

// PrijLista returns the list of the prijemnice of the filters of the main screen (GET pmagaciniid, pdoksts,
// psvimagacini and query: another filter or the search).
func (h *PrijemnicaHandler) PrijLista(c *gin.Context) {
	tbl, err := h.prijLista(c, komFilteri(c))
	if err != nil {
		komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	_ = tmpl_kom.KomPregledLista(tbl, h.translator).Render(c.Request.Context(), c.Writer)
}

// PrijDialog opens the entry dialog of a prijemnica: the saved one of the id ("Izmeni" of its row, GET id; a confirmed,
// posted or missing one gives a new prijemnica with the message why) or a new one of the magacin of the filters
// ("Dodaj", GET pmagaciniid).
func (h *PrijemnicaHandler) PrijDialog(c *gin.Context) {
	ctx := c.Request.Context()
	f := komFilteri(c)
	e, err := h.prijEkranZa(c, int64(common.StringToInt(c.Query("id"))), 0, f.MagaciniID, 0)
	if err != nil {
		komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	btnClose := domain.Button{Id: prijDialogCloseID, IsVisible: true, IdDialog: prijDialogID, BtnClass: common.ClassDialogCloseButton}
	if err := tmpl_kom.PrijemnicaDialog(prijUI(), prijKomUI(), e.dok, e.izvori, e.tbl, e.btns, e.dugmad, btnClose, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		h.prijInfo(c, e.dok.Poruka)
	}
}

// PrijNovi reloads the content of the entry dialog with a new prijemnica of the vrsta and the magacin (GET vrd and
// magaciniid: another vrsta or magacin, its next broj) or of a porudžbenica (GET vezaid, the legacy BTN_TraziDOK and
// PodaciHDR: its magacin, org. jedinica, mesto troška and dobavljač).
func (h *PrijemnicaHandler) PrijNovi(c *gin.Context) {
	h.prijOdgovor(c, 0, common.StringToInt(c.Query("vrd")), common.StringToInt(c.Query("magaciniid")),
		int64(common.StringToInt(c.Query("vezaid"))), false, "")
}

// SavePrij saves the header of the prijemnica ("Sačuvaj", the legacy BTN_Snimi1): the validation (ValidatePrijemnica,
// the legacy ValidacijaEDOK: field errors as JSON with status 422), the save (SavePrijemnica, SnimiEDOK) and the content
// of the dialog with the saved prijemnica: the header locked, the stavke open.
func (h *PrijemnicaHandler) SavePrij(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.KomPrijParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidatePrijemnica(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rdokID, err := h.service.SavePrijemnica(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.prijOdgovor(c, rdokID, 0, 0, 0, true, "")
}

// PotvrdaBrisanja asks to confirm the deletion of a prijemnica ("Briši" of its row, GET id; the legacy BTN_Brisanje
// "Brisanje dokumenta?"): "Da" deletes it, "Ne" closes the question.
func (h *PrijemnicaHandler) PotvrdaBrisanja(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("id")))
	h.prijPitanje(c, prijBrisanjeDialogID, "Brisanje", prijMsgBrisanjePitanje, "DELETE", fmt.Sprintf("%s/%d", prijBrisiURL, rdokID),
		"#"+prijPregledInfoID)
}

// DeletePrij deletes the prijemnica of the id with its stavke and their pakovanja (DELETE /brisi/:id): "Uspešno obrisani
// podaci..." and the list reloaded; a posted or confirmed one shows why.
func (h *PrijemnicaHandler) DeletePrij(c *gin.Context) {
	poruka, err := h.service.DeletePrijemnica(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), false, true)
	case poruka != "":
		komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, h.translator.Message(komsvc.ZahtevObrisanoPoruka), true, false)
	}
}

// PotvrdaPrij checks the prijemnica before its potvrda and asks to confirm it (GET rdokid: "Potvrda dok.", the legacy
// BTN_NoviDOK1): a prijemnica without stavke, already confirmed, with a stavka without serija, primljena količina or
// konto of the magacin shows why; else "Da li želite da potvrdite dokument?".
func (h *PrijemnicaHandler) PotvrdaPrij(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.ProveriPotvrdu(ctx, rdokID)
	if err != nil {
		prijGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		prijGreska(c, h.translator.Message(poruka))
		return
	}
	h.prijPitanje(c, prijPotvrdaDialogID, "Potvrda", prijMsgPotvrdaPitanje, "POST", fmt.Sprintf("%s/%d", prijPotvrdiURL, rdokID), "#"+prijBodyID)
}

// PotvrdiPrij confirms the prijemnica of the id (POST /potvrdi/:id, the legacy "Da", PotvrdaDOK: the stavke booked into
// the stanje): the message ("Dokument je potvrđen!!!" or why it was not confirmed) on the main screen, whose list is
// reloaded (the entry dialog with it).
func (h *PrijemnicaHandler) PotvrdiPrij(c *gin.Context) {
	poruka, err := h.service.PotvrdiPrijemnicu(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
		return
	}
	komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, h.translator.Message(poruka), true, false)
}

// prijPitanje renders a Da/Ne question of the screen: "Da" sends the request (method, url) whose answer replaces the
// target, "Ne" closes the dialog.
func (h *PrijemnicaHandler) prijPitanje(c *gin.Context, dialogID, naslov, pitanje, method, url, target string) {
	dialog := domain.Dialog{Id: dialogID, Title: naslov}
	btnClose := domain.Button{Id: dialogID + "-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{
		Id:                   dialogID + "-da",
		IsVisible:            true,
		LabelText:            "Da",
		IdDialog:             dialog.Id,
		BtnClass:             common.ClassDeleteButton,
		HxRequestType:        method,
		HxActionURL:          url,
		HxTarget:             target,
		HxSwap:               "outerHTML",
		HxOnAfterRequest:     "closeDialog",
		HxOnAfterRequestArgs: []any{dialog.Id},
	}
	btnNe := domain.Button{Id: dialogID + "-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message(pitanje)}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// PrijArtikal returns the artikal of a šifra with the OTK of its next prijemnica (GET rdokid and sifra) as JSON for the
// form of the stavke (the legacy exit of EDT_SIFRA and VratiOTK); found false when it does not exist.
func (h *PrijemnicaHandler) PrijArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(prijMsgArtikalNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "artikal": artikal})
}

// PrijStavka returns a saved stavka (GET rdokid and id) as JSON for the form of the stavke ("Izmeni" of its row).
func (h *PrijemnicaHandler) PrijStavka(c *gin.Context) {
	stavka, found, err := h.service.GetStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		int64(common.StringToInt(c.Query("id"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(prijMsgStavkaNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SavePrijStavka saves a stavka of the prijemnica (POST of the fields of the stavke, the legacy BTN_DODAJDET; the stavka
// of "Izmeni" changes its konto, količine, serija and OTK): the validation (ValidateStavka, the legacy ValidacijaEPRO:
// field errors as JSON with status 422), the save (SaveStavka, AddEPRO) and the content of the dialog (the stavke open).
func (h *PrijemnicaHandler) SavePrijStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.KomPrijStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveStavka(ctx, params); err != nil {
		prijStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.prijOdgovor(c, params.RdokID, 0, 0, 0, true, "")
}

// PotvrdaBrisanjaStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the legacy
// "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it.
func (h *PrijemnicaHandler) PotvrdaBrisanjaStavke(c *gin.Context) {
	h.prijPitanje(c, prijStavkaBrisanjeID, "Brisanje", prijMsgBrisanjeStavke, "DELETE",
		fmt.Sprintf("%s/%d?rdokid=%d", prijStavkaBrisiURL, common.StringToInt(c.Query("id")), common.StringToInt(c.Query("rdokid"))),
		"#"+prijBodyID)
}

// DeletePrijStavka deletes a stavka of the prijemnica with its pakovanje (DELETE …/stavka/brisi/:id with the rdokid, the
// legacy BTN_BrisiDET: the redni brojevi after it move back) and answers with the content of the dialog.
func (h *PrijemnicaHandler) DeletePrijStavka(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.DeleteStavka(c.Request.Context(), rdokID, int64(common.StringToInt(c.Param("id"))))
	h.prijStavkeOdgovor(c, rdokID, poruka, err, common.ErrMsgDeleteData)
}

// prijStavkeOdgovor answers a change of the stavke: an error or the message why in the container of the dialogs of the
// stavke, else the content of the dialog (the stavke open). greska is the format of the error.
func (h *PrijemnicaHandler) prijStavkeOdgovor(c *gin.Context, rdokID int64, poruka string, err error, greska string) {
	switch {
	case err != nil:
		prijStavkaGreska(c, fmt.Sprintf(greska, err.Error()))
	case poruka != "":
		prijStavkaGreska(c, h.translator.Message(poruka))
	default:
		h.prijOdgovor(c, rdokID, 0, 0, 0, true, "")
	}
}

// PreuzmiStavke takes the stavke of the porudžbenica into the prijemnica (POST rdokid with the selected prijizvorid, the
// legacy BTN_Dodaj, or sve, BTN_DodajSve) and answers with the content of the dialog; none to take, or an artikal
// already in the prijemnica ("Ovakvi podaci već postoje"), shows why.
func (h *PrijemnicaHandler) PreuzmiStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.PostForm("rdokid")))
	rproID := int64(common.StringToInt(c.PostForm(prijIzvorID)))
	if c.PostForm("sve") == "true" {
		rproID = 0
	} else if rproID <= 0 {
		prijStavkaGreska(c, h.translator.Message(prijMsgNijednaIzabrana))
		return
	}
	poruka, err := h.service.Preuzmi(ctx, rdokID, rproID)
	if err != nil {
		prijStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	// The stavke taken are shown with the message of the ones skipped (the legacy Info of AddEPRONAR).
	h.prijOdgovor(c, rdokID, 0, 0, 0, true, poruka)
}

// PotvrdaVracanja asks to confirm the deletion of the last stavka or of all of them (GET rdokid and sve: "Vrati
// poslednju" and "Vrati sve", the legacy "Brisanje podataka? Da li ste sigurni?"): "Da" deletes them.
func (h *PrijemnicaHandler) PotvrdaVracanja(c *gin.Context) {
	h.prijPitanje(c, prijVratiDialogID, "Brisanje", prijMsgBrisanjeStavke, "DELETE",
		fmt.Sprintf("%s/%d?sve=%t", prijVratiURL, common.StringToInt(c.Query("rdokid")), c.Query("sve") == "true"), "#"+prijBodyID)
}

// VratiStavke deletes the last stavka of the prijemnica of the id or all of them with their pakovanja (DELETE
// …/stavka/vrati/:id with sve, the legacy BTN_VratiJedan and BTN_VratiSve) and answers with the content of the dialog;
// without stavke it shows why.
func (h *PrijemnicaHandler) VratiStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	var poruka string
	var err error
	if c.Query("sve") == "true" {
		poruka, err = h.service.VratiSve(ctx, rdokID)
	} else {
		poruka, err = h.service.VratiPoslednju(ctx, rdokID)
	}
	h.prijStavkeOdgovor(c, rdokID, poruka, err, common.ErrMsgDeleteData)
}

// prijPakovanjaDialog renders the dialog "Pakovanje" of a prijemnica (the legacy pr_Pakovanja "Količine po pakovanju")
// with its stavke; without stavke it shows why.
func (h *PrijemnicaHandler) prijPakovanjaDialog(c *gin.Context, rdokID int64) {
	ctx := c.Request.Context()
	pakovanja, poruka, err := h.service.GetPakovanja(ctx, rdokID)
	if err != nil {
		prijStavkaGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		prijStavkaGreska(c, h.translator.Message(poruka))
		return
	}
	tbl := h.prijTable("", prijPakovanjaTableID, h.service.PakovanjaTableFields())
	tbl.Rows = komsvc.PrijPakovanjaTableRows(pakovanja)
	tbl.FuncClick = "komPrijPakovanjeIzbor(this)"
	dialog := domain.Dialog{Id: prijPakovanjaDialogID, Title: "Količine po pakovanju"}
	btnClose := domain.Button{Id: prijPakovanjaDialogID + "-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnSave := domain.Button{Id: prijPakovanjaDialogID + "-sacuvaj", LabelText: "Sačuvaj", Icon: "save", IsVisible: true, BtnClass: common.ClassSaveButton,
		HxActionURL: prijPakovanjeSaveURL, HxRequestType: "POST", HxInclude: "#" + prijPakovanjaDialogID + "-forma input",
		HxTarget: "#" + prijPakovanjaDialogID, HxSwap: "outerHTML"}
	btnNazad := domain.Button{Id: prijPakovanjaDialogID + "-nazad", LabelText: "Nazad", Icon: "back", IsVisible: true, BtnClass: common.ClassCloseButton,
		HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
	_ = tmpl_kom.PrijemnicaPakovanja(dialog, tbl, rdokID, btnSave, btnNazad, btnClose, h.translator).Render(ctx, c.Writer)
}

// PrijPakovanja opens the dialog "Pakovanje" of the prijemnica (GET rdokid, the legacy BTN_Pakovanja).
func (h *PrijemnicaHandler) PrijPakovanja(c *gin.Context) {
	h.prijPakovanjaDialog(c, int64(common.StringToInt(c.Query("rdokid"))))
}

// SavePrijPakovanje saves the pakovanje of the selected stavka (POST rdokid, pakrproid, kolpopak and brpak) and answers
// with the dialog "Pakovanje"; one that cannot be saved shows why.
func (h *PrijemnicaHandler) SavePrijPakovanje(c *gin.Context) {
	var params domain.KomPrijPakovanjeParams
	if err := c.ShouldBind(&params); err != nil {
		prijStavkaGreska(c, h.translator.Message(common.ErrMsgParseForm))
		return
	}
	if params.RproID <= 0 {
		prijStavkaGreska(c, h.translator.Message(prijMsgNijednaIzabrana))
		return
	}
	poruka, err := h.service.SavePakovanje(c.Request.Context(), params)
	if err != nil {
		prijStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	if poruka != "" {
		prijStavkaGreska(c, h.translator.Message(poruka))
		return
	}
	h.prijPakovanjaDialog(c, params.RdokID)
}

// PrijPrikazStavke returns the stavke of the prijemnica of the id (the click of a row of the main screen, the legacy
// PR_QRY_EPRO of "Detalji prijemnice").
func (h *PrijemnicaHandler) PrijPrikazStavke(c *gin.Context) {
	ctx := c.Request.Context()
	stavke, err := h.service.GetStavke(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, prijPregledInfoID, prijPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	tbl := h.prijTable("", prijDetaljiTableID, h.service.PrikazStavkeTableFields())
	tbl.Rows = komsvc.PrijPrikazStavkeTableRows(stavke)
	_ = tmpl_robno.RobnoFakturePrikazStavke(tbl, h.translator).Render(ctx, c.Writer)
}

// prijReportParams are the parameters of a print of the screen: the firm and the user.
func prijReportParams(session *domain.UserSession, firma domain.RobnoStampaFakturaFirmaDto, naziv string) domain.ReportParameters {
	return domain.ReportParameters{
		Orientation: "portrait",
		CompanyName: firma.Naziv,
		Adress:      firma.Adresa,
		Postcode:    firma.Pobro,
		City:        firma.Mesto,
		PIB:         firma.Pib,
		MatBroj:     firma.Matbr,
		SifDel:      firma.Sifdel,
		Telefon:     firma.Tel,
		CompanyLogo: robnosvc.RobnoStampaLogo(firma.Logo),
		UserName:    session.UserName,
		ReportName:  naziv,
		God:         session.SelectedGod,
	}
}

// PrijStampa prints a prijemnica (GET rdokid: "Štampa" of the entry dialog or of the selected prijemnica of the main
// screen, the legacy BTN_Print4 with PR_RPT_PRIJEMNICA); a prijemnica without stavke shows why.
func (h *PrijemnicaHandler) PrijStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		c.String(http.StatusUnauthorized, common.ErrMsgSessionNotFound)
		return
	}
	stampa, poruka, err := h.service.GetStampa(ctx, int64(common.StringToInt(c.Query("rdokid"))))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		c.String(http.StatusOK, h.translator.Message(poruka))
		return
	}
	_ = tmpl_rep_kom.PrijemnicaStampa(prijReportParams(session, stampa.Firma, "PRIJEMNICA"), stampa, h.translator).Render(ctx, c.Writer)
}

// PrijKarton prints the karton identifikacije of the selected stavka of the prijemnica (GET rdokid and kirproid, the
// legacy BTN_KI with RPT_KI); without a selected stavka it shows why.
func (h *PrijemnicaHandler) PrijKarton(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		c.String(http.StatusUnauthorized, common.ErrMsgSessionNotFound)
		return
	}
	karton, poruka, err := h.service.GetKarton(ctx, int64(common.StringToInt(c.Query("rdokid"))), int64(common.StringToInt(c.Query(prijKartonID))))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		c.String(http.StatusOK, h.translator.Message(poruka))
		return
	}
	_ = tmpl_rep_kom.PrijemnicaKarton(prijReportParams(session, domain.RobnoStampaFakturaFirmaDto{}, "KARTON IDENTIFIKACIJE"), karton,
		h.translator).Render(ctx, c.Writer)
}

// AddRoutes registers the routes of the "Prijemnica materijala po porudžbenici" screen.
func (h *PrijemnicaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/prijemnice-sirovina", h.Prijemnica)
	r.GET("/api/komercijalno/prijemnice-sirovina/lista", h.PrijLista)
	r.GET("/api/komercijalno/prijemnice-sirovina/dialog", h.PrijDialog)
	r.GET("/api/komercijalno/prijemnice-sirovina/novi", h.PrijNovi)
	r.POST("/api/komercijalno/prijemnice-sirovina/save", h.SavePrij)
	r.GET("/api/komercijalno/prijemnice-sirovina/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/prijemnice-sirovina/brisi/:id", h.DeletePrij)
	r.GET("/api/komercijalno/prijemnice-sirovina/potvrda", h.PotvrdaPrij)
	r.POST("/api/komercijalno/prijemnice-sirovina/potvrdi/:id", h.PotvrdiPrij)
	r.GET("/api/komercijalno/prijemnice-sirovina/artikal", h.PrijArtikal)
	r.GET("/api/komercijalno/prijemnice-sirovina/stavka", h.PrijStavka)
	r.POST("/api/komercijalno/prijemnice-sirovina/stavka/save", h.SavePrijStavka)
	r.GET("/api/komercijalno/prijemnice-sirovina/stavka/potvrda", h.PotvrdaBrisanjaStavke)
	r.DELETE("/api/komercijalno/prijemnice-sirovina/stavka/brisi/:id", h.DeletePrijStavka)
	r.POST("/api/komercijalno/prijemnice-sirovina/stavka/preuzmi", h.PreuzmiStavke)
	r.GET("/api/komercijalno/prijemnice-sirovina/stavka/vrati/potvrda", h.PotvrdaVracanja)
	r.DELETE("/api/komercijalno/prijemnice-sirovina/stavka/vrati/:id", h.VratiStavke)
	r.GET("/api/komercijalno/prijemnice-sirovina/pakovanja", h.PrijPakovanja)
	r.POST("/api/komercijalno/prijemnice-sirovina/pakovanja/save", h.SavePrijPakovanje)
	r.GET("/api/komercijalno/prijemnice-sirovina/prikaz/stavke/:id", h.PrijPrikazStavke)
	r.GET("/api/komercijalno/prijemnice-sirovina/stampa", h.PrijStampa)
	r.GET("/api/komercijalno/prijemnice-sirovina/karton", h.PrijKarton)
}
