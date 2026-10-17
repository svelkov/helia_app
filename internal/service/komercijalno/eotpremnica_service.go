package komercijalno

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// The "eOtpremnice" screen of the komercijalno (the legacy pr_eOTPREMNICA with its internal window IW_Otpremnice): the
// otpremnica of the "Otpremnice" screen (OtpremnicaResource: the header, the stavke, the potvrda, the print, the list)
// with the data of the e-otpremnica, saved with the header in its transaction. The legacy only shows the controls of
// IW_Otpremnice (their save is not in its code); the columns are of migrations/003. The fields marked "*" in the legacy
// window (obavezna pri slanju) are not required: the e-otpremnica is not sent yet.

// The messages of the e-otpremnica.
const (
	komMsgEotpTip          = "Tip otpremnice nije ispravan."
	komMsgEotpNacin        = "Način otpreme nije ispravan."
	komMsgEotpDatum        = "Datum i vreme nisu ispravni."
	komMsgEotpPartner      = "Ne postoji u kontnom planu."
	komMsgEotpFazaPrazna   = "Unesite način otpreme ili naziv faze vožnje."
	komMsgEotpOpasnaPrazna = "Unesite naziv opasne materije."
)

// komEotpJediniceMase and komEotpJediniceZapremine are the jedinice mere of the težina and of the zapremina isporuke
// (the legacy COMBO_JM1 and COMBO_JM2; "" is "-").
var (
	komEotpJediniceMase      = []string{"", "g", "kg", "t"}
	komEotpJediniceZapremine = []string{"", "m3", "l"}
)

// EotpremnicaService is the "eOtpremnice" screen: the otpremnica with the data of the e-otpremnica.
type EotpremnicaService interface {
	OtpremnicaService
	ValidateEotpremnica(ctx context.Context, params domain.KomOtpParams, e domain.KomEotpParams) []domain.FieldError
	SaveEotpremnica(ctx context.Context, params domain.KomOtpParams, e domain.KomEotpParams) (int64, error)
	GetEotpNalog(ctx context.Context, nalogID int64) (domain.KomEotpNalogPodaci, bool, error)
	GetEotpPartner(ctx context.Context, konto, sifra string) (domain.KomEotpPartner, bool, error)
	GetEotpMesto(ctx context.Context, konto, sifra string, mi int) (domain.KomEotpAdresa, bool, error)
}

// eotpRepository is what the e-otpremnica needs of repository.KomercijalnoRepository.
type eotpRepository interface {
	KomEotp(ctx context.Context, rdokID int64) (domain.KomEotpPodaci, bool, error)
	KomEotpIzmeni(ctx context.Context, tx db.Transaction, rdokID int64, u domain.KomEotpUpis) error
	KomEotpFaze(ctx context.Context, rdokID int64) ([]domain.KomEotpFaza, error)
	KomEotpOpasne(ctx context.Context, rdokID int64) ([]domain.KomEotpOpasna, error)
	KomEotpBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomEotpFazaDodaj(ctx context.Context, tx db.Transaction, rdokID int64, f domain.KomEotpFaza) error
	KomEotpOpasnaDodaj(ctx context.Context, tx db.Transaction, rdokID int64, o domain.KomEotpOpasna) error
	KomEotpPartner(ctx context.Context, konto, sifra string) (domain.KomEotpPartner, bool, error)
	KomEotpPosiljalac(ctx context.Context, firma string) (domain.KomEotpPartner, error)
	KomEotpMagacinAdresa(ctx context.Context, magaciniID int) (domain.KomEotpAdresa, bool, error)
	KomEotpMestoIsporuke(ctx context.Context, konto, sifra string, mi int) (domain.KomEotpAdresa, bool, error)
	KomEotpNalog(ctx context.Context, nalogID int64) (fkto, fana string, mi int, found bool, err error)
}

