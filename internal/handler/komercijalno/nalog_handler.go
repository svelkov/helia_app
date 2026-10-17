package komercijalno

import (
	"fmt"
	"net/http"

	"helia/config"
	tmpl_kom "helia/frontend/templates/komercijalno"
	tmpl_rep_kom "helia/frontend/templates/reports/komercijalno"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	komsvc "helia/internal/service/komercijalno"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// The "Nalozi za isporuku" screen of the komercijalno (the legacy pr_PorudzbenicaProdaja with its stavke
// pr_PorudzbenicaProdaja_DETALJI, the porudžbenice od kupaca, opened from the menu): its urls and the ids of its main
// screens. The entry dialog is the one of the profaktura (proUIZa, proButtonsZa: its ids, its script) with the urls of
// this screen.
const (
	nalURL = "/api/komercijalno/nalozi-isporuka"
	// The main screens: nalURL is the tab "Pregled naloga za isporuku" (nalListaURL its list), nalDelimicniURL the tab
	// "Pregled delimično realizovanih naloga" (nalDelimicniListaURL its list); nalDialogURL is the entry dialog of a
	// nalog (GET id: a saved one, "Izmeni" of its row; pmagaciniid: a new one, "Dodaj"), nalNoviURL reloads its content
	// with a new one (GET vrd and magaciniid).
	nalListaURL          = nalURL + "/lista"
	nalDelimicniURL      = nalURL + "/delimicni"
	nalDelimicniListaURL = nalDelimicniURL + "/lista"
	nalDialogURL         = nalURL + "/dialog"
	nalNoviURL           = nalURL + "/novi"
	// nalSaveURL saves the header (POST); nalBrisiPotvrdaURL asks to confirm the deletion and nalBrisiURL + "/:id"
	// deletes it; nalPotvrdaURL asks to confirm the potvrda and nalPotvrdiURL + "/:id" confirms it.
	nalSaveURL         = nalURL + "/save"
	nalBrisiURL        = nalURL + "/brisi"
	nalBrisiPotvrdaURL = nalBrisiURL + "/potvrda"
	nalPotvrdaURL      = nalURL + "/potvrda"
	nalPotvrdiURL      = nalURL + "/potvrdi"
	// The JSON of the header: the kupac, the mesto isporuke, the valuta and the cenovnik (like the profaktura), the
	// profaktura of the nalog (GET rdokid and magaciniid) and the finansijsko stanje of the kupac (GET fkto, fana and
	// datum).
	nalKupacURL      = nalURL + "/kupac"
	nalMestoURL      = nalURL + "/mesto"
	nalValutaURL     = nalURL + "/valuta"
	nalCenovnikURL   = nalURL + "/cenovnik"
	nalProfakturaURL = nalURL + "/profaktura"
	nalSaldoURL      = nalURL + "/saldo"
	// The stavke, like the ones of the profaktura.
	nalArtikalURL       = nalURL + "/artikal"
	nalStavkaURL        = nalURL + "/stavka"
	nalStavkaSaveURL    = nalStavkaURL + "/save"
	nalStavkaPotvrdaURL = nalStavkaURL + "/potvrda"
	nalStavkaBrisiURL   = nalStavkaURL + "/brisi"
	// nalPrikazStavkeURL + "/:id" and nalDelimicniStavkeURL + "/:id" return the stavke of the selected nalog of the
	// lists; nalStampaURL prints a nalog (GET rdokid).
	nalPrikazStavkeURL    = nalURL + "/prikaz/stavke"
	nalDelimicniStavkeURL = nalDelimicniURL + "/stavke"
	nalStampaURL          = nalURL + "/stampa"

	nalPregledID           = "kom-nal-pregled"
	nalDelimicniID         = "kom-nal-delimicni"
	nalListaID             = "kom-nal-lista"
	nalDetaljiID           = "kom-nal-detalji"
	nalPregledInfoID       = "kom-nal-pregled-info"
	nalPregledInfoDialogID = "kom-nal-pregled-info-dialog"
	nalPregledRdokID       = "kom-nal-pregled-rdokid"
	nalPregledStampaID     = "kom-nal-pregled-stampa"
	nalPretragaID          = "kom-nal-pretraga"
	nalListaTableID        = "kom-nal-lista-table"
	nalDetaljiTableID      = "kom-nal-detalji-table"
	nalMsgProfaktura       = "Izabrana profaktura ne postoji ili nije potvrđena!!!"
)

// NalogHandler is the handler of the "Nalozi za isporuku" screen.
type NalogHandler struct {
	service    komsvc.NalogService
	translator *i18n.Service
	cfg        config.Config
}

// NewNalogHandler creates the handler of the "Nalozi za isporuku" screen.
func NewNalogHandler(s komsvc.NalogService, cfg config.Config, translator *i18n.Service) *NalogHandler {
	return &NalogHandler{service: s, translator: translator, cfg: cfg}
}

// nalTabovi are the tabs of the main screens; aktivan is the open one (1 or 2).
func nalTabovi(aktivan int) []domain.TabItem {
	return []domain.TabItem{
		{ID: nalPregledID + "-tab", Label: "Pregled naloga za isporuku", HXRequestUrl: nalURL, IsActive: aktivan == 1},
		{ID: nalDelimicniID + "-tab", Label: "Pregled delimično realizovanih naloga", HXRequestUrl: nalDelimicniURL, IsActive: aktivan == 2},
	}
}

// nalPregledUI returns the ids and the urls of the main screen "Pregled naloga za isporuku" or, with delimicni, of
// "Pregled delimično realizovanih naloga" (without the filter of the status and the entry dialog).
func nalPregledUI(delimicni bool) domain.KomPregledUI {
	ui := domain.KomPregledUI{URL: nalURL, ListaURL: nalListaURL, DialogURL: nalDialogURL, RootID: nalPregledID, ListaID: nalListaID,
		DetaljiID: nalDetaljiID, InfoID: nalPregledInfoID, RdokID: nalPregledRdokID, SearchID: nalPretragaID, Tabovi: nalTabovi(1)}
	if delimicni {
		ui.URL, ui.ListaURL, ui.DialogURL, ui.RootID, ui.Tabovi, ui.BezStatusa = nalDelimicniURL, nalDelimicniListaURL, "", nalDelimicniID, nalTabovi(2), true
	}
	return ui
}

// nalKomUI returns the urls of the JSON of the header of the entry dialog.
func nalKomUI() domain.KomProUI {
	return domain.KomProUI{KupacURL: nalKupacURL, MestoURL: nalMestoURL, ValutaURL: nalValutaURL, CenovnikURL: nalCenovnikURL}
}

// nalNalogUI returns the urls of the JSON only the entry dialog of a nalog has.
func nalNalogUI() domain.KomNalogUI {
	return domain.KomNalogUI{ProfakturaURL: nalProfakturaURL, SaldoURL: nalSaldoURL}
}

// nalDugmadZa builds "Potvrda dok." (the legacy BTN_Potvrda) and the print (BTN_Print4) of the entry dialog; a new
// nalog has them disabled. A nalog has no "Form. nalog".
func nalDugmadZa(rdokID int64) domain.KomProButtons {
	if rdokID <= 0 {
		return domain.KomProButtons{
			Potvrda: domain.Button{Id: "kom-pro-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, IsDisabled: true,
				BtnClass: common.ClassSaveButton + " opacity-50 cursor-not-allowed"},
			Stampa: common.SetPrintButton("kom-pro-stampa", "Štampa", "stampa", nalStampaURL, "GET", true,
				common.ClassPrintButton+" opacity-50 pointer-events-none", "stavkardokid"),
		}
	}
	return domain.KomProButtons{
		Potvrda: domain.Button{Id: "kom-pro-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, BtnClass: common.ClassSaveButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d", nalPotvrdaURL, rdokID), HxRequestType: "GET", HxTarget: "#" + proInfoID, HxSwap: "innerHTML"},
		Stampa: common.SetPrintButton("kom-pro-stampa", "Štampa", "stampa", nalStampaURL, "GET", true, common.ClassPrintButton, "stavkardokid"),
	}
}

