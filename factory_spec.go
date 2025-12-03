package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/jmoiron/sqlx"

	"helia/config"
	"helia/frontend/templates"
	"helia/global"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/handler"
	fin "helia/internal/handler/finansijsko"
	oshandler "helia/internal/handler/os"
	"helia/internal/i18n"
	"helia/internal/middleware"
	"helia/internal/repository"
	"helia/internal/service"
	"helia/internal/validation"
	finval "helia/internal/validation/finansijsko"
	osval "helia/internal/validation/os"

	"github.com/gin-gonic/gin"
)

const (
	contextTimeout = 30 * time.Second
	tokenExpiry    = 24 * time.Hour
)

// Configuration loaded from environment
var (
	// Secret key for signing the JWT (keep this secure!)
	SESSION_SECRET = "3285f0d71eed0c41fded2115c9cc8ac09a0ab5a519565df10afdb20f8013c5268f2c19948b6af096c1cfc2921ab086be21fa5407b9d91aeb08eeeeef3c2e16c9ae30ae15f27d340f17c450468fef50795e58bb7351a94602bc045aea1a1ff3b03039081208cf067b44fd913b98b712e34ba080941f5ff8545b0eac26824f0ef4a93109939d8f917e1fac1eb588f4272ebac415975bcdc994c3a0fea7c3805d601443ad71dd9043858de5c2bfe64106683d9eaebce28442ce7bb22298d5b85cc3cc41e6f81f9c0f8f678cce559f745645edc5a5009ba20f8b5a16be4ee7dada7791913c90e3629a44b88a17d3d107bd3a6c0f3000b4865b2c015c0875901a028e" // Replace with your secret key
)

// factory initializes and starts the application
func factory() {
	cfg := loadConfig()
	global.SetConfig(cfg)
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := i18n.Init("./translations", cfg.Languages, "sr"); err != nil {
		log.Fatal("Failed to load translations:", err)
	}
	translator := i18n.GetInstance()
	// Initialize translator
	// translator := i18n.NewTranslator("sr")
	// if err := translator.LoadTranslations("./translations", cfg.Languages); err != nil {
	// 	log.Fatal("Failed to load translations:", err)
	// }

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := setupRouter(translator)

	// Entity settings
	setEntities(db, router)

	// Create server
	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(translator *i18n.Service) *gin.Engine {
	router := gin.Default()
	// Configure sessions
	store := configureSessionStore()
	router.Use(sessions.Sessions("heliasession", store))
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.I18n(translator))
	router.Use(gzip.Gzip(gzip.DefaultCompression)) // Compression¨

	// Static files BEFORE CSRF middleware (so they bypass it)
	router.Static("/css", "./frontend/static/css")
	router.Static("/js", "./frontend/static/js")
	router.Static("/frontend/static", "./frontend/static")

	// Cache control for static files
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/frontend/static/") ||
			strings.HasPrefix(c.Request.URL.Path, "/css/") ||
			strings.HasPrefix(c.Request.URL.Path, "/js/") {
			c.Header("Cache-Control", "public, max-age=3600")
		}
		c.Next()
	})

	// CSRF middleware AFTER static files
	router.Use(middleware.CSRFMiddleware()) // Apply to all routes except static
	// Add request logging to debug
	// router.Use(func(c *gin.Context) {
	// 	log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
	// 	c.Next()
	// })

	// Load HTML templates
	//router.LoadHTMLGlob("./frontend/templates/*")
	// API routes
	router.Handle("GET", "/get-menu", getMenuHandler)

	return router

}

func configureSessionStore() sessions.Store {
	secret := getSessionSecret()

	store := cookie.NewStore([]byte(secret))

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(tokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: getSameSitePolicy(),
	})

	return store
}

func getSessionSecret() string {
	if secret := os.Getenv("SESSION_SECRET"); secret != "" {
		return secret
	}

	if gin.Mode() == gin.ReleaseMode {
		log.Fatal("SESSION_SECRET environment variable is required in production")
	}

	log.Println("Warning: Using default session secret for development")
	return SESSION_SECRET
}
func getSameSitePolicy() http.SameSite {
	switch gin.Mode() {
	case gin.ReleaseMode:
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

// loadConfig loads configuration from environment or config file
func loadConfig() config.Config {
	// Read the file
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return config.Config{}
	}

	// Parse JSON into struct
	var config config.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return config
	}
	return config
}

// connectDB establishes a database connection
func connectDB(cfg config.Config) (*sqlx.DB, error) {
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

// getMenuHandler handles menu requests
func getMenuHandler(c *gin.Context) {
	menuName := c.Query("menuName")
	subMenus := common.GetTranslatedSubMenus(domain.MenuData, menuName, global.GetLanguage())
	if subMenus == nil {
		c.JSON(404, gin.H{"error": "Menu not found"})
		return
	}

	// Render templ component directly
	component := templates.Side_nav(subMenus)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.JSON(500, gin.H{"error": "Error rendering template"})
		return
	}
}

// registerGenericEntity registers a generic entity's routes
func registerGenericEntity[T any](
	r *gin.Engine,
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
func setEntities(db *sqlx.DB, r *gin.Engine) {
	// Partneri
	registerGenericEntity[domain.Partneri](
		r, db, "partneri",
		validation.PartneriValidationRules(),
		handler.SetPartneriFields(),
		domain.HandlerConfig{
			ContentTitle: "PARTNERI",
			TableID:      "partneri-table",
			APIPrefix:    "/api/partneri",
			IDField:      common.IDpartneri,
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
			APIPrefix:    "/api/drzava",
			IDField:      common.IDdrzave,
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
			IDField:      common.IDtipdok,
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
			IDField:      common.IDdokvrsta,
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
			APIPrefix:    "/api/sifop",
			IDField:      common.IDsifop,
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
			IDField:      common.IDsifmesto,
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
			APIPrefix:    "/api/mestotroska",
			IDField:      common.IDmestotr,
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
			IDField:      common.IDorgjed,
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
			IDField:      common.IDbanke,
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
			APIPrefix:    "/api/sifplizvodi",
			IDField:      common.IDsifplizv,
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
			IDField:      common.IDfvknjrac,
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
			IDField:      common.IDbnkizv,
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
			IDField:      common.IDfvepdv,
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
			IDField:      common.IDoamgrp,
		},
	)

	// Complex entities with custom services (non-generic)
	// Fkpl
	fkplRepo := repository.NewBaseRepository[domain.Fkpl](db, "fkpl")
	fkplValidator := validation.NewRuleBasedValidator[domain.Fkpl](finval.FkplValidationRules())
	fkplService := service.NewBaseService(*fkplRepo, *fkplValidator)
	fkplHandler := fin.NewFkplHandler(fkplService)
	fkplHandler.AddRoutes(r)

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
		fnalValidator,
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
		common.IDfpro,
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
