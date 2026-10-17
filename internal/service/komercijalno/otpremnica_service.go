package komercijalno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
	commonsvc "helia/internal/service/common"
	robnosvc "helia/internal/service/robno"
)

// The messages of the otpremnica (the legacy pr_Otpremnica); the ones it shares with the other documents are the komMsg
// constants of the other services.
const (
	komMsgOtpPotvrdjena        = "Dokument je potvrđen... Izmene nisu dozvoljene!!!"
	komMsgOtpPotvrdjenaBrisi   = "Dokument je potvrđen, brisanje nije dozvoljeno!!!"
	komMsgOtpNalog             = "Niste izabrali narudžbenicu po koju radite otpremnicu!!!"
	komMsgOtpNalogBezStavki    = "Porudžbenica/nalog za otpremu ne sadrži nijednu stavku. Nedozvoljeno formiranje otpremnice!!!"
	komMsgOtpNalogCena         = "Postoje stavke u dokumentu čija je cena 0!!! Nedozvoljeno snimanje dokumenta!!!"
	komMsgOtpBrotp             = "Morate uneti broj otpremnice..."
	komMsgOtpBrotpPostoji      = "Otpremnica sa ovim brojem već postoji..."
	komMsgOtpBrFakture         = "Morate uneti broj fakture..."
	komMsgOtpBrFakturePostoji  = "Dokument sa brojem: {{broj}} već postoji!!!!"
	komMsgOtpFakturaPostoji    = "Dokument sa ovim brojem fakture već postoji!!!!"
	komMsgOtpDatotp            = "Morate uneti datum otpremnice..."
	komMsgOtpDatotpNalog       = "Datum otpremnice ne može biti manji od datuma porudžbenice/naloga!!!"
	komMsgOtpDatum             = "Morate uneti datum dokumenta..."
	komMsgOtpDatumGodina       = "Nekorektan datum dokumenta, godina mora biti jednaka poslovnoj godini!"
	komMsgOtpDatumOtp          = "Datum fakture ne može biti manji od datuma otpremnice!!!"
	komMsgOtpMagacinProizvodi  = "Fakture za gotove proizvode i sirovine izdaju se samo iz magacina got. proizvoda i sirovina!!!"
	komMsgOtpMagacinRoba       = "Fakture za robu izdaju se iz magacina robe!!!"
	komMsgOtpPopdv             = "Morate uneti POPDV polje!"
	komMsgOtpPopdvNePostoji    = "POPDV polje ne postoji!"
	komMsgOtpPopdvVrsta        = "Uneto polje nije u skladu sa vrstom dokumenta!"
	komMsgOtpPopdvTip          = "Neodgovarajuće POPDV polje!"
	komMsgOtpKurs              = "Kurs mora biti veći od 0..."
	komMsgOtpVaziZa            = "Vrednost za koju kurs važi mora biti veća od 0..."
	komMsgOtpKategorija        = "Mora postojati oznaka poreske kategorije!"
	komMsgOtpNastanak          = "Morate odrediti kada nastaje PDV obaveza!"
	komMsgOtpGodinaZatvorena   = "Poslovna godina je zatvorena, knjiženje nije dozvoljeno..."
	komMsgOtpJm                = "Šifra ima nepostojeću jedinicu mere!"
	komMsgOtpStavka            = "Šifra: {{sifra}} {{naziv}}"
	komMsgOtpNulaKolicina      = "Ukupna isporučena količina za dokument je 0, dokument ne može biti potvrđen!!!"
	komMsgOtpCenaNula          = "Postoje stavke u dokumentu čija cena je 0!!! Dokument ne može biti potvrđen!!!"
	komMsgOtpZalihe            = "Nema dovoljno količine na zalihama, dokument ne može biti potvrđen!!!"
	komMsgOtpZalihaArtikla     = "Za artikal: {{sifra}} {{naziv}} zahtevano je: {{kolicina}} {{jm}}, raspoloživa zaliha: {{zaliha}} {{jm}}"
	komMsgOtpZalihaNePostoji   = "Artikal: {{sifra}} {{naziv}} na kontu: {{konto}} ne postoji u zalihama!"
	komMsgOtpStorno            = "Izabrani artikal nema nijedan ulaz, ne može biti storniran, vraćen na stanje!!!"
	komMsgOtpStornoNegativna   = "Kod storno dokumenata količina mora biti negativna!!!"
	komMsgOtpStornoVeca        = "Stornirana količina ne može biti veća od tražene količine!!!"
	komMsgOtpIsporucenaVeca    = "Isporučena količina ne može biti veća od tražene količine!!!"
	komMsgOtpIsporucenaNegativ = "Isporučena količina ne može biti negativna!!!"
	komMsgOtpStavkaNePostoji   = "Stavka koju birate za izmenu ne postoji."
)

// komOtpTipoviMagacina are the tipovi of the magacini of an otpremnica (the legacy ROB_QRY_MAGUSER with "G;V;S").
var komOtpTipoviMagacina = []string{"G", "V", "S"}

// OtpremnicaService is the "Otpremnice" screen of the komercijalno (the legacy pr_Otpremnica).
type OtpremnicaService interface {
	GetOtpremnica(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomOtpDokument, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	GetNalogOtpremnice(ctx context.Context, nalogID int64, magaciniID int) (domain.KomOtpNalog, bool, error)
	GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error)
	ValidateOtpremnica(ctx context.Context, params domain.KomOtpParams) []domain.FieldError
	SaveOtpremnica(ctx context.Context, params domain.KomOtpParams) (int64, error)
	DeleteOtpremnica(ctx context.Context, rdokID int64) (string, error)
	ProveriPotvrdu(ctx context.Context, params domain.KomOtpParams) (KomOtpProvera, error)
	PotvrdiOtpremnicu(ctx context.Context, params domain.KomOtpParams) (KomOtpProvera, error)
	GetStavke(ctx context.Context, rdokID int64) ([]domain.KomOtpStavkaRow, error)
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomOtpStavkaForma, bool, error)
	ValidateStavka(ctx context.Context, params domain.KomOtpStavkaParams) []domain.FieldError
	SaveStavka(ctx context.Context, params domain.KomOtpStavkaParams) error
	GetPrikaz(ctx context.Context, params domain.KomOtpPrikazParams) ([]domain.KomOtpPrikazRow, error)
	GetStampa(ctx context.Context, rdokID int64, komercijalniOpis bool) (domain.KomOtpStampa, string, error)
	StavkeTableFields(lot bool) []domain.Fields
	PrikazTableFields() []domain.Fields
	PrikazStavkeTableFields() []domain.Fields
}

// otpRepository is what the service needs of repository.KomercijalnoRepository.
type otpRepository interface {
	komValutaRepository
	komRezervacijaRepository
	BeginTx() (db.Transaction, error)
	KomOtpVrste(ctx context.Context) ([]domain.KomOtpVrsta, error)
	KomOtpremnica(ctx context.Context, rdokID int64) (domain.KomOtpremnicaRow, bool, error)
	KomOtpNalog(ctx context.Context, nalogID int64) (domain.KomOtpNalog, bool, error)
	KomOtpStavke(ctx context.Context, rdokID int64) ([]domain.KomOtpStavkaRow, error)
	KomOtpPrivremeniBrojevi(ctx context.Context, tx db.Transaction, vrd, mag int) (brotp, dokum int, err error)
	KomOtpBrotpPostoji(ctx context.Context, vrd, mag, brotp int, rdokID int64) (bool, error)
	KomOtpRdokPostoji(ctx context.Context, tx db.Transaction, mag, vrd, dokum int) (bool, error)
	KomOtpDodaj(ctx context.Context, tx db.Transaction, p domain.KomOtpUpis) (int64, error)
	KomOtpIzmeni(ctx context.Context, tx db.Transaction, p domain.KomOtpUpis) error
	KomOtpStavkeKopiraj(ctx context.Context, tx db.Transaction, otpID, nalogID int64, storno, preracun bool) error
	KomOtpStavkePreracunaj(ctx context.Context, tx db.Transaction, otpID int64) error
	KomOtpProdvred(ctx context.Context, tx db.Transaction, otpID int64, prodvred float64) error
	KomOtpStavkaIzmeni(ctx context.Context, tx db.Transaction, otpID, rproID int64, kolic float64, otk, serija string, roktr int64) error
	KomOtpNalogFormiran(ctx context.Context, tx db.Transaction, nalogID int64) (bool, error)
	KomOtpNalogVrati(ctx context.Context, tx db.Transaction, nalogID int64) (bool, error)
	KomOtpPrikaz(ctx context.Context, p domain.KomOtpPrikazParams) ([]domain.KomOtpPrikazRow, error)
	KomOtpGodinaZatvorena(ctx context.Context) (bool, error)
	KomOtpNoviBrojevi(ctx context.Context, tx db.Transaction, vrd, mag int) (brotp, dokum int, err error)
	KomOtpRstaProsnc(ctx context.Context, tx db.Transaction, magaciniID, sifra int) (float64, error)
	KomOtpNetoTezina(ctx context.Context, tx db.Transaction, otpID int64, tezina float64) error
	KomOtpRnal(ctx context.Context, tx db.Transaction, tipdok string, danal sql.NullTime, magaciniID int) (int64, bool, error)
	KomOtpRnalDodaj(ctx context.Context, tx db.Transaction, n domain.KomOtpRnalUpis) (int64, error)
	KomOtpRnalDokument(ctx context.Context, tx db.Transaction, rnalID int64) error
	KomOtpRdokDodaj(ctx context.Context, tx db.Transaction, otpID, rnalID int64, d domain.KomOtpRdokUpis) (int64, error)
	KomOtpRproDodaj(ctx context.Context, tx db.Transaction, p domain.KomOtpRproUpis) (int64, error)
	KomOtpRceneDodaj(ctx context.Context, tx db.Transaction, c domain.KomOtpRceneUpis) error
	KomOtpRdokBrst(ctx context.Context, tx db.Transaction, rdokID int64, brst int) error
	KomOtpAvansi(ctx context.Context, tx db.Transaction, nalogID, rdokID int64) (int64, error)
	KomOtpPotvrdi(ctx context.Context, tx db.Transaction, otpID int64, dokum, brotp int, opis string) (bool, error)
	KomOtpNalogIsporuceno(ctx context.Context, tx db.Transaction, otpID int64) error
	KomProKupac(ctx context.Context, konto, sifra string) (domain.KomProKupac, bool, error)
	KomMagacin(ctx context.Context, magaciniID int) (int, bool, error)
	KomProMagacin(ctx context.Context, magaciniID int) (tipmag string, tipzal, tipcene int, found bool, err error)
	KomZahtevBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomPorIDFkpl(ctx context.Context, konto, sifra string) (int64, error)
	KomPonudaStampaDobavljac(ctx context.Context, idFkpl int64) (domain.KomPonudaStampaDobavljac, bool, error)
	KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error)
	BankeZiroRacuni(ctx context.Context) ([]string, error)
	RporStopa(ctx context.Context, tx db.Transaction, po int, datum time.Time) (float64, error)
	DrstaLotStanje(ctx context.Context, konto string, sifra, mag int, otk, serija string, roktr int64) (repository.DrstaLot, bool, error)
	PopdvPoljeUlaz(ctx context.Context, polje string) (repository.PopdvUlaz, bool, error)
}

