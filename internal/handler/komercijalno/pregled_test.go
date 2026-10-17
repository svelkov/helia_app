package komercijalno

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"helia/config"
	tmpl_kom "helia/frontend/templates/komercijalno"
	"helia/i18n"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

// TestKomLista checks the list of a main screen: "Izmeni" (the entry dialog) and "Briši" (the question of the deletion)
// of every nepotvrđen document, none of the potvrđeni ones (they cannot be changed).
func TestKomLista(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	ui := zahtevPregledUI()
	rows := []domain.TableRow{{ID: "9", Fields: []string{"1-Sirovine", "410", "12", "02.03.2026", "", "", ""}}}
	for _, tt := range []struct {
		name         string
		nepotvrdjeni bool
	}{
		{name: "nepotvrđeni", nepotvrdjeni: true},
		{name: "potvrđeni", nepotvrdjeni: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tbl := komLista(config.Config{}, ui, zahtevListaTableID, nil, rows, tt.nepotvrdjeni, zahtevBrisiPotvrdaURL, zahtevPrikazStavkeURL)
			tbl.Headers = []domain.Fields{{Name: "magacin"}, {Name: "vrd"}, {Name: "dokum"}, {Name: "dadok"}, {Name: "foot"}, {Name: "ref"}, {Name: "oj"}}
			var buf bytes.Buffer
			if err := tmpl_kom.KomPregledLista(tbl, i18n.GetInstance()).Render(context.Background(), &buf); err != nil {
				t.Fatal(err)
			}
			html := buf.String()
			for _, s := range []string{`hx-get="` + zahtevDialogURL + `"`, `hx-get="` + zahtevBrisiPotvrdaURL + `"`} {
				if got := strings.Contains(html, s); got != tt.nepotvrdjeni {
					t.Errorf("%s in the list = %v, want %v", s, got, tt.nepotvrdjeni)
				}
			}
			if !strings.Contains(html, zahtevPrikazStavkeURL+"/9") || !strings.Contains(html, "komPregledIzbor(this)") {
				t.Errorf("a row does not load its stavke or select it for the print")
			}
		})
	}
}

// TestKomFilteri checks the filters of the list read from the request: the potvrđeni or else the nepotvrđeni
// documents, all the magacini, the search.
func TestKomFilteri(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		query string
		want  komPregledFilteri
	}{
		{query: "pvrd=410&pmagaciniid=3&pdoksts=40&psvimagacini=true&query=+hitno+",
			want: komPregledFilteri{Vrd: 410, MagaciniID: 3, Doksts: domain.KomZahtevPotvrdjen, SviMagacini: true, Tekst: "hitno"}},
		{query: "pvrd=410&pdoksts=xx", want: komPregledFilteri{Vrd: 410, Doksts: domain.KomZahtevNepotvrdjen}},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/lista?"+tt.query, nil)
		if got := komFilteri(c); got != tt.want {
			t.Errorf("%s: got %+v, want %+v", tt.query, got, tt.want)
		}
	}
}

// TestKomDialogOznaka checks that the entry dialog of every screen carries data-kom-dialog: the main screen reloads its
// list when that dialog is hidden (closed by "Nazad", its close button or Escape).
func TestKomDialogOznaka(t *testing.T) {
	if err := i18n.Init("../../../i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	for _, tt := range []struct {
		name, dialogID string
		render         func(*bytes.Buffer) error
	}{
		{name: "zahtev za nabavku", dialogID: zahtevDialogID, render: func(b *bytes.Buffer) error {
			return tmpl_kom.ZahtevNabavkaDialog(zahtevUI(), domain.KomZahtevDokument{}, domain.TableData{}, zahtevButtons(), zahtevPotvrdaButton(0),
				domain.Button{}, tr, "").Render(context.Background(), b)
		}},
		{name: "zahtev za ponudu", dialogID: ponudaDialogID, render: func(b *bytes.Buffer) error {
			return tmpl_kom.ZahtevPonudaDialog(ponudaUI(), ponudaKomUI(), domain.KomPonudaDokument{}, domain.TableData{}, domain.TableData{},
				ponudaButtons(), ponudaPotvrdaButton(0), ponudaPreuzmiButton(), domain.Button{}, tr, "").Render(context.Background(), b)
		}},
		{name: "porudžbenica dobavljaču", dialogID: porDialogID, render: func(b *bytes.Buffer) error {
			return tmpl_kom.PorudzbenicaDialog(porUI(), porKomUI(), domain.KomPorDokument{}, domain.TableData{}, domain.TableData{}, porButtons(),
				porDugmadZa(0), domain.Button{}, tr, "").Render(context.Background(), b)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tt.render(&buf); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(buf.String(), `id="`+tt.dialogID+`" data-kom-dialog="true"`) {
				t.Errorf("the dialog %s has no data-kom-dialog", tt.dialogID)
			}
		})
	}
}
