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

// The fakture usluga (and the knjižna pisma) with their stavke and the dodatni opis usluge, in the
// period of the session.

// MagacinZalihe returns how a magacin keeps its zalihe: its tip zaliha and its način vođenja
// zaliha; found is false when it is not a magacin of the period.
func (r *RobnoRepository) MagacinZalihe(ctx context.Context, magaciniID int) (tipzal, nacvodzal int, found bool, err error) {
	b := querybuilder.New(`select coalesce(tipzal, 0), coalesce(nacvodzal, 0) from magacini`).Where("magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, false, err
	}
	err = r.row(ctx, nil, "zalihe magacina", b, &tipzal, &nacvodzal)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return tipzal, nacvodzal, err == nil, err
}

// MagkontoPrviKonto returns the first konto (trimmed, not empty) of a magacin; "" without it.
func (r *RobnoRepository) MagkontoPrviKonto(ctx context.Context, magaciniID int) (string, error) {
	b := querybuilder.New(`select trim(konto) from magkonto`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	b.Where("magaciniid = ?", magaciniID).Where("coalesce(trim(konto), '') <> ''").OrderBy("magkontoid").Limit(1)
	var konto string
	err := r.row(ctx, nil, "konto magacina", b, &konto)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return konto, err
}

// RdokIzvorniGodine returns the datum, the SEF id and the SEF status of the document of the year god
// (of the kar of the session) by its magacin (mag), vrsta and broj: the izvorni dokument of a
// knjižno pismo, which may be of another year than the session. The error wraps sql.ErrNoRows when
// there is none.
func (r *RobnoRepository) RdokIzvorniGodine(ctx context.Context, god, mag, vrd, dokum int) (dadok sql.NullTime, sefID float64, status string, err error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return dadok, 0, "", err
	}
	b := querybuilder.New(`select dadok, coalesce(salesinvoiceid, 0)::float8, coalesce(status_salinv, '') from rdok`).
		Where("god = ?", god).Where("kar = ?", s.Kar).Where("mag = ?", mag).Where("vrd = ?", vrd).Where("dokum = ?", dokum).
		OrderBy("rdokid").Limit(1)
	err = r.row(ctx, nil, "izvorni dokument", b, &dadok, &sefID, &status)
	return dadok, sefID, status, err
}

