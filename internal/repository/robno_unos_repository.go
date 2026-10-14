package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The entry of the robni dokumenti (the fakture and the documents that share their entry): the
// header, the avansi, the rate, the obračun, the stavke and their artikli, in the period of the
// session.

// KontoMagacinaIliStavke returns the first konto of a magacin, else the konto of the last stavka of
// the document; "" without both.
func (r *RobnoRepository) KontoMagacinaIliStavke(ctx context.Context, magaciniID int, rdokID int64) (string, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return "", err
	}
	b := querybuilder.New(`select coalesce(
			(select m.konto from magkonto m where m.god = ? and m.kar = ? and m.magaciniid = ?
				and coalesce(m.konto, '') <> '' order by m.magkontoid limit 1),
			(select p.konto from rpro p where p.rdokid = ? and p.god = ? and p.kar = ? and coalesce(p.konto, '') <> ''
				order by p.rproid desc limit 1),
			'')`, s.God, s.Kar, magaciniID, rdokID, s.God, s.Kar)
	var konto string
	err = r.row(ctx, nil, "konto magacina", b, &konto)
	return konto, err
}

// SifreSifarnika returns the old columns of the šifarnici by their keys: key = mi of the mesto
// isporuke, c1 = sifkom of the komercijalista, c2 = sifval of the valuta, c3 = sifbank of the banka,
// c4 = keykat and c5 = category of the poreska kategorija (0 or "" for a key that does not exist; the
// mesto isporuke, the komercijalista and the banka of the period, valute and porkat are catalogs).
func (r *RobnoRepository) SifreSifarnika(ctx context.Context, fispID, komID, valutaID, bankaID, porkatID int64) (domain.SearchRowDto, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return domain.SearchRowDto{}, err
	}
	b := querybuilder.New(`select
			coalesce((select f.mi from fisp f where f.fispid = ? and f.god = ? and f.kar = ?), 0)::bigint::text as key,
			coalesce((select k.sifkom from komercijalisti k where k.komid = ? and k.god = ? and k.kar = ?), 0)::text as c1,
			coalesce((select v.sifval from valute v where v.idvalute = ?), 0)::text as c2,
			coalesce((select b.sifbank from bnkizv b where b.bnkizvid = ? and b.god = ? and b.kar = ?), 0)::text as c3,
			coalesce((select p.keykat from porkat p where p.porkatid = ?), '') as c4,
			coalesce((select p.category from porkat p where p.porkatid = ?), '') as c5`,
		fispID, s.God, s.Kar, komID, s.God, s.Kar, valutaID, bankaID, s.God, s.Kar, porkatID, porkatID)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "šifre šifarnika", &rows, b); err != nil {
		return domain.SearchRowDto{}, err
	}
	if len(rows) == 0 {
		return domain.SearchRowDto{}, nil
	}
	return rows[0], nil
}

// AvansIDPoBroju returns the avansni račun (vrsta 172) of the kupac with the broj that has an avans;
// 0 without it.
func (r *RobnoRepository) AvansIDPoBroju(ctx context.Context, dokum int, konto, sifra string) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`select coalesce((select r.rdokid from rdok r
			where r.god = ? and r.kar = ? and r.vrd = 172 and r.dokum = ? and r.fkto = ? and r.fana = ?
				and exists (select 1 from avansfakt af where af.avansid = r.rdokid)
			order by r.rdokid limit 1), 0)::bigint`, s.God, s.Kar, dokum, strings.TrimSpace(konto), strings.TrimSpace(sifra))
	var id int64
	err = r.row(ctx, nil, "avans", b, &id)
	return id, err
}

// RdokPoBrojuFana reads (and locks) the document of a magacin by its vrsta and broj with the šifra of
// its kupac; the error wraps sql.ErrNoRows when there is none.
func (r *RobnoRepository) RdokPoBrojuFana(ctx context.Context, tx db.Transaction, magaciniID, vrd, dokum int) (rdokID int64, fana string, err error) {
	b := querybuilder.New(`select rdokid, coalesce(fana, '') from rdok`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, "", err
	}
	b.Where("magaciniid = ?", magaciniID).Where("vrd = ?", vrd).Where("dokum = ?", dokum).ForUpdate()
	err = r.row(ctx, tx, "dokument", b, &rdokID, &fana)
	return rdokID, fana, err
}

