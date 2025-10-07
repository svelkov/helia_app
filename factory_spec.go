package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/jmoiron/sqlx"

	"helia/frontend/templates"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/handler"
	fin "helia/internal/handler/finansijsko"
	oshandler "helia/internal/handler/os"
	"helia/internal/repository"
	"helia/internal/service"
	"helia/internal/validation"
	finval "helia/internal/validation/finansijsko"
	osval "helia/internal/validation/os"
	"helia/pkg/utils"
)

// factory initializes and starts the application
func factory() {
	cfg := loadConfig()
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := http.NewServeMux()
	registerRoutes(r, db)

	// srv := &http.Server{
	// 	Addr:    ":8080",
	// 	Handler: r,
	// }

	// Start server
	http.ListenAndServe(":8080", r)
	// 	// Graceful shutdown
	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("Server failed: %v", err)
	// 	}
	// }()

}

// loadConfig loads configuration from environment or config file
func loadConfig() domain.AppConfig {
	// Read the file
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return domain.AppConfig{}
	}

	// Parse JSON into struct
	var config domain.AppConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return domain.AppConfig{}
	}
	return config
}

// connectDB establishes a database connection
func connectDB(cfg domain.AppConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		cfg.DBConfig.DBHost, cfg.DBConfig.DBPort, cfg.DBConfig.DBUser, cfg.DBConfig.DBPassword, cfg.DBConfig.DBName, cfg.DBConfig.DBSearchPath,
	)
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	return db, nil
}

// registerRoutes registers all routes
func registerRoutes(r *http.ServeMux, db *sqlx.DB) {
	// Static files
	fs := http.FileServer(http.Dir("./frontend/static"))
	r.Handle("/frontend/static/", http.StripPrefix("/frontend/static/", fs))

	// API routes
	r.HandleFunc("/get-menu", getMenuHandler)

	// Entity routes
	setEntities(db, r)
}

