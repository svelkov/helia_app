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

// The messages of the zahtev za ponudu (the legacy pr_ZahtevPONUDA, pr_PotvrdaPONUDE and pr_StampaPONUDA); the ones it
// shares with the zahtev za nabavku are the komMsg constants of zahtevnabavka_service.go.
const (
	komMsgPonudaBezIzvora   = "Ne postoji nijedan zahtev za nabavku ili niste selektovali nijednu stavku, ne može se formirati zahtev za ponudu!!!"
	komMsgOrgjed            = "Niste izabrali organizacionu jedinicu!!!"
	komMsgBezDobavljaca     = "Niste izabrali nijednog dobavljača kome ćete poslati ponudu!!!"
	komMsgDobavljacNema     = "Dobavljač ne postoji."
	komMsgPrevisDobavljaca  = "Previše izabranih dobavljača!!!"
	komMsgNijednaIzabrana   = "Niste selektovali nijednu stavku. Dodavanje stavke nije moguće"
	komMsgPotvrdaBezDob     = "Ne postoji nijedan izabran dobavljač, kome je poslata ponuda..."
	komMsgPotvrdaDobavljac  = "Niste izabrali dobavljača!!!"
	komMsgPonudaNijeIzmena  = "Potvrđen zahtev za ponudu se ne može menjati."
	komMsgPonudaStampaBezDb = "Ne postoji nijedan dobavljač kome se šalje ponuda..."
)

// ZahtevPonudaService is the "Zahtev za ponudu" screen of the komercijalno (the legacy pr_ZahtevPONUDA).
type ZahtevPonudaService interface {
	GetPonuda(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomPonudaDokument, error)
	GetIzvori(ctx context.Context, magaciniID int) ([]domain.KomPonudaIzvorRow, error)
	GetVrste(ctx context.Context) ([]domain.ComboItem, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	ValidatePonuda(ctx context.Context, params domain.KomPonudaParams) []domain.FieldError
	SavePonuda(ctx context.Context, params domain.KomPonudaParams) (int64, error)
	PreuzmiStavke(ctx context.Context, rdokID int64, izabrane string) (string, error)
	DeletePonuda(ctx context.Context, rdokID int64) (string, error)
	ProveriPotvrdu(ctx context.Context, rdokID int64) ([]domain.KomDobavljac, string, error)
	PotvrdiPonudu(ctx context.Context, rdokID, idFkpl int64) (string, error)
	GetStavke(ctx context.Context, rdokID int64) ([]domain.KomZahtevStavkaRow, error)
	GetArtikal(ctx context.Context, sifra string) (domain.KomArtikal, bool, error)
	ValidateStavka(ctx context.Context, params domain.KomPonudaStavkaParams) []domain.FieldError
	SaveStavka(ctx context.Context, params domain.KomPonudaStavkaParams) error
	DeleteStavka(ctx context.Context, rdokID, rproID int64) error
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaForma, bool, error)
	GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error)
	GetDobavljaci(ctx context.Context, ids string) ([]domain.KomDobavljac, error)
	GetPrikaz(ctx context.Context, params domain.KomPonudaPrikazParams) ([]domain.KomPonudaPrikazRow, error)
	GetStampaDobavljaci(ctx context.Context, rdokID int64) ([]domain.KomDobavljac, string, error)
	GetStampa(ctx context.Context, rdokID, idFkpl int64) (domain.KomPonudaStampa, string, error)
	IzvoriTableFields() []domain.Fields
	StavkeTableFields() []domain.Fields
	DobavljaciTableFields() []domain.Fields
	PrikazTableFields() []domain.Fields
	PrikazStavkeTableFields() []domain.Fields
}

// ponudaRepository is what the service needs of repository.KomercijalnoRepository.
type ponudaRepository interface {
	komOrgjedRepository
	BeginTx() (db.Transaction, error)
	KomPonudaVrste(ctx context.Context) ([]domain.ComboItem, error)
	KomZahtevNoviBroj(ctx context.Context, tx db.Transaction, vrd int) (int, error)
	KomPonuda(ctx context.Context, rdokID int64) (domain.KomPonudaRow, bool, error)
	KomPonudaDodaj(ctx context.Context, tx db.Transaction, z domain.KomPonudaUpis) (int64, error)
	KomPonudaIzmeni(ctx context.Context, tx db.Transaction, z domain.KomPonudaUpis) error
	KomPonudaPotvrdi(ctx context.Context, rdokID int64, konto, sifra string) (bool, error)
	KomPonudaBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomPonudaIzvori(ctx context.Context, magaciniID int) ([]domain.KomPonudaIzvorRow, error)
	KomPonudaIzvor(ctx context.Context, tx db.Transaction, rproID int64) (domain.KomPonudaIzvor, bool, error)
	KomPonudaIzvorRezervisi(ctx context.Context, tx db.Transaction, rproID int64) error
	KomPonudaStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomPonudaStavkaUpis) error
	KomPonudaStavkaKolicina(ctx context.Context, tx db.Transaction, rdokID, rproID int64, kolicina float64) error
	KomPonudaStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error
	KomZahtevStavke(ctx context.Context, rdokID int64) ([]domain.KomZahtevStavkaRow, error)
	KomZahtevStavka(ctx context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaRow, bool, error)
	KomZahtevNoviRbr(ctx context.Context, tx db.Transaction, rdokID int64) (int, error)
	KomArtikal(ctx context.Context, sifra int) (domain.KomArtikal, bool, error)
	KomKomercijalista(ctx context.Context, sifra int) (domain.KomKomercijalista, bool, error)
	KomKomercijalistaKorisnika(ctx context.Context, userID int64) (domain.KomKomercijalista, bool, error)
	KomMagacin(ctx context.Context, magaciniID int) (int, bool, error)
	KomDobavljaci(ctx context.Context, ids []int64) ([]domain.KomDobavljac, error)
	KomPonudePrikaz(ctx context.Context, p domain.KomPonudaPrikazParams) ([]domain.KomPonudaPrikazRow, error)
	KomPonudaStampaDobavljac(ctx context.Context, idFkpl int64) (domain.KomPonudaStampaDobavljac, bool, error)
	KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error)
	BankeZiroRacuni(ctx context.Context) ([]string, error)
	NoviDokumentDefaults(ctx context.Context, firma string) (*[]domain.SearchRowDto, error)
	Valuta(ctx context.Context, sifval int) (oznval string, vaziza int, found bool, err error)
	SrednjiKurs(ctx context.Context, sifval int, datum time.Time) (float64, bool, error)
}

