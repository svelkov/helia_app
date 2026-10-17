// Package komercijalno holds the business logic of the komercijalno module.
package komercijalno

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	commonsvc "helia/internal/service/common"
)

// The messages of the zahtev za nabavku (the legacy pr_NarudzbenicaPROIZV).
const (
	komMsgNePostoji        = "Ne postoji."
	komMsgDatum            = "Morate uneti datum dokumenta..."
	komMsgMagacin          = "Niste izabrali magacin!!!"
	komMsgVrsta            = "Niste izabrali vrstu dokumenta!!!"
	komMsgKomercijalista   = "Komercijalista sa takvom šifrom ne postoji!!!..."
	komMsgArtikal          = "Artikal sa izabranom šifrom ne postoji!!!"
	komMsgKolicina         = "Količina mora biti veća od 0!!!!"
	komMsgMestoTroskaOJ    = "Nije mesto troška izabrane OJ."
	komMsgFiskalizovan     = "Ovaj dokument je odštampan na fiskalni printer, brisanje nije dozvoljeno!!!"
	komMsgProknjizen       = "Ovaj dokument je proknjižen, brisanje nije dozvoljeno!!!"
	komMsgBezStavkiPotvrda = "Ne možete potvrditi dokument zato što nema nijednu stavku!!!"
	komMsgVecPotvrdjen     = "Dokument je već potvrđen..."
	komMsgPotvrdjen        = "Dokument je potvrđen!!!"
	komMsgBezStavkiStampa  = "Ne možete odštampati dokument zato što nema nijednu stavku..."
	komMsgNijeIzmena       = "Potvrđen zahtev za nabavku se ne može menjati."
	komMsgObrisano         = "Uspešno obrisani podaci..."
)

// The magacini of a zahtev za nabavku: the ones of the user of the tipovi P, S and D (the legacy ROB_QRY_MAGUSER with
// "P;S;D").
var komZahtevTipoviMagacina = []string{"P", "S", "D"}

// ZahtevNabavkaService is the "Zahtev za nabavku" screen of the komercijalno (the legacy pr_NarudzbenicaPROIZV).
type ZahtevNabavkaService interface {
	GetZahtev(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomZahtevDokument, error)
	ValidateZahtev(ctx context.Context, params domain.KomZahtevParams) []domain.FieldError
	SaveZahtev(ctx context.Context, params domain.KomZahtevParams) (int64, error)
	DeleteZahtev(ctx context.Context, rdokID int64) (string, error)
	ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error)
	PotvrdiZahtev(ctx context.Context, rdokID int64) (string, error)
	GetStavke(ctx context.Context, rdokID int64) ([]domain.KomZahtevStavkaRow, error)
	GetArtikal(ctx context.Context, sifra string) (domain.KomArtikal, bool, error)
	ValidateStavka(ctx context.Context, params domain.KomZahtevStavkaParams) []domain.FieldError
	SaveStavka(ctx context.Context, params domain.KomZahtevStavkaParams) error
	DeleteStavka(ctx context.Context, rdokID, rproID int64) error
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaForma, bool, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	GetVrste(ctx context.Context) ([]domain.ComboItem, error)
	GetPrikaz(ctx context.Context, params domain.KomZahtevPrikazParams) ([]domain.KomZahtevPrikazRow, error)
	GetStampa(ctx context.Context, rdokID int64) (domain.KomZahtevStampa, string, error)
	StavkeTableFields() []domain.Fields
	PrikazTableFields() []domain.Fields
	PrikazStavkeTableFields() []domain.Fields
}

