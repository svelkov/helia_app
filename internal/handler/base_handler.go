package handler

import (
	"encoding/json"
	"fmt"
	"helia/config"
	"helia/frontend/templates"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tmpl "helia/frontend/templates"
	finservice "helia/internal/service/finansijsko"

	"github.com/a-h/templ"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Helper struct for default selections
type defaultSelections struct {
	firma    string
	god      int
	kar      int
	language string
}

// Constants for configuration and business logic
const (
	contextTimeout = 30 * time.Second
	tokenExpiry    = 24 * time.Hour
	sessionName    = "app-session"

	// Session keys
	firma        = "firma"
	god          = "god"
	kar          = "kar"
	csrfTokenKey = "csrf_token"

	// Default values
	defaultMenu = "opsti_podaci"
)

// Context keys
type contextKey string

const usernameKey contextKey = "username"

// Configuration loaded from environment
var (

	// Secret key for signing the JWT (keep this secure!)
	SESSION_SECRET = "3285f0d71eed0c41fded2115c9cc8ac09a0ab5a519565df10afdb20f8013c5268f2c19948b6af096c1cfc2921ab086be21fa5407b9d91aeb08eeeeef3c2e16c9ae30ae15f27d340f17c450468fef50795e58bb7351a94602bc045aea1a1ff3b03039081208cf067b44fd913b98b712e34ba080941f5ff8545b0eac26824f0ef4a93109939d8f917e1fac1eb588f4272ebac415975bcdc994c3a0fea7c3805d601443ad71dd9043858de5c2bfe64106683d9eaebce28442ce7bb22298d5b85cc3cc41e6f81f9c0f8f678cce559f745645edc5a5009ba20f8b5a16be4ee7dada7791913c90e3629a44b88a17d3d107bd3a6c0f3000b4865b2c015c0875901a028e" // Replace with your secret key

/*
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	if len(jwtSecret) == 0 {
	    log.Fatal("JWT_SECRET environment variable not set")
	}

sessionKey = []byte(os.Getenv("SESSION_KEY"))

	if len(sessionKey) == 0 {
	    log.Fatal("SESSION_KEY environment variable not set")
	}

sessionStore = sessions.NewCookieStore(sessionKey)
*/
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// BasicHandler handles base application functionality
type BasicHandler struct {
	isLoggedIn   bool
	menuItems    domain.MenuDataItems
	subMenuItems []domain.SubMenuItem
	fvrService   finservice.FvrService
	firma        domain.Firma
	cfg          config.Config
	logger       *log.Logger
}

// NewBasicHandler creates and initializes a new BasicHandler
func NewBasicHandler(c *gin.Context, isLoggedIn bool, menuItems domain.MenuDataItems, subMenuItems []domain.SubMenuItem, fvrService finservice.FvrService, cfg config.Config) *BasicHandler {
	handler := &BasicHandler{
		isLoggedIn:   isLoggedIn,
		menuItems:    menuItems,
		subMenuItems: subMenuItems,
		fvrService:   fvrService,
		firma:        domain.Firma{},
		logger:       log.New(os.Stdout, "[BasicHandler] ", log.LstdFlags|log.Lshortfile),
		cfg:          cfg,
	}

	// Initialize firma data
	if err := handler.setFirma(c); err != nil {
		handler.logger.Printf("Failed to initialize firma data: %v", err)
	}

	return handler
}

// Error response helper with log levels
func (h *BasicHandler) respondWithError(c *gin.Context, code int, message string) {
	// Log with appropriate level based on status code
	switch {
	case code >= 500:
		h.logger.Printf("SERVER ERROR %d: %s", code, message)
	case code >= 400:
		h.logger.Printf("CLIENT ERROR %d: %s", code, message)
	default:
		h.logger.Printf("Error %d: %s", code, message)
	}

	c.JSON(code, gin.H{"error": message})
}

// If you want to use Gin's session middleware instead
func (h *BasicHandler) renderLoginPage(c *gin.Context, fvrData domain.Firma, selections defaultSelections) {
	// Generate CSRF token
	csrfToken := h.generateCSRFToken(c)

	c.Header("content-type", "text/html")
	err := templates.Base(false,
		templates.Login(i18n.GetInstance(), csrfToken),
		h.menuItems,
		h.subMenuItems,
		"Helia",
		"",
		fmt.Sprintf("%d", time.Now().Year()),
		"",
		setComboFirmaConfig(fvrData, selections.firma),
		setComboPoslGodConfig(fvrData, selections.firma, selections.god),
		setComboKarConfig(fvrData, selections.firma, selections.god, selections.kar),
		setComboLanguageConfig(selections.language, h.cfg),
		i18n.GetInstance(),
	).Render(c.Request.Context(), c.Writer)

	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "Error rendering login page")
		return
	}
}

