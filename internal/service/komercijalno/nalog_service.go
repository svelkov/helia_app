package komercijalno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

// The messages of the nalog za isporuku (the legacy pr_PorudzbenicaProdaja and pr_PorudzbenicaProdaja_DETALJI); the
// ones it shares with the other documents are the komMsg constants of the other services.
const (
	komMsgNalPotvrdjen    = "Nalog je potvrđen, izmene nisu dozvoljene!!!"
	komMsgNalKurs         = "Niste uneli kurs. Kad je valuta domicilna, kurs mora biti 1!!!"
	komMsgNalPorkat       = "Morate izabrati poresku kategoriju!"
	komMsgNalProfaktura   = "Izabrana profaktura ne postoji ili nije potvrđena!!!"
	komMsgNalProMagacin   = "Izabrana profaktura nije iz magacina naloga!!!"
	komMsgNalBanka        = "Banka za uplatu sa izabranom šifrom ne postoji!!!"
	komMsgNalCenaNula     = "Postoje stavke čija je cena = 0. Potvrda dokumenta nije dozvoljena!!!"
	komMsgNalPlanskaCena  = "Za izabrani proizvod ne postoji definisana planska cena, koja je potrebna kasnije za kontiranje. Unesite plansku cenu u normativ gotovih proizvoda, zatim nastavite unos..."
	komMsgNalBrisanjeNema = "Niste selektovali dokument za brisanje!!!!"
)

// KomNalogNaciniIsporuke are the načini isporuke of a nalog (the legacy PuniCOMBOS: the first one is the default).
var KomNalogNaciniIsporuke = []string{"KAMION", "KOMBI", "OSTALO", "-"}

// komNalogTipoviMagacina are the tipovi of the magacini of a nalog (the legacy ROB_QRY_MAGUSER with "G;V;S").
var komNalogTipoviMagacina = []string{"G", "V", "S"}

// KomNalogStatusOpis is the text of a status of a nalog (the legacy EDT_STATUS).
func KomNalogStatusOpis(doksts string) string {
	switch doksts {
	case domain.KomNalogNov:
		return "Unet dokument, nepotvrđen nalog"
	case domain.KomNalogPotvrdjen:
		return "Potvrđen nalog, spreman za otpremu"
	case domain.KomNalogOtpremljen:
		return "Formirana otpremnica, faktura"
	}
	return ""
}

// komNalogZakljucan is the message of a nalog that cannot change: a potvrđen one or one with the otpremnica (the legacy
// ValidacijaEDOK allows changes up to the otpremnica; only the new ones change here).
func komNalogZakljucan(doksts string) string {
	switch doksts {
	case domain.KomNalogPotvrdjen:
		return komMsgNalPotvrdjen
	case domain.KomNalogOtpremljen:
		return komMsgProNalog
	}
	return ""
}

// NalogService is the "Nalozi za isporuku" screen of the komercijalno (the legacy pr_PorudzbenicaProdaja).
type NalogService interface {
	GetNalog(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomNalogDokument, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	ValidateNalog(ctx context.Context, params domain.KomNalogParams) []domain.FieldError
	SaveNalog(ctx context.Context, params domain.KomNalogParams) (int64, error)
	DeleteNalog(ctx context.Context, rdokID int64) (string, error)
	ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error)
	PotvrdiNalog(ctx context.Context, rdokID int64) (string, error)
	GetProfakturaNaloga(ctx context.Context, rdokID int64, magaciniID int) (domain.KomNalogProfaktura, bool, error)
	GetSaldo(ctx context.Context, konto, sifra, datum string) (domain.KomNalogSaldo, error)
	GetKupac(ctx context.Context, konto, sifra string) (domain.KomProKupac, bool, error)
	GetMestoIsporuke(ctx context.Context, konto, sifra, mi string) (domain.KomProMestoIsporuke, bool, error)
	GetValuta(ctx context.Context, sifval, datum string) (domain.KomValuta, error)
	GetCenovnik(ctx context.Context, id int64) (naziv, opis string, err error)
	GetStavke(ctx context.Context, rdokID int64) ([]domain.KomProStavkaRow, bool, error)
	GetArtikal(ctx context.Context, rdokID int64, sifra string, saStanjem bool) (domain.KomProArtikal, bool, error)
	ValidateNalogStavka(ctx context.Context, params domain.KomNalogStavkaParams) []domain.FieldError
	SaveNalogStavka(ctx context.Context, params domain.KomNalogStavkaParams) error
	DeleteNalogStavka(ctx context.Context, rdokID, rproID int64) (string, error)
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomProStavkaForma, bool, error)
	GetNalogPrikaz(ctx context.Context, params domain.KomNalogPrikazParams) ([]domain.KomNalogPrikazRow, error)
	GetNalogStampa(ctx context.Context, rdokID int64) (domain.KomNalogStampa, string, error)
	StavkeTableFields(ino bool, oznval string, lot bool) []domain.Fields
	NalogPrikazTableFields() []domain.Fields
	NalogPrikazStavkeTableFields(delimicni bool) []domain.Fields
}

// komRezervacijaRepository is what the reservation of the zalihe needs of repository.RobnoRepository.
type komRezervacijaRepository interface {
	DrstaIDZaIzmenu(ctx context.Context, tx db.Transaction, konto string, sifra, mag int, otk, serija string, roktr int64) (int64, error)
	RstaIDZaIzmenu(ctx context.Context, tx db.Transaction, magaciniID int, konto string, sifra int) (int64, error)
	StanjeDodaj(ctx context.Context, tx db.Transaction, tabela repository.StanjeTabela, vrednosti map[string]any) error
	StanjeIzmeni(ctx context.Context, tx db.Transaction, tabela repository.StanjeTabela, rstaID int64, izmene []repository.Izmena) error
}