// otpRobno is what the potvrda needs of the robno service: the posting of the stanje of a stavka (by konto and by lot)
// and of the robni nalog, the obračun of the faktura and the nabavna cena of an artikal.
type otpRobno interface {
	AzurirajRSTA(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja) error
	AzurirajDRSTA(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja) error
	AzurirajRNAL(ctx context.Context, tx db.Transaction, rnalID int64, dugovno, potrazno bool, iznos float64) error
	ObracunFakture(ctx context.Context, rdokID int64) error
	IzracunajNCENA(ctx context.Context, danal, dadok time.Time, magaciniID, sifra int) (float64, error)
}

// OtpremnicaResource implements OtpremnicaService.
type OtpremnicaResource struct {
	repo   otpRepository
	common magaciniKorisnika
	robno  otpRobno
	now    func() time.Time
}

// NewOtpremnicaService returns the service of the "Otpremnice" screen.
func NewOtpremnicaService(repo otpRepository, common magaciniKorisnika, robno otpRobno) *OtpremnicaResource {
	return &OtpremnicaResource{repo: repo, common: common, robno: robno, now: time.Now}
}

// GetMagacini returns the magacini of the user of the tipovi G, V and S.
func (s *OtpremnicaResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.common.GetMagacinKorisnikaComboValues(ctx, commonsvc.WithIn("magacini.tipmag", komOtpTipoviMagacina...))
}

// GetValuta returns the valuta of the šifra with its kurs on the datum (the legacy BTN_TraziVAL).
func (s *OtpremnicaResource) GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error) {
	return komValutaKurs(ctx, s.repo, sifval, datum)
}

// komOtpVrsta returns the vrsta of the vrd of the list; found is false without it.
func komOtpVrsta(vrste []domain.KomOtpVrsta, vrd int) (domain.KomOtpVrsta, bool) {
	for _, v := range vrste {
		if v.Vrd == vrd {
			return v, true
		}
	}
	return domain.KomOtpVrsta{}, false
}

// komOtpPrivremen tells whether a broj is a temporary one (above 990000) or none.
func komOtpPrivremen(broj int) bool {
	return broj <= 0 || broj > domain.KomOtpPrivremeniBroj
}

// GetOtpremnica returns the content of the entry dialog with the otpremnica of the id (the legacy PrikaziHDR) or, for the
// id 0, a new one (the legacy BTN_NoviDOK): of the magacin and the vrsta (the first ones without them), with the next
// temporary broj otpremnice and fakture, the datumi of today, the valuta of the firma with the kurs 1, the oznaka of the
// poreska kategorija ("S" for a PDV obveznik, else "SS"), the default POPDV polje of the vrsta and the nastanak PDV
// obaveze (35 for a PDV obveznik, else 0). A potvrđena one cannot change: a new one is given with the message why.
func (s *OtpremnicaResource) GetOtpremnica(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomOtpDokument, error) {
	var dok domain.KomOtpDokument
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return dok, fmt.Errorf("no user session found")
	}
	var err error
	if dok.Magacini, err = s.GetMagacini(ctx); err != nil {
		return dok, err
	}
	vrste, err := s.repo.KomOtpVrste(ctx)
	if err != nil {
		return dok, err
	}
	for _, v := range vrste {
		dok.Vrste = append(dok.Vrste, domain.ComboItem{Key: strconv.Itoa(v.Vrd), Value: fmt.Sprintf("%d - %s", v.Vrd, v.Opis)})
	}
	dok.ZiroRacuni = []domain.ComboItem{{Key: "-", Value: "-"}}
	racuni, err := s.repo.BankeZiroRacuni(ctx)
	if err != nil {
		return dok, err
	}
	for _, r := range racuni {
		dok.ZiroRacuni = append(dok.ZiroRacuni, domain.ComboItem{Key: r, Value: r})
	}
	if dok.SifValFirme, _, err = komValutaFirme(ctx, s.repo, session.Firma); err != nil {
		return dok, err
	}
	firma, err := s.repo.KomFirma(ctx, session.Firma)
	if err != nil {
		return dok, err
	}
	h := &dok.Header
	h.PdvObveznik = firma.Obv
	if rdokID > 0 {
		o, found, err := s.repo.KomOtpremnica(ctx, rdokID)
		if err != nil {
			return dok, err
		}
		switch {
		case !found:
			dok.Poruka = komMsgNePostoji
		case o.Doksts != domain.KomOtpNova:
			dok.Poruka, vrd, magaciniID = komMsgOtpPotvrdjena, o.Vrd, o.MagaciniID
		default:
			h.KomOtpremnicaRow = o
			h.Datum, h.DatumOtp = o.Dadok.Time, o.Datotp.Time
			h.Vrsta, _ = komOtpVrsta(vrste, o.Vrd)
			if h.Kupac, _, err = s.repo.KomProKupac(ctx, o.Fkto, o.Fana); err != nil {
				return dok, err
			}
			h.Kupac.SistemPdv = komSistemPdv(h.Kupac.TipPdv)
			h.Ino = h.Kupac.Ter == domain.KomOtpTeritorijaStrani
			if !h.Ino {
				// The legacy PodaciHDR: a domestic kupac is invoiced in the valuta of the firm.
				h.SifVal, h.Kurs, h.VaziZa = dok.SifValFirme, 1, 1
			}
			h.Snimljen = true
			dok.Nastanci, dok.Nastanak = robnosvc.RobFaktureNastanakPDV(h.Vrsta.Storno(), h.PdvObveznik, h.Vrsta.Dokozn)
			return dok, nil
		}
	}
	h.Datum, h.DatumOtp = s.now(), s.now()
	h.Dadok = sql.NullTime{Time: h.Datum, Valid: true}
	h.Datotp = h.Dadok
	h.MagaciniID = komIzListe(dok.Magacini, magaciniID)
	h.Vrd = komIzListe(dok.Vrste, vrd)
	h.Vrsta, _ = komOtpVrsta(vrste, h.Vrd)
	h.SifVal, h.Kurs, h.VaziZa = dok.SifValFirme, 1, 1
	h.Posuslnab = "-"
	h.Category = "SS"
	if h.PdvObveznik {
		h.Category = "S"
		if h.Vrsta.FlgPopdv {
			h.Polje = h.Vrsta.Polje
		}
	}
	dok.Nastanci, dok.Nastanak = robnosvc.RobFaktureNastanakPDV(h.Vrsta.Storno(), h.PdvObveznik, h.Vrsta.Dokozn)
	if h.MagaciniID > 0 && h.Vrd > 0 {
		mag, err := s.magacin(ctx, h.MagaciniID)
		if err != nil {
			return dok, err
		}
		h.Mag = mag
		if h.Brotp, h.Dokum, err = s.repo.KomOtpPrivremeniBrojevi(ctx, nil, h.Vrd, mag); err != nil {
			return dok, err
		}
	}
	return dok, nil
}

