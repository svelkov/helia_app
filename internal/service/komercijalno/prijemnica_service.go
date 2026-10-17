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
)

// The messages of the prijemnica materijala (the legacy pr_PrijemnicaSIROVINE); the ones it shares with the other
// documents are the komMsg constants of zahtevnabavka_service.go and porudzbenica_service.go.
const (
	komMsgPrijPorudzbenica = "Niste izabrali narudžbenicu po koju radite prijemnicu!!!"
	komMsgPrijDatum        = "Morate uneti datum prijemnice..."
	komMsgPrijDatumPor     = "Datum prijemnice ne može biti manji od datuma porudžbenice..."
	komMsgPrijNijeIzmena   = "Potvrđena prijemnica se ne može menjati."
	komMsgPrijDupli        = "Ovakvi podaci već postoje. Nedozvoljen dupli unos!!!"
	komMsgPrijKonto        = "Konto nije definisan za izabrani magacin..."
	komMsgPrijTrazKolic    = "Količina na dokumentu mora biti veća od 0!!!!"
	komMsgPrijKolic        = "Primljena količina mora biti veća od 0!!!!"
	komMsgPrijSerija       = "Morate uneti broj serije!!!!"
	komMsgPrijNekompletne  = "Stavke u dokumentu nisu kompletne, fali podatak za seriju ili primlj. količinu, dokument ne može biti potvrđen!!!"
	komMsgPrijBezKonta     = "Stavke u dokumentu nisu kompletne, fali podatak za konto magacina, dokument ne može biti potvrđen!!!"
	komMsgPrijBezPakovanja = "Ne postoji nijedna uneta stavka, nedozvoljen unos pakovanja!!!"
	komMsgPrijPakovanje    = "Količina po pakovanju i broj pakovanja ne mogu biti manji od 0!!!"
	komMsgPrijKarton       = "Niste selektovali stavku za karton identifikacije!!!"
)

// PrijemnicaService is the "Prijemnica materijala po porudžbenici" screen of the komercijalno (the legacy
// pr_PrijemnicaSIROVINE).
type PrijemnicaService interface {
	GetPrijemnica(ctx context.Context, rdokID int64, vrd, magaciniID int, vezaID int64) (domain.KomPrijDokument, error)
	GetMagacini(ctx context.Context) ([]domain.ComboItem, error)
	ValidatePrijemnica(ctx context.Context, params domain.KomPrijParams) []domain.FieldError
	SavePrijemnica(ctx context.Context, params domain.KomPrijParams) (int64, error)
	DeletePrijemnica(ctx context.Context, rdokID int64) (string, error)
	ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error)
	PotvrdiPrijemnicu(ctx context.Context, rdokID int64) (string, error)
	GetStavke(ctx context.Context, rdokID int64) ([]domain.KomPrijStavkaRow, error)
	GetArtikal(ctx context.Context, rdokID int64, sifra string) (domain.KomPrijArtikal, bool, error)
	ValidateStavka(ctx context.Context, params domain.KomPrijStavkaParams) []domain.FieldError
	SaveStavka(ctx context.Context, params domain.KomPrijStavkaParams) error
	DeleteStavka(ctx context.Context, rdokID, rproID int64) (string, error)
	GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomPrijStavkaForma, bool, error)
	Preuzmi(ctx context.Context, rdokID, rproID int64) (string, error)
	VratiPoslednju(ctx context.Context, rdokID int64) (string, error)
	VratiSve(ctx context.Context, rdokID int64) (string, error)
	GetPakovanja(ctx context.Context, rdokID int64) ([]domain.KomPrijPakovanjeRow, string, error)
	SavePakovanje(ctx context.Context, params domain.KomPrijPakovanjeParams) (string, error)
	GetPrikaz(ctx context.Context, params domain.KomPrijPrikazParams) ([]domain.KomPrijPrikazRow, error)
	GetStampa(ctx context.Context, rdokID int64) (domain.KomPrijStampa, string, error)
	GetKarton(ctx context.Context, rdokID, rproID int64) (domain.KomPrijKarton, string, error)
	StavkeTableFields() []domain.Fields
	IzvoriTableFields() []domain.Fields
	PrikazTableFields() []domain.Fields
	PrikazStavkeTableFields() []domain.Fields
	PakovanjaTableFields() []domain.Fields
}

// prijRepository is what the service needs of repository.KomercijalnoRepository.
type prijRepository interface {
	BeginTx() (db.Transaction, error)
	KomPrijVrste(ctx context.Context, magaciniID int) ([]domain.ComboItem, error)
	KomPorNoviBroj(ctx context.Context, tx db.Transaction, vrd, mag int) (int, error)
	KomPrijemnica(ctx context.Context, rdokID int64) (domain.KomPrijemnicaRow, bool, error)
	KomPrijDodaj(ctx context.Context, tx db.Transaction, p domain.KomPrijemnicaUpis) (int64, error)
	KomPrijIzmeni(ctx context.Context, tx db.Transaction, p domain.KomPrijemnicaUpis) error
	KomPrijPotvrdi(ctx context.Context, tx db.Transaction, rdokID int64, doksts string) (bool, error)
	KomPrijBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomPrijVeza(ctx context.Context, rdokID int64) (domain.KomPrijVeza, bool, error)
	KomPorDobavljac(ctx context.Context, konto, sifra string) (domain.KomDobavljacPodaci, bool, error)
	KomPorIDFkpl(ctx context.Context, konto, sifra string) (int64, error)
	KomPorStavke(ctx context.Context, rdokID int64) ([]domain.KomPorStavkaRow, error)
	KomPrijStavke(ctx context.Context, rdokID int64) ([]domain.KomPrijStavkaRow, error)
	KomPrijStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomPrijStavkaUpis) error
	KomPrijStavkaIzmeni(ctx context.Context, tx db.Transaction, st domain.KomPrijStavkaIzmena) error
	KomPrijStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error
	KomPrijStavkeBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error
	KomPrijIzvor(ctx context.Context, tx db.Transaction, porudzbenicaID, rproID int64) (domain.KomPrijIzvor, bool, error)
	KomPrijImaArtikal(ctx context.Context, tx db.Transaction, rdokID, rsifID int64) (bool, error)
	KomPrijArtikal(ctx context.Context, sifra int) (domain.KomPrijArtikal, bool, error)
	KomPrijOtkBroj(ctx context.Context, tx db.Transaction, vrd, sifra int) (int, error)
	KomPrijKonta(ctx context.Context, magaciniID int) ([]domain.ComboItem, error)
	KomPrijPakovanja(ctx context.Context, rdokID int64) ([]domain.KomPrijPakovanjeRow, error)
	KomPrijPakovanjeSnimi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, kolPoPak, brPak float64) error
	KomPrijPrikaz(ctx context.Context, p domain.KomPrijPrikazParams) ([]domain.KomPrijPrikazRow, error)
	KomZahtevNoviRbr(ctx context.Context, tx db.Transaction, rdokID int64) (int, error)
	KomMagacin(ctx context.Context, magaciniID int) (int, bool, error)
	KomPonudaStampaDobavljac(ctx context.Context, idFkpl int64) (domain.KomPonudaStampaDobavljac, bool, error)
	KomFirma(ctx context.Context, firma string) (domain.RobnoStampaFakturaFirmaDto, error)
	BankeZiroRacuni(ctx context.Context) ([]string, error)
}

