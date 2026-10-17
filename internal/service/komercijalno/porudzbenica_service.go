package komercijalno

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	commonsvc "helia/internal/service/common"
)

// The messages of the porudžbenica dobavljaču (the legacy pr_PorudzbenicaDOBAVLJ and pr_PorudzbenicaDOBAVLJ_DETALJI);
// the ones it shares with the zahtevi are the komMsg constants of zahtevnabavka_service.go and zahtevponuda_service.go.
const (
	komMsgPorKonto           = "Morate uneti konto dobavljača..."
	komMsgPorSifra           = "Morate uneti šifru dobavljača..."
	komMsgPorDobavljac       = "Dobavljač sa takvom šifrom ne postoji!!!..."
	komMsgPorValuta          = "Valuta sa takvom šifrom ne postoji!!!..."
	komMsgPorKurs            = "Morate uneti kurs valute!!!..."
	komMsgPorVeza            = "Zahtev za ponudu ne postoji ili nije potvrđen."
	komMsgPorNijeIzmena      = "Potvrđena porudžbenica se ne može menjati."
	komMsgPorFiskalIzmena    = "Ovaj dokument je odštampan na fiskalni printer, izmene nisu dozvoljene!!!"
	komMsgPorKnjizenIzmena   = "Ovaj dokument je proknjižen, izmene nisu dozvoljene!!!"
	komMsgPorPotvrdjenBris   = "Dokument je potvrđen, brisanje nije dozvoljeno!!!"
	komMsgPorCenaNula        = "Postoje stavke čija je cena = 0. Potvrda dokumenta nije dozvoljena!!!"
	komMsgPorCena            = "Cena mora biti veća od 0!!!!"
	komMsgPorNemaZaDodavanje = "Ne postoji nijedna stavka za dodavanje!!!"
	komMsgPorNemaZaBrisanje  = "Ne postoji nijedna stavka za brisanje!!!"
)

// komCenaUValuti is the label of the cena of a stavka of a porudžbenica in a foreign valuta: "Jedinična cena {{valuta}}"
// (the legacy caption "Jedinična cena EUR").
const komCenaUValuti = "Jedinična cena u valuti"

// The magacini of a porudžbenica: the ones of the user of the tipovi P, S, D, L and V (the legacy ROB_QRY_MAGUSER with
// "P;S;D;L;V").
var komPorTipoviMagacina = []string{"P", "S", "D", "L", "V"}

// PorudzbenicaService is the "Porudžbenice dobavljačima" screen of the komercijalno (the legacy pr_PorudzbenicaDOBAVLJ).
type PorudzbenicaService interface {
	GetPorudzbenica(ctx context.Context, rdokID int64, vrd, magaciniID int, vezaID int64) (domain.KomPorDokument, error)
	GetVrste(ctx context.Context) ([]domain.ComboItem, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	ValidatePorudzbenica(ctx context.Context, params domain.KomPorParams) []domain.FieldError
	SavePorudzbenica(ctx context.Context, params domain.KomPorParams) (int64, error)
	DeletePorudzbenica(ctx context.Context, rdokID int64) (string, error)
	ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error)
	PotvrdiPorudzbenicu(ctx context.Context, rdokID int64) (string, error)
	GetDobavljac(ctx context.Context, konto, sifra string) (domain.KomDobavljacPodaci, bool, error)
	GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error)
	GetStavke(ctx context.Context, rdokID int64) ([]domain.KomPorStavkaRow, bool, error)
	GetArtikal(ctx context.Context, sifra string) (domain.KomArtikal, bool, error)
	ValidateStavka(ctx context.Context, params domain.KomPorStavkaParams) []domain.FieldError
	SaveStavka(ctx context.Context, params domain.KomPorStavkaParams) error
	DeleteStavka(ctx context.Context, rdokID, rproID int64) error
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomPorStavkaForma, bool, error)
	Preuzmi(ctx context.Context, rdokID, rproID int64) (string, error)
	VratiPoslednju(ctx context.Context, rdokID int64) (string, error)
	VratiSve(ctx context.Context, rdokID int64) (string, error)
	GetPrikaz(ctx context.Context, params domain.KomPonudaPrikazParams) ([]domain.KomPorPrikazRow, error)
	GetStampa(ctx context.Context, rdokID int64, naziv, komercopis bool) (domain.KomPorStampa, string, error)
	StavkeTableFields(ino bool, oznval string) []domain.Fields
	IzvoriTableFields() []domain.Fields
	PrikazTableFields() []domain.Fields
	PrikazStavkeTableFields() []domain.Fields
}

// porRepository is what the service needs of repository.KomercijalnoRepository.
type porRepository interface {
	komOrgjedRepository
	komValutaRepository
	BeginTx() (db.Transaction, error)
	KomPorVrste(ctx context.Context) ([]domain.ComboItem, error)
	KomPorNoviBroj(ctx context.Context, tx db.Transaction, vrd, mag int) (int, error)
	KomPorudzbenica(ctx context.Context, rdokID int64) (domain.KomPorudzbenicaRow, bool, error)
	KomPorDodaj(ctx context.Context, tx db.Transaction, p domain.KomPorudzbenicaUpis) (int64, error)
	KomPorIzmeni(ctx context.Context, tx db.Transaction, p domain.KomPorudzbenicaUpis) error
	KomPorPotvrdi(ctx context.Context, rdokID int64) (bool, error)
	KomPorBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomPorVeza(ctx context.Context, rdokID int64) (domain.KomPorVeza, bool, error)
	KomPorDobavljac(ctx context.Context, konto, sifra string) (domain.KomDobavljacPodaci, bool, error)
	KomPorIDFkpl(ctx context.Context, konto, sifra string) (int64, error)
	KomPorStavke(ctx context.Context, rdokID int64) ([]domain.KomPorStavkaRow, error)
	KomPorStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomPorStavkaUpis) error
	KomPorStavkaIzmeni(ctx context.Context, tx db.Transaction, rdokID, rproID int64, kolicina, cena, cenaVal, rabat float64) error
	KomPorVratiSve(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomPorIzvori(ctx context.Context, ponudaID int64) ([]domain.KomPorIzvorRow, error)
	KomPorIzvor(ctx context.Context, tx db.Transaction, ponudaID, rproID int64) (domain.KomPorIzvor, bool, error)
	KomPorPrikaz(ctx context.Context, p domain.KomPonudaPrikazParams) ([]domain.KomPorPrikazRow, error)
	KomPonudaIzvorRezervisi(ctx context.Context, tx db.Transaction, rproID int64) error
	KomPonudaStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error
	KomPonudaStampaDobavljac(ctx context.Context, idFkpl int64) (domain.KomPonudaStampaDobavljac, bool, error)
	KomZahtevNoviRbr(ctx context.Context, tx db.Transaction, rdokID int64) (int, error)
	KomArtikal(ctx context.Context, sifra int) (domain.KomArtikal, bool, error)
	KomKomercijalista(ctx context.Context, sifra int) (domain.KomKomercijalista, bool, error)
	KomKomercijalistaKorisnika(ctx context.Context, userID int64) (domain.KomKomercijalista, bool, error)
	KomMagacin(ctx context.Context, magaciniID int) (int, bool, error)
	KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error)
	BankeZiroRacuni(ctx context.Context) ([]string, error)
	RvrFdob(ctx context.Context) (string, error)
}

