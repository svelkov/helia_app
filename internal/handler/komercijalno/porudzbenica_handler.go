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

// The "Porudžbenice dobavljačima" screen of the komercijalno (the legacy pr_PorudzbenicaDOBAVLJ with its stavke
// pr_PorudzbenicaDOBAVLJ_DETALJI, opened from the menu): its urls and the ids of its elements (the templates get them by
// porUI and porKomUI).
const (
	porURL = "/api/komercijalno/porudzbenice-dobavljaca"
	// The main screen: porListaURL returns the list of the porudžbenice (the filters of the main screen), porDialogURL
	// the entry dialog of a porudžbenica (GET id: a saved one, "Izmeni" of its row; pvrd and pmagaciniid: a new one,
	// "Dodaj"), porNoviURL reloads the content of the dialog with a new one (GET vrd, magaciniid and vezaid: another vrsta
	// or magacin, a zahtev za ponudu of the "Veza pon./zahtev nab. br").
	porListaURL  = porURL + "/lista"
	porDialogURL = porURL + "/dialog"
	porNoviURL   = porURL + "/novi"
	// porSaveURL saves the header (POST); porBrisiPotvrdaURL asks to confirm the deletion and porBrisiURL + "/:id"
	// deletes it; porPotvrdaURL asks to confirm the potvrda and porPotvrdiURL + "/:id" confirms it.
	porSaveURL         = porURL + "/save"
	porBrisiURL        = porURL + "/brisi"
	porBrisiPotvrdaURL = porBrisiURL + "/potvrda"
	porPotvrdaURL      = porURL + "/potvrda"
	porPotvrdiURL      = porURL + "/potvrdi"
	// porDobavljacURL returns the dobavljač of a konto and šifra (GET fkto and fana), porValutaURL the valuta of a
	// šifra with its kurs (GET sifval and datum), both as JSON.
	porDobavljacURL = porURL + "/dobavljac"
	porValutaURL    = porURL + "/valuta"
	// The stavke: porArtikalURL returns the artikal of a šifra (GET sifra), porStavkaURL a saved stavka for the form
	// ("Izmeni" of its row, GET rdokid and id), porStavkaSaveURL saves a stavka (POST), porStavkaPotvrdaURL asks to
	// confirm the deletion of a stavka and porStavkaBrisiURL + "/:id" deletes it; porPreuzmiURL takes the stavke of the
	// zahtev za ponudu (POST rdokid and rproid: one, 0 all of them), porVratiPotvrdaURL asks to confirm the deletion of
	// the last stavka or of all of them (GET rdokid and sve) and porVratiURL + "/:id" deletes them (sve).
	porArtikalURL       = porURL + "/artikal"
	porStavkaURL        = porURL + "/stavka"
	porStavkaSaveURL    = porStavkaURL + "/save"
	porStavkaPotvrdaURL = porStavkaURL + "/potvrda"
	porStavkaBrisiURL   = porStavkaURL + "/brisi"
	porPreuzmiURL       = porStavkaURL + "/preuzmi"
	porVratiURL         = porStavkaURL + "/vrati"
	porVratiPotvrdaURL  = porVratiURL + "/potvrda"
	// porPrikazStavkeURL + "/:id" returns the stavke of the selected porudžbenica of the list; porStampaURL prints a
	// porudžbenica (GET rdokid, stampanaziv and stampakomercopis).
	porPrikazStavkeURL = porURL + "/prikaz/stavke"
	porStampaURL       = porURL + "/stampa"

	porDialogID            = "kom-por-dialog"
	porDialogCloseID       = "kom-por-dialog-close"
	porNazadID             = "kom-por-nazad"
	porBodyID              = "kom-por-body"
	porInfoID              = "kom-por-info"
	porInfoDialogID        = "kom-por-info-dialog"
	porStavkaInfoID        = "kom-por-stavka-info"
	porStavkaInfoDialogID  = "kom-por-stavka-info-dialog"
	porBrisanjeDialogID    = "kom-por-brisanje"
	porPotvrdaDialogID     = "kom-por-potvrda"
	porStavkaBrisanjeID    = "kom-por-stavka-brisanje"
	porVratiDialogID       = "kom-por-vrati"
	porHeaderPanelID       = "kom-por-header-panel"
	porStavkePanelID       = "kom-por-stavke-panel"
	porStavkeTableID       = "kom-por-stavke-table"
	porIzvoriTableID       = "kom-por-izvori-table"
	porIzvorID             = "porizvorid"
	porPregledID           = "kom-por-pregled"
	porListaID             = "kom-por-lista"
	porDetaljiID           = "kom-por-detalji"
	porPregledInfoID       = "kom-por-pregled-info"
	porPregledInfoDialogID = "kom-por-pregled-info-dialog"
	porPregledRdokID       = "kom-por-pregled-rdokid"
	porPregledStampaID     = "kom-por-pregled-stampa"
	porPretragaID          = "kom-por-pretraga"
	porListaTableID        = "kom-por-lista-table"
	porDetaljiTableID      = "kom-por-detalji-table"
	porStampaNazivID       = "stampanaziv"
	porStampaKomercopisID  = "stampakomercopis"
	porMsgPotvrdaPitanje   = "Da li želite da potvrdite dokument?"
	porMsgBrisanjePitanje  = "Brisanje dokumenta?"
	porMsgBrisanjeStavke   = "Brisanje podataka? Da li ste sigurni?"
	porMsgStavkaNePostoji  = "Stavka koju birate za izmenu ne postoji."
	porMsgArtikalNePostoji = "Artikal sa izabranom šifrom ne postoji!!!"
	porMsgDobavljac        = "Dobavljač sa takvom šifrom ne postoji!!!..."
)