// prijStanje is what the service needs of the robno service: the stanje of the stavke of a confirmed prijemnica (the
// legacy GlobalProcROBNO.AzurRSTA_ZN and AzurDRSTA_ZN).
type prijStanje interface {
	AzurirajRSTAZN(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja, zaliha bool) error
	AzurirajDRSTAZN(ctx context.Context, tx db.Transaction, p domain.RobPromenaStanja, zaliha bool) error
}

// PrijemnicaResource implements PrijemnicaService.
type PrijemnicaResource struct {
	repo   prijRepository
	common magaciniKorisnika
	stanje prijStanje
	now    func() time.Time
}

// NewPrijemnicaService returns the service of the "Prijemnica materijala po porudžbenici" screen.
func NewPrijemnicaService(repo prijRepository, common magaciniKorisnika, stanje prijStanje) *PrijemnicaResource {
	return &PrijemnicaResource{repo: repo, common: common, stanje: stanje, now: time.Now}
}

// GetMagacini returns the magacini of the user of the tipovi P, S, D, L and V (the legacy ROB_QRY_MAGUSER with
// "P;S;D;L;V"; the filter of the main screen).
func (s *PrijemnicaResource) GetMagacini(ctx context.Context) ([]domain.ComboItem, error) {
	return s.common.GetMagacinKorisnikaComboValues(ctx, komPorMagacini())
}

// komPrijOtk is the OTK of a new stavka of an artikal: the number of its stavke in the potvrđene prijemnice of the vrsta
// + 1 with three digits (the legacy VratiOTK: "001").
func komPrijOtk(broj int) string {
	return fmt.Sprintf("%03d", broj+1)
}

// komPrijDatum is the datum of a prijemnica: the datum prijemnice (datotp), else its datum.
func komPrijDatum(p domain.KomPrijemnicaRow) sql.NullTime {
	if p.Datotp.Valid {
		return p.Datotp
	}
	return p.Dadok
}

// komPrijKolicina is a primljena količina with the sign of the predznak of the vrsta (negative for a storno).
func komPrijKolicina(kolicina float64, storno bool) float64 {
	if storno {
		return -math.Abs(kolicina)
	}
	return math.Abs(kolicina)
}

// GetPrijemnica returns the content of the entry dialog with the prijemnica of the id (a nepotvrđena one, the legacy
// BTN_TraziDOK1) or, for the id 0, a new one (the legacy BTN_NoviDOK): of the magacin and the vrsta (the first ones of
// the screen without them; the vrste of the tip of the magacin, the legacy PuniVRD), the next broj of the vrsta in the
// magacin (NoviBrojEDOK) and the datum of today. A new one of a porudžbenica (vezaID, the legacy BTN_TraziDOK and
// PodaciHDR) takes its magacin, org. jedinica, mesto troška and dobavljač. A confirmed or posted prijemnica cannot be
// changed: a new one is given with the message why.
func (s *PrijemnicaResource) GetPrijemnica(ctx context.Context, rdokID int64, vrd, magaciniID int, vezaID int64) (domain.KomPrijDokument, error) {
	var dok domain.KomPrijDokument
	var err error
	if dok.Magacini, err = s.GetMagacini(ctx); err != nil {
		return dok, err
	}
	if rdokID > 0 {
		p, found, err := s.repo.KomPrijemnica(ctx, rdokID)
		if err != nil {
			return dok, err
		}
		switch {
		case !found:
			dok.Poruka = komMsgNePostoji
		case strings.EqualFold(strings.TrimSpace(p.Knjige1), "D"):
			dok.Poruka, vrd, magaciniID = komMsgPorKnjizenIzmena, p.Vrd, p.MagaciniID
		case p.Doksts != domain.KomZahtevNepotvrdjen:
			dok.Poruka, vrd, magaciniID = komMsgPrijNijeIzmena, p.Vrd, p.MagaciniID
		default:
			return s.sacuvana(ctx, dok, p)
		}
	}
	h := &dok.Header
	h.Datum = s.now()
	var veza domain.KomPrijVeza
	if vezaID > 0 {
		var found bool
		if veza, found, err = s.repo.KomPrijVeza(ctx, vezaID); err != nil {
			return dok, err
		}
		if found {
			magaciniID = veza.MagaciniID
		} else {
			dok.Poruka = komMsgPrijPorudzbenica
		}
	}
	h.MagaciniID = komIzListe(dok.Magacini, magaciniID)
	if veza.RdokID > 0 && veza.MagaciniID == h.MagaciniID {
		h.VrdokID, h.VezaDokum, h.VezaDatum, h.Fkto, h.Fana = veza.RdokID, veza.Dokum, veza.Dadok, veza.Fkto, veza.Fana
		h.OrgjedText, h.MestoTroskaText = veza.OrgjedText, veza.MtText
		if h.Dobavljac, err = s.dobavljac(ctx, h.Fkto, h.Fana); err != nil {
			return dok, err
		}
	}
	if h.MagaciniID > 0 {
		if dok.Vrste, err = s.repo.KomPrijVrste(ctx, h.MagaciniID); err != nil {
			return dok, err
		}
		if dok.Konta, err = s.repo.KomPrijKonta(ctx, h.MagaciniID); err != nil {
			return dok, err
		}
	}
	h.Vrd = komIzListe(dok.Vrste, vrd)
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
	dok.NoviRbr = 1
	return dok, nil
}

