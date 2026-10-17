package komercijalno

import (
	"fmt"
	"net/http"
	"strings"

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

// The "Otpremnice" screen of the komercijalno (the legacy pr_Otpremnica, "Otpremnica robe/gotovih proizvoda", opened from
// the menu): its urls and the ids of its elements.
const (
	otpURL = "/api/komercijalno/otpremnice"
	// The main screen: otpListaURL returns the list (the filters of the main screen), otpDialogURL the entry dialog (GET
	// id: a saved one, "Izmeni" of its row; pmagaciniid: a new one, "Dodaj"), otpNoviURL reloads its content with a new
	// one (GET vrd and magaciniid).
	otpListaURL  = otpURL + "/lista"
	otpDialogURL = otpURL + "/dialog"
	otpNoviURL   = otpURL + "/novi"
	// otpSaveURL saves the header (POST); otpBrisiPotvrdaURL asks to confirm the deletion and otpBrisiURL + "/:id" deletes
	// it; otpPotvrdaURL checks the potvrda and asks to confirm it, otpPotvrdiURL + "/:id" confirms it (with the nastanak
	// PDV obaveze and the opis of the faktura).
	otpSaveURL         = otpURL + "/save"
	otpBrisiURL        = otpURL + "/brisi"
	otpBrisiPotvrdaURL = otpBrisiURL + "/potvrda"
	otpPotvrdaURL      = otpURL + "/potvrda"
	otpPotvrdiURL      = otpURL + "/potvrdi"
	// The JSON of the header: otpNalogURL the nalog za isporuku of a new one (GET rdokid and magaciniid), otpValutaURL the
	// valuta of a šifra with its kurs (GET sifval and datum).
	otpNalogURL  = otpURL + "/nalog"
	otpValutaURL = otpURL + "/valuta"
	// The stavke: otpStavkaURL a saved stavka for the form (GET rdokid and id), otpStavkaSaveURL saves it (POST).
	otpStavkaURL     = otpURL + "/stavka"
	otpStavkaSaveURL = otpStavkaURL + "/save"
	// otpPrikazStavkeURL + "/:id" returns the stavke of the selected otpremnica of the list; otpStampaURL prints an
	// otpremnica (GET rdokid and the komercijalni opis).
	otpPrikazStavkeURL = otpURL + "/prikaz/stavke"
	otpStampaURL       = otpURL + "/stampa"

	otpDialogID            = "kom-otp-dialog"
	otpDialogCloseID       = "kom-otp-dialog-close"
	otpNazadID             = "kom-otp-nazad"
	otpBodyID              = "kom-otp-body"
	otpInfoID              = "kom-otp-info"
	otpInfoDialogID        = "kom-otp-info-dialog"
	otpStavkaInfoID        = "kom-otp-stavka-info"
	otpStavkaInfoDialogID  = "kom-otp-stavka-info-dialog"
	otpBrisanjeDialogID    = "kom-otp-brisanje"
	otpPotvrdaDialogID     = "kom-otp-potvrda"
	otpHeaderPanelID       = "kom-otp-header-panel"
	otpStavkePanelID       = "kom-otp-stavke-panel"
	otpStavkeTableID       = "kom-otp-stavke-table"
	otpPregledID           = "kom-otp-pregled"
	otpListaID             = "kom-otp-lista"
	otpDetaljiID           = "kom-otp-detalji"
	otpPregledInfoID       = "kom-otp-pregled-info"
	otpPregledInfoDialogID = "kom-otp-pregled-info-dialog"
	otpPregledRdokID       = "kom-otp-pregled-rdokid"
	otpPregledStampaID     = "kom-otp-pregled-stampa"
	otpPretragaID          = "kom-otp-pretraga"
	otpListaTableID        = "kom-otp-lista-table"
	otpDetaljiTableID      = "kom-otp-detalji-table"
	otpStampaKomercopisID  = "otpstampakomercopis"
	// otpPotvrdaPolja are the fields of the header the potvrda sends: the nastanak PDV obaveze and the opis of the faktura.
	otpPotvrdaPolja       = "#kom-otp-nastanak, #kom-otp-opisdok"
	otpMsgBezOtk          = "Niste izabrali OTK broj za stavke! Unos bez OTK broja?"
	otpMsgNalogNePostoji  = "Niste izabrali narudžbenicu po koju radite otpremnicu!!!"
	otpMsgStavkaNePostoji = "Stavka koju birate za izmenu ne postoji."
)

// otpPutanje are the urls of a screen of the otpremnice: of the "Otpremnice" screen (otpPutanjeOtp) or of the
// "eOtpremnice" screen (eotpPutanje, with the urls of the JSON of the e-otpremnica). They equal the routes of the
// AddRoutes of their screen.
type otpPutanje struct {
	url, lista, dialog, novi, save, brisi, brisiPotvrda, potvrda, potvrdi, nalog, valuta, stavka, stavkaSave, prikazStavke, stampa string
	// The e-otpremnica (the "eOtpremnice" screen): its controls are saved with the header; the JSON of the nalog, of a
	// partner and of a mesto isporuke.
	eotp                         bool
	eotpNalog, partner, mestoIsp string
	// koraci: the entry dialog of the e-otpremnica as a stepper (the test screen "Unos u koracima"); koraciDialog is the
	// url of that dialog, opened from the main screen of the e-otpremnice.
	koraci       bool
	koraciDialog string
}

// otpPutanjeOtp are the urls of the "Otpremnice" screen.
var otpPutanjeOtp = otpPutanje{url: otpURL, lista: otpListaURL, dialog: otpDialogURL, novi: otpNoviURL, save: otpSaveURL, brisi: otpBrisiURL,
	brisiPotvrda: otpBrisiPotvrdaURL, potvrda: otpPotvrdaURL, potvrdi: otpPotvrdiURL, nalog: otpNalogURL, valuta: otpValutaURL,
	stavka: otpStavkaURL, stavkaSave: otpStavkaSaveURL, prikazStavke: otpPrikazStavkeURL, stampa: otpStampaURL}

// OtpremnicaHandler is the handler of the "Otpremnice" screen, and of the "eOtpremnice" screen (EotpremnicaHandler: its
// urls and its service of the e-otpremnica).
type OtpremnicaHandler struct {
	service    komsvc.OtpremnicaService
	eotp       komsvc.EotpremnicaService
	p          otpPutanje
	translator *i18n.Service
	cfg        config.Config
}

// NewOtpremnicaHandler creates the handler of the "Otpremnice" screen.
func NewOtpremnicaHandler(s komsvc.OtpremnicaService, cfg config.Config, translator *i18n.Service) *OtpremnicaHandler {
	return &OtpremnicaHandler{service: s, p: otpPutanjeOtp, translator: translator, cfg: cfg}
}

// otpUI returns the ids and the urls of the entry dialog.
func otpUI(p otpPutanje) domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{HeaderPanelID: otpHeaderPanelID, StavkePanelID: otpStavkePanelID, ContentID: "#" + otpBodyID, DialogID: otpDialogID,
		NoviURL: p.novi, BodyID: otpBodyID, InfoID: otpInfoID, StavkaInfoID: otpStavkaInfoID}
}

