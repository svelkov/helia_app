package robno

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
)

// ---------------------------------------------------------------------------------------------------
// Popis u toku godine (the group PTG, the legacy rob_opsti_dokPOPIS): the roba counted in the magacin of the
// nalog, with the radnik (a konto and its analitička šifra of the kontni plan, rdok.pkto and rdok.pana). The
// screen is the one of the opšti dokumenti (Ptg) with the stavke of the opšti dokument (the konto, the artikal,
// the količina, the nabavna cena, the cena and the lot). A stavka changes the nalog and the document only (its
// prodajna and nabavna vrednost and its porez), not the stanje.
// ---------------------------------------------------------------------------------------------------

// robPtgDokument fills the radnik of a popis u toku godine (rdok.pkto with the naziv of its konto, rdok.pana with
// the naziv of its šifra: the legacy EDT_NAZKONTO and EDT_NAZSIFRE) and the label of the cena of its stavke by the
// tip of the magacin (the legacy caption of EDT_CENA: the veleprodaja the veleprodajna cena, the others the
// magacinska cena); the cena and the nabavna cena of the stavke are entered.
func (s *RobnoDokumentaResource) robPtgDokument(ctx context.Context, res *domain.RobnoOpstiDokument) error {
	h := &res.Header
	if h.RdokID > 0 {
		var err error
		if h.Pkto, h.Pana, err = s.robno.RdokPktoPana(ctx, h.RdokID); err != nil {
			return err
		}
		if h.Pkto != "" {
			if h.PktoNaziv, err = s.robOpstiNazivKonta(ctx, h.Pkto, ""); err != nil {
				return err
			}
			if h.Pana != "" {
				if h.GradilisteNaziv, err = s.robOpstiNazivKonta(ctx, h.Pkto, h.Pana); err != nil {
					return err
				}
			}
		}
	}
	st := &res.Stavka
	st.CenaNaziv, st.CenaZakljucana, st.NabCena = "Magacinska cena", false, false
	tipmag, err := s.robno.MagacinTipmag(ctx, h.MagaciniID)
	if err != nil {
		return err
	}
	if tipmag == "V" {
		st.CenaNaziv = "Veleprodajna cena"
	}
	return nil
}

// robPtgArtikal returns the artikal of a stavka of a popis u toku godine (the legacy exit of EDT_SIFRA of
// rob_opsti_dokPOPIS): the artikal by its šifra with its naziv and JM (the konto of the form) and its stanje in
// the magacin: in a magacin of the tip zaliha 2 the one of the lot of the form (drsta; the artikal must be of a tip
// of the magacin, else Poruka), else of the konto (rsta). The cena and the nabavna cena are the ones of the cene of
// the artikal (rcene; 0 without them: the legacy reads the ones of the stanje first and then always the ones of the
// cene). found is false for an artikal that does not exist.
func (s *RobnoDokumentaResource) robPtgArtikal(ctx context.Context, k domain.RobStavkeKontekst, sifra, konto, otk, serija, rok string) (domain.RobOpstiArtikal, bool, error) {
	a := domain.RobOpstiArtikal{}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return a, false, fmt.Errorf("no user session found")
	}
	rows, err := s.robno.RsifArtikalSaCenama(ctx, sifra)
	if err != nil {
		return a, false, err
	}
	if rows == nil || len(*rows) == 0 {
		if k.Tipzal == 2 {
			a.Poruka = "Artikal sa izabranom šifrom ne postoji."
		}
		return a, false, nil
	}
	r := (*rows)[0]
	if k.Tipzal == 2 {
		tipovi, opis := robOpstiTipArtikla(k.Tipmag)
		if !robSadrzi(tipovi, r.C3) {
			a.Poruka = "Izabrani tip artikla ne pripada " + opis + "."
			return a, false, nil
		}
	}
	a.Sifra, a.Naziv, a.Jm, a.Konto = common.StringToInt(r.Key), r.C1, r.C2, strings.TrimSpace(konto)
	a.Cena, a.Ncena = common.StringToFloat64(r.C4), common.StringToFloat64(r.C5)
	if k.Tipzal == 2 {
		datum, _ := robnoUnosDatum(rok)
		stanje, _, err := s.robOpstiDrsta(ctx, k.Mag, a.Konto, a.Sifra, otk, serija, RobWinDevDatum(datum))
		if err != nil {
			return a, true, err
		}
		a.Stanje = stanje
		return a, true, nil
	}
	st, err := s.robOpstiStanje(ctx, k.MagaciniID, a.Konto, a.Sifra)
	if err != nil {
		return a, true, err
	}
	if st.found {
		a.Stanje = st.ulaz - st.izlaz
	}
	return a, true, nil
}