// dobavljac returns the dobavljač of the konto and šifra with the text of its sistem PDV (the legacy PodaciHDR); empty
// without it.
func (s *PrijemnicaResource) dobavljac(ctx context.Context, konto, sifra string) (domain.KomDobavljacPodaci, error) {
	if strings.TrimSpace(konto) == "" || strings.TrimSpace(sifra) == "" {
		return domain.KomDobavljacPodaci{}, nil
	}
	d, found, err := s.repo.KomPorDobavljac(ctx, konto, sifra)
	if err != nil || !found {
		return domain.KomDobavljacPodaci{}, err
	}
	d.SistemPdv = komSistemPdv(d.TipPdv)
	return d, nil
}

// sacuvana is the content of the entry dialog with a saved nepotvrđena prijemnica: its header with the dobavljač, the
// vrste and the konta of its magacin, the stavke of its porudžbenica, the next redni broj and whether it has stavke.
func (s *PrijemnicaResource) sacuvana(ctx context.Context, dok domain.KomPrijDokument, p domain.KomPrijemnicaRow) (domain.KomPrijDokument, error) {
	h := domain.KomPrijHeaderView{RdokID: p.RdokID, Vrd: p.Vrd, Dokum: p.Dokum, MagaciniID: p.MagaciniID, Fkto: p.Fkto, Fana: p.Fana,
		OrgjedText: p.OrgjedText, MestoTroskaText: p.MestoTroskaText, VrdokID: p.VrdokID, VezaDokum: p.VezaDokum, VezaDatum: p.VezaDatum,
		Storno: p.Predznak == "-"}
	if d := komPrijDatum(p); d.Valid {
		h.Datum = d.Time
	}
	var err error
	if h.Dobavljac, err = s.dobavljac(ctx, p.Fkto, p.Fana); err != nil {
		return dok, err
	}
	dok.Header = h
	if dok.Vrste, err = s.repo.KomPrijVrste(ctx, p.MagaciniID); err != nil {
		return dok, err
	}
	if dok.Konta, err = s.repo.KomPrijKonta(ctx, p.MagaciniID); err != nil {
		return dok, err
	}
	if p.VrdokID > 0 {
		if dok.Izvori, err = s.repo.KomPorStavke(ctx, p.VrdokID); err != nil {
			return dok, err
		}
	}
	if dok.NoviRbr, err = s.repo.KomZahtevNoviRbr(ctx, nil, p.RdokID); err != nil {
		return dok, err
	}
	dok.ImaStavke = dok.NoviRbr > 1
	return dok, nil
}

// ValidatePrijemnica validates the header of a prijemnica (the legacy ValidacijaEDOK): a saved one is nepotvrđena; the
// magacin of a new one is one of the user, its vrsta one of the prijemnice of the magacin and its porudžbenica a
// potvrđena one of the magacin; the datum is a valid date not before the datum of the porudžbenica.
func (s *PrijemnicaResource) ValidatePrijemnica(ctx context.Context, params domain.KomPrijParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	var datumPorudzbenice sql.NullTime
	if params.RdokID > 0 {
		p, found, err := s.repo.KomPrijemnica(ctx, params.RdokID)
		switch {
		case err != nil:
			add("datum", common.ErrMsgReadData)
		case !found:
			add("datum", komMsgNePostoji)
		case p.Doksts != domain.KomZahtevNepotvrdjen:
			add("datum", komMsgPrijNijeIzmena)
		default:
			datumPorudzbenice = p.VezaDatum
		}
	} else {
		magaciniID := common.StringToInt(params.MagaciniID)
		if magacini, err := s.GetMagacini(ctx); err != nil {
			add("magaciniid", common.ErrMsgReadData)
		} else if !komUListi(magacini, magaciniID) {
			add("magaciniid", komMsgMagacin)
		}
		if vrste, err := s.repo.KomPrijVrste(ctx, magaciniID); err != nil {
			add("vrd", common.ErrMsgReadData)
		} else if !komUListi(vrste, common.StringToInt(params.Vrd)) {
			add("vrd", komMsgVrsta)
		}
		if id, given, ok := komID(params.VrdokID); !ok || !given {
			add("vrdokid", komMsgPrijPorudzbenica)
		} else if v, found, err := s.repo.KomPrijVeza(ctx, id); err != nil {
			add("vrdokid", common.ErrMsgReadData)
		} else if !found || v.MagaciniID != magaciniID {
			add("vrdokid", komMsgPrijPorudzbenica)
		} else {
			datumPorudzbenice = v.Dadok
		}
	}
	if datum, ok := komDatum(params.Datum); !ok {
		add("datum", komMsgPrijDatum)
	} else if datumPorudzbenice.Valid && datum.Format(common.HtmlLayout) < datumPorudzbenice.Time.Format(common.HtmlLayout) {
		add("datum", komMsgPrijDatumPor)
	}
	return errs
}

