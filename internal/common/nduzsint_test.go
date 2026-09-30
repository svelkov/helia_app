package common

import (
	"context"
	"testing"

	"helia/config"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
)

func TestNDuzSint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{NDuzSint: 3}

	c, _ := gin.CreateTestContext(nil)

	// No session at all: the configured default is used.
	if got := NDuzSint(c, cfg); got != 3 {
		t.Fatalf("without a session: got %d, want 3 (cfg default)", got)
	}

	// The session carries the value derived from the bookkeeping type of the firma.
	domain.SetSessionInContext(c, &domain.UserSession{DuzSin: 4})
	if got := NDuzSint(c, cfg); got != 4 {
		t.Fatalf("with session DuzSin=4: got %d, want 4", got)
	}

	// A session without the value (token issued before nDuzSIN existed) falls back as well.
	domain.SetSessionInContext(c, &domain.UserSession{DuzSin: 0})
	if got := NDuzSint(c, cfg); got != 3 {
		t.Fatalf("with session DuzSin=0: got %d, want 3 (cfg default)", got)
	}

	// Service layer variant (standard context).
	ctx := domain.SetSessionInStdContext(context.Background(), &domain.UserSession{DuzSin: 4})
	if got := NDuzSintFromContext(ctx, cfg); got != 4 {
		t.Fatalf("service layer with session DuzSin=4: got %d, want 4", got)
	}
	if got := NDuzSintFromContext(context.Background(), cfg); got != 3 {
		t.Fatalf("service layer without a session: got %d, want 3 (cfg default)", got)
	}

	// A nil session must not panic.
	var nilSession *domain.UserSession
	if got := nilSession.GetDuzSin(3); got != 3 {
		t.Fatalf("nil session: got %d, want 3", got)
	}
}
