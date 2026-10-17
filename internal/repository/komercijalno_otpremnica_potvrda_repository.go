package repository

import (
	"context"
	"database/sql"
	"errors"

	"helia/internal/domain"
	"helia/internal/infrastructure/db"
	"helia/internal/infrastructure/db/querybuilder"
)

// The potvrda of an otpremnica of the komercijalno (the legacy pr_Otpremnica BTN_POTVRDADOK with SnimiRDOK): its robna
// faktura (rdok.edokid is the otpremnica) on a robni nalog of the vrsta naloga za knjiženje, the datum and the magacin
// (an "Aut. kreiran nalog" without one), its stavke in rpro, the missing cene of the artikli, the avansi of the nalog za
// isporuku, the final brojevi of the otpremnica and the isporučene količine of the nalog za isporuku. The posting of the
// stanje is the one of the robno service.

// KomOtpGodinaZatvorena tells whether the poslovna godina of the period is closed (fvr.godzatv; the legacy "Poslovna
// godina je zatvorena, knjiženje nije dozvoljeno...").
func (r *KomercijalnoRepository) KomOtpGodinaZatvorena(ctx context.Context) (bool, error) {
	b := querybuilder.New(`select coalesce(godzatv, false) from fvr`)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	b.Limit(1)
	var zatvorena bool
	err := r.row(ctx, nil, "zatvorena godina", b, &zatvorena)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return zatvorena, err
}

// KomOtpNoviBrojevi returns the next broj otpremnice of the vrsta in the magacin (mag) of the period (the largest one
// below 990000 + 1, the temporary ones not counted) and the next broj of its robna faktura (the largest broj of rdok of
// the vrsta in the magacin + 1; the legacy NoviBrojRDOK).
func (r *KomercijalnoRepository) KomOtpNoviBrojevi(ctx context.Context, tx db.Transaction, vrd, mag int) (brotp, dokum int, err error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	b := querybuilder.New(`select
			coalesce((select max(x.brotp) from edok x where x.god = ? and x.kar = ? and x.mag = ? and x.vrd = ? and x.brotp < ?), 0)::int + 1,
			coalesce((select max(d.dokum) from rdok d where d.god = ? and d.kar = ? and d.mag = ? and d.vrd = ?), 0)::int + 1`,
		s.God, s.Kar, mag, vrd, domain.KomOtpPrivremeniBroj, s.God, s.Kar, mag, vrd)
	if err := r.row(ctx, tx, "brojevi potvrđene otpremnice", b, &brotp, &dokum); err != nil {
		return 0, 0, err
	}
	return brotp, dokum, nil
}

// KomOtpRstaProsnc returns the prosečna nabavna cena of an artikal in a magacin of the period (the first rsta of the
// šifra; the legacy RSTA.PROSNC of GODKARMAGACINIIDSIFRA); 0 without it.
func (r *KomercijalnoRepository) KomOtpRstaProsnc(ctx context.Context, tx db.Transaction, magaciniID, sifra int) (float64, error) {
	b := querybuilder.New(`select coalesce(prosnc, 0)::float8 from rsta`).Where("magaciniid = ?", magaciniID).Where("sifra = ?", sifra)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, err
	}
	b.OrderBy("rstaid").Limit(1)
	var prosnc float64
	err := r.row(ctx, tx, "prosečna nabavna cena", b, &prosnc)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return prosnc, err
}

// KomOtpNetoTezina saves the neto težina of an otpremnica of the period (the legacy IzracunajTezinu); ErrNotFound without
// it.
func (r *KomercijalnoRepository) KomOtpNetoTezina(ctx context.Context, tx db.Transaction, otpID int64, tezina float64) error {
	b := querybuilder.NewUpdate("edok").Set("netwght = ?", tezina).Where("rdokid = ?", otpID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "neto težina otpremnice", b)
}