// ZahtevPonudaResource implements ZahtevPonudaService.
type ZahtevPonudaResource struct {
	repo   ponudaRepository
	common magaciniKorisnika
	now    func() time.Time
}

// NewZahtevPonudaService returns the service of the "Zahtev za ponudu" screen.
func NewZahtevPonudaService(repo ponudaRepository, common magaciniKorisnika) *ZahtevPonudaResource {
	return &ZahtevPonudaResource{repo: repo, common: common, now: time.Now}
}

// GetPonuda returns the content of the tab "Ponude dobavljačima" with the zahtev za ponudu of the id (a nepotvrđen one,
// the legacy BTN_TraziDOK1 and the exit of EDT_DOKUM) or, for the id 0, a new one (the legacy BTN_NoviDOK): of the vrsta
// and the magacin (the first ones without them), the next broj of the vrsta (NoviBrojEDOK), the datum of today, the
// komercijalista of the user and the valuta of the firma with the kurs 1. The slobodne stavke of the zahtevi za nabavku
// of the magacin come with it (PuniSTAVKE). A confirmed zahtev cannot be changed: a new one is given with the message.
func (s *ZahtevPonudaResource) GetPonuda(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomPonudaDokument, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomPonudaDokument{}, fmt.Errorf("no user session found")
	}
	var dok domain.KomPonudaDokument
	var err error
	if dok.Vrste, err = s.repo.KomPonudaVrste(ctx); err != nil {
		return dok, err
	}
	if dok.Magacini, err = s.common.GetMagacinKorisnikaComboValues(ctx, komMagaciniTipovi()); err != nil {
		return dok, err
	}
	if rdokID > 0 {
		z, found, err := s.repo.KomPonuda(ctx, rdokID)
		if err != nil {
			return dok, err
		}
		switch {
		case !found:
			dok.Poruka = komMsgNePostoji
		case z.Doksts != "" && z.Doksts != domain.KomZahtevNepotvrdjen:
			dok.Poruka = komMsgPonudaNijeIzmena
			vrd, magaciniID = z.Vrd, z.MagaciniID
		default:
			if dok.Header, err = s.ponudaHeader(ctx, z); err != nil {
				return dok, err
			}
			if dok.NoviRbr, err = s.repo.KomZahtevNoviRbr(ctx, nil, z.RdokID); err != nil {
				return dok, err
			}
			dok.Izvori, err = s.repo.KomPonudaIzvori(ctx, z.MagaciniID)
			return dok, err
		}
	}
	h := &dok.Header
	h.Vrd = komIzListe(dok.Vrste, vrd)
	h.MagaciniID = komIzListe(dok.Magacini, magaciniID)
	h.Datum = s.now()
	h.Kurs, h.VaziZa = 1, 1
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
	if h.SifVal, h.OznVal, err = s.valutaFirme(ctx, session.Firma); err != nil {
		return dok, err
	}
	dok.NoviRbr = 1
	if h.MagaciniID > 0 {
		if dok.Izvori, err = s.repo.KomPonudaIzvori(ctx, h.MagaciniID); err != nil {
			return dok, err
		}
	}
	return dok, nil
}

// komMagaciniTipovi is the option of the magacini of the user of a zahtev: the tipovi P, S and D (the legacy
// ROB_QRY_MAGUSER with "P;S;D").
func komMagaciniTipovi() commonsvc.ComboOption {
	return commonsvc.WithIn("magacini.tipmag", komZahtevTipoviMagacina...)
}

// valutaFirme returns the šifra and the oznaka of the valuta of the firma (the legacy nFVRSIFVAL); 0 without it.
func (s *ZahtevPonudaResource) valutaFirme(ctx context.Context, firma string) (int, string, error) {
	return komValutaFirme(ctx, s.repo, firma)
}

// komValutaRepository is what the valuta of a document needs of the repository: the defaults of the firma (its
// valuta), the valute and the kursna lista.
type komValutaRepository interface {
	NoviDokumentDefaults(ctx context.Context, firma string) (*[]domain.SearchRowDto, error)
	Valuta(ctx context.Context, sifval int) (oznval string, vaziza int, found bool, err error)
	SrednjiKurs(ctx context.Context, sifval int, datum time.Time) (float64, bool, error)
}

