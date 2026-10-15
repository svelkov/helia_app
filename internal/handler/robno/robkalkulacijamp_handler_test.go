package robno

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestRobnoKmpRoutes checks that the urls the "Kalkulacije maloprodaje" screen sends its requests to (the constants of
// the buttons, of the grids and of the script) are routes of robnoKmpRoutes, and that the group KAL opens the screen.
func TestRobnoKmpRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&RobnoDokumentaUnosHandler{}).robnoKmpRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, robnoKmpDokumentURL + "/:id"},
		{http.MethodGet, robnoKmpNoviURL},
		{http.MethodPost, robnoKmpSaveURL},
		{http.MethodGet, robnoKmpPotvrdaURL},
		{http.MethodDelete, robnoKmpBrisiURL + "/:id"},
		{http.MethodGet, robnoKmpStopaURL},
		{http.MethodGet, robnoKmpArtikalURL},
		{http.MethodGet, robnoKmpStavkaURL},
		{http.MethodPost, robnoKmpStavkaSaveURL},
		{http.MethodGet, robnoKmpStavkaPotvrdaURL},
		{http.MethodDelete, robnoKmpStavkaBrisiURL + "/:id"},
		{http.MethodGet, robnoKmpPrikazURL},
		{http.MethodGet, robnoKmpPrikazDokumentaURL},
		{http.MethodGet, robnoKmpPrikazStavkeURL + "/:id"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if !routes[tt.method+" "+tt.path] {
				t.Errorf("route %s %s is not registered", tt.method, tt.path)
			}
		})
	}
	if ekran := robnoDokumentaUnosEkrani["KAL"]; ekran.open == nil {
		t.Errorf("the group KAL has no screen")
	}
}

// TestRobKmpBodyRender checks the content of the tab of the entry: the two collapsible controls (the header open and
// enabled, the stavke closed and disabled before the save; the other way round after it), the fields the service binds,
// the buttons of the stavke ("Sačuvaj", "Odustani" disabled) and the POPDV polje hidden for a vrsta that does not ask
// it.
func TestRobKmpBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	dok := domain.RobnoKmpDokument{}
	dok.Header.RnalID, dok.Header.Vrd, dok.Header.BrojDokum, dok.Header.MagaciniID = 9, "120", "12", 3
	dok.Header.ProdavnicaKonto, dok.Header.ProdavnicaSifra = "1340", "01"
	dok.Header.Epp, dok.Header.Odbitni = true, true
	dok.Header.Porezi = []domain.RobKalPorez{{ID: 5, Osnovica: 1000, Porez: 200, Po: 2, Stopa: 20}}
	dok.Stavka.Konto, dok.Stavka.Pana, dok.Stavka.Maloprodaja = "1340", "01", true
	fields := []string{`name="kupackonto"`, `name="kupacsifra"`, `name="pkto"`, `name="pana"`, `name="idorgjed"`, `name="mestotrid"`,
		`name="brdokorg"`, `name="datumizvora"`, `name="opis"`, `name="usloviplacanja"`, `name="rok"`, `name="foot"`, `name="napomporeskoosl"`,
		`name="epp"`, `name="fseppid"`, `name="popdv"`, `name="odbitni"`, `name="porezosn"`, `name="porezppor"`, `name="porezpo"`,
		`name="porezi"`, `name="sifraartikla"`, `name="nazivkmp"`, `name="jm"`, `name="tarifa"`, `name="kolicina"`, `name="fakturnacena"`,
		`name="rabatstavke"`, `name="tz"`, `name="ztro"`, `name="mma"`, `name="mpcena"`, `id="stavkarproid"`, `id="robno-kmp-stavka-odustani"`}
	for _, tt := range []struct {
		name     string
		snimljen bool
	}{
		{name: "new document", snimljen: false},
		{name: "saved header", snimljen: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := dok
			d.Header.Snimljen = tt.snimljen
			var buf bytes.Buffer
			if err := tmpl_robno.RobKmpBody(robnoKmpUI(), robnoKmpEkranUI(), d, domain.TableData{}, nil, robnoKmpButtons(9, 0, 120), tr).
				Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			header := robnoMniPanel(t, html, robnoKmpHeaderPanelID)
			stavke := robnoMniPanel(t, html, robnoKmpStavkePanelID)
			if got := robnoMniOtvoren(header); got == tt.snimljen {
				t.Errorf("header open = %v, want %v", got, !tt.snimljen)
			}
			if got := robnoMniOtvoren(stavke); got != tt.snimljen {
				t.Errorf("stavke open = %v, want %v", got, tt.snimljen)
			}
			if !strings.Contains(stavke, `data-enabled="`+robnoFaktureStateTest(tt.snimljen)+`"`) {
				t.Errorf("stavke data-enabled, want %v: %s", tt.snimljen, stavke)
			}
			// The POPDV polje is hidden for a vrsta that does not ask it.
			if i := strings.Index(html, `name="popdv"`); i < 0 || !strings.Contains(html[strings.LastIndex(html[:i], `<div class="flex items-center gap-2`):i], "hidden") {
				t.Errorf("the POPDV polje is not hidden")
			}
		})
	}
}
