package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The prijemnice materijala po porudžbenici of the komercijalno (the legacy pr_PrijemnicaSIROVINE): edok of the vrste
// of the group ULS (the magacini of the tipovi P, S, L and D, the moduli PRO and SVI) or PLT (the magacini of the tip V,
// the moduli ROB and SVI) for a potvrđena porudžbenica dobavljaču (vrdokid), with their stavke in epro and the
// pakovanja of the stavke in kolpak.

// komPrijVrstaUslov is the condition of the vrste (dokvrsta v) of the prijemnice of a magacin (magacini m) of the period
// (the legacy PR_QRY_DOKVRSTA with gsDOKPRO and "ULS" for the magacini of the tipovi P, S, L and D, gsDOKROB and "PLT"
// for the tip V), with its arguments.
const komPrijVrstaUslov = `v.god = m.god and v.kar = m.kar and (
		(upper(trim(coalesce(m.tipmag, ''))) in ('P', 'S', 'L', 'D') and upper(trim(coalesce(v.grpdok, ''))) = ?
			and upper(trim(coalesce(v.modul, ''))) in (?, ?))
		or (upper(trim(coalesce(m.tipmag, ''))) = 'V' and upper(trim(coalesce(v.grpdok, ''))) = ?
			and upper(trim(coalesce(v.modul, ''))) in (?, ?)))`

// komPrijVrstaArgs are the arguments of komPrijVrstaUslov.
var komPrijVrstaArgs = []any{"ULS", "PRO", "SVI", "PLT", "ROB", "SVI"}

// komPrijPotvrdjeni are the statuses of the potvrđene prijemnice: confirmed in a magacin with the analiza (30) or
// without it (40).
var komPrijPotvrdjeni = []any{domain.KomPrijemnicaAnalizirana, domain.KomZahtevPotvrdjen}

