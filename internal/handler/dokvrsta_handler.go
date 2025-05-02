package handler

import (
	"net/http"
	"strings"

	"helia/internal/domain"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	dokvrstaContentTitle = "VRSTE DOKUMENATA"
	dokvrstaTableID      = "dokvrsta-table"
	dokvrstaURLPrefix    = "/api/dokvrsta/"
)

var dokvrstaTableFields = []domain.Fields{
	{Name: "vrd", Label: "Vrsta dokumenta", Type: "text", Width: "5"},
	{Name: "opis", Label: "Opis vrste dokumenta", Type: "text", Width: "40"},
	{Name: "kodknj", Label: "Nacin knjizenja (d,p,dp)", Type: "text", Width: "4"},
	{Name: "predznak", Label: "Predznak", Type: "text", Width: "3"},
	{Name: "grpdok", Label: "Grupa dokumenta", Type: "text", Width: "10"},
	{Name: "dokozn", Label: "Oznaka dokumenta", Type: "text", Width: "5"},
	{Name: "stornovrd", Label: "Storno vrsta dokumenta", Type: "text", Width: "6"},
	{Name: "modul", Label: "Modul", Type: "text", Width: "8"},
}

type DokvrstaHandler struct {
	Service *service.BaseService[domain.Dokvrsta]
}

func NewDokvrstaHandler(service *service.BaseService[domain.Dokvrsta]) *DokvrstaHandler {
	return &DokvrstaHandler{Service: service}
}

func (h *DokvrstaHandler) CreateDokvrsta(w http.ResponseWriter, r *http.Request) {
	var dokvrsta domain.Dokvrsta
	utils.CreateHelper(w, r, &dokvrsta, h.Service, dokvrstaTableFields)
}

func (h *DokvrstaHandler) UpdateDokvrsta(w http.ResponseWriter, r *http.Request) {
	var dokvrsta domain.Dokvrsta
	utils.UpdateHelper(w, r, &dokvrsta, h.Service, dokvrstaTableFields, utils.IDdokvrsta)
}

func (h *DokvrstaHandler) DeleteDokvrsta(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDdokvrsta)
}

func (h *DokvrstaHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, dokvrstaTableFields)
}

func (h *DokvrstaHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(dokvrstaURLPrefix, "/"), dokvrstaTableFields)
}

func (h *DokvrstaHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, dokvrstaTableFields, utils.IDdokvrsta)
}

func (h *DokvrstaHandler) GetDokvrsta(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, dokvrstaTableFields, utils.IDdokvrsta)
}

func (h *DokvrstaHandler) GetAllDokvrsta(w http.ResponseWriter, r *http.Request) {
	var dokvrsta domain.Dokvrsta
	utils.GetAllEntityHelper(w, r, &dokvrsta, h.Service, dokvrstaTableFields, dokvrstaContentTitle, dokvrstaTableID, dokvrstaURLPrefix, utils.IDdokvrsta)
}

func (h *DokvrstaHandler) AddRoutes(r *http.ServeMux) {
	//define routes for Dokvrsta
	r.HandleFunc("POST /api/dokvrsta", h.CreateDokvrsta)
	r.HandleFunc("GET /api/dokvrsta/all", h.GetAllDokvrsta)
	r.HandleFunc("GET /api/dokvrsta/confirm-delete", h.confirmDeleteHandler)
	r.HandleFunc("GET /api/dokvrsta/confirm-update", h.confirmUpdateHandler)
	r.HandleFunc("GET /api/dokvrsta/confirm-add", h.confirmAddHandler)
	r.HandleFunc("GET /api/dokvrsta/{id}", h.GetDokvrsta)
	r.HandleFunc("PUT /api/dokvrsta/{id}", h.UpdateDokvrsta)
	r.HandleFunc("DELETE /api/dokvrsta/{id}", h.DeleteDokvrsta)
}
