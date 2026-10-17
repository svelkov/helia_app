package komercijalno

import (
	"fmt"
	"net/http"
	"strconv"

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

// The "Zahtev za ponudu" screen of the komercijalno (the legacy pr_ZahtevPONUDA, opened from the menu): its urls and
// the ids of its elements (the templates get them by ponudaUI and domain.KomPonudaUI).
const (
	ponudaURL = "/api/komercijalno/zahtevi-ponuda"
	// The main screen: ponudaListaURL returns the list of the zahtevi (the filters of the main screen), ponudaDialogURL
	// the entry dialog of a zahtev (GET id: a saved one, "Izmeni" of its row; pvrd and pmagaciniid: a new one, "Dodaj"),
	// ponudaNoviURL reloads the content of the dialog with a new zahtev (GET vrd and magaciniid: another vrsta),
	// ponudaIzvoriURL the slobodne stavke of the zahtevi za nabavku of a magacin (GET magaciniid).
	ponudaListaURL  = ponudaURL + "/lista"
	ponudaDialogURL = ponudaURL + "/dialog"
	ponudaNoviURL   = ponudaURL + "/novi"
	ponudaIzvoriURL = ponudaURL + "/izvori"
	// ponudaSaveURL saves the header with the chosen stavke (POST); ponudaBrisiPotvrdaURL asks to confirm the deletion
	// and ponudaBrisiURL + "/:id" deletes it; ponudaPotvrdaURL opens the dialog of the potvrda (the dobavljači),
	// ponudaPotvrdaPitanjeURL asks to confirm it and ponudaPotvrdiURL + "/:id" confirms it (GET or POST idfkpl).
	ponudaSaveURL           = ponudaURL + "/save"
	ponudaBrisiURL          = ponudaURL + "/brisi"
	ponudaBrisiPotvrdaURL   = ponudaBrisiURL + "/potvrda"
	ponudaPotvrdaURL        = ponudaURL + "/potvrda"
	ponudaPotvrdaPitanjeURL = ponudaPotvrdaURL + "/pitanje"
	ponudaPotvrdiURL        = ponudaURL + "/potvrdi"
	// ponudaValutaURL returns the valuta of a šifra with its kurs on a datum (GET sifval and datum).
	ponudaValutaURL = ponudaURL + "/valuta"
	// The stavke: ponudaArtikalURL returns the artikal of a šifra (GET sifra), ponudaStavkaURL a saved stavka for the
	// form ("Izmeni" of its row, GET rdokid and id), ponudaStavkaSaveURL saves a stavka (POST), ponudaPreuzmiURL takes
	// the chosen stavke of the zahtevi za nabavku (POST rdokid and izabranestavke), ponudaStavkaPotvrdaURL asks to
	// confirm the deletion of a stavka and ponudaStavkaBrisiURL + "/:id" deletes it.
	ponudaArtikalURL       = ponudaURL + "/artikal"
	ponudaStavkaURL        = ponudaURL + "/stavka"
	ponudaStavkaSaveURL    = ponudaStavkaURL + "/save"
	ponudaPreuzmiURL       = ponudaStavkaURL + "/preuzmi"
	ponudaStavkaPotvrdaURL = ponudaStavkaURL + "/potvrda"
	ponudaStavkaBrisiURL   = ponudaStavkaURL + "/brisi"
	// ponudaPrikazStavkeURL + "/:id" returns the stavke of the selected zahtev of the list; ponudaStampaDobavljaciURL
	// opens the dialog of the print (the dobavljači of the selected zahtev, GET rdokid) and ponudaStampaURL prints it (GET
	// rdokid, idfkpl).
	ponudaPrikazStavkeURL     = ponudaURL + "/prikaz/stavke"
	ponudaStampaURL           = ponudaURL + "/stampa"
	ponudaStampaDobavljaciURL = ponudaStampaURL + "/dobavljaci"

	ponudaDialogID            = "kom-ponuda-dialog"
	ponudaDialogCloseID       = "kom-ponuda-dialog-close"
	ponudaNazadID             = "kom-ponuda-nazad"
	ponudaBodyID              = "kom-ponuda-body"
	ponudaInfoID              = "kom-ponuda-info"
	ponudaInfoDialogID        = "kom-ponuda-info-dialog"
	ponudaStavkaInfoID        = "kom-ponuda-stavka-info"
	ponudaStavkaInfoDialogID  = "kom-ponuda-stavka-info-dialog"
	ponudaBrisanjeDialogID    = "kom-ponuda-brisanje"
	ponudaStavkaBrisanjeID    = "kom-ponuda-stavka-brisanje"
	ponudaPotvrdaDialogID     = "kom-ponuda-potvrda"
	ponudaPotvrdaPitanjeID    = "kom-ponuda-potvrda-pitanje"
	ponudaPotvrdaPitanjeDlgID = "kom-ponuda-potvrda-pitanje-dialog"
	ponudaPotvrdaIDFkplID     = "kom-ponuda-potvrda-idfkpl"
	ponudaPotvrdaRdokID       = "kom-ponuda-potvrda-rdokid"
	ponudaStampaDialogID      = "kom-ponuda-stampa"
	ponudaStampaIDFkplID      = "kom-ponuda-stampa-idfkpl"
	ponudaStampaRdokID        = "kom-ponuda-stampa-rdokid"
	ponudaHeaderPanelID       = "kom-ponuda-header-panel"
	ponudaStavkePanelID       = "kom-ponuda-stavke-panel"
	ponudaStavkeTableID       = "kom-ponuda-stavke-table"
	ponudaIzvoriTableID       = "kom-ponuda-izvori-table"
	ponudaIzvoriGridID        = "kom-ponuda-izvori-grid"
	ponudaIzabraneID          = "izabranestavke"
	ponudaPregledID           = "kom-ponuda-pregled"
	ponudaListaID             = "kom-ponuda-lista"
	ponudaDetaljiID           = "kom-ponuda-detalji"
	ponudaPregledInfoID       = "kom-ponuda-pregled-info"
	ponudaPregledInfoDialogID = "kom-ponuda-pregled-info-dialog"
	ponudaPregledRdokID       = "kom-ponuda-pregled-rdokid"
	ponudaPregledStampaID     = "kom-ponuda-pregled-stampa"
	ponudaPretragaID          = "kom-ponuda-pretraga"
	ponudaListaTableID        = "kom-ponuda-lista-table"
	ponudaDetaljiTableID      = "kom-ponuda-detalji-table"
	ponudaDobavljaciTableID   = "kom-ponuda-dobavljaci-table"
	ponudaMsgPotvrdaPitanje   = "Da li želite da potvrdite dokument?"
	ponudaMsgBrisanjePitanje  = "Brisanje dokumenta?"
	ponudaMsgBrisanjeStavke   = "Brisanje podataka? Da li ste sigurni?"
	ponudaMsgStavkaNePostoji  = "Stavka koju birate za izmenu ne postoji."
	ponudaMsgArtikalNePostoji = "Artikal sa izabranom šifrom ne postoji!!!"
)

// ZahtevPonudaHandler is the handler of the "Zahtev za ponudu" screen.
type ZahtevPonudaHandler struct {
	service    komsvc.ZahtevPonudaService
	translator *i18n.Service
	cfg        config.Config
}

// NewZahtevPonudaHandler creates the handler of the "Zahtev za ponudu" screen.
func NewZahtevPonudaHandler(s komsvc.ZahtevPonudaService, cfg config.Config, translator *i18n.Service) *ZahtevPonudaHandler {
	return &ZahtevPonudaHandler{service: s, translator: translator, cfg: cfg}
}

// ponudaUI returns the ids and the urls of the entry dialog.
func ponudaUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID: ponudaHeaderPanelID,
		StavkePanelID: ponudaStavkePanelID,
		ContentID:     "#" + ponudaBodyID,
		DialogID:      ponudaDialogID,
		NoviURL:       ponudaNoviURL,
		ArtikalURL:    ponudaArtikalURL,
		BodyID:        ponudaBodyID,
		InfoID:        ponudaInfoID,
		StavkaInfoID:  ponudaStavkaInfoID,
	}
}

