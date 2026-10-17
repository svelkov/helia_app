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
	"helia/internal/common"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestNalRoutes checks that the urls the screen sends its requests to are routes of AddRoutes.
func TestNalRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&NalogHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, nalURL},
		{http.MethodGet, nalListaURL},
		{http.MethodGet, nalDelimicniURL},
		{http.MethodGet, nalDelimicniListaURL},
		{http.MethodGet, nalDelimicniStavkeURL + "/:id"},
		{http.MethodGet, nalDialogURL},
		{http.MethodGet, nalNoviURL},
		{http.MethodPost, nalSaveURL},
		{http.MethodGet, nalBrisiPotvrdaURL},
		{http.MethodDelete, nalBrisiURL + "/:id"},
		{http.MethodGet, nalPotvrdaURL},
		{http.MethodPost, nalPotvrdiURL + "/:id"},
		{http.MethodGet, nalKupacURL},
		{http.MethodGet, nalMestoURL},
		{http.MethodGet, nalValutaURL},
		{http.MethodGet, nalCenovnikURL},
		{http.MethodGet, nalProfakturaURL},
		{http.MethodGet, nalSaldoURL},
		{http.MethodGet, nalArtikalURL},
		{http.MethodGet, nalStavkaURL},
		{http.MethodPost, nalStavkaSaveURL},
		{http.MethodGet, nalStavkaPotvrdaURL},
		{http.MethodDelete, nalStavkaBrisiURL + "/:id"},
		{http.MethodGet, nalPrikazStavkeURL + "/:id"},
		{http.MethodGet, nalStampaURL},
	} {
		if !routes[tt.method+" "+tt.path] {
			t.Errorf("route %s %s is not registered", tt.method, tt.path)
		}
	}
}

// TestNalBodyRender checks the entry dialog of a nalog: the fields of the profaktura with the ones of the nalog (the
// profaktura, fixed once the nalog has one; the bank; the status; the finansijsko stanje; "Skidaj zalihe po starosti"
// only in a magacin with the lots), the urls of the screen, no "Form. nalog".
func TestNalBodyRender(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	for _, tt := range []struct {
		name    string
		rdokID  int64
		vrdokID int64
		tipzal  int
	}{
		{name: "new nalog", tipzal: 1},
		{name: "saved one of a profaktura in a magacin with the lots", rdokID: 50, vrdokID: 46, tipzal: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dok := domain.KomNalogDokument{
				KomProDokument: domain.KomProDokument{
					Header: domain.KomProHeaderView{KomProfakturaRow: domain.KomProfakturaRow{RdokID: tt.rdokID, Vrd: 350, Dokum: 3, MagaciniID: 3,
						Tipzal: tt.tipzal, SifVal: 941, Kurs: 1, VaziZa: 1, Doksts: "10"}, Datum: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)},
					Vrste: []domain.ComboItem{{Key: "350", Value: "350 - Porudžbenica kupca"}}, Magacini: []domain.ComboItem{{Key: "3", Value: "1 - Roba"}},
					NoviRbr: 1, BrojLabel: "Broj porudžbenice", DatumLabel: "Datum porudžbenice",
				},
				Dodatno:    domain.KomNalogDodatno{VrdokID: tt.vrdokID, ProDokum: 7, SifBank: 2, BnkizvID: 4, BnkDes: "Banka"},
				Saldo:      domain.KomNalogSaldo{UkSaldo: 1234.5},
				StatusOpis: "Unet dokument, nepotvrđen nalog",
			}
			var buf bytes.Buffer
			if err := tmpl_kom.ProfakturaBody(proUIZa(nalNoviURL, nalArtikalURL), nalKomUI(), dok.KomProDokument, domain.TableData{},
				proButtonsZa(nalSaveURL, nalStavkaSaveURL), nalDugmadZa(tt.rdokID), tmpl_kom.NalogVarijanta(dok, nalNalogUI(), tr), tr).
				Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, s := range []string{`name="profakturaid"`, `name="bnkizvid"`, `/api/search/profakture-potvrdjene`, `/api/search/banke-izvoz`,
				`id="nalstatus"`, `id="nalsaldo-uksaldo"`, `id="kom-nal-saldo"`, `data-profaktura-url="` + nalProfakturaURL + `"`,
				`data-saldo-url="` + nalSaldoURL + `"`, `data-kupac-url="` + nalKupacURL + `"`, nalSaveURL, nalStavkaSaveURL, `id="kom-pro-potvrda-btn"`,
				`2 - Banka`} {
				if !strings.Contains(html, s) {
					t.Errorf("the body has no %s", s)
				}
			}
			if strings.Contains(html, `id="kom-pro-nalog-btn"`) || strings.Contains(html, proSaveURL) {
				t.Errorf("the nalog has the buttons of the profaktura")
			}
			pro := regexp.MustCompile(`<input[^>]*id="profakturaid-combo-input"[^>]*>`).FindString(html)
			if got := strings.Contains(pro, "disabled"); got != (tt.vrdokID > 0) {
				t.Errorf("profaktura disabled = %v: %s", got, pro)
			}
			if got := strings.Contains(html, `name="postarosti"`); got != (tt.tipzal == 2) {
				t.Errorf("po starosti shown = %v", got)
			}
			komStampaProvera(t, html, "kom-pro-stampa", proNazadID, tt.rdokID == 0)
		})
	}
}