// getMenuHandler handles menu requests
func getMenuHandler(w http.ResponseWriter, r *http.Request) {
	menuName := r.URL.Query().Get("menuName")
	subMenus := common.GetSubMenus(domain.MenuData, menuName)
	if subMenus == nil {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}
	if err := templates.Side_nav(subMenus).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// registerGenericEntity registers a generic entity's routes
func registerGenericEntity[T any](
	r *http.ServeMux,
	db *sqlx.DB,
	tableName string,
	validationRules []validation.ValidationRule,
	fields []domain.Fields,
	config domain.HandlerConfig,
) {
	repo := repository.NewBaseRepository[T](db, tableName)
	validator := validation.NewRuleBasedValidator[T](validationRules)
	svc := service.NewBaseService(*repo, *validator)
	h := handler.NewGenericHandler(svc, fields, config)
	h.RegisterRoutes(r)
}

// setEntities registers all entity routes
func setEntities(db *sqlx.DB, r *http.ServeMux) {
	// Partneri
	registerGenericEntity[domain.Partneri](
		r, db, "partneri",
		validation.PartneriValidationRules(),
		handler.SetPartneriFields(),
		domain.HandlerConfig{
			ContentTitle: "PARTNERI",
			TableID:      "partneri-table",
			APIPrefix:    "/api/partneri",
			IDField:      utils.IDpartneri,
		},
	)

	// Drzave
	registerGenericEntity[domain.Drzave](
		r, db, "drzave",
		validation.DrzavaValidationRules(),
		handler.SetDrzaveFields(),
		domain.HandlerConfig{
			ContentTitle: "DRZAVE",
			TableID:      "drzave-table",
			APIPrefix:    "/api/drzave",
			IDField:      utils.IDdrzave,
		},
	)

	// Tipdok
	registerGenericEntity[domain.Tipdok](
		r, db, "tipdok",
		validation.TipdokValidationRules(),
		handler.SetTipdokFields(),
		domain.HandlerConfig{
			ContentTitle: "VRSTE NALOGA",
			TableID:      "tipdok-table",
			APIPrefix:    "/api/tipdok",
			IDField:      utils.IDtipdok,
		},
	)

	// Dokvrsta
	registerGenericEntity[domain.Dokvrsta](
		r, db, "dokvrsta",
		validation.DokvrstaValidationRules(),
		handler.SetDokvrstaFields(),
		domain.HandlerConfig{
			ContentTitle: "VRSTE DOKUMENATA",
			TableID:      "dokvrsta-table",
			APIPrefix:    "/api/dokvrsta",
			IDField:      utils.IDdokvrsta,
		},
	)

	// Opstine
	registerGenericEntity[domain.Sifop](
		r, db, "sifop",
		validation.OpstineValidationRules(),
		handler.SetSifopFields(),
		domain.HandlerConfig{
			ContentTitle: "OPSTINE",
			TableID:      "opstine-table",
			APIPrefix:    "/api/opstine",
			IDField:      utils.IDsifop,
		},
	)

	// Sifmesto
	registerGenericEntity[domain.Sifmesto](
		r, db, "sifmesto",
		validation.SifmestoValidationRules(),
		handler.SetSifmestoFileds(),
		domain.HandlerConfig{
			ContentTitle: "MESTA",
			TableID:      "sifmesto-table",
			APIPrefix:    "/api/sifmesto",
			IDField:      utils.IDsifmesto,
		},
	)

	// Mestotr
	registerGenericEntity[domain.Mestotr](
		r, db, "mestotr",
		validation.MestotroskaValidationRules(),
		handler.SetMestotroskaFields(),
		domain.HandlerConfig{
			ContentTitle: "MESTA TROSKA",
			TableID:      "mestotr-table",
			APIPrefix:    "/api/mestotr",
			IDField:      utils.IDmestotr,
		},
	)

	// Orgjed
	registerGenericEntity[domain.Orgjed](
		r, db, "orgjed",
		validation.OrgjedValidationRules(),
		handler.SetOrgjedFields(),
		domain.HandlerConfig{
			ContentTitle: "ORGANIZACIONE JEDINICE",
			TableID:      "orgjed-table",
			APIPrefix:    "/api/orgjed",
			IDField:      utils.IDorgjed,
		},
	)

	// Banke
	registerGenericEntity[domain.Banke](
		r, db, "banke",
		validation.BankeValidationRules(),
		handler.SetBankeFields(),
		domain.HandlerConfig{
			ContentTitle: "BANKE",
			TableID:      "banke-table",
			APIPrefix:    "/api/banke",
			IDField:      utils.IDbanke,
		},
	)

	// Sifplizv
	registerGenericEntity[domain.Sifplizv](
		r, db, "sifplizv",
		validation.SifplizvValidationRules(),
		handler.SetSifplizvFields(),
		domain.HandlerConfig{
			ContentTitle: "SIFRE PLACANJA",
			TableID:      "sifplizv-table",
			APIPrefix:    "/api/sifplizv",
			IDField:      utils.IDsifplizv,
		},
	)

	// Fvknjrac
	registerGenericEntity[domain.Fvknjrac](
		r, db, "fvknjrac",
		validation.FvknjracValidationRules(),
		handler.SetFvknjracFields(),
		domain.HandlerConfig{
			ContentTitle: "FINANSIJSKI RACUNI",
			TableID:      "fvknjrac-table",
			APIPrefix:    "/api/fvknjrac",
			IDField:      utils.IDfvknjrac,
		},
	)

	// Bnkizv
	registerGenericEntity[domain.Bnkizv](
		r, db, "bnkizv",
		validation.BnkizvValidationRules(),
		handler.SetBnkizvFields(),
		domain.HandlerConfig{
			ContentTitle: "BANKOVNI IZVODI",
			TableID:      "bnkizv-table",
			APIPrefix:    "/api/bnkizv",
			IDField:      utils.IDbnkizv,
		},
	)

	// Fvepdv
	registerGenericEntity[domain.Fvepdv](
		r, db, "fvepdv",
		validation.FvepdvValidationRules(),
		handler.SetFevpdvFields(),
		domain.HandlerConfig{
			ContentTitle: "EVIDENCIJA PDV",
			TableID:      "fvepdv-table",
			APIPrefix:    "/api/fvepdv",
			IDField:      utils.IDfvepdv,
		},
	)

	// Fkpl
	registerGenericEntity[domain.Fkpl](
		r, db, "fkpl",
		finval.FkplValidationRules(),
		fin.SetFkplFields(),
		domain.HandlerConfig{
			ContentTitle: "PLAN KNJIZENJA",
			TableID:      "fkpl-table",
			APIPrefix:    "/api/fkpl",
			IDField:      utils.IDfkpl,
		},
	)

	// Oamgrp
	registerGenericEntity[domain.Oamgrp](
		r, db, "oamgrp",
		osval.OamgrpValidationRules(),
		oshandler.SetOamgrpFields(),
		domain.HandlerConfig{
			ContentTitle: "GRUPE OSNOVNIH SREDSTAVA",
			TableID:      "oamgrp-table",
			APIPrefix:    "/api/oamgrp",
			IDField:      utils.IDoamgrp,
		},
	)

	// Complex entities with custom services (non-generic)

	// Fnal
	fnalRepo := repository.NewBaseRepository[domain.Fnal](db, "fnal")
	fnalValidator := validation.NewRuleBasedValidator[domain.Fnal](finval.FnalValidationRules())
	fnalBaseService := service.NewBaseService(*fnalRepo, *fnalValidator)
	fnalService := service.NewNalogService(
		fnalBaseService,
		*fnalRepo,
		*repository.NewBaseRepository[domain.Tipdok](db, "tipdok"),
		*repository.NewBaseRepository[domain.Sf](db, "sf"),
		"",
		[]domain.Fields{},
		[]domain.Fields{},
	)
	fnalHandler := fin.NewFnalHandler(fnalService)
	fnalHandler.AddRoutes(r)

	// Fpro
	fproRepo := repository.NewBaseRepository[domain.Fpro](db, "fpro")
	fproValidator := validation.NewRuleBasedValidator[domain.Fpro](finval.FnalValidationRules())
	fproBaseService := service.NewBaseService(*fproRepo, *fproValidator)
	fproService := service.NewFproService(
		fproBaseService,
		*fproRepo,
		utils.IDfpro,
		[]domain.Fields{},
		[]domain.Fields{},
	)
	fproHandler := fin.NewFproHandler(fproService)
	fproHandler.AddRoutes(r)

	// Promet
	prometRepo := repository.NewBaseRepository[domain.PrometDto](db, "prometdto")
	prometValidator := validation.NewRuleBasedValidator[domain.PrometDto](finval.PrometValidationRules())
	prometService := service.NewPrometService(
		service.NewBaseService(*prometRepo, *prometValidator),
		prometRepo,
	)
	prometHandler := fin.NewPrometHandler(prometService)
	prometHandler.AddRoutes(r)

	// BasicHandler
	fvrRepo := repository.NewBaseRepository[domain.Fvr](db, "fvr")
	fvrService := service.NewFvrService(fvrRepo)
	basicHandler := handler.NewBasicHandler(
		IsLoggedIn,
		domain.MenuData,
		[]domain.SubMenuItem{},
		fvrService,
	)
	basicHandler.AddRoutes(r)
}