// KomOtpRnal returns the robni nalog of the period of the vrsta naloga za knjiženje, the datum and the magacin (the one
// with the largest broj; the legacy NoviBrojNALOGA); found is false without it.
func (r *KomercijalnoRepository) KomOtpRnal(ctx context.Context, tx db.Transaction, tipdok string, danal sql.NullTime, magaciniID int) (int64, bool, error) {
	b := querybuilder.New(`select rnalid::bigint from rnal`).
		Where("tipdok = ?", tipdok).
		Where("danal = ?", danal).
		Where("magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, false, err
	}
	b.OrderByExpr("nalog desc, rnalid desc").Limit(1)
	var rnalID int64
	err := r.row(ctx, tx, "robni nalog otpremnice", b, &rnalID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return rnalID, true, nil
}

// KomOtpRnalDodaj adds a robni nalog of the period (the legacy SnimiRDOK "KREIRAJ SLOG U RNAL"): "Aut. kreiran nalog" of the
// vrsta naloga za knjiženje (with its id) on the datum (danal and datob) in the magacin, the next broj of the vrsta naloga
// (the largest one of the period + 1), the iznosi and the counters 0, the operater, with the audit columns. It returns its
// rnalid.
func (r *KomercijalnoRepository) KomOtpRnalDodaj(ctx context.Context, tx db.Transaction, n domain.KomOtpRnalUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into rnal (god, kar, magaciniid, mag, dug, pot, opis, datob, danal, nalog, nalsts, oper, rbr, brst, brdo,
			idtipdok, tipdok, xdatunosa, xopunos)
		select ?, ?, ?, ?, 0, 0, 'Aut. kreiran nalog', ?, ?,
			coalesce((select max(x.nalog) from rnal x where x.tipdok = ? and x.god = ? and x.kar = ?), 0)::int + 1, '', ?, 0, 0, 0,
			(select t.idtipdok from tipdok t where t.tipdok = ? and t.god = ? and t.kar = ? limit 1), ?, now(), ?`,
		s.God, s.Kar, n.MagaciniID, n.Mag, n.Danal, n.Danal, n.Tipdok, s.God, s.Kar, s.UserName, n.Tipdok, s.God, s.Kar, n.Tipdok, s.UserName).
		Returning("rnalid")
	var rnalID int64
	if err := r.row(ctx, tx, "novi robni nalog otpremnice", b, &rnalID); err != nil {
		return 0, err
	}
	return rnalID, nil
}

// KomOtpRnalDokument counts one more document of a robni nalog of the period (rnal.brdo); ErrNotFound without it.
func (r *KomercijalnoRepository) KomOtpRnalDokument(ctx context.Context, tx db.Transaction, rnalID int64) error {
	b := querybuilder.NewUpdate("rnal").Set("brdo = coalesce(brdo, 0) + 1").Where("rnalid = ?", rnalID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "dokumenti robnog naloga", b)
}

// KomOtpRdokDodaj adds the robna faktura of an otpremnica of the period on a robni nalog (the legacy SnimiRDOK "Novi slog
// u RDOK-u"): the vrsta, the datum, the kupac, the uslovi, the ugovoreni rabat and the kasa, the valuta with its kurs, the
// transport data, the org. jedinica, the mesto troška, the cenovnik, the poreska kategorija and the POPDV polje of the
// otpremnica; the nalog, its vrsta naloga za knjiženje, datum, datum obrade and magacin; the broj, the broj otpremnice
// (dokiz) with its datum (datiz), the opis and the nastanak PDV obaveze of the potvrda; the ids of the mesto isporuke,
// the valuta and the bank of the period of the šifre of the otpremnica, the šifra of its komercijalista; edokid the
// otpremnica; the iznos and the counters 0, not posted (knjige N), with the audit columns. It returns its rdokid.
func (r *KomercijalnoRepository) KomOtpRdokDodaj(ctx context.Context, tx db.Transaction, otpID, rnalID int64, d domain.KomOtpRdokUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into rdok (god, kar, nalog, tipdok, rnalid, vrd, danal, dadok, mag, magaciniid, iznos, rbr, rok, datiz, datob,
			fkto, fana, pornapomena, opiz, foot, mi, fispid, otp, pla, kamt, brst, vozac, brvozila, vrdokid, edokid, ugrabat, vrugrabat, pkase,
			vkase, sifval, idvalute, kurs, vaziza, posuslnab, idorgjed, mestotrid, bttowght, netwght, totboxs, brpaleta, sifbank, bnkizvid,
			rcenovnikhdrid, opis, knjige_1, knjige_2, dokiz, dokum, polje, keykat, category, brugov, descriptioncode, komid, kom, bi, avansid,
			epp_polje, xdatunosa, xopunos)
		select e.god, e.kar, n.nalog, n.tipdok, n.rnalid, e.vrd, n.danal, e.dadok, n.mag, n.magaciniid, 0, 0, e.rok, e.datotp, n.datob,
			e.fkto, e.fana, e.pornapomena, '', e.foot, e.mi,
			(select f.fispid from fisp f where f.konto = e.fkto and f.sifra = e.fana and f.mi = e.mi and f.god = e.god and f.kar = e.kar
				order by f.fispid limit 1),
			e.otp, e.pla, e.kamt, 0, e.vozac, e.brvozila, 0, e.rdokid, e.ugrabat, e.vrugrabat, e.pkase, e.vkase, e.sifval,
			(select v.idvalute from valute v where v.sifval = e.sifval order by v.idvalute limit 1), e.kurs, e.vaziza, e.posuslnab,
			e.idorgjed, e.mestotrid, e.bttowght, e.netwght, e.totboxs, e.brpaleta, e.sifbank,
			(select b.bnkizvid from bnkizv b where b.sifbank = e.sifbank and b.god = e.god and b.kar = e.kar order by b.bnkizvid limit 1),
			e.rcenovnikhdrid, ?, 'N', 'N', ?, ?, e.polje, e.keykat, e.category, e.brugov, ?, e.komid,
			(select k.sifkom from komercijalisti k where k.komid = e.komid), 0, 0, '', now(), ?
		from edok e`, d.Opis, d.Dokiz, d.Dokum, d.DescriptionCode, s.UserName).
		AddJoin("join rnal n on n.rnalid = ? and n.god = e.god and n.kar = e.kar", rnalID).
		Where("e.rdokid = ?", otpID).
		Returning("rdokid")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return 0, err
	}
	var rdokID int64
	if err := r.row(ctx, tx, "faktura otpremnice", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// KomOtpRproDodaj adds a stavka of the robna faktura of an otpremnica of the period (the legacy SnimiRDOK "Novi slog u
// RPRO"): the nalog, the vrsta, the broj, the datum and the kupac of the faktura, the artikal with its lot and konto, the
// količine, the cene, the iznos, the rabat, the tarifa and the poreska kategorija of the stavka, with the audit columns.
// It returns its rproid.
func (r *KomercijalnoRepository) KomOtpRproDodaj(ctx context.Context, tx db.Transaction, p domain.KomOtpRproUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into rpro (god, kar, rnalid, rdokid, nalog, vrd, dokum, dadok, sifra, konto, jm, ncena, kolic, kolic1, iznos,
			fkto, fana, rab, po, mag, magaciniid, brst, rbr, serija, otk, roktr, rsifid, cena, pcena, fcena, cenaval, mcenap, taxcat, taxexreco,
			sifartkup, xdatunosa, xopunos)
		select d.god, d.kar, ?, d.rdokid, d.nalog, d.vrd, d.dokum, d.dadok, ?, ?, ?, ?, ?, ?, ?, d.fkto, d.fana, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, '', now(), ?
		from rdok d`, p.RnalID, p.Sifra, p.Konto, p.JM, p.Ncena, p.Kolic, p.Kolic1, p.Iznos, p.Rab, p.Po, p.Mag, p.MagaciniID, p.Brst, p.Rbr,
		p.Serija, p.Otk, p.Roktr, p.RsifID, p.Cena, p.Pcena, p.Fcena, p.CenaVal, p.Mcenap, p.Taxcat, p.Taxexreco, s.UserName).
		Where("d.rdokid = ?", p.RdokID).
		Returning("rproid")
	if err := inPeriod(ctx, b, "d."); err != nil {
		return 0, err
	}
	var rproID int64
	if err := r.row(ctx, tx, "stavka fakture otpremnice", b, &rproID); err != nil {
		return 0, err
	}
	return rproID, nil
}

// KomOtpRceneDodaj adds the cena of an artikal of the period that has none (the legacy SnimiRDOK RCENE): the nabavna
// cena, the rabat, the cena (also as the planska cena), the prodajna cena, the other cene and the marže 0.
func (r *KomercijalnoRepository) KomOtpRceneDodaj(ctx context.Context, tx db.Transaction, c domain.KomOtpRceneUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into rcene (god, kar, rsifid, sifra, mma, vra, mra, ztro, mcena, mcenap, ncena, ztrin, rab, tz, tzi, cena,
			pcena, fcena, xdatunosa, xopunos)
		select ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, ?, 0, ?, '', '', ?, ?, ?, now(), ?
		from (select 1) x`, s.God, s.Kar, c.RsifID, c.Sifra, c.Ncena, c.Rab, c.Cena, c.Cena, c.Fcena, s.UserName).
		Where("not exists (select 1 from rcene y where y.sifra = ? and y.god = ? and y.kar = ?)", c.Sifra, s.God, s.Kar)
	_, err = r.exec(ctx, tx, "cena artikla otpremnice", b)
	return err
}