// PorudzbenicaHandler is the handler of the "Porudžbenice dobavljačima" screen.
type PorudzbenicaHandler struct {
	service    komsvc.PorudzbenicaService
	translator *i18n.Service
	cfg        config.Config
}

// NewPorudzbenicaHandler creates the handler of the "Porudžbenice dobavljačima" screen.
func NewPorudzbenicaHandler(s komsvc.PorudzbenicaService, cfg config.Config, translator *i18n.Service) *PorudzbenicaHandler {
	return &PorudzbenicaHandler{service: s, translator: translator, cfg: cfg}
}

// porUI returns the ids and the urls of the entry dialog.
func porUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID: porHeaderPanelID,
		StavkePanelID: porStavkePanelID,
		ContentID:     "#" + porBodyID,
		DialogID:      porDialogID,
		NoviURL:       porNoviURL,
		ArtikalURL:    porArtikalURL,
		BodyID:        porBodyID,
		InfoID:        porInfoID,
		StavkaInfoID:  porStavkaInfoID,
	}
}

// porPregledUI returns the ids and the urls of the main screen.
func porPregledUI() domain.KomPregledUI {
	return domain.KomPregledUI{URL: porURL, ListaURL: porListaURL, DialogURL: porDialogURL, RootID: porPregledID, ListaID: porListaID,
		DetaljiID: porDetaljiID, InfoID: porPregledInfoID, RdokID: porPregledRdokID, SearchID: porPretragaID}
}

// porKomUI returns the ids and the urls of the screen the shared RobnoFaktureUI does not have.
func porKomUI() domain.KomPorUI {
	return domain.KomPorUI{DobavljacURL: porDobavljacURL, ValutaURL: porValutaURL, StampaNazivID: porStampaNazivID,
		StampaKomercopisID: porStampaKomercopisID}
}

// porButtons builds the buttons of the entry dialog of a porudžbenica, like the ones of the zahtevi: "Sačuvaj" saves the
// header and opens the stavke, "Izmeni" switches back to the header, "Nazad" closes the dialog (the list of the main
// screen is reloaded); "Sačuvaj" of the stavka saves it (a new one, or the one of "Izmeni" of its row), "Odustani"
// gives up the change of a stavka. A new porudžbenica and the deletion are "Dodaj" and "Briši" of the main screen.
func porButtons() domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "kom-por-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          porSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + porHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + porBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{porHeaderPanelID, porStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "kom-por-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{porHeaderPanelID, porStavkePanelID},
		},
		Back: domain.Button{
			Id:           porNazadID,
			LabelText:    "Nazad",
			Icon:         "back",
			BtnClass:     common.ClassCloseButton,
			HxOnClick:    "closeDialog",
			HxOnClickArg: []any{porDialogID},
		},
		SaveStavka: domain.Button{
			Id:                   "kom-por-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          porStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + porStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + porBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "kom-por-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "komPorStavkaOdustani",
		},
	}
}

