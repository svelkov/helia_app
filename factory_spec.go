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
	searchPath := "baza"

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable", host, port, user, pwd, dbname, searchPath))
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
	tipdokRepo := repository.NewBaseRepository[domain.Tipdok](db, "tipdok", cfg)
	tipdokValidator := validation.NewRuleBasedValidator[domain.Tipdok](validation.TipdokValidationRules())
	tipdokService := service.NewBaseService(*tipdokRepo, *tipdokValidator)
	tipdokHandler := handler.NewTipdokHandler(tipdokService)
	tipdokHandler.AddRoutes(r)

	// Create repository, validator, service and handler for dokvrsta
	dokvrstaRepo := repository.NewBaseRepository[domain.Dokvrsta](db, "dokvrsta", cfg)
	dokvrstaValidator := validation.NewRuleBasedValidator[domain.Dokvrsta](validation.DokvrstaValidationRules())
	dokvrstaService := service.NewBaseService(*dokvrstaRepo, *dokvrstaValidator)
	dokvrstaHandler := handler.NewDokvrstaHandler(dokvrstaService)
	dokvrstaHandler.AddRoutes(r)

	// Create repository, validator, service and handler for opstine
	opstineRepo := repository.NewBaseRepository[domain.Sifop](db, "sifop", cfg)
	opstineValidator := validation.NewRuleBasedValidator[domain.Sifop](validation.OpstineValidationRules())
	opstineService := service.NewBaseService(*opstineRepo, *opstineValidator)
	opstineHandler := handler.NewOpstineHandler(opstineService)
	opstineHandler.AddRoutes(r)

	// Create repository, validator, service and handler for opstine
	sifmestoRepo := repository.NewBaseRepository[domain.Sifmesto](db, "sifmesto", cfg)
	sifmestoValidator := validation.NewRuleBasedValidator[domain.Sifmesto](validation.SifmestoValidationRules())
	sifmestoService := service.NewBaseService(*sifmestoRepo, *sifmestoValidator)
	sifmestoHandler := handler.NewSifmestoHandler(sifmestoService)
	sifmestoHandler.AddRoutes(r)
	// Create repository, validator, service and handler for opstine
	mestotroskaRepo := repository.NewBaseRepository[domain.Mestotr](db, "mestotr", cfg)
	mestotroskaValidator := validation.NewRuleBasedValidator[domain.Mestotr](validation.MestotroskaValidationRules())
	mestotroskaService := service.NewBaseService(*mestotroskaRepo, *mestotroskaValidator)
	mestoroskaHandler := handler.NewMestotroskaHandler(mestotroskaService)
	mestoroskaHandler.AddRoutes(r)

	// Create repository, validator, service and handler for organizacione jedinice
	orgjedRepo := repository.NewBaseRepository[domain.Orgjed](db, "orgjed", cfg)
	orgjedValidator := validation.NewRuleBasedValidator[domain.Orgjed](validation.OrgjedValidationRules())
	orgjedService := service.NewBaseService(*orgjedRepo, *orgjedValidator)
	orgjedHandler := handler.NewOrgjedHandler(orgjedService)
	orgjedHandler.AddRoutes(r)

	// Create repository, validator, service and handler for banke
	bankeRepo := repository.NewBaseRepository[domain.Banke](db, "banke", cfg)
	bankeValidator := validation.NewRuleBasedValidator[domain.Banke](validation.BankeValidationRules())
	bankeService := service.NewBaseService(*bankeRepo, *bankeValidator)
	bankeHandler := handler.NewBankeHandler(bankeService)
	bankeHandler.AddRoutes(r)

	// Create repository, validator, service and handler for banke
	sifplizvRepo := repository.NewBaseRepository[domain.Sifplizv](db, "sifplizv", cfg)
	sifplizvValidator := validation.NewRuleBasedValidator[domain.Sifplizv](validation.SifplizvValidationRules())
	sifplizvService := service.NewBaseService(*sifplizvRepo, *sifplizvValidator)
	sifplizvHandler := handler.NewSifplizvHandler(sifplizvService)
	sifplizvHandler.AddRoutes(r)

	// Create repository, validator, service and handler for tipove knjige
	fvknjracRepo := repository.NewBaseRepository[domain.Fvknjrac](db, "fvknjrac", cfg)
	fvknjracValidator := validation.NewRuleBasedValidator[domain.Fvknjrac](validation.FvknjracValidationRules())
	fvknjracService := service.NewBaseService(*fvknjracRepo, *fvknjracValidator)
	fvknjracHandler := handler.NewFvknjracHandler(fvknjracService)
	fvknjracHandler.AddRoutes(r)

	// Create repository, validator, service and handler for banke za izvozne fakture
	bnkizvRepo := repository.NewBaseRepository[domain.Bnkizv](db, "bnkizv", cfg)
	bnkizvValidator := validation.NewRuleBasedValidator[domain.Bnkizv](validation.BnkizvValidationRules())
	bnkizvService := service.NewBaseService(*bnkizvRepo, *bnkizvValidator)
	bnkizvHandler := handler.NewBnkizvHandler(bnkizvService)
	bnkizvHandler.AddRoutes(r)

	// Create repository, validator, service and handler for vrste evidencije PDV
	fvepdvRepo := repository.NewBaseRepository[domain.Fvepdv](db, "fvepdv", cfg)
	fvepdvValidator := validation.NewRuleBasedValidator[domain.Fvepdv](validation.FvepdvValidationRules())
	fvepdvService := service.NewBaseService(*fvepdvRepo, *fvepdvValidator)
	fvepdvHandler := handler.NewFvepdvHandler(fvepdvService)
	fvepdvHandler.AddRoutes(r)

	basicHandler := handler.NewBasicHandler(IsLoggedIn, []domain.SubMenuItem{})
	basicHandler.AddRoutes(r)

	// Create repository, validator, service and handler for vrste evidencije PDV
	fkplRepo := repository.NewBaseRepository[domain.Fkpl](db, "fkpl", cfg)
	fkplValidator := validation.NewRuleBasedValidator[domain.Fkpl](finval.FkplValidationRules())
	fkplService := service.NewBaseService(*fkplRepo, *fkplValidator)
	fkplHandler := fin.NewFkplHandler(fkplService)
	fkplHandler.AddRoutes(r)

	// Create repository, validator, service and handler for vrste evidencije PDV
	fnalRepo := repository.NewBaseRepository[domain.Fnal](db, "fnal", cfg)
	fnalValidator := validation.NewRuleBasedValidator[domain.Fnal](finval.FnalValidationRules())
	fnalService := service.NewBaseService(*fnalRepo, *fnalValidator)
	fnalHandler := fin.NewFnalHandler(fnalService, tipdokService)
	fnalHandler.AddRoutes(r)
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