// KomOtpRdokBrst saves the broj stavki (and the redni broj) of the robna faktura of an otpremnica of the period;
// ErrNotFound without it.
func (r *KomercijalnoRepository) KomOtpRdokBrst(ctx context.Context, tx db.Transaction, rdokID int64, brst int) error {
	b := querybuilder.NewUpdate("rdok").Set("brst = ?", brst).Set("rbr = ?", brst).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "broj stavki fakture otpremnice", b)
}

// KomOtpAvansi adds to the robna faktura of an otpremnica the avansi of its nalog za isporuku of the period (avansfakt of
// the edokid; the legacy SnimiRDOK "povezivanje avansa"): the osnovice and the PDV of the avans, the parts of the faktura
// 0 (the obračun of the faktura computes them). It returns how many.
func (r *KomercijalnoRepository) KomOtpAvansi(ctx context.Context, tx db.Transaction, nalogID, rdokID int64) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into avansfakt (god, kar, avansid, rdokid, osnfakt1, pdv1fakt, osnfakt2, pdv2fakt, osnfakt3, osnavans1,
			pdv1avans, osnavans2, pdv2avans, osnavans3, edokid, xdatunosa, xopunos)
		select a.god, a.kar, a.avansid, ?, 0, 0, 0, 0, 0, a.osnavans1, a.pdv1avans, a.osnavans2, a.pdv2avans, a.osnavans3, 0, now(), ?
		from avansfakt a`, rdokID, s.UserName).
		Where("a.edokid = ?", nalogID)
	if err := inPeriod(ctx, b, "a."); err != nil {
		return 0, err
	}
	return r.exec(ctx, tx, "avansi fakture otpremnice", b)
}

// KomOtpPotvrdi marks a new otpremnica of the period potvrđena (doksts 40) with its final broj otpremnice and fakture and
// its opis (the legacy SnimiRDOK "promeni mu status"), with the audit columns. It returns false when it is not new.
func (r *KomercijalnoRepository) KomOtpPotvrdi(ctx context.Context, tx db.Transaction, otpID int64, dokum, brotp int, opis string) (bool, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	b := querybuilder.NewUpdate("edok").
		Set("doksts = ?", domain.KomOtpPotvrdjena).
		Set("dokum = ?", dokum).
		Set("brotp = ?", brotp).
		Set("opis = ?", opis).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", otpID).
		Where("coalesce(doksts, '') = ?", domain.KomOtpNova)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	n, err := r.exec(ctx, tx, "potvrda otpremnice", b)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomOtpNalogIsporuceno writes the isporučene količine of a potvrđena otpremnica of the period to the stavke of its nalog
// za isporuku (epro.kolic of the stavka of rproid1; the "Pregled delimično realizovanih naloga" of the nalozi).
func (r *KomercijalnoRepository) KomOtpNalogIsporuceno(ctx context.Context, tx db.Transaction, otpID int64) error {
	b := querybuilder.NewUpdate("epro").
		Set("kolic = abs(coalesce(o.kolic, 0))").
		From("epro o").
		Where("o.rdokid = ?", otpID).
		Where("o.god = epro.god").
		Where("o.kar = epro.kar").
		Where("epro.rproid = o.rproid1")
	if err := inPeriod(ctx, b, "epro."); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "isporučene količine naloga", b)
	return err
}
