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

// The profakture of the komercijalno (the legacy pr_Profakture): edok of the vrste of the modul PRO and the group PRF
// for one kupac, with their stavke in epro, and the nalog formed from a profaktura (a porudžbenica kupca, the vrste of
// the group NRK) with the rezervacija of the zalihe of its lots (drsta.rez).

// The groups of the vrste dokumenata of the profakture and of the nalog formed from them (the legacy PR_QRY_DOKVRSTA
// with "PRF" and "NRK").
const (
	komProGrupa      = "PRF"
	komProNalogGrupa = "NRK"
)

// komProVrstaUslov restricts edok e to the vrste of the profakture.
const komProVrstaUslov = `e.vrd in (select v.vrd from dokvrsta v where v.god = e.god and v.kar = e.kar
	and upper(trim(coalesce(v.modul, ''))) = ? and upper(trim(coalesce(v.grpdok, ''))) = ?)`

// KomProVrste returns the vrste dokumenata of the profakture of the period (dokvrsta of the modul PRO and the group PRF)
// as "vrd - opis" keyed by the vrd, in the order of the vrd.
func (r *KomercijalnoRepository) KomProVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return r.komVrste(ctx, komProGrupa)
}

// KomProNalogVrste returns the vrste dokumenata of the nalozi formed from the profakture (the group NRK, the legacy
// nVRDNALOG: the first one).
func (r *KomercijalnoRepository) KomProNalogVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return r.komVrste(ctx, komProNalogGrupa)
}

// KomProTipdok returns the vrste naloga za knjiženje (tipdok) of a profaktura of the vrsta (the legacy QRY_TIPDOK of the
// moduli PRO, ROB and SVI whose grpvrd contains the vrsta; without one of them all the tipdok of these moduli), except
// "00", as "tipdok - opis" keyed by the tipdok, in their order.
func (r *KomercijalnoRepository) KomProTipdok(ctx context.Context, vrd int) ([]domain.ComboItem, error) {
	vrsta := fmt.Sprintf("%d", vrd)
	b := querybuilder.New(`select trim(t.tipdok) as key, trim(t.tipdok) || ' - ' || coalesce(t.opis, '') as value from tipdok t`).
		Where("upper(trim(coalesce(t.grpdok, ''))) in ('PRO', 'ROB', 'SVI')").
		Where("trim(coalesce(t.tipdok, '')) not in ('', '00')").
		Where(`(position(? in coalesce(t.grpvrd, '')) > 0 or not exists (select 1 from tipdok x where x.god = t.god and x.kar = t.kar
			and upper(trim(coalesce(x.grpdok, ''))) in ('PRO', 'ROB', 'SVI') and position(? in coalesce(x.grpvrd, '')) > 0))`, vrsta, vrsta)
	if err := inPeriod(ctx, b, "t."); err != nil {
		return nil, err
	}
	b.OrderByExpr("trim(t.tipdok)")
	var rows []domain.ComboItem
	if err := r.selectRows(ctx, "vrste naloga za knjiženje profakture", &rows, b); err != nil {
		return nil, err
	}
	return rows, nil
}

// KomProfaktura returns a profaktura of the period with the texts of its šifre, its magacin, the oznaka of its valuta
// and the broj of the nalog formed from it; found is false without it.
func (r *KomercijalnoRepository) KomProfaktura(ctx context.Context, rdokID int64) (domain.KomProfakturaRow, bool, error) {
	return r.komProDokument(ctx, rdokID, komProGrupa)
}

// KomNalog returns a nalog za isporuku of the period (a document of the vrste of the group NRK, read like the
// profaktura); found is false without it.
func (r *KomercijalnoRepository) KomNalog(ctx context.Context, rdokID int64) (domain.KomProfakturaRow, bool, error) {
	return r.komProDokument(ctx, rdokID, komProNalogGrupa)
}

