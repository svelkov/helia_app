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
	drzavaContentTitle = "DRZAVE"
	drzavaTableID      = "drzave-table"
	drzavaURLPrefix    = "/api/drzava/"
)

var drzavaTableFields = []domain.Fields{
	{Name: "naziv", Label: "Naziv drzave", Width: "40"},
	{Name: "ozndrz", Label: "Oznaka", Width: "6"},
}

type DrzavaHandler struct {
	Service *service.BaseService[domain.Drzave]
}

func NewDrzavaHandler(service *service.BaseService[domain.Drzave]) *DrzavaHandler {
	return &DrzavaHandler{Service: service}
}

func (h *DrzavaHandler) CreateDrzava(w http.ResponseWriter, r *http.Request) {
	var drzava domain.Drzave
	utils.CreateHelper(w, r, &drzava, h.Service, drzavaTableFields)
}

func (h *DrzavaHandler) UpdateDrzava(w http.ResponseWriter, r *http.Request) {
	var drzava domain.Drzave
	utils.UpdateHelper(w, r, &drzava, h.Service, drzavaTableFields, utils.IDdrzave)
}

func (h *DrzavaHandler) DeleteDrzava(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDdrzave)
}

func (h *DrzavaHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, drzavaTableFields)
}

func (h *DrzavaHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(drzavaURLPrefix, "/"), drzavaTableFields)
}

func (h *DrzavaHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, drzavaTableFields, utils.IDdrzave)
}

func (h *DrzavaHandler) GetDrzava(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, drzavaTableFields, utils.IDdrzave)
}

func (h *DrzavaHandler) GetAllDrzave(w http.ResponseWriter, r *http.Request) {
	var Drzava domain.Drzave
	utils.GetAllEntityHelper(w, r, &Drzava, h.Service, drzavaTableFields, drzavaContentTitle, drzavaTableID, drzavaURLPrefix, utils.IDdrzave)
}

func (h *DrzavaHandler) AddRoutes(r *http.ServeMux) {

	r.HandleFunc("POST /api/drzava", infrastructure.AuthMiddleware(h.CreateDrzava))
	r.HandleFunc("GET /api/drzava/all", infrastructure.AuthMiddleware(h.GetAllDrzave))
	r.HandleFunc("GET /api/drzava/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/drzava/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/drzava/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/drzava/{id}", infrastructure.AuthMiddleware(h.GetDrzava))
	r.HandleFunc("PUT /api/drzava/{id}", infrastructure.AuthMiddleware(h.UpdateDrzava))
	r.HandleFunc("DELETE /api/drzava/{id}", infrastructure.AuthMiddleware(h.DeleteDrzava))
}
