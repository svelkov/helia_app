package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/jmoiron/sqlx"

	"helia/config"
	tmpl "helia/frontend/templates"
	"helia/internal/domain"
	"helia/internal/handler"
	fin "helia/internal/handler/finansijsko"
	"helia/internal/repository"
	"helia/internal/service"
	"helia/internal/validation"
	finval "helia/internal/validation/finansijsko"
)

func factory() {
	//config should come from config.json, god and kar from combo box.
	cfg := config.Config{}
	cfg.God = 2022
	cfg.Kar = 1
	cfg.PageSize = 20

	// Initialize database connection
	host := "localhost"
	port := 5432
	user := "postgres"
	pwd := "postgres"
	dbname := "helia"
	search_path := "baza"

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable", host, port, user, pwd, dbname, search_path))
	if err != nil {
		log.Fatal(err)
	}
	r := http.NewServeMux()
	//r := mux.NewRouter()
	setEntities(db, r, cfg)

	r.Handle("/frontend/static/", http.StripPrefix("/frontend/static/", http.FileServer(http.Dir("./frontend/static"))))
	// API endpoint
	r.HandleFunc("/get-menu", getMenuHandler)

	// Start server
	http.ListenAndServe(":8080", r)
}

func getMenuHandler(w http.ResponseWriter, r *http.Request) {
	menuName := r.URL.Query().Get("menuName")
	var submenu []domain.SubMenuItem

	// Get submenu items
	if val, exists := domain.MenuData[menuName]; exists {
		submenu = val
	} else {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}

	err := tmpl.Side_nav(submenu).Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func RenderBootstrapSideNav(w http.ResponseWriter, submenu []domain.SubMenuItem) {

	tmpl, err := template.ParseFiles("frontend/templates/bootstrap_side_nav.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, submenu); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func setEntities(db *sqlx.DB, r *http.ServeMux, cfg config.Config) {

	// Create repository, validator, service and handler for drzave
	partnerirepo := repository.NewBaseRepository[domain.Partneri](db, "partneri", cfg)
	partnerivalidator := validation.NewRuleBasedValidator[domain.Partneri](validation.PartneriValidationRules())
	partneriservice := service.NewBaseService(*partnerirepo, *partnerivalidator)
	partnerihandler := handler.NewPartneriHandler(partneriservice)
	partnerihandler.AddRoutes(r)
	// Create repository, validator, service and handler for drzave
	drzavarepo := repository.NewBaseRepository[domain.Drzave](db, "drzave", cfg)
	drzavavalidator := validation.NewRuleBasedValidator[domain.Drzave](validation.DrzavaValidationRules())
	drzavaservice := service.NewBaseService(*drzavarepo, *drzavavalidator)
	drzavahandler := handler.NewDrzavaHandler(drzavaservice)
	drzavahandler.AddRoutes(r)

	// Create repository, validator, service and handler for tipdok
	tipdok_repo := repository.NewBaseRepository[domain.Tipdok](db, "tipdok", cfg)
	tipdok_validator := validation.NewRuleBasedValidator[domain.Tipdok](validation.TipdokValidationRules())
	tipdok_service := service.NewBaseService(*tipdok_repo, *tipdok_validator)
	tipdok_handler := handler.NewTipdokHandler(tipdok_service)
	tipdok_handler.AddRoutes(r)

	// Create repository, validator, service and handler for dokvrsta
	dokvrsta_repo := repository.NewBaseRepository[domain.Dokvrsta](db, "dokvrsta", cfg)
	dokvrsta_validator := validation.NewRuleBasedValidator[domain.Dokvrsta](validation.DokvrstaValidationRules())
	dokvrsta_service := service.NewBaseService(*dokvrsta_repo, *dokvrsta_validator)
	dokvrsta_handler := handler.NewDokvrstaHandler(dokvrsta_service)
	dokvrsta_handler.AddRoutes(r)

	// Create repository, validator, service and handler for opstine
	opstine_repo := repository.NewBaseRepository[domain.Sifop](db, "sifop", cfg)
	opstine_validator := validation.NewRuleBasedValidator[domain.Sifop](validation.OpstineValidationRules())
	opstine_service := service.NewBaseService(*opstine_repo, *opstine_validator)
	opstine_handler := handler.NewOpstineHandler(opstine_service)
	opstine_handler.AddRoutes(r)

	// Create repository, validator, service and handler for opstine
	sifmesto_repo := repository.NewBaseRepository[domain.Sifmesto](db, "sifmesto", cfg)
	sifmesto_validator := validation.NewRuleBasedValidator[domain.Sifmesto](validation.SifmestoValidationRules())
	sifmesto_service := service.NewBaseService(*sifmesto_repo, *sifmesto_validator)
	sifmesto_handler := handler.NewSifmestoHandler(sifmesto_service)
	sifmesto_handler.AddRoutes(r)
	// Create repository, validator, service and handler for opstine
	mestotroska_repo := repository.NewBaseRepository[domain.Mestotr](db, "mestotr", cfg)
	mestotroska_validator := validation.NewRuleBasedValidator[domain.Mestotr](validation.MestotroskaValidationRules())
	mestotroska_service := service.NewBaseService(*mestotroska_repo, *mestotroska_validator)
	mestoroska_handler := handler.NewMestotroskaHandler(mestotroska_service)
	mestoroska_handler.AddRoutes(r)

	// Create repository, validator, service and handler for organizacione jedinice
	orgjed_repo := repository.NewBaseRepository[domain.Orgjed](db, "orgjed", cfg)
	orgjed_validator := validation.NewRuleBasedValidator[domain.Orgjed](validation.OrgjedValidationRules())
	orgjed_service := service.NewBaseService(*orgjed_repo, *orgjed_validator)
	orgjed_handler := handler.NewOrgjedHandler(orgjed_service)
	orgjed_handler.AddRoutes(r)

	// Create repository, validator, service and handler for banke
	banke_repo := repository.NewBaseRepository[domain.Banke](db, "banke", cfg)
	banke_validator := validation.NewRuleBasedValidator[domain.Banke](validation.BankeValidationRules())
	banke_service := service.NewBaseService(*banke_repo, *banke_validator)
	banke_handler := handler.NewBankeHandler(banke_service)
	banke_handler.AddRoutes(r)

	// Create repository, validator, service and handler for banke
	sifplizv_repo := repository.NewBaseRepository[domain.Sifplizv](db, "sifplizv", cfg)
	sifplizv_validator := validation.NewRuleBasedValidator[domain.Sifplizv](validation.SifplizvValidationRules())
	sifplizv_service := service.NewBaseService(*sifplizv_repo, *sifplizv_validator)
	sifplizv_handler := handler.NewSifplizvHandler(sifplizv_service)
	sifplizv_handler.AddRoutes(r)

	// Create repository, validator, service and handler for tipove knjige
	fvknjrac_repo := repository.NewBaseRepository[domain.Fvknjrac](db, "fvknjrac", cfg)
	fvknjrac_validator := validation.NewRuleBasedValidator[domain.Fvknjrac](validation.FvknjracValidationRules())
	fvknjrac_service := service.NewBaseService(*fvknjrac_repo, *fvknjrac_validator)
	fvknjrac_handler := handler.NewFvknjracHandler(fvknjrac_service)
	fvknjrac_handler.AddRoutes(r)

	// Create repository, validator, service and handler for banke za izvozne fakture
	bnkizv_repo := repository.NewBaseRepository[domain.Bnkizv](db, "bnkizv", cfg)
	bnkizv_validator := validation.NewRuleBasedValidator[domain.Bnkizv](validation.BnkizvValidationRules())
	bnkizv_service := service.NewBaseService(*bnkizv_repo, *bnkizv_validator)
	bnkizv_handler := handler.NewBnkizvHandler(bnkizv_service)
	bnkizv_handler.AddRoutes(r)

	// Create repository, validator, service and handler for vrste evidencije PDV
	fvepdv_repo := repository.NewBaseRepository[domain.Fvepdv](db, "fvepdv", cfg)
	fvepdv_validator := validation.NewRuleBasedValidator[domain.Fvepdv](validation.FvepdvValidationRules())
	fvepdv_service := service.NewBaseService(*fvepdv_repo, *fvepdv_validator)
	fvepdv_handler := handler.NewFvepdvHandler(fvepdv_service)
	fvepdv_handler.AddRoutes(r)

	basicHandler := handler.NewBasicHandler(IsLoggedIn, []domain.SubMenuItem{})
	basicHandler.AddRoutes(r)

	// Create repository, validator, service and handler for vrste evidencije PDV
	fkpl_repo := repository.NewBaseRepository[domain.Fkpl](db, "fkpl", cfg)
	fkpl_validator := validation.NewRuleBasedValidator[domain.Fkpl](finval.FkplValidationRules())
	fkpl_service := service.NewBaseService(*fkpl_repo, *fkpl_validator)
	fkpl_handler := fin.NewFkplHandler(fkpl_service)
	fkpl_handler.AddRoutes(r)

	// Create repository, validator, service and handler for vrste evidencije PDV
	fnal_repo := repository.NewBaseRepository[domain.Fnal](db, "fnal", cfg)
	fnal_validator := validation.NewRuleBasedValidator[domain.Fnal](finval.FnalValidationRules())
	fnal_service := service.NewBaseService(*fnal_repo, *fnal_validator)
	fnal_handler := fin.NewFnalHandler(fnal_service, tipdok_service)
	fnal_handler.AddRoutes(r)
}

// In your factory:
func NewBankeHandler(r *http.ServeMux, db *sqlx.DB, cfg config.Config) { // Added r, db, cfg
	// Create repository, validator, service and handler for banke
	bankeRepo := repository.NewBaseRepository[domain.Banke](db, "banke", cfg)
	bankeValidator := validation.NewRuleBasedValidator[domain.Banke](validation.BankeValidationRules())
	bankeService := service.NewBaseService(*bankeRepo, *bankeValidator)
	bankeHandler := handler.NewBankeHandler(bankeService) // Call the handler constructor
	bankeHandler.AddRoutes(r)
}