// zahtevRepository is what the service needs of repository.KomercijalnoRepository.
type zahtevRepository interface {
	BeginTx() (db.Transaction, error)
	KomZahtevVrste(ctx context.Context) ([]domain.ComboItem, error)
	KomZahtevNoviBroj(ctx context.Context, tx db.Transaction, vrd int) (int, error)
	KomZahtev(ctx context.Context, rdokID int64) (domain.KomZahtevRow, bool, error)
	KomZahtevDodaj(ctx context.Context, tx db.Transaction, z domain.KomZahtevUpis) (int64, error)
	KomZahtevIzmeni(ctx context.Context, tx db.Transaction, z domain.KomZahtevUpis) error
	KomZahtevPotvrdi(ctx context.Context, rdokID int64) (bool, error)
	KomZahtevBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomZahtevStavke(ctx context.Context, rdokID int64) ([]domain.KomZahtevStavkaRow, error)
	KomZahtevStavka(ctx context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaRow, bool, error)
	KomZahtevNoviRbr(ctx context.Context, tx db.Transaction, rdokID int64) (int, error)
	KomZahtevStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomZahtevStavkaUpis) error
	KomZahtevStavkaIzmeni(ctx context.Context, tx db.Transaction, st domain.KomZahtevStavkaUpis) error
	KomZahtevStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error
	KomArtikal(ctx context.Context, sifra int) (domain.KomArtikal, bool, error)
	KomKomercijalista(ctx context.Context, sifra int) (domain.KomKomercijalista, bool, error)
	KomKomercijalistaKorisnika(ctx context.Context, userID int64) (domain.KomKomercijalista, bool, error)
	KomMagacin(ctx context.Context, magaciniID int) (int, bool, error)
	KomZahteviPrikaz(ctx context.Context, p domain.KomZahtevPrikazParams) ([]domain.KomZahtevPrikazRow, error)
	KomZahtevStampaZaglavlje(ctx context.Context, rdokID int64) (domain.KomZahtevStampa, bool, error)
	KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error)
	OrgjedPostoji(ctx context.Context, id int64) (bool, error)
	MestotrPostoji(ctx context.Context, id int64) (bool, error)
	MestotrUOrgjed(ctx context.Context, mestotrID, orgjedID int64) (bool, error)
}

// magaciniKorisnika is what the service needs of the common service: the magacini of the user.
type magaciniKorisnika interface {
	GetMagacinKorisnikaComboValues(ctx context.Context, opts ...commonsvc.ComboOption) ([]domain.ComboItem, error)
}

// ZahtevNabavkaResource implements ZahtevNabavkaService.
type ZahtevNabavkaResource struct {
	repo   zahtevRepository
	common magaciniKorisnika
	now    func() time.Time
}

// NewZahtevNabavkaService returns the service of the "Zahtev za nabavku" screen.
func NewZahtevNabavkaService(repo zahtevRepository, common magaciniKorisnika) *ZahtevNabavkaResource {
	return &ZahtevNabavkaResource{repo: repo, common: common, now: time.Now}
}

// GetMagacini returns the magacini of a zahtev za nabavku: the ones the user of the session may work with of the tipovi
// P, S and D, "mag - opis" keyed by magaciniid (the legacy PuniVRD with ROB_QRY_MAGUSER).
func (s *ZahtevNabavkaResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.common.GetMagacinKorisnikaComboValues(ctx, commonsvc.WithIn("magacini.tipmag", komZahtevTipoviMagacina...))
}

// GetVrste returns the vrste dokumenata of the zahtevi za nabavku (the filter of the main screen).
func (s *ZahtevNabavkaResource) GetVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return s.repo.KomZahtevVrste(ctx)
}