// robPtgValidateStavka validates a stavka of a popis u toku godine (the legacy Validacija of rob_opsti_dokPOPIS):
// the konto is required and a konto of the magacin; the šifra is required and the artikal exists; the količina
// and the cena are greater than 0. The question "Unos bez OTK broja?" is not asked, the check of the zalihe is
// commented out in the legacy.
func (s *RobnoDokumentaResource) robPtgValidateStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) []domain.FieldError {
	var fieldErrors []domain.FieldError
	add := func(field, message string) {
		fieldErrors = append(fieldErrors, domain.FieldError{Field: field, ErrorMessage: message})
	}
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		add("sifraartikla", common.ErrMsgSessionNotFound)
		return fieldErrors
	}
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	if konto == "" {
		add("kontostavke", common.ErrMsgObavezanPodatak)
	} else if konta, err := s.GetRobKontaMagacina(ctx, k.MagaciniID); err != nil {
		add("kontostavke", common.ErrMsgReadData)
	} else if !robSadrzi(konta, konto) {
		add("kontostavke", "Nije konto magacina.")
	}
	if sifraText := strings.TrimSpace(params.Sifra); sifraText == "" {
		add("sifraartikla", common.ErrMsgObavezanPodatak)
	} else if v, err := strconv.Atoi(sifraText); err != nil {
		add("sifraartikla", robnoUnosMsgNePostoji)
	} else if exists, err := s.robno.RsifPostoji(ctx, v); err != nil {
		add("sifraartikla", common.ErrMsgReadData)
	} else if !exists {
		add("sifraartikla", robnoUnosMsgNePostoji)
	}
	if strings.TrimSpace(params.Kolic) == "" {
		add("kolicina", common.ErrMsgObavezanPodatak)
	} else if common.StringToFloat64(params.Kolic) <= 0 {
		add("kolicina", robInterniMsgVeceOdNule)
	}
	if common.StringToFloat64(params.Cena) <= 0 {
		add("cenaopd", robInterniMsgVeceOdNule)
	}
	return fieldErrors
}