// otpKomUI returns the urls of the JSON of the header (and of the controls of the e-otpremnica with their ids).
func otpKomUI(p otpPutanje) domain.KomOtpUI {
	ui := domain.KomOtpUI{NalogURL: p.nalog, ValutaURL: p.valuta}
	if p.eotp {
		ui.Eotp = domain.KomEotpUI{NalogURL: p.eotpNalog, PartnerURL: p.partner, MestoURL: p.mestoIsp, Det1PanelID: eotpDet1PanelID,
			Det2PanelID: eotpDet2PanelID, Det3PanelID: eotpDet3PanelID, Det5PanelID: eotpDet5PanelID, Koraci: p.koraci}
	}
	return ui
}

// otpPregledUI returns the ids and the urls of the main screen.
func otpPregledUI(p otpPutanje) domain.KomPregledUI {
	return domain.KomPregledUI{URL: p.url, ListaURL: p.lista, DialogURL: p.dialog, RootID: otpPregledID, ListaID: otpListaID,
		DetaljiID: otpDetaljiID, InfoID: otpPregledInfoID, RdokID: otpPregledRdokID, SearchID: otpPretragaID}
}

// otpButtons builds the buttons of the entry dialog: "Sačuvaj" saves the header and opens the stavke, "Izmeni" switches
// back to the header, "Nazad" closes the dialog; "Sačuvaj" of the stavka saves the stavka of "Izmeni", "Odustani" gives it
// up. The "Sačuvaj" of the header of the e-otpremnica sends the fields of its controls too.
func otpButtons(p otpPutanje) domain.RobnoFaktureButtons {
	include := "#" + otpHeaderPanelID + " [data-panel-fields]"
	if p.eotp {
		for _, id := range []string{eotpDet1PanelID, eotpDet2PanelID, eotpDet3PanelID, eotpDet5PanelID} {
			include += ", #" + id + " [data-panel-fields]"
		}
	}
	return domain.RobnoFaktureButtons{
		Save: domain.Button{Id: "kom-otp-sacuvaj", LabelText: "Sačuvaj", Icon: "save", BtnClass: common.ClassSaveButton, HxActionURL: p.save,
			HxRequestType: "POST", HxInclude: include, HxTarget: "#" + otpBodyID, HxSwap: "outerHTML",
			HxOnAfterRequest: "robnoFaktureAfterHeaderSave", HxOnAfterRequestArgs: []any{otpHeaderPanelID, otpStavkePanelID}},
		Modify: domain.Button{Id: "kom-otp-izmeni", LabelText: "Izmeni", Icon: "refresh", BtnClass: common.ClassNewButton,
			HxOnClick: "robnoFaktureModifyHeader", HxOnClickArg: []any{otpHeaderPanelID, otpStavkePanelID}},
		Back: domain.Button{Id: otpNazadID, LabelText: "Nazad", Icon: "back", BtnClass: common.ClassCloseButton, HxOnClick: "closeDialog",
			HxOnClickArg: []any{otpDialogID}},
		SaveStavka: domain.Button{Id: "kom-otp-stavka-sacuvaj", LabelText: "Sačuvaj", Icon: "save", BtnClass: common.ClassSaveButton,
			HxActionURL: p.stavkaSave, HxRequestType: "POST", HxInclude: "#" + otpStavkePanelID + " [data-panel-fields]", HxTarget: "#" + otpBodyID,
			HxSwap: "outerHTML", HxOnAfterRequest: "robnoFaktureAfterStavkaSave", HxOnAfterRequestArgs: []any{}},
		CancelStavka: domain.Button{Id: "kom-otp-stavka-odustani", LabelText: "Odustani", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassOdustaniButton + " opacity-50 cursor-not-allowed", HxOnClick: "komOtpStavkaOdustani"},
	}
}

