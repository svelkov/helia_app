package finansijsko

import (
	"net/http"
	"strings"

	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	"helia/pkg/utils"

	finservice "helia/internal/service/finansijsko"

	"github.com/gin-gonic/gin"
)

const (
	fkplContentTitle   string = "KONTNI PLAN"
	fkplTableID        string = "fkpl-table"
	searchKontoTableID string = "searchkonto-table"
	fkplURLPrefix      string = "/api/fkpl/"
	fkplURLGetAll      string = "/api/fkpl/all"
)

func SetFkplFields() []domain.Fields {
	return []domain.Fields{
		{Name: "konto", Label: "Konto", Width: "10"},
		{Name: "sifra", Label: "Sifra", Width: "10"},
		{Name: "naziv", Label: "Naziv", Width: "120"},
		{Name: "vkonta", Label: "Vrsta konta", Width: "4"},
	}
}

var fkplSearchTableFields = []domain.Fields{
	{Name: "konto", Label: "Konto", Width: "10"},
	{Name: "sifra", Label: "Sifra", Width: "10"},
	{Name: "naziv", Label: "Naziv", Width: "120"},
}

type FkplHandler struct {
	service     service.Service[domain.Fkpl]
	fkplService finservice.FkplService
	cfg         config.Config
}

func NewFkplHandler(service *service.BaseService[domain.Fkpl], fkplService finservice.FkplService, cfg config.Config) *FkplHandler {
	return &FkplHandler{service: service, fkplService: fkplService, cfg: cfg}
}

func (h *FkplHandler) CreateFkpl(c *gin.Context) {
	var fkpl domain.Fkpl
	utils.CreateHelper(c, &fkpl, h.service, common.IDfkpl, SetFkplFields())
}

func (h *FkplHandler) UpdateFkpl(c *gin.Context) {
	var fkpl domain.Fkpl
	utils.UpdateHelper(c, &fkpl, h.service, SetFkplFields(), common.IDfkpl)
}

func (h *FkplHandler) DeleteFkpl(c *gin.Context) {
	utils.DeleteHelper(c, h.service, common.IDfkpl)
}

func (h *FkplHandler) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, SetFkplFields())
}

func (h *FkplHandler) confirmAddHandler(c *gin.Context) {
	utils.ConfirmAddHelper(c, strings.TrimSuffix(fkplURLPrefix, "/"), SetFkplFields())
}

func (h *FkplHandler) confirmUpdateHandler(c *gin.Context) {
	utils.ConfirmUpdateHelper(c, h.service, SetFkplFields(), common.IDfkpl)
}

func (h *FkplHandler) GetFkpl(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, SetFkplFields(), common.IDfkpl)
}

func (h *FkplHandler) GetAllFkpl(c *gin.Context) {

	tbl := utils.GetAllEntityHelper(c, h.service, SetFkplFields(), fkplContentTitle, fkplTableID, fkplURLPrefix, fkplURLGetAll, common.IDfkpl, h.cfg)
	utils.RenderContent(c, *tbl)
}

func (h *FkplHandler) TraziKonto(c *gin.Context) {
	entities, err := h.fkplService.TraziKonto(c)
	if err != nil {
		c.Writer.Header().Set("Content-Type", "text/plain")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	// Check if pointer is not nil and slice is not empty
	if entities != nil && len(*entities) > 0 {
		firstElement := (*entities)[0].Naziv
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.Write([]byte(firstElement))
		return
	}

}
func (h *FkplHandler) TraziKontoSearchTable(c *gin.Context) {
	// Parse query parameters
	tbl := common.SetTableBasicData("", searchKontoTableID, fkplSearchTableFields, "", "", 0, 0, 0, 0, h.cfg)
	err := h.fkplService.KontoSearchForTable(c, &tbl)
	if err != nil {
		c.Writer.Header().Set("Content-Type", "text/plain")
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, []domain.FieldError{}, err.Error())
		return
	}
	utils.RenderContent(c, tbl)
}
func (h *FkplHandler) AddRoutes(r *gin.Engine) {
	// Create API group with prefix
	//api := r.Group(fkplURLPrefix)
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Define routes for fkpl
	r.POST("/api/fkpl", h.CreateFkpl)
	r.GET("/api/fkpl/all", h.GetAllFkpl)
	r.GET("/api/fkpl/confirm-delete", h.confirmDeleteHandler)
	r.GET("/api/fkpl/confirm-update", h.confirmUpdateHandler)
	r.GET("/api/fkpl/confirm-add", h.confirmAddHandler)
	r.GET("/api/fkpl/:id", h.GetFkpl)
	r.PUT("/api/fkpl/:id", h.UpdateFkpl)
	r.DELETE("/api/fkpl/:id", h.DeleteFkpl)
	r.GET("/api/fkpl/trazikonto", h.TraziKonto)
	r.GET("/api/fkpl/trazikontosearchtable", h.TraziKontoSearchTable)
}
