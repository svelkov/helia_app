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

// The prenosnice (a document of a magacin and its kontra dokument in the magacin where the roba
// goes) and the settings of the firm, in the period of the session.

// tipoviMagacinaPrenosnice are the tipovi of the magacini a prenosnica may send the roba to (the
// legacy ROB_QRY_MAG with "G;V;D;S;L"; the search "magacini-prenosa" has the same list).
const tipoviMagacinaPrenosnice = "('G', 'V', 'D', 'S', 'L')"

// IniVrednost returns the value of a setting of the firm (initable of the kar of the session, the
// odeljak and the ključ, without a username), trimmed; "" without it. initable has no god.
func (r *RobnoRepository) IniVrednost(ctx context.Context, odeljak, kljuc string) (string, error) {
	return r.IniVrednostTx(ctx, nil, odeljak, kljuc)
}

// RdokPrenosnicaMagacin1 returns the magacin where the roba of a prenosnica goes and the konto
// zaliha of its kontra dokument: key = magid1, c1 = mag - opis of that magacin, c2 = the konto.
func (r *RobnoRepository) RdokPrenosnicaMagacin1(ctx context.Context, rdokID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select coalesce(r.magid1, 0)::text as key,
			coalesce(concat_ws(' - ', m.mag::text, nullif(m.opis, '')), '') as c1,
			coalesce(trim(coalesce(k.pkto, k2.pkto)), '') as c2
		from rdok r`).
		AddJoin("left join magacini m on m.magaciniid = r.magid1").
		AddJoin("left join rdok k on k.rdokid = r.vrdokid").
		AddJoin("left join lateral (select x.pkto from rdok x where x.vrdokid = r.rdokid order by x.rdokid limit 1) k2 on true").
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "magacin prenosnice", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// MagacinPrenosnice returns the first magacin (by mag) other than magaciniID a prenosnica may send
// the roba to: key = magaciniid, c1 = mag - opis; no row without one.
func (r *RobnoRepository) MagacinPrenosnice(ctx context.Context, magaciniID int) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select m.magaciniid::text as key, coalesce(concat_ws(' - ', m.mag::text, nullif(m.opis, '')), '') as c1
		from magacini m`)
	if err := inPeriod(ctx, b, "m."); err != nil {
		return nil, err
	}
	b.Where("m.magaciniid <> ?", magaciniID).Where("upper(trim(coalesce(m.tipmag, ''))) in " + tipoviMagacinaPrenosnice).
		OrderBy("m.mag").Limit(1)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "magacin prenosnice", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// MagkontoNazivKonta returns the naziv (of the kontni plan) of a konto (trimmed) of a magacin; ""
// without it.
func (r *RobnoRepository) MagkontoNazivKonta(ctx context.Context, magaciniID int, konto string) (string, error) {
	b := querybuilder.New(`select coalesce((select f.naziv from fkpl f where f.idfkpl = k.idfkpl), '') from magkonto k`).
		Where("k.magaciniid = ?", magaciniID).Where("trim(k.konto) = ?", strings.TrimSpace(konto))
	if err := inPeriod(ctx, b, "k."); err != nil {
		return "", err
	}
	b.OrderBy("k.magkontoid").Limit(1)
	var naziv string
	err := r.row(ctx, nil, "naziv konta", b, &naziv)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return strings.TrimSpace(naziv), err
}

