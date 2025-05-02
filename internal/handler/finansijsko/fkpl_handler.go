package finansijsko

import (
	"net/http"
	"strings"

	"helia/internal/domain"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	fkplContentTitle = "KONTNI PLAN"
	fkplTableID      = "fkpl-table"
	fkplURLPrefix    = "/api/fkpl/"
)

var fkplTableFields = []domain.Fields{
	{Name: "konto", Label: "Konto", Width: "10"},
	{Name: "naziv", Label: "Naziv", Width: "60"},
	{Name: "sifra", Label: "Sifra", Width: "10"},
	{Name: "vkonta", Label: "Vrsta konta", Width: "4"},
}

type FkplHandler struct {
	Service *service.BaseService[domain.Fkpl]
}

func NewFkplHandler(service *service.BaseService[domain.Fkpl]) *FkplHandler {
	return &FkplHandler{Service: service}
}

func (h *FkplHandler) CreateFkpl(w http.ResponseWriter, r *http.Request) {
	var fkpl domain.Fkpl
	utils.CreateHelper(w, r, &fkpl, h.Service, fkplTableFields)
}

func (h *FkplHandler) UpdateFkpl(w http.ResponseWriter, r *http.Request) {
	var fkpl domain.Fkpl
	utils.UpdateHelper(w, r, &fkpl, h.Service, fkplTableFields, utils.IDfkpl)
}

func (h *FkplHandler) DeleteFkpl(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDfkpl)
}

func (h *FkplHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, fkplTableFields)
}

func (h *FkplHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(fkplURLPrefix, "/"), fkplTableFields)
}

func (h *FkplHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, fkplTableFields, utils.IDfkpl)
}

func (h *FkplHandler) GetFkpl(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, fkplTableFields, utils.IDfkpl)
}

func (h *FkplHandler) GetAllFkpl(w http.ResponseWriter, r *http.Request) {
	var fkpl domain.Fkpl
	utils.GetAllEntityHelper(w, r, &fkpl, h.Service, fkplTableFields, fkplContentTitle, fkplTableID, fkplURLPrefix, utils.IDfkpl)
}

func (h *FkplHandler) AddRoutes(r *http.ServeMux) {
	// Define routes for fkpl
	r.HandleFunc("POST /api/fkpl", h.CreateFkpl)
	r.HandleFunc("GET /api/fkpl/all", h.GetAllFkpl)
	r.HandleFunc("GET /api/fkpl/confirm-delete", h.confirmDeleteHandler)
	r.HandleFunc("GET /api/fkpl/confirm-update", h.confirmUpdateHandler)
	r.HandleFunc("GET /api/fkpl/confirm-add", h.confirmAddHandler)
	r.HandleFunc("GET /api/fkpl/{id}", h.GetFkpl)
	r.HandleFunc("PUT /api/fkpl/{id}", h.UpdateFkpl)
	r.HandleFunc("DELETE /api/fkpl/{id}", h.DeleteFkpl)
}