// PorudzbenicaResource implements PorudzbenicaService.
type PorudzbenicaResource struct {
	repo   porRepository
	common magaciniKorisnika
	now    func() time.Time
}

// NewPorudzbenicaService returns the service of the "Porudžbenice dobavljačima" screen.
func NewPorudzbenicaService(repo porRepository, common magaciniKorisnika) *PorudzbenicaResource {
	return &PorudzbenicaResource{repo: repo, common: common, now: time.Now}
}

// komPorMagacini is the option of the magacini of the user of a porudžbenica (the tipovi P, S, D, L and V).
func komPorMagacini() commonsvc.ComboOption {
	return commonsvc.WithIn("magacini.tipmag", komPorTipoviMagacina...)
}

// komSistemPdv is the text of the sistem PDV of a partner (partneri.tippdv; the legacy EDT_pdv).
func komSistemPdv(tippdv int) string {
	switch tippdv {
	case 1:
		return "U sistemu PDV-a"
	case 2:
		return "Van sistema PDV-a"
	case 3:
		return "Registrovani poljoprivrednik"
	case 4:
		return "Fizičko lice"
	case 5:
		return "Student"
	}
	return ""
}

// komIno tells whether a valuta is a foreign one: given and not the one of the firma (the legacy STC_IZVOZ and the
// prices of the stavke in the valuta).
func komIno(sifval, sifvalFirme int) bool {
	return sifval > 0 && sifval != sifvalFirme
}

// GetVrste returns the vrste dokumenata of the porudžbenice dobavljačima (the filter of the main screen).
func (s *PorudzbenicaResource) GetVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return s.repo.KomPorVrste(ctx)
}

// GetMagacini returns the magacini of the user of the tipovi P, S, D, L and V (the filter of the main screen).
func (s *PorudzbenicaResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.common.GetMagacinKorisnikaComboValues(ctx, komPorMagacini())
}

// GetDobavljac returns the dobavljač of the konto and šifra with the text of its sistem PDV (the legacy exit of EDT_FANA);
// found is false without it.
func (s *PorudzbenicaResource) GetDobavljac(ctx context.Context, konto, sifra string) (domain.KomDobavljacPodaci, bool, error) {
	if strings.TrimSpace(konto) == "" || strings.TrimSpace(sifra) == "" {
		return domain.KomDobavljacPodaci{}, false, nil
	}
	d, found, err := s.repo.KomPorDobavljac(ctx, konto, sifra)
	if err != nil || !found {
		return domain.KomDobavljacPodaci{}, found, err
	}
	d.SistemPdv = komSistemPdv(d.TipPdv)
	return d, true, nil
}

// GetValuta returns the valuta of the šifra with its kurs on the datum (the legacy exit of EDT_SIFVAL).
func (s *PorudzbenicaResource) GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error) {
	return komValutaKurs(ctx, s.repo, sifval, datum)
}