// komProDokument returns a document of the period of the vrste of the modul PRO and the group, read like the profaktura.
func (r *KomercijalnoRepository) komProDokument(ctx context.Context, rdokID int64, grupa string) (domain.KomProfakturaRow, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.vrd, 0)::int, trim(coalesce(e.tipdok, '')), coalesce(e.dokum, 0)::int, e.dadok,
			coalesce(e.rok, 0)::int, coalesce(e.magaciniid, 0)::int, coalesce(e.mag, 0)::int, coalesce(e.mag::text || '-' || coalesce(m.opis, ''), ''),
			coalesce(m.adresa, ''), coalesce(m.mesto, ''), upper(trim(coalesce(m.tipmag, ''))), coalesce(m.tipzal, 0)::int, coalesce(m.tipcene, 0)::int,
			coalesce(e.fkto, ''), coalesce(e.fana, ''), e.komid::bigint, coalesce(k.sifkom, 0)::int, coalesce(k.imeprezime, ''),
			e.idorgjed::bigint, coalesce(concat_ws(' - ', nullif(o.ojozn::text, ''), nullif(o.naziv, '')), ''),
			e.mestotrid::bigint, coalesce(concat_ws(' - ', nullif(t.mtroska, ''), nullif(t.opis, '')), ''),
			coalesce(e.foot, ''), coalesce(e.doksts, ''), coalesce(e.knjige_1, ''), coalesce(e.knjige_4, ''), coalesce(e.sifval, 0)::int,
			coalesce((select v.oznval from valute v where v.sifval = e.sifval order by v.idvalute limit 1), ''), coalesce(e.kurs, 0)::float8,
			coalesce(e.vaziza, 0)::int, coalesce(e.otp, ''), coalesce(e.pla, ''), coalesce(e.posuslnab, ''), coalesce(e.mi, 0)::int,
			coalesce(e.ugrabat, 0)::float8, coalesce(e.vrugrabat, 0)::float8, coalesce(e.pkase, 0)::float8, coalesce(e.vkase, 0)::float8,
			coalesce(e.iznos, 0)::float8, coalesce(e.netwght, 0)::float8, coalesce(e.rcenovnikhdrid, 0)::bigint, coalesce(e.paritet, ''),
			coalesce(e.izjizv, 0)::int, coalesce(e.keykat, ''), coalesce(e.category, ''), coalesce(e.pornapomena, ''), coalesce(e.kamt, ''),
			coalesce(e.brugov, ''), coalesce(e.vozac, ''), coalesce(e.brvozila, ''),
			coalesce((select n.dokum from edok n where n.vrdokid = e.rdokid and n.god = e.god and n.kar = e.kar order by n.rdokid limit 1), 0)::int
		from edok e`).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr t on t.mestotrid = e.mestotrid").
		Where("e.rdokid = ?", rdokID).
		Where(komProVrstaUslov, komZahtevModul, grupa)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomProfakturaRow{}, false, err
	}
	var p domain.KomProfakturaRow
	err := r.row(ctx, nil, "profaktura", b, &p.RdokID, &p.Vrd, &p.Tipdok, &p.Dokum, &p.Dadok, &p.Rok, &p.MagaciniID, &p.Mag, &p.MagacinText,
		&p.MagacinAdresa, &p.MagacinMesto, &p.Tipmag, &p.Tipzal, &p.Tipcene, &p.Fkto, &p.Fana, &p.KomID, &p.KomSifra, &p.KomIme, &p.Orgjed,
		&p.OrgjedText, &p.MestoTroska, &p.MestoTroskaText, &p.Foot, &p.Doksts, &p.Knjige1, &p.Knjige4, &p.SifVal, &p.OznVal, &p.Kurs, &p.VaziZa,
		&p.Otp, &p.Pla, &p.Posuslnab, &p.Mi, &p.Ugrabat, &p.Vrugrabat, &p.Pkase, &p.Vkase, &p.Iznos, &p.Netwght, &p.CenovnikID, &p.Paritet,
		&p.Izjizv, &p.Keykat, &p.Category, &p.Pornapomena, &p.Kamt, &p.Brugov, &p.Vozac, &p.Brvozila, &p.NalogDokum)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomProfakturaRow{}, false, nil
	}
	if err != nil {
		return domain.KomProfakturaRow{}, false, err
	}
	return p, true, nil
}

// KomProMagacin returns the tip, the tip zaliha and the tip cene of a magacin of the period; found is false without it.
func (r *KomercijalnoRepository) KomProMagacin(ctx context.Context, magaciniID int) (tipmag string, tipzal, tipcene int, found bool, err error) {
	b := querybuilder.New(`select upper(trim(coalesce(tipmag, ''))), coalesce(tipzal, 0)::int, coalesce(tipcene, 0)::int from magacini`).
		Where("magaciniid = ?", magaciniID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return "", 0, 0, false, err
	}
	err = r.row(ctx, nil, "magacin profakture", b, &tipmag, &tipzal, &tipcene)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, 0, false, nil
	}
	if err != nil {
		return "", 0, 0, false, err
	}
	return tipmag, tipzal, tipcene, true, nil
}

// KomProDodaj adds a profaktura (the legacy SnimiEDOK of a new document): not of a nalog, new (doksts 10), not posted
// (knjige N), its opis ("Profaktura br: <broj>" without one), the kupac, the uslovi, the valuta with its kurs, the poreska
// kategorija, the cenovnik, the JM of the ambalaža empty (jm1 and jm2 are not null without a default), with the audit
// columns. It returns its rdokid.
func (r *KomercijalnoRepository) KomProDodaj(ctx context.Context, tx db.Transaction, p domain.KomProfakturaUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into edok (god, kar, rnalid, vrd, tipdok, dokum, dadok, mag, magaciniid, iznos, rbr, opis, rok, fkto, fana,
			komid, pla, otp, foot, posuslnab, brst, brotp, vrdokid, idorgjed, mestotrid, sifval, kurs, vaziza, mi, ugrabat, pkase, rcenovnikhdrid,
			paritet, izjizv, keykat, category, pornapomena, kamt, brugov, doksts, knjige_1, knjige_2, jm1, jm2, xdatunosa, xopunos)
		values (?, ?, 0, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			'N', 'N', '', '', now(), ?)`,
		s.God, s.Kar, p.Vrd, p.Tipdok, p.Dokum, p.Dadok, p.Mag, p.MagaciniID, komProOpis(p), p.Rok, p.Fkto, p.Fana,
		p.KomID, p.Pla, p.Otp, p.Foot, p.Posuslnab, p.Orgjed, p.MestoTroska, p.SifVal, p.Kurs, p.VaziZa, p.Mi, p.Ugrabat, p.Pkase, p.CenovnikID,
		p.Paritet, p.Izjizv, p.Keykat, p.Category, p.Pornapomena, p.Kamt, p.Brugov, domain.KomProfakturaNova, s.UserName).
		Returning("rdokid")
	var rdokID int64
	if err := r.row(ctx, tx, "nova profaktura", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// komProOpis is the opis of a new document: its own or "Profaktura br: <broj>".
func komProOpis(p domain.KomProfakturaUpis) string {
	if p.Opis != "" {
		return p.Opis
	}
	return fmt.Sprintf("Profaktura br: %d", p.Dokum)
}

// KomProIzmeni saves the header of a profaktura of the period (the legacy SnimiEDOK of a saved document, ScreenToFile):
// its datum, rok, kupac, komercijalista, org. jedinica, mesto troška, tekst, valuta with its kurs, uslovi, mesto
// isporuke, ugovoreni rabat, kasa, cenovnik, paritet, izvozna izjava and poreska kategorija, with the audit columns (its
// vrsta, tipdok, broj and magacin stay); ErrNotFound without it.
func (r *KomercijalnoRepository) KomProIzmeni(ctx context.Context, tx db.Transaction, p domain.KomProfakturaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("edok").
		Set("dadok = ?", p.Dadok).
		Set("rok = ?", p.Rok).
		Set("fkto = ?", p.Fkto).
		Set("fana = ?", p.Fana).
		Set("komid = ?", p.KomID).
		Set("idorgjed = ?", p.Orgjed).
		Set("mestotrid = ?", p.MestoTroska).
		Set("foot = ?", p.Foot).
		Set("sifval = ?", p.SifVal).
		Set("kurs = ?", p.Kurs).
		Set("vaziza = ?", p.VaziZa).
		Set("otp = ?", p.Otp).
		Set("pla = ?", p.Pla).
		Set("posuslnab = ?", p.Posuslnab).
		Set("mi = ?", p.Mi).
		Set("ugrabat = ?", p.Ugrabat).
		Set("pkase = ?", p.Pkase).
		Set("rcenovnikhdrid = ?", p.CenovnikID).
		Set("paritet = ?", p.Paritet).
		Set("izjizv = ?", p.Izjizv).
		Set("keykat = ?", p.Keykat).
		Set("category = ?", p.Category).
		Set("pornapomena = ?", p.Pornapomena).
		Set("kamt = ?", p.Kamt).
		Set("brugov = ?", p.Brugov).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", p.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "profaktura", b)
}

// KomProUkupno saves the totals of a profaktura of the period (the legacy Preracun): the iznos at the planska cena, the
// vrednost of the ugovoreni rabat and of the kasa and the neto težina; ErrNotFound without it.
func (r *KomercijalnoRepository) KomProUkupno(ctx context.Context, tx db.Transaction, rdokID int64, iznos, vrugrabat, vkase, netwght float64) error {
	b := querybuilder.NewUpdate("edok").
		Set("iznos = ?", iznos).
		Set("vrugrabat = ?", vrugrabat).
		Set("vkase = ?", vkase).
		Set("netwght = ?", netwght).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "ukupno profakture", b)
}

