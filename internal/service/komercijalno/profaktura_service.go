package komercijalno

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	commonsvc "helia/internal/service/common"
	robnosvc "helia/internal/service/robno"
)

// The messages of the profaktura (the legacy pr_Profakture and pr_Profakture_DETALJI); the ones it shares with the
// other documents are the komMsg constants of the other services.
const (
	komMsgProNalog          = "Za ovaj nalog formirana je otpremnica... Izmene nisu dozvoljene!!!!"
	komMsgProTipdok         = "Morate izabrati vrstu naloga za knjiženje..."
	komMsgProTipdok00       = "Ne možete unositi naloge za isporuku za izabranu vrstu naloga za knjiženje..."
	komMsgProTipdokNema     = "Niste izabrali vrstu naloga za knjiženje!!!"
	komMsgProKonto          = "Morate uneti konto partnera..."
	komMsgProSifra          = "Morate uneti šifru partnera..."
	komMsgProPartner        = "Partner sa takvom šifrom ne postoji!!!..."
	komMsgProMestoTroska    = "Niste izabrali mesto troška!!!"
	komMsgProValuta         = "Valuta sa izabranom šifrom ne postoji!!!!..."
	komMsgProKurs           = "Šifra valute je različita od domicilne, niste uneli kurs!!! Za izvozne fakture obavezan je unos kurs valute!!!"
	komMsgProVaziZa         = "Šifra valute je različita od domicilne, niste uneli paritet!!! Za izvozne fakture obavezan je unos paritet valute!!!"
	komMsgProNalogPostoji   = "Za selektovanu profakturu je formiran nalog!!! "
	komMsgProNalogBezDok    = "Niste selektovali nijedan dokument, formiranje naloga nije moguće!!!"
	komMsgProNalogBezVrste  = "Ne postoji vrsta dokumenta porudžbenice kupca, formiranje naloga nije moguće!!!"
	komMsgProStavkaNePostoj = "Stavka koju birate za izmenu ne postoji."
)

// KomProNalogFormiranPoruka is the message of a formed nalog, with its broj ({{broj}}; the legacy "Uspešno formiran
// nalog....", "Broj naloga je: ").
const KomProNalogFormiranPoruka = "Uspešno formiran nalog.... Broj naloga je: {{broj}}"

// The uslovi plaćanja and the načini isporuke of a profaktura (the legacy PuniCOMBOS: the first one is the default).
var (
	KomProUsloviPlacanja = []string{"VIRMAN", "GOTOVINA", "KOMPENZACIJA", "AVANS", "OSTALO", "-"}
	KomProNaciniIsporuke = []string{"-", "KAMION", "KOMBI", "OSTALO"}
)

// komProTipoviMagacina are the tipovi of the magacini of a profaktura (the legacy ROB_QRY_MAGUSER with "G;V").
var komProTipoviMagacina = []string{"G", "V"}

// ProfakturaService is the "Profakture" screen of the komercijalno (the legacy pr_Profakture).
type ProfakturaService interface {
	GetProfaktura(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomProDokument, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	ValidateProfaktura(ctx context.Context, params domain.KomProParams) []domain.FieldError
	SaveProfaktura(ctx context.Context, params domain.KomProParams) (int64, error)
	DeleteProfaktura(ctx context.Context, rdokID int64) (string, error)
	ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error)
	PotvrdiProfakturu(ctx context.Context, rdokID int64) (string, error)
	ProveriNalog(ctx context.Context, rdokID int64) (string, error)
	FormirajNalog(ctx context.Context, rdokID int64) (int, string, error)
	GetKupac(ctx context.Context, konto, sifra string) (domain.KomProKupac, bool, error)
	GetMestoIsporuke(ctx context.Context, konto, sifra, mi string) (domain.KomProMestoIsporuke, bool, error)
	GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error)
	GetCenovnik(ctx context.Context, id int64) (naziv, opis string, err error)
	GetStavke(ctx context.Context, rdokID int64) ([]domain.KomProStavkaRow, bool, error)
	GetArtikal(ctx context.Context, rdokID int64, sifra string, saStanjem bool) (domain.KomProArtikal, bool, error)
	ValidateStavka(ctx context.Context, params domain.KomProStavkaParams) []domain.FieldError
	SaveStavka(ctx context.Context, params domain.KomProStavkaParams) error
	DeleteStavka(ctx context.Context, rdokID, rproID int64) (string, error)
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomProStavkaForma, bool, error)
	GetPrikaz(ctx context.Context, params domain.KomProPrikazParams) ([]domain.KomProPrikazRow, error)
	GetStampa(ctx context.Context, rdokID int64) (domain.KomProStampa, string, error)
	StavkeTableFields(ino bool, oznval string, lot bool) []domain.Fields
	PrikazTableFields() []domain.Fields
	PrikazStavkeTableFields() []domain.Fields
}

// proRepository is what the service needs of repository.KomercijalnoRepository.
type proRepository interface {
	komOrgjedRepository
	komValutaRepository
	komRezervacijaRepository
	BeginTx() (db.Transaction, error)
	KomProVrste(ctx context.Context) ([]domain.ComboItem, error)
	KomProNalogVrste(ctx context.Context) ([]domain.ComboItem, error)
	KomProTipdok(ctx context.Context, vrd int) ([]domain.ComboItem, error)
	KomProfaktura(ctx context.Context, rdokID int64) (domain.KomProfakturaRow, bool, error)
	KomProMagacin(ctx context.Context, magaciniID int) (tipmag string, tipzal, tipcene int, found bool, err error)
	KomProDodaj(ctx context.Context, tx db.Transaction, p domain.KomProfakturaUpis) (int64, error)
	KomProIzmeni(ctx context.Context, tx db.Transaction, p domain.KomProfakturaUpis) error
	KomProUkupno(ctx context.Context, tx db.Transaction, rdokID int64, iznos, vrugrabat, vkase, netwght float64) error
	KomProPotvrdi(ctx context.Context, rdokID int64) (bool, error)
	KomProKupac(ctx context.Context, konto, sifra string) (domain.KomProKupac, bool, error)
	KomProMestoIsporuke(ctx context.Context, konto, sifra string, mi int) (domain.KomProMestoIsporuke, bool, error)
	KomProCenovnik(ctx context.Context, id int64) (naziv, tip string, broj int, datcen sql.NullTime, found bool, err error)
	KomProStavke(ctx context.Context, rdokID int64) ([]domain.KomProStavkaRow, error)
	KomProStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomProStavkaUpis) error
	KomProStavkaIzmeni(ctx context.Context, tx db.Transaction, st domain.KomProStavkaIzmena) error
	KomProLotovi(ctx context.Context, magaciniID int, rsifID int64, saStanjem bool) ([]domain.KomProLot, error)
	KomProCenaArtikla(ctx context.Context, sifra int) (float64, error)
	KomProCenaCenovnika(ctx context.Context, cenovnikID int64, sifra int) (vpcena, rabat float64, found bool, err error)
	KomProPlanskaCena(ctx context.Context, sifra int) (float64, bool, error)
	KomProNalogDodaj(ctx context.Context, tx db.Transaction, proID int64, vrd, dokum int) (int64, error)
	KomProNalogStavke(ctx context.Context, tx db.Transaction, proID, nalogID int64, vrd, dokum int) error
	KomProNalogFormiran(ctx context.Context, tx db.Transaction, proID int64) (bool, error)
	KomProPrikaz(ctx context.Context, p domain.KomProPrikazParams) ([]domain.KomProPrikazRow, error)
	KomPrijArtikal(ctx context.Context, sifra int) (domain.KomPrijArtikal, bool, error)
	KomPrijKonta(ctx context.Context, magaciniID int) ([]domain.ComboItem, error)
	KomPorNoviBroj(ctx context.Context, tx db.Transaction, vrd, mag int) (int, error)
	KomPorIDFkpl(ctx context.Context, konto, sifra string) (int64, error)
	KomPonudaStampaDobavljac(ctx context.Context, idFkpl int64) (domain.KomPonudaStampaDobavljac, bool, error)
	KomZahtevNoviRbr(ctx context.Context, tx db.Transaction, rdokID int64) (int, error)
	KomZahtevStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error
	KomZahtevBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomKomercijalista(ctx context.Context, sifra int) (domain.KomKomercijalista, bool, error)
	KomKomercijalistaKorisnika(ctx context.Context, userID int64) (domain.KomKomercijalista, bool, error)
	KomMagacin(ctx context.Context, magaciniID int) (int, bool, error)
	KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error)
	BankeZiroRacuni(ctx context.Context) ([]string, error)
	RporStopa(ctx context.Context, tx db.Transaction, po int, datum time.Time) (float64, error)
}

