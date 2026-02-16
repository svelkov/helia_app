package finansijsko

import (
	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/service"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
)

var sfTableFields = []domain.Fields{
	{Name: "brst", Label: "Uk. Br. Stavki", Width: "10"},
	{Name: "brna", Label: "Uk. Br.Baoga", Width: "10"},
	{Name: "dug", Label: "Duguje", Width: "10"},
	{Name: "Pot", Label: "Potrazuje", Width: "10"},
}

type SfHandler struct {
	Service *service.BaseService[domain.Sf]
	cfg     config.Config
}

func NewSfHandler(service *service.BaseService[domain.Sf], cfg config.Config) *SfHandler {
	return &SfHandler{Service: service, cfg: cfg}
}

func (h *SfHandler) UpdateSf(c *gin.Context) {
	var sf domain.Sf
	utils.UpdateHelper(c, &sf, h.Service, sfTableFields, common.IDfkpl)
}

func (h *SfHandler) GetAllSf(c *gin.Context) {
	utils.GetAllEntityHelper(c, h.Service, sfTableFields, "", "", "", "", "", h.cfg)
}
