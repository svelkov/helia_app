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
	bankeContentTitle = "BANKE"
	bankeTableID      = "banke-table"
	bankeURLPrefix    = "/api/banke/"
)

var bankeTableFields = []domain.Fields{
	{Name: "brrac", Label: "Broj Racuna", Width: "45"},
	{Name: "banka", Label: "Banka", Width: "60"},
	{Name: "konto", Label: "Konto", Width: "8"},
	{Name: "sifra", Label: "Sifra", Width: "8"},
	{Name: "bnkcod", Label: "Kod Banke", Width: "4"},
	{Name: "ebank", Label: "E-Banking", Width: "60"},
	{Name: "pocnazfajl", Label: "Poc. Naz. Fajla", Width: "10"},
	{Name: "tipdok", Label: "Tip Dokumenta", Width: "6"},
	{Name: "nafakne", Label: "Na Fakt Ne", Width: "1"},
}

type BankeHandler struct {
	Service service.Service[domain.Banke] // Use the interface
}

func NewBankeHandler(service service.Service[domain.Banke]) *BankeHandler {
	return &BankeHandler{Service: service}
}

func (h *BankeHandler) CreateBanke(w http.ResponseWriter, r *http.Request) {
	var banke domain.Banke
	utils.CreateHelper(w, r, &banke, h.Service, bankeTableFields)
}

func (h *BankeHandler) UpdateBanke(w http.ResponseWriter, r *http.Request) {
	var banke domain.Banke
	utils.UpdateHelper(w, r, &banke, h.Service, bankeTableFields, utils.IDbanke)
}

func (h *BankeHandler) DeleteBanke(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDbanke)
}

func (h *BankeHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, bankeTableFields)
}

func (h *BankeHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(bankeURLPrefix, "/"), bankeTableFields)
}

func (h *BankeHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, bankeTableFields, utils.IDbanke)
}

func (h *BankeHandler) GetBanke(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, bankeTableFields, utils.IDbanke)
}

func (h *BankeHandler) GetAllBanke(w http.ResponseWriter, r *http.Request) {
	var banke domain.Banke
	tbl := utils.GetAllEntityHelper(w, r, &banke, h.Service, bankeTableFields, bankeContentTitle, bankeTableID, bankeURLPrefix, utils.IDbanke)
	utils.RenderContent(w, r, *tbl)
}

func (h *BankeHandler) AddRoutes(r *http.ServeMux) {
	//define routes for Banke
	r.HandleFunc("POST /api/banke", infrastructure.AuthMiddleware(h.CreateBanke))
	r.HandleFunc("GET /api/banke/all", infrastructure.AuthMiddleware(h.GetAllBanke))
	r.HandleFunc("GET /api/banke/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/banke/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/banke/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/banke/{id}", infrastructure.AuthMiddleware(h.GetBanke))
	r.HandleFunc("PUT /api/banke/{id}", infrastructure.AuthMiddleware(h.UpdateBanke))
	r.HandleFunc("DELETE /api/banke/{id}", infrastructure.AuthMiddleware(h.DeleteBanke))
}