// ponudaPregledUI returns the ids and the urls of the main screen.
func ponudaPregledUI() domain.KomPregledUI {
	return domain.KomPregledUI{URL: ponudaURL, ListaURL: ponudaListaURL, DialogURL: ponudaDialogURL, RootID: ponudaPregledID,
		ListaID: ponudaListaID, DetaljiID: ponudaDetaljiID, InfoID: ponudaPregledInfoID, RdokID: ponudaPregledRdokID,
		SearchID: ponudaPretragaID}
}

// ponudaKomUI returns the ids and the urls of the screen the zahtev za nabavku does not have.
func ponudaKomUI() domain.KomPonudaUI {
	return domain.KomPonudaUI{
		IzvoriURL:     ponudaIzvoriURL,
		IzvoriGridID:  ponudaIzvoriGridID,
		IzabraneID:    ponudaIzabraneID,
		ValutaURL:     ponudaValutaURL,
		PotvrdaIDFkpl: ponudaPotvrdaIDFkplID,
		StampaIDFkpl:  ponudaStampaIDFkplID,
	}
}

// ponudaButtons builds the buttons of the entry dialog of a zahtev za ponudu, like the ones of the zahtev za nabavku:
// "Sačuvaj" saves the header with the chosen stavke of the zahtevi za nabavku and opens the stavke, "Izmeni" switches
// back to the header, "Nazad" closes the dialog (the list of the main screen is reloaded); "Sačuvaj" of the stavka saves
// a stavka entered by hand (or the količina of the one of "Izmeni" of its row), "Odustani" gives up the change of a
// stavka. The potvrda and "+" of the stavke are ponudaPotvrdaButton and ponudaPreuzmiButton; a new zahtev and the
// deletion are "Dodaj" and "Briši" of the main screen.
func ponudaButtons() domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "kom-ponuda-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          ponudaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + ponudaHeaderPanelID + " [data-panel-fields], #magaciniid, #" + ponudaIzabraneID,
			HxTarget:             "#" + ponudaBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{ponudaHeaderPanelID, ponudaStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "kom-ponuda-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{ponudaHeaderPanelID, ponudaStavkePanelID},
		},
		Back: domain.Button{
			Id:           ponudaNazadID,
			LabelText:    "Nazad",
			Icon:         "back",
			BtnClass:     common.ClassCloseButton,
			HxOnClick:    "closeDialog",
			HxOnClickArg: []any{ponudaDialogID},
		},
		SaveStavka: domain.Button{
			Id:                   "kom-ponuda-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          ponudaStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + ponudaStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + ponudaBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "kom-ponuda-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "komPonudaStavkaOdustani",
		},
	}
}