// nalRepository is what the service needs of repository.KomercijalnoRepository.
type nalRepository interface {
	proRepository
	KomNalog(ctx context.Context, rdokID int64) (domain.KomProfakturaRow, bool, error)
	KomNalogDodatno(ctx context.Context, rdokID int64) (domain.KomNalogDodatno, error)
	KomNalogDodatnoSacuvaj(ctx context.Context, tx db.Transaction, rdokID, vrdokID int64, sifbank int) error
	KomNalogBanka(ctx context.Context, bnkizvID int64) (int, bool, error)
	KomNalogKopirajStavke(ctx context.Context, tx db.Transaction, proID, nalogID int64, vrd, dokum int, dadok sql.NullTime, rbrOd int) error
	KomNalogProfakturaVezi(ctx context.Context, tx db.Transaction, proID int64) (bool, error)
	KomNalogProfakturaVrati(ctx context.Context, tx db.Transaction, proID int64) (bool, error)
	KomNalogLotoviStarost(ctx context.Context, magaciniID int, konto string, sifra int) ([]domain.KomNalogLot, error)
	KomNalogFpro(ctx context.Context, konto, sifra string) ([]domain.KomNalogFpro, error)
	KomNalogRobneStavke(ctx context.Context, konto, sifra string) ([]domain.KomNalogRobnaStavka, error)
	KomNalogPrikaz(ctx context.Context, p domain.KomNalogPrikazParams) ([]domain.KomNalogPrikazRow, error)
}

// NalogResource implements NalogService: the profaktura of the vrste of the group NRK (its header, kupac, artikal and
// stavke) with the reservation of the zalihe of the stavke, the profaktura it is made of, the bank of the payment of an
// izvoz and the finansijsko stanje of the kupac.
type NalogResource struct {
	*ProfakturaResource
	repo nalRepository
}

// errKomNalogProfaktura is returned when the profaktura of a nalog stopped being a potvrđena one before the save.
var errKomNalogProfaktura = errors.New("profaktura naloga nije potvrđena")

// NewNalogService returns the service of the "Nalozi za isporuku" screen.
func NewNalogService(repo nalRepository, common magaciniKorisnika) *NalogResource {
	pro := &ProfakturaResource{repo: repo, common: common, now: time.Now, varijanta: komProVarijanta{
		tipoviMagacina: komNalogTipoviMagacina, naciniIsporuke: KomNalogNaciniIsporuke, vrste: repo.KomProNalogVrste, dokument: repo.KomNalog,
		zakljucan: komNalogZakljucan, brojLabel: "Broj porudžbenice", datumLabel: "Datum porudžbenice",
		opis: func(dokum int) string { return fmt.Sprintf("Nalog za izdavanje robe br: %d", dokum) }}}
	return &NalogResource{ProfakturaResource: pro, repo: repo}
}

// komRezervacija is the reservation of the količina of a stavka of a document in its lot.
func komRezervacija(p domain.KomProfakturaRow, st domain.KomProStavkaRow, kolicina float64) domain.KomNalogRezervacija {
	return domain.KomNalogRezervacija{MagaciniID: p.MagaciniID, Mag: p.Mag, RsifID: st.RsifID, Sifra: st.Sifra, Konto: st.Konto, Otk: st.Otk, Serija: st.Serija,
		Roktr: st.Roktr, Cena: st.Cena, Kolicina: kolicina}
}

// komRezervisi reserves (a positive količina) or releases (a negative one) the zalihe of a stavka (the legacy AddEPRO,
// the exit of a changed row and BrisiDET): drsta.rez of its lot — a lot without a row of drsta gets one with only the
// reservation (the legacy AddEPRO) — and rsta.rez of the artikal in its konto when it has a row of rsta. A release of
// rsta.rez does not go below 0: the nalozi of the legacy did not reserve it.
func komRezervisi(ctx context.Context, repo komRezervacijaRepository, tx db.Transaction, r domain.KomNalogRezervacija) error {
	if r.Kolicina == 0 {
		return nil
	}
	if err := komRezervisiLot(ctx, repo, tx, r); err != nil {
		return err
	}
	konto := strings.TrimSpace(r.Konto)
	id, err := repo.RstaIDZaIzmenu(ctx, tx, r.MagaciniID, konto, r.Sifra)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return repo.StanjeIzmeni(ctx, tx, repository.TabelaRsta, id, []repository.Izmena{{Kolona: "rez", Dodaj: true, NeNegativno: r.Kolicina < 0,
		Vrednost: r.Kolicina}})
}

// komRezervisiLot reserves or releases the zalihe of a stavka only in its lot (drsta.rez; a lot without a row of drsta
// gets one with only the reservation): the move of a reservation to another lot of the same artikal (rsta stays).
func komRezervisiLot(ctx context.Context, repo komRezervacijaRepository, tx db.Transaction, r domain.KomNalogRezervacija) error {
	if r.Kolicina == 0 {
		return nil
	}
	konto := strings.TrimSpace(r.Konto)
	id, err := repo.DrstaIDZaIzmenu(ctx, tx, konto, r.Sifra, r.Mag, r.Otk, r.Serija, r.Roktr)
	switch {
	case err == nil:
		return repo.StanjeIzmeni(ctx, tx, repository.TabelaDrsta, id, []repository.Izmena{{Kolona: "rez", Dodaj: true, Vrednost: r.Kolicina}})
	case errors.Is(err, sql.ErrNoRows):
		if r.Kolicina > 0 {
			return repo.StanjeDodaj(ctx, tx, repository.TabelaDrsta, komNoviLot(r, konto))
		}
		return nil
	}
	return err
}