// SavePrijemnica saves the header of a validated prijemnica (the legacy SnimiEDOK) in one transaction: a new one with
// the next broj of its vrsta in its magacin (the broj prijemnice), the oznaka of its magacin and the data of its
// porudžbenica; a saved one with its datum. It returns the rdokid.
func (s *PrijemnicaResource) SavePrijemnica(ctx context.Context, params domain.KomPrijParams) (int64, error) {
	datum, _ := komDatum(params.Datum)
	p := domain.KomPrijemnicaUpis{RdokID: params.RdokID, Vrd: common.StringToInt(params.Vrd), MagaciniID: common.StringToInt(params.MagaciniID),
		Datum: datum}
	if id, given, ok := komID(params.VrdokID); ok && given {
		p.VrdokID = id
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	if p.RdokID > 0 {
		if err := s.repo.KomPrijIzmeni(ctx, tx, p); err != nil {
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
		if p.RdokID, err = s.repo.KomPrijDodaj(ctx, tx, p); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit prijemnica: %w", err)
	}
	return p.RdokID, nil
}

// DeletePrijemnica deletes a prijemnica with its stavke and their pakovanja (the legacy BTN_Brisanje): not a posted
// (knjige 1 "D") or confirmed one, whose message is returned instead.
func (s *PrijemnicaResource) DeletePrijemnica(ctx context.Context, rdokID int64) (string, error) {
	p, found, err := s.repo.KomPrijemnica(ctx, rdokID)
	if err != nil {
		return "", err
	}
	switch {
	case !found:
		return "Niste selektovali dokument za brisanje!!!!", nil
	case strings.EqualFold(strings.TrimSpace(p.Knjige1), "D"):
		return komMsgProknjizen, nil
	case p.Doksts != domain.KomZahtevNepotvrdjen:
		return komMsgPorPotvrdjenBris, nil
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomPrijBrisi(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje prijemnice: %w", err)
	}
	return "", nil
}

// ProveriPotvrdu checks a prijemnica before the question of its potvrda (the legacy BTN_NoviDOK1 "Potvrda dok."): it
// has stavke, it is nepotvrđena, every stavka has a serija and a primljena količina and its konto is one of the
// magacin; the message why it cannot be confirmed is returned.
func (s *PrijemnicaResource) ProveriPotvrdu(ctx context.Context, rdokID int64) (string, error) {
	_, _, poruka, err := s.zaPotvrdu(ctx, rdokID)
	return poruka, err
}

// zaPotvrdu returns the prijemnica and its stavke after the checks of ProveriPotvrdu, or the message why it cannot be
// confirmed.
func (s *PrijemnicaResource) zaPotvrdu(ctx context.Context, rdokID int64) (domain.KomPrijemnicaRow, []domain.KomPrijStavkaRow, string, error) {
	p, found, err := s.repo.KomPrijemnica(ctx, rdokID)
	if err != nil {
		return p, nil, "", err
	}
	if !found {
		return p, nil, komMsgBezStavkiPotvrda, nil
	}
	stavke, err := s.repo.KomPrijStavke(ctx, rdokID)
	if err != nil {
		return p, nil, "", err
	}
	if len(stavke) == 0 {
		return p, nil, komMsgBezStavkiPotvrda, nil
	}
	if p.Doksts != domain.KomZahtevNepotvrdjen {
		return p, nil, komMsgVecPotvrdjen, nil
	}
	for _, st := range stavke {
		if strings.TrimSpace(st.Serija) == "" || st.Kolic == 0 {
			return p, nil, komMsgPrijNekompletne, nil
		}
	}
	konta, err := s.repo.KomPrijKonta(ctx, p.MagaciniID)
	if err != nil {
		return p, nil, "", err
	}
	for _, st := range stavke {
		if !komPrijKontoUListi(konta, st.Konto) {
			return p, nil, komMsgPrijBezKonta, nil
		}
	}
	return p, stavke, "", nil
}

// komPrijKontoUListi tells whether the konto is one of the konta of the magacin.
func komPrijKontoUListi(konta []domain.ComboItem, konto string) bool {
	konto = strings.TrimSpace(konto)
	for _, k := range konta {
		if konto != "" && k.Key == konto {
			return true
		}
	}
	return false
}

// PotvrdiPrijemnicu confirms a prijemnica (the legacy "Da" of "Potvrda dok.", PotvrdaDOK) after the checks of
// ProveriPotvrdu, in one transaction: its status 40, or 30 in a magacin with the analiza, and every stavka booked into
// the stanje by konto (AzurRSTA_ZN, on the datum of the prijemnica at its cena) and by lot (AzurDRSTA_ZN, on the datum
// prijemnice by its OTK, serija and rok trajanja), for the zalihe or, with the analiza, the neraspoložive zalihe; the
// količina and the vrednost with the sign of the predznak of the vrsta. It returns the message of the checks, or
// "Dokument je potvrđen!!!".
func (s *PrijemnicaResource) PotvrdiPrijemnicu(ctx context.Context, rdokID int64) (string, error) {
	p, stavke, poruka, err := s.zaPotvrdu(ctx, rdokID)
	if err != nil || poruka != "" {
		return poruka, err
	}
	zaliha, doksts := p.Analiza != 1, domain.KomZahtevPotvrdjen
	if !zaliha {
		doksts = domain.KomPrijemnicaAnalizirana
	}
	storno := p.Predznak == "-"
	var danal, datotp time.Time
	if p.Dadok.Valid {
		danal = p.Dadok.Time
	}
	if d := komPrijDatum(p); d.Valid {
		datotp = d.Time
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	ok, err := s.repo.KomPrijPotvrdi(ctx, tx, rdokID, doksts)
	if err != nil {
		return "", err
	}
	if !ok {
		return komMsgVecPotvrdjen, nil
	}
	for _, st := range stavke {
		kolicina := komPrijKolicina(st.Kolic, storno)
		vrednost := komPrijKolicina(st.Kolic*st.Cena, storno)
		promena := domain.RobPromenaStanja{MagaciniID: p.MagaciniID, Konto: strings.TrimSpace(st.Konto), Sifra: st.Sifra, Kolicina: kolicina,
			Iznos: vrednost, Cena: st.Cena, Danal: danal, DokvrstaID: p.DokvrstaID, Tipdok: p.Tipdok}
		if err := s.stanje.AzurirajRSTAZN(ctx, tx, promena, zaliha); err != nil {
			return "", fmt.Errorf("stanje stavke %d: %w", st.Rbr, err)
		}
		promena.Cena, promena.Danal = 0, datotp
		promena.Otk, promena.Serija, promena.Roktr = st.Otk, st.Serija, st.Roktr
		if err := s.stanje.AzurirajDRSTAZN(ctx, tx, promena, zaliha); err != nil {
			return "", fmt.Errorf("zalihe stavke %d: %w", st.Rbr, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit potvrda prijemnice: %w", err)
	}
	return komMsgPotvrdjen, nil
}

// GetStavke returns the stavke of a prijemnica in the order of their redni broj.
func (s *PrijemnicaResource) GetStavke(ctx context.Context, rdokID int64) ([]domain.KomPrijStavkaRow, error) {
	if rdokID <= 0 {
		return nil, nil
	}
	return s.repo.KomPrijStavke(ctx, rdokID)
}

// GetArtikal returns the artikal of a stavka entered by hand with the OTK of its next prijemnica of the vrsta of the
// prijemnica (the legacy exit of EDT_SIFRA and VratiOTK); found is false without it.
func (s *PrijemnicaResource) GetArtikal(ctx context.Context, rdokID int64, sifra string) (domain.KomPrijArtikal, bool, error) {
	n, err := strconv.Atoi(strings.TrimSpace(sifra))
	if err != nil || n <= 0 {
		return domain.KomPrijArtikal{}, false, nil
	}
	a, found, err := s.repo.KomPrijArtikal(ctx, n)
	if err != nil || !found {
		return a, found, err
	}
	p, found, err := s.repo.KomPrijemnica(ctx, rdokID)
	if err != nil || !found {
		return a, true, err
	}
	broj, err := s.repo.KomPrijOtkBroj(ctx, nil, p.Vrd, a.Sifra)
	if err != nil {
		return a, true, err
	}
	a.Otk = komPrijOtk(broj)
	return a, true, nil
}

// prijStavka returns the stavka of the id of a prijemnica; found is false without it.
func (s *PrijemnicaResource) prijStavka(ctx context.Context, rdokID, rproID int64) (domain.KomPrijStavkaRow, bool, error) {
	stavke, err := s.repo.KomPrijStavke(ctx, rdokID)
	if err != nil {
		return domain.KomPrijStavkaRow{}, false, err
	}
	for _, st := range stavke {
		if st.RproID == rproID {
			return st, true, nil
		}
	}
	return domain.KomPrijStavkaRow{}, false, nil
}

// nepotvrdjena returns the prijemnica of the id when it is nepotvrđena, else the message why its stavke cannot change.
func (s *PrijemnicaResource) nepotvrdjena(ctx context.Context, rdokID int64) (domain.KomPrijemnicaRow, string, error) {
	p, found, err := s.repo.KomPrijemnica(ctx, rdokID)
	switch {
	case err != nil:
		return p, "", err
	case !found:
		return p, komMsgNePostoji, nil
	case p.Doksts != domain.KomZahtevNepotvrdjen:
		return p, komMsgPrijNijeIzmena, nil
	}
	return p, "", nil
}

// ValidateStavka validates a stavka of a prijemnica (the legacy ValidacijaEPRO): its prijemnica is nepotvrđena, a new
// stavka has an artikal of the šifra and a changed one (rproid) exists; the konto is one of the magacin, the količina
// na dokumentu and the primljena količina are greater than 0 and the serija is given.
func (s *PrijemnicaResource) ValidateStavka(ctx context.Context, params domain.KomPrijStavkaParams) []domain.FieldError {
	var errs []domain.FieldError
	add := func(field, message string) {
		errs = append(errs, domain.FieldError{Field: field, ErrorMessage: message})
	}
	p, poruka, err := s.nepotvrdjena(ctx, params.RdokID)
	switch {
	case err != nil:
		add("sifraartikla", common.ErrMsgReadData)
	case poruka != "":
		add("sifraartikla", poruka)
	}
	if params.RproID > 0 {
		if _, found, err := s.prijStavka(ctx, params.RdokID, params.RproID); err != nil {
			add("sifraartikla", common.ErrMsgReadData)
		} else if !found {
			add("sifraartikla", komMsgNePostoji)
		}
	} else if n, err := strconv.Atoi(strings.TrimSpace(params.Sifra)); err != nil || n <= 0 {
		add("sifraartikla", komMsgArtikal)
	} else if _, found, err := s.repo.KomPrijArtikal(ctx, n); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !found {
		add("sifraartikla", komMsgArtikal)
	}
	if p.MagaciniID > 0 {
		if konta, err := s.repo.KomPrijKonta(ctx, p.MagaciniID); err != nil {
			add("konto", common.ErrMsgReadData)
		} else if !komPrijKontoUListi(konta, params.Konto) {
			add("konto", komMsgPrijKonto)
		}
	}
	if common.StringToFloat64(params.TrazKolic) <= 0 {
		add("trazkolic", komMsgPrijTrazKolic)
	}
	if common.StringToFloat64(params.Kolic) <= 0 {
		add("kolic", komMsgPrijKolic)
	}
	if strings.TrimSpace(params.Serija) == "" {
		add("serija", komMsgPrijSerija)
	}
	return errs
}

// SaveStavka saves a validated stavka of a prijemnica in one transaction: a new one entered by hand with the next redni
// broj (the legacy AddEPRO: the cena 0, the given OTK or the one of VratiOTK, the magacin and the vrsta, broj and datum
// of the prijemnica), a changed one (rproid) with its konto, količine, serija and OTK (the legacy change of a row of
// the grid); the primljena količina with the sign of the predznak of the vrsta.
func (s *PrijemnicaResource) SaveStavka(ctx context.Context, params domain.KomPrijStavkaParams) error {
	p, found, err := s.repo.KomPrijemnica(ctx, params.RdokID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("prijemnica %d ne postoji", params.RdokID)
	}
	trazKolic := math.Abs(common.StringToFloat64(params.TrazKolic))
	kolic := komPrijKolicina(common.StringToFloat64(params.Kolic), p.Predznak == "-")
	konto, serija, otk := strings.TrimSpace(params.Konto), strings.TrimSpace(params.Serija), strings.TrimSpace(params.Otk)
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if params.RproID > 0 {
		if err := s.repo.KomPrijStavkaIzmeni(ctx, tx, domain.KomPrijStavkaIzmena{RdokID: p.RdokID, RproID: params.RproID, Konto: konto,
			TrazKolic: trazKolic, Kolic: kolic, Serija: serija, Otk: otk}); err != nil {
			return err
		}
	} else {
		n, _ := strconv.Atoi(strings.TrimSpace(params.Sifra))
		a, found, err := s.repo.KomPrijArtikal(ctx, n)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("artikal %s ne postoji", params.Sifra)
		}
		if otk == "" {
			broj, err := s.repo.KomPrijOtkBroj(ctx, tx, p.Vrd, a.Sifra)
			if err != nil {
				return err
			}
			otk = komPrijOtk(broj)
		}
		rbr, err := s.repo.KomZahtevNoviRbr(ctx, tx, p.RdokID)
		if err != nil {
			return err
		}
		st := domain.KomPrijStavkaUpis{RdokID: p.RdokID, Rbr: rbr, Sifra: a.Sifra, RsifID: a.RsifID, Po: a.Po, Konto: konto, Mag: p.Mag,
			MagaciniID: p.MagaciniID, TrazKolic: trazKolic, Kolic: kolic, Otk: otk, Serija: serija, Vrd: p.Vrd, Dokum: p.Dokum,
			Dadok: komPrijDatum(p)}
		if err := s.repo.KomPrijStavkaDodaj(ctx, tx, st); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stavka prijemnice: %w", err)
	}
	return nil
}

// DeleteStavka deletes a stavka of a nepotvrđena prijemnica with its pakovanje (the legacy BTN_BrisiDET): the redni
// brojevi after it move one back; the message why is returned for a confirmed prijemnica.
func (s *PrijemnicaResource) DeleteStavka(ctx context.Context, rdokID, rproID int64) (string, error) {
	if _, poruka, err := s.nepotvrdjena(ctx, rdokID); err != nil || poruka != "" {
		return poruka, err
	}
	st, found, err := s.prijStavka(ctx, rdokID, rproID)
	if err != nil {
		return "", err
	}
	if !found {
		return komMsgNePostoji, nil
	}
	return "", s.brisiStavku(ctx, rdokID, st)
}

// brisiStavku deletes a stavka with its pakovanje in one transaction (the redni brojevi after it moved back).
func (s *PrijemnicaResource) brisiStavku(ctx context.Context, rdokID int64, st domain.KomPrijStavkaRow) error {
	tx, err := s.repo.BeginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomPrijStavkaBrisi(ctx, tx, rdokID, st.RproID, st.Rbr); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit brisanje stavke prijemnice: %w", err)
	}
	return nil
}

// GetStavkaForma returns a saved stavka for the form of the stavke ("Izmeni" of its row) with its količine without the
// sign of a storno; found is false without it.
func (s *PrijemnicaResource) GetStavkaForma(ctx context.Context, rdokID, rproID int64) (domain.KomPrijStavkaForma, bool, error) {
	st, found, err := s.prijStavka(ctx, rdokID, rproID)
	if err != nil || !found {
		return domain.KomPrijStavkaForma{}, found, err
	}
	return domain.KomPrijStavkaForma{RproID: st.RproID, Rbr: st.Rbr, Konto: strings.TrimSpace(st.Konto), Sifra: st.Sifra, Naziv: st.Naziv, JM: st.JM,
		TrazKolic: math.Abs(st.TrazKolic), Kolic: math.Abs(st.Kolic), Serija: st.Serija, Otk: st.Otk}, true, nil
}

// komPrijCena is the cena of a stavka taken from the porudžbenica: its cena less the rabat rounded to 2 decimals (the
// legacy AddEPRONAR).
func komPrijCena(cena, rabat float64) float64 {
	return cena - math.Round(cena*rabat/100*100)/100
}

// Preuzmi takes the stavke of the porudžbenica into a saved nepotvrđena prijemnica (the legacy BTN_Dodaj with the id of
// a stavka, BTN_DodajSve with 0, AddEPRONAR) in one transaction: each one whose artikal is not in the prijemnica yet
// becomes a stavka with the next redni broj, the first konto of the magacin, its tražena količina as the količina na
// dokumentu, the primljena količina 0, the OTK of VratiOTK, no serija, its cena less the rabat and the iznos of the
// tražena količina. An artikal already in the prijemnica is skipped ("Ovakvi podaci već postoje"); the message is
// returned.
func (s *PrijemnicaResource) Preuzmi(ctx context.Context, rdokID, rproID int64) (string, error) {
	p, poruka, err := s.nepotvrdjena(ctx, rdokID)
	if err != nil || poruka != "" {
		return poruka, err
	}
	if p.VrdokID <= 0 {
		return komMsgPorNemaZaDodavanje, nil
	}
	izvori, err := s.repo.KomPorStavke(ctx, p.VrdokID)
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
	konta, err := s.repo.KomPrijKonta(ctx, p.MagaciniID)
	if err != nil {
		return "", err
	}
	var konto string
	if len(konta) > 0 {
		konto = konta[0].Key
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	dupli := false
	for _, id := range ids {
		iz, found, err := s.repo.KomPrijIzvor(ctx, tx, p.VrdokID, id)
		if err != nil {
			return "", err
		}
		if !found {
			continue
		}
		ima, err := s.repo.KomPrijImaArtikal(ctx, tx, p.RdokID, iz.RsifID)
		if err != nil {
			return "", err
		}
		if ima {
			dupli = true
			continue
		}
		broj, err := s.repo.KomPrijOtkBroj(ctx, tx, p.Vrd, iz.Sifra)
		if err != nil {
			return "", err
		}
		rbr, err := s.repo.KomZahtevNoviRbr(ctx, tx, p.RdokID)
		if err != nil {
			return "", err
		}
		cena := komPrijCena(iz.Cena, iz.Rabat)
		st := domain.KomPrijStavkaUpis{RdokID: p.RdokID, Rbr: rbr, Sifra: iz.Sifra, RsifID: iz.RsifID, Po: iz.Po, Konto: konto, Mag: p.Mag,
			MagaciniID: p.MagaciniID, TrazKolic: iz.TrazKolic, Otk: komPrijOtk(broj), Fcena: iz.Cena, Cena: cena, Iznos: iz.TrazKolic * cena,
			Rabat: iz.Rabat, Vrd: p.Vrd, Dokum: p.Dokum, Dadok: komPrijDatum(p)}
		if err := s.repo.KomPrijStavkaDodaj(ctx, tx, st); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit stavke prijemnice: %w", err)
	}
	if dupli {
		return komMsgPrijDupli, nil
	}
	return "", nil
}

// VratiPoslednju deletes the last stavka (the largest redni broj) of a nepotvrđena prijemnica with its pakovanje (the
// legacy BTN_VratiJedan); without stavke the message why is returned.
func (s *PrijemnicaResource) VratiPoslednju(ctx context.Context, rdokID int64) (string, error) {
	if _, poruka, err := s.nepotvrdjena(ctx, rdokID); err != nil || poruka != "" {
		return poruka, err
	}
	stavke, err := s.repo.KomPrijStavke(ctx, rdokID)
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

// VratiSve deletes all the stavke of a nepotvrđena prijemnica with their pakovanja (the legacy BTN_VratiSve) in one
// transaction; without stavke the message why is returned.
func (s *PrijemnicaResource) VratiSve(ctx context.Context, rdokID int64) (string, error) {
	if _, poruka, err := s.nepotvrdjena(ctx, rdokID); err != nil || poruka != "" {
		return poruka, err
	}
	stavke, err := s.repo.KomPrijStavke(ctx, rdokID)
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
	if err := s.repo.KomPrijStavkeBrisi(ctx, tx, rdokID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit brisanje stavki prijemnice: %w", err)
	}
	return "", nil
}

// GetPakovanja returns the stavke of a prijemnica with their pakovanja (the legacy BTN_Pakovanja and pr_Pakovanja);
// without stavke the message why is returned.
func (s *PrijemnicaResource) GetPakovanja(ctx context.Context, rdokID int64) ([]domain.KomPrijPakovanjeRow, string, error) {
	rows, err := s.repo.KomPrijPakovanja(ctx, rdokID)
	if err != nil {
		return nil, "", err
	}
	if len(rows) == 0 {
		return nil, komMsgPrijBezPakovanja, nil
	}
	return rows, "", nil
}

// SavePakovanje saves the pakovanje of a stavka of a nepotvrđena prijemnica (the količina po pakovanju and the broj
// pakovanja, not below 0) in one transaction; the message why is returned when it cannot be saved.
func (s *PrijemnicaResource) SavePakovanje(ctx context.Context, params domain.KomPrijPakovanjeParams) (string, error) {
	if _, poruka, err := s.nepotvrdjena(ctx, params.RdokID); err != nil || poruka != "" {
		return poruka, err
	}
	if _, found, err := s.prijStavka(ctx, params.RdokID, params.RproID); err != nil {
		return "", err
	} else if !found {
		return komMsgNePostoji, nil
	}
	kolPoPak, brPak := common.StringToFloat64(params.KolPoPak), common.StringToFloat64(params.BrPak)
	if kolPoPak < 0 || brPak < 0 {
		return komMsgPrijPakovanje, nil
	}
	tx, err := s.repo.BeginTx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.KomPrijPakovanjeSnimi(ctx, tx, params.RdokID, params.RproID, kolPoPak, brPak); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit pakovanje stavke: %w", err)
	}
	return "", nil
}

// GetPrikaz returns the prijemnice of the list of the main screen (the legacy PR_QRY_EDOK): the nepotvrđene or the
// potvrđene ones of the magacin or of all the magacini of the user.
func (s *PrijemnicaResource) GetPrikaz(ctx context.Context, params domain.KomPrijPrikazParams) ([]domain.KomPrijPrikazRow, error) {
	if params.Doksts != domain.KomZahtevPotvrdjen {
		params.Doksts = domain.KomZahtevNepotvrdjen
	}
	if !params.SviMagacini && params.MagaciniID <= 0 {
		return nil, nil
	}
	return s.repo.KomPrijPrikaz(ctx, params)
}

// GetStampa returns the print of a prijemnica (the legacy PR_RPT_PRIJEMNICA): the firm with its tekući računi, the
// dobavljač, the broj and the datum of the prijemnica, the broj and the datum of its porudžbenica, the magacin, the
// org. jedinica and the stavke; a nepotvrđena one is marked; one without stavke is not printed (the message why is
// returned).
func (s *PrijemnicaResource) GetStampa(ctx context.Context, rdokID int64) (domain.KomPrijStampa, string, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.KomPrijStampa{}, "", fmt.Errorf("no user session found")
	}
	p, found, err := s.repo.KomPrijemnica(ctx, rdokID)
	if err != nil {
		return domain.KomPrijStampa{}, "", err
	}
	if !found {
		return domain.KomPrijStampa{}, komMsgBezStavkiStampa, nil
	}
	stampa := domain.KomPrijStampa{Dokum: p.Dokum, Datum: komPrijDatum(p), VezaDokum: p.VezaDokum, VezaDatum: p.VezaDatum, Magacin: p.MagacinText,
		MagacinAdresa: p.MagacinAdresa, OJ: p.OrgjedText, Nepotvrdjen: p.Doksts == domain.KomZahtevNepotvrdjen}
	if stampa.Stavke, err = s.repo.KomPrijStavke(ctx, rdokID); err != nil {
		return domain.KomPrijStampa{}, "", err
	}
	if len(stampa.Stavke) == 0 {
		return domain.KomPrijStampa{}, komMsgBezStavkiStampa, nil
	}
	if idFkpl, err := s.repo.KomPorIDFkpl(ctx, p.Fkto, p.Fana); err != nil {
		return domain.KomPrijStampa{}, "", err
	} else if idFkpl > 0 {
		if stampa.Dobavljac, _, err = s.repo.KomPonudaStampaDobavljac(ctx, idFkpl); err != nil {
			return domain.KomPrijStampa{}, "", err
		}
	}
	if stampa.Firma, err = s.repo.KomFirma(ctx, session.Firma); err != nil {
		return domain.KomPrijStampa{}, "", err
	}
	if stampa.ZiroRacuni, err = s.repo.BankeZiroRacuni(ctx); err != nil {
		return domain.KomPrijStampa{}, "", err
	}
	return stampa, "", nil
}

// GetKarton returns the karton identifikacije of a stavka of a prijemnica (the legacy BTN_KI with RPT_KI): the
// dobavljač, the broj and the datum of the prijemnica, the artikal, its serija, OTK and primljena količina; without a
// selected stavka the message why is returned.
func (s *PrijemnicaResource) GetKarton(ctx context.Context, rdokID, rproID int64) (domain.KomPrijKarton, string, error) {
	if rproID <= 0 {
		return domain.KomPrijKarton{}, komMsgPrijKarton, nil
	}
	p, found, err := s.repo.KomPrijemnica(ctx, rdokID)
	if err != nil {
		return domain.KomPrijKarton{}, "", err
	}
	if !found {
		return domain.KomPrijKarton{}, komMsgPrijKarton, nil
	}
	st, found, err := s.prijStavka(ctx, rdokID, rproID)
	if err != nil {
		return domain.KomPrijKarton{}, "", err
	}
	if !found {
		return domain.KomPrijKarton{}, komMsgPrijKarton, nil
	}
	d, err := s.dobavljac(ctx, p.Fkto, p.Fana)
	if err != nil {
		return domain.KomPrijKarton{}, "", err
	}
	return domain.KomPrijKarton{Dobavljac: d.Naziv, Dokum: p.Dokum, Datum: komPrijDatum(p), Sifra: st.Sifra, Naziv: st.Naziv, JM: st.JM,
		Serija: st.Serija, Otk: st.Otk, Kolicina: st.Kolic, Decimale: st.Decimale}, "", nil
}

// StavkeTableFields are the columns of the grid of the stavke of the prijemnica (the legacy TABLE_PR_QRY_EPRO1).
func (s *PrijemnicaResource) StavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "6", TextAlign: "right"},
		{Name: "konto", Label: "Konto", Width: "7"},
		{Name: "sifra", Label: "Šifra artikla", Width: "8", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla / sirovine", Width: "32"},
		{Name: "jm", Label: "JM", Width: "5"},
		{Name: "trazkolic", Label: "Količina na dokumentu", Width: "11", TextAlign: "right"},
		{Name: "kolic", Label: "Primljena količina", Width: "11", TextAlign: "right"},
		{Name: "serija", Label: "Serija", Width: "10"},
		{Name: "otk", Label: "OTK", Width: "8"},
	}
}

// IzvoriTableFields are the columns of the stavke of the porudžbenica (the legacy "Količine po porudžbenici",
// TABLE_PR_QRY_EPRO).
func (s *PrijemnicaResource) IzvoriTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "8", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "12", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla / sirovine", Width: "46"},
		{Name: "jm", Label: "JM", Width: "8"},
		{Name: "kolicina", Label: "Tražena količina", Width: "13", TextAlign: "right"},
		{Name: "cena", Label: "Jedinična cena", Width: "13", TextAlign: "right"},
	}
}

