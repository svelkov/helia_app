package komercijalno

import (
	"context"
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
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// The "Profakture" screen of the komercijalno (the legacy pr_Profakture with its stavke pr_Profakture_DETALJI, opened
// from the menu): its urls and the ids of its elements (the templates get them by proUI, proKomUI and proPregledUI).
const (
	proURL = "/api/komercijalno/profakture"
	// The main screen: proListaURL returns the list of the profakture (the filters of the main screen), proDialogURL
	// the entry dialog of a profaktura (GET id: a saved one, "Izmeni" of its row; pmagaciniid: a new one, "Dodaj"),
	// proNoviURL reloads the content of the dialog with a new one (GET vrd and magaciniid).
	proListaURL  = proURL + "/lista"
	proDialogURL = proURL + "/dialog"
	proNoviURL   = proURL + "/novi"
	// proSaveURL saves the header (POST); proBrisiPotvrdaURL asks to confirm the deletion and proBrisiURL + "/:id"
	// deletes it; proPotvrdaURL asks to confirm the potvrda and proPotvrdiURL + "/:id" confirms it; proNalogPotvrdaURL
	// asks to confirm the forming of the nalog and proNalogURL + "/:id" forms it.
	proSaveURL         = proURL + "/save"
	proBrisiURL        = proURL + "/brisi"
	proBrisiPotvrdaURL = proBrisiURL + "/potvrda"
	proPotvrdaURL      = proURL + "/potvrda"
	proPotvrdiURL      = proURL + "/potvrdi"
	proNalogURL        = proURL + "/nalog"
	proNalogPotvrdaURL = proNalogURL + "/potvrda"
	// The JSON of the header: proKupacURL the kupac of a konto and šifra (GET fkto and fana), proMestoURL a mesto
	// isporuke (GET fkto, fana and mi), proValutaURL the valuta of a šifra with its kurs (GET sifval and datum),
	// proCenovnikURL the opis of a cenovnik (GET id).
	proKupacURL    = proURL + "/kupac"
	proMestoURL    = proURL + "/mesto"
	proValutaURL   = proURL + "/valuta"
	proCenovnikURL = proURL + "/cenovnik"
	// The stavke: proArtikalURL returns the artikal of a šifra with its lots and cena (GET rdokid, sifra and
	// sastanjem), proStavkaURL a saved stavka for the form ("Izmeni" of its row, GET rdokid and id), proStavkaSaveURL
	// saves a stavka (POST), proStavkaPotvrdaURL asks to confirm the deletion of a stavka and proStavkaBrisiURL + "/:id"
	// deletes it.
	proArtikalURL       = proURL + "/artikal"
	proStavkaURL        = proURL + "/stavka"
	proStavkaSaveURL    = proStavkaURL + "/save"
	proStavkaPotvrdaURL = proStavkaURL + "/potvrda"
	proStavkaBrisiURL   = proStavkaURL + "/brisi"
	// proPrikazStavkeURL + "/:id" returns the stavke of the selected profaktura of the list; proStampaURL prints a
	// profaktura (GET rdokid).
	proPrikazStavkeURL = proURL + "/prikaz/stavke"
	proStampaURL       = proURL + "/stampa"

	proDialogID            = "kom-pro-dialog"
	proDialogCloseID       = "kom-pro-dialog-close"
	proNazadID             = "kom-pro-nazad"
	proBodyID              = "kom-pro-body"
	proInfoID              = "kom-pro-info"
	proInfoDialogID        = "kom-pro-info-dialog"
	proStavkaInfoID        = "kom-pro-stavka-info"
	proStavkaInfoDialogID  = "kom-pro-stavka-info-dialog"
	proBrisanjeDialogID    = "kom-pro-brisanje"
	proPotvrdaDialogID     = "kom-pro-potvrda"
	proNalogDialogID       = "kom-pro-nalog"
	proStavkaBrisanjeID    = "kom-pro-stavka-brisanje"
	proHeaderPanelID       = "kom-pro-header-panel"
	proStavkePanelID       = "kom-pro-stavke-panel"
	proStavkeTableID       = "kom-pro-stavke-table"
	proPregledID           = "kom-pro-pregled"
	proListaID             = "kom-pro-lista"
	proDetaljiID           = "kom-pro-detalji"
	proPregledInfoID       = "kom-pro-pregled-info"
	proPregledInfoDialogID = "kom-pro-pregled-info-dialog"
	proPregledRdokID       = "kom-pro-pregled-rdokid"
	proPregledStampaID     = "kom-pro-pregled-stampa"
	proPretragaID          = "kom-pro-pretraga"
	proListaTableID        = "kom-pro-lista-table"
	proDetaljiTableID      = "kom-pro-detalji-table"
	proMsgPotvrdaPitanje   = "Da li želite da potvrdite dokument?"
	proMsgNalogPitanje     = "Formiranje naloga za selektovanu profakturu?"
	proMsgBrisanjePitanje  = "Brisanje dokumenta?"
	proMsgBrisanjeStavke   = "Brisanje podataka? Da li ste sigurni?"
	proMsgStavkaNePostoji  = "Stavka koju birate za izmenu ne postoji."
	proMsgArtikalNePostoji = "Artikal sa izabranom šifrom ne postoji!!!"
	proMsgKupac            = "Partner sa takvom šifrom ne postoji!!!..."
)

// ProfakturaHandler is the handler of the "Profakture" screen.
type ProfakturaHandler struct {
	service    komsvc.ProfakturaService
	translator *i18n.Service
	cfg        config.Config
}

// NewProfakturaHandler creates the handler of the "Profakture" screen.
func NewProfakturaHandler(s komsvc.ProfakturaService, cfg config.Config, translator *i18n.Service) *ProfakturaHandler {
	return &ProfakturaHandler{service: s, translator: translator, cfg: cfg}
}

// proUI returns the ids and the urls of the entry dialog.
func proUI() domain.RobnoFaktureUI {
	return proUIZa(proNoviURL, proArtikalURL)
}

// proUIZa returns the ids of the entry dialog of the profaktura and of its variants (their script works with these ids)
// with the urls of a new document and of the artikal of a stavka of the screen.
func proUIZa(noviURL, artikalURL string) domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID: proHeaderPanelID,
		StavkePanelID: proStavkePanelID,
		ContentID:     "#" + proBodyID,
		DialogID:      proDialogID,
		NoviURL:       noviURL,
		ArtikalURL:    artikalURL,
		BodyID:        proBodyID,
		InfoID:        proInfoID,
		StavkaInfoID:  proStavkaInfoID,
	}
}