// komValutaFirme returns the šifra and the oznaka of the valuta of the firma (the legacy nFVRSIFVAL); 0 without it.
func komValutaFirme(ctx context.Context, repo komValutaRepository, firma string) (int, string, error) {
	defaults, err := repo.NoviDokumentDefaults(ctx, firma)
	if err != nil || defaults == nil || len(*defaults) == 0 {
		return 0, "", err
	}
	sifval := common.StringToInt((*defaults)[0].Key)
	if sifval <= 0 {
		return 0, "", nil
	}
	oznval, _, _, err := repo.Valuta(ctx, sifval)
	return sifval, oznval, err
}

// komValutaKurs returns the valuta of the šifra with its kurs on the datum (the legacy exit of EDT_SIFVAL and
// BTN_TraziVALUTU): the srednji kurs of the kursna lista of that day, else 1; the "važi za" and the oznaka of the
// valuta, else 1 and "" (found false).
func komValutaKurs(ctx context.Context, repo komValutaRepository, sifval, datum string) (domain.KomValuta, error) {
	v := domain.KomValuta{SifVal: common.StringToInt(strings.TrimSpace(sifval)), Kurs: 1, VaziZa: 1}
	if v.SifVal <= 0 {
		return v, nil
	}
	oznval, vaziza, found, err := repo.Valuta(ctx, v.SifVal)
	if err != nil {
		return v, err
	}
	if found {
		v.Found, v.OznVal, v.VaziZa = true, oznval, vaziza
	}
	if d, ok := komDatum(datum); ok {
		kurs, found, err := repo.SrednjiKurs(ctx, v.SifVal, d)
		if err != nil {
			return v, err
		}
		if found {
			v.Kurs = kurs
		}
	}
	return v, nil
}

// ponudaHeader is the header of the screen of a saved zahtev za ponudu with its dobavljači.
func (s *ZahtevPonudaResource) ponudaHeader(ctx context.Context, z domain.KomPonudaRow) (domain.KomPonudaHeaderView, error) {
	h := domain.KomPonudaHeaderView{RdokID: z.RdokID, Vrd: z.Vrd, Dokum: z.Dokum, Rok: z.Rok, MagaciniID: z.MagaciniID, KomSifra: z.KomSifra,
		KomIme: z.KomIme, Orgjed: z.Orgjed, OrgjedText: z.OrgjedText, MestoTroska: z.MestoTroska, MestoTroskaText: z.MestoTroskaText,
		Tekst: z.Foot, SifVal: z.SifVal, OznVal: z.OznVal, Kurs: z.Kurs, VaziZa: z.VaziZa, Potvrdjen: z.Doksts == domain.KomZahtevPotvrdjen}
	if z.Dadok.Valid {
		h.Datum = z.Dadok.Time
	}
	var err error
	h.Dobavljaci, err = s.GetDobavljaci(ctx, z.VezniID)
	return h, err
}

// GetVrste returns the vrste dokumenata of the zahtevi za ponudu (the filter of the main screen).
func (s *ZahtevPonudaResource) GetVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return s.repo.KomPonudaVrste(ctx)
}

// GetMagacini returns the magacini of the user of the tipovi P, S and D (the filter of the main screen).
func (s *ZahtevPonudaResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.common.GetMagacinKorisnikaComboValues(ctx, komMagaciniTipovi())
}

// GetIzvori returns the slobodne stavke of the potvrđeni zahtevi za nabavku of the magacin (the legacy PuniSTAVKE, on
// another magacin of the first tab).
func (s *ZahtevPonudaResource) GetIzvori(ctx context.Context, magaciniID int) ([]domain.KomPonudaIzvorRow, error) {
	if magaciniID <= 0 {
		return nil, nil
	}
	return s.repo.KomPonudaIzvori(ctx, magaciniID)
}

// komIDovi parses a list of ids separated by ";" or "," (the dobavljači of vezniid, the chosen stavke of the zahtevi
// za nabavku): the positive ones, each once, in their order; ok is false when a part is not a positive whole number.
func komIDovi(value string) (ids []int64, ok bool) {
	ok = true
	seen := map[int64]bool{}
	for _, part := range strings.FieldsFunc(value, func(r rune) bool { return r == ';' || r == ',' }) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			ok = false
			continue
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, ok
}

// komVezniID is the vezniid of the dobavljači: "idfkpl;" of each (the legacy opDOKID of pr_Trazi_DobPONUDE).
func komVezniID(ids []int64) string {
	var b strings.Builder
	for _, id := range ids {
		b.WriteString(strconv.FormatInt(id, 10))
		b.WriteString(";")
	}
	return b.String()
}

// GetDobavljaci returns the dobavljači of the ids (vezniid "idfkpl;idfkpl;") in their order, without the ones that do
// not exist in the period.
func (s *ZahtevPonudaResource) GetDobavljaci(ctx context.Context, ids string) ([]domain.KomDobavljac, error) {
	lista, _ := komIDovi(ids)
	if len(lista) == 0 {
		return nil, nil
	}
	rows, err := s.repo.KomDobavljaci(ctx, lista)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]domain.KomDobavljac, len(rows))
	for _, d := range rows {
		byID[d.IDFkpl] = d
	}
	out := make([]domain.KomDobavljac, 0, len(rows))
	for _, id := range lista {
		if d, found := byID[id]; found {
			out = append(out, d)
		}
	}
	return out, nil
}

