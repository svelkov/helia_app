package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The kalkulacije veleprodaje (rdok of the group PLT): the plaćen porez (rppo) and the plaćen avans (appo) of the
// document, its zavisni troškovi (rztr, with their porezi in rppo), the narudžbenica it refers to (edok), the cene
// of the kalkulacija in the stanje (rsta) and in the stavke (rpro), and the deletion of what refers to it (kepu).

// RdokKalkulacija is what the header of a kalkulacija has besides the columns of the faktura: the EPP polje
// (fsepp with its text), "PDV se može odbiti" and the narudžbenica (edok) with its broj and datum.
type RdokKalkulacija struct {
	EppPolje  string
	FseppID   sql.NullInt64
	FseppText string
	Odbitni   bool
	EdokID    int64
	EdokDokum int
	EdokDadok sql.NullTime
}

// RdokKalkulacijaPodaci returns the columns of the header of a kalkulacija (see RdokKalkulacija); the error wraps
// sql.ErrNoRows when it is not a document of the period.
func (r *RobnoRepository) RdokKalkulacijaPodaci(ctx context.Context, rdokID int64) (RdokKalkulacija, error) {
	b := querybuilder.New(`select coalesce(r.epp_polje, ''), nullif(r.fseppid, 0)::bigint,
			coalesce(concat_ws(' - ', nullif(f.polje, ''), nullif(f.opis, '')), ''),
			coalesce(r.odbitni, false), coalesce(r.edokid, 0)::bigint, coalesce(e.dokum, 0)::int, e.dadok
		from rdok r`).
		AddJoin("left join fsepp f on f.fseppid = r.fseppid and f.god = r.god and f.kar = r.kar").
		AddJoin("left join edok e on e.rdokid = r.edokid and e.god = r.god and e.kar = r.kar").
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return RdokKalkulacija{}, err
	}
	var k RdokKalkulacija
	err := r.row(ctx, nil, "kalkulacija", b, &k.EppPolje, &k.FseppID, &k.FseppText, &k.Odbitni, &k.EdokID, &k.EdokDokum, &k.EdokDadok)
	return k, err
}

// EdokNarudzbenica returns the broj and the datum of a narudžbenica (edok) of the period; found is false without it.
func (r *RobnoRepository) EdokNarudzbenica(ctx context.Context, edokID int64) (dokum int, dadok sql.NullTime, found bool, err error) {
	b := querybuilder.New(`select coalesce(dokum, 0), dadok from edok`).Where("rdokid = ?", edokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, sql.NullTime{}, false, err
	}
	err = r.row(ctx, nil, "narudžbenica", b, &dokum, &dadok)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, sql.NullTime{}, false, nil
	}
	return dokum, dadok, err == nil, err
}

// RdokIzvorniDobavljacaPostoji tells whether another document of the vrsta of the period has the izvorni dokument
// of the dobavljač (the legacy check of ValidacijaRDOK: fkto, fana, vrd and dokiz; rdokID is the document itself).
func (r *RobnoRepository) RdokIzvorniDobavljacaPostoji(ctx context.Context, konto, sifra string, vrd int, dokiz string, rdokID int64) (bool, error) {
	return r.postojiUPeriodu(ctx, "izvorni dokument dobavljača", querybuilder.New(`select count(*) from rdok`).
		Where("fkto = ?", konto).Where("fana = ?", sifra).Where("vrd = ?", vrd).Where("dokiz = ?", dokiz).
		Where("rdokid <> ?", rdokID))
}

// PopdvUlaz is a polje of the POPDV with what the kalkulacija checks of it: its tip, the vrsta of its evidencija
// PDV (fvepdv.vktip, "" without one) and the tipovi PDV of the partneri it is for.
type PopdvUlaz struct {
	Tip            int
	Vktip          string
	TippdvPartneri string
}

// PopdvPoljeUlaz returns the polje of the POPDV of the period (see PopdvUlaz); found is false without it.
func (r *RobnoRepository) PopdvPoljeUlaz(ctx context.Context, polje string) (PopdvUlaz, bool, error) {
	b := querybuilder.New(`select coalesce(p.tip, 0)::int, coalesce(e.vktip, ''), coalesce(p.tippdv_partneri, '') from popdv p`).
		AddJoin("left join fvepdv e on e.fvepdvid = p.fvepdvid")
	if err := inPeriod(ctx, b, "p."); err != nil {
		return PopdvUlaz{}, false, err
	}
	b.Where("p.polje = ?", polje).Limit(1)
	var p PopdvUlaz
	err := r.row(ctx, nil, "polje POPDV", b, &p.Tip, &p.Vktip, &p.TippdvPartneri)
	if errors.Is(err, sql.ErrNoRows) {
		return PopdvUlaz{}, false, nil
	}
	return p, err == nil, err
}

