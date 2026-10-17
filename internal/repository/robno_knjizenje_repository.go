package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The posting of the stavke of the robni dokumenti: the stanje by konto (rsta), the zalihe by lot
// (drsta), the cene of the artikli (rcene), the nalog and the document, in the period of the session.

// Izmena is a change of a column: Dodaj adds the value to it (coalesce(column, 0) + value), else the
// column gets the value. NeNegativno keeps the sum of Dodaj at 0 or more (greatest(..., 0), the legacy
// "if < 0 then 0" of the neraspoložive zalihe). The column must be one of the table (see the allow-lists
// below).
type Izmena struct {
	Kolona      string
	Dodaj       bool
	NeNegativno bool
	Vrednost    any
}

// StanjeTabela is a table of the stanje: rsta (by konto) or drsta (by lot).
type StanjeTabela string

// The tables of the stanje.
const (
	TabelaRsta  StanjeTabela = "rsta"
	TabelaDrsta StanjeTabela = "drsta"
)

// kolone are the columns the posting may write, per table.
var kolone = func() map[string]map[string]bool {
	stanje := map[string]bool{}
	for _, k := range []string{"god", "kar", "rsifid", "sifra", "konto", "mag", "magaciniid", "zdat", "brst",
		"cena", "fcena", "mcena", "ncena", "prosnc", "mcenap", "vma", "mma", "rab", "ztro", "ztrin", "tz", "tzi",
		"dug", "pot", "ulaz", "izlaz", "pov", "pos", "kolicner", "rez"} {
		stanje[k] = true
	}
	for m := 1; m <= 12; m++ {
		for _, k := range []string{"mdug", "mpot", "mul", "miz"} {
			stanje[fmt.Sprintf("%s_%d", k, m)] = true
		}
	}
	lot := map[string]bool{"otk": true, "serija": true, "roktr": true, "zstatus": true}
	for k := range stanje {
		lot[k] = true
	}
	rcene := map[string]bool{}
	for _, k := range []string{"cena", "mcenap", "pcena", "fcena", "ncena", "mcena", "rab", "vra", "mra", "vma", "mma"} {
		rcene[k] = true
	}
	return map[string]map[string]bool{"rsta": stanje, "drsta": lot, "rcene": rcene}
}()

// kolona returns the quoted column of a table; an error for a column the posting may not write.
func kolona(tabela, k string) (string, error) {
	if !kolone[tabela][k] {
		return "", fmt.Errorf("%s: kolona %q nije dozvoljena", tabela, k)
	}
	return querybuilder.QuoteIdent(k), nil
}

// setIzmene adds the changes to the SET of an update of the table.
func setIzmene(b *querybuilder.Builder, tabela string, izmene []Izmena) error {
	for _, iz := range izmene {
		k, err := kolona(tabela, iz.Kolona)
		if err != nil {
			return err
		}
		switch {
		case iz.Dodaj && iz.NeNegativno:
			b.Set(k+" = greatest(coalesce("+k+", 0) + ?, 0)", iz.Vrednost)
		case iz.Dodaj:
			b.Set(k+" = coalesce("+k+", 0) + ?", iz.Vrednost)
		default:
			b.Set(k+" = ?", iz.Vrednost)
		}
	}
	return nil
}

// StanjeDodaj adds a row of the stanje with the columns and their values (god and kar are the ones of
// the session).
func (r *RobnoRepository) StanjeDodaj(ctx context.Context, tx db.Transaction, tabela StanjeTabela, vrednosti map[string]any) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	v := make(map[string]any, len(vrednosti)+2)
	for k, x := range vrednosti {
		v[k] = x
	}
	v["god"], v["kar"] = s.God, s.Kar
	imena := make([]string, 0, len(v))
	for k := range v {
		imena = append(imena, k)
	}
	sort.Strings(imena)
	cols, args := make([]string, len(imena)), make([]any, len(imena))
	for i, k := range imena {
		if cols[i], err = kolona(string(tabela), k); err != nil {
			return err
		}
		args[i] = v[k]
	}
	b := querybuilder.New(`insert into `+string(tabela)+` (`+strings.Join(cols, ", ")+`) values (`+
		strings.TrimSuffix(strings.Repeat("?, ", len(cols)), ", ")+`)`, args...)
	_, err = r.exec(ctx, tx, "novo stanje", b)
	return err
}

