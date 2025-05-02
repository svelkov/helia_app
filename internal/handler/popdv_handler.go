package handler

import (
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
	"net/http"
	"strings"
)

const (
	popdvContentTitle = "VRSTE NALOGA"
	popdvTableID      = "popdv-table"
	popdvURLPrefix    = "/api/popdv/"
)

// key of the map must be the name of filed in the table in db (we need it for mapping)
var popdvTableFields = []domain.Fields{
	{Name: "popdv", Label: "Vrsta Naloga", ValidationText: "Morate uneti tip dokumenta...", Width: "10"},
	{Name: "opis", Label: "Opis", ValidationText: "Morate uneti opis dokumenta...", Width: "60"},
	{Name: "grpdok", Label: "Grupa Dok.", ValidationText: "Morate uneti grupu dokumenata...", Width: "20"},
	{Name: "grpvrd", Label: "Grp. Vrste Dok.", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20"},
	{Name: "magacin", Label: "Magacin", ValidationText: "", Width: "10"},
}

type PopdvHandler struct {
	Service *service.BaseService[domain.Popdv]
}

func NewPopdvHandler(service *service.BaseService[domain.Popdv]) *PopdvHandler {
	return &PopdvHandler{Service: service}
}

func (h *PopdvHandler) CreatePopdv(w http.ResponseWriter, r *http.Request) {
	var popdv domain.Popdv
	utils.CreateHelper(w, r, &popdv, h.Service, popdvTableFields)
}

func (h *PopdvHandler) UpdatePopdv(w http.ResponseWriter, r *http.Request) {
	var popdv domain.Popdv
	utils.UpdateHelper(w, r, &popdv, h.Service, popdvTableFields, utils.IDpopdv)
}

func (h *PopdvHandler) DeletePopdv(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDpopdv)
}

func (h *PopdvHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, popdvTableFields)
}

func (h *PopdvHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(popdvURLPrefix, "/"), popdvTableFields)
}

func (h *PopdvHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, popdvTableFields, utils.IDpopdv)
}

func (h *PopdvHandler) GetPopdv(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, popdvTableFields, utils.IDpopdv)
}

func (h *PopdvHandler) GetAllPopdv(w http.ResponseWriter, r *http.Request) {
	var popdv domain.Popdv
	utils.GetAllEntityHelper(w, r, &popdv, h.Service, popdvTableFields, popdvContentTitle, popdvTableID, popdvURLPrefix, utils.IDpopdv)
}

func (h *PopdvHandler) AddRoutes(r *http.ServeMux) {
	//define routes for popdv
	r.HandleFunc("POST /api/popdv", infrastructure.AuthMiddleware(h.CreatePopdv))
	r.HandleFunc("GET /api/popdv/all", infrastructure.AuthMiddleware(h.GetAllPopdv))
	r.HandleFunc("GET /api/popdv/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/popdv/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/popdv/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/popdv/{id}", infrastructure.AuthMiddleware(h.GetPopdv))
	r.HandleFunc("PUT /api/popdv/{id}", infrastructure.AuthMiddleware(h.UpdatePopdv))
	r.HandleFunc("DELETE /api/popdv/{id}", infrastructure.AuthMiddleware(h.DeletePopdv))
}