// EotpremnicaResource implements EotpremnicaService on top of the otpremnica.
type EotpremnicaResource struct {
	*OtpremnicaResource
	erepo eotpRepository
}

// NewEotpremnicaService returns the service of the "eOtpremnice" screen.
func NewEotpremnicaService(otp *OtpremnicaResource, repo eotpRepository) *EotpremnicaResource {
	return &EotpremnicaResource{OtpremnicaResource: otp, erepo: repo}
}

// GetOtpremnica returns the content of the entry dialog of the otpremnica (OtpremnicaResource.GetOtpremnica) with its
// e-otpremnica: of a saved one its data, faze vožnje and opasne materije; of a new one the tip "Eksterna" and the
// adresa otpreme of its magacin (the legacy initialization of IW_Otpremnice). The pošiljalac is the firm, the primalac
// the kupac of the otpremnica and the prevoznik the one of its način otpreme (the legacy COMBO_NACINOTP: the firm, the
// prevoznik of the kontni plan, the primalac).
func (s *EotpremnicaResource) GetOtpremnica(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomOtpDokument, error) {
	dok, err := s.OtpremnicaResource.GetOtpremnica(ctx, rdokID, vrd, magaciniID)
	if err != nil {
		return dok, err
	}
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return dok, fmt.Errorf("no user session found")
	}
	h := dok.Header
	e := &domain.KomEotpView{}
	if h.RdokID > 0 {
		podaci, _, err := s.erepo.KomEotp(ctx, h.RdokID)
		if err != nil {
			return dok, err
		}
		e.KomEotpPodaci = podaci
		if e.Faze, err = s.erepo.KomEotpFaze(ctx, h.RdokID); err != nil {
			return dok, err
		}
		if e.Opasne, err = s.erepo.KomEotpOpasne(ctx, h.RdokID); err != nil {
			return dok, err
		}
	} else if h.MagaciniID > 0 {
		adresa, _, err := s.erepo.KomEotpMagacinAdresa(ctx, h.MagaciniID)
		if err != nil {
			return dok, err
		}
		e.UlicaOtp, e.MestoOtp, e.PosBrOtp = adresa.Ulica, adresa.Mesto, adresa.Pobro
	}
	if e.TipEotp == "" {
		e.TipEotp = domain.KomEotpEksterna
	}
	if e.Posiljalac, err = s.erepo.KomEotpPosiljalac(ctx, session.Firma); err != nil {
		return dok, err
	}
	if strings.TrimSpace(h.Fkto) != "" {
		if e.Primalac, _, err = s.erepo.KomEotpPartner(ctx, h.Fkto, h.Fana); err != nil {
			return dok, err
		}
	}
	if e.Prevoznik, err = s.prevoznik(ctx, e); err != nil {
		return dok, err
	}
	dok.Eotp = e
	return dok, nil
}

// prevoznik returns the prevoznik of the način otpreme of the e-otpremnica (the legacy select of COMBO_NACINOTP): the
// firm for the sopstveni prevoz, the prevoznik of the kontni plan, the primalac for the prevoz primaoca; none for the
// lično preuzimanje or dostava.
func (s *EotpremnicaResource) prevoznik(ctx context.Context, e *domain.KomEotpView) (domain.KomEotpPartner, error) {
	switch e.NacOtp {
	case domain.KomEotpSopstveniPrevoz:
		return e.Posiljalac, nil
	case domain.KomEotpPrevozPrimaoca:
		return e.Primalac, nil
	case domain.KomEotpPrevoznik:
		if strings.TrimSpace(e.PrevKto) == "" {
			return domain.KomEotpPartner{}, nil
		}
		p, _, err := s.erepo.KomEotpPartner(ctx, e.PrevKto, e.PrevAna)
		return p, err
	}
	return domain.KomEotpPartner{}, nil
}

