package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"helia/frontend/templates"
	"helia/global"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
	"net/http"
	"strconv"
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
	menuItems    domain.MenuDataItems
	subMenuItems []domain.SubMenuItem
	fvrService   service.FvrService
	store        *sessions.CookieStore
}

func NewBasicHandler(isLoggedIn bool, menuItems domain.MenuDataItems, subMenuItems []domain.SubMenuItem, fvrService service.FvrService) *BasicHandler {
	return &BasicHandler{
		isLoggedIn:   isLoggedIn,
		menuItems:    menuItems,
		subMenuItems: subMenuItems,
		fvrService:   fvrService,
		store:        sessionStore,
	}
}

func (h *BasicHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	fvrData, poslGodina, err := h.fvrService.GetAllFvr()
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		selectedKomintent := 0
		selectedPoslGodina := 0
		if len(fvrData) > 0 {
			selectedKomintent = fvrData[0].Kar
		}
		if len(poslGodina) > 0 {
			selectedPoslGodina = poslGodina[0]
		}
		templates.Base(false, templates.Login(), h.menuItems, h.subMenuItems, "Helia", "", fmt.Sprintf("%d", time.Now().Year()), "", setComboKomintent(fvrData, selectedKomintent), setComboPoslGodConfig(poslGodina, selectedPoslGodina)).Render(r.Context(), w) // Render login page on failure.
		return
	}
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
	selectedKomintent := 0
	selectedPoslGodina := 0
	if len(fvrData) > 0 {
		selectedKomintent = fvrData[0].Kar
	}
	if len(poslGodina) > 0 {
		selectedPoslGodina = poslGodina[0]
	}
	global.SetGnGod(selectedPoslGodina)
	global.SetGnKar(selectedKomintent)
	templates.Base(false, templates.Login(), h.menuItems, h.subMenuItems, "Helia", "", fmt.Sprintf("%d", time.Now().Year()), "", setComboKomintent(fvrData, selectedKomintent), setComboPoslGodConfig(poslGodina, selectedPoslGodina)).Render(r.Context(), w) // Render login page on failure.
}

func (h *BasicHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Simulate a successful registration
		http.Redirect(w, r, "/login", http.StatusSeeOther) // Redirect to login page
		return
	}
	fvrData, poslGodina, err := h.fvrService.GetAllFvr()
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	selectedKomintent := -1
	selectedPoslGodina := -1
	if len(fvrData) > 0 {
		selectedKomintent = fvrData[0].Kar
	}
	if len(poslGodina) > 0 {
		selectedPoslGodina = poslGodina[0]
	}
	global.SetGnGod(selectedPoslGodina)
	global.SetGnKar(selectedKomintent)
	templates.Base(h.isLoggedIn, templates.Register(), h.menuItems, h.subMenuItems, "Helia", "", fmt.Sprintf("%d", time.Now().Year()), "", setComboKomintent(fvrData, selectedKomintent), setComboPoslGodConfig(poslGodina, selectedPoslGodina)).Render(r.Context(), w)
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

func (h *BasicHandler) getCurrentDate(w http.ResponseWriter, r *http.Request) {
	currentDate := time.Now().Format("2006-01-02")
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"currentDate": currentDate}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error generating JSON response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse) // Write the JSON response
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

	username, ok := getUsernameFromToken(r)

	// Check if user is authenticated
	IsLoggedIn := ok && username != ""

	menuName := r.URL.Query().Get("menuName")
	if menuName == "" {
		menuName = "opsti_podaci" // Default menu name
	}
	subMenus := common.GetSubMenus(domain.MenuData, menuName)
	// Get submenu items
	if subMenus == nil {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}
	h.subMenuItems = subMenus
	h.menuItems.CurrentMenu = menuName
	year := fmt.Sprintf("%d", time.Now().Year())
	c := tmpl.Content(IsLoggedIn)
	fvrData, _, err := h.fvrService.GetAllFvr()
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

	// Get user session
	session, err := h.store.Get(r, "app-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Get values from session with defaults
	currentGod, _ := session.Values["poslovnagodina"].(int)
	currentKomintent, _ := session.Values["komintent"].(int)
	poslGodina, err := h.fvrService.GetAllGod(currentKomintent)
	if err != nil {
		http.Error(w, "Error while getting years", http.StatusInternalServerError)
		return
	}
	global.SetGnGod(currentGod)
	global.SetGnKar(currentKomintent)
	err = tmpl.Base(IsLoggedIn, c, h.menuItems, h.subMenuItems, "HELIA", username, year, menuName, setComboKomintent(fvrData, currentKomintent), setComboPoslGodConfig(poslGodina, currentGod)).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

}

func (h *BasicHandler) SetPoslovnaGod(w http.ResponseWriter, r *http.Request) {
	// Get komintent ID from request
	komintentID := r.URL.Query().Get("komintent")
	if komintentID == "" {
		komintentID = r.FormValue("komintent")
	}

	// Get user session
	session, err := h.store.Get(r, "app-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if komintentID != "" {
		if gnKar, err := strconv.Atoi(komintentID); err == nil {
			session.Values["komintent"] = gnKar
			global.SetGnKar(gnKar)
		}
	}

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Instead of redirecting, re-render the full page
	h.renderFullPage(w, r)
}

