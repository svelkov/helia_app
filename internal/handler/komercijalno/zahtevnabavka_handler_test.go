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

// TestZahtevRoutes checks that the urls the screen sends its requests to (the constants of the buttons, of the grids
// and of the script) are routes of AddRoutes.
func TestZahtevRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&ZahtevNabavkaHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, zahtevURL},
		{http.MethodGet, zahtevListaURL},
		{http.MethodGet, zahtevDialogURL},
		{http.MethodGet, zahtevNoviURL},
		{http.MethodPost, zahtevSaveURL},
		{http.MethodGet, zahtevBrisiPotvrdaURL},
		{http.MethodDelete, zahtevBrisiURL + "/:id"},
		{http.MethodGet, zahtevPotvrdaURL},
		{http.MethodPost, zahtevPotvrdiURL + "/:id"},
		{http.MethodGet, zahtevArtikalURL},
		{http.MethodGet, zahtevStavkaURL},
		{http.MethodPost, zahtevStavkaSaveURL},
		{http.MethodGet, zahtevStavkaPotvrdaURL},
		{http.MethodDelete, zahtevStavkaBrisiURL + "/:id"},
		{http.MethodGet, zahtevPrikazStavkeURL + "/:id"},
		{http.MethodGet, zahtevStampaURL},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// zahtevPanel returns the <details> of a collapsible control of the rendered body.
func zahtevPanel(t *testing.T, html, id string) string {
	t.Helper()
	m := regexp.MustCompile(`<details[^>]*id="` + regexp.QuoteMeta(id) + `"[^>]*>`).FindString(html)
	if m == "" {
		t.Fatalf("no control %s", id)
	}
	return m
}

// TestZahtevBodyRender checks the content of the tab "Zahtev za nabavku": the two collapsible controls (a new zahtev:
// the header open and enabled, the stavke closed and disabled; a saved one the other way round), the fields the
// service binds and the vrsta and the magacin fixed for a saved zahtev.
func TestZahtevBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	fields := []string{`name="vrd"`, `name="magaciniid"`, `id="brdok"`, `name="datum"`, `name="komsifra"`, `name="idorgjed"`,
		`name="mestotrid"`, `name="foot"`, `name="rdokid"`, `name="rproid"`, `name="sifraartikla"`, `name="kolicina"`,
		`id="kom-zahtev-potvrda"`, `id="kom-zahtev-stavka-odustani"`, `/api/search/artikli-proizvodnja`}
	for _, tt := range []struct {
		name     string
		rdokID   int64
		snimljen bool
	}{
		{name: "new zahtev", rdokID: 0},
		{name: "saved header", rdokID: 9, snimljen: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomZahtevDokument{
				Header:   domain.KomZahtevHeaderView{RdokID: tt.rdokID, Vrd: 410, Dokum: 12, MagaciniID: 3, Snimljen: tt.snimljen},
				Vrste:    []domain.ComboItem{{Key: "410", Value: "410 - Zahtev"}},
				Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Sirovine"}},
				NoviRbr:  1,
			}
			btns := zahtevButtons()
			btns.Print = zahtevStampaButton(tt.rdokID)
			var buf bytes.Buffer
			if err := tmpl_kom.ZahtevNabavkaBody(zahtevUI(), dok, domain.TableData{}, btns, zahtevPotvrdaButton(tt.rdokID), tr).
				Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			komStampaProvera(t, html, "kom-zahtev-stampa", zahtevNazadID, tt.rdokID == 0)
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			header := zahtevPanel(t, html, zahtevHeaderPanelID)
			stavke := zahtevPanel(t, html, zahtevStavkePanelID)
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
			vrsta := regexp.MustCompile(`<select[^>]*id="vrd"[^>]*>`).FindString(html)
			if got := strings.Contains(vrsta, "disabled"); got != (tt.rdokID > 0) {
				t.Errorf("vrsta disabled = %v: %s", got, vrsta)
			}
		})
	}
}

// zahtevOtvoren tells whether a <details> has the attribute open (not a class such as "open:flex").
func zahtevOtvoren(details string) bool {
	return regexp.MustCompile(`\sopen(=""|\s|>)`).MatchString(details)
}

func boolText(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
