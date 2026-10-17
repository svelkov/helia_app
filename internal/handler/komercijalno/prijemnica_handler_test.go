package komercijalno

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	tmpl_kom "helia/frontend/templates/komercijalno"
	tmpl_rep_kom "helia/frontend/templates/reports/komercijalno"
	"helia/i18n"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestPrijRoutes checks that the urls the screen sends its requests to (the constants of the buttons, of the grids and
// of the script) are routes of AddRoutes.
func TestPrijRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&PrijemnicaHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, prijURL},
		{http.MethodGet, prijListaURL},
		{http.MethodGet, prijDialogURL},
		{http.MethodGet, prijNoviURL},
		{http.MethodPost, prijSaveURL},
		{http.MethodGet, prijBrisiPotvrdaURL},
		{http.MethodDelete, prijBrisiURL + "/:id"},
		{http.MethodGet, prijPotvrdaURL},
		{http.MethodPost, prijPotvrdiURL + "/:id"},
		{http.MethodGet, prijArtikalURL},
		{http.MethodGet, prijStavkaURL},
		{http.MethodPost, prijStavkaSaveURL},
		{http.MethodGet, prijStavkaPotvrdaURL},
		{http.MethodDelete, prijStavkaBrisiURL + "/:id"},
		{http.MethodPost, prijPreuzmiURL},
		{http.MethodGet, prijVratiPotvrdaURL},
		{http.MethodDelete, prijVratiURL + "/:id"},
		{http.MethodGet, prijPakovanjaURL},
		{http.MethodPost, prijPakovanjeSaveURL},
		{http.MethodGet, prijPrikazStavkeURL + "/:id"},
		{http.MethodGet, prijStampaURL},
		{http.MethodGet, prijKartonURL},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestPrijBodyRender checks the content of the entry dialog: the two collapsible controls (a new prijemnica: the header
// open, the stavke closed; a saved one the other way round), the fields the service binds, the vrsta, the magacin and
// the porudžbenica fixed for a saved prijemnica, the stavke of the porudžbenica only with one, "Pakovanje" and "Karton
// identifikacije" only for a saved one, and "STORNO DOKUMENT" only for a storno.
func TestPrijBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	fields := []string{`name="vrd"`, `name="magaciniid"`, `id="brdok"`, `name="datum"`, `name="vrdokid"`, `name="rdokid"`, `name="rproid"`,
		`name="konto"`, `name="sifraartikla"`, `name="trazkolic"`, `name="kolic"`, `name="serija"`, `name="otk"`, `id="kom-prij-potvrda-btn"`,
		`id="kom-prij-stavka-odustani"`, `/api/search/porudzbenice-potvrdjene`, `/api/search/artikli-sirovine`, `id="dobnaziv"`, `id="prijoj"`,
		`name="kirproid"`}
	for _, tt := range []struct {
		name     string
		rdokID   int64
		snimljen bool
		vrdokID  int64
		storno   bool
	}{
		{name: "new prijemnica", rdokID: 0},
		{name: "saved storno of a porudžbenica", rdokID: 35, snimljen: true, vrdokID: 25, storno: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomPrijDokument{
				Header: domain.KomPrijHeaderView{RdokID: tt.rdokID, Vrd: 198, Dokum: 4, Datum: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), MagaciniID: 3,
					VrdokID: tt.vrdokID, VezaDokum: 6, Storno: tt.storno, Snimljen: tt.snimljen},
				Vrste:    []domain.ComboItem{{Key: "198", Value: "198 - Kalkulacija ulaza sirovine"}},
				Magacini: []domain.ComboItem{{Key: "3", Value: "3 - Materijal"}},
				Konta:    []domain.ComboItem{{Key: "1010", Value: "1010"}},
				NoviRbr:  1,
			}
			var buf bytes.Buffer
			if err := tmpl_kom.PrijemnicaBody(prijUI(), prijKomUI(), dok, domain.TableData{}, domain.TableData{}, prijButtons(),
				prijDugmadZa(tt.rdokID), tr).Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			header := zahtevPanel(t, html, prijHeaderPanelID)
			stavke := zahtevPanel(t, html, prijStavkePanelID)
			if got := zahtevOtvoren(header); got == tt.snimljen {
				t.Errorf("header open = %v, want %v", got, !tt.snimljen)
			}
			if got := zahtevOtvoren(stavke); got != tt.snimljen {
				t.Errorf("stavke open = %v, want %v", got, tt.snimljen)
			}
			for _, id := range []string{"vrd", "magaciniid"} {
				sel := regexp.MustCompile(`<select[^>]*id="` + id + `"[^>]*>`).FindString(html)
				if got := strings.Contains(sel, "disabled"); got != (tt.rdokID > 0) {
					t.Errorf("%s disabled = %v: %s", id, got, sel)
				}
			}
			if got := strings.Contains(html, `id="prijizvorid"`); got != (tt.vrdokID > 0) {
				t.Errorf("the stavke of the porudžbenica shown = %v", got)
			}
			for _, id := range []string{"kom-prij-pakovanje", "kom-prij-karton"} {
				if got := strings.Contains(html, `id="`+id+`"`); got != (tt.rdokID > 0) {
					t.Errorf("%s shown = %v", id, got)
				}
			}
			komStampaProvera(t, html, "kom-prij-stampa", prijNazadID, tt.rdokID == 0)
			storno := regexp.MustCompile(`<span class="[^"]*text-red-700[^"]*"`).FindString(html)
			if got := !strings.Contains(storno, "hidden"); got != tt.storno {
				t.Errorf("STORNO shown = %v: %s", got, storno)
			}
		})
	}
}