// komNoviLot are the columns of a new row of drsta that holds only a reservation (the legacy AddEPRO "dodaj slog u
// drsta"): the lot of the artikal in the konto of the magacin, the cena of the stavka, the prometi 0 and the rezervisana
// količina.
func komNoviLot(r domain.KomNalogRezervacija, konto string) map[string]any {
	v := map[string]any{"rsifid": r.RsifID, "otk": r.Otk, "serija": r.Serija, "roktr": r.Roktr, "sifra": r.Sifra, "konto": konto, "mag": r.Mag,
		"magaciniid": r.MagaciniID, "cena": r.Cena, "dug": 0, "pot": 0, "ulaz": 0, "izlaz": 0, "pov": 0, "pos": 0, "rez": r.Kolicina}
	for m := 1; m <= 12; m++ {
		for _, k := range []string{"mdug", "mul", "mpot", "miz"} {
			v[fmt.Sprintf("%s_%d", k, m)] = 0
		}
	}
	return v
}

// GetNalog returns the content of the entry dialog with the nalog of the id (see GetProfaktura) and what only a nalog
// has: its profaktura, the bank of the payment of an izvoz, its status and the finansijsko stanje of the kupac on its
// datum. A new one has the first način isporuke (KAMION).
func (s *NalogResource) GetNalog(ctx context.Context, rdokID int64, vrd, magaciniID int) (domain.KomNalogDokument, error) {
	pro, err := s.GetProfaktura(ctx, rdokID, vrd, magaciniID)
	dok := domain.KomNalogDokument{KomProDokument: pro}
	if err != nil {
		return dok, err
	}
	h := pro.Header
	if h.RdokID > 0 {
		if dok.Dodatno, err = s.repo.KomNalogDodatno(ctx, h.RdokID); err != nil {
			return dok, err
		}
	}
	dok.StatusOpis = KomNalogStatusOpis(h.Doksts)
	if h.RdokID <= 0 {
		dok.StatusOpis = KomNalogStatusOpis(domain.KomNalogNov)
	}
	dok.Saldo, err = s.saldo(ctx, h.Fkto, h.Fana, h.Datum)
	return dok, err
}

// ValidateNalog validates the header of a nalog (the legacy ValidacijaEDOK, see ValidateProfaktura) and what only a
// nalog has: the kurs is greater than 0, the poreska kategorija is given when the firm is not a PDV obveznik, a given
// bank of the payment of an izvoz exists and a chosen profaktura (only while the nalog has none) is a potvrđena one of
// the magacin of the nalog.
func (s *NalogResource) ValidateNalog(ctx context.Context, params domain.KomNalogParams) []domain.FieldError {
	errs := s.ValidateProfaktura(ctx, params.KomProParams)
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	if !komImaGresku(errs, "kurs") && komBroj(params.Kurs, 0) <= 0 {
		add("kurs", komMsgNalKurs)
	}
	if session := domain.GetSessionFromStdContext(ctx); session != nil && strings.TrimSpace(params.Keykat) == "" {
		if firma, err := s.repo.KomFirma(ctx, session.Firma); err != nil {
			add("keykat", common.ErrMsgReadData)
		} else if !firma.Obv {
			add("keykat", komMsgNalPorkat)
		}
	}
	if id := common.StringToInt(params.BnkizvID); id > 0 {
		if _, found, err := s.repo.KomNalogBanka(ctx, int64(id)); err != nil {
			add("bnkizvid", common.ErrMsgReadData)
		} else if !found {
			add("bnkizvid", komMsgNalBanka)
		}
	}
	proID := int64(common.StringToInt(params.ProfakturaID))
	if proID <= 0 {
		return errs
	}
	magaciniID := common.StringToInt(params.MagaciniID)
	if params.RdokID > 0 {
		dodatno, err := s.repo.KomNalogDodatno(ctx, params.RdokID)
		if err != nil {
			add("profakturaid", common.ErrMsgReadData)
			return errs
		}
		if dodatno.VrdokID > 0 {
			return errs
		}
		n, _, err := s.repo.KomNalog(ctx, params.RdokID)
		if err != nil {
			add("profakturaid", common.ErrMsgReadData)
			return errs
		}
		magaciniID = n.MagaciniID
	}
	if pro, found, err := s.repo.KomProfaktura(ctx, proID); err != nil {
		add("profakturaid", common.ErrMsgReadData)
	} else if !found || pro.Doksts != domain.KomProfakturaPotvrdjena {
		add("profakturaid", komMsgNalProfaktura)
	} else if pro.MagaciniID != magaciniID {
		add("profakturaid", komMsgNalProMagacin)
	}
	return errs
}

