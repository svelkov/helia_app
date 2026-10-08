package config

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type Router struct {
	Router *http.ServeMux
	Config Config
	Logger *log.Logger
	Auth   Authenticator
}

// Config struct for passing configuration
type Config struct {
	BasePath        string        `json:"base_path"`
	DBConfig        DB_Connection `json:"db_connection"`
	PageSize        int           `json:"page_size"`
	PageSizes       []int         `json:"page_sizes"`
	Env             string        `json:"env"`
	Port            string        `json:"port"`
	Languages       []string      `json:"languages"`
	DefaultLanguage string        `json:"defaultLanguage"`
	JwtSecret       string        `json:"jwt_secret"`
	SessionSecret   string        `json:"session_secret"`
	NDuzSint        int           `json:"nDuzSint"`
	Konta           KontaConfig   `json:"konta"`
	// GrupeDokumenata are the groups of the vrste naloga and of the vrste dokumenata of the modules
	// (the legacy global variables gsDOKROB, gsDOKPRO, gsTIPDOKOSN, gsDOKULS and gsDOKULGP).
	GrupeDokumenata GrupeDokumenataConfig `json:"grupe_dokumenata"`

	// JWT Token Configuration (Option B - JWT-only authentication)
	AccessTokenTTL       int  `json:"access_token_ttl"`       // In minutes, default 15
	RefreshTokenTTL      int  `json:"refresh_token_ttl"`      // In minutes, default 1440 (24 hours)
	CSRFTokenTTL         int  `json:"csrf_token_ttl"`         // In minutes, default 1440
	EnableTokenBlacklist bool `json:"enable_token_blacklist"` // Enable token revocation, default true
}
type KontaConfig struct {
	KontoKupca      string `json:"konto_kupca"`
	KontoDobavljaca string `json:"konto_dobavljaca"`
}

// GrupeDokumenataConfig are the groups (tipdok.grpdok, dokvrsta.grpdok) of the documents of the modules
// and the vrste dokumenata of the ulazi, like the legacy global variables. An empty value of the
// configuration falls back to the legacy value (see the getters).
type GrupeDokumenataConfig struct {
	// DokRob are the groups of the robno knjigovodstvo (gsDOKROB "ROB;SVI;PST").
	DokRob []string `json:"dok_rob"`
	// DokPro are the groups of the proizvodnja (gsDOKPRO "PRO;SVI;PST").
	DokPro []string `json:"dok_pro"`
	// TipdokOsn are the groups of the vrste naloga of the osnovna sredstva (gsTIPDOKOSN "OSN;SVI;PST").
	TipdokOsn []string `json:"tipdok_osn"`
	// DokUls are the vrste dokumenata of the ulazi (gsDOKULS "198;199").
	DokUls []int `json:"dok_uls"`
	// DokUlgp are the vrste dokumenata of the ulazi (gsDOKULGP "202;203").
	DokUlgp []int `json:"dok_ulgp"`
	// TipmagRob are the tipovi of the magacini (magacini.tipmag) of the robno knjigovodstvo, the
	// magacini of the unos dokumenata (the legacy ROB_QRY_MAGUSER with "V;M;D").
	TipmagRob []string `json:"tipmag_rob"`
}

// GetDokRob returns the groups of the robno knjigovodstvo (default ROB, SVI, PST).
func (g GrupeDokumenataConfig) GetDokRob() []string {
	return orDefault(g.DokRob, []string{"ROB", "SVI", "PST"})
}

// GetDokPro returns the groups of the proizvodnja (default PRO, SVI, PST).
func (g GrupeDokumenataConfig) GetDokPro() []string {
	return orDefault(g.DokPro, []string{"PRO", "SVI", "PST"})
}

// GetTipdokOsn returns the groups of the vrste naloga of the osnovna sredstva (default OSN, SVI, PST).
func (g GrupeDokumenataConfig) GetTipdokOsn() []string {
	return orDefault(g.TipdokOsn, []string{"OSN", "SVI", "PST"})
}

// GetDokUls returns the vrste dokumenata of the ulazi (default 198, 199).
func (g GrupeDokumenataConfig) GetDokUls() []int {
	return orDefault(g.DokUls, []int{198, 199})
}

// GetDokUlgp returns the vrste dokumenata of the ulazi (default 202, 203).
func (g GrupeDokumenataConfig) GetDokUlgp() []int {
	return orDefault(g.DokUlgp, []int{202, 203})
}

// GetTipmagRob returns the tipovi of the magacini of the robno knjigovodstvo (default V, M, D).
func (g GrupeDokumenataConfig) GetTipmagRob() []string {
	return orDefault(g.TipmagRob, []string{"V", "M", "D"})
}

// orDefault returns the configured values, the default when none are configured.
func orDefault[T any](values, def []T) []T {
	if len(values) == 0 {
		return def
	}
	return values
}

type DB_Connection struct {
	DBHost       string `json:"db_host"`
	DBPort       int    `json:"db_port"`
	DBUser       string `json:"db_user"`
	DBPassword   string `json:"db_password"`
	DBName       string `json:"db_name"`
	DBSearchPath string `json:"db_search_path"`
}

func (c *Config) GetPageSize() int {
	if c.PageSize == 0 {
		return 20 // default page size
	}
	return c.PageSize
}

// NDuzSintForKnjigovod returns the length of the synthetic account prefix (nDuzSIN) for the
// bookkeeping type of a firma (FVR.KNJIGOVOD):
//
//	IF FVR.KNJIGOVOD = "Finansijsko" OR FVR.KNJIGOVOD = "Pogonsko" THEN nDuzSIN = 3
//	ELSE nDuzSIN = 4
//	END
//
// The comparison ignores case and surrounding spaces. NDuzSint from the configuration is NOT used
// here: it only serves as the fallback for requests that have no firma/bookkeeping type yet (see
// common.NDuzSint).
func (c Config) NDuzSintForKnjigovod(knjigovod string) int {
	switch strings.ToLower(strings.TrimSpace(knjigovod)) {
	case "finansijsko", "pogonsko":
		return 3
	default:
		return 4
	}
}

// GetAccessTokenTTL returns access token TTL with default fallback
func (c *Config) GetAccessTokenTTL() time.Duration {
	if c.AccessTokenTTL <= 0 {
		return 15 * time.Minute // default 15 minutes
	}
	return time.Duration(c.AccessTokenTTL) * time.Minute
}

// GetRefreshTokenTTL returns refresh token TTL with default fallback
func (c *Config) GetRefreshTokenTTL() time.Duration {
	if c.RefreshTokenTTL <= 0 {
		return 24 * time.Hour // default 24 hours
	}
	return time.Duration(c.RefreshTokenTTL) * time.Minute
}

// GetCSRFTokenTTL returns CSRF token TTL with default fallback
func (c *Config) GetCSRFTokenTTL() time.Duration {
	if c.CSRFTokenTTL <= 0 {
		return 24 * time.Hour // default 24 hours
	}
	return time.Duration(c.CSRFTokenTTL) * time.Minute
}

// IsTokenBlacklistEnabled returns whether token blacklist is enabled
func (c *Config) IsTokenBlacklistEnabled() bool {
	return c.EnableTokenBlacklist
}

// Authenticator interface for authentication
type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type App struct {
	Cfg *Config
	Ws  *Router
}