// ValidatePonuda validates the header of a zahtev za ponudu (the legacy BTN_Sacuvaj and ValidacijaEDOK): a new one takes
// at least one chosen stavka of a zahtev za nabavku and its vrsta and magacin are ones of the screen; a saved one is
// nepotvrđen; the datum is a valid date, the org. jedinica is chosen (a given mesto troška is one of it), a given
// komercijalista exists, at least one dobavljač is chosen, each of them exists and their list fits vezniid.
func (s *ZahtevPonudaResource) ValidatePonuda(ctx context.Context, params domain.KomPonudaParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if params.RdokID > 0 {
		z, found, err := s.repo.KomPonuda(ctx, params.RdokID)
		switch {
		case err != nil:
			add("brdok", common.ErrMsgReadData)
		case !found:
			add("brdok", komMsgNePostoji)
		case z.Doksts != "" && z.Doksts != domain.KomZahtevNepotvrdjen:
			add("brdok", komMsgPonudaNijeIzmena)
		}
	} else {
		if izabrane, _ := komIDovi(params.Izabrane); len(izabrane) == 0 {
			add("magaciniid", komMsgPonudaBezIzvora)
		}
		if vrste, err := s.repo.KomPonudaVrste(ctx); err != nil {
			add("vrd", common.ErrMsgReadData)
		} else if !komUListi(vrste, common.StringToInt(params.Vrd)) {
			add("vrd", komMsgVrsta)
		}
		if magacini, err := s.common.GetMagacinKorisnikaComboValues(ctx, komMagaciniTipovi()); err != nil {
			add("magaciniid", common.ErrMsgReadData)
		} else if !komUListi(magacini, common.StringToInt(params.MagaciniID)) {
			add("magaciniid", komMsgMagacin)
		}
	}
	if _, ok := komDatum(params.Datum); !ok {
		add("datum", komMsgDatum)
	}
	if _, given, ok := komID(params.Orgjed); ok && !given {
		add("idorgjed", komMsgOrgjed)
	}
	komProveriOrgjedMt(ctx, s.repo, params.Orgjed, params.MestoTroska, add)
	if sifra := strings.TrimSpace(params.KomSifra); sifra != "" && sifra != "0" && sifra != "-" {
		if n, err := strconv.Atoi(sifra); err != nil {
			add("komsifra", komMsgKomercijalista)
		} else if _, found, err := s.repo.KomKomercijalista(ctx, n); err != nil {
			add("komsifra", common.ErrMsgReadData)
		} else if !found {
			add("komsifra", komMsgKomercijalista)
		}
	}
	s.proveriDobavljace(ctx, params.Dobavljaci, add)
	return errs
}

// proveriDobavljace checks the chosen dobavljači: at least one, each of them exists, their vezniid fits the column.
func (s *ZahtevPonudaResource) proveriDobavljace(ctx context.Context, value string, add func(field, message string)) {
	ids, ok := komIDovi(value)
	switch {
	case !ok:
		add("dobavljaci", komMsgDobavljacNema)
	case len(ids) == 0:
		add("dobavljaci", komMsgBezDobavljaca)
	case len(komVezniID(ids)) > domain.KomPonudaVezniIDMax:
		add("dobavljaci", komMsgPrevisDobavljaca)
	default:
		if rows, err := s.repo.KomDobavljaci(ctx, ids); err != nil {
			add("dobavljaci", common.ErrMsgReadData)
		} else if len(rows) != len(ids) {
			add("dobavljaci", komMsgDobavljacNema)
		}
	}
}

// komBroj parses a number of the form (1.234,56 or 1234.56); def for an empty value.
func komBroj(value string, def float64) float64 {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return common.StringToFloat64(value)
}

// SavePonuda saves the header of a validated zahtev za ponudu (the legacy SnimiEDOK) and takes the chosen stavke of the
// zahtevi za nabavku (AddEPRONAR) in one transaction: a new one with the next broj of its vrsta and the oznaka of its
// magacin; a saved one with its datum, rok, komercijalista, org. jedinica, mesto troška, tekst, valuta and dobavljači.
// It returns the rdokid.
func (s *ZahtevPonudaResource) SavePonuda(ctx context.Context, params domain.KomPonudaParams) (int64, error) {
	datum, _ := komDatum(params.Datum)
	ids, _ := komIDovi(params.Dobavljaci)
	z := domain.KomPonudaUpis{RdokID: params.RdokID, Vrd: common.StringToInt(params.Vrd), MagaciniID: common.StringToInt(params.MagaciniID),
		Dadok: datum, Rok: common.StringToInt(params.Rok), Orgjed: komNullID(params.Orgjed), MestoTroska: komNullID(params.MestoTroska),
		Foot: strings.TrimSpace(params.Tekst), SifVal: common.StringToInt(params.SifVal), Kurs: komBroj(params.Kurs, 1),
		VaziZa: int(komBroj(params.VaziZa, 1)), VezniID: komVezniID(ids)}
	if sifra := common.StringToInt(strings.TrimSpace(params.KomSifra)); sifra > 0 {
		k, found, err := s.repo.KomKomercijalista(ctx, sifra)
		if err != nil {
			return 0, err
		}
		if found {
			z.KomID = sql.NullInt64{Int64: k.KomID, Valid: true}
		}
	}
	if z.RdokID > 0 {
		saved, found, err := s.repo.KomPonuda(ctx, z.RdokID)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("zahtev za ponudu %d ne postoji", z.RdokID)
		}
		z.Vrd, z.Dokum, z.MagaciniID, z.Mag = saved.Vrd, saved.Dokum, saved.MagaciniID, saved.Mag
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if z.RdokID > 0 {
		if err := s.repo.KomPonudaIzmeni(ctx, tx, z); err != nil {
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
		if z.RdokID, err = s.repo.KomPonudaDodaj(ctx, tx, z); err != nil {
			return 0, err
		}
	}
	izabrane, _ := komIDovi(params.Izabrane)
	if err := s.preuzmi(ctx, tx, z, izabrane); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit zahtev za ponudu: %w", err)
	}
	return z.RdokID, nil
}