// nalTable returns an empty grid of the screen.
func (h *NalogHandler) nalTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", nalURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, nalURL, false, false, false)
	return tbl
}

// nalEkran is the content of the entry dialog: the nalog, the grid of the stavke and the buttons.
type nalEkran struct {
	dok    domain.KomNalogDokument
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	dugmad domain.KomProButtons
}

// nalEkranZa builds the content of the entry dialog with the nalog of the id (0: a new one of the vrsta and the
// magacin): the header, the stavke with "Izmeni" and "Obriši" of every row (OTK, serija and rok upotrebe only in a
// magacin with the lots) and the buttons.
func (h *NalogHandler) nalEkranZa(c *gin.Context, rdokID int64, vrd, magaciniID int) (nalEkran, error) {
	ctx := c.Request.Context()
	var e nalEkran
	var err error
	if e.dok, err = h.service.GetNalog(ctx, rdokID, vrd, magaciniID); err != nil {
		return e, err
	}
	header := e.dok.Header
	lot := header.Tipzal == 2
	e.tbl = h.nalTable("Stavke naloga", proStavkeTableID, h.service.StavkeTableFields(header.Ino, header.OznVal, lot))
	stavke, _, err := h.service.GetStavke(ctx, header.RdokID)
	if err != nil {
		return e, err
	}
	e.tbl.Rows = komsvc.ProStavkeTableRows(stavke, header.Ino, lot, true)
	e.tbl.ShowActions = len(stavke) > 0
	e.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	e.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: nalStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "komProStavkaUcitaj(event)"}
	e.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: nalStavkaPotvrdaURL,
		HxTarget: "#" + proStavkaInfoID, HxSwap: "innerHTML"}
	e.btns = proButtonsZa(nalSaveURL, nalStavkaSaveURL)
	e.dugmad = nalDugmadZa(header.RdokID)
	return e, nil
}

