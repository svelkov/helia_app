package robno

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"helia/config"
	tmpl_rep_robno "helia/frontend/templates/reports/robno"
	tmpl_robno "helia/frontend/templates/robno"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	robnosvc "helia/internal/service/robno"
	"helia/pkg/utils"
)

const krajPoslovneGodineURL = "/api/robno/krajposlovnegodine"

type KrajPoslovneGodineHandler struct {
	service robnosvc.KrajPoslovneGodineService
	cfg     config.Config
	tabs    *domain.TabData
	subtabs *domain.TabData
}

func NewKrajPoslovneGodineHandler(service robnosvc.KrajPoslovneGodineService, cfg config.Config) *KrajPoslovneGodineHandler {
	tr := i18n.GetInstance()
	return &KrajPoslovneGodineHandler{service: service, cfg: cfg,
		tabs:    &domain.TabData{Tabs: []domain.TabItem{{ID: "kpg-popis", Label: tr.Title("Popisne liste"), HXRequestUrl: krajPoslovneGodineURL + "/popis", IsActive: true}, {ID: "kpg-visak", Label: tr.Title("Obrada viškova/manjkova"), HXRequestUrl: krajPoslovneGodineURL + "/obrada"}, {ID: "kpg-prepis", Label: tr.Title("Prepis stanja"), HXRequestUrl: krajPoslovneGodineURL + "/prepis"}}},
		subtabs: &domain.TabData{Tabs: []domain.TabItem{{ID: "kpg-sifra", Label: tr.Title("Po šifri"), HXRequestUrl: krajPoslovneGodineURL + "/popis?tip=sifra", IsActive: true}, {ID: "kpg-naziv", Label: tr.Title("Po nazivu"), HXRequestUrl: krajPoslovneGodineURL + "/popis?tip=naziv"}, {ID: "kpg-grupa", Label: tr.Title("Po grupi"), HXRequestUrl: krajPoslovneGodineURL + "/popis?tip=grupa"}, {ID: "kpg-grupa-naziv", Label: tr.Title("Po grupi i nazivu"), HXRequestUrl: krajPoslovneGodineURL + "/popis?tip=grupa-naziv"}}},
	}
}

func (h *KrajPoslovneGodineHandler) Main(c *gin.Context) {
	common.SetActiveTab(h.tabs, 0)
	common.SetActiveTab(h.subtabs, 0)
	mag, err := h.service.GetMagacini(c.Request.Context())
	if err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	tbl := h.table("Popisne liste", h.service.GetType1Fields())
	if err := tmpl_robno.KrajPoslovneGodineMain(*h.tabs, *h.subtabs, tbl, mag, i18n.GetInstance()).Render(c, c.Writer); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *KrajPoslovneGodineHandler) Popis(c *gin.Context) { h.renderRows(c, 0, false) }