// komImaGresku tells whether the field already has an error.
func komImaGresku(errs []domain.FieldError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

// SaveNalog saves the header of a validated nalog (the legacy SnimiEDOK, see SaveProfaktura) with its bank of the
// payment of an izvoz in one transaction. A chosen profaktura (a new nalog, or a saved one without a profaktura; the
// legacy KopirajStavkeizProfakture) is linked (vrdokid), its stavke are copied after the stavke of the nalog with their
// zalihe reserved (komRezervisi) and it is marked (doksts 40). It returns the rdokid.
func (s *NalogResource) SaveNalog(ctx context.Context, params domain.KomNalogParams) (int64, error) {
	sifbank := 0
	if id := common.StringToInt(params.BnkizvID); id > 0 {
		var err error
		if sifbank, _, err = s.repo.KomNalogBanka(ctx, int64(id)); err != nil {
			return 0, err
		}
	}
	var vrdokID int64
	rbrOd := 1
	if params.RdokID > 0 {
		dodatno, err := s.repo.KomNalogDodatno(ctx, params.RdokID)
		if err != nil {
			return 0, err
		}
		vrdokID = dodatno.VrdokID
		if rbrOd, err = s.repo.KomZahtevNoviRbr(ctx, nil, params.RdokID); err != nil {
			return 0, err
		}
	}
	proID := int64(common.StringToInt(params.ProfakturaID))
	var proStavke []domain.KomProStavkaRow
	kopiraj := vrdokID == 0 && proID > 0
	if kopiraj {
		var err error
		if proStavke, err = s.repo.KomProStavke(ctx, proID); err != nil {
			return 0, err
		}
		vrdokID = proID
	}
	return s.sacuvaj(ctx, params.KomProParams, func(tx db.Transaction, p domain.KomProfakturaUpis) error {
		if err := s.repo.KomNalogDodatnoSacuvaj(ctx, tx, p.RdokID, vrdokID, sifbank); err != nil {
			return err
		}
		if !kopiraj {
			return nil
		}
		ok, err := s.repo.KomNalogProfakturaVezi(ctx, tx, proID)
		if err != nil {
			return err
		}
		if !ok {
			return errKomNalogProfaktura
		}
		dadok := sql.NullTime{Time: p.Dadok, Valid: !p.Dadok.IsZero()}
		if err := s.repo.KomNalogKopirajStavke(ctx, tx, proID, p.RdokID, p.Vrd, p.Dokum, dadok, rbrOd); err != nil {
			return err
		}
		nalog := domain.KomProfakturaRow{MagaciniID: p.MagaciniID, Mag: p.Mag}
		for _, st := range proStavke {
			if err := komRezervisi(ctx, s.repo, tx, komRezervacija(nalog, st, st.Kolicina)); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteNalog deletes a nalog with its stavke (the legacy BTN_Brisanje): not a fiscalized (knjige 4 "S"), posted
// (knjige 1 "D"), potvrđen or otpremljen one, whose message is returned instead. In one transaction the zalihe of its
// stavke are released (komRezervisi) and its profaktura becomes potvrđena again (doksts 20; the legacy sets 10).
func (s *NalogResource) DeleteNalog(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.repo.KomNalog(ctx, rdokID)
	if err != nil {
		return "", err
	}
	switch {
	case !found:
		return komMsgNalBrisanjeNema, nil
	case strings.EqualFold(strings.TrimSpace(p.Knjige4), "S"):
		return komMsgFiskalizovan, nil
	case strings.EqualFold(strings.TrimSpace(p.Knjige1), "D"):
		return komMsgProknjizen, nil
	case komNalogZakljucan(p.Doksts) != "":
		return komNalogZakljucan(p.Doksts), nil
	}
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	if err != nil {
		return "", err
	}
	dodatno, err := s.repo.KomNalogDodatno(ctx, rdokID)
	if err != nil {
		return "", err
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	for _, st := range stavke {
		if err := komRezervisi(ctx, s.repo, tx, komRezervacija(p, st, -st.Kolicina)); err != nil {
			return "", err
		}
	}
	if err := s.repo.KomZahtevBrisi(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if dodatno.VrdokID > 0 {
		// A profaktura deleted or changed meanwhile has nothing to set back.
		if _, err := s.repo.KomNalogProfakturaVrati(ctx, tx, dodatno.VrdokID); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje naloga: %w", err)
	}
	return "", nil
}

// ProveriPotvrdu checks a nalog before the question of its potvrda (the legacy BTN_Potvrda): it has stavke, every
// stavka has a prodajna cena greater than 0 and a cena other than 0, and it is new; the message why it cannot be
// confirmed is returned.
func (s *NalogResource) ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.repo.KomNalog(ctx, rdokID)
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
	for _, st := range stavke {
		if st.Fcena <= 0 || st.Cena == 0 {
			return komMsgNalCenaNula, nil
		}
	}
	if p.Doksts != domain.KomNalogNov {
		return komMsgVecPotvrdjen, nil
	}
	return "", nil
}

// PotvrdiNalog confirms a nalog (the legacy "Da" of BTN_Potvrda: doksts 20) after the checks of ProveriPotvrdu; it
// returns their message, or "Dokument je potvrđen!!!".
func (s *NalogResource) PotvrdiNalog(ctx context.Context, rdokID int64) (string, error) {
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

// GetProfakturaNaloga returns what the header of a nalog takes of a potvrđena profaktura of its magacin (the legacy
// PodaciIzProfakture: the datum, the rok, the kupac, the tekst, the valuta, the org. jedinica and the mesto troška, the
// ugovoreni rabat and the kasa — its percentage; the legacy takes the vrednost —, the žiro račun and the poreska
// kategorija); found is false for another one.
func (s *NalogResource) GetProfakturaNaloga(ctx context.Context, rdokID int64, magaciniID int) (domain.KomNalogProfaktura, bool, error) {
	if rdokID <= 0 {
		return domain.KomNalogProfaktura{}, false, nil
	}
	p, found, err := s.repo.KomProfaktura(ctx, rdokID)
	if err != nil || !found || p.Doksts != domain.KomProfakturaPotvrdjena || p.MagaciniID != magaciniID {
		return domain.KomNalogProfaktura{}, false, err
	}
	return domain.KomNalogProfaktura{RdokID: p.RdokID, Dokum: p.Dokum, Dadok: p.Dadok, Datum: komDatumText(p.Dadok), MagaciniID: p.MagaciniID,
		Rok: p.Rok, Fkto: p.Fkto, Fana: p.Fana, Foot: p.Foot, SifVal: p.SifVal, OznVal: p.OznVal, VaziZa: p.VaziZa, Orgjed: p.Orgjed.Int64,
		OrgjedText: p.OrgjedText, MestoTroska: p.MestoTroska.Int64, MestoTroskaText: p.MestoTroskaText, Ugrabat: p.Ugrabat, Pkase: p.Pkase,
		Posuslnab: p.Posuslnab, Keykat: p.Keykat, Category: p.Category}, true, nil
}

// GetSaldo returns the finansijsko stanje of the kupac of the konto and šifra on the datum (today without a valid one;
// the legacy SaldoKUPCA); zero without the kupac.
func (s *NalogResource) GetSaldo(ctx context.Context, konto, sifra, datum string) (domain.KomNalogSaldo, error) {
	d, ok := komDatum(datum)
	if !ok {
		d = s.now()
	}
	return s.saldo(ctx, konto, sifra, d)
}

// saldo reads the promet and the neproknjižene fakture of the kupac and calculates its finansijsko stanje.
func (s *NalogResource) saldo(ctx context.Context, konto, sifra string, danas time.Time) (domain.KomNalogSaldo, error) {
	if strings.TrimSpace(konto) == "" || strings.TrimSpace(sifra) == "" {
		return domain.KomNalogSaldo{}, nil
	}
	fpro, err := s.repo.KomNalogFpro(ctx, konto, sifra)
	if err != nil {
		return domain.KomNalogSaldo{}, err
	}
	robne, err := s.repo.KomNalogRobneStavke(ctx, konto, sifra)
	if err != nil {
		return domain.KomNalogSaldo{}, err
	}
	return komNalogSaldo(fpro, robne, danas), nil
}

// komDospelo tells whether a document of the datum and the rok is due on the day (the legacy dDOSPECE <= dDANAS; a
// document without a datum is due).
func komDospelo(dadok sql.NullTime, rok int, danas time.Time) bool {
	if !dadok.Valid {
		return true
	}
	dan := time.Date(danas.Year(), danas.Month(), danas.Day(), 0, 0, 0, 0, time.UTC)
	d := dadok.Time
	return !time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, rok).After(dan)
}

// komNalogSaldo is the finansijsko stanje of a kupac on a day (the legacy SaldoKUPCA):
//   - proknjiženo (fpro): the realizacija is the sum of the fakture (kategorije 1 and 2), split into dospelo and
//     nedospelo by their dospeće, plaćeno the sum of the uplate (3 and 4); the uplate close the dospelo first, the rest
//     of them the nedospelo; the saldo is realizacija - plaćeno;
//   - neproknjiženo (the robne fakture): per faktura the iznos (Σ količina × cena), the rabat, the ugovoreni rabat and
//     the kasa on what is left and the PDV of the stavke (rounded per stavka), split by its dospeće;
//   - ukupno: the dospelo, the nedospelo and their sum.
func komNalogSaldo(fpro []domain.KomNalogFpro, robne []domain.KomNalogRobnaStavka, danas time.Time) domain.KomNalogSaldo {
	var s domain.KomNalogSaldo
	var dospelo, nedospelo float64
	for _, f := range fpro {
		switch f.Kat {
		case 1, 2:
			s.Realizacija += f.Iznos
			if komDospelo(f.Dadok, f.Rok, danas) {
				dospelo += f.Iznos
			} else {
				nedospelo += f.Iznos
			}
		case 3, 4:
			s.Placeno += f.Iznos
		}
	}
	s.Dospelo, s.Nedospelo = 0, nedospelo
	if dospelo-s.Placeno > 0 {
		s.Dospelo = dospelo - s.Placeno
	}
	if dospelo-s.Placeno < 0 {
		s.Nedospelo = nedospelo - (s.Placeno - dospelo)
	}
	s.Saldo = s.Realizacija - s.Placeno
	for i := 0; i < len(robne); {
		f := robne[i]
		var iznos, rabat, porez float64
		for ; i < len(robne) && robne[i].RdokID == f.RdokID; i++ {
			st := robne[i]
			iznos += st.Kolic * st.Fcena
			rabat += st.Kolic * st.Fcena * st.Rab / 100
			porez += komRound2(st.Kolic * st.Fcena * st.Stopa / 100)
		}
		iznos, rabat = komRound2(iznos), komRound2(rabat)
		ugr := komRound2((iznos - rabat) * f.Ugrabat / 100)
		kasa := komRound2((iznos - rabat - ugr) * f.Pkase / 100)
		faktura := iznos - rabat - ugr - kasa + porez
		if komDospelo(f.Dadok, f.Rok, danas) {
			s.DospeloRob += faktura
		} else {
			s.NedospeloRob += faktura
		}
	}
	s.RealizacijaRob = s.DospeloRob + s.NedospeloRob
	s.UkDospelo = s.Dospelo + s.DospeloRob
	s.UkNedospelo = s.Nedospelo + s.NedospeloRob
	s.UkSaldo = s.UkDospelo + s.UkNedospelo
	return s
}

// ValidateNalogStavka validates a stavka of a nalog (the legacy ValidacijaEPRO, see ValidateStavka): a new one in a
// magacin of the tip G needs the planska cena of its artikal (the default normativ).
func (s *NalogResource) ValidateNalogStavka(ctx context.Context, params domain.KomNalogStavkaParams) []domain.FieldError {
	errs := s.ValidateStavka(ctx, params.KomProStavkaParams)
	if params.RproID > 0 || len(errs) > 0 {
		return errs
	}
	p, found, err := s.repo.KomNalog(ctx, params.RdokID)
	if err != nil || !found || p.Tipmag != "G" {
		return errs
	}
	sifra, _ := strconv.Atoi(strings.TrimSpace(params.Sifra))
	if cena, _, err := s.repo.KomProPlanskaCena(ctx, sifra); err != nil {
		errs = append(errs, domain.FieldError{Field: "sifraartikla", ErrorMessage: common.ErrMsgReadData})
	} else if cena == 0 {
		errs = append(errs, domain.FieldError{Field: "sifraartikla", ErrorMessage: komMsgNalPlanskaCena})
	}
	return errs
}

// SaveNalogStavka saves a validated stavka of a nalog in one transaction and calculates the totals of the nalog again:
//   - a new one (the legacy AddEPRO) with the next redni broj: with "Skidaj zalihe po starosti" its količina is split
//     over the lots of the artikal in its konto with a slobodna količina, the oldest rok trajanja first (one stavka per
//     lot), and what is left goes to a stavka without a lot; else one stavka of the chosen lot. The zalihe of each
//     stavka are reserved (komRezervisi);
//   - a changed one (the legacy exit of a changed row): its količina, prodajna cena and rabat (its planska cena stays;
//     the legacy overwrites it with the prodajna cena), the reservation changed by the difference of the količina.
func (s *NalogResource) SaveNalogStavka(ctx context.Context, params domain.KomNalogStavkaParams) error {
	p, found, err := s.repo.KomNalog(ctx, params.RdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("nalog %d ne postoji", params.RdokID)
	}
	ino, err := s.ino(ctx, p.SifVal)
	if err != nil {
		return err
	}
	kolicina, rabat := common.StringToFloat64(params.Kolicina), common.StringToFloat64(params.Rabat)
	fcena, cenaVal := komProCene(common.StringToFloat64(params.Cena), ino, p.Kurs, p.VaziZa)
	if params.RproID > 0 {
		return s.izmeniStavku(ctx, p, params.RproID, kolicina, fcena, cenaVal, rabat)
	}
	sifra, _ := strconv.Atoi(strings.TrimSpace(params.Sifra))
	a, found, err := s.repo.KomPrijArtikal(ctx, sifra)
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
	st := domain.KomProStavkaUpis{RdokID: p.RdokID, Sifra: a.Sifra, RsifID: a.RsifID, Po: a.Po, Konto: strings.TrimSpace(params.Konto), Mag: p.Mag,
		MagaciniID: p.MagaciniID, Kolicina: kolicina, Cena: planska, Fcena: fcena, CenaVal: cenaVal, Rabat: rabat, Otk: strings.TrimSpace(params.Otk),
		Serija: strings.TrimSpace(params.Serija), Roktr: komWinDevDatum(params.Rok), Vrd: p.Vrd, Dokum: p.Dokum, Dadok: p.Dadok}
	stavke := []domain.KomProStavkaUpis{st}
	if common.StringToBool(params.PoStarosti) && p.Tipzal == 2 {
		lotovi, err := s.repo.KomNalogLotoviStarost(ctx, p.MagaciniID, st.Konto, a.Sifra)
		if err != nil {
			return err
		}
		stavke = komNalogPoStarosti(st, lotovi)
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	rbr, err := s.repo.KomZahtevNoviRbr(ctx, tx, p.RdokID)
	if err != nil {
		return err
	}
	for i, st := range stavke {
		st.Rbr = rbr + i
		if err := s.repo.KomProStavkaDodaj(ctx, tx, st); err != nil {
			return err
		}
		r := domain.KomNalogRezervacija{MagaciniID: p.MagaciniID, Mag: p.Mag, RsifID: st.RsifID, Sifra: st.Sifra, Konto: st.Konto, Otk: st.Otk,
			Serija: st.Serija, Roktr: st.Roktr, Cena: st.Cena, Kolicina: st.Kolicina}
		if err := komRezervisi(ctx, s.repo, tx, r); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka naloga: %w", err)
	}
	return s.preracunaj(ctx, p.RdokID)
}

// komNalogPoStarosti splits a stavka over the lots by their age (the legacy AddEPRO with CBOX_ZALSTAROST): each lot
// gets the količina still to split up to its slobodna količina, in the order of the lots; what is left goes to a
// stavka without a lot (OTK, serija and rok trajanja empty) in the konto of the stavka. Without lots the whole
// količina is without a lot (the legacy adds nothing then).
func komNalogPoStarosti(st domain.KomProStavkaUpis, lotovi []domain.KomNalogLot) []domain.KomProStavkaUpis {
	ostatak := st.Kolicina
	var stavke []domain.KomProStavkaUpis
	for _, lot := range lotovi {
		if ostatak <= 0 {
			break
		}
		kol := min(ostatak, lot.Slobodno)
		if kol <= 0 {
			continue
		}
		s := st
		s.Konto, s.Otk, s.Serija, s.Roktr, s.Kolicina = lot.Konto, lot.Otk, lot.Serija, lot.Roktr, kol
		stavke = append(stavke, s)
		ostatak -= kol
	}
	if ostatak > 0 {
		s := st
		s.Otk, s.Serija, s.Roktr, s.Kolicina = "", "", 0, ostatak
		stavke = append(stavke, s)
	}
	return stavke
}

// izmeniStavku changes a stavka of a nalog and its reservation by the difference of the količina.
func (s *NalogResource) izmeniStavku(ctx context.Context, p domain.KomProfakturaRow, rproID int64, kolicina, fcena, cenaVal, rabat float64) error {
	staro, found, err := s.proStavka(ctx, p.RdokID, rproID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("stavka %d naloga %d ne postoji", rproID, p.RdokID)
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomProStavkaIzmeni(ctx, tx, domain.KomProStavkaIzmena{RdokID: p.RdokID, RproID: rproID, Kolicina: kolicina, Fcena: fcena,
		CenaVal: cenaVal, Rabat: rabat}); err != nil {
		return err
	}
	if err := komRezervisi(ctx, s.repo, tx, komRezervacija(p, staro, kolicina-staro.Kolicina)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka naloga: %w", err)
	}
	return s.preracunaj(ctx, p.RdokID)
}

// DeleteNalogStavka deletes a stavka of a new nalog (the legacy BTN_BrisiDET: its zalihe released, the redni brojevi
// after it move one back) and calculates the totals again; the message why is returned when it cannot.
func (s *NalogResource) DeleteNalogStavka(ctx context.Context, rdokID, rproID int64) (string, error) {
	p, poruka, err := s.izmenljiva(ctx, rdokID)
	if err != nil || poruka != "" {
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
	if err := komRezervisi(ctx, s.repo, tx, komRezervacija(p, st, -st.Kolicina)); err != nil {
		return "", err
	}
	if err := s.repo.KomZahtevStavkaBrisi(ctx, tx, rdokID, st.RproID, st.Rbr); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje stavke naloga: %w", err)
	}
	return "", s.preracunaj(ctx, rdokID)
}

// GetNalogPrikaz returns the nalozi of the lists of the main screen: the nepotvrđeni or the potvrđeni (any status other
// than 10), or the delimično realizovani; nothing without a magacin (unless all of them).
func (s *NalogResource) GetNalogPrikaz(ctx context.Context, params domain.KomNalogPrikazParams) ([]domain.KomNalogPrikazRow, error) {
	if params.Doksts != domain.KomNalogNov {
		params.Doksts = domain.KomNalogPotvrdjen
	}
	if !params.SviMagacini && params.MagaciniID <= 0 {
		return nil, nil
	}
	return s.repo.KomNalogPrikaz(ctx, params)
}

// GetNalogStampa returns the print of a nalog (the legacy PR_RPT_NALOGPRODAJA): the firm with the address of the
// magacin, the kupac, the magacin, the org. jedinica, the tekući račun (the žiro račun of the nalog), its broj and datum
// (the legacy prints the broj and the datum of the otpremnica, empty before it), the klasa cena of its cenovnik, the
// napomena, the stavke with the tražena količina, the ukupna težina and zapremina, and whether it is not confirmed; a
// nalog without stavke is not printed (the message why is returned).
func (s *NalogResource) GetNalogStampa(ctx context.Context, rdokID int64) (domain.KomNalogStampa, string, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomNalogStampa{}, "", fmt.Errorf("no user session found")
	}
	p, found, err := s.repo.KomNalog(ctx, rdokID)
	if err != nil {
		return domain.KomNalogStampa{}, "", err
	}
	if !found {
		return domain.KomNalogStampa{}, komMsgBezStavkiStampa, nil
	}
	stavke, err := s.repo.KomProStavke(ctx, rdokID)
	if err != nil {
		return domain.KomNalogStampa{}, "", err
	}
	if len(stavke) == 0 {
		return domain.KomNalogStampa{}, komMsgBezStavkiStampa, nil
	}
	st := domain.KomNalogStampa{Magacin: p.MagacinText, OJ: p.OrgjedText, TekuciRacun: strings.TrimSpace(p.Posuslnab), Dokum: p.Dokum, Datum: p.Dadok,
		Napomena: p.Foot, Lot: p.Tipzal == 2, Nepotvrdjen: p.Doksts == domain.KomNalogNov}
	if st.Firma, err = s.repo.KomFirma(ctx, session.Firma); err != nil {
		return domain.KomNalogStampa{}, "", err
	}
	if adresa := strings.TrimSpace(p.MagacinAdresa); adresa != "" {
		st.Firma.Adresa = strings.TrimSpace(st.Firma.Adresa) + " , " + adresa
	}
	if idFkpl, err := s.repo.KomPorIDFkpl(ctx, p.Fkto, p.Fana); err != nil {
		return domain.KomNalogStampa{}, "", err
	} else if idFkpl > 0 {
		if st.Kupac, _, err = s.repo.KomPonudaStampaDobavljac(ctx, idFkpl); err != nil {
			return domain.KomNalogStampa{}, "", err
		}
	}
	if p.CenovnikID > 0 {
		naziv, _, broj, _, found, err := s.repo.KomProCenovnik(ctx, p.CenovnikID)
		if err != nil {
			return domain.KomNalogStampa{}, "", err
		}
		if found {
			st.KlasaCena = fmt.Sprintf("%d- %s", broj, naziv)
		}
	}
	for i, stavka := range stavke {
		opis := stavka.Naziv
		if st.Lot {
			opis += "\nSerija: " + stavka.Serija + " Rok trajanja: " + komWinDevDatumText(stavka.Roktr)
		}
		st.Stavke = append(st.Stavke, domain.KomNalogStampaStavka{Rbr: i + 1, Sifra: stavka.Sifra, Opis: opis, Otk: stavka.Otk,
			Serija: stavka.Serija, Rok: komWinDevDatumText(stavka.Roktr), JM: stavka.JM, Kolicina: stavka.Kolicina, Decimale: stavka.Decimale})
		st.Tezina += stavka.Kolicina * stavka.Tezjm
		st.Zapremina += stavka.Kolicina * stavka.Kubikaza
	}
	return st, "", nil
}

// NalogPrikazTableFields are the columns of the lists of the main screen (the legacy TABLE_QRY_EDOK and
// TABLE_QRY_EDOK1).
func (s *NalogResource) NalogPrikazTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "magacin", Label: "Magacin", Width: "12"},
		{Name: "vrd", Label: "Vrsta", Width: "5", TextAlign: "right"},
		{Name: "dokum", Label: "Broj naloga", Width: "6", TextAlign: "right"},
		{Name: "dadok", Label: "Datum", Width: "8"},
		{Name: "otpremnica", Label: "Broj otpremn.", Width: "6", TextAlign: "right"},
		{Name: "faktura", Label: "Br. fakture", Width: "6", TextAlign: "right"},
		{Name: "fkto", Label: "Konto", Width: "6"},
		{Name: "fana", Label: "Šifra", Width: "6"},
		{Name: "kupac", Label: "Naziv partnera", Width: "24"},
		{Name: "rok", Label: "Rok plaćanja", Width: "5", TextAlign: "right"},
		{Name: "ugrabat", Label: "Ugov. rabat", Width: "5", TextAlign: "right"},
		{Name: "status", Label: "Status", Width: "14"},
	}
}

// NalogPrikazStavkeTableFields are the columns of the stavke of the selected nalog of the main screen (the legacy
// TABLE_QRY_EPRO of "Pregled naloga za isporuku": the naručena količina, the cena and the cena in the valuta; of
// "Pregled delimično realizovanih naloga" (delimicni, TABLE_QRY_EPRO1): the naručena and the isporučena količina).
func (s *NalogResource) NalogPrikazStavkeTableFields(delimicni bool) []domain.Fields {
	fields := []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "5", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "7", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla/proizvoda", Width: "26"},
		{Name: "jm", Label: "JM", Width: "4"},
		{Name: "otk", Label: "OTK", Width: "5"},
		{Name: "serija", Label: "Serija", Width: "7"},
		{Name: "rok", Label: "Rok upotrebe", Width: "7"},
		{Name: "kolicina", Label: "Naručena količina", Width: "8", TextAlign: "right"},
	}
	if delimicni {
		fields = append(fields, domain.Fields{Name: "isporuceno", Label: "Isporučena količina", Width: "8", TextAlign: "right"})
	}
	fields = append(fields, domain.Fields{Name: "cena", Label: "Jedinična cena", Width: "8", TextAlign: "right"})
	if !delimicni {
		fields = append(fields, domain.Fields{Name: "cenaval", Label: "Cena u valuti", Width: "8", TextAlign: "right"})
	}
	return append(fields,
		domain.Fields{Name: "rabat", Label: "% Rabata", Width: "5", TextAlign: "right"},
		domain.Fields{Name: "vrednost", Label: "Vrednost", Width: "10", TextAlign: "right", IncludeInTotals: true},
	)
}

// NalogPrikazTableRows are the rows of the lists of the main screen: "Izmeni" and "Briši" of every new nalog (izmena),
// the status as its text.
func NalogPrikazTableRows(nalozi []domain.KomNalogPrikazRow, izmena bool) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(nalozi))
	broj := func(n int) string {
		if n <= 0 {
			return ""
		}
		return strconv.Itoa(n)
	}
	for _, n := range nalozi {
		menja := izmena && n.Doksts == domain.KomNalogNov
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(n.RdokID, 10), HasUpdate: menja, HasDelete: menja, Fields: []string{
			n.Magacin, strconv.Itoa(n.Vrd), strconv.Itoa(n.Dokum), komDatumText(n.Dadok), broj(n.Otpremnica), broj(n.Faktura), n.Fkto, n.Fana,
			n.Kupac, strconv.Itoa(n.Rok), common.FormatNumberWithSystemLocale(n.Ugrabat, 2), KomNalogStatusOpis(n.Doksts),
		}})
	}
	return rows
}