// StanjeIzmeni changes a row of the stanje (rstaid); nothing without changes.
func (r *RobnoRepository) StanjeIzmeni(ctx context.Context, tx db.Transaction, tabela StanjeTabela, rstaID int64, izmene []Izmena) error {
	if len(izmene) == 0 {
		return nil
	}
	b := querybuilder.NewUpdate(string(tabela))
	if err := setIzmene(b, string(tabela), izmene); err != nil {
		return err
	}
	b.Where("rstaid = ?", rstaID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stanje", b)
}

// RstaIDZaIzmenu reads (and locks) the stanje of an artikal in a konto of a magacin; the error wraps
// sql.ErrNoRows without it.
func (r *RobnoRepository) RstaIDZaIzmenu(ctx context.Context, tx db.Transaction, magaciniID int, konto string, sifra int) (int64, error) {
	b := querybuilder.New(`select rstaid from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", konto).Where("sifra = ?", sifra).OrderBy("rstaid").Limit(1).ForUpdate()
	var id int64
	err := r.row(ctx, tx, "stanje", b, &id)
	return id, err
}

// DrstaIDZaIzmenu reads (and locks) the zaliha of a lot of an artikal in a konto of a magacin (mag);
// the error wraps sql.ErrNoRows without it.
func (r *RobnoRepository) DrstaIDZaIzmenu(ctx context.Context, tx db.Transaction, konto string, sifra, mag int, otk, serija string, roktr int64) (int64, error) {
	b := querybuilder.New(`select rstaid from drsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	b.Where("konto = ?", konto).Where("sifra = ?", sifra).Where("mag = ?", mag).
		Where("otk = ?", otk).Where("serija = ?", serija).Where("roktr = ?", roktr).
		OrderBy("rstaid").Limit(1).ForUpdate()
	var id int64
	err := r.row(ctx, tx, "zaliha lota", b, &id)
	return id, err
}

// RsifIDIliNula returns the id of an artikal by its šifra; 0 when there is none.
func (r *RobnoRepository) RsifIDIliNula(ctx context.Context, tx db.Transaction, sifra int) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`select coalesce((select rsifid from rsif where god = ? and kar = ? and sifra = ? limit 1), 0)`, s.God, s.Kar, sifra)
	var id int64
	err = r.row(ctx, tx, "artikal", b, &id)
	return id, err
}

// RsifIDMag returns the id of an artikal by its šifra and the number (mag) of a magacin; 0 for one
// that does not exist in the period.
func (r *RobnoRepository) RsifIDMag(ctx context.Context, tx db.Transaction, sifra, magaciniID int) (rsifID int64, mag int, err error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	b := querybuilder.New(`select coalesce((select rsifid from rsif where god = ? and kar = ? and sifra = ? limit 1), 0),
			coalesce((select mag from magacini where magaciniid = ? and god = ? and kar = ?), 0)`,
		s.God, s.Kar, sifra, magaciniID, s.God, s.Kar)
	err = r.row(ctx, tx, "artikal i magacin", b, &rsifID, &mag)
	return rsifID, mag, err
}

// Knjizenje is what the posting needs of a vrsta dokumenta: its knjiženje (kodknj), group (grpdok)
// and predznak, as stored.
type Knjizenje struct {
	Kodknj, Grpdok, Predznak string
}