// PrikazTableFields are the columns of the list of the main screen (the legacy TABLE_QRY_EDOK).
func (s *PrijemnicaResource) PrikazTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "magacin", Label: "Magacin", Width: "12"},
		{Name: "vrd", Label: "Vrsta dokum.", Width: "6", TextAlign: "right"},
		{Name: "dokum", Label: "Broj dokumenta", Width: "7", TextAlign: "right"},
		{Name: "dadok", Label: "Datum dokumenta", Width: "8"},
		{Name: "fkto", Label: "Konto partnera", Width: "7"},
		{Name: "fana", Label: "Šifra", Width: "6"},
		{Name: "dobavljac", Label: "Naziv partnera", Width: "18"},
		{Name: "opis", Label: "Opis", Width: "12"},
		{Name: "rok", Label: "Rok plaćanja", Width: "5", TextAlign: "right"},
		{Name: "datiz", Label: "Datum izvornog dokum.", Width: "8"},
		{Name: "otp", Label: "Mesto isporuke", Width: "8"},
		{Name: "kamt", Label: "Uslovi kamate", Width: "8"},
	}
}

// PrikazStavkeTableFields are the columns of the stavke of the selected prijemnica of the main screen (the legacy
// "Detalji prijemnice", TABLE_EPRO2).
func (s *PrijemnicaResource) PrikazStavkeTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "6", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "9", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla / sirovine", Width: "34"},
		{Name: "jm", Label: "JM", Width: "5"},
		{Name: "trazkolic", Label: "Količina na dokumentu", Width: "11", TextAlign: "right"},
		{Name: "kolic", Label: "Primljena količina", Width: "11", TextAlign: "right"},
		{Name: "otk", Label: "OTK", Width: "8"},
		{Name: "serija", Label: "Serija", Width: "10"},
		{Name: "roktr", Label: "Rok trajanja", Width: "8"},
	}
}