// KomProPotvrdi confirms a new profaktura of the period (the legacy BTN_Potvrda: doksts 20). It returns false when the
// profaktura is not new any more.
func (r *KomercijalnoRepository) KomProPotvrdi(ctx context.Context, rdokID int64) (bool, error) {
	return r.komProStatus(ctx, nil, rdokID, domain.KomProfakturaPotvrdjena, domain.KomProfakturaNova)
}

// komProStatus sets the status of a profaktura of the period that has one of the statuses od; false without it.
func (r *KomercijalnoRepository) komProStatus(ctx context.Context, tx db.Transaction, rdokID int64, doksts string, od ...string) (bool, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	b := querybuilder.NewUpdate("edok").
		Set("doksts = ?", doksts).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", rdokID).
		Where("coalesce(doksts, '') in ("+strings.TrimSuffix(strings.Repeat("?, ", len(od)), ", ")+")", querybuilder.Args(od)...)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	n, err := r.exec(ctx, tx, "status profakture", b)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomProKupac returns the kupac of the konto and šifra of the period (the analitika of fkpl with its partner) with its
// ugovoreni rabat and rok plaćanja; found is false without it.
func (r *KomercijalnoRepository) KomProKupac(ctx context.Context, konto, sifra string) (domain.KomProKupac, bool, error) {
	b := querybuilder.New(`select coalesce(p.naziv, f.naziv, ''), coalesce(p.adresa, ''), coalesce(nullif(p.pobro, 0)::text, ''),
			coalesce(p.mesto, ''), coalesce(p.pib, ''), coalesce(p.tippdv, 0)::int, coalesce(p.ugrabat, 0)::float8, coalesce(p.proval, 0)::int,
			coalesce(p.ter, 0)::int
		from fkpl f`).
		AddJoin("left join partneri p on p.idpartneri = f.idpartneri").
		Where("f.vkonta = 1").
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra))
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.KomProKupac{}, false, err
	}
	b.OrderBy("f.idfkpl").Limit(1)
	var k domain.KomProKupac
	err := r.row(ctx, nil, "kupac", b, &k.Naziv, &k.Adresa, &k.Pobro, &k.Mesto, &k.Pib, &k.TipPdv, &k.Ugrabat, &k.Rok, &k.Ter)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomProKupac{}, false, nil
	}
	if err != nil {
		return domain.KomProKupac{}, false, err
	}
	return k, true, nil
}

