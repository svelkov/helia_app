package finansijsko

import (
	"helia/config"
	"helia/pkg/utils"
	"net/http"

	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	kamateContentTitle  string = "KAMATE"
	kamateTableID       string = "kamatetable"
	kamateURLPrefix     string = "/api/kamate/"
	kamateURLStope      string = "/api/kamate/stope"
	kamateURLFormiranje string = "/api/kamate/formiranje"
	kamateURLObracun    string = "/api/kamate/obracun"
)

const (
	hxValsStope = `js:{
            tip_kamate: document.getElementById("tip_kamate")?.value,
            stopa_od: document.getElementById("stopa_od")?.value,
            stopa_do: document.getElementById("stopa_do")?.value
        }`
	hxValsFormiranje = `js:{
            konto: document.getElementById("konto")?.value,
            od_sifre: document.getElementById("od_sifre")?.value,
            do_sifre: document.getElementById("do_sifre")?.value,
            od_datuma: document.getElementById("od_datuma")?.value,
            do_datuma: document.getElementById("do_datuma")?.value
        }`
	hxValsObracun = `js:{
            od_broja_liste: document.getElementById("od_broja_liste")?.value,
            do_broja_liste: document.getElementById("do_broja_liste")?.value,
            pod_datumom: document.getElementById("pod_datumom")?.value
        }`
)

type KamateHandler struct {
	tabData domain.TabData
	service finservice.KamateService
	cfg     config.Config
}

func NewKamateHandler(service finservice.KamateService, cfg config.Config) *KamateHandler {
	handler := &KamateHandler{
		cfg: cfg,
	}
	handler.tabData = GetKamateTabData()
	handler.service = service
	return handler
}

func (h *KamateHandler) KamateMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}

	tbl := common.SetTableBasicData(kamateContentTitle, kamateTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, kamateContentTitle, "", false, false, false)

	err := tmpl_fin.KamateMain(h.tabData, tbl, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

func (h *KamateHandler) KamatneStope(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		tbl := common.SetTableBasicData(kamateContentTitle, kamateTableID, h.service.GetTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "KAMATNE STOPE", "", false, false, false)

		h.tabData = setKamateActiveTab(h.tabData, "stope")
		err := tmpl_fin.KamatneStope(h.tabData, tbl, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", kamateTableID, h.service.GetTableFields(), "", kamateURLStope, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", kamateURLStope, false, false, false)
		err := h.service.GetKamatneStope(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetKamatneStope(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *KamateHandler) FormiranjeLista(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		tbl := common.SetTableBasicData(kamateContentTitle, kamateTableID, h.service.GetFormiranjeLisovaTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "PREGLED DOKUMENATA ZA FORMIRANJE KAMATNIH LISTOVA", kamateURLFormiranje, false, false, false)

		h.tabData = setKamateActiveTab(h.tabData, "formiranje")
		err := tmpl_fin.FormiranjeLista(h.tabData, tbl, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" || requestSource == "searchinput" {
		fieldParameters := []string{"konto", "od_datuma", "do_datuma"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", kamateTableID, h.service.GetFormiranjeLisovaTableFields(), "", kamateURLFormiranje, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", kamateURLFormiranje, false, false, false)
		err := h.service.GetFormiranjeLista(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetFormiranjeLista(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

func (h *KamateHandler) ObracunKamate(c *gin.Context) {
	requestSource := c.Request.Header.Get("X-Request-Source")
	translator := i18n.GetInstance()

	if requestSource == "menu" || requestSource == "tab" {
		session := domain.GetSessionFromContext(c)
		gnGod := 0
		if session != nil {
			gnGod = session.SelectedGod
		}

		tbl := common.SetTableBasicData(kamateContentTitle, kamateTableID, h.service.GetObracunTableFields(), "", "", 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "OBRACUN KAMATE", kamateURLObracun, false, false, false)

		h.tabData = setKamateActiveTab(h.tabData, "obracun")
		err := tmpl_fin.ObracunKamate(h.tabData, tbl, translator, gnGod).Render(c.Request.Context(), c.Writer)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}
	}
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		fieldParameters := []string{"pod_datumom"}
		fieldsError := common.ValidateRequiredParams(c, fieldParameters)
		if len(fieldsError) > 0 {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgValidation)
			return
		}

		page, pageSize := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		tbl := common.SetTableBasicData("", kamateTableID, h.service.GetObracunTableFields(), "", kamateURLObracun, 0, 0, 0, 0, h.cfg)
		common.SetTableConfig(&tbl, "", kamateURLObracun, false, false, false)
		err := h.service.GetObracunKamate(c, &tbl, true, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgGetTotalRecords)
			return
		}
		err = h.service.GetObracunKamate(c, &tbl, false, pageSize, page)
		if err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		utils.RenderContent(c, tbl)
	}
}

// RegisterRoutes registers the routes for the Kamate handler
func (h *KamateHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/kamate", h.KamateMain)
	r.GET("api/kamate/stope", h.KamatneStope)
	r.GET("api/kamate/formiranje", h.FormiranjeLista)
	r.GET("api/kamate/obracun", h.ObracunKamate)
}

func GetKamateTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "stope", Label: "Kamatne stope", HXRequestUrl: kamateURLStope, IsActive: true, Name: "stope"},
			{ID: "formiranje", Label: "Formiranje kamatnih listova", HXRequestUrl: kamateURLFormiranje, IsActive: false, Name: "formiranje"},
			{ID: "obracun", Label: "Obracun kamate", HXRequestUrl: kamateURLObracun, IsActive: false, Name: "obracun"},
		},
	}
}

func setKamateActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