// LoginHandler manages user authentication
func (h *BasicHandler) LoginHandler(c *gin.Context) {
	fvrData := h.getFirma()
	selections := h.getDefaultSelections(fvrData)

	if c.Request.Method == http.MethodGet {
		h.renderLoginPage(c, fvrData, selections)
		return
	}

	// Handle POST request
	h.handleLoginPost(c, selections)
}

func (h *BasicHandler) handleLoginPost(c *gin.Context, selections defaultSelections) {
	var loginData struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if err := c.ShouldBind(&loginData); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Username and password are required")
		return
	}

	if !h.validateCredentials(loginData.Username, loginData.Password) {
		h.logger.Printf("Failed login attempt for user: %s", loginData.Username)
		h.respondWithError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := h.generateToken(loginData.Username, selections, h.getJwtSecretFromContext(c))
	if err != nil {
		h.logger.Printf("Error generating token for user %s: %v", loginData.Username, err)
		h.respondWithError(c, http.StatusInternalServerError, "Error generating authentication token")
		return
	}

	// Set auth cookie
	h.setAuthCookie(c, token)

	// Update user session in context (per-user, request-scoped)
	h.updateUserSession(c, selections)

	// Handle response based on request type
	if c.GetHeader("HX-Request") == "true" {
		c.JSON(200, gin.H{
			"success":  true,
			"message":  "Login successful",
			"redirect": "/",
		})
	} else {
		c.Redirect(http.StatusSeeOther, "/")
	}
}

func (h *BasicHandler) setAuthCookie(c *gin.Context, token string) {
	c.SetCookie("auth_token", token, int(tokenExpiry/time.Second), "/", "", true, true)
}

func (h *BasicHandler) updateUserSession(c *gin.Context, selections defaultSelections) {
	// Get or create UserSession in context
	session := domain.GetSessionFromContext(c)
	if session == nil {
		// Create new UserSession for newly logged-in user
		userName := c.PostForm("username")
		session = &domain.UserSession{
			UserID:   0, // TODO: Set from user lookup
			UserName: userName,
		}
	}

	// Update session values
	session.Firma = selections.firma
	session.SelectedGod = selections.god
	session.SelectedKar = selections.kar
	if selections.language != "" {
		session.Language = selections.language
	}

	// Store in context for this request and subsequent handlers
	c.Set("userSession", session)
}

// getDefaultSelections returns default values for firma, god, kar, and language
func (h *BasicHandler) getDefaultSelections(fvrData domain.Firma) defaultSelections {
	selections := defaultSelections{
		language: "sr", // Default language
	}

	if len(fvrData.Firme) > 0 {
		selections.firma = fvrData.Firme[0].Naziv

		if len(fvrData.Firme[0].Godine) > 0 {
			selections.god = fvrData.Firme[0].Godine[0].God

			if len(fvrData.Firme[0].Godine[0].Kar) > 0 {
				selections.kar = fvrData.Firme[0].Godine[0].Kar[0]
			}
		}
	}

	return selections
}

// validateCredentials checks if the provided credentials are valid
func (h *BasicHandler) validateCredentials(username, password string) bool {
	// TODO: Replace with actual authentication logic
	return username == "testuser" && password == "123"
}