// proKomUI returns the urls of the entry dialog the shared RobnoFaktureUI does not have.
func proKomUI() domain.KomProUI {
	return domain.KomProUI{KupacURL: proKupacURL, MestoURL: proMestoURL, ValutaURL: proValutaURL, CenovnikURL: proCenovnikURL}
}

// proPregledUI returns the ids and the urls of the main screen.
func proPregledUI() domain.KomPregledUI {
	return domain.KomPregledUI{URL: proURL, ListaURL: proListaURL, DialogURL: proDialogURL, RootID: proPregledID, ListaID: proListaID,
		DetaljiID: proDetaljiID, InfoID: proPregledInfoID, RdokID: proPregledRdokID, SearchID: proPretragaID}
}

// proButtons builds the buttons of the entry dialog of a profaktura, like the ones of the porudžbenica: "Sačuvaj" saves
// the header and opens the stavke, "Izmeni" switches back to the header, "Nazad" closes the dialog (the list of the main
// screen is reloaded); "Sačuvaj" of the stavka saves it (a new one, or the one of "Izmeni" of its row), "Odustani"
// gives up the change of a stavka.
func proButtons() domain.RobnoFaktureButtons {
	return proButtonsZa(proSaveURL, proStavkaSaveURL)
}

// proButtonsZa builds the buttons of the entry dialog of the profaktura and of its variants with the urls of the save
// of the header and of a stavka of the screen.
func proButtonsZa(saveURL, stavkaSaveURL string) domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "kom-pro-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          saveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + proHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + proBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{proHeaderPanelID, proStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "kom-pro-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{proHeaderPanelID, proStavkePanelID},
		},
		Back: domain.Button{
			Id:           proNazadID,
			LabelText:    "Nazad",
			Icon:         "back",
			BtnClass:     common.ClassCloseButton,
			HxOnClick:    "closeDialog",
			HxOnClickArg: []any{proDialogID},
		},
		SaveStavka: domain.Button{
			Id:                   "kom-pro-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          stavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + proStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + proBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "kom-pro-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "komProStavkaOdustani",
		},
	}
}

