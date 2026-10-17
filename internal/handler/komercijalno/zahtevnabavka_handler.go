// Package komercijalno holds the handlers of the komercijalno module.
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

// The "Zahtev za nabavku" screen of the komercijalno (the legacy pr_NarudzbenicaPROIZV, opened from the menu): its urls
// and the ids of its elements (the templates get them by zahtevUI).
const (
	zahtevURL = "/api/komercijalno/zahtevi-nabavka"
	// The main screen: zahtevListaURL returns the list of the zahtevi (the filters of the main screen), zahtevDialogURL
	// the entry dialog of a zahtev (GET id: a saved one, "Izmeni" of its row; pvrd and pmagaciniid: a new one, "Dodaj"),
	// zahtevNoviURL reloads the content of the dialog with a new zahtev (GET vrd and magaciniid: another vrsta).
	zahtevListaURL  = zahtevURL + "/lista"
	zahtevDialogURL = zahtevURL + "/dialog"
	zahtevNoviURL   = zahtevURL + "/novi"
	// zahtevSaveURL saves the header (POST); zahtevBrisiPotvrdaURL asks to confirm the deletion of a zahtev ("Briši" of
	// its row) and zahtevBrisiURL + "/:id" deletes it; zahtevPotvrdaURL asks to confirm the potvrda and
	// zahtevPotvrdiURL + "/:id" confirms it.
	zahtevSaveURL         = zahtevURL + "/save"
	zahtevBrisiURL        = zahtevURL + "/brisi"
	zahtevBrisiPotvrdaURL = zahtevBrisiURL + "/potvrda"
	zahtevPotvrdaURL      = zahtevURL + "/potvrda"
	zahtevPotvrdiURL      = zahtevURL + "/potvrdi"
	// The stavke: zahtevArtikalURL returns the artikal of a šifra (GET sifra), zahtevStavkaURL a saved stavka for the
	// form ("Izmeni" of its row, GET rdokid and id), zahtevStavkaSaveURL saves a stavka (POST),
	// zahtevStavkaPotvrdaURL asks to confirm the deletion of a stavka and zahtevStavkaBrisiURL + "/:id" deletes it.
	zahtevArtikalURL       = zahtevURL + "/artikal"
	zahtevStavkaURL        = zahtevURL + "/stavka"
	zahtevStavkaSaveURL    = zahtevStavkaURL + "/save"
	zahtevStavkaPotvrdaURL = zahtevStavkaURL + "/potvrda"
	zahtevStavkaBrisiURL   = zahtevStavkaURL + "/brisi"
	// zahtevPrikazStavkeURL + "/:id" returns the stavke of the selected zahtev of the list, zahtevStampaURL prints a
	// zahtev (GET rdokid).
	zahtevPrikazStavkeURL = zahtevURL + "/prikaz/stavke"
	zahtevStampaURL       = zahtevURL + "/stampa"

	zahtevDialogID             = "kom-zahtev-dialog"
	zahtevDialogCloseID        = "kom-zahtev-dialog-close"
	zahtevNazadID              = "kom-zahtev-nazad"
	zahtevBodyID               = "kom-zahtev-body"
	zahtevInfoID               = "kom-zahtev-info"
	zahtevInfoDialogID         = "kom-zahtev-info-dialog"
	zahtevStavkaInfoID         = "kom-zahtev-stavka-info"
	zahtevStavkaInfoDialogID   = "kom-zahtev-stavka-info-dialog"
	zahtevPotvrdaDialogID      = "kom-zahtev-potvrda"
	zahtevBrisanjeDialogID     = "kom-zahtev-brisanje"
	zahtevStavkaBrisanjeID     = "kom-zahtev-stavka-brisanje"
	zahtevHeaderPanelID        = "kom-zahtev-header-panel"
	zahtevStavkePanelID        = "kom-zahtev-stavke-panel"
	zahtevStavkeTableID        = "kom-zahtev-stavke-table"
	zahtevPregledID            = "kom-zahtev-pregled"
	zahtevListaID              = "kom-zahtev-lista"
	zahtevDetaljiID            = "kom-zahtev-detalji"
	zahtevPregledInfoID        = "kom-zahtev-pregled-info"
	zahtevPregledInfoDialogID  = "kom-zahtev-pregled-info-dialog"
	zahtevPregledRdokID        = "kom-zahtev-pregled-rdokid"
	zahtevPregledStampaID      = "kom-zahtev-pregled-stampa"
	zahtevPretragaID           = "kom-zahtev-pretraga"
	zahtevListaTableID         = "kom-zahtev-lista-table"
	zahtevDetaljiTableID       = "kom-zahtev-detalji-table"
	zahtevMsgPotvrdaPitanje    = "Da li želite da potvrdite dokument?"
	zahtevMsgBrisanjePitanje   = "Brisanje dokumenta?"
	zahtevMsgBrisanjeStavkePit = "Brisanje podataka? Da li ste sigurni?"
	zahtevMsgStavkaNePostoji   = "Stavka koju birate za izmenu ne postoji."
	zahtevMsgArtikalNePostoji  = "Artikal sa izabranom šifrom ne postoji!!!"
)