// GetZahtev returns the content of the tab "Zahtev za nabavku" with the zahtev of the id (a nepotvrđen one, the legacy
// BTN_TraziDOK and FileToScreen) or, for the id 0, a new one (the legacy BTN_NoviDOK): of the vrsta and the magacin
// (the first ones without them), the next broj of the vrsta (NoviBrojEDOK), the datum of today and the komercijalista
// of the user. A confirmed zahtev cannot be changed: a new one is given with the message why.
func (s *ZahtevNabavkaResource) GetZahtev(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomZahtevDokument, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomZahtevDokument{}, fmt.Errorf("no user session found")
	}
	var dok domain.KomZahtevDokument
	var err error
	if dok.Vrste, err = s.repo.KomZahtevVrste(ctx); err != nil {
		return dok, err
	}
	if dok.Magacini, err = s.GetMagacini(ctx); err != nil {
		return dok, err
	}
	if rdokID > 0 {
		z, found, err := s.repo.KomZahtev(ctx, rdokID)
		if err != nil {
			return dok, err
		}
		switch {
		case !found:
			dok.Poruka = komMsgNePostoji
		case z.Doksts != "" && z.Doksts != domain.KomZahtevNepotvrdjen:
			dok.Poruka = komMsgNijeIzmena
			vrd, magaciniID = z.Vrd, z.MagaciniID
		default:
			dok.Header = komZahtevHeader(z)
			if dok.NoviRbr, err = s.repo.KomZahtevNoviRbr(ctx, nil, z.RdokID); err != nil {
				return dok, err
			}
			return dok, nil
		}
	}
	h := &dok.Header
	h.Vrd = komIzListe(dok.Vrste, vrd)
	h.MagaciniID = komIzListe(dok.Magacini, magaciniID)
	h.Datum = s.now()
	h.Opunos = session.UserName
	if h.Vrd > 0 {
		if h.Dokum, err = s.repo.KomZahtevNoviBroj(ctx, nil, h.Vrd); err != nil {
			return dok, err
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

// komZahtevHeader is the header of the screen of a saved zahtev.
func komZahtevHeader(z domain.KomZahtevRow) domain.KomZahtevHeaderView {
	h := domain.KomZahtevHeaderView{RdokID: z.RdokID, Vrd: z.Vrd, Dokum: z.Dokum, MagaciniID: z.MagaciniID, KomSifra: z.KomSifra,
		KomIme: z.KomIme, Opunos: z.Opunos, Orgjed: z.Orgjed, OrgjedText: z.OrgjedText, MestoTroska: z.MestoTroska,
		MestoTroskaText: z.MestoTroskaText, Tekst: z.Foot, Potvrdjen: z.Doksts == domain.KomZahtevPotvrdjen}
	if z.Dadok.Valid {
		h.Datum = z.Dadok.Time
	}
	return h
}

// komIzListe returns the key of the list that is the given one, else the first key of the list (0 for an empty list).
func komIzListe(lista []domain.ComboItem, kljuc int) int {
	for _, item := range lista {
		if common.StringToInt(item.Key) == kljuc && kljuc > 0 {
			return kljuc
		}
	}
	if len(lista) > 0 {
		return common.StringToInt(lista[0].Key)
	}
	return 0
}

// komUList tells whether the key is in the list.
func komUListi(lista []domain.ComboItem, kljuc int) bool {
	for _, item := range lista {
		if kljuc > 0 && common.StringToInt(item.Key) == kljuc {
			return true
		}
	}
	return false
}

// ValidateZahtev validates the header of a zahtev za nabavku (the legacy ValidacijaEDOK): the vrsta and the magacin
// are ones of the screen, the datum is a valid date, a given komercijalista exists, a given org. jedinica and mesto
// troška exist (the mesto troška of the org. jedinica; both not required, like the other documents); a saved zahtev
// is nepotvrđen.
func (s *ZahtevNabavkaResource) ValidateZahtev(ctx context.Context, params domain.KomZahtevParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if params.RdokID > 0 {
		z, found, err := s.repo.KomZahtev(ctx, params.RdokID)
		switch {
		case err != nil:
			add("brdok", common.ErrMsgReadData)
		case !found:
			add("brdok", komMsgNePostoji)
		case z.Doksts != "" && z.Doksts != domain.KomZahtevNepotvrdjen:
			add("brdok", komMsgNijeIzmena)
		}
	} else {
		if vrste, err := s.repo.KomZahtevVrste(ctx); err != nil {
			add("vrd", common.ErrMsgReadData)
		} else if !komUListi(vrste, common.StringToInt(params.Vrd)) {
			add("vrd", komMsgVrsta)
		}
		if magacini, err := s.GetMagacini(ctx); err != nil {
			add("magaciniid", common.ErrMsgReadData)
		} else if !komUListi(magacini, common.StringToInt(params.MagaciniID)) {
			add("magaciniid", komMsgMagacin)
		}
	}
	if _, ok := komDatum(params.Datum); !ok {
		add("datum", komMsgDatum)
	}
	if sifra := strings.TrimSpace(params.KomSifra); sifra != "" && sifra != "0" && sifra != "-" {
		if n, err := strconv.Atoi(sifra); err != nil {
			add("komsifra", komMsgKomercijalista)
		} else if _, found, err := s.repo.KomKomercijalista(ctx, n); err != nil {
			add("komsifra", common.ErrMsgReadData)
		} else if !found {
			add("komsifra", komMsgKomercijalista)
		}
	}
	s.komOrgjedMt(ctx, params.Orgjed, params.MestoTroska, add)
	return errs
}

// komOrgjedMt checks a given org. jedinica and mesto troška (not required): they exist and the mesto troška is one of
// the org. jedinica.
func (s *ZahtevNabavkaResource) komOrgjedMt(ctx context.Context, oj, mt string, add func(field, message string)) {
	komProveriOrgjedMt(ctx, s.repo, oj, mt, add)
}

// komOrgjedRepository is what the check of an org. jedinica and a mesto troška needs of the repository.
type komOrgjedRepository interface {
	OrgjedPostoji(ctx context.Context, id int64) (bool, error)
	MestotrPostoji(ctx context.Context, id int64) (bool, error)
	MestotrUOrgjed(ctx context.Context, mestotrID, orgjedID int64) (bool, error)
}

// komProveriOrgjedMt checks a given org. jedinica and mesto troška: they exist and the mesto troška is one of the org.
// jedinica.
func komProveriOrgjedMt(ctx context.Context, repo komOrgjedRepository, oj, mt string, add func(field, message string)) {
	ojID, ojGiven, ojOK := komID(oj)
	if !ojOK {
		add("idorgjed", komMsgNePostoji)
	} else if ojGiven {
		if exists, err := repo.OrgjedPostoji(ctx, ojID); err != nil {
			add("idorgjed", common.ErrMsgReadData)
			ojOK = false
		} else if !exists {
			add("idorgjed", komMsgNePostoji)
			ojOK = false
		}
	}
	mtID, mtGiven, mtOK := komID(mt)
	switch {
	case !mtOK:
		add("mestotrid", komMsgNePostoji)
	case !mtGiven:
	default:
		if exists, err := repo.MestotrPostoji(ctx, mtID); err != nil {
			add("mestotrid", common.ErrMsgReadData)
		} else if !exists {
			add("mestotrid", komMsgNePostoji)
		} else if ojOK && ojGiven {
			if svoj, err := repo.MestotrUOrgjed(ctx, mtID, ojID); err != nil {
				add("mestotrid", common.ErrMsgReadData)
			} else if !svoj {
				add("mestotrid", komMsgMestoTroskaOJ)
			}
		}
	}
}

// komID parses the key of a combo: empty or "-" is no key (given false), a positive whole number a key (ok false for
// anything else).
func komID(value string) (id int64, given, ok bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return 0, false, true
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		return 0, true, false
	}
	return n, true, true
}

// komNullID is a key of a combo as a nullable column.
func komNullID(value string) sql.NullInt64 {
	id, given, ok := komID(value)
	return sql.NullInt64{Int64: id, Valid: given && ok}
}

// komDatum parses a date of the form (the HTML date input, or dd.mm.yyyy).
func komDatum(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"2006-01-02", "02.01.2006"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// SaveZahtev saves the header of a validated zahtev za nabavku (the legacy SnimiEDOK) in one transaction: a new one
// with the next broj of its vrsta (NoviBrojEDOK again, like the legacy), the oznaka of its magacin and the
// komercijalista of the šifra; a saved one with its datum, komercijalista, org. jedinica, mesto troška and tekst (the
// komercijalista too: the legacy kept the old one, see the report). It returns the rdokid.
func (s *ZahtevNabavkaResource) SaveZahtev(ctx context.Context, params domain.KomZahtevParams) (int64, error) {
	datum, _ := komDatum(params.Datum)
	z := domain.KomZahtevUpis{RdokID: params.RdokID, Vrd: common.StringToInt(params.Vrd), MagaciniID: common.StringToInt(params.MagaciniID),
		Dadok: datum, Orgjed: komNullID(params.Orgjed), MestoTroska: komNullID(params.MestoTroska), Foot: strings.TrimSpace(params.Tekst)}
	if sifra := common.StringToInt(strings.TrimSpace(params.KomSifra)); sifra > 0 {
		k, found, err := s.repo.KomKomercijalista(ctx, sifra)
		if err != nil {
			return 0, err
		}
		if found {
			z.KomID = sql.NullInt64{Int64: k.KomID, Valid: true}
		}
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if z.RdokID > 0 {
		if err := s.repo.KomZahtevIzmeni(ctx, tx, z); err != nil {
			return 0, err
		}
	} else {
		mag, found, err := s.repo.KomMagacin(ctx, z.MagaciniID)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("magacin %d ne postoji", z.MagaciniID)
		}
		z.Mag = mag
		if z.Dokum, err = s.repo.KomZahtevNoviBroj(ctx, tx, z.Vrd); err != nil {
			return 0, err
		}
		if z.RdokID, err = s.repo.KomZahtevDodaj(ctx, tx, z); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit zahtev za nabavku: %w", err)
	}
	return z.RdokID, nil
}

// DeleteZahtev deletes a zahtev za nabavku with its stavke (the legacy BTN_Brisanje): not a fiscalized one (knjige 4
// "S") nor a posted one (knjige 1 "D"), whose message is returned instead.
func (s *ZahtevNabavkaResource) DeleteZahtev(ctx context.Context, rdokID int64) (string, error) {
	z, found, err := s.repo.KomZahtev(ctx, rdokID)
	if err != nil {
		return "", err
	}
	switch {
	case !found:
		return "Niste selektovali dokument za brisanje!!!!", nil
	case strings.EqualFold(strings.TrimSpace(z.Knjige4), "S"):
		return komMsgFiskalizovan, nil
	case strings.EqualFold(strings.TrimSpace(z.Knjige1), "D"):
		return komMsgProknjizen, nil
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomZahtevBrisi(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje zahteva za nabavku: %w", err)
	}
	return "", nil
}

// ProveriPotvrdu checks a zahtev za nabavku before the question of its potvrda (the legacy BTN_Potvrda): it has
// stavke and it is nepotvrđen; the message why it cannot be confirmed is returned.
func (s *ZahtevNabavkaResource) ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error) {
	z, found, err := s.repo.KomZahtev(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgBezStavkiPotvrda, nil
	}
	stavke, err := s.repo.KomZahtevStavke(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if len(stavke) == 0 {
		return komMsgBezStavkiPotvrda, nil
	}
	if z.Doksts != domain.KomZahtevNepotvrdjen {
		return komMsgVecPotvrdjen, nil
	}
	return "", nil
}

// PotvrdiZahtev confirms a zahtev za nabavku (the legacy "Da" of BTN_Potvrda: doksts 40) after the checks of
// ProveriPotvrdu; it returns their message, or "Dokument je potvrđen!!!".
func (s *ZahtevNabavkaResource) PotvrdiZahtev(ctx context.Context, rdokID int64) (string, error) {
	if poruka, err := s.ProveriPotvrdu(ctx, rdokID); err != nil || poruka != "" {
		return poruka, err
	}
	ok, err := s.repo.KomZahtevPotvrdi(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !ok {
		return komMsgVecPotvrdjen, nil
	}
	return komMsgPotvrdjen, nil
}

// GetStavke returns the stavke of a zahtev za nabavku (the legacy PuniDETALJ with PR_QRY_EPRO).
func (s *ZahtevNabavkaResource) GetStavke(ctx context.Context, rdokID int64) ([]domain.KomZahtevStavkaRow, error) {
	if rdokID <= 0 {
		return nil, nil
	}
	return s.repo.KomZahtevStavke(ctx, rdokID)
}

// GetArtikal returns the artikal of a stavka with its nabavna cena (the legacy exit of EDT_SIFRA); found is false
// without it.
func (s *ZahtevNabavkaResource) GetArtikal(ctx context.Context, sifra string) (domain.KomArtikal, bool, error) {
	n, err := strconv.Atoi(strings.TrimSpace(sifra))
	if err != nil || n <= 0 {
		return domain.KomArtikal{}, false, nil
	}
	return s.repo.KomArtikal(ctx, n)
}

// ValidateStavka validates a stavka of a zahtev za nabavku (the legacy ValidacijaEPRO): its zahtev is a nepotvrđen one,
// the artikal of the šifra exists and the količina is greater than 0.
func (s *ZahtevNabavkaResource) ValidateStavka(ctx context.Context, params domain.KomZahtevStavkaParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if z, found, err := s.repo.KomZahtev(ctx, params.RdokID); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !found {
		add("sifraartikla", komMsgNePostoji)
	} else if z.Doksts != "" && z.Doksts != domain.KomZahtevNepotvrdjen {
		add("sifraartikla", komMsgNijeIzmena)
	}
	if _, found, err := s.GetArtikal(ctx, params.Sifra); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !found {
		add("sifraartikla", komMsgArtikal)
	}
	if common.StringToFloat64(params.Kolicina) <= 0 {
		add("kolicina", komMsgKolicina)
	}
	if params.RproID > 0 {
		if _, found, err := s.repo.KomZahtevStavka(ctx, params.RdokID, params.RproID); err != nil {
			add("sifraartikla", common.ErrMsgReadData)
		} else if !found {
			add("sifraartikla", komMsgNePostoji)
		}
	}
	return errs
}

// SaveStavka saves a validated stavka of a zahtev za nabavku (the legacy AddEPRO) in one transaction: a new one with
// the next redni broj (NoviRBR), a changed one (rproid) in its place; the cena is the nabavna cena of the artikal, the
// magacin and the vrsta, broj and datum are the ones of the zahtev.
func (s *ZahtevNabavkaResource) SaveStavka(ctx context.Context, params domain.KomZahtevStavkaParams) error {
	z, found, err := s.repo.KomZahtev(ctx, params.RdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("zahtev za nabavku %d ne postoji", params.RdokID)
	}
	a, found, err := s.GetArtikal(ctx, params.Sifra)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("artikal %s ne postoji", params.Sifra)
	}
	st := domain.KomZahtevStavkaUpis{RproID: params.RproID, RdokID: z.RdokID, Sifra: a.Sifra, RsifID: a.RsifID, Po: a.Po, Cena: a.NCena,
		Kolicina: common.StringToFloat64(params.Kolicina), Mag: z.Mag, MagaciniID: z.MagaciniID, Vrd: z.Vrd, Dokum: z.Dokum, Dadok: z.Dadok}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if st.RproID > 0 {
		err = s.repo.KomZahtevStavkaIzmeni(ctx, tx, st)
	} else {
		if st.Rbr, err = s.repo.KomZahtevNoviRbr(ctx, tx, st.RdokID); err != nil {
			return err
		}
		err = s.repo.KomZahtevStavkaDodaj(ctx, tx, st)
	}
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka zahteva za nabavku: %w", err)
	}
	return nil
}

// DeleteStavka deletes a stavka of a zahtev za nabavku and moves the redni brojevi of the stavke after it one back
// (the legacy BTN_BrisiDET), in one transaction.
func (s *ZahtevNabavkaResource) DeleteStavka(ctx context.Context, rdokID, rproID int64) error {
	st, found, err := s.repo.KomZahtevStavka(ctx, rdokID, rproID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("stavka %d ne postoji", rproID)
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomZahtevStavkaBrisi(ctx, tx, rdokID, rproID, st.Rbr); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit brisanje stavke zahteva za nabavku: %w", err)
	}
	return nil
}

// GetStavkaForma returns a saved stavka for the form of the stavke ("Izmeni" of its row) with the current nabavna
// cena of its artikal; found is false without it.
func (s *ZahtevNabavkaResource) GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaForma, bool, error) {
	st, found, err := s.repo.KomZahtevStavka(ctx, rdokID, rproID)
	if err != nil || !found {
		return domain.KomZahtevStavkaForma{}, false, err
	}
	return domain.KomZahtevStavkaForma{RproID: st.RproID, Rbr: st.Rbr, Sifra: st.Sifra, Naziv: st.Naziv, JM: st.JM, Kolicina: st.Kolicina,
		NCena: st.Cena}, true, nil
}

// GetPrikaz returns the zahtevi za nabavku of the tab "Pregled zahteva za nabavku" (the legacy PuniHDRZAHTEV).
func (s *ZahtevNabavkaResource) GetPrikaz(ctx context.Context, params domain.KomZahtevPrikazParams) ([]domain.KomZahtevPrikazRow, error) {
	if params.Doksts != domain.KomZahtevPotvrdjen {
		params.Doksts = domain.KomZahtevNepotvrdjen
	}
	if params.Vrd <= 0 || (!params.SviMagacini && params.MagaciniID <= 0) {
		return nil, nil
	}
	return s.repo.KomZahteviPrikaz(ctx, params)
}

// GetStampa returns the print of a zahtev za nabavku (the legacy PR_RPT_ZAHTEV_NABAVKA): the firm, the header and the
// stavke; a zahtev without stavke is not printed (the message why is returned, the legacy BTN_Print4).
func (s *ZahtevNabavkaResource) GetStampa(ctx context.Context, rdokID int64) (domain.KomZahtevStampa, string, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomZahtevStampa{}, "", fmt.Errorf("no user session found")
	}
	stavke, err := s.repo.KomZahtevStavke(ctx, rdokID)
	if err != nil {
		return domain.KomZahtevStampa{}, "", err
	}
	if len(stavke) == 0 {
		return domain.KomZahtevStampa{}, komMsgBezStavkiStampa, nil
	}
	stampa, found, err := s.repo.KomZahtevStampaZaglavlje(ctx, rdokID)
	if err != nil {
		return domain.KomZahtevStampa{}, "", err
	}
	if !found {
		return domain.KomZahtevStampa{}, komMsgBezStavkiStampa, nil
	}
	if stampa.Firma, err = s.repo.KomFirma(ctx, session.Firma); err != nil {
		return domain.KomZahtevStampa{}, "", err
	}
	stampa.Stavke = stavke
	return stampa, "", nil
}

