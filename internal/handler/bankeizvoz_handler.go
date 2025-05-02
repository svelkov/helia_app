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
	bnkizvContentTitle = "BANKE ZA IZVOZNE FAKTURE"
	bnkizvTableID      = "bnkizv-table"
	bnkizvURLPrefix    = "/api/bnkizv/"
)

var bnkizvTableFields = []domain.Fields{
	{Name: "sifbank", Label: "Sifra Banke", Width: "6"},
	{Name: "bnkdes", Label: "Naziv Banka", Width: "80"},
	{Name: "swiftadr", Label: "Swift Adresa", Width: "80"},
	{Name: "brojrac", Label: "Broj Racuna", Width: "55"},
	{Name: "beneficiary", Label: "Beneficiary", Width: "70"},
	{Name: "corrbank", Label: "Korespodentna banka", Width: "90"},
	{Name: "tel", Label: "Telefon", Width: "25"},
	{Name: "fax", Label: "Fax", Width: "30"},
	{Name: "address", Label: "Adresa", Width: "40"},
	{Name: "komentar", Label: "Komentar", Width: "70"},
}

type BnkizvHandler struct {
	Service *service.BaseService[domain.Bnkizv]
}

func NewBnkizvHandler(service *service.BaseService[domain.Bnkizv]) *BnkizvHandler {
	return &BnkizvHandler{Service: service}
}

func (h *BnkizvHandler) CreateBnkizv(w http.ResponseWriter, r *http.Request) {
	var bnkizv domain.Bnkizv
	utils.CreateHelper(w, r, &bnkizv, h.Service, bnkizvTableFields)
}

func (h *BnkizvHandler) UpdateBnkizv(w http.ResponseWriter, r *http.Request) {
	var bnkizv domain.Bnkizv
	utils.UpdateHelper(w, r, &bnkizv, h.Service, bnkizvTableFields, utils.IDbnkizv)
}

func (h *BnkizvHandler) DeleteBnkizv(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDbnkizv)
}

func (h *BnkizvHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, bnkizvTableFields)
}

func (h *BnkizvHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(bnkizvURLPrefix, "/"), bnkizvTableFields)
}

func (h *BnkizvHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, bnkizvTableFields, utils.IDbnkizv)
}

func (h *BnkizvHandler) GetBnkizv(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, bnkizvTableFields, utils.IDbnkizv)
}

func (h *BnkizvHandler) GetAllBnkizv(w http.ResponseWriter, r *http.Request) {
	var bnkizv domain.Bnkizv
	utils.GetAllEntityHelper(w, r, &bnkizv, h.Service, bnkizvTableFields, bnkizvContentTitle, bnkizvTableID, bnkizvURLPrefix, utils.IDbnkizv)
}

func (h *BnkizvHandler) AddRoutes(r *http.ServeMux) {
	//define routes for bnkizv
	r.HandleFunc("POST /api/bnkizv", infrastructure.AuthMiddleware(h.CreateBnkizv))
	r.HandleFunc("GET /api/bnkizv/all", infrastructure.AuthMiddleware(h.GetAllBnkizv))
	r.HandleFunc("GET /api/bnkizv/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/bnkizv/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/bnkizv/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/bnkizv/{id}", infrastructure.AuthMiddleware(h.GetBnkizv))
	r.HandleFunc("PUT /api/bnkizv/{id}", infrastructure.AuthMiddleware(h.UpdateBnkizv))
	r.HandleFunc("DELETE /api/bnkizv/{id}", infrastructure.AuthMiddleware(h.DeleteBnkizv))
}