// proDugmadZa builds the potvrda ("Potvrda dok.", the legacy BTN_Potvrda), "Form. nalog" (BTN_FAKTURA) and the print
// (BTN_Print4); a new profaktura has them disabled (the print shown like the printer of the legacy window).
func proDugmadZa(rdokID int64) domain.KomProButtons {
	if rdokID <= 0 {
		return domain.KomProButtons{
			Potvrda: domain.Button{Id: "kom-pro-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, IsDisabled: true,
				BtnClass: common.ClassSaveButton + " opacity-50 cursor-not-allowed"},
			Nalog: domain.Button{Id: "kom-pro-nalog-btn", LabelText: "Form. nalog", Icon: "obrada", IsVisible: true, IsDisabled: true,
				BtnClass: common.ClassNewButton + " opacity-50 cursor-not-allowed"},
			Stampa: common.SetPrintButton("kom-pro-stampa", "Štampa", "stampa", proStampaURL, "GET", true,
				common.ClassPrintButton+" opacity-50 pointer-events-none", "stavkardokid"),
		}
	}
	return domain.KomProButtons{
		Potvrda: domain.Button{Id: "kom-pro-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, BtnClass: common.ClassSaveButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d", proPotvrdaURL, rdokID), HxRequestType: "GET", HxTarget: "#" + proInfoID, HxSwap: "innerHTML"},
		Nalog: domain.Button{Id: "kom-pro-nalog-btn", LabelText: "Form. nalog", Icon: "obrada", IsVisible: true, BtnClass: common.ClassNewButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d", proNalogPotvrdaURL, rdokID), HxRequestType: "GET", HxTarget: "#" + proInfoID, HxSwap: "innerHTML"},
		Stampa: common.SetPrintButton("kom-pro-stampa", "Štampa", "stampa", proStampaURL, "GET", true, common.ClassPrintButton, "stavkardokid"),
	}
}

// proTable returns an empty grid of the screen.
func (h *ProfakturaHandler) proTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", proURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, proURL, false, false, false)
	return tbl
}

// proEkran is the content of the entry dialog: the profaktura, the grid of the stavke and the buttons.
type proEkran struct {
	dok    domain.KomProDokument
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	dugmad domain.KomProButtons
}

// proEkranZa builds the content of the entry dialog with the profaktura of the id (0: a new one of the vrsta and the
// magacin): the header, the stavke with "Izmeni" and "Obriši" of every row (OTK, serija and rok upotrebe only in a
// magacin with the lots) and the buttons.
func (h *ProfakturaHandler) proEkranZa(c *gin.Context, rdokID int64, vrd, magaciniID int) (proEkran, error) {
	ctx := c.Request.Context()
	var e proEkran
	var err error
	if e.dok, err = h.service.GetProfaktura(ctx, rdokID, vrd, magaciniID); err != nil {
		return e, err
	}
	header := e.dok.Header
	lot := header.Tipzal == 2
	e.tbl = h.proTable("Stavke profakture", proStavkeTableID, h.service.StavkeTableFields(header.Ino, header.OznVal, lot))
	stavke, _, err := h.service.GetStavke(ctx, header.RdokID)
	if err != nil {
		return e, err
	}
	e.tbl.Rows = komsvc.ProStavkeTableRows(stavke, header.Ino, lot, true)
	e.tbl.ShowActions = len(stavke) > 0
	e.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	e.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: proStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "komProStavkaUcitaj(event)"}
	e.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: proStavkaPotvrdaURL,
		HxTarget: "#" + proStavkaInfoID, HxSwap: "innerHTML"}
	e.btns = proButtons()
	e.dugmad = proDugmadZa(header.RdokID)
	return e, nil
}