// komStampaProvera checks "Štampa" of an entry dialog: always in the bar above the header before "Nazad", disabled
// (nov) until the document is saved.
func komStampaProvera(t *testing.T, html, stampaID, nazadID string, nov bool) {
	t.Helper()
	stampa := regexp.MustCompile(`<[^>]*id="` + stampaID + `"[^>]*>`).FindString(html)
	if stampa == "" || strings.Index(html, `id="`+stampaID+`"`) > strings.Index(html, `id="`+nazadID+`"`) {
		t.Errorf("Štampa is not before Nazad: %q", stampa)
	}
	if got := strings.Contains(stampa, "pointer-events-none"); got != nov {
		t.Errorf("Štampa disabled = %v, want %v", got, nov)
	}
}

// TestPrijPakovanjaRender checks the dialog "Pakovanje": the form of the pakovanje the service binds and "Sačuvaj" with
// the fields of the form.
func TestPrijPakovanjaRender(t *testing.T) {
	tr := i18n.GetInstance()
	var buf bytes.Buffer
	dialog := domain.Dialog{Id: prijPakovanjaDialogID, Title: "Količine po pakovanju"}
	btnSave := domain.Button{Id: "pak-sacuvaj", LabelText: "Sačuvaj", IsVisible: true, HxActionURL: prijPakovanjeSaveURL, HxRequestType: "POST",
		HxInclude: "#" + prijPakovanjaDialogID + "-forma input"}
	if err := tmpl_kom.PrijemnicaPakovanja(dialog, domain.TableData{}, 35, btnSave, domain.Button{Id: "pak-nazad", IsVisible: true},
		domain.Button{Id: "pak-close", IsVisible: true}, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, s := range []string{`id="` + prijPakovanjaDialogID + `-forma"`, `name="rdokid" value="35"`, `name="pakrproid"`, `name="kolpopak"`,
		`name="brpak"`, prijPakovanjeSaveURL} {
		if !strings.Contains(html, s) {
			t.Errorf("the dialog has no %s", s)
		}
	}
}

// TestPrijStampaRender checks the print of a prijemnica (the title, the porudžbenica, the columns of the legacy report,
// the signatures and the warning of a nepotvrđena one) and the karton identifikacije.
func TestPrijStampaRender(t *testing.T) {
	tr := i18n.GetInstance()
	datum := sql.NullTime{Time: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Valid: true}
	p := domain.KomPrijStampa{Dokum: 4, Datum: datum, VezaDokum: 6, VezaDatum: datum, Magacin: "3-MATERIJAL", OJ: "1 - Proizvodnja", Nepotvrdjen: true,
		Dobavljac: domain.KomPonudaStampaDobavljac{Konto: "4350", Sifra: "0381", Naziv: "ROBINIA DOO"},
		Stavke:    []domain.KomPrijStavkaRow{{Rbr: 1, Sifra: 5, Naziv: "PROTILAC 25KG", JM: "KG", TrazKolic: 10, Kolic: 9.5, Otk: "001", Serija: "S-7"}}}
	var buf bytes.Buffer
	if err := tmpl_rep_kom.PrijemnicaStampa(domain.ReportParameters{}, p, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, s := range []string{"PROTILAC 25KG", ">S-7<", ">001<", "09.03.2026", "3-MATERIJAL", "1 - Proizvodnja", "ROBINIA DOO", "text-red-700"} {
		if !strings.Contains(html, s) {
			t.Errorf("the print has no %s", s)
		}
	}
	buf.Reset()
	k := domain.KomPrijKarton{Dobavljac: "ROBINIA DOO", Dokum: 4, Datum: datum, Sifra: 5, Naziv: "PROTILAC", JM: "KG", Serija: "S-7", Otk: "001", Kolicina: 9.5}
	if err := tmpl_rep_kom.PrijemnicaKarton(domain.ReportParameters{}, k, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html = buf.String()
	for _, s := range []string{"ROBINIA DOO", "09.03.2026", ">S-7<", ">001<", "PROTILAC"} {
		if !strings.Contains(html, s) {
			t.Errorf("the karton has no %s", s)
		}
	}
}