// ProfakturaResource implements ProfakturaService. The nalog za isporuku (NalogResource) is a variant of it: the same
// header, kupac, artikal and stavke of another vrsta (komProVarijanta).
type ProfakturaResource struct {
	repo      proRepository
	common    magaciniKorisnika
	now       func() time.Time
	varijanta komProVarijanta
}

// komProVarijanta is what differs between the profaktura and its variants: the tipovi of the magacini, the načini
// isporuke (the first one is the default), the vrste dokumenata, the reading of a document of the vrste, the message of
// a status that cannot change ("" when it can), the labels of the broj and the datum of the document and the opis of a
// new one (nil: the repository's "Profaktura br: <broj>").
type komProVarijanta struct {
	tipoviMagacina []string
	naciniIsporuke []string
	vrste          func(ctx context.Context) ([]domain.ComboItem, error)
	dokument       func(ctx context.Context, rdokID int64) (domain.KomProfakturaRow, bool, error)
	zakljucan      func(doksts string) string
	brojLabel      string
	datumLabel     string
	opis           func(dokum int) string
}

// NewProfakturaService returns the service of the "Profakture" screen.
func NewProfakturaService(repo proRepository, common magaciniKorisnika) *ProfakturaResource {
	return &ProfakturaResource{repo: repo, common: common, now: time.Now, varijanta: komProVarijanta{
		tipoviMagacina: komProTipoviMagacina, naciniIsporuke: KomProNaciniIsporuke, vrste: repo.KomProVrste, dokument: repo.KomProfaktura,
		zakljucan: func(doksts string) string {
			if doksts == domain.KomProfakturaNalog {
				return komMsgProNalog
			}
			return ""
		},
		brojLabel: "Broj profakture", datumLabel: "Datum profakture"}}
}

// komWinDevPocetak is the first day of the WinDev integer dates (drsta.roktr, epro.roktr: the days since it).
var komWinDevPocetak = time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)

// komWinDevDatum is the WinDev integer date of a datum (0 without it; the legacy ConvertDateToInteger).
func komWinDevDatum(datum string) int64 {
	d, ok := komDatum(datum)
	if !ok {
		return 0
	}
	return int64(d.Sub(komWinDevPocetak).Hours() / 24)
}

// komWinDevDatumText is a WinDev integer date as dd.mm.yyyy ("" for 0; the legacy ConvertIntegerToDate).
func komWinDevDatumText(n int64) string {
	if n <= 0 {
		return ""
	}
	return komWinDevPocetak.AddDate(0, 0, int(n)).Format(common.DateLayout)
}

// komRound2 rounds to 2 decimals (the legacy Round(x, 2)).
func komRound2(x float64) float64 {
	return math.Round(x*100) / 100
}

// GetMagacini returns the magacini of the user of the tipovi G and V (the filter of the main screen).
func (s *ProfakturaResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.common.GetMagacinKorisnikaComboValues(ctx, commonsvc.WithIn("magacini.tipmag", s.varijanta.tipoviMagacina...))
}

// GetValuta returns the valuta of the šifra with its kurs on the datum (the legacy exit of EDT_SIFVAL).
func (s *ProfakturaResource) GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error) {
	return komValutaKurs(ctx, s.repo, sifval, datum)
}

// GetKupac returns the kupac of the konto and šifra with the text of its sistem PDV, its ugovoreni rabat and its rok
// plaćanja (the legacy exit of EDT_FANA); found is false without it.
func (s *ProfakturaResource) GetKupac(ctx context.Context, konto, sifra string) (domain.KomProKupac, bool, error) {
	if strings.TrimSpace(konto) == "" || strings.TrimSpace(sifra) == "" {
		return domain.KomProKupac{}, false, nil
	}
	k, found, err := s.repo.KomProKupac(ctx, konto, sifra)
	if err != nil || !found {
		return domain.KomProKupac{}, found, err
	}
	k.SistemPdv = komSistemPdv(k.TipPdv)
	return k, true, nil
}

// GetMestoIsporuke returns the mesto isporuke of the kupac (the legacy BTN_Sifra3); found is false without it.
func (s *ProfakturaResource) GetMestoIsporuke(ctx context.Context, konto, sifra, mi string) (domain.KomProMestoIsporuke, bool, error) {
	n, err := strconv.Atoi(strings.TrimSpace(mi))
	if err != nil || n <= 0 {
		return domain.KomProMestoIsporuke{}, false, nil
	}
	return s.repo.KomProMestoIsporuke(ctx, konto, sifra, n)
}

// komProCenovnikTip is the text of the tip of a cenovnik (the legacy BTN_TraziCENE: "Robu", "Gotove proizvode", ...).
func komProCenovnikTip(tip string) string {
	switch tip {
	case "R":
		return "Robu"
	case "P":
		return "Gotove proizvode"
	case "U":
		return "Usluge"
	case "S":
		return "Sirovine"
	case "L":
		return "Poluproizvode"
	case "A":
		return "Ambalažu"
	case "T":
		return "Troškove"
	case "D":
		return "Rezervne delove"
	}
	return ""
}

// GetCenovnik returns the naziv of the klasa cena of a cenovnik and its opis ("Cenovnik za: ROBU", "Broj: 3 Od datuma:
// 01/03/2026"; the legacy BTN_TraziCENE); empty without it.
func (s *ProfakturaResource) GetCenovnik(ctx context.Context, id int64) (string, string, error) {
	if id <= 0 {
		return "", "", nil
	}
	naziv, tip, broj, datcen, found, err := s.repo.KomProCenovnik(ctx, id)
	if err != nil || !found {
		return "", "", err
	}
	datum := ""
	if datcen.Valid {
		datum = datcen.Time.Format("02/01/2006")
	}
	return naziv, fmt.Sprintf("Cenovnik za: %s\nBroj: %d Od datuma: %s", strings.ToUpper(komProCenovnikTip(tip)), broj, datum), nil
}