// kolonePartneraStavke are the columns of the stavke a document passes to them.
var kolonePartneraStavke = map[string]bool{"fana": true, "pkto": true, "pana": true}

// RproPostaviPartnera sets a column of the partner (fana, pkto or pana) of the stavke of a document.
func (r *RobnoRepository) RproPostaviPartnera(ctx context.Context, tx db.Transaction, rdokID int64, kolona, vrednost string) error {
	if !kolonePartneraStavke[kolona] {
		return fmt.Errorf("rpro: kolona %q nije dozvoljena", kolona)
	}
	b := querybuilder.NewUpdate("rpro").Set(querybuilder.QuoteIdent(kolona)+" = ?", vrednost).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "partner stavki", b)
	return err
}

// RnalDokumenta is what a new document copies of its nalog.
type RnalDokumenta struct {
	Tipdok                 string
	Nalog, Mag, MagaciniID int
	Brdo                   int
	Danal, Datob           sql.NullTime
}

// RnalZaDokument reads (and locks) the nalog of a new document; the error wraps sql.ErrNoRows when it
// is not a nalog of the period.
func (r *RobnoRepository) RnalZaDokument(ctx context.Context, tx db.Transaction, rnalID int64) (RnalDokumenta, error) {
	b := querybuilder.New(`select coalesce(tipdok, ''), coalesce(nalog, 0), danal, datob, coalesce(mag, 0),
		coalesce(magaciniid, 0), coalesce(brdo, 0) from rnal`).Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RnalDokumenta{}, err
	}
	b.ForUpdate()
	var n RnalDokumenta
	err := r.row(ctx, tx, "nalog", b, &n.Tipdok, &n.Nalog, &n.Danal, &n.Datob, &n.Mag, &n.MagaciniID, &n.Brdo)
	return n, err
}

// AvansfaktBrisiOstale deletes the avansi closed on a faktura other than the ones of ostaju (all of
// them without ostaju).
func (r *RobnoRepository) AvansfaktBrisiOstale(ctx context.Context, tx db.Transaction, rdokID int64, ostaju []int64) error {
	b := querybuilder.New(`delete from avansfakt`).Where("rdokid = ?", rdokID)
	if len(ostaju) > 0 {
		b.Where("avansid not in ("+strings.TrimSuffix(strings.Repeat("?, ", len(ostaju)), ", ")+")", querybuilder.Args(ostaju)...)
	}
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "avansi fakture", b)
	return err
}