// KomPrijVrste returns the vrste dokumenata of the prijemnice of a magacin of the period (by its tip, see
// komPrijVrstaUslov) as "vrd - opis" keyed by the vrd, in the order of the vrd.
func (r *KomercijalnoRepository) KomPrijVrste(ctx context.Context, magaciniID int) ([]domain.ComboItem, error) {
	b := querybuilder.New(`select coalesce(v.vrd, 0)::text as key, coalesce(v.vrd, 0)::text || ' - ' || coalesce(v.opis, '') as value
		from dokvrsta v`).
		AddJoin("join magacini m on m.magaciniid = ?", magaciniID).
		Where(komPrijVrstaUslov, komPrijVrstaArgs...)
	if err := inPeriod(ctx, b, "v."); err != nil {
		return nil, err
	}
	b.OrderBy("v.vrd")
	var rows []domain.ComboItem
	if err := r.selectRows(ctx, "vrste prijemnica", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomPrijemnica returns a prijemnica of the period (a document of a vrsta of the prijemnice of its magacin) with the
// texts of its magacin, org. jedinica and mesto troška, the analiza of the magacin, its vrsta (the first dokvrsta of
// the vrd) and the broj and datum of its porudžbenica; found is false without it.
func (r *KomercijalnoRepository) KomPrijemnica(ctx context.Context, rdokID int64) (domain.KomPrijemnicaRow, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.vrd, 0)::int,
			coalesce((select v.dokvrstaid from dokvrsta v where v.god = e.god and v.kar = e.kar and v.vrd = e.vrd order by v.dokvrstaid limit 1), 0)::bigint,
			coalesce((select trim(v.predznak) from dokvrsta v where v.god = e.god and v.kar = e.kar and v.vrd = e.vrd order by v.dokvrstaid limit 1), ''),
			coalesce(e.dokum, 0)::int, e.dadok, e.datotp, coalesce(e.tipdok, ''), coalesce(e.magaciniid, 0)::int, coalesce(e.mag, 0)::int,
			coalesce(e.mag::text || '-' || coalesce(m.opis, ''), ''), coalesce(m.adresa, ''), coalesce(m.analiza, 0)::int,
			coalesce(e.fkto, ''), coalesce(e.fana, ''),
			e.idorgjed::bigint, coalesce(concat_ws(' - ', nullif(o.ojozn::text, ''), nullif(o.naziv, '')), ''),
			e.mestotrid::bigint, coalesce(concat_ws(' - ', nullif(t.mtroska, ''), nullif(t.opis, '')), ''),
			coalesce(e.doksts, ''), coalesce(e.knjige_1, ''), coalesce(e.vrdokid, 0)::bigint, coalesce(z.dokum, 0)::int, z.dadok
		from edok e`).
		AddJoin("join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr t on t.mestotrid = e.mestotrid").
		AddJoin("left join edok z on z.rdokid = e.vrdokid and z.god = e.god and z.kar = e.kar").
		Where("e.rdokid = ?", rdokID).
		Where("e.vrd in (select v.vrd from dokvrsta v where "+komPrijVrstaUslov+")", komPrijVrstaArgs...)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomPrijemnicaRow{}, false, err
	}
	var p domain.KomPrijemnicaRow
	err := r.row(ctx, nil, "prijemnica", b, &p.RdokID, &p.Vrd, &p.DokvrstaID, &p.Predznak, &p.Dokum, &p.Dadok, &p.Datotp, &p.Tipdok,
		&p.MagaciniID, &p.Mag, &p.MagacinText, &p.MagacinAdresa, &p.Analiza, &p.Fkto, &p.Fana, &p.Orgjed, &p.OrgjedText, &p.MestoTroska,
		&p.MestoTroskaText, &p.Doksts, &p.Knjige1, &p.VrdokID, &p.VezaDokum, &p.VezaDatum)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPrijemnicaRow{}, false, nil
	}
	if err != nil {
		return domain.KomPrijemnicaRow{}, false, err
	}
	return p, true, nil
}

// KomPrijDodaj adds a prijemnica of a porudžbenica of the period (the legacy SnimiEDOK of a new document): not of a
// nalog, nepotvrđena (doksts 10), not posted (knjige N), its broj and datum also the broj and datum of the prijemnica
// (brotp, datotp), its opis "Prijemnica br: <broj>", the rok, the dobavljač, the uslovi, the valuta with its kurs, the
// komercijalista, the org. jedinica and the mesto troška of the porudžbenica (vrdokid), the JM of the ambalaža empty
// (jm1 and jm2 are not null without a default), with the audit columns. It returns its rdokid.
func (r *KomercijalnoRepository) KomPrijDodaj(ctx context.Context, tx db.Transaction, p domain.KomPrijemnicaUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into edok (god, kar, rnalid, vrd, dokum, dadok, mag, magaciniid, iznos, rbr, opis, rok, fkto, fana, pla,
			foot, otp, sifval, kurs, vaziza, komid, mi, brst, brotp, datotp, vrdokid, idorgjed, mestotrid, doksts, knjige_1, knjige_2, jm1,
			jm2, xdatunosa, xopunos)
		select z.god, z.kar, 0, ?, ?, ?, ?, ?, 0, 0, ?, z.rok, z.fkto, z.fana, z.pla, z.foot, z.otp, z.sifval, z.kurs, z.vaziza, z.komid,
			z.mi, 0, ?, ?, z.rdokid, z.idorgjed, z.mestotrid, ?, 'N', 'N', '', '', now(), ?
		from edok z`,
		p.Vrd, p.Dokum, p.Datum, p.Mag, p.MagaciniID, fmt.Sprintf("Prijemnica br: %d", p.Dokum), p.Dokum, p.Datum,
		domain.KomZahtevNepotvrdjen, s.UserName).
		Where("z.rdokid = ?", p.VrdokID).
		Returning("rdokid")
	if err := inPeriod(ctx, b, "z."); err != nil {
		return 0, err
	}
	var rdokID int64
	if err := r.row(ctx, tx, "nova prijemnica", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// KomPrijIzmeni saves the datum of a prijemnica of the period (the legacy SnimiEDOK of a saved document: the datum
// prijemnice, also its datum), with the audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomPrijIzmeni(ctx context.Context, tx db.Transaction, p domain.KomPrijemnicaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("edok").
		Set("dadok = ?", p.Datum).
		Set("datotp = ?", p.Datum).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", p.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "prijemnica", b)
}

// KomPrijPotvrdi sets the status of a nepotvrđena prijemnica of the period (the legacy PotvrdaDOK: 40, or 30 in a
// magacin with the analiza). It returns false when the prijemnica is not nepotvrđena any more.
func (r *KomercijalnoRepository) KomPrijPotvrdi(ctx context.Context, tx db.Transaction, rdokID int64, doksts string) (bool, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	b := querybuilder.NewUpdate("edok").
		Set("doksts = ?", doksts).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", rdokID).
		Where("coalesce(doksts, '') = ?", domain.KomZahtevNepotvrdjen)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	n, err := r.exec(ctx, tx, "potvrda prijemnice", b)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomPrijBrisi deletes a prijemnica of the period with its stavke and their pakovanja (the legacy BTN_Brisanje);
// ErrNotFound without it.
func (r *KomercijalnoRepository) KomPrijBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	if err := r.komPrijKolpakBrisi(ctx, tx, rdokID, 0); err != nil {
		return err
	}
	return r.KomZahtevBrisi(ctx, tx, rdokID)
}

// komPrijKolpakBrisi deletes the pakovanja of the stavke of a prijemnica of the period: of one stavka (rproID), or of
// all of them (rproID 0).
func (r *KomercijalnoRepository) komPrijKolpakBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64) error {
	b := querybuilder.New(`delete from kolpak`).Where("rdokid = ?", rdokID)
	if rproID > 0 {
		b.Where("rproid = ?", rproID)
	}
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "pakovanja prijemnice", b)
	return err
}