// robPtgStavka adds a stavka to a popis u toku godine in one transaction (the legacy CreateRpro of
// rob_opsti_dokPOPIS): the iznos is the količina by the cena (negative with the količina for a storno); the
// stavka has the columns of its document (the radnik as pkto and pana), the artikal, the konto, the cena, the
// nabavna cena, the lot and the next redni broj; the nalog gets the iznos (AzurirajRNAL), the document the
// prodajna and the nabavna vrednost, the porez (by the stopa of the tarifa of the artikal on the datum of the
// document), the iznos and a stavka more. The stanje is not changed (the legacy AzurirajRSTA and AzurirajDRSTA
// are commented out); the SR is not kept.
func (s *RobnoDokumentaResource) robPtgStavka(ctx context.Context, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (rproID int64, err error) {
	tx, err := s.unos.Rpro.BeginTx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if rproID, err = s.robPtgDodajStavku(ctx, tx, k, params); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robPtgDodajStavku adds a stavka to the popis u toku godine of k in the transaction tx (see robPtgStavka).
func (s *RobnoDokumentaResource) robPtgDodajStavku(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, params domain.RobOpstiStavkaParams) (int64, error) {
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return 0, fmt.Errorf("no user session found")
	}
	sifra := common.StringToInt(params.Sifra)
	konto := strings.ReplaceAll(strings.TrimSpace(params.Konto), " ", "")
	kolicinaForme := common.StringToFloat64(params.Kolic)
	cena, ncena := common.StringToFloat64(params.Cena), common.StringToFloat64(params.Ncena)
	kolicina, iznos := math.Abs(kolicinaForme), math.Abs(robZaokruzi(kolicinaForme*cena, 2))
	if k.Storno() {
		kolicina, iznos = -kolicina, -iznos
	}
	rok, _ := robnoUnosDatum(params.Rok)
	var rsifID int64
	var po int
	rsifID, po, found, err := s.robno.RsifIDPo(ctx, tx, sifra)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("artikal ne postoji")
	}
	pkto, pana, pdvpct, err := s.robno.RdokPktoPanaStopa(ctx, tx, k.RdokID, po)
	if err != nil {
		return 0, err
	}
	rbr, err := s.robno.RproSledeciRbrNaloga(ctx, tx, k.RnalID, k.RdokID)
	if err != nil {
		return 0, err
	}
	rproID, err := s.unos.Rpro.CreateWithTx(ctx, tx, &domain.RobRproDto{}, "rproid", []domain.Fields{
		{Name: "rnalid", Value: strconv.FormatInt(k.RnalID, 10)},
		{Name: "nalog", Value: strconv.Itoa(k.Nalog)},
		{Name: "rdokid", Value: strconv.FormatInt(k.RdokID, 10)},
		{Name: "vrd", Value: strconv.Itoa(k.Vrd)},
		{Name: "dokum", Value: strconv.Itoa(k.Dokum)},
		robPoljeNullTime("dadok", k.Dadok),
		{Name: "sifra", Value: strconv.Itoa(sifra)},
		{Name: "konto", Value: konto},
		robPoljeFloat("cena", cena),
		robPoljeFloat("kolic", kolicina),
		robPoljeFloat("iznos", iznos),
		{Name: "fkto", Value: k.Fkto},
		{Name: "fana", Value: k.Fana},
		robPoljeText("pkto", pkto),
		robPoljeText("pana", pana),
		robPoljeFloat("ncena", ncena),
		{Name: "rab", Value: "0"},
		{Name: "po", Value: strconv.Itoa(po)},
		{Name: "mma", Value: "0"},
		{Name: "mag", Value: strconv.Itoa(k.Mag)},
		{Name: "magaciniid", Value: strconv.Itoa(k.MagaciniID)},
		{Name: "brst", Value: strconv.Itoa(rbr)},
		{Name: "rbr", Value: strconv.Itoa(rbr)},
		{Name: "rsifid", Value: strconv.FormatInt(rsifID, 10)},
		robPoljeText("otk", params.Otk),
		robPoljeText("serija", params.Serija),
		{Name: "roktr", Value: strconv.FormatInt(RobWinDevDatum(rok), 10)},
		// NOT NULL without a default.
		{Name: "sifartkup", Value: ""}, {Name: "taxcat", Value: ""}, {Name: "taxexreco", Value: ""},
	})
	if err != nil {
		return 0, err
	}
	if err = s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), iznos); err != nil {
		return 0, err
	}
	if err = s.robno.RdokDodajVrednosti(ctx, tx, k.RdokID, kolicinaForme*cena, kolicinaForme*ncena, kolicina*cena*pdvpct/100, iznos); err != nil {
		return 0, err
	}
	return rproID, nil
}

// robPtgBrisi deletes a stavka of a popis u toku godine in the transaction tx (the legacy BrisiRPRO of
// rob_opsti_dokPOPIS): the document loses its iznos, its prodajna and nabavna vrednost, its porez and a stavka,
// the stavka is deleted and the nalog loses its iznos (AzurirajRNAL, after the deletion: the legacy only counts
// the stavke of the nalog, its AzurirajRNAL is commented out). The stanje is not changed; the redni brojevi of the
// other stavke stay (the legacy does not number them again).
func (s *RobnoDokumentaResource) robPtgBrisi(ctx context.Context, tx db.Transaction, k domain.RobStavkeKontekst, rproID int64) error {
	kolicina, cena, ncena, iznos, po, err := s.robno.RproPopisStavka(ctx, tx, rproID)
	if err != nil {
		return err
	}
	pdvpct, err := s.robno.RdokStopa(ctx, tx, k.RdokID, po)
	if err != nil {
		return err
	}
	kolicinaForme := math.Abs(kolicina)
	if err := s.robno.RdokOduzmiVrednosti(ctx, tx, k.RdokID, kolicinaForme*cena, kolicinaForme*ncena, kolicina*cena*pdvpct/100, iznos); err != nil {
		return err
	}
	if err := s.unos.Rpro.DeleteWithTx(ctx, tx, "rproid", rproID); err != nil {
		return err
	}
	return s.AzurirajRNAL(ctx, tx, k.RnalID, k.Dugovno(), k.Potrazno(), -iznos)
}