// zajedno fills the lists of the entry dialog every profaktura has: the magacini, the vrste, the žiro računi of the
// firm and its valuta, and whether the firm is a PDV obveznik.
func (s *ProfakturaResource) zajedno(ctx context.Context, dok *domain.KomProDokument) error {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return fmt.Errorf("no user session found")
	}
	var err error
	if dok.Magacini, err = s.GetMagacini(ctx); err != nil {
		return err
	}
	if dok.Vrste, err = s.varijanta.vrste(ctx); err != nil {
		return err
	}
	racuni, err := s.repo.BankeZiroRacuni(ctx)
	if err != nil {
		return err
	}
	for _, r := range racuni {
		dok.ZiroRacuni = append(dok.ZiroRacuni, domain.ComboItem{Key: r, Value: r})
	}
	for _, u := range KomProUsloviPlacanja {
		dok.UsloviPlacanja = append(dok.UsloviPlacanja, domain.ComboItem{Key: u, Value: u})
	}
	for _, n := range s.varijanta.naciniIsporuke {
		dok.NaciniIsporuke = append(dok.NaciniIsporuke, domain.ComboItem{Key: n, Value: n})
	}
	if dok.SifValFirme, _, err = komValutaFirme(ctx, s.repo, session.Firma); err != nil {
		return err
	}
	firma, err := s.repo.KomFirma(ctx, session.Firma)
	if err != nil {
		return err
	}
	dok.Header.PdvObveznik = firma.Obv
	return nil
}

