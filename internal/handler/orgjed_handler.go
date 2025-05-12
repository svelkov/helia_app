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
	orgjedContentTitle = "ORGANIZACIONE JEDINICE"
	orgjedTableID      = "orgjed-table"
	orgjedURLPrefix    = "/api/orgjed/"
)

var orgjedTableFields = []domain.Fields{
	{Name: "ojozn", Label: "Sifra Orgjed", Width: "6"},
	{Name: "naziv", Label: "Naziv Orgjed", Width: "45"},
}

type OrgjedHandler struct {
	Service *service.BaseService[domain.Orgjed]
}

func NewOrgjedHandler(service *service.BaseService[domain.Orgjed]) *OrgjedHandler {
	return &OrgjedHandler{Service: service}
}

func (h *OrgjedHandler) CreateOrgjed(w http.ResponseWriter, r *http.Request) {
	var orgjed domain.Orgjed
	utils.CreateHelper(w, r, &orgjed, h.Service, orgjedTableFields)
}

func (h *OrgjedHandler) UpdateOrgjed(w http.ResponseWriter, r *http.Request) {
	var orgjed domain.Orgjed
	utils.UpdateHelper(w, r, &orgjed, h.Service, orgjedTableFields, utils.IDorgjed)
}

func (h *OrgjedHandler) DeleteOrgjed(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDorgjed)
}

func (h *OrgjedHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, orgjedTableFields)
}

func (h *OrgjedHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(orgjedURLPrefix, "/"), orgjedTableFields)
}

func (h *OrgjedHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, orgjedTableFields, utils.IDorgjed)
}

func (h *OrgjedHandler) GetOrgjed(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, orgjedTableFields, utils.IDorgjed)
}

func (h *OrgjedHandler) GetAllOrgjed(w http.ResponseWriter, r *http.Request) {
	var orgjed domain.Orgjed
	tbl := utils.GetAllEntityHelper(w, r, &orgjed, h.Service, orgjedTableFields, orgjedContentTitle, orgjedTableID, orgjedURLPrefix, utils.IDorgjed)
	utils.RenderContent(w, r, *tbl)
}

func (h *OrgjedHandler) AddRoutes(r *http.ServeMux) {
	//define routes for orgjed
	r.HandleFunc("POST /api/orgjed", infrastructure.AuthMiddleware(h.CreateOrgjed))
	r.HandleFunc("GET /api/orgjed/all", infrastructure.AuthMiddleware(h.GetAllOrgjed))
	r.HandleFunc("GET /api/orgjed/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/orgjed/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/orgjed/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/orgjed/{id}", infrastructure.AuthMiddleware(h.GetOrgjed))
	r.HandleFunc("PUT /api/orgjed/{id}", infrastructure.AuthMiddleware(h.UpdateOrgjed))
	r.HandleFunc("DELETE /api/orgjed/{id}", infrastructure.AuthMiddleware(h.DeleteOrgjed))
}