// ZahtevNabavkaHandler is the handler of the "Zahtev za nabavku" screen.
type ZahtevNabavkaHandler struct {
	service    komsvc.ZahtevNabavkaService
	translator *i18n.Service
	cfg        config.Config
}

// NewZahtevNabavkaHandler creates the handler of the "Zahtev za nabavku" screen.
func NewZahtevNabavkaHandler(s komsvc.ZahtevNabavkaService, cfg config.Config, translator *i18n.Service) *ZahtevNabavkaHandler {
	return &ZahtevNabavkaHandler{service: s, translator: translator, cfg: cfg}
}

// zahtevUI returns the ids and the urls of the entry dialog.
func zahtevUI() domain.RobnoFaktureUI {
	return domain.RobnoFaktureUI{
		HeaderPanelID: zahtevHeaderPanelID,
		StavkePanelID: zahtevStavkePanelID,
		ContentID:     "#" + zahtevBodyID,
		DialogID:      zahtevDialogID,
		NoviURL:       zahtevNoviURL,
		ArtikalURL:    zahtevArtikalURL,
		BodyID:        zahtevBodyID,
		InfoID:        zahtevInfoID,
		StavkaInfoID:  zahtevStavkaInfoID,
	}
}

// zahtevPregledUI returns the ids and the urls of the main screen.
func zahtevPregledUI() domain.KomPregledUI {
	return domain.KomPregledUI{URL: zahtevURL, ListaURL: zahtevListaURL, DialogURL: zahtevDialogURL, RootID: zahtevPregledID,
		ListaID: zahtevListaID, DetaljiID: zahtevDetaljiID, InfoID: zahtevPregledInfoID, RdokID: zahtevPregledRdokID,
		SearchID: zahtevPretragaID}
}