// generateToken creates a new JWT token for the authenticated user with preferences
func (h *BasicHandler) generateToken(username string, selections defaultSelections, jwtSecret []byte) (string, error) {
	claims := domain.UserClaims{
		Username:    username,
		UserID:      0, // TODO: Set from user lookup
		Firma:       selections.firma,
		SelectedGod: selections.god,
		SelectedKar: selections.kar,
		Language:    selections.language,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "HELIA",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// getJwtSecretFromContext retrieves the JWT secret from Gin context
func (h *BasicHandler) getJwtSecretFromContext(c *gin.Context) []byte {
	if s, exists := c.Get("jwtSecret"); exists {
		if secretBytes, ok := s.([]byte); ok {
			return secretBytes
		}
	}
	return nil
}

// regenerateToken creates a new JWT token with updated UserSession preferences
func (h *BasicHandler) regenerateToken(c *gin.Context, userSession *domain.UserSession) {
	if userSession == nil {
		return
	}

	jwtSecret := h.getJwtSecretFromContext(c)
	if jwtSecret == nil {
		h.logger.Println("JWT secret not found, cannot regenerate token")
		return
	}

	// Get username from current token
	username, ok := h.getUsernameFromToken(c)
	if !ok {
		h.logger.Println("Cannot get username from token, cannot regenerate")
		return
	}

	// Create new token with updated preferences
	claims := domain.UserClaims{
		Username:    username,
		UserID:      int(userSession.UserID),
		Firma:       userSession.Firma,
		SelectedGod: userSession.SelectedGod,
		SelectedKar: userSession.SelectedKar,
		Language:    userSession.Language,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "HELIA",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		h.logger.Printf("Error regenerating token: %v", err)
		return
	}

	// Update auth cookie with new token
	h.setAuthCookie(c, tokenString)
}

// generateCSRFToken creates a new CSRF token and stores it in session
func (h *BasicHandler) generateCSRFToken(c *gin.Context) string {
	csrfToken := uuid.New().String()
	session := sessions.Default(c)
	session.Set(csrfTokenKey, csrfToken)
	if err := session.Save(); err != nil {
		h.logger.Printf("Failed to save CSRF token to session: %v", err)
	}
	return csrfToken
}

// validateCSRFToken checks if the provided CSRF token is valid
func (h *BasicHandler) validateCSRFToken(c *gin.Context) bool {
	// Get CSRF token from form
	formToken := c.PostForm("_csrf")

	if formToken == "" {
		return false
	}

	// Get CSRF token from session
	session := sessions.Default(c)
	sessionToken := session.Get(csrfTokenKey)

	if sessionToken == nil {
		return false
	}

	// Compare tokens
	return formToken == sessionToken.(string)
}

// clearCSRFToken removes the CSRF token from session after use
func (h *BasicHandler) clearCSRFToken(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(csrfTokenKey)
	if err := session.Save(); err != nil {
		h.logger.Printf("Failed to clear CSRF token from session: %v", err)
	}
}

// --- Authentication Handlers: Register and Logout ---

// RegisterHandler manages user registration
// RegisterHandler manages user registration
func (h *BasicHandler) RegisterHandler(c *gin.Context) {
	fvrData := h.getFirma()
	selections := h.getDefaultSelections(fvrData)

	if c.Request.Method == http.MethodGet {
		err := templates.Base(
			false,
			templates.Register(i18n.GetInstance()),
			h.menuItems,
			h.subMenuItems,
			"Helia - Registration",
			"",
			fmt.Sprintf("%d", time.Now().Year()),
			"",
			setComboFirmaConfig(fvrData, selections.firma),
			setComboPoslGodConfig(fvrData, selections.firma, selections.god),
			setComboKarConfig(fvrData, selections.firma, selections.god, selections.kar),
			setComboLanguageConfig(selections.language, h.cfg),
			i18n.GetInstance(),
		).Render(c.Request.Context(), c.Writer)

		if err != nil {
			h.logger.Printf("Error rendering registration page: %v", err)
			h.respondWithError(c, http.StatusInternalServerError, "Error rendering registration page")
			return
		}
		return
	}

	// Handle POST - process registration
	username := c.PostForm("username")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Validate input
	if err := h.validateRegistration(username, password, confirmPassword); err != nil {
		h.logger.Printf("Registration validation failed: %v", err)
		h.respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: Add actual user registration logic here
	// For now, just redirect to login page
	c.Redirect(http.StatusSeeOther, "/login")
}

// LogoutHandler manages user logout
func (h *BasicHandler) LogoutHandler(c *gin.Context) {
	// Clear the auth cookie
	c.SetCookie(
		"auth_token",
		"",
		-1,
		"/",
		"",
		true, // Secure
		true, // HttpOnly
	)

	// Clear the session
	// Get session
	session := sessions.Default(c)
	session.Options(sessions.Options{
		MaxAge: -1, // Configures cookie expiration
	})
	if err := session.Save(); err != nil {
		h.logger.Printf("Error clearing session during logout: %v", err)
	}

	// Clear user session from context (it will be removed by middleware on next request)
	c.Set("userSession", nil)

	h.isLoggedIn = false
	c.Redirect(http.StatusSeeOther, "/login")
}

// Helper functions for registration
func (h *BasicHandler) validateRegistration(username, password, confirmPassword string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}
	return nil
}

// --- Main Page Handlers ---

// indexHandler handles the main page rendering
func (h *BasicHandler) indexHandler(c *gin.Context) {
	username, ok := h.getUsernameFromToken(c)
	fvrData := h.getFirma()
	isLoggedIn := ok && username != ""
	// Get user session from context
	userSession := domain.GetSessionFromContext(c)

	// Get firma data and selections
	selections := h.getUserSessionSelections(userSession, fvrData)
	//set the selected language
	i18n.GetInstance().SetLanguage(selections.language)
	// if !isLoggedIn {
	// 	h.renderLoginPage(c, fvrData, selections)
	// 	return
	// }
	menuName := c.Query("menuName")
	if menuName == "" {
		menuName = defaultMenu
	}
	// Get submenu items
	subMenus := common.GetTranslatedSubMenus(domain.MenuData, menuName, selections.language)
	if subMenus == nil {
		h.logger.Printf("Menu not found: %s", menuName)
		h.respondWithError(c, http.StatusNotFound, "Menu not found")
		return
	}

	// Update handler state
	h.subMenuItems = subMenus
	h.menuItems.CurrentMenu = menuName

	// Render the page
	err := tmpl.Base(
		isLoggedIn,
		tmpl.Content(isLoggedIn, i18n.GetInstance()),
		h.menuItems,
		h.subMenuItems,
		"HELIA",
		username,
		fmt.Sprintf("%d", time.Now().Year()),
		menuName,
		setComboFirmaConfig(fvrData, selections.firma),
		setComboPoslGodConfig(fvrData, selections.firma, selections.god),
		setComboKarConfig(fvrData, selections.firma, selections.god, selections.kar),
		setComboLanguageConfig(selections.language, h.cfg),
		i18n.GetInstance(),
	).Render(c.Request.Context(), c.Writer)

	if err != nil {
		h.logger.Printf("Error rendering template: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "Error rendering template")
		return
	}
}

// HomeHandler redirects to the index page
func (h *BasicHandler) HomeHandler(c *gin.Context) {
	c.Redirect(http.StatusSeeOther, "/")
}

// getCurrentDate returns the current date in JSON format
func (h *BasicHandler) getCurrentDate(c *gin.Context) {
	currentDate := time.Now().Format(common.HtmlLayout)

	response := struct {
		CurrentDate string `json:"currentDate"`
	}{
		CurrentDate: currentDate,
	}
	// Manually encode JSON and handle errors
	c.Header("Content-Type", "application/json")
	c.Status(200)
	if err := json.NewEncoder(c.Writer).Encode(response); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "Error generating response")
		return
	}
}

// Helper functions

// getUserSessionSelections returns selections from UserSession in context
func (h *BasicHandler) getUserSessionSelections(userSession *domain.UserSession, fvrData domain.Firma) defaultSelections {
	if userSession == nil {
		// Return defaults from fvrData
		return h.getDefaultSelections(fvrData)
	}

	selections := defaultSelections{
		firma:    userSession.Firma,
		god:      userSession.SelectedGod,
		kar:      userSession.SelectedKar,
		language: userSession.Language,
	}

	// Set defaults if session values are empty
	if selections.firma == "" && len(fvrData.Firme) > 0 {
		selections.firma = fvrData.Firme[0].Naziv
	}
	if selections.god == 0 && len(fvrData.Firme) > 0 && len(fvrData.Firme[0].Godine) > 0 {
		selections.god = fvrData.Firme[0].Godine[0].God
	}
	if selections.kar == 0 && len(fvrData.Firme) > 0 && len(fvrData.Firme[0].Godine) > 0 &&
		len(fvrData.Firme[0].Godine[0].Kar) > 0 {
		selections.kar = fvrData.Firme[0].Godine[0].Kar[0]
	}
	if selections.language == "" {
		selections.language = "en" // default language
	}
	return selections
}

// getSessionSelections is deprecated - kept for backwards compatibility
func (h *BasicHandler) getSessionSelections(session sessions.Session, fvrData domain.Firma) defaultSelections {
	firmaVal := session.Get("firma")
	godVal := session.Get("god")
	karVal := session.Get("kar")
	langVal := session.Get("language")
	firma, ok1 := firmaVal.(string)
	god, ok2 := godVal.(int)
	kar, ok3 := karVal.(int)
	lang, _ := langVal.(string)

	if !ok1 || !ok2 || !ok3 {
		return defaultSelections{firma: "", god: 0, kar: 0}
	}
	selections := defaultSelections{
		firma:    firma,
		god:      god,
		kar:      kar,
		language: lang,
	}

	// Set defaults if session values are empty
	if selections.firma == "" && len(fvrData.Firme) > 0 {
		selections.firma = fvrData.Firme[0].Naziv
	}
	if selections.god == 0 && len(fvrData.Firme) > 0 && len(fvrData.Firme[0].Godine) > 0 {
		selections.god = fvrData.Firme[0].Godine[0].God
	}
	if selections.kar == 0 && len(fvrData.Firme) > 0 && len(fvrData.Firme[0].Godine) > 0 &&
		len(fvrData.Firme[0].Godine[0].Kar) > 0 {
		selections.kar = fvrData.Firme[0].Godine[0].Kar[0]
	}
	if selections.language == "" {
		selections.language = lang
	}
	return selections
}

func (h *BasicHandler) getUsernameFromToken(c *gin.Context) (string, bool) {
	tokenString, err := c.Cookie("auth_token")
	if err != nil {
		return "", false
	}

	// Get JWT secret from context
	var secret []byte
	if s, exists := c.Get("jwtSecret"); exists {
		if secretBytes, ok := s.([]byte); ok {
			secret = secretBytes
		}
	}

	if secret == nil {
		return "", false
	}

	username, err := infrastructure.VerifyJWT(tokenString, secret)
	if err != nil {
		return "", false
	}

	return username.Username, true
}

// --- Combo Box Handlers ---

// ComboBox handlers for firma, godina, and knjigovodstvo selection
func (h *BasicHandler) setComboFirma(c *gin.Context) {
	firma := c.Query("firma")
	if firma == "" {
		firma = c.PostForm("firma")
	}

	if firma != "" {
		// Update user session in context (per-user, request-scoped)
		userSession := domain.GetSessionFromContext(c)
		if userSession != nil {
			userSession.Firma = firma
			// Reset dependent fields
			userSession.SelectedGod = 0
			userSession.SelectedKar = 0
			c.Set("userSession", userSession)

			// Regenerate JWT token with new preferences
			h.regenerateToken(c, userSession)
		}
	}

	h.renderComboResponse(c)
}

func (h *BasicHandler) SelectComboFirma(c *gin.Context) {
	// UserSession already has firma from context, just render
	h.renderFullPage(c)
}

func (h *BasicHandler) SetComboGod(c *gin.Context) {
	god := c.Query("god")
	if god == "" {
		god = c.PostForm("god")
	}

	if god != "" {
		if gnGod, err := strconv.Atoi(god); err == nil {
			// Update user session in context (per-user, request-scoped)
			userSession := domain.GetSessionFromContext(c)
			if userSession != nil {
				userSession.SelectedGod = gnGod

				// Regenerate JWT token with new preferences
				h.regenerateToken(c, userSession)
				// Reset dependent field
				userSession.SelectedKar = 0
				c.Set("userSession", userSession)
			}
		}
	}

	h.renderComboResponse(c)
}

func (h *BasicHandler) SelectComboGod(c *gin.Context) {
	// UserSession already has god from context, just render
	h.renderFullPage(c)
}

func (h *BasicHandler) SetComboKar(c *gin.Context) {
	kar := c.Query("kar")
	if kar == "" {
		kar = c.PostForm("kar")
	}

	if kar != "" {
		if gnKar, err := strconv.Atoi(kar); err == nil {
			// Update user session in context (per-user, request-scoped)
			userSession := domain.GetSessionFromContext(c)

			// Regenerate JWT token with new preferences
			h.regenerateToken(c, userSession)
			if userSession != nil {
				userSession.SelectedKar = gnKar
				c.Set("userSession", userSession)
			}
		}
	}

	h.renderComboResponse(c)
}

func (h *BasicHandler) SelectComboKar(c *gin.Context) {
	// UserSession already has kar from context, just render
	h.renderFullPage(c)
}
func (h *BasicHandler) SelectComboLanguage(c *gin.Context) {
	lang := c.Query("language")

	// Update user session in context (per-user, request-scoped)
	userSession := domain.GetSessionFromContext(c)

	// Regenerate JWT token with new preferences
	h.regenerateToken(c, userSession)
	if userSession != nil {
		userSession.Language = lang
		c.Set("userSession", userSession)
	}

	i18n.GetInstance().SetLanguage(lang)

	h.renderFullPage(c)
}

// Helper functions for combo box handlers
func (h *BasicHandler) renderComboResponse(c *gin.Context) {
	fvrData := h.getFirma()
	userSession := domain.GetSessionFromContext(c)

	selections := h.getUserSessionSelections(userSession, fvrData)
	response := struct {
		FirmaConfig domain.ComboFieldConfig `json:"firmaConfig"`
		GodConfig   domain.ComboFieldConfig `json:"godConfig"`
		KarConfig   domain.ComboFieldConfig `json:"karConfig"`
	}{
		FirmaConfig: setComboFirmaConfig(fvrData, selections.firma),
		GodConfig:   setComboPoslGodConfig(fvrData, selections.firma, selections.god),
		KarConfig:   setComboKarConfig(fvrData, selections.firma, selections.god, selections.kar),
	}

	// Manually encode JSON and handle errors
	c.Header("Content-Type", "application/json")
	c.Status(200)
	if err := json.NewEncoder(c.Writer).Encode(response); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "Error generating response")
		return
	}
}

