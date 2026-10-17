package komercijalno

import (
	"helia/config"
	"helia/i18n"
	komsvc "helia/internal/service/komercijalno"

	"github.com/gin-gonic/gin"
)

// The entry dialog of the e-otpremnica as a stepper (the test screen "Unos u koracima", opened by its button on the main
// screen of the e-otpremnice): the same e-otpremnica (OtpremnicaHandler with the service of the e-otpremnica and these
// urls), its header and its controls shown one step at a time instead of as collapsible controls. The main screen, the
// list and the deletion are the ones of the e-otpremnice.
const (
	eotpKoraciURL            = "/api/komercijalno/eotpremnice-koraci"
	eotpKoraciDialogURL      = eotpKoraciURL + "/dialog"
	eotpKoraciNoviURL        = eotpKoraciURL + "/novi"
	eotpKoraciSaveURL        = eotpKoraciURL + "/save"
	eotpKoraciPotvrdaURL     = eotpKoraciURL + "/potvrda"
	eotpKoraciPotvrdiURL     = eotpKoraciURL + "/potvrdi"
	eotpKoraciNalogURL       = eotpKoraciURL + "/nalog"
	eotpKoraciValutaURL      = eotpKoraciURL + "/valuta"
	eotpKoraciStavkaURL      = eotpKoraciURL + "/stavka"
	eotpKoraciStavkaSaveURL  = eotpKoraciStavkaURL + "/save"
	eotpKoraciStampaURL      = eotpKoraciURL + "/stampa"
	eotpKoraciNalogPodaciURL = eotpKoraciURL + "/nalog-podaci"
	eotpKoraciPartnerURL     = eotpKoraciURL + "/partner"
	eotpKoraciMestoURL       = eotpKoraciURL + "/mesto-isporuke"
)

// eotpKoraciPutanje are the urls of the stepper: its own ones for the dialog, the ones of the e-otpremnice for the main
// screen (the list, the deletion, the stavke of a row).
var eotpKoraciPutanje = otpPutanje{url: eotpURL, lista: eotpListaURL, dialog: eotpKoraciDialogURL, novi: eotpKoraciNoviURL,
	save: eotpKoraciSaveURL, brisi: eotpBrisiURL, brisiPotvrda: eotpBrisiPotvrdaURL, potvrda: eotpKoraciPotvrdaURL,
	potvrdi: eotpKoraciPotvrdiURL, nalog: eotpKoraciNalogURL, valuta: eotpKoraciValutaURL, stavka: eotpKoraciStavkaURL,
	stavkaSave: eotpKoraciStavkaSaveURL, prikazStavke: eotpPrikazStavkeURL, stampa: eotpKoraciStampaURL,
	eotp: true, eotpNalog: eotpKoraciNalogPodaciURL, partner: eotpKoraciPartnerURL, mestoIsp: eotpKoraciMestoURL, koraci: true}

// EotpremnicaKoraciHandler is the handler of the stepper of the e-otpremnica.
type EotpremnicaKoraciHandler struct {
	*EotpremnicaHandler
}

// NewEotpremnicaKoraciHandler creates the handler of the stepper of the e-otpremnica.
func NewEotpremnicaKoraciHandler(s komsvc.EotpremnicaService, cfg config.Config, translator *i18n.Service) *EotpremnicaKoraciHandler {
	return &EotpremnicaKoraciHandler{EotpremnicaHandler: &EotpremnicaHandler{
		OtpremnicaHandler: &OtpremnicaHandler{service: s, eotp: s, p: eotpKoraciPutanje, translator: translator, cfg: cfg}}}
}

// AddRoutes registers the routes of the stepper of the e-otpremnica: the ones of its entry dialog.
func (h *EotpremnicaKoraciHandler) AddRoutes(r *gin.Engine) {
	r.GET("/api/komercijalno/eotpremnice-koraci/dialog", h.OtpDialog)
	r.GET("/api/komercijalno/eotpremnice-koraci/novi", h.OtpNovi)
	r.POST("/api/komercijalno/eotpremnice-koraci/save", h.SaveOtp)
	r.GET("/api/komercijalno/eotpremnice-koraci/potvrda", h.PotvrdaOtp)
	r.POST("/api/komercijalno/eotpremnice-koraci/potvrdi/:id", h.PotvrdiOtp)
	r.GET("/api/komercijalno/eotpremnice-koraci/nalog", h.OtpNalog)
	r.GET("/api/komercijalno/eotpremnice-koraci/valuta", h.OtpValuta)
	r.GET("/api/komercijalno/eotpremnice-koraci/stavka", h.OtpStavka)
	r.POST("/api/komercijalno/eotpremnice-koraci/stavka/save", h.SaveOtpStavka)
	r.GET("/api/komercijalno/eotpremnice-koraci/stampa", h.OtpStampa)
	r.GET("/api/komercijalno/eotpremnice-koraci/nalog-podaci", h.EotpNalogPodaci)
	r.GET("/api/komercijalno/eotpremnice-koraci/partner", h.EotpPartner)
	r.GET("/api/komercijalno/eotpremnice-koraci/mesto-isporuke", h.EotpMesto)
}
