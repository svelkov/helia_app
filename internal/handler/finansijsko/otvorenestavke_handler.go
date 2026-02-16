package finansijsko

import (
	"helia/config"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service/finansijsko"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	otvorenestavkeContentTitle string = "OTVORENE STAVKE"
	otvorenestavkeTableID      string = "otvorenestavketable"
)
// OtvoreneStavkeHandler handles all requests for Otvorene Stavke module
type OtvoreneStavkeHandler struct {
	tabData domain.TabData
	service finansijsko.OtvoreneStavkeService
	cfg     config.Config
}

// NewOtvoreneStavkeHandler creates a new handler instance
func NewOtvoreneStavkeHandler(
	service finansijsko.OtvoreneStavkeService,
	cfg config.Config,
) *OtvoreneStavkeHandler {
	handler := &OtvoreneStavkeHandler{
		service: service,
		cfg:     cfg,
	}
	handler.tabData = GetOtvoreneStavkeTabData()
	return handler
}

// OtvoreneStavkeMain - Main entry point for Otvorene Stavke
func (h *OtvoreneStavkeHandler) OtvoreneStavkeMain(c *gin.Context) {
	session := domain.GetSessionFromContext(c)
	gnGod := 0
	if session != nil {
		gnGod = session.SelectedGod
	}

	tbl := common.SetTableBasicData(otvorenestavkeContentTitle, otvorenestavkeTableID, h.service.GetOtvoreneStavkeFields(), "", "", 0, 0, 0, 0, h.cfg)
	common.SetTableConfig(&tbl, otvorenestavkeContentTitle, "", false, false, false)

	err := tmpl_fin.OtvoreneStavkeMain(h.tabData, tbl, i18n.GetInstance(), gnGod).Render(c.Request.Context(), c.Writer)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
		return
	}
}

