package robno

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRobnoKnpRoutes checks that the urls the "Knjižna pisma" screen sends its requests to (the constants of the
// buttons, of the grids and of the script) are routes of robnoKnpRoutes, and that the groups KNO and KNZ open the
// screen.
func TestRobnoKnpRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&RobnoDokumentaUnosHandler{}).robnoKnpRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	tests := []struct {
		method, path string
	}{
		{http.MethodGet, robnoKnpDokumentURL + "/:id"},
		{http.MethodGet, robnoKnpNoviURL},
		{http.MethodPost, robnoKnpSaveURL},
		{http.MethodGet, robnoKnpPotvrdaURL},
		{http.MethodDelete, robnoKnpBrisiURL + "/:id"},
		{http.MethodGet, robnoKnpVezniURL},
		{http.MethodGet, robnoKnpObracunURL},
		{http.MethodGet, robnoKnpStavkaURL},
		{http.MethodPost, robnoKnpStavkaSaveURL},
		{http.MethodPost, robnoKnpStavkaDodajURL + "/:id"},
		{http.MethodGet, robnoKnpStavkaPotvrdaURL},
		{http.MethodDelete, robnoKnpStavkaBrisiURL + "/:id"},
		{http.MethodPost, robnoKnpStavkeDodajURL},
		{http.MethodGet, robnoKnpStavkePotvrdaURL},
		{http.MethodDelete, robnoKnpStavkeBrisiURL + "/:id"},
		{http.MethodGet, robnoKnpPrikazURL},
		{http.MethodGet, robnoKnpPrikazDokumentaURL},
		{http.MethodGet, robnoKnpPrikazStavkeURL + "/:id"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if !routes[tt.method+" "+tt.path] {
				t.Errorf("route %s %s is not registered", tt.method, tt.path)
			}
		})
	}
	for _, grupa := range []string{"KNO", "KNZ"} {
		if ekran := robnoDokumentaUnosEkrani[grupa]; ekran.open == nil {
			t.Errorf("the group %s has no screen", grupa)
		}
	}
}

// TestRobnoDokumentaUnosAddRoutes checks that AddRoutes registers the routes of every screen of the entry whose
// group has an open function (a screen without its routes opens, but none of its buttons works).
func TestRobnoDokumentaUnosAddRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&RobnoDokumentaUnosHandler{}).AddRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	for _, url := range []string{robnoFaktureSaveURL, robnoPorSaveURL, robnoProSaveURL, robnoKnpSaveURL} {
		t.Run(url, func(t *testing.T) {
			if !routes[http.MethodPost+" "+url] {
				t.Errorf("route POST %s is not registered by AddRoutes", url)
			}
		})
	}
}

// TestRobnoSivEkran checks that the group SIV (the zaduženja sitnog inventara) opens the screen of the opšti
// dokumenti.
func TestRobnoSivEkran(t *testing.T) {
	if ekran := robnoDokumentaUnosEkrani["SIV"]; ekran.open == nil {
		t.Errorf("the group SIV has no screen")
	}
}
