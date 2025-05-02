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
	mestotroskaContentTitle = "MESTO TROSKA"
	mestotroskaTableID      = "mestotroska-table"
	mestotroskaURLPrefix    = "/api/mestotroska/"
)

var mestotroskaTableFields = []domain.Fields{
	{Name: "mtroska", Label: "Mestotroska", Width: "6"},
	{Name: "opis", Label: "Opis", Width: "45"},
	{Name: "idorgjed", Label: "Org. jedinica", Width: "20"},
}

type MestotroskaHandler struct {
	Service *service.BaseService[domain.Mestotr]
}

func NewMestotroskaHandler(service *service.BaseService[domain.Mestotr]) *MestotroskaHandler {
	return &MestotroskaHandler{Service: service}
}

func (h *MestotroskaHandler) CreateMestotroska(w http.ResponseWriter, r *http.Request) {
	var mestotroska domain.Mestotr
	utils.CreateHelper(w, r, &mestotroska, h.Service, mestotroskaTableFields)
}

func (h *MestotroskaHandler) UpdateMestotroska(w http.ResponseWriter, r *http.Request) {
	var mestotroska domain.Mestotr
	utils.UpdateHelper(w, r, &mestotroska, h.Service, mestotroskaTableFields, utils.IDmestotr)
}

func (h *MestotroskaHandler) DeleteMestotroska(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDmestotr)
}

func (h *MestotroskaHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, mestotroskaTableFields)
}

func (h *MestotroskaHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(mestotroskaURLPrefix, "/"), mestotroskaTableFields)
}

func (h *MestotroskaHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, mestotroskaTableFields, utils.IDmestotr)
}

func (h *MestotroskaHandler) GetMestotroska(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, mestotroskaTableFields, utils.IDmestotr)
}

func (h *MestotroskaHandler) GetAllMestotroska(w http.ResponseWriter, r *http.Request) {
	var mestotroska domain.Mestotr
	utils.GetAllEntityHelper(w, r, &mestotroska, h.Service, mestotroskaTableFields, mestotroskaContentTitle, mestotroskaTableID, mestotroskaURLPrefix, utils.IDmestotr)
}

func (h *MestotroskaHandler) AddRoutes(r *http.ServeMux) {
	//define routes for Mestotroska
	r.HandleFunc("POST /api/mestotroska", infrastructure.AuthMiddleware(h.CreateMestotroska))
	r.HandleFunc("GET /api/mestotroska/all", infrastructure.AuthMiddleware(h.GetAllMestotroska))
	r.HandleFunc("GET /api/mestotroska/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/mestotroska/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/mestotroska/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/mestotroska/{id}", infrastructure.AuthMiddleware(h.GetMestotroska))
	r.HandleFunc("PUT /api/mestotroska/{id}", infrastructure.AuthMiddleware(h.UpdateMestotroska))
	r.HandleFunc("DELETE /api/mestotroska/{id}", infrastructure.AuthMiddleware(h.DeleteMestotroska))
}