// RdokPoBrojuZaIzmenu reads (and locks) the document of a magacin by its vrsta and broj with its
// kontra dokument (vrdokid, 0 without it); the error wraps sql.ErrNoRows when there is none.
func (r *RobnoRepository) RdokPoBrojuZaIzmenu(ctx context.Context, tx db.Transaction, magaciniID, vrd, dokum int) (rdokID, vrdokID int64, err error) {
	b := querybuilder.New(`select rdokid, coalesce(vrdokid, 0)::bigint from rdok`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("vrd = ?", vrd).Where("dokum = ?", dokum).ForUpdate()
	err = r.row(ctx, tx, "dokument", b, &rdokID, &vrdokID)
	return rdokID, vrdokID, err
}

// MagaciniMag returns the numbers (mag) of two magacini; 0 for one that is not a magacin of the
// period.
func (r *RobnoRepository) MagaciniMag(ctx context.Context, tx db.Transaction, magaciniID, magaciniID1 int) (mag, mag1 int, err error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	b := querybuilder.New(`select coalesce((select mag from magacini where magaciniid = ? and god = ? and kar = ?), 0),
			coalesce((select mag from magacini where magaciniid = ? and god = ? and kar = ?), 0)`,
		magaciniID, s.God, s.Kar, magaciniID1, s.God, s.Kar)
	err = r.row(ctx, tx, "magacini", b, &mag, &mag1)
	return mag, mag1, err
}

// RdokSledeciBroj returns the next broj of the documents of the vrsta in a magacin.
func (r *RobnoRepository) RdokSledeciBroj(ctx context.Context, tx db.Transaction, magaciniID, vrd int) (int, error) {
	b := querybuilder.New(`select coalesce(max(dokum), 0) + 1 from rdok`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("vrd = ?", vrd)
	var dokum int
	err := r.row(ctx, tx, "broj dokumenta", b, &dokum)
	return dokum, err
}

// RdokPostaviKontru sets the kontra dokument of a document; ErrNotFound when it is not a document of
// the period.
func (r *RobnoRepository) RdokPostaviKontru(ctx context.Context, tx db.Transaction, rdokID, kontraID int64) error {
	b := querybuilder.NewUpdate("rdok").Set("vrdokid = ?", kontraID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "kontra dokument", b)
}

// RdokKontraPoVeziID returns the first document whose kontra dokument (vrdokid) is rdokID; 0
// without it.
func (r *RobnoRepository) RdokKontraPoVeziID(ctx context.Context, tx db.Transaction, rdokID int64) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`select coalesce((select rdokid from rdok where vrdokid = ? and god = ? and kar = ? order by rdokid limit 1), 0)::bigint`,
		rdokID, s.God, s.Kar)
	var id int64
	err = r.row(ctx, tx, "kontra dokument", b, &id)
	return id, err
}

// RdokKontra returns the kontra dokument of a prenosnica (its vrdokid, else the document whose
// vrdokid is the prenosnica) with its konto zaliha (pkto, trimmed); found is false without it.
func (r *RobnoRepository) RdokKontra(ctx context.Context, rdokID int64) (kontraID int64, konto1 string, found bool, err error) {
	b := querybuilder.New(`select k.rdokid, coalesce(trim(k.pkto), '') from rdok r`).
		AddJoin("join rdok k on k.rdokid = coalesce(nullif(r.vrdokid, 0), (select x.rdokid from rdok x where x.vrdokid = r.rdokid order by x.rdokid limit 1))").
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return 0, "", false, err
	}
	err = r.row(ctx, nil, "kontra dokument", b, &kontraID, &konto1)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", false, nil
	}
	return kontraID, konto1, err == nil, err
}

// RproPostaviDadok sets the datum dokumenta of the stavke of a document.
func (r *RobnoRepository) RproPostaviDadok(ctx context.Context, tx db.Transaction, rdokID int64, dadok any) error {
	b := querybuilder.NewUpdate("rpro").Set("dadok = ?", dadok).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "datum stavki", b)
	return err
}

// RnalZaKopiju is what a nalog of another magacin copies of a nalog.
type RnalZaKopiju struct {
	Tipdok       string
	Nalog        int
	IDTipdok     sql.NullInt64
	Danal, Datob sql.NullTime
}

// RnalKopija reads the vrsta, the broj, the tip and the datumi of a nalog; the error wraps
// sql.ErrNoRows when it is not a nalog of the period.
func (r *RobnoRepository) RnalKopija(ctx context.Context, tx db.Transaction, rnalID int64) (RnalZaKopiju, error) {
	b := querybuilder.New(`select coalesce(tipdok, ''), coalesce(nalog, 0), idtipdok, danal, datob from rnal`).Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RnalZaKopiju{}, err
	}
	var n RnalZaKopiju
	err := r.row(ctx, tx, "nalog", b, &n.Tipdok, &n.Nalog, &n.IDTipdok, &n.Danal, &n.Datob)
	return n, err
}