// KomProMestoIsporuke returns the mesto isporuke of a kupac of the period (fisp by its konto, šifra and MI); found is
// false without it.
func (r *KomercijalnoRepository) KomProMestoIsporuke(ctx context.Context, konto, sifra string, mi int) (domain.KomProMestoIsporuke, bool, error) {
	b := querybuilder.New(`select coalesce(f.naziv, ''), coalesce(f.adresa, ''), coalesce(nullif(f.pobro, 0)::text, ''), coalesce(f.mesto, ''),
			coalesce(f.pib, ''), coalesce(f.jib, ''), coalesce(nullif(f.gln, 0)::text, '')
		from fisp f`).
		Where("f.konto = ?", strings.TrimSpace(konto)).
		Where("f.sifra = ?", strings.TrimSpace(sifra)).
		Where("f.mi = ?", mi)
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.KomProMestoIsporuke{}, false, err
	}
	b.OrderBy("f.fispid").Limit(1)
	var m domain.KomProMestoIsporuke
	err := r.row(ctx, nil, "mesto isporuke", b, &m.Naziv, &m.Adresa, &m.Pobro, &m.Mesto, &m.Pib, &m.Jib, &m.Gln)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomProMestoIsporuke{}, false, nil
	}
	if err != nil {
		return domain.KomProMestoIsporuke{}, false, err
	}
	return m, true, nil
}

// KomProCenovnik returns a cenovnik (rcenovnikhdr, a šifarnik without a period) with the naziv of its klasa cena; found
// is false without it.
func (r *KomercijalnoRepository) KomProCenovnik(ctx context.Context, id int64) (naziv, tip string, broj int, datcen sql.NullTime, found bool, err error) {
	b := querybuilder.New(`select coalesce(k.naziv, ''), upper(trim(coalesce(h.tip, ''))), coalesce(h.broj, 0)::int, h.datcen from rcenovnikhdr h`).
		AddJoin("left join klasacen k on k.klasacenid = h.klasacenid").
		Where("h.rcenovnikhdrid = ?", id)
	err = r.row(ctx, nil, "cenovnik", b, &naziv, &tip, &broj, &datcen)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", 0, sql.NullTime{}, false, nil
	}
	if err != nil {
		return "", "", 0, sql.NullTime{}, false, err
	}
	return naziv, tip, broj, datcen, true, nil
}

