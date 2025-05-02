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
	fvepdvContentTitle = "VRSTE EVIDENCIJE PDV"
	fvepdvTableID      = "fvepdv-table"
	fvepdvURLPrefix    = "/api/fvepdv/"
)

var fvepdvTableFields = []domain.Fields{
	{Name: "vktip", Label: "Tip I/U", Width: "4"},
	{Name: "opis", Label: "Opis", Width: "100"},
	{Name: "vkrbr", Label: "Vrsta r.br.", Width: "3"},
	{Name: "obrazac", Label: "Obrazac", Width: "6"},
}

type FvepdvHandler struct {
	Service *service.BaseService[domain.Fvepdv]
}

func NewFvepdvHandler(service *service.BaseService[domain.Fvepdv]) *FvepdvHandler {
	return &FvepdvHandler{Service: service}
}
func (h *FvepdvHandler) CreateFvepdv(w http.ResponseWriter, r *http.Request) {
	var Fvepdv domain.Fvepdv
	utils.CreateHelper(w, r, &Fvepdv, h.Service, fvepdvTableFields)
}

func (h *FvepdvHandler) UpdateFvepdv(w http.ResponseWriter, r *http.Request) {
	var fvepdv domain.Fvepdv
	utils.UpdateHelper(w, r, &fvepdv, h.Service, fvepdvTableFields, utils.IDfvepdv)
}

func (h *FvepdvHandler) DeleteFvepdv(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDfvepdv)
}

func (h *FvepdvHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, fvepdvTableFields)
}

func (h *FvepdvHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(fvepdvURLPrefix, "/"), fvepdvTableFields)
}

func (h *FvepdvHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, fvepdvTableFields, utils.IDfvepdv)
}

func (h *FvepdvHandler) GetFvepdv(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, fvepdvTableFields, utils.IDfvepdv)
}

func (h *FvepdvHandler) GetAllFvepdv(w http.ResponseWriter, r *http.Request) {
	var Fvepdv domain.Fvepdv
	utils.GetAllEntityHelper(w, r, &Fvepdv, h.Service, fvepdvTableFields, fvepdvContentTitle, fvepdvTableID, fvepdvURLPrefix, utils.IDfvepdv)
}

func (h *FvepdvHandler) AddRoutes(r *http.ServeMux) {
	//define routes for Fvepdv
	r.HandleFunc("POST /api/fvepdv", infrastructure.AuthMiddleware(h.CreateFvepdv))
	r.HandleFunc("GET /api/fvepdv/all", infrastructure.AuthMiddleware(h.GetAllFvepdv))
	r.HandleFunc("GET /api/fvepdv/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/fvepdv/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/fvepdv/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/fvepdv/{id}", infrastructure.AuthMiddleware(h.GetFvepdv))
	r.HandleFunc("PUT /api/fvepdv/{id}", infrastructure.AuthMiddleware(h.UpdateFvepdv))
	r.HandleFunc("DELETE /api/fvepdv/{id}", infrastructure.AuthMiddleware(h.DeleteFvepdv))
}
