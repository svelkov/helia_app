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

// The zahtevi za ponudu of the komercijalno (the legacy pr_ZahtevPONUDA): edok of the vrste of the modul PRO and the
// group ZPN with their stavke in epro, taken mostly from the slobodne stavke of the potvrđeni zahtevi za nabavku.

// komPonudaGrupa is the group of the vrste dokumenata of the zahtevi za ponudu (the legacy PR_QRY_DOKVRSTA with "ZPN").
const komPonudaGrupa = "ZPN"

// The ststatus of a stavka of a zahtev za nabavku: slobodna (it can be taken by a zahtev za ponudu) or rezervisana
// (taken by one, the legacy AddEPRONAR).
const (
	komStavkaSlobodna    = "S"
	komStavkaRezervisana = "R"
)

// KomPonudaVrste returns the vrste dokumenata of the zahtevi za ponudu of the period (dokvrsta of the modul PRO and the
// group ZPN) as "vrd - opis" keyed by the vrd, in the order of the vrd.
func (r *KomercijalnoRepository) KomPonudaVrste(ctx context.Context) ([]domain.ComboItem, error) {
	return r.komVrste(ctx, komPonudaGrupa)
}

// KomPonuda returns a zahtev za ponudu of the period with the texts of its komercijalista, org. jedinica, mesto troška
// and valuta; found is false without it.
func (r *KomercijalnoRepository) KomPonuda(ctx context.Context, rdokID int64) (domain.KomPonudaRow, bool, error) {
	b := querybuilder.New(`select e.rdokid::bigint, coalesce(e.vrd, 0)::int, coalesce(e.dokum, 0)::int, e.dadok, coalesce(e.rok, 0)::int,
			coalesce(e.magaciniid, 0)::int, coalesce(e.mag, 0)::int, e.komid::bigint, coalesce(k.sifkom, 0)::int, coalesce(k.imeprezime, ''),
			e.idorgjed::bigint, coalesce(concat_ws(' - ', nullif(o.ojozn::text, ''), nullif(o.naziv, '')), ''),
			e.mestotrid::bigint, coalesce(concat_ws(' - ', nullif(t.mtroska, ''), nullif(t.opis, '')), ''),
			coalesce(e.foot, ''), coalesce(e.doksts, ''), coalesce(e.xopunos, ''), coalesce(e.knjige_1, ''), coalesce(e.knjige_4, ''),
			coalesce(e.sifval, 0)::int, coalesce((select v.oznval from valute v where v.sifval = e.sifval order by v.idvalute limit 1), ''),
			coalesce(e.kurs, 0)::float8, coalesce(e.vaziza, 0)::int, coalesce(e.vezniid, '')
		from edok e`).
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		AddJoin("left join mestotr t on t.mestotrid = e.mestotrid").
		Where("e.rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, "e."); err != nil {
		return domain.KomPonudaRow{}, false, err
	}
	var z domain.KomPonudaRow
	err := r.row(ctx, nil, "zahtev za ponudu", b, &z.RdokID, &z.Vrd, &z.Dokum, &z.Dadok, &z.Rok, &z.MagaciniID, &z.Mag, &z.KomID,
		&z.KomSifra, &z.KomIme, &z.Orgjed, &z.OrgjedText, &z.MestoTroska, &z.MestoTroskaText, &z.Foot, &z.Doksts, &z.Opunos, &z.Knjige1,
		&z.Knjige4, &z.SifVal, &z.OznVal, &z.Kurs, &z.VaziZa, &z.VezniID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPonudaRow{}, false, nil
	}
	if err != nil {
		return domain.KomPonudaRow{}, false, err
	}
	return z, true, nil
}