// GetPorudzbenica returns the content of the tab "Porudžbenice dobavljačima" with the porudžbenica of the id (a
// nepotvrđena one, the legacy BTN_TraziDOK1 and the exit of EDT_DOKUM) or, for the id 0, a new one (the legacy
// BTN_NoviDOK): of the vrsta and the magacin (the first ones without them), the next broj of the vrsta in the magacin
// (NoviBrojEDOK), the datum of today, the konto of the dobavljači (rvr.fdob), the komercijalista of the user and the
// valuta of the firma with the kurs 1. A new one of a zahtev za ponudu (vezaID, the legacy BTN_TraziDOK) takes its rok,
// način dopremanja, magacin, org. jedinica, mesto troška and dobavljač. A confirmed, fiscalized or posted porudžbenica
// cannot be changed: a new one is given with the message why.
func (s *PorudzbenicaResource) GetPorudzbenica(ctx context.Context, rdokID int64, vrd, magaciniID int, vezaID int64) (domain.KomPorDokument, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomPorDokument{}, fmt.Errorf("no user session found")
	}
	var dok domain.KomPorDokument
	var err error
	if dok.Vrste, err = s.repo.KomPorVrste(ctx); err != nil {
		return dok, err
	}
	if dok.Magacini, err = s.common.GetMagacinKorisnikaComboValues(ctx, komPorMagacini()); err != nil {
		return dok, err
	}
	sifvalFirme, oznvalFirme, err := komValutaFirme(ctx, s.repo, session.Firma)
	if err != nil {
		return dok, err
	}
	dok.SifValFirme = sifvalFirme
	if rdokID > 0 {
		p, found, err := s.repo.KomPorudzbenica(ctx, rdokID)
		if err != nil {
			return dok, err
		}
		switch {
		case !found:
			dok.Poruka = komMsgNePostoji
		case strings.EqualFold(strings.TrimSpace(p.Knjige4), "S"):
			dok.Poruka, vrd, magaciniID = komMsgPorFiskalIzmena, p.Vrd, p.MagaciniID
		case strings.EqualFold(strings.TrimSpace(p.Knjige1), "D"):
			dok.Poruka, vrd, magaciniID = komMsgPorKnjizenIzmena, p.Vrd, p.MagaciniID
		case p.Doksts != "" && p.Doksts != domain.KomZahtevNepotvrdjen:
			dok.Poruka, vrd, magaciniID = komMsgPorNijeIzmena, p.Vrd, p.MagaciniID
		default:
			return s.sacuvana(ctx, dok, p, sifvalFirme)
		}
	}
	h := &dok.Header
	h.Vrd = komIzListe(dok.Vrste, vrd)
	h.Datum = s.now()
	h.Kurs, h.VaziZa = 1, 1
	h.SifVal, h.OznVal = sifvalFirme, oznvalFirme
	if vezaID > 0 {
		v, found, err := s.repo.KomPorVeza(ctx, vezaID)
		if err != nil {
			return dok, err
		}
		if found {
			magaciniID = v.MagaciniID
			h.VrdokID, h.VezaDokum, h.VezaDatum, h.Rok, h.Otp = v.RdokID, v.Dokum, v.Dadok, v.Rok, v.Otp
			h.Orgjed, h.OrgjedText, h.MestoTroska, h.MestoTroskaText = v.Orgjed, v.OrgjedText, v.MestoTroska, v.MtText
			h.Fkto, h.Fana = v.Fkto, v.Fana
		} else {
			dok.Poruka = komMsgPorVeza
		}
	}
	h.MagaciniID = komIzListe(dok.Magacini, magaciniID)
	if h.Fkto == "" {
		if h.Fkto, err = s.repo.RvrFdob(ctx); err != nil {
			return dok, err
		}
	}
	if h.Fana != "" {
		if h.Dobavljac, _, err = s.GetDobavljac(ctx, h.Fkto, h.Fana); err != nil {
			return dok, err
		}
	}
	if h.Vrd > 0 && h.MagaciniID > 0 {
		mag, found, err := s.repo.KomMagacin(ctx, h.MagaciniID)
		if err != nil {
			return dok, err
		}
		if found {
			if h.Dokum, err = s.repo.KomPorNoviBroj(ctx, nil, h.Vrd, mag); err != nil {
				return dok, err
			}
		}
	}
	if k, found, err := s.repo.KomKomercijalistaKorisnika(ctx, session.UserID); err != nil {
		return dok, err
	} else if found {
		h.KomSifra, h.KomIme = k.Sifra, k.Ime
	}
	dok.NoviRbr = 1
	return dok, nil
}

// sacuvana is the content of the tab with a saved nepotvrđena porudžbenica: its header with the dobavljač, the
// slobodne stavke of its zahtev za ponudu, the next redni broj and whether it has stavke.
func (s *PorudzbenicaResource) sacuvana(ctx context.Context, dok domain.KomPorDokument, p domain.KomPorudzbenicaRow, sifvalFirme int) (domain.KomPorDokument, error) {
	h := domain.KomPorHeaderView{RdokID: p.RdokID, Vrd: p.Vrd, Dokum: p.Dokum, Rok: p.Rok, MagaciniID: p.MagaciniID, Fkto: p.Fkto, Fana: p.Fana,
		KomSifra: p.KomSifra, KomIme: p.KomIme, Orgjed: p.Orgjed, OrgjedText: p.OrgjedText, MestoTroska: p.MestoTroska,
		MestoTroskaText: p.MestoTroskaText, Tekst: p.Foot, SifVal: p.SifVal, OznVal: p.OznVal, Kurs: p.Kurs, VaziZa: p.VaziZa,
		VrdokID: p.VrdokID, VezaDokum: p.VezaDokum, VezaDatum: p.VezaDatum, Otp: p.Otp, Pla: p.Pla, Brugov: p.Brugov, Nacpak: p.Nacpak,
		Posuslnab: p.Posuslnab, Pratdok: p.Pratdok, Ino: komIno(p.SifVal, sifvalFirme)}
	if p.Dadok.Valid {
		h.Datum = p.Dadok.Time
	}
	var err error
	if h.Dobavljac, _, err = s.GetDobavljac(ctx, p.Fkto, p.Fana); err != nil {
		return dok, err
	}
	dok.Header = h
	if p.VrdokID > 0 {
		if dok.Izvori, err = s.repo.KomPorIzvori(ctx, p.VrdokID); err != nil {
			return dok, err
		}
	}
	if dok.NoviRbr, err = s.repo.KomZahtevNoviRbr(ctx, nil, p.RdokID); err != nil {
		return dok, err
	}
	dok.ImaStavke = dok.NoviRbr > 1
	return dok, nil
}