// magacin returns the oznaka (mag) of a magacin of the user, 0 for another one.
func (s *OtpremnicaResource) magacin(ctx context.Context, magaciniID int) (int, error) {
	magacini, err := s.GetMagacini(ctx)
	if err != nil || !komUListi(magacini, magaciniID) {
		return 0, err
	}
	mag, _, err := s.repo.KomMagacin(ctx, magaciniID)
	return mag, err
}

// GetNalogOtpremnice returns a potvrđen nalog za isporuku of the magacin with what the header of an otpremnica shows of
// it (the legacy BTN_TraziDOK with PodaciHDR): its kupac and whether it is a foreign one (a domestic kupac is invoiced
// in the valuta of the firm at the kurs 1); found is false for another one.
func (s *OtpremnicaResource) GetNalogOtpremnice(ctx context.Context, nalogID int64, magaciniID int) (domain.KomOtpNalog, bool, error) {
	if nalogID <= 0 {
		return domain.KomOtpNalog{}, false, nil
	}
	n, found, err := s.repo.KomOtpNalog(ctx, nalogID)
	if err != nil || !found || n.Doksts != domain.KomNalogPotvrdjen || n.MagaciniID != magaciniID {
		return domain.KomOtpNalog{}, false, err
	}
	n.Datum = komDatumText(n.Dadok)
	if n.Kupac, _, err = s.repo.KomProKupac(ctx, n.Fkto, n.Fana); err != nil {
		return n, false, err
	}
	n.Kupac.SistemPdv = komSistemPdv(n.Kupac.TipPdv)
	n.Ino = n.Kupac.Ter == domain.KomOtpTeritorijaStrani
	if !n.Ino {
		session := domain.GetSessionFromStdContext(ctx)
		if session == nil {
			return n, false, fmt.Errorf("no user session found")
		}
		sifval, oznval, err := komValutaFirme(ctx, s.repo, session.Firma)
		if err != nil {
			return n, false, err
		}
		n.SifVal, n.OznVal, n.Kurs, n.VaziZa = sifval, oznval, 1, 1
	}
	return n, true, nil
}

// ValidateOtpremnica validates the header of an otpremnica (the legacy ValidacijaEDOK and ValidacijaEPRO): a saved one is
// new; a new one has a magacin of the user, a vrsta of the otpremnice and a potvrđen nalog za isporuku of the magacin
// with stavke whose cene are not 0; the vrsta of the gotovi proizvodi (FPR) is issued from a magacin of the gotovi
// proizvodi or sirovine (G, S), the one of the roba (FAK, FRP) from a magacin of the roba (V, M); with "Dozvoli izmenu br.
// OTP i br. Fakture" the brojevi are given, the broj otpremnice is not one of another otpremnica and the broj fakture not
// one of a robni dokument; the datum otpremnice is valid and not before the datum of the nalog, the datum dokumenta is a
// valid datum of the year and not before the datum otpremnice; the POPDV polje of a PDV obveznik whose vrsta needs it;
// the kurs and "važi za" are greater than 0; the oznaka of the poreska kategorija is given (the nastanak PDV obaveze is
// checked by the potvrda, which uses it).
func (s *OtpremnicaResource) ValidateOtpremnica(ctx context.Context, params domain.KomOtpParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		add("datum", common.ErrMsgSessionNotFound)
		return errs
	}
	vrste, err := s.repo.KomOtpVrste(ctx)
	if err != nil {
		add("vrd", common.ErrMsgReadData)
		return errs
	}
	var vrsta domain.KomOtpVrsta
	var mag, magaciniID int
	var nalogDadok sql.NullTime
	if params.RdokID > 0 {
		o, found, err := s.repo.KomOtpremnica(ctx, params.RdokID)
		switch {
		case err != nil:
			add("datum", common.ErrMsgReadData)
			return errs
		case !found:
			add("datum", komMsgNePostoji)
			return errs
		case o.Doksts != domain.KomOtpNova:
			add("datum", komMsgOtpPotvrdjena)
			return errs
		}
		vrsta, _ = komOtpVrsta(vrste, o.Vrd)
		mag, magaciniID, nalogDadok = o.Mag, o.MagaciniID, o.NalogDadok
	} else {
		magaciniID = common.StringToInt(params.MagaciniID)
		if mag, err = s.magacin(ctx, magaciniID); err != nil {
			add("magaciniid", common.ErrMsgReadData)
		} else if mag == 0 {
			add("magaciniid", komMsgMagacin)
		}
		var found bool
		if vrsta, found = komOtpVrsta(vrste, common.StringToInt(params.Vrd)); !found {
			add("vrd", komMsgVrsta)
		}
		n, found, err := s.repo.KomOtpNalog(ctx, int64(common.StringToInt(params.NalogID)))
		switch {
		case err != nil:
			add("nalogid", common.ErrMsgReadData)
		case !found || n.Doksts != domain.KomNalogPotvrdjen || n.MagaciniID != magaciniID:
			add("nalogid", komMsgOtpNalog)
		default:
			nalogDadok = n.Dadok
			stavke, err := s.repo.KomOtpStavke(ctx, n.RdokID)
			switch {
			case err != nil:
				add("nalogid", common.ErrMsgReadData)
			case len(stavke) == 0:
				add("nalogid", komMsgOtpNalogBezStavki)
			default:
				for _, st := range stavke {
					if st.Fcena == 0 || st.Cena == 0 {
						add("nalogid", komMsgOtpNalogCena)
						break
					}
				}
			}
		}
	}
	if vrsta.Vrd > 0 && mag > 0 {
		if tipmag, _, _, found, err := s.repo.KomProMagacin(ctx, magaciniID); err != nil {
			add("magaciniid", common.ErrMsgReadData)
		} else if found {
			switch vrsta.Dokozn {
			case "FPR":
				if tipmag != "G" && tipmag != "S" {
					add("vrd", komMsgOtpMagacinProizvodi)
				}
			case "FAK", "FRP":
				if tipmag != "V" && tipmag != "M" {
					add("vrd", komMsgOtpMagacinRoba)
				}
			}
		}
	}
	if common.StringToBool(params.DozvoliBroj) && vrsta.Vrd > 0 && mag > 0 {
		brotp, dokum := common.StringToInt(params.Brotp), common.StringToInt(params.Dokum)
		if brotp <= 0 {
			add("brotp", komMsgOtpBrotp)
		} else if postoji, err := s.repo.KomOtpBrotpPostoji(ctx, vrsta.Vrd, mag, brotp, params.RdokID); err != nil {
			add("brotp", common.ErrMsgReadData)
		} else if postoji {
			add("brotp", komMsgOtpBrotpPostoji)
		}
		if dokum <= 0 {
			add("brfakture", komMsgOtpBrFakture)
		} else if postoji, err := s.repo.KomOtpRdokPostoji(ctx, nil, mag, vrsta.Vrd, dokum); err != nil {
			add("brfakture", common.ErrMsgReadData)
		} else if postoji {
			add("brfakture", strings.ReplaceAll(komMsgOtpBrFakturePostoji, "{{broj}}", strconv.Itoa(dokum)))
		}
	}
	datotp, okOtp := komDatum(params.Datotp)
	if !okOtp {
		add("datotp", komMsgOtpDatotp)
	} else if nalogDadok.Valid && datotp.Before(komDan(nalogDadok.Time)) {
		add("datotp", komMsgOtpDatotpNalog)
	}
	switch datum, ok := komDatum(params.Datum); {
	case !ok:
		add("datum", komMsgOtpDatum)
	case datum.Year() != session.SelectedGod:
		add("datum", komMsgOtpDatumGodina)
	case okOtp && datum.Before(datotp):
		add("datum", komMsgOtpDatumOtp)
	}
	if vrsta.FlgPopdv {
		if firma, err := s.repo.KomFirma(ctx, session.Firma); err != nil {
			add("popdv", common.ErrMsgReadData)
		} else if firma.Obv {
			s.validatePopdv(ctx, params.Polje, add)
		}
	}
	if komBroj(params.Kurs, 0) <= 0 {
		add("kurs", komMsgOtpKurs)
	}
	if komBroj(params.VaziZa, 0) <= 0 {
		add("vaziza", komMsgOtpVaziZa)
	}
	if strings.TrimSpace(params.Category) == "" {
		add("oznakakat", komMsgOtpKategorija)
	}
	return errs
}

// validatePopdv checks the POPDV polje (the legacy ValidacijaEDOK): given, existing, of an evidencija of the izlazni
// računi ("I") and of the tip 3.
func (s *OtpremnicaResource) validatePopdv(ctx context.Context, polje string, add func(field, message string)) {
	polje = strings.ReplaceAll(strings.TrimSpace(polje), ",", ".")
	if polje == "" {
		add("popdv", komMsgOtpPopdv)
		return
	}
	p, found, err := s.repo.PopdvPoljeUlaz(ctx, polje)
	switch {
	case err != nil:
		add("popdv", common.ErrMsgReadData)
	case !found:
		add("popdv", komMsgOtpPopdvNePostoji)
	case p.Vktip != "" && p.Vktip != "I":
		add("popdv", komMsgOtpPopdvVrsta)
	case p.Tip != 3:
		add("popdv", komMsgOtpPopdvTip)
	}
}

