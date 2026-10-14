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

// The artikli (rsif), the stanja (rsta) and the zalihe by lot (drsta) of the period of the session.

// RsifArtikal returns the id, the naziv and the JM of an artikal by its šifra; found is false when
// there is none.
func (r *RobnoRepository) RsifArtikal(ctx context.Context, sifra int) (rsifID int64, naziv, jm string, found bool, err error) {
	b := querybuilder.New(`select a.rsifid, coalesce(a.naziv, ''), coalesce(a.jm, '') from rsif a`)
	if err := inPeriod(ctx, b, "a."); err != nil {
		return 0, "", "", false, err
	}
	b.Where("a.sifra = ?", sifra).Limit(1)
	err = r.row(ctx, nil, "artikal", b, &rsifID, &naziv, &jm)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", "", false, nil
	}
	return rsifID, naziv, jm, err == nil, err
}

// RsifIDPo returns the id and the poreska oznaka of an artikal by its šifra; found is false when
// there is none. tx may be nil.
func (r *RobnoRepository) RsifIDPo(ctx context.Context, tx db.Transaction, sifra int) (rsifID int64, po int, found bool, err error) {
	b := querybuilder.New(`select rsifid, coalesce(po, 0) from rsif`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, false, err
	}
	b.Where("sifra = ?", sifra).Limit(1)
	err = r.row(ctx, tx, "artikal", b, &rsifID, &po)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return rsifID, po, err == nil, err
}

// RstaStanjeVpcena returns the stanje and the veleprodajna cena of an artikal in a konto of a
// magacin; found is false when there is no stanje.
func (r *RobnoRepository) RstaStanjeVpcena(ctx context.Context, magaciniID int, konto string, sifra int) (stanje, vpcena float64, found bool, err error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8, coalesce(vpcena, 0)::float8 from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, false, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", strings.TrimSpace(konto)).Where("sifra = ?", sifra).
		OrderBy("rstaid").Limit(1)
	err = r.row(ctx, nil, "stanje", b, &stanje, &vpcena)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return stanje, vpcena, err == nil, err
}

// RstaPostaviVpcena sets the veleprodajna cena of the stanje of an artikal in a konto of a
// magacin (no row when there is no stanje).
func (r *RobnoRepository) RstaPostaviVpcena(ctx context.Context, tx db.Transaction, magaciniID int, konto string, sifra int, vpcena float64) error {
	b := querybuilder.NewUpdate("rsta").Set("vpcena = ?", vpcena)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", konto).Where("sifra = ?", sifra)
	_, err := r.exec(ctx, tx, "veleprodajna cena stanja", b)
	return err
}

// DrstaStanjeArtikla returns the sum of the positive zalihe by lot of an artikal.
func (r *RobnoRepository) DrstaStanjeArtikla(ctx context.Context, rsifID int64) (float64, error) {
	b := querybuilder.New(`select coalesce(sum(coalesce(ulaz, 0) - coalesce(izlaz, 0)), 0)::float8 from drsta`).
		Where("rsifid = ?", rsifID).
		Where("coalesce(ulaz, 0) - coalesce(izlaz, 0) > 0")
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var stanje float64
	err := r.row(ctx, nil, "zalihe artikla", b, &stanje)
	return stanje, err
}

// DrstaLotovi returns the zalihe by lot (stanje >= 0) of an artikal in a magacin, in the order of
// their ids: key = stanje, c1 = OTK broj, c2 = serija, c3 = rok trajanja.
func (r *RobnoRepository) DrstaLotovi(ctx context.Context, mag int, rsifID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8::text as key, coalesce(otk, '') as c1,
		coalesce(serija, '') as c2, coalesce(roktr, 0)::bigint::text as c3 from drsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return nil, err
	}
	b.Where("mag = ?", mag).Where("rsifid = ?", rsifID).Where("coalesce(ulaz, 0) - coalesce(izlaz, 0) >= 0").OrderBy("rstaid")
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "zalihe po lotu", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// DrstaPostaviCene sets the cena and the veleprodajna cena of the zalihe by lot of an artikal in a
// magacin.
func (r *RobnoRepository) DrstaPostaviCene(ctx context.Context, tx db.Transaction, magaciniID, sifra int, cena float64) error {
	b := querybuilder.NewUpdate("drsta").Set("cena = ?, vpcena = ?", cena, cena)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	b.Where("magaciniid = ?", magaciniID).Where("sifra = ?", sifra)
	_, err := r.exec(ctx, tx, "cene zaliha po lotu", b)
	return err
}