// PakovanjaTableFields are the columns of the dialog "Pakovanje" (the legacy pr_Pakovanja "Količine po pakovanju").
func (s *PrijemnicaResource) PakovanjaTableFields() []domain.Fields {
	return []domain.Fields{
		{Name: "rbr", Label: "Red. br.", Width: "7", TextAlign: "right"},
		{Name: "sifra", Label: "Šifra artikla", Width: "10", TextAlign: "right"},
		{Name: "naziv", Label: "Naziv artikla / sirovine", Width: "38"},
		{Name: "jm", Label: "JM", Width: "6"},
		{Name: "kolpopak", Label: "Količina po pakovanju", Width: "13", TextAlign: "right"},
		{Name: "brpak", Label: "Broj pakovanja", Width: "12", TextAlign: "right"},
		{Name: "uskladisteno", Label: "Uskladištena količina", Width: "14", TextAlign: "right"},
	}
}

// komPrijKolicinaText is a količina of a grid with the decimals of its JM (at most 3).
func komPrijKolicinaText(kolicina float64, decimale int) string {
	if decimale < 0 || decimale > 3 {
		decimale = 3
	}
	return common.FormatNumberWithSystemLocale(kolicina, decimale)
}

// PrijStavkeTableRows are the rows of the stavke of a prijemnica, the last ones first like the legacy grid
// (PuniDETALJ1 "ORDER BY RBR DESC"), with "Izmeni" and "Obriši" of every row (izmena).
func PrijStavkeTableRows(stavke []domain.KomPrijStavkaRow, izmena bool) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for i := len(stavke) - 1; i >= 0; i-- {
		st := stavke[i]
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), HasUpdate: izmena, HasDelete: izmena, Fields: []string{
			strconv.Itoa(st.Rbr), strings.TrimSpace(st.Konto), strconv.Itoa(st.Sifra), st.Naziv, st.JM, komPrijKolicinaText(st.TrazKolic, st.Decimale),
			komPrijKolicinaText(st.Kolic, st.Decimale), st.Serija, st.Otk,
		}})
	}
	return rows
}

