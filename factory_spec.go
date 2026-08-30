package main

import (
	"helia/internal/validation/robno"

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
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/jmoiron/sqlx"

	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/handler"
	fin "helia/internal/handler/finansijsko"
	"helia/internal/handler/kamate"
	oshandler "helia/internal/handler/os"
	robnohand "helia/internal/handler/robno"
	"helia/internal/infrastructure/db"
	"helia/internal/middleware"
	"helia/internal/repository"
	"helia/internal/service"
	finservice "helia/internal/service/finansijsko"
	robnosvc "helia/internal/service/robno"
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
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := i18n.Init("./i18n/translations", cfg.Languages, "SR"); err != nil {
		log.Fatal("Failed to load translations:", err)
	}
	translator := i18n.GetInstance()

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Get JWT secret
	jwtSecret := getJwtSecret(cfg)

	// Get Session secret
	sessionSecret := getSessionSecret(cfg)

	// Create router
	router := setupRouter(translator, jwtSecret, sessionSecret)

	// create empty context
	c := gin.CreateTestContextOnly(nil, router)
	lockService := middleware.NewLockService(db)
	lm := middleware.NewLockMiddleware(lockService)
	// Entity settings
	setEntities(c, db, router, jwtSecret, cfg, lm, lockService)

	// Create server
	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        router,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup periodic cleanup (every 30 minutes)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			lockService.CleanupExpired(context.Background())
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

func setupRouter(translator *i18n.Service, jwtSecret []byte, sessionSecret string) *gin.Engine {
	router := gin.Default()

	// Store JWT secret in router engine so handlers can access it
	router.Use(func(c *gin.Context) {
		c.Set("jwtSecret", jwtSecret)
		c.Next()
	})

	// Configure sessions
	store := configureSessionStore(sessionSecret)
	router.Use(sessions.Sessions("heliasession", store))

	// UserSession middleware - creates UserSession in context from JWT token on every request
	router.Use(middleware.UserSession(jwtSecret))
	// Global middleware to inject UserSession into standard context
	router.Use(func(c *gin.Context) {
		userSession := domain.GetSessionFromContext(c)
		if userSession != nil {
			// Wrap the request context with userSession
			ctx := domain.SetSessionInStdContext(c.Request.Context(), userSession)
			// Replace the request's context
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	})
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.I18n(translator))
	router.Use(gzip.Gzip(gzip.DefaultCompression)) // Compression¨
	router.Use(middleware.ContextWithSessionMiddleware())

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
	router.Use(middleware.UserSession(jwtSecret))
	// Add request logging to debug
	// router.Use(func(c *gin.Context) {
	// 	log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
	// 	c.Next()
	// })

	// Load HTML templates
	//router.LoadHTMLGlob("./frontend/templates/*")
	// API routes
	//router.Handle("GET", "/get-menu", getMenuHandler)

	return router

}

func configureSessionStore(sessionSecret string) sessions.Store {
	store := cookie.NewStore([]byte(sessionSecret))

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(tokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: getSameSitePolicy(),
	})

	return store
}

func getSessionSecret(cfg config.Config) string {
	if secret := os.Getenv("SESSION_SECRET"); secret != "" {
		return secret
	}

	if cfg.SessionSecret != "" {
		return cfg.SessionSecret
	}

	if gin.Mode() == gin.ReleaseMode {
		log.Fatal("SESSION_SECRET environment variable or session_secret in config is required in production")
	}

	log.Println("Warning: Using default session secret for development")
	return SESSION_SECRET
}

