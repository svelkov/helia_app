package handler

import (
	"net/http"
	"strings"

	"helia/internal/domain"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	sifmestoContentTitle = "MESTA"
	sifmestoTableID      = "sifmesto-table"
	sifmestoURLPrefix    = "/api/sifmesto/"
)

var sifmestoTableFields = []domain.Fields{
	{Name: "sifm", Label: "Sifra Mesta", Width: "4"},
	{Name: "naziv", Label: "Naziv", Width: "80"},
	{Name: "ops", Label: "Opstina", Width: "8"},
	{Name: "pobro", Label: "Postanski broj", Width: "8"},
	{Name: "km", Label: "Km", Width: "6"},
}

type SifmestoHandler struct {
	Service *service.BaseService[domain.Sifmesto]
}

func NewSifmestoHandler(service *service.BaseService[domain.Sifmesto]) *SifmestoHandler {
	return &SifmestoHandler{Service: service}
}
func (h *SifmestoHandler) CreateSifmesto(w http.ResponseWriter, r *http.Request) {
	var sifmesto domain.Sifmesto
	utils.CreateHelper(w, r, &sifmesto, h.Service, sifmestoTableFields)
}

func (h *SifmestoHandler) UpdateSifmesto(w http.ResponseWriter, r *http.Request) {
	var sifmesto domain.Sifmesto
	utils.UpdateHelper(w, r, &sifmesto, h.Service, sifmestoTableFields, utils.IDsifmesto)
}

func (h *SifmestoHandler) DeleteSifmesto(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDsifmesto)
}

func (h *SifmestoHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, sifmestoTableFields)
}

func (h *SifmestoHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(sifmestoURLPrefix, "/"), sifmestoTableFields)
}

func (h *SifmestoHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, sifmestoTableFields, utils.IDsifmesto)
}

func (h *SifmestoHandler) GetSifmesto(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, sifmestoTableFields, utils.IDsifmesto)
}

func (h *SifmestoHandler) GetAllSifmesto(w http.ResponseWriter, r *http.Request) {
	var sifmesto domain.Sifmesto
	utils.GetAllEntityHelper(w, r, &sifmesto, h.Service, sifmestoTableFields, sifmestoContentTitle, sifmestoTableID, sifmestoURLPrefix, utils.IDsifmesto)
}

func (h *SifmestoHandler) AddRoutes(r *http.ServeMux) {
	//define routes for sifmesto
	r.HandleFunc("POST /api/sifmesto", h.CreateSifmesto)
	r.HandleFunc("GET /api/sifmesto/all", h.GetAllSifmesto)
	r.HandleFunc("GET /api/sifmesto/confirm-delete", h.confirmDeleteHandler)
	r.HandleFunc("GET /api/sifmesto/confirm-update", h.confirmUpdateHandler)
	r.HandleFunc("GET /api/sifmesto/confirm-add", h.confirmAddHandler)
	r.HandleFunc("GET /api/sifmesto/{id}", h.GetSifmesto)
	r.HandleFunc("PUT /api/sifmesto/{id}", h.UpdateSifmesto)
	r.HandleFunc("DELETE /api/sifmesto/{id}", h.DeleteSifmesto)
}
