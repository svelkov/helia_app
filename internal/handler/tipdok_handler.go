package handler

import (
	"net/http"
	"strings"

	rpt "helia/frontend/templates/reports"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	tipdokContentTitle string = "VRSTE NALOGA"
	tipdokTableID      string = "tipdok-table"
	tipdokURLPrefix    string = "/api/tipdok/"
	tipdokURLGetAll    string = "/api/tipdok/all1"
)

// key of the map must be the name of filed in the table in db (we need it for mapping)
var tipdokTableFields = []domain.Fields{
	{Name: "tipdok", Label: "Vrsta Naloga", ValidationText: "Morate uneti tip dokumenta...", Width: "10"},
	{Name: "opis", Label: "Opis", ValidationText: "Morate uneti opis dokumenta...", Width: "60"},
	{Name: "grpdok", Label: "Grupa Dok.", ValidationText: "Morate uneti grupu dokumenata...", Width: "20"},
	{Name: "grpvrd", Label: "Grp. Vrste Dok.", ValidationText: "Morate uneti grupu vrste dokumenata...", Width: "20"},
	{Name: "magacin", Label: "Magacin", ValidationText: "", Width: "10"},
}

type TipdokHandler struct {
	Service *service.BaseService[domain.Tipdok]
}

func NewTipdokHandler(service *service.BaseService[domain.Tipdok]) *TipdokHandler {
	return &TipdokHandler{Service: service}
}

func (h *TipdokHandler) CreateTipdok(w http.ResponseWriter, r *http.Request) {
	var tipdok domain.Tipdok
	utils.CreateHelper(w, r, &tipdok, h.Service, utils.IDtipdok, tipdokTableFields)
}

func (h *TipdokHandler) UpdateTipdok(w http.ResponseWriter, r *http.Request) {
	var tipdok domain.Tipdok
	utils.UpdateHelper(w, r, &tipdok, h.Service, tipdokTableFields, utils.IDtipdok)
}

func (h *TipdokHandler) DeleteTipdok(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDtipdok)
}

func (h *TipdokHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, tipdokTableFields)
}

func (h *TipdokHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(tipdokURLPrefix, "/"), tipdokTableFields)
}

func (h *TipdokHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, tipdokTableFields, utils.IDtipdok)
}

func (h *TipdokHandler) GetTipdok(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, tipdokTableFields, utils.IDtipdok)
}

func (h *TipdokHandler) GetAllTipdok(w http.ResponseWriter, r *http.Request) {
	var tipdok domain.Tipdok
	tbl := utils.GetAllEntityHelper(w, r, &tipdok, h.Service, tipdokTableFields, tipdokContentTitle, tipdokTableID, tipdokURLPrefix, tipdokURLGetAll, utils.IDtipdok)
	utils.RenderContent(w, r, *tbl)
}
func (h *TipdokHandler) ReportHandler(w http.ResponseWriter, r *http.Request) {
	params := map[string]interface{}{
		"User":    "admin",
		"Filters": "Active records only",
	}

	var tipdok domain.Tipdok
	tbl := utils.GetAllEntityHelper(w, r, &tipdok, h.Service, tipdokTableFields, tipdokContentTitle, tipdokTableID, tipdokURLPrefix, tipdokURLGetAll, utils.IDtipdok)

	component := rpt.ReportPage(*tbl, params, "Your Company")

	component.Render(r.Context(), w)
}

func (h *TipdokHandler) AddRoutes(r *http.ServeMux) {

	//define routes for tipdok
	r.HandleFunc("POST /api/tipdok", infrastructure.AuthMiddleware(h.CreateTipdok))
	r.HandleFunc("GET /api/tipdok/all", infrastructure.AuthMiddleware(h.GetAllTipdok))
	r.HandleFunc("GET /api/tipdok/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/tipdok/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/tipdok/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/tipdok/{id}", infrastructure.AuthMiddleware(h.GetTipdok))
	r.HandleFunc("PUT /api/tipdok/{id}", infrastructure.AuthMiddleware(h.UpdateTipdok))
	r.HandleFunc("DELETE /api/tipdok/{id}", infrastructure.AuthMiddleware(h.DeleteTipdok))
	r.HandleFunc("GET /api/tipdok/report", infrastructure.AuthMiddleware(h.ReportHandler))
}
