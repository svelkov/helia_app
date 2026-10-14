package robno

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/repository"
)

// AzurirajStanja posts a stavka of a robni dokument (the legacy GlobalProcROBNO.AZurirajStanja), in the
// transaction of the save of the stavka (after the stavka is added): the nalog rnalID, the document rdokID
// and the stavka rproID with its šifra and konto, the količina kol and the iznos dokIznos (negative for a
// storno, as the caller gives them). See robStanja for what is posted.
func (s *RobnoDokumentaResource) AzurirajStanja(ctx context.Context, tx db.Transaction, rnalID, rdokID, rproID int64, sifra int, konto string, kol, dokIznos float64) error {
	return s.robStanja(ctx, tx, rnalID, rdokID, rproID, sifra, konto, kol, dokIznos, 1)
}

// BrisiStanja takes a stavka of a robni dokument back (the legacy GlobalProcROBNO.BrisiStanja), in the
// transaction of the delete of the stavka (before the stavka is deleted: its cene are read): the parameters
// of AzurirajStanja with the količina and the iznos as they were posted. See robStanja.
func (s *RobnoDokumentaResource) BrisiStanja(ctx context.Context, tx db.Transaction, rnalID, rdokID, rproID int64, sifra int, konto string, kol, dokIznos float64) error {
	return s.robStanja(ctx, tx, rnalID, rdokID, rproID, sifra, konto, kol, dokIznos, -1)
}

// robStanjaStavka is what the posting of a stavka reads of it, of its document and of its nalog.
type robStanjaStavka struct {
	cena, mcenap, pcena, fcena, ncena, mcena, rab, vra, mra, vma, mma float64
	po                                                                int
	vrd                                                               int
	dadok                                                             sql.NullTime
	magaciniID, mag                                                   int
	tipdok                                                            string
	danal                                                             sql.NullTime
}

