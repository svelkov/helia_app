package komercijalno

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"

	tmpl_kom "helia/frontend/templates/komercijalno"
	tmpl_rep_kom "helia/frontend/templates/reports/komercijalno"
	"helia/i18n"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestPorRoutes checks that the urls the screen sends its requests to (the constants of the buttons, of the grids and
// of the script) are routes of AddRoutes.
func TestPorRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&PorudzbenicaHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, porURL},
		{http.MethodGet, porListaURL},
		{http.MethodGet, porDialogURL},
		{http.MethodGet, porNoviURL},
		{http.MethodPost, porSaveURL},
		{http.MethodGet, porBrisiPotvrdaURL},
		{http.MethodDelete, porBrisiURL + "/:id"},
		{http.MethodGet, porPotvrdaURL},
		{http.MethodPost, porPotvrdiURL + "/:id"},
		{http.MethodGet, porDobavljacURL},
		{http.MethodGet, porValutaURL},
		{http.MethodGet, porArtikalURL},
		{http.MethodGet, porStavkaURL},
		{http.MethodPost, porStavkaSaveURL},
		{http.MethodGet, porStavkaPotvrdaURL},
		{http.MethodDelete, porStavkaBrisiURL + "/:id"},
		{http.MethodPost, porPreuzmiURL},
		{http.MethodGet, porVratiPotvrdaURL},
		{http.MethodDelete, porVratiURL + "/:id"},
		{http.MethodGet, porPrikazStavkeURL + "/:id"},
		{http.MethodGet, porStampaURL},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestPorBodyRender checks the content of the tab "Porudžbenice dobavljačima": the two collapsible controls (a new
// porudžbenica: the header open and enabled, the stavke closed and disabled; a saved one the other way round), the
// fields the service binds, the vrsta, the magacin and the zahtev za ponudu fixed for a saved porudžbenica, the stavke
// of the zahtev za ponudu only with one and "UVOZNI DOKUMENT - INO PARTNER" only in a foreign valuta.
func TestPorBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	fields := []string{`name="vrd"`, `name="magaciniid"`, `id="brdok"`, `name="datum"`, `name="vrdokid"`, `name="fkto"`, `name="fana"`,
		`name="otp"`, `name="pla"`, `name="rok"`, `name="brugov"`, `name="nacpak"`, `name="komsifra"`, `name="sifval"`, `name="kurs"`,
		`name="vaziza"`, `name="idorgjed"`, `name="mestotrid"`, `name="foot"`, `name="posuslnab"`, `name="pratdok"`, `name="rdokid"`,
		`name="rproid"`, `name="sifraartikla"`, `name="kolicina"`, `name="cena"`, `name="rabat"`, `id="kom-por-potvrda-btn"`,
		`id="kom-por-stavka-odustani"`, `/api/search/ponude-potvrdjene`, `/api/search/analitika`,
		`/api/search/artikli-proizvodnja`, `id="porinfodobavljac"`, `id="porinfonaziv"`, `id="porinfobroj"`, `id="porinfoiznos"`}
	for _, tt := range []struct {
		name     string
		rdokID   int64
		snimljen bool
		vrdokID  int64
		ino      bool
	}{
		{name: "new porudžbenica", rdokID: 0},
		{name: "saved header of a zahtev za ponudu in a foreign valuta", rdokID: 25, snimljen: true, vrdokID: 15, ino: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomPorDokument{
				Header: domain.KomPorHeaderView{RdokID: tt.rdokID, Vrd: 320, Dokum: 6, MagaciniID: 3, Fkto: "4330", SifVal: 978, OznVal: "EUR", Kurs: 117.25,
					VaziZa: 1, VrdokID: tt.vrdokID, Ino: tt.ino, Snimljen: tt.snimljen},
				Vrste:    []domain.ComboItem{{Key: "320", Value: "320 - Porudžbenica"}},
				Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Sirovine"}},
				NoviRbr:  1,
			}
			var buf bytes.Buffer
			if err := tmpl_kom.PorudzbenicaBody(porUI(), porKomUI(), dok, domain.TableData{}, domain.TableData{}, porButtons(),
				porDugmadZa(tt.rdokID), tr).Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			header := zahtevPanel(t, html, porHeaderPanelID)
			stavke := zahtevPanel(t, html, porStavkePanelID)
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
			if got := strings.Contains(html, `id="porizvorid"`); got != (tt.vrdokID > 0) {
				t.Errorf("the stavke of the zahtev za ponudu shown = %v", got)
			}
			// "Štampa" is in the bar above the header, before "Nazad": disabled until the porudžbenica is saved.
			komStampaProvera(t, html, "kom-por-stampa", porNazadID, tt.rdokID == 0)
			ino := regexp.MustCompile(`<span id="kom-por-ino" class="[^"]*"`).FindString(html)
			if got := !strings.Contains(ino, "hidden"); got != tt.ino {
				t.Errorf("INO shown = %v: %s", got, ino)
			}
		})
	}
}

// TestPorStampaDobavljac checks the dobavljač of the print: a box marked by its four corners like the kupac of the
// faktura ("DOBAVLJAČ:" with the konto and the šifra on one line, no rounded border) and "JM" over the column of the
// JM.
func TestPorStampaDobavljac(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	p := domain.KomPorStampa{Dokum: 1, Dobavljac: domain.KomPonudaStampaDobavljac{Konto: "4350", Sifra: "0381", Naziv: "ROBINIA DOO", TipPdv: 1,
		Pib: "101448098"}, Stavke: []domain.KomPorStampaStavka{{Rbr: 1, Sifra: 5, Opis: "PROTILAC 25KG", JM: "KG", Kolicina: 10, Cena: 144.32}}}
	var buf bytes.Buffer
	if err := tmpl_rep_kom.PorudzbenicaStampa(domain.ReportParameters{}, p, i18n.GetInstance()).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, s := range []string{`absolute top-0 left-0 w-4 h-4 border-t border-l border-black`, `absolute bottom-0 right-0 w-4 h-4 border-b border-r border-black`,
		`<span class="ml-1">4350 0381</span>`, `>JM</th>`} {
		if !strings.Contains(html, s) {
			t.Errorf("the print has no %s", s)
		}
	}
	if strings.Contains(html, "border border-black rounded p-2") {
		t.Errorf("the dobavljač is still in a rounded box")
	}
}
