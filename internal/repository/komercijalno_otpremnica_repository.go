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

// The otpremnice of the komercijalno (the legacy pr_Otpremnica): edok of the vrste of the robne fakture of a potvrđen
// nalog za isporuku (vrdokid) with its stavke in epro copied from the nalog (rproid1). The potvrda and its robna
// faktura are in komercijalno_otpremnica_potvrda_repository.go.

// komOtpVrstaUslov restricts the dokvrsta v to the vrste of the otpremnice (the legacy PR_QRY_DOKVRSTA with
// "PRO;SVI;ROB" and "FPR;FAK;PRE;FZR;FRP").
const komOtpVrstaUslov = `upper(trim(coalesce(v.modul, ''))) in ('PRO', 'SVI', 'ROB')
	and upper(trim(coalesce(v.dokozn, ''))) in ('FPR', 'FAK', 'PRE', 'FZR', 'FRP')`

// KomOtpVrste returns the vrste of the otpremnice of the period in the order of the vrd.
func (r *KomercijalnoRepository) KomOtpVrste(ctx context.Context) ([]domain.KomOtpVrsta, error) {
	b := querybuilder.New(`select v.dokvrstaid::bigint as dokvrstaid, coalesce(v.vrd, 0)::int as vrd, coalesce(v.opis, '') as opis,
			upper(trim(coalesce(v.dokozn, ''))) as dokozn, trim(coalesce(v.predznak, '')) as predznak,
			upper(trim(coalesce(v.kodknj, ''))) as kodknj, coalesce(v.flgpopdv, false) as flgpopdv, trim(coalesce(v.polje, '')) as polje
		from dokvrsta v`).
		Where(komOtpVrstaUslov)
	if err := inPeriod(ctx, b, "v."); err != nil {
		return nil, err
	}
	b.OrderBy("v.vrd")
	var rows []domain.KomOtpVrsta
	if err := r.selectRows(ctx, "vrste otpremnica", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomOtpremnica returns an otpremnica of the period with the texts of its šifre (magacin, vrsta naloga za knjiženje, org.
// jedinica, mesto troška, komercijalista, mesto isporuke, valuta) and the broj and datum of its nalog; found is false
// without it.
func (r *KomercijalnoRepository) KomOtpremnica(ctx context.Context, rdokID int64) (domain.KomOtpremnicaRow, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.vrd, 0)::int, coalesce(e.dokum, 0)::int, e.dadok, coalesce(e.brotp, 0)::int,
			e.datotp, e.datiz, coalesce(e.magaciniid, 0)::int, coalesce(e.mag, 0)::int, coalesce(e.mag::text || '-' || coalesce(m.opis, ''), ''),
			coalesce(m.adresa, ''), coalesce(m.mesto, ''), upper(trim(coalesce(m.tipmag, ''))), coalesce(m.tipzal, 0)::int,
			coalesce(m.nacvodzal, 0)::int, trim(coalesce(e.tipdok, '')),
			coalesce((select trim(t.tipdok) || '-' || coalesce(t.opis, '') from tipdok t where t.tipdok = e.tipdok and t.god = e.god
				and t.kar = e.kar limit 1), ''),
			e.idorgjed::bigint, coalesce(o.ojozn::text || '-' || coalesce(o.naziv, ''), ''), e.mestotrid::bigint,
			coalesce(mt.mtroska || '-' || coalesce(mt.opis, ''), ''), coalesce(e.fkto, ''), coalesce(e.fana, ''), e.komid::bigint,
			coalesce(k.imeprezime, ''), coalesce(e.rok, 0)::int, coalesce(e.pla, ''), coalesce(e.kamt, ''), coalesce(e.mi, 0)::int,
			coalesce((select f.naziv from fisp f where f.konto = e.fkto and f.sifra = e.fana and f.mi = e.mi and f.god = e.god
				and f.kar = e.kar limit 1), ''),
			coalesce(e.otp, ''), coalesce(e.vozac, ''), coalesce(e.brvozila, ''), coalesce(e.bttowght, 0)::float8,
			coalesce(e.netwght, 0)::float8, coalesce(e.totboxs, 0)::int, coalesce(e.brpaleta, 0)::int, coalesce(e.foot, ''),
			coalesce(e.napomena, ''), coalesce(e.opis, ''), coalesce(e.sifval, 0)::int,
			coalesce((select v.oznval from valute v where v.sifval = e.sifval order by v.idvalute limit 1), ''), coalesce(e.kurs, 0)::float8,
			coalesce(e.vaziza, 0)::int, coalesce(e.keykat, ''), coalesce(e.category, ''), coalesce(e.pornapomena, ''),
			coalesce(e.brugov, ''), coalesce(e.polje, ''), coalesce(e.posuslnab, ''), coalesce(e.ugrabat, 0)::float8,
			coalesce(e.pkase, 0)::float8, coalesce(e.doksts, ''), coalesce(e.knjige_1, ''), coalesce(e.vrdokid, 0)::bigint,
			coalesce(n.dokum, 0)::int, n.dadok
		from edok e`).
		AddJoin("join dokvrsta v on v.vrd = e.vrd and v.god = e.god and v.kar = e.kar").
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr mt on mt.mestotrid = e.mestotrid").
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		AddJoin("left join edok n on n.rdokid = e.vrdokid and n.god = e.god and n.kar = e.kar").
		Where("e.rdokid = ?", rdokID).
		Where(komOtpVrstaUslov)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomOtpremnicaRow{}, false, err
	}
	var o domain.KomOtpremnicaRow
	err := r.row(ctx, nil, "otpremnica", b, &o.RdokID, &o.Vrd, &o.Dokum, &o.Dadok, &o.Brotp, &o.Datotp, &o.Datiz, &o.MagaciniID, &o.Mag,
		&o.MagacinText, &o.MagacinAdresa, &o.MagacinMesto, &o.Tipmag, &o.Tipzal, &o.Nacvodzal, &o.Tipdok, &o.TipdokText, &o.Orgjed,
		&o.OrgjedText, &o.MestoTroska, &o.MestoTroskaText, &o.Fkto, &o.Fana, &o.KomID, &o.KomIme, &o.Rok, &o.Pla, &o.Kamt, &o.Mi,
		&o.MestoIsporuke, &o.Otp, &o.Vozac, &o.Brvozila, &o.Bttowght, &o.Netwght, &o.Totboxs, &o.Brpaleta, &o.Foot, &o.Napomena, &o.Opis,
		&o.SifVal, &o.OznVal, &o.Kurs, &o.VaziZa, &o.Keykat, &o.Category, &o.Pornapomena, &o.Brugov, &o.Polje, &o.Posuslnab, &o.Ugrabat,
		&o.Pkase, &o.Doksts, &o.Knjige1, &o.VrdokID, &o.NalogDokum, &o.NalogDadok)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomOtpremnicaRow{}, false, nil
	}
	if err != nil {
		return domain.KomOtpremnicaRow{}, false, err
	}
	return o, true, nil
}

// KomOtpNalog returns a nalog za isporuku of the period (edok of the vrste of the group NRK) with what the header of an
// otpremnica takes of it (the legacy PodaciHDR); found is false without it.
func (r *KomercijalnoRepository) KomOtpNalog(ctx context.Context, nalogID int64) (domain.KomOtpNalog, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.dokum, 0)::int, e.dadok, coalesce(e.magaciniid, 0)::int, coalesce(e.doksts, ''),
			coalesce(e.fkto, ''), coalesce(e.fana, ''),
			coalesce((select trim(t.tipdok) || '-' || coalesce(t.opis, '') from tipdok t where t.tipdok = e.tipdok and t.god = e.god
				and t.kar = e.kar limit 1), ''),
			coalesce(o.ojozn::text || '-' || coalesce(o.naziv, ''), ''), coalesce(mt.mtroska || '-' || coalesce(mt.opis, ''), ''),
			coalesce(e.posuslnab, ''), coalesce(e.sifval, 0)::int,
			coalesce((select v.oznval from valute v where v.sifval = e.sifval order by v.idvalute limit 1), ''), coalesce(e.kurs, 0)::float8,
			coalesce(e.vaziza, 0)::int, coalesce(e.foot, ''), coalesce(e.napomena, ''), coalesce(e.vozac, ''), coalesce(e.brvozila, ''),
			coalesce(e.otp, ''), coalesce(e.bttowght, 0)::float8, coalesce(e.netwght, 0)::float8, coalesce(e.totboxs, 0)::int,
			coalesce(e.brpaleta, 0)::int, coalesce(e.keykat, ''), coalesce(e.category, ''), coalesce(e.pornapomena, ''), coalesce(e.brugov, '')
		from edok e`).
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr mt on mt.mestotrid = e.mestotrid").
		Where("e.rdokid = ?", nalogID).
		Where(komProVrstaUslov, komZahtevModul, komProNalogGrupa)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomOtpNalog{}, false, err
	}
	var n domain.KomOtpNalog
	err := r.row(ctx, nil, "nalog otpremnice", b, &n.RdokID, &n.Dokum, &n.Dadok, &n.MagaciniID, &n.Doksts, &n.Fkto, &n.Fana, &n.TipdokText,
		&n.OrgjedText, &n.MestoTroskaText, &n.Posuslnab, &n.SifVal, &n.OznVal, &n.Kurs, &n.VaziZa, &n.Foot, &n.Napomena, &n.Vozac, &n.Brvozila,
		&n.Otp, &n.Bttowght, &n.Netwght, &n.Totboxs, &n.Brpaleta, &n.Keykat, &n.Category, &n.Pornapomena, &n.Brugov)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomOtpNalog{}, false, nil
	}
	if err != nil {
		return domain.KomOtpNalog{}, false, err
	}
	return n, true, nil
}