// KomPonudaDodaj adds a zahtev za ponudu (the legacy SnimiEDOK of a new document): not of a nalog, nepotvrđen
// (doksts 10), not posted (knjige N), its opis "Zahtev za ponudu br: <broj>", the rok, the valuta with its kurs and the
// dobavljači (vezniid), the JM of the ambalaža empty (jm1 and jm2 are not null without a default), with the audit
// columns. It returns its rdokid.
func (r *KomercijalnoRepository) KomPonudaDodaj(ctx context.Context, tx db.Transaction, z domain.KomPonudaUpis) (int64, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return 0, err
	}
	b := querybuilder.New(`insert into edok (god, kar, rnalid, vrd, dokum, dadok, mag, magaciniid, iznos, rbr, opis, rok, fkto, fana,
			pla, otp, foot, brst, vrdokid, idorgjed, mestotrid, komid, sifval, kurs, vaziza, vezniid, doksts, knjige_1, knjige_2, jm1, jm2,
			xdatunosa, xopunos)
		values (?, ?, 0, ?, ?, ?, ?, ?, 0, 0, ?, ?, '', '', '', '', ?, 0, 0, ?, ?, ?, ?, ?, ?, ?, ?, 'N', 'N', '', '', now(), ?)`,
		s.God, s.Kar, z.Vrd, z.Dokum, z.Dadok, z.Mag, z.MagaciniID, fmt.Sprintf("Zahtev za ponudu br: %d", z.Dokum), z.Rok, z.Foot,
		z.Orgjed, z.MestoTroska, z.KomID, z.SifVal, z.Kurs, z.VaziZa, z.VezniID, domain.KomZahtevNepotvrdjen, s.UserName).
		Returning("rdokid")
	var rdokID int64
	if err := r.row(ctx, tx, "novi zahtev za ponudu", b, &rdokID); err != nil {
		return 0, err
	}
	return rdokID, nil
}

// KomPonudaIzmeni saves the header of a zahtev za ponudu of the period (the legacy SnimiEDOK of a saved document): its
// datum, rok, komercijalista, org. jedinica, mesto troška, tekst, valuta with its kurs and dobavljači, with the audit
// columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomPonudaIzmeni(ctx context.Context, tx db.Transaction, z domain.KomPonudaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("edok").
		Set("dadok = ?", z.Dadok).
		Set("rok = ?", z.Rok).
		Set("komid = ?", z.KomID).
		Set("idorgjed = ?", z.Orgjed).
		Set("mestotrid = ?", z.MestoTroska).
		Set("foot = ?", z.Foot).
		Set("sifval = ?", z.SifVal).
		Set("kurs = ?", z.Kurs).
		Set("vaziza = ?", z.VaziZa).
		Set("vezniid = ?", z.VezniID).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", z.RdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "zahtev za ponudu", b)
}

// KomPonudaPotvrdi confirms a nepotvrđen zahtev za ponudu of the period with the dobavljač it is confirmed for (the
// legacy BTN_Potvrda of pr_PotvrdaPONUDE: doksts 40, fkto and fana the konto and the šifra of the dobavljač). It returns
// false when the zahtev is not nepotvrđen any more (e.g. confirmed meanwhile).
func (r *KomercijalnoRepository) KomPonudaPotvrdi(ctx context.Context, rdokID int64, konto, sifra string) (bool, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return false, err
	}
	b := querybuilder.NewUpdate("edok").
		Set("doksts = ?", domain.KomZahtevPotvrdjen).
		Set("fkto = ?", konto).
		Set("fana = ?", sifra).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rdokid = ?", rdokID).
		Where("coalesce(doksts, '') = ?", domain.KomZahtevNepotvrdjen)
	if err := inPeriod(ctx, b, ""); err != nil {
		return false, err
	}
	n, err := r.exec(ctx, nil, "potvrda zahteva za ponudu", b)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// KomPonudaBrisi deletes a zahtev za ponudu of the period with its stavke (the legacy BTN_Brisanje): the stavke of the
// zahtevi za nabavku they were taken from are slobodne again; ErrNotFound without the zahtev.
func (r *KomercijalnoRepository) KomPonudaBrisi(ctx context.Context, tx db.Transaction, rdokID int64) error {
	oslobodi := querybuilder.NewUpdate("epro").
		Set("ststatus = ?", komStavkaSlobodna).
		Where("rproid in (select p.rproid1 from epro p where p.rdokid = ? and p.god = epro.god and p.kar = epro.kar)", rdokID)
	if err := inPeriod(ctx, oslobodi, ""); err != nil {
		return err
	}
	if _, err := r.exec(ctx, tx, "oslobađanje stavki zahteva za nabavku", oslobodi); err != nil {
		return err
	}
	return r.KomZahtevBrisi(ctx, tx, rdokID)
}