// otpDugmadZa builds "Potvrda dok." (the legacy BTN_POTVRDADOK, with the nastanak PDV obaveze and the opis of the
// faktura) and the print (BTN_Print4, with the komercijalni opis of the main screen); a new otpremnica has them disabled.
func otpDugmadZa(p otpPutanje, rdokID int64) domain.KomProButtons {
	polja := "otpstavkardokid," + otpStampaKomercopisID
	if rdokID <= 0 {
		return domain.KomProButtons{
			Potvrda: domain.Button{Id: "kom-otp-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, IsDisabled: true,
				BtnClass: common.ClassSaveButton + " opacity-50 cursor-not-allowed"},
			Stampa: common.SetPrintButton("kom-otp-stampa", "Štampa", "stampa", p.stampa, "GET", true,
				common.ClassPrintButton+" opacity-50 pointer-events-none", polja),
		}
	}
	return domain.KomProButtons{
		Potvrda: domain.Button{Id: "kom-otp-potvrda-btn", LabelText: "Potvrda dok.", Icon: "obrada", IsVisible: true, BtnClass: common.ClassSaveButton,
			HxActionURL: fmt.Sprintf("%s?rdokid=%d", p.potvrda, rdokID), HxRequestType: "GET", HxInclude: otpPotvrdaPolja, HxTarget: "#" + otpInfoID,
			HxSwap: "innerHTML"},
		Stampa: common.SetPrintButton("kom-otp-stampa", "Štampa", "stampa", p.stampa, "GET", true, common.ClassPrintButton, polja),
	}
}

// otpTable returns an empty grid of the screen.
func (h *OtpremnicaHandler) otpTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", h.p.url, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, h.p.url, false, false, false)
	return tbl
}

// otpEkran is the content of the entry dialog: the otpremnica, the grid of the stavke and the buttons.
type otpEkran struct {
	dok    domain.KomOtpDokument
	tbl    domain.TableData
	btns   domain.RobnoFaktureButtons
	dugmad domain.KomProButtons
}