// RsifArtikalSaCenama returns the artikal of a šifra (as text) with the cene of its first rcene:
// key = šifra, c1 = naziv, c2 = JM, c3 = tip, c4 = cena, c5 = nabavna cena; no row when there is none.
func (r *RobnoRepository) RsifArtikalSaCenama(ctx context.Context, sifra string) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select a.sifra::text as key, coalesce(a.naziv, '') as c1, coalesce(a.jm, '') as c2,
			upper(trim(coalesce(a.tip, ''))) as c3,
			coalesce((select c.cena from rcene c where c.god = a.god and c.kar = a.kar and c.sifra = a.sifra order by c.rceneid limit 1), 0)::float8::text as c4,
			coalesce((select c.ncena from rcene c where c.god = a.god and c.kar = a.kar and c.sifra = a.sifra order by c.rceneid limit 1), 0)::float8::text as c5
		from rsif a`)
	if err := inPeriod(ctx, b, "a."); err != nil {
		return nil, err
	}
	b.Where("a.sifra::text = ?", strings.TrimSpace(sifra)).Limit(1)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "artikal", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// RsifArtikalPoTekstu returns the šifra, the naziv and the JM of an artikal by its šifra as text;
// found is false when there is none.
func (r *RobnoRepository) RsifArtikalPoTekstu(ctx context.Context, sifra string) (sif int, naziv, jm string, found bool, err error) {
	b := querybuilder.New(`select a.sifra, coalesce(a.naziv, ''), coalesce(a.jm, '') from rsif a`)
	if err := inPeriod(ctx, b, "a."); err != nil {
		return 0, "", "", false, err
	}
	b.Where("a.sifra::text = ?", strings.TrimSpace(sifra)).Limit(1)
	err = r.row(ctx, nil, "artikal", b, &sif, &naziv, &jm)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", "", false, nil
	}
	return sif, naziv, jm, err == nil, err
}

// RsifNula returns the id and the JM of the artikal of the šifra 0 (the first one); the error wraps
// sql.ErrNoRows when there is none.
func (r *RobnoRepository) RsifNula(ctx context.Context, tx db.Transaction) (rsifID int64, jm string, err error) {
	b := querybuilder.New(`select rsifid, coalesce(jm, '') from rsif`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, "", err
	}
	b.Where("sifra = 0").OrderBy("rsifid").Limit(1)
	err = r.row(ctx, tx, "artikal šifre 0", b, &rsifID, &jm)
	return rsifID, jm, err
}

// RceneNulaPostoji tells whether the artikal of the šifra 0 has its cene.
func (r *RobnoRepository) RceneNulaPostoji(ctx context.Context, tx db.Transaction) (bool, error) {
	b := querybuilder.New(`select count(*) from rcene`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	b.Where("sifra = 0")
	var n int
	if err := r.row(ctx, tx, "cene artikla šifre 0", b, &n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// rsifKolone are the columns of rsif from the catalog (no identity column), in their order:
// key = name, c1 = data type, c2 = is_nullable (YES/NO), c3 = the table.column of its foreign key
// ("" for none), c4 = it has a default (true/false).
func (r *RobnoRepository) rsifKolone(ctx context.Context) ([]domain.SearchRowDto, error) {
	b := querybuilder.New(`select c.column_name as key, c.data_type as c1, c.is_nullable as c2,
			coalesce((select f.confrelid::regclass::text || '.' || r.attname from pg_constraint f
				join pg_attribute a on a.attrelid = f.conrelid and a.attnum = f.conkey[1]
				join pg_attribute r on r.attrelid = f.confrelid and r.attnum = f.confkey[1]
				where f.conrelid = 'rsif'::regclass and f.contype = 'f' and a.attname = c.column_name limit 1), '') as c3,
			(c.column_default is not null)::text as c4
		from information_schema.columns c`).
		Where("c.table_schema = current_schema()").Where("c.table_name = 'rsif'").Where("c.is_identity = 'NO'").
		OrderBy("c.ordinal_position")
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "kolone artikla", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// RsifKreirajNulu creates the artikal of the šifra 0 (a usluga, tip U, with the naziv) and returns
// its rsifid. rsif has many columns without a default: the artikal copies the keys of the šifarnici
// (its foreign keys: the one of the first artikal of the period when it exists in the šifarnik, else
// the first one of the šifarnik), its other columns get their default or are empty ("", 0, false;
// NULL when allowed). The error wraps sql.ErrNoRows when the period has no artikal to copy. The
// names come from the catalog; each must be a plain identifier and is quoted.
func (r *RobnoRepository) RsifKreirajNulu(ctx context.Context, tx db.Transaction, naziv string) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	kolone, err := r.rsifKolone(ctx)
	if err != nil {
		return 0, err
	}
	if len(kolone) == 0 {
		return 0, errors.New("rsif nema kolone")
	}
	names, izrazi := make([]string, 0, len(kolone)), make([]string, 0, len(kolone))
	var args []any
	for _, k := range kolone {
		if !querybuilder.IsIdent(k.Key) || strings.Contains(k.Key, ".") {
			return 0, fmt.Errorf("artikal šifre 0: kolona %q", k.Key)
		}
		kol := querybuilder.QuoteIdent(k.Key)
		izraz, arg, hasArg := "", any(nil), false
		switch k.Key {
		case "rsifid":
			continue
		case "god":
			izraz, arg, hasArg = "?", s.God, true
		case "kar":
			izraz, arg, hasArg = "?", s.Kar, true
		case "sifra":
			izraz = "0"
		case "tip":
			izraz = "'U'"
		case "naziv":
			izraz, arg, hasArg = "?", naziv, true
		case "xopunos":
			izraz, arg, hasArg = "?", s.UserName, true
		case "xdatunosa":
			izraz = "now()"
		default:
			switch {
			case k.C3 != "":
				// A key of a šifarnik (also with a default: its default may not exist).
				izraz = "rsif." + kol
			case k.C4 == "true":
				// The default of the column.
				continue
			case k.C2 == "YES":
				izraz = "null"
			case strings.Contains(k.C1, "char") || k.C1 == "text":
				izraz = "''"
			case k.C1 == "boolean":
				izraz = "false"
			case strings.Contains(k.C1, "date") || strings.Contains(k.C1, "time"):
				izraz = "rsif." + kol
			default:
				izraz = "0"
			}
		}
		// A key of a šifarnik: the one of the copied artikal when it exists in the šifarnik, else the
		// first one of the šifarnik.
		if ref := strings.SplitN(k.C3, ".", 2); k.C3 != "" && len(ref) == 2 {
			if !querybuilder.IsIdent(ref[0]) || !querybuilder.IsIdent(ref[1]) || strings.Contains(ref[1], ".") {
				return 0, fmt.Errorf("artikal šifre 0: šifarnik %q", k.C3)
			}
			tab, col := querybuilder.QuoteIdent(ref[0]), querybuilder.QuoteIdent(ref[1])
			izraz, hasArg = "coalesce((select x."+col+" from "+tab+" x where x."+col+" = rsif."+kol+" limit 1), (select min(x."+col+") from "+tab+" x))", false
		}
		names, izrazi = append(names, kol), append(izrazi, izraz)
		if hasArg {
			args = append(args, arg)
		}
	}
	b := querybuilder.New(`insert into rsif (`+strings.Join(names, ", ")+`) select `+strings.Join(izrazi, ", ")+` from rsif`, args...).
		Where("rsif.god = ?", s.God).Where("rsif.kar = ?", s.Kar).OrderBy("rsif.rsifid").Limit(1).Returning("rsifid")
	var rsifID int64
	err = r.row(ctx, tx, "artikal šifre 0", b, &rsifID)
	return rsifID, err
}

// RstaCene is the stanje of an artikal in a konto of a magacin with its cene.
type RstaCene struct {
	Stanje, Cena, Prosnc, Vma, Mma, Mcena, Mcenap float64
}

// RstaStanjeCene returns the stanje of an artikal in a konto (trimmed) of a magacin with its cena,
// prosečna nabavna cena, rabat (vma), marža (mma) and maloprodajne cene; found is false without it.
func (r *RobnoRepository) RstaStanjeCene(ctx context.Context, magaciniID int, konto string, sifra int) (RstaCene, bool, error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8,
			coalesce(cena, 0)::float8, coalesce(prosnc, 0)::float8, coalesce(vma, 0)::float8,
			coalesce(mma, 0)::float8, coalesce(mcena, 0)::float8, coalesce(mcenap, 0)::float8
		from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RstaCene{}, false, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", strings.TrimSpace(konto)).Where("sifra = ?", sifra).
		OrderBy("rstaid").Limit(1)
	var c RstaCene
	err := r.row(ctx, nil, "stanje", b, &c.Stanje, &c.Cena, &c.Prosnc, &c.Vma, &c.Mma, &c.Mcena, &c.Mcenap)
	if errors.Is(err, sql.ErrNoRows) {
		return RstaCene{}, false, nil
	}
	return c, err == nil, err
}

// RsifArtikalSaStopom returns the artikal by its šifra, or by its barkod (poBarkodu), with its
// poreska stopa on the datum: key = šifra, c1 = naziv, c2 = JM, c3 = poreska oznaka, c4 = model,
// c5 = stopa, c6 = taksa of its cene; no row when there is none.
func (r *RobnoRepository) RsifArtikalSaStopom(ctx context.Context, poBarkodu bool, vrednost string, datum any) (*[]domain.SearchRowDto, error) {
	kolona := "a.sifra::text"
	if poBarkodu {
		kolona = "a.barkod"
	}
	b := querybuilder.New(`select a.sifra::text as key, coalesce(a.naziv, '') as c1,
			coalesce(a.jm, '') as c2, coalesce(a.po, 0)::text as c3, coalesce(a.model, '') as c4,
			`+poreskaStopa("a.po", "?::date")+`::text as c5,
			coalesce((select c.itaksa from rcene c where c.god = a.god and c.kar = a.kar and c.sifra = a.sifra limit 1), 0)::float8::text as c6
		from rsif a`, datum)
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

// RsifIDNazivJm returns the id, the naziv and the JM of an artikal by its šifra; the error wraps
// sql.ErrNoRows when there is none.
func (r *RobnoRepository) RsifIDNazivJm(ctx context.Context, tx db.Transaction, sifra int) (rsifID int64, naziv, jm string, err error) {
	b := querybuilder.New(`select rsifid, coalesce(naziv, ''), coalesce(jm, '') from rsif`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, "", "", err
	}
	b.Where("sifra = ?", sifra)
	err = r.row(ctx, tx, "artikal", b, &rsifID, &naziv, &jm)
	return rsifID, naziv, jm, err
}

// RcenePostoji tells whether the artikal with the šifra has its cene.
func (r *RobnoRepository) RcenePostoji(ctx context.Context, tx db.Transaction, sifra int) (bool, error) {
	b := querybuilder.New(`select count(*) from rcene`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	b.Where("sifra = ?", sifra)
	var n int
	if err := r.row(ctx, tx, "cene artikla", b, &n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// RstaStanjeNcena returns the stanje of an artikal in a konto (trimmed) of a magacin and its
// nabavna cena; found is false without it.
func (r *RobnoRepository) RstaStanjeNcena(ctx context.Context, magaciniID int, konto string, sifra int) (stanje, ncena float64, found bool, err error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8, coalesce(ncena, 0)::float8 from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, false, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", strings.TrimSpace(konto)).Where("sifra = ?", sifra).
		OrderBy("rstaid").Limit(1)
	err = r.row(ctx, nil, "stanje", b, &stanje, &ncena)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return stanje, ncena, err == nil, err
}

// RsifArtikalSaNcenom returns the artikal by its šifra, or by its barkod (poBarkodu), with its
// poreska stopa on the datum: key = šifra, c1 = naziv, c2 = JM, c3 = poreska oznaka, c4 = stopa,
// c5 = nabavna cena of its cene; no row when there is none.
func (r *RobnoRepository) RsifArtikalSaNcenom(ctx context.Context, poBarkodu bool, vrednost string, datum any) (*[]domain.SearchRowDto, error) {
	kolona := "a.sifra::text"
	if poBarkodu {
		kolona = "a.barkod"
	}
	b := querybuilder.New(`select a.sifra::text as key, coalesce(a.naziv, '') as c1,
			coalesce(a.jm, '') as c2, coalesce(a.po, 0)::text as c3,
			`+poreskaStopa("a.po", "?::date")+`::text as c4,
			coalesce((select c.ncena from rcene c where c.god = a.god and c.kar = a.kar and c.sifra = a.sifra limit 1), 0)::float8::text as c5
		from rsif a`, datum)
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
