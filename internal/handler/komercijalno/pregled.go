package komercijalno

import (
	"strings"

	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

// The main screens of the documents of the komercijalno (the zahtev za nabavku, the zahtev za ponudu, the porudžbenica
// dobavljaču): what their handlers share. The filters of the list come with every request of the main screen
// (komPregledScriptBody): pvrd, pmagaciniid, pdoksts, psvimagacini and query.

// komPregledFilteri are the filters of the list of a main screen.
type komPregledFilteri struct {
	Vrd         int
	MagaciniID  int
	Doksts      string
	SviMagacini bool
	Tekst       string
}

// komFilteri reads the filters of the list of a main screen (the query of the request, else its form).
func komFilteri(c *gin.Context) komPregledFilteri {
	v := func(k string) string {
		if s, ok := c.GetQuery(k); ok {
			return s
		}
		return c.PostForm(k)
	}
	f := komPregledFilteri{Vrd: common.StringToInt(v("pvrd")), MagaciniID: common.StringToInt(v("pmagaciniid")), Doksts: v("pdoksts"),
		SviMagacini: v("psvimagacini") == "true", Tekst: strings.TrimSpace(v("query"))}
	if f.Doksts != domain.KomZahtevPotvrdjen {
		f.Doksts = domain.KomZahtevNepotvrdjen
	}
	return f
}

// komLista builds the list of the documents of a main screen: a click of a row loads its stavke (stavkeURL + "/:id")
// and selects it for "Štampa"; the nepotvrđeni documents have "Izmeni" (the entry dialog) and "Briši" (the question of
// the deletion, brisiPotvrdaURL with the id), the potvrđeni ones none (they cannot be changed, like the legacy).
func komLista(cfg config.Config, ui domain.KomPregledUI, tableID string, fields []domain.Fields, rows []domain.TableRow, nepotvrdjeni bool,
	brisiPotvrdaURL, stavkeURL string) domain.TableData {
	tbl := common.SetTableBasicData("", tableID, fields, "", ui.ListaURL, 0, 0, 0, 0, cfg)
	common.SetTableConfig(&tbl, "", ui.ListaURL, false, false, false)
	tbl.Rows = rows
	tbl.DetailURL = stavkeURL
	tbl.DetailTarget = "#" + ui.DetaljiID
	tbl.DetailHxRequestType = "GET"
	tbl.DetailHxSwap = "innerHTML"
	tbl.DetailHxTrigger = "click"
	tbl.FuncClick = "komPregledIzbor(this)"
	if nepotvrdjeni && len(rows) > 0 {
		tbl.ShowActions = true
		tbl.BtnUpdate = domain.Button{LabelText: "Izmeni", IsVisible: true, HxRequestType: "GET", HxActionURL: ui.DialogURL}
		tbl.BtnDelete = domain.Button{LabelText: "Briši", IsVisible: true, HxActionURL: brisiPotvrdaURL, HxTarget: "#" + ui.InfoID,
			HxSwap: "innerHTML"}
	}
	return tbl
}

// komPretraga is the search of a main screen, the same as the one of the robno screens (common.CreateSearchInput): the
// list is reloaded while typing (with the filters).
func komPretraga(ui domain.KomPregledUI, translator *i18n.Service) domain.InputControl {
	return common.CreateSearchInput(ui.SearchID, translator, ui.ListaURL, "#"+ui.ListaID, "")
}

// komDodajStampa are "Dodaj" (the entry dialog of a new document of the vrsta and the magacin of the filters) and
// "Štampa" (the print of the selected document, with the values of the fields of polja; disabled until a row is
// selected) of a main screen.
func komDodajStampa(tbl *domain.TableData, ui domain.KomPregledUI, stampaID, stampaURL, polja string) {
	tbl.BtnAdd = domain.Button{LabelText: "Dodaj", IsVisible: true, HxActionURL: ui.DialogURL}
	tbl.BtnPrint = common.SetPrintButton(stampaID, "Štampa", "stampa", stampaURL, "GET", true,
		common.ClassPrintButton+" opacity-50 pointer-events-none", polja)
}

// komPoruka shows a message on the main screen (its container of the messages instead of the target of the request);
// osvezi reloads the list (a document confirmed or deleted).
func komPoruka(c *gin.Context, infoID, dialogID, poruka string, osvezi, greska bool) {
	c.Header("HX-Retarget", "#"+infoID)
	c.Header("HX-Reswap", "innerHTML")
	if osvezi {
		c.Header("HX-Trigger", "komOsvezi")
	}
	utils.RenderDialogOK(c, dialogID, poruka, greska)
}
