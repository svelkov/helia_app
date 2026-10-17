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

// TestOtpRoutes checks that the urls the screen sends its requests to are routes of AddRoutes.
func TestOtpRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&OtpremnicaHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, otpURL},
		{http.MethodGet, otpListaURL},
		{http.MethodGet, otpDialogURL},
		{http.MethodGet, otpNoviURL},
		{http.MethodPost, otpSaveURL},
		{http.MethodGet, otpBrisiPotvrdaURL},
		{http.MethodDelete, otpBrisiURL + "/:id"},
		{http.MethodGet, otpPotvrdaURL},
		{http.MethodPost, otpPotvrdiURL + "/:id"},
		{http.MethodGet, otpNalogURL},
		{http.MethodGet, otpValutaURL},
		{http.MethodGet, otpStavkaURL},
		{http.MethodPost, otpStavkaSaveURL},
		{http.MethodGet, otpPrikazStavkeURL + "/:id"},
		{http.MethodGet, otpStampaURL},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestOtpBodyRender checks the content of the entry dialog: the fields the service binds, the magacin, the vrsta and the
// nalog fixed for a saved otpremnica, the lot only in a magacin with the lots, the valuta only for a foreign kupac, the
// nastanak and the opis outside the fields of the header (the potvrda sends them), "Štampa" before "Nazad".
func TestOtpBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	fields := []string{`name="magaciniid"`, `name="vrd"`, `name="nalogid"`, `name="brotp"`, `name="datotp"`, `name="brfakture"`, `name="datum"`,
		`name="dozvolibroj"`, `name="foot"`, `name="napomena"`, `name="keykat"`, `name="oznakakat"`, `name="napomporeskoosl"`, `name="posuslnab"`,
		`name="vozac"`, `name="brvozila"`, `name="otp"`, `name="brugov"`, `name="bttowght"`, `name="netwght"`, `name="totboxs"`, `name="brpaleta"`,
		`name="popdv"`, `name="sifval"`, `name="kurs"`, `name="vaziza"`, `name="nastanak"`, `name="opisdok"`, `name="rdokid"`, `name="rproid"`,
		`name="kolicina"`, `/api/search/nalozi-potvrdjeni`, `id="kom-otp-potvrda-btn"`, `data-nalog-url="` + otpNalogURL + `"`}
	for _, tt := range []struct {
		name   string
		rdokID int64
		tipzal int
		ino    bool
	}{
		{name: "new otpremnica", tipzal: 1},
		{name: "saved one of a foreign kupac in a magacin with the lots", rdokID: 60, tipzal: 2, ino: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomOtpDokument{
				Header: domain.KomOtpHeaderView{KomOtpremnicaRow: domain.KomOtpremnicaRow{RdokID: tt.rdokID, Vrd: 188, Brotp: 990001, Dokum: 990001,
					MagaciniID: 3, Tipzal: tt.tipzal, SifVal: 941, Kurs: 1, VaziZa: 1, Doksts: "10"}, Datum: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
					Ino: tt.ino, PdvObveznik: true, Snimljen: tt.rdokID > 0},
				Vrste:    []domain.ComboItem{{Key: "188", Value: "188 - Faktura"}},
				Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Proizvodi"}},
				Nastanci: []domain.ComboItem{{Key: "35", Value: "Datum prometa"}}, Nastanak: "35",
			}
			var buf bytes.Buffer
			if err := tmpl_kom.OtpremnicaBody(otpUI(otpPutanjeOtp), otpKomUI(otpPutanjeOtp), dok, domain.TableData{}, otpButtons(otpPutanjeOtp), otpDugmadZa(otpPutanjeOtp, tt.rdokID), tr).
				Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, f := range fields {
				if !strings.Contains(html, f) {
					t.Errorf("the body has no %s", f)
				}
			}
			for _, id := range []string{"magaciniid", "vrd"} {
				sel := regexp.MustCompile(`<select[^>]*id="` + id + `"[^>]*>`).FindString(html)
				if got := strings.Contains(sel, "disabled"); got != (tt.rdokID > 0) {
					t.Errorf("%s disabled = %v: %s", id, got, sel)
				}
			}
			if got := strings.Contains(html, `name="dozvolilot"`); got != (tt.tipzal == 2) {
				t.Errorf("lot shown = %v", got)
			}
			ino := regexp.MustCompile(`<div id="kom-otp-ino"[^>]*>`).FindString(html)
			if got := !strings.Contains(ino, "hidden"); got != tt.ino {
				t.Errorf("valuta shown = %v: %s", got, ino)
			}
			polja := html[strings.Index(html, "data-panel-fields"):]
			if strings.Index(polja, `name="nastanak"`) < strings.Index(polja, `name="brpaleta"`) {
				t.Errorf("the nastanak is among the fields of the header")
			}
			komStampaProvera(t, html, "kom-otp-stampa", otpNazadID, tt.rdokID == 0)
		})
	}
}

