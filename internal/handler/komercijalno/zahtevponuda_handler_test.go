package komercijalno

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"

	tmpl_kom "helia/frontend/templates/komercijalno"
	"helia/i18n"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestPonudaRoutes checks that the urls the screen sends its requests to (the constants of the buttons, of the grids,
// of the dialogs and of the script) are routes of AddRoutes.
func TestPonudaRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&ZahtevPonudaHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, ponudaURL},
		{http.MethodGet, ponudaListaURL},
		{http.MethodGet, ponudaDialogURL},
		{http.MethodGet, ponudaNoviURL},
		{http.MethodGet, ponudaIzvoriURL},
		{http.MethodPost, ponudaSaveURL},
		{http.MethodGet, ponudaBrisiPotvrdaURL},
		{http.MethodDelete, ponudaBrisiURL + "/:id"},
		{http.MethodGet, ponudaPotvrdaURL},
		{http.MethodGet, ponudaPotvrdaPitanjeURL},
		{http.MethodPost, ponudaPotvrdiURL + "/:id"},
		{http.MethodGet, ponudaValutaURL},
		{http.MethodGet, ponudaArtikalURL},
		{http.MethodGet, ponudaStavkaURL},
		{http.MethodPost, ponudaStavkaSaveURL},
		{http.MethodPost, ponudaPreuzmiURL},
		{http.MethodGet, ponudaStavkaPotvrdaURL},
		{http.MethodDelete, ponudaStavkaBrisiURL + "/:id"},
		{http.MethodGet, ponudaPrikazStavkeURL + "/:id"},
		{http.MethodGet, ponudaStampaDobavljaciURL},
		{http.MethodGet, ponudaStampaURL},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestPonudaBodyRender checks the content of the tab "Ponude dobavljačima": the two collapsible controls (a new zahtev:
// the header open and enabled, the stavke closed and disabled; a saved one the other way round), the fields the
// service binds, the magacin and the vrsta fixed for a saved zahtev and the dobavljači of the header.
func TestPonudaBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	fields := []string{`name="vrd"`, `name="magaciniid"`, `id="brdok"`, `name="datum"`, `name="rok"`, `id="dospeva"`, `name="komsifra"`,
		`name="idorgjed"`, `name="mestotrid"`, `name="foot"`, `name="sifval"`, `name="kurs"`, `name="vaziza"`, `name="dobavljaci"`,
		`name="izabranestavke"`, `name="rdokid"`, `name="rproid"`, `name="sifraartikla"`, `name="kolicina"`, `id="kom-ponuda-potvrda-btn"`,
		`id="kom-ponuda-preuzmi"`, `id="kom-ponuda-stavka-odustani"`, `/api/search/dobavljaci-ponude`,
		`/api/search/artikli-proizvodnja`, `/api/search/valute`}
	for _, tt := range []struct {
		name     string
		rdokID   int64
		snimljen bool
	}{
		{name: "new zahtev", rdokID: 0},
		{name: "saved header", rdokID: 15, snimljen: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomPonudaDokument{
				Header: domain.KomPonudaHeaderView{RdokID: tt.rdokID, Vrd: 310, Dokum: 4, MagaciniID: 3, Rok: 10, SifVal: 941, OznVal: "RSD", Kurs: 1,
					VaziZa: 1, Snimljen: tt.snimljen, Dobavljaci: []domain.KomDobavljac{{IDFkpl: 101, Konto: "4350", Sifra: "00012", Naziv: "Metal"}}},
				Vrste:    []domain.ComboItem{{Key: "310", Value: "310 - Zahtev za ponudu"}},
				Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Sirovine"}},
				NoviRbr:  1,
			}
			btns := ponudaButtons()
			btns.Print = ponudaStampaButton(tt.rdokID)
			var buf bytes.Buffer
			if err := tmpl_kom.ZahtevPonudaBody(ponudaUI(), ponudaKomUI(), dok, domain.TableData{}, domain.TableData{}, btns,
				ponudaPotvrdaButton(tt.rdokID), ponudaPreuzmiButton(), tr).Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			komStampaProvera(t, html, "kom-ponuda-stampa", ponudaNazadID, tt.rdokID == 0)
			if tt.rdokID > 0 && !strings.Contains(html, ponudaStampaDobavljaciURL) {
				t.Errorf("Štampa does not open the dialog of the dobavljači")
			}
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			header := zahtevPanel(t, html, ponudaHeaderPanelID)
			stavke := zahtevPanel(t, html, ponudaStavkePanelID)
			if got := zahtevOtvoren(header); got == tt.snimljen {
				t.Errorf("header open = %v, want %v", got, !tt.snimljen)
			}
			if got := zahtevOtvoren(stavke); got != tt.snimljen {
				t.Errorf("stavke open = %v, want %v", got, tt.snimljen)
			}
			if !strings.Contains(stavke, `data-enabled="`+boolText(tt.snimljen)+`"`) {
				t.Errorf("stavke data-enabled, want %v: %s", tt.snimljen, stavke)
			}
			// The vrsta and the magacin of a saved zahtev are fixed.
			for _, id := range []string{"vrd", "magaciniid"} {
				sel := regexp.MustCompile(`<select[^>]*id="` + id + `"[^>]*>`).FindString(html)
				if got := strings.Contains(sel, "disabled"); got != (tt.rdokID > 0) {
					t.Errorf("%s disabled = %v: %s", id, got, sel)
				}
			}
			// The dobavljači of the header: the hidden vezniid and their list.
			if !strings.Contains(html, `value="101;"`) || !strings.Contains(html, "4350 00012 - Metal") {
				t.Errorf("the dobavljači of the header are missing")
			}
		})
	}
}
