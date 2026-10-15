package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The opšti robni dokumenti (and the finansijski, the KOL and the nivelacije of them), in the period
// of the session.

// DokvrstaOpis returns the opis of the vrsta vrd; "" without it.
func (r *RobnoRepository) DokvrstaOpis(ctx context.Context, vrd int) (string, error) {
	b := querybuilder.New(`select coalesce(opis, '') from dokvrsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	b.Where("vrd = ?", vrd).Limit(1)
	var opis string
	err := r.row(ctx, nil, "opis vrste", b, &opis)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return opis, err
}

// RdokProtivstav returns the protivstav of an opšti dokument: key = konto, c1 = šifra (trimmed),
// c2 = org. jedinica of the roba, c3 = its oznaka - naziv, c4 = mesto troška of the roba,
// c5 = its oznaka - opis.
func (r *RobnoRepository) RdokProtivstav(ctx context.Context, rdokID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select coalesce(trim(r.tkonto), '') as key, coalesce(trim(r.tsifra), '') as c1,
			coalesce(r.idorgjed1::text, '') as c2, coalesce(concat_ws(' - ', nullif(o.ojozn, ''), nullif(o.naziv, '')), '') as c3,
			coalesce(r.mestotrid1::text, '') as c4, coalesce(concat_ws(' - ', nullif(m.mtroska, ''), nullif(m.opis, '')), '') as c5
		from rdok r`).
		AddJoin("left join orgjed o on o.idorgjed = r.idorgjed1").
		AddJoin("left join mestotr m on m.mestotrid = r.mestotrid1").
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "protivstav dokumenta", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// OrgjedPrvaSaMestom returns the first org. jedinica (by its oznaka) with its first mesto troška:
// key = idorgjed, c1 = oznaka - naziv, c2 = mestotrid, c3 = oznaka - opis.
func (r *RobnoRepository) OrgjedPrvaSaMestom(ctx context.Context) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select o.idorgjed::text as key, coalesce(concat_ws(' - ', nullif(o.ojozn, ''), nullif(o.naziv, '')), '') as c1,
			coalesce(m.mestotrid::text, '') as c2, coalesce(concat_ws(' - ', nullif(m.mtroska, ''), nullif(m.opis, '')), '') as c3
		from orgjed o`).
		AddJoin(`left join lateral (select x.mestotrid, x.mtroska, x.opis from mestotr x where x.idorgjed = o.idorgjed
			order by x.mtroska, x.mestotrid limit 1) m on true`)
	if err := inPeriod(ctx, b, "o."); err != nil {
		return nil, err
	}
	b.OrderBy("o.ojozn, o.idorgjed").Limit(1)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "org. jedinica", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// MagacinOpsti returns what the entry of the stavke needs of a magacin and of the nalog: key = tip
// zaliha, c1 = način vođenja zaliha, c2 = tip magacina, c3 = opis of the nalog.
func (r *RobnoRepository) MagacinOpsti(ctx context.Context, magaciniID int, rnalID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select coalesce(m.tipzal, 0)::text as key, coalesce(m.nacvodzal, 0)::text as c1,
			coalesce(m.tipmag, '') as c2,
			coalesce((select n.opis from rnal n where n.rnalid = ? and n.god = m.god and n.kar = m.kar), '') as c3
		from magacini m`, rnalID).Where("m.magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, "m."); err != nil {
		return nil, err
	}
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "magacin", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// FkplNaziv returns the naziv of a konto of the kontni plan (trimmed konto and šifra): the
// sintetički konto (vkonta 2) without a šifra, else the konto with the šifra; found is false without it.
func (r *RobnoRepository) FkplNaziv(ctx context.Context, konto, sifra string) (naziv string, found bool, err error) {
	b := querybuilder.New(`select coalesce(naziv, '') from fkpl`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", false, err
	}
	b.Where("konto = ?", strings.TrimSpace(konto))
	if s := strings.TrimSpace(sifra); s != "" {
		b.Where("sifra = ?", s)
	} else {
		b.Where("vkonta = 2")
	}
	b.Limit(1)
	err = r.row(ctx, nil, "naziv konta", b, &naziv)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return naziv, err == nil, err
}

// RstaOpsti is the stanje of an artikal with its kretanje and cene.
type RstaOpsti struct {
	Ulaz, Izlaz, Dug, Pot, Cena, Ncena, Prosnc float64
}