// ponudaPotvrdaButton is "Potvrda" of the header (the legacy BTN_Potvrda): it opens the dialog of the potvrda of the
// saved zahtev; a new one has nothing to confirm.
func ponudaPotvrdaButton(rdokID int64) domain.Button {
	if rdokID <= 0 {
		return domain.Button{Id: "kom-ponuda-potvrda-btn", LabelText: "Potvrda", Icon: "obrada", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassSaveButton + " opacity-50 cursor-not-allowed"}
	}
	return domain.Button{
		Id:            "kom-ponuda-potvrda-btn",
		LabelText:     "Potvrda",
		Icon:          "obrada",
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   fmt.Sprintf("%s?rdokid=%d", ponudaPotvrdaURL, rdokID),
		HxRequestType: "GET",
		HxTarget:      "#" + ponudaInfoID,
		HxSwap:        "innerHTML",
	}
}

// ponudaStampaButton is "Štampa" of the entry dialog: like "Štampa" of the main screen it opens the dialog of the
// dobavljači of the saved zahtev (the print is for one dobavljač); always shown, disabled until the zahtev is saved.
func ponudaStampaButton(rdokID int64) domain.Button {
	if rdokID <= 0 {
		return domain.Button{Id: "kom-ponuda-stampa", LabelText: "Štampa", Icon: "stampa", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassPrintButton + " opacity-50 pointer-events-none"}
	}
	return domain.Button{Id: "kom-ponuda-stampa", LabelText: "Štampa", Icon: "stampa", IsVisible: true, BtnClass: common.ClassPrintButton,
		HxActionURL: fmt.Sprintf("%s?rdokid=%d", ponudaStampaDobavljaciURL, rdokID), HxRequestType: "GET", HxTarget: "#" + ponudaInfoID,
		HxSwap: "innerHTML"}
}

// ponudaPreuzmiButton is "+" of the stavke (the legacy BTN_DodajG): it takes the chosen stavke of the zahtevi za nabavku
// into the saved zahtev.
func ponudaPreuzmiButton() domain.Button {
	return domain.Button{
		Id:            "kom-ponuda-preuzmi",
		LabelText:     "Dodaj izabrane stavke",
		Icon:          "add",
		BtnClass:      common.ClassNewButton,
		HxActionURL:   ponudaPreuzmiURL,
		HxRequestType: "POST",
		HxInclude:     "#stavkardokid, #" + ponudaIzabraneID,
		HxTarget:      "#" + ponudaBodyID,
		HxSwap:        "outerHTML",
	}
}

// ponudaTable returns an empty grid of the screen.
func (h *ZahtevPonudaHandler) ponudaTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", ponudaURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, ponudaURL, false, false, false)
	return tbl
}

// izvoriTable returns the grid of the slobodne stavke of the zahtevi za nabavku: a click of a row chooses it or not.
func (h *ZahtevPonudaHandler) izvoriTable(izvori []domain.KomPonudaIzvorRow) domain.TableData {
	tbl := h.ponudaTable("", ponudaIzvoriTableID, h.service.IzvoriTableFields())
	tbl.Rows = komsvc.PonudaIzvoriTableRows(izvori)
	tbl.FuncClick = "komPonudaIzvorIzbor(this)"
	return tbl
}

// ponudaEkran is the content of the tab "Ponude dobavljačima": the zahtev, the grids and the buttons.
type ponudaEkran struct {
	dok        domain.KomPonudaDokument
	izvori     domain.TableData
	tbl        domain.TableData
	btns       domain.RobnoFaktureButtons
	btnPotvrda domain.Button
	btnPreuzmi domain.Button
}

// ponudaEkranZa builds the content of the tab with the zahtev of the id (0: a new one of the vrsta and the magacin): the
// slobodne stavke of the zahtevi za nabavku, the header, the stavke with "Izmeni" and "Obriši" of every row and the
// buttons.
func (h *ZahtevPonudaHandler) ponudaEkranZa(c *gin.Context, rdokID int64, vrd, magaciniID int) (ponudaEkran, error) {
	ctx := c.Request.Context()
	var e ponudaEkran
	var err error
	if e.dok, err = h.service.GetPonuda(ctx, rdokID, vrd, magaciniID); err != nil {
		return e, err
	}
	header := e.dok.Header
	e.izvori = h.izvoriTable(e.dok.Izvori)
	e.tbl = h.ponudaTable("Stavke zahteva", ponudaStavkeTableID, h.service.StavkeTableFields())
	stavke, err := h.service.GetStavke(ctx, header.RdokID)
	if err != nil {
		return e, err
	}
	e.tbl.Rows = komsvc.PonudaStavkeTableRows(stavke)
	e.tbl.ShowActions = len(stavke) > 0
	e.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	e.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: ponudaStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "komPonudaStavkaUcitaj(event)"}
	e.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: ponudaStavkaPotvrdaURL,
		HxTarget: "#" + ponudaStavkaInfoID, HxSwap: "innerHTML"}
	e.btns = ponudaButtons()
	e.btns.Print = ponudaStampaButton(header.RdokID)
	e.btnPotvrda = ponudaPotvrdaButton(header.RdokID)
	e.btnPreuzmi = ponudaPreuzmiButton()
	return e, nil
}

