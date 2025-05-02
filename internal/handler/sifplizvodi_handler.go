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
	sifplizvContentTitle = "SIFRE PLACANJA ZA IZVODE"
	sifplizvTableID      = "sifplizv-table"
	sifplizvURLPrefix    = "/api/sifplizvodi/"
)

var sifplizvTableFields = []domain.Fields{
	{Name: "sifplac", Label: "Sifra Placanja", ValidationText: "morate uneti sifru placanja...", Width: "6"},
	{Name: "oblik", Label: "Oblik", ValidationText: "morate uneti oblik placanja...", Width: "6"},
	{Name: "osnov", Label: "Osnov placanja", ValidationText: "morate uneti osnov placanja", Width: "4"},
	{Name: "opis", Label: "Opis", ValidationText: "morate uneti opis", Width: "80"},
	{Name: "konto", Label: "Konto", Width: "8"},
	{Name: "sifra", Label: "Sifra", Width: "8"},
}

type SifplizvHandler struct {
	Service *service.BaseService[domain.Sifplizv]
}

func NewSifplizvHandler(service *service.BaseService[domain.Sifplizv]) *SifplizvHandler {
	return &SifplizvHandler{Service: service}
}

func (h *SifplizvHandler) CreateSifplizv(w http.ResponseWriter, r *http.Request) {
	var sifplizv domain.Sifplizv
	utils.CreateHelper(w, r, &sifplizv, h.Service, sifplizvTableFields)
}

func (h *SifplizvHandler) UpdateSifplizv(w http.ResponseWriter, r *http.Request) {
	var sifplizv domain.Sifplizv
	utils.UpdateHelper(w, r, &sifplizv, h.Service, sifplizvTableFields, utils.IDsifplizv)
}

func (h *SifplizvHandler) DeleteSifplizv(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDsifplizv)
}

func (h *SifplizvHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, sifplizvTableFields)
}

func (h *SifplizvHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(sifplizvURLPrefix, "/"), sifplizvTableFields)
}

func (h *SifplizvHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, sifplizvTableFields, utils.IDsifplizv)
}

func (h *SifplizvHandler) GetSifplizv(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, sifplizvTableFields, utils.IDsifplizv)
}

func (h *SifplizvHandler) GetAllSifplizv(w http.ResponseWriter, r *http.Request) {
	var sifplizv domain.Sifplizv
	utils.GetAllEntityHelper(w, r, &sifplizv, h.Service, sifplizvTableFields, sifplizvContentTitle, sifplizvTableID, sifplizvURLPrefix, utils.IDsifplizv)
}

func (h *SifplizvHandler) AddRoutes(r *http.ServeMux) {
	//define routes for sifplizv
	r.HandleFunc("POST /api/sifplizvodi", infrastructure.AuthMiddleware(h.CreateSifplizv))
	r.HandleFunc("GET /api/sifplizvodi/all", infrastructure.AuthMiddleware(h.GetAllSifplizv))
	r.HandleFunc("GET /api/sifplizvodi/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/sifplizvodi/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/sifplizvodi/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/sifplizvodi/{id}", infrastructure.AuthMiddleware(h.GetSifplizv))
	r.HandleFunc("PUT /api/sifplizvodi/{id}", infrastructure.AuthMiddleware(h.UpdateSifplizv))
	r.HandleFunc("DELETE /api/sifplizvodi/{id}", infrastructure.AuthMiddleware(h.DeleteSifplizv))
}
