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

// The e-otpremnice of the komercijalno (the legacy pr_eOTPREMNICA with IW_Otpremnice): the otpremnice of
// komercijalno_otpremnica_repository.go with the data of the e-otpremnica (the columns of edok of migrations/003), their
// faze vožnje (enot) and their opasne materije (eopma).

// KomEotp returns the data of the e-otpremnica of an otpremnica of the period; found is false without it.
func (r *KomercijalnoRepository) KomEotp(ctx context.Context, rdokID int64) (domain.KomEotpPodaci, bool, error) {
	b := querybuilder.New(`select coalesce(e.opasnemat, false), coalesce(e.roba3str, false), coalesce(e.tipeotp, ''), coalesce(e.jm1, ''),
			coalesce(e.jm2, ''), coalesce(e.bttovol, 0)::float8, e.datpoc, e.datzav, e.stvdatotp, coalesce(e.povratrobe, false),
			coalesce(e.brnarudz, ''), coalesce(e.brokvspor, ''), coalesce(e.kontime, ''), coalesce(e.konttel, ''), coalesce(e.kontmail, ''),
			coalesce(e.referkup, ''), coalesce(e.telkup, ''), coalesce(e.mailkup, ''), coalesce(e.nacotp, 0)::int, coalesce(e.licime, ''),
			coalesce(e.licprez, ''), coalesce(e.licbrlk, ''), coalesce(e.mestoutov, ''), coalesce(e.mestoisp, ''), coalesce(e.prevkto, ''),
			coalesce(e.prevana, ''), coalesce(e.vozprez, ''), coalesce(e.vozmail, ''), coalesce(e.voztel, ''), coalesce(e.vozdozv, ''),
			coalesce(e.ulicaotp, ''), coalesce(e.ulbrotp, ''), coalesce(e.mestootp, ''), coalesce(e.posbrotp, ''), coalesce(e.sifobjotp, ''),
			coalesce(e.ulicaodr, ''), coalesce(e.ulbrodr, ''), coalesce(e.mestoodr, ''), coalesce(e.posbrodr, ''), coalesce(e.sifobjodr, ''),
			coalesce(e.fisnap, ''), coalesce(e.roba3strtxt, '')
		from edok e`).
		Where("e.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomEotpPodaci{}, false, err
	}
	var p domain.KomEotpPodaci
	err := r.row(ctx, nil, "e-otpremnica", b, &p.OpasneMat, &p.Roba3Str, &p.TipEotp, &p.Jm1, &p.Jm2, &p.Bttovol, &p.Datpoc, &p.Datzav,
		&p.StvDatotp, &p.PovratRobe, &p.BrNarudz, &p.BrOkvSpor, &p.KontIme, &p.KontTel, &p.KontMail, &p.ReferKup, &p.TelKup, &p.MailKup,
		&p.NacOtp, &p.LicIme, &p.LicPrez, &p.LicBrLk, &p.MestoUtov, &p.MestoIsp, &p.PrevKto, &p.PrevAna, &p.VozPrez, &p.VozMail, &p.VozTel,
		&p.VozDozv, &p.UlicaOtp, &p.UlBrOtp, &p.MestoOtp, &p.PosBrOtp, &p.SifObjOtp, &p.UlicaOdr, &p.UlBrOdr, &p.MestoOdr, &p.PosBrOdr,
		&p.SifObjOdr, &p.FisNap, &p.Roba3StrTxt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomEotpPodaci{}, false, nil
	}
	if err != nil {
		return domain.KomEotpPodaci{}, false, err
	}
	return p, true, nil
}

// KomEotpIzmeni saves the data of the e-otpremnica of an otpremnica of the period with its mesto isporuke (mi), with
// the audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomEotpIzmeni(ctx context.Context, tx db.Transaction, rdokID int64, u domain.KomEotpUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("edok").
		Set("opasnemat = ?", u.OpasneMat).
		Set("roba3str = ?", u.Roba3Str).
		Set("tipeotp = ?", u.TipEotp).
		Set("jm1 = ?", u.Jm1).
		Set("jm2 = ?", u.Jm2).
		Set("bttovol = ?", u.Bttovol).
		Set("datpoc = ?", u.Datpoc).
		Set("datzav = ?", u.Datzav).
		Set("stvdatotp = ?", u.StvDatotp).
		Set("povratrobe = ?", u.PovratRobe).
		Set("brnarudz = ?", u.BrNarudz).
		Set("brokvspor = ?", u.BrOkvSpor).
		Set("kontime = ?", u.KontIme).
		Set("konttel = ?", u.KontTel).
		Set("kontmail = ?", u.KontMail).
		Set("referkup = ?", u.ReferKup).
		Set("telkup = ?", u.TelKup).
		Set("mailkup = ?", u.MailKup).
		Set("nacotp = ?", u.NacOtp).
		Set("licime = ?", u.LicIme).
		Set("licprez = ?", u.LicPrez).
		Set("licbrlk = ?", u.LicBrLk).
		Set("mestoutov = ?", u.MestoUtov).
		Set("mestoisp = ?", u.MestoIsp).
		Set("prevkto = ?", u.PrevKto).
		Set("prevana = ?", u.PrevAna).
		Set("vozprez = ?", u.VozPrez).
		Set("vozmail = ?", u.VozMail).
		Set("voztel = ?", u.VozTel).
		Set("vozdozv = ?", u.VozDozv).
		Set("ulicaotp = ?", u.UlicaOtp).
		Set("ulbrotp = ?", u.UlBrOtp).
		Set("mestootp = ?", u.MestoOtp).
		Set("posbrotp = ?", u.PosBrOtp).
		Set("sifobjotp = ?", u.SifObjOtp).
		Set("mi = ?", u.Mi).
		Set("ulicaodr = ?", u.UlicaOdr).
		Set("ulbrodr = ?", u.UlBrOdr).
		Set("mestoodr = ?", u.MestoOdr).
		Set("posbrodr = ?", u.PosBrOdr).
		Set("sifobjodr = ?", u.SifObjOdr).
		Set("fisnap = ?", u.FisNap).
		Set("roba3strtxt = ?", u.Roba3StrTxt).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "e-otpremnica", b)
}