// KomPonudaIzvori returns the slobodne stavke (ststatus S) of the potvrđeni zahtevi za nabavku (doksts 40, the vrste
// of the group INR) of the magacin in the period, with the godina, broj, datum and org. jedinica of their zahtev and the
// naziv and JM of their artikal (the legacy PuniSTAVKE).
func (r *KomercijalnoRepository) KomPonudaIzvori(ctx context.Context, magaciniID int) ([]domain.KomPonudaIzvorRow, error) {
	b := querybuilder.New(`select p.rproid::bigint as rproid, e.rdokid::bigint as rdokid, coalesce(e.god, 0)::int as god,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok,
			coalesce(o.ojozn::text || '-' || coalesce(o.naziv, ''), '') as oj, coalesce(p.sifra, 0)::int as sifra,
			coalesce(s.naziv, '') as naziv, coalesce(s.jm, '') as jm, coalesce(p.trazkolic, 0)::float8 as kolicina
		from epro p`).
		AddJoin("join edok e on e.rdokid = p.rdokid and e.god = p.god and e.kar = p.kar").
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		AddJoin("left join orgjed o on o.idorgjed = e.idorgjed").
		Where("coalesce(e.doksts, '') = ?", domain.KomZahtevPotvrdjen).
		Where("e.magaciniid = ?", magaciniID).
		Where("coalesce(p.ststatus, '') = ?", komStavkaSlobodna).
		Where(`e.vrd in (select v.vrd from dokvrsta v where v.god = e.god and v.kar = e.kar
			and upper(trim(coalesce(v.modul, ''))) = ? and upper(trim(coalesce(v.grpdok, ''))) = ?)`, komZahtevModul, komZahtevGrupa)
	if err := inPeriod(ctx, b, "p."); err != nil {
		return nil, err
	}
	b.OrderByExpr("e.dadok nulls last, e.dokum, p.rbr")
	var rows []komPonudaIzvorDto
	if err := r.selectRows(ctx, "stavke zahteva za nabavku za ponudu", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPonudaIzvorRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPonudaIzvorRow{RproID: row.RproID, RdokID: row.RdokID, God: row.God, Dokum: row.Dokum, Dadok: row.Dadok, OJ: row.OJ,
			Sifra: row.Sifra, Naziv: row.Naziv, JM: row.JM, Kolicina: row.Kolicina}
	}
	return out, nil
}

// komPonudaIzvorDto is a row of KomPonudaIzvori.
type komPonudaIzvorDto struct {
	RproID   int64        `db:"rproid"`
	RdokID   int64        `db:"rdokid"`
	God      int          `db:"god"`
	Dokum    int          `db:"dokum"`
	Dadok    sql.NullTime `db:"dadok"`
	OJ       string       `db:"oj"`
	Sifra    int          `db:"sifra"`
	Naziv    string       `db:"naziv"`
	JM       string       `db:"jm"`
	Kolicina float64      `db:"kolicina"`
}

