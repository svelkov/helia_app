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

// The nalozi za isporuku of the komercijalno (the legacy pr_PorudzbenicaProdaja, the porudžbenice od kupaca): edok of
// the vrste of the modul PRO and the group NRK, read and written like the profakture (komercijalno_profaktura_
// repository.go), with what only a nalog has: its profaktura (vrdokid), the bank of the payment of an izvoz (sifbank),
// the rezervacija of the zalihe by the age of the lots, the finansijsko stanje of the kupac and the lists of the main
// screen.

// KomNalogDodatno returns what a nalog of the period has besides the fields of the profaktura: its profaktura (vrdokid,
// with its broj and datum), the bank of the payment of an izvoz (sifbank with the bnkizv of the period of the šifra and
// its opis), the broj of its otpremnica and of its faktura (see komNalogOtpremnica). Zero values without the nalog.
func (r *KomercijalnoRepository) KomNalogDodatno(ctx context.Context, rdokID int64) (domain.KomNalogDodatno, error) {
	b := querybuilder.New(`select coalesce(e.vrdokid, 0)::bigint, coalesce(p.dokum, 0)::int, p.dadok, coalesce(e.sifbank, 0)::int,
			coalesce(k.bnkizvid, 0)::bigint, coalesce(k.bnkdes, ''), `+komNalogOtpremnica+`, `+komNalogFaktura+`
		from edok e`).
		AddJoin("left join edok p on p.rdokid = e.vrdokid and p.god = e.god and p.kar = e.kar").
		AddJoin(`left join lateral (select b.bnkizvid, b.bnkdes from bnkizv b where b.sifbank = e.sifbank and b.god = e.god and b.kar = e.kar
			order by b.bnkizvid limit 1) k on true`).
		Where("e.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomNalogDodatno{}, err
	}
	var d domain.KomNalogDodatno
	err := r.row(ctx, nil, "nalog za isporuku", b, &d.VrdokID, &d.ProDokum, &d.ProDadok, &d.SifBank, &d.BnkizvID, &d.BnkDes, &d.Otpremnica,
		&d.Faktura)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomNalogDodatno{}, nil
	}
	return d, err
}

// komNalogOtpremnica is the broj of the otpremnica of the nalog e (the edok of the period whose vrdokid is the nalog;
// pr_Otpremnica) and komNalogFaktura the broj of its faktura (the rdok of the period formed from the potvrđena
// otpremnica, rdok.edokid).
const (
	komNalogOtpremnica = `coalesce((select o.brotp from edok o where o.vrdokid = e.rdokid and o.god = e.god and o.kar = e.kar
			and o.doksts = '40' order by o.rdokid limit 1), 0)::int`
	komNalogFaktura = `coalesce((select f.dokum from edok o join rdok f on f.edokid = o.rdokid and f.god = o.god and f.kar = o.kar
			where o.vrdokid = e.rdokid and o.god = e.god and o.kar = e.kar order by f.rdokid limit 1), 0)::int`
)

// KomNalogDodatnoSacuvaj saves the profaktura (vrdokid) and the bank of the payment of an izvoz (sifbank) of a nalog of
// the period (the legacy SnimiEDOK); ErrNotFound without it.
func (r *KomercijalnoRepository) KomNalogDodatnoSacuvaj(ctx context.Context, tx db.Transaction, rdokID, vrdokID int64, sifbank int) error {
	b := querybuilder.NewUpdate("edok").
		Set("vrdokid = ?", vrdokID).
		Set("sifbank = ?", sifbank).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "profaktura i banka naloga", b)
}

// KomNalogBanka returns the šifra of a bank of the payment of an izvoz of the period (bnkizv; the legacy EDT_SIFBANK);
// found is false without it.
func (r *KomercijalnoRepository) KomNalogBanka(ctx context.Context, bnkizvID int64) (int, bool, error) {
	b := querybuilder.New(`select coalesce(sifbank, 0)::int from bnkizv`).Where("bnkizvid = ?", bnkizvID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return 0, false, err
	}
	var sifbank int
	err := r.row(ctx, nil, "banka za uplatu", b, &sifbank)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return sifbank, true, nil
}