// komDan is the datum of a time (00:00).
func komDan(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// komOtpUpis returns the columns of the header of the params.
func komOtpUpis(params domain.KomOtpParams) domain.KomOtpUpis {
	datotp, _ := komDatum(params.Datotp)
	datum, _ := komDatum(params.Datum)
	return domain.KomOtpUpis{RdokID: params.RdokID, NalogID: int64(common.StringToInt(params.NalogID)), Vrd: common.StringToInt(params.Vrd),
		Datotp: datotp, Dadok: datum, Otp: strings.TrimSpace(params.Otp), Vozac: strings.TrimSpace(params.Vozac),
		Brvozila: strings.TrimSpace(params.Brvozila), Bttowght: komBroj(params.Bttowght, 0), Netwght: komBroj(params.Netwght, 0),
		Totboxs: common.StringToInt(params.Totboxs), Brpaleta: common.StringToInt(params.Brpaleta), Foot: strings.TrimSpace(params.Foot),
		Napomena: strings.TrimSpace(params.Napomena), SifVal: common.StringToInt(params.SifVal), Kurs: komBroj(params.Kurs, 1),
		VaziZa: int(komBroj(params.VaziZa, 1)), Posuslnab: strings.TrimSpace(params.Posuslnab),
		Polje: strings.ReplaceAll(strings.TrimSpace(params.Polje), ",", "."), Keykat: strings.TrimSpace(params.Keykat),
		Category: strings.TrimSpace(params.Category), Pornapomena: strings.TrimSpace(params.Pornapomena), Brugov: strings.TrimSpace(params.Brugov)}
}

// SaveOtpremnica saves the header of a validated otpremnica in one transaction (the legacy SnimiEDOK with PuniDETALJ):
//   - a new one of the potvrđen nalog (marked 40) in its magacin: the temporary brojevi (or the typed ones with "Dozvoli
//     izmenu br. OTP i br. Fakture"), the fields of the screen and the other ones of the nalog; its stavke are the ones of
//     the nalog with the isporučena količina the tražena one (negative for a storno vrsta);
//   - a saved one: its fields (its brojevi only with "Dozvoli izmenu"), the stavke of the nalog it does not have yet.
//
// A kurs or "važi za" other than the one of the nalog (of a saved one: its own) sets the prodajne cene of the stavke to
// the cena u valuti at the new kurs. Then the prodajna vrednost of the otpremnica is calculated again (the legacy
// PreracunFAKTURAPROIZVGLOBEDOK; the legacy only after a new kurs). It returns the rdokid.
func (s *OtpremnicaResource) SaveOtpremnica(ctx context.Context, params domain.KomOtpParams) (int64, error) {
	return s.saveOtpremnica(ctx, params, nil)
}

// saveOtpremnica is SaveOtpremnica; posle (when given) writes in its transaction what the screen saves with the header
// of the otpremnica (the data of the e-otpremnica), after the otpremnica and its stavke are written.
func (s *OtpremnicaResource) saveOtpremnica(ctx context.Context, params domain.KomOtpParams, posle func(tx db.Transaction, rdokID int64) error) (int64, error) {
	p := komOtpUpis(params)
	vrste, err := s.repo.KomOtpVrste(ctx)
	if err != nil {
		return 0, err
	}
	dozvoli := common.StringToBool(params.DozvoliBroj)
	var stariKurs float64
	var staraVaziZa int
	var nalogID int64
	if p.RdokID > 0 {
		o, found, err := s.repo.KomOtpremnica(ctx, p.RdokID)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("otpremnica %d ne postoji", p.RdokID)
		}
		p.Vrd, nalogID, stariKurs, staraVaziZa = o.Vrd, o.VrdokID, o.Kurs, o.VaziZa
		p.Brotp, p.Dokum = o.Brotp, o.Dokum
	} else {
		n, found, err := s.repo.KomOtpNalog(ctx, p.NalogID)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("nalog %d ne postoji", p.NalogID)
		}
		mag, err := s.magacin(ctx, n.MagaciniID)
		if err != nil {
			return 0, err
		}
		if p.Brotp, p.Dokum, err = s.repo.KomOtpPrivremeniBrojevi(ctx, nil, p.Vrd, mag); err != nil {
			return 0, err
		}
		nalogID, stariKurs, staraVaziZa = n.RdokID, n.Kurs, n.VaziZa
	}
	if dozvoli {
		p.Brotp, p.Dokum = common.StringToInt(params.Brotp), common.StringToInt(params.Dokum)
	}
	vrsta, _ := komOtpVrsta(vrste, p.Vrd)
	preracun := stariKurs != p.Kurs || staraVaziZa != p.VaziZa
	tx, err := s.repo.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if p.RdokID > 0 {
		if err := s.repo.KomOtpIzmeni(ctx, tx, p); err != nil {
			return 0, err
		}
		if preracun {
			if err := s.repo.KomOtpStavkePreracunaj(ctx, tx, p.RdokID); err != nil {
				return 0, err
			}
		}
	} else {
		ok, err := s.repo.KomOtpNalogFormiran(ctx, tx, nalogID)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, errors.New(komMsgOtpNalog)
		}
		if p.RdokID, err = s.repo.KomOtpDodaj(ctx, tx, p); err != nil {
			return 0, err
		}
	}
	if err := s.repo.KomOtpStavkeKopiraj(ctx, tx, p.RdokID, nalogID, vrsta.Storno(), preracun); err != nil {
		return 0, err
	}
	if posle != nil {
		if err := posle(tx, p.RdokID); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit otpremnica: %w", err)
	}
	return p.RdokID, s.prodvred(ctx, p.RdokID)
}

// prodvred calculates the prodajna vrednost of an otpremnica again and saves it (the legacy PreracunFAKTURAPROIZVGLOBEDOK).
func (s *OtpremnicaResource) prodvred(ctx context.Context, otpID int64) error {
	o, found, err := s.repo.KomOtpremnica(ctx, otpID)
	if err != nil || !found {
		return err
	}
	stavke, err := s.repo.KomOtpStavke(ctx, otpID)
	if err != nil {
		return err
	}
	stope := map[int]float64{}
	if o.Dadok.Valid {
		for _, st := range stavke {
			if _, ok := stope[st.Po]; ok {
				continue
			}
			if stope[st.Po], err = s.repo.RporStopa(ctx, nil, st.Po, o.Dadok.Time); err != nil {
				return err
			}
		}
	}
	return s.repo.KomOtpProdvred(ctx, nil, otpID, komOtpProdvred(stavke, o.Ugrabat, o.Pkase, o.Pornapomena, stope))
}

// komOtpProdvred is the prodajna vrednost of a document (the legacy PreracunFAKTURAPROIZVGLOBEDOK): the iznos (Σ tražena
// količina × prodajna cena) less the rabat of the stavke, the ugovoreni rabat and the kasa of the document (each on what
// is left, rounded to 2 decimals), plus the PDV of the stavke (on the osnovica of each stavka by the stopa of its tarifa,
// rounded; none with a napomena o poreskom oslobođenju).
func komOtpProdvred(stavke []domain.KomOtpStavkaRow, ugrabat, pkase float64, pornapomena string, stope map[int]float64) float64 {
	var iznos, rabat, pdv float64
	for _, st := range stavke {
		iznos += st.Trazkolic * st.Fcena
		rabat += st.Trazkolic * st.Fcena * st.Rab / 100
		bruto := komRound2(st.Trazkolic * st.Fcena)
		vrab := komRound2(bruto * st.Rab / 100)
		vugr := komRound2((bruto - vrab) * ugrabat / 100)
		vkasa := komRound2((bruto - vrab - vugr) * pkase / 100)
		osnovica := komRound2(bruto - vrab - vugr - vkasa)
		pdv += komRound2(osnovica * stope[st.Po] / 100)
	}
	if strings.TrimSpace(pornapomena) != "" {
		pdv = 0
	}
	iznos, rabat = komRound2(iznos), komRound2(rabat)
	ugr := komRound2((iznos - rabat) * ugrabat / 100)
	kasa := komRound2((iznos - rabat - ugr) * pkase / 100)
	return komRound2(iznos - rabat - kasa - ugr + pdv)
}

// DeleteOtpremnica deletes a new otpremnica with its stavke (the legacy BTN_Brisanje): not a posted (knjige 1 "D") or a
// potvrđena one, whose message is returned instead. In one transaction its nalog za isporuku becomes potvrđen again
// (doksts 20; the legacy sets 10); the reservations stay with the nalog.
func (s *OtpremnicaResource) DeleteOtpremnica(ctx context.Context, rdokID int64) (string, error) {
	return s.deleteOtpremnica(ctx, rdokID, nil)
}