// otpEkranZa builds the content of the entry dialog with the otpremnica of the id (0: a new one of the vrsta and the
// magacin): the header, the stavke with "Izmeni" of every row (OTK, serija and rok trajanja only in a magacin with the
// lots) and the buttons.
func (h *OtpremnicaHandler) otpEkranZa(c *gin.Context, rdokID int64, vrd, magaciniID int) (otpEkran, error) {
	ctx := c.Request.Context()
	var e otpEkran
	var err error
	if e.dok, err = h.service.GetOtpremnica(ctx, rdokID, vrd, magaciniID); err != nil {
		return e, err
	}
	header := e.dok.Header
	lot := header.Tipzal == 2
	e.tbl = h.otpTable("Količine po nalogu za otpremu", otpStavkeTableID, h.service.StavkeTableFields(lot))
	stavke, err := h.service.GetStavke(ctx, header.RdokID)
	if err != nil {
		return e, err
	}
	e.tbl.Rows = komsvc.OtpStavkeTableRows(stavke, lot, true)
	e.tbl.ShowActions = len(stavke) > 0
	e.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	e.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: h.p.stavka, HxTarget: "this",
		HxSwap: "none", HxOnAfterRequest: "komOtpStavkaUcitaj(event)"}
	e.btns = otpButtons(h.p)
	e.dugmad = otpDugmadZa(h.p, header.RdokID)
	return e, nil
}

// otpInfo renders a message of the entry dialog in its container of the messages.
func (h *OtpremnicaHandler) otpInfo(c *gin.Context, poruka string) {
	dialog := domain.Dialog{Id: otpInfoDialogID, Title: "Info"}
	btnClose := domain.Button{Id: "kom-otp-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnOk := domain.Button{Id: "kom-otp-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
		HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
	_ = tmpl_robno.RobnoFaktureInfo(otpInfoID, h.translator.Message(poruka), dialog, btnClose, btnOk, h.translator).Render(c.Request.Context(), c.Writer)
}

// otpOdgovor renders the content of the entry dialog with the otpremnica of the id (0: a new one of the vrsta and the
// magacin), the header saved (snimljen: the stavke open) or not; an error is shown in the dialog of the screen.
func (h *OtpremnicaHandler) otpOdgovor(c *gin.Context, rdokID int64, vrd, magaciniID int, snimljen bool) {
	e, err := h.otpEkranZa(c, rdokID, vrd, magaciniID)
	if err != nil {
		otpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	e.dok.Header.Snimljen = snimljen && e.dok.Header.RdokID > 0
	if err := tmpl_kom.OtpremnicaBody(otpUI(h.p), otpKomUI(h.p), e.dok, e.tbl, e.btns, e.dugmad, h.translator).Render(c.Request.Context(), c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		h.otpInfo(c, e.dok.Poruka)
	}
}

// otpGreska shows an error in the dialog of the messages of the entry dialog instead of the target of the request.
func otpGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+otpInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, otpInfoDialogID, message, true)
}

// otpStavkaGreska shows an error of the stavke in the container of their dialogs.
func otpStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+otpStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, otpStavkaInfoDialogID, message, true)
}

// otpProveraTekst is the text of a result of the checks of the potvrda: its translated message with its lines.
func (h *OtpremnicaHandler) otpProveraTekst(r komsvc.KomOtpProvera) string {
	return strings.Join(append([]string{h.translator.Message(r.Poruka)}, r.Detalji...), "\n")
}

