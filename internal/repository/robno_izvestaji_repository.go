package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
)

// The reports and calculations of the robno module that read the documents of the period.

// Ulaz/izlaz of the promet RUC report (domain.RobnoPrometParams.UlazIzlaz).
const (
	PrometUlaz      = "ulaz"
	PrometIzlaz     = "izlaz"
	PrometUlazIzlaz = "ulazizlaz"
)

// PrometUlazIzlaz returns, per magacin and artikal (or per artikal with zbirMagacina), the promet
// of the chosen ulaz/izlaz before and in the period od-do datuma, and the stanje, in the period of
// the session. ulazIzlaz is PrometUlaz, PrometIzlaz or anything else for both.
func (r *RobnoRepository) PrometUlazIzlaz(ctx context.Context, p domain.RobnoPrometParams) (*[]domain.RobnoPrometDto, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// The kodknj of the stavke of the chosen promet and the sign of their količina and magacinska
	// vrednost: for the ulaz i izlaz the izlazi subtract; the fakturna vrednost and the rabat are summed
	// only for the ulaz or the izlaz.
	kodknj, sign, fakturna := "('D', 'P')", "case when upper(dokvrsta.kodknj) = 'P' then -1 else 1 end", "0"
	switch p.UlazIzlaz {
	case PrometUlaz:
		kodknj, sign, fakturna = "('D')", "1", "1"
	case PrometIzlaz:
		kodknj, sign, fakturna = "('P')", "1", "1"
	}
	pocetak := fmt.Sprintf("%d-01-01", s.God) // the first day of the year
	// The key of the rows: the magacin and the article, or the article for the sum of the magacini.
	key, magacin, magaciniID := "s.mag, s.sifra", "s.mag", "min(s.magaciniid)"
	stanje := `(select coalesce(sum(coalesce(r.ulaz, 0) - coalesce(r.izlaz, 0)), 0) from rsta r
			where r.god = ? and r.kar = ? and r.magaciniid = agg.magaciniid and r.sifra = agg.sifra_art)`
	stanjeArgs := []any{s.God, s.Kar}
	if p.ZbirMagacina {
		key, magacin, magaciniID = "s.sifra", "0", "0"
		stanje = `(select coalesce(sum(coalesce(r.ulaz, 0) - coalesce(r.izlaz, 0)), 0) from rsta r
			where r.god = ? and r.kar = ? and r.sifra = agg.sifra_art and r.mag >= ? and r.mag <= ?)`
		stanjeArgs = []any{s.God, s.Kar, p.OdMagacina, p.DoMagacina}
	}
	// The stavke before the period (danal < odDatuma) and of the period (danal >= odDatuma): every
	// use is its own placeholder, so the args list it once per use, in the order of the text.
	before, period := "s.danal < ?", "s.danal >= ?"
	sumOf := func(cond, expr string) string {
		return "coalesce(sum(case when " + cond + " then " + expr + " end), 0)"
	}
	agg := `select ` + magacin + ` as mag, ` + magaciniID + ` as magaciniid, s.sifra as sifra_art,
			coalesce(min(s.naziv), '') as naziv,
			coalesce(min(s.jm), '') as jm,
			coalesce(min(s.pro), '') as pro,
			coalesce(min(s.proizsifra), '') as proizsifra,
			coalesce(min(s.komercopis), '') as komercopis,
			` + sumOf(before, "s.znak * s.kolic") + ` as koldodat,
			` + sumOf(period+" and s.kodknj = 'D'", "s.kolic") + ` as ulaz,
			` + sumOf(period+" and s.kodknj = 'P'", "s.kolic") + ` as izlaz,
			coalesce(sum(s.znak * s.kolic), 0) as kolnadan,
			` + sumOf(before, "s.fakvred") + ` as fakvreddo,
			` + sumOf(before, "s.rabat") + ` as rabatdo,
			` + sumOf(before, "s.znak * s.magvred") + ` as magvreddo,
			` + sumOf(period, "s.fakvred") + ` as fakvrednost,
			` + sumOf(period, "s.rabat") + ` as rabat,
			` + sumOf(period, "s.znak * s.magvred") + ` as magvrednost,
			coalesce(sum(s.fakvred), 0) as fakvrednostnadan,
			coalesce(sum(s.rabat), 0) as rabatnadan,
			coalesce(sum(s.znak * s.magvred), 0) as magvrednostnadan
		from stavke s
		group by ` + key
	aggArgs := make([]any, 9)
	for i := range aggArgs {
		aggArgs[i] = p.OdDatuma
	}
	b := querybuilder.New(`select agg.*, `+stanje+` as stanje from agg`, stanjeArgs...).
		WithCTE("stavke", `select rdok.mag, rdok.magaciniid, rdok.danal, rpro.sifra,
				rsif.naziv, rsif.jm, rsif.pro, rsif.proizsifra, rsif.komercopis,
				upper(dokvrsta.kodknj) as kodknj,
				`+sign+` as znak,
				coalesce(rpro.kolic, 0) as kolic,
				round(coalesce(rpro.kolic, 0) * coalesce(rpro.cena, 0), 2) as magvred,
				`+fakturna+` * round(coalesce(rpro.kolic, 0) * coalesce(rpro.fcena, 0), 2) as fakvred,
				`+fakturna+` * round(coalesce(rpro.kolic, 0) * coalesce(rpro.fcena, 0) * coalesce(rpro.rab, 0) / 100, 2) as rabat
			from rdok
			inner join rpro on rpro.rdokid = rdok.rdokid
			inner join rsif on rsif.rsifid = rpro.rsifid
			inner join dokvrsta on dokvrsta.god = rdok.god and dokvrsta.kar = rdok.kar and dokvrsta.vrd = rdok.vrd
			where rpro.god = ? and rpro.kar = ?
				and rdok.mag >= ? and rdok.mag <= ?
				and rdok.danal >= ? and rdok.danal <= ?
				and upper(dokvrsta.kodknj) in `+kodknj+`
				and rpro.sifra > 0`, s.God, s.Kar, p.OdMagacina, p.DoMagacina, pocetak, p.DoDatuma).
		WithCTE("agg", agg, aggArgs...).
		OrderBy("agg.sifra_art, agg.mag")
	var rows []domain.RobnoPrometDto
	if err := r.selectRows(ctx, "promet ulaz/izlaz", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// VrdGrupeNaloga returns the vrste dokumenata (vrd) of the period whose dokvrsta.grpnal holds the
// vrsta naloga tipdok (one, e.g. "70", or a list, "70,71,").
func (r *RobnoRepository) VrdGrupeNaloga(ctx context.Context, tipdok string) ([]int, error) {
	b := querybuilder.New(`select vrd from dokvrsta d`).
		Where("exists (select 1 from regexp_split_to_table(coalesce(d.grpnal, ''), '[,;]') g where trim(g) = ?)", strings.TrimSpace(tipdok)).
		OrderBy("vrd")
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	var vrd []int
	err := r.selectRows(ctx, "vrste dokumenata naloga", &vrd, b)
	return vrd, err
}

// NcenaPromet returns the stavke of an artikal in a magacin up to the datum of the nalog and of the
// document, in the period of the session, for the calculation of the nabavna cena: key = vrd,
// c1 = količina, c2 = cena, c3 = nabavna cena, c4 = kodknj of the vrsta.
func (r *RobnoRepository) NcenaPromet(ctx context.Context, sifra, magaciniID int, danal, dadok time.Time) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select d.vrd::text as key, coalesce(p.kolic, 0)::float8::text as c1,
			coalesce(p.cena, 0)::float8::text as c2, coalesce(p.ncena, 0)::float8::text as c3,
			upper(trim(coalesce(v.kodknj, ''))) as c4
		from rdok d`).
		AddJoin("join rpro p on p.rdokid = d.rdokid").
		AddJoin("left join dokvrsta v on v.god = d.god and v.kar = d.kar and v.vrd = d.vrd").
		Where("d.vrd <> 119 and d.vrd <> 133")
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	b.Where("p.sifra = ?", sifra).
		Where("d.magaciniid = ?", magaciniID).
		Where("d.danal <= ?", danal).
		Where("d.dadok <= ?", dadok).
		OrderBy("d.danal asc, d.dadok asc, p.rproid")
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "promet artikla", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// NcenaPrviUlaz returns the nabavna cena of the first ulaz of an artikal in a magacin from the
// datum of the nalog, in the period of the session (key = nabavna cena; no row when none).
func (r *RobnoRepository) NcenaPrviUlaz(ctx context.Context, sifra, magaciniID int, danal time.Time) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select coalesce(p.ncena, 0)::float8::text as key from rdok d`).
		AddJoin("join rpro p on p.rdokid = d.rdokid")
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	b.Where("d.danal >= ?", danal).
		Where("d.vrd in (101, 110, 113, 117, 121, 129, 198, 131, 189)").
		Where("p.sifra = ?", sifra).
		Where("d.magaciniid = ?", magaciniID).
		OrderBy("d.danal asc").
		Limit(1)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "prvi ulaz artikla", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}
