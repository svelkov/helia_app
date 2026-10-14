package robno

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/infrastructure/db"
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
	userSession := domain.GetSessionFromStdContext(ctx)
	if userSession == nil {
		return fmt.Errorf("no user session found")
	}
	god, kar := userSession.SelectedGod, userSession.SelectedKar
	st := robStanjaStavka{}
	err := tx.QueryRowContext(ctx, `select coalesce(cena, 0)::float8, coalesce(mcenap, 0)::float8, coalesce(pcena, 0)::float8,
			coalesce(fcena, 0)::float8, coalesce(ncena, 0)::float8, coalesce(mcena, 0)::float8, coalesce(rab, 0)::float8,
			coalesce(vra, 0)::float8, coalesce(mra, 0)::float8, coalesce(vma, 0)::float8, coalesce(mma, 0)::float8, coalesce(po, 0)
		from rpro where rproid = $1`, rproID).Scan(&st.cena, &st.mcenap, &st.pcena, &st.fcena, &st.ncena, &st.mcena, &st.rab,
		&st.vra, &st.mra, &st.vma, &st.mma, &st.po)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err = tx.QueryRowContext(ctx, `select coalesce(vrd, 0), dadok from rdok where rdokid = $1`, rdokID).Scan(&st.vrd, &st.dadok); err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, `select coalesce(magaciniid, 0), coalesce(mag, 0), coalesce(tipdok, ''), danal from rnal where rnalid = $1`,
		rnalID).Scan(&st.magaciniID, &st.mag, &st.tipdok, &st.danal); err != nil {
		return err
	}
	var k robKnjizenje
	err = tx.QueryRowContext(ctx, `select upper(trim(coalesce(kodknj, ''))), upper(trim(coalesce(grpdok, ''))), trim(coalesce(predznak, ''))
		from dokvrsta where god = $1 and kar = $2 and vrd = $3 order by dokvrstaid limit 1`, god, kar, st.vrd).Scan(&k.kodknj, &k.grpdok, &k.predznak)
	if err == sql.ErrNoRows {
		return fmt.Errorf("vrsta dokumenta %d ne postoji", st.vrd)
	}
	if err != nil {
		return err
	}
	dugovno, potrazno := k.kodknj == "D", k.kodknj == "P"
	niv := k.grpdok == "NIV"
	iznos := float64(znak) * dokIznos

	// The cene of the artikal.
	if robSadrzi([]string{"PLS", "PLT", "PLM", "NIV", "POP", "PRE"}, k.grpdok) {
		set := []string{"cena = $4"}
		args := []any{god, kar, sifra, st.cena}
		dodaj := func(kolona string, vrednost float64) {
			args = append(args, vrednost)
			set = append(set, fmt.Sprintf("%s = $%d", kolona, len(args)))
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
		if _, err = tx.ExecContext(ctx, `update rcene set `+strings.Join(set, ", ")+`
			where god = $1 and kar = $2 and sifra = $3 and exists (select 1 from rsif where god = $1 and kar = $2 and sifra = $3)`,
			args...); err != nil {
			return err
		}
	}

	// The nalog.
	if _, err = tx.ExecContext(ctx, `update rnal set brst = coalesce(brst, 0) + $2,
			dug = round((coalesce(dug, 0) + $3)::numeric, 2), pot = round((coalesce(pot, 0) + $4)::numeric, 2)
		where rnalid = $1`, rnalID, znak, robStanjaIznos(dugovno, iznos), robStanjaIznos(potrazno, iznos)); err != nil {
		return err
	}

	// The document.
	if znak > 0 {
		stopa := 0.0
		if st.dadok.Valid {
			if err = tx.QueryRowContext(ctx, `select `+fmt.Sprintf(robPoreskaStopaSQL, "$1", "$2"), st.po, st.dadok.Time).Scan(&stopa); err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, `update rdok set rbr = coalesce(rbr, 0) + 1, brst = coalesce(brst, 0) + 1,
				iznos = round((coalesce(iznos, 0) + $2)::numeric, 2), prodvred = round((coalesce(prodvred, 0) + $2)::numeric, 2),
				vporez = round((coalesce(vporez, 0) + $3)::numeric, 2)
			where rdokid = $1`, rdokID, iznos, iznos*stopa/100)
	} else {
		_, err = tx.ExecContext(ctx, `update rdok set rbr = coalesce(rbr, 0) - 1, brst = coalesce(brst, 0) - 1,
				iznos = round((coalesce(iznos, 0) + $2)::numeric, 2)
			where rdokid = $1`, rdokID, iznos)
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
	var rstaID int64
	err = tx.QueryRowContext(ctx, `select rstaid from rsta where god = $1 and kar = $2 and magaciniid = $3 and konto = $4 and sifra = $5
		order by rstaid limit 1 for update`, god, kar, st.magaciniID, konto, sifra).Scan(&rstaID)
	if err == sql.ErrNoRows {
		var rsifID int64
		if err = tx.QueryRowContext(ctx, `select coalesce((select rsifid from rsif where god = $1 and kar = $2 and sifra = $3 limit 1), 0)`,
			god, kar, sifra).Scan(&rsifID); err != nil {
			return err
		}
		if rsifID == 0 {
			return nil
		}
		kolone := robStanjeKolone(p, k, m)
		kolone["god"], kolone["kar"] = god, kar
		kolone["rsifid"], kolone["sifra"], kolone["konto"] = rsifID, sifra, konto
		kolone["mag"], kolone["magaciniid"] = st.mag, st.magaciniID
		if st.danal.Valid {
			kolone["zdat"] = st.danal.Time.Format(common.HtmlLayout)
		}
		kolone["cena"], kolone["brst"] = st.cena, 1
		kolone["tz"], kolone["tzi"] = "", ""
		return robInsertStanja(ctx, tx, "rsta", kolone)
	}
	if err != nil {
		return err
	}
	set, args := robStanjeKretanje(p, k, m, 1)
	if robSadrzi([]string{"PLS", "PLT", "PLM", "NIV"}, k.grpdok) {
		args = append(args, st.cena)
		set = append(set, fmt.Sprintf("cena = $%d", len(args)))
	}
	args = append(args, znak)
	set = append(set, fmt.Sprintf("brst = coalesce(brst, 0) + $%d", len(args)))
	return robUpdateStanja(ctx, tx, "rsta", rstaID, set, args)
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