// nalOdgovor renders the content of the entry dialog with the nalog of the id (0: a new one of the vrsta and the
// magacin), the header saved (snimljen: the stavke open) or not, with its message; an error is shown in the dialog of
// the screen.
func (h *NalogHandler) nalOdgovor(c *gin.Context, rdokID int64, vrd, magaciniID int, snimljen bool) {
	e, err := h.nalEkranZa(c, rdokID, vrd, magaciniID)
	if err != nil {
		proGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	e.dok.Header.Snimljen = snimljen && e.dok.Header.RdokID > 0
	if err := tmpl_kom.ProfakturaBody(proUIZa(nalNoviURL, nalArtikalURL), nalKomUI(), e.dok.KomProDokument, e.tbl, e.btns, e.dugmad,
		tmpl_kom.NalogVarijanta(e.dok, nalNalogUI(), h.translator), h.translator).Render(c.Request.Context(), c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		komProInfo(c, h.translator, e.dok.Poruka)
	}
}

// nalPregledGreska shows an error on the main screen.
func (h *NalogHandler) nalPregledGreska(c *gin.Context, err error) {
	komPoruka(c, nalPregledInfoID, nalPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
}

// Nalog opens the "Nalozi za isporuku" screen from the menu with its tab "Pregled naloga za isporuku" (the legacy
// pr_PorudzbenicaProdaja: the first magacin of the user, the nepotvrđeni nalozi of all the vrste of the group NRK),
// the search with "Dodaj" and "Štampa", the list of the nalozi and the stavke of the selected one.
func (h *NalogHandler) Nalog(c *gin.Context) {
	h.nalPregled(c, false)
}

// NalogDelimicni opens the tab "Pregled delimično realizovanih naloga" (the legacy PuniTABLEHDR): the nalozi with the
// otpremnica whose stavke were not all delivered, with "Štampa".
func (h *NalogHandler) NalogDelimicni(c *gin.Context) {
	h.nalPregled(c, true)
}

// nalPregled renders a main screen of the nalozi.
func (h *NalogHandler) nalPregled(c *gin.Context, delimicni bool) {
	ctx := c.Request.Context()
	magacini, err := h.service.GetMagacini(ctx)
	if err != nil {
		utils.RenderDialogOK(c, nalPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	f := komPregledFilteri{Doksts: domain.KomZahtevNepotvrdjen}
	if len(magacini) > 0 {
		f.MagaciniID = common.StringToInt(magacini[0].Key)
	}
	tbl, err := h.nalLista(c, f, delimicni)
	if err != nil {
		utils.RenderDialogOK(c, nalPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	ui := nalPregledUI(delimicni)
	komDodajStampa(&tbl, ui, nalPregledStampaID, nalStampaURL, nalPregledRdokID)
	tab := "Pregled naloga za isporuku"
	if delimicni {
		tbl.BtnAdd = domain.Button{}
		tab = "Pregled delimično realizovanih naloga"
	}
	stavkeTbl := h.nalTable("", nalDetaljiTableID, h.service.NalogPrikazStavkeTableFields(delimicni))
	if err := tmpl_kom.NalogPregled(ui, tab, magacini, f.MagaciniID, tbl, komPretraga(ui, h.translator), stavkeTbl,
		h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, nalPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// nalLista builds the list of the nalozi of the filters: "Izmeni" and "Briši" of every new nalog (the legacy changes
// only the new ones too); the delimično realizovani have none.
func (h *NalogHandler) nalLista(c *gin.Context, f komPregledFilteri, delimicni bool) (domain.TableData, error) {
	nalozi, err := h.service.GetNalogPrikaz(c.Request.Context(), domain.KomNalogPrikazParams{MagaciniID: f.MagaciniID, Doksts: f.Doksts,
		SviMagacini: f.SviMagacini, Delimicni: delimicni, Tekst: f.Tekst})
	if err != nil {
		return domain.TableData{}, err
	}
	stavkeURL := nalPrikazStavkeURL
	if delimicni {
		stavkeURL = nalDelimicniStavkeURL
	}
	return komLista(h.cfg, nalPregledUI(delimicni), nalListaTableID, h.service.NalogPrikazTableFields(),
		komsvc.NalogPrikazTableRows(nalozi, !delimicni), !delimicni, nalBrisiPotvrdaURL, stavkeURL), nil
}

// NalLista returns the list of the nalozi of the filters of "Pregled naloga za isporuku" (GET pmagaciniid, pdoksts,
// psvimagacini and query).
func (h *NalogHandler) NalLista(c *gin.Context) {
	h.nalListaOdgovor(c, false)
}

// NalDelimicniLista returns the list of "Pregled delimično realizovanih naloga" (GET pmagaciniid, psvimagacini and
// query).
func (h *NalogHandler) NalDelimicniLista(c *gin.Context) {
	h.nalListaOdgovor(c, true)
}

func (h *NalogHandler) nalListaOdgovor(c *gin.Context, delimicni bool) {
	tbl, err := h.nalLista(c, komFilteri(c), delimicni)
	if err != nil {
		h.nalPregledGreska(c, err)
		return
	}
	_ = tmpl_kom.KomPregledLista(tbl, h.translator).Render(c.Request.Context(), c.Writer)
}

// NalDialog opens the entry dialog of a nalog: the saved one of the id ("Izmeni" of its row, GET id; a potvrđen,
// otpremljen, fiscalized, posted or missing one gives a new nalog with the message why) or a new one of the magacin of
// the filters ("Dodaj", GET pmagaciniid).
func (h *NalogHandler) NalDialog(c *gin.Context) {
	ctx := c.Request.Context()
	f := komFilteri(c)
	e, err := h.nalEkranZa(c, int64(common.StringToInt(c.Query("id"))), 0, f.MagaciniID)
	if err != nil {
		h.nalPregledGreska(c, err)
		return
	}
	btnClose := domain.Button{Id: proDialogCloseID, IsVisible: true, IdDialog: proDialogID, BtnClass: common.ClassDialogCloseButton}
	if err := tmpl_kom.ProfakturaDialog(proUIZa(nalNoviURL, nalArtikalURL), nalKomUI(), e.dok.KomProDokument, e.tbl, e.btns, e.dugmad,
		tmpl_kom.NalogVarijanta(e.dok, nalNalogUI(), h.translator), btnClose, h.translator, common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		komProInfo(c, h.translator, e.dok.Poruka)
	}
}

// NalNovi reloads the content of the entry dialog with a new nalog of the vrsta and the magacin (GET vrd and
// magaciniid).
func (h *NalogHandler) NalNovi(c *gin.Context) {
	h.nalOdgovor(c, 0, common.StringToInt(c.Query("vrd")), common.StringToInt(c.Query("magaciniid")), false)
}

// SaveNal saves the header of the nalog ("Sačuvaj", the legacy BTN_Snimi1): the validation (ValidateNalog, the legacy
// ValidacijaEDOK: field errors as JSON with status 422), the save (SaveNalog, SnimiEDOK with KopirajStavkeizProfakture)
// and the content of the dialog with the saved nalog: the header locked, the stavke open.
func (h *NalogHandler) SaveNal(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.KomNalogParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateNalog(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rdokID, err := h.service.SaveNalog(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.nalOdgovor(c, rdokID, 0, 0, true)
}

// PotvrdaBrisanja asks to confirm the deletion of a nalog ("Briši" of its row, GET id; the legacy BTN_Brisanje
// "Brisanje dokumenta?").
func (h *NalogHandler) PotvrdaBrisanja(c *gin.Context) {
	komPitanje(c, h.translator, proBrisanjeDialogID, "Brisanje", proMsgBrisanjePitanje, "DELETE",
		fmt.Sprintf("%s/%d", nalBrisiURL, common.StringToInt(c.Query("id"))), "#"+nalPregledInfoID)
}

// DeleteNal deletes the nalog of the id with its stavke (DELETE /brisi/:id): "Uspešno obrisani podaci..." and the list
// reloaded; one that is not new, fiscalized or posted shows why.
func (h *NalogHandler) DeleteNal(c *gin.Context) {
	poruka, err := h.service.DeleteNalog(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, nalPregledInfoID, nalPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), false, true)
	case poruka != "":
		komPoruka(c, nalPregledInfoID, nalPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, nalPregledInfoID, nalPregledInfoDialogID, h.translator.Message(komsvc.ZahtevObrisanoPoruka), true, false)
	}
}

// PotvrdaNal checks the nalog before its potvrda and asks to confirm it (GET rdokid: "Potvrda dok.", the legacy
// BTN_Potvrda): one without stavke, with a cena 0 or already confirmed shows why; else "Da li želite da potvrdite
// dokument?".
func (h *NalogHandler) PotvrdaNal(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.ProveriPotvrdu(c.Request.Context(), rdokID)
	if err != nil {
		proGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		proGreska(c, h.translator.Message(poruka))
		return
	}
	komPitanje(c, h.translator, proPotvrdaDialogID, "Potvrda", proMsgPotvrdaPitanje, "POST", fmt.Sprintf("%s/%d", nalPotvrdiURL, rdokID),
		"#"+proBodyID)
}

// PotvrdiNal confirms the nalog of the id (POST /potvrdi/:id, the legacy "Da": doksts 20): the message on the main
// screen, whose list is reloaded (the entry dialog with it).
func (h *NalogHandler) PotvrdiNal(c *gin.Context) {
	poruka, err := h.service.PotvrdiNalog(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, nalPregledInfoID, nalPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
		return
	}
	komPoruka(c, nalPregledInfoID, nalPregledInfoDialogID, h.translator.Message(poruka), true, false)
}

// NalKupac returns the kupac of a konto and šifra (komProKupacJSON).
func (h *NalogHandler) NalKupac(c *gin.Context) { komProKupacJSON(c, h.service, h.translator) }

// NalMesto returns a mesto isporuke of the kupac (komProMestoJSON).
func (h *NalogHandler) NalMesto(c *gin.Context) { komProMestoJSON(c, h.service) }

// NalValuta returns the valuta of a šifra with its kurs (komProValutaJSON).
func (h *NalogHandler) NalValuta(c *gin.Context) { komProValutaJSON(c, h.service) }

// NalCenovnik returns the naziv and the opis of a cenovnik (komProCenovnikJSON).
func (h *NalogHandler) NalCenovnik(c *gin.Context) { komProCenovnikJSON(c, h.service) }

// NalArtikal returns the artikal of a šifra with its lots and its cena (komProArtikalJSON).
func (h *NalogHandler) NalArtikal(c *gin.Context) { komProArtikalJSON(c, h.service, h.translator) }

// NalStavka returns a saved stavka for the form of the stavke (komProStavkaJSON).
func (h *NalogHandler) NalStavka(c *gin.Context) { komProStavkaJSON(c, h.service, h.translator) }

// NalProfaktura returns what the header takes of a potvrđena profaktura of the magacin (GET rdokid and magaciniid) as
// JSON (the legacy PodaciIzProfakture); found false for another one.
func (h *NalogHandler) NalProfaktura(c *gin.Context) {
	p, found, err := h.service.GetProfakturaNaloga(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		common.StringToInt(c.Query("magaciniid")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(nalMsgProfaktura)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "profaktura": p})
}

// NalSaldo returns the finansijsko stanje of the kupac of a konto and šifra on the datum (GET fkto, fana and datum) as
// JSON (the legacy SaldoKUPCA).
func (h *NalogHandler) NalSaldo(c *gin.Context) {
	saldo, err := h.service.GetSaldo(c.Request.Context(), c.Query("fkto"), c.Query("fana"), c.Query("datum"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saldo)
}

// SaveNalStavka saves a stavka of the nalog (POST of the fields of the stavke, the legacy BTN_DODAJDET; the stavka of
// "Izmeni" changes its količina, cena and rabat): the validation (ValidateNalogStavka, the legacy ValidacijaEPRO: field
// errors as JSON with status 422), the save (SaveNalogStavka, AddEPRO with the reservation) and the content of the
// dialog.
func (h *NalogHandler) SaveNalStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.KomNalogStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateNalogStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveNalogStavka(ctx, params); err != nil {
		proStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.nalOdgovor(c, params.RdokID, 0, 0, true)
}

// PotvrdaBrisanjaStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id).
func (h *NalogHandler) PotvrdaBrisanjaStavke(c *gin.Context) {
	komPitanje(c, h.translator, proStavkaBrisanjeID, "Brisanje", proMsgBrisanjeStavke, "DELETE",
		fmt.Sprintf("%s/%d?rdokid=%d", nalStavkaBrisiURL, common.StringToInt(c.Query("id")), common.StringToInt(c.Query("rdokid"))),
		"#"+proBodyID)
}

// DeleteNalStavka deletes a stavka of the nalog (DELETE …/stavka/brisi/:id with the rdokid, the legacy BTN_BrisiDET:
// its zalihe released) and answers with the content of the dialog.
func (h *NalogHandler) DeleteNalStavka(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.DeleteNalogStavka(c.Request.Context(), rdokID, int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		proStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
	case poruka != "":
		proStavkaGreska(c, h.translator.Message(poruka))
	default:
		h.nalOdgovor(c, rdokID, 0, 0, true)
	}
}

// NalPrikazStavke returns the stavke of the nalog of the id of "Pregled naloga za isporuku" with the total of their
// vrednost (the click of a row, the legacy PR_QRY_EPRO).
func (h *NalogHandler) NalPrikazStavke(c *gin.Context) {
	h.nalPrikazStavke(c, false)
}

// NalDelimicniStavke returns the stavke of the nalog of the id of "Pregled delimično realizovanih naloga" with the
// isporučena količina (the legacy PuniTABLEDET; the stavke not delivered fully in red).
func (h *NalogHandler) NalDelimicniStavke(c *gin.Context) {
	h.nalPrikazStavke(c, true)
}

func (h *NalogHandler) nalPrikazStavke(c *gin.Context, delimicni bool) {
	ctx := c.Request.Context()
	stavke, _, err := h.service.GetStavke(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		h.nalPregledGreska(c, err)
		return
	}
	tbl := h.nalTable("", nalDetaljiTableID, h.service.NalogPrikazStavkeTableFields(delimicni))
	var ukupno float64
	tbl.Rows, ukupno = komsvc.NalogPrikazStavkeTableRows(stavke, delimicni)
	porUkupno(&tbl, ukupno)
	_ = tmpl_robno.RobnoFakturePrikazStavke(tbl, h.translator).Render(ctx, c.Writer)
}

// NalStampa prints a nalog (GET rdokid: "Štampa" of the entry dialog or of the selected nalog of the main screens, the
// legacy BTN_Print4 with PR_RPT_NALOGPRODAJA); a nalog without stavke shows why.
func (h *NalogHandler) NalStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		c.String(http.StatusUnauthorized, common.ErrMsgSessionNotFound)
		return
	}
	stampa, poruka, err := h.service.GetNalogStampa(ctx, int64(common.StringToInt(c.Query("rdokid"))))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		c.String(http.StatusOK, h.translator.Message(poruka))
		return
	}
	_ = tmpl_rep_kom.NalogStampa(prijReportParams(session, stampa.Firma, "NALOG ZA OTPREMU"), stampa, h.translator).Render(ctx, c.Writer)
}

// AddRoutes registers the routes of the "Nalozi za isporuku" screen.
func (h *NalogHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/nalozi-isporuka", h.Nalog)
	r.GET("/api/komercijalno/nalozi-isporuka/lista", h.NalLista)
	r.GET("/api/komercijalno/nalozi-isporuka/delimicni", h.NalogDelimicni)
	r.GET("/api/komercijalno/nalozi-isporuka/delimicni/lista", h.NalDelimicniLista)
	r.GET("/api/komercijalno/nalozi-isporuka/delimicni/stavke/:id", h.NalDelimicniStavke)
	r.GET("/api/komercijalno/nalozi-isporuka/dialog", h.NalDialog)
	r.GET("/api/komercijalno/nalozi-isporuka/novi", h.NalNovi)
	r.POST("/api/komercijalno/nalozi-isporuka/save", h.SaveNal)
	r.GET("/api/komercijalno/nalozi-isporuka/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/nalozi-isporuka/brisi/:id", h.DeleteNal)
	r.GET("/api/komercijalno/nalozi-isporuka/potvrda", h.PotvrdaNal)
	r.POST("/api/komercijalno/nalozi-isporuka/potvrdi/:id", h.PotvrdiNal)
	r.GET("/api/komercijalno/nalozi-isporuka/kupac", h.NalKupac)
	r.GET("/api/komercijalno/nalozi-isporuka/mesto", h.NalMesto)
	r.GET("/api/komercijalno/nalozi-isporuka/valuta", h.NalValuta)
	r.GET("/api/komercijalno/nalozi-isporuka/cenovnik", h.NalCenovnik)
	r.GET("/api/komercijalno/nalozi-isporuka/profaktura", h.NalProfaktura)
	r.GET("/api/komercijalno/nalozi-isporuka/saldo", h.NalSaldo)
	r.GET("/api/komercijalno/nalozi-isporuka/artikal", h.NalArtikal)
	r.GET("/api/komercijalno/nalozi-isporuka/stavka", h.NalStavka)
	r.POST("/api/komercijalno/nalozi-isporuka/stavka/save", h.SaveNalStavka)
	r.GET("/api/komercijalno/nalozi-isporuka/stavka/potvrda", h.PotvrdaBrisanjaStavke)
	r.DELETE("/api/komercijalno/nalozi-isporuka/stavka/brisi/:id", h.DeleteNalStavka)
	r.GET("/api/komercijalno/nalozi-isporuka/prikaz/stavke/:id", h.NalPrikazStavke)
	r.GET("/api/komercijalno/nalozi-isporuka/stampa", h.NalStampa)
}