// RporIDNaDan returns the rpor row of a poreska oznaka (tip 0) in force on the datum (the last one from before
// it); 0 without one (rpor is a catalog without a period).
func (r *RobnoRepository) RporIDNaDan(ctx context.Context, po int, datum time.Time) (int64, error) {
	b := querybuilder.New(`select coalesce((select rporid from rpor where po = ? and coalesce(tip, 0) = 0 and datum <= ?
		order by datum desc limit 1), 0)::bigint`, po, datum)
	var id int64
	err := r.row(ctx, nil, "poreska oznaka", b, &id)
	return id, err
}

// FseppPolje returns the polje of a row of the EPP (fsepp) of the period; found is false without it.
func (r *RobnoRepository) FseppPolje(ctx context.Context, fseppID int64) (polje string, found bool, err error) {
	b := querybuilder.New(`select coalesce(polje, '') from fsepp`).Where("fseppid = ?", fseppID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", false, err
	}
	err = r.row(ctx, nil, "polje EPP", b, &polje)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return polje, err == nil, err
}

// RsifKalkulacija is an artikal of a stavka of a kalkulacija: its id, šifra, naziv, JM, konto (rsif.konto), poreska
// oznaka and its stopa on a datum.
type RsifKalkulacija struct {
	RsifID    int64
	Sifra     int
	Naziv, Jm string
	Konto     string
	Po        int
	Stopa     float64
}

// RsifArtikalKalkulacije returns the artikal by its šifra, or by its barkod (poBarkodu), with its stopa on the datum
// (see RsifKalkulacija); found is false without it.
func (r *RobnoRepository) RsifArtikalKalkulacije(ctx context.Context, poBarkodu bool, vrednost string, datum time.Time) (RsifKalkulacija, bool, error) {
	kolona := "a.sifra::text"
	if poBarkodu {
		kolona = "a.barkod"
	}
	b := querybuilder.New(`select a.rsifid, a.sifra, coalesce(a.naziv, ''), coalesce(a.jm, ''), coalesce(trim(a.konto), ''),
			coalesce(a.po, 0)::int, `+poreskaStopa("a.po", "?::date")+`
		from rsif a`, datum)
	if err := inPeriod(ctx, b, "a."); err != nil {
		return RsifKalkulacija{}, false, err
	}
	b.Where(kolona+" = ?", vrednost).Limit(1)
	var a RsifKalkulacija
	err := r.row(ctx, nil, "artikal", b, &a.RsifID, &a.Sifra, &a.Naziv, &a.Jm, &a.Konto, &a.Po, &a.Stopa)
	if errors.Is(err, sql.ErrNoRows) {
		return RsifKalkulacija{}, false, nil
	}
	return a, err == nil, err
}

// --- The plaćen porez (rppo) and the plaćen avans (appo) ---

// RppoDokumenta returns the plaćen porez of a document (its rows of rppo without a zavisni trošak) in their order.
func (r *RobnoRepository) RppoDokumenta(ctx context.Context, rdokID int64) ([]domain.RobKalPorez, error) {
	b := querybuilder.New(`select rppoid::bigint as id, coalesce(osn, 0)::float8 as osn, coalesce(ppor, 0)::float8 as ppor,
			coalesce(po, 0)::int as po, coalesce(pdv, 0)::float8 as pdv, coalesce(rporid, 0)::bigint as rporid,
			coalesce(rac, '') as rac, dadok
		from rppo`).Where("rdokid = ?", rdokID).Where("coalesce(rztrid, 0) = 0")
	return r.kalPorezi(ctx, "plaćen porez", b, "rppoid")
}

// RppoTroska returns the plaćen porez of a zavisni trošak (its rows of rppo) in their order.
func (r *RobnoRepository) RppoTroska(ctx context.Context, rztrID int64) ([]domain.RobKalPorez, error) {
	b := querybuilder.New(`select rppoid::bigint as id, coalesce(osn, 0)::float8 as osn, coalesce(ppor, 0)::float8 as ppor,
			coalesce(po, 0)::int as po, coalesce(pdv, 0)::float8 as pdv, coalesce(rporid, 0)::bigint as rporid,
			coalesce(rac, '') as rac, dadok
		from rppo`).Where("rztrid = ?", rztrID)
	return r.kalPorezi(ctx, "plaćen porez troška", b, "rppoid")
}