// porDugmadZa builds the potvrda ("Potvrda dok.", the legacy BTN_NoviDOK1), the print of the porudžbenica (the legacy
// BTN_STAMPA) and the buttons of the stavke of the zahtev za ponudu: "Dodaj" the selected one (BTN_Dodaj), "Dodaj sve"
// (BTN_DodajSve), "Vrati poslednju" (BTN_VratiJedan) and "Vrati sve" (BTN_VratiSve); a new porudžbenica has the potvrda
// and "Štampa" disabled (enabled after the save) and none of the others.
func porDugmadZa(rdokID int64) domain.KomPorButtons {
	if rdokID <= 0 {
		return domain.KomPorButtons{
			Potvrda: domain.Button{Id: "kom-por-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, IsDisabled: true,
				BtnClass: common.ClassSaveButton + " opacity-50 cursor-not-allowed"},
			Stampa: common.SetPrintButton("kom-por-stampa", "Štampa", "stampa", porStampaURL, "GET", true,
				common.ClassPrintButton+" opacity-50 pointer-events-none", "stavkardokid,"+porStampaNazivID+","+porStampaKomercopisID),
		}
	}
	stavke := func(id, label, icon, method, url, vals string) domain.Button {
		return domain.Button{Id: id, LabelText: label, Icon: icon, IsVisible: true, BtnClass: common.ClassNewButton, HxActionURL: url,
			HxRequestType: method, HxVals: vals, HxInclude: "#" + porIzvorID, HxTarget: "#" + porBodyID, HxSwap: "outerHTML"}
	}
	vrati := func(id, label string, sve bool) domain.Button {
		return domain.Button{Id: id, LabelText: label, Icon: "delete", IsVisible: true, BtnClass: common.ClassDeleteButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d&sve=%t", porVratiPotvrdaURL, rdokID, sve), HxRequestType: "GET",
			HxTarget: "#" + porStavkaInfoID, HxSwap: "innerHTML"}
	}
	return domain.KomPorButtons{
		Potvrda: domain.Button{Id: "kom-por-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, BtnClass: common.ClassSaveButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d", porPotvrdaURL, rdokID), HxRequestType: "GET", HxTarget: "#" + porInfoID, HxSwap: "innerHTML"},
		Stampa: common.SetPrintButton("kom-por-stampa", "Štampa", "stampa", porStampaURL, "GET", true, common.ClassPrintButton,
			"stavkardokid,"+porStampaNazivID+","+porStampaKomercopisID),
		Dodaj:          stavke("kom-por-dodaj", "Dodaj", "add", "POST", porPreuzmiURL, fmt.Sprintf(`{"rdokid": "%d"}`, rdokID)),
		DodajSve:       stavke("kom-por-dodaj-sve", "Dodaj sve", "add", "POST", porPreuzmiURL, fmt.Sprintf(`{"rdokid": "%d", "sve": "true"}`, rdokID)),
		VratiPoslednju: vrati("kom-por-vrati-poslednju", "Vrati poslednju", false),
		VratiSve:       vrati("kom-por-vrati-sve", "Vrati sve", true),
	}
}

// porTable returns an empty grid of the screen.
func (h *PorudzbenicaHandler) porTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", porURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, porURL, false, false, false)
	return tbl
}

// porEkran is the content of the tab "Porudžbenice dobavljačima": the porudžbenica, the grids and the buttons.
type porEkran struct {
	dok    domain.KomPorDokument
	izvori domain.TableData
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	dugmad domain.KomPorButtons
}

