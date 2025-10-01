package finansijsko

import (
	"fmt"
	"net/http"

	tmpl "helia/frontend/templates"
	tmpl_fin "helia/frontend/templates/finansijsko"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	prometContentTitle string = "PROMET"
	prometTableID      string = "promettable"
	prometURLPrefix    string = "/api/promet/"
	prometURLGetAll    string = "/api/promet/all"
)

type PrometHandler struct {
	tabData domain.TabData
	service *service.PrometResource
}

func NewPrometHandler(service *service.PrometResource) *PrometHandler {
	handler := &PrometHandler{}
	handler.tabData = GetTabData()
	handler.service = service
	return handler
}

func (h *PrometHandler) PrometMain(w http.ResponseWriter, r *http.Request) {
	// Create configuration
	// popupConfig := domain.NewSearchPopupConfig("konto", "/api/fkpl/trazikonto", "Trazi konto...")
	// popupConfig.Width = "w-296"         // Custom width
	// popupConfig.MaxHeight = "max-h-280" // Custom height
	hxVals := `js:{
            konto: document.getElementById("konto")?.value,
            sifra: document.getElementById("sifra")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value
        }`
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/analitickakonta", "#promettable", "innerHTML", "GET", "", hxVals, true)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false

	w.Header().Set("Content-Type", "text/html")

	err := tmpl_fin.PrometMain(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (h *PrometHandler) PrometAnalitickihKonta(w http.ResponseWriter, r *http.Request) {
	// Get our custom header
	requestSource := r.Header.Get("X-Request-Source")
	if requestSource == "menu" || requestSource == "tab" {
	hxVals := `js:{
            konto: document.getElementById("konto")?.value,
            sifra: document.getElementById("sifra")?.value,
            oddatuma: document.getElementById("oddatuma")?.value,
            dodatuma: document.getElementById("dodatuma")?.value
        }`
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/analitickakonta", "#promettable", "innerHTML", "GET", "", hxVals, true)

	//if the call come from menu click or tab click then render the page with parameters and empty table
		tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnkontaTableFields(), "", "", 0, 0, 0, 0)
		tbl.ShowActions = false

		w.Header().Set("Content-Type", "text/html")
		h.tabData = setActiveTab(h.tabData, "analitickakonta")
		err := tmpl_fin.PrometAnalitickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			return
		}
	}
	// If it's a POST request, the make obrada
	if requestSource == "btnobrada" || requestSource == "btnpage" {
		page, pageSize := common.GetPageAndPageSizeFromRequest(r)
		response, err := h.service.GetPrometAnalitickihKonta(r, true, 0, 0)
		if err != nil {
			http.Error(w, "Failed to get total records", http.StatusInternalServerError)
			return
		}
		totalRecords := response.TotalRecords
		totalPages := (totalRecords + pageSize - 1) / pageSize
		// Get paginated data
		response, err = h.service.GetPrometAnalitickihKonta(r, false, pageSize, page)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			return
		}
		tbl := common.SetTableBasicData("", prometTableID, h.service.GetAnkontaTableFields(), "", "/api/promet/analitickakonta", pageSize, page, totalPages, totalRecords)
		tbl.ShowActions = false
		tbl.ShowPagination = true
		tbl.Pagination.HxInclude = "#konto, #sifra, #oddatuma, #dodatuma"
		// Prepare TableData for UI
		tblRows, err := common.SetTableRows(&tbl, response.Data, h.service.GetAnkontaTableFields(), "idfpro", "", h.service.GetFieldCache())
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
}

