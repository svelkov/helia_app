package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TestUserSessionDuzSin verifies that the bookkeeping dependent nDuzSIN stored in the token is put on
// the request session, where the handlers and services read it from (common.NDuzSint).
func TestUserSessionDuzSin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret")
	cfg := config.Config{NDuzSint: 3}

	newToken := func(duzSin int) string {
		claims := domain.UserClaims{
			Username: "tester",
			UserID:   1,
			Firma:    "ERA-TEX",
			DuzSin:   duzSin,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(secret)
		if err != nil {
			t.Fatalf("signing: %v", err)
		}
		return signed
	}

	for _, tc := range []struct {
		name   string
		duzSin int
		want   int
	}{
		{"Pogonska firma (4)", 4, 4},
		{"Finansijska firma (3)", 3, 3},
		{"Token without the value falls back to config", 0, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotSession int
			var gotCommon int

			r := gin.New()
			r.Use(UserSession(secret))
			r.GET("/probe", func(c *gin.Context) {
				gotSession = domain.GetSessionFromContext(c).DuzSin
				gotCommon = common.NDuzSint(c, cfg)
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			req.AddCookie(&http.Cookie{Name: "auth_token", Value: newToken(tc.duzSin)})
			r.ServeHTTP(httptest.NewRecorder(), req)

			if gotSession != tc.duzSin {
				t.Errorf("session DuzSin = %d, want %d", gotSession, tc.duzSin)
			}
			if gotCommon != tc.want {
				t.Errorf("common.NDuzSint = %d, want %d", gotCommon, tc.want)
			}
		})
	}
}
