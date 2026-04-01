package finansijsko

import (
	"fmt"
	"helia/config"
	"net/http"

	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	dnevnikContentTitle string = "DNEVNIK KNJIŽENJA"
	dnevnikTableID      string = "dnevnik-table"
	dnevnikURLPrefix    string = "/api/dnevnik/"
	dnevnikURLMain      string = "/api/dnevnik"
)

type DnevnikHandler struct {
	cfg     config.Config
	service *finservice.DnevnikResource
}

const (
	hxValsDnevnik = `js:{
            "oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
			"query": document.getElementById("search-input")?.value
        }`
)

func NewDnevnikHandler(service *finservice.DnevnikResource, cfg config.Config) *DnevnikHandler {
	handler := &DnevnikHandler{
		cfg:     cfg,
		service: service,
	}
	return handler
}

// DnevnikKnjizenja handles the dnevnik knjizenja view
func (h *DnevnikHandler) DnevnikKnjizenja(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")

	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}

	tbl := common.SetTableBasicData(dnevnikContentTitle, dnevnikTableID, h.service.GetDnevnikTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, "", dnevnikURLMain, false, false, false)
	searchInput := common.CreateSearchInput("search-input", i18n.GetInstance(), dnevnikURLMain, fmt.Sprintf("#%s", dnevnikTableID), hxValsDnevnik)
	tbl.HasTotals = true
	if requestSource == "menu" || requestSource == "" {
		btnObrada := common.SetButton("btnobrada", "Obrada", "fin_obrada", dnevnikURLMain, fmt.Sprintf("#%s", dnevnikTableID), "innerHTML", "GET", "", hxValsDnevnik, true, common.ClassSaveButton, "handleDialogResponse")
		btnPrint := common.SetButton("stampa", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true, common.ClassPrintButton, "")

		err := tmpl_fin.DnevnikKnjizenja(tbl, searchInput, btnObrada, btnPrint, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
		return
	}

	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		//validacija input parametre:
		fieldParameters := []string{"oddatuma", "dodatuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		odDatuma := c.Query("oddatuma")
		doDatuma := c.Query("dodatuma")
		searchText := c.Query("search-input")
		ctx := c.Request.Context()

		err := h.service.GetDnevnikKnjizenja(ctx, &tbl, true, page, pageSize, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.Pagination.HxVals = hxValsDnevnik
		err = h.service.GetDnevnikKnjizenja(ctx, &tbl, false, page, pageSize, odDatuma, doDatuma, searchText)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// AddRoutes registers all dnevnik routes
func (h *DnevnikHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for dnevnik
	r.GET("api/dnevnik", h.DnevnikKnjizenja)
}