// GetProfaktura returns the content of the entry dialog with the profaktura of the id (the legacy PrikaziPODHDR) or, for
// the id 0, a new one (the legacy BTN_NoviDOK): of the magacin and the vrsta (the first ones without them), the first
// vrsta naloga za knjiženje of the vrsta, the next broj of the vrsta in the magacin (NoviBrojEDOK), the datum of today,
// the konto of the kupci, the komercijalista of the user, the valuta of the firma with the kurs 1, the first uslovi
// plaćanja, način isporuke and žiro račun, the oznaka of the poreska kategorija ("S" for a PDV obveznik, else "SS"). A
// fiscalized or posted profaktura, or one with a nalog, cannot be changed: a new one is given with the message why.
func (s *ProfakturaResource) GetProfaktura(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomProDokument, error) {
	var dok domain.KomProDokument
	if err := s.zajedno(ctx, &dok); err != nil {
		return dok, err
	}
	dok.BrojLabel, dok.DatumLabel = s.varijanta.brojLabel, s.varijanta.datumLabel
	if rdokID > 0 {
		p, found, err := s.varijanta.dokument(ctx, rdokID)
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
		case s.varijanta.zakljucan(p.Doksts) != "":
			dok.Poruka, vrd, magaciniID = s.varijanta.zakljucan(p.Doksts), p.Vrd, p.MagaciniID
		default:
			return s.sacuvana(ctx, dok, p)
		}
	}
	session := domain.GetSessionFromStdContext(ctx)
	h := &dok.Header
	h.Datum = s.now()
	h.Dadok = sql.NullTime{Time: h.Datum, Valid: true}
	h.MagaciniID = komIzListe(dok.Magacini, magaciniID)
	h.Vrd = komIzListe(dok.Vrste, vrd)
	h.Kurs, h.VaziZa, h.SifVal = 1, 1, dok.SifValFirme
	h.Pla, h.Otp = KomProUsloviPlacanja[0], s.varijanta.naciniIsporuke[0]
	if len(dok.ZiroRacuni) > 0 {
		h.Posuslnab = dok.ZiroRacuni[0].Key
	}
	h.Category = "SS"
	if h.PdvObveznik {
		h.Category = "S"
	}
	var err error
	if dok.Tipdok, err = s.repo.KomProTipdok(ctx, h.Vrd); err != nil {
		return dok, err
	}
	if len(dok.Tipdok) > 0 {
		h.Tipdok = dok.Tipdok[0].Key
	}
	if defaults, err := s.repo.NoviDokumentDefaults(ctx, session.Firma); err != nil {
		return dok, err
	} else if defaults != nil && len(*defaults) > 0 {
		h.Fkto = strings.TrimSpace((*defaults)[0].C3)
	}
	if h.MagaciniID > 0 {
		if h.Tipmag, h.Tipzal, h.Tipcene, _, err = s.repo.KomProMagacin(ctx, h.MagaciniID); err != nil {
			return dok, err
		}
		if dok.Konta, err = s.repo.KomPrijKonta(ctx, h.MagaciniID); err != nil {
			return dok, err
		}
		if h.Vrd > 0 {
			mag, found, err := s.repo.KomMagacin(ctx, h.MagaciniID)
			if err != nil {
				return dok, err
			}
			if found {
				h.Mag = mag
				if h.Dokum, err = s.repo.KomPorNoviBroj(ctx, nil, h.Vrd, mag); err != nil {
					return dok, err
				}
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

// sacuvana is the content of the entry dialog with a saved profaktura that can change: its header with the kupac, the
// mesto isporuke and the cenovnik, the vrste naloga za knjiženje of its vrsta, the konta of its magacin, the next
// redni broj and whether it has stavke.
func (s *ProfakturaResource) sacuvana(ctx context.Context, dok domain.KomProDokument, p domain.KomProfakturaRow) (domain.KomProDokument, error) {
	h := &dok.Header
	obveznik := h.PdvObveznik
	*h = domain.KomProHeaderView{KomProfakturaRow: p, PdvObveznik: obveznik, Ino: komIno(p.SifVal, dok.SifValFirme)}
	if p.Dadok.Valid {
		h.Datum = p.Dadok.Time
	}
	var err error
	if kupac, found, err := s.GetKupac(ctx, p.Fkto, p.Fana); err != nil {
		return dok, err
	} else if found {
		h.Kupac = kupac.KomDobavljacPodaci
	}
	if p.Mi > 0 {
		if h.MestoIsporuke, _, err = s.repo.KomProMestoIsporuke(ctx, p.Fkto, p.Fana, p.Mi); err != nil {
			return dok, err
		}
	}
	if h.CenovnikText, h.CenovnikOpis, err = s.GetCenovnik(ctx, p.CenovnikID); err != nil {
		return dok, err
	}
	if dok.Tipdok, err = s.repo.KomProTipdok(ctx, p.Vrd); err != nil {
		return dok, err
	}
	if dok.Konta, err = s.repo.KomPrijKonta(ctx, p.MagaciniID); err != nil {
		return dok, err
	}
	if dok.NoviRbr, err = s.repo.KomZahtevNoviRbr(ctx, nil, p.RdokID); err != nil {
		return dok, err
	}
	h.ImaStavke = dok.NoviRbr > 1
	return dok, nil
}

// ValidateProfaktura validates the header of a profaktura (the legacy ValidacijaEDOK): a saved one has no nalog; the
// magacin of a new one is one of the user, its vrsta one of the profakture and its vrsta naloga za knjiženje one of the
// vrsta (not "00"); the datum is a valid date, the kupac (konto and šifra) exists, the org. jedinica and the mesto
// troška are given and exist, a given komercijalista exists, the valuta exists and a foreign valuta has a kurs and a
// "važi za".
func (s *ProfakturaResource) ValidateProfaktura(ctx context.Context, params domain.KomProParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if params.RdokID > 0 {
		p, found, err := s.varijanta.dokument(ctx, params.RdokID)
		switch {
		case err != nil:
			add("datum", common.ErrMsgReadData)
		case !found:
			add("datum", komMsgNePostoji)
		case s.varijanta.zakljucan(p.Doksts) != "":
			add("datum", s.varijanta.zakljucan(p.Doksts))
		}
	} else {
		magaciniID, vrd := common.StringToInt(params.MagaciniID), common.StringToInt(params.Vrd)
		if magacini, err := s.GetMagacini(ctx); err != nil {
			add("magaciniid", common.ErrMsgReadData)
		} else if !komUListi(magacini, magaciniID) {
			add("magaciniid", komMsgMagacin)
		}
		if vrste, err := s.varijanta.vrste(ctx); err != nil {
			add("vrd", common.ErrMsgReadData)
		} else if !komUListi(vrste, vrd) {
			add("vrd", komMsgVrsta)
		}
		tipdok := strings.TrimSpace(params.Tipdok)
		switch tipdok {
		case "":
			add("tipdok", komMsgProTipdok)
		case "00":
			add("tipdok", komMsgProTipdok00)
		default:
			if lista, err := s.repo.KomProTipdok(ctx, vrd); err != nil {
				add("tipdok", common.ErrMsgReadData)
			} else if !komProUListi(lista, tipdok) {
				add("tipdok", komMsgProTipdokNema)
			}
		}
	}
	if _, ok := komDatum(params.Datum); !ok {
		add("datum", komMsgDatum)
	}
	konto, sifra := strings.TrimSpace(params.Fkto), strings.TrimSpace(params.Fana)
	switch {
	case konto == "":
		add("fkto", komMsgProKonto)
	case sifra == "":
		add("fana", komMsgProSifra)
	default:
		if _, found, err := s.repo.KomProKupac(ctx, konto, sifra); err != nil {
			add("fana", common.ErrMsgReadData)
		} else if !found {
			add("fana", komMsgProPartner)
		}
	}
	// The org. jedinica and the mesto troška are required (the legacy "Niste izabrali organizacionu jedinicu!!!" and
	// "Niste izabrali mesto troška!!!").
	if id, given, ok := komID(params.Orgjed); !ok || !given {
		add("idorgjed", komMsgOrgjed)
	} else if exists, err := s.repo.OrgjedPostoji(ctx, id); err != nil {
		add("idorgjed", common.ErrMsgReadData)
	} else if !exists {
		add("idorgjed", komMsgOrgjed)
	}
	if id, given, ok := komID(params.MestoTroska); !ok || !given {
		add("mestotrid", komMsgProMestoTroska)
	} else if exists, err := s.repo.MestotrPostoji(ctx, id); err != nil {
		add("mestotrid", common.ErrMsgReadData)
	} else if !exists {
		add("mestotrid", komMsgProMestoTroska)
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
	sifval := common.StringToInt(strings.TrimSpace(params.SifVal))
	if _, _, found, err := s.repo.Valuta(ctx, sifval); err != nil {
		add("valutaid", common.ErrMsgReadData)
	} else if !found {
		add("valutaid", komMsgProValuta)
	} else if session := domain.GetSessionFromStdContext(ctx); session != nil {
		if sifvalFirme, _, err := komValutaFirme(ctx, s.repo, session.Firma); err == nil && sifval != sifvalFirme {
			if komBroj(params.Kurs, 0) <= 0 {
				add("kurs", komMsgProKurs)
			}
			if komBroj(params.VaziZa, 0) <= 0 {
				add("vaziza", komMsgProVaziZa)
			}
		}
	}
	return errs
}

// komProUListi tells whether the key (a text) is in the list.
func komProUListi(lista []domain.ComboItem, kljuc string) bool {
	for _, item := range lista {
		if kljuc != "" && item.Key == kljuc {
			return true
		}
	}
	return false
}

// SaveProfaktura saves the header of a validated profaktura (the legacy SnimiEDOK) in one transaction: a new one with
// the next broj of its vrsta in its magacin, the oznaka of its magacin, its vrsta naloga za knjiženje; a saved one with
// its datum, kupac, uslovi, valuta, poreska kategorija and the other fields of the header (its vrsta, vrsta naloga za
// knjiženje, broj and magacin stay). The cenovnik is kept only in a magacin with the tip cene 2. Then the totals of
// the profaktura are calculated again (the legacy SnimiEDOK). It returns the rdokid.
func (s *ProfakturaResource) SaveProfaktura(ctx context.Context, params domain.KomProParams) (int64, error) {
	return s.sacuvaj(ctx, params, nil)
}

// sacuvaj saves the header of a validated document of the variant (see SaveProfaktura); posle (when given) runs in the
// same transaction after the header is saved, with what was saved (its rdokid, vrsta, broj, datum and magacin).
func (s *ProfakturaResource) sacuvaj(ctx context.Context, params domain.KomProParams, posle func(tx db.Transaction, p domain.KomProfakturaUpis) error) (int64, error) {
	datum, _ := komDatum(params.Datum)
	p := domain.KomProfakturaUpis{RdokID: params.RdokID, Vrd: common.StringToInt(params.Vrd), Tipdok: strings.TrimSpace(params.Tipdok),
		MagaciniID: common.StringToInt(params.MagaciniID), Dadok: datum, Rok: common.StringToInt(params.Rok), Fkto: strings.TrimSpace(params.Fkto),
		Fana: strings.TrimSpace(params.Fana), Orgjed: komNullID(params.Orgjed), MestoTroska: komNullID(params.MestoTroska),
		Foot: strings.TrimSpace(params.Foot), SifVal: common.StringToInt(params.SifVal), Kurs: komBroj(params.Kurs, 1),
		VaziZa: int(komBroj(params.VaziZa, 1)), Otp: strings.TrimSpace(params.Otp), Pla: strings.TrimSpace(params.Pla),
		Posuslnab: strings.TrimSpace(params.Posuslnab), Mi: common.StringToInt(params.Mi), Ugrabat: komBroj(params.Ugrabat, 0),
		Pkase: komBroj(params.Pkase, 0), Paritet: strings.TrimSpace(params.Paritet), Izjizv: common.StringToInt(params.Izjizv),
		Keykat: strings.TrimSpace(params.Keykat), Category: strings.TrimSpace(params.Category), Pornapomena: strings.TrimSpace(params.Pornapomena),
		Kamt: strings.TrimSpace(params.Kamt), Brugov: strings.TrimSpace(params.Brugov)}
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
		saved, found, err := s.varijanta.dokument(ctx, p.RdokID)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("profaktura %d ne postoji", p.RdokID)
		}
		p.MagaciniID, p.Mag, p.Vrd, p.Dokum = saved.MagaciniID, saved.Mag, saved.Vrd, saved.Dokum
	}
	_, _, tipcene, _, err := s.repo.KomProMagacin(ctx, p.MagaciniID)
	if err != nil {
		return 0, err
	}
	if tipcene == 2 {
		p.CenovnikID = int64(common.StringToInt(params.CenovnikID))
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if p.RdokID > 0 {
		if err := s.repo.KomProIzmeni(ctx, tx, p); err != nil {
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
		if s.varijanta.opis != nil {
			p.Opis = s.varijanta.opis(p.Dokum)
		}
		if p.RdokID, err = s.repo.KomProDodaj(ctx, tx, p); err != nil {
			return 0, err
		}
	}
	if posle != nil {
		if err := posle(tx, p); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit profaktura: %w", err)
	}
	return p.RdokID, s.preracunaj(ctx, p.RdokID)
}

// komProUkupno are the totals of a profaktura (the legacy Preracun and SnimiEDOK): the iznos at the planska cena (Σ
// količina × cena), the vrednost of the ugovoreni rabat ((Σ količina × fcena - the rabat) × ugovoreni rabat %) and of
// the kasa ((… - the ugovoreni rabat) × kasa %) and the neto težina (Σ težina × količina).
func komProUkupno(stavke []domain.KomProStavkaRow, ugrabat, pkase float64) (iznos, vrugrabat, vkase, netwght float64) {
	var vrednost, rabat float64
	for _, st := range stavke {
		vrednost += st.Kolicina * st.Fcena
		iznos += st.Kolicina * st.Cena
		rabat += st.Kolicina * st.Fcena * st.Rabat / 100
		netwght += st.Tezjm * st.Kolicina
	}
	vrednost, rabat, iznos = komRound2(vrednost), komRound2(rabat), komRound2(iznos)
	vrugrabat = komRound2((vrednost - rabat) * ugrabat / 100)
	vkase = komRound2((vrednost - rabat - vrugrabat) * pkase / 100)
	return iznos, vrugrabat, vkase, netwght
}

// preracunaj calculates the totals of a profaktura again and saves them (the legacy Preracun).
func (s *ProfakturaResource) preracunaj(ctx context.Context, rdokID int64) error {
	p, found, err := s.varijanta.dokument(ctx, rdokID)
	if err != nil || !found {
		return err
	}
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	if err != nil {
		return err
	}
	iznos, vrugrabat, vkase, netwght := komProUkupno(stavke, p.Ugrabat, p.Pkase)
	return s.repo.KomProUkupno(ctx, nil, rdokID, iznos, vrugrabat, vkase, netwght)
}

// DeleteProfaktura deletes a profaktura with its stavke (the legacy BTN_Brisanje): not a fiscalized (knjige 4 "S"),
// posted (knjige 1 "D") one or one with a nalog, whose message is returned instead.
func (s *ProfakturaResource) DeleteProfaktura(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.varijanta.dokument(ctx, rdokID)
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
	case p.Doksts == domain.KomProfakturaNalog:
		return komMsgProNalogPostoji, nil
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
		return "", fmt.Errorf("commit brisanje profakture: %w", err)
	}
	return "", nil
}

// ProveriPotvrdu checks a profaktura before the question of its potvrda (the legacy BTN_Potvrda): it has stavke and it
// is new (status below 20); the message why it cannot be confirmed is returned.
func (s *ProfakturaResource) ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.varijanta.dokument(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgBezStavkiPotvrda, nil
	}
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if len(stavke) == 0 {
		return komMsgBezStavkiPotvrda, nil
	}
	if common.StringToInt(p.Doksts) >= 20 {
		return komMsgVecPotvrdjen, nil
	}
	return "", nil
}

// PotvrdiProfakturu confirms a profaktura (the legacy "Da" of BTN_Potvrda: doksts 20) after the checks of
// ProveriPotvrdu; it returns their message, or "Dokument je potvrđen!!!".
func (s *ProfakturaResource) PotvrdiProfakturu(ctx context.Context, rdokID int64) (string, error) {
	if poruka, err := s.ProveriPotvrdu(ctx, rdokID); err != nil || poruka != "" {
		return poruka, err
	}
	ok, err := s.repo.KomProPotvrdi(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !ok {
		return komMsgVecPotvrdjen, nil
	}
	return komMsgPotvrdjen, nil
}

// ProveriNalog checks a profaktura before the question of the forming of its nalog (the legacy BTN_FAKTURA "Form.
// nalog"): it exists and has no nalog yet; the message why is returned.
func (s *ProfakturaResource) ProveriNalog(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.varijanta.dokument(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgProNalogBezDok, nil
	}
	if p.Doksts == domain.KomProfakturaNalog {
		return komMsgProNalogPostoji, nil
	}
	return "", nil
}

// FormirajNalog forms the nalog of a profaktura (the legacy FormirajNALOG) after the checks of ProveriNalog, in one
// transaction: a porudžbenica kupca of the first vrsta of the group NRK with its next broj in the magacin of the
// profaktura, its header and stavke copied, the zalihe of the stavke reserved (komRezervisi: drsta.rez of the lot and
// rsta.rez; the legacy reserves only drsta.rez of an existing lot) and the profaktura marked (doksts 40). It returns
// the broj of the nalog, or the message why it was not formed.
func (s *ProfakturaResource) FormirajNalog(ctx context.Context, rdokID int64) (int, string, error) {
	if poruka, err := s.ProveriNalog(ctx, rdokID); err != nil || poruka != "" {
		return 0, poruka, err
	}
	p, _, err := s.varijanta.dokument(ctx, rdokID)
	if err != nil {
		return 0, "", err
	}
	vrste, err := s.repo.KomProNalogVrste(ctx)
	if err != nil {
		return 0, "", err
	}
	if len(vrste) == 0 {
		return 0, komMsgProNalogBezVrste, nil
	}
	vrd := common.StringToInt(vrste[0].Key)
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	if err != nil {
		return 0, "", err
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = tx.Rollback() }()
	ok, err := s.repo.KomProNalogFormiran(ctx, tx, rdokID)
	if err != nil {
		return 0, "", err
	}
	if !ok {
		return 0, komMsgProNalogPostoji, nil
	}
	broj, err := s.repo.KomPorNoviBroj(ctx, tx, vrd, p.Mag)
	if err != nil {
		return 0, "", err
	}
	nalogID, err := s.repo.KomProNalogDodaj(ctx, tx, rdokID, vrd, broj)
	if err != nil {
		return 0, "", err
	}
	if err := s.repo.KomProNalogStavke(ctx, tx, rdokID, nalogID, vrd, broj); err != nil {
		return 0, "", err
	}
	for _, st := range stavke {
		if err := komRezervisi(ctx, s.repo, tx, komRezervacija(p, st, st.Kolicina)); err != nil {
			return 0, "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, "", fmt.Errorf("commit nalog profakture: %w", err)
	}
	return broj, "", nil
}

// GetStavke returns the stavke of a profaktura and whether its prices are in a foreign valuta.
func (s *ProfakturaResource) GetStavke(ctx context.Context, rdokID int64) ([]domain.KomProStavkaRow, bool, error) {
	if rdokID <= 0 {
		return nil, false, nil
	}
	p, found, err := s.varijanta.dokument(ctx, rdokID)
	if err != nil || !found {
		return nil, false, err
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return nil, false, err
	}
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	return stavke, ino, err
}

// ino tells whether the valuta is a foreign one for the firma of the session.
func (s *ProfakturaResource) ino(ctx context.Context, sifval int) (bool, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return false, fmt.Errorf("no user session found")
	}
	sifvalFirme, _, err := komValutaFirme(ctx, s.repo, session.Firma)
	return komIno(sifval, sifvalFirme), err
}

// GetArtikal returns the artikal of a stavka (the legacy BTN_Sifra and the exit of EDT_SIFRA): its lots in the magacin
// of the profaktura (only the ones with stanje with saStanjem), its prodajna cena with the rabat by the tip cene of the
// magacin (2: the cenovnik of the profaktura, vpcena and rabat; else rcene.cena); found is false without it.
func (s *ProfakturaResource) GetArtikal(ctx context.Context, rdokID int64, sifra string, saStanjem bool) (domain.KomProArtikal, bool, error) {
	n, err := strconv.Atoi(strings.TrimSpace(sifra))
	if err != nil || n <= 0 {
		return domain.KomProArtikal{}, false, nil
	}
	p, found, err := s.varijanta.dokument(ctx, rdokID)
	if err != nil || !found {
		return domain.KomProArtikal{}, false, err
	}
	a, found, err := s.repo.KomPrijArtikal(ctx, n)
	if err != nil || !found {
		return domain.KomProArtikal{}, found, err
	}
	art := domain.KomProArtikal{RsifID: a.RsifID, Sifra: a.Sifra, Naziv: a.Naziv, JM: a.JM, Po: a.Po}
	if art.Lotovi, err = s.repo.KomProLotovi(ctx, p.MagaciniID, a.RsifID, saStanjem); err != nil {
		return art, true, err
	}
	for i := range art.Lotovi {
		art.Lotovi[i].Rok = komWinDevDatumText(art.Lotovi[i].Roktr)
	}
	if p.Tipcene == 2 {
		if art.Cena, art.Rabat, _, err = s.repo.KomProCenaCenovnika(ctx, p.CenovnikID, a.Sifra); err != nil {
			return art, true, err
		}
	} else if art.Cena, err = s.repo.KomProCenaArtikla(ctx, a.Sifra); err != nil {
		return art, true, err
	}
	return art, true, nil
}

// izmenljiva returns the profaktura of the id when its stavke can change (it has no nalog), else the message why.
func (s *ProfakturaResource) izmenljiva(ctx context.Context, rdokID int64) (domain.KomProfakturaRow, string, error) {
	p, found, err := s.varijanta.dokument(ctx, rdokID)
	switch {
	case err != nil:
		return p, "", err
	case !found:
		return p, komMsgNePostoji, nil
	case s.varijanta.zakljucan(p.Doksts) != "":
		return p, s.varijanta.zakljucan(p.Doksts), nil
	}
	return p, "", nil
}

// proStavka returns the stavka of the id of a profaktura; found is false without it.
func (s *ProfakturaResource) proStavka(ctx context.Context, rdokID, rproID int64) (domain.KomProStavkaRow, bool, error) {
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	if err != nil {
		return domain.KomProStavkaRow{}, false, err
	}
	for _, st := range stavke {
		if st.RproID == rproID {
			return st, true, nil
		}
	}
	return domain.KomProStavkaRow{}, false, nil
}

// ValidateStavka validates a stavka of a profaktura (the legacy ValidacijaEPRO): its profaktura has no nalog, a new
// stavka has an artikal of the šifra and a changed one (rproid) exists; the količina and the cena are greater than 0.
func (s *ProfakturaResource) ValidateStavka(ctx context.Context, params domain.KomProStavkaParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if _, poruka, err := s.izmenljiva(ctx, params.RdokID); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if poruka != "" {
		add("sifraartikla", poruka)
	}
	if params.RproID > 0 {
		if _, found, err := s.proStavka(ctx, params.RdokID, params.RproID); err != nil {
			add("sifraartikla", common.ErrMsgReadData)
		} else if !found {
			add("sifraartikla", komMsgProStavkaNePostoj)
		}
	} else if n, err := strconv.Atoi(strings.TrimSpace(params.Sifra)); err != nil || n <= 0 {
		add("sifraartikla", komMsgArtikal)
	} else if _, found, err := s.repo.KomPrijArtikal(ctx, n); err != nil {
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

// komProCene returns the prodajna cena of a stavka in RSD (fcena) and in the valuta (cenaval) of an entered cena (the
// legacy AddEPRO): in a foreign valuta the entered cena is the cenaval and the fcena is converted by the kurs and the
// "važi za", rounded to 2 decimals; else both are the entered cena.
func komProCene(unesena float64, ino bool, kurs float64, vaziza int) (fcena, cenaVal float64) {
	if !ino {
		return unesena, unesena
	}
	if vaziza <= 0 {
		vaziza = 1
	}
	return komRound2(unesena * kurs / float64(vaziza)), unesena
}

// SaveStavka saves a validated stavka of a profaktura in one transaction: a new one with the next redni broj (the
// legacy AddEPRO: the artikal with its lot, the konto, the količina, the planska cena — of the default normativ in a
// magacin of the tip G, else the prodajna cena —, the prodajna cena in RSD and in the valuta, the rabat), a changed one
// (rproid) with its količina, prodajna cena and rabat. Then the totals of the profaktura are calculated again.
func (s *ProfakturaResource) SaveStavka(ctx context.Context, params domain.KomProStavkaParams) error {
	p, found, err := s.varijanta.dokument(ctx, params.RdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("profaktura %d ne postoji", params.RdokID)
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return err
	}
	kolicina, rabat := common.StringToFloat64(params.Kolicina), common.StringToFloat64(params.Rabat)
	fcena, cenaVal := komProCene(common.StringToFloat64(params.Cena), ino, p.Kurs, p.VaziZa)
	st := domain.KomProStavkaUpis{}
	if params.RproID <= 0 {
		n, _ := strconv.Atoi(strings.TrimSpace(params.Sifra))
		a, found, err := s.repo.KomPrijArtikal(ctx, n)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("artikal %s ne postoji", params.Sifra)
		}
		planska := fcena
		if p.Tipmag == "G" {
			if planska, _, err = s.repo.KomProPlanskaCena(ctx, a.Sifra); err != nil {
				return err
			}
		}
		st = domain.KomProStavkaUpis{RdokID: p.RdokID, Sifra: a.Sifra, RsifID: a.RsifID, Po: a.Po, Konto: strings.TrimSpace(params.Konto),
			Mag: p.Mag, MagaciniID: p.MagaciniID, Kolicina: kolicina, Cena: planska, Fcena: fcena, CenaVal: cenaVal, Rabat: rabat,
			Otk: strings.TrimSpace(params.Otk), Serija: strings.TrimSpace(params.Serija), Roktr: komWinDevDatum(params.Rok), Vrd: p.Vrd,
			Dokum: p.Dokum, Dadok: p.Dadok, Isporuceno: kolicina}
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if params.RproID > 0 {
		if err := s.repo.KomProStavkaIzmeni(ctx, tx, domain.KomProStavkaIzmena{RdokID: p.RdokID, RproID: params.RproID, Kolicina: kolicina,
			Fcena: fcena, CenaVal: cenaVal, Rabat: rabat}); err != nil {
			return err
		}
	} else {
		if st.Rbr, err = s.repo.KomZahtevNoviRbr(ctx, tx, p.RdokID); err != nil {
			return err
		}
		if err := s.repo.KomProStavkaDodaj(ctx, tx, st); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka profakture: %w", err)
	}
	return s.preracunaj(ctx, p.RdokID)
}

// DeleteStavka deletes a stavka of a profaktura without a nalog (the legacy BTN_BrisiDET: the redni brojevi after it
// move one back) and calculates the totals again; the message why is returned when it cannot.
func (s *ProfakturaResource) DeleteStavka(ctx context.Context, rdokID, rproID int64) (string, error) {
	if _, poruka, err := s.izmenljiva(ctx, rdokID); err != nil || poruka != "" {
		return poruka, err
	}
	st, found, err := s.proStavka(ctx, rdokID, rproID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgProStavkaNePostoj, nil
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomZahtevStavkaBrisi(ctx, tx, rdokID, st.RproID, st.Rbr); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje stavke profakture: %w", err)
	}
	return "", s.preracunaj(ctx, rdokID)
}

// GetStavkaForma returns a saved stavka for the form of the stavke ("Izmeni" of its row) with its cena in the valuta of
// the profaktura; found is false without it.
func (s *ProfakturaResource) GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomProStavkaForma, bool, error) {
	stavke, ino, err := s.GetStavke(ctx, rdokID)
	if err != nil {
		return domain.KomProStavkaForma{}, false, err
	}
	for _, st := range stavke {
		if st.RproID == rproID {
			return domain.KomProStavkaForma{RproID: st.RproID, Rbr: st.Rbr, Konto: strings.TrimSpace(st.Konto), Sifra: st.Sifra, Naziv: st.Naziv,
				JM: st.JM, Otk: st.Otk, Serija: st.Serija, Rok: komWinDevDatumText(st.Roktr), Kolicina: st.Kolicina, Cena: komProCenaPrikaz(st, ino),
				Rabat: st.Rabat}, true, nil
		}
	}
	return domain.KomProStavkaForma{}, false, nil
}

// komProCenaPrikaz is the cena of a stavka in the valuta of its profaktura (cenaval in a foreign valuta, else fcena; the
// legacy PuniDETALJ1).
func komProCenaPrikaz(st domain.KomProStavkaRow, ino bool) float64 {
	if ino {
		return st.CenaVal
	}
	return st.Fcena
}

// GetPrikaz returns the profakture of the list of the main screen (the legacy PR_QRY_EDOK).
func (s *ProfakturaResource) GetPrikaz(ctx context.Context, params domain.KomProPrikazParams) ([]domain.KomProPrikazRow, error) {
	if params.Doksts != domain.KomProfakturaNova {
		params.Doksts = domain.KomProfakturaPotvrdjena
	}
	if !params.SviMagacini && params.MagaciniID <= 0 {
		return nil, nil
	}
	return s.repo.KomProPrikaz(ctx, params)
}

// GetStampa returns the print of a profaktura (the legacy PR_RPT_PROFAKTURA, PR_RPT_PROFAKTURAIZV in a foreign valuta):
// the firm with its tekući računi, the kupac, the mesto isporuke, the header and the stavke with the totals (see
// komProObracun); a profaktura without stavke is not printed (the message why is returned).
func (s *ProfakturaResource) GetStampa(ctx context.Context, rdokID int64) (domain.KomProStampa, string, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomProStampa{}, "", fmt.Errorf("no user session found")
	}
	p, found, err := s.varijanta.dokument(ctx, rdokID)
	if err != nil {
		return domain.KomProStampa{}, "", err
	}
	if !found {
		return domain.KomProStampa{}, komMsgBezStavkiStampa, nil
	}
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	if err != nil {
		return domain.KomProStampa{}, "", err
	}
	if len(stavke) == 0 {
		return domain.KomProStampa{}, komMsgBezStavkiStampa, nil
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return domain.KomProStampa{}, "", err
	}
	st := domain.KomProStampa{Ino: ino, OznVal: p.OznVal, Dokum: p.Dokum, Datum: p.Dadok, Pla: p.Pla, Otp: p.Otp, Rok: p.Rok, Komercijalista: p.KomIme,
		Vozac: p.Vozac, Brvozila: p.Brvozila, Paritet: p.Paritet, Izjizv: p.Izjizv == 1, Napomena: p.Foot, Pornapomena: strings.TrimSpace(p.Pornapomena),
		NetoTezina: p.Netwght}
	if p.Dadok.Valid {
		st.Valuta = sql.NullTime{Time: p.Dadok.Time.AddDate(0, 0, p.Rok), Valid: true}
	}
	if st.Firma, err = s.repo.KomFirma(ctx, session.Firma); err != nil {
		return domain.KomProStampa{}, "", err
	}
	st.PdvObveznik = st.Firma.Obv
	st.MestoIzdavanja = strings.TrimSpace(p.MagacinMesto)
	if st.MestoIzdavanja == "" {
		st.MestoIzdavanja = strings.TrimSpace(st.Firma.Mesto)
	}
	if strings.TrimSpace(p.MagacinAdresa) != "" {
		st.Firma.Adresa = strings.TrimSpace(st.Firma.Adresa) + " , " + strings.TrimSpace(p.MagacinAdresa)
	}
	if !ino {
		if st.ZiroRacuni, err = s.repo.BankeZiroRacuni(ctx); err != nil {
			return domain.KomProStampa{}, "", err
		}
	}
	if idFkpl, err := s.repo.KomPorIDFkpl(ctx, p.Fkto, p.Fana); err != nil {
		return domain.KomProStampa{}, "", err
	} else if idFkpl > 0 {
		if st.Kupac, _, err = s.repo.KomPonudaStampaDobavljac(ctx, idFkpl); err != nil {
			return domain.KomProStampa{}, "", err
		}
	}
	if p.Mi > 0 {
		mi, found, err := s.repo.KomProMestoIsporuke(ctx, p.Fkto, p.Fana, p.Mi)
		if err != nil {
			return domain.KomProStampa{}, "", err
		}
		if found {
			st.MestoIsporuke = mi.Naziv
			st.MestoIsporuke2 = strings.TrimSpace(mi.Adresa + ", " + strings.TrimSpace(mi.Pobro+" "+mi.Mesto))
			if mi.Gln != "" {
				st.MestoIsporuke2 += " GLN:" + mi.Gln
			}
		}
	}
	stope := map[int]float64{}
	for _, stavka := range stavke {
		if _, ok := stope[stavka.Po]; ok || ino || !p.Dadok.Valid {
			continue
		}
		if stope[stavka.Po], err = s.repo.RporStopa(ctx, nil, stavka.Po, p.Dadok.Time); err != nil {
			return domain.KomProStampa{}, "", err
		}
	}
	komProObracun(&st, p, stavke, stope)
	// Slovima only for the dinar firm (the legacy nFVRSIFVAL = 941), on the domestic print.
	if !ino {
		sifvalFirme, _, err := komValutaFirme(ctx, s.repo, session.Firma)
		if err != nil {
			return domain.KomProStampa{}, "", err
		}
		if sifvalFirme == 941 {
			st.Slovima = robnosvc.BrojSlovima(st.ZaNaplatu)
		}
	}
	return st, "", nil
}

// komProStavkaNaziv is the naziv of a printed stavka (the legacy ITEM_NAZIVART): with the serija and the rok trajanja in
// a magacin with the tip zaliha 2, else with the komercijalni opis, the proizvođač and the proizvođačka šifra.
func komProStavkaNaziv(st domain.KomProStavkaRow, tipzal int, ino bool) string {
	if tipzal == 2 {
		if ino {
			return st.Naziv + "\nSeries: " + st.Serija + " Expiration date: " + komWinDevDatumText(st.Roktr)
		}
		return st.Naziv + "\nSerija: " + st.Serija + " Rok trajanja: " + komWinDevDatumText(st.Roktr)
	}
	delovi := []string{st.Naziv}
	for _, d := range []string{st.Komercopis, st.Pro, st.Proizsifra} {
		if d = strings.TrimSpace(d); d != "" {
			delovi = append(delovi, d)
		}
	}
	return strings.Join(delovi, " ")
}

// komProObracun fills the stavke and the totals of the print of a profaktura (the legacy BODY and BREAK_FOTRDOKID):
//   - a stavka: the iznos (količina × prodajna cena, in the valuta for a foreign one), the rabat on it, the ugovoreni
//     rabat and the kasa of the document on what is left, the poreska osnovica, the stopa PDV of its tarifa on the
//     datum (0 with a napomena o poreskom oslobođenju or in a foreign valuta) and the PDV of one unit;
//   - BRUTO, RABAT, NETO with the ugovoreni rabat and the kasa (when the document has them), the PDV per stopa (only
//     for a PDV obveznik and not in a foreign valuta) and the ukupna naknada (za naplatu).
func komProObracun(st *domain.KomProStampa, p domain.KomProfakturaRow, stavke []domain.KomProStavkaRow, stope map[int]float64) {
	type poStopi struct{ osnovica, pdv float64 }
	perStopa := map[float64]*poStopi{}
	var pdvUkupno float64
	for i, s := range stavke {
		cena := s.Fcena
		if st.Ino {
			cena = s.CenaVal
		}
		stopa := stope[s.Po]
		if st.Pornapomena != "" || st.Ino {
			stopa = 0
		}
		iznos := komRound2(s.Kolicina * cena)
		rabat := komRound2(iznos * s.Rabat / 100)
		ugr := komRound2((iznos - rabat) * p.Ugrabat / 100)
		kasa := komRound2((iznos - rabat - ugr) * p.Pkase / 100)
		osnovica := komRound2(iznos - rabat - ugr - kasa)
		pdv := komRound2(osnovica * stopa / 100)
		var pdvJM float64
		if s.Kolicina != 0 {
			pdvJM = komRound2(osnovica * stopa / 100 / s.Kolicina)
		}
		st.Stavke = append(st.Stavke, domain.KomProStampaStavka{Rbr: i + 1, Sifra: s.Sifra, Naziv: komProStavkaNaziv(s, p.Tipzal, st.Ino), JM: s.JM,
			Kolicina: s.Kolicina, Decimale: s.Decimale, Cena: cena, ProcRab: s.Rabat, Rabat: rabat, Stopa: stopa, PdvJM: pdvJM,
			Iznos: s.Kolicina * cena})
		if perStopa[stopa] == nil {
			perStopa[stopa] = &poStopi{}
		}
		perStopa[stopa].osnovica += osnovica
		perStopa[stopa].pdv += pdv
		st.Svega += iznos
		st.Rabat += rabat
	}
	st.Neto = st.Svega - st.Rabat
	if p.Vrugrabat != 0 {
		st.NetoPrikaz, st.UgrabatPct = true, p.Ugrabat
		st.Ugrabat = komRound2(st.Neto * p.Ugrabat / 100)
	}
	if p.Vkase != 0 {
		st.NetoPrikaz, st.KasaPct = true, p.Pkase
		st.Kasa = komRound2((st.Neto - st.Ugrabat) * p.Pkase / 100)
	}
	if st.PdvObveznik && !st.Ino {
		stopeRed := make([]float64, 0, len(perStopa))
		for stopa := range perStopa {
			stopeRed = append(stopeRed, stopa)
		}
		sort.Float64s(stopeRed)
		for _, stopa := range stopeRed {
			st.PdvPoStopama = append(st.PdvPoStopama, domain.KomProStampaPdv{Stopa: stopa, Osnovica: perStopa[stopa].osnovica, Pdv: perStopa[stopa].pdv})
			pdvUkupno += perStopa[stopa].pdv
		}
	}
	st.ZaNaplatu = st.Svega + pdvUkupno - st.Rabat - st.Ugrabat - st.Kasa
}

// StavkeTableFields are the columns of the grid of the stavke of the profaktura (the legacy TABLE_PR_QRY_EPRO): the
// cena in the valuta of the document; OTK, serija and rok upotrebe only in a magacin with the lots (tip zaliha 2).
func (s *ProfakturaResource) StavkeTableFields(ino bool, oznval string, lot bool) []domain.Fields {
	cena := domain.Fields{Name: "cena", Label: "Jedinična cena", Width: "10", TextAlign: "right"}
	if ino && strings.TrimSpace(oznval) != "" {
		cena.Label, cena.Params = komCenaUValuti, map[string]string{"valuta": strings.ToUpper(strings.TrimSpace(oznval))}
	}
	fields := []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "5", TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "6"},
		{Name: "sifra", Label: "Šifra artikla", Width: "7", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla/proizvoda", Width: "30"},
		{Name: "jm", Label: "JM", Width: "4"},
	}
	if lot {
		fields = append(fields, domain.Fields{Name: "otk", Label: "OTK", Width: "6"}, domain.Fields{Name: "serija", Label: "Serija", Width: "8"},
			domain.Fields{Name: "rok", Label: "Rok upotrebe", Width: "8"})
	}
	return append(fields,
		domain.Fields{Name: "kolicina", Label: "Količina", Width: "9", TextAlign: "right"},
		cena,
		domain.Fields{Name: "rabat", Label: "% Rabata", Width: "6", TextAlign: "right"},
	)
}

// PrikazTableFields are the columns of the list of the main screen (the legacy TABLE_QRY_EDOK).
func (s *ProfakturaResource) PrikazTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "magacin", Label: "Magacin", Width: "12"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "6", TextAlign: "right"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "7", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8"},
		{Name: "fkto", Label: "Konto partnera", Width: "7"},
		{Name: "fana", Label: "Šifra", Width: "6"},
		{Name: "kupac", Label: "Naziv partnera", Width: "22"},
		{Name: "opis", Label: "Opis", Width: "14"},
		{Name: "rok", Label: "Rok plaćanja", Width: "5", TextAlign: "right"},
		{Name: "otp", Label: "Način isporuke", Width: "8"},
		{Name: "nalog", Label: "Nalog", Width: "6", TextAlign: "right"},
	}
}

