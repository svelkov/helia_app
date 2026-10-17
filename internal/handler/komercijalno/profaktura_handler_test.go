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

// TestProRoutes checks that the urls the screen sends its requests to are routes of AddRoutes.
func TestProRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&ProfakturaHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, proURL},
		{http.MethodGet, proListaURL},
		{http.MethodGet, proDialogURL},
		{http.MethodGet, proNoviURL},
		{http.MethodPost, proSaveURL},
		{http.MethodGet, proBrisiPotvrdaURL},
		{http.MethodDelete, proBrisiURL + "/:id"},
		{http.MethodGet, proPotvrdaURL},
		{http.MethodPost, proPotvrdiURL + "/:id"},
		{http.MethodGet, proNalogPotvrdaURL},
		{http.MethodPost, proNalogURL + "/:id"},
		{http.MethodGet, proKupacURL},
		{http.MethodGet, proMestoURL},
		{http.MethodGet, proValutaURL},
		{http.MethodGet, proCenovnikURL},
		{http.MethodGet, proArtikalURL},
		{http.MethodGet, proStavkaURL},
		{http.MethodPost, proStavkaSaveURL},
		{http.MethodGet, proStavkaPotvrdaURL},
		{http.MethodDelete, proStavkaBrisiURL + "/:id"},
		{http.MethodGet, proPrikazStavkeURL + "/:id"},
		{http.MethodGet, proStampaURL},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestProBodyRender checks the content of the entry dialog: the two collapsible controls, the fields the service binds,
// the tipdok, the magacin and the vrsta fixed for a saved profaktura, the lot only in a magacin with the lots, the
// cenovnik only in a magacin with the tip cene 2, "Štampa" before "Nazad" (disabled for a new one).
func TestProBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	fields := []string{`name="tipdok"`, `name="magaciniid"`, `name="vrd"`, `name="datum"`, `name="rok"`, `name="fkto"`, `name="fana"`, `name="ugrabat"`,
		`name="pkase"`, `name="otp"`, `name="pla"`, `name="cenovnikid"`, `name="paritet"`, `name="izjizv"`, `name="mi"`, `name="komsifra"`,
		`name="idorgjed"`, `name="mestotrid"`, `name="foot"`, `name="posuslnab"`, `name="keykat"`, `name="oznakakat"`, `name="napomporeskoosl"`,
		`name="sifval"`, `name="kurs"`, `name="vaziza"`, `name="kamt"`, `name="brugov"`, `name="rdokid"`, `name="rproid"`, `name="konto"`,
		`name="sifraartikla"`, `name="otk"`, `name="serija"`, `name="roktr"`, `name="kolicina"`, `name="cena"`, `name="rabat"`,
		`id="kom-pro-potvrda-btn"`, `id="kom-pro-nalog-btn"`, `/api/search/artikli-zalihe`, `/api/search/poreske-kategorije`, `id="prosastanjem"`}
	for _, tt := range []struct {
		name     string
		rdokID   int64
		snimljen bool
		tipzal   int
		tipcene  int
	}{
		{name: "new profaktura", rdokID: 0, tipzal: 1, tipcene: 1},
		{name: "saved one in a magacin with the lots and a cenovnik", rdokID: 45, snimljen: true, tipzal: 2, tipcene: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomProDokument{
				Header: domain.KomProHeaderView{KomProfakturaRow: domain.KomProfakturaRow{RdokID: tt.rdokID, Vrd: 360, Dokum: 7, MagaciniID: 3,
					Tipzal: tt.tipzal, Tipcene: tt.tipcene, SifVal: 941, Kurs: 1, VaziZa: 1}, Datum: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
					Snimljen: tt.snimljen},
				Vrste:    []domain.ComboItem{{Key: "360", Value: "360 - Predračun"}},
				Tipdok:   []domain.ComboItem{{Key: "43", Value: "43 - Mešovito"}},
				Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Veleprodaja"}},
				NoviRbr:  1,
			}
			var buf bytes.Buffer
			if err := tmpl_kom.ProfakturaBody(proUI(), proKomUI(), dok, domain.TableData{}, proButtons(), proDugmadZa(tt.rdokID), tmpl_kom.ProfakturaVarijanta(), tr).
				Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			if got := zahtevOtvoren(zahtevPanel(t, html, proHeaderPanelID)); got == tt.snimljen {
				t.Errorf("header open = %v", got)
			}
			for _, id := range []string{"tipdok", "magaciniid", "vrd"} {
				sel := regexp.MustCompile(`<select[^>]*id="` + id + `"[^>]*>`).FindString(html)
				if got := strings.Contains(sel, "disabled"); got != (tt.rdokID > 0) {
					t.Errorf("%s disabled = %v: %s", id, got, sel)
				}
			}
			if got := strings.Contains(html, `id="prolot"`); got != (tt.tipzal == 2) {
				t.Errorf("lot shown = %v", got)
			}
			cenovnik := regexp.MustCompile(`<input[^>]*id="cenovnikid-combo-input"[^>]*>`).FindString(html)
			if got := strings.Contains(cenovnik, "disabled"); got != (tt.tipcene != 2) {
				t.Errorf("cenovnik disabled = %v: %s", got, cenovnik)
			}
			komStampaProvera(t, html, "kom-pro-stampa", proNazadID, tt.rdokID == 0)
		})
	}
}

