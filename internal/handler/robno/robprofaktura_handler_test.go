package robno

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRobnoProRoutes checks that the urls the "Profakture" screen sends its requests to (the constants of the
// buttons, of the grids and of the script) are routes of robnoProRoutes, and that the group PRO opens the screen.
func TestRobnoProRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	(&RobnoDokumentaUnosHandler{}).robnoProRoutes(r)
	routes := map[string]bool{}
	for _, rt := range r.Routes() {
		routes[rt.Method+" "+rt.Path] = true
	}
	tests := []struct {
		method, path string
	}{
		{http.MethodGet, robnoProDokumentURL + "/:id"},
		{http.MethodGet, robnoProNoviURL},
		{http.MethodPost, robnoProSaveURL},
		{http.MethodGet, robnoProPotvrdaURL},
		{http.MethodDelete, robnoProBrisiURL + "/:id"},
		{http.MethodGet, robnoProArtikalURL},
		{http.MethodGet, robnoProArtikalURL + "/stopa"},
		{http.MethodGet, robnoProStavkaURL},
		{http.MethodPost, robnoProStavkaSaveURL},
		{http.MethodGet, robnoProStavkaPotvrdaURL},
		{http.MethodDelete, robnoProStavkaBrisiURL + "/:id"},
		{http.MethodGet, robnoProPrikazURL},
		{http.MethodGet, robnoProPrikazDokumentaURL},
		{http.MethodGet, robnoProPrikazStavkeURL + "/:id"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if !routes[tt.method+" "+tt.path] {
				t.Errorf("route %s %s is not registered", tt.method, tt.path)
			}
		})
	}
	if ekran := robnoDokumentaUnosEkrani["PRO"]; ekran.open == nil {
		t.Errorf("the group PRO has no screen")
	}
}