func (h *KrajPoslovneGodineHandler) Obrada(c *gin.Context) {
	common.SetActiveTab(h.tabs, 1)
	var rows []domain.KrajVisakManjakRow
	if c.Request.Header.Get("X-Request-Source") == "btnobrada" {
		var err error
		p := h.params(c)
		if errors := h.service.ValidateObrada(p); len(errors) > 0 {
			common.WriteJSONResponse(c, 400, false, errors, common.ErrMsgValidation)
			return
		}
		rows, err = h.service.GetVisakManjak(c.Request.Context(), p)
		if err != nil {
			common.WriteJSONResponse(c, 500, false, nil, err.Error())
			return
		}
	}
	mag, _ := h.service.GetMagacini(c.Request.Context())
	if err := tmpl_robno.KrajPoslovneGodineObrada(*h.tabs, h.tableVisakRows("Obrada viškova i manjkova", rows), mag, i18n.GetInstance()).Render(c, c.Writer); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, common.ErrMsgRenderTemplate)
	}
}
func (h *KrajPoslovneGodineHandler) Prepis(c *gin.Context) {
	common.SetActiveTab(h.tabs, 2)
	if c.Request.Header.Get("X-Request-Source") == "btnprepis" {
		year, _ := strconv.Atoi(c.Query("novagod"))
		p := h.params(c)
		p.NovaGod = year
		if errors := h.service.ValidateRSTA(p); len(errors) > 0 {
			common.WriteJSONResponse(c, 400, false, errors, common.ErrMsgValidation)
			return
		}
		if err := h.service.PrepisStanja(c.Request.Context(), p, year); err != nil {
			common.WriteJSONResponse(c, 500, false, nil, err.Error())
			return
		}
	}
	mag, _ := h.service.GetMagacini(c.Request.Context())
	if err := tmpl_robno.KrajPoslovneGodinePrepis(*h.tabs, h.table("Prepis stanja", h.service.GetType1Fields()), mag, i18n.GetInstance()).Render(c, c.Writer); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *KrajPoslovneGodineHandler) renderRows(c *gin.Context, tab int, print bool) {
	common.SetActiveTab(h.tabs, tab)
	if tab == 0 {
		common.SetActiveTab(h.subtabs, h.subtabIndex(c.Query("tip")))
	}
	data, err := h.service.GetPopis(c.Request.Context(), h.params(c))
	if err != nil {
		common.WriteJSONResponse(c, 500, false, nil, err.Error())
		return
	}
	tbl := h.tablePopisRows("Popisne liste", data)
	if print {
		h.renderReport(c, tbl)
		return
	}
	if c.Request.Header.Get("X-Request-Source") == "btnobrada" || c.Request.Header.Get("X-Request-Source") == "btnpage" {
		utils.RenderContent(c, tbl)
		return
	}
	mag, _ := h.service.GetMagacini(c.Request.Context())
	if err := tmpl_robno.KrajPoslovneGodinePopis(*h.tabs, *h.subtabs, tbl, mag, i18n.GetInstance()).Render(c, c.Writer); err != nil {
		common.WriteJSONResponse(c, 500, false, nil, common.ErrMsgRenderTemplate)
	}
}