// PrikazStavkeTableFields are the columns of the stavke of the selected profaktura of the main screen (the legacy
// TABLE_QRY_EPRO: the vrednost is the količina by the cena).
func (s *ProfakturaResource) PrikazStavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "6", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "9", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla/proizvoda", Width: "34"},
		{Name: "jm", Label: "JM", Width: "5"},
		{Name: "kolicina", Label: "Količina", Width: "10", TextAlign: "right"},
		{Name: "cena", Label: "Jedinična cena", Width: "10", TextAlign: "right"},
		{Name: "rabat", Label: "% Rabata", Width: "7", TextAlign: "right"},
		{Name: "vrednost", Label: "Vrednost", Width: "12", TextAlign: "right", IncludeInTotals: true},
	}
}

// komProKolicinaText is a količina of a grid with the decimals of its JM (at most 3).
func komProKolicinaText(kolicina float64, decimale int) string {
	if decimale < 0 || decimale > 3 {
		decimale = 3
	}
	return common.FormatNumberWithSystemLocale(kolicina, decimale)
}

// ProStavkeTableRows are the rows of the stavke of a profaktura, the last ones first like the legacy grid (TableSort
// "-COL_RBR"), with "Izmeni" and "Obriši" of every row (izmena); the cena in the valuta of the document, OTK, serija
// and rok upotrebe only with the lots.
func ProStavkeTableRows(stavke []domain.KomProStavkaRow, ino, lot, izmena bool) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		fields := []string{strconv.Itoa(st.Rbr), strings.TrimSpace(st.Konto), strconv.Itoa(st.Sifra), st.Naziv, st.JM}
		if lot {
			fields = append(fields, st.Otk, st.Serija, komWinDevDatumText(st.Roktr))
		}
		fields = append(fields, komProKolicinaText(st.Kolicina, st.Decimale), common.FormatNumberWithSystemLocale(komProCenaPrikaz(st, ino), 2),
			common.FormatNumberWithSystemLocale(st.Rabat, 2))
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), HasUpdate: izmena, HasDelete: izmena, Fields: fields})
	}
	return rows
}

