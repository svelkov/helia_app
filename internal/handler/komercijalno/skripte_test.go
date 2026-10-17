package komercijalno

import (
	"bytes"
	"context"
	"strings"
	"testing"

	tmpl_kom "helia/frontend/templates/komercijalno"
	"helia/i18n"
	"helia/internal/domain"

	"github.com/a-h/templ"
)

// TestKomercijalnoSkripte checks that every screen of the komercijalno (opened from the menu into a page that loads
// no shared scripts) defines the functions its buttons and fields call: closeDialog ("Nazad" and the close button),
// clearFieldError, handleBackendResponse (the field errors of a save), openPrintWithParams (the print buttons) and
// komPregledUcitaj (the list of the main screen).
func TestKomercijalnoSkripte(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	for _, tt := range []struct {
		name   string
		screen templ.Component
	}{
		{name: "zahtev za nabavku", screen: tmpl_kom.ZahtevNabavkaPregled(zahtevPregledUI(), nil, nil, 0, 0, domain.TableData{},
			komPretraga(zahtevPregledUI(), tr), domain.TableData{}, tr)},
		{name: "zahtev za ponudu", screen: tmpl_kom.ZahtevPonudaPregled(ponudaPregledUI(), nil, nil, 0, 0, domain.TableData{},
			komPretraga(ponudaPregledUI(), tr), domain.TableData{}, tr)},
		{name: "porudžbenica dobavljaču", screen: tmpl_kom.PorudzbenicaPregled(porPregledUI(), porKomUI(), nil, nil, 0, 0, domain.TableData{},
			komPretraga(porPregledUI(), tr), domain.TableData{}, tr)},
		{name: "prijemnica materijala", screen: tmpl_kom.PrijemnicaPregled(prijPregledUI(), nil, 0, domain.TableData{},
			komPretraga(prijPregledUI(), tr), domain.TableData{}, tr)},
		{name: "profaktura", screen: tmpl_kom.ProfakturaPregled(proPregledUI(), nil, 0, domain.TableData{},
			komPretraga(proPregledUI(), tr), domain.TableData{}, tr)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tt.screen.Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}
			html := buf.String()
			for _, fn := range []string{"function closeDialog(", "function clearFieldError(", "function handleBackendResponse(",
				"function openPrintWithParams(", "function komPregledUcitaj("} {
				if !strings.Contains(html, fn) {
					t.Errorf("the screen does not define %s", strings.TrimSuffix(strings.TrimPrefix(fn, "function "), "("))
				}
			}
		})
	}
}

// TestPrijemnicaSkripta checks that the prijemnica defines the functions its buttons, grids and fields call (prijUI,
// prijButtons, prijEkranZa and the dialog "Pakovanje").
func TestPrijemnicaSkripta(t *testing.T) {
	tr := i18n.GetInstance()
	var buf bytes.Buffer
	if err := tmpl_kom.PrijemnicaPregled(prijPregledUI(), nil, 0, domain.TableData{}, komPretraga(prijPregledUI(), tr), domain.TableData{},
		tr).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	for _, fn := range []string{"function komPrijVeza(", "function komPrijIzbor(", "function komPrijStavkaUcitaj(", "function komPrijStavkaOdustani(",
		"function komPrijPakovanjeIzbor(", "function robnoFaktureAfterHeaderSave(", "function robnoUpisiBroj("} {
		if !strings.Contains(html, fn) {
			t.Errorf("the screen does not define %s", strings.TrimSuffix(strings.TrimPrefix(fn, "function "), "("))
		}
	}
	if strings.Contains(html, `id="pvrd"`) {
		t.Errorf("the main screen of the prijemnica has the filter of the vrsta")
	}
}

// TestProfakturaSkripta checks that the profaktura defines the functions its buttons, grids and fields call (proButtons,
// proEkranZa and the OnSelect of its combos).
func TestProfakturaSkripta(t *testing.T) {
	tr := i18n.GetInstance()
	var buf bytes.Buffer
	if err := tmpl_kom.ProfakturaPregled(proPregledUI(), nil, 0, domain.TableData{}, komPretraga(proPregledUI(), tr), domain.TableData{},
		tr).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	for _, fn := range []string{"function komProKupac(", "function komProMesto(", "function komProKategorija(", "function komProValuta(",
		"function komProCenovnik(", "function komProStavkaUcitaj(", "function komProStavkaOdustani(", "function robnoFaktureKategorijaSelected("} {
		if !strings.Contains(html, fn) {
			t.Errorf("the screen does not define %s", strings.TrimSuffix(strings.TrimPrefix(fn, "function "), "("))
		}
	}
}