// preuzmi takes the stavke of the zahtevi za nabavku of the ids into the zahtev za ponudu (the legacy AddEPRONAR): each
// one not taken yet becomes a stavka with the next redni broj, its artikal and its tražena količina and its rproid in
// rproid1, and is marked rezervisana; one already taken (or not slobodna) is skipped, like the legacy.
func (s *ZahtevPonudaResource) preuzmi(ctx context.Context, tx db.Transaction, z domain.KomPonudaUpis, ids []int64) error {
	dadok := sql.NullTime{Time: z.Dadok, Valid: !z.Dadok.IsZero()}
	for _, id := range ids {
		izvor, found, err := s.repo.KomPonudaIzvor(ctx, tx, id)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		rbr, err := s.repo.KomZahtevNoviRbr(ctx, tx, z.RdokID)
		if err != nil {
			return err
		}
		st := domain.KomPonudaStavkaUpis{RdokID: z.RdokID, Rbr: rbr, Sifra: izvor.Sifra, RsifID: izvor.RsifID, Po: izvor.Po,
			Kolicina: izvor.Kolicina, Mag: z.Mag, MagaciniID: z.MagaciniID, Vrd: z.Vrd, Dokum: z.Dokum, Dadok: dadok, RproID1: izvor.RproID}
		if err := s.repo.KomPonudaStavkaDodaj(ctx, tx, st); err != nil {
			return err
		}
		if err := s.repo.KomPonudaIzvorRezervisi(ctx, tx, izvor.RproID); err != nil {
			return err
		}
	}
	return nil
}

// PreuzmiStavke takes the chosen stavke of the zahtevi za nabavku into a saved nepotvrđen zahtev za ponudu ("+" of the
// stavke, the legacy BTN_DodajG) in one transaction; without a chosen one its message is returned.
func (s *ZahtevPonudaResource) PreuzmiStavke(ctx context.Context, rdokID int64, izabrane string) (string, error) {
	ids, _ := komIDovi(izabrane)
	if len(ids) == 0 {
		return komMsgNijednaIzabrana, nil
	}
	z, found, err := s.repo.KomPonuda(ctx, rdokID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgNePostoji, nil
	}
	if z.Doksts != "" && z.Doksts != domain.KomZahtevNepotvrdjen {
		return komMsgPonudaNijeIzmena, nil
	}
	upis := domain.KomPonudaUpis{RdokID: z.RdokID, Vrd: z.Vrd, Dokum: z.Dokum, MagaciniID: z.MagaciniID, Mag: z.Mag}
	if z.Dadok.Valid {
		upis.Dadok = z.Dadok.Time
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.preuzmi(ctx, tx, upis, ids); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit stavke zahteva za ponudu: %w", err)
	}
	return "", nil
}

// DeletePonuda deletes a zahtev za ponudu with its stavke, their stavke of the zahtevi za nabavku slobodne again (the
// legacy BTN_Brisanje): not a fiscalized one (knjige 4 "S") nor a posted one (knjige 1 "D"), whose message is
// returned instead.
func (s *ZahtevPonudaResource) DeletePonuda(ctx context.Context, rdokID int64) (string, error) {
	z, found, err := s.repo.KomPonuda(ctx, rdokID)
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
	if err := s.repo.KomPonudaBrisi(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje zahteva za ponudu: %w", err)
	}
	return "", nil
}

// ProveriPotvrdu checks a zahtev za ponudu before its potvrda (the legacy BTN_Potvrda and pr_PotvrdaPONUDE): it has
// stavke, it is nepotvrđen and it has dobavljači; the dobavljači (the list of the dialog of the potvrda) are returned,
// or the message why it cannot be confirmed.
func (s *ZahtevPonudaResource) ProveriPotvrdu(ctx context.Context, rdokID int64) ([]domain.KomDobavljac, string, error) {
	z, found, err := s.repo.KomPonuda(ctx, rdokID)
	if err != nil {
		return nil, "", err
	}
	if !found {
		return nil, komMsgBezStavkiPotvrda, nil
	}
	stavke, err := s.repo.KomZahtevStavke(ctx, rdokID)
	if err != nil {
		return nil, "", err
	}
	if len(stavke) == 0 {
		return nil, komMsgBezStavkiPotvrda, nil
	}
	if z.Doksts != domain.KomZahtevNepotvrdjen {
		return nil, komMsgVecPotvrdjen, nil
	}
	dobavljaci, err := s.GetDobavljaci(ctx, z.VezniID)
	if err != nil {
		return nil, "", err
	}
	if len(dobavljaci) == 0 {
		return nil, komMsgPotvrdaBezDob, nil
	}
	return dobavljaci, "", nil
}