// komEotpDatumVreme parses a datum i vreme of the e-otpremnica (a datetime-local field, or dd.mm.yyyy hh:mm); empty is
// no datum (ok), a wrong one is not ok.
func komEotpDatumVreme(value string) (sql.NullTime, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullTime{}, true
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05", "02.01.2006 15:04", "2006-01-02"} {
		if t, err := time.Parse(layout, value); err == nil {
			return sql.NullTime{Time: t, Valid: true}, true
		}
	}
	return sql.NullTime{}, false
}

// komEotpU tells whether the value is one of the list.
func komEotpU(value string, lista []string) bool {
	for _, v := range lista {
		if v == value {
			return true
		}
	}
	return false
}

// komEotpJM is the jedinica mere of the form ("-" is none).
func komEotpJM(value string) string {
	value = strings.TrimSpace(value)
	if value == "-" {
		return ""
	}
	return value
}

// komEotpRed returns the value of the column of the row i of a grid ("" without it).
func komEotpRed(kolona []string, i int) string {
	if i < len(kolona) {
		return strings.TrimSpace(kolona[i])
	}
	return ""
}

// komEotpFaze returns the faze vožnje of the rows of their grid (the empty rows left out) and the index of the first
// row that is not complete (-1 without one).
func komEotpFaze(e domain.KomEotpParams) ([]domain.KomEotpFaza, int) {
	var faze []domain.KomEotpFaza
	for i := range e.FazaNaziv {
		nacin, naziv := common.StringToInt(komEotpRed(e.FazaNacOtp, i)), komEotpRed(e.FazaNaziv, i)
		if nacin == 0 && naziv == "" {
			continue
		}
		if nacin < domain.KomEotpSopstveniPrevoz || nacin > domain.KomEotpLicnaDostava || naziv == "" {
			return faze, i
		}
		faze = append(faze, domain.KomEotpFaza{Rbr: len(faze) + 1, NacOtp: nacin, Naziv: naziv})
	}
	return faze, -1
}

// komEotpOpasne returns the opasne materije of the rows of their grid (the empty rows left out) and the index of the
// first row without a naziv (-1 without one).
func komEotpOpasne(e domain.KomEotpParams) ([]domain.KomEotpOpasna, int) {
	var opasne []domain.KomEotpOpasna
	for i := range e.OpmaNaziv {
		o := domain.KomEotpOpasna{Naziv: komEotpRed(e.OpmaNaziv, i), Opis: komEotpRed(e.OpmaOpis, i), Komentar: komEotpRed(e.OpmaKoment, i)}
		if o.Naziv == "" && o.Opis == "" && o.Komentar == "" {
			continue
		}
		if o.Naziv == "" {
			return opasne, i
		}
		o.Rbr = len(opasne) + 1
		opasne = append(opasne, o)
	}
	return opasne, -1
}

// ValidateEotpremnica validates the header of the otpremnica (ValidateOtpremnica) and its e-otpremnica: the tip, the
// jedinice mere and the način otpreme are ones of their lists, the datumi i vremena are valid, a given prevoznik exists
// in the kontni plan, every faza vožnje has its način otpreme and naziv, every opasna materija its naziv.
func (s *EotpremnicaResource) ValidateEotpremnica(ctx context.Context, params domain.KomOtpParams, e domain.KomEotpParams) []domain.FieldError {
	fieldErrors := s.ValidateOtpremnica(ctx, params)
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if !komEotpU(strings.TrimSpace(e.TipEotp), []string{domain.KomEotpEksterna, domain.KomEotpInterna}) {
		add("tipeotp", komMsgEotpTip)
	}
	if !komEotpU(komEotpJM(e.Jm1), komEotpJediniceMase) {
		add("jm1", common.ErrMsgObavezanPodatak)
	}
	if !komEotpU(komEotpJM(e.Jm2), komEotpJediniceZapremine) {
		add("jm2", common.ErrMsgObavezanPodatak)
	}
	if n := common.StringToInt(e.NacOtp); n < 0 || n > domain.KomEotpLicnaDostava {
		add("nacotp", komMsgEotpNacin)
	}
	for field, value := range map[string]string{"datpoc": e.Datpoc, "datzav": e.Datzav, "stvdatotp": e.StvDatotp} {
		if _, ok := komEotpDatumVreme(value); !ok {
			add(field, komMsgEotpDatum)
		}
	}
	if common.StringToInt(e.NacOtp) == domain.KomEotpPrevoznik && strings.TrimSpace(e.PrevKto) != "" {
		if _, found, err := s.erepo.KomEotpPartner(ctx, e.PrevKto, e.PrevAna); err != nil {
			add("prevana", common.ErrMsgReadData)
		} else if !found {
			add("prevana", komMsgEotpPartner)
		}
	}
	if _, i := komEotpFaze(e); i >= 0 {
		// The rows come from the line of "Dodaj" (it checks them), so the error goes to its naziv.
		add("eotpfazanaziv", komMsgEotpFazaPrazna)
	}
	if _, i := komEotpOpasne(e); i >= 0 {
		add("eotpopmanaziv", komMsgEotpOpasnaPrazna)
	}
	return fieldErrors
}

