package finansijsko

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"helia/global"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	fkplContentTitle string = "KONTNI PLAN"
	fkplTableID      string = "fkpl-table"
	fkplURLPrefix    string = "/api/fkpl/"
	fkplURLGetAll    string = "/api/fkpl/all"
)

var fkplTableFields = []domain.Fields{
	{Name: "konto", Label: "Konto", Width: "10"},
	{Name: "sifra", Label: "Sifra", Width: "10"},
	{Name: "naziv", Label: "Naziv", Width: "120"},
	{Name: "vkonta", Label: "Vrsta konta", Width: "4"},
}
var fkplSearchTableFields = []domain.Fields{
	{Name: "konto", Label: "Konto", Width: "10"},
	{Name: "sifra", Label: "Sifra", Width: "10"},
	{Name: "naziv", Label: "Naziv", Width: "120"},
}

type FkplHandler struct {
	Service *service.BaseService[domain.Fkpl]
}

func NewFkplHandler(service *service.BaseService[domain.Fkpl]) *FkplHandler {
	return &FkplHandler{Service: service}
}

func (h *FkplHandler) CreateFkpl(w http.ResponseWriter, r *http.Request) {
	var fkpl domain.Fkpl
	utils.CreateHelper(w, r, &fkpl, h.Service, utils.IDfkpl, fkplTableFields)
}

func (h *FkplHandler) UpdateFkpl(w http.ResponseWriter, r *http.Request) {
	var fkpl domain.Fkpl
	utils.UpdateHelper(w, r, &fkpl, h.Service, fkplTableFields, utils.IDfkpl)
}

func (h *FkplHandler) DeleteFkpl(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper[domain.Fkpl](w, r, h.Service, utils.IDfkpl)
}

func (h *FkplHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, fkplTableFields)
}

func (h *FkplHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(fkplURLPrefix, "/"), fkplTableFields)
}

func (h *FkplHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper[domain.Fkpl](w, r, h.Service, fkplTableFields, utils.IDfkpl)
}

func (h *FkplHandler) GetFkpl(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, fkplTableFields, utils.IDfkpl)
}

func (h *FkplHandler) GetAllFkpl(w http.ResponseWriter, r *http.Request) {
	var fkpl domain.Fkpl
	tbl := utils.GetAllEntityHelper(w, r, &fkpl, h.Service, fkplTableFields, fkplContentTitle, fkplTableID, fkplURLPrefix, fkplURLGetAll, utils.IDfkpl)
	utils.RenderContent(w, r, *tbl)
}

func (h *FkplHandler) TraziKonto(w http.ResponseWriter, r *http.Request) {

	args := []interface{}{}
	// Parse query parameters from the URL

	konto := r.URL.Query().Get("konto")
	sifra := r.URL.Query().Get("sifra")
	vkonta := r.URL.Query().Get("vkonta")
	if vkonta == "" && konto == "" && sifra == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Nedostaje parametar query ili vkonta"))

		return

	}
	// Custom SQL query for searching konto, sifra, or naziv
	sqlQuery := `SELECT f.naziv
				FROM baza.fkpl as f`
	whereText := `WHERE 1 = 1  `
	hasGod, hasKAr := h.Service.Repo.GetHasGodHasKar()
	param := 1
	if hasGod {
		whereText += fmt.Sprintf(" AND f.god = $%d ", param)
		param++
		args = append(args, global.GetGnGod())
	}
	if hasKAr {
		whereText += fmt.Sprintf(" AND f.kar = $%d ", param)
		param++
		args = append(args, global.GetGnKar())
	}
	if konto != "" {
		whereText += fmt.Sprintf(" AND f.konto = $%d ", param)
		param++
		args = append(args, konto)
	}

	if sifra != "" {
		whereText += fmt.Sprintf(" AND f.sifra = $%d ", param)
		param++
		args = append(args, sifra)
	}
	if vkonta != "" {
		whereText += fmt.Sprintf(" AND f.vkonta = $%d", param)
		param++
		args = append(args, vkonta)
	}

	entities, err := h.Service.GetAllCustom(sqlQuery, whereText, args, "", "")
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Greška prilikom pretrage konta"))
		return
	}
	// Check if pointer is not nil and slice is not empty
	if entities != nil && len(*entities) > 0 {
		firstElement := (*entities)[0].Naziv
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(firstElement))
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if vkonta == "2" {
		w.Write([]byte("Nije pronađen konto"))
		return
	}
	if vkonta == "1" {
		w.Write([]byte("Nije pronađena šifra"))
		return
	}

}
func (h *FkplHandler) TraziKontoSearchTable(w http.ResponseWriter, r *http.Request) {

	args := []interface{}{}
	// Parse query parameters from the URL
	searchValue := r.URL.Query().Get("query")
	vkonta := r.URL.Query().Get("vkonta")
	if searchValue == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Nedostaje parametar za pretrazivanje"))

		return

	}
	// Custom SQL query for searching konto, sifra, or naziv
	sqlQuery := `SELECT idfkpl, f.konto, f.sifra, f.naziv
				FROM baza.fkpl as f`
	whereText := `WHERE f.god = $1 AND f.kar = $2 AND f.vkonta = $3 AND (f.konto ILIKE '%' || $4 || '%' 
										OR f.sifra ILIKE '%' || $5 || '%' 
										OR f.naziv ILIKE '%' || $6 || '%' ) ORDER BY konto LIMIT 20`

	args = append(args, global.GetGnGod(), global.GetGnKar(), vkonta, searchValue, searchValue, searchValue)
	entities, err := h.Service.GetAllCustom(sqlQuery, whereText, args, "", "")
	if err != nil {
		response := utils.CreateResponse(w, false, []domain.FieldError{}, utils.ReadDataErrMsg, http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	// Convert the fetched data into the format expected by the template
	tbl := common.SetTableBasicData("", prometTableID, fkplSearchTableFields, "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	tbl.ShowPagination = false
	// Prepare TableData for UI
	tblRows, err := common.SetTableRows(&tbl, *entities, fkplSearchTableFields, "idfkpl", "", h.Service.GetFieldCache())
	if err != nil {
		http.Error(w, "Failed to set table rows", http.StatusInternalServerError)
		return
	}
	tbl.Rows = tblRows.Rows
	tbl.BtnAdd = domain.Button{IsVisible: false}   // Hide Add button in this view
	tbl.BtnPrint = domain.Button{IsVisible: false} // Hide Print button in this view

	w.Header().Set("Content-Type", "text/html")
	utils.RenderContent(w, r, tbl)
}
func (h *FkplHandler) AddRoutes(r *http.ServeMux) {
	// Define routes for fkpl
	r.HandleFunc("POST /api/fkpl", infrastructure.AuthMiddleware(h.CreateFkpl))
	r.HandleFunc("GET /api/fkpl/all", infrastructure.AuthMiddleware(h.GetAllFkpl))
	r.HandleFunc("GET /api/fkpl/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/fkpl/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/fkpl/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/fkpl/{id}", infrastructure.AuthMiddleware(h.GetFkpl))
	r.HandleFunc("PUT /api/fkpl/{id}", infrastructure.AuthMiddleware(h.UpdateFkpl))
	r.HandleFunc("DELETE /api/fkpl/{id}", infrastructure.AuthMiddleware(h.DeleteFkpl))
	r.HandleFunc("GET /api/fkpl/trazikonto", infrastructure.AuthMiddleware(h.TraziKonto))
	r.HandleFunc("GET /api/fkpl/trazikontosearchtable", infrastructure.AuthMiddleware(h.TraziKontoSearchTable))
}