// proInfo renders a message of the entry dialog in its container of the messages (komProInfo).
func (h *ProfakturaHandler) proInfo(c *gin.Context, poruka string) {
	komProInfo(c, h.translator, poruka)
}

// komProInfo renders a message of the entry dialog of the profaktura (and of its variants) in its container of the
// messages.
func komProInfo(c *gin.Context, translator *i18n.Service, poruka string) {
	dialog := domain.Dialog{Id: proInfoDialogID, Title: "Info"}
	btnClose := domain.Button{Id: "kom-pro-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnOk := domain.Button{Id: "kom-pro-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
		HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
	_ = tmpl_robno.RobnoFaktureInfo(proInfoID, translator.Message(poruka), dialog, btnClose, btnOk, translator).Render(c.Request.Context(), c.Writer)
}

// proOdgovor renders the content of the entry dialog with the profaktura of the id (0: a new one of the vrsta and the
// magacin), the header saved (snimljen: the stavke open) or not, with a message; an error is shown in the dialog of the
// screen.
func (h *ProfakturaHandler) proOdgovor(c *gin.Context, rdokID int64, vrd, magaciniID int, snimljen bool, poruka string) {
	e, err := h.proEkranZa(c, rdokID, vrd, magaciniID)
	if err != nil {
		proGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	e.dok.Header.Snimljen = snimljen && e.dok.Header.RdokID > 0
	if err := tmpl_kom.ProfakturaBody(proUI(), proKomUI(), e.dok, e.tbl, e.btns, e.dugmad, tmpl_kom.ProfakturaVarijanta(), h.translator).Render(c.Request.Context(), c.Writer); err != nil {
		return
	}
	if poruka == "" {
		poruka = e.dok.Poruka
	}
	if poruka != "" {
		h.proInfo(c, poruka)
	}
}

// proGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func proGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+proInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, proInfoDialogID, message, true)
}

// proStavkaGreska shows an error of the stavke in the container of their dialogs.
func proStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+proStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, proStavkaInfoDialogID, message, true)
}