// komEotpUpis returns what the save writes of the e-otpremnica of the form (a validated one).
func komEotpUpis(e domain.KomEotpParams) domain.KomEotpUpis {
	t := strings.TrimSpace
	datpoc, _ := komEotpDatumVreme(e.Datpoc)
	datzav, _ := komEotpDatumVreme(e.Datzav)
	stvdat, _ := komEotpDatumVreme(e.StvDatotp)
	u := domain.KomEotpUpis{KomEotpPodaci: domain.KomEotpPodaci{
		OpasneMat: common.StringToBool(e.OpasneMat), Roba3Str: common.StringToBool(e.Roba3Str), TipEotp: t(e.TipEotp),
		Jm1: komEotpJM(e.Jm1), Jm2: komEotpJM(e.Jm2), Bttovol: komBroj(e.Bttovol, 0), Datpoc: datpoc, Datzav: datzav, StvDatotp: stvdat,
		PovratRobe: common.StringToBool(e.PovratRobe), BrNarudz: t(e.BrNarudz), BrOkvSpor: t(e.BrOkvSpor), KontIme: t(e.KontIme),
		KontTel: t(e.KontTel), KontMail: t(e.KontMail), ReferKup: t(e.ReferKup), TelKup: t(e.TelKup), MailKup: t(e.MailKup),
		NacOtp: common.StringToInt(e.NacOtp), LicIme: t(e.LicIme), LicPrez: t(e.LicPrez), LicBrLk: t(e.LicBrLk), MestoUtov: t(e.MestoUtov),
		MestoIsp: t(e.MestoIsp), VozPrez: t(e.VozPrez), VozMail: t(e.VozMail), VozTel: t(e.VozTel), VozDozv: t(e.VozDozv),
		UlicaOtp: t(e.UlicaOtp), UlBrOtp: t(e.UlBrOtp), MestoOtp: t(e.MestoOtp), PosBrOtp: t(e.PosBrOtp), SifObjOtp: t(e.SifObjOtp),
		UlicaOdr: t(e.UlicaOdr), UlBrOdr: t(e.UlBrOdr), MestoOdr: t(e.MestoOdr), PosBrOdr: t(e.PosBrOdr), SifObjOdr: t(e.SifObjOdr),
		FisNap: t(e.FisNap), Roba3StrTxt: t(e.Roba3StrTxt)},
		Mi: common.StringToInt(e.Mi)}
	// The prevoznik of the kontni plan only for the način otpreme "Prevoznik" (the legacy shows its konto and šifra only
	// then); the lično preuzimanje / dostava only for theirs.
	if u.NacOtp == domain.KomEotpPrevoznik {
		u.PrevKto, u.PrevAna = t(e.PrevKto), t(e.PrevAna)
	}
	if u.NacOtp != domain.KomEotpLicnoPreuzimanje && u.NacOtp != domain.KomEotpLicnaDostava {
		u.LicIme, u.LicPrez, u.LicBrLk = "", "", ""
	}
	if !u.Roba3Str {
		u.Roba3StrTxt = ""
	}
	u.Faze, _ = komEotpFaze(e)
	if u.OpasneMat {
		u.Opasne, _ = komEotpOpasne(e)
	}
	return u
}