// KomOtpStavke returns the stavke of a document of the period (an otpremnica or its nalog) in the order of their redni
// broj with the naziv, the JM (its decimals, whether it exists and has a length or a width), the komercijalni opis, the
// proizvođač, the tip and the težina of their artikal (the legacy PR_QRY_EPRO).
func (r *KomercijalnoRepository) KomOtpStavke(ctx context.Context, rdokID int64) ([]domain.KomOtpStavkaRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, coalesce(p.rproid1, 0)::bigint as rproid1, coalesce(p.rbr, 0)::int as rbr,
			coalesce(p.brst, 0)::int as brst, trim(coalesce(p.konto, '')) as konto, coalesce(p.sifra, 0)::int as sifra,
			coalesce(p.rsifid, 0)::bigint as rsifid, coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm,
			coalesce(s.komercopis, '') as komercopis, coalesce(s.pro, '') as pro, coalesce(s.proizsifra, '') as proizsifra,
			upper(trim(coalesce(s.tip, ''))) as tip, coalesce(p.otk, '') as otk, coalesce(p.serija, '') as serija,
			coalesce(p.roktr, 0)::bigint as roktr, coalesce(p.trazkolic, 0)::float8 as trazkolic, coalesce(p.kolic, 0)::float8 as kolic,
			coalesce(p.kolic1, 0)::float8 as kolic1, coalesce(p.cena, 0)::float8 as cena, coalesce(p.ncena, 0)::float8 as ncena,
			coalesce(p.pcena, 0)::float8 as pcena, coalesce(p.fcena, 0)::float8 as fcena, coalesce(p.cenaval, 0)::float8 as cenaval,
			coalesce(p.mcenap, 0)::float8 as mcenap, coalesce(p.rab, 0)::float8 as rab, coalesce(p.po, 0)::int as po,
			coalesce(s.tezjm, 0)::float8 as tezjm, coalesce(j.brdecimala, 2)::int as decimale, j.jm is not null as jmpostoji,
			coalesce(j.ima_duzinu, false) as imaduzinu, coalesce(j.ima_sirinu, false) as imasirinu
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		AddJoin(`left join lateral (select x.jm, x.brdecimala, x.ima_duzinu, x.ima_sirinu from jedmere x where x.jm = s.jm and x.god = p.god
			and x.kar = p.kar limit 1) j on true`).
		Where("p.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr").OrderBy("p.rproid")
	var rows []domain.KomOtpStavkaRow
	if err := r.selectRows(ctx, "stavke otpremnice", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomOtpPrivremeniBrojevi returns the next temporary broj otpremnice and broj fakture of the vrsta in the magacin (mag) of
// the period (the legacy NoviBrojEDOK: the largest one above 990000 + 1, else 990001).
func (r *KomercijalnoRepository) KomOtpPrivremeniBrojevi(ctx context.Context, tx db.Transaction, vrd, mag int) (brotp, dokum int, err error) {
	b := querybuilder.New(`select coalesce(max(brotp) filter (where brotp > ?), ?)::int, coalesce(max(dokum) filter (where dokum > ?), ?)::int
		from edok`, domain.KomOtpPrivremeniBroj, domain.KomOtpPrviPrivremeni-1, domain.KomOtpPrivremeniBroj, domain.KomOtpPrviPrivremeni-1).
		Where("vrd = ?", vrd).
		Where("mag = ?", mag)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, 0, err
	}
	if err := r.row(ctx, tx, "privremeni brojevi otpremnice", b, &brotp, &dokum); err != nil {
		return 0, 0, err
	}
	return brotp + 1, dokum + 1, nil
}

// KomOtpBrotpPostoji tells whether another otpremnica (or document) of the vrsta in the magacin (mag) of the period has
// the broj otpremnice (the legacy GODKARVRDMAGBROTP: "Otpremnica sa ovim brojem već postoji...").
func (r *KomercijalnoRepository) KomOtpBrotpPostoji(ctx context.Context, vrd, mag, brotp int, rdokID int64) (bool, error) {
	b := querybuilder.New(`select count(*) from edok`).
		Where("vrd = ?", vrd).
		Where("mag = ?", mag).
		Where("brotp = ?", brotp).
		Where("rdokid <> ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	var n int
	if err := r.row(ctx, nil, "broj otpremnice", b, &n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomOtpRdokPostoji tells whether a robni dokument of the vrsta in the magacin (mag) of the period has the broj (the
// legacy GODKARMAGVRDDOKUM: "Dokument sa brojem: N već postoji!!!!").
func (r *KomercijalnoRepository) KomOtpRdokPostoji(ctx context.Context, tx db.Transaction, mag, vrd, dokum int) (bool, error) {
	b := querybuilder.New(`select count(*) from rdok`).
		Where("mag = ?", mag).
		Where("vrd = ?", vrd).
		Where("dokum = ?", dokum)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	var n int
	if err := r.row(ctx, tx, "broj fakture", b, &n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomOtpDodaj adds an otpremnica of a nalog of the period (the legacy SnimiEDOK of a new one): new (doksts 10), not posted
// (knjige N), not of a robni nalog, its opis "Otpremnica br: <broj>", the vrsta, the brojevi, the datumi (datiz the datum
// otpremnice), the fields of the screen, and the other ones of the nalog (magacin, rok, kupac, uslovi, ugovoreni rabat,
// kasa, iznos, komercijalista, kamata, mesto isporuke, vrsta naloga za knjiženje, org. jedinica, mesto troška, cenovnik,
// prodajna vrednost, kurs firme); vrdokid is the nalog. It returns its rdokid.
func (r *KomercijalnoRepository) KomOtpDodaj(ctx context.Context, tx db.Transaction, p domain.KomOtpUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into edok (god, kar, rnalid, vrd, dokum, dadok, mag, magaciniid, iznos, rbr, opis, rok, fkto, fana, brst,
			brotp, datotp, datiz, vrdokid, otp, pla, posuslnab, vrugrabat, ugrabat, vkase, pkase, sifval, kurs, vaziza, komid, kamt, brugov, mi,
			tipdok, idorgjed, mestotrid, foot, brvozila, vozac, bttowght, netwght, totboxs, brpaleta, rcenovnikhdrid, doksts, knjige_1, knjige_2,
			prodvred, kursfirme, napomena, polje, keykat, category, pornapomena, jm1, jm2, xdatunosa, xopunos)
		select n.god, n.kar, 0, ?, ?, ?, n.mag, n.magaciniid, n.iznos, 0, ?, n.rok, n.fkto, n.fana, 0, ?, ?, ?, n.rdokid, ?, n.pla, ?,
			n.vrugrabat, n.ugrabat, n.vkase, n.pkase, ?, ?, ?, n.komid, n.kamt, ?, n.mi, n.tipdok, n.idorgjed, n.mestotrid, ?, ?, ?, ?, ?, ?, ?,
			n.rcenovnikhdrid, ?, 'N', 'N', n.prodvred, n.kursfirme, ?, ?, ?, ?, ?, '', '', now(), ?
		from edok n`,
		p.Vrd, p.Dokum, p.Dadok, fmt.Sprintf("Otpremnica br: %d", p.Brotp), p.Brotp, p.Datotp, p.Datotp, p.Otp, p.Posuslnab, p.SifVal, p.Kurs,
		p.VaziZa, p.Brugov, p.Foot, p.Brvozila, p.Vozac, p.Bttowght, p.Netwght, p.Totboxs, p.Brpaleta, domain.KomOtpNova, p.Napomena,
		p.Polje, p.Keykat, p.Category, p.Pornapomena, s.UserName).
		Where("n.rdokid = ?", p.NalogID).
		Returning("rdokid")
	if err := inPeriod(ctx, b, "n."); err != nil {
		return 0, err
	}
	var rdokID int64
	if err := r.row(ctx, tx, "nova otpremnica", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// KomOtpIzmeni saves the header of an otpremnica of the period (the legacy SnimiEDOK of a saved one): the brojevi, the
// datumi, the transport data, the tekst and the napomena, the valuta with its kurs, the žiro račun, the POPDV polje, the ugovor and the
// poreska kategorija (its vrsta, magacin and nalog stay), with the audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomOtpIzmeni(ctx context.Context, tx db.Transaction, p domain.KomOtpUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("edok").
		Set("brotp = ?", p.Brotp).
		Set("dokum = ?", p.Dokum).
		Set("opis = ?", fmt.Sprintf("Otpremnica br: %d", p.Brotp)).
		Set("dadok = ?", p.Dadok).
		Set("datotp = ?", p.Datotp).
		Set("datiz = ?", p.Datotp).
		Set("vozac = ?", p.Vozac).
		Set("brvozila = ?", p.Brvozila).
		Set("bttowght = ?", p.Bttowght).
		Set("netwght = ?", p.Netwght).
		Set("totboxs = ?", p.Totboxs).
		Set("brpaleta = ?", p.Brpaleta).
		Set("otp = ?", p.Otp).
		Set("foot = ?", p.Foot).
		Set("napomena = ?", p.Napomena).
		Set("sifval = ?", p.SifVal).
		Set("kurs = ?", p.Kurs).
		Set("vaziza = ?", p.VaziZa).
		Set("posuslnab = ?", p.Posuslnab).
		Set("polje = ?", p.Polje).
		Set("brugov = ?", p.Brugov).
		Set("keykat = ?", p.Keykat).
		Set("category = ?", p.Category).
		Set("pornapomena = ?", p.Pornapomena).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", p.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "otpremnica", b)
}

// KomOtpStavkeKopiraj adds the stavke of the nalog of the period an otpremnica does not have yet (the legacy PuniDETALJ:
// a stavka of the nalog is linked by rproid1): the artikal with its lot and konto, the tražena and the isporučena količina
// (the one of the nalog, negative for a storno vrsta), the cene (the prodajna cena of the cena u valuti at the kurs and
// the "važi za" of the otpremnica with preracun, else the one of the nalog), the rabat, the vrsta, the broj and the datum
// of the otpremnica, with the audit columns.
func (r *KomercijalnoRepository) KomOtpStavkeKopiraj(ctx context.Context, tx db.Transaction, otpID, nalogID int64, storno, preracun bool) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	znak := 1
	if storno {
		znak = -1
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, konto, sifra, po, rsifid, rbr, mag, magaciniid, cena, ncena, fcena, cenaval,
			mcenap, trazkolic, kolic, vrd, dokum, dadok, rab, otk, serija, roktr, rproid1, xdatunosa, xopunos)
		select p.god, p.kar, o.rdokid, p.konto, p.sifra, p.po, p.rsifid, p.rbr, p.mag, p.magaciniid, p.cena, p.ncena,
			case when ? then round((coalesce(p.cenaval, 0) * coalesce(o.kurs, 1) / coalesce(nullif(o.vaziza, 0), 1))::numeric, 2) else p.fcena end,
			p.cenaval, p.mcenap, ? * abs(coalesce(p.trazkolic, 0)), ? * abs(coalesce(p.trazkolic, 0)), o.vrd, o.dokum, o.dadok, p.rab, p.otk,
			p.serija, p.roktr, p.rproid, now(), ?
		from epro p`, preracun, znak, znak, s.UserName).
		AddJoin("join edok o on o.rdokid = ? and o.god = p.god and o.kar = p.kar", otpID).
		Where("p.rdokid = ?", nalogID).
		Where(`not exists (select 1 from epro x where x.rdokid = o.rdokid and x.rproid1 = p.rproid and x.god = p.god and x.kar = p.kar)`)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return err
	}
	_, err = r.exec(ctx, tx, "stavke otpremnice", b)
	return err
}

// KomOtpStavkePreracunaj sets the prodajna cena of the stavke of an otpremnica of the period to the cena u valuti at its
// kurs and "važi za" (the legacy PuniDETALJ with a changed kurs).
func (r *KomercijalnoRepository) KomOtpStavkePreracunaj(ctx context.Context, tx db.Transaction, otpID int64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("epro").
		Set("fcena = round((coalesce(epro.cenaval, 0) * coalesce(o.kurs, 1) / coalesce(nullif(o.vaziza, 0), 1))::numeric, 2)").
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		From("edok o").
		Where("o.rdokid = epro.rdokid").
		Where("o.god = epro.god").
		Where("o.kar = epro.kar").
		Where("epro.rdokid = ?", otpID)
	if err := inPeriod(ctx, b, "epro."); err != nil {
		return err
	}
	_, err = r.exec(ctx, tx, "cene stavki otpremnice", b)
	return err
}

// KomOtpProdvred saves the prodajna vrednost of an otpremnica of the period (the legacy PreracunFAKTURAPROIZVGLOBEDOK);
// ErrNotFound without it.
func (r *KomercijalnoRepository) KomOtpProdvred(ctx context.Context, tx db.Transaction, otpID int64, prodvred float64) error {
	b := querybuilder.NewUpdate("edok").Set("prodvred = ?", prodvred).Where("rdokid = ?", otpID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "prodajna vrednost otpremnice", b)
}

// KomOtpStavkaIzmeni changes the isporučena količina and the lot of a stavka of an otpremnica of the period (the legacy
// exit of a changed row of TABLE_PR_QRY_EPRO), with the audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomOtpStavkaIzmeni(ctx context.Context, tx db.Transaction, otpID, rproID int64, kolic float64, otk, serija string,
	roktr int64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("epro").
		Set("kolic = ?", kolic).
		Set("otk = ?", otk).
		Set("serija = ?", serija).
		Set("roktr = ?", roktr).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rproid = ?", rproID).
		Where("rdokid = ?", otpID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stavka otpremnice", b)
}

// KomOtpNalogFormiran marks the potvrđen nalog of the period of a new otpremnica (doksts 40; the legacy SnimiEDOK "PRONADJI
// NALOG ZA ISPORUKU I PROMENI MU STATUS"). It returns false when the nalog is not a potvrđen one (any more).
func (r *KomercijalnoRepository) KomOtpNalogFormiran(ctx context.Context, tx db.Transaction, nalogID int64) (bool, error) {
	return r.komProStatus(ctx, tx, nalogID, domain.KomNalogOtpremljen, domain.KomNalogPotvrdjen)
}

// KomOtpNalogVrati sets the nalog of the period of a deleted otpremnica back to potvrđen (doksts 20; the legacy
// BTN_Brisanje sets 10). It returns false when the nalog is not marked.
func (r *KomercijalnoRepository) KomOtpNalogVrati(ctx context.Context, tx db.Transaction, nalogID int64) (bool, error) {
	return r.komProStatus(ctx, tx, nalogID, domain.KomNalogPotvrdjen, domain.KomNalogOtpremljen)
}

// KomOtpPrikaz returns the otpremnice of the list of the main screen (the legacy PR_QRY_EDOK of "Pregled otpremnica"): the
// ones of the vrste of the otpremnice with the status (the nepotvrđene 10, else the potvrđene 40) in the magacini of the
// user of the session (maguser), of one magacin or of all of them (SviMagacini), the last ones first, with the "mag-opis"
// of the magacin, the "vrd-opis" of the vrsta, the broj of the nalog, the naziv of the kupac and of the mesto isporuke.
func (r *KomercijalnoRepository) KomOtpPrikaz(ctx context.Context, p domain.KomOtpPrikazParams) ([]domain.KomOtpPrikazRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select e.rdokid::bigint as rdokid, coalesce(m.mag::text || '-' || coalesce(m.opis, ''), '') as magacin,
			coalesce(e.vrd, 0)::int as vrd, coalesce(v.vrd::text || '-' || coalesce(v.opis, ''), '') as vrstaopis,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok, coalesce(e.brotp, 0)::int as brotp, coalesce(n.dokum, 0)::int as nalogdokum,
			coalesce(e.fkto, '') as fkto, coalesce(e.fana, '') as fana, coalesce(f.naziv, '') as kupac, coalesce(e.doksts, '') as doksts,
			coalesce(e.rok, 0)::int as rok, e.datiz as datiz, coalesce(e.opis, '') as opis,
			coalesce((select x.naziv from fisp x where x.konto = e.fkto and x.sifra = e.fana and x.mi = e.mi and x.god = e.god
				and x.kar = e.kar limit 1), '') as mestoisporuke,
			coalesce(e.kamt, '') as kamt
		from edok e`).
		AddJoin("join dokvrsta v on v.vrd = e.vrd and v.god = e.god and v.kar = e.kar").
		AddJoin("join (select distinct mu.mag from maguser mu where upper(mu.iduser) = ?) u on u.mag = e.mag",
			strings.ToUpper(strings.TrimSpace(s.UserName))).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join edok n on n.rdokid = e.vrdokid and n.god = e.god and n.kar = e.kar").
		AddJoin("left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana").
		Where(komOtpVrstaUslov)
	if p.Doksts == domain.KomOtpNova {
		b.Where("coalesce(e.doksts, '') = ?", domain.KomOtpNova)
	} else {
		b.Where("coalesce(e.doksts, '') = ?", domain.KomOtpPotvrdjena)
	}
	if !p.SviMagacini {
		b.Where("e.magaciniid = ?", p.MagaciniID)
	}
	komPretraga(b, p.Tekst, "e.dokum::text", "e.brotp::text", "e.foot", "f.naziv", "e.fana")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return nil, err
	}
	b.OrderByExpr("e.dadok desc nulls last, e.vrd desc, e.dokum desc")
	var rows []domain.KomOtpPrikazRow
	if err := r.selectRows(ctx, "pregled otpremnica", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}