// NalogPrikazStavkeTableRows are the rows of the stavke of the selected nalog of the main screen with their total: the
// vrednost is the količina by the cena less the rabat (the legacy PuniTABLEDET: rounded to 2 decimals) — the naručena
// količina, of the delimično realizovani (delimicni) the isporučena one, whose rows that differ from the naručena are
// marked (ClassRow komNalogRazlika).
func NalogPrikazStavkeTableRows(stavke []domain.KomProStavkaRow, delimicni bool) ([]domain.TableRow, float64) {
	rows := make([]domain.TableRow, 0, len(stavke))
	ukupno := 0.0
	for _, st := range stavke {
		kolicina := st.Kolicina
		if delimicni {
			kolicina = st.Isporuceno
		}
		vrednost := komRound2(kolicina*st.Fcena - kolicina*st.Fcena*st.Rabat/100)
		ukupno += vrednost
		fields := []string{strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM, st.Otk, st.Serija, komWinDevDatumText(st.Roktr),
			komProKolicinaText(st.Kolicina, st.Decimale)}
		row := domain.TableRow{ID: strconv.FormatInt(st.RproID, 10)}
		if delimicni {
			fields = append(fields, komProKolicinaText(st.Isporuceno, st.Decimale), common.FormatNumberWithSystemLocale(st.Fcena, 2))
			if st.Isporuceno != st.Kolicina {
				row.ClassRow = komNalogRazlika
			}
		} else {
			fields = append(fields, common.FormatNumberWithSystemLocale(st.Fcena, 2), common.FormatNumberWithSystemLocale(st.CenaVal, 2))
		}
		row.Fields = append(fields, common.FormatNumberWithSystemLocale(st.Rabat, 2), common.FormatNumberWithSystemLocale(vrednost, 2))
		rows = append(rows, row)
	}
	return rows, ukupno
}

// komNalogRazlika is the class of a row of a stavka whose isporučena količina differs from the naručena one (the legacy
// red rows of TABLE_QRY_EPRO1).
const komNalogRazlika = common.ClassRowRazlika