// PrijIzvoriTableRows are the rows of the stavke of the porudžbenica of a prijemnica.
func PrijIzvoriTableRows(izvori []domain.KomPorStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(izvori))
	for _, iz := range izvori {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(iz.RproID, 10), Fields: []string{
			strconv.Itoa(iz.Rbr), strconv.Itoa(iz.Sifra), iz.Naziv, iz.JM, komPrijKolicinaText(iz.Kolicina, iz.Decimale),
			common.FormatNumberWithSystemLocale(iz.Cena, 2),
		}})
	}
	return rows
}

// PrijPrikazTableRows are the rows of the list of the main screen.
func PrijPrikazTableRows(prijemnice []domain.KomPrijPrikazRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(prijemnice))
	for _, p := range prijemnice {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(p.RdokID, 10), Fields: []string{
			p.Magacin, strconv.Itoa(p.Vrd), strconv.Itoa(p.Dokum), komDatumText(p.Dadok), p.Fkto, p.Fana, p.Dobavljac, p.Opis, strconv.Itoa(p.Rok),
			komDatumText(p.Datiz), p.Otp, p.Kamt,
		}})
	}
	return rows
}

// PrijPrikazStavkeTableRows are the rows of the stavke of the selected prijemnica of the main screen, in the order of
// their redni broj (the rok trajanja, a WinDev integer date, as dd.mm.yyyy).
func PrijPrikazStavkeTableRows(stavke []domain.KomPrijStavkaRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(stavke))
	for _, st := range stavke {
		roktr := komWinDevDatumText(st.Roktr)
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(st.RproID, 10), Fields: []string{
			strconv.Itoa(st.Rbr), strconv.Itoa(st.Sifra), st.Naziv, st.JM, komPrijKolicinaText(st.TrazKolic, st.Decimale),
			komPrijKolicinaText(st.Kolic, st.Decimale), st.Otk, st.Serija, roktr,
		}})
	}
	return rows
}

// PrijPakovanjaTableRows are the rows of the dialog "Pakovanje": the količina po pakovanju, the broj pakovanja and the
// uskladištena količina of every stavka.
func PrijPakovanjaTableRows(pakovanja []domain.KomPrijPakovanjeRow) []domain.TableRow {
	rows := make([]domain.TableRow, 0, len(pakovanja))
	for _, p := range pakovanja {
		rows = append(rows, domain.TableRow{ID: strconv.FormatInt(p.RproID, 10), Fields: []string{
			strconv.Itoa(p.Rbr), strconv.Itoa(p.Sifra), p.Naziv, p.JM, common.FormatNumberWithSystemLocale(p.KolPoPak, 3),
			common.FormatNumberWithSystemLocale(p.BrPak, 0), common.FormatNumberWithSystemLocale(p.Uskladisteno, 3),
		}})
	}
	return rows
}