// Profaktura opens the "Profakture" screen from the menu (the legacy pr_Profakture, its tab "Pregled profaktura" as the
// main screen): the filters (the first magacin of the user, the nepotvrđene profakture; all the vrste of the
// profakture), the search with "Dodaj" and "Štampa", the list of the profakture and the stavke of the selected one.
func (h *ProfakturaHandler) Profaktura(c *gin.Context) {
	ctx := c.Request.Context()
	magacini, err := h.service.GetMagacini(ctx)
	if err != nil {
		utils.RenderDialogOK(c, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	f := komPregledFilteri{Doksts: domain.KomZahtevNepotvrdjen}
	if len(magacini) > 0 {
		f.MagaciniID = common.StringToInt(magacini[0].Key)
	}
	tbl, err := h.proLista(c, f)
	if err != nil {
		utils.RenderDialogOK(c, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	ui := proPregledUI()
	komDodajStampa(&tbl, ui, proPregledStampaID, proStampaURL, proPregledRdokID)
	stavkeTbl := h.proTable("", proDetaljiTableID, h.service.PrikazStavkeTableFields())
	if err := tmpl_kom.ProfakturaPregled(ui, magacini, f.MagaciniID, tbl, komPretraga(ui, h.translator), stavkeTbl,
		h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// proLista builds the list of the profakture of the filters (the legacy PR_QRY_EDOK): "Izmeni" and "Briši" of every
// profaktura (the confirmed ones can change too, like the legacy; one with a nalog shows why it cannot).
func (h *ProfakturaHandler) proLista(c *gin.Context, f komPregledFilteri) (domain.TableData, error) {
	profakture, err := h.service.GetPrikaz(c.Request.Context(), domain.KomProPrikazParams{MagaciniID: f.MagaciniID, Doksts: f.Doksts,
		SviMagacini: f.SviMagacini, Tekst: f.Tekst})
	if err != nil {
		return domain.TableData{}, err
	}
	return komLista(h.cfg, proPregledUI(), proListaTableID, h.service.PrikazTableFields(), komsvc.ProPrikazTableRows(profakture, true),
		true, proBrisiPotvrdaURL, proPrikazStavkeURL), nil
}

// ProLista returns the list of the profakture of the filters of the main screen (GET pmagaciniid, pdoksts,
// psvimagacini and query).
func (h *ProfakturaHandler) ProLista(c *gin.Context) {
	tbl, err := h.proLista(c, komFilteri(c))
	if err != nil {
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	_ = tmpl_kom.KomPregledLista(tbl, h.translator).Render(c.Request.Context(), c.Writer)
}

// ProDialog opens the entry dialog of a profaktura: the saved one of the id ("Izmeni" of its row, GET id; a
// fiscalized, posted one, one with a nalog or a missing one gives a new profaktura with the message why) or a new one of
// the magacin of the filters ("Dodaj", GET pmagaciniid).
func (h *ProfakturaHandler) ProDialog(c *gin.Context) {
	ctx := c.Request.Context()
	f := komFilteri(c)
	e, err := h.proEkranZa(c, int64(common.StringToInt(c.Query("id"))), 0, f.MagaciniID)
	if err != nil {
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	btnClose := domain.Button{Id: proDialogCloseID, IsVisible: true, IdDialog: proDialogID, BtnClass: common.ClassDialogCloseButton}
	if err := tmpl_kom.ProfakturaDialog(proUI(), proKomUI(), e.dok, e.tbl, e.btns, e.dugmad, tmpl_kom.ProfakturaVarijanta(), btnClose, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		h.proInfo(c, e.dok.Poruka)
	}
}

// ProNovi reloads the content of the entry dialog with a new profaktura of the vrsta and the magacin (GET vrd and
// magaciniid: another vrsta or magacin, its next broj and the vrste naloga za knjiženje of the vrsta).
func (h *ProfakturaHandler) ProNovi(c *gin.Context) {
	h.proOdgovor(c, 0, common.StringToInt(c.Query("vrd")), common.StringToInt(c.Query("magaciniid")), false, "")
}

// SavePro saves the header of the profaktura ("Sačuvaj", the legacy BTN_Snimi1): the validation (ValidateProfaktura,
// the legacy ValidacijaEDOK: field errors as JSON with status 422), the save (SaveProfaktura, SnimiEDOK) and the content
// of the dialog with the saved profaktura: the header locked, the stavke open (the legacy opens pr_Profakture_DETALJI).
func (h *ProfakturaHandler) SavePro(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.KomProParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateProfaktura(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rdokID, err := h.service.SaveProfaktura(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.proOdgovor(c, rdokID, 0, 0, true, "")
}

// PotvrdaBrisanja asks to confirm the deletion of a profaktura ("Briši" of its row, GET id; the legacy BTN_Brisanje
// "Brisanje dokumenta?").
func (h *ProfakturaHandler) PotvrdaBrisanja(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("id")))
	h.proPitanje(c, proBrisanjeDialogID, "Brisanje", proMsgBrisanjePitanje, "DELETE", fmt.Sprintf("%s/%d", proBrisiURL, rdokID), "#"+proPregledInfoID)
}

// DeletePro deletes the profaktura of the id with its stavke (DELETE /brisi/:id): "Uspešno obrisani podaci..." and the
// list reloaded; a fiscalized or posted one, or one with a nalog, shows why.
func (h *ProfakturaHandler) DeletePro(c *gin.Context) {
	poruka, err := h.service.DeleteProfaktura(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), false, true)
	case poruka != "":
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, h.translator.Message(komsvc.ZahtevObrisanoPoruka), true, false)
	}
}

// PotvrdaPro checks the profaktura before its potvrda and asks to confirm it (GET rdokid: "Potvrda dok.", the legacy
// BTN_Potvrda): one without stavke or already confirmed shows why; else "Da li želite da potvrdite dokument?".
func (h *ProfakturaHandler) PotvrdaPro(c *gin.Context) {
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
	h.proPitanje(c, proPotvrdaDialogID, "Potvrda", proMsgPotvrdaPitanje, "POST", fmt.Sprintf("%s/%d", proPotvrdiURL, rdokID), "#"+proBodyID)
}

// PotvrdiPro confirms the profaktura of the id (POST /potvrdi/:id, the legacy "Da": doksts 20): the message on the main
// screen, whose list is reloaded (the entry dialog with it).
func (h *ProfakturaHandler) PotvrdiPro(c *gin.Context) {
	poruka, err := h.service.PotvrdiProfakturu(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
		return
	}
	komPoruka(c, proPregledInfoID, proPregledInfoDialogID, h.translator.Message(poruka), true, false)
}

// PotvrdaNaloga checks the profaktura before the forming of its nalog and asks to confirm it (GET rdokid: "Form.
// nalog", the legacy BTN_FAKTURA): one with a nalog shows why; else "Formiranje naloga za selektovanu profakturu?".
func (h *ProfakturaHandler) PotvrdaNaloga(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.ProveriNalog(c.Request.Context(), rdokID)
	if err != nil {
		proGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		proGreska(c, h.translator.Message(poruka))
		return
	}
	h.proPitanje(c, proNalogDialogID, "Form. nalog", proMsgNalogPitanje, "POST", fmt.Sprintf("%s/%d", proNalogURL, rdokID), "#"+proBodyID)
}

// FormirajNalog forms the nalog of the profaktura of the id (POST /nalog/:id, the legacy FormirajNALOG): "Uspešno
// formiran nalog.... Broj naloga je: N" (or why it was not formed) on the main screen, whose list is reloaded (the
// entry dialog with it).
func (h *ProfakturaHandler) FormirajNalog(c *gin.Context) {
	broj, poruka, err := h.service.FormirajNalog(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
	case poruka != "":
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, h.translator.MessageWithParams(komsvc.KomProNalogFormiranPoruka,
			map[string]string{"broj": strconv.Itoa(broj)}), true, false)
	}
}

// proPitanje renders a Da/Ne question of the screen (komPitanje).
func (h *ProfakturaHandler) proPitanje(c *gin.Context, dialogID, naslov, pitanje, method, url, target string) {
	komPitanje(c, h.translator, dialogID, naslov, pitanje, method, url, target)
}

// komPitanje renders a Da/Ne question of a screen: "Da" sends the request (method, url) whose answer replaces the
// target, "Ne" closes the dialog.
func komPitanje(c *gin.Context, translator *i18n.Service, dialogID, naslov, pitanje, method, url, target string) {
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
	_ = tmpl.DialogConfirm([]string{translator.Message(pitanje)}, dialog, btnClose, btnDa, btnNe, translator,
		common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// ProKupac returns the kupac of a konto and šifra (komProKupacJSON).
func (h *ProfakturaHandler) ProKupac(c *gin.Context) { komProKupacJSON(c, h.service, h.translator) }

// ProMesto returns a mesto isporuke of the kupac (komProMestoJSON).
func (h *ProfakturaHandler) ProMesto(c *gin.Context) { komProMestoJSON(c, h.service) }

// ProValuta returns the valuta of a šifra with its kurs (komProValutaJSON).
func (h *ProfakturaHandler) ProValuta(c *gin.Context) { komProValutaJSON(c, h.service) }

// ProCenovnik returns the naziv and the opis of a cenovnik (komProCenovnikJSON).
func (h *ProfakturaHandler) ProCenovnik(c *gin.Context) { komProCenovnikJSON(c, h.service) }

// ProArtikal returns the artikal of a šifra with its lots and its cena (komProArtikalJSON).
func (h *ProfakturaHandler) ProArtikal(c *gin.Context) { komProArtikalJSON(c, h.service, h.translator) }

// ProStavka returns a saved stavka for the form of the stavke (komProStavkaJSON).
func (h *ProfakturaHandler) ProStavka(c *gin.Context) { komProStavkaJSON(c, h.service, h.translator) }

// komProJSONService is what the JSON of the header and of the stavke of the profaktura and of its variants needs of
// their service.
type komProJSONService interface {
	GetKupac(ctx context.Context, konto, sifra string) (domain.KomProKupac, bool, error)
	GetMestoIsporuke(ctx context.Context, konto, sifra, mi string) (domain.KomProMestoIsporuke, bool, error)
	GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error)
	GetCenovnik(ctx context.Context, id int64) (naziv, opis string, err error)
	GetArtikal(ctx context.Context, rdokID int64, sifra string, saStanjem bool) (domain.KomProArtikal, bool, error)
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomProStavkaForma, bool, error)
}

// komProKupacJSON returns the kupac of a konto and šifra (GET fkto and fana) as JSON for the header (the legacy exit of
// EDT_FANA: also its ugovoreni rabat and rok plaćanja); found false when it does not exist.
func komProKupacJSON(c *gin.Context, s komProJSONService, translator *i18n.Service) {
	k, found, err := s.GetKupac(c.Request.Context(), c.Query("fkto"), c.Query("fana"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": translator.Message(proMsgKupac)})
		return
	}
	k.SistemPdv = translator.Label(k.SistemPdv)
	c.JSON(http.StatusOK, gin.H{"found": true, "kupac": k})
}

// komProMestoJSON returns a mesto isporuke of the kupac (GET fkto, fana and mi) as JSON for the header (the legacy
// BTN_Sifra3).
func komProMestoJSON(c *gin.Context, s komProJSONService) {
	m, found, err := s.GetMestoIsporuke(c.Request.Context(), c.Query("fkto"), c.Query("fana"), c.Query("mi"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "mesto": m})
}

// komProValutaJSON returns the valuta of a šifra with its kurs on the datum (GET sifval and datum) as JSON for the
// header.
func komProValutaJSON(c *gin.Context, s komProJSONService) {
	valuta, err := s.GetValuta(c.Request.Context(), c.Query("sifval"), c.Query("datum"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, valuta)
}

// komProCenovnikJSON returns the naziv and the opis of a cenovnik (GET id) as JSON for the header (the legacy
// BTN_TraziCENE).
func komProCenovnikJSON(c *gin.Context, s komProJSONService) {
	naziv, opis, err := s.GetCenovnik(c.Request.Context(), int64(common.StringToInt(c.Query("id"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"naziv": naziv, "opis": opis})
}

// komProArtikalJSON returns the artikal of a šifra with its lots and its cena (GET rdokid, sifra and sastanjem) as JSON
// for the form of the stavke (the legacy BTN_Sifra and the exit of EDT_SIFRA); found false when it does not exist.
func komProArtikalJSON(c *gin.Context, s komProJSONService, translator *i18n.Service) {
	artikal, found, err := s.GetArtikal(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))), c.Query("sifra"),
		c.Query("sastanjem") == "true")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": translator.Message(proMsgArtikalNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "artikal": artikal})
}

// komProStavkaJSON returns a saved stavka (GET rdokid and id) as JSON for the form of the stavke ("Izmeni" of its row).
func komProStavkaJSON(c *gin.Context, s komProJSONService, translator *i18n.Service) {
	stavka, found, err := s.GetStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		int64(common.StringToInt(c.Query("id"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": translator.Message(proMsgStavkaNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SaveProStavka saves a stavka of the profaktura (POST of the fields of the stavke, the legacy BTN_DODAJDET; the stavka
// of "Izmeni" changes its količina, cena and rabat): the validation (ValidateStavka, the legacy ValidacijaEPRO: field
// errors as JSON with status 422), the save (SaveStavka, AddEPRO and Preracun) and the content of the dialog.
func (h *ProfakturaHandler) SaveProStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.KomProStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveStavka(ctx, params); err != nil {
		proStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.proOdgovor(c, params.RdokID, 0, 0, true, "")
}

// PotvrdaBrisanjaStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id).
func (h *ProfakturaHandler) PotvrdaBrisanjaStavke(c *gin.Context) {
	h.proPitanje(c, proStavkaBrisanjeID, "Brisanje", proMsgBrisanjeStavke, "DELETE",
		fmt.Sprintf("%s/%d?rdokid=%d", proStavkaBrisiURL, common.StringToInt(c.Query("id")), common.StringToInt(c.Query("rdokid"))),
		"#"+proBodyID)
}

// DeleteProStavka deletes a stavka of the profaktura (DELETE …/stavka/brisi/:id with the rdokid, the legacy
// BTN_BrisiDET) and answers with the content of the dialog.
func (h *ProfakturaHandler) DeleteProStavka(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.DeleteStavka(c.Request.Context(), rdokID, int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		proStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
	case poruka != "":
		proStavkaGreska(c, h.translator.Message(poruka))
	default:
		h.proOdgovor(c, rdokID, 0, 0, true, "")
	}
}

// ProPrikazStavke returns the stavke of the profaktura of the id with the total of their vrednost (the click of a row of
// the main screen, the legacy PR_QRY_EPRO).
func (h *ProfakturaHandler) ProPrikazStavke(c *gin.Context) {
	ctx := c.Request.Context()
	stavke, ino, err := h.service.GetStavke(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, proPregledInfoID, proPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	tbl := h.proTable("", proDetaljiTableID, h.service.PrikazStavkeTableFields())
	var ukupno float64
	tbl.Rows, ukupno = komsvc.ProPrikazStavkeTableRows(stavke, ino)
	porUkupno(&tbl, ukupno)
	_ = tmpl_robno.RobnoFakturePrikazStavke(tbl, h.translator).Render(ctx, c.Writer)
}

// ProStampa prints a profaktura (GET rdokid: "Štampa" of the entry dialog or of the selected profaktura of the main
// screen, the legacy BTN_Print4 with PR_RPT_PROFAKTURA, PR_RPT_PROFAKTURAIZV in a foreign valuta); a profaktura without
// stavke shows why.
func (h *ProfakturaHandler) ProStampa(c *gin.Context) {
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
	naziv := "PROFAKTURA"
	if stampa.Ino {
		naziv = "PROFORMA INVOICE"
	}
	_ = tmpl_rep_kom.ProfakturaStampa(prijReportParams(session, stampa.Firma, naziv), stampa, h.translator).Render(ctx, c.Writer)
}

// AddRoutes registers the routes of the "Profakture" screen.
func (h *ProfakturaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/profakture", h.Profaktura)
	r.GET("/api/komercijalno/profakture/lista", h.ProLista)
	r.GET("/api/komercijalno/profakture/dialog", h.ProDialog)
	r.GET("/api/komercijalno/profakture/novi", h.ProNovi)
	r.POST("/api/komercijalno/profakture/save", h.SavePro)
	r.GET("/api/komercijalno/profakture/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/profakture/brisi/:id", h.DeletePro)
	r.GET("/api/komercijalno/profakture/potvrda", h.PotvrdaPro)
	r.POST("/api/komercijalno/profakture/potvrdi/:id", h.PotvrdiPro)
	r.GET("/api/komercijalno/profakture/nalog/potvrda", h.PotvrdaNaloga)
	r.POST("/api/komercijalno/profakture/nalog/:id", h.FormirajNalog)
	r.GET("/api/komercijalno/profakture/kupac", h.ProKupac)
	r.GET("/api/komercijalno/profakture/mesto", h.ProMesto)
	r.GET("/api/komercijalno/profakture/valuta", h.ProValuta)
	r.GET("/api/komercijalno/profakture/cenovnik", h.ProCenovnik)
	r.GET("/api/komercijalno/profakture/artikal", h.ProArtikal)
	r.GET("/api/komercijalno/profakture/stavka", h.ProStavka)
	r.POST("/api/komercijalno/profakture/stavka/save", h.SaveProStavka)
	r.GET("/api/komercijalno/profakture/stavka/potvrda", h.PotvrdaBrisanjaStavke)
	r.DELETE("/api/komercijalno/profakture/stavka/brisi/:id", h.DeleteProStavka)
	r.GET("/api/komercijalno/profakture/prikaz/stavke/:id", h.ProPrikazStavke)
	r.GET("/api/komercijalno/profakture/stampa", h.ProStampa)
}