// deleteOtpremnica is DeleteOtpremnica; pre (when given) deletes in its transaction, before the otpremnica, what the
// screen keeps with it (the faze vožnje and the opasne materije of the e-otpremnica).
func (s *OtpremnicaResource) deleteOtpremnica(ctx context.Context, rdokID int64, pre func(tx db.Transaction) error) (string, error) {
	o, found, err := s.repo.KomOtpremnica(ctx, rdokID)
	if err != nil {
		return "", err
	}
	switch {
	case !found:
		return komMsgNalBrisanjeNema, nil
	case strings.EqualFold(strings.TrimSpace(o.Knjige1), "D"):
		return komMsgProknjizen, nil
	case o.Doksts != domain.KomOtpNova:
		return komMsgOtpPotvrdjenaBrisi, nil
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if pre != nil {
		if err := pre(tx); err != nil {
			return "", err
		}
	}
	if err := s.repo.KomZahtevBrisi(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if o.VrdokID > 0 {
		// A nalog deleted or changed meanwhile has nothing to set back.
		if _, err := s.repo.KomOtpNalogVrati(ctx, tx, o.VrdokID); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje otpremnice: %w", err)
	}
	return "", nil
}

// GetStavke returns the stavke of an otpremnica.
func (s *OtpremnicaResource) GetStavke(ctx context.Context, rdokID int64) ([]domain.KomOtpStavkaRow, error) {
	if rdokID <= 0 {
		return nil, nil
	}
	return s.repo.KomOtpStavke(ctx, rdokID)
}

// GetStavkaForma returns a saved stavka for the form of the stavke ("Izmeni" of its row); found is false without it.
func (s *OtpremnicaResource) GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomOtpStavkaForma, bool, error) {
	st, found, err := s.stavka(ctx, rdokID, rproID)
	if err != nil || !found {
		return domain.KomOtpStavkaForma{}, found, err
	}
	return domain.KomOtpStavkaForma{RproID: st.RproID, Rbr: st.Rbr, Sifra: st.Sifra, Naziv: st.Naziv, JM: st.JM, Otk: st.Otk, Serija: st.Serija,
		Rok: komWinDevDatumText(st.Roktr), Trazkolic: st.Trazkolic, Kolicina: st.Kolic}, true, nil
}

// stavka returns the stavka of the id of an otpremnica; found is false without it.
func (s *OtpremnicaResource) stavka(ctx context.Context, rdokID, rproID int64) (domain.KomOtpStavkaRow, bool, error) {
	stavke, err := s.GetStavke(ctx, rdokID)
	if err != nil {
		return domain.KomOtpStavkaRow{}, false, err
	}
	for _, st := range stavke {
		if st.RproID == rproID {
			return st, true, nil
		}
	}
	return domain.KomOtpStavkaRow{}, false, nil
}

// ValidateStavka validates a changed stavka (the legacy exit of a row of TABLE_PR_QRY_EPRO): its otpremnica is new and the
// stavka exists; the isporučena količina of a storno vrsta is not positive and not below the tražena one (both negative),
// of the other vrste not negative (the legacy does not check it) and not above the tražena one.
func (s *OtpremnicaResource) ValidateStavka(ctx context.Context, params domain.KomOtpStavkaParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	o, found, err := s.repo.KomOtpremnica(ctx, params.RdokID)
	switch {
	case err != nil:
		add("kolicina", common.ErrMsgReadData)
		return errs
	case !found:
		add("kolicina", komMsgNePostoji)
		return errs
	case o.Doksts != domain.KomOtpNova:
		add("kolicina", komMsgOtpPotvrdjena)
		return errs
	}
	st, found, err := s.stavka(ctx, params.RdokID, params.RproID)
	if err != nil {
		add("kolicina", common.ErrMsgReadData)
		return errs
	}
	if !found {
		add("kolicina", komMsgOtpStavkaNePostoji)
		return errs
	}
	vrste, err := s.repo.KomOtpVrste(ctx)
	if err != nil {
		add("kolicina", common.ErrMsgReadData)
		return errs
	}
	vrsta, _ := komOtpVrsta(vrste, o.Vrd)
	kolicina := common.StringToFloat64(params.Kolicina)
	if vrsta.Storno() {
		if kolicina > 0 {
			add("kolicina", komMsgOtpStornoNegativna)
		} else if kolicina != 0 && kolicina < st.Trazkolic {
			add("kolicina", komMsgOtpStornoVeca)
		}
		return errs
	}
	if kolicina < 0 {
		add("kolicina", komMsgOtpIsporucenaNegativ)
	} else if kolicina > st.Trazkolic {
		add("kolicina", komMsgOtpIsporucenaVeca)
	}
	return errs
}

// SaveStavka saves a validated stavka in one transaction: its isporučena količina and, with "Dozvoli izmenu OTK, Seriju,
// Rok" in a magacin with the lots, its lot. A changed lot moves the reservation of its tražena količina from the old lot to
// the new one (drsta.rez; the legacy takes the tražena količina off the old lot but adds the isporučena one to the new
// one).
func (s *OtpremnicaResource) SaveStavka(ctx context.Context, params domain.KomOtpStavkaParams) error {
	o, found, err := s.repo.KomOtpremnica(ctx, params.RdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("otpremnica %d ne postoji", params.RdokID)
	}
	st, found, err := s.stavka(ctx, params.RdokID, params.RproID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("stavka %d otpremnice %d ne postoji", params.RproID, params.RdokID)
	}
	otk, serija, roktr := st.Otk, st.Serija, st.Roktr
	if o.Tipzal == 2 && common.StringToBool(params.DozvoliLot) {
		otk, serija, roktr = strings.TrimSpace(params.Otk), strings.TrimSpace(params.Serija), komWinDevDatum(params.Rok)
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomOtpStavkaIzmeni(ctx, tx, o.RdokID, st.RproID, common.StringToFloat64(params.Kolicina), otk, serija, roktr); err != nil {
		return err
	}
	if otk != st.Otk || serija != st.Serija || roktr != st.Roktr {
		stari := komOtpRezervacija(o, st, -math.Abs(st.Trazkolic))
		novi := stari
		novi.Otk, novi.Serija, novi.Roktr, novi.Kolicina = otk, serija, roktr, math.Abs(st.Trazkolic)
		if err := komRezervisiLot(ctx, s.repo, tx, stari); err != nil {
			return err
		}
		if err := komRezervisiLot(ctx, s.repo, tx, novi); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka otpremnice: %w", err)
	}
	return nil
}

// komOtpRezervacija is the reservation of the količina of a stavka of an otpremnica in its lot.
func komOtpRezervacija(o domain.KomOtpremnicaRow, st domain.KomOtpStavkaRow, kolicina float64) domain.KomNalogRezervacija {
	return domain.KomNalogRezervacija{MagaciniID: o.MagaciniID, Mag: o.Mag, RsifID: st.RsifID, Sifra: st.Sifra, Konto: st.Konto, Otk: st.Otk,
		Serija: st.Serija, Roktr: st.Roktr, Cena: st.Cena, Kolicina: kolicina}
}

// KomOtpProvera is the result of the checks of the potvrda of an otpremnica: the message why it cannot be confirmed (a
// message of the translations) with its lines (the stavke, the zalihe: data, not translated), and the stavke without an
// OTK the question of the potvrda asks about. Potvrdjena tells the potvrda was done (Poruka "Dokument je potvrđen!!!").
type KomOtpProvera struct {
	Poruka     string
	Detalji    []string
	BezOtk     []string
	Potvrdjena bool
}

// komOtpLot is the key of a lot of a stavka (the legacy aaSTAVKE of BTN_POTVRDADOK).
type komOtpLot struct {
	konto, otk, serija string
	sifra              int
	roktr              int64
}

// ProveriPotvrdu checks an otpremnica before the question of its potvrda (the legacy BTN_POTVRDADOK): the poslovna godina
// is not closed, it has stavke, it is new, the nastanak PDV obaveze is chosen; in a magacin with the lots every JM exists
// (a stavka of a JM with a length and no width without an OTK is asked about: pitanja); every prodajna cena is above 0;
// the isporučena količina is not 0 in all; the zaliha of every lot covers the isporučena količina of the stavke of the
// lot (the legacy compares the tražena one), not for the usluge (tip U); a lot of a storno vrsta had an ulaz; a broj
// fakture typed with "Dozvoli izmenu" is not one of a robni dokument. The message why it cannot be confirmed is returned
// with its lines, else the stavke without an OTK.
func (s *OtpremnicaResource) ProveriPotvrdu(ctx context.Context, params domain.KomOtpParams) (KomOtpProvera, error) {
	var r KomOtpProvera
	poruka := func(p string, detalji ...string) (KomOtpProvera, error) {
		return KomOtpProvera{Poruka: p, Detalji: detalji}, nil
	}
	o, found, err := s.repo.KomOtpremnica(ctx, params.RdokID)
	if err != nil {
		return r, err
	}
	if !found {
		return poruka(komMsgBezStavkiPotvrda)
	}
	if zatvorena, err := s.repo.KomOtpGodinaZatvorena(ctx); err != nil {
		return r, err
	} else if zatvorena {
		return poruka(komMsgOtpGodinaZatvorena)
	}
	stavke, err := s.repo.KomOtpStavke(ctx, o.RdokID)
	if err != nil {
		return r, err
	}
	if len(stavke) == 0 {
		return poruka(komMsgBezStavkiPotvrda)
	}
	if o.Doksts != domain.KomOtpNova {
		return poruka(komMsgVecPotvrdjen)
	}
	if n := strings.TrimSpace(params.Nastanak); n == "" || n == robnosvc.RobNastanakNeIzabran || n == "-1" {
		return poruka(komMsgOtpNastanak)
	}
	vrste, err := s.repo.KomOtpVrste(ctx)
	if err != nil {
		return r, err
	}
	vrsta, _ := komOtpVrsta(vrste, o.Vrd)
	zameni := func(poruka string, st domain.KomOtpStavkaRow) string {
		return strings.NewReplacer("{{sifra}}", strconv.Itoa(st.Sifra), "{{naziv}}", st.Naziv, "{{jm}}", st.JM, "{{konto}}", st.Konto).Replace(poruka)
	}
	var ukupno float64
	ceneOk := true
	lotovi := map[komOtpLot]float64{}
	var redLotova []komOtpLot
	prva := map[komOtpLot]domain.KomOtpStavkaRow{}
	for _, st := range stavke {
		if o.Tipzal == 2 {
			if !st.JmPostoji {
				return poruka(komMsgOtpJm, zameni(komMsgOtpStavka, st))
			}
			if st.ImaDuzinu && !st.ImaSirinu && strings.TrimSpace(st.Otk) == "" && st.Kolic > 0 {
				r.BezOtk = append(r.BezOtk, zameni(komMsgOtpStavka, st))
			}
		}
		if st.Fcena <= 0 {
			ceneOk = false
		}
		ukupno += st.Kolic
		if st.Kolic == 0 || st.Tip == "U" {
			continue
		}
		k := komOtpLot{konto: st.Konto, otk: st.Otk, serija: st.Serija, sifra: st.Sifra, roktr: st.Roktr}
		if _, ok := lotovi[k]; !ok {
			redLotova = append(redLotova, k)
			prva[k] = st
		}
		lotovi[k] += st.Kolic
	}
	if ukupno == 0 {
		return poruka(komMsgOtpNulaKolicina)
	}
	if !ceneOk {
		return poruka(komMsgOtpCenaNula)
	}
	var nedostaje []string
	for _, k := range redLotova {
		st := prva[k]
		lot, found, err := s.repo.DrstaLotStanje(ctx, k.konto, k.sifra, o.Mag, k.otk, k.serija, k.roktr)
		if err != nil {
			return r, err
		}
		switch {
		case !found:
			if vrsta.Storno() {
				return poruka(komMsgOtpStorno)
			}
			nedostaje = append(nedostaje, zameni(komMsgOtpZalihaNePostoji, st))
		case vrsta.Storno():
			if lot.Ulaz == 0 {
				return poruka(komMsgOtpStorno)
			}
		case math.Round((lot.Ulaz-lot.Izlaz)*1000)/1000 < lotovi[k]:
			nedostaje = append(nedostaje, strings.NewReplacer("{{kolicina}}", common.FormatNumberWithSystemLocale(lotovi[k], 3),
				"{{zaliha}}", common.FormatNumberWithSystemLocale(lot.Ulaz-lot.Izlaz, 3)).Replace(zameni(komMsgOtpZalihaArtikla, st)))
		}
	}
	if len(nedostaje) > 0 {
		return poruka(komMsgOtpZalihe, nedostaje...)
	}
	if !komOtpPrivremen(o.Dokum) {
		if postoji, err := s.repo.KomOtpRdokPostoji(ctx, nil, o.Mag, o.Vrd, o.Dokum); err != nil {
			return r, err
		} else if postoji {
			return poruka(komMsgOtpFakturaPostoji, strconv.Itoa(o.Dokum))
		}
	}
	return r, nil
}

// PotvrdiOtpremnicu confirms an otpremnica after the checks of ProveriPotvrdu (the legacy "Da" of BTN_POTVRDADOK with
// SnimiRDOK) in one transaction:
//   - the brojevi: the typed ones (below 990000) or the next broj otpremnice and the next broj of the robna faktura;
//   - the neto težina when it is 0 (Σ isporučena količina × težina; the legacy IzracunajTezinu);
//   - the robni nalog of the vrsta naloga za knjiženje, the datum and the magacin (a new "Aut. kreiran nalog" without one)
//     counts one more document; the robna faktura of the otpremnica with the opis ("<opis dok.> Dok.br:<broj>") and the
//     nastanak PDV obaveze of the screen;
//   - every stavka with an isporučena količina (negative for a storno vrsta): the stavka of the faktura (see komOtpRpro),
//     the cena of its artikal when it has none, the stanje of the artikal in its konto and lot (the robno AzurirajRSTA and
//     AzurirajDRSTA); every stavka releases the reservation of its tražena količina (the legacy only the delivered ones);
//   - the iznos on the robni nalog (AzurirajRNAL), the broj stavki of the faktura, the avansi of the nalog za isporuku,
//     the otpremnica potvrđena (doksts 40) with its brojevi and opis "Otpremnica br.:<broj>", the isporučene količine on
//     the stavke of the nalog za isporuku.
//
// Then the obračun of the faktura (the robno ObracunFakture: its iznos, ugovoreni rabat, kasa and avansi; the legacy only
// with avansi). It returns the result of the checks, or "Dokument je potvrđen!!!".
func (s *OtpremnicaResource) PotvrdiOtpremnicu(ctx context.Context, params domain.KomOtpParams) (KomOtpProvera, error) {
	if r, err := s.ProveriPotvrdu(ctx, params); err != nil || r.Poruka != "" {
		return r, err
	}
	poruka, err := s.potvrdi(ctx, params)
	if err != nil || poruka != "" {
		return KomOtpProvera{Poruka: poruka}, err
	}
	return KomOtpProvera{Poruka: komMsgPotvrdjen, Potvrdjena: true}, nil
}

// potvrdi is the posting of PotvrdiOtpremnicu; it returns the message why it was not done ("" when done).
func (s *OtpremnicaResource) potvrdi(ctx context.Context, params domain.KomOtpParams) (string, error) {
	o, _, err := s.repo.KomOtpremnica(ctx, params.RdokID)
	if err != nil {
		return "", err
	}
	stavke, err := s.repo.KomOtpStavke(ctx, o.RdokID)
	if err != nil {
		return "", err
	}
	vrste, err := s.repo.KomOtpVrste(ctx)
	if err != nil {
		return "", err
	}
	vrsta, _ := komOtpVrsta(vrste, o.Vrd)
	ino := false
	if session := domain.GetSessionFromStdContext(ctx); session != nil {
		sifvalFirme, _, err := komValutaFirme(ctx, s.repo, session.Firma)
		if err != nil {
			return "", err
		}
		ino = komIno(o.SifVal, sifvalFirme)
	}
	// The nabavne cene of a magacin kept at the nabavna cena (nacvodzal 3): the prosečna, else the computed one.
	nabavne := map[int]float64{}
	if o.Nacvodzal == 3 {
		for _, st := range stavke {
			if _, ok := nabavne[st.Sifra]; ok || st.Kolic == 0 {
				continue
			}
			c, err := s.repo.KomOtpRstaProsnc(ctx, nil, o.MagaciniID, st.Sifra)
			if err != nil {
				return "", err
			}
			if c <= 0 {
				if c, err = s.robno.IzracunajNCENA(ctx, o.Dadok.Time, o.Dadok.Time, o.MagaciniID, st.Sifra); err != nil {
					return "", err
				}
			}
			nabavne[st.Sifra] = komRound2(c)
		}
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	brotp, dokum, err := s.repo.KomOtpNoviBrojevi(ctx, tx, o.Vrd, o.Mag)
	if err != nil {
		return "", err
	}
	if !komOtpPrivremen(o.Brotp) {
		brotp = o.Brotp
	}
	if !komOtpPrivremen(o.Dokum) {
		dokum = o.Dokum
	}
	if o.Netwght == 0 {
		var tezina float64
		for _, st := range stavke {
			tezina += st.Kolic * st.Tezjm
		}
		if err := s.repo.KomOtpNetoTezina(ctx, tx, o.RdokID, tezina); err != nil {
			return "", err
		}
	}
	rnalID, found, err := s.repo.KomOtpRnal(ctx, tx, o.Tipdok, o.Dadok, o.MagaciniID)
	if err != nil {
		return "", err
	}
	if !found {
		if rnalID, err = s.repo.KomOtpRnalDodaj(ctx, tx, domain.KomOtpRnalUpis{Tipdok: o.Tipdok, Danal: o.Dadok.Time, MagaciniID: o.MagaciniID,
			Mag: o.Mag}); err != nil {
			return "", err
		}
	}
	if err := s.repo.KomOtpRnalDokument(ctx, tx, rnalID); err != nil {
		return "", err
	}
	opis := fmt.Sprintf("Dok.br:%d", dokum)
	if od := strings.TrimSpace(params.OpisDok); od != "" {
		opis = od + " " + opis
	}
	rdokID, err := s.repo.KomOtpRdokDodaj(ctx, tx, o.RdokID, rnalID, domain.KomOtpRdokUpis{Dokum: dokum, Dokiz: brotp, Opis: opis,
		DescriptionCode: common.StringToInt(params.Nastanak)})
	if err != nil {
		return "", err
	}
	var ukupno float64
	brst := 0
	for _, st := range stavke {
		if st.Kolic != 0 {
			brst++
			r := komOtpRpro(o, st, vrsta.Storno(), nabavne[st.Sifra], ino)
			r.RnalID, r.RdokID = rnalID, rdokID
			if _, err := s.repo.KomOtpRproDodaj(ctx, tx, r); err != nil {
				return "", err
			}
			fcena := st.Fcena
			if ino && o.VaziZa > 0 {
				fcena = komRound2(st.Fcena * o.Kurs / float64(o.VaziZa))
			}
			if err := s.repo.KomOtpRceneDodaj(ctx, tx, domain.KomOtpRceneUpis{RsifID: st.RsifID, Sifra: st.Sifra, Ncena: r.Ncena, Rab: st.Rab,
				Cena: st.Cena, Fcena: fcena}); err != nil {
				return "", err
			}
			p := domain.RobPromenaStanja{MagaciniID: o.MagaciniID, Konto: st.Konto, Sifra: st.Sifra, Kolicina: r.Kolic, Iznos: r.Iznos, Cena: r.Cena,
				Danal: o.Dadok.Time, DokvrstaID: vrsta.DokvrstaID, Tipdok: o.Tipdok, Brst: 1}
			if err := s.robno.AzurirajRSTA(ctx, tx, p); err != nil {
				return "", err
			}
			if o.Tipzal == 1 || o.Tipzal == 2 {
				p.Danal, p.Tipdok, p.Otk, p.Serija, p.Roktr = o.Datotp.Time, "", st.Otk, st.Serija, st.Roktr
				if err := s.robno.AzurirajDRSTA(ctx, tx, p); err != nil {
					return "", err
				}
			}
			ukupno += r.Iznos
		}
		if err := komRezervisi(ctx, s.repo, tx, komOtpRezervacija(o, st, -math.Abs(st.Trazkolic))); err != nil {
			return "", err
		}
	}
	if err := s.robno.AzurirajRNAL(ctx, tx, rnalID, vrsta.Kodknj == "D", vrsta.Kodknj == "P", komRound2(ukupno)); err != nil {
		return "", err
	}
	if err := s.repo.KomOtpRdokBrst(ctx, tx, rdokID, brst); err != nil {
		return "", err
	}
	if _, err := s.repo.KomOtpAvansi(ctx, tx, o.VrdokID, rdokID); err != nil {
		return "", err
	}
	ok, err := s.repo.KomOtpPotvrdi(ctx, tx, o.RdokID, dokum, brotp, fmt.Sprintf("Otpremnica br.:%d", brotp))
	if err != nil {
		return "", err
	}
	if !ok {
		return komMsgVecPotvrdjen, nil
	}
	if err := s.repo.KomOtpNalogIsporuceno(ctx, tx, o.RdokID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit potvrda otpremnice: %w", err)
	}
	return "", s.robno.ObracunFakture(ctx, rdokID)
}

// komOtpRpro is the stavka of the robna faktura of a stavka of an otpremnica (the legacy SnimiRDOK "Novi slog u RPRO") by
// the way the magacin keeps its zalihe (nacvodzal): 3 at the nabavna cena (the cena and the nabavna cena are the nabavna
// one, the iznos the količina by it), 1 at the veleprodajna cena (the cena is the prodajna one, the iznos the količina by
// it), else at the cena of the stavka with the iznos 0 (the legacy). The količine and the iznos are negative for a storno
// vrsta. The poreska kategorija: for the oznaka "S" S20 or S10 by the tarifa (1, 2), else the oznaka.
func komOtpRpro(o domain.KomOtpremnicaRow, st domain.KomOtpStavkaRow, storno bool, nabavna float64, ino bool) domain.KomOtpRproUpis {
	znak := 1.0
	if storno {
		znak = -1
	}
	kolic := znak * math.Abs(st.Kolic)
	r := domain.KomOtpRproUpis{Sifra: st.Sifra, RsifID: st.RsifID, Konto: st.Konto, JM: st.JM, Kolic: kolic, Kolic1: znak * math.Abs(st.Kolic1),
		Pcena: st.Pcena, Fcena: st.Fcena, CenaVal: st.CenaVal, Mcenap: st.Mcenap, Rab: st.Rab, Po: st.Po, Brst: st.Brst, Rbr: st.Rbr, Otk: st.Otk,
		Serija: st.Serija, Roktr: st.Roktr, MagaciniID: o.MagaciniID, Mag: o.Mag}
	switch o.Nacvodzal {
	case 3:
		r.Cena, r.Ncena = nabavna, nabavna
		r.Iznos = znak * math.Abs(komRound2(st.Kolic*nabavna))
	case 1:
		r.Cena = st.Fcena
		r.Iznos = znak * math.Abs(komRound2(st.Kolic*st.Fcena))
	default:
		r.Cena = st.Cena
	}
	if strings.TrimSpace(o.Category) == "S" {
		switch st.Po {
		case 1:
			r.Taxcat, r.Taxexreco = "S20", "-"
		case 2:
			r.Taxcat, r.Taxexreco = "S10", "-"
		}
	} else {
		r.Taxcat = strings.TrimSpace(o.Category)
	}
	return r
}

// GetPrikaz returns the otpremnice of the list of the main screen: the nepotvrđene or the potvrđene (any status other than
// 10); nothing without a magacin (unless all of them).
func (s *OtpremnicaResource) GetPrikaz(ctx context.Context, params domain.KomOtpPrikazParams) ([]domain.KomOtpPrikazRow, error) {
	if params.Doksts != domain.KomOtpNova {
		params.Doksts = domain.KomOtpPotvrdjena
	}
	if !params.SviMagacini && params.MagaciniID <= 0 {
		return nil, nil
	}
	return s.repo.KomOtpPrikaz(ctx, params)
}

// GetStampa returns the print of an otpremnica (the legacy PR_RPT_OTPREMNICA): the firm with the address of the magacin
// and the tekući račun of the otpremnica (all the računi of the firm without one), the kupac, the header, the stavke with an
// isporučena količina (the naziv with the serija and the rok trajanja in a magacin with the lots, with the komercijalni
// opis, the proizvođač and his šifra when asked), the prodajna cena and the iznos (in a foreign valuta at the cena u
// valuti; the legacy divides the prodajna cena by the kurs and "važi za"), the totals, the weights and the napomene; "Storno
// otpremnica" for a storno vrsta (the legacy checks the knjiženje of the vrsta); a document without stavke is not printed
// (the message why is returned).
func (s *OtpremnicaResource) GetStampa(ctx context.Context, rdokID int64, komercijalniOpis bool) (domain.KomOtpStampa, string, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomOtpStampa{}, "", fmt.Errorf("no user session found")
	}
	o, found, err := s.repo.KomOtpremnica(ctx, rdokID)
	if err != nil {
		return domain.KomOtpStampa{}, "", err
	}
	if !found {
		return domain.KomOtpStampa{}, komMsgBezStavkiStampa, nil
	}
	stavke, err := s.repo.KomOtpStavke(ctx, rdokID)
	if err != nil {
		return domain.KomOtpStampa{}, "", err
	}
	if len(stavke) == 0 {
		return domain.KomOtpStampa{}, komMsgBezStavkiStampa, nil
	}
	vrste, err := s.repo.KomOtpVrste(ctx)
	if err != nil {
		return domain.KomOtpStampa{}, "", err
	}
	vrsta, _ := komOtpVrsta(vrste, o.Vrd)
	sifvalFirme, _, err := komValutaFirme(ctx, s.repo, session.Firma)
	if err != nil {
		return domain.KomOtpStampa{}, "", err
	}
	st := domain.KomOtpStampa{Storno: vrsta.Storno(), Ino: komIno(o.SifVal, sifvalFirme), OznVal: o.OznVal, SifVal: o.SifVal, Kurs: o.Kurs,
		VaziZa: o.VaziZa, Brotp: o.Brotp, Datotp: o.Datotp, NalogDokum: o.NalogDokum, Pla: o.Pla, Rok: o.Rok, Komercijalista: o.KomIme,
		Vozac: o.Vozac, Brvozila: o.Brvozila, Magacin: o.MagacinText, OJ: o.OrgjedText, Bttowght: o.Bttowght, Netwght: o.Netwght,
		Totboxs: o.Totboxs, Brpaleta: o.Brpaleta, Napomena: o.Foot, Pornapomena: strings.TrimSpace(o.Pornapomena),
		Nepotvrdjen: o.Doksts == domain.KomOtpNova}
	if st.Firma, err = s.repo.KomFirma(ctx, session.Firma); err != nil {
		return domain.KomOtpStampa{}, "", err
	}
	st.MestoIzdavanja = strings.TrimSpace(o.MagacinMesto)
	if st.MestoIzdavanja == "" {
		st.MestoIzdavanja = strings.TrimSpace(st.Firma.Mesto)
	}
	if adresa := strings.TrimSpace(o.MagacinAdresa); adresa != "" {
		st.Firma.Adresa = strings.TrimSpace(st.Firma.Adresa) + " , " + adresa
	}
	st.TekuciRacun = strings.TrimSpace(o.Posuslnab)
	if st.TekuciRacun == "" || st.TekuciRacun == "-" || st.TekuciRacun == "-1" {
		racuni, err := s.repo.BankeZiroRacuni(ctx)
		if err != nil {
			return domain.KomOtpStampa{}, "", err
		}
		st.TekuciRacun = strings.Join(racuni, "; ")
	}
	if idFkpl, err := s.repo.KomPorIDFkpl(ctx, o.Fkto, o.Fana); err != nil {
		return domain.KomOtpStampa{}, "", err
	} else if idFkpl > 0 {
		if st.Kupac, _, err = s.repo.KomPonudaStampaDobavljac(ctx, idFkpl); err != nil {
			return domain.KomOtpStampa{}, "", err
		}
	}
	rbr := 0
	for _, stavka := range stavke {
		if stavka.Kolic == 0 {
			continue
		}
		rbr++
		naziv := stavka.Naziv
		if o.Tipzal == 2 {
			naziv += "\nSerija: " + stavka.Serija + " Rok trajanja: " + komWinDevDatumText(stavka.Roktr)
		}
		if komercijalniOpis {
			for _, d := range []string{stavka.Komercopis, stavka.Pro, stavka.Proizsifra} {
				if d = strings.TrimSpace(d); d != "" {
					naziv += " " + d
				}
			}
		}
		cena := stavka.Fcena
		if st.Ino {
			cena = stavka.CenaVal
		}
		iznos := komRound2(stavka.Kolic * cena)
		st.Stavke = append(st.Stavke, domain.KomOtpStampaStavka{Rbr: rbr, Sifra: stavka.Sifra, Naziv: naziv, JM: stavka.JM, Otk: stavka.Otk,
			Serija: stavka.Serija, Rok: komWinDevDatumText(stavka.Roktr), Kolicina: stavka.Kolic, Decimale: stavka.Decimale, Cena: cena,
			ProcRab: stavka.Rab, Rabat: komRound2(iznos * stavka.Rab / 100), Iznos: iznos})
		st.Svega += iznos
		st.UkupnoRSD += stavka.Kolic * stavka.Fcena
	}
	st.Svega, st.UkupnoRSD = komRound2(st.Svega), komRound2(st.UkupnoRSD)
	return st, "", nil
}

// StavkeTableFields are the columns of the grid of the stavke of the entry dialog ("KOLIČINE PO NALOGU ZA OTPREMU", the
// legacy TABLE_PR_QRY_EPRO): OTK, serija and rok trajanja only in a magacin with the lots.
func (s *OtpremnicaResource) StavkeTableFields(lot bool) []domain.Fields {
	fields := []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "5", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla/proizvoda", Width: "8", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla/proizvoda", Width: "28"},
		{Name: "komercopis", Label: "Komercijalni opis", Width: "16"},
		{Name: "jm", Label: "JM", Width: "5"},
	}
	if lot {
		fields = append(fields, domain.Fields{Name: "otk", Label: "OTK", Width: "7"}, domain.Fields{Name: "serija", Label: "Serija", Width: "8"},
			domain.Fields{Name: "rok", Label: "Rok trajanja", Width: "8"})
	}
	return append(fields,
		domain.Fields{Name: "trazkolic", Label: "Tražena količina", Width: "9", TextAlign: "right"},
		domain.Fields{Name: "kolic", Label: "Isporučena količina", Width: "9", TextAlign: "right"},
	)
}

