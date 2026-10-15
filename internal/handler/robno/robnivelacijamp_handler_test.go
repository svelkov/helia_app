package robno

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"

	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestRobnoMniRoutes checks that the urls the "Nivelacije maloprodaje" screen sends its requests to (the constants of
// the buttons, of the grids and of the script) are routes of robnoMniRoutes, and that the group MNI opens the screen.
func TestRobnoMniRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&RobnoDokumentaUnosHandler{}).robnoMniRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	tests := []struct {
		method, path string
	}{
		{http.MethodGet, robnoMniDokumentURL + "/:id"},
		{http.MethodGet, robnoMniNoviURL},
		{http.MethodPost, robnoMniSaveURL},
		{http.MethodGet, robnoMniPotvrdaURL},
		{http.MethodDelete, robnoMniBrisiURL + "/:id"},
		{http.MethodGet, robnoMniArtikalURL + "/stopa"},
		{http.MethodGet, robnoMniStavkaURL},
		{http.MethodPost, robnoMniStavkaSaveURL},
		{http.MethodGet, robnoMniStavkaPotvrdaURL},
		{http.MethodDelete, robnoMniStavkaBrisiURL + "/:id"},
		{http.MethodGet, robnoMniPrikazURL},
		{http.MethodGet, robnoMniPrikazDokumentaURL},
		{http.MethodGet, robnoMniPrikazStavkeURL + "/:id"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if !routes[tt.method+" "+tt.path] {
				t.Errorf("route %s %s is not registered", tt.method, tt.path)
			}
		})
	}
	if ekran := robnoDokumentaUnosEkrani["MNI"]; ekran.open == nil {
		t.Errorf("the group MNI has no screen")
	}
}

// TestRobnoMniButtons checks the buttons of a new and of a saved document: a new one has nothing to delete.
func TestRobnoMniButtons(t *testing.T) {
	tests := []struct {
		name         string
		rdokID       int64
		wantDisabled bool
	}{
		{name: "new document", rdokID: 0, wantDisabled: true},
		{name: "saved document", rdokID: 41, wantDisabled: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btns := robnoMniButtons(9, tt.rdokID, 150)
			if btns.Delete.IsDisabled != tt.wantDisabled {
				t.Errorf("Delete.IsDisabled = %v, want %v", btns.Delete.IsDisabled, tt.wantDisabled)
			}
			if !strings.Contains(btns.New.HxActionURL, "rnalid=9&vrd=150") {
				t.Errorf("New.HxActionURL = %q, want the nalog and the vrsta", btns.New.HxActionURL)
			}
		})
	}
}

// TestRobMniBodyRender checks the content of the tab of the entry: the two collapsible controls (the header open
// and enabled, the stavke closed and disabled for a document that is not saved; the other way round after the save)
// with the fields the service binds (RobnoMniParams, RobMniStavkaParams).
func TestRobMniBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	header := domain.RobnoMniHeaderView{VrstaOpis: "Nivelacija MP"}
	header.RnalID, header.Vrd, header.BrojDokum = 9, "150", "3"
	header.KupacKonto, header.KupacSifra, header.KupacNaziv = "1340", "15", "Prodavnica 1"
	stavka := domain.RobnoMniStavka{RdokID: 0, Vrd: "150", Brdok: "3", PromeniTarifu: false}
	fields := []string{`name="kupackonto"`, `name="kupacsifra"`, `name="idorgjed"`, `name="mestotrid"`, `name="brdokorg"`,
		`name="datumizvora"`, `name="opis"`, `name="datumfakturisanja"`, `name="nazivmni"`, `name="kolicina"`, `name="jm"`,
		`name="staracena"`, `name="novacena"`, `name="procenat"`, `name="unesiprocenat"`, `name="staratarifa"`,
		`name="promenitarifu"`, `name="novatarifa"`, `id="stavkardokid"`, `id="stavkarproid"`}
	tests := []struct {
		name     string
		snimljen bool
	}{
		{name: "new document", snimljen: false},
		{name: "saved header", snimljen: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := header
			h.Snimljen = tt.snimljen
			var buf bytes.Buffer
			if err := tmpl_robno.RobMniBody(robnoMniUI(), h, stavka, domain.TableData{}, robnoMniButtons(9, 0, 150), tr).Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			header := robnoMniPanel(t, html, robnoMniHeaderPanelID)
			stavke := robnoMniPanel(t, html, robnoMniStavkePanelID)
			if got := robnoMniOtvoren(header); got == tt.snimljen {
				t.Errorf("header panel open = %v, want %v", got, !tt.snimljen)
			}
			if got := robnoMniOtvoren(stavke); got != tt.snimljen {
				t.Errorf("stavke panel open = %v, want %v", got, tt.snimljen)
			}
			// The nova tarifa is entered only with "Promeni tarifu", the procenat only with "Unesi procenat".
			for _, id := range []string{"novatarifa", "procenat"} {
				if !robnoMniDisabled(html, id) {
					t.Errorf("%s is not disabled", id)
				}
			}
		})
	}
}

// robnoMniPanel returns the opening tag of the <details> of the panel id.
func robnoMniPanel(t *testing.T, html, id string) string {
	t.Helper()
	i := strings.Index(html, `id="`+id+`"`)
	if i < 0 {
		t.Fatalf("no panel %s", id)
	}
	start := strings.LastIndex(html[:i], "<details")
	end := strings.Index(html[i:], ">")
	if start < 0 || end < 0 {
		t.Fatalf("panel %s is not a <details>", id)
	}
	return html[start : i+end]
}

// robnoMniOtvorenAtribut is the attribute open of a <details> (not the Tailwind variant "open:" of its class).
var robnoMniOtvorenAtribut = regexp.MustCompile(`\sopen(\s|=|$)`)

// robnoMniOtvoren tells whether the opening tag of a <details> has the attribute open.
func robnoMniOtvoren(tag string) bool {
	return robnoMniOtvorenAtribut.MatchString(tag)
}

// robnoMniDisabled tells whether the input of the id is disabled.
func robnoMniDisabled(html, id string) bool {
	i := strings.Index(html, `id="`+id+`"`)
	if i < 0 {
		return false
	}
	start := strings.LastIndex(html[:i], "<input")
	end := strings.Index(html[i:], ">")
	return start >= 0 && end >= 0 && strings.Contains(html[start:i+end], "disabled")
}
