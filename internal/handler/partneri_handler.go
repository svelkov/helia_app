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
	partneriContentTitle = "POSLOVNI PARTNERI"
	partneriTableID      = "partneri-table"
	partneriURLPrefix    = "/api/partneri/"
)

var partneriTableFields = []domain.Fields{
	{Name: "sifra", Label: "Sifra", ValidationText: "morate uneti sifru partnera...", Width: "8"},
	{Name: "naziv", Label: "Naziv", ValidationText: "morate uneti naziv partnera..", Width: "60"},
	{Name: "adresa", Label: "Adresa", Width: "40"},
	{Name: "pobro", Label: "Postanski broj", Width: "8"},
	{Name: "mesto", Label: "Mesto", Width: "40"},
	{Name: "pib", Label: "PIB", Width: "12"},
	{Name: "jmbg", Label: "JMBG", Width: "15"},
	{Name: "bpg", Label: "BPG"},
	{Name: "index", Label: "Index"},
	{Name: "gln", Label: "GLN"},
	{Name: "jib", Label: "JIB"},
	{Name: "ziro", Label: "Ziro"},
	{Name: "matbr", Label: "Maticni Broj"},
	{Name: "konta", Label: "Konta"},
	{Name: "tippdv", Label: "Tip PDV"},
	{Name: "email", Label: "E-Mail"},
	{Name: "telefon", Label: "Telefon"},
	{Name: "kontaktosb", Label: "Kontakt osoba"},
	{Name: "budzetski", Label: "Budzetski"},
	{Name: "jbkjs", Label: "JBKJS"},
	{Name: "napomena", Label: "Naponema"},
	{Name: "idpartneri", Label: "ID Partneri"},
}

type PartneriHandler struct {
	Service *service.BaseService[domain.Partneri]
}

func NewPartneriHandler(service *service.BaseService[domain.Partneri]) *PartneriHandler {
	return &PartneriHandler{Service: service}
}
func (h *PartneriHandler) CreatePartneri(w http.ResponseWriter, r *http.Request) {
	var partneri domain.Partneri
	utils.CreateHelper(w, r, &partneri, h.Service, partneriTableFields)
}

func (h *PartneriHandler) UpdatePartneri(w http.ResponseWriter, r *http.Request) {
	var partneri domain.Partneri
	utils.UpdateHelper(w, r, &partneri, h.Service, partneriTableFields, utils.IDpartneri)
}

func (h *PartneriHandler) DeletePartneri(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDpartneri)
}

func (h *PartneriHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, partneriTableFields)
}

func (h *PartneriHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(partneriURLPrefix, "/"), partneriTableFields)
}

func (h *PartneriHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, partneriTableFields, utils.IDpartneri)
}

func (h *PartneriHandler) GetPartneri(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, partneriTableFields, utils.IDpartneri)
}

func (h *PartneriHandler) GetAllPartneri(w http.ResponseWriter, r *http.Request) {
	var partneri domain.Partneri
	utils.GetAllEntityHelper(w, r, &partneri, h.Service, partneriTableFields, partneriContentTitle, partneriTableID, partneriURLPrefix, utils.IDpartneri)
}

func (h *PartneriHandler) AddRoutes(r *http.ServeMux) {
	//define routes for partneri
	r.HandleFunc("POST /api/partneri", infrastructure.AuthMiddleware(h.CreatePartneri))
	r.HandleFunc("GET /api/partneri/all", infrastructure.AuthMiddleware(h.GetAllPartneri))
	r.HandleFunc("GET /api/partneri/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/partneri/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/partneri/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/partneri/{id}", infrastructure.AuthMiddleware(h.GetPartneri))
	r.HandleFunc("PUT /api/partneri/{id}", infrastructure.AuthMiddleware(h.UpdatePartneri))
	r.HandleFunc("DELETE /api/partneri/{id}", infrastructure.AuthMiddleware(h.DeletePartneri))
}