// PrikazTableFields are the columns of the list of the main screen (the legacy TABLE_QRY_EDOK).
func (s *OtpremnicaResource) PrikazTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "magacin", Label: "Magacin", Width: "10"},
		{Name: "vrsta", Label: "Vrsta dokum.", Width: "10"},
		{Name: "dokum", Label: "Broj fakture", Width: "5", TextAlign: "right"},
		{Name: "dadok", Label: "Datum fakture", Width: "6"},
		{Name: "brotp", Label: "Broj otpr.", Width: "5", TextAlign: "right"},
		{Name: "nalog", Label: "Broj naloga za isporuku", Width: "5", TextAlign: "right"},
		{Name: "fkto", Label: "Konto partnera", Width: "5"},
		{Name: "fana", Label: "Šifra", Width: "5"},
		{Name: "kupac", Label: "Naziv partnera", Width: "16"},
		{Name: "doksts", Label: "Status dokum.", Width: "4", TextAlign: "right"},
		{Name: "rok", Label: "Rok plaćanja", Width: "4", TextAlign: "right"},
		{Name: "datiz", Label: "Datum izvornog dokum.", Width: "6"},
		{Name: "opis", Label: "Opis", Width: "10"},
		{Name: "mestoisporuke", Label: "Mesto isporuke", Width: "8"},
		{Name: "kamt", Label: "Uslovi kamate", Width: "6"},
	}
}