// TestNalPregledRender checks the two main screens: their tabs, the filter of the status only on the first one, "Dodaj"
// only on the first one, the functions of the script of the nalog.
func TestNalPregledRender(t *testing.T) {
	tr := i18n.GetInstance()
	for _, delimicni := range []bool{false, true} {
		ui := nalPregledUI(delimicni)
		tbl := domain.TableData{}
		komDodajStampa(&tbl, ui, nalPregledStampaID, nalStampaURL, nalPregledRdokID)
		if delimicni {
			tbl.BtnAdd = domain.Button{}
		}
		var buf bytes.Buffer
		if err := tmpl_kom.NalogPregled(ui, "", nil, 0, tbl, komPretraga(ui, tr), domain.TableData{}, tr).Render(context.Background(), &buf); err != nil {
			t.Fatalf("render: %v", err)
		}
		html := buf.String()
		for _, s := range []string{`hx-get="` + nalURL + `"`, `hx-get="` + nalDelimicniURL + `"`, "function komNalProfaktura(", "function komNalSaldo(",
			"function komProKupacPolja(", "function komPregledUcitaj(", `data-lista-url="` + ui.ListaURL + `"`} {
			if !strings.Contains(html, s) {
				t.Errorf("delimični %v: the screen has no %s", delimicni, s)
			}
		}
		if got := strings.Contains(html, `id="pdoksts10"`); got == delimicni {
			t.Errorf("delimični %v: the filter of the status shown = %v", delimicni, got)
		}
	}
}

// TestNalStampaRender checks the print: the title with the broj, the kupac, the columns of the lots, the totals and
// "Nepotvrđen dokument!".
func TestNalStampaRender(t *testing.T) {
	tr := i18n.GetInstance()
	p := domain.KomNalogStampa{Dokum: 3, Datum: sql.NullTime{Time: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Valid: true},
		Kupac: domain.KomPonudaStampaDobavljac{Konto: "2040", Sifra: "00031", Naziv: "KUPAC DOO"}, Magacin: "30-Roba", OJ: "1 - Prodaja",
		KlasaCena: "3- VP", Tezina: 6, Zapremina: 1.25, Lot: true, Nepotvrdjen: true,
		Stavke: []domain.KomNalogStampaStavka{{Rbr: 1, Sifra: 1001, Opis: "BRAŠNO", Otk: "001", Serija: "S1", Rok: "31.12.2026", JM: "KG", Kolicina: 4}}}
	var buf bytes.Buffer
	if err := tmpl_rep_kom.NalogStampa(domain.ReportParameters{}, p, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, s := range []string{"KUPAC DOO", "BRAŠNO", "31.12.2026", "09.03.2026", "30-Roba", "1 - Prodaja", "3- VP", "__________",
		common.FormatNumberWithSystemLocale(1.25, 3)} {
		if !strings.Contains(html, s) {
			t.Errorf("the print has no %s", s)
		}
	}
	if !strings.Contains(html, `colspan="9"`) {
		t.Errorf("the print has no columns of the lots")
	}
	p.Lot, p.Nepotvrdjen = false, false
	buf.Reset()
	if err := tmpl_rep_kom.NalogStampa(domain.ReportParameters{}, p, tr).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	if html := buf.String(); !strings.Contains(html, `colspan="6"`) || strings.Contains(html, ">S1<") {
		t.Errorf("the print without the lots has their columns")
	}
}