// KomEotpFaze returns the faze vožnje of an otpremnica of the period in the order of their redni broj.
func (r *KomercijalnoRepository) KomEotpFaze(ctx context.Context, rdokID int64) ([]domain.KomEotpFaza, error) {
	b := querybuilder.New(`select coalesce(n.rbr, 0)::int as rbr, coalesce(n.nacotp, 0)::int as nacotp, coalesce(n.naziv, '') as naziv
		from enot n`).
		Where("n.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "n."); err != nil {
		return nil, err
	}
	b.OrderBy("n.rbr").OrderBy("n.enotid")
	var faze []domain.KomEotpFaza
	if err := r.selectRows(ctx, "faze vožnje", &faze, b); err != nil {
		return nil, err
	}
	return faze, nil
}

// KomEotpOpasne returns the opasne materije of an otpremnica of the period in the order of their redni broj.
func (r *KomercijalnoRepository) KomEotpOpasne(ctx context.Context, rdokID int64) ([]domain.KomEotpOpasna, error) {
	b := querybuilder.New(`select coalesce(o.rbr, 0)::int as rbr, coalesce(o.naziv, '') as naziv, coalesce(o.opis, '') as opis,
			coalesce(o.komentar, '') as komentar
		from eopma o`).
		Where("o.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "o."); err != nil {
		return nil, err
	}
	b.OrderBy("o.rbr").OrderBy("o.eopmaid")
	var opasne []domain.KomEotpOpasna
	if err := r.selectRows(ctx, "opasne materije", &opasne, b); err != nil {
		return nil, err
	}
	return opasne, nil
}

// KomEotpBrisi deletes the faze vožnje and the opasne materije of an otpremnica of the period (with the otpremnica or
// before they are written again).
func (r *KomercijalnoRepository) KomEotpBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	faze := querybuilder.New(`delete from enot`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, faze, ""); err != nil {
		return err
	}
	if _, err := r.exec(ctx, tx, "faze vožnje", faze); err != nil {
		return err
	}
	opasne := querybuilder.New(`delete from eopma`).Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, opasne, ""); err != nil {
		return err
	}
	_, err := r.exec(ctx, tx, "opasne materije", opasne)
	return err
}

// KomEotpFazaDodaj adds a faza vožnje to an otpremnica in the period of the session, with the audit columns.
func (r *KomercijalnoRepository) KomEotpFazaDodaj(ctx context.Context, tx db.Transaction, rdokID int64, f domain.KomEotpFaza) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into enot (god, kar, rdokid, rbr, nacotp, naziv, xdatunosa, xopunos) values (?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, rdokID, f.Rbr, f.NacOtp, f.Naziv, s.UserName)
	_, err = r.exec(ctx, tx, "faza vožnje", b)
	return err
}

// KomEotpOpasnaDodaj adds an opasna materija to an otpremnica in the period of the session, with the audit columns.
func (r *KomercijalnoRepository) KomEotpOpasnaDodaj(ctx context.Context, tx db.Transaction, rdokID int64, o domain.KomEotpOpasna) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into eopma (god, kar, rdokid, rbr, naziv, opis, komentar, xdatunosa, xopunos) values (?, ?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, rdokID, o.Rbr, o.Naziv, o.Opis, o.Komentar, s.UserName)
	_, err = r.exec(ctx, tx, "opasna materija", b)
	return err
}