func (h *KrajPoslovneGodineHandler) params(c *gin.Context) domain.KrajPoslovneGodineParams {
	p := domain.KrajPoslovneGodineParams{Magacin: common.StringToInt(c.Query("magacin")), OdKonta: c.Query("odkonta"), DoKonta: c.Query("dokonta"), OdSifre: common.StringToInt(c.Query("odsifre")), DoSifre: common.StringToInt(c.Query("dosifre")), Cena: common.StringToInt(c.Query("cena")), Tip: c.Query("tip"), TipZal: common.StringToInt(c.Query("tipzal")), Nalog: common.StringToInt(c.Query("nalog")), Dokum: common.StringToInt(c.Query("dokum")), Vrd: common.StringToInt(c.Query("vrd")), NovaGod: common.StringToInt(c.Query("novagod")), StaraGod: common.StringToInt(c.Query("staragod")), OrgJed: common.StringToInt(c.Query("orgjed")), MestoTroska: common.StringToInt(c.Query("mestotroska")), DanObrade: c.Query("danobrade"), Opis: c.Query("opis"), DanNalog: c.Query("dannalog"), DanDokum: c.Query("dandokum")}
	p.ObradiNule = c.Query("obradinule") == "on" || c.Query("obradinule") == "true"
	return p
}
func (h *KrajPoslovneGodineHandler) subtabIndex(tip string) int {
	switch tip {
	case "naziv":
		return 1
	case "grupa":
		return 2
	case "grupa-naziv":
		return 3
	default:
		return 0
	}
}
func (h *KrajPoslovneGodineHandler) tablePopisRows(title string, data domain.KrajPopisData) domain.TableData {
	if data.TipZal == 2 {
		t := h.table(title, h.service.GetType2Fields())
		for _, row := range data.Type2 {
			t.Rows = append(t.Rows, domain.TableRow{Fields: []string{fmt.Sprint(row.RedBr), row.Konto, fmt.Sprint(row.Sifra), row.Naziv, row.JM, fmt.Sprint(row.Grupa), row.NazivGrupe, row.OTK, row.Serija, row.Rok, fmt.Sprintf("%.3f", row.Kolicina1), fmt.Sprintf("%.3f", row.Kolicina2), fmt.Sprintf("%.3f", row.Kolicina3), fmt.Sprintf("%.3f", row.UkupnaKolicina), fmt.Sprintf("%.3f", row.StanjeZaliha)}})
		}
		return t
	}
	t := h.table(title, h.service.GetType1Fields())
	for _, row := range data.Type1 {
		t.Rows = append(t.Rows, domain.TableRow{Fields: []string{fmt.Sprint(row.RedBr), row.Konto, fmt.Sprint(row.Sifra), row.Naziv, row.JM, fmt.Sprint(row.Grupa), row.NazivGrupe, fmt.Sprintf("%.2f", row.Cena), fmt.Sprintf("%.3f", row.Kolicina1), fmt.Sprintf("%.3f", row.Kolicina2), fmt.Sprintf("%.3f", row.Kolicina3), fmt.Sprintf("%.3f", row.UkupnaKolicina), fmt.Sprintf("%.3f", row.StanjeZaliha)}})
	}
	return t
}
func (h *KrajPoslovneGodineHandler) tableVisakRows(title string, rows []domain.KrajVisakManjakRow) domain.TableData {
	t := h.table(title, h.service.GetVisakManjakFields())
	for _, row := range rows {
		t.Rows = append(t.Rows, domain.TableRow{Fields: []string{fmt.Sprint(row.RedBr), fmt.Sprint(row.Sifra), row.Konto, row.Naziv, row.JM, fmt.Sprintf("%.2f", row.Cena), fmt.Sprintf("%.3f", row.KolicinaPopisa), fmt.Sprintf("%.2f", row.IznosPopisa), fmt.Sprintf("%.3f", row.StanjeKnjigovodstveno), fmt.Sprintf("%.2f", row.IznosKnjigovodstveno), fmt.Sprintf("%.3f", row.Visak), fmt.Sprintf("%.2f", row.IznosViska), fmt.Sprintf("%.3f", row.Manjak), fmt.Sprintf("%.2f", row.IznosManjka), fmt.Sprintf("%.2f", row.FinansijskiVisak), fmt.Sprintf("%.2f", row.FinansijskiManjak)}})
	}
	return t
}
func (h *KrajPoslovneGodineHandler) renderReport(c *gin.Context, tbl domain.TableData) {
	params := h.params(c)
	var err error
	if c.Request.URL.Path == "/api/robno/krajposlovnegodine/popis/stampa2" {
		err = tmpl_rep_robno.ROB_RPT_POPISNALISTA2(params, tbl, i18n.GetInstance()).Render(c, c.Writer)
	} else {
		err = tmpl_rep_robno.ROB_RPT_POPISNALISTA(params, tbl, i18n.GetInstance()).Render(c, c.Writer)
	}
	if err != nil {
		common.WriteJSONResponse(c, 500, false, nil, common.ErrMsgRenderTemplate)
	}
}
func (h *KrajPoslovneGodineHandler) table(title string, fields []domain.Fields) domain.TableData {
	t := common.SetTableBasicData(title, "kpg-table", fields, "", "", 0, 0, 0, 0, h.cfg)
	t.SearchEnabled = false
	t.ShowPagination = false
	return t
}
func (h *KrajPoslovneGodineHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth())
	r.GET("/api/robno/krajposlovnegodine", h.Main)
	r.GET("/api/robno/krajposlovnegodine/popis", h.Popis)
	r.GET("/api/robno/krajposlovnegodine/obrada", h.Obrada)
	r.GET("/api/robno/krajposlovnegodine/prepis", h.Prepis)
	r.GET("/api/robno/krajposlovnegodine/popis/stampa", func(c *gin.Context) { h.renderRows(c, 0, true) })
	r.GET("/api/robno/krajposlovnegodine/popis/stampa2", func(c *gin.Context) { h.renderRows(c, 0, true) })
}