// KomPonudaIzvor returns a stavka of a zahtev za nabavku of the period a zahtev za ponudu can take (the legacy
// AddEPRONAR): a slobodna stavka of a potvrđen zahtev za nabavku that no stavka has taken yet (no epro with its rproid
// in rproid1), with the poreska stopa of its artikal; found is false otherwise.
func (r *KomercijalnoRepository) KomPonudaIzvor(ctx context.Context, tx db.Transaction, rproID int64) (domain.KomPonudaIzvor, bool, error) {
	b := querybuilder.New(`select p.rproid::bigint, coalesce(p.sifra, 0)::int, coalesce(s.po, 0)::int, coalesce(p.rsifid, 0)::bigint,
			coalesce(p.trazkolic, 0)::float8
		from epro p`).
		AddJoin("join edok e on e.rdokid = p.rdokid and e.god = p.god and e.kar = p.kar").
		AddJoin("left join rsif s on s.rsifid = p.rsifid").
		Where("p.rproid = ?", rproID).
		Where("coalesce(e.doksts, '') = ?", domain.KomZahtevPotvrdjen).
		Where("coalesce(p.ststatus, '') = ?", komStavkaSlobodna).
		Where("not exists (select 1 from epro x where x.rproid1 = p.rproid and x.god = p.god and x.kar = p.kar)")
	if err := inPeriod(ctx, b, "p."); err != nil {
		return domain.KomPonudaIzvor{}, false, err
	}
	var st domain.KomPonudaIzvor
	err := r.row(ctx, tx, "stavka zahteva za nabavku", b, &st.RproID, &st.Sifra, &st.Po, &st.RsifID, &st.Kolicina)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPonudaIzvor{}, false, nil
	}
	if err != nil {
		return domain.KomPonudaIzvor{}, false, err
	}
	return st, true, nil
}

// KomPonudaIzvorRezervisi marks a stavka of a zahtev za nabavku of the period taken by a zahtev za ponudu (ststatus R,
// the legacy AddEPRONAR); ErrNotFound without it.
func (r *KomercijalnoRepository) KomPonudaIzvorRezervisi(ctx context.Context, tx db.Transaction, rproID int64) error {
	b := querybuilder.NewUpdate("epro").Set("ststatus = ?", komStavkaRezervisana).Where("rproid = ?", rproID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "rezervacija stavke zahteva za nabavku", b)
}

// KomPonudaStavkaDodaj adds a stavka of a zahtev za ponudu (the legacy AddEPRO for a stavka entered by hand, AddEPRONAR
// for a stavka of a zahtev za nabavku with its rproid in rproid1): the artikal, the tražena količina, the magacin and
// the vrsta, broj and datum of the document, with the audit columns.
func (r *KomercijalnoRepository) KomPonudaStavkaDodaj(ctx context.Context, tx db.Transaction, st domain.KomPonudaStavkaUpis) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.New(`insert into epro (god, kar, rdokid, sifra, po, rsifid, rbr, mag, magaciniid, trazkolic, rproid1, vrd, dokum, dadok,
			xdatunosa, xopunos)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)`,
		s.God, s.Kar, st.RdokID, st.Sifra, st.Po, st.RsifID, st.Rbr, st.Mag, st.MagaciniID, st.Kolicina, st.RproID1, st.Vrd, st.Dokum,
		st.Dadok, s.UserName)
	_, err = r.exec(ctx, tx, "nova stavka zahteva za ponudu", b)
	return err
}

// KomPonudaStavkaKolicina changes the tražena količina of a stavka of a zahtev za ponudu of the period (the legacy exit
// of a changed row of TABLE_PR_QRY_EPRO1), with the audit columns; ErrNotFound without it.
func (r *KomercijalnoRepository) KomPonudaStavkaKolicina(ctx context.Context, tx db.Transaction, rdokID, rproID int64, kolicina float64) error {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return err
	}
	b := querybuilder.NewUpdate("epro").
		Set("trazkolic = ?", kolicina).
		Set("xdatizmene = now()").
		Set("xopizmene = ?", s.UserName).
		Where("rproid = ?", rproID).
		Where("rdokid = ?", rdokID)
	if err := inPeriod(ctx, b, ""); err != nil {
		return err
	}
	return r.execOne(ctx, tx, "stavka zahteva za ponudu", b)
}

