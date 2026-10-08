package handler

import (
	"net/http"
	"strings"

	"helia/frontend/components"
	"helia/i18n"
	commonsvc "helia/internal/service/common"

	"github.com/gin-gonic/gin"
)

// searchURL is the endpoint of the shared searches: GET /api/search/:entity (partneri, konta,
// artikli, ...; see commonsvc.SearchService). It is the SearchURL of a components.TableCombo, e.g.
// "/api/search/partneri".
const searchURL = "/api/search/:entity"

// SearchHandler answers the searches of components.TableCombo for every screen, like CommonService
// serves the combos: one endpoint, the searches are defined in the SearchService.
type SearchHandler struct {
	service    commonsvc.SearchService
	translator *i18n.Service
}

// NewSearchHandler creates the handler of the shared searches.
func NewSearchHandler(service commonsvc.SearchService, translator *i18n.Service) *SearchHandler {
	return &SearchHandler{service: service, translator: translator}
}

// Search answers GET /api/search/:entity with the rows that match the text of the combo (the
// parameter q, or the one named by "param") and the filters of the request (the other parameters,
// e.g. konto from the HxVals of the combo), rendered by components.TableComboResults. The id of the
// combo is the parameter "combo" or else the id of the input that fired the request (HX-Trigger,
// "<combo>-input").
func (h *SearchHandler) Search(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Request.URL.Query()
	param := query.Get("param")
	if param == "" {
		param = "q"
	}
	filters := map[string]string{}
	for name, values := range query {
		if name != param && name != "combo" && name != "param" && len(values) > 0 {
			filters[name] = values[0]
		}
	}
	res, err := h.service.Search(ctx, c.Param("entity"), query.Get(param), filters)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	res.ComboID = query.Get("combo")
	if res.ComboID == "" {
		res.ComboID = strings.TrimSuffix(c.GetHeader("HX-Trigger"), "-input")
	}
	if err := components.TableComboResults(res, h.translator).Render(ctx, c.Writer); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// AddRoutes registers the endpoint of the shared searches.
func (h *SearchHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/search/:entity", h.Search)
}