func (h *PrometHandler) PrometAnalitickihKontaPoMI(w http.ResponseWriter, r *http.Request) {
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/analitickakontami", "#promettable", "innerHTML", "POST", "", "", true)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnKontaMiTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	w.Header().Set("Content-Type", "text/html")
	h.tabData = setActiveTab(h.tabData, "analitickakontami")
	err := tmpl_fin.AnalitickaKarticaPoMI(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
func (h *PrometHandler) PrometDeviznihAnalitickihKonta(w http.ResponseWriter, r *http.Request) {
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/deviznihanalitickihkonta", "#promettable", "innerHTML", "POST", "", "", true)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetAnDeviznaKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	w.Header().Set("Content-Type", "text/html")
	h.tabData = setActiveTab(h.tabData, "deviznihanalitickihkonta")
	err := tmpl_fin.PrometDeviznihAnalitickihKonta(h.tabData, tbl, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (h *PrometHandler) PrometSubsintetickihKonta(w http.ResponseWriter, r *http.Request) {
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/subsintetickakonta", "#promettable", "innerHTML", "POST", "", "", true)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSubsintetickihKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	w.Header().Set("Content-Type", "text/html")
	h.tabData = setActiveTab(h.tabData, "subsintetickakonta")
	err := tmpl_fin.PrometSubsintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
func (h *PrometHandler) PrometSintetickihKonta(w http.ResponseWriter, r *http.Request) {
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/sintetickakonta", "#promet-table", "innerHTML", "POST", "", "", true)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetSintetickihKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	w.Header().Set("Content-Type", "text/html")
	h.tabData = setActiveTab(h.tabData, "sintetickakonta")
	err := tmpl_fin.PrometSintetickihKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
func (h *PrometHandler) KarticaSintetickihKonta(w http.ResponseWriter, r *http.Request) {
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/karticasintetickihkonta", "#promettable", "innerHTML", "POST", "", "", true)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKarticaSintetickihKontaTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	w.Header().Set("Content-Type", "text/html")
	h.tabData = setActiveTab(h.tabData, "karticasintetickihkonta")
	err := tmpl_fin.KarticaSintetickiKonta(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (h *PrometHandler) PrometKontaAnaliticki(w http.ResponseWriter, r *http.Request) {
	btnPrint := common.SetButton("print-btn", "Štampa", "stampa", "", "#tab-content", "innerHTML", "GET", "", "", true)
	btnObrada := common.SetButton("obrada-btn", "Obrada", "fin_obrada", "/api/promet/kontaanaliticki", "#promettable", "innerHTML", "POST", "", "", true)

	tbl := common.SetTableBasicData(prometContentTitle, prometTableID, h.service.GetKontaAnalitickiTableFields(), "", "", 0, 0, 0, 0)
	tbl.ShowActions = false
	w.Header().Set("Content-Type", "text/html")
	h.tabData = setActiveTab(h.tabData, "kontaanaliticki") // Activate the "Promet konta analitički" tab
	err := tmpl_fin.PrometKontaAnaliticki(h.tabData, tbl, btnPrint, btnObrada, domain.PrometTotalValues{}).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (h *PrometHandler) PrometTotalValues(w http.ResponseWriter, r *http.Request) {

	// Get totals data
	response, err := h.service.GetPrometTotals(r)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl_fin.TotalValues(response.Totals).Render(r.Context(), w)

}
func (h *PrometHandler) SearchButtonDialog(w http.ResponseWriter, r *http.Request) {
	hxVals := ""
	vkonta := r.URL.Query().Get("vkonta")
	values := r.URL.Query()
	id := ""
	placeholder := "trazi konto..."
	if _, exists := values["konto"]; exists {
		// Parameter exists (even if empty)
		id = "konto"
		placeholder = "trazi konto..."

	}
	if _, exists := values["sifra"]; exists {
		// Parameter exists (even if empty)
		id = "sifra"
		placeholder = "trazi sifru..."
	}
	w.Header().Set("Content-Type", "text/html")

	if vkonta != "" {
		hxVals = fmt.Sprintf(`{"vkonta": "%s"}`, vkonta)
	}
	tmpl.SearchButtonDialog(id, id, placeholder, "/api/fkpl/trazikontosearchtable", "#search-results", "innerHTML", hxVals).Render(r.Context(), w)

}

func (h *PrometHandler) AddRoutes(r *http.ServeMux) {
	// Define routes for fkpl
	r.HandleFunc("GET /api/promet", infrastructure.AuthMiddleware(h.PrometMain))
	r.HandleFunc("GET /api/promet/analitickakonta", infrastructure.AuthMiddleware(h.PrometAnalitickihKonta))
	r.HandleFunc("GET /api/promet/analitickakontami", infrastructure.AuthMiddleware(h.PrometAnalitickihKontaPoMI))
	r.HandleFunc("GET /api/promet/deviznihanalitickihkonta", infrastructure.AuthMiddleware(h.PrometDeviznihAnalitickihKonta))
	r.HandleFunc("GET /api/promet/subsintetickakonta", infrastructure.AuthMiddleware(h.PrometSubsintetickihKonta))
	r.HandleFunc("GET /api/promet/sintetickakonta", infrastructure.AuthMiddleware(h.PrometSintetickihKonta))
	r.HandleFunc("GET /api/promet/karticasintetickihkonta", infrastructure.AuthMiddleware(h.KarticaSintetickihKonta))
	r.HandleFunc("GET /api/promet/kontaanaliticki", infrastructure.AuthMiddleware(h.PrometKontaAnaliticki))
	r.HandleFunc("GET /api/promet/totalvalues", infrastructure.AuthMiddleware(h.PrometTotalValues))
	r.HandleFunc("GET /api/promet/searchbutton", infrastructure.AuthMiddleware(h.SearchButtonDialog))

}

func GetTabData() domain.TabData {
	return domain.TabData{
		Tabs: []domain.TabItem{
			{ID: "prometankonta", Label: "Promet an. konta", HXRequestUrl: fmt.Sprintf("%sanalitickakonta", prometURLPrefix), IsActive: true, Name: "analitickakonta"},
			{ID: "prometankontami", Label: "Promet an. konta po MI", HXRequestUrl: fmt.Sprintf("%sanalitickakontami", prometURLPrefix), IsActive: false, Name: "analitickakontami"},
			{ID: "deviznihanalitickihkonta", Label: "Promet deviznih an. konta", HXRequestUrl: fmt.Sprintf("%sdeviznihanalitickihkonta", prometURLPrefix), IsActive: false, Name: "deviznihanalitickihkonta"},
			{ID: "subsintetickakonta", Label: "Promet subsintetičkih konta", HXRequestUrl: fmt.Sprintf("%ssubsintetickakonta", prometURLPrefix), IsActive: false, Name: "subsintetickakonta"},
			{ID: "sintetickakonta", Label: "Promet sintetičkih konta", HXRequestUrl: fmt.Sprintf("%ssintetickakonta", prometURLPrefix), IsActive: false, Name: "sintetickakonta"},
			{ID: "karticasintetickihkonta", Label: "Kartica sintetičkih konta", HXRequestUrl: fmt.Sprintf("%skarticasintetickihkonta", prometURLPrefix), IsActive: false, Name: "karticasintetickihkonta"},
			{ID: "kontapovrd", Label: "Promet konta po VRD", HXRequestUrl: fmt.Sprintf("%skontapovrd", prometURLPrefix), IsActive: false, Name: "kontapovrd"},
			{ID: "kontaanaliticki", Label: "Promet konta analitički", HXRequestUrl: fmt.Sprintf("%skontaanaliticki", prometURLPrefix), IsActive: false, Name: "kontaanaliticki"},
		},
	}
}

func setActiveTab(tabs domain.TabData, tabName string) domain.TabData {
	for i, tab := range tabs.Tabs {
		if tab.Name == tabName {
			tabs.Tabs[i].IsActive = true
		} else {
			tabs.Tabs[i].IsActive = false
		}
	}
	return tabs
}
