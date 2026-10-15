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

// TestRobnoKalRoutes checks that the urls the "Kalkulacije veleprodaje" screen sends its requests to (the constants of
// the buttons, of the grids and of the script) are routes of robnoKalRoutes, and that the group PLT opens the screen.
func TestRobnoKalRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&RobnoDokumentaUnosHandler{}).robnoKalRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	tests := []struct {
		method, path string
	}{
		{http.MethodGet, robnoKalDokumentURL + "/:id"},
		{http.MethodGet, robnoKalNoviURL},
		{http.MethodPost, robnoKalSaveURL},
		{http.MethodGet, robnoKalPotvrdaURL},
		{http.MethodDelete, robnoKalBrisiURL + "/:id"},
		{http.MethodGet, robnoKalStopaURL},
		{http.MethodGet, robnoKalArtikalURL},
		{http.MethodGet, robnoKalStavkaURL},
		{http.MethodPost, robnoKalStavkaSaveURL},
		{http.MethodGet, robnoKalStavkaPotvrdaURL},
		{http.MethodDelete, robnoKalStavkaBrisiURL + "/:id"},
		{http.MethodGet, robnoKalTrosakURL},
		{http.MethodPost, robnoKalTrosakSaveURL},
		{http.MethodGet, robnoKalTrosakPotvrdaURL},
		{http.MethodDelete, robnoKalTrosakBrisiURL + "/:id"},
		{http.MethodGet, robnoKalPrikazURL},
		{http.MethodGet, robnoKalPrikazDokumentaURL},
		{http.MethodGet, robnoKalPrikazStavkeURL + "/:id"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if !routes[tt.method+" "+tt.path] {
				t.Errorf("route %s %s is not registered", tt.method, tt.path)
			}
		})
	}
	if ekran := robnoDokumentaUnosEkrani["PLT"]; ekran.open == nil {
		t.Errorf("the group PLT has no screen")
	}
}

// TestRobnoKalButtons checks the buttons of a new and of a saved document: a new one has nothing to delete; "Izmeni"
// switches back from both other controls.
func TestRobnoKalButtons(t *testing.T) {
	for _, tt := range []struct {
		name         string
		rdokID       int64
		wantDisabled bool
	}{
		{name: "new document", rdokID: 0, wantDisabled: true},
		{name: "saved document", rdokID: 41, wantDisabled: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			btns := robnoKalButtons(9, tt.rdokID, 110)
			if btns.Delete.IsDisabled != tt.wantDisabled {
				t.Errorf("Delete.IsDisabled = %v, want %v", btns.Delete.IsDisabled, tt.wantDisabled)
			}
			if len(btns.Modify.HxOnClickArg) != 3 {
				t.Errorf("Modify.HxOnClickArg = %v, want the three controls", btns.Modify.HxOnClickArg)
			}
		})
	}
}