// zahtevButtons builds the buttons of the entry dialog of a zahtev, like the ones of the faktura: "Sačuvaj" saves the
// header and opens the stavke, "Izmeni" switches back to the header, "Nazad" closes the dialog (the list of the main
// screen is reloaded); "Sačuvaj" of the stavka saves it (a new one, or the one of "Izmeni" of its row), "Odustani" gives
// up the change of a stavka. The potvrda is zahtevPotvrdaButton; a new zahtev and the deletion are "Dodaj" and "Briši"
// of the main screen.
func zahtevButtons() domain.RobnoFaktureButtons {
	return domain.RobnoFaktureButtons{
		Save: domain.Button{
			Id:                   "kom-zahtev-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          zahtevSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + zahtevHeaderPanelID + " [data-panel-fields]",
			HxTarget:             "#" + zahtevBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterHeaderSave",
			HxOnAfterRequestArgs: []any{zahtevHeaderPanelID, zahtevStavkePanelID},
		},
		Modify: domain.Button{
			Id:           "kom-zahtev-izmeni",
			LabelText:    "Izmeni",
			Icon:         "refresh",
			BtnClass:     common.ClassNewButton,
			HxOnClick:    "robnoFaktureModifyHeader",
			HxOnClickArg: []any{zahtevHeaderPanelID, zahtevStavkePanelID},
		},
		Back: domain.Button{
			Id:           zahtevNazadID,
			LabelText:    "Nazad",
			Icon:         "back",
			BtnClass:     common.ClassCloseButton,
			HxOnClick:    "closeDialog",
			HxOnClickArg: []any{zahtevDialogID},
		},
		SaveStavka: domain.Button{
			Id:                   "kom-zahtev-stavka-sacuvaj",
			LabelText:            "Sačuvaj",
			Icon:                 "save",
			BtnClass:             common.ClassSaveButton,
			HxActionURL:          zahtevStavkaSaveURL,
			HxRequestType:        "POST",
			HxInclude:            "#" + zahtevStavkePanelID + " [data-panel-fields]",
			HxTarget:             "#" + zahtevBodyID,
			HxSwap:               "outerHTML",
			HxOnAfterRequest:     "robnoFaktureAfterStavkaSave",
			HxOnAfterRequestArgs: []any{},
		},
		CancelStavka: domain.Button{
			Id:         "kom-zahtev-stavka-odustani",
			LabelText:  "Odustani",
			IsVisible:  true,
			IsDisabled: true,
			BtnClass:   common.ClassOdustaniButton + " opacity-50 cursor-not-allowed",
			HxOnClick:  "komZahtevStavkaOdustani",
		},
	}
}

// zahtevPotvrdaButton is "Potvrda" of the header (the legacy BTN_Potvrda): it asks to confirm the saved zahtev; a new
// one has nothing to confirm.
func zahtevPotvrdaButton(rdokID int64) domain.Button {
	if rdokID <= 0 {
		return domain.Button{Id: "kom-zahtev-potvrda", LabelText: "Potvrda", Icon: "obrada", IsVisible: true, IsDisabled: true,
			BtnClass: common.ClassSaveButton + " opacity-50 cursor-not-allowed"}
	}
	return domain.Button{
		Id:            "kom-zahtev-potvrda",
		LabelText:     "Potvrda",
		Icon:          "obrada",
		BtnClass:      common.ClassSaveButton,
		HxActionURL:   fmt.Sprintf("%s?rdokid=%d", zahtevPotvrdaURL, rdokID),
		HxRequestType: "GET",
		HxTarget:      "#" + zahtevInfoID,
		HxSwap:        "innerHTML",
	}
}

// zahtevStampaButton is "Štampa" of the entry dialog (the print of the zahtev, like "Štampa" of the main screen): always
// shown, disabled until the zahtev is saved.
func zahtevStampaButton(rdokID int64) domain.Button {
	class := common.ClassPrintButton
	if rdokID <= 0 {
		class += " opacity-50 pointer-events-none"
	}
	return common.SetPrintButton("kom-zahtev-stampa", "Štampa", "stampa", zahtevStampaURL, "GET", true, class, "stavkardokid")
}

// zahtevTable returns an empty grid of the screen.
func (h *ZahtevNabavkaHandler) zahtevTable(title, tableID string, fields []domain.Fields) domain.TableData {
	tbl := common.SetTableBasicData(title, tableID, fields, "", zahtevURL, 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, title, zahtevURL, false, false, false)
	return tbl
}

// zahtevEkran is the content of the tab "Zahtev za nabavku": the zahtev, the grid of its stavke and the buttons.
type zahtevEkran struct {
	dok        domain.KomZahtevDokument
	tbl        domain.TableData
	btns       domain.RobnoFaktureButtons
	btnPotvrda domain.Button
}