// PotvrdiPonudu confirms a zahtev za ponudu for the dobavljač of the idfkpl, one of its dobavljači (the legacy "Da" of
// BTN_Potvrda of pr_PotvrdaPONUDE: doksts 40, fkto and fana of the dobavljač) after the checks of ProveriPotvrdu; it
// returns their message, or "Dokument je potvrđen!!!".
func (s *ZahtevPonudaResource) PotvrdiPonudu(ctx context.Context, rdokID, idFkpl int64) (string, error) {
	dobavljaci, poruka, err := s.ProveriPotvrdu(ctx, rdokID)
	if err != nil || poruka != "" {
		return poruka, err
	}
	for _, d := range dobavljaci {
		if d.IDFkpl != idFkpl {
			continue
		}
		ok, err := s.repo.KomPonudaPotvrdi(ctx, rdokID, d.Konto, d.Sifra)
		if err != nil {
			return "", err
		}
		if !ok {
			return komMsgVecPotvrdjen, nil
		}
		return komMsgPotvrdjen, nil
	}
	return komMsgPotvrdaDobavljac, nil
}

// GetStavke returns the stavke of a zahtev za ponudu (the legacy PuniDETALJ1).
func (s *ZahtevPonudaResource) GetStavke(ctx context.Context, rdokID int64) ([]domain.KomZahtevStavkaRow, error) {
	if rdokID <= 0 {
		return nil, nil
	}
	return s.repo.KomZahtevStavke(ctx, rdokID)
}

// GetArtikal returns the artikal of a stavka (the legacy exit of EDT_SIFRA: the artikli of the tipovi S, A, R, D and
// K); found is false without it.
func (s *ZahtevPonudaResource) GetArtikal(ctx context.Context, sifra string) (domain.KomArtikal, bool, error) {
	n, err := strconv.Atoi(strings.TrimSpace(sifra))
	if err != nil || n <= 0 {
		return domain.KomArtikal{}, false, nil
	}
	return s.repo.KomArtikal(ctx, n)
}

// ValidateStavka validates a stavka of a zahtev za ponudu (the legacy ValidacijaEPRO): its zahtev is a nepotvrđen one,
// the količina is greater than 0; a new stavka has an artikal of the šifra, a changed one (rproid) exists (only its
// količina changes).
func (s *ZahtevPonudaResource) ValidateStavka(ctx context.Context, params domain.KomPonudaStavkaParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if z, found, err := s.repo.KomPonuda(ctx, params.RdokID); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !found {
		add("sifraartikla", komMsgNePostoji)
	} else if z.Doksts != "" && z.Doksts != domain.KomZahtevNepotvrdjen {
		add("sifraartikla", komMsgPonudaNijeIzmena)
	}
	if params.RproID > 0 {
		if _, found, err := s.repo.KomZahtevStavka(ctx, params.RdokID, params.RproID); err != nil {
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
	return errs
}

// SaveStavka saves a validated stavka of a zahtev za ponudu in one transaction: a new one entered by hand with the next
// redni broj (the legacy AddEPRO: the magacin and the vrsta, broj and datum of the zahtev, no stavka of a zahtev za
// nabavku), a changed one (rproid) with its new količina.
func (s *ZahtevPonudaResource) SaveStavka(ctx context.Context, params domain.KomPonudaStavkaParams) error {
	z, found, err := s.repo.KomPonuda(ctx, params.RdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("zahtev za ponudu %d ne postoji", params.RdokID)
	}
	kolicina := common.StringToFloat64(params.Kolicina)
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if params.RproID > 0 {
		if err := s.repo.KomPonudaStavkaKolicina(ctx, tx, z.RdokID, params.RproID, kolicina); err != nil {
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
		rbr, err := s.repo.KomZahtevNoviRbr(ctx, tx, z.RdokID)
		if err != nil {
			return err
		}
		st := domain.KomPonudaStavkaUpis{RdokID: z.RdokID, Rbr: rbr, Sifra: a.Sifra, RsifID: a.RsifID, Po: a.Po, Kolicina: kolicina, Mag: z.Mag,
			MagaciniID: z.MagaciniID, Vrd: z.Vrd, Dokum: z.Dokum, Dadok: z.Dadok}
		if err := s.repo.KomPonudaStavkaDodaj(ctx, tx, st); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka zahteva za ponudu: %w", err)
	}
	return nil
}

// DeleteStavka deletes a stavka of a zahtev za ponudu (the legacy BTN_BrisiDET) in one transaction: its stavka of the
// zahtev za nabavku is slobodna again and the redni brojevi after it move one back.
func (s *ZahtevPonudaResource) DeleteStavka(ctx context.Context, rdokID, rproID int64) error {
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
	if err := s.repo.KomPonudaStavkaBrisi(ctx, tx, rdokID, rproID, st.Rbr); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit brisanje stavke zahteva za ponudu: %w", err)
	}
	return nil
}

// GetStavkaForma returns a saved stavka for the form of the stavke ("Izmeni" of its row: its količina); found is false
// without it.
func (s *ZahtevPonudaResource) GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomZahtevStavkaForma, bool, error) {
	st, found, err := s.repo.KomZahtevStavka(ctx, rdokID, rproID)
	if err != nil || !found {
		return domain.KomZahtevStavkaForma{}, false, err
	}
	return domain.KomZahtevStavkaForma{RproID: st.RproID, Rbr: st.Rbr, Sifra: st.Sifra, Naziv: st.Naziv, JM: st.JM, Kolicina: st.Kolicina},
		true, nil
}

// GetValuta returns the valuta of the šifra with its kurs on the datum (the legacy exit of EDT_SIFVAL and
// BTN_TraziVALUTU): the srednji kurs of the kursna lista of that day, else 1; the "važi za" and the oznaka of the
// valuta, else 1 and "" (found false).
func (s *ZahtevPonudaResource) GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error) {
	return komValutaKurs(ctx, s.repo, sifval, datum)
}