// KomProStavke returns the stavke of a profaktura (or of a nalog za isporuku) of the period in the order of their redni
// broj, with the naziv, the JM, the komercijalni opis, the proizvođač, the težina and the zapremina of their artikal, the
// decimals of the JM and the isporučena količina.
func (r *KomercijalnoRepository) KomProStavke(ctx context.Context, rdokID int64) ([]domain.KomProStavkaRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, coalesce(p.rbr, 0)::int as rbr, coalesce(p.konto, '') as konto,
			coalesce(p.sifra, 0)::int as sifra, coalesce(p.rsifid, 0)::bigint as rsifid, coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm, coalesce(s.komercopis, '') as komercopis,
			coalesce(s.pro, '') as pro, coalesce(s.proizsifra, '') as proizsifra, coalesce(p.otk, '') as otk, coalesce(p.serija, '') as serija,
			coalesce(p.roktr, 0)::bigint as roktr, coalesce(p.trazkolic, 0)::float8 as kolicina, coalesce(p.cena, 0)::float8 as cena,
			coalesce(p.fcena, 0)::float8 as fcena, coalesce(p.cenaval, 0)::float8 as cenaval, coalesce(p.rab, 0)::float8 as rabat,
			coalesce(p.po, 0)::int as po, coalesce(s.tezjm, 0)::float8 as tezjm,
			coalesce((select j.brdecimala from jedmere j where j.jm = s.jm and j.god = p.god and j.kar = p.kar limit 1), 2)::int as decimale,
			coalesce(p.kolic, 0)::float8 as isporuceno, coalesce(s.kubikaza, 0)::float8 as kubikaza
		from epro p`).
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderBy("p.rbr")
	var rows []komProStavkaDto
	if err := r.selectRows(ctx, "stavke profakture", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomProStavkaRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomProStavkaRow{RproID: row.RproID, Rbr: row.Rbr, Konto: row.Konto, Sifra: row.Sifra, RsifID: row.RsifID, Naziv: row.Naziv, JM: row.JM,
			Komercopis: row.Komercopis, Pro: row.Pro, Proizsifra: row.Proizsifra, Otk: row.Otk, Serija: row.Serija, Roktr: row.Roktr,
			Kolicina: row.Kolicina, Cena: row.Cena, Fcena: row.Fcena, CenaVal: row.CenaVal, Rabat: row.Rabat, Po: row.Po, Tezjm: row.Tezjm,
			Decimale: row.Decimale, Isporuceno: row.Isporuceno, Kubikaza: row.Kubikaza}
	}
	return out, nil
}

// komProStavkaDto is a row of KomProStavke.
type komProStavkaDto struct {
	RproID     int64   `db:"rproid"`
	Rbr        int     `db:"rbr"`
	Konto      string  `db:"konto"`
	Sifra      int     `db:"sifra"`
	RsifID     int64   `db:"rsifid"`
	Naziv      string  `db:"naziv"`
	JM         string  `db:"jm"`
	Komercopis string  `db:"komercopis"`
	Pro        string  `db:"pro"`
	Proizsifra string  `db:"proizsifra"`
	Otk        string  `db:"otk"`
	Serija     string  `db:"serija"`
	Roktr      int64   `db:"roktr"`
	Kolicina   float64 `db:"kolicina"`
	Cena       float64 `db:"cena"`
	Fcena      float64 `db:"fcena"`
	CenaVal    float64 `db:"cenaval"`
	Rabat      float64 `db:"rabat"`
	Po         int     `db:"po"`
	Tezjm      float64 `db:"tezjm"`
	Decimale   int     `db:"decimale"`
	Isporuceno float64 `db:"isporuceno"`
	Kubikaza   float64 `db:"kubikaza"`
}

// KomProStavkaDodaj adds a stavka of a profaktura (the legacy AddEPRO): the artikal with its lot, the konto, the
// tražena količina (trazkolic) and the isporučena one (kolic), the planska cena (cena, ncena, pcena), the prodajna cena in RSD (fcena) and in the
// valuta (cenaval), the rabat, the magacin and the vrsta, broj and datum of the profaktura, with the audit columns.
func (r *KomercijalnoRepository) KomProStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomProStavkaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, sifra, po, rsifid, rbr, mag, magaciniid, konto, trazkolic, kolic, ncena, cena,
			pcena, fcena, cenaval, rab, otk, serija, roktr, vrd, dokum, dadok, xdatunosa, xopunos)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, st.RdokID, st.Sifra, st.Po, st.RsifID, st.Rbr, st.Mag, st.MagaciniID, st.Konto, st.Kolicina, st.Isporuceno, st.Cena, st.Cena,
		st.Cena, st.Fcena, st.CenaVal, st.Rabat, st.Otk, st.Serija, st.Roktr, st.Vrd, st.Dokum, st.Dadok, s.UserName)
	_, err = r.exec(ctx, tx, "nova stavka profakture", b)
	return err
}

// KomProStavkaIzmeni changes the količina, the prodajna cena (fcena in RSD, cenaval in the valuta) and the rabat of a
// stavka of a profaktura of the period (the legacy exit of a changed row of TABLE_PR_QRY_EPRO), with the audit columns;
// ErrNotFound without it.
func (r *KomercijalnoRepository) KomProStavkaIzmeni(ctx context.Context, tx db.Transaction, st domain.KomProStavkaIzmena) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("epro").
		Set("trazkolic = ?", st.Kolicina).
		Set("fcena = ?", st.Fcena).
		Set("cenaval = ?", st.CenaVal).
		Set("rab = ?", st.Rabat).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rproid = ?", st.RproID).
		Where("rdokid = ?", st.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stavka profakture", b)
}

// KomProLotovi returns the lots (drsta) of an artikal in a magacin of the period with their stanje (ulaz - izlaz -
// rezervisano), the ones with stanje first by their OTK (the legacy pr_Trazi_GPROIZVOD_ET); with sStanjem only the ones
// with stanje greater than 0.
func (r *KomercijalnoRepository) KomProLotovi(ctx context.Context, magaciniID int, rsifID int64, saStanjem bool) ([]domain.KomProLot, error) {
	b := querybuilder.New(`select d.rstaid::bigint as rstaid, trim(coalesce(d.konto, '')) as konto, coalesce(d.otk, '') as otk,
			coalesce(d.serija, '') as serija, coalesce(d.roktr, 0)::bigint as roktr,
			(coalesce(d.ulaz, 0) - coalesce(d.izlaz, 0) - coalesce(d.rez, 0))::float8 as stanje
		from drsta d`).
		Where("d.magaciniid = ?", magaciniID).
		Where("d.rsifid = ?", rsifID)
	if saStanjem {
		b.Where("coalesce(d.ulaz, 0) - coalesce(d.izlaz, 0) - coalesce(d.rez, 0) > 0")
	}
	if err := inPeriod(ctx, b, "d."); err != nil {
		return nil, err
	}
	b.OrderByExpr("(coalesce(d.ulaz, 0) - coalesce(d.izlaz, 0) - coalesce(d.rez, 0) > 0) desc, d.otk, d.serija, d.rstaid")
	var rows []komProLotDto
	if err := r.selectRows(ctx, "lotovi artikla", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomProLot, len(rows))
	for i, row := range rows {
		out[i] = domain.KomProLot{RstaID: row.RstaID, Konto: row.Konto, Otk: row.Otk, Serija: row.Serija, Roktr: row.Roktr, Stanje: row.Stanje}
	}
	return out, nil
}

// komProLotDto is a row of KomProLotovi.
type komProLotDto struct {
	RstaID int64   `db:"rstaid"`
	Konto  string  `db:"konto"`
	Otk    string  `db:"otk"`
	Serija string  `db:"serija"`
	Roktr  int64   `db:"roktr"`
	Stanje float64 `db:"stanje"`
}

// KomProCenaArtikla returns the prodajna cena of an artikal of the period (the first rcene of the šifra; the legacy
// RCENE.CENA of a magacin with the tip cene 1); 0 without it.
func (r *KomercijalnoRepository) KomProCenaArtikla(ctx context.Context, sifra int) (float64, error) {
	b := querybuilder.New(`select coalesce(c.cena, 0)::float8 from rcene c`).Where("c.sifra = ?", sifra)
	if err := inPeriod(ctx, b, "c."); err != nil {
		return 0, err
	}
	b.OrderBy("c.rceneid").Limit(1)
	var cena float64
	err := r.row(ctx, nil, "cena artikla", b, &cena)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return cena, err
}

// KomProCenaCenovnika returns the veleprodajna cena and the rabat of an artikal in a cenovnik of the period (rcenovnik
// of the broj of its header; the legacy RCENOVNIK of a magacin with the tip cene 2); found is false without it.
func (r *KomercijalnoRepository) KomProCenaCenovnika(ctx context.Context, cenovnikID int64, sifra int) (vpcena, rabat float64, found bool, err error) {
	b := querybuilder.New(`select coalesce(c.vpcena, 0)::float8, coalesce(c.rabat, 0)::float8 from rcenovnik c`).
		AddJoin("join rcenovnikhdr h on h.rcenovnikhdrid = c.rcenovnikhdrid and h.broj = c.broj").
		Where("c.rcenovnikhdrid = ?", cenovnikID).
		Where("c.sifra = ?", sifra)
	if err := inPeriod(ctx, b, "c."); err != nil {
		return 0, 0, false, err
	}
	b.OrderBy("c.rcenovnikid").Limit(1)
	err = r.row(ctx, nil, "cena cenovnika", b, &vpcena, &rabat)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	return vpcena, rabat, true, nil
}

// KomProPlanskaCena returns the planska cena of an artikal of the period (the first default normativ, pinorm.dftnort;
// the legacy planska cena of a magacin of the tip G); found is false without it.
func (r *KomercijalnoRepository) KomProPlanskaCena(ctx context.Context, sifra int) (float64, bool, error) {
	b := querybuilder.New(`select coalesce(n.plcena, 0)::float8 from pinorm n`).Where("n.dftnort").Where("n.sifra = ?", sifra)
	if err := inPeriod(ctx, b, "n."); err != nil {
		return 0, false, err
	}
	b.OrderBy("n.pinormid").Limit(1)
	var cena float64
	err := r.row(ctx, nil, "planska cena", b, &cena)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return cena, true, nil
}

// KomProNalogDodaj adds the nalog of a profaktura of the period (the legacy FormirajNALOG): a document of the vrsta
// with the broj, not of a nalog, new (doksts 10), not posted (knjige N), its opis "Nalog za izdavanje robe br: <broj>",
// the header of the profaktura (vrdokid: the profaktura; the legacy copies neither the kurs nor the poreska kategorija),
// the JM of the ambalaža empty, with the audit columns. It returns its rdokid.
func (r *KomercijalnoRepository) KomProNalogDodaj(ctx context.Context, tx db.Transaction, proID int64, vrd, dokum int) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into edok (god, kar, rnalid, vrd, tipdok, dokum, dadok, mag, magaciniid, iznos, rbr, opis, rok, fkto, fana,
			komid, pla, foot, otp, posuslnab, brst, brotp, vaziza, sifval, idorgjed, mestotrid, vrugrabat, ugrabat, vkase, pkase, doksts,
			rcenovnikhdrid, bttowght, netwght, totboxs, brpaleta, vrdokid, knjige_1, knjige_2, jm1, jm2, xdatunosa, xopunos)
		select z.god, z.kar, 0, ?, z.tipdok, ?, z.dadok, z.mag, z.magaciniid, z.iznos, 0, ?, z.rok, z.fkto, z.fana, z.komid, z.pla, z.foot, z.otp,
			z.posuslnab, 0, 0, z.vaziza, z.sifval, z.idorgjed, z.mestotrid, z.vrugrabat, z.ugrabat, z.vkase, z.pkase, ?, z.rcenovnikhdrid,
			z.bttowght, z.netwght, z.totboxs, z.brpaleta, z.rdokid, 'N', 'N', '', '', now(), ?
		from edok z`,
		vrd, dokum, fmt.Sprintf("Nalog za izdavanje robe br: %d", dokum), domain.KomProfakturaNova, s.UserName).
		Where("z.rdokid = ?", proID).
		Returning("rdokid")
	if err := inPeriod(ctx, b, "z."); err != nil {
		return 0, err
	}
	var rdokID int64
	if err := r.row(ctx, tx, "nalog profakture", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// KomProNalogStavke copies the stavke of a profaktura of the period into its nalog (the legacy FormirajNALOG "KOPIRAJ
// DETALJE NALOGA"): the artikal with its lot and konto, the količina, the cene and the rabat, the vrsta and the broj of
// the nalog and the datum of the stavka, with the audit columns.
func (r *KomercijalnoRepository) KomProNalogStavke(ctx context.Context, tx db.Transaction, proID, nalogID int64, vrd, dokum int) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, sifra, po, rsifid, rbr, mag, magaciniid, trazkolic, ncena, cena, fcena, cenaval,
			rab, otk, serija, roktr, konto, vrd, dokum, dadok, xdatunosa, xopunos)
		select p.god, p.kar, ?, p.sifra, p.po, p.rsifid, p.rbr, p.mag, p.magaciniid, p.trazkolic, p.ncena, p.cena, p.fcena, p.cenaval, p.rab,
			p.otk, p.serija, p.roktr, p.konto, ?, ?, p.dadok, now(), ?
		from epro p`, nalogID, vrd, dokum, s.UserName).
		Where("p.rdokid = ?", proID)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return err
	}
	_, err = r.exec(ctx, tx, "stavke naloga profakture", b)
	return err
}

// KomProNalogFormiran marks a profaktura of the period whose nalog was formed (doksts 40; the legacy FormirajNALOG). It
// returns false when the profaktura already has a nalog.
func (r *KomercijalnoRepository) KomProNalogFormiran(ctx context.Context, tx db.Transaction, proID int64) (bool, error) {
	return r.komProStatus(ctx, tx, proID, domain.KomProfakturaNalog, domain.KomProfakturaNova, domain.KomProfakturaPotvrdjena, "")
}

// KomProPrikaz returns the profakture of the list of the main screen (the legacy PR_QRY_EDOK of the tab "Pregled
// profaktura"): the ones of the vrste of the profakture with the status (the nepotvrđene 10, else the potvrđene 20 and
// 40) in the magacini of the user of the session (maguser), of one magacin or of all of them (SviMagacini), the last
// ones first, with the "mag-opis" of the magacin, the naziv of the kupac and the broj of the nalog formed from them.
func (r *KomercijalnoRepository) KomProPrikaz(ctx context.Context, p domain.KomProPrikazParams) ([]domain.KomProPrikazRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select e.rdokid::bigint as rdokid,
			coalesce(m.mag::text || '-' || coalesce(m.opis, ''), '') as magacin, coalesce(e.vrd, 0)::int as vrd,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok, coalesce(e.fkto, '') as fkto, coalesce(e.fana, '') as fana,
			coalesce(f.naziv, '') as kupac, coalesce(e.opis, '') as opis, coalesce(e.rok, 0)::int as rok, coalesce(e.otp, '') as otp,
			coalesce(e.doksts, '') as doksts,
			coalesce((select n.dokum from edok n where n.vrdokid = e.rdokid and n.god = e.god and n.kar = e.kar order by n.rdokid limit 1), 0)::int
				as nalogdokum
		from edok e`).
		AddJoin("join (select distinct mu.mag from maguser mu where upper(mu.iduser) = ?) u on u.mag = e.mag",
			strings.ToUpper(strings.TrimSpace(s.UserName))).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join fkpl f on f.god = e.god and f.kar = e.kar and f.vkonta = 1 and f.konto = e.fkto and f.sifra = e.fana").
		Where(komProVrstaUslov, komZahtevModul, komProGrupa)
	if p.Doksts == domain.KomProfakturaNova {
		b.Where("coalesce(e.doksts, '') = ?", domain.KomProfakturaNova)
	} else {
		b.Where("coalesce(e.doksts, '') in (?, ?)", domain.KomProfakturaPotvrdjena, domain.KomProfakturaNalog)
	}
	if !p.SviMagacini {
		b.Where("e.magaciniid = ?", p.MagaciniID)
	}
	komPretraga(b, p.Tekst, "e.dokum::text", "e.foot", "f.naziv", "e.fana")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return nil, err
	}
	b.OrderByExpr("e.dadok desc nulls last, e.vrd desc, e.dokum desc")
	var rows []komProPrikazDto
	if err := r.selectRows(ctx, "pregled profaktura", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomProPrikazRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomProPrikazRow{RdokID: row.RdokID, Magacin: row.Magacin, Vrd: row.Vrd, Dokum: row.Dokum, Dadok: row.Dadok, Fkto: row.Fkto,
			Fana: row.Fana, Kupac: row.Kupac, Opis: row.Opis, Rok: row.Rok, Otp: row.Otp, Doksts: row.Doksts, NalogDokum: row.NalogDokum}
	}
	return out, nil
}

// komProPrikazDto is a row of KomProPrikaz.
type komProPrikazDto struct {
	RdokID     int64        `db:"rdokid"`
	Magacin    string       `db:"magacin"`
	Vrd        int          `db:"vrd"`
	Dokum      int          `db:"dokum"`
	Dadok      sql.NullTime `db:"dadok"`
	Fkto       string       `db:"fkto"`
	Fana       string       `db:"fana"`
	Kupac      string       `db:"kupac"`
	Opis       string       `db:"opis"`
	Rok        int          `db:"rok"`
	Otp        string       `db:"otp"`
	Doksts     string       `db:"doksts"`
	NalogDokum int          `db:"nalogdokum"`
}