// ProPrikazTableRows are the rows of the list of the main screen: "Izmeni" and "Briši" of every profaktura without a
// nalog (izmena), the broj of the nalog of the others.
func ProPrikazTableRows(profakture []domain.KomProPrikazRow, izmena bool) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(profakture))
	for _, p := range profakture {
		nalog := ""
		if p.NalogDokum > 0 {
			nalog = strconv.Itoa(p.NalogDokum)
		}
		menja := izmena && p.Doksts != domain.KomProfakturaNalog
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(p.RdokID, 10), HasUpdate: menja, HasDelete: menja, Fields: []string{
			p.Magacin, strconv.Itoa(p.Vrd), strconv.Itoa(p.Dokum), komDatumText(p.Dadok), p.Fkto, p.Fana, p.Kupac, p.Opis, strconv.Itoa(p.Rok), p.Otp,
			nalog,
		}})
	}
	return rows
}

// ProPrikazStavkeTableRows are the rows of the stavke of the selected profaktura of the main screen (the vrednost the
// količina by the cena in the valuta of the document) with their total.
func ProPrikazStavkeTableRows(stavke []domain.KomProStavkaRow, ino bool) ([]domain.TableRow, float64) {
	rows := make([]domain.TableRow, 0, len(stavke))
	ukupno := 0.0
	for _, st := range stavke {
		cena := komProCenaPrikaz(st, ino)
		vrednost := st.Kolicina * cena
		ukupno += vrednost
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), Fields: []string{
			strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM, komProKolicinaText(st.Kolicina, st.Decimale),
			common.FormatNumberWithSystemLocale(cena, 2), common.FormatNumberWithSystemLocale(st.Rabat, 2), common.FormatNumberWithSystemLocale(vrednost, 2),
		}})
	}
	return rows, ukupno
}
