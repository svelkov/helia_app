package handler

import (
	"net/http"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	sifopContentTitle = "OPSTINE"
	sifopTableID      = "sifop-table"
	sifopURLPrefix    = "/api/sifop/"
)

var sifopTableFields = []domain.Fields{
	{Name: "ops", Label: "Sifra Opstine", Width: "6"},
	{Name: "naziv", Label: "Naziv Opstine", Width: "60"},
}

type SifopHandler struct {
	Service *service.BaseService[domain.Sifop]
}

func NewOpstineHandler(service *service.BaseService[domain.Sifop]) *SifopHandler {
	return &SifopHandler{Service: service}
}
func (h *SifopHandler) CreateSifop(w http.ResponseWriter, r *http.Request) {
	var sifop domain.Sifop
	utils.CreateHelper(w, r, &sifop, h.Service, sifopTableFields)
}

func (h *SifopHandler) UpdateSifop(w http.ResponseWriter, r *http.Request) {
	var sifop domain.Sifop
	utils.UpdateHelper(w, r, &sifop, h.Service, sifopTableFields, utils.IDsifop)
}

func (h *SifopHandler) DeleteSifop(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDsifop)
}

func (h *SifopHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, sifopTableFields)
}

func (h *SifopHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(sifopURLPrefix, "/"), sifopTableFields)
}

func (h *SifopHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, sifopTableFields, utils.IDsifop)
}

func (h *SifopHandler) GetSifop(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, sifopTableFields, utils.IDsifop)
}

func (h *SifopHandler) GetAllSifop(w http.ResponseWriter, r *http.Request) {
	var sifop domain.Sifop
	utils.GetAllEntityHelper(w, r, &sifop, h.Service, sifopTableFields, sifopContentTitle, sifopTableID, sifopURLPrefix, utils.IDsifop)
}

func (h *SifopHandler) AddRoutes(r *http.ServeMux) {
	//define routes for sifop
	r.HandleFunc("POST /api/sifop", infrastructure.AuthMiddleware(h.CreateSifop))
	r.HandleFunc("GET /api/sifop/all", infrastructure.AuthMiddleware(h.GetAllSifop))
	r.HandleFunc("GET /api/sifop/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/sifop/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/sifop/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/sifop/{id}", infrastructure.AuthMiddleware(h.GetSifop))
	r.HandleFunc("PUT /api/sifop/{id}", infrastructure.AuthMiddleware(h.UpdateSifop))
	r.HandleFunc("DELETE /api/sifop/{id}", infrastructure.AuthMiddleware(h.DeleteSifop))
}