// AvansfaktPostoji tells whether the avans is closed on the faktura.
func (r *RobnoRepository) AvansfaktPostoji(ctx context.Context, tx db.Transaction, avansID, rdokID int64) (bool, error) {
	b := querybuilder.New(`select count(*) from avansfakt`).Where("avansid = ?", avansID).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	var n int
	if err := r.row(ctx, tx, "avans fakture", b, &n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// AvansIznosi are the osnovice and the PDV of an avans.
type AvansIznosi struct {
	Osn1, Osn2, Osn3, Pdv1, Pdv2 float64
}

// AvansfaktIznosi returns the osnovice and the PDV of an avans (its first row); the error wraps
// sql.ErrNoRows without it.
func (r *RobnoRepository) AvansfaktIznosi(ctx context.Context, tx db.Transaction, avansID int64) (AvansIznosi, error) {
	b := querybuilder.New(`select coalesce(osnavans1, 0)::float8, coalesce(osnavans2, 0)::float8,
		coalesce(osnavans3, 0)::float8, coalesce(pdv1avans, 0)::float8, coalesce(pdv2avans, 0)::float8 from avansfakt`).
		Where("avansid = ?", avansID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return AvansIznosi{}, err
	}
	b.OrderBy("avansfaktid").Limit(1)
	var a AvansIznosi
	err := r.row(ctx, tx, "iznosi avansa", b, &a.Osn1, &a.Osn2, &a.Osn3, &a.Pdv1, &a.Pdv2)
	return a, err
}

// AvansiObracuna returns the avansi closed on a faktura for its obračun (by the datum of the avansni
// račun): the osnovice and the PDV of the avans and what the other fakture closed of them.
func (r *RobnoRepository) AvansiObracuna(ctx context.Context, rdokID int64) ([]domain.RobObracunAvansDto, error) {
	b := querybuilder.New(`select af.avansfaktid as avansid,
			coalesce(af.osnavans1, 0)::float8 as osnavans1, coalesce(af.osnavans2, 0)::float8 as osnavans2,
			coalesce(af.osnavans3, 0)::float8 as osnavans3, coalesce(af.pdv1avans, 0)::float8 as pdv1avans,
			coalesce(af.pdv2avans, 0)::float8 as pdv2avans,
			coalesce(o.osnfakt1, 0)::float8 as osnfakt1, coalesce(o.osnfakt2, 0)::float8 as osnfakt2,
			coalesce(o.osnfakt3, 0)::float8 as osnfakt3, coalesce(o.pdv1fakt, 0)::float8 as pdv1fakt,
			coalesce(o.pdv2fakt, 0)::float8 as pdv2fakt
		from avansfakt af`).
		AddJoin("left join rdok a on a.rdokid = af.avansid").
		AddJoin(`left join lateral (select sum(x.osnfakt1) as osnfakt1, sum(x.osnfakt2) as osnfakt2, sum(x.osnfakt3) as osnfakt3,
				sum(x.pdv1fakt) as pdv1fakt, sum(x.pdv2fakt) as pdv2fakt
			from avansfakt x where x.avansid = af.avansid and x.rdokid <> af.rdokid and x.rdokid <> 0) o on true`).
		Where("af.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "af."); err != nil {
		return nil, err
	}
	b.OrderByExpr("a.dadok nulls last, af.avansfaktid")
	var rows []domain.RobObracunAvansDto
	if err := r.selectRows(ctx, "avansi obračuna", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// FaktrateDokumenta returns the rate of a document in their order.
func (r *RobnoRepository) FaktrateDokumenta(ctx context.Context, rdokID int64) ([]domain.RobRataDto, error) {
	b := querybuilder.New(`select faktrateid, coalesce(rbrrate, 0) as rbrrate,
		coalesce(osnovica, 0)::float8 as osnovica, coalesce(procenat, 0)::float8 as procenat from faktrate`).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return nil, err
	}
	b.OrderBy("rbrrate, faktrateid")
	var rows []domain.RobRataDto
	if err := r.selectRows(ctx, "rate", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// StavkeObracuna returns the stavke of a document for its obračun, with the stopa of their oznaka on
// the datum of the document; a document without stavke gives one row with zeros.
func (r *RobnoRepository) StavkeObracuna(ctx context.Context, rdokID int64) ([]domain.RobObracunStavkaDto, error) {
	b := querybuilder.New(`select coalesce(r.ugrabat, 0)::float8 as ugrabat,
			coalesce(r.pkase, 0)::float8 as pkase, coalesce(r.pornapomena, '') as pornapomena,
			coalesce(st.kolic, 0)::float8 as kolic, coalesce(st.fcena, 0)::float8 as fcena,
			coalesce(st.cena, 0)::float8 as cena, coalesce(st.rab, 0)::float8 as rab,
			coalesce(st.taxcat, '') as taxcat, coalesce(ps.pp, 0)::float8 as stopa
		from rdok r`).
		AddJoin("left join rpro st on st.rdokid = r.rdokid").
		AddJoin(`left join lateral (select x.pp from rpor x where x.po = st.po and x.datum <= r.dadok
			order by x.datum desc limit 1) ps on true`).
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	var rows []domain.RobObracunStavkaDto
	if err := r.selectRows(ctx, "stavke obračuna", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// SaldoKupca returns the sums of the saldo of the kupac (konto, šifra; trimmed) on the datum: its
// fakture, the dospele and the nedospele ones and its uplate of the finansijsko (fpro), and the
// dospele and nedospele fakture of the robno not yet posted (rdok).
func (r *RobnoRepository) SaldoKupca(ctx context.Context, konto, sifra string, datum any) ([]domain.RobSaldoKupcaDto, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	konto, sifra = strings.TrimSpace(konto), strings.TrimSpace(sifra)
	b := querybuilder.New(`select fin.fakture::float8 as fakture, fin.dospelo::float8 as dospelo, fin.nedospelo::float8 as nedospelo,
			fin.uplate::float8 as uplate,
			coalesce((select sum(iznfakt) from rob where dadok + rok <= ?), 0)::float8 as dospelorob,
			coalesce((select sum(iznfakt) from rob where dadok is null or dadok + rok > ?), 0)::float8 as nedospelorob
		from fin`, datum, datum).
		WithCTE("fin", `select coalesce(sum(case when kat in (1, 2) then iznos end), 0) as fakture,
				coalesce(sum(case when kat in (1, 2) and dadok + coalesce(rok, 0) <= ? then iznos end), 0) as dospelo,
				coalesce(sum(case when kat in (1, 2) and (dadok is null or dadok + coalesce(rok, 0) > ?) then iznos end), 0) as nedospelo,
				coalesce(sum(case when kat in (3, 4) then iznos end), 0) as uplate
			from fpro where god = ? and kar = ? and konto = ? and sifra = ? and vrd in (10, 30, 80, 90)`,
			datum, datum, s.God, s.Kar, konto, sifra).
		WithCTE("dok", `select r.rdokid, r.dadok, coalesce(r.rok, 0) as rok, coalesce(r.ugrabat, 0) as ugrabat, coalesce(r.pkase, 0) as pkase,
				round(coalesce(sum(st.kolic * st.fcena), 0), 2) as iznos,
				round(coalesce(sum(st.kolic * st.fcena * coalesce(st.rab, 0) / 100), 0), 2) as rab,
				coalesce(sum(round(st.kolic * st.fcena * coalesce(ps.pp, 0) / 100, 2)), 0) as porez
			from rdok r
			left join rpro st on st.rdokid = r.rdokid
			left join lateral (select x.pp from rpor x where x.po = st.po and x.datum <= st.dadok
				order by x.datum desc limit 1) ps on true
			where r.god = ? and r.kar = ? and r.fkto = ? and r.fana = ? and coalesce(r.knjige_1, '') <> 'D'
				and r.vrd in (select d.vrd from dokvrsta d where d.god = ? and d.kar = ?
					and d.dokozn in ('FAK', 'FPR', 'FRP', 'KPK'))
			group by r.rdokid, r.dadok, r.rok, r.ugrabat, r.pkase`,
			s.God, s.Kar, konto, sifra, s.God, s.Kar).
		WithCTE("rob", `select dadok, rok, iznos - rab - u - round((iznos - rab - u) * pkase / 100, 2) + porez as iznfakt
			from (select dok.*, round((iznos - rab) * ugrabat / 100, 2) as u from dok) d`)
	var rows []domain.RobSaldoKupcaDto
	if err := r.selectRows(ctx, "saldo kupca", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// AvansiDokumenta returns the avansi closed on a document: per avans its avansni račun, its iznos,
// what of it is still open after all the fakture it is closed on and the iznos closed on this
// document.
func (r *RobnoRepository) AvansiDokumenta(ctx context.Context, rdokID int64) ([]domain.RobAvansDto, error) {
	b := querybuilder.New(`select af.avansid::bigint as avansid, coalesce(a.dokum, 0) as dokum, a.dadok,
			(coalesce(af.osnavans1, 0) + coalesce(af.osnavans2, 0) + coalesce(af.osnavans3, 0)
				+ coalesce(af.pdv1avans, 0) + coalesce(af.pdv2avans, 0))::float8 as iznos,
			(coalesce(af.osnavans1, 0) + coalesce(af.osnavans2, 0) + coalesce(af.osnavans3, 0)
				+ coalesce(af.pdv1avans, 0) + coalesce(af.pdv2avans, 0)
				- coalesce((select sum(coalesce(x.osnfakt1, 0) + coalesce(x.osnfakt2, 0) + coalesce(x.osnfakt3, 0)
					+ coalesce(x.pdv1fakt, 0) + coalesce(x.pdv2fakt, 0)) from avansfakt x where x.avansid = af.avansid), 0))::float8 as ostatak,
			(coalesce(af.osnfakt1, 0) + coalesce(af.osnfakt2, 0) + coalesce(af.osnfakt3, 0)
				+ coalesce(af.pdv1fakt, 0) + coalesce(af.pdv2fakt, 0))::float8 as zatvoreno
		from avansfakt af`).
		AddJoin("left join rdok a on a.rdokid = af.avansid").
		Where("af.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "af."); err != nil {
		return nil, err
	}
	b.OrderBy("af.avansfaktid")
	var rows []domain.RobAvansDto
	if err := r.selectRows(ctx, "avansi dokumenta", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// BankeZiroRacuni returns the žiro računi of the firm the documents show ("broj računa  banka"),
// without the ones not for the fakture.
func (r *RobnoRepository) BankeZiroRacuni(ctx context.Context) ([]string, error) {
	b := querybuilder.New(`select coalesce(brrac, '') || '  ' || coalesce(banka, '') from banke`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return nil, err
	}
	b.Where("not coalesce(nafakne, false)").OrderBy("idbanke")
	var rows []string
	if err := r.selectRows(ctx, "žiro računi", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// NoviDokumentDefaults returns the defaults of a new document of the firm (by its naziv): key = the
// šifra of the domaća valuta, c1 = its idvalute, c2 = its text, c3 = the konto of the kupci; no row
// without the firma.
func (r *RobnoRepository) NoviDokumentDefaults(ctx context.Context, firma string) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select coalesce(f.sifval, 0)::text as key, coalesce(v.idvalute::text, '') as c1,
			coalesce(concat_ws(' - ', v.sifval::text, nullif(v.naziv, '')), '') as c2, coalesce(rv.fkup, '') as c3
		from fvr f`).
		AddJoin(`left join lateral (select x.idvalute, x.sifval, x.naziv from valute x where x.sifval = f.sifval
			order by x.idvalute limit 1) v on true`).
		AddJoin("left join rvr rv on rv.god = f.god and rv.kar = f.kar")
	if err := inPeriod(ctx, b, "f."); err != nil {
		return nil, err
	}
	b.Where("f.naziv = ?", firma).Limit(1)
	var rows []domain.SearchRowDto
	if err := r.selectRows(ctx, "podrazumevane vrednosti", &rows, b); err != nil {
		return nil, err
	}
	return &rows, nil
}

// RproStavkeDokumenta returns the stavke of a document, the last one first, with the naziv and the JM
// of their artikal.
func (r *RobnoRepository) RproStavkeDokumenta(ctx context.Context, rdokID int64) ([]domain.RobStavkaRow, error) {
	b := querybuilder.New(`select r.rproid, coalesce(r.rdokid, 0)::bigint as rdokid,
			coalesce(r.rbr, 0) as rbr, coalesce(r.sifra, 0) as sifra, coalesce(r.konto, '') as konto,
			coalesce(nullif(r.naz1, ''), s.naziv, '') as naziv, coalesce(r.naz1, '') as naz1, coalesce(nullif(r.jm, ''), s.jm, '') as jm,
			coalesce(r.kolic, 0)::float8 as kolic, coalesce(r.cena, 0)::float8 as cena, coalesce(r.fcena, 0)::float8 as fcena,
			coalesce(r.mcena, 0)::float8 as mcena, coalesce(r.mcenap, 0)::float8 as mcenap, coalesce(r.ncena, 0)::float8 as ncena,
			coalesce(r.rab, 0)::float8 as rab, coalesce(r.po, 0) as po, coalesce(r.dani, 0) as dani, coalesce(r.mag, 0) as mag,
			coalesce(r.iznos, 0)::float8 as iznos, (coalesce(r.kolic, 0) * coalesce(r.fcena, 0))::float8 as iznosfakpr,
			coalesce(r.cenaval, 0)::float8 as cenaval, round(coalesce(r.kolic, 0) * coalesce(r.cenaval, 0), 2)::float8 as iznval,
			coalesce(r.taxcat, '') as taxcat, coalesce(r.taxexreco, '') as taxexreco, coalesce(r.otk, '') as otk,
			coalesce(r.serija, '') as serija, coalesce(r.roktr, 0)::float8 as roktr, coalesce(r.rsifid, 0)::bigint as rsifid,
			coalesce(r.rproid1, 0)::bigint as rproid1, coalesce(r.vma, 0)::float8 as vma, coalesce(r.mma, 0)::float8 as mma,
			coalesce(r.vra, 0)::float8 as vra, coalesce(r.mra, 0)::float8 as mra, coalesce(r.pcena, 0)::float8 as pcena,
			coalesce(r.itaksa, 0)::float8 as itaksa, `+poreskaStopa("r.po", "r.dadok")+` as stopa,
			coalesce(r.ptaksa, 0)::float8 as ptaksa, coalesce(r.ztrof, 0)::float8 as ztrof, coalesce(r.ztrin, 0)::float8 as ztrin,
			coalesce(r.ztro, 0)::float8 as ztro, coalesce(r.tz, '') as tz, coalesce(r.tzi, '') as tzi
		from rpro r`).
		AddJoin("left join rsif s on s.god = r.god and s.kar = r.kar and s.sifra = r.sifra").
		Where("r.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "r."); err != nil {
		return nil, err
	}
	b.OrderBy("r.rbr desc")
	var rows []domain.RobStavkaRow
	if err := r.selectRows(ctx, "stavke dokumenta", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// RdokKontekst returns what the stavke of a document need of it, its nalog, its magacin and its
// vrsta; no row when it is not a document of the period.
func (r *RobnoRepository) RdokKontekst(ctx context.Context, rdokID int64) ([]domain.RobStavkeKontekst, error) {
	b := querybuilder.New(`select d.rdokid, coalesce(d.rnalid, 0)::bigint as rnalid,
			coalesce(d.tipdok, '') as tipdok, coalesce(d.nalog, 0) as nalog, d.danal, coalesce(d.vrd, 0) as vrd,
			coalesce(v.dokvrstaid, 0)::bigint as dokvrstaid, coalesce(d.dokum, 0) as dokum, d.dadok,
			coalesce(d.fkto, '') as fkto, coalesce(d.fana, '') as fana, coalesce(d.kurs, 0)::float8 as kurs,
			coalesce(d.vaziza, 0)::float8 as vaziza, coalesce(d.magaciniid, 0) as magaciniid, coalesce(d.mag, 0) as mag,
			coalesce(m.tipzal, 0) as tipzal, coalesce(v.predznak, '') as predznak, coalesce(v.kodknj, '') as kodknj,
			coalesce(m.nacvodzal, 0) as nacvodzal, coalesce(m.tipmag, '') as tipmag, coalesce(m.tipart, '') as tipart,
			coalesce(d.sifval, 0) as sifval, coalesce(f.sifval, 0) as domacavaluta, coalesce(val.oznval, '') as valuta
		from rdok d`).
		AddJoin("left join magacini m on m.magaciniid = d.magaciniid").
		AddJoin("left join dokvrsta v on v.god = d.god and v.kar = d.kar and v.vrd = d.vrd").
		AddJoin("left join lateral (select x.sifval from fvr x where x.god = d.god and x.kar = d.kar limit 1) f on true").
		AddJoin("left join lateral (select x.oznval from valute x where x.sifval = d.sifval order by x.idvalute limit 1) val on true").
		Where("d.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	var rows []domain.RobStavkeKontekst
	if err := r.selectRows(ctx, "kontekst stavki", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// RsifArtikalStanje returns the artikal of a šifra with its stanje in a konto (trimmed) of a magacin
// and its veleprodajna cena; no row when the artikal does not exist.
func (r *RobnoRepository) RsifArtikalStanje(ctx context.Context, magaciniID int, konto string, sifra int) ([]domain.RobArtikalStanjeDto, error) {
	b := querybuilder.New(`select a.rsifid, a.sifra, coalesce(a.naziv, '') as naziv,
			coalesce(a.jm, '') as jm, coalesce(a.tip, '') as tip, coalesce(a.po, 0) as po,
			coalesce(st.stanje, 0)::float8 as stanje, coalesce(st.vpcena, 0)::float8 as vpcena
		from rsif a`).
		AddJoin(`left join lateral (select sum(coalesce(r.ulaz, 0) - coalesce(r.izlaz, 0)) as stanje, max(r.vpcena) as vpcena
			from rsta r where r.god = a.god and r.kar = a.kar and r.magaciniid = ? and r.konto = ? and r.sifra = a.sifra) st on true`,
			magaciniID, strings.TrimSpace(konto))
	if err := inPeriod(ctx, b, "a."); err != nil {
		return nil, err
	}
	b.Where("a.sifra = ?", sifra).Limit(1)
	var rows []domain.RobArtikalStanjeDto
	if err := r.selectRows(ctx, "artikal i stanje", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// MagkontoKonta returns the konta of a magacin (trimmed, not empty, in their order).
func (r *RobnoRepository) MagkontoKonta(ctx context.Context, magaciniID int) ([]string, error) {
	b := querybuilder.New(`select distinct trim(konto) from magkonto`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return nil, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("coalesce(trim(konto), '') <> ''").OrderBy("1")
	var konta []string
	if err := r.selectRows(ctx, "konta magacina", &konta, b); err != nil {
		return nil, err
	}
	return konta, nil
}

// RproZadnjaCenaKupca returns the last cena of the artikal to the kupac of a document: the prodajna
// cena, the rabat and the datum of the last stavka of the kupac with the šifra in a document of the
// same vrsta up to the datum of the document; no row without it.
func (r *RobnoRepository) RproZadnjaCenaKupca(ctx context.Context, rdokID int64, sifra int) ([]domain.RobZadnjaCenaDto, error) {
	b := querybuilder.New(`select coalesce(r.fcena, 0)::float8 as fcena, coalesce(r.rab, 0)::float8 as rab, r.dadok from rdok d`).
		AddJoin(`join rpro r on r.god = d.god and r.kar = d.kar and r.fkto = d.fkto and r.fana = d.fana and r.vrd = d.vrd
			and r.dadok <= d.dadok`).
		Where("d.rdokid = ?", rdokID).Where("r.sifra = ?", sifra)
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	b.OrderBy("r.dadok desc, r.rdokid desc, r.rproid desc").Limit(1)
	var rows []domain.RobZadnjaCenaDto
	if err := r.selectRows(ctx, "zadnja cena kupca", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// RsifIDJmDani reads (and locks) the id, the JM and the dani of an artikal by its šifra; the error
// wraps sql.ErrNoRows when there is none.
func (r *RobnoRepository) RsifIDJmDani(ctx context.Context, tx db.Transaction, sifra int) (rsifID int64, jm string, dani int, err error) {
	b := querybuilder.New(`select rsifid, coalesce(jm, ''), coalesce(dani, 0) from rsif`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, "", 0, err
	}
	b.Where("sifra = ?", sifra).ForUpdate()
	err = r.row(ctx, tx, "artikal", b, &rsifID, &jm, &dani)
	return rsifID, jm, dani, err
}

// RproBrojNaloga counts the stavke of a nalog.
func (r *RobnoRepository) RproBrojNaloga(ctx context.Context, tx db.Transaction, rnalID int64) (int, error) {
	b := querybuilder.New(`select count(*) from rpro`).Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	var n int
	err := r.row(ctx, tx, "stavke naloga", b, &n)
	return n, err
}

// RnalDugPot reads (and locks) the duguje and the potražuje of a nalog; the error wraps sql.ErrNoRows
// when it is not a nalog of the period.
func (r *RobnoRepository) RnalDugPot(ctx context.Context, tx db.Transaction, rnalID int64) (dug, pot float64, err error) {
	b := querybuilder.New(`select coalesce(dug, 0)::float8, coalesce(pot, 0)::float8 from rnal`).Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, err
	}
	b.ForUpdate()
	err = r.row(ctx, tx, "nalog", b, &dug, &pot)
	return dug, pot, err
}

// RsifArtikalStavke returns the artikal by its šifra, or by its barkod (poBarkodu), for a stavka:
// key = šifra, c1 = naziv, c2 = JM, c3 = poreska oznaka, c4 = dani, c5 = its konto (else the konto
// given), c6 = tip, c7 = its stopa on the datum, c8 = the cena of the proizvoda; no row without it.
func (r *RobnoRepository) RsifArtikalStavke(ctx context.Context, poBarkodu bool, vrednost, konto string, datum time.Time) (*[]domain.SearchRowDto, error) {
	kolona := "a.sifra::text"
	if poBarkodu {
		kolona = "a.barkod"
	}
	b := querybuilder.New(`select a.sifra::text as key, coalesce(a.naziv, '') as c1,
			coalesce(a.jm, '') as c2, coalesce(a.po, 0)::text as c3, coalesce(a.dani, 0)::text as c4,
			coalesce(nullif(trim(a.konto), ''), ?) as c5, coalesce(a.tip, '') as c6,
			`+poreskaStopa("a.po", "?::date")+`::text as c7,
			coalesce((select c.pcena from rcene c where c.god = a.god and c.kar = a.kar and c.sifra = a.sifra limit 1), 0)::float8::text as c8
		from rsif a`, strings.TrimSpace(konto), datum)
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

// RstaStavke is the stanje of an artikal with the cene a stavka of a faktura uses.
type RstaStavke struct {
	Stanje, Vpcena, Cena, Ncena, Prosnc float64
}

// RstaStanjeStavke returns the stanje of an artikal in a konto of a magacin with its veleprodajna,
// magacinska, nabavna and prosečna nabavna cena; found is false without it.
func (r *RobnoRepository) RstaStanjeStavke(ctx context.Context, magaciniID int, konto string, sifra int) (RstaStavke, bool, error) {
	b := querybuilder.New(`select (coalesce(ulaz, 0) - coalesce(izlaz, 0))::float8,
			coalesce(vpcena, 0)::float8, coalesce(cena, 0)::float8, coalesce(ncena, 0)::float8, coalesce(prosnc, 0)::float8
		from rsta`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return RstaStavke{}, false, err
	}
	b.Where("magaciniid = ?", magaciniID).Where("konto = ?", konto).Where("sifra = ?", sifra).OrderBy("rstaid").Limit(1)
	var st RstaStavke
	err := r.row(ctx, nil, "stanje", b, &st.Stanje, &st.Vpcena, &st.Cena, &st.Ncena, &st.Prosnc)
	if errors.Is(err, sql.ErrNoRows) {
		return RstaStavke{}, false, nil
	}
	return st, err == nil, err
}

// RproStavkaForma returns a saved stavka of a faktura for its form: key = rproid, c1 = šifra,
// c2 = naziv, c3 = konto, c4 = JM, c5 = količina|fakturna cena|rabat|poreska oznaka|dani|nabavna
// cena|magacinska cena|rok trajanja|stopa|stanje of the artikal, c6 = OTK broj, c7 = serija,
// c8 = poreska kategorija; no row without it.
func (r *RobnoRepository) RproStavkaForma(ctx context.Context, rproID, rdokID int64) (*[]domain.SearchRowDto, error) {
	b := querybuilder.New(`select r.rproid::text as key, coalesce(r.sifra, 0)::text as c1,
			coalesce(nullif(r.naz1, ''), a.naziv, '') as c2, coalesce(r.konto, '') as c3, coalesce(r.jm, '') as c4,
			concat_ws('|', coalesce(r.kolic, 0)::float8, coalesce(r.fcena, 0)::float8, coalesce(r.rab, 0)::float8,
				coalesce(r.po, 0), coalesce(r.dani, 0), coalesce(r.ncena, 0)::float8, coalesce(r.cena, 0)::float8,
				coalesce(r.roktr, 0)::bigint,
				`+poreskaStopa("r.po", "r.dadok")+`,
				coalesce((select sum(coalesce(t.ulaz, 0) - coalesce(t.izlaz, 0)) from rsta t
					where t.god = r.god and t.kar = r.kar and t.magaciniid = r.magaciniid and t.konto = r.konto and t.sifra = r.sifra), 0)::float8) as c5,
			coalesce(r.otk, '') as c6, coalesce(r.serija, '') as c7, coalesce(r.taxcat, '') as c8
		from rpro r`).
		AddJoin("left join rsif a on a.god = r.god and a.kar = r.kar and a.sifra = r.sifra").
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

// RproTaxexreco returns the osnov oslobođenja of a stavka; "" without it.
func (r *RobnoRepository) RproTaxexreco(ctx context.Context, rproID int64) (string, error) {
	b := querybuilder.New(`select coalesce(taxexreco, '') from rpro`).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", err
	}
	var v string
	err := r.row(ctx, nil, "osnov oslobođenja", b, &v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}
