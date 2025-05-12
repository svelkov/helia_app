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
	fvknjracContentTitle = "VRSTE PORESKIH KNJIGA (KPR i KIR)"
	fvknjracTableID      = "fvknjrac-table"
	fvknjracURLPrefix    = "/api/fvknjrac/"
)

var fvknjracTableFields = []domain.Fields{
	{Name: "vktip", Label: "Tip knjige (ulazni/izlazni racun)", Width: "2"},
	{Name: "opis", Label: "Opis", Width: "50"},
	{Name: "vkrbr", Label: "Vrsta knjige racuna", Width: "2"},
	{Name: "konta", Label: "Konta", Width: "60"},
}

type FvknjracHandler struct {
	Service *service.BaseService[domain.Fvknjrac]
}

func NewFvknjracHandler(service *service.BaseService[domain.Fvknjrac]) *FvknjracHandler {
	return &FvknjracHandler{Service: service}
}
func (h *FvknjracHandler) CreateFvknjrac(w http.ResponseWriter, r *http.Request) {
	var fvknjrac domain.Fvknjrac
	utils.CreateHelper(w, r, &fvknjrac, h.Service, fvknjracTableFields)
}

func (h *FvknjracHandler) UpdateFvknjrac(w http.ResponseWriter, r *http.Request) {
	var fvknjrac domain.Fvknjrac
	utils.UpdateHelper(w, r, &fvknjrac, h.Service, fvknjracTableFields, utils.IDfvknjrac)
}

func (h *FvknjracHandler) DeleteFvknjrac(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDfvknjrac)
}

func (h *FvknjracHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, fvknjracTableFields)
}

func (h *FvknjracHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(fvknjracURLPrefix, "/"), fvknjracTableFields)
}

func (h *FvknjracHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, fvknjracTableFields, utils.IDfvknjrac)
}

func (h *FvknjracHandler) GetFvknjrac(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, fvknjracTableFields, utils.IDfvknjrac)
}

func (h *FvknjracHandler) GetAllFvknjrac(w http.ResponseWriter, r *http.Request) {
	var fvknjrac domain.Fvknjrac
	tbl := utils.GetAllEntityHelper(w, r, &fvknjrac, h.Service, fvknjracTableFields, fvknjracContentTitle, fvknjracTableID, fvknjracURLPrefix, utils.IDfvknjrac)
	utils.RenderContent(w, r, *tbl)
}

func (h *FvknjracHandler) AddRoutes(r *http.ServeMux) {
	//define routes for fvknjrac
	r.HandleFunc("POST /api/fvknjrac", infrastructure.AuthMiddleware(h.CreateFvknjrac))
	r.HandleFunc("GET /api/fvknjrac/all", infrastructure.AuthMiddleware(h.GetAllFvknjrac))
	r.HandleFunc("GET /api/fvknjrac/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/fvknjrac/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/fvknjrac/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/fvknjrac/{id}", infrastructure.AuthMiddleware(h.GetFvknjrac))
	r.HandleFunc("PUT /api/fvknjrac/{id}", infrastructure.AuthMiddleware(h.UpdateFvknjrac))
	r.HandleFunc("DELETE /api/fvknjrac/{id}", infrastructure.AuthMiddleware(h.DeleteFvknjrac))
}