// ValidatePorudzbenica validates the header of a porudžbenica (the legacy ValidacijaEDOK): a saved one is nepotvrđena;
// the vrsta and the magacin of a new one are ones of the screen and its zahtev za ponudu (when given) is a potvrđen
// one; the datum is a valid date, the dobavljač (konto and šifra) exists, a given org. jedinica and mesto troška exist
// (the mesto troška of the org. jedinica; both not required, "-"), the valuta exists, a given komercijalista exists and
// a foreign valuta has a kurs.
func (s *PorudzbenicaResource) ValidatePorudzbenica(ctx context.Context, params domain.KomPorParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if params.RdokID > 0 {
		p, found, err := s.repo.KomPorudzbenica(ctx, params.RdokID)
		switch {
		case err != nil:
			add("brdok", common.ErrMsgReadData)
		case !found:
			add("brdok", komMsgNePostoji)
		case p.Doksts != "" && p.Doksts != domain.KomZahtevNepotvrdjen:
			add("brdok", komMsgPorNijeIzmena)
		}
	} else {
		if vrste, err := s.repo.KomPorVrste(ctx); err != nil {
			add("vrd", common.ErrMsgReadData)
		} else if !komUListi(vrste, common.StringToInt(params.Vrd)) {
			add("vrd", komMsgVrsta)
		}
		if magacini, err := s.common.GetMagacinKorisnikaComboValues(ctx, komPorMagacini()); err != nil {
			add("magaciniid", common.ErrMsgReadData)
		} else if !komUListi(magacini, common.StringToInt(params.MagaciniID)) {
			add("magaciniid", komMsgMagacin)
		}
		if id, given, ok := komID(params.VrdokID); !ok {
			add("vezabroj", komMsgPorVeza)
		} else if given {
			if _, found, err := s.repo.KomPorVeza(ctx, id); err != nil {
				add("vezabroj", common.ErrMsgReadData)
			} else if !found {
				add("vezabroj", komMsgPorVeza)
			}
		}
	}
	if _, ok := komDatum(params.Datum); !ok {
		add("datum", komMsgDatum)
	}
	konto, sifra := strings.TrimSpace(params.Fkto), strings.TrimSpace(params.Fana)
	switch {
	case konto == "":
		add("fkto", komMsgPorKonto)
	case sifra == "":
		add("fana", komMsgPorSifra)
	default:
		if _, found, err := s.repo.KomPorDobavljac(ctx, konto, sifra); err != nil {
			add("fana", common.ErrMsgReadData)
		} else if !found {
			add("fana", komMsgPorDobavljac)
		}
	}
	// The org. jedinica and the mesto troška are not required ("-", like the other documents; the legacy required both):
	// a given one exists and the mesto troška is one of the org. jedinica.
	komProveriOrgjedMt(ctx, s.repo, params.Orgjed, params.MestoTroska, add)
	sifval := common.StringToInt(strings.TrimSpace(params.SifVal))
	if _, _, found, err := s.repo.Valuta(ctx, sifval); err != nil {
		add("valutaid", common.ErrMsgReadData)
	} else if !found {
		add("valutaid", komMsgPorValuta)
	} else if session := domain.GetSessionFromStdContext(ctx); session != nil {
		if sifvalFirme, _, err := komValutaFirme(ctx, s.repo, session.Firma); err == nil && sifval != sifvalFirme && komBroj(params.Kurs, 0) <= 0 {
			add("kurs", komMsgPorKurs)
		}
	}
	if sk := strings.TrimSpace(params.KomSifra); sk != "" && sk != "0" && sk != "-" {
		if n, err := strconv.Atoi(sk); err != nil {
			add("komsifra", komMsgKomercijalista)
		} else if _, found, err := s.repo.KomKomercijalista(ctx, n); err != nil {
			add("komsifra", common.ErrMsgReadData)
		} else if !found {
			add("komsifra", komMsgKomercijalista)
		}
	}
	return errs
}

// SavePorudzbenica saves the header of a validated porudžbenica (the legacy SnimiEDOK) in one transaction: a new one
// with the next broj of its vrsta in its magacin, the oznaka of its magacin and its zahtev za ponudu; a saved one with
// its datum, rok, dobavljač, komercijalista, org. jedinica, mesto troška, tekst, valuta and uslovi nabavke (its vrsta,
// broj, magacin and zahtev za ponudu stay). It returns the rdokid.
func (s *PorudzbenicaResource) SavePorudzbenica(ctx context.Context, params domain.KomPorParams) (int64, error) {
	datum, _ := komDatum(params.Datum)
	p := domain.KomPorudzbenicaUpis{RdokID: params.RdokID, Vrd: common.StringToInt(params.Vrd), MagaciniID: common.StringToInt(params.MagaciniID),
		Dadok: datum, Rok: common.StringToInt(params.Rok), Fkto: strings.TrimSpace(params.Fkto), Fana: strings.TrimSpace(params.Fana),
		Orgjed: komNullID(params.Orgjed), MestoTroska: komNullID(params.MestoTroska), Foot: strings.TrimSpace(params.Tekst),
		SifVal: common.StringToInt(params.SifVal), Kurs: komBroj(params.Kurs, 1), VaziZa: int(komBroj(params.VaziZa, 1)),
		Otp: strings.TrimSpace(params.Otp), Pla: strings.TrimSpace(params.Pla), Brugov: strings.TrimSpace(params.Brugov),
		Nacpak: strings.TrimSpace(params.Nacpak), Posuslnab: strings.TrimSpace(params.Posuslnab), Pratdok: strings.TrimSpace(params.Pratdok)}
	if id, given, ok := komID(params.VrdokID); ok && given {
		p.VrdokID = id
	}
	if sifra := common.StringToInt(strings.TrimSpace(params.KomSifra)); sifra > 0 {
		k, found, err := s.repo.KomKomercijalista(ctx, sifra)
		if err != nil {
			return 0, err
		}
		if found {
			p.KomID = sql.NullInt64{Int64: k.KomID, Valid: true}
		}
	}
	if p.RdokID > 0 {
		saved, found, err := s.repo.KomPorudzbenica(ctx, p.RdokID)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("porudžbenica %d ne postoji", p.RdokID)
		}
		p.Vrd, p.Dokum, p.MagaciniID, p.Mag, p.VrdokID = saved.Vrd, saved.Dokum, saved.MagaciniID, saved.Mag, saved.VrdokID
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if p.RdokID > 0 {
		if err := s.repo.KomPorIzmeni(ctx, tx, p); err != nil {
			return 0, err
		}
	} else {
		mag, found, err := s.repo.KomMagacin(ctx, p.MagaciniID)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("magacin %d ne postoji", p.MagaciniID)
		}
		p.Mag = mag
		if p.Dokum, err = s.repo.KomPorNoviBroj(ctx, tx, p.Vrd, mag); err != nil {
			return 0, err
		}
		if p.RdokID, err = s.repo.KomPorDodaj(ctx, tx, p); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit porudžbenica: %w", err)
	}
	return p.RdokID, nil
}