// KomPrijVeza returns the potvrđena porudžbenica dobavljaču (doksts 40, the vrste of the modul PRO and the group NRD) of
// the period a prijemnica is made for (the legacy BTN_TraziDOK and PodaciHDR); found is false without it.
func (r *KomercijalnoRepository) KomPrijVeza(ctx context.Context, rdokID int64) (domain.KomPrijVeza, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.dokum, 0)::int, e.dadok, coalesce(e.magaciniid, 0)::int,
			e.idorgjed::bigint, coalesce(concat_ws(' - ', nullif(o.ojozn::text, ''), nullif(o.naziv, '')), ''),
			e.mestotrid::bigint, coalesce(concat_ws(' - ', nullif(t.mtroska, ''), nullif(t.opis, '')), ''), coalesce(e.fkto, ''), coalesce(e.fana, '')
		from edok e`).
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr t on t.mestotrid = e.mestotrid").
		Where("e.rdokid = ?", rdokID).
		Where("coalesce(e.doksts, '') = ?", domain.KomZahtevPotvrdjen).
		Where(`e.vrd in (select v.vrd from dokvrsta v where v.god = e.god and v.kar = e.kar
			and upper(trim(coalesce(v.modul, ''))) = ? and upper(trim(coalesce(v.grpdok, ''))) = ?)`, komZahtevModul, komPorGrupa)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomPrijVeza{}, false, err
	}
	var v domain.KomPrijVeza
	err := r.row(ctx, nil, "porudžbenica prijemnice", b, &v.RdokID, &v.Dokum, &v.Dadok, &v.MagaciniID, &v.Orgjed, &v.OrgjedText, &v.MestoTroska,
		&v.MtText, &v.Fkto, &v.Fana)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPrijVeza{}, false, nil
	}
	if err != nil {
		return domain.KomPrijVeza{}, false, err
	}
	return v, true, nil
}

// KomPrijStavke returns the stavke of a prijemnica of the period in the order of their redni broj, with the naziv and
// the JM of their artikal and the decimals of the JM (jedmere.brdecimala, 2 without it; the legacy PuniDETALJ1 and
// PR_QRY_EPRO).
func (r *KomercijalnoRepository) KomPrijStavke(ctx context.Context, rdokID int64) ([]domain.KomPrijStavkaRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, coalesce(p.rbr, 0)::int as rbr, coalesce(p.konto, '') as konto,
			coalesce(p.sifra, 0)::int as sifra, coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm,
			coalesce(p.trazkolic, 0)::float8 as trazkolic, coalesce(p.kolic, 0)::float8 as kolic, coalesce(p.serija, '') as serija,
			coalesce(p.otk, '') as otk, coalesce(p.roktr, 0)::bigint as roktr, coalesce(p.cena, 0)::float8 as cena,
			coalesce((select j.brdecimala from jedmere j where j.jm = s.jm and j.god = p.god and j.kar = p.kar limit 1), 2)::int as decimale
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr")
	var rows []komPrijStavkaDto
	if err := r.selectRows(ctx, "stavke prijemnice", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPrijStavkaRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPrijStavkaRow{RproID: row.RproID, Rbr: row.Rbr, Konto: row.Konto, Sifra: row.Sifra, Naziv: row.Naziv, JM: row.JM,
			TrazKolic: row.TrazKolic, Kolic: row.Kolic, Serija: row.Serija, Otk: row.Otk, Roktr: row.Roktr, Cena: row.Cena, Decimale: row.Decimale}
	}
	return out, nil
}

// komPrijStavkaDto is a row of KomPrijStavke.
type komPrijStavkaDto struct {
	RproID    int64   `db:"rproid"`
	Rbr       int     `db:"rbr"`
	Konto     string  `db:"konto"`
	Sifra     int     `db:"sifra"`
	Naziv     string  `db:"naziv"`
	JM        string  `db:"jm"`
	TrazKolic float64 `db:"trazkolic"`
	Kolic     float64 `db:"kolic"`
	Serija    string  `db:"serija"`
	Otk       string  `db:"otk"`
	Roktr     int64   `db:"roktr"`
	Cena      float64 `db:"cena"`
	Decimale  int     `db:"decimale"`
}

// KomPrijStavkaDodaj adds a stavka of a prijemnica (the legacy AddEPRO for a stavka entered by hand, AddEPRONAR for a
// stavka of the porudžbenica): the artikal, the konto, the količina na dokumentu and the primljena količina, the OTK,
// the serija, the cena (fcena before the rabat, cena and ncena after it), the iznos, the rabat, the magacin and the
// vrsta, broj and datum of the prijemnica, with the audit columns.
func (r *KomercijalnoRepository) KomPrijStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomPrijStavkaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, sifra, po, rsifid, rbr, mag, magaciniid, konto, trazkolic, kolic, otk, serija,
			fcena, cena, ncena, iznos, rab, vrd, dokum, dadok, xdatunosa, xopunos)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, st.RdokID, st.Sifra, st.Po, st.RsifID, st.Rbr, st.Mag, st.MagaciniID, st.Konto, st.TrazKolic, st.Kolic, st.Otk,
		st.Serija, st.Fcena, st.Cena, st.Cena, st.Iznos, st.Rabat, st.Vrd, st.Dokum, st.Dadok, s.UserName)
	_, err = r.exec(ctx, tx, "nova stavka prijemnice", b)
	return err
}

// KomPrijStavkaIzmeni changes a stavka of a prijemnica of the period (the legacy exit of a changed row of
// TABLE_PR_QRY_EPRO1): its konto, količina na dokumentu, primljena količina, serija and OTK, its iznos the cena by the
// primljena količina, with the audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomPrijStavkaIzmeni(ctx context.Context, tx db.Transaction, st domain.KomPrijStavkaIzmena) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("epro").
		Set("konto = ?", st.Konto).
		Set("trazkolic = ?", st.TrazKolic).
		Set("kolic = ?", st.Kolic).
		Set("serija = ?", st.Serija).
		Set("otk = ?", st.Otk).
		Set("iznos = coalesce(cena, 0) * ?", st.Kolic).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rproid = ?", st.RproID).
		Where("rdokid = ?", st.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stavka prijemnice", b)
}

// KomPrijStavkaBrisi deletes a stavka of a prijemnica of the period with its pakovanje and moves the redni brojevi of
// the stavke after it one back (the legacy BTN_BrisiDET and BTN_VratiJedan); ErrNotFound without it.
func (r *KomercijalnoRepository) KomPrijStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error {
	if err := r.komPrijKolpakBrisi(ctx, tx, rdokID, rproID); err != nil {
		return err
	}
	return r.KomZahtevStavkaBrisi(ctx, tx, rdokID, rproID, rbr)
}

// KomPrijStavkeBrisi deletes all the stavke of a prijemnica of the period with their pakovanja (the legacy
// BTN_VratiSve).
func (r *KomercijalnoRepository) KomPrijStavkeBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	if err := r.komPrijKolpakBrisi(ctx, tx, rdokID, 0); err != nil {
		return err
	}
	b := querybuilder.New(`delete from epro`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "stavke prijemnice", b)
	return err
}

// KomPrijIzvor returns a stavka of the porudžbenica of the period a prijemnica takes (the legacy AddEPRONAR): its
// artikal with its poreska stopa, its tražena količina, its cena and its rabat; found is false without it.
func (r *KomercijalnoRepository) KomPrijIzvor(ctx context.Context, tx db.Transaction, porudzbenicaID, rproID int64) (domain.KomPrijIzvor, bool, error) {
	b := querybuilder.New(`select p.rproid::bigint, coalesce(p.sifra, 0)::int, coalesce(s.po, 0)::int, coalesce(p.rsifid, 0)::bigint,
			coalesce(p.trazkolic, 0)::float8, coalesce(p.cena, 0)::float8, coalesce(p.rab, 0)::float8
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rproid = ?", rproID).
		Where("p.rdokid = ?", porudzbenicaID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return domain.KomPrijIzvor{}, false, err
	}
	var st domain.KomPrijIzvor
	err := r.row(ctx, tx, "stavka porudžbenice", b, &st.RproID, &st.Sifra, &st.Po, &st.RsifID, &st.TrazKolic, &st.Cena, &st.Rabat)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPrijIzvor{}, false, nil
	}
	if err != nil {
		return domain.KomPrijIzvor{}, false, err
	}
	return st, true, nil
}