// PrikazStavkeTableFields are the columns of the stavke of the selected otpremnica of the main screen ("DETALJI
// DOKUMENTA", the legacy TABLE_EPRO2).
func (s *OtpremnicaResource) PrikazStavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "5", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla/sirovine", Width: "8", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla/sirovine", Width: "28"},
		{Name: "jm", Label: "JM", Width: "4"},
		{Name: "otk", Label: "OTK", Width: "6"},
		{Name: "serija", Label: "Serija", Width: "7"},
		{Name: "rok", Label: "Rok", Width: "7"},
		{Name: "trazkolic", Label: "Zahtevana količina", Width: "8", TextAlign: "right"},
		{Name: "kolic", Label: "Isporučena količina", Width: "8", TextAlign: "right"},
		{Name: "fcena", Label: "Jedinična cena", Width: "8", TextAlign: "right"},
		{Name: "cenaval", Label: "Cena u valuti", Width: "8", TextAlign: "right"},
		{Name: "rab", Label: "% Rab.", Width: "5", TextAlign: "right"},
	}
}

// OtpStavkeTableRows are the rows of the stavke of the entry dialog with "Izmeni" of every row (izmena): OTK, serija and
// rok trajanja only with the lots.
func OtpStavkeTableRows(stavke []domain.KomOtpStavkaRow, lot, izmena bool) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		fields := []string{strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.Komercopis, st.JM}
		if lot {
			fields = append(fields, st.Otk, st.Serija, komWinDevDatumText(st.Roktr))
		}
		fields = append(fields, komProKolicinaText(st.Trazkolic, st.Decimale), komProKolicinaText(st.Kolic, st.Decimale))
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), HasUpdate: izmena, Fields: fields})
	}
	return rows
}