// DeletePorudzbenica deletes a porudžbenica with its stavke, their stavke of the zahtev za ponudu slobodne again (the
// legacy BTN_Brisanje): not a fiscalized (knjige 4 "S"), posted (knjige 1 "D") or confirmed one, whose message is
// returned instead.
func (s *PorudzbenicaResource) DeletePorudzbenica(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.repo.KomPorudzbenica(ctx, rdokID)
	if err != nil {
		return "", err
	}
	switch {
	case !found:
		return "Niste selektovali dokument za brisanje!!!!", nil
	case strings.EqualFold(strings.TrimSpace(p.Knjige4), "S"):
		return komMsgFiskalizovan, nil
	case strings.EqualFold(strings.TrimSpace(p.Knjige1), "D"):
		return komMsgProknjizen, nil
	case p.Doksts == domain.KomZahtevPotvrdjen:
		return komMsgPorPotvrdjenBris, nil
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomPorBrisi(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje porudžbenice: %w", err)
	}
	return "", nil
}

// ProveriPotvrdu checks a porudžbenica before the question of its potvrda (the legacy BTN_NoviDOK1 "Potvrda dok."): it
// has stavke, the cena of each one is greater than 0 and it is nepotvrđena; the message why it cannot be confirmed is
// returned.
func (s *PorudzbenicaResource) ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.repo.KomPorudzbenica(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgBezStavkiPotvrda, nil
	}
	stavke, err := s.repo.KomPorStavke(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if len(stavke) == 0 {
		return komMsgBezStavkiPotvrda, nil
	}
	for _, st := range stavke {
		if st.Cena <= 0 {
			return komMsgPorCenaNula, nil
		}
	}
	if p.Doksts != domain.KomZahtevNepotvrdjen {
		return komMsgVecPotvrdjen, nil
	}
	return "", nil
}

// PotvrdiPorudzbenicu confirms a porudžbenica (the legacy "Da" of "Potvrda dok.": doksts 40) after the checks of
// ProveriPotvrdu; it returns their message, or "Dokument je potvrđen!!!".
func (s *PorudzbenicaResource) PotvrdiPorudzbenicu(ctx context.Context, rdokID int64) (string, error) {
	if poruka, err := s.ProveriPotvrdu(ctx, rdokID); err != nil || poruka != "" {
		return poruka, err
	}
	ok, err := s.repo.KomPorPotvrdi(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !ok {
		return komMsgVecPotvrdjen, nil
	}
	return komMsgPotvrdjen, nil
}

// GetStavke returns the stavke of a porudžbenica (the legacy PuniDETALJ1) and whether its prices are in a foreign
// valuta.
func (s *PorudzbenicaResource) GetStavke(ctx context.Context, rdokID int64) ([]domain.KomPorStavkaRow, bool, error) {
	if rdokID <= 0 {
		return nil, false, nil
	}
	p, found, err := s.repo.KomPorudzbenica(ctx, rdokID)
	if err != nil || !found {
		return nil, false, err
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return nil, false, err
	}
	stavke, err := s.repo.KomPorStavke(ctx, rdokID)
	return stavke, ino, err
}

// ino tells whether the valuta is a foreign one for the firma of the session.
func (s *PorudzbenicaResource) ino(ctx context.Context, sifval int) (bool, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return false, fmt.Errorf("no user session found")
	}
	sifvalFirme, _, err := komValutaFirme(ctx, s.repo, session.Firma)
	return komIno(sifval, sifvalFirme), err
}

// GetArtikal returns the artikal of a stavka with its nabavna cena (the legacy exit of EDT_SIFRA: rcene.ncena); found is
// false without it.
func (s *PorudzbenicaResource) GetArtikal(ctx context.Context, sifra string) (domain.KomArtikal, bool, error) {
	n, err := strconv.Atoi(strings.TrimSpace(sifra))
	if err != nil || n <= 0 {
		return domain.KomArtikal{}, false, nil
	}
	return s.repo.KomArtikal(ctx, n)
}

// porStavka returns the stavka of the id of a porudžbenica; found is false without it.
func (s *PorudzbenicaResource) porStavka(ctx context.Context, rdokID, rproID int64) (domain.KomPorStavkaRow, bool, error) {
	stavke, err := s.repo.KomPorStavke(ctx, rdokID)
	if err != nil {
		return domain.KomPorStavkaRow{}, false, err
	}
	for _, st := range stavke {
		if st.RproID == rproID {
			return st, true, nil
		}
	}
	return domain.KomPorStavkaRow{}, false, nil
}

// ValidateStavka validates a stavka of a porudžbenica (the legacy ValidacijaEPRO): its porudžbenica is a nepotvrđena
// one, the količina and the cena are greater than 0; a new stavka has an artikal of the šifra, a changed one (rproid)
// exists.
func (s *PorudzbenicaResource) ValidateStavka(ctx context.Context, params domain.KomPorStavkaParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if p, found, err := s.repo.KomPorudzbenica(ctx, params.RdokID); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !found {
		add("sifraartikla", komMsgNePostoji)
	} else if p.Doksts != "" && p.Doksts != domain.KomZahtevNepotvrdjen {
		add("sifraartikla", komMsgPorNijeIzmena)
	}
	if params.RproID > 0 {
		if _, found, err := s.porStavka(ctx, params.RdokID, params.RproID); err != nil {
			add("sifraartikla", common.ErrMsgReadData)
		} else if !found {
			add("sifraartikla", komMsgNePostoji)
		}
	} else if _, found, err := s.GetArtikal(ctx, params.Sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !found {
		add("sifraartikla", komMsgArtikal)
	}
	if common.StringToFloat64(params.Kolicina) <= 0 {
		add("kolicina", komMsgKolicina)
	}
	if common.StringToFloat64(params.Cena) <= 0 {
		add("cena", komMsgPorCena)
	}
	return errs
}

// komCene returns the cena of a stavka in RSD (cena and ncena) and in the valuta (cenaval) of an entered cena (the
// legacy AddEPRO): in a foreign valuta the entered cena is the cenaval and the cena is converted by the kurs and the
// "važi za"; else all of them are the entered cena.
func komCene(unesena float64, ino bool, kurs float64, vaziza int) (cena, cenaVal float64) {
	if !ino {
		return unesena, unesena
	}
	if vaziza <= 0 {
		vaziza = 1
	}
	return unesena * kurs / float64(vaziza), unesena
}

// SaveStavka saves a validated stavka of a porudžbenica in one transaction: a new one entered by hand with the next
// redni broj (the legacy AddEPRO: the magacin and the vrsta, broj and datum of the porudžbenica, no stavka of a zahtev
// za ponudu), a changed one (rproid) with its new količina, cena and rabat.
func (s *PorudzbenicaResource) SaveStavka(ctx context.Context, params domain.KomPorStavkaParams) error {
	p, found, err := s.repo.KomPorudzbenica(ctx, params.RdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("porudžbenica %d ne postoji", params.RdokID)
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return err
	}
	kolicina, rabat := common.StringToFloat64(params.Kolicina), common.StringToFloat64(params.Rabat)
	cena, cenaVal := komCene(common.StringToFloat64(params.Cena), ino, p.Kurs, p.VaziZa)
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if params.RproID > 0 {
		if err := s.repo.KomPorStavkaIzmeni(ctx, tx, p.RdokID, params.RproID, kolicina, cena, cenaVal, rabat); err != nil {
			return err
		}
	} else {
		a, found, err := s.GetArtikal(ctx, params.Sifra)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("artikal %s ne postoji", params.Sifra)
		}
		rbr, err := s.repo.KomZahtevNoviRbr(ctx, tx, p.RdokID)
		if err != nil {
			return err
		}
		st := domain.KomPorStavkaUpis{RdokID: p.RdokID, Rbr: rbr, Sifra: a.Sifra, RsifID: a.RsifID, Po: a.Po, Kolicina: kolicina, Cena: cena,
			CenaVal: cenaVal, Rabat: rabat, Mag: p.Mag, MagaciniID: p.MagaciniID, Vrd: p.Vrd, Dokum: p.Dokum, Dadok: p.Dadok}
		if err := s.repo.KomPorStavkaDodaj(ctx, tx, st); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka porudžbenice: %w", err)
	}
	return nil
}

// DeleteStavka deletes a stavka of a porudžbenica (the legacy BTN_BrisiDET) in one transaction: its stavka of the
// zahtev za ponudu is slobodna again and the redni brojevi after it move one back.
func (s *PorudzbenicaResource) DeleteStavka(ctx context.Context, rdokID, rproID int64) error {
	st, found, err := s.porStavka(ctx, rdokID, rproID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("stavka %d ne postoji", rproID)
	}
	return s.brisiStavku(ctx, rdokID, st)
}

// brisiStavku deletes a stavka in one transaction (its source slobodna again, the redni brojevi after it moved back).
func (s *PorudzbenicaResource) brisiStavku(ctx context.Context, rdokID int64, st domain.KomPorStavkaRow) error {
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomPonudaStavkaBrisi(ctx, tx, rdokID, st.RproID, st.Rbr); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit brisanje stavke porudžbenice: %w", err)
	}
	return nil
}