// KomPrijImaArtikal tells whether a prijemnica of the period has a stavka of the artikal (the legacy check of AddEPRONAR
// "Ovakvi podaci već postoje").
func (r *KomercijalnoRepository) KomPrijImaArtikal(ctx context.Context, tx db.Transaction, rdokID, rsifID int64) (bool, error) {
	b := querybuilder.New(`select count(*)::int from epro`).Where("rdokid = ?", rdokID).Where("rsifid = ?", rsifID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	var n int
	if err := r.row(ctx, tx, "artikal u prijemnici", b, &n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomPrijArtikal returns the artikal of the šifra of the period (rsif, the first by its id; the legacy exit of
// EDT_SIFRA); found is false without it.
func (r *KomercijalnoRepository) KomPrijArtikal(ctx context.Context, sifra int) (domain.KomPrijArtikal, bool, error) {
	b := querybuilder.New(`select r.rsifid::bigint, coalesce(r.sifra, 0)::int, coalesce(r.naziv, ''), coalesce(r.jm, ''), coalesce(r.po, 0)::int
		from rsif r`).
		Where("r.sifra = ?", sifra)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return domain.KomPrijArtikal{}, false, err
	}
	b.OrderBy("r.rsifid").Limit(1)
	var a domain.KomPrijArtikal
	err := r.row(ctx, nil, "artikal prijemnice", b, &a.RsifID, &a.Sifra, &a.Naziv, &a.JM, &a.Po)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPrijArtikal{}, false, nil
	}
	if err != nil {
		return domain.KomPrijArtikal{}, false, err
	}
	return a, true, nil
}

// KomPrijOtkBroj returns the number of the stavke of the artikal in the potvrđene prijemnice (not doksts 10) of the
// vrsta of the period (the legacy VratiOTK, the base of the OTK of a new stavka).
func (r *KomercijalnoRepository) KomPrijOtkBroj(ctx context.Context, tx db.Transaction, vrd, sifra int) (int, error) {
	b := querybuilder.New(`select count(p.sifra)::int from epro p`).
		AddJoin("join edok e on e.rdokid = p.rdokid and e.god = p.god and e.kar = p.kar").
		Where("e.vrd = ?", vrd).
		Where("p.sifra = ?", sifra).
		Where("coalesce(e.doksts, '') <> ?", domain.KomZahtevNepotvrdjen)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return 0, err
	}
	var n int
	if err := r.row(ctx, tx, "OTK artikla", b, &n); err != nil {
		return 0, err
	}
	return n, nil
}

// KomPrijKonta returns the konta of a magacin of the period (magkonto) in their order, the konto of a stavka of a
// prijemnica (the first one is the default, the legacy MAGKONTO of the magacin).
func (r *KomercijalnoRepository) KomPrijKonta(ctx context.Context, magaciniID int) ([]domain.ComboItem, error) {
	b := querybuilder.New(`select distinct trim(k.konto) as key, trim(k.konto) as value from magkonto k`).
		Where("k.magaciniid = ?", magaciniID).
		Where("trim(coalesce(k.konto, '')) <> ''")
	if err := inPeriod(ctx, b, "k."); err != nil {
		return nil, err
	}
	b.OrderByExpr("trim(k.konto)")
	var rows []domain.ComboItem
	if err := r.selectRows(ctx, "konta magacina", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomPrijPakovanja returns the stavke of a prijemnica of the period in the order of their redni broj with their
// pakovanje (the first kolpak of the stavka; 0 without it), the legacy pr_Pakovanja.
func (r *KomercijalnoRepository) KomPrijPakovanja(ctx context.Context, rdokID int64) ([]domain.KomPrijPakovanjeRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, coalesce(p.rbr, 0)::int as rbr, coalesce(p.sifra, 0)::int as sifra,
			coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm,
			coalesce((select k.kolpopak from kolpak k where k.rproid = p.rproid and k.rdokid = p.rdokid and k.god = p.god and k.kar = p.kar
				order by k.kolpakid limit 1), 0)::float8 as kolpopak,
			coalesce((select k.brpak from kolpak k where k.rproid = p.rproid and k.rdokid = p.rdokid and k.god = p.god and k.kar = p.kar
				order by k.kolpakid limit 1), 0)::float8 as brpak
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr")
	var rows []komPrijPakovanjeDto
	if err := r.selectRows(ctx, "pakovanja prijemnice", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPrijPakovanjeRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPrijPakovanjeRow{RproID: row.RproID, Rbr: row.Rbr, Sifra: row.Sifra, Naziv: row.Naziv, JM: row.JM, KolPoPak: row.KolPoPak,
			BrPak: row.BrPak, Uskladisteno: row.KolPoPak * row.BrPak}
	}
	return out, nil
}

// komPrijPakovanjeDto is a row of KomPrijPakovanja.
type komPrijPakovanjeDto struct {
	RproID   int64   `db:"rproid"`
	Rbr      int     `db:"rbr"`
	Sifra    int     `db:"sifra"`
	Naziv    string  `db:"naziv"`
	JM       string  `db:"jm"`
	KolPoPak float64 `db:"kolpopak"`
	BrPak    float64 `db:"brpak"`
}

// KomPrijPakovanjeSnimi saves the pakovanje of a stavka of a prijemnica of the period (kolpak): the količina po
// pakovanju and the broj pakovanja of its pakovanje, a new one when it has none, with the audit columns.
func (r *KomercijalnoRepository) KomPrijPakovanjeSnimi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, kolPoPak, brPak float64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("kolpak").
		Set("kolpopak = ?", kolPoPak).
		Set("brpak = ?", brPak).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rproid = ?", rproID).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	n, err := r.exec(ctx, tx, "pakovanje stavke", b)
	if err != nil || n > 0 {
		return err
	}
	novo := querybuilder.New(`insert into kolpak (god, kar, rdokid, rproid, kolpopak, brpak, xdatunosa, xopunos)
		values (?, ?, ?, ?, ?, ?, now(), ?)`, s.God, s.Kar, rdokID, rproID, kolPoPak, brPak, s.UserName)
	_, err = r.exec(ctx, tx, "novo pakovanje stavke", novo)
	return err
}

// KomPrijPrikaz returns the prijemnice of the list of the main screen (the legacy PR_QRY_EDOK of the tab "Pregled
// prijemnica"): the ones of the vrste of the prijemnice of their magacin with the status (the nepotvrđene, else the
// potvrđene: 30 and 40) in the magacini of the user of the session (maguser), of one magacin or of all of them
// (SviMagacini), the last ones first, with the "mag-opis" of the magacin and the naziv of the dobavljač.
func (r *KomercijalnoRepository) KomPrijPrikaz(ctx context.Context, p domain.KomPrijPrikazParams) ([]domain.KomPrijPrikazRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select e.rdokid::bigint as rdokid,
			coalesce(m.mag::text || '-' || coalesce(m.opis, ''), '') as magacin, coalesce(e.vrd, 0)::int as vrd,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok, coalesce(e.fkto, '') as fkto, coalesce(e.fana, '') as fana,
			coalesce(f.naziv, '') as dobavljac, coalesce(e.opis, '') as opis, coalesce(e.rok, 0)::int as rok, e.datiz as datiz,
			coalesce(e.otp, '') as otp, coalesce(e.kamt, '') as kamt
		from edok e`).
		AddJoin("join (select distinct mu.mag from maguser mu where upper(mu.iduser) = ?) u on u.mag = e.mag",
			strings.ToUpper(strings.TrimSpace(s.UserName))).
		AddJoin("join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana").
		Where("e.vrd in (select v.vrd from dokvrsta v where "+komPrijVrstaUslov+")", komPrijVrstaArgs...)
	if p.Doksts == domain.KomZahtevNepotvrdjen {
		b.Where("coalesce(e.doksts, '') = ?", domain.KomZahtevNepotvrdjen)
	} else {
		b.Where("coalesce(e.doksts, '') in (?, ?)", komPrijPotvrdjeni...)
	}
	if !p.SviMagacini {
		b.Where("e.magaciniid = ?", p.MagaciniID)
	}
	komPretraga(b, p.Tekst, "e.dokum::text", "e.opis", "f.naziv", "e.fana")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return nil, err
	}
	b.OrderByExpr("e.dadok desc nulls last, e.vrd desc, e.dokum desc")
	var rows []komPrijPrikazDto
	if err := r.selectRows(ctx, "pregled prijemnica", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPrijPrikazRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPrijPrikazRow{RdokID: row.RdokID, Magacin: row.Magacin, Vrd: row.Vrd, Dokum: row.Dokum, Dadok: row.Dadok, Fkto: row.Fkto,
			Fana: row.Fana, Dobavljac: row.Dobavljac, Opis: row.Opis, Rok: row.Rok, Datiz: row.Datiz, Otp: row.Otp, Kamt: row.Kamt}
	}
	return out, nil
}

// komPrijPrikazDto is a row of KomPrijPrikaz.
type komPrijPrikazDto struct {
	RdokID    int64        `db:"rdokid"`
	Magacin   string       `db:"magacin"`
	Vrd       int          `db:"vrd"`
	Dokum     int          `db:"dokum"`
	Dadok     sql.NullTime `db:"dadok"`
	Fkto      string       `db:"fkto"`
	Fana      string       `db:"fana"`
	Dobavljac string       `db:"dobavljac"`
	Opis      string       `db:"opis"`
	Rok       int          `db:"rok"`
	Datiz     sql.NullTime `db:"datiz"`
	Otp       string       `db:"otp"`
	Kamt      string       `db:"kamt"`
}
