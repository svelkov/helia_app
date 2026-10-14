package main

import (
	"testing"

	"helia/i18n"
	"helia/internal/infrastructure/db"
	"helia/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestStartupWiring runs the wiring of the app (setupRouter and setEntities, as factory does) without
// a database: NewBaseRepository panics at startup for an invalid table name or db tag, so a panic here
// is a broken repository. The database is never reached (sqlx.Open does not connect).
func TestStartupWiring(t *testing.T) {
	cfg := loadConfig()
	if err := i18n.Init("./i18n/translations", cfg.Languages, "SR"); err != nil {
		t.Fatal(err)
	}
	conn, err := sqlx.Open("postgres", "host=127.0.0.1 port=1 user=x dbname=x sslmode=disable connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	database := db.NewPostgresDB(conn)
	gin.SetMode(gin.TestMode)
	router := setupRouter(i18n.GetInstance(), getJwtSecret(cfg), getSessionSecret(cfg))
	c := gin.CreateTestContextOnly(nil, router)
	lockService := middleware.NewLockService(database)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("startup panicked: %v", r)
		}
	}()
	setEntities(c, database, router, getJwtSecret(cfg), cfg, middleware.NewLockMiddleware(lockService), lockService, i18n.GetInstance())
	if n := len(router.Routes()); n == 0 {
		t.Fatal("no routes registered")
	}
	t.Logf("%d routes registered", len(router.Routes()))
}