// KomNalogKopirajStavke copies the stavke of a profaktura of the period into a nalog (the legacy
// KopirajStavkeizProfakture): the artikal with its lot and konto, the količina, the cene and the rabat, the vrsta, the
// broj and the datum of the nalog, the redni brojevi from rbrOd on, with the audit columns.
func (r *KomercijalnoRepository) KomNalogKopirajStavke(ctx context.Context, tx db.Transaction, proID, nalogID int64, vrd, dokum int,
	dadok sql.NullTime, rbrOd int) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, sifra, po, rsifid, rbr, mag, magaciniid, trazkolic, ncena, cena, fcena, cenaval,
			rab, otk, serija, roktr, konto, vrd, dokum, dadok, xdatunosa, xopunos)
		select p.god, p.kar, ?, p.sifra, p.po, p.rsifid, coalesce(p.rbr, 0) + ? - 1, p.mag, p.magaciniid, p.trazkolic, p.ncena, p.cena, p.fcena,
			p.cenaval, p.rab, p.otk, p.serija, p.roktr, p.konto, ?, ?, ?, now(), ?
		from epro p`, nalogID, rbrOd, vrd, dokum, dadok, s.UserName).
		Where("p.rdokid = ?", proID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return err
	}
	_, err = r.exec(ctx, tx, "stavke profakture naloga", b)
	return err
}

// KomNalogProfakturaVezi marks the potvrđena profaktura of the period a nalog is made of (doksts 40; the legacy
// KopirajStavkeizProfakture). It returns false when the profaktura is not a potvrđena one (any more).
func (r *KomercijalnoRepository) KomNalogProfakturaVezi(ctx context.Context, tx db.Transaction, proID int64) (bool, error) {
	return r.komProStatus(ctx, tx, proID, domain.KomProfakturaNalog, domain.KomProfakturaPotvrdjena)
}

// KomNalogProfakturaVrati sets the profaktura of the period of a deleted nalog back to potvrđena (doksts 20; the legacy
// BTN_Brisanje sets 10). It returns false when the profaktura is not marked.
func (r *KomercijalnoRepository) KomNalogProfakturaVrati(ctx context.Context, tx db.Transaction, proID int64) (bool, error) {
	return r.komProStatus(ctx, tx, proID, domain.KomProfakturaPotvrdjena, domain.KomProfakturaNalog)
}

// KomNalogLotoviStarost returns the lots (drsta) of an artikal in a konto of a magacin of the period with a slobodna
// količina (ulaz - izlaz - rezervisano greater than 0), the oldest rok trajanja first (the legacy AddEPRO with "Skidaj
// zalihe po starosti": order by roktr, serija). Not locked: the reservation locks each lot (DrstaIDZaIzmenu).
func (r *KomercijalnoRepository) KomNalogLotoviStarost(ctx context.Context, magaciniID int, konto string, sifra int) ([]domain.KomNalogLot, error) {
	b := querybuilder.New(`select trim(coalesce(d.konto, '')) as konto, coalesce(d.otk, '') as otk, coalesce(d.serija, '') as serija,
			coalesce(d.roktr, 0)::bigint as roktr, (coalesce(d.ulaz, 0) - coalesce(d.izlaz, 0) - coalesce(d.rez, 0))::float8 as slobodno
		from drsta d`).
		Where("d.magaciniid = ?", magaciniID).
		Where("d.konto = ?", strings.TrimSpace(konto)).
		Where("d.sifra = ?", sifra).
		Where("coalesce(d.ulaz, 0) - coalesce(d.izlaz, 0) - coalesce(d.rez, 0) > 0")
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	b.OrderByExpr("coalesce(d.roktr, 0), d.serija, d.rstaid")
	var rows []domain.KomNalogLot
	if err := r.selectRows(ctx, "lotovi po starosti", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomNalogFpro returns the promet of a kupac of the period in the finansijsko knjigovodstvo (fpro of the konto and the
// šifra of the vrste 10, 30 and 80; the legacy SaldoKUPCA): the kategorija, the iznos, the datum and the rok.
func (r *KomercijalnoRepository) KomNalogFpro(ctx context.Context, konto, sifra string) ([]domain.KomNalogFpro, error) {
	b := querybuilder.New(`select coalesce(f.kat, 0)::int as kat, coalesce(f.iznos, 0)::float8 as iznos, f.dadok as dadok,
			coalesce(f.rok, 0)::int as rok
		from fpro f`).
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra)).
		Where("f.vrd in (10, 30, 80)")
	if err := inPeriod(ctx, b, "f."); err != nil {
		return nil, err
	}
	b.OrderBy("f.idfpro")
	var rows []domain.KomNalogFpro
	if err := r.selectRows(ctx, "promet kupca", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomNalogRobneStavke returns the stavke of the robne fakture of a kupac of the period not posted yet (rdok of the
// konto and the šifra of the vrste FAK, FPR and KPK with knjige 1 other than "D", with their rpro; the legacy
// SaldoKUPCA): the document with its datum, rok, ugovoreni rabat and kasa, the stavka with its količina, cena, rabat and
// the stopa PDV of its tarifa on its datum.
func (r *KomercijalnoRepository) KomNalogRobneStavke(ctx context.Context, konto, sifra string) ([]domain.KomNalogRobnaStavka, error) {
	b := querybuilder.New(`select d.rdokid::bigint as rdokid, d.dadok as dadok, coalesce(d.rok, 0)::int as rok,
			coalesce(d.ugrabat, 0)::float8 as ugrabat, coalesce(d.pkase, 0)::float8 as pkase, coalesce(p.kolic, 0)::float8 as kolic,
			coalesce(p.fcena, 0)::float8 as fcena, coalesce(p.rab, 0)::float8 as rab, `+poreskaStopa("p.po", "p.dadok")+` as stopa
		from rdok d`).
		AddJoin("join rpro p on p.rdokid = d.rdokid and p.god = d.god and p.kar = d.kar").
		Where("d.fkto = ?", strings.TrimSpace(konto)).
		Where("d.fana = ?", strings.TrimSpace(sifra)).
		Where(`d.vrd in (select v.vrd from dokvrsta v where v.god = d.god and v.kar = d.kar
			and upper(trim(coalesce(v.dokozn, ''))) in ('FAK', 'FPR', 'KPK'))`).
		Where("coalesce(d.knjige_1, '') <> 'D'")
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	b.OrderBy("d.rdokid").OrderBy("p.rproid")
	var rows []domain.KomNalogRobnaStavka
	if err := r.selectRows(ctx, "neproknjižene fakture kupca", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomNalogPrikaz returns the nalozi of the lists of the main screen in the magacini of the user of the session
// (maguser), of one magacin or of all of them (SviMagacini), with the "mag-opis" of the magacin, the broj of the
// otpremnica and of the faktura, the naziv of the kupac:
//   - "Pregled naloga za isporuku" (the legacy PR_QRY_EDOK): the nepotvrđeni (10), else the potvrđeni (20 and 40), the
//     last ones first;
//   - "Pregled delimično realizovanih naloga" (Delimicni, the legacy PuniTABLEHDR): the ones with the otpremnica (40)
//     with a stavka whose isporučena količina differs from the tražena one, in the order of their broj; only when the
//     otpremnica is potvrđena (it writes the isporučena količina).
func (r *KomercijalnoRepository) KomNalogPrikaz(ctx context.Context, p domain.KomNalogPrikazParams) ([]domain.KomNalogPrikazRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select e.rdokid::bigint as rdokid,
			coalesce(m.mag::text || '-' || coalesce(m.opis, ''), '') as magacin, coalesce(e.vrd, 0)::int as vrd,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok, `+komNalogOtpremnica+` as otpremnica,
			`+komNalogFaktura+` as faktura,
			coalesce(e.fkto, '') as fkto, coalesce(e.fana, '') as fana, coalesce(f.naziv, '') as kupac, coalesce(e.rok, 0)::int as rok,
			coalesce(e.ugrabat, 0)::float8 as ugrabat, coalesce(e.doksts, '') as doksts
		from edok e`).
		AddJoin("join (select distinct mu.mag from maguser mu where upper(mu.iduser) = ?) u on u.mag = e.mag",
			strings.ToUpper(strings.TrimSpace(s.UserName))).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana").
		Where(komProVrstaUslov, komZahtevModul, komProNalogGrupa)
	switch {
	case p.Delimicni:
		b.Where("coalesce(e.doksts, '') = ?", domain.KomNalogOtpremljen).
			Where(`exists (select 1 from epro x where x.rdokid = e.rdokid and x.god = e.god and x.kar = e.kar
				and coalesce(x.trazkolic, 0) <> coalesce(x.kolic, 0))`).
			Where(`exists (select 1 from edok o where o.vrdokid = e.rdokid and o.god = e.god and o.kar = e.kar and o.doksts = ?)`,
				domain.KomNalogOtpremljen)
	case p.Doksts == domain.KomNalogNov:
		b.Where("coalesce(e.doksts, '') = ?", domain.KomNalogNov)
	default:
		b.Where("coalesce(e.doksts, '') in (?, ?)", domain.KomNalogPotvrdjen, domain.KomNalogOtpremljen)
	}
	if !p.SviMagacini {
		b.Where("e.magaciniid = ?", p.MagaciniID)
	}
	komPretraga(b, p.Tekst, "e.dokum::text", "e.foot", "f.naziv", "e.fana")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return nil, err
	}
	if p.Delimicni {
		b.OrderByExpr("e.dokum, e.rdokid")
	} else {
		b.OrderByExpr("e.dadok desc nulls last, e.vrd desc, e.dokum desc")
	}
	var rows []domain.KomNalogPrikazRow
	if err := r.selectRows(ctx, "pregled naloga za isporuku", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}