// RstaStanjeOpsti returns the stanje of an artikal in a konto (trimmed) of a magacin; found is false
// without it.
func (r *RobnoRepository) RstaStanjeOpsti(ctx context.Context, magaciniID int, konto string, sifra int) (RstaOpsti, bool, error) {
	b := querybuilder.New(`select coalesce(ulaz, 0)::float8, coalesce(izlaz, 0)::float8,
			coalesce(dug, 0)::float8, coalesce(pot, 0)::float8, coalesce(cena, 0)::float8,
			coalesce(ncena, 0)::float8, coalesce(prosnc, 0)::float8
		from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RstaOpsti{}, false, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", strings.TrimSpace(konto)).Where("sifra = ?", sifra).OrderBy("rstaid").Limit(1)
	var st RstaOpsti
	err := r.row(ctx, nil, "stanje", b, &st.Ulaz, &st.Izlaz, &st.Dug, &st.Pot, &st.Cena, &st.Ncena, &st.Prosnc)
	if errors.Is(err, sql.ErrNoRows) {
		return RstaOpsti{}, false, nil
	}
	return st, err == nil, err
}

// DrstaLotKolicina returns the stanje of the lot (OTK broj, serija, rok trajanja) of an artikal in a
// konto of a magacin (mag; the texts trimmed); found is false without it.
func (r *RobnoRepository) DrstaLotKolicina(ctx context.Context, mag int, konto string, sifra int, otk, serija string, roktr int64) (float64, bool, error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8 from drsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, false, err
	}
	b.Where("konto = ?", strings.TrimSpace(konto)).Where("sifra = ?", sifra).Where("mag = ?", mag).
		Where("otk = ?", strings.TrimSpace(otk)).Where("serija = ?", strings.TrimSpace(serija)).Where("roktr = ?", roktr).
		OrderBy("rstaid").Limit(1)
	var stanje float64
	err := r.row(ctx, nil, "zaliha lota", b, &stanje)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return stanje, err == nil, err
}

// DrstaLotCene returns the stanje, the cena and the nabavna cena of the lot (OTK broj, serija, rok trajanja) of an
// artikal in a konto of a magacin (mag; the texts trimmed); found is false without it.
func (r *RobnoRepository) DrstaLotCene(ctx context.Context, mag int, konto string, sifra int, otk, serija string, roktr int64) (stanje, cena, ncena float64, found bool, err error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8, coalesce(cena, 0)::float8, coalesce(ncena, 0)::float8 from drsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, 0, false, err
	}
	b.Where("konto = ?", strings.TrimSpace(konto)).Where("sifra = ?", sifra).Where("mag = ?", mag).
		Where("otk = ?", strings.TrimSpace(otk)).Where("serija = ?", strings.TrimSpace(serija)).Where("roktr = ?", roktr).
		OrderBy("rstaid").Limit(1)
	err = r.row(ctx, nil, "zaliha i cene lota", b, &stanje, &cena, &ncena)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, 0, false, nil
	}
	return stanje, cena, ncena, err == nil, err
}

// RsifArtikalOpsti returns the artikal by its šifra, or by its barkod (poBarkodu): key = šifra,
// c1 = naziv, c2 = JM, c3 = konto (trimmed), c4 = tip (upper case); no row when there is none.
func (r *RobnoRepository) RsifArtikalOpsti(ctx context.Context, poBarkodu bool, vrednost string) (*[]domain.SearchRowDto, error) {
	kolona := "a.sifra::text"
	if poBarkodu {
		kolona = "a.barkod"
	}
	b := querybuilder.New(`select a.sifra::text as key, coalesce(a.naziv, '') as c1, coalesce(a.jm, '') as c2,
			trim(coalesce(a.konto, '')) as c3, upper(trim(coalesce(a.tip, ''))) as c4
		from rsif a`)
	if err := inPeriod(ctx, b, "a."); err != nil {
		return nil, err
	}
	b.Where(kolona+" = ?", vrednost).Limit(1)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "artikal", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// RsifIDPoNaziv returns the id, the poreska oznaka and the naziv of an artikal by its šifra; the
// error wraps sql.ErrNoRows when there is none.
func (r *RobnoRepository) RsifIDPoNaziv(ctx context.Context, tx db.Transaction, sifra int) (rsifID int64, po int, naziv string, err error) {
	b := querybuilder.New(`select rsifid, coalesce(po, 0), coalesce(naziv, '') from rsif`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, "", err
	}
	b.Where("sifra = ?", sifra).Limit(1)
	err = r.row(ctx, tx, "artikal", b, &rsifID, &po, &naziv)
	return rsifID, po, naziv, err
}

// DrstaBrisiPrazanLot deletes the zaliha of a lot (trimmed texts) of an artikal in a konto of a
// magacin (mag) when it has no stanje, no vrednost and no stavka.
func (r *RobnoRepository) DrstaBrisiPrazanLot(ctx context.Context, tx db.Transaction, konto string, sifra, mag int, otk, serija string, roktr int64) error {
	b := querybuilder.New(`delete from drsta d`)
	if err := inPeriod(ctx, b, "d."); err != nil {
		return err
	}
	b.Where("d.konto = ?", strings.TrimSpace(konto)).Where("d.sifra = ?", sifra).Where("d.mag = ?", mag).
		Where("d.otk = ?", strings.TrimSpace(otk)).Where("d.serija = ?", strings.TrimSpace(serija)).Where("d.roktr = ?", roktr).
		Where("coalesce(d.ulaz, 0) - coalesce(d.izlaz, 0) = 0").Where("coalesce(d.dug, 0) - coalesce(d.pot, 0) = 0").
		Where(`not exists (select 1 from rpro p where p.god = d.god and p.kar = d.kar and p.konto = d.konto and p.sifra = d.sifra
			and p.mag = d.mag and coalesce(p.otk, '') = d.otk and coalesce(p.serija, '') = d.serija and coalesce(p.roktr, 0) = d.roktr)`)
	_, err := r.exec(ctx, tx, "prazan lot", b)
	return err
}

// DrstaPostaviProsncSifre sets the prosečna nabavna cena of the zalihe by lot of the artikal (by its
// šifra) in a magacin (mag).
func (r *RobnoRepository) DrstaPostaviProsncSifre(ctx context.Context, mag, sifra int, prosnc float64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("drsta").Set("prosnc = ?", prosnc).
		Where("god = ?", s.God).Where("kar = ?", s.Kar).Where("mag = ?", mag).
		Where("rsifid = (select a.rsifid from rsif a where a.god = ? and a.kar = ? and a.sifra = ? limit 1)", s.God, s.Kar, sifra)
	_, err = r.exec(ctx, nil, "prosečna nabavna cena zaliha", b)
	return err
}

// RdokDodajStavku counts one stavka more in a document; ErrNotFound when it is not a document of
// the period.
func (r *RobnoRepository) RdokDodajStavku(ctx context.Context, tx db.Transaction, rdokID int64) error {
	b := querybuilder.NewUpdate("rdok").Set("brst = coalesce(brst, 0) + 1").Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stavke dokumenta", b)
}