// ponudaBody renders the content of the tab (the answer of the save, of a selected or new zahtev, of the deletion and of
// the potvrda) with the message of the screen (poruka, else the one of the zahtev) in its dialog.
func (h *ZahtevPonudaHandler) ponudaBody(c *gin.Context, e ponudaEkran, poruka string) {
	ctx := c.Request.Context()
	if err := tmpl_kom.ZahtevPonudaBody(ponudaUI(), ponudaKomUI(), e.dok, e.izvori, e.tbl, e.btns, e.btnPotvrda, e.btnPreuzmi,
		h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if poruka == "" {
		poruka = e.dok.Poruka
	}
	if poruka != "" {
		dialog := domain.Dialog{Id: ponudaInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "kom-ponuda-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "kom-ponuda-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(ponudaInfoID, h.translator.Message(poruka), dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// ponudaOdgovor renders the content of the tab with the zahtev of the id (0: a new one of the vrsta and the magacin), the
// header saved (snimljen: the stavke open) or not, with a message; an error is shown in the dialog of the screen.
func (h *ZahtevPonudaHandler) ponudaOdgovor(c *gin.Context, rdokID int64, vrd, magaciniID int, snimljen bool, poruka string) {
	e, err := h.ponudaEkranZa(c, rdokID, vrd, magaciniID)
	if err != nil {
		ponudaGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	e.dok.Header.Snimljen = snimljen && e.dok.Header.RdokID > 0
	h.ponudaBody(c, e, poruka)
}

// ponudaGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func ponudaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+ponudaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, ponudaInfoDialogID, message, true)
}

// ponudaStavkaGreska shows an error of the stavke in the container of their dialogs.
func ponudaStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+ponudaStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, ponudaStavkaInfoDialogID, message, true)
}

// ZahtevPonuda opens the "Zahtev za ponudu" screen from the menu (the legacy pr_ZahtevPONUDA): the main screen with the
// filters (the first vrsta and magacin, the nepotvrđeni zahtevi), the search with "Dodaj" and "Štampa" (the dialog of
// the dobavljači of the selected zahtev, the legacy BTN_Print4 and pr_StampaPONUDA), the list of the zahtevi and the
// stavke of the selected one.
func (h *ZahtevPonudaHandler) ZahtevPonuda(c *gin.Context) {
	ctx := c.Request.Context()
	vrste, err := h.service.GetVrste(ctx)
	if err != nil {
		utils.RenderDialogOK(c, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	magacini, err := h.service.GetMagacini(ctx)
	if err != nil {
		utils.RenderDialogOK(c, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	f := komPregledFilteri{Doksts: domain.KomZahtevNepotvrdjen}
	if len(vrste) > 0 {
		f.Vrd = common.StringToInt(vrste[0].Key)
	}
	if len(magacini) > 0 {
		f.MagaciniID = common.StringToInt(magacini[0].Key)
	}
	tbl, err := h.ponudaLista(c, f)
	if err != nil {
		utils.RenderDialogOK(c, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	ui := ponudaPregledUI()
	komDodajStampa(&tbl, ui, ponudaPregledStampaID, ponudaStampaURL, ponudaPregledRdokID)
	tbl.BtnPrint = domain.Button{Id: ponudaPregledStampaID, LabelText: "Štampa", Icon: "stampa", IsVisible: true, OpenDialog: true,
		BtnClass: common.ClassPrintButton + " opacity-50 pointer-events-none", HxActionURL: ponudaStampaDobavljaciURL, HxRequestType: "GET",
		HxInclude: "#" + ponudaPregledRdokID, HxTarget: "#" + ponudaPregledInfoID, HxSwap: "innerHTML"}
	stavkeTbl := h.ponudaTable("", ponudaDetaljiTableID, h.service.PrikazStavkeTableFields())
	if err := tmpl_kom.ZahtevPonudaPregled(ui, vrste, magacini, f.Vrd, f.MagaciniID, tbl, komPretraga(ui, h.translator), stavkeTbl,
		h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// ponudaLista builds the list of the zahtevi za ponudu of the filters (the legacy PuniHDRPONUDE): "Izmeni" and "Briši" of
// every nepotvrđen zahtev.
func (h *ZahtevPonudaHandler) ponudaLista(c *gin.Context, f komPregledFilteri) (domain.TableData, error) {
	ponude, err := h.service.GetPrikaz(c.Request.Context(), domain.KomPonudaPrikazParams{Vrd: f.Vrd, MagaciniID: f.MagaciniID, Doksts: f.Doksts,
		SviMagacini: f.SviMagacini, Tekst: f.Tekst})
	if err != nil {
		return domain.TableData{}, err
	}
	return komLista(h.cfg, ponudaPregledUI(), ponudaListaTableID, h.service.PrikazTableFields(), komsvc.PonudaPrikazTableRows(ponude),
		f.Doksts == domain.KomZahtevNepotvrdjen, ponudaBrisiPotvrdaURL, ponudaPrikazStavkeURL), nil
}

// PonudaLista returns the list of the zahtevi za ponudu of the filters of the main screen (GET pvrd, pmagaciniid,
// pdoksts, psvimagacini and query: another filter or the search).
func (h *ZahtevPonudaHandler) PonudaLista(c *gin.Context) {
	tbl, err := h.ponudaLista(c, komFilteri(c))
	if err != nil {
		komPoruka(c, ponudaPregledInfoID, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	_ = tmpl_kom.KomPregledLista(tbl, h.translator).Render(c.Request.Context(), c.Writer)
}

// PonudaDialog opens the entry dialog of a zahtev za ponudu: the saved one of the id ("Izmeni" of its row, GET id; a
// confirmed or a missing one gives a new zahtev with the message why) or a new one of the vrsta and the magacin of the
// filters ("Dodaj", GET pvrd and pmagaciniid) with the slobodne stavke of the zahtevi za nabavku of the magacin.
func (h *ZahtevPonudaHandler) PonudaDialog(c *gin.Context) {
	ctx := c.Request.Context()
	f := komFilteri(c)
	e, err := h.ponudaEkranZa(c, int64(common.StringToInt(c.Query("id"))), f.Vrd, f.MagaciniID)
	if err != nil {
		komPoruka(c, ponudaPregledInfoID, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	btnClose := domain.Button{Id: ponudaDialogCloseID, IsVisible: true, IdDialog: ponudaDialogID, BtnClass: common.ClassDialogCloseButton}
	if err := tmpl_kom.ZahtevPonudaDialog(ponudaUI(), ponudaKomUI(), e.dok, e.izvori, e.tbl, e.btns, e.btnPotvrda, e.btnPreuzmi, btnClose,
		h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		dialog := domain.Dialog{Id: ponudaInfoDialogID, Title: "Info"}
		btnX := domain.Button{Id: "kom-ponuda-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "kom-ponuda-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(ponudaInfoID, h.translator.Message(e.dok.Poruka), dialog, btnX, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// PonudaNovi reloads the content of the entry dialog with a new zahtev of the vrsta and the magacin (GET vrd and
// magaciniid: another vrsta of a new zahtev, its next broj).
func (h *ZahtevPonudaHandler) PonudaNovi(c *gin.Context) {
	h.ponudaOdgovor(c, 0, common.StringToInt(c.Query("vrd")), common.StringToInt(c.Query("magaciniid")), false, "")
}

// PonudaIzvori returns the grid of the slobodne stavke of the zahtevi za nabavku of a magacin (GET magaciniid: another
// magacin of a new zahtev, the legacy "Select a row in Combo_MAGACIN").
func (h *ZahtevPonudaHandler) PonudaIzvori(c *gin.Context) {
	ctx := c.Request.Context()
	izvori, err := h.service.GetIzvori(ctx, common.StringToInt(c.Query("magaciniid")))
	if err != nil {
		ponudaGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	_ = tmpl_kom.PonudaIzvoriGrid(ponudaKomUI(), h.izvoriTable(izvori), h.translator).Render(ctx, c.Writer)
}

// SavePonuda saves the header of the zahtev with the chosen stavke of the zahtevi za nabavku ("Sačuvaj", the legacy
// BTN_Sacuvaj): the validation (ValidatePonuda, the legacy ValidacijaEDOK: field errors as JSON with status 422), the
// save (SavePonuda, SnimiEDOK and AddEPRONAR) and the content of the tab with the saved zahtev: the header locked, the
// stavke open.
func (h *ZahtevPonudaHandler) SavePonuda(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.KomPonudaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidatePonuda(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rdokID, err := h.service.SavePonuda(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.ponudaOdgovor(c, rdokID, 0, 0, true, "")
}

// PotvrdaBrisanja asks to confirm the deletion of a zahtev za ponudu ("Briši" of its row, GET id; the legacy BTN_Brisanje
// "Brisanje dokumenta?"): "Da" deletes it, "Ne" closes the question.
func (h *ZahtevPonudaHandler) PotvrdaBrisanja(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("id")))
	h.ponudaPitanje(c, ponudaBrisanjeDialogID, "Brisanje", ponudaMsgBrisanjePitanje, "DELETE", fmt.Sprintf("%s/%d", ponudaBrisiURL, rdokID),
		"#"+ponudaPregledInfoID, ponudaBrisanjeDialogID)
}

// DeletePonuda deletes the zahtev za ponudu of the id with its stavke (DELETE /brisi/:id; their stavke of the zahtevi za
// nabavku slobodne again): "Uspešno obrisani podaci..." and the list reloaded; a fiscalized or posted one shows why.
func (h *ZahtevPonudaHandler) DeletePonuda(c *gin.Context) {
	poruka, err := h.service.DeletePonuda(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, ponudaPregledInfoID, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), false, true)
	case poruka != "":
		komPoruka(c, ponudaPregledInfoID, ponudaPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, ponudaPregledInfoID, ponudaPregledInfoDialogID, h.translator.Message(komsvc.ZahtevObrisanoPoruka), true, false)
	}
}

// PotvrdaPonude opens the dialog of the potvrda of the zahtev (GET rdokid: "Potvrda", the legacy BTN_Potvrda and
// pr_PotvrdaPONUDE): a zahtev without stavke, already confirmed or without dobavljači shows why; else its dobavljači,
// the first one chosen, with "Potvrda" and "Nazad".
func (h *ZahtevPonudaHandler) PotvrdaPonude(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dobavljaci, poruka, err := h.service.ProveriPotvrdu(ctx, rdokID)
	if err != nil {
		ponudaGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		ponudaGreska(c, h.translator.Message(poruka))
		return
	}
	btnPotvrda := domain.Button{
		Id:            ponudaPotvrdaDialogID + "-potvrdi",
		LabelText:     "Potvrda",
		Icon:          "obrada",
		IsVisible:     true,
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   ponudaPotvrdaPitanjeURL,
		HxRequestType: "GET",
		HxInclude:     "#" + ponudaPotvrdaRdokID + ", #" + ponudaPotvrdaIDFkplID,
		HxTarget:      "#" + ponudaPotvrdaPitanjeID,
		HxSwap:        "innerHTML",
	}
	h.dobavljaciDialog(c, ponudaPotvrdaDialogID, "Izabrani dobavljači", dobavljaci, rdokID, ponudaPotvrdaRdokID, ponudaPotvrdaIDFkplID,
		btnPotvrda, ponudaPotvrdaPitanjeID)
}

// dobavljaciDialog renders a dialog of the dobavljači of a zahtev (the potvrda, the print): the grid (a click of a row
// chooses the dobavljač, the first one is chosen), the hidden rdokid and idfkpl, the button of the action, "Nazad" and
// the container of the question of the action.
func (h *ZahtevPonudaHandler) dobavljaciDialog(c *gin.Context, dialogID, naslov string, dobavljaci []domain.KomDobavljac, rdokID int64,
	rdokInputID, idfkplInputID string, btnAkcija domain.Button, pitanjeID string) {
	tbl := h.ponudaTable("", ponudaDobavljaciTableID, h.service.DobavljaciTableFields())
	tbl.Rows = komsvc.PonudaDobavljaciTableRows(dobavljaci)
	tbl.FuncClick = fmt.Sprintf("komPonudaDobavljacIzbor(this, '%s')", idfkplInputID)
	prvi := ""
	if len(dobavljaci) > 0 {
		prvi = strconv.FormatInt(dobavljaci[0].IDFkpl, 10)
	}
	dialog := domain.Dialog{Id: dialogID, Title: naslov}
	btnClose := domain.Button{Id: dialogID + "-close", IsVisible: true, IdDialog: dialogID, BtnClass: common.ClassDialogCloseButton}
	btnNazad := domain.Button{Id: dialogID + "-nazad", LabelText: "Nazad", Icon: "back", IsVisible: true, BtnClass: common.ClassCloseButton,
		HxOnClick: "closeDialog", HxOnClickArg: []any{dialogID}}
	_ = tmpl_kom.PonudaDobavljaciDialog(dialog, tbl, strconv.FormatInt(rdokID, 10), rdokInputID, prvi, idfkplInputID, btnAkcija, btnNazad,
		btnClose, pitanjeID, h.translator).Render(c.Request.Context(), c.Writer)
}

// PotvrdaPitanje checks the potvrda again and asks to confirm it for the chosen dobavljač (GET rdokid and idfkpl, the
// legacy "Da li želite da potvrdite dokument?"): "Da" confirms it, "Ne" closes the question (the dialog of the potvrda
// stays).
func (h *ZahtevPonudaHandler) PotvrdaPitanje(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	idFkpl := int64(common.StringToInt(c.Query("idfkpl")))
	if _, poruka, err := h.service.ProveriPotvrdu(ctx, rdokID); err != nil {
		ponudaGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	} else if poruka != "" {
		ponudaGreska(c, h.translator.Message(poruka))
		return
	}
	h.ponudaPitanje(c, ponudaPotvrdaPitanjeDlgID, "Potvrda", ponudaMsgPotvrdaPitanje, "POST",
		fmt.Sprintf("%s/%d?idfkpl=%d", ponudaPotvrdiURL, rdokID, idFkpl), "#"+ponudaBodyID, ponudaPotvrdaPitanjeDlgID)
}

// PotvrdiPonudu confirms the zahtev of the id for the dobavljač of the idfkpl (POST /potvrdi/:id, the legacy "Da" of
// pr_PotvrdaPONUDE: doksts 40, fkto and fana of the dobavljač): the message ("Dokument je potvrđen!!!" or why it was not
// confirmed) on the main screen, whose list is reloaded (the entry dialog with it).
func (h *ZahtevPonudaHandler) PotvrdiPonudu(c *gin.Context) {
	poruka, err := h.service.PotvrdiPonudu(c.Request.Context(), int64(common.StringToInt(c.Param("id"))), int64(common.StringToInt(c.Query("idfkpl"))))
	if err != nil {
		komPoruka(c, ponudaPregledInfoID, ponudaPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
		return
	}
	komPoruka(c, ponudaPregledInfoID, ponudaPregledInfoDialogID, h.translator.Message(poruka), true, false)
}

// ponudaPitanje renders a Da/Ne question of the screen: "Da" sends the request (method, url) whose answer replaces the
// target and closes the dialogs of zatvori, "Ne" closes the question.
func (h *ZahtevPonudaHandler) ponudaPitanje(c *gin.Context, dialogID, naslov, pitanje, method, url, target string, zatvori ...string) {
	dialog := domain.Dialog{Id: dialogID, Title: naslov}
	btnClose := domain.Button{Id: dialogID + "-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	args := make([]any, 0, len(zatvori))
	for _, id := range zatvori {
		args = append(args, id)
	}
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
		HxOnAfterRequestArgs: args,
	}
	btnNe := domain.Button{Id: dialogID + "-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm([]string{h.translator.Message(pitanje)}, dialog, btnClose, btnDa, btnNe, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// PonudaValuta returns the valuta of a šifra with its kurs on the datum (GET sifval and datum) as JSON for the header
// (the legacy exit of EDT_SIFVAL).
func (h *ZahtevPonudaHandler) PonudaValuta(c *gin.Context) {
	valuta, err := h.service.GetValuta(c.Request.Context(), c.Query("sifval"), c.Query("datum"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, valuta)
}

// PonudaArtikal returns the artikal of a šifra (GET sifra) as JSON for the form of the stavke (the legacy exit of
// EDT_SIFRA); found false when it does not exist.
func (h *ZahtevPonudaHandler) PonudaArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetArtikal(c.Request.Context(), c.Query("sifra"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(ponudaMsgArtikalNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "artikal": artikal})
}

// PonudaStavka returns a saved stavka (GET rdokid and id) as JSON for the form of the stavke ("Izmeni" of its row).
func (h *ZahtevPonudaHandler) PonudaStavka(c *gin.Context) {
	stavka, found, err := h.service.GetStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		int64(common.StringToInt(c.Query("id"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(ponudaMsgStavkaNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SavePonudaStavka saves a stavka of the zahtev (POST of the fields of the stavke, the legacy BTN_DODAJDET; the stavka
// of "Izmeni" changes its količina): the validation (ValidateStavka, the legacy ValidacijaEPRO: field errors as JSON
// with status 422), the save (SaveStavka, AddEPRO) and the content of the tab with the zahtev (the stavke open).
func (h *ZahtevPonudaHandler) SavePonudaStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.KomPonudaStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveStavka(ctx, params); err != nil {
		ponudaStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.ponudaOdgovor(c, params.RdokID, 0, 0, true, "")
}

// PreuzmiStavke takes the chosen stavke of the zahtevi za nabavku into the zahtev (POST rdokid and izabranestavke: "+"
// of the stavke, the legacy BTN_DodajG) and answers with the content of the tab with the zahtev (the stavke open); none
// chosen shows why.
func (h *ZahtevPonudaHandler) PreuzmiStavke(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.PostForm("rdokid")))
	poruka, err := h.service.PreuzmiStavke(ctx, rdokID, c.PostForm("izabranestavke"))
	if err != nil {
		ponudaStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	if poruka != "" {
		ponudaStavkaGreska(c, h.translator.Message(poruka))
		return
	}
	h.ponudaOdgovor(c, rdokID, 0, 0, true, "")
}

// PotvrdaBrisanjaStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the legacy
// "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it.
func (h *ZahtevPonudaHandler) PotvrdaBrisanjaStavke(c *gin.Context) {
	h.ponudaPitanje(c, ponudaStavkaBrisanjeID, "Brisanje", ponudaMsgBrisanjeStavke, "DELETE",
		fmt.Sprintf("%s/%d?rdokid=%d", ponudaStavkaBrisiURL, common.StringToInt(c.Query("id")), common.StringToInt(c.Query("rdokid"))),
		"#"+ponudaBodyID, ponudaStavkaBrisanjeID)
}

// DeletePonudaStavka deletes a stavka of the zahtev (DELETE …/stavka/brisi/:id with the rdokid, the legacy
// BTN_BrisiDET: its stavka of the zahtev za nabavku is slobodna again, the redni brojevi after it move back) and
// answers with the content of the tab with the zahtev.
func (h *ZahtevPonudaHandler) DeletePonudaStavka(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	if err := h.service.DeleteStavka(c.Request.Context(), rdokID, int64(common.StringToInt(c.Param("id")))); err != nil {
		ponudaStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.ponudaOdgovor(c, rdokID, 0, 0, true, "")
}

// PonudaPrikazStavke returns the stavke of the zahtev of the id (the click of a row of the list of the main screen, the
// legacy PR_QRY_EPRO).
func (h *ZahtevPonudaHandler) PonudaPrikazStavke(c *gin.Context) {
	ctx := c.Request.Context()
	stavke, err := h.service.GetStavke(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		ponudaGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tbl := h.ponudaTable("", ponudaDetaljiTableID, h.service.PrikazStavkeTableFields())
	tbl.Rows = komsvc.ZahtevPrikazStavkeTableRows(stavke)
	_ = tmpl_robno.RobnoFakturePrikazStavke(tbl, h.translator).Render(ctx, c.Writer)
}

// StampaDobavljaci opens the dialog of the print of the selected zahtev (GET rdokid: "Štampa" of the main screen, the
// legacy BTN_Print4 and pr_StampaPONUDA): its dobavljači, the first one chosen, with "Štampaj" (the print
// for the chosen dobavljač) and "Nazad"; a zahtev without stavke or dobavljači shows why.
func (h *ZahtevPonudaHandler) StampaDobavljaci(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	dobavljaci, poruka, err := h.service.GetStampaDobavljaci(ctx, rdokID)
	if err != nil {
		ponudaGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		ponudaGreska(c, h.translator.Message(poruka))
		return
	}
	btnStampa := common.SetPrintButton(ponudaStampaDialogID+"-stampaj", "Štampaj", "stampa", ponudaStampaURL, "GET", true,
		common.ClassPrintButton, ponudaStampaRdokID+","+ponudaStampaIDFkplID)
	h.dobavljaciDialog(c, ponudaStampaDialogID, "Spisak dobavljača kojima se šalje ponuda", dobavljaci, rdokID, ponudaStampaRdokID,
		ponudaStampaIDFkplID, btnStampa, "")
}

// PonudaStampa prints a zahtev za ponudu for one of its dobavljači (GET rdokid and idfkpl: "Štampaj" of the dialog of the
// print, the legacy PR_RPT_ZAHTEV_PONUDA); a zahtev without stavke shows why.
func (h *ZahtevPonudaHandler) PonudaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		c.String(http.StatusUnauthorized, common.ErrMsgSessionNotFound)
		return
	}
	stampa, poruka, err := h.service.GetStampa(ctx, int64(common.StringToInt(c.Query("rdokid"))), int64(common.StringToInt(c.Query("idfkpl"))))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		c.String(http.StatusOK, h.translator.Message(poruka))
		return
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
		ReportName:  "ZAHTEV ZA PONUDU",
		God:         session.SelectedGod,
	}
	_ = tmpl_rep_kom.ZahtevPonudaStampa(repParams, stampa, h.translator).Render(ctx, c.Writer)
}

// AddRoutes registers the routes of the "Zahtev za ponudu" screen.
func (h *ZahtevPonudaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/zahtevi-ponuda", h.ZahtevPonuda)
	r.GET("/api/komercijalno/zahtevi-ponuda/lista", h.PonudaLista)
	r.GET("/api/komercijalno/zahtevi-ponuda/dialog", h.PonudaDialog)
	r.GET("/api/komercijalno/zahtevi-ponuda/novi", h.PonudaNovi)
	r.GET("/api/komercijalno/zahtevi-ponuda/izvori", h.PonudaIzvori)
	r.POST("/api/komercijalno/zahtevi-ponuda/save", h.SavePonuda)
	r.GET("/api/komercijalno/zahtevi-ponuda/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/zahtevi-ponuda/brisi/:id", h.DeletePonuda)
	r.GET("/api/komercijalno/zahtevi-ponuda/potvrda", h.PotvrdaPonude)
	r.GET("/api/komercijalno/zahtevi-ponuda/potvrda/pitanje", h.PotvrdaPitanje)
	r.POST("/api/komercijalno/zahtevi-ponuda/potvrdi/:id", h.PotvrdiPonudu)
	r.GET("/api/komercijalno/zahtevi-ponuda/valuta", h.PonudaValuta)
	r.GET("/api/komercijalno/zahtevi-ponuda/artikal", h.PonudaArtikal)
	r.GET("/api/komercijalno/zahtevi-ponuda/stavka", h.PonudaStavka)
	r.POST("/api/komercijalno/zahtevi-ponuda/stavka/save", h.SavePonudaStavka)
	r.POST("/api/komercijalno/zahtevi-ponuda/stavka/preuzmi", h.PreuzmiStavke)
	r.GET("/api/komercijalno/zahtevi-ponuda/stavka/potvrda", h.PotvrdaBrisanjaStavke)
	r.DELETE("/api/komercijalno/zahtevi-ponuda/stavka/brisi/:id", h.DeletePonudaStavka)
	r.GET("/api/komercijalno/zahtevi-ponuda/prikaz/stavke/:id", h.PonudaPrikazStavke)
	r.GET("/api/komercijalno/zahtevi-ponuda/stampa/dobavljaci", h.StampaDobavljaci)
	r.GET("/api/komercijalno/zahtevi-ponuda/stampa", h.PonudaStampa)
}
