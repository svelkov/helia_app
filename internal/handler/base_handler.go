package handler

import (
	"context"
	"fmt"
	"helia/frontend/templates"
	"helia/internal/domain"
	"net/http"
	"time"

	tmpl "helia/frontend/templates"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
)

// Context key for username
type contextKey string

const usernameKey contextKey = "username"

// Secret key for signing the JWT (keep this secure!)
var jwtSecret = []byte("3285f0d71eed0c41fded2115c9cc8ac09a0ab5a519565df10afdb20f8013c5268f2c19948b6af096c1cfc2921ab086be21fa5407b9d91aeb08eeeeef3c2e16c9ae30ae15f27d340f17c450468fef50795e58bb7351a94602bc045aea1a1ff3b03039081208cf067b44fd913b98b712e34ba080941f5ff8545b0eac26824f0ef4a93109939d8f917e1fac1eb588f4272ebac415975bcdc994c3a0fea7c3805d601443ad71dd9043858de5c2bfe64106683d9eaebce28442ce7bb22298d5b85cc3cc41e6f81f9c0f8f678cce559f745645edc5a5009ba20f8b5a16be4ee7dada7791913c90e3629a44b88a17d3d107bd3a6c0f3000b4865b2c015c0875901a028e")

// Secret key for signing the JWT (keep this secure!)
var sessionStore = sessions.NewCookieStore([]byte("3285f0d71eed0c41fded2115c9cc8ac09a0ab5a519565df10afdb20f8013c5268f2c19948b6af096c1cfc2921ab086be21fa5407b9d91aeb08eeeeef3c2e16c9ae30ae15f27d340f17c450468fef50795e58bb7351a94602bc045aea1a1ff3b03039081208cf067b44fd913b98b712e34ba080941f5ff8545b0eac26824f0ef4a93109939d8f917e1fac1eb588f4272ebac415975bcdc994c3a0fea7c3805d601443ad71dd9043858de5c2bfe64106683d9eaebce28442ce7bb22298d5b85cc3cc41e6f81f9c0f8f678cce559f745645edc5a5009ba20f8b5a16be4ee7dada7791913c90e3629a44b88a17d3d107bd3a6c0f3000b4865b2c015c0875901a028e")) // Replace with your secret key

type BasicHandler struct {
	isLoggedIn   bool
	subMenuItems []domain.SubMenuItem
}

func NewBasicHandler(isLoggedIn bool, subMeniItems []domain.SubMenuItem) *BasicHandler {
	return &BasicHandler{isLoggedIn: isLoggedIn, subMenuItems: subMeniItems}
}

func (h *BasicHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Dummy authentication (replace with actual logic)
	if username == "testuser" && password == "123" {
		token, err := GenerateJWT(username)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Set JWT token as an HTTP-only cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
		})

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	templates.Base(false, templates.Login(), h.subMenuItems, "Helia", "", fmt.Sprintf("%d", time.Now().Year())).Render(r.Context(), w) // Render login page on failure.
}

func (h *BasicHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Simulate a successful registration
		http.Redirect(w, r, "/login", http.StatusSeeOther) // Redirect to login page
		return
	}
	templates.Base(h.isLoggedIn, templates.Register(), h.subMenuItems, "Helia", "", fmt.Sprintf("%d", time.Now().Year())).Render(r.Context(), w)
}

func (h *BasicHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	// Simulate a logout
	h.isLoggedIn = false
	http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect to home page
}
func (h *BasicHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect to home page
}

// Middleware to check if the user is logged in.
func (h *BasicHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tokenString := cookie.Value

		// Parse and validate token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Extract username and store in context
		username, ok := claims["username"].(string)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), usernameKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Handler for the main index page
func (h *BasicHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from context
	username, ok := r.Context().Value(usernameKey).(string)

	// Check if user is authenticated
	IsLoggedIn := ok && username != ""

	menuName := r.URL.Query().Get("menuName")
	var submenu []domain.SubMenuItem
	if menuName == "" {
		menuName = "Opšti podaci"
	}
	// Get submenu items
	if val, exists := domain.MenuData[menuName]; exists {
		submenu = val
	} else {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}

	year := fmt.Sprintf("%d", time.Now().Year())
	c := tmpl.Content(IsLoggedIn)
	err := tmpl.Base(IsLoggedIn, c, submenu, "HELIA", username, year).Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

}

// Generate JWT token
func GenerateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Handlers
func (h *BasicHandler) AddRoutes(r *http.ServeMux) {
	r.HandleFunc("/login", h.LoginHandler)
	r.HandleFunc("/register", h.RegisterHandler)
	r.HandleFunc("/logout", h.LogoutHandler)
	r.HandleFunc("/home", h.AuthMiddleware(h.HomeHandler))
	r.HandleFunc("/", h.AuthMiddleware(h.indexHandler))
}