// StavkeTableFields are the columns of the grid of the stavke of the tab "Zahtev za nabavku" (the legacy
// TABLE_PR_QRY_EPRO): the redni broj, the artikal, the zahtevana količina, the nabavna cena and the vrednost.
func (s *ZahtevNabavkaResource) StavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "6", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla / sirovine", Width: "10", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla / sirovine", Width: "40"},
		{Name: "jm", Label: "JM", Width: "6"},
		{Name: "kolicina", Label: "Zahtevana količina", Width: "12", TextAlign: "right"},
		{Name: "cena", Label: "Nabavna cena", Width: "12", TextAlign: "right"},
		{Name: "vrednost", Label: "Vrednost", Width: "14", TextAlign: "right"},
	}
}

// PrikazTableFields are the columns of the zahtevi of the tab "Pregled zahteva za nabavku" (the legacy TABLE_QRY_EDOK).
func (s *ZahtevNabavkaResource) PrikazTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "magacin", Label: "Magacin", Width: "16"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "8", TextAlign: "right"},
		{Name: "dokum", Label: "Broj zahteva", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum zahteva", Width: "10"},
		{Name: "napomena", Label: "Napomena", Width: "30"},
		{Name: "referent", Label: "Referent", Width: "16"},
		{Name: "oj", Label: "OJ", Width: "12"},
	}
}