// OtvoreneStavke - Tab 1: Otvorene stavke (Open Items)
func (h *OtvoreneStavkeHandler) OtvoreneStavke(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("otvorenestavke")
		//renderOtvoreneStavkeContent(c, h, 0)
		return
	}

	if source == "btnobrada" || source == "search" {
		// params := map[string]interface{}{
		// 	"konto":           c.PostForm("konto"),
		// 	"pocetakkonta":    c.PostForm("pocetakkonta"),
		// 	"zavrsenisifra":   c.PostForm("zavrsenisifra"),
		// 	"poddatumom":      c.PostForm("poddatumom"),
		// 	"otv_stavke_dana": c.PostForm("otv_stavke_dana"),
		// }

		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetOtvoreneStavke(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("otvorenestavke")
		//renderTableContent(c, h, 0, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("otvorenestavke")
	//renderOtvoreneStavkeContent(c, h, 0)
}

// ZatvoreneStavke - Tab 2: Zatvorene stavke (Closed Items)
func (h *OtvoreneStavkeHandler) ZatvoreneStavke(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("zatvorenestavke")
		//renderOtvoreneStavkeContent(c, h, 1)
		return
	}

	if source == "btnobrada" || source == "search" {

		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetZatvoreneStavke(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("zatvorenestavke")
		//renderTableContent(c, h, 1, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("zatvorenestavke")
	//renderOtvoreneStavkeContent(c, h, 1)
}

// IOS - Tab 3: IOS (Izvod otvorenih stavki)
func (h *OtvoreneStavkeHandler) IOS(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("ios")
		//renderOtvoreneStavkeContent(c, h, 2)
		return
	}

	if source == "btnobrada" || source == "search" {

		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetIOS(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("ios")
		//renderTableContent(c, h, 2, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("ios")
	//renderOtvoreneStavkeContent(c, h, 2)
}

// Opomene - Tab 4: Opomene (Reminders)
func (h *OtvoreneStavkeHandler) Opomene(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("opomene")
		//renderOtvoreneStavkeContent(c, h, 3)
		return
	}

	if source == "btnobrada" || source == "search" {
		// params := map[string]interface{}{
		// 	"konto":           c.PostForm("konto4"),
		// 	"pocetakkonta":    c.PostForm("pocetakkonta4"),
		// 	"zavrsenisifra":   c.PostForm("zavrsenisifra4"),
		// 	"poddatumom":      c.PostForm("poddatumom4"),
		// 	"otv_stavke_dana": c.PostForm("otv_stavke_dana4"),
		// }

		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetOpomene(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("opomene")
		//renderTableContent(c, h, 3, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("opomene")
	//renderOtvoreneStavkeContent(c, h, 3)
}

// DospelaPotraživanja - Tab 5: Dospela potraživanja/dugovanja
func (h *OtvoreneStavkeHandler) DospelaPotraživanja(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("dospelapotraživanja")
		//renderOtvoreneStavkeContent(c, h, 4)
		return
	}

	if source == "btnobrada" || source == "search" {

		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetDospelaPotraživanja(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("dospelapotraživanja")
		//renderTableContent(c, h, 4, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("dospelapotraživanja")
	//renderOtvoreneStavkeContent(c, h, 4)
}

// PregledDugovanjaPredStarosti - Tab 6: Pregled dugovanja po starosti
func (h *OtvoreneStavkeHandler) PregledDugovanjaPredStarosti(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("pregleddugovanja")
		//renderOtvoreneStavkeContent(c, h, 5)
		return
	}

	if source == "btnobrada" || source == "search" {

		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPregledDugovanjaPredStarosti(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("pregleddugovanja")
		//renderTableContent(c, h, 5, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("pregleddugovanja")
	//renderOtvoreneStavkeContent(c, h, 5)
}

// PregledDospelogDuga - Tab 7: Pregled dospelog duga
func (h *OtvoreneStavkeHandler) PregledDospelogDuga(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("pregleddugatospelogduga")
		//renderOtvoreneStavkeContent(c, h, 6)
		return
	}

	if source == "btnobrada" || source == "search" {
		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPregledDospelogDuga(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("pregleddugatospelogduga")
		//renderTableContent(c, h, 6, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("pregleddugatospelogduga")
	//renderOtvoreneStavkeContent(c, h, 6)
}

// PovezivanjRacunaIUplata - Tab 8: Povezivanje računa i uplata
func (h *OtvoreneStavkeHandler) PovezivanjRacunaIUplata(c *gin.Context) {
	source := c.Query("source")
	tbl := &domain.TableData{}

	if source == "menu" || source == "tab" {
		h.setOtvoreneStavkeActiveTab("povezivanje")
		//renderOtvoreneStavkeContent(c, h, 7)
		return
	}

	if source == "btnobrada" || source == "search" {

		pageSize, currentPage := common.GetPageAndPageSizeFromRequest(c, h.cfg)
		if err := h.service.GetPovezivanjRacunaIUplata(c, tbl, true, pageSize, currentPage); err != nil {
			common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, common.ErrMsgRenderTemplate)
			return
		}

		h.setOtvoreneStavkeActiveTab("povezivanje")
		//renderTableContent(c, h, 7, tbl)
		return
	}

	h.setOtvoreneStavkeActiveTab("povezivanje")
	//renderOtvoreneStavkeContent(c, h, 7)
}

// RegisterRoutes registers the routes for the Otvorene Stavke handler
func (h *OtvoreneStavkeHandler) RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())

	r.GET("api/otvorenestavke", h.OtvoreneStavkeMain)
	r.POST("api/otvorenestavke", h.OtvoreneStavke)
	r.GET("api/otvorenestavke/zatvorenestavke", h.ZatvoreneStavke)
	r.GET("api/otvorenestavke/ios", h.IOS)
	r.POST("api/otvorenestavke/ios", h.IOS)
	r.GET("api/otvorenestavke/opomene", h.Opomene)
	r.POST("api/otvorenestavke/opomene", h.Opomene)
	r.GET("api/otvorenestavke/dospelapotraživanja", h.DospelaPotraživanja)
	r.POST("api/otvorenestavke/dospelapotraživanja", h.DospelaPotraživanja)
	r.GET("api/otvorenestavke/pregleddugovanja", h.PregledDugovanjaPredStarosti)
	r.POST("api/otvorenestavke/pregleddugovanja", h.PregledDugovanjaPredStarosti)
	r.GET("api/otvorenestavke/pregleddugatospelogduga", h.PregledDospelogDuga)
	r.POST("api/otvorenestavke/pregleddugatospelogduga", h.PregledDospelogDuga)
	r.GET("api/otvorenestavke/povezivanje", h.PovezivanjRacunaIUplata)
	r.POST("api/otvorenestavke/povezivanje", h.PovezivanjRacunaIUplata)
}

// Helper functions

func GetOtvoreneStavkeTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "otvorenestavke", Label: "Otvorene stavke", HXRequestUrl: "/api/otvorenestavke/otvorenestavke", IsActive: true, Name: "otvorenestavke"},
			{ID: "zatvorenestavke", Label: "Zatvorene stavke", HXRequestUrl: "/api/otvorenestavke/zatvorenestavke", IsActive: false, Name: "zatvorenestavke"},
			{ID: "ios", Label: "IOS", HXRequestUrl: "/api/otvorenestavke/ios", IsActive: false, Name: "ios"},
			{ID: "opomene", Label: "Opomene", HXRequestUrl: "/api/otvorenestavke/opomene", IsActive: false, Name: "opomene"},
			{ID: "dospelapotraživanja", Label: "Dospela potraživanja/dugovanja", HXRequestUrl: "/api/otvorenestavke/dospelapotraživanja", IsActive: false, Name: "dospelapotraživanja"},
			{ID: "pregleddugovanja", Label: "Pregled Dugovanja/Obaveze", HXRequestUrl: "/api/otvorenestavke/pregleddugovanja", IsActive: false, Name: "pregleddugovanja"},
			{ID: "pregleddugatospelogduga", Label: "Pregled dospelog duga", HXRequestUrl: "/api/otvorenestavke/pregleddugatospelogduga", IsActive: false, Name: "pregleddugatospelogduga"},
			{ID: "povezivanje", Label: "Povezivanje računa i uplata", HXRequestUrl: "/api/otvorenestavke/povezivanje", IsActive: false, Name: "povezivanje"},
		},
	}
}

func (h *OtvoreneStavkeHandler) setOtvoreneStavkeActiveTab(tab string) {
	for i, t := range h.tabData.Tabs {
		if t.Name == tab {
			h.tabData.Tabs[i].IsActive = true
		} else {
			h.tabData.Tabs[i].IsActive = false
		}
	}
}
