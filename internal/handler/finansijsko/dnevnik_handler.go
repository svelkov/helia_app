package finansijsko

import (
	"fmt"
	"helia/config"
	"net/http"
	"time"

	tmpl_fin "helia/frontend/templates/finansijsko"
	tmpl_rep_fin "helia/frontend/templates/reports/finansijsko"
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
	dnevnikURLStampa    string = "/api/dnevnik/stampa"
)

type DnevnikHandler struct {
	cfg     config.Config
	service finservice.DnevnikService
}

const (
	hxValsDnevnik = `js:{
            "oddatuma": document.getElementById("oddatuma")?.value,
			"dodatuma": document.getElementById("dodatuma")?.value,
			"query": document.getElementById("search-input")?.value
        }`
)

func NewDnevnikHandler(service finservice.DnevnikService, cfg config.Config) *DnevnikHandler {
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
		btnPrint := domain.Button{
			Id:            "btn-print-dnevnik",
			IsVisible:     true,
			LabelText:     "Štampa ",
			BtnClass:      common.ClassPrintButton,
			HxActionURL:   dnevnikURLStampa,
			DataFields:    "oddatuma,dodatuma",
			HxRequestType: "GET",
		}

		btnObrada := common.SetButton("btnobrada", "Obrada", "fin_obrada", dnevnikURLMain, fmt.Sprintf("#%s", dnevnikTableID), "innerHTML", "GET", "", hxValsDnevnik, true, common.ClassSaveButton, "handleBackendResponse")
		tmpl_fin.DnevnikKnjizenja(tbl, searchInput, btnObrada, btnPrint, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
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
		searchText := c.Query("query")
		ctx := c.Request.Context()

		err := h.service.GetDnevnikKnjizenja(ctx, &tbl, true, page, pageSize, odDatuma, doDatuma, searchText, "O")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		tbl.Pagination.HxVals = hxValsDnevnik
		err = h.service.GetDnevnikKnjizenja(ctx, &tbl, false, page, pageSize, odDatuma, doDatuma, searchText, "O")
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// DnevnikKnjizenjaStampa renders a full-page printable Dnevnik knjizenja report.
func (h *DnevnikHandler) DnevnikKnjizenjaStampa(c *gin.Context) {
	ctx := c.Request.Context()
	translator := i18n.GetInstance()
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	odDatuma := c.Query("oddatuma")
	doDatuma := c.Query("dodatuma")

	// Format dates for display (input: 2006-01-02, output: DD.MM.YYYY)
	odDatumaFmt := odDatuma
	doDatumaFmt := doDatuma
	if t, err := time.Parse("2006-01-02", odDatuma); err == nil {
		odDatumaFmt = t.Format(common.DateLayout)
	}
	if t, err := time.Parse("2006-01-02", doDatuma); err == nil {
		doDatumaFmt = t.Format(common.DateLayout)
	}

	// Build table — fetch all records (no pagination for print)
	tbl := common.SetTableBasicData(dnevnikContentTitle, dnevnikTableID, h.service.GetDnevnikStampaTableFields(), "", "", 999999, 1, 1, 0, h.cfg)
	tbl.HasTotals = true

	if err := h.service.GetDnevnikKnjizenja(ctx, &tbl, true, 1, 999999, odDatuma, doDatuma, "", "S"); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	if err := h.service.GetDnevnikKnjizenja(ctx, &tbl, false, 1, 999999, odDatuma, doDatuma, "", "S"); err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	fvrData, err := h.service.GetFvrData(ctx)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgReadData+" "+err.Error())
		return
	}
	paramItems := make(map[string]domain.ParameterItem)
	params := domain.ReportParameters{
		ReportName:     translator.Title("DNEVNIK KNJIŽENJA"),
		CompanyName:    fvrData.Naziv,
		Adress:         fvrData.Adresa,
		Postcode:       fvrData.Pobro,
		City:           fvrData.Mesto,
		Orientation:    "landscape",
		ParameterItems: paramItems,
	}
	paramItems["OdDatuma"] = domain.ParameterItem{Name: translator.Label("Za period od"), Value: odDatumaFmt}
	paramItems["DoDatuma"] = domain.ParameterItem{Name: translator.Label("do"), Value: doDatumaFmt}
	params.ParameterItems = paramItems
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl_rep_fin.DnevnikKnjizenjaStampa(tbl, domain.TableData{}, params, translator).Render(ctx, c.Writer)
}

// AddRoutes registers all dnevnik routes
func (h *DnevnikHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for dnevnik
	r.GET("api/dnevnik", h.DnevnikKnjizenja)
	r.GET("api/dnevnik/stampa", h.DnevnikKnjizenjaStampa)
}