// Otpremnica opens the "Otpremnice" screen from the menu (the legacy pr_Otpremnica, its tab "Pregled otpremnica" as the
// main screen): the filters (the first magacin of the user, the nepotvrđene otpremnice), "Komercijalni opis u štampi",
// the search with "Dodaj" and "Štampa", the list and the stavke of the selected otpremnica.
func (h *OtpremnicaHandler) Otpremnica(c *gin.Context) {
	ctx := c.Request.Context()
	magacini, err := h.service.GetMagacini(ctx)
	if err != nil {
		utils.RenderDialogOK(c, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	f := komPregledFilteri{Doksts: domain.KomZahtevNepotvrdjen}
	if len(magacini) > 0 {
		f.MagaciniID = common.StringToInt(magacini[0].Key)
	}
	tbl, err := h.otpLista(c, f)
	if err != nil {
		utils.RenderDialogOK(c, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	ui := otpPregledUI(h.p)
	komDodajStampa(&tbl, ui, otpPregledStampaID, h.p.stampa, otpPregledRdokID+","+otpStampaKomercopisID)
	stavkeTbl := h.otpTable("", otpDetaljiTableID, h.service.PrikazStavkeTableFields())
	if err := tmpl_kom.OtpremnicaPregled(ui, otpStampaKomercopisID, magacini, f.MagaciniID, tbl, komPretraga(ui, h.translator), stavkeTbl,
		h.p.eotp, h.p.koraciDialog, h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// otpLista builds the list of the otpremnice of the filters: "Izmeni" and "Briši" of every new one.
func (h *OtpremnicaHandler) otpLista(c *gin.Context, f komPregledFilteri) (domain.TableData, error) {
	otpremnice, err := h.service.GetPrikaz(c.Request.Context(), domain.KomOtpPrikazParams{MagaciniID: f.MagaciniID, Doksts: f.Doksts,
		SviMagacini: f.SviMagacini, Tekst: f.Tekst})
	if err != nil {
		return domain.TableData{}, err
	}
	return komLista(h.cfg, otpPregledUI(h.p), otpListaTableID, h.service.PrikazTableFields(), komsvc.OtpPrikazTableRows(otpremnice, true), true,
		h.p.brisiPotvrda, h.p.prikazStavke), nil
}

// OtpLista returns the list of the otpremnice of the filters of the main screen (GET pmagaciniid, pdoksts, psvimagacini
// and query).
func (h *OtpremnicaHandler) OtpLista(c *gin.Context) {
	tbl, err := h.otpLista(c, komFilteri(c))
	if err != nil {
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	_ = tmpl_kom.KomPregledLista(tbl, h.translator).Render(c.Request.Context(), c.Writer)
}

// OtpDialog opens the entry dialog of an otpremnica: the saved one of the id ("Izmeni" of its row, GET id; a potvrđena or
// a missing one gives a new otpremnica with the message why) or a new one of the magacin of the filters ("Dodaj", GET
// pmagaciniid).
func (h *OtpremnicaHandler) OtpDialog(c *gin.Context) {
	ctx := c.Request.Context()
	f := komFilteri(c)
	e, err := h.otpEkranZa(c, int64(common.StringToInt(c.Query("id"))), 0, f.MagaciniID)
	if err != nil {
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	btnClose := domain.Button{Id: otpDialogCloseID, IsVisible: true, IdDialog: otpDialogID, BtnClass: common.ClassDialogCloseButton}
	if err := tmpl_kom.OtpremnicaDialog(otpUI(h.p), otpKomUI(h.p), e.dok, e.tbl, e.btns, e.dugmad, btnClose, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		h.otpInfo(c, e.dok.Poruka)
	}
}

// OtpNovi reloads the content of the entry dialog with a new otpremnica of the vrsta and the magacin (GET vrd and
// magaciniid: their temporary brojevi).
func (h *OtpremnicaHandler) OtpNovi(c *gin.Context) {
	h.otpOdgovor(c, 0, common.StringToInt(c.Query("vrd")), common.StringToInt(c.Query("magaciniid")), false)
}

// SaveOtp saves the header of the otpremnica ("Sačuvaj", the legacy BTN_Snimi1): the validation (ValidateOtpremnica, the
// legacy ValidacijaEDOK and ValidacijaEPRO: field errors as JSON with status 422), the save (SaveOtpremnica, SnimiEDOK
// with PuniDETALJ) and the content of the dialog with the saved otpremnica: the header locked, the stavke open.
func (h *OtpremnicaHandler) SaveOtp(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.KomOtpParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	var rdokID int64
	var err error
	if h.eotp != nil {
		// The "eOtpremnice" screen: the e-otpremnica of its controls with the header.
		var eparams domain.KomEotpParams
		if err := c.ShouldBind(&eparams); err != nil {
			common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
			return
		}
		if fieldErrors := h.eotp.ValidateEotpremnica(ctx, params, eparams); len(fieldErrors) > 0 {
			common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
			return
		}
		rdokID, err = h.eotp.SaveEotpremnica(ctx, params, eparams)
	} else {
		if fieldErrors := h.service.ValidateOtpremnica(ctx, params); len(fieldErrors) > 0 {
			common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
			return
		}
		rdokID, err = h.service.SaveOtpremnica(ctx, params)
	}
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.otpOdgovor(c, rdokID, 0, 0, true)
}

// PotvrdaBrisanja asks to confirm the deletion of an otpremnica ("Briši" of its row, GET id; the legacy "Brisanje
// dokumenta?").
func (h *OtpremnicaHandler) PotvrdaBrisanja(c *gin.Context) {
	komPitanje(c, h.translator, otpBrisanjeDialogID, "Brisanje", proMsgBrisanjePitanje, "DELETE",
		fmt.Sprintf("%s/%d", h.p.brisi, common.StringToInt(c.Query("id"))), "#"+otpPregledInfoID)
}

// DeleteOtp deletes the otpremnica of the id with its stavke (DELETE /brisi/:id): "Uspešno obrisani podaci..." and the
// list reloaded; a potvrđena or posted one shows why.
func (h *OtpremnicaHandler) DeleteOtp(c *gin.Context) {
	poruka, err := h.service.DeleteOtpremnica(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), false, true)
	case poruka != "":
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, h.translator.Message(komsvc.ZahtevObrisanoPoruka), true, false)
	}
}

// PotvrdaOtp checks the otpremnica before its potvrda and asks to confirm it (GET rdokid, nastanak and opisdok: "Potvrda
// dok.", the legacy BTN_POTVRDADOK): one that cannot be confirmed shows why (with its lines, e.g. the zalihe that are
// missing); else the stavke without an OTK ("Unos bez OTK broja?") and "Da li želite da potvrdite dokument?".
func (h *OtpremnicaHandler) PotvrdaOtp(c *gin.Context) {
	var params domain.KomOtpParams
	if err := c.ShouldBind(&params); err != nil {
		otpGreska(c, common.ErrMsgParseForm)
		return
	}
	r, err := h.service.ProveriPotvrdu(c.Request.Context(), params)
	if err != nil {
		otpGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if r.Poruka != "" {
		otpGreska(c, h.otpProveraTekst(r))
		return
	}
	var pitanje []string
	if len(r.BezOtk) > 0 {
		pitanje = append(append(pitanje, h.translator.Message(otpMsgBezOtk)), r.BezOtk...)
	}
	pitanje = append(pitanje, h.translator.Message(proMsgPotvrdaPitanje))
	dialog := domain.Dialog{Id: otpPotvrdaDialogID, Title: "Potvrda"}
	btnClose := domain.Button{Id: otpPotvrdaDialogID + "-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
	btnDa := domain.Button{Id: otpPotvrdaDialogID + "-da", IsVisible: true, LabelText: "Da", IdDialog: dialog.Id, BtnClass: common.ClassDeleteButton,
		HxRequestType: "POST", HxActionURL: fmt.Sprintf("%s/%d", h.p.potvrdi, params.RdokID), HxInclude: otpPotvrdaPolja, HxTarget: "#" + otpBodyID,
		HxSwap: "outerHTML", HxOnAfterRequest: "closeDialog", HxOnAfterRequestArgs: []any{dialog.Id}}
	btnNe := domain.Button{Id: otpPotvrdaDialogID + "-ne", IsVisible: true, LabelText: "Ne", IdDialog: dialog.Id, BtnClass: common.ClassOdustaniButton}
	_ = tmpl.DialogConfirm(pitanje, dialog, btnClose, btnDa, btnNe, h.translator, common.GetCsrfTokenFromSession(c)).Render(c.Request.Context(), c.Writer)
}

// PotvrdiOtp confirms the otpremnica of the id (POST /potvrdi/:id with nastanak and opisdok, the legacy "Da" with
// SnimiRDOK): the message on the main screen, whose list is reloaded (the entry dialog with it); one that cannot be
// confirmed (any more) shows why in the entry dialog.
func (h *OtpremnicaHandler) PotvrdiOtp(c *gin.Context) {
	var params domain.KomOtpParams
	if err := c.ShouldBind(&params); err != nil {
		otpGreska(c, common.ErrMsgParseForm)
		return
	}
	params.RdokID = int64(common.StringToInt(c.Param("id")))
	r, err := h.service.PotvrdiOtpremnicu(c.Request.Context(), params)
	switch {
	case err != nil:
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
	case !r.Potvrdjena:
		otpGreska(c, h.otpProveraTekst(r))
	default:
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, h.translator.Message(r.Poruka), true, false)
	}
}

// OtpNalog returns a potvrđen nalog za isporuku of the magacin (GET rdokid and magaciniid) as JSON for the header of a new
// otpremnica (the legacy BTN_TraziDOK with PodaciHDR); found false for another one.
func (h *OtpremnicaHandler) OtpNalog(c *gin.Context) {
	n, found, err := h.service.GetNalogOtpremnice(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		common.StringToInt(c.Query("magaciniid")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(otpMsgNalogNePostoji)})
		return
	}
	n.Kupac.SistemPdv = h.translator.Label(n.Kupac.SistemPdv)
	c.JSON(http.StatusOK, gin.H{"found": true, "nalog": n})
}

// OtpValuta returns the valuta of a šifra with its kurs on the datum (GET sifval and datum) as JSON for the header.
func (h *OtpremnicaHandler) OtpValuta(c *gin.Context) {
	valuta, err := h.service.GetValuta(c.Request.Context(), c.Query("sifval"), c.Query("datum"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, valuta)
}

// OtpStavka returns a saved stavka (GET rdokid and id) as JSON for the form of the stavke ("Izmeni" of its row).
func (h *OtpremnicaHandler) OtpStavka(c *gin.Context) {
	stavka, found, err := h.service.GetStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		int64(common.StringToInt(c.Query("id"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(otpMsgStavkaNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SaveOtpStavka saves the stavka of "Izmeni" (POST of the fields of the stavke: its isporučena količina and lot): the
// validation (ValidateStavka: field errors as JSON with status 422), the save (SaveStavka) and the content of the
// dialog.
func (h *OtpremnicaHandler) SaveOtpStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.KomOtpStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveStavka(ctx, params); err != nil {
		otpStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.otpOdgovor(c, params.RdokID, 0, 0, true)
}

// OtpPrikazStavke returns the stavke of the otpremnica of the id (the click of a row of the main screen, the legacy
// TABLE_EPRO2).
func (h *OtpremnicaHandler) OtpPrikazStavke(c *gin.Context) {
	ctx := c.Request.Context()
	stavke, err := h.service.GetStavke(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, otpPregledInfoID, otpPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	tbl := h.otpTable("", otpDetaljiTableID, h.service.PrikazStavkeTableFields())
	tbl.Rows = komsvc.OtpPrikazStavkeTableRows(stavke)
	_ = tmpl_robno.RobnoFakturePrikazStavke(tbl, h.translator).Render(ctx, c.Writer)
}

// OtpStampa prints an otpremnica (GET rdokid and the komercijalni opis: "Štampa" of the entry dialog or of the selected
// otpremnica of the main screen, the legacy BTN_Print4 with PR_RPT_OTPREMNICA); one without stavke shows why.
func (h *OtpremnicaHandler) OtpStampa(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		c.String(http.StatusUnauthorized, common.ErrMsgSessionNotFound)
		return
	}
	stampa, poruka, err := h.service.GetStampa(ctx, int64(common.StringToInt(c.Query("rdokid"))), c.Query(otpStampaKomercopisID) == "true")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		c.String(http.StatusOK, h.translator.Message(poruka))
		return
	}
	_ = tmpl_rep_kom.OtpremnicaStampa(prijReportParams(session, stampa.Firma, "OTPREMNICA"), stampa, h.translator).Render(ctx, c.Writer)
}

// AddRoutes registers the routes of the "Otpremnice" screen.
func (h *OtpremnicaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/otpremnice", h.Otpremnica)
	r.GET("/api/komercijalno/otpremnice/lista", h.OtpLista)
	r.GET("/api/komercijalno/otpremnice/dialog", h.OtpDialog)
	r.GET("/api/komercijalno/otpremnice/novi", h.OtpNovi)
	r.POST("/api/komercijalno/otpremnice/save", h.SaveOtp)
	r.GET("/api/komercijalno/otpremnice/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/otpremnice/brisi/:id", h.DeleteOtp)
	r.GET("/api/komercijalno/otpremnice/potvrda", h.PotvrdaOtp)
	r.POST("/api/komercijalno/otpremnice/potvrdi/:id", h.PotvrdiOtp)
	r.GET("/api/komercijalno/otpremnice/nalog", h.OtpNalog)
	r.GET("/api/komercijalno/otpremnice/valuta", h.OtpValuta)
	r.GET("/api/komercijalno/otpremnice/stavka", h.OtpStavka)
	r.POST("/api/komercijalno/otpremnice/stavka/save", h.SaveOtpStavka)
	r.GET("/api/komercijalno/otpremnice/prikaz/stavke/:id", h.OtpPrikazStavke)
	r.GET("/api/komercijalno/otpremnice/stampa", h.OtpStampa)
}