// PrikazStavkeTableFields are the columns of the stavke of the selected zahtev of the tab "Pregled zahteva za nabavku"
// (the legacy TABLE_QRY_EPRO).
func (s *ZahtevNabavkaResource) PrikazStavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "8", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "12", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "50"},
		{Name: "jm", Label: "JM", Width: "8"},
		{Name: "kolicina", Label: "Naručena količina", Width: "14", TextAlign: "right"},
	}
}

// ZahtevStavkeTableRows are the rows of the grid of the stavke (the legacy display of a row of TABLE_PR_QRY_EPRO: the
// vrednost is the cena by the zahtevana količina).
func ZahtevStavkeTableRows(stavke []domain.KomZahtevStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), HasUpdate: true, HasDelete: true, Fields: []string{
			strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM,
			common.FormatNumberWithSystemLocale(st.Kolicina, 3),
			common.FormatNumberWithSystemLocale(st.Cena, 2),
			common.FormatNumberWithSystemLocale(st.Cena*st.Kolicina, 2),
		}})
	}
	return rows
}

// ZahtevPrikazTableRows are the rows of the zahtevi of the tab "Pregled zahteva za nabavku".
func ZahtevPrikazTableRows(zahtevi []domain.KomZahtevPrikazRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(zahtevi))
	for _, z := range zahtevi {
		datum := ""
		if z.Dadok.Valid {
			datum = z.Dadok.Time.Format(common.DateLayout)
		}
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(z.RdokID, 10), Fields: []string{
			z.Magacin, strconv.Itoa(z.Vrd), strconv.Itoa(z.Dokum), datum, z.Foot, z.Referent, z.OJ,
		}})
	}
	return rows
}

// ZahtevPrikazStavkeTableRows are the rows of the stavke of the selected zahtev of the tab "Pregled zahteva za
// nabavku".
func ZahtevPrikazStavkeTableRows(stavke []domain.KomZahtevStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), Fields: []string{
			strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM, common.FormatNumberWithSystemLocale(st.Kolicina, 3),
		}})
	}
	return rows
}

// ZahtevObrisanoPoruka is the message after the deletion of a zahtev (the legacy "Uspešno obrisani podaci...").
const ZahtevObrisanoPoruka = komMsgObrisano
