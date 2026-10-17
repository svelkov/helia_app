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
	"helia/i18n"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestEotpRoutes checks that the urls the screen sends its requests to are routes of AddRoutes.
func TestEotpRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&EotpremnicaHandler{OtpremnicaHandler: &OtpremnicaHandler{}}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, eotpPutanje.url},
		{http.MethodGet, eotpPutanje.lista},
		{http.MethodGet, eotpPutanje.dialog},
		{http.MethodGet, eotpPutanje.novi},
		{http.MethodPost, eotpPutanje.save},
		{http.MethodGet, eotpPutanje.brisiPotvrda},
		{http.MethodDelete, eotpPutanje.brisi + "/:id"},
		{http.MethodGet, eotpPutanje.potvrda},
		{http.MethodPost, eotpPutanje.potvrdi + "/:id"},
		{http.MethodGet, eotpPutanje.nalog},
		{http.MethodGet, eotpPutanje.valuta},
		{http.MethodGet, eotpPutanje.stavka},
		{http.MethodPost, eotpPutanje.stavkaSave},
		{http.MethodGet, eotpPutanje.prikazStavke + "/:id"},
		{http.MethodGet, eotpPutanje.stampa},
		{http.MethodGet, eotpPutanje.eotpNalog},
		{http.MethodGet, eotpPutanje.partner},
		{http.MethodGet, eotpPutanje.mestoIsp},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestEotpBodyRender checks the entry dialog of an e-otpremnica: the four controls of the e-otpremnica around the stavke