// RnalPoBroju returns the nalog of a magacin by its broj and vrsta; the error wraps sql.ErrNoRows
// when there is none.
func (r *RobnoRepository) RnalPoBroju(ctx context.Context, tx db.Transaction, nalog int, tipdok string, magaciniID int) (int64, error) {
	b := querybuilder.New(`select rnalid from rnal`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	b.Where("nalog = ?", nalog).Where("tipdok = ?", tipdok).Where("magaciniid = ?", magaciniID).OrderBy("rnalid").Limit(1)
	var id int64
	err := r.row(ctx, tx, "nalog", b, &id)
	return id, err
}

// DrstaLot is the zaliha of a lot of an artikal.
type DrstaLot struct {
	Ulaz, Izlaz, Cena, Prosnc, Ncena float64
}

// DrstaLotStanje returns the zaliha of the lot (OTK broj, serija, rok trajanja; trimmed) of an
// artikal in a konto of a magacin (mag); found is false without it.
func (r *RobnoRepository) DrstaLotStanje(ctx context.Context, konto string, sifra, mag int, otk, serija string, roktr int64) (DrstaLot, bool, error) {
	b := querybuilder.New(`select coalesce(ulaz, 0)::float8, coalesce(izlaz, 0)::float8,
			coalesce(cena, 0)::float8, coalesce(prosnc, 0)::float8, coalesce(ncena, 0)::float8
		from drsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return DrstaLot{}, false, err
	}
	b.Where("konto = ?", konto).Where("sifra = ?", sifra).Where("mag = ?", mag).
		Where("otk = ?", strings.TrimSpace(otk)).Where("serija = ?", strings.TrimSpace(serija)).Where("roktr = ?", roktr).
		OrderBy("rstaid").Limit(1)
	var l DrstaLot
	err := r.row(ctx, nil, "zaliha lota", b, &l.Ulaz, &l.Izlaz, &l.Cena, &l.Prosnc, &l.Ncena)
	if errors.Is(err, sql.ErrNoRows) {
		return DrstaLot{}, false, nil
	}
	return l, err == nil, err
}

// RstaUlazIzlazProsnc returns the ulaz and the izlaz (rounded to 3 decimals) and the prosečna nabavna
// cena of an artikal in a konto of a magacin; the error wraps sql.ErrNoRows without its stanje.
func (r *RobnoRepository) RstaUlazIzlazProsnc(ctx context.Context, tx db.Transaction, magaciniID int, konto string, sifra int) (ulaz, izlaz, prosnc float64, err error) {
	b := querybuilder.New(`select round(coalesce(ulaz, 0), 3)::float8, round(coalesce(izlaz, 0), 3)::float8, coalesce(prosnc, 0)::float8
		from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, 0, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", konto).Where("sifra = ?", sifra).OrderBy("rstaid").Limit(1)
	err = r.row(ctx, tx, "stanje", b, &ulaz, &izlaz, &prosnc)
	return ulaz, izlaz, prosnc, err
}

// RstaPostaviProsnc sets the prosečna nabavna cena of the stanje of an artikal in a konto of a
// magacin.
func (r *RobnoRepository) RstaPostaviProsnc(ctx context.Context, tx db.Transaction, magaciniID int, konto string, sifra int, prosnc float64) error {
	b := querybuilder.NewUpdate("rsta").Set("prosnc = ?", prosnc)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", konto).Where("sifra = ?", sifra)
	_, err := r.exec(ctx, tx, "prosečna nabavna cena stanja", b)
	return err
}

// DrstaPostaviProsnc sets the prosečna nabavna cena of the zalihe by lot of an artikal in a magacin
// (mag).
func (r *RobnoRepository) DrstaPostaviProsnc(ctx context.Context, tx db.Transaction, mag int, rsifID int64, prosnc float64) error {
	b := querybuilder.NewUpdate("drsta").Set("prosnc = ?", prosnc)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	b.Where("mag = ?", mag).Where("rsifid = ?", rsifID)
	_, err := r.exec(ctx, tx, "prosečna nabavna cena zaliha", b)
	return err
}

// RproLotBrisanje is a stavka with its lot, to delete.
type RproLotBrisanje struct {
	Sifra, Rbr             int
	Konto, Otk, Serija     string
	Kolicina, Iznos, Roktr float64
}

// RproStavkaLotBrisanje reads (and locks) a stavka with its lot to delete; the error wraps
// sql.ErrNoRows when it is not a stavka of the period.
func (r *RobnoRepository) RproStavkaLotBrisanje(ctx context.Context, tx db.Transaction, rproID int64) (RproLotBrisanje, error) {
	b := querybuilder.New(`select coalesce(sifra, 0), coalesce(rbr, 0), coalesce(konto, ''), coalesce(otk, ''), coalesce(serija, ''),
			coalesce(kolic, 0)::float8, coalesce(iznos, 0)::float8, coalesce(roktr, 0)::float8
		from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RproLotBrisanje{}, err
	}
	b.ForUpdate()
	var st RproLotBrisanje
	err := r.row(ctx, tx, "stavka", b, &st.Sifra, &st.Rbr, &st.Konto, &st.Otk, &st.Serija, &st.Kolicina, &st.Iznos, &st.Roktr)
	return st, err
}

// RproKontraStavka returns the stavka of the kontra dokument kontraID made for the stavka rproID
// (rproid1); 0 without it.
func (r *RobnoRepository) RproKontraStavka(ctx context.Context, tx db.Transaction, kontraID, rproID int64) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`select coalesce((select rproid from rpro where rdokid = ? and rproid1 = ? and god = ? and kar = ?
		order by rproid limit 1), 0)::bigint`, kontraID, rproID, s.God, s.Kar)
	var id int64
	err = r.row(ctx, tx, "stavka kontra dokumenta", b, &id)
	return id, err
}