// porEkranZa builds the content of the tab with the porudžbenica of the id (0: a new one of the vrsta, the magacin and
// the zahtev za ponudu): the header, the slobodne stavke of its zahtev za ponudu (a click of a row selects it for
// "Dodaj"), the stavke with "Izmeni" and "Obriši" of every row and their total, and the buttons.
func (h *PorudzbenicaHandler) porEkranZa(c *gin.Context, rdokID int64, vrd, magaciniID int, vezaID int64) (porEkran, error) {
	ctx := c.Request.Context()
	var e porEkran
	var err error
	if e.dok, err = h.service.GetPorudzbenica(ctx, rdokID, vrd, magaciniID, vezaID); err != nil {
		return e, err
	}
	header := e.dok.Header
	e.izvori = h.porTable("", porIzvoriTableID, h.service.IzvoriTableFields())
	e.izvori.Rows = komsvc.PorIzvoriTableRows(e.dok.Izvori)
	e.izvori.FuncClick = fmt.Sprintf("komPorIzvorIzbor(this, '%s')", porIzvorID)
	fields := h.service.StavkeTableFields(header.Ino, header.OznVal)
	fields[len(fields)-1].IncludeInTotals = true
	e.tbl = h.porTable("Stavke porudžbenice", porStavkeTableID, fields)
	stavke, _, err := h.service.GetStavke(ctx, header.RdokID)
	if err != nil {
		return e, err
	}
	var ukupno float64
	e.tbl.Rows, ukupno = komsvc.PorStavkeTableRows(stavke, header.Ino, true)
	porUkupno(&e.tbl, ukupno)
	e.dok.Ukupno = ukupno
	e.tbl.ShowActions = len(stavke) > 0
	e.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	e.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: porStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "komPorStavkaUcitaj(event)"}
	e.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: porStavkaPotvrdaURL,
		HxTarget: "#" + porStavkaInfoID, HxSwap: "innerHTML"}
	e.btns = porButtons()
	e.dugmad = porDugmadZa(header.RdokID)
	return e, nil
}

// porUkupno puts the total of the vrednost (the last column) in the footer of a grid of stavke.
func porUkupno(tbl *domain.TableData, ukupno float64) {
	if len(tbl.Rows) == 0 {
		return
	}
	tbl.Totals = make([]string, len(tbl.Headers))
	tbl.Totals[len(tbl.Headers)-1] = common.FormatNumberWithSystemLocale(ukupno, 2)
	tbl.HasTotals = true
}