// (DET1-DET3 before, DET5 after), their fields, the fields that moved out of the header (once each), the groups of the
// način otpreme by its value, the groups of the opasne materije and of the roba treće strane enabled by their
// checkboxes, the rows of the grids, the locked controls of a saved one, "Sačuvaj" sending the controls; the otpremnica
// without the controls.
func TestEotpBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	e := &domain.KomEotpView{
		KomEotpPodaci: domain.KomEotpPodaci{OpasneMat: true, TipEotp: domain.KomEotpInterna, Jm1: "kg", NacOtp: domain.KomEotpPrevoznik, PrevKto: "4350",
			PrevAna: "12", Datpoc: sql.NullTime{Time: time.Date(2026, 3, 10, 8, 15, 0, 0, time.UTC), Valid: true}, UlicaOdr: "Bulevar"},
		Faze:       []domain.KomEotpFaza{{Rbr: 1, NacOtp: 2, Naziv: "Prevoz DOO"}},
		Opasne:     []domain.KomEotpOpasna{{Rbr: 1, Naziv: "UN 1203", Opis: "Benzin"}},
		Posiljalac: domain.KomEotpPartner{Naziv: "FIRMA DOO"},
		Primalac:   domain.KomEotpPartner{Naziv: "KUPAC DOO", Jbkjs: "12345"},
		Prevoznik:  domain.KomEotpPartner{Naziv: "PREVOZ DOO"},
	}
	for _, tt := range []struct {
		name   string
		rdokID int64
		eotp   *domain.KomEotpView
	}{
		{name: "new e-otpremnica", eotp: e},
		{name: "saved e-otpremnica", rdokID: 60, eotp: e},
		{name: "otpremnica", rdokID: 60},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := otpPutanjeOtp
			if tt.eotp != nil {
				p = eotpPutanje
			}
			dok := domain.KomOtpDokument{
				Header: domain.KomOtpHeaderView{KomOtpremnicaRow: domain.KomOtpremnicaRow{RdokID: tt.rdokID, Vrd: 188, Brotp: 990001, MagaciniID: 3,
					Fkto: "2040", Fana: "15", Vozac: "Pera", Doksts: "10"}, Snimljen: tt.rdokID > 0},
				Vrste: []domain.ComboItem{{Key: "188", Value: "188 - Faktura"}}, Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Proizvodi"}},
				Eotp: tt.eotp,
			}
			var buf bytes.Buffer
			btns := otpButtons(p)
			if err := tmpl_kom.OtpremnicaBody(otpUI(p), otpKomUI(p), dok, domain.TableData{}, btns, otpDugmadZa(p, tt.rdokID), tr).
				Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			paneli := []string{eotpDet1PanelID, eotpDet2PanelID, eotpDet3PanelID, eotpDet5PanelID}
			if tt.eotp == nil {
				for _, id := range paneli {
					if strings.Contains(html, id) {
						t.Errorf("the otpremnica has the control %s", id)
					}
				}
				if strings.Count(html, `name="vozac"`) != 1 || strings.Contains(btns.Save.HxInclude, eotpDet1PanelID) {
					t.Errorf("the header of the otpremnica changed")
				}
				return
			}
			// The order: header, DET1, DET2, DET3, stavke, DET5.
			redosled := []string{otpHeaderPanelID, eotpDet1PanelID, eotpDet2PanelID, eotpDet3PanelID, otpStavkePanelID, eotpDet5PanelID}
			poz := -1
			for _, id := range redosled {
				i := strings.Index(html, `id="`+id+`"`)
				if i <= poz {
					t.Fatalf("the control %s is missing or out of order", id)
				}
				poz = i
			}
			for _, f := range []string{`name="tipeotp"`, `name="jm1"`, `name="jm2"`, `name="bttovol"`, `name="stvdatotp"`, `name="datpoc"`, `name="datzav"`,
				`name="povratrobe"`, `name="brnarudz"`, `name="brokvspor"`, `name="kontime"`, `name="konttel"`, `name="kontmail"`, `name="referkup"`,
				`name="telkup"`, `name="mailkup"`, `name="nacotp"`, `name="licime"`, `name="licprez"`, `name="licbrlk"`, `name="mestoutov"`,
				`name="mestoisp"`, `name="prevkto"`, `name="prevana"`, `name="vozprez"`, `name="vozmail"`, `name="voztel"`, `name="vozdozv"`,
				`name="ulicaotp"`, `name="ulbrotp"`, `name="mestootp"`, `name="posbrotp"`, `name="sifobjotp"`, `name="eotpmi"`, `name="ulicaodr"`,
				`name="ulbrodr"`, `name="mestoodr"`, `name="posbrodr"`, `name="sifobjodr"`, `name="fisnap"`, `name="roba3strtxt"`, `name="opasnemat"`,
				`name="roba3str"`, `value="2026-03-10T08:15"`, `KUPAC DOO`, `FIRMA DOO`, `PREVOZ DOO`, `data-eotp-nalog-url="` + eotpNalogPodaciURL + `"`,
				`name="fazanaziv" value="Prevoz DOO"`, `name="opmanaziv" value="UN 1203"`} {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			// The fields of the header in the controls, once each.
			for _, f := range []string{"vozac", "brvozila", "brugov", "bttowght", "totboxs", "napomena", "otp", "netwght", "brpaleta"} {
				if n := strings.Count(html, `name="`+f+`"`); n != 1 {
					t.Errorf("%s is %d times in the body", f, n)
				}
			}
			for id, skriven := range map[string]bool{"kom-eotp-licno": true, "kom-eotp-prevoz": false, "kom-eotp-prevoznik-sifra": false} {
				el := regexp.MustCompile(`<[a-z]+ id="` + id + `"[^>]*>`).FindString(html)
				if got := strings.Contains(el, "hidden"); got != skriven {
					t.Errorf("%s hidden = %v: %s", id, got, el)
				}
			}
			// The groups of the opasne materije and of the roba treće strane are always shown, disabled without their
			// checkbox.
			for id, disabled := range map[string]bool{"kom-eotp-opasne": false, "kom-eotp-roba3str": true} {
				el := regexp.MustCompile(`<fieldset id="` + id + `"[^>]*>`).FindString(html)
				if strings.Contains(el, "hidden") || strings.Contains(el, " disabled") != disabled {
					t.Errorf("%s: %s", id, el)
				}
			}
			for _, id := range paneli {
				if !strings.Contains(btns.Save.HxInclude, "#"+id+" [data-panel-fields]") {
					t.Errorf("Sačuvaj does not send %s", id)
				}
				el := regexp.MustCompile(`<details[^>]*id="` + id + `"[^>]*>`).FindString(html)
				if got := strings.Contains(el, `data-enabled="false"`); got != (tt.rdokID > 0) {
					t.Errorf("%s locked = %v: %s", id, got, el)
				}
			}
		})
	}
}

// TestEotpPregledRender checks the main screen of the e-otpremnice: the script of the controls of the e-otpremnica, and
// its list loaded from the url of the e-otpremnice.
func TestEotpPregledRender(t *testing.T) {
	tr := i18n.GetInstance()
	ui := otpPregledUI(eotpPutanje)
	var buf bytes.Buffer
	if err := tmpl_kom.OtpremnicaPregled(ui, otpStampaKomercopisID, nil, 0, domain.TableData{}, komPretraga(ui, tr), domain.TableData{}, true, eotpKoraciDialogURL, tr).
		Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	for _, s := range []string{"function komEotpNalog(", "function komEotpNacin(", "function komEotpFazaDodaj(", "function komEotpOpasnaDodaj(",
		"function komEotpPaneli(", "function komOtpNalog(", eotpListaURL, `id="kom-eotp-koraci-btn"`, `hx-get="` + eotpKoraciDialogURL + `"`,
		"function komKoraciIdi("} {
		if !strings.Contains(html, s) {
			t.Errorf("the screen has no %s", s)
		}
	}
}