// GetStavkaForma returns a saved stavka for the form of the stavke ("Izmeni" of its row) with its cena in the valuta of
// the porudžbenica; found is false without it.
func (s *PorudzbenicaResource) GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomPorStavkaForma, bool, error) {
	stavke, ino, err := s.GetStavke(ctx, rdokID)
	if err != nil {
		return domain.KomPorStavkaForma{}, false, err
	}
	for _, st := range stavke {
		if st.RproID == rproID {
			return domain.KomPorStavkaForma{RproID: st.RproID, Rbr: st.Rbr, Sifra: st.Sifra, Naziv: st.Naziv, JM: st.JM, Kolicina: st.Kolicina,
				Cena: komCenaPrikaz(st, ino), Rabat: st.Rabat}, true, nil
		}
	}
	return domain.KomPorStavkaForma{}, false, nil
}

// komCenaPrikaz is the cena of a stavka in the valuta of its porudžbenica (cenaval in a foreign valuta, else cena).
func komCenaPrikaz(st domain.KomPorStavkaRow, ino bool) float64 {
	if ino {
		return st.CenaVal
	}
	return st.Cena
}

// komVrednost is the vrednost of a stavka: the količina by the cena less the rabat (the legacy ITEM_VRED: the rabat
// rounded to 2 decimals).
func komVrednost(kolicina, cena, rabat float64) float64 {
	iznos := kolicina * cena
	return iznos - math.Round(iznos*rabat/100*100)/100
}