// SaveEotpremnica saves the header of a validated otpremnica (OtpremnicaResource.SaveOtpremnica) and in its transaction
// its e-otpremnica: its data and its mesto isporuke, its faze vožnje and opasne materije written again (the opasne
// materije only with "Opasne materije"). It returns the rdokid.
func (s *EotpremnicaResource) SaveEotpremnica(ctx context.Context, params domain.KomOtpParams, e domain.KomEotpParams) (int64, error) {
	u := komEotpUpis(e)
	return s.saveOtpremnica(ctx, params, func(tx db.Transaction, rdokID int64) error {
		if err := s.erepo.KomEotpIzmeni(ctx, tx, rdokID, u); err != nil {
			return err
		}
		if err := s.erepo.KomEotpBrisi(ctx, tx, rdokID); err != nil {
			return err
		}
		for _, f := range u.Faze {
			if err := s.erepo.KomEotpFazaDodaj(ctx, tx, rdokID, f); err != nil {
				return err
			}
		}
		for _, o := range u.Opasne {
			if err := s.erepo.KomEotpOpasnaDodaj(ctx, tx, rdokID, o); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteOtpremnica deletes a new otpremnica (OtpremnicaResource.DeleteOtpremnica) with its faze vožnje and opasne
// materije in its transaction.
func (s *EotpremnicaResource) DeleteOtpremnica(ctx context.Context, rdokID int64) (string, error) {
	return s.deleteOtpremnica(ctx, rdokID, func(tx db.Transaction) error {
		return s.erepo.KomEotpBrisi(ctx, tx, rdokID)
	})
}

// GetEotpNalog returns what a new e-otpremnica takes of its nalog za isporuku: the primalac (its kupac) and the adresa
// odredišta of its mesto isporuke; found is false without the nalog.
func (s *EotpremnicaResource) GetEotpNalog(ctx context.Context, nalogID int64) (domain.KomEotpNalogPodaci, bool, error) {
	var n domain.KomEotpNalogPodaci
	fkto, fana, mi, found, err := s.erepo.KomEotpNalog(ctx, nalogID)
	if err != nil || !found {
		return n, found, err
	}
	n.Fkto, n.Fana, n.Mi = fkto, fana, mi
	if n.Primalac, _, err = s.erepo.KomEotpPartner(ctx, fkto, fana); err != nil {
		return n, true, err
	}
	if mi > 0 {
		if n.Odrediste, _, err = s.erepo.KomEotpMestoIsporuke(ctx, fkto, fana, mi); err != nil {
			return n, true, err
		}
	}
	return n, true, nil
}

// GetEotpPartner returns a partner of the konto and šifra (the prevoznik of the kontni plan, the legacy
// PrikaziPrevoznika); found is false without it.
func (s *EotpremnicaResource) GetEotpPartner(ctx context.Context, konto, sifra string) (domain.KomEotpPartner, bool, error) {
	return s.erepo.KomEotpPartner(ctx, konto, sifra)
}

// GetEotpMesto returns the adresa of a mesto isporuke of the primalac (the adresa odredišta, the legacy PrikaziMI);
// found is false without it.
func (s *EotpremnicaResource) GetEotpMesto(ctx context.Context, konto, sifra string, mi int) (domain.KomEotpAdresa, bool, error) {
	return s.erepo.KomEotpMestoIsporuke(ctx, konto, sifra, mi)
}