// GetPrikaz returns the zahtevi za ponudu of the tab "Pregled ponuda" (the legacy PuniHDRPONUDE).
func (s *ZahtevPonudaResource) GetPrikaz(ctx context.Context, params domain.KomPonudaPrikazParams) ([]domain.KomPonudaPrikazRow, error) {
	if params.Doksts != domain.KomZahtevPotvrdjen {
		params.Doksts = domain.KomZahtevNepotvrdjen
	}
	if params.Vrd <= 0 || (!params.SviMagacini && params.MagaciniID <= 0) {
		return nil, nil
	}
	return s.repo.KomPonudePrikaz(ctx, params)
}

// GetStampaDobavljaci returns the dobavljači a zahtev za ponudu can be printed for (the legacy BTN_Print4 and
// pr_StampaPONUDA); a zahtev without stavke or without dobavljači is not printed (the message why is returned).
func (s *ZahtevPonudaResource) GetStampaDobavljaci(ctx context.Context, rdokID int64) ([]domain.KomDobavljac, string, error) {
	z, found, err := s.repo.KomPonuda(ctx, rdokID)
	if err != nil {
		return nil, "", err
	}
	if !found {
		return nil, komMsgBezStavkiStampa, nil
	}
	stavke, err := s.repo.KomZahtevStavke(ctx, rdokID)
	if err != nil {
		return nil, "", err
	}
	if len(stavke) == 0 {
		return nil, komMsgBezStavkiStampa, nil
	}
	dobavljaci, err := s.GetDobavljaci(ctx, z.VezniID)
	if err != nil {
		return nil, "", err
	}
	if len(dobavljaci) == 0 {
		return nil, komMsgPonudaStampaBezDb, nil
	}
	return dobavljaci, "", nil
}

// GetStampa returns the print of a zahtev za ponudu for one of its dobavljači (the legacy PR_RPT_ZAHTEV_PONUDA): the
// firm with its žiro računi, the dobavljač, the header and the stavke; a zahtev without stavke or a dobavljač that is
// not one of it is not printed (the message why is returned).
func (s *ZahtevPonudaResource) GetStampa(ctx context.Context, rdokID, idFkpl int64) (domain.KomPonudaStampa, string, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomPonudaStampa{}, "", fmt.Errorf("no user session found")
	}
	dobavljaci, poruka, err := s.GetStampaDobavljaci(ctx, rdokID)
	if err != nil || poruka != "" {
		return domain.KomPonudaStampa{}, poruka, err
	}
	jedan := false
	for _, d := range dobavljaci {
		jedan = jedan || d.IDFkpl == idFkpl
	}
	if !jedan {
		return domain.KomPonudaStampa{}, komMsgPotvrdaDobavljac, nil
	}
	z, found, err := s.repo.KomPonuda(ctx, rdokID)
	if err != nil {
		return domain.KomPonudaStampa{}, "", err
	}
	if !found {
		return domain.KomPonudaStampa{}, komMsgBezStavkiStampa, nil
	}
	stampa := domain.KomPonudaStampa{Dokum: z.Dokum, Datum: z.Dadok, Rok: z.Rok, Napomena: z.Foot, Nepotvrdjen: z.Doksts == domain.KomZahtevNepotvrdjen}
	if stampa.Dobavljac, _, err = s.repo.KomPonudaStampaDobavljac(ctx, idFkpl); err != nil {
		return domain.KomPonudaStampa{}, "", err
	}
	if stampa.Firma, err = s.repo.KomFirma(ctx, session.Firma); err != nil {
		return domain.KomPonudaStampa{}, "", err
	}
	if stampa.ZiroRacuni, err = s.repo.BankeZiroRacuni(ctx); err != nil {
		return domain.KomPonudaStampa{}, "", err
	}
	if stampa.Stavke, err = s.repo.KomZahtevStavke(ctx, rdokID); err != nil {
		return domain.KomPonudaStampa{}, "", err
	}
	return stampa, "", nil
}

// IzvoriTableFields are the columns of the slobodne stavke of the zahtevi za nabavku of the first tab (the legacy
// TABLE_PR_QRY_EPRO); the last one is the choice of the stavka.
func (s *ZahtevPonudaResource) IzvoriTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "5", TextAlign: "right"},
		{Name: "god", Label: "Godina", Width: "6", TextAlign: "right"},
		{Name: "dokum", Label: "Broj zahteva", Width: "8", TextAlign: "right"},
		{Name: "dadok", Label: "Datum zahteva", Width: "9"},
		{Name: "oj", Label: "OJ", Width: "14"},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "30"},
		{Name: "jm", Label: "JM", Width: "5"},
		{Name: "kolicina", Label: "Tražena količina", Width: "10", TextAlign: "right"},
		{Name: "izaberi", Label: "Izaberi stavku", Width: "6", TextAlign: "center", Type: "checkbox"},
	}
}