// Preuzmi takes the slobodne stavke of the zahtev za ponudu into a saved nepotvrđena porudžbenica (the legacy BTN_Dodaj
// with the id of a stavka, BTN_DodajSve with 0) in one transaction: each one not taken yet becomes a stavka with the
// next redni broj, its artikal, its tražena količina and its cena in the valuta and its rproid in rproid1, and is
// marked rezervisana; without one the message why is returned.
func (s *PorudzbenicaResource) Preuzmi(ctx context.Context, rdokID, rproID int64) (string, error) {
	p, found, err := s.repo.KomPorudzbenica(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgNePostoji, nil
	}
	if p.Doksts != "" && p.Doksts != domain.KomZahtevNepotvrdjen {
		return komMsgPorNijeIzmena, nil
	}
	if p.VrdokID <= 0 {
		return komMsgPorNemaZaDodavanje, nil
	}
	izvori, err := s.repo.KomPorIzvori(ctx, p.VrdokID)
	if err != nil {
		return "", err
	}
	var ids []int64
	for _, iz := range izvori {
		if rproID <= 0 || iz.RproID == rproID {
			ids = append(ids, iz.RproID)
		}
	}
	if len(ids) == 0 {
		return komMsgPorNemaZaDodavanje, nil
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return "", err
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	for _, id := range ids {
		iz, found, err := s.repo.KomPorIzvor(ctx, tx, p.VrdokID, id)
		if err != nil {
			return "", err
		}
		if !found {
			continue
		}
		rbr, err := s.repo.KomZahtevNoviRbr(ctx, tx, p.RdokID)
		if err != nil {
			return "", err
		}
		cena, cenaVal := komCene(iz.CenaVal, ino, p.Kurs, p.VaziZa)
		st := domain.KomPorStavkaUpis{RdokID: p.RdokID, Rbr: rbr, Sifra: iz.Sifra, RsifID: iz.RsifID, Po: iz.Po, Kolicina: iz.Kolicina, Cena: cena,
			CenaVal: cenaVal, Mag: p.Mag, MagaciniID: p.MagaciniID, Vrd: p.Vrd, Dokum: p.Dokum, Dadok: p.Dadok, RproID1: iz.RproID}
		if err := s.repo.KomPorStavkaDodaj(ctx, tx, st); err != nil {
			return "", err
		}
		if err := s.repo.KomPonudaIzvorRezervisi(ctx, tx, iz.RproID); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit stavke porudžbenice: %w", err)
	}
	return "", nil
}

// VratiPoslednju deletes the last stavka of a porudžbenica (the legacy BTN_VratiJedan; its source slobodna again);
// without stavke the message why is returned.
func (s *PorudzbenicaResource) VratiPoslednju(ctx context.Context, rdokID int64) (string, error) {
	stavke, err := s.repo.KomPorStavke(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if len(stavke) == 0 {
		return komMsgPorNemaZaBrisanje, nil
	}
	poslednja := stavke[0]
	for _, st := range stavke {
		if st.Rbr > poslednja.Rbr {
			poslednja = st
		}
	}
	return "", s.brisiStavku(ctx, rdokID, poslednja)
}

// VratiSve deletes all the stavke of a porudžbenica (the legacy BTN_VratiSve; their sources slobodne again) in one
// transaction; without stavke the message why is returned.
func (s *PorudzbenicaResource) VratiSve(ctx context.Context, rdokID int64) (string, error) {
	stavke, err := s.repo.KomPorStavke(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if len(stavke) == 0 {
		return komMsgPorNemaZaBrisanje, nil
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomPorVratiSve(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje stavki porudžbenice: %w", err)
	}
	return "", nil
}

// GetPrikaz returns the porudžbenice of the tab "Pregled porudžbenice" (the legacy PR_QRY_EDOK).
func (s *PorudzbenicaResource) GetPrikaz(ctx context.Context, params domain.KomPonudaPrikazParams) ([]domain.KomPorPrikazRow, error) {
	if params.Doksts != domain.KomZahtevPotvrdjen {
		params.Doksts = domain.KomZahtevNepotvrdjen
	}
	if params.Vrd <= 0 || (!params.SviMagacini && params.MagaciniID <= 0) {
		return nil, nil
	}
	return s.repo.KomPorPrikaz(ctx, params)
}

// GetStampa returns the print of a porudžbenica (the legacy PR_RPT_PORUDZB_DOB, PR_RPT_PORUDZB_DOBINO in a foreign
// valuta): the firm with its tekući računi (the posebni uslovi nabavke instead of them when given, like the legacy),
// the dobavljač, the header, the uslovi nabavke and the stavke, whose opis is the naziv and/or the komercijalni opis of
// the artikal; a porudžbenica without stavke is not printed (the message why is returned).
func (s *PorudzbenicaResource) GetStampa(ctx context.Context, rdokID int64, naziv, komercopis bool) (domain.KomPorStampa, string, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomPorStampa{}, "", fmt.Errorf("no user session found")
	}
	p, found, err := s.repo.KomPorudzbenica(ctx, rdokID)
	if err != nil {
		return domain.KomPorStampa{}, "", err
	}
	if !found {
		return domain.KomPorStampa{}, komMsgBezStavkiStampa, nil
	}
	stavke, err := s.repo.KomPorStavke(ctx, rdokID)
	if err != nil {
		return domain.KomPorStampa{}, "", err
	}
	if len(stavke) == 0 {
		return domain.KomPorStampa{}, komMsgBezStavkiStampa, nil
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return domain.KomPorStampa{}, "", err
	}
	stampa := domain.KomPorStampa{Ino: ino, OznVal: p.OznVal, Dokum: p.Dokum, Datum: p.Dadok, Rok: p.Rok, Napomena: p.Foot, Otp: p.Otp, Pla: p.Pla,
		Brugov: p.Brugov, Nacpak: p.Nacpak, Posuslnab: p.Posuslnab, Magacin: p.MagacinText, MagacinAdresa: p.MagacinAdresa, OJ: p.OrgjedText}
	if p.KomIme != "" {
		stampa.Referent = "(" + strings.TrimSpace(p.KomIme) + ")"
	}
	if idFkpl, err := s.repo.KomPorIDFkpl(ctx, p.Fkto, p.Fana); err != nil {
		return domain.KomPorStampa{}, "", err
	} else if idFkpl > 0 {
		if stampa.Dobavljac, _, err = s.repo.KomPonudaStampaDobavljac(ctx, idFkpl); err != nil {
			return domain.KomPorStampa{}, "", err
		}
	}
	if stampa.Firma, err = s.repo.KomFirma(ctx, session.Firma); err != nil {
		return domain.KomPorStampa{}, "", err
	}
	if u := strings.TrimSpace(p.Posuslnab); !ino && u != "" && u != "-" && u != "-1" {
		stampa.ZiroRacuni = []string{u}
	} else if stampa.ZiroRacuni, err = s.repo.BankeZiroRacuni(ctx); err != nil {
		return domain.KomPorStampa{}, "", err
	}
	for _, st := range stavke {
		var opis []string
		if naziv {
			opis = append(opis, st.Naziv)
		}
		if komercopis && strings.TrimSpace(st.Komercopis) != "" {
			opis = append(opis, strings.TrimSpace(st.Komercopis))
		}
		cena := komCenaPrikaz(st, ino)
		vrednost := komVrednost(st.Kolicina, cena, st.Rabat)
		stampa.Stavke = append(stampa.Stavke, domain.KomPorStampaStavka{Rbr: st.Rbr, Sifra: st.Sifra, Opis: strings.Join(opis, " "), JM: st.JM,
			Kolicina: st.Kolicina, Decimale: st.Decimale, Cena: cena, Rabat: st.Rabat, Vrednost: vrednost})
		stampa.Ukupno += vrednost
	}
	return stampa, "", nil
}

// StavkeTableFields are the columns of the grid of the stavke of the porudžbenica, in one line like the fakture (the
// legacy TABLE_PR_QRY_EPRO1 without the proizvođač, the proizvođačka šifra and the komercijalni opis): the cena is the
// one of the valuta of the document.
func (s *PorudzbenicaResource) StavkeTableFields(ino bool, oznval string) []domain.Fields {
	cena := domain.Fields{Name: "cena", Label: "Jedinična cena", Width: "11", TextAlign: "right"}
	if ino && strings.TrimSpace(oznval) != "" {
		cena.Label, cena.Params = komCenaUValuti, map[string]string{"valuta": strings.ToUpper(strings.TrimSpace(oznval))}
	}
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "6", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "9", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "36"},
		{Name: "jm", Label: "JM", Width: "5"},
		{Name: "kolicina", Label: "Količina za nabavku", Width: "11", TextAlign: "right"},
		cena,
		{Name: "rabat", Label: "% Rabata", Width: "7", TextAlign: "right"},
		{Name: "vrednost", Label: "Vrednost", Width: "12", TextAlign: "right"},
	}
}

// IzvoriTableFields are the columns of the slobodne stavke of the zahtev za ponudu (the legacy "Količine po ponudi").
func (s *PorudzbenicaResource) IzvoriTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "8", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "12", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla / sirovine", Width: "50"},
		{Name: "jm", Label: "JM", Width: "8"},
		{Name: "kolicina", Label: "Tražena količina", Width: "14", TextAlign: "right"},
	}
}