// KomPonudaStavkaBrisi deletes a stavka of a zahtev za ponudu of the period (the legacy BTN_BrisiDET): the stavka of the
// zahtev za nabavku it was taken from is slobodna again and the redni brojevi of the stavke after it move one back;
// ErrNotFound without it.
func (r *KomercijalnoRepository) KomPonudaStavkaBrisi(ctx context.Context, tx db.Transaction, rdokID, rproID int64, rbr int) error {
	oslobodi := querybuilder.NewUpdate("epro").
		Set("ststatus = ?", komStavkaSlobodna).
		Where("rproid in (select p.rproid1 from epro p where p.rproid = ? and p.rdokid = ? and p.god = epro.god and p.kar = epro.kar)",
			rproID, rdokID)
	if err := inPeriod(ctx, oslobodi, ""); err != nil {
		return err
	}
	if _, err := r.exec(ctx, tx, "oslobađanje stavke zahteva za nabavku", oslobodi); err != nil {
		return err
	}
	return r.KomZahtevStavkaBrisi(ctx, tx, rdokID, rproID, rbr)
}

// KomDobavljaci returns the dobavljači of the ids (fkpl of the period with the adresa and the mesto of their partner),
// in any order; an empty list is an error.
func (r *KomercijalnoRepository) KomDobavljaci(ctx context.Context, ids []int64) ([]domain.KomDobavljac, error) {
	if len(ids) == 0 {
		return nil, errors.New("dobavljači: no ids")
	}
	ph := strings.TrimSuffix(strings.Repeat("?, ", len(ids)), ", ")
	b := querybuilder.New(`select f.idfkpl::bigint as idfkpl, coalesce(f.konto, '') as konto, coalesce(f.sifra, '') as sifra,
			coalesce(f.naziv, '') as naziv, coalesce(p.adresa, '') as adresa, coalesce(p.mesto, '') as mesto, f.idpartneri::bigint as idpartneri
		from fkpl f`).
		AddJoin("left join partneri p on p.idpartneri = f.idpartneri").
		Where("f.idfkpl in ("+ph+")", querybuilder.Args(ids)...)
	if err := inPeriod(ctx, b, "f."); err != nil {
		return nil, err
	}
	var rows []komDobavljacDto
	if err := r.selectRows(ctx, "dobavljači", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomDobavljac, len(rows))
	for i, row := range rows {
		out[i] = domain.KomDobavljac{IDFkpl: row.IDFkpl, Konto: row.Konto, Sifra: row.Sifra, Naziv: row.Naziv, Adresa: row.Adresa,
			Mesto: row.Mesto, IDPartneri: row.IDPartneri}
	}
	return out, nil
}

// komDobavljacDto is a row of KomDobavljaci.
type komDobavljacDto struct {
	IDFkpl     int64         `db:"idfkpl"`
	Konto      string        `db:"konto"`
	Sifra      string        `db:"sifra"`
	Naziv      string        `db:"naziv"`
	Adresa     string        `db:"adresa"`
	Mesto      string        `db:"mesto"`
	IDPartneri sql.NullInt64 `db:"idpartneri"`
}