// TestRobKalBodyRender checks the content of the tab of the entry: the three collapsible controls in the order header,
// zavisni troškovi, stavke; before the save only the header is open and enabled, after it the header is locked and
// the zavisni troškovi and the stavke are enabled (one of them open); the fields the service binds are there.
func TestRobKalBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	dok := domain.RobnoKalDokument{}
	dok.Header.RnalID, dok.Header.Vrd, dok.Header.BrojDokum, dok.Header.MagaciniID = 9, "110", "12", 3
	dok.Header.Epp, dok.Header.Odbitni = true, true
	dok.Header.Porezi = []domain.RobKalPorez{{ID: 5, Osnovica: 1000, Porez: 200, Po: 2, Stopa: 20}}
	dok.Stavka.Tipzal = 2
	fields := []string{`name="kupackonto"`, `name="kupacsifra"`, `name="fispid"`, `name="idvalute"`, `name="kurs"`, `name="vaziza"`,
		`name="idorgjed"`, `name="mestotrid"`, `name="brdokorg"`, `name="datumizvora"`, `name="opis"`, `name="edokid"`,
		`name="nacindopremanja"`, `name="usloviplacanja"`, `name="rok"`, `name="uslovikamate"`, `name="foot"`, `name="napomporeskoosl"`,
		`name="epp"`, `name="fseppid"`, `name="popdv"`, `name="odbitni"`, `name="porezosn"`, `name="porezppor"`, `name="porezpo"`,
		`name="avansrac"`, `name="avansdatum"`, `name="porezi"`, `name="avansi"`,
		`name="trkonto"`, `name="trsifra"`, `name="trdokdob"`, `name="trdadok"`, `name="trrok"`, `name="tropis"`, `name="trnapomena"`,
		`name="truvozpdv"`, `name="trpopdv"`, `name="triznfak"`, `name="triznos"`, `name="trtkonto"`, `name="trporezosn"`, `name="trporezi"`,
		`name="kontostavke"`, `name="sifraartikla"`, `name="kolicina"`, `name="fakturnacena"`, `name="rabatstavke"`, `name="tz"`,
		`name="ztro"`, `name="tzi"`, `name="ztrin"`, `name="vma"`, `name="vpcena"`, `name="mma"`, `name="mpcena"`, `name="otk"`,
		`name="serija"`, `name="rokstavke"`}
	tests := []struct {
		name                         string
		snimljen                     bool
		otvoreni                     string
		wantHeader, wantTros, wantSt bool
	}{
		{name: "new document", snimljen: false, otvoreni: robnoKalOtvoreneStavke, wantHeader: true},
		{name: "saved header, the stavke open", snimljen: true, otvoreni: robnoKalOtvoreneStavke, wantSt: true},
		{name: "saved header, the zavisni troškovi open", snimljen: true, otvoreni: robnoKalOtvoreniTroskovi, wantTros: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dok
			d.Header.Snimljen = tt.snimljen
			save, cancel := robnoKalTrosakButtons()
			var buf bytes.Buffer
			if err := tmpl_robno.RobKalBody(robnoKalUI(), robnoKalEkranUI(), d, domain.TableData{}, domain.TableData{}, nil,
				robnoKalButtons(9, 0, 110), save, cancel, tt.otvoreni, tr).Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			header := robnoMniPanel(t, html, robnoKalHeaderPanelID)
			troskovi := robnoMniPanel(t, html, robnoKalTroskoviPanelID)
			stavke := robnoMniPanel(t, html, robnoKalStavkePanelID)
			if i, j, k := strings.Index(html, `id="`+robnoKalHeaderPanelID+`"`), strings.Index(html, `id="`+robnoKalTroskoviPanelID+`"`),
				strings.Index(html, `id="`+robnoKalStavkePanelID+`"`); !(i < j && j < k) {
				t.Errorf("the controls are not in the order header, zavisni troškovi, stavke")
			}
			for _, p := range []struct {
				name      string
				tag       string
				open      bool
				enabled   bool
				wantOpen  bool
				wantState string
			}{
				{"header", header, robnoMniOtvoren(header), true, tt.wantHeader, robnoFaktureStateTest(!tt.snimljen)},
				{"zavisni troškovi", troskovi, robnoMniOtvoren(troskovi), true, tt.wantTros, robnoFaktureStateTest(tt.snimljen)},
				{"stavke", stavke, robnoMniOtvoren(stavke), true, tt.wantSt, robnoFaktureStateTest(tt.snimljen)},
			} {
				if p.open != p.wantOpen {
					t.Errorf("%s open = %v, want %v", p.name, p.open, p.wantOpen)
				}
				if !strings.Contains(p.tag, `data-enabled="`+p.wantState+`"`) {
					t.Errorf("%s: want data-enabled=%q in %s", p.name, p.wantState, p.tag)
				}
			}
			// The saved row of the plaćen porez is in its grid.
			if !strings.Contains(html, `name="porezosn" value="1000"`) {
				t.Errorf("the row of the plaćen porez is not rendered")
			}
		})
	}
}

// robnoFaktureStateTest is the value of the data-enabled attribute of a control.
func robnoFaktureStateTest(enabled bool) string {
	if enabled {
		return "true"
	}
	return "false"
}