// AppoDokumenta returns the plaćen avans of a document (appo) in their order.
func (r *RobnoRepository) AppoDokumenta(ctx context.Context, rdokID int64) ([]domain.RobKalPorez, error) {
	b := querybuilder.New(`select appoid::bigint as id, coalesce(osn, 0)::float8 as osn, coalesce(ppor, 0)::float8 as ppor,
			coalesce(po, 0)::int as po, coalesce(pdv, 0)::float8 as pdv, coalesce(rporid, 0)::bigint as rporid,
			coalesce(rac, '') as rac, dadok
		from appo`).Where("rdokid = ?", rdokID)
	return r.kalPorezi(ctx, "plaćen avans", b, "appoid")
}

// kalPorezi reads the rows of a grid of the porezi of the period by their id.
func (r *RobnoRepository) kalPorezi(ctx context.Context, what string, b *querybuilder.Builder, id string) ([]domain.RobKalPorez, error) {
	if err := inPeriod(ctx, b, ""); err != nil {
		return nil, err
	}
	b.OrderBy(id)
	var rows []domain.RobKalPorez
	if err := r.selectRows(ctx, what, &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KalPorezDokument is the document (or the zavisni trošak) a row of rppo or appo belongs to: its vrsta, broj and
// datum, the dobavljač (konto, šifra), the broj of the računa (rac of the rppo), the document and the trošak.
type KalPorezDokument struct {
	Vrd, Dokum   int
	Dadok        sql.NullTime
	Konto, Sifra string
	Rac          string
	RdokID       int64
	RztrID       int64
}

// RppoDodaj adds a row of the plaćen porez (rppo) of a document or of a zavisni trošak (god and kar of the session).
func (r *RobnoRepository) RppoDodaj(ctx context.Context, tx db.Transaction, d KalPorezDokument, p domain.RobKalPorezUnos) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	var rztrID any
	if d.RztrID > 0 {
		rztrID = d.RztrID
	}
	b := querybuilder.New(`insert into rppo (god, kar, vrd, dokum, dadok, po, pdv, konto, sifra, rac, rdokid, rztrid, rporid, osn, ppor,
		xdatunosa, xopunos) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, d.Vrd, d.Dokum, d.Dadok, p.Po, p.Stopa, d.Konto, d.Sifra, d.Rac, d.RdokID, rztrID, p.RporID, p.Osnovica, p.Porez,
		s.UserName)
	_, err = r.exec(ctx, tx, "plaćen porez", b)
	return err
}

// AppoDodaj adds a row of the plaćen avans (appo) of a document (god and kar of the session).
func (r *RobnoRepository) AppoDodaj(ctx context.Context, tx db.Transaction, d KalPorezDokument, p domain.RobKalPorezUnos) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into appo (god, kar, vrd, dokum, dadok, po, pdv, konto, sifra, rac, rdokid, rporid, osn, ppor,
		xdatunosa, xopunos) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, d.Vrd, d.Dokum, p.Datum, p.Po, p.Stopa, d.Konto, d.Sifra, p.Rac, d.RdokID, p.RporID, p.Osnovica, p.Porez, s.UserName)
	_, err = r.exec(ctx, tx, "plaćen avans", b)
	return err
}

// RppoDokumentaBrisi deletes the plaćen porez of a document (its rows of rppo without a zavisni trošak).
func (r *RobnoRepository) RppoDokumentaBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	b := querybuilder.New(`delete from rppo`).Where("rdokid = ?", rdokID).Where("coalesce(rztrid, 0) = 0")
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "plaćen porez", b)
	return err
}

// RppoTroskaBrisi deletes the plaćen porez of a zavisni trošak.
func (r *RobnoRepository) RppoTroskaBrisi(ctx context.Context, tx db.Transaction, rztrID int64) error {
	b := querybuilder.New(`delete from rppo`).Where("rztrid = ?", rztrID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "plaćen porez troška", b)
	return err
}

// RppoTroskovaDokumentaBrisi deletes the plaćen porez of all the zavisni troškovi of a document.
func (r *RobnoRepository) RppoTroskovaDokumentaBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`delete from rppo`).
		Where("rztrid in (select z.rztrid from rztr z where z.rdokid = ? and z.god = ? and z.kar = ?)", rdokID, s.God, s.Kar)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err = r.exec(ctx, tx, "plaćen porez troškova", b)
	return err
}

// AppoDokumentaBrisi deletes the plaćen avans of a document.
func (r *RobnoRepository) AppoDokumentaBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	return r.brisiPoDokumentu(ctx, tx, "plaćen avans", "appo", rdokID)
}

// RztrDokumentaBrisi deletes the zavisni troškovi of a document (their porezi first: RppoTroskovaDokumentaBrisi).
func (r *RobnoRepository) RztrDokumentaBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	return r.brisiPoDokumentu(ctx, tx, "zavisni troškovi", "rztr", rdokID)
}

// KepuDokumentaBrisi deletes the rows of the knjiga evidencije prometa (kepu) of a document.
func (r *RobnoRepository) KepuDokumentaBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	return r.brisiPoDokumentu(ctx, tx, "knjiga evidencije prometa", "kepu", rdokID)
}

// --- The zavisni troškovi (rztr) ---

// rztrKolone are the columns of a zavisni trošak with the naziv of the dobavljač and of the konto troška.
const rztrKolone = `select z.rztrid, coalesce(z.fkto, '') as fkto, coalesce(z.fana, '') as fana, coalesce(k.naziv, '') as naziv,
			coalesce(z.dokdob, '') as dokdob, z.dadok, coalesce(z.rok, 0)::int as rok, coalesce(z.opis, '') as opis,
			coalesce(z.pornapomena, '') as pornapomena, coalesce(z.flguvozpdv, false) as flguvozpdv,
			coalesce(z.polje, '') as polje, coalesce(z.iznfak, 0)::float8 as iznfak, coalesce(z.iznos, 0)::float8 as iznos,
			coalesce(z.tkonto, '') as tkonto, coalesce(t.naziv, '') as tkontonaziv
		from rztr z`

// rztrUpit is the query of the zavisni troškovi with the nazivi (the dobavljač of the kontni plan, the konto
// troška: fkpl.vkonta 2).
func rztrUpit() *querybuilder.Builder {
	return querybuilder.New(rztrKolone).
		AddJoin(`left join lateral (select f.naziv from fkpl f where f.god = z.god and f.kar = z.kar and f.konto = z.fkto
			and f.sifra = z.fana limit 1) k on true`).
		AddJoin(`left join lateral (select f.naziv from fkpl f where f.god = z.god and f.kar = z.kar and f.konto = z.tkonto
			and f.vkonta = 2 limit 1) t on true`)
}

// RztrDokumenta returns the zavisni troškovi of a document in their order (the legacy ROB_QRY_RZTR).
func (r *RobnoRepository) RztrDokumenta(ctx context.Context, rdokID int64) ([]domain.RobKalTrosakRow, error) {
	b := rztrUpit().Where("z.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "z."); err != nil {
		return nil, err
	}
	b.OrderBy("z.rztrid")
	var rows []domain.RobKalTrosakRow
	if err := r.selectRows(ctx, "zavisni troškovi", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// RztrTrosak returns a zavisni trošak of a document; found is false when it is not one of the document in the
// period.
func (r *RobnoRepository) RztrTrosak(ctx context.Context, rztrID, rdokID int64) (domain.RobKalTrosakRow, bool, error) {
	b := rztrUpit().Where("z.rztrid = ?", rztrID).Where("z.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "z."); err != nil {
		return domain.RobKalTrosakRow{}, false, err
	}
	var rows []domain.RobKalTrosakRow
	if err := r.selectRows(ctx, "zavisni trošak", &rows, b); err != nil {
		return domain.RobKalTrosakRow{}, false, err
	}
	if len(rows) == 0 {
		return domain.RobKalTrosakRow{}, false, nil
	}
	return rows[0], true, nil
}

// RztrUnos are the values of a zavisni trošak as it is saved.
type RztrUnos struct {
	Fkto, Fana, Dokdob, Opis, Pornapomena, Polje, Tkonto string
	Dadok                                                sql.NullTime
	Rok                                                  int
	Flguvozpdv                                           bool
	Iznfak, Iznos                                        float64
}

// RztrDodaj adds a zavisni trošak of a document (its vrsta and broj; god and kar of the session) and returns its id.
func (r *RobnoRepository) RztrDodaj(ctx context.Context, tx db.Transaction, rdokID int64, vrd, dokum int, z RztrUnos) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into rztr (god, kar, rdokid, vrd, dokum, fkto, fana, dokdob, dadok, rok, opis, pornapomena,
		flguvozpdv, polje, iznfak, iznos, tkonto) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.God, s.Kar, rdokID, vrd, dokum, z.Fkto, z.Fana, z.Dokdob, z.Dadok, z.Rok, z.Opis, z.Pornapomena, z.Flguvozpdv, z.Polje,
		z.Iznfak, z.Iznos, z.Tkonto).Returning("rztrid")
	var id int64
	err = r.row(ctx, tx, "zavisni trošak", b, &id)
	return id, err
}

// RztrIzmeni changes a zavisni trošak of a document; ErrNotFound when it is not one of the document in the period.
func (r *RobnoRepository) RztrIzmeni(ctx context.Context, tx db.Transaction, rztrID, rdokID int64, z RztrUnos) error {
	b := querybuilder.NewUpdate("rztr").
		Set(`fkto = ?, fana = ?, dokdob = ?, dadok = ?, rok = ?, opis = ?, pornapomena = ?, flguvozpdv = ?, polje = ?,
			iznfak = ?, iznos = ?, tkonto = ?`, z.Fkto, z.Fana, z.Dokdob, z.Dadok, z.Rok, z.Opis, z.Pornapomena, z.Flguvozpdv,
			z.Polje, z.Iznfak, z.Iznos, z.Tkonto).
		Where("rztrid = ?", rztrID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "zavisni trošak", b)
}

// RztrBrisi deletes a zavisni trošak of a document; ErrNotFound when it is not one of the document in the period.
func (r *RobnoRepository) RztrBrisi(ctx context.Context, tx db.Transaction, rztrID, rdokID int64) error {
	b := querybuilder.New(`delete from rztr`).Where("rztrid = ?", rztrID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "zavisni trošak", b)
}

// RztrProcenat returns what the procenat of the external zavisni troškovi of a document is computed from (the legacy
// IzracunajEKSTZAVTROSAK): the sum of the iznosi of its troškovi and the fakturna vrednost robe of the last one.
func (r *RobnoRepository) RztrProcenat(ctx context.Context, rdokID int64) (ukupno, iznfak float64, err error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	b := querybuilder.New(`select coalesce(sum(coalesce(z.iznos, 0)), 0)::float8,
			coalesce((select coalesce(x.iznfak, 0) from rztr x where x.rdokid = ? and x.god = ? and x.kar = ?
				order by x.rztrid desc limit 1), 0)::float8
		from rztr z`, rdokID, s.God, s.Kar).Where("z.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "z."); err != nil {
		return 0, 0, err
	}
	err = r.row(ctx, nil, "zavisni troškovi", b, &ukupno, &iznfak)
	return ukupno, iznfak, err
}

// --- The stanje and the stavke of a kalkulacija ---

// RstaKalkulacijaCene returns the stanje of an artikal in a konto (trimmed) of a magacin with the cene of the
// kalkulacija (see domain.RobKalRstaCene); found is false without it.
func (r *RobnoRepository) RstaKalkulacijaCene(ctx context.Context, magaciniID int, konto string, sifra int) (domain.RobKalRstaCene, bool, error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8 as stanje, coalesce(cena, 0)::float8 as cena,
			coalesce(fcena, 0)::float8 as fcena, coalesce(rab, 0)::float8 as rab, coalesce(ncena, 0)::float8 as ncena,
			coalesce(vma, 0)::float8 as vma, coalesce(mma, 0)::float8 as mma, coalesce(mcenap, 0)::float8 as mcenap
		from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return domain.RobKalRstaCene{}, false, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", strings.TrimSpace(konto)).Where("sifra = ?", sifra).
		OrderBy("rstaid").Limit(1)
	var rows []domain.RobKalRstaCene
	if err := r.selectRows(ctx, "stanje", &rows, b); err != nil {
		return domain.RobKalRstaCene{}, false, err
	}
	if len(rows) == 0 {
		return domain.RobKalRstaCene{}, false, nil
	}
	return rows[0], true, nil
}

// RproKalkulacija reads (and, in a transaction, locks) a stavka of a kalkulacija (see domain.RobKalStavkaDto); the
// error wraps sql.ErrNoRows when it is not a stavka of the document in the period.
func (r *RobnoRepository) RproKalkulacija(ctx context.Context, tx db.Transaction, rproID, rdokID int64) (domain.RobKalStavkaDto, error) {
	b := querybuilder.New(`select rproid, coalesce(rbr, 0)::int as rbr, coalesce(sifra, 0)::int as sifra, coalesce(konto, '') as konto,
			coalesce(kolic, 0)::float8 as kolic, coalesce(iznos, 0)::float8 as iznos, coalesce(cena, 0)::float8 as cena,
			coalesce(fcena, 0)::float8 as fcena, coalesce(cenaval, 0)::float8 as cenaval, coalesce(rab, 0)::float8 as rab,
			coalesce(ncena, 0)::float8 as ncena, coalesce(mcena, 0)::float8 as mcena, coalesce(mcenap, 0)::float8 as mcenap,
			coalesce(vma, 0)::float8 as vma, coalesce(mma, 0)::float8 as mma, coalesce(ztro, 0)::float8 as ztro,
			coalesce(ztrin, 0)::float8 as ztrin, coalesce(tz, '') as tz, coalesce(tzi, '') as tzi, coalesce(otk, '') as otk,
			coalesce(serija, '') as serija, coalesce(roktr, 0)::float8 as roktr, coalesce(naz1, '') as naz1, coalesce(jm, '') as jm,
			coalesce(po, 0)::int as po
		from rpro`).Where("rproid = ?", rproID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return domain.RobKalStavkaDto{}, err
	}
	if tx != nil {
		b.ForUpdate()
	}
	var s domain.RobKalStavkaDto
	err := r.row(ctx, tx, "stavka", b, &s.RproID, &s.Rbr, &s.Sifra, &s.Konto, &s.Kolicina, &s.Iznos, &s.Cena, &s.Fcena, &s.Cenaval,
		&s.Rab, &s.Ncena, &s.Mcena, &s.Mcenap, &s.Vma, &s.Mma, &s.Ztro, &s.Ztrin, &s.Tz, &s.Tzi, &s.Otk, &s.Serija, &s.Roktr,
		&s.Naz1, &s.Jm, &s.Po)
	return s, err
}

// RceneMaloprodajne returns the maloprodajna cena without and with the porez of an artikal (rcene.mcena, rcene.mcenap:
// the last ones of a kalkulacija); found is false without its cene.
func (r *RobnoRepository) RceneMaloprodajne(ctx context.Context, sifra int) (mcena, mcenap float64, found bool, err error) {
	b := querybuilder.New(`select coalesce(mcena, 0)::float8, coalesce(mcenap, 0)::float8 from rcene`).Where("sifra = ?", sifra)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, false, err
	}
	b.OrderBy("rceneid").Limit(1)
	err = r.row(ctx, nil, "cene artikla", b, &mcena, &mcenap)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return mcena, mcenap, err == nil, err
}

// RproZaliheBrisanja returns the artikli of the stavke of a document whose stanje in the konto of the magacin of the
// document would be negative after the deletion of the document (the legacy check of BTN_Brisanje).
func (r *RobnoRepository) RproZaliheBrisanja(ctx context.Context, rdokID int64) ([]domain.RobKalZalihaBrisanja, error) {
	b := querybuilder.New(`select p.sifra, coalesce(s.naziv, '') as naziv,
			(coalesce(t.ulaz, 0) - coalesce(t.izlaz, 0) - coalesce(p.kolic, 0))::float8 as stanje
		from rpro p`).
		AddJoin("join rdok d on d.rdokid = p.rdokid").
		AddJoin(`join rsta t on t.god = p.god and t.kar = p.kar and t.magaciniid = d.magaciniid and t.konto = p.konto
			and t.sifra = p.sifra`).
		AddJoin("left join rsif s on s.god = p.god and s.kar = p.kar and s.sifra = p.sifra").
		Where("p.rdokid = ?", rdokID).
		Where("coalesce(t.ulaz, 0) - coalesce(t.izlaz, 0) - coalesce(p.kolic, 0) < 0")
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr")
	var rows []domain.RobKalZalihaBrisanja
	if err := r.selectRows(ctx, "zalihe posle brisanja", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// DrstaLotPostaviVpcena sets the veleprodajna cena of a lot (OTK broj, serija, rok) of an artikal in a konto of a
// magacin (mag).
func (r *RobnoRepository) DrstaLotPostaviVpcena(ctx context.Context, tx db.Transaction, konto string, sifra, mag int, otk, serija string, roktr int64, vpcena float64) error {
	b := querybuilder.NewUpdate("drsta").Set("vpcena = ?", vpcena)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	b.Where("konto = ?", konto).Where("sifra = ?", sifra).Where("mag = ?", mag).Where("otk = ?", otk).
		Where("serija = ?", serija).Where("roktr = ?", roktr)
	_, err := r.exec(ctx, tx, "veleprodajna cena zaliha", b)
	return err
}