// DokvrstaKnjizenjePoID returns the knjiženje of a vrsta dokumenta (dokvrstaid); empty when it is
// not one of the period.
func (r *RobnoRepository) DokvrstaKnjizenjePoID(ctx context.Context, tx db.Transaction, dokvrstaID int64) (Knjizenje, error) {
	b := querybuilder.New(`select coalesce(kodknj, ''), coalesce(grpdok, ''), coalesce(predznak, '') from dokvrsta`).
		Where("dokvrstaid = ?", dokvrstaID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return Knjizenje{}, err
	}
	var k Knjizenje
	err := r.row(ctx, tx, "knjiženje vrste", b, &k.Kodknj, &k.Grpdok, &k.Predznak)
	if errors.Is(err, sql.ErrNoRows) {
		return Knjizenje{}, nil
	}
	return k, err
}

// DokvrstaKnjizenjeVrste returns the knjiženje of the vrsta vrd (the upper case kodknj and grpdok,
// the trimmed predznak); the error wraps sql.ErrNoRows when it is not a vrsta of the period.
func (r *RobnoRepository) DokvrstaKnjizenjeVrste(ctx context.Context, tx db.Transaction, vrd int) (Knjizenje, error) {
	b := querybuilder.New(`select upper(trim(coalesce(kodknj, ''))), upper(trim(coalesce(grpdok, ''))), trim(coalesce(predznak, ''))
		from dokvrsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return Knjizenje{}, err
	}
	b.Where("vrd = ?", vrd).OrderBy("dokvrstaid").Limit(1)
	var k Knjizenje
	err := r.row(ctx, tx, "knjiženje vrste", b, &k.Kodknj, &k.Grpdok, &k.Predznak)
	return k, err
}

// IniVrednostTx is IniVrednost in a transaction.
func (r *RobnoRepository) IniVrednostTx(ctx context.Context, tx db.Transaction, odeljak, kljuc string) (string, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return "", err
	}
	b := querybuilder.New(`select coalesce(trim(defvrednost), '') from initable`).
		Where("kar = ?", s.Kar).Where("odeljak = ?", odeljak).Where("kljuc = ?", kljuc).Where("coalesce(username, '') = ''").
		OrderBy("initableid").Limit(1)
	var v string
	err = r.row(ctx, tx, "podešavanje", b, &v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

// RproCene are the cene and the poreska oznaka of a stavka.
type RproCene struct {
	Cena, Mcenap, Pcena, Fcena, Ncena, Mcena, Rab, Vra, Mra, Vma, Mma float64
	Po                                                                int
}

// RproCeneStavke returns the cene of a stavka; the error wraps sql.ErrNoRows when it is not a stavka
// of the period.
func (r *RobnoRepository) RproCeneStavke(ctx context.Context, tx db.Transaction, rproID int64) (RproCene, error) {
	b := querybuilder.New(`select coalesce(cena, 0)::float8, coalesce(mcenap, 0)::float8, coalesce(pcena, 0)::float8,
			coalesce(fcena, 0)::float8, coalesce(ncena, 0)::float8, coalesce(mcena, 0)::float8, coalesce(rab, 0)::float8,
			coalesce(vra, 0)::float8, coalesce(mra, 0)::float8, coalesce(vma, 0)::float8, coalesce(mma, 0)::float8, coalesce(po, 0)
		from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RproCene{}, err
	}
	var c RproCene
	err := r.row(ctx, tx, "cene stavke", b, &c.Cena, &c.Mcenap, &c.Pcena, &c.Fcena, &c.Ncena, &c.Mcena, &c.Rab,
		&c.Vra, &c.Mra, &c.Vma, &c.Mma, &c.Po)
	return c, err
}

// RdokVrdDadok returns the vrsta and the datum of a document; the error wraps sql.ErrNoRows when it
// is not a document of the period.
func (r *RobnoRepository) RdokVrdDadok(ctx context.Context, tx db.Transaction, rdokID int64) (vrd int, dadok sql.NullTime, err error) {
	b := querybuilder.New(`select coalesce(vrd, 0), dadok from rdok`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, dadok, err
	}
	err = r.row(ctx, tx, "dokument", b, &vrd, &dadok)
	return vrd, dadok, err
}

// RnalMagacin is what the posting needs of a nalog.
type RnalMagacin struct {
	MagaciniID, Mag int
	Tipdok          string
	Danal           sql.NullTime
}