// zahtevEkranZa builds the content of the tab with the zahtev of the id (0: a new one of the vrsta and the magacin):
// the header, the stavke with "Izmeni" and "Obriši" of every row and the buttons.
func (h *ZahtevNabavkaHandler) zahtevEkranZa(c *gin.Context, rdokID int64, vrd, magaciniID int) (zahtevEkran, error) {
	ctx := c.Request.Context()
	var e zahtevEkran
	var err error
	if e.dok, err = h.service.GetZahtev(ctx, rdokID, vrd, magaciniID); err != nil {
		return e, err
	}
	header := e.dok.Header
	e.tbl = h.zahtevTable("Stavke zahteva", zahtevStavkeTableID, h.service.StavkeTableFields())
	stavke, err := h.service.GetStavke(ctx, header.RdokID)
	if err != nil {
		return e, err
	}
	e.tbl.Rows = komsvc.ZahtevStavkeTableRows(stavke)
	e.tbl.ShowActions = len(stavke) > 0
	e.tbl.HxVals = fmt.Sprintf(`{"rdokid": "%d"}`, header.RdokID)
	e.tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: zahtevStavkaURL,
		HxTarget: "this", HxSwap: "none", HxOnAfterRequest: "komZahtevStavkaUcitaj(event)"}
	e.tbl.BtnDelete = domain.Button{LabelText: "Obriši", IsVisible: true, HxActionURL: zahtevStavkaPotvrdaURL,
		HxTarget: "#" + zahtevStavkaInfoID, HxSwap: "innerHTML"}
	e.btns = zahtevButtons()
	e.btns.Print = zahtevStampaButton(header.RdokID)
	e.btnPotvrda = zahtevPotvrdaButton(header.RdokID)
	return e, nil
}