// KomPonudePrikaz returns the zahtevi za ponudu of the tab "Pregled ponuda" (the legacy PuniHDRPONUDE): the ones of the
// vrsta and the status in the magacini of the user of the session (maguser), of one magacin or of all of them
// (SviMagacini), the last ones first, with the "mag-opis" of the magacin, the nazivi of the dobavljači (vezniid, in
// their order) and the "šifra-ime" of the komercijalista.
func (r *KomercijalnoRepository) KomPonudePrikaz(ctx context.Context, p domain.KomPonudaPrikazParams) ([]domain.KomPonudaPrikazRow, error) {
	s, err := sessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	b := querybuilder.New(`select e.rdokid::bigint as rdokid,
			coalesce(m.mag::text || '-' || coalesce(m.opis, ''), '') as magacin, coalesce(e.vrd, 0)::int as vrd,
			coalesce(e.dokum, 0)::int as dokum, e.dadok as dadok, coalesce(e.rok, 0)::int as rok,
			coalesce((select string_agg(coalesce(f.naziv, ''), ', ' order by array_position(string_to_array(e.vezniid, ';'), f.idfkpl::text))
				from fkpl f where f.god = e.god and f.kar = e.kar and f.idfkpl::text = any(string_to_array(e.vezniid, ';'))), '') as dobavljaci,
			coalesce(k.sifkom::text || '-' || coalesce(k.imeprezime, ''), '') as referent, coalesce(e.foot, '') as foot
		from edok e`).
		AddJoin("join (select distinct mu.mag from maguser mu where upper(mu.iduser) = ?) u on u.mag = e.mag",
			strings.ToUpper(strings.TrimSpace(s.UserName))).
		AddJoin("left join magacini m on m.magaciniid = e.magaciniid").
		AddJoin("left join komercijalisti k on k.komid = e.komid").
		Where("e.vrd = ?", p.Vrd).
		Where("coalesce(e.doksts, '') = ?", p.Doksts)
	if !p.SviMagacini {
		b.Where("e.magaciniid = ?", p.MagaciniID)
	}
	komPretraga(b, p.Tekst, "e.dokum::text", "e.foot", "k.imeprezime")
	if err := inPeriod(ctx, b, "e."); err != nil {
		return nil, err
	}
	b.OrderByExpr("e.dadok desc nulls last, e.dokum desc")
	var rows []komPonudaPrikazDto
	if err := r.selectRows(ctx, "pregled zahteva za ponudu", &rows, b); err != nil {
		return nil, err
	}
	out := make([]domain.KomPonudaPrikazRow, len(rows))
	for i, row := range rows {
		out[i] = domain.KomPonudaPrikazRow{RdokID: row.RdokID, Magacin: row.Magacin, Vrd: row.Vrd, Dokum: row.Dokum, Dadok: row.Dadok,
			Rok: row.Rok, Dobavljaci: row.Dobavljaci, Referent: row.Referent, Foot: row.Foot}
	}
	return out, nil
}

// komPonudaPrikazDto is a row of KomPonudePrikaz.
type komPonudaPrikazDto struct {
	RdokID     int64        `db:"rdokid"`
	Magacin    string       `db:"magacin"`
	Vrd        int          `db:"vrd"`
	Dokum      int          `db:"dokum"`
	Dadok      sql.NullTime `db:"dadok"`
	Rok        int          `db:"rok"`
	Dobavljaci string       `db:"dobavljaci"`
	Referent   string       `db:"referent"`
	Foot       string       `db:"foot"`
}

// KomPonudaStampaDobavljac returns the dobavljač a zahtev za ponudu of the period is printed for (the legacy
// PR_RPT_ZAHTEV_PONUDA: the konto and šifra of the analitika, the naziv and the data of its partner); found is false
// without the analitika.
func (r *KomercijalnoRepository) KomPonudaStampaDobavljac(ctx context.Context, idFkpl int64) (domain.KomPonudaStampaDobavljac, bool, error) {
	b := querybuilder.New(`select coalesce(f.konto, ''), coalesce(f.sifra, ''), coalesce(p.naziv, f.naziv, ''), coalesce(p.adresa, ''),
			coalesce(p.mesto, ''), coalesce(p.tippdv, 0)::int, coalesce(p.pib, ''), coalesce(p.budzetski, false), coalesce(p.jbkjs, ''),
			coalesce(p.jmbg, ''), coalesce(p.bpg, ''), coalesce(p.index, ''), coalesce(p.telefon, '')
		from fkpl f`).
		AddJoin("left join partneri p on p.idpartneri = f.idpartneri").
		Where("f.idfkpl = ?", idFkpl)
	if err := inPeriod(ctx, b, "f."); err != nil {
		return domain.KomPonudaStampaDobavljac{}, false, err
	}
	var d domain.KomPonudaStampaDobavljac
	err := r.row(ctx, nil, "dobavljač zahteva za ponudu", b, &d.Konto, &d.Sifra, &d.Naziv, &d.Adresa, &d.Mesto, &d.TipPdv, &d.Pib,
		&d.Budzetski, &d.Jbkjs, &d.Jmbg, &d.Bpg, &d.Index, &d.Telefon)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KomPonudaStampaDobavljac{}, false, nil
	}
	if err != nil {
		return domain.KomPonudaStampaDobavljac{}, false, err
	}
	return d, true, nil
}