// RnalMagacinNaloga returns the magacin, the vrsta and the datum of a nalog; the error wraps
// sql.ErrNoRows when it is not a nalog of the period.
func (r *RobnoRepository) RnalMagacinNaloga(ctx context.Context, tx db.Transaction, rnalID int64) (RnalMagacin, error) {
	b := querybuilder.New(`select coalesce(magaciniid, 0), coalesce(mag, 0), coalesce(tipdok, ''), danal from rnal`).Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RnalMagacin{}, err
	}
	var n RnalMagacin
	err := r.row(ctx, tx, "nalog", b, &n.MagaciniID, &n.Mag, &n.Tipdok, &n.Danal)
	return n, err
}

// RcenePostaviCene sets the cene of an artikal (rcene of its šifra) when the artikal exists.
func (r *RobnoRepository) RcenePostaviCene(ctx context.Context, tx db.Transaction, sifra int, izmene []Izmena) error {
	if len(izmene) == 0 {
		return nil
	}
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("rcene")
	if err := setIzmene(b, "rcene", izmene); err != nil {
		return err
	}
	b.Where("god = ?", s.God).Where("kar = ?", s.Kar).Where("sifra = ?", sifra).
		Where("exists (select 1 from rsif where god = ? and kar = ? and sifra = ?)", s.God, s.Kar, sifra)
	_, err = r.exec(ctx, tx, "cene artikla", b)
	return err
}

// RnalKnjizi posts a stavka on a nalog: brst stavki more (negative: less) and the iznosi on its
// duguje and potražuje (rounded to 2 decimals); ErrNotFound when it is not a nalog of the period.
func (r *RobnoRepository) RnalKnjizi(ctx context.Context, tx db.Transaction, rnalID int64, brst int, dug, pot float64) error {
	b := querybuilder.NewUpdate("rnal").
		Set(`brst = coalesce(brst, 0) + ?, dug = round((coalesce(dug, 0) + ?)::numeric, 2), pot = round((coalesce(pot, 0) + ?)::numeric, 2)`,
			brst, dug, pot).
		Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "nalog", b)
}

// RdokKnjiziStavku posts a stavka on a document: the redni broj and the broj stavki one more, the
// iznos added to its iznos and prodajna vrednost, the porez to its porez (rounded to 2 decimals);
// ErrNotFound when it is not a document of the period.
func (r *RobnoRepository) RdokKnjiziStavku(ctx context.Context, tx db.Transaction, rdokID int64, iznos, porez float64) error {
	b := querybuilder.NewUpdate("rdok").
		Set(`rbr = coalesce(rbr, 0) + 1, brst = coalesce(brst, 0) + 1,
			iznos = round((coalesce(iznos, 0) + ?)::numeric, 2), prodvred = round((coalesce(prodvred, 0) + ?)::numeric, 2),
			vporez = round((coalesce(vporez, 0) + ?)::numeric, 2)`, iznos, iznos, porez).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "dokument", b)
}

// RdokVratiStavku takes a stavka back from a document: the redni broj and the broj stavki one less,
// the iznos (negative as given) added to its iznos; ErrNotFound when it is not a document of the
// period.
func (r *RobnoRepository) RdokVratiStavku(ctx context.Context, tx db.Transaction, rdokID int64, iznos float64) error {
	b := querybuilder.NewUpdate("rdok").
		Set(`rbr = coalesce(rbr, 0) - 1, brst = coalesce(brst, 0) - 1, iznos = round((coalesce(iznos, 0) + ?)::numeric, 2)`, iznos).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "dokument", b)
}

// RporStopa returns the PDV stopa of a poreska oznaka on a datum (rpor is a catalog without a
// period).
func (r *RobnoRepository) RporStopa(ctx context.Context, tx db.Transaction, po int, datum time.Time) (float64, error) {
	b := querybuilder.New(`select `+poreskaStopa("?", "?"), po, datum)
	var stopa float64
	err := r.row(ctx, tx, "poreska stopa", b, &stopa)
	return stopa, err
}