func getJwtSecret(cfg config.Config) []byte {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}

	if cfg.JwtSecret != "" {
		return []byte(cfg.JwtSecret)
	}

	if gin.Mode() == gin.ReleaseMode {
		log.Fatal("JWT_SECRET environment variable or jwt_secret in config is required in production")
	}

	log.Println("Warning: Using default JWT secret for development")
	return []byte(SESSION_SECRET)
}
func getSameSitePolicy() http.SameSite {
	switch gin.Mode() {
	case gin.ReleaseMode:
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

// parseJWT parses and validates a JWT token, returning the user claims
func parseJWT(tokenString string, jwtSecret []byte) (*domain.UserClaims, error) {
	claims := &domain.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
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
func connectDB(cfg config.Config) (db.Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		cfg.DBConfig.DBHost, cfg.DBConfig.DBPort, cfg.DBConfig.DBUser, cfg.DBConfig.DBPassword, cfg.DBConfig.DBName, cfg.DBConfig.DBSearchPath,
	)
	conn, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	// Wrap sqlx.DB with PostgresDB adapter
	db := db.NewPostgresDB(conn)
	return db, nil
}

// registerGenericEntity registers a generic entity's routes
func registerGenericEntity[T any](
	r *gin.Engine,
	db db.Database,
	tableName string,
	validationRules []validation.ValidationRule,
	fields []domain.Fields,
	config domain.HandlerConfig,
	cfg config.Config,
	lm *middleware.LockMiddleware, // Optional lock middleware for entities that require it
) {
	fvrRepo := repository.NewBaseRepository[domain.Fvr](db, "fvr")
	repo := repository.NewBaseRepository[T](db, tableName)
	validator := validation.NewRuleBasedValidator[T](validationRules)
	svc := service.NewBaseService(*repo, validator)
	h := handler.NewGenericHandler(svc, fields, config, cfg, lm, fvrRepo)
	h.RegisterRoutes(r)
}

// setEntities registers all entity routes
func setEntities(c *gin.Context, db db.Database, r *gin.Engine, jwtSecret []byte, cfg config.Config, lm *middleware.LockMiddleware, ls *middleware.LockService) {
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
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
	)

	// Sifmesto
	registerGenericEntity[domain.Sifmesto](
		r, db, "sifmesto",
		validation.SifmestoValidationRules(),
		handler.SetSifmestoFields(),
		domain.HandlerConfig{
			ContentTitle: "MESTA",
			TableID:      "sifmesto-table",
			APIPrefix:    "/api/sifmesto",
			IDField:      common.IDsifmesto,
		},
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
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
		cfg,
		lm,
	)
	// Tipanalitike
	registerGenericEntity[domain.Tipanalitike](
		r, db, "tipanalitike",
		validation.TipanalitikeValidationRules(),
		handler.SetTipanalitikeFields(),
		domain.HandlerConfig{
			ContentTitle: "TIPOVI ANALITIKE",
			TableID:      "tipanalitike-table",
			APIPrefix:    "/api/tipanalitike",
			IDField:      common.IDtipanalitike,
		},
		cfg,
		lm,
	)
	// ROBNO
	// Rgru
	registerGenericEntity[domain.Rgru](
		r, db, "rgru",
		robno.RgruValidationRules(),
		handler.SetRgruFields(),
		domain.HandlerConfig{
			ContentTitle: "ROBNE GRUPE",
			TableID:      "rgru-table",
			APIPrefix:    "/api/rgru",
			IDField:      common.IDrgru,
		},
		cfg,
		lm,
	)

	// Rpgru
	registerGenericEntity[domain.Rpgru](
		r, db, "rpgru",
		robno.RpgruValidationRules(),
		handler.SetRpgruFields(),
		domain.HandlerConfig{
			ContentTitle: "ROBNE PODGRUPE",
			TableID:      "rpgru-table",
			APIPrefix:    "/api/rpgru",
			IDField:      common.IDrpgru,
		},
		cfg,
		lm,
	)

	// Jedmere
	registerGenericEntity[domain.Jedmere](
		r, db, "jedmere",
		robno.JedmereValidationRules(),
		handler.SetJedmereFields(),
		domain.HandlerConfig{
			ContentTitle: "JEDINICE MERE",
			TableID:      "jedmere-table",
			APIPrefix:    "/api/jedmere",
			IDField:      common.IDjedmere,
		},
		cfg,
		lm,
	)

	// Rpor
	registerGenericEntity[domain.Rpor](
		r, db, "rpor",
		robno.RporValidationRules(),
		handler.SetRporFields(),
		domain.HandlerConfig{
			ContentTitle: "SIFARNIK PORESKIH STOPA",
			TableID:      "rpor-table",
			APIPrefix:    "/api/rpor",
			IDField:      common.IDrpor,
		},
		cfg,
		lm,
	)

	// OSNOVNA SREDSTVA
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
		cfg,
		lm,
	)

	fvrRepo := repository.NewBaseRepository[domain.Fvr](db, "fvr")

	// MAGACINI
	magaciniRepo := repository.NewBaseRepository[domain.Magacini](db, "magacini")
	magaciniValidator := validation.NewRuleBasedValidator[domain.Magacini]([]validation.ValidationRule{})
	magaciniBaseService := service.NewBaseService(*magaciniRepo, magaciniValidator)
	magaciniService := robnosvc.NewMagaciniResource(magaciniBaseService, magaciniRepo, fvrRepo, cfg)
	magaciniHandler := robnohand.NewMagaciniHandler(magaciniBaseService, magaciniService, cfg, lm)
	magaciniHandler.AddRoutes(r)

	// MAGACIN KONTO
	magacinKontoRepo := repository.NewBaseRepository[domain.Magkonto](db, "magkonto")
	magacinKontoValidator := validation.NewRuleBasedValidator[domain.Magkonto]([]validation.ValidationRule{})
	magacinKontoBaseService := service.NewBaseService(*magacinKontoRepo, magacinKontoValidator)
	magacinKontoService := robnosvc.NewMagacinKontoResource(magacinKontoBaseService, magacinKontoRepo, fvrRepo, cfg)
	magacinKontoHandler := robnohand.NewMagacinKontoHandler(magacinKontoBaseService, magacinKontoService, cfg, lm)
	magacinKontoHandler.AddRoutes(r)

	// KOMERCIJALISTI
	komercijalistiRepo := repository.NewBaseRepository[domain.Komercijalisti](db, "komercijalisti")
	komercijalistiValidator := validation.NewRuleBasedValidator[domain.Komercijalisti]([]validation.ValidationRule{})
	komercijalistiBaseService := service.NewBaseService(*komercijalistiRepo, komercijalistiValidator)
	komercijalistiService := robnosvc.NewKomercijalistiResource(komercijalistiBaseService, komercijalistiRepo, fvrRepo, cfg)
	komercijalistiHandler := robnohand.NewKomercijalistiHandler(komercijalistiBaseService, komercijalistiService, cfg, lm)
	komercijalistiHandler.AddRoutes(r)

	// ARTIKLI
	artikliRepo := repository.NewBaseRepository[domain.Rsif](db, "rsif")
	artikliValidator := validation.NewRuleBasedValidator[domain.Rsif]([]validation.ValidationRule{})
	artikliBaseService := service.NewBaseService(*artikliRepo, artikliValidator)
	artikliService := robnosvc.NewArtikliResource(artikliBaseService, artikliRepo, fvrRepo, cfg)
	artikliHandler := robnohand.NewArtikliHandler(artikliBaseService, artikliService, cfg, lm)
	artikliHandler.AddRoutes(r)

	// Complex entities with custom services (non-generic)
	//partneri
	tekracuniRepo := repository.NewBaseRepository[domain.TekRacuni](db, "tekracuni")
	partneriRepo := repository.NewBaseRepository[domain.Partneri](db, "partneri")
	tipAnalitikeRepo := repository.NewBaseRepository[domain.Tipanalitike](db, "tipanalitike")
	partneriValidator := validation.NewRuleBasedValidator[domain.Partneri](validation.PartneriValidationRules())
	partnerBaseService := service.NewBaseService(*partneriRepo, partneriValidator)
	partneriService := service.NewPartneriService(partnerBaseService, partneriValidator, partneriRepo, tekracuniRepo, tipAnalitikeRepo, fvrRepo)
	partneriHandler := handler.NewPartneriHandler(partneriService, cfg, lm)
	partneriHandler.AddRoutes(r)
	// Fkpl
	fkplRepo := repository.NewBaseRepository[domain.Fkpl](db, "fkpl")
	fkplValidator := validation.NewRuleBasedValidator[domain.Fkpl](finval.FkplValidationRules())
	baseService := service.NewBaseService(*fkplRepo, fkplValidator)
	fkplService := finservice.NewFkplResource(baseService, fkplRepo, fvrRepo, tipAnalitikeRepo, partneriRepo, cfg)
	fkplHandler := fin.NewFkplHandler(baseService, fkplService, cfg, lm)
	fkplHandler.AddRoutes(r)

	// Fpro
	fproRepo := repository.NewBaseRepository[domain.Fpro](db, "fpro")
	fproFnalRepo := repository.NewBaseRepository[domain.Fnal](db, "fnal")
	fproFkplRepo := repository.NewBaseRepository[domain.Fkpl](db, "fkpl")
	fproOrgjedRepo := repository.NewBaseRepository[domain.Orgjed](db, "orgjed")
	fproMestotrRepo := repository.NewBaseRepository[domain.Mestotr](db, "mestotr")
	fproFvrRepo := repository.NewBaseRepository[domain.Fvr](db, "fvr")
	fproFvknjracRepo := repository.NewBaseRepository[domain.Fvknjrac](db, "fvknjrac")
	fproValuteRepo := repository.NewBaseRepository[domain.Valute](db, "valute")
	fproKomercijalistiRepo := repository.NewBaseRepository[domain.Komercijalisti](db, "komercijalisti")
	fproMiRepo := repository.NewBaseRepository[domain.Fisp](db, "fisp")
	fproMagaciniRepo := repository.NewBaseRepository[domain.Magacini](db, "magacini")
	fproValidator := validation.NewRuleBasedValidator[domain.Fpro](finval.FnalValidationRules())
	fproBaseService := service.NewBaseService(*fproRepo, fproValidator)
	fproService := finservice.NewFproService(
		fproBaseService,
		*fproRepo,
		*fproFnalRepo,
		*fproFkplRepo,
		*fproOrgjedRepo,
		*fproMestotrRepo,
		*fproFvrRepo,
		*fproFvknjracRepo,
		*fproValuteRepo,
		*fproKomercijalistiRepo,
		*fproMiRepo,
		*fproMagaciniRepo,
		common.IDfpro,
		[]domain.Fields{},
		[]domain.Fields{},
		cfg,
	)
	fproHandler := fin.NewFproHandler(fproService, cfg, lm)
	fproHandler.AddRoutes(r)
	// Fnal
	fnalRepo := repository.NewBaseRepository[domain.Fnal](db, "fnal")
	fnalValidator := finval.NewFnalValidator()
	fnalBaseService := service.NewBaseService(*fnalRepo, fnalValidator)
	fnalService := finservice.NewNalogService(
		fnalBaseService,
		fproService,
		fnalValidator,
		*fnalRepo,
		*repository.NewBaseRepository[domain.Tipdok](db, "tipdok"),
		*repository.NewBaseRepository[domain.Sf](db, "sf"),
		*repository.NewBaseRepository[domain.Orgjed](db, "orgjed"),
		*repository.NewBaseRepository[domain.Mestotr](db, "mestotr"),
		repository.NewBaseRepository[domain.Fvr](db, "fvr"),
		*fproRepo,
		"",
		cfg,
	)
	fnalHandler := fin.NewFnalHandler(fnalService, fnalBaseService, cfg, lm, ls)
	fnalHandler.AddRoutes(r)

	// Promet
	prometRepo := repository.NewBaseRepository[domain.PrometDto](db, "prometdto")
	prometValidator := validation.NewRuleBasedValidator[domain.PrometDto](finval.PrometValidationRules())
	prometService := finservice.NewPrometService(
		service.NewBaseService(*prometRepo, prometValidator),
		prometRepo,
		fkplRepo,
		fvrRepo,
	)
	prometHandler := fin.NewPrometHandler(prometService, cfg)
	prometHandler.AddRoutes(r)

	// Robno kartica artikla
	robnoKarticaRepo := repository.NewBaseRepository[domain.RobnoStanjeDto](db, "robnostanjedto")
	robnoKarticaService := robnosvc.NewRobnoKarticaService(robnoKarticaRepo)
	robnoKarticaHandler := robnohand.NewRobnoKarticaHandler(robnoKarticaService, cfg)
	robnoKarticaHandler.AddRoutes(r)

	// Robno stanja
	robnoStanjaRepo := repository.NewBaseRepository[domain.RobnoStanjeDto](db, "robnostanjedto")
	robnoStanjaService := robnosvc.NewRobnoStanjaService(*robnoStanjaRepo)
	robnoStanjaHandler := robnohand.NewRobnoStanjaHandler(robnoStanjaService, cfg)
	robnoStanjaHandler.AddRoutes(r)

	// Robno promet reports
	robnoprometService := robnosvc.NewRobnoPrometService(prometService)
	robnoprometHandler := robnohand.NewRobnoPrometHandler(robnoprometService, cfg)
	robnoprometHandler.AddRoutes(r)

	// Salda
	saldaRepo := repository.NewBaseRepository[domain.SaldaDto](db, "saldadto")
	saldaValidator := validation.NewRuleBasedValidator[domain.SaldaDto](finval.SaldaValidationRules())
	saldaService := finservice.NewSaldaService(
		service.NewBaseService(*saldaRepo, saldaValidator),
		saldaRepo,
		fkplRepo,
		fproRepo,
		repository.NewBaseRepository[domain.Partneri](db, "partneri"),
		repository.NewBaseRepository[domain.SaldaPartnerDto](db, "saldapartneridto"),
		repository.NewBaseRepository[domain.SaldaKomercijalistiDto](db, "saldakomercijalistidto"),
		repository.NewBaseRepository[domain.Fvr](db, "fvr"),
	)

	saldaHandler := fin.NewSaldaHandler(saldaService, cfg)
	saldaHandler.AddRoutes(r)

	// SaKompenzacije
	kompRepo := repository.NewBaseRepository[domain.KompenzacijeDto](db, "kompenzacijedto")
	kompValidator := validation.NewRuleBasedValidator[domain.KompenzacijeDto](finval.KompenzacijeValidationRules())
	kompService := finservice.NewKompenzacijeService(
		service.NewBaseService(*kompRepo, kompValidator),
		fproRepo,
		// fproRepo,
		// repository.NewBaseRepository[domain.Partneri](db, "partneri"),
		// repository.NewBaseRepository[domain.SaldaPartnerDto](db, "saldapartneridto"),
		// repository.NewBaseRepository[domain.SaldaKomercijalistiDto](db, "saldakomercijalistidto"),
	)

	kompHandler := fin.NewKompenzacijeHandler(kompService, cfg, lm)
	kompHandler.AddRoutes(r)

	// DnevnikHandler

	dnevnikRepo := repository.NewBaseRepository[domain.DnevnikDto](db, "fpro")
	dnevnikValidator := validation.NewValidator[domain.DnevnikDto]()
	dnevnikService := finservice.NewDnevnikService(
		service.NewBaseService(*dnevnikRepo, dnevnikValidator),
		fproRepo,
		fvrRepo,
	)
	dnevnikHandler := fin.NewDnevnikHandler(dnevnikService, cfg)
	dnevnikHandler.AddRoutes(r)

	// Izvodi
	izvhdrRepo := repository.NewBaseRepository[domain.Fizvzag](db, "fizvzag")
	izvdetRepo := repository.NewBaseRepository[domain.Fizvdet](db, "fizvdet")
	bankeRepo := repository.NewBaseRepository[domain.Banke](db, "banke")
	tipdokRepo := repository.NewBaseRepository[domain.Tipdok](db, "tipdok")
	sifplizvRepo := repository.NewBaseRepository[domain.Sifplizv](db, "sifplizv")
	izvodiService := finservice.NewIzvodiResource(izvhdrRepo, izvdetRepo, bankeRepo, tipdokRepo, fnalRepo, partneriRepo, tekracuniRepo, sifplizvRepo, fkplRepo, fvrRepo, cfg)
	izvodiHandler := fin.NewIzvodiHandler(izvodiService, cfg, lm)
	izvodiHandler.AddRoutes(r)

	// BasicHandler
	fvrService := finservice.NewFvrService(fvrRepo)
	menuService := service.NewMenuService(
		repository.NewBaseRepository[domain.MenuItem](db, "menuitems"),
		repository.NewBaseRepository[domain.SubMenuItem](db, "submenuitems"),
	)

	// UserService for authentication and user management
	userRepo := repository.NewBaseRepository[domain.User](db, "appusers")
	userValidator := validation.NewRuleBasedValidator[domain.User]([]validation.ValidationRule{})
	userService := service.NewUserService(userRepo, userValidator)

	basicHandler := handler.NewBasicHandler(
		c,
		menuService,
		IsLoggedIn,
		fvrService,
		cfg,
		userService,
	)
	basicHandler.AddRoutes(r)

	// PoreskeKnjige
	kirRepo := repository.NewBaseRepository[domain.Kir](db, "kir")
	kprRepo := repository.NewBaseRepository[domain.Kpr](db, "kpr")
	fvknjracRepo := repository.NewBaseRepository[domain.Fvknjrac](db, "fvknjrac")
	kirValidator := validation.NewRuleBasedValidator[domain.Kir]([]validation.ValidationRule{})
	kprValidator := validation.NewRuleBasedValidator[domain.Kpr]([]validation.ValidationRule{})
	kirService := service.NewBaseService(*kirRepo, kirValidator)
	kprService := service.NewBaseService(*kprRepo, kprValidator)
	poreskeKnjigeService := finservice.NewPoreskeKnjigeService(kirService, kprService, kirRepo, kprRepo, fvknjracRepo, tipdokRepo, fvrRepo)
	poreskeKnjigeHandler := fin.NewPoreskeKnjigeHandler(poreskeKnjigeService, cfg, lm)
	poreskeKnjigeHandler.RegisterRoutes(r)

	//otvorene stavke
	otvFvrRepo := repository.NewBaseRepository[domain.Fvr](db, "fvr")
	otvPartneriRepo := repository.NewBaseRepository[domain.Partneri](db, "partneri")
	otvoreneStavkeService := finservice.NewOtvoreneStavkeService(*fproRepo, otvFvrRepo, otvPartneriRepo)
	otvoreneStavkeHandler := fin.NewOtvoreneStavkeHandler(otvoreneStavkeService, cfg)
	otvoreneStavkeHandler.RegisterRoutes(r)

	// Kamatne stope (Kam)
	kamRepo := repository.NewBaseRepository[domain.Kam](db, "kam")
	// Tipovi kamate (Tkam)
	tkamRepo := repository.NewBaseRepository[domain.Tkam](db, "tkam")

	kamateService := finservice.NewKamateService(fkplRepo, kamRepo, tkamRepo, fproRepo)
	kamateHandler := kamate.NewKamateHandler(kamateService, cfg, lm)
	kamateHandler.RegisterRoutes(r)

	// Bilansi
	biluRepo := repository.NewBaseRepository[domain.Bilu](db, "bilu")
	biluValidator := validation.NewRuleBasedValidator[domain.Bilu]([]validation.ValidationRule{})
	biluService := service.NewBaseService(*biluRepo, biluValidator)
	bilsRepo := repository.NewBaseRepository[domain.Bils](db, "bils")
	bilsValidator := validation.NewRuleBasedValidator[domain.Bils]([]validation.ValidationRule{})
	bilsService := service.NewBaseService(*bilsRepo, bilsValidator)

	bilansiService := finservice.NewBilansiService(
		biluService,
		biluRepo,
		bilsService,
		bilsRepo,
		repository.NewBaseRepository[domain.FproDto](db, "fprodto"),
		repository.NewBaseRepository[domain.Fvr](db, "fvr"),
		repository.NewBaseRepository[domain.Fkpl](db, "fkpl"),
		cfg,
	)
	bilansiHandler := fin.NewBilansiHandler(bilansiService, cfg, lm)
	bilansiHandler.RegisterRoutes(r)

	// FSEPP
	fseppRepo := repository.NewBaseRepository[domain.Fsepp](db, "fsepp")
	fseppValidator := validation.NewRuleBasedValidator[domain.Fsepp]([]validation.ValidationRule{})
	fseppBaseService := service.NewBaseService(*fseppRepo, fseppValidator)

	fseppSefKprRepo := repository.NewBaseRepository[domain.FseppSefKpr](db, "epp_sef_kpr")
	fseppSefKprValidator := validation.NewRuleBasedValidator[domain.FseppSefKpr]([]validation.ValidationRule{})
	fseppBaseSefKprService := service.NewBaseService(*fseppSefKprRepo, fseppSefKprValidator)

	fseppService := finservice.NewFseppService(
		fseppBaseService,
		fseppBaseSefKprService,
		fseppRepo,
		fseppSefKprRepo,
		kprRepo,
	)
	fseppHandler := fin.NewFseppHandler(fseppService, cfg, lm)
	fseppHandler.RegisterRoutes(r)
	// POPDV
	popdvRepo := repository.NewBaseRepository[domain.Popdv](db, "popdv")
	popdvValidator := validation.NewRuleBasedValidator[domain.Popdv]([]validation.ValidationRule{})
	popdvBaseService := service.NewBaseService(*popdvRepo, popdvValidator)

	popdvService := finservice.NewPopdvService(
		popdvBaseService,
		popdvRepo,
	)
	popdvHandler := fin.NewPopdvHandler(popdvService, cfg, lm)
	popdvHandler.RegisterRoutes(r)
}