func (h *BasicHandler) SelectPoslovnaGodina(w http.ResponseWriter, r *http.Request) {
	poslGod := r.URL.Query().Get("poslovnagodina")
	komintent := r.URL.Query().Get("komintent")

	// Get user session
	session, err := h.store.Get(r, "app-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Store in session
	if poslGod != "" {
		if gnGod, err := strconv.Atoi(poslGod); err == nil {
			session.Values["poslovnagodina"] = gnGod
			global.SetGnGod(gnGod)
		}
	}

	if komintent != "" {
		if gnKar, err := strconv.Atoi(komintent); err == nil {
			session.Values["komintent"] = gnKar
			global.SetGnKar(gnKar)
		}
	}

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Instead of redirecting, re-render the full page
	h.renderFullPage(w, r)
}

// Helper function to render the full page
func (h *BasicHandler) renderFullPage(w http.ResponseWriter, r *http.Request) {
	// Get username from context
	username, ok := getUsernameFromToken(r)
	IsLoggedIn := ok && username != ""

	menuName := r.URL.Query().Get("menuName")
	if menuName == "" {
		menuName = "opsti_podaci" // Default menu name
	}

	subMenus := common.GetSubMenus(domain.MenuData, menuName)
	if subMenus == nil {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}

	h.subMenuItems = subMenus
	h.menuItems.CurrentMenu = menuName
	year := fmt.Sprintf("%d", time.Now().Year())
	c := tmpl.Content(IsLoggedIn)

	fvrData, _, err := h.fvrService.GetAllFvr()
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

	// Get user session
	session, err := h.store.Get(r, "app-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Get values from session with defaults
	currentGod, _ := session.Values["poslovnagodina"].(int)
	currentKomintent, _ := session.Values["komintent"].(int)

	poslGodina, err := h.fvrService.GetAllGod(currentKomintent)
	if err != nil {
		http.Error(w, "Error while getting years", http.StatusInternalServerError)
		return
	}

	// Create combo configs with current values
	komintentConfig := setComboKomintent(fvrData, currentKomintent)
	poslGodConfig := setComboPoslGodConfig(poslGodina, currentGod)

	err = tmpl.Base(
		IsLoggedIn, c, h.menuItems, h.subMenuItems, "HELIA", username, year, "",
		komintentConfig,
		poslGodConfig).Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func getUsernameFromToken(r *http.Request) (string, bool) {
	// Get the auth_token cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return "", false // No cookie found, user is not authenticated
	}
	username, err := infrastructure.VerifyJWT(cookie.Value)

	if err != nil {
		return "", false // Invalid token, user is not authenticated
	}
	return username.Username, true // User is authenticated
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
	r.HandleFunc("/home", h.HomeHandler)
	r.HandleFunc("/", h.indexHandler)
	r.HandleFunc("api/get-current-date", h.getCurrentDate)
	r.HandleFunc("/api/selectposlgod", h.SelectPoslovnaGodina)
	r.HandleFunc("/api/setposlgod", h.SetPoslovnaGod)
}

func populateComboIntItems(poslGodina []int) []domain.ComboItem {
	poslGodComboItems := []domain.ComboItem{}
	for i := 0; i < len(poslGodina); i++ {
		poslGodComboItems = append(poslGodComboItems, domain.ComboItem{
			Key:   fmt.Sprintf("%d", poslGodina[i]),
			Value: fmt.Sprintf("%d", poslGodina[i]),
		})
	}
	return poslGodComboItems
}

func setComboPoslGodConfig(poslovnaGod []int, selectedValue int) domain.ComboFieldConfig {
	config := domain.ComboFieldConfig{
		ID:             "poslovnagodina",
		Name:           "poslovnagodina",
		LabelText:      "Poslovna Godina",
		HasLabel:       true,
		Disabled:       false,
		ClassSelect:    utils.ClassSelect + " min-w-[80px] ",
		ClassLabel:     utils.ClassLabel + " font-medium text-white text-sm whitespace-nowrap",
		HxVals:         `js:{"komintent": document.getElementById("komintent").value, "poslovnagodina": this.value}`,
		ChangeEndpoint: "/api/selectposlgod",
		HxSwap:         "outerHTML", // Change to outerHTML to replace the entire select
		HxChangeTarget: "body",      // Target the entire body to replace the whole page
	}

	optionItems := []domain.ComboItem{}
	for _, item := range poslovnaGod {
		key := fmt.Sprintf("%d", item) // Convert to string to match selectedValue
		optionItems = append(optionItems, domain.ComboItem{
			Key:   key,
			Value: key, // Use the number as both key and value
		})
	}
	config.OptionValues = optionItems
	config.SelectedValue = fmt.Sprintf("%d", selectedValue)
	return config
}

func setComboKomintent(items []domain.Fvr, selectedValue int) domain.ComboFieldConfig {
	config := domain.ComboFieldConfig{
		ID:             "komintent",
		Name:           "komintent",
		LabelText:      "Komintent",
		HasLabel:       true,
		Disabled:       false,
		ClassSelect:    utils.ClassSelect + " min-w-[80px] ",
		ClassLabel:     utils.ClassLabel + " font-medium text-white text-sm whitespace-nowrap ",
		HxVals:         `{"komintent": document.getElementById("komintent").value}`,
		ChangeEndpoint: "/api/setposlgod",
		HxSwap:         "outerHTML", // Change to outerHTML to replace the entire select
		HxChangeTarget: "body",      // Target the entire body to replace the whole page
	}

	optionItems := []domain.ComboItem{}
	for _, item := range items {
		key := fmt.Sprintf("%d", item.Kar) // Convert to string to match selectedValue
		optionItems = append(optionItems, domain.ComboItem{
			Key:   key,
			Value: item.Naziv,
		})
	}
	config.OptionValues = optionItems
	config.SelectedValue = fmt.Sprintf("%d", selectedValue)
	return config
}