// OtpPrikazTableRows are the rows of the list of the main screen: "Izmeni" and "Briši" of every new otpremnica (izmena).
func OtpPrikazTableRows(otpremnice []domain.KomOtpPrikazRow, izmena bool) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(otpremnice))
	broj := func(n int) string {
		if n <= 0 {
			return ""
		}
		return strconv.Itoa(n)
	}
	for _, o := range otpremnice {
		menja := izmena && o.Doksts == domain.KomOtpNova
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(o.RdokID, 10), HasUpdate: menja, HasDelete: menja, Fields: []string{
			o.Magacin, o.VrstaOpis, broj(o.Dokum), komDatumText(o.Dadok), broj(o.Brotp), broj(o.NalogDokum), o.Fkto, o.Fana, o.Kupac, o.Doksts,
			strconv.Itoa(o.Rok), komDatumText(o.Datiz), o.Opis, o.MestoIsporuke, o.Kamt,
		}})
	}
	return rows
}

// OtpPrikazStavkeTableRows are the rows of the stavke of the selected otpremnica of the main screen.
func OtpPrikazStavkeTableRows(stavke []domain.KomOtpStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), Fields: []string{
			strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM, st.Otk, st.Serija, komWinDevDatumText(st.Roktr),
			komProKolicinaText(st.Trazkolic, st.Decimale), komProKolicinaText(st.Kolic, st.Decimale), common.FormatNumberWithSystemLocale(st.Fcena, 2),
			common.FormatNumberWithSystemLocale(st.CenaVal, 2), common.FormatNumberWithSystemLocale(st.Rab, 2),
		}})
	}
	return rows
}