// TestOtpPregledRender checks the main screen: the komercijalni opis of the print and the functions of the script.
func TestOtpPregledRender(t *testing.T) {
	tr := i18n.GetInstance()
	ui := otpPregledUI(otpPutanjeOtp)
	var buf bytes.Buffer
	if err := tmpl_kom.OtpremnicaPregled(ui, otpStampaKomercopisID, nil, 0, domain.TableData{}, komPretraga(ui, tr), domain.TableData{}, false, "", tr).
		Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	for _, s := range []string{`id="` + otpStampaKomercopisID + `"`, "function komOtpNalog(", "function komOtpStavkaUcitaj(", "function komOtpStavkaOdustani(",
		"function komOtpKategorija(", "function komOtpValuta(", "function robnoFaktureAfterHeaderSave(", "function komPregledUcitaj("} {
		if !strings.Contains(html, s) {
			t.Errorf("the screen has no %s", s)
		}
	}
}

// TestOtpStampaRender checks the domestic print (the title, the nalog, the stavke, the totals, the fixed texts) and the
// one in a foreign valuta (in English, the currency, the total in RSD).
func TestOtpStampaRender(t *testing.T) {
	tr := i18n.GetInstance()
	p := domain.KomOtpStampa{Brotp: 15, Datotp: sql.NullTime{Time: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), Valid: true}, NalogDokum: 3,
		Kupac: domain.KomPonudaStampaDobavljac{Naziv: "KUPAC DOO"}, Magacin: "30-Proizvodi", Bttowght: 12.5, Pornapomena: "član 24",
		Nepotvrdjen: true, Stavke: []domain.KomOtpStampaStavka{{Rbr: 1, Sifra: 1001, Naziv: "BRAŠNO", JM: "KG", Otk: "001", Kolicina: 6, Cena: 100,
			Iznos: 600}}, Svega: 600, UkupnoRSD: 600}
	var buf bytes.Buffer
	if err := tmpl_rep_kom.OtpremnicaStampa(domain.ReportParameters{}, p, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, s := range []string{"KUPAC DOO", "BRAŠNO", "001", "10.03.2026", "30-Proizvodi", "SUP", "član 24", `colspan="10"`} {
		if !strings.Contains(html, s) {
			t.Errorf("the print has no %s", s)
		}
	}
	p.Ino, p.Storno, p.OznVal, p.SifVal, p.Kurs, p.VaziZa = true, true, "eur", 978, 117.25, 1
	buf.Reset()
	if err := tmpl_rep_kom.OtpremnicaStampa(domain.ReportParameters{}, p, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html = buf.String()
	for _, s := range []string{"CUSTOMER", "Despatch note:", "978 - EUR EXCHANGE RATE: 1 EUR = 117.2500", "Total RSD", "Goods received by", `colspan="9"`} {
		if !strings.Contains(html, s) {
			t.Errorf("the izvozna print has no %s", s)
		}
	}
}