// PrikazTableFields are the columns of the porudžbenice of the tab "Pregled porudžbenice" (the legacy TABLE_QRY_EDOK).
func (s *PorudzbenicaResource) PrikazTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "magacin", Label: "Magacin", Width: "14"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "6", TextAlign: "right"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "7", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "9"},
		{Name: "fkto", Label: "Konto partnera", Width: "7"},
		{Name: "fana", Label: "Šifra", Width: "7"},
		{Name: "dobavljac", Label: "Naziv partnera", Width: "24"},
		{Name: "rok", Label: "Rok nabavke", Width: "6", TextAlign: "right"},
		{Name: "otp", Label: "Način dopremanja", Width: "12"},
		{Name: "referent", Label: "Referent nabavke", Width: "12"},
	}
}

// PrikazStavkeTableFields are the columns of the stavke of the selected porudžbenica of the tab "Pregled porudžbenice"
// (the legacy TABLE_QRY_EPRO) with the total of the vrednost.
func (s *PorudzbenicaResource) PrikazStavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "6", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "9", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "36"},
		{Name: "jm", Label: "JM", Width: "5"},
		{Name: "kolicina", Label: "Naručena količina", Width: "11", TextAlign: "right"},
		{Name: "cena", Label: "Jedinična cena", Width: "11", TextAlign: "right"},
		{Name: "rabat", Label: "% Rabata", Width: "7", TextAlign: "right"},
		{Name: "vrednost", Label: "Vrednost", Width: "12", TextAlign: "right", IncludeInTotals: true},
	}
}

// PorStavkeTableRows are the rows of the stavke of a porudžbenica (the cena in its valuta, the vrednost less the rabat)
// with "Izmeni" and "Obriši" of every row (izmena), and the total of the vrednost.
func PorStavkeTableRows(stavke []domain.KomPorStavkaRow, ino, izmena bool) ([]domain.TableRow, float64) {
	rows := make([]domain.TableRow, 0, len(stavke))
	ukupno := 0.0
	for _, st := range stavke {
		cena := komCenaPrikaz(st, ino)
		vrednost := komVrednost(st.Kolicina, cena, st.Rabat)
		ukupno += vrednost
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), HasUpdate: izmena, HasDelete: izmena, Fields: []string{
			strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM, common.FormatNumberWithSystemLocale(st.Kolicina, 3),
			common.FormatNumberWithSystemLocale(cena, 2), common.FormatNumberWithSystemLocale(st.Rabat, 2),
			common.FormatNumberWithSystemLocale(vrednost, 2),
		}})
	}
	return rows, ukupno
}

// PorIzvoriTableRows are the rows of the slobodne stavke of the zahtev za ponudu.
func PorIzvoriTableRows(izvori []domain.KomPorIzvorRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(izvori))
	for _, iz := range izvori {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(iz.RproID, 10), Fields: []string{
			strconv.Itoa(iz.Rbr), strconv.Itoa(iz.Sifra), iz.Naziv, iz.JM, common.FormatNumberWithSystemLocale(iz.Kolicina, 3),
		}})
	}
	return rows
}

// PorPrikazTableRows are the rows of the porudžbenice of the tab "Pregled porudžbenice".
func PorPrikazTableRows(porudzbenice []domain.KomPorPrikazRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(porudzbenice))
	for _, p := range porudzbenice {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(p.RdokID, 10), Fields: []string{
			p.Magacin, strconv.Itoa(p.Vrd), strconv.Itoa(p.Dokum), komDatumText(p.Dadok), p.Fkto, p.Fana, p.Dobavljac, strconv.Itoa(p.Rok), p.Otp,
			p.Referent,
		}})
	}
	return rows
}