// TestProStampaRender checks the domestic print (the totals with the PDV, slovima, the poresko oslobođenje) and the
// izvozna one (in English, the valuta, the declaration of the exporter).
func TestProStampaRender(t *testing.T) {
	tr := i18n.GetInstance()
	datum := sql.NullTime{Time: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Valid: true}
	p := domain.KomProStampa{Dokum: 7, Datum: datum, Valuta: datum, Kupac: domain.KomPonudaStampaDobavljac{Konto: "2040", Sifra: "00031", Naziv: "KUPAC DOO"},
		MestoIsporuke: "Magacin kupca", MestoIsporuke2: "Put 1, 21000 Novi Sad", Stavke: []domain.KomProStampaStavka{{Rbr: 1, Sifra: 5, Naziv: "BRAŠNO",
			JM: "KG", Kolicina: 10, Cena: 100, ProcRab: 5, Rabat: 50, Stopa: 20, PdvJM: 18.4, Iznos: 1000}}, Svega: 1000, Rabat: 50, NetoPrikaz: true,
		Neto: 950, Ugrabat: 95, UgrabatPct: 10, PdvPoStopama: []domain.KomProStampaPdv{{Stopa: 20, Osnovica: 855, Pdv: 171}}, ZaNaplatu: 1026,
		Slovima: "hiljadu dvadeset šest", Pornapomena: "član 24 stav 1", NetoTezina: 10}
	var buf bytes.Buffer
	if err := tmpl_rep_kom.ProfakturaStampa(domain.ReportParameters{}, p, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, s := range []string{"KUPAC DOO", "M. Isporuke:Magacin kupca", "BRAŠNO", "hiljadu dvadeset šest", "član 24 stav 1", "09.03.2026"} {
		if !strings.Contains(html, s) {
			t.Errorf("the print has no %s", s)
		}
	}
	p.Ino, p.OznVal, p.Izjizv, p.PdvPoStopama, p.Slovima = true, "eur", true, nil, ""
	buf.Reset()
	if err := tmpl_rep_kom.ProfakturaStampa(domain.ReportParameters{}, p, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html = buf.String()
	for _, s := range []string{"PROFORMA No.:", "TOTAL FOR PAYMENT EUR", "Delivery. Place:Magacin kupca", "The exporter of the products"} {
		if !strings.Contains(html, s) {
			t.Errorf("the izvozna print has no %s", s)
		}
	}
}
