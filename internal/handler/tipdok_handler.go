package handler

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	tipdokContentTitle = "VRSTE NALOGA"
	tipdokTableID      = "tipdok-table"
	tipdokURLPrefix    = "/api/tipdok/"
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
	utils.CreateHelper(w, r, &tipdok, h.Service, tipdokTableFields)
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
	utils.GetAllEntityHelper(w, r, &tipdok, h.Service, tipdokTableFields, tipdokContentTitle, tipdokTableID, tipdokURLPrefix, utils.IDtipdok)
}
func (h *TipdokHandler) GetAllTipdokNalozi(w http.ResponseWriter, r *http.Request) {

	// Fetch all "tipdok" entities from the service layer
	args := []interface{}{}
	hasGod, hasKar := h.Service.Repo.CheckGogKar()
	basicWhere := h.Service.Repo.CreateBasicWhere(tipdokTableFields, &args, hasGod, hasKar)
	queryText := "SELECT idtipdok, tipdok, opis FROM tipdok "
	whereText := fmt.Sprintf(" %s AND (grpdok = 'FIN' OR grpdok = 'SVI') ", basicWhere)
	orderBy := " ORDER BY tipdok"

	tipdokData, err := h.Service.GetAllCustom(queryText, whereText, args, "", orderBy)
	if err != nil {
		http.Error(w, "Error getting tipdok's", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("options").Parse(`
        {{range .}}
            <option value="{{.TipDok}}">{{.TipDok}} - {{.Opis}}</option>
        {{end}}
    `)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8") // Important: set content type to HTML
	err = tmpl.Execute(w, tipdokData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// responseData := make([]domain.TipdokComboItem, len(*tipdokData))
	// for i, td := range *tipdokData {
	// 	responseData[i] = domain.TipdokComboItem{
	// 		TipDok: td.TipDok,
	// 		Opis:   td.Opis,
	// 	}
	// }

	// w.Header().Set("Content-Type", "application/json")
	// err = json.NewEncoder(w).Encode(responseData)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
}

func (h *TipdokHandler) AddRoutes(r *http.ServeMux) {

	//define routes for tipdok
	r.HandleFunc("POST /api/tipdok", infrastructure.AuthMiddleware(h.CreateTipdok))
	r.HandleFunc("GET /api/tipdok/all", infrastructure.AuthMiddleware(h.GetAllTipdok))
	r.HandleFunc("GET /api/tipdok/all/nalozi", infrastructure.AuthMiddleware(h.GetAllTipdokNalozi))
	r.HandleFunc("GET /api/tipdok/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/tipdok/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/tipdok/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/tipdok/{id}", infrastructure.AuthMiddleware(h.GetTipdok))
	r.HandleFunc("PUT /api/tipdok/{id}", infrastructure.AuthMiddleware(h.UpdateTipdok))
	r.HandleFunc("DELETE /api/tipdok/{id}", infrastructure.AuthMiddleware(h.DeleteTipdok))
}