// zahtevBody renders the content of the tab (the answer of the save, of a selected or new zahtev, of the deletion and
// of the potvrda) with the message of the screen (poruka, else the one of the zahtev) in its dialog.
func (h *ZahtevNabavkaHandler) zahtevBody(c *gin.Context, e zahtevEkran, poruka string) {
	ctx := c.Request.Context()
	if err := tmpl_kom.ZahtevNabavkaBody(zahtevUI(), e.dok, e.tbl, e.btns, e.btnPotvrda, h.translator).Render(ctx, c.Writer); err != nil {
		return
	}
	if poruka == "" {
		poruka = e.dok.Poruka
	}
	if poruka != "" {
		dialog := domain.Dialog{Id: zahtevInfoDialogID, Title: "Info"}
		btnClose := domain.Button{Id: "kom-zahtev-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "kom-zahtev-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(zahtevInfoID, h.translator.Message(poruka), dialog, btnClose, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// zahtevOdgovor renders the content of the tab with the zahtev of the id (0: a new one of the vrsta and the magacin),
// the header saved (snimljen: the stavke open) or not, with a message; an error is shown in the dialog of the screen.
func (h *ZahtevNabavkaHandler) zahtevOdgovor(c *gin.Context, rdokID int64, vrd, magaciniID int, snimljen bool, poruka string) {
	e, err := h.zahtevEkranZa(c, rdokID, vrd, magaciniID)
	if err != nil {
		zahtevGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	e.dok.Header.Snimljen = snimljen && e.dok.Header.RdokID > 0
	h.zahtevBody(c, e, poruka)
}

// zahtevGreska shows an error in the dialog of the messages of the screen instead of the target of the request.
func zahtevGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+zahtevInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, zahtevInfoDialogID, message, true)
}

// zahtevStavkaGreska shows an error of the stavke in the container of their dialogs.
func zahtevStavkaGreska(c *gin.Context, message string) {
	c.Header("HX-Retarget", "#"+zahtevStavkaInfoID)
	c.Header("HX-Reswap", "innerHTML")
	utils.RenderDialogOK(c, zahtevStavkaInfoDialogID, message, true)
}

// ZahtevNabavka opens the "Zahtev za nabavku" screen from the menu (the legacy pr_NarudzbenicaPROIZV): the main screen
// with the filters (the first vrsta and magacin, the nepotvrđeni zahtevi), the search with "Dodaj" and "Štampa", the
// list of the zahtevi and the stavke of the selected one.
func (h *ZahtevNabavkaHandler) ZahtevNabavka(c *gin.Context) {
	ctx := c.Request.Context()
	vrste, err := h.service.GetVrste(ctx)
	if err != nil {
		utils.RenderDialogOK(c, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	magacini, err := h.service.GetMagacini(ctx)
	if err != nil {
		utils.RenderDialogOK(c, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	f := komPregledFilteri{Doksts: domain.KomZahtevNepotvrdjen}
	if len(vrste) > 0 {
		f.Vrd = common.StringToInt(vrste[0].Key)
	}
	if len(magacini) > 0 {
		f.MagaciniID = common.StringToInt(magacini[0].Key)
	}
	tbl, err := h.zahtevLista(c, f)
	if err != nil {
		utils.RenderDialogOK(c, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
		return
	}
	ui := zahtevPregledUI()
	komDodajStampa(&tbl, ui, zahtevPregledStampaID, zahtevStampaURL, zahtevPregledRdokID)
	stavkeTbl := h.zahtevTable("", zahtevDetaljiTableID, h.service.PrikazStavkeTableFields())
	if err := tmpl_kom.ZahtevNabavkaPregled(ui, vrste, magacini, f.Vrd, f.MagaciniID, tbl, komPretraga(ui, h.translator), stavkeTbl,
		h.translator).Render(ctx, c.Writer); err != nil {
		utils.RenderDialogOK(c, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), true)
	}
}

// zahtevLista builds the list of the zahtevi of the filters (the legacy PuniHDRZAHTEV): "Izmeni" and "Briši" of every
// nepotvrđen zahtev.
func (h *ZahtevNabavkaHandler) zahtevLista(c *gin.Context, f komPregledFilteri) (domain.TableData, error) {
	zahtevi, err := h.service.GetPrikaz(c.Request.Context(), domain.KomZahtevPrikazParams{Vrd: f.Vrd, MagaciniID: f.MagaciniID, Doksts: f.Doksts,
		SviMagacini: f.SviMagacini, Tekst: f.Tekst})
	if err != nil {
		return domain.TableData{}, err
	}
	return komLista(h.cfg, zahtevPregledUI(), zahtevListaTableID, h.service.PrikazTableFields(), komsvc.ZahtevPrikazTableRows(zahtevi),
		f.Doksts == domain.KomZahtevNepotvrdjen, zahtevBrisiPotvrdaURL, zahtevPrikazStavkeURL), nil
}

// ZahtevLista returns the list of the zahtevi of the filters of the main screen (GET pvrd, pmagaciniid, pdoksts,
// psvimagacini and query: another filter or the search).
func (h *ZahtevNabavkaHandler) ZahtevLista(c *gin.Context) {
	tbl, err := h.zahtevLista(c, komFilteri(c))
	if err != nil {
		komPoruka(c, zahtevPregledInfoID, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	_ = tmpl_kom.KomPregledLista(tbl, h.translator).Render(c.Request.Context(), c.Writer)
}

// ZahtevDialog opens the entry dialog of a zahtev: the saved one of the id ("Izmeni" of its row, GET id; a confirmed
// or a missing one gives a new zahtev with the message why) or a new one of the vrsta and the magacin of the filters
// ("Dodaj", GET pvrd and pmagaciniid).
func (h *ZahtevNabavkaHandler) ZahtevDialog(c *gin.Context) {
	ctx := c.Request.Context()
	f := komFilteri(c)
	e, err := h.zahtevEkranZa(c, int64(common.StringToInt(c.Query("id"))), f.Vrd, f.MagaciniID)
	if err != nil {
		komPoruka(c, zahtevPregledInfoID, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()), false, true)
		return
	}
	btnClose := domain.Button{Id: zahtevDialogCloseID, IsVisible: true, IdDialog: zahtevDialogID, BtnClass: common.ClassDialogCloseButton}
	if err := tmpl_kom.ZahtevNabavkaDialog(zahtevUI(), e.dok, e.tbl, e.btns, e.btnPotvrda, btnClose, h.translator,
		common.GetCsrfTokenFromSession(c)).Render(ctx, c.Writer); err != nil {
		return
	}
	if e.dok.Poruka != "" {
		dialog := domain.Dialog{Id: zahtevInfoDialogID, Title: "Info"}
		btnX := domain.Button{Id: "kom-zahtev-info-close", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassDialogCloseButton}
		btnOk := domain.Button{Id: "kom-zahtev-info-ok", LabelText: "OK", IsVisible: true, IdDialog: dialog.Id, BtnClass: common.ClassSaveButton,
			HxOnClick: "closeDialog", HxOnClickArg: []any{dialog.Id}}
		_ = tmpl_robno.RobnoFaktureInfo(zahtevInfoID, h.translator.Message(e.dok.Poruka), dialog, btnX, btnOk, h.translator).Render(ctx, c.Writer)
	}
}

// ZahtevNovi reloads the content of the entry dialog with a new zahtev of the vrsta and the magacin (GET vrd and
// magaciniid: another vrsta of a new zahtev, its next broj).
func (h *ZahtevNabavkaHandler) ZahtevNovi(c *gin.Context) {
	h.zahtevOdgovor(c, 0, common.StringToInt(c.Query("vrd")), common.StringToInt(c.Query("magaciniid")), false, "")
}

// SaveZahtev saves the header of the zahtev ("Sačuvaj", the legacy BTN_Snimi1): the validation (ValidateZahtev, the
// legacy ValidacijaEDOK: field errors as JSON with status 422), the save (SaveZahtev, SnimiEDOK) and the content of
// the tab with the saved zahtev: the header locked, the stavke open.
func (h *ZahtevNabavkaHandler) SaveZahtev(c *gin.Context) {
	ctx := c.Request.Context()
	if domain.GetSessionFromStdContext(ctx) == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgSessionNotFound)
		return
	}
	var params domain.KomZahtevParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateZahtev(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	rdokID, err := h.service.SaveZahtev(ctx, params)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.zahtevOdgovor(c, rdokID, 0, 0, true, "")
}

// PotvrdaBrisanja asks to confirm the deletion of a zahtev ("Briši" of its row, GET id; the legacy BTN_Brisanje
// "Brisanje dokumenta?"): "Da" deletes it, "Ne" closes the question.
func (h *ZahtevNabavkaHandler) PotvrdaBrisanja(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("id")))
	h.zahtevPitanje(c, zahtevBrisanjeDialogID, "Brisanje", zahtevMsgBrisanjePitanje, "DELETE", fmt.Sprintf("%s/%d", zahtevBrisiURL, rdokID),
		"#"+zahtevPregledInfoID)
}

// DeleteZahtev deletes the zahtev of the id with its stavke (DELETE /brisi/:id): "Uspešno obrisani podaci..." and the
// list reloaded; a fiscalized or posted one shows why.
func (h *ZahtevNabavkaHandler) DeleteZahtev(c *gin.Context) {
	poruka, err := h.service.DeleteZahtev(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	switch {
	case err != nil:
		komPoruka(c, zahtevPregledInfoID, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()), false, true)
	case poruka != "":
		komPoruka(c, zahtevPregledInfoID, zahtevPregledInfoDialogID, h.translator.Message(poruka), false, true)
	default:
		komPoruka(c, zahtevPregledInfoID, zahtevPregledInfoDialogID, h.translator.Message(komsvc.ZahtevObrisanoPoruka), true, false)
	}
}

// PotvrdaZahteva checks the zahtev before its potvrda and asks to confirm it (GET rdokid: "Potvrda", the legacy
// BTN_Potvrda): a zahtev without stavke or already confirmed shows why; else "Da li želite da potvrdite dokument?".
func (h *ZahtevNabavkaHandler) PotvrdaZahteva(c *gin.Context) {
	ctx := c.Request.Context()
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	poruka, err := h.service.ProveriPotvrdu(ctx, rdokID)
	if err != nil {
		zahtevGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	if poruka != "" {
		zahtevGreska(c, h.translator.Message(poruka))
		return
	}
	h.zahtevPitanje(c, zahtevPotvrdaDialogID, "Potvrda", zahtevMsgPotvrdaPitanje, "POST", fmt.Sprintf("%s/%d", zahtevPotvrdiURL, rdokID),
		"#"+zahtevBodyID)
}

// PotvrdiZahtev confirms the zahtev of the id (POST /potvrdi/:id, the legacy "Da" of BTN_Potvrda: doksts 40): the
// message ("Dokument je potvrđen!!!" or why it was not confirmed) on the main screen, whose list is reloaded (the entry
// dialog with it).
func (h *ZahtevNabavkaHandler) PotvrdiZahtev(c *gin.Context) {
	poruka, err := h.service.PotvrdiZahtev(c.Request.Context(), int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		komPoruka(c, zahtevPregledInfoID, zahtevPregledInfoDialogID, fmt.Sprintf(common.ErrMsgSaveData+": %s", err.Error()), false, true)
		return
	}
	komPoruka(c, zahtevPregledInfoID, zahtevPregledInfoDialogID, h.translator.Message(poruka), true, false)
}

// zahtevPitanje renders a Da/Ne question of the screen: "Da" sends the request (method, url) whose answer replaces the
// target, "Ne" closes the dialog.
func (h *ZahtevNabavkaHandler) zahtevPitanje(c *gin.Context, dialogID, naslov, pitanje, method, url, target string) {
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

// ZahtevArtikal returns the artikal of a šifra with its nabavna cena (GET sifra) as JSON for the form of the stavke (the
// legacy exit of EDT_SIFRA); found false when it does not exist.
func (h *ZahtevNabavkaHandler) ZahtevArtikal(c *gin.Context) {
	artikal, found, err := h.service.GetArtikal(c.Request.Context(), c.Query("sifra"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(zahtevMsgArtikalNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "artikal": artikal})
}

// ZahtevStavka returns a saved stavka (GET rdokid and id) as JSON for the form of the stavke ("Izmeni" of its row).
func (h *ZahtevNabavkaHandler) ZahtevStavka(c *gin.Context) {
	stavka, found, err := h.service.GetStavkaForma(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))),
		int64(common.StringToInt(c.Query("id"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(zahtevMsgStavkaNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "stavka": stavka})
}

// SaveZahtevStavka saves a stavka of the zahtev (POST of the fields of the stavke, the legacy BTN_DODAJDET; a stavka of
// "Izmeni" as a change): the validation (ValidateStavka, the legacy ValidacijaEPRO: field errors as JSON with status
// 422), the save (SaveStavka, AddEPRO) and the content of the tab with the zahtev (the stavke open).
func (h *ZahtevNabavkaHandler) SaveZahtevStavka(c *gin.Context) {
	ctx := c.Request.Context()
	var params domain.KomZahtevStavkaParams
	if err := c.ShouldBind(&params); err != nil {
		common.WriteJSONResponse(c, http.StatusBadRequest, false, nil, common.ErrMsgParseForm)
		return
	}
	if fieldErrors := h.service.ValidateStavka(ctx, params); len(fieldErrors) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldErrors, common.ErrMsgValidation)
		return
	}
	if err := h.service.SaveStavka(ctx, params); err != nil {
		zahtevStavkaGreska(c, common.ErrMsgSaveData+": "+err.Error())
		return
	}
	h.zahtevOdgovor(c, params.RdokID, 0, 0, true, "")
}

// PotvrdaBrisanjaStavke asks to confirm the deletion of a stavka ("Obriši" of its row, GET rdokid and id, the legacy
// "Brisanje podataka? Da li ste sigurni?"): "Da" deletes it.
func (h *ZahtevNabavkaHandler) PotvrdaBrisanjaStavke(c *gin.Context) {
	h.zahtevPitanje(c, zahtevStavkaBrisanjeID, "Brisanje", zahtevMsgBrisanjeStavkePit, "DELETE",
		fmt.Sprintf("%s/%d?rdokid=%d", zahtevStavkaBrisiURL, common.StringToInt(c.Query("id")), common.StringToInt(c.Query("rdokid"))),
		"#"+zahtevBodyID)
}

// DeleteZahtevStavka deletes a stavka of the zahtev (DELETE …/stavka/brisi/:id with the rdokid, the legacy
// BTN_BrisiDET: the redni brojevi after it move back) and answers with the content of the tab with the zahtev.
func (h *ZahtevNabavkaHandler) DeleteZahtevStavka(c *gin.Context) {
	rdokID := int64(common.StringToInt(c.Query("rdokid")))
	if err := h.service.DeleteStavka(c.Request.Context(), rdokID, int64(common.StringToInt(c.Param("id")))); err != nil {
		zahtevStavkaGreska(c, fmt.Sprintf(common.ErrMsgDeleteData, err.Error()))
		return
	}
	h.zahtevOdgovor(c, rdokID, 0, 0, true, "")
}

// ZahtevPrikazStavke returns the stavke of the zahtev of the id (the click of a row of the list of the main screen).
func (h *ZahtevNabavkaHandler) ZahtevPrikazStavke(c *gin.Context) {
	ctx := c.Request.Context()
	stavke, err := h.service.GetStavke(ctx, int64(common.StringToInt(c.Param("id"))))
	if err != nil {
		zahtevGreska(c, fmt.Sprintf(common.ErrMsgDataFetch, err.Error()))
		return
	}
	tbl := h.zahtevTable("", zahtevDetaljiTableID, h.service.PrikazStavkeTableFields())
	tbl.Rows = komsvc.ZahtevPrikazStavkeTableRows(stavke)
	_ = tmpl_robno.RobnoFakturePrikazStavke(tbl, h.translator).Render(ctx, c.Writer)
}

// ZahtevStampa prints a zahtev za nabavku (GET rdokid: "Štampaj" of the selected zahtev of the tab "Pregled zahteva za
// nabavku", the legacy BTN_Print4 with PR_RPT_ZAHTEV_NABAVKA); a zahtev without stavke shows why.
func (h *ZahtevNabavkaHandler) ZahtevStampa(c *gin.Context) {
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
		ReportName:  "ZAHTEV ZA NABAVKU",
		God:         session.SelectedGod,
	}
	_ = tmpl_rep_kom.ZahtevNabavkaStampa(repParams, stampa, h.translator).Render(ctx, c.Writer)
}

// AddRoutes registers the routes of the "Zahtev za nabavku" screen.
func (h *ZahtevNabavkaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/zahtevi-nabavka", h.ZahtevNabavka)
	r.GET("/api/komercijalno/zahtevi-nabavka/lista", h.ZahtevLista)
	r.GET("/api/komercijalno/zahtevi-nabavka/dialog", h.ZahtevDialog)
	r.GET("/api/komercijalno/zahtevi-nabavka/novi", h.ZahtevNovi)
	r.POST("/api/komercijalno/zahtevi-nabavka/save", h.SaveZahtev)
	r.GET("/api/komercijalno/zahtevi-nabavka/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/zahtevi-nabavka/brisi/:id", h.DeleteZahtev)
	r.GET("/api/komercijalno/zahtevi-nabavka/potvrda", h.PotvrdaZahteva)
	r.POST("/api/komercijalno/zahtevi-nabavka/potvrdi/:id", h.PotvrdiZahtev)
	r.GET("/api/komercijalno/zahtevi-nabavka/artikal", h.ZahtevArtikal)
	r.GET("/api/komercijalno/zahtevi-nabavka/stavka", h.ZahtevStavka)
	r.POST("/api/komercijalno/zahtevi-nabavka/stavka/save", h.SaveZahtevStavka)
	r.GET("/api/komercijalno/zahtevi-nabavka/stavka/potvrda", h.PotvrdaBrisanjaStavke)
	r.DELETE("/api/komercijalno/zahtevi-nabavka/stavka/brisi/:id", h.DeleteZahtevStavka)
	r.GET("/api/komercijalno/zahtevi-nabavka/prikaz/stavke/:id", h.ZahtevPrikazStavke)
	r.GET("/api/komercijalno/zahtevi-nabavka/stampa", h.ZahtevStampa)
}