// RsifArtikalUsluge returns the artikal by its šifra, or by its barkod (poBarkodu), for a stavka of
// a faktura usluga: key = šifra, c1 = naziv, c2 = JM, c3 = poreska oznaka, c4 = tip, c5 = konto
// (trimmed), c6 = its stopa on the datum, c7 = the cena of its cene; no row when there is none.
func (r *RobnoRepository) RsifArtikalUsluge(ctx context.Context, poBarkodu bool, vrednost string, datum any) (*[]domain.SearchRowDto, error) {
	kolona := "a.sifra::text"
	if poBarkodu {
		kolona = "a.barkod"
	}
	b := querybuilder.New(`select a.sifra::text as key, coalesce(a.naziv, '') as c1,
			coalesce(a.jm, '') as c2, coalesce(a.po, 0)::text as c3, coalesce(a.tip, '') as c4, trim(coalesce(a.konto, '')) as c5,
			`+poreskaStopa("a.po", "?::date")+`::text as c6,
			coalesce((select c.cena from rcene c where c.god = a.god and c.kar = a.kar and c.sifra = a.sifra limit 1), 0)::float8::text as c7
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

// RstaUsluge is the stanje of an artikal with the cene a stavka of a faktura usluga uses.
type RstaUsluge struct {
	Stanje, Prosnc, Vpcena, Cena float64
}

// RstaStanjeUsluge returns the stanje of an artikal in a konto of a magacin with its prosečna
// nabavna, veleprodajna and magacinska cena; found is false without it.
func (r *RobnoRepository) RstaStanjeUsluge(ctx context.Context, magaciniID int, konto string, sifra int) (RstaUsluge, bool, error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8,
			coalesce(prosnc, 0)::float8, coalesce(vpcena, 0)::float8, coalesce(cena, 0)::float8
		from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RstaUsluge{}, false, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", konto).Where("sifra = ?", sifra).OrderBy("rstaid").Limit(1)
	var st RstaUsluge
	err := r.row(ctx, nil, "stanje", b, &st.Stanje, &st.Prosnc, &st.Vpcena, &st.Cena)
	if errors.Is(err, sql.ErrNoRows) {
		return RstaUsluge{}, false, nil
	}
	return st, err == nil, err
}

// RdokStopaUsluge returns the stopa of the tarifa on the datum of a document; for a document with
// the stari PDV (flgstaripdv) the stopa of the day before the last change of the stopa of the
// tarifa. found is false when it is not a document of the period.
func (r *RobnoRepository) RdokStopaUsluge(ctx context.Context, rdokID int64, tarifa int) (stopa float64, found bool, err error) {
	b := querybuilder.New(`select (case when coalesce(d.flgstaripdv, false) then
			coalesce((select x.pp from rpor x where x.po = ? and x.datum <= (select max(y.datum) from rpor y where y.po = ?) - 1
				order by x.datum desc limit 1), 0)
		else `+poreskaStopa("?", "d.dadok")+` end)::float8
		from rdok d`, tarifa, tarifa, tarifa).Where("d.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "d."); err != nil {
		return 0, false, err
	}
	err = r.row(ctx, nil, "stopa tarife", b, &stopa)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return stopa, err == nil, err
}

// RsifTipStanje returns the tip of an artikal and its stanje in a konto (trimmed) of a magacin (0
// without it); found is false when the artikal does not exist.
func (r *RobnoRepository) RsifTipStanje(ctx context.Context, sifra, magaciniID int, konto string) (tip string, stanje float64, found bool, err error) {
	b := querybuilder.New(`select coalesce(a.tip, ''),
			coalesce((select (coalesce(r.ulaz, 0) - coalesce(r.izlaz, 0))::float8 from rsta r where r.god = a.god and r.kar = a.kar
				and r.magaciniid = ? and r.konto = ? and r.sifra = a.sifra order by r.rstaid limit 1), 0)::float8
		from rsif a`, magaciniID, strings.TrimSpace(konto))
	if err := inPeriod(ctx, b, "a."); err != nil {
		return "", 0, false, err
	}
	b.Where("a.sifra = ?", sifra).Limit(1)
	err = r.row(ctx, nil, "stanje artikla", b, &tip, &stanje)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, false, nil
	}
	return strings.TrimSpace(tip), stanje, err == nil, err
}

// RsifIDTip returns the id and the tip of an artikal by its šifra; the error wraps sql.ErrNoRows
// when there is none.
func (r *RobnoRepository) RsifIDTip(ctx context.Context, tx db.Transaction, sifra int) (rsifID int64, tip string, err error) {
	b := querybuilder.New(`select rsifid, coalesce(tip, '') from rsif`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, "", err
	}
	b.Where("sifra = ?", sifra).Limit(1)
	err = r.row(ctx, tx, "artikal", b, &rsifID, &tip)
	return rsifID, tip, err
}

// DokusldetSnimi sets the dodatni opis usluge of a document: it changes the one it has, else adds it.
func (r *RobnoRepository) DokusldetSnimi(ctx context.Context, tx db.Transaction, rdokID int64, opis string) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	c := querybuilder.New(`select count(*) from dokusldet`).Where("rdokid = ?", rdokID).Where("god = ?", s.God).Where("kar = ?", s.Kar)
	var n int
	if err := r.row(ctx, tx, "opis usluge", c, &n); err != nil {
		return err
	}
	b := querybuilder.NewUpdate("dokusldet").Set("opisdet = ?", opis).
		Where("rdokid = ?", rdokID).Where("god = ?", s.God).Where("kar = ?", s.Kar)
	if n == 0 {
		b = querybuilder.New(`insert into dokusldet (god, kar, edokid, rdokid, opisdet) values (?, ?, 0, ?, ?)`, s.God, s.Kar, rdokID, opis)
	}
	_, err = r.exec(ctx, tx, "opis usluge", b)
	return err
}

// RproUsluga is a stavka of a faktura usluga to delete.
type RproUsluga struct {
	RdokID                 int64
	Sifra, Rbr             int
	Konto, Otk, Serija     string
	Tip                    string
	Kolicina, Iznos, Roktr float64
}

// RproUslugaBrisanje reads (and locks) a stavka of a faktura usluga to delete, with the tip of its
// artikal; the error wraps sql.ErrNoRows when it is not a stavka of the period.
func (r *RobnoRepository) RproUslugaBrisanje(ctx context.Context, tx db.Transaction, rproID int64) (RproUsluga, error) {
	b := querybuilder.New(`select coalesce(p.rdokid, 0)::bigint, coalesce(p.sifra, 0), coalesce(p.konto, ''),
			coalesce(p.kolic, 0)::float8, coalesce(p.iznos, 0)::float8, coalesce(p.otk, ''), coalesce(p.serija, ''),
			coalesce(p.roktr, 0)::float8, coalesce(p.rbr, 0), coalesce((select a.tip from rsif a where a.rsifid = p.rsifid), '')
		from rpro p`).Where("p.rproid = ?", rproID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return RproUsluga{}, err
	}
	b.ForUpdate()
	var st RproUsluga
	err := r.row(ctx, tx, "stavka", b, &st.RdokID, &st.Sifra, &st.Konto, &st.Kolicina, &st.Iznos, &st.Otk, &st.Serija, &st.Roktr, &st.Rbr, &st.Tip)
	return st, err
}

// RdokOduzmiIznos takes an iznos from a document (rounded to 2 decimals), its number of stavke
// stays; ErrNotFound when it is not a document of the period.
func (r *RobnoRepository) RdokOduzmiIznos(ctx context.Context, tx db.Transaction, rdokID int64, iznos float64) error {
	b := querybuilder.NewUpdate("rdok").Set("iznos = round((coalesce(iznos, 0) - ?)::numeric, 2)", iznos).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "iznos dokumenta", b)
}

// RproUslugaForma returns a saved stavka of a faktura usluga for its form: key = šifra, c1 = naziv,
// c2 = konto|JM|poreska kategorija|osnov oslobođenja|OTK broj|serija, c3 = poreska oznaka, c4 = its
// stopa on the datum of the document, c5 = |količina|fakturna cena|rabat|nabavna cena,
// c6 = rok trajanja, c7 = the dodatni opis usluge of the document.
func (r *RobnoRepository) RproUslugaForma(ctx context.Context, rproID, rdokID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select coalesce(r.sifra, 0)::text as key, coalesce(r.naz1, '') as c1,
			(coalesce(r.konto, '') || '|' || coalesce(r.jm, '') || '|' || coalesce(r.taxcat, '') || '|' || coalesce(r.taxexreco, '') || '|' ||
				coalesce(r.otk, '') || '|' || coalesce(r.serija, '')) as c2,
			coalesce(r.po, 0)::text as c3, `+poreskaStopa("r.po", "d.dadok")+`::text as c4,
			(abs(coalesce(r.kolic, 0))::float8::text || '|' || coalesce(r.fcena, 0)::float8::text || '|' || coalesce(r.rab, 0)::float8::text || '|' ||
				coalesce(r.ncena, 0)::float8::text) as c5,
			coalesce(r.roktr, 0)::bigint::text as c6,
			coalesce((select u.opisdet from dokusldet u where u.rdokid = r.rdokid order by u.dokusldetid limit 1), '') as c7
		from rpro r`).
		AddJoin("join rdok d on d.rdokid = r.rdokid").
		Where("r.rproid = ?", rproID).Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "stavka", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// RdokVrdStavke returns the vrsta of the document of a stavka; found is false when it is not a
// stavka of a document of the period.
func (r *RobnoRepository) RdokVrdStavke(ctx context.Context, rproID int64) (vrd int, found bool, err error) {
	b := querybuilder.New(`select coalesce(vrd, 0) from rdok`).Where("rdokid = (select rdokid from rpro where rproid = ?)", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, false, err
	}
	err = r.row(ctx, nil, "vrsta dokumenta stavke", b, &vrd)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return vrd, err == nil, err
}

// DokvrstaVrsteGrupe returns the vrste of the documents of the group of the vrsta vrd (grpdok)
// between vrdOd and vrdDo, in their order.
func (r *RobnoRepository) DokvrstaVrsteGrupe(ctx context.Context, vrd, vrdOd, vrdDo int) ([]int, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select v.vrd from dokvrsta v`).
		Where("v.god = ?", s.God).Where("v.kar = ?", s.Kar).Where("v.vrd between ? and ?", vrdOd, vrdDo).
		Where(`upper(trim(coalesce(v.grpdok, ''))) = (select upper(trim(coalesce(g.grpdok, ''))) from dokvrsta g
			where g.god = ? and g.kar = ? and g.vrd = ? limit 1)`, s.God, s.Kar, vrd).
		OrderBy("v.vrd")
	var vrste []int
	if err := r.selectRows(ctx, "vrste grupe", &vrste, b); err != nil {
		return nil, err
	}
	return vrste, nil
}