// KomEotpPartner returns a partner of the konto and šifra of the period as the e-otpremnica shows it (the analitika of
// fkpl with its partner: the legacy PrikaziKupca and PrikaziPrevoznika); found is false without it.
func (r *KomercijalnoRepository) KomEotpPartner(ctx context.Context, konto, sifra string) (domain.KomEotpPartner, bool, error) {
	b := querybuilder.New(`select coalesce(p.naziv, f.naziv, ''), coalesce(p.adresa, ''),
			trim(coalesce(nullif(p.pobro, 0)::text, '') || ' ' || coalesce(p.mesto, '')), coalesce(p.pib, ''), coalesce(p.matbr, ''),
			coalesce(p.jbkjs, '')
		from fkpl f`).
		AddJoin("left join partneri p on p.idpartneri = f.idpartneri").
		Where("f.vkonta = 1").
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra))
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.KomEotpPartner{}, false, err
	}
	b.OrderBy("f.idfkpl").Limit(1)
	var p domain.KomEotpPartner
	err := r.row(ctx, nil, "partner e-otpremnice", b, &p.Naziv, &p.Adresa, &p.Mesto, &p.Pib, &p.Matbr, &p.Jbkjs)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomEotpPartner{}, false, nil
	}
	if err != nil {
		return domain.KomEotpPartner{}, false, err
	}
	return p, true, nil
}

// KomEotpPosiljalac returns the firm of the session as the pošiljalac of the e-otpremnica (fvr of the period with the
// naziv of the firm; the legacy EDT_FVR*).
func (r *KomercijalnoRepository) KomEotpPosiljalac(ctx context.Context, firma string) (domain.KomEotpPartner, error) {
	b := querybuilder.New(`select coalesce(f.naziv, ''), coalesce(f.adresa, ''),
			trim(coalesce(nullif(f.pobro, 0)::text, '') || ' ' || coalesce(f.mesto, '')), coalesce(f.pib, ''), coalesce(f.matbr, ''),
			coalesce(f.jbkjs, '')
		from fvr f`).
		Where("f.naziv = ?", firma)
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.KomEotpPartner{}, err
	}
	b.Limit(1)
	var p domain.KomEotpPartner
	err := r.row(ctx, nil, "pošiljalac", b, &p.Naziv, &p.Adresa, &p.Mesto, &p.Pib, &p.Matbr, &p.Jbkjs)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomEotpPartner{}, nil
	}
	return p, err
}

// KomEotpMagacinAdresa returns the adresa of a magacin of the period (the adresa otpreme of a new e-otpremnica);
// found is false without it.
func (r *KomercijalnoRepository) KomEotpMagacinAdresa(ctx context.Context, magaciniID int) (domain.KomEotpAdresa, bool, error) {
	b := querybuilder.New(`select coalesce(adresa, ''), coalesce(mesto, ''), coalesce(nullif(pobro, 0)::text, '') from magacini`).
		Where("magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return domain.KomEotpAdresa{}, false, err
	}
	var a domain.KomEotpAdresa
	err := r.row(ctx, nil, "adresa magacina", b, &a.Ulica, &a.Mesto, &a.Pobro)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomEotpAdresa{}, false, nil
	}
	if err != nil {
		return domain.KomEotpAdresa{}, false, err
	}
	return a, true, nil
}

// KomEotpMestoIsporuke returns the adresa of a mesto isporuke of a partner of the period (fisp by its konto, šifra and
// MI: the legacy PrikaziMI); found is false without it.
func (r *KomercijalnoRepository) KomEotpMestoIsporuke(ctx context.Context, konto, sifra string, mi int) (domain.KomEotpAdresa, bool, error) {
	b := querybuilder.New(`select coalesce(f.adresa, ''), coalesce(f.mesto, ''), coalesce(nullif(f.pobro, 0)::text, '') from fisp f`).
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra)).
		Where("f.mi = ?", mi)
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.KomEotpAdresa{}, false, err
	}
	b.OrderBy("f.fispid").Limit(1)
	var a domain.KomEotpAdresa
	err := r.row(ctx, nil, "mesto isporuke e-otpremnice", b, &a.Ulica, &a.Mesto, &a.Pobro)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomEotpAdresa{}, false, nil
	}
	if err != nil {
		return domain.KomEotpAdresa{}, false, err
	}
	return a, true, nil
}

// KomEotpNalog returns the kupac (konto and šifra) and the mesto isporuke of a nalog za isporuku of the period (what a
// new e-otpremnica takes of it); found is false without it.
func (r *KomercijalnoRepository) KomEotpNalog(ctx context.Context, nalogID int64) (fkto, fana string, mi int, found bool, err error) {
	b := querybuilder.New(`select coalesce(n.fkto, ''), coalesce(n.fana, ''), coalesce(n.mi, 0)::int from edok n`).
		Where("n.rdokid = ?", nalogID)
	if err = inPeriod(ctx, b, "n."); err != nil {
		return "", "", 0, false, err
	}
	err = r.row(ctx, nil, "nalog e-otpremnice", b, &fkto, &fana, &mi)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", 0, false, nil
	}
	if err != nil {
		return "", "", 0, false, err
	}
	return fkto, fana, mi, true, nil
}