// robStanja is the posting of a stavka (znak 1, AzurirajStanja) or its taking back (znak -1, BrisiStanja) by
// the knjiženje of the vrsta of the document (dokvrsta of the god, kar and vrd; without it nothing is posted,
// the legacy HTransactionCancel):
//   - the vrste of the groups PLS, PLT, PLM, NIV, POP and PRE set the cena of the artikal (rcene.cena, and
//     its mcenap when it is posted; the PLT also the other cene, the rabat and the marže of the stavka). The
//     legacy adds the missing rsif with only its šifra, which rsif does not allow (NOT NULL, foreign keys):
//     an artikal without its rsif keeps its cene;
//   - the nalog: the broj stavki ± 1, the iznos on its duguje (dugovno) or potražuje (potražno);
//   - the document: the redni broj and the broj stavki ± 1, the iznos; a posting also adds the iznos to the
//     prodajna vrednost and its porez (the poreska stopa of the oznaka of the stavka on the datum of the
//     document, the legacy VratiPorez2) to the porez;
//   - the stanje of the artikal in the konto of the magacin of the nalog (rsta; not the group DIR): the
//     kretanje of the month of the document by the knjiženje (dugovno: dug and ulaz, potražno: pot and izlaz,
//     the početno stanje for the vrsta naloga "00"; the NIV not the količina), the broj stavki ± 1, the cena
//     of the stavka for the PLS, PLT, PLM and NIV. A new stanje gets the cena of the stavka, the datum of the
//     nalog and the broj stavki 1; rsta.rsifid references rsif, so an artikal without its rsif (e.g. the
//     stavke of the šifra 0) gets no new stanje.
//
// The legacy SR, RSAL and RSIN are not kept.
func (s *RobnoDokumentaResource) robStanja(ctx context.Context, tx db.Transaction, rnalID, rdokID, rproID int64, sifra int, konto string, kol, dokIznos float64, znak int) error {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return fmt.Errorf("no user session found")
	}
	st := robStanjaStavka{}
	c, err := s.robno.RproCeneStavke(ctx, tx, rproID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	st.cena, st.mcenap, st.pcena, st.fcena, st.ncena, st.mcena = c.Cena, c.Mcenap, c.Pcena, c.Fcena, c.Ncena, c.Mcena
	st.rab, st.vra, st.mra, st.vma, st.mma, st.po = c.Rab, c.Vra, c.Mra, c.Vma, c.Mma, c.Po
	if st.vrd, st.dadok, err = s.robno.RdokVrdDadok(ctx, tx, rdokID); err != nil {
		return err
	}
	n, err := s.robno.RnalMagacinNaloga(ctx, tx, rnalID)
	if err != nil {
		return err
	}
	st.magaciniID, st.mag, st.tipdok, st.danal = n.MagaciniID, n.Mag, n.Tipdok, n.Danal
	kv, err := s.robno.DokvrstaKnjizenjeVrste(ctx, tx, st.vrd)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("vrsta dokumenta %d ne postoji", st.vrd)
	}
	if err != nil {
		return err
	}
	k := robKnjizenje{kodknj: kv.Kodknj, grpdok: kv.Grpdok, predznak: kv.Predznak}
	dugovno, potrazno := k.kodknj == "D", k.kodknj == "P"
	niv := k.grpdok == "NIV"
	iznos := float64(znak) * dokIznos

	// The cene of the artikal.
	if robSadrzi([]string{"PLS", "PLT", "PLM", "NIV", "POP", "PRE"}, k.grpdok) {
		cene := []repository.Izmena{{Kolona: "cena", Vrednost: st.cena}}
		dodaj := func(kolona string, vrednost float64) {
			cene = append(cene, repository.Izmena{Kolona: kolona, Vrednost: vrednost})
		}
		if znak > 0 {
			dodaj("mcenap", st.mcenap)
		}
		if k.grpdok == "PLT" {
			dodaj("pcena", st.pcena)
			dodaj("fcena", st.fcena)
			dodaj("ncena", st.ncena)
			dodaj("mcena", st.mcena)
			dodaj("rab", st.rab)
			dodaj("vra", st.vra)
			dodaj("mra", st.mra)
			dodaj("vma", st.vma)
			dodaj("mma", st.mma)
		}
		if err = s.robno.RcenePostaviCene(ctx, tx, sifra, cene); err != nil {
			return err
		}
	}

	// The nalog.
	if err = s.robno.RnalKnjizi(ctx, tx, rnalID, znak, robStanjaIznos(dugovno, iznos), robStanjaIznos(potrazno, iznos)); err != nil {
		return err
	}

	// The document.
	if znak > 0 {
		stopa := 0.0
		if st.dadok.Valid {
			if stopa, err = s.robno.RporStopa(ctx, tx, st.po, st.dadok.Time); err != nil {
				return err
			}
		}
		err = s.robno.RdokKnjiziStavku(ctx, tx, rdokID, iznos, iznos*stopa/100)
	} else {
		err = s.robno.RdokVratiStavku(ctx, tx, rdokID, iznos)
	}
	if err != nil {
		return err
	}

	// The stanje of the artikal.
	if k.grpdok == "DIR" {
		return nil
	}
	p := domain.RobPromenaStanja{Kolicina: float64(znak) * kol, Iznos: iznos, Tipdok: st.tipdok}
	if niv {
		p.Kolicina = 0
	}
	m := robStanjeMesec(robStanjaDatum(st.dadok))
	rstaID, err := s.robno.RstaIDZaIzmenu(ctx, tx, st.magaciniID, konto, sifra)
	if errors.Is(err, sql.ErrNoRows) {
		rsifID, err := s.robno.RsifIDIliNula(ctx, tx, sifra)
		if err != nil {
			return err
		}
		if rsifID == 0 {
			return nil
		}
		kolone := robStanjeKolone(p, k, m)
		kolone["rsifid"], kolone["sifra"], kolone["konto"] = rsifID, sifra, konto
		kolone["mag"], kolone["magaciniid"] = st.mag, st.magaciniID
		if st.danal.Valid {
			kolone["zdat"] = st.danal.Time.Format(common.HtmlLayout)
		}
		kolone["cena"], kolone["brst"] = st.cena, 1
		kolone["tz"], kolone["tzi"] = "", ""
		return s.robno.StanjeDodaj(ctx, tx, repository.TabelaRsta, kolone)
	}
	if err != nil {
		return err
	}
	izmene := robStanjeKretanje(p, k, m)
	if robSadrzi([]string{"PLS", "PLT", "PLM", "NIV"}, k.grpdok) {
		izmene = append(izmene, repository.Izmena{Kolona: "cena", Vrednost: st.cena})
	}
	izmene = append(izmene, repository.Izmena{Kolona: "brst", Dodaj: true, Vrednost: znak})
	return s.robno.StanjeIzmeni(ctx, tx, repository.TabelaRsta, rstaID, izmene)
}

// robStanjaIznos is the iznos posted on a side of the nalog (its duguje or potražuje): the iznos for the
// side of the knjiženje, else 0.
func robStanjaIznos(strana bool, iznos float64) float64 {
	if !strana {
		return 0
	}
	return math.Round(iznos*100) / 100
}

// robStanjaDatum is the datum of a document (the zero time without it).
func robStanjaDatum(d sql.NullTime) time.Time {
	if !d.Valid {
		return time.Time{}
	}
	return d.Time
}