// renderFullPage handles the full page rendering with all components
func (h *BasicHandler) renderFullPage(c *gin.Context) {
	// Get user session from context (per-user, request-scoped)
	userSession := domain.GetSessionFromContext(c)

	// Track if we need to regenerate token
	needsTokenRegeneration := false

	// Update session from query parameters if provided
	if userSession != nil {
		currentLanguage := c.Query("language")
		if currentLanguage != "" && currentLanguage != userSession.Language {
			userSession.Language = currentLanguage
			needsTokenRegeneration = true
		}
		currentFirma := c.Query("firma")
		if currentFirma != "" && currentFirma != userSession.Firma {
			userSession.Firma = currentFirma
			needsTokenRegeneration = true
		}
		currentGod := c.Query("god")
		if currentGod != "" {
			if gnGod, err := strconv.Atoi(currentGod); err == nil && gnGod != userSession.SelectedGod {
				userSession.SelectedGod = gnGod
				needsTokenRegeneration = true
			}
		}
		currentKar := c.Query("kar")
		if currentKar != "" {
			if gnKar, err := strconv.Atoi(currentKar); err == nil && gnKar != userSession.SelectedKar {
				userSession.SelectedKar = gnKar
				needsTokenRegeneration = true
			}
		}
		c.Set("userSession", userSession)

		// Regenerate token if any preference changed
		if needsTokenRegeneration {
			h.regenerateToken(c, userSession)
		}
	}
	username, ok := h.getUsernameFromToken(c)
	if !ok {
		h.logger.Print("No valid user token found")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Get menu information
	menuName := c.Query("menuName")
	if menuName == "" {
		menuName = defaultMenu
	}

	// Get submenu items
	subMenus := common.GetTranslatedSubMenus(domain.MenuData, menuName, userSession.Language)
	if subMenus == nil {
		h.logger.Printf("Menu not found: %s", menuName)
		h.respondWithError(c, http.StatusNotFound, "Menu not found")
		return
	}

	// Update handler state
	h.subMenuItems = subMenus
	h.menuItems.CurrentMenu = menuName

	fvrData := h.getFirma()
	selections := h.getUserSessionSelections(userSession, fvrData)

	// Prepare page data
	pageData := struct {
		IsLoggedIn   bool
		Content      templ.Component
		MenuItems    domain.MenuDataItems
		SubMenus     []domain.SubMenuItem
		Title        string
		Username     string
		Year         string
		MenuName     string
		FirmaConf    domain.ComboFieldConfig
		GodConf      domain.ComboFieldConfig
		KarConf      domain.ComboFieldConfig
		LanguageConf domain.ComboFieldConfig
	}{
		IsLoggedIn:   true,
		Content:      tmpl.Content(true, i18n.GetInstance()),
		MenuItems:    h.menuItems,
		SubMenus:     h.subMenuItems,
		Title:        "HELIA",
		Username:     username,
		Year:         fmt.Sprintf("%d", time.Now().Year()),
		MenuName:     menuName,
		FirmaConf:    setComboFirmaConfig(fvrData, selections.firma),
		GodConf:      setComboPoslGodConfig(fvrData, selections.firma, selections.god),
		KarConf:      setComboKarConfig(fvrData, selections.firma, selections.god, selections.kar),
		LanguageConf: setComboLanguageConfig(selections.language, h.cfg),
	}

	// Render the page
	err := tmpl.Base(
		pageData.IsLoggedIn,
		pageData.Content,
		pageData.MenuItems,
		pageData.SubMenus,
		pageData.Title,
		pageData.Username,
		pageData.Year,
		pageData.MenuName,
		pageData.FirmaConf,
		pageData.GodConf,
		pageData.KarConf,
		pageData.LanguageConf,
		i18n.GetInstance(),
	).Render(c.Request.Context(), c.Writer)

	if err != nil {
		h.logger.Printf("Error rendering template: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "Error rendering template")
		return
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// AuthMiddleware verifies the user is authenticated

// AddRoutes registers all HTTP routes for the BasicHandler
func (h *BasicHandler) AddRoutes(r *gin.Engine) {
	// Authentication routes
	// Create API group with prefix

	//r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	// Handle favicon to prevent auto-requests
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204) // No content
	})

	r.GET("/login", h.LoginHandler)
	r.POST("/login", h.LoginHandler)
	r.GET("/register", h.RegisterHandler)
	r.GET("/logout", h.LogoutHandler)

	// Main page routes
	r.GET("/", h.indexHandler)
	r.GET("/home", h.HomeHandler)

	// r routes
	r.GET("/api/get-current-date", h.getCurrentDate)

	// Combo box routes
	r.GET("/api/setfirma", h.setComboFirma)
	r.GET("/api/selectfirma", h.SelectComboFirma)
	r.GET("/api/setgod", h.SetComboGod)
	r.GET("/api/selectgod", h.SelectComboGod)
	r.GET("/api/selectkar", h.SelectComboKar)
	r.GET("/api/setkar", h.SetComboKar)
	r.GET("/api/selectlanguage", h.SelectComboLanguage)
	//r.GET("/api/setlanguage", h.SetComboLanguage)
}

// setFirma initializes the firma data from the service
func (h *BasicHandler) setFirma(c *gin.Context) error {
	fvrData, err := h.fvrService.GetAllFvr(c)
	if err != nil {
		h.logger.Printf("Error fetching firma data: %v", err)
		return fmt.Errorf("failed to fetch firma data: %w", err)
	}

	if fvrData == nil || len(fvrData.Firme) == 0 {
		return fmt.Errorf("no firma data available")
	}

	h.firma = *fvrData
	return nil
}

// getFirma returns the cached firma data
func (h *BasicHandler) getFirma() domain.Firma {
	return h.firma
}

// setComboFirmaConfig creates configuration for the firma combo box
func setComboFirmaConfig(fvrData domain.Firma, selectedValue string) domain.ComboFieldConfig {
	configCombo := domain.ComboFieldConfig{
		ID:             "firma",
		Name:           "firma",
		LabelText:      "Preduzece/Komintent",
		HasLabel:       true,
		Disabled:       false,
		ClassSelect:    common.ClassSelect + " min-w-[80px]",
		ClassLabel:     common.ClassLabel + " font-medium text-white text-sm whitespace-nowrap",
		HxVals:         `js:{"firma": this.value}`,
		ChangeEndpoint: "/api/selectfirma",
		HxSwap:         "outerHTML",
		HxChangeTarget: "body",
	}

	var optionItems []domain.ComboItem
	for _, firma := range fvrData.Firme {
		optionItems = append(optionItems, domain.ComboItem{
			Key:   firma.Naziv,
			Value: firma.Naziv,
		})
	}

	configCombo.OptionValues = optionItems
	configCombo.SelectedValue = selectedValue
	return configCombo
}

// setComboPoslGodConfig creates configuration for the poslovna godina combo box
func setComboPoslGodConfig(fvrData domain.Firma, selectedFirma string, selectedValue int) domain.ComboFieldConfig {
	configCombo := domain.ComboFieldConfig{
		ID:             "god",
		Name:           "god",
		LabelText:      "Poslovna Godina",
		HasLabel:       true,
		Disabled:       selectedFirma == "",
		ClassSelect:    common.ClassSelect + " min-w-[80px]",
		ClassLabel:     common.ClassLabel + " font-medium text-white text-sm whitespace-nowrap",
		HxVals:         `js:{"firma": document.getElementById("firma").value, "god": this.value}`,
		ChangeEndpoint: "/api/selectgod",
		HxSwap:         "outerHTML",
		HxChangeTarget: "body",
	}

	var optionItems []domain.ComboItem
	for _, firma := range fvrData.Firme {
		if firma.Naziv == selectedFirma {
			for _, god := range firma.Godine {
				key := fmt.Sprintf("%d", god.God)
				optionItems = append(optionItems, domain.ComboItem{
					Key:   key,
					Value: key,
				})
			}
			break
		}
	}

	configCombo.OptionValues = optionItems
	configCombo.SelectedValue = fmt.Sprintf("%d", selectedValue)
	return configCombo
}

// setComboKarConfig creates configuration for the knjigovodstvo combo box
func setComboKarConfig(fvrData domain.Firma, selectedFirma string, selectedGod, selectedValue int) domain.ComboFieldConfig {
	configCombo := domain.ComboFieldConfig{
		ID:             "kar",
		Name:           "kar",
		LabelText:      "Knjigovodstvo",
		HasLabel:       true,
		Disabled:       selectedFirma == "" || selectedGod == 0,
		ClassSelect:    common.ClassSelect + " min-w-[80px]",
		ClassLabel:     common.ClassLabel + " font-medium text-white text-sm whitespace-nowrap",
		HxVals:         `js:{"firma": document.getElementById("firma").value, "god": document.getElementById("god").value, "kar": this.value}`,
		ChangeEndpoint: "/api/selectkar",
		HxSwap:         "outerHTML",
		HxChangeTarget: "body",
	}

	var optionItems []domain.ComboItem
	for _, firma := range fvrData.Firme {
		if firma.Naziv == selectedFirma {
			for _, god := range firma.Godine {
				if god.God == selectedGod {
					for _, kar := range god.Kar {
						key := fmt.Sprintf("%d", kar)
						optionItems = append(optionItems, domain.ComboItem{
							Key:   key,
							Value: key,
						})
					}
					break
				}
			}
			break
		}
	}

	configCombo.OptionValues = optionItems
	configCombo.SelectedValue = fmt.Sprintf("%d", selectedValue)
	return configCombo
}

// setComboKarConfig creates configuration for the knjigovodstvo combo box
func setComboLanguageConfig(selectedValue string, cfg config.Config) domain.ComboFieldConfig {

	configCombo := domain.ComboFieldConfig{
		ID:             "language",
		Name:           "language",
		LabelText:      "",
		HasLabel:       false,
		ClassSelect:    common.ClassSelect + " min-w-[40px]",
		HxVals:         `js:{"language": document.getElementById("language").value}`,
		ChangeEndpoint: "/api/selectlanguage",
		HxSwap:         "outerHTML",
		HxChangeTarget: "body",
	}

	var optionItems []domain.ComboItem
	languages := cfg.Languages
	for _, l := range languages {
		key := fmt.Sprintf("%s", l)
		optionItems = append(optionItems, domain.ComboItem{
			Key:   key,
			Value: key,
		})
	}

	configCombo.OptionValues = optionItems
	configCombo.SelectedValue = fmt.Sprintf("%s", selectedValue)
	return configCombo
}