// StavkeTableFields are the columns of the stavke of the zahtev za ponudu (the legacy TABLE_PR_QRY_EPRO1).
func (s *ZahtevPonudaResource) StavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "8", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "12", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "50"},
		{Name: "jm", Label: "JM", Width: "8"},
		{Name: "kolicina", Label: "Količina za nabavku", Width: "14", TextAlign: "right"},
	}
}

// DobavljaciTableFields are the columns of the dobavljači of the dialogs of the potvrda and of the print (the legacy
// TBL_DOB of pr_PotvrdaPONUDE and pr_StampaPONUDA).
func (s *ZahtevPonudaResource) DobavljaciTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "konto", Label: "Konto dobavljača", Width: "10"},
		{Name: "sifra", Label: "Šifra dobavljača", Width: "10"},
		{Name: "naziv", Label: "Naziv dobavljača", Width: "40"},
		{Name: "adresa", Label: "Adresa", Width: "25"},
		{Name: "mesto", Label: "Mesto", Width: "15"},
	}
}

// PrikazTableFields are the columns of the zahtevi of the tab "Pregled ponuda" (the legacy TABLE_QRY_EDOK).
func (s *ZahtevPonudaResource) PrikazTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "magacin", Label: "Magacin", Width: "14"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "6", TextAlign: "right"},
		{Name: "dokum", Label: "Broj ponude", Width: "6", TextAlign: "right"},
		{Name: "dadok", Label: "Datum ponude", Width: "9"},
		{Name: "rok", Label: "Rok nabavke", Width: "6", TextAlign: "right"},
		{Name: "dospece", Label: "Očekivani datum nabavke", Width: "9"},
		{Name: "dobavljaci", Label: "Dobavljači kojima je poslata ponuda", Width: "26"},
		{Name: "referent", Label: "Referent nabavke", Width: "14"},
		{Name: "napomena", Label: "Napomena", Width: "14"},
	}
}

// PrikazStavkeTableFields are the columns of the stavke of the selected zahtev of the tab "Pregled ponuda" (the legacy
// TABLE_QRY_EPRO).
func (s *ZahtevPonudaResource) PrikazStavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "8", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "12", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla", Width: "50"},
		{Name: "jm", Label: "JM", Width: "8"},
		{Name: "kolicina", Label: "Naručena količina", Width: "14", TextAlign: "right"},
	}
}

// komDatumText is a datum of a grid (dd.mm.yyyy, "" without it).
func komDatumText(d sql.NullTime) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format(common.DateLayout)
}

// PonudaIzvoriTableRows are the rows of the slobodne stavke of the zahtevi za nabavku (the legacy PuniSTAVKE: the redni
// broj of the row, the godina, broj, datum and org. jedinica of the zahtev, the artikal, the tražena količina and the
// choice, not chosen).
func PonudaIzvoriTableRows(izvori []domain.KomPonudaIzvorRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(izvori))
	for i, st := range izvori {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), Fields: []string{
			strconv.Itoa(i + 1), strconv.Itoa(st.God), strconv.Itoa(st.Dokum), komDatumText(st.Dadok), st.OJ, strconv.Itoa(st.Sifra), st.Naziv,
			st.JM, common.FormatNumberWithSystemLocale(st.Kolicina, 3), "false",
		}})
	}
	return rows
}

// PonudaStavkeTableRows are the rows of the stavke of the zahtev za ponudu ("Izmeni" and "Obriši" of every row).
func PonudaStavkeTableRows(stavke []domain.KomZahtevStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), HasUpdate: true, HasDelete: true, Fields: []string{
			strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM, common.FormatNumberWithSystemLocale(st.Kolicina, 3),
		}})
	}
	return rows
}

// PonudaDobavljaciTableRows are the rows of the dobavljači (keyed by the idfkpl).
func PonudaDobavljaciTableRows(dobavljaci []domain.KomDobavljac) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(dobavljaci))
	for _, d := range dobavljaci {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(d.IDFkpl, 10), Fields: []string{d.Konto, d.Sifra, d.Naziv, d.Adresa, d.Mesto}})
	}
	return rows
}

// PonudaPrikazTableRows are the rows of the zahtevi of the tab "Pregled ponuda" (the očekivani datum nabavke is the
// datum plus the rok in days, the legacy dDOSPECE).
func PonudaPrikazTableRows(ponude []domain.KomPonudaPrikazRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(ponude))
	for _, z := range ponude {
		dospece := ""
		if z.Dadok.Valid {
			dospece = z.Dadok.Time.AddDate(0, 0, z.Rok).Format(common.DateLayout)
		}
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(z.RdokID, 10), Fields: []string{
			z.Magacin, strconv.Itoa(z.Vrd), strconv.Itoa(z.Dokum), komDatumText(z.Dadok), strconv.Itoa(z.Rok), dospece, z.Dobavljaci, z.Referent,
			z.Foot,
		}})
	}
	return rows
}
