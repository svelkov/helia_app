package komercijalno

import (
	"net/http"

	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	komsvc "helia/internal/service/komercijalno"

	"github.com/gin-gonic/gin"
)

// The "eOtpremnice" screen of the komercijalno (the legacy pr_eOTPREMNICA with its internal window IW_Otpremnice,
// opened from the menu "eOtpremnica"): the screen of the otpremnice (OtpremnicaHandler with these urls) with the
// controls of the e-otpremnica in the entry dialog.
const (
	eotpURL             = "/api/komercijalno/eotpremnice"
	eotpListaURL        = eotpURL + "/lista"
	eotpDialogURL       = eotpURL + "/dialog"
	eotpNoviURL         = eotpURL + "/novi"
	eotpSaveURL         = eotpURL + "/save"
	eotpBrisiURL        = eotpURL + "/brisi"
	eotpBrisiPotvrdaURL = eotpBrisiURL + "/potvrda"
	eotpPotvrdaURL      = eotpURL + "/potvrda"
	eotpPotvrdiURL      = eotpURL + "/potvrdi"
	eotpNalogURL        = eotpURL + "/nalog"
	eotpValutaURL       = eotpURL + "/valuta"
	eotpStavkaURL       = eotpURL + "/stavka"
	eotpStavkaSaveURL   = eotpStavkaURL + "/save"
	eotpPrikazStavkeURL = eotpURL + "/prikaz/stavke"
	eotpStampaURL       = eotpURL + "/stampa"
	// The JSON of the controls of the e-otpremnica: eotpNalogPodaciURL what a new one takes of its nalog za isporuku (GET
	// rdokid), eotpPartnerURL a partner of the kontni plan (GET konto and sifra: the prevoznik), eotpMestoURL the adresa of
	// a mesto isporuke of the primalac (GET konto, sifra and mi).
	eotpNalogPodaciURL = eotpURL + "/nalog-podaci"
	eotpPartnerURL     = eotpURL + "/partner"
	eotpMestoURL       = eotpURL + "/mesto-isporuke"

	// The collapsible controls of the e-otpremnica (the screenshots DET1, DET2, DET3 and DET5 of IW_Otpremnice; DET4 is
	// the control of the stavke of the otpremnica).
	eotpDet1PanelID = "kom-eotp-det1-panel"
	eotpDet2PanelID = "kom-eotp-det2-panel"
	eotpDet3PanelID = "kom-eotp-det3-panel"
	eotpDet5PanelID = "kom-eotp-det5-panel"

	eotpMsgNalogNePostoji   = "Niste izabrali narudžbenicu po koju radite otpremnicu!!!"
	eotpMsgPartnerNePostoji = "Ne postoji u kontnom planu."
)

// eotpPutanje are the urls of the "eOtpremnice" screen.
var eotpPutanje = otpPutanje{url: eotpURL, lista: eotpListaURL, dialog: eotpDialogURL, novi: eotpNoviURL, save: eotpSaveURL, brisi: eotpBrisiURL,
	brisiPotvrda: eotpBrisiPotvrdaURL, potvrda: eotpPotvrdaURL, potvrdi: eotpPotvrdiURL, nalog: eotpNalogURL, valuta: eotpValutaURL,
	stavka: eotpStavkaURL, stavkaSave: eotpStavkaSaveURL, prikazStavke: eotpPrikazStavkeURL, stampa: eotpStampaURL,
	eotp: true, eotpNalog: eotpNalogPodaciURL, partner: eotpPartnerURL, mestoIsp: eotpMestoURL, koraciDialog: eotpKoraciDialogURL}

// EotpremnicaHandler is the handler of the "eOtpremnice" screen: the handler of the otpremnice with its urls and the
// service of the e-otpremnica.
type EotpremnicaHandler struct {
	*OtpremnicaHandler
}

// NewEotpremnicaHandler creates the handler of the "eOtpremnice" screen.
func NewEotpremnicaHandler(s komsvc.EotpremnicaService, cfg config.Config, translator *i18n.Service) *EotpremnicaHandler {
	return &EotpremnicaHandler{OtpremnicaHandler: &OtpremnicaHandler{service: s, eotp: s, p: eotpPutanje, translator: translator, cfg: cfg}}
}

// EotpNalogPodaci returns what a new e-otpremnica takes of its nalog za isporuku (GET rdokid) as JSON: the primalac and
// the adresa odredišta of its mesto isporuke; found false without it.
func (h *EotpremnicaHandler) EotpNalogPodaci(c *gin.Context) {
	n, found, err := h.eotp.GetEotpNalog(c.Request.Context(), int64(common.StringToInt(c.Query("rdokid"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(eotpMsgNalogNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "nalog": n})
}

// EotpPartner returns a partner of the kontni plan (GET konto and sifra: the prevoznik, the legacy PrikaziPrevoznika)
// as JSON; found false without it.
func (h *EotpremnicaHandler) EotpPartner(c *gin.Context) {
	p, found, err := h.eotp.GetEotpPartner(c.Request.Context(), c.Query("konto"), c.Query("sifra"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusOK, gin.H{"found": false, "message": h.translator.Message(eotpMsgPartnerNePostoji)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": true, "partner": p})
}

// EotpMesto returns the adresa of a mesto isporuke of the primalac (GET konto, sifra and mi: the adresa odredišta, the
// legacy PrikaziMI) as JSON; found false without it (the legacy empties the adresa).
func (h *EotpremnicaHandler) EotpMesto(c *gin.Context) {
	a, found, err := h.eotp.GetEotpMesto(c.Request.Context(), c.Query("konto"), c.Query("sifra"), common.StringToInt(c.Query("mi")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"found": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"found": found, "adresa": a})
}

// AddRoutes registers the routes of the "eOtpremnice" screen: the ones of the otpremnice and the JSON of the controls
// of the e-otpremnica.
func (h *EotpremnicaHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/eotpremnice", h.Otpremnica)
	r.GET("/api/komercijalno/eotpremnice/lista", h.OtpLista)
	r.GET("/api/komercijalno/eotpremnice/dialog", h.OtpDialog)
	r.GET("/api/komercijalno/eotpremnice/novi", h.OtpNovi)
	r.POST("/api/komercijalno/eotpremnice/save", h.SaveOtp)
	r.GET("/api/komercijalno/eotpremnice/brisi/potvrda", h.PotvrdaBrisanja)
	r.DELETE("/api/komercijalno/eotpremnice/brisi/:id", h.DeleteOtp)
	r.GET("/api/komercijalno/eotpremnice/potvrda", h.PotvrdaOtp)
	r.POST("/api/komercijalno/eotpremnice/potvrdi/:id", h.PotvrdiOtp)
	r.GET("/api/komercijalno/eotpremnice/nalog", h.OtpNalog)
	r.GET("/api/komercijalno/eotpremnice/valuta", h.OtpValuta)
	r.GET("/api/komercijalno/eotpremnice/stavka", h.OtpStavka)
	r.POST("/api/komercijalno/eotpremnice/stavka/save", h.SaveOtpStavka)
	r.GET("/api/komercijalno/eotpremnice/prikaz/stavke/:id", h.OtpPrikazStavke)
	r.GET("/api/komercijalno/eotpremnice/stampa", h.OtpStampa)
	r.GET("/api/komercijalno/eotpremnice/nalog-podaci", h.EotpNalogPodaci)
	r.GET("/api/komercijalno/eotpremnice/partner", h.EotpPartner)
	r.GET("/api/komercijalno/eotpremnice/mesto-isporuke", h.EotpMesto)
}
