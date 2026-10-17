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

// TestEotpKoraciRoutes checks that the urls the stepper sends its requests to are routes: its own ones (AddRoutes of
// the stepper) and the ones of the main screen of the e-otpremnice.
func TestEotpKoraciRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&EotpremnicaHandler{OtpremnicaHandler: &OtpremnicaHandler{}}).AddRoutes(r)
	(&EotpremnicaKoraciHandler{EotpremnicaHandler: &EotpremnicaHandler{OtpremnicaHandler: &OtpremnicaHandler{}}}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	p := eotpKoraciPutanje
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, p.url},
		{http.MethodGet, p.lista},
		{http.MethodGet, p.dialog},
		{http.MethodGet, p.novi},
		{http.MethodPost, p.save},
		{http.MethodGet, p.brisiPotvrda},
		{http.MethodDelete, p.brisi + "/:id"},
		{http.MethodGet, p.potvrda},
		{http.MethodPost, p.potvrdi + "/:id"},
		{http.MethodGet, p.nalog},
		{http.MethodGet, p.valuta},
		{http.MethodGet, p.stavka},
		{http.MethodPost, p.stavkaSave},
		{http.MethodGet, p.prikazStavke + "/:id"},
		{http.MethodGet, p.stampa},
		{http.MethodGet, p.eotpNalog},
		{http.MethodGet, p.partner},
		{http.MethodGet, p.mestoIsp},
		{http.MethodGet, eotpPutanje.koraciDialog},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestEotpKoraciBodyRender checks the stepper: the six steps in their order with the ids of the controls, one of them
// shown (the header of a new otpremnica, the stavke of a saved one), the steps of the data editable until the save and
// the stavke after it, the fields of every step, "Prethodni", "Sledeći", "Sačuvaj" sending the steps of the data and
// "Izmeni".
func TestEotpKoraciBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	p := eotpKoraciPutanje
	for _, tt := range []struct {
		name    string
		rdokID  int64
		pocetni string
	}{
		{name: "new e-otpremnica", pocetni: "1"},
		{name: "saved e-otpremnica", rdokID: 60, pocetni: "5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomOtpDokument{
				Header: domain.KomOtpHeaderView{KomOtpremnicaRow: domain.KomOtpremnicaRow{RdokID: tt.rdokID, Vrd: 188, MagaciniID: 3, Doksts: "10"},
					Snimljen: tt.rdokID > 0},
				Vrste: []domain.ComboItem{{Key: "188", Value: "188 - Faktura"}}, Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Proizvodi"}},
				Eotp: &domain.KomEotpView{KomEotpPodaci: domain.KomEotpPodaci{TipEotp: domain.KomEotpEksterna}},
			}
			btns := otpButtons(p)
			var buf bytes.Buffer
			if err := tmpl_kom.OtpremnicaBody(otpUI(p), otpKomUI(p), dok, domain.TableData{}, btns, otpDugmadZa(p, tt.rdokID), tr).
				Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			if !strings.Contains(html, `data-koraci="true"`) || !strings.Contains(html, `data-korak="`+tt.pocetni+`"`) {
				t.Fatalf("not a stepper on the step %s", tt.pocetni)
			}
			koraci := []string{otpHeaderPanelID, eotpDet1PanelID, eotpDet2PanelID, eotpDet3PanelID, otpStavkePanelID, eotpDet5PanelID}
			poz := -1
			for i, id := range koraci {
				sekcija := regexp.MustCompile(`<section id="` + id + `"[^>]*>`).FindString(html)
				if sekcija == "" || !strings.Contains(sekcija, `data-korak-sadrzaj="`+string(rune('1'+i))+`"`) {
					t.Fatalf("the step %d (%s) is missing: %s", i+1, id, sekcija)
				}
				if j := strings.Index(html, sekcija); j <= poz {
					t.Errorf("the step %s is out of order", id)
				} else {
					poz = j
				}
				prikazan := !strings.Contains(sekcija, "hidden")
				if prikazan != (string(rune('1'+i)) == tt.pocetni) {
					t.Errorf("the step %d shown = %v", i+1, prikazan)
				}
				enabled := id != otpStavkePanelID
				if tt.rdokID > 0 {
					enabled = !enabled
				}
				if !strings.Contains(sekcija, `data-enabled="`+map[bool]string{true: "true", false: "false"}[enabled]+`"`) {
					t.Errorf("the step %s enabled should be %v: %s", id, enabled, sekcija)
				}
			}
			if strings.Contains(html, "<details") {
				t.Errorf("the stepper has collapsible controls")
			}
			for _, f := range []string{`name="vrd"`, `name="tipeotp"`, `name="nacotp"`, `name="ulicaodr"`, `name="kolicina"`, `name="roba3strtxt"`,
				`id="kom-eotp-koraci-prethodni"`, `id="kom-eotp-koraci-sledeci"`, `id="kom-otp-sacuvaj"`, `id="kom-otp-izmeni"`, `data-korak-dugme="6"`} {
				if !strings.Contains(html, f) {
					t.Errorf("the stepper has no %s", f)
				}
			}
			for _, id := range []string{otpHeaderPanelID, eotpDet1PanelID, eotpDet2PanelID, eotpDet3PanelID, eotpDet5PanelID} {
				if !strings.Contains(btns.Save.HxInclude, "#"+id+" [data-panel-fields]") {
					t.Errorf("Sačuvaj does not send %s", id)
				}
			}
			if !strings.Contains(html, `hx-post="`+eotpKoraciSaveURL+`"`) {
				t.Errorf("Sačuvaj does not post to %s", eotpKoraciSaveURL)
			}
		})
	}
}
