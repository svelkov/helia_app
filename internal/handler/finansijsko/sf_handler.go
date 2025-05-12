package finansijsko

import (
	"net/http"

	"helia/internal/domain"
	"helia/internal/service"
	"helia/pkg/utils"
)

var sfTableFields = []domain.Fields{
	{Name: "brst", Label: "Uk. Br. Stavki", Width: "10"},
	{Name: "brna", Label: "Uk. Br.Baoga", Width: "10"},
	{Name: "dug", Label: "Duguje", Width: "10"},
	{Name: "Pot", Label: "Potrazuje", Width: "10"},
}

type SfHandler struct {
	Service *service.BaseService[domain.Sf]
}

func NewSfHandler(service *service.BaseService[domain.Sf]) *SfHandler {
	return &SfHandler{Service: service}
}

func (h *SfHandler) UpdateSf(w http.ResponseWriter, r *http.Request) {
	var sf domain.Sf
	utils.UpdateHelper(w, r, &sf, h.Service, sfTableFields, utils.IDfkpl)
}

func (h *SfHandler) GetAllSf(w http.ResponseWriter, r *http.Request) {
	var sf domain.Sf
	utils.GetAllEntityHelper(w, r, &sf, h.Service, sfTableFields, "", "", "", "")
}