// porBody renders the content of the tab (the answer of the save, of a selected or new porudžbenica, of the deletion
// and of the potvrda) with the message of the screen (poruka, else the one of the porudžbenica) in its dialog.
func (h *PorudzbenicaHandler) porBody(c *gin.Context, e porEkran, poruka string) {
	ctx := c.Request.Context()
	if err := tmpl_kom.PorudzbenicaBody(porUI(), porKomUI(), e.dok, e.izvori, e.tbl, e.btns, e.dugmad, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if poruka == "" {
		poruka = e.dok.Poruka
	}
	if poruka != "" {
		dialog := domain.Dialog{Id: porInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "kom-por-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "kom-por-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(porInfoID, h.translator.Message(poruka), dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// porOdgovor renders the content of the tab with the porudžbenica of the id (0: a new one of the vrsta, the magacin and
// the zahtev za ponudu), the header saved (snimljen: the stavke open) or not, with a message; an error is shown in the
// dialog of the screen.
func (h *PorudzbenicaHandler) porOdgovor(c *gin.Context, rdokID int64, vrd, magaciniID int, vezaID int64, snimljen bool, poruka string) {
	e, err := h.porEkranZa(c, rdokID, vrd, magaciniID, vezaID)
	if err != nil {
		porGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	e.dok.Header.Snimljen = snimljen && e.dok.Header.RdokID > 0
	h.porBody(c, e, poruka)
}

// porGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func porGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+porInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, porInfoDialogID, message, true)
}

// porStavkaGreska shows an error of the stavke in the container of their dialogs.
func porStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+porStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, porStavkaInfoDialogID, message, true)
}

// Porudzbenica opens the "Porudžbenice dobavljačima" screen from the menu (the legacy pr_PorudzbenicaDOBAVLJ): the main
// screen with the filters (the first vrsta and magacin, the nepotvrđene porudžbenice) and the checkboxes of the print,
// the search with "Dodaj" and "Štampa", the list of the porudžbenice and the stavke of the selected one.
func (h *PorudzbenicaHandler) Porudzbenica(c *gin.Context) {
	ctx := c.Request.Context()
	vrste, err := h.service.GetVrste(ctx)
	if err != nil {
		utils.RenderDialogOK(c, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	magacini, err := h.service.GetMagacini(ctx)
	if err != nil {
		utils.RenderDialogOK(c, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	f := komPregledFilteri{Doksts: domain.KomZahtevNepotvrdjen}
	if len(vrste) > 0 {
		f.Vrd = common.StringToInt(vrste[0].Key)
	}
	if len(magacini) > 0 {
		f.MagaciniID = common.StringToInt(magacini[0].Key)
	}
	tbl, err := h.porLista(c, f)
	if err != nil {
		utils.RenderDialogOK(c, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	ui := porPregledUI()
	komDodajStampa(&tbl, ui, porPregledStampaID, porStampaURL, porPregledRdokID+","+porStampaNazivID+","+porStampaKomercopisID)
	stavkeTbl := h.porTable("", porDetaljiTableID, h.service.PrikazStavkeTableFields())
	if err := tmpl_kom.PorudzbenicaPregled(ui, porKomUI(), vrste, magacini, f.Vrd, f.MagaciniID, tbl, komPretraga(ui, h.translator), stavkeTbl,
		h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// porLista builds the list of the porudžbenice of the filters (the legacy PR_QRY_EDOK): "Izmeni" and "Briši" of every
// nepotvrđena porudžbenica.
func (h *PorudzbenicaHandler) porLista(c *gin.Context, f komPregledFilteri) (domain.TableData, error) {
	porudzbenice, err := h.service.GetPrikaz(c.Request.Context(), domain.KomPonudaPrikazParams{Vrd: f.Vrd, MagaciniID: f.MagaciniID,
		Doksts: f.Doksts, SviMagacini: f.SviMagacini, Tekst: f.Tekst})
	if err != nil {
		return domain.TableData{}, err
	}
	return komLista(h.cfg, porPregledUI(), porListaTableID, h.service.PrikazTableFields(), komsvc.PorPrikazTableRows(porudzbenice),
		f.Doksts == domain.KomZahtevNepotvrdjen, porBrisiPotvrdaURL, porPrikazStavkeURL), nil
}

// PorLista returns the list of the porudžbenice of the filters of the main screen (GET pvrd, pmagaciniid, pdoksts,
// psvimagacini and query: another filter or the search).
func (h *PorudzbenicaHandler) PorLista(c *gin.Context) {
	tbl, err := h.porLista(c, komFilteri(c))
	if err != nil {
		komPoruka(c, porPregledInfoID, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	_ = tmpl_kom.KomPregledLista(tbl, h.translator).Render(c.Request.Context(), c.Writer)
}

// PorDialog opens the entry dialog of a porudžbenica: the saved one of the id ("Izmeni" of its row, GET id; a
// confirmed, fiscalized, posted or missing one gives a new porudžbenica with the message why) or a new one of the vrsta
// and the magacin of the filters ("Dodaj", GET pvrd and pmagaciniid).
func (h *PorudzbenicaHandler) PorDialog(c *gin.Context) {
	ctx := c.Request.Context()
	f := komFilteri(c)
	e, err := h.porEkranZa(c, int64(common.StringToInt(c.Query("id"))), f.Vrd, f.MagaciniID, 0)
	if err != nil {
		komPoruka(c, porPregledInfoID, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	btnClose := domain.Button{Id: porDialogCloseID, IsVisible: true, IdDialog: porDialogID, BtnClass: common.ClassDialogCloseButton}
	if err := tmpl_kom.PorudzbenicaDialog(porUI(), porKomUI(), e.dok, e.izvori, e.tbl, e.btns, e.dugmad, btnClose, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		dialog := domain.Dialog{Id: porInfoDialogID, Title: "Info"}
		btnX := domain.Button{Id: "kom-por-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "kom-por-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(porInfoID, h.translator.Message(e.dok.Poruka), dialog, btnX, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// PorNovi reloads the content of the entry dialog with a new porudžbenica of the vrsta and the magacin (GET vrd and
// magaciniid: another vrsta or magacin, its next broj) or of a zahtev za ponudu (GET vezaid, the legacy
// BTN_TraziDOK: its rok, način dopremanja, magacin, org. jedinica, mesto troška and dobavljač).
func (h *PorudzbenicaHandler) PorNovi(c *gin.Context) {
	h.porOdgovor(c, 0, common.StringToInt(c.Query("vrd")), common.StringToInt(c.Query("magaciniid")),
		int64(common.StringToInt(c.Query("vezaid"))), false, "")
}

// SavePor saves the header of the porudžbenica ("Sačuvaj", the legacy BTN_Snimi1): the validation
// (ValidatePorudzbenica, the legacy ValidacijaEDOK: field errors as JSON with status 422), the save (SavePorudzbenica,
// SnimiEDOK) and the content of the tab with the saved porudžbenica: the header locked, the stavke open (the legacy
// opens pr_PorudzbenicaDOBAVLJ_DETALJI).
func (h *PorudzbenicaHandler) SavePor(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.KomPorParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidatePorudzbenica(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rdokID, err := h.service.SavePorudzbenica(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.porOdgovor(c, rdokID, 0, 0, 0, true, "")
}

// PotvrdaBrisanja asks to confirm the deletion of a porudžbenica ("Briši" of its row, GET id; the legacy BTN_Brisanje
// "Brisanje dokumenta?"): "Da" deletes it, "Ne" closes the question.
func (h *PorudzbenicaHandler) PotvrdaBrisanja(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("id")))
	h.porPitanje(c, porBrisanjeDialogID, "Brisanje", porMsgBrisanjePitanje, "DELETE", fmt.Sprintf("%s/%d", porBrisiURL, rdokID),
		"#"+porPregledInfoID)
}

// DeletePor deletes the porudžbenica of the id with its stavke (DELETE /brisi/:id; their stavke of the zahtev za ponudu
// slobodne again): "Uspešno obrisani podaci..." and the list reloaded; a fiscalized, posted or confirmed one shows why.
func (h *PorudzbenicaHandler) DeletePor(c *gin.Context) {
	poruka, err := h.service.DeletePorudzbenica(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, porPregledInfoID, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), false, true)
	case poruka != "":
		komPoruka(c, porPregledInfoID, porPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, porPregledInfoID, porPregledInfoDialogID, h.translator.Message(komsvc.ZahtevObrisanoPoruka), true, false)
	}
}

// PotvrdaPor checks the porudžbenica before its potvrda and asks to confirm it (GET rdokid: "Potvrda dok.", the legacy
// BTN_NoviDOK1): a porudžbenica without stavke, with a stavka without cena or already confirmed shows why; else "Da li
// želite da potvrdite dokument?".
func (h *PorudzbenicaHandler) PotvrdaPor(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.ProveriPotvrdu(ctx, rdokID)
	if err != nil {
		porGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		porGreska(c, h.translator.Message(poruka))
		return
	}
	h.porPitanje(c, porPotvrdaDialogID, "Potvrda", porMsgPotvrdaPitanje, "POST", fmt.Sprintf("%s/%d", porPotvrdiURL, rdokID), "#"+porBodyID)
}

// PotvrdiPor confirms the porudžbenica of the id (POST /potvrdi/:id, the legacy "Da": doksts 40): the message
// ("Dokument je potvrđen!!!" or why it was not confirmed) on the main screen, whose list is reloaded (the entry dialog
// with it).
func (h *PorudzbenicaHandler) PotvrdiPor(c *gin.Context) {
	poruka, err := h.service.PotvrdiPorudzbenicu(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, porPregledInfoID, porPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
		return
	}
	komPoruka(c, porPregledInfoID, porPregledInfoDialogID, h.translator.Message(poruka), true, false)
}

// porPitanje renders a Da/Ne question of the screen: "Da" sends the request (method, url) whose answer replaces the
// target, "Ne" closes the dialog.
func (h *PorudzbenicaHandler) porPitanje(c *gin.Context, dialogID, naslov, pitanje, method, url, target string) {
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

// PorDobavljac returns the dobavljač of a konto and šifra (GET fkto and fana) as JSON for the header (the legacy exit of
// EDT_FANA); found false when it does not exist.
func (h *PorudzbenicaHandler) PorDobavljac(c *gin.Context) {
	d, found, err := h.service.GetDobavljac(c.Request.Context(), c.Query("fkto"), c.Query("fana"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(porMsgDobavljac)})
		return
	}
	d.SistemPdv = h.translator.Label(d.SistemPdv)
	c.JSON(http.StatusOK, gin.H{"found": true, "dobavljac": d})
}

// PorValuta returns the valuta of a šifra with its kurs on the datum (GET sifval and datum) as JSON for the header (the
// legacy exit of EDT_SIFVAL).
func (h *PorudzbenicaHandler) PorValuta(c *gin.Context) {
	valuta, err := h.service.GetValuta(c.Request.Context(), c.Query("sifval"), c.Query("datum"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, valuta)
}

// PorArtikal returns the artikal of a šifra with its nabavna cena (GET sifra) as JSON for the form of the stavke (the
// legacy exit of EDT_SIFRA); found false when it does not exist.
func (h *PorudzbenicaHandler) PorArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetArtikal(c.Request.Context(), c.Query("sifra"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(porMsgArtikalNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "artikal": artikal})
}

// PorStavka returns a saved stavka (GET rdokid and id) as JSON for the form of the stavke ("Izmeni" of its row).
func (h *PorudzbenicaHandler) PorStavka(c *gin.Context) {
	stavka, found, err := h.service.GetStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		int64(common.StringToInt(c.Query("id"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(porMsgStavkaNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SavePorStavka saves a stavka of the porudžbenica (POST of the fields of the stavke, the legacy BTN_DODAJDET; the
// stavka of "Izmeni" changes its količina, cena and rabat): the validation (ValidateStavka, the legacy ValidacijaEPRO:
// field errors as JSON with status 422), the save (SaveStavka, AddEPRO) and the content of the tab with the
// porudžbenica (the stavke open).
func (h *PorudzbenicaHandler) SavePorStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.KomPorStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveStavka(ctx, params); err != nil {
		porStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.porOdgovor(c, params.RdokID, 0, 0, 0, true, "")
}

// PotvrdaBrisanjaStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the legacy
// "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it.
func (h *PorudzbenicaHandler) PotvrdaBrisanjaStavke(c *gin.Context) {
	h.porPitanje(c, porStavkaBrisanjeID, "Brisanje", porMsgBrisanjeStavke, "DELETE",
		fmt.Sprintf("%s/%d?rdokid=%d", porStavkaBrisiURL, common.StringToInt(c.Query("id")), common.StringToInt(c.Query("rdokid"))),
		"#"+porBodyID)
}

// DeletePorStavka deletes a stavka of the porudžbenica (DELETE …/stavka/brisi/:id with the rdokid, the legacy
// BTN_BrisiDET: its stavka of the zahtev za ponudu is slobodna again, the redni brojevi after it move back) and answers
// with the content of the tab with the porudžbenica.
func (h *PorudzbenicaHandler) DeletePorStavka(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	if err := h.service.DeleteStavka(c.Request.Context(), rdokID, int64(common.StringToInt(c.Param("id")))); err != nil {
		porStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.porOdgovor(c, rdokID, 0, 0, 0, true, "")
}

// PreuzmiStavke takes the stavke of the zahtev za ponudu into the porudžbenica (POST rdokid with the selected porizvorid,
// the legacy BTN_Dodaj, or sve, BTN_DodajSve) and answers with the content of the tab with the porudžbenica (the stavke
// open); none to take shows why.
func (h *PorudzbenicaHandler) PreuzmiStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.PostForm("rdokid")))
	rproID := int64(common.StringToInt(c.PostForm(porIzvorID)))
	if c.PostForm("sve") == "true" {
		rproID = 0
	} else if rproID <= 0 {
		porStavkaGreska(c, h.translator.Message("Niste selektovali nijednu stavku. Dodavanje stavke nije moguće"))
		return
	}
	poruka, err := h.service.Preuzmi(ctx, rdokID, rproID)
	if err != nil {
		porStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	if poruka != "" {
		porStavkaGreska(c, h.translator.Message(poruka))
		return
	}
	h.porOdgovor(c, rdokID, 0, 0, 0, true, "")
}

// PotvrdaVracanja asks to confirm the deletion of the last stavka or of all of them (GET rdokid and sve: "Vrati
// poslednju" and "Vrati sve", the legacy "Brisanje podataka? Da li ste sigurni?"): "Da" deletes them.
func (h *PorudzbenicaHandler) PotvrdaVracanja(c *gin.Context) {
	h.porPitanje(c, porVratiDialogID, "Brisanje", porMsgBrisanjeStavke, "DELETE",
		fmt.Sprintf("%s/%d?sve=%t", porVratiURL, common.StringToInt(c.Query("rdokid")), c.Query("sve") == "true"), "#"+porBodyID)
}

// VratiStavke deletes the last stavka of the porudžbenica of the id or all of them (DELETE …/stavka/vrati/:id with sve,
// the legacy BTN_VratiJedan and BTN_VratiSve: their stavke of the zahtev za ponudu slobodne again) and answers with the
// content of the tab with the porudžbenica; without stavke it shows why.
func (h *PorudzbenicaHandler) VratiStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Param("id")))
	var poruka string
	var err error
	if c.Query("sve") == "true" {
		poruka, err = h.service.VratiSve(ctx, rdokID)
	} else {
		poruka, err = h.service.VratiPoslednju(ctx, rdokID)
	}
	if err != nil {
		porStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	if poruka != "" {
		porStavkaGreska(c, h.translator.Message(poruka))
		return
	}
	h.porOdgovor(c, rdokID, 0, 0, 0, true, "")
}

// PorPrikazStavke returns the stavke of the porudžbenica of the id with the total of their vrednost (the click of a row
// of the tab "Pregled porudžbenice", the legacy PR_QRY_EPRO).
func (h *PorudzbenicaHandler) PorPrikazStavke(c *gin.Context) {
	ctx := c.Request.Context()
	stavke, ino, err := h.service.GetStavke(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		porGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tbl := h.porTable("", porDetaljiTableID, h.service.PrikazStavkeTableFields())
	var ukupno float64
	tbl.Rows, ukupno = komsvc.PorStavkeTableRows(stavke, ino, false)
	porUkupno(&tbl, ukupno)
	_ = tmpl_robno.RobnoFakturePrikazStavke(tbl, h.translator).Render(ctx, c.Writer)
}

// PorStampa prints a porudžbenica (GET rdokid, stampanaziv and stampakomercopis: "Štampa" of the stavke or "Štampaj" of
// the selected porudžbenica of the tab "Pregled porudžbenice", the legacy BTN_STAMPA and BTN_Print4 with
// PR_RPT_PORUDZB_DOB, PR_RPT_PORUDZB_DOBINO in a foreign valuta); a porudžbenica without stavke shows why.
func (h *PorudzbenicaHandler) PorStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		c.String(http.StatusUnauthorized, common.ErrMsgSessionNotFound)
		return
	}
	stampa, poruka, err := h.service.GetStampa(ctx, int64(common.StringToInt(c.Query("rdokid"))), c.Query(porStampaNazivID) != "false",
		c.Query(porStampaKomercopisID) == "true")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		c.String(http.StatusOK, h.translator.Message(poruka))
		return
	}
	naziv := "PORUDŽBENICA"
	if stampa.Ino {
		naziv = "PURCHASE ORDER"
	}
	repParams := domain.ReportParameters{
		Orientation: "portrait",
		CompanyName: stampa.Firma.Naziv,
		Adress:      stampa.Firma.Adresa,
		Postcode:    stampa.Firma.Pobro,
		City:        stampa.Firma.Mesto,
		PIB:         stampa.Firma.Pib,
		MatBroj:     stampa.Firma.Matbr,
		SifDel:      stampa.Firma.Sifdel,
		Telefon:     stampa.Firma.Tel,
		CompanyLogo: robnosvc.RobnoStampaLogo(stampa.Firma.Logo),
		UserName:    session.UserName,
		ReportName:  naziv,
		God:         session.SelectedGod,
	}
	_ = tmpl_rep_kom.PorudzbenicaStampa(repParams, stampa, h.translator).Render(ctx, c.Writer)
}

// AddRoutes registers the routes of the "Porudžbenice dobavljačima" screen.
func (h *PorudzbenicaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/porudzbenice-dobavljaca", h.Porudzbenica)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/lista", h.PorLista)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/dialog", h.PorDialog)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/novi", h.PorNovi)
	r.POST("/api/komercijalno/porudzbenice-dobavljaca/save", h.SavePor)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/porudzbenice-dobavljaca/brisi/:id", h.DeletePor)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/potvrda", h.PotvrdaPor)
	r.POST("/api/komercijalno/porudzbenice-dobavljaca/potvrdi/:id", h.PotvrdiPor)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/dobavljac", h.PorDobavljac)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/valuta", h.PorValuta)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/artikal", h.PorArtikal)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/stavka", h.PorStavka)
	r.POST("/api/komercijalno/porudzbenice-dobavljaca/stavka/save", h.SavePorStavka)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/stavka/potvrda", h.PotvrdaBrisanjaStavke)
	r.DELETE("/api/komercijalno/porudzbenice-dobavljaca/stavka/brisi/:id", h.DeletePorStavka)
	r.POST("/api/komercijalno/porudzbenice-dobavljaca/stavka/preuzmi", h.PreuzmiStavke)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/stavka/vrati/potvrda", h.PotvrdaVracanja)
	r.DELETE("/api/komercijalno/porudzbenice-dobavljaca/stavka/vrati/:id", h.VratiStavke)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/prikaz/stavke/:id", h.PorPrikazStavke)
	r.GET("/api/komercijalno/porudzbenice-dobavljaca/stampa", h.PorStampa)
}
